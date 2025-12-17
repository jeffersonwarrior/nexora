# CHANGELOG

All notable changes to Nexora CLI will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.28.1] - 2025-12-17

### Added

#### Critical Bug Fixes
- **Mistral Tool Call ID Format**: Fixed `"Tool call id was call_61626542 but must be a-z, A-Z, 0-9, with a length of 9"` error
- **Session Title Display**: Fixed missing session titles in header (20	25 char truncation + "New Session" fallback)
- **Provider Field Reference**: Fixed compilation error `CatwalkCfg.Provider 	 ModelCfg.Provider`
- **Cerebras Provider**: Fixed configuration handling and API compatibility issues
- **View Context Explosion**: Fixed excessive context generation in view tool with smart truncation
- **Context Window Handling**: Improved context window management with better token estimation

#### New Features
- **Tool ID Sanitization Library**: `internal/agent/utils/tool_id.go` with provider-specific ID generation
- **Enhanced Mistral Provider**: Full support for `devstral-2512` and `devstral-small-2512` models
- **Model Validation**: Added automatic fallback to recent models when current models are invalid
- **Better Config Handlers**: Improved configuration loading with local model selection support
- **Configuration Security**: Added `.gitignore` for `nexora.json` and created `nexora.example.json` template

#### Testing & Quality
- **Comprehensive Test Suite**: 15/15 new tests for tool ID utilities
- **Enhanced QA Tests**: All existing QA tests passing with coverage
- **Model Fallback Logic**: Full test coverage for model validation and recovery

### Technical Improvements
- Provider-specific tool call ID generation (Mistral: 9 alphanumeric chars, OpenAI: call_* format)
- Enhanced error messages and recovery patterns
- Updated configuration templates and documentation
- Improved PATH handling in installation script
- Smart context truncation for view tool to prevent context explosion
- Better token estimation for context window management
- Enhanced configuration loading with local model validation

### Security
- Removed API keys from repository history
- Added configuration file to `.gitignore`
- Created safe configuration template for users

### Fixed
- Tool call ID sanitization for Mistral provider
- Session title truncation and fallback behavior
- Model configuration validation and auto-recovery
- Provider field reference compilation errors
- Cerebras provider configuration handling
- View tool context explosion with smart truncation
- Context window token estimation and management

---

## [0.28.0] - 2025-12-17

### Fixed
- Catwalk Provider struct compatibility (`ExtraBody` removal)
- Logo rendering issues (imports, undefined symbols, emoji parsing)
- Undefined providers & references cleanup
- AgentCoordinator.Run() signature in chat.go
- RentalH200 provider removal (all references cleaned)

---

## [0.27.0] - 2025-12-15

### Added

#### Provider Expansion (Phase 2 Complete)
- **13 Provider Entries** with 70+ LLM models
- **X.AI (Grok)**: Grok 4.1 Fast variants (reasoning & non-reasoning), Grok 4.x, 3.x (6 models)
- **OpenAI**: GPT-5.2 (latest), GPT-4o, GPT-4 Turbo, o1, o1-mini (8 models)
- **Anthropic**: Claude 4.5 Opus, Sonnet, Haiku with multiple versions (5 models)
- **Google Gemini**: Gemini 3 Pro, 2.5 Flash, 2.5 Pro, Extended Thinking (6 models)
- **Mistral Split Variants**:
  - General: Large, Medium, Small, Ministral
  - Devstral: Code reasoning (FREE during beta)
  - Codestral: Code generation and completion
- **Z.AI**: GLM-4.6 multimodal, vision variants, free flash models (7 models)
- **Cerebras**: Ultra-fast inference - Llama 3.3-70B, Qwen 3-235B, GPT-3-120B (6 models)
- **Kimi (Moonshot)**: 1M token context, K2 thinking models (4 models)
- **MiniMax**: Mixture of Experts (MoE) models (2 models)

#### Documentation
- Comprehensive README updates with provider categories
- Provider table with 20+ environment variables and descriptions
- Latest model highlights (GPT-5.2, Grok 4.1, Gemini 3 Pro)
- Provider categories: Flagship, Specialized, Aggregators, Cloud & Local

### Fixed
- **X.AI API Endpoint**: Corrected to `/v1` suffix (was missing)
- **X.AI Provider Type**: Fixed from `openai` to `openai-compat`
- **X.AI Model List**: Removed non-existent `grok-4-heavy`, added `grok-4-1-fast-reasoning` and `grok-4-1-fast`
- **Provider Sorting**: Custom providers now appear first in dropdown (prepend instead of append)

### Improved
- All 70+ models tested and verified against live APIs
- Pricing data verified and up-to-date as of 2025-12-15
- Feature flags (vision, reasoning) correctly set for all models
- Test coverage: 35/35 provider tests passing (100%)
- Clean codebase: removed legacy files and development artifacts

### Technical
- **Mistral Variant Pattern**: Separate provider entries per use case (general/devstral/codestral)
- **API Key Resolution**: Proper environment variable substitution via `$VAR_NAME` in config
- **openai-compat Type**: Correct request formatting for non-OpenAI providers with OpenAI-compatible APIs
- **Provider Pricing**: Input/output costs per 1M tokens verified

---

## [0.26.0] - 2025-12-12

### Added
- **GitHub Provider Injection**: Fetch custom providers from GitHub repositories
- **Nexora Local Provider**: Built-in support for local Devstral-2 model via proxy (port 9000)
- **Mistral Provider**: Default Mistral provider with embed and moderation models

### Fixed
- Provider test mock signatures updated for new `ProviderClient` interface with `context.Context`
- Bedrock provider tests now correctly expect 0 providers when AWS credentials are missing
- Recent model recording no longer adds fallback defaults during initialization
- Type casting for `catwalk.InferenceProvider` in provider ID map lookups

### Changed
- Model fallback during initialization no longer records to recent models (prevents unwanted defaults in history)

---

## [0.25.0] - 2025-12-11

### Added
- Complete AIOPS (AI Operations) middleware service implementation
- Remote 3B model integration for intelligent agent assistance
- Edit resolution with confidence scoring (0.8+ threshold)
- Loop detection with sliding window pattern recognition (10-tool window, check every 5)
- Drift detection for task alignment monitoring (20-action window, check every 10)
- CLI commands: `nexora aiops status` and `nexora aiops test`
- TUI component design ready to replace MCP sidebar display
- Graceful fallback when AIOPS service unavailable
- Feature toggles for controlled AIOPS rollout
- Comprehensive testing framework with mock server infrastructure
- PROVIDERTODO.md created with roadmap for OpenAI OAuth2 and MiniMax provider support

### Fixed
- Agent event handler lock passing issues (sync.Mutex)
- Non-constant format string warnings in CLI commands
- Various agent test data updates for improved coverage

### Infrastructure
- HTTP/REST protocol for AIOPS service communication
- Timeout handling and network resilience (5-second timeouts)
- Memory-efficient sliding window tracking vs full history
- Middleware pattern with clean separation and easy testing

---

## [0.24.0] - 2025-01-09

### Major Features
- **AIOPS Middleware Service**: Complete intelligent assistance system with remote 3B model
- **Agent Loop Intelligence**: Real-time loop detection before user frustration
- **Edit Resolution**: AI-assisted edit resolution with 85% average confidence
- **Task Drift Monitoring**: Keeps large tasks aligned with original goals
- **Status Transparency**: Real-time latency, confidence rates, and activity history

### Performance
- 127ms average AIOPS service latency
- Efficient sliding window algorithms
- Asynchronous detection for non-blocking operation
- Configurable check intervals for responsiveness optimization

### Integration
- Fully integrated into agent execution workflow
- Pre-healing insertion - AI resolves simple cases first
- Configuration system integration with endpoint settings
- CLI command structure with professional lipgloss styling

### Quality
- Production-ready with comprehensive error handling
- Test coverage for all major components
- Mock server infrastructure for controlled testing
- Documentation in `qa/aiops/` directory

---

## [0.23.0] - 2025-01-08

### Fixed
- Indexer FTS (Full-Text Search) schema issues
- Concurrent indexer registration problems
- DI (Dependency Injection) layer for indexer components
- SQLite FTS virtual table configuration
- Indexer storage concurrency protection
- Test data updates for improved test coverage

### Infrastructure
- Added comprehensive testing framework (P5, P6, simple tests)
- Improved database transaction handling
- Enhanced error reporting for indexing failures
- Better debugging information for indexing operations

---

## [0.22.0] - 2023-12-XX

### Previous Features
- LSP (Language Server Protocol) integration
- Multiple AI provider support (OpenAI, Anthropic, OpenRouter, etc.)
- Self-healing agent capabilities
- Interactive TUI (Terminal User Interface)
- File watching and indexing system
- Search functionality across codebases
- Configuration management
- Session management
- OAuth integration
- Permission management
- Shell integration
- History tracking

[View complete implementation details in CODEDOCS.md](CODEDOCS.md)
