package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"ai_avengers/backend/internal/capability"
)

// loadCapabilityDeclarations reads the capability/language each domain has
// declared in domain_profiles.config.
//
// WHY the profile row and not the expert: capability is a property of the DOMAIN
// ("go programming" writes Go), so declaring it once covers every expert of that
// domain, and a new domain is onboarded by adding a profile row — no Go change.
// Missing keys leave the domain unclassified, which never satisfies a code or QA
// requirement (an unclassified expert must not be mistaken for a programmer).
func (r *WorkflowRunner) loadCapabilityDeclarations(ctx context.Context) map[string]capability.Declaration {
	out := map[string]capability.Declaration{}
	rows, err := r.db.Query(ctx, `SELECT domain, config FROM domain_profiles`)
	if err != nil {
		r.logger.Warn("capability: could not read domain profiles", zap.Error(err))
		return out
	}
	defer rows.Close()

	for rows.Next() {
		var domain string
		var raw []byte
		if err := rows.Scan(&domain, &raw); err != nil {
			continue
		}
		var cfg struct {
			Capability string `json:"capability"`
			Language   string `json:"language"`
		}
		if len(raw) > 0 {
			if err := json.Unmarshal(raw, &cfg); err != nil {
				continue
			}
		}
		out[capability.NormName(domain)] = capability.Declaration{
			Kind:     capability.Kind(strings.ToLower(strings.TrimSpace(cfg.Capability))),
			Language: capability.LanguageAliases(cfg.Language),
		}
	}
	return out
}

// expertsForPhase narrows the workflow's experts to those allowed to run a
// phase, per the declared capabilities.
//
// WHY it returns an error instead of falling back to "all experts": the incident
// this fixes was exactly that fallback — a workflow with only system-design
// experts ran its implementation phase and produced Go code none of them owned.
// Blocking with the reason ("implementation phase needs a Go implementation
// expert") is the correct, actionable behaviour.
func (r *WorkflowRunner) expertsForPhase(
	ctx context.Context,
	phase string,
	experts []workflowExpert,
	language string,
) ([]workflowExpert, error) {
	decls := r.loadCapabilityDeclarations(ctx)

	in := make([]capability.Expert, 0, len(experts))
	for _, e := range experts {
		in = append(in, capability.Expert{
			ID:          e.ID.String(),
			Name:        e.Name,
			Declaration: decls[capability.NormName(e.Domain)],
		})
	}

	matched, err := capability.SelectForPhase(phase, in, language)
	if err != nil {
		return nil, err
	}
	allowed := make(map[string]bool, len(matched))
	for _, m := range matched {
		allowed[m.ID] = true
	}
	out := make([]workflowExpert, 0, len(matched))
	for _, e := range experts {
		if allowed[e.ID.String()] {
			out = append(out, e)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no expert may run the %s phase", phase)
	}
	return out, nil
}
