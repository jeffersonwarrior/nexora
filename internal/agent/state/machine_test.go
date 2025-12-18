package state

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestStateTransitions(t *testing.T) {
	tests := []struct {
		name      string
		from      AgentState
		to        AgentState
		wantValid bool
	}{
		{"Idle to Processing", StateIdle, StateProcessingPrompt, true},
		{"Processing to Streaming", StateProcessingPrompt, StateStreamingResponse, true},
		{"Streaming to Executing", StateStreamingResponse, StateExecutingTool, true},
		{"Executing to Streaming", StateExecutingTool, StateStreamingResponse, true},
		{"Streaming to ProgressCheck", StateStreamingResponse, StateProgressCheck, true},
		{"ProgressCheck to PhaseTransition", StateProgressCheck, StatePhaseTransition, true},
		{"PhaseTransition to Processing", StatePhaseTransition, StateProcessingPrompt, true},
		
		// Invalid transitions
		{"Idle to Executing", StateIdle, StateExecutingTool, false},
		{"Halted to anything", StateHalted, StateProcessingPrompt, false},
		{"Processing to PhaseTransition", StateProcessingPrompt, StatePhaseTransition, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewStateMachine(Config{
				SessionID: "test",
				Context:   context.Background(),
			})

			// Set initial state (bypass validation for test setup)
			sm.mu.Lock()
			sm.currentState = tt.from
			sm.mu.Unlock()

			err := sm.TransitionTo(tt.to)
			if tt.wantValid && err != nil {
				t.Errorf("expected valid transition, got error: %v", err)
			}
			if !tt.wantValid && err == nil {
				t.Errorf("expected invalid transition, got no error")
			}
		})
	}
}

func TestProgressTrackerLoop(t *testing.T) {
	pt := NewProgressTracker()

	// Simulate same file, same error, 3 times
	for i := 0; i < 3; i++ {
		pt.RecordAction("edit", "test.go", "", "old_string not found", false)
	}

	stuck, reason := pt.IsStuck()
	if !stuck {
		t.Error("expected stuck condition after 3 identical errors")
	}
	if reason == "" {
		t.Error("expected reason for stuck condition")
	}
	t.Logf("Stuck detected: %s", reason)
}

func TestProgressTrackerNoLoopOnProgress(t *testing.T) {
	pt := NewProgressTracker()

	// Simulate productive work: different files, successes
	files := []string{"a.go", "b.go", "c.go", "d.go", "e.go"}
	for i := 0; i < 100; i++ {
		file := files[i%len(files)]
		pt.RecordAction("edit", file, "", "", true)
	}

	stuck, _ := pt.IsStuck()
	if stuck {
		t.Error("expected no stuck condition with productive work")
	}
}

func TestProgressTrackerOscillation(t *testing.T) {
	pt := NewProgressTracker()

	// Simulate A->B->A->B pattern
	for i := 0; i < 4; i++ {
		if i%2 == 0 {
			pt.RecordAction("edit", "a.go", "", "", false)
		} else {
			pt.RecordAction("edit", "b.go", "", "", false)
		}
	}

	stuck, reason := pt.IsStuck()
	if !stuck {
		t.Error("expected stuck condition for oscillation")
	}
	t.Logf("Oscillation detected: %s", reason)
}

func TestMessageDeduplication(t *testing.T) {
	pt := NewProgressTracker()

	msg := "same message repeated"

	// First time - not a duplicate
	dup := pt.RecordMessage(msg)
	if dup {
		t.Error("first message should not be duplicate")
	}

	// Second time - duplicate
	dup = pt.RecordMessage(msg)
	if !dup {
		t.Error("second identical message should be duplicate")
	}

	// Different message - not duplicate
	dup = pt.RecordMessage("different message")
	if dup {
		t.Error("different message should not be duplicate")
	}
}

func TestProgressReset(t *testing.T) {
	pt := NewProgressTracker()

	// Create error history
	for i := 0; i < 3; i++ {
		pt.RecordAction("edit", "test.go", "", "error", false)
	}

	// Should be stuck
	stuck, _ := pt.IsStuck()
	if !stuck {
		t.Error("expected stuck before reset")
	}

	// Reset
	pt.Reset()

	// Should not be stuck after reset
	stuck, _ = pt.IsStuck()
	if stuck {
		t.Error("expected not stuck after reset")
	}

	// File modifications should be preserved
	pt.RecordFileModification("test.go", "hash123")
	pt.Reset()
	if len(pt.filesModified) == 0 {
		t.Error("file modifications should be preserved after reset")
	}
}

func TestPhaseContext(t *testing.T) {
	pc := NewPhaseContext(10)

	// Start phase 1
	pc.StartPhase(1, "Refactor package A", 15*time.Minute)

	info := pc.GetCurrentPhaseInfo()
	if info.PhaseNumber != 1 {
		t.Errorf("expected phase 1, got %d", info.PhaseNumber)
	}
	if info.Description != "Refactor package A" {
		t.Errorf("unexpected description: %s", info.Description)
	}

	// Record some progress
	pc.RecordFileChange("a.go")
	pc.RecordFileChange("b.go")
	pc.MarkTestsPassed(true)

	info = pc.GetCurrentPhaseInfo()
	if info.FilesChanged != 2 {
		t.Errorf("expected 2 files changed, got %d", info.FilesChanged)
	}
	if !info.TestsPassed {
		t.Error("expected tests passed")
	}

	// Complete phase
	pc.CompletePhase(true)

	progress := pc.GetTotalProgress()
	if progress.CompletedPhases != 1 {
		t.Errorf("expected 1 completed phase, got %d", progress.CompletedPhases)
	}
	if progress.SuccessfulPhases != 1 {
		t.Errorf("expected 1 successful phase, got %d", progress.SuccessfulPhases)
	}
}

func TestStateMachineWithPhases(t *testing.T) {
	sm := NewStateMachine(Config{
		SessionID:   "test-session",
		Context:     context.Background(),
		TotalPhases: 3,
	})

	// Start phase 1
	err := sm.StartPhase(1, "Phase 1", 10*time.Minute)
	if err != nil {
		t.Fatalf("failed to start phase: %v", err)
	}

	if sm.GetState() != StatePhaseTransition {
		t.Errorf("expected StatePhaseTransition, got %s", sm.GetState())
	}

	// Record some work
	sm.RecordToolCall("edit", "test.go", "", "", true)
	sm.RecordToolCall("bash", "", "go test", "", true)
	sm.RecordTest("go test", true, "PASS")

	// Don't manually complete - StartPhase for phase 2 will auto-complete phase 1

	// Start phase 2 (progress should reset)
	err = sm.StartPhase(2, "Phase 2", 10*time.Minute)
	if err != nil {
		t.Fatalf("failed to start phase 2: %v", err)
	}

	stats := sm.GetProgress()
	if stats.ConsecutiveErrors != 0 {
		t.Errorf("expected errors reset after phase transition, got %d", stats.ConsecutiveErrors)
	}

	progress := sm.GetTotalProgress()
	if progress.CompletedPhases != 1 {
		t.Errorf("expected 1 completed phase, got %d", progress.CompletedPhases)
	}
	if progress.CurrentPhase != 2 {
		t.Errorf("expected current phase 2, got %d", progress.CurrentPhase)
	}
}

func TestStateMachineCallbacks(t *testing.T) {
	var (
		stateChanges   int
		stuckCalled    bool
		progressCalled bool
		mu             sync.Mutex
	)

	sm := NewStateMachine(Config{
		SessionID: "test",
		Context:   context.Background(),
		OnStateChange: func(from, to AgentState) {
			mu.Lock()
			stateChanges++
			mu.Unlock()
		},
		OnStuck: func(reason string) {
			mu.Lock()
			stuckCalled = true
			mu.Unlock()
		},
		OnProgress: func(stats ProgressStats) {
			mu.Lock()
			progressCalled = true
			mu.Unlock()
		},
	})

	// Trigger state change
	err := sm.TransitionTo(StateProcessingPrompt)
	if err != nil {
		t.Fatalf("transition failed: %v", err)
	}

	// Give callback time to execute
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	sc := stateChanges
	mu.Unlock()

	if sc == 0 {
		t.Error("expected state change callback to be called")
	}

	// Trigger stuck condition
	for i := 0; i < 3; i++ {
		sm.RecordToolCall("edit", "test.go", "", "same error", false)
	}

	// Give callback time to execute
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	stuck := stuckCalled
	mu.Unlock()

	if !stuck {
		t.Error("expected stuck callback to be called")
	}

	// Trigger progress callback (every 10 calls)
	for i := 0; i < 10; i++ {
		sm.RecordToolCall("edit", "file.go", "", "", true)
	}

	// Give callback time to execute
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	prog := progressCalled
	mu.Unlock()

	if !prog {
		t.Error("expected progress callback to be called")
	}
}

func TestStateMachineLongRunningTask(t *testing.T) {
	sm := NewStateMachine(Config{
		SessionID:   "long-task",
		Context:     context.Background(),
		TotalPhases: 10,
	})

	// Simulate 10-phase refactor with 1000 tool calls
	for phase := 1; phase <= 10; phase++ {
		err := sm.StartPhase(phase, fmt.Sprintf("Phase %d", phase), 15*time.Minute)
		if err != nil {
			t.Fatalf("failed to start phase %d: %v", phase, err)
		}

		// Simulate 100 tool calls per phase (1000 total)
		for call := 0; call < 100; call++ {
			fileName := fmt.Sprintf("file_%d_%d.go", phase, call)
			sm.RecordToolCall("edit", fileName, "", "", true)
		}

		// Mark tests passed
		sm.RecordTest("go test ./...", true, "PASS")

		// Note: Don't manually call CompletePhase - it will be auto-called
		// when starting the next phase (or can be called at the very end)

		// Should not be stuck (making progress)
		stuck, reason := sm.IsStuck()
		if stuck {
			t.Errorf("phase %d: unexpected stuck condition: %s", phase, reason)
		}
	}

	// Manually complete the last phase
	err := sm.CompletePhase(true)
	if err != nil {
		t.Fatalf("failed to complete final phase: %v", err)
	}

	// Verify final state
	progress := sm.GetTotalProgress()
	if progress.CompletedPhases != 10 {
		t.Errorf("expected 10 completed phases, got %d", progress.CompletedPhases)
	}
	if progress.SuccessfulPhases != 10 {
		t.Errorf("expected 10 successful phases, got %d", progress.SuccessfulPhases)
	}

	toolCalls := sm.GetToolCallCount()
	if toolCalls != 1000 {
		t.Errorf("expected 1000 tool calls, got %d", toolCalls)
	}

	t.Logf("✅ Successfully completed 10-phase refactor with %d tool calls in %v",
		toolCalls, sm.GetElapsedTime())
}

func TestStateMachineStuckDetection(t *testing.T) {
	sm := NewStateMachine(Config{
		SessionID: "stuck-test",
		Context:   context.Background(),
	})

	// Simulate stuck loop: same error 3 times
	for i := 0; i < 3; i++ {
		sm.RecordToolCall("edit", "stuck.go", "", "old_string not found", false)
	}

	stuck, reason := sm.IsStuck()
	if !stuck {
		t.Error("expected stuck condition after 3 identical errors")
	}
	t.Logf("✅ Stuck detected: %s", reason)

	// Verify we can continue after recovery
	sm.progressTracker.Reset()
	stuck, _ = sm.IsStuck()
	if stuck {
		t.Error("expected not stuck after reset")
	}
}
