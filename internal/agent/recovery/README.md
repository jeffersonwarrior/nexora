# Error Recovery System

The error recovery system provides automatic recovery from transient errors in Nexora agents, improving reliability and reducing user intervention.

## Architecture

### Core Components

- **RecoveryStrategy Interface**: Defines how to recover from specific error types
- **RecoveryRegistry**: Manages all recovery strategies and coordinates recovery
- **AgentExecutionContext**: Tracks execution state and recovery attempts
- **RecoverableError**: Wraps errors with recovery context

### Error Types Supported

1. **File Outdated** - Files modified since read
2. **Edit Failed** - Whitespace mismatches, text not found
3. **Loop Detected** - Infinite loops in agent execution
4. **Timeout** - Operations exceeding time limits
5. **Resource Limit** - CPU/memory/disk constraints
6. **Panic** - Runtime panics with stack traces

## Usage

### Basic Recovery

```go
// Create registry with default strategies
registry := recovery.NewRecoveryRegistry()

// Create execution context
execCtx := state.NewAgentExecutionContext("session-123")

// Attempt recovery for an error
err := NewFileOutdatedError(os.ErrNotExist, "/tmp/file.txt")
recoveryErr := registry.AttemptRecovery(ctx, err, execCtx)
```

### Custom Strategies

```go
type CustomRecoveryStrategy struct{}

func (s *CustomRecoveryStrategy) CanRecover(err error) bool {
    return strings.Contains(err.Error(), "custom error")
}

func (s *CustomRecoveryStrategy) Recover(ctx context.Context, err error, execCtx *state.AgentExecutionContext) error {
    // Custom recovery logic
    return nil
}

func (s *CustomRecoveryStrategy) MaxRetries() int {
    return 3
}

func (s *CustomRecoveryStrategy) Name() string {
    return "Custom Recovery"
}

// Add to registry
registry.AddStrategy(&CustomRecoveryStrategy{})
```

## Strategy Behaviors

### File Outdated Strategy
- **Max Retries**: 2
- **Action**: Re-read file, refresh cached content
- **Context**: Requires `file_path` in error context

### Edit Failed Strategy  
- **Max Retries**: 3
- **Action**: Attempt whitespace normalization, suggest write tool fallback
- **Context**: Requires `file_path`, optionally `old_text` and `new_text`

### Loop Detected Strategy
- **Max Retries**: 0
- **Action**: Immediately halt execution to prevent infinite loops
- **Behavior**: Always fails for safety

### Timeout Strategy
- **Max Retries**: 1
- **Action**: Suggest longer timeout or task decomposition
- **Context**: Uses `operation` and `timeout` from error context

### Resource Limit Strategy
- **Max Retries**: 5
- **Action**: Pause execution, wait for resources to become available
- **Context**: Uses `resource_type`, `current`, and `limit` values

### Panic Strategy
- **Max Retries**: 1
- **Action**: Log panic, capture stack trace, graceful recovery
- **Context**: Optionally includes `stack_trace`

## Integration with Agent

The recovery system integrates with the agent state machine:

1. **Error Detection**: Errors are wrapped as `RecoverableError`
2. **Strategy Selection**: Registry finds appropriate recovery strategy
3. **Recovery Attempt**: Strategy attempts recovery with retry limits
4. **State Transition**: Success/failure triggers appropriate state transitions
5. **Retry Tracking**: Execution context tracks recovery attempts

## Configuration

### Global Retry Limit

```go
registry := recovery.NewRecoveryRegistry()
registry.SetMaxAttempts(5) // Override default of 3
```

### Custom Thresholds per Strategy

Each strategy defines its own `MaxRetries()` method.

## Testing

The recovery system includes comprehensive tests:

- Unit tests for each strategy
- Integration tests for end-to-end scenarios
- Load tests for concurrent recovery
- Statistics tracking for diagnostics

```bash
go test ./internal/agent/recovery/... -v
```

## Diagnostics

### Registry Information

```go
diagnostics := registry.GetDiagnostics()
fmt.Printf("Total Strategies: %d\n", diagnostics.TotalStrategies)
fmt.Printf("Names: %v\n", diagnostics.StrategyNames)
```

### Statistics Tracking

```go
stats := recovery.NewRecoveryStatistics()
stats.RecordAttempt("File Outdated Recovery", "file_outdated", true, 1)
```