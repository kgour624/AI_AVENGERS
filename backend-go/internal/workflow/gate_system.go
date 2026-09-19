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

// gate1SimilarityThreshold: minimum rerank score for Gate 1 to pass.
// WHY 0.70 (not 0.75): 0.75 is too strict for APPLY_PRINCIPLES mode.
// Principles transfer at lower similarity ("consistent hashing" applies
// to "URL shortener" at ~0.68-0.72 similarity).
// 0.70 is the middle of the suggested 0.68-0.72 range.
const gate1SimilarityThreshold = 0.70

// gate2PollTimeout: how long Gate 2 waits for peer responses.
// WHY 10s: enough for parallel DB queries, not so long it blocks workflow.
// Peers that don't respond within 10s are skipped — partial knowledge
// is better than blocking forever.
const gate2PollTimeout = 10 * time.Second

// GateResult is the output of RunGates.
// AgentLoop uses this to build the LLM context.
type GateResult struct {
	// TrainingChunks: expert's own training chunks (Gate 1 result).
	// Non-empty when Gate 1 passes (similarity >= 0.70).
	TrainingChunks []chinawall.CourseChunk

	// PeerContributions: knowledge from other experts (Gate 2 result).
	// Non-empty when Gate 1 fails but peers have relevant knowledge.
	PeerContributions []PeerContribution

	// GenericAllowed: true only when Gate 1 AND Gate 2 both fail to
	// cover the task. Gate 3 is only enabled when this is true.
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
// Gate 1 — Own training (Vector DB):
//   Expert's course_chunks searched for task.
//   If similarity >= 0.70 -> PASS -> Generic BLOCKED.
//   Expert uses only their training. Citations: [CHUNK_uuid].
//
// Gate 2 — Peer knowledge (Blackboard Poll):
//   All other workflow experts' chunks searched in parallel (10s timeout).
//   Coverage check: do combined chunks cover the task?
//   YES -> Generic BLOCKED. PARTIAL/NO -> Gate 3 enabled.
//
// Gate 3 — Gap filling (Generic, restricted):
//   Only enabled when Gate 1 AND Gate 2 fail.
//   Prompt: "Fill ONLY this gap: [gap]. Minimal generic knowledge."
//   Generic claims saved to pending_experience for admin review.
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

	// Filter by similarity threshold
	var qualifiedChunks []chinawall.CourseChunk
	for _, c := range ownChunks {
		if float64(c.RerankScore) >= gate1SimilarityThreshold {
			qualifiedChunks = append(qualifiedChunks, c)
		}
	}

	if len(qualifiedChunks) > 0 {
		// Gate 1 PASS: expert has sufficient training knowledge
		result.TrainingChunks = qualifiedChunks
		result.Gate1Passed = true
		result.GenericAllowed = false
		g.logger.Info("gate system: Gate 1 PASS — generic blocked",
			zap.String("expert", expert.Name),
			zap.Int("qualified_chunks", len(qualifiedChunks)),
		)
		return result, nil
	}

	// Diagnostics for "chunks were found but none qualified".
	//
	// WHY: total_chunks alone cannot tell these two cases apart, and they
	// need opposite fixes:
	//   a) the reranker (ml-sidecar) is unreachable — getCourseChunks falls
	//      back to a hardcoded RerankScore of 0.5 for every chunk, which is
	//      below this threshold, so Gate 1 can NEVER pass no matter how good
	//      the training data is. Signature: every score is exactly 0.5.
	//   b) the reranker ran and the best chunk genuinely scored below the
	//      threshold. Signature: varied scores; top_score shows how close.
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
	}
	g.logger.Info("gate system: Gate 1 FAIL — proceeding to Gate 2",
		zap.String("expert", expert.Name),
		zap.Int("total_chunks", len(ownChunks)),
		zap.Float64("threshold", gate1SimilarityThreshold),
		zap.Float64("top_score", topScore),
		zap.Bool("rerank_fallback_suspected", allExactlyHalf),
	)

	// ============================================================
	// GATE 2: Peer knowledge poll (parallel, 10s timeout)
	// ============================================================
	peerContribs := g.pollPeers(ctx, expert.ID, taskDescription, allExperts)
	result.PeerContributions = peerContribs

	// Coverage check: do combined peer chunks cover the task?
	coverage, gap := g.checkCoverage(ctx, workflowID, taskDescription, peerContribs)
	result.Gate2Coverage = coverage

	switch coverage {
	case "YES":
		// Gate 2 PASS: peers cover the task fully
		result.GenericAllowed = false
		g.logger.Info("gate system: Gate 2 PASS — peer knowledge sufficient",
			zap.String("expert", expert.Name),
			zap.Int("peer_contributions", len(peerContribs)),
		)
	case "PARTIAL":
		// Gate 2 PARTIAL: peers cover some, generic allowed for gap only
		result.GenericAllowed = true
		result.CoverageGap = gap
		g.logger.Info("gate system: Gate 2 PARTIAL — generic allowed for gap",
			zap.String("expert", expert.Name),
			zap.String("gap", gap),
		)
	default: // "NO"
		// Gate 2 FAIL: no peer knowledge, generic allowed for full task
		result.GenericAllowed = true
		result.CoverageGap = taskDescription
		g.logger.Info("gate system: Gate 2 FAIL — generic allowed for full task",
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
	peerContribs []PeerContribution,
) (coverage string, gap string) {
	if len(peerContribs) == 0 {
		return "NO", taskDescription
	}

	// Build peer knowledge summary
	var sb strings.Builder
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
		UserPrompt: fmt.Sprintf("TASK: %s\n\nPEER KNOWLEDGE:\n%s", taskDescription, peerText),
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

	if result.Gate1Passed && len(result.TrainingChunks) > 0 {
		sb.WriteString(fmt.Sprintf("[GATE 1 PASS: YOUR TRAINING — %d chunks]\n", len(result.TrainingChunks)))
		sb.WriteString("These are YOUR principles. Use them. Cite each with [CHUNK_uuid].\n\n")
		for _, c := range result.TrainingChunks {
			if c.Topic != "" {
				sb.WriteString(fmt.Sprintf("[CHUNK_%s | Topic: %s | Score: %.2f]\n",
					c.ID.String()[:8], c.Topic, c.RerankScore))
			}
			sb.WriteString(c.Text)
			sb.WriteString("\n\n")
		}
		sb.WriteString("RULE: Generic knowledge is BLOCKED. Use only the above training material.\n\n")
		return sb.String()
	}

	if len(result.PeerContributions) > 0 {
		sb.WriteString("[GATE 2: PEER KNOWLEDGE]\n")
		sb.WriteString("Your training didn't cover this. Your peers contributed the following.\n")
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

	if result.GenericAllowed {
		sb.WriteString("[GATE 3: GAP FILLING ALLOWED]\n")
		sb.WriteString(fmt.Sprintf("ONLY fill this specific gap: %s\n", result.CoverageGap))
		sb.WriteString("Rules:\n")
		sb.WriteString("- Use minimal generic knowledge\n")
		sb.WriteString("- Mark every generic claim with [GENERIC]\n")
		sb.WriteString("- Do NOT rewrite what peers already covered\n")
		sb.WriteString("- Generic knowledge will be reviewed by admin for future training\n\n")
	} else if len(result.PeerContributions) > 0 {
		sb.WriteString("RULE: Generic knowledge is BLOCKED. Use only peer knowledge above.\n\n")
	} else {
		sb.WriteString("[NO TRAINING OR PEER KNOWLEDGE FOUND]\n")
		sb.WriteString("Use your general domain expertise. Mark all claims with [GENERIC].\n\n")
	}

	return sb.String()
}
