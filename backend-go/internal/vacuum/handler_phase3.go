package vacuum

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"ai_avengers/backend/internal/observability"
	"ai_avengers/backend/internal/response"
	"ai_avengers/backend/internal/vacuum/engine"
)

// PreviewDiff — Phase 3: DSA vs LLM vs Combined diff for AdminVacuum UI.
// I: {text} -> P: Clean (DSA only) + CleanHybrid (LLM if wired) -> compare -> O: {dsa_cleaned, llm_added_spans, combined_cleaned, diff_stat}
func (h *Handler) PreviewDiff(c *gin.Context) {
	var req struct {
		Text string `json:"text" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}
	text := req.Text
	// DSA only
	dsaRes, err := h.service.Engine().Clean(c.Request.Context(), text)
	if err != nil {
		response.InternalError(c)
		return
	}
	// Combined (hybrid if LLM wired)
	combinedRes, _ := h.service.Engine().CleanHybrid(c.Request.Context(), text, h.service.KachraDetector(), h.service.KachraVerifier(), nil, "")
	if combinedRes == nil {
		combinedRes = dsaRes
	}
	// diff stat: count chars differing - heuristic
	llmOnlyHits := 0
	if combinedRes != nil && dsaRes != nil {
		llmOnlyHits = len(combinedRes.Hits) - len(dsaRes.Hits)
		if llmOnlyHits < 0 {
			llmOnlyHits = 0
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"dsa_cleaned":      dsaRes.CleanedText,
		"dsa_hits":         dsaRes.Hits,
		"dsa_p1_hits":      dsaRes.P1Hits,
		"combined_cleaned": combinedRes.CleanedText,
		"combined_hits":    combinedRes.Hits,
		"combined_p1_hits": combinedRes.P1Hits,
		"llm_added_hits":   llmOnlyHits,
		"sha256_in":        dsaRes.SHA256In,
		"sha256_out_dsa":   dsaRes.SHA256Out,
		"sha256_out_combined": combinedRes.SHA256Out,
		"filler_threshold": engine.FillerThreshold(),
	})
}

// EvalStats — Phase 3 observability: last N vacuum_eval rows + avg eval_score.
func (h *Handler) EvalStats(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := h.db.Query(c.Request.Context(), `SELECT file_job_id::text, dsa_hit_count, llm_suggested, llm_verified, llm_mapped, combined_hit_count, filler_filtered, duration_ms, created_at FROM vacuum_eval ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		response.InternalError(c)
		return
	}
	defer rows.Close()
	type row struct {
		FileJobID        string  `json:"file_job_id"`
		DSA              int     `json:"dsa_hit_count"`
		Suggested        int     `json:"llm_suggested"`
		Verified         int     `json:"llm_verified"`
		Mapped           int     `json:"llm_mapped"`
		Combined         int     `json:"combined_hit_count"`
		FillerFiltered   int     `json:"filler_filtered"`
		DurationMs       int     `json:"duration_ms"`
		CreatedAt        string  `json:"created_at"`
	}
	var out []row
	for rows.Next() {
		var r row
		var ts string
		_ = rows.Scan(&r.FileJobID, &r.DSA, &r.Suggested, &r.Verified, &r.Mapped, &r.Combined, &r.FillerFiltered, &r.DurationMs, &ts)
		r.CreatedAt = ts
		out = append(out, r)
	}
	if out == nil {
		out = []row{}
	}
	var avgEval *float64
	_ = h.db.QueryRow(c.Request.Context(), `SELECT AVG(eval_score) FROM file_jobs WHERE status='done' AND eval_score IS NOT NULL`).Scan(&avgEval)
	snap := observability.Global.Snapshot()
	response.OK(c, gin.H{"eval": out, "avg_eval_score": avgEval, "metrics": snap})
}

// DriftCheck — Phase 3 drift detection endpoint.
func (h *Handler) DriftCheck(c *gin.Context) {
	drift, recent, older, err := h.service.DriftCheck(c.Request.Context())
	if err != nil {
		response.InternalError(c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"drift": drift, "recent_avg": recent, "older_avg": older, "threshold": 0.15})
}

// AutoPromote — Phase 3 human-in-loop reducer: promote confidence>=0.90 & hit_count>=3.
func (h *Handler) AutoPromote(c *gin.Context) {
	n, err := h.service.AutoPromote(c.Request.Context())
	if err != nil {
		response.InternalError(c)
		return
	}
	// optional feedback: echo promoted
	c.JSON(http.StatusOK, gin.H{"promoted": n, "brain_version": h.service.Brain().Version()})
}

// helper to satisfy import usage
var _ = strings.TrimSpace
