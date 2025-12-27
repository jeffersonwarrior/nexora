package task

import (
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
