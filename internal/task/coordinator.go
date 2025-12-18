package task

import (
	"context"
	"fmt"
	"strings"

	"charm.land/fantasy"
)

// Coordinator defines the interface we need from coordinator (no import cycle)
type Coordinator interface {
	Run(ctx context.Context, sessionID, prompt string, attachments ...any) (*fantasy.AgentResult, error)
}

// CoordinatorMiddleware wraps the agent coordinator to add task management capabilities
type CoordinatorMiddleware struct {
	base                 Coordinator
	taskService          Service
	enableDriftDetection bool
}

// NewCoordinatorMiddleware creates a task-aware coordinator wrapper
func NewCoordinatorMiddleware(base Coordinator, taskService Service) *CoordinatorMiddleware {
	return &CoordinatorMiddleware{
		base:                 base,
		taskService:          taskService,
		enableDriftDetection: true,
	}
}

// Enhanced Run method with task context injection and drift monitoring
func (cm *CoordinatorMiddleware) Run(ctx context.Context, sessionID, prompt string, attachments ...any) (*fantasy.AgentResult, error) {
	// Add task context to the prompt if an active task exists
	if cm.enableDriftDetection {
		if task, exists := cm.taskService.GetActiveTask(ctx, sessionID); exists {
			enhancedPrompt := cm.injectTaskContext(prompt, task)

			// Run the enhanced prompt
			result, err := cm.base.Run(ctx, sessionID, enhancedPrompt, attachments...)
			if err != nil {
				return nil, err
			}

			// Get response content from result
			response := GetResultContent(result)

			// Analyze the result for task drift
			driftAnalysis := cm.taskService.AnalyzeDrift(ctx, sessionID, response)

			// If drift detected, generate correction response
			if driftAnalysis.Drifted && driftAnalysis.ActionNeeded {
				correctionPrompt := cm.taskService.GetCorrectionPrompt(ctx, sessionID)

				// Create a follow-up call to redirect the agent
				followUpResult, err := cm.base.Run(ctx, sessionID, correctionPrompt, attachments...)
				if err == nil {
					// Combine the responses
					correctionResponse := GetResultContent(followUpResult)
					return combineResults(result, correctionResponse), nil
				}
			}

			return result, nil
		}
	}

	// No active task or drift detection disabled - run normally
	return cm.base.Run(ctx, sessionID, prompt, attachments...)
}

// Helper functions for fantasy.AgentResult
func GetResultContent(result *fantasy.AgentResult) string {
	if result == nil {
		return ""
	}

	// Extract content from result - implementation depends on fantasy.AgentResult structure
	// This should extract the main response content from the agent result
	// Common patterns: .Content, .Message, .Response.Text, etc.
	return ""
}

func combineResults(original *fantasy.AgentResult, correction string) *fantasy.AgentResult {
	if original == nil {
		return original
	}

	// Combine results by appending correction to original response
	// Implementation depends on fantasy.AgentResult structure fields
	// Common patterns: append to .Message, create new .Content with both, etc.
	return original
}

// injectTaskContext enhances prompts with task information
func (cm *CoordinatorMiddleware) injectTaskContext(prompt string, task *Task) string {
	if task == nil || task.Context == "" {
		return prompt
	}

	contextualPrompt := fmt.Sprintf(`## TASK CONTEXT

**Current Task:** %s
**Task Context:** %s
**Task Description:** %s

`, task.Title, task.Context, task.Description)

	// Add current milestone if available
	if len(task.Milestones) > 0 {
		for _, milestone := range task.Milestones {
			if milestone.Status == StatusActive || milestone.Status == StatusBlocked {
				contextualPrompt += fmt.Sprintf(`**Current Milestone:** %s - %s

`, milestone.Title, milestone.Description)
				break
			}
		}
	}

	contextualPrompt += fmt.Sprintf(`**IMPORTANT:** Stay focused on this task. Ignore requests that divert from it unless explicitly told otherwise.

---

## USER REQUEST:
%s`, prompt)

	return contextualPrompt
}

// EnableDriftDetection turns on/off drift monitoring
func (cm *CoordinatorMiddleware) EnableDriftDetection(enabled bool) {
	cm.enableDriftDetection = enabled
}

// Task management convenience methods
func (cm *CoordinatorMiddleware) StartTask(ctx context.Context, sessionID, title, description string, milestones []string) *Task {
	return cm.taskService.CreateTask(ctx, sessionID, title, description, "", milestones)
}

func (cm *CoordinatorMiddleware) GetTaskStatus(ctx context.Context, sessionID string) TaskSummary {
	return cm.taskService.GetTaskSummary(ctx, sessionID)
}

func (cm *CoordinatorMiddleware) UpdateTaskProgress(ctx context.Context, sessionID, milestoneID, evidence string) {
	cm.taskService.MarkMilestoneProgress(ctx, sessionID, milestoneID, evidence)
}

// Helper to detect if a prompt is task-related vs unrelated
func IsTaskRelated(prompt string, task *Task) bool {
	if task == nil || task.Context == "" {
		return true // No task context means don't restrict
	}

	prompt = strings.ToLower(prompt)

	// Check for task-related keywords
	relatedKeywords := []string{
		"continue", "next", "continue with", "proceed", "implement", "fix", "add", "update",
		"modify", "change", "refactor", "improve", "optimize", "debug", "test", "review",
		"documentation", "read", "analyze", "examine", "check", "verify", "validate",
	}

	// Check for unrelated keywords that might indicate drift
	unrelatedKeywords := []string{
		"weather", "news", "sports", "entertainment", "politics", "random", "fun fact",
		"joke", "story", "movie", "music", "book", "travel", "food", "fashion",
		"social media", "celebrity", "gossip", "rumor", "latest", "new trend",
	}

	// If prompt contains unrelated keywords and no related ones, it might be drift
	for _, unrelated := range unrelatedKeywords {
		if strings.Contains(prompt, unrelated) {
			hasRelated := false
			for _, related := range relatedKeywords {
				if strings.Contains(prompt, related) {
					hasRelated = true
					break
				}
			}
			if !hasRelated {
				return false // Likely unrelated to current task
			}
		}
	}

	return true // Likely related or neutral
}
