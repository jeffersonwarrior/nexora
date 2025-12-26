# Nexora v0.29.1-RC1 Completion: Phases 3-5 with TDD Swarm

## Objective
Complete Phases 3 (Test Coverage 36%→50%), 4 (Tool Consolidation), and 5 (TUI Enhancements) using test-driven development with parallel swarm workers.

## Task Description for Orchestrator

Initialize swarm orchestration for completing Nexora v0.29.1-RC1 Phases 3-5:

**Phase 3 - Test Coverage (36% → 50%)**
Increase overall test coverage from 36.0% to 50.0% by adding comprehensive tests for critical packages. Focus on TUI components (chat, page/chat), agent tools, and core agent functionality. Each feature must write tests FIRST, then verify coverage increases.

**Phase 4 - Tool Consolidation**
Consolidate 27 tools to 19 by merging bash tools, fetch tools, agent delegation tools, and removing analytics tools. Implement transparent aliasing for backward compatibility. Each consolidation must have tests validating both new unified tool and alias compatibility.

**Phase 5 - TUI Enhancements**
Implement auto-LSP detection/installation, settings panel UI, unified delegate command with resource-based pooling, and prompt repository import CLI. Each feature must have tests before implementation.

**TDD Validation Workflow:**
1. Worker creates broad integration tests defining expected behavior
2. Worker creates narrow unit tests for edge cases
3. Worker implements feature to pass tests
4. Worker validates: `go test ./[package]/... -v -race`
5. Worker validates coverage increase: `go test -coverprofile=coverage.out ./[package]/...`
6. Orchestrator validates overall coverage before marking complete
7. Final validation: `go test ./... -race && go build ./...`

**Success Criteria:**
- All tests pass: `go test ./... -race`
- Coverage ≥ 50%: `go tool cover -func=coverage.out | grep total`
- Build succeeds: `go build ./...`
- No vet warnings: `go vet ./...`

---

## Feature Breakdown for Swarm

### PHASE 3: Test Coverage (Features 1-8)

#### Feature 1: TUI Chat Component Tests (3.2% → 30%)
**Package:** `internal/tui/components/chat/messages`
**Priority:** CRITICAL (lowest coverage)
**Depends on:** None

**TDD Workflow:**
1. Create test file `messages_test.go` with broad integration tests:
   - Test thinking animation display logic (verify Fix 1 from Phase 1)
   - Test message rendering with markdown
   - Test citation handling
   - Test footer state management
2. Create narrow unit tests for edge cases:
   - Empty thinking content
   - Long markdown content
   - Multiple citations
   - State transitions
3. Implement any missing functionality to pass tests
4. Validate: `go test ./internal/tui/components/chat/messages/... -v -race -coverprofile=coverage.out`
5. Verify coverage: `go tool cover -func=coverage.out | grep total` (target: 30%+)

**Completion Criteria:**
- [ ] All tests pass
- [ ] Coverage ≥ 30%
- [ ] Thinking animation fix validated by tests
- [ ] No race conditions detected

---

#### Feature 2: TUI Chat Page Tests (8.1% → 30%)
**Package:** `internal/tui/page/chat`
**Priority:** CRITICAL
**Depends on:** None

**TDD Workflow:**
1. Create broad integration tests in `chat_page_test.go`:
   - Test "/" command trigger logic (verify Fix 3 from Phase 2)
   - Test editor value retrieval
   - Test chat page state transitions
   - Test keyboard shortcut handling
2. Create narrow unit tests:
   - Empty editor "/" trigger
   - Non-empty editor "/" passthrough
   - Editor value edge cases
3. Implement missing functionality
4. Validate: `go test ./internal/tui/page/chat/... -v -race -coverprofile=coverage.out`
5. Verify coverage ≥ 30%

**Completion Criteria:**
- [ ] All tests pass
- [ ] Coverage ≥ 30%
- [ ] "/" command fix validated by tests
- [ ] EditorValue() method tested

---

#### Feature 3: Agent Tools Tests (Need current % → 50%)
**Package:** `internal/agent/tools`
**Priority:** CRITICAL (highest gap)
**Depends on:** None

**TDD Workflow:**
1. Check current coverage: `go test -coverprofile=coverage.out ./internal/agent/tools/... && go tool cover -func=coverage.out | grep total`
2. Create broad integration tests for undertested tools:
   - Bash tool execution and error handling
   - Fetch operations with timeout
   - Edit operations with validation
   - File operations (read, write, glob, grep)
   - MCP tool integration
3. Create narrow unit tests:
   - Tool parameter validation
   - Error handling edge cases
   - Timeout behavior
   - Resource cleanup
4. Implement missing error handling/validation
5. Validate: `go test ./internal/agent/tools/... -v -race -coverprofile=coverage.out`
6. Verify coverage ≥ 50%

**Completion Criteria:**
- [ ] All tests pass
- [ ] Coverage ≥ 50%
- [ ] All major tools have tests
- [ ] Error handling tested

---

#### Feature 4: Core Agent Tests (Need current % → 50%)
**Package:** `internal/agent`
**Priority:** HIGH
**Depends on:** Feature 3 (tools tests provide foundation)

**TDD Workflow:**
1. Check current coverage
2. Create broad integration tests:
   - Conversation state management
   - Tool orchestration flow
   - ProjectID threading (verify Fix 2 from Phase 1)
   - Message handling
   - LLM integration mocks
3. Create narrow unit tests:
   - State transitions
   - Error recovery
   - SessionAgentOptions initialization
   - ToolExecution with projectID
4. Validate: `go test ./internal/agent/... -v -race -coverprofile=coverage.out`
5. Verify coverage ≥ 50%

**Completion Criteria:**
- [ ] All tests pass
- [ ] Coverage ≥ 50%
- [ ] ProjectID threading validated by tests
- [ ] Observations capture tested

---

#### Feature 5: CLI Commands Tests (Need current % → 40%)
**Package:** `internal/cmd`
**Priority:** HIGH
**Depends on:** None

**TDD Workflow:**
1. Check current coverage
2. Create tests for undertested commands:
   - Run command logic
   - Import/export functionality
   - Indexing operations
   - Error handling
3. Validate: `go test ./internal/cmd/... -v -race -coverprofile=coverage.out`
4. Verify coverage ≥ 40%

**Completion Criteria:**
- [ ] All tests pass
- [ ] Coverage ≥ 40%
- [ ] All CLI commands tested

---

#### Feature 6: Agent Delegation Tests (45.8% → 55%)
**Package:** `internal/agent/delegation`
**Priority:** MEDIUM
**Depends on:** None

**TDD Workflow:**
1. Add tests for gaps:
   - Pool management
   - Agent registry
   - Completion tracking
   - Queue operations
2. Validate and verify coverage ≥ 55%

**Completion Criteria:**
- [ ] All tests pass
- [ ] Coverage ≥ 55%

---

#### Feature 7: Agent Memory Tests (46.9% → 55%)
**Package:** `internal/agent/memory`
**Priority:** MEDIUM
**Depends on:** None

**TDD Workflow:**
1. Add tests for gaps in memory system
2. Validate and verify coverage ≥ 55%

**Completion Criteria:**
- [ ] All tests pass
- [ ] Coverage ≥ 55%

---

#### Feature 8: Database Tests (32.8% → 55%)
**Package:** `internal/db`
**Priority:** MEDIUM
**Depends on:** None

**TDD Workflow:**
1. Add tests for undertested queries
2. Test migration rollback
3. Validate and verify coverage ≥ 55%

**Completion Criteria:**
- [ ] All tests pass
- [ ] Coverage ≥ 55%

---

### PHASE 4: Tool Consolidation (Features 9-13)

#### Feature 9: Bash Tool Consolidation
**Files:** `internal/agent/tools/bash.go`, `bash_monitored.go`
**Priority:** HIGH
**Depends on:** Feature 3 (agent tools tests)

**TDD Workflow:**
1. Create `bash_consolidated_test.go` with tests for:
   - Standard bash execution mode
   - AI-monitored mode (with purpose + completion_criteria)
   - Mode detection logic
   - Parameter validation
   - Both modes produce correct results
2. Create narrow tests:
   - Empty purpose/criteria → standard mode
   - Both provided → monitored mode
   - Timeout handling in both modes
3. Implement consolidation:
   - Rename `bash_monitored.go` → `bash.go` (backup old first)
   - Update `BashParams` struct with optional fields
   - Implement mode detection logic
4. Delete old `bash.go`
5. Validate: `go test ./internal/agent/tools/... -v -race`

**Completion Criteria:**
- [ ] All tests pass (both modes)
- [ ] Old `bash.go` deleted
- [ ] Mode detection works correctly
- [ ] No functionality regression

---

#### Feature 10: Fetch Tool Consolidation
**Files:** `internal/agent/tools/fetch.go`, `web_fetch.go`, `agentic_fetch_tool.go`
**Priority:** HIGH
**Depends on:** Feature 3

**TDD Workflow:**
1. Create `fetch_consolidated_test.go`:
   - Test text, markdown, html formats
   - Test web_reader and raw modes
   - Test auto-fallback (web_reader → raw)
   - Test timeout handling
2. Implement consolidation:
   - Rename `web_fetch.go` → `fetch.go`
   - Update `FetchParams` with format/mode options
3. Delete old files
4. Validate

**Completion Criteria:**
- [ ] All tests pass (all formats/modes)
- [ ] Old files deleted
- [ ] Auto-fallback tested and working

---

#### Feature 11: Agent Tools Consolidation (Delegate)
**Files:** `internal/agent/tools/agents.go`, `agent_list.go`, `agent_status.go`, `agent_run.go`, `delegate.go`
**Priority:** HIGH
**Depends on:** Feature 6 (delegation tests)

**TDD Workflow:**
1. Create `delegate_consolidated_test.go`:
   - Test action=spawn with task description
   - Test action=list
   - Test action=status with session_id
   - Test action=stop with session_id
   - Test action=run with prompt
   - Test blocking vs non-blocking
2. Enhance `delegate.go` with action parameter
3. Delete old agent_*.go files
4. Validate

**Completion Criteria:**
- [ ] All tests pass (all actions)
- [ ] Old files deleted
- [ ] All actions work correctly

---

#### Feature 12: Remove Analytics Tools
**Files:** `internal/agent/tools/track_prompt_usage.go`, `prompt_analytics.go`
**Priority:** LOW
**Depends on:** None

**TDD Workflow:**
1. Verify no active usage: `grep -r "track_prompt_usage\|prompt_analytics" internal/`
2. Delete files
3. Validate: `go build ./...`

**Completion Criteria:**
- [ ] Files deleted
- [ ] Build succeeds
- [ ] No references found

---

#### Feature 13: Tool Aliasing System
**File:** `internal/agent/tools/aliases.go` (NEW)
**Priority:** HIGH
**Depends on:** Features 9, 10, 11 (all consolidations complete)

**TDD Workflow:**
1. Create `aliases_test.go`:
   - Test ResolveToolName for all aliases
   - Test old names map to new names
   - Test removed tools return empty string
   - Test new names pass through unchanged
2. Implement `aliases.go`:
   - Create ToolAliases map
   - Implement ResolveToolName function
   - Add logging for alias usage
3. Integrate into tool dispatch logic
4. Validate: AI can use old names, TUI shows new names

**Completion Criteria:**
- [ ] All tests pass
- [ ] All old tool names aliased
- [ ] Logging confirms alias usage
- [ ] Backward compatibility verified

---

### PHASE 5: TUI Enhancements (Features 14-17)

#### Feature 14: Auto-LSP Detection and Installation
**File:** `internal/lsp/autodetect.go` (NEW)
**Priority:** HIGH
**Depends on:** None
**Complexity:** HIGH (use competitive planning)

**TDD Workflow:**
1. Create `autodetect_test.go`:
   - Test Go project detection (go.mod → gopls)
   - Test Rust detection (Cargo.toml → rust-analyzer)
   - Test Node detection (package.json → typescript-language-server)
   - Test Python detection (pyproject.toml/requirements.txt → pyright)
   - Test installation command generation
   - Test LSP executable existence check
   - Test auto-enable logic
2. Implement autodetection logic
3. Integrate with TUI initialization
4. Validate: Create test projects, verify detection

**Completion Criteria:**
- [ ] All tests pass
- [ ] All 4 languages detected correctly
- [ ] Installation commands correct
- [ ] Auto-enable works

---

#### Feature 15: TUI Settings Panel
**File:** `internal/tui/components/dialogs/settings/settings.go` (NEW)
**Priority:** HIGH
**Depends on:** Feature 14 (LSP settings depend on LSP system)
**Complexity:** MEDIUM

**TDD Workflow:**
1. Create `settings_test.go`:
   - Test settings initialization with defaults
   - Test toggle state changes
   - Test persistence across sessions
   - Test immediate application
   - Test keyboard shortcut trigger
2. Implement settings dialog:
   - Yolo Mode toggle
   - Sidebar toggle
   - Help toggle
   - LSP Auto-detect toggle
   - LSP Auto-install toggle
3. Add keyboard shortcut (ctrl+,)
4. Validate: Manual TUI testing

**Completion Criteria:**
- [ ] All tests pass
- [ ] All 5 settings implemented
- [ ] Persistence works
- [ ] Keyboard shortcut works

---

#### Feature 16: Unified Delegate with Resource Pool
**Files:** Multiple in `internal/agent/delegation/`
**Priority:** HIGH
**Depends on:** Feature 11 (delegate consolidation)
**Complexity:** HIGH (use competitive planning)

**TDD Workflow:**
1. Create tests:
   - Test resource calculation (CPU, memory per agent)
   - Test dynamic pool sizing based on available resources
   - Test queue with timeout (30min)
   - Test agent spawning when resources available
   - Test queue behavior when resources exhausted
2. Implement ResourceConfig and monitoring
3. Implement dynamic agent spawning
4. Add queue with timeout
5. Validate under load

**Completion Criteria:**
- [ ] All tests pass
- [ ] Resource limits enforced
- [ ] Queue timeout works
- [ ] Dynamic spawning works

---

#### Feature 17: Completion Banner Notifications
**File:** `internal/tui/components/banner/banner.go` (NEW)
**Priority:** MEDIUM
**Depends on:** Feature 16 (delegate completion events)
**Complexity:** LOW

**TDD Workflow:**
1. Create `banner_test.go`:
   - Test banner creation with success/error state
   - Test auto-dismiss after 10s
   - Test styling (green for success, red for error)
   - Test message formatting
2. Implement banner component
3. Integrate with delegate completion events
4. Validate: Manual TUI testing

**Completion Criteria:**
- [ ] All tests pass
- [ ] Auto-dismiss works
- [ ] Styling correct
- [ ] Shows on delegate completion

---

#### Feature 18: Prompt Repository Import CLI
**File:** `internal/cmd/import_prompts.go` (enhance existing)
**Priority:** LOW
**Depends on:** None
**Complexity:** LOW

**TDD Workflow:**
1. Create tests:
   - Test default repository import
   - Test custom repository import (-r flag)
   - Test update/sync (-u flag)
   - Test conflict handling
   - Test progress reporting
2. Implement CLI commands
3. Validate: Import test repository

**Completion Criteria:**
- [ ] All tests pass
- [ ] Default repo import works
- [ ] Custom repo (-r) works
- [ ] Update (-u) works

---

## Final Validation Checklist

After all features complete, orchestrator must verify:

### Phase 3 Validation
```bash
# Overall coverage check
go test -coverprofile=coverage.out ./...
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
echo "Coverage: $COVERAGE% (target: ≥50%)"

# Per-package coverage verification
go test -coverprofile=coverage.out ./internal/tui/components/chat/messages/...
go tool cover -func=coverage.out | grep total  # Should be ≥30%

go test -coverprofile=coverage.out ./internal/tui/page/chat/...
go tool cover -func=coverage.out | grep total  # Should be ≥30%

go test -coverprofile=coverage.out ./internal/agent/tools/...
go tool cover -func=coverage.out | grep total  # Should be ≥50%

go test -coverprofile=coverage.out ./internal/agent/...
go tool cover -func=coverage.out | grep total  # Should be ≥50%
```

### Phase 4 Validation
```bash
# Verify old tool files deleted
! test -f internal/agent/tools/bash.go || echo "ERROR: Old bash.go still exists"
! test -f internal/agent/tools/fetch.go || echo "ERROR: Old fetch.go still exists"
! test -f internal/agent/tools/agents.go || echo "ERROR: Old agents.go still exists"

# Verify new consolidated files exist
test -f internal/agent/tools/bash.go && echo "✓ Consolidated bash.go exists"
test -f internal/agent/tools/fetch.go && echo "✓ Consolidated fetch.go exists"
test -f internal/agent/tools/aliases.go && echo "✓ aliases.go exists"

# Verify aliasing works
go test ./internal/agent/tools/... -v -run TestAlias
```

### Phase 5 Validation
```bash
# Verify new files exist
test -f internal/lsp/autodetect.go && echo "✓ autodetect.go exists"
test -f internal/tui/components/dialogs/settings/settings.go && echo "✓ settings.go exists"
test -f internal/tui/components/banner/banner.go && echo "✓ banner.go exists"

# Run TUI feature tests
go test ./internal/lsp/... -v
go test ./internal/tui/components/dialogs/settings/... -v
go test ./internal/tui/components/banner/... -v
```

### Final Build & Test Validation
```bash
# All tests must pass with race detector
go test ./... -race -v

# Build must succeed
go build ./...

# No vet warnings
go vet ./...

# Final coverage check
go test -coverprofile=coverage.out ./...
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
if (( $(echo "$COVERAGE >= 50" | bc -l) )); then
  echo "✓ Coverage PASSED: $COVERAGE%"
else
  echo "✗ Coverage FAILED: $COVERAGE% (need ≥50%)"
  exit 1
fi
```

---

## Orchestration Strategy

**Recommended Approach:**

1. **Initialize orchestrator:**
   ```
   Use orchestrator_init with this task description and 18 features
   ```

2. **Phase 3 parallelization (Features 1-8):**
   - Start Features 1, 2, 3, 5 in parallel (no dependencies)
   - After Feature 3 completes, start Feature 4
   - Start Features 6, 7, 8 in parallel

3. **Phase 4 parallelization (Features 9-13):**
   - Start Features 9, 10, 12 in parallel (Feature 12 is trivial)
   - After Features 9, 10, 11 complete, start Feature 13

4. **Phase 5 parallelization (Features 14-18):**
   - Start Features 14, 18 in parallel
   - After Feature 14 completes, start Feature 15
   - After Feature 11 completes, start Feature 16
   - After Feature 16 completes, start Feature 17

5. **Use competitive planning for:**
   - Feature 14 (Auto-LSP Detection) - complex detection logic
   - Feature 16 (Resource Pool) - complex resource management

6. **Verification cadence:**
   - After Phase 3 complete: Run full coverage validation
   - After Phase 4 complete: Run aliasing and consolidation tests
   - After Phase 5 complete: Run TUI feature tests
   - Final: Run complete validation checklist

**Worker Model Selection:**
- Use `haiku` for simple features (12, 17, 18)
- Use `sonnet` for most features (1-11, 13, 14, 15)
- Use `opus` for competitive planning (14, 16)

---

## Success Metrics

**Phase 3:** Coverage increases from 36% → ≥50%
**Phase 4:** 9 files deleted, 1 new file (aliases.go), all tests pass
**Phase 5:** 4 new features implemented with tests, all manual validations pass
**Overall:** All tests pass, build succeeds, no vet warnings, ≥50% coverage

**Estimated completion:** 6-8 hours with 5 parallel workers
