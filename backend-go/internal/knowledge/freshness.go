// C6: knowledge freshness / staleness detection.
//
// WHY: an expert's corpus can silently go out of date, and an embedding
// provider/model change leaves old vectors incompatible with new queries
// (the admin embedding screen already warns about this). This file turns
// those risks into explicit, actionable refresh tasks:
//
//   - embedding_mismatch: chunks stamped with a provider/model other than
//     the one now configured → they must be re-embedded before queries can
//     trust them.
//   - stale_corpus: the newest chunk is older than the policy age → the
//     knowledge should be refreshed.
//   - orphan_reference: a saved answer (provenance, C1) cites a chunk that
//     no longer exists (re-ingested/removed) → the citation no longer
//     verifies (P9 feedback).
//   - empty_corpus: an expert has no chunks at all.
//   - corpus_contradiction: separate sources under one topic contain explicit
//     opposite-polarity, strongly overlapping statements; they need review.
//
// Signals are deterministic (counts/ages/model comparison/string checks) — no
// LLM — so scans are cheap and reproducible. Contradiction findings are review
// candidates, never an automatic choice of which source is correct.
package knowledge

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Task types / severities / statuses (stable DB + API values).
const (
	TaskEmbeddingMismatch   = "embedding_mismatch"
	TaskStaleCorpus         = "stale_corpus"
	TaskOrphanReference     = "orphan_reference"
	TaskEmptyCorpus         = "empty_corpus"
	TaskCorpusContradiction = "corpus_contradiction"

	SeverityLow    = "low"
	SeverityMedium = "medium"
	SeverityHigh   = "high"

	StatusOpen         = "open"
	StatusAcknowledged = "acknowledged"
	StatusResolved     = "resolved"

	// Freshness statuses (expert-level).
	FreshnessFresh        = "fresh"
	FreshnessStale        = "stale"
	FreshnessNeedsReembed = "needs_reembed"
	FreshnessEmpty        = "empty"
)

// Policy configures the freshness thresholds.
type Policy struct {
	// MaxCorpusAgeDays: newest chunk older than this → stale_corpus.
	// <= 0 → default 180.
	MaxCorpusAgeDays int
}

// Task is one refresh action.
type Task struct {
	ID         uuid.UUID       `json:"id"`
	ExpertID   uuid.UUID       `json:"expert_id"`
	ExpertName string          `json:"expert_name,omitempty"`
	TaskType   string          `json:"task_type"`
	Severity   string          `json:"severity"`
	Title      string          `json:"title"`
	Details    json.RawMessage `json:"details"`
	Status     string          `json:"status"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
	ResolvedAt *time.Time      `json:"resolved_at,omitempty"`
}

// ExpertFreshness is the expert-level summary.
type ExpertFreshness struct {
	ExpertID         uuid.UUID  `json:"expert_id"`
	ChunkCount       int        `json:"chunk_count"`
	NewestChunkAt    *time.Time `json:"newest_chunk_at,omitempty"`
	CorpusAgeDays    int        `json:"corpus_age_days"` // -1 = unknown (no chunks)
	Stale            bool       `json:"stale"`
	CurrentProvider  string     `json:"current_provider"`
	CurrentModel     string     `json:"current_model"`
	PresentModels    []string   `json:"present_models"`
	MismatchChunks   int        `json:"mismatch_chunks"`
	OrphanReferences int        `json:"orphan_references"`
	Status           string     `json:"status"`
	OpenTasks        int        `json:"open_tasks"`
	ScannedAt        time.Time  `json:"scanned_at"`
}

// signal is an internal candidate task.
type signal struct {
	TaskType  string
	Severity  string
	Title     string
	Details   map[string]interface{}
	DedupeKey string
}

// Freshness computes and persists knowledge-freshness signals.
type Freshness struct {
	db     *pgxpool.Pool
	policy Policy
	logger *zap.Logger
}

// NewFreshness builds the freshness service.
func NewFreshness(db *pgxpool.Pool, policy Policy, logger *zap.Logger) *Freshness {
	if policy.MaxCorpusAgeDays <= 0 {
		policy.MaxCorpusAgeDays = 180
	}
	return &Freshness{db: db, policy: policy, logger: logger}
}

// ScanAndMeasureAll refreshes deterministic freshness signals, then remeasures
// only experts whose corpus changed since the last completed eval. At most one
// expensive eval is started per scheduled sweep; the caller invokes this on a
// daily schedule and logs the returned error.
func (f *Freshness) ScanAndMeasureAll(ctx context.Context, measure func(context.Context, uuid.UUID) error) error {
	if !f.Enabled() {
		return fmt.Errorf("freshness service not enabled")
	}
	if _, err := f.ScanAll(ctx); err != nil {
		return err
	}
	if measure == nil {
		return nil
	}

	var expertID uuid.UUID
	err := f.db.QueryRow(ctx, `
		SELECT e.id
		  FROM experts e
		 WHERE e.deleted_at IS NULL
		   AND EXISTS (SELECT 1 FROM course_chunks cc WHERE cc.expert_id = e.id)
		   AND EXISTS (
		       SELECT 1 FROM expert_capabilities ec
		        WHERE ec.expert_id = e.id AND ec.topic IS NOT NULL AND ec.topic <> '' AND ec.topic <> 'general'
		   )
		   AND NOT EXISTS (
		       SELECT 1 FROM expert_capability_eval_runs r
		        WHERE r.expert_id = e.id AND r.status = 'running'
		   )
		   AND (
		       NOT EXISTS (SELECT 1 FROM expert_capability_eval_runs r WHERE r.expert_id=e.id AND r.status='complete')
		       OR EXISTS (
		           SELECT 1 FROM course_chunks cc
		            WHERE cc.expert_id=e.id AND NOT EXISTS (
		                SELECT 1 FROM expert_capability_eval_runs r
		                 WHERE r.expert_id=e.id AND r.status='complete'
		                   AND COALESCE(r.corpus_updated_at, r.completed_at) >= cc.created_at
		            )
		       )
		   )
		 ORDER BY e.created_at
		 LIMIT 1`).Scan(&expertID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("freshness: find changed expert for capability measurement: %w", err)
	}
	if err := measure(ctx, expertID); err != nil {
		return fmt.Errorf("freshness: measure changed expert %s: %w", expertID, err)
	}
	return nil
}

// Enabled reports whether the service can operate. Nil-safe.
func (f *Freshness) Enabled() bool { return f != nil && f.db != nil }

// MaxAgeDays exposes the configured threshold (for API responses).
func (f *Freshness) MaxAgeDays() int {
	if f == nil {
		return 180
	}
	return f.policy.MaxCorpusAgeDays
}

// ScanExpert recomputes signals for one expert and upserts its open tasks.
func (f *Freshness) ScanExpert(ctx context.Context, expertID uuid.UUID) (*ExpertFreshness, error) {
	if !f.Enabled() {
		return nil, fmt.Errorf("freshness service not enabled")
	}
	ef, signals, err := f.compute(ctx, expertID)
	if err != nil {
		return nil, err
	}
	contradictions, err := f.detectCorpusContradictions(ctx, expertID)
	if err != nil {
		return nil, fmt.Errorf("freshness: detect corpus contradictions: %w", err)
	}
	signals = append(signals, contradictions...)
	for _, s := range signals {
		if err := f.upsertTask(ctx, expertID, s); err != nil {
			return nil, err
		}
	}
	ef.OpenTasks = f.countOpenTasks(ctx, expertID)
	return ef, nil
}

// ScanAll scans every active expert. Returns per-expert results; individual
// failures are logged and skipped (one bad expert must not stop the scan).
func (f *Freshness) ScanAll(ctx context.Context) ([]ExpertFreshness, error) {
	if !f.Enabled() {
		return []ExpertFreshness{}, nil
	}
	rows, err := f.db.Query(ctx,
		`SELECT id FROM experts WHERE deleted_at IS NULL ORDER BY created_at`)
	if err != nil {
		return nil, fmt.Errorf("freshness scan: list experts: %w", err)
	}
	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			continue
		}
		ids = append(ids, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	out := make([]ExpertFreshness, 0, len(ids))
	for _, id := range ids {
		ef, err := f.ScanExpert(ctx, id)
		if err != nil {
			f.logger.Warn("freshness scan: expert failed", zap.String("expert_id", id.String()), zap.Error(err))
			continue
		}
		out = append(out, *ef)
	}
	return out, nil
}

// GetExpert returns the current signals WITHOUT writing tasks (read-only).
func (f *Freshness) GetExpert(ctx context.Context, expertID uuid.UUID) (*ExpertFreshness, error) {
	if !f.Enabled() {
		return nil, fmt.Errorf("freshness service not enabled")
	}
	ef, _, err := f.compute(ctx, expertID)
	if err != nil {
		return nil, err
	}
	ef.OpenTasks = f.countOpenTasks(ctx, expertID)
	return ef, nil
}

// detectCorpusContradictions flags direct, deterministic polarity conflicts
// between different source files sharing a topic. It never resolves the conflict.
func (f *Freshness) detectCorpusContradictions(ctx context.Context, expertID uuid.UUID) ([]signal, error) {
	rows, err := f.db.Query(ctx, `
		WITH topics AS (
		    SELECT topic FROM course_chunks
		     WHERE expert_id=$1 AND topic IS NOT NULL AND topic<>''
	     GROUP BY topic ORDER BY COUNT(*) DESC, topic LIMIT 100
		), ranked_sources AS (
		    SELECT cc.topic, COALESCE(cc.source_file, '') AS source_file, cc.chunk_text,
		           ROW_NUMBER() OVER (PARTITION BY cc.topic, COALESCE(cc.source_file, '') ORDER BY cc.chunk_index, cc.id) AS source_rank,
		           DENSE_RANK() OVER (PARTITION BY cc.topic ORDER BY COALESCE(cc.source_file, '')) AS file_rank
		      FROM course_chunks cc JOIN topics t USING (topic)
		     WHERE cc.expert_id=$1 AND cc.chunk_text<>''
		)
		SELECT topic, source_file, chunk_text FROM ranked_sources
		 WHERE source_rank=1 AND file_rank<=16
		 ORDER BY topic, source_file`, expertID)
	if err != nil {
		return nil, fmt.Errorf("query corpus statements: %w", err)
	}
	defer rows.Close()
	type statement struct{ topic, source, text string }
	byTopic := make(map[string][]statement)
	seen := make(map[string]bool)
	for rows.Next() {
		var s statement
		if err := rows.Scan(&s.topic, &s.source, &s.text); err != nil {
			return nil, fmt.Errorf("scan corpus statement: %w", err)
		}
		key := s.topic + "\x00" + s.source
		if seen[key] {
			continue
		}
		seen[key] = true
		s.text = strings.TrimSpace(s.text)
		if len(s.text) > 1200 {
			s.text = s.text[:1200]
		}
		byTopic[s.topic] = append(byTopic[s.topic], s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read corpus statements: %w", err)
	}
	var out []signal
	for topic, sources := range byTopic {
		for i := 0; i < len(sources); i++ {
			for j := i + 1; j < len(sources); j++ {
				if sources[i].source == sources[j].source {
					continue
				}
				term, ok := polarityConflict(sources[i].text, sources[j].text)
				if !ok {
					continue
				}
				keyBytes := sha256.Sum256([]byte(topic + "\x00" + sources[i].source + "\x00" + sources[j].source + "\x00" + term))
				out = append(out, signal{
					TaskType: TaskCorpusContradiction, Severity: SeverityHigh,
					Title: "Knowledge sources may contradict each other",
					Details: map[string]interface{}{
						"topic": topic, "source_a": sources[i].source, "source_b": sources[j].source,
						"statement_a": sources[i].text, "statement_b": sources[j].text,
						"overlap_term": term, "resolution": "review_sources",
					},
					DedupeKey: hex.EncodeToString(keyBytes[:]),
				})
			}
		}
	}
	return out, nil
}

func polarityConflict(a, b string) (string, bool) {
	words := func(s string) []string {
		return strings.Fields(strings.Map(func(r rune) rune {
			if r >= 'A' && r <= 'Z' {
				r += 'a' - 'A'
			}
			if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
				return r
			}
			return ' '
		}, s))
	}
	left, right := words(a), words(b)
	negative := func(ws []string) bool {
		for _, w := range ws {
			switch w {
			case "not", "never", "cannot", "cant", "no", "isnt", "doesnt":
				return true
			}
		}
		return false
	}
	if negative(left) == negative(right) {
		return "", false
	}
	stop := map[string]bool{"not": true, "never": true, "cannot": true, "cant": true, "no": true, "isnt": true, "doesnt": true, "this": true, "that": true, "with": true, "from": true, "into": true, "when": true, "then": true, "than": true, "which": true, "their": true, "there": true, "these": true, "those": true, "about": true, "should": true, "would": true, "could": true, "must": true, "will": true, "have": true, "has": true, "does": true, "only": true, "also": true, "used": true, "using": true, "based": true, "after": true, "before": true, "under": true, "over": true, "between": true, "each": true, "such": true, "more": true, "most": true, "some": true, "many": true, "other": true, "same": true, "different": true, "true": true, "false": true}
	terms := make(map[string]bool, len(left))
	for _, w := range left {
		if len(w) >= 5 && !stop[w] {
			terms[w] = true
		}
	}
	for _, w := range right {
		if terms[w] {
			return w, true
		}
	}
	return "", false
}

// ListTasks returns refresh tasks, filtered by status and/or expert.
func (f *Freshness) ListTasks(ctx context.Context, status string, expertID *uuid.UUID, limit int) ([]Task, error) {
	if !f.Enabled() {
		return []Task{}, nil
	}
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	rows, err := f.db.Query(ctx, `
		SELECT t.id, t.expert_id, COALESCE(e.name,''), t.task_type, t.severity,
		       t.title, t.details, t.status, t.created_at, t.updated_at, t.resolved_at
		FROM knowledge_refresh_tasks t
		LEFT JOIN experts e ON e.id = t.expert_id
		WHERE ($1 = '' OR t.status = $1)
		  AND ($2::uuid IS NULL OR t.expert_id = $2)
		ORDER BY CASE t.severity WHEN 'high' THEN 0 WHEN 'medium' THEN 1 ELSE 2 END,
		         t.created_at DESC
		LIMIT $3`,
		status, expertID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("freshness list tasks: %w", err)
	}
	defer rows.Close()

	out := []Task{}
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.ExpertID, &t.ExpertName, &t.TaskType, &t.Severity,
			&t.Title, &t.Details, &t.Status, &t.CreatedAt, &t.UpdatedAt, &t.ResolvedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// AcknowledgeTask marks an open task acknowledged.
func (f *Freshness) AcknowledgeTask(ctx context.Context, taskID uuid.UUID) error {
	if !f.Enabled() {
		return fmt.Errorf("freshness service not enabled")
	}
	tag, err := f.db.Exec(ctx,
		`UPDATE knowledge_refresh_tasks SET status='acknowledged', updated_at=NOW()
		 WHERE id=$1 AND status='open'`, taskID)
	if err != nil {
		return fmt.Errorf("acknowledge task: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("task not found or already acknowledged")
	}
	return nil
}

// ResolveTask marks a task resolved.
func (f *Freshness) ResolveTask(ctx context.Context, taskID uuid.UUID) error {
	if !f.Enabled() {
		return fmt.Errorf("freshness service not enabled")
	}
	tag, err := f.db.Exec(ctx,
		`UPDATE knowledge_refresh_tasks SET status='resolved', resolved_at=NOW(), updated_at=NOW()
		 WHERE id=$1 AND status IN ('open','acknowledged')`, taskID)
	if err != nil {
		return fmt.Errorf("resolve task: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("task not found or already resolved")
	}
	return nil
}

// ---- internals ----

func (f *Freshness) compute(ctx context.Context, expertID uuid.UUID) (*ExpertFreshness, []signal, error) {
	now := time.Now().UTC()
	ef := &ExpertFreshness{ExpertID: expertID, ScannedAt: now, PresentModels: []string{}}

	var newest *time.Time
	if err := f.db.QueryRow(ctx,
		`SELECT COUNT(*), MAX(created_at) FROM course_chunks WHERE expert_id=$1`, expertID,
	).Scan(&ef.ChunkCount, &newest); err != nil {
		return nil, nil, fmt.Errorf("freshness: count chunks: %w", err)
	}
	ef.NewestChunkAt = newest
	ef.CorpusAgeDays = CorpusAgeDays(newest, now)

	ef.CurrentProvider, ef.CurrentModel = f.currentEmbedding(ctx)

	// Distinct models present in the corpus (for display + mismatch detail).
	if rows, err := f.db.Query(ctx,
		`SELECT DISTINCT embedding_model FROM course_chunks
		 WHERE expert_id=$1 AND embedding_model IS NOT NULL`, expertID); err == nil {
		for rows.Next() {
			var m string
			if rows.Scan(&m) == nil {
				ef.PresentModels = append(ef.PresentModels, m)
			}
		}
		rows.Close()
	}

	if err := f.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM course_chunks
		 WHERE expert_id=$1 AND embedding_provider IS NOT NULL
		   AND (embedding_provider <> $2
		        OR (embedding_model IS NOT NULL AND $3 <> '' AND embedding_model <> $3))`,
		expertID, ef.CurrentProvider, ef.CurrentModel,
	).Scan(&ef.MismatchChunks); err != nil {
		return nil, nil, fmt.Errorf("freshness: mismatch: %w", err)
	}

	if err := f.db.QueryRow(ctx, orphanReferencesSQL, expertID).Scan(&ef.OrphanReferences); err != nil {
		return nil, nil, fmt.Errorf("freshness: orphan refs: %w", err)
	}

	ef.Stale = IsStale(ef.CorpusAgeDays, f.policy.MaxCorpusAgeDays)
	ef.Status = Classify(ef.ChunkCount, ef.MismatchChunks, ef.Stale)

	return ef, f.signalsFor(ef), nil
}

func (f *Freshness) signalsFor(ef *ExpertFreshness) []signal {
	var out []signal

	if ef.ChunkCount == 0 {
		out = append(out, signal{
			TaskType: TaskEmptyCorpus, Severity: SeverityHigh,
			Title:   "Expert has no knowledge chunks",
			Details: map[string]interface{}{"chunk_count": 0},
		})
		return out // nothing else is meaningful on an empty corpus
	}

	if ef.MismatchChunks > 0 {
		out = append(out, signal{
			TaskType: TaskEmbeddingMismatch, Severity: SeverityHigh,
			Title: "Corpus embeddings do not match the active embedding model",
			Details: map[string]interface{}{
				"mismatch_chunks":  ef.MismatchChunks,
				"current_provider": ef.CurrentProvider,
				"current_model":    ef.CurrentModel,
				"present_models":   ef.PresentModels,
			},
			DedupeKey: ef.CurrentProvider + ":" + ef.CurrentModel,
		})
	}

	if ef.Stale {
		details := map[string]interface{}{
			"corpus_age_days": ef.CorpusAgeDays,
			"max_age_days":    f.policy.MaxCorpusAgeDays,
		}
		if ef.NewestChunkAt != nil {
			details["newest_chunk_at"] = ef.NewestChunkAt.Format(time.RFC3339)
		}
		out = append(out, signal{
			TaskType: TaskStaleCorpus, Severity: SeverityMedium,
			Title:   "Corpus has not been refreshed within the freshness window",
			Details: details,
		})
	}

	if ef.OrphanReferences > 0 {
		out = append(out, signal{
			TaskType: TaskOrphanReference, Severity: SeverityMedium,
			Title:   "Saved answers cite chunks that no longer exist",
			Details: map[string]interface{}{"orphan_references": ef.OrphanReferences},
		})
	}

	return out
}

func (f *Freshness) upsertTask(ctx context.Context, expertID uuid.UUID, s signal) error {
	details, err := json.Marshal(s.Details)
	if err != nil {
		details = []byte("{}")
	}
	_, err = f.db.Exec(ctx,
		`INSERT INTO knowledge_refresh_tasks (expert_id, task_type, severity, title, details, dedupe_key)
		 VALUES ($1,$2,$3,$4,$5,$6)
		 ON CONFLICT (expert_id, task_type, dedupe_key) DO UPDATE
		 SET severity   = EXCLUDED.severity,
		     title      = EXCLUDED.title,
		     details    = EXCLUDED.details,
		     updated_at = NOW(),
		     -- A resolved task that recurs is reopened; open/ack stay as-is.
		     status     = CASE WHEN knowledge_refresh_tasks.status='resolved' THEN 'open'
		                       ELSE knowledge_refresh_tasks.status END,
		     resolved_at = CASE WHEN knowledge_refresh_tasks.status='resolved' THEN NULL
		                        ELSE knowledge_refresh_tasks.resolved_at END`,
		expertID, s.TaskType, s.Severity, s.Title, details, s.DedupeKey,
	)
	if err != nil {
		return fmt.Errorf("freshness upsert task: %w", err)
	}
	return nil
}

func (f *Freshness) countOpenTasks(ctx context.Context, expertID uuid.UUID) int {
	var n int
	_ = f.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM knowledge_refresh_tasks WHERE expert_id=$1 AND status <> 'resolved'`,
		expertID).Scan(&n)
	return n
}

// currentEmbedding reads the active embedding provider/model from
// system_settings (the same keys the admin embedding screen writes).
func (f *Freshness) currentEmbedding(ctx context.Context) (provider, model string) {
	provider = "sidecar"
	readSetting := func(key string) string {
		var raw []byte
		if err := f.db.QueryRow(ctx,
			`SELECT value FROM system_settings WHERE key=$1`, key).Scan(&raw); err != nil {
			return ""
		}
		var v string
		if json.Unmarshal(raw, &v) != nil {
			return ""
		}
		return v
	}
	if p := readSetting("embedding_provider"); p != "" {
		provider = p
	}
	model = readSetting("embedding_model")
	return provider, model
}

// orphanReferencesSQL counts provenance citations (C1) whose chunk_id no
// longer exists in course_chunks for the same expert — a citation that can
// no longer be verified. jsonb_typeof guards against non-array values.
const orphanReferencesSQL = `
	SELECT COUNT(*)
	FROM provenance_records pr
	CROSS JOIN LATERAL jsonb_array_elements(
		CASE WHEN jsonb_typeof(pr.citations) = 'array' THEN pr.citations ELSE '[]'::jsonb END
	) AS c
	WHERE pr.expert_id = $1
	  AND (c->>'chunk_id') IS NOT NULL
	  AND NOT EXISTS (SELECT 1 FROM course_chunks cc WHERE cc.id::text = (c->>'chunk_id'))`

// ---- pure helpers (unit-tested) ----

// CorpusAgeDays returns whole days between newest and now. -1 when newest is
// nil (no chunks → age unknown, never stale). Negative deltas clamp to 0.
func CorpusAgeDays(newest *time.Time, now time.Time) int {
	if newest == nil {
		return -1
	}
	d := int(now.UTC().Sub(newest.UTC()).Hours() / 24)
	if d < 0 {
		return 0
	}
	return d
}

// IsStale reports whether an age (in days) exceeds the window. Unknown age
// (-1) and non-positive windows are never stale.
func IsStale(ageDays, maxDays int) bool {
	if ageDays < 0 || maxDays <= 0 {
		return false
	}
	return ageDays > maxDays
}

// Classify returns the expert-level freshness status.
func Classify(chunkCount, mismatchChunks int, stale bool) string {
	switch {
	case chunkCount == 0:
		return FreshnessEmpty
	case mismatchChunks > 0:
		return FreshnessNeedsReembed
	case stale:
		return FreshnessStale
	default:
		return FreshnessFresh
	}
}
