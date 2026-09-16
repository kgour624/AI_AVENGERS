package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	appcontext "ai_avengers/backend/internal/context"
	"ai_avengers/backend/internal/category"
	"ai_avengers/backend/internal/chinawall"
	"ai_avengers/backend/internal/decision"
	"ai_avengers/backend/internal/memory"
	"ai_avengers/backend/internal/observability"
	"ai_avengers/backend/internal/ratelimit"
	"ai_avengers/backend/internal/selflearning"
)

// OrchestratorRequest is the input to the orchestrator.
type OrchestratorRequest struct {
	ProjectID   uuid.UUID
	ClientID    uuid.UUID
	ChatID      uuid.UUID
	Message     string
	ExpertIDs   []uuid.UUID
	TurnNumber  int
	// ReplyToMessageID/IncludeFullThread (CT-C1/C2): nil/false for every
	// fresh (non-reply) question — the existing behavior for every
	// request sent before this feature. Passed through to
	// appcontext.Assembler.Assemble unchanged.
	ReplyToMessageID  *uuid.UUID
	IncludeFullThread bool
	// UserMessageID (CT-C4): the already-saved id of the user message
	// that triggered this request (message/handler.go's Send saves it
	// BEFORE calling orchestrator.Process). Used only by
	// processWithExpert's structure-permission-ASK branch, to set that
	// ASK message's own reply_to_message_id back to this id so a LATER
	// reply-to-the-ASK can walk one more parent level and recover the
	// original question (decision/engine.go's gateStructurePermission).
	UserMessageID uuid.UUID
	// TokenCh: when non-nil, Gate 5 generation streams tokens here.
	// message/handler.go creates this channel and forwards tokens to SSE.
	// nil = blocking (used by tests, smoke test, non-streaming callers).
	TokenCh chan<- string
}

// OrchestratorResponse is the full output including all expert responses.
type OrchestratorResponse struct {
	ExpertResponses []ExpertResponse
	Synthesis       *SynthesisResult
	TurnNumber      int
	TotalTokens     int
	DurationMs      int64
	// PhaseTimings holds per-phase latency breakdown for observability.
	// Populated by processWithExpert via PhaseTimer (Observer pattern).
	// Nil when no experts ran (error path).
	PhaseTimings []observability.PhaseResult
}

// ExpertResponse is one expert's response.
type ExpertResponse struct {
	ExpertID    uuid.UUID            `json:"expert_id"`
	ExpertName  string               `json:"expert_name"`
	Domain      string               `json:"domain"`
	Mode        decision.ResponseMode `json:"mode"`
	Content     string               `json:"content"`
	Citations   []chinawall.Citation `json:"citations"`
	Confidence  float64              `json:"confidence"`
	GateStopped int                  `json:"gate_stopped"`
	Warning     string               `json:"warning,omitempty"`
	Questions   []string             `json:"questions,omitempty"`
	Error       string               `json:"error,omitempty"`
	// TemplateSections (CT-B4): populated only when this expert has a
	// category with a non-empty template_schema. nil for every flat-text
	// expert response (CT-L2) — frontend (CT-D5, not yet built) must
	// check len(TemplateSections) > 0 before rendering structured UI,
	// falling back to plain Content otherwise, exactly like every layer
	// below this one already does.
	TemplateSections []chinawall.TemplateSectionResult `json:"template_sections,omitempty"`
	// ReplyToUserMessageID (CT-C4): set ONLY when GateStopped==-1 (this
	// response IS a structure-permission ASK, decision/engine.go's
	// gateStructurePermission sentinel). message/handler.go's
	// saveAssistantMessage uses this as the ASK message's OWN
	// reply_to_message_id when saving it, so a later reply-to-this-ASK
	// can walk one more parent level and recover the original question.
	// nil for every other response.
	ReplyToUserMessageID *uuid.UUID `json:"-"`
}

// SynthesisResult holds the combined view when multiple experts respond.
type SynthesisResult struct {
	Agreements     []string       `json:"agreements"`
	Contradictions []Contradiction `json:"contradictions"`
	Summary        string         `json:"summary"`
}

// Contradiction is a point where two experts disagree.
type Contradiction struct {
	Topic     string `json:"topic"`
	ExpertA   string `json:"expert_a"`
	PositionA string `json:"position_a"`
	ExpertB   string `json:"expert_b"`
	PositionB string `json:"position_b"`
}

// expertRecord holds DB data for an expert.
type expertRecord struct {
	ID                   uuid.UUID
	Name                 string
	Domain               string
	ReasoningCharter     string
	ClarificationCharter map[string][]string
	// CategoryID (CT-B4): nullable per CT-L2. nil means flat-text expert.
	CategoryID *uuid.UUID
}

// Orchestrator coordinates multiple domain experts for a single request.
//
// WHY parallel goroutines:
// 3 experts sequential = 3x latency (15+ seconds).
// 3 experts parallel = same latency as 1 (5 seconds).
// Go goroutines make this trivial — this is exactly what they're for.
type Orchestrator struct {
	db          *pgxpool.Pool
	assembler   *appcontext.Assembler
	decisionEng *decision.Engine
	memManager  *memory.Manager
	// categoryRegistry (CT-B4): looked up per expert in loadExperts to
	// resolve category_id -> TemplateSections/DefaultLanguage. nil is a
	// valid state (server started before CT-A wiring, or category feature
	// disabled) — loadExperts treats nil registry exactly like "expert has
	// no category_id", never panics on nil dereference (see loadExperts).
	categoryRegistry *category.Registry
	// selfLearning (Self-Learning Mode): nil = disabled (zero regression).
	// When non-nil, processWithExpert runs Understand → Extract → Verify
	// on the raw question before passing it to the decision engine.
	// This converts story-noisy or domain-agnostic questions into
	// domain-specific signal, improving RAG retrieval accuracy.
	// WHY nil-safe: allows disabling self-learning without code change.
	selfLearning *selflearning.QuestionProcessor
	// expertLimiter enforces per-expert request rate limits.
	// Prevents a single expert from being overwhelmed by concurrent requests
	// which would cause LLM provider rate limit hits.
	// Strategy pattern: swap NoopLimiter for testing, TokenBucketLimiter for prod.
	expertLimiter ratelimit.RateLimiter
	// expertSemaphores enforces per-expert concurrency limits.
	// Each expert gets a buffered channel of size expertMaxConcurrency.
	// Acquiring = send to channel. Releasing = receive from channel.
	// WHY separate from rate limiter:
	//   Rate limiter: controls request rate (requests/second).
	//   Semaphore: controls concurrent in-flight requests.
	//   Both are needed: rate limiter prevents burst, semaphore prevents
	//   goroutine explosion when requests are slow (LLM latency 2-10s).
	// WHY sync.Map: keys are expert IDs (strings), created lazily.
	//   sync.Map is optimized for write-once, read-many — perfect here.
	expertSemaphores sync.Map // map[string]chan struct{}
	// expertMaxConcurrency: max concurrent requests per expert.
	// WHY 3: single admin user, 3 concurrent experts is the realistic max.
	// Higher = goroutine explosion under LLM latency. Lower = unnecessary queuing.
	expertMaxConcurrency int
	logger      *zap.Logger
}

// NewOrchestrator creates a new orchestrator.
func NewOrchestrator(
	db *pgxpool.Pool,
	assembler *appcontext.Assembler,
	decisionEng *decision.Engine,
	memManager *memory.Manager,
	categoryRegistry *category.Registry,
	selfLearning *selflearning.QuestionProcessor,
	logger *zap.Logger,
) *Orchestrator {
	return &Orchestrator{
		db:               db,
		assembler:        assembler,
		decisionEng:      decisionEng,
		memManager:       memManager,
		categoryRegistry: categoryRegistry,
		selfLearning:     selfLearning,
		logger:           logger,
		// Default: 5 burst, 2 requests/second per expert.
		// WHY these numbers: LLM providers typically allow 5-10 RPM per key.
		// 2 req/s = 120 RPM, well within provider limits.
		// Burst=5 allows short spikes (e.g. user sends 3 messages quickly).
		expertLimiter:        ratelimit.NewTokenBucketLimiter(5, 2.0),
		// WHY 3: single admin user, 3 concurrent experts is the realistic max.
		// Prevents goroutine explosion when LLM calls take 2-10s each.
		expertMaxConcurrency: 3,
	}
}

// Process handles a client message through all selected experts.
//
// Mental execution:
// Client: "How should I design the user table?"
// Selected experts: [DB Expert, System Design Expert]
//
// Step 1: Load both experts from DB
// Step 2: Launch 2 goroutines simultaneously
//   Goroutine 1: DB Expert processes question
//   Goroutine 2: SD Expert processes question
// Step 3: Collect results (timeout: 120s)
// Step 4: If 2+ experts responded → synthesize
// Step 5: Update memory async
// Step 6: Return combined response
//
// WHY parallel (not sequential):
//   Chat flow = conversational Q&A. Each expert answers independently
//   from their own training corpus. No expert needs another's output
//   to answer a chat question — they have different knowledge domains.
//
//   Multi-agent COLLABORATION (where Expert B reads Expert A's output)
//   happens in the WORKFLOW flow (/api/v1/workflows), not here.
//   Workflow uses the Blackboard pattern — experts post artifacts,
//   others read them, OTA loop drives each expert.
func (o *Orchestrator) Process(ctx context.Context, req OrchestratorRequest) (*OrchestratorResponse, error) {
	start := time.Now()

	if len(req.ExpertIDs) == 0 {
		return nil, fmt.Errorf("no experts selected")
	}

	// Load expert records
	experts, err := o.loadExperts(ctx, req.ExpertIDs)
	if err != nil {
		return nil, fmt.Errorf("load experts failed: %w", err)
	}
	if len(experts) == 0 {
		return nil, fmt.Errorf("no active experts found")
	}

	// Run experts in parallel.
	// WHY parallel: each expert has independent knowledge corpus.
	// DB Expert and System Design Expert don't need each other's output
	// to answer a chat question — they answer from their own training.
	resultCh := make(chan ExpertResponse, len(experts))
	var wg sync.WaitGroup

	for _, exp := range experts {
		wg.Add(1)
		go func(expert expertRecord) {
			defer wg.Done()
			result := o.processWithExpert(ctx, req, expert)
			resultCh <- result
		}(exp)
	}

	// Close channel when all goroutines finish
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Collect results with timeout
	// WHY 120s timeout: LLM calls + self-learning steps can take 30-60s for complex problems.
	timeoutCtx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	var expertResponses []ExpertResponse
	for {
		select {
		case result, ok := <-resultCh:
			if !ok {
				goto collected
			}
			expertResponses = append(expertResponses, result)
		case <-timeoutCtx.Done():
			o.logger.Warn("orchestrator timeout", zap.Int("collected", len(expertResponses)))
			goto collected
		}
	}
collected:

	if len(expertResponses) == 0 {
		return nil, fmt.Errorf("all experts failed or timed out")
	}

	// Synthesize if multiple experts
	var synthesis *SynthesisResult
	if len(expertResponses) > 1 {
		synthesis = o.synthesize(expertResponses)
	}

	// Update memory async (non-blocking)
	go o.updateMemory(context.Background(), req, expertResponses)

	return &OrchestratorResponse{
		ExpertResponses: expertResponses,
		Synthesis:       synthesis,
		TurnNumber:      req.TurnNumber,
		DurationMs:      time.Since(start).Milliseconds(),
	}, nil
}

// processWithExpert runs one expert through the full pipeline.
// Uses PhaseTimer (Observer pattern) to record per-phase latency.
func (o *Orchestrator) processWithExpert(ctx context.Context, req OrchestratorRequest, expert expertRecord) ExpertResponse {
	timer := observability.NewPhaseTimer()

	// Per-expert rate limiting (Strategy pattern).
	// Prevents a single expert from being overwhelmed by concurrent requests.
	// expertLimiter is a TokenBucketLimiter in production, NoopLimiter in tests.
	if !o.expertLimiter.Allow(ctx, ratelimit.ExpertKey(expert.ID.String())) {
		o.logger.Warn("expert rate limit exceeded",
			zap.String("expert_id", expert.ID.String()),
			zap.String("expert_name", expert.Name),
		)
		return ExpertResponse{
			ExpertID:   expert.ID,
			ExpertName: expert.Name,
			Domain:     expert.Domain,
			Mode:       decision.ModeREFUSE,
			Content:    "Expert is busy. Please try again in a moment.",
			Error:      "rate_limit_exceeded",
		}
	}

	// Per-expert concurrency limit (Semaphore pattern).
	// Prevents goroutine explosion when LLM calls take 2-10s each.
	// Acquire semaphore slot. If full: return busy response immediately.
	// WHY non-blocking (select with default): we never want to block the
	// caller goroutine. Queuing would hide backpressure from the user.
	sem := o.getExpertSemaphore(expert.ID.String())
	select {
	case sem <- struct{}{}:
		// Slot acquired. Release when function returns.
		defer func() { <-sem }()
	default:
		// All slots occupied. Return busy immediately.
		o.logger.Warn("expert concurrency limit exceeded",
			zap.String("expert_id", expert.ID.String()),
			zap.String("expert_name", expert.Name),
			zap.Int("max_concurrency", o.expertMaxConcurrency),
		)
		return ExpertResponse{
			ExpertID:   expert.ID,
			ExpertName: expert.Name,
			Domain:     expert.Domain,
			Mode:       decision.ModeREFUSE,
			Content:    "Expert is handling too many requests. Please try again in a moment.",
			Error:      "concurrency_limit_exceeded",
		}
	}

	// Phase: context assembly (parallel fan-out inside Assemble)
	timer.Start("context_assembly")
	assembledCtx, err := o.assembler.Assemble(
		ctx, req.ChatID, req.ProjectID, expert.ID, req.Message, req.TurnNumber,
		req.ReplyToMessageID, req.IncludeFullThread,
	)
	timer.Stop("context_assembly")
	if err != nil {
		o.logger.Warn("context assembly failed", zap.String("expert", expert.Name), zap.Error(err))
		return ExpertResponse{
			ExpertID: expert.ID, ExpertName: expert.Name, Domain: expert.Domain,
			Mode: decision.ModeREFUSE, Content: "Context assembly failed", Error: err.Error(),
		}
	}

	// Get project summary for Gate 4
	projectSummary := ""
	if assembledCtx.RollingSummary != "" {
		projectSummary = assembledCtx.RollingSummary
	}

	// CT-B4: resolve category_id -> TemplateSections/DefaultLanguage.
	// Both stay nil/"" (flat-text path, CT-L2) unless ALL of:
	//   1. categoryRegistry was actually wired in (not nil)
	//   2. expert.CategoryID is non-nil
	//   3. that category exists in the cache AND has >=1 template section
	// Any of these being false is a normal, common state — not an error.
	var templateSections []category.TemplateSection
	defaultLanguage := ""
	askStructurePermission := false
	if o.categoryRegistry != nil && expert.CategoryID != nil {
		if cat := o.categoryRegistry.Get(*expert.CategoryID); cat != nil {
			askStructurePermission = cat.AskStructurePermission
			if len(cat.TemplateSchema.Sections) > 0 {
				templateSections = cat.TemplateSchema.Sections
				defaultLanguage = cat.DefaultLanguage
			}
		}
	}

	// Run decision engine. req.ReplyToMessageID (CT-C4) is nil for every
	// fresh question — gateStructurePermission (decision/engine.go) is a
	// no-op in that case regardless, since it also checks
	// expert.AskStructurePermission first.
	//
	// replyContext: pre-formatted string from assembledCtx.ReplyThread.
	// Empty string for fresh questions (no reply target) — enforcer
	// injects it into the LLM system prompt only when non-empty.
	// WHY format here not in enforcer: avoids importing appcontext
	// from chinawall (would create a circular dependency).
	replyContext := formatReplyContext(assembledCtx.ReplyThread)

	// Phase: self-learning
	// WHY self-learning is CRITICAL for DSA/problem-solving domains:
	//   DSA problems are deliberately story-wrapped:
	//   "Alice on chessboard, knight moves, find minimum moves"
	//   RAG searches for "chessboard", "Alice", "knight" → wrong chunks.
	//   Self-learning converts this to:
	//   "BFS shortest path on grid, minimum distance between two cells"
	//   → correct chunks retrieved.
	//   Problem-solving domains need self-learning MORE than factual domains,
	//   not less. Factual domains (medical/legal) have literal questions;
	//   DSA problems are always obfuscated with story noise.
	//
	// Previous (WRONG) logic: skip for IsProblemSolvingDomain.
	// Correct logic: run for ALL domains when selfLearning is non-nil.
	timer.Start("self_learning")
	questionForRAG := req.Message
	if o.selfLearning != nil {
		processed := o.selfLearning.Process(
			ctx,
			req.Message,
			expert.Name,
			expert.Domain,
			expert.ReasoningCharter,
			assembledCtx.CourseChunks,
		)
		if processed.VerificationPassed {
			questionForRAG = processed.Extracted
			o.logger.Info("self-learning: using extracted question",
				zap.String("expert", expert.Name),
				zap.String("domain", expert.Domain),
			)
		} else if processed.SkippedReason != "" {
			o.logger.Debug("self-learning: using original question",
				zap.String("expert", expert.Name),
				zap.String("reason", processed.SkippedReason),
			)
		}
	}
	timer.Stop("self_learning")

	// Phase: decision engine (Gates 1-5 including LLM generation)
	timer.Start("decision_engine")
	result, err := o.decisionEng.Process(
		ctx,
		questionForRAG,
		decision.Expert{
			ID:                     expert.ID,
			Name:                   expert.Name,
			Domain:                 expert.Domain,
			ReasoningCharter:       expert.ReasoningCharter,
			ClarificationCharter:   expert.ClarificationCharter,
			TemplateSections:       templateSections,
			DefaultLanguage:        defaultLanguage,
			AskStructurePermission: askStructurePermission,
		},
		assembledCtx.CourseChunks,
		projectSummary,
		replyContext,
		1,
		req.ReplyToMessageID,
		req.TokenCh,
	)
	timer.Stop("decision_engine")

	// Log phase breakdown for observability
	for _, p := range timer.Results() {
		o.logger.Info("pipeline phase timing",
			zap.String("expert", expert.Name),
			zap.String("phase", p.Phase),
			zap.Int64("ms", p.Duration.Milliseconds()),
		)
	}

	if err != nil {
		o.logger.Error("decision engine failed", zap.String("expert", expert.Name), zap.Error(err))
		return ExpertResponse{
			ExpertID: expert.ID, ExpertName: expert.Name, Domain: expert.Domain,
			Mode: decision.ModeREFUSE, Content: "Processing failed", Error: err.Error(),
		}
	}

	return ExpertResponse{
		ExpertID:    expert.ID,
		ExpertName:  expert.Name,
		Domain:      expert.Domain,
		Mode:        result.Mode,
		Content:     result.Content,
		Citations:   result.Citations,
		Confidence:  result.Confidence,
		GateStopped: result.GateStopped,
		Warning:     result.Warning,
		Questions:   result.Questions,
		TemplateSections: result.TemplateSections,
		// ReplyToUserMessageID (CT-C4): only set when this IS a
		// structure-permission ASK (sentinel GateStopped==-1, see
		// decision/engine.go's gateStructurePermission). userMsgID copy
		// taken here, not a pointer into req, so each goroutine gets its
		// own value — req is shared read-only across all expert goroutines
		// (Process launches one per expert), but taking &local avoids any
		// doubt about aliasing a shared struct's field across goroutines.
		ReplyToUserMessageID: structurePermissionAskParent(result.GateStopped, req.UserMessageID),
	}
}

// structurePermissionAskParent returns &userMessageID when gateStopped
// indicates this response is a structure-permission ASK (sentinel -1,
// decision/engine.go's gateStructurePermission), or nil for every other
// response. Named function (not inlined) so the sentinel-value meaning
// has exactly one place to live, per Step 5's "magic values -> named
// constants" rule (gateStopped==-1 itself is documented at its one
// source of truth, gateStructurePermission).
func structurePermissionAskParent(gateStopped int, userMessageID uuid.UUID) *uuid.UUID {
	if gateStopped != -1 {
		return nil
	}
	id := userMessageID
	return &id
}

// synthesize finds agreements and contradictions between expert responses.
func (o *Orchestrator) synthesize(responses []ExpertResponse) *SynthesisResult {
	// Simple synthesis: find ADVISE responses and note any WARN/REFUSE
	var advising []ExpertResponse
	var warnings []ExpertResponse

	for _, r := range responses {
		switch r.Mode {
		case decision.ModeADVISE:
			advising = append(advising, r)
		case decision.ModeWARN, decision.ModePUSHBACK:
			warnings = append(warnings, r)
		}
	}

	result := &SynthesisResult{}

	if len(advising) > 1 {
		result.Agreements = []string{"Multiple experts have relevant knowledge on this topic"}
	}

	for _, w := range warnings {
		for _, a := range advising {
			result.Contradictions = append(result.Contradictions, Contradiction{
				Topic:     "approach",
				ExpertA:   a.ExpertName,
				PositionA: "Proceed with implementation",
				ExpertB:   w.ExpertName,
				PositionB: w.Warning,
			})
		}
	}

	result.Summary = fmt.Sprintf("%d expert(s) responded. Review each response carefully.", len(responses))
	return result
}

// allSameDomain returns true when every expert in the slice shares the same domain.
// Used by Process() to decide between collaborative and parallel modes.
//
// WHY case-insensitive: domain field is user-entered (admin panel).
// "DSA" and "dsa" should be treated as the same domain.
func allSameDomain(experts []expertRecord) bool {
	if len(experts) < 2 {
		return false // single expert: no collaboration needed
	}
	first := strings.ToLower(experts[0].Domain)
	for _, e := range experts[1:] {
		if strings.ToLower(e.Domain) != first {
			return false
		}
	}
	return true
}

// processCollaborative runs same-domain experts in Plan-Execute sequence.
//
// PATTERN: Plan-Execute (Arpit Bhiyani AI Masterclass)
//   Expert 1 (Analyst): analyze problem, propose approach, identify key concepts
//   Expert 2 (Solver):  read Expert 1's analysis, produce enhanced final answer
//   Expert N:           each reads all previous outputs, adds its perspective
//
// WHY sequential not parallel:
//   Expert 2 NEEDS Expert 1's output as input.
//   Parallel would give 2 independent answers — not collaboration.
//
// WHY inject prior analysis as replyContext:
//   replyContext is already wired through the entire pipeline:
//   orchestrator → decision engine → chinawall enforcer → LLM system prompt.
//   Reusing this path means zero new wiring — Expert 2 sees Expert 1's
//   analysis as "prior context" in its system prompt, exactly like a
//   human expert reading a colleague's notes before answering.
//
// OUTPUT: single ExpertResponse from the last expert (the final answer).
// All intermediate analyses are included in the response content so the
// user can see the collaboration chain.
func (o *Orchestrator) processCollaborative(
	ctx context.Context,
	req OrchestratorRequest,
	experts []expertRecord,
	start time.Time,
) (*OrchestratorResponse, error) {
	o.logger.Info("collaborative mode: same-domain experts",
		zap.Int("expert_count", len(experts)),
		zap.String("domain", experts[0].Domain),
	)

	// collaborativeContext accumulates each expert's output.
	// Passed to the next expert as additional context.
	// Format: "[ExpertName Analysis]\n{content}\n"
	var collaborativeContext strings.Builder

	var expertResponses []ExpertResponse

	for i, expert := range experts {
		// Build request for this expert.
		// For Expert 2+: inject prior analyses as replyContext.
		// We do this by appending to the message — the simplest, most
		// reliable injection point that requires no new pipeline wiring.
		expertReq := req
		if i > 0 && collaborativeContext.Len() > 0 {
			// Inject prior expert analyses into the question.
			// WHY append to message not a separate field:
			//   self-learning reads req.Message for RAG query.
			//   context assembler reads req.Message for embedding.
			//   Both benefit from seeing the prior analysis.
			//   A separate field would require wiring through 5 layers.
			expertReq.Message = fmt.Sprintf(
				"%s\n\n--- PRIOR EXPERT ANALYSIS (read before answering) ---\n%s\n--- END PRIOR ANALYSIS ---",
				req.Message,
				collaborativeContext.String(),
			)
			// Only the last expert streams tokens (user sees final answer).
			// Intermediate experts run blocking (no streaming).
			expertReq.TokenCh = nil
		}
		// Last expert gets the original TokenCh for streaming.
		if i == len(experts)-1 {
			expertReq.TokenCh = req.TokenCh
		}

		result := o.processWithExpert(ctx, expertReq, expert)
		expertResponses = append(expertResponses, result)

		// If this expert produced a useful answer, add it to the
		// collaborative context for the next expert.
		if result.Error == "" && result.Content != "" &&
			result.Mode != decision.ModeREFUSE {
			collaborativeContext.WriteString(fmt.Sprintf(
				"[%s]:\n%s\n\n",
				expert.Name,
				result.Content,
			))
		}

		o.logger.Info("collaborative: expert completed",
			zap.Int("step", i+1),
			zap.Int("total", len(experts)),
			zap.String("expert", expert.Name),
			zap.String("mode", string(result.Mode)),
		)
	}

	// Update memory async (non-blocking)
	go o.updateMemory(context.Background(), req, expertResponses)

	return &OrchestratorResponse{
		ExpertResponses: expertResponses,
		TurnNumber:      req.TurnNumber,
		DurationMs:      time.Since(start).Milliseconds(),
	}, nil
}

// updateMemory records the turn in all memory levels.
func (o *Orchestrator) updateMemory(ctx context.Context, req OrchestratorRequest, responses []ExpertResponse) {
	for _, resp := range responses {
		if resp.Error != "" {
			continue
		}
		importance := 3
		if resp.Mode == decision.ModeADVISE {
			importance = 4
		}
		// Bug 3.2 fix (docs bug list): pass nil, not uuid.New(). The real
		// assistant message row does not exist yet at this point (it is
		// saved separately by message/handler.go's saveAssistantMessage,
		// possibly in a goroutine that has not completed) - a fabricated
		// random UUID here violated master_event_log's message_id FK on
		// every turn. RecordTurn's messageID param is now *uuid.UUID
		// (nullable), matching the nullable FK column exactly.
		o.memManager.RecordTurn(
			ctx,
			req.ProjectID, resp.ExpertID, req.ClientID,
			req.ChatID, nil,
			req.TurnNumber,
			req.Message, resp.Content,
			string(resp.Mode), importance,
		)
	}
}

// loadExperts fetches expert records from DB.
func (o *Orchestrator) loadExperts(ctx context.Context, expertIDs []uuid.UUID) ([]expertRecord, error) {
	var experts []expertRecord
	for _, id := range expertIDs {
		var e expertRecord
		var clarJSON []byte
		err := o.db.QueryRow(ctx,
			`SELECT id, name, domain, COALESCE(reasoning_charter,''), clarification_charter, category_id
			 FROM experts
			 WHERE id=$1 AND is_active=TRUE AND is_training=FALSE AND deleted_at IS NULL`,
			id,
		).Scan(&e.ID, &e.Name, &e.Domain, &e.ReasoningCharter, &clarJSON, &e.CategoryID)
		if err != nil {
			o.logger.Warn("expert not found or inactive", zap.String("id", id.String()))
			continue
		}
		// Parse clarification charter
		if len(clarJSON) > 0 {
			_ = parseJSON(clarJSON, &e.ClarificationCharter)
		}
		experts = append(experts, e)
	}
	return experts, nil
}

// formatReplyContext converts a reply thread into a pre-formatted string
// for injection into the LLM system prompt.
//
// WHY format here (not in enforcer/chinawall):
//   appcontext.ReplyThreadEntry lives in the context package.
//   chinawall imports context would create a circular dependency
//   (context already imports chinawall for CourseChunk).
//   Formatting here keeps the dependency direction clean.
//
// Returns empty string for fresh questions (no reply thread) —
// enforcer skips injection when empty.
func formatReplyContext(thread []appcontext.ReplyThreadEntry) string {
	if len(thread) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("## REPLY CONTEXT (the user is replying to this prior message):\n")
	for _, entry := range thread {
		role := "ASSISTANT"
		if entry.Role == "user" {
			role = "USER"
		}
		sb.WriteString(fmt.Sprintf("%s (turn %d):\n%s\n\n", role, entry.TurnNumber, entry.Content))
	}
	sb.WriteString("Answer the user's follow-up question with full awareness of the above context.\n")
	return sb.String()
}

// parseJSON unmarshals raw JSON bytes into dst.
// Used by loadExperts to decode clarification_charter JSONB column.
// Non-fatal: if unmarshal fails, dst is left at its zero value.
func parseJSON(data []byte, dst interface{}) error {
	return json.Unmarshal(data, dst)
}

// getExpertSemaphore returns the semaphore channel for the given expert ID.
// Creates a new buffered channel lazily on first access.
//
// WHY sync.Map.LoadOrStore:
//   Multiple goroutines may call this simultaneously for the same expert.
//   LoadOrStore is atomic — only one channel is ever created per expert.
//   The "loser" goroutine discards its newly created channel and uses
//   the winner's channel. No mutex needed.
//
// WHY buffered channel as semaphore:
//   Buffered channel of size N = semaphore with N slots.
//   Send = acquire. Receive = release.
//   Non-blocking select in caller = try-acquire without waiting.
func (o *Orchestrator) getExpertSemaphore(expertID string) chan struct{} {
	newSem := make(chan struct{}, o.expertMaxConcurrency)
	actual, _ := o.expertSemaphores.LoadOrStore(expertID, newSem)
	return actual.(chan struct{})
}
