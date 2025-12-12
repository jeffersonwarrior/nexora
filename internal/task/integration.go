package task

import (
	"context"
	"fmt"
	"strings"
)

// Config for task management behavior
type Config struct {
	EnableAutoTaskCreation   bool    `yaml:"enable_auto_task_creation"`
	EnableDriftDetection     bool    `yaml:"enable_drift_detection"`
	DriftThreshold           float64 `yaml:"drift_threshold"` // confidence level to trigger correction
	MaxActiveTasksPerSession int     `yaml:"max_active_tasks_per_session"`
}

// DefaultConfig returns default task management settings
func DefaultConfig() Config {
	return Config{
		EnableAutoTaskCreation:   true,
		EnableDriftDetection:     true,
		DriftThreshold:           0.7,
		MaxActiveTasksPerSession: 1,
	}
}

// TaskAwareCoordinator adds task management to coordinator
type TaskAwareCoordinator struct {
	base        Coordinator
	taskService Service
	config      Config
}

// NewTaskAwareCoordinator creates a coordinator with task management capabilities
func NewTaskAwareCoordinator(
	baseCoordinator Coordinator,
	taskService Service,
	config Config,
) *TaskAwareCoordinator {
	return &TaskAwareCoordinator{
		base:        baseCoordinator,
		taskService: taskService,
		config:      config,
	}
}

// Run implements Coordinator with task context injection
func (tac *TaskAwareCoordinator) Run(ctx context.Context, sessionID, prompt string, attachments ...any) (interface{}, error) {
	// Auto-create task if enabled and no active task exists
	if tac.config.EnableAutoTaskCreation {
		_, exists := tac.taskService.GetActiveTask(ctx, sessionID)
		if !exists {
			// Create automatic task from prompt
			taskTitle := summarizePrompt(prompt)
			taskDescription := prompt
			tac.taskService.CreateTask(ctx, sessionID, taskTitle, taskDescription, prompt, []string{"Complete the requested work"})
		}
	}

	// Inject task context
	enhancedPrompt := tac.enhancePromptWithContext(ctx, sessionID, prompt)

	// Run the agent with enhanced prompt
	result, err := tac.base.Run(ctx, sessionID, enhancedPrompt, attachments...)
	if err != nil {
		return nil, err
	}

	// Check for drift if enabled
	if tac.config.EnableDriftDetection {
		response := extractResponseContent(result)
		analysis := tac.taskService.AnalyzeDrift(ctx, sessionID, response)

		if analysis.Drifted && analysis.Confidence >= tac.config.DriftThreshold {
			// Get correction prompt and re-run
			correctionPrompt := tac.taskService.GetCorrectionPrompt(ctx, sessionID)
			enhancedPrompt = correctionPrompt + "\n\n" + enhancedPrompt
			return tac.base.Run(ctx, sessionID, enhancedPrompt, attachments...)
		}
	}

	return result, nil
}

// enhancePromptWithContext adds task information to user prompts
func (tac *TaskAwareCoordinator) enhancePromptWithContext(ctx context.Context, sessionID, prompt string) string {
	task, exists := tac.taskService.GetActiveTask(ctx, sessionID)
	if !exists {
		return prompt
	}

	var enhanced strings.Builder

	enhanced.WriteString("## CURRENT TASK CONTEXT\n\n")
	enhanced.WriteString(fmt.Sprintf("**Task:** %s\n", task.Title))
	enhanced.WriteString(fmt.Sprintf("**Description:** %s\n", task.Description))

	if task.Context != "" {
		enhanced.WriteString(fmt.Sprintf("**Working Context:** %s\n", task.Context))
	}

	// Show current milestone
	if len(task.Milestones) > 0 {
		for _, milestone := range task.Milestones {
			if milestone.Status == StatusActive || milestone.Status == StatusBlocked {
				enhanced.WriteString(fmt.Sprintf("**Current Milestone:** %s - %s\n", milestone.Title, milestone.Description))
				break
			}
		}
	}

	enhanced.WriteString("\n**Reminder:** Stay focused on this task. Address the user request but maintain alignment with the current task objectives.\n\n")
	enhanced.WriteString("## USER REQUEST\n\n")
	enhanced.WriteString(prompt)

	return enhanced.String()
}

// Helper functions
func summarizePrompt(prompt string) string {
	prompt = strings.TrimSpace(prompt)

	// Find first sentence
	if idx := strings.Index(prompt, "."); idx > 0 && idx < 60 {
		return strings.TrimSpace(prompt[:idx])
	}

	// If short enough, use directly
	if len(prompt) < 60 {
		return prompt
	}

	// Truncate with ellipsis
	words := strings.Fields(prompt)
	var result []string

	for _, word := range words {
		if len(strings.Join(result, " ")) > 50 {
			break
		}
		result = append(result, word)
	}

	return strings.Join(result, " ") + "..."
}

func extractResponseContent(result interface{}) string {
	// Since we can't see the actual result structure,
	// this is a placeholder that needs to be implemented
	// based on the actual return type
	return ""
}

// TaskCommandFactory creates CLI commands for task management
type TaskCommandFactory struct {
	taskService Service
}

func NewTaskCommandFactory(service Service) *TaskCommandFactory {
	return &TaskCommandFactory{taskService: service}
}

// CreateTaskCommands returns CLI commands for task management
func (tcf *TaskCommandFactory) CreateTaskCommands(sessionID string) map[string]interface{} {
	tool := NewTaskTool(tcf.taskService, sessionID)
	return map[string]interface{}{
		"task": tool.GetTaskCommand(),
	}
}
