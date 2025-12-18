# Advanced Tools Installer

A comprehensive installer script for AI development tools on Linux and macOS systems.

## Quick Start

```bash
# Install all tools
./scripts/install-tools.sh

# Install specific category
./scripts/install-tools.sh git          # Git enhancement tools
./scripts/install-tools.sh search       # Search and navigation tools
./scripts/install-tools.sh text         # Text processing and editors
./scripts/install-tools.sh network      # Network and monitoring tools
./scripts/install-tools.sh sys          # System utilities
./scripts/install-tools.sh dev          # Development tools
```

## Options

- `--dry-run` - Preview what would be installed without installing
- `--skip-update` - Skip package manager updates
- `--verbose` - Show detailed output
- `--help` - Show help message

## Tool Categories

### Git Tools
- **Git** - Version control
- **GitHub CLI** (`gh`) - GitHub command-line interface
- **Git-delta** - Better git diffs
- **Lazygit** - Terminal Git UI
- **Tig** - Text interface for Git

### Search & Navigation Tools
- **ripgrep** (`rg`) - Ultra-fast text search
- **fd** - Modern find replacement
- **fzf** - Command-line fuzzy finder
- **bat** - Better cat with syntax highlighting
- **exa/eza** - Modern ls replacement
- **tree** - Directory tree viewer
- **the silver searcher** (`ag`) - Fast code search

### Text Processing Tools
- **Neovim** - Modern Vim fork
- **jq** - JSON processor
- **yq** - YAML processor
- **AWK/GAWK** - Text processing
- **Markdownlint** - Markdown linter

### Network Tools
- **htop** - Process viewer
- **iftop** - Network bandwidth monitor
- **nmap** - Network scanner
- **curl/wget** - HTTP clients
- **xh** - Modern HTTP client
- **bandwhich** - Network bandwidth utilization

### System Utilities
- **tmux** - Terminal multiplexer
- **ncdu** - NCurses disk usage
- **dust** - Disk usage analyzer
- **p7zip/zip/unzip** - Archive utilities
- **lsof** - List open files
- **strace** - System call tracer

### Development Tools
- **Python 3** + **pip** - Python development
- **Node.js** + **npm** - JavaScript development
- **Build tools** - gcc, g++, make, cmake
- **Docker** - Container platform
- **AWS CLI** & **Google Cloud CLI** - Cloud tools
- **Watchexec** - Run commands on file changes
- **Serve** - Static file server
- **gum** - Interactive CLI tool

## Examples

```bash
# Preview installation without installing
./scripts/install-tools.sh --dry-run

# Install just search tools with detailed output
./scripts/install-tools.sh search --verbose

# Install dev tools without updating package manager
./scripts/install-tools.sh dev --skip-update

# Install git tools and show help
./scripts/install-tools.sh git --help
```

## Platform Support

- **Linux**: Debian/Ubuntu, RedHat/CentOS, Arch Linux, Alpine
- **macOS**: via Homebrew

## Multi-Package Manager Support

The installer automatically detects and uses:
- **apt-get** (Debian/Ubuntu)
- **yum** (RedHat/CentOS)
- **pacman** (Arch Linux)
- **apk** (Alpine)
- **brew** (macOS)
- **cargo** (Rust packages)
- **npm** (Node.js packages)
- **go install** (Go packages)

## Post-Installation

After installation, you may need to:
1. Restart your terminal or run `source ~/.bashrc`
2. Verify tool installations: `which rg`, `which fd`, `which fzf`
3. Check tool versions: `rg --version`, `fd --version`, `fzf --version`