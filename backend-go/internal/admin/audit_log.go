package admin

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/response"
)

type AuditLogRow struct {
	ID         uuid.UUID   `json:"id"`
	Action     string      `json:"action"`
	ActorID    *uuid.UUID  `json:"actor_id"`
	ActorEmail *string     `json:"actor_email"`
	Details    interface{} `json:"details"`
	CreatedAt  time.Time   `json:"created_at"`
}

func (h *AdminHandler) GetAuditLog(c *gin.Context) {
	limit := 20
	offset := 0
	if v := c.Query("limit"); v != "" {
		if n, err := auditAtoiBounded(v, 1, 100); err == nil {
			limit = n
		}
	}
	if v := c.Query("offset"); v != "" {
		if n, err := auditAtoiBounded(v, 0, 1000000); err == nil {
			offset = n
		}
	}
	action := c.Query("action")
	rows := []AuditLogRow{}
	if h.db != nil {
		q := `SELECT id, action, actor_id, actor_email, details, created_at FROM audit_logs WHERE true`
		args := []interface{}{}
		idx := 1
		if action != "" {
			q += ` AND action = $` + auditItoa(idx)
			args = append(args, action)
			idx++
		}
		q += ` ORDER BY created_at DESC LIMIT $` + auditItoa(idx) + ` OFFSET $` + auditItoa(idx+1)
		args = append(args, limit, offset)
		rs, err := h.db.Query(c.Request.Context(), q, args...)
		if err == nil {
			defer rs.Close()
			for rs.Next() {
				var r AuditLogRow
				if err := rs.Scan(&r.ID, &r.Action, &r.ActorID, &r.ActorEmail, &r.Details, &r.CreatedAt); err == nil {
					rows = append(rows, r)
				}
			}
		} else {
			h.logger.Info("audit_logs fallback (table may not exist)", zap.Error(err))
		}
	}
	if len(rows) == 0 && h.db != nil {
		var updatedAt *time.Time
		var updatedBy *uuid.UUID
		_ = h.db.QueryRow(c.Request.Context(), `SELECT updated_at, updated_by FROM system_settings WHERE key='retrieval_config'`).Scan(&updatedAt, &updatedBy)
		if updatedAt != nil {
			rows = append(rows, AuditLogRow{ID: uuid.New(), Action: "retrieval_config.updated", ActorID: updatedBy, CreatedAt: *updatedAt, Details: map[string]string{"note": "synthesized from retrieval_config.updated_at (audit_logs empty)"}})
		}
	}
	response.OK(c, gin.H{"items": rows, "limit": limit, "offset": offset, "total": len(rows)})
}

func auditAtoiBounded(s string, min, max int) (int, error) {
	n := 0
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return 0, &auditParseErr{"invalid integer: " + s}
		}
		n = n*10 + int(ch-'0')
		if n > max {
			return max, nil
		}
	}
	if n < min {
		return min, nil
	}
	return n, nil
}

type auditParseErr struct{ msg string }

func (e *auditParseErr) Error() string { return e.msg }

func auditItoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := [20]byte{}
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[pos:])
}
