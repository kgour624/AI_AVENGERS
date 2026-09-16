package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/blackboard"
	appcontext "ai_avengers/backend/internal/context"
	"ai_avengers/backend/internal/chinawall"
	"ai_avengers/backend/internal/gateway"
)

// doneMarker: LLM outputs this as last line to signal task completion.
const doneMarker = "TASK_COMPLETE"

// contextSummarizeThreshold: when blackboard has more than this many
// artifacts, summarize old ones to prevent context explosion.
// WHY 10: 10 artifacts * ~2k tokens each = 20k tokens. Above this,
// cost and context limits become a real problem.
const contextSummarizeThreshold = 10

// recentArtifactsToKeepRaw: always keep this many recent artifacts
// in full detail (not summarized). Most recent = most relevant.
const recentArtifactsToKeepRaw = 5

// AgentLoop runs one expert through the OTA (Observe-Think-Act) loop.
//
// PATTERN: OTA Loop (Arpit Bhiyani AI Masterclass)
//   Observe: ReadBlackboard(since=lastSeq) + summarize if needed
//   Think:   LLM call with charter + task + context
//   Act:     Execute tool calls: PostArtifact, AskExpert
//   Repeat until TASK_COMPLETE or max iterations
//
// STATUS UPDATES: via blackboard events (not direct DB writes).
//   PostTaskStatus() -> task_status_changed event -> Projector -> workflow_tasks
//   PostTaskFailed() -> task_failed event -> Projector -> workflow_tasks
//
// CONTEXT MANAGEMENT (Fix 4):
//   > contextSummarizeThreshold artifacts -> summarize old ones
//   Keep recentArtifactsToKeepRaw in full detail
//   Prevents 40k+ token context explosion
type AgentLoop struct {
	db      *pgxpool.Pool
	tools   *Tools
	store   *blackboard.Store
	gateway *gateway.ModelGateway
	logger  *zap.Logger
}

func NewAgentLoop(db *pgxpool.Pool, tools *Tools, store *blackboard.Store, gw *gateway.ModelGateway, logger *zap.Logger) *AgentLoop {
	return &AgentLoop{db: db, tools: tools, store: store, gateway: gw, logger: logger}
}

type AgentLoopRequest struct {
	WorkflowID      uuid.UUID
	Expert          workflowExpert
	TaskID          uuid.UUID
	TaskTitle       string
	TaskDescription string
}

type AgentLoopResult struct {
	ArtifactEventID *uuid.UUID
	Iterations      int
	Completed       bool
}

// Run executes the OTA loop for one expert's task.
//
// Mental execution:
//   Expert: System Design, Task: "Design URL shortener architecture"
//
//   Iteration 1:
//     Observe: ReadBlackboard(since=0) -> [requirement_captured event]
//     Context: 1 artifact, no summarization needed
//     Think:   LLM(system=charter+task, user="Context: [req]\nTask: Design...")
//     Act:     LLM outputs <tool_call>PostArtifact{architecture_decision}</tool_call>
//              -> PostArtifact -> blackboard event
//              -> PostTaskStatus(done) -> task_status_changed event
//              -> LLM outputs TASK_COMPLETE -> exit
//
//   Iteration 2 (if no TASK_COMPLETE):
//     Observe: ReadBlackboard(since=lastSeq) -> new events
//     Context: if > 10 artifacts -> summarize old ones
//     Think:   LLM with updated context
//     Act:     ...
func (a *AgentLoop) Run(ctx context.Context, req AgentLoopRequest) (*AgentLoopResult, error) {
	a.logger.Info("agent loop started",
		zap.String("workflow_id", req.WorkflowID.String()),
		zap.String("expert", req.Expert.Name),
		zap.String("task", req.TaskTitle),
	)

	// Signal task start via blackboard event (not direct DB write).
	// Projector will pick this up and update workflow_tasks.
	if err := PostTaskStatus(ctx, a.store, req.WorkflowID, req.Expert.ID, "in_progress"); err != nil {
		a.logger.Warn("agent loop: PostTaskStatus in_progress failed", zap.Error(err))
		// Non-fatal: continue anyway.
	}

	maxIter := req.Expert.MaxLoopIterations
	if maxIter <= 0 {
		maxIter = 5
	}

	result := &AgentLoopResult{}
	var lastBlackboardSeq int64 = 0
	// allArtifacts accumulates ALL artifacts seen across iterations.
	// Used for context summarization when count exceeds threshold.
	var allArtifacts []blackboard.Event

	for iter := 1; iter <= maxIter; iter++ {
		a.logger.Debug("agent loop iteration",
			zap.String("expert", req.Expert.Name),
			zap.Int("iter", iter),
		)

		// ============================================================
		// OBSERVE: Read new blackboard events since last cursor.
		// ============================================================
		newArtifacts, err := a.tools.ReadBlackboard(ctx, ReadBlackboardRequest{
			WorkflowID: req.WorkflowID,
			Since:      lastBlackboardSeq,
		})
		if err != nil {
			a.logger.Warn("agent loop: ReadBlackboard failed (continuing)", zap.Error(err))
		} else {
			allArtifacts = append(allArtifacts, newArtifacts...)
			for _, ev := range newArtifacts {
				if ev.SequenceNumber > lastBlackboardSeq {
					lastBlackboardSeq = ev.SequenceNumber
				}
			}
		}

		// ============================================================
		// CONTEXT MANAGEMENT: Summarize if too many artifacts.
		// WHY: 10+ artifacts * ~2k tokens = 20k+ tokens -> cost + limit.
		// Fix: summarize old artifacts, keep recent ones in full.
		// ============================================================
		contextText, err := a.buildContext(ctx, allArtifacts)
		if err != nil {
			a.logger.Warn("agent loop: buildContext failed, using raw", zap.Error(err))
			contextText = formatArtifacts(allArtifacts)
		}

		// ============================================================
		// THINK: LLM call.
		// ============================================================
		systemPrompt := buildAgentSystemPrompt(req)
		userPrompt := buildAgentUserPrompt(req, contextText, iter)

		llmResp, err := a.gateway.Call(ctx, gateway.LLMRequest{
			Model:        gateway.ModelStrong,
			SystemPrompt: systemPrompt,
			UserPrompt:   userPrompt,
			MaxTokens:    4000,
			Temperature:  req.Expert.loopTemperature(),
		})
		if err != nil {
			a.logger.Error("agent loop: LLM call failed",
				zap.String("expert", req.Expert.Name),
				zap.Int("iter", iter),
				zap.Error(err),
			)
			// Fatal: post task_failed event and return.
			_ = PostTaskFailed(ctx, a.store, req.WorkflowID, req.Expert.ID,
				fmt.Sprintf("LLM failed on iter %d: %s", iter, err.Error()))
			return result, fmt.Errorf("agent loop: LLM failed: %w", err)
		}
		result.Iterations = iter

		// ============================================================
		// ACT: Execute tool calls from LLM response.
		// ============================================================
		artifactID, actErr := a.act(ctx, req, llmResp.Content)
		if actErr != nil {
			a.logger.Warn("agent loop: act failed (non-fatal)",
				zap.String("expert", req.Expert.Name),
				zap.Error(actErr),
			)
		}
		if artifactID != nil {
			result.ArtifactEventID = artifactID
		}

		if strings.Contains(llmResp.Content, doneMarker) {
			result.Completed = true
			break
		}
	}

	if !result.Completed {
		a.logger.Warn("agent loop: max iterations reached",
			zap.String("expert", req.Expert.Name),
			zap.Int("max", maxIter),
		)
	}

	// Signal task done via blackboard event.
	_ = PostTaskStatus(ctx, a.store, req.WorkflowID, req.Expert.ID, "done")

	a.logger.Info("agent loop finished",
		zap.String("expert", req.Expert.Name),
		zap.Int("iterations", result.Iterations),
		zap.Bool("completed", result.Completed),
	)
	return result, nil
}

// buildContext builds the context string for the LLM.
// If artifacts > contextSummarizeThreshold, summarizes old ones.
//
// Mental execution:
//   15 artifacts total
//   old = artifacts[0:10], recent = artifacts[10:15]
//   summary = LLM("Summarize: [old 10 artifacts]")
//   return "[SUMMARY]\n" + summary + "\n\n[RECENT 5]\n" + format(recent)
func (a *AgentLoop) buildContext(ctx context.Context, artifacts []blackboard.Event) (string, error) {
	if len(artifacts) <= contextSummarizeThreshold {
		return formatArtifacts(artifacts), nil
	}

	// Split: old (to summarize) + recent (keep raw).
	splitAt := len(artifacts) - recentArtifactsToKeepRaw
	if splitAt < 0 {
		splitAt = 0
	}
	old := artifacts[:splitAt]
	recent := artifacts[splitAt:]

	// Summarize old artifacts.
	oldText := formatArtifacts(old)
	summaryResp, err := a.gateway.Call(ctx, gateway.LLMRequest{
		Model: gateway.ModelCheap,
		SystemPrompt: "You are a technical summarizer. Summarize the following expert artifacts concisely. Preserve key decisions, data models, API contracts, and design choices. Max 500 words.",
		UserPrompt:   oldText,
		MaxTokens:    700,
		Temperature:  0.1,
	})
	if err != nil {
		return "", fmt.Errorf("summarizer LLM failed: %w", err)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[SUMMARY OF %d PRIOR ARTIFACTS]\n", len(old)))
	sb.WriteString(summaryResp.Content)
	sb.WriteString("\n\n")
	sb.WriteString(fmt.Sprintf("[RECENT %d ARTIFACTS (full detail)]\n", len(recent)))
	sb.WriteString(formatArtifacts(recent))
	return sb.String(), nil
}

// formatArtifacts formats a slice of blackboard events as readable text.
func formatArtifacts(artifacts []blackboard.Event) string {
	if len(artifacts) == 0 {
		return "No prior artifacts."
	}
	var sb strings.Builder
	for _, ev := range artifacts {
		posterID := "system"
		if ev.PostedByExpertID != nil {
			posterID = ev.PostedByExpertID.String()
		}
		sb.WriteString(fmt.Sprintf("[%s by %s (seq=%d)]:\n%s\n\n",
			ev.EventType, posterID, ev.SequenceNumber, string(ev.Content)))
	}
	return sb.String()
}

// act parses LLM response and executes tool calls.
func (a *AgentLoop) act(ctx context.Context, req AgentLoopRequest, llmContent string) (*uuid.UUID, error) {
	blocks := extractToolCallBlocks(llmContent)
	if len(blocks) == 0 {
		return nil, nil
	}

	var artifactEventID *uuid.UUID
	for _, block := range blocks {
		var call struct {
			Tool       string          `json:"tool"`
			EventType  string          `json:"event_type"`
			Content    json.RawMessage `json:"content"`
			ToExpertID string          `json:"to_expert_id"`
			Question   string          `json:"question"`
		}
		if err := json.Unmarshal([]byte(block), &call); err != nil {
			a.logger.Warn("agent loop: malformed tool call",
				zap.String("block", truncate(block, 100)),
				zap.Error(err),
			)
			continue
		}

		switch call.Tool {
		case ToolPostArtifact:
			if call.EventType == "" {
				continue
			}
			var contentObj interface{}
			if len(call.Content) > 0 {
				_ = json.Unmarshal(call.Content, &contentObj)
			} else {
				contentObj = map[string]string{"text": llmContent}
			}
			event, err := a.tools.PostArtifact(ctx, PostArtifactRequest{
				WorkflowID:       req.WorkflowID,
				PostedByExpertID: req.Expert.ID,
				EventType:        call.EventType,
				Content:          contentObj,
			})
			if err != nil {
				return nil, fmt.Errorf("PostArtifact: %w", err)
			}
			artifactEventID = &event.ID

		case ToolAskExpert:
			if call.ToExpertID == "" || call.Question == "" {
				continue
			}
			toID, err := uuid.Parse(call.ToExpertID)
			if err != nil {
				continue
			}
			_, _ = a.tools.AskExpert(ctx, AskExpertRequest{
				WorkflowID:   req.WorkflowID,
				FromExpertID: req.Expert.ID,
				ToExpertID:   toID,
				Question:     call.Question,
			})
		}
	}
	return artifactEventID, nil
}

// buildAgentSystemPrompt builds the LLM system prompt for this expert's task.
func buildAgentSystemPrompt(req AgentLoopRequest) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("You are %s, a domain expert in %s.\n\n", req.Expert.Name, req.Expert.Domain))
	sb.WriteString(fmt.Sprintf("REASONING CHARTER:\n%s\n\n", req.Expert.ReasoningCharter))
	sb.WriteString(fmt.Sprintf("YOUR TASK:\n%s\n\n", req.TaskDescription))
	sb.WriteString(`AVAILABLE TOOLS:
Call tools by outputting a JSON block:

<tool_call>
{"tool": "PostArtifact", "event_type": "<type>", "content": {<artifact>}}
</tool_call>

Tool: PostArtifact
  event_type: architecture_decision | data_model_proposed | api_contract_proposed |
              module_design_proposed | code_artifact_produced | test_case_proposed |
              requirement_captured

Tool: AskExpert
  <tool_call>{"tool": "AskExpert", "to_expert_id": "<uuid>", "question": "<text>"}</tool_call>

WHEN DONE: Output TASK_COMPLETE as the last line.
Do NOT output TASK_COMPLETE until you have posted your artifact via PostArtifact.
`)
	return sb.String()
}

// buildAgentUserPrompt builds the user prompt for each OTA iteration.
func buildAgentUserPrompt(req AgentLoopRequest, contextText string, iter int) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("BLACKBOARD CONTEXT:\n%s\n\n", contextText))
	sb.WriteString(fmt.Sprintf("ITERATION: %d\n", iter))
	sb.WriteString(fmt.Sprintf("YOUR TASK: %s\n\n", req.TaskTitle))
	sb.WriteString("Produce your artifact. Use PostArtifact to publish it. Output TASK_COMPLETE when done.")
	return sb.String()
}

// extractToolCallBlocks finds all <tool_call>...</tool_call> blocks.
func extractToolCallBlocks(text string) []string {
	const open = "<tool_call>"
	const close = "</tool_call>"
	var blocks []string
	for {
		start := strings.Index(text, open)
		if start == -1 {
			break
		}
		end := strings.Index(text[start:], close)
		if end == -1 {
			break
		}
		content := strings.TrimSpace(text[start+len(open) : start+end])
		if content != "" {
			blocks = append(blocks, content)
		}
		text = text[start+end+len(close):]
	}
	return blocks
}

// loopTemperature returns LLM temperature based on loop pattern.
func (e workflowExpert) loopTemperature() float64 {
	if e.LoopPattern == "react" {
		return 0.4
	}
	return 0.2
}
