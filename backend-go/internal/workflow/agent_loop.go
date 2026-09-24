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
	"ai_avengers/backend/internal/gateway"
	"ai_avengers/backend/internal/observability"
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
//   Observe: ReadBlackboard + 3-Gate knowledge access
//   Think:   LLM call with gate-controlled context
//   Act:     Execute tool calls: PostArtifact, AskExpert
//   Repeat until TASK_COMPLETE or max iterations
//
// 3-GATE KNOWLEDGE ACCESS (Human Brain Model):
//   Gate 1: Own training (70-80% knowledge) — generic BLOCKED if passes
//   Gate 2: Peer knowledge (team's work) — generic BLOCKED if covers task
//   Gate 3: Generic gap filling (20-30%) — only when Gates 1+2 fail
//   Generic claims saved to pending_experience for admin review.
//
// 2-PHASE BEHAVIOR:
//   Design phase (high_level_design): Gates 1+2+3 active
//   Implementation phase (implementation): Gate 1 only, China Wall strict
type AgentLoop struct {
	db             *pgxpool.Pool
	tools          *Tools
	store          *blackboard.Store
	gateway        *gateway.ModelGateway
	assembler      *appcontext.Assembler
	gateSystem     *GateSystem
	experienceBank *ExperienceBank
	logger         *zap.Logger
}

func NewAgentLoop(db *pgxpool.Pool, tools *Tools, store *blackboard.Store, gw *gateway.ModelGateway, assembler *appcontext.Assembler, logger *zap.Logger) *AgentLoop {
	var gs *GateSystem
	var eb *ExperienceBank
	if assembler != nil {
		gs = NewGateSystem(assembler, logger)
		eb = NewExperienceBank(db, logger)
	}
	return &AgentLoop{
		db: db, tools: tools, store: store, gateway: gw,
		assembler: assembler, gateSystem: gs, experienceBank: eb,
		logger: logger,
	}
}

type AgentLoopRequest struct {
	WorkflowID      uuid.UUID
	Expert          workflowExpert
	TaskID          uuid.UUID
	TaskTitle       string
	TaskDescription string
	// WorkflowPhase: current phase of the workflow.
	// Design phases (high_level_design, detailed_design): Gates 1+2+3 active.
	// Implementation phase: Gate 1 only (China Wall strict, no generic).
	WorkflowPhase string
	// AllExperts: all experts in this workflow.
	// Used by Gate 2 to poll peers for knowledge.
	AllExperts []workflowExpert
	// GenericAllowancePct: the client's generic ceiling (0-30) read from
	// workflows.generic_allowance_pct by the runner. 0 = trained + peer
	// knowledge only. Re-read by the runner on every phase attempt, so a
	// client raising it and re-running takes effect immediately.
	GenericAllowancePct float64
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
		observability.Global.IncAgentLoopIter()
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
		// GATE SYSTEM: 3-gate knowledge access (iter==1 only).
		// Design phases: Gates 1+2+3 active.
		// Implementation phase: Gate 1 only (no generic allowed).
		// ============================================================
		var gateResult *GateResult
		if iter == 1 && a.gateSystem != nil {
			isDesignPhase := req.WorkflowPhase == PhaseHighLevelDesign ||
				req.WorkflowPhase == PhaseDetailedDesign ||
				req.WorkflowPhase == ""

			if isDesignPhase {
				// Design phase: full 3-gate system
				gr, gateErr := a.gateSystem.RunGates(
					ctx, req.WorkflowID, req.Expert, req.TaskDescription, req.AllExperts,
					req.GenericAllowancePct,
				)
				if gateErr != nil {
					a.logger.Warn("agent loop: gate system failed (continuing)",
						zap.String("expert", req.Expert.Name),
						zap.Error(gateErr),
					)
				} else {
					gateResult = gr
				}
			} else {
				// Implementation phase: Gate 1 only, generic BLOCKED
				chunks, fetchErr := a.assembler.GetCourseChunksForWorkflow(
					ctx, req.Expert.ID, req.TaskDescription, 10,
				)
				if fetchErr == nil {
					gateResult = &GateResult{
						TrainingChunks: chunks,
						Gate1Passed:    len(chunks) > 0,
						GenericAllowed: false, // NEVER in implementation phase
					}
				}
				a.logger.Info("agent loop: implementation phase — Gate 1 only, generic blocked",
					zap.String("expert", req.Expert.Name),
				)
			}
		}

		// ============================================================
		// CONTEXT MANAGEMENT: Build context from gate result + blackboard
		// + project memory (A6: workflow experts previously never saw L1/L2).
		// ============================================================
		var contextText string
		if gateResult != nil {
			// Gate system ran: use gate-controlled context
			gateCtx := FormatGateContext(gateResult)
			blackboardCtx, bErr := a.buildBlackboardContext(ctx, req.WorkflowID, allArtifacts)
			if bErr != nil {
				blackboardCtx = formatArtifacts(allArtifacts)
			}
			contextText = gateCtx + "\n[BLACKBOARD — peers' work]\n" + blackboardCtx
		} else {
			// No gate system (assembler nil): fallback to blackboard only
			var bErr error
			contextText, bErr = a.buildBlackboardContext(ctx, req.WorkflowID, allArtifacts)
			if bErr != nil {
				contextText = formatArtifacts(allArtifacts)
			}
		}
		// Project memory: established decisions across this project's experts.
		// Non-fatal — empty/missing memory just means no extra section.
		if memText := a.loadProjectMemoryText(ctx, req); memText != "" {
			contextText = memText + "\n" + contextText
		}

		// ============================================================
		// THINK: LLM call.
		// ============================================================
		systemPrompt := buildAgentSystemPrompt(req)
		userPrompt := buildAgentUserPrompt(req, contextText, iter)

		llmResp, err := a.gateway.Call(ctx, gateway.LLMRequest{
			Model:        gateway.ModelStrong,
			WorkflowID:   &req.WorkflowID,
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

		// Measure sourcing on every iteration's output.
		//
		// WHY measured and not assumed: "every claim must cite its source" is
		// only a prompt instruction — nothing enforces it. Counting the
		// markers is the only way to know whether the trained material was
		// actually used, instead of eyeballing an artifact in the UI. A zero
		// count while generic is blocked is a real problem, so it is logged
		// at WARN.
		chunkCites := strings.Count(llmResp.Content, "[CHUNK_")
		peerCites := strings.Count(llmResp.Content, "[PEER:")
		genericTags := strings.Count(llmResp.Content, "[GENERIC]")
		notCovered := strings.Count(llmResp.Content, "[NOT_COVERED")
		sourcingFields := []zap.Field{
			zap.String("expert", req.Expert.Name),
			zap.Int("iteration", iter),
			zap.Int("chunk_citations", chunkCites),
			zap.Int("peer_citations", peerCites),
			zap.Int("generic_tags", genericTags),
			zap.Int("not_covered_tags", notCovered),
			zap.Float64("generic_allowance_pct", req.GenericAllowancePct),
		}
		if chunkCites == 0 && peerCites == 0 && req.GenericAllowancePct == 0 {
			a.logger.Warn("agent loop: output cites no trained or peer source", sourcingFields...)
		} else {
			a.logger.Info("agent loop: output sourcing", sourcingFields...)
		}

		if strings.Contains(llmResp.Content, doneMarker) {
			// Save generic claims to Experience Bank before exiting
			if a.experienceBank != nil && gateResult != nil && gateResult.GenericAllowed {
				claims := ExtractGenericClaims(llmResp.Content)
				if len(claims) > 0 {
					a.experienceBank.SaveGenericClaims(
						ctx, req.Expert.ID, req.WorkflowID,
						req.TaskTitle, req.Expert.Domain, claims,
					)
				}
			}
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

// buildBlackboardContext builds context from blackboard artifacts only.
// Summarizes old artifacts if count exceeds threshold.
// Renamed from buildContext — gate context is now handled by FormatGateContext.
// loadProjectMemoryText resolves the workflow's project and returns the
// Assembler's project-memory prompt block (or "" if unavailable).
//
// Non-fatal by design: a missing project_id, DB hiccup, or nil assembler
// must never abort a wave — experts just proceed without the extra section.
func (a *AgentLoop) loadProjectMemoryText(ctx context.Context, req AgentLoopRequest) string {
	if a == nil || a.assembler == nil || a.db == nil {
		return ""
	}
	var projectID uuid.UUID
	err := a.db.QueryRow(ctx,
		`SELECT project_id FROM workflows WHERE id = $1`,
		req.WorkflowID,
	).Scan(&projectID)
	if err != nil || projectID == uuid.Nil {
		return ""
	}
	query := req.TaskDescription
	if query == "" {
		query = req.TaskTitle
	}
	return a.assembler.GetProjectMemoryText(ctx, projectID, req.Expert.ID, query)
}

func (a *AgentLoop) buildBlackboardContext(ctx context.Context, workflowID uuid.UUID, artifacts []blackboard.Event) (string, error) {
	if len(artifacts) == 0 {
		return "No prior artifacts from peers.", nil
	}
	if len(artifacts) <= contextSummarizeThreshold {
		return formatArtifacts(artifacts), nil
	}
	splitAt := len(artifacts) - recentArtifactsToKeepRaw
	if splitAt < 0 {
		splitAt = 0
	}
	old := artifacts[:splitAt]
	recent := artifacts[splitAt:]
	oldText := formatArtifacts(old)
	summaryResp, err := a.gateway.Call(ctx, gateway.LLMRequest{
		Model:        gateway.ModelCheap,
		WorkflowID:   &workflowID,
		SystemPrompt: "Summarize these expert artifacts concisely. Preserve key decisions, data models, API contracts. Max 500 words.",
		UserPrompt:   oldText,
		MaxTokens:    700,
		Temperature:  0.1,
	})
	if err != nil {
		return "", fmt.Errorf("summarizer failed: %w", err)
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[SUMMARY OF %d OLDER ARTIFACTS]\n", len(old)))
	sb.WriteString(summaryResp.Content)
	sb.WriteString("\n\n")
	sb.WriteString(fmt.Sprintf("[RECENT %d ARTIFACTS]\n", len(recent)))
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
		// Take the first balanced {...} object inside the block instead of
		// requiring the block to be exactly one JSON value.
		//
		// WHY: LLM tool-call output is not deterministic. Observed in
		// production: a block whose object was followed by a stray ']',
		// which made json.Unmarshal fail with
		// "invalid character ']' after top-level value" and silently
		// dropped the expert's artifact. The same tolerance also handles a
		// block wrapped in an array ("[{...}]") and a block with trailing
		// prose after the object.
		jsonBlock := extractFirstJSONObject(block)
		if jsonBlock == "" {
			a.logger.Warn("agent loop: tool call has no JSON object",
				zap.String("block", truncate(block, 100)),
			)
			continue
		}
		if err := json.Unmarshal([]byte(jsonBlock), &call); err != nil {
			a.logger.Warn("agent loop: malformed tool call",
				zap.String("block", truncate(jsonBlock, 100)),
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
// APPLY_PRINCIPLES mode: training = principles to apply, not exact answers.
func buildAgentSystemPrompt(req AgentLoopRequest) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("You are %s, a domain expert in %s.\n\n", req.Expert.Name, req.Expert.Domain))
	sb.WriteString(fmt.Sprintf("REASONING CHARTER:\n%s\n\n", req.Expert.ReasoningCharter))
	sb.WriteString(fmt.Sprintf("YOUR TASK:\n%s\n\n", req.TaskDescription))
	sb.WriteString(phaseFocus(req.WorkflowPhase))
	sb.WriteString("HOW TO USE YOUR CONTEXT:\n")
	sb.WriteString("1. PROJECT MEMORY: Settled decisions for this project (if present). Do not contradict them silently.\n")
	sb.WriteString("2. TRAINING MATERIAL: Principles from your training. Apply them to the new problem.\n")
	sb.WriteString("   You do NOT need an exact match. 'consistent hashing' training applies to 'URL shortener'.\n")
	sb.WriteString("   Think like a senior engineer: use past experience on new problems.\n")
	sb.WriteString("3. BLACKBOARD: Work done by other experts. Build on it, don't repeat it.\n\n")
	sb.WriteString("AVAILABLE TOOLS:\n")
	sb.WriteString("<tool_call>{\"tool\": \"PostArtifact\", \"event_type\": \"<type>\", \"content\": {<artifact>}}</tool_call>\n")
	sb.WriteString("event_type options:\n")
	sb.WriteString("  architecture_decision | data_model_proposed | api_contract_proposed | module_design_proposed\n")
	sb.WriteString("  requirement_captured | test_case_proposed\n")
	sb.WriteString("  code_artifact_produced  <- SPECIAL: content MUST have {\"filename\": \"file.go\", \"code\": \"...\"}\n")
	sb.WriteString("    WHY: code artifacts are validated (compile + typecheck) before publishing.\n")
	sb.WriteString("    If validation fails after 3 fix attempts, artifact is BLOCKED.\n\n")
	sb.WriteString("<tool_call>{\"tool\": \"AskExpert\", \"to_expert_id\": \"<uuid>\", \"question\": \"<text>\"}</tool_call>\n\n")
	sb.WriteString("WHEN DONE: Output TASK_COMPLETE as the last line after posting your artifact.\n")
	return sb.String()
}

// phaseFocus returns the phase-specific section of the system prompt.
// Returns "" for phases that need no extra constraint.
//
// WHY this exists: runner.go deliberately executes the SAME planned task
// list once per phase (see its Step 8 comment — "each design phase runs the
// full wave set"). But WorkflowPhase was only used to decide which gates
// run; it never reached the prompt. So an expert received a byte-identical
// system prompt in high_level_design and in detailed_design, produced
// near-identical output, and because blackboard dedup_key is a hash of
// (event_type + poster + content), that second artifact was silently
// dropped as a duplicate. The phase's LLM call was paid for and produced
// nothing. Telling the expert which phase it is in is what makes the
// re-run purposeful instead of duplicate work.
//
// The per-phase artifact focus mirrors the workflow state machine in
// DOMAIN_EXPERT_COLLABORATION_DESIGN.md §9.
//
// Only design phases appear here: runner.go routes implementation and qa to
// AiderRunner (which builds its own prompt), so AgentLoop never sees them.
func phaseFocus(phase string) string {
	switch phase {
	case PhaseHighLevelDesign:
		return "CURRENT PHASE: high_level_design\n" +
			"Stay at the HIGH-LEVEL view: component boundaries, major technology\n" +
			"choices with their trade-offs, and how a request flows through the system.\n" +
			"Expected artifacts: architecture_decision, requirement_captured.\n" +
			"Do NOT write concrete schemas, endpoint signatures or code in this phase.\n\n"
	case PhaseDetailedDesign:
		return "CURRENT PHASE: detailed_design\n" +
			"The high-level architecture is already on the blackboard. Read it and go\n" +
			"one level deeper — do NOT restate it.\n" +
			"Expected artifacts: data_model_proposed, api_contract_proposed,\n" +
			"module_design_proposed, test_case_proposed.\n" +
			"Be concrete: table/field definitions, endpoint signatures with request and\n" +
			"response shapes, per-module responsibilities.\n\n"
	default:
		// intake / handoff / unset: no phase-specific constraint.
		return ""
	}
}

// buildAgentUserPrompt builds the user prompt for each OTA iteration.
func buildAgentUserPrompt(req AgentLoopRequest, contextText string, iter int) string {
	var sb strings.Builder
	// contextText already contains [TRAINING MATERIAL] + [BLACKBOARD] sections
	sb.WriteString(fmt.Sprintf("CONTEXT:\n%s\n\n", contextText))
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

// extractFirstJSONObject returns the first balanced {...} object found in s,
// or "" if there is none / it never closes.
//
// Brace counting is string-aware: braces inside JSON string literals are
// ignored (a code artifact's "code" field legitimately contains { and }),
// and a backslash escape does not terminate the string.
//
// Mental execution:
//   `{"a":1}]`              -> `{"a":1}`   (stray trailing bracket dropped)
//   `[{"a":1}]`             -> `{"a":1}`   (array wrapper skipped)
//   `{"code":"func(){}"}`   -> whole object (braces inside string ignored)
//   `{"s":"a\"}b"}`         -> whole object (escaped quote handled)
//   `{"a":1`                -> ""          (never balances)
//   `no json here`          -> ""
func extractFirstJSONObject(s string) string {
	start := strings.Index(s, "{")
	if start == -1 {
		return ""
	}
	depth := 0
	inString := false
	escaped := false
	for i := start; i < len(s); i++ {
		ch := s[i]
		if inString {
			switch {
			case escaped:
				escaped = false
			case ch == '\\':
				escaped = true
			case ch == '"':
				inString = false
			}
			continue
		}
		switch ch {
		case '"':
			inString = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return s[start : i+1]
			}
		}
	}
	return ""
}

// loopTemperature returns LLM temperature based on loop pattern.
func (e workflowExpert) loopTemperature() float64 {
	if e.LoopPattern == "react" {
		return 0.4
	}
	return 0.2
}
