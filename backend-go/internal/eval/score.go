package eval

import (
	"fmt"
	"strings"
)

// Check is one component-wise evaluation result.
type Check struct {
	Name   string `json:"name"`
	Passed bool   `json:"passed"`
	Detail string `json:"detail,omitempty"`
}

// Verdict is the scored result for one case.
type Verdict struct {
	CaseID string  `json:"case_id"`
	Checks []Check `json:"checks"`
	Passed bool    `json:"passed"`
}

// Score runs every expectation component against the observed answer.
// Deterministic and pure — unit-tested. A case passes only when ALL its
// component checks pass (no partial credit; vital failure is counted by
// the runner, not here).
func Score(c Case, o Observed) Verdict {
	var checks []Check

	// Always: answer must be non-empty (min_chars, default 1).
	minChars := c.Expect.MinChars
	if minChars <= 0 {
		minChars = 1
	}
	contentLen := len(strings.TrimSpace(o.Content))
	checks = append(checks, Check{
		Name:   "non_empty",
		Passed: contentLen >= minChars,
		Detail: detailf("len=%d min=%d", contentLen, minChars),
	})

	// Refusal expectation. A refusal is either an explicit mode OR an
	// answer that carries no citations and no material content (the
	// China Wall refusal path returns a short reason string).
	refused := o.Refused || o.Mode == "REFUSE"
	checks = append(checks, Check{
		Name:   "refusal",
		Passed: c.Expect.Refuse == refused,
		Detail: detailf("expected_refuse=%v got=%v mode=%s", c.Expect.Refuse, refused, o.Mode),
	})

	// Citation requirement (skipped entirely when not expected).
	if c.Expect.MustCite {
		checks = append(checks, Check{
			Name:   "citation",
			Passed: o.Citations > 0,
			Detail: detailf("citations=%d", o.Citations),
		})
	}

	// Keyword coverage (each expected keyword must appear).
	if len(c.Expect.Keywords) > 0 {
		lower := strings.ToLower(o.Content)
		var missing []string
		for _, kw := range c.Expect.Keywords {
			if kw == "" {
				continue
			}
			if !strings.Contains(lower, strings.ToLower(kw)) {
				missing = append(missing, kw)
			}
		}
		checks = append(checks, Check{
			Name:   "keywords",
			Passed: len(missing) == 0,
			Detail: detailf("missing=%v", missing),
		})
	}

	passed := true
	for _, ch := range checks {
		if !ch.Passed {
			passed = false
			break
		}
	}
	return Verdict{CaseID: c.ID, Checks: checks, Passed: passed}
}

func detailf(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}
