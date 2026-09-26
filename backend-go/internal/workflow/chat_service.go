package workflow

// Workflow chat — a client conversing with the experts who produced a
// deliverable. See docs/COLLABORATIVE_DESIGN_ARCHITECTURE.md §6.
//
// SEPARATION FROM THE PRODUCT CHAT (§6.1)
//
// This shares no code path and no table with the product chat:
//   product chat : chats, messages         internal/chat, internal/message,
//                                          internal/orchestrator, internal/decision
//   workflow chat: workflow_chats,         this file + chat_handler.go
//                  workflow_chat_messages,
//                  workflow_chat_participants
//
// The product chat is shipped and paid for. Threading workflow concerns through
// message/handler.go would put every future workflow-chat change one bug away
// from breaking it. Nothing here imports internal/chat or internal/message.
//
// WHY THIS LIVES IN package workflow, NOT internal/workflow/chat
//
// The doc proposed a subpackage. That does not compile: GateSystem,
// workflowExpert, Tools and DesignSectionStore are unexported members of package
// workflow, so a subpackage would force exporting them or duplicating them. The
// requirement was separation from the PRODUCT chat, and that is fully satisfied
// here — different package from internal/chat, different tables, different
// handler. Separation from the workflow engine was never the goal; this chat is
// part of the workflow engine.
//
// WHY context.Assembler IS NOT REUSED
//
// Assembler.Assemble reads chat_summaries, messages and chat_index — all product
// chat tables (assembler.go lines 459, 555, 571, 606). Pointing it at a
// workflow_chats id would silently query the wrong conversation. Only its
// expert-scoped retrieval is reusable, and GateSystem already calls that
// (GetCourseChunksForWorkflow), so training chunks and peer knowledge arrive
// through RunGates rather than through a second retrieval path.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/blackboard"
	"ai_avengers/backend/internal/gateway"
)

// Errors a caller is expected to branch on.
var (
	// ErrChatNotFound covers both "no such chat" and "not this client's chat".
	// Deliberately one error: telling a caller that a chat exists but belongs to
	// somebody else leaks the existence of other clients' workflows.
	ErrChatNotFound = errors.New("workflow chat not found")
	// ErrNotParticipant is returned when a question is addressed to an expert
	// that has not been added to the chat.
	ErrNotParticipant = errors.New("expert is not a participant in this chat")
	// ErrNoResponder is returned when a chat has no pinned deliverable author and
	// no participants, so there is nobody to answer.
	ErrNoResponder = errors.New("chat has no expert to answer; add a participant")
)

// defaultGenericCeiling is the fallback when system_settings has no
// generic_allowance_ceiling row — the same 30 that migration 017 hardcoded as a
// CHECK, kept as a last resort so a missing settings row cannot silently widen
// the policy.
const defaultGenericCeiling = 30.0

// chatRecentMessages caps how much prior conversation is replayed to the model.
// Small on purpose: the deliverable and the training chunks are the substance of
// an answer here, not the chat history.
const chatRecentMessages = 10

// WorkflowChat is one deliverable-scoped conversation.
type WorkflowChat struct {
	ID         uuid.UUID `json:"id"`
	WorkflowID uuid.UUID `json:"workflow_id"`
	ClientID   uuid.UUID `json:"client_id"`
	// PinnedEventID is the deliverable. nil = a chat about the workflow overall.
	PinnedEventID *uuid.UUID `json:"pinned_event_id,omitempty"`
	Title         string     `json:"title"`
	// GenericAllowancePct nil = inherit workflows.generic_allowance_pct (§6.4).
	GenericAllowancePct *float64  `json:"generic_allowance_pct,omitempty"`
	MessageCount        int       `json:"message_count"`
	IsArchived          bool      `json:"is_archived"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// WorkflowChatMessage is one message, or one tool-loop step once §7 lands.
type WorkflowChatMessage struct {
	ID         uuid.UUID       `json:"id"`
	ChatID     uuid.UUID       `json:"chat_id"`
	Role       string          `json:"role"`
	ExpertID   *uuid.UUID      `json:"expert_id,omitempty"`
	ExpertName string          `json:"expert_name,omitempty"`
	Content    string          `json:"content"`
	TurnNumber int             `json:"turn_number"`
	ToolName   *string         `json:"tool_name,omitempty"`
	ToolInput  json.RawMessage `json:"tool_input,omitempty"`
	ToolResult json.RawMessage `json:"tool_result,omitempty"`
	StepNumber *int            `json:"step_number,omitempty"`
	Citations  json.RawMessage `json:"citations,omitempty"`
	GateResult json.RawMessage `json:"gate_result,omitempty"`
	TokensUsed int             `json:"tokens_used"`
	CostUSD    float64         `json:"cost_usd"`
	CreatedAt  time.Time       `json:"created_at"`
}

// ChatParticipant is an expert taking part in a chat.
type ChatParticipant struct {
	ExpertID   uuid.UUID `json:"expert_id"`
	ExpertName string    `json:"expert_name"`
	Domain     string    `json:"domain"`
	AddedAt    time.Time `json:"added_at"`
}

// WorkflowChatService owns workflow-chat state and the answer pipeline.
// WHY no blackboard.Store field: deliverables are read by id and by id-array
// here, which the Store does not expose (it has Post / GetByType / GetSince).
// Carrying an unused dependency would suggest a coupling that does not exist.
// The two reads live in deliverableContext and citedEvents.
type WorkflowChatService struct {
	db            *pgxpool.Pool
	store         *blackboard.Store
	gates         *GateSystem
	gw            *gateway.ModelGateway
	sections      *DesignSectionStore
	tools         *ToolRegistry
	workspaceRoot string
	logger        *zap.Logger
}

// NewWorkflowChatService wires the service.
//
// store is the SAME blackboard.Store instance main.go already constructed with
// a real Redis client (bbStore), passed in rather than rebuilt here.
// Store.Post unconditionally calls publishNotification, which calls
// s.redis.Publish with no nil check (store.go:194, 328) — a Store built with a
// nil Redis client would panic the first time a mutating tool in the loop
// posts an event.
//
// workspaceRoot is the same AIDER_WORKSPACE_ROOT main.go already reads for
// AiderRunner — the tool loop's read_design/search_design tools read the
// identical {workspaceRoot}/{workflowID}/main/ layout that
// seedWorkspace/WorkspaceMerger already produce. One convention, not two.
func NewWorkflowChatService(
	db *pgxpool.Pool,
	store *blackboard.Store,
	gates *GateSystem,
	gw *gateway.ModelGateway,
	sections *DesignSectionStore,
	tools *ToolRegistry,
	workspaceRoot string,
	logger *zap.Logger,
) *WorkflowChatService {
	return &WorkflowChatService{
		db:            db,
		store:         store,
		gates:         gates,
		gw:            gw,
		sections:      sections,
		tools:         tools,
		workspaceRoot: workspaceRoot,
		logger:        logger,
	}
}

// ============================================================
// Chat lifecycle
// ============================================================

// CreateChat opens a chat, optionally pinned to a deliverable.
//
// When pinned, the deliverable's author is added as the first participant — that
// is the whole point of pinning, and it needs no extra table: the author is
// blackboard_events.posted_by_expert_id.
func (s *WorkflowChatService) CreateChat(
	ctx context.Context,
	workflowID, clientID uuid.UUID,
	pinnedEventID *uuid.UUID,
	title string,
) (*WorkflowChat, error) {
	// The workflow must exist AND belong to this client. Checked before anything
	// is written, so a wrong id cannot create an orphan chat.
	var ownerID uuid.UUID
	err := s.db.QueryRow(ctx,
		`SELECT client_id FROM workflows WHERE id = $1`, workflowID,
	).Scan(&ownerID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("create workflow chat: %w", ErrChatNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("create workflow chat: load workflow: %w", err)
	}
	if ownerID != clientID {
		return nil, fmt.Errorf("create workflow chat: %w", ErrChatNotFound)
	}

	// A pinned event must belong to the same workflow. Without this check a
	// client could pin another workflow's deliverable and pull its author and
	// content into this conversation.
	var authorID *uuid.UUID
	if pinnedEventID != nil {
		var evWorkflow uuid.UUID
		err := s.db.QueryRow(ctx,
			`SELECT workflow_id, posted_by_expert_id
			 FROM blackboard_events WHERE id = $1`, *pinnedEventID,
		).Scan(&evWorkflow, &authorID)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("create workflow chat: pinned event not found")
		}
		if err != nil {
			return nil, fmt.Errorf("create workflow chat: load pinned event: %w", err)
		}
		if evWorkflow != workflowID {
			return nil, fmt.Errorf("create workflow chat: pinned event belongs to a different workflow")
		}
	}

	if strings.TrimSpace(title) == "" {
		title = "Deliverable discussion"
	}

	var ch WorkflowChat
	err = s.db.QueryRow(ctx,
		`INSERT INTO workflow_chats (workflow_id, client_id, pinned_event_id, title)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, workflow_id, client_id, pinned_event_id, title,
		           generic_allowance_pct, message_count, is_archived,
		           created_at, updated_at`,
		workflowID, clientID, pinnedEventID, title,
	).Scan(&ch.ID, &ch.WorkflowID, &ch.ClientID, &ch.PinnedEventID, &ch.Title,
		&ch.GenericAllowancePct, &ch.MessageCount, &ch.IsArchived,
		&ch.CreatedAt, &ch.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create workflow chat: insert: %w", err)
	}

	// Author becomes the first participant. ON CONFLICT DO NOTHING because a
	// retried create must not fail on the participant row.
	if authorID != nil {
		if _, err := s.db.Exec(ctx,
			`INSERT INTO workflow_chat_participants (chat_id, expert_id, added_by)
			 VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`,
			ch.ID, *authorID, clientID,
		); err != nil {
			return nil, fmt.Errorf("create workflow chat: add author: %w", err)
		}
	}

	s.logger.Info("workflow chat created",
		zap.String("chat_id", ch.ID.String()),
		zap.String("workflow_id", workflowID.String()),
		zap.Bool("pinned", pinnedEventID != nil),
	)
	return &ch, nil
}

// GetChat loads a chat, enforcing client ownership.
func (s *WorkflowChatService) GetChat(ctx context.Context, chatID, clientID uuid.UUID) (*WorkflowChat, error) {
	var ch WorkflowChat
	err := s.db.QueryRow(ctx,
		`SELECT id, workflow_id, client_id, pinned_event_id, title,
		        generic_allowance_pct, message_count, is_archived,
		        created_at, updated_at
		 FROM workflow_chats
		 WHERE id = $1 AND client_id = $2`,
		chatID, clientID,
	).Scan(&ch.ID, &ch.WorkflowID, &ch.ClientID, &ch.PinnedEventID, &ch.Title,
		&ch.GenericAllowancePct, &ch.MessageCount, &ch.IsArchived,
		&ch.CreatedAt, &ch.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrChatNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get workflow chat: %w", err)
	}
	return &ch, nil
}

// ListChats returns a workflow's chats, newest first.
func (s *WorkflowChatService) ListChats(ctx context.Context, workflowID, clientID uuid.UUID) ([]WorkflowChat, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, workflow_id, client_id, pinned_event_id, title,
		        generic_allowance_pct, message_count, is_archived,
		        created_at, updated_at
		 FROM workflow_chats
		 WHERE workflow_id = $1 AND client_id = $2 AND is_archived = FALSE
		 ORDER BY updated_at DESC`,
		workflowID, clientID,
	)
	if err != nil {
		return nil, fmt.Errorf("list workflow chats: %w", err)
	}
	defer rows.Close()

	out := []WorkflowChat{}
	for rows.Next() {
		var ch WorkflowChat
		if err := rows.Scan(&ch.ID, &ch.WorkflowID, &ch.ClientID, &ch.PinnedEventID,
			&ch.Title, &ch.GenericAllowancePct, &ch.MessageCount, &ch.IsArchived,
			&ch.CreatedAt, &ch.UpdatedAt); err != nil {
			return nil, fmt.Errorf("list workflow chats: scan: %w", err)
		}
		out = append(out, ch)
	}
	return out, rows.Err()
}

// SetKnowledgeMode sets the chat's generic allowance (§6.4).
//
// pct nil clears the override, so the chat inherits the workflow's value again.
// A value above the configured business ceiling is clamped, not rejected: the
// client asked for "as much generic as allowed", and failing the request would
// be a worse answer to that than giving them the maximum.
func (s *WorkflowChatService) SetKnowledgeMode(ctx context.Context, chatID, clientID uuid.UUID, pct *float64) (*WorkflowChat, error) {
	if _, err := s.GetChat(ctx, chatID, clientID); err != nil {
		return nil, err
	}

	if pct != nil {
		if *pct < 0 {
			return nil, fmt.Errorf("set knowledge mode: percentage cannot be negative")
		}
		ceiling := s.genericCeiling(ctx)
		if *pct > ceiling {
			s.logger.Info("workflow chat: generic allowance clamped to ceiling",
				zap.String("chat_id", chatID.String()),
				zap.Float64("requested", *pct),
				zap.Float64("ceiling", ceiling),
			)
			clamped := ceiling
			pct = &clamped
		}
	}

	if _, err := s.db.Exec(ctx,
		`UPDATE workflow_chats SET generic_allowance_pct = $1, updated_at = NOW()
		 WHERE id = $2 AND client_id = $3`,
		pct, chatID, clientID,
	); err != nil {
		return nil, fmt.Errorf("set knowledge mode: %w", err)
	}
	return s.GetChat(ctx, chatID, clientID)
}

// ============================================================
// Participants
// ============================================================

// AddParticipant brings another expert into the conversation (§6.3).
func (s *WorkflowChatService) AddParticipant(ctx context.Context, chatID, clientID, expertID uuid.UUID) error {
	if _, err := s.GetChat(ctx, chatID, clientID); err != nil {
		return err
	}

	// The expert must exist and be active. Without this, a stale id from the UI
	// would create a participant row whose expert can never answer.
	var exists bool
	if err := s.db.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM experts
		                WHERE id = $1 AND is_active = TRUE AND deleted_at IS NULL)`,
		expertID,
	).Scan(&exists); err != nil {
		return fmt.Errorf("add participant: expert check: %w", err)
	}
	if !exists {
		return fmt.Errorf("add participant: %w: %s", ErrExpertNotFound, expertID)
	}

	if _, err := s.db.Exec(ctx,
		`INSERT INTO workflow_chat_participants (chat_id, expert_id, added_by)
		 VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`,
		chatID, expertID, clientID,
	); err != nil {
		return fmt.Errorf("add participant: %w", err)
	}
	s.logger.Info("workflow chat participant added",
		zap.String("chat_id", chatID.String()),
		zap.String("expert_id", expertID.String()),
	)
	return nil
}

// RemoveParticipant drops an expert from the conversation.
func (s *WorkflowChatService) RemoveParticipant(ctx context.Context, chatID, clientID, expertID uuid.UUID) error {
	if _, err := s.GetChat(ctx, chatID, clientID); err != nil {
		return err
	}
	if _, err := s.db.Exec(ctx,
		`DELETE FROM workflow_chat_participants WHERE chat_id = $1 AND expert_id = $2`,
		chatID, expertID,
	); err != nil {
		return fmt.Errorf("remove participant: %w", err)
	}
	return nil
}

// ListParticipants returns the experts in a chat, in join order.
func (s *WorkflowChatService) ListParticipants(ctx context.Context, chatID uuid.UUID) ([]ChatParticipant, error) {
	rows, err := s.db.Query(ctx,
		`SELECT p.expert_id, e.name, e.domain, p.added_at
		 FROM workflow_chat_participants p
		 JOIN experts e ON e.id = p.expert_id
		 WHERE p.chat_id = $1
		 ORDER BY p.added_at`,
		chatID,
	)
	if err != nil {
		return nil, fmt.Errorf("list participants: %w", err)
	}
	defer rows.Close()

	out := []ChatParticipant{}
	for rows.Next() {
		var p ChatParticipant
		if err := rows.Scan(&p.ExpertID, &p.ExpertName, &p.Domain, &p.AddedAt); err != nil {
			return nil, fmt.Errorf("list participants: scan: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// ============================================================
// Messages
// ============================================================

// ListMessages returns a chat's messages oldest-first, tool steps included.
func (s *WorkflowChatService) ListMessages(ctx context.Context, chatID, clientID uuid.UUID, limit int) ([]WorkflowChatMessage, error) {
	if _, err := s.GetChat(ctx, chatID, clientID); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 500 {
		limit = 200
	}

	rows, err := s.db.Query(ctx,
		`SELECT m.id, m.chat_id, m.role, m.expert_id, COALESCE(e.name, ''),
		        m.content, m.turn_number, m.tool_name, m.tool_input, m.tool_result,
		        m.step_number, m.citations, m.gate_result, m.tokens_used,
		        m.cost_usd, m.created_at
		 FROM workflow_chat_messages m
		 LEFT JOIN experts e ON e.id = m.expert_id
		 WHERE m.chat_id = $1
		 ORDER BY m.created_at
		 LIMIT $2`,
		chatID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	defer rows.Close()

	out := []WorkflowChatMessage{}
	for rows.Next() {
		var m WorkflowChatMessage
		if err := rows.Scan(&m.ID, &m.ChatID, &m.Role, &m.ExpertID, &m.ExpertName,
			&m.Content, &m.TurnNumber, &m.ToolName, &m.ToolInput, &m.ToolResult,
			&m.StepNumber, &m.Citations, &m.GateResult, &m.TokensUsed,
			&m.CostUSD, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("list messages: scan: %w", err)
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// Send records the client's question, produces one expert answer, stores it and
// returns it.
//
// expertID selects the responder. nil means the default from §6.3: the pinned
// deliverable's author. Fanning out to every participant on every message is
// deliberately NOT the default — three participants would mean three retrievals
// and three completions per message, which is the same waste that had the Aider
// loop re-running an identical embedding five times per task before 31becac.
func (s *WorkflowChatService) Send(
	ctx context.Context,
	chatID, clientID uuid.UUID,
	expertID *uuid.UUID,
	question string,
) (*WorkflowChatMessage, error) {
	question = strings.TrimSpace(question)
	if question == "" {
		return nil, fmt.Errorf("send: message is required")
	}

	ch, err := s.GetChat(ctx, chatID, clientID)
	if err != nil {
		return nil, err
	}

	responderID, err := s.resolveResponder(ctx, ch, expertID)
	if err != nil {
		return nil, err
	}

	expert, err := s.loadChatExpert(ctx, responderID)
	if err != nil {
		return nil, err
	}

	turn, err := s.nextTurnNumber(ctx, chatID)
	if err != nil {
		return nil, err
	}

	// The question is stored before the model is called, so a failed or slow
	// answer never loses what the client asked.
	if _, err := s.db.Exec(ctx,
		`INSERT INTO workflow_chat_messages (chat_id, role, content, turn_number)
		 VALUES ($1, 'user', $2, $3)`,
		chatID, question, turn,
	); err != nil {
		return nil, fmt.Errorf("send: save question: %w", err)
	}

	// Gates decide what knowledge this answer may draw on — the same call the
	// design phase makes, so an answer about a deliverable obeys the rules that
	// produced it.
	genericPct := s.effectiveGenericPct(ctx, ch)
	participants, err := s.participantExperts(ctx, chatID)
	if err != nil {
		return nil, err
	}
	gateResult, err := s.gates.RunGates(ctx, ch.WorkflowID, expert, question, participants, genericPct)
	if err != nil {
		return nil, fmt.Errorf("send: gates: %w", err)
	}

	deliverable, err := s.deliverableContext(ctx, ch)
	if err != nil {
		return nil, err
	}
	history, err := s.recentHistory(ctx, chatID, chatRecentMessages)
	if err != nil {
		return nil, err
	}
	sectionList, err := s.sectionList(ctx, ch.WorkflowID)
	if err != nil {
		return nil, err
	}

	systemPrompt := s.buildSystemPrompt(expert, gateResult, genericPct)
	// Ground the expert in the files that ACTUALLY exist. Without this it only
	// sees design sections (empty for a code-producing run), guesses paths from
	// the client's message, and reports it cannot read its own output — the
	// production bug where a completed workflow's files were unreadable in chat.
	if files := workspaceFiles(s.workspaceRoot, ch.WorkflowID); len(files) > 0 {
		systemPrompt += "\n\nFILES THAT ACTUALLY EXIST IN THIS WORKFLOW:\n- " + strings.Join(files, "\n- ") +
			"\nRead any of them with read_design(path). Never claim a file cannot be read without first calling read_design on a path from this list."
	}
	toolCatalogue := promptCatalogue(s.tools.ForExpert(expert))
	userPrompt := s.buildUserPrompt(deliverable, sectionList, history, question)

	// Audit record: which knowledge rules produced this answer (§6.4).
	// GenericAllowed is true when training + peers together do NOT cover the
	// task, so "blocked" is its negation. Recorded from the gate result rather
	// than re-derived, so the audit row cannot drift from what the gates decided.
	gateJSON, _ := json.Marshal(map[string]any{
		"training_chunks":       len(gateResult.TrainingChunks),
		"peer_contributions":    len(gateResult.PeerContributions),
		"gate1_passed":          gateResult.Gate1Passed,
		"generic_allowed":       gateResult.GenericAllowed,
		"generic_blocked":       !gateResult.GenericAllowed,
		"generic_allowance_pct": gateResult.GenericAllowancePct,
		"coverage_gap":          gateResult.CoverageGap,
	})

	// The tool loop (§7.3): gather -> act -> verify. Every step — including
	// every intermediate tool call — is persisted as its own row before the
	// loop continues, so a client can see what the expert read before it
	// answered (§7.6 "shows its work"), and so a crash mid-loop loses at most
	// the step in flight, not the ones already done.
	loopCtx := &toolLoopContext{
		db:            s.db,
		store:         s.store, // the real Store — see NewWorkflowChatService's comment
		sections:      s.sections,
		gates:         s.gates,
		gw:            s.gw,
		workflowID:    ch.WorkflowID,
		expert:        expert,
		chatID:        chatID,
		workspaceRoot: s.workspaceRoot,
	}

	maxSteps := s.toolLoopMaxSteps(ctx)
	stepNo := 0
	convo := userPrompt + "\n\n" + toolCatalogue
	var finalContent string
	var totalTokens int
	var totalCost float64

	for {
		stepNo++
		resp, err := s.gw.Call(ctx, gateway.LLMRequest{
			Model:        gateway.ModelStrong,
			WorkflowID:   &ch.WorkflowID, // chat spend lands on the workflow budget
			SystemPrompt: systemPrompt,
			UserPrompt:   convo,
			MaxTokens:    4000,
		})
		if err != nil {
			return nil, fmt.Errorf("send: llm (step %d): %w", stepNo, err)
		}
		totalTokens += resp.InputTokens + resp.OutputTokens
		totalCost += resp.CostUSD

		calls := parseToolCalls(resp.Content)
		if len(calls) == 0 {
			// No tool call: this is the final answer.
			finalContent = resp.Content
			break
		}

		// Persist the assistant's tool-call turn itself (content minus the
		// tool_call blocks, which may be empty — the model sometimes emits
		// nothing but the call).
		reasoning := stripToolCallBlocks(resp.Content)
		if reasoning != "" {
			if _, err := s.db.Exec(ctx,
				`INSERT INTO workflow_chat_messages (chat_id, role, expert_id, content, turn_number, step_number)
				 VALUES ($1, 'assistant', $2, $3, $4, $5)`,
				chatID, expert.ID, reasoning, turn, stepNo,
			); err != nil {
				s.logger.Warn("workflow chat: could not persist reasoning step",
					zap.String("chat_id", chatID.String()), zap.Error(err))
			}
		}

		var toolResultsText strings.Builder
		for _, call := range calls {
			tool, known := s.tools.lookup(call.Tool, expert)
			var resultJSON json.RawMessage
			var resultErr error
			if !known {
				resultErr = fmt.Errorf("tool %q is not available to this expert", call.Tool)
			} else {
				var out any
				out, resultErr = tool.Handler(ctx, loopCtx, call.Input)
				if resultErr == nil {
					resultJSON, _ = json.Marshal(out)
				}
			}

			var resultForPrompt string
			if resultErr != nil {
				resultForPrompt = fmt.Sprintf("ERROR: %s", resultErr.Error())
				resultJSON, _ = json.Marshal(map[string]string{"error": resultErr.Error()})
			} else {
				resultForPrompt = string(resultJSON)
			}
			toolResultsText.WriteString(fmt.Sprintf("[%s result]\n%s\n\n", call.Tool, resultForPrompt))

			inputJSON := call.Input
			if len(inputJSON) == 0 {
				inputJSON = json.RawMessage("{}")
			}
			if _, err := s.db.Exec(ctx,
				`INSERT INTO workflow_chat_messages
				     (chat_id, role, tool_name, tool_input, tool_result, turn_number, step_number)
				 VALUES ($1, 'tool', $2, $3, $4, $5, $6)`,
				chatID, call.Tool, inputJSON, resultJSON, turn, stepNo,
			); err != nil {
				s.logger.Warn("workflow chat: could not persist tool step",
					zap.String("chat_id", chatID.String()), zap.String("tool", call.Tool),
					zap.Error(err))
			}
		}

		if stepNo >= maxSteps {
			// The cap exists so a confused model cannot loop until the
			// workflow's budget is gone (§7.3). Ending here — rather than
			// erroring — still gives the client an answer, built from
			// whatever the loop already gathered.
			s.logger.Warn("workflow chat: tool loop hit max steps",
				zap.String("chat_id", chatID.String()), zap.Int("max_steps", maxSteps))
			finalContent = "I gathered the following before reaching my step limit:\n\n" + toolResultsText.String()
			break
		}

		// Feed the tool results back in and let the model continue.
		convo = convo + "\n\n=== YOUR PREVIOUS TOOL CALLS ===\n" + resp.Content +
			"\n\n=== TOOL RESULTS ===\n" + toolResultsText.String()
	}

	var out WorkflowChatMessage
	err = s.db.QueryRow(ctx,
		`INSERT INTO workflow_chat_messages
		     (chat_id, role, expert_id, content, turn_number,
		      gate_result, tokens_used, cost_usd)
		 VALUES ($1, 'assistant', $2, $3, $4, $5, $6, $7)
		 RETURNING id, chat_id, role, expert_id, content, turn_number,
		           gate_result, tokens_used, cost_usd, created_at`,
		chatID, expert.ID, finalContent, turn, gateJSON,
		totalTokens, totalCost,
	).Scan(&out.ID, &out.ChatID, &out.Role, &out.ExpertID, &out.Content,
		&out.TurnNumber, &out.GateResult, &out.TokensUsed, &out.CostUSD,
		&out.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("send: save answer: %w", err)
	}
	out.ExpertName = expert.Name

	// Two rows were added this turn (question + answer).
	if _, err := s.db.Exec(ctx,
		`UPDATE workflow_chats SET message_count = message_count + 2, updated_at = NOW()
		 WHERE id = $1`, chatID,
	); err != nil {
		// Non-fatal: a stale counter must not discard an answer the client paid
		// for. Same treatment as the cost write in gateway.addWorkflowCost.
		s.logger.Warn("workflow chat: message_count update failed",
			zap.String("chat_id", chatID.String()), zap.Error(err))
	}

	s.logger.Info("workflow chat answered",
		zap.String("chat_id", chatID.String()),
		zap.String("expert", expert.Name),
		zap.Int("turn", turn),
		zap.Int("training_chunks", len(gateResult.TrainingChunks)),
		zap.Int("peers", len(gateResult.PeerContributions)),
		zap.Float64("generic_pct", genericPct),
		zap.Int("tool_steps", stepNo),
		zap.Float64("cost_usd", totalCost),
	)
	return &out, nil
}

// ============================================================
// Internals
// ============================================================

// resolveResponder picks which expert answers.
//
// An explicitly addressed expert must already be a participant. Otherwise a
// client could pull any expert in the system into the conversation by id,
// bypassing AddParticipant and its active/exists checks.
func (s *WorkflowChatService) resolveResponder(ctx context.Context, ch *WorkflowChat, requested *uuid.UUID) (uuid.UUID, error) {
	if requested != nil {
		var ok bool
		if err := s.db.QueryRow(ctx,
			`SELECT EXISTS (SELECT 1 FROM workflow_chat_participants
			                WHERE chat_id = $1 AND expert_id = $2)`,
			ch.ID, *requested,
		).Scan(&ok); err != nil {
			return uuid.Nil, fmt.Errorf("resolve responder: %w", err)
		}
		if !ok {
			return uuid.Nil, fmt.Errorf("resolve responder: %w", ErrNotParticipant)
		}
		return *requested, nil
	}

	// Default: the pinned deliverable's author, which is the first participant
	// added at creation. Falling back to the earliest participant covers an
	// unpinned chat.
	var id uuid.UUID
	err := s.db.QueryRow(ctx,
		`SELECT expert_id FROM workflow_chat_participants
		 WHERE chat_id = $1 ORDER BY added_at LIMIT 1`,
		ch.ID,
	).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, ErrNoResponder
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("resolve responder: %w", err)
	}
	return id, nil
}

// loadChatExpert reads one expert into the same struct the gate system and
// agent loop use, with the same query shape as
// WorkflowRunner.loadWorkflowExperts — one loader shape, not two.
func (s *WorkflowChatService) loadChatExpert(ctx context.Context, expertID uuid.UUID) (workflowExpert, error) {
	var e workflowExpert
	var toolsJSON []byte
	err := s.db.QueryRow(ctx,
		`SELECT id, name, domain,
		        COALESCE(reasoning_charter, ''),
		        COALESCE(loop_pattern, 'ota'),
		        COALESCE(max_loop_iterations, 5),
		        COALESCE(allowed_tools, '[]'::jsonb)
		 FROM experts
		 WHERE id = $1 AND is_active = TRUE AND deleted_at IS NULL`,
		expertID,
	).Scan(&e.ID, &e.Name, &e.Domain, &e.ReasoningCharter,
		&e.LoopPattern, &e.MaxLoopIterations, &toolsJSON)
	if errors.Is(err, pgx.ErrNoRows) {
		return workflowExpert{}, fmt.Errorf("load chat expert: %w: %s", ErrExpertNotFound, expertID)
	}
	if err != nil {
		return workflowExpert{}, fmt.Errorf("load chat expert: %w", err)
	}
	if len(toolsJSON) > 0 {
		_ = json.Unmarshal(toolsJSON, &e.AllowedTools)
	}
	return e, nil
}

// participantExperts loads every participant, which Gate 2 polls for peer
// knowledge. The responder is included; RunGates skips itself.
func (s *WorkflowChatService) participantExperts(ctx context.Context, chatID uuid.UUID) ([]workflowExpert, error) {
	rows, err := s.db.Query(ctx,
		`SELECT e.id, e.name, e.domain,
		        COALESCE(e.reasoning_charter, ''),
		        COALESCE(e.loop_pattern, 'ota'),
		        COALESCE(e.max_loop_iterations, 5),
		        COALESCE(e.allowed_tools, '[]'::jsonb)
		 FROM workflow_chat_participants p
		 JOIN experts e ON e.id = p.expert_id
		 WHERE p.chat_id = $1 AND e.is_active = TRUE AND e.deleted_at IS NULL
		 ORDER BY p.added_at`,
		chatID,
	)
	if err != nil {
		return nil, fmt.Errorf("participant experts: %w", err)
	}
	defer rows.Close()

	var out []workflowExpert
	for rows.Next() {
		var e workflowExpert
		var toolsJSON []byte
		if err := rows.Scan(&e.ID, &e.Name, &e.Domain, &e.ReasoningCharter,
			&e.LoopPattern, &e.MaxLoopIterations, &toolsJSON); err != nil {
			return nil, fmt.Errorf("participant experts: scan: %w", err)
		}
		if len(toolsJSON) > 0 {
			_ = json.Unmarshal(toolsJSON, &e.AllowedTools)
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// nextTurnNumber returns the next turn for a chat. One turn holds the client's
// question and the answer(s) to it.
func (s *WorkflowChatService) nextTurnNumber(ctx context.Context, chatID uuid.UUID) (int, error) {
	var turn int
	if err := s.db.QueryRow(ctx,
		`SELECT COALESCE(MAX(turn_number), 0) + 1
		 FROM workflow_chat_messages WHERE chat_id = $1`, chatID,
	).Scan(&turn); err != nil {
		return 0, fmt.Errorf("next turn: %w", err)
	}
	return turn, nil
}

// effectiveGenericPct resolves the knowledge mode (§6.4):
// chat override -> workflow value -> clamped to the configured ceiling.
func (s *WorkflowChatService) effectiveGenericPct(ctx context.Context, ch *WorkflowChat) float64 {
	pct := 0.0
	if ch.GenericAllowancePct != nil {
		pct = *ch.GenericAllowancePct
	} else {
		if err := s.db.QueryRow(ctx,
			`SELECT COALESCE(generic_allowance_pct, 0) FROM workflows WHERE id = $1`,
			ch.WorkflowID,
		).Scan(&pct); err != nil {
			// Safe default is strict. A settings read failure must never widen
			// what an expert is allowed to invent.
			s.logger.Warn("workflow chat: could not read workflow generic allowance, using 0",
				zap.String("workflow_id", ch.WorkflowID.String()), zap.Error(err))
			return 0
		}
	}

	if ceiling := s.genericCeiling(ctx); pct > ceiling {
		return ceiling
	}
	if pct < 0 {
		return 0
	}
	return pct
}

// toolLoopMaxSteps reads the step cap from system_settings (§11, §7.3).
// Falls back to toolLoopMaxStepsDefault (tool_loop.go) on any read failure —
// a missing settings row must bound the loop, never leave it unbounded.
func (s *WorkflowChatService) toolLoopMaxSteps(ctx context.Context) int {
	var raw []byte
	if err := s.db.QueryRow(ctx,
		`SELECT value FROM system_settings WHERE key = 'tool_loop_max_steps'`,
	).Scan(&raw); err != nil {
		return toolLoopMaxStepsDefault
	}
	var v int
	if err := json.Unmarshal(raw, &v); err != nil || v <= 0 {
		return toolLoopMaxStepsDefault
	}
	return v
}

// genericCeiling reads the business ceiling from system_settings (§11).
//
// The safety bound (0..100) lives in the table CHECK; this is the policy value
// an admin may change without a migration. On any failure it returns the old
// hardcoded 30 — a missing settings row must not silently widen the policy.
func (s *WorkflowChatService) genericCeiling(ctx context.Context) float64 {
	var raw []byte
	if err := s.db.QueryRow(ctx,
		`SELECT value FROM system_settings WHERE key = 'generic_allowance_ceiling'`,
	).Scan(&raw); err != nil {
		return defaultGenericCeiling
	}
	var v float64
	if err := json.Unmarshal(raw, &v); err != nil || v < 0 {
		return defaultGenericCeiling
	}
	return v
}

// deliverableContext renders the pinned deliverable and the artifacts it cites.
//
// This is the subject of the conversation, so it is rendered first in the prompt
// (see buildUserPrompt) — not last, where a model skims.
func (s *WorkflowChatService) deliverableContext(ctx context.Context, ch *WorkflowChat) (string, error) {
	if ch.PinnedEventID == nil {
		return "", nil
	}

	// Only scalars are scanned. references_event_ids is a Postgres uuid[], and
	// nothing in this codebase has ever scanned one — blackboard.Store writes it
	// (store.go:140, cast from JSON) and copies it back from the request, but
	// never reads it out of a row. Rather than be the first place to rely on an
	// unproven array decode, citedEvents resolves the references entirely in SQL
	// with a LATERAL unnest, so no array ever crosses into Go.
	var (
		eventType string
		content   []byte
		refCount  int
		postedAt  time.Time
	)
	err := s.db.QueryRow(ctx,
		`SELECT event_type, content,
		        COALESCE(array_length(references_event_ids, 1), 0),
		        posted_at
		 FROM blackboard_events WHERE id = $1`, *ch.PinnedEventID,
	).Scan(&eventType, &content, &refCount, &postedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		// The chat outlives a deleted event only if the workflow was removed,
		// which cascades this chat too — so this is a data oddity, not a client
		// error. Answer without the deliverable rather than failing the turn.
		s.logger.Warn("workflow chat: pinned event missing",
			zap.String("chat_id", ch.ID.String()))
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("deliverable context: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("=== THE DELIVERABLE THIS CONVERSATION IS ABOUT ===\n")
	sb.WriteString(fmt.Sprintf("Type: %s\nProduced: %s\n\n", eventType, postedAt.Format(time.RFC3339)))
	sb.WriteString(string(content))
	sb.WriteString("\n")

	if refCount > 0 {
		cited, err := s.citedEvents(ctx, *ch.PinnedEventID)
		if err != nil {
			// Cited artifacts are supporting context. Losing them degrades the
			// answer; failing the turn denies it entirely.
			s.logger.Warn("workflow chat: could not load cited events", zap.Error(err))
		} else if cited != "" {
			sb.WriteString("\n--- ARTIFACTS THIS DELIVERABLE CITES ---\n")
			sb.WriteString(cited)
		}
	}
	return sb.String(), nil
}

// citedEvents renders the artifacts a deliverable references.
//
// The reference list is expanded inside Postgres (LATERAL unnest) instead of
// being read into Go and sent back as a parameter. Two reasons: one round trip
// instead of two, and no dependency on decoding a uuid[] — see deliverableContext.
func (s *WorkflowChatService) citedEvents(ctx context.Context, pinnedEventID uuid.UUID) (string, error) {
	rows, err := s.db.Query(ctx,
		`SELECT c.event_type, c.content
		 FROM blackboard_events p
		 CROSS JOIN LATERAL unnest(p.references_event_ids) AS r(id)
		 JOIN blackboard_events c ON c.id = r.id
		 WHERE p.id = $1
		 ORDER BY c.sequence_number`,
		pinnedEventID,
	)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var sb strings.Builder
	for rows.Next() {
		var eventType string
		var content []byte
		if err := rows.Scan(&eventType, &content); err != nil {
			return "", err
		}
		sb.WriteString(fmt.Sprintf("[%s]\n%s\n\n", eventType, string(content)))
	}
	return sb.String(), rows.Err()
}

// recentHistory renders the last N messages, oldest-first.
func (s *WorkflowChatService) recentHistory(ctx context.Context, chatID uuid.UUID, limit int) (string, error) {
	// Newest-first in SQL to get the last N, then reversed for the prompt so the
	// conversation reads forwards.
	rows, err := s.db.Query(ctx,
		`SELECT m.role, COALESCE(e.name, ''), m.content
		 FROM workflow_chat_messages m
		 LEFT JOIN experts e ON e.id = m.expert_id
		 WHERE m.chat_id = $1 AND m.role IN ('user', 'assistant')
		 ORDER BY m.created_at DESC
		 LIMIT $2`,
		chatID, limit,
	)
	if err != nil {
		return "", fmt.Errorf("recent history: %w", err)
	}
	defer rows.Close()

	type entry struct{ role, name, content string }
	var entries []entry
	for rows.Next() {
		var e entry
		if err := rows.Scan(&e.role, &e.name, &e.content); err != nil {
			return "", fmt.Errorf("recent history: scan: %w", err)
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return "", err
	}

	var sb strings.Builder
	for i := len(entries) - 1; i >= 0; i-- {
		e := entries[i]
		who := "Client"
		if e.role == "assistant" {
			who = e.name
			if who == "" {
				who = "Expert"
			}
		}
		sb.WriteString(fmt.Sprintf("%s: %s\n\n", who, e.content))
	}
	return sb.String(), nil
}

// sectionList renders the design sections that exist, so the expert can refer to
// them by name. Generated from workflow_design_sections — never hand-written.
func (s *WorkflowChatService) sectionList(ctx context.Context, workflowID uuid.UUID) (string, error) {
	secs, err := s.sections.ListSections(ctx, workflowID)
	if err != nil {
		return "", err
	}
	if len(secs) == 0 {
		return "", nil
	}

	// Owner names in one query, joined through workflow_design_sections rather
	// than passing an array of ids back as a parameter. Same reason as
	// citedEvents: no array crosses the boundary, and it is one round trip
	// instead of one per section.
	names := map[uuid.UUID]string{}
	rows, err := s.db.Query(ctx,
		`SELECT s.expert_id, e.name
		 FROM workflow_design_sections s
		 JOIN experts e ON e.id = s.expert_id
		 WHERE s.workflow_id = $1`,
		workflowID,
	)
	if err != nil {
		return "", fmt.Errorf("section list: names: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var id uuid.UUID
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return "", fmt.Errorf("section list: scan: %w", err)
		}
		names[id] = name
	}
	if err := rows.Err(); err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString("=== DESIGN SECTIONS IN THIS PROJECT ===\n")
	for _, sec := range secs {
		sb.WriteString(fmt.Sprintf("  %s  (owner: %s)\n", sec.SectionPath, names[sec.ExpertID]))
	}
	return sb.String(), nil
}

// buildSystemPrompt states who the expert is and what it may draw on.
func (s *WorkflowChatService) buildSystemPrompt(expert workflowExpert, gate *GateResult, genericPct float64) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("You are %s, a domain expert in %s.\n", expert.Name, expert.Domain))
	sb.WriteString("You are answering a client's question about a deliverable YOU produced in this project.\n\n")

	if expert.ReasoningCharter != "" {
		sb.WriteString("=== YOUR RULES (follow strictly) ===\n")
		sb.WriteString(expert.ReasoningCharter)
		sb.WriteString("\n\n")
	}

	// Training chunks and peer knowledge, rendered by the same formatter the
	// design phase uses, so the citation markers are identical across both.
	if gateCtx := FormatGateContext(gate); gateCtx != "" {
		sb.WriteString(gateCtx)
		sb.WriteString("\n")
	}

	sb.WriteString("=== KNOWLEDGE RULES ===\n")
	switch {
	case genericPct <= 0:
		sb.WriteString("Trained-only mode. Answer ONLY from your training above and from the\n")
		sb.WriteString("deliverable and design sections given below. If neither covers the\n")
		sb.WriteString("question, say exactly what is missing and what you would need. Do NOT\n")
		sb.WriteString("fill the gap with general knowledge.\n")
	default:
		sb.WriteString(fmt.Sprintf("Your training and the approved design are the primary sources. At most\n"+
			"%.0f%% of this answer may come from general knowledge, and only for gaps the\n"+
			"sources do not cover. Tag every such sentence with [GENERIC].\n", genericPct))
	}
	sb.WriteString("\nCite training with [CHUNK_xxxxxxxx] and peer knowledge with [PEER:Name],\n")
	sb.WriteString("exactly as the markers appear above. Refer to design files by their path.\n")

	return sb.String()
}

// ProposeChange stores a change request from this chat and posts it on the
// blackboard so the runner can pick it up.
//
// This is the entry point for the coordinated redesign flow:
//   Client types change in chat -> ProposeChange -> ChangeRequestService.ProposeChange
//   -> blackboard event -> runner.watchForChangeRequest detects it
//   -> relevant experts re-run design sections -> approval gate
//
// The chat message is stored first (same pattern as Send) so the client's
// intent is never lost even if the downstream steps fail.
func (s *WorkflowChatService) ProposeChange(
	ctx context.Context,
	chatID, clientID uuid.UUID,
	changeGoal string,
	crSvc *ChangeRequestService,
) (*ChangeRequest, error) {
	changeGoal = strings.TrimSpace(changeGoal)
	if changeGoal == "" {
		return nil, fmt.Errorf("propose change: change_goal is required")
	}

	ch, err := s.GetChat(ctx, chatID, clientID)
	if err != nil {
		return nil, err
	}

	turn, err := s.nextTurnNumber(ctx, chatID)
	if err != nil {
		return nil, err
	}

	// Store the client's change message in the chat history.
	var msgID uuid.UUID
	err = s.db.QueryRow(ctx,
		`INSERT INTO workflow_chat_messages (chat_id, role, content, turn_number)
		 VALUES ($1, 'user', $2, $3)
		 RETURNING id`,
		chatID, "[CHANGE REQUEST] "+changeGoal, turn,
	).Scan(&msgID)
	if err != nil {
		return nil, fmt.Errorf("propose change: save message: %w", err)
	}

	// Delegate to ChangeRequestService which owns the change_requests table.
	cr, err := crSvc.ProposeChange(ctx, ch.WorkflowID, clientID, changeGoal, &chatID, &msgID)
	if err != nil {
		return nil, fmt.Errorf("propose change: %w", err)
	}

	// Acknowledge in the chat so the client sees the request was received.
	if _, err := s.db.Exec(ctx,
		`INSERT INTO workflow_chat_messages (chat_id, role, content, turn_number)
		 VALUES ($1, 'assistant', $2, $3)`,
		chatID,
		fmt.Sprintf("Your change request has been received (ID: %s). "+
			"The relevant experts will re-run their design sections. "+
			"You will be asked to approve the updated design when ready.", cr.ID),
		turn,
	); err != nil {
		// Non-fatal: the change request was created. A missing ack message
		// is a UX gap, not a data loss.
		s.logger.Warn("propose change: ack message failed (non-fatal)",
			zap.String("chat_id", chatID.String()), zap.Error(err))
	}

	// Update message count (question + ack = 2 rows).
	if _, err := s.db.Exec(ctx,
		`UPDATE workflow_chats SET message_count = message_count + 2, updated_at = NOW()
		 WHERE id = $1`, chatID,
	); err != nil {
		s.logger.Warn("propose change: message_count update failed (non-fatal)",
			zap.String("chat_id", chatID.String()), zap.Error(err))
	}

	s.logger.Info("change request proposed from chat",
		zap.String("chat_id", chatID.String()),
		zap.String("workflow_id", ch.WorkflowID.String()),
		zap.String("change_request_id", cr.ID.String()),
	)
	return cr, nil
}

// buildUserPrompt puts the deliverable first, then the design map, then history,
// then the question.
func (s *WorkflowChatService) buildUserPrompt(deliverable, sections, history, question string) string {
	var sb strings.Builder

	if deliverable != "" {
		sb.WriteString(deliverable)
		sb.WriteString("\n")
	}
	if sections != "" {
		sb.WriteString(sections)
		sb.WriteString("\n")
	}
	if history != "" {
		sb.WriteString("=== CONVERSATION SO FAR ===\n")
		sb.WriteString(history)
		sb.WriteString("\n")
	}
	sb.WriteString("=== THE CLIENT'S QUESTION ===\n")
	sb.WriteString(question)

	return sb.String()
}
