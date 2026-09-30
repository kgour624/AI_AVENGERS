package mcpv2

import (
	"context"
	"fmt"

	"ai_avengers/backend/internal/gateway"
)

// RealGatewayAdapter maps tier (cheap|strong|fast) -> ModelGateway.Call via Provider.
// No new LLM integration — reuses ModelGateway + system_settings + C5 usage single choke point.
type RealGatewayAdapter struct {
	gateway *gateway.ModelGateway
}

func NewRealGatewayAdapter(g *gateway.ModelGateway) *RealGatewayAdapter {
	return &RealGatewayAdapter{gateway: g}
}

func (r *RealGatewayAdapter) Call(ctx context.Context, tier string, messages []Message) (string, GatewayUsage, error) {
	if r.gateway == nil {
		return "", GatewayUsage{}, fmt.Errorf("gateway not wired")
	}
	if tier != "cheap" && tier != "strong" && tier != "fast" {
		tier = "strong"
	}
	// Map tier -> ModelType via gateway types
	var providerMsgs []gateway.ProviderMessage
	for _, m := range messages {
		role := m.Role
		if role == "" {
			role = "user"
		}
		providerMsgs = append(providerMsgs, gateway.ProviderMessage{Role: role, Content: m.Content})
	}

	var modelType gateway.ModelType
	switch tier {
	case "strong":
		modelType = gateway.ModelStrong
	case "fast":
		modelType = gateway.ModelFast
	default:
		modelType = gateway.ModelCheap
	}

	req := gateway.LLMRequest{
		Model:    modelType,
		Messages: providerMsgs,
	}
	resp, err := r.gateway.Call(ctx, req)
	if err != nil {
		return "", GatewayUsage{}, err
	}
	return resp.Content, GatewayUsage{InputTokens: resp.InputTokens, OutputTokens: resp.OutputTokens, CostUSD: resp.CostUSD}, nil
}
