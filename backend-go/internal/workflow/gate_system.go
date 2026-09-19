package workflow

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	appcontext "ai_avengers/backend/internal/context"
	"ai_avengers/backend/internal/chinawall"
	"ai_avengers/backend/internal/gateway"
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

	// CoverageGap: what's missing after Gate 1 + Gate 2.
	// Used as Gate 3 prompt: "Fill ONLY this gap: [CoverageGap]"
	// Empty when GenericAllowed=false.
	CoverageGap string

	// Gate1Passed: true when expert's own training covered the task.
	Gate1Passed bool

	// Gate2Coverage: "YES" | "PARTIAL" | "NO" from coverage check.
	Gate2Coverage string
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
type GateSystem struct {
	assembler *appcontext.Assembler
	gateway   *gateway.ModelGateway
	logger    *zap.Logger
}

// NewGateSystem creates a new GateSystem.
func NewGateSystem(assembler *appcontext.Assembler, gw *gateway.ModelGateway, logger *zap.Logger) *GateSystem {
	return &GateSystem{assembler: assembler, gateway: gw, logger: logger}
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
		if s >= gate1StrongThreshold {
			strongChunks = append(strongChunks, c)
		}
		if s >= gate1UsableThreshold {
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
		zap.Int("total_chunks", len(ownChunks)),
		zap.Int("usable_chunks", len(usableChunks)),
		zap.Int("strong_chunks", len(strongChunks)),
		zap.Float64("top_score", topScore),
		zap.Float64("usable_threshold", gate1UsableThreshold),
		zap.Float64("strong_threshold", gate1StrongThreshold),
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

	// Own training is strong: block generic outright and skip the coverage
	// LLM call — there is nothing to decide.
	if len(strongChunks) > 0 {
		result.GenericAllowed = false
		result.Gate2Coverage = "YES"
		g.logger.Info("gate system: strong own training — generic blocked",
			zap.String("expert", expert.Name),
			zap.Int("strong_chunks", len(strongChunks)),
			zap.Int("peer_contributions", len(peerContribs)),
		)
		return result, nil
	}

	// Coverage is judged over BOTH own usable training and peer knowledge,
	// so generic is permitted only for what neither of them covers.
	coverage, gap := g.checkCoverage(ctx, workflowID, taskDescription, result.TrainingChunks, peerContribs)
	result.Gate2Coverage = coverage

	switch coverage {
	case "YES":
		// Training + peers cover the task fully.
		result.GenericAllowed = false
		g.logger.Info("gate system: training + peers cover task — generic blocked",
			zap.String("expert", expert.Name),
			zap.Int("training_chunks", len(result.TrainingChunks)),
			zap.Int("peer_contributions", len(peerContribs)),
		)
	case "PARTIAL":
		// Generic allowed for the named gap only, tagged [GENERIC].
		result.GenericAllowed = true
		result.CoverageGap = gap
		g.logger.Info("gate system: partial coverage — generic allowed for gap only",
			zap.String("expert", expert.Name),
			zap.Int("training_chunks", len(result.TrainingChunks)),
			zap.String("gap", gap),
		)
	default: // "NO"
		// Neither training nor peers cover it: generic for the full task.
		result.GenericAllowed = true
		result.CoverageGap = taskDescription
		g.logger.Info("gate system: no coverage — generic allowed for full task",
			zap.String("expert", expert.Name),
		)
	}

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

// checkCoverage asks LLM: do these peer chunks cover the task?
// Returns: coverage ("YES"|"PARTIAL"|"NO") + gap description.
//
// WHY cheap LLM (ModelCheap):
//   This is a binary classification, not generation.
//   ModelCheap is sufficient and costs 10x less.
//
// Mental execution:
//   task: "Design rate limiter"
//   peerChunks: [api_design, rate_limiter_pattern]
//   LLM: "PARTIAL\nMISSING: distributed rate limiting across nodes"
//   return: "PARTIAL", "distributed rate limiting across nodes"
func (g *GateSystem) checkCoverage(
	ctx context.Context,
	workflowID uuid.UUID,
	taskDescription string,
	ownChunks []chinawall.CourseChunk,
	peerContribs []PeerContribution,
) (coverage string, gap string) {
	if len(ownChunks) == 0 && len(peerContribs) == 0 {
		return "NO", taskDescription
	}

	// Build the available-knowledge summary from BOTH sources.
	//
	// WHY own training is included: this used to judge peer chunks only, so
	// an expert whose own training partly covered the task was still told
	// "NO coverage — generic allowed for the full task". The gap has to be
	// measured against everything the expert actually has in front of it.
	var sb strings.Builder
	for _, c := range ownChunks {
		if sb.Len() == 0 {
			sb.WriteString("[Your own training]\n")
		}
		sb.WriteString(c.Text)
		sb.WriteString("\n")
	}
	if sb.Len() > 0 {
		sb.WriteString("\n")
	}
	for _, contrib := range peerContribs {
		if len(contrib.Chunks) == 0 {
			continue
		}
		sb.WriteString(fmt.Sprintf("[Expert: %s]\n", contrib.ExpertName))
		for _, c := range contrib.Chunks {
			sb.WriteString(c.Text)
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	peerText := sb.String()
	if strings.TrimSpace(peerText) == "" {
		return "NO", taskDescription
	}

	resp, err := g.gateway.Call(ctx, gateway.LLMRequest{
		Model:      gateway.ModelCheap,
		WorkflowID: &workflowID,
		SystemPrompt: `You are a coverage checker. Given a task and peer knowledge, determine if the knowledge covers the task.

Output EXACTLY one of:
YES
PARTIAL\nMISSING: <what is missing, max 50 words>
NO

No other output.`,
		UserPrompt: fmt.Sprintf("TASK: %s\n\nAVAILABLE KNOWLEDGE:\n%s", taskDescription, peerText),
		MaxTokens:  100,
		Temperature: 0.0,
	})
	if err != nil {
		g.logger.Warn("gate system: coverage check LLM failed", zap.Error(err))
		// Fail safe: allow generic
		return "NO", taskDescription
	}

	raw := strings.TrimSpace(resp.Content)
	if strings.HasPrefix(raw, "YES") {
		return "YES", ""
	}
	if strings.HasPrefix(raw, "PARTIAL") {
		lines := strings.SplitN(raw, "\n", 2)
		missing := ""
		if len(lines) > 1 {
			missing = strings.TrimPrefix(strings.TrimSpace(lines[1]), "MISSING:")
			missing = strings.TrimSpace(missing)
		}
		if missing == "" {
			missing = taskDescription
		}
		return "PARTIAL", missing
	}
	return "NO", taskDescription
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
		sb.WriteString("[GATE 3: GAP FILLING ALLOWED]\n")
		sb.WriteString(fmt.Sprintf("ONLY fill this specific gap: %s\n", result.CoverageGap))
		sb.WriteString("Rules:\n")
		sb.WriteString("- Use minimal generic knowledge\n")
		sb.WriteString("- Mark every generic claim with [GENERIC]\n")
		sb.WriteString("- Do NOT rewrite what your training or your peers already covered\n")
		sb.WriteString("- Generic knowledge will be reviewed by admin for future training\n\n")

	case hasTraining || hasPeers:
		// Something was found and coverage was judged sufficient.
		sb.WriteString("RULE: Generic knowledge is BLOCKED. Build your answer only from\n")
		sb.WriteString("the material above.\n\n")

	default:
		// Nothing at all was found — the honest fallback.
		sb.WriteString("[NO TRAINING OR PEER KNOWLEDGE FOUND]\n")
		sb.WriteString("Use your general domain expertise. Mark all claims with [GENERIC].\n\n")
	}

	return sb.String()
}
