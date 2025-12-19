# Test Timeouts and Resource Limits

This project includes comprehensive timeout and resource limiting features to prevent Go tests from running forever and hogging CPU/RAM resources.

## Overview

The test infrastructure includes multiple layers of timeout protection:

1. **Go built-in timeout** - Applied to all test commands
2. **System-level timeout** - Forces test termination if Go timeout fails
3. **Memory limits** - Prevents tests from consuming unlimited RAM
4. **CPU limits** - Controls CPU usage during test execution
5. **Test helper functions** - For individual test timeout control

## Using Timeouts

### From Makefile

```bash
# Regular tests with 10-minute timeout
make test

# QA tests with 5-minute timeout
make test-qa

# Quick tests with 5-minute timeout
make test-quick

# Tests with resource limits (memory/CPU)
make test-limited
```

### From Taskfile

```bash
# Regular tests with timeout
task test

# Tests with resource limits
task test:limited

# QA tests specifically
task test:qa
```

### Direct Go Test with Timeout

```bash
# Individual packages with timeout
go test -timeout=5m ./internal/config/...

# All tests with custom timeout
go test ./... -timeout=15m
```

## Environment Variables

Customize timeouts without changing code:

```bash
# Export before running tests
export NEXORA_TEST_TIMEOUT_SHORT=1m
export NEXORA_TEST_TIMEOUT_MEDIUM=10m
export NEXORA_TEST_TIMEOUT_LONG=30m
export NEXORA_TEST_TIMEOUT_QA=7m

# Resource limit customization
export TEST_TIMEOUT=15m
export TEST_MEMORY_LIMIT=4G
export TEST_CPU_LIMIT=4
export TEST_TIMEOUT_KILL_AFTER=1m

# Then run tests
make test
```

## Test Helper Functions

For individual tests that might hang:

```go
import "github.com/nexora/nexora/internal/testutil"

func TestSlowOperation(t *testing.T) {
    testutil.RunWithTimeout(t, 30*time.Second, func(ctx *testing.T) {
        result := slowOperation(ctx.Context())
        if result != expected {
            t.Error("unexpected result")
        }
    })
}

func TestWithDeadline(t *testing.T) {
    deadline := time.Now().Add(5 * time.Minute)
    testutil.RunWithTimeoutAndDeadline(t, 2*time.Minute, deadline, func(ctx *testing.T) {
        // Must complete before 2min timeout and 5min deadline
        performOperation(ctx.Context())
    })
}
```

## Resource Limits

The `run-tests-with-limits.sh` script provides OS-level resource limiting:

- **Memory limits** - Prevents tests from consuming unlimited RAM
- **Timeout enforcement** - Kills test processes after timeout
- **CPU throttling** - Optional CPU usage limits

### Requirements for Full Feature Set

- **Linux**: Uses `prlimit` or `ulimit` for memory limits
- **macOS**: Uses `gtimeout` from GNU coreutils (install via brew)
- **Windows**: Built-in Go timeout only (install WSL for full features)

### Installation

```bash
# On macOS with Homebrew
brew install coreutils bc

# On Ubuntu/Debian
sudo apt-get install coreutils bc

# On other Linux distributions
# coreutils should already be installed
```

## CI/CD Integration

The GitHub Actions workflow already includes timeout protection:

```yaml
- name: Run tests
  run: go test -v -race -timeout=10m -coverprofile=coverage.out $(go list ./... | grep -v '/qa$')
```

This ensures CI tests won't run forever, consuming resources.

## Troubleshooting

### Timeouts Too Aggressive

If tests are timing out prematurely:

1. Increase timeout: `go test -timeout=30m`
2. Set environment variable: `export TEST_TIMEOUT=30m`
3. Check for actual performance issues

### Memory Limits Too Restrictive

If tests fail due to memory limits:

1. Increase limit: `export TEST_MEMORY_LIMIT=8G`
2. Run without memory limits: `make test` (instead of `make test-limited`)

### Test Still Hanging

If a test bypasses all timeout mechanisms:

1. Find the process: `ps aux | grep "go test"`
2. Kill manually: `kill -9 <pid>`
3. Report the issue - this indicates a serious bug

## Best Practices

1. **Always use timeouts** - Never run tests without `-timeout` flag
2. **Set разумные limits** - Match timeouts to expected test duration
3. **Test in CI** - Ensure resource limits work in your CI environment
4. **Monitor resources** - Check memory/CPU usage during test execution
5. **Use test helpers** - Wrap potentially slow operations in timeout helpers

## Examples

### Running Integration Tests with Extended Timeouts

```bash
# For slow integration tests
go test ./integration/... -timeout=30m -v

# With resource limits
TEST_TIMEOUT=30m TEST_MEMORY_LIMIT=4G make test-limited
```

### Debugging Timeouts

```bash
# Run with verbose output to see which test times out
go test -v -timeout=5m ./...

# Run specific test
go test -run TestSpecificSlowFunction -timeout=10m ./pkg/...
```