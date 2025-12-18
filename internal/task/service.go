package task

import (
	"context"
)

// Service defines the task management interface
type Service interface {
	// Task lifecycle management
	CreateTask(ctx context.Context, sessionID, title, description, context string, milestones []string) *Task
	GetActiveTask(ctx context.Context, sessionID string) (*Task, bool)
	UpdateTaskContext(ctx context.Context, sessionID, newContext string)
	CompleteTask(ctx context.Context, sessionID string)
	CloseTask(ctx context.Context, sessionID string)

	// Drift detection and correction
	AnalyzeDrift(ctx context.Context, sessionID, aiResponse string) DriftAnalysis
	GetCorrectionPrompt(ctx context.Context, sessionID string) string

	// Progress tracking
	MarkMilestoneProgress(ctx context.Context, sessionID, milestoneID, evidence string)
	GetTaskSummary(ctx context.Context, sessionID string) TaskSummary
}

// TaskSummary provides a concise overview of task state
type TaskSummary struct {
	TaskID        string        `json:"task_id"`
	Title         string        `json:"title"`
	Status        TaskStatus    `json:"status"`
	Progress      float64       `json:"progress"` // 0.0 to 1.0
	CurrentPhase  string        `json:"current_phase"`
	NextMilestone *Milestone    `json:"next_milestone,omitempty"`
	DriftRisk     DriftAnalysis `json:"drift_risk,omitempty"`
}

// Service implementation
type taskService struct {
	manager *Manager
}

// NewService creates a new task service
func NewService() Service {
	return &taskService{
		manager: NewManager(),
	}
}

// Task lifecycle methods
func (s *taskService) CreateTask(ctx context.Context, sessionID, title, description, context string, milestones []string) *Task {
	return s.manager.CreateTask(sessionID, title, description, context, milestones)
}

func (s *taskService) GetActiveTask(ctx context.Context, sessionID string) (*Task, bool) {
	return s.manager.GetTask(sessionID)
}

func (s *taskService) UpdateTaskContext(ctx context.Context, sessionID, newContext string) {
	s.manager.UpdateTaskContext(sessionID, newContext)
}

func (s *taskService) CompleteTask(ctx context.Context, sessionID string) {
	s.manager.CompleteTask(sessionID)
}

func (s *taskService) CloseTask(ctx context.Context, sessionID string) {
	s.manager.CloseTask(sessionID)
}

// Drift detection methods
func (s *taskService) AnalyzeDrift(ctx context.Context, sessionID, aiResponse string) DriftAnalysis {
	return s.manager.AnalyzeDrift(sessionID, aiResponse)
}

func (s *taskService) GetCorrectionPrompt(ctx context.Context, sessionID string) string {
	return s.manager.GetCorrectionPrompt(sessionID)
}

// Progress tracking methods
func (s *taskService) MarkMilestoneProgress(ctx context.Context, sessionID, milestoneID, evidence string) {
	s.manager.MarkMilestoneProgress(sessionID, milestoneID, evidence)
}

func (s *taskService) GetTaskSummary(ctx context.Context, sessionID string) TaskSummary {
	task, exists := s.manager.GetTask(sessionID)
	if !exists {
		return TaskSummary{}
	}

	var completedMilestones int
	var currentMilestone *Milestone

	for i, milestone := range task.Milestones {
		if milestone.Status == StatusCompleted {
			completedMilestones++
		} else if currentMilestone == nil {
			currentMilestone = &task.Milestones[i]
		}
	}

	progress := float64(completedMilestones) / float64(len(task.Milestones))

	return TaskSummary{
		TaskID:        task.ID,
		Title:         task.Title,
		Status:        task.Status,
		Progress:      progress,
		CurrentPhase:  task.Context,
		NextMilestone: currentMilestone,
		DriftRisk:     s.manager.AnalyzeDrift(sessionID, ""), // Check for potential drift without analyzing specific response
	}
}
