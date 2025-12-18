# ADR-002: 100-Line Chunks for VIEW Tool Default

**Date**: 2025-12-18  
**Status**: Accepted  
**Deciders**: Development Team  
**Technical Story**: Context window exhaustion and token management

## Context

The VIEW tool reads and displays file contents with line numbers for code examination. The tool needs to balance two competing concerns:

1. **Completeness**: Show enough context for the agent to understand code
2. **Efficiency**: Avoid exhausting the context window with large files

### The Problem

Initial implementation had a higher default line limit (2000 lines), which caused issues:

- **Context window exhaustion**: Large files consumed significant portions of the 200K token budget
- **Slow responses**: Processing thousands of lines takes time
- **Reduced conversation capacity**: Less room for actual conversation and tool outputs
- **Memory pressure**: Large file contents cached in session state

### Token Math

For a typical Go file:
- Average line: ~40 characters → ~10 tokens
- 2000 lines: ~20K tokens
- Context budget: 200K tokens
- Impact: **10% of context window for a single file view**

With multiple file views, auto-summarization triggers prematurely.

### Real-World Usage Patterns

Analysis of agent behavior showed:
- Agents typically need **local context** around specific functions
- Rarely need entire file at once
- When searching, use grep/glob to locate first, then view specific sections
- Multiple small views preferred over one large view

## Decision

We will **default VIEW tool limit to 100 lines** per call, with an offset parameter to navigate large files.

This means:
- First call shows lines 1-100
- Use `offset=100` to see lines 101-200
- Use `offset=200` to see lines 201-300
- And so on...

Users can still request larger chunks by setting a higher `limit` parameter, but 100 lines is the default.

## Consequences

### Positive

- **Context preservation**: 100 lines ≈ 1K tokens vs 20K tokens (95% reduction)
- **Faster responses**: Less data to process and transmit
- **More conversation room**: More tokens available for actual problem-solving
- **Better agent behavior**: Encourages targeted viewing instead of blind reading
- **Lazy loading**: Only load what's needed, when it's needed
- **Clear messaging**: Agent sees "showing lines 1-100 of 500 total" and knows to request more

### Negative

- **Multiple calls required**: Large files need several VIEW calls
- **Agent learning curve**: Must understand offset parameter
- **Navigation complexity**: Requires calculating offsets
- **Potential for missing context**: Might miss important code between chunks

### Risks

- **Agent confusion**: Might not realize there's more content beyond line 100
  - **Mitigation**: Clear messaging "showing lines 1-100 of 500 total" in output
  - **Mitigation**: Documentation emphasizes using `offset` parameter
  - **Mitigation**: System prompt includes navigation instructions

- **Workflow disruption**: Existing patterns assuming full file views break
  - **Mitigation**: Gradual rollout, monitor agent behavior
  - **Mitigation**: Higher limits available via explicit `limit` parameter
  
- **Missed context**: Critical information might be outside viewed chunk
  - **Mitigation**: Agents have grep/find tools to locate relevant sections
  - **Mitigation**: Can request larger chunks when needed
  - **Mitigation**: Multiple views with different offsets encouraged

## Alternatives Considered

### Option A: Keep Higher Default (2000 lines)

**Description**: Maintain current behavior with 2000-line default

**Pros**:
- Complete file view in one call
- Simple mental model
- No navigation needed

**Cons**:
- High token cost (20K+ tokens per large file)
- Context window exhaustion
- Encourages wasteful viewing
- Slower responses

**Why not chosen**: Token efficiency is critical for long conversations. The cost outweighs convenience.

### Option B: Dynamic Sizing Based on File Size

**Description**: Small files show completely, large files auto-chunk

**Pros**:
- Best of both worlds
- Intelligent behavior
- No navigation for small files

**Cons**:
- Inconsistent behavior (confusing)
- Complexity in implementation
- Hard to predict behavior
- Still consumes tokens for "small" files that add up

**Why not chosen**: Consistency and predictability are more valuable than perceived convenience.

### Option C: Aggressive Limit (50 lines)

**Description**: Even smaller default chunks

**Pros**:
- Maximum token efficiency
- Forces very targeted viewing

**Cons**:
- Too restrictive
- Many files naturally have 50-100 line functions
- Excessive navigation required
- Poor UX

**Why not chosen**: 50 lines is too restrictive. Most functions/classes fit in 100 lines.

### Option D: Summary + Sections

**Description**: Show file outline first, then load sections on demand

**Pros**:
- Best information architecture
- Efficient token use
- Clear navigation

**Cons**:
- Requires new tool or complex VIEW mode
- Parsing complexity
- Language-specific
- Major implementation effort

**Why not chosen**: Too complex for current timeline. Could be future enhancement.

## Implementation Notes

### Files Affected
- `internal/agent/tools/view.go` - Change default `limit` from 2000 to 100
- `internal/agent/tools/view.md` - Update documentation with examples
- System prompt - Add navigation guidance for large files

### Migration Path

1. ✅ Update default limit constant
2. ✅ Update documentation with offset examples
3. ✅ Add "showing lines X-Y of Z total" message to output
4. ✅ Update system prompt with navigation patterns
5. Monitor agent behavior for confusion or issues

### Testing Strategy

- ✅ Unit tests: Verify 100-line default
- ✅ Integration tests: Test offset navigation
- ✅ UX tests: Agent successfully navigates large files
- ✅ Performance tests: Measure token savings

### Code Example

```go
// Before: Default showed 2000 lines
view file_path="/home/user/largefile.go"
// Output: 2000 lines → 20K tokens

// After: Default shows 100 lines
view file_path="/home/user/largefile.go"
// Output: lines 1-100 of 500 total → 1K tokens

// Navigate to next chunk
view file_path="/home/user/largefile.go" offset=100
// Output: lines 101-200 of 500 total → 1K tokens
```

### Rollback Plan

If agents struggle with navigation:
1. Increase default to 200 lines (compromise)
2. Add auto-suggestion in error messages
3. Consider Option D (summary + sections) for future

## References

- [VIEW Tool Documentation](../../internal/agent/tools/view.go)
- [Context Window Management Strategy](../ROADMAP.md#context-window-management)
- [Token Budget Analysis](../../docs/PERFORMANCE_IMPROVEMENTS_2025_12_18.md)

## Revision History

- **2025-12-18**: Initial draft and acceptance
