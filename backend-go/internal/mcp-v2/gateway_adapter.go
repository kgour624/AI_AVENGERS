package mcpv2

import (
	"context"
	"fmt"
)

// GatewayAdapter wraps existing internal/gateway ModelGateway via GatewayPort (reuse 3 tiers cheap|strong|fast).
// Reuses system_settings.models + llm_settings — never creates parallel config ( §3.8 ).
// Vault: api_key_ref never plaintext — decrypted via existing AES-256 pattern from repo_connections.
type GatewayAdapter struct {
	// underlying is injected via interface to avoid import cycle — real wiring in cmd/server/main.go will pass gateway.Service
	underlying GatewayPort
}

func NewGatewayAdapter(underlying GatewayPort) *GatewayAdapter {
	return &GatewayAdapter{underlying: underlying}
}

func (g *GatewayAdapter) Call(ctx context.Context, tier string, messages []Message) (string, GatewayUsage, error) {
	if tier != "cheap" && tier != "strong" && tier != "fast" {
		tier = "strong"
	}
	if g.underlying == nil {
		return "", GatewayUsage{}, fmt.Errorf("gateway not wired — pending §6 mount")
	}
	return g.underlying.Call(ctx, tier, messages)
}

// ResolveProvider implements §3.8 Resolution Flow: mcp_expert_provider_map -> fallback system_settings.models[tier] -> llm_settings check.
func (g *GatewayAdapter) ResolveProvider(ctx context.Context, expertID, platform, tier string) (provider, model string, err error) {
	// 1. Check mcp_expert_provider_map via DBPort if row exists
	// 2. Fallback to global system_settings (via same DB)
	// 3. Validate provider is active in llm_settings — if not, return "Provider not enabled"
	// Real SQL in storage/postgres.go — stub keeps interface stable
	return "openrouter", "anthropic/claude-3-5-sonnet", nil
}