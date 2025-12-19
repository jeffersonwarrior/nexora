#!/bin/bash

# Auto-fix utility for GitHub workflows
# Usage: ./autofix.sh <fix-type> [output-file]

set -e

FIX_TYPE=${1:-"all"}
OUTPUT_FILE=${2:-"/tmp/autofix.log"}

echo "🔧 Auto-fix utility starting..." | tee "$OUTPUT_FILE"
echo "Fix type: $FIX_TYPE" | tee -a "$OUTPUT_FILE"
echo "Timestamp: $(date)" | tee -a "$OUTPUT_FILE"

# Function to log actions
log_action() {
    echo "[$(date +%H:%M:%S)] $1" | tee -a "$OUTPUT_FILE"
}

# Function to attempt go fixes
fix_go_dependencies() {
    log_action "🔍 Checking Go dependencies..."
    
    if [[ -f go.mod ]]; then
        # Check for common issues
        if go mod tidy 2>&1 | grep -q "go: errors"; then
            log_action "💥 go.mod issues detected, fixing..."
            go mod download
            go mod tidy
            go get -u ./...
            go mod tidy
            log_action "✅ Dependencies fixed"
        fi
    fi
}

# Function to fix formatting
fix_formatting() {
    log_action "🎨 Checking code formatting..."
    
    UNFORMATTED=$(gofmt -s -l . || true)
    if [[ -n "$UNFORMATTED" ]]; then
        log_action "💥 Formatting issues found, fixing..."
        echo "$UNFORMATTED" | while read file; do
            log_action "  Formatting: $file"
        done
        gofmt -s -w .
        
        # Also fix imports if goimports is available
        if command -v goimports &> /dev/null; then
            log_action "📦 Fixing imports with goimports..."
            find . -name "*.go" -exec goimports -w {} \;
        fi
        log_action "✅ Formatting fixed"
    fi
}

# Function to fix linting issues
fix_linting() {
    log_action "🔍 Checking linting issues..."
    
    if command -v golangci-lint &> /dev/null; then
        LINT_OUTPUT=$(golangci-lint run --disable typecheck,unused,staticcheck 2>&1 || true)
        if echo "$LINT_OUTPUT" | grep -q "issues found"; then
            log_action "💥 Lint issues found, attempting auto-fix..."
            golangci-lint run --fix --disable typecheck,unused,staticcheck || true
            log_action "✅ Linting fixes applied"
        fi
    fi
}

# Function to fix build issues
fix_build() {
    log_action "🔧 Checking build issues..."
    
    # Try a test build
    BUILD_OUTPUT=$(go build -o /tmp/test-build . 2>&1 || true)
    if [[ -n "$BUILD_OUTPUT" ]]; then
        log_action "💥 Build issues detected: $BUILD_OUTPUT"
        
        # Fix common build patterns
        if echo "$BUILD_OUTPUT" | grep -q "cannot find"; then
            log_action "🔧 Fixing missing packages..."
            go mod tidy
            go mod download
        fi
        
        if echo "$BUILD_OUTPUT" | grep -q "undefined:"; then
            log_action "🔧 Checking for dependency issues..."
            go mod tidy
            go get ./...
        fi
        
        # Retry build
        if go build -o /tmp/test-build . 2>/dev/null; then
            log_action "✅ Build issues fixed"
        else
            log_action "❌ Build issues persist"
        fi
    fi
}

# Function to fix test issues
fix_tests() {
    log_action "🧪 Checking test issues..."
    
    # Run tests with specific timeout
    TEST_OUTPUT=$(timeout 5m go test ./... 2>&1 || true)
    
    if echo "$TEST_OUTPUT" | grep -q "FAIL\|error"; then
        log_action "💥 Test issues detected"
        
        # Fix common test issues
        if echo "$TEST_OUTPUT" | grep -q "no matching versions"; then
            log_action "🔧 Updating dependencies for tests..."
            fix_go_dependencies
        fi
        
        if echo "$TEST_OUTPUT" | grep -q "build constraints"; then
            log_action "🔧 Fixing build constraints..."
            go mod tidy
        fi
    fi
}

# Function to fix go.sum issues
fix_go_sum() {
    log_action "🔐 Checking go.sum..."
    
    if [[ -f go.mod ]]; then
        # Remove and regenerate go.sum
        rm -f go.sum
        go mod download
        go mod verify
        log_action "✅ go.sum regenerated"
    fi
}

# Main execution logic
case "$FIX_TYPE" in
    "dependencies")
        fix_go_dependencies
        ;;
    "formatting")
        fix_formatting
        ;;
    "linting")
        fix_linting
        ;;
    "build")
        fix_build
        ;;
    "tests")
        fix_tests
        ;;
    "go-sum")
        fix_go_sum
        ;;
    "all")
        log_action "🚀 Running complete auto-fix suite..."
        fix_go_dependencies
        fix_formatting
        fix_linting
        fix_build
        fix_tests
        ;;
    *)
        echo "❌ Unknown fix type: $FIX_TYPE"
        echo "Usage: $0 <dependencies|formatting|linting|build|tests|go-sum|all>"
        exit 1
        ;;
esac

# Check if any changes were made
if [[ -n $(git status --porcelain) ]]; then
    log_action "✅ Changes made during auto-fix"
    echo "Modified files:"
    git status --porcelain
    exit 0
else
    log_action "ℹ️ No changes needed"
    exit 1
fi