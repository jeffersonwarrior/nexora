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

// Test strategies implementation
func TestRecoveryStrategies(t *testing.T) {
	ctx := context.Background()
	execCtx := &state.AgentExecutionContext{
		RetryCount: 0,
		ErrorCount: 0,
		LastError:  nil,
	}

	t.Run("FileOutdatedStrategy", func(t *testing.T) {
		strategy := &FileOutdatedStrategy{}

		// Create a temp file for testing
		tmpFile := "/tmp/test_recovery_file.txt"
		defer os.Remove(tmpFile)
		
		// Create the file
		if err := os.WriteFile(tmpFile, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}

		// Test CanRecover
		err := NewFileOutdatedError(errors.New("file modified"), tmpFile)
		if !strategy.CanRecover(err) {
			t.Error("FileOutdatedStrategy should recover from file outdated errors")
		}

		// Test generic error detection
		genericErr := errors.New("file has been modified externally")
		if !strategy.CanRecover(genericErr) {
			t.Error("FileOutdatedStrategy should recover from generic file modified errors")
		}

		// Test recovery with existing file
		recoveryErr := strategy.Recover(ctx, err, execCtx)
		if recoveryErr != nil {
			t.Errorf("FileOutdatedStrategy recovery failed: %v", recoveryErr)
		}
	})

	t.Run("EditFailedStrategy", func(t *testing.T) {
		strategy := &EditFailedStrategy{}

		// Test CanRecover
		err := NewEditFailedError(errors.New("whitespace mismatch"), "/tmp/test.txt", "old", "new")
		if !strategy.CanRecover(err) {
			t.Error("EditFailedStrategy should recover from edit failed errors")
		}

		// Test recovery with missing context
		badErr := errors.New("edit failed")
		recoveryErr := strategy.Recover(ctx, badErr, execCtx)
		if recoveryErr == nil {
			t.Error("EditFailedStrategy should fail recovery without context")
		}
	})

	t.Run("LoopDetectedStrategy", func(t *testing.T) {
		strategy := &LoopDetectedStrategy{}

		// Test CanRecover
		err := NewLoopDetectedError("processing", 100)
		if !strategy.CanRecover(err) {
			t.Error("LoopDetectedStrategy should recover from loop detected errors")
		}

		// Test MaxRetries
		if strategy.MaxRetries() != 0 {
			t.Error("LoopDetectedStrategy should have 0 max retries")
		}

		// Test recovery
		recoveryErr := strategy.Recover(ctx, err, execCtx)
		if recoveryErr == nil {
			t.Error("LoopDetectedStrategy should halt execution")
		}
	})

	t.Run("TimeoutStrategy", func(t *testing.T) {
		strategy := &TimeoutStrategy{}

		// Test CanRecover
		err := NewTimeoutError("operation", 30*time.Second)
		if !strategy.CanRecover(err) {
			t.Error("TimeoutStrategy should recover from timeout errors")
		}

		// Test MaxRetries
		if strategy.MaxRetries() != 1 {
			t.Error("TimeoutStrategy should have 1 max retry")
		}
	})

	t.Run("ResourceLimitStrategy", func(t *testing.T) {
		strategy := &ResourceLimitStrategy{}

		// Test CanRecover
		err := NewResourceLimitError("memory", "8GB", "16GB")
		if !strategy.CanRecover(err) {
			t.Error("ResourceLimitStrategy should recover from resource limit errors")
		}

		// Test MaxRetries
		if strategy.MaxRetries() != 5 {
			t.Error("ResourceLimitStrategy should have 5 max retries")
		}
	})

	t.Run("PanicStrategy", func(t *testing.T) {
		strategy := &PanicStrategy{}

		// Test CanRecover
		err := NewPanicError(errors.New("runtime panic"), "stack trace here")
		if !strategy.CanRecover(err) {
			t.Error("PanicStrategy should recover from panic errors")
		}

		// Test MaxRetries
		if strategy.MaxRetries() != 1 {
			t.Error("PanicStrategy should have 1 max retry")
		}
	})
}

func TestRecoveryRegistry(t *testing.T) {
	registry := NewRecoveryRegistry()
	execCtx := &state.AgentExecutionContext{
		RetryCount: 0,
		ErrorCount: 0,
		LastError:  nil,
	}

	t.Run("FindStrategy", func(t *testing.T) {
		// Test finding strategy for known error
		editErr := NewEditFailedError(errors.New("edit failed"), "/tmp/test.txt", "old", "new")
		strategy := registry.FindStrategy(editErr)
		if strategy == nil {
			t.Error("Registry should find strategy for edit failed error")
		}
		if strategy.Name() != "Edit Failed Recovery" {
			t.Errorf("Expected 'Edit Failed Recovery', got '%s'", strategy.Name())
		}

		// Test no strategy for unknown error
		unknownErr := errors.New("unknown error")
		strategy = registry.FindStrategy(unknownErr)
		if strategy != nil {
			t.Error("Registry should not find strategy for unknown error")
		}
	})

	t.Run("AttemptRecovery", func(t *testing.T) {
		// Test successful recovery - create temp file first
		tmpFile := "/tmp/test_recovery.txt"
		err := []byte("test content")
		if writeErr := os.WriteFile(tmpFile, err, 0644); writeErr != nil {
			t.Fatalf("Failed to create test file: %v", writeErr)
		}
		defer os.Remove(tmpFile)

		// Use the same file path in the error
		fileErr := NewFileOutdatedError(errors.New("file modified"), tmpFile)

		ctx := context.Background()
		recoveryErr := registry.AttemptRecovery(ctx, fileErr, execCtx)
		if recoveryErr != nil {
			t.Errorf("Recovery failed: %v", recoveryErr)
		}

		// Test retry limit enforcement
		execCtx.RetryCount = 5 // Exceed max retries
		recoveryErr = registry.AttemptRecovery(ctx, fileErr, execCtx)
		if recoveryErr == nil {
			t.Error("Should fail due to retry limit")
		}
	})

	t.Run("SetMaxAttempts", func(t *testing.T) {
		registry.SetMaxAttempts(5)
		if registry.maxAttempts != 5 {
			t.Error("SetMaxAttempts failed")
		}
	})

	t.Run("AddStrategy", func(t *testing.T) {
		initialCount := len(registry.GetAllStrategies())
		customStrategy := &CustomTestStrategy{}
		registry.AddStrategy(customStrategy)

		newCount := len(registry.GetAllStrategies())
		if newCount != initialCount+1 {
			t.Errorf("Expected %d strategies, got %d", initialCount+1, newCount)
		}
	})
}

// Custom test strategy for testing AddStrategy
type CustomTestStrategy struct{}

func (s *CustomTestStrategy) CanRecover(err error) bool {
	return err.Error() == "custom test error"
}

func (s *CustomTestStrategy) Recover(ctx context.Context, err error, execCtx *state.AgentExecutionContext) error {
	return nil
}

func (s *CustomTestStrategy) MaxRetries() int {
	return 1
}

func (s *CustomTestStrategy) Name() string {
	return "Custom Test Recovery"
}

func TestRecoverableError(t *testing.T) {
	originalErr := errors.New("original error")
	context := map[string]interface{}{
		"key":       "value",
		"timestamp": time.Now(),
	}

	recoverableErr := NewRecoverableError(originalErr, "test_type", context)

	// Test Error method
	if !strings.Contains(recoverableErr.Error(), "test_type") {
		t.Error("Error string should contain error type")
	}

	// Test Unwrap method
	if errors.Unwrap(recoverableErr) != originalErr {
		t.Error("Unwrap should return original error")
	}
}

// Benchmark tests for performance validation
func BenchmarkFindStrategy(b *testing.B) {
	registry := NewRecoveryRegistry()
	err := NewEditFailedError(errors.New("test"), "/tmp/test", "old", "new")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		registry.FindStrategy(err)
	}
}

func BenchmarkAttemptRecovery(b *testing.B) {
	registry := NewRecoveryRegistry()
	ctx := context.Background()
	execCtx := &state.AgentExecutionContext{
		RetryCount: 0,
		ErrorCount: 0,
		LastError:  nil,
	}
	err := NewFileOutdatedError(errors.New("test"), "/tmp/test")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		registry.AttemptRecovery(ctx, err, execCtx)
	}
}
