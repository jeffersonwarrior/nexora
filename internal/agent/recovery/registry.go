package recovery

import (
	"context"
	"fmt"

	"github.com/nexora/cli/internal/agent/state"
)

// RecoveryRegistry manages all recovery strategies and coordinates error recovery
type RecoveryRegistry struct {
	strategies  []RecoveryStrategy
	maxAttempts int // Global limit (default: 3)
}

// NewRecoveryRegistry creates a new recovery registry with default strategies
func NewRecoveryRegistry() *RecoveryRegistry {
	registry := &RecoveryRegistry{
		maxAttempts: 3,
		strategies: []RecoveryStrategy{
			&FileOutdatedStrategy{},
			&EditFailedStrategy{},
			&LoopDetectedStrategy{},
			&TimeoutStrategy{},
			&ResourceLimitStrategy{},
			&PanicStrategy{},
		},
	}
	return registry
}

// SetMaxAttempts sets the global maximum retry attempts
func (rr *RecoveryRegistry) SetMaxAttempts(max int) {
	if max < 1 {
		max = 3 // Minimum of 3 attempts
	}
	rr.maxAttempts = max
}

// AddStrategy adds a new recovery strategy to the registry
func (rr *RecoveryRegistry) AddStrategy(strategy RecoveryStrategy) {
	rr.strategies = append(rr.strategies, strategy)
}

// RemoveStrategy removes a recovery strategy by name
func (rr *RecoveryRegistry) RemoveStrategy(name string) {
	for i, strategy := range rr.strategies {
		if strategy.Name() == name {
			rr.strategies = append(rr.strategies[:i], rr.strategies[i+1:]...)
			break
		}
	}
}

// FindStrategy finds the appropriate recovery strategy for a given error
func (rr *RecoveryRegistry) FindStrategy(err error) RecoveryStrategy {
	for _, strategy := range rr.strategies {
		if strategy.CanRecover(err) {
			return strategy
		}
	}
	return nil
}

// AttemptRecovery tries to recover from an error using an appropriate strategy
func (rr *RecoveryRegistry) AttemptRecovery(ctx context.Context, err error, execCtx *state.AgentExecutionContext) error {
	strategy := rr.FindStrategy(err)
	if strategy == nil {
		return fmt.Errorf("no recovery strategy found for error: %w", err)
	}
	
	// Check both global and strategy-specific retry limits
	if execCtx.RetryCount >= rr.maxAttempts {
		return fmt.Errorf("global retry limit (%d) exceeded: %w", rr.maxAttempts, err)
	}
	
	if execCtx.RetryCount >= strategy.MaxRetries() {
		return fmt.Errorf("strategy retry limit (%d) exceeded for %s: %w", 
			strategy.MaxRetries(), strategy.Name(), err)
	}
	
	// Attempt recovery
	recoveryErr := strategy.Recover(ctx, err, execCtx)
	if recoveryErr != nil {
		return fmt.Errorf("recovery failed using %s: %w", strategy.Name(), recoveryErr)
	}
	
	// Increment retry count
	execCtx.RetryCount++
	return nil
}

// GetAllStrategies returns all registered strategies (for testing/diagnostics)
func (rr *RecoveryRegistry) GetAllStrategies() []RecoveryStrategy {
	return append([]RecoveryStrategy{}, rr.strategies...)
}

// GetStrategyNames returns the names of all registered strategies
func (rr *RecoveryRegistry) GetStrategyNames() []string {
	names := make([]string, len(rr.strategies))
	for i, strategy := range rr.strategies {
		names[i] = strategy.Name()
	}
	return names
}

// CanRecover returns true if any strategy can handle the given error
func (rr *RecoveryRegistry) CanRecover(err error) bool {
	return rr.FindStrategy(err) != nil
}

// GetMaxAttempts returns the global maximum retry attempts
func (rr *RecoveryRegistry) GetMaxAttempts() int {
	return rr.maxAttempts
}

// Reset resets the registry to default state
func (rr *RecoveryRegistry) Reset() {
	rr.maxAttempts = 3
	rr.strategies = []RecoveryStrategy{
		&FileOutdatedStrategy{},
		&EditFailedStrategy{},
		&LoopDetectedStrategy{},
		&TimeoutStrategy{},
		&ResourceLimitStrategy{},
		&PanicStrategy{},
	}
}

// Diagnostics provides information about the recovery system
type RecoveryDiagnostics struct {
	TotalStrategies    int      `json:"total_strategies"`
	StrategyNames      []string `json:"strategy_names"`
	GlobalMaxAttempts  int      `json:"global_max_attempts"`
	LastRecoveredError string   `json:"last_recovered_error,omitempty"`
}

// GetDiagnostics returns diagnostic information about the recovery registry
func (rr *RecoveryRegistry) GetDiagnostics() RecoveryDiagnostics {
	return RecoveryDiagnostics{
		TotalStrategies:   len(rr.strategies),
		StrategyNames:     rr.GetStrategyNames(),
		GlobalMaxAttempts: rr.maxAttempts,
	}
}

// CreateRecoveryStatistics tracks recovery attempts and outcomes
type RecoveryStatistics struct {
	TotalAttempts    int                        `json:"total_attempts"`
	SuccessCount     int                        `json:"success_count"`
	FailureCount     int                        `json:"failure_count"`
	StrategyCounts   map[string]int             `json:"strategy_counts"`
	ErrorTypeCounts  map[string]int             `json:"error_type_counts"`
	AverageRetries   float64                    `json:"average_retries"`
}

// NewRecoveryStatistics creates empty statistics
func NewRecoveryStatistics() *RecoveryStatistics {
	return &RecoveryStatistics{
		StrategyCounts:  make(map[string]int),
		ErrorTypeCounts: make(map[string]int),
	}
}

// RecordAttempt records a recovery attempt
func (rs *RecoveryStatistics) RecordAttempt(strategy string, errorType string, success bool, retryCount int) {
	rs.TotalAttempts++
	if success {
		rs.SuccessCount++
	} else {
		rs.FailureCount++
	}
	
	rs.StrategyCounts[strategy]++
	rs.ErrorTypeCounts[errorType]++
	
	// Update average retries
	rs.AverageRetries = float64(retryCount)
}