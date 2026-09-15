package chinawall

// Default MaxTokens values used when a DomainProfile does not specify
// its own override (MaxTokensFlat/MaxTokensStructured == 0). Named
// constants, not magic numbers scattered in enforcer.go, so admin
// panel guidance and code stay in sync at one source of truth.
//
// WHY two separate defaults: a flat prose+code answer fits comfortably
// in far fewer tokens than a structured JSON answer, which must pack
// prose + a full code block + 4 test-case buckets into ONE JSON object
// (see CATEGORY_TEMPLATE_HANDOFF.md §4 and the 2026-09-08 RCA in
// HANDOFF.md documenting a real truncation bug this exact setting
// caused for a DSA expert).
const (
	DefaultMaxTokensFlat       = 5000
	DefaultMaxTokensStructured = 12000
)

// CoverageMode controls how Layer 2 checks if chunks can answer the question.
type CoverageMode string

const (
	// CoverageModeApplyPrinciples: checks if expert can APPLY principles to solve.
	// Correct for DSA, algorithms, coding — principles transfer to new problems.
	// WHY: A DSA expert trained on "two pointers" SHOULD solve new sliding window
	// problems. Asking "is this exact problem in transcript?" would refuse everything.
	CoverageModeApplyPrinciples CoverageMode = "APPLY_PRINCIPLES"

	// CoverageModeLiteralMatch: checks if chunks CONTAIN the answer.
	// Correct for medical, legal, finance — facts must be in the transcript.
	// WHY: A medical expert must not apply general knowledge. Only transcript facts.
	CoverageModeLiteralMatch CoverageMode = "LITERAL_MATCH"
)

// CitationMode controls how Layer 3 generates the answer.
type CitationMode string

const (
	// CitationModeLoose: cite principles in explanation, code blocks exempt.
	// Correct for DSA — code IS the application of cited principles.
	CitationModeLoose CitationMode = "LOOSE"

	// CitationModeStrict: every factual claim must have inline citation.
	// Correct for medical, legal, finance — no uncited claims allowed.
	CitationModeStrict CitationMode = "STRICT"
)

// StripMode controls how Layer 4 strips uncited content.
type StripMode string

const (
	// StripModeCodeExempt: fenced code blocks kept unconditionally.
	// Correct for DSA — code does not need per-line citations.
	StripModeCodeExempt StripMode = "CODE_EXEMPT"

	// StripModeFull: every sentence without citation is stripped.
	// Correct for medical, legal, finance — strict grounding required.
	StripModeFull StripMode = "FULL_STRIP"
)

// DomainProfile defines how the China Wall behaves for a specific domain.
//
// INHERITANCE MODEL:
//   BaseProfile (immutable defaults) ← domain profile overrides specific fields
//
// PRIORITY MODEL:
//   Domain rules apply first.
//   IF Layer 4 output is empty AND StripMode != CODE_EXEMPT:
//     → retry with BaseProfile (safety net kicks in)
//   IF StripMode == CODE_EXEMPT AND output has code block:
//     → valid output, domain rules win
//
// WHY this conflict resolution:
//   Empty output = domain rules over-relaxed the wall.
//   Code block present = DSA answer is valid even without prose citations.
//   No LLM judgment needed — purely structural check.
type DomainProfile struct {
	// Domain identifier — matches Expert.Domain field (case-insensitive).
	Domain string

	// Gate1Skip: if true, skip vagueness check entirely for this domain.
	// WHY: DSA experts answer every technical question directly.
	// A medical expert may need clarification on ambiguous symptoms.
	Gate1Skip bool

	// CoverageMode: how Layer 2 checks chunk coverage.
	CoverageMode CoverageMode

	// CitationMode: how Layer 3 generates the answer.
	CitationMode CitationMode

	// StripMode: how Layer 4 strips uncited content.
	StripMode StripMode

	// SystemPromptExt: appended to Layer 3 system prompt.
	// Domain-specific instructions for the LLM.
	// Example for DSA: "Always include time and space complexity."
	// Example for Medical: "Always recommend consulting a doctor."
	SystemPromptExt string

	// DomainKeywords: topics this domain covers.
	// Used by Gate 2 to verify question is in domain scope.
	// Empty = Gate 2 does not filter by keyword for this domain.
	DomainKeywords []string

	// CustomRules: domain-specific rules evaluated before base rules.
	// Updated by AI based on conversation patterns.
	// Stored in DB as JSONB, loaded at startup, cached in memory.
	CustomRules []DomainRule

	// MaxTokensFlat: LLM response token cap for this domain's flat-text
	// generation (chinawall.Enforcer.generateFlatText). 0 means "use
	// DefaultMaxTokensFlat" - this is what every domain has today, so
	// admins upgrading from an older DB row (no this field yet) keep
	// identical behavior with zero migration required.
	//
	// WHY admin-configurable (2026-09-08, real production incident):
	// a hardcoded token limit caused a real DSA expert's structured
	// answers to get truncated mid-JSON, producing garbled output (see
	// HANDOFF.md's 2026-09-08 round-3 RCA). A fixed code-level constant
	// meant fixing this for one domain risked being wrong for the next
	// domain with different answer-length needs - now every domain
	// admin can tune this independently from the admin panel, without a
	// backend redeploy, exactly like every other field on this struct.
	MaxTokensFlat int

	// MaxTokensStructured: same as MaxTokensFlat, but for this domain's
	// STRUCTURED (categorized-expert) generation path
	// (chinawall.Enforcer.generateStructured). 0 means "use
	// DefaultMaxTokensStructured". Separate from MaxTokensFlat because a
	// structured answer must fit prose + a full code block + test-case
	// buckets all inside ONE JSON object - routinely needs a much higher
	// cap than a flat answer for the SAME domain.
	MaxTokensStructured int
}

// DomainRule is a single domain-specific rule.
// AI can add/update rules based on conversation patterns.
type DomainRule struct {
	// ID: unique identifier for this rule.
	ID string

	// Description: human-readable explanation of the rule.
	Description string

	// Condition: when this rule applies.
	// Example: "question contains 'complexity'"
	Condition string

	// Action: what to do when condition is met.
	// Example: "always include Big-O analysis"
	Action string
}

// BaseProfile is the immutable fallback used when:
//   1. No domain profile exists for the expert's domain.
//   2. Domain profile produced empty output (safety net).
//
// WHY strict defaults:
//   Unknown domain = treat as factual domain.
//   Better to refuse than to hallucinate.
//   Admin can always create a domain profile to relax rules.
var BaseProfile = &DomainProfile{
	Domain:       "__base__",
	Gate1Skip:    false,
	CoverageMode: CoverageModeLiteralMatch,
	CitationMode: CitationModeStrict,
	StripMode:    StripModeFull,
}

// DefaultProfiles are seeded at startup if not present in DB.
// Admin can override any field from the admin panel.
// AI can update CustomRules based on conversation patterns.
var DefaultProfiles = []*DomainProfile{
	{
		Domain:          "dsa",
		Gate1Skip:       true,
		CoverageMode:    CoverageModeApplyPrinciples,
		CitationMode:    CitationModeLoose,
		StripMode:       StripModeCodeExempt,
		SystemPromptExt: "Always include time complexity (Big-O) and space complexity. Provide complete, runnable code.",
		DomainKeywords:  []string{"algorithm", "data structure", "complexity", "array", "tree", "graph", "sort", "search", "dynamic programming", "recursion"},
	},
	{
		Domain:          "algorithms",
		Gate1Skip:       true,
		CoverageMode:    CoverageModeApplyPrinciples,
		CitationMode:    CitationModeLoose,
		StripMode:       StripModeCodeExempt,
		SystemPromptExt: "Always explain the algorithmic approach before code. Include complexity analysis.",
		DomainKeywords:  []string{"algorithm", "complexity", "optimization", "greedy", "divide and conquer"},
	},
	{
		Domain:          "system_design",
		Gate1Skip:       false, // System design questions can be vague — clarification helps
		CoverageMode:    CoverageModeApplyPrinciples,
		CitationMode:    CitationModeLoose,
		StripMode:       StripModeCodeExempt,
		SystemPromptExt: "Always discuss scalability, trade-offs, and failure modes. Use diagrams in text form where helpful.",
		DomainKeywords:  []string{"design", "architecture", "scale", "database", "cache", "load balancer", "microservice", "api"},
	},
	{
		Domain:          "coding",
		Gate1Skip:       true,
		CoverageMode:    CoverageModeApplyPrinciples,
		CitationMode:    CitationModeLoose,
		StripMode:       StripModeCodeExempt,
		SystemPromptExt: "Provide complete, working code. Explain the approach before the code.",
		DomainKeywords:  []string{"code", "function", "implement", "write", "program"},
	},
	{
		Domain:          "oops",
		Gate1Skip:       true,
		CoverageMode:    CoverageModeApplyPrinciples,
		CitationMode:    CitationModeLoose,
		StripMode:       StripModeCodeExempt,
		SystemPromptExt: "Always explain the OOP principle being applied. Show class diagrams in text form. Provide code examples.",
		DomainKeywords:  []string{"class", "object", "inheritance", "polymorphism", "encapsulation", "abstraction", "interface", "design pattern"},
	},
	{
		Domain:          "medical",
		Gate1Skip:       false,
		CoverageMode:    CoverageModeLiteralMatch,
		CitationMode:    CitationModeStrict,
		StripMode:       StripModeFull,
		SystemPromptExt: "Always recommend consulting a qualified medical professional. Only cite information from the provided course material.",
		DomainKeywords:  []string{},
	},
	{
		Domain:          "finance",
		Gate1Skip:       false,
		CoverageMode:    CoverageModeLiteralMatch,
		CitationMode:    CitationModeStrict,
		StripMode:       StripModeFull,
		SystemPromptExt: "Always include a disclaimer that this is educational content, not financial advice.",
		DomainKeywords:  []string{},
	},
	{
		Domain:          "legal",
		Gate1Skip:       false,
		CoverageMode:    CoverageModeLiteralMatch,
		CitationMode:    CitationModeStrict,
		StripMode:       StripModeFull,
		SystemPromptExt: "Always include a disclaimer that this is educational content, not legal advice. Recommend consulting a qualified lawyer.",
		DomainKeywords:  []string{},
	},
}
