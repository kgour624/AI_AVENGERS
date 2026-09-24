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

	"ai_avengers/backend/internal/category"
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
	// TemplateSections/DefaultLanguage (CT-B4): populated by the caller
	// (orchestrator.go) ONLY when this expert has a category_id whose
	// category has a non-empty template_schema. nil/"" for every other
	// expert (CT-L2) — Process()/chinaWall.Enforce() take the existing
	// flat-text path unchanged in that case.
	TemplateSections []category.TemplateSection
	DefaultLanguage  string
	// AskStructurePermission (CT-C4): true only when this expert's
	// category has ask_structure_permission=true. false for every
	// non-categorized expert or category without the flag (CT-L2).
	AskStructurePermission bool
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
	// TemplateSections (CT-B4): mirrors chinawall.EnforceResult.TemplateSections.
	// nil for every flat-text expert response (CT-L2).
	TemplateSections []chinawall.TemplateSectionResult
	// Claims (B8): mirrors chinawall.EnforceResult.Claims. nil when claim
	// verify disabled / fail-open / structured path.
	Claims []chinawall.ClaimReport
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
//
// replyToMessageID (CT-C4): nil for a fresh question (the vast majority
// of calls, and always nil on Gate 5's internal retry recursion below —
// see that call site). Non-nil only on the top-level call for an actual
// reply. Only relevant when expert.AskStructurePermission is true;
// otherwise gateStructurePermission is a no-op and this parameter has
// zero effect on behavior (CT-L2-style fallback, applied to categories
// instead of experts).
func (e *Engine) Process(
	ctx context.Context,
	question string,
	expert Expert,
	chunks []chinawall.CourseChunk,
	projectSummary string,
	replyContext string,
	attempt int,
	replyToMessageID *uuid.UUID,
	// tokenCh: non-nil enables streaming for Gate 5 generation.
	// nil = blocking (backward compatible).
	tokenCh chan<- string,
) (*DecisionResult, error) {

	e.logger.Debug("decision engine processing",
		zap.String("expert", expert.Name),
		zap.Int("attempt", attempt),
		zap.Int("chunks", len(chunks)),
	)

	// GATE 0: Structure-permission gate (CT-C4, CATEGORY_TEMPLATE_HANDOFF.md §6).
	// No-op (returns nil, "") when expert.AskStructurePermission is false —
	// every existing expert's behavior is completely unaffected.
	if gateResult, rewrittenQuestion := e.gateStructurePermission(ctx, question, expert, replyToMessageID); gateResult != nil {
		return gateResult, nil
	} else if rewrittenQuestion != "" {
		question = rewrittenQuestion
	}

	// GATE 1: Information Sufficiency
	// needsLLM is true only when the keyword path was inconclusive
	// (vague + short + no charter match). Domain-skip and "already
	// specific" return needsLLM=false so gate1WithLLM is never called
	// (A13 — previously every nil from gate1 defeated Gate1Skip).
	if result, needsLLM := e.gate1(question, expert); result != nil {
		return result, nil
	} else if needsLLM {
		if result := e.gate1WithLLM(ctx, question, expert); result != nil {
			return result, nil
		}
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
	// Pass expert.TemplateSections/DefaultLanguage (CT-B4) — nil/"" for
	// any expert with no category or an empty template_schema, in which
	// case Enforce() takes its existing flat-text path unchanged (CT-L2).
	enforceResult, err := e.chinaWall.Enforce(
		ctx, question, chunks, expert.Name, expert.Domain, expert.ReasoningCharter, replyContext, attempt,
		expert.TemplateSections, expert.DefaultLanguage, tokenCh,
	)
	if err != nil {
		return nil, fmt.Errorf("gate 5 failed: %w", err)
	}

	switch enforceResult.Status {
	case "retry":
		if attempt < 5 {
			// replyToMessageID=nil on retry: Gate 0 already resolved (if
			// applicable) on the top-level call and rewrote `question` in
			// place — re-passing a non-nil replyToMessageID here would
			// re-run gateStructurePermission's DB lookup against the
			// ALREADY-rewritten question and incorrectly re-append the
			// preference text a second time.
			//
			// tokenCh is passed through (not nil) so streaming continues
			// on retry — user sees continuous token flow, not frozen stream.
		return e.Process(ctx, question, expert, chunks, projectSummary, replyContext, attempt+1, nil, tokenCh)
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
			Mode:             ModeADVISE,
			Content:          enforceResult.Answer,
			Citations:        enforceResult.Citations,
			Confidence:       enforceResult.Confidence,
			Warning:          warning,
			TemplateSections: enforceResult.TemplateSections,
			Claims:           enforceResult.Claims,
		}, nil

	case "partial":
		// China Wall returned partial coverage after multiple attempts.
		// enforceResult.Answer is empty (enforcer.go:174 only sets Reason).
		// Return clear refusal explaining what's missing.
		return &DecisionResult{
			Mode:        ModeREFUSE,
			Content:     fmt.Sprintf("Cannot provide complete answer: %s", enforceResult.Reason),
			GateStopped: 5,
			Reason:      enforceResult.Reason,
		}, nil

	default:
		// Should never happen - log for debugging
		e.logger.Error("Unknown China Wall status",
			zap.String("status", enforceResult.Status),
			zap.String("reason", enforceResult.Reason))
		return &DecisionResult{
			Mode:    ModeREFUSE,
			Content: fmt.Sprintf("Unexpected status: %s", enforceResult.Status),
		}, nil
	}
}

// gate1 checks if we have enough information to answer.
// Returns (result, needsLLM):
//   - problem-solving domain  -> (nil, false)  // Gate1Skip, no LLM
//   - already specific/long   -> (nil, false)  // no LLM
//   - charter match           -> (ASK, false)  // keyword path settled
//   - vague+short+no charter  -> (nil, true)   // inconclusive -> LLM
// WHY the bool (A13): previously every nil defeated Gate1Skip because
// Process always fell through to gate1WithLLM.
func (e *Engine) gate1(question string, expert Expert) (*DecisionResult, bool) {
	questionLower := strings.ToLower(question)

	// Domain bypass: problem-solving domain experts (DSA, coding, etc.)
	// never need clarification — every technical question is specific enough.
	// WHY domain not question: the expert's domain determines its behavior.
	// A DSA expert answers "what is python" directly. A medical expert might
	// need clarification on "what is python" (snake? programming language?).
	if chinawall.IsProblemSolvingDomain(expert.Domain) {
		return nil, false
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

	// Specific/long questions skip the LLM fallback.
	// WHY length check: Long questions usually have enough context.
	if !isVague || len(question) >= 80 {
		return nil, false
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
		}, false
	}

	// Keyword path flagged vague + short but found no charter questions.
	// Fall back to the cheap LLM check.
	return nil, true
}

// gate1WithLLM is the enhanced version using LLM for vagueness detection.
// Called only when gate1's keyword check is inconclusive (needsLLM=true).
// Uses zero-shot structured output (Byte by Byte AI course pattern).
// Cheap: ModelCheap, Temperature 0.1, single-pass, no CoT (§3.1 P11).
// Fail-safe: any LLM/parse error returns nil (do not block, §3.1 P3).
func (e *Engine) gate1WithLLM(ctx context.Context, question string, expert Expert) *DecisionResult {
	// Defense-in-depth: never ask clarification of problem-solving domains
	// even if a future call site forgets the needsLLM gate (A13).
	if chinawall.IsProblemSolvingDomain(expert.Domain) {
		return nil
	}

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

// gateStructurePermission implements CT-C4 / CATEGORY_TEMPLATE_HANDOFF.md §6.
// Returns (result, rewrittenQuestion):
//   - (nil, "") when this gate does not apply — the overwhelmingly common
//     case (expert.AskStructurePermission is false). Caller (Process)
//     continues with the original question exactly as before this
//     feature existed.
//   - (askResult, "") when a FRESH (non-reply) question arrives for an
//     expert whose category has ask_structure_permission=true — emits
//     an ASK-mode message asking the client's structure/boilerplate
//     preference, and Process returns immediately without reaching
//     Gate 1-5. Reuses the existing reply mechanism (CT-L9): the
//     client's answer will arrive as a normal message with
//     reply_to_message_id pointing at this ASK message's id, which is
//     how the second branch below recognizes it.
//   - (nil, rewrittenQuestion) when the incoming message IS a reply to
//     a structure-permission ASK this function itself previously
//     emitted — walks one level up via replyToMessageID to recover the
//     ORIGINAL question (not the client's short preference answer),
//     appends the client's stated preference to it, and returns the
//     combined text for Process to use for Gate 1 onward. This is how
//     "reuses the reply mechanism entirely, no separate flow" (CT-L9)
//     is satisfied: no new message type, no new DB column beyond the
//     one reply_to_message_id migration 010 already added.
//
// Mental execution (happy path, full round trip):
// 1. Client asks categorized-with-ask_structure_permission expert:
//    "reverse a linked list" (replyToMessageID=nil, fresh question)
//    -> isReplyToOwnAsk=false (no replyToMessageID to check)
//    -> emits ASK message "Chahiye structure/boilerplate ya sirf logic
//       likh doon?", Process returns immediately, GateStopped=0 (CT-C4
//       uses its own sentinel, not colliding with Gate 1's ASK — see
//       gateStopped: -1 below).
// 2. Client replies to THAT ASK message with "sirf logic"
//    (replyToMessageID = the ASK message's id)
//    -> isReplyToOwnAsk=true (parent row's decision_mode='ASK' AND
//       parent.expert_id == this expert AND parent has no
//       reply_to_message_id of its own, i.e. it's a root-level ASK,
//       not itself a reply — distinguishing a structure-permission ASK
//       from an unrelated ASK, e.g. Gate 1's clarification ASK, which
//       also sets decision_mode='ASK'. Gate 1 ASKs are also technically
//       valid parents for a generic reply, so this check alone is a
//       heuristic, not a perfect discriminator — documented as a known
//       limitation, not silently assumed correct)
//    -> walks up ONE more level (parent.reply_to_message_id) to find
//       the ORIGINAL question ("reverse a linked list")
//    -> returns rewrittenQuestion = "reverse a linked list\n\n(Client's
//       stated preference: sirf logic)"
//    -> Process continues Gate 1 onward with this rewritten question.
//
// Edge case: client replies to the ASK but the original question's
// parent lookup fails (deleted, or ASK was somehow a root message with
// no reply_to_message_id of its own — should not normally happen since
// this gate always sets it, but defensive nonetheless) -> falls back to
// using the client's raw reply text as the question, logs a warning,
// does not crash or return an error to the client.
func (e *Engine) gateStructurePermission(
	ctx context.Context,
	question string,
	expert Expert,
	replyToMessageID *uuid.UUID,
) (*DecisionResult, string) {
	if !expert.AskStructurePermission {
		return nil, ""
	}

	if replyToMessageID == nil {
		// Fresh question to a structure-permission category — ask first.
		// GateStopped=-1 (not 0/1/2/3/4/5): distinguishes this from every
		// existing gate stop reason so a frontend/log consumer can tell
		// "this is the structure-permission prompt" apart from a genuine
		// Gate 1 clarification ASK, without needing a new Mode value.
		return &DecisionResult{
			Mode:        ModeASK,
			Questions:   []string{"Chahiye structure/boilerplate ya sirf logic likh doon?"},
			GateStopped: -1,
			Content:     "Before I answer, let me know your preference.",
		}, ""
	}

	// This message IS a reply. Check whether it's specifically a reply
	// to THIS gate's own ASK (not some other reply chain).
	var parentMode, parentContent string
	var parentExpertID *uuid.UUID
	var parentReplyTo *uuid.UUID
	err := e.db.QueryRow(ctx,
		`SELECT COALESCE(decision_mode,''), content, expert_id, reply_to_message_id
		 FROM messages WHERE id=$1`,
		*replyToMessageID,
	).Scan(&parentMode, &parentContent, &parentExpertID, &parentReplyTo)
	if err != nil {
		e.logger.Warn("gateStructurePermission: parent message lookup failed, treating as fresh question",
			zap.Error(err),
		)
		return nil, ""
	}

	isOwnStructureAsk := parentMode == string(ModeASK) &&
		parentExpertID != nil && *parentExpertID == expert.ID &&
		parentContent == "Before I answer, let me know your preference."
	if !isOwnStructureAsk {
		// Reply to something else (e.g. a normal answer, or a Gate 1
		// clarification ASK) — not this gate's concern.
		return nil, ""
	}

	if parentReplyTo == nil {
		// Defensive: the ASK should always have been a reply to the
		// original question (see the fresh-question branch above, which
		// does NOT currently set reply_to_message_id on the ASK itself —
		// see known gap noted in CATEGORY_TEMPLATE_HANDOFF.md). Fall back
		// to the client's raw preference text rather than erroring.
		e.logger.Warn("gateStructurePermission: ASK message has no parent question, using reply text as question")
		return nil, question
	}

	var originalQuestion string
	err = e.db.QueryRow(ctx,
		`SELECT content FROM messages WHERE id=$1`,
		*parentReplyTo,
	).Scan(&originalQuestion)
	if err != nil || originalQuestion == "" {
		e.logger.Warn("gateStructurePermission: original question lookup failed, using reply text as question",
			zap.Error(err),
		)
		return nil, question
	}

	return nil, fmt.Sprintf("%s\n\n(Client's stated preference: %s)", originalQuestion, question)
}

// neverRulePattern compiled once at package level for performance.
var neverRulePattern = regexp.MustCompile(`never\s+(\w+(?:\s+\w+){0,5})`)

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}


