# Testing Guidelines for Nexora

## Problem Description

The QA suite test in `qa/qa_suite_test.go` was running `go test ./...` which created a recursive loop:

1. Running `go test ./...` would find and run the QA tests
2. The QA tests would run `go test ./...` again
3. This created an infinite recursion of test runs
4. Each test invocation spawned multiple parallel sub-tests using `t.Parallel()`
5. Leading to exponential growth in processes

## Solution Implemented

Added a environment variable check to prevent nested test runs:

```go
// Check if we're already in a nested test run to avoid infinite recursion
if os.Getenv("NESTED_TEST_RUN") == "1" {
    t.Skip("skipping nested test run to avoid infinite recursion")
}

// Set environment variable to detect nested test runs
cmd := exec.Command("sh", "-c", "NESTED_TEST_RUN=1 go test ./... -timeout=5m 2>&1 || true")
```

## Recommendations

1. **Avoid recursive test calls**: Test suites should not invoke `go test` on themselves
2. **Use proper test dependencies**: If testing multiple packages, explicitly list them
3. **Consider separate test commands**: Use Makefile targets for different test levels
4. **Set appropriate timeouts**: Always use timeout flags to prevent hanging tests

## Improved Testing Structure

Consider these alternatives instead of recursive `go test`:

```go
// Option 1: Test specific packages explicitly
packages := []string{
    "./internal/agent/...",
    "./internal/config/...",
    "./internal/...",
    "./cmd/...",
}

for _, pkg := range packages {
    cmd := exec.Command("go", "test", pkg, "-timeout=5m")
    // ... run command
}

// Option 2: Use a separate integration test script
// that doesn't run under `go test ./...`

// Option 3: Skip the recursive test entirely and rely on CI/CD
// or Makefile targets for full test runs
```

## Monitoring Test Processes

To check for similar issues in the future:

```bash
# Count running test processes
ps aux | grep -E "go test" | grep -v grep | wc -l

# Check parent-child relationships
ps -eo pid,ppid,cmd | grep -E "go test|sh -c" | grep -v grep

# Kill stuck test processes
pkill -f "go test"
```

## Test Timeout Best Practices

1. Always set explicit timeouts
2. Use the `run-tests-with-limits.sh` script for additional protection
3. Configure appropriate memory limits
4. Monitor resource usage during test runs