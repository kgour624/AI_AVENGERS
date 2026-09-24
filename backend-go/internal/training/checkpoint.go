package training

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// ErrJobPaused is returned by IngestTranscript when the pipeline
// intentionally stops due to a charter LLM failure and sets
// status='paused'. The caller (admin_handler goroutine) must treat
// this as a non-error stop — the job is not broken, it is waiting
// for admin action (Retry Now or Pause & Wait).
//
// WHY sentinel error not bool return:
//   IngestTranscript already returns (*IngestionResult, error).
//   Adding a third return value would require updating every caller.
//   A sentinel error is idiomatic Go for "stop, but not a crash".
var ErrJobPaused = errors.New("ingestion job paused: charter LLM failure — awaiting admin action")

// Stage constants — must match migration 007 CHECK constraint on
// current_stage AND migration 011 CHECK constraint on status.
const (
	StagePending           = "pending"
	StageChunking          = "chunking"
	StageTopicExtraction   = "topic_extraction"
	StageCharterExtraction = "charter_extraction"
	StageEmbedding         = "embedding"
	StageStoring           = "storing"
	StageSmokeTest         = "smoke_test"
	StageComplete          = "complete"
	StageFailed            = "failed"
	// StagePaused: pipeline stopped at charter_extraction due to LLM
	// failure. Job is NOT failed — it is waiting for admin action.
	// Admin can: Retry Now (re-run charter LLM) or leave paused
	// (auto-fails after 24h via background checker in main.go).
	StagePaused = "paused"
)

// StageLabels maps stage constants to human-readable UI labels.
var StageLabels = map[string]string{
	StagePending:           "Waiting to start",
	StageChunking:          "Step 1/6: Splitting transcript into chunks",
	StageTopicExtraction:   "Step 2/6: Extracting topics (LLM)",
	StageCharterExtraction: "Step 3/6: Extracting expert charter (LLM)",
	StageEmbedding:         "Step 4/6: Generating vector embeddings",
	StageStoring:           "Step 5/6: Saving to database",
	StageSmokeTest:         "Step 6/6: Running smoke test",
	StageComplete:          "Complete ✅",
	StageFailed:            "Failed ❌",
	StagePaused:            "Paused ⏸ — Charter LLM failed. Use Retry Now or wait.",
}

// StageOrder maps stage → index for progress % calculation.
var StageOrder = map[string]int{
	StagePending:           0,
	StageChunking:          1,
	StageTopicExtraction:   2,
	StageCharterExtraction: 3,
	StageEmbedding:         4,
	StageStoring:           5,
	StageSmokeTest:         6,
	StageComplete:          7,
	StagePaused:            3, // paused at charter stage — same order as charter_extraction
}

// JobCheckpoint is the durable state written to ingestion_jobs.checkpoint_data.
// Written every checkpointEveryN chunks. Survives server crashes.
//
// WHY JSONB not separate table:
//   One row per job, one JSONB column. Simple, no joins.
//   Checkpoint is only read on resume — not hot path.
//
// Resume algorithm:
//   1. Load checkpoint from DB
//   2. If Stage >= topic_extraction: load existing chunks from DB (skip chunking)
//   3. If Stage >= charter_extraction: skip topic extraction up to LastBatchIndex
//   4. If CharterExtracted: skip charter extraction
//   5. If Stage >= embedding: skip embedding up to LastBatchIndex
//   6. Stage storing: ON CONFLICT DO NOTHING = always safe to re-run
type JobCheckpoint struct {
	Stage            string    `json:"stage"`
	ChunksDone       int       `json:"chunks_done"`
	ChunksTotal      int       `json:"chunks_total"`
	CharterExtracted bool      `json:"charter_extracted"`
	LastBatchIndex   int       `json:"last_batch_index"` // last completed batch (0-indexed)
	CostUSDSoFar     float64   `json:"cost_usd_so_far"`
	StartedAt        time.Time `json:"started_at"`
}

// checkpointEveryN is how often (in chunks) we write a checkpoint.
// 50 chunks = worst case: redo 50 chunks on resume (idempotent via chunk_hash).
// 50 chunks at 8 chunks/sec = 6 seconds of re-work maximum.
const checkpointEveryN = 50

// CheckpointWriter writes job checkpoints to the DB.
type CheckpointWriter struct {
	db     *pgxpool.Pool
	jobID  uuid.UUID
	logger *zap.Logger
}

func NewCheckpointWriter(db *pgxpool.Pool, jobID uuid.UUID, logger *zap.Logger) *CheckpointWriter {
	return &CheckpointWriter{db: db, jobID: jobID, logger: logger}
}

// Write persists a checkpoint to the DB.
// Called every checkpointEveryN chunks and at every stage transition.
func (w *CheckpointWriter) Write(ctx context.Context, cp JobCheckpoint) {
	cpJSON, err := json.Marshal(cp)
	if err != nil {
		w.logger.Warn("checkpoint marshal failed", zap.Error(err))
		return
	}
	_, err = w.db.Exec(ctx,
		`UPDATE ingestion_jobs SET
			checkpoint_data = $1,
			current_stage   = $2,
			cost_usd        = $3
		 WHERE id = $4`,
		string(cpJSON), cp.Stage, cp.CostUSDSoFar, w.jobID,
	)
	if err != nil {
		// Non-fatal: checkpoint failure doesn't stop ingestion.
		// Worst case: resume from an earlier checkpoint.
		w.logger.Warn("checkpoint write failed (non-fatal)",
			zap.String("job_id", w.jobID.String()),
			zap.Error(err),
		)
	}
}

// Load reads the latest checkpoint for a job.
// Returns nil if no checkpoint exists (fresh job).
func LoadCheckpoint(ctx context.Context, db *pgxpool.Pool, jobID uuid.UUID) (*JobCheckpoint, error) {
	var cpJSON []byte
	err := db.QueryRow(ctx,
		`SELECT checkpoint_data FROM ingestion_jobs WHERE id = $1`,
		jobID,
	).Scan(&cpJSON)
	if err != nil {
		return nil, fmt.Errorf("load checkpoint: %w", err)
	}
	if len(cpJSON) == 0 || string(cpJSON) == "{}" {
		return nil, nil // fresh job, no checkpoint
	}
	var cp JobCheckpoint
	if err := json.Unmarshal(cpJSON, &cp); err != nil {
		return nil, fmt.Errorf("unmarshal checkpoint: %w", err)
	}
	return &cp, nil
}

// ProgressTracker tracks ingestion speed and estimates remaining time.
// In-memory only — written to DB every checkpointEveryN chunks.
//
// T3: the topic/embed batches now run concurrently, so the tracker is read
// from multiple goroutines (UpdateDB computes the ETA from the running total).
// A plain sync.Mutex is enough — both AddCost and TotalCost write/read one
// float, there is no read-heavy path to justify an RWMutex (same reasoning as
// observability.PhaseTimer).
type ProgressTracker struct {
	db          *pgxpool.Pool
	jobID       uuid.UUID
	totalChunks int
	startedAt   time.Time
	mu          sync.Mutex
	costUSD     float64
	logger      *zap.Logger
}

func NewProgressTracker(db *pgxpool.Pool, jobID uuid.UUID, totalChunks int, logger *zap.Logger) *ProgressTracker {
	return &ProgressTracker{
		db:          db,
		jobID:       jobID,
		totalChunks: totalChunks,
		startedAt:   time.Now(),
		logger:      logger,
	}
}

// AddCost accumulates LLM cost.
func (t *ProgressTracker) AddCost(usd float64) {
	t.mu.Lock()
	t.costUSD += usd
	t.mu.Unlock()
}

// TotalCost returns accumulated cost.
func (t *ProgressTracker) TotalCost() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.costUSD
}

// UpdateDB writes current progress + ETA to the DB.
// Called every checkpointEveryN chunks.
//
// Mental execution:
//   totalChunks=15000, done=4200, elapsed=504s
//   speed = 4200/504 = 8.33 chunks/sec
//   remaining = (15000-4200)/8.33 = 1296 seconds = ~21 min
func (t *ProgressTracker) UpdateDB(ctx context.Context, done int, stage, detail string) {
	elapsed := time.Since(t.startedAt).Seconds()
	var etaSec int
	if elapsed > 0 && done > 0 {
		speed := float64(done) / elapsed
		remaining := float64(t.totalChunks-done) / speed
		if remaining > 0 {
			etaSec = int(remaining)
		}
	}

	_, err := t.db.Exec(ctx,
		`UPDATE ingestion_jobs SET
			processed_chunks             = $1,
			total_chunks                 = $2,
			current_stage                = $3,
			stage_detail                 = $4,
			cost_usd                     = $5,
			estimated_seconds_remaining  = $6
		 WHERE id = $7`,
		done, t.totalChunks, stage, detail, t.TotalCost(), etaSec, t.jobID,
	)
	if err != nil {
		t.logger.Warn("progress update failed", zap.Error(err))
	}
}
