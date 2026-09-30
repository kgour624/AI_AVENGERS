package mcpv2

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler is API layer — only HTTP/SSE, delegates to Service (RULE 8-A:28, SRP).
type Handler struct {
	svc    *Service
	logger *zap.Logger
}

func NewHandler(svc *Service, logger *zap.Logger) *Handler {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Handler{svc: svc, logger: logger}
}

// RegisterRoutes mounts MCP-V2 isolated routes — composition only, no edit in old mcp (R2).
// Call from cmd/server/main.go: mcpv2.RegisterRoutes(r, deps) — this line will be in §6 Pending until you approve.
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	// OAuth discovery + CORS (MCP 7&8 spec)
	r.GET("/.well-known/oauth-protected-resource", withCors(h.handleDiscovery))
	r.GET("/.well-known/oauth-protected-resource/mcp", withCors(h.handleDiscovery))
	// MCP Streamable HTTP + SSE (POST /mcp/v2)
	mcp := r.Group("/mcp/v2")
	mcp.POST("", h.handleMCP)
	mcp.GET("/sse", h.handleSSE)
}

// withCors adds CORS headers for Inspector (MCP 7&8 — withCors pattern).
func withCors(next gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "MCP-Protocol-Version, Authorization, Content-Type")
		if c.Request.Method == http.MethodOptions {
			c.Status(http.StatusNoContent)
			return
		}
		next(c)
	}
}

func (h *Handler) handleDiscovery(c *gin.Context) {
	origin := schemeHost(c)
	c.JSON(http.StatusOK, gin.H{
		"resource":               origin + "/mcp/v2",
		"authorization_servers":  []string{origin},
		"bearer_methods_supported": []string{"header"},
		"scopes_supported":       []string{"expert:read", "chunk:read", "expert:write", "admin:write"},
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
	JSONRPC string `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  any    `json:"result,omitempty"`
	Error   *rpcErr `json:"error,omitempty"`
}
type rpcErr struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (h *Handler) handleMCP(c *gin.Context) {
	// 401 if no Authorization (OAuth 2.1 — WWW-Authenticate)
	if c.GetHeader("Authorization") == "" {
		c.Header("WWW-Authenticate", fmt.Sprintf(`Bearer realm="AI Avengers", resource_metadata="%s/.well-known/oauth-protected-resource/mcp"`, schemeHost(c)))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing_token"})
		return
	}
	var req rpcReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, rpcResp{JSONRPC: "2.0", Error: &rpcErr{Code: -32700, Message: err.Error()}})
		return
	}
	result, rpcErr := h.dispatch(c, req)
	resp := rpcResp{JSONRPC: "2.0", ID: req.ID, Result: result, Error: rpcErr}
	c.JSON(http.StatusOK, resp)
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

func (h *Handler) dispatch(c *gin.Context, req rpcReq) (any, *rpcErr) {
	switch req.Method {
	case "initialize":
		return gin.H{"protocolVersion": "2024-11-05", "capabilities": gin.H{"tools": gin.H{}, "resources": gin.H{}, "prompts": gin.H{}}}, nil
	case "tools/list":
		return gin.H{"tools": []gin.H{
			{"name": "list_experts", "title": "List Experts", "description": "List active domain experts (robots)", "inputSchema": gin.H{"type": "object", "properties": gin.H{}}},
			{"name": "get_expert", "title": "Get Expert", "description": "Get expert charter by id", "inputSchema": gin.H{"type": "object", "properties": gin.H{"id": gin.H{"type": "string"}}} , "annotations": gin.H{"readOnlyHint": true}},
			{"name": "search_chunks", "title": "Search Chunks", "description": "Expert-scoped pgvector search. Use before ask_expert.", "inputSchema": gin.H{"type": "object", "properties": gin.H{"expertId": gin.H{"type": "string"}, "query": gin.H{"type": "string"}, "topK": gin.H{"type": "number"}, "cursor": gin.H{"type": "string"}}} , "annotations": gin.H{"readOnlyHint": true, "openWorldHint": false}},
			{"name": "ask_expert", "title": "Ask Expert", "description": "Ask expert with cited chunks. China Wall enforced.", "inputSchema": gin.H{"type": "object", "properties": gin.H{"expertId": gin.H{"type": "string"}, "question": gin.H{"type": "string"}}} , "annotations": gin.H{"destructiveHint": false, "idempotentHint": true}},
		}}, nil
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
	default:
		return nil, &rpcErr{Code: -32601, Message: "method not found: " + req.Method}
	}
}

func (h *Handler) handleToolCall(c *gin.Context, name string, args map[string]any) (any, *rpcErr) {
	// Kill Switch check (Gap 4 — handle once, fail fast)
	if IsKilled() {
		return nil, &rpcErr{Code: -32099, Message: "mcp_disabled"}
	}
	ctx := c.Request.Context()
	switch name {
	case "list_experts":
		experts, err := h.svc.ListExperts(ctx)
		if err != nil {
			return nil, h.toRpcErr(err)
		}
		links := make([]gin.H, 0, len(experts))
		for _, e := range experts {
			links = append(links, gin.H{"type": "resource_link", "uri": "expert://experts/" + e.ID, "name": "🤖 " + e.Name})
		}
		return gin.H{"content": []gin.H{{"type": "text", "text": fmt.Sprintf("found %d experts", len(experts))}}, "structuredContent": gin.H{"experts": experts}, "_meta": gin.H{"links": links}}, nil
	case "get_expert":
		id, _ := args["id"].(string)
		e, err := h.svc.GetExpert(ctx, id)
		if err != nil {
			return nil, h.toRpcErr(err)
		}
		return gin.H{"content": []gin.H{{"type": "text", "text": e.Charter}}, "structuredContent": gin.H{"expert": e}}, nil
	case "search_chunks":
		expertID, _ := args["expertId"].(string)
		query, _ := args["query"].(string)
		cursor, _ := args["cursor"].(string)
		limit := 20
		if v, ok := args["topK"].(float64); ok {
			limit = int(v)
		}
		if cursor != "" {
			if _, err := base64.StdEncoding.DecodeString(cursor); err != nil {
				return gin.H{"content": []gin.H{{"type": "text", "text": "invalid_cursor"}}, "isError": true}, nil
			}
		}
		chunks, nextCursor, err := h.svc.SearchChunks(ctx, expertID, query, limit, cursor)
		if err != nil {
			return nil, h.toRpcErr(err)
		}
		return gin.H{"content": []gin.H{{"type": "text", "text": fmt.Sprintf("found %d chunks", len(chunks))}}, "structuredContent": gin.H{"chunks": chunks, "nextCursor": nextCursor}}, nil
	case "ask_expert":
		expertID, _ := args["expertId"].(string)
		question, _ := args["question"].(string)
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
			return nil, h.toRpcErr(err)
		}
		return gin.H{"content": []gin.H{{"type": "text", "text": answer}}, "structuredContent": gin.H{"answer": answer, "citations": citations}}, nil
	default:
		return nil, &rpcErr{Code: -32601, Message: "tool not found: " + name}
	}
}