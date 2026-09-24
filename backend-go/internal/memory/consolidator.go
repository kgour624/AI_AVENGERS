package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/gateway"
	"ai_avengers/backend/internal/usage"
)

// Consolidation knobs (B7). Tuned for single-admin production:
// enough entries to be worth an LLM call, cheap model, bounded payload.
const (
	// consolidateMinEntries: below this, project stays as raw L2 rows.
	consolidateMinEntries = 12
	// consolidateMaxEntries: hard cap on source rows fed to the LLM
	// (P5 budget; oldest beyond this stay as raw — external, re-retrievable).
	consolidateMaxEntries = 40
	// consolidateInterval: how often the background job wakes.
	consolidateInterval = 6 * time.Hour
	// consolidateCooldown: skip a project that already has a fresh summary.
	consolidateCooldown = 24 * time.Hour
	// decayHalfLife: preference weight halves every N idle days
	// (exponential on days-since-last-access).
	decayHalfLifeDays = 30.0
	// decayFloor: weight below this → supersede (forget), never DELETE.
	decayFloor = 0.05
	// criticalImportance: importance >= this never decays (P4 critical flag).
	criticalImportance = 5
)

// Consolidator periodically:
//  1. Decays preference weights (facts untouched).
//  2. Hierarchical-summarizes dense L2 projects into one 'summary' row
//     and supersedes the covered raw entries (kept, not deleted).
//
// Wired from cmd/server/main.go as a long-lived goroutine. Failures are
// logged and skipped — never blocks request paths (async, P4).
type Consolidator struct {
	manager *Manager
	gateway *gateway.ModelGateway
	logger  *zap.Logger
	// interval overrides consolidateInterval in tests; 0 → default.
	interval time.Duration
}

// NewConsolidator builds a Consolidator. gateway may be nil — then
// decay still runs, consolidate is a no-op (no LLM).
func NewConsolidator(m *Manager, gw *gateway.ModelGateway, logger *zap.Logger) *Consolidator {
	return &Consolidator{
		manager:  m,
		gateway:  gw,
		logger:   logger,
		interval: consolidateInterval,
	}
}

// Run blocks until ctx is cancelled. Tick + one immediate pass on start
// so a restart does not wait a full interval before first cleanup.
func (c *Consolidator) Run(ctx context.Context) {
	if c == nil || c.manager == nil {
		return
	}
	interval := c.interval
	if interval <= 0 {
		interval = consolidateInterval
	}
	c.logger.Info("memory consolidator started", zap.Duration("interval", interval))
	c.runOnce(ctx)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			c.logger.Info("memory consolidator stopped")
			return
		case <-ticker.C:
			c.runOnce(ctx)
		}
	}
}

func (c *Consolidator) runOnce(ctx context.Context) {
	if err := c.DecayPreferences(ctx); err != nil {
		c.logger.Warn("preference decay failed", zap.Error(err))
	}
	if c.gateway == nil {
		return
	}
	projects, err := c.listCandidates(ctx)
	if err != nil {
		c.logger.Warn("list consolidate candidates failed", zap.Error(err))
		return
	}
	for _, p := range projects {
		if ctx.Err() != nil {
			return
		}
		res, err := c.ConsolidateProject(ctx, p)
		if err != nil {
			c.logger.Warn("consolidate project failed",
				zap.String("project_id", p.String()),
				zap.Error(err),
			)
			continue
		}
		if res != nil && res.WroteSummary {
			c.logger.Info("project memory consolidated",
				zap.String("project_id", p.String()),
				zap.Int("source_entries", res.SourceCount),
				zap.Int("superseded", res.Superseded),
			)
		}
	}
}

// ConsolidateResult is observability for one project pass.
type ConsolidateResult struct {
	SourceCount  int
	Superseded   int
	WroteSummary bool
	Skipped      string // reason when no write
}

// consolidationVerdict is the strict JSON the cheap model must emit.
type consolidationVerdict struct {
	Decisions   []string `json:"decisions"`
	Numbers     []string `json:"numbers"`
	Constraints []string `json:"constraints"`
	OpenItems   []string `json:"open_items"`
	SNR         float64  `json:"snr"` // 0-1 self-reported signal quality
}

// ConsolidateProject summarizes active non-summary L2 rows for one
// project into a single 'summary' row (growing-doc: prior summary is
// superseded). Covered raw rows are superseded (externalized, still in
// DB for re-retrieve / audit) — never DELETE (P4: external > blind
// eviction; L3 remains append-only truth).
func (c *Consolidator) ConsolidateProject(ctx context.Context, projectID uuid.UUID) (*ConsolidateResult, error) {
	if c == nil || c.manager == nil || c.manager.l2 == nil {
		return nil, fmt.Errorf("consolidator not configured")
	}
	// C5: attribute the (cheap) consolidation LLM call to this project.
	ctx = usage.WithAttribution(ctx, usage.Attribution{
		ProjectID: &projectID, UseCase: usage.UseCaseConsolidate,
	})
	res := &ConsolidateResult{}

	// Fresh summary within cooldown → skip (end-of-phase, not per-event).
	if recent, err := c.manager.l2.hasRecentSummary(ctx, projectID, consolidateCooldown); err != nil {
		return nil, err
	} else if recent {
		res.Skipped = "recent_summary"
		return res, nil
	}

	entries, err := c.manager.l2.listActiveForConsolidate(ctx, projectID, consolidateMaxEntries)
	if err != nil {
		return nil, err
	}
	res.SourceCount = len(entries)
	if len(entries) < consolidateMinEntries {
		res.Skipped = "below_min_entries"
		return res, nil
	}

	verdict, err := c.summarize(ctx, entries)
	if err != nil {
		return nil, err
	}
	body := formatConsolidationSummary(verdict)
	if strings.TrimSpace(body) == "" {
		res.Skipped = "empty_summary"
		return res, nil
	}

	// Attribute the summary to the expert with the most source rows
	// (expert_id NOT NULL on L2). Never invent a UUID.
	owner := majorityExpert(entries)
	if owner == uuid.Nil {
		res.Skipped = "no_expert"
		return res, nil
	}

	// Importance high so summary stays on top of GetRecent / search;
	// criticalImportance so decay (if ever mistyped as preference) skips it.
	summaryEntry := L2Entry{
		ProjectID:     projectID,
		ExpertID:      owner,
		MemoryType:    "summary",
		Content:       body,
		Context:       fmt.Sprintf("consolidated from %d L2 entries; snr=%.2f", len(entries), verdict.SNR),
		TurnReference: 0,
		Importance:    criticalImportance,
		Weight:        1.0,
	}
	if err := c.manager.l2.Append(ctx, summaryEntry); err != nil {
		return nil, fmt.Errorf("write summary: %w", err)
	}
	res.WroteSummary = true

	// Supersede prior summaries first (growing-doc: one live summary).
	n, _ := c.manager.l2.supersedeOlderSummaries(ctx, projectID)
	res.Superseded += n

	// Supersede the source rows that fed this summary. Keep them in DB
	// (is_superseded=TRUE) so search skips them but audit/re-retrieve works.
	ids := make([]uuid.UUID, 0, len(entries))
	for _, e := range entries {
		if e.MemoryType == "summary" {
			continue
		}
		ids = append(ids, e.ID)
	}
	n, err = c.manager.l2.supersedeMany(ctx, ids)
	if err != nil {
		c.logger.Warn("supersede source entries partial failure", zap.Error(err))
	}
	res.Superseded += n

	// L3 audit trail (append-only). client_id taken from projects row.
	c.recordConsolidationEvent(ctx, projectID, res)
	return res, nil
}

// DecayPreferences applies exponential decay to preference rows only.
// Facts (decision/code/…/summary) are untouched. importance >= 5 is
// never decayed. Weight below floor → supersede (forget), never DELETE.
func (c *Consolidator) DecayPreferences(ctx context.Context) error {
	if c == nil || c.manager == nil || c.manager.l2 == nil {
		return nil
	}
	return c.manager.l2.decayPreferences(ctx, decayHalfLifeDays, decayFloor, criticalImportance)
}

func (c *Consolidator) listCandidates(ctx context.Context) ([]uuid.UUID, error) {
	return c.manager.l2.listConsolidateCandidates(ctx, consolidateMinEntries, consolidateCooldown)
}

func (c *Consolidator) summarize(ctx context.Context, entries []L2Entry) (*consolidationVerdict, error) {
	if c.gateway == nil {
		return nil, fmt.Errorf("gateway nil")
	}
	var sb strings.Builder
	for i, e := range entries {
		sb.WriteString(fmt.Sprintf(
			"[%d type=%s importance=%d weight=%.2f]\nCONTENT: %s\nCONTEXT: %s\n\n",
			i+1, e.MemoryType, e.Importance, e.Weight,
			truncateRunes(e.Content, 400),
			truncateRunes(e.Context, 200),
		))
	}

	systemPrompt := `You consolidate project memory into a high-SNR summary for future agents.
Prescriptive: extract ONLY what changes future behaviour.

Output ONE JSON object, no markdown, no preamble:
{
  "decisions": ["..."],      // settled decisions that must not be silently contradicted
  "numbers": ["..."],        // concrete figures, IDs, versions, rates
  "constraints": ["..."],    // hard limits, must-not rules, tech choices locked
  "open_items": ["..."],     // unresolved questions still blocking work
  "snr": 0.0                 // 0-1 self-rated signal quality of the source set
}

Rules:
1. Prefer fewer, sharper items over long prose. Empty arrays are fine.
2. Do NOT invent facts not present in the source entries.
3. Drop chit-chat, clarifications, and superseded vibes.
4. snr low when source is noisy/contradictory.`

	resp, err := c.gateway.Call(ctx, gateway.LLMRequest{
		Model:        gateway.ModelCheap, // async, batched, cheap (P4)
		SystemPrompt: systemPrompt,
		UserPrompt:   "SOURCE L2 ENTRIES:\n\n" + sb.String(),
		MaxTokens:    800,
		Temperature:  0.1,
	})
	if err != nil {
		return nil, err
	}
	return parseConsolidationVerdict(resp.Content)
}

func parseConsolidationVerdict(raw string) (*consolidationVerdict, error) {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "```") {
		raw = strings.TrimPrefix(raw, "```json")
		raw = strings.TrimPrefix(raw, "```JSON")
		raw = strings.TrimPrefix(raw, "```")
		if i := strings.LastIndex(raw, "```"); i >= 0 {
			raw = raw[:i]
		}
		raw = strings.TrimSpace(raw)
	}
	if start := strings.Index(raw, "{"); start >= 0 {
		if end := strings.LastIndex(raw, "}"); end > start {
			raw = raw[start : end+1]
		}
	}
	var v consolidationVerdict
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return nil, fmt.Errorf("consolidation verdict unmarshal: %w", err)
	}
	if v.SNR < 0 {
		v.SNR = 0
	}
	if v.SNR > 1 {
		v.SNR = 1
	}
	return &v, nil
}

func formatConsolidationSummary(v *consolidationVerdict) string {
	if v == nil {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("PROJECT SUMMARY (consolidated)\n")
	writeSection(&sb, "Decisions", v.Decisions)
	writeSection(&sb, "Numbers", v.Numbers)
	writeSection(&sb, "Constraints", v.Constraints)
	writeSection(&sb, "Open items", v.OpenItems)
	sb.WriteString(fmt.Sprintf("\nSNR: %.2f\n", v.SNR))
	return strings.TrimSpace(sb.String())
}

func writeSection(sb *strings.Builder, title string, items []string) {
	cleaned := make([]string, 0, len(items))
	for _, it := range items {
		it = strings.TrimSpace(it)
		if it != "" {
			cleaned = append(cleaned, it)
		}
	}
	if len(cleaned) == 0 {
		return
	}
	sb.WriteString("\n")
	sb.WriteString(title)
	sb.WriteString(":\n")
	for _, it := range cleaned {
		sb.WriteString("- ")
		sb.WriteString(it)
		sb.WriteString("\n")
	}
}

func majorityExpert(entries []L2Entry) uuid.UUID {
	counts := map[uuid.UUID]int{}
	var best uuid.UUID
	bestN := 0
	for _, e := range entries {
		if e.ExpertID == uuid.Nil {
			continue
		}
		counts[e.ExpertID]++
		if counts[e.ExpertID] > bestN {
			bestN = counts[e.ExpertID]
			best = e.ExpertID
		}
	}
	return best
}

func (c *Consolidator) recordConsolidationEvent(ctx context.Context, projectID uuid.UUID, res *ConsolidateResult) {
	if c.manager == nil || c.manager.l3 == nil || res == nil {
		return
	}
	clientID, err := c.manager.l2.projectClientID(ctx, projectID)
	if err != nil || clientID == uuid.Nil {
		c.logger.Warn("consolidation L3 skip: no client_id",
			zap.String("project_id", projectID.String()),
			zap.Error(err),
		)
		return
	}
	_ = c.manager.l3.Append(ctx, L3Event{
		ProjectID: projectID,
		ClientID:  clientID,
		EventType: EventMemoryConsolidated,
		EventData: map[string]interface{}{
			"source_count":  res.SourceCount,
			"superseded":    res.Superseded,
			"wrote_summary": res.WroteSummary,
		},
		Reasoning:    fmt.Sprintf("consolidated %d L2 entries, superseded %d", res.SourceCount, res.Superseded),
		DecisionMade: "memory_consolidated",
	})
}

// decayWeight is pure exponential half-life math — unit-tested.
// weight' = weight * 0.5^(idleDays / halfLifeDays). Critical/non-pref skip upstream.
func decayWeight(current float64, idleDays, halfLifeDays float64) float64 {
	if halfLifeDays <= 0 {
		return current
	}
	if idleDays <= 0 {
		return current
	}
	if current <= 0 {
		return 0
	}
	factor := math.Pow(0.5, idleDays/halfLifeDays)
	next := current * factor
	if next < 0 {
		return 0
	}
	if next > 1 {
		return 1
	}
	return next
}

func truncateRunes(s string, max int) string {
	if max <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "…"
}
