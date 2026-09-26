package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/blackboard"
	"ai_avengers/backend/internal/gateway"
)

// UnderstandingEvent is the blackboard event type for "this expert says it
// understood the requirement like this".
//
// WHY this exists (the root of every downstream error): once a design starts on
// a misunderstood requirement, a perfectly trained expert still produces a
// perfectly wrong artifact, and every later phase inherits that fault. Before
// this, the client approved a plan without ever seeing what the expert THOUGHT
// it was asked to build. This artifact is that missing checkpoint: plain-language
// restatement + explicit boundaries + what is still unclear, shown to the client
// before any design work begins.
const UnderstandingEvent = "understanding_captured"

// maxUnderstandingAttempts bounds the restate loop: client adds missing info,
// experts restate, client reviews again. Without a cap a client that keeps
// asking for changes would loop forever.
const maxUnderstandingAttempts = 3

// UnderstandingArtifact is one expert's restatement of the requirement.
//
// Every field is prose the CLIENT can judge, not internal vocabulary: the point
// is to make a misunderstanding visible to a non-engineer before it is built on.
type UnderstandingArtifact struct {
	ExpertID   uuid.UUID `json:"expert_id"`
	ExpertName string    `json:"expert_name"`
	// Restatement: the requirement in the expert's own words, plainly. "Man ki
	// baat" — if another expert (or the client) cannot follow this, the expert
	// has not understood yet.
	Restatement string   `json:"restatement"`
	Goal        string   `json:"goal"`
	InScope     []string `json:"in_scope"`
	// OutOfScope: what the expert believes is explicitly NOT part of the task.
	// This is where unwanted additions usually hide.
	OutOfScope      []string `json:"out_of_scope"`
	Assumptions     []string `json:"assumptions"`
	Unknowns        []string `json:"unknowns"`
	SuccessCriteria []string `json:"success_criteria"`
}

// understandingSystemPrompt forces a structured, client-judgeable answer and
// forbids designing: this step is about comprehension, not solutions.
const understandingSystemPrompt = `You are a domain expert asked to prove you UNDERSTOOD a requirement before any design work starts.
Do NOT design, propose solutions, or write code. Only restate and bound the problem.
Answer with ONLY one JSON object:
{
  "restatement": "the requirement in your own plain words, 2-4 sentences, understandable by a non-engineer",
  "goal": "the single outcome the client wants",
  "in_scope": ["..."],
  "out_of_scope": ["what is explicitly NOT part of this task"],
  "assumptions": ["things you are taking for granted — flag them so they can be corrected"],
  "unknowns": ["what you genuinely do not know and need the client to answer"],
  "success_criteria": ["how we will know it is done correctly"]
}
Be specific and honest: an empty "unknowns" list is suspicious unless the requirement is truly unambiguous.`

// generateUnderstanding asks one expert to restate the requirement.
func generateUnderstanding(
	ctx context.Context,
	gw *gateway.ModelGateway,
	expert workflowExpert,
	requirement string,
) (*UnderstandingArtifact, error) {
	if gw == nil {
		return nil, fmt.Errorf("understanding: no model gateway")
	}
	if strings.TrimSpace(requirement) == "" {
		return nil, fmt.Errorf("understanding: requirement is empty")
	}

	resp, err := gw.Call(ctx, gateway.LLMRequest{
		Model:        gateway.ModelStrong,
		SystemPrompt: understandingSystemPrompt,
		UserPrompt:   fmt.Sprintf("EXPERT DOMAIN: %s\nEXPERT CHARTER:\n%s\n\nREQUIREMENT:\n%s", expert.Domain, expert.ReasoningCharter, requirement),
		MaxTokens:    1500,
		Temperature:  0.2,
	})
	if err != nil {
		return nil, fmt.Errorf("understanding: llm: %w", err)
	}

	raw := stripJSONFencePublic(resp.Content)
	var art UnderstandingArtifact
	if err := json.Unmarshal([]byte(raw), &art); err != nil {
		return nil, fmt.Errorf("understanding: parse: %w", err)
	}
	if strings.TrimSpace(art.Restatement) == "" {
		return nil, fmt.Errorf("understanding: empty restatement")
	}
	art.ExpertID = expert.ID
	art.ExpertName = expert.Name
	return &art, nil
}

// stripJSONFencePublic removes a ```json fence if the model added one despite the
// instruction. Kept local and small: callers here must not fail on a cosmetic
// wrapper.
func stripJSONFencePublic(raw string) string {
	s := strings.TrimSpace(raw)
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```json")
		s = strings.TrimPrefix(s, "```")
		s = strings.TrimSuffix(s, "```")
		s = strings.TrimSpace(s)
	}
	if start := strings.IndexByte(s, '{'); start > 0 {
		s = s[start:]
	}
	if end := strings.LastIndexByte(s, '}'); end >= 0 && end < len(s)-1 {
		s = s[:end+1]
	}
	return s
}

// captureUnderstandings generates and stores one artifact per expert.
//
// WHY every expert and not a lead: any domain can be on the team (design, LLD,
// product, data, ...), so "who leads" is not knowable up front and picking one
// would hide exactly the misunderstanding this step exists to expose. Each
// expert states its own reading; the client sees them all side by side.
func (r *WorkflowRunner) captureUnderstandings(
	ctx context.Context,
	workflowID uuid.UUID,
	experts []workflowExpert,
	requirement string,
) ([]UnderstandingArtifact, error) {
	out := make([]UnderstandingArtifact, 0, len(experts))
	for _, e := range experts {
		art, err := generateUnderstanding(ctx, r.gateway, e, requirement)
		if err != nil {
			// One expert failing must not hide the others: the client needs to see
			// who understood and who did not.
			r.logger.Warn("understanding: expert could not restate the requirement",
				zap.String("expert", e.Name), zap.Error(err))
			continue
		}
		content, _ := json.Marshal(art)
		expertID := e.ID
		if _, postErr := r.store.Post(ctx, blackboard.PostRequest{
			WorkflowID:       workflowID,
			EventType:        UnderstandingEvent,
			PostedByExpertID: &expertID,
			Content:          json.RawMessage(content),
		}); postErr != nil {
			r.logger.Warn("understanding: could not persist artifact",
				zap.String("expert", e.Name), zap.Error(postErr))
		}
		out = append(out, *art)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("understanding: no expert produced a restatement")
	}
	return out, nil
}

// askUnderstandingGate shows every expert's restatement to the client and waits
// for the decision. Returns the decision, the client's notes (what was missing),
// and any error. Mirrors the design gate so both behave identically.
func (r *WorkflowRunner) askUnderstandingGate(
	ctx context.Context,
	workflowID uuid.UUID,
	arts []UnderstandingArtifact,
	attempt int,
) (string, string, error) {
	var from uuid.UUID
	if len(arts) > 0 {
		from = arts[0].ExpertID
	}
	approvalID, err := r.tools.AskClient(ctx, AskClientRequest{
		WorkflowID:      workflowID,
		FromExpertID:    from,
		GateName:        "understanding",
		Summary:         fmt.Sprintf("Before any design: does your team understand the requirement correctly? (review %d of %d)", attempt, maxUnderstandingAttempts),
		ArtifactContent: arts,
	})
	if err != nil {
		return "", "", fmt.Errorf("understanding gate: %w", err)
	}
	if err := r.waitForResume(ctx, workflowID); err != nil {
		return "", "", fmt.Errorf("understanding gate wait: %w", err)
	}
	decision, notes := r.lastApprovalDecision(ctx, workflowID, approvalID, "understanding")
	r.logger.Info("runner: understanding gate answered",
		zap.String("workflow_id", workflowID.String()),
		zap.Int("attempt", attempt),
		zap.String("decision", decision))
	return decision, notes, nil
}
