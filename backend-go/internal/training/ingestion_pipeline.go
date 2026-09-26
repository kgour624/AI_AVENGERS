package training

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/docextract"
	"ai_avengers/backend/internal/gateway"
	"ai_avengers/backend/internal/jobevents"
	"ai_avengers/backend/internal/ml"
	"ai_avengers/backend/internal/observability"
)

// defaultIngestionWorkers bounds how many topic/embed batches run at once (T3).
//
// WHY 3: it mirrors orchestrator.expertMaxConcurrency (3). The bottleneck in
// both parallel stages is external — the LLM provider's rate limit for topic
// extraction, the ML sidecar's CPU for embeddings — not this process's CPU.
// Adding workers past the provider's capacity only moves the queue to the
// provider; 3 keeps chunks flowing without tripping rate limits or spawning a
// goroutine per batch on a 15k-chunk transcript.
//
// Overridable with INGESTION_WORKERS (bounded 1..32) so a beefier sidecar or a
// higher-rate provider key can be exploited without a code change.
const defaultIngestionWorkers = 3

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
	// events is the durable job timeline (T1). Nil-safe: when unwired, emits
	// are dropped and ingestion behaves exactly as before.
	events  *jobevents.Store
	// extractor converts an uploaded document into text (stage 0). Nil is a
	// valid value: it disables the office/PDF formats while plain text keeps
	// working — see docextract.Extract.
	extractor *docextract.Extractor
	// workers caps concurrent topic/embed batches (T3).
	workers int
	// capabilityMeasurer runs the measurement pass the ingest gate requires (I3).
	// Nil-safe: when unwired the gate reports that capability was never measured and
	// the expert stays in draft, which is the correct outcome for an environment
	// that cannot measure.
	capabilityMeasurer CapabilityMeasurer
	logger             *zap.Logger
}

// NewIngestionPipeline creates a new ingestion pipeline.
//
// embedder satisfies ml.Embedder — either *ml.SidecarClient (default) or
// *ml.DynamicEmbedder (when CodeCraftAPI embeddings are enabled).
//
// sidecar is kept as a separate *ml.SidecarClient because the smoke test
// step calls Rerank(), which is sidecar-only and NOT part of the Embedder
// interface (locked decision: reranking always stays on the Python sidecar).
//
// events is the durable job event log (T1). Nil is a valid value (events
// unwired / table not yet migrated) — every emit site is nil-safe, so
// ingestion never depends on the timeline being wired.
//
// extractor converts an uploaded document to text (see PrepareTranscript).
// Nil is a valid value: it disables the PDF/office formats while .txt/.md keep
// working, which is exactly the pre-existing behaviour.
func NewIngestionPipeline(
	db *pgxpool.Pool,
	embedder ml.Embedder,
	sidecar *ml.SidecarClient,
	gw *gateway.ModelGateway,
	events *jobevents.Store,
	extractor *docextract.Extractor,
	logger *zap.Logger,
) *IngestionPipeline {
	workers := defaultIngestionWorkers
	if raw := os.Getenv("INGESTION_WORKERS"); raw != "" {
		if n, parseErr := strconv.Atoi(raw); parseErr == nil && n >= 1 && n <= 32 {
			workers = n
		} else {
			logger.Warn("invalid INGESTION_WORKERS, using default",
				zap.String("value", raw), zap.Int("default", defaultIngestionWorkers))
		}
	}

	return &IngestionPipeline{
		db:         db,
		embedder:   embedder,
		sidecar:    sidecar,
		gateway:    gw,
		chunker:    NewTextChunker(DefaultChunkerConfig()),
		topics:     NewTopicExtractor(gw, embedder, logger),
		charters:   NewCharterExtractor(gw, logger),
		capability: NewCapabilityBuilder(gw, logger),
		events:     events,
		extractor:  extractor,
		workers:    workers,
		logger:     logger,
	}
}

// emit appends one timeline event (T1). Best-effort by design: a broken
// timeline must never abort an ingestion run, so a failure is logged, not
// propagated.
func (p *IngestionPipeline) emit(ctx context.Context, jobID, expertID uuid.UUID, stage, kind string, detail interface{}) {
	if p.events == nil {
		return
	}
	if _, err := p.events.Append(ctx, jobevents.AppendRequest{
		JobID:    jobID,
		ExpertID: expertID,
		Stage:    stage,
		Kind:     kind,
		Detail:   detail,
	}); err != nil {
		p.logger.Warn("ingestion event append failed (non-fatal)",
			zap.String("kind", kind),
			zap.String("stage", stage),
			zap.Error(err),
		)
	}
}

// emitFinal appends a terminal event (complete/failed/paused) using a fresh
// context when the run's context is already cancelled — otherwise a timeout or
// shutdown would erase the one event an admin most needs to see.
func (p *IngestionPipeline) emitFinal(ctx context.Context, jobID, expertID uuid.UUID, stage, kind string, detail interface{}) {
	if p.events == nil {
		return
	}
	if ctx.Err() != nil {
		fresh, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		ctx = fresh
	}
	p.emit(ctx, jobID, expertID, stage, kind, detail)
}

// beginStage records the stage transition (job row) and emits stage_started.
// Returns the start time to hand to endStage.
func (p *IngestionPipeline) beginStage(ctx context.Context, jobID, expertID uuid.UUID, stage, detail string) time.Time {
	p.updateStage(ctx, jobID, stage, detail)
	p.emit(ctx, jobID, expertID, stage, jobevents.KindStageStarted, map[string]interface{}{
		"detail": detail,
	})
	return time.Now()
}

// endStage emits stage_done with the measured duration plus any extra facts
// the caller wants on the timeline.
func (p *IngestionPipeline) endStage(ctx context.Context, jobID, expertID uuid.UUID, stage string, startedAt time.Time, extra map[string]interface{}) {
	detail := map[string]interface{}{
		"duration_ms": time.Since(startedAt).Milliseconds(),
	}
	for k, v := range extra {
		detail[k] = v
	}
	p.emit(ctx, jobID, expertID, stage, jobevents.KindStageDone, detail)
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

	// T1: open the timeline for this run. `resumed` + the checkpoint stage tell
	// the admin at a glance whether this is a fresh run or a continuation.
	resumedFrom := ""
	if isResume && cp != nil {
		resumedFrom = cp.Stage
	}
	p.emit(ctx, jobID, expertID, StagePending, jobevents.KindRunStarted, map[string]interface{}{
		"resumed":          isResume,
		"resumed_from":     resumedFrom,
		"workers":          p.workers,
		"transcript_chars": len(transcript),
		"source_file":      sourceFile,
	})

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
		// NOTE: cleaning is not one of the 6 UI stages, so it does not touch
		// current_stage (which has a CHECK constraint). It is emitted as its
		// own timeline entry so the log still shows where the time went.
		cleanStarted := time.Now()
		p.emit(ctx, jobID, expertID, "cleaning", jobevents.KindStageStarted, map[string]interface{}{
			"label":          "Cleaning transcript",
			"original_chars": originalLen,
		})
		ptimer.Start("clean")
		transcript = cleaner.Clean(transcript)
		ptimer.Stop("clean")
		cleanedLen := len(transcript)
		reductionPct := 0
		if originalLen > 0 {
			reductionPct = (originalLen - cleanedLen) * 100 / originalLen
		}
		p.emit(ctx, jobID, expertID, "cleaning", jobevents.KindStageDone, map[string]interface{}{
			"duration_ms":    time.Since(cleanStarted).Milliseconds(),
			"original_chars": originalLen,
			"cleaned_chars":  cleanedLen,
			"reduction_pct":  reductionPct,
		})
		p.logger.Info("transcript cleaned",
			zap.Int("original_chars", originalLen),
			zap.Int("cleaned_chars", cleanedLen),
			zap.Int("reduction_pct", reductionPct),
		)
		if cleanedLen == 0 {
			err := fmt.Errorf("transcript is empty after cleaning")
			p.updateJobStatus(ctx, jobID, "failed", err.Error(), 0, 0)
			p.emitFinal(ctx, jobID, expertID, "cleaning", jobevents.KindFailed, map[string]interface{}{
				"reason": err.Error(),
			})
			return nil, err
		}
	}

	// ============================================================
	// STEP 1: CHUNKING
	// ============================================================
	// If resuming from topic_extraction or later, load chunks from DB
	// instead of re-chunking (chunking is deterministic but expensive for large transcripts).
	var chunks []TextChunk
	chunkStarted := time.Now()
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
		// Resume path: chunking was already done in an earlier run. Record it
		// so the timeline shows where the stored chunks came from instead of
		// silently skipping step 1.
		p.emit(ctx, jobID, expertID, StageChunking, jobevents.KindStageDone, map[string]interface{}{
			"chunks":          len(chunks),
			"from_checkpoint": true,
		})
	} else {
		chunkStarted = p.beginStage(ctx, jobID, expertID, StageChunking, "Splitting transcript...")
		ptimer.Start("chunk")
		chunks = p.chunker.Chunk(transcript)
		ptimer.Stop("chunk")
		if len(chunks) == 0 {
			err := fmt.Errorf("no chunks created from transcript")
			p.updateJobStatus(ctx, jobID, "failed", err.Error(), 0, 0)
			p.emitFinal(ctx, jobID, expertID, StageChunking, jobevents.KindFailed, map[string]interface{}{
				"reason": err.Error(),
			})
			return nil, err
		}
		p.endStage(ctx, jobID, expertID, StageChunking, chunkStarted, map[string]interface{}{
			"chunks": len(chunks),
		})
	}
	p.logger.Info("chunking complete", zap.Int("chunks", len(chunks)))
	// Publish the denominator immediately so the UI can render "0 / N" before
	// the first batch finishes.
	p.updateJobProgress(ctx, jobID, 0, len(chunks))

	// Init progress tracker and checkpoint writer
	tracker := NewProgressTracker(p.db, jobID, len(chunks), p.logger)
	cpWriter := NewCheckpointWriter(p.db, jobID, p.logger)
	if isResume && cp != nil {
		tracker.costUSD = cp.CostUSDSoFar // restore accumulated cost
	}

	// ============================================================
	// STEP 2: TOPIC EXTRACTION (batched, resumable)
	// ============================================================
	topicResults := make([]TopicResult, len(chunks))
	// Default fallback for all chunks
	for i := range topicResults {
		topicResults[i] = TopicResult{Topic: "general", Confidence: 0.5}
	}

	const topicBatchSize = 50
	resumeTopicBatch := 0

	// Topic results live in memory until step 6 (store), so a checkpoint alone
	// cannot say which batches are safely recoverable:
	//   * cp.Stage == topic_extraction → nothing was stored yet, so a "skip the
	//     first N batches" resume would silently keep the "general" fallback for
	//     those chunks (pre-existing bug: they were never persisted).
	//   * cp.Stage >= charter_extraction → the chunks ARE in the DB, so the
	//     stored topics can be reused. This is what makes Retry cheap (P1).
	// A partial match (fewer stored rows than chunks) falls back to a full
	// recompute rather than storing fallback "general" topics — an all-general
	// corpus makes Gate 2 refuse every future question.
	totalTopicBatches := (len(chunks) + topicBatchSize - 1) / topicBatchSize
	if isResume && cp != nil && StageOrder[cp.Stage] >= StageOrder[StageCharterExtraction] {
		if stored, loadErr := p.loadTopicsFromDB(ctx, expertID, sourceFile, len(chunks)); loadErr == nil && len(stored) == len(chunks) {
			topicResults = stored
			resumeTopicBatch = totalTopicBatches
			p.logger.Info("resume: reused stored topics (no LLM call)", zap.Int("count", len(stored)))
		} else {
			p.logger.Warn("resume: stored topics unavailable or incomplete — recomputing topic extraction",
				zap.Int("expected_chunks", len(chunks)),
				zap.Error(loadErr),
			)
		}
	}

	// Only announce / enter the topic stage when we are actually going to tag
	// chunks. On a resume past this stage, current_stage must NOT be
	// downgraded back to topic_extraction.
	topicStarted := time.Now()
	if resumeTopicBatch < totalTopicBatches {
		topicStarted = p.beginStage(ctx, jobID, expertID, StageTopicExtraction,
			fmt.Sprintf("0/%d chunks tagged", len(chunks)))
	}

	ptimer.Start("topic_extract")

	// T3: parallel topic extraction.
	// WHY safe to parallelize: batches are independent — each reads its own
	// slice of chunks and writes its own slice of topicResults, so there is no
	// shared mutable state. One LLM call per batch IS the stage's latency.
	//
	// WHY the contiguous prefix: batches finish out of order. Resume skips the
	// first N completed batches, which is only correct if N is a PREFIX that is
	// fully done. Anything after a gap is re-done on resume — cheap (one batch
	// of LLM work) and always correct, versus a corrupted resume position.
	//
	// WHY a semaphore: bounds concurrency to p.workers so we stay inside the
	// provider's rate limit and never spawn a goroutine per batch on a
	// 15k-chunk transcript.
	topicCtx, topicCancel := context.WithCancel(ctx)
	defer topicCancel()

	var (
		topicWG          sync.WaitGroup
		topicMu          sync.Mutex
		topicFatalErr    error
		topicCompleted   = make([]bool, totalTopicBatches)
		topicPrefix      = resumeTopicBatch
		topicBatchesDone = resumeTopicBatch
	)
	topicSem := make(chan struct{}, p.workers)

	for batchIdx := resumeTopicBatch; batchIdx < totalTopicBatches; batchIdx++ {
		topicWG.Add(1)
		go func(batchIdx int) {
			defer topicWG.Done()

			// Fail-fast: a fatal provider error (credits/auth) cancels the rest.
			if topicCtx.Err() != nil {
				return
			}
			select {
			case topicSem <- struct{}{}:
			case <-topicCtx.Done():
				return
			}
			defer func() { <-topicSem }()

			batchStart := batchIdx * topicBatchSize
			batchEnd := min(batchStart+topicBatchSize, len(chunks))
			batch := chunks[batchStart:batchEnd]

			batchResults, batchErr := p.topics.ExtractBatch(topicCtx, batch)

			topicMu.Lock()
			defer topicMu.Unlock()

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
						zap.Int("chunks_total", len(chunks)),
						zap.Error(batchErr),
					)
					if topicFatalErr == nil {
						topicFatalErr = fmt.Errorf("batch %d: %w", batchIdx, batchErr)
						topicCancel()
					}
					return
				}
				p.logger.Warn("topic batch failed (transient), using fallback topic",
					zap.Int("batch", batchIdx), zap.Error(batchErr))
			} else {
				copy(topicResults[batchStart:batchEnd], batchResults)
			}

			topicCompleted[batchIdx] = true
			topicBatchesDone++
			for topicPrefix < totalTopicBatches && topicCompleted[topicPrefix] {
				topicPrefix++
			}
			chunksDone := min(topicPrefix*topicBatchSize, len(chunks))

			// Checkpoint on every completed batch (topics are expensive).
			cpWriter.Write(ctx, JobCheckpoint{
				Stage:          StageTopicExtraction,
				ChunksDone:     chunksDone,
				ChunksTotal:    len(chunks),
				LastBatchIndex: topicPrefix,
				CostUSDSoFar:   tracker.TotalCost(),
				StartedAt:      start,
			})
			tracker.UpdateDB(ctx, chunksDone, StageTopicExtraction,
				fmt.Sprintf("%d/%d chunks tagged", chunksDone, len(chunks)))
			p.emit(ctx, jobID, expertID, StageTopicExtraction, jobevents.KindBatchDone, map[string]interface{}{
				"batch_index":   batchIdx,
				"batches_total": totalTopicBatches,
				"batches_done":  topicBatchesDone,
				"chunks_done":   chunksDone,
				"chunks_total":  len(chunks),
				"workers":       p.workers,
			})
		}(batchIdx)
	}
	topicWG.Wait()
	ptimer.Stop("topic_extract")

	if topicFatalErr != nil {
		return nil, p.pauseOnLLMFailure(
			ctx, jobID, expertID, cpWriter,
			len(chunks), tracker.TotalCost(), start,
			fmt.Sprintf("Topic extraction fatal LLM error: %s", topicFatalErr.Error()),
		)
	}
	if resumeTopicBatch < totalTopicBatches {
		p.endStage(ctx, jobID, expertID, StageTopicExtraction, topicStarted, map[string]interface{}{
			"chunks":  len(chunks),
			"batches": totalTopicBatches,
			"topics":  countUniqueTopics(topicResults),
			"workers": p.workers,
		})
	}
	p.logger.Info("topic extraction complete", zap.Int("workers", p.workers))

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
		charterStarted := p.beginStage(ctx, jobID, expertID, StageCharterExtraction, "Extracting expert charter...")
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
				ctx, jobID, expertID, cpWriter,
				len(chunks), tracker.TotalCost(), start,
				fmt.Sprintf("Charter LLM failed: %s", extractErr.Error()),
			)
		}
		charter = extractedCharter
		p.endStage(ctx, jobID, expertID, StageCharterExtraction, charterStarted, map[string]interface{}{
			"charter_chars":                  len(charter.ReasoningCharter),
			"clarification_charter_entries":  len(charter.ClarificationCharter),
		})
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
	} else {
		// Resume: charter was already extracted and loaded from the DB.
		p.emit(ctx, jobID, expertID, StageCharterExtraction, jobevents.KindStageDone, map[string]interface{}{
			"from_checkpoint": true,
			"charter_chars":   len(charter.ReasoningCharter),
		})
	}
	p.logger.Info("charter extraction complete")

	// ============================================================
	// STEP 4: EMBEDDING (batched, resumable)
	// ============================================================
	embedStarted := p.beginStage(ctx, jobID, expertID, StageEmbedding,
		fmt.Sprintf("0/%d embeddings generated", len(chunks)))

	embeddings := make([][]float32, len(chunks))

	// WHY 25 not 100:
	//   100 chunks × bge-base-en-v1.5 on CPU ≈ 40s > 30s HTTP timeout.
	//   25 chunks ≈ 10s per batch — well within 300s timeout.
	//   Smaller batches also mean more frequent checkpoints.
	const embedBatchSize = 25
	// NOTE: embeddings are NOT skipped on resume. They exist only in memory
	// until step 6 stores them, so "skip the first N batches" would leave those
	// chunks with a nil vector and make the INSERT fail (or store a zero
	// vector). Embedding runs on the local ML sidecar, so recomputing is cheap;
	// correctness is not. Checkpointing still records progress for the UI.
	totalEmbedBatches := (len(chunks) + embedBatchSize - 1) / embedBatchSize

	ptimer.Start("embed")

	// T3: parallel embedding. Same design as the topic stage above:
	// independent batches, bounded workers, contiguous-prefix checkpointing,
	// fail-fast on the first hard error (an unavailable sidecar will fail
	// every batch — aborting beats burning minutes on doomed calls).
	embedCtx, embedCancel := context.WithCancel(ctx)
	defer embedCancel()

	var (
		embedWG          sync.WaitGroup
		embedMu          sync.Mutex
		embedErr         error
		embedCompleted   = make([]bool, totalEmbedBatches)
		embedPrefix      int
		embedBatchesDone int
	)
	embedSem := make(chan struct{}, p.workers)

	for batchIdx := 0; batchIdx < totalEmbedBatches; batchIdx++ {
		embedWG.Add(1)
		go func(batchIdx int) {
			defer embedWG.Done()

			if embedCtx.Err() != nil {
				return
			}
			select {
			case embedSem <- struct{}{}:
			case <-embedCtx.Done():
				return
			}
			defer func() { <-embedSem }()

			batchStart := batchIdx * embedBatchSize
			batchEnd := min(batchStart+embedBatchSize, len(chunks))

			texts := make([]string, batchEnd-batchStart)
			for i, c := range chunks[batchStart:batchEnd] {
				texts[i] = c.Text
			}

			batchEmbeds, err := p.embedder.Embed(embedCtx, texts)
			if err != nil {
				embedMu.Lock()
				if embedErr == nil {
					embedErr = fmt.Errorf("embedding batch %d failed: %w", batchIdx, err)
					embedCancel()
				}
				embedMu.Unlock()
				return
			}
			copy(embeddings[batchStart:batchEnd], batchEmbeds)

			embedMu.Lock()
			defer embedMu.Unlock()

			embedCompleted[batchIdx] = true
			embedBatchesDone++
			for embedPrefix < totalEmbedBatches && embedCompleted[embedPrefix] {
				embedPrefix++
			}
			chunksDone := min(embedPrefix*embedBatchSize, len(chunks))

			// Checkpoint on every completed batch.
			cpWriter.Write(ctx, JobCheckpoint{
				Stage:            StageEmbedding,
				ChunksDone:       chunksDone,
				ChunksTotal:      len(chunks),
				CharterExtracted: true,
				LastBatchIndex:   embedPrefix,
				CostUSDSoFar:     tracker.TotalCost(),
				StartedAt:        start,
			})
			tracker.UpdateDB(ctx, chunksDone, StageEmbedding,
				fmt.Sprintf("%d/%d embeddings generated", chunksDone, len(chunks)))
			p.emit(ctx, jobID, expertID, StageEmbedding, jobevents.KindBatchDone, map[string]interface{}{
				"batch_index":   batchIdx,
				"batches_total": totalEmbedBatches,
				"batches_done":  embedBatchesDone,
				"chunks_done":   chunksDone,
				"chunks_total":  len(chunks),
				"workers":       p.workers,
			})
		}(batchIdx)
	}
	embedWG.Wait()
	ptimer.Stop("embed")

	if embedErr != nil {
		p.updateJobStatus(ctx, jobID, "failed",
			"ML sidecar unavailable: "+embedErr.Error(), 0, len(chunks))
		p.emitFinal(ctx, jobID, expertID, StageEmbedding, jobevents.KindFailed, map[string]interface{}{
			"reason": embedErr.Error(),
		})
		return nil, fmt.Errorf("embedding failed: %w", embedErr)
	}
	p.endStage(ctx, jobID, expertID, StageEmbedding, embedStarted, map[string]interface{}{
		"chunks":  len(chunks),
		"batches": totalEmbedBatches,
		"workers": p.workers,
	})
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
	storeStarted := p.beginStage(ctx, jobID, expertID, StageStoring,
		fmt.Sprintf("Saving %d chunks to database...", len(chunks)))
	ptimer.Start("store")
	chunkIDs, storeStats, err := p.storeChunks(ctx, jobID, expertID, chunks, topicResults, embeddings, sourceFile)
	ptimer.Stop("store")
	if err != nil {
		p.updateJobStatus(ctx, jobID, "failed", "storage failed: "+err.Error(), 0, 0)
		p.emitFinal(ctx, jobID, expertID, StageStoring, jobevents.KindFailed, map[string]interface{}{
			"reason": err.Error(),
		})
		return nil, fmt.Errorf("chunk storage failed: %w", err)
	}
	p.logger.Info("chunks stored",
		zap.Int("resolved", len(chunkIDs)),
		zap.Int("parsed", storeStats.Parsed),
		zap.Int("inserted", storeStats.Inserted),
		zap.Int("duplicates", storeStats.Duplicates),
		zap.Int("reused", storeStats.Reused),
	)
	p.endStage(ctx, jobID, expertID, StageStoring, storeStarted, map[string]interface{}{
		"chunks":     len(chunkIDs),
		"parsed":     storeStats.Parsed,
		"inserted":   storeStats.Inserted,
		"duplicates": storeStats.Duplicates,
		"reused":     storeStats.Reused,
	})

	// T2: double confirmation. The pipeline is not trusted to grade its own
	// homework — re-read the database and compare it against what this run
	// claims it produced. The result is recorded as a `verified` event so the
	// admin sees "claim vs reality" without opening psql.
	//
	// WHY the claim is now ExpectedStored() and not len(chunks): the conflict
	// clause makes "parsed" and "stored" legitimately different numbers, so
	// comparing stored rows against the parse count reported a mismatched
	// database on every document that repeats text. The identity that actually
	// detects loss is stored == preexisting_for_file + inserted. The breakdown
	// (parsed / duplicates / inserted / reused) is emitted alongside it, so the
	// gap is explained rather than hidden.
	verifyTopics := countUniqueTopics(topicResults)
	verifyFallback := 0
	for _, t := range topicResults {
		if t.Topic == "" {
			verifyFallback++
		}
	}
	ledger := RunLedger{
		JobID:           jobID,
		ExpertID:        expertID,
		SourceFile:      sourceFile,
		Stats:           storeStats,
		FallbackClaimed: verifyFallback,
	}
	if verifyErr := p.verifyStoredCorpus(ctx, &ledger, verifyTopics); verifyErr != nil {
		p.logger.Warn("post-store verification failed (non-fatal)", zap.Error(verifyErr))
		ledger.VerificationStatus = VerificationNotChecked
		ledger.MismatchReason = verifyErr.Error()
	}
	p.writeRunLedger(ctx, ledger)

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
	// NOTE: "capability_build" is NOT a valid current_stage value (the column
	// has a CHECK constraint from migration 011), so this stage is tracked on
	// the event timeline only — the job row keeps showing 'storing'/'smoke_test'.
	capabilityStarted := time.Now()
	p.emit(ctx, jobID, expertID, "capability_build", jobevents.KindStageStarted, map[string]interface{}{
		"label": "Building capability table",
	})
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

	// Then derive the capability NUMBERS from the whole corpus, overriding
	// whatever the run just wrote.
	//
	// WHY unconditionally, and after the write above: the builder only ever sees
	// this run's chunks, and in append mode (the default) the corpus is the union
	// of every run. Without this step a topic touched by the newest course keeps
	// only the newest course's chunk_count, the topics the newest run never
	// mentions keep stale numbers, and the capability table stops matching the
	// corpus counts the expert row reports. Same function the reconcile action
	// and the LLM-failure fallback use, so all three can never disagree.
	p.buildCapabilitiesFromChunks(ctx, expertID)

	ptimer.Stop("capability_build")
	p.emit(ctx, jobID, expertID, "capability_build", jobevents.KindStageDone, map[string]interface{}{
		"duration_ms": time.Since(capabilityStarted).Milliseconds(),
		"topics":      len(capabilities),
		"llm_used":    err == nil,
	})

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
	smokeStarted := p.beginStage(ctx, jobID, expertID, StageSmokeTest, "Running smoke test...")
	ptimer.Start("smoke_test")
	smoke, smokeErr := p.runSmokeTest(ctx, expertID, expertName)
	ptimer.Stop("smoke_test")
	if smokeErr != nil {
		p.logger.Warn("smoke test error (non-fatal, expert stays draft)",
			zap.String("expert_id", expertID.String()),
			zap.Error(smokeErr),
		)
	}
	p.endStage(ctx, jobID, expertID, StageSmokeTest, smokeStarted, map[string]interface{}{
		"passed":        smoke.Passed,
		"probes_passed": smoke.PassedProbes,
		"probes_uncited": smoke.Uncited,
		"probes_unanswered": smoke.NoAnswer,
		"probes_total":  smokeTestProbeCount,
		"error":        errorString(smokeErr),
	})

	// I3: the training gate. DOMAIN_EXPERT_COLLABORATION_DESIGN.md §5.3 lists five
	// conditions for 'trained'; they are now evaluated explicitly and in one place,
	// and every unmet one is named in the job's warning text. Two of them were
	// enforced before (storage verification, and a smoke test that measured retrieval
	// only) and three were never read at all: charter rules, clarification coverage
	// and capability depth.
	//
	// The capability condition is measured DURING ingest. A gate that only passes
	// after a separate manual pass is a gate that mostly reports "unknown", and
	// "unknown" must not read as a pass. The measurer is injected, so an environment
	// without one fails the condition instead of silently skipping it.
	if p.capabilityMeasurer != nil {
		p.emit(ctx, jobID, expertID, StageSmokeTest, jobevents.KindCapabilityMeasured, map[string]interface{}{
			"status": "started",
			"topics": gateMeasureTopics,
		})
		if measureErr := p.capabilityMeasurer(ctx, expertID, gateMeasureTopics); measureErr != nil {
			// Non-fatal: the gate reports the capability as unmeasured and the expert
			// stays in draft, which is the correct outcome when the pass cannot run.
			p.logger.Warn("capability measurement failed during ingest (gate will report it unmeasured)",
				zap.String("expert_id", expertID.String()),
				zap.Error(measureErr))
			p.emit(ctx, jobID, expertID, StageSmokeTest, jobevents.KindCapabilityMeasured, map[string]interface{}{
				"status": "failed",
				"error":  measureErr.Error(),
			})
		}
	} else {
		p.logger.Warn("no capability measurer wired — the gate cannot see measured capability",
			zap.String("expert_id", expertID.String()))
	}

	gateInputs, gateErr := p.collectGateInputs(ctx, expertID, charter, smoke)
	if gateErr != nil {
		// Fail closed. An unreadable gate is not a passed gate, and silently treating
		// it as one would reintroduce exactly the problem this gate exists to fix.
		p.logger.Warn("ingest gate could not be evaluated — expert stays draft",
			zap.String("expert_id", expertID.String()),
			zap.Error(gateErr))
	}
	gateConditions := EvaluateIngestGate(gateInputs)
	gateOK := gateErr == nil && GatePassed(gateConditions)

	verificationOK := ledger.VerificationStatus == VerificationVerified
	warnReasons := make([]string, 0, len(gateConditions)+1)
	if !verificationOK {
		reason := ledger.MismatchReason
		if reason == "" {
			reason = "corpus verification could not be completed"
		}
		warnReasons = append(warnReasons, "corpus verification: "+reason)
	}
	warnReasons = append(warnReasons, UnmetGateReasons(gateConditions)...)

	finalStatus := "complete"
	warningText := ""
	if len(warnReasons) > 0 {
		finalStatus = "complete_with_warnings"
		warningText = "completed with warnings — " + strings.Join(warnReasons, "; ")
	}

	p.emitFinal(ctx, jobID, expertID, StageComplete, jobevents.KindGateEvaluated, map[string]interface{}{
		"passed":     gateOK,
		"conditions": gateConditions,
	})

	// 'trained' is what makes an expert publicly usable, so it now requires the whole
	// gate AND the storage verification. Fail closed: an expert that cannot answer
	// about its own corpus must not be handed to users.
	if gateOK && verificationOK {
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
		p.logger.Info("ingest gate passed and corpus verified — expert is trained",
			zap.String("expert_id", expertID.String()),
			zap.Int("measured_topics", gateInputs.MeasuredTopics),
			zap.Int("probes_passed", smoke.PassedProbes),
		)
	} else {
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
		p.logger.Warn("expert stays in draft — ingest gate not satisfied",
			zap.String("expert_id", expertID.String()),
			zap.Int("chunks", gateInputs.Chunks),
			zap.Int("charter_rules", gateInputs.CharterRules),
			zap.Int("clarification_topics", gateInputs.ClarificationTopics),
			zap.Int("measured_topics", gateInputs.MeasuredTopics),
			zap.Bool("capability_measured", gateInputs.CapabilityMeasured),
			zap.Bool("smoke_passed", smoke.Passed),
			zap.Bool("verification_ok", verificationOK),
			zap.Strings("unmet", UnmetGateReasons(gateConditions)),
		)
	}

	duration := time.Since(start).Milliseconds()
	p.updateJobStatus(ctx, jobID, finalStatus, warningText, len(chunks), len(chunks))

	p.logger.Info("ingestion complete",
		zap.String("expert_id", expertID.String()),
		zap.Int("chunks", len(chunks)),
		zap.Int("topics", uniqueTopics),
		zap.Int64("duration_ms", duration),
		zap.Bool("smoke_test_passed", smoke.Passed),
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

	// T1: close the timeline with the final ledger — everything an admin would
	// otherwise have to reconstruct from server logs, in one durable record.
	phaseTotals := map[string]int64{}
	for _, phase := range ptimer.Results() {
		phaseTotals[phase.Phase] = phase.Duration.Milliseconds()
	}
	p.emitFinal(ctx, jobID, expertID, StageComplete, jobevents.KindComplete, map[string]interface{}{
		"chunks_this_run": len(chunks),
		// Phase B breakdown: `chunks_this_run` is the parse count, which is not
		// the same as what landed in the corpus. Sending both lets the timeline
		// state the real numbers instead of implying 395 stored rows when 118 of
		// them were repeats of text already held.
		"chunks_parsed":     storeStats.Parsed,
		"chunks_duplicate":  storeStats.Duplicates,
		"chunks_inserted":   storeStats.Inserted,
		"chunks_reused":     storeStats.Reused,
		"verification":      ledger.VerificationStatus,
		"status":            finalStatus,
		"warnings":          warningText,
		"topics":            uniqueTopics,
		"corpus_total":      actualTotalChunks,
		"corpus_topics":     actualTotalTopics,
		"avg_depth":         avgDepth,
		"duration_ms":       duration,
		"smoke_test_passed": smoke.Passed,
		"smoke_probes":      smoke.PassedProbes,
		"gate_passed":       gateOK,
		"gate_conditions":   gateConditions,
		"cost_usd":          tracker.TotalCost(),
		"workers":           p.workers,
		"phase_ms":          phaseTotals,
		"source_file":       sourceFile,
	})

	return &IngestionResult{
		ExpertID:      expertID,
		TotalChunks:   len(chunks),
		TotalTopics:   uniqueTopics,
		AvgDepthLevel: avgDepth,
		DurationMs:    duration,
	}, nil
}

// storeChunks inserts all chunks into the database and links prev/next ids.
//
// T4 (batched): the original implementation issued one INSERT plus up to two
// link UPDATEs per chunk — ~45k round trips for a 15k-chunk transcript, which
// is why "Saving to database" dominated the run. It now:
//  1. INSERTs `storeBatchSize` rows per statement (ON CONFLICT DO NOTHING kept,
//     so append-mode dedup semantics are unchanged).
//  2. Resolves ids for the whole batch in ONE query (covers both freshly
//     inserted rows and pre-existing duplicates).
//  3. Links prev/next id in one batched UPDATE via unnest — after all inserts,
//     because a chunk's neighbours may live in another batch.
//
// Ordering is preserved: chunk_index still comes from the chunker, and the
// link pass walks chunkIDs in index order.
//
// WHY it returns StoreStats: ON CONFLICT DO NOTHING makes "how many chunks were
// parsed" and "how many rows exist" legitimately different numbers, and the
// command tag is the only place the true insert count exists. Discarding it is
// what let a 395-chunk document report 277 stored rows as a database failure.
// The stats are read from the database's own row counts, never from intent.
func (p *IngestionPipeline) storeChunks(
	ctx context.Context,
	jobID uuid.UUID,
	expertID uuid.UUID,
	chunks []TextChunk,
	topics []TopicResult,
	embeddings [][]float32,
	sourceFile string,
) ([]uuid.UUID, StoreStats, error) {
	chunkIDs := make([]uuid.UUID, len(chunks))
	stats := StoreStats{Parsed: len(chunks)}

	// C6: stamp each new chunk with the provider/model that produced its
	// vector, so a later embedding-provider/model change is detectable as a
	// re-embed requirement. Read once per ingestion (cheap). Existing
	// duplicate chunks keep their original stamp (ON CONFLICT DO NOTHING).
	embedProvider, embedModel := p.currentEmbedding(ctx)
	var embedModelArg interface{} // empty model → SQL NULL (sidecar has no model name)
	if embedModel != "" {
		embedModelArg = embedModel
	}

	// Safety: the chunker always populates ChunkHash, but a caller that built
	// TextChunk directly must still get a dedup key (never NULL).
	for i := range chunks {
		if chunks[i].ChunkHash == "" {
			chunks[i].ChunkHash = HashChunkText(chunks[i].Text)
		}
	}

	// Distinct dedup keys cap the rows this run can create, and the gap between
	// parsed and distinct is exactly the duplicate count the admin needs to see.
	// Counted AFTER the hash backfill above, so a chunk that arrived without a
	// hash is never mistaken for a repeat of another chunk.
	stats.UniqueHashes, stats.Duplicates = countDistinctHashes(chunks)

	// Measured BEFORE the inserts, so the storage identity can be checked after:
	// stored_for_file == preexisting_for_file + inserted. Every row this pass
	// creates carries sourceFile, and a row skipped by the conflict clause adds
	// nothing — so a failure of that identity means a row was really lost.
	if err := p.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM course_chunks WHERE expert_id = $1 AND source_file = $2`,
		expertID, sourceFile,
	).Scan(&stats.PreexistingForFile); err != nil {
		return nil, stats, fmt.Errorf("count existing chunks for file: %w", err)
	}

	// covered tracks the distinct chunks present in the corpus as the batches
	// land, which is what the stored-count actually is — not the parse count.
	covered := make(map[string]struct{}, len(chunks))

	// storeBatchSize trades statement size against round trips. 200 rows × 10
	// params = 2000 bind parameters — comfortably under Postgres's 65535 limit.
	const storeBatchSize = 200

	for batchStart := 0; batchStart < len(chunks); batchStart += storeBatchSize {
		batchEnd := min(batchStart+storeBatchSize, len(chunks))

		var sb strings.Builder
		sb.WriteString(`INSERT INTO course_chunks
			(expert_id, chunk_text, chunk_index, topic, subtopic, source_file, embedding, chunk_hash,
			 embedding_provider, embedding_model, section_path)
		 VALUES `)
		args := make([]interface{}, 0, (batchEnd-batchStart)*11)
		for i := batchStart; i < batchEnd; i++ {
			if i > batchStart {
				sb.WriteString(",")
			}
			base := len(args)
			sb.WriteString(fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)",
				base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8, base+9, base+10, base+11))

			topic := topics[i].Topic
			subtopic := ""
			if i < len(topics) {
				subtopic = topics[i].Subtopic
			}
			args = append(args,
				expertID, chunks[i].Text, chunks[i].Index, topic, subtopic,
				sourceFile, pgvector.NewVector(embeddings[i]), chunks[i].ChunkHash,
				embedProvider, embedModelArg, chunks[i].SectionPath,
			)
		}
		sb.WriteString(` ON CONFLICT (expert_id, chunk_hash) DO NOTHING`)

		tag, err := p.db.Exec(ctx, sb.String(), args...)
		if err != nil {
			return nil, stats, fmt.Errorf("failed to insert chunks %d..%d: %w", batchStart, batchEnd-1, err)
		}
		// The command tag is the only honest source for "how many rows this
		// statement created": ON CONFLICT DO NOTHING reports skipped rows by
		// simply not counting them, so len(batch) would overstate the store.
		batchInserted := int(tag.RowsAffected())
		stats.Inserted += batchInserted

		// Resolve ids for this batch (new rows AND deduplicated rows).
		idByHash, err := p.lookupChunkIDs(ctx, expertID, chunks[batchStart:batchEnd])
		if err != nil {
			return nil, stats, err
		}
		for i := batchStart; i < batchEnd; i++ {
			id, ok := idByHash[chunks[i].ChunkHash]
			if !ok {
				// The row must exist: it was either inserted above or already
				// present (ON CONFLICT). Missing means the dedup key changed
				// mid-write — surface it instead of storing a zero UUID.
				return nil, stats, fmt.Errorf("chunk %d missing after insert (hash %s)", i, chunks[i].ChunkHash)
			}
			chunkIDs[i] = id
			covered[chunks[i].ChunkHash] = struct{}{}
		}

		// The job row keeps counting PARSED chunks (batchEnd of len(chunks)): the
		// progress bar and ETA describe pipeline throughput, and every chunk sent
		// to the store really was processed. The event below reports distinct
		// chunks present in the corpus, which is the number that used to be
		// misreported as "stored 395/395" while the corpus held 277.
		p.updateJobProgress(ctx, jobID, batchEnd, len(chunks))
		p.emit(ctx, jobID, expertID, StageStoring, jobevents.KindChunkStored, map[string]interface{}{
			"chunks_done":    len(covered),
			"chunks_total":   stats.UniqueHashes,
			"batch_size":     batchEnd - batchStart,
			"batch_inserted": batchInserted,
		})
	}

	// Reused is derived, never counted per batch: each distinct hash either
	// created a row or already existed, so deriving it from UniqueHashes and
	// Inserted means the two can never disagree.
	stats.Reused = stats.UniqueHashes - stats.Inserted
	if stats.Reused < 0 {
		// Only reachable if the database reported more inserted rows than
		// distinct hashes were sent. Clamp so the ledger never stores a negative
		// count, and leave the contradiction for verification to report.
		stats.Reused = 0
	}

	// Link pass (after every insert, so cross-batch neighbours resolve).
	p.linkChunks(ctx, chunkIDs)

	return chunkIDs, stats, nil
}

// lookupChunkIDs resolves chunk_hash → id for the given chunks in one query.
//
// WHY by hash and not RETURNING: ON CONFLICT DO NOTHING returns nothing for a
// duplicate, so a single INSERT cannot tell us the id of a pre-existing row.
// One SELECT for the whole batch covers both cases.
func (p *IngestionPipeline) lookupChunkIDs(ctx context.Context, expertID uuid.UUID, chunks []TextChunk) (map[string]uuid.UUID, error) {
	if len(chunks) == 0 {
		return map[string]uuid.UUID{}, nil
	}
	hashes := make([]string, 0, len(chunks))
	for _, c := range chunks {
		hashes = append(hashes, c.ChunkHash)
	}
	rows, err := p.db.Query(ctx,
		`SELECT chunk_hash, id FROM course_chunks
		  WHERE expert_id = $1 AND chunk_hash = ANY($2)`,
		expertID, hashes,
	)
	if err != nil {
		return nil, fmt.Errorf("lookup chunk ids: %w", err)
	}
	defer rows.Close()

	out := make(map[string]uuid.UUID, len(hashes))
	for rows.Next() {
		var hash string
		var id uuid.UUID
		if scanErr := rows.Scan(&hash, &id); scanErr != nil {
			continue
		}
		out[hash] = id
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("lookup chunk ids: rows error: %w", err)
	}
	return out, nil
}

// linkChunks writes prev_chunk_id / next_chunk_id for the run's chunks.
//
// Batched with unnest: one UPDATE per linkBatch rows instead of two UPDATEs per
// chunk. uuid.Nil is the "no neighbour" sentinel and is turned back into SQL
// NULL by NULLIF, so the first/last chunk keep NULL links.
//
// Non-fatal: navigation links are a context convenience, not correctness.
func (p *IngestionPipeline) linkChunks(ctx context.Context, chunkIDs []uuid.UUID) {
	if len(chunkIDs) == 0 {
		return
	}

	// Duplicate text inside one document resolves to the same row id, so this
	// slice can carry one id at two positions. Linking it verbatim would make a
	// row its own neighbour (two source rows updating the same target, with
	// Postgres free to pick either). Collapse to first-appearance order so the
	// chain runs over distinct rows.
	unique := make([]uuid.UUID, 0, len(chunkIDs))
	seen := make(map[uuid.UUID]struct{}, len(chunkIDs))
	for _, id := range chunkIDs {
		if id == uuid.Nil {
			continue
		}
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	if len(unique) == 0 {
		return
	}

	const linkBatch = 500

	for start := 0; start < len(unique); start += linkBatch {
		end := min(start+linkBatch, len(unique))

		ids := make([]uuid.UUID, 0, end-start)
		prev := make([]uuid.UUID, 0, end-start)
		next := make([]uuid.UUID, 0, end-start)
		for i := start; i < end; i++ {
			var prevID, nextID uuid.UUID
			if i > 0 {
				prevID = unique[i-1]
			}
			if i+1 < len(unique) {
				nextID = unique[i+1]
			}
			ids = append(ids, unique[i])
			prev = append(prev, prevID)
			next = append(next, nextID)
		}

		_, err := p.db.Exec(ctx,
			`UPDATE course_chunks AS c
			    SET prev_chunk_id = NULLIF(v.prev_id, '00000000-0000-0000-0000-000000000000'::uuid),
			        next_chunk_id = NULLIF(v.next_id, '00000000-0000-0000-0000-000000000000'::uuid)
			   FROM (SELECT unnest($1::uuid[]) AS id,
			                unnest($2::uuid[]) AS prev_id,
			                unnest($3::uuid[]) AS next_id) AS v
			  WHERE c.id = v.id`,
			ids, prev, next)
		if err != nil {
			p.logger.Warn("failed to link chunk navigation (non-fatal)", zap.Error(err))
			return
		}
	}
}

// verifyStoredCorpus is the T2 "double confirmation" step: it re-reads the
// database, compares it against what this run claims it produced, and records
// the verdict on the ledger and on the timeline as a `verified` event.
//
// WHY the check stays but the claim changed: previously the only way to confirm
// ingestion had really stored what it claimed was to open psql. The check is
// sound; the expectation was not. It compared stored rows against the number of
// chunks PARSED, which made every document that repeats text look like a
// database mismatch — 395 parsed, 277 stored, 118 of them repeats of text the
// corpus already held. A warning that fires on the normal case is worse than no
// warning, because it teaches the admin to ignore the amber box.
//
// The claim is now the storage identity:
//
//	stored_for_file == preexisting_for_file + inserted
//
// Every inserted row carries this file, and a row skipped by the conflict clause
// adds nothing — so this can only fail when a row was genuinely lost, or
// something else wrote to the same expert+file. The parsed/duplicate/inserted
// breakdown travels with it, so the gap is explained instead of hidden.
//
// Topic counts are REPORTED but do not gate the verdict: dedup drops repeated
// chunks, so the number of distinct topics behind the surviving rows is not
// comparable to the topics the run produced. Gating on it would reintroduce the
// same class of false alarm.
func (p *IngestionPipeline) verifyStoredCorpus(
	ctx context.Context,
	l *RunLedger,
	claimedTopics int,
) error {
	var stored, storedTopics, storedGeneral, nullEmbeddings int
	err := p.db.QueryRow(ctx, `
		SELECT COUNT(*),
		       COUNT(DISTINCT topic),
		       COUNT(*) FILTER (WHERE topic IS NULL OR topic = '' OR topic = 'general'),
		       COUNT(*) FILTER (WHERE embedding IS NULL)
		  FROM course_chunks
		 WHERE expert_id = $1 AND source_file = $2`,
		l.ExpertID, l.SourceFile,
	).Scan(&stored, &storedTopics, &storedGeneral, &nullEmbeddings)
	if err != nil {
		return fmt.Errorf("verify stored corpus: %w", err)
	}

	// Corpus-wide count gives the admin the "after append" picture, not just
	// this file's contribution (append mode is the default).
	var corpusTotal int
	if scanErr := p.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM course_chunks WHERE expert_id = $1`, l.ExpertID,
	).Scan(&corpusTotal); scanErr != nil {
		p.logger.Warn("verify: corpus total query failed (non-fatal)", zap.Error(scanErr))
	}

	l.StoredForFile = stored
	l.GeneralStored = storedGeneral
	l.NullEmbeddings = nullEmbeddings
	l.CorpusTotal = corpusTotal

	expected := l.ExpectedStored()
	switch {
	case stored != expected:
		l.VerificationStatus = VerificationMismatch
		l.MismatchReason = fmt.Sprintf(
			"expected %d rows for this file (%d pre-existing + %d inserted), found %d",
			expected, l.Stats.PreexistingForFile, l.Stats.Inserted, stored,
		)
	case nullEmbeddings > 0:
		l.VerificationStatus = VerificationMismatch
		l.MismatchReason = fmt.Sprintf("%d stored chunks have no embedding", nullEmbeddings)
	default:
		l.VerificationStatus = VerificationVerified
		l.MismatchReason = ""
	}
	ok := l.VerificationStatus == VerificationVerified

	p.emit(ctx, l.JobID, l.ExpertID, StageStoring, jobevents.KindVerified, map[string]interface{}{
		"ok": ok,
		"claim": map[string]interface{}{
			"chunks": expected,
			"topics": claimedTopics,
		},
		"reality": map[string]interface{}{
			"chunks":          stored,
			"topics":          storedTopics,
			"general_chunks":  storedGeneral,
			"null_embeddings": nullEmbeddings,
		},
		// The breakdown turns "277 is not 395" into something the admin can act
		// on: 118 of those chunks were repeats of text already in the corpus.
		"breakdown": map[string]interface{}{
			"parsed":               l.Stats.Parsed,
			"duplicate":            l.Stats.Duplicates,
			"inserted":             l.Stats.Inserted,
			"reused":               l.Stats.Reused,
			"preexisting_for_file": l.Stats.PreexistingForFile,
			"stored_for_file":      stored,
		},
		"fallback_chunks":     l.FallbackClaimed,
		"corpus_total_chunks": corpusTotal,
		"source_file":         l.SourceFile,
		"mismatch_reason":     l.MismatchReason,
		"summary":             l.describeLedger(),
	})
	return nil
}

// errorString renders err for a JSON event detail, using "" for nil so the
// timeline never contains the literal "<nil>".
func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// buildCapabilitiesFromChunks (re)derives expert_capabilities from the topics the
// corpus actually holds, without LLM analysis.
//
// Two callers, one definition: the fallback when the capability build fails, and
// the end of every successful ingest — where it corrects the per-run numbers the
// builder wrote with corpus-wide ones. Delegates to
// upsertCapabilitiesFromChunks (reconcile.go) so the reconcile action, the
// fallback and the ingest step can never disagree about what a chunk-derived
// capability is. Best-effort by contract: it logs and returns, because neither a
// failed fallback nor a failed correction may fail an otherwise-complete run.
func (p *IngestionPipeline) buildCapabilitiesFromChunks(ctx context.Context, expertID uuid.UUID) {
	written, err := upsertCapabilitiesFromChunks(ctx, p.db, p.logger, expertID)
	if err != nil {
		p.logger.Warn("buildCapabilitiesFromChunks: query failed", zap.Error(err))
		return
	}
	p.logger.Info("buildCapabilitiesFromChunks: done",
		zap.String("expert_id", expertID.String()),
		zap.Int("topics_inserted", written),
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
	expertID uuid.UUID,
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
			p.emitFinal(ctx, jobID, expertID, StageFailed, jobevents.KindFailed, map[string]interface{}{
				"reason": fmt.Sprintf("could not persist pause state: %s", fallbackErr.Error()),
			})
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

	// T1: record the pause as a first-class event, not just a status column —
	// this is the row the UI turns into the amber "action required" timeline
	// entry, and it survives refresh.
	p.emitFinal(ctx, jobID, expertID, StagePaused, jobevents.KindPaused, map[string]interface{}{
		"reason":          reason,
		"chunks_total":    chunksTotal,
		"cost_usd":        costSoFar,
		"resume_stage":    StageCharterExtraction,
		"needs_admin":     true,
		"hint":            "Fix the cause (provider credits or the model's token budget), then use Retry now.",
	})

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

// loadTopicsFromDB reads the already-stored topic/subtopic for each chunk of
// THIS job's source file, ordered by chunk_index so the result aligns 1:1 with
// the chunks loaded by loadChunksFromDB.
//
// WHY: on resume past topic extraction we must NOT re-run the topic LLM, but
// storeChunks/capability.Build still need the real topics (not the "general"
// fallback). The stored rows are the source of truth.
func (p *IngestionPipeline) loadTopicsFromDB(ctx context.Context, expertID uuid.UUID, sourceFile string, want int) ([]TopicResult, error) {
	rows, err := p.db.Query(ctx,
		`SELECT COALESCE(topic,''), COALESCE(subtopic,'')
		   FROM course_chunks
		  WHERE expert_id=$1 AND source_file=$2
		  ORDER BY chunk_index ASC`,
		expertID, sourceFile,
	)
	if err != nil {
		return nil, fmt.Errorf("loadTopicsFromDB: %w", err)
	}
	defer rows.Close()

	out := make([]TopicResult, 0, want)
	for rows.Next() {
		t := TopicResult{Confidence: 1.0} // already accepted at first extraction
		if err := rows.Scan(&t.Topic, &t.Subtopic); err != nil {
			return nil, err
		}
		if t.Topic == "" {
			t.Topic = "general"
		}
		out = append(out, t)
	}
	return out, rows.Err()
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
	if status == "complete" || status == "complete_with_warnings" || status == "failed" {
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
			current_stage = CASE WHEN $1::varchar IN ('complete', 'complete_with_warnings') THEN 'complete'
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

// SmokeTestResult is what the smoke test measured, per §5.3(5): an expert must answer
// its own topic probes with citations, not merely retrieve a matching chunk.
//
// WHY separate counters instead of one pass/fail: "retrieved but never cited" and
// "declined to answer" are different defects with different fixes — the first is a
// grounding problem, the second a coverage or prompt problem — and a single boolean
// would hide which one the expert has.
type SmokeTestResult struct {
	Passed       bool
	Probes       int
	PassedProbes int
	// Uncited counts probes whose retrieval matched but whose answer carried no
	// citation marker.
	Uncited int
	// NoAnswer counts probes where the answer was empty or an explicit refusal.
	NoAnswer int
}

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
) (SmokeTestResult, error) {
	// §5.3(5) requires an actual response with citations, so the probe now answers as
	// well as retrieves. The answerer is shared with the capability evaluation, so the
	// citation rule cannot drift between the two places that enforce it.
	answerer := newGroundedAnswerer(p.gateway, p.logger)
	result := SmokeTestResult{}

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
		return result, fmt.Errorf("smoke test: failed to load topics: %w", err)
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
		return result, fmt.Errorf("smoke test: topic scan error: %w", err)
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
	passCount := 0
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
			return result, fmt.Errorf("smoke test: ML sidecar unavailable: %w", embedErr)
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
			return result, fmt.Errorf("smoke test: rerank unavailable: %w", rerankErr)
		}

		// Step 6: Check if best rerank score meets threshold.
		bestScore := float32(0)
		for _, r := range rankResults {
			if r.Score > bestScore {
				bestScore = r.Score
			}
		}

		probePassed := bestScore >= smokeTestRerankThreshold

		// §5.3(5) asks for "non-REFUSE responses WITH CITATIONS", not for a good
		// retrieval score. Until now this test only measured retrieval, so an expert
		// that retrieved the right chunk and then answered without grounding it —
		// or declined to answer at all — passed the gate. The answer is only checked
		// when retrieval already matched: a probe that never found the chunk is
		// already explained, and asking the model to answer from poor context would
		// spend a call to learn the same thing.
		if probePassed {
			contextBlock := buildNumberedContext(candidates, promptPassageChars)
			answer, answerErr := answerer.Answer(ctx, probeQuestion, contextBlock)
			switch {
			case answerErr != nil:
				// Infrastructure, not corpus quality: count it as unusable and let the
				// caller decide, exactly as the embed/rerank failures above do.
				p.logger.Warn("smoke test: answering failed for probe",
					zap.Int("probe_index", i),
					zap.String("topic", topic),
					zap.Error(answerErr),
				)
				result.NoAnswer++
				probePassed = false
			default:
				switch refusalOrCitationFailure(answer) {
				case FailureNotCited:
					result.Uncited++
					probePassed = false
				case FailureRefused, FailureEmpty:
					result.NoAnswer++
					probePassed = false
				}
			}
		}

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
		result.Passed = false
		return result, nil
	}

	passed := passCount >= effectiveThreshold
	result.Passed = passed
	result.PassedProbes = passCount
	result.Probes = len(topics)

	p.logger.Info("smoke test complete",
		zap.String("expert_id", expertID.String()),
		zap.Int("probes_run", len(topics)),
		zap.Int("probes_passed", passCount),
		zap.Int("probes_uncited", result.Uncited),
		zap.Int("probes_unanswered", result.NoAnswer),
		zap.Int("effective_threshold", effectiveThreshold),
		zap.Bool("passed", passed),
	)

	return result, nil
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
// isFatalLLMError reports whether a topic-extraction LLM error is systemic
// (payment / auth / provider misconfiguration) rather than transient.
//
// WHY empty-content is fatal too: when a reasoning model spends its whole
// completion budget on internal reasoning, providers return an empty content
// ("empty content in response"). Treating that as transient is exactly how we
// ended up tagging EVERY chunk with topic="general" — the "all-general-topic
// disaster" this function's caller exists to avoid. Pausing is the correct,
// fail-closed outcome; the admin sees a job that needs action instead of a
// silently degraded expert.
func isFatalLLMError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	// Match the exact format from providers/common.go:
	// fmt.Errorf("provider returned status %d", resp.StatusCode)
	if strings.Contains(msg, "status 402") ||
		strings.Contains(msg, "status 401") ||
		strings.Contains(msg, "status 403") {
		return true
	}
	return isEmptyContentError(msg)
}

// isEmptyContentError reports whether an LLM error means "the provider
// returned a 200 with no usable text" — see providers/common.go. Kept separate
// so both topic extraction and the charter path classify it the same way.
func isEmptyContentError(msg string) bool {
	return strings.Contains(msg, "empty content in response") ||
		strings.Contains(msg, "empty choices in response") ||
		strings.Contains(msg, "token budget exhausted")
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
