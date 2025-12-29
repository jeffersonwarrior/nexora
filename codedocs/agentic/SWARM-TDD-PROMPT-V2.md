# Nexora v0.29.1-RC1 Completion: Phases 0-5 with TDD Swarm (Version 2)

## CRITICAL: Lessons Learned from Previous Run

**Previous swarm FAILED validation:**
- ❌ Coverage dropped from 36% to 10.8% (target was 50%)
- ❌ 9 tests failing (workers marked features "complete" without running tests)
- ❌ Phase 4 half-done (duplicate files exist: bash.go + bash_monitored.go)
- ❌ Workers stuck/timed out on Features 7 & 8

**New Requirements:**
1. **MANDATORY BLOCKING VALIDATION** - Workers CANNOT mark complete until ALL validations pass
2. **ORCHESTRATOR RE-VALIDATION** - Orchestrator must independently verify worker claims
3. **NO PARTIAL COMPLETION** - Fix issues or report blocked, never "done with failures"
4. **TIMEOUT INVESTIGATION** - Tests taking >10s must be debugged, not skipped

---

## Objective
Complete Phases 0 (Fix Failing Tests), 3 (Test Coverage 36%→50%), 4 (Tool Consolidation), and 5 (TUI Enhancements + Bug Fixes) using test-driven development with parallel swarm workers.

---

## Task Description for Orchestrator

Initialize swarm orchestration for completing Nexora v0.29.1-RC1 Phases 0-5:

**Phase 0 - Fix Failing Tests (NEW - BLOCKING)**
Fix 9 failing tests from previous swarm run before proceeding with any new work. All tests must pass before Phase 3 begins.

**Phase 3 - Test Coverage (36% → 50%)**
Increase overall test coverage from 36.0% to 50.0% by adding comprehensive tests for critical packages. Focus on TUI components (chat, page/chat), agent tools, and core agent functionality. Each feature must write tests FIRST, then verify coverage increases.

**Phase 4 - Tool Consolidation**
Consolidate 27 tools to 19 by merging bash tools, fetch tools, agent delegation tools, and removing analytics tools. Implement transparent aliasing for backward compatibility. Each consolidation must have tests validating both new unified tool and alias compatibility.

**Phase 5 - TUI Enhancements & Critical Bug Fixes**
Implement auto-LSP detection/installation, settings panel UI, unified delegate command with resource-based pooling, and prompt repository import CLI. **PLUS fix 4 critical bugs:** delegate banner breaking TUI, delegate reliability issues, "/" key passthrough, and MCP system overhaul.

---

## MANDATORY VALIDATION PROTOCOL

### Every Feature Must Complete These Steps (NO EXCEPTIONS):

#### Step 1: Write Tests First (TDD)
```bash
# Create test file with:
# 1. Broad integration tests (happy path)
# 2. Narrow unit tests (edge cases)
# 3. Error handling tests

# Tests should FAIL initially (no implementation yet)
go test ./[package]/... -v
# Expected: FAIL (this is correct - tests written first)
```

#### Step 2: Implement Feature
```bash
# Write code to make tests pass
# Iterate until all tests pass
```

#### Step 3: BLOCKING VALIDATION (Required Before Marking Complete)

**3a. Run Tests with Race Detector:**
```bash
go test ./[package]/... -v -race -timeout 5m
```
**PASS CRITERIA:**
- ✅ ALL tests PASS (no FAIL allowed)
- ✅ No race conditions detected
- ✅ No timeouts >10s (if found, investigate deadlock/performance issue)
- ✅ No panics or crashes

**3b. Verify Coverage Target Met:**
```bash
go test -coverprofile=coverage.out ./[package]/...
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
echo "Package Coverage: $COVERAGE% (target: [X]%)"
```
**PASS CRITERIA:**
- ✅ Coverage ≥ feature target (30%, 40%, 50%, or 55% depending on package)
- ✅ Report exact percentage in completion message

**3c. Build Validation:**
```bash
go build ./...
```
**PASS CRITERIA:**
- ✅ Build succeeds with no errors

**3d. Vet Validation:**
```bash
go vet ./[package]/...
```
**PASS CRITERIA:**
- ✅ No warnings or errors

**3e. Full Test Suite Check:**
```bash
go test ./... -race -timeout 10m
```
**PASS CRITERIA:**
- ✅ ALL tests across entire codebase still pass
- ✅ Your changes didn't break other packages

#### Step 4: Report Completion with Evidence

Worker MUST include in completion message:
```
✅ VALIDATION COMPLETE - Feature [X]

Test Results:
- Package tests: [N] passed, 0 failed
- Race detector: CLEAN
- Coverage achieved: [X.X]% (target: [X]%)
- Build: SUCCESS
- Vet: CLEAN
- Full suite: [N] packages passed

Files Changed:
- Added: [list files]
- Modified: [list files]
- Deleted: [list files]

Evidence:
[paste relevant test output showing PASS]
```

**If ANY validation fails:**
- ❌ DO NOT mark feature complete
- 🔧 Fix the issue
- 🔁 Re-run all validations
- 📝 Report the blocker and continue working

---

## ORCHESTRATOR OVERSIGHT PROTOCOL

### When Worker Reports Feature Complete:

**Orchestrator MUST independently verify:**

```bash
# 1. Re-run package tests
go test ./[reported-package]/... -v -race -timeout 5m

# 2. Verify coverage claim
go test -coverprofile=coverage.out ./[reported-package]/...
go tool cover -func=coverage.out | grep total
# Compare to worker's reported coverage

# 3. Check file changes match expectations
git status
git diff

# 4. Run full test suite
go test ./... -race -timeout 10m

# 5. Verify build still works
go build ./...
```

**Orchestrator Acceptance Criteria:**
- ✅ Independent test run confirms worker's PASS claim
- ✅ Coverage matches worker's report (±1%)
- ✅ File changes are minimal and relevant
- ✅ Full test suite still passes
- ✅ Build succeeds

**If verification fails:**
- 🚫 REJECT completion
- 📩 Send worker back with specific failures
- 🔁 Worker must fix and re-validate

**Only mark complete when:**
- ✅ All independent verifications pass
- ✅ No regressions detected
- ✅ Coverage targets met

---

## Feature Breakdown for Swarm

### PHASE 0: Fix Failing Tests (Features 0.1-0.6) - BLOCKING

**CRITICAL:** These must ALL pass before starting Phase 3. No new features until baseline is clean.

---

#### Feature 0.1: Fix TestConnectionHealth_StartStop
**Package:** `internal/agent`
**Priority:** BLOCKING
**Depends on:** None

**Investigation & Fix:**
1. Run test to understand failure:
   ```bash
   go test -v ./internal/agent/... -run TestConnectionHealth_StartStop
   ```
2. Read test code to understand expectations
3. Debug failure (likely race condition or improper cleanup)
4. Fix implementation
5. **MANDATORY VALIDATION:**
   ```bash
   go test ./internal/agent/... -v -race -run TestConnectionHealth_StartStop -count=10
   # Must pass 10 times (no flakiness)
   ```

**Completion Criteria:**
- [ ] Test passes 10 consecutive times
- [ ] No race conditions
- [ ] Fix is minimal and targeted

---

#### Feature 0.2: Fix TestDependencyValidation_ValidDependencies
**Package:** `internal/agent/delegation`
**Priority:** BLOCKING (101-second timeout suggests deadlock!)
**Depends on:** None

**Investigation & Fix:**
1. Run test to see timeout:
   ```bash
   go test -v ./internal/agent/delegation/... -run TestDependencyValidation_ValidDependencies -timeout 2m
   ```
2. Check for mutex deadlock or channel blocking
3. Add timeout/context to prevent infinite waits
4. Fix deadlock
5. **MANDATORY VALIDATION:**
   ```bash
   go test ./internal/agent/delegation/... -v -race -run TestDependencyValidation -timeout 30s
   # Must complete in <5s (previously took 101s)
   ```

**Completion Criteria:**
- [ ] Test passes in <5 seconds
- [ ] No deadlock detected
- [ ] Proper timeout handling added

---

#### Feature 0.3: Fix 4 Shell Tests (Database Timeouts)
**Package:** `internal/agent/shell`
**Priority:** BLOCKING
**Depends on:** None

**Failing Tests:**
- TestShellMonitor_ReviewShellOutput
- TestStoreShellSafetyReview (7.65s)
- TestStoreShellActivity (7.66s)
- TestStoreShellReview (7.72s)

**Investigation & Fix:**
1. Run tests:
   ```bash
   go test -v ./internal/agent/shell/... -run TestShell
   ```
2. All 4 tests taking ~7.6s suggests database timeout
3. Check database initialization in tests
4. Fix DB migration/connection issues
5. **MANDATORY VALIDATION:**
   ```bash
   go test ./internal/agent/shell/... -v -race
   # All 4 tests must pass in <2s total
   ```

**Completion Criteria:**
- [ ] All 4 tests pass
- [ ] Complete in <2 seconds total
- [ ] Database properly initialized in tests

---

#### Feature 0.4: Fix TestClient (LSP)
**Package:** `internal/lsp`
**Priority:** BLOCKING
**Depends on:** None

**Investigation & Fix:**
1. Run test:
   ```bash
   go test -v ./internal/lsp/... -run TestClient
   ```
2. Understand LSP client initialization failure
3. Fix implementation or test expectations
4. **MANDATORY VALIDATION:**
   ```bash
   go test ./internal/lsp/... -v -race
   ```

**Completion Criteria:**
- [ ] Test passes
- [ ] LSP client properly initialized

---

#### Feature 0.5: Fix TestMetadataExtractor_ExtractSubcategory (Flaky)
**Package:** `internal/prompts`
**Priority:** BLOCKING
**Depends on:** None

**Known Issue:**
- Test fails intermittently
- `ExtractSubcategory() = "Database", want "Backend"`
- Suggests non-deterministic behavior (map iteration order?)

**Investigation & Fix:**
1. Run test 20 times to reproduce flakiness:
   ```bash
   go test -v ./internal/prompts/... -run TestMetadataExtractor_ExtractSubcategory -count=20
   ```
2. Identify source of non-determinism
3. Fix implementation (likely need deterministic keyword matching)
4. **MANDATORY VALIDATION:**
   ```bash
   go test ./internal/prompts/... -v -race -run TestMetadataExtractor -count=50
   # Must pass 50 consecutive times (no flakiness)
   ```

**Completion Criteria:**
- [ ] Test passes 50 consecutive times
- [ ] No non-deterministic behavior
- [ ] Subcategory extraction is deterministic

---

#### Feature 0.6: Fix TestQA_PanicRecovery_E2E_Standalone
**Package:** `qa`
**Priority:** BLOCKING
**Depends on:** None

**Investigation & Fix:**
1. Run test:
   ```bash
   go test -v ./qa/... -run TestQA_PanicRecovery_E2E_Standalone
   ```
2. Debug panic recovery mechanism
3. Fix implementation
4. **MANDATORY VALIDATION:**
   ```bash
   go test ./qa/... -v -race -run TestQA_PanicRecovery
   ```

**Completion Criteria:**
- [ ] Test passes
- [ ] Panic recovery works correctly

---

### PHASE 3: Test Coverage (Features 3.1-3.8)

**PREREQUISITE:** Phase 0 must be 100% complete (all 6 features passing)

---

#### Feature 3.1: TUI Chat Component Tests (10.8% → 30%)
**Package:** `internal/tui/components/chat/messages`
**Priority:** CRITICAL (lowest coverage)
**Depends on:** Phase 0 complete

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
4. **MANDATORY VALIDATION (see protocol above)**

**Completion Criteria:**
- [ ] All tests pass with race detector
- [ ] Coverage ≥ 30% (verified by orchestrator)
- [ ] Thinking animation fix validated by tests
- [ ] No race conditions detected
- [ ] Full test suite still passes

---

#### Feature 3.2: TUI Chat Page Tests (8.1% → 30%)
**Package:** `internal/tui/page/chat`
**Priority:** CRITICAL
**Depends on:** Phase 0 complete

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
4. **MANDATORY VALIDATION**

**Completion Criteria:**
- [ ] All tests pass
- [ ] Coverage ≥ 30%
- [ ] "/" command fix validated by tests
- [ ] EditorValue() method tested

---

#### Feature 3.3: Agent Tools Tests (33.8% → 50%)
**Package:** `internal/agent/tools`
**Priority:** CRITICAL (highest gap)
**Depends on:** Phase 0 complete

**TDD Workflow:**
1. Check current coverage baseline
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
5. **MANDATORY VALIDATION**

**Completion Criteria:**
- [ ] All tests pass
- [ ] Coverage ≥ 50%
- [ ] All major tools have tests
- [ ] Error handling tested

---

#### Feature 3.4: Core Agent Tests (27.6% → 50%)
**Package:** `internal/agent`
**Priority:** HIGH
**Depends on:** Phase 0 complete, Feature 3.3 (tools tests provide foundation)

**TDD Workflow:**
1. Check current coverage baseline
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
4. **MANDATORY VALIDATION**

**Completion Criteria:**
- [ ] All tests pass
- [ ] Coverage ≥ 50%
- [ ] ProjectID threading validated by tests
- [ ] Observations capture tested

---

#### Feature 3.5: CLI Commands Tests (26.8% → 40%)
**Package:** `internal/cmd`
**Priority:** HIGH
**Depends on:** Phase 0 complete

**TDD Workflow:**
1. Check current coverage
2. Create tests for undertested commands:
   - Run command logic
   - Import/export functionality
   - Indexing operations
   - Error handling
3. **MANDATORY VALIDATION**

**Completion Criteria:**
- [ ] All tests pass
- [ ] Coverage ≥ 40%
- [ ] All CLI commands tested

---

#### Feature 3.6: Agent Delegation Tests (56.1% → 55%)
**Package:** `internal/agent/delegation`
**Priority:** LOW (already exceeds target!)
**Depends on:** Phase 0 complete

**TDD Workflow:**
1. Verify current coverage (already at 56.1%)
2. Add any missing edge case tests
3. **MANDATORY VALIDATION**

**Completion Criteria:**
- [ ] All tests pass
- [ ] Coverage maintained ≥ 55%

---

#### Feature 3.7: Agent Memory Tests (50.7% → 55%)
**Package:** `internal/agent/memory`
**Priority:** MEDIUM (close to target)
**Depends on:** Phase 0 complete

**TDD Workflow:**
1. Check current coverage (50.7%)
2. Add tests for gaps (+4.3% needed)
3. **MANDATORY VALIDATION**

**Completion Criteria:**
- [ ] All tests pass
- [ ] Coverage ≥ 55%

---

#### Feature 3.8: Database Tests (32.8% → 55%)
**Package:** `internal/db`
**Priority:** HIGH
**Depends on:** Phase 0 complete

**TDD Workflow:**
1. Add tests for undertested queries
2. Test migration rollback
3. **MANDATORY VALIDATION**

**Completion Criteria:**
- [ ] All tests pass
- [ ] Coverage ≥ 55%

---

### PHASE 4: Tool Consolidation (Features 4.1-4.5)

**PREREQUISITE:** Phase 3 must be complete (≥50% overall coverage)

---

#### Feature 4.1: Bash Tool Consolidation
**Files:** `internal/agent/tools/bash.go`, `bash_monitored.go`
**Priority:** HIGH
**Depends on:** Feature 3.3 (agent tools tests)

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
   - Backup current `bash.go` to `bash.go.OLD`
   - Copy `bash_monitored.go` → `bash.go`
   - Update `BashParams` struct with optional fields
   - Implement mode detection logic
4. Delete `bash.go.OLD` and `bash_monitored.go`
5. **MANDATORY VALIDATION**

**Completion Criteria:**
- [ ] All tests pass (both modes)
- [ ] Only one bash.go exists
- [ ] Mode detection works correctly
- [ ] No functionality regression

---

#### Feature 4.2: Fetch Tool Consolidation
**Files:** `internal/agent/tools/fetch.go`, `web_fetch.go`, `internal/agent/agentic_fetch_tool.go`
**Priority:** HIGH
**Depends on:** Feature 3.3

**TDD Workflow:**
1. Create `fetch_consolidated_test.go`:
   - Test text, markdown, html formats
   - Test web_reader and raw modes
   - Test auto-fallback (web_reader → raw)
   - Test timeout handling
2. Implement consolidation:
   - Backup current `fetch.go` to `fetch.go.OLD`
   - Copy `web_fetch.go` → `fetch.go`
   - Update `FetchParams` with format/mode options
3. Delete old files (`fetch.go.OLD`, `agentic_fetch_tool.go`)
4. **MANDATORY VALIDATION**

**Completion Criteria:**
- [ ] All tests pass (all formats/modes)
- [ ] Only one fetch.go exists
- [ ] Auto-fallback tested and working

---

#### Feature 4.3: Agent Tools Consolidation (Delegate)
**Files:** `internal/agent/tools/agents.go`, `agent_list.go`, `agent_status.go`, `agent_run.go`, `delegate.go`
**Priority:** HIGH
**Depends on:** Feature 3.6 (delegation tests)

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
4. **MANDATORY VALIDATION**

**Completion Criteria:**
- [ ] All tests pass (all actions)
- [ ] Old files deleted
- [ ] All actions work correctly

---

#### Feature 4.4: Remove Analytics Tools
**Files:** `internal/agent/tools/track_prompt_usage.go`, `prompt_analytics.go`
**Priority:** LOW
**Depends on:** None

**Investigation & Cleanup:**
1. Verify no active usage:
   ```bash
   grep -r "track_prompt_usage\|prompt_analytics" internal/
   ```
2. Delete files
3. **MANDATORY VALIDATION** (build only)

**Completion Criteria:**
- [ ] Files deleted
- [ ] Build succeeds
- [ ] No references found

---

#### Feature 4.5: Tool Aliasing System
**File:** `internal/agent/tools/aliases.go` (already exists - enhance/test)
**Priority:** HIGH
**Depends on:** Features 4.1, 4.2, 4.3 (all consolidations complete)

**TDD Workflow:**
1. Create `aliases_test.go`:
   - Test ResolveToolName for all aliases
   - Test old names map to new names
   - Test removed tools return empty string
   - Test new names pass through unchanged
2. Review/enhance existing `aliases.go`
3. Integrate into tool dispatch logic
4. **MANDATORY VALIDATION**

**Completion Criteria:**
- [ ] All tests pass
- [ ] All old tool names aliased
- [ ] Logging confirms alias usage
- [ ] Backward compatibility verified

---

### PHASE 5: TUI Enhancements & Critical Bug Fixes (Features 5.1-5.8)

**PREREQUISITE:** Phase 4 must be complete (tool consolidation done)

---

#### Feature 5.1: Auto-LSP Detection and Installation
**File:** `internal/lsp/autodetect.go` (already exists - test/enhance)
**Priority:** HIGH
**Depends on:** Feature 0.4 (LSP tests passing)
**Complexity:** HIGH (consider competitive planning)

**TDD Workflow:**
1. Create `autodetect_test.go`:
   - Test Go project detection (go.mod → gopls)
   - Test Rust detection (Cargo.toml → rust-analyzer)
   - Test Node detection (package.json → typescript-language-server)
   - Test Python detection (pyproject.toml/requirements.txt → pyright)
   - Test installation command generation
   - Test LSP executable existence check
   - Test auto-enable logic
2. Review/enhance existing autodetect.go
3. Integrate with TUI initialization
4. **MANDATORY VALIDATION**

**Completion Criteria:**
- [ ] All tests pass
- [ ] All 4 languages detected correctly
- [ ] Installation commands correct
- [ ] Auto-enable works

---

#### Feature 5.2: TUI Settings Panel
**File:** `internal/tui/components/dialogs/settings/settings.go` (already exists - test/enhance)
**Priority:** HIGH
**Depends on:** Feature 5.1 (LSP settings depend on LSP system)
**Complexity:** MEDIUM

**TDD Workflow:**
1. Create `settings_test.go`:
   - Test settings initialization with defaults
   - Test toggle state changes
   - Test persistence across sessions
   - Test immediate application
   - Test keyboard shortcut trigger
2. Review/enhance existing settings.go
3. Add keyboard shortcut (ctrl+,)
4. **MANDATORY VALIDATION**
5. Manual TUI testing required

**Completion Criteria:**
- [ ] All tests pass
- [ ] All 5 settings implemented
- [ ] Persistence works
- [ ] Keyboard shortcut works

---

#### Feature 5.3: Fix Delegate Banner Breaking TUI (CRITICAL BUG)
**Files:** `internal/tui/components/banner/banner.go`, delegate integration
**Priority:** CRITICAL
**Depends on:** None

**Bug Description:**
- Running delegate agent in Nexora shows giant green banner across top of TUI
- Banner breaks TUI layout
- Should be inline chat message instead

**TDD Workflow:**
1. Create test reproducing banner issue:
   - Test delegate completion triggers inline message (not banner)
   - Test banner component is NOT shown for delegate events
   - Test chat receives inline completion message
2. Fix banner integration:
   - Remove banner trigger for delegate completion
   - Add inline chat message for delegate completion
   - Ensure banner only shows for appropriate events
3. **MANDATORY VALIDATION**
4. Manual TUI testing: run delegate, verify no banner

**Completion Criteria:**
- [ ] All tests pass
- [ ] Delegate completion shows inline message
- [ ] No banner displayed for delegate
- [ ] TUI layout not broken
- [ ] Manual testing confirms fix

---

#### Feature 5.4: Fix Delegate Reliability Issues (CRITICAL BUG)
**Files:** `internal/agent/tools/delegate.go`, delegation system
**Priority:** CRITICAL
**Depends on:** Feature 4.3 (delegate consolidation)

**Bug Description:**
- Delegate agent fails frequently due to variety of errors
- Needs real testing - should be simpler to run as command
- Current interface too complex/fragile

**TDD Workflow:**
1. Create comprehensive delegate tests:
   - Test delegate with simple task (e.g., "list files")
   - Test error handling for common failures
   - Test timeout handling
   - Test resource cleanup on failure
   - Test simplified command interface
2. Simplify delegate interface:
   - Add simple mode: just task description (auto-detect agent type)
   - Better error messages
   - Graceful degradation
   - Automatic retry for transient failures
3. **MANDATORY VALIDATION**
4. Real-world testing: run 10 delegate commands, all should succeed

**Completion Criteria:**
- [ ] All tests pass
- [ ] Simplified interface implemented
- [ ] 10/10 real delegate commands succeed
- [ ] Error messages are clear and actionable
- [ ] Automatic retry works for transient failures

---

#### Feature 5.5: Fix "/" Key Passthrough (CRITICAL BUG)
**Files:** `internal/tui/tui.go`, `internal/tui/page/chat/chat.go`
**Priority:** CRITICAL
**Depends on:** Feature 3.2 (chat page tests)

**Bug Description:**
- Current: "/" only typeable when editor is completely empty
- Expected: "/" should be typeable after first character
- Example: User types "Can you help" then "/" → should work
- Example: "/" at start → should trigger command menu (current behavior OK)

**TDD Workflow:**
1. Create tests for "/" key behavior:
   - Test empty editor + "/" → triggers command menu ✓ (existing)
   - Test "text" + "/" → "/" passes through to editor ✓ (NEW)
   - Test "text /" + more text → works correctly ✓ (NEW)
   - Test "/" at position 0 vs position >0
2. Fix logic in `internal/tui/tui.go:551-554`:
   - Check cursor position, not just editor emptiness
   - Only trigger menu if "/" is first character AND editor empty
3. **MANDATORY VALIDATION**
4. Manual TUI testing: type "help me /" and verify "/" appears

**Completion Criteria:**
- [ ] All tests pass
- [ ] "/" works at position >0
- [ ] "/" at position 0 still triggers menu
- [ ] Manual testing confirms fix

---

#### Feature 5.6: MCP System Overhaul (CRITICAL BUG)
**Files:** `internal/mcp/*`, MCP integration layer
**Priority:** CRITICAL
**Depends on:** None
**Complexity:** HIGH (consider competitive planning)

**Bug Description:**
- MCP system needs overhaul (user reported general issues)
- Likely: connection reliability, error handling, configuration

**Investigation & Fix:**
1. Audit current MCP implementation:
   - Review connection handling
   - Review error handling
   - Review configuration system
   - Identify pain points
2. Create comprehensive MCP tests:
   - Test MCP server connection
   - Test tool discovery
   - Test tool execution
   - Test error recovery
   - Test configuration loading
3. Implement improvements:
   - Better error messages
   - Connection retry logic
   - Configuration validation
   - Graceful degradation
4. **MANDATORY VALIDATION**
5. Real-world testing with actual MCP servers

**Completion Criteria:**
- [ ] All tests pass
- [ ] MCP servers connect reliably
- [ ] Error messages are clear
- [ ] Configuration is simple
- [ ] Manual testing with 3 different MCP servers succeeds

---

#### Feature 5.7: Unified Delegate with Resource Pool
**Files:** Multiple in `internal/agent/delegation/`
**Priority:** HIGH
**Depends on:** Feature 5.4 (delegate reliability fixed)
**Complexity:** HIGH (consider competitive planning)

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
5. **MANDATORY VALIDATION**

**Completion Criteria:**
- [ ] All tests pass
- [ ] Resource limits enforced
- [ ] Queue timeout works
- [ ] Dynamic spawning works

---

#### Feature 5.8: Prompt Repository Import CLI
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
3. **MANDATORY VALIDATION**
4. Real-world test: import test repository

**Completion Criteria:**
- [ ] All tests pass
- [ ] Default repo import works
- [ ] Custom repo (-r) works
- [ ] Update (-u) works

---

## Final Validation Checklist

After all features complete, orchestrator must verify:

### Phase 0 Validation (Blocking)
```bash
# All 9 failing tests must pass
go test ./internal/agent/... -v -race -run TestConnectionHealth_StartStop -count=10
go test ./internal/agent/delegation/... -v -race -run TestDependencyValidation -timeout 30s
go test ./internal/agent/shell/... -v -race
go test ./internal/lsp/... -v -race -run TestClient
go test ./internal/prompts/... -v -race -run TestMetadataExtractor -count=50
go test ./qa/... -v -race -run TestQA_PanicRecovery

# All tests must pass before proceeding
echo "Phase 0 Status: ALL TESTS PASSING ✓"
```

### Phase 3 Validation
```bash
# Overall coverage check
go test -coverprofile=coverage.out ./...
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
echo "Overall Coverage: $COVERAGE% (target: ≥50%)"

if (( $(echo "$COVERAGE >= 50" | bc -l) )); then
  echo "✓ Phase 3 PASSED"
else
  echo "✗ Phase 3 FAILED: Coverage too low"
  exit 1
fi

# Per-package coverage verification
echo "Verifying package coverage targets..."

go test -coverprofile=coverage.out ./internal/tui/components/chat/messages/...
COV=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
echo "TUI Chat Messages: $COV% (target: ≥30%)"

go test -coverprofile=coverage.out ./internal/tui/page/chat/...
COV=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
echo "TUI Chat Page: $COV% (target: ≥30%)"

go test -coverprofile=coverage.out ./internal/agent/tools/...
COV=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
echo "Agent Tools: $COV% (target: ≥50%)"

go test -coverprofile=coverage.out ./internal/agent/...
COV=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
echo "Core Agent: $COV% (target: ≥50%)"
```

### Phase 4 Validation
```bash
# Verify old tool files deleted
echo "Checking for duplicate tool files..."
! test -f internal/agent/tools/bash_monitored.go || echo "ERROR: bash_monitored.go still exists"
! test -f internal/agent/tools/fetch.go.OLD || echo "ERROR: fetch.go.OLD still exists"
! test -f internal/agent/tools/web_fetch.go || echo "ERROR: web_fetch.go still exists"
! test -f internal/agent/tools/agents.go || echo "ERROR: agents.go still exists"
! test -f internal/agent/tools/agent_list.go || echo "ERROR: agent_list.go still exists"
! test -f internal/agent/tools/agent_status.go || echo "ERROR: agent_status.go still exists"
! test -f internal/agent/tools/agent_run.go || echo "ERROR: agent_run.go still exists"
! test -f internal/agent/tools/track_prompt_usage.go || echo "ERROR: track_prompt_usage.go still exists"
! test -f internal/agent/tools/prompt_analytics.go || echo "ERROR: prompt_analytics.go still exists"

# Verify new consolidated files exist
test -f internal/agent/tools/bash.go && echo "✓ bash.go exists"
test -f internal/agent/tools/fetch.go && echo "✓ fetch.go exists"
test -f internal/agent/tools/delegate.go && echo "✓ delegate.go exists"
test -f internal/agent/tools/aliases.go && echo "✓ aliases.go exists"

# Verify aliasing works
go test ./internal/agent/tools/... -v -run TestAlias
echo "✓ Phase 4 PASSED"
```

### Phase 5 Validation
```bash
# Verify new files exist
test -f internal/lsp/autodetect.go && echo "✓ autodetect.go exists"
test -f internal/tui/components/dialogs/settings/settings.go && echo "✓ settings.go exists"
test -f internal/tui/components/banner/banner.go && echo "✓ banner.go exists"

# Run TUI feature tests
go test ./internal/lsp/... -v -race
go test ./internal/tui/components/dialogs/settings/... -v -race
go test ./internal/tui/components/banner/... -v -race

echo "✓ Phase 5 PASSED"
```

### Final Build & Test Validation
```bash
echo "Running final validation suite..."

# All tests must pass with race detector
go test ./... -race -v -timeout 10m
if [ $? -eq 0 ]; then
  echo "✓ All tests PASS"
else
  echo "✗ Tests FAILED"
  exit 1
fi

# Build must succeed
go build ./...
if [ $? -eq 0 ]; then
  echo "✓ Build SUCCESS"
else
  echo "✗ Build FAILED"
  exit 1
fi

# No vet warnings
go vet ./...
if [ $? -eq 0 ]; then
  echo "✓ Vet CLEAN"
else
  echo "✗ Vet WARNINGS"
  exit 1
fi

# Final coverage check
go test -coverprofile=coverage.out ./...
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
echo "Final Coverage: $COVERAGE%"

if (( $(echo "$COVERAGE >= 50" | bc -l) )); then
  echo "✅ FINAL VALIDATION PASSED"
  echo "Coverage: $COVERAGE% ✓"
  echo "Tests: ALL PASS ✓"
  echo "Build: SUCCESS ✓"
  echo "Vet: CLEAN ✓"
else
  echo "❌ FINAL VALIDATION FAILED"
  echo "Coverage: $COVERAGE% (need ≥50%)"
  exit 1
fi
```

### Manual TUI Testing (Required for Phase 5 bugs)
```bash
# Build and run nexora for manual testing
go build -o nexora ./cmd/nexora

# Test 1: Delegate banner fix
./nexora
# In TUI: run a delegate command
# Verify: inline message appears, NO banner

# Test 2: Delegate reliability
# Run 10 different delegate commands
# Verify: all succeed

# Test 3: "/" key passthrough
# In TUI: type "Can you help /"
# Verify: "/" appears in text

# Test 4: MCP system
# Connect to 3 different MCP servers
# Verify: all connect successfully

# All 4 manual tests must pass
echo "✅ MANUAL TUI TESTING PASSED"
```

---

## Orchestration Strategy

**Recommended Approach:**

### 1. Initialize orchestrator:
```bash
Use orchestrator_init with this task description and 22 features (6 + 8 + 5 + 3)
```

### 2. Phase 0 - Fix Failing Tests (BLOCKING - Features 0.1-0.6)
**Must complete BEFORE any Phase 3 work:**
- Start all 6 features in parallel (independent fixes)
- Wait for ALL to complete
- Run full validation: `go test ./... -race`
- Only proceed to Phase 3 when ALL tests pass

### 3. Phase 3 - Test Coverage (Features 3.1-3.8)
**After Phase 0 complete:**
- Start Features 3.1, 3.2, 3.3, 3.5 in parallel (no dependencies)
- After Feature 3.3 completes, start Feature 3.4
- Start Features 3.6, 3.7, 3.8 in parallel
- After all complete, verify ≥50% overall coverage

### 4. Phase 4 - Tool Consolidation (Features 4.1-4.5)
**After Phase 3 complete:**
- Start Features 4.1, 4.2, 4.4 in parallel
- After Features 4.1, 4.2, 4.3 complete, start Feature 4.5
- Verify old files deleted, aliases work

### 5. Phase 5 - TUI Enhancements & Bugs (Features 5.1-5.8)
**After Phase 4 complete:**
- Start Features 5.1, 5.3, 5.4, 5.5, 5.6 in parallel (critical bugs)
- After Feature 5.1 completes, start Feature 5.2
- After Feature 5.4 completes, start Feature 5.7
- Feature 5.8 can run anytime
- Manual TUI testing required for 5.3, 5.4, 5.5, 5.6

### 6. Use competitive planning for:
- Feature 5.1 (Auto-LSP Detection) - complex detection logic
- Feature 5.6 (MCP Overhaul) - complex system redesign
- Feature 5.7 (Resource Pool) - complex resource management

### 7. Verification cadence:
- After Phase 0: Run full test suite (must be 100% pass)
- After Phase 3: Run coverage validation (must be ≥50%)
- After Phase 4: Run aliasing and consolidation tests
- After Phase 5: Run TUI feature tests + manual testing
- Final: Run complete validation checklist

**Worker Model Selection:**
- Use `haiku` for simple features (3.6, 4.4, 5.8)
- Use `sonnet` for most features (all Phase 0, 3.1-3.5, 3.7-3.8, 4.1-4.3, 4.5, 5.2-5.5)
- Use `opus` for competitive planning (5.1, 5.6, 5.7)

---

## Success Metrics

**Phase 0:** All 9 failing tests pass (100% baseline)
**Phase 3:** Coverage increases from 36% → ≥50%
**Phase 4:** 9 files deleted, 1 file enhanced (aliases.go), all tests pass
**Phase 5:** 4 critical bugs fixed + 4 features implemented, all manual validations pass
**Overall:** All tests pass, build succeeds, no vet warnings, ≥50% coverage, TUI bugs fixed

**Estimated completion:** 8-12 hours with 5 parallel workers

---

## Critical Success Factors

1. **BLOCKING validation at every step** - no marking complete without proof
2. **Orchestrator independent verification** - don't trust worker self-reports
3. **Phase 0 must be 100% complete** - no new work until baseline is clean
4. **Coverage must be verified** - run go tool cover, not just claim percentage
5. **Manual TUI testing is mandatory** - Phase 5 bugs require human verification
6. **No partial completion** - fix issues or report blocked, never "done with failures"

**Previous run failed because workers self-reported without validation. This time, validation is MANDATORY and BLOCKING.**
