# Fix: GLM-4.7 Tool Call Degradation from Aggressive Loop Detection

## Problem Summary

GLM-4.7 crashed during long audit tasks, outputting malformed pseudo-commands:
```xml
<file_command> grep -r "permission.Service" /home/nexora --include="*.go" | head -5 </file_command>
```

Instead of making actual tool calls.

## Root Cause

**Cascading failure from aggressive loop detection:**

1. During audit task (338 Go files), GLM-4.7 made many read-only operations (view, grep, ls)
2. Loop detection triggered "No meaningful progress" (no file modifications in 15 actions)
3. Loop detection **returned an error**, halting execution
4. GLM-4.7's tool calling got corrupted:
   - Tool names: `grep_path_pattern</arg_key><arg_value>internal/permission</arg_value>`
   - XML/JSON serialization leaked into tool names
5. After repeated failures, model degraded to outputting `<file_command>` tags as text

## Database Evidence

Session: `9e48c4c0-5615-4df2-8317-022f18451b55` ("Audit 338 Go Files")

```
assistant|glm-4.7|reasoning|[{"type":"reasoning","data":{"thinking":"I'll search with a standard approach:"}},
  {"type":"text","data":{"text":"<file_command>\ngrep -r \"permission.Service\" /home/nexora --include=\"*.go\" | head -5\n</file_command>"}}]
```

Multiple loop detection warnings before corruption:
```
system||text|[{"type":"text","data":{"text":"🛑 Loop detected: No meaningful progress in last 15 actions (0 meaningful successes, 0 unique targets)\n\nStopping execution to prevent infinite loop."}}]
```

## Fix Applied

### 1. Soften Loop Detection (internal/agent/agent.go:833-869)

```go
// Differentiate between hard errors and "no progress" warnings
isErrorLoop := strings.Contains(strings.ToLower(reason), "error") ||
               strings.Contains(strings.ToLower(reason), "oscillating")
isNoProgress := strings.Contains(strings.ToLower(reason), "no meaningful progress")

if isErrorLoop {
    // Hard error loop - halt execution
    return fmt.Errorf("loop detected: %s", reason)
} else if isNoProgress {
    // Soft warning - don't halt, just warn
    // Don't return error - let agent continue
}
```

**Impact:** Read-only operations (audits, exploration) no longer halt execution

### 2. Tool Name Sanitization (internal/agent/utils/tool_id.go:94-114)

```go
// SanitizeToolName removes XML/JSON artifacts that some models (GLM, etc.) leak
func SanitizeToolName(toolName string) string {
    // Strip XML: "tool_name</arg_key>" -> "tool_name"
    if idx := strings.Index(toolName, "<"); idx > 0 {
        return toolName[:idx]
    }
    // Strip JSON: "tool_name{\"arg\":" -> "tool_name"
    if idx := strings.Index(toolName, "{"); idx > 0 {
        return toolName[:idx]
    }
    // Clean: only allow alphanumeric, underscore, hyphen
    ...
}
```

Applied in tool name resolution (internal/agent/tools/aliases.go:78-94):
```go
func ResolveToolName(name string) string {
    // Sanitize first to handle models that leak serialization format
    name = utils.SanitizeToolName(name)
    ...
}
```

**Impact:** Corrupted tool names like `grep_path_pattern</arg_key><arg_value>...` → `grep_path_pattern`

## Testing

```bash
go test ./internal/agent/utils -v -run=TestSanitizeToolName
```

All 8 test cases pass:
- Clean tool names preserved
- XML corruption stripped
- JSON corruption stripped
- Special characters removed

## Expected Behavior After Fix

1. **Audit tasks work:** Read-only operations no longer trigger hard loop detection
2. **Graceful degradation:** Soft warnings instead of execution halts
3. **Tool call recovery:** Malformed tool names sanitized automatically
4. **No pseudo-commands:** Model less likely to output `<file_command>` tags

## Recommendations

1. **Monitor GLM models:** Track tool call quality metrics
2. **Consider provider-specific handling:** GLM may need different loop detection thresholds
3. **Add telemetry:** Log when sanitization fixes corrupted tool names
4. **Document limits:** Clarify that GLM-4.7 may struggle with very long reasoning chains

## Files Modified

- `internal/agent/agent.go` - Softened loop detection
- `internal/agent/tools/aliases.go` - Added sanitization to ResolveToolName
- `internal/agent/utils/tool_id.go` - New SanitizeToolName function
- `internal/agent/utils/tool_name_test.go` - Test coverage (NEW)
