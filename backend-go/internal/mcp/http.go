package mcp

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

// maxRequestBytes caps one request body. Tool payloads carry a diff, not a
// repository, so this is generous while still refusing an unbounded upload.
const maxRequestBytes = 2 << 20 // 2 MiB

// ServeHTTP answers MCP over HTTP.
//
// WHY this transport exists: a stdio session only works when the client can
// launch a process, which rules out anything hosted. HTTP lets a team point
// their coding agent at one URL with a per-user token, and it is where the
// token scope and the audit trail become meaningful — a stdio session belongs to
// whoever started the process and has no identity to attribute calls to.
//
// Every request is authorised BEFORE it reaches the registry: an unknown or
// revoked token never gets as far as listing tools.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		// MCP over HTTP is a POST endpoint; anything else is a client mistake
		// rather than a server fault.
		w.Header().Set("Allow", http.MethodPost)
		writeHTTPJSON(w, http.StatusMethodNotAllowed, map[string]any{
			"jsonrpc": "2.0",
			"error":   jsonRPCError{Code: -32600, Message: "method not allowed"},
		})
		return
	}
	if s.tokens == nil {
		// Refusing is the honest answer: a server with no token store cannot
		// authenticate anyone, and running unauthenticated would expose every
		// expert to whoever finds the URL.
		writeHTTPJSON(w, http.StatusServiceUnavailable, map[string]any{
			"jsonrpc": "2.0",
			"error":   jsonRPCError{Code: -32603, Message: "this server has no token store configured"},
		})
		return
	}

	scope, info, err := s.tokens.Resolve(r.Context(), bearerToken(r))
	if err != nil {
		status := http.StatusUnauthorized
		var toolErr *ToolError
		if errors.As(err, &toolErr) && toolErr.Code == "FORBIDDEN" {
			status = http.StatusUnauthorized
		}
		writeHTTPJSON(w, status, map[string]any{
			"jsonrpc": "2.0",
			"error":   jsonRPCError{Code: -32001, Message: err.Error()},
		})
		return
	}
	scope.TokenID = info.ID

	body, err := io.ReadAll(io.LimitReader(r.Body, maxRequestBytes+1))
	if err != nil {
		writeHTTPJSON(w, http.StatusBadRequest, map[string]any{
			"jsonrpc": "2.0",
			"error":   jsonRPCError{Code: -32700, Message: "could not read request"},
		})
		return
	}
	if len(body) > maxRequestBytes {
		writeHTTPJSON(w, http.StatusRequestEntityTooLarge, map[string]any{
			"jsonrpc": "2.0",
			"error":   jsonRPCError{Code: -32600, Message: "request too large"},
		})
		return
	}

	var req jsonRPCRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeHTTPJSON(w, http.StatusBadRequest, map[string]any{
			"jsonrpc": "2.0",
			"error":   jsonRPCError{Code: -32700, Message: "parse error: " + err.Error()},
		})
		return
	}

	// The request context is the client's: if the agent disconnects, the expert
	// answer is cancelled with it instead of running on for nobody.
	resp, replyDue := s.dispatch(r.Context(), scope, req)

	// Usage is recorded after the call so a bookkeeping failure cannot fail it.
	if touchErr := s.tokens.Touch(r.Context(), info.ID); touchErr != nil {
		s.logger.Warn("mcp token touch failed", zap.Error(touchErr))
	}

	if !replyDue {
		w.WriteHeader(http.StatusAccepted)
		return
	}
	writeHTTPJSON(w, http.StatusOK, resp)
}

// bearerToken reads the token from the Authorization header. The header is the
// only accepted source: a token in a query string ends up in access logs and
// browser history.
func bearerToken(r *http.Request) string {
	header := strings.TrimSpace(r.Header.Get("Authorization"))
	const prefix = "Bearer "
	if len(header) > len(prefix) && strings.EqualFold(header[:len(prefix)], prefix) {
		return strings.TrimSpace(header[len(prefix):])
	}
	return ""
}

// writeHTTPJSON writes one reply with the status the transport decided.
func writeHTTPJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
