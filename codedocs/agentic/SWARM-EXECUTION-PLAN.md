# Swarm Execution Plan: Nexora Headless Delegate

**Task:** Implement headless delegate system for Nexora
**Orchestrator:** Claude-Swarm (`/opt/claude-swarm`)
**Project:** `/home/nexora`

---

## Architecture Overview

### Delegation Pool Integration

The existing delegate system uses `delegation.NewPool()` with resource-aware spawning and queue management.
This plan introduces tmux-based execution as an **alternative executor** that plugs into the existing pool:

```
┌─────────────────────────────────────────────────────────────────┐
│                     Coordinator                                  │
│  ┌─────────────────┐    ┌──────────────────────────────────┐   │
│  │  delegateTool() │───▶│       delegation.Pool             │   │
│  └─────────────────┘    │  - Resource monitoring            │   │
│                         │  - Queue management               │   │
│                         │  - Concurrency limits             │   │
│                         └──────────────┬───────────────────┘   │
│                                        │                        │
│                         ┌──────────────┴───────────────────┐   │
│                         │        Pool.SetExecutor()         │   │
│                         └──────────────┬───────────────────┘   │
│                                        │                        │
│              ┌─────────────────────────┼─────────────────────┐ │
│              │                         │                     │ │
│              ▼                         ▼                     │ │
│   ┌──────────────────┐     ┌──────────────────────────┐     │ │
│   │ executeDelegated │     │ executeDelegatedTaskTmux │     │ │
│   │      Task()      │     │  (NEW - tmux spawn)      │     │ │
│   │  (inline agent)  │     └──────────────────────────┘     │ │
│   └──────────────────┘                                       │ │
│                                                              │ │
└──────────────────────────────────────────────────────────────┘
```

**Pool integration approach:**
1. Keep existing `delegation.Pool` for resource management
2. Add `executeDelegatedTaskTmux()` as new executor function
3. Configure pool to use tmux executor: `c.delegatePool.SetExecutor(c.executeDelegatedTaskTmux)`
4. Existing inline executor remains available as fallback

**Benefits:**
- Existing queue timeout, resource monitoring, and stats continue to work
- No changes to `delegateTool()` or permission handling
- Can switch between inline and tmux execution via config

---

## Audit Concerns

The following concerns were identified during plan review and must be addressed:

### Concern 11: Configuration Integration

**Issue:** The plan assumed `CoordinatorConfig` but coordinator uses `*config.Config`.

**Resolution:** Add headless fields to `Options` struct in `internal/config/config.go`:

```go
type Options struct {
    // ... existing fields

    // Headless mode configuration
    HeadlessMode    bool   `json:"headless_mode,omitempty" jsonschema:"description=Run without TUI"`
    PromptFile      string `json:"prompt_file,omitempty" jsonschema:"description=Read prompt from file"`
    ContextFile     string `json:"context_file,omitempty" jsonschema:"description=Read context from file"`
    OutputFile      string `json:"output_file,omitempty" jsonschema:"description=Write output to file"`
    StatusFile      string `json:"status_file,omitempty" jsonschema:"description=Write status updates to file"`
    OutputFormat    string `json:"output_format,omitempty" jsonschema:"description=Headless output format (text/json/minimal)"`

    // Delegate mode fields (set when running as spawned delegate)
    ParentSession   string `json:"parent_session,omitempty" jsonschema:"description=Parent session ID"`
    TaskID          string `json:"task_id,omitempty" jsonschema:"description=Delegate task ID"`
}
```

**Impact:** Update F5 to modify `internal/config/config.go` instead of coordinator.

---

### Concern 12: CLI Flag Propagation

**Issue:** Flags need to propagate through the application stack.

**Resolution:** Add propagation chain:

```go
// internal/cmd/root.go
var (
    headlessMode   bool
    promptFile     string
    outputFile     string
    // ...
)

func init() {
    rootCmd.PersistentFlags().BoolVar(&headlessMode, "headless", false, "Run without TUI")
    // ... other flags
}

// In run command, propagate to config
func runNexora(cmd *cobra.Command, args []string) error {
    cfg, err := config.Load()
    if err != nil {
        return err
    }

    // CLI flags override config file settings
    if headlessMode {
        cfg.Options.HeadlessMode = true
    }
    if promptFile != "" {
        cfg.Options.PromptFile = promptFile
    }
    // ... propagate all flags

    // Pass to coordinator
    coord := agent.NewCoordinator(ctx, cfg, ...)
}
```

**Impact:** F1-F4 must include propagation logic, not just flag registration.

---

### Concern 13: Tool Access Parity

**Issue:** Headless delegates must have same tool access as main agent.

**Resolution:** Ensure environment inheritance in tmux spawn:

```go
func (c *coordinator) executeDelegatedTaskTmux(ctx context.Context, task *delegation.Task) (string, error) {
    // Build environment for delegate
    env := os.Environ()

    // Ensure MCP configuration is available
    mcpConfigPath := filepath.Join(os.Getenv("HOME"), ".config", "nexora", "mcp.json")
    if _, err := os.Stat(mcpConfigPath); err == nil {
        // MCP servers will be discovered from config
    }

    // Build command with inherited environment
    nexoraCmd := fmt.Sprintf(
        "env %s nexora --headless --prompt-file=%s ...",
        strings.Join(criticalEnvVars(env), " "),
        promptPath,
    )
    // ...
}

func criticalEnvVars(env []string) []string {
    critical := []string{"HOME", "PATH", "NEXORA_", "ANTHROPIC_", "OPENAI_"}
    var result []string
    for _, e := range env {
        for _, prefix := range critical {
            if strings.HasPrefix(e, prefix) || strings.HasPrefix(e, prefix+"=") {
                result = append(result, e)
                break
            }
        }
    }
    return result
}
```

**Impact:** Add to F9 implementation. Add acceptance criterion for MCP server availability.

---

### Concern 14: Permission Model Safety

**Issue:** Auto-approving permissions may be unsafe for some tasks.

**Resolution:** Add `--require-confirmation` flag:

```go
// CLI flag
rootCmd.PersistentFlags().BoolVar(&requireConfirmation, "require-confirmation", false,
    "Require confirmation for destructive operations even in headless mode")

// In coordinator
func (c *coordinator) RunHeadless(ctx context.Context) error {
    if !c.cfg.Options.RequireConfirmation {
        c.permissions.AutoApproveSession(session.ID)
    }
    // Otherwise, permissions will fail without human interaction
    // which is appropriate for safety-critical automation
}
```

**Impact:** Add to F1 and F6. Consider adding `--allowed-tools` whitelist flag.

---

### Concern 15: Resource Exhaustion Limits

**Issue:** Need configurable limits for concurrent headless delegates.

**Resolution:** Add to Options and pool configuration:

```go
type Options struct {
    // ...
    MaxConcurrentDelegates int `json:"max_concurrent_delegates,omitempty" jsonschema:"description=Maximum concurrent delegate processes,default=5"`
}

// In coordinator initialization
func (c *coordinator) initDelegatePool() {
    maxDelegates := c.cfg.Options.MaxConcurrentDelegates
    if maxDelegates <= 0 {
        maxDelegates = 5 // sensible default
    }

    poolConfig := delegation.PoolConfig{
        MaxWorkers: maxDelegates,
        // ...
    }
    c.delegatePool = delegation.NewPool(poolConfig, c.resourceMonitor)
}
```

**Impact:** Add to F5 config fields. Update pool initialization.

---

### Concern 16: Configurable Cleanup Retention

**Issue:** 1-hour cleanup delay should be configurable.

**Resolution:** Add retention option:

```go
type Options struct {
    // ...
    DelegateRetentionMinutes int `json:"delegate_retention_minutes,omitempty" jsonschema:"description=Minutes to retain delegate directories for debugging,default=60"`
}

func (c *coordinator) cleanupDelegateDir(delegateDir string) {
    retention := time.Duration(c.cfg.Options.DelegateRetentionMinutes) * time.Minute
    if retention <= 0 {
        retention = 1 * time.Hour
    }

    go func() {
        time.Sleep(retention)
        os.RemoveAll(delegateDir)
    }()
}
```

**Impact:** Update F14 cleanup logic.

---

### Concern 17: Monitor Error Handling and Backoff

**Issue:** Monitor needs SIGTERM handling and should use exponential backoff.

**Resolution:** Update F10 with:

```go
func (c *coordinator) monitorDelegate(
    ctx context.Context,
    task *delegation.Task,
    tmuxSessionID string,
    delegateDir string,
) {
    // Initial poll interval
    pollInterval := 2 * time.Second
    maxInterval := 30 * time.Second
    consecutiveEmpty := 0

    donePath := filepath.Join(delegateDir, "done")
    outputPath := filepath.Join(delegateDir, "output.txt")

    // Handle SIGTERM for graceful shutdown
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
    defer signal.Stop(sigChan)

    ticker := time.NewTicker(pollInterval)
    defer ticker.Stop()

    for {
        select {
        case <-sigChan:
            slog.Info("delegate monitor received shutdown signal", "task_id", task.ID)
            c.gracefulDelegateShutdown(task, tmuxSessionID, delegateDir)
            return

        case <-ctx.Done():
            // ... existing handling

        case <-ticker.C:
            if _, err := os.Stat(donePath); err == nil {
                // ... completion handling
                return
            }

            // Exponential backoff when idle
            consecutiveEmpty++
            if consecutiveEmpty > 5 && pollInterval < maxInterval {
                pollInterval = min(pollInterval*2, maxInterval)
                ticker.Reset(pollInterval)
            }
        }
    }
}
```

**Impact:** Update F10 implementation.

---

### Concern 18: Move --output-format to Phase 1

**Issue:** Output format flag needed early for F7.

**Resolution:** Update Phase 1 feature list:

| ID | Feature | Dependencies |
|----|---------|--------------|
| **F1** | Add `--headless` flag | None |
| **F2** | Add `--prompt-file` flag | None |
| **F3** | Add `--output-file` flag | None |
| **F4** | Add `--model` flag | None |
| **F4b** | Add `--output-format` flag | None |

**Impact:** Update feature decomposition table.

---

### Concern 19: Extract Atomic File Utilities

**Issue:** Multiple features need atomic file operations.

**Resolution:** Create `internal/fsutil/atomic.go`:

```go
package fsutil

import (
    "os"
    "path/filepath"
)

// WriteAtomic writes content atomically via temp file + rename
func WriteAtomic(path string, content []byte, perm os.FileMode) error {
    dir := filepath.Dir(path)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return err
    }

    tmp, err := os.CreateTemp(dir, ".tmp-*")
    if err != nil {
        return err
    }
    tmpPath := tmp.Name()
    defer func() {
        tmp.Close()
        os.Remove(tmpPath)
    }()

    if _, err := tmp.Write(content); err != nil {
        return err
    }
    if err := tmp.Sync(); err != nil {
        return err
    }
    if err := tmp.Close(); err != nil {
        return err
    }

    return os.Rename(tmpPath, path)
}

// ReadIfExists reads a file, returning nil if it doesn't exist
func ReadIfExists(path string) ([]byte, error) {
    data, err := os.ReadFile(path)
    if os.IsNotExist(err) {
        return nil, nil
    }
    return data, err
}
```

**Impact:** Add new feature F0 or integrate into F8. Update F8, F11, F12, F13 to use `fsutil.WriteAtomic()`.

---

### Concern 20: Configuration Validation in RunHeadless

**Issue:** Need to validate file paths before execution.

**Resolution:** Add validation function:

```go
func (c *coordinator) validateHeadlessConfig() error {
    opts := c.cfg.Options

    // Prompt file must exist and be readable
    if opts.PromptFile == "" {
        return fmt.Errorf("--prompt-file is required in headless mode")
    }
    if _, err := os.Stat(opts.PromptFile); err != nil {
        return fmt.Errorf("prompt file not accessible: %w", err)
    }

    // Output directory must be writable
    if opts.OutputFile != "" {
        dir := filepath.Dir(opts.OutputFile)
        if err := os.MkdirAll(dir, 0755); err != nil {
            return fmt.Errorf("cannot create output directory: %w", err)
        }
        // Test write access
        testFile := filepath.Join(dir, ".write-test")
        if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
            return fmt.Errorf("output directory not writable: %w", err)
        }
        os.Remove(testFile)
    }

    // Context file is optional but must be readable if specified
    if opts.ContextFile != "" {
        if _, err := os.Stat(opts.ContextFile); err != nil {
            return fmt.Errorf("context file not accessible: %w", err)
        }
    }

    return nil
}

func (c *coordinator) RunHeadless(ctx context.Context) error {
    if err := c.validateHeadlessConfig(); err != nil {
        return err
    }
    // ... rest of implementation
}
```

**Impact:** Add to F6 implementation.

---

### Concern 21: Health Checks in monitorDelegate

**Issue:** Need periodic health checks for delegates.

**Resolution:** Add health check logic to F10:

```go
func (c *coordinator) checkDelegateHealth(tmuxSessionID, delegateDir string) error {
    tmuxMgr := shell.GetTmuxManager()

    // Check tmux session is responsive
    if !tmuxMgr.IsSessionRunning(tmuxSessionID) {
        return fmt.Errorf("tmux session no longer running")
    }

    // Check status file is being updated (stale = >5 min without update)
    statusPath := filepath.Join(delegateDir, "status.json")
    info, err := os.Stat(statusPath)
    if err == nil {
        if time.Since(info.ModTime()) > 5*time.Minute {
            return fmt.Errorf("delegate appears stalled (no status update in 5 min)")
        }
    }

    // Check disk space
    var stat syscall.Statfs_t
    if err := syscall.Statfs(delegateDir, &stat); err == nil {
        freeBytes := stat.Bavail * uint64(stat.Bsize)
        if freeBytes < 100*1024*1024 { // Less than 100MB
            return fmt.Errorf("low disk space: %d bytes remaining", freeBytes)
        }
    }

    return nil
}

// Called periodically in monitorDelegate loop
case <-healthTicker.C:
    if err := c.checkDelegateHealth(tmuxSessionID, delegateDir); err != nil {
        slog.Warn("delegate health check failed", "task_id", task.ID, "error", err)
        // Could trigger early timeout or notification
    }
```

**Impact:** Add to F10 implementation and acceptance criteria.

---

### Concern 22: Dependency Adjustments

**Issue:** Some dependencies need adjustment.

**Resolution:** Updated dependency configuration:

```
set_dependencies:
  # Phase 1 → Phase 2
  F5 depends on: [F1, F2, F3, F4, F4b]  # Added F4b for output-format

  # Phase 2 chain
  F6 depends on: [F5, F7]  # F6 needs F7 design decisions
  F7 depends on: [F5]      # Changed: F7 before F6

  # Phase 3 chain (unchanged)
  F8 depends on: [F7]
  F9 depends on: [F8]
  F10 depends on: [F9]
  F11 depends on: [F10]

  # Phase 4 (unchanged)
  F12 depends on: [F7]
  F13 depends on: [F7]
  F14 depends on: [F11]

  # Phase 5 - explicit F14 dependency for full lifecycle testing
  F15 depends on: [F7]
  F16 depends on: [F11]
  F17 depends on: [F14, F15, F16]  # Already correct
```

**Impact:** Update dependency configuration section.

---

### Concern 23: Tmux Availability Check

**Issue:** Plan assumes tmux is available but lacks runtime checks or graceful fallbacks.

**Resolution:** Add availability check and fallback flag:

```go
// In coordinator initialization or executeDelegatedTaskTmux
func (c *coordinator) ensureTmuxAvailable() error {
    if !shell.IsTmuxAvailable() {
        if c.cfg.Options.ForceInlineDelegate {
            slog.Warn("tmux not available, falling back to inline delegate")
            return nil // Will use inline executor
        }
        return fmt.Errorf("tmux not available and --force-inline not set")
    }
    return nil
}

// CLI flag
rootCmd.PersistentFlags().BoolVar(&forceInline, "force-inline", false,
    "Fall back to inline delegate if tmux unavailable")
```

**Add to Options:**
```go
ForceInlineDelegate bool `json:"force_inline_delegate,omitempty"`
```

**Impact:** Add check to F9, add flag to F1-F4 group, update F5 Options.

---

### Concern 24: Session ID Collision Risk

**Issue:** `sessionID := fmt.Sprintf("delegate-%s", task.ID[:8])` uses only 8 characters. Task IDs sharing the same prefix will collide.

**Resolution:** Use more characters and add timestamp suffix:

```go
func generateDelegateSessionID(taskID string) string {
    // Use 12 chars of task ID + nano timestamp modulo for uniqueness
    suffix := time.Now().UnixNano() % 100000
    return fmt.Sprintf("delegate-%s-%05d", taskID[:min(12, len(taskID))], suffix)
}

// Usage in F9
sessionID := generateDelegateSessionID(task.ID)
```

**Impact:** Update F9 session ID generation.

---

### Concern 25: Race in ProcessPendingDelegateReports

**Issue:** Function deletes from map before processing files. If another delegate completes during processing, its report could be lost.

**Resolution:** Copy then delete, or delete after successful processing:

```go
func (c *coordinator) processPendingDelegateReports(ctx context.Context, sessionID string) error {
    // Take a snapshot of pending reports
    c.delegateReportsMu.Lock()
    pending := make([]string, len(c.pendingDelegateReports[sessionID]))
    copy(pending, c.pendingDelegateReports[sessionID])
    c.delegateReportsMu.Unlock()

    if len(pending) == 0 {
        return nil
    }

    // Process reports...
    var processed []string
    for _, taskID := range pending {
        // ... process each report
        if err := c.processReport(taskID); err == nil {
            processed = append(processed, taskID)
        }
    }

    // Only remove successfully processed reports
    c.delegateReportsMu.Lock()
    remaining := c.pendingDelegateReports[sessionID]
    c.pendingDelegateReports[sessionID] = removeProcessed(remaining, processed)
    if len(c.pendingDelegateReports[sessionID]) == 0 {
        delete(c.pendingDelegateReports, sessionID)
    }
    c.delegateReportsMu.Unlock()

    return nil
}
```

**Impact:** Update F11 implementation.

---

### Concern 26: Add --delegate-timeout Flag

**Issue:** 30-minute timeout is hardcoded in F10. Operators need flexibility for long-running tasks.

**Resolution:** Add configurable timeout:

```go
// Add to Options
DelegateTimeoutMinutes int `json:"delegate_timeout_minutes,omitempty" jsonschema:"description=Delegate timeout in minutes,default=30"`

// CLI flag
rootCmd.PersistentFlags().IntVar(&delegateTimeout, "delegate-timeout", 30,
    "Timeout in minutes for delegate tasks")

// In monitorDelegate
func (c *coordinator) monitorDelegate(...) {
    timeoutMinutes := c.cfg.Options.DelegateTimeoutMinutes
    if timeoutMinutes <= 0 {
        timeoutMinutes = 30
    }
    // Also honor task-specific timeout if set
    if task.Timeout > 0 {
        timeoutMinutes = int(task.Timeout.Minutes())
    }
    timeout := time.After(time.Duration(timeoutMinutes) * time.Minute)
    // ...
}
```

**Impact:** Add to F5 Options, update F10.

---

### Concern 27: Headless Validation Mandatory

**Issue:** `validateHeadlessConfig()` (Concern 20) isn't in F6 acceptance criteria. Should be mandatory.

**Resolution:** Add to F6 acceptance criteria and make call prominent:

```go
func (c *coordinator) RunHeadless(ctx context.Context) error {
    // MANDATORY: Validate before any execution
    if err := c.validateHeadlessConfig(); err != nil {
        return fmt.Errorf("headless configuration invalid: %w", err)
    }

    // Rest of implementation...
}
```

**F6 Acceptance Criteria Addition:**
- [ ] **validateHeadlessConfig() called at start (MANDATORY)**
- [ ] Fails fast with clear error if prompt file missing
- [ ] Fails fast if output directory not writable

**Impact:** Update F6 acceptance criteria.

---

### Concern 28: Output Format Consistency

**Issue:** F7 shows `c.headlessWriter` but F6 references `c.cfg.Options.OutputFormat`. Access patterns inconsistent.

**Resolution:** Standardize on Options access with lazy writer initialization:

```go
// Coordinator field
type coordinator struct {
    // ...
    headlessWriter *HeadlessOutputWriter
}

// Lazy initialization in RunHeadless
func (c *coordinator) RunHeadless(ctx context.Context) error {
    // Initialize writer from config
    format := c.cfg.Options.OutputFormat
    if format == "" {
        format = "text"
    }
    c.headlessWriter = NewHeadlessOutputWriter(os.Stdout, format)

    // Use c.headlessWriter throughout
}

// In outputContent, check writer exists
func (c *coordinator) outputContent(content string, contentType string) {
    if c.cfg.Options.HeadlessMode && c.headlessWriter != nil {
        // Use headlessWriter
    }
}
```

**Impact:** Update F6 and F7 for consistent access.

---

### Concern 29: Add Logging to Spawn Command

**Issue:** `nexoraCmd` in F9 lacks logging/debugging flags. Delegate failures are hard to debug.

**Resolution:** Add debug and log file options to spawned command:

```go
func (c *coordinator) executeDelegatedTaskTmux(ctx context.Context, task *delegation.Task) (string, error) {
    // ...

    // Add logging to delegate command
    logFile := filepath.Join(delegateDir, "delegate.log")

    nexoraCmd := fmt.Sprintf(
        "nexora --headless --prompt-file=%s --output-file=%s --status-file=%s --model=%s --working-dir=%s --parent-session=%s --task-id=%s",
        promptPath, outputPath, statusPath, task.Model, task.WorkingDir, task.ParentSession, task.ID,
    )

    // Add debug flag if parent is in debug mode
    if c.cfg.Options.Debug {
        nexoraCmd += " --debug"
    }

    // Redirect stderr to log file for debugging
    nexoraCmd = fmt.Sprintf("(%s) 2>%s", nexoraCmd, logFile)

    // ...
}
```

**Impact:** Update F9 command building.

---

### Concern 30: Model Override Precedence

**Issue:** F4 adds `--model` flag but precedence unclear: CLI flag vs config file vs task-specified model.

**Resolution:** Document and implement clear precedence (highest to lowest):

```go
// Precedence (highest wins):
// 1. Task-specific model (from delegation.Task.Model)
// 2. CLI --model flag
// 3. Config file model_override
// 4. Default model from provider config

func (c *coordinator) resolveModel(taskModel string) string {
    // 1. Task-specific takes highest priority
    if taskModel != "" {
        return taskModel
    }

    // 2. CLI override (already in Options from flag propagation)
    if c.cfg.Options.ModelOverride != "" {
        return c.cfg.Options.ModelOverride
    }

    // 3/4. Fall back to provider default
    return c.cfg.DefaultModel()
}
```

**Documentation for F4:**
```
Model Override Precedence:
1. Task.Model (set by delegate tool)
2. --model CLI flag
3. Options.ModelOverride in config file
4. Provider default
```

**Impact:** Add documentation to F4, implement in F6/F9.

---

### Concern 31: Start Periodic Cleanup Goroutine

**Issue:** F14's `periodicDelegateCleanup` is defined but never started.

**Resolution:** Start during coordinator initialization:

```go
// In NewCoordinator or initialization
func NewCoordinator(ctx context.Context, cfg *config.Config, ...) *coordinator {
    c := &coordinator{
        cfg: cfg,
        // ...
    }

    // Start background cleanup if delegate pool is enabled
    if cfg.Options.EnableDelegatePool {
        go c.periodicDelegateCleanup(ctx)
    }

    return c
}

// Also recover orphaned delegates on startup
func (c *coordinator) initDelegateSystem(ctx context.Context) {
    c.recoverOrphanedDelegates(ctx)
    go c.periodicDelegateCleanup(ctx)
}
```

**F14 Acceptance Criteria Addition:**
- [ ] **periodicDelegateCleanup started during initialization**
- [ ] Cleanup goroutine respects context cancellation

**Impact:** Update coordinator initialization, update F14 acceptance criteria.

---

### Concern 32: Config Field Reference Consistency

**Issue:** F6 uses `c.cfg.PromptFile` but F5 shows `c.cfg.Options.PromptFile`. Must standardize.

**Resolution:** All headless fields are in Options. Verify all references:

```go
// CORRECT - all headless config in Options
c.cfg.Options.HeadlessMode
c.cfg.Options.PromptFile
c.cfg.Options.ContextFile
c.cfg.Options.OutputFile
c.cfg.Options.StatusFile
c.cfg.Options.OutputFormat
c.cfg.Options.ParentSession
c.cfg.Options.TaskID

// INCORRECT - these don't exist at top level
c.cfg.PromptFile      // WRONG
c.cfg.Headless        // WRONG
c.cfg.OutputFile      // WRONG
```

**Update F6 code to use correct paths:**
```go
func (c *coordinator) RunHeadless(ctx context.Context) error {
    opts := c.cfg.Options  // Alias for cleaner code

    prompt, err := os.ReadFile(opts.PromptFile)
    // ...

    if opts.ContextFile != "" {
        contextData, err := os.ReadFile(opts.ContextFile)
        // ...
    }

    // ...

    if opts.OutputFile != "" {
        if err := os.WriteFile(opts.OutputFile, []byte(finalResult), 0644); err != nil {
            // ...
        }
    }
}
```

**Impact:** Audit and fix all config references in F6, F7, F9, F10, F12, F13, F14.

---

### Concern 33: Circuit Breaker for Delegate Spawning

**Issue:** Rapid delegate failures could cascade without backoff, exhausting resources.

**Resolution:** Add circuit breaker pattern:

```go
type CircuitBreaker struct {
    failures      int
    threshold     int
    resetAfter    time.Duration
    lastFailure   time.Time
    state         string // "closed", "open", "half-open"
    mu            sync.Mutex
}

func NewCircuitBreaker(threshold int, resetAfter time.Duration) *CircuitBreaker {
    return &CircuitBreaker{
        threshold:  threshold,
        resetAfter: resetAfter,
        state:      "closed",
    }
}

func (cb *CircuitBreaker) Allow() bool {
    cb.mu.Lock()
    defer cb.mu.Unlock()

    switch cb.state {
    case "open":
        if time.Since(cb.lastFailure) > cb.resetAfter {
            cb.state = "half-open"
            return true
        }
        return false
    case "half-open":
        return true // Allow one test request
    default:
        return true
    }
}

func (cb *CircuitBreaker) RecordSuccess() {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    cb.failures = 0
    cb.state = "closed"
}

func (cb *CircuitBreaker) RecordFailure() {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    cb.failures++
    cb.lastFailure = time.Now()
    if cb.failures >= cb.threshold {
        cb.state = "open"
    }
}

// Usage in coordinator
type coordinator struct {
    // ...
    delegateCircuitBreaker *CircuitBreaker
}

func (c *coordinator) canSpawnDelegate() error {
    if !c.delegateCircuitBreaker.Allow() {
        return fmt.Errorf("delegate spawning circuit breaker open - too many recent failures")
    }
    return nil
}
```

**Impact:** Add to F9 before spawning, add to coordinator struct initialization.

---

### Concern 34: Resource Backpressure System

**Issue:** System lacks backpressure when approaching resource limits.

**Resolution:** Add resource pressure monitoring:

```go
type ResourcePressure struct {
    threshold float64
    monitor   *resourceMonitor
}

func (c *coordinator) checkResourcePressure() error {
    if c.resourceMonitor == nil {
        return nil // No monitoring configured
    }

    pressure := c.resourceMonitor.Pressure()
    if pressure > 0.9 {
        return fmt.Errorf("system under critical load (%.0f%%), refusing new delegates", pressure*100)
    }
    if pressure > 0.8 {
        slog.Warn("system under high load", "pressure", pressure)
        // Could implement throttling here instead of refusing
    }
    return nil
}

// Call before spawning
func (c *coordinator) executeDelegatedTaskTmux(ctx context.Context, task *delegation.Task) (string, error) {
    if err := c.checkResourcePressure(); err != nil {
        return "", err
    }
    // ...
}
```

**Impact:** Add to F9 pre-spawn checks.

---

### Concern 35: Retry Logic with Exponential Backoff

**Issue:** Transient failures (tmux hiccup, temp file issues) cause immediate failure without retry.

**Resolution:** Add retry wrapper:

```go
func (c *coordinator) spawnDelegateWithRetry(ctx context.Context, task *delegation.Task) (string, error) {
    backoff := time.Second
    maxRetries := 3

    var lastErr error
    for i := 0; i < maxRetries; i++ {
        sessionID, err := c.executeDelegatedTaskTmux(ctx, task)
        if err == nil {
            c.delegateCircuitBreaker.RecordSuccess()
            return sessionID, nil
        }

        lastErr = err

        // Check if error is retryable
        if !isRetryableError(err) {
            c.delegateCircuitBreaker.RecordFailure()
            return "", err
        }

        slog.Warn("delegate spawn failed, retrying",
            "task_id", task.ID,
            "attempt", i+1,
            "error", err,
            "backoff", backoff,
        )

        if i < maxRetries-1 {
            select {
            case <-ctx.Done():
                return "", ctx.Err()
            case <-time.After(backoff):
            }
            backoff *= 2
        }
    }

    c.delegateCircuitBreaker.RecordFailure()
    return "", fmt.Errorf("delegate spawn failed after %d retries: %w", maxRetries, lastErr)
}

func isRetryableError(err error) bool {
    errStr := err.Error()
    retryable := []string{
        "tmux server not found",
        "temporary file",
        "resource temporarily unavailable",
        "connection refused",
    }
    for _, s := range retryable {
        if strings.Contains(errStr, s) {
            return true
        }
    }
    return false
}
```

**Impact:** Replace direct executeDelegatedTaskTmux calls with spawnDelegateWithRetry in pool executor.

---

### Concern 36: Security Hardening - Command Validation

**Issue:** Delegate commands could potentially be manipulated.

**Resolution:** Add command validation:

```go
func (c *coordinator) validateDelegateCommand(task *delegation.Task) error {
    // Validate working directory is within allowed paths
    allowedRoots := []string{
        c.cfg.WorkingDir(),
        os.TempDir(),
        filepath.Join(os.Getenv("HOME"), ".nexora"),
    }

    absWorkDir, err := filepath.Abs(task.WorkingDir)
    if err != nil {
        return fmt.Errorf("invalid working directory: %w", err)
    }

    allowed := false
    for _, root := range allowedRoots {
        absRoot, _ := filepath.Abs(root)
        if strings.HasPrefix(absWorkDir, absRoot) {
            allowed = true
            break
        }
    }
    if !allowed {
        return fmt.Errorf("working directory %s not in allowed paths", absWorkDir)
    }

    // Validate task description doesn't contain injection attempts
    dangerous := []string{"$(", "`", "&&", "||", ";", "|", ">", "<"}
    for _, d := range dangerous {
        if strings.Contains(task.Description, d) {
            return fmt.Errorf("task description contains potentially dangerous characters: %s", d)
        }
    }

    return nil
}
```

**Impact:** Add to F9 before tmux spawn.

---

### Concern 37: Prometheus Metrics for Observability

**Issue:** No metrics collection for monitoring delegate health in production.

**Resolution:** Add metrics interface:

```go
type DelegateMetrics interface {
    IncSpawned()
    IncCompleted(success bool)
    ObserveDuration(d time.Duration)
    SetActive(count int)
}

// Default no-op implementation
type noopMetrics struct{}

func (noopMetrics) IncSpawned()                    {}
func (noopMetrics) IncCompleted(success bool)     {}
func (noopMetrics) ObserveDuration(d time.Duration) {}
func (noopMetrics) SetActive(count int)           {}

// Prometheus implementation (optional dependency)
type prometheusMetrics struct {
    spawned   prometheus.Counter
    completed *prometheus.CounterVec
    duration  prometheus.Histogram
    active    prometheus.Gauge
}

func (c *coordinator) recordDelegateSpawned(taskID string) {
    c.delegateMetrics.IncSpawned()
    c.delegateMetrics.SetActive(len(c.activeDelegates))
}

func (c *coordinator) recordDelegateCompleted(taskID string, success bool, duration time.Duration) {
    c.delegateMetrics.IncCompleted(success)
    c.delegateMetrics.ObserveDuration(duration)
    c.delegateMetrics.SetActive(len(c.activeDelegates))
}
```

**Impact:** Add metrics interface to coordinator, call in F9 spawn and F11 completion.

---

### Concern 38: Graceful Shutdown with Drain

**Issue:** Abrupt shutdown loses in-progress delegate work.

**Resolution:** Implement comprehensive shutdown:

```go
func (c *coordinator) shutdownDelegates(ctx context.Context) error {
    c.activeDelegatesMu.Lock()
    delegates := make([]string, 0, len(c.activeDelegates))
    for taskID := range c.activeDelegates {
        delegates = append(delegates, taskID)
    }
    c.activeDelegatesMu.Unlock()

    if len(delegates) == 0 {
        return nil
    }

    slog.Info("shutting down delegates", "count", len(delegates))

    var wg sync.WaitGroup
    errChan := make(chan error, len(delegates))

    for _, taskID := range delegates {
        wg.Add(1)
        go func(id string) {
            defer wg.Done()
            if err := c.gracefulShutdownDelegate(ctx, id); err != nil {
                errChan <- fmt.Errorf("failed to shutdown delegate %s: %w", id, err)
            }
        }(taskID)
    }

    wg.Wait()
    close(errChan)

    var errs []error
    for err := range errChan {
        errs = append(errs, err)
    }

    if len(errs) > 0 {
        return fmt.Errorf("shutdown errors: %v", errs)
    }
    return nil
}

func (c *coordinator) gracefulShutdownDelegate(ctx context.Context, taskID string) error {
    c.activeDelegatesMu.RLock()
    info, exists := c.activeDelegates[taskID]
    c.activeDelegatesMu.RUnlock()

    if !exists {
        return nil
    }

    // Write interrupt signal file
    interruptPath := filepath.Join(info.DelegateDir, "interrupt")
    if err := os.WriteFile(interruptPath, []byte("shutdown"), 0644); err != nil {
        slog.Warn("failed to write interrupt file", "task_id", taskID, "error", err)
    }

    // Wait briefly for graceful exit
    waitCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()

    ticker := time.NewTicker(500 * time.Millisecond)
    defer ticker.Stop()

    for {
        select {
        case <-waitCtx.Done():
            // Force kill
            tmuxMgr := shell.GetTmuxManager()
            tmuxMgr.KillSession(info.SessionID)
            return nil
        case <-ticker.C:
            donePath := filepath.Join(info.DelegateDir, "done")
            if _, err := os.Stat(donePath); err == nil {
                return nil // Clean exit
            }
        }
    }
}
```

**Impact:** Add to coordinator shutdown path, integrate with context cancellation.

---

### Concern 39: Empty Prompt File Validation

**Issue:** F6 validates file existence but not content - empty prompt causes confusing behavior.

**Resolution:** Add content validation:

```go
func (c *coordinator) validateHeadlessConfig() error {
    opts := c.cfg.Options

    // Prompt file must exist, be readable, AND have content
    if opts.PromptFile == "" {
        return fmt.Errorf("--prompt-file is required in headless mode")
    }

    content, err := os.ReadFile(opts.PromptFile)
    if err != nil {
        return fmt.Errorf("prompt file not readable: %w", err)
    }

    // Check for meaningful content (not just whitespace)
    trimmed := strings.TrimSpace(string(content))
    if len(trimmed) == 0 {
        return fmt.Errorf("prompt file is empty or contains only whitespace")
    }

    if len(trimmed) < 10 {
        slog.Warn("prompt file contains very short content", "length", len(trimmed))
    }

    // ... rest of validation
    return nil
}
```

**Impact:** Update F6 validateHeadlessConfig().

---

### Concern 40: Cross-Platform Disk Space Check

**Issue:** Concern 21's `syscall.Statfs_t` is Linux-only, won't compile on Windows/macOS.

**Resolution:** Use cross-platform approach:

```go
import "golang.org/x/sys/unix" // For Unix systems

func checkDiskSpace(path string) (uint64, error) {
    // Use build tags or runtime check
    switch runtime.GOOS {
    case "linux", "darwin", "freebsd":
        return checkDiskSpaceUnix(path)
    case "windows":
        return checkDiskSpaceWindows(path)
    default:
        return 0, nil // Skip check on unknown platforms
    }
}

// +build !windows

func checkDiskSpaceUnix(path string) (uint64, error) {
    var stat unix.Statfs_t
    if err := unix.Statfs(path, &stat); err != nil {
        return 0, err
    }
    return stat.Bavail * uint64(stat.Bsize), nil
}

// Alternative: use github.com/shirou/gopsutil/disk for cross-platform
import "github.com/shirou/gopsutil/v3/disk"

func checkDiskSpacePortable(path string) (uint64, error) {
    usage, err := disk.Usage(path)
    if err != nil {
        return 0, err
    }
    return usage.Free, nil
}
```

**Impact:** Update F10 health check to use portable implementation or skip on unsupported platforms.

---

### Concern 41: Add --dry-run Flag for Validation

**Issue:** No way to validate headless configuration without actual execution.

**Resolution:** Add dry-run mode:

```go
// CLI flag
rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false,
    "Validate configuration and show what would be executed without running")

// In Options
DryRun bool `json:"dry_run,omitempty"`

// In RunHeadless
func (c *coordinator) RunHeadless(ctx context.Context) error {
    if err := c.validateHeadlessConfig(); err != nil {
        return err
    }

    opts := c.cfg.Options

    if opts.DryRun {
        fmt.Println("=== DRY RUN MODE ===")
        fmt.Printf("Prompt file: %s\n", opts.PromptFile)
        fmt.Printf("Output file: %s\n", opts.OutputFile)
        fmt.Printf("Output format: %s\n", opts.OutputFormat)
        fmt.Printf("Model override: %s\n", opts.ModelOverride)
        fmt.Printf("Working directory: %s\n", c.cfg.WorkingDir())

        // Show prompt preview
        content, _ := os.ReadFile(opts.PromptFile)
        preview := string(content)
        if len(preview) > 200 {
            preview = preview[:200] + "..."
        }
        fmt.Printf("Prompt preview:\n%s\n", preview)
        fmt.Println("=== VALIDATION PASSED ===")
        return nil
    }

    // ... actual execution
}
```

**Impact:** Add to F1-F4 flags, add handling to F6.

---

### Concern 42: Missing removeProcessed Helper Function

**Issue:** Concern 25 references `removeProcessed()` helper but doesn't define it.

**Resolution:** Add helper function:

```go
// removeProcessed returns a new slice with processed items removed
func removeProcessed(original, processed []string) []string {
    processedSet := make(map[string]bool, len(processed))
    for _, p := range processed {
        processedSet[p] = true
    }

    result := make([]string, 0, len(original)-len(processed))
    for _, item := range original {
        if !processedSet[item] {
            result = append(result, item)
        }
    }
    return result
}

// Full context for Concern 25:
func (c *coordinator) processPendingDelegateReports(ctx context.Context, sessionID string) error {
    // Take a snapshot of pending reports
    c.delegateReportsMu.Lock()
    pending := make([]string, len(c.pendingDelegateReports[sessionID]))
    copy(pending, c.pendingDelegateReports[sessionID])
    c.delegateReportsMu.Unlock()

    if len(pending) == 0 {
        return nil
    }

    // Process reports
    var processed []string
    for _, taskID := range pending {
        if err := c.processReport(ctx, taskID); err != nil {
            slog.Warn("failed to process delegate report",
                "task_id", taskID,
                "error", err,
            )
            continue // Don't add to processed, will retry
        }
        processed = append(processed, taskID)
    }

    // Only remove successfully processed reports
    c.delegateReportsMu.Lock()
    remaining := c.pendingDelegateReports[sessionID]
    c.pendingDelegateReports[sessionID] = removeProcessed(remaining, processed)
    if len(c.pendingDelegateReports[sessionID]) == 0 {
        delete(c.pendingDelegateReports, sessionID)
    }
    c.delegateReportsMu.Unlock()

    return nil
}
```

**Impact:** Add to F11 implementation.

---

### Concern 43: Session ID Should Use Full UUID

**Issue:** Concern 24 fix still has collision risk with 12 chars + timestamp modulo.

**Resolution:** Use full task ID (which is already a UUID):

```go
func generateDelegateSessionID(taskID string) string {
    // Task IDs are UUIDs - use full ID, tmux handles long session names
    // Just ensure it's valid for tmux (alphanumeric, dash, underscore)
    safe := strings.Map(func(r rune) rune {
        if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
            (r >= '0' && r <= '9') || r == '-' || r == '_' {
            return r
        }
        return '-'
    }, taskID)

    return fmt.Sprintf("delegate-%s", safe)
}

// If task ID isn't UUID, generate one:
func generateDelegateSessionID(taskID string) string {
    if len(taskID) < 32 {
        // Short ID, add UUID suffix
        return fmt.Sprintf("delegate-%s-%s", taskID, uuid.New().String()[:8])
    }
    // Full UUID, use as-is
    return fmt.Sprintf("delegate-%s", taskID)
}
```

**Impact:** Update F9 session ID generation, supersedes Concern 24.

---

## Feature Decomposition

### Phase 0: Shared Utilities (Pre-requisite)

| ID | Feature | Dependencies |
|----|---------|--------------|
| **F0** | Create `internal/fsutil/atomic.go` with shared atomic file utilities | None |

**Sequential:** F0 must complete before Phase 2 (used by F8, F11, F12, F13)

### Phase 1: CLI Flags (Parallel)

| ID | Feature | Dependencies |
|----|---------|--------------|
| **F1** | Add `--headless` flag to root command | None |
| **F2** | Add `--prompt-file` flag to root command | None |
| **F2b** | Add `--context-file` flag to root command | None |
| **F3** | Add `--output-file` flag to root command | None |
| **F3b** | Add `--status-file` flag to root command | None |
| **F4** | Add `--model` flag to root command | None |
| **F4b** | Add `--output-format` flag to root command | None |
| **F4c** | Add `--working-dir` flag to root command | None |

**Parallel Group A:** F1, F2, F2b, F3, F3b, F4, F4b, F4c (all independent)

### Phase 2: Headless Coordinator (Sequential)

| ID | Feature | Dependencies |
|----|---------|--------------|
| **F5** | Add headless fields to `config.Options` + coordinator struct fields | F0, F1-F4c |
| **F7** | Add stdout streaming for headless mode | F5 |
| **F6** | Implement RunHeadless() with continuation loop + interrupt polling | F5, F7 |

**Adjusted order:** F5 → F7 → F6 (F7 before F6 per Concern 22)

### Phase 3: Tmux Delegate Spawning (Sequential)

| ID | Feature | Dependencies |
|----|---------|--------------|
| **F8** | Create .nexora/delegates/ directory structure + state management | F0, F7 |
| **F9** | Implement executeDelegatedTaskTmux() with circuit breaker | F8 |
| **F9b** | Configure pool to use tmux executor + pool.Start() integration | F9 |
| **F10** | Implement delegate monitor goroutine with health checks | F9b |
| **F11** | Implement result injection + processPendingDelegateReports() integration | F10 |

**Sequential:** F8 → F9 → F9b → F10 → F11

### Phase 4: File-Based State (Partial Parallel)

| ID | Feature | Dependencies |
|----|---------|--------------|
| **F12** | Write status file on tool calls | F7 |
| **F13** | Write .done file on completion | F7 |
| **F14** | Cleanup tmux sessions and delegate dirs | F11 |

**Parallel Group B:** F12, F13 (both depend on F7)
**Sequential after B:** F14

### Phase 5: Testing & Validation

| ID | Feature | Dependencies |
|----|---------|--------------|
| **F15** | Add tests for headless mode | F7 |
| **F16** | Add tests for delegate spawning | F11 |
| **F17** | Integration test: full delegate flow | F14 |

**Sequential:** F15, F16 can parallel, F17 last

---

## Swarm Initialization

```
orchestrator_init with:
  projectDir: /home/nexora
  taskDescription: |
    Implement headless delegate system for Nexora.

    Goal: delegate_tool spawns full Nexora in tmux session instead of
    running limited inline sub-agent. This provides:
    - Full tool access (same as main Nexora)
    - True parallelism (separate processes)
    - CLI-first (nexora --headless --prompt-file=task.txt)
    - Swarm compatible (file-based I/O)

    Key files:
    - internal/cmd/root.go (CLI flags)
    - internal/agent/coordinator.go (headless mode)
    - internal/agent/delegate_tool.go (tmux spawning)
    - internal/shell/tmux.go (existing tmux infrastructure)

    Design doc: codedocs/agentic/NEXORA-SWARM-INTEGRATION.md

  existingFeatures:
    - "F0: Create internal/fsutil/atomic.go with WriteAtomic() and ReadIfExists() utilities"
    - "F1: Add --headless flag to internal/cmd/root.go"
    - "F2: Add --prompt-file flag to internal/cmd/root.go"
    - "F2b: Add --context-file flag to internal/cmd/root.go"
    - "F3: Add --output-file flag to internal/cmd/root.go"
    - "F3b: Add --status-file flag to internal/cmd/root.go"
    - "F4: Add --model flag to internal/cmd/root.go"
    - "F4b: Add --output-format flag to internal/cmd/root.go"
    - "F4c: Add --working-dir flag to internal/cmd/root.go"
    - "F5: Add HeadlessMode, PromptFile, OutputFile, OutputFormat, ModelOverride fields to config.Options + coordinator struct fields"
    - "F6: Implement RunHeadless() method with continuation loop and interrupt file polling"
    - "F7: Add stdout streaming in headless mode - tool outputs and responses go to stdout"
    - "F8: Create .nexora/delegates/ directory structure with state management and crash recovery"
    - "F9: Implement executeDelegatedTaskTmux() with circuit breaker and retry logic"
    - "F9b: Configure delegation pool to use tmux executor + pool.Start() integration"
    - "F10: Implement monitorDelegate() goroutine with health checks and exponential backoff"
    - "F11: Implement reportDelegateResult() + processPendingDelegateReports() integration in coordinator.Run()"
    - "F12: Write status JSON file on each tool call in headless mode"
    - "F13: Write .done file on completion in headless mode"
    - "F14: Cleanup tmux session after capturing result + start periodicDelegateCleanup goroutine"
    - "F15: Add unit tests for headless coordinator mode"
    - "F16: Add unit tests for delegate tmux spawning"
    - "F17: Integration test: delegate full flow from spawn to result injection"
```

---

## Dependency Configuration

**Updated per Concerns 22, 44-48:**

```
set_dependencies:
  # Phase 0 (Pre-requisite)
  F0 depends on: []  # Atomic file utilities - no dependencies

  # Phase 1 → Phase 2
  F5 depends on: [F0, F1, F2, F2b, F3, F3b, F4, F4b, F4c]  # All CLI flags + utilities

  # Phase 2 chain (reordered: F7 before F6)
  F7 depends on: [F5]      # Stdout streaming first
  F6 depends on: [F5, F7]  # RunHeadless needs F7 design

  # Phase 3 chain
  F8 depends on: [F0, F7]  # Uses atomic file utilities
  F9 depends on: [F8]
  F9b depends on: [F9]     # Pool executor switch
  F10 depends on: [F9b]
  F11 depends on: [F10]

  # Phase 4
  F12 depends on: [F0, F7]  # Uses atomic file utilities
  F13 depends on: [F0, F7]  # Uses atomic file utilities
  F14 depends on: [F11]

  # Phase 5
  F15 depends on: [F7]
  F16 depends on: [F11]
  F17 depends on: [F14, F15, F16]  # Full lifecycle test
```

---

## Execution Plan

### Step 1: Initialize Session

```
orchestrator_init(projectDir, taskDescription, existingFeatures)
```

### Step 2: Configure Dependencies

```
set_dependencies(feature dependencies as above)
```

### Step 3: Validate Before Starting

```
validate_workers(featureIds: ["F1", "F2", "F3", "F4"])
```

### Step 4: Execute Phase 1 (Parallel)

```
start_parallel_workers(featureIds: ["F1", "F2", "F3", "F4"])
sleep 180
check_all_workers()
# Mark complete as they finish
run_verification("go build ./... && go vet ./...")
commit_progress("feat: add headless CLI flags")
```

### Step 5: Execute Phase 2 (Sequential - F7 before F6 per Concern 22)

```
start_worker("F5")
# Wait, check, mark complete
start_worker("F7")  # Stdout streaming first (needed by F6)
# Wait, check, mark complete
start_worker("F6")  # RunHeadless depends on F7
# Wait, check, mark complete
run_verification("go build ./... && go vet ./...")
commit_progress("feat: implement headless coordinator mode")
```

### Step 6: Execute Phase 3 (Sequential)

```
start_worker("F8")
start_worker("F9")
start_worker("F10")
start_worker("F11")
# Sequential execution
run_verification("go build ./... && go vet ./...")
commit_progress("feat: replace inline delegate with tmux spawn")
```

### Step 7: Execute Phase 4 (Partial Parallel)

```
start_parallel_workers(featureIds: ["F12", "F13"])
# Wait, complete both
start_worker("F14")
run_verification("go build ./... && go vet ./...")
commit_progress("feat: add file-based state for delegates")
```

### Step 8: Execute Phase 5 (Tests)

```
start_parallel_workers(featureIds: ["F15", "F16"])
# Wait, complete both
start_worker("F17")
run_verification("go test ./internal/agent/... -race")
commit_progress("test: add delegate system tests")
```

### Step 9: Final Verification

```
run_verification("go build ./... && go test ./... -race")
```

---

## Feature Details

### F0: Atomic File Utilities

**File:** `internal/fsutil/atomic.go`

Creates shared atomic file utilities used by F8, F11, F12, F13. This prevents code duplication
and ensures consistent atomic write semantics across all delegate file operations.

```go
package fsutil

import (
    "os"
    "path/filepath"
)

// WriteAtomic writes content atomically via temp file + rename.
// This prevents partial reads and ensures file integrity on crash.
func WriteAtomic(path string, content []byte, perm os.FileMode) error {
    dir := filepath.Dir(path)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return err
    }

    // Write to temp file in same directory (required for atomic rename)
    tmp, err := os.CreateTemp(dir, ".tmp-*")
    if err != nil {
        return err
    }
    tmpPath := tmp.Name()
    defer func() {
        tmp.Close()
        os.Remove(tmpPath) // Cleanup on failure
    }()

    if _, err := tmp.Write(content); err != nil {
        return err
    }
    if err := tmp.Sync(); err != nil {
        return err
    }
    if err := tmp.Close(); err != nil {
        return err
    }

    // Atomic rename - this is the commit point
    return os.Rename(tmpPath, path)
}

// ReadIfExists reads a file, returning nil content if it doesn't exist.
// This is useful for optional files like context.txt.
func ReadIfExists(path string) ([]byte, error) {
    data, err := os.ReadFile(path)
    if os.IsNotExist(err) {
        return nil, nil
    }
    return data, err
}

// WriteJSON marshals and atomically writes JSON content.
func WriteJSON(path string, v interface{}) error {
    data, err := json.MarshalIndent(v, "", "  ")
    if err != nil {
        return err
    }
    return WriteAtomic(path, data, 0644)
}
```

**Acceptance:**
- [ ] Package `internal/fsutil` created
- [ ] `WriteAtomic()` uses temp file + rename pattern
- [ ] `ReadIfExists()` returns nil for missing files
- [ ] `WriteJSON()` helper for JSON content
- [ ] Unit tests for atomic write edge cases
- [ ] Build passes

---

### F1: Add --headless Flag

**File:** `internal/cmd/root.go`

```go
var headlessMode bool

func init() {
    rootCmd.PersistentFlags().BoolVar(&headlessMode, "headless", false,
        "Run without TUI (for delegates and automation)")
}
```

**Acceptance:**
- [ ] Flag registered on root command
- [ ] `nexora --help` shows --headless flag
- [ ] Flag value accessible in run function

---

### F2: Add --prompt-file Flag

**File:** `internal/cmd/root.go`

```go
var promptFile string

func init() {
    rootCmd.PersistentFlags().StringVar(&promptFile, "prompt-file", "",
        "Read initial prompt from file (headless mode)")
}
```

**Acceptance:**
- [ ] Flag registered on root command
- [ ] `nexora --help` shows --prompt-file flag
- [ ] Flag value accessible in run function

---

### F2b: Add --context-file Flag

**File:** `internal/cmd/root.go`

```go
var contextFile string

func init() {
    rootCmd.PersistentFlags().StringVar(&contextFile, "context-file", "",
        "Read additional context from file (headless mode)")
}
```

**Acceptance:**
- [ ] Flag registered on root command
- [ ] `nexora --help` shows --context-file flag
- [ ] Flag value accessible in run function

---

### F3: Add --output-file Flag

**File:** `internal/cmd/root.go`

```go
var outputFile string

func init() {
    rootCmd.PersistentFlags().StringVar(&outputFile, "output-file", "",
        "Write final result to file (headless mode)")
}
```

**Acceptance:**
- [ ] Flag registered on root command
- [ ] `nexora --help` shows --output-file flag

---

### F3b: Add --status-file Flag

**File:** `internal/cmd/root.go`

```go
var statusFile string

func init() {
    rootCmd.PersistentFlags().StringVar(&statusFile, "status-file", "",
        "Write status updates to file (headless mode)")
}
```

**Acceptance:**
- [ ] Flag registered on root command
- [ ] `nexora --help` shows --status-file flag

---

### F4: Add --model Flag

**File:** `internal/cmd/root.go`

```go
var modelOverride string

func init() {
    rootCmd.PersistentFlags().StringVar(&modelOverride, "model", "",
        "Model to use (overrides default)")
}
```

**Acceptance:**
- [ ] Flag registered on root command
- [ ] `nexora --help` shows --model flag

---

### F4b: Add --output-format Flag

**File:** `internal/cmd/root.go`

```go
var outputFormat string

func init() {
    rootCmd.PersistentFlags().StringVar(&outputFormat, "output-format", "text",
        "Headless output format: text, json, minimal")
}
```

**Acceptance:**
- [ ] Flag registered on root command
- [ ] `nexora --help` shows --output-format flag
- [ ] Default value is "text"

---

### F4c: Add --working-dir Flag

**File:** `internal/cmd/root.go`

```go
var workingDir string

func init() {
    rootCmd.PersistentFlags().StringVar(&workingDir, "working-dir", "",
        "Working directory for agent operations (defaults to current directory)")
}
```

**Acceptance:**
- [ ] Flag registered on root command
- [ ] `nexora --help` shows --working-dir flag
- [ ] Defaults to current working directory if not specified

---

### F5: Add Headless Fields to config.Options

**File:** `internal/config/config.go` (per Concern 11)

The coordinator uses `*config.Config`, not a separate struct. Add fields to existing `Options`:

```go
type Options struct {
    // ... existing fields (ContextPaths, TUI, Debug, etc.)

    // Headless mode configuration
    HeadlessMode           bool   `json:"headless_mode,omitempty" jsonschema:"description=Run without TUI"`
    PromptFile             string `json:"prompt_file,omitempty" jsonschema:"description=Read prompt from file"`
    ContextFile            string `json:"context_file,omitempty" jsonschema:"description=Read context from file"`
    OutputFile             string `json:"output_file,omitempty" jsonschema:"description=Write output to file"`
    StatusFile             string `json:"status_file,omitempty" jsonschema:"description=Write status updates to file"`
    OutputFormat           string `json:"output_format,omitempty" jsonschema:"description=Headless output format (text/json/minimal),default=text"`
    RequireConfirmation    bool   `json:"require_confirmation,omitempty" jsonschema:"description=Require confirmation in headless mode"`

    // Delegate mode fields (set when running as spawned delegate)
    ParentSession          string `json:"parent_session,omitempty" jsonschema:"description=Parent session ID"`
    TaskID                 string `json:"task_id,omitempty" jsonschema:"description=Delegate task ID"`

    // Resource limits (per Concern 15)
    MaxConcurrentDelegates int    `json:"max_concurrent_delegates,omitempty" jsonschema:"description=Max concurrent delegates,default=5"`

    // Cleanup configuration (per Concern 16)
    DelegateRetentionMinutes int  `json:"delegate_retention_minutes,omitempty" jsonschema:"description=Minutes to retain delegate dirs,default=60"`
}
```

**CLI flags registration (internal/cmd/root.go):**
```go
var (
    headlessMode        bool
    promptFile          string
    contextFile         string
    outputFile          string
    statusFile          string
    outputFormat        string
    modelOverride       string
    requireConfirmation bool
    parentSession       string
    taskID              string
)

func init() {
    rootCmd.PersistentFlags().BoolVar(&headlessMode, "headless", false, "Run without TUI")
    rootCmd.PersistentFlags().StringVar(&promptFile, "prompt-file", "", "Read prompt from file")
    rootCmd.PersistentFlags().StringVar(&contextFile, "context-file", "", "Read context from file")
    rootCmd.PersistentFlags().StringVar(&outputFile, "output-file", "", "Write output to file")
    rootCmd.PersistentFlags().StringVar(&statusFile, "status-file", "", "Write status updates to file")
    rootCmd.PersistentFlags().StringVar(&outputFormat, "output-format", "text", "Output format: text, json, minimal")
    rootCmd.PersistentFlags().StringVar(&modelOverride, "model", "", "Model override")
    rootCmd.PersistentFlags().BoolVar(&requireConfirmation, "require-confirmation", false, "Require confirmation for destructive ops")
    rootCmd.PersistentFlags().StringVar(&parentSession, "parent-session", "", "Parent session ID (delegate mode)")
    rootCmd.PersistentFlags().StringVar(&taskID, "task-id", "", "Task ID (delegate mode)")
}
```

**Flag propagation (per Concern 12):**
```go
func runNexora(cmd *cobra.Command, args []string) error {
    cfg, err := config.Load()
    if err != nil {
        return err
    }

    // CLI flags override config file settings
    if headlessMode {
        cfg.Options.HeadlessMode = true
    }
    if promptFile != "" {
        cfg.Options.PromptFile = promptFile
    }
    if contextFile != "" {
        cfg.Options.ContextFile = contextFile
    }
    if outputFile != "" {
        cfg.Options.OutputFile = outputFile
    }
    if statusFile != "" {
        cfg.Options.StatusFile = statusFile
    }
    if outputFormat != "" {
        cfg.Options.OutputFormat = outputFormat
    }
    if requireConfirmation {
        cfg.Options.RequireConfirmation = true
    }
    if parentSession != "" {
        cfg.Options.ParentSession = parentSession
    }
    if taskID != "" {
        cfg.Options.TaskID = taskID
    }

    // Pass to coordinator
    coord := agent.NewCoordinator(ctx, cfg, ...)
}
```

**Acceptance:**
- [ ] All headless fields added to `config.Options` struct
- [ ] JSON schema tags for config file support
- [ ] CLI flags registered for all fields
- [ ] Flag propagation to config implemented
- [ ] Resource limit and retention fields included

**Coordinator struct fields (internal/agent/coordinator.go):**

These fields must be added to the coordinator struct to support headless delegate operations:

```go
type coordinator struct {
    // ... existing fields (cfg, sessions, messages, permissions, delegatePool, etc.)

    // Headless mode support
    headlessWriter *HeadlessOutputWriter // For stdout streaming (F7)

    // Delegate tracking (F8, F9, F10, F11)
    activeDelegates      map[string]*activeDelegateInfo
    activeDelegatesMu    sync.RWMutex

    // Delegate report queue (F11)
    pendingDelegateReports map[string][]string // sessionID -> []taskID
    delegateReportsMu      sync.Mutex

    // Circuit breaker for delegate spawning (Concern 33)
    delegateCircuitBreaker *CircuitBreaker

    // Metrics interface (Concern 37)
    delegateMetrics DelegateMetrics
}

// activeDelegateInfo tracks a running delegate
type activeDelegateInfo struct {
    SessionID   string           // tmux session ID
    DelegateDir string           // .nexora/delegates/<task-id>
    Task        *delegation.Task
    StartedAt   time.Time
}

// Initialize in NewCoordinator:
func NewCoordinator(ctx context.Context, cfg *config.Config, ...) *coordinator {
    c := &coordinator{
        cfg:                    cfg,
        activeDelegates:        make(map[string]*activeDelegateInfo),
        pendingDelegateReports: make(map[string][]string),
        delegateCircuitBreaker: NewCircuitBreaker(5, 5*time.Minute),
        delegateMetrics:        noopMetrics{},
        // ... other fields
    }

    // Start cleanup goroutine if delegate pool enabled
    if cfg.Options.MaxConcurrentDelegates > 0 {
        go c.periodicDelegateCleanup(ctx)
    }

    return c
}
```

**Acceptance:**
- [ ] All headless fields added to `config.Options` struct
- [ ] All coordinator struct fields added
- [ ] JSON schema tags for config file support
- [ ] CLI flags registered for all fields
- [ ] Flag propagation to config implemented
- [ ] Resource limit and retention fields included
- [ ] Circuit breaker initialized
- [ ] Cleanup goroutine started
- [ ] Build passes

---

### F6: Implement RunHeadless()

**File:** `internal/agent/coordinator.go`

**Critical:** The existing `executeDelegatedTask` in delegate_tool.go uses a sophisticated
10-iteration continuation loop with work-in-progress detection. Headless mode MUST replicate
this behavior to ensure agents actually complete work rather than just describing plans.

```go
func (c *coordinator) RunHeadless(ctx context.Context) error {
    // MANDATORY: Validate before any execution (Concern 27)
    if err := c.validateHeadlessConfig(); err != nil {
        return fmt.Errorf("headless configuration invalid: %w", err)
    }

    opts := c.cfg.Options  // Alias for cleaner code (Concern 32)

    // Initialize headless output writer (Concern 28)
    format := opts.OutputFormat
    if format == "" {
        format = "text"
    }
    c.headlessWriter = NewHeadlessOutputWriter(os.Stdout, format)

    // Read prompt from file
    prompt, err := os.ReadFile(opts.PromptFile)
    if err != nil {
        return fmt.Errorf("failed to read prompt: %w", err)
    }

    // Read optional context file
    var fullPrompt string
    if opts.ContextFile != "" {
        contextData, err := os.ReadFile(opts.ContextFile)
        if err != nil {
            return fmt.Errorf("failed to read context: %w", err)
        }
        fullPrompt = fmt.Sprintf("Context:\n%s\n\nTask:\n%s", string(contextData), string(prompt))
    } else {
        fullPrompt = string(prompt)
    }

    // Derive delegate directory for interrupt file polling
    var delegateDir string
    if opts.StatusFile != "" {
        delegateDir = filepath.Dir(opts.StatusFile)
    }

    // Create session (task session if we have parent info)
    var session *session.Session
    if opts.ParentSession != "" && opts.TaskID != "" {
        agentToolSessionID := c.sessions.CreateAgentToolSessionID(opts.ParentSession, opts.TaskID)
        taskTitle := "Headless: " + truncate(fullPrompt, 50)
        session, err = c.sessions.CreateTaskSession(ctx, agentToolSessionID, opts.ParentSession, taskTitle)
    } else {
        session, err = c.sessions.Create(ctx, c.cfg.WorkingDir())
    }
    if err != nil {
        return fmt.Errorf("failed to create session: %w", err)
    }

    // Auto-approve permissions in headless mode (unless RequireConfirmation is set)
    if !opts.RequireConfirmation {
        c.permissions.AutoApproveSession(session.ID)
    }

    // Run with continuation loop (matches executeDelegatedTask logic)
    const maxIterations = 10
    var allTextParts []string
    var totalToolCalls int

    for iteration := 0; iteration < maxIterations; iteration++ {
        // Check for interrupt file (graceful shutdown signal from parent)
        if delegateDir != "" {
            interruptPath := filepath.Join(delegateDir, "interrupt")
            if _, err := os.Stat(interruptPath); err == nil {
                slog.Info("interrupt file detected, stopping gracefully", "iteration", iteration)
                break
            }
        }

        iterPrompt := fullPrompt
        if iteration > 0 {
            iterPrompt = "Continue with the task. Use the available tools to complete the work. Do not just describe what you will do - actually do it using tools."
        }

        // Write status file for monitoring
        c.writeHeadlessStatus(iteration, totalToolCalls, "running")

        result, err := c.Run(ctx, session.ID, iterPrompt)
        if err != nil {
            if len(allTextParts) > 0 {
                slog.Warn("headless iteration failed, returning partial", "iteration", iteration, "error", err)
                break
            }
            return fmt.Errorf("agent run failed: %w", err)
        }

        // Count tool calls from session messages
        msgs, _ := c.messages.List(ctx, session.ID)
        iterToolCalls := 0
        for _, msg := range msgs {
            if msg.Role == message.Assistant {
                iterToolCalls += len(msg.ToolCalls())
            }
        }
        newToolCalls := iterToolCalls - totalToolCalls
        totalToolCalls = iterToolCalls

        // Collect text content
        allTextParts = append(allTextParts, result)

        // Check completion/continuation indicators (same logic as executeDelegatedTask)
        responseText := strings.ToLower(result)

        workInProgress := containsAny(responseText, []string{
            "now let me", "next, i'll", "let me create", "i'll now",
            "let me check", "let me examine", "let me implement",
        })

        taskComplete := containsAny(responseText, []string{
            "task completed", "finished", "done", "complete",
            "successfully completed", "here are the results", "summary:",
        })

        if taskComplete {
            slog.Debug("headless task complete", "iteration", iteration, "tool_calls", totalToolCalls)
            break
        }

        if totalToolCalls > 0 && !workInProgress {
            slog.Debug("headless finished work", "iteration", iteration, "tool_calls", totalToolCalls)
            break
        }

        if newToolCalls == 0 && !workInProgress && iteration > 0 {
            slog.Warn("headless appears stuck", "iteration", iteration)
            break
        }
    }

    // Final status
    c.writeHeadlessStatus(maxIterations, totalToolCalls, "completed")

    // Write output using atomic file utilities (F0)
    finalResult := strings.Join(allTextParts, "\n\n")
    if opts.OutputFile != "" {
        if err := fsutil.WriteAtomic(opts.OutputFile, []byte(finalResult), 0644); err != nil {
            return fmt.Errorf("failed to write output: %w", err)
        }
    }

    // Write done marker using atomic file utilities (F0)
    if delegateDir != "" {
        donePath := filepath.Join(delegateDir, "done")
        content := fmt.Sprintf("completed_at=%s\n", time.Now().Format(time.RFC3339))
        if err := fsutil.WriteAtomic(donePath, []byte(content), 0644); err != nil {
            slog.Warn("failed to write done marker", "error", err)
        }
    }

    return nil
}

// writeHeadlessStatus writes status.json for monitoring
func (c *coordinator) writeHeadlessStatus(iteration, toolCalls int, phase string) {
    opts := c.cfg.Options
    if opts.StatusFile == "" {
        return
    }
    status := map[string]interface{}{
        "iteration":   iteration,
        "tool_calls":  toolCalls,
        "phase":       phase,
        "last_update": time.Now().Format(time.RFC3339),
    }
    // Use atomic file utilities (F0)
    fsutil.WriteJSON(opts.StatusFile, status)
}

// containsAny checks if text contains any of the indicators
func containsAny(text string, indicators []string) bool {
    for _, ind := range indicators {
        if strings.Contains(text, ind) {
            return true
        }
    }
    return false
}
```

**Acceptance:**
- [ ] Reads prompt from file
- [ ] Reads optional context file
- [ ] Creates task session when parent info provided
- [ ] Auto-approves permissions in headless mode
- [ ] Implements 10-iteration continuation loop (matching existing delegate logic)
- [ ] Detects work-in-progress vs completion indicators
- [ ] Writes status.json for external monitoring
- [ ] Writes done marker file on completion
- [ ] Writes result to output file
- [ ] Build passes

---

### F7: Add Stdout Streaming

**File:** `internal/agent/coordinator.go`

Modify output handling to stream to stdout when headless. Output should be
structured for both human readability and machine parsing.

**Output format:**
```
[TOOL] tool_name
<tool output>

[RESPONSE]
<agent response text>

[STATUS] iteration=N tool_calls=M
```

```go
// HeadlessOutputWriter handles formatted output in headless mode
type HeadlessOutputWriter struct {
    mu     sync.Mutex
    out    io.Writer
    format string // "text" (default), "json", "minimal"
}

func NewHeadlessOutputWriter(out io.Writer, format string) *HeadlessOutputWriter {
    if format == "" {
        format = "text"
    }
    return &HeadlessOutputWriter{out: out, format: format}
}

func (w *HeadlessOutputWriter) WriteToolOutput(toolName, output string) {
    w.mu.Lock()
    defer w.mu.Unlock()

    switch w.format {
    case "json":
        entry := map[string]string{"type": "tool", "name": toolName, "output": output}
        data, _ := json.Marshal(entry)
        fmt.Fprintln(w.out, string(data))
    case "minimal":
        fmt.Fprintln(w.out, output)
    default: // text
        fmt.Fprintf(w.out, "[TOOL] %s\n%s\n\n", toolName, output)
    }
}

func (w *HeadlessOutputWriter) WriteResponse(response string) {
    w.mu.Lock()
    defer w.mu.Unlock()

    switch w.format {
    case "json":
        entry := map[string]string{"type": "response", "content": response}
        data, _ := json.Marshal(entry)
        fmt.Fprintln(w.out, string(data))
    case "minimal":
        fmt.Fprintln(w.out, response)
    default: // text
        fmt.Fprintf(w.out, "[RESPONSE]\n%s\n\n", response)
    }
}

func (w *HeadlessOutputWriter) WriteStatus(iteration, toolCalls int, phase string) {
    w.mu.Lock()
    defer w.mu.Unlock()

    switch w.format {
    case "json":
        entry := map[string]interface{}{
            "type": "status", "iteration": iteration,
            "tool_calls": toolCalls, "phase": phase,
        }
        data, _ := json.Marshal(entry)
        fmt.Fprintln(w.out, string(data))
    case "minimal":
        // No status in minimal mode
    default: // text
        fmt.Fprintf(w.out, "[STATUS] iteration=%d tool_calls=%d phase=%s\n", iteration, toolCalls, phase)
    }
}

// Integration in coordinator
func (c *coordinator) outputContent(content string, contentType string) {
    if c.cfg.Headless && c.headlessWriter != nil {
        switch contentType {
        case "tool":
            c.headlessWriter.WriteToolOutput(c.currentToolName, content)
        case "response":
            c.headlessWriter.WriteResponse(content)
        default:
            fmt.Fprint(os.Stdout, content)
        }
    } else {
        // existing TUI message handling
    }
}
```

**CLI flag for format:**
```go
rootCmd.PersistentFlags().StringVar(&outputFormat, "output-format", "text",
    "Headless output format: text, json, minimal")
```

**Acceptance:**
- [ ] Tool outputs go to stdout with [TOOL] prefix in text mode
- [ ] Agent responses go to stdout with [RESPONSE] prefix
- [ ] Supports json format for machine parsing
- [ ] Supports minimal format for simple output
- [ ] Thread-safe output writing
- [ ] No TUI messages in headless mode
- [ ] Build passes

---

### F8: Create Delegate Directory Structure

**File:** `internal/agent/delegate_tool.go`

Define the file-based state structure for delegates. This addresses fragility concerns
by using atomic operations and clear ownership semantics.

**Directory layout:**
```
.nexora/
├── delegates/                    # Active delegate workspaces
│   └── <task-id>/
│       ├── prompt.txt            # Input prompt (written by spawner)
│       ├── context.txt           # Optional context (written by spawner)
│       ├── status.json           # Progress updates (written by delegate)
│       ├── output.txt            # Final result (written by delegate)
│       └── done                  # Completion marker (written by delegate)
│
├── delegate-reports/             # Completed delegate reports for parent injection
│   └── <task-id-prefix>-<nanos>.json
│
└── delegate-state.json           # Recovery state for crash handling
```

**Atomic file operations:**
```go
// writeAtomicFile writes content atomically via temp file + rename
func writeAtomicFile(path string, content []byte, perm os.FileMode) error {
    dir := filepath.Dir(path)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return err
    }

    // Write to temp file
    tmp, err := os.CreateTemp(dir, ".tmp-*")
    if err != nil {
        return err
    }
    tmpPath := tmp.Name()
    defer func() {
        tmp.Close()
        os.Remove(tmpPath) // Cleanup on failure
    }()

    if _, err := tmp.Write(content); err != nil {
        return err
    }
    if err := tmp.Sync(); err != nil {
        return err
    }
    if err := tmp.Close(); err != nil {
        return err
    }

    // Atomic rename
    return os.Rename(tmpPath, path)
}
```

**Delegate state for crash recovery:**
```go
// DelegateState tracks active delegates for recovery on restart
type DelegateState struct {
    Delegates map[string]DelegateEntry `json:"delegates"`
    UpdatedAt time.Time               `json:"updated_at"`
}

type DelegateEntry struct {
    TaskID        string    `json:"task_id"`
    TmuxSessionID string    `json:"tmux_session_id"`
    ParentSession string    `json:"parent_session"`
    DelegateDir   string    `json:"delegate_dir"`
    StartedAt     time.Time `json:"started_at"`
    Status        string    `json:"status"` // "running", "completed", "failed", "orphaned"
}

// saveDelegateState persists state for crash recovery
func (c *coordinator) saveDelegateState() error {
    c.activeDelegatesMu.Lock()
    defer c.activeDelegatesMu.Unlock()

    state := DelegateState{
        Delegates: make(map[string]DelegateEntry),
        UpdatedAt: time.Now(),
    }
    for id, info := range c.activeDelegates {
        state.Delegates[id] = DelegateEntry{
            TaskID:        id,
            TmuxSessionID: info.SessionID,
            ParentSession: info.Task.ParentSession,
            DelegateDir:   info.DelegateDir,
            StartedAt:     info.StartedAt,
            Status:        "running",
        }
    }

    data, _ := json.MarshalIndent(state, "", "  ")
    statePath := filepath.Join(c.cfg.WorkingDir(), ".nexora", "delegate-state.json")
    return writeAtomicFile(statePath, data, 0644)
}

// recoverOrphanedDelegates checks for delegates from a previous crash
// NOTE: Must acquire activeDelegatesMu when modifying activeDelegates (Concern 7 - race fix)
func (c *coordinator) recoverOrphanedDelegates(ctx context.Context) {
    statePath := filepath.Join(c.cfg.WorkingDir(), ".nexora", "delegate-state.json")
    data, err := os.ReadFile(statePath)
    if err != nil {
        return // No state file, nothing to recover
    }

    var state DelegateState
    if json.Unmarshal(data, &state) != nil {
        return
    }

    tmuxMgr := shell.GetTmuxManager()
    for _, entry := range state.Delegates {
        // Check if tmux session still exists
        if tmuxMgr.IsSessionRunning(entry.TmuxSessionID) {
            // Check if done file exists
            donePath := filepath.Join(entry.DelegateDir, "done")
            if _, err := os.Stat(donePath); err == nil {
                // Completed but not processed - read and report
                outputPath := filepath.Join(entry.DelegateDir, "output.txt")
                result, _ := os.ReadFile(outputPath)
                c.reportDelegateResult(&delegation.Task{
                    ID:            entry.TaskID,
                    ParentSession: entry.ParentSession,
                }, string(result), nil)
            } else {
                // Still running - add to active delegates and resume monitoring
                // CRITICAL: Acquire lock before modifying activeDelegates map
                c.activeDelegatesMu.Lock()
                c.activeDelegates[entry.TaskID] = &activeDelegateInfo{
                    SessionID:   entry.TmuxSessionID,
                    DelegateDir: entry.DelegateDir,
                    Task: &delegation.Task{
                        ID:            entry.TaskID,
                        ParentSession: entry.ParentSession,
                    },
                    StartedAt: entry.StartedAt,
                }
                c.activeDelegatesMu.Unlock()

                go c.monitorDelegate(ctx, &delegation.Task{
                    ID:            entry.TaskID,
                    ParentSession: entry.ParentSession,
                }, entry.TmuxSessionID, entry.DelegateDir)
            }
        } else {
            // Tmux session gone - mark as orphaned and cleanup
            slog.Warn("orphaned delegate found", "task_id", entry.TaskID)
            c.cleanupDelegateDir(entry.DelegateDir)
        }
    }

    // Clear state file after recovery
    os.Remove(statePath)
}
```

**Acceptance:**
- [ ] Directory structure documented and consistent
- [ ] Uses atomic file writes for reliability
- [ ] Tracks active delegates in state file
- [ ] Recovers orphaned delegates on startup
- [ ] Handles concurrent delegate operations safely
- [ ] Build passes

---

### F9: Implement executeDelegatedTaskTmux()

**File:** `internal/agent/delegate_tool.go`

Implements tmux-based delegate spawning. This integrates with the existing delegation pool
by providing an alternative executor that spawns tmux processes instead of inline sub-agents.

**Important:** The existing `executeDelegatedTask` has a sophisticated 10-iteration continuation loop
with work-in-progress detection. The tmux-spawned headless mode must replicate this behavior
internally (see F6 for headless continuation logic).

```go
func (c *coordinator) executeDelegatedTaskTmux(ctx context.Context, task *delegation.Task) (string, error) {
    // Create delegate directory structure
    delegateDir := filepath.Join(c.cfg.WorkingDir(), ".nexora", "delegates", task.ID)
    if err := os.MkdirAll(delegateDir, 0755); err != nil {
        return "", fmt.Errorf("failed to create delegate directory: %w", err)
    }

    // Write prompt file
    promptPath := filepath.Join(delegateDir, "prompt.txt")
    if err := os.WriteFile(promptPath, []byte(task.Description), 0600); err != nil {
        return "", fmt.Errorf("failed to write prompt file: %w", err)
    }

    // Write context file if present
    if task.Context != "" {
        contextPath := filepath.Join(delegateDir, "context.txt")
        if err := os.WriteFile(contextPath, []byte(task.Context), 0600); err != nil {
            return "", fmt.Errorf("failed to write context file: %w", err)
        }
    }

    // Output paths
    outputPath := filepath.Join(delegateDir, "output.txt")
    statusPath := filepath.Join(delegateDir, "status.json")
    donePath := filepath.Join(delegateDir, "done")

    // Build nexora command with all required flags
    nexoraCmd := fmt.Sprintf(
        "nexora --headless --prompt-file=%s --output-file=%s --status-file=%s --model=%s --working-dir=%s --parent-session=%s --task-id=%s",
        promptPath, outputPath, statusPath, task.Model, task.WorkingDir, task.ParentSession, task.ID,
    )

    // Use singleton TmuxManager with correct API
    tmuxMgr := shell.GetTmuxManager()
    sessionID := fmt.Sprintf("delegate-%s", task.ID[:8])
    description := fmt.Sprintf("Delegate: %s", truncate(task.Description, 50))

    session, err := tmuxMgr.NewTmuxSession(sessionID, task.WorkingDir, nexoraCmd, description)
    if err != nil {
        // Cleanup delegate directory on failure
        os.RemoveAll(delegateDir)
        return "", fmt.Errorf("failed to spawn delegate tmux session: %w", err)
    }

    // Store session reference for cleanup
    c.activeDelegatesMu.Lock()
    c.activeDelegates[task.ID] = &activeDelegateInfo{
        SessionID:   session.ID,
        DelegateDir: delegateDir,
        Task:        task,
        StartedAt:   time.Now(),
    }
    c.activeDelegatesMu.Unlock()

    // Monitor in background with cancellation support
    go c.monitorDelegate(ctx, task, session.ID, delegateDir)

    return session.ID, nil
}

// Helper to truncate strings
func truncate(s string, max int) string {
    if len(s) <= max {
        return s
    }
    return s[:max-3] + "..."
}
```

**New struct for tracking:**
```go
type activeDelegateInfo struct {
    SessionID   string
    DelegateDir string
    Task        *delegation.Task
    StartedAt   time.Time
}
```

**Acceptance:**
- [ ] Creates `.nexora/delegates/<task-id>/` directory structure
- [ ] Writes prompt.txt and optional context.txt
- [ ] Uses `shell.GetTmuxManager()` singleton (not NewTmuxManager)
- [ ] Uses correct `NewTmuxSession(sessionID, workingDir, command, description)` API
- [ ] Handles errors on file operations
- [ ] Cleans up on spawn failure
- [ ] Tracks active delegates for cleanup/recovery
- [ ] Returns immediately with session ID
- [ ] Calls saveDelegateState() after adding to activeDelegates
- [ ] Build passes

---

### F9b: Configure Pool Executor Switch

**File:** `internal/agent/coordinator.go`

This feature integrates the tmux executor with the existing delegation pool,
allowing the coordinator to switch between inline and tmux-based delegation.

```go
// initDelegatePool configures the delegation pool with the appropriate executor
func (c *coordinator) initDelegatePool(ctx context.Context) {
    opts := c.cfg.Options

    maxDelegates := opts.MaxConcurrentDelegates
    if maxDelegates <= 0 {
        maxDelegates = 5 // sensible default
    }

    poolConfig := delegation.PoolConfig{
        MaxWorkers:   maxDelegates,
        QueueTimeout: 30 * time.Second,
    }

    c.delegatePool = delegation.NewPool(poolConfig, c.resourceMonitor)

    // Configure executor based on tmux availability
    if shell.IsTmuxAvailable() && !opts.ForceInlineDelegate {
        // Use tmux-based delegation for full tool access
        c.delegatePool.SetExecutor(c.spawnDelegateWithRetry)
        slog.Info("delegate pool using tmux executor")
    } else {
        // Fall back to inline delegation
        c.delegatePool.SetExecutor(c.executeDelegatedTask)
        if !shell.IsTmuxAvailable() {
            slog.Warn("tmux not available, using inline delegate executor")
        }
    }

    // Start the pool
    c.delegatePool.Start(ctx)

    // Recover any orphaned delegates from previous crash
    c.recoverOrphanedDelegates(ctx)
}

// spawnDelegateWithRetry wraps executeDelegatedTaskTmux with retry logic (Concern 35)
func (c *coordinator) spawnDelegateWithRetry(ctx context.Context, task *delegation.Task) (string, error) {
    // Check circuit breaker first (Concern 33)
    if !c.delegateCircuitBreaker.Allow() {
        return "", fmt.Errorf("delegate spawning circuit breaker open - too many recent failures")
    }

    // Check resource pressure (Concern 34)
    if err := c.checkResourcePressure(); err != nil {
        return "", err
    }

    backoff := time.Second
    maxRetries := 3
    var lastErr error

    for i := 0; i < maxRetries; i++ {
        sessionID, err := c.executeDelegatedTaskTmux(ctx, task)
        if err == nil {
            c.delegateCircuitBreaker.RecordSuccess()
            c.delegateMetrics.IncSpawned()
            return sessionID, nil
        }

        lastErr = err

        // Check if error is retryable
        if !isRetryableError(err) {
            c.delegateCircuitBreaker.RecordFailure()
            return "", err
        }

        slog.Warn("delegate spawn failed, retrying",
            "task_id", task.ID,
            "attempt", i+1,
            "error", err,
            "backoff", backoff,
        )

        if i < maxRetries-1 {
            select {
            case <-ctx.Done():
                return "", ctx.Err()
            case <-time.After(backoff):
            }
            backoff *= 2
        }
    }

    c.delegateCircuitBreaker.RecordFailure()
    return "", fmt.Errorf("delegate spawn failed after %d retries: %w", maxRetries, lastErr)
}

func isRetryableError(err error) bool {
    errStr := err.Error()
    retryable := []string{
        "tmux server not found",
        "temporary file",
        "resource temporarily unavailable",
        "connection refused",
    }
    for _, s := range retryable {
        if strings.Contains(errStr, s) {
            return true
        }
    }
    return false
}
```

**Call from NewCoordinator:**
```go
func NewCoordinator(ctx context.Context, cfg *config.Config, ...) *coordinator {
    c := &coordinator{...}

    // Initialize delegate pool with appropriate executor
    c.initDelegatePool(ctx)

    return c
}
```

**Acceptance:**
- [ ] Pool configured with MaxConcurrentDelegates from config
- [ ] Tmux executor used when tmux available and ForceInlineDelegate=false
- [ ] Inline executor used as fallback
- [ ] Pool.Start() called during initialization
- [ ] Orphaned delegates recovered on startup
- [ ] Retry logic with exponential backoff
- [ ] Circuit breaker integration
- [ ] Resource pressure checks
- [ ] Build passes

---

### F10: Implement Delegate Monitor

**File:** `internal/agent/delegate_tool.go`

The monitor watches for delegate completion via the done file, handles timeouts,
and supports context cancellation for graceful shutdown.

```go
func (c *coordinator) monitorDelegate(
    ctx context.Context,
    task *delegation.Task,
    tmuxSessionID string,
    delegateDir string,
) {
    donePath := filepath.Join(delegateDir, "done")
    outputPath := filepath.Join(delegateDir, "output.txt")
    statusPath := filepath.Join(delegateDir, "status.json")

    ticker := time.NewTicker(2 * time.Second)
    defer ticker.Stop()

    // Configurable timeout from task or default
    timeoutDuration := 30 * time.Minute
    if task.Timeout > 0 {
        timeoutDuration = task.Timeout
    }
    timeout := time.After(timeoutDuration)

    tmuxMgr := shell.GetTmuxManager()

    defer func() {
        // Cleanup: remove from active delegates tracking
        c.activeDelegatesMu.Lock()
        delete(c.activeDelegates, task.ID)
        c.activeDelegatesMu.Unlock()
    }()

    for {
        select {
        case <-ctx.Done():
            // Context cancelled - graceful shutdown
            slog.Info("delegate monitor cancelled", "task_id", task.ID)
            tmuxMgr.KillSession(tmuxSessionID)
            c.cleanupDelegateDir(delegateDir)
            return

        case <-timeout:
            slog.Warn("delegate timed out", "task_id", task.ID, "timeout", timeoutDuration)
            tmuxMgr.KillSession(tmuxSessionID)
            c.reportDelegateResult(task, "", fmt.Errorf("delegate timed out after %v", timeoutDuration))
            c.cleanupDelegateDir(delegateDir)
            return

        case <-ticker.C:
            // Check if done file exists
            if _, err := os.Stat(donePath); err == nil {
                // Read result
                result, readErr := os.ReadFile(outputPath)
                if readErr != nil {
                    slog.Error("failed to read delegate output", "task_id", task.ID, "error", readErr)
                    c.reportDelegateResult(task, "", fmt.Errorf("failed to read output: %w", readErr))
                } else {
                    c.reportDelegateResult(task, string(result), nil)
                }

                // Cleanup tmux session
                tmuxMgr.KillSession(tmuxSessionID)
                c.cleanupDelegateDir(delegateDir)
                return
            }

            // Optionally check status file for progress updates
            if statusData, err := os.ReadFile(statusPath); err == nil {
                var status delegateStatus
                if json.Unmarshal(statusData, &status) == nil {
                    slog.Debug("delegate progress",
                        "task_id", task.ID,
                        "iteration", status.Iteration,
                        "tool_calls", status.ToolCalls,
                    )
                }
            }
        }
    }
}

// cleanupDelegateDir removes delegate directory after a delay
// to allow debugging of failed delegates
func (c *coordinator) cleanupDelegateDir(delegateDir string) {
    // Keep failed delegate dirs for 1 hour for debugging
    go func() {
        time.Sleep(1 * time.Hour)
        os.RemoveAll(delegateDir)
    }()
}

// delegateStatus represents the status.json structure
type delegateStatus struct {
    Iteration  int       `json:"iteration"`
    ToolCalls  int       `json:"tool_calls"`
    LastUpdate time.Time `json:"last_update"`
    Phase      string    `json:"phase"` // "running", "completing", "error"
}
```

**Acceptance:**
- [ ] Uses `shell.GetTmuxManager()` singleton correctly
- [ ] Uses correct `KillSession(sessionID)` API
- [ ] Polls for done file every 2 seconds
- [ ] Supports configurable timeout from task
- [ ] Handles context cancellation for graceful shutdown
- [ ] Reads and handles output file errors
- [ ] Optionally monitors status.json for progress
- [ ] Cleans up tmux session after completion
- [ ] Removes from active delegates tracking
- [ ] Delayed cleanup of delegate directories for debugging

---

### F11: Implement Result Injection

**File:** `internal/agent/delegate_tool.go`

**Note:** This pattern already exists in the current `executeDelegatedTask` (lines 417-439).
The tmux version builds on this but adds session locking to prevent conflicts when
the parent session is mid-turn.

**Problem:** Calling `c.Run()` on a parent session that's already processing a turn
can cause race conditions or unexpected behavior.

**Solution:** Use a message queue that the parent session polls, rather than directly
injecting via `c.Run()`.

```go
// DelegateReport represents a completed delegate's result
type DelegateReport struct {
    TaskID       string    `json:"task_id"`
    ParentSession string   `json:"parent_session"`
    Success      bool      `json:"success"`
    Result       string    `json:"result"`
    Error        string    `json:"error,omitempty"`
    CompletedAt  time.Time `json:"completed_at"`
}

// reportDelegateResult queues a delegate result for the parent session to process.
// Instead of calling c.Run() directly (which could conflict with an active turn),
// we write to a queue file that the parent session checks between turns.
func (c *coordinator) reportDelegateResult(task *delegation.Task, result string, err error) {
    report := DelegateReport{
        TaskID:        task.ID,
        ParentSession: task.ParentSession,
        Success:       err == nil,
        Result:        result,
        CompletedAt:   time.Now(),
    }
    if err != nil {
        report.Error = err.Error()
    }

    // Write to delegate reports queue
    reportsDir := filepath.Join(c.cfg.WorkingDir(), ".nexora", "delegate-reports")
    if mkErr := os.MkdirAll(reportsDir, 0755); mkErr != nil {
        slog.Error("failed to create reports dir", "error", mkErr)
        return
    }

    reportPath := filepath.Join(reportsDir, fmt.Sprintf("%s-%d.json", task.ID[:8], time.Now().UnixNano()))
    data, _ := json.Marshal(report)
    if writeErr := os.WriteFile(reportPath, data, 0644); writeErr != nil {
        slog.Error("failed to write delegate report", "error", writeErr)
        return
    }

    slog.Info("delegate report queued",
        "task_id", task.ID,
        "parent_session", task.ParentSession,
        "success", report.Success,
    )

    // Optionally: signal the parent session that a report is ready
    // This could be done via a channel if the parent is in the same process
    c.notifyDelegateComplete(task.ParentSession, task.ID)
}

// notifyDelegateComplete signals that a delegate has completed.
// The parent coordinator checks pending reports between agent turns.
func (c *coordinator) notifyDelegateComplete(parentSession, taskID string) {
    c.delegateReportsMu.Lock()
    defer c.delegateReportsMu.Unlock()

    if c.pendingDelegateReports == nil {
        c.pendingDelegateReports = make(map[string][]string)
    }
    c.pendingDelegateReports[parentSession] = append(
        c.pendingDelegateReports[parentSession],
        taskID,
    )
}

// processPendingDelegateReports is called between agent turns to inject
// delegate results into the conversation. This avoids race conditions
// with an active turn.
func (c *coordinator) processPendingDelegateReports(ctx context.Context, sessionID string) error {
    c.delegateReportsMu.Lock()
    pending := c.pendingDelegateReports[sessionID]
    delete(c.pendingDelegateReports, sessionID)
    c.delegateReportsMu.Unlock()

    if len(pending) == 0 {
        return nil
    }

    reportsDir := filepath.Join(c.cfg.WorkingDir(), ".nexora", "delegate-reports")
    var prompts []string

    for _, taskID := range pending {
        // Find matching report file
        pattern := filepath.Join(reportsDir, taskID[:8]+"-*.json")
        matches, _ := filepath.Glob(pattern)
        if len(matches) == 0 {
            continue
        }

        // Read most recent report
        reportPath := matches[len(matches)-1]
        data, err := os.ReadFile(reportPath)
        if err != nil {
            continue
        }

        var report DelegateReport
        if json.Unmarshal(data, &report) != nil {
            continue
        }

        // Build injection prompt
        if report.Success {
            prompts = append(prompts, fmt.Sprintf(
                "[DELEGATE REPORT - Task ID: %s]\n\nThe delegated sub-agent has completed its task.\n\n## Delegate's Findings:\n\n%s",
                report.TaskID, report.Result,
            ))
        } else {
            prompts = append(prompts, fmt.Sprintf(
                "[DELEGATE FAILED - Task ID: %s]\n\nError: %s",
                report.TaskID, report.Error,
            ))
        }

        // Cleanup report file
        os.Remove(reportPath)
    }

    if len(prompts) == 0 {
        return nil
    }

    // Inject combined delegate reports
    combinedPrompt := strings.Join(prompts, "\n\n---\n\n")
    _, err := c.Run(ctx, sessionID, combinedPrompt)
    return err
}
```

**New coordinator fields:**
```go
type coordinator struct {
    // ... existing fields

    // Delegate report tracking
    delegateReportsMu      sync.Mutex
    pendingDelegateReports map[string][]string // sessionID -> []taskID
}
```

**Integration point:** Call `processPendingDelegateReports()` at the start of each
`coordinator.Run()` call, before processing user input.

**Acceptance:**
- [ ] Uses file-based queue instead of direct c.Run() injection
- [ ] Avoids race condition with active parent turns
- [ ] Queues reports with proper locking
- [ ] Parent processes reports between turns
- [ ] Handles both success and error cases
- [ ] Cleans up report files after processing

**Integration point in coordinator.Run():**

The processPendingDelegateReports() function must be called at the start of each
coordinator.Run() call to inject delegate results before processing user input:

```go
func (c *coordinator) Run(ctx context.Context, sessionID string, prompt string) (string, error) {
    // Process any pending delegate reports before handling new input
    // This ensures delegate results are injected between turns, avoiding race conditions
    if err := c.processPendingDelegateReports(ctx, sessionID); err != nil {
        slog.Warn("failed to process pending delegate reports", "error", err)
        // Non-fatal - continue with normal processing
    }

    // ... rest of existing Run() implementation
}
```

**Acceptance:**
- [ ] Uses file-based queue instead of direct c.Run() injection
- [ ] Avoids race condition with active parent turns
- [ ] Queues reports with proper locking
- [ ] Parent processes reports between turns via Run() integration
- [ ] Handles both success and error cases
- [ ] Cleans up report files after processing
- [ ] processPendingDelegateReports() called at start of coordinator.Run()
- [ ] Build passes

---

### F12: Write Status File on Tool Calls

**File:** `internal/agent/coordinator.go`

In headless mode, write status updates to the status.json file for external monitoring.
This is integrated into the RunHeadless() method (see F6) and the tool execution hooks.

```go
// updateHeadlessStatus is called after each tool execution
func (c *coordinator) updateHeadlessStatus(toolName string, iteration, toolCalls int) {
    if !c.cfg.Headless || c.cfg.StatusFile == "" {
        return
    }

    status := map[string]interface{}{
        "phase":       "running",
        "iteration":   iteration,
        "tool_calls":  toolCalls,
        "last_tool":   toolName,
        "last_update": time.Now().Format(time.RFC3339),
    }

    data, _ := json.Marshal(status)
    writeAtomicFile(c.cfg.StatusFile, data, 0644)
}
```

**Integration:** Hook into the tool execution path to call `updateHeadlessStatus()` after
each tool completes. The status file uses atomic writes to prevent partial reads.

**Acceptance:**
- [ ] Status file updated after each tool call in headless mode
- [ ] Uses atomic file writes
- [ ] Includes iteration, tool count, and timestamp
- [ ] No-op when not in headless mode or no status file configured
- [ ] Build passes

---

### F13: Write Done File on Completion

**File:** `internal/agent/coordinator.go`

The done marker file signals completion to the monitor goroutine.
This is integrated into RunHeadless() (see F6).

```go
// writeDoneMarker signals delegate completion
func (c *coordinator) writeDoneMarker() error {
    if !c.cfg.Headless || c.cfg.StatusFile == "" {
        return nil
    }

    // Derive done path from status file location
    delegateDir := filepath.Dir(c.cfg.StatusFile)
    donePath := filepath.Join(delegateDir, "done")

    // Write completion timestamp
    content := fmt.Sprintf("completed_at=%s\n", time.Now().Format(time.RFC3339))
    return writeAtomicFile(donePath, []byte(content), 0644)
}
```

**Called at:** End of RunHeadless() after writing output file.

**Acceptance:**
- [ ] Done marker written on successful completion
- [ ] Done marker written on error (with error info)
- [ ] Uses atomic write
- [ ] Monitor detects done file and reads output
- [ ] Build passes

---

### F14: Cleanup Tmux Sessions and Delegate Directories

**File:** `internal/agent/delegate_tool.go`

Cleanup is handled in multiple places:
1. `monitorDelegate()` - cleans up after successful completion or timeout
2. `cleanupDelegateDir()` - delayed cleanup preserving debug info
3. `recoverOrphanedDelegates()` - cleans up orphans on startup

```go
// cleanupDelegate performs full cleanup of a delegate
func (c *coordinator) cleanupDelegate(taskID, tmuxSessionID, delegateDir string) {
    // Kill tmux session if still running
    tmuxMgr := shell.GetTmuxManager()
    if tmuxMgr.IsSessionRunning(tmuxSessionID) {
        tmuxMgr.KillSession(tmuxSessionID)
    }

    // Remove from active delegates
    c.activeDelegatesMu.Lock()
    delete(c.activeDelegates, taskID)
    c.activeDelegatesMu.Unlock()

    // Update state file
    c.saveDelegateState()

    // Schedule directory cleanup (delayed for debugging)
    c.cleanupDelegateDir(delegateDir)
}

// periodicDelegateCleanup runs periodically to clean stale delegates
func (c *coordinator) periodicDelegateCleanup(ctx context.Context) {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            c.cleanupStaleDelegates()
        }
    }
}

// cleanupStaleDelegates removes delegates older than 2 hours
func (c *coordinator) cleanupStaleDelegates() {
    delegatesDir := filepath.Join(c.cfg.WorkingDir(), ".nexora", "delegates")
    entries, err := os.ReadDir(delegatesDir)
    if err != nil {
        return
    }

    cutoff := time.Now().Add(-2 * time.Hour)
    for _, entry := range entries {
        if !entry.IsDir() {
            continue
        }

        info, err := entry.Info()
        if err != nil {
            continue
        }

        if info.ModTime().Before(cutoff) {
            delegateDir := filepath.Join(delegatesDir, entry.Name())
            slog.Info("cleaning stale delegate", "dir", delegateDir)
            os.RemoveAll(delegateDir)
        }
    }
}
```

**Acceptance:**
- [ ] Tmux session killed on completion/timeout/error
- [ ] Delegate removed from active tracking
- [ ] State file updated after cleanup
- [ ] Delegate directory cleaned up after delay
- [ ] Periodic cleanup removes stale delegates
- [ ] Build passes

---

### F15: Add Tests for Headless Mode

**File:** `internal/agent/coordinator_headless_test.go`

```go
func TestRunHeadless_BasicPrompt(t *testing.T) {
    // Setup temp files
    promptFile := filepath.Join(t.TempDir(), "prompt.txt")
    outputFile := filepath.Join(t.TempDir(), "output.txt")
    os.WriteFile(promptFile, []byte("What is 2+2?"), 0644)

    cfg := &CoordinatorConfig{
        Headless:   true,
        PromptFile: promptFile,
        OutputFile: outputFile,
    }
    coord := newTestCoordinator(cfg)

    err := coord.RunHeadless(context.Background())
    require.NoError(t, err)

    output, _ := os.ReadFile(outputFile)
    assert.Contains(t, string(output), "4")
}

func TestRunHeadless_ContinuationLoop(t *testing.T) {
    // Test that headless mode continues until task complete
}

func TestRunHeadless_StatusFileUpdates(t *testing.T) {
    // Test status.json is written during execution
}

func TestRunHeadless_DoneMarker(t *testing.T) {
    // Test done file is written on completion
}
```

**Acceptance:**
- [ ] Tests basic headless prompt execution
- [ ] Tests continuation loop behavior
- [ ] Tests status file updates
- [ ] Tests done marker creation
- [ ] All tests pass with race detector

---

### F16: Add Tests for Delegate Spawning

**File:** `internal/agent/delegate_tmux_test.go`

```go
func TestExecuteDelegatedTaskTmux_SpawnSuccess(t *testing.T) {
    if !shell.IsTmuxAvailable() {
        t.Skip("tmux not available")
    }

    task := &delegation.Task{
        ID:          "test-task-123",
        Description: "Test delegate task",
        WorkingDir:  t.TempDir(),
    }

    coord := newTestCoordinator(nil)
    sessionID, err := coord.executeDelegatedTaskTmux(context.Background(), task)

    require.NoError(t, err)
    assert.NotEmpty(t, sessionID)

    // Cleanup
    shell.GetTmuxManager().KillSession(sessionID)
}

func TestMonitorDelegate_CompletionDetection(t *testing.T) {
    // Test monitor detects done file and reads output
}

func TestMonitorDelegate_Timeout(t *testing.T) {
    // Test monitor times out and cleans up
}

func TestRecoverOrphanedDelegates(t *testing.T) {
    // Test orphan recovery on startup
}
```

**Acceptance:**
- [ ] Tests tmux session spawning
- [ ] Tests monitor completion detection
- [ ] Tests timeout handling
- [ ] Tests orphan recovery
- [ ] All tests pass with race detector

---

### F17: Integration Test - Full Delegate Flow

**File:** `internal/agent/delegate_integration_test.go`

```go
func TestDelegateFullFlow(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }
    if !shell.IsTmuxAvailable() {
        t.Skip("tmux not available")
    }

    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
    defer cancel()

    // Create coordinator with tmux executor
    coord := newTestCoordinator(nil)
    coord.delegatePool.SetExecutor(coord.executeDelegatedTaskTmux)
    coord.delegatePool.Start(ctx)

    // Submit delegate task
    taskID, _, err := coord.delegatePool.Submit(
        "List files in current directory",
        "",
        t.TempDir(),
        4096,
        "test-parent-session",
    )
    require.NoError(t, err)

    // Wait for completion
    var completed bool
    for i := 0; i < 60; i++ { // 2 min max
        time.Sleep(2 * time.Second)

        coord.delegateReportsMu.Lock()
        reports := coord.pendingDelegateReports["test-parent-session"]
        coord.delegateReportsMu.Unlock()

        if len(reports) > 0 {
            completed = true
            break
        }
    }

    assert.True(t, completed, "delegate should complete within timeout")
}
```

**Acceptance:**
- [ ] Tests full flow: spawn → execute → monitor → report
- [ ] Validates file-based communication works
- [ ] Validates result injection to parent
- [ ] Tests cleanup after completion
- [ ] Passes with race detector

---

## Verification Commands

After each phase:

```bash
# Build check
go build ./... && go vet ./...

# Test check
go test ./internal/agent/... -race -v

# Manual test (after F7)
echo "What is 2+2?" > /tmp/test.txt
nexora --headless --prompt-file=/tmp/test.txt --output-file=/tmp/result.txt
cat /tmp/result.txt
```

---

## Protocol (Optional)

For safety, register a protocol:

```json
{
  "id": "nexora-delegate-dev",
  "name": "Nexora Delegate Development",
  "version": "1.0.0",
  "constraints": [
    {
      "id": "no-external-deps",
      "type": "file_access",
      "rule": {
        "allowedPaths": ["internal/**", "codedocs/**", "go.mod", "go.sum"]
      },
      "severity": "warning"
    }
  ]
}
```

---

## Quick Start Commands

```bash
# 1. Ensure claude-swarm MCP is available
claude mcp list | grep claude-swarm

# 2. Start swarm session
# (Claude Code will use orchestrator_init)

# 3. Monitor progress
cat /home/nexora/claude-progress.txt

# 4. Check workers
tmux ls | grep cc-worker
```

---

## Concern Traceability Matrix

Maps each concern to the feature(s) that address it for verification:

| Concern | Description | Addressed By | Verified |
|---------|-------------|--------------|----------|
| C11 | Configuration Integration | F5 | [ ] |
| C12 | CLI Flag Propagation | F5 | [ ] |
| C13 | Tool Access Parity | F9 | [ ] |
| C14 | Permission Model Safety | F1, F5, F6 | [ ] |
| C15 | Resource Exhaustion Limits | F5, F9b | [ ] |
| C16 | Configurable Cleanup Retention | F5, F14 | [ ] |
| C17 | Monitor Error Handling and Backoff | F10 | [ ] |
| C18 | Move --output-format to Phase 1 | F4b | [ ] |
| C19 | Extract Atomic File Utilities | F0 | [ ] |
| C20 | Configuration Validation | F6 | [ ] |
| C21 | Health Checks in monitorDelegate | F10 | [ ] |
| C22 | Dependency Adjustments | All Features | [ ] |
| C23 | Tmux Availability Check | F9b | [ ] |
| C24 | Session ID Collision Risk | F9 | [ ] |
| C25 | Race in ProcessPendingDelegateReports | F11 | [ ] |
| C26 | Add --delegate-timeout Flag | F5, F10 | [ ] |
| C27 | Headless Validation Mandatory | F6 | [ ] |
| C28 | Output Format Consistency | F6, F7 | [ ] |
| C29 | Add Logging to Spawn Command | F9 | [ ] |
| C30 | Model Override Precedence | F4, F6, F9 | [ ] |
| C31 | Start Periodic Cleanup Goroutine | F5, F14 | [ ] |
| C32 | Config Field Reference Consistency | F5, F6, F7, F9-F14 | [ ] |
| C33 | Circuit Breaker for Delegate Spawning | F5, F9b | [ ] |
| C34 | Resource Backpressure System | F9b | [ ] |
| C35 | Retry Logic with Exponential Backoff | F9b | [ ] |
| C36 | Security Hardening - Command Validation | F9 | [ ] |
| C37 | Prometheus Metrics for Observability | F5, F9b | [ ] |
| C38 | Graceful Shutdown with Drain | F10 | [ ] |
| C39 | Empty Prompt File Validation | F6 | [ ] |
| C40 | Cross-Platform Disk Space Check | F10 | [ ] |
| C41 | Add --dry-run Flag for Validation | F1-F4, F6 | [ ] |
| C42 | Missing removeProcessed Helper Function | F11 | [ ] |
| C43 | Session ID Should Use Full UUID | F9 | [ ] |

---

## Rollback Procedures

If implementation fails at any phase, use these procedures to safely abort:

### Phase 1 Rollback (CLI Flags)
```bash
# Revert flag additions
git checkout internal/cmd/root.go
```

### Phase 2 Rollback (Headless Coordinator)
```bash
# Revert config and coordinator changes
git checkout internal/config/config.go
git checkout internal/agent/coordinator.go
```

### Phase 3 Rollback (Tmux Spawning)
```bash
# Revert delegate tool changes
git checkout internal/agent/delegate_tool.go
# Kill any orphaned delegate sessions
tmux ls | grep delegate- | cut -d: -f1 | xargs -I{} tmux kill-session -t {}
# Clean up delegate directories
rm -rf .nexora/delegates/*
rm -f .nexora/delegate-state.json
```

### Phase 4 Rollback (File-Based State)
```bash
# Same as Phase 3 - already covered
git checkout internal/agent/delegate_tool.go
git checkout internal/agent/coordinator.go
```

### Full Rollback
```bash
# Revert all changes to pre-implementation state
git stash  # or git checkout -f
# Clean up runtime state
rm -rf .nexora/delegates .nexora/delegate-reports .nexora/delegate-state.json
tmux ls | grep delegate- | cut -d: -f1 | xargs -I{} tmux kill-session -t {} 2>/dev/null || true
```

---

## Required Environment Variables

The headless delegate system requires these environment variables for proper operation:

| Variable | Required | Description | Example |
|----------|----------|-------------|---------|
| `HOME` | Yes | User home directory for config paths | `/home/user` |
| `PATH` | Yes | System path for nexora binary | `/usr/local/bin:...` |
| `NEXORA_*` | No | Nexora-specific configuration | `NEXORA_DEBUG=1` |
| `ANTHROPIC_API_KEY` | Yes* | API key for Claude models | `sk-ant-...` |
| `OPENAI_API_KEY` | No | API key for OpenAI models (if used) | `sk-...` |

*Or configured via psst secret management.

**Environment Inheritance:**
When spawning delegates via tmux, these environment variables are automatically
inherited from the parent process. Additional variables can be passed via the
`--env` flag (if implemented) or environment file.

---

## Success Criteria

The implementation is complete when:

1. **All Features Pass Acceptance Criteria**
   - Each feature's checkbox list is fully checked
   - All tests pass with race detector enabled

2. **Build Verification**
   ```bash
   go build ./... && go vet ./... && go test ./... -race
   ```

3. **Manual Verification**
   ```bash
   # Test headless mode
   echo "What is 2+2?" > /tmp/test.txt
   nexora --headless --prompt-file=/tmp/test.txt --output-file=/tmp/result.txt
   cat /tmp/result.txt

   # Test delegate spawning (if tmux available)
   nexora  # Interactive mode, use delegate_tool
   ```

4. **Traceability Matrix Complete**
   - All concerns mapped to features
   - All verification checkboxes checked

---

## Document Version

- **Version:** 2.0
- **Last Updated:** 2025-12-29
- **Changes:**
  - Added F0 (Atomic File Utilities)
  - Added F2b, F3b, F4b, F4c (missing CLI flags)
  - Added F9b (Pool Executor Switch)
  - Fixed F6 config path references (Concern 32)
  - Added coordinator struct field declarations (F5)
  - Added interrupt file polling to F6
  - Fixed race condition in recoverOrphanedDelegates (F8)
  - Added processPendingDelegateReports integration (F11)
  - Added Concern Traceability Matrix
  - Added Rollback Procedures
  - Added Required Environment Variables
