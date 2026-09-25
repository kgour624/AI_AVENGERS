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

// Context budget policy (B9, §3.1 P5).
//
// CONTEXT IS A BUDGET. Every turn must stay within MaxTokens. The system
// prompt is NOT part of this budget: it is built separately by the
// enforcer (chinawall.generateFlatText) and concatenated at call time, so
// eviction of chat context can never delete it.
//
// Allocation (WHY these percentages):
// 10% rolling summary — compressed history (prescriptive), keep
// 20% L2 project memory — what other experts decided
// 20% recent messages — last N turns for continuity
// 15% semantic history — relevant past turns
// 35% course chunks — expert knowledge (most important)
//
// Eviction priority when the hard ceiling is hit (lowest value first):
// semantic history → oldest recent → L2 tail → repo/course tail.
// Never evicted: rolling summary (compressed) and the reply thread
// (explicit client action — "if unsure whether an old step is needed
// later, don't evict").
//
// WHY this order: Most important context last = less lost-in-middle.
const (
	budgetSummaryPct   = 10
	budgetL2Pct        = 20
	budgetRecentPct    = 20
	budgetHistoryPct   = 15
	budgetChunksPct    = 35
	// budgetSummarySlack: rolling summary is compressed and high-value;
	// allow it a little room beyond its strict 10% before dropping it.
	budgetSummarySlack = 500
)

// Assembler builds smart context for LLM calls.
// Respects token budget: never exceeds MaxTokens.
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
		if tokensUsed+summaryTokens <= budget*budgetSummaryPct/100+budgetSummarySlack {
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
		l2Budget := budget * budgetL2Pct / 100
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
		recentBudget := budget * budgetRecentPct / 100
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
		historyBudget := budget * budgetHistoryPct / 100
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

	// 5. Course chunks — most important, always include (from fan-out result).
	// B9: capped at the documented 35% slice (was unbounded — the main
	// source of over-budget turns). Highest-rerank chunks win.
	// chunkTokensUsed tracks the shared course+repo chunk slice separately
	// from the cumulative total, so sections 1-4 don't consume chunk room.
	chunksBudget := budget * budgetChunksPct / 100
	chunkTokensUsed := 0
	if chunksRes.err != nil {
		a.logger.Warn("getCourseChunks failed — question will see zero course chunks, likely causing an incorrect Gate 2 refusal",
			zap.String("expert_id", expertID.String()),
			zap.Error(chunksRes.err),
		)
	} else {
		kept, chunkTokens := trimChunksToBudget(chunksRes.chunks, chunksBudget)
		assembled.CourseChunks = kept
		chunkTokensUsed += chunkTokens
		tokensUsed += chunkTokens
		if len(kept) < len(chunksRes.chunks) {
			a.logger.Debug("course chunks truncated to stay within their 35% budget",
				zap.Int("kept", len(kept)),
				zap.Int("total", len(chunksRes.chunks)),
			)
		}
		if len(chunksRes.chunks) == 0 {
			a.logger.Warn("getCourseChunks returned zero chunks (no error) — verify expert_id has ingested content",
				zap.String("expert_id", expertID.String()),
				zap.String("question_preview", question[:minInt(80, len(question))]),
			)
		}
	}

	// 6. Connected repo code chunks (additive, optional) — from fan-out result.
	// B9: shares the same 35% chunk slice; repo chunks are appended after
	// course chunks, so they are the first evicted at the tail.
	if repoChunksRes.err != nil {
		a.logger.Warn("getRepoChunks failed — continuing without connected-repo context",
			zap.String("project_id", projectID.String()),
			zap.Error(repoChunksRes.err),
		)
	} else if len(repoChunksRes.chunks) > 0 {
		if room := chunksBudget - chunkTokensUsed; room > 0 {
			kept, repoTokens := trimChunksToBudget(repoChunksRes.chunks, room)
			assembled.CourseChunks = append(assembled.CourseChunks, kept...)
			chunkTokensUsed += repoTokens
			tokensUsed += repoTokens
		}
	}

	// 7. Reply thread (CT-C2). Runs after fan-out — depends on replyToMessageID.
	// Runs regardless of remaining budget: a reply is an explicit client action
	// and is never evicted (B9 P5: unsure whether needed later → keep).
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

	// 8. HARD CEILING (B9): the per-section caps above sum to ~100%, but
	// the summary slack + reply thread can still push past the budget.
	// Enforce an absolute cap by evicting lowest-value sections first.
	tokensUsed = enforceHardCeiling(assembled, budget, tokensUsed, a.logger)
	assembled.TotalTokens = tokensUsed

	a.logger.Debug("context assembled",
		zap.Int("tokens", tokensUsed),
		zap.Int("budget", budget),
		zap.Int("chunks", len(assembled.CourseChunks)),
		zap.Int("turn", turnNumber),
	)

	return assembled, nil
}

// trimChunksToBudget returns as many leading chunks (best-reranked first)
// as fit within tokenBudget, plus the tokens used. The last chunk is NOT
// split — a half-chunk is worse than a missing one for grounding.
// Pure — unit-tested.
func trimChunksToBudget(chunks []chinawall.CourseChunk, tokenBudget int) ([]chinawall.CourseChunk, int) {
	if tokenBudget <= 0 || len(chunks) == 0 {
		return nil, 0
	}
	used := 0
	var kept []chinawall.CourseChunk
	for _, c := range chunks {
		t := estimateTokens(c.Text)
		if used+t > tokenBudget && len(kept) > 0 {
			break
		}
		kept = append(kept, c)
		used += t
	}
	return kept, used
}

// enforceHardCeiling guarantees TotalTokens <= budget by evicting
// low-value sections in a fixed priority order (B9 P5). Never evicts the
// rolling summary (compressed, high value) or the reply thread (explicit
// client action). Returns the (reduced) token count. Deterministic.
func enforceHardCeiling(assembled *AssembledContext, budget, tokensUsed int, logger *zap.Logger) int {
	if assembled == nil || tokensUsed <= budget {
		return tokensUsed
	}
	over := func() int { return tokensUsed - budget }

	// Priority 1: semantic history (luxury — most queries don't need it).
	for len(assembled.RelevantHistory) > 0 && tokensUsed > budget {
		last := len(assembled.RelevantHistory) - 1
		tokensUsed -= estimateTokens(assembled.RelevantHistory[last].OneLineSummary)
		assembled.RelevantHistory = assembled.RelevantHistory[:last]
	}
	// Priority 2: oldest recent messages (keep newest for continuity).
	for len(assembled.RecentMessages) > 1 && tokensUsed > budget {
		tokensUsed -= estimateTokens(assembled.RecentMessages[0].Content)
		assembled.RecentMessages = assembled.RecentMessages[1:]
	}
	// Priority 3: L2 project memory tail.
	if assembled.ProjectContext != nil {
		for len(assembled.ProjectContext.L2Entries) > 0 && tokensUsed > budget {
			last := len(assembled.ProjectContext.L2Entries) - 1
			tokensUsed -= estimateTokens(assembled.ProjectContext.L2Entries[last].Content)
			assembled.ProjectContext.L2Entries = assembled.ProjectContext.L2Entries[:last]
		}
	}
	// Priority 4: chunk tail (repo appended last → evicted first).
	for len(assembled.CourseChunks) > 0 && tokensUsed > budget {
		last := len(assembled.CourseChunks) - 1
		tokensUsed -= estimateTokens(assembled.CourseChunks[last].Text)
		assembled.CourseChunks = assembled.CourseChunks[:last]
	}
	// Priority 5 (last resort): rolling summary. Only if still over after
	// every evictable section is gone — means the reply thread alone is
	// over budget; drop the summary rather than violate the hard cap.
	if tokensUsed > budget && assembled.RollingSummary != "" {
		tokensUsed -= estimateTokens(assembled.RollingSummary)
		assembled.RollingSummary = ""
	}

	if logger != nil && tokensUsed > budget {
		// Reply thread alone exceeds the budget — explicit client action
		// wins; log so operators can see a genuinely oversized reply.
		logger.Warn("context still over budget after eviction (reply thread pinned)",
			zap.Int("tokens", tokensUsed),
			zap.Int("budget", budget),
			zap.Int("over", over()),
		)
	}
	return tokensUsed
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

// GetCourseChunksForWorkflow retrieves training chunks for a workflow expert.
// Used by AgentLoop to give experts their training knowledge as context.
//
// WHY exported (not getCourseChunks which is unexported):
//   getCourseChunks is chat-flow only (called inside Assemble()).
//   Workflow needs the same retrieval but without the full Assemble() overhead
//   (no rolling summary, no L2 memory, no recent messages — just chunks).
//
// WHY topK=5 default:
//   Workflow context already has blackboard artifacts.
//   5 chunks * ~500 tokens = 2500 tokens — enough for principles without
//   crowding out blackboard context.
//
// Failure policy: non-fatal.
//   Empty slice returned on any error — AgentLoop continues with
//   blackboard-only context. Expert never refuses due to missing training.
func (a *Assembler) GetCourseChunksForWorkflow(
	ctx context.Context,
	expertID uuid.UUID,
	taskDescription string,
	topK int,
) ([]chinawall.CourseChunk, error) {
	if topK <= 0 {
		topK = 5
	}
	return a.getCourseChunks(ctx, expertID, taskDescription, topK)
}

// GetProjectMemoryText loads project L1/L2 memory for a workflow expert and
// returns a prompt-ready block (or "" when empty / unavailable).
//
// Used by workflow AgentLoop so design-phase experts see the same project
// decisions standalone chat already injects. Nil-safe: missing memManager,
// missing project, or fetch failure → empty string (non-fatal).
//
// Budget: L2 entries are capped by GetProjectContext (top 8); the rendered
// block is hard-capped at ~3000 chars so a chatty project cannot crowd out
// training chunks / blackboard in the workflow prompt.
func (a *Assembler) GetProjectMemoryText(
	ctx context.Context,
	projectID uuid.UUID,
	expertID uuid.UUID,
	query string,
) string {
	if a == nil || a.memManager == nil || projectID == uuid.Nil {
		return ""
	}
	pc, err := a.memManager.GetProjectContext(ctx, projectID, expertID, query)
	if err != nil || pc == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("[PROJECT MEMORY — established decisions for this project]\n")
	sb.WriteString("Treat these as settled context. Do not contradict them silently.\n\n")

	wrote := false
	if pc.L1Memory != nil {
		if pc.L1Memory.ContextSummary != "" {
			sb.WriteString("Your recent context summary:\n")
			sb.WriteString(pc.L1Memory.ContextSummary)
			sb.WriteString("\n\n")
			wrote = true
		}
		if len(pc.L1Memory.KeyDecisions) > 0 {
			sb.WriteString("Your recent key decisions:\n")
			for _, d := range pc.L1Memory.KeyDecisions {
				if d.Content == "" {
					continue
				}
				sb.WriteString(fmt.Sprintf("- %s\n", d.Content))
				wrote = true
			}
			sb.WriteString("\n")
		}
	}
	if len(pc.L2Entries) > 0 {
		// B7: surface consolidated summary first (highest-SNR block), then
		// remaining non-summary rows. Decayed/superseded already filtered
		// by L2 search.
		var summaryBlock strings.Builder
		var otherBlock strings.Builder
		for _, e := range pc.L2Entries {
			if e.Content == "" {
				continue
			}
			if e.MemoryType == "summary" {
				summaryBlock.WriteString(e.Content)
				summaryBlock.WriteString("\n\n")
			} else {
				otherBlock.WriteString(fmt.Sprintf("- [%s] %s\n", e.MemoryType, e.Content))
			}
		}
		if summaryBlock.Len() > 0 {
			sb.WriteString("Project consolidated summary:\n")
			sb.WriteString(summaryBlock.String())
			wrote = true
		}
		if otherBlock.Len() > 0 {
			sb.WriteString("Cross-expert project decisions:\n")
			sb.WriteString(otherBlock.String())
			sb.WriteString("\n")
			wrote = true
		}
	}
	if !wrote {
		return ""
	}

	out := sb.String()
	const maxChars = 3000
	if len(out) > maxChars {
		return out[:maxChars] + "\n…[project memory truncated]\n"
	}
	return out
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
	// Step 1: Vector search — one RANKED list.
	embedding, err := a.embedder.EmbedSingle(ctx, question)
	if err != nil {
		return nil, fmt.Errorf("embed failed: %w", err)
	}

	// Feature #23: source_file and chunk_index are selected so the citation modal can
	// name the transcript and the position.
	vectorRows, err := a.db.Query(ctx,
		`SELECT id, chunk_text, COALESCE(topic,''), COALESCE(source_file,''), chunk_index
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
		ID         uuid.UUID
		Text       string
		Topic      string
		SourceFile string // Feature #23: transcript filename
		ChunkIndex int    // Feature #23: position in transcript
	}

	byID := make(map[uuid.UUID]rawChunk)
	var vectorIDs []uuid.UUID

	for vectorRows.Next() {
		var c rawChunk
		if err := vectorRows.Scan(&c.ID, &c.Text, &c.Topic, &c.SourceFile, &c.ChunkIndex); err != nil {
			continue
		}
		if _, dup := byID[c.ID]; dup {
			continue
		}
		byID[c.ID] = c
		vectorIDs = append(vectorIDs, c.ID)
	}

	// Step 2: Keyword search — the second RANKED list.
	//
	// WHY the terms are extracted instead of passing the question: plainto_tsquery
	// ANDs every lexeme, so a natural-language question required every one of its
	// words to appear in the same chunk and usually matched nothing. The terms are
	// whitelisted to [a-z0-9_] before they reach the tsquery, which is what makes
	// joining them with ' | ' safe.
	//
	// ORDER BY ts_rank is the other half of the fix: without it, LIMIT 10 returned ten
	// arbitrary matching rows rather than the ten best.
	var keywordIDs []uuid.UUID
	terms := salientTerms(question)
	if len(terms) > 0 {
		const keywordSQL = `
			SELECT id, chunk_text, COALESCE(topic,''), COALESCE(source_file,''), chunk_index
			  FROM course_chunks
			 WHERE expert_id = $1
			   AND chunk_text_tsv @@ to_tsquery('english', array_to_string($2::text[], ' | '))
			 ORDER BY ts_rank(chunk_text_tsv, to_tsquery('english', array_to_string($2::text[], ' | '))) DESC
			 LIMIT 10`
		keywordRows, keywordErr := a.db.Query(ctx, keywordSQL, expertID, terms)
		if keywordErr == nil {
			defer keywordRows.Close()
			for keywordRows.Next() {
				var c rawChunk
				if err := keywordRows.Scan(&c.ID, &c.Text, &c.Topic, &c.SourceFile, &c.ChunkIndex); err != nil {
					continue
				}
				if _, dup := byID[c.ID]; !dup {
					byID[c.ID] = c
				}
				keywordIDs = append(keywordIDs, c.ID)
			}
		}
		// Keyword search failure is non-fatal — the vector list still stands.
	}

	if len(byID) == 0 {
		return nil, nil
	}

	// Step 3: Fuse the two ranked lists.
	//
	// WHY fusion rather than concatenation (what this used to do): appending keyword
	// results after vector results meant the keyword hits sat at the end of the
	// candidate list, so on the reranker-unavailable path — which returns the first
	// `limit` candidates and scores them all 0.5 — they were discarded entirely. With
	// RRF the fused order is the real order, and that path inherits it.
	order := reciprocalRankFusion([][]uuid.UUID{vectorIDs, keywordIDs})
	candidates := make([]rawChunk, 0, len(order))
	for _, id := range order {
		if c, ok := byID[id]; ok {
			candidates = append(candidates, c)
		}
	}

	// Step 3: Rerank merged candidates
	texts := make([]string, len(candidates))
	for i, c := range candidates {
		texts[i] = c.Text
	}

	reranked, err := a.sidecar.Rerank(ctx, question, texts, limit)
	if err != nil {
		// Fallback: return top K without reranking.
		//
		// This was silent before. It must not be: the flat 0.5 score below
		// is under workflow Gate 1's 0.70 threshold, so while the sidecar is
		// unreachable every expert silently drops to "generic allowed" mode
		// even with perfect training data — and no error surfaced anywhere.
		a.logger.Warn("reranker unavailable — falling back to flat 0.5 scores",
			zap.Int("candidates", len(candidates)),
			zap.Error(err),
		)
		var chunks []chinawall.CourseChunk
		for i, c := range candidates {
			if i >= limit {
				break
			}
			// Feature #23: Include SourceFile and ChunkIndex in fallback path
			chunks = append(chunks, chinawall.CourseChunk{
				ID:          c.ID,
				Text:        c.Text,
				Topic:       c.Topic,
				RerankScore: 0.5,
				SourceFile:  c.SourceFile,
				ChunkIndex:  c.ChunkIndex,
			})
		}
		return chunks, nil
	}

	var chunks []chinawall.CourseChunk
	for _, r := range reranked {
		if r.Index < len(candidates) {
			c := candidates[r.Index]
			// Feature #23: Include SourceFile and ChunkIndex in reranked results
			chunks = append(chunks, chinawall.CourseChunk{
				ID:          c.ID,
				Text:        c.Text,
				Topic:       c.Topic,
				RerankScore: r.Score,
				SourceFile:  c.SourceFile,
				ChunkIndex:  c.ChunkIndex,
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
