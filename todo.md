# Nexora TODO & Refactoring Plan

## 🎯 COMPLETED TODAY - December 18, 2025

### ✅ Option 2: CRUSH 	 NEXORA Cleanup (10 minutes)
**Status**: ✅ **COMPLETE**

- ✅ Fixed AGENTS.md (4 lines: CRUSH.md 	 NEXORA.md)
- ✅ Fixed Taskfile.yaml (1 line: CRUSH_PROFILE 	 NEXORA_PROFILE)
- ✅ All 4 tests passing in `internal/config/crush_test.go`

### ✅ Option 1: State Machine Integration (2-3 hours)
**Status**: ✅ **COMPLETE**

**Files Modified**:
- `internal/agent/agent.go`:
  - ✅ Added state machine import
  - ✅ Added `stateMachines` field to sessionAgent
  - ✅ Created `getOrCreateStateMachine()` helper
  - ✅ State transition on Run() start: 	 StateProcessingPrompt
  - ✅ Tool call tracking in OnToolResult callback
  - ✅ Loop detection with automatic halt
  - ✅ System message on stuck condition

**Result**: State machine now actively prevents conversation loops!

### ✅ Week 2: Resource Monitoring (3-4 hours)
**Status**: ✅ **COMPLETE** - December 18, 2025

**Files Created**:
```
internal/resources/
├── monitor.go       # Resource monitoring engine (318 lines) ✅
├── types.go         # Data structures (61 lines) ✅
└── monitor_test.go  # Comprehensive tests (249 lines) ✅
Total: 628 lines
```

**Features Implemented**:
- ✅ CPU usage monitoring (using gopsutil/v3)
- ✅ Memory usage monitoring (using gopsutil/v3)
- ✅ Disk space monitoring (using gopsutil/v3)
- ✅ Configurable thresholds (CPU 80%, Memory 85%, Disk 5GB)
- ✅ Violation tracking (last 20 violations with timestamps)
- ✅ Callbacks: OnCPUHigh, OnMemHigh, OnDiskLow, OnViolation
- ✅ State machine integration (auto-pause on max violations)
- ✅ Thread-safe with mutex protection
- ✅ Start/Stop lifecycle management
- ✅ Added `StateResourcePaused` to state machine

**Test Results**: 
```bash
✅ go test ./internal/agent/state/...    # 11/11 tests passing
✅ go test ./internal/resources/...      # 9/9 tests passing
✅ go test ./internal/config/...         # All tests passing
✅ go build ./...                        # Build successful
```

**How It Works**:
```
Monitor.Start(ctx)
  ↓
Every 5 seconds:
  - Check CPU usage
  - Check Memory usage
  - Check Disk space
  ↓
If threshold exceeded:
  - Log violation
  - Call callbacks
  - Add to violation history
  ↓
If max violations (3) reached:
  - Transition state machine to StateResourcePaused
  - Agent pauses execution
```

---

## 🚨 NEXT PRIORITIES

### Week 1: State Machine Architecture ✅ **COMPLETE** - December 18, 2025
**Goal**: Replace ad-hoc agent flow with explicit state machine

**Files Created**:
```
internal/agent/state/
├── machine.go          # Core state machine ✅
├── progress.go         # Progress tracking & loop detection ✅
├── phase.go            # Multi-phase task management ✅
├── states.go           # State constants and transitions ✅
├── machine_test.go     # Comprehensive tests (11 tests, 100% coverage) ✅
└── README.md           # Full documentation ✅
```

**State Definitions**:
```go
type AgentState int

const (
    StateIdle AgentState = iota
    StateProcessingPrompt
    StateStreamingResponse
    StateExecutingTool
    StateAwaitingPermission
    StateErrorRecovery
    StateResourcePaused
    StatePanicRecovery
    StateHalted
)

type AgentExecutionContext struct {
    State           AgentState
    SessionID       string
    CancelFunc      context.CancelFunc
    StartTime       time.Time
    ToolCallCount   int
    ErrorCount      int
    RetryCount      int
    LastError       error
    
    // Resource tracking
    CPUUsage        float64
    MemoryUsage     uint64
    DiskFree        uint64
}
```

**Integration Points**:
- Modify `internal/agent/agent.go` - Replace Run() flow with state machine
- Modify `internal/agent/coordinator.go` - Initialize state machine per session
- Add state transition logging to slog

**Tasks**:
- [x] Create state package structure ✅
- [x] Implement StateMachine with mutex-protected transitions ✅
- [x] Define valid transition map (9 states, validated transitions) ✅
- [x] Add state transition callbacks (OnStateChange, OnStuck, OnProgress) ✅
- [x] Add comprehensive state transition tests (11 tests, all passing) ✅
- [x] Document state machine with diagram ✅
- [ ] Integrate into sessionAgent.Run() (NEXT STEP)

**What Was Built**:
- ✅ Progress-based loop detection (not aggressive circuit breakers)
- ✅ Phase-aware tracking (resets between phases)
- ✅ 3-layer loop detection (same error, oscillation, no progress)
- ✅ Message deduplication
- ✅ 100% test coverage with race detector
- ✅ Allows 1000+ step productive tasks
- ✅ Halts on 3 identical errors

**Test Results**: PASS (1.056s) - 11/11 tests passing, no race conditions

**Documentation**: See `internal/agent/state/README.md` and `codedocs/STATE_MACHINE_IMPLEMENTATION_COMPLETE.md`

---

### Week 2: Resource Monitoring
**Goal**: Prevent resource exhaustion with active monitoring

**Files to Create**:
```
internal/resources/
├── monitor.go          # Resource monitoring goroutine
├── limits.go           # Threshold configuration
├── metrics.go          # Resource usage collection
└── guard.go            # Circuit breaker integration
```

**Implementation**:
```go
type ResourceMonitor struct {
    cpuThreshold    float64          // 80%
    memThreshold    uint64           // 85% of available
    diskThreshold   uint64           // 5GB minimum
    checkInterval   time.Duration    // 5 seconds
    
    // State
    stateMachine    *state.StateMachine
    violations      []Violation
    
    // Callbacks
    onCPUHigh       func(usage float64)
    onMemHigh       func(usage uint64)
    onDiskLow       func(free uint64)
}

func (rm *ResourceMonitor) Start(ctx context.Context)
func (rm *ResourceMonitor) Stop()
func (rm *ResourceMonitor) CurrentUsage() ResourceSnapshot
```

**Integration Points**:
- Launch monitor goroutine in coordinator.NewCoordinator()
- Pass state machine reference to monitor
- Monitor triggers state transitions (PAUSED, HALTED)
- Add metrics to Prometheus (if enabled)

**Tasks**:
- [ ] Implement CPU usage detection (using gopsutil or /proc/stat)
- [ ] Implement memory usage detection (runtime.MemStats + system)
- [ ] Implement disk space detection (syscall.Statfs)
- [ ] Create violation history (last 10 violations)
- [ ] Integrate with state machine for auto-pause
- [ ] Add configuration via config.yaml and env vars
- [ ] Test with artificial resource exhaustion
- [ ] Add graceful resume when resources available

---

### Week 3: Error Recovery System
**Goal**: Automatic recovery from transient errors

**Files to Create**:
```
internal/agent/recovery/
├── registry.go         # Maps errors to strategies
├── strategy.go         # RecoveryStrategy interface
├── file_outdated.go    # Re-read file recovery
├── edit_failed.go      # AIOPS-powered edit retry
├── loop_detected.go    # Break loop, return control
├── timeout.go          # Timeout handling
└── panic.go            # Panic recovery strategy
```

**Recovery Strategy Interface**:
```go
type RecoveryStrategy interface {
    CanRecover(err error) bool
    Recover(ctx context.Context, err error, execCtx *state.AgentExecutionContext) error
    MaxRetries() int
}

type RecoveryRegistry struct {
    strategies []RecoveryStrategy
    maxAttempts int  // Global limit (default: 3)
}
```

**Error Types to Handle**:
1. **FileOutdatedError** 	 Re-read file, retry tool call
2. **EditFailedError** 	 Use AIOPS to fix edit, retry
3. **LoopDetectedError** 	 Stop iteration, return to user
4. **TimeoutError** 	 Cancel, log, return to user
5. **ResourceLimitError** 	 Pause, wait, resume
6. **PanicError** 	 Recover, log stack trace, safe shutdown

**Integration Points**:
- Add recovery registry to sessionAgent
- Call recovery on error in StateErrorRecovery
- Track retry counts in AgentExecutionContext
- Log recovery attempts and outcomes

**Tasks**:
- [ ] Define typed error hierarchy (wrap existing errors)
- [ ] Implement RecoveryStrategy interface
- [ ] Create file outdated recovery (simple re-read)
- [ ] Create edit failed recovery (integrate with existing AIOPS)
- [ ] Create loop detection recovery (stop + user message)
- [ ] Implement panic recovery with defer/recover
- [ ] Add retry limit enforcement (max 3 per error type)
- [ ] Test each recovery strategy independently
- [ ] Integration test: trigger errors and verify recovery

---

### Week 4: Per-Tool Timeouts & Polish
**Goal**: Fine-grained timeout control, prevent tool hangs

**Tool Interface Extension**:
```go
type Tool interface {
    Name() string
    Execute(ctx context.Context, params any) (any, error)
    
    // NEW
    DefaultTimeout() time.Duration
    MaxRetries() int
    RequiresPermission() bool
    IsLongRunning() bool  // NEW: If true, run in background
}

// Updated timeout defaults
var DefaultTimeouts = map[string]time.Duration{
    "bash":   1 * time.Minute,   // CHANGED: Quick commands only
    "view":   1 * time.Minute,   // CHANGED: Large file support
    "edit":   30 * time.Second,
    "glob":   5 * time.Second,
    "grep":   30 * time.Second,
    "mcp":    2 * time.Minute,
    "write":  30 * time.Second,
}

// Long-running commands go to background automatically
var BackgroundCommands = []string{
    "npm start", "npm run dev", "go run", "python -m",
    "docker-compose up", "make watch", "yarn dev",
}
```

**Background Process Management**:
```go
// When bash tool detects long-running command:
func (t *BashTool) Execute(ctx context.Context, params BashParams) (any, error) {
    if isLongRunning(params.Command) {
        // Send to existing BackgroundShellManager
        jobID, err := backgroundManager.Start(ctx, params.Command)
        return JobStartedResponse{
            JobID: jobID,
            Message: "Command started in background. Use job_output to check status.",
        }, nil
    }
    
    // Regular command with 1-minute timeout
    ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
    defer cancel()
    
    return executeSync(ctx, params.Command)
}
```

**Tasks**:
- [ ] Change bash default timeout to 1 minute
- [ ] Change view default timeout to 1 minute
- [ ] Add IsLongRunning detection to bash tool
- [ ] Auto-route long commands to BackgroundShellManager
- [ ] Update tool documentation with new timeouts
- [ ] Test with intentionally slow commands

---

## P0: Install Octofriend as Supervisor

### Goal: Bolt octofriend into Nexora as oversight/fallback layer

**Architecture**:
```
┌─────────────────────────────────────────────────────────────────┐
│                         USER REQUEST                            │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
              ┌──────────────────────┐
              │   Nexora Agent       │ ◄──┐
              │   (Primary)          │    │
              └──────┬───────────────┘    │
                     │                     │
         ┌───────────┴───────────┐        │
         │                       │        │
         ▼                       ▼        │
    ┌─────────┐           ┌──────────┐   │
    │ Success │           │  Failure  │   │
    └─────────┘           └────┬─────┘   │
                               │          │
                               ▼          │
                     ┌──────────────────┐ │
                     │  Octofriend      │ │ Fallback
                     │  (Supervisor)    │─┘ on error
                     └──────────────────┘
                               │
                               ▼
                        ┌────────────┐
                        │  Success   │
                        └────────────┘
```

**Use Cases**:
1. **Primary Use**: Nexora handles all requests
2. **Fallback**: After 3 failed retries, hand off to Octofriend
3. **Second Opinion**: User can explicitly ask Octofriend via command
4. **Compare Results**: Run both, compare outputs (dev/debug mode)

**Implementation Plan**:

### Phase 1: Install Octofriend (Week 4)
**Tasks**:
- [ ] Install octofriend globally: `npm install -g octofriend`
- [ ] Add octofriend to Nexora's dependencies/tooling
- [ ] Create wrapper package: `internal/octofriend/`
- [ ] Test octofriend standalone execution

### Phase 2: Create Supervisor Bridge (Week 4)
**Files to Create**:
```
internal/supervisor/
├── octofriend.go       # Octofriend wrapper/client
├── bridge.go           # Bridge between Nexora <-> Octofriend
├── fallback.go         # Fallback logic (when to delegate)
└── compare.go          # Compare results (dev mode)
```

**Octofriend Wrapper**:
```go
type OctofriendClient struct {
    execPath    string  // Path to octofriend binary
    workingDir  string
    config      OctoConfig
}

type OctoConfig struct {
    Model       string  // Override model for octofriend
    APIKey      string  // Reuse Nexora's API keys
    DockerMode  bool    // Use Docker sandbox
}

func (oc *OctofriendClient) Execute(ctx context.Context, prompt string) (string, error) {
    // Execute octofriend with prompt
    // Capture output
    // Parse result
    // Return to Nexora
}

func (oc *OctofriendClient) ExecuteInDocker(ctx context.Context, prompt string, containerName string) (string, error) {
    // Use: octo docker connect <container>
    // Execute prompt
    // Return result
}
```

**Fallback Logic**:
```go
type FallbackStrategy struct {
    maxRetriesBeforeFallback int  // Default: 3
    enableComparison         bool // Dev mode: compare both results
}

func (fs *FallbackStrategy) ShouldFallback(execCtx *state.AgentExecutionContext) bool {
    // Check if Nexora has failed 3x on same task
    if execCtx.RetryCount >= fs.maxRetriesBeforeFallback {
        return true
    }
    
    // Check if error type suggests octofriend might do better
    // e.g., complex file editing, diff application
    if isComplexEditError(execCtx.LastError) {
        return true
    }
    
    return false
}

func (fs *FallbackStrategy) ExecuteWithFallback(
    ctx context.Context,
    nexoraAgent Agent,
    octoClient *OctofriendClient,
    prompt string,
) (Result, error) {
    // Try Nexora first
    result, err := nexoraAgent.Execute(ctx, prompt)
    if err == nil {
        return result, nil
    }
    
    // Fallback to Octofriend
    slog.Info("Falling back to Octofriend supervisor", "reason", err)
    return octoClient.Execute(ctx, prompt)
}
```

### Phase 3: Integration Points (Week 4)

**In State Machine**:
```go
// Add new state
const (
    // ... existing states
    StateFallbackToSupervisor  // NEW
)

// In error recovery
func (sm *StateMachine) EnterErrorRecovery(err error) {
    if sm.context.RetryCount >= 3 {
        // Hand off to octofriend
        sm.Transition(StateFallbackToSupervisor)
        return
    }
    
    // Normal recovery
    // ...
}
```

**In Coordinator**:
```go
type coordinator struct {
    // ... existing fields
    octofriend *supervisor.OctofriendClient  // NEW
    fallbackStrategy *supervisor.FallbackStrategy  // NEW
}

func (c *coordinator) Run(ctx context.Context, sessionID, prompt string) (*Result, error) {
    // Check if we should use fallback
    if c.fallbackStrategy.ShouldFallback(execContext) {
        return c.octofriend.Execute(ctx, prompt)
    }
    
    // Normal Nexora execution
    return c.currentAgent.Run(ctx, call)
}
```

**Configuration**:
```yaml
supervisor:
  enabled: true
  type: "octofriend"  # Future: could add other supervisors
  fallback:
    enable: true
    max_retries_before_fallback: 3
    prefer_for_tasks:
      - "complex_edits"
      - "diff_application"
      - "json_repair"
  octofriend:
    exec_path: "/usr/local/bin/octofriend"  # or auto-detect
    model_override: null  # null = use Nexora's model
    docker_mode: false
    autofix_models:
      enable: true  # Use octofriend's custom autofix models
```

### Phase 4: User Commands & TUI Switching (Week 4)

**Command Options**:
```go
// In chat, user can type:

// Option 1: Switch to Octofriend TUI (RECOMMENDED)
"/octo"                    // Exit Nexora, launch octofriend TUI in same terminal
"/octo <prompt>"           // Switch to octo TUI with initial prompt

// Option 2: Use Octofriend programmatically (headless)
"/octo-exec <prompt>"      // Run octofriend headless, return results to Nexora

// Other commands:
"/compare <prompt>"        // Run both, show comparison
"/fallback enable"         // Enable fallback mode
"/fallback disable"        // Disable fallback mode
"/supervisor status"       // Show supervisor status
"/back"                    // Return to Nexora from octofriend (if applicable)
```

**Implementation: TUI Switching**:

```go
// internal/tui/commands/octo.go
func HandleOctoCommand(app *App, prompt string) error {
    // 1. Save Nexora state
    app.SaveSession()
    
    // 2. Cleanup Nexora TUI
    app.Program.Quit()
    app.Cleanup()
    
    // 3. Execute octofriend in same terminal
    cmd := exec.Command("octofriend")
    if prompt != "" {
        // Pass initial prompt via stdin or as argument
        cmd.Stdin = strings.NewReader(prompt)
    }
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    cmd.Stdin = os.Stdin
    
    // Give user full control of terminal
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("octofriend failed: %w", err)
    }
    
    // 4. After octofriend exits, user can restart nexora manually
    //    OR auto-restart nexora (optional)
    fmt.Println("\n✨ Octofriend session ended.")
    fmt.Println("💡 Run 'nexora' to return to Nexora")
    
    return nil
}
```

**Alternative: Side-by-side (tmux/screen integration)**:
```go
func HandleOctoSideBySide(app *App, prompt string) error {
    // If running in tmux:
    // 1. Split pane vertically
    // 2. Launch octofriend in new pane
    // 3. Keep Nexora running in original pane
    
    if !isInTmux() {
        return errors.New("Side-by-side requires tmux")
    }
    
    cmd := exec.Command("tmux", "split-window", "-h", "octofriend")
    return cmd.Run()
}
```

**User Experience Flow**:
```
[In Nexora TUI]
User: /octo debug the webpack config

[Nexora saves session and quits]
💾 Session saved
🔄 Switching to Octofriend...

[Octofriend TUI launches in same terminal]
Octo> debug the webpack config
[... octofriend does its work ...]

[User exits octofriend: Ctrl+C or normal quit]
✨ Octofriend session ended.
💡 Run 'nexora' to return to Nexora

[User runs: nexora]
[Nexora TUI launches, restores previous session]
✅ Session restored
Welcome back!
```

**Configuration**:
```yaml
supervisor:
  octofriend:
    switch_behavior: "replace"  # "replace", "sidebyside", "exec"
    auto_return: false           # Auto-restart nexora after octo exits
    save_session_on_switch: true
    share_working_dir: true      # Same working directory
    share_api_keys: true         # Use Nexora's API keys
```

**Tasks**:
- [ ] Install octofriend globally
- [ ] Create `internal/supervisor/` package
- [ ] Implement OctofriendClient wrapper (headless mode)
- [ ] Implement TUI switching logic
- [ ] Add /octo command handler
- [ ] Session save/restore on switch
- [ ] Detect tmux/screen for side-by-side option
- [ ] Add fallback strategy for programmatic use
- [ ] Add StateFallbackToSupervisor
- [ ] Integrate into coordinator
- [ ] Add configuration support
- [ ] Test switching flow
- [ ] Add logging/metrics for fallback usage
- [ ] Document all command options

### Phase 5: Advanced Features (Future)

**Docker Sandbox Integration**:
```go
// Octofriend can run in Docker containers
// Nexora can delegate risky operations to sandboxed octofriend
func (oc *OctofriendClient) ExecuteInSandbox(ctx context.Context, prompt string) (string, error) {
    // Start fresh alpine container
    // Run octofriend inside
    // Execute prompt
    // Capture results
    // Destroy container
}
```

**Comparison Mode** (Dev/Testing):
```go
func (fs *FallbackStrategy) ExecuteWithComparison(
    ctx context.Context,
    nexoraAgent Agent,
    octoClient *OctofriendClient,
    prompt string,
) (ComparisonResult, error) {
    var wg sync.WaitGroup
    var nexoraResult, octoResult Result
    var nexoraErr, octoErr error
    
    // Run both in parallel
    wg.Add(2)
    go func() {
        nexoraResult, nexoraErr = nexoraAgent.Execute(ctx, prompt)
        wg.Done()
    }()
    go func() {
        octoResult, octoErr = octoClient.Execute(ctx, prompt)
        wg.Done()
    }()
    wg.Wait()
    
    // Compare and return both results
    return ComparisonResult{
        Nexora:    nexoraResult,
        Octofriend: octoResult,
        BothSucceeded: nexoraErr == nil && octoErr == nil,
        Differences: diffResults(nexoraResult, octoResult),
    }, nil
}
```

**Metrics**:
```go
type SupervisorMetrics struct {
    FallbackCount       int
    FallbackSuccessRate float64
    TasksImproved       []string  // Tasks where octofriend did better
    TasksWorse          []string  // Tasks where octofriend did worse
}
```

---



**Implementation in Coordinator**:
```go
func (c *coordinator) executeToolWithTimeout(
    ctx context.Context, 
    tool Tool, 
    params any,
) (any, error) {
    timeout := tool.DefaultTimeout()
    if timeout == 0 {
        timeout = DefaultTimeouts[tool.Name()]
    }
    
    ctx, cancel := context.WithTimeout(ctx, timeout)
    defer cancel()
    
    result, err := tool.Execute(ctx, params)
    if errors.Is(err, context.DeadlineExceeded) {
        return nil, &TimeoutError{
            Tool:    tool.Name(),
            Timeout: timeout,
        }
    }
    return result, err
}
```

**Tasks**:
- [ ] Add timeout methods to all tool implementations
- [ ] Update tool execution to use timeout wrapper
- [ ] Make timeouts configurable per tool in config.yaml
- [ ] Add timeout warnings (e.g., 80% elapsed)
- [ ] Test with intentionally slow tools
- [ ] Document timeout configuration

---

## Configuration

**New config.yaml section**:
```yaml
agent:
  state_machine:
    enable_recovery: true
    max_recovery_attempts: 3
    recovery_timeout: 30s
    log_transitions: true
  
  resources:
    enable_monitoring: true
    cpu_threshold: 80.0       # percent
    memory_threshold: 85.0     # percent
    disk_min_free: 5368709120  # 5GB in bytes
    check_interval: 5s
  
  limits:
    max_tool_calls: 100
    max_session_duration: 30m
    max_retries_per_tool: 3
  
  timeouts:
    default: 5m
    bash: 10m
    edit: 30s
    view: 10s
    glob: 5s
    grep: 30s
    mcp: 2m
    write: 30s
```

**Environment Variable Overrides**:
```bash
NEXORA_STATE_MACHINE_ENABLE_RECOVERY=true
NEXORA_RESOURCES_CPU_THRESHOLD=80
NEXORA_RESOURCES_MEMORY_THRESHOLD=85
NEXORA_RESOURCES_DISK_MIN_FREE=5GB
NEXORA_MAX_TOOL_CALLS=100
NEXORA_MAX_SESSION_DURATION=30m
NEXORA_TIMEOUT_BASH=10m
NEXORA_TIMEOUT_EDIT=30s
```

---

## Quick Fixes (Parallel Work)

### Fix TODOs and FIXMEs
- [ ] `internal/agent/coordinator.go:105` - Make agent selection dynamic
- [ ] `internal/agent/coordinator.go:258` - Dynamic model config per agent
- [ ] `internal/agent/coordinator.go:517-518` - Enhance agent execution
- [ ] `internal/sessionlog/integration.go:78` - Read from environment
- [ ] `internal/agent/native/bash.go:35,107` - Implement proper context handling

### Replace context.TODO()
- All instances should be replaced with proper context propagation
- Add timeout contexts where long operations occur
- Pattern: `ctx, cancel := context.WithTimeout(parent, duration)`

### Add Panic Recovery
Add to critical paths:
```go
defer func() {
    if r := recover(); r != nil {
        err = fmt.Errorf("panic recovered: %v\nstack: %s", r, debug.Stack())
        slog.Error("Tool execution panic", "panic", r, "stack", string(debug.Stack()))
        // Transition to StatePanicRecovery
    }
}()
```

Locations:
- Tool execution callbacks (OnToolCall)
- LLM streaming (agent.Stream)
- Message creation/update operations

---

## TUI Enhancement: Debug/Monitoring Window (P1)

### Feature: Control-Key Accessible Debug Window
**Goal**: Add a toggle-able debug/monitoring window accessible via Ctrl+D

**Motivation**: 
- Need to see tool output (bash, logs, etc.) without interrupting chat flow
- Want to monitor state machine transitions in real-time
- Need visibility into resource usage (CPU/memory/disk)
- Current flow: tool output appears inline, gets buried in chat history

**Design**:
```
┌─────────────────────────────────────────────────────────────┐
│ Nexora Chat (Main Window)                                   │
│                                                              │
│ User: read the file                                         │
│ Assistant: I'll read that for you...                        │
│ [Tool: view /path/to/file]                                  │
│                                                              │
│ > _                                    Press Ctrl+D for Debug│
└─────────────────────────────────────────────────────────────┘

[User presses Ctrl+D]

┌─────────────────────────────────────────────────────────────┐
│ Chat (60% width)                │ Debug Window (40% width)  │
│                                 │                           │
│ User: read the file             │ STATE: ExecutingTool      │
│ Assistant: I'll read...         │ Tool: view                │
│ [Tool: view /path/to/file]      │ Duration: 0.5s            │
│                                 │                           │
│                                 │ Resources:                │
│                                 │  CPU: 12%                 │
│                                 │  Mem: 847MB               │
│                                 │  Disk: 102GB free         │
│                                 │                           │
│                                 │ Tool Output:              │
│                                 │ 1| package main          │
│                                 │ 2| import "fmt"          │
│ > _                             │ 3| ...                   │
│                                 │                           │
│                    Ctrl+D Close │ [scroll with ↑/↓]        │
└─────────────────────────────────┴───────────────────────────┘
```

**Implementation**:

**Files to Create**:
```
internal/tui/components/debug/
├── window.go           # Main debug window component
├── state_panel.go      # State machine display
├── resources_panel.go  # Resource usage display
├── output_panel.go     # Tool output display
└── tabs.go             # Tab switching (State/Resources/Output/Logs)
```

**Integration**:
- Modify `internal/tui/app.go` - Add debug window toggle
- Add keybinding: Ctrl+D toggles debug window
- Connect to state machine for real-time updates
- Connect to resource monitor for metrics
- Buffer tool outputs for display

**Features**:
1. **State Panel** (default tab)
   - Current agent state with color coding
   - State transition history (last 10)
   - Tool call count, error count, retry count
   - Time in current state

2. **Resource Panel**
   - Real-time CPU/Memory/Disk usage
   - Historical sparkline graphs (last 60s)
   - Threshold indicators (yellow warning, red breach)
   - Watchdog status (active/triggered)

3. **Output Panel**
   - Live tool output streaming
   - Scrollable with vim keys (j/k or ↑/↓)
   - Searchable (/)
   - Copy-able (y to yank)

4. **Logs Panel**
   - Live slog output
   - Filter by level (Debug/Info/Warn/Error)
   - Searchable

**Keybindings**:
- `Ctrl+D` - Toggle debug window
- `Tab` - Switch between tabs
- `j/k` or `↑/↓` - Scroll
- `/` - Search
- `y` - Copy current line
- `Esc` - Close debug window

**Tasks**:
- [ ] Create debug window package structure
- [ ] Implement state panel with lipgloss styling
- [ ] Implement resource panel with sparklines
- [ ] Implement output panel with scrolling
- [ ] Add keybinding handler in app.go
- [ ] Connect to state machine events
- [ ] Connect to resource monitor
- [ ] Add tool output buffering
- [ ] Test with various window sizes
- [ ] Add configuration (enable/disable, default tab, etc.)

**Configuration**:
```yaml
tui:
  debug_window:
    enabled: true
    default_open: false
    default_tab: "state"
    width_percent: 40
    max_history_lines: 1000
    refresh_interval: 100ms
```

---



### Current State
- Using `charm.land/fantasy` v0.5.1 for LLM abstraction
- Using `github.com/charmbracelet/catwalk` v0.9.5 for provider/model config
- Provider config in `internal/config/provider.go` (~420 lines)

Replace fantasy/catwalk with a unified provider abstraction that:
1. **Reduces coupling** - Single interface for all providers
2. **Adds provider support** - Gemini, Groq, Cohere, Perplexity, local llama.cpp, ollama
3. **Cleaner request/response translation** - One place for OpenAI↔Mistral↔Anthropic format conversions
4. **Better error handling** - Provider-specific error mapping
5. **Config management** - Single source of truth for models and their capabilities

### Implementation Phases

#### Phase 1: Create Provider Interface (LOW EFFORT, HIGH VALUE)
**Effort**: ~4-6 hours
**Files to create**:
- `internal/llm/provider.go` - Core interface (150-200 lines)
  - `Provider` interface with Chat/Completion/Embedding methods
  - `Model` struct with capabilities (context, cost, reasoning, etc)
  - `Request`/`Response` wrappers for normalized format
  - `ChatOptions` for streaming, tools, temperature, etc

- `internal/llm/client.go` - Provider registry (100-150 lines)
  - `ProviderRegistry` to load/switch providers at runtime
  - Discovery mechanism for available models
  - Cost tracking per model

**Backward Compat**: Wrap existing fantasy/catwalk, no breaking changes yet

#### Phase 2: Implement Provider Adapters (MEDIUM EFFORT)
**Effort**: ~8-12 hours (1-2 providers per hour once pattern is established)

**To Create**:
1. `internal/llm/providers/openai.go` - Direct OpenAI client
2. `internal/llm/providers/anthropic.go` - Anthropic API
3. `internal/llm/providers/mistral.go` - Mistral API
4. `internal/llm/providers/gemini.go` - Google Gemini
5. `internal/llm/providers/groq.go` - Groq (fast inference)
6. `internal/llm/providers/openai_compat.go` - OpenAI-compatible (Mistral local, vLLM, LM Studio)
7. `internal/llm/providers/fantasy_wrapper.go` - Wrap existing for migration

**Per adapter includes**:
- Request normalization (tools, messages, parameters)
- Response translation to standard format
- Error mapping
- Token counting
- Rate limit handling

#### Phase 3: Response Format Normalization (MEDIUM EFFORT)
**Effort**: ~6-8 hours

**Create**:
- `internal/llm/format/` directory
  - `openai.go` - OpenAI request/response types + translation
  - `anthropic.go` - Anthropic types + translation
  - `mistral.go` - Mistral types + translation
  - `gemini.go` - Gemini types + translation
  - `converter.go` - Unified conversion logic

**Why separate**: Each provider has different:
- Message/tool formats
- Streaming chunk structure
- Error codes/messages
- Token usage fields
- Model naming conventions

#### Phase 4: Update Config System (LOW-MEDIUM EFFORT)
**Effort**: ~4-6 hours

**Changes to `internal/config/`**:
- Replace catwalk model structs with unified `llm.Model`
- Keep provider.go for loading from disk/env
- Remove provider.go injection logic → use new registry
- Update provider_test.go with new interface

#### Phase 5: Update Agent/Coordinator (MEDIUM EFFORT)
**Effort**: ~6-8 hours

**Files to update**:
- `internal/agent/coordinator.go` - Switch to new provider interface
- `internal/agent/event.go` - Use unified usage tracking
- `internal/agent/common_test.go` - Update test fixtures

**What changes**:
- Remove fantasy imports
- Use new provider.Chat() method
- Standardized error handling
- Same business logic, cleaner implementation

#### Phase 6: Update Config/Provider Tests (LOW EFFORT)
**Effort**: ~2-3 hours

**Tests to update**:
- `internal/config/provider_test.go`
- `internal/config/provider_empty_test.go`
- `internal/config/recent_models_test.go`
- Add new provider adapter tests

### Total Effort Estimate
- **Aggressive** (dedicated): 25-35 hours (~1 week full-time)
- **Gradual** (part-time): 4-5 hours/week
- **With refactoring surprises**: Add 20-30%

### Benefits
- ✅ Smaller, more maintainable code
- ✅ Easier to add new providers (1-2 hours per provider)
- ✅ Better error handling per provider
- ✅ Unified testing across all providers
- ✅ Can deprecate fantasy/catwalk (remove 2 dependencies)
- ✅ Better performance (direct API calls vs abstraction)
- ✅ Easier to support bleeding-edge provider features

### Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Break existing agent flow | Wrap old code, migrate incrementally, test with existing VCR cassettes |
| Lose fantasy's abstraction benefits | Gain them back with cleaner design, fewer dependencies |
| Model metadata conflicts | Centralize in new `Model` struct, single source of truth |
| Token counting differences | Implement per-provider, validate against actual API usage |

### Dependencies to Add
- `google.golang.org/genai` - Gemini (or REST if preferred)
- `github.com/groq/groq-go` - Groq SDK (if available)
- Keep: `github.com/openai/openai-go`
- Keep: `github.com/charmbracelet/anthropic-sdk-go`
- Consider: `github.com/mistralai/client-go`

### Dependencies to Remove
- `charm.land/fantasy` - Full replacement
- `github.com/charmbracelet/catwalk` - Config only, can replace with JSON/YAML

---

## Core Reliability Fixes

### UI & Context Issues
- [ ] **UI cursor adjustments** - Fix FIXME comments in `splash.go` and `models.go`
- [ ] **Remaining context.TODO()** (20 instances) - Replace with proper contexts

---

## 🧪 TEST FAILURES (Fix Immediately)
### Database & Indexing
- [x] **SQLite FTS schema errors** - "no such column: fts" and "no such table: symbols_fts" in search queries
- [x] **Database migration concurrency** - Concurrent read/write failures during index operations
- [x] **Index table creation** - Missing FTS tables in new databases

### Command Line Interface
- [ ] **Nexora crash handling** - Generic crash messages instead of specific errors in index/query commands
- [ ] **Invalid path handling** - "no such file" messages not appearing for invalid paths
- [ ] **Bedrock credential validation** - Failing credential tests expecting 0/1 counts

---

## Known Issues

### Agent Loop with Devstral-2 (BLOCKER)
**Issue**: When reading large files like todo.md, nexora enters infinite loop with devstral-2
**Root Cause**: Agent is sending full system context (including all tool definitions) to LLM, causing devstral to repeatedly try tool invocations
**Solution**: Implement tool result aggregation and prevent re-prompting with same context
**Status**: NEEDS INVESTIGATION - may be devstral-2 tool handling, not nexora

### Streaming Response Conversion
**Issue**: Devstral proxy converts streaming→non-streaming internally, then converts back to SSE
**Better approach**: Support true streaming passthrough or use polling with proper chunking

---

## Other TODOs

### Bugs/Issues
- [ ] Agent loop prevention (max iterations, cycle detection)
- [ ] Provider connection timeouts not always handled gracefully
- [ ] Model cost tracking doesn't account for cache hits

### Features
- [ ] Add function calling to all providers (currently partial)
- [ ] Implement vision support across providers
- [ ] Add batch processing API
- [ ] Rate limiting per provider with backoff
- [ ] Provider health checks
- [ ] Fallback provider when primary fails

### Performance
- [ ] Cache model list on startup
- [ ] Connection pooling for HTTP clients
- [ ] Streaming response optimization
- [ ] Token counting optimization

### Testing
- [ ] Expand provider adapter tests to 90%+ coverage
- [ ] Add integration tests with real APIs (VCR cassettes)
- [ ] Load testing with multiple concurrent requests
- [ ] Error scenario coverage

### Documentation
- [ ] Add provider setup guides (API keys, regions, etc)
- [ ] Document new provider interface for contributors
- [ ] API compatibility matrix
- [ ] Migration guide from fantasy→new system

---

## Banned Commands

**DO NOT DO**:
- `pkill nexora` - Kills the running nexora process ungracefully
- `pkill go` - Kills all Go processes, including the development server and builds