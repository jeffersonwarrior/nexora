# Immediate Edit Tool Fixes for AI Compatibility

## Top 3 Critical Fixes Needed

### 1. Fix Tab Display vs Reality Mismatch (PRIORITY 1)
**Problem**: VIEW shows `→	` but EDIT expects `	` - this is the #1 cause of failures

**Immediate Fix**:
```go
// Add to edit.go
func normalizeTabIndicators(content string) string {
    // Convert display format back to actual tabs
    content = strings.ReplaceAll(content, "→\t", "\t")
    content = strings.ReplaceAll(content, "→", "\t")
    return content
}

// Modify replaceContent function - add at the very beginning:
func replaceContent(edit editContext, filePath, oldString, newString string, replaceAll bool, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
    // NORMALIZE TAB INDICATORS FIRST
    oldString = normalizeTabIndicators(oldString)
    newString = normalizeTabIndicators(newString)
    
    // ... rest of existing function
}
```

**Testing**:
- Create test file with tabs
- View file and copy text with `→	` indicators
- Attempt edit with copied text
- Verify edit succeeds after normalization

### 2. Add AI Mode Parameter (PRIORITY 2)
**Problem**: Current exact matching too strict for AI

**Immediate Fix**:
```go
// Modify EditParams in edit.go
type EditParams struct {
    FilePath    string `json:"file_path" description:"The absolute path to the file to modify"`
    OldString   string `json:"old_string" description:"The text to replace"`
    NewString   string `json:"new_string" description:"The text to replace it with"`
    ReplaceAll  bool   `json:"replace_all,omitempty" description:"Replace all occurrences of old_string (default false)"`
    AIMode      bool   `json:"ai_mode,omitempty" description:"Enable AI-optimized editing with automatic context expansion and retries"`
}

// Add to replaceContent function after tab normalization:
if params.AIMode {
    // Automatically expand context if minimal
    if strings.Count(params.OldString, "\n") < 2 {
        expanded, err := autoExpandContext(filePath, params.OldString)
        if err == nil {
            params.OldString = expanded
        }
    }
    
    // Lower AIOPS confidence threshold for AI mode
    if edit.aiops != nil {
        resolution, err := edit.aiops.ResolveEdit(edit.ctx, oldContent, params.OldString, params.NewString)
        if err == nil && resolution.Confidence > 0.5 { // Lowered from 0.8
            params.OldString = resolution.ExactOldString
        }
    }
}
```

**Testing**:
- Test edits with minimal context using AI mode
- Verify automatic context expansion works
- Test AIOPS resolution with lower confidence threshold

### 3. Improve Error Messages for AI (PRIORITY 3)
**Problem**: Current errors not actionable for AI

**Immediate Fix**:
```go
// Add to edit_diagnostics.go
func createAIErrorMessage(err error, fileContent, oldString string) string {
    analysis := AnalyzeWhitespaceDifference(fileContent, oldString)
    
    if analysis.HasTabMismatch {
        return fmt.Sprintf("TAB_MISMATCH: Expected %d tabs but found %d. " +
                          "The VIEW tool shows tabs as '→\t' but EDIT needs raw tabs. " +
                          "Try using AI mode (ai_mode=true) for automatic normalization.",
                          analysis.ExpectedTabs, analysis.FoundTabs)
    }
    
    if analysis.HasSpaceMismatch {
        return fmt.Sprintf("SPACE_MISMATCH: Expected %d spaces but found %d. " +
                          "Count spaces carefully or use more surrounding context. " +
                          "AI mode (ai_mode=true) can help with this.",
                          analysis.ExpectedSpaces, analysis.FoundSpaces)
    }
    
    if analysis.PatternNotFound {
        return fmt.Sprintf("PATTERN_NOT_FOUND: The text '%s' was not found. " +
                          "This often happens when copying from VIEW output. " +
                          "Try: 1) Use AI mode (ai_mode=true), " +
                          "2) Include more surrounding context, " +
                          "3) Check for tab/space differences.",
                          truncatePreview(oldString, 50))
    }
    
    return err.Error()
}

// Modify all error returns in edit.go to use:
return fantasy.NewTextErrorResponse(createAIErrorMessage(err, oldContent, params.OldString)), nil
```

**Testing**:
- Force various error conditions
- Verify new error messages are returned
- Test that AI can understand and act on messages

## Quick Wins (Can Implement Immediately)

### 1. Add Simple Context Expansion
```go
// Add to edit.go
func autoExpandContext(filePath, partialString string) (string, error) {
    content, err := os.ReadFile(filePath)
    if err != nil {
        return "", err
    }
    
    fileContent := string(content)
    lines := strings.Split(fileContent, "\n")
    
    // Simple line-based expansion
    for i, line := range lines {
        if strings.Contains(line, partialString) {
            start := max(0, i-2)
            end := min(len(lines), i+3)
            return strings.Join(lines[start:end], "\n"), nil
        }
    }
    
    return partialString, nil // Fallback to original
}
```

### 2. Add Basic Tab Detection to Error Messages
```go
// Add to edit_diagnostics.go
func AnalyzeWhitespaceDifference(fileContent, pattern string) map[string]interface{} {
    fileTabs := strings.Count(fileContent, "\t")
    patternTabs := strings.Count(pattern, "\t")
    fileDisplayTabs := strings.Count(pattern, "→\t")
    
    return map[string]interface{} {
        "has_tab_mismatch": patternTabs != fileTabs || fileDisplayTabs > 0,
        "expected_tabs": fileTabs,
        "found_tabs": patternTabs,
        "display_tabs_found": fileDisplayTabs,
        "pattern_in_file": strings.Contains(fileContent, pattern),
        "pattern_in_file_after_normalization": strings.Contains(fileContent, normalizeTabIndicators(pattern)),
    }
}
```

## Implementation Order

1. **Fix tab normalization** (1-2 hours)
2. **Add AI mode parameter** (2-3 hours)
3. **Improve error messages** (3-4 hours)
4. **Add simple context expansion** (2-3 hours)
5. **Test all changes together** (4-6 hours)

**Total**: 12-18 hours for critical fixes

## Testing Strategy

### Unit Tests to Add
```go
// Test tab normalization
func TestTabNormalization(t *testing.T) {
    tests := []struct {
        input    string
        expected string
    }{
        {"func\tfoo()", "func\tfoo()"}, // Real tab
        {"func→\tfoo()", "func\tfoo()"}, // Display tab
        {"func→foo()", "func\tfoo()"}, // Partial display tab
        {"no\ttabs", "no\ttabs"}, // No change needed
    }
    
    for _, tt := range tests {
        result := normalizeTabIndicators(tt.input)
        require.Equal(t, tt.expected, result)
    }
}

// Test AI mode context expansion
func TestAutoExpandContext(t *testing.T) {
    // Create test file
    tmpFile := createTestFile(t, "line1\nline2\ntarget\nline4\nline5")
    
    expanded, err := autoExpandContext(tmpFile, "target")
    require.NoError(t, err)
    require.Contains(t, expanded, "line2")
    require.Contains(t, expanded, "target")
    require.Contains(t, expanded, "line4")
}
```

### Integration Tests
1. Test full edit workflow with tab-containing files
2. Test AI mode with minimal context
3. Verify error messages are helpful
4. Test backward compatibility

## Monitoring

Add these metrics to track success:
```go
// Add to edit operations
func LogEditOperation(filePath string, success bool, aiMode bool, errorType string) {
    metrics.Increment("edit.operations.total")
    if success {
        metrics.Increment("edit.operations.success")
        if aiMode {
            metrics.Increment("edit.operations.ai_success")
        }
    } else {
        metrics.Increment("edit.operations.failure")
        metrics.Increment(fmt.Sprintf("edit.operations.failure.%s", errorType))
        if aiMode {
            metrics.Increment("edit.operations.ai_failure")
        }
    }
}
```

## Rollout Plan

1. **Develop and test** fixes in isolation
2. **Integrate** all changes
3. **Test** with real AI agents
4. **Deploy** with feature flags
5. **Monitor** success rates
6. **Iterate** based on feedback

## Expected Impact

- **Tab-related failures**: Reduced by 80-90%
- **Overall edit success rate**: Improved by 40-60%
- **AI agent productivity**: Significant improvement
- **User satisfaction**: Increased due to better error messages

## Risk Mitigation

- All changes are backward compatible
- Feature flags allow gradual rollout
- Comprehensive testing before deployment
- Monitoring for quick issue detection
- Rollback plan in place

## Next Steps

1. Implement tab normalization fix
2. Add AI mode parameter
3. Improve error messages
4. Add basic context expansion
5. Write and run tests
6. Deploy to staging for validation
