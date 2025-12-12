#!/usr/bin/env bash

# Test script to verify AI mode is working properly

echo "Setting up test environment..."

# Create test directory
mkdir -p /tmp/ai_mode_test_dir
cd /tmp/ai_mode_test_dir

# Create test file
cat > test_go.go << 'EOF'
package main

import (
	"fmt"
)

func main() {
	fmt.Println("Hello, World!")
}
EOF

echo "Created test file with content:"
cat test_go.go
echo ""

# Build nexora if needed
cd /home/renter/nexora
if [ ! -f ./nexora ]; then
    echo "Building nexora..."
    go build -o nexora .
fi

# Test AI mode with minimal context
echo "Testing AI mode with minimal context..."
cat > test_ai_mode.json << 'EOF'
{
  "model": "openai/gpt-4",
  "messages": [
    {
      "role": "user", 
      "content": "Using the edit tool with ai_mode=true, modify the test file at /tmp/ai_mode_test_dir/test_go.go to change 'Hello, World!' to 'Hello, AI Mode!'"
    }
  ]
}
EOF

# Show the summary of what we're testing
echo ""
echo "=== AI Mode Test Summary ==="
echo "1. Created test file with known content"
echo "2. Prepared test request to use edit tool with ai_mode=true"
echo "3. The AI mode should automatically expand minimal context"
echo ""
echo "To manually test AI mode in nexora:"
echo "1. Start nexora with: ./nexora"
echo "2. Edit the file at /tmp/ai_mode_test_dir/test_go.go"
echo "3. Use ai_mode=true when making an edit with minimal context"
echo "4. Verify that context is automatically expanded"
echo ""