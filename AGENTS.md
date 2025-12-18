# Nexora Agent Development Guide

This guide helps AI agents effectively work in the Nexora codebase. Nexora is a terminal-based AI assistant written in Go that provides multi-model LLM support, session management, LSP integration, and extensible MCP capabilities.

## Essential Commands

### Build & Run

```bash
# Quick build
go build .

# Run directly
go run .

# Run with profiling enabled (CPU, heap, etc.)
go run . -NEXORA_PROFILE=true
# Then access pprof at localhost:6060/debug/pprof

# Build and install to PATH
go install .

# Verbose build
go build -v .
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run specific package
go test ./internal/agent -v

# Run single test
go test ./internal/agent -run TestCoderAgent

# Run tests and update golden files (for snapshot tests)
go test ./... -update

# Update golden files for specific package
go test ./internal/tui/components/core -update

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Code Quality

```bash
# Complete quality check (recommended before commits)
./scripts/lint.sh

# Format code with gofumpt (stricter than gofmt)
gofumpt -w .

# Run linters only
golangci-lint run --config=".golangci.yml" --timeout=5m

# Fix linting issues automatically
golangci-lint run --config=".golangci.yml" --timeout=5m --fix

# Format code via task
task fmt
```

### Task Runner (Taskfile.yaml)

Nexora uses `task` (https://taskfile.dev) for common operations:

```bash
task lint              # Run base linters
task lint:fix          # Run linters and fix issues
task lint:install      # Install golangci-lint
task build             # Compile the project
task test              # Run test suite
task dev               # Run with profiling enabled
task fmt               # Format with gofumpt
task install           # Install to PATH with version info
task profile:cpu       # 10s CPU profile
task profile:heap      # Heap memory profile
task profile:allocs    # Allocations profile
task schema            # Generate JSON schema for config
task release           # Create and push new semver tag
```

## Project Structure

### Directory Organization

```
nexora/
├── internal/                  # Private application packages
│   ├── agent/                 # LLM agent orchestration (largest, most complex)
│   │   ├── agent.go          # Main SessionAgent logic
│   │   ├── coordinator.go    # Agent loop and streaming
│   │   ├── tools/            # Tool implementations (view, edit, bash, etc.)
│   │   ├── prompt/           # Prompt engineering templates
│   │   ├── metrics/          # Tool execution metrics
│   │   └── templates/        # Embedded markdown prompts
│   ├── tui/                   # Terminal UI (Bubble Tea based)
│   │   ├── components/       # Reusable UI components
│   │   ├── page/             # Full-page components
│   │   ├── styles/           # Lipgloss styling
│   │   └── exp/              # Experimental UI elements
│   ├── config/                # Provider and model configuration
│   ├── message/               # Message storage and conversion
│   ├── session/               # Session management
│   ├── db/                    # Database layer
│   ├── lsp/                   # Language Server Protocol client
│   ├── permission/            # Permission system for tools
│   ├── cmd/                   # CLI command handlers
│   ├── indexer/               # Code indexing
│   ├── aiops/                 # AI operations (loop detection, drift)
│   └── [other packages]       # Utilities, extensions, etc.
├── scripts/                   # Build and maintenance scripts
│   ├── lint.sh               # Comprehensive quality checks
│   └── run-labeler.sh        # GitHub PR labeling
├── main.go                    # Entry point
├── go.mod / go.sum           # Dependencies
├── .golangci.yml             # Linter configuration
├── Taskfile.yaml             # Task definitions
├── NEXORA.md                  # Architecture and code style guide
├── PROJECT_OPERATIONS.md     # Operational procedures
├── CODEDOCS.md               # Code reference index
└── README.md                 # Project overview
```

### Key Packages

#### `internal/agent/`
The heart of Nexora. Contains:
- `SessionAgent`: Main agent loop that processes prompts and executes tools
- `Coordinator`: Manages agent streaming responses
- Tool implementations (view, edit, bash, fetch, etc.)
- Prompt templates for system messages, titles, summaries
- AIOPS for detecting tool loops and task drift

#### `internal/tui/`
Bubble Tea-based terminal UI:
- `Chat` page with message list, input, sidebar
- `Dialogs` for model selection, settings
- Reusable `Components` (lists, inputs, buttons)
- Responsive layout system

#### `internal/config/`
Provider and model configuration:
- Multi-model support (OpenAI, Anthropic, Google, Bedrock, OpenRouter, local)
- Provider initialization via `charm.land/fantasy`
- Mock providers for testing

#### `internal/message/`
Message abstraction layer:
- Converts between internal format and LLM provider APIs
- Handles tool calls and results
- Supports reasoning, text, images, binary content
- JSON serialization for persistence

## Code Organization & Conventions

### Package Structure

1. **Single responsibility**: Each package handles one concern
2. **Interfaces in consumers**: Define interfaces in packages that use them, not in packages being used
3. **Unexported by default**: Only export what's needed in package names; use lowercase for internal

Example pattern:
```go
// internal/message/content.go - defines types used elsewhere
type Message struct { ... }

// internal/agent/agent.go - defines interface it needs
type MessageService interface {
    Create(ctx context.Context, sessionID string, params CreateParams) error
    List(ctx context.Context, sessionID string) ([]Message, error)
}
```

### Naming Conventions

- **PascalCase**: Exported types, functions, constants (`type Message struct`, `func (m *Message) Create()`)
- **camelCase**: Unexported identifiers (`func (m *Message) processContent()`)
- **Type aliases for clarity**: Use explicit names for custom types
  ```go
  type AgentName string
  type SessionID string
  ```
- **Enums with iota**: Group related constants
  ```go
  const (
      MessageRoleUser MessageRole = "user"
      MessageRoleAssistant MessageRole = "assistant"
      MessageRoleTool MessageRole = "tool"
  )
  ```

### Imports Organization

Group imports in this order:
1. Standard library (`fmt`, `os`, `context`, etc.)
2. Third-party packages (`charm.land/`, `github.com/charmbracelet/`, etc.)
3. Internal packages (`github.com/nexora/nexora/internal/...`)

Example:
```go
import (
    "context"
    "fmt"
    "strings"
    
    "charm.land/fantasy"
    "github.com/charmbracelet/log"
    
    "github.com/nexora/nexora/internal/agent"
    "github.com/nexora/nexora/internal/message"
)
```

### Error Handling

- Return errors explicitly using `fmt.Errorf` with context
- Wrap errors with `%w` for error chain preservation
- Use `errors.Is()` and `errors.As()` for error checking

Example:
```go
if err := service.Create(ctx, params); err != nil {
    return fmt.Errorf("failed to create message: %w", err)
}
```

### Context Usage

- Always pass `context.Context` as the first parameter for operations
- Use `ctx.Context()` on `*testing.T` to get test context in tests
- Respect context cancellation in loops and long operations

### Struct Composition

Use embedding for composition:
```go
type ExtendedAgent struct {
    *Agent  // Embed agent, gains its methods
    extra   string
}
```

## Code Style Guidelines

### Formatting

**CRITICAL**: Format all Go code before committing.

```bash
# Primary: Use gofumpt (stricter than gofmt)
gofumpt -w .

# Fallback: If gofumpt unavailable, use goimports
goimports -w .

# Fallback: If goimports unavailable, use gofmt
gofmt -w .

# Via task
task fmt
```

The `.golangci.yml` config enforces `gofumpt` formatting via linter.

### Comments

- **Line comments**: Start with capital letter, end with period
- **Wrap at 78 columns**: Break long comments into multiple lines
- **Document exported symbols**: Every exported type, function, const should have a comment
- **Avoid inline comments**: Prefer explaining in surrounding code structure

Example:
```go
// Message represents a single message in conversation history.
// It can contain text, tool calls, or tool results.
type Message struct {
    ID   string
    Role MessageRole
}

// Create persists a new message to the database.
func (s *Service) Create(ctx context.Context, params CreateParams) error {
    // Validate parameters first
    if err := params.Validate(); err != nil {
        return fmt.Errorf("invalid params: %w", err)
    }
    // ...
}
```

### Constants & Enums

Use typed constants with iota for enums:
```go
type FinishReason string

const (
    FinishReasonEndTurn     FinishReason = "end_turn"
    FinishReasonMaxTokens   FinishReason = "max_tokens"
    FinishReasonToolUse     FinishReason = "tool_use"
    FinishReasonCanceled    FinishReason = "canceled"
    FinishReasonError       FinishReason = "error"
    FinishReasonUnknown     FinishReason = "unknown"
)
```

### JSON Tags

Use `snake_case` for JSON field names:
```go
type Config struct {
    ModelName   string `json:"model_name"`
    MaxTokens   int    `json:"max_tokens"`
    Temperature float64 `json:"temperature,omitempty"`
}
```

### File Permissions

Use octal notation:
```go
os.WriteFile(path, data, 0o644)  // Regular file
os.Mkdir(path, 0o755)             // Directory
```

### Interfaces

Keep interfaces small and focused:
```go
// Good: Single responsibility
type MessageService interface {
    Create(ctx context.Context, params CreateParams) error
    List(ctx context.Context, sessionID string) ([]Message, error)
}

// Bad: Too many methods
type MessageService interface {
    Create(...) error
    List(...) ([]Message, error)
    Update(...) error
    Delete(...) error
    Archive(...) error
    Export(...) error
    Import(...) error
}
```

## Testing

### Test Patterns

Nexora uses `testify/require` for assertions and parallel testing:

```go
import "github.com/stretchr/testify/require"

func TestMyFunction(t *testing.T) {
    t.Parallel()  // Run in parallel with other tests
    
    result, err := MyFunction(context.Background(), input)
    require.NoError(t, err)
    require.Equal(t, expected, result)
}
```

### Setting Environment Variables

Use `t.SetEnv()` instead of `os.Setenv()`:
```go
func TestWithEnv(t *testing.T) {
    t.SetEnv("MY_VAR", "value")
    // Variable is set for duration of test, auto-restored
}
```

### Temporary Directories

Use `t.Tempdir()` for test files (auto-cleaned):
```go
func TestFileOp(t *testing.T) {
    tmpdir := t.Tempdir()
    testFile := filepath.Join(tmpdir, "test.txt")
    // File operations...
    // Automatically cleaned up after test
}
```

### Testing with Mock Providers

When tests need LLM provider interactions, use mock providers to avoid API calls:

```go
func TestAgentWithLLM(t *testing.T) {
    // Enable mock providers
    originalUseMock := config.UseMockProviders
    config.UseMockProviders = true
    defer func() {
        config.UseMockProviders = originalUseMock
        config.ResetProviders()  // Clear mock data
    }()
    
    // Reset to ensure fresh mock data
    config.ResetProviders()
    
    // Tests now use mock LLM responses
    providers := config.Providers()
    // ... test logic
}
```

### Golden Files (Snapshot Tests)

Nexora uses golden files for snapshot-based testing, especially for UI components:

```bash
# Update all golden files when expected output changes
go test ./... -update

# Update specific package
go test ./internal/tui/components/core -update

# Run normally (compares against golden files)
go test ./internal/tui/components/core
```

The test suite will fail if output differs from `.golden` files. Use `-update` flag to regenerate.

### Parallel Testing

Use `t.Parallel()` in all new tests to speed up test suite:
```go
func TestSomething(t *testing.T) {
    t.Parallel()
    // Test body
}
```

## Important Gotchas & Non-Obvious Patterns

### 1. Agent Loop Management

The agent loop in `internal/agent/agent.go` uses `fantasy.Agent.Stream()` with callbacks:
- `OnAssistantContent`: Receives text/reasoning from LLM
- `OnToolCall`: Called when LLM requests tool execution
- `OnToolResult`: Called after tool completes (must add result to message history)
- `OnStepFinish`: Called at end of each LLM step

**Critical**: Tool results MUST be added to message history via `OnToolResult`, or the agent won't see the results.

### 2. Tool Call Execution

Tool calls are managed in `internal/agent/tools/`:
- Each tool is a separate package implementing `ExecutableTool` interface
- Tools run in goroutines and communicate via channels
- Results are cached in memory but must be persisted via message service

Tools handle permissions via `internal/permission` service - they may prompt user for approval.

### 3. Message Conversion

Messages convert between:
- **Internal format**: `internal/message.Message` (stored in DB)
- **Fantasy format**: `fantasy.Message` (sent to LLM)
- **Provider-specific formats**: OpenAI, Anthropic, etc. (handled by fantasy library)

The conversion happens in `Message.ToAIMessage()` method in `internal/message/content.go`.

### 4. Session & Message Persistence

- Sessions stored in database via `internal/session` service
- Messages stored via `internal/message` service  
- Both use `internal/db` package for database operations
- Always use context-aware methods for cancellation support

### 5. TUI & Bubble Tea Integration

The TUI uses Bubble Tea v2 (from charm.land):
- Each page is a `Model` implementing `tea.Model` interface
- `Update` method handles messages/events
- `View` method renders to terminal
- Use `lipgloss.v2` for styling (not pure ANSI codes)

### 6. Devstral-2 & Mistral Tool Message Handling

**Known Issue Fixed**: Mistral API requires tool result messages (`role: "tool"`) in message history. Previous versions dropped them, causing infinite tool call loops.

The proxy at `/home/renter/devstral-proxy/devstral_proxy/utils.py` now:
- Preserves tool result messages from OpenAI format
- Converts them to Mistral format with `tool_call_id` matching
- Includes them in request body so LLM sees execution results

See `NEXORA.md` for detailed explanation.

### 7. Provider Configuration

Providers are initialized via `internal/config/provider.go`:
- Supports OpenAI, Anthropic, Google, Bedrock, OpenRouter, local compatible APIs
- Uses `charm.land/fantasy` library for unified interface
- Configuration read from environment or config file
- Mock providers available for testing

### 8. Loop Detection & Drift Detection

`internal/agent/agent.go` includes two detection mechanisms:
- **Loop detection** (every 5 tool calls): Detects when same tool is called repeatedly
- **Drift detection** (every 10 tool calls): Detects when agent is drifting from original task

Both run asynchronously and create system messages to alert the agent. They don't stop execution but inform the agent of potential issues.

### 9. LSP Integration

`internal/lsp/` provides Language Server Protocol client:
- Used to get additional context about code (definitions, references, hover info)
- Integrated with agent to enhance code understanding
- Non-critical: Agent works without LSP, but works better with it

### 10. Embedding Resources

The agent uses `//go:embed` for markdown templates:
```go
//go:embed templates/title.md
var titlePrompt []byte

//go:embed templates/summary.md
var summaryPrompt []byte
```

These are compiled into the binary and used for generating titles/summaries.

## Working with Dependencies

### Key Dependencies

- **charm.land/fantasy**: LLM abstraction layer (OpenAI, Anthropic, Google, etc.)
- **charm.land/bubbletea/v2**: Terminal UI framework
- **charm.land/lipgloss/v2**: Terminal styling
- **charmbracelet/glamour**: Markdown rendering
- **charmbracelet/catwalk**: Model metadata
- **modelcontextprotocol/go-sdk**: MCP support
- **lib/pq**: PostgreSQL driver

### Updating Dependencies

```bash
# Check for updates
go list -u -m all

# Update specific package
go get -u github.com/charm.land/fantasy

# Update all
go get -u ./...

# Verify dependencies
go mod verify

# Clean up unused
go mod tidy
```

## Debugging Tips

### Profiling

Enable profiling with environment variable:
```bash
NEXORA_PROFILE=true go run .
```

Then access profiles at `http://localhost:6060/debug/pprof`:
```bash
# CPU profile (10 seconds)
go tool pprof -http :6061 'http://localhost:6060/debug/pprof/profile?seconds=10'

# Heap profile
go tool pprof -http :6061 'http://localhost:6060/debug/pprof/heap'

# Allocations
go tool pprof -http :6061 'http://localhost:6060/debug/pprof/allocs'
```

### Logging

Uses `log/slog` package:
```go
import "log/slog"

slog.Info("Processing message", "id", id)
slog.Error("Failed to create", "error", err)
slog.Warn("Deprecated API", "version", "1.0")
slog.Debug("Internal state", "value", state)
```

### Debugging Tests

```bash
# Run with verbose output
go test -v ./internal/agent

# Run single test with output
go test -v -run TestCoderAgent ./internal/agent

# Show print statements
go test -v -run TestCoderAgent ./internal/agent 2>&1 | grep -A5 "TestCoderAgent"
```

## Version Management

- **Semantic versioning**: `MAJOR.MINOR.PATCH`
- **Version location**: `internal/version/version.go`
- **Update before release**: Edit `var Version = "x.x.x"`
- **Development version**: Append `+dev` to version for unreleased builds

## Committing Code

**Always use semantic commit messages**:
- `feat:` - New feature
- `fix:` - Bug fix
- `chore:` - Build, dependencies, non-code changes
- `refactor:` - Code reorganization without behavior change
- `docs:` - Documentation changes
- `sec:` - Security fixes
- `test:` - Test-related changes

Example: `feat: add support for streaming responses`

Keep commit messages to one line when possible. Use multi-line commits only when additional context is necessary.

## Code Review Considerations

Before submitting code:

1. Run `./scripts/lint.sh` - Complete quality check
2. Ensure all tests pass: `go test ./...`
3. Check formatting: `gofumpt -w .`
4. Verify no TODO/FIXME comments are leftover
5. Document public exports with comments
6. Use semantic commit message

The CI will run the same checks on pull requests, so local verification prevents failures.

## Documentation Structure

The project has comprehensive documentation:

- **NEXORA.md** (this directory): Architecture, code style, testing patterns
- **PROJECT_OPERATIONS.md**: Version management, build processes, release procedures
- **CODEDOCS.md**: Index of 40,000+ lines of code with symbol references
- **README.md**: Project overview, features, installation
- **todo.md**: Current issues and planned improvements

When investigating the codebase:
1. Check CODEDOCS.md for symbol locations
2. Read NEXORA.md for style and pattern examples
3. Review PROJECT_OPERATIONS.md for procedural guidance
4. Check todo.md for known issues affecting your work

## Performance & Best Practices

### Memory Management

- Use `t.Tempdir()` in tests (auto-cleanup)
- Limit goroutine creation in loops
- Be aware that messages are cached in memory

### Concurrency

- Agent loop is streaming (async callbacks)
- Tool execution happens in goroutines
- Use channels for tool result communication
- Always consider context cancellation

### API Integrations

- LLM calls via fantasy library (abstracts provider differences)
- LSP calls are synchronous, may timeout
- MCP connections are streaming (stdio, http, SSE)

## When in Doubt

1. **Check existing patterns**: Look at similar code in the codebase
2. **Read comments**: Exported items should have doc comments
3. **Check tests**: Test files show usage patterns
4. **Consult CODEDOCS.md**: Full reference of symbols and their locations
5. **Ask via git history**: `git log -p --all -S "symbol_name"` shows usage evolution

## Quick Reference Checklist

Before committing:

- [ ] Code formatted with `gofumpt -w .`
- [ ] All tests pass: `go test ./...`
- [ ] Lint checks pass: `./scripts/lint.sh`
- [ ] Exported items have doc comments
- [ ] No TODO/FIXME left behind
- [ ] Semantic commit message
- [ ] No API keys or secrets in code
- [ ] Interfaces are small and focused
- [ ] Error messages use `%w` for wrapping
- [ ] Context passed as first parameter where needed
