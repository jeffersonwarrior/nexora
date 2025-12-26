#!/bin/bash
# Record VCR cassettes for agent integration tests
# Usage: ./scripts/record-vcr.sh [provider]
# 
# Examples:
#   ./scripts/record-vcr.sh              # Record all providers
#   ./scripts/record-vcr.sh anthropic   # Record only Anthropic
#   ./scripts/record-vcr.sh openai       # Record only OpenAI

set -e

PROVIDER="${1:-all}"
ENV_FILE="/home/nexora/.env"

# Load environment from .env if it exists
if [ -f "$ENV_FILE" ]; then
    set -a
    source "$ENV_FILE"
    set +a
fi

echo "=============================================="
echo "VCR Cassette Recording"
echo "=============================================="
echo "Provider: $PROVIDER"
echo ""

# Check if required env vars are set
check_env() {
    local var_name="$1"
    local value="${!var_name}"
    if [ -z "$value" ]; then
        echo "⚠️  Warning: $var_name is not set"
        return 1
    else
        echo "✅ $var_name is set"
        return 0
    fi
}

# Record tests for a specific provider
record_provider() {
    local pattern="$1"
    local env_vars="$2"
    
    echo ""
    echo "Recording tests matching: $pattern"
    echo "---"
    
    $env_vars go test ./internal/agent -run "$pattern" -v -timeout 30m
}

# Run tests
case "$PROVIDER" in
    anthropic)
        echo "Checking Anthropic API key..."
        check_env "NEXORA_ANTHROPIC_API_KEY"
        record_provider "TestCoderAgent/anthropic" "NEXORA_ANTHROPIC_API_KEY=$NEXORA_ANTHROPIC_API_KEY"
        ;;
        
    openai)
        echo "Checking OpenAI API key..."
        check_env "NEXORA_OPENAI_API_KEY"
        record_provider "TestCoderAgent/openai-gpt-5" "NEXORA_OPENAI_API_KEY=$NEXORA_OPENAI_API_KEY"
        ;;
        
    openrouter)
        echo "Checking OpenRouter API key..."
        check_env "NEXORA_OPENROUTER_API_KEY"
        record_provider "TestCoderAgent/openrouter" "NEXORA_OPENROUTER_API_KEY=$NEXORA_OPENROUTER_API_KEY"
        ;;
        
    zai)
        echo "Checking ZAI API key..."
        check_env "NEXORA_ZAI_API_KEY"
        record_provider "TestCoderAgent/zai-glm" "NEXORA_ZAI_API_KEY=$NEXORA_ZAI_API_KEY"
        ;;
        
    all)
        echo "Checking all API keys..."
        echo ""
        
        # Anthropic
        echo "---"
        if check_env "NEXORA_ANTHROPIC_API_KEY"; then
            record_provider "TestCoderAgent/anthropic" "NEXORA_ANTHROPIC_API_KEY=$NEXORA_ANTHROPIC_API_KEY"
        fi
        
        # OpenAI
        echo "---"
        if check_env "NEXORA_OPENAI_API_KEY"; then
            record_provider "TestCoderAgent/openai-gpt-5" "NEXORA_OPENAI_API_KEY=$NEXORA_OPENAI_API_KEY"
        fi
        
        # OpenRouter
        echo "---"
        if check_env "NEXORA_OPENROUTER_API_KEY"; then
            record_provider "TestCoderAgent/openrouter" "NEXORA_OPENROUTER_API_KEY=$NEXORA_OPENROUTER_API_KEY"
        fi
        
        # ZAI
        echo "---"
        if check_env "NEXORA_ZAI_API_KEY"; then
            record_provider "TestCoderAgent/zai-glm" "NEXORA_ZAI_API_KEY=$NEXORA_ZAI_API_KEY"
        fi
        ;;
        
    *)
        echo "Usage: $0 [anthropic|openai|openrouter|zai|all]"
        exit 1
        ;;
esac

echo ""
echo "=============================================="
echo "Recording complete!"
echo "=============================================="
echo ""
echo "To run tests without recording:"
echo "  go test ./internal/agent -run TestCoderAgent -v"
echo ""
echo "To check coverage:"
echo "  go test -coverprofile=cov.out ./internal/agent"
echo "  go tool cover -func=cov.out | grep total"
