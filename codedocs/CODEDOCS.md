# Nexora Codebase Documentation

**Version:** v0.29.3  
**Date:** 2025-12-28  
**Status:** Production Ready

---

## Project Overview

Nexora is an **AI-powered CLI agent** designed as an execution layer for autonomous agent orchestration. It provides intelligent file editing, shell command execution, web fetching, and multi-provider AI model support.

### Key Characteristics
- **YOLO Mode by Default**: No permission prompts, designed for dedicated/isolated environments
- **Orchestration-Ready**: Built to be controlled by higher-level systems (e.g., Zackor)
- **70+ Models Supported**: OpenAI, Anthropic, xAI, Mistral, Google Gemini, Local (Ollama/LM-Studio)
- **20+ Built-in Tools**: edit, write, bash, grep, glob, fetch, agent, delegate, and more
- **TMUX Integration**: Persistent shell sessions for interactive workflows

---

## Directory Structure

```
/
├── main.go                     # Entry point
├── go.mod / go.sum            # Go module dependencies
├── Makefile                    # Build and development commands
├── Taskfile.yaml              # Task automation
├── package.json               # Node.js dependencies (build tools)
├── install.sh                 # Installation script
├── README.md                  # Project documentation
├── CHANGELOG.md               # Release notes
├── CLAUDE.md                  # Claude AI instructions
├── NEXORA.md                  # Nexora system prompt
├── AGENTS.md                  # Agent-specific instructions
├── LICENSE.md                 # MIT License
├── SECURITY.md                # Security policy
├── ROADMAP.md                 # Project roadmap
├── TMUX.md                    # TMUX integration guide
├── schema.json                # JSON schema for configuration
│
├── internal/                  # Core implementation
│   ├── agent/                 # AI agent implementation
│   ├── cmd/                   # CLI commands
│   ├── config/                # Configuration management
│   ├── tui/                   # Terminal UI (bubble tea)
│   ├── db/                    # Database (SQLite)
│   ├── session/               # Session management
│   ├── task/                  # Task coordination
│   ├── shell/                 # Shell execution
│   ├── indexer/               # Code indexing
│   ├── mcp/                   # MCP protocol support
│   ├── lsp/                   # Language server protocol
│   ├── oauth/                 # OAuth authentication
│   ├── pubsub/                # Event publishing
│   ├── resources/             # Resource monitoring
│   ├── version/               # Version management
│   └── [utilities]/           # Utility packages
│
├── sdk/                       # SDK for AI providers
│   ├── base/                  # Base client interfaces
│   ├── openai/                # OpenAI provider implementation
│   └── mistral/               # Mistral provider implementation
│
├── scripts/                   # Build and development scripts
├── docs/                      # Documentation and ADRs
│   └── adr/                   # Architecture Decision Records
│
└── codedocs/                  # Codebase documentation
    ├── BASH-SAFETY-AUDIT.md   # Bash security documentation
    ├── TMUX-INTERACTION-PROTOCOL.md  # TMUX integration guide
    ├── agentic/               # Agentic features documentation
    └── phase4/                # Phase 4 execution logs
```

---

## Core Modules

### 1. Agent (`internal/agent/`)

The core AI agent that processes messages and executes tools.

| File | Purpose |
|------|---------|
| `agent.go` | Main agent loop, message processing, tool dispatch |
| `coordinator.go` | Multi-session coordination, resource limits |
| `delegate_tool.go` | Sub-agent delegation with resource pooling |
| `agent_tool.go` | Agent spawning and management |
| `summarizer.go` | Context summarization for long conversations |
| `prompts.go` | Prompt template management |
| `conversation_loop.go` | Main conversation handling |
| `recovery/` | Error recovery and self-healing |

**State Management:**
| File | Purpose |
|------|---------|
| `state/machine.go` | Finite state machine for execution phases |
| `state/progress.go` | Progress tracking and loop detection |
| `state/context.go` | Context window management |

### 2. Tools (`internal/agent/tools/`)

20+ built-in tools for agent operations.

**Core Tools:**
| Tool | File | Purpose |
|------|------|---------|
| view | `view.go` | Read files (100-line chunks by default) |
| edit | `edit.go` | Edit files (AI-mode available) |
| write | `write.go` | Create/overwrite files |
| multiedit | `multiedit.go` | Multi-file editing |
| bash | `bash.go` | Shell command execution (TMUX support) |
| glob | `glob.go` | File pattern matching |
| grep | `grep.go` | Content search |
| fetch | `fetch.go` | Web content fetching (smart routing) |
| download | `download.go` | File downloads |
| sourcegraph | `sourcegraph.go` | Code search |
| agent | `agents.go` | Sub-agent spawning |
| delegate | `delegate_tool.go` | Resource-pooled delegation |

**Tool Support Files:**
| File | Purpose |
|------|---------|
| `aliases.go` | Natural language aliases (47 supported) |
| `output_manager.go` | Context-aware output handling |
| `bash_safety_test.go` | 35 safety tests for bash tool |
| `impact_analysis.go` | Edit impact analysis |

### 3. Configuration (`internal/config/`)

| File | Purpose |
|------|---------|
| `config.go` | Main configuration structure |
| `load.go` | Configuration file loading |
| `providers/` | AI provider implementations |
| `provider.go` | Provider selection and fallback |
| `resolve.go` | Model resolution |

**Supported Providers:**
| Provider | File | Models |
|----------|------|--------|
| OpenAI | `providers/openai.go` | GPT-5.2, GPT-4o, o1 |
| Anthropic | `providers/anthropic.go` | Claude 4.5 Opus, Sonnet |
| xAI | `providers/xai.go` | Grok 4.1 |
| Mistral | `providers/mistral_*.go` | Devstral, Codestral |
| Gemini | `providers/gemini.go` | Gemini 3 Pro |
| MiniMax | `providers/minimax.go` | Kimi |
| Z.AI | `providers/zai.go` | GLM-4.6 (MCP) |
| Local | `providers/local_detector.go` | Ollama, LM-Studio |

### 4. Terminal UI (`internal/tui/`)

Bubble Tea-based interactive terminal UI.

| Component | Purpose |
|-----------|---------|
| `tui.go` | Main TUI loop |
| `components/chat/` | Chat interface, messages, editor |
| `components/dialogs/` | Settings, models, commands, sessions |
| `components/core/` | Status bar, layout |
| `exp/diffview/` | Side-by-side diff viewer |
| `styles/` | Theming and styling |

### 5. Database (`internal/db/`)

SQLite database for session persistence.

| File | Purpose |
|------|---------|
| `db.go` | Database connection and migrations |
| `models.go` | Data models (Session, Message, File, Checkpoint) |
| `sessions.sql.go` | Session queries (sqlc generated) |
| `messages.sql.go` | Message queries (sqlc generated) |
| `checkpoints.sql.go` | Checkpoint queries (sqlc generated) |

### 6. Session Management (`internal/session/`)

| File | Purpose |
|------|---------|
| `session.go` | Session state management |
| `checkpoint.go` | Session checkpoint for crash recovery |

### 7. Task System (`internal/task/`)

| File | Purpose |
|------|---------|
| `manager.go` | Task queue and execution |
| `coordinator.go` | Task graph coordination |
| `graph.go` | Dependency graph for tasks |
| `agent_tool.go` | Task agent integration |

### 8. Shell & TMUX (`internal/shell/`)

| File | Purpose |
|------|---------|
| `shell.go` | Shell execution with safety |
| `tmux.go` | TMUX session management |
| `background.go` | Background process handling |
| `command_block_test.go` | Command blocking tests |

### 9. Code Indexer (`internal/indexer/`)

| File | Purpose |
|------|---------|
| `indexer.go` | Code indexing service |
| `graph.go` | Code dependency graph |
| `ast_parser.go` | AST-based parsing |
| `embeddings.go` | Vector embeddings |
| `storage.go` | SQLite storage |

---

## SDK Modules

### OpenAI SDK (`sdk/openai/`)
```go
- client.go     # OpenAI API client
- chat.go       # Chat completions
- model.go      # Model support
```

### Mistral SDK (`sdk/mistral/`)
```go
- client.go     # Mistral API client
- chat.go       # Chat completions
- types.go      # Type definitions
```

### Base SDK (`sdk/base/`)
```go
- client.go     # Abstract client interface
- types.go      # Common type definitions
```

---

## Key Dependencies

| Dependency | Purpose |
|------------|---------|
| `charmbracelet/bubbletea/v2` | Terminal UI framework |
| `charm.land/bubbles/v2` | UI components |
| `github.com/openai/openai-go/v2` | OpenAI API |
| `github.com/charmbracelet/anthropic-sdk-go` | Anthropic API |
| `github.com/sahilm/fuzzy` | Fuzzy matching |
| `github.com/mattn/go-isatty` | Terminal detection |
| `github.com/stretchr/testify` | Testing utilities |
| `github.com/google/uuid` | UUID generation |
| `github.com/ncruces/go-sqlite3` | SQLite driver |

---

## Build & Test Commands

```bash
# Build
make build                    # Build binary
make setup                    # Setup with 9 providers

# Test
make test                     # Run all tests
make test-qa                  # QA test suite
go test ./... -race           # Race detection
go test ./... -coverprofile   # Coverage report

# Lint
make lint                     # Run linters
gofumpt -w .                  # Format code

# Development
make dev                      # Development mode
```

---

## Configuration Files

| File | Purpose |
|------|---------|
| `~/.config/nexora/nexora.json` | User configuration |
| `.env` | API keys (git-ignored) |
| `schema.json` | Config JSON schema |

---

## GitHub Release Checklist

- [x] Version bumped in `internal/version/version.go`
- [x] CHANGELOG.md updated with release notes
- [x] README.md updated with current features
- [x] LICENSE.md present and correct
- [x] SECURITY.md security policy defined
- [x] All tests passing
- [x] Build successful
- [x] No temporary files in repository
- [x] No API keys or secrets committed
- [x] .gitignore properly configured

---

## Additional Documentation

| Document | Purpose |
|----------|---------|
| `TMUX.md` | TMUX integration protocol |
| `CLAUDE.md` | Claude AI instructions |
| `NEXORA.md` | Nexora system prompt |
| `AGENTS.md` | Agent-specific instructions |
| `codedocs/BASH-SAFETY-AUDIT.md` | Bash security guide |
| `codedocs/TMUX-INTERACTION-PROTOCOL.md` | TMUX workflows |
| `docs/adr/*.md` | Architecture decisions |

---

*Last Updated: 2025-12-28*
