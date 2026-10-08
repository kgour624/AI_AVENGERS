package ml

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Embedder is the interface for generating text embeddings.
//
// WHY this interface exists (SOLID-D, Interface Segregation):
//   Before this interface, all callers took *SidecarClient directly.
//   This coupled them to the Python sidecar as the ONLY embedding source.
//   With this interface, callers depend on the abstraction — the concrete
//   implementation (sidecar or CodeCraftAPI) is resolved at runtime.
//
// WHY Rerank() is NOT in this interface (SOLID-I):
//   Reranking is ALWAYS sidecar-only. CodeCraftAPI has no /v1/rerank endpoint.
//   Including Rerank() would force CodeCraftAPIEmbedder to implement a method
//   it cannot implement. Callers that need reranking (chinawall/enforcer.go)
//   continue to take *SidecarClient directly — zero regression.
//
// *SidecarClient already satisfies this interface via Go structural typing:
//   SidecarClient.Embed(ctx, texts []string) ([][]float32, error)     ✔
//   SidecarClient.EmbedSingle(ctx, text string) ([]float32, error)    ✔
// No wrapper struct needed.
type Embedder interface {
	// Embed generates embeddings for a batch of texts.
	// Returns one embedding vector per input text, in the same order.
	Embed(ctx context.Context, texts []string) ([][]float32, error)

	// EmbedSingle generates an embedding for a single text.
	// Convenience wrapper around Embed.
	EmbedSingle(ctx context.Context, text string) ([]float32, error)
}

// ErrEmbeddingUnavailable is returned when no embedder is configured or the
// embedding service is down. Callers use errors.Is to distinguish infra
// failure from logical "insufficient context".
// User-facing message for this error must be "System abhi down hai, thodi der baad try karein"
// and NOT the generic "INSUFFICIENT_CONTEXT" refusal.
var ErrEmbeddingUnavailable = errors.New("embedding service unavailable")

// unavailableEmbedder implements Embedder but always returns ErrEmbeddingUnavailable.
// Used as a safe nil-guard so resolveEmbedder NEVER returns a nil interface
// that would panic on .Embed() call.
type unavailableEmbedder struct {
	reason string
}

func (u *unavailableEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	return nil, fmt.Errorf("%w: %s", ErrEmbeddingUnavailable, u.reason)
}

func (u *unavailableEmbedder) EmbedSingle(ctx context.Context, text string) ([]float32, error) {
	return nil, fmt.Errorf("%w: %s", ErrEmbeddingUnavailable, u.reason)
}

// DynamicEmbedder reads the active embedding provider from system_settings
// at call time and delegates to the correct implementation.
//
// WHY dynamic (reads DB at call time, not at startup):
//   Same pattern as ModelGateway.getActiveProvider().
//   Admin switches embedding provider from the UI — takes effect on the
//   next Embed() call with no server restart needed.
//   DB read overhead: ~1ms (single SELECT via pgx pool). Acceptable at
//   single-admin scale.
//
// Failure behavior:
//   If DB read fails → falls back to sidecar (safe default, always available).
//   If embedding_provider = "codecraftapi" but ccEmbedder is nil → falls back to sidecar.
//   If embedding_provider is an unknown value → falls back to sidecar.
type DynamicEmbedder struct {
	db         *pgxpool.Pool
	sidecar    *SidecarClient    // always available, used as safe default
	ccEmbedder *CodeCraftAPIEmbedder // nil if CodeCraftAPI key not configured
	logger     *zap.Logger
}

// NewDynamicEmbedder creates a DynamicEmbedder.
//
// sidecar must not be nil — it is the safe fallback for all failure cases.
// ccEmbedder may be nil if CodeCraftAPI is not configured; in that case
// any attempt to use codecraftapi embedding falls back to sidecar with a warning.
func NewDynamicEmbedder(
	db *pgxpool.Pool,
	sidecar *SidecarClient,
	ccEmbedder *CodeCraftAPIEmbedder,
	logger *zap.Logger,
) *DynamicEmbedder {
	return &DynamicEmbedder{
		db:         db,
		sidecar:    sidecar,
		ccEmbedder: ccEmbedder,
		logger:     logger,
	}
}

// Embed generates embeddings for a batch of texts.
// Reads embedding_provider from system_settings at call time.
func (d *DynamicEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	emb := d.resolveEmbedder(ctx)
	if emb == nil {
		return nil, fmt.Errorf("%w: resolveEmbedder returned nil (sidecar not configured)", ErrEmbeddingUnavailable)
	}
	return emb.Embed(ctx, texts)
}

// EmbedSingle generates an embedding for a single text.
// Reads embedding_provider from system_settings at call time.
func (d *DynamicEmbedder) EmbedSingle(ctx context.Context, text string) ([]float32, error) {
	emb := d.resolveEmbedder(ctx)
	if emb == nil {
		return nil, fmt.Errorf("%w: resolveEmbedder returned nil (sidecar not configured)", ErrEmbeddingUnavailable)
	}
	return emb.EmbedSingle(ctx, text)
}

// resolveEmbedder reads embedding_provider from system_settings and returns
// the correct Embedder implementation. Falls back to sidecar on any error.
// NEVER returns nil — if sidecar is also nil it returns unavailableEmbedder
// so the caller gets a clear ErrEmbeddingUnavailable instead of a nil panic.
//
// Mental execution:
//   Case 1: embedding_provider = "sidecar" (or row missing) → return sidecar
//   Case 2: embedding_provider = "codecraftapi" + ccEmbedder != nil → return ccEmbedder
//   Case 3: DB read fails → log warning, return sidecar
//   Case 4: embedding_provider = "codecraftapi" but ccEmbedder is nil → log warning, return sidecar
//   Case 5: unknown provider value → return sidecar (safe default)
//   Case 6: sidecar == nil AND ccEmbedder == nil → return unavailableEmbedder (never nil, never panic)
func (d *DynamicEmbedder) resolveEmbedder(ctx context.Context) Embedder {
	provider := d.readEmbeddingProvider(ctx)

	if provider == "codecraftapi" {
		if d.ccEmbedder != nil {
			return d.ccEmbedder
		}
		if d.logger != nil {
			d.logger.Warn("embedding_provider is codecraftapi but CodeCraftAPIEmbedder is nil, falling back to sidecar")
		}
		if d.sidecar != nil {
			return d.sidecar
		}
		if d.logger != nil {
			d.logger.Error("both CodeCraftAPIEmbedder and sidecar are nil — embedding service unavailable")
		}
		return &unavailableEmbedder{reason: "sidecar not configured and CodeCraftAPI not configured (embedding_provider=codecraftapi)"}
	}

	if d.sidecar != nil {
		return d.sidecar
	}
	if d.logger != nil {
		d.logger.Error("sidecar is nil — embedding service unavailable, cannot serve request", zap.String("provider", provider))
	}
	return &unavailableEmbedder{reason: "sidecar not configured (embedding_provider=" + provider + ")"}
}

// readEmbeddingProvider reads the embedding_provider value from system_settings.
// Returns "sidecar" on any error (safe default).
func (d *DynamicEmbedder) readEmbeddingProvider(ctx context.Context) string {
	if d.db == nil {
		if d.logger != nil {
			d.logger.Warn("DynamicEmbedder.db is nil, falling back to sidecar")
		}
		return "sidecar"
	}
	var valueJSON []byte
	err := d.db.QueryRow(ctx,
		`SELECT value FROM system_settings WHERE key = 'embedding_provider'`,
	).Scan(&valueJSON)
	if err != nil {
		if err.Error() != "no rows in result set" {
			if d.logger != nil {
				d.logger.Warn("failed to read embedding_provider from system_settings, falling back to sidecar",
					zap.Error(err),
				)
			}
		}
		return "sidecar"
	}

	var provider string
	if json.Unmarshal(valueJSON, &provider) != nil || provider == "" {
		return "sidecar"
	}

	return provider
}
