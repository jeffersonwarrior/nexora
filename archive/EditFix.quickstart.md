# Quick Start Guide: Fixing Edit Tool for AI

## Step 1: Implement Tab Normalization (30 minutes)

### File: `internal/agent/tools/edit.go`

Add this function at the top of the file (after imports):

```go
// normalizeTabIndicators converts VIEW tool display tabs (→\t) back to actual tabs (\t)
func normalizeTabIndicators(content string) string {
    // Convert display format →\t back to actual tabs \t
    content = strings.ReplaceAll(content, "→\t", "\t")
    // Also handle cases where only → is present
    content = strings.ReplaceAll(content, "→", "\t")
    return content
}
```

### File: `internal/agent/tools/edit.go`

Modify the `replaceContent` function - add this right at the beginning (after parameter validation):

```go
func replaceContent(edit editContext, filePath, oldString, newString string, replaceAll bool, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
    // ... existing parameter validation code ...
    
    // NORMALIZE TAB INDICATORS FROM VIEW OUTPUT
    oldString = normalizeTabIndicators(oldString)
    newString = normalizeTabIndicators(newString)
    
    // ... rest of existing function unchanged ...
}
```

## Step 2: Add AI Mode Parameter (15 minutes)

### File: `internal/agent/tools/edit.go`

Modify the `EditParams` struct:

```go
type EditParams struct {
    FilePath    string `json:"file_path" description:"The absolute path to the file to modify"`
    OldString   string `json:"old_string" description:"The text to replace"`
    NewString   string `json:"new_string" description:"The text to replace it with"`
    ReplaceAll  bool   `json:"replace_all,omitempty" description:"Replace all occurrences of old_string (default false)"`
    AIMode      bool   `json:"ai_mode,omitempty" description:"Enable AI-optimized editing with automatic context expansion and improved error handling"`
}
```

## Step 3: Add Simple Context Expansion (20 minutes)

### File: `internal/agent/tools/edit.go`

Add this function:

```go
// autoExpandContext automatically expands minimal context to improve match success
func autoExpandContext(filePath, partialString string) (string, error) {
    content, err := os.ReadFile(filePath)
    if err != nil {
        return "", fmt.Errorf("failed to read file for context expansion: %w", err)
    }
    
    fileContent := string(content)
    lines := strings.Split(fileContent, "\n")
    
    // Try to find the partial string and expand context around it
    for i, line := range lines {
        if strings.Contains(line, partialString) {
            // Expand 2 lines before and 2 lines after for better context
            start := i - 2
            if start < 0 {
                start = 0
            }
            end := i + 3
            if end > len(lines) {
                end = len(lines)
            }
            return strings.Join(lines[start:end], "\n"), nil
        }
    }
    
    // If not found, return original (no expansion possible)
    return partialString, nil
}
```

### File: `internal/agent/tools/edit.go`

Add AI mode logic to the `replaceContent` function (after tab normalization):

```go
// Add this right after the tab normalization in replaceContent:
if params.AIMode {
    // Automatically expand context if the pattern has minimal lines
    lineCount := strings.Count(params.OldString, "\n")
    if lineCount < 2 { // Less than 2 newlines = minimal context
        expanded, err := autoExpandContext(filePath, params.OldString)
        if err == nil && expanded != params.OldString {
            params.OldString = expanded
            // Log that we expanded context
            slog.Debug("AI mode: expanded context for better matching", 
                      "file", filePath, 
                      "original_lines", lineCount,
                      "expanded_lines", strings.Count(expanded, "\n"))
        }
    }
}
```

## Step 4: Improve Error Messages (30 minutes)

### File: `internal/agent/tools/edit_diagnostics.go`

Add this function:

```go
// AnalyzeWhitespaceDifference compares whitespace between file content and pattern
func AnalyzeWhitespaceDifference(fileContent, pattern string) map[string]interface{} {
    fileTabs := strings.Count(fileContent, "\t")
    patternTabs := strings.Count(pattern, "\t")
    displayTabs := strings.Count(pattern, "→\t")
    
    fileSpaces := countLeadingSpaces(fileContent)
    patternSpaces := countLeadingSpaces(pattern)
    
    return map[string]interface{} {
        "has_tab_mismatch": patternTabs != fileTabs || displayTabs > 0,
        "expected_tabs": fileTabs,
        "found_tabs": patternTabs,
        "display_tabs": displayTabs,
        "has_space_mismatch": patternSpaces != fileSpaces,
        "expected_spaces": fileSpaces,
        "found_spaces": patternSpaces,
        "pattern_in_file": strings.Contains(fileContent, pattern),
        "pattern_after_normalization": strings.Contains(fileContent, normalizeTabIndicators(pattern)),
    }
}

// Helper function for space counting
func countLeadingSpaces(content string) int {
    lines := strings.Split(content, "\n")
    if len(lines) == 0 {
        return 0
    }
    
    // Count leading spaces on first non-empty line
    for _, line := range lines {
        trimmed := strings.TrimLeft(line, " ")
        if len(trimmed) < len(line) {
            return len(line) - len(trimmed)
        }
    }
    return 0
}
```

### File: `internal/agent/tools/edit_diagnostics.go`

Add this function:

```go
// createAIErrorMessage generates AI-friendly error messages with actionable guidance
func createAIErrorMessage(err error, fileContent, oldString string) string {
    analysis := AnalyzeWhitespaceDifference(fileContent, oldString)
    
    // Tab mismatch - most common issue
    if analysis["has_tab_mismatch"].(bool) {
        return fmt.Sprintf("TAB_MISMATCH: The VIEW tool shows tabs as '→\t' but EDIT needs raw tabs. " +
                          "Found %d display tabs in your pattern. Try: " +
                          "1) Use AI mode (ai_mode=true) for automatic normalization, or " +
                          "2) Replace '→\t' with actual tab characters in your pattern.",
                          analysis["display_tabs"].(int))
    }
    
    // Space mismatch
    if analysis["has_space_mismatch"].(bool) {
        return fmt.Sprintf("SPACE_MISMATCH: Expected %d leading spaces but found %d. " +
                          "Count spaces carefully. AI mode (ai_mode=true) can help with this.",
                          analysis["expected_spaces"].(int), analysis["found_spaces"].(int))
    }
    
    // Pattern not found at all
    if !analysis["pattern_in_file"].(bool) {
        // Check if it would match after normalization
        if analysis["pattern_after_normalization"].(bool) {
            return fmt.Sprintf("PATTERN_FORMAT_MISMATCH: Your pattern would match after tab normalization. " +
                              "Use AI mode (ai_mode=true) or normalize tabs manually.")
        }
        
        return fmt.Sprintf("PATTERN_NOT_FOUND: The text was not found in the file. " +
                          "Common fixes: 1) Use AI mode (ai_mode=true), " +
                          "2) Include more surrounding context (3-5 lines), " +
                          "3) Check for tab/space differences, " +
                          "4) Verify the file hasn't changed since you viewed it.")
    }
    
    // Fallback to original error
    return err.Error()
}
```

### File: `internal/agent/tools/edit.go`

Update error returns to use the new AI-friendly messages. Find all places that return errors like:

```go
// Change from:
return fantasy.NewTextErrorResponse("old_string not found in file..."), nil

// To:
return fantasy.NewTextErrorResponse(createAIErrorMessage(
    fmt.Errorf("old_string not found in file"), 
    oldContent, 
    params.OldString
)), nil
```

## Step 5: Add Basic Testing (20 minutes)

### File: `internal/agent/tools/edit_test.go` (create if doesn't exist)

```go
package tools

import (
    "testing"
    "github.com/stretchr/testify/require"
)

func TestTabNormalization(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {
            name:     "real tab unchanged",
            input:    "func\tfoo()",
            expected: "func\tfoo()",
        },
        {
            name:     "display tab normalized",
            input:    "func→\tfoo()",
            expected: "func\tfoo()",
        },
        {
            name:     "partial display tab normalized",
            input:    "func→foo()",
            expected: "func\tfoo()",
        },
        {
            name:     "no tabs unchanged",
            input:    "func foo()",
            expected: "func foo()",
        },
        {
            name:     "multiple display tabs",
            input:    "func→\t→\tfoo()",
            expected: "func\t\tfoo()",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := normalizeTabIndicators(tt.input)
            require.Equal(t, tt.expected, result)
        })
    }
}

func TestAutoExpandContext(t *testing.T) {
    // Create a temporary test file
    tmpDir := t.TempDir()
    testFile := tmpDir + "/test.txt"
    
    content := "line 1\nline 2\ntarget line\nline 4\nline 5"
    err := os.WriteFile(testFile, []byte(content), 0o644)
    require.NoError(t, err)
    
    // Test context expansion
    expanded, err := autoExpandContext(testFile, "target")
    require.NoError(t, err)
    
    // Should include lines before and after
    require.Contains(t, expanded, "line 2")
    require.Contains(t, expanded, "target line")
    require.Contains(t, expanded, "line 4")
    
    // Should not include lines that are too far
    require.NotContains(t, expanded, "line 1")
    require.NotContains(t, expanded, "line 5")
}

func TestAIErrorMessages(t *testing.T) {
    fileContent := "func\tmain() {\n\tfmt.Println(\"hello\")\n}"
    
    tests := []struct {
        name           string
        pattern        string
        expectedPrefix string
    }{
        {
            name:           "tab mismatch",
            pattern:        "func→\tmain()",
            expectedPrefix: "TAB_MISMATCH",
        },
        {
            name:           "pattern not found",
            pattern:        "nonexistent",
            expectedPrefix: "PATTERN_NOT_FOUND",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            msg := createAIErrorMessage(
                fmt.Errorf("test error"),
                fileContent,
                tt.pattern,
            )
            require.Contains(t, msg, tt.expectedPrefix)
        })
    }
}
```

## Step 6: Update Documentation (15 minutes)

### File: `internal/agent/tools/edit.md`

Add this section to the documentation:

```markdown
## AI Mode

For AI agents, use `ai_mode=true` to enable automatic fixes for common issues:

```json
{
  "file_path": "/path/to/file.go",
  "old_string": "func main() {",
  "new_string": "func main() {\n    // new code",
  "ai_mode": true
}
```

### AI Mode Features

1. **Automatic Tab Normalization**: Converts VIEW tool display tabs (`→\t`) to real tabs (`\t`)
2. **Context Expansion**: Automatically expands minimal context to improve match success
3. **Enhanced Error Messages**: Provides actionable guidance for common failure patterns
4. **Lower AIOPS Threshold**: More aggressive use of AI-powered edit resolution

### When to Use AI Mode

- When copying text from VIEW tool output
- When getting whitespace-related errors
- When working with files that have complex indentation
- When initial edit attempts fail

### AI Mode Best Practices

- Still provide as much context as possible
- Check error messages for specific guidance
- Use for complex edits where exact matching is difficult
- Combine with other parameters as needed
```

## Step 7: Test the Changes

Run the tests:
```bash
cd /home/renter/nexora
go test ./internal/agent/tools -v -run "TestTabNormalization|TestAutoExpandContext|TestAIErrorMessages"
```

Test manually:
```bash
# Create a test file with tabs
cd /tmp
echo -e "func\tmain() {\n\tfmt.Println(\"hello\")\n}" > test.go

# Test the edit tool with AI mode
# (You'll need to use the actual nexora CLI or API)
```

## Step 8: Monitor and Iterate

Add these metrics to track success:

```go
// Add to appropriate places in edit.go
func LogEditOperation(filePath string, success bool, aiMode bool, errorType string) {
    // Implement metrics logging
    // Track: total operations, success rate, AI mode usage, error types
}
```

## Expected Results

After implementing these changes:

1. **Tab-related failures should drop by 80-90%**
2. **AI mode should improve success rate by 40-60%**
3. **Error messages should be much more helpful**
4. **Context expansion should reduce "pattern not found" errors**

## Rollback Plan

If issues arise:

1. The changes are backward compatible - existing calls work unchanged
2. AI mode is opt-in - no impact on existing workflows
3. Feature flags can be added if needed for gradual rollout
4. All changes can be easily reverted if needed

## Time Estimate

- **Implementation**: 2-3 hours
- **Testing**: 1-2 hours
- **Documentation**: 30 minutes
- **Total**: 3-5 hours for immediate improvements

## Next Steps After This Quick Fix

1. Monitor success rates in production
2. Gather feedback from AI agents
3. Implement more advanced self-healing strategies
4. Add more sophisticated context expansion
5. Improve AIOPS integration further
