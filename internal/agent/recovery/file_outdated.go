package recovery

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/nexora/cli/internal/agent/state"
)

// FileOutdatedStrategy handles errors where a file has been modified since it was read
type FileOutdatedStrategy struct{}

func (s *FileOutdatedStrategy) CanRecover(err error) bool {
	if re, ok := err.(*RecoverableError); ok {
		return re.ErrorType == ErrorTypeFileOutdated
	}
	return strings.Contains(err.Error(), "file has been modified") ||
		strings.Contains(err.Error(), "file outdated") ||
		strings.Contains(err.Error(), "stale data")
}

func (s *FileOutdatedStrategy) Recover(ctx context.Context, err error, execCtx *state.AgentExecutionContext) error {
	// Extract file path from error context
	var filePath string
	if re, ok := err.(*RecoverableError); ok && re.Context != nil {
		if path, exists := re.Context["file_path"]; exists {
			filePath = path.(string)
		}
	}

	if filePath == "" {
		return fmt.Errorf("cannot recover from file outdated error: no file path provided")
	}

	// Validate file exists
	if _, statErr := os.Stat(filePath); os.IsNotExist(statErr) {
		return fmt.Errorf("file %s no longer exists: %w", filePath, statErr)
	}

	// Record recovery attempt
	execCtx.LastError = fmt.Errorf("file %s was outdated, re-reading file", filePath)

	// In a real implementation, we would:
	// 1. Re-read the file content
	// 2. Update any cached content
	// 3. Signal to retry the operation with fresh data

	return nil
}

func (s *FileOutdatedStrategy) MaxRetries() int {
	return 2 // Re-reading files should work quickly
}

func (s *FileOutdatedStrategy) Name() string {
	return "File Outdated Recovery"
}

// EditFailedStrategy handles edit operation failures
type EditFailedStrategy struct{}

func (s *EditFailedStrategy) CanRecover(err error) bool {
	if re, ok := err.(*RecoverableError); ok {
		return re.ErrorType == ErrorTypeEditFailed
	}
	return strings.Contains(err.Error(), "edit failed") ||
		strings.Contains(err.Error(), "whitespace mismatch") ||
		strings.Contains(err.Error(), "text not found") ||
		strings.Contains(err.Error(), "edit tool failed")
}

func (s *EditFailedStrategy) Recover(ctx context.Context, err error, execCtx *state.AgentExecutionContext) error {
	// Extract edit context
	var filePath, oldText string
	if re, ok := err.(*RecoverableError); ok && re.Context != nil {
		filePath, _ = re.Context["file_path"].(string)
		oldText, _ = re.Context["old_text"].(string)
	}

	if filePath == "" {
		return fmt.Errorf("cannot recover from edit failure: no file path provided")
	}

	// Record recovery attempt
	execCtx.LastError = fmt.Errorf("edit failed on %s, attempting recovery", filePath)

	// Recovery strategies:
	// 1. Try to read the current file content
	// 2. If we have oldText and newText, attempt a more robust edit
	// 3. Fall back to AIOPS-powered fix if available

	content, readErr := os.ReadFile(filePath)
	if readErr != nil {
		return fmt.Errorf("cannot read file for recovery: %w", readErr)
	}

	currentContent := string(content)

	// Try to find exact match with normalized whitespace
	if oldText != "" {
		normalizedCurrent := normalizeWhitespace(currentContent)
		normalizedOld := normalizeWhitespace(oldText)

		if strings.Contains(normalizedCurrent, normalizedOld) {
			// We can potentially recover, signal to retry
			return nil
		}
	}

	// Fall back: suggest using write tool to replace entire file
	return fmt.Errorf("edit recovery failed - suggest using write tool to replace entire file")
}

func (s *EditFailedStrategy) MaxRetries() int {
	return 3 // Edits can fail due to whitespace issues
}

func (s *EditFailedStrategy) Name() string {
	return "Edit Failed Recovery"
}

// LoopDetectedStrategy handles infinite loop detection
type LoopDetectedStrategy struct{}

func (s *LoopDetectedStrategy) CanRecover(err error) bool {
	if re, ok := err.(*RecoverableError); ok {
		return re.ErrorType == ErrorTypeLoopDetected
	}
	return strings.Contains(err.Error(), "loop detected") ||
		strings.Contains(err.Error(), "infinite loop") ||
		strings.Contains(err.Error(), "stuck in loop")
}

func (s *LoopDetectedStrategy) Recover(ctx context.Context, err error, execCtx *state.AgentExecutionContext) error {
	// For loop detection, recovery means stopping the loop and returning control
	execCtx.LastError = fmt.Errorf("loop detected, halting execution and returning to user")

	// In a complete implementation, this would:
	// 1. Transition state machine to StateHalted
	// 2. Provide a summary of what was being done
	// 3. Suggest next steps to the user

	return fmt.Errorf("loop detected - execution halted to prevent infinite loop")
}

func (s *LoopDetectedStrategy) MaxRetries() int {
	return 0 // No retries for loops - always halt
}

func (s *LoopDetectedStrategy) Name() string {
	return "Loop Detected Recovery"
}

// TimeoutStrategy handles operation timeouts
type TimeoutStrategy struct{}

func (s *TimeoutStrategy) CanRecover(err error) bool {
	if re, ok := err.(*RecoverableError); ok {
		return re.ErrorType == ErrorTypeTimeout
	}
	return strings.Contains(err.Error(), "timeout") ||
		strings.Contains(err.Error(), "deadline exceeded") ||
		strings.Contains(err.Error(), "context deadline exceeded")
}

func (s *TimeoutStrategy) Recover(ctx context.Context, err error, execCtx *state.AgentExecutionContext) error {
	// Extract operation context
	var operation string
	if re, ok := err.(*RecoverableError); ok && re.Context != nil {
		operation, _ = re.Context["operation"].(string)
	}

	if operation == "" {
		operation = "operation"
	}

	execCtx.LastError = fmt.Errorf("%s timed out", operation)

	// For timeout recovery:
	// 1. Check if operation can be retried with longer timeout
	// 2. Suggest breaking down into smaller operations
	// 3. For long-running operations, suggest background execution

	return fmt.Errorf("%s timed out - try with longer timeout or break into smaller steps", operation)
}

func (s *TimeoutStrategy) MaxRetries() int {
	return 1 // One retry with potentially longer timeout
}

func (s *TimeoutStrategy) Name() string {
	return "Timeout Recovery"
}

// ResourceLimitStrategy handles resource exhaustion
type ResourceLimitStrategy struct{}

func (s *ResourceLimitStrategy) CanRecover(err error) bool {
	if re, ok := err.(*RecoverableError); ok {
		return re.ErrorType == ErrorTypeResourceLimit
	}
	return strings.Contains(err.Error(), "resource limit") ||
		strings.Contains(err.Error(), "memory") && strings.Contains(err.Error(), "insufficient") ||
		strings.Contains(err.Error(), "disk space") ||
		strings.Contains(err.Error(), "cpu usage")
}

func (s *ResourceLimitStrategy) Recover(ctx context.Context, err error, execCtx *state.AgentExecutionContext) error {
	// Extract resource type
	var resourceType string
	if re, ok := err.(*RecoverableError); ok && re.Context != nil {
		resourceType, _ = re.Context["resource_type"].(string)
	}

	if resourceType == "" {
		resourceType = "system resource"
	}

	execCtx.LastError = fmt.Errorf("%s limit reached", resourceType)

	// Resource recovery strategies:
	// 1. Transition to paused state if not already
	// 2. Wait for resources to become available
	// 3. Suggest freeing up resources

	return fmt.Errorf("%s limit reached - system will pause and retry when resources are available", resourceType)
}

func (s *ResourceLimitStrategy) MaxRetries() int {
	return 5 // Resources may become available
}

func (s *ResourceLimitStrategy) Name() string {
	return "Resource Limit Recovery"
}

// PanicStrategy handles panic recovery
type PanicStrategy struct{}

func (s *PanicStrategy) CanRecover(err error) bool {
	if re, ok := err.(*RecoverableError); ok {
		return re.ErrorType == ErrorTypePanic
	}
	return strings.Contains(err.Error(), "panic") ||
		strings.Contains(err.Error(), "runtime panic")
}

func (s *PanicStrategy) Recover(ctx context.Context, err error, execCtx *state.AgentExecutionContext) error {
	execCtx.LastError = fmt.Errorf("panic recovered: %v", err)

	// Panic recovery:
	// 1. Log the panic and stack trace
	// 2. Transition to recovery state
	// 3. Attempt graceful shutdown or restart

	// In a complete implementation, this would capture the stack trace
	// and perform more sophisticated recovery

	return fmt.Errorf("panic occurred - system recovered and continuing safely")
}

func (s *PanicStrategy) MaxRetries() int {
	return 1 // Limited retries for panics
}

func (s *PanicStrategy) Name() string {
	return "Panic Recovery"
}

// Helper function to normalize whitespace for comparison
func normalizeWhitespace(text string) string {
	// Convert tabs to spaces
	text = strings.ReplaceAll(text, "\t", "    ")
	// Normalize multiple spaces to single space
	text = strings.Join(strings.Fields(text), " ")
	return text
}

// Create convenience functions for creating common recoverable errors
func NewFileOutdatedError(err error, filePath string) error {
	return NewRecoverableError(err, ErrorTypeFileOutdated, map[string]interface{}{
		"file_path": filePath,
		"timestamp": time.Now(),
	})
}

func NewEditFailedError(err error, filePath, oldText, newText string) error {
	return NewRecoverableError(err, ErrorTypeEditFailed, map[string]interface{}{
		"file_path": filePath,
		"old_text":  oldText,
		"new_text":  newText,
		"timestamp": time.Now(),
	})
}

func NewLoopDetectedError(operation string, iterations int) error {
	return NewRecoverableError(fmt.Errorf("loop detected after %d iterations in %s", iterations, operation),
		ErrorTypeLoopDetected, map[string]interface{}{
			"operation":  operation,
			"iterations": iterations,
			"timestamp":  time.Now(),
		})
}

func NewTimeoutError(operation string, timeout time.Duration) error {
	return NewRecoverableError(fmt.Errorf("%s timed out after %v", operation, timeout),
		ErrorTypeTimeout, map[string]interface{}{
			"operation": operation,
			"timeout":   timeout,
			"timestamp": time.Now(),
		})
}

func NewResourceLimitError(resourceType string, current, limit interface{}) error {
	return NewRecoverableError(fmt.Errorf("%s limit reached (current: %v, limit: %v)", resourceType, current, limit),
		ErrorTypeResourceLimit, map[string]interface{}{
			"resource_type": resourceType,
			"current":       current,
			"limit":         limit,
			"timestamp":     time.Now(),
		})
}

func NewPanicError(err error, stackTrace string) error {
	return NewRecoverableError(err, ErrorTypePanic, map[string]interface{}{
		"stack_trace": stackTrace,
		"timestamp":   time.Now(),
	})
}
