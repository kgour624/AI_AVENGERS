// Package mcp exposes this system's domain experts to an external agent
// (Claude Code, Cursor, any MCP-capable client) over the Model Context
// Protocol.
//
// WHY a separate package and a separate binary: the MCP server must never be
// able to change how chat or workflows run. It imports the business data it
// needs (experts) and nothing that executes a workflow or a chat turn, so the
// existing paths are untouched — a bug here can only break MCP, never the
// product. Imports still only go down (API → app → business → foundation).
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// Tool is one capability this server offers.
//
// WHY an interface: the registry must not know which concrete tools exist.
// Adding a tool is adding a struct and one line in main — nothing in the
// dispatch path changes (open for extension, closed for modification).
// One method, like every interface here: a tool is a single job, and a
// single-method interface is the smallest thing a caller can depend on.
type Tool interface {
	Name() string
	Description() string
	// Schema is the MCP input schema handed to the client so it knows how to
	// call this tool.
	Schema() json.RawMessage
	Invoke(ctx context.Context, scope Scope, args json.RawMessage) (ToolResult, error)
}

// ToolResult is what the client receives. Text only: these tools answer in
// prose/markdown, and a text-only result keeps every tool uniform.
type ToolResult struct {
	Text string
}

// ToolError is a caller-facing failure. Anything else that escapes a tool is a
// server fault and must not leak its internals to the client (handle once, at
// the boundary).
type ToolError struct {
	Code    string
	Message string
}

func (e *ToolError) Error() string { return e.Message }

// NewToolError builds a caller-facing error.
func NewToolError(code, message string) *ToolError {
	return &ToolError{Code: code, Message: message}
}

// Scope is what one caller (one token) is allowed to do.
//
// WHY a value and not a pointer: it is copied into the tool call and never
// mutated, so two concurrent calls can never observe each other's scope.
type Scope struct {
	// Label identifies the token in logs and audits (never the token itself).
	Label string
	// TokenID is the token row this scope came from. Carried so an audit entry
	// can point at the token without re-resolving it. Empty for stdio sessions,
	// which have no token.
	TokenID string
	// Domains restricts which experts may be reached. Empty = every domain,
	// which is what the local stdio case (the owner's own machine) wants.
	Domains []string
	// Tools restricts which tools may be called. Empty = every tool.
	Tools []string
}

// AllowsTool reports whether this scope may call the named tool.
func (s Scope) AllowsTool(name string) bool {
	if len(s.Tools) == 0 {
		return true
	}
	for _, t := range s.Tools {
		if strings.EqualFold(strings.TrimSpace(t), name) {
			return true
		}
	}
	return false
}

// AllowsDomain reports whether this scope may reach the given domain. Domain
// comparison reuses the product's own normalisation rule so "System Design" and
// "system_design" are one domain here exactly as they are inside the app.
func (s Scope) AllowsDomain(domain string) bool {
	if len(s.Domains) == 0 {
		return true
	}
	want := NormDomainKey(domain)
	for _, d := range s.Domains {
		if NormDomainKey(d) == want {
			return true
		}
	}
	return false
}

// NormDomainKey lowercases and strips separators so a domain compares equal
// regardless of how a human or a token wrote it.
func NormDomainKey(domain string) string {
	d := strings.ToLower(strings.TrimSpace(domain))
	d = strings.ReplaceAll(d, "_", " ")
	d = strings.ReplaceAll(d, "-", " ")
	return strings.Join(strings.Fields(d), " ")
}

// decodeArgs unmarshals tool arguments with a caller-facing error, so a bad
// payload is reported as a bad payload rather than as a server crash.
func decodeArgs(args json.RawMessage, dst any) error {
	if len(args) == 0 {
		return nil
	}
	if err := json.Unmarshal(args, dst); err != nil {
		return NewToolError("INVALID_ARGS", fmt.Sprintf("could not read arguments: %v", err))
	}
	return nil
}
