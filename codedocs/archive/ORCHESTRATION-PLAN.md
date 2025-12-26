# Nexora RC1 Autonomous Orchestration Plan

**Generated:** 2025-12-26T08:10:00Z
**Mode:** Full Auto (User sleeping, laptop on)
**Objective:** Complete Phases 0-5 for v0.29.1-RC1 release

---

## Current Status

### Completed

1. ✅ **SWARM-TDD-PROMPT-V2.md created** - Enhanced prompt with:
   - Mandatory blocking validation protocol
   - Orchestrator oversight requirements
   - 4 new critical bugs integrated
   - 22 total features (6 bug fixes + 8 coverage + 5 consolidation + 3 enhancements)

2. ✅ **Phase 4 baseline rollback** - Removed incomplete consolidation attempts:
   - Archived: `bash_consolidated.go.backup`, `fetch_consolidated.go.backup`, `delegate_enhanced.go`
   - Created: `PHASE4-BASELINE.md` documenting current state
   - Baseline: 47 tool files, clean state for consolidation

3. ✅ **PHASE4-BASELINE.md created** - Documentation of:
   - Current tool file count (47 implementation, 31 tests)
   - Files to consolidate (bash, fetch, agent tools)
   - Consolidation plan
   - Success criteria

4. 🔄 **Phase 0 in progress** - 6 parallel agents fixing failing tests:
   - Agent a9ee121: ✅ COMPLETE - Fixed TestMetadataExtractor flakiness (map → slice)
   - Agent aea6b50: 🔄 Running - Fixing TestConnectionHealth_StartStop
   - Agent a5a4035: 🔄 Running - Fixing TestDependencyValidation deadlock
   - Agent aee3ec7: 🔄 Running - Fixing 4 shell tests (DB timeouts)
   - Agent aee92cd: 🔄 Running - Fixing TestClient LSP
   - Agent a2bdddd: 🔄 Running - Fixing TestQA_PanicRecovery

---

## Execution Plan

### Phase 0: Fix Failing Tests (IN PROGRESS)

**Blocking:** Must be 100% complete before Phase 3

**Agent Status:**
- ✅ 1/6 complete (TestMetadataExtractor)
- 🔄 5/6 running

**Next Steps:**
1. Wait for all 6 agents to complete
2. Verify all fixes independently
3. Run full test suite: `go test ./... -race`
4. Verify build: `go build ./...`
5. Only proceed to Phase 3 when ALL tests pass

---

### Phase 3: Test Coverage (36% → 50%)

**Prerequisite:** Phase 0 must be 100% complete

**Features (8):**
1. TUI Chat Component Tests (10.8% → 30%) - CRITICAL
2. TUI Chat Page Tests (8.1% → 30%) - CRITICAL
3. Agent Tools Tests (33.8% → 50%) - CRITICAL
4. Core Agent Tests (27.6% → 50%) - HIGH
5. CLI Commands Tests (26.8% → 40%) - HIGH
6. Agent Delegation Tests (56.1% → 55%) - LOW (already exceeds!)
7. Agent Memory Tests (50.7% → 55%) - MEDIUM
8. Database Tests (32.8% → 55%) - HIGH

**Parallelization Strategy:**
- Batch 1 (parallel): Features 1, 2, 3, 5, 6, 7, 8
- Batch 2 (after Feature 3): Feature 4 (depends on tool tests)

**Models:**
- Features 1-5: sonnet
- Features 6-8: haiku (simpler, less coverage gap)

**Validation:**
- Each feature must pass mandatory validation protocol
- Orchestrator re-validates each completion claim
- Final check: overall coverage ≥ 50%

---

### Phase 4: Tool Consolidation (27 → 19 tools)

**Prerequisite:** Phase 3 complete (≥50% coverage)

**Features (5):**
1. Bash Tool Consolidation (bash.go + bash_monitored.go → bash.go)
2. Fetch Tool Consolidation (fetch.go + web_fetch.go + agentic_fetch_tool.go → fetch.go)
3. Agent Tools Consolidation (agents.go + agent_*.go → delegate.go)
4. Remove Analytics Tools (track_prompt_usage.go, prompt_analytics.go)
5. Tool Aliasing System (enhance aliases.go, integration testing)

**Parallelization Strategy:**
- Batch 1 (parallel): Features 1, 2, 4
- Batch 2 (after 1,2,3): Feature 5 (needs all consolidations done)

**Models:**
- Features 1-3: sonnet
- Feature 4: haiku (trivial deletion)
- Feature 5: sonnet (testing/integration)

**Validation:**
- Old files deleted
- New consolidated files work for both modes
- Aliases map correctly
- All tests pass
- Backward compatibility verified

---

### Phase 5: TUI Enhancements & Critical Bugs

**Prerequisite:** Phase 4 complete

**Features (8):**
1. Auto-LSP Detection (test/enhance autodetect.go)
2. TUI Settings Panel (test/enhance settings.go)
3. **FIX: Delegate Banner Breaking TUI** (CRITICAL)
4. **FIX: Delegate Reliability** (CRITICAL)
5. **FIX: "/" Key Passthrough** (CRITICAL)
6. **FIX: MCP System Overhaul** (CRITICAL)
7. Unified Delegate with Resource Pool
8. Prompt Repository Import CLI

**Parallelization Strategy:**
- Batch 1 (parallel - critical bugs): Features 3, 4, 5, 6
- Batch 2 (parallel - enhancements): Features 1, 8
- Batch 3 (after Feature 1): Feature 2 (depends on LSP)
- Batch 4 (after Feature 4): Feature 7 (depends on delegate reliability)

**Models:**
- Features 1, 6, 7: opus (use competitive planning for complex features)
- Features 2-5, 8: sonnet

**Validation:**
- All automated tests pass
- **MANUAL TUI TESTING REQUIRED** for bugs 3, 4, 5, 6
- Build succeeds
- No regressions

---

## Validation Gates

### After Phase 0:
```bash
go test ./... -race -timeout 10m
# Expected: ALL PASS (no failures)

go build ./...
# Expected: SUCCESS

go vet ./...
# Expected: CLEAN
```

### After Phase 3:
```bash
go test -coverprofile=coverage.out ./...
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
# Expected: COVERAGE >= 50.0

# Per-package verification
go test -coverprofile=coverage.out ./internal/tui/components/chat/messages/...
# Expected: >= 30%

go test -coverprofile=coverage.out ./internal/tui/page/chat/...
# Expected: >= 30%

go test -coverprofile=coverage.out ./internal/agent/tools/...
# Expected: >= 50%

go test -coverprofile=coverage.out ./internal/agent/...
# Expected: >= 50%
```

### After Phase 4:
```bash
# Verify old files deleted
! test -f internal/agent/tools/bash_monitored.go
! test -f internal/agent/tools/web_fetch.go
! test -f internal/agent/tools/agents.go
# ... etc

# Verify new files exist
test -f internal/agent/tools/bash.go
test -f internal/agent/tools/fetch.go
test -f internal/agent/tools/aliases.go

# Test aliasing
go test ./internal/agent/tools/... -v -run TestAlias
```

### After Phase 5:
```bash
# Automated tests
go test ./internal/lsp/... -v -race
go test ./internal/tui/components/dialogs/settings/... -v -race
go test ./internal/tui/components/banner/... -v -race

# MANUAL TUI TESTING (required!)
# 1. Build: go build -o nexora ./cmd/nexora
# 2. Run: ./nexora
# 3. Test delegate banner fix (should be inline, not giant banner)
# 4. Test delegate reliability (run 10 commands, all should succeed)
# 5. Test "/" key (type "Can you help /" - should work)
# 6. Test MCP system (connect to 3 servers, all should work)
```

### Final Validation:
```bash
# Complete test suite
go test ./... -race -v -timeout 10m

# Build
go build ./...

# Vet
go vet ./...

# Coverage
go test -coverprofile=coverage.out ./...
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
echo "Final Coverage: $COVERAGE%"
# Expected: >= 50%

# Manual TUI testing
# All 4 Phase 5 bugs must be verified fixed
```

---

## Contingency Plans

### If Phase 0 agents fail:
1. Collect agent outputs
2. Identify common failure patterns
3. Fix issues manually or spawn new targeted agents
4. Do NOT proceed to Phase 3 until baseline is clean

### If Phase 3 coverage targets not met:
1. Identify packages still below target
2. Spawn additional test-writing agents
3. Use competitive planning for complex packages
4. Iterate until ≥50% overall coverage

### If Phase 4 consolidation breaks tests:
1. Rollback to baseline
2. Review consolidation approach
3. Use smaller increments (one file at a time)
4. Re-test after each consolidation

### If Phase 5 manual testing reveals bugs:
1. Document exact reproduction steps
2. Create targeted bug fix agents
3. Re-test manually
4. Iterate until all manual tests pass

---

## Autonomous Execution Timeline

**Estimated Duration:** 8-12 hours

### Hour 0-1: Phase 0 (Fix Failing Tests)
- ✅ Agents launched
- 🔄 5/6 agents running, 1/6 complete
- Expected completion: ~30-60 min total

### Hour 1-2: Phase 0 Validation + Phase 3 Launch
- Verify all 6 fixes
- Run full test suite
- Launch 8 Phase 3 agents (7 in parallel)

### Hour 2-5: Phase 3 (Test Coverage)
- Agents write tests + implementation
- Incremental coverage verification
- Target: 36% → 50%+

### Hour 5-6: Phase 3 Validation + Phase 4 Launch
- Verify ≥50% coverage
- Launch 4 Phase 4 agents (3 in parallel)

### Hour 6-8: Phase 4 (Tool Consolidation)
- Consolidate bash, fetch, agent tools
- Implement aliasing
- Verify backward compatibility

### Hour 8-9: Phase 4 Validation + Phase 5 Launch
- Verify consolidation complete
- Launch 8 Phase 5 agents (4 in parallel for bugs)

### Hour 9-11: Phase 5 (TUI Enhancements + Bug Fixes)
- Fix 4 critical bugs
- Implement enhancements
- Automated testing

### Hour 11-12: Final Validation + Manual Testing
- Run complete validation suite
- Perform manual TUI testing
- Create completion report
- **WAKE USER IF MANUAL TESTING NEEDED**

---

## Success Metrics

**Phase 0:**
- ✅ All 9 failing tests pass
- ✅ Build succeeds
- ✅ No vet warnings

**Phase 3:**
- ✅ Overall coverage ≥ 50%
- ✅ All package targets met
- ✅ All tests pass

**Phase 4:**
- ✅ 9 files deleted
- ✅ Aliases work
- ✅ No functionality regression

**Phase 5:**
- ✅ 4 bugs fixed (verified manually)
- ✅ 4 features implemented
- ✅ All automated tests pass

**Final:**
- ✅ All tests pass with race detector
- ✅ Build succeeds
- ✅ Coverage ≥ 50%
- ✅ 4 TUI bugs fixed
- ✅ Ready for RC1 release

---

## Monitoring & Reporting

### Progress Tracking
- Update todo list after each phase
- Create phase completion reports
- Log any blockers or issues

### Completion Notification
When all phases complete:
1. Generate final validation report
2. Document any manual testing requirements
3. Create release checklist
4. Update user on completion

**Autonomous mode: Will continue until complete or blocked on manual testing.**
