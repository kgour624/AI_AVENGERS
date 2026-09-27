package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"go.uber.org/zap"
)

// protocolVersion is the MCP revision this server speaks.
const protocolVersion = "2024-11-05"

// Server answers MCP requests for one client session.
type Server struct {
	registry *Registry
	logger   *zap.Logger
	version  string
	scope    Scope
	// auditor may be nil: the stdio session on a developer's machine has
	// nothing to audit against, and a nil auditor must never be special-cased
	// at every call site.
	auditor *Auditor
	// tokens is set only for the HTTP transport. A nil store means "this server
	// cannot authenticate anybody", which HTTP refuses instead of running open.
	tokens TokenStore
}

// NewServer builds a server over a registry. The scope is fixed for the session:
// a stdio session belongs to whoever launched the binary, so there is no token
// to resolve per request.
func NewServer(registry *Registry, logger *zap.Logger, version string, scope Scope) *Server {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Server{registry: registry, logger: logger, version: version, scope: scope}
}

// SetTokenStore enables token authentication for the HTTP transport.
func (s *Server) SetTokenStore(tokens TokenStore) { s.tokens = tokens }

// SetAuditor enables the audit trail for this server.
func (s *Server) SetAuditor(auditor *Auditor) { s.auditor = auditor }

// jsonRPCRequest is one incoming message. Notifications (no id) are legal in the
// protocol and must not be answered.
type jsonRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// jsonRPCError is the failure shape MCP expects.
type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type jsonRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  any             `json:"result,omitempty"`
	Error   *jsonRPCError   `json:"error,omitempty"`
}

// ServeStdio reads newline-delimited JSON-RPC from in and writes replies to out.
//
// WHY one request at a time: an MCP stdio client sends a request and waits for
// its reply, so serving them concurrently would gain nothing and would make the
// reply order non-deterministic. The HTTP transport, when it lands, owns its own
// bounded worker pool — same rule as everywhere else in this codebase: the
// parent owns every goroutine it starts.
func (s *Server) ServeStdio(ctx context.Context, in io.Reader, out io.Writer) error {
	reader := bufio.NewReaderSize(in, 1<<20)
	writer := bufio.NewWriter(out)
	defer writer.Flush()

	for {
		// Cancelling ctx is how a shutdown arrives (the client exiting closes
		// stdin, a signal cancels the parent context).
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			s.handleLine(ctx, writer, line)
			if flushErr := writer.Flush(); flushErr != nil {
				return fmt.Errorf("mcp: write reply: %w", flushErr)
			}
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return fmt.Errorf("mcp: read request: %w", err)
		}
	}
}

// handleLine decodes and answers one message. A malformed line is reported as a
// protocol error and the session continues — one bad request must not kill the
// session the client is using.
func (s *Server) handleLine(ctx context.Context, w io.Writer, line []byte) {
	var req jsonRPCRequest
	if err := json.Unmarshal(line, &req); err != nil {
		s.writeJSON(w, jsonRPCResponse{
			JSONRPC: "2.0",
			Error:   &jsonRPCError{Code: -32700, Message: "parse error: " + err.Error()},
		})
		return
	}

	resp, replyDue := s.dispatch(ctx, s.scope, req)
	if !replyDue {
		return
	}
	s.writeJSON(w, resp)
}

// writeJSON writes one reply. A write error means the client is gone; log it and
// move on — there is nothing left to respond to.
func (s *Server) writeJSON(w io.Writer, payload jsonRPCResponse) {
	data, err := json.Marshal(payload)
	if err != nil {
		s.logger.Error("mcp: marshal reply failed", zap.Error(err))
		return
	}
	data = append(data, '\n')
	if _, err := w.Write(data); err != nil {
		s.logger.Warn("mcp: write reply failed", zap.Error(err))
	}
}
