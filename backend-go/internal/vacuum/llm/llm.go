package llm

import "context"

// LLMCaller is the minimal gateway contract vacuum needs (subset of gateway.ModelGateway).
// Implemented by gateway.ModelGateway and by test doubles.
type LLMCaller interface {
	Call(ctx context.Context, req LLMRequest) (*LLMResponse, error)
}

// LLMRequest mirrors gateway.LLMRequest but decoupled to avoid import cycle.
type LLMRequest struct {
	Model        string
	SystemPrompt string
	UserPrompt   string
	MaxTokens    int
	Temperature  float64
}

// LLMResponse mirrors gateway.LLMResponse trimmed.
type LLMResponse struct {
	Content      string
	InputTokens  int
	OutputTokens int
	ModelUsed    string
}

// Classifier — Phase 5 tiny classifier (Gemini Flash temp 0.1).
// I: chunk text -> P: LLM classify -> O: label + confidence
type Classifier interface {
	Classify(ctx context.Context, text string) (label string, confidence float64, err error)
}

// HeadingGenerator — Phase 5 heading gen (Claude Sonnet, ## / ### only).
// I: cleaned text / chunk -> P: LLM headings -> O: []Heading (no content mutation)
type Heading struct {
	Level int    // 2 or 3
	Text  string // heading text without prefix
	Raw   string // full line e.g. "## Introduction"
}

type HeadingGenerator interface {
	Generate(ctx context.Context, text string) ([]Heading, error)
	Validate(headings []Heading) error
}

// PreservationGuard — Phase 5 No-Trust SHA256 guard.
// I: original SHA + cleaned text + claimed headings -> P: hash verify -> O: verified bool
type PreservationGuard interface {
	Verify(ctx context.Context, originalSHA256 string, cleanedText string) (verified bool, computedSHA string, err error)
	VerifyHash(originalSHA256, cleanedSHA256 string) bool
}
