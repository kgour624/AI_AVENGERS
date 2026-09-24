package workflow

// QA phase runner — §19 of docs/COLLABORATIVE_DESIGN_ARCHITECTURE.md.
//
// WHAT CHANGED AND WHY
//
// The old QA path called AiderRunner.Run, which runs go build / go test
// against the workspace. The workspace now holds design markdown (authoring
// phase), not compiled source code. AiderRunner's verifyWorkspace finds no
// go.mod or package.json, reports "no project manifest found", and
// taskComplete is never true — so the phase silently burned 5 Aider
// iterations, spent real money, and produced nothing. This is the live bug
// described in the §19 gap note.
//
// WHAT QA MEANS NOW
//
// Testing experts (java-tester, react-tester, js-tester, etc.) read the
// completed design and propose test cases as acceptance criteria. Their
// output is a set of propose_acceptance blackboard events — one per test
// case — which the client approves exactly like any other amendment (§7.5).
// Gaps (sections with no AC, or AC with no verify command) are posted as
// qa_gap_found events so the admin can see them without blocking the
// workflow.
//
// WHAT IS REUSED, NOT REBUILT
//
//   documentCheck (authoring.go)     — workflow-level completeness check
//   GateSystem.RunGates              — loads expert training chunks
//   gateway.ModelGateway.Call        — the LLM call
//   blackboard.Store.Post            — event posting
//   DesignSectionStore.ListSections  — reads the real section roster
//
// WHAT IS NOT USED
//
//   AiderRunner  — no git workspace, no go build, no aider-service HTTP call
//   WorkspaceMerger — no per-expert workspaces to merge
//   verify.go    — build/test checks have no meaning for markdown

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/blackboard"
	"ai_avengers/backend/internal/gateway"
)

// QARunner runs one testing expert's QA turn.
//
// A thin struct — it holds only what every method needs. No workspace
// directory, no HTTP client, no aider-service URL: none of those are used.
type QARunner struct {
	db       *pgxpool.Pool
	store    *blackboard.Store
	gates    *GateSystem
	gw       *gateway.ModelGateway
	sections *DesignSectionStore
	// workspaceRoot is the same root AiderRunner and AuthoringRunner use.
	// QARunner reads from {workspaceRoot}/{workflowID}/main/ — the merged
	// workspace that authoring waves already produced. It never writes there.
	workspaceRoot string
	logger        *zap.Logger
}

// NewQARunner wires the QA runner.
// db is required so parseAndPostProposals can call proposeAmendment
// (creates the approval_requests row the client actually acts on).
func NewQARunner(
	db *pgxpool.Pool,
	store *blackboard.Store,
	gates *GateSystem,
	gw *gateway.ModelGateway,
	sections *DesignSectionStore,
	workspaceRoot string,
	logger *zap.Logger,
) *QARunner {
	return &QARunner{
		db:            db,
		store:         store,
		gates:         gates,
		gw:            gw,
		sections:      sections,
		workspaceRoot: workspaceRoot,
		logger:        logger,
	}
}

// QAResult mirrors AuthoringResult's shape so runner.go can handle both
// with the same log pattern.
type QAResult struct {
	// ProposedCount is how many propose_acceptance events were posted.
	ProposedCount int
	// GapCount is how many qa_gap_found events were posted.
	GapCount int
	// Completed is true when the document check passed AND at least one
	// test case was proposed. False means the expert ran but found nothing
	// to propose — still a valid outcome (the design may already be complete).
	Completed bool
}

// Run executes one testing expert's QA turn.
//
// Mental execution:
//
//	expert = java-tester, workflow has 3 sections
//
//	1. Read main/ workspace: final.md + design/10-pm.md + design/20-go.md
//	   + design/30-react.md + ACCEPTANCE.md
//	2. documentCheck(main/) — workflow-level: are all sections present?
//	   gaps -> qa_gap_found events
//	3. GateSystem.RunGates -> java-tester's training chunks
//	4. LLM call: "You are java-tester. Read the design. Propose test cases."
//	5. Parse LLM response -> propose_acceptance events on blackboard
//	6. Return QAResult{ProposedCount: N, GapCount: M, Completed: N > 0}
func (r *QARunner) Run(ctx context.Context, req AiderRunRequest) (*QAResult, error) {
	mainWorkspace := filepath.Join(r.workspaceRoot, req.WorkflowID.String(), "main")

	// Step 1: Read the design context from main/.
	designContext, err := r.buildDesignContext(ctx, req.WorkflowID, mainWorkspace)
	if err != nil {
		return nil, fmt.Errorf("qa: build design context: %w", err)
	}

	// Step 2: Workflow-level document check.
	// documentCheck is per-expert-workspace in authoring.go; here we run it
	// against main/ (the merged workspace) so it checks ALL sections, not
	// just one expert's. This is the cross-section check that per-expert
	// authoring cannot do.
	checkResult := documentCheck(mainWorkspace)
	gapCount := 0
	if !checkResult.Passed {
		for _, problem := range checkResult.Problems {
			_, postErr := r.store.Post(ctx, blackboard.PostRequest{
				WorkflowID:       req.WorkflowID,
				EventType:        "qa_gap_found",
				PostedByExpertID: &req.Expert.ID,
				Content: map[string]interface{}{
					"expert":  req.Expert.Name,
					"problem": problem,
					"phase":   PhaseQA,
				},
			})
			if postErr != nil {
				// Non-fatal: a lost gap event must not fail the QA turn.
				// Same treatment as publishDesignArtifact in authoring.go.
				r.logger.Warn("qa: could not post qa_gap_found",
					zap.String("expert", req.Expert.Name),
					zap.String("problem", problem),
					zap.Error(postErr),
				)
			} else {
				gapCount++
			}
		}
	}

	// Step 3: Load expert training via GateSystem.
	// allExperts is just this expert — QA is a solo turn, not a peer-poll.
	// genericAllowancePct = 0: test cases must come from the design and the
	// expert's training, not from generic knowledge.
	gateResult, err := r.gates.RunGates(ctx, req.WorkflowID, req.Expert,
		req.TaskDescription, []workflowExpert{req.Expert}, 0)
	if err != nil {
		// Non-fatal: proceed without training chunks rather than failing the
		// whole QA turn. The design context alone is enough to propose tests.
		r.logger.Warn("qa: gate system failed, continuing without training chunks",
			zap.String("expert", req.Expert.Name),
			zap.Error(err),
		)
	}

	// Step 4: Build the LLM prompt and call the model.
	systemPrompt := r.buildQASystemPrompt(req.Expert, gateResult)
	userPrompt := r.buildQAUserPrompt(req, designContext)

	resp, err := r.gw.Call(ctx, gateway.LLMRequest{
		Model:        gateway.ModelStrong,
		WorkflowID:   &req.WorkflowID,
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		MaxTokens:    4000,
	})
	if err != nil {
		return nil, fmt.Errorf("qa: llm call: %w", err)
	}

	// Step 5: Parse the LLM response and post propose_acceptance events.
	proposedCount, parseErr := r.parseAndPostProposals(ctx, req, resp.Content)
	if parseErr != nil {
		// Non-fatal: log and return what we have.
		r.logger.Warn("qa: some proposals could not be posted",
			zap.String("expert", req.Expert.Name),
			zap.Error(parseErr),
		)
	}

	r.logger.Info("qa turn completed",
		zap.String("expert", req.Expert.Name),
		zap.Int("proposed", proposedCount),
		zap.Int("gaps", gapCount),
	)

	return &QAResult{
		ProposedCount: proposedCount,
		GapCount:      gapCount,
		Completed:     proposedCount > 0 || gapCount == 0,
	}, nil
}

// buildDesignContext reads the spine, every section file, and ACCEPTANCE.md
// from main/ and renders them as one string for the LLM prompt.
//
// WHY read from main/ not from a per-expert workspace: QA runs AFTER all
// authoring waves have merged into main/. main/ is the single source of
// truth for the completed design (§2.1). A per-expert workspace would only
// hold what that expert seeded — not the full picture.
func (r *QARunner) buildDesignContext(
	ctx context.Context, workflowID uuid.UUID, mainWorkspace string,
) (string, error) {
	var sb strings.Builder

	// Spine first — it is the index every expert and the coding agent reads
	// before anything else (§3.3, §8).
	for _, name := range []string{finalMD, acceptanceMD, decisionsMD} {
		data, err := os.ReadFile(filepath.Join(mainWorkspace, name))
		if err != nil {
			if os.IsNotExist(err) {
				// Not yet authored — skip rather than fail. The document
				// check above already recorded this as a gap.
				continue
			}
			return "", fmt.Errorf("read %s: %w", name, err)
		}
		sb.WriteString(fmt.Sprintf("=== %s ===\n%s\n\n", name, string(data)))
	}

	// Section files in section_no order (ListSections returns them ordered).
	sections, err := r.sections.ListSections(ctx, workflowID)
	if err != nil {
		return "", fmt.Errorf("list sections: %w", err)
	}
	for _, sec := range sections {
		data, err := os.ReadFile(filepath.Join(mainWorkspace, sec.SectionPath))
		if err != nil {
			if os.IsNotExist(err) {
				continue // not yet authored — skip
			}
			return "", fmt.Errorf("read section %s: %w", sec.SectionPath, err)
		}
		sb.WriteString(fmt.Sprintf("=== %s ===\n%s\n\n", sec.SectionPath, string(data)))
	}

	return sb.String(), nil
}

// buildQASystemPrompt tells the testing expert who it is and what it may
// draw on. Same structure as WorkflowChatService.buildSystemPrompt so the
// knowledge rules are consistent across the two surfaces.
func (r *QARunner) buildQASystemPrompt(expert workflowExpert, gate *GateResult) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("You are %s, a testing expert in %s.\n", expert.Name, expert.Domain))
	sb.WriteString("Your job is to review the completed design and propose test cases as\n")
	sb.WriteString("acceptance criteria. You are NOT writing code — you are writing\n")
	sb.WriteString("machine-checkable test specifications.\n\n")

	if expert.ReasoningCharter != "" {
		sb.WriteString("=== YOUR RULES ===\n")
		sb.WriteString(expert.ReasoningCharter)
		sb.WriteString("\n\n")
	}

	// Training chunks from GateSystem — same FormatGateContext the workflow
	// chat uses, so citation markers are consistent.
	if gate != nil {
		if gateCtx := FormatGateContext(gate); gateCtx != "" {
			sb.WriteString(gateCtx)
			sb.WriteString("\n")
		}
	}

	sb.WriteString("=== KNOWLEDGE RULES ===\n")
	sb.WriteString("Test cases must come from the design and your training.\n")
	sb.WriteString("Do NOT invent requirements the design does not state.\n")
	sb.WriteString("If the design is silent on something, say so — do not fill the gap.\n\n")

	sb.WriteString("=== OUTPUT FORMAT ===\n")
	sb.WriteString("For each test case, output one block in EXACTLY this format:\n\n")
	sb.WriteString("PROPOSE_TEST\n")
	sb.WriteString("section: <section_path from the design, e.g. design/20-go.md>\n")
	sb.WriteString("statement: <what must be true — one sentence>\n")
	sb.WriteString("verify: <a command that exits 0 on pass, non-zero on fail>\n")
	sb.WriteString("done_when: <the observable condition>\n")
	sb.WriteString("END_TEST\n\n")
	sb.WriteString("Output as many PROPOSE_TEST blocks as needed. After all blocks,\n")
	sb.WriteString("write a short summary of what you checked and any gaps you found.\n")

	return sb.String()
}

// buildQAUserPrompt puts the full design context and the task in front of
// the testing expert.
func (r *QARunner) buildQAUserPrompt(req AiderRunRequest, designContext string) string {
	var sb strings.Builder

	sb.WriteString("=== DESIGN TO REVIEW ===\n")
	sb.WriteString(designContext)
	sb.WriteString("\n")
	sb.WriteString("=== YOUR TASK ===\n")
	sb.WriteString(req.TaskTitle)
	if req.TaskDescription != "" && req.TaskDescription != req.TaskTitle {
		sb.WriteString("\n")
		sb.WriteString(req.TaskDescription)
	}
	sb.WriteString("\n\nPropose test cases for your domain. Use the PROPOSE_TEST format above.")

	return sb.String()
}

// parseAndPostProposals parses PROPOSE_TEST blocks from the LLM response
// and records each via proposeAmendment (design_amendment_proposed event +
// pending approval_requests row). Nothing is written to ACCEPTANCE.md until
// the client approves — same §7.5 path as tool_loop's propose_acceptance.
//
// WHY NOT a standalone qa_test_proposed event:
// That event type had no consumer — no approval row, no list/approve handler.
// "pending_approval: true" was a dead flag. The real approval surface is
// gate_name='design_amendment' on approval_requests (amendment.go).
//
// WHY parse a structured block instead of asking for JSON:
// The LLM is already told to write design documents in markdown. Asking it
// to switch to JSON mid-response produces mixed output that is harder to
// parse reliably. A labelled-line block (PROPOSE_TEST / END_TEST) is the
// same shape as Aider's SEARCH/REPLACE — a format the model has seen in
// training and produces consistently.
func (r *QARunner) parseAndPostProposals(
	ctx context.Context, req AiderRunRequest, content string,
) (int, error) {
	const blockOpen = "PROPOSE_TEST"
	const blockClose = "END_TEST"

	// Section path → section number, so AC ids match toolProposeAcceptance
	// (AC-<sectionNo>-<n>) and documentCheck's validation.
	sectionNoByPath := map[string]int{}
	if r.sections != nil {
		if secs, listErr := r.sections.ListSections(ctx, req.WorkflowID); listErr == nil {
			for _, sec := range secs {
				sectionNoByPath[sec.SectionPath] = sec.SectionNo
			}
		}
	}

	// Existing ACCEPTANCE.md content (for next AC index). Best-effort:
	// missing file → start at 01.
	existingAcceptance := ""
	if full, pathErr := designPath(r.workspaceRoot, req.WorkflowID, acceptanceMD); pathErr == nil {
		if data, readErr := os.ReadFile(full); readErr == nil {
			existingAcceptance = string(data)
		}
	}
	// Track per-section next index so multiple proposals in one LLM response
	// don't collide on the same AC id.
	nextIdxBySection := map[int]int{}

	posted := 0
	var firstErr error
	remaining := content

	for {
		start := strings.Index(remaining, blockOpen)
		if start == -1 {
			break
		}
		rest := remaining[start+len(blockOpen):]
		end := strings.Index(rest, blockClose)
		if end == -1 {
			break
		}
		block := strings.TrimSpace(rest[:end])
		remaining = rest[end+len(blockClose):]

		proposal := parseTestBlock(block)
		if proposal == nil {
			// Malformed block — log and skip rather than failing the whole turn.
			r.logger.Warn("qa: malformed PROPOSE_TEST block, skipping",
				zap.String("expert", req.Expert.Name),
				zap.String("block", block),
			)
			continue
		}

		sectionNo := sectionNoByPath[proposal.Section]
		if sectionNo == 0 {
			// Unknown section — still propose as an appendable criterion
			// against ACCEPTANCE.md so the client can decide; use section 0
			// index so the id is still unique within this batch.
			r.logger.Warn("qa: proposal section not in design roster (proposing anyway)",
				zap.String("expert", req.Expert.Name),
				zap.String("section", proposal.Section),
			)
		}

		next := nextIdxBySection[sectionNo]
		if next == 0 {
			next = acceptanceIDsForSection(existingAcceptance, sectionNo) + 1
		}
		nextIdxBySection[sectionNo] = next + 1

		acID := fmt.Sprintf("AC-%d-%02d", sectionNo, next)
		owner := ownerTokenFromSectionPath(proposal.Section)
		if owner == "" {
			owner = req.Expert.Name
		}
		acBlock := formatAcceptanceBlock(
			acID, owner, proposal.Section,
			proposal.Statement, proposal.Verify, proposal.DoneWhen,
		)

		_, postErr := proposeAmendment(ctx, r.db, r.store, ProposeAmendmentRequest{
			WorkflowID: req.WorkflowID,
			ExpertID:   req.Expert.ID,
			ExpertName: req.Expert.Name,
			ChatID:     uuid.Nil, // QA phase, not chat-sourced
			Kind:       amendmentKindAcceptance,
			Target:     acceptanceMD,
			OldText:    "", // append — nothing to replace
			NewText:    acBlock,
			Reason:     fmt.Sprintf("%s proposes %s for %s (QA)", req.Expert.Name, acID, proposal.Section),
		})
		if postErr != nil {
			r.logger.Warn("qa: could not propose acceptance amendment",
				zap.String("expert", req.Expert.Name),
				zap.String("section", proposal.Section),
				zap.String("ac_id", acID),
				zap.Error(postErr),
			)
			if firstErr == nil {
				firstErr = postErr
			}
			continue
		}
		posted++
	}

	return posted, firstErr
}

// testProposal is one parsed PROPOSE_TEST block.
type testProposal struct {
	Section   string
	Statement string
	Verify    string
	DoneWhen  string
}

// parseTestBlock parses the labelled-line format inside a PROPOSE_TEST block.
// Returns nil if any required field is missing.
//
// Mental execution:
//
//	input: "section: design/20-go.md\nstatement: API returns 200\nverify: go test ./...\ndone_when: all tests pass"
//	output: &testProposal{Section:"design/20-go.md", Statement:"API returns 200", ...}
//
//	input: "section: design/20-go.md\nstatement: API returns 200"  (missing verify)
//	output: nil
func parseTestBlock(block string) *testProposal {
	p := &testProposal{}
	for _, line := range strings.Split(block, "\n") {
		line = strings.TrimSpace(line)
		if after, ok := strings.CutPrefix(line, "section:"); ok {
			p.Section = strings.TrimSpace(after)
		} else if after, ok := strings.CutPrefix(line, "statement:"); ok {
			p.Statement = strings.TrimSpace(after)
		} else if after, ok := strings.CutPrefix(line, "verify:"); ok {
			p.Verify = strings.TrimSpace(after)
		} else if after, ok := strings.CutPrefix(line, "done_when:"); ok {
			p.DoneWhen = strings.TrimSpace(after)
		}
	}
	// All four fields are required — a test case without a verify command
	// is not machine-checkable (§3.4's core requirement).
	if p.Section == "" || p.Statement == "" || p.Verify == "" || p.DoneWhen == "" {
		return nil
	}
	return p
}
