# ModelScan Integration Plan - Multi-Agent AI SDK

**Location**: `/home/nexora/.local/tools/modelscan`  
**Purpose**: Production-ready Go SDK for 21+ LLM providers with multi-agent framework  
**Status**: 🟢 **READY FOR INTEGRATION** (80.1% test coverage, 149 tests passing)

---

## Executive Summary

**ModelScan** is a comprehensive Go SDK providing:
1. **21 Production SDKs** - All major LLM providers (OpenAI, Anthropic, Google, xAI, etc.)
2. **Multi-Agent Framework** - Coordinator, teams, messaging, workflows
3. **Smart Routing** - Cheapest, fastest, balanced strategies with health tracking
4. **Rate Limiting** - Token bucket algorithm with SQLite persistence
5. **Unified Streaming** - SSE, WebSocket, HTTP chunked across all providers
6. **Zero Dependencies** - Pure Go stdlib

**This is a SEPARATE product** that Nexora can integrate with, not replace.

---

## Project Stats

### Code Size
- **7,168 lines** of Go code (22 files)
- **149 tests passing** (100% pass rate)
- **80.1% average coverage** (90%+ on critical paths)
- **3 CLI tools** (modelscan, seed-db, demo)

### Core Packages
```
sdk/
├── agent/           # Multi-agent framework (86.5% coverage)
│   ├── coordinator  # Task distribution strategies
│   ├── team         # Team management & messaging
│   ├── messagebus   # Inter-agent communication
│   ├── memory       # Agent memory systems
│   ├── workflow     # Orchestration engine
│   ├── tools        # Tool registry & execution
│   └── react        # ReAct planning algorithm
├── router/          # Intelligent routing (86.2% coverage)
├── ratelimit/       # Token bucket rate limiting (90.1% coverage)
├── stream/          # Unified streaming API (92.3% coverage)
├── storage/         # SQLite persistence (74.7% coverage)
└── cli/             # CLI orchestration (71.7% coverage)

providers/           # Provider discovery (68.1% coverage)
storage/             # Rate limit storage (73.1% coverage)
```

---

## Key Features

### 1. Multi-Agent Coordination
```go
// Create a team with specialized agents
team := coordinator.CreateTeam("dev-team")
coordinator.AddAgent(team.ID, dataAgent)
coordinator.AddAgent(team.ID, codeAgent)

// Distribute task with load balancing
result := coordinator.DistributeTasks(team.ID, tasks, LoadBalance)
```

**Features:**
- Task Distribution: RoundRobin, LoadBalance, Priority strategies
- Capability Matching: Agents matched to tasks based on skills
- Team Context: Shared memory and message broadcasting
- Message Bus: Inter-agent communication with statistics

### 2. Intelligent Router
```go
// Route to cheapest provider under $0.10 budget
provider := router.RouteRequest(request, RouterOptions{
    Strategy: Cheapest,
    Constraints: RouterConstraints{
        MaxCost: 0.10,
        MaxLatency: 500 * time.Millisecond,
    },
})
```

**Strategies:**
- `Cheapest` - Optimize for cost
- `Fastest` - Optimize for latency
- `Balanced` - Balance cost/latency
- `RoundRobin` - Distribute evenly
- `Fallback` - Try primary, fallback on failure

**Health Tracking:**
- Exponential moving average latency
- Failure counting with automatic fallback
- Provider blacklisting after 3+ failures

### 3. Rate Limiting
```go
// Token bucket with multiple limits
limiter := ratelimit.NewBucket(Limits{
    RPM: 500,  // Requests per minute
    TPM: 200_000,  // Tokens per minute
    RPD: 10_000,  // Requests per day
    Concurrent: 10,  // Max concurrent
})

// Wait for rate limit clearance
err := limiter.Wait(ctx, 1000) // 1000 tokens
```

**Features:**
- Multi-limit coordination (RPM + TPM + RPD)
- Burst allowance support
- Automatic rollback on failure
- Thread-safe (tested with 10 concurrent goroutines)
- SQLite WAL persistence

### 4. Unified Streaming
```go
// Stream from any provider
stream := client.Stream(ctx, request)
for event := range stream.Events() {
    fmt.Print(event.Content)
}

// Stream operators
filtered := stream.
    Filter(func(e Event) bool { return len(e.Content) > 0 }).
    Map(func(e Event) Event { e.Content = strings.ToUpper(e.Content); return e })
```

**Formats Supported:**
- SSE (Server-Sent Events) - OpenAI, Anthropic
- WebSocket - Google, xAI
- HTTP Chunked - Mistral, Together
- Generic line-delimited JSON

### 5. Provider SDKs (21 total)
```go
// Consistent API across all providers
import "github.com/jeffersonwarrior/modelscan/sdk/openai"
import "github.com/jeffersonwarrior/modelscan/sdk/anthropic"
import "github.com/jeffersonwarrior/modelscan/sdk/google"

client := openai.NewClient("key")
resp, _ := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
    Model: "gpt-4",
    Messages: []openai.ChatMessage{
        {Role: "user", Content: "Hello!"},
    },
})
```

**Coverage:**
- Core: OpenAI, Anthropic, Google, Mistral
- Direct: xAI, DeepSeek, Minimax, Kimi, Z.AI, Cohere
- Aggregators: OpenRouter, Synthetic, Vibe, NanoGPT
- Inference: Together AI, Fireworks, Groq, Replicate, DeepInfra, Hyperbolic, Perplexity

---

## Database Schema

### SQLite Tables
```sql
-- Rate limit tracking
CREATE TABLE rate_limits (
    provider TEXT NOT NULL,
    limit_type TEXT NOT NULL,
    value INTEGER NOT NULL,
    tier TEXT DEFAULT 'standard'
);

-- Pricing data
CREATE TABLE pricing (
    provider TEXT NOT NULL,
    model TEXT NOT NULL,
    input_cost REAL NOT NULL,
    output_cost REAL NOT NULL,
    currency TEXT DEFAULT 'USD'
);

-- Agent persistence
CREATE TABLE agents (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    capabilities TEXT,  -- JSON array
    status TEXT DEFAULT 'active',
    created_at TIMESTAMP
);

-- Task tracking
CREATE TABLE tasks (
    id TEXT PRIMARY KEY,
    description TEXT,
    assigned_agent_id TEXT,
    status TEXT DEFAULT 'pending',
    result TEXT,
    created_at TIMESTAMP,
    completed_at TIMESTAMP
);

-- Message history
CREATE TABLE messages (
    id TEXT PRIMARY KEY,
    sender_id TEXT,
    recipient_id TEXT,
    team_id TEXT,
    content TEXT,
    timestamp TIMESTAMP
);
```

---

## Integration with Nexora

### Option 1: Embedded Library (Recommended)
**Use ModelScan as a library within Nexora**

```go
// In nexora/internal/agent/providers/modelscan.go
import "github.com/jeffersonwarrior/modelscan/sdk/agent"
import "github.com/jeffersonwarrior/modelscan/sdk/router"

type ModelScanProvider struct {
    coordinator *agent.Coordinator
    router      *router.Router
}

func (p *ModelScanProvider) ExecuteTask(ctx context.Context, task Task) (Result, error) {
    // Use ModelScan's multi-agent coordination
    return p.coordinator.DistributeTask(task)
}
```

**Benefits:**
- Direct integration, no IPC overhead
- Share context and state
- Unified error handling
- Single binary

**Drawbacks:**
- Tight coupling
- Harder to version independently

### Option 2: CLI Subprocess (Current)
**Spawn modelscan CLI as subprocess**

```go
// In nexora/internal/agent/tools/modelscan.go
func CallModelScan(ctx context.Context, args []string) (string, error) {
    cmd := exec.CommandContext(ctx, "modelscan", args...)
    output, err := cmd.CombinedOutput()
    return string(output), err
}
```

**Benefits:**
- Loose coupling
- Independent versioning
- Process isolation

**Drawbacks:**
- IPC overhead
- JSON serialization costs
- Process management complexity

### Option 3: Hybrid Approach (Best)
**Use library for core, CLI for advanced features**

```go
// Core SDK embedded
import "github.com/jeffersonwarrior/modelscan/sdk/router"

// Advanced features via CLI
func (n *Nexora) CreateAgentTeam(name string) error {
    return exec.Command("modelscan", "create-team", "--name", name).Run()
}
```

**Benefits:**
- Best of both worlds
- Performance where it matters
- Flexibility for complex workflows

---

## Roadmap Alignment

### ModelScan Roadmap (8 phases)
1. ✅ **Phase 0**: Foundation (Rate limiting, routing, streaming)
2. ✅ **Phase 0.5**: Agent Framework (Multi-agent coordination)
3. 🔄 **Phase 1**: Database Integration (SQLite persistence)
4. 🔜 **Phase 2**: CLI Orchestration (Main CLI layer)
5. 🔜 **Phase 3**: Tool/Skill/MCP Management
6. 🔜 **Phase 4**: Advanced Task Management
7. 🔜 **Phase 5**: Security & Isolation
8. 🔜 **Phase 6**: Observability & Monitoring

### Nexora Roadmap
- **0.29.0**: CLI agent with tool system ✅ (current)
- **0.30.0**: PostgreSQL migration (2 weeks)
- **0.31.0**: VNC Gateway alpha (2 weeks)
- **0.32.0**: VNC Gateway beta (2 weeks)
- **0.3.0**: Production release (2 weeks)

### Integration Timeline
**Option A: Embed Now**
- Week 1: Import ModelScan SDK packages
- Week 2: Integrate router + rate limiting
- Week 3: Test with real workloads
- Week 4: Documentation + examples

**Option B: CLI Integration**
- Week 1: Add modelscan as git submodule
- Week 2: Build system integration
- Week 3: Subprocess IPC layer
- Week 4: Error handling + testing

**Option C: Defer to 0.30.0+**
- Keep ModelScan separate for now
- Focus on 0.3.0 VNC architecture
- Integrate after VNC stabilizes
- Use as provider abstraction layer

---

## Recommendations

### For 0.29.0 (Now)
**Do NOT integrate yet**. Reasons:
1. ModelScan is still in Phase 1 (database work ongoing)
2. Nexora 0.29.0 audit fixes should ship clean
3. 0.3.0 VNC migration is bigger priority
4. Better to integrate after both stabilize

**Action**: Document integration plan, defer to 0.30.0+

### For 0.30.0+ (Future)
**Use Option 3: Hybrid Approach**
1. Embed ModelScan SDK for core routing/rate-limiting
2. Use ModelScan CLI for advanced multi-agent features
3. Share SQLite database for persistence
4. Coordinate releases (Nexora 0.30.0 with ModelScan 1.0)

**Benefits:**
- Performance where it matters (embedded SDK)
- Flexibility for complex workflows (CLI)
- Clean separation of concerns
- Independent versioning

---

## Files to Review

### Core SDK
- `sdk/agent/coordinator.go` - Multi-agent coordination
- `sdk/router/router.go` - Intelligent routing
- `sdk/ratelimit/bucket.go` - Rate limiting
- `sdk/stream/stream.go` - Unified streaming
- `sdk/storage/database.go` - SQLite persistence

### Documentation
- `README.md` - Quick start guide
- `ROADMAP.md` - 8-phase development plan
- `TODO.md` - Current tasks and strategy
- `SESSION_SUMMARY.md` - Latest progress
- `SDK_USAGE_GUIDE.md` - Integration examples

### Tests
- `sdk/agent/*_test.go` - 149 tests, 86.5% coverage
- `sdk/router/*_test.go` - Router tests, 86.2% coverage
- `sdk/ratelimit/*_test.go` - Rate limit tests, 90.1% coverage

---

## Summary

**ModelScan is a production-ready, separate Go SDK** with:
- ✅ 21 provider SDKs
- ✅ Multi-agent framework
- ✅ Smart routing
- ✅ Rate limiting
- ✅ 80% test coverage
- ✅ Zero dependencies

**Integration Strategy:**
1. **Now (0.29.0)**: Keep separate, document plan
2. **Later (0.30.0+)**: Hybrid approach (embedded + CLI)
3. **Goal**: Use ModelScan as provider abstraction layer

**Current Priority**: Ship Nexora 0.29.0 clean, integrate ModelScan after 0.3.0 VNC stabilizes.

---

**Date**: December 18, 2025  
**Location**: `/home/nexora/.local/tools/modelscan`  
**Status**: Ready for integration planning
