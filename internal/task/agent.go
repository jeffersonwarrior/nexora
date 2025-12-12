package task

import (
	"context"
	"fmt"
	"strings"
)

// AgentTaskMiddleware provides task-aware capabilities for agents
type AgentTaskMiddleware struct {
	taskService    Service
	agentSessionID string
}

// NewAgentTaskMiddleware creates task middleware for a specific agent session
func NewAgentTaskMiddleware(taskService Service, sessionID string) *AgentTaskMiddleware {
	return &AgentTaskMiddleware{
		taskService:    taskService,
		agentSessionID: sessionID,
	}
}

// StartFocusedTask begins a new focused task for the agent
func (atm *AgentTaskMiddleware) StartFocusedTask(ctx context.Context, title, description string, milestones []string) (*Task, error) {
	if len(milestones) == 0 {
		// Default milestone if none provided
		milestones = []string{"Complete the requested task"}
	}

	task := atm.taskService.CreateTask(ctx, atm.agentSessionID, title, description, "Working on the current user request", milestones)

	// Log task start
	fmt.Printf("🎯 Task Started: %s\n", title)
	fmt.Printf("📍 Context: %s\n", description)

	return task, nil
}

// UpdateTaskContext dynamically updates the current working context
func (atm *AgentTaskMiddleware) UpdateTaskContext(ctx context.Context, newContext string) {
	atm.taskService.UpdateTaskContext(ctx, atm.agentSessionID, newContext)
}

// CheckAndCorrectDrift analyzes responses and applies corrections if needed
func (atm *AgentTaskMiddleware) CheckAndCorrectDrift(ctx context.Context, response string) (string, bool) {
	analysis := atm.taskService.AnalyzeDrift(ctx, atm.agentSessionID, response)

	if analysis.Drifted {
		// Return correction prompt
		correction := atm.taskService.GetCorrectionPrompt(ctx, atm.agentSessionID)
		return fmt.Sprintf("[TASK CORRECTION NEEDED]\n%s\n\nContinue with your focused task.\n\n[ORIGINAL RESPONSE FOLLOWS]\n%s",
			correction, response), true
	}

	return response, false
}

// MarkProgress updates milestone based on agent work
func (atm *AgentTaskMiddleware) MarkProgress(ctx context.Context, milestoneDescription string) {
	task, exists := atm.taskService.GetActiveTask(ctx, atm.agentSessionID)
	if !exists {
		return
	}

	// Find matching milestone
	for _, milestone := range task.Milestones {
		if strings.Contains(strings.ToLower(milestone.Description), strings.ToLower(milestoneDescription)) {
			atm.taskService.MarkMilestoneProgress(ctx, atm.agentSessionID, milestone.ID, milestoneDescription)
			fmt.Printf("✅ Milestone Completed: %s\n", milestone.Title)
			break
		}
	}
}

// GetTaskSummary returns current task progress
func (atm *AgentTaskMiddleware) GetTaskSummary(ctx context.Context) TaskSummary {
	return atm.taskService.GetTaskSummary(ctx, atm.agentSessionID)
}

// CompleteCurrentTask finishes the current task
func (atm *AgentTaskMiddleware) CompleteCurrentTask(ctx context.Context) {
	summary := atm.taskService.GetTaskSummary(ctx, atm.agentSessionID)
	if summary.Title != "" {
		atm.taskService.CompleteTask(ctx, atm.agentSessionID)
		fmt.Printf("🎉 Task Completed: %s (Progress: %.0f%%)\n", summary.Title, summary.Progress*100)
	}
}

// AutoTaskCreator automatically creates tasks from user prompts
func (atm *AgentTaskMiddleware) AutoTaskCreator(ctx context.Context, userPrompt string) (*Task, error) {
	// Extract task info from prompt
	title := extractTaskTitle(userPrompt)
	description := userPrompt
	milestones := generateMilestones(userPrompt)

	return atm.StartFocusedTask(ctx, title, description, milestones)
}

// extractTaskTitle creates a concise title from user prompt
func extractTaskTitle(prompt string) string {
	prompt = strings.TrimSpace(prompt)

	// If prompt is short, use it directly
	if len(prompt) < 50 {
		return prompt
	}

	// Take first sentence or first ~40 chars
	if idx := strings.Index(prompt, "."); idx > 0 && idx < 40 {
		return strings.TrimSpace(prompt[:idx])
	}
	return prompt[:40] + "..."
}

// generateMilestones creates relevant milestones from the prompt
func generateMilestones(prompt string) []string {
	var milestones []string

	promptLower := strings.ToLower(prompt)

	// Common patterns and their milestones
	patterns := []struct {
		hints     []string
		milestone string
	}{
		{[]string{"add feature", "implement feature"}, "Add the new feature"},
		{[]string{"fix bug", "address issue", "resolve problem"}, "Fix the identified bug"},
		{[]string{"refactor", "improve", "optimize"}, "Refactor the existing code"},
		{[]string{"test", "write test"}, "Write comprehensive tests"},
		{[]string{"document", "add docs", "update documentation"}, "Update documentation"},
		{[]string{"remove", "delete", "clean up"}, "Remove unnecessary code"},
		{[]string{"update", "modify", "change"}, "Update the existing implementation"},
	}

	highestMatchScore := 0
	var bestMilestone string

	for _, pattern := range patterns {
		score := 0
		for _, hint := range pattern.hints {
			if strings.Contains(promptLower, hint) {
				score++
			}
		}
		if score > highestMatchScore {
			highestMatchScore = score
			bestMilestone = pattern.milestone
		}
	}

	if bestMilestone != "" {
		milestones = append(milestones, bestMilestone)
	}

	// Always add a completion milestone
	milestones = append(milestones, "Verify the implementation works correctly")
	milestones = append(milestones, "Format and commit the changes")

	return milestones
}
