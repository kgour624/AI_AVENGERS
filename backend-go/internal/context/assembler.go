package context

import (
	"context"
	"encoding/json"
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

// jsonUnmarshalInt unmarshals a system_settings JSONB value (stored as
// e.g. the bare text "10") into an int. Kept as a tiny named helper
// (not inlined) so getReplyThreadMaxDepth's error handling reads as one
// line, matching this file's existing style of small single-purpose
// helpers (estimateTokens, etc.) at the bottom of the file.
func jsonUnmarshalInt(data []byte, v *int) error {
	return json.Unmarshal(data, v)
}

// AssembledContext holds everything needed for one LLM call.
// Built fresh for every turn. Never cached.
type AssembledContext struct {
	RollingSummary  string
	RecentMessages  []Message
	RelevantHistory []HistoryEntry
	CourseChunks    []chinawall.CourseChunk
	ProjectContext  *memory.ProjectContext
	// ReplyThread (CT-C2): populated ONLY when the incoming message has
	// a non-nil ReplyToMessageID (CATEGORY_TEMPLATE_HANDOFF.md §5). nil
	// for every fresh (non-reply) question — the vast majority of
	// messages, both before and after this feature existed. This is a
	// new, isolated budgeted slice; it does not replace or alter
	// RecentMessages/RelevantHistory/CourseChunks logic above (additive
	// only, per CATEGORY_TEMPLATE_HANDOFF.md §5).
	ReplyThread []ReplyThreadEntry
	TotalTokens int
}

// ReplyThreadEntry is one message in a resolved reply thread, pinned
// (single entry) or full-chain (multiple entries, root-ward order).
type ReplyThreadEntry struct {
	Role       string
	Content    string
	TurnNumber int
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
	embedder    ml.Embedder       // ml.Embedder interface: sidecar or CodeCraftAPI, resolved at call time
	sidecar     *ml.SidecarClient // kept separately for Rerank() — Rerank is sidecar-only, not in Embedder interface
	memManager  *memory.Manager
	maxTokens   int
	recentMsgs  int
	semanticTopK int
	chunksTopK  int
	logger      *zap.Logger
}

// NewAssembler creates a new context assembler.
// NewAssembler creates a new context assembler.
// embedder satisfies ml.Embedder — either *ml.SidecarClient (default) or
// *ml.DynamicEmbedder (when CodeCraftAPI embeddings are enabled).
//
// sidecar is kept as a separate *ml.SidecarClient because getCourseChunks
// and getRepoChunks call Rerank(), which is sidecar-only and NOT part of
// the Embedder interface (locked decision: reranking always stays on the
// Python sidecar).
func NewAssembler(
	db *pgxpool.Pool,
	embedder ml.Embedder,
	sidecar *ml.SidecarClient,
	memManager *memory.Manager,
	maxTokens, recentMsgs, semanticTopK, chunksTopK int,
	logger *zap.Logger,
) *Assembler {
	return &Assembler{
		db:           db,
		embedder:     embedder,
		sidecar:      sidecar,
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
// Assemble builds the full context for a turn.
//
// DESIGN PATTERN: Fan-Out / Parallel Fetch
//   Steps 1-6 are independent DB/ML queries. Running them sequentially
//   wastes wall-clock time proportional to N*avg_query_latency.
//   Fan-out runs all independent steps concurrently, then merges results
//   and enforces per-step token budgets.
//
// SOLID:
//   SRP: each private helper fetches exactly one context source.
//   OCP: new sources added as new goroutine + channel, no budget logic change.
//
// Mental execution:
// Turn 25, Expert: DB Expert, Question: "should I shard?"
// Fan-out (all concurrent):
//   g1: getRollingSummary  → 500 tokens
//   g2: getL2Memory        → 300 tokens
//   g3: getRecentMessages  → 400 tokens
//   g4: searchChatHistory  → 200 tokens
//   g5: getCourseChunks    → 2000 tokens
//   g6: getRepoChunks      → 0 tokens (no repo connected)
// Merge + budget enforce → ~3400 tokens total
func (a *Assembler) Assemble(
	ctx context.Context,
	chatID uuid.UUID,
	projectID uuid.UUID,
	expertID uuid.UUID,
	question string,
	turnNumber int,
	// replyToMessageID/includeFullThread (CT-C2): nil/false for every
	// fresh (non-reply) question — every existing behavior above this
	// point in Assemble is completely unaffected. Only orchestrator.go's
	// processWithExpert passes non-nil values through, and only when the
	// client actually replied to a specific prior message.
	replyToMessageID *uuid.UUID,
	includeFullThread bool,
) (*AssembledContext, error) {

	budget := a.maxTokens

	// ── Fan-out: all independent context sources run concurrently ──────
	// Each goroutine writes into its own typed result channel (buffered=1).
	// No shared mutable state between goroutines — no mutex needed.
	type summaryResult struct {
		text string
		err  error
	}
	type l2Result struct {
		projCtx *memory.ProjectContext
		err     error
	}
	type recentResult struct {
		msgs []Message
		err  error
	}
	type historyResult struct {
		entries []HistoryEntry
		err     error
	}
	type chunksResult struct {
		chunks []chinawall.CourseChunk
		err    error
	}

	summaryCh    := make(chan summaryResult, 1)
	l2Ch         := make(chan l2Result, 1)
	recentCh     := make(chan recentResult, 1)
	historyCh    := make(chan historyResult, 1)
	chunksCh     := make(chan chunksResult, 1)
	repoChunksCh := make(chan chunksResult, 1)

	go func() {
		s, err := a.getRollingSummary(ctx, chatID)
		summaryCh <- summaryResult{s, err}
	}()
	go func() {
		pc, err := a.memManager.GetProjectContext(ctx, projectID, expertID, question)
		l2Ch <- l2Result{pc, err}
	}()
	go func() {
		msgs, err := a.getRecentMessages(ctx, chatID, a.recentMsgs)
		recentCh <- recentResult{msgs, err}
	}()
	go func() {
		h, err := a.searchChatHistory(ctx, chatID, question, a.semanticTopK)
		historyCh <- historyResult{h, err}
	}()
	go func() {
		c, err := a.getCourseChunks(ctx, expertID, question, a.chunksTopK)
		chunksCh <- chunksResult{c, err}
	}()
	go func() {
		c, err := a.getRepoChunks(ctx, projectID, question, a.chunksTopK)
		repoChunksCh <- chunksResult{c, err}
	}()

	// Collect all fan-out results
	summaryRes    := <-summaryCh
	l2Res         := <-l2Ch
	recentRes     := <-recentCh
	historyRes    := <-historyCh
	chunksRes     := <-chunksCh
	repoChunksRes := <-repoChunksCh

	// ── Merge + budget enforcement ──────────────────────────────────────
	assembled := &AssembledContext{}
	tokensUsed := 0

	// 1. Rolling summary (10% budget)
	if summaryRes.err == nil && summaryRes.text != "" {
		summaryTokens := estimateTokens(summaryRes.text)
		if tokensUsed+summaryTokens <= budget*10/100+500 {
			assembled.RollingSummary = summaryRes.text
			tokensUsed += summaryTokens
		}
	}

	// 2. L2 project memory (20% budget)
	// FIX (2026-09-08 RCA): this step previously had NO budget
	// enforcement at all (unlike steps 3/4/5, which at least check
	// tokensUsed before running). Every L2Entry's tokens were added
	// unconditionally, so a long-running conversation with many
	// cross-expert decisions could silently consume far more than its
	// 2. L2 project memory (20% budget) — from fan-out result
	if l2Res.err == nil && l2Res.projCtx != nil {
		l2Budget := budget * 20 / 100
		l2Tokens := 0
		kept := l2Res.projCtx.L2Entries[:0:0]
		for _, entry := range l2Res.projCtx.L2Entries {
			t := estimateTokens(entry.Content)
			if l2Tokens+t > l2Budget && len(kept) > 0 {
				break
			}
			kept = append(kept, entry)
			l2Tokens += t
		}
		if len(kept) < len(l2Res.projCtx.L2Entries) {
			a.logger.Debug("L2 project memory truncated to stay within its 20% budget",
				zap.Int("kept", len(kept)),
				zap.Int("total", len(l2Res.projCtx.L2Entries)),
			)
		}
		l2Res.projCtx.L2Entries = kept
		assembled.ProjectContext = l2Res.projCtx
		tokensUsed += l2Tokens
	}

	// 3. Recent messages (20% budget) — from fan-out result
	if recentRes.err == nil {
		recentBudget := budget * 20 / 100
		recentTokens := 0
		var kept []Message
		for i := len(recentRes.msgs) - 1; i >= 0; i-- {
			t := estimateTokens(recentRes.msgs[i].Content)
			if recentTokens+t > recentBudget && len(kept) > 0 {
				break
			}
			kept = append(kept, recentRes.msgs[i])
			recentTokens += t
		}
		for i, j := 0, len(kept)-1; i < j; i, j = i+1, j-1 {
			kept[i], kept[j] = kept[j], kept[i]
		}
		if len(kept) < len(recentRes.msgs) {
			a.logger.Debug("recent messages truncated to stay within 20% budget",
				zap.Int("kept", len(kept)),
				zap.Int("total", len(recentRes.msgs)),
			)
		}
		assembled.RecentMessages = kept
		tokensUsed += recentTokens
	}

	// 4. Semantic history (15% budget) — from fan-out result
	if historyRes.err == nil {
		historyBudget := budget * 15 / 100
		historyTokens := 0
		var kept []HistoryEntry
		for _, h := range historyRes.entries {
			t := estimateTokens(h.OneLineSummary)
			if historyTokens+t > historyBudget && len(kept) > 0 {
				break
			}
			kept = append(kept, h)
			historyTokens += t
		}
		assembled.RelevantHistory = kept
		tokensUsed += historyTokens
	}

	// 5. Course chunks — most important, always include (from fan-out result)
	if chunksRes.err != nil {
		a.logger.Warn("getCourseChunks failed — question will see zero course chunks, likely causing an incorrect Gate 2 refusal",
			zap.String("expert_id", expertID.String()),
			zap.Error(chunksRes.err),
		)
	} else {
		assembled.CourseChunks = chunksRes.chunks
		for _, c := range chunksRes.chunks {
			tokensUsed += estimateTokens(c.Text)
		}
		if len(chunksRes.chunks) == 0 {
			a.logger.Warn("getCourseChunks returned zero chunks (no error) — verify expert_id has ingested content",
				zap.String("expert_id", expertID.String()),
				zap.String("question_preview", question[:minInt(80, len(question))]),
			)
		}
	}

	// 6. Connected repo code chunks (additive, optional) — from fan-out result
	if repoChunksRes.err != nil {
		a.logger.Warn("getRepoChunks failed — continuing without connected-repo context",
			zap.String("project_id", projectID.String()),
			zap.Error(repoChunksRes.err),
		)
	} else if len(repoChunksRes.chunks) > 0 && tokensUsed < budget*95/100 {
		assembled.CourseChunks = append(assembled.CourseChunks, repoChunksRes.chunks...)
		for _, c := range repoChunksRes.chunks {
			tokensUsed += estimateTokens(c.Text)
		}
	}

	// 7. Reply thread (CT-C2). Runs after fan-out — depends on replyToMessageID.
	// Runs regardless of remaining budget: a reply is an explicit client action.
	if replyToMessageID != nil {
		thread, err := a.getReplyThread(ctx, *replyToMessageID, includeFullThread)
		if err != nil {
			a.logger.Warn("reply thread resolution failed, continuing without it",
				zap.String("reply_to", replyToMessageID.String()),
				zap.Error(err),
			)
		} else {
			assembled.ReplyThread = thread
			for _, e := range thread {
				tokensUsed += estimateTokens(e.Content)
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

// getReplyThread resolves reply context for a message that has a
// non-nil ReplyToMessageID (CT-C2, CATEGORY_TEMPLATE_HANDOFF.md §5).
//
// includeFullThread=false (CT-L6 default): returns exactly ONE entry —
// the pinned message itself. Cheapest, matches "sirf pin kra sirf uska".
//
// includeFullThread=true (explicit opt-in only): walks reply_to_message_id
// pointers upward (root-ward), capped at system_settings.reply_thread_max_depth
// (falls back to 10 if the setting is missing or unparseable — same
// default the migration seeds, so a missing row behaves identically to
// the seeded row).
//
// Mental execution (happy path):
// pinned message id=X, X.reply_to_message_id=Y, Y.reply_to_message_id=NULL
// includeFullThread=true, max_depth=10
// -> loadOne(X) -> entries=[X], next=Y
// -> loadOne(Y) -> entries=[X,Y], next=nil (Y has no parent)
// -> loop ends (next==nil), return [X,Y] (2 entries, well under cap)
//
// Edge case (max depth exceeded): a 15-message-deep reply chain with
// max_depth=10 -> loop stops after 10 iterations even though more
// ancestors exist — partial thread is returned, not an error, and is
// logged so an admin can see truncation happened.
//
// Edge case (pinned message deleted): messages.reply_to_message_id has
// ON DELETE SET NULL (migration 010), so a deleted parent simply breaks
// the chain at that point — loadOne on a since-deleted id returns
// pgx.ErrNoRows, treated as "no more ancestors", not a hard failure.
func (a *Assembler) getReplyThread(ctx context.Context, pinnedID uuid.UUID, includeFullThread bool) ([]ReplyThreadEntry, error) {
	entry, nextParent, err := a.loadOneThreadMessage(ctx, pinnedID)
	if err != nil {
		return nil, fmt.Errorf("load pinned message: %w", err)
	}
	entries := []ReplyThreadEntry{entry}

	if !includeFullThread {
		return entries, nil
	}

	maxDepth := a.getReplyThreadMaxDepth(ctx)
	for depth := 1; depth < maxDepth && nextParent != nil; depth++ {
		next, parent, err := a.loadOneThreadMessage(ctx, *nextParent)
		if err != nil {
			// Deleted/missing ancestor — chain ends here, not an error
			// for the caller (partial thread is still useful context).
			break
		}
		entries = append(entries, next)
		nextParent = parent
	}
	if nextParent != nil {
		a.logger.Warn("reply thread truncated at max depth",
			zap.Int("max_depth", maxDepth),
			zap.String("pinned_id", pinnedID.String()),
		)
	}

	return entries, nil
}

// loadOneThreadMessage loads role/content/turn_number for one message,
// plus its own parent pointer (for continuing the walk upward).
//
// FIX (2026-09-08 RCA, round 9): previously selected ONLY `content`,
// which is the empty string for every categorized (structured-JSON)
// expert's answer - the structured generation path only ever
// populates chinawall.EnforceResult.TemplateSections, never Answer/
// Content (see round 7's HANDOFF entry for the same underlying fact,
// applied there to the messages table's own persistence gap). Replying
// to a structured answer therefore pinned a genuinely EMPTY string as
// context - the new expert's prompt got "## Replying To\nASSISTANT
// (turn N): " with nothing after the colon, so it had no real basis to
// continue the conversation and understandably asked "what are we
// talking about?" or refused. Fix: also select template_sections, and
// when content is empty but sections exist, flatten them into a single
// readable text block so the reply thread always carries the real
// answer, structured or flat.
func (a *Assembler) loadOneThreadMessage(ctx context.Context, id uuid.UUID) (ReplyThreadEntry, *uuid.UUID, error) {
	var e ReplyThreadEntry
	var parent *uuid.UUID
	var templateSectionsJSON []byte
	err := a.db.QueryRow(ctx,
		`SELECT role, content, turn_number, reply_to_message_id, template_sections
		 FROM messages WHERE id=$1`,
		id,
	).Scan(&e.Role, &e.Content, &e.TurnNumber, &parent, &templateSectionsJSON)
	if err != nil {
		return ReplyThreadEntry{}, nil, err
	}
	if strings.TrimSpace(e.Content) == "" && len(templateSectionsJSON) > 0 {
		if flattened := flattenTemplateSections(templateSectionsJSON); flattened != "" {
			e.Content = flattened
		}
	}
	return e, parent, nil
}

// flattenTemplateSections converts a messages.template_sections JSONB
// value (array mirroring chinawall.TemplateSectionResult) into a single
// human-readable text block, so reply-thread context (and any other
// consumer needing plain text) has a real, complete answer to work
// with instead of an empty Content string. Returns "" on any parse
// failure or empty input - callers already handle "" as "nothing to
// add", matching every other non-fatal-degradation pattern in this file.
func flattenTemplateSections(raw []byte) string {
	var sections []struct {
		Label   string `json:"label"`
		Content string `json:"content"`
	}
	if err := json.Unmarshal(raw, &sections); err != nil {
		return ""
	}
	var sb strings.Builder
	for _, s := range sections {
		if strings.TrimSpace(s.Content) == "" {
			continue
		}
		if sb.Len() > 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString(s.Label)
		sb.WriteString(":\n")
		sb.WriteString(s.Content)
	}
	return sb.String()
}

// getReplyThreadMaxDepth reads system_settings.reply_thread_max_depth.
// Falls back to 10 (same default migration 010 seeds) if the row is
// missing or its value is not a valid integer — never blocks reply
// resolution on a settings-read failure.
func (a *Assembler) getReplyThreadMaxDepth(ctx context.Context) int {
	const fallback = 10
	var valueJSON []byte
	err := a.db.QueryRow(ctx,
		`SELECT value FROM system_settings WHERE key='reply_thread_max_depth'`,
	).Scan(&valueJSON)
	if err != nil {
		return fallback
	}
	var depth int
	if err := jsonUnmarshalInt(valueJSON, &depth); err != nil || depth <= 0 {
		return fallback
	}
	return depth
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
	embedding, err := a.embedder.EmbedSingle(ctx, query)
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

// getCourseChunks retrieves and reranks course chunks using HYBRID search.
// Hybrid = vector similarity + keyword match combined.
//
// WHY hybrid search (Byte by Byte AI course):
// Course taught two indexing methods:
// 1. Vector-based: captures semantic meaning ("database partitioning" matches "sharding")
// 2. Keyword-based: captures exact terms ("PostgreSQL" exact match)
// Hybrid combines both: better recall than either alone.
//
// Algorithm:
// Step 1: Vector search -> top 15 candidates (semantic)
// Step 2: Keyword search -> top 10 candidates (exact)
// Step 3: Merge + deduplicate
// Step 4: Rerank merged set -> top K
func (a *Assembler) getCourseChunks(ctx context.Context, expertID uuid.UUID, question string, limit int) ([]chinawall.CourseChunk, error) {
	// Step 1: Vector search
	embedding, err := a.embedder.EmbedSingle(ctx, question)
	if err != nil {
		return nil, fmt.Errorf("embed failed: %w", err)
	}

	vectorRows, err := a.db.Query(ctx,
		`SELECT id, chunk_text, COALESCE(topic,'')
		 FROM course_chunks
		 WHERE expert_id=$1
		 ORDER BY embedding <=> $2
		 LIMIT 15`,
		expertID, pgvector.NewVector(embedding),
	)
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}
	defer vectorRows.Close()

	type rawChunk struct {
		ID    uuid.UUID
		Text  string
		Topic string
	}

	seen := make(map[uuid.UUID]bool)
	var candidates []rawChunk

	for vectorRows.Next() {
		var c rawChunk
		if err := vectorRows.Scan(&c.ID, &c.Text, &c.Topic); err != nil {
			continue
		}
		if !seen[c.ID] {
			seen[c.ID] = true
			candidates = append(candidates, c)
		}
	}

	// Step 2: Keyword search (full-text)
	// WHY: Exact terms like "PostgreSQL", "Kafka", "Redis" may not be
	// captured well by semantic search alone.
	keywordRows, err := a.db.Query(ctx,
		`SELECT id, chunk_text, COALESCE(topic,'')
		 FROM course_chunks
		 WHERE expert_id=$1
		   AND chunk_text_tsv @@ plainto_tsquery('english', $2)
		 LIMIT 10`,
		expertID, question,
	)
	if err == nil {
		defer keywordRows.Close()
		for keywordRows.Next() {
			var c rawChunk
			if err := keywordRows.Scan(&c.ID, &c.Text, &c.Topic); err != nil {
				continue
			}
			if !seen[c.ID] {
				seen[c.ID] = true
				candidates = append(candidates, c)
			}
		}
	}
	// Keyword search failure is non-fatal — vector results still usable

	if len(candidates) == 0 {
		return nil, nil
	}

	// Step 3: Rerank merged candidates
	texts := make([]string, len(candidates))
	for i, c := range candidates {
		texts[i] = c.Text
	}

	reranked, err := a.sidecar.Rerank(ctx, question, texts, limit)
	if err != nil {
		// Fallback: return top K without reranking
		var chunks []chinawall.CourseChunk
		for i, c := range candidates {
			if i >= limit {
				break
			}
			chunks = append(chunks, chinawall.CourseChunk{
				ID: c.ID, Text: c.Text, Topic: c.Topic, RerankScore: 0.5,
			})
		}
		return chunks, nil
	}

	var chunks []chinawall.CourseChunk
	for _, r := range reranked {
		if r.Index < len(candidates) {
			c := candidates[r.Index]
			chunks = append(chunks, chinawall.CourseChunk{
				ID: c.ID, Text: c.Text, Topic: c.Topic, RerankScore: r.Score,
			})
		}
	}
	return chunks, nil
}

// getRepoChunks retrieves and reranks connected-repo code chunks,
// using the exact same hybrid (vector + keyword) search pattern as
// getCourseChunks above, scoped by project_id instead of expert_id.
//
// Bug 3.3 fix (docs bug list): this method did not exist at all -
// repo_chunks was written to by the sync pipeline but never read by
// anything. Returns chinawall.CourseChunk (not a new type) so results
// flow through the same `chunks` parameter every expert call already
// consumes - see the call site in Assemble() above for why that
// matters.
func (a *Assembler) getRepoChunks(ctx context.Context, projectID uuid.UUID, question string, limit int) ([]chinawall.CourseChunk, error) {
	embedding, err := a.embedder.EmbedSingle(ctx, question)
	if err != nil {
		return nil, fmt.Errorf("embed failed: %w", err)
	}

	vectorRows, err := a.db.Query(ctx,
		`SELECT id, chunk_text, file_path
		 FROM repo_chunks
		 WHERE project_id=$1
		 ORDER BY embedding <=> $2
		 LIMIT 15`,
		projectID, pgvector.NewVector(embedding),
	)
	if err != nil {
		return nil, fmt.Errorf("repo vector search failed: %w", err)
	}
	defer vectorRows.Close()

	type rawRepoChunk struct {
		ID       uuid.UUID
		Text     string
		FilePath string
	}

	seen := make(map[uuid.UUID]bool)
	var candidates []rawRepoChunk

	for vectorRows.Next() {
		var c rawRepoChunk
		if err := vectorRows.Scan(&c.ID, &c.Text, &c.FilePath); err != nil {
			continue
		}
		if !seen[c.ID] {
			seen[c.ID] = true
			candidates = append(candidates, c)
		}
	}

	// Keyword search (full-text) - migration 003 already adds
	// chunk_text_tsv as a generated column + GIN index for exactly this.
	keywordRows, err := a.db.Query(ctx,
		`SELECT id, chunk_text, file_path
		 FROM repo_chunks
		 WHERE project_id=$1
		   AND chunk_text_tsv @@ plainto_tsquery('english', $2)
		 LIMIT 10`,
		projectID, question,
	)
	if err == nil {
		defer keywordRows.Close()
		for keywordRows.Next() {
			var c rawRepoChunk
			if err := keywordRows.Scan(&c.ID, &c.Text, &c.FilePath); err != nil {
				continue
			}
			if !seen[c.ID] {
				seen[c.ID] = true
				candidates = append(candidates, c)
			}
		}
	}
	// Keyword search failure is non-fatal - vector results still usable

	if len(candidates) == 0 {
		return nil, nil
	}

	texts := make([]string, len(candidates))
	for i, c := range candidates {
		texts[i] = c.Text
	}

	reranked, err := a.sidecar.Rerank(ctx, question, texts, limit)
	if err != nil {
		var chunks []chinawall.CourseChunk
		for i, c := range candidates {
			if i >= limit {
				break
			}
			// WHY prefix "repo:": lets a citation/log consumer tell a
			// course chunk from a code chunk at a glance without a
			// separate type or field - CourseChunk has no "source" field
			// and adding one would ripple into chinawall/decision/every
			// caller that constructs or reads a CourseChunk today.
			chunks = append(chunks, chinawall.CourseChunk{
				ID: c.ID, Text: c.Text, Topic: "repo:" + c.FilePath, RerankScore: 0.5,
			})
		}
		return chunks, nil
	}

	var chunks []chinawall.CourseChunk
	for _, r := range reranked {
		if r.Index < len(candidates) {
			c := candidates[r.Index]
			chunks = append(chunks, chinawall.CourseChunk{
				ID: c.ID, Text: c.Text, Topic: "repo:" + c.FilePath, RerankScore: r.Score,
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

	// CT-C2: reply thread, if this question is a reply. Placed before
	// RecentMessages/RelevantHistory so the specific pinned/threaded
	// context the client explicitly asked to reference is not lost in
	// the middle of the prompt.
	if len(ctx.ReplyThread) > 0 {
		sb.WriteString("## Replying To\n")
		for _, e := range ctx.ReplyThread {
			sb.WriteString(fmt.Sprintf("%s (turn %d): %s\n", strings.ToUpper(e.Role), e.TurnNumber, e.Content))
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

// minInt is a tiny local helper — avoids importing another package
// just for a two-value min, matching the pattern used in decision/
// engine.go and chinawall/enforcer.go's own package-local minInt.
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
