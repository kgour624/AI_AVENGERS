package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/gateway"
)

// TaskSpec is one expert's task produced by the Planner.
type TaskSpec struct {
	ExpertID           uuid.UUID
	Title              string
	Description        string
	DependsOnExpertIDs []uuid.UUID
}

// plannerMaxRetries: LLMs hallucinate ~30% on structured JSON output.
// 3 attempts covers the vast majority of transient failures.
const plannerMaxRetries = 3

// Planner uses a single LLM call to decompose a requirement into per-expert tasks.
// Domain-agnostic: no hardcoded domain strings. LLM reads expert charters.
// Reliable: retries up to plannerMaxRetries on parse failure or empty result.
type Planner struct {
	gateway *gateway.ModelGateway
	logger  *zap.Logger
}

func NewPlanner(gw *gateway.ModelGateway, logger *zap.Logger) *Planner {
	return &Planner{gateway: gw, logger: logger}
}

// workflowExpert holds expert data needed by Planner and AgentLoop.
// Separate from orchestrator.expertRecord: workflow needs loop_pattern,
// max_loop_iterations, allowed_tools which chat orchestrator does not.
type workflowExpert struct {
	ID                uuid.UUID
	Name              string
	Domain            string
	ReasoningCharter  string
	LoopPattern       string
	MaxLoopIterations int
	AllowedTools      []string
}

// Plan decomposes requirementText into one TaskSpec per expert.
// Retries up to plannerMaxRetries on JSON parse failure or 0 valid tasks.
//
// Mental execution:
//   Input: "Build URL shortener", experts=[PM, SD, DSA, LLD]
//   Attempt 1: LLM -> JSON -> validate UUIDs -> return []TaskSpec
//   Attempt 1 fail: LLM hallucinated UUID -> 0 valid tasks -> retry
//   Attempt 2: LLM -> JSON -> validate -> return
//   Attempt 3 fail: return error
func (p *Planner) Plan(ctx context.Context, workflowID uuid.UUID, requirementText string, experts []workflowExpert) ([]TaskSpec, error) {
	if len(experts) == 0 {
		return nil, fmt.Errorf("planner: no experts provided")
	}
	if strings.TrimSpace(requirementText) == "" {
		return nil, fmt.Errorf("planner: requirement text is empty")
	}

	// Build expert index for O(1) UUID validation.
	expertIndex := make(map[string]workflowExpert, len(experts))
	for _, e := range experts {
		expertIndex[e.ID.String()] = e
	}

	var expertList strings.Builder
	for _, e := range experts {
		expertList.WriteString(fmt.Sprintf(
			"- Expert ID: %s\n  Name: %s\n  Domain: %s\n  Charter: %s\n\n",
			e.ID.String(), e.Name, e.Domain, e.ReasoningCharter,
		))
	}
	userPrompt := fmt.Sprintf("EXPERTS:\n%s\nREQUIREMENT:\n%s", expertList.String(), requirementText)

	var lastErr error
	for attempt := 1; attempt <= plannerMaxRetries; attempt++ {
		tasks, err := p.planOnce(ctx, workflowID, userPrompt, expertIndex)
		if err != nil {
			lastErr = err
			p.logger.Warn("planner: attempt failed",
				zap.Int("attempt", attempt),
				zap.Error(err),
			)
			continue
		}
		p.logger.Info("planner: plan created",
			zap.Int("attempt", attempt),
			zap.Int("tasks", len(tasks)),
		)
		return tasks, nil
	}
	return nil, fmt.Errorf("planner: all %d attempts failed: %w", plannerMaxRetries, lastErr)
}

// planOnce makes one LLM call and validates the result.
func (p *Planner) planOnce(ctx context.Context, workflowID uuid.UUID, userPrompt string, expertIndex map[string]workflowExpert) ([]TaskSpec, error) {
	resp, err := p.gateway.Call(ctx, gateway.LLMRequest{
		Model:      gateway.ModelStrong,
		WorkflowID: &workflowID,
		SystemPrompt: `You are a workflow planner for a multi-agent software development system.

Given a list of domain experts and a requirement, create exactly one task per expert.
Each task must:
1. Be specific to that expert's domain and charter
2. Describe what artifact the expert should produce
3. List which other expert IDs this task depends on

Rules:
- Every expert MUST get exactly one task
- depends_on_expert_ids must only contain IDs from the provided expert list
- An expert cannot depend on itself
- Tasks with no dependencies run first

CRITICAL: Output ONLY a valid JSON array. No explanation, no markdown, no code fences.
Use exact UUIDs from the expert list.

Format:
[
  {
    "expert_id": "<exact UUID>",
    "title": "<short title, max 100 chars>",
    "description": "<detailed description, min 50 chars>",
    "depends_on_expert_ids": ["<exact UUID>", ...]
  }
]`,
		UserPrompt:  userPrompt,
		// 0 = let the gateway use the model's configured output limit. A hard
		// 2000 here is what made planning fail on a reasoning model: it spent
		// the whole budget thinking, returned no JSON, three times, and the
		// workflow died at intake. The number belongs to the model, which is an
		// admin setting (Admin → LLM Settings).
		MaxTokens:   0,
		Temperature: 0.1,
	})
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	raw := strings.TrimSpace(resp.Content)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	start := strings.Index(raw, "[")
	end := strings.LastIndex(raw, "]")
	if start == -1 || end == -1 || end <= start {
		return nil, fmt.Errorf("no JSON array in output: %q", truncate(raw, 200))
	}
	raw = raw[start : end+1]

	type rawTask struct {
		ExpertID           string   `json:"expert_id"`
		Title              string   `json:"title"`
		Description        string   `json:"description"`
		DependsOnExpertIDs []string `json:"depends_on_expert_ids"`
	}
	var rawTasks []rawTask
	if err := json.Unmarshal([]byte(raw), &rawTasks); err != nil {
		return nil, fmt.Errorf("JSON parse failed: %w", err)
	}
	if len(rawTasks) == 0 {
		return nil, fmt.Errorf("empty task list")
	}

	var tasks []TaskSpec
	for _, rt := range rawTasks {
		if _, ok := expertIndex[rt.ExpertID]; !ok {
			p.logger.Warn("planner: unknown expert_id skipped", zap.String("id", rt.ExpertID))
			continue
		}
		if strings.TrimSpace(rt.Title) == "" {
			continue
		}
		expertUUID, _ := uuid.Parse(rt.ExpertID)
		seen := make(map[uuid.UUID]bool)
		var deps []uuid.UUID
		for _, depStr := range rt.DependsOnExpertIDs {
			// Parse dependency UUID first to handle case-insensitive comparison
			depUUID, err := uuid.Parse(depStr)
			if err != nil {
				p.logger.Warn("planner: invalid dep UUID skipped", zap.String("dep", depStr))
				continue
			}
			// Check self-dependency using parsed UUIDs (case-insensitive)
			if depUUID == expertUUID {
				p.logger.Warn("planner: self-dependency removed",
					zap.String("expert_id", expertUUID.String()),
					zap.String("dep_id", depStr),
				)
				continue
			}
			// Check duplicate using parsed UUID
			if seen[depUUID] {
				continue
			}
			if _, ok := expertIndex[depStr]; !ok {
				p.logger.Warn("planner: unknown dep removed", zap.String("dep", depStr))
				continue
			}
			seen[depUUID] = true
			deps = append(deps, depUUID)
		}
		tasks = append(tasks, TaskSpec{
			ExpertID:           expertUUID,
			Title:              strings.TrimSpace(rt.Title),
			Description:        strings.TrimSpace(rt.Description),
			DependsOnExpertIDs: deps,
		})
	}

	if len(tasks) == 0 {
		return nil, fmt.Errorf("all %d tasks invalid after UUID validation", len(rawTasks))
	}
	return tasks, nil
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}
