package training

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

	"ai_avengers/backend/internal/gateway"
	"ai_avengers/backend/internal/ml"
	"ai_avengers/backend/internal/observability"
)

// IngestionPipeline orchestrates the full transcript ingestion process.
// Flow: Load text -> Clean -> Chunk -> Extract topics -> Embed -> Store
//
// WHY this order:
// 1. Chunk first: Smaller units are better for topic extraction
// 2. Topic extract: Needed before storing (metadata)
// 3. Embed: Needed for vector search
// 4. Store: All data ready, single transaction
type IngestionPipeline struct {
	db       *pgxpool.Pool
	embedder ml.Embedder        // ml.Embedder interface: sidecar or CodeCraftAPI, resolved at call time
	sidecar  *ml.SidecarClient  // kept separately for Rerank() — Rerank is sidecar-only, not in Embedder interface
	gateway  *gateway.ModelGateway
	chunker  *TextChunker
	topics   *TopicExtractor
	charters *CharterExtractor
	capability *CapabilityBuilder
	logger   *zap.Logger
}

// NewIngestionPipeline creates a new ingestion pipeline.
//
// embedder satisfies ml.Embedder — either *ml.SidecarClient (default) or
// *ml.DynamicEmbedder (when CodeCraftAPI embeddings are enabled).
//
// sidecar is kept as a separate *ml.SidecarClient because the smoke test
// step calls Rerank(), which is sidecar-only and NOT part of the Embedder
// interface (locked decision: reranking always stays on the Python sidecar).
func NewIngestionPipeline(
	db *pgxpool.Pool,
	embedder ml.Embedder,
	sidecar *ml.SidecarClient,
	gw *gateway.ModelGateway,
	logger *zap.Logger,
) *IngestionPipeline {
	return &IngestionPipeline{
		db:         db,
		embedder:   embedder,
		sidecar:    sidecar,
		gateway:    gw,
		chunker:    NewTextChunker(DefaultChunkerConfig()),
		topics:     NewTopicExtractor(gw, embedder, logger),
		charters:   NewCharterExtractor(gw, logger),
		capability: NewCapabilityBuilder(gw, logger),
		logger:     logger,
	}
}

// currentEmbedding reads the active embedding provider/model from
// system_settings (same keys the admin embedding screen writes). Defaults to
// "sidecar" with an empty model. Best-effort — a read failure yields defaults
// so ingestion is never blocked by a missing settings row.
func (p *IngestionPipeline) currentEmbedding(ctx context.Context) (provider, model string) {
	provider = "sidecar"
	readSetting := func(key string) string {
		var raw []byte
		if err := p.db.QueryRow(ctx,
			`SELECT value FROM system_settings WHERE key=$1`, key).Scan(&raw); err != nil {
			return ""
		}
		var v string
		if json.Unmarshal(raw, &v) != nil {
			return ""
		}
		return v
	}
	if v := readSetting("embedding_provider"); v != "" {
		provider = v
	}
	model = readSetting("embedding_model")
	return provider, model
}

// IngestionResult holds statistics from a completed ingestion.
type IngestionResult struct {
	ExpertID      uuid.UUID
	TotalChunks   int
	TotalTopics   int
	AvgDepthLevel float64
	DurationMs    int64
}

// IngestTranscript runs the full ingestion pipeline for a transcript.
// Called by background worker after admin uploads a transcript.
//
// replaceExisting controls how the expert's existing corpus is treated:
//   - true: DELETE all existing chunks for this expert first, then insert new.
//     Use for full retraining when the source material is being replaced
//     end-to-end.
//   - false: append mode. New chunks are inserted; any chunk whose
//     (expert_id, chunk_hash) already exists is silently skipped via
//     ON CONFLICT ... DO NOTHING. Use when adding supplementary transcripts
//     to an existing corpus. This is the DEFAULT behavior per
//     DOMAIN_EXPERT_COLLABORATION_DESIGN.md §5.4.
//
// Mental execution:
// Input: expertID, transcript text, replaceExisting
// 1. Update job status to "running"
// 2. Chunk text (chunker computes ChunkHash on every chunk)
// 3. Extract topics (batched LLM calls)
// 4. Extract charters (strong LLM call)
// 5. Generate embeddings (ML sidecar)
// 6. If replaceExisting=true: DELETE existing chunks for expert
// 7. Store chunks with ON CONFLICT DO NOTHING (dedup)
// 8. Build capability table
// 9. Update expert stats
// 10. Update job status to "complete"
// On any error: update job status to "failed", cleanup partial data
func (p *IngestionPipeline) IngestTranscript(
	ctx context.Context,
	jobID uuid.UUID,
	expertID uuid.UUID,
	expertName string,
	transcript string,
	sourceFile string,
	replaceExisting bool,
) (*IngestionResult, error) {
	start := time.Now()
	// PhaseTimer records per-step latency for ingestion observability.
	// Logged at the end of IngestTranscript so admin can see which step
	// is the bottleneck (embed is typically 60-80% of total time).
	ptimer := observability.NewPhaseTimer()

	p.logger.Info("ingestion started",
		zap.String("expert_id", expertID.String()),
		zap.String("job_id", jobID.String()),
		zap.Int("transcript_length", len(transcript)),
	)

	// ============================================================
	// CHECKPOINT LOAD: resume from last known good state
	// ============================================================
	cp, err := LoadCheckpoint(ctx, p.db, jobID)
	if err != nil {
		p.logger.Warn("checkpoint load failed, starting fresh", zap.Error(err))
		cp = nil
	}
	isResume := cp != nil && cp.Stage != StagePending
	if isResume {
		p.logger.Info("resuming from checkpoint",
			zap.String("stage", cp.Stage),
			zap.Int("chunks_done", cp.ChunksDone),
			zap.Int("last_batch", cp.LastBatchIndex),
		)
		_, _ = p.db.Exec(ctx,
			`UPDATE ingestion_jobs SET resumed_from_checkpoint=TRUE WHERE id=$1`, jobID)
	}

	// Update job status to running
	p.updateJobStatus(ctx, jobID, "running", "", 0, 0)

	// ============================================================
	// STEP 0: TRANSCRIPT CLEANING (always runs, before chunking)
	// ============================================================
	// WHY clean before chunking:
	//   Raw transcripts contain 45-55% noise (greetings, timestamps,
	//   quiz mechanics, filler words). This noise gets embedded into
	//   chunks and lowers reranker scores below the China Wall threshold
	//   (0.35), causing Gate 2 to refuse all questions even when the
	//   expert has relevant knowledge.
	//
	//   Cleaning BEFORE chunking ensures every chunk contains only
	//   DSA/technical content. Noise patterns documented in:
	//   TRANSCRIPT_NOISE_TABLE.md
	if transcript != "" {
		cleaner := NewTranscriptCleaner()
		originalLen := len(transcript)
		ptimer.Start("clean")
		transcript = cleaner.Clean(transcript)
		ptimer.Stop("clean")
		cleanedLen := len(transcript)
		reductionPct := 0
		if originalLen > 0 {
			reductionPct = (originalLen - cleanedLen) * 100 / originalLen
		}
		p.logger.Info("transcript cleaned",
			zap.Int("original_chars", originalLen),
			zap.Int("cleaned_chars", cleanedLen),
			zap.Int("reduction_pct", reductionPct),
		)
		if cleanedLen == 0 {
			err := fmt.Errorf("transcript is empty after cleaning")
			p.updateJobStatus(ctx, jobID, "failed", err.Error(), 0, 0)
			return nil, err
		}
	}

	// ============================================================
	// STEP 1: CHUNKING
	// ============================================================
	// If resuming from topic_extraction or later, load chunks from DB
	// instead of re-chunking (chunking is deterministic but expensive for large transcripts).
	var chunks []TextChunk
	if isResume && StageOrder[cp.Stage] >= StageOrder[StageTopicExtraction] {
		// Load existing chunks from DB — scoped to THIS job's source file.
		// WHY sourceFile filter: without it, loadChunksFromDB returns ALL
		// chunks for the expert across every uploaded transcript file.
		// On resume, this inflates chunk count (e.g. 873 → 1741) and
		// causes topic extraction + embedding to re-process chunks from
		// OTHER files that were already correctly stored.
		chunks, err = p.loadChunksFromDB(ctx, expertID, sourceFile)
		if err != nil || len(chunks) == 0 {
			p.logger.Warn("could not load chunks from DB, re-chunking", zap.Error(err))
			chunks = p.chunker.Chunk(transcript)
		}
		p.logger.Info("resume: loaded chunks from DB", zap.Int("count", len(chunks)))
	} else {
		p.updateStage(ctx, jobID, StageChunking, "Splitting transcript...")
		ptimer.Start("chunk")
		chunks = p.chunker.Chunk(transcript)
		ptimer.Stop("chunk")
		if len(chunks) == 0 {
			err := fmt.Errorf("no chunks created from transcript")
			p.updateJobStatus(ctx, jobID, "failed", err.Error(), 0, 0)
			return nil, err
		}
	}
	p.logger.Info("chunking complete", zap.Int("chunks", len(chunks)))

	// Init progress tracker and checkpoint writer
	tracker := NewProgressTracker(p.db, jobID, len(chunks), p.logger)
	cpWriter := NewCheckpointWriter(p.db, jobID, p.logger)
	if isResume && cp != nil {
		tracker.costUSD = cp.CostUSDSoFar // restore accumulated cost
	}

	// ============================================================
	// STEP 2: TOPIC EXTRACTION (batched, resumable)
	// ============================================================
	p.updateStage(ctx, jobID, StageTopicExtraction,
		fmt.Sprintf("0/%d chunks tagged", len(chunks)))

	topicResults := make([]TopicResult, len(chunks))
	// Default fallback for all chunks
	for i := range topicResults {
		topicResults[i] = TopicResult{Topic: "general", Confidence: 0.5}
	}

	const topicBatchSize = 50
	resumeTopicBatch := 0
	if isResume && cp != nil && cp.Stage == StageTopicExtraction {
		resumeTopicBatch = cp.LastBatchIndex
		p.logger.Info("resume: skipping topic batches", zap.Int("skip_to_batch", resumeTopicBatch))
	}

	ptimer.Start("topic_extract")
	for batchStart := 0; batchStart < len(chunks); batchStart += topicBatchSize {
		batchIdx := batchStart / topicBatchSize
		// Skip already-processed batches on resume
		if batchIdx < resumeTopicBatch {
			continue
		}

		batchEnd := batchStart + topicBatchSize
		if batchEnd > len(chunks) {
			batchEnd = len(chunks)
		}
		batch := chunks[batchStart:batchEnd]

		batchResults, batchErr := p.topics.ExtractBatch(ctx, batch)
		if batchErr != nil {
			// Distinguish fatal errors (payment/auth) from transient errors.
			// Fatal: 402 (credits exhausted), 401 (invalid key), 403 (forbidden).
			//   → Pause the job. Continuing would store all remaining chunks
			//     with topic="general", destroying topic diversity and making
			//     Gate 2 fail for every future question. This is worse than
			//     not training at all.
			// Transient: 429 (rate limit), 5xx (server error), network timeout.
			//   → Use fallback topic="general" for this batch and continue.
			//     A few "general" chunks are acceptable; all chunks being
			//     "general" is not.
			if isFatalLLMError(batchErr) {
				p.logger.Error("topic extraction: fatal LLM error — pausing job to prevent all-general-topic disaster",
					zap.Int("batch", batchIdx),
					zap.Int("chunks_done", batchStart),
					zap.Int("chunks_total", len(chunks)),
					zap.Error(batchErr),
				)
				return nil, p.pauseOnLLMFailure(
					ctx, jobID, cpWriter,
					len(chunks), tracker.TotalCost(), start,
					fmt.Sprintf("Topic extraction fatal LLM error at batch %d: %s", batchIdx, batchErr.Error()),
				)
			}
			// Transient error — use fallback for this batch, continue.
			p.logger.Warn("topic batch failed (transient), using fallback topic",
				zap.Int("batch", batchIdx), zap.Error(batchErr))
		} else {
			copy(topicResults[batchStart:batchEnd], batchResults)
		}

		// Checkpoint every batch
		if batchIdx%1 == 0 { // every batch for topics (they're expensive)
			cpWriter.Write(ctx, JobCheckpoint{
				Stage:          StageTopicExtraction,
				ChunksDone:     batchEnd,
				ChunksTotal:    len(chunks),
				LastBatchIndex: batchIdx + 1,
				CostUSDSoFar:   tracker.TotalCost(),
				StartedAt:      start,
			})
			tracker.UpdateDB(ctx, batchEnd, StageTopicExtraction,
				fmt.Sprintf("%d/%d chunks tagged", batchEnd, len(chunks)))
		}
	}
	ptimer.Stop("topic_extract")
	p.logger.Info("topic extraction complete")

	// ============================================================
	// STEP 3: CHARTER EXTRACTION (single call, resumable)
	// ============================================================
	var charter *Charter
	charterAlreadyDone := isResume && cp != nil && cp.CharterExtracted

	if charterAlreadyDone {
		// Load existing charter from DB
		charter, err = p.loadCharterFromDB(ctx, expertID)
		if err != nil || charter == nil {
			p.logger.Warn("could not load charter from DB, re-extracting", zap.Error(err))
			charterAlreadyDone = false
		} else {
			p.logger.Info("resume: charter loaded from DB")
		}
	}

	if !charterAlreadyDone {
		p.updateStage(ctx, jobID, StageCharterExtraction, "Extracting expert charter...")
		ptimer.Start("charter_extract")
		extractedCharter, extractErr := p.charters.Extract(ctx, transcript, expertName)
		ptimer.Stop("charter_extract")
		if extractErr != nil {
			// Charter LLM failed (API error, rate limit, timeout, or bad output).
			// Do NOT silently fall back to a default charter — charter quality
			// is non-negotiable. A default charter produces generic, uncited
			// answers that defeat the purpose of domain-expert training.
			//
			// Instead: pause the job and wait for admin action.
			// Admin options (via admin panel):
			//   • Retry Now — re-run charter LLM from this stage
			//   • Pause & Wait — leave paused, resume manually later
			// Auto-fail: if paused for >24h without action (main.go checker).
			//
			// WHY pause not fail:
			//   LLM errors are transient. Chunks + topics are already in DB.
			//   Failing would require re-uploading the transcript and re-running
			//   all 6 stages from scratch. Pausing preserves all prior work.
			p.logger.Warn("charter extraction failed — pausing job for admin action",
				zap.String("expert_id", expertID.String()),
				zap.String("job_id", jobID.String()),
				zap.Error(extractErr),
			)
			return nil, p.pauseOnLLMFailure(
				ctx, jobID, cpWriter,
				len(chunks), tracker.TotalCost(), start,
				fmt.Sprintf("Charter LLM failed: %s", extractErr.Error()),
			)
		}
		charter = extractedCharter
		// Checkpoint: charter done
		cpWriter.Write(ctx, JobCheckpoint{
			Stage:            StageCharterExtraction,
			ChunksDone:       len(chunks),
			ChunksTotal:      len(chunks),
			CharterExtracted: true,
			LastBatchIndex:   0,
			CostUSDSoFar:     tracker.TotalCost(),
			StartedAt:        start,
		})
	}
	p.logger.Info("charter extraction complete")

	// ============================================================
	// STEP 4: EMBEDDING (batched, resumable)
	// ============================================================
	p.updateStage(ctx, jobID, StageEmbedding,
		fmt.Sprintf("0/%d embeddings generated", len(chunks)))

	embeddings := make([][]float32, len(chunks))

	// WHY 25 not 100:
	//   100 chunks × bge-base-en-v1.5 on CPU ≈ 40s > 30s HTTP timeout.
	//   25 chunks ≈ 10s per batch — well within 300s timeout.
	//   Smaller batches also mean more frequent checkpoints.
	const embedBatchSize = 25
	resumeEmbedBatch := 0
	if isResume && cp != nil && cp.Stage == StageEmbedding {
		resumeEmbedBatch = cp.LastBatchIndex
		p.logger.Info("resume: skipping embed batches", zap.Int("skip_to_batch", resumeEmbedBatch))
	}

	ptimer.Start("embed")
	for batchStart := 0; batchStart < len(chunks); batchStart += embedBatchSize {
		batchIdx := batchStart / embedBatchSize
		if batchIdx < resumeEmbedBatch {
			continue
		}

		batchEnd := batchStart + embedBatchSize
		if batchEnd > len(chunks) {
			batchEnd = len(chunks)
		}

		texts := make([]string, batchEnd-batchStart)
		for i, c := range chunks[batchStart:batchEnd] {
			texts[i] = c.Text
		}

		batchEmbeds, embedErr := p.embedder.Embed(ctx, texts)
		if embedErr != nil {
			p.updateJobStatus(ctx, jobID, "failed",
				"ML sidecar unavailable: "+embedErr.Error(), batchStart, len(chunks))
			return nil, fmt.Errorf("embedding batch %d failed: %w", batchIdx, embedErr)
		}
		copy(embeddings[batchStart:batchEnd], batchEmbeds)

		// Checkpoint every batch
		cpWriter.Write(ctx, JobCheckpoint{
			Stage:            StageEmbedding,
			ChunksDone:       batchEnd,
			ChunksTotal:      len(chunks),
			CharterExtracted: true,
			LastBatchIndex:   batchIdx + 1,
			CostUSDSoFar:     tracker.TotalCost(),
			StartedAt:        start,
		})
		tracker.UpdateDB(ctx, batchEnd, StageEmbedding,
			fmt.Sprintf("%d/%d embeddings generated", batchEnd, len(chunks)))
	}
	ptimer.Stop("embed")
	p.logger.Info("embeddings generated", zap.Int("count", len(embeddings)))

	// Step 5: Optional cleanup for full-replace mode.
	// Append-mode (replaceExisting=false) is the default and preserves the
	// existing corpus, relying on ON CONFLICT (expert_id, chunk_hash) DO NOTHING
	// in storeChunks() to skip duplicates.
	if replaceExisting {
		_, err = p.db.Exec(ctx,
			`DELETE FROM course_chunks WHERE expert_id = $1`,
			expertID,
		)
		if err != nil {
			p.updateJobStatus(ctx, jobID, "failed", "cleanup failed: "+err.Error(), 0, 0)
			return nil, fmt.Errorf("cleanup failed: %w", err)
		}
		p.logger.Info("replaceExisting=true: existing chunks deleted",
			zap.String("expert_id", expertID.String()),
		)
	}

	// Step 6: Store chunks. In append mode, ON CONFLICT dedups against
	// existing (expert_id, chunk_hash) pairs so repeat ingestion is idempotent.
	p.updateStage(ctx, jobID, StageStoring, fmt.Sprintf("Saving %d chunks to database...", len(chunks)))
	ptimer.Start("store")
	chunkIDs, err := p.storeChunks(ctx, jobID, expertID, chunks, topicResults, embeddings, sourceFile)
	ptimer.Stop("store")
	if err != nil {
		p.updateJobStatus(ctx, jobID, "failed", "storage failed: "+err.Error(), 0, 0)
		return nil, fmt.Errorf("chunk storage failed: %w", err)
	}
	p.logger.Info("chunks stored", zap.Int("count", len(chunkIDs)))

	// Step 7: Update expert charters
	//
	// CHARTER OVERWRITE POLICY:
	//   replaceExisting=true  → full retrain, overwrite charter completely
	//   replaceExisting=false → append mode, APPEND new charter to existing
	//                           WHY: admin uploads multiple transcripts for
	//                           the same expert. Each transcript adds new
	//                           principles. Overwriting destroys previous
	//                           training. Appending builds a richer charter.
	//
	// If existing charter is empty (first ingestion), just write the new one.
	clarificationJSON, _ := json.Marshal(charter.ClarificationCharter)

	if replaceExisting {
		// Full retrain: replace charter completely
		_, err = p.db.Exec(ctx,
			`UPDATE experts SET
				reasoning_charter = $1,
				clarification_charter = $2,
				updated_at = NOW()
			 WHERE id = $3`,
			charter.ReasoningCharter,
			string(clarificationJSON),
			expertID,
		)
	} else {
		// Append mode: append new charter to existing, preserve old principles
		// CASE WHEN existing is empty: just write new charter directly
		// CASE WHEN existing has content: append with separator
		_, err = p.db.Exec(ctx,
			`UPDATE experts SET
				reasoning_charter = CASE
					WHEN COALESCE(reasoning_charter, '') = '' THEN $1
					ELSE reasoning_charter || E'\n\n--- Additional Principles (from new transcript) ---\n\n' || $1
				END,
				clarification_charter = $2,
				updated_at = NOW()
			 WHERE id = $3`,
			charter.ReasoningCharter,
			string(clarificationJSON),
			expertID,
		)
	}
	if err != nil {
		p.logger.Warn("failed to update charters", zap.Error(err))
	}

	// Step 8: Build capability table
	ptimer.Start("capability_build")
	capabilities, err := p.capability.Build(ctx, chunks, topicResults)
	if err != nil {
		p.logger.Warn("capability build failed (LLM), falling back to chunk-based capabilities",
			zap.Error(err),
		)
		// Fallback: populate capabilities directly from course_chunks topic data.
		// WHY: Chunks are stored with topics (Step 2). Capability table is just
		// an aggregation. We can compute it without LLM — less detailed but correct.
		p.buildCapabilitiesFromChunks(ctx, expertID)
	} else {
		p.storeCapabilities(ctx, expertID, capabilities)
	}
	ptimer.Stop("capability_build")

	// Step 9: Update expert stats.
	// BUG FIX (2026-09-08): previously used len(chunks)/uniqueTopics
	// directly, which are ONLY the current transcript file's counts.
	// In append mode (the DEFAULT per DOMAIN_EXPERT_COLLABORATION_DESIGN.md
	// §5.4), course_chunks keeps BOTH the old and new chunks, but this
	// UPDATE overwrote experts.total_chunks with just the new file's
	// count (e.g. 420), silently discarding the previous total (e.g.
	// 745) instead of reflecting the true 1165 rows now in course_chunks.
	// Fix: query the actual aggregate from course_chunks/expert_capabilities
	// (the real source of truth this column is supposed to mirror) instead
	// of trusting the in-memory count from only this run. Falls back to
	// the old (wrong but non-fatal) values on query error rather than
	// leaving total_chunks unset entirely.
	uniqueTopics := countUniqueTopics(topicResults)
	avgDepth := calculateAvgDepth(capabilities)

	actualTotalChunks := len(chunks)
	actualTotalTopics := uniqueTopics
	if err := p.db.QueryRow(ctx,
		`SELECT COUNT(*), COUNT(DISTINCT topic)
		 FROM course_chunks
		 WHERE expert_id = $1`,
		expertID,
	).Scan(&actualTotalChunks, &actualTotalTopics); err != nil {
		p.logger.Warn("failed to compute actual total_chunks/total_topics from DB, falling back to this run's counts",
			zap.String("expert_id", expertID.String()),
			zap.Error(err),
		)
		actualTotalChunks = len(chunks)
		actualTotalTopics = uniqueTopics
	}

	_, err = p.db.Exec(ctx,
		`UPDATE experts SET
			total_chunks = $1,
			total_topics = $2,
			avg_depth_level = $3,
			is_training = FALSE,
			updated_at = NOW()
		 WHERE id = $4`,
		actualTotalChunks, actualTotalTopics, avgDepth, expertID,
	)
	if err != nil {
		p.logger.Warn("failed to update expert stats", zap.Error(err))
	}

	// Step 10: Smoke test — verify the expert is actually retrievable.
	// WHY here: chunks + capabilities are in DB, so retrieval is possible.
	// WHY before marking complete: training_status must reflect real state.
	ptimer.Start("smoke_test")
	smokeTestPassed, smokePassCount, smokeErr := p.runSmokeTest(ctx, expertID, expertName)
	ptimer.Stop("smoke_test")
	if smokeErr != nil {
		p.logger.Warn("smoke test error (non-fatal, expert stays draft)",
			zap.String("expert_id", expertID.String()),
			zap.Error(smokeErr),
		)
	}

	if smokeTestPassed {
		// Mark expert as trained and publicly visible.
		_, err = p.db.Exec(ctx,
			`UPDATE experts SET
				training_status = 'trained',
				is_training = FALSE,
				updated_at = NOW()
			 WHERE id = $1`,
			expertID,
		)
		if err != nil {
			p.logger.Warn("failed to set training_status=trained", zap.Error(err))
		}
		p.logger.Info("smoke test PASSED — expert is trained",
			zap.String("expert_id", expertID.String()),
			zap.Int("probes_passed", smokePassCount),
		)
	} else {
		// Keep expert in draft — ingestion succeeded but retrieval is weak.
		// Admin should upload more transcripts and re-ingest.
		_, err = p.db.Exec(ctx,
			`UPDATE experts SET
				training_status = 'draft',
				is_training = FALSE,
				updated_at = NOW()
			 WHERE id = $1`,
			expertID,
		)
		if err != nil {
			p.logger.Warn("failed to reset training_status=draft", zap.Error(err))
		}
		p.logger.Warn("smoke test FAILED — expert stays in draft, upload more transcripts",
			zap.String("expert_id", expertID.String()),
			zap.Int("probes_passed", smokePassCount),
		)
	}

	duration := time.Since(start).Milliseconds()
	p.updateJobStatus(ctx, jobID, "complete", "", len(chunks), len(chunks))

	p.logger.Info("ingestion complete",
		zap.String("expert_id", expertID.String()),
		zap.Int("chunks", len(chunks)),
		zap.Int("topics", uniqueTopics),
		zap.Int64("duration_ms", duration),
		zap.Bool("smoke_test_passed", smokeTestPassed),
	)

	// Log per-step phase breakdown for ingestion observability.
	// Admin can see which step is the bottleneck (embed is typically 60-80%).
	for _, phase := range ptimer.Results() {
		p.logger.Info("ingestion phase timing",
			zap.String("expert_id", expertID.String()),
			zap.String("phase", phase.Phase),
			zap.Int64("ms", phase.Duration.Milliseconds()),
		)
	}

	return &IngestionResult{
		ExpertID:      expertID,
		TotalChunks:   len(chunks),
		TotalTopics:   uniqueTopics,
		AvgDepthLevel: avgDepth,
		DurationMs:    duration,
	}, nil
}

// storeChunks inserts all chunks into the database.
// Links prev/next chunk IDs for context navigation.
//
// Dedup: uses ON CONFLICT (expert_id, chunk_hash) DO NOTHING. If a chunk
// with the same hash already exists for this expert, INSERT is a no-op
// and RETURNING id returns no rows (pgx.QueryRow -> pgx.ErrNoRows).
// In that case, look up the existing chunk's id and use it — this keeps
// prev/next linking intact even when partially-duplicate transcripts are
// re-ingested. In practice, this is rare in replace mode (table was just
// wiped) and common in append mode (that's the whole point).
//
// Mental execution:
// Insert chunk 0 (new) -> get new ID
// Insert chunk 1 (dup) -> ON CONFLICT, no rows returned -> lookup existing ID
// Update chunk 0 -> set next_chunk_id = chunk 1's existing ID
// ... repeat for all chunks
func (p *IngestionPipeline) storeChunks(
	ctx context.Context,
	jobID uuid.UUID,
	expertID uuid.UUID,
	chunks []TextChunk,
	topics []TopicResult,
	embeddings [][]float32,
	sourceFile string,
	) ([]uuid.UUID, error) {
	chunkIDs := make([]uuid.UUID, len(chunks))

	// C6: stamp each new chunk with the provider/model that produced its
	// vector, so a later embedding-provider/model change is detectable as a
	// re-embed requirement. Read once per ingestion (cheap). Existing
	// duplicate chunks keep their original stamp (ON CONFLICT DO NOTHING).
	embedProvider, embedModel := p.currentEmbedding(ctx)
	var embedModelArg interface{} // empty model → SQL NULL (sidecar has no model name)
	if embedModel != "" {
		embedModelArg = embedModel
	}

	for i, chunk := range chunks {
		topic := topics[i].Topic
		subtopic := ""
		if i < len(topics) {
			subtopic = topics[i].Subtopic
		}

		embedding := pgvector.NewVector(embeddings[i])

		// Safety: chunker should always populate ChunkHash, but if a caller
		// bypassed the chunker and constructed TextChunk directly, compute it
		// here so the dedup key is never NULL.
		if chunk.ChunkHash == "" {
			chunk.ChunkHash = HashChunkText(chunk.Text)
		}

		var chunkID uuid.UUID
		err := p.db.QueryRow(ctx,
			`INSERT INTO course_chunks
				(expert_id, chunk_text, chunk_index, topic, subtopic, source_file, embedding, chunk_hash,
				 embedding_provider, embedding_model)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			 ON CONFLICT (expert_id, chunk_hash) DO NOTHING
			 RETURNING id`,
			expertID, chunk.Text, chunk.Index, topic, subtopic, sourceFile, embedding, chunk.ChunkHash,
			embedProvider, embedModelArg,
		).Scan(&chunkID)
		if err != nil {
			// ON CONFLICT DO NOTHING + RETURNING id yields zero rows on conflict,
			// which pgx surfaces as pgx.ErrNoRows. Treat that as "duplicate":
			// look up the existing chunk id and continue.
			if err.Error() == "no rows in result set" {
				lookupErr := p.db.QueryRow(ctx,
					`SELECT id FROM course_chunks
					 WHERE expert_id = $1 AND chunk_hash = $2
					 LIMIT 1`,
					expertID, chunk.ChunkHash,
				).Scan(&chunkID)
				if lookupErr != nil {
					return nil, fmt.Errorf("dedup lookup failed for chunk %d: %w", i, lookupErr)
				}
				p.logger.Debug("chunk deduplicated",
					zap.Int("index", i),
					zap.String("chunk_hash", chunk.ChunkHash),
				)
			} else {
				return nil, fmt.Errorf("failed to insert chunk %d: %w", i, err)
			}
		}

		chunkIDs[i] = chunkID

		// Link to previous chunk
		if i > 0 {
			// Update current chunk's prev_chunk_id
			_, err = p.db.Exec(ctx,
				`UPDATE course_chunks SET prev_chunk_id = $1 WHERE id = $2`,
				chunkIDs[i-1], chunkID,
			)
			if err != nil {
				p.logger.Warn("failed to link prev chunk", zap.Error(err))
			}

			// Update previous chunk's next_chunk_id
			_, err = p.db.Exec(ctx,
				`UPDATE course_chunks SET next_chunk_id = $1 WHERE id = $2`,
				chunkID, chunkIDs[i-1],
			)
			if err != nil {
				p.logger.Warn("failed to link next chunk", zap.Error(err))
			}
		}

		// Update job progress every 50 chunks
		if i%50 == 0 {
			p.updateJobProgress(ctx, jobID, i+1, len(chunks))
		}
	}

	return chunkIDs, nil
}

// buildCapabilitiesFromChunks populates expert_capabilities from course_chunks
// topic data without LLM analysis. Used as fallback when capability build fails.
//
// Mental execution:
//   745 chunks, 37 topics → 37 capability rows
//   Each row: topic, depth_level=1, chunk_count=N, complexity_ceiling="basic"
func (p *IngestionPipeline) buildCapabilitiesFromChunks(ctx context.Context, expertID uuid.UUID) {
	rows, err := p.db.Query(ctx,
		`SELECT topic, COUNT(*) as chunk_count
		 FROM course_chunks
		 WHERE expert_id = $1
		   AND topic IS NOT NULL
		   AND topic != ''
		 GROUP BY topic
		 ORDER BY chunk_count DESC`,
		expertID,
	)
	if err != nil {
		p.logger.Warn("buildCapabilitiesFromChunks: query failed", zap.Error(err))
		return
	}
	defer rows.Close()

	inserted := 0
	for rows.Next() {
		var topic string
		var chunkCount int
		if err := rows.Scan(&topic, &chunkCount); err != nil {
			continue
		}
		_, err := p.db.Exec(ctx,
			`INSERT INTO expert_capabilities
				(expert_id, topic, depth_level, chunk_count, complexity_ceiling)
			 VALUES ($1, $2, 1, $3, 'basic')
			 ON CONFLICT (expert_id, topic) DO UPDATE SET
				chunk_count = EXCLUDED.chunk_count,
				updated_at  = NOW()`,
			expertID, topic, chunkCount,
		)
		if err != nil {
			p.logger.Warn("buildCapabilitiesFromChunks: insert failed",
				zap.String("topic", topic), zap.Error(err))
			continue
		}
		inserted++
	}
	p.logger.Info("buildCapabilitiesFromChunks: done",
		zap.String("expert_id", expertID.String()),
		zap.Int("topics_inserted", inserted),
	)
}

// storeCapabilities inserts capability records for an expert.
func (p *IngestionPipeline) storeCapabilities(ctx context.Context, expertID uuid.UUID, capabilities []CapabilityResult) {
	// Delete existing capabilities
	_, err := p.db.Exec(ctx,
		`DELETE FROM expert_capabilities WHERE expert_id = $1`,
		expertID,
	)
	if err != nil {
		p.logger.Warn("failed to delete old capabilities", zap.Error(err))
	}

	for _, cap := range capabilities {
		// BUG FIX (2026-09-08): can_handle / cannot_handle / example_questions
		// are TEXT[] columns (Postgres native array), NOT JSONB — verified
		// against migrations/001_initial_schema.up.sql's expert_capabilities
		// table definition. The previous code called json.Marshal() on each
		// []string and sent the resulting JSON string (e.g. `["a","b"]`) as
		// the value for a TEXT[] column. Postgres rejected this with
		// SQLSTATE 22P02 "malformed array literal" because a JSON string is
		// not valid Postgres array literal syntax (`{"a","b"}` would be).
		// Fix: pass the []string slices directly. pgx/v5 (used throughout
		// this codebase via pgxpool.Pool) natively encodes Go []string as
		// a Postgres text[] parameter — no json.Marshal, no manual array
		// literal construction needed. Nil-safety: default nil slices to
		// an empty slice (not NULL) so admin UI / GetTopics callers always
		// get a real (possibly empty) array back, never have to special-
		// case NULL vs [].
		canHandle := cap.CanHandle
		if canHandle == nil {
			canHandle = []string{}
		}
		cannotHandle := cap.CannotHandle
		if cannotHandle == nil {
			cannotHandle = []string{}
		}
		exampleQuestions := cap.ExampleQuestions
		if exampleQuestions == nil {
			exampleQuestions = []string{}
		}

		_, err := p.db.Exec(ctx,
			`INSERT INTO expert_capabilities
				(expert_id, topic, depth_level, chunk_count, complexity_ceiling,
				 can_handle, cannot_handle, example_questions)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			 ON CONFLICT (expert_id, topic) DO UPDATE SET
				depth_level = EXCLUDED.depth_level,
				chunk_count = EXCLUDED.chunk_count,
				complexity_ceiling = EXCLUDED.complexity_ceiling,
				can_handle = EXCLUDED.can_handle,
				cannot_handle = EXCLUDED.cannot_handle,
				example_questions = EXCLUDED.example_questions,
				updated_at = NOW()`,
			expertID, cap.Topic, cap.DepthLevel, cap.ChunkCount, cap.ComplexityCeiling,
			canHandle, cannotHandle, exampleQuestions,
		)
		if err != nil {
			p.logger.Warn("failed to store capability",
				zap.String("topic", cap.Topic),
				zap.Error(err),
			)
		}
	}
}

// updateStage writes the current pipeline stage to the DB.
// Called at every stage transition for live UI updates.
func (p *IngestionPipeline) updateStage(ctx context.Context, jobID uuid.UUID, stage, detail string) {
	_, err := p.db.Exec(ctx,
		`UPDATE ingestion_jobs SET current_stage=$1, stage_detail=$2 WHERE id=$3`,
		stage, detail, jobID,
	)
	if err != nil {
		p.logger.Warn("updateStage failed", zap.String("stage", stage), zap.Error(err))
	}
}

// pauseOnLLMFailure transitions the job to status='paused' when the
// charter LLM call fails. Writes a checkpoint so resume/retry can
// restart from charter_extraction without re-chunking or re-embedding.
//
// Returns ErrJobPaused always — caller must return this to the goroutine
// so IngestTranscript stops cleanly.
//
// Mental execution:
//   Input: jobID, reason="charter LLM: context deadline exceeded"
//   1. Write checkpoint at StagePaused (chunks + topics already done)
//   2. UPDATE status='paused', paused_at=NOW(), error_message=reason
//   3. Return ErrJobPaused
//   Caller (IngestTranscript): return nil, ErrJobPaused
//   Goroutine (admin_handler): errors.Is(err, ErrJobPaused) → log info, no retry
func (p *IngestionPipeline) pauseOnLLMFailure(
	ctx context.Context,
	jobID uuid.UUID,
	cpWriter *CheckpointWriter,
	chunksTotal int,
	costSoFar float64,
	start time.Time,
	reason string,
) error {
	// Write checkpoint at paused stage so resume/retry starts from
	// charter_extraction, not from the beginning.
	// CharterExtracted=false: charter was NOT successfully extracted —
	// that is exactly why we are pausing. Resume must re-run it.
	cpWriter.Write(ctx, JobCheckpoint{
		Stage:            StagePaused,
		ChunksDone:       chunksTotal,
		ChunksTotal:      chunksTotal,
		CharterExtracted: false,
		LastBatchIndex:   0,
		CostUSDSoFar:     costSoFar,
		StartedAt:        start,
	})

	// Transition job to paused state.
	// paused_at is used by the 24h auto-fail checker in main.go.
	//
	// WHY fallback strategy:
	//   If the context is already cancelled (deadline exceeded during LLM retry),
	//   the DB update will fail immediately. We need a fallback with a fresh
	//   context to ensure the job status is updated. Otherwise the job stays
	//   'running' forever with no way to recover.
	updateQuery := `UPDATE ingestion_jobs SET
		status        = 'paused',
		current_stage = 'paused',
		stage_detail  = $1,
		paused_at     = NOW(),
		error_message = $2
	 WHERE id = $3`

	_, err := p.db.Exec(ctx, updateQuery,
		StageLabels[StagePaused],
		reason,
		jobID,
	)
	if err != nil {
		p.logger.Error("pauseOnLLMFailure: primary DB update failed, trying fallback",
			zap.String("job_id", jobID.String()),
			zap.Error(err),
		)

		// Fallback: try with a fresh context (5s timeout).
		// If the original context was cancelled, this gives us one last
		// chance to update the DB before giving up.
		fallbackCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, fallbackErr := p.db.Exec(fallbackCtx, updateQuery,
			StageLabels[StagePaused],
			reason,
			jobID,
		)
		if fallbackErr != nil {
			// Both attempts failed. This is serious — DB is likely down or
			// the connection is broken. Mark the job as 'failed' instead of
			// 'paused' so admin knows something is wrong.
			p.logger.Error("pauseOnLLMFailure: fallback DB update also failed — marking job as failed",
				zap.String("job_id", jobID.String()),
				zap.Error(fallbackErr),
			)

			// Last-ditch attempt: mark as 'failed' with a fresh context.
			// If this also fails, there's nothing more we can do.
			failCtx, failCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer failCancel()
			_, _ = p.db.Exec(failCtx,
				`UPDATE ingestion_jobs SET
					status        = 'failed',
					current_stage = 'failed',
					stage_detail  = 'DB update failed',
					error_message = $1,
					completed_at  = NOW()
				 WHERE id = $2`,
				fmt.Sprintf("pauseOnLLMFailure: DB update failed twice: %s", fallbackErr.Error()),
				jobID,
			)
			// Return the original error, not ErrJobPaused, because the job
			// is not actually paused — it's failed.
			return fmt.Errorf("pauseOnLLMFailure: failed to update job status: %w", fallbackErr)
		}

		p.logger.Info("pauseOnLLMFailure: fallback DB update succeeded",
			zap.String("job_id", jobID.String()),
		)
	}

	p.logger.Warn("ingestion paused: charter LLM failure — awaiting admin action",
		zap.String("job_id", jobID.String()),
		zap.String("reason", reason),
	)

	return ErrJobPaused
}

// loadChunksFromDB loads existing TextChunks for a specific source file.
// Scoped to (expert_id, source_file) to prevent cross-file chunk leakage
// when an expert has multiple uploaded transcripts.
//
// WHY source_file filter:
//   Without it, resume loads ALL chunks for the expert across every
//   transcript file ever uploaded. For an expert with 2 files
//   (868 + 873 chunks), resume would load 1741 chunks instead of 873,
//   causing topic extraction and embedding to re-process the wrong set.
func (p *IngestionPipeline) loadChunksFromDB(ctx context.Context, expertID uuid.UUID, sourceFile string) ([]TextChunk, error) {
	rows, err := p.db.Query(ctx,
		`SELECT chunk_text, chunk_index, COALESCE(chunk_hash,'') FROM course_chunks
		 WHERE expert_id=$1 AND source_file=$2 ORDER BY chunk_index ASC`,
		expertID, sourceFile,
	)
	if err != nil {
		return nil, fmt.Errorf("loadChunksFromDB: %w", err)
	}
	defer rows.Close()

	var chunks []TextChunk
	for rows.Next() {
		var c TextChunk
		if err := rows.Scan(&c.Text, &c.Index, &c.ChunkHash); err != nil {
			continue
		}
		chunks = append(chunks, c)
	}
	return chunks, rows.Err()
}

// loadCharterFromDB loads the existing charter from the experts table.
// Used on resume to skip re-extraction.
func (p *IngestionPipeline) loadCharterFromDB(ctx context.Context, expertID uuid.UUID) (*Charter, error) {
	var reasoningCharter string
	var clarificationJSON []byte
	err := p.db.QueryRow(ctx,
		`SELECT COALESCE(reasoning_charter,''), COALESCE(clarification_charter,'{}') FROM experts WHERE id=$1`,
		expertID,
	).Scan(&reasoningCharter, &clarificationJSON)
	if err != nil {
		return nil, fmt.Errorf("loadCharterFromDB: %w", err)
	}
	var clarification map[string][]string
	if err := json.Unmarshal(clarificationJSON, &clarification); err != nil {
		clarification = defaultClarificationCharter()
	}
	return &Charter{
		ReasoningCharter:     reasoningCharter,
		ClarificationCharter: clarification,
	}, nil
}

// updateJobStatus updates the ingestion job status.
func (p *IngestionPipeline) updateJobStatus(
	ctx context.Context,
	jobID uuid.UUID,
	status string,
	errorMsg string,
	processed int,
	total int,
) {
	var completedAt interface{}
	if status == "complete" || status == "failed" {
		completedAt = time.Now()
	}

	// BUG FIX (2026-09-08): PostgreSQL rejected this query with
	// SQLSTATE 42P08 "inconsistent types deduced for parameter $1".
	// Root cause: $1 was used twice — once as a bare positional param
	// in `status = $1` (implicit type from the status column) and once
	// explicitly cast as `$1::text` inside the CASE WHEN. The planner
	// could not reconcile the implicit type from the first usage with
	// the explicit cast in the second usage, so it rejected the whole
	// query. Because this UPDATE never ran, the job row was permanently
	// stuck at whatever status/stage it had before this call (e.g.
	// "pending" / "57% embedding"), even though the pipeline had
	// actually finished successfully and training_status was already
	// 'trained' on the experts row — the frontend's ingestion modal
	// reads job status from ingestion_jobs, not from experts, so it kept
	// showing stale progress forever.
	// Fix: cast $1 the SAME way (::varchar, matching status's actual
	// column type per migrations/001_initial_schema.up.sql) in BOTH
	// usages, so the planner sees one consistent type throughout.
	_, err := p.db.Exec(ctx,
		`UPDATE ingestion_jobs SET
			status = $1::varchar,
			error_message = NULLIF($2, ''),
			processed_chunks = $3,
			total_chunks = $4,
			completed_at = $5,
			started_at = CASE WHEN $1::varchar = 'running' THEN NOW() ELSE started_at END,
			current_stage = CASE WHEN $1::varchar = 'complete' THEN 'complete'
			                     WHEN $1::varchar = 'failed'   THEN current_stage
			                     ELSE current_stage END
		 WHERE id = $6`,
		status, errorMsg, processed, total, completedAt, jobID,
	)
	if err != nil {
		p.logger.Warn("failed to update job status",
			zap.String("job_id", jobID.String()),
			zap.Error(err),
		)
	}
}

// ============================================================
// SMOKE TEST
// ============================================================

// Smoke test constants.
// WHY named constants not magic numbers: KNOWLEDGE_HUB.md §7.4 anti-pattern checklist.
const (
	// smokeTestProbeCount is the number of topics to probe.
	// 5 gives a representative sample without being expensive.
	smokeTestProbeCount = 5

	// smokeTestPassThreshold is the minimum number of probes that must
	// return a chunk with rerank score >= smokeTestRerankThreshold.
	// Majority (3/5) allows lightly-covered topics to fail without
	// blocking the expert from being marked trained.
	smokeTestPassThreshold = 3

	// smokeTestRerankThreshold is the minimum rerank score for a probe
	// to be considered "passing". Matches China Wall Layer 1 threshold.
	smokeTestRerankThreshold = float32(0.35)

	// smokeTestTopK is the number of vector-search candidates to fetch
	// per probe before reranking. 10 gives the reranker enough to work with.
	smokeTestTopK = 10
)

// runSmokeTest verifies that the expert's corpus is retrievable.
//
// Algorithm:
// 1. Load top-N topics from expert_capabilities (by depth_level DESC).
// 2. For each topic, build a generic probe question.
// 3. Embed the probe question.
// 4. Vector-search course_chunks for this expert (top smokeTestTopK).
// 5. Rerank candidates against the probe.
// 6. Pass if best rerank score >= smokeTestRerankThreshold.
// 7. Return (passed, passCount, err).
//    passed = passCount >= smokeTestPassThreshold.
//
// Mental execution:
// Input: expertID="abc", expertName="Arpit — System Design"
// Load topics: ["system_design", "databases", "caching", "load_balancing", "microservices"]
// Probe 0: "What does Arpit — System Design teach about system_design?"
//   embed -> [0.1, 0.3, ...] (768D)
//   vector search -> 10 chunks
//   rerank -> best score 0.72 -> PASS
// Probe 1: "What does Arpit — System Design teach about databases?"
//   ... best score 0.41 -> PASS
// ... (5 probes total)
// passCount=4 >= 3 -> passed=true
func (p *IngestionPipeline) runSmokeTest(
	ctx context.Context,
	expertID uuid.UUID,
	expertName string,
) (passed bool, passCount int, err error) {

	p.logger.Info("smoke test starting",
		zap.String("expert_id", expertID.String()),
		zap.String("expert_name", expertName),
	)

	// Step 1: Load top topics from expert_capabilities.
	// WHY expert_capabilities not course_chunks:
	//   capabilities are already aggregated by topic with depth_level.
	//   Picking by depth_level DESC gives us the best-covered topics first.
	rows, err := p.db.Query(ctx,
		`SELECT topic FROM expert_capabilities
		 WHERE expert_id = $1
		 ORDER BY depth_level DESC, chunk_count DESC
		 LIMIT $2`,
		expertID, smokeTestProbeCount,
	)
	if err != nil {
		return false, 0, fmt.Errorf("smoke test: failed to load topics: %w", err)
	}
	defer rows.Close()

	var topics []string
	for rows.Next() {
		var topic string
		if scanErr := rows.Scan(&topic); scanErr == nil {
			topics = append(topics, topic)
		}
	}
	if err = rows.Err(); err != nil {
		return false, 0, fmt.Errorf("smoke test: topic scan error: %w", err)
	}

	// If expert_capabilities is empty (capability build failed), load
	// distinct topics directly from course_chunks.
	// WHY: Topic extraction (Step 2) runs before capability build (Step 8).
	// Chunk topics are stored even when capability build fails.
	// This gives real topic probes instead of a weak "general" fallback.
	if len(topics) == 0 {
		p.logger.Warn("smoke test: no capabilities found, loading topics from course_chunks",
			zap.String("expert_id", expertID.String()),
		)
		chunkTopicRows, chunkTopicErr := p.db.Query(ctx,
			`SELECT DISTINCT topic
			 FROM course_chunks
			 WHERE expert_id = $1
			   AND topic IS NOT NULL
			   AND topic != ''
			   AND topic != 'general'
			 ORDER BY topic
			 LIMIT $2`,
			expertID, smokeTestProbeCount,
		)
		if chunkTopicErr == nil {
			defer chunkTopicRows.Close()
			for chunkTopicRows.Next() {
				var t string
				if scanErr := chunkTopicRows.Scan(&t); scanErr == nil {
					topics = append(topics, t)
				}
			}
		}
		if len(topics) > 0 {
			p.logger.Info("smoke test: loaded topics from course_chunks",
				zap.Int("count", len(topics)),
				zap.Strings("topics", topics),
			)
		} else {
			// Last resort: generic probe
			p.logger.Warn("smoke test: no topics in chunks either, using generic probe",
				zap.String("expert_id", expertID.String()),
			)
			topics = []string{"general"}
		}
	}

	// Step 2–6: Probe each topic.
	passCount = 0
	for i, topic := range topics {
		probeQuestion := fmt.Sprintf(
			"What does %s teach about %s?",
			expertName, topic,
		)

		// Step 3: Embed the probe question.
		embeddings, embedErr := p.embedder.Embed(ctx, []string{probeQuestion})
		if embedErr != nil {
			p.logger.Warn("smoke test: embed failed for probe",
				zap.Int("probe_index", i),
				zap.String("topic", topic),
				zap.Error(embedErr),
			)
			// ML sidecar failure is infrastructure, not corpus quality.
			// Return error so caller can decide (non-fatal).
			return false, passCount, fmt.Errorf("smoke test: ML sidecar unavailable: %w", embedErr)
		}
		queryVec := pgvector.NewVector(embeddings[0])

		// Step 4: Vector search — fetch top-K candidates for this expert.
		chunkRows, searchErr := p.db.Query(ctx,
			`SELECT chunk_text
			 FROM course_chunks
			 WHERE expert_id = $1
			 ORDER BY embedding <=> $2
			 LIMIT $3`,
			expertID, queryVec, smokeTestTopK,
		)
		if searchErr != nil {
			p.logger.Warn("smoke test: vector search failed",
				zap.Int("probe_index", i),
				zap.Error(searchErr),
			)
			continue // Skip this probe, don't fail the whole test
		}

		var candidates []string
		for chunkRows.Next() {
			var text string
			if scanErr := chunkRows.Scan(&text); scanErr == nil {
				candidates = append(candidates, text)
			}
		}
		chunkRows.Close()

		if len(candidates) == 0 {
			p.logger.Warn("smoke test: no candidates returned",
				zap.Int("probe_index", i),
				zap.String("topic", topic),
			)
			continue
		}

		// Step 5: Rerank candidates against the probe question.
		// WHY p.sidecar.Rerank (not p.embedder.Rerank):
		// Rerank() is sidecar-only — it is NOT in the Embedder interface.
		// CodeCraftAPI has no /v1/rerank endpoint. Locked decision.
		rankResults, rerankErr := p.sidecar.Rerank(ctx, probeQuestion, candidates, len(candidates))
		if rerankErr != nil {
			p.logger.Warn("smoke test: rerank failed",
				zap.Int("probe_index", i),
				zap.Error(rerankErr),
			)
			// Rerank failure = ML sidecar issue, return error.
			return false, passCount, fmt.Errorf("smoke test: rerank unavailable: %w", rerankErr)
		}

		// Step 6: Check if best rerank score meets threshold.
		bestScore := float32(0)
		for _, r := range rankResults {
			if r.Score > bestScore {
				bestScore = r.Score
			}
		}

		probePassed := bestScore >= smokeTestRerankThreshold
		if probePassed {
			passCount++
		}

		p.logger.Debug("smoke test probe result",
			zap.Int("probe_index", i),
			zap.String("topic", topic),
			zap.Float32("best_score", bestScore),
			zap.Bool("passed", probePassed),
		)
	}

	// Dynamic threshold: min(smokeTestPassThreshold, len(topics))
	// WHY dynamic:
	//   If expert_capabilities is empty (capability build failed), fallback
	//   creates 1 generic probe. 1 >= 3 is always false — mathematically
	//   impossible to pass even with perfect retrieval.
	//   Dynamic threshold: 1 probe → threshold=1, 5 probes → threshold=3.
	effectiveThreshold := smokeTestPassThreshold
	if len(topics) < smokeTestPassThreshold {
		effectiveThreshold = len(topics)
	}
	if effectiveThreshold == 0 {
		// No probes ran at all — no data to verify. Stay draft.
		p.logger.Warn("smoke test: no probes ran, expert stays draft",
			zap.String("expert_id", expertID.String()),
		)
		return false, 0, nil
	}

	passed = passCount >= effectiveThreshold

	p.logger.Info("smoke test complete",
		zap.String("expert_id", expertID.String()),
		zap.Int("probes_run", len(topics)),
		zap.Int("probes_passed", passCount),
		zap.Int("effective_threshold", effectiveThreshold),
		zap.Bool("passed", passed),
	)

	return passed, passCount, nil
}

// isFatalLLMError returns true when the error indicates a permanent
// provider-level failure that will not resolve by retrying.
//
// Fatal errors (pause job):
//   - 402 Payment Required: API credits exhausted
//   - 401 Unauthorized:     API key invalid or revoked
//   - 403 Forbidden:        Account suspended or access denied
//
// Non-fatal errors (use fallback, continue):
//   - 429 Too Many Requests: rate limit, transient
//   - 5xx Server Error:      provider outage, transient
//   - network timeout:       transient
//
// WHY this distinction matters for topic extraction:
//   A fatal error on batch 3 of 50 means batches 4-50 will also fail.
//   Continuing stores all remaining chunks with topic="general",
//   destroying topic diversity. Gate 2 then fails for every question
//   because no relevant topic-specific chunk exists.
//   Pausing preserves the work done so far and lets admin fix the
//   API key/credits before resuming from the last checkpoint.
func isFatalLLMError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	// Match the exact format from providers/common.go:
	// fmt.Errorf("provider returned status %d", resp.StatusCode)
	return strings.Contains(msg, "status 402") ||
		strings.Contains(msg, "status 401") ||
		strings.Contains(msg, "status 403")
}

// updateJobProgress updates processed chunk count.
func (p *IngestionPipeline) updateJobProgress(ctx context.Context, jobID uuid.UUID, processed, total int) {
	_, _ = p.db.Exec(ctx,
		`UPDATE ingestion_jobs SET processed_chunks = $1, total_chunks = $2 WHERE id = $3`,
		processed, total, jobID,
	)
}

// countUniqueTopics counts distinct topics in results.
func countUniqueTopics(topics []TopicResult) int {
	seen := make(map[string]bool)
	for _, t := range topics {
		if t.Topic != "" {
			seen[t.Topic] = true
		}
	}
	return len(seen)
}

// calculateAvgDepth calculates average depth level across capabilities.
func calculateAvgDepth(capabilities []CapabilityResult) float64 {
	if len(capabilities) == 0 {
		return 0
	}
	total := 0
	for _, c := range capabilities {
		total += c.DepthLevel
	}
	return float64(total) / float64(len(capabilities))
}

// parseClarificationJSON is used by charter_extractor.
func parseClarificationJSON(response string) (map[string][]string, error) {
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")
	if start == -1 || end == -1 || start >= end {
		return nil, fmt.Errorf("no JSON object found")
	}
	response = response[start : end+1]

	var result map[string][]string
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, err
	}

	normalized := make(map[string][]string)
	for topic, questions := range result {
		normalizedTopic := normalizeTopicName(topic)
		if len(questions) > 5 {
			questions = questions[:5]
		}
		normalized[normalizedTopic] = questions
	}
	return normalized, nil
}
