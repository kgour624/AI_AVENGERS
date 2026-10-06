package llm

import "context"

// GatewayAdapter bridges gateway.ModelGateway to vacuum/llm.LLMCaller without import cycle.
// The gateway package builds this via interface; vacuum never imports gateway directly.
//
// Usage in cmd/server/main.go:
//   vacuumAdapter := llm.NewGatewayAdapter(modelGateway)
//   classifier := llm.NewGeminiClassifier(vacuumAdapter)
//   headingGen := llm.NewClaudeHeadingGenerator(vacuumAdapter)
//
// modelGateway must implement Call(ctx, req) (*Resp, error) where req has Model/SystemPrompt/UserPrompt/MaxTokens/Temperature.
// We use a tiny interface to avoid importing gateway types (breaks cycle).
type GatewayCaller interface {
	Call(ctx context.Context, model, systemPrompt, userPrompt string, maxTokens int, temperature float64) (string, error)
}

// gatewayAdapter wraps any object that has the full gateway LLMRequest/LLMResponse Call.
type gatewayAdapter struct {
	inner interface {
		Call(ctx context.Context, req interface{}) (interface{}, error)
	}
}

// Simple functional adapter — caller provides a closure.
type funcAdapter struct {
	fn func(ctx context.Context, req LLMRequest) (*LLMResponse, error)
}

func NewFuncAdapter(fn func(ctx context.Context, req LLMRequest) (*LLMResponse, error)) LLMCaller {
	return &funcAdapter{fn: fn}
}

func (f *funcAdapter) Call(ctx context.Context, req LLMRequest) (*LLMResponse, error) {
	return f.fn(ctx, req)
}
