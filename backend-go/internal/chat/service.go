package chat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/response"
)

// Chat is a single chat window within a project.
type Chat struct {
	ID           uuid.UUID `json:"id"`
	ProjectID    uuid.UUID `json:"project_id"`
	ClientID     uuid.UUID `json:"client_id"`
	Title        string    `json:"title"`
	MessageCount int       `json:"message_count"`
	IsArchived   bool      `json:"is_archived"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Message is a single message in a chat.
type Message struct {
	ID                   uuid.UUID  `json:"id"`
	ChatID               uuid.UUID  `json:"chat_id"`
	Role                 string     `json:"role"`
	Content              string     `json:"content"`
	TurnNumber           int        `json:"turn_number"`
	ExpertID             *uuid.UUID `json:"expert_id,omitempty"`
	DecisionMode         string     `json:"decision_mode,omitempty"`
	Citations            interface{} `json:"citations,omitempty"`
	Confidence           float64    `json:"confidence,omitempty"`
	TokensUsed           int        `json:"tokens_used,omitempty"`
	CostUSD              float64    `json:"cost_usd,omitempty"`
	// WarningText is the Gate 3 WARN reasoning text.
	// NULL/empty means no charter warning was triggered.
	WarningText          string     `json:"warning_text,omitempty"`
	// ClarifyingQuestions is the Gate 1 ASK question list.
	// Empty means no clarification was needed.
	ClarifyingQuestions  []string   `json:"clarifying_questions,omitempty"`
	// ReplyToMessageID (CT-C1, migration 010): immediate parent message
	// this one replies to. nil = fresh question, not a reply (the vast
	// majority of messages, both before and after this feature). Optional
	// on this struct so every existing SaveMessage/ListMessages call site
	// that does not set it keeps working unchanged — zero regression.
	ReplyToMessageID     *uuid.UUID `json:"reply_to_message_id,omitempty"`
	// TemplateSections (migration 011, 2026-09-08 RCA round 7): JSONB
	// mirroring chinawall.TemplateSectionResult. nil for every flat-text
	// message (CT-L2). Stored as interface{} (raw JSON), not a typed
	// slice, matching this struct's existing convention for Citations
	// above - chat/service.go intentionally has no dependency on the
	// chinawall package, so the concrete shape lives at the call site
	// (message/handler.go) that already imports chinawall.
	TemplateSections     interface{} `json:"template_sections,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
}

// ErrNotFound is returned when a chat is not found.
var ErrNotFound = errors.New("chat not found")

// Service handles chat business logic.
type Service struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

// NewService creates a new chat service.
func NewService(db *pgxpool.Pool, logger *zap.Logger) *Service {
	return &Service{db: db, logger: logger}
}

// GetByID returns a chat, verifying client ownership.
func (s *Service) GetByID(ctx context.Context, chatID, clientID uuid.UUID) (*Chat, error) {
	var ch Chat
	err := s.db.QueryRow(ctx,
		`SELECT id, project_id, client_id, title, message_count, is_archived, created_at, updated_at
		 FROM chats
		 WHERE id=$1 AND client_id=$2`,
		chatID, clientID,
	).Scan(&ch.ID, &ch.ProjectID, &ch.ClientID, &ch.Title,
		&ch.MessageCount, &ch.IsArchived, &ch.CreatedAt, &ch.UpdatedAt)
	if err != nil {
		return nil, ErrNotFound
	}
	return &ch, nil
}

// UpdateTitle updates the chat title.
func (s *Service) UpdateTitle(ctx context.Context, chatID, clientID uuid.UUID, title string) error {
	result, err := s.db.Exec(ctx,
		`UPDATE chats SET title=$1, updated_at=NOW() WHERE id=$2 AND client_id=$3`,
		title, chatID, clientID,
	)
	if err != nil {
		return fmt.Errorf("update chat failed: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// Archive marks a chat as archived (soft disable).
func (s *Service) Archive(ctx context.Context, chatID, clientID uuid.UUID) error {
	result, err := s.db.Exec(ctx,
		`UPDATE chats SET is_archived=TRUE, updated_at=NOW() WHERE id=$1 AND client_id=$2`,
		chatID, clientID,
	)
	if err != nil {
		return fmt.Errorf("archive chat failed: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// Unarchive restores an archived chat to active.
func (s *Service) Unarchive(ctx context.Context, chatID, clientID uuid.UUID) error {
	result, err := s.db.Exec(ctx,
		`UPDATE chats SET is_archived=FALSE, updated_at=NOW() WHERE id=$1 AND client_id=$2`,
		chatID, clientID,
	)
	if err != nil {
		return fmt.Errorf("unarchive chat failed: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// PermanentDelete hard-deletes a chat and all its messages.
// WHY hard delete: user explicitly confirmed deletion.
// Messages are cascade-deleted by FK constraint.
func (s *Service) PermanentDelete(ctx context.Context, chatID, clientID uuid.UUID) error {
	// Verify ownership before deleting
	var exists bool
	s.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM chats WHERE id=$1 AND client_id=$2)`,
		chatID, clientID,
	).Scan(&exists)
	if !exists {
		return ErrNotFound
	}
	_, err := s.db.Exec(ctx,
		`DELETE FROM chats WHERE id=$1 AND client_id=$2`,
		chatID, clientID,
	)
	if err != nil {
		return fmt.Errorf("permanent delete chat failed: %w", err)
	}
	return nil
}

// DeleteMessage hard-deletes a single user message.
func (s *Service) DeleteMessage(ctx context.Context, messageID, clientID uuid.UUID) error {
	var exists bool
	s.db.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM messages m
			JOIN chats c ON c.id = m.chat_id
			WHERE m.id=$1 AND c.client_id=$2 AND m.role='user'
		)`,
		messageID, clientID,
	).Scan(&exists)
	if !exists {
		return ErrNotFound
	}
	_, err := s.db.Exec(ctx, `DELETE FROM messages WHERE id=$1`, messageID)
	if err != nil {
		return fmt.Errorf("delete message failed: %w", err)
	}
	return nil
}

// UpdateMessageContent edits the text of a user message.
func (s *Service) UpdateMessageContent(ctx context.Context, messageID, clientID uuid.UUID, content string) error {
	if content == "" {
		return fmt.Errorf("content cannot be empty")
	}
	result, err := s.db.Exec(ctx,
		`UPDATE messages m SET content=$1
		 FROM chats c
		 WHERE m.id=$2 AND m.chat_id=c.id AND c.client_id=$3 AND m.role='user'`,
		content, messageID, clientID,
	)
	if err != nil {
		return fmt.Errorf("update message failed: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// GetNextTurnNumber returns the next turn number for a chat.
// WHY: Turn number tracks conversation position (one number per user question
// + all expert replies in that round). Used by memory manager and the
// rolling-summary trigger (every 10 turns).
//
// WHY MAX(turn_number) not message_count+1:
//   message_count increments once per row (user + each assistant), so the
//   old formula produced 1,3,5,7… and turnNumber%10==0 never fired —
//   rolling summaries never ran. MAX of the turn column advances by exactly
//   one per completed round.
func (s *Service) GetNextTurnNumber(ctx context.Context, chatID uuid.UUID) (int, error) {
	var maxTurn int
	err := s.db.QueryRow(ctx,
		`SELECT COALESCE(MAX(turn_number), 0) FROM messages WHERE chat_id=$1`,
		chatID,
	).Scan(&maxTurn)
	if err != nil {
		return 0, err
	}
	return maxTurn + 1, nil
}

// IncrementMessageCount increments the message count for a chat.
func (s *Service) IncrementMessageCount(ctx context.Context, chatID uuid.UUID) error {
	_, err := s.db.Exec(ctx,
		`UPDATE chats SET message_count=message_count+1, updated_at=NOW() WHERE id=$1`,
		chatID,
	)
	return err
}

// SaveMessage saves a message to the database.
// Includes warning_text and clarifying_questions so Gate 3 WARN
// and Gate 1 ASK data survive page reloads.
//
// Mental execution:
// msg.ClarifyingQuestions = ["What scale?", "What team size?"]
// pgx marshals []string -> '["What scale?","What team size?"]'::jsonb
// On read: pgx unmarshals jsonb -> []string automatically
func (s *Service) SaveMessage(ctx context.Context, msg Message) (uuid.UUID, error) {
	// Ensure clarifying_questions is never nil (DB default is '[]')
	if msg.ClarifyingQuestions == nil {
		msg.ClarifyingQuestions = []string{}
	}
	if msg.Citations == nil {
		msg.Citations = []interface{}{}
	}

	// decision_mode must be NULL or one of the allowed values.
	// User messages have no decision_mode — empty string "" violates
	// messages_mode_check (SQLSTATE 23514). Convert to nil so pgx
	// sends NULL, which the constraint allows.
	var decisionMode interface{}
	if msg.DecisionMode != "" {
		decisionMode = msg.DecisionMode
	}
	// Same for WarningText — empty string is fine for TEXT columns,
	// but NULL is cleaner and consistent with the intent.
	var warningText interface{}
	if msg.WarningText != "" {
		warningText = msg.WarningText
	}

	var id uuid.UUID
	err := s.db.QueryRow(ctx,
		`INSERT INTO messages
			(chat_id, role, content, turn_number, expert_id, decision_mode,
			 confidence, warning_text, clarifying_questions, citations,
			 reply_to_message_id, template_sections)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		 RETURNING id`,
		msg.ChatID, msg.Role, msg.Content, msg.TurnNumber,
		msg.ExpertID, decisionMode, msg.Confidence,
		warningText, msg.ClarifyingQuestions, msg.Citations, msg.ReplyToMessageID,
		msg.TemplateSections,
	).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("save message failed: %w", err)
	}
	return id, nil
}

// ListMessages returns paginated messages for a chat.
func (s *Service) ListMessages(ctx context.Context, chatID, clientID uuid.UUID, limit, offset int) ([]Message, error) {
	// Verify ownership
	var exists bool
	s.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM chats WHERE id=$1 AND client_id=$2)`,
		chatID, clientID,
	).Scan(&exists)
	if !exists {
		return nil, ErrNotFound
	}

	rows, err := s.db.Query(ctx,
		`SELECT id, chat_id, role, content, turn_number,
		        expert_id, COALESCE(decision_mode,''), COALESCE(confidence,0),
		        COALESCE(tokens_used,0), COALESCE(cost_usd,0),
		        COALESCE(warning_text,''),
		        COALESCE(clarifying_questions,'[]'::jsonb),
		        COALESCE(citations,'[]'::jsonb),
		        reply_to_message_id,
		        template_sections,
		        created_at
		 FROM messages
		 WHERE chat_id=$1
		 ORDER BY turn_number ASC
		 LIMIT $2 OFFSET $3`,
		chatID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var m Message
		// Bug 3.4 fix: citations JSONB is now selected and scanned back
		// into m.Citations (as []map[string]interface{}, matching the
		// pattern already used for clarifying_questions below).
		var clarifyingJSON, citationsJSON, templateSectionsJSON []byte
		if err := rows.Scan(
			&m.ID, &m.ChatID, &m.Role, &m.Content, &m.TurnNumber,
			&m.ExpertID, &m.DecisionMode, &m.Confidence,
			&m.TokensUsed, &m.CostUSD,
			&m.WarningText, &clarifyingJSON, &citationsJSON,
			&m.ReplyToMessageID,
			&templateSectionsJSON,
			&m.CreatedAt,
		); err != nil {
			continue
		}
		// Unmarshal clarifying_questions JSONB -> []string
		if len(clarifyingJSON) > 0 {
			_ = json.Unmarshal(clarifyingJSON, &m.ClarifyingQuestions)
		}
		if m.ClarifyingQuestions == nil {
			m.ClarifyingQuestions = []string{}
		}
		// Unmarshal citations JSONB -> generic slice (interface{} field)
		if len(citationsJSON) > 0 {
			var cit []map[string]interface{}
			_ = json.Unmarshal(citationsJSON, &cit)
			m.Citations = cit
		}
		// Migration 011 (2026-09-08 RCA round 7): unmarshal
		// template_sections JSONB -> generic slice, same pattern as
		// citations above. NULL column (flat-text message, or any row
		// saved before migration 011) leaves templateSectionsJSON empty
		// and m.TemplateSections stays nil - the frontend adapter's
		// existing `templateSections && length > 0` check already
		// handles nil/undefined correctly, zero regression for flat
		// messages.
		if len(templateSectionsJSON) > 0 {
			var sections []map[string]interface{}
			_ = json.Unmarshal(templateSectionsJSON, &sections)
			m.TemplateSections = sections
		}
		messages = append(messages, m)
	}
	if messages == nil {
		messages = []Message{}
	}
	return messages, nil
}

// IndexTurn saves a turn to the chat_index for semantic search.
// Called async after each turn.
// WHY: Chat index enables semantic search in conversation history.
// Context assembler uses this to find relevant past turns.
//
// WHY pgvector.NewVector not string interpolation (Bug 7 fix):
// String interpolation is SQL injection risk and breaks on NaN/Inf.
// pgvector.NewVector sends embedding as binary protocol parameter.
// Same pattern used correctly in l2_store.go.
func (s *Service) IndexTurn(
	ctx context.Context,
	chatID uuid.UUID,
	messageID uuid.UUID,
	turnNumber int,
	summary string,
	topic string,
	importance int,
	embedding []float32,
) error {
	if len(embedding) == 0 {
		// ML sidecar unavailable — store with zero vector
		// WHY not skip: chat_index entry is needed for turn tracking
		// even without semantic search capability
		_, err := s.db.Exec(ctx,
			`INSERT INTO chat_index
				(chat_id, message_id, turn_number, one_line_summary, topic, importance, embedding)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			chatID, messageID, turnNumber, summary, topic, importance,
			pgvector.NewVector(make([]float32, 768)), // zero vector
		)
		return err
	}

	_, err := s.db.Exec(ctx,
		`INSERT INTO chat_index
			(chat_id, message_id, turn_number, one_line_summary, topic, importance, embedding)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		chatID, messageID, turnNumber, summary, topic, importance,
		pgvector.NewVector(embedding),
	)
	return err
}

// GenerateRollingSummary generates and stores a rolling summary.
// Called every 10 turns.
// WHY rolling summary (Arpit Bhiyani + Byte by Byte AI):
// Context window is finite. Summary compresses history.
// Prevents lost-in-middle problem for long conversations.
func (s *Service) SaveRollingSummary(
	ctx context.Context,
	chatID uuid.UUID,
	summaryText string,
	turnStart, turnEnd int,
) error {
	_, err := s.db.Exec(ctx,
		`INSERT INTO chat_summaries
			(chat_id, summary_text, turn_range_start, turn_range_end)
		 VALUES ($1, $2, $3, $4)`,
		chatID, summaryText, turnStart, turnEnd,
	)
	return err
}

// GetDB returns the database pool.
// Used by message handler for direct queries.
func (s *Service) GetDB() *pgxpool.Pool {
	return s.db
}

// Handler handles HTTP requests for chats.
type Handler struct {
	svc    *Service
	logger *zap.Logger
}

// NewHandler creates a new chat handler.
func NewHandler(svc *Service, logger *zap.Logger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

// GetByID GET /chats/:id
func (h *Handler) GetByID(c *gin.Context) {
	clientID := c.MustGet("user_id").(uuid.UUID)
	chatID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid chat ID")
		return
	}
	ch, err := h.svc.GetByID(c.Request.Context(), chatID, clientID)
	if err != nil {
		response.NotFound(c, "chat")
		return
	}
	response.OK(c, ch)
}

// Update PATCH /chats/:id
func (h *Handler) Update(c *gin.Context) {
	clientID := c.MustGet("user_id").(uuid.UUID)
	chatID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid chat ID")
		return
	}
	var req struct {
		Title string `json:"title" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}
	if err := h.svc.UpdateTitle(c.Request.Context(), chatID, clientID, req.Title); err != nil {
		if errors.Is(err, ErrNotFound) {
			response.NotFound(c, "chat")
			return
		}
		response.InternalError(c)
		return
	}
	response.OK(c, map[string]string{"status": "updated"})
}

// Archive DELETE /chats/:id — soft disable (is_archived=true)
func (h *Handler) Archive(c *gin.Context) {
	clientID := c.MustGet("user_id").(uuid.UUID)
	chatID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid chat ID")
		return
	}
	if err := h.svc.Archive(c.Request.Context(), chatID, clientID); err != nil {
		if errors.Is(err, ErrNotFound) {
			response.NotFound(c, "chat")
			return
		}
		response.InternalError(c)
		return
	}
	response.OK(c, map[string]string{"status": "archived"})
}

// Unarchive POST /chats/:id/unarchive — restore archived chat to active
func (h *Handler) Unarchive(c *gin.Context) {
	clientID := c.MustGet("user_id").(uuid.UUID)
	chatID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid chat ID")
		return
	}
	if err := h.svc.Unarchive(c.Request.Context(), chatID, clientID); err != nil {
		if errors.Is(err, ErrNotFound) {
			response.NotFound(c, "chat")
			return
		}
		response.InternalError(c)
		return
	}
	response.OK(c, map[string]string{"status": "active"})
}

// PermanentDelete DELETE /chats/:id/permanent — hard delete with cascade
func (h *Handler) PermanentDelete(c *gin.Context) {
	clientID := c.MustGet("user_id").(uuid.UUID)
	chatID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid chat ID")
		return
	}
	if err := h.svc.PermanentDelete(c.Request.Context(), chatID, clientID); err != nil {
		if errors.Is(err, ErrNotFound) {
			response.NotFound(c, "chat")
			return
		}
		response.InternalError(c)
		return
	}
	response.OK(c, map[string]string{"status": "deleted"})
}

// ListMessages GET /chats/:id/messages
func (h *Handler) ListMessages(c *gin.Context) {
	clientID := c.MustGet("user_id").(uuid.UUID)
	chatID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid chat ID")
		return
	}
	messages, err := h.svc.ListMessages(c.Request.Context(), chatID, clientID, 50, 0)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.NotFound(c, "chat")
			return
		}
		response.InternalError(c)
		return
	}
	response.OK(c, messages)
}
