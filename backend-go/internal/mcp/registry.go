package mcp

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
)

// Registry holds the tools this server exposes.
//
// WHY the map is built once and never written again: it is read by every
// request, so making it immutable after construction removes the need for a
// lock entirely — no shared mutable state, no races to reason about.
type Registry struct {
	byName map[string]Tool
	order  []string
}

// NewRegistry builds a registry from the tools main decides to expose.
func NewRegistry(tools ...Tool) *Registry {
	r := &Registry{byName: make(map[string]Tool, len(tools))}
	for _, t := range tools {
		if t == nil {
			continue
		}
		name := t.Name()
		if _, dup := r.byName[name]; dup {
			// A duplicate name would silently shadow a tool; refuse the second
			// one and keep the first rather than guess which was intended.
			continue
		}
		r.byName[name] = t
		r.order = append(r.order, name)
	}
	sort.Strings(r.order)
	return r
}

// ToolDescriptor is one entry of the tools/list reply.
type ToolDescriptor struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

// List returns the tools a scope may call, in a stable order.
func (r *Registry) List(scope Scope) []ToolDescriptor {
	out := make([]ToolDescriptor, 0, len(r.order))
	for _, name := range r.order {
		t := r.byName[name]
		if !scope.AllowsTool(name) {
			continue
		}
		out = append(out, ToolDescriptor{
			Name:        t.Name(),
			Description: t.Description(),
			InputSchema: t.Schema(),
		})
	}
	return out
}

// Call invokes one tool. Authorisation is checked here, before the tool runs,
// so no tool has to remember to check its own scope.
func (r *Registry) Call(ctx context.Context, scope Scope, name string, args json.RawMessage) (ToolResult, error) {
	t, ok := r.byName[strings.TrimSpace(name)]
	if !ok {
		return ToolResult{}, NewToolError("UNKNOWN_TOOL", "no such tool: "+name)
	}
	if !scope.AllowsTool(name) {
		return ToolResult{}, NewToolError("FORBIDDEN", "this token may not call "+name)
	}
	return t.Invoke(ctx, scope, args)
}
