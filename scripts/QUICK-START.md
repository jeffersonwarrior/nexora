# Quick Usage Guide

## Created Files

✅ **`/home/nexora/scripts/install-tools.sh`** - Main installer script
✅ **`/home/nexora/scripts/README-tools.md`** - Complete documentation

## Available Tools

Your system already has these essential AI development tools installed:

- **`rg` (ripgrep)** - Ultra-fast text search
- **`fd`** - Modern find replacement  
- **`fzf`** - Command-line fuzzy finder
- **`bat`** - Better cat with syntax highlighting
- **`exa`** - Modern ls replacement
- **`jq`** - JSON processor
- **`gh`** - GitHub CLI
- **`htop`** - Process viewer

## Quick Examples

### Search
```bash
# Find files containing "TODO" 
rg "TODO" .

# Find files by name
fd "*.go"

# Interactive search with fzf
fzf
```

### File Management
```bash
# Better cat with syntax highlighting
bat main.go

# Better directory listing
exa --tree --long

# Interactive directory navigation
fzf --preview='bat --color=always {}'
```

### Git Integration
```bash
# Enhanced git experience
git log --oneline | fzf

# GitHub operations
gh issue list
gh pr list
gh repo view
```

## Install More Tools

```bash
# Install all tools (comprehensive)
./scripts/install-tools.sh

# Install specific categories
./scripts/install-tools.sh git          # Git tools
./scripts/install-tools.sh search       # Search tools
./scripts/install-tools.sh text         # Text editing tools
./scripts/install-tools.sh network      # Network tools
./scripts/install-tools.sh sys          # System utilities
./scripts/install-tools.sh dev          # Development tools

# Preview without installing
./scripts/install-tools.sh all --dry-run
```

## AI Development Benefits

These tools improve AI-assisted development by:

1. **Better Search**: `rg` finds code faster than grep
2. **Efficient Navigation**: `fd` + `fzf` = instant file discovery
3. **Better Code Reading**: `bat` highlights syntax for easier analysis
4. **Clean Output**: `exa` provides clearer directory views
5. **GitHub Integration**: `gh` for API interactions
6. **Performance**: Faster commands reduce context switching

The installer script is ready to enhance any Linux/macOS system for optimal AI development work!