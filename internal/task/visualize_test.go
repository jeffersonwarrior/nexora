package task

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVisualize_ASCII(t *testing.T) {
	t.Parallel()
	t.Run("renders single task", func(t *testing.T) {
		g := NewTaskGraph()
		task := &Task{
			ID:    "A",
			Title: "Task A",
			Milestones: []Milestone{
				{Status: StatusCompleted},
			},
		}
		g.AddTask(task)

		output := g.VisualizeASCII("A")
		assert.NotEmpty(t, output)
		assert.Contains(t, output, "Task A")
		assert.Contains(t, output, "100%")
	})

	t.Run("renders task with dependencies", func(t *testing.T) {
		g := NewTaskGraph()
		taskA := &Task{
			ID:    "A",
			Title: "Build Feature X",
			Milestones: []Milestone{
				{Status: StatusActive},
			},
		}
		taskB := &Task{
			ID:    "B",
			Title: "Design API",
			Milestones: []Milestone{
				{Status: StatusCompleted},
			},
		}
		taskC := &Task{
			ID:    "C",
			Title: "Write tests",
			Milestones: []Milestone{
				{Status: StatusBlocked},
			},
		}

		g.AddTask(taskA)
		g.AddTask(taskB)
		g.AddTask(taskC)
		g.AddDependency("A", "B")
		g.AddDependency("A", "C")

		output := g.VisualizeASCII("A")
		t.Logf("Output:\n%s", output)
		assert.NotEmpty(t, output)
		assert.Contains(t, output, "Build Feature X")
		assert.Contains(t, output, "Design API")
		assert.Contains(t, output, "Write tests")
		// Should contain tree characters
		hasTreeChars := strings.Contains(output, "├──") || strings.Contains(output, "└──")
		if !hasTreeChars {
			t.Logf("Output does not contain tree characters")
		}
		assert.True(t, hasTreeChars)
	})

	t.Run("renders deep dependency chain", func(t *testing.T) {
		g := NewTaskGraph()
		taskA := &Task{ID: "A", Title: "A", Milestones: []Milestone{{Status: StatusActive}}}
		taskB := &Task{ID: "B", Title: "B", Milestones: []Milestone{{Status: StatusActive}}}
		taskC := &Task{ID: "C", Title: "C", Milestones: []Milestone{{Status: StatusCompleted}}}

		g.AddTask(taskA)
		g.AddTask(taskB)
		g.AddTask(taskC)
		g.AddDependency("A", "B")
		g.AddDependency("B", "C")

		output := g.VisualizeASCII("A")
		assert.NotEmpty(t, output)
		assert.Contains(t, output, "A")
		assert.Contains(t, output, "B")
		assert.Contains(t, output, "C")
	})

	t.Run("returns error message for non-existent task", func(t *testing.T) {
		g := NewTaskGraph()
		output := g.VisualizeASCII("non-existent")
		assert.NotEmpty(t, output)
		assert.Contains(t, output, "not found")
	})

	t.Run("shows status indicators", func(t *testing.T) {
		g := NewTaskGraph()
		taskA := &Task{
			ID:     "A",
			Title:  "Task A",
			Status: StatusActive,
			Milestones: []Milestone{
				{Status: StatusActive},
			},
		}
		taskB := &Task{
			ID:     "B",
			Title:  "Task B",
			Status: StatusCompleted,
			Milestones: []Milestone{
				{Status: StatusCompleted},
			},
		}
		taskC := &Task{
			ID:     "C",
			Title:  "Task C",
			Status: StatusBlocked,
			Milestones: []Milestone{
				{Status: StatusBlocked},
			},
		}

		g.AddTask(taskA)
		g.AddTask(taskB)
		g.AddTask(taskC)
		g.AddDependency("A", "B")
		g.AddDependency("A", "C")

		output := g.VisualizeASCII("A")
		assert.NotEmpty(t, output)
		// Should contain progress indicators
		assert.Contains(t, output, "%")
	})
}

func TestFormatProgress(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		progress float64
		want     string
	}{
		{"zero percent", 0.0, "0%"},
		{"fifty percent", 0.5, "50%"},
		{"hundred percent", 1.0, "100%"},
		{"partial percent", 0.333, "33%"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatProgress(tt.progress)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestGetStatusSymbol(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		status   TaskStatus
		progress float64
		want     string
	}{
		{"completed task", StatusCompleted, 1.0, "done"},
		{"active task with progress", StatusActive, 0.5, "progress"},
		{"active task no progress", StatusActive, 0.0, "pending"},
		{"blocked task", StatusBlocked, 0.0, "blocked"},
		{"paused task", StatusPaused, 0.0, "paused"},
		{"cancelled task", StatusCancelled, 0.0, "cancelled"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getStatusSymbol(tt.status, tt.progress)
			assert.Equal(t, tt.want, result)
		})
	}
}
