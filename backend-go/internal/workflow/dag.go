package workflow

import (
	"fmt"

	"github.com/google/uuid"
)

// ExecutionWave is one group of tasks that can run in parallel.
// All tasks in a wave have no unmet dependencies within the wave.
// Waves execute sequentially; tasks within a wave execute in parallel.
//
// Example:
//   Wave 0: [PM]           ← no deps, runs first
//   Wave 1: [SD, DSA]      ← both depend only on PM
//   Wave 2: [LLD]          ← depends on SD + DSA
type ExecutionWave []TaskSpec

// BuildDAG validates the task dependency graph and returns execution waves.
//
// Algorithm: Kahn's topological sort (BFS-based).
//   WHY Kahn's not DFS: Kahn's naturally groups tasks into parallel waves.
//   DFS gives a linear order; Kahn's gives the optimal parallel grouping.
//
// Mental execution:
//   tasks = [PM(deps:[]), SD(deps:[PM]), DSA(deps:[PM]), LLD(deps:[SD,DSA])]
//
//   in-degree: PM=0, SD=1, DSA=1, LLD=2
//   queue: [PM]
//
//   Wave 0: process PM → SD.in-degree=0, DSA.in-degree=0
//   Wave 1: process SD, DSA → LLD.in-degree=0
//   Wave 2: process LLD
//
//   Result: [[PM], [SD, DSA], [LLD]]
//
// Cycle detection:
//   If total processed < len(tasks) → cycle exists → return error.
//   Example cycle: A→B→A → both have in-degree > 0 forever → never enter queue.
func BuildDAG(tasks []TaskSpec) ([]ExecutionWave, error) {
	if len(tasks) == 0 {
		return nil, fmt.Errorf("dag: no tasks provided")
	}

	// Build expert_id → TaskSpec index map for O(1) lookup.
	taskIndex := make(map[string]int, len(tasks))
	for i, t := range tasks {
		taskIndex[t.ExpertID.String()] = i
	}

	// Validate: all depends_on_expert_ids must exist in task list.
	// WHY here not in Planner: Planner validates against expert list.
	// DAG validates against task list (a subset of experts may have tasks).
	for _, t := range tasks {
		for _, depID := range t.DependsOnExpertIDs {
			if _, ok := taskIndex[depID.String()]; !ok {
				return nil, fmt.Errorf(
					"dag: task for expert %s depends on expert %s which has no task",
					t.ExpertID, depID,
				)
			}
		}
	}

	// Compute in-degree for each task.
	// in-degree[i] = number of tasks that must complete before task i.
	//
	// BUG FIX: this used to do `idx := taskIndex[depID.String()]; inDegree[idx]++`,
	// which incremented the in-degree of the DEPENDENCY (depID) instead of the
	// DEPENDENT (t). That's backwards: a task with zero dependencies would get
	// its in-degree bumped by every task that depends on it, so it would never
	// reach in-degree 0 and would be misreported as part of a cycle — exactly
	// the "cycle detected among 1 tasks: [expertID]" failure seen in production
	// for a task that had no dependencies at all. in-degree[i] is simply the
	// number of dependencies task i itself declares.
	inDegree := make([]int, len(tasks))
	for i, t := range tasks {
		inDegree[i] = len(t.DependsOnExpertIDs)
	}

	// Build adjacency list: task i → tasks that depend on i.
	// When task i completes, we decrement in-degree of its dependents.
	adj := make([][]int, len(tasks))
	for i, t := range tasks {
		for _, depID := range t.DependsOnExpertIDs {
			depIdx := taskIndex[depID.String()]
			// depIdx → i: when depIdx completes, i's in-degree decreases.
			adj[depIdx] = append(adj[depIdx], i)
		}
	}

	// Kahn's BFS: process tasks wave by wave.
	var waves []ExecutionWave
	processed := 0

	// Initial wave: all tasks with in-degree 0.
	var currentWave []int
	for i, deg := range inDegree {
		if deg == 0 {
			currentWave = append(currentWave, i)
		}
	}

	for len(currentWave) > 0 {
		// Build this wave's TaskSpecs.
		wave := make(ExecutionWave, len(currentWave))
		for i, idx := range currentWave {
			wave[i] = tasks[idx]
		}
		waves = append(waves, wave)
		processed += len(currentWave)

		// Find next wave: tasks whose in-degree drops to 0.
		var nextWave []int
		for _, idx := range currentWave {
			for _, dependent := range adj[idx] {
				inDegree[dependent]--
				if inDegree[dependent] == 0 {
					nextWave = append(nextWave, dependent)
				}
			}
		}
		currentWave = nextWave
	}

	// Cycle detection: if not all tasks were processed, a cycle exists.
	if processed < len(tasks) {
		// Find the cycle members for a useful error message.
		var cycleMembers []uuid.UUID
		for i, deg := range inDegree {
			if deg > 0 {
				cycleMembers = append(cycleMembers, tasks[i].ExpertID)
			}
		}
		return nil, fmt.Errorf(
			"dag: cycle detected among %d tasks: %v",
			len(cycleMembers), cycleMembers,
		)
	}

	return waves, nil
}
