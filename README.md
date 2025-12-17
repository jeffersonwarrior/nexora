# Nexora - 🚀 AI-Powered CLI Agent

> Nexora is a powerful command-line AI assistant that combines modern language models with intelligent tool execution for seamless development workflows.

## ✨ Quick Start

```bash
# Install Nexora
curl -fsSL https://nexora.land/install.sh | sh

# Or clone and build
git clone https://github.com/nexora/cli.git
cd cli
./install.sh

# Start chatting
nexora chat
```

## 🎯 What's New

### v0.28.1 (2025-12-17) - Critical Bug Fixes

**🔥 Critical Fixes:**
- **🛠️ Mistral Tool Call ID Format**: Fixed `"Tool call id was call_61626542 but must be a-z, A-Z, 0-9, with a length of 9"` error
- **🏷️ Session Title Display**: Fixed missing session titles in header (20→25 char truncation + "New Session" fallback)
- **🔧 Provider Field Reference**: Fixed compilation error `CatwalkCfg.Provider → ModelCfg.Provider`
- **🔀 Model Validation**: Added automatic fallback to recent models when current models are invalid
- **🛡️ Configuration Security**: Removed API keys from repository, added `.gitignore` and template

**🆕 New Features:**
- **📦 Tool ID Sanitization Library**: Provider-specific ID generation with comprehensive testing
- **🔄 Enhanced Mistral Provider**: Full support for `devstral-2512` and `devstral-small-2512` models
- **📋 Auto Model Fallback**: Automatic recovery from invalid model configurations
- **🧪 Comprehensive Test Suite**: 15/15 new tests + existing QA tests all passing

**⚙️ Technical Improvements:**
- Provider-specific tool call ID generation (Mistral: 9 alphanumeric chars, OpenAI: call_* format)
- Enhanced error messages and recovery patterns
- Updated `nexora.example.json` template for user configuration

**📦 Installation:**
```bash
# Clean install with all fixes
./install.sh 0.28.1
nexora --version  # → "nexora version 0.28.1"
```

**⚙️ Configuration:**
```bash
# Copy template and configure
cp nexora.example.json nexora.json
# Edit nexora.json with your API keys and preferred models
```

**🚀 Status**: Production ready ✓ All tests passing ✓ Clean build ✓ Deployed

---

## 🧠 AI Provider Support

Nexora supports **70+ models** from **13 providers**:

### 🏆 Premium Providers
| Provider | Models | Best For |
|----------|--------|----------|
| **OpenAI** | GPT-5.2, GPT-4o, o1, o1-mini | General purpose, reasoning |
| **Anthropic** | Claude 4.5 Opus, Sonnet, Haiku | Coding, analysis, reasoning |
| **Google** | Gemini 3 Pro, 2.5 Flash | Multimodal, reasoning |
| **Mistral** | Devstral 2, Codestral | Code reasoning, generation |

### 🚀 Fast & Specialized
| Provider | Models | Best For |
|----------|--------|----------|
| **X.AI (Grok)** | Grok 4.1, Grok 4.x | Real-time analysis |
| **Cerebras** | Llama 3.3-70B, Qwen 3-235B | Ultra-fast inference |
| **Z.AI** | GLM-4.6, vision models | Multimodal, free tier |

### 📚 Free & Open Source
| Provider | Models | Features |
|----------|--------|----------|
| **Kimi** | K2 thinking models | 1M token context |
| **MiniMax** | MoE models | Mixture of Experts |

## 🛠️ Core Features

### 🎯 Smart Tool Execution
- **File Operations**: `view`, `edit`, `multiedit`, `write`, `ls`, `find`
- **Development**: `bash`, `git`, build tools, testing frameworks
- **Web Integration**: `fetch`, `web-search`, `sourcegraph` code search
- **MCP Tools**: Web reader, enhanced search capabilities

### 🧠 AI-Powered Assistance
- **Context Awareness**: Reads your codebase automatically
- **Error Recovery**: Intelligent retry and fallback mechanisms
- **Session Management**: Persistent conversations with summaries
- **Multi-Model Support**: Switch between providers seamlessly

### 🎨 User Experience
- **Beautiful TUI**: Modern terminal interface with chat
- **Syntax Highlighting**: Code-aware editing with validation
- **Auto-completion**: Smart command and file suggestions
- **Cross-platform**: Linux, macOS, Windows support

## ⚙️ Configuration

### Quick Setup
```bash
# Create configuration
cp nexora.example.json nexora.json

# Add your API keys
{
  "models": {
    "large": {
      "provider": "openai",
      "model": "gpt-4o"
    },
    "small": {
      "provider": "openai",
      "model": "gpt-4o-mini"
    }
  },
  "providers": {
    "openai": {
      "api_key": "$OPENAI_API_KEY"
    }
  }
}
```

### Environment Variables
```bash
# OpenAI
export OPENAI_API_KEY="sk-..."

# Anthropic  
export ANTHROPIC_API_KEY="sk-ant-..."

# Google Gemini
export GOOGLE_API_KEY="AI..."

# Mistral
export MISTRAL_API_KEY="..."

# ... and 9 more providers
```

## 🚀 Advanced Usage

### Session Management
```bash
# Start new session
nexora chat "Build a REST API in Go"

# Continue session  
nexora chat --continue

# List sessions
nexora sessions list

# Delete session
nexora sessions delete <session-id>
```

### Tool Examples
```bash
# Chat with automatic tool execution
nexora chat "Fix the build errors in main.go"

# Search codebase
nexora search "TODO comments" --type go

# Edit files safely
nexora edit main.go --find "func main" --replace "// Updated main function"
```

### Provider Switching
```bash
# Use specific model
nexora chat --provider mistral --model devstral-2512 "Review this code"

# List available models
nexora models list

# Set default model
nexora config set models.large.provider mistral
nexora config set models.large.model devstral-2512
```

## 🧪 Development

### Build from Source
```bash
# Clone repository
git clone https://github.com/nexora/cli.git
cd cli

# Install dependencies
go mod download

# Build binary
make build

# Run tests
make test

# Install locally
make install-user
```

### Development Tools
```bash
# Install all dev tools
make install-tools

# Run specific tests
go test ./qa/ -v

# Run integration tests
go test ./internal/agent/... -v
```

### Project Structure
```
├── internal/
│   ├── agent/          # Core AI agent logic
│   ├── config/         # Provider and model configuration
│   ├── tui/            # Terminal user interface
│   ├── message/        # Message handling and persistence
│   └── tools/          # Tool implementations
├── qa/                 # Quality assurance tests
└── providers/          # AI provider integrations
```

## 🐛 Troubleshooting

### Common Issues

**🔑 API Key Problems**
```bash
# Verify environment variables
env | grep API_KEY

# Test connection
nexora chat --provider openai --model gpt-4o "test"
```

**🛠️ Build Issues**
```bash
# Clean and rebuild
make clean
make build

# Update dependencies
go mod tidy
```

**🏷️ Session Title Not Showing**
```bash
# This was fixed in v0.28.1
# Update to latest version
./install.sh
```

**🔧 Mistral Tool Call ID Error**
```bash
# Fixed in v0.28.1
# Tool call IDs are automatically sanitized to 9 alphanumeric chars
nexora chat --provider mistral "test tool execution"
```

### Get Help
- **Documentation**: [docs.nexora.land](https://docs.nexora.land)
- **Issues**: [GitHub Issues](https://github.com/nexora/cli/issues)
- **Discussions**: [GitHub Discussions](https://github.com/nexora/cli/discussions)

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

## 🤝 Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

---

## 📈 Changelog

### v0.28.1 (2025-12-17) - Critical Bug Fixes
- Fixed Mistral tool call ID format error
- Enhanced session title display
- Added model validation and fallback
- Improved configuration security
- Comprehensive test coverage

### v0.28.0 (2025-12-17) - Maintenance Release
- Build and stability fixes
- Provider compatibility updates
- Logo rendering improvements

### v0.27.0 (2025-12-15) - Major Provider Expansion
- Added 13 providers with 70+ models
- Comprehensive model catalog
- Enhanced documentation

---

**Made with ❤️ by the Nexora team**