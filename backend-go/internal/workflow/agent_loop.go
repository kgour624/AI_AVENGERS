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
	"ai_avengers/backend/internal/gateway"
)

// doneMarker is the exact string the LLM must output to signal task completion.
// WHY a marker not a tool call: simpler for the LLM to output, less hallucination risk.
// The LLM is instructed to output this as the LAST line of its response.
const doneMarker = "TASK_COMPLETE"

// AgentLoop runs one expert through the OTA (Observe-Think-Act) loop
// for a single workflow task.
//
// PATTERN: OTA Loop (Arpit Bhiyani AI Masterclass)
//   Observe: Read blackboard for prior expert artifacts
//   Think:   LLM call with task + blackboard context + tool definitions
//   Act:     Execute tool calls from LLM response
//   Repeat until: LLM outputs TASK_COMPLETE or max iterations reached
//
// WHY OTA not React:
//   React (Reason+Act) is for open-ended exploration where thought is precious.
//   OTA is for structured artifact production where the task is well-defined.
//   Each expert has a clear deliverable (architecture doc, class diagram, etc.).
//   OTA is simpler, cheaper, and sufficient for this use case.
//
// WHY per-expert loop not shared loop:
//   Each expert has independent knowledge corpus and independent task.
//   Shared loop would mix contexts and confuse the LLM.
//   Independent loops = clean separation of concerns.
type AgentLoop struct {
	db      *pgxpool.Pool
	tools   *Tools
	gateway *gateway.ModelGateway
	logger  *zap.Logger
}

// NewAgentLoop creates a new AgentLoop.
func NewAgentLoop(db *pgxpool.Pool, tools *Tools, gw *gateway.ModelGateway, logger *zap.Logger) *AgentLoop {
	return &AgentLoop{
		db:      db,
		tools:   tools,
		gateway: gw,
		logger:  logger,
	}
}

// AgentLoopRequest is the input to AgentLoop.Run.
type AgentLoopRequest struct {
	WorkflowID      uuid.UUID
	Expert          workflowExpert
	TaskID          uuid.UUID
	TaskTitle       string
	TaskDescription string
	// TokenCh: when non-nil, the last expert's final LLM response streams here.
	// nil for intermediate experts (their output goes to blackboard only).
	TokenCh chan<- string
}

// AgentLoopResult is the output of AgentLoop.Run.
type AgentLoopResult struct {
	ArtifactEventID *uuid.UUID // blackboard event ID of the produced artifact
	Iterations      int        // how many OTA iterations ran
	Completed       bool       // true if TASK_COMPLETE was reached
}

// Run executes the OTA loop for one expert's task.
//
// Mental execution:
//   Expert: System Design, Task: "Design URL shortener architecture"
//   workflowID: abc, taskID: xyz
//
//   Iteration 1:
//     Observe: ReadBlackboard(since=0) → [] (no prior events)
//     Think:   LLM(system=charter+task, user="Blackboard: none\nTask: Design...")
//     Act:     LLM outputs PostArtifact tool call
//              → PostArtifact executed → blackboard event created
//              → LLM also outputs TASK_COMPLETE → exit loop
//
//   Iteration 2 (if no TASK_COMPLETE):
//     Observe: ReadBlackboard(since=lastSeq) → new events from other experts
//     Think:   LLM with updated context
//     Act:     ...
func (a *AgentLoop) Run(ctx context.Context, req AgentLoopRequest) (*AgentLoopResult, error) {
	a.logger.Info("agent loop started",
		zap.String("workflow_id", req.WorkflowID.String()),
		zap.String("expert", req.Expert.Name),
		zap.String("task_id", req.TaskID.String()),
		zap.String("task", req.TaskTitle),
	)

	// Mark task as in_progress.
	a.updateTaskStatus(ctx, req.TaskID, "in_progress", nil)

	maxIter := req.Expert.MaxLoopIterations
	if maxIter <= 0 {
		maxIter = 5 // safe default
	}

	result := &AgentLoopResult{}
	var lastBlackboardSeq int64 = 0

	for iter := 1; iter <= maxIter; iter++ {
		a.logger.Debug("agent loop iteration",
			zap.String("expert", req.Expert.Name),
			zap.Int("iter", iter),
			zap.Int("max", maxIter),
		)

		// ============================================================
		// OBSERVE: Read blackboard for prior expert artifacts.
		// WHY since=lastBlackboardSeq: only read NEW events each iteration.
		// Avoids re-processing events already in the LLM context.
		// ============================================================
		priorArtifacts, err := a.tools.ReadBlackboard(ctx, ReadBlackboardRequest{
			WorkflowID: req.WorkflowID,
			Since:      lastBlackboardSeq,
		})
		if err != nil {
			a.logger.Warn("agent loop: ReadBlackboard failed",
				zap.String("expert", req.Expert.Name),
				zap.Error(err),
			)
			// Non-fatal: continue with empty context.
			priorArtifacts = nil
		}
		// Update cursor to latest seen sequence.
		for _, ev := range priorArtifacts {
			if ev.SequenceNumber > lastBlackboardSeq {
				lastBlackboardSeq = ev.SequenceNumber
			}
		}

		// ============================================================
		// THINK: LLM call with task + blackboard context.
		// ============================================================
		systemPrompt := a.buildSystemPrompt(req)
		userPrompt := a.buildUserPrompt(req, priorArtifacts, iter)

		var llmContent string
		var llmErr error

		// Stream only for the last expert's final iteration (when TokenCh is set).
		// All other calls are blocking.
		if req.TokenCh != nil && iter == maxIter {
			llmContent, llmErr = a.callStreaming(ctx, systemPrompt, userPrompt, req.Expert, req.TokenCh)
		} else {
			var resp *gateway.LLMResponse
			resp, llmErr = a.gateway.Call(ctx, gateway.LLMRequest{
				Model:        gateway.ModelStrong,
				SystemPrompt: systemPrompt,
				UserPrompt:   userPrompt,
				MaxTokens:    4000,
				Temperature:  req.Expert.loopTemperature(),
			})
			if resp != nil {
				llmContent = resp.Content
			}
		}
		if llmErr != nil {
			a.logger.Error("agent loop: LLM call failed",
				zap.String("expert", req.Expert.Name),
				zap.Int("iter", iter),
				zap.Error(llmErr),
			)
			// Fatal: can't think without LLM.
			a.updateTaskStatus(ctx, req.TaskID, "blocked", nil)
			return result, fmt.Errorf("agent loop: LLM failed on iter %d: %w", iter, llmErr)
		}
		result.Iterations = iter

		// ============================================================
		// ACT: Parse and execute tool calls from LLM response.
		// ============================================================
		artifactID, actErr := a.act(ctx, req, llmContent)
		if actErr != nil {
			a.logger.Warn("agent loop: act failed (non-fatal, continuing)",
				zap.String("expert", req.Expert.Name),
				zap.Int("iter", iter),
				zap.Error(actErr),
			)
			// Non-fatal: log and continue to next iteration.
		}
		if artifactID != nil {
			result.ArtifactEventID = artifactID
		}

		// Check for TASK_COMPLETE marker.
		if strings.Contains(llmContent, doneMarker) {
			a.logger.Info("agent loop: task complete",
				zap.String("expert", req.Expert.Name),
				zap.Int("iterations", iter),
			)
			result.Completed = true
			break
		}
	}

	// Mark task done regardless of whether TASK_COMPLETE was reached.
	// Max iterations = best effort — whatever was produced is the artifact.
	if !result.Completed {
		a.logger.Warn("agent loop: max iterations reached without TASK_COMPLETE",
			zap.String("expert", req.Expert.Name),
			zap.Int("max_iter", maxIter),
		)
	}
	a.updateTaskStatus(ctx, req.TaskID, "done", result.ArtifactEventID)

	a.logger.Info("agent loop finished",
		zap.String("expert", req.Expert.Name),
		zap.Int("iterations", result.Iterations),
		zap.Bool("completed", result.Completed),
	)
	return result, nil
}

// buildSystemPrompt constructs the LLM system prompt for this expert's task.
// Includes: expert charter, task description, available tools, done marker instruction.
func (a *AgentLoop) buildSystemPrompt(req AgentLoopRequest) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("You are %s, a domain expert in %s.\n\n", req.Expert.Name, req.Expert.Domain))
	sb.WriteString(fmt.Sprintf("REASONING CHARTER:\n%s\n\n", req.Expert.ReasoningCharter))
	sb.WriteString(fmt.Sprintf("YOUR TASK:\n%s\n\n", req.TaskDescription))

	sb.WriteString(`AVAILABLE TOOLS:
You have access to these tools. Call them by outputting a JSON block with this format:

<tool_call>
{"tool": "PostArtifact", "event_type": "<type>", "content": {<your artifact>}}
</tool_call>

Tool: PostArtifact
  Purpose: Publish your artifact to the shared blackboard so other experts can read it.
  event_type options:
    - "requirement_captured"    (for requirement analysis)
    - "architecture_decision"   (for system design decisions)
    - "data_model_proposed"     (for data models)
    - "api_contract_proposed"   (for API contracts)
    - "module_design_proposed"  (for module/class designs)
    - "code_artifact_produced"  (for code)
    - "test_case_proposed"      (for test cases)
    - "task_plan_ready"         (for task plans)
  content: your artifact as a JSON object

Tool: AskExpert
  Purpose: Ask another expert a specific question.
  Format: <tool_call>{"tool": "AskExpert", "to_expert_id": "<uuid>", "question": "<text>"}</tool_call>

WHEN YOU ARE DONE:
Output TASK_COMPLETE as the last line of your response.
Do NOT output TASK_COMPLETE until you have posted your artifact via PostArtifact.
`)

	return sb.String()
}

// buildUserPrompt constructs the user prompt for each OTA iteration.
// Includes: blackboard context (prior artifacts) + task reminder.
func (a *AgentLoop) buildUserPrompt(req AgentLoopRequest, priorArtifacts []blackboard.Event, iter int) string {
	var sb strings.Builder

	if len(priorArtifacts) == 0 {
		sb.WriteString("BLACKBOARD CONTEXT: No prior artifacts yet.\n\n")
	} else {
		sb.WriteString(fmt.Sprintf("BLACKBOARD CONTEXT (%d new artifacts since last iteration):\n", len(priorArtifacts)))
		for _, ev := range priorArtifacts {
			// Format: [event_type by expert_id]: content
			posterID := "client"
			if ev.PostedByExpertID != nil {
				posterID = ev.PostedByExpertID.String()
			}
			sb.WriteString(fmt.Sprintf("\n[%s by %s (seq=%d)]:\n%s\n",
				ev.EventType, posterID, ev.SequenceNumber, string(ev.Content)))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("ITERATION: %d\n", iter))
	sb.WriteString(fmt.Sprintf("YOUR TASK: %s\n\n", req.TaskTitle))
	sb.WriteString("Now produce your artifact. Use PostArtifact to publish it. Output TASK_COMPLETE when done.")

	return sb.String()
}

// act parses the LLM response and executes any tool calls found.
// Returns the artifact event ID if PostArtifact was called.
//
// Mental execution:
//   LLM output: "I will design the architecture...\n<tool_call>{...}</tool_call>\nTASK_COMPLETE"
//   → find <tool_call> blocks
//   → parse JSON inside each block
//   → execute PostArtifact or AskExpert
//   → return artifact event ID
func (a *AgentLoop) act(ctx context.Context, req AgentLoopRequest, llmContent string) (*uuid.UUID, error) {
	var artifactEventID *uuid.UUID

	// Extract all <tool_call>...</tool_call> blocks.
	blocks := extractToolCallBlocks(llmContent)
	if len(blocks) == 0 {
		// No tool calls — LLM is still thinking. Valid for intermediate iterations.
		return nil, nil
	}

	for _, block := range blocks {
		var call struct {
			Tool        string          `json:"tool"`
			EventType   string          `json:"event_type"`
			Content     json.RawMessage `json:"content"`
			ToExpertID  string          `json:"to_expert_id"`
			Question    string          `json:"question"`
		}
		if err := json.Unmarshal([]byte(block), &call); err != nil {
			a.logger.Warn("agent loop: malformed tool call JSON",
				zap.String("expert", req.Expert.Name),
				zap.String("block", block[:min(100, len(block))]),
				zap.Error(err),
			)
			continue
		}

		switch call.Tool {
		case ToolPostArtifact:
			if call.EventType == "" {
				a.logger.Warn("agent loop: PostArtifact missing event_type",
					zap.String("expert", req.Expert.Name),
				)
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
				a.logger.Error("agent loop: PostArtifact failed",
					zap.String("expert", req.Expert.Name),
					zap.Error(err),
				)
				return nil, fmt.Errorf("PostArtifact: %w", err)
			}
			artifactEventID = &event.ID
			a.logger.Info("agent loop: artifact posted",
				zap.String("expert", req.Expert.Name),
				zap.String("event_type", call.EventType),
				zap.Int64("seq", event.SequenceNumber),
			)

		case ToolAskExpert:
			if call.ToExpertID == "" || call.Question == "" {
				continue
			}
			toID, err := uuid.Parse(call.ToExpertID)
			if err != nil {
				continue
			}
			result, err := a.tools.AskExpert(ctx, AskExpertRequest{
				WorkflowID:   req.WorkflowID,
				FromExpertID: req.Expert.ID,
				ToExpertID:   toID,
				Question:     call.Question,
			})
			if err != nil {
				a.logger.Warn("agent loop: AskExpert failed",
					zap.String("expert", req.Expert.Name),
					zap.Error(err),
				)
			} else if result.TimedOut {
				a.logger.Warn("agent loop: AskExpert timed out",
					zap.String("expert", req.Expert.Name),
					zap.String("to", call.ToExpertID),
				)
			}
			// Answer (if any) will appear in next iteration's ReadBlackboard.
		}
	}

	return artifactEventID, nil
}

// callStreaming calls the LLM with streaming enabled.
// Used for the last expert's final response so the user sees tokens live.
func (a *AgentLoop) callStreaming(
	ctx context.Context,
	systemPrompt, userPrompt string,
	expert workflowExpert,
	tokenCh chan<- string,
) (string, error) {
	// Use blocking call — streaming wiring requires StreamCall which is
	// already implemented in chinawall/enforcer.go for the chat flow.
	// For workflow, blocking is acceptable: user sees Kanban updates, not tokens.
	// TODO: wire StreamCall here if real-time token streaming is needed for workflow.
	resp, err := a.gateway.Call(ctx, gateway.LLMRequest{
		Model:        gateway.ModelStrong,
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		MaxTokens:    4000,
		Temperature:  expert.loopTemperature(),
	})
	if err != nil {
		return "", err
	}
	// Send full content to tokenCh so frontend can display it.
	if tokenCh != nil {
		select {
		case tokenCh <- resp.Content:
		default:
		}
	}
	return resp.Content, nil
}

// updateTaskStatus updates workflow_tasks.status and optionally sets
// produced_artifact_event_id when the task produces an artifact.
func (a *AgentLoop) updateTaskStatus(ctx context.Context, taskID uuid.UUID, status string, artifactEventID *uuid.UUID) {
	var err error
	if artifactEventID != nil {
		_, err = a.db.Exec(ctx,
			`UPDATE workflow_tasks SET
				status = $1,
				produced_artifact_event_id = $2,
				completed_at = CASE WHEN $1 = 'done' THEN NOW() ELSE completed_at END,
				started_at = CASE WHEN $1 = 'in_progress' AND started_at IS NULL THEN NOW() ELSE started_at END,
				updated_at = NOW()
			 WHERE id = $3`,
			status, artifactEventID, taskID,
		)
	} else {
		_, err = a.db.Exec(ctx,
			`UPDATE workflow_tasks SET
				status = $1,
				completed_at = CASE WHEN $1 = 'done' THEN NOW() ELSE completed_at END,
				started_at = CASE WHEN $1 = 'in_progress' AND started_at IS NULL THEN NOW() ELSE started_at END,
				updated_at = NOW()
			 WHERE id = $2`,
			status, taskID,
		)
	}
	if err != nil {
		// Non-fatal: Kanban state is derived from blackboard anyway.
		a.logger.Warn("agent loop: updateTaskStatus failed",
			zap.String("task_id", taskID.String()),
			zap.String("status", status),
			zap.Error(err),
		)
	}
}

// extractToolCallBlocks finds all <tool_call>...</tool_call> blocks in text.
// Returns the JSON content inside each block.
//
// Mental execution:
//   Input: "I will design...\n<tool_call>{\"tool\": \"PostArtifact\", ...}</tool_call>\nTASK_COMPLETE"
//   Output: ["{\"tool\": \"PostArtifact\", ...}"]
func extractToolCallBlocks(text string) []string {
	const openTag = "<tool_call>"
	const closeTag = "</tool_call>"

	var blocks []string
	for {
		start := strings.Index(text, openTag)
		if start == -1 {
			break
		}
		end := strings.Index(text[start:], closeTag)
		if end == -1 {
			break
		}
		content := strings.TrimSpace(text[start+len(openTag) : start+end])
		if content != "" {
			blocks = append(blocks, content)
		}
		text = text[start+end+len(closeTag):]
	}
	return blocks
}

// loopTemperature returns the appropriate LLM temperature for this expert's loop pattern.
// react: 0.4 (more exploratory)
// ota/plan_execute: 0.2 (more deterministic)
func (e workflowExpert) loopTemperature() float64 {
	if e.LoopPattern == "react" {
		return 0.4
	}
	return 0.2
}
