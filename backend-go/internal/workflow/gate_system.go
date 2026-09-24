package workflow

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	appcontext "ai_avengers/backend/internal/context"
	"ai_avengers/backend/internal/chinawall"
)

// Gate 1 works in two relevance bands instead of one pass/fail line.
//
// WHY the band exists: with a single 0.70 cutoff, real runs on correctly
// trained experts scored 0.504 and 0.668 as their BEST chunk. Everything was
// therefore discarded and the answer fell through to fully generic — the
// trained material was retrieved from the DB and then thrown away, which is
// the opposite of what a domain-expert platform is for.
//
// gate1StrongThreshold: own training covers the task well enough that
// generic knowledge is blocked outright and no coverage check is needed.
const gate1StrongThreshold = 0.70

// gate1UsableThreshold: own training is relevant enough to apply as
// principles. It enters the prompt, peers are still consulted, and generic
// knowledge is permitted ONLY for what neither covers — tagged [GENERIC] so
// the gap is visible to the admin instead of silently invented.
//
// WHY 0.40: the observed real scores for on-topic trained material sat in
// the 0.50-0.67 range, and a cross-encoder score below ~0.4 is weak enough
// that forcing the expert to build on it produces worse answers than letting
// it reason. Tune with the measured top_score in the Gate 1 log.
const gate1UsableThreshold = 0.40

// gate1MaxChunks caps how many own-training chunks reach the prompt, so that
// weak-but-passing chunks cannot crowd out the strongest ones. Chunks arrive
// sorted best-first, so this keeps the top N.
const gate1MaxChunks = 5

// gate2PollTimeout: how long Gate 2 waits for peer responses.
// WHY 10s: enough for parallel DB queries, not so long it blocks workflow.
// Peers that don't respond within 10s are skipped — partial knowledge
// is better than blocking forever.
const gate2PollTimeout = 10 * time.Second

// GateResult is the output of RunGates.
// AgentLoop uses this to build the LLM context.
type GateResult struct {
	// TrainingChunks: expert's own training chunks (Gate 1 result).
	// Non-empty when at least one chunk scored >= gate1UsableThreshold,
	// capped at gate1MaxChunks and ordered best-first.
	TrainingChunks []chinawall.CourseChunk

	// PeerContributions: knowledge from other experts (Gate 2 result).
	// Non-empty when Gate 1 fails but peers have relevant knowledge.
	PeerContributions []PeerContribution

	// GenericAllowed: true when own training + peer knowledge together do
	// NOT cover the task. Always false when any chunk reached
	// gate1StrongThreshold. Gate 3 is only enabled when this is true.
	GenericAllowed bool

	// CoverageGap: the task text the generic allowance applies to.
	// Empty when GenericAllowed=false.
	CoverageGap string

	// Gate1Passed: true when the expert's own training produced usable chunks.
	Gate1Passed bool

	// GenericAllowancePct: the client's ceiling (0-30), copied from the
	// workflow so the prompt can state the exact limit.
	GenericAllowancePct float64
}

// PeerContribution is one expert's knowledge contribution in Gate 2.
type PeerContribution struct {
	ExpertID   uuid.UUID
	ExpertName string
	Chunks     []chinawall.CourseChunk
}

// GateSystem implements the 3-gate knowledge access control.
//
// DESIGN: "Pehle Ghar mein Dhoondo, Fir Dost se Pucho, Fir Google Karo"
//
// Gate 1 — Own training (Vector DB), two bands:
//   >= gate1StrongThreshold (0.70): training covers the task -> Generic
//     BLOCKED outright, no coverage check needed.
//   >= gate1UsableThreshold (0.40): training is relevant as principles ->
//     it goes into the prompt (top gate1MaxChunks) and the coverage check
//     decides whether any generic filling is permitted.
//   Below both: no own training in the prompt.
//   Citations: [CHUNK_uuid].
//
// Gate 2 — Peer knowledge (Blackboard Poll) — ALWAYS runs:
//   All other workflow experts' chunks searched in parallel (10s timeout).
//   Runs even when Gate 1 found strong training, because cross-expert
//   collaboration is mandatory: an expert must see what its peers were
//   trained on for the same task, not just its own material.
//   Coverage check (own training + peers): does the combined knowledge
//   cover the task? YES -> Generic BLOCKED. PARTIAL/NO -> Gate 3 enabled.
//
// Gate 3 — Gap filling (Generic, restricted):
//   Only enabled when own training AND peers together leave a gap.
//   Prompt: "Fill ONLY this gap: [gap]. Minimal generic knowledge."
//   Generic claims saved to pending_experience for admin review, so an
//   uncovered area becomes visible training debt instead of a silent guess.
//
// SOLID:
//   SRP: GateSystem only decides knowledge access, doesn't generate answers.
//   OCP: New gates can be added without changing AgentLoop.
//   DIP: AgentLoop depends on GateResult (interface), not gate internals.
// No ModelGateway field: GateSystem no longer makes any LLM call. The only
// one it had was the coverage check, and that decision now belongs to the
// client (workflows.generic_allowance_pct). Keeping an unused LLM dependency
// here would misrepresent what this type does.
// GateThresholds is the usable/strong pair Gate 1 compares rerank scores
// against. Defaults equal the historical package constants so a missing
// DB row produces byte-identical behaviour to pre-B4.
type GateThresholds struct {
	Usable float64
	Strong float64
	// Source is "default" when the package constants are used, otherwise
	// the gate_thresholds.source of the applied/manual row that won.
	Source string
	Domain string
}

// DefaultGateThresholds returns the historical constants. Public so the
// calibration path and admin GET can surface the baseline.
func DefaultGateThresholds() GateThresholds {
	return GateThresholds{
		Usable: gate1UsableThreshold,
		Strong: gate1StrongThreshold,
		Source: "default",
	}
}

// thresholdCacheTTL: how long a per-domain lookup stays hot. Admin apply
// after a calibration becomes visible within this window without a restart.
const thresholdCacheTTL = 30 * time.Second

type cachedThresholds struct {
	t       GateThresholds
	expires time.Time
}

type GateSystem struct {
	assembler *appcontext.Assembler
	// db (B4): optional. When set, thresholdsFor looks up gate_thresholds
	// for applied/manual rows. Nil keeps the historical const path —
	// AgentLoop that has no pool still works.
	db     *pgxpool.Pool
	logger *zap.Logger

	// thrCache avoids a DB round-trip on every expert task. Keyed by
	// domain (empty key = global default, reserved for future).
	thrMu    sync.RWMutex
	thrCache map[string]cachedThresholds
}

// NewGateSystem creates a new GateSystem.
// db may be nil (const-only mode); pass the pool in production so per-domain
// applied/manual overrides from gate_thresholds are honoured (B4/P8).
func NewGateSystem(assembler *appcontext.Assembler, db *pgxpool.Pool, logger *zap.Logger) *GateSystem {
	return &GateSystem{
		assembler: assembler,
		db:        db,
		logger:    logger,
		thrCache:  make(map[string]cachedThresholds),
	}
}

// thresholdsFor returns the Gate 1 usable/strong pair for a domain.
// Order of preference: cache hit → applied/manual DB row → package defaults.
// Never errors out to the caller — a DB hiccup falls back to defaults so a
// config outage cannot abort a wave.
func (g *GateSystem) thresholdsFor(ctx context.Context, domain string) GateThresholds {
	def := DefaultGateThresholds()
	def.Domain = domain
	if g == nil {
		return def
	}

	key := strings.ToLower(strings.TrimSpace(domain))
	now := time.Now()
	g.thrMu.RLock()
	if hit, ok := g.thrCache[key]; ok && now.Before(hit.expires) {
		g.thrMu.RUnlock()
		return hit.t
	}
	g.thrMu.RUnlock()

	t := def
	if g.db != nil && key != "" {
		var usable, strong float64
		var source string
		err := g.db.QueryRow(ctx, `
			SELECT usable, strong, source
			  FROM gate_thresholds
			 WHERE lower(domain) = $1
			   AND source IN ('applied', 'manual')
			 LIMIT 1`, key).Scan(&usable, &strong, &source)
		if err == nil && strong > usable && usable > 0 {
			t = GateThresholds{
				Usable: usable,
				Strong: strong,
				Source: source,
				Domain: domain,
			}
		}
		// err != nil (no row / table missing / etc.) → keep defaults.
	}

	g.thrMu.Lock()
	g.thrCache[key] = cachedThresholds{t: t, expires: now.Add(thresholdCacheTTL)}
	g.thrMu.Unlock()
	return t
}

// InvalidateThresholdCache drops the in-process cache so the next
// thresholdsFor re-reads the DB. Called by admin apply/PATCH.
func (g *GateSystem) InvalidateThresholdCache() {
	if g == nil {
		return
	}
	g.thrMu.Lock()
	g.thrCache = make(map[string]cachedThresholds)
	g.thrMu.Unlock()
}

// RunGates executes Gate 1 → Gate 2 → Gate 3 decision.
//
// Mental execution:
//   Expert: System Design, Task: "Design rate limiter for URL shortener"
//   allExperts: [PM, System Design, LLD, DSA]
//
//   Gate 1:
//     Search System Design's chunks for "rate limiter"
//     Result: [token bucket chunk, score=0.82] -> PASS
//     Return: GateResult{TrainingChunks: [token_bucket], Gate1Passed: true}
//
//   Gate 1 FAIL scenario:
//     No chunks with score >= 0.70
//     -> Gate 2:
//       Parallel search: PM chunks, LLD chunks, DSA chunks (10s timeout)
//       PM: [api_design chunk, score=0.65]
//       LLD: [rate_limiter_pattern chunk, score=0.78]
//       Coverage check: "PARTIAL — missing: distributed rate limiting"
//       -> Gate 3 enabled, CoverageGap = "distributed rate limiting"
func (g *GateSystem) RunGates(
	ctx context.Context,
	workflowID uuid.UUID,
	expert workflowExpert,
	taskDescription string,
	allExperts []workflowExpert,
	genericAllowancePct float64,
) (*GateResult, error) {
	result := &GateResult{}

	// ============================================================
	// GATE 1: Expert's own training
	// ============================================================
	g.logger.Debug("gate system: running Gate 1",
		zap.String("expert", expert.Name),
	)

	ownChunks, err := g.assembler.GetCourseChunksForWorkflow(
		ctx, expert.ID, taskDescription, 10,
	)
	if err != nil {
		g.logger.Warn("gate system: Gate 1 fetch failed",
			zap.String("expert", expert.Name),
			zap.Error(err),
		)
		// Non-fatal: treat as Gate 1 fail, proceed to Gate 2
	}

	// B4: per-domain thresholds (applied/manual DB row, else package
	// defaults). Resolved once per RunGates so log + band split agree.
	thr := g.thresholdsFor(ctx, expert.Domain)

	// Split own training into the two bands. ownChunks arrive sorted by
	// rerank score, best first.
	var strongChunks, usableChunks []chinawall.CourseChunk
	topScore := 0.0
	allExactlyHalf := len(ownChunks) > 0
	for _, c := range ownChunks {
		s := float64(c.RerankScore)
		if s > topScore {
			topScore = s
		}
		if s != 0.5 {
			allExactlyHalf = false
		}
		if s >= thr.Strong {
			strongChunks = append(strongChunks, c)
		}
		if s >= thr.Usable {
			usableChunks = append(usableChunks, c)
		}
	}
	if len(usableChunks) > gate1MaxChunks {
		usableChunks = usableChunks[:gate1MaxChunks]
	}
	result.TrainingChunks = usableChunks
	result.Gate1Passed = len(usableChunks) > 0

	// rerank_fallback_suspected: every score being exactly 0.5 is the
	// signature of getCourseChunks' fallback when ml-sidecar's reranker is
	// unreachable — real reranker output is never uniformly 0.5. Without
	// this flag, "sidecar down" and "scores genuinely low" look identical
	// in the logs and need opposite fixes.
	g.logger.Info("gate system: Gate 1 result",
		zap.String("expert", expert.Name),
		zap.String("domain", expert.Domain),
		zap.Int("total_chunks", len(ownChunks)),
		zap.Int("usable_chunks", len(usableChunks)),
		zap.Int("strong_chunks", len(strongChunks)),
		zap.Float64("top_score", topScore),
		zap.Float64("usable_threshold", thr.Usable),
		zap.Float64("strong_threshold", thr.Strong),
		zap.String("threshold_source", thr.Source),
		zap.Bool("rerank_fallback_suspected", allExactlyHalf),
	)

	// ============================================================
	// GATE 2: Peer knowledge poll — ALWAYS runs
	// ============================================================
	// Peers are polled on every task now, not only when Gate 1 fails.
	// Cross-expert collaboration is the whole point of the workflow: the
	// System Design expert's architecture has to reach the PMO's task and
	// vice versa. Previously a Gate 1 pass returned early and the peer's
	// material never entered the prompt at all.
	//
	// Cost note: pollPeers only does embed + vector search + rerank per
	// peer (no LLM call), is parallel, and is bounded by gate2PollTimeout.
	peerContribs := g.pollPeers(ctx, expert.ID, taskDescription, allExperts)
	result.PeerContributions = peerContribs

	// ============================================================
	// GATE 3: Generic knowledge — the CLIENT decides, not a model
	// ============================================================
	// genericAllowancePct comes from workflows.generic_allowance_pct, which
	// the client sets from the approval gate after reading the deliverables.
	//   0 (default) -> generic BLOCKED. Trained + peer knowledge only.
	//   >0          -> at most that share of the answer may be generic, and
	//                  every such claim must be tagged [GENERIC].
	//
	// WHY this replaced the LLM coverage check: that check asked a cheap
	// model "does this knowledge cover the task?". For any broad task the
	// answer was "NO", which set CoverageGap to the ENTIRE task and told the
	// expert generic was fine for everything — so the trained material was
	// retrieved, put in the prompt, and then effectively thrown away. The
	// experts are trained on specific industry course material; whether to
	// dilute that with generic knowledge is a product decision for the
	// client, not an inference a model should make per task. Removing it
	// also drops one LLM call per expert per phase.
	result.GenericAllowancePct = genericAllowancePct
	result.GenericAllowed = genericAllowancePct > 0
	if result.GenericAllowed {
		// The gap is not model-decided any more. The expert is told how much
		// generic is permitted and must still cite everything else.
		result.CoverageGap = taskDescription
	}

	g.logger.Info("gate system: generic policy",
		zap.String("workflow_id", workflowID.String()),
		zap.String("expert", expert.Name),
		zap.Float64("generic_allowance_pct", genericAllowancePct),
		zap.Bool("generic_allowed", result.GenericAllowed),
		zap.Int("training_chunks", len(result.TrainingChunks)),
		zap.Int("strong_chunks", len(strongChunks)),
		zap.Int("peer_contributions", len(peerContribs)),
	)

	return result, nil
}

// pollPeers searches all other experts' chunks in parallel.
// Returns within gate2PollTimeout regardless of how many experts respond.
//
// Mental execution:
//   allExperts = [PM, System Design, LLD, DSA]
//   asking expert = System Design
//   poll: PM, LLD, DSA in parallel goroutines
//   timeout: 10s
//   result: whoever responded within 10s
func (g *GateSystem) pollPeers(
	ctx context.Context,
	askingExpertID uuid.UUID,
	taskDescription string,
	allExperts []workflowExpert,
) []PeerContribution {
	// Create timeout context for the poll
	pollCtx, cancel := context.WithTimeout(ctx, gate2PollTimeout)
	defer cancel()

	type peerResult struct {
		contrib PeerContribution
		err     error
	}

	// Count peers (exclude asking expert)
	var peers []workflowExpert
	for _, e := range allExperts {
		if e.ID != askingExpertID {
			peers = append(peers, e)
		}
	}

	if len(peers) == 0 {
		return nil
	}

	resultCh := make(chan peerResult, len(peers))
	var wg sync.WaitGroup

	for _, peer := range peers {
		wg.Add(1)
		go func(p workflowExpert) {
			defer wg.Done()
			chunks, err := g.assembler.GetCourseChunksForWorkflow(
				pollCtx, p.ID, taskDescription, 3,
			)
			resultCh <- peerResult{
				contrib: PeerContribution{
					ExpertID:   p.ID,
					ExpertName: p.Name,
					Chunks:     chunks,
				},
				err: err,
			}
		}(peer)
	}

	// Close channel when all goroutines finish
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Collect results within timeout
	var contributions []PeerContribution
	for {
		select {
		case res, ok := <-resultCh:
			if !ok {
				// All peers responded
				return contributions
			}
			if res.err == nil && len(res.contrib.Chunks) > 0 {
				contributions = append(contributions, res.contrib)
			}
		case <-pollCtx.Done():
			// Timeout: return whatever we have
			g.logger.Warn("gate system: Gate 2 poll timeout, using partial results",
				zap.Int("received", len(contributions)),
				zap.Int("total_peers", len(peers)),
			)
			return contributions
		}
	}
}

// HasKnowledge returns true when the expert has at least one training chunk
// for the given topic that scores at or above the usable threshold for its
// domain (B4: domain override if applied, else the package default 0.40).
//
// Used by the conflict auto-resolver (tool_loop.go toolRaiseConflict) to
// check whether a higher-rank expert actually knows the topic before letting
// it decide. Rank gives authority; knowledge gives legitimacy.
//
// WHY reuse the Gate 1 usable bar: it is the same bar Gate 1 uses to decide
// "this training is relevant enough to apply as principles". A score below
// it means the expert's training is too weak to reason from — the same
// conclusion applies here. domain is optional; empty → package default.
func (g *GateSystem) HasKnowledge(ctx context.Context, expertID uuid.UUID, topic string, domain string) bool {
	chunks, err := g.assembler.GetCourseChunksForWorkflow(ctx, expertID, topic, 1)
	if err != nil || len(chunks) == 0 {
		return false
	}
	thr := g.thresholdsFor(ctx, domain)
	return float64(chunks[0].RerankScore) >= thr.Usable
}

// FormatGateContext builds the context string from GateResult.
// Called by AgentLoop.buildContext() to replace the old training-only context.
//
// Context structure (in order):
//   [GATE 1: YOUR TRAINING]  — if Gate 1 passed
//   [GATE 2: PEER KNOWLEDGE] — if Gate 2 has contributions
//   [GATE 3: GAP FILLING]    — if generic allowed, with strict instructions
func FormatGateContext(result *GateResult) string {
	var sb strings.Builder

	hasTraining := len(result.TrainingChunks) > 0
	hasPeers := len(result.PeerContributions) > 0

	// Own training first.
	//
	// WHY no early return here any more: this block used to end with
	// `return sb.String()`, so the moment an expert's own training
	// qualified, the peer section below was never rendered — the other
	// experts' work silently never reached the prompt. Both sources are
	// required input now.
	if hasTraining {
		sb.WriteString(fmt.Sprintf("[GATE 1: YOUR TRAINING — %d chunks]\n", len(result.TrainingChunks)))
		sb.WriteString("These are YOUR principles. Apply them first. Cite each with [CHUNK_uuid].\n\n")
		for _, c := range result.TrainingChunks {
			if c.Topic != "" {
				sb.WriteString(fmt.Sprintf("[CHUNK_%s | Topic: %s | Score: %.2f]\n",
					c.ID.String()[:8], c.Topic, c.RerankScore))
			}
			sb.WriteString(c.Text)
			sb.WriteString("\n\n")
		}
	}

	if hasPeers {
		sb.WriteString("[GATE 2: PEER KNOWLEDGE]\n")
		if hasTraining {
			sb.WriteString("What your peers were trained on for this same task.\n")
		} else {
			sb.WriteString("Your training didn't cover this. Your peers contributed the following.\n")
		}
		sb.WriteString("Cite peer knowledge as [PEER:ExpertName].\n\n")
		for _, contrib := range result.PeerContributions {
			if len(contrib.Chunks) == 0 {
				continue
			}
			sb.WriteString(fmt.Sprintf("[From %s]:\n", contrib.ExpertName))
			for _, c := range contrib.Chunks {
				sb.WriteString(c.Text)
				sb.WriteString("\n")
			}
			sb.WriteString("\n")
		}
	}

	// Both sources present: say explicitly that neither may be dropped.
	// Without this the model tends to answer from whichever block it read
	// last and quietly ignore the other.
	if hasTraining && hasPeers {
		sb.WriteString("REQUIRED: your own training AND your peers' knowledge above are both\n")
		sb.WriteString("input to this task. Reconcile them. Do not ignore either one.\n\n")
	}

	switch {
	case result.GenericAllowed:
		// The client has opened a bounded generic allowance.
		sb.WriteString("[GENERIC ALLOWANCE OPENED BY THE CLIENT]\n")
		sb.WriteString(fmt.Sprintf(
			"At most %.0f%% of your output may come from generic knowledge.\n",
			result.GenericAllowancePct))
		sb.WriteString("Rules:\n")
		sb.WriteString("- The trained and peer material above stays the primary source.\n")
		sb.WriteString("- Tag EVERY generic sentence with [GENERIC]. Untagged generic content is a violation.\n")
		sb.WriteString("- Everything not tagged [GENERIC] MUST carry [CHUNK_id] or [PEER:ExpertName].\n")
		sb.WriteString("- Do NOT restate what your training or your peers already cover.\n")
		sb.WriteString("- Generic claims are reviewed by the admin as future training material.\n\n")

	case hasTraining || hasPeers:
		// Default posture: trained + peer knowledge only.
		sb.WriteString("[GENERIC KNOWLEDGE IS BLOCKED]\n")
		sb.WriteString("Build the answer ONLY from the material above.\n")
		sb.WriteString("- EVERY claim must carry its source: [CHUNK_id] for your training,\n")
		sb.WriteString("  [PEER:ExpertName] for a peer's. A claim with no source is not allowed.\n")
		sb.WriteString("- If something the task asks for is genuinely not in the material above,\n")
		sb.WriteString("  do NOT invent it. Write [NOT_COVERED: <what is missing>] instead.\n")
		sb.WriteString("  Saying it is missing is correct behaviour; guessing is not.\n\n")

	default:
		// Nothing at all was found. Say so rather than silently guessing.
		sb.WriteString("[NO TRAINING OR PEER KNOWLEDGE FOUND]\n")
		sb.WriteString("You have no trained material for this task and no peer supplied any.\n")
		sb.WriteString("State this plainly as [NOT_COVERED: <what you would need>] rather than\n")
		sb.WriteString("answering from generic knowledge, unless the client has opened a\n")
		sb.WriteString("generic allowance.\n\n")
	}

	return sb.String()
}
