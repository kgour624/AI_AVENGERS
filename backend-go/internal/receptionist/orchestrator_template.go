package receptionist

import (
    "context"
    "fmt"
    "github.com/google/uuid"
    "github.com/jackc/pgx/v5"
    "go.uber.org/zap"
)

type TemplateOrchestrator struct {
    store     *Store
    search    SearchProvider
    publisher *Publisher
    llm       LLMClient
    logger    *zap.Logger
}

func NewTemplateOrchestrator(store *Store, search SearchProvider, publisher *Publisher, llm LLMClient, logger *zap.Logger) *TemplateOrchestrator {
    return &TemplateOrchestrator{store: store, search: search, publisher: publisher, llm: llm, logger: logger}
}

func (o *TemplateOrchestrator) WithLLM(llm LLMClient) *TemplateOrchestrator {
    o.llm = llm
    return o
}

// EnsureTemplate - Empty -> Search -> Save -> SSE notify (Postgres truth + Redis notify)
func (o *TemplateOrchestrator) EnsureTemplate(ctx context.Context, sessionID, checkpointID, tenantID uuid.UUID, idx int, agenda string) ([]Point, error) {
    // 1. Postgres truth check
    pts, _, err := o.store.GetTemplate(ctx, checkpointID)
    if err == nil && len(pts) > 0 {
        o.logger.Info("template exists", zap.Int("count", len(pts)))
        return pts, nil // already exists, reuse
    }
    if err != pgx.ErrNoRows {
        o.logger.Error("get template error", zap.Error(err))
    }

    // 2. Typing event before search - Phase 2 pipe reuse
    o.store.AppendEvent(ctx, sessionID, tenantID, "typing", map[string]any{"msg": "Search kar rahi hu, 2026 standard se points la rahi hu..."})
    o.publisher.NotifyNewEvent(ctx, sessionID.String(), 0)

    // 3. Cheap search - dynamic, pre-store nahi
    query := fmt.Sprintf("%s required sections checklist", agenda) // agenda = "LLD", "PRD for JIRA" anything
    pts, err = o.search.Search(ctx, query)
    if err != nil {
        o.logger.Error("search failed fallback empty", zap.Error(err))
        return nil, err
    }

    // 4. Postgres upsert (source of truth)
    if err := o.store.UpsertTemplate(ctx, checkpointID, sessionID, tenantID, idx, query, pts); err != nil {
        return nil, err
    }

    // 5. SSE notify - frontend ko template_ready
    ev, _ := o.store.AppendEvent(ctx, sessionID, tenantID, "template_ready", map[string]any{"checkpoint_id": checkpointID, "idx": idx, "points": pts})
    o.publisher.NotifyNewEvent(ctx, sessionID.String(), ev.ID)
    o.logger.Info("template created", zap.Int("points", len(pts)), zap.String("query", query))

    return pts, nil
}

// OnNextCheckpoint - purana clear + naya search trigger (tune bola: har checkpoint pe clear)
func (o *TemplateOrchestrator) OnNextCheckpoint(ctx context.Context, oldCheckpointID, newCheckpointID, sessionID, tenantID uuid.UUID, newIdx int, agenda string) ([]Point, error) {
    // clear old (best effort)
    _ = o.store.DeleteTemplate(ctx, oldCheckpointID)
    o.store.AppendEvent(ctx, sessionID, tenantID, "phase_change", map[string]any{"msg": "Purana checklist clear, naya la rahi hu..."})
    return o.EnsureTemplate(ctx, sessionID, newCheckpointID, tenantID, newIdx, agenda)
}
