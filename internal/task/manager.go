package task

import (
	"fmt"
	"strings"
	"time"
)

// Task represents a focused task with milestone tracking
type Task struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Context     string                 `json:"context"` // Current working context
	Milestones  []Milestone            `json:"milestones"`
	Status      TaskStatus             `json:"status"`
	Priority    Priority               `json:"priority"`
	SessionID   string                 `json:"session_id"`
	Created     time.Time              `json:"created"`
	Updated     time.Time              `json:"updated"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// Milestone represents a checkable progress point
type Milestone struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      TaskStatus `json:"status"`
	DueTime     time.Time  `json:"due_time,omitempty"`
	Evidence    string     `json:"evidence,omitempty"` // What was done to complete this
}

type TaskStatus string

const (
	StatusActive    TaskStatus = "active"
	StatusBlocked   TaskStatus = "blocked"
	StatusCompleted TaskStatus = "completed"
	StatusPaused    TaskStatus = "paused"
	StatusCancelled TaskStatus = "cancelled"
)

type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityMedium Priority = "medium"
	PriorityLow    Priority = "low"
)

// Manager tracks active tasks and detects drift
type Manager struct {
	activeTasks map[string]*Task // sessionID -> Task
	driftRules  []DriftRule
}

type DriftRule struct {
	Name     string
	Pattern  string
	Action   string
	Keywords []string
}

// NewManager creates a new task manager
func NewManager() *Manager {
	return &Manager{
		activeTasks: make(map[string]*Task),
		driftRules:  defaultDriftRules(),
	}
}

// CreateTask creates a new focused task
func (m *Manager) CreateTask(sessionID, title, description, context string, milestones []string) *Task {
	task := &Task{
		ID:          fmt.Sprintf("task-%d", time.Now().UnixNano()),
		Title:       title,
		Description: description,
		Context:     context,
		Status:      StatusActive,
		Priority:    PriorityHigh,
		SessionID:   sessionID,
		Created:     time.Now(),
		Updated:     time.Now(),
		Milestones:  make([]Milestone, 0),
	}

	// Create milestones from descriptions
	for i, milestoneDesc := range milestones {
		milestone := Milestone{
			ID:          fmt.Sprintf("milestone-%d-%d", time.Now().UnixNano(), i),
			Title:       fmt.Sprintf("Milestone %d", i+1),
			Description: milestoneDesc,
			Status:      StatusBlocked,
		}
		task.Milestones = append(task.Milestones, milestone)
	}

	m.activeTasks[sessionID] = task
	return task
}

// GetTask retrieves active task for a session
func (m *Manager) GetTask(sessionID string) (*Task, bool) {
	task, exists := m.activeTasks[sessionID]
	return task, exists
}

// UpdateTaskContext updates the current working context
func (m *Manager) UpdateTaskContext(sessionID, newContext string) {
	if task, exists := m.activeTasks[sessionID]; exists {
		task.Context = newContext
		task.Updated = time.Now()
	}
}

// AnalyzeDrift analyzes AI response for task drift
func (m *Manager) AnalyzeDrift(sessionID, aiResponse string) DriftAnalysis {
	task, exists := m.activeTasks[sessionID]
	if !exists {
		return DriftAnalysis{Drifted: false, Message: "No active task"}
	}

	analysis := DriftAnalysis{
		Drifted:         false,
		Confidence:      1.0,
		Recommendations: make([]string, 0),
		ActionNeeded:    false,
	}

	// Check for task relevance
	if !isRelevantToTask(aiResponse, task) {
		analysis.Drifted = true
		analysis.Confidence = 0.8
		analysis.Recommendations = append(analysis.Recommendations,
			fmt.Sprintf("Your response doesn't seem to relate to the current task: %s", task.Title))
	}

	// Check for drift rules
	for _, rule := range m.driftRules {
		if matchesDriftRule(aiResponse, rule) {
			analysis.Drifted = true
			analysis.Recommendations = append(analysis.Recommendations,
				fmt.Sprintf("Detected potential drift: %s. Please refocus on: %s", rule.Name, task.Title))
			analysis.ActionNeeded = true
			break
		}
	}

	// Check milestone completion hints
	if suggestsMilestoneCompletion(aiResponse, task) {
		analysis.MilestoneProgress = m.checkMilestoneProgress(aiResponse, task)
	}

	// Set appropriate message
	if analysis.Drifted {
		analysis.Message = "Drift detected - see recommendations"
	} else {
		analysis.Message = "Response is on track"
	}

	return analysis
}

// GetCorrectionPrompt returns a prompt to steer AI back on track
func (m *Manager) GetCorrectionPrompt(sessionID string) string {
	task, exists := m.activeTasks[sessionID]
	if !exists {
		return ""
	}

	prompt := fmt.Sprintf("## Task Focus Required\n\nYou appear to have drifted from your current task: **%s**\n\n", task.Title)
	prompt += fmt.Sprintf("**Context:** %s\n\n", task.Context)

	if len(task.Milestones) > 0 {
		prompt += "**Current Milestones:**\n"
		for _, milestone := range task.Milestones {
			if milestone.Status == StatusActive || milestone.Status == StatusBlocked {
				prompt += fmt.Sprintf("- %s: %s\n", milestone.Title, milestone.Description)
				break // Show next milestone to focus on
			}
		}
		prompt += "\n"
	}

	prompt += "Please refocus on this task and ignore unrelated requests or distractions."

	return prompt
}

// MarkMilestoneProgress updates milestone based on AI work
func (m *Manager) MarkMilestoneProgress(sessionID, milestoneID, evidence string) {
	if task, exists := m.activeTasks[sessionID]; exists {
		for i, milestone := range task.Milestones {
			if milestone.ID == milestoneID {
				task.Milestones[i].Status = StatusCompleted
				task.Milestones[i].Evidence = evidence
				task.Updated = time.Now()
				break
			}
		}
	}
}

// CompleteTask marks task as completed
func (m *Manager) CompleteTask(sessionID string) {
	if task, exists := m.activeTasks[sessionID]; exists {
		task.Status = StatusCompleted
		task.Updated = time.Now()
	}
}

// CloseTask removes task from active tracking
func (m *Manager) CloseTask(sessionID string) {
	delete(m.activeTasks, sessionID)
}

type DriftAnalysis struct {
	Drifted           bool                `json:"drifted"`
	Confidence        float64             `json:"confidence"`
	Recommendations   []string            `json:"recommendations"`
	MilestoneProgress []MilestoneProgress `json:"milestone_progress,omitempty"`
	ActionNeeded      bool                `json:"action_needed"`
	Message           string              `json:"message"`
}

type MilestoneProgress struct {
	MilestoneID string `json:"milestone_id"`
	Completed   bool   `json:"completed"`
	Evidence    string `json:"evidence"`
}

// Helper functions
func isRelevantToTask(response string, task *Task) bool {
	response = strings.ToLower(response)
	taskContext := strings.ToLower(task.Context + " " + task.Title + " " + task.Description)

	// Simple keyword overlap check
	responseWords := strings.Fields(response)
	taskWords := strings.Fields(taskContext)

	overlap := 0
	for _, word := range responseWords {
		for _, taskWord := range taskWords {
			if strings.Contains(taskWord, word) || strings.Contains(word, taskWord) {
				overlap++
				break
			}
		}
	}

	return float64(overlap)/float64(len(responseWords)) > 0.1 // At least 10% word overlap
}

func matchesDriftRule(response string, rule DriftRule) bool {
	response = strings.ToLower(response)

	// Check for rule keywords
	for _, keyword := range rule.Keywords {
		if strings.Contains(response, strings.ToLower(keyword)) {
			return true
		}
	}

	return false
}

func suggestsMilestoneCompletion(response string, task *Task) bool {
	completionWords := []string{"done", "completed", "finished", "implemented", "fixed", "added", "updated"}
	response = strings.ToLower(response)

	for _, word := range completionWords {
		if strings.Contains(response, word) {
			return true
		}
	}
	return false
}

func (m *Manager) checkMilestoneProgress(response string, task *Task) []MilestoneProgress {
	var progress []MilestoneProgress

	for _, milestone := range task.Milestones {
		if milestone.Status == StatusActive || milestone.Status == StatusBlocked {
			if strings.Contains(strings.ToLower(response), strings.ToLower(milestone.Title)) ||
				strings.Contains(strings.ToLower(response), strings.ToLower(milestone.Description)) {
				progress = append(progress, MilestoneProgress{
					MilestoneID: milestone.ID,
					Completed:   true,
					Evidence:    response,
				})
			}
		}
	}

	return progress
}

func defaultDriftRules() []DriftRule {
	return []DriftRule{
		{
			Name:     "Unrelated Topic",
			Pattern:  ".*",
			Action:   "refocus",
			Keywords: []string{"weather", "news", "sports", "celebrities", "politics", "random", "fun fact"},
		},
		{
			Name:     "Technical Distraction",
			Pattern:  ".*",
			Action:   "refocus",
			Keywords: []string{"latest iPhone", "new framework", "another database", "different approach", "alternative solution"},
		},
		{
			Name:     "Scope Creep",
			Pattern:  ".*",
			Action:   "refocus",
			Keywords: []string{"also", "additionally", "could also", "might want to", "while we're at it"},
		},
	}
}
