package state

import (
	"crypto/sha256"
	"fmt"
	"sync"
	"time"
)

// ProgressTracker tracks semantic progress to distinguish productive work from loops.
type ProgressTracker struct {
	mu sync.RWMutex

	// Semantic progress markers
	filesModified    map[string]string // file path -> content hash
	commandsExecuted map[string]int    // command -> execution count
	testsRun         []TestResult
	milestones       []Milestone

	// Loop detection
	recentActions    []ActionFingerprint
	maxRecentActions int // Keep last 10-20 actions

	// Error tracking
	recentErrors     []ErrorFingerprint
	maxRecentErrors  int
	consecutiveErrors int

	// Message deduplication
	recentMessageHashes []string
	maxMessageHistory   int
}

// ActionFingerprint uniquely identifies an action for loop detection.
type ActionFingerprint struct {
	Timestamp  time.Time
	ToolName   string
	TargetFile string // For file operations
	Command    string // For bash operations
	ErrorHash  string // Hash of error message if failed
	Success    bool
}

// ErrorFingerprint tracks errors for pattern detection.
type ErrorFingerprint struct {
	Timestamp time.Time
	ToolName  string
	Target    string // File or command
	ErrorMsg  string
	ErrorHash string
}

// TestResult tracks test execution outcomes.
type TestResult struct {
	Timestamp time.Time
	Command   string
	Passed    bool
	Output    string
}

// Milestone represents a significant progress point.
type Milestone struct {
	Timestamp   time.Time
	Description string
	Phase       int
	Metadata    map[string]interface{}
}

// NewProgressTracker creates a new progress tracker.
func NewProgressTracker() *ProgressTracker {
	return &ProgressTracker{
		filesModified:       make(map[string]string),
		commandsExecuted:    make(map[string]int),
		testsRun:            make([]TestResult, 0),
		milestones:          make([]Milestone, 0),
		recentActions:       make([]ActionFingerprint, 0),
		maxRecentActions:    20,
		recentErrors:        make([]ErrorFingerprint, 0),
		maxRecentErrors:     10,
		recentMessageHashes: make([]string, 0),
		maxMessageHistory:   5,
	}
}

// RecordFileModification records a file change.
func (pt *ProgressTracker) RecordFileModification(filePath, contentHash string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	pt.filesModified[filePath] = contentHash
}

// RecordCommand records a command execution.
func (pt *ProgressTracker) RecordCommand(command string, success bool) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	pt.commandsExecuted[command]++
}

// RecordTest records a test execution.
func (pt *ProgressTracker) RecordTest(command string, passed bool, output string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	pt.testsRun = append(pt.testsRun, TestResult{
		Timestamp: time.Now(),
		Command:   command,
		Passed:    passed,
		Output:    output,
	})
}

// RecordMilestone records a progress milestone.
func (pt *ProgressTracker) RecordMilestone(description string, phase int, metadata map[string]interface{}) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	pt.milestones = append(pt.milestones, Milestone{
		Timestamp:   time.Now(),
		Description: description,
		Phase:       phase,
		Metadata:    metadata,
	})
}

// RecordAction records an action for loop detection.
func (pt *ProgressTracker) RecordAction(toolName, targetFile, command, errorMsg string, success bool) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	errorHash := ""
	if errorMsg != "" {
		errorHash = hashString(errorMsg)
	}

	action := ActionFingerprint{
		Timestamp:  time.Now(),
		ToolName:   toolName,
		TargetFile: targetFile,
		Command:    command,
		ErrorHash:  errorHash,
		Success:    success,
	}

	pt.recentActions = append(pt.recentActions, action)
	if len(pt.recentActions) > pt.maxRecentActions {
		pt.recentActions = pt.recentActions[1:]
	}

	// Track errors separately
	if !success && errorMsg != "" {
		pt.consecutiveErrors++
		pt.recentErrors = append(pt.recentErrors, ErrorFingerprint{
			Timestamp: time.Now(),
			ToolName:  toolName,
			Target:    getTarget(targetFile, command),
			ErrorMsg:  errorMsg,
			ErrorHash: errorHash,
		})
		if len(pt.recentErrors) > pt.maxRecentErrors {
			pt.recentErrors = pt.recentErrors[1:]
		}
	} else if success {
		pt.consecutiveErrors = 0 // Reset on success
	}
}

// RecordMessage records a message hash for deduplication.
func (pt *ProgressTracker) RecordMessage(message string) bool {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	hash := hashString(message)

	// Check if duplicate
	for _, h := range pt.recentMessageHashes {
		if h == hash {
			return true // Duplicate detected
		}
	}

	pt.recentMessageHashes = append(pt.recentMessageHashes, hash)
	if len(pt.recentMessageHashes) > pt.maxMessageHistory {
		pt.recentMessageHashes = pt.recentMessageHashes[1:]
	}

	return false // Not a duplicate
}

// IsStuck determines if the agent is stuck in a loop.
func (pt *ProgressTracker) IsStuck() (bool, string) {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	// Check 1: Consecutive identical errors (same file, same error, 5+ times)
	// Increased threshold to reduce false positives
	if pt.consecutiveErrors >= 5 {
		if stuck, reason := pt.hasSameFileError(); stuck {
			return true, reason
		}
	}

	// Check 2: Oscillating actions (edit A->B, edit B->A repeatedly)
	if stuck, reason := pt.hasOscillatingActions(); stuck {
		return true, reason
	}

	// Check 3: No progress (same actions, no file changes)
	// Only check if we have enough actions and no recent success
	if len(pt.recentActions) >= 15 { // Increased from 10
		if stuck, reason := pt.hasNoProgress(); stuck {
			return true, reason
		}
	}

	return false, ""
}

// hasSameFileError checks if the same file has the same error 3+ times.
func (pt *ProgressTracker) hasSameFileError() (bool, string) {
	if len(pt.recentErrors) < 3 {
		return false, ""
	}

	// Get last 3 errors
	recent := pt.recentErrors[max(0, len(pt.recentErrors)-3):]

	// Check if same target and error
	if len(recent) >= 3 {
		first := recent[0]
		matchCount := 1
		for i := 1; i < len(recent); i++ {
			if recent[i].Target == first.Target && recent[i].ErrorHash == first.ErrorHash {
				matchCount++
			}
		}
		if matchCount >= 3 {
			return true, fmt.Sprintf("Same error on '%s' repeated 3 times", first.Target)
		}
	}

	return false, ""
}

// hasOscillatingActions detects alternating actions (A->B->A->B pattern).
func (pt *ProgressTracker) hasOscillatingActions() (bool, string) {
	if len(pt.recentActions) < 4 {
		return false, ""
	}

	// Get last 4 actions
	recent := pt.recentActions[max(0, len(pt.recentActions)-4):]

	// Check for A-B-A-B pattern (same target, alternating operations)
	if len(recent) == 4 {
		if recent[0].TargetFile == recent[2].TargetFile &&
			recent[1].TargetFile == recent[3].TargetFile &&
			recent[0].TargetFile != "" &&
			recent[1].TargetFile != "" &&
			recent[0].TargetFile != recent[1].TargetFile {
			return true, fmt.Sprintf("Oscillating between '%s' and '%s'", recent[0].TargetFile, recent[1].TargetFile)
		}
	}

	return false, ""
}

// hasNoProgress checks if there's been no meaningful progress.
func (pt *ProgressTracker) hasNoProgress() (bool, string) {
	// Only check if we have enough actions
	if len(pt.recentActions) < 15 {
		return false, ""
	}

	// Get last 15 actions to match the error message
	recent := pt.recentActions[max(0, len(pt.recentActions)-15):]

	// Count successful operations and analyze progress patterns
	uniqueTargets := make(map[string]bool)
	successCount := 0
	meaningfulSuccessCount := 0 // Only counts edits, writes, etc. (not just views)
	successfulOps := make([]ActionFingerprint, 0)
	
	for _, action := range recent {
		if action.Success {
			target := getTarget(action.TargetFile, action.Command)
			if target != "" {
				uniqueTargets[target] = true
			}
			successCount++
			
			// Count only meaningful operations as progress (not just viewing)
			if action.ToolName != "view" && action.ToolName != "ls" && action.ToolName != "grep" {
				meaningfulSuccessCount++
			}
			
			successfulOps = append(successfulOps, action)
		}
	}

	// More sophisticated progress detection:
	// Only declare stuck if we have very few meaningful successes AND very little variety
	// AND no meaningful file modifications in recent actions
	
	// Check if we've made any actual file modifications (strong progress indicator)
	hasFileMods := len(pt.filesModified) > 0
	
	// Check for variety in successful operations (different tools, targets, etc.)
	uniqueTools := make(map[string]bool)
	for _, op := range successfulOps {
		uniqueTools[op.ToolName] = true
	}
	
	// More lenient conditions: need both low meaningful success rate AND low variety
	if meaningfulSuccessCount < 2 && len(uniqueTargets) < 3 && len(uniqueTools) < 2 && !hasFileMods {
		return true, fmt.Sprintf("No meaningful progress in last 15 actions (%d meaningful successes, %d unique targets)", meaningfulSuccessCount, len(uniqueTargets))
	}

	return false, ""
}

// Reset clears all progress tracking (called when phase transitions).
func (pt *ProgressTracker) Reset() {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	// Keep milestones and file modifications (historical progress)
	// But reset error tracking and action history
	pt.recentActions = make([]ActionFingerprint, 0)
	pt.recentErrors = make([]ErrorFingerprint, 0)
	pt.consecutiveErrors = 0
	pt.recentMessageHashes = make([]string, 0)
}

// GetStats returns current progress statistics.
func (pt *ProgressTracker) GetStats() ProgressStats {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	successCount := 0
	failureCount := 0
	for _, action := range pt.recentActions {
		if action.Success {
			successCount++
		} else {
			failureCount++
		}
	}

	return ProgressStats{
		FilesModified:     len(pt.filesModified),
		CommandsExecuted:  len(pt.commandsExecuted),
		TestsRun:          len(pt.testsRun),
		MilestonesReached: len(pt.milestones),
		RecentSuccesses:   successCount,
		RecentFailures:    failureCount,
		ConsecutiveErrors: pt.consecutiveErrors,
	}
}

// ProgressStats summarizes current progress.
type ProgressStats struct {
	FilesModified     int
	CommandsExecuted  int
	TestsRun          int
	MilestonesReached int
	RecentSuccesses   int
	RecentFailures    int
	ConsecutiveErrors int
}

// Helper functions

func hashString(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return fmt.Sprintf("%x", h.Sum(nil))[:16]
}

func getTarget(file, command string) string {
	if file != "" {
		return file
	}
	if command != "" {
		// Take first 50 chars of command as target
		if len(command) > 50 {
			return command[:50]
		}
		return command
	}
	return ""
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
