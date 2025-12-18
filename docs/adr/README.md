# Architecture Decision Records (ADRs)

This directory contains Architecture Decision Records (ADRs) for the Nexora project. ADRs document important architectural decisions, explaining the context, the decision made, and its consequences.

## What is an ADR?

An Architecture Decision Record captures a single architectural decision and its rationale. It's a lightweight documentation format that helps teams understand why certain technical choices were made.

## Format

Each ADR follows a standard template:
- **Status**: Proposed, Accepted, Deprecated, or Superseded
- **Context**: The issue or challenge being addressed
- **Decision**: The chosen solution
- **Consequences**: Positive, negative, and risks
- **Alternatives Considered**: Other options that were evaluated

## Index of ADRs

### Accepted

- [ADR-001: Force AI Mode in Edit Tool](001-force-ai-mode-edit-tool.md) - Default to AI-assisted editing to reduce whitespace failures
- [ADR-002: 100-Line Chunks for VIEW Tool](002-view-tool-100-line-chunks.md) - Limit view output to prevent context exhaustion
- [ADR-003: Auto-Summarization at 80% Context](003-auto-summarization-threshold.md) - Trigger summarization before context window fills
- [ADR-004: Preserve Provider Options on Model Switch](004-preserve-provider-options-model-switch.md) - Keep user settings when switching models for summarization
- [ADR-005: Fuzzy Match Confidence Threshold](005-fuzzy-match-confidence.md) - Require 90%+ confidence for fuzzy string matching
- [ADR-006: Environment Detection in System Prompt](006-environment-detection.md) - Include runtime environment details in prompts
- [ADR-007: AIOPS Fallback Strategy](007-aiops-fallback-strategy.md) - Tiered recovery for edit failures (fuzzy 	 AIOPS 	 retry)

## Creating a New ADR

1. Copy `template.md` to a new file with the next number
2. Fill in the template sections
3. Submit as a pull request or commit directly
4. Update this README to include the new ADR in the index

## When to Write an ADR

Write an ADR when:
- Making a significant architectural decision
- Choosing between multiple technical approaches
- Establishing a pattern or convention
- Changing an existing architectural decision
- Resolving a contentious technical debate

## References

- [Documenting Architecture Decisions](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions) by Michael Nygard
- [ADR GitHub Organization](https://adr.github.io/)
