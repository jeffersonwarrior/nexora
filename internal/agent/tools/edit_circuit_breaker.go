package tools

import (
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// EditCircuitBreaker tracks edit failures per file and triggers alternative suggestions
// after repeated failures. This helps models like GLM-4.7 that get stuck in retry loops.
type EditCircuitBreaker struct {
	mu       sync.RWMutex
	failures map[string]*fileEditState // key: sessionID:filePath
}

type fileEditState struct {
	failureCount  int
	lastFailure   time.Time
	lastOldString string
	suggestions   []string
	circuitOpen   bool
}

const (
	// MaxEditFailures before circuit breaker triggers
	MaxEditFailures = 3
	// CircuitResetTime after which failures are cleared
	CircuitResetTime = 5 * time.Minute
)

var (
	globalCircuitBreaker     *EditCircuitBreaker
	globalCircuitBreakerOnce sync.Once
)

// GetEditCircuitBreaker returns the global circuit breaker instance
func GetEditCircuitBreaker() *EditCircuitBreaker {
	globalCircuitBreakerOnce.Do(func() {
		globalCircuitBreaker = &EditCircuitBreaker{
			failures: make(map[string]*fileEditState),
		}
	})
	return globalCircuitBreaker
}

// RecordFailure records an edit failure and returns alternative suggestions if threshold exceeded
func (cb *EditCircuitBreaker) RecordFailure(sessionID, filePath, oldString, errorMsg string) (bool, string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	key := sessionID + ":" + filePath
	state, exists := cb.failures[key]

	if !exists {
		state = &fileEditState{}
		cb.failures[key] = state
	}

	// Reset if enough time has passed
	if time.Since(state.lastFailure) > CircuitResetTime {
		state.failureCount = 0
		state.circuitOpen = false
		state.suggestions = nil
	}

	state.failureCount++
	state.lastFailure = time.Now()
	state.lastOldString = oldString

	slog.Debug("Edit circuit breaker: failure recorded",
		"file", filePath,
		"count", state.failureCount,
		"threshold", MaxEditFailures)

	// Check if we should trip the circuit
	if state.failureCount >= MaxEditFailures && !state.circuitOpen {
		state.circuitOpen = true
		suggestion := cb.generateSuggestion(filePath, errorMsg, state.failureCount)
		state.suggestions = append(state.suggestions, suggestion)

		slog.Warn("Edit circuit breaker TRIPPED",
			"file", filePath,
			"failures", state.failureCount,
			"suggestion", suggestion)

		return true, suggestion
	}

	return false, ""
}

// RecordSuccess resets the failure counter for a file
func (cb *EditCircuitBreaker) RecordSuccess(sessionID, filePath string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	key := sessionID + ":" + filePath
	if state, exists := cb.failures[key]; exists {
		state.failureCount = 0
		state.circuitOpen = false
		state.suggestions = nil
	}
}

// IsCircuitOpen checks if the circuit is open for a file
func (cb *EditCircuitBreaker) IsCircuitOpen(sessionID, filePath string) bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	key := sessionID + ":" + filePath
	if state, exists := cb.failures[key]; exists {
		// Auto-reset after timeout
		if time.Since(state.lastFailure) > CircuitResetTime {
			return false
		}
		return state.circuitOpen
	}
	return false
}

// GetFailureCount returns the current failure count for a file
func (cb *EditCircuitBreaker) GetFailureCount(sessionID, filePath string) int {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	key := sessionID + ":" + filePath
	if state, exists := cb.failures[key]; exists {
		return state.failureCount
	}
	return 0
}

// generateSuggestion creates helpful suggestions based on failure patterns
func (cb *EditCircuitBreaker) generateSuggestion(filePath, errorMsg string, failureCount int) string {
	suggestions := []string{
		fmt.Sprintf("\n\n⚠️ EDIT CIRCUIT BREAKER TRIPPED (%d consecutive failures on %s)\n", failureCount, filePath),
		"The edit tool has failed multiple times. Consider these alternatives:\n",
		"1. **Use bash with sed**: `sed -i 's/old_pattern/new_pattern/g' " + filePath + "`\n",
		"2. **Rewrite entire file**: Use the `write` tool to replace the whole file content\n",
		"3. **View file first**: Use `view` tool to see exact current content before editing\n",
		"4. **Try different context**: Include 5+ lines of surrounding context in old_string\n",
	}

	// Add error-specific suggestions
	if errorMsg != "" {
		if containsAny(errorMsg, "not found", "no match") {
			suggestions = append(suggestions,
				"5. **The text you're trying to match may have changed** - re-read the file\n")
		}
		if containsAny(errorMsg, "multiple", "appears") {
			suggestions = append(suggestions,
				"5. **Multiple matches found** - add more unique context or use replace_all=true\n")
		}
	}

	result := ""
	for _, s := range suggestions {
		result += s
	}
	return result
}

// containsAny checks if s contains any of the substrings
func containsAny(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if len(sub) > 0 && len(s) >= len(sub) {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
		}
	}
	return false
}

// Cleanup removes stale entries older than CircuitResetTime
func (cb *EditCircuitBreaker) Cleanup() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	for key, state := range cb.failures {
		if time.Since(state.lastFailure) > CircuitResetTime*2 {
			delete(cb.failures, key)
		}
	}
}
