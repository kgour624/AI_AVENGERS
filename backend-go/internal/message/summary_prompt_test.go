package message

import (
	"strings"
	"testing"
)

func TestBuildRollingSummaryPrompt_Prescriptive(t *testing.T) {
	got := buildRollingSummaryPrompt([]string{"Turn 1 [db]: chose postgres", "Turn 2 [api]: p95 < 200ms"})
	// Must name what to keep (B9 P5) rather than a generic "summarize".
	for _, need := range []string{
		"decisions", "numbers", "constraints", "open questions", "commands",
		"Turn 1 [db]: chose postgres", "HISTORY:", "ROLLING SUMMARY:",
	} {
		if !strings.Contains(got, need) {
			t.Errorf("missing %q in prompt:\n%s", need, got)
		}
	}
}
