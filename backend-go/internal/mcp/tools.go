package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// ListExpertsTool answers "which experts do I have?" — the first thing a client
// needs, because every other call names a domain or an expert.
type ListExpertsTool struct {
	catalog Catalog
}

// NewListExpertsTool builds the tool over its catalog.
func NewListExpertsTool(catalog Catalog) *ListExpertsTool {
	return &ListExpertsTool{catalog: catalog}
}

// Name implements Tool.
func (t *ListExpertsTool) Name() string { return "list_experts" }

// Description implements Tool.
func (t *ListExpertsTool) Description() string {
	return "List the trained domain experts available, grouped by domain."
}

// Schema implements Tool.
func (t *ListExpertsTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{},"additionalProperties":false}`)
}

// Invoke implements Tool.
func (t *ListExpertsTool) Invoke(ctx context.Context, scope Scope, _ json.RawMessage) (ToolResult, error) {
	experts, err := t.catalog.ListExperts(ctx)
	if err != nil {
		return ToolResult{}, err
	}

	var sb strings.Builder
	sb.WriteString("# Available experts\n\n")
	sb.WriteString("Use the `id` value as `expert_id` in the next call; copy it exactly — do not guess.\n\n")
	shown := 0
	for _, e := range experts {
		if !scope.AllowsDomain(e.Domain) {
			continue
		}
		if !scope.AllowsExpert(e.ID) {
			continue
		}
		shown++
		sb.WriteString(fmt.Sprintf("- **%s** — id: `%s` — domain: `%s` (slug: `%s`)\n", e.Name, e.ID, e.Domain, e.Slug))
	}
	if shown == 0 {
		return ToolResult{Text: "No experts are available to this token."}, nil
	}
	return ToolResult{Text: sb.String()}, nil
}

// GetStandardsTool returns one expert's standards as a "teacher packet": the
// text a coding agent should follow instead of its generic defaults.
type GetStandardsTool struct {
	catalog Catalog
}

// NewGetStandardsTool builds the tool over its catalog.
func NewGetStandardsTool(catalog Catalog) *GetStandardsTool {
	return &GetStandardsTool{catalog: catalog}
}

// Name implements Tool.
func (t *GetStandardsTool) Name() string { return "get_standards" }

// Description implements Tool.
func (t *GetStandardsTool) Description() string {
	return "Return a domain expert's standards (charter, principles) to follow while writing code."
}

// Schema implements Tool.
func (t *GetStandardsTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"domain":{"type":"string","description":"Domain, e.g. \"system design\" or \"frontend\""},"expert":{"type":"string","description":"Optional expert name or slug; defaults to the domain's expert"},"expert_id":{"type":"string","description":"Stable expert UUID from list_experts — preferred, copy exactly"}},"additionalProperties":false}`)
}

// Invoke implements Tool.
func (t *GetStandardsTool) Invoke(ctx context.Context, scope Scope, args json.RawMessage) (ToolResult, error) {
	var in struct {
		Domain   string `json:"domain"`
		Expert   string `json:"expert"`
		ExpertID string `json:"expert_id"`
	}
	if err := decodeArgs(args, &in); err != nil {
		return ToolResult{}, err
	}
	in.Domain = strings.TrimSpace(in.Domain)
	in.ExpertID = strings.TrimSpace(in.ExpertID)
	// expert_id is the stable roundtrip ID from list_experts — try it first.
	if in.ExpertID != "" {
		expert, found, err := t.catalog.ExpertByID(ctx, in.ExpertID)
		if err != nil {
			return ToolResult{}, err
		}
		if found {
			if !scope.AllowsExpert(expert.ID) {
				return ToolResult{}, NewToolError("FORBIDDEN", "this token may not use expert "+expert.Name)
			}
			if !scope.AllowsDomain(expert.Domain) {
				return ToolResult{}, NewToolError("FORBIDDEN", "this token may not use domain "+expert.Domain)
			}
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("# Standards: %s (domain: %s)\n\n", expert.Name, expert.Domain))
			sb.WriteString("Treat the following as the authoritative rules for this task. Where they and your\n")
			sb.WriteString("generic defaults disagree, these win. If something you need is not covered here,\n")
			sb.WriteString("say so explicitly instead of inventing a rule.\n\n")
			sb.WriteString("## Reasoning charter\n\n")
			if strings.TrimSpace(expert.ReasoningCharter) == "" {
				sb.WriteString("_No charter recorded for this expert yet._\n")
			} else {
				sb.WriteString(expert.ReasoningCharter)
				sb.WriteString("\n")
			}
			if strings.TrimSpace(expert.Description) != "" {
				sb.WriteString("\n## Scope\n\n")
				sb.WriteString(expert.Description)
				sb.WriteString("\n")
			}
			return ToolResult{Text: sb.String()}, nil
		}
		// fall through to domain lookup so we can give a helpful error with the real IDs
	}
	if in.Domain == "" {
		return ToolResult{}, NewToolError("INVALID_ARGS", "domain or expert_id is required — call list_experts to get the correct id")
	}
	if !scope.AllowsDomain(in.Domain) {
		return ToolResult{}, NewToolError("FORBIDDEN", "this token may not use domain "+in.Domain)
	}

	expert, found, err := t.catalog.ExpertFor(ctx, in.Domain, in.Expert)
	if err != nil {
		return ToolResult{}, err
	}
	if !found {
		return ToolResult{}, NewToolError("UNKNOWN_EXPERT", "no expert found for domain "+in.Domain+". "+availableExpertsHint(ctx, t.catalog, scope))
	}
	if !scope.AllowsExpert(expert.ID) {
		return ToolResult{}, NewToolError("FORBIDDEN", "this token may not use expert "+expert.Name)
	}

	// The packet is framed as instructions to the coding agent on purpose: a
	// bare charter dump reads as background, while "these rules win over your
	// defaults" is what makes the agent apply them to the code it writes.
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Standards: %s (domain: %s)\n\n", expert.Name, expert.Domain))
	sb.WriteString("Treat the following as the authoritative rules for this task. Where they and your\n")
	sb.WriteString("generic defaults disagree, these win. If something you need is not covered here,\n")
	sb.WriteString("say so explicitly instead of inventing a rule.\n\n")
	sb.WriteString("## Reasoning charter\n\n")
	if strings.TrimSpace(expert.ReasoningCharter) == "" {
		sb.WriteString("_No charter recorded for this expert yet._\n")
	} else {
		sb.WriteString(expert.ReasoningCharter)
		sb.WriteString("\n")
	}
	if strings.TrimSpace(expert.Description) != "" {
		sb.WriteString("\n## Scope\n\n")
		sb.WriteString(expert.Description)
		sb.WriteString("\n")
	}
	return ToolResult{Text: sb.String()}, nil
}
