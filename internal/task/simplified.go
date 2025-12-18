package task

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// SimpleTask represents a focused task without circular dependencies
type SimpleTask struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	SessionID   string    `json:"session_id"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated"`
}

// SimpleService provides the essential task management functionality
type SimpleService interface {
	CreateTask(ctx context.Context, sessionID, title, description string) *SimpleTask
	GetActiveTask(ctx context.Context, sessionID string) (*SimpleTask, bool)
	UpdateTask(ctx context.Context, sessionID, description string) *SimpleTask
	CompleteTask(ctx context.Context, sessionID string) bool
	CheckFocus(ctx context.Context, sessionID, response string) (bool, string)
	GetFocusPrompt(ctx context.Context, sessionID string) string
}

// simpleService implements SimpleService
type simpleService struct {
	tasks map[string]*SimpleTask // sessionID -> task
}

// NewSimpleService creates a new task service
func NewSimpleService() SimpleService {
	return &simpleService{
		tasks: make(map[string]*SimpleTask),
	}
}

// CreateTask creates a new focused task
func (s *simpleService) CreateTask(ctx context.Context, sessionID, title, description string) *SimpleTask {
	task := &SimpleTask{
		ID:          fmt.Sprintf("task-%d", time.Now().UnixNano()),
		Title:       title,
		Description: description,
		Status:      "active",
		SessionID:   sessionID,
		Created:     time.Now(),
		Updated:     time.Now(),
	}

	s.tasks[sessionID] = task
	return task
}

// GetActiveTask retrieves the current task for a session
func (s *simpleService) GetActiveTask(ctx context.Context, sessionID string) (*SimpleTask, bool) {
	task, exists := s.tasks[sessionID]
	return task, exists
}

// UpdateTask updates the task description (working context)
func (s *simpleService) UpdateTask(ctx context.Context, sessionID, description string) *SimpleTask {
	task, exists := s.tasks[sessionID]
	if !exists {
		return nil
	}

	task.Description = description
	task.Updated = time.Now()
	return task
}

// CompleteTask marks tasks as completed
func (s *simpleService) CompleteTask(ctx context.Context, sessionID string) bool {
	task, exists := s.tasks[sessionID]
	if exists {
		task.Status = "completed"
		task.Updated = time.Now()
		delete(s.tasks, sessionID) // Remove from active tracking
		return true
	}
	return false
}

// CheckFocus analyzes if the AI is drifting from its current task
func (s *simpleService) CheckFocus(ctx context.Context, sessionID, response string) (bool, string) {
	task, exists := s.tasks[sessionID]
	if !exists {
		return true, "No active task - all work accepted"
	}

	response = strings.ToLower(response)
	taskContext := strings.ToLower(task.Title + " " + task.Description)

	// If no task context, assume focused
	if taskContext == "" {
		return true, "No specific task context"
	}

	// Check for task relevance
	responseWords := strings.Fields(response)
	taskWords := strings.Fields(taskContext)

	var overlap int
	for _, rword := range responseWords {
		for _, tword := range taskWords {
			if len(rword) > 2 && len(tword) > 2 {
				if strings.Contains(tword, rword) || strings.Contains(rword, tword) {
					overlap++
					break
				}
			}
			if rword == tword {
				overlap++
				break
			}
		}
	}

	// Check for clear drift indicators
	driftIndicators := []string{
		"weather", "news", "sports", "celebrity", "random", "joke",
		"fun fact", "trivia", "entertainment", "music", "movie",
		"could also", "while we're at it", "by the way", "also",
	}

	for _, indicator := range driftIndicators {
		if strings.Contains(response, indicator) {
			return false, "Detected potential drift: " + indicator
		}
	}

	// Require at least 10% relevance to stay on task
	if len(responseWords) > 0 {
		relevance := float64(overlap) / float64(len(responseWords))
		if relevance < 0.1 {
			return false, "Low relevance to current task"
		}
	}

	return true, "Maintaining focus"
}

// GetFocusPrompt returns a correction prompt if the agent is drifting
func (s *simpleService) GetFocusPrompt(ctx context.Context, sessionID string) string {
	task, exists := s.tasks[sessionID]
	if !exists {
		return ""
	}

	return fmt.Sprintf(`## TASK FOCUS REQUIRED

You appear to be drifting from your current task. Please refocus:

**Current Task:** %s
**Task Description:** %s

Stay focused on this task. Ignore unrelated requests or distractions.
Proceed with the focused approach to complete this specific task.

---
`, task.Title, task.Description)
}
