package state

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// StateMachine manages agent execution state with progress tracking.
type StateMachine struct {
	mu sync.RWMutex

	// Core state
	currentState AgentState
	sessionID    string
	startTime    time.Time

	// Progress tracking
	progressTracker *ProgressTracker
	phaseContext    *PhaseContext

	// Context management
	ctx        context.Context
	cancelFunc context.CancelFunc

	// Metrics
	toolCallCount int
	stateHistory  []StateMetrics
	maxHistory    int

	// Callbacks
	onStateChange func(from, to AgentState)
	onStuck       func(reason string)
	onProgress    func(stats ProgressStats)
}

// Config for state machine initialization.
type Config struct {
	SessionID     string
	Context       context.Context
	TotalPhases   int
	MaxHistory    int
	OnStateChange func(from, to AgentState)
	OnStuck       func(reason string)
	OnProgress    func(stats ProgressStats)
}

// NewStateMachine creates a new state machine.
func NewStateMachine(cfg Config) *StateMachine {
	ctx := cfg.Context
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithCancel(ctx)

	maxHistory := cfg.MaxHistory
	if maxHistory == 0 {
		maxHistory = 100
	}

	var phaseContext *PhaseContext
	if cfg.TotalPhases > 0 {
		phaseContext = NewPhaseContext(cfg.TotalPhases)
	}

	return &StateMachine{
		currentState:    StateIdle,
		sessionID:       cfg.SessionID,
		startTime:       time.Now(),
		progressTracker: NewProgressTracker(),
		phaseContext:    phaseContext,
		ctx:             ctx,
		cancelFunc:      cancel,
		toolCallCount:   0,
		stateHistory:    make([]StateMetrics, 0, maxHistory),
		maxHistory:      maxHistory,
		onStateChange:   cfg.OnStateChange,
		onStuck:         cfg.OnStuck,
		onProgress:      cfg.OnProgress,
	}
}

// TransitionTo attempts to transition to a new state.
func (sm *StateMachine) TransitionTo(newState AgentState) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	oldState := sm.currentState

	// Check if transition is valid
	if !oldState.CanTransitionTo(newState) {
		return &TransitionError{
			From:   oldState,
			To:     newState,
			Reason: "invalid transition",
		}
	}

	// Record state exit metrics
	if len(sm.stateHistory) > 0 {
		lastMetric := &sm.stateHistory[len(sm.stateHistory)-1]
		lastMetric.ExitTime = time.Now()
		lastMetric.Duration = lastMetric.ExitTime.Sub(lastMetric.EnterTime)
		lastMetric.TransitionTo = newState
	}

	// Transition
	sm.currentState = newState

	// Record state entry
	metric := StateMetrics{
		State:     newState,
		EnterTime: time.Now(),
	}
	sm.stateHistory = append(sm.stateHistory, metric)

	// Trim history if needed
	if len(sm.stateHistory) > sm.maxHistory {
		sm.stateHistory = sm.stateHistory[1:]
	}

	// Log transition
	slog.Info("state transition",
		"session_id", sm.sessionID,
		"from", oldState.String(),
		"to", newState.String(),
		"tool_calls", sm.toolCallCount,
	)

	// Callback
	if sm.onStateChange != nil {
		go sm.onStateChange(oldState, newState)
	}

	return nil
}

// GetState returns the current state.
func (sm *StateMachine) GetState() AgentState {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.currentState
}

// RecordToolCall records a tool execution.
func (sm *StateMachine) RecordToolCall(toolName, targetFile, command, errorMsg string, success bool) {
	sm.mu.Lock()
	sm.toolCallCount++
	sm.mu.Unlock()

	// Record in progress tracker
	sm.progressTracker.RecordAction(toolName, targetFile, command, errorMsg, success)

	// Record in phase context if applicable
	if sm.phaseContext != nil && success && targetFile != "" {
		sm.phaseContext.RecordFileChange(targetFile)
	}

	// Check for stuck condition
	if stuck, reason := sm.progressTracker.IsStuck(); stuck {
		slog.Warn("stuck condition detected",
			"session_id", sm.sessionID,
			"reason", reason,
			"tool_calls", sm.toolCallCount,
		)

		if sm.onStuck != nil {
			go sm.onStuck(reason)
		}
	}

	// Progress callback
	if sm.onProgress != nil && sm.toolCallCount%10 == 0 {
		stats := sm.progressTracker.GetStats()
		go sm.onProgress(stats)
	}
}

// RecordMessage records a message for deduplication.
func (sm *StateMachine) RecordMessage(message string) bool {
	return sm.progressTracker.RecordMessage(message)
}

// RecordFileModification records a file change with content hash.
func (sm *StateMachine) RecordFileModification(filePath, contentHash string) {
	sm.progressTracker.RecordFileModification(filePath, contentHash)

	if sm.phaseContext != nil {
		sm.phaseContext.RecordFileChange(filePath)
	}
}

// RecordTest records a test execution.
func (sm *StateMachine) RecordTest(command string, passed bool, output string) {
	sm.progressTracker.RecordTest(command, passed, output)

	if sm.phaseContext != nil {
		sm.phaseContext.MarkTestsPassed(passed)
	}
}

// RecordMilestone records a progress milestone.
func (sm *StateMachine) RecordMilestone(description string, metadata map[string]interface{}) {
	phase := 0
	if sm.phaseContext != nil {
		phase = sm.phaseContext.CurrentPhase
	}

	sm.progressTracker.RecordMilestone(description, phase, metadata)
}

// StartPhase begins a new phase (resets progress tracking).
func (sm *StateMachine) StartPhase(phaseNumber int, description string, expectedDuration time.Duration) error {
	if sm.phaseContext == nil {
		return fmt.Errorf("phase context not initialized")
	}

	// Complete previous phase if transitioning from one phase to another
	// (but not on first phase)
	if phaseNumber > 1 && sm.phaseContext.CurrentPhase > 0 {
		sm.phaseContext.CompletePhase(true)
	}

	// Start new phase
	sm.phaseContext.StartPhase(phaseNumber, description, expectedDuration)

	// Reset progress tracker (keep history, reset error tracking)
	sm.progressTracker.Reset()

	slog.Info("phase started",
		"session_id", sm.sessionID,
		"phase", phaseNumber,
		"description", description,
		"expected_duration", expectedDuration,
	)

	// Transition to phase transition state
	return sm.TransitionTo(StatePhaseTransition)
}

// CompletePhase marks the current phase as complete.
func (sm *StateMachine) CompletePhase(success bool) error {
	if sm.phaseContext == nil {
		return fmt.Errorf("phase context not initialized")
	}

	sm.phaseContext.CompletePhase(success)

	phaseInfo := sm.phaseContext.GetCurrentPhaseInfo()
	slog.Info("phase completed",
		"session_id", sm.sessionID,
		"phase", phaseInfo.PhaseNumber,
		"success", success,
		"duration", phaseInfo.Elapsed,
		"files_changed", phaseInfo.FilesChanged,
		"tests_passed", phaseInfo.TestsPassed,
	)

	return nil
}

// IsStuck checks if the agent is stuck in a loop.
func (sm *StateMachine) IsStuck() (bool, string) {
	return sm.progressTracker.IsStuck()
}

// GetProgress returns current progress statistics.
func (sm *StateMachine) GetProgress() ProgressStats {
	return sm.progressTracker.GetStats()
}

// GetPhaseInfo returns current phase information.
func (sm *StateMachine) GetPhaseInfo() *PhaseInfo {
	if sm.phaseContext == nil {
		return nil
	}

	info := sm.phaseContext.GetCurrentPhaseInfo()
	return &info
}

// GetTotalProgress returns overall progress across all phases.
func (sm *StateMachine) GetTotalProgress() *TotalProgress {
	if sm.phaseContext == nil {
		return nil
	}

	progress := sm.phaseContext.GetTotalProgress()
	return &progress
}

// GetToolCallCount returns the total number of tool calls.
func (sm *StateMachine) GetToolCallCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.toolCallCount
}

// GetElapsedTime returns time since state machine started.
func (sm *StateMachine) GetElapsedTime() time.Duration {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return time.Since(sm.startTime)
}

// Cancel cancels the state machine context.
func (sm *StateMachine) Cancel() {
	if sm.cancelFunc != nil {
		sm.cancelFunc()
	}

	_ = sm.TransitionTo(StateHalted)
}

// Context returns the state machine context.
func (sm *StateMachine) Context() context.Context {
	return sm.ctx
}

// GetStateHistory returns recent state transitions.
func (sm *StateMachine) GetStateHistory(limit int) []StateMetrics {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if limit <= 0 || limit > len(sm.stateHistory) {
		limit = len(sm.stateHistory)
	}

	start := len(sm.stateHistory) - limit
	history := make([]StateMetrics, limit)
	copy(history, sm.stateHistory[start:])

	return history
}
