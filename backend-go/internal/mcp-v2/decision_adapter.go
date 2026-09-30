package mcpv2

import (
	"context"

	"github.com/google/uuid"
	"ai_avengers/backend/internal/chinawall"
	"ai_avengers/backend/internal/decision"
)

// DecisionAdapter wraps real decision.Engine — implements DecisionPort (Check) via Engine.Process.
// No edit in internal/decision/engine.go (R2). Business/generic_adapter also wraps this.
type DecisionAdapter struct {
	engine *decision.Engine
}

func NewDecisionAdapter(engine *decision.Engine) *DecisionAdapter {
	return &DecisionAdapter{engine: engine}
}

// Check maps mcp-v2 Chunks -> chinawall.CourseChunk and calls Engine.Process (gate2 coverage).
// Returns GateStopped + Allowed for AskExpert generic relief flow.
func (d *DecisionAdapter) Check(ctx context.Context, expertID, question string, chunks []Chunk) (GateResult, error) {
	if d.engine == nil {
		return GateResult{GateStopped: 0, Allowed: true}, nil
	}
	var cwChunks []chinawall.CourseChunk
	for _, c := range chunks {
		id, _ := uuid.Parse(c.ID)
		cwChunks = append(cwChunks, chinawall.CourseChunk{ID: id, Text: c.Text})
	}
	// Build minimal Expert — domain lookup not needed for gate2 coverage check; gate2 checks chunks non-empty.
	expertUUID, _ := uuid.Parse(expertID)
	expert := decision.Expert{
		ID:   expertUUID,
		Name: expertID,
	}
	res, err := d.engine.Process(ctx, question, expert, cwChunks, "", "", 0, nil, nil)
	if err != nil {
		return GateResult{}, err
	}
	allowed := res.Mode != decision.ModeREFUSE
	gateStopped := res.GateStopped
	// Gate2 refuse maps to GateStopped==2 in old port; Process returns GateStopped 2 when knowledge not covered.
	if res.Mode == decision.ModeREFUSE && res.Reason != "" {
		// Heuristic: if chunks empty, treat as gate2
		if len(cwChunks) == 0 {
			gateStopped = 2
			allowed = false
		}
	}
	return GateResult{GateStopped: gateStopped, Allowed: allowed}, nil
}