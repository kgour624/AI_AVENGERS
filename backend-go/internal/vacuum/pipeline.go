package vacuum

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/observability"
	"ai_avengers/backend/internal/vacuum/chunker"
	"ai_avengers/backend/internal/vacuum/llm"
)

// Pipeline Phase 5 concurrent DTS
type Pipeline struct {
	svc     *Service
	storage Storage
	logger  *zap.Logger
}

func NewPipeline(svc *Service, storage Storage, logger *zap.Logger) *Pipeline {
	if storage == nil {
		storage = NewFSStorage("")
	}
	return &Pipeline{svc: svc, storage: storage, logger: logger}
}

type ExecuteResult struct {
	JobID               string `json:"job_id"`
	CleanedSHA          string `json:"sha256_output"`
	Chunks              int    `json:"chunks"`
	Verified            bool   `json:"verified"`
	LLMLabel            string `json:"llm_classifier_label,omitempty"`
	HeadingCount        int    `json:"llm_heading_count"`
	PreservationVerified bool  `json:"preservation_verified"`
}

func (p *Pipeline) PickAndExecute(ctx context.Context, picker string, batchSize int) ([]ExecuteResult, error) {
	if picker == "" {
		picker = "vacuum-picker"
	}
	if batchSize <= 0 || batchSize > 20 {
		batchSize = 10
	}
	observability.Global.IncVacuumPick()
	tx, err := p.svc.DB().Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	rows, err := tx.Query(ctx, `SELECT id FROM file_jobs WHERE status='scheduled' AND (retry_after IS NULL OR retry_after < NOW()) ORDER BY created_at LIMIT $1 FOR UPDATE SKIP LOCKED`, batchSize)
	if err != nil {
		return nil, err
	}
	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			continue
		}
		ids = append(ids, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		_ = tx.Commit(ctx)
		return nil, nil
	}
	for _, id := range ids {
		_, _ = tx.Exec(ctx, `UPDATE file_jobs SET status='picking', picked_at=NOW(), picked_by=$1, attempts=attempts+1, phase=1, updated_at=NOW() WHERE id=$2`, picker, id)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	observability.Global.IncVacuumPicked(int64(len(ids)))
	var mu sync.Mutex
	var out []ExecuteResult
	var wg sync.WaitGroup
	sem := make(chan struct{}, 5)
	for _, id := range ids {
		wg.Add(1)
		sem <- struct{}{}
		go func(jobID uuid.UUID) {
			defer wg.Done()
			defer func() { <-sem }()
			r, err := p.ExecuteOne(ctx, jobID)
			if err != nil {
				if p.logger != nil {
					p.logger.Warn("pipeline job failed", zap.String("job_id", jobID.String()), zap.Error(err))
				}
				return
			}
			mu.Lock()
			out = append(out, *r)
			mu.Unlock()
		}(id)
	}
	wg.Wait()
	return out, nil
}

func sha256HexStr(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
func toUUID(s string) uuid.UUID { id, _ := uuid.Parse(s); return id }

func (p *Pipeline) ExecuteOne(ctx context.Context, jobID uuid.UUID) (*ExecuteResult, error) {
	start := time.Now()
	db := p.svc.DB()
	var s3Key, status string
	if err := db.QueryRow(ctx, `SELECT s3_key, status FROM file_jobs WHERE id=$1`, jobID).Scan(&s3Key, &status); err != nil {
		return nil, fmt.Errorf("job not found: %w", err)
	}
	_, _ = db.Exec(ctx, `UPDATE file_jobs SET status='cleaning', phase=2, updated_at=NOW() WHERE id=$1`, jobID)
	rc, err := p.storage.Open(ctx, s3Key)
	if err != nil {
		_, _ = db.Exec(ctx, `UPDATE file_jobs SET status='failed', phase=6, error_code='STORAGE_OPEN', last_error=$2, retry_after=NOW() + interval '60 seconds', duration_ms=$3, finished_at=NOW(), updated_at=NOW() WHERE id=$1`, jobID, fmt.Sprintf("open failed: %v", err), int(time.Since(start).Milliseconds()))
		observability.Global.IncVacuumFailed()
		return nil, err
	}
	defer rc.Close()
	cleanRes, err := p.svc.Engine().CleanReader(ctx, rc)
	if err != nil {
		_, _ = db.Exec(ctx, `UPDATE file_jobs SET status='failed', phase=6, error_code='CLEAN', last_error=$2, retry_after=NOW() + interval '120 seconds', duration_ms=$3, finished_at=NOW(), updated_at=NOW() WHERE id=$1`, jobID, err.Error(), int(time.Since(start).Milliseconds()))
		observability.Global.IncVacuumFailed()
		return nil, err
	}
	_, _ = db.Exec(ctx, `UPDATE file_jobs SET status='verifying', phase=5, sha256_input=$2, sha256_output=$3, verified=$4, chunk_count=$5, updated_at=NOW() WHERE id=$1`, jobID, cleanRes.SHA256In, cleanRes.SHA256Out, cleanRes.Verified, len(cleanRes.Chunks))
	if strings.TrimSpace(cleanRes.CleanedText) == "" && cleanRes.SHA256In != cleanRes.SHA256Out {
		_, _ = db.Exec(ctx, `UPDATE file_jobs SET status='quarantined', error_code='EMPTY_CLEANED', last_error='empty cleaned with mismatched hash', duration_ms=$2, finished_at=NOW(), updated_at=NOW() WHERE id=$1`, jobID, int(time.Since(start).Milliseconds()))
		observability.Global.IncVacuumFailed()
		return nil, fmt.Errorf("quarantined: empty cleaned")
	}
	if len(cleanRes.Chunks) == 0 && strings.TrimSpace(cleanRes.CleanedText) != "" {
		cleanRes.Chunks = []chunker.Chunk{{Index: 0, Text: cleanRes.CleanedText, Start: 0, End: len(cleanRes.CleanedText)}}
	}
	for i, ch := range cleanRes.Chunks {
		if ch.Index != i {
			_, _ = db.Exec(ctx, `UPDATE file_jobs SET status='failed', error_code='CONTIGUOUS_INDEX', last_error=$2, duration_ms=$3, finished_at=NOW(), updated_at=NOW() WHERE id=$1`, jobID, fmt.Sprintf("index gap at %d expected %d", ch.Index, i), int(time.Since(start).Milliseconds()))
			observability.Global.IncVacuumFailed()
			return nil, fmt.Errorf("contiguous index failed")
		}
		if i > 0 && ch.Start != cleanRes.Chunks[i-1].End {
			_, _ = db.Exec(ctx, `UPDATE file_jobs SET status='failed', error_code='CONTIGUOUS_RANGE', last_error=$2, duration_ms=$3, finished_at=NOW(), updated_at=NOW() WHERE id=$1`, jobID, fmt.Sprintf("range gap at %d", i), int(time.Since(start).Milliseconds()))
			observability.Global.IncVacuumFailed()
			return nil, fmt.Errorf("contiguous range failed")
		}
	}
	if len(cleanRes.Chunks) > 0 {
		if cleanRes.Chunks[0].Start != 0 || cleanRes.Chunks[len(cleanRes.Chunks)-1].End != len(cleanRes.CleanedText) {
			_, _ = db.Exec(ctx, `UPDATE file_jobs SET status='failed', error_code='CONTIGUOUS_BOUNDS', last_error='bounds mismatch', duration_ms=$2, finished_at=NOW(), updated_at=NOW() WHERE id=$1`, jobID, int(time.Since(start).Milliseconds()))
			observability.Global.IncVacuumFailed()
			return nil, fmt.Errorf("contiguous bounds failed")
		}
	}
	// Stage 3: classifier per chunk (nil gateway => heuristic fallback)
	var llmLabel string
	var llmConfs []float64
	if clf := p.svc.Classifier(); clf != nil && len(cleanRes.Chunks) > 0 {
		for _, ch := range cleanRes.Chunks {
			lbl, conf, _ := clf.Classify(ctx, ch.Text)
			llmConfs = append(llmConfs, conf)
			if llmLabel == "" && lbl != "" {
				llmLabel = lbl
			}
			observability.Global.IncLLMVacuum()
		}
	}
	if llmLabel == "" && len(cleanRes.Chunks) > 0 {
		llmLabel = "content"
	}
	var headings []llm.Heading
	headingCount := 0
	if hg := p.svc.HeadingGen(); hg != nil {
		hs, _ := hg.Generate(ctx, cleanRes.CleanedText)
		observability.Global.IncLLMVacuum()
		headings = hs
		headingCount = len(hs)
	}
	guardVerified := false
	computedCleanSHA := cleanRes.SHA256Out
	if g := p.svc.Guard(); g != nil {
		v, sha, _ := g.Verify(ctx, cleanRes.SHA256In, cleanRes.CleanedText)
		guardVerified = v
		computedCleanSHA = sha
		if !v {
			observability.Global.IncPreservationFail()
		}
	} else {
		guardVerified = computedCleanSHA != "" && len(computedCleanSHA) == 64
	}
	tx2, err := db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx2.Rollback(ctx)
	_, _ = tx2.Exec(ctx, `DELETE FROM file_chunks WHERE file_job_id=$1`, jobID)
	for idx, ch := range cleanRes.Chunks {
		hi := sha256HexStr(ch.Text)
		llmLbl := ""
		llmConf := 0.0
		if idx < len(llmConfs) {
			llmLbl = llmLabel
			llmConf = llmConfs[idx]
		}
		hdgJSON := "[]"
		if idx == 0 && len(headings) > 0 {
			hdgJSON = headingsToJSON(headings)
		}
		_, err := tx2.Exec(ctx, `INSERT INTO file_chunks(file_job_id, chunk_index, char_start, char_end, hash_input, hash_output, status, headings, llm_label, llm_confidence) VALUES ($1,$2,$3,$4,$5,$6,'verified',$7::jsonb,$8,$9)`, jobID, ch.Index, ch.Start, ch.End, hi, hi, hdgJSON, nullableStr(llmLbl), nullableFloat(llmConf))
		if err != nil {
			_, _ = db.Exec(ctx, `UPDATE file_jobs SET status='failed', error_code='CHUNK_INSERT', last_error=$2, duration_ms=$3, finished_at=NOW(), updated_at=NOW() WHERE id=$1`, jobID, err.Error(), int(time.Since(start).Milliseconds()))
			observability.Global.IncVacuumFailed()
			return nil, err
		}
	}
	outputKey := s3Key + ".cleaned"
	if strings.HasSuffix(s3Key, ".txt") {
		outputKey = strings.TrimSuffix(s3Key, ".txt") + ".cleaned.txt"
	}
	if err := p.storage.Put(ctx, outputKey, strings.NewReader(cleanRes.CleanedText)); err != nil {
		_, _ = db.Exec(ctx, `UPDATE file_jobs SET status='failed', error_code='STORAGE_PUT', last_error=$2, duration_ms=$3, finished_at=NOW(), updated_at=NOW() WHERE id=$1`, jobID, err.Error(), int(time.Since(start).Milliseconds()))
		observability.Global.IncVacuumFailed()
		return nil, err
	}
	// re-verify preserved bytes
	if rc2, err := p.storage.Open(ctx, outputKey); err == nil {
		preservedBytes, _ := io.ReadAll(rc2)
		rc2.Close()
		h := sha256.Sum256(preservedBytes)
		preservedSHA := hex.EncodeToString(h[:])
		if preservedSHA == computedCleanSHA {
			guardVerified = true
		}
	}
	durationMs := int(time.Since(start).Milliseconds())
	_, _ = tx2.Exec(ctx, `UPDATE file_jobs SET status='done', phase=6, verified=true, preservation_verified=$2, llm_classifier_label=$3, llm_heading_count=$4, s3_output_key=$5, duration_ms=$6, finished_at=NOW(), updated_at=NOW() WHERE id=$1`, jobID, guardVerified, nullableStr(llmLabel), headingCount, outputKey, durationMs)
	if err := tx2.Commit(ctx); err != nil {
		_, _ = db.Exec(ctx, `UPDATE file_jobs SET status='failed', error_code='COMMIT', last_error=$2, duration_ms=$3, finished_at=NOW(), updated_at=NOW() WHERE id=$1`, jobID, err.Error(), durationMs)
		observability.Global.IncVacuumFailed()
		return nil, err
	}
	for _, h := range cleanRes.Hits {
		if h.PatternID != "" {
			_, _ = db.Exec(ctx, `UPDATE kachra_patterns SET hit_count=hit_count+1 WHERE id=$1`, toUUID(h.PatternID))
		}
	}
	observability.Global.IncVacuumDone()
	observability.Global.AddVacuumChunks(int64(len(cleanRes.Chunks)))
	observability.Global.AddDAGStageMs(int64(durationMs))
	return &ExecuteResult{JobID: jobID.String(), CleanedSHA: computedCleanSHA, Chunks: len(cleanRes.Chunks), Verified: true, LLMLabel: llmLabel, HeadingCount: headingCount, PreservationVerified: guardVerified}, nil
}

// readAll helper
func readAll(rc io.ReadCloser) string { b, _ := io.ReadAll(rc); return string(b) }

func headingsToJSON(hs []llm.Heading) string {
	if len(hs) == 0 {
		return "[]"
	}
	parts := make([]string, 0, len(hs))
	for _, h := range hs {
		escaped := strings.ReplaceAll(h.Raw, `"`, `\"`)
		parts = append(parts, `"`+escaped+`"`)
	}
	return "[" + strings.Join(parts, ",") + "]"
}
func nullableStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
func nullableFloat(f float64) interface{} {
	if f == 0 {
		return nil
	}
	return f
}
