package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/blackboard"
	"ai_avengers/backend/internal/gateway"
)

// TraceabilityEvent records that one artifact was compared against its parent.
//
// WHY this exists: "trained-only" proves WHERE the knowledge came from, not that
// the child artifact stayed INSIDE its parent. A perfectly trained LLD expert can
// still add things the design never asked for, and a Java expert can implement
// things the LLD never specified — and nothing today would notice. This compares
// child against parent and records two lists:
//
//	out_of_parent   — child content with no parent source. FLAGGED, not deleted:
//	                  the client may have wanted it (decided with the client), but
//	                  it must be visible and removable.
//	uncovered_parent— parent requirements the child never addressed. This is the
//	                  "did you follow it COMPLETELY?" half of the guarantee.
const TraceabilityEvent = "traceability_checked"

// TraceLink maps one child section to the parent section(s) it derives from.
type TraceLink struct {
	ChildSection  string   `json:"child_section"`
	ParentSources []string `json:"parent_sources"`
}

// TraceabilityReport is the comparison result for one parent/child pair.
type TraceabilityReport struct {
	ParentLabel string        `json:"parent_label"`
	ChildLabel  string        `json:"child_label"`
	Links       []TraceLink   `json:"links"`
	OutOfParent []string      `json:"out_of_parent"`
	Uncovered   []string      `json:"uncovered_parent"`
	CoveragePct float64       `json:"coverage_pct"`
}

const traceabilitySystemPrompt = `You compare a CHILD artifact against its PARENT artifact to prove the child stayed inside the parent.
Output ONLY one JSON object:
{
  "links": [{"child_section": "name of child section/file/function", "parent_sources": ["name of the parent section it comes from"]}],
  "out_of_parent": ["child content that has NO source in the parent — quote it briefly so it can be reviewed and removed"],
  "uncovered_parent": ["parent requirement/section the child never addressed"],
  "coverage_pct": 0-100
}
Rules:
1. Judge ONLY against the parent text given. Do not use outside knowledge.
2. Every child part must appear in exactly one link entry.
3. "out_of_parent" is for EXTRA content the child added: it is not automatically
   wrong, but the client must be told so they can decide to remove it.
4. "uncovered_parent" is for parent items the child MISSED: report them all.
5. Be precise and quote short phrases; no prose outside the JSON.`

// checkTraceability compares child against parent and returns the report.
func checkTraceability(
	ctx context.Context,
	gw *gateway.ModelGateway,
	parentLabel, parentText, childLabel, childText string,
) (*TraceabilityReport, error) {
	if gw == nil {
		return nil, fmt.Errorf("traceability: no model gateway")
	}
	if strings.TrimSpace(parentText) == "" || strings.TrimSpace(childText) == "" {
		return nil, fmt.Errorf("traceability: parent and child must both have content")
	}

	userPrompt := "PARENT (the artifact this must follow): " + parentLabel + "\n" + truncateForTrace(parentText) +
		"\n\nCHILD (the artifact being checked): " + childLabel + "\n" + truncateForTrace(childText)

	resp, err := gw.Call(ctx, gateway.LLMRequest{
		Model:        gateway.ModelCheap,
		SystemPrompt: traceabilitySystemPrompt,
		UserPrompt:   userPrompt,
		MaxTokens:    2000,
		Temperature:  0.0,
	})
	if err != nil {
		return nil, fmt.Errorf("traceability: llm: %w", err)
	}

	raw := stripJSONFencePublic(resp.Content)
	var rep TraceabilityReport
	if err := json.Unmarshal([]byte(raw), &rep); err != nil {
		return nil, fmt.Errorf("traceability: parse: %w", err)
	}
	rep.ParentLabel = parentLabel
	rep.ChildLabel = childLabel
	if rep.CoveragePct < 0 {
		rep.CoveragePct = 0
	}
	if rep.CoveragePct > 100 {
		rep.CoveragePct = 100
	}
	return &rep, nil
}

// traceCharLimit bounds each artifact pasted into the comparison prompt.
const traceCharLimit = 24000

func truncateForTrace(s string) string {
	if len(s) <= traceCharLimit {
		return s
	}
	return s[:traceCharLimit] + "\n...[truncated]"
}

// runTraceability compares the newest parent artifact with the newest child
// artifact and posts the report.
//
// It never fails the run: the report is evidence for the client (which parts to
// remove, which parent items were missed). Blocking a completed build on a
// judgement call would be worse than showing it, and the client already has the
// flag list to act on.
func (r *WorkflowRunner) runTraceability(
	ctx context.Context,
	workflowID uuid.UUID,
	parentTypes []string,
	childTypes []string,
	parentLabel, childLabel string,
) *TraceabilityReport {
	parent, ok := r.latestArtifactContent(ctx, workflowID, parentTypes)
	if !ok {
		return nil
	}
	child, ok := r.latestArtifactContent(ctx, workflowID, childTypes)
	if !ok {
		return nil
	}

	rep, err := checkTraceability(ctx, r.gateway, parentLabel, parent, childLabel, child)
	if err != nil {
		r.logger.Warn("traceability: check failed",
			zap.String("workflow_id", workflowID.String()), zap.Error(err))
		return nil
	}

	if len(rep.OutOfParent) > 0 || len(rep.Uncovered) > 0 {
		r.logger.Warn("traceability: child is outside or incomplete vs its parent",
			zap.String("workflow_id", workflowID.String()),
			zap.String("child", childLabel),
			zap.Int("out_of_parent", len(rep.OutOfParent)),
			zap.Int("uncovered_parent", len(rep.Uncovered)),
			zap.Float64("coverage_pct", rep.CoveragePct))
	}

	content, _ := json.Marshal(rep)
	if _, postErr := r.store.Post(ctx, blackboard.PostRequest{
		WorkflowID: workflowID,
		EventType:  TraceabilityEvent,
		Content:    json.RawMessage(content),
	}); postErr != nil {
		r.logger.Warn("traceability: could not persist report",
			zap.String("workflow_id", workflowID.String()), zap.Error(postErr))
	}
	return rep
}

// latestArtifactContent returns the body text of the newest event of the given
// types. code_artifact_produced keeps its text under "content"; design events
// keep prose or JSON under a few known keys, so each is tried in turn.
func (r *WorkflowRunner) latestArtifactContent(ctx context.Context, workflowID uuid.UUID, types []string) (string, bool) {
	events, err := r.store.GetByType(ctx, workflowID, types, 0)
	if err != nil || len(events) == 0 {
		return "", false
	}
	latest := events[len(events)-1]
	var m map[string]interface{}
	if err := json.Unmarshal(latest.Content, &m); err != nil {
		return string(latest.Content), len(latest.Content) > 0
	}
	for _, key := range []string{"content", "text", "design", "lld", "body", "summary"} {
		if v, ok := m[key].(string); ok && strings.TrimSpace(v) != "" {
			return v, true
		}
	}
	return string(latest.Content), len(latest.Content) > 0
}
