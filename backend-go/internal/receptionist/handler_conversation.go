package receptionist

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

// POST /receptionist/sessions/:id/checkpoints/:cid/answer
// Body: {point_id, answer: "Haan"|"Nahi"|"Samjhao"}
func (h *Handler) AnswerPoint(c *gin.Context) {
    sessionID, _ := uuid.Parse(c.Param("id"))
    checkpointID, _ := uuid.Parse(c.Param("cid"))
    tenantID, _ := c.Get("tenant_id")
    tid := tenantID.(uuid.UUID)

    sess, _ := h.store.GetSession(c.Request.Context(), sessionID)
    if sess.State != StatePhase2Conversation {
        c.JSON(http.StatusBadRequest, gin.H{"code": "CHECKLIST_NOT_READY"})
        return
    }

    var req struct {
        PointID string `json:"point_id" binding:"required"`
        Answer  string `json:"answer" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"code": "INVALID_REQUEST"})
        return
    }
    pts, explain, err := h.convEngine.HandleAnswer(c.Request.Context(), sessionID, checkpointID, tid, req.PointID, req.Answer)
    if err != nil {
        c.JSON(500, gin.H{"code": "ANSWER_FAILED", "message": err.Error()})
        return
    }
    // Smart Judge case - explain return
    if explain != "" {
        c.JSON(200, gin.H{"points": pts, "explain": explain, "judge": true})
        return
    }
    c.JSON(200, gin.H{"points": pts, "all_discussed": h.convEngine.IsAllDiscussed(pts)})
}

// POST /receptionist/sessions/:id/checkpoints/:cid/add-more
// Body: {agenda}
func (h *Handler) AddMore(c *gin.Context) {
    sessionID, _ := uuid.Parse(c.Param("id"))
    checkpointID, _ := uuid.Parse(c.Param("cid"))
    tenantID, _ := c.Get("tenant_id")
    tid := tenantID.(uuid.UUID)
    var req struct{ Agenda string `json:"agenda"` }
    c.ShouldBindJSON(&req)
    sess, _ := h.store.GetSession(c.Request.Context(), sessionID)
    if req.Agenda == "" { req.Agenda = sess.Agenda }
    pts, _, _ := h.store.GetTemplate(c.Request.Context(), checkpointID)
    newPts, err := h.convEngine.AddMorePoints(c.Request.Context(), sessionID, checkpointID, tid, req.Agenda, pts)
    if err != nil { c.JSON(500, gin.H{"code":"ADD_MORE_FAILED"}); return }
    c.JSON(200, gin.H{"points": newPts})
}

// POST /receptionist/sessions/:id/checkpoints/:cid/approve
// Body: {} -> approve gate: "Ye 6 points leke jau?"
func (h *Handler) ApproveCheckpoint(c *gin.Context) {
    sessionID, _ := uuid.Parse(c.Param("id"))
    _ = sessionID
    checkpointID, _ := uuid.Parse(c.Param("cid"))
    // check all discussed
    pts, _, _ := h.store.GetTemplate(c.Request.Context(), checkpointID)
    if !h.convEngine.IsAllDiscussed(pts) {
        c.JSON(400, gin.H{"code": "NOT_ALL_DISCUSSED", "message": "Sab points pe Haan/Nahi karna baki hai"})
        return
    }
    // mark checkpoint COMMITTED
    h.store.db.Exec(c.Request.Context(), `UPDATE receptionist_checkpoints SET status='committed' WHERE id=$1`, checkpointID)
    // build prompt for experts = approved points only
    approved := []Point{}
    for _, p := range pts { if p.Status == "approved" { approved = append(approved, p) } }
    // save approved prompt JSON for Phase 5 expert call
    c.JSON(200, gin.H{"approved_count": len(approved), "points": approved, "msg": "Approved! Ab experts se baat kar sakte ho."})
}
