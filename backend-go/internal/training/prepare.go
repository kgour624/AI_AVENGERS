package training

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/docextract"
	"ai_avengers/backend/internal/jobevents"
)

// StageExtracting is the pre-pipeline stage: the uploaded file is being
// converted to text.
//
// WHY it is not one of the six pipeline stages: chunking, topic extraction,
// charter extraction, embedding, storing and the smoke test all operate on
// text and are checkpointed. Extraction happens *before* any of them, writes no
// checkpoint, and is not resumable (on failure the admin re-uploads). It is a
// real stage to the admin — it can take seconds and it can fail — so it is
// surfaced in ingestion_jobs.current_stage (migration 037) and on the event
// timeline, but it never appears in checkpoint_data.stage.
const StageExtracting = "extracting"

// PrepareTranscript converts an uploaded document into the text the pipeline
// ingests, and records the conversion on the timeline (T1).
//
// WHY this exists as a separate step rather than inside IngestTranscript:
// every stage after this one works on text, and the resume/retry paths already
// re-run from a transcript stored in the database. Converting here means those
// paths are untouched: they read the extracted text exactly as they previously
// read an uploaded .txt.
//
// Mental execution:
//   Input: job J, expert E, "lecture.pdf", 12MB
//   1. Mark the job 'running' / 'extracting' so the UI shows stage 0.
//   2. Extract (plain text is decoded in-process; everything else via sidecar).
//   3. On failure → job 'failed' + a `failed` event carrying reason and message,
//      and return the error. Nothing downstream runs.
//   4. On success → store the extracted text as transcript_content (the resume
//      source) and emit `stage_done` with format/pages/chars.
//
// Fail-closed: step 4 is NOT best-effort. Every later resume reads that column,
// so silently continuing after a failed write would produce a job that looks
// healthy and cannot be resumed.
func (p *IngestionPipeline) PrepareTranscript(
	ctx context.Context,
	jobID, expertID uuid.UUID,
	filename string,
	data []byte,
) (string, error) {
	started := time.Now()

	if _, err := p.db.Exec(ctx,
		`UPDATE ingestion_jobs
		    SET status = 'running',
		        current_stage = $2,
		        stage_detail = $3,
		        started_at = COALESCE(started_at, NOW()),
		        updated_at = NOW()
		  WHERE id = $1`,
		jobID, StageExtracting, fmt.Sprintf("Extracting text from %s...", filename),
	); err != nil {
		return "", fmt.Errorf("mark job as extracting: %w", err)
	}

	p.emit(ctx, jobID, expertID, StageExtracting, jobevents.KindStageStarted, map[string]interface{}{
		"filename": filename,
		"bytes":    len(data),
	})

	result, err := p.extractor.Extract(ctx, filename, data)
	if err != nil {
		reason, message := describeExtractError(err)
		p.logger.Warn("document extraction failed",
			zap.String("job_id", jobID.String()),
			zap.String("filename", filename),
			zap.String("reason", reason),
			zap.Error(err),
		)
		p.updateJobStatus(ctx, jobID, "failed", message, 0, 0)
		p.emitFinal(ctx, jobID, expertID, StageExtracting, jobevents.KindFailed, map[string]interface{}{
			"reason":   reason,
			"message":  message,
			"filename": filename,
		})
		return "", err
	}

	if result.Chars == 0 {
		result.Chars = len(result.Text)
	}

	// The extracted text is the durable source for every later resume/retry.
	_, writeErr := p.db.Exec(ctx,
		`UPDATE ingestion_jobs SET transcript_content = $2, updated_at = NOW() WHERE id = $1`,
		jobID, result.Text,
	)
	if writeErr != nil {
		reason := "store_failed"
		message := "Could not store the extracted text: " + writeErr.Error()
		p.logger.Error("failed to store extracted transcript",
			zap.String("job_id", jobID.String()),
			zap.Error(writeErr),
		)
		p.updateJobStatus(ctx, jobID, "failed", message, 0, 0)
		p.emitFinal(ctx, jobID, expertID, StageExtracting, jobevents.KindFailed, map[string]interface{}{
			"reason":  reason,
			"message": message,
		})
		return "", fmt.Errorf("store extracted transcript: %w", writeErr)
	}

	p.updateStage(ctx, jobID, StageExtracting,
		fmt.Sprintf("Extracted %d characters from %s", result.Chars, filename))
	p.emit(ctx, jobID, expertID, StageExtracting, jobevents.KindStageDone, map[string]interface{}{
		"duration_ms": time.Since(started).Milliseconds(),
		"filename":    filename,
		"format":      result.Format,
		"chars":       result.Chars,
		"pages":       result.Pages,
		"warnings":    result.Warnings,
	})
	p.logger.Info("document extracted",
		zap.String("job_id", jobID.String()),
		zap.String("filename", filename),
		zap.String("format", result.Format),
		zap.Int("chars", result.Chars),
		zap.Int("pages", result.Pages),
	)

	return result.Text, nil
}

// describeExtractError splits a conversion failure into the stable reason code
// (for logs/metrics) and the admin-facing message (shown verbatim in the UI).
func describeExtractError(err error) (reason string, message string) {
	var extractErr *docextract.Error
	if errors.As(err, &extractErr) {
		return extractErr.Reason, extractErr.Message
	}
	return "extract_failed", err.Error()
}
