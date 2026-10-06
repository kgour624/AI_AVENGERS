import pathlib
audit = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\internal\admin\audit_log.go")
fixed_audit = """package admin

import (
\t\"time\"

\t\"github.com/gin-gonic/gin\"
\t\"github.com/google/uuid\"
\t\"go.uber.org/zap\"

\t\"ai_avengers/backend/internal/response\"
)

type AuditLogRow struct {
\tID         uuid.UUID   `json:\"id\"`
\tAction     string      `json:\"action\"`
\tActorID    *uuid.UUID  `json:\"actor_id\"`
\tActorEmail *string     `json:\"actor_email\"`
\tDetails    interface{} `json:\"details\"`
\tCreatedAt  time.Time   `json:\"created_at\"`
}

func (h *AdminHandler) GetAuditLog(c *gin.Context) {
\tlimit := 20
\toffset := 0
\tif v := c.Query(\"limit\"); v != \"\" {
\t\tif n, err := auditAtoiBounded(v, 1, 100); err == nil {
\t\t\tlimit = n
\t\t}
\t}
\tif v := c.Query(\"offset\"); v != \"\" {
\t\tif n, err := auditAtoiBounded(v, 0, 1000000); err == nil {
\t\t\toffset = n
\t\t}
\t}
\taction := c.Query(\"action\")
\trows := []AuditLogRow{}
\tif h.db != nil {
\t\tq := `SELECT id, action, actor_id, actor_email, details, created_at FROM audit_logs WHERE true`
\t\targs := []interface{}{}
\t\tidx := 1
\t\tif action != \"\" {
\t\t\tq += ` AND action = $` + auditItoa(idx)
\t\t\targs = append(args, action)
\t\t\tidx++
\t\t}
\t\tq += ` ORDER BY created_at DESC LIMIT $` + auditItoa(idx) + ` OFFSET $` + auditItoa(idx+1)
\t\targs = append(args, limit, offset)
\t\trs, err := h.db.Query(c.Request.Context(), q, args...)
\t\tif err == nil {
\t\t\tdefer rs.Close()
\t\t\tfor rs.Next() {
\t\t\t\tvar r AuditLogRow
\t\t\t\tif err := rs.Scan(&r.ID, &r.Action, &r.ActorID, &r.ActorEmail, &r.Details, &r.CreatedAt); err == nil {
\t\t\t\t\trows = append(rows, r)
\t\t\t\t}
\t\t\t}
\t\t} else {
\t\t\th.logger.Warn(\"audit_logs query failed (table may not exist)\", zap.Error(err))
\t\t}
\t}
\tif len(rows) == 0 && h.db != nil {
\t\tvar updatedAt *time.Time
\t\tvar updatedBy *uuid.UUID
\t\t_ = h.db.QueryRow(c.Request.Context(), `SELECT updated_at, updated_by FROM retrieval_config WHERE id = '00000000-0000-0000-0000-000000000001'`).Scan(&updatedAt, &updatedBy)
\t\tif updatedAt != nil {
\t\t\trows = append(rows, AuditLogRow{ID: uuid.New(), Action: \"retrieval_config.updated\", ActorID: updatedBy, CreatedAt: *updatedAt, Details: map[string]string{\"note\": \"synthesized from retrieval_config.updated_at (audit_logs empty)\"}})
\t\t}
\t}
\tresponse.OK(c, gin.H{\"items\": rows, \"limit\": limit, \"offset\": offset, \"total\": len(rows)})
}

func auditAtoiBounded(s string, min, max int) (int, error) {
\tn := 0
\tfor _, ch := range s {
\t\tif ch < '0' || ch > '9' {
\t\t\treturn 0, &auditParseErr{\"invalid integer: \" + s}
\t\t}
\t\tn = n*10 + int(ch-'0')
\t\tif n > max {
\t\t\treturn max, nil
\t\t}
\t}
\tif n < min {
\t\treturn min, nil
\t}
\treturn n, nil
}

type auditParseErr struct{ msg string }

func (e *auditParseErr) Error() string { return e.msg }

func auditItoa(n int) string {
\tif n == 0 {
\t\treturn \"0\"
\t}
\tbuf := [20]byte{}
\tpos := len(buf)
\tfor n > 0 {
\t\tpos--
\t\tbuf[pos] = byte('0' + n%10)
\t\tn /= 10
\t}
\treturn string(buf[pos:])
}
"""
audit.write_text(fixed_audit, encoding="utf-8")
print("fixed audit_log.go", len(fixed_audit))
