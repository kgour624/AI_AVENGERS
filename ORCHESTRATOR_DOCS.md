# Orchestrator.go & Collaborative Mode - Deep Dive (Software Engineer's Guide)

## Part 1: orchestrator.go

**File Path:** `backend-go/internal/orchestrator/orchestrator.go`

This file serves as the **Core Controller (The Brain)** of the entire chat system. Every incoming user message routes through this file first.

### 1. Core Structs (Data Models)

```go
type OrchestratorRequest struct {
	ChatID              uuid.UUID
	Message             string
	ExpertIDs           []uuid.UUID
	GenericAllowancePct float64
}

type OrchestratorResponse struct {
	ExpertResponses []ExpertResponse
	Synthesis       *SynthesisResult
	CollabSections  []collab.Section
}
```

### 2. `Process` (Independent Mode - Parallel Execution)

Invoked when a user selects multiple experts in "Independent" mode.

- **Goroutine Fan-out**: Instead of running sequentially, it utilizes `go func()` to execute all experts **in parallel**.

```go
func (o *Orchestrator) Process(ctx context.Context, req OrchestratorRequest) (*OrchestratorResponse, error) {
	// ... load experts ...
	resultCh := make(chan ExpertResponse, len(experts))
	var wg sync.WaitGroup

	for _, exp := range experts {
		wg.Add(1)
		go func(expert expertRecord) {
			defer wg.Done()
			// Peer context is "" because independent mode isolates experts
			result := o.processWithExpert(ctx, req, expert, "")
			resultCh <- result
		}(exp)
	}

	wg.Wait()
	close(resultCh)
    // ... timeout logic & synthesize() ...
}
```

### 3. `ProcessCollaborative` (Collaborative Mode - Sequential Execution)

- **Sequential Execution**: Deliberately avoids parallel goroutines.

```go
func (o *Orchestrator) ProcessCollaborative(...) (*OrchestratorResponse, error) {
    // 1. Plan the order
	plan := collab.PlanSections(ctx, o.gw, req.Message, infos, o.logger)

    // 2. Define how an expert runs
	runExpert := func(runCtx context.Context, expertID uuid.UUID, peerContext string) (string, error) {
		expert := byID[expertID]
		resp := o.processWithExpert(runCtx, req, expert, peerContext)
		return resp.Content, nil
	}

    // 3. Execute sequentially
	relayResult := collab.RunRelay(ctx, plan, runExpert, onProgress)
    // ...
}
```

### 4. `processWithExpert` (The Core Pipeline)

Regardless of the mode, every expert is routed through this exact pipeline:

```go
func (o *Orchestrator) processWithExpert(ctx context.Context, req OrchestratorRequest, expert expertRecord, peerContext string) ExpertResponse {
	// 1. Concurrency Control (Semaphore)
	sem := o.getExpertSemaphore(expert.ID.String())
	select {
	case sem <- struct{}{}: defer func() { <-sem }()
	default: return ExpertResponse{Error: "concurrency_limit_exceeded"} // Backpressure
	}

	// 2. Context Assembly (RAG)
	assembledCtx, _ := o.assembler.Assemble(ctx, req.ChatID, req.ProjectID, expert.ID, req.Message, ...)

	// 3. Self-Learning (Noise Reduction)
	questionForRAG := req.Message
	if o.selfLearning != nil {
		processed := o.selfLearning.Process(ctx, req.Message, expert.Name, ...)
		if processed.VerificationPassed {
			questionForRAG = processed.Extracted
		}
	}

	// 4. Decision Engine (China Wall Gates)
	result, _ := o.decisionEng.Process(ctx, questionForRAG, expert, assembledCtx.CourseChunks, ...)

	return ExpertResponse{ /* mapped data */ }
}
```

---

## Part 2: plan.go

**File Path:** `backend-go/internal/collab/plan.go`

This file is the critical first step of **Collaborative Mode**. It dictates **which expert speaks first and what their section title should be.**

### 1. Structs (Data Models)

```go
type Section struct {
	ExpertID     uuid.UUID `json:"expert_id"`
	ExpertName   string    `json:"expert_name"`
	SectionTitle string    `json:"section_title"`
	Content      string    `json:"content,omitempty"`
}
```

### 2. `PlanSections` (LLM-Based Dynamic Planning)

```go
func PlanSections(ctx context.Context, gw *gateway.ModelGateway, question string, experts []ExpertInfo, logger *zap.Logger) []Section {
	fallback := FallbackPlan(experts)
	if gw == nil { return fallback }

    // LLM call asking for JSON ordering
	resp, err := gw.Call(ctx, gateway.LLMRequest{
		Model:       gateway.ModelCheap,
		UserPrompt:  "Decide the best section order and a short section title...",
		Temperature: 0.1,
	})

	if err != nil || parseFails(resp) {
		return fallback // Fail-Open Design
	}
    // ... map JSON to []Section ...
}
```

### 3. `FallbackPlan` (Fail-Safe)

```go
func FallbackPlan(experts []ExpertInfo) []Section {
	var sections []Section
	for _, e := range experts {
		title := e.CategoryName
		if title == "" { title = e.Domain }
		sections = append(sections, Section{
			ExpertID: e.ID, ExpertName: e.Name, SectionTitle: title,
		})
	}
	return sections
}
```

---

## Part 3: relay.go

**File Path:** `backend-go/internal/collab/relay.go`

The **Execution Engine** of Collaborative Mode. Runs experts in strict sequence.

### 1. `RunRelay` (The Core Sequential Loop)

```go
type RunExpert func(ctx context.Context, expertID uuid.UUID, peerContext string) (content string, err error)

func RunRelay(ctx context.Context, plan []Section, run RunExpert, onProgress ProgressFunc) RelayResult {
	result := RelayResult{Sections: make([]Section, 0, len(plan))}
	var peerContext strings.Builder

	for i, section := range plan {
        // Run expert with accumulated prior context
		content, err := run(ctx, section.ExpertID, formatPeerContext(peerContext.String()))

		if err != nil { // Fail-Soft Posture
			section.Content = "This expert could not answer: " + err.Error()
			result.Sections = append(result.Sections, section)
			continue
		}

		section.Content = content
		result.Sections = append(result.Sections, section)

		// Append to peerContext for the NEXT expert (truncated to save tokens)
		peerContent := content
		if len(peerContent) > 1500 { peerContent = peerContent[:1500] + "\n... [truncated]" }
		fmt.Fprintf(&peerContext, "--- %s already answered ---\n%s\n\n", section.ExpertName, peerContent)
	}

	return result
}
```

---

## Part 4: assemble.go

**File Path:** `backend-go/internal/collab/assemble.go`

The **final presentation layer**. Deterministic Markdown builder. No LLM used here.

```go
func AssembleMarkdown(sections []Section) string {
	var sb strings.Builder
	for i, section := range sections {
		if i > 0 { sb.WriteString("\n\n") }

		// Format: ## Section Title (Expert Name)
		sb.WriteString("## ")
		sb.WriteString(section.SectionTitle)
		sb.WriteString(" (")
		sb.WriteString(section.ExpertName)
		sb.WriteString(")\n\n")

		// Append content
		sb.WriteString(strings.TrimSpace(section.Content))
	}
	return sb.String()
}
```

---

## Part 5: enforcer.go (The China Wall)

**File Path:** `backend-go/internal/chinawall/enforcer.go`

The **Security Guard** of the application.

### 1. `Enforce` (The 4-Layer Pipeline)

````go
func (e *Enforcer) Enforce(ctx context.Context, question string, chunks []CourseChunk, expertDomain string, ...) (*EnforceResult, error) {
	profile := e.registry.Get(expertDomain)

	// LAYER 1: Reranker Threshold (Fast Reject)
	if float64(bestScore) < threshold && !allowGeneric {
		return e.buildRefusal("insufficient_relevance", "Score below threshold"), nil
	}

	// LAYER 2: Coverage Check (LLM Judge)
	coverage, _ := e.checkCoverage(ctx, question, chunks, profile.CoverageMode)
	if coverage == "NO" && !allowGeneric {
		return e.buildRefusal("not_covered", "Content does not cover this question"), nil
	}

	// LAYER 3: Generative Citation Enforcement
	generated, _ := e.generateWithCitations(ctx, question, chunks, profile, ...)
	if len(generated.Citations) == 0 {
		return &EnforceResult{Status: "retry", LayerFailed: 3}, nil // Force retry
	}

	// LAYER 4: Uncited Claim Stripping (Regex)
	cleanAnswer, _ := e.stripUncited(generated.Answer, profile.StripMode)
	cleanAnswer = e.cleanChunkIDs(cleanAnswer) // Security Cleansing

	// Conflict Resolution / Base Wall Safety Net
	if strings.TrimSpace(cleanAnswer) == "" {
		if profile.StripMode == StripModeCodeExempt && strings.Contains(generated.Answer, "```") {
			cleanAnswer = generated.Answer // Keep code blocks
		} else {
			// Domain rules stripped everything -> fallback to BaseProfile
			baseGenerated, _ := e.generateWithCitations(ctx, question, chunks, BaseProfile, ...)
			cleanAnswer, _ = e.stripUncited(baseGenerated.Answer, BaseProfile.StripMode)
		}
	}

	// Apply Quality Gate (LLM-as-a-Judge)
	ans, cites, qScore := e.applyQualityGate(ctx, question, chunks, profile, cleanAnswer, generated.Citations)

	return &EnforceResult{Status: "success", Answer: ans, Citations: cites, Coverage: coverage}, nil
}
````

---

## Part 6: processor.go (Self-Learning Mode)

**File Path:** `backend-go/internal/selflearning/processor.go`

This file is responsible for **Noise Reduction and Query Optimization**. Users often ask questions wrapped in stories (e.g., _"Alice is on a chessboard and the knight moves..."_). Standard RAG vectors search for "Alice" and "chessboard", returning garbage. This processor translates the story into technical domain-signal (_"BFS shortest path on grid"_) **before** RAG search begins.

### 1. `Process` (The 3-Step Pipeline)

The pipeline is designed to be **Fail-Closed**: if any step fails or drops constraints, it safely aborts and uses the original user question. It never crashes the request.

```go
func (p *QuestionProcessor) Process(ctx context.Context, question string, expertName string, expertDomain string, reasoningCharter string) *ProcessedQuestion {
	result := &ProcessedQuestion{Original: question}

	// Fast skip for very short questions (already domain-aligned)
	if !shouldProcess(question) { return result }

	// STEP 1: UNDERSTAND (LLM identifies the core concepts)
	understanding, _ := p.stepUnderstand(ctx, question, expertName, expertDomain, reasoningCharter)
	// Example Output: CORE_PROBLEM: min knight moves, DOMAIN_CONCEPTS: BFS, grid

	// STEP 2: EXTRACT (LLM rewrites the question dropping the story)
	extracted, _ := p.stepExtract(ctx, question, expertName, expertDomain, understanding)

	// STEP 2.5: DETERMINISTIC GUARDS (Regex/Logic Checks)
	// Crucial: Numbers, ALL-CAPS (SQL/BFS), and Quotes must survive extraction!
	missingConstraints := missingConstraintTokens(question, extracted)
	if len(missingConstraints) > 0 {
		result.SkippedReason = "guard_constraint_dropped"
		return result // Fallback to original question
	}

	// STEP 3: VERIFY (LLM double-checks itself)
	verified, _, _ := p.stepVerify(ctx, question, extracted)
	if !verified {
		result.SkippedReason = "step3_verify_no"
		return result // Fallback to original question
	}

	result.VerificationPassed = true
	result.Extracted = extracted
	return result
}
```

### 2. Guard Logic (`missingConstraintTokens`)

A deterministic safety check (no LLM required) to ensure technical accuracy isn't lost during the rewrite.

```go
func missingConstraintTokens(original, extracted string) []string {
	var missing []string
	// constraintTokens() extracts digits (\d+), acronyms ([A-Z]+), and "quotes"
	for _, tok := range constraintTokens(original) {
		if !strings.Contains(extracted, tok) {backend-go/internal/workflow/gate_system.go
			missing = append(missing, tok)
		}
	}
	return missing
}
```

---

## Part 7: engine.go (The Decision Engine / 5 Gates)

**File Path:** `backend-go/internal/decision/engine.go`

This file is the **Pre-Flight Reasoning Brain**. Before the heavy Generation LLM is ever called, this engine runs the user's question through 5 strict gates. It catches vague questions, out-of-domain topics, charter violations, and over-engineering attempts.

### 1. `Process` (The 5-Gate System)

Each gate catches a specific failure mode. If a gate fails, the system immediately returns a specific `ResponseMode` (like `ASK`, `WARN`, `PUSH_BACK`, `REFUSE`).

```go
func (e *Engine) Process(...) (*DecisionResult, error) {
	// GATE 0: Structure Permission (Optional)
	// Asks user: "Chahiye structure/boilerplate ya sirf logic likh doon?"
	if gateResult, rewritten := e.gateStructurePermission(...); gateResult != nil {
		return gateResult, nil
	}

	// GATE 1: Information Sufficiency (Is the question too vague?)
	// Checks for words like "how do i", "best way". If vague, asks for clarification.
	if result, needsLLM := e.gate1(question, expert); result != nil {
		return result, nil
	} else if needsLLM {
		if result := e.gate1WithLLM(ctx, question, expert); result != nil { return result, nil }
	}

	// GATE 2: Knowledge Coverage (Is this in my domain?)
	if result := e.gate2(question, chunks, expert); result != nil {
		return result, nil
	}

	// GATE 3: Charter Compliance (Does this violate my specific rules?)
	if result := e.gate3(ctx, question, expert); result != nil {
		if result.Mode == ModeWARN { warning = result.Warning } else { return result, nil }
	}

	// GATE 4: Necessity Check (Is the user over-engineering?)
	if result := e.gate4(ctx, question, projectSummary, expert); result != nil {
		return result, nil
	}

	// GATE 5: Generate Answer (Calls the China Wall Enforcer)
	enforceResult, _ := e.chinaWall.Enforce(...)
	// ... maps success/refusal ...
}
```

### 2. Specific Gate Implementations

**Gate 2 (Knowledge Coverage):** Refuses if the RAG search found nothing relevant.

```go
func (e *Engine) gate2(question string, chunks []chinawall.CourseChunk, expert Expert) *DecisionResult {
	if len(chunks) == 0 || bestScore < 0.20 {
		return &DecisionResult{
			Mode:    ModeREFUSE,
			Content: fmt.Sprintf("This topic is not in my training material. I specialize in: %s", expert.Domain),
		}
	}
	return nil
}
```

**Gate 4 (Necessity Check):** Detects over-engineering keywords and uses a cheap LLM to push back if premature.

```go
func (e *Engine) gate4(ctx context.Context, question string, projectSummary string, expert Expert) *DecisionResult {
	overEngineeringTerms := []string{"kubernetes", "microservices", "kafka", "redis cluster"}

	for _, term := range overEngineeringTerms {
		if strings.Contains(strings.ToLower(question), term) {
			// Ask cheap LLM: "Is this premature for this project?"
			if result := e.checkNecessity(ctx, question, projectSummary, term); result != nil {
				return result // ModePUSHBACK
			}
		}
	return nil
}
```

---

## Part 8: assembler.go (Context Assembler)

**File Path:** `backend-go/internal/context/assembler.go`

This file is responsible for building the **Context Payload** sent to the LLM. It aggregates data from Postgres and VectorDBs (Chat history, summaries, RAG chunks, and project memory) while enforcing a strict **Token Budget** so the LLM context window never overflows.

### 1. The Token Budget Policy

Context is treated as a strict budget. If the system hits the ceiling, it evicts data based on predefined priorities (saving RAG chunks and explicit replies for last).

```go
// Budget Allocation
const (
	budgetSummaryPct   = 10 // Rolling summary
	budgetL2Pct        = 20 // Project memory
	budgetRecentPct    = 20 // Recent N messages
	budgetHistoryPct   = 15 // Semantic relevant history
	budgetChunksPct    = 35 // RAG Training Chunks (Most Important)
)
```

### 2. `Assemble` (Parallel Fan-Out Design)

Instead of querying the database 6 times sequentially (which would cause massive latency), this function uses **Goroutine Fan-Out**. It launches all 6 independent context fetches simultaneously.

```go
func (a *Assembler) Assemble(...) (*AssembledContext, error) {
	budget := a.maxTokens

	// 1. Fan-out: independent queries run concurrently
	summaryCh := make(chan summaryResult, 1)
	l2Ch := make(chan l2Result, 1)
	recentCh := make(chan recentResult, 1)
	chunksCh := make(chan chunksResult, 1)
    // ...

	go func() { s, err := a.getRollingSummary(ctx, chatID); summaryCh <- summaryResult{s, err} }()
	go func() { pc, err := a.memManager.GetProjectContext(...); l2Ch <- l2Result{pc, err} }()
	go func() { msgs, err := a.getRecentMessages(ctx, chatID, a.recentMsgs); recentCh <- recentResult{msgs, err} }()
	go func() { c, err := a.getCourseChunks(...); chunksCh <- chunksResult{c, err} }()

	// 2. Collect all parallel results
	summaryRes := <-summaryCh
	l2Res := <-l2Ch
	recentRes := <-recentCh
	chunksRes := <-chunksCh

	// 3. Merge & Enforce Token Budget
	assembled := &AssembledContext{}
	tokensUsed := 0
    // ... trims each slice based on the percentage budget (e.g. 35% for chunks) ...

	// 4. Hard Ceiling Enforcement
    // Evicts lowest priority context (semantic history) before highest (RAG chunks)
	tokensUsed = enforceHardCeiling(assembled, budget, tokensUsed, a.logger)
	assembled.TotalTokens = tokensUsed

	return assembled, nil
}
```

### Assembler.go Summary

`assembler.go` is designed for **Speed** (via concurrent fan-out) and **Safety** (via token budget enforcement). It guarantees the LLM receives the maximum possible context without ever crashing due to token limit errors.

---

## Part 9: service.go (Chat & Message DB Layer)

**File Path:** `backend-go/internal/chat/service.go`

This file handles the **Persistence and Indexing** of the chat system. While the Orchestrator thinks and generates, `service.go` writes those decisions securely to Postgres. It also prepares data for semantic search.

### 1. `Message` Struct (The Data Model)

This struct stores every detail of an interaction, including all China Wall metadata. It uses `interface{}` for complex JSON objects (like Citations and TemplateSections) to avoid cyclic dependencies with the `chinawall` package.

```go
type Message struct {
	ID                   uuid.UUID   `json:"id"`
	ChatID               uuid.UUID   `json:"chat_id"`
	Role                 string      `json:"role"`
	Content              string      `json:"content"`
	TurnNumber           int         `json:"turn_number"`
	ExpertID             *uuid.UUID  `json:"expert_id,omitempty"`
	DecisionMode         string      `json:"decision_mode,omitempty"`
	Citations            interface{} `json:"citations,omitempty"` // China Wall Citations
	WarningText          string      `json:"warning_text,omitempty"` // Gate 3 Warn
	ClarifyingQuestions  []string    `json:"clarifying_questions,omitempty"` // Gate 1 Ask
	ReplyToMessageID     *uuid.UUID  `json:"reply_to_message_id,omitempty"`
	TemplateSections     interface{} `json:"template_sections,omitempty"`
	CollabSections       interface{} `json:"collab_sections,omitempty"` // Collaborative mode output
}
```

### 2. `SaveMessage`

Writes the message to Postgres. Normalizes empty fields (like `""`) to `NULL` to satisfy database constraints.

```go
func (s *Service) SaveMessage(ctx context.Context, msg Message) (uuid.UUID, error) {
	// ... normalizes empty strings to NULL to pass DB constraints ...
	err := s.db.QueryRow(ctx,
		`INSERT INTO messages
			(chat_id, role, content, turn_number, expert_id, decision_mode,
			 confidence, warning_text, clarifying_questions, citations,
			 reply_to_message_id, template_sections,
			 quality_score, coverage, refusal_reason, collab_sections)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		 RETURNING id`,
        // ... parameters ...
	).Scan(&id)
	return id, err
}
```

### 3. `IndexTurn` (Semantic Search Vectorization)

Called asynchronously after a turn is completed. It stores a vector embedding of the turn using `pgvector`. This is exactly what allows `assembler.go` (Step 4: Semantic History) to retrieve relevant past messages based on meaning, not just recency.

```go
func (s *Service) IndexTurn(...) error {
	// Uses pgvector.NewVector to safely inject embeddings, preventing SQL injection
	_, err := s.db.Exec(ctx,
		`INSERT INTO chat_index
			(chat_id, message_id, turn_number, one_line_summary, topic, importance, embedding)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		chatID, messageID, turnNumber, summary, topic, importance,
		pgvector.NewVector(embedding),
	)
	return err
}
```

---

## Part 10: engine.go (Workflow Engine & State Machine)

**File Path:** `backend-go/internal/workflow/engine.go`

This file implements the core **State Machine** for long-running workflows (agents working in sequence). It ensures strict forward progression through predefined phases (Intake → Understanding → High Level Design → Detailed Design → Implementation → QA → Handoff) and manages state checkpoints.

### 1. Phase and Status Management

The `Engine` enforces a valid transition path. No expert or agent can directly mutate a workflow's state or skip phases randomly.

```go
// phaseOrder defines the valid forward progression.
var phaseOrder = []string{
	PhaseIntake,
	PhaseUnderstanding,
	PhaseHighLevelDesign,
	PhaseDetailedDesign,
	PhaseImplementation,
	PhaseQA,
	PhaseHandoff,
	PhaseCompleted,
}

func isValidTransition(current, next string) bool {
	// ... logic to ensure next is strictly after current in phaseOrder
	// (or allows early termination to PhaseCompleted) ...
}
```

### 2. `TransitionPhase` & Checkpoints

Before moving a workflow into a new phase, it saves a snapshot of the current state. This allows a crashed workflow to recover and resume perfectly from its last known phase.

```go
func (e *Engine) TransitionPhase(...) (*Workflow, error) {
	// 1. Validate forward transition
	if !isValidTransition(w.CurrentPhase, nextPhase) { return nil, err }

	// 2. Write phase checkpoint BEFORE transitioning
	if err := e.writeCheckpoint(ctx, workflowID, w.CurrentPhase, checkpointSnapshot, lastBlackboardSeq); err != nil {
		// logs warning...
	}

	// 3. Update DB state to new phase
	err = e.db.QueryRow(ctx, `UPDATE workflows SET current_phase = $1 ...`, nextPhase).Scan(...)

	return &updated, nil
}
```

### 3. Token Cost Budgeting

Long-running agents can consume huge amounts of API credits. The engine checks costs continuously and tracks Soft Limits (75% budget) and Hard Limits (100% budget, pause workflow).

```go
func (e *Engine) UpdateCostSpent(ctx context.Context, workflowID uuid.UUID, additionalCostUSD float64) (*CostLimitResult, error) {
    // 1. Update total spent in DB
	// 2. Calculate thresholds based on the Workflow's CostBudgetUSD
	softThreshold := w.CostBudgetUSD * (w.CostSoftLimitPct / 100.0)
	hardThreshold := w.CostBudgetUSD * (w.CostHardLimitPct / 100.0)

	result := &CostLimitResult{
		CostSpentUSD: w.CostSpentUSD,
		SoftLimitHit: w.CostSpentUSD >= softThreshold,
		HardLimitHit: w.CostSpentUSD >= hardThreshold,
	}
    // ... logs warnings if limits are hit ...
	return result, nil
}
```

---

## Part 11: runner.go (Workflow Task Runner & Recovery)

**File Path:** `backend-go/internal/workflow/runner.go`

This file is the **Master Controller** for long-running workflows. It runs asynchronously in a goroutine and orchestrates the entire lifecycle: planning, building the DAG (Directed Acyclic Graph), executing waves in parallel, handling client approvals, and recovering from server crashes.

### 1. `Run` (The Master Goroutine)

The `Run` method executes the workflow step-by-step. It uses a **Single-Flight Lock** (`sync.Map`) to guarantee that two parallel processes (e.g. an API call and a server restart) don't accidentally execute the same workflow twice, which would waste money and produce duplicate design documents.

```go
func (r *WorkflowRunner) Run(ctx context.Context, workflowID uuid.UUID) {
	// Single-flight lock: exactly one runner may drive a workflow at a time.
	if _, loaded := r.inflightRuns.LoadOrStore(workflowID, struct{}{}); loaded {
		log.Warn("runner: a runner is already driving this workflow; refusing a second one")
		return
	}
	defer r.inflightRuns.Delete(workflowID)

    // 1. Load Workflow and Experts
    // 2. Load Checkpoint (if recovering from crash)
    // 3. Planner.Plan() -> Tasks
    // 4. BuildDAG() -> Execution Waves
    // 5. AskClient for Plan Approval (Pauses Workflow)
    // 6. Execute Phases (Understanding -> Design -> Implementation -> QA)
    // 7. Complete()
}
```

### 2. Pod-Restart Recovery (Checkpoints)

If the backend crashes midway through a 30-minute workflow, `runner.go` will automatically resume when the server restarts. It reads `workflows.runner_state` and intelligently skips phases and tasks that are already done.

```go
	// Step 3: Load checkpoint (pod-restart recovery).
	resumePhase := ""
	var savedState *runnerState
	if saved, _ := r.loadRunnerState(ctx, workflowID); saved != nil {
		savedState = saved
		resumePhase = saved.Phase
        // Skips experts already completed in the resumed phase
		resumeCompleted = append([]string(nil), saved.CompletedExpertIDs...)
	}
```

### 3. Separation of Concerns (Single Write Path)

The runner **never** mutates tasks directly in the database. It only posts events to the `blackboard`. A separate `Projector` listens to the blackboard and updates the UI (`workflow_tasks` table). This decoupled architecture prevents race conditions.

```go
	// Post task_plan_ready -> Projector inserts workflow_tasks.
    // Runner does NOT write to workflow_tasks directly.
	_, err = r.store.Post(ctx, blackboard.PostRequest{
		WorkflowID:     workflowID,
		EventType:      "task_plan_ready",
		PostedByClient: true,
		Content:        planContent,
	})
```

---

## Part 12: agent_loop.go (The OTA Agent Loop)

**File Path:** `backend-go/internal/workflow/agent_loop.go`

This file is the brain of a **single expert** working on a **single task**. It implements the **Observe-Think-Act (OTA)** loop. The expert will loop through these steps until it explicitly outputs `TASK_COMPLETE`.

### 1. The OTA Loop (`Run`)

1. **Observe:** Reads the blackboard for new artifacts from peer experts. Fetches its own RAG training data.
2. **Think:** Merges the blackboard, project memory, and RAG data into a context string. Calls the LLM.
3. **Act:** Extracts tool calls (like `PostArtifact` or `AskExpert`) from the LLM's response and executes them.

```go
func (a *AgentLoop) Run(ctx context.Context, req AgentLoopRequest) (*AgentLoopResult, error) {
	for iter := 1; iter <= maxIter; iter++ {
		// 1. OBSERVE: Read blackboard
		newArtifacts, _ := a.tools.ReadBlackboard(ctx, ReadBlackboardRequest{...})

		// 2. 3-Gate Knowledge System (Training -> Peers -> Generic)
		gateResult, _ := a.gateSystem.RunGates(...)

		// 3. THINK: Build prompt and call LLM
		contextText := FormatGateContext(gateResult) + "\n[BLACKBOARD]\n" + a.buildBlackboardContext(...)
		llmResp, err := a.gateway.Call(ctx, gateway.LLMRequest{...})

		// 4. ACT: Parse and execute tool calls
		artifactID, _ := a.act(ctx, req, llmResp.Content)

		// Exit if finished
		if strings.Contains(llmResp.Content, "TASK_COMPLETE") {
			result.Completed = true
			break
		}
	}
	return result, nil
}
```

### 2. Context Summarization

If an expert reads 20 artifacts from the blackboard, the LLM context window would blow up. `buildBlackboardContext` automatically summarizes older artifacts using a cheap LLM if there are more than 10, keeping only the 5 most recent in raw detail.

```go
const contextSummarizeThreshold = 10
const recentArtifactsToKeepRaw = 5

func (a *AgentLoop) buildBlackboardContext(...) (string, error) {
	if len(artifacts) <= contextSummarizeThreshold {
		return formatArtifacts(artifacts), nil
	}
	// Summarizes older artifacts to prevent context limit explosion...
}
```

### 3. Bulletproof JSON Parsing (`extractFirstJSONObject`)

LLMs often hallucinate stray brackets or wrap tool calls in markdown formatting. This loop doesn't just call `json.Unmarshal`. It manually walks the string and counts braces `{}` to extract the exact JSON object, guaranteeing the agent won't crash on bad formatting.

---

## Part 13: gate_system.go (3-Gate Knowledge Access)

**File Path:** `backend-go/internal/workflow/gate_system.go`

This file implements the strict knowledge control system for workflow experts. The design principle is **"Pehle Ghar mein Dhoondo, Fir Dost se Pucho, Fir Google Karo"** (Search your own house, then ask a friend, then Google). It prevents the LLM from hallucinating generic answers when it should be relying on specific company training material.

### 1. Gate 1: Own Training (Vector DB)

The system searches the expert's own RAG (Retrieval-Augmented Generation) chunks. It uses a **Two-Band Threshold** instead of a simple pass/fail.

- **Strong (>= 0.70):** The training covers the task perfectly. Generic LLM knowledge is BLOCKED completely.
- **Usable (>= 0.40):** The training is somewhat relevant. It is added to the prompt as principles, but peers and generic knowledge might still be needed.

### 2. Gate 2: Peer Knowledge (Parallel Polling)

Even if Gate 1 passes, the system _always_ polls the other experts in the workflow. It runs parallel goroutines with a strict 10-second timeout (`gate2PollTimeout`) to fetch what the other experts (like PM, LLD, DSA) were trained on for this specific task.

```go
func (g *GateSystem) pollPeers(ctx context.Context, askingExpertID uuid.UUID, taskDescription string, allExperts []workflowExpert) []PeerContribution {
	pollCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	// Spawns goroutines to fetch RAG chunks from all peers concurrently...
}
```

### 3. Gate 3: Generic Gap Filling

If the training and peers don't fully cover the task, the **Client** decides if generic knowledge is allowed via `genericAllowancePct` (e.g., 0-30%). The model doesn't get to decide this itself.

### 4. `FormatGateContext` (The Prompt Builder)

This function builds the highly constrained prompt injected into the LLM. It forces the LLM to cite its sources (`[CHUNK_uuid]`, `[PEER:ExpertName]`) and tag any hallucinated/generic knowledge with `[GENERIC]`.

```go
func FormatGateContext(result *GateResult) string {
	// ...
	if result.GenericAllowed {
		sb.WriteString("- Tag EVERY generic sentence with [GENERIC]. Untagged generic content is a violation.\n")
		sb.WriteString("- Everything not tagged [GENERIC] MUST carry [CHUNK_id] or [PEER:ExpertName].\n")
	} else {
		sb.WriteString("[GENERIC KNOWLEDGE IS BLOCKED]\n")
		sb.WriteString("Build the answer ONLY from the material above.\n")
		sb.WriteString("- If something the task asks for is genuinely not in the material above,\n")
		sb.WriteString("  do NOT invent it. Write [NOT_COVERED: <what is missing>] instead.\n")
	}
	// ...
}
```

---

## Part 15: blackboard package (`store.go` & `subscriber.go`)

**Directory Path:** `backend-go/internal/blackboard/`

The **Blackboard** is the central communication hub for the entire workflow. Experts don't talk to each other directly; they post artifacts (designs, code, requirements) to the blackboard, and other experts read them. It is built using **PostgreSQL (Source of Truth)** and **Redis Pub/Sub (Real-time Notifications)**.

### 1. `store.go` (The Single Write Path)

All writes to the blackboard must go through the `Post()` method. This ensures no expert can bypass the system and mutate data directly.

- **Idempotency & Deduplication:** It generates a `dedup_key` (SHA-256 hash of the event type, poster ID, and JSON content). It uses Postgres' `ON CONFLICT DO NOTHING`. If an expert crashes and retries a task, it won't accidentally post the exact same design document twice.
- **Redis Notification:** After safely writing to Postgres, it publishes a lightweight notification payload to Redis. This is "best-effort"—if Redis is down, the system doesn't fail; the data is safely in Postgres.

```go
func (s *Store) Post(ctx context.Context, req PostRequest) (*Event, error) {
	// 1. Hash content to create dedup_key
	dedupKey := sha256.Sum256(...)

	// 2. Safe Insert (Idempotent)
	err = s.db.QueryRow(ctx,
		`INSERT INTO blackboard_events ...
		 ON CONFLICT (workflow_id, dedup_key) DO NOTHING
		 RETURNING id, sequence_number, posted_at`, ...).Scan(...)

	// 3. Publish lightweight notification to Redis
	s.publishNotification(ctx, event)
}
```

### 2. `subscriber.go` (The Reliable Reader)

The `Subscriber` listens to the Redis channel for new events.

- **Cursor Management:** It keeps track of the last `sequence_number` it processed (saved in Redis). If the server restarts, it knows exactly where it left off.
- **Safe Catch-Up:** It doesn't trust the Redis payload for the actual data. When it gets a ping from Redis, it runs `catchUp()` to fetch the real, full data from Postgres. If the Redis connection drops, it automatically reconnects and fetches any missed events from Postgres using its cursor.

```go
func (s *Subscriber) run(...) {
	// 1. Catch up on events missed while offline
	s.catchUp(ctx, workflowID, &cursor, eventCh)

	// 2. Listen to Redis
	pubsub := s.redis.Subscribe(ctx, channel)

	for {
		msg := <-pubsub.Channel()
		// 3. On ping, fetch new events from Postgres
		s.catchUp(ctx, workflowID, &cursor, eventCh)

        // 4. Save cursor so we don't read them again on crash
		s.SaveCursor(ctx, workflowID, expertID, cursor)
	}
}
```

---

## Part 16: expert package (`handler.go`)

**File Path:** `backend-go/internal/expert/handler.go`

This file provides the HTTP API endpoints used by the frontend to list, view, and interact with Domain Experts.

### 1. `ListActive` (GET /experts)

This endpoint returns the catalog of experts available to the user. It has strict filtering rules:

- **Fully Trained Only:** It only returns experts where `training_status='trained'` and `is_training=FALSE`. Draft experts (in ingestion) are hidden so clients don't see broken/empty bots.
- **Tenant & RBAC Scoping:** If a user belongs to Tenant A, they cannot see Tenant B's custom experts. If the user is a `domain_expert` role, they only see experts explicitly granted to their account.

### 2. `GetTopics` (GET /experts/:id/topics)

Returns what the expert actually knows (Capabilities). It queries the `expert_capabilities` table to return the topics, their depth levels, and evaluation metrics (`eval_passed`, `can_handle`, `cannot_handle`). This allows the frontend to show a "What I can do" section for the expert.

### 3. `AnswerFormats` (GET /experts/:id/answer-formats)

Returns the named answer formats available for this expert's category. This allows the chat UI to show a dropdown so the user can choose how they want the answer formatted (e.g., "Code Only", "Detailed Explanation", "Summary").

---

## Part 17: training package (`ingestion_pipeline.go`)

**File Path:** `backend-go/internal/training/ingestion_pipeline.go`

This file handles the end-to-end ingestion and training of an Expert from an uploaded transcript. The process is extremely robust, heavily checkpointed, and designed to never waste expensive LLM/Embedding calls if it crashes.

### The Pipeline Flow (The 6 Stages)

1.  **Clean (Step 0):** Before chunking, it removes 45-55% of transcript noise (greetings, timestamps, filler words). Noise lowers vector rerank scores, so removing it first is critical.
2.  **Chunk:** Splits the cleaned text into logical chunks.
3.  **Topic Extraction:** Runs parallel, batched LLM calls to assign topics to chunks. If the LLM provider hits a hard error (like a 402 Payment Required), the pipeline intentionally **pauses** the job instead of silently failing or marking everything as a "general" topic.
4.  **Charter Extraction:** Uses a strong LLM to extract the expert's "Reasoning Charter" and "Clarification Charter".
5.  **Embedding:** Generates vector embeddings for all chunks in parallel by calling the Python ML sidecar.
6.  **Store:** Stores the chunks in PostgreSQL (`course_chunks`). It uses `ON CONFLICT DO NOTHING` so it can safely resume or append new transcripts to an existing expert without duplicating chunks.

### Checkpointing and Resumability

Every time a batch of topics or embeddings finishes, the state is saved to the database.

```go
cpWriter.Write(ctx, JobCheckpoint{
    Stage:          StageTopicExtraction,
    ChunksDone:     chunksDone,
    // ...
})
```

If the server restarts, `LoadCheckpoint()` figures out exactly where the pipeline left off. For example, if it crashes during embedding, it will load the previously extracted topics directly from the DB (`loadTopicsFromDB`) instead of paying for the LLM calls again.

---

## Part 18: ml-sidecar (Python ML Microservice)

**Directory Path:** `ml-sidecar/`

The ML sidecar is a local Python (FastAPI) microservice.
**WHY keep this separate from Go?** Because Python has the most mature ML and document parsing libraries. Running heavy CPU-bound ML inference in the Go API process would block its event loop and crash the web server.

It provides three main endpoints:

1.  **`/embed`:** Generates 768-dimensional vector embeddings using the `bge-base-en-v1.5` model. This is used during ingestion to store chunks in PGVector, and during chat to embed the user's query.
2.  **`/rerank`:** Uses a cross-encoder model (`bge-reranker-base`) to score retrieved chunks against a query. Vector similarity (Cosine) alone is often inaccurate. A cross-encoder reads both the query and the chunk together to provide a highly accurate relevance score.
3.  **`/extract`:** Parses uploaded files (e.g., PDFs, Word docs) into plain text for the ingestion pipeline. Parsing untrusted documents is risky and CPU-heavy, so the sidecar acts as a sandbox to protect the main Go API.

---

## Part 19: gateway package (`model_gateway.go`)

**File Path:** `backend-go/internal/gateway/model_gateway.go`

This is the central LLM router for the entire application. It abstracts away specific LLM providers (Anthropic, DeepSeek, Gemini, CodeCraftAPI) so the rest of the codebase just calls `gateway.Call()`.

**Key Features:**

1.  **Provider Abstraction:** Translates internal requests into provider-specific API calls.
2.  **Streaming & Sync:** Supports both standard `Call()` and `StreamCall()` (used by Gate 5 for real-time typing effect).
3.  **Resilience (Breaker & Fallback):** If the primary LLM provider goes down (or hits a rate limit), the gateway automatically trips a circuit breaker and switches to a configured Fallback Provider. It will retry up to 3 times before doing this.
4.  **Cost Tracking:** Automatically calculates token usage and dollar cost for every single LLM call and persists it to the usage metrics DB.

---

## Part 20: message package (`handler.go`)

**File Path:** `backend-go/internal/message/handler.go`

This file exposes the main POST `/chats/:id/messages` HTTP endpoint that the frontend calls when the user presses "Send" in the chat UI.

**Key Features:**

1.  **SSE Streaming:** It establishes a Server-Sent Events (SSE) connection so it can push chunks of text to the frontend in real-time. (e.g. `{"type": "chunk", "content": "Hello..."}`)
2.  **Attachment Parsing:** If the user attaches a PDF/Word file to the chat, it intercepts the file and calls `docextract` (which calls the ML Sidecar) to turn it into plain text before feeding it to the Orchestrator.
3.  **Orchestrator Trigger:** It triggers `h.orchestrator.Process()` (or `ProcessCollaborative` if multiple experts are selected in relay mode) and pipes the tokens back to the user.
4.  **Persistence:** Saves the user message and the final expert responses (including Citations, Template Sections, and Quality Scores) synchronously into the PostgreSQL database so the chat history is preserved on page reload.

---

## Part 14: handler.go (Workflow HTTP API)

**File Path:** `backend-go/internal/workflow/handler.go`

This file provides the HTTP API endpoints that the frontend uses to manage workflows. It bridges the gap between client actions (clicking "Approve" on the Kanban board) and the internal state machine.

### 1. `RunWorkflow` (Starting Background Jobs)

When the user clicks "Start", the server responds immediately (HTTP 202 Accepted) and launches the `Projector` and `WorkflowRunner` as background goroutines.

- **Order matters:** The `Projector` (which listens to events and updates the DB) is explicitly started _before_ the `Runner` (which emits events) so that the very first plan event isn't missed.

```go
func (h *Handler) RunWorkflow(runner *WorkflowRunner) gin.HandlerFunc {
	// ... validation ...
	runCtx := context.Background()

	// Start Projector FIRST: must be subscribed before Runner posts events.
	go h.projector.Run(runCtx, id)

	// Start Runner: drives workflow to completion.
	go runner.Run(runCtx, id)

	c.JSON(202, map[string]interface{}{"status": "accepted"})
}
```

### 2. `RespondToApproval` (Gate Keeper)

When a workflow pauses (e.g., waiting for client approval on the Understanding phase), the client sends an API request here.
It translates API terms to strict Database constraints. If the client clicks "Request Changes", they can also send a `GenericAllowancePct` (e.g., 30%), which updates the workflow to allow up to 30% generic LLM knowledge when it resumes.

```go
	// decisionToStatus translates API vocabulary to strict DB vocabulary.
	decisionToStatus := map[string]string{
		"approve":                  "approved",
		"approve_with_notes":       "approved_with_notes",
		"request_changes":          "changes_requested",
		"reject_and_restart_phase": "rejected",
		"cancel_workflow":          "cancelled",
	}
    // ... updates generic allowance if provided, then resumes the workflow ...
```

### 3. Edge Case Recovery (`RetryTask` & `EnableCodeDelivery`)

- **`RetryTask`:** If an LLM call fails due to a rate limit, the client can manually retry that specific task without restarting the 30-minute workflow.
- **`EnableCodeDelivery`:** Allows a client to turn a "Design-Only" workflow into a "Deliver-Code" workflow after it has already finished, resuming from the implementation phase using the approved design.
