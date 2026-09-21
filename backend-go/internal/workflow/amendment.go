package workflow

// §7.5 of docs/COLLABORATIVE_DESIGN_ARCHITECTURE.md — propose, approve, apply.
//
// Before this file, every mutating tool in the chat (§7.4) and every code
// feedback finding (§17) posted a proposal event and stopped there. The design
// could be discussed and questioned but never actually changed, which made the
// whole §7 tool loop a very well-instrumented read-only system.
//
// THE FLOW, AND WHERE EACH STEP LIVES
//
//	1 PROPOSE  proposeAmendment          — blackboard event + approval_requests row
//	2 APPROVE  AmendmentService.Respond  — the client's decision, idempotent
//	3 APPLY    applyAmendmentText + git  — deterministic text edit, one commit
//	4 RECORD   design_amended event      — commit sha, approval id, who approved
//	5 MERGE    (not needed — see below)
//	6 LOG      DECISIONS.md, newest first
//
// TWO PLACES THIS DEPARTS FROM THE DOC, DELIBERATELY
//
// **No Aider session.** §7.5 step 3 says "one Aider session on branch
// amend/{chat_id}/{n}". An amendment is a structured {target, old_text,
// new_text}: applying it is an exact string replacement. Handing that to an LLM
// buys nothing and costs three things we have already paid for on the Aider
// path — non-determinism (it may rewrite text nobody asked it to), money per
// amendment, and the failure mode where the model proposes no edits at all and
// the run silently produces nothing (fixed twice already: 1faf3e5, d9fa1c3).
// applyAmendmentText does the same job with a defined answer for every input,
// including refusing to guess when old_text is ambiguous.
//
// **No branch.** §7.5 gave the branch two jobs. The first, "gives us the diff to
// render at step 2", is already satisfied: old_text/new_text IS the diff, stored
// structurally in artifact_content, and it is what the approval screen renders.
// The second, "keeps the proposal out of main/ until approved", is satisfied by
// writing nothing at all until approval — which is the actual behaviour here. A
// branch would add a real hazard in exchange: main/ is the shared workspace that
// a running wave rsyncs into, and a `git checkout` there would swap files under
// a wave in progress. Step 5 disappears for the same reason — there is nothing
// to merge back.
//
// WHAT IS STILL A LIMITATION, NAMED
//
// An approved amendment is refused while the workflow is actively running. A
// wave merge rsyncs each expert workspace over main/, and those workspaces were
// seeded from main/ BEFORE the amendment, so a merge landing after an apply
// would silently overwrite it. Refusing with "the design is being written right
// now, try again in a moment" is honest; making the merger amendment-aware is
// the real fix and is a larger change than this one.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/blackboard"
)

// amendmentGate is the approval_requests.gate_name for an amendment.
//
// Already legal: migration 018 widened approval_requests_gate_check to include
// 'design_amendment' and 'design_conflict'. No migration is needed here, which
// is why 018 widened it — this is the code that was missing.
const amendmentGate = "design_amendment"

// Amendment kinds. The kind decides how the proposal is rendered and nothing
// about how it is applied: every kind ends up as the same text operation.
const (
	amendmentKindStatement  = "statement"  // a change to existing design text
	amendmentKindAcceptance = "acceptance" // a new acceptance criterion
)

// Amendment errors.
var (
	// ErrAmendmentNotFound covers "no such amendment" and "belongs to another
	// client", merged for the same reason ErrWorkflowNotFound merges them.
	ErrAmendmentNotFound = errors.New("amendment not found")

	// ErrAmendmentResolved means someone already decided this one. Distinct so
	// the handler can answer 409 rather than 400 — a duplicate click is not a
	// bad request.
	ErrAmendmentResolved = errors.New("this amendment has already been decided")

	// ErrDesignBusy means a wave is writing the design right now. See the
	// limitation named in the file comment.
	ErrDesignBusy = errors.New("the design is being written right now — approve this again in a moment")

	// ErrAmendmentStale is declared in amendment_text.go, next to the function
	// that returns it.

	// ErrBadAmendmentTarget means the target file is not part of this
	// workflow's harness.
	ErrBadAmendmentTarget = errors.New("that file is not part of this workflow's design")
)

// ============================================================
// 1. PROPOSE
// ============================================================

// ProposeAmendmentRequest is one proposal.
type ProposeAmendmentRequest struct {
	WorkflowID uuid.UUID
	ExpertID   uuid.UUID
	ExpertName string
	// ChatID is uuid.Nil when the proposal did not come from a chat.
	ChatID  uuid.UUID
	Kind    string
	Target  string
	OldText string // empty means "append New Text to the file"
	NewText string
	Reason  string
}

// amendmentArtifact is what goes into approval_requests.artifact_content. It is
// the whole proposal: everything needed to render the diff for the client and
// to apply it later, with no second lookup.
type amendmentArtifact struct {
	Kind               string `json:"kind"`
	Target             string `json:"target"`
	OldText            string `json:"old_text"`
	NewText            string `json:"new_text"`
	Reason             string `json:"reason"`
	ProposedByExpertID string `json:"proposed_by_expert_id,omitempty"`
	ProposedByExpert   string `json:"proposed_by_expert,omitempty"`
	ChatID             string `json:"chat_id,omitempty"`
	ProposalEventID    string `json:"proposal_event_id"`
}

// ProposedAmendment is what a proposal returns to the tool that made it.
type ProposedAmendment struct {
	ApprovalID uuid.UUID
	EventID    uuid.UUID
}

// proposeAmendment records a proposal: one blackboard event, one pending
// approval row. Writes nothing to the design.
//
// Order is event-then-approval, the reverse of Tools.AskClient. AskClient
// creates the approval first so the event can carry approval_id, because its
// client-facing path is an SSE event the UI reads that id out of. An amendment
// is not delivered that way — it is read from the amendments list endpoint — so
// the more useful link is the other direction, and approval_requests already has
// the column for it: cited_event_ids. Using it here is also the first real
// exercise of that column, which is how the uuid[] literal bug it had was found.
//
// A package-level function, not a method: the two callers are tool handlers
// whose toolLoopContext already carries a pool and a store, and giving them a
// service dependency would mean threading it through NewToolRegistry for no
// gain.
func proposeAmendment(
	ctx context.Context, db *pgxpool.Pool, store *blackboard.Store, req ProposeAmendmentRequest,
) (*ProposedAmendment, error) {
	if req.Target == "" || req.NewText == "" || req.Reason == "" {
		return nil, fmt.Errorf("amendment: target, new text and reason are all required")
	}
	if req.Kind == "" {
		req.Kind = amendmentKindStatement
	}

	content := map[string]any{
		"kind":     req.Kind,
		"target":   req.Target,
		"old_text": req.OldText,
		"new_text": req.NewText,
		"reason":   req.Reason,
	}
	if req.ChatID != uuid.Nil {
		content["chat_id"] = req.ChatID.String()
	}

	post := blackboard.PostRequest{
		WorkflowID: req.WorkflowID,
		EventType:  "design_amendment_proposed",
		Content:    content,
	}
	if req.ExpertID != uuid.Nil {
		expertID := req.ExpertID
		post.PostedByExpertID = &expertID
	} else {
		// blackboard_events_poster_check (migration 006) requires exactly one of
		// posted_by_expert_id / posted_by_client.
		post.PostedByClient = true
	}

	ev, err := store.Post(ctx, post)
	if err != nil {
		return nil, fmt.Errorf("amendment: post proposal event: %w", err)
	}

	artifact := amendmentArtifact{
		Kind:             req.Kind,
		Target:           req.Target,
		OldText:          req.OldText,
		NewText:          req.NewText,
		Reason:           req.Reason,
		ProposedByExpert: req.ExpertName,
		ProposalEventID:  ev.ID.String(),
	}
	if req.ExpertID != uuid.Nil {
		artifact.ProposedByExpertID = req.ExpertID.String()
	}
	if req.ChatID != uuid.Nil {
		artifact.ChatID = req.ChatID.String()
	}

	approvalID, err := createApprovalRequest(ctx, db, AskClientRequest{
		WorkflowID:      req.WorkflowID,
		GateName:        amendmentGate,
		Summary:         amendmentSummary(req),
		ArtifactContent: artifact,
		CitedEventIDs:   []uuid.UUID{ev.ID},
	})
	if err != nil {
		// The event is already on the blackboard. That is acceptable and not
		// worth compensating for: the event log is append-only by design, the
		// tool returns the error, and a retry with identical content is
		// suppressed by blackboard.Post's dedup_key — so a retry produces one
		// event and one approval, not two of either.
		return nil, fmt.Errorf("amendment: create approval: %w", err)
	}

	return &ProposedAmendment{ApprovalID: approvalID, EventID: ev.ID}, nil
}

// amendmentSummary is the one line the client sees in a list.
//
// Built from the proposal rather than asked for separately: another required
// argument on the tool is another thing a model can fill with a restatement of
// the reason.
func amendmentSummary(req ProposeAmendmentRequest) string {
	who := req.ExpertName
	if who == "" {
		who = "an expert"
	}
	switch req.Kind {
	case amendmentKindAcceptance:
		return fmt.Sprintf("%s proposes a new acceptance criterion in %s", who, req.Target)
	default:
		if req.OldText == "" {
			return fmt.Sprintf("%s proposes an addition to %s", who, req.Target)
		}
		return fmt.Sprintf("%s proposes a change to %s", who, req.Target)
	}
}

// ============================================================
// 2. APPROVE and 3. APPLY
// ============================================================

// AmendmentService reads and resolves amendment approvals, and applies the
// approved ones to the harness.
type AmendmentService struct {
	db            *pgxpool.Pool
	store         *blackboard.Store
	sections      *DesignSectionStore
	workspaceRoot string
	logger        *zap.Logger

	// applyMu serialises read-modify-write of the harness files.
	//
	// Two amendments approved at the same moment both read final.md, both edit
	// their own copy, and the second write erases the first. A mutex is the
	// right tool for the actual deployment: one api process owns one workspace
	// directory, because the workspace is a local volume, not shared storage.
	// If that ever stops being true, this needs a Postgres advisory lock — named
	// here so the assumption is visible rather than discovered.
	applyMu sync.Mutex
}

// NewAmendmentService wires the service.
func NewAmendmentService(
	db *pgxpool.Pool,
	store *blackboard.Store,
	sections *DesignSectionStore,
	workspaceRoot string,
	logger *zap.Logger,
) *AmendmentService {
	return &AmendmentService{
		db:            db,
		store:         store,
		sections:      sections,
		workspaceRoot: workspaceRoot,
		logger:        logger,
	}
}

// AmendmentView is one amendment as the client sees it.
type AmendmentView struct {
	ApprovalID uuid.UUID `json:"approval_id"`
	Status     string    `json:"status"`
	Summary    string    `json:"summary"`
	Kind       string    `json:"kind"`
	Target     string    `json:"target"`
	// OldText and NewText together are the diff the client approves. Sent in
	// full: an amendment is a change to a contract, and a truncated contract
	// change is not something anyone can responsibly approve.
	OldText          string     `json:"old_text"`
	NewText          string     `json:"new_text"`
	Reason           string     `json:"reason"`
	ProposedByExpert string     `json:"proposed_by_expert,omitempty"`
	ChatID           string     `json:"chat_id,omitempty"`
	RequestedAt      time.Time  `json:"requested_at"`
	RespondedAt      *time.Time `json:"responded_at,omitempty"`
}

// List returns a workflow's amendments, newest first.
//
// status filters on approval_requests.status; empty means "pending", because
// that is the only list anyone acts on and defaulting to everything would put a
// month of resolved history in front of the one decision that is waiting.
func (s *AmendmentService) List(ctx context.Context, workflowID, clientID uuid.UUID, status string) ([]AmendmentView, error) {
	if _, err := workflowProject(ctx, s.db, workflowID, clientID); err != nil {
		return nil, err
	}
	if status == "" {
		status = "pending"
	}

	query := `SELECT id, status, summary, artifact_content, requested_at, responded_at
	            FROM approval_requests
	           WHERE workflow_id = $1 AND gate_name = $2`
	args := []any{workflowID, amendmentGate}
	if status != "all" {
		query += ` AND status = $3`
		args = append(args, status)
	}
	query += ` ORDER BY requested_at DESC`

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list amendments: %w", err)
	}
	defer rows.Close()

	out := []AmendmentView{}
	for rows.Next() {
		var v AmendmentView
		var artifactJSON []byte
		if err := rows.Scan(&v.ApprovalID, &v.Status, &v.Summary, &artifactJSON,
			&v.RequestedAt, &v.RespondedAt); err != nil {
			return nil, fmt.Errorf("list amendments: scan: %w", err)
		}
		var a amendmentArtifact
		// A row whose artifact does not parse is still listed, with the summary
		// and status it has. Dropping it would hide a pending decision.
		if err := json.Unmarshal(artifactJSON, &a); err != nil {
			s.logger.Warn("amendment: artifact_content did not parse",
				zap.String("approval_id", v.ApprovalID.String()), zap.Error(err))
		} else {
			v.Kind, v.Target, v.OldText, v.NewText = a.Kind, a.Target, a.OldText, a.NewText
			v.Reason, v.ProposedByExpert, v.ChatID = a.Reason, a.ProposedByExpert, a.ChatID
		}
		out = append(out, v)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list amendments: rows: %w", err)
	}
	return out, nil
}

// AmendmentDecision is the client's answer.
type AmendmentDecision struct {
	// Decision is "approve", "approve_with_edit" or "reject".
	Decision string
	// EditedText replaces the proposal's new_text when Decision is
	// approve_with_edit. §7.5: "Reject and edit-then-approve are both
	// first-class" — and the existing RespondToApproval endpoint has no way to
	// carry edited content, which is the main reason amendments need their own
	// endpoint rather than reusing that one.
	EditedText string
	Notes      string
}

// AmendmentOutcome is what happened.
type AmendmentOutcome struct {
	ApprovalID uuid.UUID `json:"approval_id"`
	Status     string    `json:"status"`
	Applied    bool      `json:"applied"`
	Target     string    `json:"target,omitempty"`
	CommitSHA  string    `json:"commit_sha,omitempty"`
	DecisionID string    `json:"decision_id,omitempty"`
}

// Respond records the client's decision and, on approval, applies it.
//
// The whole body holds applyMu, not just the write: the check that old_text
// still appears exactly once, and the write that depends on that being true,
// have to be the same critical section. Validating outside the lock would make
// the check advisory — two amendments touching the same paragraph could both
// pass it and the second would then apply to text the first had already changed.
func (s *AmendmentService) Respond(
	ctx context.Context, workflowID, approvalID, clientID uuid.UUID, dec AmendmentDecision,
) (*AmendmentOutcome, error) {
	if _, err := workflowProject(ctx, s.db, workflowID, clientID); err != nil {
		return nil, err
	}

	approve := dec.Decision == "approve" || dec.Decision == "approve_with_edit"
	if !approve && dec.Decision != "reject" {
		return nil, fmt.Errorf("unknown decision %q — use approve, approve_with_edit or reject", dec.Decision)
	}

	s.applyMu.Lock()
	defer s.applyMu.Unlock()

	artifact, err := s.loadPending(ctx, workflowID, approvalID)
	if err != nil {
		return nil, err
	}

	if !approve {
		return s.reject(ctx, workflowID, approvalID, clientID, artifact, dec)
	}

	newText := artifact.NewText
	if dec.Decision == "approve_with_edit" {
		if strings.TrimSpace(dec.EditedText) == "" {
			return nil, fmt.Errorf("approve_with_edit needs edited_text — send approve to accept the proposal as written")
		}
		newText = dec.EditedText
	}

	// A wave rsyncing over main/ would erase this. See the file comment.
	var wfStatus string
	if err := s.db.QueryRow(ctx,
		`SELECT status FROM workflows WHERE id = $1`, workflowID,
	).Scan(&wfStatus); err != nil {
		return nil, fmt.Errorf("amendment: read workflow status: %w", err)
	}
	if wfStatus == StatusRunning {
		return nil, ErrDesignBusy
	}

	targets, err := amendableTargets(ctx, s.sections, workflowID)
	if err != nil {
		return nil, err
	}
	if !targets[artifact.Target] {
		return nil, fmt.Errorf("%w: %s", ErrBadAmendmentTarget, artifact.Target)
	}

	mainPath := mainWorkspacePath(s.workspaceRoot, workflowID)
	targetFull := filepath.Join(mainPath, artifact.Target)
	existing, err := os.ReadFile(targetFull)
	if err != nil {
		return nil, fmt.Errorf("%w (%s is not in the merged workspace)", ErrNoHarness, artifact.Target)
	}

	// Computed BEFORE the status is claimed, so a stale proposal is refused
	// without consuming the approval — the client can still reject it, and the
	// expert can propose again against the current text.
	updated, err := applyAmendmentText(string(existing), artifact.OldText, newText)
	if err != nil {
		return nil, err
	}

	// Claim it. The conditional UPDATE is both the lookup and the resolution,
	// the pattern RespondToApproval already uses (handler.go) — a second click
	// affects zero rows and is reported as a conflict instead of applying twice.
	clientResponse, _ := json.Marshal(map[string]string{
		"decision": dec.Decision,
		"notes":    dec.Notes,
	})
	tag, err := s.db.Exec(ctx,
		`UPDATE approval_requests
		    SET status = 'approved', client_response = $1, responded_at = NOW()
		  WHERE id = $2 AND workflow_id = $3 AND gate_name = $4 AND status = 'pending'`,
		clientResponse, approvalID, workflowID, amendmentGate)
	if err != nil {
		return nil, fmt.Errorf("amendment: claim approval: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, ErrAmendmentResolved
	}

	outcome, applyErr := s.applyApproved(ctx, workflowID, approvalID, clientID, artifact, updated, dec)
	if applyErr != nil {
		// Put it back so the client is not left with an amendment marked
		// approved that never reached the design. The rollback is the only way
		// out of that state — there is no worker that would retry it later.
		if _, rbErr := s.db.Exec(ctx,
			`UPDATE approval_requests SET status = 'pending', client_response = NULL, responded_at = NULL
			  WHERE id = $1 AND workflow_id = $2`, approvalID, workflowID); rbErr != nil {
			s.logger.Error("amendment: apply failed AND the rollback failed — this amendment is marked approved but was not applied",
				zap.String("approval_id", approvalID.String()),
				zap.NamedError("apply_error", applyErr), zap.NamedError("rollback_error", rbErr))
		}
		return nil, applyErr
	}
	return outcome, nil
}

// loadPending reads one pending amendment's artifact.
func (s *AmendmentService) loadPending(ctx context.Context, workflowID, approvalID uuid.UUID) (amendmentArtifact, error) {
	var artifactJSON []byte
	var status string
	err := s.db.QueryRow(ctx,
		`SELECT status, artifact_content FROM approval_requests
		  WHERE id = $1 AND workflow_id = $2 AND gate_name = $3`,
		approvalID, workflowID, amendmentGate,
	).Scan(&status, &artifactJSON)
	if errors.Is(err, pgx.ErrNoRows) {
		return amendmentArtifact{}, ErrAmendmentNotFound
	}
	if err != nil {
		return amendmentArtifact{}, fmt.Errorf("amendment: load: %w", err)
	}
	if status != "pending" {
		return amendmentArtifact{}, ErrAmendmentResolved
	}

	var a amendmentArtifact
	if err := json.Unmarshal(artifactJSON, &a); err != nil {
		return amendmentArtifact{}, fmt.Errorf("amendment: artifact_content did not parse: %w", err)
	}
	if a.Target == "" || a.NewText == "" {
		return amendmentArtifact{}, fmt.Errorf("amendment: artifact_content has no target or no new text")
	}
	return a, nil
}

// reject resolves an amendment without writing anything.
func (s *AmendmentService) reject(
	ctx context.Context, workflowID, approvalID, clientID uuid.UUID,
	artifact amendmentArtifact, dec AmendmentDecision,
) (*AmendmentOutcome, error) {
	clientResponse, _ := json.Marshal(map[string]string{"decision": "reject", "notes": dec.Notes})
	tag, err := s.db.Exec(ctx,
		`UPDATE approval_requests
		    SET status = 'rejected', client_response = $1, responded_at = NOW()
		  WHERE id = $2 AND workflow_id = $3 AND gate_name = $4 AND status = 'pending'`,
		clientResponse, approvalID, workflowID, amendmentGate)
	if err != nil {
		return nil, fmt.Errorf("amendment: reject: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, ErrAmendmentResolved
	}

	refs := citedProposalEvents(artifact)
	if _, err := s.store.Post(ctx, blackboard.PostRequest{
		WorkflowID:     workflowID,
		EventType:      "design_amendment_rejected",
		PostedByClient: true,
		Content: map[string]any{
			"approval_id": approvalID.String(),
			"target":      artifact.Target,
			"reason":      artifact.Reason,
			"notes":       dec.Notes,
			"rejected_by": clientID.String(),
		},
		ReferencesEventIDs: refs,
	}); err != nil {
		// The decision is recorded in approval_requests either way; the event is
		// the audit copy. Losing it must not tell the client their rejection
		// failed, because it did not.
		s.logger.Error("amendment: rejected, but the event could not be recorded",
			zap.String("approval_id", approvalID.String()), zap.Error(err))
	}

	return &AmendmentOutcome{
		ApprovalID: approvalID, Status: "rejected", Applied: false, Target: artifact.Target,
	}, nil
}

// applyApproved writes the amended file, appends the decision log, commits, and
// records design_amended.
func (s *AmendmentService) applyApproved(
	ctx context.Context, workflowID, approvalID, clientID uuid.UUID,
	artifact amendmentArtifact, updatedContent string, dec AmendmentDecision,
) (*AmendmentOutcome, error) {
	mainPath := mainWorkspacePath(s.workspaceRoot, workflowID)
	if err := ensureGitRepo(ctx, mainPath); err != nil {
		return nil, fmt.Errorf("amendment: prepare workspace: %w", err)
	}

	if err := writeWorkspaceFile(mainPath, artifact.Target, updatedContent); err != nil {
		return nil, fmt.Errorf("amendment: write %s: %w", artifact.Target, err)
	}

	// §7.5 step 6. Written here and not by an expert turn: the decision log
	// records the CLIENT's decision, and no expert is in a position to author it.
	decisionID, err := s.appendDecision(mainPath, approvalID, clientID, artifact, dec)
	if err != nil {
		return nil, fmt.Errorf("amendment: append decision log: %w", err)
	}

	message := fmt.Sprintf("design(amend): %s — %s\n\nApproval: %s\nDecision: %s",
		artifact.Target, firstLine(artifact.Reason), approvalID, decisionID)
	if err := gitAddCommit(ctx, mainPath, message); err != nil {
		return nil, fmt.Errorf("amendment: commit: %w", err)
	}
	sha, err := headSHA(ctx, mainPath)
	if err != nil {
		return nil, fmt.Errorf("amendment: %w", err)
	}

	content := map[string]any{
		"approval_id": approvalID.String(),
		"target":      artifact.Target,
		"kind":        artifact.Kind,
		"reason":      artifact.Reason,
		"commit_sha":  sha,
		"decision_id": decisionID,
		"decision":    dec.Decision,
		"approved_by": clientID.String(),
		"notes":       dec.Notes,
	}
	if artifact.ChatID != "" {
		content["chat_id"] = artifact.ChatID
	}
	if _, err := s.store.Post(ctx, blackboard.PostRequest{
		WorkflowID:         workflowID,
		EventType:          "design_amended",
		PostedByClient:     true,
		Content:            content,
		ReferencesEventIDs: citedProposalEvents(artifact),
	}); err != nil {
		// The design IS amended and committed. Failing the request now would
		// tell the client nothing happened while their design has changed, and a
		// retry would then fail on old_text no longer being present.
		s.logger.Error("amendment: applied and committed, but the event could not be recorded",
			zap.String("approval_id", approvalID.String()),
			zap.String("commit", sha), zap.Error(err))
	}

	s.logger.Info("design amended",
		zap.String("workflow_id", workflowID.String()),
		zap.String("approval_id", approvalID.String()),
		zap.String("target", artifact.Target),
		zap.String("commit", sha),
		zap.String("decision_id", decisionID),
	)

	return &AmendmentOutcome{
		ApprovalID: approvalID, Status: "approved", Applied: true,
		Target: artifact.Target, CommitSHA: sha, DecisionID: decisionID,
	}, nil
}

// amendableTargets is the set of files an amendment may touch: the spine, the
// acceptance file, and the design sections actually assigned to this workflow.
//
// An allowlist, not a path-sanitising check. The target string comes from a
// model, and every sanitising approach has to anticipate what it might produce;
// an allowlist only has to know what is legitimate. That also makes path
// traversal a non-question: "../../etc/passwd" is simply not in the set.
//
// DECISIONS.md is absent on purpose — it is the append-only record of these
// decisions, written by appendDecision, and an amendment able to rewrite it
// could rewrite its own history.
//
// Package-level so the propose path and the apply path share one definition. The
// tool checks it to fail fast; the service checks it again at apply time, because
// the section list can change between a proposal and its approval.
func amendableTargets(ctx context.Context, sections *DesignSectionStore, workflowID uuid.UUID) (map[string]bool, error) {
	out := map[string]bool{finalMD: true, acceptanceMD: true}
	secs, err := sections.ListSections(ctx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("amendable targets: list sections: %w", err)
	}
	for _, sec := range secs {
		out[sec.SectionPath] = true
	}
	return out, nil
}

// appendDecision adds one entry to DECISIONS.md, newest first, and returns its
// id.
func (s *AmendmentService) appendDecision(
	mainPath string, approvalID, clientID uuid.UUID, artifact amendmentArtifact, dec AmendmentDecision,
) (string, error) {
	existing := ""
	if data, err := os.ReadFile(filepath.Join(mainPath, decisionsMD)); err == nil {
		existing = string(data)
	} else {
		existing = protocolTemplates[decisionsMD]
	}

	decisionID := nextDecisionID(existing)
	var entry strings.Builder
	fmt.Fprintf(&entry, "## %s — %s\n\n", decisionID, firstLine(artifact.Reason))
	fmt.Fprintf(&entry, "- Target: `%s`\n", artifact.Target)
	if artifact.ProposedByExpert != "" {
		fmt.Fprintf(&entry, "- Proposed by: %s\n", artifact.ProposedByExpert)
	}
	fmt.Fprintf(&entry, "- Decision: %s, by client %s on %s\n",
		dec.Decision, clientID, time.Now().UTC().Format(time.RFC3339))
	if dec.Notes != "" {
		fmt.Fprintf(&entry, "- Notes: %s\n", dec.Notes)
	}
	fmt.Fprintf(&entry, "- Amendment: %s\n\n", approvalID)

	return decisionID, writeWorkspaceFile(mainPath, decisionsMD, insertNewestFirst(existing, entry.String()))
}

// ============================================================
// Helpers that need more than the standard library
//
// The pure text operations — applyAmendmentText, nextDecisionID,
// insertNewestFirst, formatAcceptanceBlock and friends — live in
// amendment_text.go. See that file's header for why they are separate.
// ============================================================

// citedProposalEvents returns the proposal event id as a reference list, or nil.
func citedProposalEvents(artifact amendmentArtifact) []uuid.UUID {
	if artifact.ProposalEventID == "" {
		return nil
	}
	id, err := uuid.Parse(artifact.ProposalEventID)
	if err != nil {
		return nil
	}
	return []uuid.UUID{id}
}
