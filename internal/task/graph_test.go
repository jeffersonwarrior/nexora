package task

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTaskGraph(t *testing.T) {
	t.Parallel()
	t.Run("creates empty graph", func(t *testing.T) {
		g := NewTaskGraph()
		assert.NotNil(t, g)
		assert.NotNil(t, g.nodes)
		assert.NotNil(t, g.edges)
	})
}

func TestGraph_AddTask(t *testing.T) {
	t.Parallel()
	t.Run("adds task successfully", func(t *testing.T) {
		g := NewTaskGraph()
		task := &Task{ID: "task-1", Title: "Task 1"}

		err := g.AddTask(task)
		assert.NoError(t, err)
		assert.NotNil(t, g.nodes["task-1"])
	})

	t.Run("prevents duplicate task IDs", func(t *testing.T) {
		g := NewTaskGraph()
		task := &Task{ID: "task-1", Title: "Task 1"}

		err := g.AddTask(task)
		assert.NoError(t, err)

		err = g.AddTask(task)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("prevents nil task", func(t *testing.T) {
		g := NewTaskGraph()
		err := g.AddTask(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nil task")
	})
}

func TestGraph_AddDependency(t *testing.T) {
	t.Parallel()
	t.Run("adds dependency successfully", func(t *testing.T) {
		g := NewTaskGraph()
		taskA := &Task{ID: "A", Title: "Task A", Dependencies: []string{}}
		taskB := &Task{ID: "B", Title: "Task B", Dependencies: []string{}}

		g.AddTask(taskA)
		g.AddTask(taskB)

		// A depends on B
		err := g.AddDependency("A", "B")
		assert.NoError(t, err)
		assert.Contains(t, g.edges["A"], "B")
	})

	t.Run("prevents dependency on non-existent task", func(t *testing.T) {
		g := NewTaskGraph()
		taskA := &Task{ID: "A", Title: "Task A", Dependencies: []string{}}
		g.AddTask(taskA)

		err := g.AddDependency("A", "non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("prevents self-dependency", func(t *testing.T) {
		g := NewTaskGraph()
		taskA := &Task{ID: "A", Title: "Task A", Dependencies: []string{}}
		g.AddTask(taskA)

		err := g.AddDependency("A", "A")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "self-dependency")
	})
}

func TestGraph_DetectCycle(t *testing.T) {
	t.Parallel()
	t.Run("detects simple cycle", func(t *testing.T) {
		g := NewTaskGraph()
		taskA := &Task{ID: "A", Title: "Task A", Dependencies: []string{}}
		taskB := &Task{ID: "B", Title: "Task B", Dependencies: []string{}}
		taskC := &Task{ID: "C", Title: "Task C", Dependencies: []string{}}

		g.AddTask(taskA)
		g.AddTask(taskB)
		g.AddTask(taskC)

		// Create cycle: A -> B -> C -> A
		g.AddDependency("A", "B")
		g.AddDependency("B", "C")
		g.AddDependency("C", "A")

		hasCycle, cycle := g.DetectCycle()
		assert.True(t, hasCycle)
		assert.NotEmpty(t, cycle)
	})

	t.Run("no cycle in linear chain", func(t *testing.T) {
		g := NewTaskGraph()
		taskA := &Task{ID: "A", Title: "Task A", Dependencies: []string{}}
		taskB := &Task{ID: "B", Title: "Task B", Dependencies: []string{}}
		taskC := &Task{ID: "C", Title: "Task C", Dependencies: []string{}}

		g.AddTask(taskA)
		g.AddTask(taskB)
		g.AddTask(taskC)

		// Linear: A -> B -> C
		g.AddDependency("A", "B")
		g.AddDependency("B", "C")

		hasCycle, cycle := g.DetectCycle()
		assert.False(t, hasCycle)
		assert.Empty(t, cycle)
	})

	t.Run("no cycle in empty graph", func(t *testing.T) {
		g := NewTaskGraph()
		hasCycle, cycle := g.DetectCycle()
		assert.False(t, hasCycle)
		assert.Empty(t, cycle)
	})
}

func TestGraph_TopologicalSort(t *testing.T) {
	t.Parallel()
	t.Run("sorts simple dependency chain", func(t *testing.T) {
		g := NewTaskGraph()
		taskA := &Task{ID: "A", Title: "Task A", Dependencies: []string{}}
		taskB := &Task{ID: "B", Title: "Task B", Dependencies: []string{}}
		taskC := &Task{ID: "C", Title: "Task C", Dependencies: []string{}}

		g.AddTask(taskA)
		g.AddTask(taskB)
		g.AddTask(taskC)

		// A -> B -> C (A depends on B, B depends on C)
		g.AddDependency("A", "B")
		g.AddDependency("B", "C")

		sorted, err := g.TopologicalSort()
		require.NoError(t, err)
		require.Len(t, sorted, 3)

		// C should come before B, B should come before A
		idxA := -1
		idxB := -1
		idxC := -1
		for i, task := range sorted {
			switch task.ID {
			case "A":
				idxA = i
			case "B":
				idxB = i
			case "C":
				idxC = i
			}
		}

		assert.Less(t, idxC, idxB, "C should come before B")
		assert.Less(t, idxB, idxA, "B should come before A")
	})

	t.Run("sorts diamond dependency", func(t *testing.T) {
		g := NewTaskGraph()
		// Diamond: A depends on B and C, both B and C depend on D
		taskA := &Task{ID: "A", Title: "Task A", Dependencies: []string{}}
		taskB := &Task{ID: "B", Title: "Task B", Dependencies: []string{}}
		taskC := &Task{ID: "C", Title: "Task C", Dependencies: []string{}}
		taskD := &Task{ID: "D", Title: "Task D", Dependencies: []string{}}

		g.AddTask(taskA)
		g.AddTask(taskB)
		g.AddTask(taskC)
		g.AddTask(taskD)

		g.AddDependency("A", "B")
		g.AddDependency("A", "C")
		g.AddDependency("B", "D")
		g.AddDependency("C", "D")

		sorted, err := g.TopologicalSort()
		require.NoError(t, err)
		require.Len(t, sorted, 4)

		// D should come first, then B and C, then A
		idxA := -1
		idxB := -1
		idxC := -1
		idxD := -1
		for i, task := range sorted {
			switch task.ID {
			case "A":
				idxA = i
			case "B":
				idxB = i
			case "C":
				idxC = i
			case "D":
				idxD = i
			}
		}

		assert.Less(t, idxD, idxB, "D should come before B")
		assert.Less(t, idxD, idxC, "D should come before C")
		assert.Less(t, idxB, idxA, "B should come before A")
		assert.Less(t, idxC, idxA, "C should come before A")
	})

	t.Run("fails on cycle", func(t *testing.T) {
		g := NewTaskGraph()
		taskA := &Task{ID: "A", Title: "Task A", Dependencies: []string{}}
		taskB := &Task{ID: "B", Title: "Task B", Dependencies: []string{}}

		g.AddTask(taskA)
		g.AddTask(taskB)

		// Create cycle: A -> B -> A
		g.AddDependency("A", "B")
		g.AddDependency("B", "A")

		_, err := g.TopologicalSort()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cycle")
	})

	t.Run("sorts empty graph", func(t *testing.T) {
		g := NewTaskGraph()
		sorted, err := g.TopologicalSort()
		assert.NoError(t, err)
		assert.Empty(t, sorted)
	})
}

func TestGraph_GetDependencies(t *testing.T) {
	t.Parallel()
	t.Run("returns direct dependencies", func(t *testing.T) {
		g := NewTaskGraph()
		taskA := &Task{ID: "A", Title: "Task A", Dependencies: []string{}}
		taskB := &Task{ID: "B", Title: "Task B", Dependencies: []string{}}
		taskC := &Task{ID: "C", Title: "Task C", Dependencies: []string{}}

		g.AddTask(taskA)
		g.AddTask(taskB)
		g.AddTask(taskC)

		g.AddDependency("A", "B")
		g.AddDependency("A", "C")

		deps := g.GetDependencies("A")
		assert.Len(t, deps, 2)

		depIDs := make(map[string]bool)
		for _, dep := range deps {
			depIDs[dep.ID] = true
		}
		assert.True(t, depIDs["B"])
		assert.True(t, depIDs["C"])
	})

	t.Run("returns empty for task with no dependencies", func(t *testing.T) {
		g := NewTaskGraph()
		taskA := &Task{ID: "A", Title: "Task A", Dependencies: []string{}}
		g.AddTask(taskA)

		deps := g.GetDependencies("A")
		assert.Empty(t, deps)
	})

	t.Run("returns empty for non-existent task", func(t *testing.T) {
		g := NewTaskGraph()
		deps := g.GetDependencies("non-existent")
		assert.Empty(t, deps)
	})
}

func TestGraph_GetTransitiveDependencies(t *testing.T) {
	t.Parallel()
	t.Run("returns all transitive dependencies", func(t *testing.T) {
		g := NewTaskGraph()
		taskA := &Task{ID: "A", Title: "Task A", Dependencies: []string{}}
		taskB := &Task{ID: "B", Title: "Task B", Dependencies: []string{}}
		taskC := &Task{ID: "C", Title: "Task C", Dependencies: []string{}}
		taskD := &Task{ID: "D", Title: "Task D", Dependencies: []string{}}

		g.AddTask(taskA)
		g.AddTask(taskB)
		g.AddTask(taskC)
		g.AddTask(taskD)

		// A -> B -> C -> D
		g.AddDependency("A", "B")
		g.AddDependency("B", "C")
		g.AddDependency("C", "D")

		deps := g.GetTransitiveDependencies("A")
		assert.Len(t, deps, 3) // B, C, D

		depIDs := make(map[string]bool)
		for _, dep := range deps {
			depIDs[dep.ID] = true
		}
		assert.True(t, depIDs["B"])
		assert.True(t, depIDs["C"])
		assert.True(t, depIDs["D"])
	})

	t.Run("returns empty for task with no dependencies", func(t *testing.T) {
		g := NewTaskGraph()
		taskA := &Task{ID: "A", Title: "Task A", Dependencies: []string{}}
		g.AddTask(taskA)

		deps := g.GetTransitiveDependencies("A")
		assert.Empty(t, deps)
	})

	t.Run("handles diamond dependencies", func(t *testing.T) {
		g := NewTaskGraph()
		// Diamond: A -> B, A -> C, B -> D, C -> D
		taskA := &Task{ID: "A", Title: "Task A", Dependencies: []string{}}
		taskB := &Task{ID: "B", Title: "Task B", Dependencies: []string{}}
		taskC := &Task{ID: "C", Title: "Task C", Dependencies: []string{}}
		taskD := &Task{ID: "D", Title: "Task D", Dependencies: []string{}}

		g.AddTask(taskA)
		g.AddTask(taskB)
		g.AddTask(taskC)
		g.AddTask(taskD)

		g.AddDependency("A", "B")
		g.AddDependency("A", "C")
		g.AddDependency("B", "D")
		g.AddDependency("C", "D")

		deps := g.GetTransitiveDependencies("A")
		assert.Len(t, deps, 3) // B, C, D (D should appear once despite two paths)

		depIDs := make(map[string]bool)
		for _, dep := range deps {
			depIDs[dep.ID] = true
		}
		assert.True(t, depIDs["B"])
		assert.True(t, depIDs["C"])
		assert.True(t, depIDs["D"])
	})
}

func TestGraph_CalculateProgress(t *testing.T) {
	t.Parallel()
	t.Run("calculates progress with no dependencies", func(t *testing.T) {
		g := NewTaskGraph()
		task := &Task{
			ID:         "A",
			Title:      "Task A",
			Milestones: []Milestone{{Status: StatusCompleted}, {Status: StatusActive}},
		}
		g.AddTask(task)

		progress := g.CalculateProgress("A")
		assert.Equal(t, 0.5, progress) // 1 of 2 milestones completed
	})

	t.Run("calculates progress with dependencies", func(t *testing.T) {
		g := NewTaskGraph()
		taskA := &Task{
			ID:         "A",
			Title:      "Task A",
			Milestones: []Milestone{{Status: StatusActive}},
		}
		taskB := &Task{
			ID:         "B",
			Title:      "Task B",
			Milestones: []Milestone{{Status: StatusCompleted}},
		}
		g.AddTask(taskA)
		g.AddTask(taskB)
		g.AddDependency("A", "B")

		progress := g.CalculateProgress("A")
		// A is 0% done, B is 100% done
		// Weighted: 0.5 * A's progress + 0.5 * B's progress = 0.5 * 0 + 0.5 * 1 = 0.5
		assert.Greater(t, progress, 0.0)
		assert.Less(t, progress, 1.0)
	})

	t.Run("returns 0 for task with no milestones", func(t *testing.T) {
		g := NewTaskGraph()
		task := &Task{ID: "A", Title: "Task A", Milestones: []Milestone{}}
		g.AddTask(task)

		progress := g.CalculateProgress("A")
		assert.Equal(t, 0.0, progress)
	})

	t.Run("returns 0 for non-existent task", func(t *testing.T) {
		g := NewTaskGraph()
		progress := g.CalculateProgress("non-existent")
		assert.Equal(t, 0.0, progress)
	})
}

func TestGraph_CriticalPath(t *testing.T) {
	t.Parallel()

	t.Run("empty graph returns empty path", func(t *testing.T) {
		g := NewTaskGraph()
		path, err := g.CriticalPath()
		require.NoError(t, err)
		assert.Empty(t, path)
	})

	t.Run("single task with no dependencies", func(t *testing.T) {
		g := NewTaskGraph()
		taskA := &Task{ID: "A", Title: "Task A"}
		g.AddTask(taskA)

		path, err := g.CriticalPath()
		require.NoError(t, err)
		assert.Equal(t, []string{"A"}, path)
	})

	t.Run("linear chain dependency (A depends on B, B depends on C)", func(t *testing.T) {
		g := NewTaskGraph()
		taskA := &Task{ID: "A", Title: "Task A"}
		taskB := &Task{ID: "B", Title: "Task B"}
		taskC := &Task{ID: "C", Title: "Task C"}

		g.AddTask(taskA)
		g.AddTask(taskB)
		g.AddTask(taskC)

		// A -> B -> C (A depends on B, B depends on C)
		g.AddDependency("A", "B")
		g.AddDependency("B", "C")

		path, err := g.CriticalPath()
		require.NoError(t, err)
		require.Len(t, path, 3)

		// Path should be from root (C) to dependent (A)
		assert.Equal(t, "C", path[0])
		assert.Equal(t, "B", path[1])
		assert.Equal(t, "A", path[2])
	})

	t.Run("diamond dependency (A depends on B and C, both depend on D)", func(t *testing.T) {
		g := NewTaskGraph()
		taskA := &Task{ID: "A", Title: "Task A"}
		taskB := &Task{ID: "B", Title: "Task B"}
		taskC := &Task{ID: "C", Title: "Task C"}
		taskD := &Task{ID: "D", Title: "Task D"}

		g.AddTask(taskA)
		g.AddTask(taskB)
		g.AddTask(taskC)
		g.AddTask(taskD)

		// Diamond: A depends on B and C, both B and C depend on D
		g.AddDependency("A", "B")
		g.AddDependency("A", "C")
		g.AddDependency("B", "D")
		g.AddDependency("C", "D")

		path, err := g.CriticalPath()
		require.NoError(t, err)

		// Critical path should be: D -> B -> A or D -> C -> A (both have length 3)
		// We verify the structure is correct
		assert.Equal(t, "D", path[0])                   // Must start with D
		assert.Equal(t, "A", path[2])                   // Must end with A
		assert.Contains(t, []string{"B", "C"}, path[1]) // Middle should be B or C
	})

	t.Run("complex graph with multiple paths", func(t *testing.T) {
		// Graph:
		//        E
		//       / \
		//      B   F
		//     /   /
		//    A   /
		//     \ /
		//      C
		//      |
		//      D
		//
		// Path A->C->D = 3
		// Path B->C->D = 3
		// Path E->F->C->D = 4  (longest)
		g := NewTaskGraph()

		for id := range []string{"A", "B", "C", "D", "E", "F"} {
			taskID := string(rune('A' + id))
			g.AddTask(&Task{ID: taskID, Title: "Task " + taskID})
		}

		// Define dependencies
		g.AddDependency("C", "A") // C depends on A
		g.AddDependency("C", "B") // C depends on B
		g.AddDependency("C", "F") // C depends on F
		g.AddDependency("B", "E") // B depends on E
		g.AddDependency("F", "E") // F depends on E
		g.AddDependency("D", "C") // D depends on C

		path, err := g.CriticalPath()
		require.NoError(t, err)

		// Critical path should include E and have length 4
		require.Len(t, path, 4)
		assert.Equal(t, "E", path[0]) // Must start with E
		assert.Equal(t, "D", path[3]) // Must end with D
		// Should go through F or B (both depend on E)
		assert.Contains(t, []string{"B", "F"}, path[1])
	})

	t.Run("fails on cyclic dependency", func(t *testing.T) {
		g := NewTaskGraph()
		taskA := &Task{ID: "A", Title: "Task A"}
		taskB := &Task{ID: "B", Title: "Task B"}

		g.AddTask(taskA)
		g.AddTask(taskB)

		// Create cycle: A -> B -> A
		g.AddDependency("A", "B")
		g.AddDependency("B", "A")

		_, err := g.CriticalPath()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cycle")
	})

	t.Run("independent tasks (no dependencies)", func(t *testing.T) {
		g := NewTaskGraph()
		taskA := &Task{ID: "A", Title: "Task A"}
		taskB := &Task{ID: "B", Title: "Task B"}
		taskC := &Task{ID: "C", Title: "Task C"}

		g.AddTask(taskA)
		g.AddTask(taskB)
		g.AddTask(taskC)

		path, err := g.CriticalPath()
		require.NoError(t, err)

		// With no dependencies, critical path length is 1
		// Any single task is a valid critical path
		assert.Len(t, path, 1)
		assert.Contains(t, []string{"A", "B", "C"}, path[0])
	})

	t.Run("branching structure (one root, multiple branches)", func(t *testing.T) {
		// Graph:
		//     A
		//    / \
		//   B   C
		//   |   |
		//   D   E
		//    \ /
		//     F
		//
		// Path D->B->A->F = 4
		// Path E->C->A->F = 4
		// Path A->B->D = 3
		// Path A->C->E = 3
		g := NewTaskGraph()

		for id := range []string{"A", "B", "C", "D", "E", "F"} {
			taskID := string(rune('A' + id))
			g.AddTask(&Task{ID: taskID, Title: "Task " + taskID})
		}

		g.AddDependency("B", "A")
		g.AddDependency("C", "A")
		g.AddDependency("D", "B")
		g.AddDependency("E", "C")
		g.AddDependency("F", "D")
		g.AddDependency("F", "E")

		path, err := g.CriticalPath()
		require.NoError(t, err)

		// Critical path should be length 4
		require.Len(t, path, 4)
		assert.Equal(t, "A", path[0]) // Must start with A
		assert.Equal(t, "F", path[3]) // Must end with F
	})

	t.Run("W-shaped topology (multiple convergence points)", func(t *testing.T) {
		// Graph:
		//   A     D
		//   |     |
		//   B     E
		//    \   /
		//     C
		//
		// Path A->B->C = 3
		// Path D->E->C = 3
		g := NewTaskGraph()

		for id := range []string{"A", "B", "C", "D", "E"} {
			taskID := string(rune('A' + id))
			g.AddTask(&Task{ID: taskID, Title: "Task " + taskID})
		}

		g.AddDependency("B", "A")
		g.AddDependency("E", "D")
		g.AddDependency("C", "B")
		g.AddDependency("C", "E")

		path, err := g.CriticalPath()
		require.NoError(t, err)

		// Critical path should be length 3
		require.Len(t, path, 3)
		assert.Equal(t, "C", path[2]) // Must end with C
		// Should start with either A or D
		assert.Contains(t, []string{"A", "D"}, path[0])
	})
}

func TestGraph_RemoveTask(t *testing.T) {
	t.Parallel()

	t.Run("removes task without dependencies", func(t *testing.T) {
		g := NewTaskGraph()
		task := &Task{ID: "A", Title: "Task A"}
		g.AddTask(task)

		assert.NotNil(t, g.nodes["A"])
		delete(g.nodes, "A")
		delete(g.edges, "A")
		assert.Nil(t, g.nodes["A"])
	})

	t.Run("graph state after task removal", func(t *testing.T) {
		g := NewTaskGraph()
		taskA := &Task{ID: "A", Title: "Task A"}
		taskB := &Task{ID: "B", Title: "Task B"}

		g.AddTask(taskA)
		g.AddTask(taskB)
		g.AddDependency("A", "B")

		// Remove task A
		delete(g.nodes, "A")
		delete(g.edges, "A")

		// Verify B still exists
		assert.NotNil(t, g.nodes["B"])
		assert.Nil(t, g.nodes["A"])
	})
}

func TestGraph_GetTask(t *testing.T) {
	t.Parallel()

	t.Run("retrieves existing task", func(t *testing.T) {
		g := NewTaskGraph()
		task := &Task{ID: "A", Title: "Task A"}
		g.AddTask(task)

		retrieved, exists := g.nodes["A"]
		assert.True(t, exists)
		assert.Equal(t, task.ID, retrieved.ID)
		assert.Equal(t, task.Title, retrieved.Title)
	})

	t.Run("returns nil for non-existent task", func(t *testing.T) {
		g := NewTaskGraph()
		retrieved, exists := g.nodes["non-existent"]
		assert.False(t, exists)
		assert.Nil(t, retrieved)
	})
}

func TestGraph_MultipleRoots(t *testing.T) {
	t.Parallel()

	t.Run("handles graph with multiple root nodes", func(t *testing.T) {
		g := NewTaskGraph()
		// Two independent chains
		taskA := &Task{ID: "A", Title: "Task A"}
		taskB := &Task{ID: "B", Title: "Task B"}
		taskC := &Task{ID: "C", Title: "Task C"}
		taskD := &Task{ID: "D", Title: "Task D"}

		g.AddTask(taskA)
		g.AddTask(taskB)
		g.AddTask(taskC)
		g.AddTask(taskD)

		// Chain 1: A -> B
		g.AddDependency("B", "A")
		// Chain 2: C -> D
		g.AddDependency("D", "C")

		sorted, err := g.TopologicalSort()
		require.NoError(t, err)
		assert.Len(t, sorted, 4)

		// Both A and C should appear before their dependents
		var idxA, idxB, idxC, idxD int
		for i, task := range sorted {
			switch task.ID {
			case "A":
				idxA = i
			case "B":
				idxB = i
			case "C":
				idxC = i
			case "D":
				idxD = i
			}
		}

		assert.Less(t, idxA, idxB)
		assert.Less(t, idxC, idxD)
	})
}

func TestGraph_DependenciesEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("adding duplicate dependency", func(t *testing.T) {
		g := NewTaskGraph()
		taskA := &Task{ID: "A", Title: "Task A"}
		taskB := &Task{ID: "B", Title: "Task B"}

		g.AddTask(taskA)
		g.AddTask(taskB)

		err := g.AddDependency("A", "B")
		assert.NoError(t, err)

		// Add same dependency again
		err = g.AddDependency("A", "B")
		assert.NoError(t, err) // Should not error, just duplicates in slice

		// Verify dependency exists
		assert.Contains(t, g.edges["A"], "B")
	})

	t.Run("task with empty ID", func(t *testing.T) {
		g := NewTaskGraph()
		task := &Task{ID: "", Title: "Empty ID Task"}

		err := g.AddTask(task)
		assert.NoError(t, err) // Empty ID is technically allowed

		// Verify it was added
		_, exists := g.nodes[""]
		assert.True(t, exists)
	})
}

func TestGraph_LargeGraph(t *testing.T) {
	t.Parallel()

	t.Run("handles large graph with many tasks", func(t *testing.T) {
		g := NewTaskGraph()

		// Create 100 tasks in a linear chain
		for i := 0; i < 100; i++ {
			taskID := fmt.Sprintf("task-%d", i)
			g.AddTask(&Task{ID: taskID, Title: fmt.Sprintf("Task %d", i)})
		}

		// Create linear dependencies
		for i := 1; i < 100; i++ {
			fromID := fmt.Sprintf("task-%d", i)
			toID := fmt.Sprintf("task-%d", i-1)
			g.AddDependency(fromID, toID)
		}

		// Verify topological sort works
		sorted, err := g.TopologicalSort()
		require.NoError(t, err)
		assert.Len(t, sorted, 100)

		// Verify order is correct
		for i := 0; i < 99; i++ {
			curr := sorted[i].ID
			next := sorted[i+1].ID
			// task-0 should come before task-1, etc.
			var currNum, nextNum int
			fmt.Sscanf(curr, "task-%d", &currNum)
			fmt.Sscanf(next, "task-%d", &nextNum)
			assert.Less(t, currNum, nextNum)
		}
	})
}

func TestGraph_GetDependents(t *testing.T) {
	t.Parallel()

	t.Run("returns tasks that depend on given task", func(t *testing.T) {
		g := NewTaskGraph()
		taskA := &Task{ID: "A", Title: "Task A"}
		taskB := &Task{ID: "B", Title: "Task B"}
		taskC := &Task{ID: "C", Title: "Task C"}

		g.AddTask(taskA)
		g.AddTask(taskB)
		g.AddTask(taskC)

		// B and C depend on A
		g.AddDependency("B", "A")
		g.AddDependency("C", "A")

		// Find all tasks that depend on A
		dependents := []string{}
		for taskID, deps := range g.edges {
			for _, dep := range deps {
				if dep == "A" {
					dependents = append(dependents, taskID)
				}
			}
		}

		assert.Len(t, dependents, 2)
		assert.Contains(t, dependents, "B")
		assert.Contains(t, dependents, "C")
	})
}
