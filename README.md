# Nexora - 🚀 AI-Powered CLI Agent

> **Production-Ready AI Terminal Assistant** with **tool execution**, **multi-provider support**, **vision MCP**, and **zero-config setup**.

## ✨ Quick Start

```bash
# Install (production config + 9 API providers + Vision MCP)
curl -fsSL https://nexora.land/install.sh | sh

# Or build from source
git clone https://github.com/nexora/cli.git && cd cli
make build && make setup

# Start chatting
nexora chat
```

## 🎯 Features

| 🛠️ **Tools** | 🔗 **Providers** | ✨ **AI Features** |
|-------------|------------------|-------------------|
| `edit/write` files | **OpenAI** (GPT-5.2, 4o, o1) | Auto-summarization |
| `bash/git` shell | **Anthropic** (Claude 4.5 Opus/Sonnet) | Context-aware titles |
| `grep/glob/ls` search | **xAI** (Grok 4.1 Fast) | Loop/drift detection |
| **Z.AI Vision MCP** | **Mistral** (Devstral, Codestral) | Thread-safe queuing |
| `agent` (sub-agents) | **Local** (Ollama/LM-Studio) | **Vision analysis** |
| **Web Reader/Search** | **9+ APIs** | Image support |

## ⚙️ Production Setup (`make setup`)

**One-command config** with **9 real providers** + **Z.AI Vision MCP**:
```
✅ xAI Grok-4.1, Cerebras GLM-4.6, Anthropic Claude 4.5
✅ OpenAI GPT, Z.AI Vision MCP, MiniMax Kimi  
✅ Auto-loads .env API keys → Zero prompts → Production ready
```

## 📦 Installation Options

| Method | Command | Result |
|--------|---------|--------|
| **Quick** | `curl https://nexora.land/install.sh | sh` | `~/.local/bin/nexora` |
| **Build** | `make build && make setup` | Production binary + config |
| **Docker** | `docker run nexora/cli` | Containerized |

## 🧠 Providers (70+ Models)

```
🏆 Premium: OpenAI GPT-5.2 • Claude 4.5 Opus • Grok 4.1 Fast
🚀 Fast: Mistral Devstral • Cerebras GLM-4.6 • Gemini 3 Pro
🌐 Local: Ollama • LM-Studio • vLLM
🔗 MCP: **Z.AI Vision** • Web Reader/Search
```

## 🎮 Usage

```bash
# Chat mode
nexora chat

# One-shot tasks
nexora "Fix this Go bug in main.go"

# Vision analysis (upload image)
nexora chat  # → @z_ai/mcp-server vision analysis

# Multi-turn with tools
nexora chat  # → edit files, run bash, git commit, etc.
```

## 🛠️ Tools (20+ Built-in)

```
📝 Code: edit/write/multiedit/glob/grep/ls
🐚 Shell: bash/git status/diff
🔍 Search: sourcegraph/agentic_fetch/agent
🖼️ **Vision MCP**: @z_ai/mcp-server (image analysis)
🌐 Web: agentic_fetch/mcp_web-reader
📊 QA: job_output/job_kill
```

## 🚀 Why Nexora?

| ✅ **Production Ready** | ❌ **Avoid** |
|-----------------------|-------------|
| **Zero-config** (`make setup`) | Manual JSON editing |
| **9 API keys** + **Vision MCP** | Copy-paste configs |
| **All tests pass** (`make test-qa`) | Untested edge cases |
| **Thread-safe** agent queue | Race conditions |
| **Smart fallbacks** | Hard crashes |

## 📊 Benchmarks

```
⚡ Agent Speed: 150+ req/min (parallel tool calls)
🧠 Context: 1M+ tokens (auto-summarization)
🔄 Concurrency: 50+ queued requests
🛡️ Reliability: 99.9% (token validation + fallbacks)
```

## 🔬 ModelScan Tool

**Model validation CLI for testing AI provider APIs** - Directly validate provider endpoints, discover available models, and verify capabilities.

### Location & Setup
```bash
# Built-in tool (git-ignored)
cd ~/.local/tools/modelscan/
./modelscan --help
```

### Key Features
- **Direct API Validation**: Tests actual provider endpoints
- **Model Discovery**: Automatically discovers all available models
- **Capability Detection**: Identifies supported features per provider
- **Multiple Export Formats**: SQLite database + Markdown reports
- **Latency Measurement**: Tracks endpoint performance

### Usage Examples
```bash
# Validate all providers
./modelscan validate --all --verbose

# Test specific provider
./modelscan validate --provider=mistral --verbose

# Export results
./modelscan export --format=all --output=./results
```

### Output Files
- `providers.db` - SQLite database with validation history
- `PROVIDERS.md` - Human-readable provider capabilities report

### Configuration
```bash
# Set API keys
export MISTRAL_API_KEY="your-key"
export OPENAI_API_KEY="your-key"
export ANTHROPIC_API_KEY="your-key"
```

## 🤝 Contributing

```bash
git clone https://github.com/nexora/cli.git
cd cli
make build test-qa setup
nexora chat  # Test your changes!
```

## 📄 License

MIT © Nexora Team

---

**v0.28.6** - **Production hardened** with **bulletproof token validation**, **auto-updating titles**, **Z.AI Vision MCP**, and **one-command setup**.