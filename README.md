# Nexora - 🚀 AI-Powered CLI Agent

<div align="center">

[![Release](https://img.shields.io/github/v/release/jeffersonwarrior/nexora?style=for-the-badge&logo=github&color=blue)](https://github.com/jeffersonwarrior/nexora/releases)
[![CI/CD](https://img.shields.io/github/actions/workflow/status/jeffersonwarrior/nexora/ci.yml?branch=main&style=for-the-badge&logo=github)](.github/workflows/ci.yml)
[![Discord](https://img.shields.io/badge/Discord-Join%20Community-5865F2?style=for-the-badge&logo=discord&logoColor=white)](https://discord.gg/GCyC6qT79M)
[![Twitter/X](https://img.shields.io/badge/X-Follow%20Community-000000?style=for-the-badge&logo=x&logoColor=white)](https://x.com/i/communities/2004598673062216166/)
[![Reddit](https://img.shields.io/badge/Reddit-r%2FZackor-FF4500?style=for-the-badge&logo=reddit&logoColor=white)](https://www.reddit.com/r/Zackor/)

**Production-Ready AI Terminal Assistant** with **intelligent state management**, **adaptive resource monitoring**, and **self-healing execution**.

</div>

## ⚠️ Architecture Notice

**Nexora is the EXECUTION layer** - not the "thinking" layer.

- **YOLO mode by default**: No permission prompts, no safeguards
- **Designed for dedicated environments**: Run on isolated VM or dedicated hardware
- **Orchestration-ready**: Built to be controlled by higher-level systems (e.g., **Zackor**)
- **Direct execution**: AI commands run immediately without human approval

**⚡ This is intentional** - Nexora executes, orchestrators think.

## ✨ Quick Start

```bash
# Install (production config + 9 API providers + Vision MCP)
curl -fsSL https://nexora.land/install.sh | sh

# Or build from source
git clone https://github.com/jeffersonwarrior/nexora.git && cd nexora
make build && make setup

# First run launches TUI for setup
nexora chat  # Interactive TUI for API keys & provider configuration
```

## 🎯 Features

| 🛠️ **Tools** | 🔗 **Providers** | ✨ **AI Features** |
|-------------|------------------|-------------------|
| `edit/write` files | **OpenAI** (GPT-5.2, 4o, o1) | **Intelligent state management** |
| `bash/git` shell | **Anthropic** (Claude 4.5 Opus/Sonnet) | **Adaptive resource monitoring** |
| `grep/glob/ls` search | **xAI** (Grok 4.1 Fast) | Loop/drift detection |
| **Z.AI Vision MCP** | **Mistral** (Devstral, Codestral) | Progress tracking |
| `agent` (sub-agents) | **Local** (Ollama/LM-Studio) | Self-healing execution |
| **Web Reader/Search** | **9+ APIs** | Auto-summarization |

## ⚙️ Setup & Configuration

**Interactive TUI by default** - First run launches guided setup:
```bash
nexora chat  # Opens TUI for API key input & provider configuration
```

**Or use `make setup`** for one-command config with **9 real providers** + **Z.AI Vision MCP**:
```
✅ xAI Grok-4.1, Cerebras GLM-4.6, Anthropic Claude 4.5
✅ OpenAI GPT, Z.AI Vision MCP, MiniMax Kimi  
✅ Auto-loads .env API keys → Production ready
```

---

⚙️ See [CICD.md](CICD.md) for CI/CD pipeline documentation





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
🐚 Shell: bash/git status/diff (with TMUX session support)
🔍 Search: sourcegraph/agentic_fetch/agent
🖼️ **Vision MCP**: @z_ai/mcp-server (image analysis)
🌐 Web: fetch (smart routing) / web_fetch / web_search
📊 QA: job_output/job_kill (aliased to bash)
⚠️ **DELEGATE**: Not yet implemented (coming soon)
```

## 🔤 Natural Language Tool Aliases

Use natural language to invoke tools - 47 aliases supported:

| Tool | Aliases |
|------|---------|
| **fetch** | curl, wget, http-get, http_get, web-fetch, webfetch, web_fetch, http |
| **view** | read, cat, open |
| **ls** | dir, directory |
| **edit** | modify, change, replace, update |
| **write** | create, make, new |
| **grep** | search, find, rg |
| **bash** | shell, exec, execute, run, command |
| **web_search** | web-search, websearch, search-web |
| **sourcegraph** | sg, code-search |
| **job_kill** | 	 bash |
| **job_output** | 	 bash |

## 🌐 Smart Fetch

**Intelligent web content fetching** with context-aware handling:

### Features
- **MCP Auto-Routing**: Automatically uses MCP web_reader if available, falls back to built-in
- **Context-Aware**: Counts tokens in content, writes large content to tmp files if needed
- **Session-Scoped Storage**: Tmp files in `./tmp/nexora-{session-id}/`, auto-cleaned on session end
- **Format Support**: Returns content as text, markdown, or HTML

### Usage
```bash
# Natural language aliases supported
curl https://example.com          # 	 fetch
wget https://example.com          # 	 fetch
http-get https://example.com      # 	 fetch
web-fetch https://example.com     # 	 fetch
## 🤖 AI Code Review with Greptile

Nexora integrates with [Greptile](https://greptile.com) for automated AI code reviews on pull requests.

### Setup
1. Get your Greptile API key from the [Greptile Dashboard](https://greptile.com/dashboard)
2. Add the API key to your repository secrets:
   ```bash
   # In GitHub repository settings > Secrets and variables > Actions
   GREPTILE_API_KEY=your_api_key_here
   ```

### Features
- **Automatic PR Reviews**: Greptile analyzes pull requests for:
  - Code quality and best practices
  - Security vulnerabilities
  - Performance considerations
  - Documentation completeness
  - Test coverage
  - Adherence to Go conventions

- **Repository Indexing**: Automatically indexes your codebase for context-aware reviews
- **Customizable Rules**: Uses `greptile.json` for project-specific review rules

### Configuration
Configure review settings in your `greptile.json`:

```json
{
  "strictness": 3,
  "commentTypes": ["logic", "syntax", "style"],
  "model": "gpt-4o",
  "instructions": "Focus on production-grade code quality and adherence to Go patterns",
  "includeBranches": ["main", "feature/**", "fix/**"],
  "excludeAuthors": ["dependabot[bot]", "github-actions[bot]"]
}
```

### Usage
Once configured, Greptile will automatically review pull requests on:
- PR creation
- PR updates/new commits
- Manual workflow dispatch

## 🚀 Why Nexora?
|-----------------------|-------------|
| **YOLO mode**: Immediate execution | Human-in-the-loop workflows |
| **Orchestration-ready**: API-first design | Standalone end-user apps |
| **Dedicated environments**: VM/container isolation | Shared development machines |
| **Intelligent state management** | Ad-hoc execution |
| **Adaptive resource monitoring** | Runaway processes |

## 📊 Benchmarks

```
⚡ Agent Speed: 150+ req/min (parallel tool calls)
🧠 Context: 1M+ tokens (auto-summarization)
🔄 Concurrency: 50+ queued requests
🛡️ Reliability: 99.9% (token validation + fallbacks)
```

## 🎭 Architecture: Nexora vs Zackor

**Nexora** = **Execution Layer**
- Direct CLI execution with AI
- YOLO mode (no safeguards)
- State management & resource monitoring
- Runs on dedicated/isolated environments

**Zackor** = **Orchestration Layer** _(coming soon)_
- High-level planning & strategy
- Multi-agent coordination
- Safety policies & approval workflows
- Manages multiple Nexora instances

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
git clone https://github.com/jeffersonwarrior/nexora.git
cd nexora
make build test-qa setup
nexora chat  # Test your changes!
```

## 📄 License

MIT © Nexora Team

## 👨‍💻 Credits

Built by **Jefferson Nunn** with the help of:
- Claude Opus, Sonnet, Haiku
- Synthetic, GLM, Kimi
- GPT, OSS, Cerebras GLM

---

**v0.29.3** - **Production-ready CLI** with **version display**, **about command**, **task/checkpoint system**, and **unified command palette**.
## 📺 TMUX Integration

Nexora supports **persistent TMUX sessions** for AI-driven interactive terminal workflows.

- Interactive editor control (vi, helix, emacs)
- Human observation and real-time intervention
- Multi-model orchestration (Opus + Sonnet + Haiku)

See [TMUX.md](TMUX.md) for the full protocol documentation.
