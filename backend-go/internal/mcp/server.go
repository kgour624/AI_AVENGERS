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

	notification := len(req.ID) == 0

	switch req.Method {
	case "initialize":
		if notification {
			return
		}
		s.writeJSON(w, jsonRPCResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{
			"protocolVersion": protocolVersion,
			"capabilities":    map[string]any{"tools": map[string]any{}},
			"serverInfo":      map[string]any{"name": "ai-avengers-experts", "version": s.version},
		}})
	case "notifications/initialized", "initialized":
		// Client acknowledgement. Nothing to do, nothing to reply.
	case "ping":
		if notification {
			return
		}
		s.writeJSON(w, jsonRPCResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{}})
	case "tools/list":
		if notification {
			return
		}
		s.writeJSON(w, jsonRPCResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{
			"tools": s.registry.List(s.scope),
		}})
	case "tools/call":
		if notification {
			return
		}
		s.handleToolCall(ctx, w, req)
	default:
		if notification {
			return
		}
		s.writeJSON(w, jsonRPCResponse{JSONRPC: "2.0", ID: req.ID,
			Error: &jsonRPCError{Code: -32601, Message: "method not found: " + req.Method}})
	}
}

// handleToolCall runs one tool. A ToolError is a caller mistake (unknown domain,
// forbidden tool) and is reported as a tool-level error; anything else is a
// server fault, logged with its detail while the client only learns that the
// call failed. Handled once, here, at the boundary.
func (s *Server) handleToolCall(ctx context.Context, w io.Writer, req jsonRPCRequest) {
	var params struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		s.writeJSON(w, jsonRPCResponse{JSONRPC: "2.0", ID: req.ID,
			Error: &jsonRPCError{Code: -32602, Message: "invalid params: " + err.Error()}})
		return
	}

	result, err := s.registry.Call(ctx, s.scope, params.Name, params.Arguments)
	if err != nil {
		var toolErr *ToolError
		if errors.As(err, &toolErr) {
			s.logger.Info("mcp tool rejected",
				zap.String("tool", params.Name),
				zap.String("code", toolErr.Code),
			)
			// isError keeps this a successful JSON-RPC response: MCP models a
			// tool-level failure this way, and the client needs to see why.
			s.writeToolError(w, req.ID, toolErr.Message)
			return
		}
		s.logger.Error("mcp tool failed", zap.String("tool", params.Name), zap.Error(err))
		s.writeToolError(w, req.ID, "The tool failed on the server.")
		return
	}

	s.writeJSON(w, jsonRPCResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{
		"content": []map[string]any{{"type": "text", "text": result.Text}},
		"isError": false,
	}})
}

func (s *Server) writeToolError(w io.Writer, id json.RawMessage, message string) {
	s.writeJSON(w, jsonRPCResponse{JSONRPC: "2.0", ID: id, Result: map[string]any{
		"content": []map[string]any{{"type": "text", "text": message}},
		"isError": true,
	}})
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
