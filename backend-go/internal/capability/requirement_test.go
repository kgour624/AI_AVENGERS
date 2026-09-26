package capability

import "testing"

func expert(id, kind, lang string) Expert {
	return Expert{ID: id, Name: id, Declaration: Declaration{Kind: Kind(kind), Language: LanguageAliases(lang)}}
}

// The incident this package fixes: a design expert must never be handed the
// implementation phase.
func TestImplementationPhaseNeedsImplementationExpert(t *testing.T) {
	designers := []Expert{expert("d1", "design", "")}
	if _, err := SelectForPhase(PhaseImplementation, designers, "go"); err == nil {
		t.Fatal("design-only experts must not satisfy the implementation phase")
	}

	goExpert := []Expert{expert("g1", "implementation", "go")}
	got, err := SelectForPhase(PhaseImplementation, goExpert, "go")
	if err != nil || len(got) != 1 {
		t.Fatalf("go implementation expert should match: %v %d", err, len(got))
	}
}

// A Python programmer must not receive a Go task, and an implementation expert
// that declares no language must not receive a language-specific task.
func TestLanguageMismatchIsExcluded(t *testing.T) {
	py := []Expert{expert("p1", "implementation", "python")}
	if _, err := SelectForPhase(PhaseImplementation, py, "go"); err == nil {
		t.Fatal("python expert must not match a go requirement")
	}
	if _, err := SelectForPhase(PhaseImplementation, py, "python"); err != nil {
		t.Fatalf("python expert should match python: %v", err)
	}
	noLang := []Expert{expert("x1", "implementation", "")}
	if _, err := SelectForPhase(PhaseImplementation, noLang, "go"); err == nil {
		t.Fatal("an implementation expert with no declared language must not take a language-specific task")
	}
}

// QA is a testing expert's job only.
func TestQAPhaseNeedsTestingExpert(t *testing.T) {
	impl := []Expert{expert("g1", "implementation", "go")}
	if _, err := SelectForPhase(PhaseQA, impl, "go"); err == nil {
		t.Fatal("implementation experts must not run QA")
	}
	testers := []Expert{expert("t1", "testing", "go")}
	if got, err := SelectForPhase(PhaseQA, testers, "go"); err != nil || len(got) != 1 {
		t.Fatalf("testing expert should match QA: %v", err)
	}
}

// Unclassified experts satisfy nothing, and unknown phases fall back to nobody.
func TestUnclassifiedAndUnknownPhase(t *testing.T) {
	none := []Expert{expert("u1", "other", "")}
	if _, err := SelectForPhase(PhaseImplementation, none, "go"); err == nil {
		t.Fatal("unclassified/other must not satisfy implementation")
	}
	if got, err := SelectForPhase("some_future_phase", []Expert{expert("d1", "design", "")}, ""); err != nil || len(got) != 0 {
		t.Fatalf("unknown phase must select nobody: %v %d", err, len(got))
	}
}
