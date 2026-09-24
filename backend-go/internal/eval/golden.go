// Package eval implements C3: a golden-set evaluation harness with a
// deterministic scorer, baseline storage, and a vital-failure merge gate.
//
// §3.1 P7: a golden set (user-like + adversarial) is run on every
// PR/model/prompt change; baseline scores are stored; VITAL failures
// block (tiered must-have vs good-to-have); each component is scored
// separately before chaining. The scorer is deterministic (keyword /
// citation / refusal checks) so CI is reproducible; an LLM judge can be
// layered on later without changing the gate.
package eval

import (
	"bufio"
	"embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

//go:embed golden/*.jsonl
var goldenFS embed.FS

// CaseType distinguishes realistic user queries from adversarial probes.
type CaseType string

const (
	TypeUserLike    CaseType = "user_like"
	TypeAdversarial CaseType = "adversarial"
)

// Expectation is the machine-checkable expectation for a case. Each field
// is a separate component check (component-wise eval, P7).
type Expectation struct {
	// Keywords must all appear (case-insensitive substring) in the answer.
	Keywords []string `json:"keywords"`
	// MustCite requires at least one citation.
	MustCite bool `json:"must_cite"`
	// Refuse requires the answer to be a refusal (adversarial/out-of-scope).
	Refuse bool `json:"refuse"`
	// MinChars requires at least this many characters of content.
	MinChars int `json:"min_chars"`
}

// Case is one golden-set entry.
type Case struct {
	ID         string      `json:"id"`
	Suite      string      `json:"suite"`
	Type       CaseType    `json:"type"`
	ExpertSlug string      `json:"expert_slug"`
	Domain     string      `json:"domain"`
	Input      string      `json:"input"`
	Expect     Expectation `json:"expect"`
	// Vital marks a must-have case: any vital failure blocks the merge.
	Vital bool `json:"vital"`
}

// GoldenSet is a named, ordered collection of cases.
type GoldenSet struct {
	Name  string
	Cases []Case
}

// LoadGoldenSet loads an embedded golden set by name (file golden/<name>.jsonl).
func LoadGoldenSet(name string) (*GoldenSet, error) {
	path := "golden/" + name + ".jsonl"
	data, err := goldenFS.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("eval: golden set %q not found: %w", name, err)
	}
	cases, err := parseJSONL(string(data))
	if err != nil {
		return nil, fmt.Errorf("eval: parse %q: %w", name, err)
	}
	if len(cases) == 0 {
		return nil, fmt.Errorf("eval: golden set %q is empty", name)
	}
	for i := range cases {
		if cases[i].Suite == "" {
			cases[i].Suite = name
		}
	}
	return &GoldenSet{Name: name, Cases: cases}, nil
}

// GoldenSetNames lists the embedded golden set names, sorted.
func GoldenSetNames() ([]string, error) {
	entries, err := goldenFS.ReadDir("golden")
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		n := strings.TrimSuffix(e.Name(), ".jsonl")
		if n != e.Name() {
			names = append(names, n)
		}
	}
	sort.Strings(names)
	return names, nil
}

// parseJSONL parses one JSON object per non-empty, non-comment (#) line.
// Pure — unit-tested.
func parseJSONL(text string) ([]Case, error) {
	var cases []Case
	sc := bufio.NewScanner(strings.NewReader(text))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	line := 0
	for sc.Scan() {
		line++
		raw := strings.TrimSpace(sc.Text())
		if raw == "" || strings.HasPrefix(raw, "#") {
			continue
		}
		var c Case
		if err := json.Unmarshal([]byte(raw), &c); err != nil {
			return nil, fmt.Errorf("line %d: %w", line, err)
		}
		if c.ID == "" {
			return nil, fmt.Errorf("line %d: case missing id", line)
		}
		if c.Input == "" {
			return nil, fmt.Errorf("line %d (id=%s): missing input", line, c.ID)
		}
		cases = append(cases, c)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return cases, nil
}
