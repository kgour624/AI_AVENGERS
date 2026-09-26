package chinawall

import (
	"testing"

	"go.uber.org/zap"
)

// Pins the fix for the production refusal: an expert whose domain is stored as
// "system design" (space) must resolve to the SAME tuned profile as
// "system_design" (underscore). Before this, the lookup missed, the strict
// BaseProfile was used, and a trained expert refused its own domain.
func TestNormDomainKeyUnifiesSpacesHyphensAndCase(t *testing.T) {
	want := "system_design"
	for _, in := range []string{"system_design", "system design", "System Design", "  System   Design  ", "system-design"} {
		if got := normDomainKey(in); got != want {
			t.Fatalf("normDomainKey(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRegistryGetMatchesSpaceFormOfStoredDomain(t *testing.T) {
	r := NewDomainRegistry(nil, zap.NewNop())
	// Simulate a stored profile registered under the underscore key.
	r.profiles["system_design"] = &DomainProfile{Domain: "system_design", CoverageMode: CoverageModeApplyPrinciples}

	got := r.Get("system design")
	if got == nil || got.CoverageMode != CoverageModeApplyPrinciples {
		t.Fatalf("Get(\"system design\") returned %+v, want the registered APPLY_PRINCIPLES profile", got)
	}
}
