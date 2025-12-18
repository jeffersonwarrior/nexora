package recovery

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/nexora/nexora/internal/agent/state"
)

// Integration tests for the complete recovery system
func TestRecoverySystemIntegration(t *testing.T) {
	ctx := context.Background()

	// Test 1: File outdated recovery with actual file
	t.Run("FileOutdatedIntegration", func(t *testing.T) {
		// Create a test file
		testFile := "/tmp/test_integration.txt"
		initialContent := "initial content"
		err := os.WriteFile(testFile, []byte(initialContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		defer os.Remove(testFile)

		registry := NewRecoveryRegistry()
		execCtx := &state.AgentExecutionContext{
			RetryCount: 0,
			ErrorCount: 0,
			LastError:  nil,
		}

		// Simulate file outdated error
		fileErr := NewFileOutdatedError(errors.New("file was modified"), testFile)

		// Attempt recovery
		recoveryErr := registry.AttemptRecovery(ctx, fileErr, execCtx)
		if recoveryErr != nil {
			t.Errorf("Recovery should succeed: %v", recoveryErr)
		}

		if execCtx.RetryCount != 1 {
			t.Errorf("Expected retry count 1, got %d", execCtx.RetryCount)
		}

		if execCtx.LastError == nil {
			t.Error("Last error should be set after recovery")
		}
	})

	// Test 2: Edit failed recovery with file operations
	t.Run("EditFailedIntegration", func(t *testing.T) {
		testFile := "/tmp/test_edit.txt"
		content := "line 1\nline 2\nline 3\n"
		err := os.WriteFile(testFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		defer os.Remove(testFile)

		registry := NewRecoveryRegistry()
		execCtx := &state.AgentExecutionContext{
			RetryCount: 0,
			ErrorCount: 0,
			LastError:  nil,
		}

		// Simulate edit failed error
		editErr := NewEditFailedError(errors.New("whitespace mismatch"), testFile, "line 2", "line 2 updated")

		// Attempt recovery
		recoveryErr := registry.AttemptRecovery(ctx, editErr, execCtx)
		if recoveryErr != nil {
			// Recovery may fail but should provide useful error
			t.Logf("Recovery failed as expected: %v", recoveryErr)
		}
	})

	// Test 3: Loop detection should halt immediately
	t.Run("LoopDetectionIntegration", func(t *testing.T) {
		registry := NewRecoveryRegistry()
		execCtx := &state.AgentExecutionContext{
			RetryCount: 0,
			ErrorCount: 0,
			LastError:  nil,
		}

		loopErr := NewLoopDetectedError("recursive processing", 150)

		// Attempt recovery
		recoveryErr := registry.AttemptRecovery(ctx, loopErr, execCtx)
		if recoveryErr == nil {
			t.Error("Loop detection should halt execution")
		}

		// Retry count should not increase for loops
		if execCtx.RetryCount > 0 {
			t.Errorf("Loop detection should not increment retry count, got %d", execCtx.RetryCount)
		}
	})

	// Test 4: Timeout recovery
	t.Run("TimeoutIntegration", func(t *testing.T) {
		registry := NewRecoveryRegistry()
		execCtx := &state.AgentExecutionContext{
			RetryCount: 0,
			ErrorCount: 0,
			LastError:  nil,
		}

		timeoutErr := NewTimeoutError("long-running operation", 5*time.Minute)

		// First recovery attempt
		recoveryErr := registry.AttemptRecovery(ctx, timeoutErr, execCtx)
		if recoveryErr == nil {
			t.Error("Timeout recovery should provide guidance, not silent success")
		}

		// Second attempt should fail due to retry limit
		recoveryErr = registry.AttemptRecovery(ctx, timeoutErr, execCtx)
		if recoveryErr == nil {
			t.Error("Should fail due to retry limit")
		}
	})

	// Test 5: Resource limit recovery
	t.Run("ResourceLimitIntegration", func(t *testing.T) {
		registry := NewRecoveryRegistry()
		execCtx := &state.AgentExecutionContext{
			RetryCount: 0,
			ErrorCount: 0,
			LastError:  nil,
		}

		// Test multiple resource types
		resourceErrs := []error{
			NewResourceLimitError("memory", "14.2GB", "16GB"),
			NewResourceLimitError("cpu", "85%", "80%"),
			NewResourceLimitError("disk", "500MB", "1GB"),
		}

		for _, err := range resourceErrs {
			recoveryErr := registry.AttemptRecovery(ctx, err, execCtx)
			if recoveryErr == nil {
				t.Error("Resource recovery should provide guidance")
			}
		}
	})

	// Test 6: Panic recovery
	t.Run("PanicRecoveryIntegration", func(t *testing.T) {
		registry := NewRecoveryRegistry()
		execCtx := &state.AgentExecutionContext{
			RetryCount: 0,
			ErrorCount: 0,
			LastError:  nil,
		}

		panicErr := NewPanicError(errors.New("runtime panic: index out of range"), "")

		recoveryErr := registry.AttemptRecovery(ctx, panicErr, execCtx)
		if recoveryErr == nil {
			t.Error("Panic recovery should acknowledge but continue")
		}

		if execCtx.LastError == nil {
			t.Error("Last error should be set after panic recovery")
		}
	})
}

// Test recovery system under load
func TestRecoverySystemLoad(t *testing.T) {
	ctx := context.Background()
	registry := NewRecoveryRegistry()

	// Simulate concurrent recovery attempts
	const numConcurrent = 10
	results := make(chan error, numConcurrent)

	for i := 0; i < numConcurrent; i++ {
		go func(id int) {
			execCtx := &state.AgentExecutionContext{
				RetryCount: 0,
				ErrorCount: 0,
				LastError:  nil,
			}

			var err error
			switch id % 6 {
			case 0:
				err = NewFileOutdatedError(errors.New("test"), "/tmp/test.txt")
			case 1:
				err = NewEditFailedError(errors.New("test"), "/tmp/test", "old", "new")
			case 2:
				err = NewLoopDetectedError("test", 100)
			case 3:
				err = NewTimeoutError("test", time.Second)
			case 4:
				err = NewResourceLimitError("memory", "8GB", "16GB")
			case 5:
				err = NewPanicError(errors.New("test panic"), "")
			}

			results <- registry.AttemptRecovery(ctx, err, execCtx)
		}(i)
	}

	// Collect results
	successCount := 0
	for i := 0; i < numConcurrent; i++ {
		err := <-results
		if err == nil {
			successCount++
		}
	}

	t.Logf("Load test: %d/%d recoveries succeeded", successCount, numConcurrent)
}

// Test statistics tracking
func TestRecoveryStatistics(t *testing.T) {
	stats := NewRecoveryStatistics()

	// Record various recovery attempts
	stats.RecordAttempt("File Outdated Recovery", ErrorTypeFileOutdated, true, 1)
	stats.RecordAttempt("Edit Failed Recovery", ErrorTypeEditFailed, false, 2)
	stats.RecordAttempt("Loop Detected Recovery", ErrorTypeLoopDetected, true, 0)

	// Verify statistics
	if stats.TotalAttempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", stats.TotalAttempts)
	}

	if stats.SuccessCount != 2 {
		t.Errorf("Expected 2 successes, got %d", stats.SuccessCount)
	}

	if stats.FailureCount != 1 {
		t.Errorf("Expected 1 failure, got %d", stats.FailureCount)
	}

	if len(stats.StrategyCounts) != 3 {
		t.Errorf("Expected 3 strategy types, got %d", len(stats.StrategyCounts))
	}
}

// Test registry diagnostics
func TestRegistryDiagnostics(t *testing.T) {
	registry := NewRecoveryRegistry()

	// Add custom strategy
	customStrategy := &CustomTestStrategy{}
	registry.AddStrategy(customStrategy)

	// Get diagnostics
	diagnostics := registry.GetDiagnostics()

	if diagnostics.TotalStrategies != 7 { // 6 default + 1 custom
		t.Errorf("Expected 7 strategies, got %d", diagnostics.TotalStrategies)
	}

	if diagnostics.GlobalMaxAttempts != 3 {
		t.Errorf("Expected max attempts 3, got %d", diagnostics.GlobalMaxAttempts)
	}

	// Verify strategy name is included
	found := false
	for _, name := range diagnostics.StrategyNames {
		if name == "Custom Test Recovery" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Custom strategy not found in diagnostics")
	}
}

// Error propagation test
func TestErrorPropagation(t *testing.T) {
	registry := NewRecoveryRegistry()
	ctx := context.Background()
	execCtx := &state.AgentExecutionContext{
		RetryCount: 0,
		ErrorCount: 0,
		LastError:  nil,
	}

	// Original error should be wrapped, not lost
	originalErr := errors.New("something went wrong")
	wrappedErr := NewFileOutdatedError(originalErr, "/tmp/test.txt")

	recoveryErr := registry.AttemptRecovery(ctx, wrappedErr, execCtx)

	// Recovery might succeed (nil) or fail with error, both valid
	if recoveryErr != nil {
		// Check that error contains original error information
		if !strings.Contains(recoveryErr.Error(), originalErr.Error()) {
			t.Errorf("Recovery error should contain original error: %v, got: %v", originalErr, recoveryErr)
		}
	} else {
		// Successful recovery - verify retry count was incremented
		if execCtx.RetryCount == 0 {
			t.Error("Retry count should be incremented after successful recovery")
		}
	}
}
