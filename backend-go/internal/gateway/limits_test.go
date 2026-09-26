package gateway

import "testing"

// The limit precedence decides which number a real LLM call is capped by, so it
// is pinned here rather than inferred from the query's ORDER BY.
func TestPickModelLimitPrecedence(t *testing.T) {
	rows := []ModelLimit{
		{Provider: "anthropic", Tier: "*", MaxInputTokens: 100, MaxOutputTokens: 100},
		{Provider: "anthropic", Tier: LimitTierStrong, MaxInputTokens: 200, MaxOutputTokens: 200},
		{Provider: "deepseek", Tier: "*", MaxInputTokens: 300, MaxOutputTokens: 300},
	}

	tests := []struct {
		name        string
		provider    string
		tier        string
		wantOK      bool
		wantOutput  int
	}{
		{name: "exact tier row wins over the wildcard", provider: "anthropic", tier: LimitTierStrong, wantOK: true, wantOutput: 200},
		{name: "wildcard covers a tier without its own row", provider: "anthropic", tier: LimitTierCheap, wantOK: true, wantOutput: 100},
		{name: "other provider's row is irrelevant", provider: "openrouter", tier: LimitTierStrong, wantOK: false},
		{name: "wildcard only, different provider", provider: "deepseek", tier: LimitTierFast, wantOK: true, wantOutput: 300},
		{name: "provider case is not significant", provider: "Anthropic", tier: LimitTierStrong, wantOK: true, wantOutput: 200},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := pickModelLimit(rows, tc.provider, tc.tier)
			if ok != tc.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tc.wantOK)
			}
			if ok && got.MaxOutputTokens != tc.wantOutput {
				t.Fatalf("MaxOutputTokens = %d, want %d", got.MaxOutputTokens, tc.wantOutput)
			}
		})
	}
}

func TestPickModelLimitEmpty(t *testing.T) {
	if _, ok := pickModelLimit(nil, "anthropic", LimitTierStrong); ok {
		t.Fatal("an empty table must mean 'nothing configured', not a zero limit")
	}
}

// effectiveOutputBudget is what stops a 2000-token planner call from dying on a
// reasoning model, and what stops a 32000-token admin value from being sent to a
// model that only accepts 8192.
func TestEffectiveOutputBudget(t *testing.T) {
	tests := []struct {
		name      string
		requested int
		ceiling   int
		want      int
	}{
		{name: "no opinion uses the ceiling", requested: 0, ceiling: 16000, want: 16000},
		{name: "small request is kept", requested: 200, ceiling: 8192, want: 200},
		{name: "request above the ceiling is narrowed", requested: 32000, ceiling: 8192, want: 8192},
		{name: "at the ceiling", requested: 8192, ceiling: 8192, want: 8192},
		{name: "no ceiling and no request keeps the historical 2000", requested: 0, ceiling: 0, want: 2000},
		{name: "no ceiling with an explicit request", requested: 900, ceiling: 0, want: 900},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := effectiveOutputBudget(tc.requested, tc.ceiling); got != tc.want {
				t.Fatalf("effectiveOutputBudget(%d, %d) = %d, want %d", tc.requested, tc.ceiling, got, tc.want)
			}
		})
	}
}

func TestLimitTierMapping(t *testing.T) {
	cases := map[ModelType]string{
		ModelStrong: LimitTierStrong,
		ModelFast:   LimitTierFast,
		ModelCheap:  LimitTierCheap,
		ModelType(""): LimitTierDefault,
	}
	for in, want := range cases {
		if got := limitTier(in); got != want {
			t.Fatalf("limitTier(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestEstimateTokensAndMessages(t *testing.T) {
	if got := estimateTokens(""); got != 0 {
		t.Fatalf("empty text = %d, want 0", got)
	}
	if got := estimateTokens("abcd"); got != 1 {
		t.Fatalf("4 chars = %d, want 1", got)
	}
	msgs := []ProviderMessage{
		{Role: "system", Content: "abcd"},
		{Role: "user", Content: "abcdefgh"},
	}
	if got := estimateMessagesTokens(msgs); got != 3 {
		t.Fatalf("messages = %d, want 3 (1+2)", got)
	}
}

// resolveCeiling is the rule that lets an admin RAISE a limit above the stale
// per-provider default. Without it the screen would accept a number the gateway
// then ignores, and the reasoning-model failure this feature exists to fix would
// survive the fix.
func TestResolveCeiling(t *testing.T) {
	tests := []struct {
		name          string
		providerMax   int
		configuredMax int
		want          int
	}{
		{name: "not configured uses the provider default", providerMax: 8192, configuredMax: 0, want: 8192},
		{name: "configured lower narrows it", providerMax: 8192, configuredMax: 2000, want: 2000},
		{name: "configured higher raises it", providerMax: 8192, configuredMax: 16000, want: 16000},
		{name: "no provider default but configured", providerMax: 0, configuredMax: 4096, want: 4096},
		{name: "neither known", providerMax: 0, configuredMax: 0, want: 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := resolveCeiling(tc.providerMax, tc.configuredMax); got != tc.want {
				t.Fatalf("resolveCeiling(%d, %d) = %d, want %d", tc.providerMax, tc.configuredMax, got, tc.want)
			}
		})
	}
}
