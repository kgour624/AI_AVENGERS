package workflow

import "testing"

// This is the rule that decides whether work is accepted, so every branch is
// pinned. The property that matters most: a split is NEVER an approval.
func TestCombineCriticVerdicts(t *testing.T) {
	tests := []struct {
		name       string
		opinions   []CriticOpinion
		wantStatus string
	}{
		{
			name:       "both approve",
			opinions:   []CriticOpinion{{Critic: "sec", Status: "approved"}, {Critic: "sec", Status: "approved"}},
			wantStatus: "approved",
		},
		{
			name:       "one blocks, one approves — blocked",
			opinions:   []CriticOpinion{{Critic: "sec", Status: "blocked", Comment: "auth bypass"}, {Critic: "sec", Status: "approved"}},
			wantStatus: "blocked",
		},
		{
			name:       "one asks for changes, one approves — changes requested, not approval",
			opinions:   []CriticOpinion{{Critic: "lld", Status: "changes_requested", Comment: "no retry policy"}, {Critic: "lld", Status: "approved"}},
			wantStatus: "changes_requested",
		},
		{
			name:       "blocked outranks changes requested",
			opinions:   []CriticOpinion{{Critic: "a", Status: "changes_requested"}, {Critic: "b", Status: "blocked"}},
			wantStatus: "blocked",
		},
		{
			name:       "no opinion is not consent",
			opinions:   nil,
			wantStatus: "",
		},
		{
			name:       "an unreadable verdict cannot act as a yes",
			opinions:   []CriticOpinion{{Critic: "a", Status: "approved"}, {Critic: "b", Status: "wat"}},
			wantStatus: "changes_requested",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := combineCriticVerdicts(tc.opinions)
			if got != tc.wantStatus {
				t.Fatalf("combineCriticVerdicts = %q, want %q", got, tc.wantStatus)
			}
		})
	}
}

// A disagreement must be readable: the comment names the dissenting critic and its
// reason, because the doubt itself is the finding.
func TestCombineCriticVerdictsNamesTheDissent(t *testing.T) {
	_, comment := combineCriticVerdicts([]CriticOpinion{
		{Critic: "security", Model: "cheap", Status: "approved"},
		{Critic: "security", Model: "fast", Status: "changes_requested", Comment: "no rate limit"},
	})
	if comment == "" {
		t.Fatal("a split must come with a readable reason")
	}
	for _, want := range []string{"security", "fast", "no rate limit"} {
		if !contains(comment, want) {
			t.Fatalf("comment %q should name %q", comment, want)
		}
	}
}

// The second critic must not be the same tier as the first, or two samples from
// one model share its blind spots — the exact failure this policy exists to catch.
func TestDefaultSecondCriticIsADifferentTier(t *testing.T) {
	p := DefaultMultiCriticPolicy()
	if !p.Enabled {
		t.Fatal("the second critic is on by default; a silently disabled review is the thing nobody notices")
	}
	if p.SecondCritic == "" {
		t.Fatal("the second critic must have a tier")
	}
}

func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
