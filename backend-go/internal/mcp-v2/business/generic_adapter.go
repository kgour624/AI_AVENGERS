// Package business — Generic relief wrapper, NO edit in internal/decision/engine.go (R2).
package business

import (
	"context"
	"fmt"
)

// GenericConfig mirrors mcp_expert_generic_config row.
type GenericConfig struct {
	AllowGeneric   bool
	GenericPercent int // 0|5|10|20 CHECK
	Reason         string
}

// DecisionChecker is Business-local port — Business never imports App (Ultimate Go §1,§5).
type DecisionChecker interface {
	Check(ctx context.Context, expertID, question string, chunks []Chunk) (GateResult, error)
}
type GatewayCaller interface {
	Call(ctx context.Context, tier string, messages []Message) (string, GatewayUsage, error)
}

// McpDecisionAdapter wraps DecisionChecker for generic relief (5-10%).
type McpDecisionAdapter struct {
	decision DecisionChecker
	gateway  GatewayCaller
}

func NewMcpDecisionAdapter(d DecisionChecker, g GatewayCaller) *McpDecisionAdapter {
	return &McpDecisionAdapter{decision: d, gateway: g}
}

// CheckWithGeneric calls gate2, if GateStopped==2 and allow_generic then second LLM call with tagged budget.
// Returns answer with [DOMAIN:90%][GENERIC:10%] tags + genericUsed flag.
func (a *McpDecisionAdapter) CheckWithGeneric(ctx context.Context, expertID, question string, chunks []Chunk, cfg GenericConfig) (answer string, genericUsed bool, percent int, err error) {
	if a.decision != nil {
		gate, err := a.decision.Check(ctx, expertID, question, chunks)
		if err == nil && gate.GateStopped == 2 && !gate.Allowed {
			if !cfg.AllowGeneric || cfg.GenericPercent == 0 {
				return "", false, 0, fmt.Errorf("knowledge not covered and generic not allowed")
			}
			// Generic budget call — tag parts
			prompt := fmt.Sprintf("%s\n\nYou may use up to %d%% general knowledge outside cited chunks. Tag generic parts [GENERIC:%d%%] and domain parts [DOMAIN:%d%%].", question, cfg.GenericPercent, cfg.GenericPercent, 100-cfg.GenericPercent)
			msgs := []Message{{Role: "user", Content: prompt}}
			for _, c := range chunks {
				msgs = append(msgs, Message{Role: "user", Content: "CITATION: " + c.Text})
			}
			out, _, err := a.gateway.Call(ctx, "strong", msgs)
			if err != nil {
				return "", false, 0, fmt.Errorf("generic gateway: %w", err)
			}
			return out, true, cfg.GenericPercent, nil
		}
		if err != nil {
			return "", false, 0, err
		}
	}
	// No gate block — normal domain answer (caller will do gateway call)
	return "", false, 0, nil
}