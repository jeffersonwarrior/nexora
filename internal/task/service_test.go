package task

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewService(t *testing.T) {
	t.Run("creates service successfully", func(t *testing.T) {
		// Act
		service := NewService()

		// Assert
		assert.NotNil(t, service)
		assert.Implements(t, (*Service)(nil), service)
	})
}

func TestTaskService_CreateTask(t *testing.T) {
	t.Run("creates task with valid inputs", func(t *testing.T) {
		// Arrange
		service := NewService()
		ctx := context.Background()
		sessionID := "session-123"
		title := "Test Task"
		description := "A test task"
		taskContext := "Working on tests"
		milestones := []string{"milestone1", "milestone2"}

		// Act
		task := service.CreateTask(ctx, sessionID, title, description, taskContext, milestones)

		// Assert
		assert.NotNil(t, task)
		assert.Equal(t, sessionID, task.SessionID)
		assert.Equal(t, title, task.Title)
		assert.Equal(t, description, task.Description)
		assert.Equal(t, taskContext, task.Context)
		assert.Equal(t, StatusActive, task.Status)
		assert.Equal(t, PriorityHigh, task.Priority) // Default priority is high
		assert.Len(t, task.Milestones, 2)

		// Check milestones - titles are auto-generated as "Milestone N"
		assert.Equal(t, "Milestone 1", task.Milestones[0].Title)
		assert.Equal(t, "Milestone 2", task.Milestones[1].Title)
		assert.Equal(t, StatusBlocked, task.Milestones[0].Status) // Initial milestone status
		assert.Equal(t, StatusBlocked, task.Milestones[1].Status)
	})

	t.Run("creates task with empty milestones", func(t *testing.T) {
		// Arrange
		service := NewService()
		ctx := context.Background()

		// Act
		task := service.CreateTask(ctx, "session", "title", "desc", "context", []string{})

		// Assert
		assert.NotNil(t, task)
		assert.Len(t, task.Milestones, 0)
	})
}

func TestTaskService_GetActiveTask(t *testing.T) {
	t.Run("returns newly created active task", func(t *testing.T) {
		// Arrange
		service := NewService()
		ctx := context.Background()
		sessionID := "session-123"

		// Act
		createdTask := service.CreateTask(ctx, sessionID, "Test", "Desc", "Context", []string{"m1"})
		activeTask, found := service.GetActiveTask(ctx, sessionID)

		// Assert
		assert.True(t, found)
		assert.NotNil(t, activeTask)
		assert.Equal(t, createdTask.ID, activeTask.ID)
		assert.Equal(t, createdTask.Title, activeTask.Title)
	})

	t.Run("returns false for non-existent session", func(t *testing.T) {
		// Arrange
		service := NewService()
		ctx := context.Background()

		// Act
		_, found := service.GetActiveTask(ctx, "non-existent")

		// Assert
		assert.False(t, found)
	})
}

func TestTaskService_UpdateTaskContext(t *testing.T) {
	t.Run("updates context for active task", func(t *testing.T) {
		// Arrange
		service := NewService()
		ctx := context.Background()
		sessionID := "session-123"
		service.CreateTask(ctx, sessionID, "Test", "Desc", "Original context", []string{})

		// Act
		service.UpdateTaskContext(ctx, sessionID, "Updated context")

		// Assert
		task, found := service.GetActiveTask(ctx, sessionID)
		assert.True(t, found)
		assert.Equal(t, "Updated context", task.Context)
	})

	t.Run("handles non-existent session gracefully", func(t *testing.T) {
		// Arrange
		service := NewService()
		ctx := context.Background()

		// Act & Assert - should not panic
		service.UpdateTaskContext(ctx, "non-existent", "new context")
	})
}

func TestTaskService_CompleteTask(t *testing.T) {
	t.Run("completes active task", func(t *testing.T) {
		// Arrange
		service := NewService()
		ctx := context.Background()
		sessionID := "session-123"
		service.CreateTask(ctx, sessionID, "Test", "Desc", "Context", []string{"m1"})

		// Act
		service.CompleteTask(ctx, sessionID)

		// Assert - completed task is still in active map but with completed status
		task, found := service.GetActiveTask(ctx, sessionID)
		assert.True(t, found) // Still in active map
		assert.NotNil(t, task)
		assert.Equal(t, StatusCompleted, task.Status)
	})

	t.Run("handles non-existent session gracefully", func(t *testing.T) {
		// Arrange
		service := NewService()
		ctx := context.Background()

		// Act & Assert - should not panic
		service.CompleteTask(ctx, "non-existent")
	})
}

func TestTaskService_CloseTask(t *testing.T) {
	t.Run("closes active task", func(t *testing.T) {
		// Arrange
		service := NewService()
		ctx := context.Background()
		sessionID := "session-123"
		service.CreateTask(ctx, sessionID, "Test", "Desc", "Context", []string{"m1"})

		// Act
		service.CloseTask(ctx, sessionID)

		// Assert
		task, found := service.GetActiveTask(ctx, sessionID)
		assert.False(t, found) // No longer active
		if found {
			assert.Equal(t, StatusCancelled, task.Status)
		}
	})

	t.Run("handles non-existent session gracefully", func(t *testing.T) {
		// Arrange
		service := NewService()
		ctx := context.Background()

		// Act & Assert - should not panic
		service.CloseTask(ctx, "non-existent")
	})
}

func TestTaskService_DriftAnalysis(t *testing.T) {
	t.Run("analyzes drift for simple response", func(t *testing.T) {
		// Arrange
		service := NewService()
		ctx := context.Background()
		sessionID := "session-123"
		service.CreateTask(ctx, sessionID, "Test", "Desc", "Context", []string{"m1"})

		// Act
		analysis := service.AnalyzeDrift(ctx, sessionID, "This is a response about something else")

		// Assert - DriftAnalysis is a struct, check its fields
		assert.NotEmpty(t, analysis.Message)
		assert.GreaterOrEqual(t, analysis.Confidence, 0.0)
		assert.LessOrEqual(t, analysis.Confidence, 1.0)
	})

	t.Run("handles non-existent session gracefully", func(t *testing.T) {
		// Arrange
		service := NewService()
		ctx := context.Background()

		// Act
		analysis := service.AnalyzeDrift(ctx, "non-existent", "response")

		// Assert
		assert.NotNil(t, analysis)
		assert.Empty(t, analysis.MilestoneProgress)
		assert.Empty(t, analysis.Recommendations)
	})
}

func TestTaskService_GetCorrectionPrompt(t *testing.T) {
	t.Run("returns prompt for active task", func(t *testing.T) {
		// Arrange
		service := NewService()
		ctx := context.Background()
		sessionID := "session-123"
		service.CreateTask(ctx, sessionID, "Test Task", "Desc", "Context", []string{"milestone1"})

		// Act
		prompt := service.GetCorrectionPrompt(ctx, sessionID)

		// Assert
		assert.NotEmpty(t, prompt)
		assert.Contains(t, prompt, "Test Task")
		assert.Contains(t, prompt, "Context")
	})

	t.Run("returns empty for non-existent session", func(t *testing.T) {
		// Arrange
		service := NewService()
		ctx := context.Background()

		// Act
		prompt := service.GetCorrectionPrompt(ctx, "non-existent")

		// Assert
		assert.Empty(t, prompt)
	})
}

func TestTaskService_MarkMilestoneProgress(t *testing.T) {
	t.Run("marks milestone progress", func(t *testing.T) {
		// Arrange
		service := NewService()
		ctx := context.Background()
		sessionID := "session-123"
		task := service.CreateTask(ctx, sessionID, "Test", "Desc", "Context", []string{"milestone1"})
		milestoneID := task.Milestones[0].ID

		// Act
		service.MarkMilestoneProgress(ctx, sessionID, milestoneID, "Completed the work")

		// Assert
		updatedTask, found := service.GetActiveTask(ctx, sessionID)
		assert.True(t, found)

		for _, milestone := range updatedTask.Milestones {
			if milestone.ID == milestoneID {
				assert.Equal(t, StatusCompleted, milestone.Status)
				assert.Equal(t, "Completed the work", milestone.Evidence)
			}
		}
	})

	t.Run("handles non-existent session gracefully", func(t *testing.T) {
		// Arrange
		service := NewService()
		ctx := context.Background()

		// Act & Assert - should not panic
		service.MarkMilestoneProgress(ctx, "non-existent", "milestone-id", "evidence")
	})
}

func TestTaskService_GetTaskSummary(t *testing.T) {
	t.Run("returns summary for active task", func(t *testing.T) {
		// Arrange
		service := NewService()
		ctx := context.Background()
		sessionID := "session-123"
		service.CreateTask(ctx, sessionID, "Test Task", "Description", "Context", []string{"milestone1"})

		// Act
		summary := service.GetTaskSummary(ctx, sessionID)

		// Assert
		assert.NotEmpty(t, summary.TaskID)
		assert.Equal(t, "Test Task", summary.Title)
		assert.Equal(t, StatusActive, summary.Status)
		assert.GreaterOrEqual(t, summary.Progress, 0.0)
		assert.LessOrEqual(t, summary.Progress, 1.0)
	})

	t.Run("returns empty summary for non-existent session", func(t *testing.T) {
		// Arrange
		service := NewService()
		ctx := context.Background()

		// Act
		summary := service.GetTaskSummary(ctx, "non-existent")

		// Assert - empty summary has zero values
		assert.Empty(t, summary.TaskID)
		assert.Empty(t, summary.Title)
		assert.Empty(t, summary.Status)
	})
}

// Task data structure tests
func TestTask_StatusTransitions(t *testing.T) {
	t.Run("valid status values", func(t *testing.T) {
		// Arrange & Act & Assert
		validStatuses := []TaskStatus{
			StatusActive,
			StatusBlocked,
			StatusCompleted,
			StatusPaused,
			StatusCancelled,
		}

		for _, status := range validStatuses {
			assert.NotEmpty(t, string(status))
		}
	})
}

func TestTask_PriorityLevels(t *testing.T) {
	t.Run("valid priority values", func(t *testing.T) {
		// Arrange & Act & Assert
		validPriorities := []Priority{
			PriorityHigh,
			PriorityMedium,
			PriorityLow,
		}

		for _, priority := range validPriorities {
			assert.NotEmpty(t, string(priority))
		}
	})
}

// Integration tests to verify the complete flow
func TestTaskService_Workflow(t *testing.T) {
	t.Run("complete task workflow", func(t *testing.T) {
		// Arrange
		service := NewService()
		ctx := context.Background()
		sessionID := "workflow-test"

		// 1. Create task
		task := service.CreateTask(ctx, sessionID, "Workflow Task", "Test workflow", "Initial context", []string{"Research", "Implementation", "Testing"})
		assert.NotNil(t, task)
		assert.Equal(t, 3, len(task.Milestones))

		// 2. Get active task
		activeTask, found := service.GetActiveTask(ctx, sessionID)
		assert.True(t, found)
		assert.Equal(t, task.ID, activeTask.ID)

		// 3. Update context
		service.UpdateTaskContext(ctx, sessionID, "Updated context after research")
		updatedTask, _ := service.GetActiveTask(ctx, sessionID)
		assert.Equal(t, "Updated context after research", updatedTask.Context)

		// 4. Mark milestone progress
		milestoneID := updatedTask.Milestones[0].ID
		service.MarkMilestoneProgress(ctx, sessionID, milestoneID, "Research completed")

		// 5. Get summary
		summary := service.GetTaskSummary(ctx, sessionID)
		assert.Equal(t, "Workflow Task", summary.Title)
		assert.Equal(t, StatusActive, summary.Status)

		// 6. Complete task
		service.CompleteTask(ctx, sessionID)
		completedTask, found := service.GetActiveTask(ctx, sessionID)
		assert.True(t, found) // Still in active map after completion
		assert.Equal(t, StatusCompleted, completedTask.Status)
	})
}
