# VCR Cassette Recording Guide

## What is VCR?

VCR (Video Cassette Recorder) records HTTP interactions for replay in tests:
- **First run:** Makes real API calls, records them to YAML files
- **Subsequent runs:** Replays recorded responses (no real API calls needed)
- **Benefits:** Fast, deterministic, no API keys needed after recording

## Setup

### 1. Set Environment Variables

Create `/home/nexora/.env` with your synthetic keys:

```bash
# Required for agent tests
NEXORA_ANTHROPIC_API_KEY=your_synthetic_key_here
NEXORA_OPENAI_API_KEY=your_synthetic_key_here
NEXORA_OPENROUTER_API_KEY=your_synthetic_key_here
NEXORA_ZAI_API_KEY=your_synthetic_key_here

# Optional (for other tests)
NEXORA_GOOGLE_API_KEY=
NEXORA_BEDROCK_ACCESS_KEY=
NEXORA_BEDROCK_SECRET_KEY=
NEXORA_BEDROCK_REGION=us-east-1
```

### 2. Load Environment

```bash
# Option A: Export in shell
export NEXORA_ANTHROPIC_API_KEY="your_key"
export NEXORA_OPENAI_API_KEY="your_key"

# Option B: Source .env file
source /home/nexora/.env

# Option C: Run test with env vars inline
NEXORA_ANTHROPIC_API_KEY="key" NEXORA_OPENAI_API_KEY="key" go test ./internal/agent -run TestCoderAgent -v
```

## Recording Cassettes

### Option 1: Run All Agent Tests (Records All Providers)

```bash
# Record cassettes for all providers
NEXORA_ANTHROPIC_API_KEY="..." \
NEXORA_OPENAI_API_KEY="..." \
NEXORA_OPENROUTER_API_KEY="..." \
NEXORA_ZAI_API_KEY="..." \
go test ./internal/agent -run TestCoderAgent -v -record

# Note: VCR records automatically when cassettes don't exist
```

### Option 2: Record Single Provider

```bash
# Record only Anthropic
NEXORA_ANTHROPIC_API_KEY="..." \
go test ./internal/agent -run "TestCoderAgent/anthropic-sonnet" -v

# Record only OpenAI
NEXORA_OPENAI_API_KEY="..." \
go test ./internal/agent -run "TestCoderAgent/openai-gpt-5" -v
```

### Option 3: Record Single Test Case

```bash
# Record only "simple test" for anthropic
NEXORA_ANTHROPIC_API_KEY="..." \
go test ./internal/agent -run "TestCoderAgent/anthropic-sonnet/simple_test" -v
```

## Running Tests (Playback Mode)

Once cassettes are recorded, tests run without API keys:

```bash
# No API keys needed - replays recorded cassettes
go test ./internal/agent -run TestCoderAgent -v

# Check coverage
go test -coverprofile=cov.out ./internal/agent
go tool cover -func=cov.out | grep total
```

## Cassette Storage

Recorded cassettes are stored in:

```
internal/agent/testdata/TestCoderAgent/
├── anthropic-sonnet/      # Needs recording
│   ├── simple_test.yaml
│   ├── read_a_file.yaml
│   ├── update_a_file.yaml
│   └── ...
├── openai-gpt-5/          # Already exists
│   ├── simple_test.yaml
│   └── ...
├── openrouter-kimi-k2/    # Needs recording
│   └── ...
└── zai-glm4.6/            # Needs recording
    └── ...
```

## Re-recording Cassettes

If you want fresh recordings (e.g., different model responses):

```bash
# Delete old cassettes
rm -rf internal/agent/testdata/TestCoderAgent/*/simple_test.yaml

# Re-record with new keys
NEXORA_ANTHROPIC_API_KEY="new_key" go test ./internal/agent -run "TestCoderAgent/anthropic-sonnet/simple_test" -v
```

## Troubleshooting

### "API key not set" Error

```bash
# Check if env var is loaded
echo $NEXORA_ANTHROPIC_API_KEY

# Export it
export NEXORA_ANTHROPIC_API_KEY="your_key"
```

### Test Fails with " cassette not found"

VCR automatically records if cassette doesn't exist. Just run the test:

```bash
NEXORA_ANTHROPIC_API_KEY="..." go test ./internal/agent -run TestName -v
# VCR will create the cassette file automatically
```

### Cassette Recording Too Slow

- Tests record real API calls which take time
- First run will be slow (recording)
- Subsequent runs will be fast (playback)
- This is expected behavior

### Different Response Each Time

This is normal for LLMs. Options:
1. Accept variability (tests check structure, not exact text)
2. Use deterministic prompts (same input = similar output)
3. Mock responses (bypass VCR for unit tests)

## Coverage Impact

Recording these cassettes will significantly increase coverage:

| Provider | Current Coverage | After Recording |
|----------|-----------------|-----------------|
| anthropic-sonnet | 0% | ~15% |
| openai-gpt-5 | ~10% | ~15% |
| openrouter-kimi-k2 | 0% | ~15% |
| zai-glm4.6 | 0% | ~15% |

**Estimated total coverage increase:** +10-15%

## Quick Reference

```bash
# Record all tests (needs ~30 minutes with API calls)
NEXORA_*_API_KEY="keys" go test ./internal/agent -run TestCoderAgent -v

# Record single provider (~5-10 minutes)
NEXORA_ANTHROPIC_API_KEY="key" go test ./internal/agent -run "anthropic" -v

# Run tests without recording (fast)
go test ./internal/agent -run TestCoderAgent -v

# Check coverage
go test -coverprofile=cov.out ./internal/agent
go tool cover -func=cov.out | grep total
```

## After Recording

Once cassettes are recorded:

1. **Commit to git** (these are test fixtures)
2. **Tests will run in CI** without API keys
3. **Coverage will increase** by 10-15%
4. **CI will be fast** (playback mode)

```bash
git add internal/agent/testdata/TestCoderAgent/
git commit -m "test: add VCR cassettes for integration tests"
git push
```
