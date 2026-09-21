package workflow

// The authoring turn — §9 of docs/COLLABORATIVE_DESIGN_ARCHITECTURE.md. This is
// what Aider's role becomes: instead of generating application code, each
// expert edits its OWN design section (assigned by DesignSectionStore, §3.2)
// and, append-only, the spine.
//
// WHY A NEW FILE RATHER THAN EDITING aider_runner.go's EXISTING PATH
//
// aider_runner.go's runAiderIteration/observeWorkspace/verifyWorkspace assume
// the artifact being produced is compilable source code — that is the entire
// point of verify.go (build/test per project). An authoring turn produces
// markdown, and "does it build" has no meaning for markdown. Branching
// aider_runner.go's existing functions on "is this authoring or
// implementation" would put an if-else for a completely different kind of
// artifact inside functions that are already 2000+ lines and already carry
// six root-cause fixes (d9fa1c3, 007f88c, 0a4d3a4, 31becac, 1faf3e5). A
// parallel file that reuses the HTTP-call plumbing but owns its own prompt and
// its own completion check is the smaller, checkable diff.
//
// WHAT IS REUSED, NOT REBUILT
//
//   AiderRunner.initWorkspace   — git init + config + empty commit
//   AiderRunner.listEditableFiles (git ls-files, filtering design/)
//   AiderRunner.checkGenericApproval, .loadExpertTraining
//   AiderRunner's HTTP call to aider-service /iterate — the exact request/
//     response shapes, status-before-decode check, and the "no reason
//     reported" fallback that turned a real production incident into a named
//     error (0a4d3a4) all apply unchanged
//   WorkspaceMerger.MergeWave  — same per-wave merge, same workspace layout
//   DesignSectionStore          — this is the whole reason authoring can be
//     dynamic-file-per-expert instead of one contested final.md
//
// WHAT IS NEW
//
//   buildAuthoringPrompt   — a document-writing task, not a coding task
//   documentCheck          — verifies a markdown workspace, never runs go/npm
//   seedAuthoringWorkspace — seeds the SPINE and every OTHER expert's already-
//     written section as read-only context; the calling expert's OWN section
//     (existing content, if any) is the one editable file

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/blackboard"
)

// finalMD, acceptanceMD, decisionsMD are the four protocol filenames (§2.2 —
// fixed in code because they are a property of the system, not the roster).
const (
	finalMD      = "final.md"
	acceptanceMD = "ACCEPTANCE.md"
	decisionsMD  = "DECISIONS.md"
)

// protocolTemplates seeds the four files on the very first authoring turn of a
// workflow. Written once; every turn after this appends to what is already
// there — see the append-only rule in buildAuthoringPrompt.
var protocolTemplates = map[string]string{
	finalMD: "# Design Spine\n\n" +
		"One paragraph: what is being built.\n\n" +
		"## Module Map\n\n" +
		"| Module | Owner | Section | Status |\n" +
		"|--------|-------|---------|--------|\n\n" +
		"## Statements\n\n" +
		"| ID | Statement | Owner | Status | Decided in |\n" +
		"|----|-----------|-------|--------|------------|\n\n" +
		"See ACCEPTANCE.md for the definition of done and DECISIONS.md for the " +
		"decision log.\n",
	acceptanceMD: "# Acceptance Criteria\n\n" +
		"Each entry: AC-<SECTION_NO>-<n>, owner, section, statement, verify " +
		"command, done-when condition.\n\n",
	decisionsMD: "# Decision Log\n\nNewest first.\n\n",
}

// AuthoringRunner runs one expert's authoring turn: read the spine and every
// other section, write this expert's own section, append to the spine.
//
// A thin wrapper around AiderRunner rather than a duplicate of its fields: the
// workspace directory, the HTTP client, and the aider-service URL are the
// SAME workspace convention and the SAME service AiderRunner already talks to
// — reusing AiderRunner's own methods (below) means a fix to that HTTP call
// (like 0a4d3a4) never has to be applied twice.
type AuthoringRunner struct {
	aider    *AiderRunner
	sections *DesignSectionStore
	logger   *zap.Logger
}

// NewAuthoringRunner wires the authoring runner.
func NewAuthoringRunner(aider *AiderRunner, sections *DesignSectionStore, logger *zap.Logger) *AuthoringRunner {
	return &AuthoringRunner{aider: aider, sections: sections, logger: logger}
}

// AuthoringResult mirrors AiderRunResult's shape so callers already handling
// one can handle the other with no new branching on the result type itself —
// only on which runner produced it.
type AuthoringResult struct {
	CommitSHAs  []string
	Iterations  int
	Completed   bool
	SectionPath string
}

// Run executes one expert's authoring turn.
//
// Unlike AiderRunner.Run (fixed 5 iterations regardless of outcome), an
// authoring turn that passes the document check on iteration 1 stops there —
// there is no "keep iterating on a document that is already internally
// consistent" the way there is for code that might still fail a later test.
func (r *AuthoringRunner) Run(ctx context.Context, req AiderRunRequest) (*AuthoringResult, error) {
	section, err := r.sections.AssignSection(ctx, req.WorkflowID, req.Expert.ID)
	if err != nil {
		return nil, fmt.Errorf("authoring: assign section: %w", err)
	}

	workspacePath := filepath.Join(r.aider.workspaceDir, req.WorkflowID.String(), req.Expert.ID.String())
	if err := r.aider.initWorkspace(ctx, workspacePath); err != nil {
		return nil, fmt.Errorf("authoring: init workspace: %w", err)
	}

	readOnlySections, err := r.seedAuthoringWorkspace(ctx, req.WorkflowID, section, workspacePath)
	if err != nil {
		return nil, fmt.Errorf("authoring: seed workspace: %w", err)
	}

	training, trainingChunks, err := r.aider.loadExpertTraining(ctx, req.Expert.ID, req.TaskDescription)
	if err != nil {
		r.logger.Warn("authoring: failed to load expert training, continuing without it",
			zap.Error(err), zap.String("expert_id", req.Expert.ID.String()))
	}
	genericApproved, genericTopics := r.aider.checkGenericApproval(ctx, req.WorkflowID, req.Expert.ID)

	const maxIterations = 3 // a document task needs far fewer retries than code
	var commitSHAs []string
	var lastCheckErr error
	var result documentCheckResult

	for iteration := 1; iteration <= maxIterations; iteration++ {
		r.logger.Info("authoring iteration starting",
			zap.Int("iteration", iteration), zap.Int("max", maxIterations),
			zap.String("section", section.SectionPath))

		message := r.buildAuthoringPrompt(req, section, readOnlySections, lastCheckErr)
		if training != "" {
			message += "\n\n" + training
		}
		message += r.aider.build7030EnforcementInstructions(
			genericApproved, genericTopics, trainingChunks > 0, len(readOnlySections) > 0,
		)

		editFiles, listErr := r.aider.listEditableFiles(ctx, workspacePath)
		if listErr != nil {
			editFiles = []string{}
		}

		commitSHA, callErr := r.callAiderService(ctx, req, workspacePath, message, editFiles, readOnlySections, iteration)
		if callErr != nil {
			lastCheckErr = callErr
			r.logger.Warn("authoring iteration produced nothing, continuing",
				zap.Int("iteration", iteration), zap.Error(callErr))
			continue
		}
		if commitSHA != "" {
			commitSHAs = append(commitSHAs, commitSHA)
		}

		result = documentCheck(workspacePath)
		r.logger.Info("authoring document check",
			zap.Int("iteration", iteration),
			zap.Bool("passed", result.Passed),
			zap.Strings("problems", result.Problems),
		)
		if result.Passed {
			r.publishDesignArtifact(ctx, req, section, workspacePath, commitSHAs)
			return &AuthoringResult{
				CommitSHAs:  commitSHAs,
				Iterations:  iteration,
				Completed:   true,
				SectionPath: section.SectionPath,
			}, nil
		}
		lastCheckErr = fmt.Errorf("document check failed: %s", strings.Join(result.Problems, "; "))
	}

	if len(commitSHAs) == 0 {
		if lastCheckErr != nil {
			return nil, fmt.Errorf("authoring: no commit after %d iterations: %w", maxIterations, lastCheckErr)
		}
		return nil, fmt.Errorf("authoring: no commit after %d iterations", maxIterations)
	}
	// Committed something even though the document check never fully passed
	// — still real, git-committed work. Publishing it is what makes the UI's
	// Files panel match what `git log` in this workspace already shows.
	r.publishDesignArtifact(ctx, req, section, workspacePath, commitSHAs)
	return &AuthoringResult{
		CommitSHAs:  commitSHAs,
		Iterations:  maxIterations,
		Completed:   false, // committed something, but the document check never passed
		SectionPath: section.SectionPath,
	}, nil
}

// seedAuthoringWorkspace writes the four protocol files (creating them from
// the template on the very first turn) and every OTHER expert's section as
// read-only context, then commits. Returns the read-only file list.
//
// WHY every other section, not just declared dependencies: doc §4.1 scopes a
// CODING turn to declared dependencies to bound context as the roster grows.
// An authoring turn is different — it edits the SPINE, which by definition
// touches the whole module map, so an authoring expert needs to see who else
// has already written what, not just its own upstream tasks.
func (r *AuthoringRunner) seedAuthoringWorkspace(
	ctx context.Context, workflowID uuid.UUID, mySection DesignSection, workspacePath string,
) ([]string, error) {
	mainWorkspace := filepath.Join(r.aider.workspaceDir, workflowID.String(), "main")

	readOnly := []string{}
	for name, template := range protocolTemplates {
		content := template
		if existing, err := readFromMain(mainWorkspace, name); err == nil {
			content = existing
		}
		if err := writeWorkspaceFile(workspacePath, name, content); err != nil {
			return nil, err
		}
	}

	allSections, err := r.sections.ListSections(ctx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("list sections: %w", err)
	}
	for _, sec := range allSections {
		if sec.ExpertID == mySection.ExpertID {
			continue // my own section is editable, not read-only — handled below
		}
		content, err := readFromMain(mainWorkspace, sec.SectionPath)
		if err != nil {
			continue // not yet authored — nothing to seed
		}
		if err := writeWorkspaceFile(workspacePath, sec.SectionPath, content); err != nil {
			return nil, err
		}
		readOnly = append(readOnly, sec.SectionPath)
	}
	readOnly = append(readOnly, finalMD, acceptanceMD, decisionsMD)

	// My own section: seed with its existing content (a revision), or leave
	// absent (a fresh section — Aider creates it, same as any new file).
	if content, err := readFromMain(mainWorkspace, mySection.SectionPath); err == nil {
		if err := writeWorkspaceFile(workspacePath, mySection.SectionPath, content); err != nil {
			return nil, err
		}
	}

	if err := gitAddCommit(ctx, workspacePath, "chore: seed authoring workspace"); err != nil {
		return nil, err
	}
	return readOnly, nil
}

// readFromMain reads relPath from the merged main/ workspace. Returns an
// error if main/ or the file does not exist yet — the normal case before the
// first wave has merged anything.
func readFromMain(mainWorkspace, relPath string) (string, error) {
	data, err := os.ReadFile(filepath.Join(mainWorkspace, relPath))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func writeWorkspaceFile(workspacePath, relPath, content string) error {
	full := filepath.Join(workspacePath, relPath)
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		return fmt.Errorf("mkdir for %s: %w", relPath, err)
	}
	if err := os.WriteFile(full, []byte(content), 0644); err != nil {
		return fmt.Errorf("write %s: %w", relPath, err)
	}
	return nil
}

func gitAddCommit(ctx context.Context, workspacePath, message string) error {
	add := exec.CommandContext(ctx, "git", "add", ".")
	add.Dir = workspacePath
	if err := add.Run(); err != nil {
		return fmt.Errorf("git add: %w", err)
	}
	commit := exec.CommandContext(ctx, "git", "commit", "-m", message)
	commit.Dir = workspacePath
	if output, err := commit.CombinedOutput(); err != nil {
		if !strings.Contains(string(output), "nothing to commit") {
			return fmt.Errorf("git commit: %w (output: %s)", err, string(output))
		}
	}
	return nil
}

// buildAuthoringPrompt is the document-writing equivalent of
// AiderRunner.buildImplementationPrompt. Explicitly NOT that function with an
// if-branch: the two tasks ("write code that compiles" vs "write/amend a
// design section and update the spine") share no sentence, and forcing them
// into one function would make every future edit to either risk the other.
func (r *AuthoringRunner) buildAuthoringPrompt(
	req AiderRunRequest, section DesignSection, readOnlySections []string, lastCheckErr error,
) string {
	var sb strings.Builder

	sb.WriteString("DESIGN AUTHORING TASK — write or revise a design document. This is NOT a\n")
	sb.WriteString("coding task: produce markdown, not source code.\n\n")
	sb.WriteString(fmt.Sprintf("Your assignment: %s\n\n", req.TaskTitle))
	if req.TaskDescription != "" && req.TaskDescription != req.TaskTitle {
		sb.WriteString(fmt.Sprintf("Scope:\n%s\n\n", req.TaskDescription))
	}

	sb.WriteString(fmt.Sprintf("YOUR SECTION (edit this): %s\n\n", section.SectionPath))

	sb.WriteString("RULES:\n")
	sb.WriteString(fmt.Sprintf("  - Edit ONLY %s and, append-only, %s and %s.\n", section.SectionPath, finalMD, acceptanceMD))
	sb.WriteString("  - Do NOT edit any other expert's section file — they are open for you to\n")
	sb.WriteString("    read, not to change.\n")
	sb.WriteString(fmt.Sprintf("  - In %s: add exactly one row to the Module Map for your module, and\n", finalMD))
	sb.WriteString("    you may ADD statement rows (never remove or edit an existing row — if you\n")
	sb.WriteString("    disagree with one, that is a conflict, not an edit).\n")
	sb.WriteString(fmt.Sprintf("  - In %s: add at least one criterion for your section, in the exact\n", acceptanceMD))
	sb.WriteString("    format already in the file (AC-<id>, owner, section, statement, verify,\n")
	sb.WriteString("    done when — the verify line MUST be a real command).\n")
	sb.WriteString("  - Use SEARCH/REPLACE blocks for every change, including appends.\n\n")

	if len(readOnlySections) > 0 {
		sb.WriteString("READ-ONLY CONTEXT (other experts' work, already open):\n")
		for _, f := range readOnlySections {
			sb.WriteString(fmt.Sprintf("  - %s\n", f))
		}
		sb.WriteString("\n")
	}

	if lastCheckErr != nil {
		sb.WriteString("THE PREVIOUS ATTEMPT FAILED THE DOCUMENT CHECK — fix this first:\n")
		sb.WriteString(lastCheckErr.Error())
		sb.WriteString("\n\n")
	}

	return sb.String()
}

// callAiderService is AiderRunner.runAiderIteration's HTTP call, factored out
// so both callers share it verbatim. Every behaviour documented at the
// original call site (status-before-decode, the "no reason reported"
// fallback, non-nil edit/read-only lists to avoid the null-vs-empty-list 422)
// applies here unchanged — see aider_runner.go:1089-1114 for why each one
// exists; duplicating that reasoning here would drift the next time one side
// is fixed and the other is not.
func (r *AuthoringRunner) callAiderService(
	ctx context.Context, req AiderRunRequest, workspacePath, message string,
	editFiles, readOnlyFiles []string, iteration int,
) (string, error) {
	type aiderServiceRequest struct {
		WorkspacePath string   `json:"workspace_path"`
		Message       string   `json:"message"`
		ExpertID      string   `json:"expert_id"`
		WorkflowID    string   `json:"workflow_id"`
		EditFiles     []string `json:"edit_files"`
		ReadOnlyFiles []string `json:"read_only_files"`
	}
	type aiderServiceResponse struct {
		CommitSHA string `json:"commit_sha"`
		Success   bool   `json:"success"`
		Error     string `json:"error"`
	}

	body, err := json.Marshal(aiderServiceRequest{
		WorkspacePath: workspacePath,
		Message:       message,
		ExpertID:      req.Expert.ID.String(),
		WorkflowID:    req.WorkflowID.String(),
		EditFiles:     editFiles,
		ReadOnlyFiles: readOnlyFiles,
	})
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		r.aider.aiderServiceURL+"/iterate", strings.NewReader(string(body)))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Workflow-ID", req.WorkflowID.String())
	httpReq.Header.Set("X-Expert-ID", req.Expert.ID.String())

	httpResp, err := r.aider.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("aider service: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}
	if httpResp.StatusCode != http.StatusOK {
		snippet := string(respBody)
		if len(snippet) > 1000 {
			snippet = snippet[:1000] + "...(truncated)"
		}
		return "", fmt.Errorf("aider service status %d: %s", httpResp.StatusCode, snippet)
	}

	var resp aiderServiceResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return "", fmt.Errorf("decode response: %w (body: %s)", err, string(respBody))
	}
	if !resp.Success {
		reason := resp.Error
		if reason == "" {
			reason = fmt.Sprintf("no reason reported by aider-service (body: %s)", string(respBody))
		}
		return "", fmt.Errorf("aider: %s", reason)
	}
	return resp.CommitSHA, nil
}

// ============================================================
// Document check — the authoring-phase completion criterion (§9's load-
// bearing row). This replaces verify.go's build/test check for a markdown
// workspace: verify.go correctly reports "unavailable" when it finds no
// go.mod/package.json, and AllPassed() correctly treats unavailable as
// not-complete — which means an authoring turn run through verify.go would
// NEVER complete. This function is the missing verifier, shipped in the same
// change that introduces the authoring path, not after.
// ============================================================

// documentCheckResult is the authoring-turn equivalent of
// WorkspaceVerification (verify.go) — same shape (Passed + a list of
// problems), different subject matter (documents, not compiled code).
type documentCheckResult struct {
	Passed   bool
	Problems []string
}

// acceptanceLineRe matches one "## AC-..." heading line in ACCEPTANCE.md.
var acceptanceLineRe = regexp.MustCompile(`(?m)^## (AC-[A-Za-z0-9_-]+)\b.*owner:\s*(\S+).*section:\s*(\S+)`)

// statementRowRe matches one statement-table row in final.md:
// | C-014 | text | owner | accepted | DEC-007 |
//
// The status group is [^|]+? (anything but a pipe), not a letters-only class.
// A letters-only class was tried first and missed the exact contradiction
// this function exists to catch: "superseded by C-015" contains digits and a
// hyphen, so a row with that status silently failed to match at all, and the
// accepted/superseded check below saw only the "accepted" row — passing a
// workspace that should have failed. Caught by running this check against a
// contradictory pair before trusting it, not by inspection.
var statementRowRe = regexp.MustCompile(`(?m)^\|\s*(C-[A-Za-z0-9_-]+)\s*\|[^|]*\|[^|]*\|\s*([^|]+?)\s*\|`)

// documentCheck runs the checks §9 specifies against ONE expert's authoring
// workspace:
//   - final.md and ACCEPTANCE.md and DECISIONS.md exist and are non-empty
//   - every acceptance criterion has an owner and a verify line
//   - no statement is both "accepted" and "superseded" — literally, no
//     statement ID appears twice with contradictory status; §5's rule is that
//     a superseded row is a NEW row referencing the old one, not an edit, so
//     two rows sharing an ID is itself the defect this check exists to catch
//
// §9 also asks that every path in workflow_design_sections exist on disk.
// Deliberately not checked here: this workspace only ever holds the sections
// seedAuthoringWorkspace chose to seed (peers already merged into main/ at
// seed time) — a section authored by a LATER wave cannot exist here yet, and
// checking for it would fail every early-wave turn for a reason unrelated to
// what this expert wrote. That check belongs after the wave merge, against
// main/, which is the per-wave integration turn's job (§4.3), not this one
// expert's document check.
func documentCheck(workspacePath string) documentCheckResult {
	var problems []string

	for _, f := range []string{finalMD, acceptanceMD, decisionsMD} {
		data, err := os.ReadFile(filepath.Join(workspacePath, f))
		if err != nil {
			problems = append(problems, fmt.Sprintf("%s: missing or unreadable (%v)", f, err))
			continue
		}
		if len(strings.TrimSpace(string(data))) == 0 {
			problems = append(problems, fmt.Sprintf("%s: exists but is empty", f))
		}
	}
	if len(problems) > 0 {
		// The rest of the checks read these same files; no point compounding
		// confusing follow-on errors onto a missing-file report.
		return documentCheckResult{Passed: false, Problems: problems}
	}

	acceptance, _ := os.ReadFile(filepath.Join(workspacePath, acceptanceMD))
	matches := acceptanceLineRe.FindAllStringSubmatch(string(acceptance), -1)
	if len(matches) == 0 {
		problems = append(problems, "ACCEPTANCE.md: no criteria found (need at least one '## AC-...' entry with owner and section)")
	}
	seenAC := map[string]bool{}
	for _, m := range matches {
		id := m[1]
		if seenAC[id] {
			problems = append(problems, fmt.Sprintf("ACCEPTANCE.md: duplicate criterion id %s", id))
		}
		seenAC[id] = true
		if !strings.Contains(acceptance2VerifyBlock(string(acceptance), id), "Verify:") {
			problems = append(problems, fmt.Sprintf("%s: missing a Verify: line", id))
		}
	}

	finalContent, _ := os.ReadFile(filepath.Join(workspacePath, finalMD))
	rows := statementRowRe.FindAllStringSubmatch(string(finalContent), -1)
	statusByID := map[string][]string{}
	for _, row := range rows {
		id, status := row[1], strings.ToLower(strings.TrimSpace(row[2]))
		statusByID[id] = append(statusByID[id], status)
	}
	for id, statuses := range statusByID {
		hasAccepted := false
		hasSuperseded := false
		for _, s := range statuses {
			if s == "accepted" {
				hasAccepted = true
			}
			if strings.HasPrefix(s, "superseded") {
				hasSuperseded = true
			}
		}
		if hasAccepted && hasSuperseded {
			problems = append(problems, fmt.Sprintf("final.md: statement %s is both accepted and superseded", id))
		}
	}

	return documentCheckResult{Passed: len(problems) == 0, Problems: problems}
}

// acceptance2VerifyBlock returns the text of one AC-<id> entry up to the next
// "## " heading or end of file, so the Verify: check only looks inside that
// criterion's own block rather than anywhere later in the file.
func acceptance2VerifyBlock(content, id string) string {
	idx := strings.Index(content, "## "+id)
	if idx == -1 {
		return ""
	}
	rest := content[idx:]
	if next := strings.Index(rest[1:], "\n## "); next != -1 {
		return rest[:next+1]
	}
	return rest
}
