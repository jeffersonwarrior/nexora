# Nexora Agentic Features Roadmap

## v0.29.2 Agentic Features (Released)

### 1. TMUX-Powered Interactive Editing (Phase 4.5)

- Human-in-the-Loop Paradigm: Real-time collaboration between AI and human
- Persistent bash shell sessions via TMUX
- Interactive editor control (vi, helix, emacs)
- 10 concurrent TMUX sessions per conversation
- Special key mapping (, , , , etc.)
- Human can watch AI work and intervene instantly (Ctrl+C to stop, demonstrate, resume)
- Session reuse across multiple commands

### 2. Tool Architecture Simplification (Phase 4.5)

- Reduced from 20+47 tools → 9 core tools + ~30 aliases
- Smart fetch with MCP auto-routing and context-aware handling
- Token optimization with session-scoped tmp file fallback
- 42% reduction in total invocation names (67 → 39)

### 3. Multi-Agent Orchestration System (RC1 Plan)

From ORCHESTRATION-PLAN.md and SWARM-TRACKER.md:
- Parallel worker execution: 6+ concurrent agents on different features
- Test-Driven Development workflow: Mandatory blocking validation
- Competitive planning: 2 workers create competing plans, best wins (for complex features using Opus)
- Phase-based execution: Phases 0-5 with dependency management
  - Phase 0: Fix failing tests
  - Phase 3: Test coverage (36% → 50%)
  - Phase 4: Tool consolidation
  - Phase 5: TUI enhancements + critical bugs

### 4. Unified Delegate with Resource Pool (Phase 5, Feature 16)

- Dynamic agent spawning based on available resources
- Resource calculation (CPU, memory per agent)
- Queue with 30min timeout
- Prevents resource exhaustion

### 5. Auto-LSP Detection (Phase 5, Feature 14)

- Go (go.mod → gopls)
- Rust (Cargo.toml → rust-analyzer)
- Node (package.json → typescript-language-server)
- Python (pyproject.toml → pyright)
- Installation command generation
- Auto-enable based on project type

---

## v0.29.3 Planned Agentic Features (GitHub Issues Created)

### 1. Task Graph Enrichment (Issue #6)

- Dependency visualization (ASCII graph)
- Parallel execution detection
- Critical path analysis
- Prevents circular dependencies

### 2. Session Checkpoint System (Issue #7)

- Auto-save session state every N messages
- Crash recovery and rollback
- Manual checkpoint creation
- State verification on restore

### 3. Prompt Library Restoration (Issue #3)

- TUI dashboard for browsing prompts
- SQLite with FTS5 full-text search
- Rating/usage tracking
- Performance metrics (tokens, latency)

---

## v0.29.4-0.29.5 Future Agentic Features

### 1. Agent-to-Agent Communication (Issue #8)

- Direct messaging protocol between agents
- Capability discovery and advertising
- Shared context and state
- Message routing and delivery

### 2. Protocol Composition & Conflict Resolution (Issue #9)

- Detect conflicting operations across agents
- Merge/sequence operations automatically
- Multi-agent workflow composition
- Priority-based conflict resolution

---

## Key Philosophy Shifts

> Phase 4.5 Insight: "Tools for AI efficiency, bash for everything else, TMUX for human collaboration"

### Human-in-the-Loop

Changed from "AI edits, user reviews after" → "Human watches AI edit in real-time"

### Benefits

- Faster correction when AI goes wrong
- Collaborative demonstration of correct approaches
- AI learns from successful patterns
- Parallel work (human + AI simultaneously)

---

All of these features support the core goal: transforming Nexora into an autonomous, multi-agent orchestration platform with human oversight capabilities.
