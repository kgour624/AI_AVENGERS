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
// Planner decides who does what based on expert charters — no hardcoded domains.
type TaskSpec struct {
	ExpertID           uuid.UUID   // which expert owns this task
	Title              string      // short task title for Kanban card
	Description        string      // full task description for the expert's LLM prompt
	DependsOnExpertIDs []uuid.UUID // expert IDs whose tasks must complete before this one starts
}

// Planner uses a single LLM call to decompose a requirement into per-expert tasks.
//
// DESIGN DECISION: LLM-driven, not hardcoded.
//   Domain experts are not fixed — any combination is valid.
//   LLM reads each expert's reasoning_charter and decides:
//     - what task fits that expert's domain
//     - which tasks depend on which other tasks
//   No domain strings are hardcoded anywhere in this file.
//
// WHY one LLM call not N calls:
//   Planner needs a global view of all experts to assign tasks correctly.
//   N separate calls would not know about each other's assignments.
//   One call = one coherent plan.
//
// FAILURE POLICY:
//   JSON parse failure → return error → WorkflowRunner fails the workflow.
//   Empty task list → return error → WorkflowRunner fails the workflow.
//   Unknown expert_id in LLM output → skip that task, log warning.
//   Self-dependency in depends_on → silently removed.
type Planner struct {
	gateway *gateway.ModelGateway
	logger  *zap.Logger
}

// NewPlanner creates a new Planner.
func NewPlanner(gw *gateway.ModelGateway, logger *zap.Logger) *Planner {
	return &Planner{gateway: gw, logger: logger}
}

// workflowExpert is the expert data the Planner needs.
// Separate from orchestrator.expertRecord — workflow needs loop_pattern,
// max_loop_iterations, allowed_tools which the chat orchestrator doesn't use.
type workflowExpert struct {
	ID                uuid.UUID
	Name              string
	Domain            string
	ReasoningCharter  string
	LoopPattern       string // ota | react | plan_execute
	MaxLoopIterations int
	AllowedTools      []string
}

// Plan decomposes requirementText into one TaskSpec per expert.
//
// Mental execution:
//   Input: "Build a URL shortener", experts=[SystemDesign, DSA, LLD]
//
//   System prompt: "You are a workflow planner..."
//   User prompt:   "Experts:\n- SystemDesign (charter: ...)\n...
//                  Requirement: Build a URL shortener"
//
//   LLM output (JSON):
//   [
//     {"expert_id": "uuid1", "title": "Design architecture",
//      "description": "...", "depends_on_expert_ids": []},
//     {"expert_id": "uuid2", "title": "Design hash function",
//      "description": "...", "depends_on_expert_ids": ["uuid1"]},
//     {"expert_id": "uuid3", "title": "Design class diagram",
//      "description": "...", "depends_on_expert_ids": ["uuid1", "uuid2"]}
//   ]
//
//   Validate: all expert_ids exist in input list
//   Remove: self-dependencies
//   Return: []TaskSpec
func (p *Planner) Plan(ctx context.Context, requirementText string, experts []workflowExpert) ([]TaskSpec, error) {
	if len(experts) == 0 {
		return nil, fmt.Errorf("planner: no experts provided")
	}
	if strings.TrimSpace(requirementText) == "" {
		return nil, fmt.Errorf("planner: requirement text is empty")
	}

	// Build expert index for validation.
	// Key: expert_id string → workflowExpert
	// WHY map not slice: O(1) lookup when validating LLM output.
	expertIndex := make(map[string]workflowExpert, len(experts))
	for _, e := range experts {
		expertIndex[e.ID.String()] = e
	}

	// Build the user prompt: expert list + requirement.
	var expertList strings.Builder
	for _, e := range experts {
		expertList.WriteString(fmt.Sprintf(
			"- Expert ID: %s\n  Name: %s\n  Domain: %s\n  Charter: %s\n\n",
			e.ID.String(), e.Name, e.Domain, e.ReasoningCharter,
		))
	}

	userPrompt := fmt.Sprintf(
		"EXPERTS:\n%s\nREQUIREMENT:\n%s",
		expertList.String(),
		requirementText,
	)

	resp, err := p.gateway.Call(ctx, gateway.LLMRequest{
		Model: gateway.ModelStrong,
		SystemPrompt: `You are a workflow planner for a multi-agent software development system.

Given a list of domain experts and a requirement, create exactly one task per expert.
Each task must:
1. Be specific to that expert's domain and charter
2. Describe what artifact the expert should produce
3. List which other expert IDs this task depends on (whose output must be read first)

Rules:
- Every expert in the list MUST get exactly one task
- depends_on_expert_ids must only contain IDs from the provided expert list
- An expert cannot depend on itself
- Tasks with no dependencies run first (they read only the requirement)
- Tasks with dependencies run after their dependencies complete

Output ONLY a valid JSON array. No explanation, no markdown, no code fences.
Format:
[
  {
    "expert_id": "<uuid>",
    "title": "<short task title, max 100 chars>",
    "description": "<detailed description of what to produce, min 50 chars>",
    "depends_on_expert_ids": ["<uuid>", ...]
  }
]`,
		UserPrompt:  userPrompt,
		MaxTokens:   2000,
		Temperature: 0.2, // low temperature: deterministic planning
	})
	if err != nil {
		return nil, fmt.Errorf("planner: LLM call failed: %w", err)
	}

	// Parse LLM output.
	// LLM may wrap JSON in markdown fences despite instructions — strip them.
	raw := strings.TrimSpace(resp.Content)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	// Find JSON array boundaries defensively.
	// WHY: LLM sometimes prepends a sentence before the JSON.
	start := strings.Index(raw, "[")
	end := strings.LastIndex(raw, "]")
	if start == -1 || end == -1 || end <= start {
		return nil, fmt.Errorf("planner: LLM output contains no JSON array: %q", raw[:min(200, len(raw))])
	}
	raw = raw[start : end+1]

	// Unmarshal into raw structs first — validate fields before building TaskSpec.
	type rawTask struct {
		ExpertID           string   `json:"expert_id"`
		Title              string   `json:"title"`
		Description        string   `json:"description"`
		DependsOnExpertIDs []string `json:"depends_on_expert_ids"`
	}
	var rawTasks []rawTask
	if err := json.Unmarshal([]byte(raw), &rawTasks); err != nil {
		return nil, fmt.Errorf("planner: JSON parse failed: %w (raw: %q)", err, raw[:min(200, len(raw))])
	}
	if len(rawTasks) == 0 {
		return nil, fmt.Errorf("planner: LLM produced empty task list")
	}

	// Validate and convert to TaskSpec.
	// Skip tasks with unknown expert_ids (LLM hallucination guard).
	var tasks []TaskSpec
	for _, rt := range rawTasks {
		// Validate expert_id exists in input list.
		if _, ok := expertIndex[rt.ExpertID]; !ok {
			p.logger.Warn("planner: unknown expert_id in LLM output — skipping task",
				zap.String("expert_id", rt.ExpertID),
			)
			continue
		}
		if strings.TrimSpace(rt.Title) == "" {
			p.logger.Warn("planner: task has empty title — skipping",
				zap.String("expert_id", rt.ExpertID),
			)
			continue
		}

		expertUUID, _ := uuid.Parse(rt.ExpertID)

		// Validate and deduplicate depends_on_expert_ids.
		// Remove: unknown IDs, self-references, duplicates.
		seen := make(map[string]bool)
		var deps []uuid.UUID
		for _, depIDStr := range rt.DependsOnExpertIDs {
			if depIDStr == rt.ExpertID {
				// Self-dependency: silently remove.
				continue
			}
			if _, ok := expertIndex[depIDStr]; !ok {
				p.logger.Warn("planner: unknown dependency expert_id — removing",
					zap.String("task_expert", rt.ExpertID),
					zap.String("dep_expert", depIDStr),
				)
				continue
			}
			if seen[depIDStr] {
				continue // duplicate
			}
			seen[depIDStr] = true
			depUUID, _ := uuid.Parse(depIDStr)
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
		return nil, fmt.Errorf("planner: all tasks were invalid after validation")
	}

	p.logger.Info("planner: task plan created",
		zap.Int("expert_count", len(experts)),
		zap.Int("task_count", len(tasks)),
	)
	return tasks, nil
}

// min returns the smaller of two ints.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
