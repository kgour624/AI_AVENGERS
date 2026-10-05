package receptionist

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

// GET /receptionist/sessions/:id/experts/suggest?agenda=LLD -> dynamic suggest, client final tick
func (h *Handler) SuggestExperts(c *gin.Context) {
    sessionID, _ := uuid.Parse(c.Param("id"))
    tenantID, _ := c.Get("tenant_id")
    tid := tenantID.(uuid.UUID)
    agenda := c.Query("agenda")
    if agenda == "" {
        sess, _ := h.store.GetSession(c.Request.Context(), sessionID)
        agenda = sess.Agenda
    }
    experts, _ := h.store.SuggestExperts(c.Request.Context(), tid, agenda, 6)
    c.JSON(200, gin.H{"experts": experts, "hint": "Client jo tick kare wahi final, N=1..all"})
}

// POST /receptionist/sessions/:id/checkpoints/:cid/ask-experts
// Body: {expert_ids: ["uuid1","uuid2"]} - client-selected final, NOT fixed 3 (tune bola: secretary suggests, client final)
func (h *Handler) AskExpertsV2(c *gin.Context) {
    sessionID, _ := uuid.Parse(c.Param("id"))
    checkpointID, _ := uuid.Parse(c.Param("cid"))
    tenantID, _ := c.Get("tenant_id")
    tid := tenantID.(uuid.UUID)

    // Phase 1 lock reuse - fail-closed
    sess, _ := h.store.GetSession(c.Request.Context(), sessionID)
    if sess.State != StatePhase2Conversation {
        c.JSON(http.StatusBadRequest, gin.H{"code": "CHECKLIST_NOT_READY", "message": "Phase 2 me hi experts allowed"})
        return
    }
    // Approve gate check - bina approve ke expert call nahi
    pts, _, _ := h.store.GetTemplate(c.Request.Context(), checkpointID)
    if !h.convEngine.IsAllDiscussed(pts) {
        c.JSON(400, gin.H{"code": "NOT_APPROVED", "message": "Pehle Haan/Nahi + Approve gate karo, fir expert call"})
        return
    }
    approved := []Point{}
    for _, p := range pts { if p.Status == "approved" { approved = append(approved, p) } }

    var req struct { ExpertIDs []string `json:"expert_ids" binding:"required"` }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"code": "INVALID_REQUEST"})
        return
    }
    eids := make([]uuid.UUID, 0, len(req.ExpertIDs))
    for _, s := range req.ExpertIDs {
        id, _ := uuid.Parse(s)
        eids = append(eids, id)
    }

    // Typing event before fanout
    h.store.AppendEvent(c.Request.Context(), sessionID, tid, "typing", map[string]any{"msg": "Experts se baat kar rahi hu..."})
    results, err := h.expertFanOut.CallExperts(c.Request.Context(), sessionID, checkpointID, tid, eids, approved, sess.Agenda)
    if err != nil {
        c.JSON(500, gin.H{"code": "EXPERT_CALL_FAILED", "message": err.Error()})
        return
    }
    c.JSON(200, gin.H{"results": results, "count": len(results)})
}

// POST /receptionist/expert-responses/:id/rate  Body: {rating: 1..5}
func (h *Handler) RateExpert(c *gin.Context) {
    rid, _ := uuid.Parse(c.Param("id"))
    var req struct { Rating int `json:"rating" binding:"required"` }
    c.ShouldBindJSON(&req)
    if req.Rating <1 || req.Rating >5 { c.JSON(400, gin.H{"code":"INVALID_RATING"}); return }
    if err := h.store.RateExpertResponse(c.Request.Context(), rid, req.Rating); err != nil { c.JSON(500, gin.H{"code":"RATE_FAILED"}); return }
    // >3 Approve & Next, ≤3 Reopen same point - pressure update
    status := "REOPENED"
    if req.Rating > 3 { status = "APPROVED" }
    c.JSON(200, gin.H{"status": status, "msg": status})
}
