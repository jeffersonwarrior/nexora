## Testing Infrastructure Status and Plan

### Current State Analysis
- **Overall Coverage**: 12.0% (need: 80%)
- **Major blockers**: QA test utilities broken, missing fantasy API compatibility
- **Test Strategy**: Simple mocks + temp directories + targeted coverage

### Priority Packages (by impact + current coverage)

#### HIGH PRIORITY (Immediate wins >10% coverage boost each)
1. **agent/tools** - 15.4% coverage, 30+ tools need basic tests
2. **db** - 32.0% coverage, database operations need test coverage  
3. **cmd** - 29.1% coverage, CLI commands need tests
4. **message** - 40.3% coverage, message handling needs tests
5. **lsp** - 16.1% coverage, language server protocol needs tests

#### MEDIUM PRIORITY (5-10% coverage boost each)
6. **config/providers** - 28.3% coverage, provider configs need tests
7. **config** - 49.7% coverage, configuration logic needs tests
8. **tui/components/dialogs/models** - 21.5% coverage, UI models need tests
9. **indexer** - 42.3% coverage, file indexing needs tests
10. **update** - 48.5% coverage, update logic needs tests

#### ZERO COVERAGE PACKAGES (Quick wins with basic tests)
11. All TUI components (20+ packages with 0% coverage)
12. Session log, task, app, history, format utilities
13. OAuth variants, LSP utilities, MCP tools

### Implementation Strategy

#### Phase 1: Foundation (Week 1)
1. **Fix QA utilities** - Remove broken mock_provider, create simple test helpers
2. **Update Makefile** - Add coverage targets, test runners
3. **Create test templates** - Standardized test patterns for common scenarios
4. **Setup temporary directories** - Reusable temp dir management

#### Phase 2: High Impact (Week 2-3) 
1. **agent/tools** - Create comprehensive tests for 30+ tools
2. **database operations** - Test CRUD, migrations, edge cases
3. **CLI commands** - Test all command scenarios
4. **message handling** - Test parsing, validation, transformation

#### Phase 3: Coverage Completion (Week 4)
1. **TUI components** - Basic interaction tests
2. **Zero coverage packages** - Create initial test files
3. **Integration scenarios** - Cross-package workflows
4. **Performance tests** - Critical path validation

### Testing Standards

#### Simple Mock Pattern
```go
func TestXxx(t *testing.T) {
    t.Run("success", func(t *testing.T) {
        // Arrange - setup temp dir, simple mock
        // Act - call function  
        // Assert - verify result
    })
    t.Run("error case", func(t *testing.T) {
        // Arrange - setup error condition
        // Act - call function
        // Assert - verify error handling
    })
}
```

#### Temporary Directory Pattern
```go
func setupTestDir(t *testing.T) string {
    t.Helper()
    dir := t.TempDir()
    // Create test files as needed
    return dir
}
```

#### Coverage Target Strategy
- **Each function**: at least basic success/failure test
- **Error paths**: must be tested 
- **Edge cases**: null/empty values, boundary conditions
- **Public APIs**: comprehensive testing
- **Private functions**: tested indirectly via public calls

### Success Metrics 
- **Week 1**: 25% coverage (fix foundation + quick wins)
- **Week 2**: 50% coverage (high impact packages)
- **Week 3**: 70% coverage (complete critical packages)
- **Week 4**: 80%+ coverage (final gaps + integration)

### Automated QA Process
1. **Pre-commit**: `make test-quick` (unit tests only)
2. **CI/CD**: `make test-full` (unit + integration)
3. **Coverage monitoring**: Automated reports, PR gates
4. **Quality checks**: `make qa` (lint + test validation)

This plan achieves the 80% coverage target with simple, maintainable tests using temporary directories and basic mocks.