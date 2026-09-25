package training

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/ml"
)

// Reconciliation actions accepted by Reconcile.
//
// WHY named constants: the action names cross an HTTP boundary and are echoed
// back in the response, so a typo must be a compile error on the sending side and
// a rejected request on the receiving side — never a silently ignored string.
const (
	// ReconcileActionExpertStats recomputes experts.total_chunks / total_topics /
	// avg_depth_level from the corpus they are supposed to mirror.
	ReconcileActionExpertStats = "expert_stats"

	// ReconcileActionCapabilities rebuilds expert_capabilities from chunk topics.
	ReconcileActionCapabilities = "capabilities"

	// ReconcileActionEmbeddings fills in vectors for stored chunks that have
	// none. Bounded per call — see reconcileEmbeddingBatch.
	ReconcileActionEmbeddings = "embeddings"

	// ReconcileActionResolveRun marks one run's ledger row as repaired, recording
	// that an admin looked at the discrepancy. It changes no chunk data.
	ReconcileActionResolveRun = "resolve_run"
)

// File integrity verdicts. Strings, because they are shown to the admin and
// stored in the response, not compared inside the pipeline.
const (
	integrityComplete         = "complete"
	integrityDuplicatesMerged = "duplicates_merged"
	integrityTailMissing      = "tail_missing"
	integrityUnchecked        = "unchecked"
)

const (
	// reconcileAuditRunLimit caps the ledger rows returned for one expert. The
	// audit answers "is anything wrong now?", and the newest rows are the only
	// ones that can be.
	reconcileAuditRunLimit = 50

	// reconcileEmbeddingBatch bounds how many missing vectors one repair call
	// fills. WHY bounded: a corpus that was ingested before the sidecar existed
	// has every vector missing, and an unbounded repair would turn one admin click
	// into thousands of ML calls inside a single request. The response reports how
	// many remain, so the admin decides whether to continue.
	reconcileEmbeddingBatch = 100

	// reconcileEmbedBatchSize mirrors the ingestion pipeline's 25 — the same
	// reason applies (a 100-text batch on CPU can exceed the sidecar's HTTP
	// timeout).
	reconcileEmbedBatchSize = 25
)

// Reconciler audits an expert's corpus against the record of how it was built,
// and repairs the parts of it that are DERIVED rather than stored.
//
// WHY it exists: verification (Phase B) records what each run actually stored and
// Phase C refuses to call a degraded corpus a clean success — but neither could
// act on a run that already happened. This is the acting half: read the ledger,
// compare it with the corpus, and fix what is reconstructible.
//
// WHAT IT CANNOT DO: it cannot invent rows that were never stored. A run whose
// chunks are genuinely missing is reported, not repaired — the remedy there is a
// re-ingest. Everything this type writes is derived state that is reconstructible
// from course_chunks, with one exception (resolve_run, which only records a human
// decision).
type Reconciler struct {
	db     *pgxpool.Pool
	logger *zap.Logger

	// embedder is optional. When nil, the embeddings action reports itself as
	// skipped instead of failing — a repair screen must stay usable in a
	// deployment where the ML sidecar is not configured.
	embedder ml.Embedder
}

// NewReconciler builds a reconciler. embedder may be nil.
func NewReconciler(db *pgxpool.Pool, embedder ml.Embedder, logger *zap.Logger) *Reconciler {
	return &Reconciler{db: db, embedder: embedder, logger: logger}
}

// RunAudit is one ingestion_runs row as the audit screen reads it.
type RunAudit struct {
	JobID              uuid.UUID `json:"job_id"`
	SourceFile         string    `json:"source_file"`
	VerificationStatus string    `json:"verification_status"`
	MismatchReason     string    `json:"mismatch_reason"`
	Parsed             int       `json:"parsed"`
	Duplicates         int       `json:"duplicates"`
	Inserted           int       `json:"inserted"`
	Reused             int       `json:"reused"`
	StoredForFile      int       `json:"stored_for_file"`
	GeneralStored      int       `json:"general_stored"`
	FallbackChunks     int       `json:"fallback_chunks"`
	NullEmbeddings     int       `json:"null_embeddings"`
	CreatedAt          time.Time `json:"created_at"`
}

// ExpertAudit answers "is this expert's corpus what it claims to be?".
//
// It carries both the numbers and the sentences derived from them (Findings), so
// the UI does not have to re-implement the judgement — two places deciding what
// "drift" means is how they end up disagreeing.
type ExpertAudit struct {
	ExpertID uuid.UUID `json:"expert_id"`

	// Reality, read from course_chunks.
	CorpusChunks   int `json:"corpus_chunks"`
	CorpusTopics   int `json:"corpus_topics"`
	NullEmbeddings int `json:"null_embeddings"`
	GeneralChunks  int `json:"general_chunks"`

	// What the experts row claims. These are a cache of the corpus, and a cache
	// that disagrees with its source is the drift this audit exists to find.
	DeclaredChunks int  `json:"declared_chunks"`
	DeclaredTopics int  `json:"declared_topics"`
	StatsDrift     bool `json:"stats_drift"`

	CapabilityRows int `json:"capability_rows"`
	// StaleCapabilityRows counts capability topics that no longer have any chunk.
	// Reported, never deleted: an LLM-built capability may legitimately use a
	// topic name that does not appear verbatim in the chunks, so automatic pruning
	// would destroy real data.
	StaleCapabilityRows int  `json:"stale_capability_rows"`
	MissingCapabilities bool `json:"missing_capabilities"`

	MismatchedRuns int `json:"mismatched_runs"`
	UncheckedRuns  int `json:"unchecked_runs"`

	Runs     []RunAudit `json:"runs"`
	Findings []string   `json:"findings"`
}

// FileIntegrity explains why one file holds the number of rows it holds.
type FileIntegrity struct {
	SourceFile string `json:"source_file"`

	StoredRows         int `json:"stored_rows"`
	ExpectedFromLedger int `json:"expected_from_ledger"`

	MinIndex        int `json:"min_index"`
	MaxIndex        int `json:"max_index"`
	DistinctIndices int `json:"distinct_indices"`
	// MissingIndices counts positions skipped inside [MinIndex, MaxIndex].
	MissingIndices int `json:"missing_indices"`
	// Shortfall is ExpectedFromLedger - StoredRows. It is what the storage
	// identity failed by, and it is explained (never excused) by Verdict.
	Shortfall int `json:"shortfall"`
	// TailMissing means the LAST parsed chunk never landed, which is the
	// signature of a store that stopped early rather than a document that
	// repeated itself.
	TailMissing bool `json:"tail_missing"`

	NullEmbeddings int `json:"null_embeddings"`
	GeneralChunks  int `json:"general_chunks"`
	FallbackChunks int `json:"fallback_chunks"`

	Verdict string `json:"verdict"`
}

// RunDiagnostics is the per-file explanation for one job.
type RunDiagnostics struct {
	JobID    uuid.UUID       `json:"job_id"`
	ExpertID uuid.UUID       `json:"expert_id"`
	Runs     []RunAudit      `json:"runs"`
	Files    []FileIntegrity `json:"files"`
	Findings []string        `json:"findings"`
}

// ReconcileRequest is one repair pass. DryRun defaults to false here; the HTTP
// layer defaults it to true, so an omitted field cannot cause a write.
type ReconcileRequest struct {
	Actions    []string
	DryRun     bool
	JobID      *uuid.UUID
	SourceFile string
}

// RepairResult reports what one action did, or would do.
type RepairResult struct {
	Action string `json:"action"`
	// Applied is true only when the action really ran. A dry run, an unknown
	// action and a skipped action all report false, with Detail saying which.
	Applied bool   `json:"applied"`
	DryRun  bool   `json:"dry_run"`
	Changed int    `json:"changed"`
	Detail  string `json:"detail"`
}

// classifyFileIntegrity decides why a file holds fewer rows than the run parsed.
//
// This is the diagnostic the ledger alone cannot give: `stored 277, parsed 395`
// says something differs but not what. Two causes look identical in the counts
// and need opposite responses:
//
//   - duplicates_merged: the document repeated text the corpus already held, so
//     the conflict clause kept one row per distinct chunk. Nothing is lost; the
//     content is present once, which is what retrieval wants.
//   - tail_missing: the highest stored index is below the last parsed index, so
//     the store stopped early and real content is absent. This needs a re-ingest.
//
// Splitting them is why index continuity is checked rather than just the totals.
// Pure on purpose: the judgement is worth testing without a database.
func classifyFileIntegrity(sourceFile string, storedRows, minIndex, maxIndex, distinctIndices, expected int) FileIntegrity {
	out := FileIntegrity{
		SourceFile:         sourceFile,
		StoredRows:         storedRows,
		ExpectedFromLedger: expected,
		MinIndex:           minIndex,
		MaxIndex:           maxIndex,
		DistinctIndices:    distinctIndices,
	}

	if storedRows == 0 {
		// No ledger row, or nothing stored for this file. "unchecked" rather than
		// "complete": an absent expectation is not evidence of correctness.
		out.Verdict = integrityUnchecked
		return out
	}

	out.Shortfall = expected - storedRows
	if out.Shortfall < 0 {
		// More rows than parsed means something else wrote to this file (an
		// earlier append, or a second run). Not a loss, so no shortfall.
		out.Shortfall = 0
	}
	if maxIndex >= minIndex {
		span := maxIndex - minIndex + 1
		if distinct := span - distinctIndices; distinct > 0 {
			out.MissingIndices = distinct
		}
	}
	// The last parsed chunk sits at index expected-1. If the highest stored index
	// is below that, the end of the document never landed.
	out.TailMissing = expected > 0 && maxIndex < expected-1

	switch {
	case out.TailMissing:
		out.Verdict = integrityTailMissing
	case out.Shortfall == 0 && out.MissingIndices == 0:
		out.Verdict = integrityComplete
	default:
		out.Verdict = integrityDuplicatesMerged
	}
	return out
}

// classifySentence renders a verdict for the audit screen.
func (f FileIntegrity) classifySentence() string {
	switch f.Verdict {
	case integrityComplete:
		return fmt.Sprintf("%s: all %d chunks present", f.SourceFile, f.StoredRows)
	case integrityTailMissing:
		return fmt.Sprintf("%s: store stopped early — %d of %d chunks stored, highest index %d of %d",
			f.SourceFile, f.StoredRows, f.ExpectedFromLedger, f.MaxIndex, f.ExpectedFromLedger-1)
	case integrityDuplicatesMerged:
		return fmt.Sprintf("%s: %d of %d rows stored, %d repeat text already in the corpus",
			f.SourceFile, f.StoredRows, f.ExpectedFromLedger, f.Shortfall)
	default:
		return fmt.Sprintf("%s: no stored chunks recorded", f.SourceFile)
	}
}

// Audit reads the corpus, the experts cache and the run ledger for one expert.
func (r *Reconciler) Audit(ctx context.Context, expertID uuid.UUID) (*ExpertAudit, error) {
	out := &ExpertAudit{ExpertID: expertID, Runs: []RunAudit{}, Findings: []string{}}

	if err := r.db.QueryRow(ctx, `
		SELECT COUNT(*),
		       COUNT(DISTINCT topic),
		       COUNT(*) FILTER (WHERE embedding IS NULL),
		       COUNT(*) FILTER (WHERE topic IS NULL OR topic = '' OR topic = 'general')
		  FROM course_chunks
		 WHERE expert_id = $1`, expertID,
	).Scan(&out.CorpusChunks, &out.CorpusTopics, &out.NullEmbeddings, &out.GeneralChunks); err != nil {
		return nil, fmt.Errorf("audit: read corpus: %w", err)
	}

	if err := r.db.QueryRow(ctx, `
		SELECT COALESCE(total_chunks,0), COALESCE(total_topics,0)
		  FROM experts WHERE id = $1`, expertID,
	).Scan(&out.DeclaredChunks, &out.DeclaredTopics); err != nil {
		return nil, fmt.Errorf("audit: read expert: %w", err)
	}
	out.StatsDrift = out.DeclaredChunks != out.CorpusChunks

	if err := r.db.QueryRow(ctx, `
		SELECT COUNT(*),
		       COUNT(*) FILTER (WHERE NOT EXISTS (
		           SELECT 1 FROM course_chunks k
		            WHERE k.expert_id = c.expert_id AND k.topic = c.topic
		       ))
		  FROM expert_capabilities c
		 WHERE c.expert_id = $1`, expertID,
	).Scan(&out.CapabilityRows, &out.StaleCapabilityRows); err != nil {
		return nil, fmt.Errorf("audit: read capabilities: %w", err)
	}
	out.MissingCapabilities = out.CapabilityRows == 0 && out.CorpusChunks > 0

	rows, err := r.db.Query(ctx, `
		SELECT job_id, source_file, verification_status, mismatch_reason,
		       chunks_parsed, chunks_duplicate, chunks_inserted, chunks_reused,
		       chunks_stored_for_file, chunks_general_stored, chunks_fallback,
		       null_embeddings, created_at
		  FROM ingestion_runs
		 WHERE expert_id = $1
		 ORDER BY created_at DESC
		 LIMIT $2`, expertID, reconcileAuditRunLimit)
	if err != nil {
		return nil, fmt.Errorf("audit: read runs: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var run RunAudit
		if scanErr := rows.Scan(
			&run.JobID, &run.SourceFile, &run.VerificationStatus, &run.MismatchReason,
			&run.Parsed, &run.Duplicates, &run.Inserted, &run.Reused,
			&run.StoredForFile, &run.GeneralStored, &run.FallbackChunks,
			&run.NullEmbeddings, &run.CreatedAt,
		); scanErr != nil {
			return nil, fmt.Errorf("audit: scan run: %w", scanErr)
		}
		switch run.VerificationStatus {
		case VerificationMismatch:
			out.MismatchedRuns++
		case VerificationNotChecked:
			out.UncheckedRuns++
		}
		out.Runs = append(out.Runs, run)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("audit: read runs: %w", err)
	}

	out.Findings = r.auditFindings(out)
	return out, nil
}

// auditFindings turns the audit's numbers into sentences, in the order an admin
// should act on them.
func (r *Reconciler) auditFindings(a *ExpertAudit) []string {
	findings := make([]string, 0, 4)

	if a.CorpusChunks == 0 {
		findings = append(findings, "This expert has no stored chunks — nothing is retrievable yet.")
	}
	if a.MismatchedRuns > 0 {
		findings = append(findings, fmt.Sprintf(
			"%d run(s) failed storage verification — open the diagnostics to see whether rows were lost or were duplicates.",
			a.MismatchedRuns))
	}
	if a.UncheckedRuns > 0 {
		findings = append(findings, fmt.Sprintf(
			"%d run(s) have no verification verdict, so they cannot be trusted either way.", a.UncheckedRuns))
	}
	if a.StatsDrift {
		findings = append(findings, fmt.Sprintf(
			"Expert totals are stale: the experts row says %d chunks, the corpus holds %d. Repair: expert_stats.",
			a.DeclaredChunks, a.CorpusChunks))
	}
	if a.MissingCapabilities {
		findings = append(findings, "Capabilities are empty while chunks exist — the expert cannot be matched to topics. Repair: capabilities.")
	}
	if a.StaleCapabilityRows > 0 {
		findings = append(findings, fmt.Sprintf(
			"%d capability topic(s) have no chunks. Reported only — a topic name built by the LLM need not match chunk text, so this is never deleted automatically.",
			a.StaleCapabilityRows))
	}
	if a.NullEmbeddings > 0 {
		findings = append(findings, fmt.Sprintf(
			"%d stored chunk(s) have no embedding and are invisible to retrieval. Repair: embeddings.",
			a.NullEmbeddings))
	}
	if a.GeneralChunks > 0 && a.CorpusChunks > 0 {
		pct := (a.GeneralChunks * 100) / a.CorpusChunks
		if pct >= 50 {
			findings = append(findings, fmt.Sprintf(
				"%d%% of chunks have a general/empty topic — topic extraction likely failed for part of the corpus.",
				pct))
		}
	}

	if len(findings) == 0 {
		findings = append(findings, "No drift found: the corpus, the expert totals and the run ledger agree.")
	}
	return findings
}

// Diagnose explains one job file by file.
//
// Ownership is enforced in SQL (expert_id = $1), not by comparing ids afterwards,
// so a guessed job id returns nothing rather than another expert's data.
func (r *Reconciler) Diagnose(ctx context.Context, expertID, jobID uuid.UUID) (*RunDiagnostics, error) {
	out := &RunDiagnostics{JobID: jobID, ExpertID: expertID, Runs: []RunAudit{}, Files: []FileIntegrity{}, Findings: []string{}}

	rows, err := r.db.Query(ctx, `
		SELECT job_id, source_file, verification_status, mismatch_reason,
		       chunks_parsed, chunks_duplicate, chunks_inserted, chunks_reused,
		       chunks_stored_for_file, chunks_general_stored, chunks_fallback,
		       null_embeddings, created_at
		  FROM ingestion_runs
		 WHERE expert_id = $1 AND job_id = $2
		 ORDER BY created_at DESC`, expertID, jobID)
	if err != nil {
		return nil, fmt.Errorf("diagnose: read runs: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var run RunAudit
		if scanErr := rows.Scan(
			&run.JobID, &run.SourceFile, &run.VerificationStatus, &run.MismatchReason,
			&run.Parsed, &run.Duplicates, &run.Inserted, &run.Reused,
			&run.StoredForFile, &run.GeneralStored, &run.FallbackChunks,
			&run.NullEmbeddings, &run.CreatedAt,
		); scanErr != nil {
			return nil, fmt.Errorf("diagnose: scan run: %w", scanErr)
		}
		out.Runs = append(out.Runs, run)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("diagnose: read runs: %w", err)
	}

	for _, run := range out.Runs {
		var stored, minIndex, maxIndex, distinctIndices, nullEmbeddings, general, fallback int
		if scanErr := r.db.QueryRow(ctx, `
			SELECT COUNT(*),
			       COALESCE(MIN(chunk_index),0),
			       COALESCE(MAX(chunk_index),0),
			       COUNT(DISTINCT chunk_index),
			       COUNT(*) FILTER (WHERE embedding IS NULL),
			       COUNT(*) FILTER (WHERE topic IS NULL OR topic = '' OR topic = 'general'),
			       COUNT(*) FILTER (WHERE topic IS NULL OR topic = '')
			  FROM course_chunks
			 WHERE expert_id = $1 AND source_file = $2`, expertID, run.SourceFile,
		).Scan(&stored, &minIndex, &maxIndex, &distinctIndices, &nullEmbeddings, &general, &fallback); scanErr != nil {
			return nil, fmt.Errorf("diagnose: read file %q: %w", run.SourceFile, scanErr)
		}

		file := classifyFileIntegrity(run.SourceFile, stored, minIndex, maxIndex, distinctIndices, run.Parsed)
		file.NullEmbeddings = nullEmbeddings
		file.GeneralChunks = general
		file.FallbackChunks = fallback
		out.Files = append(out.Files, file)

		if file.Verdict == integrityTailMissing {
			out.Findings = append(out.Findings, fmt.Sprintf(
				"%s is INCOMPLETE — the store stopped before the end of the document. Re-ingest this file; repair cannot recreate the missing rows.",
				run.SourceFile))
		}
	}
	for _, file := range out.Files {
		if file.Verdict == integrityDuplicatesMerged {
			out.Findings = append(out.Findings, file.classifySentence())
		}
	}
	if len(out.Findings) == 0 && len(out.Runs) > 0 {
		out.Findings = append(out.Findings, "No storage discrepancy recorded for this job.")
	}
	if len(out.Runs) == 0 {
		out.Findings = append(out.Findings, "No ledger row for this job — it ran before run accounting existed, or the ledger write failed.")
	}
	return out, nil
}

// ValidateActions rejects an unknown action before anything runs, so a request
// cannot apply half of its actions and then fail on a typo.
func ValidateActions(actions []string) error {
	if len(actions) == 0 {
		return fmt.Errorf("no actions given — one of %s is required", strings.Join(ReconcileActions(), ", "))
	}
	for _, action := range actions {
		known := false
		for _, valid := range ReconcileActions() {
			if action == valid {
				known = true
				break
			}
		}
		if !known {
			return fmt.Errorf("unknown action %q — valid actions are %s", action, strings.Join(ReconcileActions(), ", "))
		}
	}
	return nil
}

// ReconcileActions lists the accepted actions, in the order they are worth
// running: rebuild the corpus's derived state first, then record the decision.
func ReconcileActions() []string {
	return []string{
		ReconcileActionExpertStats,
		ReconcileActionCapabilities,
		ReconcileActionEmbeddings,
		ReconcileActionResolveRun,
	}
}

// Reconcile runs the requested repairs.
//
// DryRun computes and reports without writing, which is the intended default at
// the HTTP layer: a repair screen must be safe to open and look at.
func (r *Reconciler) Reconcile(ctx context.Context, expertID uuid.UUID, req ReconcileRequest) ([]RepairResult, error) {
	if err := ValidateActions(req.Actions); err != nil {
		return nil, err
	}

	results := make([]RepairResult, 0, len(req.Actions))
	for _, action := range req.Actions {
		var (
			result RepairResult
			err    error
		)
		switch action {
		case ReconcileActionExpertStats:
			result, err = r.repairExpertStats(ctx, expertID, req.DryRun)
		case ReconcileActionCapabilities:
			result, err = r.repairCapabilities(ctx, expertID, req.DryRun)
		case ReconcileActionEmbeddings:
			result, err = r.repairEmbeddings(ctx, expertID, req.DryRun)
		case ReconcileActionResolveRun:
			result, err = r.resolveRun(ctx, expertID, req)
		default:
			return results, fmt.Errorf("unknown action %q", action)
		}
		if err != nil {
			return results, fmt.Errorf("action %s: %w", action, err)
		}
		results = append(results, result)
	}
	return results, nil
}

// repairExpertStats recomputes the denormalised totals on the experts row.
//
// Safe to write: every value is a COUNT over course_chunks/expert_capabilities, so
// the "repair" is just making the cache agree with its source. Nothing here can
// destroy information — the worst case is that it recomputes the same numbers.
func (r *Reconciler) repairExpertStats(ctx context.Context, expertID uuid.UUID, dryRun bool) (RepairResult, error) {
	result := RepairResult{Action: ReconcileActionExpertStats, DryRun: dryRun}

	var corpusChunks, corpusTopics int
	if err := r.db.QueryRow(ctx, `
		SELECT COUNT(*), COUNT(DISTINCT topic)
		  FROM course_chunks WHERE expert_id = $1`, expertID,
	).Scan(&corpusChunks, &corpusTopics); err != nil {
		return result, fmt.Errorf("count corpus: %w", err)
	}

	var avgDepth float64
	if err := r.db.QueryRow(ctx, `
		SELECT COALESCE(AVG(depth_level), 0)
		  FROM expert_capabilities WHERE expert_id = $1`, expertID,
	).Scan(&avgDepth); err != nil {
		return result, fmt.Errorf("average depth: %w", err)
	}

	var declaredChunks, declaredTopics int
	var declaredDepth float64
	if err := r.db.QueryRow(ctx, `
		SELECT COALESCE(total_chunks,0), COALESCE(total_topics,0), COALESCE(avg_depth_level,0)
		  FROM experts WHERE id = $1`, expertID,
	).Scan(&declaredChunks, &declaredTopics, &declaredDepth); err != nil {
		return result, fmt.Errorf("read expert: %w", err)
	}

	if declaredChunks == corpusChunks && declaredTopics == corpusTopics && declaredDepth == avgDepth {
		result.Detail = fmt.Sprintf("already accurate (%d chunks, %d topics)", corpusChunks, corpusTopics)
		return result, nil
	}

	if dryRun {
		result.Changed = 1
		result.Detail = fmt.Sprintf("would set chunks %d -> %d, topics %d -> %d",
			declaredChunks, corpusChunks, declaredTopics, corpusTopics)
		return result, nil
	}

	if _, err := r.db.Exec(ctx, `
		UPDATE experts SET
			total_chunks = $1,
			total_topics = $2,
			avg_depth_level = $3,
			updated_at = NOW()
		 WHERE id = $4`, corpusChunks, corpusTopics, avgDepth, expertID); err != nil {
		return result, fmt.Errorf("update expert stats: %w", err)
	}

	result.Applied = true
	result.Changed = 1
	result.Detail = fmt.Sprintf("set chunks %d -> %d, topics %d -> %d",
		declaredChunks, corpusChunks, declaredTopics, corpusTopics)
	return result, nil
}

// repairCapabilities rebuilds expert_capabilities from chunk topics.
//
// It reuses the ingestion fallback's implementation rather than repeating the
// aggregation, so "capabilities from chunks" has one definition.
func (r *Reconciler) repairCapabilities(ctx context.Context, expertID uuid.UUID, dryRun bool) (RepairResult, error) {
	result := RepairResult{Action: ReconcileActionCapabilities, DryRun: dryRun}

	var topics int
	if err := r.db.QueryRow(ctx, `
		SELECT COUNT(DISTINCT topic)
		  FROM course_chunks
		 WHERE expert_id = $1 AND topic IS NOT NULL AND topic != ''`, expertID,
	).Scan(&topics); err != nil {
		return result, fmt.Errorf("count topics: %w", err)
	}

	if topics == 0 {
		result.Detail = "no chunk has a topic — nothing to rebuild; fix topic extraction first"
		return result, nil
	}

	if dryRun {
		result.Changed = topics
		result.Detail = fmt.Sprintf("would upsert %d capability topic(s)", topics)
		return result, nil
	}

	upserted, err := upsertCapabilitiesFromChunks(ctx, r.db, r.logger, expertID)
	if err != nil {
		return result, err
	}
	result.Applied = true
	result.Changed = upserted
	result.Detail = fmt.Sprintf("upserted %d capability topic(s)", upserted)
	return result, nil
}

// repairEmbeddings fills vectors for chunks that have none.
//
// Bounded per call (reconcileEmbeddingBatch). The provider/model stamp is copied
// from an already-embedded row of the same corpus, so a backfilled vector claims
// the same model as its neighbours — reading the current setting instead could
// stamp a vector with a model that never produced it.
func (r *Reconciler) repairEmbeddings(ctx context.Context, expertID uuid.UUID, dryRun bool) (RepairResult, error) {
	result := RepairResult{Action: ReconcileActionEmbeddings, DryRun: dryRun}

	var missing int
	if err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM course_chunks
		 WHERE expert_id = $1 AND embedding IS NULL`, expertID,
	).Scan(&missing); err != nil {
		return result, fmt.Errorf("count missing embeddings: %w", err)
	}
	if missing == 0 {
		result.Detail = "every stored chunk already has an embedding"
		return result, nil
	}

	if r.embedder == nil {
		result.Detail = fmt.Sprintf("%d chunk(s) have no embedding, and no embedder is wired in this deployment", missing)
		return result, nil
	}

	batch := missing
	if batch > reconcileEmbeddingBatch {
		batch = reconcileEmbeddingBatch
	}

	if dryRun {
		result.Changed = batch
		result.Detail = fmt.Sprintf("would embed %d of %d chunk(s) with no vector", batch, missing)
		return result, nil
	}

	// Stamp from the corpus itself, not from current settings.
	var provider, model string
	_ = r.db.QueryRow(ctx, `
		SELECT COALESCE(embedding_provider,''), COALESCE(embedding_model,'')
		  FROM course_chunks
		 WHERE expert_id = $1 AND embedding IS NOT NULL
		 LIMIT 1`, expertID).Scan(&provider, &model)
	var modelArg interface{}
	if model != "" {
		modelArg = model
	}

	rows, err := r.db.Query(ctx, `
		SELECT id, chunk_text FROM course_chunks
		 WHERE expert_id = $1 AND embedding IS NULL
		 ORDER BY chunk_index
		 LIMIT $2`, expertID, batch)
	if err != nil {
		return result, fmt.Errorf("read chunks without embeddings: %w", err)
	}
	type pending struct {
		id   uuid.UUID
		text string
	}
	var todo []pending
	for rows.Next() {
		var p pending
		if scanErr := rows.Scan(&p.id, &p.text); scanErr != nil {
			rows.Close()
			return result, fmt.Errorf("scan chunk: %w", scanErr)
		}
		todo = append(todo, p)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return result, fmt.Errorf("read chunks without embeddings: %w", err)
	}

	written := 0
	for start := 0; start < len(todo); start += reconcileEmbedBatchSize {
		end := start + reconcileEmbedBatchSize
		if end > len(todo) {
			end = len(todo)
		}
		texts := make([]string, 0, end-start)
		for _, p := range todo[start:end] {
			texts = append(texts, p.text)
		}

		vectors, embedErr := r.embedder.Embed(ctx, texts)
		if embedErr != nil {
			// Report progress rather than discarding it: the rows already written
			// are real, and the admin needs to know how far it got.
			result.Applied = written > 0
			result.Changed = written
			result.Detail = fmt.Sprintf("embedded %d of %d before failing: %v", written, batch, embedErr)
			return result, nil
		}
		if len(vectors) != len(texts) {
			result.Applied = written > 0
			result.Changed = written
			result.Detail = fmt.Sprintf("embedder returned %d vectors for %d texts after %d written",
				len(vectors), len(texts), written)
			return result, nil
		}

		for i, p := range todo[start:end] {
			if _, err := r.db.Exec(ctx, `
				UPDATE course_chunks
				   SET embedding = $1, embedding_provider = $2, embedding_model = $3
				 WHERE id = $4`,
				pgvector.NewVector(vectors[i]), provider, modelArg, p.id); err != nil {
				result.Applied = written > 0
				result.Changed = written
				result.Detail = fmt.Sprintf("embedded %d of %d before a write failed: %v", written, batch, err)
				return result, nil
			}
			written++
		}
	}

	result.Applied = true
	result.Changed = written
	result.Detail = fmt.Sprintf("embedded %d chunk(s); %d still without a vector — run again to continue",
		written, missing-written)
	return result, nil
}

// resolveRun records that an admin reviewed one run's discrepancy.
//
// WHY an explicit action and not an automatic side effect of the other repairs:
// the other repairs fix derived state, which says nothing about whether missing
// chunks are acceptable. Deciding that is a human judgement, so it is recorded as
// one, with the reason preserved on the row.
func (r *Reconciler) resolveRun(ctx context.Context, expertID uuid.UUID, req ReconcileRequest) (RepairResult, error) {
	result := RepairResult{Action: ReconcileActionResolveRun, DryRun: req.DryRun}

	if req.JobID == nil {
		result.Detail = "skipped: resolve_run needs job_id"
		return result, nil
	}

	note := strings.TrimSpace(req.SourceFile)
	if note == "" {
		note = "reviewed by admin via reconcile"
	} else {
		note = "reviewed by admin via reconcile: " + note
	}

	// Ownership in the WHERE clause, so a guessed job id cannot touch another
	// expert's ledger.
	var status string
	if err := r.db.QueryRow(ctx, `
		SELECT verification_status FROM ingestion_runs
		 WHERE job_id = $1 AND expert_id = $2
		 ORDER BY created_at DESC LIMIT 1`, *req.JobID, expertID).Scan(&status); err != nil {
		result.Detail = "skipped: no ledger row found for this job and expert"
		return result, nil
	}

	if status == VerificationVerified {
		result.Detail = "already verified — nothing to resolve"
		return result, nil
	}
	if req.DryRun {
		result.Changed = 1
		result.Detail = fmt.Sprintf("would mark this run as repaired (currently %s)", status)
		return result, nil
	}

	// Only this run's own table is touched. course_chunks is never modified by
	// this action: the rows that exist stay exactly as they are.
	tag, err := r.db.Exec(ctx, `
		UPDATE ingestion_runs
		   SET verification_status = 'repaired',
		       mismatch_reason = $1,
		       updated_at = NOW()
		 WHERE job_id = $2 AND expert_id = $3`, note, *req.JobID, expertID)
	if err != nil {
		return result, fmt.Errorf("mark run repaired: %w", err)
	}

	result.Applied = true
	result.Changed = int(tag.RowsAffected())
	result.Detail = fmt.Sprintf("marked %d run(s) as repaired (was %s)", result.Changed, status)
	return result, nil
}

// upsertCapabilitiesFromChunks populates expert_capabilities from chunk topics
// without an LLM, returning how many topics were written.
//
// Shared by the ingestion fallback and the reconcile action so there is ONE
// definition of what "capabilities from chunks" means; two copies would drift, and
// the drift would stay invisible until an expert's topics stopped matching the
// chunks behind them.
//
// It never deletes. A capability topic that no longer matches any chunk might be an
// LLM-normalised name rather than a stale row, so automatic pruning could destroy
// real data — the audit reports those rows instead.
func upsertCapabilitiesFromChunks(ctx context.Context, db *pgxpool.Pool, logger *zap.Logger, expertID uuid.UUID) (int, error) {
	rows, err := db.Query(ctx, `
		SELECT topic, COUNT(*) as chunk_count
		 FROM course_chunks
		 WHERE expert_id = $1
		   AND topic IS NOT NULL
		   AND topic != ''
		 GROUP BY topic
		 ORDER BY chunk_count DESC`, expertID)
	if err != nil {
		return 0, fmt.Errorf("read chunk topics: %w", err)
	}
	defer rows.Close()

	type topicCount struct {
		topic string
		count int
	}
	var topics []topicCount
	for rows.Next() {
		var t topicCount
		if scanErr := rows.Scan(&t.topic, &t.count); scanErr != nil {
			continue
		}
		topics = append(topics, t)
	}
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("read chunk topics: %w", err)
	}

	written := 0
	for _, t := range topics {
		if _, err := db.Exec(ctx, `
			INSERT INTO expert_capabilities
				(expert_id, topic, depth_level, chunk_count, complexity_ceiling)
			 VALUES ($1, $2, 1, $3, 'basic')
			 ON CONFLICT (expert_id, topic) DO UPDATE SET
				chunk_count = EXCLUDED.chunk_count,
				updated_at  = NOW()`, expertID, t.topic, t.count); err != nil {
			logger.Warn("upsert capability failed (non-fatal)",
				zap.String("expert_id", expertID.String()),
				zap.String("topic", t.topic),
				zap.Error(err),
			)
			continue
		}
		written++
	}
	return written, nil
}
