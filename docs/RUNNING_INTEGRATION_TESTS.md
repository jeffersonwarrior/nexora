# Running Agent Integration Tests with VCR

## Quick Start

### Option 1: Use Existing Cassettes (Fast, No API Keys)

Tests will fail if cassettes don't match current code:

```bash
go test ./internal/agent -run TestCoderAgent -v
# Will fail - cassettes don't match current request format
```

### Option 2: Record New Cassettes (Uses API Keys)

This will create new cassettes with your synthetic keys:

```bash
# Set your API keys
export NEXORA_ANTHROPIC_API_KEY="your_synthetic_key"
export NEXORA_OPENAI_API_KEY="your_synthetic_key"
export NEXORA_OPENROUTER_API_KEY="your_synthetic_key"
export NEXORA_ZAI_API_KEY="your_synthetic_key"

# Or use the helper script
./scripts/run-agent-tests-in-record-mode.sh
```

## What This Does

1. **Makes real API calls** to each provider
2. **Records responses** to YAML files in `internal/agent/testdata/TestCoderAgent/`
3. **Increases coverage** by ~10-15%
4. **Enables CI testing** without API keys (playback mode)

## Estimated Costs

| Provider | Test Count | Est. Cost |
|----------|------------|-----------|
| Anthropic | ~13 tests | ~$0.50-1.00 |
| OpenAI | ~13 tests | ~$0.50-1.00 |
| OpenRouter | ~13 tests | ~$0.25-0.50 |
| ZAI | ~13 tests | ~$0.25-0.50 |
| **Total** | **~52 tests** | **~$1.50-3.00** |

## After Recording

Once cassettes are recorded:

```bash
# Run tests without API keys (fast playback)
go test ./internal/agent -run TestCoderAgent -v

# Check coverage increase
go test -coverprofile=cov.out ./internal/agent
go tool cover -func=cov.out | grep total
```

## Cassette Files

Recorded files stored in:
```
internal/agent/testdata/TestCoderAgent/
├── anthropic-sonnet/
│   ├── simple_test.yaml
│   ├── read_a_file.yaml
│   ├── update_a_file.yaml
│   └── ...
├── openai-gpt-5/
│   └── ...
├── openrouter-kimi-k2/
│   └── ...
└── zai-glm4.6/
    └── ...
```

## Troubleshooting

### "Request interaction not found"

The cassette was recorded with different request data. Solution: Re-record.

### "API key not set"

Set the environment variable before running tests:
```bash
export NEXORA_OPENAI_API_KEY="your_key"
```

### Tests timeout

Some tests run long LLM operations. Increase timeout:
```bash
go test ./internal/agent -run TestCoderAgent -v -timeout 30m
```

## Committing Cassettes

Cassettes are test fixtures - commit them to git:
```bash
git add internal/agent/testdata/TestCoderAgent/
git commit -m "test: add VCR cassettes for agent integration tests"
git push
```

## Impact on Coverage

| Before Recording | After Recording |
|-----------------|-----------------|
| 28.5% agent coverage | ~40-45% agent coverage |
| 36.2% overall coverage | ~45-50% overall coverage |

## Files Created/Modified

- `/home/nexora/scripts/run-agent-tests-in-record-mode.sh` - Helper script
- `/home/nexora/scripts/record-vcr.sh` - Per-provider recording
- `/home/nexora/docs/VCR_RECORDING.md` - Full documentation
- `internal/agent/testdata/TestCoderAgent/*/*.yaml` - Recorded cassettes
