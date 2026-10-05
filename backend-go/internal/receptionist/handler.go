// Receptionist HTTP layer - gin based.
// FIX (bug #5/#9): pehle ye file github.com/go-chi/chi/v5 use karti thi, jo is
// module ki dependency hi nahi hai (go.mod me sirf gin hai) -
// isliye package compile hi nahi hota tha. Ab poora handler gin pe hai.
package receptionist

import (
"context"
"encoding/json"
"fmt"
"net/http"
"time"

"github.com/gin-gonic/gin"
"github.com/google/uuid"
)

// Handler - HTTP layer. Business logic Orchestrator me hai (App layer ko import nahi karta).
type Handler struct {
    orch      *Orchestrator
    store     *Store
    publisher *Publisher

    // Phase 2/3/5 engines - main.go wiring se inject hote hain
    convEngine   *ConversationEngine
    templateOrch *TemplateOrchestrator
    expertFanOut *ExpertFanOut
    finalSynth   *FinalSynthesizer
    svc          *Service
}

func NewHandler(orch *Orchestrator, store *Store, pub *Publisher) *Handler {
    return &Handler{orch: orch, store: store, publisher: pub}
}

// WithEngines - saare engines ek hi call me inject (main.go wiring)
func (h *Handler) WithEngines(conv *ConversationEngine, tpl *TemplateOrchestrator, fanout *ExpertFanOut, synth *FinalSynthesizer) *Handler {
    h.convEngine = conv
    h.templateOrch = tpl
    h.expertFanOut = fanout
    h.finalSynth = synth
    return h
}

func (h *Handler) emitAndNotify(ctx context.Context, sessionID, tenantID uuid.UUID, typ string, payload any) {
	if h.store != nil {
		ev, err := h.store.AppendEvent(ctx, sessionID, tenantID, typ, payload)
		if err == nil && ev != nil && h.publisher != nil {
			_ = h.publisher.NotifyNewEvent(ctx, sessionID.String(), ev.ID)
		}
	}
}

// RegisterRoutes - main.go ise /api/receptionist group pe mount karta hai.
// NOTE: gin (httprouter) me ek hi level par static + wildcard sibling allowed
// nahi hai, isliye koi bhi "checkpoints/next" type static route use nahi kiya.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
// Phase 1 - setup
rg.POST("/sessions", h.CreateSession)
rg.GET("/sessions/:id", h.GetSession)
rg.DELETE("/sessions/:id", h.Delete)
rg.POST("/sessions/:id/setup", h.CompleteSetup)

// Phase 2 - one-by-one conversation
rg.POST("/sessions/:id/chat", h.Chat)
rg.POST("/sessions/:id/change", h.Change)
rg.POST("/sessions/:id/next-checkpoint", h.NextCheckpoint)
rg.POST("/sessions/:id/add-checkpoints", h.AddCheckpoints)
rg.POST("/sessions/:id/template/:idx/ensure", h.EnsureTemplate)
rg.POST("/sessions/:id/checkpoints/:cid/answer", h.AnswerPoint)
rg.POST("/sessions/:id/checkpoints/:cid/add-more", h.AddMore)
rg.POST("/sessions/:id/checkpoints/:cid/approve", h.ApproveCheckpoint)
rg.POST("/sessions/:id/checkpoints/:cid/rate", h.RateCheckpoint)
rg.POST("/sessions/:id/checkpoints/:cid/reopen", h.Reopen)

// Experts (Phase 3)
rg.GET("/sessions/:id/experts/suggest", h.SuggestExperts)
rg.POST("/sessions/:id/ask-experts", h.AskExperts)
rg.POST("/sessions/:id/checkpoints/:cid/ask-experts", h.AskExpertsV2)
rg.POST("/expert-responses/:id/rate", h.RateExpert)

// Notes ledger
rg.GET("/sessions/:id/notes", h.ListNotes)
rg.POST("/sessions/:id/notes", h.AddNote)
rg.DELETE("/sessions/:id/notes/:noteId", h.DeleteNote)

// Attachments
rg.POST("/sessions/:id/upload", h.Upload)

// Phase 4/5 - summary loop, approve, final
rg.POST("/sessions/:id/done", h.Done)
rg.GET("/sessions/:id/summary", h.GetSummary)
rg.POST("/sessions/:id/summary/rate", h.RateSummary)
rg.POST("/sessions/:id/approve", h.Approve)
rg.GET("/sessions/:id/download", h.Download)
rg.GET("/sessions/:id/progress", h.GetProgress)
rg.POST("/sessions/:id/end", h.EndConversation)

// SSE live stream (durable replay + redis pubsub)
rg.GET("/sessions/:id/events", h.StreamEvents)
}

func writeJSON(c *gin.Context, code int, v any) { c.JSON(code, v) }

// getTenantAdmin - App layer JWT middleware se tenant/admin nikalta hai;
// dev/testing ke liye header fallback.
func getTenantAdmin(c *gin.Context) (uuid.UUID, uuid.UUID) {
	var tenantID, adminID uuid.UUID
	if v, ok := c.Get("tenant_id"); ok {
		if t, ok := v.(uuid.UUID); ok {
			tenantID = t
		}
	}
	if v, ok := c.Get("user_id"); ok {
		if u, ok := v.(uuid.UUID); ok {
			if tenantID == uuid.Nil { tenantID = u }
			if adminID == uuid.Nil { adminID = u }
		}
	}
	if v, ok := c.Get("admin_id"); ok {
		if a, ok := v.(uuid.UUID); ok {
			adminID = a
		}
	}
	if tenantID == uuid.Nil {
		tenantID, _ = uuid.Parse(c.GetHeader("X-Tenant-ID"))
	}
	if adminID == uuid.Nil {
		adminID, _ = uuid.Parse(c.GetHeader("X-Admin-ID"))
	}
	if tenantID == uuid.Nil {
		tenantID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	}
	if adminID == uuid.Nil {
		adminID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	}
	return tenantID, adminID
}

// POST /sessions
func (h *Handler) CreateSession(c *gin.Context) {
tenantID, adminID := getTenantAdmin(c)
if tenantID == uuid.Nil {
writeJSON(c, 401, gin.H{"code": "TENANT_REQUIRED"})
return
}
var req struct {
Language string `json:"language"`
Persona  string `json:"persona"`
Agenda   string `json:"agenda"`
}
if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
writeJSON(c, 400, gin.H{"code": "INVALID_BODY"})
return
}
if req.Language == "" {
req.Language = string(LangEN)
}
sess, err := h.orch.CreateSession(c.Request.Context(), tenantID, adminID, Language(req.Language), req.Persona, req.Agenda)
if err != nil {
writeJSON(c, 400, gin.H{"error": err.Error()})
return
}
sessionsTotal.WithLabelValues(string(sess.State)).Inc()
writeJSON(c, 200, gin.H{"session": sess})
}

// GET /sessions/:id
func (h *Handler) GetSession(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		writeJSON(c, 400, gin.H{"code": "INVALID_SESSION_ID"})
		return
	}
	sess, err := h.store.GetSession(c.Request.Context(), id)
	if err != nil {
		writeJSON(c, 404, gin.H{"error": "session not found"})
		return
	}
	cps, _ := h.store.GetCheckpoints(c.Request.Context(), id)
	writeJSON(c, 200, gin.H{"session": sess, "checkpoints": cps})
}

// POST /sessions/:id/setup
func (h *Handler) CompleteSetup(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil {
writeJSON(c, 400, gin.H{"code": "INVALID_SESSION_ID"})
return
}
var req struct {
Checklist   []string `json:"checklist"`
NoChecklist bool     `json:"no_checklist"`
}
_ = json.NewDecoder(c.Request.Body).Decode(&req)
if err := h.orch.CompleteSetup(c.Request.Context(), id, req.Checklist, req.NoChecklist); err != nil {
writeJSON(c, 400, gin.H{"error": err.Error()})
return
}
h.emitAndNotify(c.Request.Context(), id, uuid.Nil, "phase_change", map[string]any{"state": string(StatePhase2Conversation)})
writeJSON(c, 200, gin.H{"status": "setup_done"})
}

// POST /sessions/:id/chat
func (h *Handler) Chat(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		CheckpointID string `json:"checkpoint_id"`
		Message      string `json:"message"`
	}
	_ = json.NewDecoder(c.Request.Body).Decode(&req)

	if h.svc != nil {
		reply, err := h.svc.HandleChat(c.Request.Context(), id, req.Message)
		if err == nil {
			writeJSON(c, 200, gin.H{"reply": reply})
			return
		}
	}

	sessionID, err := uuid.Parse(id)
	if err != nil {
		writeJSON(c, 400, gin.H{"code": "INVALID_SESSION_ID"})
		return
	}
	cid, err := uuid.Parse(req.CheckpointID)
	if err != nil {
		reply, err := h.orch.HandleUserMessage(c.Request.Context(), sessionID, req.Message)
		if err != nil {
			writeJSON(c, 500, gin.H{"error": err.Error()})
			return
		}
		writeJSON(c, 200, gin.H{"reply": reply})
		return
	}
	reply, err := h.orch.HandleChatMessage(c.Request.Context(), sessionID, cid, req.Message)
	if err != nil {
		writeJSON(c, 500, gin.H{"error": err.Error()})
		return
	}
	writeJSON(c, 200, gin.H{"reply": reply})
}

// POST /sessions/:id/change
func (h *Handler) Change(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		writeJSON(c, 400, gin.H{"code": "INVALID_SESSION_ID"})
		return
	}
	var req struct {
		CheckpointID string `json:"checkpoint_id"`
		Change       string `json:"change"`
	}
	_ = json.NewDecoder(c.Request.Body).Decode(&req)
	cid, _ := uuid.Parse(req.CheckpointID)
	if err := h.orch.HandleChange(c.Request.Context(), id, cid, req.Change); err != nil {
		writeJSON(c, 400, gin.H{"error": err.Error()})
		return
	}
	writeJSON(c, 200, gin.H{"status": "change_recorded"})
}

// POST /sessions/:id/ask-experts
func (h *Handler) AskExperts(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		writeJSON(c, 400, gin.H{"code": "INVALID_SESSION_ID"})
		return
	}
	session, _ := h.store.GetSession(c.Request.Context(), id)
	var req struct {
		CheckpointID string      `json:"checkpoint_id"`
		ExpertIDs    []uuid.UUID `json:"expert_ids"`
		Question     string      `json:"question"`
	}
	_ = json.NewDecoder(c.Request.Body).Decode(&req)
	var cidPtr *uuid.UUID
	if u, perr := uuid.Parse(req.CheckpointID); perr == nil {
		cidPtr = &u
	}
	lang := Language("EN")
	if session != nil {
		lang = session.Language
	}
	results, err := h.orch.coordinator.AskExperts(c.Request.Context(), AskExpertsRequest{
		SessionID: id, CheckpointID: cidPtr, ExpertIDs: req.ExpertIDs, Question: req.Question, ClientLang: lang,
	})
	if err != nil {
		writeJSON(c, 500, gin.H{"error": err.Error()})
		return
	}
	writeJSON(c, 200, gin.H{"results": results})
}

// POST /sessions/:id/checkpoints/:cid/rate
func (h *Handler) RateCheckpoint(c *gin.Context) {
	sid, err := uuid.Parse(c.Param("id"))
	if err != nil {
		writeJSON(c, 400, gin.H{"code": "INVALID_SESSION_ID"})
		return
	}
	session, _ := h.store.GetSession(c.Request.Context(), sid)
	if session == nil || session.State != StatePhase2Conversation {
		writeJSON(c, 400, gin.H{"code": "CHECKLIST_NOT_READY"})
		return
	}
	cid, err := uuid.Parse(c.Param("cid"))
	if err != nil {
		writeJSON(c, 400, gin.H{"code": "INVALID_CHECKPOINT_ID"})
		return
	}
	var req struct {
		Stars   int    `json:"stars"`
		Summary string `json:"summary"`
	}
	_ = json.NewDecoder(c.Request.Body).Decode(&req)
	committed, err := h.orch.RateCheckpoint(c.Request.Context(), sid, cid, req.Stars, req.Summary)
	if err != nil {
		writeJSON(c, 400, gin.H{"error": err.Error()})
		return
	}
	writeJSON(c, 200, gin.H{"committed": committed})
}

// POST /sessions/:id/done
func (h *Handler) Done(c *gin.Context) {
	id := c.Param("id")
	if h.svc != nil {
		if err := h.svc.Done(c.Request.Context(), id); err == nil {
			writeJSON(c, 200, gin.H{"status": "moved_to_summary"})
			return
		}
	}
	sessionID, err := uuid.Parse(id)
	if err != nil {
		writeJSON(c, 400, gin.H{"code": "INVALID_SESSION_ID"})
		return
	}
	if err := h.orch.HandleDone(c.Request.Context(), sessionID); err != nil {
		writeJSON(c, 500, gin.H{"error": err.Error()})
		return
	}
	writeJSON(c, 200, gin.H{"status": "moved_to_summary"})
}

// POST /sessions/:id/add-checkpoints
func (h *Handler) AddCheckpoints(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil {
writeJSON(c, 400, gin.H{"code": "INVALID_SESSION_ID"})
return
}
var req struct {
Texts []string `json:"texts"`
}
_ = json.NewDecoder(c.Request.Body).Decode(&req)
if err := h.orch.AddMoreCheckpoints(c.Request.Context(), id, req.Texts); err != nil {
writeJSON(c, 400, gin.H{"error": err.Error()})
return
}
writeJSON(c, 200, gin.H{"status": "added"})
}

// GET /sessions/:id/summary
func (h *Handler) GetSummary(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil {
writeJSON(c, 400, gin.H{"code": "INVALID_SESSION_ID"})
return
}
summary, err := h.orch.GenerateSummary(c.Request.Context(), id)
if err != nil {
writeJSON(c, 500, gin.H{"error": err.Error()})
return
}
writeJSON(c, 200, gin.H{"summary": summary})
}

// POST /sessions/:id/summary/rate
func (h *Handler) RateSummary(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil {
writeJSON(c, 400, gin.H{"code": "INVALID_SESSION_ID"})
return
}
var req struct {
Stars    int    `json:"stars"`
Feedback string `json:"feedback"`
}
_ = json.NewDecoder(c.Request.Body).Decode(&req)
approved, err := h.orch.RateSummary(c.Request.Context(), id, req.Stars, req.Feedback)
if err != nil {
writeJSON(c, 400, gin.H{"error": err.Error()})
return
}
writeJSON(c, 200, gin.H{"approved": approved})
}

// POST /sessions/:id/approve
func (h *Handler) Approve(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil {
writeJSON(c, 400, gin.H{"code": "INVALID_SESSION_ID"})
return
}
doc, err := h.orch.ApproveAndGenerateFinal(c.Request.Context(), id)
if err != nil {
writeJSON(c, 400, gin.H{"error": err.Error()})
return
}
writeJSON(c, 200, gin.H{"final_response": doc})
}

// GET /sessions/:id/download
func (h *Handler) Download(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil {
writeJSON(c, 400, gin.H{"code": "INVALID_SESSION_ID"})
return
}
doc, err := h.orch.ApproveAndGenerateFinal(c.Request.Context(), id)
if err != nil {
// already completed -> last note hi final doc maano
notes, nerr := h.store.ListNotes(c.Request.Context(), id)
if nerr == nil && len(notes) > 0 {
doc = notes[len(notes)-1].Text
} else {
writeJSON(c, 404, gin.H{"error": "not found"})
return
}
}
c.Header("Content-Type", "text/markdown; charset=utf-8")
c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"receptionist-%s.md\"", id.String()))
c.String(http.StatusOK, "%s", doc)
}

// DELETE /sessions/:id
func (h *Handler) Delete(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil {
writeJSON(c, 400, gin.H{"code": "INVALID_SESSION_ID"})
return
}
if err := h.store.DeleteSession(c.Request.Context(), id); err != nil {
writeJSON(c, 500, gin.H{"error": err.Error()})
return
}
writeJSON(c, 200, gin.H{"status": "deleted"})
}

// GET /sessions/:id/notes (legacy name - ListNotes handler_polish.go me hai)
func (h *Handler) GetNotes(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil {
writeJSON(c, 400, gin.H{"code": "INVALID_SESSION_ID"})
return
}
notes, _ := h.store.ListNotes(c.Request.Context(), id)
writeJSON(c, 200, gin.H{"notes": notes})
}

// GET /sessions/:id/events - snapshot + keep-alive SSE (durable replay SSE
// ka full version handler_events.go ki StreamEvents me hai).
func (h *Handler) Events(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil {
writeJSON(c, 400, gin.H{"code": "INVALID_SESSION_ID"})
return
}
c.Header("Content-Type", "text/event-stream")
c.Header("Cache-Control", "no-cache")
c.Header("Connection", "keep-alive")
flusher, ok := c.Writer.(http.Flusher)
if !ok {
writeJSON(c, 500, gin.H{"error": "SSE not supported"})
return
}
sess, cps, notes, _ := h.orch.GetSessionWithDetails(c.Request.Context(), id)
payload, _ := json.Marshal(map[string]interface{}{"session": sess, "checkpoints": cps, "notes": notes, "seq": time.Now().UnixNano()})
fmt.Fprintf(c.Writer, "event: snapshot\ndata: %s\n\n", string(payload))
flusher.Flush()
ticker := time.NewTicker(15 * time.Second)
defer ticker.Stop()
notify := c.Request.Context().Done()
for {
select {
case <-notify:
return
case <-ticker.C:
fmt.Fprintf(c.Writer, ": keep-alive %d\n\n", time.Now().Unix())
flusher.Flush()
}
}
}

// POST /sessions/:id/checkpoints/:cid/reopen
// :cid yahan section_key hai (auth, db, scale ...)
func (h *Handler) Reopen(c *gin.Context) {
sessionID, err := uuid.Parse(c.Param("id"))
if err != nil {
writeJSON(c, 400, gin.H{"code": "INVALID_SESSION_ID"})
return
}
section := c.Param("cid")
var body struct {
Reason string `json:"reason"`
}
_ = json.NewDecoder(c.Request.Body).Decode(&body)
cp, err := h.store.ReopenCheckpoint(c.Request.Context(), sessionID, section, body.Reason)
if err != nil {
writeJSON(c, 400, gin.H{"error": err.Error()})
return
}
writeJSON(c, 200, cp)
}