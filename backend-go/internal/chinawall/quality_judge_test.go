package chinawall

import (
	"strings"
	"testing"
)

func TestParseQualityVerdict_Pass(t *testing.T) {
	raw := `{"accuracy":0.9,"coverage":0.8,"structure":0.85,"overall":0.88,"feedback":""}`
	v, err := parseQualityVerdict(raw, 0.7)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !v.Pass {
		t.Errorf("expected Pass, got Overall=%f", v.Overall)
	}
	if v.Overall != 0.88 || v.Accuracy != 0.9 {
		t.Errorf("unexpected scores: %+v", v)
	}
}

func TestParseQualityVerdict_FailFloor(t *testing.T) {
	raw := `{"accuracy":0.5,"coverage":0.4,"structure":0.6,"overall":0.45,"feedback":"missing complexity analysis"}`
	v, err := parseQualityVerdict(raw, 0.7)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if v.Pass {
		t.Errorf("expected fail below floor, got Pass with Overall=%f", v.Overall)
	}
	if v.Feedback != "missing complexity analysis" {
		t.Errorf("feedback = %q", v.Feedback)
	}
}

func TestParseQualityVerdict_TopLevelNotAverage(t *testing.T) {
	// Components high, overall deliberately low — top-level filter must
	// honour overall alone, not average the rest.
	raw := `{"accuracy":1.0,"coverage":1.0,"structure":1.0,"overall":0.4,"feedback":"does not answer the question"}`
	v, err := parseQualityVerdict(raw, 0.7)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if v.Pass {
		t.Errorf("overall 0.4 must fail even when components are 1.0")
	}
}

func TestParseQualityVerdict_MarkdownFence(t *testing.T) {
	raw := "```json\n{\"accuracy\":0.8,\"coverage\":0.8,\"structure\":0.8,\"overall\":0.8,\"feedback\":\"\"}\n```"
	v, err := parseQualityVerdict(raw, 0.7)
	if err != nil {
		t.Fatalf("parse fenced: %v", err)
	}
	if !v.Pass || v.Overall != 0.8 {
		t.Errorf("fenced parse failed: %+v", v)
	}
}

func TestParseQualityVerdict_Clamp(t *testing.T) {
	raw := `{"accuracy":5,"coverage":-1,"structure":0.5,"overall":2,"feedback":"x"}`
	v, err := parseQualityVerdict(raw, 0.7)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if v.Accuracy != 1 || v.Coverage != 0 || v.Overall != 1 {
		t.Errorf("clamp failed: %+v", v)
	}
	if !v.Pass {
		t.Errorf("clamped overall 1 must pass floor 0.7")
	}
}

func TestParseQualityVerdict_BadJSON(t *testing.T) {
	_, err := parseQualityVerdict("not json at all", 0.7)
	if err == nil {
		t.Fatal("expected error on bad JSON")
	}
}

func TestBuildJudgeReference_Empty(t *testing.T) {
	if got := buildJudgeReference(nil); got != "(no reference chunks)" {
		t.Errorf("empty = %q", got)
	}
}

func TestBuildJudgeReference_Caps(t *testing.T) {
	chunks := make([]CourseChunk, qualityJudgeMaxChunks+3)
	for i := range chunks {
		chunks[i] = CourseChunk{Text: "chunk", Topic: "t", RerankScore: float32(1 - float32(i)*0.01)}
	}
	got := buildJudgeReference(chunks)
	// Should only mention REF 1..5, not 6+.
	if !strings.Contains(got, "[REF 5") || strings.Contains(got, "[REF 6") {
		t.Errorf("cap failed: %s", got)
	}
}
