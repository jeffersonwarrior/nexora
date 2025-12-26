# Nexora Development Reflections

Cross-session learnings and patterns discovered during development.

---

## Patterns Discovered

### Database Operations
- SQLite WAL mode essential for concurrent access
- Always use `_journal_mode=WAL&_synchronous=NORMAL` in connection strings
- Test with race detector: `go test -race ./...`

### TUI Development
- Bubble Tea updates are synchronous - batch when possible
- Key bindings conflict across dialogs - check all before adding new
- Status indicators need state machine, not ad-hoc updates

### Agent System
- ProjectID must thread through entire call chain
- Observations capture fails silently if context missing
- Tool execution tracking needs explicit capture points

---

## Decisions Made

### v0.29.1
- Deferred tool consolidation to v0.29.2 (stability first)
- Keyboard shortcuts: ctrl+e for models, removed ctrl+p/ctrl+n from dialogs
- Observations capture: thread ProjectID through SessionAgentOptions

---

## Mistakes to Avoid

1. Adding keyboard shortcuts without checking all dialog handlers
2. Assuming ProjectID flows through - verify explicitly
3. Making TUI changes without testing all dialog states
4. Batch-committing unrelated changes

---

*Updated: Session entries synthesized here periodically*
