# Bash Safety Audit & Recommendations
**Date**: December 26, 2025
**Status**: Safety gaps identified, expansion needed

---

## Current Safety Implementation

### 1. Safe Commands List (`internal/agent/tools/safe.go`)

**Purpose**: Read-only commands that bypass permission system

**Current Coverage** (55 commands):
- Core utils: ls, pwd, date, echo, env, whoami, etc.
- Git read-only: status, diff, log, show, branch, etc.
- System info: ps, top, df, du, uname, etc.
- Windows-specific: ipconfig, tasklist, systeminfo, etc.

### 2. BlockFuncs System (`internal/shell/shell.go`)

**Purpose**: Block dangerous commands before execution

**Current Status**: ❌ **NOT IMPLEMENTED**
```go
func blockFuncs() []shell.BlockFunc {
    return nil  // ← No blockers active!
}
```

### 3. Permission System

**Purpose**: Require user approval for non-safe commands

**Current Behavior**:
- Safe commands → Auto-approved
- Everything else → Permission request
- User can approve/deny

---

## Safety Gaps Identified 🚨

### Critical Gaps

1. **No Destructive Command Blocking**
   ```bash
   rm -rf /              # ← Not blocked!
   rm -rf *              # ← Not blocked!
   rm -rf ~/.ssh         # ← Not blocked!
   ```

2. **No Process Kill Protection**
   ```bash
   pkill nexora          # ← Could kill self!
   killall -9 tmux       # ← Could kill all sessions!
   kill -9 1             # ← Could kill init (if root)!
   ```

3. **No Logic Bomb Detection**
   ```bash
   while true; do :; done                    # Fork bomb
   dd if=/dev/zero of=/dev/sda               # Disk wipe
   :(){ :|:& };:                             # Classic fork bomb
   find / -exec rm {} \;                     # Mass deletion
   ```

4. **No Dangerous Flag Detection**
   ```bash
   git push --force                          # Force push
   docker rm -f $(docker ps -aq)             # Nuke all containers
   npm install --global malicious-package    # Global install
   ```

5. **No Path Safety**
   ```bash
   cd / && rm -rf *      # Root deletion
   chmod 777 ~/.ssh      # Unsafe permissions
   ```

---

## Recommended Safety Additions

### Phase 1: Critical Blockers (Immediate)

Add to `blockFuncs()`:

```go
func blockFuncs() []shell.BlockFunc {
    return []shell.BlockFunc{
        // Block recursive force removal
        shell.ArgumentsBlocker("rm", []string{}, []string{"-rf", "--recursive --force"}),
        shell.ArgumentsBlocker("rm", []string{}, []string{"-fr"}),

        // Block killing Nexora/tmux processes
        func(args []string) bool {
            if len(args) == 0 {
                return false
            }
            cmd := args[0]
            if cmd == "pkill" || cmd == "killall" {
                for _, arg := range args[1:] {
                    if strings.Contains(arg, "nexora") ||
                       strings.Contains(arg, "tmux") {
                        return true
                    }
                }
            }
            return false
        },

        // Block init/systemd kills
        func(args []string) bool {
            if len(args) >= 2 && args[0] == "kill" {
                for _, arg := range args[1:] {
                    if arg == "1" || arg == "-1" {
                        return true
                    }
                }
            }
            return false
        },

        // Block format/wipe commands
        shell.CommandsBlocker([]string{
            "mkfs",
            "fdisk",
            "dd",
            "shred",
        }),

        // Block fork bombs (simple patterns)
        func(args []string) bool {
            cmdStr := strings.Join(args, " ")
            forkBombs := []string{
                ":()",
                "while true",
                ":|:",
            }
            for _, pattern := range forkBombs {
                if strings.Contains(cmdStr, pattern) {
                    return true
                }
            }
            return false
        },
    }
}
```

### Phase 2: Dangerous Flags (High Priority)

```go
// Block force flags on destructive ops
shell.ArgumentsBlocker("git", []string{"push"}, []string{"--force", "-f"}),
shell.ArgumentsBlocker("docker", []string{"rm"}, []string{"-f", "--force"}),
shell.ArgumentsBlocker("npm", []string{"install"}, []string{"-g", "--global"}),
shell.ArgumentsBlocker("chmod", []string{"777"}, []string{}),
shell.ArgumentsBlocker("chmod", []string{"-R", "777"}, []string{}),
```

### Phase 3: Path Safety (Medium Priority)

```go
// Block operations in dangerous directories
func(args []string) bool {
    if len(args) < 2 {
        return false
    }

    dangerousPaths := []string{"/", "/bin", "/usr", "/etc", "/sys"}
    cmd := args[0]

    if cmd == "rm" || cmd == "mv" || cmd == "chmod" {
        for _, arg := range args {
            for _, dangerous := range dangerousPaths {
                if strings.HasPrefix(arg, dangerous) {
                    return true
                }
            }
        }
    }
    return false
},
```

### Phase 4: Smart Pattern Detection (Future/aiops)

- **ML-based anomaly detection**: Unusual command sequences
- **Rate limiting**: Prevent rapid-fire destructive commands
- **Confirmation prompts**: Double-check for critical operations
- **Undo/rollback**: Transaction log for file operations
- **Sandbox mode**: Test commands in isolated environment first

---

## Implementation Priority

### Immediate (This Week)
- ✅ Audit current safety system (DONE)
- 🔲 Implement Phase 1 critical blockers
- 🔲 Test blocker functions
- 🔲 Update safe commands list

### High Priority (Next Sprint)
- 🔲 Implement Phase 2 dangerous flags
- 🔲 Add logging for blocked commands
- 🔲 User notification system for blocks

### Medium Priority (Future)
- 🔲 Implement Phase 3 path safety
- 🔲 Rate limiting system
- 🔲 Command history analysis

### Future (aiops Project)
- 🔲 ML-based anomaly detection
- 🔲 Smart pattern learning
- 🔲 Sandbox execution mode

---

## Testing Strategy

### Unit Tests
```go
func TestBlockDestructiveCommands(t *testing.T) {
    tests := []struct{
        command string
        shouldBlock bool
    }{
        {"rm -rf /", true},
        {"pkill nexora", true},
        {"kill -9 1", true},
        {"dd if=/dev/zero of=/dev/sda", true},
        {"rm file.txt", false},
        {"ls -la", false},
    }
    // ... test implementation
}
```

### Integration Tests
- Test with TMUX sessions
- Verify permission system interaction
- Check error messages are helpful

### Manual Tests
- Try to break out of safety (ethical hacking)
- Test edge cases
- Verify user experience isn't degraded

---

## Open Questions

1. **Should we block `sudo`?**
   - Pro: Prevents privilege escalation
   - Con: Legitimate admin tasks need it
   - **Recommendation**: Block by default, allow with explicit permission

2. **How aggressive should fork bomb detection be?**
   - Current: Simple string matching
   - Better: Parse command structure
   - **Recommendation**: Start simple, improve with feedback

3. **Should blockers be configurable?**
   - Allow users to customize blocked commands
   - **Recommendation**: Yes, via config file (later)

4. **What about TMUX send-keys commands?**
   - Can bypass blockers by sending raw keys
   - **Recommendation**: Apply blockers before send-keys

---

## References

- `internal/shell/shell.go` - BlockFunc system
- `internal/shell/command_block_test.go` - Blocker tests
- `internal/agent/tools/safe.go` - Safe commands list
- `internal/agent/tools/bash.go` - Permission integration

---

**Next Steps**: Implement Phase 1 blockers, test thoroughly, then move to Phase 2.
