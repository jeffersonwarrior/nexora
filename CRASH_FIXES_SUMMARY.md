# Crash Fixes Summary - 2025-12-18

## Issues Fixed

Based on log analysis from `/home/nexora/.nexora/logs/nexora-errors.log`, the following 4 critical crash issues were identified and fixed:

### 1. UTF-8 Encoding Issues (✅ FIXED)
**Problem**: Invalid UTF-8 sequences in files (especially in `~/.local/tools/modelscan/`) were causing view operations to fail with "invalid UTF-8" errors.

**Solution**: Modified `/home/nexora/internal/agent/tools/view.go` to sanitize invalid UTF-8 sequences instead of failing:
- Uses `strings.ToValidUTF8()` to fix invalid sequences
- Provides multi-level fallback sanitization
- Logs warnings for debugging but continues processing
- No more crashes on non-UTF-8 binary files

### 2. Message Sequence Validation Too Strict (✅ FIXED)  
**Problem**: `validateMessageSequence()` in agent.go was failing when tool results were pending, causing session crashes.

**Solution**: Modified `/home/nexora/internal/agent/agent.go` to be more forgiving:
- Changed hard failures to warnings with auto-recovery
- Allows sessions to continue even with sequence issues
- Logs warnings for debugging but doesn't crash

### 3. State Machine Transition Issues (✅ FIXED)
**Problem**: State machine didn't allow re-prompting after errors, preventing recovery loops.

**Solution**: Modified `/home/nexora/internal/agent/state/states.go`:
- Added `StateProcessingPrompt -> StateProcessingPrompt` transition
- Enables better error recovery and re-prompting

### 4. Progress Detection Too Aggressive (✅ FIXED)
**Problem**: Stuck detection was triggering false positives, marking sessions as "stuck" when they were still working.

**Solution**: Modified `/home/nexora/internal/agent/state/progress.go`:
- Increased consecutive error threshold from 3 to 5
- Increased no-progress check from 10 to 15 actions
- Made uniqueness checks more lenient (2 unique targets instead of 3)
- Reduced false stuck detections significantly

## Test Updates
Updated tests in `/home/nexora/internal/agent/state/machine_test.go` to match new thresholds:
- Tests now expect 5 consecutive errors instead of 3 for stuck detection
- All state machine tests passing with new thresholds

## Verification
- All agent tests passing: `go test ./internal/agent/...`
- UTF-8 handling now graceful instead of crashing
- Sessions auto-recover from sequence issues
- State machine allows proper error recovery loops
- Progress detection less aggressive, fewer false positives

## Files Modified
1. `/home/nexora/internal/agent/tools/view.go` - UTF-8 sanitization
2. `/home/nexora/internal/agent/agent.go` - Message sequence validation
3. `/home/nexora/internal/agent/state/states.go` - State transitions
4. `/home/nexora/internal/agent/state/progress.go` - Stuck detection thresholds
5. `/home/nexora/internal/agent/state/machine_test.go` - Updated test thresholds

## Impact
These fixes should resolve the crash issues reported in the logs. The system is now more resilient:
- Handles invalid UTF-8 files gracefully
- Recovers from message sequence issues automatically
- Allows proper error recovery through state machine
- Reduces false stuck detections, allowing longer-running tasks to complete