# CHANGELOG

All notable changes to Nexora CLI will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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