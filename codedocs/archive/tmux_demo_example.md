# TMUX Bash Shell Example

## Simple Command (Creates New TMUX Session)
```
User: bash "echo hello world"
```

**Expected Response:**
```json
{
  "content": "hello world\n\n<cwd>/home/nexora</cwd>",
  "metadata": {
    "start_time": 1735161724123,
    "end_time": 1735161724156,
    "output": "hello world",
    "description": "",
    "working_directory": "/home/nexora",
    "shell_id": "nexora-session-abc123-1735161724123456789",
    "tmux_session_id": "nexora-nexora-session-abc123-1735161724123456789",
    "tmux_pane_id": "%0",
    "tmux_available": true
  }
}
```

## Persistent Session (Reuses TMUX Session)
```
User: bash --shell_id="hello-world-demo" "echo hello"
User: bash --shell_id="hello-world-demo" "echo world" 
User: bash --shell_id="hello-world-demo" "echo hello world"
```

**First Command Response:**
```json
{
  "content": "hello\n\n<cwd>/home/nexora</cwd>",
  "metadata": {
    "shell_id": "hello-world-demo",
    "tmux_session_id": "nexora-hello-world-demo",
    "tmux_pane_id": "%0",
    "tmux_available": true
  }
}
```

**Subsequent Commands** would reuse the same TMUX session, maintaining state between commands.

## Interactive Workflow Example
```
# Start interactive session
bash --shell_id="edit-session" "vi test.txt"
bash --shell_id="edit-session" "i"
bash --shell_id="edit-session" "hello world"
bash --shell_id="edit-session" "<Esc>"
bash --shell_id="edit-session" ":wq"
```

## Key TMUX Features Demonstrated:

1. **Session Persistence**: Same shell_id reuses TMUX session
2. **State Maintenance**: Working directory, environment variables preserved  
3. **Interactive Programs**: vi, emacs, etc. work via send-keys
4. **Human Visibility**: User can attach to TMUX session to watch/intervene
5. **Auto-Cleanup**: Sessions cleaned up on conversation end

## Shell ID Pattern:
- **Format**: `{sessionID}-{nanosecond-timestamp}`
- **Example**: `nexora-session-123-1735161724123456789`
- **TMUX Name**: `nexora-{shellID}` → `nexora-nexora-session-123-1735161724123456789`
