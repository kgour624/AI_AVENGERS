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

// ── Phase 1 — Kachra Brain (Hybrid DSA+LLM foundation) ──

// KachraSpan is LLM's suggestion — strictly {text, reason, type, confidence}.
// LLM NEVER mutates content; DSA is the sole remover via byte-exact mapping.
type KachraSpan struct {
	Text       string  `json:"text"`
	Reason     string  `json:"reason"`
	Type       string  `json:"type"`       // filler|repetition|asr_error|classroom_meta|hinglish|logistics|semantic_noise|smalltalk
	Confidence float64 `json:"confidence"` // 0.0 - 1.0
}

// KachraDetector — LLM suggests kachra spans, DSA validates & removes.
// Contract: Input chunk text -> Process LLM strict JSON temp 0.0 -> Output []KachraSpan (no mutation).
type KachraDetector interface {
	Detect(ctx context.Context, chunkText string) ([]KachraSpan, error)
}

// KachraVerifier — LLM self cross-verify: re-checks each suggested span.
// Contract: Input (chunkText + candidate spans) -> Process LLM temp 0.0 -> Output filtered spans (confidence >= 0.70).
type KachraVerifier interface {
	Verify(ctx context.Context, chunkText string, spans []KachraSpan) ([]KachraSpan, error)
}

// KachraSink — persists LLM suggestions to candidate_kachra (status=pending) for human approve.
// Future seed: approval bumps kachra_patterns + brain_version -> 30s hot-reload.
type KachraSink interface {
	Save(ctx context.Context, spans []KachraSpan, chunkText string) (int, error)
	SaveWithFileID(ctx context.Context, fileJobID string, spans []KachraSpan, chunkText string) (int, error)
}
