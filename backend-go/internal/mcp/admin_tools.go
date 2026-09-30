package mcp

import (
    "encoding/json"
    "regexp"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

var toolNameRe = regexp.MustCompile(`^[a-z][a-z0-9_]{2,64}$`)

type AdminToolHandler struct { store *ToolStore }

func NewAdminToolHandler(store *ToolStore) *AdminToolHandler { return &AdminToolHandler{store: store} }

// GET /admin/mcp-tools -> dropdown ke liye active tools
func (h *AdminToolHandler) ListTools(c *gin.Context) {
    tools, err := h.store.ListActive(c.Request.Context())
    if err != nil { c.JSON(500, gin.H{"error": err.Error()}); return }
    c.JSON(200, gin.H{"data": tools})
}
// GET /admin/mcp-tool-definitions -> manage table ke liye all
func (h *AdminToolHandler) ListAllTools(c *gin.Context) {
    tools, err := h.store.ListAll(c.Request.Context())
    if err != nil { c.JSON(500, gin.H{"error": err.Error()}); return }
    c.JSON(200, gin.H{"data": tools})
}
func (h *AdminToolHandler) CreateTool(c *gin.Context) {
    var req ToolDefinition
    if err := c.ShouldBindJSON(&req); err != nil { c.JSON(400, gin.H{"error": err.Error()}); return }
    req.Name = strings.ToLower(strings.TrimSpace(req.Name))
    if !toolNameRe.MatchString(req.Name) { c.JSON(400, gin.H{"error": "invalid name, use ^[a-z][a-z0-9_]{2,64}$"}); return }
    if req.HandlerKey == "" { req.HandlerKey = req.Name }
    if len(req.InputSchema)==0 { req.InputSchema = json.RawMessage(`{"type":"object","properties":{}}`) }
    created, err := h.store.Create(c.Request.Context(), req)
    if err != nil { c.JSON(409, gin.H{"error": "tool already exists or db error: "+err.Error()}); return }
    c.JSON(201, gin.H{"data": created})
}
func (h *AdminToolHandler) UpdateTool(c *gin.Context) {
    id, _ := uuid.Parse(c.Param("id"))
    var req ToolDefinition
    if err := c.ShouldBindJSON(&req); err != nil { c.JSON(400, gin.H{"error": err.Error()}); return }
    if err := h.store.Update(c.Request.Context(), id, req); err != nil { c.JSON(500, gin.H{"error": err.Error()}); return }
    c.JSON(200, gin.H{"status": "updated"})
}
func (h *AdminToolHandler) DeleteTool(c *gin.Context) {
    id, _ := uuid.Parse(c.Param("id"))
    if err := h.store.Delete(c.Request.Context(), id); err != nil { c.JSON(500, gin.H{"error": err.Error()}); return }
    c.JSON(200, gin.H{"status": "deleted"})
}
