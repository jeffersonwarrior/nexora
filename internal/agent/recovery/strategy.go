package recovery

import (
	"context"
	"fmt"

	"github.com/nexora/nexora/internal/agent/state"
)

// RecoveryStrategy defines how to recover from specific types of errors
type RecoveryStrategy interface {
	// CanRecover returns true if this strategy can handle the given error
	CanRecover(err error) bool

	// Recover attempts to recover from the error
	Recover(ctx context.Context, err error, execCtx *state.AgentExecutionContext) error

	// MaxRetries returns the maximum number of times this strategy should be attempted
	MaxRetries() int

	// Name returns a human-readable name for this recovery strategy
	Name() string
}

// Error types
const (
	ErrorTypeFileOutdated  = "file_outdated"
	ErrorTypeEditFailed    = "edit_failed"
	ErrorTypeLoopDetected  = "loop_detected"
	ErrorTypeTimeout       = "timeout"
	ErrorTypeResourceLimit = "resource_limit"
	ErrorTypePanic         = "panic"
)

// RecoverableError wraps errors to indicate they are recoverable
type RecoverableError struct {
	Err       error
	ErrorType string
	Context   map[string]interface{}
}

func (re *RecoverableError) Error() string {
	return fmt.Sprintf("recoverable error [%s]: %v", re.ErrorType, re.Err)
}

func (re *RecoverableError) Unwrap() error {
	return re.Err
}

// NewRecoverableError creates a new recoverable error
func NewRecoverableError(err error, errorType string, context map[string]interface{}) *RecoverableError {
	return &RecoverableError{
		Err:       err,
		ErrorType: errorType,
		Context:   context,
	}
}
