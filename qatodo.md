# Nexora QA Testing TODO

> Generated: December 18, 2025
> Focus: Above-provider-level testing (provider coverage handled by modelscan)

## 📊 Current Coverage Summary

### Critical Gaps (0% coverage)
| Package | Purpose | LOC | Priority |
|---------|---------|-----|----------|
| `internal/session` | Session lifecycle CRUD | ~200 | 🔴 HIGH |
| `internal/db` | SQLite persistence layer | ~300 | 🔴 HIGH |
| `internal/message` | Message content types | ~400 | 🔴 HIGH |
| `internal/agent/native` | Native tool dispatch | ~150 | 🔴 HIGH |
| `internal/agent/tools/mcp` | MCP server integration | ~300 | 🟡 MEDIUM |
| `internal/app` | App bootstrap/lifecycle | ~200 | 🟡 MEDIUM |
| `internal/history` | Command history | ~100 | 🟢 LOW |
| `internal/format` | Output formatting | ~100 | 🟢 LOW |
| `internal/sessionlog` | Session logging | ~150 | 🟢 LOW |

### Low Coverage (needs improvement)
| Package | Current | Target | Notes |
|---------|---------|--------|-------|
| `internal/agent` | 0.7% | 50%+ | Core orchestration logic |
| `internal/agent/tools` | 12.2% | 60%+ | Tool implementations |
| `internal/lsp` | 16.1% | 40%+ | LSP integration |
| `internal/cmd` | 29.1% | 50%+ | CLI commands |
| `internal/config` | 49.7% | 70%+ | Config loading |

### Good Coverage (maintain)
| Package | Coverage |
|---------|----------|
| `internal/ansiext` | 100% |
| `internal/diff` | 100% |
| `internal/oauth` | 100% |
| `internal/stringext` | 100% |
| `internal/term` | 100% |
| `internal/pubsub` | 97.8% |
| `internal/csync` | 95.9% |
| `internal/resources` | 93.3% |

---

## 🎯 Testing Layers

### Layer 1: Unit Tests (per-package)
Individual package testing with mocked dependencies.

**TODO:**
- [ ] `internal/session/service_test.go` - Mock db.Querier, test CRUD
- [ ] `internal/db/db_test.go` - In-memory SQLite, test migrations
- [ ] `internal/message/message_test.go` - Content part serialization
- [ ] `internal/message/service_test.go` - Mock db, test CRUD + pubsub
- [ ] `internal/agent/tools/bash_test.go` - Command execution, timeouts
- [ ] `internal/agent/tools/fetch_test.go` - HTTP mocking
- [ ] `internal/agent/native/native_test.go` - Tool dispatch logic

### Layer 2: Integration Tests (cross-package)
Test component interactions without real providers.

**TODO:**
- [ ] `qa/agent_flow_test.go` - Agent → Tool → Response cycle
- [ ] `qa/session_lifecycle_test.go` - Create → Use → Persist → Resume
- [ ] `qa/message_flow_test.go` - User input → Message → Storage
- [ ] `qa/state_machine_test.go` - State transitions under various conditions
- [ ] `qa/recovery_test.go` - Error injection at each layer

### Layer 3: E2E Tests (full system)
Complete flows with mock provider responses.

**TODO:**
- [ ] `qa/e2e_conversation_test.go` - Multi-turn conversation
- [ ] `qa/e2e_tool_execution_test.go` - Tool calls with filesystem
- [ ] `qa/e2e_session_resume_test.go` - Session persistence across restarts
- [ ] `qa/e2e_error_recovery_test.go` - Recovery from various failure modes

---

## 🔧 Infrastructure Needed

### Mock Provider Framework
```go
// qa/testutil/mock_provider.go
type MockProvider struct {
    Responses []fantasy.Response  // Pre-canned responses
    Calls     []fantasy.Request   // Captured requests
    Errors    []error             // Errors to inject
}
```

### Test Fixtures
```
qa/testdata/
├── sessions/           # Sample session JSON
├── messages/           # Sample message content
├── tools/              # Tool input/output fixtures
└── configs/            # Test configurations
```

### Test Helpers
- [ ] `qa/testutil/db.go` - In-memory SQLite factory
- [ ] `qa/testutil/session.go` - Session factory helpers
- [ ] `qa/testutil/message.go` - Message builders
- [ ] `qa/testutil/agent.go` - Agent with mock provider

---

## ❓ Questions for User

1. **Session persistence priority**: Should session tests focus on SQLite directly, or through the service layer? (affects mock complexity)

2. **Tool testing scope**: For tools like `bash` and `edit`, should tests use real filesystem in temp dirs, or mock filesystem interfaces?

3. **E2E test environment**: Should E2E tests run in isolated Docker containers, or is temp directory isolation sufficient?

4. **Coverage targets**: What's the minimum acceptable coverage for a "release-ready" state? (e.g., 60% overall, 80% for critical paths)

5. **CI integration**: Should QA tests run on every commit, or only on PR/release branches? (affects test runtime budget)

---

## 💡 Suggestions

### 1. Start with Session/Message/DB triangle
These three packages form the persistence core. Testing them together provides:
- Foundation for all other tests (sessions needed everywhere)
- High impact (0% → 70%+ coverage for critical code)
- Low complexity (mostly CRUD operations)

**Estimated effort**: 4-6 hours

### 2. Create MockProvider in qa/testutil
A reusable mock provider that:
- Returns pre-defined responses
- Captures requests for assertions
- Injects errors on demand
- Simulates streaming responses

This unblocks ALL integration/E2E tests without touching real providers.

**Estimated effort**: 2-3 hours

### 3. Add table-driven tool tests
Tools like `edit`, `grep`, `glob` have deterministic behavior. Table-driven tests with fixtures:
```go
tests := []struct{
    name     string
    input    EditInput
    files    map[string]string  // initial state
    expected map[string]string  // final state
    wantErr  bool
}{...}
```

**Estimated effort**: 6-8 hours for all tools

### 4. Implement error injection framework
For recovery testing, add injection points:
```go
type ErrorInjector struct {
    FailAt    string  // "db.Create", "tool.Execute", etc.
    FailAfter int     // Fail after N successful calls
    Error     error
}
```

This enables systematic testing of all recovery paths.

**Estimated effort**: 3-4 hours

### 5. Add golden file tests for agent responses
Capture "known good" agent responses and compare:
```go
func TestAgentResponse_GoldenFile(t *testing.T) {
    got := runAgent(input)
    golden.Assert(t, got, "testdata/agent_response.golden")
}
```

Run with `-update` flag to refresh goldens. Catches regressions in response formatting.

**Estimated effort**: 2-3 hours

---

## 📅 Suggested Phases

### Phase 1: Foundation (1-2 days)
- [ ] Session/Message/DB unit tests
- [ ] MockProvider framework
- [ ] Test fixtures directory

### Phase 2: Tools (2-3 days)
- [ ] Table-driven tool tests
- [ ] Filesystem test helpers
- [ ] Tool timeout tests

### Phase 3: Integration (2-3 days)
- [ ] Agent flow tests
- [ ] State machine integration
- [ ] Error injection framework

### Phase 4: E2E (1-2 days)
- [ ] Full conversation tests
- [ ] Session resume tests
- [ ] Recovery scenario tests

**Total estimated effort**: 6-10 days for comprehensive coverage

---

## 📝 Notes

- TUI packages excluded (UI testing is a separate concern)
- Provider-level testing deferred to modelscan project
- Focus on deterministic, fast tests (no network, no real providers)
- All tests should pass with `-race` flag
