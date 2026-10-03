package workflow

// Combined design document assembly (handoff phase).
//
// WHY a separate file: runner.go's handoff block in Run() is already large;
// isolating markdown assembly and versioning here lets it be unit tested
// without exercising the full Run() state machine — the same reason
// export_git.go and client_repo.go are their own files in this package
// rather than inlined into runner.go.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"ai_avengers/backend/internal/blackboard"
)

// combinedDesignDocEvent is the blackboard event type posted after a combined
// design markdown file is written. kanban_sse.go / files_sse.go already show
// the pattern of a UI subscribing to a fixed event-type string for a terminal
// action; this is that same pattern for a "Download Unified Design" button.
const combinedDesignDocEvent = "combined_design_doc_produced"

// designArtifactSections maps blackboard event_type -> markdown header, in the
// fixed order a design document should read in. Not alphabetical and not
// GetByType's return order (DB order is not a documented guarantee here);
// this is the order a human reviewing the handoff doc end to end wants.
var designArtifactSections = []struct {
	EventType string
	Header    string
}{
	{"requirement_captured", "## Requirement"},
	{"architecture_decision", "## Architecture"},
	{"data_model_proposed", "## Database Schema"},
	{"api_contract_proposed", "## API Contract"},
	{"module_design_proposed", "## Module Design"},
	{"code_artifact_produced", "## Code Artifacts"},
}

// combinedDesignDocResult is handed back to the caller so it can decide
// whether/what to post to the blackboard. Kept separate from the write
// itself so a write failure (logged, non-fatal — see runner.go call site)
// never has to be inspected for partial success.
type combinedDesignDocResult struct {
	FilePath string
	FileName string
	Version  int // 0 = final_design.md (first run); 1,2,... = redesign_N.md
}

// writeCombinedDesignDoc assembles every design artifact fetched for the
// handoff gate into one markdown file and resolves the correct filename
// under versioning rules:
//
//   - designRevisionID == "" (first run, never redesigned): final_design.md.
//     If it already exists (a resumed workflow re-entering handoff without a
//     redesign), it is overwritten — it is still revision zero of the same
//     design, not a new one.
//   - designRevisionID != "": this is a client-requested redesign. The next
//     sequential redesign_N.md is created, never overwritten, so every past
//     redesign stays on disk for comparison.
//
// WHY the sequence is counted from disk, not from an in-memory attempt
// counter: Run()'s maxDesignAttempts loop variable is local to one call and
// is not visible here, and a workflow resumed after a pod restart re-enters
// this function with no attempt counter in memory at all (see runner.go's
// RECOVERY doc comment). The workspace directory is the only durable record
// of how many redesigns have already been written.
func writeCombinedDesignDoc(
	workflowWorkspaceMain string,
	workflowID uuid.UUID,
	designRevisionID string,
	redesignGoal string,
	artifacts []blackboard.Event,
) (*combinedDesignDocResult, error) {
	if err := os.MkdirAll(workflowWorkspaceMain, 0755); err != nil {
		return nil, fmt.Errorf("combined design doc: mkdir workspace: %w", err)
	}

	fileName := "final_design.md"
	version := 0
	if designRevisionID != "" {
		n, err := nextRedesignSequence(workflowWorkspaceMain)
		if err != nil {
			return nil, fmt.Errorf("combined design doc: resolve redesign sequence: %w", err)
		}
		version = n
		fileName = fmt.Sprintf("redesign_%d.md", n)
	}

	content := assembleCombinedMarkdown(workflowID, designRevisionID, redesignGoal, artifacts)

	fullPath := filepath.Join(workflowWorkspaceMain, fileName)
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("combined design doc: write %s: %w", fileName, err)
	}

	return &combinedDesignDocResult{FilePath: fullPath, FileName: fileName, Version: version}, nil
}

// nextRedesignSequence scans the workspace for existing redesign_N.md files
// and returns the next unused N, starting at 1.
func nextRedesignSequence(workflowWorkspaceMain string) (int, error) {
	entries, err := os.ReadDir(workflowWorkspaceMain)
	if err != nil {
		if os.IsNotExist(err) {
			return 1, nil
		}
		return 0, err
	}
	max := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, "redesign_") || !strings.HasSuffix(name, ".md") {
			continue
		}
		numPart := strings.TrimSuffix(strings.TrimPrefix(name, "redesign_"), ".md")
		var n int
		if _, scanErr := fmt.Sscanf(numPart, "%d", &n); scanErr == nil && n > max {
			max = n
		}
	}
	return max + 1, nil
}

// assembleCombinedMarkdown concatenates artifact content under fixed section
// headers. An artifact type not in designArtifactSections (a future event
// type) is appended under a trailing "## Other Artifacts" section instead of
// silently dropped — a new artifact type must become visible here, not
// disappear from the handoff document.
func assembleCombinedMarkdown(
	workflowID uuid.UUID,
	designRevisionID string,
	redesignGoal string,
	artifacts []blackboard.Event,
) string {
	byType := make(map[string][]blackboard.Event)
	for _, ev := range artifacts {
		byType[ev.EventType] = append(byType[ev.EventType], ev)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "# Design Document — Workflow %s\n\n", workflowID.String())
	fmt.Fprintf(&b, "_Generated: %s_\n\n", time.Now().UTC().Format(time.RFC3339))
	if designRevisionID != "" {
		fmt.Fprintf(&b, "_Redesign revision: %s_\n", designRevisionID)
		if redesignGoal != "" {
			fmt.Fprintf(&b, "_Redesign goal: %s_\n", redesignGoal)
		}
		b.WriteString("\n")
	}

	seen := make(map[string]bool)
	for _, section := range designArtifactSections {
		events := byType[section.EventType]
		seen[section.EventType] = true
		if len(events) == 0 {
			continue
		}
		b.WriteString(section.Header + "\n\n")
		for _, ev := range events {
			b.WriteString(extractEventText(ev))
			b.WriteString("\n\n")
		}
	}

	var otherTypes []string
	for t := range byType {
		if !seen[t] {
			otherTypes = append(otherTypes, t)
		}
	}
	if len(otherTypes) > 0 {
		b.WriteString("## Other Artifacts\n\n")
		for _, t := range otherTypes {
			for _, ev := range byType[t] {
				fmt.Fprintf(&b, "### %s\n\n", t)
				b.WriteString(extractEventText(ev))
				b.WriteString("\n\n")
			}
		}
	}
	return b.String()
}

// lookupCombinedDesignDoc returns the most recently posted combined design
// document's file path, name and version, so a download handler can serve the
// current on-disk file without re-deriving the versioning logic above.
//
// WHY "most recent wins" and not "first": a redesign posts a new
// combinedDesignDocEvent for the new redesign_N.md, and the download button
// must always offer the latest revision, exactly like fileContentsForWorkflow
// in publish.go treats later events as the current content for any other
// artifact path.
//
// WHY a separate lookup instead of extending fileContentsForWorkflow /
// artifactPathAndBody (files_list.go, publish.go): those two are shared by the
// FILES tab, the chat file picker and publishing, and are keyed on
// PostedByExpertID != nil plus the file_path/filename keys an expert-authored
// artifact carries. The combined design doc is system-posted
// (PostedByExpertID == nil) and carries file_path/file_name/version keys of
// its own shape — bending the shared helpers to also accept a nil-expert,
// differently-keyed event would make them harder to reason about for every
// other artifact type that already relies on them.
func lookupCombinedDesignDoc(events []blackboard.Event) (filePath string, fileName string, version int, found bool) {
	for i := len(events) - 1; i >= 0; i-- {
		ev := events[i]
		if ev.EventType != combinedDesignDocEvent {
			continue
		}
		var m map[string]interface{}
		if err := json.Unmarshal(ev.Content, &m); err != nil {
			continue
		}
		fp, _ := m["file_path"].(string)
		fn, _ := m["file_name"].(string)
		if fp == "" {
			continue
		}
		v := 0
		switch n := m["version"].(type) {
		case float64:
			v = int(n)
		case int:
			v = n
		}
		return fp, fn, v, true
	}
	return "", "", 0, false
}

// extractEventText pulls human-readable text out of one event's JSON
// content. Same key list and same unmarshal-then-raw-fallback algorithm as
// traceability.go's latestArtifactContent — reused here instead of
// redefined because that function already established which keys design
// and code artifacts actually store text under; it just operates on a
// single already-fetched event instead of calling GetByType itself, so it
// is a free function here, not a *WorkflowRunner method.
func extractEventText(ev blackboard.Event) string {
	var m map[string]interface{}
	if err := json.Unmarshal(ev.Content, &m); err != nil {
		return string(ev.Content)
	}
	for _, key := range []string{"content", "text", "design", "lld", "body", "summary"} {
		if v, ok := m[key].(string); ok && strings.TrimSpace(v) != "" {
			return v
		}
	}
	return string(ev.Content)
}
