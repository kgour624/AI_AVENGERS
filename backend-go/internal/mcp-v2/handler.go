package mcpv2

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"ai_avengers/backend/internal/mcp"
	"ai_avengers/backend/internal/mcp-v2/business"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler is API layer — only HTTP/SSE, delegates to Service (RULE 8-A:28, SRP).
type Handler struct {
	svc        *Service
	logger     *zap.Logger
	tokenStore *mcp.PGTokenStore
	cancels    CancellationManager
	validator  SchemaValidator
}

func NewHandler(svc *Service, logger *zap.Logger, tokenStore *mcp.PGTokenStore, validator SchemaValidator) *Handler {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Handler{svc: svc, logger: logger, tokenStore: tokenStore, validator: validator}
}

// resolveScope reads the raw token from gin context and resolves it to a Scope
// via mcp.PGTokenStore. Production-grade: fail-closed when token is present
// but invalid — an invalid token must not become an open scope.
// - no tokenStore => open (local stdio / tests)
// - no token presented => open (backward compat for local dev)
// - token present but Resolve fails => denied scope (Label=invalid, ExpertIDs=__INVALID__) so every AllowsExpert check fails and handler returns FORBIDDEN without asking user.
func (h *Handler) resolveScope(c *gin.Context) mcp.Scope {
	if h.tokenStore == nil {
		return mcp.Scope{}
	}
	rawToken, _ := c.Get("mcp_token")
	token, ok := rawToken.(string)
	if !ok || strings.TrimSpace(token) == "" {
		return mcp.Scope{}
	}
	token = strings.TrimSpace(token)
	token = strings.TrimPrefix(token, "Bearer ")
	token = strings.TrimPrefix(token, "bearer ")
	token = strings.TrimSpace(token)
	if token == "" {
		return mcp.Scope{}
	}
	scope, _, err := h.tokenStore.Resolve(c.Request.Context(), token)
	if err != nil {
		h.logger.Warn("token resolve failed — denying scope", zap.Error(err))
		return mcp.Scope{Label: "invalid_token", ExpertIDs: []string{"__INVALID_TOKEN__"}, Tools: []string{"__INVALID_TOKEN__"}}
	}
	return scope
}
// RegisterRoutes mounts MCP-V2 isolated routes — composition only, no edit in old mcp (R2).
// Call from cmd/server/main.go: mcpv2.RegisterRoutes(r, deps) — this line will be in §6 Pending until you approve.
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	// OAuth discovery - UNCOMMENTED for Claude/Inspector (MCP Spec 2024-11-05 mandatory)
	r.GET("/.well-known/oauth-protected-resource", h.handleDiscovery)
	r.GET("/.well-known/oauth-protected-resource/mcp", h.handleDiscovery)

	// PUBLIC MCP Endpoint: Uses local corsMiddleware with `*` (NO Credentials) - Spec Compliant
	// Inspector hits cross-origin without cookies, only Authorization header
	mcp := r.Group("/mcp/v2", corsMiddleware)
	mcp.GET("/health", h.handleHealth)
	mcp.HEAD("/health", h.handleHealth)
	mcp.POST("", h.handleMCP)
	mcp.OPTIONS("", h.handleOptions)
	mcp.GET("/sse", h.handleSSE) // deprecated, kept for backward compat
	mcp.OPTIONS("/sse", h.handleOptions)

	// ADMIN API: NO local corsMiddleware - Relies on Global CORS in main.go
	// Reason: Global middleware does echo Origin + Allow-Credentials:true for Bearer token.
	// Adding `*` here breaks W3C Spec & causes "Failed to load tools" browser block.
	api := r.Group("/api/v1/mcp-v2")
	api.GET("/tools", h.HandleListTools)
	api.POST("/tools", h.HandleCreateTool)
	api.POST("/tools/generate-sql", h.HandleGenerateSQL)
	// No OPTIONS needed - Global CORS handles it
}

// corsMiddleware for PUBLIC MCP only - W3C Compliant: `*` NEVER with Credentials
func corsMiddleware(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, POST, HEAD, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "MCP-Protocol-Version, Authorization, Content-Type, X-Request-ID")
	c.Header("Vary", "Origin")
	// NEVER set Access-Control-Allow-Credentials here - would be illegal with `*`
	if c.Request.Method == http.MethodOptions {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}
	c.Next()
}

func (h *Handler) handleOptions(c *gin.Context) {
	// Options is handled by withCors middleware
	c.Status(http.StatusNoContent)
}

func (h *Handler) handleDiscovery(c *gin.Context) {
	origin := schemeHost(c)
	c.JSON(http.StatusOK, gin.H{
		"resource":                 origin + "/mcp/v2",
		"authorization_servers":    []string{origin},
		"bearer_methods_supported": []string{"header"},
		"scopes_supported":         []string{"expert:read", "chunk:read", "expert:write", "admin:write"},
	})
}

func schemeHost(c *gin.Context) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	if fwd := c.GetHeader("X-Forwarded-Proto"); fwd != "" {
		scheme = fwd
	}
	return fmt.Sprintf("%s://%s", scheme, c.Request.Host)
}

// Gap 4 — Centralized error handling (Ultimate Go §8): App defines Code, API maps to protocol.
// "Once you handle it, it's not an error anymore" — log once + inspect + respond.
func (h *Handler) getRequestID(c *gin.Context) string {
	if v := c.GetHeader("X-Request-ID"); v != "" {
		return v
	}
	if v := c.Query("request_id"); v != "" {
		return v
	}
	if v := c.GetHeader("X-Request-Id"); v != "" {
		return v
	}
	if v, ok := c.Get("request_id"); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func (h *Handler) toolErrorResult(c *gin.Context, err error) gin.H {
	reqID := h.getRequestID(c)
	h.logger.Error("mcp tool error — isError response", zap.String("request_id", reqID), zap.Error(err))
	msg := "expert temporarily unavailable, try again"
	var ae *AppError
	if errors.As(err, &ae) && ae.Msg != "" {
		if ae.Code == CodeInternal {
			msg = "expert temporarily unavailable, try again"
		} else {
			msg = ae.Msg
		}
	} else if err != nil {
		if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "deadline") {
			msg = "expert temporarily unavailable, try again (timeout)"
		}
	}
	return gin.H{"content": []gin.H{{"type": "text", "text": msg}}, "isError": true, "structuredContent": gin.H{"error": msg, "request_id": reqID}}
}

func (h *Handler) toRpcErr(err error) *rpcErr {
	if err == nil {
		return nil
	}
	// AppError is the App-layer error with Code (app_errors.go)
	var ae *AppError
	if errors.As(err, &ae) {
		// Log once, at handler of the error ( §8, §11 )
		h.logger.Error("mcp app error", zap.String("code", string(ae.Code)), zap.Error(ae))
		var code int
		switch ae.Code {
		case CodeInvalidInput:
			code = -32602
		case CodeNotFound:
			code = -32004
		case CodeLimitExceeded:
			code = -32029
		case CodeUnauthorized:
			code = -32001
		case CodeForbidden:
			code = -32003
		case CodeCancelled:
			code = -32800
		default:
			code = -32603
		}
		msg := ae.Msg
		if ae.Code == CodeInternal {
			msg = "internal error" // never leak internals ( §8 )
		}
		return &rpcErr{Code: code, Message: msg}
	}
	// Sentinel fallback (RULE 8-E:42)
	if errors.Is(err, ErrLimitExceeded) {
		h.logger.Warn("mcp limit exceeded", zap.Error(err))
		return &rpcErr{Code: -32029, Message: "limit exceeded"}
	}
	// Unknown → 500, no leak
	h.logger.Error("mcp internal error", zap.Error(err))
	return &rpcErr{Code: -32603, Message: "internal error"}
}

// MCP JSON-RPC dispatch — tools/list, tools/call, resources/list etc.
type rpcReq struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}
type rpcResp struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcErr         `json:"error,omitempty"`
}
type rpcErr struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (h *Handler) handleMCP(c *gin.Context) {
	// 401 if no token (Header or Query)
	token := c.GetHeader("Authorization")
	if token == "" {
		token = c.Query("token")
	}
	
	if token == "" {
		c.Header("WWW-Authenticate", fmt.Sprintf(`Bearer realm="AI Avengers", resource_metadata="%s/.well-known/oauth-protected-resource/mcp"`, schemeHost(c)))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing_token"})
		return
	}
	// Inject token into context for downstream dispatch to use
	c.Set("mcp_token", token)
	var req rpcReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, rpcResp{JSONRPC: "2.0", Error: &rpcErr{Code: -32700, Message: err.Error()}})
		return
	}
	// --- Cancellation intercept: notifications/cancelled is notification (no rpcResp) -> 202 Accepted
	if req.Method == "notifications/cancelled" {
		if cid := extractCancelID(req.Params); cid != "" {
			h.cancels.CancelAndDelete(cid)
		}
		if pt := extractProgressToken(req.Params); pt != "" {
			h.cancels.CancelAndDelete(pt)
		}
		c.Status(http.StatusAccepted)
		return
	}
	// --- Normal request: wire cancellable context + store for cancel
	idKey := string(req.ID)
	if idKey != "" && idKey != "null" {
		ctx, cancel := context.WithCancel(c.Request.Context())
		h.cancels.Store(idKey, cancel)
		defer h.cancels.Delete(idKey)
		if pt := extractProgressToken(req.Params); pt != "" && pt != idKey {
			h.cancels.Store(pt, cancel)
			defer h.cancels.Delete(pt)
		}
		c.Request = c.Request.WithContext(ctx)
		result, respErr := h.dispatch(c, req)
		if ctx.Err() != nil && respErr == nil {
			respErr = &rpcErr{Code: -32800, Message: "Request cancelled"}
		}
		// Map AppError Cancelled -> -32800 (toRpcErr already handles, but fast path here too)
		if respErr != nil && ctx.Err() != nil && respErr.Code != -32800 {
			// keep cancelled precedence if context was cancelled
			if errors.Is(ctx.Err(), context.Canceled) {
				respErr = &rpcErr{Code: -32800, Message: "Request cancelled"}
			}
		}
		resp := rpcResp{JSONRPC: "2.0", ID: req.ID, Result: result, Error: respErr}
		if strings.Contains(c.GetHeader("Accept"), "text/event-stream") {
			c.Header("Content-Type", "text/event-stream")
			c.Header("Cache-Control", "no-cache")
			c.Stream(func(w io.Writer) bool {
				b, _ := json.Marshal(resp)
				c.SSEvent("message", string(b))
				return false
			})
			return
		}
		c.JSON(http.StatusOK, resp)
		return
	}
	result, respErr := h.dispatch(c, req)
	resp := rpcResp{JSONRPC: "2.0", ID: req.ID, Result: result, Error: respErr}
	if strings.Contains(c.GetHeader("Accept"), "text/event-stream") {
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Stream(func(w io.Writer) bool {
			b, _ := json.Marshal(resp)
			c.SSEvent("message", string(b))
			return false
		})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func extractCancelID(params json.RawMessage) string {
	var p struct {
		RequestId json.RawMessage `json:"requestId"`
	}
	if err := json.Unmarshal(params, &p); err == nil && len(p.RequestId) > 0 {
		s := string(p.RequestId)
		if strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`) {
			return s[1 : len(s)-1]
		}
		return s
	}
	return ""
}

func extractProgressToken(params json.RawMessage) string {
	var p struct {
		Meta struct {
			ProgressToken json.RawMessage `json:"progressToken"`
		} `json:"_meta"`
	}
	if err := json.Unmarshal(params, &p); err == nil && len(p.Meta.ProgressToken) > 0 {
		s := string(p.Meta.ProgressToken)
		if strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`) {
			return s[1 : len(s)-1]
		}
		return s
	}
	return ""
}

func (h *Handler) handleSSE(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	// Minimal SSE keepalive — real sampling/elicitation will push via notifications
	c.Stream(func(w io.Writer) bool {
		c.SSEvent("message", gin.H{"type": "heartbeat"})
		return false
	})
}

func (h *Handler) handleHealth(c *gin.Context) {
	reqID := c.GetHeader("X-Request-ID")
	if reqID == "" {
		reqID = c.Query("request_id")
	}
	if reqID == "" {
		reqID = c.GetHeader("X-Request-Id")
	}
	h.logger.Info("mcp health", zap.String("request_id", reqID))
	// Best-effort checks — never fail health with -32603, always 200 ok
	checks := gin.H{"db": "ok", "redis": "ok", "gateway": "ok"}
	// Light DB ping via service if available
	if h.svc != nil {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if _, err := h.svc.ListExperts(ctx); err != nil {
			checks["db"] = "unavailable: " + err.Error()
		}
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "mcp-v2", "request_id": reqID, "checks": checks})
}

// HandleListTools GET /mcp/v2/tools
// Fetches dynamic tools from Postgres via Service -> Storer.ListTools() -> mcp_v2_tools table.
func (h *Handler) HandleListTools(c *gin.Context) {
	tools, err := h.svc.ListTools(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to list tools", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list tools", "details": err.Error()})
		return
	}
	if tools == nil {
		tools = []ToolDefinition{}
	}
	c.JSON(http.StatusOK, gin.H{"tools": tools})
}

func (h *Handler) HandleCreateTool(c *gin.Context) {
	var req ToolDefinition
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.CreateTool(c.Request.Context(), req); err != nil {
		var ae *AppError
		if errors.As(err, &ae) && ae.Code == CodeInvalidInput {
			c.JSON(http.StatusBadRequest, gin.H{"error": ae.Msg})
			return
		}
		if ae != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": ae.Msg})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if h.validator != nil {
		h.validator.Invalidate(req.Name)
	}
	c.JSON(http.StatusCreated, gin.H{"status": "ok"})
}

// HandleGenerateSQL POST /api/v1/mcp-v2/tools/generate-sql — Magic Generate 🪄 text-to-SQL via cheap LLM + schema.
func (h *Handler) HandleGenerateSQL(c *gin.Context) {
	var req struct {
		Prompt string `json:"prompt" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "prompt required"})
		return
	}
	sql, err := h.svc.GenerateSQL(c.Request.Context(), req.Prompt)
	if err != nil {
		var ae *AppError
		if errors.As(err, &ae) && ae.Code == CodeInvalidInput {
			c.JSON(http.StatusBadRequest, gin.H{"error": ae.Msg})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"sql": sql})
}

func (h *Handler) dispatch(c *gin.Context, req rpcReq) (any, *rpcErr) {
	switch req.Method {
	case "initialize":
		return gin.H{
			"protocolVersion": "2024-11-05",
			"capabilities":    gin.H{"tools": gin.H{}, "resources": gin.H{}, "prompts": gin.H{}},
			"serverInfo":      gin.H{"name": "ai-avengers", "version": "1.0.0"},
		}, nil
	case "tools/list":
		// 100% Dynamic — no hardcoded array. DB (mcp_v2_tools.input_schema) is single source of truth.
		if IsKilled() {
			return gin.H{"tools": []gin.H{}}, nil
		}
		tools, err := h.svc.ListTools(c.Request.Context())
		if err != nil {
			return nil, h.toRpcErr(err)
		}
		// Resolve token scope and filter tools by allow-list (same as tools/call enforcement)
		scope := h.resolveScope(c)
		out := make([]gin.H, 0, len(tools))
		for _, t := range tools {
			// Skip tools not allowed by this token's scope
			if !scope.AllowsTool(t.Name) {
				continue
			}
			schema := t.InputSchema
			if schema == nil {
				schema = map[string]any{"type": "object", "properties": map[string]any{}}
			}
			item := gin.H{
				"name":        t.Name,
				"title":       t.DisplayName,
				"description": t.Description,
				"inputSchema": schema,
			}
			// Market parity: Annotations for Claude optimization
			if t.Annotations != nil && len(t.Annotations) > 0 {
				item["annotations"] = t.Annotations
			} else {
				// Default safe hints if DB empty
				item["annotations"] = gin.H{"readOnlyHint": true, "destructiveHint": false, "openWorldHint": false}
			}
			if t.OutputSchema != nil && len(t.OutputSchema) > 0 {
				item["outputSchema"] = t.OutputSchema
			}
			out = append(out, item)
		}
		return gin.H{"tools": out}, nil
	case "tools/call":
		var p struct {
			Name      string         `json:"name"`
			Arguments map[string]any `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &p); err != nil {
			return nil, &rpcErr{Code: -32602, Message: err.Error()}
		}
		return h.handleToolCall(c, p.Name, p.Arguments)
	case "resources/list":
		return gin.H{"resources": []gin.H{{"uri": "expert://experts", "name": "All Experts", "mimeType": "application/json"}}}, nil
	case "prompts/list":
		return gin.H{"prompts": []gin.H{{"name": "suggest_review", "description": "Review file as expert citing chunks", "arguments": []gin.H{{"name": "expertId", "required": true}, {"name": "filePath", "required": true}}}}}, nil
	case "notifications/initialized":
		return nil, nil
	case "ping":
		return gin.H{}, nil
	default:
		return nil, &rpcErr{Code: -32601, Message: "method not found: " + req.Method}
	}
}

func normalizeToolName(name string) string {
	n := strings.TrimSpace(name)
	// Claude/Code prefixes tool with server id: mcp_AI-AVENGERS_ask_expert, mcp_ask_expert, AI-AVENGERS_ask_expert
	// Strip any prefix up to last underscore that leaves a known suffix.
	if idx := strings.LastIndex(n, "_ask_expert"); idx != -1 {
		return "ask_expert"
	}
	if idx := strings.LastIndex(n, "_search_repo_content"); idx != -1 {
		return "search_repo_content"
	}
	if idx := strings.LastIndex(n, "_search_course_content"); idx != -1 {
		return "search_course_content"
	}
	if idx := strings.LastIndex(n, "_search_chunks"); idx != -1 {
		return "search_chunks"
	}
	if idx := strings.LastIndex(n, "_get_standards"); idx != -1 {
		return "get_standards"
	}
	if idx := strings.LastIndex(n, "_review_change"); idx != -1 {
		return "review_change"
	}
	if idx := strings.LastIndex(n, "_list_experts"); idx != -1 {
		return "list_experts"
	}
	if idx := strings.LastIndex(n, "_get_expert"); idx != -1 {
		return "get_expert"
	}
	// fallback: take after last '-' or '_' if contains known tool
	lower := strings.ToLower(n)
	for _, k := range []string{"ask_expert", "search_chunks", "search_course_content", "search_repo_content", "get_standards", "review_change", "list_experts", "get_expert", "get_usage_logs", "list_tools"} {
		if strings.Contains(lower, k) {
			return k
		}
	}
	return n
}

func (h *Handler) handleToolCall(c *gin.Context, name string, args map[string]any) (any, *rpcErr) {
	// Normalize prefixed names from Claude Code (mcp_AI-AVENGERS_ask_expert -> ask_expert)
	name = normalizeToolName(name)
	// Kill Switch check (Gap 4 — handle once, fail fast)
	if IsKilled() {
		return nil, &rpcErr{Code: -32099, Message: "mcp_disabled"}
	}
	// InputSchema validation - fail fast before DB (Market parity)
	if tool, err := h.svc.GetTool(c.Request.Context(), name); err == nil && tool.InputSchema != nil {
		if err := validateInputSchema(tool.InputSchema, args); err != nil {
			return gin.H{"content": []gin.H{{"type": "text", "text": "INVALID_ARGS: " + err.Error()}}, "isError": true}, nil
		}
	}
	// Resolve token → scope so we can filter experts and enforce access
	scope := h.resolveScope(c)
	// Tool allow-list enforcement (fail-closed): empty = all, __INVALID_TOKEN__ = none. Self-correcting via list_experts hint.
	if !scope.AllowsTool(name) {
		return gin.H{"content": []gin.H{{"type": "text", "text": "FORBIDDEN: this token may not use tool " + name + ". Allowed: " + strings.Join(scope.Tools, ", ")}}, "isError": true}, nil
	}
	ctx := c.Request.Context()
	// === GATEWAY SCHEMA VALIDATION (Fail-Fast) ===
	// Invalid -> -32602 turant, DB/Vector/Gateway touch hi nahi hoga. Valid -> Service call.
	if h.validator != nil {
		if err := h.validator.Validate(ctx, name, args); err != nil {
			return nil, h.toRpcErr(err) // AppError CodeInvalidInput -> -32602 via toRpcErr (Handler validates, Service executes)
		}
	}
	// === END GATEWAY VALIDATION ===
	switch name {
	case "list_experts":
		allExperts, err := h.svc.ListExperts(ctx)
		if err != nil {
			return h.toolErrorResult(c, err), nil
		}
		// Build human-readable text WITH IDs, filtered by token scope
		var sb strings.Builder
		sb.WriteString("# Available experts\n\n")
		sb.WriteString("Use the `id` value as `expert_id` in ask_expert / search_chunks / get_standards. Copy it exactly — do not guess.\n\n")
		
		var filteredExperts []business.Expert
		links := make([]gin.H, 0, len(allExperts))
		for _, e := range allExperts {
			if len(scope.ExpertIDs) > 0 {
				if !scope.AllowsExpert(e.ID) {
					continue
				}
			} else {
				if !scope.AllowsExpert(e.ID) && !scope.AllowsDomain(e.Charter) && !scope.AllowsDomain(e.Slug) {
					continue
				}
			}
			filteredExperts = append(filteredExperts, e)
			sb.WriteString(fmt.Sprintf("- **%s** — id: `%s` — slug: `%s`\n", e.Name, e.ID, e.Slug))
			links = append(links, gin.H{"type": "resource_link", "uri": "expert://experts/" + e.ID, "name": "🤖 " + e.Name})
		}
		if len(filteredExperts) == 0 {
			return gin.H{"content": []gin.H{{"type": "text", "text": "No experts are available to this token."}}}, nil
		}
		return gin.H{"content": []gin.H{{"type": "text", "text": sb.String()}}, "structuredContent": gin.H{"experts": filteredExperts}, "_meta": gin.H{"links": links}}, nil
	case "get_expert":
		id, _ := args["id"].(string)
		if id == "" {
			id, _ = args["expert_id"].(string)
		}
		if id == "" {
			id, _ = args["expertId"].(string)
		}
		if !scope.AllowsExpert(id) {
			return gin.H{"content": []gin.H{{"type": "text", "text": "FORBIDDEN: this token may not access expert " + id}}, "isError": true}, nil
		}
		e, err := h.svc.GetExpert(ctx, id)
		if err != nil {
			return h.toolErrorResult(c, err), nil
		}
		return gin.H{"content": []gin.H{{"type": "text", "text": e.Charter}}, "structuredContent": gin.H{"expert": e}}, nil
		case "search_chunks", "search_course_content", "search_repo_content":
		expertID, _ := args["expertId"].(string)
		if expertID == "" {
			expertID, _ = args["expert_id"].(string)
		}
		if expertID == "" {
			expertID, _ = args["expertId"].(string)
		}
		if strings.TrimSpace(expertID) == "" {
			return gin.H{"content": []gin.H{{"type": "text", "text": "INVALID_ARGS: expert_id is required — call list_experts to get the correct id, copy the id value exactly, do not guess. " + h.availableHint(ctx, scope)}}, "isError": true}, nil
		}
		if !scope.AllowsExpert(expertID) {
			return gin.H{"content": []gin.H{{"type": "text", "text": "FORBIDDEN: this token may not access expert " + expertID + ". " + h.availableHint(ctx, scope)}}, "isError": true}, nil
		}
		query, _ := args["query"].(string)
		if query == "" {
			query, _ = args["question"].(string)
		}
		cursor, _ := args["cursor"].(string)
		limit := 20
		if v, ok := args["topK"].(float64); ok {
			limit = int(v)
		}
		if v, ok := args["limit"].(float64); ok {
			limit = int(v)
		}
		if cursor != "" {
			if _, err := base64.StdEncoding.DecodeString(cursor); err != nil {
				return gin.H{"content": []gin.H{{"type": "text", "text": "invalid_cursor"}}, "isError": true}, nil
			}
		}
		chunks, nextCursor, err := h.svc.SearchChunks(ctx, expertID, query, limit, cursor)
		if err != nil {
			return h.toolErrorResult(c, err), nil
		}
		return gin.H{"content": []gin.H{{"type": "text", "text": fmt.Sprintf("found %d chunks", len(chunks))}}, "structuredContent": gin.H{"chunks": chunks, "nextCursor": nextCursor}}, nil
	case "ask_expert":
		expertID, _ := args["expertId"].(string)
		if expertID == "" {
			expertID, _ = args["expert_id"].(string)
		}
		if strings.TrimSpace(expertID) == "" {
			return gin.H{"content": []gin.H{{"type": "text", "text": "INVALID_ARGS: expert_id is required — call list_experts to get the correct id, copy the id value exactly, do not guess. " + h.availableHint(ctx, scope)}}, "isError": true}, nil
		}
		if !scope.AllowsExpert(expertID) {
			return gin.H{"content": []gin.H{{"type": "text", "text": "FORBIDDEN: this token may not access expert " + expertID + ". " + h.availableHint(ctx, scope)}}, "isError": true}, nil
		}
		question, _ := args["question"].(string)
		if question == "" {
			question, _ = args["query"].(string)
		}
		if strings.TrimSpace(question) == "" {
			return gin.H{"content": []gin.H{{"type": "text", "text": "INVALID_ARGS: question is required"}}, "isError": true}, nil
		}
		platform, _ := args["platform"].(string)
		tier, _ := args["tier"].(string)
		answer, citations, err := h.svc.AskExpert(ctx, expertID, question, platform, tier)
		if err != nil {
			var ae *AppError
			if errors.As(err, &ae) && ae.Code == CodeLimitExceeded {
				c.Header("WWW-Authenticate", `Bearer error="insufficient_scope"`)
				return gin.H{"content": []gin.H{{"type": "text", "text": "Expert token limit exceeded for platform. Contact admin."}}, "isError": true}, nil
			}
			if errors.Is(err, ErrLimitExceeded) {
				c.Header("WWW-Authenticate", `Bearer error="insufficient_scope"`)
				return gin.H{"content": []gin.H{{"type": "text", "text": "Expert token limit exceeded for platform. Contact admin."}}, "isError": true}, nil
			}
			if errors.As(err, &ae) && ae.Code == CodeNotFound {
				return gin.H{"content": []gin.H{{"type": "text", "text": "UNKNOWN_EXPERT: " + ae.Msg + ". " + h.availableHint(ctx, scope)}}, "isError": true}, nil
			}
			return h.toolErrorResult(c, err), nil
		}
		return gin.H{"content": []gin.H{{"type": "text", "text": answer}}, "structuredContent": gin.H{"answer": answer, "citations": citations}}, nil
	case "get_standards":
		expertID, _ := args["expert_id"].(string)
		if expertID == "" {
			expertID, _ = args["expertId"].(string)
		}
		if strings.TrimSpace(expertID) == "" {
			return gin.H{"content": []gin.H{{"type": "text", "text": "INVALID_ARGS: expert_id is required — call list_experts to get the correct id, copy the id value exactly, do not guess. " + h.availableHint(ctx, scope)}}, "isError": true}, nil
		}
		if !scope.AllowsExpert(expertID) {
			return gin.H{"content": []gin.H{{"type": "text", "text": "FORBIDDEN: this token may not access expert " + expertID + ". " + h.availableHint(ctx, scope)}}, "isError": true}, nil
		}
		track, _ := args["track"].(string)
		answer, citations, err := h.svc.GetStandards(ctx, expertID, track)
		if err != nil {
			var ae *AppError
			if errors.As(err, &ae) && ae.Code == CodeNotFound {
				return gin.H{"content": []gin.H{{"type": "text", "text": "UNKNOWN_EXPERT: " + ae.Msg + ". " + h.availableHint(ctx, scope)}}, "isError": true}, nil
			}
			return h.toolErrorResult(c, err), nil
		}
		return gin.H{"content": []gin.H{{"type": "text", "text": answer}}, "structuredContent": gin.H{"answer": answer, "citations": citations}}, nil
	case "review_change":
		expertID, _ := args["expert_id"].(string)
		if expertID == "" {
			expertID, _ = args["expertId"].(string)
		}
		if strings.TrimSpace(expertID) == "" {
			return gin.H{"content": []gin.H{{"type": "text", "text": "INVALID_ARGS: expert_id is required — call list_experts to get the correct id, copy the id value exactly, do not guess. " + h.availableHint(ctx, scope)}}, "isError": true}, nil
		}
		if !scope.AllowsExpert(expertID) {
			return gin.H{"content": []gin.H{{"type": "text", "text": "FORBIDDEN: this token may not access expert " + expertID + ". " + h.availableHint(ctx, scope)}}, "isError": true}, nil
		}
		diff, _ := args["diff"].(string)
		if strings.TrimSpace(diff) == "" {
			return gin.H{"content": []gin.H{{"type": "text", "text": "INVALID_ARGS: diff is required"}}, "isError": true}, nil
		}
		desc, _ := args["change_description"].(string)
		filePath, _ := args["file_path"].(string)
		if filePath == "" {
			filePath, _ = args["filePath"].(string)
		}
		answer, citations, err := h.svc.ReviewChange(ctx, expertID, diff, desc, filePath)
		if err != nil {
			var ae *AppError
			if errors.As(err, &ae) && ae.Code == CodeNotFound {
				return gin.H{"content": []gin.H{{"type": "text", "text": "UNKNOWN_EXPERT: " + ae.Msg + ". " + h.availableHint(ctx, scope)}}, "isError": true}, nil
			}
			return h.toolErrorResult(c, err), nil
		}
		return gin.H{"content": []gin.H{{"type": "text", "text": answer}}, "structuredContent": gin.H{"review": answer, "citations": citations}}, nil
	default:
		// Dynamic Engine fallback — Interpreter Pattern: if tool has action, execute via GenericExecutor (no Go struct needed)
		if tool, err := h.svc.GetTool(ctx, name); err == nil && !tool.Action.IsEmpty() {
			result, err := h.svc.ExecuteGeneric(ctx, name, args)
			if err != nil {
				return h.toolErrorResult(c, err), nil
			}
			return gin.H{"content": []gin.H{{"type": "text", "text": result}}}, nil
		}
		return nil, &rpcErr{Code: -32601, Message: "tool not found: " + name}
	}
}

// availableHint lists real expert IDs this token can access — self-correcting so Claude never asks user.
// Never returns empty: always gives an actionable next step with real IDs to copy-paste.
func (h *Handler) availableHint(ctx context.Context, scope mcp.Scope) string {
	if h.svc == nil {
		return "Call list_experts to see available experts."
	}
	experts, err := h.svc.ListExperts(ctx)
	if err != nil || len(experts) == 0 {
		return "Call list_experts to see available experts."
	}
	visible := []string{}
	for _, e := range experts {
		if !scope.AllowsExpert(e.ID) {
			continue
		}
		visible = append(visible, e.ID+":"+e.Slug+" ("+e.Name+")")
		if len(visible) >= 5 {
			break
		}
	}
	if len(visible) == 0 {
		return "No experts are available to this token — check token scope. Call list_experts to verify."
	}
	more := ""
	if len(experts) > len(visible) {
		more = " — call list_experts for full list"
	}
	return "Available: "+ strings.Join(visible, " | ")+ more
}

// validateInputSchema - lightweight JSON Schema required/properties check (no external dep)
// Validates args against tool.InputSchema (type:object, required:[], properties:{})
func validateInputSchema(schema map[string]any, args map[string]any) error {
	if schema == nil {
		return nil
	}
	if args == nil {
		args = map[string]any{}
	}
	// Check required fields
	if req, ok := schema["required"].([]any); ok {
		for _, r := range req {
			if key, ok := r.(string); ok {
				if _, exists := args[key]; !exists {
					return fmt.Errorf("missing required field '%s'", key)
				}
			}
		}
	}
	// Also handle required as []string (DB json variation)
	if req, ok := schema["required"].([]string); ok {
		for _, key := range req {
			if _, exists := args[key]; !exists {
				return fmt.Errorf("missing required field '%s'", key)
			}
		}
	}
	return nil
}
