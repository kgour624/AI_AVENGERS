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

	"ai_avengers/backend/internal/category"
	"ai_avengers/backend/internal/chinawall"
	appcontext "ai_avengers/backend/internal/context"
	"ai_avengers/backend/internal/decision"
	"ai_avengers/backend/internal/gateway"
	"ai_avengers/backend/internal/memory"
	"ai_avengers/backend/internal/observability"
	"ai_avengers/backend/internal/ratelimit"
	"ai_avengers/backend/internal/selflearning"
	"ai_avengers/backend/internal/usage"
)

// OrchestratorRequest is the input to the orchestrator.
type OrchestratorRequest struct {
	ProjectID  uuid.UUID
	ClientID   uuid.UUID
	ChatID     uuid.UUID
	Message    string
	ExpertIDs  []uuid.UUID
	TurnNumber int
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
	// TemplateName (T-CAT): optional named answer format from the expert's
	// category ("Code" / "Approach" / ...). Empty = the category default.
	TemplateName string
	// GenericAllowancePct (0-30): the client's generic ceiling for this
	// message. 0 = strict China Wall (the default, unchanged behaviour).
	GenericAllowancePct float64
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
	ExpertID    uuid.UUID             `json:"expert_id"`
	ExpertName  string                `json:"expert_name"`
	Domain      string                `json:"domain"`
	Mode        decision.ResponseMode `json:"mode"`
	Content     string                `json:"content"`
	Citations   []chinawall.Citation  `json:"citations"`
	Confidence  float64               `json:"confidence"`
	GateStopped int                   `json:"gate_stopped"`
	Warning     string                `json:"warning,omitempty"`
	Questions   []string              `json:"questions,omitempty"`
	Error       string                `json:"error,omitempty"`
	// TemplateSections (CT-B4): populated only when this expert has a
	// category with a non-empty template_schema. nil for every flat-text
	// expert response (CT-L2) — frontend (CT-D5, not yet built) must
	// check len(TemplateSections) > 0 before rendering structured UI,
	// falling back to plain Content otherwise, exactly like every layer
	// below this one already does.
	TemplateSections []chinawall.TemplateSectionResult `json:"template_sections,omitempty"`
	// Claims (B8): atomic claim→evidence reports. nil when claim verify
	// off / fail-open / structured. Frontend may render labels already
	// embedded in Content; this is the structured form for C1 provenance.
	Claims []chinawall.ClaimReport `json:"claims,omitempty"`
	// Coverage (C9): China Wall coverage verdict YES|PARTIAL|NO. "" when no
	// generation happened. Surfaced by the explanation view.
	Coverage string `json:"coverage,omitempty"`
	// QualityScore (C9): B6 judge overall [0,1]; 0 when the judge was
	// off/failed-open. Explanation/observability only.
	QualityScore float64 `json:"quality_score,omitempty"`
	// Reason (C9): Gate-5 / partial refusal explanation. "" unless a refusal
	// carried a China Wall reason. Surfaced by the explanation view.
	Reason string `json:"reason,omitempty"`
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
	Agreements     []string        `json:"agreements"`
	Contradictions []Contradiction `json:"contradictions"`
	Summary        string          `json:"summary"`
	// Method (B1): "llm" when the real synthesis call succeeded, "fallback"
	// when it was skipped/failed and the deterministic non-LLM merge ran
	// instead (§3.1 P3 — a synthesis failure must degrade, not crash chat).
	Method string `json:"method"`
	// NeedsEscalation (B2): true when at least one contradiction's
	// Resolution is escalate — the client must decide rather than just read
	// a flagged note. Derived from len(Escalations) > 0, never set
	// independently, so it can never disagree with the per-item policies.
	NeedsEscalation bool `json:"needs_escalation"`
	// Escalations (B2b): the subset of Contradictions whose Resolution is
	// escalate, surfaced separately so the frontend does not have to filter
	// Contradictions itself to build a "needs your decision" banner. Chat
	// has no approval_requests table (that is workflow-only, G1) — the
	// client acts by replying, same as a Gate 1 ASK's Questions.
	Escalations []Contradiction `json:"escalations"`
	// EscalationSummary is a one-line human summary of Escalations, empty
	// when there are none.
	EscalationSummary string `json:"escalation_summary,omitempty"`
}

// Contradiction classification values (B2). Kept as named constants so the
// policy is judgeable in one place, not scattered as string literals.
const (
	// ContradictionTypeContextConflict (Type-1): one expert's position
	// conflicts with the provided context/evidence — catchable, so it can
	// often be resolved deterministically against the source.
	ContradictionTypeContextConflict = "context_conflict"
	// ContradictionTypeFabrication (Type-2): an expert asserts something
	// with no supporting evidence — hard to catch, so it is treated with
	// more caution (never auto-noted away).
	ContradictionTypeFabrication = "fabrication"

	// ResolutionNoted: a real but minor difference the client can read past.
	ResolutionNoted = "noted"
	// ResolutionEscalate: the disagreement needs a client decision.
	ResolutionEscalate = "escalate"
)

// Contradiction is a point where two experts disagree (B2). Beyond the two
// positions it now carries the classification (Type) and the resolution
// policy (Resolution) the synthesis decided for this specific disagreement.
type Contradiction struct {
	Topic     string `json:"topic"`
	ExpertA   string `json:"expert_a"`
	PositionA string `json:"position_a"`
	ExpertB   string `json:"expert_b"`
	PositionB string `json:"position_b"`
	// Type is one of ContradictionType* — what kind of disagreement this is.
	// Defaults to ContradictionTypeFabrication when the model omitted or
	// returned an unknown value (fail-closed: treat as the harder case).
	Type string `json:"type"`
	// Resolution is one of Resolution* — what should happen. Defaults to
	// ResolutionEscalate when missing/unknown (fail-closed §3.1 P3).
	Resolution string `json:"resolution"`
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
	// gw is used by synthesize (B1) for the real LLM synthesis call.
	// Never nil in production (NewOrchestrator requires it); a nil gw only
	// happens in tests that construct Orchestrator{} directly, and
	// synthesize's nil-check falls back to the deterministic non-LLM merge
	// so those tests keep working unchanged.
	gw *gateway.ModelGateway
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
	logger               *zap.Logger
}

// NewOrchestrator creates a new orchestrator.
func NewOrchestrator(
	db *pgxpool.Pool,
	assembler *appcontext.Assembler,
	decisionEng *decision.Engine,
	memManager *memory.Manager,
	categoryRegistry *category.Registry,
	selfLearning *selflearning.QuestionProcessor,
	gw *gateway.ModelGateway,
	logger *zap.Logger,
) *Orchestrator {
	return &Orchestrator{
		db:               db,
		assembler:        assembler,
		decisionEng:      decisionEng,
		memManager:       memManager,
		categoryRegistry: categoryRegistry,
		selfLearning:     selfLearning,
		gw:               gw,
		logger:           logger,
		// Default: 5 burst, 2 requests/second per expert.
		// WHY these numbers: LLM providers typically allow 5-10 RPM per key.
		// 2 req/s = 120 RPM, well within provider limits.
		// Burst=5 allows short spikes (e.g. user sends 3 messages quickly).
		expertLimiter: ratelimit.NewTokenBucketLimiter(5, 2.0),
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
//
//	Goroutine 1: DB Expert processes question
//	Goroutine 2: SD Expert processes question
//
// Step 3: Collect results (timeout: 120s)
// Step 4: If 2+ experts responded → synthesize
// Step 5: Update memory async
// Step 6: Return combined response
//
// WHY parallel (not sequential):
//
//	Chat flow = conversational Q&A. Each expert answers independently
//	from their own training corpus. No expert needs another's output
//	to answer a chat question — they have different knowledge domains.
//
//	Multi-agent COLLABORATION (where Expert B reads Expert A's output)
//	happens in the WORKFLOW flow (/api/v1/workflows), not here.
//	Workflow uses the Blackboard pattern — experts post artifacts,
//	others read them, OTA loop drives each expert.
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
	timeoutCtx, cancel := context.WithTimeout(ctx, 300*time.Second)
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

	// Synthesize if multiple experts. Uses timeoutCtx (not the outer ctx)
	// so a slow synthesis call cannot run past the same 120s budget the
	// expert collection above is already bound to.
	var synthesis *SynthesisResult
	if len(expertResponses) > 1 {
		// C5: attribute the synthesis LLM call to this chat/project.
		spid, scid := req.ProjectID, req.ChatID
		synthCtx := usage.WithAttribution(timeoutCtx, usage.Attribution{
			ProjectID: &spid, ChatID: &scid, UseCase: usage.UseCaseSynthesis,
		})
		synthesis = o.synthesize(synthCtx, expertResponses)
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

	// C5: attribute every downstream LLM call in this expert's pipeline
	// (self-learning, gates, generation) to this project/chat/expert so the
	// usage surface groups spend correctly. Per-tenant spend is derived from
	// the project at write time.
	pid, cid, eid := req.ProjectID, req.ChatID, expert.ID
	ctx = usage.WithAttribution(ctx, usage.Attribution{
		ProjectID: &pid, ChatID: &cid, ExpertID: &eid, UseCase: usage.UseCaseChat,
	})

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
			// ResolveSections also handles the named-variant case, so the same
			// expert can answer "Code" or "Approach" per request; an unknown or
			// empty name falls back to the category default, then to the legacy
			// single template — never an error.
			if sections, _ := cat.ResolveSections(req.TemplateName); len(sections) > 0 {
				templateSections = sections
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
		// B5: the dead CourseChunks argument was removed — it was
		// retrieved with the ORIGINAL (story-noisy) question, so passing
		// it into the extractor would reinforce wrong retrieval. The
		// extracted question is what the decision engine re-retrieves with.
		processed := o.selfLearning.Process(
			ctx,
			req.Message,
			expert.Name,
			expert.Domain,
			expert.ReasoningCharter,
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
			GenericAllowancePct:    req.GenericAllowancePct,
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
		ExpertID:         expert.ID,
		ExpertName:       expert.Name,
		Domain:           expert.Domain,
		Mode:             result.Mode,
		Content:          result.Content,
		Citations:        result.Citations,
		Confidence:       result.Confidence,
		GateStopped:      result.GateStopped,
		Warning:          result.Warning,
		Questions:        result.Questions,
		TemplateSections: result.TemplateSections,
		Claims:           result.Claims,
		Coverage:         result.Coverage,
		QualityScore:     result.QualityScore,
		Reason:           result.Reason,
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

// synthesize merges multiple experts' actual answers into one coherent view
// (B1). Previously this built pseudo-agreements/contradictions from Mode
// alone (ADVISE vs WARN) without ever reading the answer text — two experts
// giving the same advice in different words showed up as "agreement" with no
// content, and a real semantic disagreement between two ADVISE responses was
// invisible.
//
// One dedicated LLM call reads every expert's Content and returns a strict
// JSON verdict (§3.1 P1: structured output, not prose we regex). Low
// temperature, single pass (P11) — this is a synthesis/aggregation step, not
// a reasoning task. On any failure (gw nil, call error, malformed JSON) it
// falls back to the deterministic non-LLM merge (P3: degrade, don't crash
// chat over a synthesis failure) — synthesizeFallback below is the original
// logic, unchanged, so existing tests/behaviour for that path still hold.
func (o *Orchestrator) synthesize(ctx context.Context, responses []ExpertResponse) *SynthesisResult {
	if o.gw == nil {
		return o.synthesizeFallback(responses)
	}

	var sb strings.Builder
	sb.WriteString("Multiple domain experts answered the same question independently. ")
	sb.WriteString("Read their actual answers below and produce a synthesis.\n\n")
	for i, r := range responses {
		content := r.Content
		if len(content) > 1500 {
			content = content[:1500] + "\n... [truncated]"
		}
		fmt.Fprintf(&sb, "--- Expert %d: %s (domain: %s) ---\n%s\n\n", i+1, r.ExpertName, r.Domain, content)
	}
	sb.WriteString(
		"Treat every expert's answer as a final verdict from their domain — do not " +
			"discard or downweight any of them just because they differ from the majority.\n\n" +
			"For each disagreement, classify it:\n" +
			`  - "type": "context_conflict" if one position contradicts the provided ` +
			"question/context, otherwise \"fabrication\" if a position has no supporting evidence.\n" +
			`  - "resolution": "escalate" if the client must decide (the two positions are ` +
			`materially incompatible or high-stakes), otherwise "noted" if it is a minor ` +
			"difference the client can read past. When unsure, choose 'escalate'.\n\n" +
			"Return JSON only, no prose outside the JSON:\n" +
			`{"agreements": ["point experts agree on", ...], ` +
			`"disagreements": [{"topic": "...", "expert_a": "name", "position_a": "...", ` +
			`"expert_b": "name", "position_b": "...", "type": "context_conflict|fabrication", ` +
			`"resolution": "noted|escalate"}], ` +
			`"summary": "one paragraph synthesis for the client"}` +
			"\nIf there are no real disagreements, disagreements must be an empty array — do not invent one.",
	)

	resp, err := o.gw.Call(ctx, gateway.LLMRequest{
		Model:       gateway.ModelCheap,
		UserPrompt:  sb.String(),
		MaxTokens:   700,
		Temperature: 0.1,
	})
	if err != nil {
		o.logger.Warn("synthesize: LLM call failed, using fallback merge", zap.Error(err))
		return o.synthesizeFallback(responses)
	}

	var parsed struct {
		Agreements    []string `json:"agreements"`
		Disagreements []struct {
			Topic      string `json:"topic"`
			ExpertA    string `json:"expert_a"`
			PositionA  string `json:"position_a"`
			ExpertB    string `json:"expert_b"`
			PositionB  string `json:"position_b"`
			Type       string `json:"type"`
			Resolution string `json:"resolution"`
		} `json:"disagreements"`
		Summary string `json:"summary"`
	}
	clean := strings.TrimSpace(resp.Content)
	clean = strings.TrimPrefix(clean, "```json")
	clean = strings.TrimPrefix(clean, "```")
	clean = strings.TrimSuffix(clean, "```")
	clean = strings.TrimSpace(clean)
	start := strings.Index(clean, "{")
	end := strings.LastIndex(clean, "}")
	if start == -1 || end == -1 || end < start {
		o.logger.Warn("synthesize: LLM response had no JSON object, using fallback merge")
		return o.synthesizeFallback(responses)
	}
	if err := json.Unmarshal([]byte(clean[start:end+1]), &parsed); err != nil {
		o.logger.Warn("synthesize: LLM response JSON parse failed, using fallback merge", zap.Error(err))
		return o.synthesizeFallback(responses)
	}

	result := &SynthesisResult{
		Agreements: parsed.Agreements,
		Summary:    parsed.Summary,
		Method:     "llm",
	}
	for _, d := range parsed.Disagreements {
		c := Contradiction{
			Topic:      d.Topic,
			ExpertA:    d.ExpertA,
			PositionA:  d.PositionA,
			ExpertB:    d.ExpertB,
			PositionB:  d.PositionB,
			Type:       normalizeContradictionType(d.Type),
			Resolution: normalizeResolution(d.Resolution),
		}
		// Fail-closed (B2/§3.1 P3): a fabrication-typed disagreement is
		// never silently "noted" away — force escalation for it even if
		// the model said noted. context_conflict keeps the model's policy.
		if c.Type == ContradictionTypeFabrication && c.Resolution == ResolutionNoted {
			c.Resolution = ResolutionEscalate
		}
		result.Contradictions = append(result.Contradictions, c)
	}
	if result.Summary == "" {
		result.Summary = fmt.Sprintf("%d expert(s) responded.", len(responses))
	}
	applyEscalations(result)
	return result
}

// applyEscalations derives Escalations/EscalationSummary/NeedsEscalation
// from Contradictions (B2b) — the single place that decides what counts as
// "needs a client decision", so synthesize() and synthesizeFallback() can
// never disagree with each other about it.
func applyEscalations(result *SynthesisResult) {
	for _, c := range result.Contradictions {
		if c.Resolution == ResolutionEscalate {
			result.Escalations = append(result.Escalations, c)
		}
	}
	result.NeedsEscalation = len(result.Escalations) > 0
	if result.NeedsEscalation {
		result.EscalationSummary = fmt.Sprintf(
			"%d point(s) need your decision: %s",
			len(result.Escalations),
			escalationTopics(result.Escalations),
		)
	}
}

// escalationTopics joins each escalation's Topic (falling back to
// "ExpertA vs ExpertB" when Topic is empty) into a short comma-separated
// list for EscalationSummary.
func escalationTopics(escalations []Contradiction) string {
	topics := make([]string, 0, len(escalations))
	for _, e := range escalations {
		t := e.Topic
		if t == "" {
			t = fmt.Sprintf("%s vs %s", e.ExpertA, e.ExpertB)
		}
		topics = append(topics, t)
	}
	return strings.Join(topics, "; ")
}

// normalizeContradictionType maps the model's raw "type" to a known value,
// defaulting to fabrication (the harder case) when missing or unrecognised
// — fail-closed: never let an unknown value be treated as the mild case.
func normalizeContradictionType(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case ContradictionTypeContextConflict:
		return ContradictionTypeContextConflict
	case ContradictionTypeFabrication:
		return ContradictionTypeFabrication
	default:
		return ContradictionTypeFabrication
	}
}

// normalizeResolution maps the model's raw "resolution" to a known value,
// defaulting to escalate when missing or unrecognised — fail-closed: an
// unclassified disagreement needs a client decision, not a silent note.
func normalizeResolution(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case ResolutionNoted:
		return ResolutionNoted
	case ResolutionEscalate:
		return ResolutionEscalate
	default:
		return ResolutionEscalate
	}
}

// synthesizeFallback is the original heuristic merge: notes that multiple
// experts responded and pairs any WARN/PUSHBACK against ADVISE responses as
// a contradiction. Used only when the real LLM synthesis (above) is
// unavailable or fails — never the primary path anymore, but kept exactly as
// it was so the degraded case has known, tested behaviour.
func (o *Orchestrator) synthesizeFallback(responses []ExpertResponse) *SynthesisResult {
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

	result := &SynthesisResult{Method: "fallback"}

	if len(advising) > 1 {
		result.Agreements = []string{"Multiple experts have relevant knowledge on this topic"}
	}

	for _, w := range warnings {
		for _, a := range advising {
			// Fail-closed policy for the degraded path too (B2): a
			// warn-vs-advise split is treated as an escalate-worthy
			// context conflict rather than silently noted.
			result.Contradictions = append(result.Contradictions, Contradiction{
				Topic:      "approach",
				ExpertA:    a.ExpertName,
				PositionA:  "Proceed with implementation",
				ExpertB:    w.ExpertName,
				PositionB:  w.Warning,
				Type:       ContradictionTypeContextConflict,
				Resolution: ResolutionEscalate,
			})
		}
	}

	result.Summary = fmt.Sprintf("%d expert(s) responded. Review each response carefully.", len(responses))
	applyEscalations(result)
	return result
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
		// Bug 3.2 fix (docs bug list): pass nil messageID, not uuid.New().
		// The real assistant message row does not exist yet at this point
		// (it is saved separately by message/handler.go's
		// saveAssistantMessage, possibly in a goroutine that has not
		// completed) - a fabricated random UUID here violated
		// master_event_log's message_id FK on every turn. RecordTurn's
		// messageID param is *uuid.UUID (nullable), matching the nullable
		// FK column exactly. B3: chatID is also *uuid.UUID — chat path
		// still passes the real chat; workflow path will pass nil.
		chatID := req.ChatID
		o.memManager.RecordTurn(
			ctx,
			req.ProjectID, resp.ExpertID, req.ClientID,
			&chatID, nil,
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
//
//	appcontext.ReplyThreadEntry lives in the context package.
//	chinawall imports context would create a circular dependency
//	(context already imports chinawall for CourseChunk).
//	Formatting here keeps the dependency direction clean.
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
//
//	Multiple goroutines may call this simultaneously for the same expert.
//	LoadOrStore is atomic — only one channel is ever created per expert.
//	The "loser" goroutine discards its newly created channel and uses
//	the winner's channel. No mutex needed.
//
// WHY buffered channel as semaphore:
//
//	Buffered channel of size N = semaphore with N slots.
//	Send = acquire. Receive = release.
//	Non-blocking select in caller = try-acquire without waiting.
func (o *Orchestrator) getExpertSemaphore(expertID string) chan struct{} {
	newSem := make(chan struct{}, o.expertMaxConcurrency)
	actual, _ := o.expertSemaphores.LoadOrStore(expertID, newSem)
	return actual.(chan struct{})
}
