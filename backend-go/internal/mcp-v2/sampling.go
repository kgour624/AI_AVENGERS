package mcpv2

import "context"

// Sampling borrows Host LLM via server.server.createMessage (MCP 3&4 spec).
// Checks getClientCapabilities()?.sampling else sendLoggingMessage.
// includeContext:"none" never allServers, void fire-and-forget.

type SamplingRequest struct {
	SystemPrompt string
	Messages     []Message
	MaxTokens    int
}

type SamplingClient interface {
	CreateMessage(ctx context.Context, req SamplingRequest) (string, error)
	HasSampling() bool
}

// SampleWithHost demonstrates sampling flow for ask_expert enrichment.
func SampleWithHost(ctx context.Context, client SamplingClient, req SamplingRequest) (string, error) {
	if client == nil || !client.HasSampling() {
		return "", nil
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = 100
	}
	return client.CreateMessage(ctx, req)
}