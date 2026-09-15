package chinawall

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"ai_avengers/backend/internal/category"
)

// Hardcoded test-case buckets (CT-L4, CATEGORY_TEMPLATE_HANDOFF.md §2).
// Client's explicit choice: these labels are NEVER admin-configurable,
// unlike every other part of a category's template_schema. Order is
// significant — buckets always render in this fixed order regardless of
// what order the model returns them in.
const (
	TestCaseBucketBase   = "BASE"
	TestCaseBucketEdge   = "EDGE"
	TestCaseBucketCorner = "CORNER"
	TestCaseBucketStress = "STRESS"
)

// TestCaseBuckets is the fixed, ordered list of buckets every test_cases
// section renders, regardless of category or expert.
var TestCaseBuckets = []string{
	TestCaseBucketBase, TestCaseBucketEdge, TestCaseBucketCorner, TestCaseBucketStress,
}

// TemplateSectionResult is one generated+processed section of a structured
// answer. Carried ALONGSIDE EnforceResult.Answer/Citations, never instead
// of — flat-text-only consumers (any expert with no category, CT-L2) see
// zero change: TemplateSections is simply nil/empty for them.
type TemplateSectionResult struct {
	Key       string               `json:"key"`
	Label     string               `json:"label"`
	Type      category.SectionType `json:"type"`
	Content   string               `json:"content"`
	Citations []Citation           `json:"citations"`
}

// buildStructuredPrompt builds the Layer 3 system prompt instructing the
// LLM to return STRICT JSON with exactly the category's section keys.
//
// WHY JSON not markdown headings (CT-L3): reliability over flexibility —
// explicit client choice over the more flexible but format-fragile
// markdown-heading convention alternative.
//
// profile parameter (2026-09-08 RCA fix): previously this function had
// NO access to the expert's DomainProfile at all, unlike generateFlatText
// (enforcer.go), which uses profile.CitationMode and profile.SystemPromptExt
// to shape its prompt. Root-cause of a real production symptom: a LOOSE-
// citation domain (e.g. DSA — principles transfer to new problems, code is
// exempt from per-line citations) was forced through a hardcoded
// STRICT-only citation rule here ("every claim must cite"), and its
// SystemPromptExt ("Always include time/space complexity, complete
// runnable code") never reached the model at all. Result: the model,
// facing a strict citation demand it could not satisfy with LOOSE-style
// reasoning, produced thin, citation-only prose sections instead of full
// explanations — exactly the reported "domain expert only gives sources"
// symptom. Fixed by branching on profile.CitationMode (mirrors
// generateFlatText's own branch) and appending profile.SystemPromptExt
// as an additional numbered rule.
func buildStructuredPrompt(
	expertName string,
	reasoningCharter string,
	contextText string,
	sections []category.TemplateSection,
	defaultLanguage string,
	profile *DomainProfile,
) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("You are %s, a domain expert.\n\nREASONING CHARTER:\n%s\n\n", expertName, reasoningCharter))
	sb.WriteString("YOUR TRAINING MATERIAL (use these principles to solve problems):\n")
	sb.WriteString(contextText)
	sb.WriteString("\n\nCRITICAL RULES:\n")
	sb.WriteString("1. Respond with STRICT JSON ONLY — no markdown, no prose outside the JSON object.\n")
	sb.WriteString(fmt.Sprintf("2. The JSON object must have EXACTLY these keys: %s\n", sectionKeysList(sections)))
	if profile.CitationMode == CitationModeLoose {
		// LOOSE: mirrors generateFlatText's loose-mode system prompt.
		// WHY: forcing a strict per-claim citation rule on a domain whose
		// whole point is applying principles to NEW problems (the exact
		// problem is often not verbatim in the training material) leaves
		// the model unable to satisfy the rule honestly — it then falls
		// back to citing without explaining, rather than explaining fully.
		sb.WriteString("3. In prose-type sections, cite the training-material principles you are APPLYING using [CHUNK_uuid] format where relevant — but explain your full reasoning first; a citation supports the explanation, it does not replace it. NEVER answer with citations alone and no real explanation.\n")
	} else {
		sb.WriteString("3. Every factual claim in a prose-type section's text MUST cite a source using [CHUNK_uuid] format. If information is not in your training material, say so explicitly inside that section rather than omitting it.\n")
	}
	sb.WriteString(fmt.Sprintf("4. For any code-type section, the value must be a plain JSON string containing complete, working %s code (escape newlines as \\n). NO [CHUNK_xxx] tokens inside the code.\n", defaultLanguage))
	sb.WriteString(fmt.Sprintf("5. For any test_cases-type section, the value must be a JSON object with EXACTLY these bucket keys: %s — each mapping to an array of test case strings (empty array if none apply).\n", strings.Join(TestCaseBuckets, ", ")))
	sb.WriteString("6. If a technique is NOT in your training material, say so explicitly inside the relevant section's text — do not omit the key.\n")
	sb.WriteString("7. Every prose-type section must contain SUBSTANTIVE content — multiple full sentences of real explanation, not a citation-only stub or a one-line placeholder.\n")
	if profile.SystemPromptExt != "" {
		// RCA fix (2026-09-08): previously never reached this prompt at
		// all — only generateFlatText applied SystemPromptExt. A DSA
		// domain's "Always include time and space complexity (Big-O).
		// Provide complete, runnable code." had zero effect on any
		// categorized (structured-JSON) expert until now.
		sb.WriteString(fmt.Sprintf("8. %s\n", profile.SystemPromptExt))
	}
	sb.WriteString("\nSection descriptions (what to actually write in each section):\n")
	for _, s := range sections {
		sb.WriteString(fmt.Sprintf("- %q (type=%s, label=%q): %s\n", s.Key, s.Type, s.Label, sectionGuidance(s)))
	}
	sb.WriteString("\nReturn ONLY the JSON object, nothing else.")
	return sb.String()
}

// sectionGuidance returns human-readable guidance on WHAT CONTENT belongs
// in a section — distinct from Label (a display name, e.g. "Pattern"),
// which by itself told the model nothing about what to actually write
// (2026-09-08 RCA — see TemplateSection.Description's doc comment for
// the full incident this fixes).
//
// Priority:
//  1. s.Description, if the admin set one — always wins, any category.
//  2. Built-in guidance for code/test_cases types (their format is
//     already fully specified by rules 4/5 above; this just labels
//     what belongs there content-wise).
//  3. Built-in guidance keyed by common key/label patterns (pattern,
//     idea, walkthrough, complexity) — covers the seeded "coding"
//     category (migration 010) and any admin-created category reusing
//     the same conventional names, with ZERO admin action required.
//  4. Generic fallback for anything else — still explicitly instructs
//     "not thin/citation-only", rather than saying nothing.
func sectionGuidance(s category.TemplateSection) string {
	if s.Description != "" {
		return s.Description
	}
	if s.Type == category.SectionTypeCode {
		return "Complete, working, runnable code implementing the solution (see rule 4 for format)."
	}
	if s.Type == category.SectionTypeTestCases {
		return "Concrete test cases covering base/edge/corner/stress scenarios (see rule 5 for format)."
	}
	key := strings.ToLower(strings.TrimSpace(s.Key))
	label := strings.ToLower(strings.TrimSpace(s.Label))
	switch {
	case key == "pattern" || strings.Contains(label, "pattern"):
		return "Identify and NAME the core algorithmic pattern/technique this problem needs (e.g. two pointers, sliding window, sorting + greedy, binary search, dynamic programming). Briefly explain WHY that pattern applies here."
	case key == "idea" || strings.Contains(label, "idea") || strings.Contains(label, "approach"):
		return "Explain the core idea/approach in plain language — the key insight that makes the solution work — before any code details."
	case key == "walkthrough" || strings.Contains(label, "walkthrough") || strings.Contains(label, "trace") || strings.Contains(label, "example"):
		return "Trace through a concrete example input step-by-step, showing exactly how the algorithm executes and arrives at the final output."
	case strings.Contains(label, "complex"):
		return "State the time and space complexity (Big-O) with a brief justification."
	default:
		return "Write clear, complete, substantive content for this section — do not leave it thin or citation-only."
	}
}

// findMatchingBrace finds the index of the closing } that matches the
// opening { at position start in s. Returns -1 if not found.
// Handles nested braces correctly. Ignores braces inside JSON strings
// (quoted with ") to avoid false matches on string values containing {}.
func findMatchingBrace(s string, start int) int {
	depth := 0
	inString := false
	escaped := false
	for i := start; i < len(s); i++ {
		ch := s[i]
		if escaped {
			escaped = false
			continue
		}
		if ch == '\\' && inString {
			escaped = true
			continue
		}
		if ch == '"' {
			inString = !inString
			continue
		}
		if inString {
			continue
		}
		if ch == '{' {
			depth++
		} else if ch == '}' {
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

func sectionKeysList(sections []category.TemplateSection) string {
	keys := make([]string, len(sections))
	for i, s := range sections {
		keys[i] = s.Key
	}
	return strings.Join(keys, ", ")
}

// thinkBlockRe strips <think>...</think> reasoning blocks that DeepSeek
// and other reasoning models emit before their actual JSON response.
// These blocks often contain code examples with {} braces, which break
// the naive strings.LastIndex("}") JSON extraction below.
// (?s) flag makes . match newlines (multi-line think blocks).
var thinkBlockRe = regexp.MustCompile(`(?s)<think>.*?</think>`)

// parseStructuredResponse parses the LLM's JSON response into one raw
// string per section key, keyed by section.Key. Prose/code sections are
// expected as plain JSON strings; test_cases sections are expected as a
// JSON object of bucket->[]string, which is re-serialized into a
// deterministic, bucket-ordered display string (via renderTestCaseBuckets)
// so every section can be treated as plain text uniformly by the caller
// (citation extraction + per-section stripUncited).
//
// Mental execution:
// raw = `{"pattern": "Sliding window [CHUNK_abc]", "code": "public int f(){}",
//         "test_cases": {"BASE": ["n=5"], "EDGE": [], "CORNER": [], "STRESS": ["n=1e6"]}}`
// sections = [pattern(prose), code(code), test_cases(test_cases)]
// -> result["pattern"]    = "Sliding window [CHUNK_abc]"
// -> result["code"]       = "public int f(){}"
// -> result["test_cases"] = "BASE:\n- n=5\nEDGE:\n(none)\nCORNER:\n(none)\nSTRESS:\n- n=1e6\n"
//
// Edge case: a key missing entirely from the model's JSON -> result[key]=""
// (caller decides how to treat an empty required section, not this func).
// Edge case: model returns a nested object where a plain string was
// expected -> falls back to the raw JSON text rather than silently
// dropping the section (no empty catch / silent error, per Step 5 checklist).
func parseStructuredResponse(raw string, sections []category.TemplateSection) (map[string]string, error) {
	clean := strings.TrimSpace(raw)

	// Strip <think>...</think> reasoning blocks FIRST.
	// WHY: DeepSeek and other reasoning models emit a thinking block
	// before the actual JSON. These blocks often contain code examples
	// with {} braces. strings.LastIndex("}") below would find the last
	// brace INSIDE the thinking block, not the JSON's closing brace,
	// causing json.Unmarshal to fail on every structured response.
	clean = thinkBlockRe.ReplaceAllString(clean, "")
	clean = strings.TrimSpace(clean)

	clean = strings.TrimPrefix(clean, "```json")
	clean = strings.TrimPrefix(clean, "```")
	clean = strings.TrimSuffix(clean, "```")
	clean = strings.TrimSpace(clean)

	// Find the JSON object using balanced brace matching instead of
	// naive strings.LastIndex("}").
	// WHY: LastIndex finds the LAST } in the string. If the model
	// included any trailing text or a partial second object after the
	// real JSON, LastIndex would include that garbage. Balanced matching
	// finds the FIRST complete, balanced JSON object — exactly what we want.
	start := strings.Index(clean, "{")
	if start == -1 {
		return nil, fmt.Errorf("no JSON object found in structured response")
	}
	end := findMatchingBrace(clean, start)
	if end == -1 {
		return nil, fmt.Errorf("no JSON object found in structured response")
	}
	clean = clean[start : end+1]

	var rawMap map[string]json.RawMessage
	if err := json.Unmarshal([]byte(clean), &rawMap); err != nil {
		return nil, fmt.Errorf("structured response JSON parse failed: %w", err)
	}

	result := make(map[string]string, len(sections))
	for _, s := range sections {
		val, ok := rawMap[s.Key]
		if !ok {
			result[s.Key] = ""
			continue
		}
		if s.Type == category.SectionTypeTestCases {
			text, err := renderTestCaseBuckets(val)
			if err != nil {
				result[s.Key] = ""
				continue
			}
			result[s.Key] = text
			continue
		}
		var text string
		if err := json.Unmarshal(val, &text); err != nil {
			text = string(val)
		}
		result[s.Key] = text
	}
	return result, nil
}

// renderTestCaseBuckets turns the model's {"BASE": [...], "EDGE": [...], ...}
// object into a fixed-order, human-readable string. Buckets always render
// in TestCaseBuckets order regardless of the key order the model returned
// (CT-L4 — hardcoded ordering, never model/admin-controlled).
func renderTestCaseBuckets(val json.RawMessage) (string, error) {
	var buckets map[string][]string
	if err := json.Unmarshal(val, &buckets); err != nil {
		return "", err
	}
	var sb strings.Builder
	for _, bucket := range TestCaseBuckets {
		cases := buckets[bucket]
		sb.WriteString(bucket + ":\n")
		if len(cases) == 0 {
			sb.WriteString("(none)\n")
			continue
		}
		for _, c := range cases {
			sb.WriteString("- " + c + "\n")
		}
	}
	return sb.String(), nil
}
