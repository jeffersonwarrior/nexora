# Nexora Codebase Documentation

|**Version:** v0.29.3  
|**Date:** 2025-12-29  
|**Status:** Production Ready|

---

## Project Overview

Nexora is an **AI-powered CLI agent** designed as an execution layer for autonomous agent orchestration. It provides intelligent file editing, shell command execution, web fetching, and multi-provider AI model support.

### Key Characteristics
- **YOLO Mode by Default**: No permission prompts, designed for dedicated/isolated environments
- **Orchestration-Ready**: Built to be controlled by higher-level systems (e.g., Zackor)
- **70+ Models Supported**: OpenAI, Anthropic, xAI, Mistral, Google Gemini, Local (Ollama/LM-Studio)
- **25+ Built-in Tools**: edit, write, bash, grep, glob, fetch, agent, delegate, and more
- **TMUX Integration**: Persistent shell sessions for interactive workflows
- **VCR Testing**: Deterministic test recordings for reliable CI/CD

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
├── internal/                  # Core implementation (37 modules)
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
│   ├── aiops/                 # AI operations utilities
│   ├── ansiext/               # ANSI escape sequences
│   ├── app/                   # Application core
│   ├── csync/                 # Concurrent synchronization
│   ├── diff/                  # Diff utilities
│   ├── env/                   # Environment utilities
│   ├── filepathext/           # File path extensions
│   ├── format/                # Formatting utilities
│   ├── fsext/                 # Filesystem extensions
│   ├── history/               # History management
│   ├── home/                  # Home directory utilities
│   ├── log/                   # Logging
│   ├── message/               # Message handling
│   ├── permission/            # Permission management
│   ├── sessionlog/            # Session logging
│   ├── stringext/             # String extensions
│   ├── term/                  # Terminal utilities
│   ├── testutil/              # Testing utilities
│   ├── update/                # Update management
│   └── [utilities]/           # Additional utility packages
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

|**File**|**Purpose**|
|--------|---------|
| `agent.go` | Main agent loop, message processing, tool dispatch (63KB) |
| `coordinator.go` | Sequential edit coordination, EditLocation and SequentialEditSolver types |
| `delegate_tool.go` | Sub-agent delegation with resource pooling and enhanced validation |
| `agent_tool.go` | Agent spawning and management |
| `summarizer.go` | Context summarization for long conversations |
| `prompts.go` | Prompt template management |
| `conversation_loop.go` | Main conversation handling |
| `conversation_state.go` | Conversation state management |
| `background_compaction.go` | Background compaction for memory optimization |
| `compaction.go` | Message compaction for context management |
| `multi_session_coordinator.go` | Multi-session coordination (TODO: incomplete) |
| `agentic_fetch_tool.go` | Agentic web fetching with AI analysis |
| `task_coordinator.go` | Task coordination |
| `session_window.go` | Session window management |
| `event.go` | Event handling |
| `errors.go` | Error definitions |

|**State Management:**||
|---|---|
| `state/machine.go` | Finite state machine for execution phases |
| `state/progress.go` | Progress tracking and loop detection |
| `state/context.go` | Context window management |

|**Sub-Agents Module:**||
|---|---|
| `agents/agent_executor.go` | Sub-agent execution |
| `agents/agent_parser.go` | Agent response parsing |
| `agents/integration.go` | Agent integration |

|**Delegation Module (NEW):**||
|---|---|
| `delegation/pool.go` | Resource-aware concurrent agent delegation with pooling |
| `delegation/pool_test.go` | Pool functionality tests |

|**Recovery Module:**||
|---|---|
| `recovery/registry.go` | Error recovery registry |
| `recovery/strategy.go` | Recovery strategies |
| `recovery/file_outdated.go` | File outdated detection |

|**Metrics Module:**||
|---|---|
| `metrics/metrics.go` | Agent metrics collection |
| `metrics/hook.go` | Metrics hooks |

|**Prompt Module:**||
|---|---|
| `prompt/prompt.go` | Prompt management |
| `prompt/cache.go` | Prompt caching |
| `prompt/cache_test.go` | Cache tests |

### 2. Tools (`internal/agent/tools/`)

25+ built-in tools for agent operations.

|**Core Tools:**||
|---|---|
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
| ls | `ls.go` | Directory listing |
| find | `find.go` | File finding |
| fuzzy_match | `fuzzy_match.go` | Fuzzy matching |
| job_output | `job_output.go` | Background job output |
| job_kill | `job_kill.go` | Kill background jobs |
| web_search | `web_search.go` | Web search (MCP) |
| smart_edit | `smart_edit.go` | AI-assisted smart editing |
| self_heal | `self_heal.go` | Self-healing capabilities |
| safe | `safe.go` | Safety utilities |
| aliases | `aliases.go` | Natural language aliases (47+ supported) |
| output_manager | `output_manager.go` | Context-aware output handling |
| impact_analysis | `impact_analysis.go` | Edit impact analysis |
| diagnostics | `diagnostics.go` | Edit validation |
| fetch_helpers | `fetch_helpers.go` | Fetch utilities |
| fetch_types | `fetch_types.go` | Fetch type definitions |
| error_messages | `error_messages.go` | Error message handling |
| file | `file.go` | File utilities |
| fd_helper | `fd_helper.go` | File descriptor helpers |
| hooks | `hooks.go` | Tool hooks |
| install_manager | `install_manager.go` | Installation management |
| prompts | `prompts.go` | Tool prompts |
| references | `references.go` | Reference tracking |
| reliability | `reliability.go` | Reliability utilities |
| rg | `rg.go` | Ripgrep integration |
| search | `search.go` | Search interface |
| search_indexed | `search_indexed.go` | Indexed search |
| temp_dir | `temp_dir.go` | Temporary directory management |
| tools | `tools.go` | Tools interface |

|**MCP Tools Module:**||
|---|---|
| `mcp/tools.go` | MCP tool implementations |
| `mcp/init.go` | MCP initialization |
| `mcp/prompts.go` | MCP prompts |
| `mcp/reliability.go` | MCP reliability |

---

### 3. Configuration (`internal/config/`)

|**File**|**Purpose**|
|--------|---------|
| `config.go` | Main configuration structure |
| `load.go` | Configuration file loading |
| `providers/` | AI provider implementations |
| `provider.go` | Provider selection and fallback |
| `resolve.go` | Model resolution |

|**Supported Providers:**||
|----------|------|--------|
| Provider | File | Models |
|----------|------|--------|
| OpenAI | `providers/openai.go` | GPT-5.2, GPT-4o, o1, o1-mini, o3-mini |
| Anthropic | `providers/anthropic.go` | Claude 4.5 Opus, Sonnet, Haiku |
| xAI | `providers/xai.go` | Grok 4.1, Grok-2 |
| Mistral | `providers/mistral_*.go` | Devstral, Codestral, Native |
| Gemini | `providers/gemini.go` | Gemini 3 Pro, 2.0 Flash |
| MiniMax | `providers/minimax.go` | Kimi, MiniMax |
| Z.AI | `providers/zai.go` | GLM-4.6 (MCP) |
| Cerebras | `providers/cerebras.go` | Cerebras models |
| Kimi | `providers/kimi.go` | Kimi models |
| OpenRouter | `providers/openrouter.go` | Various models |
| Local | `providers/local_detector.go` | Ollama, LM-Studio |
| DeepSeek | `providers/deepseek.go` | DeepSeek models |
| Qwen | `providers/qwen.go` | Qwen models |

---

### 4. Terminal UI (`internal/tui/`)

Bubble Tea-based interactive terminal UI.

|**Component**|**Purpose**|
|------------|---------|
| `tui.go` | Main TUI loop (21KB) |
| `keys.go` | Key bindings |
| `components/chat/` | Chat interface, messages, editor |
| `components/dialogs/` | Settings, models, commands, sessions |
| `components/core/` | Status bar, layout |
| `components/page/` | Page components |
| `exp/diffview/` | Side-by-side diff viewer |
| `styles/` | Theming and styling |
| `util/` | TUI utilities |
| `highlight/` | Syntax highlighting |

---

### 5. Database (`internal/db/`)

SQLite database for session persistence.

|**File**|**Purpose**|
|--------|---------|
| `db.go` | Database connection and migrations |
| `models.go` | Data models (Session, Message, File, Checkpoint) |
| `sessions.sql.go` | Session queries (sqlc generated) |
| `messages.sql.go` | Message queries (sqlc generated) |
| `checkpoints.sql.go` | Checkpoint queries (sqlc generated) |

---

### 6. Session Management (`internal/session/`)

|**File**|**Purpose**|
|--------|---------|
| `session.go` | Session state management |
| `checkpoint.go` | Session checkpoint for crash recovery |

---

### 7. Task System (`internal/task/`)

|**File**|**Purpose**|
|--------|---------|
| `manager.go` | Task queue and execution |
| `coordinator.go` | Task graph coordination |
| `graph.go` | Dependency graph for tasks |
| `agent_tool.go` | Task agent integration |

---

### 8. Shell & TMUX (`internal/shell/`)

|**File**|**Purpose**|
|--------|---------|
| `shell.go` | Shell execution with safety |
| `tmux.go` | TMUX session management |
| `background.go` | Background process handling |
| `command_block_test.go` | Command blocking tests |

---

### 9. Code Indexer (`internal/indexer/`)

|**File**|**Purpose**|
|--------|---------|
| `indexer.go` | Code indexing service |
| `graph.go` | Code dependency graph |
| `ast_parser.go` | AST-based parsing |
| `embeddings.go` | Vector embeddings |
| `storage.go` | SQLite storage |

---

### 10. MCP Protocol (`internal/mcp/`)

|**File**|**Purpose**|
|--------|---------|
| `client.go` | MCP client implementation |
| `server.go` | MCP server implementation |
| `protocol.go` | Protocol definitions |

---

### 11. Delegation Pool (`internal/agent/delegation/`)

Resource-aware agent delegation with pooling for concurrent sub-agent execution.

|**File**|**Purpose**|
|--------|---------|
| `pool.go` | Pool manager for concurrent delegate agents with resource monitoring (11KB) |
| `pool_test.go` | Pool functionality tests (16KB) |

|**Key Types:**|
|
- `Pool`: Manages concurrent delegate agents with resource awareness
- `PoolConfig`: Configuration for pool behavior (max concurrent, timeouts, resource limits)
- `Task`: Represents a delegated task with status tracking

---

### 12. LSP Support (`internal/lsp/`)

|**File**|**Purpose**|
|--------|---------|
| `lsp.go` | Language server protocol implementation |
| `server.go` | LSP server |
| `client.go` | LSP client |

---

### 13. OAuth (`internal/oauth/`)

|**File**|**Purpose**|
|--------|---------|
| `oauth.go` | OAuth authentication |
| `providers.go` | OAuth providers |

---

### 14. Message Handling (`internal/message/`)

|**File**|**Purpose**|
|--------|---------|
| `message.go` | Message types and handling |
| `queue.go` | Message queue |
| `processor.go` | Message processing |

---

### 15. Format & Diff (`internal/format/`, `internal/diff/`)

|**File**|**Purpose**|
|--------|---------|
| `format.go` | Formatting utilities |
| `diff.go` | Diff computation |
| `unified.go` | Unified diff format |

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

|**Dependency**|**Purpose**|
|-------------|---------|
| `charmbracelet/bubbletea/v2` | Terminal UI framework |
| `charm.land/bubbles/v2` | UI components |
| `github.com/openai/openai-go/v2` | OpenAI API |
| `github.com/charmbracelet/anthropic-sdk-go` | Anthropic API |
| `github.com/sahilm/fuzzy` | Fuzzy matching |
| `github.com/mattn/go-isatty` | Terminal detection |
| `github.com/stretchr/testify` | Testing utilities |
| `github.com/google/uuid` | UUID generation |
| `github.com/ncruces/go-sqlite3` | SQLite driver |
| `charm.land/x/vcr` | VCR test recordings |
| `github.com/alecthomas/chroma/v2` | Syntax highlighting |
| `github.com/aymanbagabas/go-udiff` | Diff utilities |

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

# VCR Testing
./scripts/record-vcr.sh       # Record VCR cassettes
./scripts/run-tests-with-limits.sh  # Run tests with limits
```

---

## Configuration Files

|**File**|**Purpose**|
|--------|---------|
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

|**Document**|**Purpose**|
|----------|---------|
| `TMUX.md` | TMUX integration protocol |
| `CLAUDE.md` | Claude AI instructions |
| `NEXORA.md` | Nexora system prompt |
| `AGENTS.md` | Agent-specific instructions |
| `codedocs/BASH-SAFETY-AUDIT.md` | Bash security guide |
| `codedocs/TMUX-INTERACTION-PROTOCOL.md` | TMUX workflows |
| `docs/adr/*.md` | Architecture decisions |
| `docs/AI_AUTOFIX_GUIDE.md` | AI autofix guide |
| `docs/CONTEXT_MANAGEMENT_V2.md` | Context management |
| `docs/VCR_RECORDING.md` | VCR testing guide |

---

## Recent Changes (v0.29.3)

### Agent Module Updates
- **coordinator.go**: Refactored for sequential edit coordination with new `SequentialEditSolver` and `EditLocation` types (33KB)
- **delegate_tool.go**: Enhanced delegation with new `DelegateParams`, `DelegatePermissionsParams`, and `delegateValidationResult` structures; improved validation logic (13KB)
- **delegation/pool.go**: New module for resource-aware concurrent agent delegation with dynamic sizing based on system resources (11KB)
- **compaction.go**: New compaction system for message optimization (11KB)
- **background_compaction.go**: Background compaction for memory optimization (9KB)
- **delegate_prompt.md.tpl**: Updated template with enhanced guidelines for focused sub-agent execution

### Configuration Updates
- Multiple provider implementations updated for consistency (OpenAI, Anthropic, xAI, Mistral, Gemini, MiniMax, Z.AI, Cerebras, Kimi, OpenRouter, DeepSeek, Qwen)
- Configuration loading and validation improvements
- New provider implementations: OpenRouter, DeepSeek, Qwen

### TUI Updates
- **tui.go**: Major refactoring with 21KB implementation
- New dialog components and improved styling
- Enhanced diff viewer in `exp/diffview/`
- New `page/` component module

### Tools Updates
- **smart_edit.go**: AI-assisted smart editing enhancements
- **self_heal.go**: Self-healing capabilities
- **output_manager.go**: Context-aware output handling
- **mcp/tools.go**: MCP tool implementations
- New tools: `fd_helper.go`, `install_manager.go`, `prompts.go`, `temp_dir.go`, `references.go`

### Test Data Updates
- Test fixtures updated for various model providers (OpenAI GPT-4o, Anthropic Sonnet, Z.AI GLM-4.6, OpenRouter Kimi-K2)
- VCR cassettes for deterministic testing
- New test coverage for delegation pool

### Git Commits (Recent 10)
```
0ba0ed5 fix: delegate output extraction and VCR cassette refresh
85caad4 test: add coverage for tool call pairing in dropToolResults compaction
c3ec3d7 fix: prevent orphaned tool results in MiniMax compaction
1cd8a7d fix: session title race, TMUX output duplication, compaction tool pairing
1cd8a7d (tag: v0.29.3) release: v0.29.3 - Production Polish & Advanced Features
7cb4393 fix: rebrand user-facing Claude Code references to Anthropic
f09eb16 fix: update system prompt from Claude Code to Nexora
1c881ae feat: add CLI commands for tasks and checkpoints
da760ee feat: implement TMUX session pooling (fixes #18)
9239ca5 test: add session leak detection tests
```

---

## Known TODOs and Future Work

### Agent Module
- `coordinator.go`: TODO - make session management dynamic when supporting multiple agents
- `multi_session_coordinator.go`: TODO - file is incomplete and needs to be finished
- Future enhancements: execution-first prompting, incremental execution pipeline, self-correction loops, tool-chain orchestration

---

*Last Updated: 2025-12-29*
*Version: v0.29.3*
*Total Go Files: 300+*
*Test Coverage: Comprehensive with VCR cassettes*
