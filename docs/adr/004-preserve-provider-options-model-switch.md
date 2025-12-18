# ADR-004: Preserve Provider Options When Switching Models

**Date**: December 18, 2025  
**Status**: Accepted  
**Context**: Summarization reliability with Cerebras GLM-4.6 model

---

## Context

### The Problem

When the Nexora agent reaches 80% of its context window, it triggers automatic summarization to maintain conversation continuity. This summarization process faced reliability issues with specific models:

1. **Cerebras GLM-4.6** - Large (32B+) parameter model that struggles with summarization at 180K tokens (90% context threshold)
2. **Model Switch Solution** - Use smaller, more reliable model (llama3.1-8b) for summarization only

### The Question

When switching from the large model to the small model for summarization, should we:
- **A) Preserve provider options** (temperature, topP, etc.) from the original model?
- **B) Clear provider options** and use defaults for the small model?

### Why It Matters

Provider options control model behavior:
- **Temperature**: Controls randomness (0.0 = deterministic, 2.0 = creative)
- **Top-P**: Nucleus sampling threshold
- **Max Tokens**: Output length limit
- **Frequency/Presence Penalties**: Repetition control

These settings may have been carefully tuned by the user for their workflow. Losing them during model switches could cause:
- Unexpected behavior changes
- User confusion ("Why did the model suddenly become more/less creative?")
- Loss of conversation context/tone

---

## Decision

**Preserve provider options when switching models for summarization.**

When Cerebras GLM-4.6 or similar models require switching to smaller models for summarization, we maintain the user's configured provider options (temperature, topP, maxTokens, etc.) on the new model.

### Implementation

```go
// Use smaller model for Cerebras summarization (more reliable at 180K)
var agent fantasy.Agent
var summarizationModel Model
var summarizationOpts fantasy.ProviderOptions
if strings.Contains(a.largeModel.Model.Model(), "glm-4.6") || 
   strings.Contains(a.largeModel.Model.Provider(), "cerebras") {
    agent = fantasy.NewAgent(a.smallModel.Model,
        fantasy.WithSystemPrompt(string(summaryPrompt)),
    )
    summarizationModel = a.smallModel
    // Preserve provider options when switching models
    // (temperature, topP, etc. are generally compatible across models)
    summarizationOpts = opts  // <-- KEY LINE
} else {
    agent = fantasy.NewAgent(a.largeModel.Model,
        fantasy.WithSystemPrompt(string(summaryPrompt)),
    )
    summarizationModel = a.largeModel
    summarizationOpts = opts
}
```

---

## Consequences

### Positive

1. **Consistency** - User settings remain active across model switches
2. **Predictability** - Model behavior stays consistent (temperature, creativity, etc.)
3. **User Trust** - No "magical" behavior changes that confuse users
4. **Generality** - Most provider options are compatible across models (temperature, topP, etc.)

### Negative

1. **Potential Incompatibility** - Some options might not work on smaller models:
   - Smaller models may have different max token limits
   - Some providers have model-specific options
2. **Suboptimal Settings** - Settings tuned for large models might not be optimal for small models:
   - Temperature 0.8 on GLM-4.6 ≠ Temperature 0.8 on llama3.1-8b
3. **Hidden Behavior** - Users might not realize a model switch occurred during summarization

### Risks

1. **Provider Validation** - If small model rejects large model's options, summarization could fail
   - **Mitigation**: Fantasy library handles validation, falls back to defaults
2. **Performance Impact** - Incompatible options might slow down small model
   - **Mitigation**: Small models are typically faster, so small overhead acceptable
3. **Quality Degradation** - Wrong settings could produce poor summaries
   - **Mitigation**: Summarization prompt is carefully crafted to guide output regardless of temperature

---

## Alternatives Considered

### Alternative 1: Clear Provider Options (Use Defaults)

**Approach**: When switching to small model, use default provider options.

```go
summarizationOpts = fantasy.ProviderOptions{} // Clear all options
```

**Pros**:
- Guaranteed compatibility (no option conflicts)
- Small model runs with "factory defaults"
- Simpler logic (no option copying)

**Cons**:
- User settings lost during summarization
- Behavior changes unexpectedly
- Inconsistent model personality/tone
- User confusion ("Why is the model acting different?")

**Why Rejected**: Breaks user expectations. If user sets temperature=0.1 for deterministic behavior, they expect that across all operations.

---

### Alternative 2: Model-Specific Option Mapping

**Approach**: Maintain a mapping of compatible options per model, translate accordingly.

```go
optionsMap := map[string]ProviderOptions{
    "glm-4.6": {Temperature: 0.8, TopP: 0.9},
    "llama3.1-8b": {Temperature: 0.7, TopP: 0.85}, // Adjusted for small model
}
summarizationOpts = optionsMap[summarizationModel.Model.Model()]
```

**Pros**:
- Optimized settings per model
- No compatibility issues
- Best possible summarization quality

**Cons**:
- Complex maintenance (new models = new mappings)
- Magic translations (0.8 → 0.7) are arbitrary
- Doesn't respect user's actual settings
- Overkill for summarization (transient operation)

**Why Rejected**: Over-engineering for a background operation. User settings should be respected.

---

### Alternative 3: Conditional Preservation (Whitelist Approach)

**Approach**: Only preserve "safe" options (temperature, topP), clear others (maxTokens, etc.).

```go
summarizationOpts = fantasy.ProviderOptions{
    Temperature: opts.Temperature,
    TopP: opts.TopP,
    // maxTokens, penalties, etc. cleared
}
```

**Pros**:
- Balances compatibility and consistency
- Preserves core behavior settings
- Reduces risk of option conflicts

**Cons**:
- Arbitrary line between "safe" and "unsafe"
- Still loses some user settings
- More complex logic

**Why Rejected**: Still loses user intent. Better to preserve all and let provider validate.

---

## Implementation Notes

### Files Affected

- `internal/agent/agent.go` - Line ~947 (already implemented)
  - Preserves `opts` as `summarizationOpts` during model switch
  - Used when calling `agent.Chat()` for summarization

### Migration Required

None - already implemented in codebase as of December 2025.

### Testing Strategy

1. **Unit Tests**:
   - Verify options preserved across model switches
   - Test with various option combinations (temperature, topP, maxTokens)
   - Verify summarization succeeds with preserved options

2. **Integration Tests**:
   - Trigger summarization with Cerebras GLM-4.6 at 80% context
   - Verify llama3.1-8b receives same provider options
   - Check summary quality with various temperature settings

3. **Edge Cases**:
   - Options with very high/low values (temperature 0.0, 2.0)
   - Provider-specific options (e.g., Anthropic's `thinking`)
   - Options incompatible with small models (verify graceful fallback)

---

## Monitoring & Validation

### Success Metrics

- **Summarization Success Rate**: >99% (provider handles incompatible options gracefully)
- **User Complaints**: Zero reports of "model behavior changed during conversation"
- **Summary Quality**: Comparable quality with preserved vs. default options

### Error Scenarios

If provider rejects options:
1. Fantasy library falls back to defaults automatically
2. Summarization proceeds with default settings
3. Log warning: "Provider options adjusted for model compatibility"

---

## Related Decisions

- **ADR-003**: Auto-Summarization at 80% Context - Defines when this model switch occurs
- **ADR-006**: Environment Detection - Provides runtime info for model selection

---

## References

- Issue: Cerebras GLM-4.6 summarization failures at 90% context threshold
- Solution: Use llama3.1-8b for Cerebras summarization (faster, more reliable)
- Code: `internal/agent/agent.go:934-956`

---

## Appendix: Common Provider Options

| Option | Description | Typical Range | Compatibility |
|--------|-------------|---------------|---------------|
| Temperature | Randomness/creativity | 0.0 - 2.0 | ✅ All models |
| TopP | Nucleus sampling | 0.0 - 1.0 | ✅ All models |
| MaxTokens | Output length limit | 1 - context_size | ⚠️ Model-dependent |
| FrequencyPenalty | Penalize repetition | -2.0 - 2.0 | ✅ Most models |
| PresencePenalty | Encourage new topics | -2.0 - 2.0 | ✅ Most models |
| TopK | Sampling constraint | 1 - 100 | ⚠️ Provider-dependent |

**Legend**:
- ✅ Highly compatible across models
- ⚠️ May require adjustment per model

---

## Future Considerations

1. **Explicit Model Switch Notification** - Consider logging model switches for transparency
2. **Per-Model Option Profiles** - Allow users to define option sets per model (future UX enhancement)
3. **Option Validation** - Pre-validate options before model switch (fail fast if incompatible)
