#!/usr/bin/env bash

# Auto lint and code quality script for Crush
# Usage: ./scripts/lint.sh

set -e

echo "🔧 Running code quality checks..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

echo -e "${BLUE}📁 Working directory: $(pwd)${NC}"

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to run a check with status
run_check() {
    local cmd="$1"
    local description="$2"
    
    echo -e "${YELLOW}Running: $description${NC}"
    if eval "$cmd"; then
        echo -e "${GREEN}✓ $description${NC}"
        return 0
    else
        echo -e "${RED}✗ $description${NC}"
        return 1
    fi
}

# Track failures
failures=0

echo -e "${BLUE}🧹 Cleaning up code formatting...${NC}"

# Format code
if command_exists gofumpt; then
    run_check "gofumpt -w ." "Formatting with gofumpt"
elif command_exists goimports; then
    run_check "goimports -w ." "Formatting with goimports"
else
    run_check "gofmt -w ." "Formatting with gofmt"
fi

echo -e "${BLUE}🔍 Running static analysis...${NC}"

# Go vet
if ! run_check "go vet ./..." "Static analysis (go vet)"; then
    failures=$((failures + 1))
fi

# Go mod verification
if ! run_check "go mod verify" "Module verification"; then
    failures=$((failures + 1))
fi

# Check for unused dependencies
if command_exists go; then
    if ! run_check "go mod tidy" "Cleaning module dependencies"; then
        failures=$((failures + 1))
    fi
fi

# Run golangci-lint if available
if command_exists golangci-lint; then
    echo -e "${BLUE}🏃 Running golangci-lint...${NC}"
    if golangci-lint run; then
        echo -e "${GREEN}✓ golangci-lint${NC}"
    else
        echo -e "${RED}✗ golangci-lint${NC}"
        failures=$((failures + 1))
    fi
fi

if run_check "go test ./... -timeout=10m" "Running test suite"; then
    echo -e "${RED}✗ Some tests failed${NC}"
    failures=$((failures + 1))
fi

# Check for TODO/FIXME comments
echo -e "${BLUE}📝 Checking TODO/FIXME comments...${NC}"
TODO_COUNT=$(grep -r "TODO\|FIXME" --include="*.go" . | wc -l || true)
if [ "$TODO_COUNT" -gt 0 ]; then
    echo -e "${YELLOW}⚠️ Found $TODO_COUNT TODO/FIXME comments${NC}"
    grep -r "TODO\|FIXME" --include="*.go" . | head -10
else
    echo -e "${GREEN}✓ No TODO/FIXME comments found${NC}"
fi

# Check for common issues
echo -e "${BLUE}🔎 Checking for common issues...${NC}"

# Check for interface{} that could be any
INTERFACE_COUNT=$(grep -r "interface{}" --include="*.go" . | wc -l || true)
if [ "$INTERFACE_COUNT" -gt 0 ]; then
    echo -e "${YELLOW}⚠️ Found $INTERFACE_COUNT interface{} that could be 'any'${NC}"
else
    echo -e "${GREEN}✓ No interface{} found${NC}"
fi

# Check for context.TODO usage
TODO_CTX_COUNT=$(grep -r "context\.TODO" --include="*.go" . | wc -l || true)
if [ "$TODO_CTX_COUNT" -gt 0 ]; then
    echo -e "${YELLOW}⚠️ Found $TODO_CTX_COUNT context.TODO() calls${NC}"
else
    echo -e "${GREEN}✓ No context.TODO() calls found${NC}"
fi

# Build check
echo -e "${BLUE}🏗️  Checking build...${NC}"
if run_check "go build ." "Build check"; then
    echo -e "${GREEN}✓ Project builds successfully${NC}"
else
    echo -e "${RED}✗ Build failed${NC}"
    failures=$((failures + 1))
fi

# Summary
echo "=================================="
if [ $failures -eq 0 ]; then
    echo -e "${GREEN}🎉 All checks passed! Code is ready.${NC}"
    exit 0
else
    echo -e "${RED}❌ $failures check(s) failed. Please fix issues before committing.${NC}"
    exit 1
fi