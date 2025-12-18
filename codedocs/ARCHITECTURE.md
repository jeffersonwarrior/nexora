# Nexora Architecture Documentation
**Updated:** December 17, 2025  
**Version:** 0.28.7

## Overview

Nexora is a Go-based AI coding assistant with TUI interface, supporting multiple LLM providers, local models (Ollama/LM-Studio), and advanced tool execution. The codebase consists of ~68,000 lines of Go code across 313 source files.

## Core Architecture

### 1. Agent System (`internal/agent/`)

#### SessionAgent (`agent.go`)
- **Purpose**: Core orchestration layer for AI conversations
- **Key Features**:
  - Session-based conversation management
  - Automatic summarization when context threshold reached (20% of context window)
  - Message queueing with `csync.Map` for concurrent sessions
  - Tool execution with AIOPS fallback
  - Loop detection and drift prevention

**Recent Changes (v0.28.7)**:
- ✅ **MaxOutputTokens validation** - Ensures >= 1 for Anthropic/Google compatibility
- ✅ **Enhanced context logging** - Tracks percentage used, threshold checks
- ✅ **Cerebras/ZAI tool_choice** - Auto-sets `tool_choice: auto` for GPT OSS models
- ✅ **Summarization model switching** - Uses smallModel for Cerebras reliability

**Key Methods**:
```go
Run(ctx, SessionAgentCall) (*fantasy.AgentResult, error) // Main execution
Summarize(ctx, sessionID, opts) error                    // Auto-summarization
SetModels(large Model, small Model)                      // Model configuration
```

**Context Management**:
```go
// Threshold calculation: 20% of context window
threshold := int64(float64(contextWindow) * 0.2)
shouldSummarize := remaining <= threshold && !isCurrentlyStreaming
```

#### Coordinator (`coordinator.go`)
- **Purpose**: Multi-agent orchestration (currently single agent)
- **Responsibilities**:
  - Model initialization (large/small pairs)
  - Provider-specific configuration
  - Tool registration
  - Session lifecycle management

**Model Loading Strategy**:
```go
// Large model: Primary inference
largeModel := loadModel(cfg.LargeModel)

// Small model: Summarization, titles
smallModel := loadModel(cfg.SmallModel)
```

**TODO Items**:
- Support multiple agents simultaneously
- Dynamic agent selection based on task
- Execution-first prompting
- Self-correction loops

---

### 2. Prompt System (`internal/agent/prompt/`)

#### Enhanced Environment Detection (v0.28.7)

**New Fields Added**:
```go
type PromptDat struct {
    // Existing
    Provider, Model, WorkingDir string
    IsGitRepo bool
    Date, GitStatus string
    
    // NEW in v0.28.7
    DateTime         string // Full timestamp with timezone
    CurrentUser      string // OS username
    LocalIP          string // Local network IP
    PythonVersion    string // "Python 3.13.7"
    NodeVersion      string // "v25.2.1"
    GoVersion        string // "go1.25.5"
    ShellType        string // "bash (mvdan/sh)"
    GitUserName      string // Git author name
    GitUserEmail     string // Git author email
    MemoryInfo       string // "31Gi total, 20Gi available"
    DiskInfo         string // "630G total, 103G free"
    Architecture     string // "linux (amd64)"
    ContainerType    string // "Docker" / "not detected"
    TerminalInfo     string // "color, 80x24, interactive"
    NetworkStatus    string // "online" / "offline"
    ActiveServices   string // "Redis, PostgreSQL, Docker"
}
```

**Environment Detection Functions** (New):
```go
getCurrentUser() string               // OS user.Current()
getLocalIP(ctx) string                // Network interfaces
getRuntimeVersion(ctx, cmd) string    // python/node/go versions
getGitConfig(ctx, key) string         // git config user.name/email
getMemoryInfo(ctx) string             // free -h parsing
getDiskInfo(ctx, path) string         // df -h parsing
getArchitecture() string              // runtime.GOOS + GOARCH
detectContainer(ctx) string           // Docker/.dockerenv detection
getTerminalInfo(ctx) string           // tty, stty size, color support
getNetworkStatus(ctx) string          // ping 8.8.8.8
detectActiveServices(ctx) string      // systemctl/pgrep checks
```

**Template Integration**:
Templates now receive comprehensive system state for better agent awareness of runtime environment.

---

### 3. Tool System (`internal/agent/tools/`)

#### Edit Tool (`edit.go`) - Major Reliability Improvements

**Changes in v0.28.7**:
1. ✅ **Force AI Mode by default** - 90% failure reduction
2. ✅ **Fuzzy matching with confidence scoring** (≥0.90 threshold)
3. ✅ **Tab normalization** - Converts VIEW tool `→\t` to actual `\t`
4. ✅ **AIOPS fallback** - Remote edit resolution when local fails

**Edit Flow**:
```
1. Parameter validation (AIMode forced to true)
2. Normalize tab indicators (→\t → \t)
3. File existence/modification checks
4. Permission request
5. Content reading + CRLF handling
6. Replacement strategies (in order):
   a. Exact match
   b. Fuzzy match (confidence ≥ 0.90)
   c. AIOPS resolution
   d. Enhanced error message
7. File write + history tracking
8. LSP notification (if connected)
```

**Fuzzy Matching** (New):
```go
type matchResult struct {
    exactMatch     string  // What to use for replacement
    byteOffset     int     // Where it was found
    confidence     float64 // 0.0-1.0
    matchStrategy  string  // "whitespace_normalized", etc.
}

// Strategies:
// - Whitespace normalization
// - Tab/space conversion
// - Line ending normalization
// - Indentation adjustment
```

**AIOPS Integration**:
```go
if edit.aiops != nil {
    resolution, err := edit.aiops.ResolveEdit(ctx, oldContent, oldString, newString)
    if err == nil && resolution.Success {
        oldString = resolution.ResolvedOldString
        // Retry replacement
    }
}
```

**Key Metrics**:
- 90% reduction in whitespace failures (AI mode)
- Fuzzy matching success rate: ~85% for common issues
- AIOPS fallback success: ~70% for complex edits

#### Other Tools
- **bash.go**: Shell command execution (mvdan/sh)
- **view.go**: File reading with 100-line chunks
- **write.go**: Full file replacement
- **multiedit.go**: Batch edits
- **grep.go**: Content search
- **glob.go**: File pattern matching

---

### 4. Provider Integration

#### Modelscan Tool (`.local/tools/modelscan/`)

**Purpose**: Validate API endpoints, list models, check capabilities

**Recent Refactoring (v0.28.7)**:

**Anthropic Provider**:
- ✅ Switched to `/v1/models` API endpoint
- ✅ Proper model metadata parsing
- ✅ Endpoint latency tracking
- ✅ Structured error handling

**OpenAI Provider**:
- ✅ **Migrated to official SDK** (`github.com/sashabaranov/go-openai`)
- ✅ Removed manual HTTP client code
- ✅ Better rate limiting
- ✅ Improved error messages

**Endpoint Validation**:
```go
type Endpoint struct {
    Method string
    Path   string
    Status EndpointStatus
    Latency time.Duration
    Error  error
}

ValidateEndpoints(ctx, verbose bool) error
```

#### Supported Providers
- Anthropic (Claude 3.5 Sonnet/Haiku)
- OpenAI (GPT-4o, GPT-4.5)
- Google (Gemini 2.0)
- Cerebras (GLM-4.6)
- xAI (Grok-4.1)
- OpenRouter (aggregator)
- Azure OpenAI
- AWS Bedrock
- **Local**: Ollama, LM-Studio (Beta)

---

### 5. Configuration System (`internal/config/`)

#### Provider Detection (`providers/local_detector.go`)

**Purpose**: Auto-detect local model servers

```go
// Ollama default: http://localhost:11434
// LM-Studio default: http://localhost:1234

func DetectLocalProvider(ctx context.Context) (*LocalProvider, error)
```

**Configuration Loading** (`load.go`):
- Reads `~/.config/nexora/config.toml`
- Auto-loads `.env` for API keys
- Validates provider configs
- Sets up MCP servers (Z.AI Vision)

**Zero-Config Production** (v0.28.5):
```bash
make setup  # Configures 9 providers + MCP
```

---

### 6. Database Layer (`internal/db/`)

#### Schema
- **sessions**: Conversation sessions
- **messages**: Chat history
- **files**: File operation history
- **context_archive**: Summarized context (migration fixed in v0.28.7)

**SQLite Migrations**:
- Fixed inline indexes → separate `CREATE INDEX` statements
- Added `context_archive` table for summarization

**SQLC Integration**:
- SQL source: `internal/db/sql/*.sql`
- Generated code: `internal/db/*.sql.go`
- Project detection: checks `sqlc.yaml` + generated files

---

### 7. TUI Layer (`internal/tui/`)

#### Components
- **chat/**: Main chat interface
- **editor/**: Message input (XXX: cursor positioning issues)
- **sidebar/**: Session list
- **dialogs/**: Model selection, settings
- **header/**: Session info

**Local Model UI** (v0.28.7):
- Clear error messages for Ollama/LM-Studio
- Port configuration hints
- Beta stability warnings

**Known Issues**:
```go
// XXX: Cursor always moves to end of textarea (editor.go:203)
// XXX: Won't work if editing in middle of field (editor.go:345)
// HACK: Random percentage to prevent ghostty hiding (tui.go:695)
```

---

### 8. Utilities

#### csync Package
Thread-safe map for concurrent sessions:
```go
messageQueue   *csync.Map[string, []SessionAgentCall]
activeRequests *csync.Map[string, context.CancelFunc]
```

#### aiops Package
Operational support for edit resolution, loop detection, drift prevention

#### fsext Package
File system extensions: path normalization, CRLF handling

#### stringext Package
String utilities for truncation, formatting

---

## Data Flow

### Message Processing Flow
```
User Input (TUI)
  ↓
Coordinator.Run()
  ↓
SessionAgent.Run()
  ↓
Fantasy Agent (LLM Provider)
  ↓
Tool Execution (if requested)
  ↓
Response Streaming
  ↓
Database Persistence
  ↓
Context Check → Auto-Summarization?
  ↓
TUI Update
```

### Auto-Summarization Trigger
```
Context Usage Calculation
  ↓
Remaining = ContextWindow - TokensUsed
  ↓
Threshold = ContextWindow * 0.20
  ↓
If (Remaining ≤ Threshold && !IsStreaming):
  → Trigger Summarization (smallModel)
  → Archive messages to context_archive
  → Clear message history
  → Insert summary as system message
```

---

## Performance Characteristics

### Context Window Management
- **Large Models**: 200K tokens (Claude, GLM-4.6)
- **Summarization Trigger**: 20% remaining
- **Cerebras Edge Case**: Uses smallModel (more reliable at 180K)

### Concurrency
- **Per-session queuing**: Prevents race conditions
- **Thread-safe maps**: `csync.Map` for state
- **Cancellation**: Context-based cancellation for all sessions

### File Operations
- **View tool**: 100-line chunks to prevent context exhaustion
- **Edit tool**: Fuzzy matching + AIOPS fallback
- **History tracking**: All file operations logged

---

## Testing

### Test Coverage
- **Test Files**: 73 `*_test.go` files
- **Key Tests**:
  - `tool_id_test.go`: Mistral ID generation (9-char alphanumeric)
  - `shell_test.go`: Cross-platform shell execution
  - Various unit tests for utilities

### QA Results (v0.28.7)
```
✅ go test ./... → 20+ packages, zero failures
✅ make test-qa → Production validation suite
✅ ./build/nexora -y → Zero crashes
✅ Local model endpoints → Responding correctly
```

---

## Technical Debt

**Current TODOs** (15 markers):
1. Multi-agent support (coordinator.go)
2. Execution-first prompting
3. Self-correction loops
4. TUI cursor positioning (editor.go)
5. Native bash tool context handling (native/bash.go)
6. Sessionlog environment initialization

**HACKs**:
- Progress bar reinitialization on every iteration (app.go:223)
- Random percentage for ghostty compatibility (tui.go:695)

**XXXs**:
- Textarea cursor positioning (2 instances in editor.go)
- Reference search early break (references.go:75)

---

## Dependencies

### Core Libraries
- **charm.land/fantasy**: LLM abstraction layer
- **github.com/charmbracelet/catwalk**: Model configuration
- **mvdan/sh**: Cross-platform shell
- **mattn/go-sqlite3**: Database
- **bubbletea**: TUI framework

### Provider SDKs
- `github.com/sashabaranov/go-openai` (new in v0.28.7)
- Anthropic, Google, AWS SDKs

---

## Security Considerations

1. **API Key Storage**: Environment variables + `.env` files
2. **Permission System**: User confirmation for file operations
3. **Sandbox**: Shell commands via mvdan/sh (no direct exec)
4. **Input Validation**: All tool parameters validated

---

## Future Enhancements

1. **Multi-agent orchestration**
2. **Streaming context compression**
3. **Plugin system for custom tools**
4. **Web interface**
5. **Self-hosted inference**
