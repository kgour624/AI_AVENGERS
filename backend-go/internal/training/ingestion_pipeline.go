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
	db         *pgxpool.Pool
	ml         *ml.SidecarClient
	gateway    *gateway.ModelGateway
	chunker    *TextChunker
	topics     *TopicExtractor
	charters   *CharterExtractor
	capability *CapabilityBuilder
	logger     *zap.Logger
}

// NewIngestionPipeline creates a new ingestion pipeline.
func NewIngestionPipeline(
	db *pgxpool.Pool,
	mlClient *ml.SidecarClient,
	gw *gateway.ModelGateway,
	logger *zap.Logger,
) *IngestionPipeline {
	return &IngestionPipeline{
		db:         db,
		ml:         mlClient,
		gateway:    gw,
		chunker:    NewTextChunker(DefaultChunkerConfig()),
		topics:     NewTopicExtractor(gw, mlClient, logger),
		charters:   NewCharterExtractor(gw, logger),
		capability: NewCapabilityBuilder(gw, logger),
		logger:     logger,
	}
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
	// STEP 1: CHUNKING
	// ============================================================
	// If resuming from topic_extraction or later, load chunks from DB
	// instead of re-chunking (chunking is deterministic but expensive for large transcripts).
	var chunks []TextChunk
	if isResume && StageOrder[cp.Stage] >= StageOrder[StageTopicExtraction] {
		// Load existing chunks from DB
		chunks, err = p.loadChunksFromDB(ctx, expertID)
		if err != nil || len(chunks) == 0 {
			p.logger.Warn("could not load chunks from DB, re-chunking", zap.Error(err))
			chunks = p.chunker.Chunk(transcript)
		}
		p.logger.Info("resume: loaded chunks from DB", zap.Int("count", len(chunks)))
	} else {
		p.updateStage(ctx, jobID, StageChunking, "Splitting transcript...")
		chunks = p.chunker.Chunk(transcript)
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
			p.logger.Warn("topic batch failed, using fallback",
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
		charter, err = p.charters.Extract(ctx, transcript, expertName)
		if err != nil {
			p.logger.Warn("charter extraction failed, using default", zap.Error(err))
			charter = &Charter{
				ReasoningCharter:     defaultReasoningCharter(expertName),
				ClarificationCharter: defaultClarificationCharter(),
			}
		}
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

	const embedBatchSize = 100
	resumeEmbedBatch := 0
	if isResume && cp != nil && cp.Stage == StageEmbedding {
		resumeEmbedBatch = cp.LastBatchIndex
		p.logger.Info("resume: skipping embed batches", zap.Int("skip_to_batch", resumeEmbedBatch))
	}

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

		batchEmbeds, embedErr := p.ml.Embed(ctx, texts)
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
	chunkIDs, err := p.storeChunks(ctx, expertID, chunks, topicResults, embeddings, sourceFile)
	if err != nil {
		p.updateJobStatus(ctx, jobID, "failed", "storage failed: "+err.Error(), 0, 0)
		return nil, fmt.Errorf("chunk storage failed: %w", err)
	}
	p.logger.Info("chunks stored", zap.Int("count", len(chunkIDs)))

	// Step 7: Update expert charters
	clarificationJSON, _ := json.Marshal(charter.ClarificationCharter)
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
	if err != nil {
		p.logger.Warn("failed to update charters", zap.Error(err))
	}

	// Step 8: Build capability table
	capabilities, err := p.capability.Build(ctx, chunks, topicResults)
	if err != nil {
		p.logger.Warn("capability build failed", zap.Error(err))
	} else {
		p.storeCapabilities(ctx, expertID, capabilities)
	}

	// Step 9: Update expert stats
	uniqueTopics := countUniqueTopics(topicResults)
	avgDepth := calculateAvgDepth(capabilities)

	_, err = p.db.Exec(ctx,
		`UPDATE experts SET
			total_chunks = $1,
			total_topics = $2,
			avg_depth_level = $3,
			is_training = FALSE,
			updated_at = NOW()
		 WHERE id = $4`,
		len(chunks), uniqueTopics, avgDepth, expertID,
	)
	if err != nil {
		p.logger.Warn("failed to update expert stats", zap.Error(err))
	}

	// Step 10: Smoke test — verify the expert is actually retrievable.
	// WHY here: chunks + capabilities are in DB, so retrieval is possible.
	// WHY before marking complete: training_status must reflect real state.
	smokeTestPassed, smokePassCount, smokeErr := p.runSmokeTest(ctx, expertID, expertName)
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
	expertID uuid.UUID,
	chunks []TextChunk,
	topics []TopicResult,
	embeddings [][]float32,
	sourceFile string,
) ([]uuid.UUID, error) {
	chunkIDs := make([]uuid.UUID, len(chunks))

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
				(expert_id, chunk_text, chunk_index, topic, subtopic, source_file, embedding, chunk_hash)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			 ON CONFLICT (expert_id, chunk_hash) DO NOTHING
			 RETURNING id`,
			expertID, chunk.Text, chunk.Index, topic, subtopic, sourceFile, embedding, chunk.ChunkHash,
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
			p.updateJobProgress(ctx, chunkIDs[0], i+1, len(chunks))
		}
	}

	return chunkIDs, nil
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
		canHandleJSON, _ := json.Marshal(cap.CanHandle)
		cannotHandleJSON, _ := json.Marshal(cap.CannotHandle)
		exampleQJSON, _ := json.Marshal(cap.ExampleQuestions)

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
			string(canHandleJSON), string(cannotHandleJSON), string(exampleQJSON),
		)
		if err != nil {
			p.logger.Warn("failed to store capability",
				zap.String("topic", cap.Topic),
				zap.Error(err),
			)
		}
	}
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

	_, err := p.db.Exec(ctx,
		`UPDATE ingestion_jobs SET
			status = $1,
			error_message = NULLIF($2, ''),
			processed_chunks = $3,
			total_chunks = $4,
			completed_at = $5,
			started_at = CASE WHEN $1 = 'running' THEN NOW() ELSE started_at END
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

	// If no capabilities yet (edge case: capability build failed), fall back
	// to a single generic probe about the expert's domain.
	if len(topics) == 0 {
		p.logger.Warn("smoke test: no capabilities found, using generic probe",
			zap.String("expert_id", expertID.String()),
		)
		topics = []string{"general"}
	}

	// Step 2–6: Probe each topic.
	passCount = 0
	for i, topic := range topics {
		probeQuestion := fmt.Sprintf(
			"What does %s teach about %s?",
			expertName, topic,
		)

		// Step 3: Embed the probe question.
		embeddings, embedErr := p.ml.Embed(ctx, []string{probeQuestion})
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
		rankResults, rerankErr := p.ml.Rerank(ctx, probeQuestion, candidates, len(candidates))
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

	passed = passCount >= smokeTestPassThreshold

	p.logger.Info("smoke test complete",
		zap.String("expert_id", expertID.String()),
		zap.Int("probes_run", len(topics)),
		zap.Int("probes_passed", passCount),
		zap.Bool("passed", passed),
	)

	return passed, passCount, nil
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
