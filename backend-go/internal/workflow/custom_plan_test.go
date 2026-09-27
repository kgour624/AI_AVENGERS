package workflow

import (
	"testing"

	"github.com/google/uuid"
)

func TestValidateWorkflowPlanOrdersDependenciesAndPreservesIndependentSteps(t *testing.T) {
	a, b, c := uuid.New(), uuid.New(), uuid.New()
	steps := []WorkflowPlanStep{
		{ID: "code", ExpertID: b, Output: "Go service", Instructions: "Implement the API", Kind: "implementation", DependsOn: []string{"design"}},
		{ID: "design", ExpertID: a, Output: "Design doc", Instructions: "Specify the contract", Kind: "design"},
		{ID: "qa", ExpertID: c, Output: "Test plan", Instructions: "Test failure modes", Kind: "testing", DependsOn: []string{"code"}},
	}
	ordered, err := ValidateWorkflowPlan(steps, []uuid.UUID{a, b, c})
	if err != nil {
		t.Fatal(err)
	}
	if ordered[0].ID != "design" || ordered[1].ID != "code" || ordered[2].ID != "qa" {
		t.Fatalf("topological order = %v, want design -> code -> qa", []string{ordered[0].ID, ordered[1].ID, ordered[2].ID})
	}
	tasks := customPlanTasks(ordered)
	if len(tasks) != 3 || tasks[1].CustomPlanOutput != "Go service" || tasks[1].CustomPlanStepID != "code" {
		t.Fatalf("custom tasks did not preserve expert/output mapping: %+v", tasks)
	}
}

func TestValidateWorkflowPlanRejectsMissingExpertUnknownDependencyAndCycle(t *testing.T) {
	a, b := uuid.New(), uuid.New()
	base := func(id string, expert uuid.UUID) WorkflowPlanStep {
		return WorkflowPlanStep{ID: id, ExpertID: expert, Output: id + " output", Instructions: "do " + id}
	}

	if _, err := ValidateWorkflowPlan([]WorkflowPlanStep{base("a", a)}, []uuid.UUID{b}); err == nil {
		t.Fatal("must reject expert not selected in workflow")
	}
	x := base("x", a)
	x.DependsOn = []string{"missing"}
	if _, err := ValidateWorkflowPlan([]WorkflowPlanStep{x}, []uuid.UUID{a}); err == nil {
		t.Fatal("must reject unknown dependency")
	}
	x, y := base("x", a), base("y", b)
	x.DependsOn = []string{"y"}
	y.DependsOn = []string{"x"}
	if _, err := ValidateWorkflowPlan([]WorkflowPlanStep{x, y}, []uuid.UUID{a, b}); err == nil {
		t.Fatal("must reject a dependency cycle")
	}
}
