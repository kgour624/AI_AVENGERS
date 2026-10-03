package collab

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestFallbackPlanUsesCategoryThenDomainInOrder(t *testing.T) {
	e1, e2 := uuid.New(), uuid.New()
	experts := []ExpertInfo{
		{ID: e1, Name: "Alice", Domain: "backend", CategoryName: "Database Schema"},
		{ID: e2, Name: "Bob", Domain: "frontend", CategoryName: ""},
	}
	plan := FallbackPlan(experts)
	if len(plan) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(plan))
	}
	if plan[0].ExpertID != e1 || plan[0].SectionTitle != "Database Schema" {
		t.Errorf("section 0 = %+v, want expert %s with title 'Database Schema'", plan[0], e1)
	}
	if plan[1].ExpertID != e2 || plan[1].SectionTitle != "frontend" {
		t.Errorf("section 1 = %+v, want expert %s with title 'frontend' (domain fallback)", plan[1], e2)
	}
}

func TestPlanSectionsWithNilGatewayUsesFallback(t *testing.T) {
	experts := []ExpertInfo{{ID: uuid.New(), Name: "Alice", Domain: "backend"}}
	plan := PlanSections(context.Background(), nil, "how should I design this?", experts, nil)
	if len(plan) != 1 || plan[0].SectionTitle != "backend" {
		t.Fatalf("expected fallback plan with 1 section titled 'backend', got %+v", plan)
	}
}

func TestRunRelaySequencesAndPassesPeerContext(t *testing.T) {
	e1, e2 := uuid.New(), uuid.New()
	plan := []Section{
		{ExpertID: e1, ExpertName: "Alice", SectionTitle: "Design"},
		{ExpertID: e2, ExpertName: "Bob", SectionTitle: "Implementation"},
	}

	var receivedPeerContextForSecond string
	run := func(ctx context.Context, expertID uuid.UUID, peerContext string) (string, error) {
		if expertID == e1 {
			if peerContext != "" {
				t.Errorf("first expert should get empty peerContext, got %q", peerContext)
			}
			return "Alice's answer", nil
		}
		receivedPeerContextForSecond = peerContext
		return "Bob's answer", nil
	}

	result := RunRelay(context.Background(), plan, run, nil)

	if len(result.Sections) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(result.Sections))
	}
	if result.Sections[0].Content != "Alice's answer" {
		t.Errorf("section 0 content = %q", result.Sections[0].Content)
	}
	if result.Sections[1].Content != "Bob's answer" {
		t.Errorf("section 1 content = %q", result.Sections[1].Content)
	}
	if receivedPeerContextForSecond == "" {
		t.Error("second expert should have received non-empty peer context from the first expert's answer")
	}
	if len(result.Failed) != 0 {
		t.Errorf("expected no failures, got %v", result.Failed)
	}
}

func TestRunRelayFailSoftOnExpertError(t *testing.T) {
	e1, e2 := uuid.New(), uuid.New()
	plan := []Section{
		{ExpertID: e1, ExpertName: "Alice", SectionTitle: "Design"},
		{ExpertID: e2, ExpertName: "Bob", SectionTitle: "Implementation"},
	}

	run := func(ctx context.Context, expertID uuid.UUID, peerContext string) (string, error) {
		if expertID == e1 {
			return "", errors.New("rate_limit_exceeded")
		}
		return "Bob's answer", nil
	}

	var progressCalls int
	result := RunRelay(context.Background(), plan, run, func(section Section, index, total int) {
		progressCalls++
	})

	if len(result.Sections) != 2 {
		t.Fatalf("expected both sections present even after a failure, got %d", len(result.Sections))
	}
	if result.Sections[0].Content == "" {
		t.Error("failed expert's section should still have a placeholder Content, not empty")
	}
	if _, ok := result.Failed[e1.String()]; !ok {
		t.Error("expected e1 to be recorded in Failed")
	}
	if result.Sections[1].Content != "Bob's answer" {
		t.Errorf("second expert should still run after the first failed, got %q", result.Sections[1].Content)
	}
	if progressCalls != 2 {
		t.Errorf("expected onProgress called twice (once per section, including the failed one), got %d", progressCalls)
	}
}

func TestAssembleMarkdownJoinsInOrder(t *testing.T) {
	sections := []Section{
		{ExpertName: "Alice", SectionTitle: "Design", Content: "Use a queue."},
		{ExpertName: "Bob", SectionTitle: "Implementation", Content: "Here is the code."},
	}
	md := AssembleMarkdown(sections)
	wantFirst := "## Design (Alice)\n\nUse a queue."
	wantSecond := "## Implementation (Bob)\n\nHere is the code."
	if idx1 := indexOf(md, wantFirst); idx1 == -1 {
		t.Errorf("missing or malformed first section, got:\n%s", md)
	}
	if idx2 := indexOf(md, wantSecond); idx2 == -1 || indexOf(md, wantSecond) < indexOf(md, wantFirst) {
		t.Errorf("second section missing or out of order, got:\n%s", md)
	}
}

func indexOf(haystack, needle string) int {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}
