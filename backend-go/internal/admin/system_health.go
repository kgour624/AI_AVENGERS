package admin

import (
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/response"
	"ai_avengers/backend/internal/training"
)

type SystemHealthResponse struct {
	RetrievalConfig *training.RetrievalConfig `json:"retrieval_config"`
	FlagEnabled     bool                      `json:"flag_enabled"`
	FlagSource      string                    `json:"flag_source"`
	Runtime         RuntimeStats              `json:"runtime"`
	DBPool          DBPoolStats               `json:"db_pool"`
	Counts          HealthCounts              `json:"counts"`
	CheckedAt       time.Time                 `json:"checked_at"`
}

type RuntimeStats struct {
	GoRoutines int     `json:"go_routines"`
	AllocMB    float64 `json:"alloc_mb"`
	NumCPU     int     `json:"num_cpu"`
	UptimeSec  int64   `json:"uptime_sec"`
}

type DBPoolStats struct {
	AcquiredConns int32 `json:"acquired_conns"`
	IdleConns     int32 `json:"idle_conns"`
	TotalConns    int32 `json:"total_conns"`
}

type HealthCounts struct {
	TotalParents  int `json:"total_parents"`
	TotalChildren int `json:"total_children"`
	Orphans       int `json:"orphans"`
	Experts       int `json:"experts"`
}

var serverStartTime = time.Now()

func (h *AdminHandler) GetSystemHealth(c *gin.Context) {
	cfg := training.DefaultRetrievalConfig()
	source := "default"
	// FIX1: load actual DB config instead of hardcoded defaults
	if h.db != nil {
		dbCfg := training.LoadRetrievalConfig(c.Request.Context(), h.db)
		var exists bool
		_ = h.db.QueryRow(c.Request.Context(), `SELECT EXISTS(SELECT 1 FROM system_settings WHERE key='retrieval_config')`).Scan(&exists)
		if exists {
			cfg = dbCfg
			source = "system_settings"
		}
	}
	var counts HealthCounts
	var dbStats DBPoolStats
	if h.db != nil {
		_ = h.db.QueryRow(c.Request.Context(), `SELECT count(*) FROM expert_pages`).Scan(&counts.TotalParents)
		_ = h.db.QueryRow(c.Request.Context(), `SELECT count(*) FROM course_chunks WHERE is_child = true`).Scan(&counts.TotalChildren)
		_ = h.db.QueryRow(c.Request.Context(), `SELECT count(*) FROM course_chunks WHERE parent_id IS NULL AND is_child = true`).Scan(&counts.Orphans)
		_ = h.db.QueryRow(c.Request.Context(), `SELECT count(*) FROM experts`).Scan(&counts.Experts)
		s := h.db.Stat()
		dbStats = DBPoolStats{AcquiredConns: s.AcquiredConns(), IdleConns: s.IdleConns(), TotalConns: s.TotalConns()}
		h.logger.Debug("system health", zap.Int("parents", counts.TotalParents), zap.Int("orphans", counts.Orphans))
	}
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	resp := SystemHealthResponse{
		RetrievalConfig: &cfg,
		FlagEnabled:     cfg.EnableParentChild,
		FlagSource:      source,
		Runtime: RuntimeStats{GoRoutines: runtime.NumGoroutine(), AllocMB: float64(m.Alloc)/1024/1024, NumCPU: runtime.NumCPU(), UptimeSec: int64(time.Since(serverStartTime).Seconds())},
		DBPool:    dbStats,
		Counts:    counts,
		CheckedAt: time.Now().UTC(),
	}
	response.OK(c, resp)
}
