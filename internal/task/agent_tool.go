package task

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// TaskTool provides agent-accessible task management capabilities
type TaskTool struct {
	taskService Service
	sessionID   string
}

// NewTaskTool creates a task management tool for agent use
func NewTaskTool(taskService Service, sessionID string) *TaskTool {
	return &TaskTool{
		taskService: taskService,
		sessionID:   sessionID,
	}
}

// AgentTaskInput represents input for task operations
type AgentTaskInput struct {
	TaskID      string   `json:"task_id,omitempty"`
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	Milestones  []string `json:"milestones,omitempty"`
	Message     string   `json:"message,omitempty"`
}

// AgentTaskOutput represents output from task operations
type AgentTaskOutput struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message"`
	TaskStatus TaskSummary `json:"task_status,omitempty"`
	DriftAlert bool        `json:"drift_alert,omitempty"`
	DriftMsg   string      `json:"drift_message,omitempty"`
}

// StartTask begins a new focused task
func (tt *TaskTool) StartTask(ctx context.Context, input AgentTaskInput) (*AgentTaskOutput, error) {
	if input.Title == "" {
		return &AgentTaskOutput{
			Success: false,
			Message: "Task title is required",
		}, nil
	}

	task := tt.taskService.CreateTask(ctx, tt.sessionID, input.Title, input.Description, fmt.Sprintf("Focus: %s", input.Description), input.Milestones)

	return &AgentTaskOutput{
		Success: true,
		Message: fmt.Sprintf("Started task: %s", task.Title),
	}, nil
}

// UpdateProgress marks milestone progress
func (tt *TaskTool) UpdateProgress(ctx context.Context, input AgentTaskInput) (*AgentTaskOutput, error) {
	if input.Message == "" {
		return &AgentTaskOutput{
			Success: false,
			Message: "Progress message is required",
		}, nil
	}

	// Update task context with progress info
	tt.taskService.UpdateTaskContext(ctx, tt.sessionID, input.Message)

	return &AgentTaskOutput{
		Success: true,
		Message: "Task context updated",
	}, nil
}

// CheckDrift monitors for focus drift
func (tt *TaskTool) CheckDrift(ctx context.Context, input AgentTaskInput) (*AgentTaskOutput, error) {
	if input.Message == "" {
		return &AgentTaskOutput{
			Success: false,
			Message: "Message to analyze is required",
		}, nil
	}

	analysis := tt.taskService.AnalyzeDrift(ctx, tt.sessionID, input.Message)

	if analysis.Drifted {
		return &AgentTaskOutput{
			Success:    true,
			Message:    "Drift detected",
			DriftAlert: true,
			DriftMsg:   strings.Join(analysis.Recommendations, "; "),
		}, nil
	}

	return &AgentTaskOutput{
		Success:    true,
		Message:    "Focus maintained",
		DriftAlert: false,
	}, nil
}

// GetStatus returns current task status
func (tt *TaskTool) GetStatus(ctx context.Context, input AgentTaskInput) (*AgentTaskOutput, error) {
	summary := tt.taskService.GetTaskSummary(ctx, tt.sessionID)

	if summary.Title == "" {
		return &AgentTaskOutput{
			Success: true,
			Message: "No active task",
		}, nil
	}

	return &AgentTaskOutput{
		Success:    true,
		Message:    fmt.Sprintf("Task: %s (%.0f%% complete)", summary.Title, summary.Progress*100),
		TaskStatus: summary,
	}, nil
}

// CompleteTask marks the current task as completed
func (tt *TaskTool) CompleteTask(ctx context.Context, input AgentTaskInput) (*AgentTaskOutput, error) {
	summary := tt.taskService.GetTaskSummary(ctx, tt.sessionID)

	if summary.Title == "" {
		return &AgentTaskOutput{
			Success: false,
			Message: "No active task to complete",
		}, nil
	}

	tt.taskService.CompleteTask(ctx, tt.sessionID)

	return &AgentTaskOutput{
		Success: true,
		Message: fmt.Sprintf("Task completed: %s", summary.Title),
	}, nil
}

// TaskToolSet returns the complete set of task management tools
func (tt *TaskTool) TaskToolSet() map[string]func(context.Context, AgentTaskInput) (*AgentTaskOutput, error) {
	return map[string]func(context.Context, AgentTaskInput) (*AgentTaskOutput, error){
		"start_task":      tt.StartTask,
		"update_progress": tt.UpdateProgress,
		"check_drift":     tt.CheckDrift,
		"get_status":      tt.GetStatus,
		"complete_task":   tt.CompleteTask,
	}
}

// Convert to Cobra CLI commands for testing/admin
func (tt *TaskTool) GetTaskCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Manage AI task focus and progress",
		Long:  "Task management tools for keeping AI agents focused and on track",
	}

	// Status command
	statusCmd := &cobra.Command{
		Use: "status",
		Run: func(cmd *cobra.Command, args []string) {
			output, _ := tt.GetStatus(cmd.Context(), AgentTaskInput{})
			if output.Success {
				if output.TaskStatus.Title != "" {
					fmt.Printf("Task: %s\n", output.TaskStatus.Title)
					fmt.Printf("Progress: %.0f%%\n", output.TaskStatus.Progress*100)
					if output.TaskStatus.NextMilestone != nil {
						fmt.Printf("Next: %s - %s\n", output.TaskStatus.NextMilestone.Title, output.TaskStatus.NextMilestone.Description)
					}
				} else {
					fmt.Println("No active task")
				}
			} else {
				fmt.Printf("Error: %s\n", output.Message)
			}
		},
	}

	// Complete command
	completeCmd := &cobra.Command{
		Use: "complete",
		Run: func(cmd *cobra.Command, args []string) {
			output, _ := tt.CompleteTask(cmd.Context(), AgentTaskInput{})
			if output.Success {
				fmt.Println(output.Message)
			} else {
				fmt.Printf("Error: %s\n", output.Message)
			}
		},
	}

	cmd.AddCommand(statusCmd, completeCmd)
	return cmd
}
