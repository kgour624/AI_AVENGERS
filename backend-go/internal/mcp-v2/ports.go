package mcpv2

import (
	"context"
	"ai_avengers/backend/internal/mcp-v2/business"
)

// Re-export Business models as App models — App imports Business (Ultimate Go §1, §7 — imports down only).
// App has its own JSON-tagged models if needed, but reuses Business core via mapping + parse().
type Expert = business.Expert
type Chunk = business.Chunk
type RepoChunk = business.RepoChunk
type UsageLog = business.UsageLog
type ExpertLimits = business.ExpertLimits
type Message = business.Message
type GatewayUsage = business.GatewayUsage
type GateResult = business.GateResult

// Small interfaces at boundaries (RULE 8-B:32) — Discover, don't design. Input uses interface, return concrete.
// Business defines Storer; App reuses it via business.Storer ( §7 — storage under business domain).
type DBPort = business.Storer

// RedisPort for atomic counters + Pub/Sub wakeup only (RULE 8-F:46).
type RedisPort interface {
	IncrTokens(ctx context.Context, key string, delta int64) (int64, error)
	Publish(ctx context.Context, channel string, payload []byte) error
}

// GatewayPort wraps internal/gateway ModelGateway.Call (cheap|strong|fast).
type GatewayPort interface {
	Call(ctx context.Context, tier string, messages []Message) (output string, usage GatewayUsage, err error)
}

// DecisionPort wraps internal/decision.Engine gate2 (no mutation).
type DecisionPort interface {
	Check(ctx context.Context, expertID, question string, chunks []Chunk) (GateResult, error)
}

// VectorPort wraps ml-sidecar bge-base-en-v1.5 embed.
type VectorPort interface {
	Embed(ctx context.Context, text string) ([]float32, error)
}
