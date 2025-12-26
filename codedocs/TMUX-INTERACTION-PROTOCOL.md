# TMUX Interaction Protocol
**Date**: December 26, 2025
**Purpose**: AI-driven interactive terminal sessions via TMUX

---

## Overview

Nexora provides AI agents with **persistent TMUX sessions** for executing complex workflows that would be difficult with simple bash commands. This enables:

- ✅ Interactive program control (vi, helix, emacs, etc.)
- ✅ Session persistence across commands
- ✅ Human observation/intervention in real-time
- ✅ Reduced edit failures through visual feedback

---

## Core Concepts

### 1. Session Management

**Naming Convention**: `{word}-{word}-{number}`
```bash
# Good examples
crimson-tiger-42
azure-phoenix-17
edit-agent-refactor
test-integration-runner

# Auto-generated format
nexora-{sessionID}-{timestamp}
```

**Session Limits**:
- Max 10 TMUX sessions per Nexora conversation
- Sessions auto-cleanup on `/clear` or exit
- `pkill tmux` decrements counter

**Creating a Session**:
```bash
bash --shell_id="crimson-tiger-42" "cd /home/nexora && ls"
```

**Reusing a Session**:
```bash
# First call creates session
bash --shell_id="edit-workflow" "vi internal/agent/agent.go"

# Subsequent calls reuse it
bash --shell_id="edit-workflow" "/ProcessMessage"  # Search in vi
bash --shell_id="edit-workflow" "5j"               # Navigate
bash --shell_id="edit-workflow" ":wq"              # Save & quit
```

### 2. Output Handling

Nexora provides **three output modes** (combined):

**Mode 1: Incremental** (for interactive workflows)
```json
{
  "mode": "incremental",
  "captures": [
    {"after_command": "vi file.go", "output": "[vi opened]"},
    {"after_command": "/search", "output": "[found at line 142]"}
  ]
}
```

**Mode 2: Summarized** (for long workflows)
```json
{
  "mode": "summarized",
  "commands_executed": 5,
  "key_outputs": [
    "vi: opened file.go (250 lines)",
    "search: found 'ProcessMessage' at line 142",
    "edit: changed 'oldName' to 'newName'",
    "save: wrote 250 lines"
  ],
  "full_output_path": "/tmp/nexora-session-abc123.log"
}
```

**Mode 3: Final State** (default)
```json
{
  "mode": "final_state",
  "final_pane_output": "[last 50 lines of terminal]",
  "session_id": "edit-agent-go",
  "can_resume": true
}
```

### 3. Special Key Mapping

When sending interactive commands, special keys need escape sequences:

```bash
<Esc>      # Escape key (exit insert mode in vi)
<Enter>    # Enter/Return
<Tab>      # Tab key
<C-c>      # Ctrl+C
<C-d>      # Ctrl+D (EOF)
<Space>    # Space bar
```

**Example VI workflow**:
```bash
bash --shell_id="edit-1" 'vi file.go'
bash --shell_id="edit-1" '/searchTerm<Enter>'
bash --shell_id="edit-1" '5j'                # Down 5 lines
bash --shell_id="edit-1" 'cwnewName<Esc>'    # Change word
bash --shell_id="edit-1" ':wq<Enter>'        # Save & quit
```

---

## Common Workflows

### Workflow 1: Single-File Edit with VI

```bash
# Open file
bash --shell_id="edit-agent-go" "vi internal/agent/agent.go"

# Search for function
bash --shell_id="edit-agent-go" "/ProcessMessage"

# Navigate
bash --shell_id="edit-agent-go" "5j"      # Down 5 lines
bash --shell_id="edit-agent-go" "w"       # Forward one word

# Edit
bash --shell_id="edit-agent-go" "ciwnewHandler"  # Change inner word

# Save
bash --shell_id="edit-agent-go" "<Esc>:wq"

# Verify
bash --shell_id="edit-agent-go" "git diff internal/agent/agent.go"
```

### Workflow 2: Multi-File Edits with VIM

**Option A**: Sequential files
```bash
for file in internal/agent/*.go; do
  bash --shell_id="refactor-imports" "vi $file -c '%s/old.Package/new.Package/g' -c 'wq'"
done
```

**Option B**: VIM argdo (interactive)
```bash
bash --shell_id="refactor-imports" "vim internal/agent/*.go"
bash --shell_id="refactor-imports" ":argdo %s/old/new/gc"
bash --shell_id="refactor-imports" ":wa"
```

**Option C**: Simple sed (non-interactive)
```bash
bash "sed -i 's/old.Package/new.Package/g' internal/agent/*.go"
```

### Workflow 3: Testing with Real-Time Feedback

```bash
# Start test runner
bash --shell_id="test-runner" "go test -v ./internal/agent/..."

# Watch output
bash --shell_id="test-runner" ""  # Capture current state

# If test fails, fix and re-run
bash --shell_id="test-runner" "go test -v ./internal/agent/..."
```

### Workflow 4: Git Workflow

```bash
# Check status
bash "git status"

# Review changes
bash "git diff"

# Stage files
bash "git add internal/agent/agent.go internal/agent/tools/"

# Commit
bash 'git commit -m "Refactor agent handler logic"'

# Push
bash "git push"
```

---

## Editor Comparison

### VI/VIM (Classic)
**Pros**: Ubiquitous, well-tested, fast
**Cons**: Modal (learning curve), terse commands

**Best for**: Quick edits, search/replace, known workflows

**Example**:
```bash
vi file.go
/function
5j
cwnewName
<Esc>:wq
```

### Helix (Modern)
**Pros**: Multiple cursors, modern UX, selection-first
**Cons**: Newer, less familiar

**Best for**: Complex multi-cursor edits, visual selection

**Example**:
```bash
hx file.go
/function<Enter>
5j
cw newName<Esc>
:wq
```

### Recommendation
- **Try both**: Test helix for next complex edit
- **Fall back to VI**: If helix has issues
- **Human can watch**: Intervene if AI struggles

---

## Human Observation

### Watching AI Work

**Human can attach to any TMUX session**:
```bash
# List sessions
tmux list-sessions

# Attach to AI's session
tmux attach -t nexora-edit-workflow

# Detach without killing (Ctrl+B, then D)
```

### Intervention Patterns

**Scenario 1**: AI is editing wrong section
```
Human: *sees AI navigating wrong file*
Human: Ctrl+C (interrupt)
Human: *makes correction manually*
Human: "Resume from here" → AI continues
```

**Scenario 2**: AI struggling with complex edit
```
Human: *watches 10 failed edit attempts*
Human: "Stop, let me show you"
Human: *demonstrates correct edit sequence*
Human: "Now do the same for other files"
```

**Scenario 3**: Faster manual intervention
```
AI: "Going to edit 20 files..."
Human: "Wait, I'll do half, you do half"
# Parallel work!
```

---

## Batching Commands

### Simple Batching (newlines)
```bash
bash '
cd internal/agent
ls -la
grep "ProcessMessage" *.go
'
```

### Complex Batching (session-based)
```bash
# Setup
bash --shell_id="batch-1" "cd internal/agent"

# Execute sequence
bash --shell_id="batch-1" "ls -la"
bash --shell_id="batch-1" "grep ProcessMessage *.go"
bash --shell_id="batch-1" "vi agent.go"
```

### Special Key Mapping for Batches
```bash
# Map special characters
bash --shell_id="vi-batch" '
vi file.go
/searchTerm<Enter>
5j
ciwnewName<Esc>
:wq<Enter>
'
```

**Note**: 101 keyboard keys limit what we need to handle!

---

## Error Handling

### Blocked Commands
```bash
bash "rm -rf /"
# → Blocked: "Command not allowed for security reasons"
```

### Session Not Found
```bash
bash --shell_id="nonexistent" "ls"
# → Creates new session with that ID
```

### TMUX Not Available
```bash
bash "ls"  # Falls back to legacy execution
# → Works, but no session persistence
```

### Session Limit Reached
```bash
# Already have 10 sessions
bash --shell_id="new-session" "ls"
# → Error: "Session limit reached (10/10). Kill a session first."
```

---

## Best Practices

### DO ✅

- **Name sessions descriptively**: `edit-agent-refactor` not `session-1`
- **Reuse sessions**: Continue editing in same session
- **Capture output**: Check vi opened correctly before sending keys
- **Use simple commands**: `ls` over `find $(which ls) -exec ...`
- **Clean up**: Kill sessions when done

### DON'T ❌

- **Don't nest TMUX**: Avoid `tmux attach` inside bash tool
- **Don't assume state**: Always capture-pane to verify
- **Don't batch complex**: Interactive workflows need step-by-step
- **Don't ignore errors**: Check exit codes
- **Don't leak sessions**: 10 session limit is real

---

## Advanced: Multi-Model Orchestration

### Scenario: Research → Plan → Execute

**Terminal 1**: Opus (deep research)
```bash
bash --shell_id="research-opus" "nexora --model opus"
# Opus researches complex algorithm
```

**Terminal 2**: Sonnet (orchestration)
```bash
bash --shell_id="orchestrator" "# Current session"
# Monitors Opus, plans implementation
```

**Terminal 3**: Haiku (execution)
```bash
bash --shell_id="executor-haiku" "nexora --model haiku"
# Haiku implements based on plan
```

**Use Cases**:
- Parallel feature development
- Code review (multiple perspectives)
- Complex debugging (divide & conquer)

---

## Troubleshooting

### Issue: VI commands not working
**Cause**: Special keys not escaped
**Fix**: Use `<Esc>`, `<Enter>` notation

### Issue: Session disappeared
**Cause**: Exceeded 10-session limit
**Fix**: `bash "tmux list-sessions"` to check, kill old ones

### Issue: Output truncated
**Cause**: Output > 30KB limit
**Fix**: Check `/tmp/nexora-output-{sessionID}.log`

### Issue: Human can't see AI's work
**Cause**: Not attached to session
**Fix**: `tmux attach -t nexora-{sessionID}`

---

## References

- `internal/shell/tmux.go` - TMUX session manager
- `internal/agent/tools/bash.go` - Bash tool integration
- `tmux(1)` man page - TMUX manual
- VI/VIM tutorials - Editor commands

---

**Last Updated**: December 26, 2025
**Status**: Production-ready
**Feedback**: Report issues to improve this protocol
