#!/bin/bash
# Script to verify Ollama models are detected with proper context

echo "Testing local model detection..."
curl -s http://localhost:11434/api/tags | jq '.'

echo ""
echo "Testing model details for llama3.1:8b..."
curl -s http://localhost:11434/api/show -d '{"name":"llama3.1:8b"}' | jq '.details.context_length // .model_info.context_length // "Not found in details"'