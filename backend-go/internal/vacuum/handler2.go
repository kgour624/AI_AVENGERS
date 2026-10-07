package vacuum

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/observability"
	"ai_avengers/backend/internal/response"
)

func (h *Handler) ApproveCandidate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid id")
		return
	}
	var pattern, ptype, category string
	err = h.db.QueryRow(c.Request.Context(), `SELECT pattern, pattern_type, category FROM candidate_kachra WHERE id=$1`, id).Scan(&pattern, &ptype, &category)
	if err != nil {
		response.NotFound(c, "candidate")
		return
	}
	_, _ = h.db.Exec(c.Request.Context(), `INSERT INTO kachra_patterns(pattern, pattern_type, category) VALUES ($1,$2,$3) ON CONFLICT (lower(trim(pattern)), pattern_type, category) DO UPDATE SET hit_count=kachra_patterns.hit_count+1, updated_at=NOW()`, pattern, ptype, category)
	_, _ = h.db.Exec(c.Request.Context(), `UPDATE candidate_kachra SET status='approved', updated_at=NOW() WHERE id=$1`, id)
	go func() { _ = h.service.Brain().Reload(context.Background()) }()
	response.OK(c, gin.H{"status": "approved", "pattern": pattern})
}

func (h *Handler) RejectCandidate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid id")
		return
	}
	_, _ = h.db.Exec(c.Request.Context(), `UPDATE candidate_kachra SET status='rejected', updated_at=NOW() WHERE id=$1`, id)
	response.OK(c, gin.H{"status": "rejected"})
}

// BulkApproveCandidates — Phase 7 self-learning: pending -> approved -> kachra_patterns -> brain_version++ -> hot-reload
func (h *Handler) BulkApproveCandidates(c *gin.Context) {
	var req struct {
		IDs []string `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}
	if len(req.IDs) == 0 {
		response.BadRequest(c, "EMPTY", "ids required")
		return
	}
	ids := make([]uuid.UUID, 0, len(req.IDs))
	for _, s := range req.IDs {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		u, err := uuid.Parse(s)
		if err != nil {
			continue
		}
		ids = append(ids, u)
	}
	if len(ids) == 0 {
		response.BadRequest(c, "INVALID_IDS", "no valid ids")
		return
	}
	approved := 0
	for _, id := range ids {
		var pattern, ptype, category string
		err := h.db.QueryRow(c.Request.Context(), `SELECT pattern, pattern_type, category FROM candidate_kachra WHERE id=$1 AND status='pending'`, id).Scan(&pattern, &ptype, &category)
		if err != nil {
			continue
		}
		_, _ = h.db.Exec(c.Request.Context(), `INSERT INTO kachra_patterns(pattern, pattern_type, category) VALUES ($1,$2,$3) ON CONFLICT (lower(trim(pattern)), pattern_type, category) DO UPDATE SET hit_count=kachra_patterns.hit_count+1, updated_at=NOW()`, pattern, ptype, category)
		tag, _ := h.db.Exec(c.Request.Context(), `UPDATE candidate_kachra SET status='approved', updated_at=NOW() WHERE id=$1`, id)
		if tag.RowsAffected() > 0 {
			approved++
		}
	}
	if approved > 0 {
		go func() { _ = h.service.Brain().Reload(context.Background()) }()
		h.logger.Info("bulk approve", zap.Int("requested", len(ids)), zap.Int("approved", approved))
	}
	response.OK(c, gin.H{"approved": approved, "requested": len(ids), "brain_version": h.service.Brain().Version()})
}

// BulkRejectCandidates — pending -> rejected (no brain bump)
func (h *Handler) BulkRejectCandidates(c *gin.Context) {
	var req struct {
		IDs []string `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}
	if len(req.IDs) == 0 {
		response.BadRequest(c, "EMPTY", "ids required")
		return
	}
	ids := make([]uuid.UUID, 0, len(req.IDs))
	for _, s := range req.IDs {
		u, err := uuid.Parse(strings.TrimSpace(s))
		if err == nil {
			ids = append(ids, u)
		}
	}
	if len(ids) == 0 {
		response.BadRequest(c, "INVALID_IDS", "no valid ids")
		return
	}
	rejected := 0
	for _, id := range ids {
		tag, _ := h.db.Exec(c.Request.Context(), `UPDATE candidate_kachra SET status='rejected', updated_at=NOW() WHERE id=$1 AND status='pending'`, id)
		if tag.RowsAffected() > 0 {
			rejected++
		}
	}
	response.OK(c, gin.H{"rejected": rejected, "requested": len(ids)})
}

func (h *Handler) ListJobs(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset, _ := strconv.Atoi(c.Query("offset"))
	if offset < 0 {
		offset = 0
	}
	status := c.Query("status")
	q := `SELECT id::text, s3_key, file_name, status, phase, picked_at, attempts, last_error, error_code, verified, preservation_verified, llm_classifier_label, llm_heading_count, download_count, last_downloaded_at, chunk_count, duration_ms FROM file_jobs`
	args := []interface{}{}
	if status != "" {
		q += ` WHERE status=$1`
		args = append(args, status)
		q += ` ORDER BY created_at DESC LIMIT $2 OFFSET $3`
		args = append(args, limit, offset)
	} else {
		q += ` ORDER BY created_at DESC LIMIT $1 OFFSET $2`
		args = append(args, limit, offset)
	}
	r, err := h.db.Query(c.Request.Context(), q, args...)
	if err != nil {
		response.InternalError(c)
		return
	}
	defer r.Close()
	type row struct {
		ID string `json:"id"`; S3Key string `json:"s3_key"`; FileName string `json:"file_name"`; Status string `json:"status"`; Phase int `json:"phase"`; PickedAt *string `json:"picked_at"`; Attempts int `json:"attempts"`; LastError *string `json:"last_error"`; ErrorCode *string `json:"error_code"`; Verified bool `json:"verified"`; PreservationVerified bool `json:"preservation_verified"`; LLMClassifierLabel *string `json:"llm_classifier_label"`; LLMHeadingCount int `json:"llm_heading_count"`; DownloadCount int `json:"download_count"`; LastDownloadedAt *string `json:"last_downloaded_at"`; ChunkCount int `json:"chunk_count"`; DurationMs *int `json:"duration_ms"`
	}
	var out []row
	for r.Next() {
		var rr row
		var picked, lastErr, errCode, llmLabel, lastDL *string
		var dur *int
		if err := r.Scan(&rr.ID, &rr.S3Key, &rr.FileName, &rr.Status, &rr.Phase, &picked, &rr.Attempts, &lastErr, &errCode, &rr.Verified, &rr.PreservationVerified, &llmLabel, &rr.LLMHeadingCount, &rr.DownloadCount, &lastDL, &rr.ChunkCount, &dur); err != nil {
			continue
		}
		rr.PickedAt = picked
		rr.LastError = lastErr
		rr.ErrorCode = errCode
		rr.LLMClassifierLabel = llmLabel
		rr.LastDownloadedAt = lastDL
		rr.DurationMs = dur
		out = append(out, rr)
	}
	if out == nil {
		out = []row{}
	}
	response.OK(c, out)
}

func (h *Handler) CreateJob(c *gin.Context) {
	var req struct {
		S3Key string `json:"s3_key" binding:"required"`; FileName string `json:"file_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}
	if req.FileName == "" {
		req.FileName = req.S3Key
	}
	var id string
	err := h.db.QueryRow(c.Request.Context(), `INSERT INTO file_jobs(s3_key, file_name) VALUES ($1,$2) RETURNING id::text`, req.S3Key, req.FileName).Scan(&id)
	if err != nil {
		h.logger.Error("create job failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.Created(c, gin.H{"id": id})
}
func (h *Handler) GetJob(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid id")
		return
	}
	var s3key, fname, status string
	var phase, attempts, chunkCount int
	var picked, lastErr, errCode, s3out *string
	var verified bool
	var dur *int
	err = h.db.QueryRow(c.Request.Context(), `SELECT s3_key, file_name, status, phase, picked_at, attempts, last_error, error_code, verified, chunk_count, duration_ms, s3_output_key FROM file_jobs WHERE id=$1`, id).Scan(&s3key, &fname, &status, &phase, &picked, &attempts, &lastErr, &errCode, &verified, &chunkCount, &dur, &s3out)
	if err != nil {
		response.NotFound(c, "job")
		return
	}
	response.OK(c, gin.H{"id": id.String(), "s3_key": s3key, "file_name": fname, "status": status, "phase": phase, "picked_at": picked, "attempts": attempts, "last_error": lastErr, "error_code": errCode, "verified": verified, "chunk_count": chunkCount, "duration_ms": dur, "s3_output_key": s3out})
}

func (h *Handler) PickJob(c *gin.Context) {
	picker := c.Query("picker")
	if picker == "" {
		picker = "vacuum-picker"
	}
	pipeline := NewPipeline(h.service, NewFSStorage(""), h.logger)
	// also heal stuck verifying jobs: reset to scheduled for retry
	_, _ = h.db.Exec(context.Background(), `UPDATE file_jobs SET status='scheduled', phase=0, last_error=NULL, error_code=NULL, retry_after=NULL, updated_at=NOW() WHERE status='verifying' AND updated_at < NOW() - interval '2 minutes'`)
	picked, err := pipeline.PickAndExecute(context.Background(), picker, 10)
	if err != nil {
		h.logger.Error("pick failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	if picked == nil {
		picked = []ExecuteResult{}
	}
	response.OK(c, picked)
}

func (h *Handler) Preview(c *gin.Context) {
	var req struct {
		Text string `json:"text" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}
	res, err := h.service.Engine().Clean(c.Request.Context(), req.Text)
	if err != nil {
		h.logger.Error("preview failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	// self-learning: unseen hits -> candidate_kachra (so admin can promote to trie)
	for _, hit := range res.Hits {
		snippet := hit.Pattern
		if len(res.CleanedText) > 120 {
			snippet = res.CleanedText[:120]
		}
		_, _ = h.db.Exec(c.Request.Context(), `INSERT INTO candidate_kachra(pattern, pattern_type, context_snippet, confidence, hit_count) VALUES ($1,'PHRASE',$2,0.9,1) ON CONFLICT (lower(trim(pattern)), pattern_type, category) DO UPDATE SET hit_count=candidate_kachra.hit_count+1, last_seen_at=NOW()`, hit.Pattern, snippet)
	}
	response.OK(c, gin.H{"cleaned": res.CleanedText, "chunks": res.Chunks, "hits": res.Hits, "p1_hits": res.P1Hits, "sha256_in": res.SHA256In, "sha256_out": res.SHA256Out, "verified": res.Verified, "brain_version": h.service.Brain().Version()})
}

func (h *Handler) BrainVersion(c *gin.Context) {
	trie, ver := h.service.Brain().Snapshot()
	c.JSON(http.StatusOK, gin.H{"version": ver, "patterns": trie.Size()})
}

// VacuumMetrics — Phase 7 observability
func (h *Handler) VacuumMetrics(c *gin.Context) {
	trie, ver := h.service.Brain().Snapshot()
	snap := observability.Global.Snapshot()
	snap["brain_version"] = ver
	snap["brain_patterns"] = trie.Size()
	var pending, approved, rejected int64
	_ = h.db.QueryRow(c.Request.Context(), `SELECT COUNT(*) FROM candidate_kachra WHERE status='pending'`).Scan(&pending)
	_ = h.db.QueryRow(c.Request.Context(), `SELECT COUNT(*) FROM candidate_kachra WHERE status='approved'`).Scan(&approved)
	_ = h.db.QueryRow(c.Request.Context(), `SELECT COUNT(*) FROM candidate_kachra WHERE status='rejected'`).Scan(&rejected)
	snap["candidates_pending"] = pending
	snap["candidates_approved"] = approved
	snap["candidates_rejected"] = rejected
	response.OK(c, snap)
}

func (h *Handler) ExecuteJob(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid id")
		return
	}
	pipeline := NewPipeline(h.service, NewFSStorage(""), h.logger)
	// async: immediate 202 so large file (380 chunks) not canceled by HTTP 30s timeout
	go func(jid uuid.UUID) {
		bg, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		if _, err := pipeline.ExecuteOne(bg, jid); err != nil {
			h.logger.Warn("async execute failed", zap.String("job_id", jid.String()), zap.Error(err))
		}
	}(id)
	response.Created(c, gin.H{"status": "accepted", "job_id": id.String()})
}

func (h *Handler) UploadAndEnqueue(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "INVALID_FILE", "file required")
		return
	}
	defer file.Close()
	safeName := header.Filename
	if safeName == "" {
		safeName = "upload.txt"
	}
	if !(strings.HasSuffix(strings.ToLower(safeName), ".txt") || strings.HasSuffix(strings.ToLower(safeName), ".md")) {
		safeName = safeName + ".txt"
	}
	s3Key := "vacuum/" + uuid.NewString() + "/" + safeName
	storage := NewFSStorage("")
	if err := storage.Put(c.Request.Context(), s3Key, file); err != nil {
		h.logger.Error("upload put failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	var id string
	err = h.db.QueryRow(c.Request.Context(), `INSERT INTO file_jobs(s3_key, file_name, file_size) VALUES ($1,$2,$3) RETURNING id::text`, s3Key, safeName, header.Size).Scan(&id)
	if err != nil {
		h.logger.Error("create job failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.Created(c, gin.H{"id": id, "s3_key": s3Key})
}

func (h *Handler) ListChunks(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid id")
		return
	}
	rows, err := h.db.Query(c.Request.Context(), `SELECT chunk_index, char_start, char_end, hash_input, hash_output, status, llm_label, llm_confidence, headings FROM file_chunks WHERE file_job_id=$1 ORDER BY chunk_index`, id)
	if err != nil {
		response.InternalError(c)
		return
	}
	defer rows.Close()
	type row struct {
		Index int `json:"chunk_index"`; Start int64 `json:"char_start"`; End int64 `json:"char_end"`; HIn string `json:"hash_input"`; HOut string `json:"hash_output"`; Status string `json:"status"`; LLM *string `json:"llm_label"`; Conf *float64 `json:"llm_confidence"`; Headings interface{} `json:"headings"`
	}
	var out []row
	for rows.Next() {
		var r row
		var lbl *string
		var conf *float64
		var headings []byte
		_ = rows.Scan(&r.Index, &r.Start, &r.End, &r.HIn, &r.HOut, &r.Status, &lbl, &conf, &headings)
		r.LLM = lbl
		r.Conf = conf
		// headings kept raw; frontend parses JSON
		_ = headings
		out = append(out, r)
	}
	if out == nil {
		out = []row{}
	}
	response.OK(c, out)
}

func (h *Handler) DownloadCleaned(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid id")
		return
	}
	var s3out *string
	var fname string
	if err := h.db.QueryRow(c.Request.Context(), `SELECT s3_output_key, file_name FROM file_jobs WHERE id=$1`, id).Scan(&s3out, &fname); err != nil {
		response.NotFound(c, "job")
		return
	}
	if s3out == nil || *s3out == "" {
		response.NotFound(c, "cleaned output not ready")
		return
	}
	storage := NewFSStorage("")
	rc, err := storage.Open(c.Request.Context(), *s3out)
	if err != nil {
		response.NotFound(c, "cleaned file")
		return
	}
	defer rc.Close()
	// bump download counter (best-effort)
	_, _ = h.db.Exec(c.Request.Context(), `UPDATE file_jobs SET download_count=download_count+1, last_downloaded_at=NOW() WHERE id=$1`, id)
	c.Header("Content-Disposition", "attachment; filename=\""+fname+".cleaned\"")
	c.Header("Content-Type", "text/plain; charset=utf-8")
	_, _ = c.Writer.Write([]byte{})
	// stream
	buf := make([]byte, 32*1024)
	for {
		n, rerr := rc.Read(buf)
		if n > 0 {
			_, _ = c.Writer.Write(buf[:n])
		}
		if rerr != nil {
			break
		}
	}
}

func (h *Handler) JobsStats(c *gin.Context) {
	rows, err := h.db.Query(c.Request.Context(), `SELECT status, COUNT(*)::bigint FROM file_jobs GROUP BY status`)
	if err != nil {
		response.InternalError(c)
		return
	}
	defer rows.Close()
	m := map[string]int64{}
	for rows.Next() {
		var s string
		var n int64
		_ = rows.Scan(&s, &n)
		m[s] = n
	}
	ct := int64(0)
	_ = h.db.QueryRow(c.Request.Context(), `SELECT COUNT(*) FROM file_chunks WHERE status='verified'`).Scan(&ct)
	avg := int64(0)
	_ = h.db.QueryRow(c.Request.Context(), `SELECT COALESCE(AVG(duration_ms),0)::bigint FROM file_jobs WHERE status='done'`).Scan(&avg)
	response.OK(c, gin.H{"by_status": m, "verified_chunks": ct, "avg_duration_ms": avg})
}

func (h *Handler) RetryJob(c *gin.Context) {
        id, err := uuid.Parse(c.Param("id"))
        if err != nil {
                response.BadRequest(c, "INVALID_ID", "invalid id")
                return
        }
        // Debounce: verifying = 78s background (106 chunks) - block retry within 2m
        var status string
        var updated time.Time
        var attempts int
        err = h.db.QueryRow(c.Request.Context(), `SELECT status, updated_at, attempts FROM file_jobs WHERE id=$1`, id).Scan(&status, &updated, &attempts)
        if err != nil {
                response.NotFound(c, "job")
                return
        }
        if status == "verifying" && time.Since(updated) < 2*time.Minute {
                response.BadRequest(c, "ALREADY_PROCESSING", "Already processing in background (verifying). Please wait 78s. Download will appear on done. Do not press Retry repeatedly.")
                return
        }
        if status != "failed" && status != "quarantined" && !(status == "verifying" && time.Since(updated) >= 2*time.Minute) {
                response.BadRequest(c, "NOT_RETRYABLE", "only failed/quarantined or stuck verifying>2m can be retried")
                return
        }
        if attempts >= 3 {
                response.BadRequest(c, "MAX_RETRIES", "max 3 attempts exceeded")
                return
        }
        tag, _ := h.db.Exec(c.Request.Context(), `UPDATE file_jobs SET status='scheduled', phase=1, attempts=attempts+1, last_error=NULL, error_code=NULL, retry_after=NULL, picked_at=NULL, updated_at=NOW() WHERE id=$1`, id)
        if tag.RowsAffected() == 0 {
                response.BadRequest(c, "NOT_RETRYABLE", "update failed")
                return
        }
        response.OK(c, gin.H{"status": "scheduled"})
}

