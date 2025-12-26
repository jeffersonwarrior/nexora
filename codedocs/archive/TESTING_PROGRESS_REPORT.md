# Session Title Fix & VCR Testing - Progress Report

## Summary
Working on fixing session title re-generation and re-recording VCR cassettes for agent tests.

## Completed Work

### 1. Session Title Fix ✅
**File Modified:** `/home/nexora/internal/agent/agent.go` (line 540)

**Change:**
```go
// Before:
if currentSession.MessageCount == 0 {

// After:
if currentSession.MessageCount == 0 || currentSession.Title == "New Session" {
```

**Rationale:** Ensures that sessions with the default "New Session" title get retitled even if MessageCount is not 0 (handles race conditions and edge cases).

**Reference:** TODO.md - Session Title Re-generation issue

### 2. New Test Added ✅
**File Created:** `/home/nexora/internal/agent/session_title_test.go`

**Test:** `TestTitleGenerationForNewSession`
- Verifies sessions with "New Session" title get retitled on first message
- Tests all 4 providers (anthropic, openai, openrouter, zai)
- Validates title is updated, non-empty, and <= 50 characters

### 3. VCR Cassettes Re-recorded ✅
**Providers Completed:**
- ✅ **Anthropic** - All 14 test cassettes re-recorded
- ✅ **OpenAI** - All 14 test cassettes re-recorded
- ⏳ **OpenRouter** - Needs re-recording
- ⏳ **ZAI** - Partially completed

**Total Cassettes Created:** ~40+ new cassettes

**Issue Resolved:** Previous cassettes contained 401 Unauthorized errors from invalid/missing API keys. Fresh recordings with valid keys now available.

## Test Coverage

**Current Coverage:** ~26.6%
**Target Coverage:** 50% (per TODO.md Phase 3)

**Coverage by Package:**
- internal/agent: Needs re-testing after VCR re-recording
- internal/session: 88.3% ✅
- internal/config: 49.7%
- internal/db: 32.0%
- Other packages: Various coverage levels

## What's Working

1. ✅ Session title generation fix implemented
2. ✅ New test validates title generation
3. ✅ Anthropic VCR cassettes re-recorded
4. ✅ OpenAI VCR cassettes re-recorded  
5. ✅ Test infrastructure (VCR, fixtures) functional

## What's Pending

### High Priority
1. **Complete VCR re-recording:**
   - OpenRouter cassettes (remaining tests)
   - ZAI cassettes (remaining tests)
   - Fix any failing tests

2. **Run full test suite:**
   - Verify all agent tests pass with new cassettes
   - Check overall test coverage

3. **Fix any test failures:**
   - Investigate zai-glm4.6 failures
   - Verify anthropic tests pass

### Medium Priority
4. **Documentation:**
   - Update SESSION_TITLE_FIX.md with final solution
   - Document VCR re-recording process

5. **Additional tests:**
   - Add edge case tests for title generation
   - Test with different placeholder titles

## Next Steps

### Immediate (Next 1-2 hours)
1. Wait for OpenRouter/ZAI tests to complete or re-run them
2. Run full agent test suite: `go test ./internal/agent -v`
3. Verify TestTitleGenerationForNewSession passes for all providers
4. Check coverage: `go tool cover -func=cov.out | grep total`

### Short-term (Today)
5. Complete all VCR re-recording if needed
6. Run full test suite with coverage: `go test ./... -coverprofile=cov.out`
7. Target 50% coverage (currently ~26.6%, was 32.4% before)
8. Update TODO.md with completed items

### This Week
9. Complete Phase 3 (Test Coverage) of v0.29.1-RC1
10. Move to Phase 4 (Tool Consolidation)
11. Continue to Phase 5 (TUI Enhancements)

## Commands Reference

### Running Tests
```bash
# Run specific test
go test ./internal/agent -run "TestTitleGenerationForNewSession/anthropic-sonnet" -v

# Run all agent tests
go test ./internal/agent -v

# Run with coverage
go test ./internal/agent -coverprofile=cov.out

# Check coverage
go tool cover -func=cov.out | grep total
```

### VCR Re-recording
```bash
# Re-record specific provider
source /home/nexora/.env
NEXORA_ANTHROPIC_API_KEY="$NEXORA_ANTHROPIC_API_KEY" \
  go test ./internal/agent -run "TestCoderAgent/anthropic-sonnet" -v

# Delete cassettes to force re-recording
rm -rf internal/agent/testdata/TestCoderAgent/anthropic-sonnet/*.yaml
```

### Coverage Report
```bash
# Full coverage
go test ./... -coverprofile=full_cov.out
go tool cover -func=full_cov.out | grep total
```

## Files Modified/Created

### Modified
- `/home/nexora/internal/agent/agent.go` - Session title fix (line 540)

### Created
- `/home/nexora/internal/agent/session_title_test.go` - New test
- `/home/nexora/SESSION_TITLE_FIX.md` - Design document
- Multiple VCR cassette files in `internal/agent/testdata/TestCoderAgent/`

### Deleted (for re-recording)
- Old VCR cassettes with 401 errors

## Notes

### VCR Cassette Structure
Cassettes stored in: `internal/agent/testdata/TestCoderAgent/{provider}/{test_name}.yaml`

Example:
```
internal/agent/testdata/TestCoderAgent/
├── anthropic-sonnet/
│   ├── simple_test.yaml
│   ├── read_a_file.yaml
│   └── ...
├── openai-gpt-5/
│   └── ...
└── ...
```

### Test Naming
- TestCoderAgent: Main integration tests
- TestTitleGenerationForNewSession: Specific title generation test

## Issue Resolution

### Session Title Problem
**Original Issue:** Sessions with "New Session" title don't get retitled on first message.

**Root Cause:** Only checked `MessageCount == 0`, didn't check if title was placeholder.

**Solution:** Added `|| currentSession.Title == "New Session"` check.

### VCR Cassette Problem
**Original Issue:** Cassettes contained 401 Unauthorized errors.

**Root Cause:** Recorded with invalid/missing API keys.

**Solution:** Deleted old cassettes, re-recorded with valid API keys.

## Success Criteria

- [x] Session title fix implemented
- [x] New test added
- [x] Anthropic cassettes re-recorded
- [x] OpenAI cassettes re-recorded
- [ ] OpenRouter cassettes re-recorded
- [ ] ZAI cassettes re-recorded
- [ ] All agent tests pass
- [ ] Coverage >= 50%
- [ ] TODO.md updated

---

**Last Updated:** 2025-12-26
**Status:** In Progress - VCR re-recording ~75% complete
