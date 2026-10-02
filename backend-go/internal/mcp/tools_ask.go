package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// AskExpertTool asks one expert one question — the agent's "teacher on call".
type AskExpertTool struct {
	catalog  Catalog
	answerer ExpertAnswerer
}

// NewAskExpertTool builds the tool over its catalog and answerer.
func NewAskExpertTool(catalog Catalog, answerer ExpertAnswerer) *AskExpertTool {
	return &AskExpertTool{catalog: catalog, answerer: answerer}
}

// Name implements Tool.
func (t *AskExpertTool) Name() string { return "ask_expert" }

// Description implements Tool.
func (t *AskExpertTool) Description() string {
	return "Ask a domain expert a question and get its China-Wall-enforced answer, with citations to its training."
}

// Schema implements Tool.
func (t *AskExpertTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"domain":{"type":"string","description":"Domain, e.g. \"system design\" or \"frontend\" — either domain or expert_id is required"},"question":{"type":"string","description":"The question to ask"},"expert":{"type":"string","description":"Optional expert name or slug; defaults to the domain's expert"},"expert_id":{"type":"string","description":"Stable expert UUID from list_experts — preferred, copy exactly"}},"additionalProperties":false}`)
}

// Invoke implements Tool.
func (t *AskExpertTool) Invoke(ctx context.Context, scope Scope, args json.RawMessage) (ToolResult, error) {
	var in struct {
		Domain   string `json:"domain"`
		Question string `json:"question"`
		Expert   string `json:"expert"`
		ExpertID string `json:"expert_id"`
	}
	if err := decodeArgs(args, &in); err != nil {
		return ToolResult{}, err
	}
	in.Domain = strings.TrimSpace(in.Domain)
	in.Question = strings.TrimSpace(in.Question)
	in.ExpertID = strings.TrimSpace(in.ExpertID)
	if in.Question == "" {
		return ToolResult{}, NewToolError("INVALID_ARGS", "question is required")
	}
	if t.answerer == nil {
		return ToolResult{}, NewToolError("NOT_CONFIGURED", "ask_expert needs the API server to be configured")
	}
	// Stable ID roundtrip: list_experts -> expert_id -> answer. Try it first so a copied UUID never needs domain.
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
			answer, err := t.answerer.Answer(ctx, AnswerRequest{
				Domain:   expert.Domain,
				ExpertID: expert.ID,
				Question: in.Question,
			})
			if err != nil {
				return ToolResult{}, err
			}
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("# %s (domain: %s)\n\n", expert.Name, expert.Domain))
			sb.WriteString(answer.Answer)
			sb.WriteString("\n")
			if answer.Citations > 0 {
				sb.WriteString(fmt.Sprintf("\n_Citations: %d. This answer is grounded in the expert's training._\n", answer.Citations))
			}
			return ToolResult{Text: sb.String()}, nil
		}
		// unknown expert_id -> fall through to give helpful hint with real IDs
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

	answer, err := t.answerer.Answer(ctx, AnswerRequest{
		Domain:   expert.Domain,
		ExpertID: expert.ID,
		Question: in.Question,
	})
	if err != nil {
		return ToolResult{}, err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s (domain: %s)\n\n", expert.Name, expert.Domain))
	sb.WriteString(answer.Answer)
	sb.WriteString("\n")
	if answer.Citations > 0 {
		sb.WriteString(fmt.Sprintf("\n_Citations: %d. This answer is grounded in the expert's training._\n", answer.Citations))
	}
	return ToolResult{Text: sb.String()}, nil
}

// ReviewChangeTool has an expert review a change the agent just wrote.
//
// WHY it is built on the same answer path and not on the workflow's
// cross-verifier: the cross-verifier needs a workflow, tasks, reviewers and a
// blackboard to write to. A coding agent asks for a verdict on a diff, right
// now, in a few seconds — so the expert is asked to review, and the review comes
// back through the same China-Wall pipeline as any other answer. One answer
// implementation, no workflow state touched.
type ReviewChangeTool struct {
	catalog  Catalog
	answerer ExpertAnswerer
}

// NewReviewChangeTool builds the tool over its catalog and answerer.
func NewReviewChangeTool(catalog Catalog, answerer ExpertAnswerer) *ReviewChangeTool {
	return &ReviewChangeTool{catalog: catalog, answerer: answerer}
}

// Name implements Tool.
func (t *ReviewChangeTool) Name() string { return "review_change" }

// Description implements Tool.
func (t *ReviewChangeTool) Description() string {
	return "Ask a domain expert to review a code change and return a verdict with concrete issues."
}

// Schema implements Tool.
func (t *ReviewChangeTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"domain":{"type":"string","description":"Domain, e.g. \"system design\" or \"frontend\" — either domain or expert_id is required"},"expert_id":{"type":"string","description":"Stable expert UUID from list_experts — preferred"},"change":{"type":"string","description":"The diff or the code to review"},"context":{"type":"string","description":"Optional: what the change is meant to do"},"expert":{"type":"string","description":"Optional expert name or slug; defaults to the domain's expert"}},"additionalProperties":false}`)
}

// Invoke implements Tool.
func (t *ReviewChangeTool) Invoke(ctx context.Context, scope Scope, args json.RawMessage) (ToolResult, error) {
	var in struct {
		Domain   string `json:"domain"`
		ExpertID string `json:"expert_id"`
		Change   string `json:"change"`
		Context  string `json:"context"`
		Expert   string `json:"expert"`
	}
	if err := decodeArgs(args, &in); err != nil {
		return ToolResult{}, err
	}
	in.Domain = strings.TrimSpace(in.Domain)
	in.ExpertID = strings.TrimSpace(in.ExpertID)
	if strings.TrimSpace(in.Change) == "" {
		return ToolResult{}, NewToolError("INVALID_ARGS", "change is required")
	}
	if t.answerer == nil {
		return ToolResult{}, NewToolError("NOT_CONFIGURED", "review_change needs the API server to be configured")
	}
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
			sb.WriteString("You are reviewing a code change against your own standards.\n\n")
			sb.WriteString("Start your reply with exactly one line:\n")
			sb.WriteString("VERDICT: APPROVED\n")
			sb.WriteString("or\n")
			sb.WriteString("VERDICT: CHANGES_REQUESTED\n\n")
			sb.WriteString("Then list each issue as a separate line beginning with \"- \", naming the file or\n")
			sb.WriteString("symbol, what is wrong, and the principle from your training it violates (cite it\n")
			sb.WriteString("the way you normally do). If nothing violates your standards, say so plainly\n")
			sb.WriteString("instead of inventing issues.\n")
			if in.Context != "" {
				sb.WriteString("\nWhat the change is meant to do:\n")
				sb.WriteString(strings.TrimSpace(in.Context))
				sb.WriteString("\n")
			}
			sb.WriteString("\nThe change:\n\n")
			sb.WriteString(in.Change)
			answer, err := t.answerer.Answer(ctx, AnswerRequest{
				Domain:   expert.Domain,
				ExpertID: expert.ID,
				Question: sb.String(),
			})
			if err != nil {
				return ToolResult{}, err
			}
			var out strings.Builder
			out.WriteString(fmt.Sprintf("# Review by %s (domain: %s)\n\n", expert.Name, expert.Domain))
			out.WriteString(answer.Answer)
			out.WriteString("\n")
			return ToolResult{Text: out.String()}, nil
		}
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

	// The prompt is the contract: a verdict line the agent can act on without
	// parsing prose, then the specific issues. "Cite the principle you are
	// enforcing" is what keeps the review grounded in the expert's training
	// instead of generic style advice.
	var sb strings.Builder
	sb.WriteString("You are reviewing a code change against your own standards.\n\n")
	sb.WriteString("Start your reply with exactly one line:\n")
	sb.WriteString("VERDICT: APPROVED\n")
	sb.WriteString("or\n")
	sb.WriteString("VERDICT: CHANGES_REQUESTED\n\n")
	sb.WriteString("Then list each issue as a separate line beginning with \"- \", naming the file or\n")
	sb.WriteString("symbol, what is wrong, and the principle from your training it violates (cite it\n")
	sb.WriteString("the way you normally do). If nothing violates your standards, say so plainly\n")
	sb.WriteString("instead of inventing issues.\n")
	if in.Context != "" {
		sb.WriteString("\nWhat the change is meant to do:\n")
		sb.WriteString(strings.TrimSpace(in.Context))
		sb.WriteString("\n")
	}
	sb.WriteString("\nThe change:\n\n")
	sb.WriteString(in.Change)

	answer, err := t.answerer.Answer(ctx, AnswerRequest{
		Domain:   expert.Domain,
		ExpertID: expert.ID,
		Question: sb.String(),
	})
	if err != nil {
		return ToolResult{}, err
	}

	var out strings.Builder
	out.WriteString(fmt.Sprintf("# Review by %s (domain: %s)\n\n", expert.Name, expert.Domain))
	out.WriteString(answer.Answer)
	out.WriteString("\n")
	return ToolResult{Text: out.String()}, nil
}
