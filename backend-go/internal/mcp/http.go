package mcp

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// maxRequestBytes caps one request body. Tool payloads carry a diff, not a
// repository, so this is generous while still refusing an unbounded upload.
const maxRequestBytes = 2 << 20 // 2 MiB

// sessionHeader is the header MCP's Streamable HTTP transport uses to tie a
// client's requests together after initialize.
const sessionHeader = "Mcp-Session-Id"

// protocolVersionHeader carries the version the client settled on.
const protocolVersionHeader = "MCP-Protocol-Version"

// sessionTTL bounds how long an idle session stays valid. A coding agent that
// goes quiet for an hour re-initializes, which costs it one round trip and
// keeps this server from holding sessions for machines that are long gone.
const sessionTTL = time.Hour

// session is one initialized client.
type session struct {
	scope    Scope
	lastSeen time.Time
}

// sessionStore keeps initialized sessions.
//
// WHY in memory and not in Postgres: a session carries no decision and no data
// worth surviving a restart — losing one costs the client a re-initialize,
// which the protocol already requires it to handle (a 404 on an unknown session
// is the defined signal). Putting it in the database would add a write to every
// request for nothing.
//
// The mutex is held only around map access; no tool call ever runs under it, so
// a slow expert cannot block another client's request.
type sessionStore struct {
	mu       sync.Mutex
	sessions map[string]*session
}

func newSessionStore() *sessionStore {
	return &sessionStore{sessions: make(map[string]*session)}
}

// create registers a session and returns its id.
func (st *sessionStore) create(scope Scope) (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	id := hex.EncodeToString(buf)

	st.mu.Lock()
	defer st.mu.Unlock()
	// Expiry is swept here rather than by a background goroutine: sessions only
	// accumulate when clients connect, so the work belongs on that path and no
	// extra goroutine has to be owned and shut down.
	now := time.Now()
	for key, sess := range st.sessions {
		if now.Sub(sess.lastSeen) > sessionTTL {
			delete(st.sessions, key)
		}
	}
	st.sessions[id] = &session{scope: scope, lastSeen: now}
	return id, nil
}

// touch validates a session id and refreshes it.
func (st *sessionStore) touch(id string) (Scope, bool) {
	if id == "" {
		return Scope{}, false
	}
	st.mu.Lock()
	defer st.mu.Unlock()
	sess, ok := st.sessions[id]
	if !ok {
		return Scope{}, false
	}
	if time.Since(sess.lastSeen) > sessionTTL {
		delete(st.sessions, id)
		return Scope{}, false
	}
	sess.lastSeen = time.Now()
	return sess.scope, true
}

// drop ends a session (the client said it is done).
func (st *sessionStore) drop(id string) {
	st.mu.Lock()
	defer st.mu.Unlock()
	delete(st.sessions, id)
}

// ServeHTTP implements MCP's Streamable HTTP transport.
//
// WHY this transport exists: a stdio session only works when the client can
// launch a process, which rules out anything hosted. HTTP lets a team point
// their coding agent at one URL with a per-user token, and it is where the token
// scope and the audit trail become meaningful — a stdio session belongs to
// whoever started the process and has no identity to attribute calls to.
//
// Shape (what a Claude Code style client expects):
//   - POST  — one JSON-RPC message; the reply is SSE when the client accepts an
//     event stream, plain JSON otherwise.
//   - GET   — a server-initiated stream. This server never pushes anything on its
//     own, so it answers 405, which the spec allows and clients handle.
//   - DELETE — end the session.
//
// Every request is authorised BEFORE it reaches the registry: an unknown or
// revoked token never gets as far as listing tools.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !originAllowed(r) {
		// DNS-rebinding guard: a page in a browser must not be able to drive a
		// local MCP server just because it can reach localhost.
		writeHTTPJSON(w, http.StatusForbidden, errorEnvelope(-32600, "origin not allowed"))
		return
	}

	switch r.Method {
	case http.MethodGet:
		// No server-initiated messages exist here, so there is nothing to stream.
		w.Header().Set("Allow", "POST, DELETE")
		writeHTTPJSON(w, http.StatusMethodNotAllowed, errorEnvelope(-32600, "this server does not offer a server-initiated stream"))
		return
	case http.MethodDelete:
		s.sessions.drop(strings.TrimSpace(r.Header.Get(sessionHeader)))
		w.WriteHeader(http.StatusNoContent)
		return
	case http.MethodPost:
		// handled below
	default:
		w.Header().Set("Allow", "POST, DELETE")
		writeHTTPJSON(w, http.StatusMethodNotAllowed, errorEnvelope(-32600, "method not allowed"))
		return
	}

	if s.tokens == nil {
		// Refusing is the honest answer: a server with no token store cannot
		// authenticate anyone, and running unauthenticated would expose every
		// expert to whoever finds the URL.
		writeHTTPJSON(w, http.StatusServiceUnavailable, errorEnvelope(-32603, "this server has no token store configured"))
		return
	}

	scope, info, err := s.tokens.Resolve(r.Context(), bearerToken(r))
	if err != nil {
		writeHTTPJSON(w, http.StatusUnauthorized, errorEnvelope(-32001, err.Error()))
		return
	}
	scope.TokenID = info.ID

	body, err := io.ReadAll(io.LimitReader(r.Body, maxRequestBytes+1))
	if err != nil {
		writeHTTPJSON(w, http.StatusBadRequest, errorEnvelope(-32700, "could not read request"))
		return
	}
	if len(body) > maxRequestBytes {
		writeHTTPJSON(w, http.StatusRequestEntityTooLarge, errorEnvelope(-32600, "request too large"))
		return
	}

	var req jsonRPCRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeHTTPJSON(w, http.StatusBadRequest, errorEnvelope(-32700, "parse error: "+err.Error()))
		return
	}

	// Session handling. initialize mints one; everything else must present it.
	// An unknown id is 404 — the protocol's signal for "re-initialize", which is
	// exactly what a client that reconnected to a restarted server should do.
	sessionID := strings.TrimSpace(r.Header.Get(sessionHeader))
	if req.Method == "initialize" {
		newID, sErr := s.sessions.create(scope)
		if sErr != nil {
			writeHTTPJSON(w, http.StatusInternalServerError, errorEnvelope(-32603, "could not start a session"))
			return
		}
		sessionID = newID
		w.Header().Set(sessionHeader, sessionID)
	} else if sessionID != "" {
		if storedScope, ok := s.sessions.touch(sessionID); ok {
			// The session's scope wins: the token that initialized the session is
			// the identity for its whole life.
			storedScope.TokenID = scope.TokenID
			scope = storedScope
		} else {
			writeHTTPJSON(w, http.StatusNotFound, errorEnvelope(-32001, "unknown or expired session; re-initialize"))
			return
		}
	}
	// A client that sends no session id at all is still served: some clients
	// (and curl) drive a single call without initializing, and refusing them
	// would buy no safety — the bearer token already authorised this request.

	// The request context is the client's: if the agent disconnects, the expert
	// answer is cancelled with it instead of running on for nobody.
	resp, replyDue := s.dispatch(r.Context(), scope, req)

	// Usage is recorded after the call so a bookkeeping failure cannot fail it.
	if touchErr := s.tokens.Touch(r.Context(), info.ID); touchErr != nil {
		s.logger.Warn("mcp token touch failed", zap.Error(touchErr))
	}

	if !replyDue {
		// Notifications and responses get no body, only an acknowledgement.
		w.WriteHeader(http.StatusAccepted)
		return
	}

	if acceptsEventStream(r) {
		s.writeSSE(w, resp)
		return
	}
	writeHTTPJSON(w, http.StatusOK, resp)
}

// writeSSE sends one JSON-RPC reply as a single SSE event and ends the stream.
//
// WHY a single event: this server answers a request and has nothing further to
// say, so holding the stream open would only make the client wait for an end it
// will never be told about.
func (s *Server) writeSSE(w http.ResponseWriter, payload jsonRPCResponse) {
	data, err := json.Marshal(payload)
	if err != nil {
		s.logger.Error("mcp: marshal reply failed", zap.Error(err))
		writeHTTPJSON(w, http.StatusInternalServerError, errorEnvelope(-32603, "could not encode the reply"))
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write([]byte("event: message\ndata: " + string(data) + "\n\n")); err != nil {
		s.logger.Warn("mcp: write sse reply failed", zap.Error(err))
		return
	}
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}

// acceptsEventStream reports whether the client wants SSE. Clients that accept
// both formats get SSE, because that is the shape the Streamable HTTP clients
// are built around; a plain JSON client (curl, a script) still gets JSON.
func acceptsEventStream(r *http.Request) bool {
	return strings.Contains(strings.ToLower(r.Header.Get("Accept")), "text/event-stream")
}

// originAllowed guards against a browser page driving this server. Requests
// without an Origin (CLI clients, agents) are normal and pass; a browser origin
// must be local or explicitly allowed.
func originAllowed(r *http.Request) bool {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		return true
	}
	lower := strings.ToLower(origin)
	return strings.Contains(lower, "://localhost") || strings.Contains(lower, "://127.0.0.1")
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

// errorEnvelope builds a transport-level JSON-RPC error body.
func errorEnvelope(code int, message string) map[string]any {
	return map[string]any{
		"jsonrpc": "2.0",
		"error":   jsonRPCError{Code: code, Message: message},
	}
}

// writeHTTPJSON writes one reply with the status the transport decided.
func writeHTTPJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
