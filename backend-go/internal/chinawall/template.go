package chinawall

import (
	"encoding/json"
	"fmt"
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
func buildStructuredPrompt(
	expertName string,
	reasoningCharter string,
	contextText string,
	sections []category.TemplateSection,
	defaultLanguage string,
) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("You are %s, a domain expert.\n\nREASONING CHARTER:\n%s\n\n", expertName, reasoningCharter))
	sb.WriteString("YOUR TRAINING MATERIAL (use these principles to solve problems):\n")
	sb.WriteString(contextText)
	sb.WriteString("\n\nCRITICAL RULES:\n")
	sb.WriteString("1. Respond with STRICT JSON ONLY — no markdown, no prose outside the JSON object.\n")
	sb.WriteString(fmt.Sprintf("2. The JSON object must have EXACTLY these keys: %s\n", sectionKeysList(sections)))
	sb.WriteString("3. For every claim in a prose-type section's text, cite the source using [CHUNK_uuid] format.\n")
	sb.WriteString(fmt.Sprintf("4. For any code-type section, the value must be a plain JSON string containing complete, working %s code (escape newlines as \\n). NO [CHUNK_xxx] tokens inside the code.\n", defaultLanguage))
	sb.WriteString(fmt.Sprintf("5. For any test_cases-type section, the value must be a JSON object with EXACTLY these bucket keys: %s — each mapping to an array of test case strings (empty array if none apply).\n", strings.Join(TestCaseBuckets, ", ")))
	sb.WriteString("6. If a technique is NOT in your training material, say so explicitly inside the relevant section's text — do not omit the key.\n")
	sb.WriteString("\nSection descriptions:\n")
	for _, s := range sections {
		sb.WriteString(fmt.Sprintf("- %q (type=%s, label=%q)\n", s.Key, s.Type, s.Label))
	}
	sb.WriteString("\nReturn ONLY the JSON object, nothing else.")
	return sb.String()
}

func sectionKeysList(sections []category.TemplateSection) string {
	keys := make([]string, len(sections))
	for i, s := range sections {
		keys[i] = s.Key
	}
	return strings.Join(keys, ", ")
}

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
	clean = strings.TrimPrefix(clean, "```json")
	clean = strings.TrimPrefix(clean, "```")
	clean = strings.TrimSuffix(clean, "```")
	clean = strings.TrimSpace(clean)

	start := strings.Index(clean, "{")
	end := strings.LastIndex(clean, "}")
	if start == -1 || end == -1 || start >= end {
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
