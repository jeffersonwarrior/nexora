# Message Sequence Validation Fix - 2025-12-18

## Problem
Encountered Anthropic API error:
```
messages.25: `tool_use` ids were found without `tool_result` blocks immediately after
```

This indicates that tool calls were made by Claude, but the user sent a new message ("tidy. lint. cq.") before the tool results could be returned, breaking Anthropic's strict message sequencing requirements.

## Root Causes Identified

### 1. **User Interruption of Tool Execution Cycle**
Users could send new messages while tool results were still pending, causing the conversation state to become invalid. The system didn't block user input during tool execution.

### 2. **Missing Message Validation**
No validation existed to catch malformed message sequences before sending to the API. This meant errors were only caught by Anthropic's API after the fact.

### 3. **Insufficient Debugging Information**
When message sequence errors occurred, there was no detailed logging to understand what messages were being sent to the API or why the sequence was invalid.

## Solutions Implemented

### Fix #1: Block User Input During Tool Execution
**Location**: `internal/agent/agent.go:278-289`

Added check in `Run()` method to prevent user messages when tool results are pending:

```go
// Check for pending tool results before accepting new user input
// This prevents users from interrupting the tool execution cycle
if !isContinuation {
    if hasPending, err := a.hasPendingToolResults(ctx, call.SessionID); err == nil && hasPending {
        return nil, fmt.Errorf("session %s: waiting for tool results to complete before accepting new input", call.SessionID)
    }
}
```

**Helper Function**: `hasPendingToolResults()` (line 1621)
- Scans conversation history for tool_use IDs without matching tool_result blocks
- Returns true if any tool calls are unresolved
- Logs pending tool call IDs for debugging

### Fix #2: Pre-API Validation
**Location**: `internal/agent/agent.go:393-400`

Added validation in `PrepareStep` callback before messages are sent to API:

```go
// Validate message sequence before sending to API
if err := a.validateMessageSequence(prepared.Messages); err != nil {
    slog.Error("Invalid message sequence detected",
        "error", err,
        "session_id", call.SessionID,
        "message_count", len(prepared.Messages))
    return callContext, prepared, fmt.Errorf("message validation failed: %w", err)
}
```

**Helper Function**: `validateMessageSequence()` (line 1651)
- Walks through message array tracking tool_use and tool_result pairs
- Ensures every tool_use has a corresponding tool_result before the sequence ends
- Returns detailed error with list of unmatched tool call IDs

### Fix #3: Detailed Message Sequence Logging
**Location**: `internal/agent/agent.go:447`

Added comprehensive logging just before API call:

```go
// Debug log message sequence before API call
a.logMessageSequence(prepared.Messages, call.SessionID)
```

**Helper Function**: `logMessageSequence()` (line 1683)
- Only runs when debug logging enabled (performance-conscious)
- Logs detailed breakdown of each message:
  - Role (user/assistant/system)
  - Content types (text, tool_use, tool_result)
  - Tool names and IDs
  - Message counts and totals
- Example output:
  ```
  message sequence before API call:
    session_id=abc123
    total_messages=25
    last_message_role=user
    tool_use_count=3
    tool_result_count=3
    sequence=[
      "0:system [text:42 chars]",
      "1:user [text:15 chars]",
      "2:assistant [tool_use:bash] [tool_use:view] [text:45 chars]",
      "3:user [tool_result:toolu_01xyz] [tool_result:toolu_01abc]",
      ...
    ]
  ```

## Message Flow (Corrected)

### Before Fix
1. User sends "do X"
2. Claude responds with tool_use blocks (IDs: abc, def, ghi)
3. **User immediately sends "tidy. lint. cq."** ← PROBLEM
4. Previous tool calls (abc, def, ghi) never get tool_result blocks
5. API rejects: "tool_use ids found without tool_result blocks"

### After Fix
1. User sends "do X"
2. Claude responds with tool_use blocks (IDs: abc, def, ghi)
3. `IsSessionBusy()` returns true, user input blocked
4. Tools execute, tool_result blocks created for abc, def, ghi
5. Validation confirms all tool_use blocks matched
6. Session marked not busy, user can send next message

## Testing

Compilation: ✅ Success
```bash
go build -o /tmp/nexora-test .
```

Tests: ✅ All passing
```bash
go test ./internal/agent/...
```

## Files Modified

- `internal/agent/agent.go`:
  - Line 278-289: Added pending tool result check
  - Line 393-400: Added message validation
  - Line 447: Added debug logging call
  - Line 1621-1644: New `hasPendingToolResults()` function
  - Line 1651-1681: New `validateMessageSequence()` function
  - Line 1683-1734: New `logMessageSequence()` function

## Type Corrections

During implementation, corrected fantasy library type references:
- `fantasy.ToolUsePart` → `fantasy.ToolCallPart`
- `message.ToolResult.ID` → `message.ToolResult.ToolCallID`

## Debug Usage

To see detailed message sequence logging, run with debug level:
```bash
NEXORA_LOG_LEVEL=debug nexora chat
```

Or set in config:
```json
{
  "log_level": "debug"
}
```

## Future Considerations

1. **UI Feedback**: Consider showing "⏳ Waiting for tool results..." in TUI when input is blocked
2. **Timeout**: Add timeout for stuck tool executions (e.g., 60s max)
3. **Force Continue**: Allow advanced users to force-skip pending tools (with warning)
4. **Metrics**: Track how often users attempt to send messages during tool execution

## Related Issues

This fix addresses the core issue in the error report where users could interrupt the tool execution cycle. The three-pronged approach ensures:
- **Prevention**: Block problematic input at the source
- **Detection**: Catch issues before they reach the API
- **Debugging**: Provide detailed information when issues occur
