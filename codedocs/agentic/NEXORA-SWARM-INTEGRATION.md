# Nexora + Claude-Swarm Integration Design

**Version:** v0.29.3+
**Status:** Design Document
**Last Updated:** 2025-12-29

---

## Executive Summary

Nexora's delegate system will spawn **full Nexora instances in tmux sessions**, matching the claude-swarm pattern. This provides:

- **Full agent capabilities** - Same tools, same logic as main session
- **True parallelism** - Separate processes, no blocking
- **CLI-first** - Works standalone or as part of any swarm
- **Fast** - No TUI overhead in headless mode
- **Observable** - Streaming output via tmux

---

## Architecture

### Current (Broken)

```
Nexora TUI Session
    └── delegate_tool (inline)
        └── Limited sub-agent (4 tools)
            └── Flawed stop logic
            └── Same process (blocking)
            └── No persistence
```

### Proposed (Reliable)

```
Nexora TUI Session
    │
    └── delegate_tool
            │
            ├── Write task → .nexora/delegates/{id}.prompt
            │
            ├── Spawn tmux session: nexora-delegate-{id}
            │   └── nexora --headless --prompt-file=... --output-file=...
            │       └── Full SessionAgent
            │       └── All tools (glob, grep, view, edit, write, bash, etc.)
            │       └── Same Run() logic as TUI
            │       └── Streams to tmux pane
            │
            ├── Return immediately: "Spawned: nexora-delegate-{id}"
            │
            └── Background monitor: .done file → inject result to parent
```

### Key Insight

**Delegate = Headless Nexora in tmux**

Not a limited sub-agent. The actual Nexora binary running without TUI.

---

## Implementation Plan

### Phase 1: Headless Mode (v0.29.3)

#### 1.1 Add `--headless` Flag

**File:** `internal/cmd/root.go`

```go
var headlessMode bool

func init() {
    rootCmd.PersistentFlags().BoolVar(&headlessMode, "headless", false,
        "Run without TUI (for delegates and automation)")
}
```

#### 1.2 Add `--prompt-file` Flag

**File:** `internal/cmd/root.go`

```go
var promptFile string

func init() {
    rootCmd.PersistentFlags().StringVar(&promptFile, "prompt-file", "",
        "Read initial prompt from file (headless mode)")
}
```

#### 1.3 Add `--output-file` Flag

**File:** `internal/cmd/root.go`

```go
var outputFile string

func init() {
    rootCmd.PersistentFlags().StringVar(&outputFile, "output-file", "",
        "Write final result to file (headless mode)")
}
```

#### 1.4 Add `--model` Flag

**File:** `internal/cmd/root.go`

```go
var modelOverride string

func init() {
    rootCmd.PersistentFlags().StringVar(&modelOverride, "model", "",
        "Model to use (e.g., 'claude-sonnet-4-20250514', 'gpt-4o', 'deepseek-v3')")
}
```

**Use cases:**
- Fast delegates: `--model=claude-haiku-3-5`
- Complex tasks: `--model=claude-sonnet-4`
- Cost optimization: `--model=deepseek-v3`

#### 1.5 Headless Coordinator Mode

**File:** `internal/agent/coordinator.go`

```go
type CoordinatorConfig struct {
    // ... existing fields
    Headless   bool   // No TUI, stdout output
    PromptFile string // Read task from file
    OutputFile string // Write result to file
}

func (c *coordinator) RunHeadless(ctx context.Context) error {
    // 1. Read prompt from file
    prompt, err := os.ReadFile(c.cfg.PromptFile)
    if err != nil {
        return fmt.Errorf("failed to read prompt file: %w", err)
    }

    // 2. Create session
    session, err := c.sessions.Create(ctx, c.cfg.WorkingDir())
    if err != nil {
        return fmt.Errorf("failed to create session: %w", err)
    }

    // 3. Write status file
    c.writeStatusFile(session.ID, "running", 0)

    // 4. Run with full agent (streaming to stdout)
    result, err := c.Run(ctx, session.ID, string(prompt))

    // 5. Write completion
    if c.cfg.OutputFile != "" {
        os.WriteFile(c.cfg.OutputFile, []byte(result), 0644)
    }
    c.writeDoneFile(session.ID, result)

    return nil
}
```

#### 1.5 Stdout Streaming in Headless Mode

**File:** `internal/agent/coordinator.go`

When `Headless == true`:
- Tool outputs go to stdout
- Agent responses go to stdout
- No TUI messages (tea.Msg)
- Progress written to status file

```go
func (c *coordinator) streamToOutput(content string) {
    if c.cfg.Headless {
        fmt.Print(content) // Direct to stdout (captured by tmux)
    } else {
        // Send to TUI via tea.Msg
    }
}
```

---

### Phase 2: Tmux Delegate Spawning (v0.29.3)

#### 2.1 Replace Inline Executor

**File:** `internal/agent/delegate_tool.go`

```go
func (c *coordinator) executeDelegatedTask(ctx context.Context, task *delegation.Task) (string, error) {
    // 1. Create delegate directory
    delegateDir := filepath.Join(c.cfg.WorkingDir(), ".nexora", "delegates")
    os.MkdirAll(delegateDir, 0755)

    // 2. Write prompt file
    promptPath := filepath.Join(delegateDir, task.ID+".prompt")
    promptContent := buildDelegatePrompt(task)
    os.WriteFile(promptPath, []byte(promptContent), 0600)

    // 3. Define output paths
    outputPath := filepath.Join(delegateDir, task.ID+".output")
    statusPath := filepath.Join(delegateDir, task.ID+".status")
    donePath := filepath.Join(delegateDir, task.ID+".done")

    // 4. Build nexora command
    nexoraCmd := fmt.Sprintf(
        "nexora --headless --prompt-file=%s --output-file=%s --working-dir=%s --model=%s",
        promptPath, outputPath, task.WorkingDir, task.Model,
    )

    // 5. Spawn in tmux session
    sessionName := fmt.Sprintf("nexora-delegate-%s", task.ID[:8])
    err := c.tmux.NewSession(sessionName, nexoraCmd)
    if err != nil {
        return "", fmt.Errorf("failed to spawn delegate: %w", err)
    }

    slog.Info("delegate spawned",
        "task_id", task.ID,
        "tmux_session", sessionName,
        "prompt_file", promptPath,
    )

    // 6. Monitor for completion (background)
    go c.monitorDelegate(ctx, task, sessionName, donePath, outputPath)

    return sessionName, nil
}
```

#### 2.2 Delegate Monitor

**File:** `internal/agent/delegate_tool.go`

```go
func (c *coordinator) monitorDelegate(
    ctx context.Context,
    task *delegation.Task,
    sessionName string,
    donePath string,
    outputPath string,
) {
    ticker := time.NewTicker(2 * time.Second)
    defer ticker.Stop()

    timeout := time.After(30 * time.Minute)

    for {
        select {
        case <-ctx.Done():
            return
        case <-timeout:
            slog.Warn("delegate timed out", "task_id", task.ID)
            c.tmux.KillSession(sessionName)
            c.reportDelegateResult(task, "", fmt.Errorf("timeout after 30 minutes"))
            return
        case <-ticker.C:
            // Check for .done file
            if _, err := os.Stat(donePath); err == nil {
                // Read result
                result, _ := os.ReadFile(outputPath)
                c.reportDelegateResult(task, string(result), nil)
                return
            }

            // Optional: Read status file for progress logging
        }
    }
}
```

#### 2.3 Result Injection to Parent

**File:** `internal/agent/delegate_tool.go`

```go
func (c *coordinator) reportDelegateResult(task *delegation.Task, result string, err error) {
    reportPrompt := fmt.Sprintf(
        "[DELEGATE COMPLETE - Task ID: %s]\n\n%s",
        task.ID,
        result,
    )

    if err != nil {
        reportPrompt = fmt.Sprintf(
            "[DELEGATE FAILED - Task ID: %s]\n\nError: %s",
            task.ID,
            err.Error(),
        )
    }

    // Inject to parent session
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    c.Run(ctx, task.ParentSession, reportPrompt)
}
```

---

### Phase 3: File-Based Communication (v0.29.3)

#### 3.1 Status File Schema

**Path:** `.nexora/delegates/{task_id}.status`

```json
{
    "task_id": "abc123",
    "status": "running",
    "started_at": "2025-12-29T04:00:00Z",
    "tool_calls": 15,
    "last_tool": "edit",
    "last_file": "src/main.go",
    "last_activity": "2025-12-29T04:02:30Z",
    "iteration": 4
}
```

#### 3.2 Done File

**Path:** `.nexora/delegates/{task_id}.done`

Existence indicates completion. Contains exit status:
```
success
```
or
```
error: <message>
```

#### 3.3 Output File

**Path:** `.nexora/delegates/{task_id}.output`

Full result text from the delegate's work.

---

### Phase 4: CLI Standalone Mode (v0.29.3)

With headless mode, Nexora can run standalone:

```bash
# Direct execution
nexora --headless --prompt-file=task.txt --output-file=result.txt

# Piped input (future enhancement)
echo "fix the bug" | nexora --headless --stdin

# Part of any script/swarm
for task in tasks/*.prompt; do
    nexora --headless --prompt-file="$task" &
done
wait
```

---

### Phase 5: Claude-Swarm Compatibility (v0.29.4/5)

#### 5.1 Worker Mode Detection

```go
func isClaudeSwarmWorker() bool {
    // Check if spawned by claude-swarm
    return os.Getenv("CLAUDE_SWARM_WORKER") == "true"
}
```

#### 5.2 Claude-Swarm File Paths

When running as claude-swarm worker:
- Read from: `.claude/orchestrator/workers/{feature}.prompt`
- Write to: `.claude/orchestrator/workers/{feature}.log`
- Status: `.claude/orchestrator/workers/{feature}.status`
- Done: `.claude/orchestrator/workers/{feature}.done`

---

## File Structure

```
.nexora/
└── delegates/
    ├── {task_id}.prompt   # Task input
    ├── {task_id}.status   # Running status (JSON)
    ├── {task_id}.output   # Final result
    ├── {task_id}.done     # Completion marker
    └── {task_id}.log      # Full execution log (optional)
```

---

## User Experience

### From TUI

```
User: delegate "fix the authentication bug in auth.go"

Nexora: Spawned delegate in tmux session: nexora-delegate-abc12345

        Monitor with: tmux attach -t nexora-delegate-abc12345

        I'll notify you when it completes.

[5 minutes later]

Nexora: [DELEGATE COMPLETE - Task ID: abc12345]

        Fixed the authentication bug:
        - Updated token validation in auth.go:45
        - Added expiry check in auth.go:78
        - All tests pass
```

### From CLI

```bash
# Standalone delegate
$ nexora --headless --prompt-file=fix-auth.txt --output-file=result.txt
[streams progress to stdout]
$ cat result.txt
Fixed the authentication bug...

# Multiple parallel delegates
$ for task in tasks/*.prompt; do
    nexora --headless --prompt-file="$task" --output-file="${task%.prompt}.result" &
done
$ wait
```

### Monitor Live

```bash
# Attach to running delegate
$ tmux attach -t nexora-delegate-abc12345

# View output without attaching
$ tmux capture-pane -t nexora-delegate-abc12345 -p

# List all delegates
$ tmux ls | grep nexora-delegate
```

---

## Implementation Priority

### v0.29.3 (Current Sprint)

| Priority | Task | File | Effort |
|----------|------|------|--------|
| **P0** | Add `--headless` flag | `internal/cmd/root.go` | 1h |
| **P0** | Add `--prompt-file` flag | `internal/cmd/root.go` | 30m |
| **P0** | Add `--output-file` flag | `internal/cmd/root.go` | 30m |
| **P0** | Implement headless coordinator mode | `internal/agent/coordinator.go` | 4h |
| **P0** | Replace inline executor with tmux spawn | `internal/agent/delegate_tool.go` | 3h |
| **P1** | Add delegate monitor (background) | `internal/agent/delegate_tool.go` | 2h |
| **P1** | Implement result injection | `internal/agent/delegate_tool.go` | 1h |
| **P2** | Status file updates | `internal/agent/coordinator.go` | 2h |
| **P2** | Cleanup on session end | `internal/agent/delegate_tool.go` | 1h |

**Total Effort:** ~15h

### v0.29.4

- Claude-swarm worker mode detection
- Protocol constraint support
- Enhanced monitoring/dashboard

---

## Success Criteria

- [ ] `nexora --headless --prompt-file=task.txt` works standalone
- [ ] `delegate` tool spawns tmux session
- [ ] Delegate has full tool access (same as main Nexora)
- [ ] Live output visible via `tmux attach`
- [ ] Results injected back to parent session
- [ ] Multiple delegates can run in parallel
- [ ] Status files updated during execution
- [ ] Cleanup occurs on completion/error

---

## Comparison: Before vs After

| Aspect | Before (Inline) | After (Tmux) |
|--------|-----------------|--------------|
| **Tools** | 4 (glob, grep, view, bash) | All (same as main) |
| **Reliability** | Flawed heuristics | Full agent logic |
| **Parallelism** | Same process | Separate processes |
| **Monitoring** | None | tmux attach |
| **Persistence** | None | File-based |
| **Speed** | Blocked by parent | Independent |
| **CLI usable** | No | Yes |
| **Swarm compatible** | No | Yes |

---

## References

- `/opt/claude-swarm/` - Claude-Swarm implementation
- `/home/nexora/internal/shell/tmux.go` - Existing tmux infrastructure
- `/home/nexora/internal/agent/delegate_tool.go` - Current delegate (to be replaced)
- `/home/nexora/codedocs/agentic/Agentic_Design_Patterns.json` - Pattern reference
