#!/bin/bash
# Verify synthetic keys are set up correctly
# Usage: ./scripts/verify-keys-setup.sh

echo "=========================================="
echo "Nexora Synthetic Keys Verification"
echo "=========================================="
echo ""

ERRORS=0

# Check if .env file exists
if [ -f /home/nexora/.env ]; then
    echo "✅ .env file exists"
else
    echo "❌ .env file missing"
    echo "   Create it with: cp /home/nexora/.env.example /home/nexora/.env"
    ERRORS=$((ERRORS + 1))
fi

echo ""
echo "Checking API keys..."
echo ""

# Load environment
if [ -f /home/nexora/.env ]; then
    set -a
    source /home/nexora/.env
    set +a
fi

# Check each provider directly
if [ -n "$NEXORA_ANTHROPIC_API_KEY" ]; then
    masked="${NEXORA_ANTHROPIC_API_KEY:0:8}..."
    echo "✅ Anthropic API key set ($masked)"
else
    echo "⚠️  Anthropic API key NOT set"
    ERRORS=$((ERRORS + 1))
fi

if [ -n "$NEXORA_OPENAI_API_KEY" ]; then
    masked="${NEXORA_OPENAI_API_KEY:0:8}..."
    echo "✅ OpenAI API key set ($masked)"
else
    echo "⚠️  OpenAI API key NOT set"
    ERRORS=$((ERRORS + 1))
fi

if [ -n "$NEXORA_OPENROUTER_API_KEY" ]; then
    masked="${NEXORA_OPENROUTER_API_KEY:0:8}..."
    echo "✅ OpenRouter API key set ($masked)"
else
    echo "⚠️  OpenRouter API key NOT set"
    ERRORS=$((ERRORS + 1))
fi

if [ -n "$NEXORA_ZAI_API_KEY" ]; then
    masked="${NEXORA_ZAI_API_KEY:0:8}..."
    echo "✅ ZAI API key set ($masked)"
else
    echo "⚠️  ZAI API key NOT set"
    ERRORS=$((ERRORS + 1))
fi

echo ""
echo "=========================================="

if [ $ERRORS -eq 0 ]; then
    echo "✅ All keys are set up correctly!"
    echo ""
    echo "You can now run VCR recording:"
    echo "  ./scripts/run-agent-tests-in-record-mode.sh"
    exit 0
else
    echo "❌ $ERRORS issue(s) found"
    echo ""
    echo "To fix:"
    echo "  1. cp /home/nexora/.env.example /home/nexora/.env"
    echo "  2. nano /home/nexora/.env"
    echo "  3. source /home/nexora/.env"
    echo "  4. ./scripts/verify-keys-setup.sh"
    exit 1
fi
