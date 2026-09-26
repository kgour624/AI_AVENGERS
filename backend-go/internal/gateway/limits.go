package gateway

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ErrInputTooLong marks a request whose prompt cannot fit the model it is
// addressed to.
//
// It is fail-closed on purpose: the provider would reject the request anyway,
// but with a vendor-specific message the admin cannot act on. Refusing here
// names the model, the estimated input size and the configured limit, all of
// which are editable in Admin → LLM Settings.
var ErrInputTooLong = errors.New("input exceeds the configured token limit for this model")

// ModelLimit is the maximum this system will ask one model for.
//
// A zero value means "nothing configured" — the caller then falls back to the
// provider's own maximum, so an empty table changes no behaviour at all.
type ModelLimit struct {
	Provider string `json:"provider"`
	// Tier is one of the tiers the gateway selects models by ('strong', 'fast',
	// 'cheap') or '*' for a provider-wide row.
	Tier string `json:"tier"`
	// MaxInputTokens caps the prompt. 0 = not enforced.
	//
	// WHY the input side is a limit and not a trimming step: silently dropping
	// context changes the answer without telling anyone, and the caller (chat
	// assembler, planner, authoring prompt) is the only side that knows which
	// part is expendable. Refusing with the numbers lets that caller trim on
	// purpose or the admin raise the limit.
	MaxInputTokens int `json:"max_input_tokens"`
	// MaxOutputTokens caps the completion, reasoning tokens included.
	//
	// When set, this number is the ceiling — up OR down — because it is the only
	// place that knows what the specific model behind the key can do. The
	// per-provider defaults in the code are conservative guesses (anthropic 8192,
	// long outdated for current Claude models) and cannot be trusted as a cap for
	// a model the admin configured by name. A value the vendor rejects comes back
	// as a provider error, and lowering it here needs no redeploy.
	MaxOutputTokens int       `json:"max_output_tokens"`
	UpdatedAt       time.Time `json:"updated_at,omitempty"`
}

// Tier keys used by llm_model_limits.tier.
const (
	LimitTierStrong  = "strong"
	LimitTierFast    = "fast"
	LimitTierCheap   = "cheap"
	LimitTierDefault = "*"
)

// modelLimitsCacheTTL mirrors the gate-threshold cache: long enough that a busy
// gateway does not hit the database on every call, short enough that an admin
// edit takes effect without a restart.
const modelLimitsCacheTTL = 30 * time.Second

// limitTier maps a model tier to the key its limit row is stored under.
func limitTier(t ModelType) string {
	switch t {
	case ModelStrong:
		return LimitTierStrong
	case ModelFast:
		return LimitTierFast
	case ModelCheap:
		return LimitTierCheap
	default:
		return LimitTierDefault
	}
}

// pickModelLimit chooses the row that applies: the tier's own row first, then
// the provider's wildcard row, then nothing.
//
// Pure, so the precedence is unit-tested rather than inferred.
func pickModelLimit(rows []ModelLimit, provider, tier string) (ModelLimit, bool) {
	var wildcard ModelLimit
	haveWildcard := false
	for _, r := range rows {
		if !strings.EqualFold(r.Provider, provider) {
			continue
		}
		if r.Tier == tier {
			return r, true
		}
		if r.Tier == LimitTierDefault {
			wildcard = r
			haveWildcard = true
		}
	}
	return wildcard, haveWildcard
}

// effectiveOutputBudget decides the completion budget to send.
//
// requested <= 0 means the caller has no opinion, and the answer is the ceiling
// — that is what makes a planner or authoring call work on a reasoning model
// without every call site hard-coding a number it cannot know.
//
// An explicit request is honoured, only ever narrowed by the ceiling. Callers
// that ask for little (a judge asking for a 200-token verdict) mean it; the
// empty-content escalation in Call() is what rescues those when a reasoning
// model spends the budget on thinking.
func effectiveOutputBudget(requested, ceiling int) int {
	if ceiling <= 0 {
		// No ceiling known: keep the historical 2000-token default rather than
		// sending 0, which providers read as "use your default".
		if requested <= 0 {
			return 2000
		}
		return requested
	}
	if requested <= 0 || requested > ceiling {
		return ceiling
	}
	return requested
}

// resolveCeiling decides the completion ceiling for one call: the admin's
// configured maximum when one is set, otherwise the provider's own.
//
// Pure, so the precedence is unit-tested. WHY the configured value wins even
// when it is LARGER than the provider default: those defaults are stale guesses
// baked into the code, and the number this feature exists to set is exactly the
// one that has to be able to go up. A reasoning model that needs 16000 has no
// use for a ceiling of 8192 — and the escalation retry cannot rescue a call
// whose ceiling is the same number that was already exhausted.
func resolveCeiling(providerMax, configuredMax int) int {
	if configuredMax > 0 {
		return configuredMax
	}
	return providerMax
}

// estimateTokens approximates a token count from text, mirroring the ratio used
// by internal/context's estimator (chars/4). Duplicated rather than imported:
// internal/context depends on this package, so importing it back would invert
// the layering.
func estimateTokens(text string) int {
	if text == "" {
		return 0
	}
	return len(text) / 4
}

// estimateMessagesTokens sums the estimated size of a whole prompt.
func estimateMessagesTokens(msgs []ProviderMessage) int {
	total := 0
	for _, m := range msgs {
		total += estimateTokens(m.Content)
	}
	return total
}

// resolveModelLimit returns the configured limit for (provider, tier) and
// whether one was configured at all.
func (g *ModelGateway) resolveModelLimit(ctx context.Context, provider string, tier ModelType) (ModelLimit, bool) {
	return pickModelLimit(g.modelLimits(ctx), provider, limitTier(tier))
}

// modelLimits returns the cached limit rows, refreshing them at most once per
// modelLimitsCacheTTL.
//
// A read failure is logged and treated as "nothing configured": a limit table
// that cannot be read must never take the whole gateway down with it.
func (g *ModelGateway) modelLimits(ctx context.Context) []ModelLimit {
	g.limitsMu.RLock()
	snap := g.limitsSnap
	g.limitsMu.RUnlock()
	if snap != nil && time.Since(snap.at) < modelLimitsCacheTTL {
		return snap.rows
	}
	if g.db == nil {
		return nil
	}
	rows, err := g.ListModelLimits(ctx)
	if err != nil {
		g.logger.Warn("model limits read failed; using provider defaults", zap.Error(err))
		return nil
	}
	g.limitsMu.Lock()
	g.limitsSnap = &modelLimitsSnapshot{rows: rows, at: time.Now()}
	g.limitsMu.Unlock()
	return rows
}

// modelLimitsSnapshot is the cached read of llm_model_limits.
type modelLimitsSnapshot struct {
	rows []ModelLimit
	at   time.Time
}

// ListModelLimits reads every configured limit row.
func (g *ModelGateway) ListModelLimits(ctx context.Context) ([]ModelLimit, error) {
	if g.db == nil {
		return nil, nil
	}
	rows, err := g.db.Query(ctx,
		`SELECT provider, tier, max_input_tokens, max_output_tokens, updated_at
		   FROM llm_model_limits
		  ORDER BY provider, tier`,
	)
	if err != nil {
		return nil, fmt.Errorf("list model limits: %w", err)
	}
	defer rows.Close()

	var out []ModelLimit
	for rows.Next() {
		var l ModelLimit
		if err := rows.Scan(&l.Provider, &l.Tier, &l.MaxInputTokens, &l.MaxOutputTokens, &l.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan model limit: %w", err)
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

// UpsertModelLimits writes the given rows and drops that cache so the change is
// live within the next call rather than after the TTL.
func (g *ModelGateway) UpsertModelLimits(ctx context.Context, updatedBy *uuid.UUID, limits []ModelLimit) error {
	if g.db == nil {
		return fmt.Errorf("model limits: database not configured")
	}
	tx, err := g.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("model limits: begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	for _, l := range limits {
		if _, err := tx.Exec(ctx,
			`INSERT INTO llm_model_limits
			     (provider, tier, max_input_tokens, max_output_tokens, updated_by, updated_at)
			 VALUES ($1, $2, $3, $4, $5, NOW())
			 ON CONFLICT (provider, tier) DO UPDATE SET
			     max_input_tokens  = EXCLUDED.max_input_tokens,
			     max_output_tokens = EXCLUDED.max_output_tokens,
			     updated_by        = EXCLUDED.updated_by,
			     updated_at        = NOW()`,
			l.Provider, l.Tier, l.MaxInputTokens, l.MaxOutputTokens, updatedBy,
		); err != nil {
			return fmt.Errorf("upsert model limit %s/%s: %w", l.Provider, l.Tier, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("model limits: commit: %w", err)
	}

	g.limitsMu.Lock()
	g.limitsSnap = nil
	g.limitsMu.Unlock()
	return nil
}

// invalidateModelLimits drops the cache so the next call re-reads the table.
// Called after an admin write.
func (g *ModelGateway) invalidateModelLimits() {
	g.limitsMu.Lock()
	g.limitsSnap = nil
	g.limitsMu.Unlock()
}
