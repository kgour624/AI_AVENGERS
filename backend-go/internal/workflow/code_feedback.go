package workflow

// §17 of docs/COLLABORATIVE_DESIGN_ARCHITECTURE.md — the code-feedback loop.
//
// Before this, a workflow ended when the client took the harness away. This
// closes the loop: the repo the coding agent actually built comes back, the
// design is compared against it, and every mismatch becomes a question routed
// to the expert who owns that part of the design.
//
// THE FIVE STEPS, AND WHERE EACH ONE LIVES
//
//	INGEST     cloneClientRepo                 — shallow clone, token never on disk
//	DIFF       code_feedback_scan.go           — Verify commands + contract set diff
//	REPORT     postReport                      — one code_feedback_ingested event
//	AMEND      postAmendments                  — one code_feedback_question per finding,
//	                                            routed to the expert who owns it
//	GATE       amendment.go (§7.5)             — the expert's answer becomes a real
//	                                            amendment, the client approves it,
//	                                            and DECISIONS.md records the round
//
// WHERE THE ROUND GETS RECORDED
//
// §17.2 step 6 asks for a DECISIONS.md entry. That entry is written by the §7.5
// apply path (amendment.go), not from here, and that is the right seam: the log
// records a decision a CLIENT made, and nothing in this file has one. A finding
// is a question; it becomes a decision only after an expert turns it into a
// concrete amendment and the client approves that. Writing DECISIONS.md from the
// ingest would be an unapproved change to the contract the code was built
// against — exactly what §7.5 step 2 exists to prevent.
//
// WHY EVERY FINDING IS A QUESTION AND NOT A VERDICT
//
// The comparator is heuristic — see code_feedback_scan.go's header for exactly
// how far it reaches. Routing a finding to the owning expert as "still
// required, or drop it?" is correct whether the finding is right or wrong. A
// false positive costs one question; an automatic write would cost the
// contract.

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/blackboard"
)

// Timeouts. The ingest runs detached from the HTTP request, so every step needs
// its own bound — without one a hung clone or a test suite that never finishes
// would leave a goroutine and a temp directory behind forever.
const (
	codeFeedbackTotalTimeout  = 20 * time.Minute
	codeFeedbackCloneTimeout  = 5 * time.Minute
	codeFeedbackVerifyTimeout = 2 * time.Minute
	// codeFeedbackReportTimeout bounds the WRITES at the end, on a context of
	// their own. See reportContext.
	codeFeedbackReportTimeout = 60 * time.Second
)

// reportContext returns a fresh context for recording results.
//
// Derived from context.Background rather than from the ingest's context on
// purpose. The ingest context carries the 20-minute analysis budget, and if a
// large repository's checks exhaust it, every write made with that context fails
// — which would mean the one run that most needed a report (the slow, failing
// one) is the one that silently produces none. The findings are the product of
// this feature; they must not be lost to the deadline of the work that found
// them.
func reportContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), codeFeedbackReportTimeout)
}

// codeFeedbackMaxFindings caps each reported bucket.
//
// A repo with two hundred endpoints and a design naming five would otherwise
// produce one blackboard event per unmatched endpoint and flood the amendment
// queue — turning a useful report into something no human reads. The count of
// what was dropped is reported, so the cap is visible rather than silent.
const codeFeedbackMaxFindings = 25

// codeFeedbackOutputLimit truncates a failing command's output before it goes
// into a JSONB event.
const codeFeedbackOutputLimit = 800

// CodeFeedbackService ingests a built repository and turns the gap between it
// and the design into routed questions.
type CodeFeedbackService struct {
	db            *pgxpool.Pool
	store         *blackboard.Store
	sections      *DesignSectionStore
	tokens        RepoTokenSource
	workspaceRoot string
	logger        *zap.Logger
}

// NewCodeFeedbackService wires the service.
func NewCodeFeedbackService(
	db *pgxpool.Pool,
	store *blackboard.Store,
	sections *DesignSectionStore,
	tokens RepoTokenSource,
	workspaceRoot string,
	logger *zap.Logger,
) *CodeFeedbackService {
	return &CodeFeedbackService{
		db:            db,
		store:         store,
		sections:      sections,
		tokens:        tokens,
		workspaceRoot: workspaceRoot,
		logger:        logger,
	}
}

// CodeFeedbackRequest is one ingest.
type CodeFeedbackRequest struct {
	WorkflowID uuid.UUID
	ClientID   uuid.UUID
	Provider   string
	RepoURL    string // the repo the client built — client-supplied, validated
	Branch     string // branch to read; empty means the remote's default
}

// acFinding is one acceptance criterion that did not come back clean.
type acFinding struct {
	ID      string `json:"id"`
	Section string `json:"section"`
	Owner   string `json:"owner"`
	Verify  string `json:"verify,omitempty"`
	Reason  string `json:"reason,omitempty"`
	Output  string `json:"output,omitempty"`
}

// codeFeedbackReport is the whole DIFF result — what goes on the blackboard and
// what the amendments are generated from.
type codeFeedbackReport struct {
	Provider      string         `json:"provider"`
	RepoURL       string         `json:"repo_url"`
	Branch        string         `json:"branch"`
	CommitSHA     string         `json:"commit_sha"`
	CriteriaTotal int            `json:"criteria_total"`
	Satisfied     []string       `json:"satisfied"`
	Unsatisfied   []acFinding    `json:"unsatisfied"`
	Unverifiable  []acFinding    `json:"unverifiable"`
	MissingInCode []contractItem `json:"missing_in_code"`
	NotInDesign   []contractItem `json:"not_in_design"`
	FilesScanned  int            `json:"files_scanned"`
	Dropped       map[string]int `json:"dropped,omitempty"`
	Notes         []string       `json:"notes,omitempty"`
}

// Start validates everything that can be checked cheaply, then runs the ingest
// in the background.
//
// Split this way because a clone plus a test suite is minutes of work and an
// HTTP request is not the place for it. The precedent is repo.Service.SyncRepo,
// which returns as soon as the request is known to be valid and does the fetch
// in a goroutine; the result surfaces the same way everything else in a
// workflow surfaces — as blackboard events the UI already streams.
//
// The synchronous part is chosen so that every error a CLIENT can fix is
// returned to the client: not your workflow, unsupported remote, no harness, no
// criteria, no provider connected. Only errors the client cannot act on (the
// clone failed, a test crashed) land on the blackboard instead.
func (s *CodeFeedbackService) Start(ctx context.Context, req CodeFeedbackRequest) error {
	projectID, err := workflowProject(ctx, s.db, req.WorkflowID, req.ClientID)
	if err != nil {
		return err
	}

	// Host allowlist before the token is read — client_repo.go rule 1.
	target, err := parseRemoteTarget(req.Provider, req.RepoURL)
	if err != nil {
		return err
	}
	if branch := strings.TrimSpace(req.Branch); branch != "" {
		if err := validateBranchName(branch); err != nil {
			return err
		}
	}

	mainPath := mainWorkspacePath(s.workspaceRoot, req.WorkflowID)
	acceptanceRaw, err := os.ReadFile(filepath.Join(mainPath, acceptanceMD))
	if err != nil {
		return fmt.Errorf("%w (no %s to compare the code against)", ErrNoHarness, acceptanceMD)
	}
	criteria := parseAcceptance(string(acceptanceRaw))
	if len(criteria) == 0 {
		return fmt.Errorf("%s has no acceptance criteria yet, so there is nothing to check the code against", acceptanceMD)
	}

	token, err := clientToken(ctx, s.tokens, projectID, target.Provider)
	if err != nil {
		return err
	}

	s.logger.Info("code feedback ingest starting",
		zap.String("workflow_id", req.WorkflowID.String()),
		zap.String("repo", target.CleanURL()),
		zap.Int("criteria", len(criteria)),
	)

	// context.Background, not ctx: ctx dies with the HTTP request, and this
	// work outlives it by design. Same choice repo.Service.SyncRepo makes.
	go func() {
		bg, cancel := context.WithTimeout(context.Background(), codeFeedbackTotalTimeout)
		defer cancel()
		s.ingest(bg, req, target, token, mainPath, criteria)
	}()
	return nil
}

// ingest is the detached body: clone, diff, report, amend.
func (s *CodeFeedbackService) ingest(
	ctx context.Context,
	req CodeFeedbackRequest,
	target remoteTarget,
	token string,
	mainPath string,
	criteria []acceptanceCriterion,
) {
	tempRoot, err := os.MkdirTemp("", "codefeedback-")
	if err != nil {
		s.failIngest(req, target, fmt.Errorf("create temp directory: %w", err))
		return
	}
	// The clone is deleted whatever happens. It holds a copy of the client's
	// source; leaving it in the container after the report is written is not
	// something to depend on a later cleanup for.
	defer func() {
		if rmErr := os.RemoveAll(tempRoot); rmErr != nil {
			s.logger.Warn("code feedback: could not remove clone",
				zap.String("path", tempRoot), zap.Error(rmErr))
		}
	}()
	repoPath := filepath.Join(tempRoot, "repo")

	sha, branch, err := s.cloneClientRepo(ctx, target, token, req.Branch, repoPath)
	if err != nil {
		s.failIngest(req, target, err)
		return
	}

	report := codeFeedbackReport{
		Provider:      target.Provider,
		RepoURL:       target.CleanURL(),
		Branch:        branch,
		CommitSHA:     sha,
		Dropped:       map[string]int{},
		CriteriaTotal: len(criteria),
	}

	// --- DIFF part 1: every criterion's own Verify command ---
	for _, ac := range criteria {
		argv, reason := splitVerifyCommand(ac.Verify)
		if reason != "" {
			report.Unverifiable = append(report.Unverifiable, acFinding{
				ID: ac.ID, Section: ac.SectionPath, Owner: ac.Owner, Verify: ac.Verify, Reason: reason,
			})
			continue
		}

		// runProjectChecks (verify.go) unchanged: it already separates "the
		// check ran and failed" from "the check could not run", already refuses
		// to treat a missing toolchain as a code failure, and already reports
		// the command with its output. Passing one command and using the AC id
		// as the "project" name is all this needs from it — the reason §17.3
		// lists verify.go as the piece to reuse rather than a thing to rebuild.
		cmdCtx, cancel := context.WithTimeout(ctx, codeFeedbackVerifyTimeout)
		result := runProjectChecks(cmdCtx, repoPath, ac.ID, [][]string{argv})
		cancel()

		switch result.Status {
		case VerifyPassed:
			report.Satisfied = append(report.Satisfied, ac.ID)
		case VerifyFailed:
			report.Unsatisfied = append(report.Unsatisfied, acFinding{
				ID: ac.ID, Section: ac.SectionPath, Owner: ac.Owner, Verify: ac.Verify,
				Output: truncateForEvent(result.Output, codeFeedbackOutputLimit),
			})
		default:
			report.Unverifiable = append(report.Unverifiable, acFinding{
				ID: ac.ID, Section: ac.SectionPath, Owner: ac.Owner, Verify: ac.Verify,
				Reason: result.Reason,
			})
		}

		if ctx.Err() != nil {
			report.Notes = append(report.Notes,
				"the ingest ran out of time part-way through the acceptance checks; the remaining criteria were not run")
			break
		}
	}

	// --- DIFF part 2: what the design promises against what the code declares ---
	sectionPaths := s.sectionPaths(ctx, req.WorkflowID)
	designItems, err := readDesignContract(mainPath, sectionPaths)
	if err != nil {
		s.failIngest(req, target, err)
		return
	}
	codeItems, filesScanned, err := scanRepoContract(repoPath)
	if err != nil {
		s.failIngest(req, target, err)
		return
	}
	report.FilesScanned = filesScanned
	missing, undesigned := diffContract(designItems, codeItems)
	var droppedMissing, droppedUndesigned int
	report.MissingInCode, droppedMissing = capItems(missing, codeFeedbackMaxFindings)
	report.NotInDesign, droppedUndesigned = capItems(undesigned, codeFeedbackMaxFindings)
	// Only recorded when something was actually dropped, so a clean report does
	// not carry two zeroes the reader has to interpret.
	if droppedMissing > 0 {
		report.Dropped["missing_in_code"] = droppedMissing
	}
	if droppedUndesigned > 0 {
		report.Dropped["not_in_design"] = droppedUndesigned
	}

	if filesScanned == 0 {
		report.Notes = append(report.Notes,
			"no source files were read in the built repository — check the branch, or whether the code was pushed")
	}
	if len(designItems) == 0 {
		report.Notes = append(report.Notes,
			"the design names no endpoints or tables in a form this comparison can read, so only the acceptance criteria were checked")
	}

	// --- REPORT and AMEND, on their own context ---
	writeCtx, cancelWrite := reportContext()
	defer cancelWrite()

	reportEventID, err := s.postReport(writeCtx, req.WorkflowID, report)
	if err != nil {
		s.logger.Error("code feedback: could not post the report",
			zap.String("workflow_id", req.WorkflowID.String()), zap.Error(err))
		return
	}

	posted := s.postAmendments(writeCtx, req.WorkflowID, reportEventID, report)

	s.logger.Info("code feedback ingest complete",
		zap.String("workflow_id", req.WorkflowID.String()),
		zap.String("commit", sha),
		zap.Int("satisfied", len(report.Satisfied)),
		zap.Int("unsatisfied", len(report.Unsatisfied)),
		zap.Int("unverifiable", len(report.Unverifiable)),
		zap.Int("missing_in_code", len(report.MissingInCode)),
		zap.Int("not_in_design", len(report.NotInDesign)),
		zap.Int("amendments_proposed", posted),
	)
}

// cloneClientRepo makes a shallow clone of the built repo and returns its HEAD
// commit and the branch that was read.
//
// --depth 1 --single-branch because nothing here reads history: the comparison
// is against the current state of one branch. On a large repository that is the
// difference between seconds and minutes.
//
// THE REMOTE IS REMOVED IMMEDIATELY. Unlike push, `git clone` PERSISTS the URL
// it was given as remote.origin.url — which means the token would sit in
// .git/config for as long as the clone exists, and any Verify command running
// `git remote -v` inside it would print it. Removing origin costs one command
// and is then asserted, not assumed.
func (s *CodeFeedbackService) cloneClientRepo(
	ctx context.Context, target remoteTarget, token, wantBranch, dest string,
) (sha string, branch string, err error) {
	cloneCtx, cancel := context.WithTimeout(ctx, codeFeedbackCloneTimeout)
	defer cancel()

	args := []string{"clone", "--depth", "1", "--single-branch"}
	if b := strings.TrimSpace(wantBranch); b != "" {
		args = append(args, "--branch", b)
	}
	args = append(args, target.authedURL(token), dest)

	out, cloneErr := gitRun(cloneCtx, "", args...)
	clean := strings.TrimSpace(scrubSecrets(out, token))

	// Runs even on failure: a failed clone can still have created the directory
	// and written the config.
	if _, rmErr := gitRun(ctx, dest, "remote", "remove", "origin"); rmErr != nil && cloneErr == nil {
		s.logger.Warn("code feedback: could not remove the clone's origin remote",
			zap.String("repo", target.CleanURL()))
	}
	if diskErr := assertNoSecretOnDisk(dest, token); diskErr != nil {
		return "", "", diskErr
	}

	if cloneErr != nil {
		lower := strings.ToLower(clean)
		switch {
		case strings.Contains(lower, "authentication failed") || strings.Contains(lower, "could not read username"):
			return "", "", fmt.Errorf("the connected %s account cannot read %s — reconnect the provider. (git: %s)",
				target.Provider, target.CleanURL(), clean)
		case strings.Contains(lower, "not found") || strings.Contains(lower, "repository not found"):
			return "", "", fmt.Errorf("%s does not exist or is not visible to the connected %s account. (git: %s)",
				target.CleanURL(), target.Provider, clean)
		case strings.Contains(lower, "remote branch") && strings.Contains(lower, "not found"):
			return "", "", fmt.Errorf("branch %q does not exist in %s. (git: %s)", wantBranch, target.CleanURL(), clean)
		default:
			return "", "", fmt.Errorf("clone of %s failed: %w (git: %s)", target.CleanURL(), cloneErr, clean)
		}
	}

	sha, err = headSHA(ctx, dest)
	if err != nil {
		return "", "", err
	}
	branch, err = currentBranch(ctx, dest)
	if err != nil {
		// A detached HEAD is possible if the caller named a tag. The comparison
		// works regardless, so this is reported, not fatal.
		branch = strings.TrimSpace(wantBranch)
	}
	return sha, branch, nil
}

// postReport writes the one code_feedback_ingested event.
//
// PostedByClient because the client triggered the ingest and no expert authored
// it; migration 006's blackboard_events_poster_check requires exactly one of
// posted_by_expert_id / posted_by_client to be set.
func (s *CodeFeedbackService) postReport(ctx context.Context, workflowID uuid.UUID, report codeFeedbackReport) (uuid.UUID, error) {
	ev, err := s.store.Post(ctx, blackboard.PostRequest{
		WorkflowID:     workflowID,
		EventType:      "code_feedback_ingested",
		PostedByClient: true,
		Content:        report,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("post code_feedback_ingested: %w", err)
	}
	return ev.ID, nil
}

// postAmendments turns each finding into one routed code_feedback_question
// event. Returns how many were posted.
//
// Every event references the report event, so the amendment queue can always
// answer "which build produced this question" — the traceability §17.2 step 6
// asks DECISIONS.md for, available immediately rather than after the apply path
// is built.
func (s *CodeFeedbackService) postAmendments(
	ctx context.Context, workflowID, reportEventID uuid.UUID, report codeFeedbackReport,
) int {
	owners := s.sectionOwners(ctx, workflowID)
	integrator := s.resolveIntegrator(ctx, workflowID)

	type amendment struct {
		finding  string
		target   string
		subject  string
		question string
		detail   string
		routeTo  *uuid.UUID
	}
	var queue []amendment

	// An unsatisfied criterion is the client's own definition of done coming
	// back failed. The owning expert decides whether it still stands.
	for _, f := range report.Unsatisfied {
		owner := owners[f.Section]
		if owner == nil {
			owner = integrator
		}
		queue = append(queue, amendment{
			finding: "acceptance_unsatisfied",
			target:  acceptanceMD,
			subject: f.ID,
			question: fmt.Sprintf("%s did not pass in the built code. Is this criterion still required "+
				"(the code is incomplete), or should it be dropped or rewritten (the criterion is wrong)?", f.ID),
			detail:  f.Output,
			routeTo: owner,
		})
	}

	// Designed but not present in the code.
	for _, item := range report.MissingInCode {
		owner := owners[item.Source]
		if owner == nil {
			owner = integrator
		}
		queue = append(queue, amendment{
			finding: "contract_missing_in_code",
			target:  item.Source,
			subject: item.Key,
			question: fmt.Sprintf("the design names the %s %q but the built code does not appear to declare it. "+
				"Is it still required, or should the design drop it?", item.Kind, item.Key),
			routeTo: owner,
		})
	}

	// Built but never designed — §17.5's "invented". Routed to the integrator,
	// not to a section owner: deciding whether the contract adopts something is
	// a cross-cutting call, and by §16.6 the integrator is who makes it.
	for _, item := range report.NotInDesign {
		queue = append(queue, amendment{
			finding: "contract_not_in_design",
			target:  finalMD,
			subject: item.Key,
			question: fmt.Sprintf("the built code declares the %s %q, which the design never defined. "+
				"Should the contract adopt it, or is it a defect to remove?", item.Kind, item.Key),
			detail:  "found in " + item.Source,
			routeTo: integrator,
		})
	}

	posted := 0
	for _, a := range queue {
		if a.routeTo == nil {
			// No expert to ask. Reporting it as an unrouted amendment would put
			// a question in the queue that nobody owns, so it is logged and
			// left in the report instead, where the client can still see it.
			s.logger.Warn("code feedback: finding has no expert to route to",
				zap.String("workflow_id", workflowID.String()),
				zap.String("finding", a.finding),
				zap.String("subject", a.subject))
			continue
		}
		content := map[string]any{
			"kind":        "code_feedback",
			"finding":     a.finding,
			"target":      a.target,
			"subject":     a.subject,
			"question":    a.question,
			"repo_url":    report.RepoURL,
			"commit_sha":  report.CommitSHA,
			"routed_to":   a.routeTo.String(),
			"needs_reply": true,
		}
		if a.detail != "" {
			content["detail"] = a.detail
		}
		// code_feedback_question, NOT design_amendment_proposed.
		//
		// They were the same event type at first, and that conflated two things
		// that need different actions from different people. A
		// design_amendment_proposed now carries a concrete {old_text, new_text}
		// and a pending approval row: the CLIENT approves it and it is written
		// (§7.5, amendment.go). A finding here has no such text — it is a
		// question an EXPERT has to answer ("still required, or drop it?"), and
		// the expert's answer is what then becomes a real amendment through
		// propose_amendment. One event type for both would leave the UI unable
		// to tell "you must approve this" from "an expert must answer this", and
		// would put un-approvable rows in the amendments list.
		if _, err := s.store.Post(ctx, blackboard.PostRequest{
			WorkflowID:         workflowID,
			EventType:          "code_feedback_question",
			PostedByClient:     true,
			ToExpertID:         a.routeTo,
			Content:            content,
			ReferencesEventIDs: []uuid.UUID{reportEventID},
		}); err != nil {
			s.logger.Error("code feedback: could not post an amendment",
				zap.String("workflow_id", workflowID.String()),
				zap.String("subject", a.subject), zap.Error(err))
			continue
		}
		posted++
	}
	return posted
}

// failIngest records a failure the client cannot fix from the request itself.
//
// Posted rather than only logged: the ingest is detached from the HTTP request,
// so a log line is invisible to the person who asked for it. The blackboard is
// where everything else about a workflow already surfaces.
//
// Takes no context, and uses reportContext instead. The most common reason to be
// here is that the ingest context expired — writing the failure with that same
// expired context would fail too, and the client would be left watching an
// ingest that never says anything at all.
func (s *CodeFeedbackService) failIngest(req CodeFeedbackRequest, target remoteTarget, cause error) {
	s.logger.Error("code feedback ingest failed",
		zap.String("workflow_id", req.WorkflowID.String()),
		zap.String("repo", target.CleanURL()),
		zap.Error(cause))

	ctx, cancel := reportContext()
	defer cancel()

	if _, err := s.store.Post(ctx, blackboard.PostRequest{
		WorkflowID:     req.WorkflowID,
		EventType:      "code_feedback_failed",
		PostedByClient: true,
		Content: map[string]any{
			"provider": target.Provider,
			"repo_url": target.CleanURL(),
			"branch":   req.Branch,
			"error":    cause.Error(),
		},
	}); err != nil {
		s.logger.Error("code feedback: could not record the failure either",
			zap.String("workflow_id", req.WorkflowID.String()), zap.Error(err))
	}
}

// ============================================================
// Routing helpers
// ============================================================

// sectionPaths returns the assigned section files for a workflow.
func (s *CodeFeedbackService) sectionPaths(ctx context.Context, workflowID uuid.UUID) []string {
	secs, err := s.sections.ListSections(ctx, workflowID)
	if err != nil {
		s.logger.Warn("code feedback: could not list design sections, comparing against the spine only",
			zap.String("workflow_id", workflowID.String()), zap.Error(err))
		return nil
	}
	out := make([]string, 0, len(secs))
	for _, sec := range secs {
		out = append(out, sec.SectionPath)
	}
	return out
}

// sectionOwners maps a section file path to the expert that owns it.
//
// Keyed on the path rather than on the owner name in ACCEPTANCE.md because the
// path is machine-generated (DesignSectionStore builds it from experts.slug) and
// the name in the markdown is not: acceptanceLineRe captures owner as \S+, so a
// two-word expert name arrives truncated. Matching on a truncated name would
// route a question to the wrong expert, which is worse than not routing it.
func (s *CodeFeedbackService) sectionOwners(ctx context.Context, workflowID uuid.UUID) map[string]*uuid.UUID {
	owners := map[string]*uuid.UUID{}
	secs, err := s.sections.ListSections(ctx, workflowID)
	if err != nil {
		s.logger.Warn("code feedback: could not list design sections for routing",
			zap.String("workflow_id", workflowID.String()), zap.Error(err))
		return owners
	}
	for _, sec := range secs {
		id := sec.ExpertID
		owners[sec.SectionPath] = &id
	}
	return owners
}

// resolveIntegrator returns the expert that answers cross-cutting questions,
// per §16.6: a manual override if the workflow sets one, otherwise the
// most-upstream participant.
//
// "Most upstream" is read as the lowest expert_categories.authoring_rank among
// the workflow's participants, with NULL ranks last, breaking ties on
// section_no (first to author). §16.2 defines authoring_rank as the ordering
// hint, and §19 records that ranks are unseeded today — so in practice this
// falls through to "the first expert that authored", which is the safe answer
// and is exactly what §19 says the consequence of unseeded ranks is.
//
// Returns nil when the workflow has no sections at all. Callers must handle it;
// posting a routed event with a nil expert would violate
// blackboard_events.to_expert_id's foreign key.
func (s *CodeFeedbackService) resolveIntegrator(ctx context.Context, workflowID uuid.UUID) *uuid.UUID {
	var auto bool
	var manual *uuid.UUID
	err := s.db.QueryRow(ctx,
		`SELECT integrator_auto, integrator_expert_id FROM workflows WHERE id = $1`,
		workflowID,
	).Scan(&auto, &manual)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		s.logger.Warn("code feedback: could not read the integrator setting",
			zap.String("workflow_id", workflowID.String()), zap.Error(err))
	}
	if !auto && manual != nil {
		return manual
	}

	var expertID uuid.UUID
	err = s.db.QueryRow(ctx,
		`SELECT s.expert_id
		   FROM workflow_design_sections s
		   JOIN experts e ON e.id = s.expert_id
		   LEFT JOIN expert_categories c ON c.id = e.category_id
		  WHERE s.workflow_id = $1
		  ORDER BY c.authoring_rank ASC NULLS LAST, s.section_no ASC
		  LIMIT 1`,
		workflowID,
	).Scan(&expertID)
	if err != nil {
		// No sections yet is an expected state, not a failure worth logging.
		if !errors.Is(err, pgx.ErrNoRows) {
			s.logger.Warn("code feedback: could not resolve an integrator",
				zap.String("workflow_id", workflowID.String()), zap.Error(err))
		}
		return nil
	}
	return &expertID
}

// ============================================================
// Small shared helpers
// ============================================================

// capItems returns at most limit items and how many were dropped.
func capItems(items []contractItem, limit int) ([]contractItem, int) {
	if len(items) <= limit {
		return items, 0
	}
	return items[:limit], len(items) - limit
}

// truncateForEvent bounds a string that is about to be stored in a JSONB
// column.
//
// Not planner.go's truncate, and not named truncate either — that name is
// already taken in this package, which is how this function was caught before it
// ever compiled. Nor is planner's reused: it cuts silently, and a command's
// output stored in an event needs to say that it was cut, or a reader debugging
// a failed criterion will take a half-printed stack trace at face value.
func truncateForEvent(s string, limit int) string {
	s = strings.TrimSpace(s)
	if len(s) <= limit {
		return s
	}
	return s[:limit] + "\n...(truncated)"
}
