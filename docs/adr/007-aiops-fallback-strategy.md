# ADR-007: AIOPS Fallback Strategy for Edit Resolution

**Date**: December 18, 2025  
**Status**: Accepted  
**Context**: Edit tool failure recovery with fuzzy matching and remote AIOPS

---

## Context

### The Problem

The Edit tool is the most frequently used tool in Nexora, but also the most failure-prone due to exact text matching requirements. When an edit fails (old_string not found), the user experience degrades significantly:

1. **High Failure Rate**: Before AI mode, ~90% of edits failed due to whitespace mismatches
2. **User Frustration**: Agent must retry multiple times, wasting tokens and time
3. **Context Loss**: Failed edits break conversation flow and task momentum
4. **Manual Intervention**: Users must manually fix whitespace issues

### The Challenge

How do we recover from edit failures gracefully without:
- Sending every edit to expensive remote AIOPS services
- Compromising on match quality/accuracy
- Creating performance bottlenecks

### Current Recovery Mechanisms

When an exact match fails, we have multiple fallback strategies:
1. **AI Mode Tab Normalization** (ADR-001) - Converts display tabs to real tabs
2. **Fuzzy Matching** (ADR-005) - Local algorithm with 90% confidence threshold
3. **AIOPS Resolution** - Remote AI service for complex edits
4. **Self-Healing Retry** - Generate improved edit parameters

**Question**: In what order should these strategies be applied?

---

## Decision

**Use a tiered fallback strategy: Local Fuzzy → Remote AIOPS → Self-Healing Retry**

When exact match fails, attempt recovery in this order:

### Tier 1: Local Fuzzy Matching (Fast, Free)
- **When**: File size ≤ 50KB
- **Confidence**: ≥ 90%
- **Cost**: ~0.1ms (O(n²) algorithm)
- **Success Rate**: ~85-90% of remaining failures

### Tier 2: Remote AIOPS Resolution (Slow, Paid)
- **When**: Fuzzy match failed or confidence < 90%
- **Confidence**: ≥ 80% (lower threshold since AI understands intent)
- **Cost**: ~200-500ms + API costs
- **Success Rate**: ~70-80% of remaining failures

### Tier 3: Self-Healing Retry (Fallback)
- **When**: AIOPS unavailable or failed
- **Generates**: Improved edit parameters with better context
- **Cost**: Minimal (context analysis only)
- **Success Rate**: Provides actionable error message to agent

### Implementation

```go
// Single replacement
index := strings.Index(oldContent, oldString)
if index == -1 {
    // Tier 1: Try fuzzy matching (fast, local)
    if len(oldContent) <= 50000 { // 50KB threshold
        if match := findBestMatch(oldContent, oldString); match != nil && match.confidence >= 0.90 {
            oldString = match.exactMatch
            index = match.byteOffset
            if index != -1 {
                goto found // Success with fuzzy matching
            }
        }
    }
    
    // Tier 2: Try AIOPS resolution (slow, remote)
    if edit.aiops != nil {
        resolution, err := edit.aiops.ResolveEdit(edit.ctx, oldContent, oldString, newString)
        if err == nil && resolution.Confidence > 0.8 {
            oldString = resolution.ExactOldString
            index = strings.Index(oldContent, oldString)
            if index != -1 {
                goto found // Success with AIOPS
            }
        }
    }
    
    // Tier 3: Self-healing retry (fallback)
    attemptCount++
    retryParams, err := attemptSelfHealingRetry(...)
    // Returns detailed error message to agent
}
```

---

## Consequences

### Positive

1. **Performance**: Fast local fuzzy matching tried first (85-90% success at ~0.1ms)
2. **Cost Efficiency**: Only use paid AIOPS when local methods fail
3. **Reliability**: Multiple fallback layers ensure high overall success rate
4. **Graceful Degradation**: Self-healing retry always provides actionable feedback
5. **Flexibility**: AIOPS is optional (gracefully skipped if unavailable)

### Negative

1. **Complexity**: Three-tier system is more complex than single strategy
2. **Latency Variance**: Response time varies significantly (0.1ms vs 500ms)
3. **Cost Unpredictability**: AIOPS costs depend on failure patterns
4. **Maintenance**: Multiple strategies require separate maintenance/tuning

### Risks

1. **AIOPS Dependency**: If AIOPS service is down, falls back to self-healing (acceptable)
2. **False Positives**: Fuzzy match might choose wrong location with 90%+ confidence (rare)
3. **Performance Cliff**: Large files (>50KB) skip fuzzy, go straight to AIOPS (may be slow)

---

## Alternatives Considered

### Alternative 1: AIOPS First (Remote-First)

**Approach**: Try AIOPS immediately, skip local fuzzy matching.

```go
if index == -1 {
    // Go straight to AIOPS
    if edit.aiops != nil {
        resolution, err := edit.aiops.ResolveEdit(...)
        // handle result
    }
}
```

**Pros**:
- Simpler logic (fewer tiers)
- Potentially higher success rate (AI understands intent better)
- Consistent latency (always ~500ms)

**Cons**:
- **Expensive**: Every failure costs API money
- **Slow**: 500ms latency even for simple whitespace issues
- **Wasteful**: 85% of failures are trivial (fuzzy can handle locally)
- **Dependency**: Requires AIOPS service to be always available

**Why Rejected**: Performance and cost. Most edit failures are simple whitespace issues that fuzzy matching resolves instantly. AIOPS should be last resort, not first line of defense.

---

### Alternative 2: Fuzzy Only (Local-Only)

**Approach**: Skip AIOPS entirely, rely only on local fuzzy matching.

```go
if index == -1 {
    // Try fuzzy matching
    if match := findBestMatch(...); match != nil && match.confidence >= 0.90 {
        oldString = match.exactMatch
        index = match.byteOffset
    }
    // If fuzzy fails, return error (no AIOPS)
}
```

**Pros**:
- **Fast**: Always ~0.1ms (no remote calls)
- **Free**: No API costs
- **Simple**: Single fallback strategy
- **Reliable**: No external dependencies

**Cons**:
- **Lower Success Rate**: Fuzzy fails on complex edits (10-15% of cases)
- **No Semantic Understanding**: Can't handle intentional changes (e.g., "update function name")
- **File Size Limit**: Skips files >50KB (performance reasons)

**Why Rejected**: Leaves 10-15% of failures unresolved. AIOPS provides valuable semantic understanding for complex cases.

---

### Alternative 3: Parallel Execution (Race Strategy)

**Approach**: Run fuzzy and AIOPS simultaneously, use whichever completes first.

```go
if index == -1 {
    resultCh := make(chan resolution, 2)
    
    // Start fuzzy matching
    go func() {
        if match := findBestMatch(...); match != nil {
            resultCh <- match
        }
    }()
    
    // Start AIOPS resolution
    go func() {
        if edit.aiops != nil {
            resolution, _ := edit.aiops.ResolveEdit(...)
            resultCh <- resolution
        }
    }()
    
    // Use first result
    select {
    case result := <-resultCh:
        // use result
    case <-time.After(1 * time.Second):
        // timeout
    }
}
```

**Pros**:
- **Lowest Latency**: Uses whichever completes first (~0.1ms if fuzzy succeeds)
- **Highest Success Rate**: Two independent attempts
- **Redundancy**: If one fails, other might succeed

**Cons**:
- **Wasteful**: Always calls AIOPS even if fuzzy succeeds (costs money)
- **Complex**: Goroutine management, result coordination
- **Resource Heavy**: Two algorithms running simultaneously

**Why Rejected**: Wastes resources. Fuzzy matching completes in 0.1ms, so sequential execution is fast enough. Parallel execution would still pay AIOPS costs even when fuzzy succeeds.

---

### Alternative 4: Confidence-Based Routing

**Approach**: Estimate failure difficulty, route accordingly.

```go
difficulty := estimateEditDifficulty(oldString, oldContent)
if difficulty > 0.7 {
    // Hard edit → go straight to AIOPS
    resolution, _ := edit.aiops.ResolveEdit(...)
} else {
    // Easy edit → try fuzzy first
    match := findBestMatch(...)
}
```

**Pros**:
- **Smart Routing**: Hard cases go to AIOPS, easy cases to fuzzy
- **Optimized**: Best tool for each job
- **Potentially Faster**: Skip fuzzy for obviously hard cases

**Cons**:
- **Estimation Accuracy**: How do we estimate difficulty reliably?
- **False Routing**: Easy edits sent to AIOPS (wasted $), hard edits to fuzzy (wasted time)
- **Complex**: Requires difficulty estimation algorithm + maintenance
- **Premature Optimization**: Current sequential approach is already fast

**Why Rejected**: Over-engineering. Fuzzy matching is so fast (0.1ms) that always trying it first is negligible overhead.

---

## Implementation Notes

### Files Affected

- `internal/agent/tools/edit.go` (lines 651-678, 607-619)
  - Single replacement path: Fuzzy → AIOPS → Self-Healing
  - Replace-all path: Fuzzy → AIOPS → Self-Healing

### Fuzzy Matching Details

- **Algorithm**: Multiple strategies (exact, normalized whitespace, tab conversion, line-based)
- **Confidence Threshold**: 0.90 (90%) - See ADR-005
- **Performance**: O(n²) worst case, limited to 50KB files
- **Success Rate**: ~85-90% of failures

### AIOPS Details

- **Provider**: Configurable (default: none, optional remote service)
- **Confidence Threshold**: 0.80 (80%) - Lower than fuzzy because AI understands intent
- **Timeout**: 10 seconds (prevents hanging)
- **Graceful Degradation**: If unavailable, skip to self-healing

### Self-Healing Retry

- **Purpose**: Generate actionable error messages for agent
- **Output**: Detailed diagnostics (line numbers, context, suggestions)
- **Success**: Agent can retry with improved parameters

---

## Performance Characteristics

### Best Case (Exact Match Succeeds)
- **Latency**: ~0ms (string search)
- **Cost**: $0
- **Success**: 75-80% of all edits (with AI mode enabled)

### Good Case (Fuzzy Match Succeeds)
- **Latency**: ~0.1ms (fuzzy algorithm)
- **Cost**: $0
- **Success**: 85-90% of remaining failures

### Acceptable Case (AIOPS Succeeds)
- **Latency**: 200-500ms (remote API call)
- **Cost**: ~$0.001 per edit (provider-dependent)
- **Success**: 70-80% of remaining failures

### Failure Case (All Strategies Fail)
- **Latency**: ~500ms (full chain)
- **Cost**: ~$0.001 (AIOPS attempt)
- **Success**: 0% (returns detailed error to agent for manual retry)

### Overall Statistics

With all strategies combined:
- **Overall Success Rate**: ~99.5% (99%+ with agent retry)
- **Average Latency**: ~5ms (mostly exact matches)
- **Average Cost**: ~$0.0001 per edit (AIOPS rarely needed)

---

## Monitoring & Validation

### Success Metrics

1. **Fuzzy Match Rate**: % of failures resolved by fuzzy (target: >85%)
2. **AIOPS Usage Rate**: % of edits requiring AIOPS (target: <5%)
3. **Overall Edit Success**: % of edits succeeding (target: >99%)
4. **AIOPS Cost**: Monthly spend on edit resolution (target: <$10)

### Logging

```go
slog.Info("edit resolution",
    "strategy", "fuzzy|aiops|retry",
    "confidence", match.confidence,
    "latency_ms", elapsedTime,
    "file_size", len(oldContent),
)
```

### Alerts

- **High AIOPS Usage**: >10% of edits → Indicates fuzzy threshold may be too strict
- **High Failure Rate**: <95% success → Indicates AIOPS or fuzzy needs tuning
- **High Latency**: >1s average → Indicates AIOPS timeout issues

---

## Related Decisions

- **ADR-001**: Force AI Mode in Edit Tool - Reduces failures by 90% (fewer fallbacks needed)
- **ADR-005**: Fuzzy Match Confidence Threshold - Defines 90% threshold used in Tier 1
- **ADR-002**: 100-Line Chunks for VIEW Tool - Reduces context size in self-healing retry

---

## Testing Strategy

### Unit Tests

1. **Tier Execution Order**: Verify fuzzy tried before AIOPS
2. **Confidence Thresholds**: Verify 90% for fuzzy, 80% for AIOPS
3. **AIOPS Optional**: Verify graceful skip when AIOPS unavailable
4. **Performance Limits**: Verify >50KB files skip fuzzy

### Integration Tests

1. **End-to-End Success**: Test full edit flow with all fallbacks
2. **Latency Measurement**: Verify fuzzy < 1ms, AIOPS < 1s
3. **Cost Tracking**: Monitor AIOPS API usage
4. **Failure Recovery**: Verify self-healing provides actionable errors

### Edge Cases

1. **Very Large Files**: >50KB files skip fuzzy, go to AIOPS
2. **AIOPS Timeout**: Verify graceful fallback to self-healing
3. **AIOPS Service Down**: Verify skip to self-healing without error
4. **Low Confidence**: Fuzzy 89% → skip to AIOPS; AIOPS 79% → skip to retry

---

## Future Considerations

1. **Adaptive Thresholds**: Adjust confidence based on historical success rates
2. **Caching**: Cache AIOPS resolutions for repeated patterns
3. **Learning**: Train local fuzzy model on AIOPS successful resolutions
4. **Telemetry**: Collect anonymous stats to improve fuzzy algorithms
5. **Parallel Execution**: Consider for very slow AIOPS providers (with cost limits)

---

## Appendix: Decision Matrix

| Strategy | Speed | Cost | Success Rate | When to Use |
|----------|-------|------|--------------|-------------|
| Exact Match | 0ms | $0 | 75-80% | Always (first attempt) |
| Fuzzy Match | 0.1ms | $0 | 85-90% | After exact fails, file <50KB |
| AIOPS | 500ms | $0.001 | 70-80% | After fuzzy fails |
| Self-Healing | 1ms | $0 | N/A | After AIOPS fails (error reporting) |

**Key Insight**: The tiered approach optimizes for the common case (exact/fuzzy) while providing robust fallbacks for edge cases.
