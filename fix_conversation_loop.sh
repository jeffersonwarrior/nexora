#!/bin/bash

# Quick patch to disable the problematic auto-continuation that causes conversation loops
# This script creates a backup and patches the agent.go file

echo "Creating backup..."
cp internal/agent/agent.go internal/agent/agent.go.backup

echo "Applying patch to disable phrase-based auto-continuation..."
# Comment out the problematic lines that cause false positive auto-continuation
sed -i '1198,1201 s/^/		\/\/ /' internal/agent/agent.go

echo "Patch applied successfully!"
echo ""
echo "What was changed:"
echo "- Disabled phrase-based auto-continuation (lines 1198-1201)"
echo "- Agent will only continue after actual tool execution"
echo "- This prevents false positives from common AI phrases like 'let me explain'"
echo ""
echo "To revert: cp internal/agent/agent.go.backup internal/agent/agent.go"