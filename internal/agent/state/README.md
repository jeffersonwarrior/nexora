# Agent State Machine

**Progress-aware state machine for Nexora AI agents with intelligent loop detection.**

## Philosophy

**Allow productive long-running tasks. Detect and halt true stuck loops.**

- ✅ **1000 tool calls OK** if making forward progress (unique file edits, tests passing)
- ❌ **3 identical errors** = stuck loop (halt and recover)
- 🔄 **Phase transitions reset error tracking** (fresh start per phase)
- 📊 **Track semantic progress**, not just step counts

## Architecture

```
┌──────────────────────────────────────────────────────┐
│              NEXORA STATE MACHINE                    │
└──────────────────────────────────────────────────────┘

                    ┌──────────┐
                    │   IDLE   │◄────────────────┐
                    └────┬─────┘                 │
                         │ User Prompt           │
                         ▼                       │
                ┌────────────────┐               │
                │   PROCESSING   │               │
                │     PROMPT     │               │
                └────┬───────────┘               │
                     │                           │
          LLM Response│                           │
                     ▼                           │
            ┌────────────────┐                   │
            │   STREAMING    │                   │
            │    RESPONSE    │                   │
            └────┬───────────┘                   │
                 │                               │
      Tool Calls │                               │
                 ▼                               │
        ┌──────────────────┐                     │
        │   EXECUTING      │                     │
        │     TOOL         │                     │
        └────┬─────────────┘                     │
             │                                   │
  Success    │                                   │
             ▼                                   │
   ┌─────────────────┐                          │
   │   PROGRESS      │                          │
   │     CHECK       │                          │
   └────┬────────────┘                          │
        │                                        │
        │ Making Progress?                      │
        ├─► YES → Continue Loop                 │
        ├─► NO (3x same error) → ERROR RECOVERY │
        └─► Phase Complete → PHASE TRANSITION ──┤
                                                │
   ┌─────────────────┐                          │
   │     PHASE       │                          │
   │   TRANSITION    │                          │
   └────┬────────────┘                          │
        │                                        │
        │ Reset Error Tracking                  │
        │ Start Next Phase                      │
        └────────────────────────────────────────┘
```

## Components

### 1. **StateMachine** (`machine.go`)

Core state machine with context management.

**Key Methods**:
- `TransitionTo(state)` - Change state with validation
- `RecordToolCall(tool, file, cmd, err, success)` - Track tool execution
- `RecordMessage(msg)` - Detect duplicate messages
- `StartPhase(n, desc, duration)` - Begin new phase (auto-resets tracking)
- `IsStuck()` - Check for stuck condition

### 2. **ProgressTracker** (`progress.go`)

Tracks semantic progress to distinguish productive work from loops.

**Detection Logic**:
- ❌ **Same file + same error 3 times** = stuck
- ❌ **Oscillating edits (A→B→A→B)** = stuck
- ❌ **No unique progress in 10 actions** = stuck
- ✅ **Unique file edits + test passes** = progress

**Metrics Tracked**:
- Files modified (path → content hash)
- Commands executed
- Test results
- Recent actions (last 20)
- Recent errors (last 10)
- Consecutive error count

### 3. **PhaseContext** (`phase.go`)

Manages multi-phase tasks (e.g., 10-phase refactor).

**Per-Phase Tracking**:
- Phase number & description
- Start time & expected duration
- Files changed in this phase
- Tests passed
- Blockers

**Features**:
- Automatic phase completion on `StartPhase(n+1)`
- Historical record of completed phases
- Total progress across all phases

### 4. **States** (`states.go`)

State definitions and valid transitions.

**States**:
- `StateIdle` - Waiting for work
- `StateProcessingPrompt` - Processing user input
- `StateStreamingResponse` - Streaming LLM output
- `StateExecutingTool` - Running tool call
- `StateAwaitingPermission` - Waiting for approval
- `StateErrorRecovery` - Recovering from error
- `StateProgressCheck` - Validating progress
- `StatePhaseTransition` - Moving between phases
- `StateHalted` - Terminal state (stopped)

## Usage Example

### Basic Usage
import "github.com/nexora/nexora/internal/agent/state"
sm := state.NewStateMachine(state.Config{
    SessionID: "session-123",
    Context:   ctx,
})

// Transition through states
sm.TransitionTo(state.StateProcessingPrompt)
sm.TransitionTo(state.StateStreamingResponse)
sm.TransitionTo(state.StateExecutingTool)

// Record tool execution
sm.RecordToolCall("edit", "main.go", "", "", true)

// Check for loops
if stuck, reason := sm.IsStuck(); stuck {
    log.Printf("Stuck: %s", reason)
    sm.TransitionTo(state.StateHalted)
}
```

### Multi-Phase Refactor

```go
// Create state machine with 10 phases
sm := state.NewStateMachine(state.Config{
    SessionID:   "refactor-task",
    Context:     ctx,
    TotalPhases: 10,
})

// Execute 10-phase refactor
for phase := 1; phase <= 10; phase++ {
    // Start phase (auto-resets error tracking)
    sm.StartPhase(phase, fmt.Sprintf("Refactor package %d", phase), 15*time.Minute)
    
    // Do work (100 tool calls per phase)
    for i := 0; i < 100; i++ {
        file := fmt.Sprintf("pkg%d/file%d.go", phase, i)
        sm.RecordToolCall("edit", file, "", "", true)
        
        // Check for stuck condition
        if stuck, reason := sm.IsStuck(); stuck {
            log.Printf("Phase %d stuck: %s", phase, reason)
            break
        }
    }
    
    // Run tests
    sm.RecordTest("go test ./...", true, "PASS")
}

// Complete final phase
sm.CompletePhase(true)

// Get results
progress := sm.GetTotalProgress()
log.Printf("Completed %d/%d phases in %v",
    progress.CompletedPhases,
    progress.TotalPhases,
    sm.GetElapsedTime())
```

### With Callbacks

```go
sm := state.NewStateMachine(state.Config{
    SessionID: "session-123",
    Context:   ctx,
    
    // State change callback
    OnStateChange: func(from, to state.AgentState) {
        log.Printf("State: %s → %s", from, to)
    },
    
    // Stuck detection callback
    OnStuck: func(reason string) {
        log.Printf("⚠️  Stuck detected: %s", reason)
        // Auto-recovery logic here
    },
    
    // Progress callback (every 10 tool calls)
    OnProgress: func(stats state.ProgressStats) {
        log.Printf("Progress: %d files, %d successes, %d failures",
            stats.FilesModified,
            stats.RecentSuccesses,
            stats.RecentFailures)
    },
})
```

## Loop Detection Examples

### ✅ Allowed: Productive Long Task

```
Phase 1: 100 tool calls, 25 files changed, tests pass ✅
Phase 2: 150 tool calls, 35 files changed, tests pass ✅
Phase 3: 200 tool calls, 40 files changed, tests pass ✅
...
Phase 10: 120 tool calls, 30 files changed, tests pass ✅

Total: 1,000 tool calls, 2 hours ✅ SUCCESS
```

**Why allowed**: Each phase makes unique progress (different files modified).

### ❌ Halted: Stuck Loop

```
Call 1: edit main.go → error: "old_string not found"
Call 2: edit main.go → error: "old_string not found"
Call 3: edit main.go → error: "old_string not found"

🛑 HALTED: Same error on 'main.go' repeated 3 times
```

**Why halted**: Same file, same error, 3 consecutive times = stuck.

### ❌ Halted: Oscillation

```
Call 1: edit a.go → change X
Call 2: edit b.go → change Y
Call 3: edit a.go → change X (undo)
Call 4: edit b.go → change Y (undo)

🛑 HALTED: Oscillating between 'a.go' and 'b.go'
```

**Why halted**: Alternating between same two files repeatedly.

### ✅ Allowed: Recovery After Errors

```
Phase 1:
  Call 1: edit main.go → error
  Call 2: edit main.go → error
  Call 3: edit main.go → error
  🔧 Auto-recovery: switch to write tool
  Call 4: write main.go → success ✅

Progress tracker reset for Phase 2 (fresh start) ✅
```

**Why allowed**: Errors within a phase, but recovery successful. Phase transition resets tracking.

## Testing

Comprehensive test suite in `machine_test.go`:

```bash
go test -v ./internal/agent/state/...
```

**Tests Include**:
- ✅ State transitions (valid/invalid)
- ✅ Progress tracking (stuck detection)
- ✅ Message deduplication
- ✅ Oscillation detection
- ✅ Phase context management
- ✅ **1000-step long-running task** (no false positives)
- ✅ **3-error stuck loop** (detects and halts)
- ✅ Callbacks (state change, stuck, progress)

## Integration with Agent

To integrate into `internal/agent/agent.go`:

```go
type sessionAgent struct {
    // ... existing fields ...
    
    stateMachines *csync.Map[string, *state.StateMachine]
}

func (a *sessionAgent) Run(ctx context.Context, call SessionAgentCall) (*fantasy.AgentResult, error) {
    // Create or get state machine for session
    sm, _ := a.stateMachines.GetOrCreate(call.SessionID, func() *state.StateMachine {
        return state.NewStateMachine(state.Config{
            SessionID: call.SessionID,
            Context:   ctx,
            OnStuck: func(reason string) {
                log.Printf("Session %s stuck: %s", call.SessionID, reason)
                // Trigger recovery or return to user
            },
        })
    })
    
    // Transition to processing
    sm.TransitionTo(state.StateProcessingPrompt)
    
    // Main agent loop
    for {
        // Execute LLM call
        sm.TransitionTo(state.StateStreamingResponse)
        result := a.callLLM(ctx, call)
        
        // Execute tools
        if len(result.ToolCalls) > 0 {
            sm.TransitionTo(state.StateExecutingTool)
            for _, tc := range result.ToolCalls {
                output, err := a.executeTool(ctx, tc)
                sm.RecordToolCall(tc.Name, getFile(tc), getCmd(tc), errMsg(err), err == nil)
                
                // Check if stuck
                if stuck, reason := sm.IsStuck(); stuck {
                    return nil, fmt.Errorf("stuck: %s", reason)
                }
            }
        }
        
        // Check progress
        sm.TransitionTo(state.StateProgressCheck)
        if done {
            break
        }
    }
    
    // Return to idle
    sm.TransitionTo(state.StateIdle)
    return result, nil
}
```

## Configuration

Future configuration options (via `config.yaml`):

```yaml
agent:
  state_machine:
    enable: true
    stuck_threshold: 3              # Same error N times = stuck
    oscillation_window: 4           # Check last N actions
    progress_check_interval: 10     # Check progress every N calls
    max_recent_actions: 20          # Track last N actions
    max_recent_errors: 10           # Track last N errors
```

## Benefits

1. **No False Positives**: Long productive tasks (1000 steps) work fine
2. **Reliable Loop Detection**: Catches stuck conditions in 3 errors
3. **Phase Awareness**: Multi-phase tasks reset tracking per phase
4. **Semantic Understanding**: Tracks actual progress, not just counts
5. **Observable**: State transitions logged, callbacks for monitoring
6. **Testable**: Comprehensive test suite with real-world scenarios

## Future Enhancements

- [ ] Resource monitoring integration (CPU, memory, disk)
- [ ] Per-tool timeout configuration
- [ ] Automatic recovery strategies per error type
- [ ] State machine visualization in TUI
- [ ] Metrics export (Prometheus, etc.)
- [ ] State persistence across restarts
