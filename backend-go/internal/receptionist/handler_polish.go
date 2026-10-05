package receptionist

import (
   "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

// NOTES
func (h *Handler) AddNote(c *gin.Context) {
    sid, _ := uuid.Parse(c.Param("id"))
    tid, _ := c.Get("tenant_id")
    var req struct{ Text string `json:"text" binding:"required"` }
    c.ShouldBindJSON(&req)
    id, err := h.store.AddNote(c.Request.Context(), sid, tid.(uuid.UUID), req.Text)
    if err != nil { c.JSON(500, gin.H{"code":"NOTE_FAILED"}); return }
    ev, _ := h.store.AppendEvent(c.Request.Context(), sid, tid.(uuid.UUID), "note", map[string]any{"id": id, "text": req.Text})
    h.publisher.NotifyNewEvent(c.Request.Context(), sid.String(), ev.ID)
    c.JSON(200, gin.H{"id": id, "text": req.Text})
}
func (h *Handler) ListNotes(c *gin.Context) {
    sid, _ := uuid.Parse(c.Param("id"))
    notes, _ := h.store.ListNotes(c.Request.Context(), sid)
    c.JSON(200, gin.H{"notes": notes})
}
func (h *Handler) DeleteNote(c *gin.Context) {
    nid, _ := uuid.Parse(c.Param("noteId"))
    h.store.DeleteNote(c.Request.Context(), nid)
    c.JSON(200, gin.H{"deleted": true})
}

// PROGRESS
func (h *Handler) GetProgress(c *gin.Context) {
    sid, _ := uuid.Parse(c.Param("id"))
    total, committed, contiguous, nextIdx := h.store.GetProgress(c.Request.Context(), sid)
    c.JSON(200, gin.H{"total": total, "committed": committed, "contiguous": contiguous, "next_idx": nextIdx, "percent": int(float64(contiguous)/float64(max(1,total))*100)})
}
func max(a,b int) int { if a>b {return a}; return b }

// END
func (h *Handler) EndConversation(c *gin.Context) {
    sid, _ := uuid.Parse(c.Param("id"))
    tid, _ := c.Get("tenant_id")
    // Single End button - idempotent
    sess, _ := h.store.GetSession(c.Request.Context(), sid)
    if sess.State == StateCompleted {
        md, hing, _ := h.store.GetFinal(c.Request.Context(), sid)
        c.JSON(200, gin.H{"final_md": md, "hinglish": hing, "bye": "Dhanyawaad, aapka PRD taiyaar hai!"})
        return
    }
    if sess.State != StatePhase2Conversation {
        c.JSON(400, gin.H{"code": "NOT_IN_CONVERSATION"})
        return
    }
    md, hing, err := h.finalSynth.EndConversation(c.Request.Context(), sid, tid.(uuid.UUID))
    if err != nil { c.JSON(500, gin.H{"code":"FINAL_FAILED","message":err.Error()}); return }
    c.JSON(200, gin.H{"final_md": md, "hinglish": hing, "phase": "PHASE_3_ENDED"})
}
