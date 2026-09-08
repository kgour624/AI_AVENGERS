package decision

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/chinawall"
	"ai_avengers/backend/internal/gateway"
)

// ResponseMode is the output mode of the decision engine.
type ResponseMode string

const (
	ModeASK      ResponseMode = "ASK"
	ModeWARN     ResponseMode = "WARN"
	ModePUSHBACK ResponseMode = "PUSH_BACK"
	ModeREFUSE   ResponseMode = "REFUSE"
	ModeADVISE   ResponseMode = "ADVISE"
)

// Expert holds the data needed by the decision engine.
type Expert struct {
	ID                   uuid.UUID
	Name                 string
	Domain               string
	ReasoningCharter     string
	ClarificationCharter map[string][]string
}

// DecisionResult is the output of the 5-gate system.
type DecisionResult struct {
	Mode        ResponseMode
	Content     string
	Citations   []chinawall.Citation
	Confidence  float64
	GateStopped int      // 0 = reached Gate 5 successfully
	Warning     string   // Set when Gate 3 triggers WARN but continues
	Questions   []string // Set when Gate 1 triggers ASK
	Reason      string   // Set when Gate 5 refuses (China Wall reason)
}

// Engine implements the 5-gate decision system.
//
// Gate 1: Information Sufficiency — do we have enough context?
// Gate 2: Knowledge Coverage — is this in our domain?
// Gate 3: Charter Compliance — does charter say this is wrong?
// Gate 4: Necessity Check — is this actually needed?
// Gate 5: Generate Answer — China Wall enforced response
//
// WHY 5 gates:
// Each gate catches a different failure mode.
// Without Gate 1: experts answer vague questions with wrong assumptions.
// Without Gate 2: experts hallucinate outside their domain.
// Without Gate 3: experts give advice that contradicts their own principles.
// Without Gate 4: experts over-engineer simple problems.
// Without Gate 5: experts hallucinate even with good context.
type Engine struct {
	db        *pgxpool.Pool
	gateway   *gateway.ModelGateway
	chinaWall *chinawall.Enforcer
	logger    *zap.Logger
}

// NewEngine creates a new decision engine.
func NewEngine(db *pgxpool.Pool, gw *gateway.ModelGateway, cw *chinawall.Enforcer, logger *zap.Logger) *Engine {
	return &Engine{db: db, gateway: gw, chinaWall: cw, logger: logger}
}

// Process runs the question through all 5 gates.
// Returns a DecisionResult with mode and content.
func (e *Engine) Process(
	ctx context.Context,
	question string,
	expert Expert,
	chunks []chinawall.CourseChunk,
	projectSummary string,
	attempt int,
) (*DecisionResult, error) {

	e.logger.Debug("decision engine processing",
		zap.String("expert", expert.Name),
		zap.Int("attempt", attempt),
		zap.Int("chunks", len(chunks)),
	)

	// GATE 1: Information Sufficiency
	if result := e.gate1(question, expert); result != nil {
		return result, nil
	}
	// Bug 3.5 fix (docs bug list): gate1WithLLM was fully implemented
	// (zero-shot structured vagueness check) but never called from
	// anywhere - dead code. Its own doc comment says "Called when
	// keyword check is inconclusive", which is exactly this: gate1's
	// keyword-based check found nothing (returned nil), so fall back to
	// the LLM check before assuming the question is clear enough.
	if result := e.gate1WithLLM(ctx, question, expert); result != nil {
		return result, nil
	}

	// GATE 2: Knowledge Coverage
	if result := e.gate2(question, chunks, expert); result != nil {
		return result, nil
	}

	// GATE 3: Charter Compliance (may set warning but continue)
	warning := ""
	if result := e.gate3(ctx, question, expert); result != nil {
		if result.Mode == ModeWARN {
			warning = result.Warning // Continue with warning
		} else {
			return result, nil
		}
	}

	// GATE 4: Necessity Check
	if result := e.gate4(ctx, question, projectSummary, expert); result != nil {
		return result, nil
	}

	// GATE 5: Generate Answer (China Wall)
	// Pass expert.Domain so China Wall uses domain-based mode,
	// not question-based keyword detection.
	enforceResult, err := e.chinaWall.Enforce(
		ctx, question, chunks, expert.Name, expert.Domain, expert.ReasoningCharter, attempt,
	)
	if err != nil {
		return nil, fmt.Errorf("gate 5 failed: %w", err)
	}

	switch enforceResult.Status {
	case "retry":
		if attempt < 5 {
			return e.Process(ctx, question, expert, chunks, projectSummary, attempt+1)
		}
		return &DecisionResult{
			Mode:        ModeREFUSE,
			Content:     "Cannot provide a properly cited answer after multiple attempts.",
			GateStopped: 5,
		}, nil

	case "refused":
		return &DecisionResult{
			Mode:        ModeREFUSE,
			Content:     enforceResult.Answer,
			GateStopped: 5,
			Reason:      enforceResult.Reason,
		}, nil

	case "success":
		return &DecisionResult{
			Mode:       ModeADVISE,
			Content:    enforceResult.Answer,
			Citations:  enforceResult.Citations,
			Confidence: enforceResult.Confidence,
			Warning:    warning,
		}, nil
	}

	return &DecisionResult{Mode: ModeREFUSE, Content: "Unexpected error"}, nil
}

// gate1 checks if we have enough information to answer.
// Uses double-check pattern:
// 1. Fast keyword check (free) — catches obvious vague questions
// 2. LLM check (cheap) — only if keyword check flags vagueness
//
// WHY LLM for vagueness (Byte by Byte AI course):
// Course taught: zero-shot prompting with structured output.
// "Is this question specific enough? Return JSON: {clear: bool, missing: [...]}"
// Keyword matching alone misses nuanced vagueness.
// LLM double-check only runs when needed — avoids cost on clear questions.
func (e *Engine) gate1(question string, expert Expert) *DecisionResult {
	questionLower := strings.ToLower(question)

	// Domain bypass: problem-solving domain experts (DSA, coding, etc.)
	// never need clarification — every technical question is specific enough.
	// WHY domain not question: the expert's domain determines its behavior.
	// A DSA expert answers "what is python" directly. A medical expert might
	// need clarification on "what is python" (snake? programming language?).
	if chinawall.IsProblemSolvingDomain(expert.Domain) {
		return nil // Problem-solving experts skip Gate 1 entirely
	}

	// Fast check: obvious vague indicators
	vagueIndicators := []string{"how do i", "what should i", "best way", "recommend", "suggest"}
	isVague := false
	for _, indicator := range vagueIndicators {
		if strings.Contains(questionLower, indicator) {
			isVague = true
			break
		}
	}

	// Only proceed to LLM check if keyword check flagged AND question is short
	// WHY length check: Long questions usually have enough context
	if !isVague || len(question) >= 80 {
		return nil // Question is specific enough
	}

	// Find relevant clarification questions from charter
	var clarificationQs []string
	for topic, questions := range expert.ClarificationCharter {
		if strings.Contains(questionLower, strings.ToLower(topic)) {
			clarificationQs = append(clarificationQs, questions...)
			if len(clarificationQs) >= 3 {
				break
			}
		}
	}

	if len(clarificationQs) > 0 {
		return &DecisionResult{
			Mode:        ModeASK,
			Questions:   clarificationQs[:minInt(3, len(clarificationQs))],
			GateStopped: 1,
			Content:     "I need more information to give you an accurate answer.",
		}
	}

	return nil
}

// gate1WithLLM is the enhanced version using LLM for vagueness detection.
// Called when keyword check is inconclusive.
// Uses zero-shot structured output (Byte by Byte AI course pattern).
func (e *Engine) gate1WithLLM(ctx context.Context, question string, expert Expert) *DecisionResult {
	prompt := fmt.Sprintf(`Is this question specific enough to answer accurately?

Question: "%s"
Expert domain: %s

Return JSON only: {"clear": true/false, "missing": ["what info is missing"]}
If clear=true, missing should be empty.`, question, expert.Domain)

	resp, err := e.gateway.Call(ctx, gateway.LLMRequest{
		Model:       gateway.ModelCheap,
		UserPrompt:  prompt,
		MaxTokens:   100,
		Temperature: 0.1,
		UseCache:    true,
	})
	if err != nil {
		return nil // On error, don't block
	}

	// Parse response
	var result struct {
		Clear   bool     `json:"clear"`
		Missing []string `json:"missing"`
	}

	clean := strings.TrimSpace(resp.Content)
	clean = strings.TrimPrefix(clean, "```json")
	clean = strings.TrimPrefix(clean, "```")
	clean = strings.TrimSuffix(clean, "```")
	clean = strings.TrimSpace(clean)

	start := strings.Index(clean, "{")
	end := strings.LastIndex(clean, "}")
	if start == -1 || end == -1 {
		return nil
	}

	if err := parseJSONDecision([]byte(clean[start:end+1]), &result); err != nil {
		return nil
	}

	if !result.Clear && len(result.Missing) > 0 {
		// Build clarification questions from missing info
		var questions []string
		for _, m := range result.Missing {
			questions = append(questions, "What is your "+m+"?")
		}
		return &DecisionResult{
			Mode:        ModeASK,
			Questions:   questions[:minInt(3, len(questions))],
			GateStopped: 1,
			Content:     "I need more information to give you an accurate answer.",
		}
	}
	return nil
}

// gate2 checks if the question is in our knowledge domain.
func (e *Engine) gate2(question string, chunks []chinawall.CourseChunk, expert Expert) *DecisionResult {
	if len(chunks) == 0 {
		return &DecisionResult{
			Mode:        ModeREFUSE,
			Content:     fmt.Sprintf("This topic is not in my training material. I specialize in: %s", expert.Domain),
			GateStopped: 2,
		}
	}

	// Check best rerank score
	bestScore := float32(0)
	for _, c := range chunks {
		if c.RerankScore > bestScore {
			bestScore = c.RerankScore
		}
	}

	// Hard refuse if score is very low
	if bestScore < 0.20 {
		return &DecisionResult{
			Mode:        ModeREFUSE,
			Content:     fmt.Sprintf("This question doesn't match my training content (relevance: %.0f%%). I cover: %s", bestScore*100, expert.Domain),
			GateStopped: 2,
		}
	}

	return nil
}

// gate3 checks if the charter says this approach is wrong.
func (e *Engine) gate3(ctx context.Context, question string, expert Expert) *DecisionResult {
	if expert.ReasoningCharter == "" {
		return nil
	}

	// Quick keyword check before LLM call
	// WHY: Avoid LLM call for questions that clearly don't violate charter
	charterLower := strings.ToLower(expert.ReasoningCharter)
	questionLower := strings.ToLower(question)

	// Check for "never" violations
	neverPattern := neverRulePattern
	matches := neverPattern.FindAllStringSubmatch(charterLower, -1)

	for _, match := range matches {
		if len(match) > 1 && strings.Contains(questionLower, match[1]) {
			return &DecisionResult{
				Mode:        ModeWARN,
				Warning:     fmt.Sprintf("My charter says: Never %s. Proceeding with caution.", match[1]),
				GateStopped: 0, // Warning, not stop
			}
		}
	}

	return nil
}

// gate4 checks if the solution is actually necessary.
func (e *Engine) gate4(ctx context.Context, question string, projectSummary string, expert Expert) *DecisionResult {
	// Only check if we have project context
	if projectSummary == "" {
		return nil
	}

	// Check for over-engineering signals
	overEngineeringTerms := []string{
		"kubernetes", "microservices", "distributed", "kafka", "redis cluster",
		"sharding", "multi-region", "auto-scaling",
	}

	questionLower := strings.ToLower(question)
	for _, term := range overEngineeringTerms {
		if strings.Contains(questionLower, term) {
			// Ask LLM if this is premature for the project
			result := e.checkNecessity(ctx, question, projectSummary, term)
			if result != nil {
				return result
			}
			break // Only check first match
		}
	}

	return nil
}

// checkNecessity uses cheap LLM to check if solution is premature.
func (e *Engine) checkNecessity(ctx context.Context, question, projectSummary, term string) *DecisionResult {
	prompt := fmt.Sprintf(`Project context: %s

User wants to implement: %s

Is this premature for this project? Reply with only YES or NO.`,
		projectSummary[:minInt(300, len(projectSummary))], term)

	resp, err := e.gateway.Call(ctx, gateway.LLMRequest{
		Model:       gateway.ModelCheap,
		UserPrompt:  prompt,
		MaxTokens:   5,
		Temperature: 0.1,
		UseCache:    true,
	})
	if err != nil {
		return nil // On error, don't push back
	}

	if strings.Contains(strings.ToUpper(resp.Content), "YES") {
		return &DecisionResult{
			Mode:        ModePUSHBACK,
			Content:     fmt.Sprintf("Based on your project context, %s may be premature. Let's verify you actually need this before implementing.", term),
			GateStopped: 4,
		}
	}
	return nil
}

// parseJSONDecision is a helper to unmarshal JSON for decision engine.
func parseJSONDecision(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// neverRulePattern compiled once at package level for performance.
var neverRulePattern = regexp.MustCompile(`never\s+(\w+(?:\s+\w+){0,5})`)

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}


