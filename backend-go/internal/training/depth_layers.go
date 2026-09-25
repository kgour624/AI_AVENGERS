package training

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/gateway"
)

// Depth layers (I5, option A): what KIND of content each chunk is.
//
// The expert's declared depth was derived from how many chunks mention a topic, so
// fourteen introductory definitions scored the same as fourteen failure modes and
// trade-offs. This classifies each existing chunk instead:
//
//	1 = WHAT / WHY        a definition, a purpose, when to use the thing
//	2 = HOW / TRADE-OFFS  mechanics, comparisons, costs and benefits
//	3 = FAILURE / EDGE    what breaks, limitations, debugging, war stories
//
// Option A means NOTHING IS INVENTED. The course already contains whatever depth it
// contains; this only says which kind each passage is. That is why the result is
// trustworthy enough to show a buyer, and why a topic with no layer-3 chunks is a
// finding about the course rather than a guess about the expert.
//
// Classification is on demand and resumable — it costs model calls, so it is not
// hidden inside ingest, and a half-classified corpus reports how much is left instead
// of pretending to be finished.

const (
	// LayerDefinition, LayerMechanics and LayerFailure are the three content kinds.
	LayerDefinition = 1
	LayerMechanics  = 2
	LayerFailure    = 3

	// layerBatchSize is how many passages travel in one classification call. Small
	// enough that each answer is checkable, large enough that a 1600-chunk corpus does
	// not take 1600 calls.
	layerBatchSize = 20

	// layerCallChunkChars caps one passage in the prompt.
	layerCallChunkChars = 700

	// layerMaxChunksPerRun bounds one call to Classify, so a button press is a bounded
	// amount of work and the response can say how much is left. The caller runs it again
	// to continue.
	layerMaxChunksPerRun = 200

	// depthJobTimeout bounds a background pass. It is intentionally much longer
	// than the 30-second browser request because the request only starts the job;
	// the UI polls its durable status while model calls run independently.
	depthJobTimeout = 30 * time.Minute

	// depthJobStaleAfter is longer than the worker deadline. A row older than this
	// with status queued/running can only be from a crashed/restarted API process.
	depthJobStaleAfter = 35 * time.Minute
)

// ErrDepthClassificationRunning means this expert already has an active pass.
var ErrDepthClassificationRunning = fmt.Errorf("depth classification is already running")

// DepthLayerName renders a layer for the screen.
func DepthLayerName(layer int) string {
	switch layer {
	case LayerDefinition:
		return "what / why"
	case LayerMechanics:
		return "how / trade-offs"
	case LayerFailure:
		return "failure / edge cases"
	default:
		return "not classified"
	}
}

// DepthClassifyResult reports one classification pass.
type DepthClassifyResult struct {
	// Total is the bounded batch selected by this pass, not the whole corpus.
	Total      int `json:"total"`
	Classified int `json:"classified"`
	// Remaining is how many chunks still have no layer, so the caller knows whether to
	// run again rather than guessing.
	Remaining int `json:"remaining"`
	Calls     int `json:"calls"`
	// FailedBatches distinguishes "all classified" from "the model did not answer".
	// The latter must leave the job failed with its remaining count, so the admin can
	// retry rather than seeing a false completion.
	FailedBatches int `json:"failed_batches"`
	// Rejected counts entries the model produced that named an out-of-range passage or
	// layer. Shown, not hidden: a rising number means the prompt or the model is drifting.
	Rejected int `json:"rejected"`
}

// DepthClassifyProgress is written after every model batch, so the UI can show
// that work is advancing instead of displaying an eternal spinner.
type DepthClassifyProgress struct {
	Total      int
	Classified int
	Calls      int
	Rejected   int
}

// DepthClassificationJob is the durable status of one background pass.
type DepthClassificationJob struct {
	ID           uuid.UUID  `json:"id"`
	ExpertID     uuid.UUID  `json:"expert_id"`
	Status       string     `json:"status"`
	Total        int        `json:"total"`
	Classified   int        `json:"classified"`
	Remaining    int        `json:"remaining"`
	Calls        int        `json:"calls"`
	Rejected     int        `json:"rejected"`
	ErrorMessage string     `json:"error_message,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
}

// TopicLayerCoverage is one topic's content kinds.
type TopicLayerCoverage struct {
	Topic      string `json:"topic"`
	Definition int    `json:"definition"`
	Mechanics  int    `json:"mechanics"`
	Failure    int    `json:"failure"`
	Total      int    `json:"total"`
}

// DepthLayerReport answers "is this expert deep?" from the corpus itself.
type DepthLayerReport struct {
	ExpertID     uuid.UUID `json:"expert_id"`
	TotalChunks  int       `json:"total_chunks"`
	Classified   int       `json:"classified"`
	Unclassified int       `json:"unclassified"`
	Definition   int       `json:"definition"`
	Mechanics    int       `json:"mechanics"`
	Failure      int       `json:"failure"`
	// Topics that have no failure-mode content, and none at mechanics level. These are
	// the actionable numbers: they name where the expert can only answer "what is it".
	TopicsWithoutFailure   int                  `json:"topics_without_failure"`
	TopicsWithoutMechanics int                  `json:"topics_without_mechanics"`
	Topics                 []TopicLayerCoverage `json:"topics"`
	Findings               []string             `json:"findings"`
}

// DepthClassifier labels chunks by content kind and reports coverage.
type DepthClassifier struct {
	db      *pgxpool.Pool
	gateway *gateway.ModelGateway
	logger  *zap.Logger
}

// NewDepthClassifier builds a depth classifier.
func NewDepthClassifier(db *pgxpool.Pool, gw *gateway.ModelGateway, logger *zap.Logger) *DepthClassifier {
	return &DepthClassifier{db: db, gateway: gw, logger: logger}
}

// Classify labels up to layerMaxChunksPerRun unclassified chunks, then reports how many
// are still left. Resumable by design: a second call continues where this one stopped.
func (c *DepthClassifier) Classify(ctx context.Context, expertID uuid.UUID) (DepthClassifyResult, error) {
	return c.ClassifyWithProgress(ctx, expertID, nil)
}

// ClassifyWithProgress is Classify with an optional progress callback. The
// callback runs after each model batch, never while a DB row set is open. A callback
// failure is returned: if durable progress cannot be written, the job must not
// pretend it is tracking the work.
func (c *DepthClassifier) ClassifyWithProgress(
	ctx context.Context,
	expertID uuid.UUID,
	onProgress func(DepthClassifyProgress) error,
) (DepthClassifyResult, error) {
	var result DepthClassifyResult

	type pending struct {
		ID   uuid.UUID
		Text string
	}
	rows, err := c.db.Query(ctx, `
		SELECT id, chunk_text
		  FROM course_chunks
		 WHERE expert_id = $1 AND layer IS NULL
		 ORDER BY chunk_index
		 LIMIT $2`, expertID, layerMaxChunksPerRun)
	if err != nil {
		return result, fmt.Errorf("depth layers: read unclassified chunks: %w", err)
	}
	var todo []pending
	for rows.Next() {
		var p pending
		if scanErr := rows.Scan(&p.ID, &p.Text); scanErr != nil {
			continue
		}
		todo = append(todo, p)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return result, fmt.Errorf("depth layers: read unclassified chunks: %w", err)
	}
	result.Total = len(todo)

	for start := 0; start < len(todo); start += layerBatchSize {
		end := start + layerBatchSize
		if end > len(todo) {
			end = len(todo)
		}
		batch := todo[start:end]

		texts := make([]string, 0, len(batch))
		for _, p := range batch {
			texts = append(texts, p.Text)
		}

		response, callErr := c.requestLayers(ctx, texts)
		result.Calls++
		if callErr != nil {
			// One failed batch must not discard the batches that worked.
			c.logger.Warn("depth layers: classification call failed",
				zap.String("expert_id", expertID.String()),
				zap.Error(callErr))
			result.FailedBatches++
			if onProgress != nil {
				if progressErr := onProgress(DepthClassifyProgress{
					Total: len(todo), Classified: result.Classified,
					Calls: result.Calls, Rejected: result.Rejected,
				}); progressErr != nil {
					return result, fmt.Errorf("depth layers: persist progress after failed batch: %w", progressErr)
				}
			}
			continue
		}

		layers, rejected, parseErr := parseDepthLayers(response, len(batch))
		if parseErr != nil {
			c.logger.Warn("depth layers: unparseable response",
				zap.String("expert_id", expertID.String()),
				zap.Error(parseErr))
			result.FailedBatches++
			if onProgress != nil {
				if progressErr := onProgress(DepthClassifyProgress{
					Total: len(todo), Classified: result.Classified,
					Calls: result.Calls, Rejected: result.Rejected,
				}); progressErr != nil {
					return result, fmt.Errorf("depth layers: persist progress after invalid response: %w", progressErr)
				}
			}
			continue
		}
		result.Rejected += rejected

		for index, layer := range layers {
			if _, err := c.db.Exec(ctx,
				`UPDATE course_chunks SET layer = $1 WHERE id = $2`, layer, batch[index].ID); err != nil {
				return result, fmt.Errorf("depth layers: store layer: %w", err)
			}
			result.Classified++
		}
		if onProgress != nil {
			if progressErr := onProgress(DepthClassifyProgress{
				Total: len(todo), Classified: result.Classified,
				Calls: result.Calls, Rejected: result.Rejected,
			}); progressErr != nil {
				return result, fmt.Errorf("depth layers: persist progress: %w", progressErr)
			}
		}
	}

	if err := c.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM course_chunks WHERE expert_id = $1 AND layer IS NULL`, expertID,
	).Scan(&result.Remaining); err != nil {
		return result, fmt.Errorf("depth layers: count remaining: %w", err)
	}
	if result.FailedBatches > 0 {
		return result, fmt.Errorf("%d model batch(es) failed; %d chunks remain unclassified; press Classify more to retry", result.FailedBatches, result.Remaining)
	}

	c.logger.Info("depth layers classified",
		zap.String("expert_id", expertID.String()),
		zap.Int("classified", result.Classified),
		zap.Int("remaining", result.Remaining),
		zap.Int("calls", result.Calls),
		zap.Int("rejected", result.Rejected),
	)
	return result, nil
}

// StartBackground creates a durable queued job and starts classification on a
// detached, bounded context. The HTTP request only creates the job row; when that
// request returns (or the browser times out), the worker keeps running.
func (c *DepthClassifier) StartBackground(ctx context.Context, expertID uuid.UUID) (*DepthClassificationJob, error) {
	// Recover a job left behind by a crashed API process. The staleness threshold is
	// longer than the worker deadline, so a live worker is never reclaimed just
	// because one model call is slow.
	if err := c.recoverStaleJob(ctx, expertID); err != nil {
		return nil, err
	}

	var pending int
	if err := c.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM course_chunks WHERE expert_id = $1 AND layer IS NULL`, expertID,
	).Scan(&pending); err != nil {
		return nil, fmt.Errorf("depth layers: count pending chunks: %w", err)
	}
	batchTotal := pending
	if batchTotal > layerMaxChunksPerRun {
		batchTotal = layerMaxChunksPerRun
	}

	job := &DepthClassificationJob{ExpertID: expertID, Status: "queued"}
	err := c.db.QueryRow(ctx, `
		INSERT INTO expert_depth_classification_jobs (expert_id, status, total, remaining)
		VALUES ($1, 'queued', $2, $3)
		ON CONFLICT (expert_id) WHERE status IN ('queued','running') DO NOTHING
		RETURNING id, expert_id, status, total, classified, remaining, calls, rejected,
		          error_message, created_at, updated_at, completed_at`, expertID, batchTotal, pending,
	).Scan(&job.ID, &job.ExpertID, &job.Status, &job.Total, &job.Classified, &job.Remaining,
		&job.Calls, &job.Rejected, &job.ErrorMessage, &job.CreatedAt, &job.UpdatedAt, &job.CompletedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrDepthClassificationRunning
		}
		return nil, fmt.Errorf("depth layers: create job: %w", err)
	}

	// Copy values — never carry the request context into the worker. Its cancellation
	// when the handler returns is exactly the 30-second timeout this background job
	// exists to avoid.
	jobID := job.ID
	go c.runBackground(jobID, expertID)
	return job, nil
}

func (c *DepthClassifier) runBackground(jobID, expertID uuid.UUID) {
	ctx, cancel := context.WithTimeout(context.Background(), depthJobTimeout)
	defer cancel()

	if _, err := c.db.Exec(ctx, `
		UPDATE expert_depth_classification_jobs
		   SET status = 'running', updated_at = now()
		 WHERE id = $1 AND status = 'queued'`, jobID); err != nil {
		c.logger.Error("depth layers: could not mark job running",
			zap.String("job_id", jobID.String()), zap.Error(err))
		c.finishBackground(jobID, "failed", err.Error())
		return
	}

	result, err := c.ClassifyWithProgress(ctx, expertID, func(p DepthClassifyProgress) error {
		remaining := pendingCount(ctx, c.db, expertID)
		// pendingCount can fail only while the database is unavailable. Return that
		// error so the job fails visibly rather than showing a made-up progress value.
		if remaining < 0 {
			return fmt.Errorf("depth layers: count remaining while updating job")
		}
		_, updateErr := c.db.Exec(ctx, `
			UPDATE expert_depth_classification_jobs
			   SET total = $1, classified = $2, remaining = $3, calls = $4, rejected = $5,
			       updated_at = now()
			 WHERE id = $6 AND status = 'running'`,
			p.Total, p.Classified, remaining, p.Calls, p.Rejected, jobID)
		return updateErr
	})
	if err != nil {
		c.finishBackgroundWithResult(jobID, result, err.Error())
		return
	}

	status := "complete"
	message := ""
	if result.Classified == 0 && result.Remaining > 0 {
		// The model calls all failed or returned unusable output. Do not label an
		// empty pass "complete"; the remaining count lets the admin retry.
		status = "failed"
		message = "no chunks were classified; check the model response/logs and retry"
	}
	if _, err := c.db.Exec(ctx, `
		UPDATE expert_depth_classification_jobs
		   SET status = $1, total = $2, classified = $3, remaining = $4,
		       calls = $5, rejected = $6, error_message = $7,
	       completed_at = now(), updated_at = now()
		 WHERE id = $8`, status, result.Total, result.Classified, result.Remaining,
		result.Calls, result.Rejected, message, jobID); err != nil {
		c.logger.Error("depth layers: could not finish job row",
			zap.String("job_id", jobID.String()), zap.Error(err))
	}
}

// pendingCount returns -1 on a DB error so a progress callback can fail the job
// visibly rather than silently displaying an invented remainder.
func pendingCount(ctx context.Context, db interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, expertID uuid.UUID) int {
	var count int
	if err := db.QueryRow(ctx,
		`SELECT COUNT(*) FROM course_chunks WHERE expert_id = $1 AND layer IS NULL`, expertID,
	).Scan(&count); err != nil {
		return -1
	}
	return count
}

func (c *DepthClassifier) finishBackgroundWithResult(jobID uuid.UUID, result DepthClassifyResult, message string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// ClassifyWithProgress can time out before its final COUNT query. On failure, read
	// the real remainder here; a timed-out job that shows "0 left" would hide the work
	// the Retry button needs to resume.
	var expertID uuid.UUID
	if err := c.db.QueryRow(ctx,
		`SELECT expert_id FROM expert_depth_classification_jobs WHERE id = $1`, jobID,
	).Scan(&expertID); err == nil {
		if remaining := pendingCount(ctx, c.db, expertID); remaining >= 0 {
			result.Remaining = remaining
		}
	}
	if _, err := c.db.Exec(ctx, `
		UPDATE expert_depth_classification_jobs
		   SET status = 'failed',
		       total = CASE WHEN $1 > 0 THEN GREATEST(total, $1) ELSE total END,
		       classified = GREATEST(classified, $2), remaining = $3,
		       calls = $4, rejected = $5, error_message = $6,
		       completed_at = now(), updated_at = now()
		 WHERE id = $7`,
		result.Total, result.Classified, result.Remaining,
		result.Calls, result.Rejected, message, jobID); err != nil {
		c.logger.Error("depth layers: could not record partial job failure",
			zap.String("job_id", jobID.String()), zap.Error(err))
	}
}

func (c *DepthClassifier) finishBackground(jobID uuid.UUID, status, message string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := c.db.Exec(ctx, `
		UPDATE expert_depth_classification_jobs
		   SET status = $1, error_message = $2, completed_at = now(), updated_at = now()
		 WHERE id = $3`, status, message, jobID); err != nil {
		c.logger.Error("depth layers: could not record job failure",
			zap.String("job_id", jobID.String()), zap.Error(err))
	}
}

// LatestJob returns the most recent durable classification job, or nil when none
// has ever been started.
func (c *DepthClassifier) LatestJob(ctx context.Context, expertID uuid.UUID) (*DepthClassificationJob, error) {
	// A worker from a crashed API process can be left queued/running forever. Expire
	// only after the worker's full timeout plus five minutes; a genuinely slow worker
	// must not be marked failed while it can still finish.
	if err := c.recoverStaleJob(ctx, expertID); err != nil {
		return nil, err
	}

	job := &DepthClassificationJob{}
	err := c.db.QueryRow(ctx, `
		SELECT id, expert_id, status, total, classified, remaining, calls, rejected,
		       error_message, created_at, updated_at, completed_at
		  FROM expert_depth_classification_jobs
		 WHERE expert_id = $1
		 ORDER BY created_at DESC LIMIT 1`, expertID,
	).Scan(&job.ID, &job.ExpertID, &job.Status, &job.Total, &job.Classified, &job.Remaining,
		&job.Calls, &job.Rejected, &job.ErrorMessage, &job.CreatedAt, &job.UpdatedAt, &job.CompletedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("depth layers: read latest job: %w", err)
	}
	return job, nil
}

// recoverStaleJob marks jobs abandoned by an API-process restart as failed. The
// deadline is deliberately longer than the worker's timeout, so a slow but alive
// worker is not reclaimed. Called on both start and read, otherwise a page refresh
// after restart could show a permanently-running job and disable the only retry action.
func (c *DepthClassifier) recoverStaleJob(ctx context.Context, expertID uuid.UUID) error {
	if _, err := c.db.Exec(ctx, `
		UPDATE expert_depth_classification_jobs
		   SET status = 'failed',
		       error_message = 'worker stopped before finishing; press Classify more to resume',
		       completed_at = now(), updated_at = now()
		 WHERE expert_id = $1 AND status IN ('queued','running')
		   AND updated_at < now() - ($2 * interval '1 minute')`,
		expertID, int(depthJobStaleAfter.Minutes())); err != nil {
		return fmt.Errorf("depth layers: recover stale job: %w", err)
	}
	return nil
}

// requestLayers asks the model what kind of content each passage is.
func (c *DepthClassifier) requestLayers(ctx context.Context, texts []string) (string, error) {
	var sb strings.Builder
	sb.WriteString("For each numbered passage from a technical course, decide what KIND of content it is.\n\n")
	sb.WriteString(`1 = WHAT / WHY       — explains what something is, why it matters, or when to use it
2 = HOW / TRADE-OFFS — explains how it works, compares options, or weighs costs and benefits
3 = FAILURE / EDGE   — explains what breaks, limitations, debugging, or a real incident

If a passage mixes kinds, choose the dominant one. Judge only the passage.

Passages:
`)
	for i, text := range texts {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, clipPromptText(strings.ReplaceAll(text, "\n", " "), layerCallChunkChars)))
	}
	sb.WriteString(`
Return ONLY JSON, one entry per passage:
{"layers":[{"n":1,"layer":2}]}`)

	resp, err := c.gateway.Call(ctx, gateway.LLMRequest{
		Model:       gateway.ModelCheap,
		UserPrompt:  sb.String(),
		MaxTokens:   2048,
		Temperature: 0,
	})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

// parseDepthLayers reads the model's answer into 0-based indexes.
//
// Pure — unit-tested. Entries naming a passage outside the batch are DROPPED, not
// clamped: a layer applied to the wrong chunk is worse than a chunk left unclassified,
// because an unclassified chunk is visibly pending while a mislabelled one is silently
// wrong for as long as the corpus lives.
func parseDepthLayers(response string, count int) (map[int]int, int, error) {
	body := extractJSONObject(response)
	if body == "" {
		return nil, 0, fmt.Errorf("no JSON object in response")
	}

	var payload struct {
		Layers []struct {
			N     int `json:"n"`
			Layer int `json:"layer"`
		} `json:"layers"`
	}
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		return nil, 0, err
	}

	rejected := 0
	out := make(map[int]int, len(payload.Layers))
	for _, entry := range payload.Layers {
		index := entry.N - 1
		if index < 0 || index >= count {
			rejected++
			continue
		}
		if entry.Layer < LayerDefinition || entry.Layer > LayerFailure {
			rejected++
			continue
		}
		if _, dup := out[index]; dup {
			// The model repeating itself is not an invention; keep the first answer.
			continue
		}
		out[index] = entry.Layer
	}
	return out, rejected, nil
}

// Report reads the corpus's content-kind coverage.
func (c *DepthClassifier) Report(ctx context.Context, expertID uuid.UUID) (*DepthLayerReport, error) {
	report := &DepthLayerReport{ExpertID: expertID, Topics: []TopicLayerCoverage{}, Findings: []string{}}

	if err := c.db.QueryRow(ctx, `
		SELECT COUNT(*),
		       COUNT(*) FILTER (WHERE layer IS NULL),
		       COUNT(*) FILTER (WHERE layer = $2),
		       COUNT(*) FILTER (WHERE layer = $3),
		       COUNT(*) FILTER (WHERE layer = $4)
		  FROM course_chunks
		 WHERE expert_id = $1`, expertID, LayerDefinition, LayerMechanics, LayerFailure,
	).Scan(&report.TotalChunks, &report.Unclassified, &report.Definition, &report.Mechanics, &report.Failure); err != nil {
		return nil, fmt.Errorf("depth layers: read coverage: %w", err)
	}
	report.Classified = report.TotalChunks - report.Unclassified

	rows, err := c.db.Query(ctx, `
		SELECT topic,
		       COUNT(*) FILTER (WHERE layer = $2),
		       COUNT(*) FILTER (WHERE layer = $3),
		       COUNT(*) FILTER (WHERE layer = $4),
		       COUNT(*)
		  FROM course_chunks
		 WHERE expert_id = $1
		   AND topic IS NOT NULL AND topic <> '' AND topic <> 'general'
		 GROUP BY topic
		 ORDER BY COUNT(*) DESC`, expertID, LayerDefinition, LayerMechanics, LayerFailure)
	if err != nil {
		return nil, fmt.Errorf("depth layers: read topic coverage: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var t TopicLayerCoverage
		if scanErr := rows.Scan(&t.Topic, &t.Definition, &t.Mechanics, &t.Failure, &t.Total); scanErr != nil {
			continue
		}
		if t.Failure == 0 {
			report.TopicsWithoutFailure++
		}
		if t.Mechanics == 0 {
			report.TopicsWithoutMechanics++
		}
		report.Topics = append(report.Topics, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("depth layers: read topic coverage: %w", err)
	}

	report.Findings = layerFindings(report)
	return report, nil
}

// layerFindings turns the coverage numbers into sentences, in the order worth acting on.
//
// Pure — unit-tested. The point of the whole feature is that these are statements about
// the COURSE ("this topic has no failure-mode content"), not a score someone has to
// interpret.
func layerFindings(r *DepthLayerReport) []string {
	findings := make([]string, 0, 4)

	if r.TotalChunks == 0 {
		return []string{"No chunks stored for this expert yet."}
	}
	if r.Unclassified > 0 {
		findings = append(findings, fmt.Sprintf(
			"%d of %d chunks are not classified yet — run the classification to see the full picture.",
			r.Unclassified, r.TotalChunks))
	}
	if r.Classified == 0 {
		return findings
	}

	if r.TopicsWithoutFailure > 0 {
		findings = append(findings, fmt.Sprintf(
			"%d topic(s) have NO failure-mode or edge-case content — the expert will explain them but cannot answer \"what breaks under load\".",
			r.TopicsWithoutFailure))
	}
	if r.TopicsWithoutMechanics > 0 {
		findings = append(findings, fmt.Sprintf(
			"%d topic(s) have no mechanics content — they will answer \"what is it\" but not \"how does it work\".",
			r.TopicsWithoutMechanics))
	}

	definitionShare := r.Definition * 100 / r.Classified
	if definitionShare >= 60 {
		findings = append(findings, fmt.Sprintf(
			"%d%% of classified content is definitions (what/why). The corpus is broad but shallow: mostly explaining what things are.",
			definitionShare))
	}
	if r.Failure > 0 && r.Failure*100/r.Classified <= 5 {
		findings = append(findings, fmt.Sprintf(
			"Only %d%% of content covers failure modes and edge cases — the part of a course that makes an expert useful in an incident.",
			r.Failure*100/r.Classified))
	}
	if len(findings) == 0 {
		findings = append(findings, "Content kinds look balanced across the corpus.")
	}
	return findings
}
