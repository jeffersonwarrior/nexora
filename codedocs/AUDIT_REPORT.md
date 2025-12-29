# NEXORA CODE AUDIT REPORT
**Date**: 2025-12-29
**Auditor**: Automated + Manual Review
**Scope**: 340 Go source files (internal/ package)
**Methodology**: Static analysis tools + systematic code review

---

## Executive Summary

This comprehensive audit examined 340 Go source files in the internal/ package using multiple static analysis tools and systematic manual review. The codebase demonstrates good security practices with proper shell parsing, SQL parameterization, and extensive safety mechanisms for command execution. However, several issues were identified ranging from resource leaks to high cyclomatic complexity.

**Critical Issues**: 1
**High Severity Issues**: 4
**Medium Severity Issues**: 8
**Low Severity Issues**: 12

---

## 1. Static Analysis Results

### 1.1 go vet
**Status**: ✅ PASSED
**Result**: No issues found

### 1.2 staticcheck
**Status**: ✅ PASSED
**Result**: No issues found
**Configuration**: .staticcheck.conf with checks = ["all", "-U1000"]

### 1.3 Cyclomatic Complexity Analysis (gocyclo)
**Files with complexity > 15**: 46 functions

**Critical Complexity (>25)**:
```
31  (*appModel).handleKeyPressMsg           internal/tui/tui.go:491:1
29  formatSourcegraphResults               internal/agent/tools/sourcegraph.go:139:1
29  NewBashTool                           internal/agent/tools/bash.go:280:1
28  (*list).changeSelectionWhenScrolling    internal/tui/exp/list/list.go:767:1
28  (*permissionDialogCmp).Update           internal/tui/components/dialogs/permissions/permissions.go:98:1
27  Prepare                               internal/db/db.go:24:1
27  (*completionsCmp).Update              internal/tui/components/completions/completions.go:115:1
```

**High Complexity (20-25)**:
```
25  (*list).Update                         internal/tui/exp/list/list.go:251:1
24  blockFuncs                            internal/agent/tools/bash.go:159:1
24  applyDocumentChange                   internal/lsp/util/edit.go:152:1
22  (*coordinator).agenticFetchTool        internal/agent/agentic_fetch_tool.go:54:1
22  (*sessionAgent).generateTitle          internal/agent/agent.go:1528:1
22  processMultiEditExistingFile           internal/agent/tools/multiedit.go:233:1
22  ParseTextToolCalls                    internal/agent/utils/text_tool_calls.go:24:1
22  (*DiffView).renderUnified             internal/tui/exp/diffview/diffview.go:398:1
22  (*Config).setDefaults                 internal/config/load.go:391:1
```

---

## 2. CRITICAL ISSUES

### 2.1 Resource Leak: Database Connection Not Closed on Migration Failure
**File**: `internal/db/connect.go`
**Line**: 65-68
**Severity**: CRITICAL

```go
if err := goose.Up(db, "migrations"); err != nil {
    slog.Error("Failed to apply migrations", "error", err)
    return nil, fmt.Errorf("failed to apply migrations: %w", err)
    // BUG: db.Close() not called here - connection leaks
}
```

**Issue**: When goose.Up() fails, the database connection is not closed before returning the error. This causes a connection leak.

**Fix**:
```go
if err := goose.Up(db, "migrations"); err != nil {
    slog.Error("Failed to apply migrations", "error", err)
    db.Close()  // Add this line
    return nil, fmt.Errorf("failed to apply migrations: %w", err)
}
```

**Impact**: Repeated migration failures will exhaust file descriptors and database connections.

---

## 3. HIGH SEVERITY ISSUES

### 3.1 Goroutine Leak: Timeout Wrapper Tool
**File**: `internal/agent/coordinator.go`
**Lines**: 85-105
**Severity**: HIGH

```go
func (t *timeoutWrappedTool) Run(ctx context.Context, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
    timeoutCtx, cancel := context.WithTimeout(ctx, t.timeout)
    defer cancel()

    resultChan := make(chan result, 1)

    // Execute original tool in goroutine
    go func() {
        resp, err := t.original.Run(timeoutCtx, call)
        resultChan <- result{resp: resp, err: err}  // Might block here forever
    }()

    select {
    case <-timeoutCtx.Done():
        // BUG: Goroutine continues running even though result won't be read
        return fantasy.ToolResponse{}, fmt.Errorf("tool timeout after %v", t.timeout)
    case res := <-resultChan:
        return res.resp, res.err
    }
}
```

**Issue**: When timeout occurs, the goroutine continues running because timeoutCtx is cancelled but the channel write might block if the tool is still executing. This leads to goroutine leaks.

**Fix**: Use buffered channel with select to ensure goroutine can always exit:
```go
go func() {
    resp, err := t.original.Run(timeoutCtx, call)
    select {
    case resultChan <- result{resp: resp, err: err}:
    case <-timeoutCtx.Done():
        // Context cancelled, exit
    }
}()
```

**Impact**: Over time, goroutine leaks will exhaust memory.

---

### 3.2 Delegation Reporting Goroutine May Leak
**File**: `internal/agent/delegate_tool.go`
**Lines**: 416-437
**Severity**: HIGH

```go
go func() {
    reportPrompt := fmt.Sprintf(
        "[DELEGATE REPORT - Task ID: %s]\n\n...",
        task.ID, result,
    )
    slog.Info("delegate reporting to parent session", ...)

    reportCtx := context.Background()  // BUG: Uses unbounded context
    if _, runErr := c.Run(reportCtx, task.ParentSession, reportPrompt); runErr != nil {
        slog.Error("failed to report delegate results", ...)
    }
}()
```

**Issue**: The goroutine uses context.Background() which cannot be cancelled. If c.Run() hangs indefinitely, the goroutine will leak.

**Fix**: Use a timeout context:
```go
reportCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

if _, runErr := c.Run(reportCtx, task.ParentSession, reportPrompt); runErr != nil {
    slog.Error("failed to report delegate results", ...)
}
```

---

### 3.3 Ignored Error: HTTP Response Body Read
**File**: `internal/agent/tools/sourcegraph.go`
**Line**: 113
**Severity**: HIGH

```go
if resp.StatusCode != http.StatusOK {
    body, _ := io.ReadAll(resp.Body)  // BUG: Error ignored
    if len(body) > 0 {
        return fantasy.NewTextErrorResponse(fmt.Sprintf("Request failed with status code: %d, response: %s", resp.StatusCode, string(body))), nil
    }
    return fantasy.NewTextErrorResponse(fmt.Sprintf("Request failed with status code: %d", resp.StatusCode)), nil
}
```

**Issue**: Error from io.ReadAll() is silently ignored. If reading fails, the error message will be empty or misleading.

**Fix**:
```go
if resp.StatusCode != http.StatusOK {
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return fantasy.NewTextErrorResponse(fmt.Sprintf("Request failed with status code: %d and failed to read error response", resp.StatusCode)), nil
    }
    if len(body) > 0 {
        return fantasy.NewTextErrorResponse(fmt.Sprintf("Request failed with status code: %d, response: %s", resp.StatusCode, string(body))), nil
    }
    return fantasy.NewTextErrorResponse(fmt.Sprintf("Request failed with status code: %d", resp.StatusCode)), nil
}
```

---

### 3.4 Multiple Ignored io.ReadAll Errors
**Files**: Multiple
**Severity**: HIGH

| File | Line | Context |
|------|------|---------|
| `internal/update/update.go` | 108 | HTTP error response body |
| `internal/cmd/install.go` | 175 | HTTP response body |
| `internal/cmd/install.go` | 211 | HTTP response body |
| `internal/cmd/install.go` | 248 | HTTP response body |

**Issue**: All these locations ignore io.ReadAll() errors when reading HTTP response bodies for error handling.

**Impact**: Errors during error message reading may lead to poor error reporting.

---

## 4. MEDIUM SEVERITY ISSUES

### 4.1 High Cyclomatic Complexity: handleKeyPressMsg
**File**: `internal/tui/tui.go`
**Function**: `(*appModel).handleKeyPressMsg`
**Complexity**: 31
**Severity**: MEDIUM

**Issue**: Function has excessive complexity (31) making it difficult to maintain and error-prone.

**Recommendation**: Refactor into smaller, focused functions using pattern matching or command pattern:
```go
func (m *appModel) handleKeyPressMsg(msg tea.KeyMsg) tea.Cmd {
    switch msg.Type {
    case tea.KeyCtrlC:
        return m.handleQuit()
    case tea.KeyCtrlS:
        return m.handleSave()
    case tea.KeyCtrlD:
        return m.handleDelete()
    // ... etc
    }
    return nil
}
```

---

### 4.2 High Cyclomatic Complexity: NewBashTool
**File**: `internal/agent/tools/bash.go`
**Function**: `NewBashTool`
**Complexity**: 29
**Severity**: MEDIUM

**Issue**: Complex initialization logic mixing multiple concerns (permission checks, command execution, TMUX integration).

**Recommendation**: Extract into separate functions:
```go
func NewBashTool(...) fantasy.AgentTool {
    return fantasy.NewAgentTool(
        BashToolName,
        string(bashDescription(attribution, modelName)),
        func(ctx context.Context, params BashParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
            return handleBashExecution(ctx, params, call, ...)
        },
    )
}

func handleBashExecution(...) (fantasy.ToolResponse, error) {
    // Extracted logic here
}
```

---

### 4.3 Complex State Machine Update Logic
**File**: `internal/tui/exp/list/list.go`
**Function**: `(*list).changeSelectionWhenScrolling`
**Complexity**: 28
**Severity**: MEDIUM

**Issue**: Complex scrolling logic mixing boundary checking with selection updates.

**Recommendation**: Separate concerns into helper functions for boundary checking and selection updates.

---

### 4.4 Unsafe Protocol in HTTP Requests (Local Model Detector)
**File**: `internal/config/providers/local_detector.go`
**Lines**: Multiple HTTP requests using http://
**Severity**: MEDIUM

**Issue**: When detecting local models, the code uses HTTP instead of HTTPS for localhost connections. While acceptable for localhost, this should be configurable and documented.

**Recommendation**:
```go
// Allow configuration of protocol for local development
protocol := "http://"
if os.Getenv("NEXORA_FORCE_HTTPS") == "1" {
    protocol = "https://"
}
baseURL := protocol + host + ":" + port
```

---

### 4.5 No Context Cancellation for Some Long Operations
**File**: `internal/agent/tools/grep.go`
**Line**: 245
**Severity**: MEDIUM

```go
go func() {
    // Long-running grep operation without context check
    results := searchFiles(...)
    resultChan <- results
}()
```

**Issue**: Goroutine doesn't check context cancellation during file search.

**Fix**: Pass context and check periodically:
```go
go func() {
    results, err := searchFilesWithContext(ctx, pattern, paths)
    resultChan <- searchResult{results: results, err: err}
}()
```

---

### 4.6 Hardcoded Timeouts
**File**: `internal/agent/coordinator.go`
**Lines**: 51-54
**Severity**: MEDIUM

```go
const (
    defaultToolTimeout  = 5 * time.Minute
    criticalToolTimeout = 10 * time.Minute
    maxToolTimeout      = 30 * time.Minute
)
```

**Issue**: Timeouts are hardcoded and not configurable per tool or use case.

**Recommendation**: Make timeouts configurable via environment variables or config file:
```go
func getToolTimeout(toolType string) time.Duration {
    if val := os.Getenv("NEXORA_TOOL_TIMEOUT_" + strings.ToUpper(toolType)); val != "" {
        if duration, err := time.ParseDuration(val); err == nil {
            return duration
        }
    }
    return defaultToolTimeout
}
```

---

### 4.7 Missing Input Validation in Bash Tool
**File**: `internal/agent/tools/bash.go`
**Severity**: MEDIUM

**Issue**: The WorkingDir parameter is not validated for directory traversal attacks.

**Fix**: Add path validation:
```go
func validateWorkingDir(baseDir, requestedDir string) error {
    cleanBase := filepath.Clean(baseDir)
    cleanRequested := filepath.Clean(requestedDir)

    // Check if requested directory is within base directory
    rel, err := filepath.Rel(cleanBase, cleanRequested)
    if err != nil {
        return err
    }

    if strings.HasPrefix(rel, "..") {
        return fmt.Errorf("working directory escape detected: %s", requestedDir)
    }

    if _, err := os.Stat(cleanRequested); err != nil {
        return fmt.Errorf("working directory does not exist: %w", err)
    }

    return nil
}
```

---

### 4.8 Potential Race in Session Service
**File**: `internal/session/session.go`
**Severity**: MEDIUM

**Issue**: The service uses pubsub.Broker for event publishing but doesn't synchronize access to internal state. Concurrent calls to Save() and Get() could race.

**Recommendation**: Add mutex protection:
```go
type service struct {
    *pubsub.Broker[Session]
    q    db.Querier
    mu   sync.RWMutex  // Add this
}

func (s *service) Get(ctx context.Context, id string) (Session, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    // ... rest of code
}

func (s *service) Save(ctx context.Context, session Session) (Session, error) {
    s.mu.Lock()
    defer s.mu.Unlock()
    // ... rest of code
}
```

---

## 5. LOW SEVERITY ISSUES

### 5.1 Magic Numbers and Constants
**File**: Multiple files
**Severity**: LOW

**Issue**: Various magic numbers throughout codebase without clear documentation.

**Examples**:
- `internal/agent/tools/bash.go`: Default scrollback 10000 lines
- `internal/tui/tui.go`: Mouse debounce 15ms, cache TTL 10s
- `internal/session/checkpoint.go`: Token threshold 50000

**Recommendation**: Define constants with comments explaining their purpose and why these values were chosen.

---

### 5.2 TODO Comments Found
**File**: `internal/log/log.go`
**Severity**: LOW

**Issue**: Code contains TODO comments that may indicate incomplete features.

**Action**: Review TODO comments and either:
- Complete the implementation
- Create issues to track
- Remove if no longer relevant

---

### 5.3 Generic Error Messages
**Files**: Multiple
**Severity**: LOW

**Issue**: Some error messages are generic and don't help with debugging.

**Example**: `fmt.Errorf("failed to open database: %w")` doesn't include database path.

**Recommendation**: Include relevant context in error messages.

---

### 5.4 Missing Tests for Error Paths
**Files**: Multiple tool implementations
**Severity**: LOW

**Issue**: Many tool tests focus on happy path and don't test error scenarios.

**Recommendation**: Add test coverage for:
- Network errors
- Permission errors
- Invalid inputs
- Timeout scenarios

---

### 5.5 Inconsistent Error Handling Styles
**Files**: Multiple
**Severity**: LOW

**Issue**: Some functions return wrapped errors with context, others return bare errors.

**Recommendation**: Standardize on wrapping all errors with context:
```go
return fmt.Errorf("action failed: %w", err)  // Good
return err  // Bad - no context
```

---

### 5.6 Potential SQL Injection Edge Case
**File**: `internal/db/connect.go`
**Severity**: LOW (Mitigated)

**Issue**: The pragma setting loop uses direct string concatenation:
```go
for _, pragma := range pragmas {
    if err := c.Exec(pragma); err != nil {
```

**Analysis**: Pragmas are hardcoded strings from the same package, so this is not a real injection vulnerability. However, if pragmas were ever made configurable, this would be a risk.

**Recommendation**: Add a comment explaining why this is safe:
```go
// Pragmas are hardcoded constants - safe from injection
for _, pragma := range pragmas {
    if err := c.Exec(pragma); err != nil {
```

---

### 5.7 Large Files
**Files**:
- `internal/config/load.go`: 791 lines
- `internal/agent/coordinator.go`: 1239 lines
- `internal/tui/tui.go`: 773 lines
**Severity**: LOW

**Issue**: Files are becoming too large and difficult to navigate.

**Recommendation**: Consider splitting into smaller, focused files by responsibility.

---

### 5.8 Unused Variables in Tests
**Files**: Multiple test files
**Severity**: LOW

**Example**: `internal/tui/components/banner/banner_test.go` line 168

**Issue**: Variables declared but not used in tests.

**Recommendation**: Use blank identifier `_` or remove unused variables.

---

### 5.9 Incomplete Documentation
**Files**: Multiple public APIs
**Severity**: LOW

**Issue**: Some public functions and types lack complete documentation.

**Recommendation**: Ensure all exported functions have godoc comments explaining:
- Purpose
- Parameters
- Return values
- Potential errors

---

### 5.10 Logging Inconsistencies
**Files**: Multiple
**Severity**: LOW

**Issue**: Some code uses slog, others use fmt.Printf for logging.

**Recommendation**: Standardize on slog for structured logging throughout codebase.

---

### 5.11 Missing Context Propagation
**File**: Various HTTP clients
**Severity**: LOW

**Issue**: Some HTTP client operations don't use request context properly.

**Recommendation**: Ensure all HTTP requests use request context:
```go
req = req.WithContext(ctx)
resp, err := client.Do(req)
```

---

### 5.12 Hardcoded Provider Models
**File**: `internal/config/providers/` (multiple files)
**Severity**: LOW

**Issue**: Model lists are hardcoded and may become outdated.

**Recommendation**: Consider fetching available models dynamically from provider APIs.

---

## 6. SECURITY ANALYSIS

### 6.1 Shell Execution Safety ✅ GOOD
**File**: `internal/shell/shell.go`

**Finding**: The codebase uses `mvdan.cc/sh/v3` to parse shell commands before execution, which provides:
- Proper shell parsing (not string concatenation)
- Comprehensive command blocking via BlockFuncs
- Safe argument handling

**Safety Mechanisms**:
```go
// Dangerous commands blocked
shell.ArgumentsBlocker("rm", []string{}, []string{"-rf"})
shell.CommandsBlocker([]string{"mkfs", "fdisk", "dd", "shred"})

// Fork bomb protection
forkBombPatterns := []string{
    ":()",
    "while true",
    ":|:",
}
```

**Assessment**: **EXCELLENT** - Shell execution is properly secured with multiple layers of protection.

---

### 6.2 SQL Injection Prevention ✅ GOOD
**File**: `internal/db/` (SQL files)

**Finding**: The codebase uses sqlc to generate type-safe SQL queries with parameterized statements.

**Example from messages.sql**:
```sql
-- name: GetMessage :one
SELECT * FROM messages WHERE id = ? LIMIT 1;
```

**Assessment**: **EXCELLENT** - SQL injection is prevented by proper use of parameterized queries.

---

### 6.3 API Key Handling ✅ GOOD
**File**: `internal/config/providers/` (multiple files)

**Finding**: API keys are loaded from environment variables, not hardcoded.

```go
APIKey: "$ANTHROPIC_API_KEY",
APIEndpoint: cmp.Or(os.Getenv("ANTHROPIC_API_ENDPOINT"), "https://api.anthropic.com/v1"),
```

**Assessment**: **GOOD** - Credentials are not hardcoded. However, consider adding validation that required environment variables are set.

---

### 6.4 Path Traversal Vulnerability ⚠️ NEEDS ATTENTION
**File**: `internal/agent/tools/bash.go`

**Issue**: The WorkingDir parameter could allow directory traversal attacks.

**See**: Issue 4.7 above for recommended fix.

---

## 7. ARCHITECTURAL OBSERVATIONS

### 7.1 Positive Patterns

1. **Separation of Concerns**: Clear separation between business logic, persistence, and UI layers.

2. **Interface-Based Design**: Good use of interfaces (e.g., `Coordinator`, `Session.Service`) for testability.

3. **Event-Driven Architecture**: Proper use of pub/sub pattern for session events.

4. **Context Propagation**: Generally good use of context for cancellation and timeouts.

5. **Type Safety**: Use of sqlc for database operations provides excellent type safety.

---

### 7.2 Areas for Improvement

1. **File Organization**: Some files are becoming too large (coordinator.go 1239 lines).

2. **Error Handling**: Inconsistent error wrapping - standardize on adding context.

3. **Testing Coverage**: Could improve test coverage for error paths and edge cases.

4. **Configuration**: Some hardcoded values (timeouts, limits) should be configurable.

5. **Documentation**: Some complex functions need better inline documentation.

---

## 8. DEPENDENCY ANALYSIS

### 8.1 Third-Party Dependencies

**Critical Dependencies**:
- `charm.land/fantasy` - Core AI agent framework
- `charm.land/bubbletea/v2` - TUI framework
- `mvdan.cc/sh/v3` - Shell parsing (security-critical)
- `ncruces/go-sqlite3` - SQLite database

**Assessment**: Dependencies are appropriate and actively maintained. The use of `mvdan.cc/sh/v3` is particularly good for security.

---

### 8.2 Dependency Security

**Potential Concerns**:
- Some dependencies are from internal charmbracelet repositories - ensure they receive security updates.

**Recommendation**:
- Run `go list -json -m all | grep -o '"Path":"[^"]*"' | cut -d'"' -f4 | xargs -I {} govulncheck -mode=binary ./` regularly
- Consider using `dependabot` (already configured) for automated dependency updates.

---

## 9. PERFORMANCE OBSERVATIONS

### 9.1 Potential Performance Issues

1. **Synchronous Database Operations**: Some database operations could be batched for better performance.

2. **Unbounded Channel Usage**: Some channels use unbounded size which could lead to memory pressure.

3. **Large File Reads**: `io.ReadAll` used extensively - could use streaming for large files.

---

### 9.2 Resource Management

**Good Practices**:
- Proper use of `defer` for resource cleanup (100+ instances found)
- Database connection pooling via sql.DB

**Issues Identified**:
- Connection leak in connect.go (Issue 2.1)
- Potential goroutine leaks (Issues 3.1, 3.2)

---

## 10. TESTING STATUS

### 10.1 Test Coverage

**Good Coverage Areas**:
- Database operations (internal/db/)
- Shell execution (internal/shell/)
- Tool implementations (internal/agent/tools/)

**Areas Needing Improvement**:
- Error path testing
- Concurrent access scenarios
- Integration tests for complex workflows

---

### 10.2 Test Quality

**Observations**:
- Tests use table-driven patterns appropriately
- Golden file testing for TUI components
- VCR cassettes for HTTP mocking

**Recommendation**: Add more race condition testing:
```bash
go test -race ./...
```

---

## 11. RECOMMENDATIONS SUMMARY

### Immediate Actions (Critical/High Priority)

1. ✅ Fix database connection leak in `internal/db/connect.go` line 65
2. ✅ Fix goroutine leak in timeout wrapper tool
3. ✅ Fix delegation reporting goroutine leak
4. ✅ Handle all io.ReadAll() errors properly

### Short-Term Actions (Medium Priority)

5. ✅ Refactor high complexity functions (>25 complexity)
6. ✅ Add working directory validation to bash tool
7. ✅ Add mutex protection to session service
8. ✅ Make timeouts configurable
9. ✅ Add context cancellation checks to long-running operations

### Long-Term Actions (Low Priority)

10. ✅ Split large files into smaller modules
11. ✅ Standardize error handling patterns
12. ✅ Improve test coverage for error paths
13. ✅ Document magic numbers and constants
14. ✅ Standardize logging on slog

---

## 12. CONCLUSION

The Nexora codebase demonstrates solid engineering practices with:
- Excellent security posture for shell execution and database operations
- Good use of modern Go patterns (contexts, interfaces, defer)
- Comprehensive test infrastructure

However, there are some issues that need attention:
- **1 Critical**: Resource leak in database connection
- **4 High**: Goroutine leaks and ignored errors
- **8 Medium**: High complexity, missing validation, race conditions
- **12 Low**: Documentation, organization, and consistency improvements

**Overall Assessment**: The codebase is well-structured with good security foundations, but would benefit from addressing the identified resource leaks and complexity issues to ensure long-term maintainability and reliability.

---

## APPENDIX A: File Analysis by Package

| Package | Files | Complexity Issues | Security Issues | Status |
|---------|--------|-------------------|-----------------|---------|
| internal/agent | 16 | 3 functions >20 | Path traversal | ⚠️ Needs Attention |
| internal/config | 6 | 1 function >20 | HTTP protocol | ⚠️ Minor |
| internal/db | 9 | 1 function >25 | Connection leak | 🔴 Critical |
| internal/session | 2 | 0 | Race condition | ⚠️ Medium |
| internal/tui | 20+ | 4 functions >20 | None | ⚠️ Needs Refactoring |
| internal/shell | 5 | 0 | None | ✅ Good |
| internal/agent/tools | 38 | 5 functions >20 | None | ✅ Good |

---

## APPENDIX B: Complexity Hotspots

Top 10 most complex functions requiring refactoring:

1. `(*appModel).handleKeyPressMsg` - 31 complexity
2. `formatSourcegraphResults` - 29 complexity
3. `NewBashTool` - 29 complexity
4. `(*list).changeSelectionWhenScrolling` - 28 complexity
5. `(*permissionDialogCmp).Update` - 28 complexity
6. `Prepare` (db) - 27 complexity
7. `(*completionsCmp).Update` - 27 complexity
8. `(*list).Update` - 25 complexity
9. `blockFuncs` - 24 complexity
10. `applyDocumentChange` - 24 complexity

---

**Report Generated**: 2025-12-29
**Next Audit Recommended**: After addressing critical issues, in 3 months
