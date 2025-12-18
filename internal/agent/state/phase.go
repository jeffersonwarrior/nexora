package state

import (
	"sync"
	"time"
)

// PhaseContext tracks multi-phase task execution (e.g., 10-phase refactor).
type PhaseContext struct {
	mu sync.RWMutex

	// Phase identification
	CurrentPhase     int
	TotalPhases      int
	PhaseDescription string

	// Timing
	PhaseStartTime   time.Time
	ExpectedDuration time.Duration
	TotalStartTime   time.Time

	// Progress indicators
	FilesChanged []string
	TestsPassed  bool
	Blockers     []string

	// Phase history
	CompletedPhases []CompletedPhase
}

// CompletedPhase represents a finished phase.
type CompletedPhase struct {
	PhaseNumber int
	Description string
	StartTime   time.Time
	EndTime     time.Time
	Duration    time.Duration
	FileCount   int
	TestsPassed bool
	Success     bool
}

// NewPhaseContext creates a new phase context.
func NewPhaseContext(totalPhases int) *PhaseContext {
	return &PhaseContext{
		CurrentPhase:    0,
		TotalPhases:     totalPhases,
		TotalStartTime:  time.Now(),
		FilesChanged:    make([]string, 0),
		CompletedPhases: make([]CompletedPhase, 0),
		Blockers:        make([]string, 0),
	}
}

// StartPhase begins a new phase.
func (pc *PhaseContext) StartPhase(phaseNumber int, description string, expectedDuration time.Duration) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	pc.CurrentPhase = phaseNumber
	pc.PhaseDescription = description
	pc.PhaseStartTime = time.Now()
	pc.ExpectedDuration = expectedDuration
	pc.FilesChanged = make([]string, 0)
	pc.TestsPassed = false
	pc.Blockers = make([]string, 0)
}

// CompletePhase marks the current phase as complete.
func (pc *PhaseContext) CompletePhase(success bool) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	completed := CompletedPhase{
		PhaseNumber: pc.CurrentPhase,
		Description: pc.PhaseDescription,
		StartTime:   pc.PhaseStartTime,
		EndTime:     time.Now(),
		Duration:    time.Since(pc.PhaseStartTime),
		FileCount:   len(pc.FilesChanged),
		TestsPassed: pc.TestsPassed,
		Success:     success,
	}

	pc.CompletedPhases = append(pc.CompletedPhases, completed)
}

// RecordFileChange adds a file to the current phase's changed files.
func (pc *PhaseContext) RecordFileChange(filePath string) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	// Avoid duplicates
	for _, existing := range pc.FilesChanged {
		if existing == filePath {
			return
		}
	}
	pc.FilesChanged = append(pc.FilesChanged, filePath)
}

// MarkTestsPassed marks tests as passing for the current phase.
func (pc *PhaseContext) MarkTestsPassed(passed bool) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	pc.TestsPassed = passed
}

// AddBlocker adds a blocker to the current phase.
func (pc *PhaseContext) AddBlocker(blocker string) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	pc.Blockers = append(pc.Blockers, blocker)
}

// GetCurrentPhaseInfo returns information about the current phase.
func (pc *PhaseContext) GetCurrentPhaseInfo() PhaseInfo {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	return PhaseInfo{
		PhaseNumber:      pc.CurrentPhase,
		TotalPhases:      pc.TotalPhases,
		Description:      pc.PhaseDescription,
		Elapsed:          time.Since(pc.PhaseStartTime),
		ExpectedDuration: pc.ExpectedDuration,
		FilesChanged:     len(pc.FilesChanged),
		TestsPassed:      pc.TestsPassed,
		HasBlockers:      len(pc.Blockers) > 0,
	}
}

// GetTotalProgress returns overall progress information.
func (pc *PhaseContext) GetTotalProgress() TotalProgress {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	completedCount := len(pc.CompletedPhases)
	successCount := 0
	totalFileChanges := 0

	for _, phase := range pc.CompletedPhases {
		if phase.Success {
			successCount++
		}
		totalFileChanges += phase.FileCount
	}

	return TotalProgress{
		TotalPhases:      pc.TotalPhases,
		CompletedPhases:  completedCount,
		SuccessfulPhases: successCount,
		CurrentPhase:     pc.CurrentPhase,
		TotalElapsed:     time.Since(pc.TotalStartTime),
		TotalFileChanges: totalFileChanges,
	}
}

// IsPhaseTimeout checks if the current phase has exceeded expected duration.
func (pc *PhaseContext) IsPhaseTimeout(multiplier float64) bool {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	if pc.ExpectedDuration == 0 {
		return false // No timeout set
	}

	elapsed := time.Since(pc.PhaseStartTime)
	threshold := time.Duration(float64(pc.ExpectedDuration) * multiplier)

	return elapsed > threshold
}

// PhaseInfo provides snapshot of current phase.
type PhaseInfo struct {
	PhaseNumber      int
	TotalPhases      int
	Description      string
	Elapsed          time.Duration
	ExpectedDuration time.Duration
	FilesChanged     int
	TestsPassed      bool
	HasBlockers      bool
}

// TotalProgress provides overview of all phases.
type TotalProgress struct {
	TotalPhases      int
	CompletedPhases  int
	SuccessfulPhases int
	CurrentPhase     int
	TotalElapsed     time.Duration
	TotalFileChanges int
}
