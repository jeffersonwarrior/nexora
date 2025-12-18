# Nexora TODO & Refactoring Plan

## High Priority: Provider/Model Abstraction Refactor

### Current State
- Using `charm.land/fantasy` v0.5.1 for LLM abstraction
- Using `github.com/charmbracelet/catwalk` v0.9.5 for provider/model config
- Provider config in `internal/config/provider.go` (~420 lines)

Replace fantasy/catwalk with a unified provider abstraction that:
1. **Reduces coupling** - Single interface for all providers
2. **Adds provider support** - Gemini, Groq, Cohere, Perplexity, local llama.cpp, ollama
3. **Cleaner request/response translation** - One place for OpenAI↔Mistral↔Anthropic format conversions
4. **Better error handling** - Provider-specific error mapping
5. **Config management** - Single source of truth for models and their capabilities

### Implementation Phases

#### Phase 1: Create Provider Interface (LOW EFFORT, HIGH VALUE)
**Effort**: ~4-6 hours
**Files to create**:
- `internal/llm/provider.go` - Core interface (150-200 lines)
  - `Provider` interface with Chat/Completion/Embedding methods
  - `Model` struct with capabilities (context, cost, reasoning, etc)
  - `Request`/`Response` wrappers for normalized format
  - `ChatOptions` for streaming, tools, temperature, etc

- `internal/llm/client.go` - Provider registry (100-150 lines)
  - `ProviderRegistry` to load/switch providers at runtime
  - Discovery mechanism for available models
  - Cost tracking per model

**Backward Compat**: Wrap existing fantasy/catwalk, no breaking changes yet

#### Phase 2: Implement Provider Adapters (MEDIUM EFFORT)
**Effort**: ~8-12 hours (1-2 providers per hour once pattern is established)

**To Create**:
1. `internal/llm/providers/openai.go` - Direct OpenAI client
2. `internal/llm/providers/anthropic.go` - Anthropic API
3. `internal/llm/providers/mistral.go` - Mistral API
4. `internal/llm/providers/gemini.go` - Google Gemini
5. `internal/llm/providers/groq.go` - Groq (fast inference)
6. `internal/llm/providers/openai_compat.go` - OpenAI-compatible (Mistral local, vLLM, LM Studio)
7. `internal/llm/providers/fantasy_wrapper.go` - Wrap existing for migration

**Per adapter includes**:
- Request normalization (tools, messages, parameters)
- Response translation to standard format
- Error mapping
- Token counting
- Rate limit handling

#### Phase 3: Response Format Normalization (MEDIUM EFFORT)
**Effort**: ~6-8 hours

**Create**:
- `internal/llm/format/` directory
  - `openai.go` - OpenAI request/response types + translation
  - `anthropic.go` - Anthropic types + translation
  - `mistral.go` - Mistral types + translation
  - `gemini.go` - Gemini types + translation
  - `converter.go` - Unified conversion logic

**Why separate**: Each provider has different:
- Message/tool formats
- Streaming chunk structure
- Error codes/messages
- Token usage fields
- Model naming conventions

#### Phase 4: Update Config System (LOW-MEDIUM EFFORT)
**Effort**: ~4-6 hours

**Changes to `internal/config/`**:
- Replace catwalk model structs with unified `llm.Model`
- Keep provider.go for loading from disk/env
- Remove provider.go injection logic → use new registry
- Update provider_test.go with new interface

#### Phase 5: Update Agent/Coordinator (MEDIUM EFFORT)
**Effort**: ~6-8 hours

**Files to update**:
- `internal/agent/coordinator.go` - Switch to new provider interface
- `internal/agent/event.go` - Use unified usage tracking
- `internal/agent/common_test.go` - Update test fixtures

**What changes**:
- Remove fantasy imports
- Use new provider.Chat() method
- Standardized error handling
- Same business logic, cleaner implementation

#### Phase 6: Update Config/Provider Tests (LOW EFFORT)
**Effort**: ~2-3 hours

**Tests to update**:
- `internal/config/provider_test.go`
- `internal/config/provider_empty_test.go`
- `internal/config/recent_models_test.go`
- Add new provider adapter tests

### Total Effort Estimate
- **Aggressive** (dedicated): 25-35 hours (~1 week full-time)
- **Gradual** (part-time): 4-5 hours/week
- **With refactoring surprises**: Add 20-30%

### Benefits
- ✅ Smaller, more maintainable code
- ✅ Easier to add new providers (1-2 hours per provider)
- ✅ Better error handling per provider
- ✅ Unified testing across all providers
- ✅ Can deprecate fantasy/catwalk (remove 2 dependencies)
- ✅ Better performance (direct API calls vs abstraction)
- ✅ Easier to support bleeding-edge provider features

### Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Break existing agent flow | Wrap old code, migrate incrementally, test with existing VCR cassettes |
| Lose fantasy's abstraction benefits | Gain them back with cleaner design, fewer dependencies |
| Model metadata conflicts | Centralize in new `Model` struct, single source of truth |
| Token counting differences | Implement per-provider, validate against actual API usage |

### Dependencies to Add
- `google.golang.org/genai` - Gemini (or REST if preferred)
- `github.com/groq/groq-go` - Groq SDK (if available)
- Keep: `github.com/openai/openai-go`
- Keep: `github.com/charmbracelet/anthropic-sdk-go`
- Consider: `github.com/mistralai/client-go`

### Dependencies to Remove
- `charm.land/fantasy` - Full replacement
- `github.com/charmbracelet/catwalk` - Config only, can replace with JSON/YAML

---

## Core Reliability Fixes

### UI & Context Issues
- [ ] **UI cursor adjustments** - Fix FIXME comments in `splash.go` and `models.go`
- [ ] **Remaining context.TODO()** (20 instances) - Replace with proper contexts

---

## 🧪 TEST FAILURES (Fix Immediately)
### Database & Indexing
- [x] **SQLite FTS schema errors** - "no such column: fts" and "no such table: symbols_fts" in search queries
- [x] **Database migration concurrency** - Concurrent read/write failures during index operations
- [x] **Index table creation** - Missing FTS tables in new databases

### Command Line Interface
- [ ] **Nexora crash handling** - Generic crash messages instead of specific errors in index/query commands
- [ ] **Invalid path handling** - "no such file" messages not appearing for invalid paths
- [ ] **Bedrock credential validation** - Failing credential tests expecting 0/1 counts

---

## Known Issues

### Agent Loop with Devstral-2 (BLOCKER)
**Issue**: When reading large files like todo.md, nexora enters infinite loop with devstral-2
**Root Cause**: Agent is sending full system context (including all tool definitions) to LLM, causing devstral to repeatedly try tool invocations
**Solution**: Implement tool result aggregation and prevent re-prompting with same context
**Status**: NEEDS INVESTIGATION - may be devstral-2 tool handling, not nexora

### Streaming Response Conversion
**Issue**: Devstral proxy converts streaming→non-streaming internally, then converts back to SSE
**Better approach**: Support true streaming passthrough or use polling with proper chunking

---

## Other TODOs

### Bugs/Issues
- [ ] Agent loop prevention (max iterations, cycle detection)
- [ ] Provider connection timeouts not always handled gracefully
- [ ] Model cost tracking doesn't account for cache hits

### Features
- [ ] Add function calling to all providers (currently partial)
- [ ] Implement vision support across providers
- [ ] Add batch processing API
- [ ] Rate limiting per provider with backoff
- [ ] Provider health checks
- [ ] Fallback provider when primary fails

### Performance
- [ ] Cache model list on startup
- [ ] Connection pooling for HTTP clients
- [ ] Streaming response optimization
- [ ] Token counting optimization

### Testing
- [ ] Expand provider adapter tests to 90%+ coverage
- [ ] Add integration tests with real APIs (VCR cassettes)
- [ ] Load testing with multiple concurrent requests
- [ ] Error scenario coverage

### Documentation
- [ ] Add provider setup guides (API keys, regions, etc)
- [ ] Document new provider interface for contributors
- [ ] API compatibility matrix
- [ ] Migration guide from fantasy→new system

---

## Banned Commands

**DO NOT DO**:
- `pkill nexora` - Kills the running nexora process ungracefully
- `pkill go` - Kills all Go processes, including the development server and builds