package memory

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/ml"
)

// Manager is the single interface for all memory operations.
// Coordinates L1 (Redis), L2 (PostgreSQL), and L3 (event log).
//
// Usage pattern:
// 1. Before turn: GetProjectContext() — load what experts know
// 2. After turn: RecordTurn() — update all three levels
//
// WHY single Manager:
// Callers don't need to know which level stores what.
// Manager decides: hot data → L1, group data → L2, all events → L3.
type Manager struct {
	l1     *L1Store
	l2     *L2Store
	l3     *L3Store
	logger *zap.Logger
}

// NewManager creates a new memory manager.
// embedder satisfies ml.Embedder — either *ml.SidecarClient (default) or
// *ml.DynamicEmbedder (when CodeCraftAPI embeddings are enabled).
// Passed through to L2Store which uses it for semantic search embeddings.
func NewManager(
	db *pgxpool.Pool,
	redisClient *redis.Client,
	embedder ml.Embedder,
	logger *zap.Logger,
) *Manager {
	return &Manager{
		l1:     NewL1Store(redisClient, logger),
		l2:     NewL2Store(db, embedder, logger),
		l3:     NewL3Store(db, logger),
		logger: logger,
	}
}

// ProjectContext holds all memory loaded for a turn.
type ProjectContext struct {
	L1Memory       *L1Memory  // Hot: this expert's recent decisions
	L2Entries      []L2Entry  // Group: all experts' relevant decisions
	L1CacheHit     bool       // Was L1 from cache or rebuilt?
}

// GetProjectContext loads all relevant memory for an expert's turn.
// Called by Context Assembler before every LLM call.
//
// Mental execution:
// Expert: DB Expert, Project: E-Commerce, Query: "should I shard?"
// 1. L1 Get — DB Expert's last 5 decisions for this project (Redis, <1ms)
// 2. L2 Search — all experts' decisions related to "sharding" (pgvector, ~5ms)
// Result: DB Expert knows SD Expert said "use PostgreSQL" at turn 8
func (m *Manager) GetProjectContext(
	ctx context.Context,
	projectID uuid.UUID,
	expertID uuid.UUID,
	query string,
) (*ProjectContext, error) {
	// L1: always fast
	l1, err := m.l1.Get(ctx, projectID, expertID)
	if err != nil {
		m.logger.Warn("L1 get failed", zap.Error(err))
		l1 = &L1Memory{ProjectID: projectID, ExpertID: expertID}
	}

	// L2: semantic search for relevant project context
	l2, err := m.l2.SearchByQuery(ctx, projectID, query, 8)
	if err != nil {
		m.logger.Warn("L2 search failed", zap.Error(err))
		l2 = []L2Entry{}
	}

	return &ProjectContext{
		L1Memory:   l1,
		L2Entries:  l2,
		L1CacheHit: l1.LastUpdated != time.Time{},
	}, nil
}

// RecordTurn updates all memory levels after a completed turn.
// Called async after response is sent to client.
//
// WHY async: Memory update should not block the response.
// If memory update fails, the response was already sent — log and continue.
func (m *Manager) RecordTurn(
	ctx context.Context,
	projectID uuid.UUID,
	expertID uuid.UUID,
	clientID uuid.UUID,
	// chatID is a POINTER (B3). Standalone chat passes &req.ChatID;
	// the workflow path has no chat row, so it passes nil. Matches
	// master_event_log.chat_id which is nullable (no NOT NULL) —
	// same honest-nil pattern as messageID below. Never fabricate a
	// chat UUID that does not exist in the chats table.
	chatID *uuid.UUID,
	// messageID is a POINTER, not a value - Bug 3.2 fix (docs bug list).
	// The caller (orchestrator.updateMemory) genuinely does not know the
	// real messages.id at this point: the assistant message is saved by
	// message/handler.go's saveAssistantMessage in a SEPARATE goroutine
	// that may not have run yet. The previous code passed uuid.New() -
	// a fabricated ID that never exists in the messages table - which
	// violated master_event_log's `message_id UUID REFERENCES
	// messages(id)` FK on every single turn. That column is nullable in
	// the schema (no NOT NULL) precisely for cases like this - passing
	// nil here is the honest, schema-correct fix, not a fabricated ID.
	messageID *uuid.UUID,
	turnNumber int,
	userMessage string,
	assistantResponse string,
	decisionMode string,
	importance int,
) {
	// Only record high-importance turns in L1/L2
	// WHY threshold 3: Low-importance turns (clarifications, small details)
	// don't need to be in memory. Saves space and keeps memory relevant.
	if importance >= 3 {
		// Update L1 (async, non-blocking)
		go func() {
			bgCtx := context.Background()
			_ = m.l1.Append(bgCtx, projectID, expertID, L1Decision{
				Content:    userMessage[:minInt(200, len(userMessage))],
				Reasoning:  assistantResponse[:minInt(300, len(assistantResponse))],
				TurnNumber: turnNumber,
				MemoryType: decisionMode,
				Importance: importance,
			})
		}()

		// Update L2 (async, non-blocking)
		go func() {
			bgCtx := context.Background()
			_ = m.l2.Append(bgCtx, L2Entry{
				ProjectID:     projectID,
				ExpertID:      expertID,
				MemoryType:    "decision",
				Content:       userMessage[:minInt(500, len(userMessage))],
				Context:       assistantResponse[:minInt(500, len(assistantResponse))],
				TurnReference: turnNumber,
				Importance:    importance,
			})
		}()
	}

	// L3: always record every turn (append-only log)
	go func() {
		bgCtx := context.Background()
		expertIDPtr := &expertID
		_ = m.l3.Append(bgCtx, L3Event{
			ProjectID: projectID,
			ExpertID:  expertIDPtr,
			ClientID:  clientID,
			ChatID:    chatID, // already *uuid.UUID, nil-safe (chat OR workflow)
			MessageID: messageID, // already *uuid.UUID, nil-safe
			EventType: EventResponseGenerated,
			EventData: map[string]interface{}{
				"turn_number":   turnNumber,
				"decision_mode": decisionMode,
				"importance":    importance,
			},
			Reasoning:    assistantResponse[:minInt(200, len(assistantResponse))],
			DecisionMade: decisionMode,
		})
	}()
}

// ProjectMemoryEntry is an L2 group-memory record enriched with the
// expert's display name, for the project's consolidated cross-expert
// decisions view (feature gap #5, docs bug list).
type ProjectMemoryEntry struct {
	ID            uuid.UUID `json:"id"`
	ProjectID     uuid.UUID `json:"projectId"`
	ExpertID      uuid.UUID `json:"expertId"`
	ExpertName    string    `json:"expertName"`
	MemoryType    string    `json:"memoryType"`
	Content       string    `json:"content"`
	Context       string    `json:"context"`
	TurnReference int       `json:"turnReference"`
	Importance    int       `json:"importance"`
	IsSuperseded  bool      `json:"isSuperseded"`
	CreatedAt     time.Time `json:"createdAt"`
}

// GetProjectMemory returns L2 group-memory entries (decisions any
// expert made in this project), enriched with expert names.
//
// Feature #5 fix (docs bug list): GET /projects/:id/memory already
// existed as a registered route, but its handler
// (handleGetProjectMemory, cmd/server/main.go) called GetTimeline -
// the exact same L3 event-log method /timeline calls, just with
// limit=20 instead of 50. It never queried project_memory_l2 at all,
// despite L2Store.GetRecent already existing and doing the right
// query (just without the expert-name join a UI needs). This is the
// real L2 fetch; main.go's handler is repointed to call it in the
// same commit as this change.
//
// WHY camelCase JSON tags directly (not snake_case): this endpoint
// was never in openapi.yaml (found missing during the Interface-First
// audit, docs/INTERFACE_FIRST_CONTRACT.md) - since it needs adding to
// the spec anyway, it is added camelCase from the start, matching the
// §4 locked decision (Go owns wire casing directly, no frontend bridge).
func (m *Manager) GetProjectMemory(ctx context.Context, projectID uuid.UUID, limit int) ([]ProjectMemoryEntry, error) {
	rows, err := m.l2.db.Query(ctx,
		`SELECT l2.id, l2.project_id, l2.expert_id, e.name, l2.memory_type, l2.content,
		        COALESCE(l2.context,''), COALESCE(l2.turn_reference,0),
		        l2.importance, l2.is_superseded, l2.created_at
		 FROM project_memory_l2 l2
		 JOIN experts e ON e.id = l2.expert_id
		 WHERE l2.project_id=$1 AND l2.is_superseded=FALSE
		 ORDER BY l2.importance DESC, l2.created_at DESC
		 LIMIT $2`,
		projectID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []ProjectMemoryEntry
	for rows.Next() {
		var e ProjectMemoryEntry
		if err := rows.Scan(
			&e.ID, &e.ProjectID, &e.ExpertID, &e.ExpertName, &e.MemoryType, &e.Content,
			&e.Context, &e.TurnReference, &e.Importance, &e.IsSuperseded, &e.CreatedAt,
		); err != nil {
			continue
		}
		entries = append(entries, e)
	}
	if entries == nil {
		entries = []ProjectMemoryEntry{}
	}
	return entries, nil
}

// RecordViolation logs a China Wall violation to L3.
func (m *Manager) RecordViolation(
	ctx context.Context,
	projectID uuid.UUID,
	expertID uuid.UUID,
	clientID uuid.UUID,
	violationType string,
	details string,
) {
	go func() {
		bgCtx := context.Background()
		expertIDPtr := &expertID
		_ = m.l3.Append(bgCtx, L3Event{
			ProjectID: projectID,
			ExpertID:  expertIDPtr,
			ClientID:  clientID,
			EventType: EventChinaWallViolation,
			EventData: map[string]interface{}{
				"violation_type": violationType,
				"details":        details,
			},
			Reasoning: details,
		})
	}()
}

// RecordRating logs a rating event to L3.
// WHY separate from RecordViolation:
// A rating is client feedback, not a China Wall violation.
// Using RecordViolation for ratings polluted the admin violations list.
// Separate method ensures correct EventType in master_event_log.
func (m *Manager) RecordRating(
	ctx context.Context,
	projectID uuid.UUID,
	expertID uuid.UUID,
	clientID uuid.UUID,
	chatID uuid.UUID,
	messageID uuid.UUID,
	score int,
	feedbackType string,
) {
	go func() {
		bgCtx := context.Background()
		expertIDPtr := &expertID
		chatIDPtr := &chatID
		msgIDPtr := &messageID
		_ = m.l3.Append(bgCtx, L3Event{
			ProjectID: projectID,
			ExpertID:  expertIDPtr,
			ClientID:  clientID,
			ChatID:    chatIDPtr,
			MessageID: msgIDPtr,
			EventType: EventRatingRecorded,
			EventData: map[string]interface{}{
				"score":         score,
				"feedback_type": feedbackType,
			},
			Reasoning:    fmt.Sprintf("Client rated response %d/5", score),
			DecisionMade: feedbackType,
		})
	}()
}

// GetTimeline returns project event timeline for UI.
func (m *Manager) GetTimeline(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]L3EventRecord, error) {
	return m.l3.GetProjectTimeline(ctx, projectID, limit, offset)
}

// GetViolations returns China Wall violations for admin panel.
func (m *Manager) GetViolations(ctx context.Context, limit int) ([]L3EventRecord, error) {
	return m.l3.GetViolations(ctx, limit)
}

// minInt returns the smaller of two ints.
// Named minInt to avoid conflict with Go 1.21+ builtin min.
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
