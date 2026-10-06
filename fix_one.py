import pathlib
sh = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\internal\admin\system_health.go")
fixed_sh = """package admin

import (
\t\"runtime\"
\t\"time\"

\t\"github.com/gin-gonic/gin\"
\t\"go.uber.org/zap\"

\t\"ai_avengers/backend/internal/response\"
\t\"ai_avengers/backend/internal/training\"
)

type SystemHealthResponse struct {
\tRetrievalConfig *training.RetrievalConfig `json:\"retrieval_config\"`
\tFlagEnabled     bool                      `json:\"flag_enabled\"`
\tFlagSource      string                    `json:\"flag_source\"`
\tRuntime         RuntimeStats              `json:\"runtime\"`
\tDBPool          DBPoolStats               `json:\"db_pool\"`
\tCounts          HealthCounts              `json:\"counts\"`
\tCheckedAt       time.Time                 `json:\"checked_at\"`
}

type RuntimeStats struct {
\tGoRoutines int     `json:\"go_routines\"`
\tAllocMB    float64 `json:\"alloc_mb\"`
\tNumCPU     int     `json:\"num_cpu\"`
\tUptimeSec  int64   `json:\"uptime_sec\"`
}

type DBPoolStats struct {
\tAcquiredConns int32 `json:\"acquired_conns\"`
\tIdleConns     int32 `json:\"idle_conns\"`
\tTotalConns    int32 `json:\"total_conns\"`
}

type HealthCounts struct {
\tTotalParents  int `json:\"total_parents\"`
\tTotalChildren int `json:\"total_children\"`
\tOrphans       int `json:\"orphans\"`
\tExperts       int `json:\"experts\"`
}

var serverStartTime = time.Now()

func (h *AdminHandler) GetSystemHealth(c *gin.Context) {
\tcfg := training.DefaultRetrievalConfig()
\tsource := \"default\"
\tvar counts HealthCounts
\tvar dbStats DBPoolStats
\tif h.db != nil {
\t\t_ = h.db.QueryRow(c.Request.Context(), `SELECT count(*) FROM expert_pages`).Scan(&counts.TotalParents)
\t\t_ = h.db.QueryRow(c.Request.Context(), `SELECT count(*) FROM course_chunks WHERE is_child = true`).Scan(&counts.TotalChildren)
\t\t_ = h.db.QueryRow(c.Request.Context(), `SELECT count(*) FROM course_chunks WHERE parent_id IS NULL AND is_child = true`).Scan(&counts.Orphans)
\t\t_ = h.db.QueryRow(c.Request.Context(), `SELECT count(*) FROM experts`).Scan(&counts.Experts)
\t\ts := h.db.Stat()
\t\tdbStats = DBPoolStats{AcquiredConns: s.AcquiredConns(), IdleConns: s.IdleConns(), TotalConns: s.TotalConns()}
\t\th.logger.Debug(\"system health\", zap.Int(\"parents\", counts.TotalParents), zap.Int(\"orphans\", counts.Orphans))
\t}
\tvar m runtime.MemStats
\truntime.ReadMemStats(&m)
\tresp := SystemHealthResponse{
\t\tRetrievalConfig: &cfg,
\t\tFlagEnabled:     cfg.EnableParentChild,
\t\tFlagSource:      source,
\t\tRuntime: RuntimeStats{GoRoutines: runtime.NumGoroutine(), AllocMB: float64(m.Alloc)/1024/1024, NumCPU: runtime.NumCPU(), UptimeSec: int64(time.Since(serverStartTime).Seconds())},
\t\tDBPool:    dbStats,
\t\tCounts:    counts,
\t\tCheckedAt: time.Now().UTC(),
\t}
\tresponse.OK(c, resp)
}
"""
sh.write_text(fixed_sh, encoding="utf-8")
print("fixed system_health.go len", len(fixed_sh))
