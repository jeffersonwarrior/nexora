# Nexora v0.29.1-RC1 Swarm Orchestration Tracker

## Overall Status: INITIALIZING

**Target:** Complete Phases 3-5 (18 features) with TDD workflow
**Success Criteria:** ≥50% coverage, all tests pass, build succeeds

---

## PHASE 3: Test Coverage (36% → 50%)

### ✅ Feature 1: TUI Chat Component Tests
- **Package:** `internal/tui/components/chat/messages`
- **Target:** 3.2% → 30%
- **Priority:** CRITICAL
- **Model:** sonnet
- **Status:** PENDING
- **Tests Required:**
  - [ ] Thinking animation display logic
  - [ ] Message rendering with markdown
  - [ ] Citation handling
  - [ ] Footer state management
  - [ ] Empty thinking content (edge case)
  - [ ] Long markdown content (edge case)
  - [ ] Multiple citations (edge case)
  - [ ] State transitions (edge case)

### ✅ Feature 2: TUI Chat Page Tests
- **Package:** `internal/tui/page/chat`
- **Target:** 8.1% → 30%
- **Priority:** CRITICAL
- **Model:** sonnet
- **Status:** PENDING
- **Tests Required:**
  - [ ] "/" command trigger logic
  - [ ] Editor value retrieval
  - [ ] Chat page state transitions
  - [ ] Keyboard shortcut handling
  - [ ] Empty editor "/" trigger (edge case)
  - [ ] Non-empty editor "/" passthrough (edge case)

### ✅ Feature 3: Agent Tools Tests
- **Package:** `internal/agent/tools`
- **Target:** 33.8% → 50%
- **Priority:** CRITICAL
- **Model:** sonnet
- **Status:** PENDING
- **Tests Required:**
  - [ ] Bash tool execution and error handling
  - [ ] Fetch operations with timeout
  - [ ] Edit operations with validation
  - [ ] File operations (read, write, glob, grep)
  - [ ] MCP tool integration
  - [ ] Tool parameter validation (edge case)
  - [ ] Error handling edge cases
  - [ ] Timeout behavior (edge case)
  - [ ] Resource cleanup (edge case)

### ⏳ Feature 4: Core Agent Tests
- **Package:** `internal/agent`
- **Target:** 27.6% → 50%
- **Priority:** HIGH
- **Model:** sonnet
- **Depends on:** Feature 3
- **Status:** BLOCKED (waiting for Feature 3)
- **Tests Required:**
  - [ ] Conversation state management
  - [ ] Tool orchestration flow
  - [ ] ProjectID threading
  - [ ] Message handling
  - [ ] LLM integration mocks
  - [ ] State transitions (edge case)
  - [ ] Error recovery (edge case)
  - [ ] SessionAgentOptions initialization (edge case)

### ✅ Feature 5: CLI Commands Tests
- **Package:** `internal/cmd`
- **Target:** 26.8% → 40%
- **Priority:** HIGH
- **Model:** sonnet
- **Status:** PENDING
- **Tests Required:**
  - [ ] Run command logic
  - [ ] Import/export functionality
  - [ ] Indexing operations
  - [ ] Error handling

### ⏳ Feature 6: Agent Delegation Tests
- **Package:** `internal/agent/delegation`
- **Target:** 56.1% → 55% (already passing!)
- **Priority:** MEDIUM
- **Model:** haiku
- **Status:** PENDING
- **Tests Required:**
  - [ ] Pool management
  - [ ] Agent registry
  - [ ] Completion tracking
  - [ ] Queue operations

### ⏳ Feature 7: Agent Memory Tests
- **Package:** `internal/agent/memory`
- **Target:** 50.7% → 55% (almost there!)
- **Priority:** MEDIUM
- **Model:** haiku
- **Status:** PENDING
- **Tests Required:**
  - [ ] Memory system gaps

### ⏳ Feature 8: Database Tests
- **Package:** `internal/db`
- **Target:** 32.8% → 55%
- **Priority:** MEDIUM
- **Model:** haiku
- **Status:** PENDING
- **Tests Required:**
  - [ ] Undertested queries
  - [ ] Migration rollback

---

## PHASE 4: Tool Consolidation

### ⏳ Feature 9: Bash Tool Consolidation
- **Files:** `bash.go`, `bash_monitored.go`
- **Priority:** HIGH
- **Model:** sonnet
- **Depends on:** Feature 3
- **Status:** BLOCKED (waiting for Feature 3)
- **Completion Criteria:**
  - [ ] Tests for standard and AI-monitored modes
  - [ ] Mode detection logic implemented
  - [ ] Old bash.go deleted
  - [ ] Both modes produce correct results

### ⏳ Feature 10: Fetch Tool Consolidation
- **Files:** `fetch.go`, `web_fetch.go`, `agentic_fetch_tool.go`
- **Priority:** HIGH
- **Model:** sonnet
- **Depends on:** Feature 3
- **Status:** BLOCKED (waiting for Feature 3)
- **Completion Criteria:**
  - [ ] Tests for text, markdown, html formats
  - [ ] Tests for web_reader and raw modes
  - [ ] Auto-fallback tested
  - [ ] Old files deleted

### ⏳ Feature 11: Agent Tools Consolidation (Delegate)
- **Files:** `agents.go`, `agent_list.go`, `agent_status.go`, `agent_run.go`, `delegate.go`
- **Priority:** HIGH
- **Model:** sonnet
- **Depends on:** Feature 6
- **Status:** BLOCKED (waiting for Feature 6)
- **Completion Criteria:**
  - [ ] Tests for all actions (spawn, list, status, stop, run)
  - [ ] Blocking vs non-blocking tested
  - [ ] Old files deleted

### ⏳ Feature 12: Remove Analytics Tools
- **Files:** `track_prompt_usage.go`, `prompt_analytics.go`
- **Priority:** LOW
- **Model:** haiku
- **Status:** PENDING
- **Completion Criteria:**
  - [ ] No active usage found
  - [ ] Files deleted
  - [ ] Build succeeds

### ⏳ Feature 13: Tool Aliasing System
- **File:** `internal/agent/tools/aliases.go` (NEW)
- **Priority:** HIGH
- **Model:** sonnet
- **Depends on:** Features 9, 10, 11
- **Status:** BLOCKED (waiting for 9, 10, 11)
- **Completion Criteria:**
  - [ ] Tests for ResolveToolName
  - [ ] All old tool names aliased
  - [ ] Backward compatibility verified

---

## PHASE 5: TUI Enhancements

### ⏳ Feature 14: Auto-LSP Detection
- **File:** `internal/lsp/autodetect.go` (NEW)
- **Priority:** HIGH
- **Model:** opus (competitive planning)
- **Status:** PENDING
- **Tests Required:**
  - [ ] Go project detection (go.mod → gopls)
  - [ ] Rust detection (Cargo.toml → rust-analyzer)
  - [ ] Node detection (package.json → typescript-language-server)
  - [ ] Python detection (pyproject.toml/requirements.txt → pyright)
  - [ ] Installation command generation
  - [ ] LSP executable existence check
  - [ ] Auto-enable logic

### ⏳ Feature 15: TUI Settings Panel
- **File:** `internal/tui/components/dialogs/settings/settings.go` (NEW)
- **Priority:** HIGH
- **Model:** sonnet
- **Depends on:** Feature 14
- **Status:** BLOCKED (waiting for Feature 14)
- **Tests Required:**
  - [ ] Settings initialization with defaults
  - [ ] Toggle state changes
  - [ ] Persistence across sessions
  - [ ] Immediate application
  - [ ] Keyboard shortcut trigger

### ⏳ Feature 16: Unified Delegate with Resource Pool
- **Files:** Multiple in `internal/agent/delegation/`
- **Priority:** HIGH
- **Model:** opus (competitive planning)
- **Depends on:** Feature 11
- **Status:** BLOCKED (waiting for Feature 11)
- **Tests Required:**
  - [ ] Resource calculation (CPU, memory per agent)
  - [ ] Dynamic pool sizing based on available resources
  - [ ] Queue with timeout (30min)
  - [ ] Agent spawning when resources available
  - [ ] Queue behavior when resources exhausted

### ⏳ Feature 17: Completion Banner Notifications
- **File:** `internal/tui/components/banner/banner.go` (NEW)
- **Priority:** MEDIUM
- **Model:** haiku
- **Depends on:** Feature 16
- **Status:** BLOCKED (waiting for Feature 16)
- **Tests Required:**
  - [ ] Banner creation with success/error state
  - [ ] Auto-dismiss after 10s
  - [ ] Styling (green for success, red for error)
  - [ ] Message formatting

### ⏳ Feature 18: Prompt Repository Import CLI
- **File:** `internal/cmd/import_prompts.go` (enhance)
- **Priority:** LOW
- **Model:** haiku
- **Status:** PENDING
- **Tests Required:**
  - [ ] Default repository import
  - [ ] Custom repository import (-r flag)
  - [ ] Update/sync (-u flag)
  - [ ] Conflict handling
  - [ ] Progress reporting

---

## Orchestration Log

### Initialization Phase
- [ ] Check current coverage baseline
- [ ] Verify test failures (1 in internal/prompts)
- [ ] Initialize swarm workers

### Phase 3 Execution
- [ ] Start Features 1, 2, 3, 5 in parallel
- [ ] Monitor progress every 30 minutes
- [ ] After Feature 3 completes, start Feature 4
- [ ] Start Features 6, 7, 8 in parallel
- [ ] Phase 3 validation: Coverage ≥50%

### Phase 4 Execution
- [ ] Start Features 9, 10, 12 in parallel
- [ ] After Features 9, 10, 11 complete, start Feature 13
- [ ] Phase 4 validation: 9 files deleted, 1 new file

### Phase 5 Execution
- [ ] Start Features 14, 18 in parallel
- [ ] After Feature 14 completes, start Feature 15
- [ ] After Feature 11 completes, start Feature 16
- [ ] After Feature 16 completes, start Feature 17
- [ ] Phase 5 validation: All TUI features working

### Final Validation
- [ ] Run: `go test ./... -race`
- [ ] Run: `go build ./...`
- [ ] Run: `go vet ./...`
- [ ] Final coverage check: ≥50%
