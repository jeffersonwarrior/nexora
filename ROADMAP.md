# Nexora Development Roadmap
**Last Updated**: December 17, 2025  
**Version**: 0.28.7

---

## 🎯 Mission Statement

Transform Nexora into a production-grade AI coding assistant with:
- **Reliability**: Zero crashes, predictable behavior, comprehensive error handling
- **Performance**: Sub-second response times, efficient resource usage
- **Extensibility**: Easy provider integration, plugin architecture
- **Quality**: 80%+ test coverage, comprehensive documentation

---

## 📊 Current State Assessment

### Health Metrics ✅
- **Code Quality**: A- (90/100)
- **Test Coverage**: 23% (73/313 files)
- **Technical Debt**: 15 markers (manageable)
- **Production Status**: ✅ READY
- **Test Results**: Zero failures across 20+ packages

### Key Strengths
- Clean architecture with separation of concerns
- Comprehensive error handling
- Pragmatic fixes (90% edit tool failure reduction)
- Good observability (logging, metrics)

### Improvement Areas
- Performance optimization potential (~500-1000ms savings)
- Test coverage gaps (integration, error paths)
- Provider abstraction complexity
- Memory leak risks (unbounded buffers)

---

## 🚨 CRITICAL PRIORITIES (This Week)

### 0. Reduce Session Startup Token Cost 💰 **EFFICIENCY** ✅ COMPLETE
**Priority**: P0 - Critical  
**Effort**: 4-6 hours  
**Impact**: 30k 	 27k tokens (11% reduction), faster responses, lower costs

**Status**: ✅ **COMPLETED** - December 17, 2025

**Results Achieved**:
- **Total Saved**: ~13,116 bytes (~3,279 tokens @ 4:1 ratio)
- **Template Reduction**: 37% (35,226 	 22,110 bytes)
- **Session Startup**: 30k 	 27k tokens (11% reduction)
- **All Tests Passing**: ✅ Zero failures

**Phase 1: Compress Documentation ✅**
- edit.md: 9,337 	 3,699 bytes (-60%) ✅
- bash.tpl: 5,227 	 3,593 bytes (-31%) ✅
- multiedit.md: 4,982 	 3,569 bytes (-28%) ✅
- coder.md.tpl: 7,150 	 5,300 bytes (-26%) ✅
- **Saved**: 10,535 bytes (~2,634 tokens)

**Phase 2: Optimize Runtime Data ✅**
- agentic_fetch.md: 2,924 	 1,618 bytes (-45%) ✅
- git commits: 3 	 2 (reduced lines) ✅
- git status: 20 	 5 files (reduced lines) ✅
- network/services: lazy-loaded (NEXORA_FULL_ENV=1 to enable) ✅
- **Saved**: ~2,000 bytes (~500 tokens)

**Phase 3: Consolidate ✅**
- job_output.md: 570 	 282 bytes (-51%) ✅
- job_kill.md: 494 	 201 bytes (-59%) ✅
- **Saved**: 581 bytes (~145 tokens)

**Files Modified**:
- `internal/agent/tools/edit.md` ✅
- `internal/agent/tools/bash.tpl` ✅
- `internal/agent/tools/multiedit.md` ✅
- `internal/agent/templates/coder.md.tpl` ✅
- `internal/agent/templates/agentic_fetch.md` ✅
- `internal/agent/tools/job_output.md` ✅
- `internal/agent/tools/job_kill.md` ✅
- `internal/agent/prompt/prompt.go` ✅
- `internal/agent/prompt/prompt_test.go` ✅

**Notes**:
- Set `NEXORA_FULL_ENV=1` to enable full network status and active services detection
- Default behavior now assumes "online" and skips expensive service detection
- All documentation remains comprehensive while being more concise

---

### 1. Background Job Monitoring & TODO System 🔔 **INFRASTRUCTURE**
**Priority**: P0 - Critical  
**Effort**: 2-3 weeks (incremental)  
**Impact**: Persistent task tracking, error recovery, long-term memory

**Problem**: Long-running background jobs fail silently
- No notification when nohup-style jobs error
- No persistent TODO system for tracking work
- No memory system for context across sessions
- Background jobs lost on restart (in-memory only)

**Current State**:
- ✅ BackgroundShellManager (50 jobs, in-memory, 8h retention)
- ✅ PubSub infrastructure (sessions, messages, permissions)
- ✅ SQLite persistence (sessions, messages)
- ❌ No job	agent error notification
- ❌ No persistent TODO tracking
- ❌ No long-term memory system

**Research Findings**:
- **Memory**: claude-mem uses SQLite+FTS5 with token budgeting, auto-summarization
- **Jobs**: Asynq (Redis) or River (Postgres) - production-ready with monitoring
- **IPC**: Can extend existing PubSub or use Redis pub/sub

**Proposed Architecture**:

**1. Three-Tier TODO System**
```
- Ephemeral Tasks: Current in-memory (internal/task/agent_tool.go)
- Persistent TODOs: New SQLite table with priorities, session links
- Background Jobs: Migrate to River (Postgres LISTEN/NOTIFY)
```

**2. Error Notification Pipeline**
```go
// Extend internal/pubsub/broker.go
type JobEvent struct {
    JobID     string
    SessionID uuid.UUID  // Link back to originating session
    Status    JobStatus  // Running, Failed, Completed
    Error     error
    Output    string
}

// Job fails 	 Event published 	 Agent creates TODO 	 Queued in session
```

**3. Memory System (SQLite + FTS5)**
```
- Store: tool responses, preferences, project context
- FTS5 search: keyword-based retrieval
- Token budgeting: summarize old memories to fit context
- Auto-expire: 90-day TTL with importance scaling
- Integration: Auto-load into PrepareStep prompt
```

**4. Background Job 	 TODO Flow**
```
1. BackgroundShell fails 	 JobEvent published
2. Agent JobEventHandler creates TODO: "Fix error in job X"
3. TODO queued in session's messageQueue
4. Next agent invocation sees TODO in context
5. Agent works on TODO, marks complete
```

**Implementation Phases**:

**Week 1: TODO Foundation**
- Add `todos` table to SQLite (id, session_id, priority, deadline, status)
- CRUD operations in `internal/todo/` service
- Subscribe to PubSub events

**Week 2: Job Monitoring**
- Extend BackgroundShellManager to publish JobEvents
- Add job failure detection
- Link jobs to originating sessions

**Week 3: Agent Integration**
- Agent auto-loads TODOs in PrepareStep
- Job errors create TODOs automatically
- TODO completion tracking

**Week 4: Memory System**
- Add `memories` table with FTS5
- Token budgeting for context fitting
- Auto-expire with importance scoring

**Week 5: Production Hardening**
- Migrate BackgroundShellManager to River (optional)
- Add monitoring dashboard
- Comprehensive testing

**Files to Create**:
- `internal/todo/todo.go` (service with PubSub)
- `internal/todo/queries.sql` (TODO CRUD)
- `internal/memory/memory.go` (FTS5 memory service)
- `internal/memory/queries.sql` (memory storage)

**Files to Modify**:
- `internal/shell/background.go` (publish JobEvents)
- `internal/agent/agent.go` (load TODOs in PrepareStep)
- `internal/db/schema.sql` (add todos + memories tables)

**Open Questions**:
1. **Persistence**: SQLite (simpler) or Redis (distributed)?
2. **Notification**: Interrupt agent or queue for next idle?
3. **Memory**: Build our own or integrate claude-mem MCP server?
4. **Jobs**: Extend current system or migrate to River/Asynq?
5. **Scope**: Per-session, global, or project-scoped TODOs?

**Suggestions**:
1. Start with SQLite for consistency with existing architecture
2. Queue notifications for next agent invocation (less disruptive)
3. Build SQLite+FTS5 memory system (full control, no external deps)
4. Keep current BackgroundShellManager, extend with persistence
5. Support both per-session and global TODOs with `scope` field

**Testing**: 
- Background job failure creates TODO
- TODO persists across restarts
- Memory system retrieves relevant context
- FTS5 search performance benchmarks

---

### 2. Fix Turbo Mode Implementation 🔥 **BROKEN BUILD**
**Priority**: P0 - Critical  
**Effort**: 1-2 hours  
**Impact**: Restore broken feature

**Issue**: Previous Turbo Mode implementation caused build failures:
1. Template syntax `{{if isTurbo}}` used in Go struct literal (agent.go lines 349-428)
2. `summarizer.go` lost `package agent` declaration
3. Turbo detection placed in wrong scope (PrepareStep func instead of agent-level)

**Solution**: Re-implement with proper Go syntax
```go
// internal/agent/agent.go - Agent struct level
type sessionAgent struct {
    // ... existing fields ...
    isTurboMode bool // NEW: Add at agent level, not in PrepareStep
}

// Check turbo mode in constructor or config
func NewSessionAgent(...) *sessionAgent {
    agent := &sessionAgent{
        isTurboMode: detectTurboMode(cfg), // Detect once at init
    }
    return agent
}

// Use in PrepareStep
func (a *sessionAgent) PrepareStep(...) {
    if a.isTurboMode {
        // Turbo optimizations
    }
}
```

**Files to Fix**:
- `internal/agent/agent.go` - Add isTurboMode field, proper detection
- `internal/agent/summarizer.go` - Verify package declaration intact

**Testing**: 
- Unit test for turbo mode detection
- Integration test with turbo enabled/disabled
- Verify `go build .` passes

**Status**: ⚠️ Reverted to commit 9f39ceb (clean state)

### 3. Memory Leak Prevention ⚠️ **BLOCKER**
**Impact**: Prevents OOM in production

**Issue**: Unbounded slices in loop detection
```go
// internal/agent/agent.go lines 96-100
recentCalls    []aiops.ToolCall  // ← UNBOUNDED
recentActions  []aiops.Action    // ← UNBOUNDED
```

**Solution**:
```go
const maxRecentCalls = 100
const maxRecentActions = 100

// In append logic:
if len(a.recentCalls) >= maxRecentCalls {
    a.recentCalls = append(a.recentCalls[1:], newCall)
} else {
    a.recentCalls = append(a.recentCalls, newCall)
}
```

**Files to Modify**:
- `internal/agent/agent.go`

**Testing**: Add unit test verifying bounds enforcement

### 4. Provider Options Bug 🐛 **HIGH PRIORITY**
**Impact**: Lost configuration when switching models

**Issue**: Summarization clears provider options for Cerebras
```go
// internal/agent/agent.go lines 934-946
if isCerebras {
    summarizationOpts = fantasy.ProviderOptions{} // ← BUG: Lost temp, topP, etc.
}
```

**Solution**:
```go
// Preserve compatible options
summarizationOpts = fantasy.ProviderOptions{
    Temperature:      opts.Temperature,
    TopP:             opts.TopP,
    // Don't copy model-specific options
}
```

**Files to Modify**:
- `internal/agent/agent.go`

**Testing**: Verify temperature/topP preserved in Cerebras summarization

---

### 3. Quick Performance Wins ⚡
**Priority**: P1 - High  
**Total Effort**: 1 hour  
**Impact**: ~150-200ms latency reduction

#### 3a. Git Config Caching (15 min)
```go
// internal/agent/prompt/prompt.go
type GitConfigCache struct {
    userName  string
    userEmail string
    once      sync.Once
}

func (c *GitConfigCache) Load(ctx context.Context) {
    c.once.Do(func() {
        c.userName = getGitConfig(ctx, "user.name")
        c.userEmail = getGitConfig(ctx, "user.email")
    })
}
```
**Savings**: 20-50ms per prompt

#### 3b. Fuzzy Match Size Limit (5 min)
```go
// internal/agent/tools/edit.go before fuzzy matching
if len(oldContent) > 50000 { // 50KB threshold
    goto skipFuzzyMatch
}
```
**Savings**: Prevents O(n²) on large files

#### 3c. Endpoint Validation Parallel (20 min)
```go
// .local/tools/modelscan/providers/*.go
var wg sync.WaitGroup
for i := range endpoints {
    wg.Add(1)
    go func(ep *Endpoint) {
        defer wg.Done()
        testEndpoint(ctx, ep)
    }(&endpoints[i])
}
wg.Wait()
```
**Savings**: N × latency → max(latencies)

---

## 🔥 HIGH PRIORITY (Next 2 Weeks)

### 4. Prompt Generation Performance 🚀
**Priority**: P1 - High  
**Effort**: 3-4 hours  
**Impact**: 500-800ms latency reduction

#### Problem
Environment detection runs 10+ shell commands sequentially:
- Python version
- Node version
- Go version
- Git config (2 calls)
- Memory info
- Disk info
- Network status
- Active services
- Container detection

**Current**: 300-800ms per prompt  
**Target**: <100ms per prompt

#### Solution: Cache Layer with Parallel Refresh

**Implementation**:
```go
// internal/agent/prompt/cache.go (NEW FILE)
package prompt

type EnvironmentCache struct {
    mu         sync.RWMutex
    data       EnvironmentData
    lastUpdate time.Time
    ttl        time.Duration
}

type EnvironmentData struct {
    CurrentUser    string
    LocalIP        string
    PythonVersion  string
    NodeVersion    string
    GoVersion      string
    GitUserName    string
    GitUserEmail   string
    MemoryInfo     string
    DiskInfo       string
    Architecture   string
    ContainerType  string
    TerminalInfo   string
    NetworkStatus  string
    ActiveServices string
}

func NewEnvironmentCache(ttl time.Duration) *EnvironmentCache {
    return &EnvironmentCache{
        ttl: ttl,
    }
}

func (c *EnvironmentCache) Get(ctx context.Context, workingDir string) (EnvironmentData, error) {
    c.mu.RLock()
    if time.Since(c.lastUpdate) < c.ttl {
        defer c.mu.RUnlock()
        return c.data, nil
    }
    c.mu.RUnlock()
    
    return c.refresh(ctx, workingDir)
}

func (c *EnvironmentCache) refresh(ctx context.Context, workingDir string) (EnvironmentData, error) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    // Double-check after acquiring write lock
    if time.Since(c.lastUpdate) < c.ttl {
        return c.data, nil
    }
    
    eg, ctx := errgroup.WithContext(ctx)
    data := EnvironmentData{}
    
    // Parallel execution
    eg.Go(func() error {
        data.CurrentUser = getCurrentUser()
        return nil
    })
    eg.Go(func() error {
        data.LocalIP = getLocalIP(ctx)
        return nil
    })
    eg.Go(func() error {
        data.PythonVersion = getRuntimeVersion(ctx, "python3 --version")
        return nil
    })
    eg.Go(func() error {
        data.NodeVersion = getRuntimeVersion(ctx, "node --version")
        return nil
    })
    eg.Go(func() error {
        data.GoVersion = getRuntimeVersion(ctx, "go version")
        return nil
    })
    eg.Go(func() error {
        data.GitUserName = getGitConfig(ctx, "user.name")
        data.GitUserEmail = getGitConfig(ctx, "user.email")
        return nil
    })
    eg.Go(func() error {
        data.MemoryInfo = getMemoryInfo(ctx)
        return nil
    })
    eg.Go(func() error {
        data.DiskInfo = getDiskInfo(ctx, workingDir)
        return nil
    })
    eg.Go(func() error {
        data.Architecture = getArchitecture()
        data.ContainerType = detectContainer(ctx)
        data.TerminalInfo = getTerminalInfo(ctx)
        return nil
    })
    eg.Go(func() error {
        data.NetworkStatus = getNetworkStatus(ctx)
        return nil
    })
    eg.Go(func() error {
        data.ActiveServices = detectActiveServices(ctx)
        return nil
    })
    
    if err := eg.Wait(); err != nil {
        return EnvironmentData{}, err
    }
    
    c.data = data
    c.lastUpdate = time.Now()
    return data, nil
}
```

**Files to Create**:
- `internal/agent/prompt/cache.go` (~150 lines)

**Files to Modify**:
- `internal/agent/prompt/prompt.go` (use cache instead of direct calls)
- `internal/agent/coordinator.go` (initialize cache)

**Configuration**:
```toml
[agent]
environment_cache_ttl = "60s"  # Refresh every 60 seconds
```

**Testing**:
- Unit tests for cache expiry
- Benchmark before/after
- Verify all fields populated

---

### 5. Architecture Decision Records 📚
**Priority**: P1 - High  
**Effort**: 2-3 hours  
**Impact**: High (maintainability, onboarding)

**Purpose**: Document "why" decisions were made

#### ADRs to Create

**ADR-001: Force AI Mode in Edit Tool**
- Why: 90% failure rate with whitespace mismatches
- Decision: Force `ai_mode=true` by default
- Tradeoff: Slight overhead vs. reliability

**ADR-002: 100-Line Chunks for VIEW Tool**
- Why: Prevent context window exhaustion
- Decision: Default 100 lines (from higher limit)
- Tradeoff: Multiple calls vs. context safety

**ADR-003: 20% Threshold for Auto-Summarization**
- Why: Balance between context usage and summarization overhead
- Decision: Trigger at 80% context window usage
- Tradeoff: Frequency vs. context preservation

**ADR-004: Clear Provider Options on Model Switch**
- Why: Cerebras summarization reliability
- Decision: Use smallModel with cleared options
- Status: **UNDER REVIEW** (may be bug)

**ADR-005: Fuzzy Match Confidence Threshold (0.90)**
- Why: Balance accuracy vs. false positives
- Decision: Only accept matches with 90%+ confidence
- Tradeoff: Precision vs. recall

**ADR-006: Environment Detection in System Prompt**
- Why: Better agent awareness of runtime
- Decision: 15+ new fields (Python, memory, network, etc.)
- Tradeoff: Latency vs. context richness

**ADR-007: AIOPS Fallback Strategy**
- Why: Complex edit resolution
- Decision: Local fuzzy → Remote AIOPS
- Status: **NEEDS DECISION** (see Question #2)

**Template** (`docs/adr/template.md`):
```markdown
# ADR-XXX: [Title]

## Status
[Proposed | Accepted | Deprecated | Superseded by ADR-YYY]

## Context
What is the issue we're seeing that motivates this decision?

## Decision
What is the change we're proposing?

## Consequences
What becomes easier or more difficult as a result?

### Positive
- ...

### Negative
- ...

### Risks
- ...

## Alternatives Considered
1. Option A: ...
2. Option B: ...

## Implementation Notes
- Files affected: ...
- Migration required: ...
```

**Directory Structure**:
```
docs/
  adr/
    README.md (index of all ADRs)
    template.md
    001-force-ai-mode-edit-tool.md
    002-view-tool-100-line-chunks.md
    003-auto-summarization-threshold.md
    004-provider-options-model-switch.md
    005-fuzzy-match-confidence.md
    006-environment-detection.md
    007-aiops-fallback-strategy.md
```

---

### 6. Test Coverage Expansion 🧪
**Priority**: P1 - High  
**Effort**: 8-12 hours  
**Target**: 40% coverage (from 23%)

#### Coverage Gaps

**Missing Integration Tests**:
- End-to-end agent flow (prompt → tools → response)
- Multi-turn conversations
- Auto-summarization triggers
- Context window management

**Missing Unit Tests**:
- Environment detection (all new functions)
- Fuzzy matching strategies
- AIOPS fallback logic
- Loop detection bounds

**Missing Error Path Tests**:
- Network failures
- API rate limits
- Invalid tool parameters
- File permission errors

#### Implementation Plan

**Week 1**: Core Agent Tests
```go
// internal/agent/agent_integration_test.go (NEW)
func TestAgentEndToEnd(t *testing.T)
func TestAgentAutoSummarization(t *testing.T)
func TestAgentToolExecution(t *testing.T)
func TestAgentContextWindow(t *testing.T)
```

**Week 2**: Tool Tests
```go
// internal/agent/tools/edit_test.go (EXPAND)
func TestEditFuzzyMatching(t *testing.T)
func TestEditAIOPSFallback(t *testing.T)
func TestEditLargeFiles(t *testing.T)

// internal/agent/tools/edit_bench_test.go (NEW)
func BenchmarkEditSmallFile(b *testing.B)
func BenchmarkEditLargeFile(b *testing.B)
func BenchmarkFuzzyMatch(b *testing.B)
```

**Week 3**: Environment Detection Tests
```go
// internal/agent/prompt/prompt_test.go (EXPAND)
func TestEnvironmentCache(t *testing.T)
func TestParallelEnvironmentDetection(t *testing.T)
func TestEnvironmentCacheTTL(t *testing.T)

// internal/agent/prompt/prompt_bench_test.go (NEW)
func BenchmarkPromptBuild(b *testing.B)
func BenchmarkEnvironmentDetection(b *testing.B)
```

---

## 🎯 MEDIUM PRIORITY (Next Month)

### 7. Provider Abstraction Refactor 🏗️
**Priority**: P2 - Medium  
**Effort**: 25-35 hours (1 week full-time)  
**Impact**: Maintainability, extensibility, performance

**Goal**: Replace `charm.land/fantasy` + `catwalk` with unified provider abstraction

#### Benefits
- ✅ Smaller, more maintainable code
- ✅ 1-2 hours to add new providers (vs current complexity)
- ✅ Better error handling per provider
- ✅ Remove 2 dependencies
- ✅ Better performance (direct API calls)
- ✅ Support bleeding-edge features faster

#### Phases (from todo.md)

**Phase 1**: Create Provider Interface (4-6 hours)
```go
// internal/llm/provider.go (NEW)
type Provider interface {
    Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
    Stream(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error)
    ListModels(ctx context.Context) ([]Model, error)
    Name() string
    Capabilities() Capabilities
}

type Model struct {
    ID              string
    Name            string
    ContextWindow   int
    MaxOutputTokens int
    SupportsTools   bool
    SupportsVision  bool
    CanReason       bool
    CostPer1M       ModelCost
}

type ChatRequest struct {
    Model       string
    Messages    []Message
    Tools       []Tool
    MaxTokens   int
    Temperature *float64
    TopP        *float64
    Stream      bool
}

type ChatResponse struct {
    Content      string
    ToolCalls    []ToolCall
    Usage        Usage
    FinishReason string
}
```

**Phase 2**: Implement Provider Adapters (8-12 hours)
- `internal/llm/providers/openai.go`
- `internal/llm/providers/anthropic.go`
- `internal/llm/providers/mistral.go`
- `internal/llm/providers/gemini.go`
- `internal/llm/providers/groq.go`
- `internal/llm/providers/openai_compat.go` (LM Studio, vLLM, etc.)

**Phase 3**: Response Format Normalization (6-8 hours)
- `internal/llm/format/converter.go`
- Unified message/tool format conversion

**Phase 4**: Update Config System (4-6 hours)
- Replace catwalk model structs
- Use new provider registry

**Phase 5**: Update Agent/Coordinator (6-8 hours)
- Switch to new provider interface
- Remove fantasy imports

**Phase 6**: Tests (2-3 hours)
- Update existing tests
- Add provider adapter tests

#### Migration Strategy
1. **Wrap existing** fantasy providers initially (no breaking changes)
2. **Migrate one provider** at a time (OpenAI first)
3. **Run both systems** in parallel with feature flag
4. **Switch default** once validated
5. **Remove fantasy** after full migration

#### Dependencies to Add
- `google.golang.org/genai` - Gemini
- `github.com/groq/groq-go` - Groq SDK (if available)
- Keep: `github.com/openai/openai-go`
- Keep: `github.com/charmbracelet/anthropic-sdk-go`

#### Dependencies to Remove
- `charm.land/fantasy`
- `github.com/charmbracelet/catwalk` (config only)

---

### 8. Move Provider-Specific Logic to Config 🔧
**Priority**: P2 - Medium  
**Effort**: 4-6 hours  
**Impact**: Code clarity, maintainability

**Issue**: Provider quirks scattered in code

**Current State**:
```go
// internal/agent/agent.go:387-390
if provider == "cerebras" || provider == "zai" {
    toolChoice := fantasy.ToolChoiceAuto
    prepared.ToolChoice = &toolChoice
}

// internal/agent/agent.go:934-946
if strings.Contains(model, "glm-4.6") || strings.Contains(provider, "cerebras") {
    agent = fantasy.NewAgent(a.smallModel.Model, ...)
}
```

**Proposed State**:
```toml
# config.toml
[providers.cerebras]
quirks = ["force_tool_choice", "use_small_model_for_summarization"]

[providers.zai]
quirks = ["force_tool_choice"]

[providers.mistral]
quirks = ["alphanumeric_tool_ids"]
```

```go
// internal/config/provider_quirks.go (NEW)
type ProviderQuirks struct {
    ForceToolChoice              bool
    UseSmallModelForSummarization bool
    AlphanumericToolIDs          bool
}

func (p *Provider) Quirks() ProviderQuirks {
    return parseQuirks(p.Config.Quirks)
}
```

**Files to Create**:
- `internal/config/provider_quirks.go`

**Files to Modify**:
- `internal/agent/agent.go` (use config instead of inline checks)
- `internal/config/provider.go` (add quirks field)

---

### 9. TUI Performance Optimization 🖥️
**Priority**: P2 - Medium  
**Effort**: 6-8 hours  
**Impact**: User experience (responsiveness)

#### Issues
1. **Cursor positioning** (editor.go:203, 345)
   - Cursor always moves to end
   - Editing in middle doesn't work
   
2. **Message list performance**
   - Re-renders entire list on every update
   - No virtualization for long sessions

#### Solutions

**9a. Fix Cursor Positioning** (4-6 hours)
```go
// internal/tui/components/chat/editor/editor.go
// Track cursor position independently
type Editor struct {
    textarea     textarea.Model
    cursorOffset int  // NEW: explicit cursor tracking
}

// Update cursor after edits
func (e *Editor) handleKeyPress(msg tea.KeyMsg) {
    switch msg.Type {
    case tea.KeyLeft:
        e.cursorOffset--
    case tea.KeyRight:
        e.cursorOffset++
    // ... handle all navigation keys
    }
}
```

**9b. Virtualized Message List** (2-3 hours)
```go
// internal/tui/components/chat/message_list.go
type MessageList struct {
    messages       []Message
    viewportStart  int  // NEW: only render visible range
    viewportEnd    int
    visibleCount   int  // e.g., 20 messages
}

func (m *MessageList) View() string {
    // Only render messages[viewportStart:viewportEnd]
    visible := m.messages[m.viewportStart:m.viewportEnd]
    // ... render only visible
}
```

---

## 🔮 STRATEGIC INITIATIVES (3-6 Months)

### 10. Multi-Agent Architecture 🤖
**Priority**: P3 - Low (Decide first)  
**Effort**: TBD (depends on scope)  
**Status**: ⚠️ NEEDS DECISION

**Current State**: 4 TODO comments, unused `map[string]SessionAgent`

**Options**:
1. **Implement Now** - Keep TODOs, build architecture
2. **Plan Later** - Create GitHub issues, remove TODOs, simplify code
3. **Not Needed** - Remove entirely, commit to single-agent model

**Questions to Answer**:
- What's the use case for multiple agents?
- Different agents for different tasks (coding, research, etc.)?
- Concurrent agents in same session?
- Agent handoff/coordination?

**Recommendation**: **Defer until Q1 2026** - Focus on single-agent quality first

**If Deferred**: Remove TODOs, simplify coordinator to single agent

---

### 11. Plugin System 🔌
**Priority**: P3 - Low  
**Effort**: 15-20 hours  
**Impact**: Extensibility, community contributions

**Goal**: Allow custom tools without modifying core code

**Design**:
```go
// internal/plugin/plugin.go
type Plugin interface {
    Name() string
    Description() string
    Tools() []fantasy.AgentTool
    Initialize(ctx context.Context) error
    Cleanup() error
}

// Example plugin
type GitHubPlugin struct {
    token string
}

func (p *GitHubPlugin) Tools() []fantasy.AgentTool {
    return []fantasy.AgentTool{
        CreateIssue,
        CreatePR,
        ListRepos,
    }
}
```

**Discovery**:
- Plugins in `~/.config/nexora/plugins/`
- Load via Go plugin system or gRPC
- Configuration in `config.toml`

---

### 12. Distributed Tracing 📊
**Priority**: P3 - Low  
**Effort**: 8-12 hours  
**Impact**: Observability, debugging

**Goal**: OpenTelemetry integration for performance insights

**Implementation**:
```go
// internal/telemetry/trace.go
import "go.opentelemetry.io/otel"

func TraceAgentRun(ctx context.Context, sessionID string) (context.Context, trace.Span) {
    return otel.Tracer("nexora").Start(ctx, "agent.run",
        trace.WithAttributes(
            attribute.String("session.id", sessionID),
        ),
    )
}
```

**Spans to Trace**:
- Agent.Run (full lifecycle)
- Tool execution (each tool)
- Provider API calls
- Database queries
- AIOPS calls

**Export**:
- Jaeger (local dev)
- Honeycomb (production)
- Console (debug)

---

## 🚀 ENHANCEMENTS (Backlog)

### Performance
- [ ] Connection pooling for HTTP clients
- [ ] Streaming response optimization
- [ ] Token counting optimization
- [ ] Model list caching on startup

### Features
- [ ] Vision support across all providers
- [ ] Batch processing API
- [ ] Rate limiting with exponential backoff
- [ ] Provider health checks
- [ ] Fallback provider on primary failure
- [ ] Web interface (React/Svelte)
- [ ] VS Code extension

### Developer Experience
- [ ] Hot reload for config changes
- [ ] Interactive setup wizard
- [ ] Shell completions (bash, zsh, fish)
- [ ] Man pages
- [ ] Docker image

### Quality
- [ ] Property-based testing (edit tool)
- [ ] Chaos testing (network failures)
- [ ] Load testing (concurrent requests)
- [ ] Security audit (SAST, dependency scan)

---

## ❓ OPEN QUESTIONS (Need Decisions)

### Question 1: Environment Detection Configuration
**Context**: 15+ new fields add 200-500ms latency

**Options**:
1. **Always enabled** (current) - Best context, slower
2. **Configurable flag** (`--minimal-prompt`) - User choice
3. **Cached with TTL** (60s) - Balance of both

**Recommendation**: **Option 3** (cache with TTL) - Implement in Priority #4

---

### Question 2: AIOPS Fallback Strategy
**Context**: Edit tool has local fuzzy → remote AIOPS fallback

**Options**:
1. **Local-first** - Skip AIOPS if fuzzy fails (faster)
2. **Accuracy-first** - Always try AIOPS (current, slower)
3. **Configurable** - User preference in config

**Recommendation**: **Option 3** (configurable)
```toml
[agent.edit]
fallback_strategy = "accuracy-first"  # or "local-first"
```

---

### Question 3: Multi-Agent Timeline
**Context**: 4 TODO comments, unused code

**Options**:
1. **Near-term (1-2 months)** - Keep TODOs, plan architecture
2. **Mid-term (3-6 months)** - Create issues, remove TODOs
3. **Long-term/uncertain** - Simplify to single-agent

**Recommendation**: **Option 2** (mid-term) - Create GitHub issues, clean up code now

---

### Question 4: Test Coverage Target
**Context**: Current 23% coverage

**Options**:
1. **Conservative (40-50%)** - Double current, achievable
2. **Standard (60-70%)** - Industry average
3. **High (80%+)** - Comprehensive coverage

**Recommendation**: **Option 1** (40-50%) for Q1 2026, then reevaluate

---

### Question 5: Provider Configuration Location
**Context**: Quirks scattered in code vs. centralized config

**Options**:
1. **Code** (current) - Fast, but scattered
2. **Configuration** - Centralized, schema overhead
3. **Provider plugins** - Flexible, complex

**Recommendation**: **Option 2** (configuration) - Implement in Priority #8

---

## 📅 Timeline

### Week 1 (Current)
- ✅ Code audit complete
- ✅ Documentation updated
- 🔴 **FIX TURBO MODE** (Priority #0) - BROKEN, needs re-implementation
- 🔴 Fix memory leak (Priority #1)
- 🔴 Fix provider options bug (Priority #2)
- 🟡 Quick performance wins (Priority #3)

### Week 2-3
- 🟡 Prompt cache implementation (Priority #4)
- 🟡 Architecture Decision Records (Priority #5)
- 🟡 Test coverage expansion start (Priority #6)

### Week 4-5
- 🟡 Test coverage expansion continue
- 🟢 TUI performance optimization (Priority #9)
- 🟢 Provider config refactor (Priority #8)

### Month 2
- 🟢 Provider abstraction refactor start (Priority #7)
- 🟢 Continue test coverage (target 40%)

### Month 3
- 🟢 Provider abstraction refactor complete
- 🟢 Documentation updates
- 🟢 Performance benchmarking

### Q1 2026
- 🔵 Plugin system exploration (Priority #11)
- 🔵 Multi-agent decision (Priority #10)
- 🔵 Distributed tracing (Priority #12)

---

## 🎯 Success Metrics

### Performance
- [ ] Prompt generation: <100ms (from 300-800ms)
- [ ] Edit tool success rate: >95% (from 90%)
- [ ] Memory usage: Stable over 24hr run
- [ ] Zero OOM crashes

### Quality
- [ ] Test coverage: 40%+ (from 23%)
- [ ] Zero P0 bugs in production
- [ ] All ADRs documented
- [ ] Code quality: A+ (from A-)

### Adoption
- [ ] GitHub stars: +20%
- [ ] Active users: +30%
- [ ] Community contributions: +5 PRs/month
- [ ] Documentation completeness: 100%

---

## 🎓 Lessons Learned

### 2025-12-17: Turbo Mode Syntax Errors
**What Happened**: Attempted to add Turbo Mode detection in agent.go but introduced multiple syntax errors:
1. Used template syntax `{{if isTurbo}}` in Go struct literal (lines 349-428)
2. Accidentally removed `package agent` declaration from summarizer.go
3. Placed Turbo detection logic in wrong scope (inside PrepareStep func instead of agent-level)

**Impact**: Build broken, required full revert to commit 9f39ceb

**Root Causes**:
- Mixed template syntax with Go code
- Insufficient validation before committing
- Scope confusion (function-local vs agent-level)

**Preventions**:
- ✅ Always run `go build .` before committing
- ✅ Verify package declarations in all modified files
- ✅ Review scope placement (where should variables live?)
- ✅ Test incremental changes, not bulk edits
- ✅ Use `git diff` to review all changes before commit

**Next Steps**:
- Re-implement Turbo Mode with proper Go syntax
- Add test coverage for Turbo Mode detection
- Document Turbo Mode behavior in ADR

---

## 🔗 References

- **Architecture**: `codedocs/ARCHITECTURE.md`
- **Code Audit**: `codedocs/codeaudit-12-17-2025.md`
- **Recent Changes**: `codedocs/RECENT_CHANGES.md`
- **Operations**: `PROJECT_OPERATIONS.md`
- **TODOs**: `todo.md` (strategic backlog)
- **Code Navigation**: `CODEDOCS.md`

---

## 📝 Change Log

- **2025-12-17**: Initial roadmap created from code audit + todo.md
- **2025-12-17 (Evening)**: Turbo Mode implementation reverted due to syntax errors
  - **Issue**: Template syntax `{{if isTurbo}}` used in Go struct literal (lines 349-428 in agent.go)
  - **Issue**: `summarizer.go` lost `package agent` declaration
  - **Issue**: Turbo detection logic placed in wrong scope (PrepareStep func instead of agent scope)
  - **Resolution**: Reverted to clean state (commit 9f39ceb)
  - **Status**: ✅ Build PASS, ✅ Tests PASS, ✅ Git CLEAN
  - **Lesson**: Avoid template syntax in Go code; validate package declarations; verify scope placement
  - **Next**: Re-implement Turbo Mode with proper Go syntax and testing
- **Next Review**: 2026-01-17 (monthly cadence)

---

**Maintained by**: Nexora Core Team  
**Questions?**: Open GitHub issue or discussion
