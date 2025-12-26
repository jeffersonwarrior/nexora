# Model Selection for Agents

## Haiku is Sufficient For (Default)
- File discovery, keyword searching, structure mapping
- Pattern matching, API discovery, feature location
- Bug FINDING (patterns, race conditions, code smells)
- Simple refactoring tasks
- Documentation reading and summarization
- Test execution and output parsing

## Use Sonnet For
- Complex multi-file refactoring
- API design decisions
- Code review with nuanced feedback
- Integration work spanning multiple systems

## Use Opus For
- Bug FIXING (judgment about correct solution required)
- Security AUDITS (adversarial reasoning needed)
- Safety-critical paths (auth, crypto, data handling)
- Root cause analysis of complex issues
- Architectural decisions with long-term impact
- Novel problem-solving without clear precedent

## Auto-escalate Rule
If Haiku returns uncertainty or empty results, automatically retry with Sonnet.
If Sonnet struggles, escalate to Opus.

## Cost Reference
- Haiku: ~3-10x cheaper than Opus
- 80-90% of exploration tasks achieve same quality with Haiku
- Reserve Opus for judgment-heavy work
