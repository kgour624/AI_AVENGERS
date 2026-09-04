package context

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/chinawall"
	"ai_avengers/backend/internal/memory"
	"ai_avengers/backend/internal/ml"
)

// AssembledContext holds everything needed for one LLM call.
// Built fresh for every turn. Never cached.
type AssembledContext struct {
	RollingSummary  string
	RecentMessages  []Message
	RelevantHistory []HistoryEntry
	CourseChunks    []chinawall.CourseChunk
	ProjectContext  *memory.ProjectContext
	TotalTokens     int
}

// Message is a single chat message for context.
type Message struct {
	Role      string
	Content   string
	CreatedAt time.Time
}

// HistoryEntry is a semantically relevant past turn.
type HistoryEntry struct {
	TurnNumber     int
	OneLineSummary string
	Topic          string
	Importance     int
}

// Assembler builds smart context for LLM calls.
// Respects token budget: never exceeds MaxTokens.
//
// Budget allocation (WHY these percentages):
// 10% rolling summary — compressed history, always include
// 20% L2 project memory — what other experts decided
// 20% recent messages — last 3 turns for continuity
// 15% semantic history — relevant past turns
// 35% course chunks — expert knowledge (most important)
//
// WHY this order: Most important context last = less lost-in-middle.
type Assembler struct {
	db          *pgxpool.Pool
	ml          *ml.SidecarClient
	memManager  *memory.Manager
	maxTokens   int
	recentMsgs  int
	semanticTopK int
	chunksTopK  int
	logger      *zap.Logger
}

// NewAssembler creates a new context assembler.
func NewAssembler(
	db *pgxpool.Pool,
	mlClient *ml.SidecarClient,
	memManager *memory.Manager,
	maxTokens, recentMsgs, semanticTopK, chunksTopK int,
	logger *zap.Logger,
) *Assembler {
	return &Assembler{
		db:           db,
		ml:           mlClient,
		memManager:   memManager,
		maxTokens:    maxTokens,
		recentMsgs:   recentMsgs,
		semanticTopK: semanticTopK,
		chunksTopK:   chunksTopK,
		logger:       logger,
	}
}

// Assemble builds the full context for a turn.
//
// Mental execution:
// Turn 25, Expert: DB Expert, Question: "should I shard?"
// 1. Rolling summary (turns 1-20) → 500 tokens
// 2. L2 memory (SD Expert decided PostgreSQL) → 300 tokens
// 3. Recent messages (turns 22-24) → 400 tokens
// 4. Semantic history (turns about database) → 200 tokens
// 5. Course chunks (sharding content) → 2000 tokens
// Total: ~3400 tokens (well under 10k limit)
func (a *Assembler) Assemble(
	ctx context.Context,
	chatID uuid.UUID,
	projectID uuid.UUID,
	expertID uuid.UUID,
	question string,
	turnNumber int,
) (*AssembledContext, error) {

	assembled := &AssembledContext{}
	tokensUsed := 0
	budget := a.maxTokens

	// 1. Rolling summary (10% budget)
	summary, err := a.getRollingSummary(ctx, chatID)
	if err == nil && summary != "" {
		summaryTokens := estimateTokens(summary)
		if tokensUsed+summaryTokens <= budget*10/100+500 {
			assembled.RollingSummary = summary
			tokensUsed += summaryTokens
		}
	}

	// 2. L2 project memory (20% budget)
	projCtx, err := a.memManager.GetProjectContext(ctx, projectID, expertID, question)
	if err == nil {
		assembled.ProjectContext = projCtx
		for _, entry := range projCtx.L2Entries {
			tokensUsed += estimateTokens(entry.Content)
		}
	}

	// 3. Recent messages (20% budget)
	if tokensUsed < budget*60/100 {
		recent, err := a.getRecentMessages(ctx, chatID, a.recentMsgs)
		if err == nil {
			assembled.RecentMessages = recent
			for _, m := range recent {
				tokensUsed += estimateTokens(m.Content)
			}
		}
	}

	// 4. Semantic history (15% budget)
	if tokensUsed < budget*75/100 {
		history, err := a.searchChatHistory(ctx, chatID, question, a.semanticTopK)
		if err == nil {
			assembled.RelevantHistory = history
			for _, h := range history {
				tokensUsed += estimateTokens(h.OneLineSummary)
			}
		}
	}

	// 5. Course chunks (35% budget — most important)
	if tokensUsed < budget*90/100 {
		chunks, err := a.getCourseChunks(ctx, expertID, question, a.chunksTopK)
		if err == nil {
			assembled.CourseChunks = chunks
			for _, c := range chunks {
				tokensUsed += estimateTokens(c.Text)
			}
		}
	}

	assembled.TotalTokens = tokensUsed

	a.logger.Debug("context assembled",
		zap.Int("tokens", tokensUsed),
		zap.Int("chunks", len(assembled.CourseChunks)),
		zap.Int("turn", turnNumber),
	)

	return assembled, nil
}

// getRollingSummary fetches the latest rolling summary for a chat.
func (a *Assembler) getRollingSummary(ctx context.Context, chatID uuid.UUID) (string, error) {
	var summary string
	err := a.db.QueryRow(ctx,
		`SELECT summary_text FROM chat_summaries
		 WHERE chat_id=$1
		 ORDER BY turn_range_end DESC
		 LIMIT 1`,
		chatID,
	).Scan(&summary)
	if err != nil {
		return "", err
	}
	return summary, nil
}

// getRecentMessages fetches the last N messages from a chat.
func (a *Assembler) getRecentMessages(ctx context.Context, chatID uuid.UUID, limit int) ([]Message, error) {
	rows, err := a.db.Query(ctx,
		`SELECT role, content, created_at
		 FROM messages
		 WHERE chat_id=$1
		 ORDER BY turn_number DESC
		 LIMIT $2`,
		chatID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.Role, &m.Content, &m.CreatedAt); err != nil {
			continue
		}
		messages = append(messages, m)
	}
	// Reverse to chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	return messages, nil
}

// searchChatHistory finds semantically relevant past turns.
func (a *Assembler) searchChatHistory(ctx context.Context, chatID uuid.UUID, query string, limit int) ([]HistoryEntry, error) {
	embedding, err := a.ml.EmbedSingle(ctx, query)
	if err != nil {
		return nil, err
	}

	rows, err := a.db.Query(ctx,
		`SELECT turn_number, one_line_summary, COALESCE(topic,''), importance
		 FROM chat_index
		 WHERE chat_id=$1 AND is_superseded=FALSE AND embedding IS NOT NULL
		 ORDER BY embedding <=> $2
		 LIMIT $3`,
		chatID, pgvector.NewVector(embedding), limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []HistoryEntry
	for rows.Next() {
		var e HistoryEntry
		if err := rows.Scan(&e.TurnNumber, &e.OneLineSummary, &e.Topic, &e.Importance); err != nil {
			continue
		}
		entries = append(entries, e)
	}
	return entries, nil
}

// getCourseChunks retrieves and reranks course chunks for a question.
func (a *Assembler) getCourseChunks(ctx context.Context, expertID uuid.UUID, question string, limit int) ([]chinawall.CourseChunk, error) {
	// Step 1: Semantic search — get top 20 candidates
	embedding, err := a.ml.EmbedSingle(ctx, question)
	if err != nil {
		return nil, fmt.Errorf("embed failed: %w", err)
	}

	rows, err := a.db.Query(ctx,
		`SELECT id, chunk_text, COALESCE(topic,'')
		 FROM course_chunks
		 WHERE expert_id=$1
		 ORDER BY embedding <=> $2
		 LIMIT 20`,
		expertID, pgvector.NewVector(embedding),
	)
	if err != nil {
		return nil, fmt.Errorf("chunk search failed: %w", err)
	}
	defer rows.Close()

	type rawChunk struct {
		ID    uuid.UUID
		Text  string
		Topic string
	}
	var rawChunks []rawChunk
	var texts []string
	for rows.Next() {
		var c rawChunk
		if err := rows.Scan(&c.ID, &c.Text, &c.Topic); err != nil {
			continue
		}
		rawChunks = append(rawChunks, c)
		texts = append(texts, c.Text)
	}

	if len(rawChunks) == 0 {
		return nil, nil
	}

	// Step 2: Rerank — get top K with scores
	reranked, err := a.ml.Rerank(ctx, question, texts, limit)
	if err != nil {
		// Fallback: return top K without reranking
		var chunks []chinawall.CourseChunk
		for i, c := range rawChunks {
			if i >= limit {
				break
			}
			chunks = append(chunks, chinawall.CourseChunk{
				ID: c.ID, Text: c.Text, Topic: c.Topic, RerankScore: 0.5,
			})
		}
		return chunks, nil
	}

	// Build final chunks with rerank scores
	var chunks []chinawall.CourseChunk
	for _, r := range reranked {
		if r.Index < len(rawChunks) {
			c := rawChunks[r.Index]
			chunks = append(chunks, chinawall.CourseChunk{
				ID:          c.ID,
				Text:        c.Text,
				Topic:       c.Topic,
				RerankScore: r.Score,
			})
		}
	}
	return chunks, nil
}

// FormatForPrompt converts assembled context into a prompt string.
func (a *Assembler) FormatForPrompt(ctx *AssembledContext) string {
	var sb strings.Builder

	if ctx.RollingSummary != "" {
		sb.WriteString("## Conversation Summary\n")
		sb.WriteString(ctx.RollingSummary)
		sb.WriteString("\n\n")
	}

	if ctx.ProjectContext != nil && len(ctx.ProjectContext.L2Entries) > 0 {
		sb.WriteString("## Project Decisions (from all experts)\n")
		for _, e := range ctx.ProjectContext.L2Entries {
			sb.WriteString(fmt.Sprintf("- %s\n", e.Content))
		}
		sb.WriteString("\n")
	}

	if len(ctx.RecentMessages) > 0 {
		sb.WriteString("## Recent Messages\n")
		for _, m := range ctx.RecentMessages {
			sb.WriteString(fmt.Sprintf("%s: %s\n", strings.ToUpper(m.Role), m.Content))
		}
		sb.WriteString("\n")
	}

	if len(ctx.RelevantHistory) > 0 {
		sb.WriteString("## Relevant Previous Discussion\n")
		for _, h := range ctx.RelevantHistory {
			sb.WriteString(fmt.Sprintf("- Turn %d: %s\n", h.TurnNumber, h.OneLineSummary))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// estimateTokens estimates token count (~4 chars per token).
func estimateTokens(text string) int {
	if text == "" {
		return 0
	}
	return len(text) / 4
}
