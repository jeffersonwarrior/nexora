# VCR Testing - Final Resolution

## Root Cause
API keys weren't being sourced from `/home/nexora/.env` before running tests.

## Verification
```bash
# BEFORE (fails - env not sourced):
NEXORA_ANTHROPIC_API_KEY length: 0

# AFTER (works - env sourced):
source /home/nexora/.env
NEXORA_ANTHROPIC_API_KEY length: 108
```

## Test Results
✅ **OpenAI** - PASSES with sourced env
✅ **Session tests** - 88.3% coverage, all pass  
❌ **Anthropic** - Needs `anthropic-version` header
❌ **All agent tests** - Need env sourcing + fresh cassettes

## What Went Wrong
1. Tests ran without `source /home/nexora/.env`
2. VCR recorded 401 Unauthorized errors (not real API responses)
3. Subsequent test runs tried to replay corrupted cassettes
4. Request mismatch = "requested interaction not found"

## Solution
1. Always source env: `source /home/nexora/.env`
2. Delete all corrupted cassettes
3. Re-record with valid API keys
4. Verify tests pass

## Commands to Run

### Delete all cassettes
```bash
rm -rf internal/agent/testdata/TestCoderAgent/*/*.yaml
rm -rf internal/agent/testdata/TestTitleGenerationForNewSession/*/*.yaml
```

### Re-record with valid keys
```bash
source /home/nexora/.env
NEXORA_ANTHROPIC_API_KEY="$NEXORA_ANTHROPIC_API_KEY" \
NEXORA_OPENAI_API_KEY="$NEXORA_OPENAI_API_KEY" \
NEXORA_OPENROUTER_API_KEY="$NEXORA_OPENROUTER_API_KEY" \
NEXORA_ZAI_API_KEY="$NEXORA_ZAI_API_KEY" \
go test ./internal/agent -run "TestCoderAgent" -v
```

### Run tests (playback mode)
```bash
source /home/nexora/.env
go test ./internal/agent -run "TestCoderAgent" -v -count=1
```

## Status: Ready to Re-record
All issues identified. Proceeding with clean re-recording.
