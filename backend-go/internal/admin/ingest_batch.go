package admin

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/docextract"
	"ai_avengers/backend/internal/response"
	"ai_avengers/backend/internal/training"
)

// Batch ingestion limits.
//
// WHY bounded concurrency (not one goroutine per file): 50 uploads at once would
// open 50 PDF parses and 50 embed loops against the same provider. The pipeline
// itself already bounds its internal work with a worker semaphore; this bounds
// how many JOBS run at once on top of that, which is the pattern the Ultimate Go
// worker-pool lesson describes — a fixed number of workers, a queue of work, and
// no unbounded goroutine growth.
const (
	maxBatchFiles        = 50
	batchWorkerCount     = 3
	batchFileSizeCeiling = 50 * 1024 * 1024
)

// BatchIngestResult is the per-file outcome, so one bad PDF cannot hide the
// other 49 (and the admin can see exactly which file failed and why).
type BatchIngestResult struct {
	Filename string `json:"filename"`
	JobID    string `json:"job_id,omitempty"`
	Status   string `json:"status"` // accepted | rejected
	Reason   string `json:"reason,omitempty"`
}

// batchFile is one accepted upload waiting for a worker.
type batchFile struct {
	filename string
	jobID    uuid.UUID
	content  []byte
}

// IngestTranscriptBatch POST /admin/experts/:id/ingest-batch
//
// Multipart form with repeated "transcripts" fields. Every accepted file gets
// its own ingestion job; up to batchWorkerCount jobs run together.
//
// WHY append-only: several jobs writing the same expert's corpus at once cannot
// safely share a "replace the corpus first" step, so replace_existing is
// intentionally not honoured in batch mode. The admin still has it for a single
// upload, which is where it is actually needed (re-chunking one course).
func (h *AdminHandler) IngestTranscriptBatch(c *gin.Context) {
	expertID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid expert ID")
		return
	}

	var expertName string
	if err := h.db.QueryRow(c.Request.Context(),
		`SELECT name FROM experts WHERE id=$1 AND deleted_at IS NULL`, expertID,
	).Scan(&expertName); err != nil || expertName == "" {
		response.NotFound(c, "expert")
		return
	}

	form, err := c.MultipartForm()
	if err != nil || form == nil {
		response.BadRequest(c, "FILES_REQUIRED", "multipart form with transcript files is required")
		return
	}
	headers := form.File["transcripts"]
	if len(headers) == 0 {
		headers = form.File["transcript"] // accept the single-file field name too
	}
	if len(headers) == 0 {
		response.BadRequest(c, "FILES_REQUIRED", "at least one transcript file is required")
		return
	}
	if len(headers) > maxBatchFiles {
		response.BadRequest(c, "TOO_MANY_FILES",
			fmt.Sprintf("at most %d files per batch (received %d); split the batch and upload again", maxBatchFiles, len(headers)))
		return
	}

	results := make([]BatchIngestResult, 0, len(headers))
	accepted := make([]batchFile, 0, len(headers))
	for _, header := range headers {
		result := BatchIngestResult{Filename: header.Filename}
		switch {
		case header.Size > batchFileSizeCeiling:
			result.Status, result.Reason = "rejected", "larger than 50MB"
		case !docextract.IsSupported(header.Filename):
			result.Status, result.Reason = "rejected", fmt.Sprintf("unsupported format (.%s)", docextract.Extension(header.Filename))
		default:
			opened, openErr := header.Open()
			if openErr != nil {
				result.Status, result.Reason = "rejected", "could not read the upload"
				break
			}
			content := make([]byte, header.Size)
			_, readErr := io.ReadFull(opened, content)
			opened.Close()
			if readErr != nil {
				result.Status, result.Reason = "rejected", "upload was truncated"
				break
			}
			var jobID uuid.UUID
			if err := h.db.QueryRow(c.Request.Context(),
				`INSERT INTO ingestion_jobs (expert_id, job_type, status, source_path)
				 VALUES ($1, 'transcript', 'pending', $2) RETURNING id`,
				expertID, header.Filename,
			).Scan(&jobID); err != nil {
				result.Status, result.Reason = "rejected", "could not create the ingestion job"
				break
			}
			result.Status, result.JobID = "accepted", jobID.String()
			accepted = append(accepted, batchFile{filename: header.Filename, jobID: jobID, content: content})
		}
		results = append(results, result)
	}

	if len(accepted) == 0 {
		response.OK(c, gin.H{"results": results, "accepted": 0})
		return
	}

	_, _ = h.db.Exec(c.Request.Context(),
		`UPDATE experts SET is_training=TRUE, updated_at=NOW() WHERE id=$1`, expertID)

	// Bounded worker pool: `batchWorkerCount` workers pull from the queue.
	// Results are written through their own index, so the shared slice is only
	// written by the collector below — the race the worker lesson warns about.
	h.logger.Info("batch ingestion started",
		zap.String("expert_id", expertID.String()),
		zap.Int("accepted", len(accepted)),
		zap.Int("workers", batchWorkerCount),
	)

	queue := make(chan batchFile)
	var wg sync.WaitGroup
	for worker := 0; worker < batchWorkerCount; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for file := range queue {
				h.runBatchJob(expertID, expertName, file)
			}
		}()
	}
	for _, file := range accepted {
		queue <- file
	}
	close(queue)
	// The workers run in the background; the HTTP response reports acceptance,
	// exactly like the single-file endpoint returns the job id immediately.
	go func() { wg.Wait() }()

	response.OK(c, gin.H{"results": results, "accepted": len(accepted), "workers": batchWorkerCount})
}

// runBatchJob prepares and ingests one file. Each job gets its own timeout and
// its failure is contained to that file.
func (h *AdminHandler) runBatchJob(expertID uuid.UUID, expertName string, file batchFile) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
	defer cancel()

	transcript, prepErr := h.ingestion.PrepareTranscript(ctx, file.jobID, expertID, file.filename, file.content)
	if prepErr != nil {
		h.logger.Warn("batch ingestion: preparation failed",
			zap.String("job_id", file.jobID.String()),
			zap.String("filename", file.filename),
			zap.Error(prepErr))
		return
	}

	if _, err := h.ingestion.IngestTranscript(ctx, file.jobID, expertID, expertName, transcript, file.filename, false); err != nil {
		if errors.Is(err, training.ErrJobPaused) {
			h.logger.Info("batch ingestion: job paused for a decision",
				zap.String("job_id", file.jobID.String()), zap.String("filename", file.filename))
			return
		}
		h.logger.Warn("batch ingestion: job failed",
			zap.String("job_id", file.jobID.String()),
			zap.String("filename", file.filename),
			zap.Error(err))
		return
	}
	h.logger.Info("batch ingestion: job finished",
		zap.String("job_id", file.jobID.String()), zap.String("filename", file.filename))
}

// batchFilenameList is a small helper kept for logging/summaries without pulling
// the whole slice into a log line.
func batchFilenameList(files []batchFile) string {
	names := make([]string, 0, len(files))
	for _, f := range files {
		names = append(names, f.filename)
	}
	return strings.Join(names, ", ")
}
