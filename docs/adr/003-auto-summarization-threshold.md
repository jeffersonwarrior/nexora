# ADR-003: Auto-Summarization at 80% Context Window

**Date**: 2025-12-18  
**Status**: Accepted  
**Deciders**: Development Team  
**Technical Story**: Context window management and conversation continuation

## Context

Nexora operates within a finite context window (typically 128K-200K tokens depending on the model). As conversations progress, the context accumulates:

- User messages
- Assistant responses  
- Tool calls and outputs
- System prompts
- File contents
- Error messages

### The Problem

Without active context management:
1. **Hard limit**: Conversations hit the context window limit and crash
2. **Degraded performance**: Models perform worse near capacity
3. **Cost escalation**: Full context sent with every API call
4. **Lost history**: No graceful degradation when limit is reached

### Token Growth Pattern

Typical conversation trajectory:
```
Turn 1: 10K tokens (system prompt + initial message)
Turn 5: 30K tokens (with tool outputs)
Turn 10: 60K tokens (viewing files, making edits)
Turn 20: 120K tokens (multiple file operations)
Turn 25: 160K tokens (approaching limit)
Turn 30: 200K tokens → CRASH
```

### Summarization Strategy

When context is full:
1. Identify older messages that can be summarized
2. Generate concise summaries of message clusters
3. Replace original messages with summaries
4. Preserve recent context for continuity
5. Continue conversation with reduced token count

## Decision

We will **trigger auto-summarization when context window usage reaches 80%** (or 20% remaining capacity).

This threshold provides:
- **Safety margin**: 20% buffer before hard limit
- **Processing time**: Room for summarization operation itself
- **Grace period**: Multiple turns available before limit

The summarization process:
1. Monitor token count after each turn
2. Calculate percentage: `currentTokens / maxTokens`
3. If ≥ 80%, trigger summarization
4. Use fast summarizer model (Cerebras, xAI)
5. Preserve last N turns (configurable, default 5)

## Consequences

### Positive

- **Conversation continuity**: No hard crashes at context limit
- **Better UX**: Seamless experience, user unaware of limit
- **Cost efficiency**: Summarized context is cheaper to process
- **Scalability**: Supports longer, more complex tasks
- **Graceful degradation**: Older context summarized, recent preserved
- **Proactive management**: Prevents emergency situations

### Negative

- **Information loss**: Summaries inevitably lose details
- **Summarization cost**: Extra API call and processing
- **Latency spike**: User experiences delay during summarization
- **Context discontinuity**: Summaries may miss nuances
- **Complexity**: More moving parts, more failure modes

### Risks

- **Premature summarization**: 80% might be too early, causing unnecessary summarization
  - **Mitigation**: Threshold tuned based on empirical data
  - **Mitigation**: Fast models (Cerebras) minimize cost
  - **Mitigation**: Can be adjusted per-deployment
  
- **Poor summaries**: Summary quality varies by model
  - **Mitigation**: Use tested fast models (Cerebras, xAI)
  - **Mitigation**: Preserve recent context (last 5 turns)
  - **Mitigation**: Summarization prompt optimized
  
- **Cascading summarization**: Multiple summarization rounds lose too much
  - **Mitigation**: Monitor summarization frequency
  - **Mitigation**: Consider persisting important context externally
  - **Mitigation**: Future: Use memory system for long-term storage

- **Race conditions**: Token count calculation inaccurate
  - **Mitigation**: Conservative estimates
  - **Mitigation**: 20% buffer handles estimation errors

## Alternatives Considered

### Option A: 90% Threshold

**Description**: Wait until 90% full before summarizing

**Pros**:
- Less frequent summarization
- More complete context preserved longer
- Lower costs

**Cons**:
- Narrower safety margin (10% buffer)
- Higher risk of hitting hard limit
- Summarization itself might push over limit
- Less time to recover from errors

**Why not chosen**: 10% buffer is too risky. Summarization operation needs room to execute.

### Option B: 70% Threshold

**Description**: Trigger earlier at 70% capacity

**Pros**:
- Large safety margin
- Plenty of room for summarization
- Low risk of hitting limit

**Cons**:
- Premature summarization
- Higher costs (more frequent summarization)
- Information loss earlier than necessary
- Poor UX (delays more often)

**Why not chosen**: Too conservative. Wastes context capacity and increases costs unnecessarily.

### Option C: Fixed Message Count

**Description**: Summarize after N messages regardless of token count

**Pros**:
- Simple logic
- Predictable behavior
- No token counting needed

**Cons**:
- Ignores actual token usage
- Some messages are 100 tokens, others 10K tokens
- Inefficient (might summarize unnecessarily)
- Might hit limit despite summarization

**Why not chosen**: Token count is what matters, not message count. This approach is blind to actual resource usage.

### Option D: Lazy Summarization (At 95%)

**Description**: Only summarize when absolutely necessary

**Pros**:
- Maximum context preservation
- Minimal cost
- Rare summarization

**Cons**:
- Emergency-mode behavior (rushed summarization)
- High risk of hitting hard limit
- No room for errors or edge cases
- Panic mode UX

**Why not chosen**: Creates stressful situations. Proactive management is better than reactive crisis handling.

## Implementation Notes

### Files Affected
- `internal/agent/agent.go` - Implement summarization trigger logic
- `internal/agent/summarizer.go` - Summarization configuration
- `internal/config/config.go` - Add threshold configuration option

### Configuration

```toml
[agent]
# Trigger summarization at 80% context usage
summarization_threshold = 0.80

# Preserve last N turns (don't summarize recent context)
summarization_preserve_turns = 5

# Fast model for summarization
summarization_provider = "cerebras"
summarization_model = "llama3.1-8b"
```

### Migration Path

1. ✅ Implement token tracking in agent
2. ✅ Add threshold check after each turn
3. ✅ Integrate summarization routine
4. ✅ Test with various conversation patterns
5. Monitor production usage and adjust

### Testing Strategy

- ✅ Unit tests: Threshold calculation
- ✅ Integration tests: Full summarization flow
- ✅ Stress tests: Conversations approaching limit
- ✅ Performance tests: Summarization latency
- ✅ Cost analysis: Summarization vs full context

### Pseudocode

```go
func (a *Agent) afterTurn() {
    currentTokens := a.calculateTokens()
    maxTokens := a.getModelMaxTokens()
    usage := float64(currentTokens) / float64(maxTokens)
    
    if usage >= 0.80 {
        log.Info("Context at 80%, triggering summarization")
        a.summarize()
    }
}

func (a *Agent) summarize() {
    // Keep last 5 turns
    recentTurns := a.messages[len(a.messages)-10:]
    oldTurns := a.messages[:len(a.messages)-10]
    
    // Summarize old turns
    summary := a.generateSummary(oldTurns)
    
    // Replace with summary
    a.messages = append([]Message{summary}, recentTurns...)
}
```

### Rollback Plan

If summarization causes issues:
1. Increase threshold to 90% (more conservative)
2. Increase preserved turns to 10 (more context)
3. Disable auto-summarization (manual only)
4. Investigate model quality or prompt issues

## References

- [Summarization Implementation](../../internal/agent/summarizer.go)
- [Context Window Analysis](../../docs/ROADMAP.md#context-window-management)
- [Fast Model Performance](../../internal/agent/summarizer.go#L14-L25)
- [Token Budget Documentation](../../docs/PERFORMANCE_IMPROVEMENTS_2025_12_18.md)

## Revision History

- **2025-12-18**: Initial draft and acceptance
