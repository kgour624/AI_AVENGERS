package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"

	"ai_avengers/backend/internal/capability"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const WorkflowPlanEvent = "workflow_plan_configured"

// WorkflowPlanStep is one client-authored step. Artifact names and step kinds
// are user data, not a built-in list: new domains and outputs need no code edit.
type WorkflowPlanStep struct {
	ID           string    `json:"id"`
	ExpertID     uuid.UUID `json:"expert_id"`
	Output       string    `json:"output"`
	Instructions string    `json:"instructions"`
	Kind         string    `json:"kind,omitempty"` // capability kind: design, implementation, testing, data, product, other
	Language     string    `json:"language,omitempty"`
	Capability   string    `json:"capability,omitempty"`
	DependsOn    []string  `json:"depends_on,omitempty"`
}

// ValidateWorkflowPlan validates the user-authored DAG and returns its
// deterministic topological order. It rejects missing experts, duplicate IDs,
// self/cyclic dependencies, and empty outputs before anything is launched.
func ValidateWorkflowPlan(steps []WorkflowPlanStep, selected []uuid.UUID) ([]WorkflowPlanStep, error) {
	if len(steps) == 0 {
		return nil, nil
	}
	selectedSet := make(map[uuid.UUID]bool, len(selected))
	for _, id := range selected {
		selectedSet[id] = true
	}
	byID := make(map[string]int, len(steps))
	for i := range steps {
		s := &steps[i]
		s.ID = strings.TrimSpace(s.ID)
		s.Output = strings.TrimSpace(s.Output)
		s.Instructions = strings.TrimSpace(s.Instructions)
		s.Kind = strings.ToLower(strings.TrimSpace(s.Kind))
		if s.ID == "" || s.Output == "" || s.Instructions == "" {
			return nil, fmt.Errorf("plan step %d needs id, output and instructions", i+1)
		}
		if _, exists := byID[s.ID]; exists {
			return nil, fmt.Errorf("duplicate plan step id %q", s.ID)
		}
		if !selectedSet[s.ExpertID] {
			return nil, fmt.Errorf("step %q uses an expert not selected for this workflow", s.ID)
		}
		byID[s.ID] = i
	}
	indegree := make([]int, len(steps))
	children := make([][]int, len(steps))
	for i, s := range steps {
		for _, dep := range s.DependsOn {
			j, ok := byID[dep]
			if !ok {
				return nil, fmt.Errorf("step %q depends on unknown step %q", s.ID, dep)
			}
			if i == j {
				return nil, fmt.Errorf("step %q cannot depend on itself", s.ID)
			}
			indegree[i]++
			children[j] = append(children[j], i)
		}
	}
	ordered := make([]WorkflowPlanStep, 0, len(steps))
	used := make([]bool, len(steps))
	for len(ordered) < len(steps) {
		ready := make([]int, 0, len(steps))
		for i := range steps {
			if !used[i] && indegree[i] == 0 {
				ready = append(ready, i)
			}
		}
		if len(ready) == 0 {
			return nil, fmt.Errorf("workflow plan contains a dependency cycle")
		}
		// Preserve the client's displayed order for independent steps. The saved
		// dependencies still determine which steps may start together.
		sort.Ints(ready)
		for _, found := range ready {
			used[found] = true
			ordered = append(ordered, steps[found])
			for _, child := range children[found] {
				indegree[child]--
			}
		}
	}
	return ordered, nil
}

func customPlanTasks(steps []WorkflowPlanStep) []TaskSpec {
	if len(steps) == 0 {
		return nil
	}
	out := make([]TaskSpec, 0, len(steps))
	idByStep := make(map[string]uuid.UUID, len(steps))
	for _, step := range steps {
		idByStep[step.ID] = step.ExpertID
	}
	for _, step := range steps {
		deps := make([]uuid.UUID, 0, len(step.DependsOn))
		for _, dep := range step.DependsOn {
			deps = append(deps, idByStep[dep])
		}
		out = append(out, TaskSpec{
			ExpertID:               step.ExpertID,
			Title:                  step.Output,
			Description:            step.Instructions,
			DependsOnExpertIDs:     deps,
			CustomPlanKind:         step.Kind,
			CustomPlanOutput:       step.Output,
			CustomPlanStepID:       step.ID,
			CustomPlanInstructions: step.Instructions,
			CustomPlanLanguage:     step.Language,
			CustomPlanCapability:   step.Capability,
			DependsOnStepIDs:       append([]string(nil), step.DependsOn...),
		})
	}
	return out
}

func (r *WorkflowRunner) loadCustomPlan(ctx context.Context, workflowID uuid.UUID, selected []workflowExpert) ([]TaskSpec, bool, error) {
	events, err := r.store.GetByType(ctx, workflowID, []string{WorkflowPlanEvent}, 0)
	if err != nil {
		return nil, false, fmt.Errorf("load configured workflow plan: %w", err)
	}
	if len(events) == 0 {
		return nil, false, nil
	}
	steps, err := decodeWorkflowPlan(events[len(events)-1].Content)
	if err != nil || len(steps) == 0 {
		r.logger.Error("workflow: saved custom plan is invalid", zap.String("workflow_id", workflowID.String()), zap.Error(err))
		if err == nil {
			err = fmt.Errorf("configured plan is empty")
		}
		return nil, true, err
	}
	selectedIDs := make([]uuid.UUID, 0, len(selected))
	for _, e := range selected {
		selectedIDs = append(selectedIDs, e.ID)
	}
	validated, err := ValidateWorkflowPlan(steps, selectedIDs)
	if err != nil {
		r.logger.Error("workflow: saved custom plan no longer validates", zap.String("workflow_id", workflowID.String()), zap.Error(err))
		return nil, true, err
	}
	return customPlanTasks(validated), true, nil
}

func decodeWorkflowPlan(raw json.RawMessage) ([]WorkflowPlanStep, error) {
	var steps []WorkflowPlanStep
	if err := json.Unmarshal(raw, &steps); err != nil {
		return nil, err
	}
	return steps, nil
}

func customPlanPhase(kind string) (string, capability.Kind) {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "implementation":
		return PhaseImplementation, capability.KindImplementation
	case "testing":
		return PhaseQA, capability.KindTesting
	case "design":
		return PhaseHighLevelDesign, capability.KindDesign
	default:
		// Any user-defined kind uses the generic design executor and is checked
		// against the selected expert's declared kind below when it is one of the
		// built-in kinds. New domain/output names require no code change.
		return PhaseDetailedDesign, capability.Kind(strings.ToLower(strings.TrimSpace(kind)))
	}
}

// runCustomPlanStep executes one client-authored step with the executor its
// declared capability names. Kept in this file so the workflow's custom-plan
// behaviour lives together, and so the runner's phase machinery cannot
// reinterpret a client step as a legacy phase task.
func (r *WorkflowRunner) runCustomPlanStep(
	ctx context.Context,
	workflowID uuid.UUID,
	expert workflowExpert,
	t TaskSpec,
	allExperts []workflowExpert,
	state *runnerState,
	errMu *sync.Mutex,
	recordSuccess func(*uuid.UUID),
	recordFailure func(string),
	firstErr *error,
) error {
	// Capability guard: a declared mismatch (for example a design expert handed a
	// code step) is refused with a reason instead of silently producing the wrong
	// kind of artifact. An UNCLASSIFIED domain profile means the client has not
	// declared a capability yet, so the explicit step choice is honoured rather
	// than blocking work on missing metadata.
	requiredKind := capability.Kind(t.CustomPlanCapability)
	if requiredKind == capability.KindUnclassified {
		requiredKind = capability.Kind(t.CustomPlanKind)
	}
	if requiredKind != capability.KindUnclassified {
		declarations := r.loadCapabilityDeclarations(ctx)
		decl := declarations[capability.NormName(expert.Domain)]
		if decl.Kind != capability.KindUnclassified && decl.Kind != requiredKind {
			return fmt.Errorf("step %q needs a %s expert, but %q is declared as %s; fix the expert's domain profile or change this step",
				t.CustomPlanOutput, requiredKind, expert.Name, decl.Kind)
		}
		if (requiredKind == capability.KindImplementation || requiredKind == capability.KindTesting) &&
			decl.Language != "" && t.CustomPlanLanguage != "" &&
			capability.LanguageAliases(decl.Language) != capability.LanguageAliases(t.CustomPlanLanguage) {
			return fmt.Errorf("step %q needs language %q, but %q is declared for %q",
				t.CustomPlanOutput, t.CustomPlanLanguage, expert.Name, decl.Language)
		}
	}

	switch t.CustomPlanCapability {
	case string(capability.KindImplementation):
		_, _, err := r.runImplementationTask(ctx, AiderRunRequest{
			WorkflowID: workflowID, Expert: expert, TaskID: uuid.Nil,
			TaskTitle: planTaskTitle(t), TaskDescription: planTaskDescription(t),
			WorkflowPhase: PhaseImplementation,
		}, true)
		if err != nil {
			return err
		}
	case string(capability.KindTesting):
		_, err := r.qaRunner.Run(ctx, AiderRunRequest{
			WorkflowID: workflowID, Expert: expert, TaskID: uuid.Nil,
			TaskTitle: planTaskTitle(t), TaskDescription: planTaskDescription(t),
			WorkflowPhase: PhaseQA,
		})
		if err != nil {
			return err
		}
	default:
		res, err := r.agentLoop.Run(ctx, AgentLoopRequest{
			WorkflowID: workflowID, Expert: expert, TaskID: uuid.Nil,
			TaskTitle: planTaskTitle(t), TaskDescription: planTaskDescription(t),
			WorkflowPhase: PhaseDetailedDesign, AllExperts: allExperts,
		})
		if err != nil {
			return err
		}
		recordSuccess(res.ArtifactEventID)
		errMu.Lock()
		state.CompletedExpertIDs = append(state.CompletedExpertIDs, expert.ID.String())
		errMu.Unlock()
		return nil
	}
	recordSuccess(nil)
	errMu.Lock()
	state.CompletedExpertIDs = append(state.CompletedExpertIDs, expert.ID.String())
	errMu.Unlock()
	return nil
}
