#!/bin/bash
# Test runner script with resource limits to prevent tests from running forever

set -euo pipefail

# Default values
TIMEOUT=${TEST_TIMEOUT:-10m}
MEMORY_LIMIT=${TEST_MEMORY_LIMIT:-2G}
CPU_LIMIT=${TEST_CPU_LIMIT:-2}
TIMEOUT_KILL_AFTER=${TEST_TIMEOUT_KILL_AFTER:-30s}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Running tests with limits:${NC}"
echo "  Timeout: ${TIMEOUT}"
echo "  Memory Limit: ${MEMORY_LIMIT}"
echo "  CPU Limit: ${CPU_LIMIT}"
echo "  Timeout Kill After: ${TIMEOUT_KILL_AFTER}"
echo ""

# Check if timeout command is available
if ! command -v timeout &> /dev/null; then
    echo -e "${YELLOW}Warning: 'timeout' command not found. Install coreutils for better timeout handling.${NC}"
    # No timeout command available, run tests directly with Go's built-in timeout
    go test -timeout="${TIMEOUT}" "$@"
    exit $?
fi

# Check if we can limit memory (Linux only with ulimit or Docker)
LIMIT_MEMORY=false
if command -v prlimit &> /dev/null; then
    # Modern Linux with prlimit
    LIMIT_MEMORY=true
elif [[ "$OSTYPE" == "linux-gnu"* ]] && command -v ulimit &> /dev/null; then
    # Try to set memory limit with ulimit (Linux)
    if ulimit -Sv $(echo "$MEMORY_LIMIT" | sed 's/G/*1024*1024/' | sed 's/M/*1024/' | bc) &>/dev/null; then
        LIMIT_MEMORY=true
    fi
fi

# Run tests with resource limits
if [ "$LIMIT_MEMORY" = true ]; then
    echo -e "${GREEN}Running with memory and timeout limits${NC}"
    if command -v prlimit &> /dev/null; then
        # Use prlimit for memory limits (modern Linux)
        prlimit --as=$(echo "$MEMORY_LIMIT" | sed 's/G/*1024*1024*1024/' | sed 's/M/*1024*1024/' | bc) \
        timeout --kill-after="${TIMEOUT_KILL_AFTER}" "${TIMEOUT}" \
        go test -timeout="${TIMEOUT}" "$@"
    else
        # Fallback: timeout + ulimit
        timeout --kill-after="${TIMEOUT_KILL_AFTER}" "${TIMEOUT}" \
        go test -timeout="${TIMEOUT}" "$@"
    fi
else
    echo -e "${YELLOW}Memory limiting not available, using timeout only${NC}"
    timeout --kill-after="${TIMEOUT_KILL_AFTER}" "${TIMEOUT}" \
    go test -timeout="${TIMEOUT}" "$@"
fi

EXIT_CODE=$?

if [ $EXIT_CODE -eq 124 ]; then
    echo -e "${RED}Tests timed out after ${TIMEOUT}${NC}"
    exit 124
elif [ $EXIT_CODE -eq 137 ]; then
    echo -e "${RED}Tests were killed (likely due to memory limit)${NC}"
    exit 137
elif [ $EXIT_CODE -ne 0 ]; then
    echo -e "${RED}Tests failed with exit code $EXIT_CODE${NC}"
    exit $EXIT_CODE
else
    echo -e "${GREEN}All tests passed successfully!${NC}"
fi