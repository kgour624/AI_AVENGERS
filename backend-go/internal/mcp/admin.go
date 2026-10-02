package mcp

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/response"
)

// AdminHandler is the in-app surface for MCP tokens.
//
// WHY it exists: minting a token from a CLI is fine for an engineer and
// impossible for the person who actually runs this product. Without this, every
// new developer machine would need shell access to the server. The handler is
// mounted under the admin group (admin role required) and never exposes a token
// after creation.
type AdminHandler struct {
	tokens *PGTokenStore
	logger *zap.Logger
}

// NewAdminHandler builds the handler over the token store.
func NewAdminHandler(tokens *PGTokenStore, logger *zap.Logger) *AdminHandler {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &AdminHandler{tokens: tokens, logger: logger}
}

// ListMCPTokens GET /admin/mcp-tokens
func (h *AdminHandler) ListMCPTokens(c *gin.Context) {
	records, err := h.tokens.List(c.Request.Context())
	if err != nil {
		h.logger.Error("list mcp tokens failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, records)
}

// CreateMCPToken POST /admin/mcp-tokens
//
// Returns the plaintext token in this reply and never again — the store keeps
// only its hash, so this is the one moment it can be copied.
func (h *AdminHandler) CreateMCPToken(c *gin.Context) {
	var req struct {
			Label     string   `json:"label"`
			Domains   []string `json:"domains"`
			ExpertIDs []string `json:"expert_ids"`
			Tools     []string `json:"tools"`
		}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_REQUEST", "label is required")
		return
	}
	req.Label = strings.TrimSpace(req.Label)
	if req.Label == "" {
		// The label is how a human recognises the token in the audit log later;
		// an unlabelled token is one nobody dares revoke.
		response.BadRequest(c, "INVALID_REQUEST", "label is required (e.g. \"sneha laptop\")")
		return
	}

	userID, _ := c.Get("user_id")
	ownerID, ok := userID.(uuid.UUID)
	if !ok {
		response.Unauthorized(c, "no user in context")
		return
	}

	raw, record, err := h.tokens.Create(c.Request.Context(), ownerID.String(), req.Label, req.Domains, req.ExpertIDs, req.Tools)
	if err != nil {
		h.logger.Error("create mcp token failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	h.logger.Info("mcp token created", zap.String("token_id", record.ID), zap.String("label", record.Label))
	response.Created(c, gin.H{"token": raw, "record": record})
}

// RevokeMCPToken POST /admin/mcp-tokens/:id/revoke
//
// Revoking is idempotent from the caller's point of view but truthful about
// what happened: an already-revoked or unknown id is NOT_FOUND rather than a
// silent success, so the UI can tell the truth instead of always showing green.
func (h *AdminHandler) RevokeMCPToken(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		response.BadRequest(c, "INVALID_ID", "token id is required")
		return
	}
	if err := h.tokens.Revoke(c.Request.Context(), id); err != nil {
		var toolErr *ToolError
		if errors.As(err, &toolErr) {
			response.NotFound(c, "no active token with that id")
			return
		}
		h.logger.Error("revoke mcp token failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	h.logger.Info("mcp token revoked", zap.String("token_id", id))
	response.OK(c, gin.H{"status": "revoked"})
}
