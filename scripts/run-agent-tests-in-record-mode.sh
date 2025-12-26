#!/bin/bash
# Run agent tests in VCR record mode
# This creates new cassettes with your API keys
# Usage: ./scripts/run-agent-tests-in-record-mode.sh

set -e

ENV_FILE="/home/nexora/.env"

# Load environment from .env if it exists
if [ -f "$ENV_FILE" ]; then
    set -a
    source "$ENV_FILE"
    set +a
fi

echo "=============================================="
echo "Running Agent Tests in VCR Record Mode"
echo "=============================================="
echo ""
echo "This will:"
echo "1. Make real API calls (uses your credits)"
echo "2. Record responses to cassettes"
echo "3. Increase test coverage"
echo ""

# Check for at least one API key
has_key=false
for var in NEXORA_ANTHROPIC_API_KEY NEXORA_OPENAI_API_KEY NEXORA_OPENROUTER_API_KEY NEXORA_ZAI_API_KEY; do
    if [ -n "${!var}" ]; then
        echo "✅ $var is set"
        has_key=true
    else
        echo "⚠️  $var not set"
    fi
done

if [ "$has_key" = false ]; then
    echo ""
    echo "❌ No API keys found. Set at least one:"
    echo "   export NEXORA_ANTHROPIC_API_KEY='your_key'"
    echo "   export NEXORA_OPENAI_API_KEY='your_key'"
    exit 1
fi

echo ""
echo "Starting tests in record mode..."
echo "=============================================="

# Run tests - VCR will automatically record if cassettes don't match
# Using -count=1 to disable test caching
NEXORA_ANTHROPIC_API_KEY="$NEXORA_ANTHROPIC_API_KEY" \
NEXORA_OPENAI_API_KEY="$NEXORA_OPENAI_API_KEY" \
NEXORA_OPENROUTER_API_KEY="$NEXORA_OPENROUTER_API_KEY" \
NEXORA_ZAI_API_KEY="$NEXORA_ZAI_API_KEY" \
go test ./internal/agent -run TestCoderAgent -v -count=1

echo ""
echo "=============================================="
echo "Tests complete!"
echo "=============================================="
echo ""
echo "Recorded cassettes saved to:"
echo "  internal/agent/testdata/TestCoderAgent/"
echo ""
echo "To run tests without recording (playback mode):"
echo "  go test ./internal/agent -run TestCoderAgent -v"
echo ""
echo "To check coverage:"
echo "  go test -coverprofile=cov.out ./internal/agent"
echo "  go tool cover -func=cov.out | grep total"
