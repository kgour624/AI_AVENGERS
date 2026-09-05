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
// Mental execution:
// Input: expertID, transcript text
// 1. Update job status to "running"
// 2. Clean and chunk text
// 3. Extract topics (batched LLM calls)
// 4. Extract charters (strong LLM call)
// 5. Generate embeddings (ML sidecar)
// 6. Store chunks in DB
// 7. Build capability table
// 8. Update expert stats
// 9. Update job status to "complete"
// On any error: update job status to "failed", cleanup partial data
func (p *IngestionPipeline) IngestTranscript(
	ctx context.Context,
	jobID uuid.UUID,
	expertID uuid.UUID,
	expertName string,
	transcript string,
	sourceFile string,
) (*IngestionResult, error) {
	start := time.Now()

	p.logger.Info("ingestion started",
		zap.String("expert_id", expertID.String()),
		zap.String("job_id", jobID.String()),
		zap.Int("transcript_length", len(transcript)),
	)

	// Update job status to running
	p.updateJobStatus(ctx, jobID, "running", "", 0, 0)

	// Step 1: Chunk text
	chunks := p.chunker.Chunk(transcript)
	if len(chunks) == 0 {
		err := fmt.Errorf("no chunks created from transcript")
		p.updateJobStatus(ctx, jobID, "failed", err.Error(), 0, 0)
		return nil, err
	}
	p.logger.Info("chunking complete", zap.Int("chunks", len(chunks)))

	// Step 2: Extract topics (batched)
	topicResults, err := p.topics.ExtractBatch(ctx, chunks)
	if err != nil {
		p.logger.Warn("topic extraction failed, using fallback", zap.Error(err))
		topicResults = make([]TopicResult, len(chunks))
		for i := range topicResults {
			topicResults[i] = TopicResult{Topic: "general", Confidence: 0.5}
		}
	}
	p.logger.Info("topic extraction complete")

	// Step 3: Extract charters
	charter, err := p.charters.Extract(ctx, transcript, expertName)
	if err != nil {
		p.logger.Warn("charter extraction failed, using default", zap.Error(err))
		charter = &Charter{
			ReasoningCharter:     defaultReasoningCharter(expertName),
			ClarificationCharter: defaultClarificationCharter(),
		}
	}
	p.logger.Info("charter extraction complete")

	// Step 4: Generate embeddings
	texts := make([]string, len(chunks))
	for i, c := range chunks {
		texts[i] = c.Text
	}

	embeddings, err := p.ml.Embed(ctx, texts)
	if err != nil {
		p.updateJobStatus(ctx, jobID, "failed", "ML sidecar unavailable: "+err.Error(), 0, 0)
		return nil, fmt.Errorf("embedding generation failed: %w", err)
	}
	p.logger.Info("embeddings generated", zap.Int("count", len(embeddings)))

	// Step 5: Delete existing chunks (re-ingestion support)
	_, err = p.db.Exec(ctx,
		`DELETE FROM course_chunks WHERE expert_id = $1`,
		expertID,
	)
	if err != nil {
		p.updateJobStatus(ctx, jobID, "failed", "cleanup failed: "+err.Error(), 0, 0)
		return nil, fmt.Errorf("cleanup failed: %w", err)
	}

	// Step 6: Store chunks
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

	duration := time.Since(start).Milliseconds()
	p.updateJobStatus(ctx, jobID, "complete", "", len(chunks), len(chunks))

	p.logger.Info("ingestion complete",
		zap.String("expert_id", expertID.String()),
		zap.Int("chunks", len(chunks)),
		zap.Int("topics", uniqueTopics),
		zap.Int64("duration_ms", duration),
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
// Mental execution:
// Insert chunk 0 -> get ID
// Insert chunk 1 -> get ID, set prev_chunk_id = chunk 0 ID
// Update chunk 0 -> set next_chunk_id = chunk 1 ID
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

		var chunkID uuid.UUID
		err := p.db.QueryRow(ctx,
			`INSERT INTO course_chunks
				(expert_id, chunk_text, chunk_index, topic, subtopic, source_file, embedding)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)
			 RETURNING id`,
			expertID, chunk.Text, chunk.Index, topic, subtopic, sourceFile, embedding,
		).Scan(&chunkID)
		if err != nil {
			return nil, fmt.Errorf("failed to insert chunk %d: %w", i, err)
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
