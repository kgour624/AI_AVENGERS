package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"go.uber.org/zap"
)

// dispatch answers one request for a scope and reports whether a reply is due
// (notifications get none).
//
// WHY one function for both transports: stdio and HTTP differ only in how bytes
// arrive and leave. Keeping the protocol logic here means a fix — a new method,
// a changed error mapping — lands once, and the two transports cannot drift
// apart. It is also where the audit entry is written, because this is the one
// place that sees the tool, the outcome and the latency.
func (s *Server) dispatch(ctx context.Context, scope Scope, req jsonRPCRequest) (jsonRPCResponse, bool) {
	notification := len(req.ID) == 0

	switch req.Method {
	case "initialize":
		if notification {
			return jsonRPCResponse{}, false
		}
		return jsonRPCResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{
			"protocolVersion": protocolVersion,
			"capabilities":    map[string]any{"tools": map[string]any{}},
			"serverInfo":      map[string]any{"name": "ai-avengers-experts", "version": s.version},
		}}, true

	case "notifications/initialized", "initialized":
		// Client acknowledgement. Nothing to do, nothing to reply.
		return jsonRPCResponse{}, false

	case "ping":
		if notification {
			return jsonRPCResponse{}, false
		}
		return jsonRPCResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{}}, true

	case "tools/list":
		if notification {
			return jsonRPCResponse{}, false
		}
		return jsonRPCResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{
			"tools": s.registry.List(scope),
		}}, true

	case "tools/call":
		if notification {
			return jsonRPCResponse{}, false
		}
		return s.dispatchToolCall(ctx, scope, req), true

	default:
		if notification {
			return jsonRPCResponse{}, false
		}
		return jsonRPCResponse{JSONRPC: "2.0", ID: req.ID,
			Error: &jsonRPCError{Code: -32601, Message: "method not found: " + req.Method}}, true
	}
}

// dispatchToolCall runs one tool and turns its outcome into a reply and an audit
// record.
//
// A ToolError is the caller's mistake (bad domain, forbidden tool, unknown
// expert) and is reported as a tool-level error the client can act on. Anything
// else is a server fault: it is logged with full detail while the client only
// learns that the call failed (never leak internals).
func (s *Server) dispatchToolCall(ctx context.Context, scope Scope, req jsonRPCRequest) jsonRPCResponse {
	var params struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return jsonRPCResponse{JSONRPC: "2.0", ID: req.ID,
			Error: &jsonRPCError{Code: -32602, Message: "invalid params: " + err.Error()}}
	}

	entry := AuditEntry{
		TokenID:     scope.TokenID,
		TokenLabel:  scope.Label,
		Tool:        params.Name,
		Domain:      domainFromArgs(params.Arguments),
		InputBytes:  len(params.Arguments),
		InputSHA256: HashInput(params.Arguments),
	}

	start := time.Now()
	result, err := s.registry.Call(ctx, scope, params.Name, params.Arguments)
	entry.LatencyMs = int(time.Since(start).Milliseconds())

	if err != nil {
		var toolErr *ToolError
		if errors.As(err, &toolErr) {
			entry.Status = "tool_error"
			entry.ErrorCode = toolErr.Code
			s.audit(entry)
			s.logger.Info("mcp tool rejected",
				zap.String("tool", params.Name),
				zap.String("code", toolErr.Code),
			)
			// isError keeps this a successful JSON-RPC response: MCP models a
			// tool-level failure this way, and the client needs to see why.
			return s.toolErrorResponse(req.ID, toolErr.Message)
		}
		entry.Status = "server_error"
		s.audit(entry)
		s.logger.Error("mcp tool failed", zap.String("tool", params.Name), zap.Error(err))
		return s.toolErrorResponse(req.ID, "The tool failed on the server.")
	}

	entry.Status = "ok"
	entry.ErrorCode = ""
	s.audit(entry)
	return jsonRPCResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{
		"content": []map[string]any{{"type": "text", "text": result.Text}},
		"isError": false,
	}}
}

// audit records the entry when auditing is enabled for this server.
func (s *Server) audit(entry AuditEntry) {
	if s.auditor == nil {
		return
	}
	s.auditor.Record(entry)
}

// toolErrorResponse builds the MCP tool-level failure reply.
func (s *Server) toolErrorResponse(id json.RawMessage, message string) jsonRPCResponse {
	return jsonRPCResponse{JSONRPC: "2.0", ID: id, Result: map[string]any{
		"content": []map[string]any{{"type": "text", "text": message}},
		"isError": true,
	}}
}

// domainFromArgs reads the domain out of a tool payload for the audit trail.
// A payload without one (list_experts) simply records an empty domain.
func domainFromArgs(args json.RawMessage) string {
	if len(args) == 0 {
		return ""
	}
	var probe struct {
		Domain string `json:"domain"`
	}
	if err := json.Unmarshal(args, &probe); err != nil {
		return ""
	}
	return probe.Domain
}
