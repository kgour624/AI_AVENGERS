package receptionist

import (
    "net/http"
    "strconv"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

// POST /receptionist/sessions/:id/checkpoints/:idx/template/ensure
// Body: {agenda: "PRD for JIRA"} - ya session.agenda se lega
func (h *Handler) EnsureTemplate(c *gin.Context) {
    sessionID, _ := uuid.Parse(c.Param("id"))
    idxStr := c.Param("idx")
    idx, _ := strconv.Atoi(idxStr)
    tenantID, _ := c.Get("tenant_id")
    tid := tenantID.(uuid.UUID)

    // Phase check - Phase 2 me hi template search allowed (Phase 1 lock reuse)
    sess, _ := h.store.GetSession(c.Request.Context(), sessionID)
    if sess.State != StatePhase2Conversation {
        c.JSON(http.StatusBadRequest, gin.H{"code": "CHECKLIST_NOT_READY", "message": "Phase 2 me hi template search hota hai"})
        return
    }

    var req struct { Agenda string `json:"agenda"` }
    c.ShouldBindJSON(&req)
    agenda := req.Agenda
    if agenda == "" { agenda = sess.Agenda } // fallback session agenda

    // checkpoint fetch
    var checkpointID uuid.UUID
    h.store.db.QueryRow(c.Request.Context(), `SELECT id FROM receptionist_checkpoints WHERE session_id=$1 AND order_idx=$2`, sessionID, idx).Scan(&checkpointID)
    if checkpointID == uuid.Nil {
        c.JSON(http.StatusNotFound, gin.H{"code": "CHECKPOINT_NOT_FOUND"})
        return
    }

    pts, err := h.templateOrch.EnsureTemplate(c.Request.Context(), sessionID, checkpointID, tid, idx, agenda)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"code": "SEARCH_FAILED", "message": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"points": pts, "search_query": agenda, "cached": false})
}

// POST /receptionist/sessions/:id/checkpoints/next  -> clear old + ensure new (frontend Next button)
func (h *Handler) NextCheckpoint(c *gin.Context) {
    sessionID, _ := uuid.Parse(c.Param("id"))
    tenantID, _ := c.Get("tenant_id")
    tid := tenantID.(uuid.UUID)
    var req struct { OldID string `json:"old_checkpoint_id"`; NewIdx int `json:"new_idx"`; Agenda string `json:"agenda"` }
    c.ShouldBindJSON(&req)
    oldID, _ := uuid.Parse(req.OldID)
    var newID uuid.UUID
    // new checkpoint already created via POST /checkpoints, fetch its ID
    h.store.db.QueryRow(c.Request.Context(), `SELECT id FROM receptionist_checkpoints WHERE session_id=$1 AND order_idx=$2`, sessionID, req.NewIdx).Scan(&newID)
    pts, err := h.templateOrch.OnNextCheckpoint(c.Request.Context(), oldID, newID, sessionID, tid, req.NewIdx, req.Agenda)
    if err != nil { c.JSON(500, gin.H{"code":"SEARCH_FAILED"}); return }
    c.JSON(200, gin.H{"points": pts})
}
