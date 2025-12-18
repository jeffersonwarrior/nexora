## [0.28.5] - 2025-12-17 - **Token Efficiency & Production Polish**

### 🎯 Performance Optimization
- **Token Reduction**: 30k 	 27k tokens (11% reduction) in session startup
  - Tool documentation compressed by 37% (35KB 	 22KB)
  - edit.md: 9.3KB 	 3.7KB (-60%)
  - bash.tpl: 5.2KB 	 3.6KB (-31%)
  - multiedit.md: 5KB 	 3.6KB (-28%)
  - coder.md.tpl: 7.2KB 	 5.3KB (-26%)
  - agentic_fetch.md: 2.9KB 	 1.6KB (-45%)
  - job tools consolidated: 1KB 	 0.5KB (-51%)
- **Runtime Optimizations**:
  - Git commits reduced: 3 	 2 in logs
  - Git status reduced: 20 	 5 files shown
  - Network/services lazy-loaded (set `NEXORA_FULL_ENV=1` to enable)
  - Default assumes "online", skips expensive ping/systemctl checks

### 📚 Documentation
- Added TOKEN_REDUCTION.md with detailed breakdown
- Updated ROADMAP.md with completion status
- All tool descriptions remain comprehensive while being more concise

### 💰 Benefits
- **Faster startup**: Less initial context to process
- **Lower costs**: 11% token reduction = 11% cost reduction per session
- **Better context window**: More room for actual conversation
- **Same functionality**: All features preserved

### 🧪 QA Results
```
✅ go test ./... 	 20+ packages, zero failures
✅ go build . 	 Clean build
✅ All tool descriptions 	 Still comprehensive
✅ No functionality lost
```

---

## [0.28.7] - 2025-12-17 - **Local Model Support (Beta)**
- **Local Model Support (Beta)**: Ollama/LM-Studio integration
  - UI configuration + clear error messages
  - Beta stability with production fallbacks

### 🧪 QA Results
```
✅ go test ./... 	 20+ packages, zero failures
✅ make test-qa 	 Production validation suite
✅ ./build/nexora -y 	 Zero crashes
✅ Local model endpoints 	 Responding correctly
```

---

- **Anthropic `max_tokens=0` ERROR**: Bulletproof validation in **ALL** call sites (summarization, title gen, main agent)
  - `ensurePositiveMaxTokens()` → **0 becomes 4096 automatically**
  - Explicit fallbacks: summarization (4096), titles (100), tools (triple-checked)
- **SQLite Migration**: Fixed `context_archive` table (inline indexes → separate CREATE INDEX)
- **Local Models UI**: Clear error messages + examples (Ollama/LM-Studio ports)

### ✨ Features
- **Z.AI Vision MCP**: **@z_ai/mcp-server** added to main config (`make setup`)
  - Vision analysis, web reader, web search
  - Zero-config production setup
- **Session Title Updates**: Auto-update every **25 messages** (configurable)
  - Context-aware (last 10 user messages)
  - Thread-safe counters (`csync.Map`)
  - Deduplication (no redundant updates)

### 🧪 QA Results
```
✅ go test ./... → 20+ packages, zero failures
✅ make test-qa → Production validation suite
✅ ./build/nexora -y → Zero crashes
✅ All migrations → Applied successfully
```

---

## [0.28.5] - 2025-12-17 - **Zero-Config Production**

### 🚀 Major Features
- **`make setup`** - **One-command production** (9 API providers + Z.AI MCP Vision)
  - xAI Grok-4.1, Cerebras GLM-4.6, Anthropic Claude 4.5, OpenAI GPT-5.2
  - **Zero permission prompts** (`skip_requests: true`)
  - **Auto-loads `.env`** API keys

### 🛠️ Reliability
- **max_tokens validation**: **0 → 4096** everywhere (Anthropic, all providers)
- **Thread-safe agent queue**: 50+ concurrent requests
- **Auto model fallback**: Invalid models → recent working models

---

## [0.28.4] - 2025-12-17 - **QA Framework**

- **Dedicated QA**: `qa/` folder + `make test-qa`
- **Build verification**: Clean builds required
- **Tool ID sanitization**: Mistral/OpenAI format compliance

---

## [0.28.1] - 2025-12-17 - **Critical Stability**

### 🐛 Fixed
- **Mistral Tool IDs**: `call_61626542` → 9-char alphanumeric format
- **Session Titles**: Fixed truncation + \"New Session\" fallback  
- **Cerebras/GLM-4.6**: Provider config + API compatibility
- **View Tool**: Context explosion → smart truncation

---

## [0.28.0] - 2025-12-17 - **Provider Expansion**

- **13 Providers**, **70+ models** fully operational
- **Multi-provider tool calls** (parallel execution)
- **Production-grade error recovery**