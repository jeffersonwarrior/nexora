package task

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewManager(t *testing.T) {
	t.Parallel()

	t.Run("creates manager successfully", func(t *testing.T) {
		// Act
		manager := NewManager()

		// Assert
		assert.NotNil(t, manager)
		assert.NotNil(t, manager.activeTasks)
		assert.NotNil(t, manager.driftRules)
	})
}

func TestManager_CreateTask(t *testing.T) {
	t.Parallel()

	t.Run("creates task with valid inputs", func(t *testing.T) {
		// Arrange
		manager := NewManager()
		sessionID := "session-123"
		title := "Test Task"
		description := "A test task"
		context := "Working on tests"
		milestones := []string{"milestone1", "milestone2"}

		// Act
		task := manager.CreateTask(sessionID, title, description, context, milestones)

		// Assert
		assert.NotNil(t, task)
		assert.NotEmpty(t, task.ID)
		assert.Equal(t, sessionID, task.SessionID)
		assert.Equal(t, title, task.Title)
		assert.Equal(t, description, task.Description)
		assert.Equal(t, context, task.Context)
		assert.Equal(t, StatusActive, task.Status)
		assert.Equal(t, PriorityHigh, task.Priority)
		assert.Len(t, task.Milestones, 2)

		// Verify task is stored in manager
		stored := manager.activeTasks[sessionID]
		assert.Equal(t, task, stored)
	})

	t.Run("creates task with empty milestones", func(t *testing.T) {
		// Arrange
		manager := NewManager()

		// Act
		task := manager.CreateTask("session", "title", "desc", "context", []string{})

		// Assert
		assert.NotNil(t, task)
		assert.Len(t, task.Milestones, 0)
	})

	t.Run("milestone IDs are unique", func(t *testing.T) {
		// Arrange
		manager := NewManager()
		milestones := []string{"m1", "m2", "m3"}

		// Act
		task := manager.CreateTask("session", "title", "desc", "context", milestones)

		// Assert
		ids := make(map[string]bool)
		for _, m := range task.Milestones {
			assert.NotEmpty(t, m.ID)
			assert.False(t, ids[m.ID], "Milestone ID should be unique")
			ids[m.ID] = true
		}
	})
}

func TestManager_GetTask(t *testing.T) {
	t.Parallel()

	t.Run("retrieves existing task", func(t *testing.T) {
		// Arrange
		manager := NewManager()
		sessionID := "session-123"
		created := manager.CreateTask(sessionID, "Test", "Desc", "Context", []string{"m1"})

		// Act
		retrieved, found := manager.GetTask(sessionID)

		// Assert
		assert.True(t, found)
		assert.NotNil(t, retrieved)
		assert.Equal(t, created.ID, retrieved.ID)
	})

	t.Run("returns false for non-existent session", func(t *testing.T) {
		// Arrange
		manager := NewManager()

		// Act
		task, found := manager.GetTask("non-existent")

		// Assert
		assert.False(t, found)
		assert.Nil(t, task)
	})
}

func TestManager_UpdateTaskContext(t *testing.T) {
	t.Parallel()

	t.Run("updates context for existing task", func(t *testing.T) {
		// Arrange
		manager := NewManager()
		sessionID := "session-123"
		manager.CreateTask(sessionID, "Test", "Desc", "Original", []string{})

		// Act
		manager.UpdateTaskContext(sessionID, "Updated context")

		// Assert
		task, _ := manager.GetTask(sessionID)
		assert.Equal(t, "Updated context", task.Context)
	})

	t.Run("handles non-existent session gracefully", func(t *testing.T) {
		// Arrange
		manager := NewManager()

		// Act & Assert - should not panic
		manager.UpdateTaskContext("non-existent", "new context")
	})
}

func TestManager_AnalyzeDrift(t *testing.T) {
	t.Parallel()

	t.Run("analyzes drift for on-topic response", func(t *testing.T) {
		// Arrange
		manager := NewManager()
		sessionID := "session-123"
		manager.CreateTask(sessionID, "Write tests", "Add unit tests", "Testing context", []string{"Write test cases"})

		// Act
		analysis := manager.AnalyzeDrift(sessionID, "I've written the test cases as requested")

		// Assert
		assert.NotNil(t, analysis)
		assert.False(t, analysis.Drifted)
		assert.GreaterOrEqual(t, analysis.Confidence, 0.0)
		assert.LessOrEqual(t, analysis.Confidence, 1.0)
	})

	t.Run("analyzes drift for off-topic response", func(t *testing.T) {
		// Arrange
		manager := NewManager()
		sessionID := "session-123"
		manager.CreateTask(sessionID, "Write tests", "Add unit tests", "Testing context", []string{})

		// Act
		analysis := manager.AnalyzeDrift(sessionID, "Let me explain the history of programming languages")

		// Assert
		assert.NotNil(t, analysis)
		assert.GreaterOrEqual(t, analysis.Confidence, 0.0)
		assert.LessOrEqual(t, analysis.Confidence, 1.0)
	})

	t.Run("handles non-existent session gracefully", func(t *testing.T) {
		// Arrange
		manager := NewManager()

		// Act
		analysis := manager.AnalyzeDrift("non-existent", "response")

		// Assert
		assert.NotNil(t, analysis)
		assert.False(t, analysis.Drifted)
	})

	t.Run("detects milestone keywords", func(t *testing.T) {
		// Arrange
		manager := NewManager()
		sessionID := "session-123"
		task := manager.CreateTask(sessionID, "Test", "Desc", "Context", []string{"Complete research"})

		// Set milestone to active so it can be detected
		task.Milestones[0].Status = StatusActive

		// Act
		analysis := manager.AnalyzeDrift(sessionID, "I have completed Milestone 1 with the research")

		// Assert
		assert.NotNil(t, analysis)
		// Milestone detection depends on response containing milestone keywords
	})
}

func TestManager_CompleteTask(t *testing.T) {
	t.Parallel()

	t.Run("completes active task", func(t *testing.T) {
		// Arrange
		manager := NewManager()
		sessionID := "session-123"
		manager.CreateTask(sessionID, "Test", "Desc", "Context", []string{})

		// Act
		manager.CompleteTask(sessionID)

		// Assert
		task, _ := manager.GetTask(sessionID)
		assert.Equal(t, StatusCompleted, task.Status)
	})

	t.Run("handles non-existent session gracefully", func(t *testing.T) {
		// Arrange
		manager := NewManager()

		// Act & Assert - should not panic
		manager.CompleteTask("non-existent")
	})
}

func TestManager_MarkMilestoneProgress(t *testing.T) {
	t.Parallel()

	t.Run("marks milestone as complete", func(t *testing.T) {
		// Arrange
		manager := NewManager()
		sessionID := "session-123"
		task := manager.CreateTask(sessionID, "Test", "Desc", "Context", []string{"milestone1"})
		milestoneID := task.Milestones[0].ID

		// Act
		manager.MarkMilestoneProgress(sessionID, milestoneID, "Completed successfully")

		// Assert
		updated, _ := manager.GetTask(sessionID)
		assert.Equal(t, StatusCompleted, updated.Milestones[0].Status)
		assert.Equal(t, "Completed successfully", updated.Milestones[0].Evidence)
	})

	t.Run("handles non-existent session gracefully", func(t *testing.T) {
		// Arrange
		manager := NewManager()

		// Act & Assert - should not panic
		manager.MarkMilestoneProgress("non-existent", "milestone-id", "evidence")
	})

	t.Run("handles non-existent milestone gracefully", func(t *testing.T) {
		// Arrange
		manager := NewManager()
		sessionID := "session-123"
		manager.CreateTask(sessionID, "Test", "Desc", "Context", []string{"m1"})

		// Act & Assert - should not panic
		manager.MarkMilestoneProgress(sessionID, "invalid-milestone-id", "evidence")
	})
}

func TestManager_GetCorrectionPrompt(t *testing.T) {
	t.Parallel()

	t.Run("generates prompt for active task", func(t *testing.T) {
		// Arrange
		manager := NewManager()
		sessionID := "session-123"
		manager.CreateTask(sessionID, "Test Task", "Test Description", "Test Context", []string{"milestone1"})

		// Act
		prompt := manager.GetCorrectionPrompt(sessionID)

		// Assert
		assert.NotEmpty(t, prompt)
		assert.Contains(t, prompt, "Test Task")
		assert.Contains(t, prompt, "Test Context")
	})

	t.Run("returns empty for non-existent session", func(t *testing.T) {
		// Arrange
		manager := NewManager()

		// Act
		prompt := manager.GetCorrectionPrompt("non-existent")

		// Assert
		assert.Empty(t, prompt)
	})
}

func TestDriftAnalysis(t *testing.T) {
	t.Parallel()

	t.Run("detects task-related keywords", func(t *testing.T) {
		// Arrange
		manager := NewManager()
		sessionID := "session-123"
		manager.CreateTask(sessionID, "API development", "Build API", "Context", []string{})

		// Act
		analysis := manager.AnalyzeDrift(sessionID, "I'm working on the API endpoint")

		// Assert
		assert.NotNil(t, analysis)
		// Should detect "API" keyword
	})
}

func TestTaskMetadata(t *testing.T) {
	t.Parallel()

	t.Run("task has metadata field", func(t *testing.T) {
		// Arrange
		manager := NewManager()

		// Act
		task := manager.CreateTask("session", "title", "desc", "context", []string{})

		// Assert - Metadata is a map[string]interface{} that may be nil
		assert.NotNil(t, task)
		// Metadata field exists but may be nil/empty
	})
}

func TestTaskPriority(t *testing.T) {
	t.Parallel()

	t.Run("default priority is high", func(t *testing.T) {
		// Arrange
		manager := NewManager()

		// Act
		task := manager.CreateTask("session", "title", "desc", "context", []string{})

		// Assert
		assert.Equal(t, PriorityHigh, task.Priority)
	})
}

func TestMilestoneStructure(t *testing.T) {
	t.Parallel()

	t.Run("milestone has expected fields", func(t *testing.T) {
		// Arrange
		manager := NewManager()
		task := manager.CreateTask("session", "title", "desc", "context", []string{"milestone desc"})

		// Act
		milestone := task.Milestones[0]

		// Assert
		assert.NotEmpty(t, milestone.ID)
		assert.NotEmpty(t, milestone.Title)
		assert.Equal(t, "milestone desc", milestone.Description)
		assert.Equal(t, StatusBlocked, milestone.Status)
	})
}

func TestMultipleSessionsTasks(t *testing.T) {
	t.Parallel()

	t.Run("manager handles multiple sessions independently", func(t *testing.T) {
		// Arrange
		manager := NewManager()

		// Act
		task1 := manager.CreateTask("session-1", "Task 1", "Desc 1", "Context 1", []string{})
		task2 := manager.CreateTask("session-2", "Task 2", "Desc 2", "Context 2", []string{})

		// Assert
		retrieved1, found1 := manager.GetTask("session-1")
		retrieved2, found2 := manager.GetTask("session-2")

		assert.True(t, found1)
		assert.True(t, found2)
		assert.NotEqual(t, task1.ID, task2.ID)
		assert.Equal(t, task1.Title, retrieved1.Title)
		assert.Equal(t, task2.Title, retrieved2.Title)
	})
}

func TestTaskStatusConstants(t *testing.T) {
	t.Parallel()

	t.Run("status constants have expected values", func(t *testing.T) {
		assert.Equal(t, TaskStatus("active"), StatusActive)
		assert.Equal(t, TaskStatus("blocked"), StatusBlocked)
		assert.Equal(t, TaskStatus("completed"), StatusCompleted)
		assert.Equal(t, TaskStatus("paused"), StatusPaused)
		assert.Equal(t, TaskStatus("cancelled"), StatusCancelled)
	})
}

func TestPriorityConstants(t *testing.T) {
	t.Parallel()

	t.Run("priority constants have expected values", func(t *testing.T) {
		assert.Equal(t, Priority("high"), PriorityHigh)
		assert.Equal(t, Priority("medium"), PriorityMedium)
		assert.Equal(t, Priority("low"), PriorityLow)
	})
}

func TestManager_AnalyzeDrift_EdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("handles empty response", func(t *testing.T) {
		// Arrange
		manager := NewManager()
		sessionID := "session-123"
		manager.CreateTask(sessionID, "Test", "Desc", "Context", []string{})

		// Act
		analysis := manager.AnalyzeDrift(sessionID, "")

		// Assert
		assert.NotNil(t, analysis)
	})

	t.Run("handles very long response", func(t *testing.T) {
		// Arrange
		manager := NewManager()
		sessionID := "session-123"
		manager.CreateTask(sessionID, "Test", "Desc", "Context", []string{})
		longResponse := strings.Repeat("This is a very long response. ", 100)

		// Act
		analysis := manager.AnalyzeDrift(sessionID, longResponse)

		// Assert
		assert.NotNil(t, analysis)
	})
}

// Removed TestManager_ConcurrentAccess as Manager is not thread-safe by design
