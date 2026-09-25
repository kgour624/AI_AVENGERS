package training

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Verification status values written to ingestion_runs.verification_status.
//
// These are strings in the DB (a CHECK constraint enforces the set) and
// constants here, so a typo is a compile error rather than a row that silently
// reads as "mismatch" on the audit screen.
const (
	// VerificationNotChecked: the check itself failed (DB error), so the run has
	// no verdict. Not the same as "verified" — an unread ledger is not a pass.
	VerificationNotChecked = "not_checked"

	// VerificationVerified: the storage identity held and no stored chunk is
	// missing its embedding.
	VerificationVerified = "verified"

	// VerificationMismatch: the identity failed — rows are actually missing.
	VerificationMismatch = "mismatch"
)

// StoreStats is the honest record of what one store pass did.
//
// WHY this exists: the pipeline used to claim it stored every chunk it parsed,
// while the INSERT silently skipped duplicates (`ON CONFLICT DO NOTHING`). The
// gap between the two was invisible — the command tag was discarded — so the
// only way to notice was to count rows by hand. These numbers come from the
// database's own row counts, not from what the pipeline intended.
type StoreStats struct {
	// Parsed is the number of chunks handed to the store.
	Parsed int

	// UniqueHashes is the number of distinct dedup keys among Parsed. The corpus
	// can hold at most this many rows for the run, because (expert_id,
	// chunk_hash) is unique.
	UniqueHashes int

	// Duplicates is Parsed - UniqueHashes: identical text appearing more than
	// once inside a single document. These are not lost — they are the same
	// content — but they are why a run can parse more chunks than it stores.
	Duplicates int

	// Inserted is the number of rows the INSERT statements actually created.
	Inserted int

	// Reused is the number of distinct chunks that already existed in the corpus
	// (append mode, or a resumed run). UniqueHashes - Inserted.
	Reused int

	// PreexistingForFile is how many rows already carried this expert+source_file
	// before the store. Read before the insert so the storage identity can be
	// checked afterwards: stored == PreexistingForFile + Inserted.
	PreexistingForFile int
}

// RunLedger is one ingestion run's accounting, ready to persist.
//
// It deliberately separates the RUN's claim (Stats, FallbackClaimed) from the
// DATABASE's reality (StoredForFile, GeneralStored, NullEmbeddings, CorpusTotal),
// because the whole point is that those two are not assumed to agree.
type RunLedger struct {
	JobID      uuid.UUID
	ExpertID   uuid.UUID
	SourceFile string

	// Stats is what the store pass reported.
	Stats StoreStats

	// FallbackClaimed counts chunks whose topic extraction failed and left them
	// with an empty topic (transient LLM error, batch skipped). Distinct from a
	// topic that is legitimately "general": this one means the model never
	// answered for those chunks, which is a corpus-quality signal.
	FallbackClaimed int

	// StoredForFile is the row count for this expert+source_file after the store.
	StoredForFile int

	// GeneralStored counts those rows whose topic is NULL/''/'general'.
	GeneralStored int

	// NullEmbeddings counts rows with no vector. Retrieval cannot see them, so a
	// non-zero value is a real defect even when every row is present.
	NullEmbeddings int

	// CorpusTotal is the expert's entire corpus after the run (append mode makes
	// this larger than StoredForFile).
	CorpusTotal int

	VerificationStatus string
	MismatchReason     string
}

// ExpectedStored is the row count the identity predicts for this expert+file.
//
// Mental execution: 277 rows already existed for the file, the store inserted 0
// new ones (a resumed run) → the file must still hold 277. If it holds 276, a
// row was lost, and that is what the verification reports.
func (l RunLedger) ExpectedStored() int {
	return l.Stats.PreexistingForFile + l.Stats.Inserted
}

// countDistinctHashes returns how many distinct dedup keys the run holds and how
// many of its chunks repeat a key already counted.
//
// Split out of storeChunks so the rule can be tested without a database: this is
// the arithmetic that decides whether parsed and stored are allowed to differ.
//
// Callers MUST set ChunkHash first — storeChunks backfills any empty hash before
// calling this. A hash-less chunk is not a repeat of anything, and counting it as
// one would understate the corpus.
func countDistinctHashes(chunks []TextChunk) (unique, duplicates int) {
	distinct := make(map[string]struct{}, len(chunks))
	for i := range chunks {
		distinct[chunks[i].ChunkHash] = struct{}{}
	}
	unique = len(distinct)
	return unique, len(chunks) - unique
}

// writeRunLedger persists one run's accounting.
//
// WHY non-fatal: the corpus is already stored by the time this runs. Failing the
// job because the audit row could not be written would turn a bookkeeping
// problem into a data problem — the same reasoning the event timeline uses.
func (p *IngestionPipeline) writeRunLedger(ctx context.Context, l RunLedger) {
	if l.VerificationStatus == "" {
		l.VerificationStatus = VerificationNotChecked
	}

	_, err := p.db.Exec(ctx, `
		INSERT INTO ingestion_runs (
			job_id, expert_id, source_file,
			chunks_parsed, chunks_duplicate, chunks_inserted, chunks_reused,
			chunks_preexisting_for_file, chunks_stored_for_file,
			chunks_general_stored, chunks_fallback, null_embeddings, corpus_total,
			verification_status, mismatch_reason, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, NOW())
		ON CONFLICT (job_id, source_file) DO UPDATE SET
			chunks_parsed               = EXCLUDED.chunks_parsed,
			chunks_duplicate            = EXCLUDED.chunks_duplicate,
			chunks_inserted             = EXCLUDED.chunks_inserted,
			chunks_reused               = EXCLUDED.chunks_reused,
			chunks_preexisting_for_file = EXCLUDED.chunks_preexisting_for_file,
			chunks_stored_for_file      = EXCLUDED.chunks_stored_for_file,
			chunks_general_stored       = EXCLUDED.chunks_general_stored,
			chunks_fallback             = EXCLUDED.chunks_fallback,
			null_embeddings             = EXCLUDED.null_embeddings,
			corpus_total                = EXCLUDED.corpus_total,
			verification_status         = EXCLUDED.verification_status,
			mismatch_reason             = EXCLUDED.mismatch_reason,
			updated_at                  = NOW()`,
		l.JobID, l.ExpertID, l.SourceFile,
		l.Stats.Parsed, l.Stats.Duplicates, l.Stats.Inserted, l.Stats.Reused,
		l.Stats.PreexistingForFile, l.StoredForFile,
		l.GeneralStored, l.FallbackClaimed, l.NullEmbeddings, l.CorpusTotal,
		l.VerificationStatus, l.MismatchReason,
	)
	if err != nil {
		p.logger.Warn("ingestion run ledger write failed (non-fatal)",
			zap.String("job_id", l.JobID.String()),
			zap.String("source_file", l.SourceFile),
			zap.Error(err),
		)
	}
}

// describeLedger renders the run for the timeline: parsed → duplicates merged →
// inserted → stored, so one event line explains the whole gap.
func (l RunLedger) describeLedger() string {
	return fmt.Sprintf(
		"parsed %d, duplicates merged %d, inserted %d, reused %d, stored %d (expected %d), corpus %d",
		l.Stats.Parsed, l.Stats.Duplicates, l.Stats.Inserted, l.Stats.Reused,
		l.StoredForFile, l.ExpectedStored(), l.CorpusTotal,
	)
}
