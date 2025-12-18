#!/usr/bin/env bash

# Advanced Tools Installer for Linux/macOS
# Installs essential development tools for AI-assisted development
# Usage: ./scripts/install-tools.sh [category]
# Categories: all, git, search, text, network, sys, dev (default: all)
# Note: Automatically elevates to sudo for system-wide installation

set -e

# Colors for output - force color output
if [ -t 1 ]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    BLUE='\033[0;34m'
    PURPLE='\033[0;35m'
    CYAN='\033[0;36m'
    NC='\033[0m' # No Color
else
    RED='' GREEN='' YELLOW='' BLUE='' PURPLE='' CYAN='' NC=''
fi

# Configuration
VERBOSE=0
DRY_RUN=0
SKIP_UPDATE=0
TO_INSTALL_CATEGORY="${1:-all}"

# Help message
show_help() {
    cat << EOF
Advanced Tools Installer - AI Development Tools

USAGE:
    $0 [CATEGORY] [OPTIONS]

CATEGORIES:
    all         Install all tools (default)
    git         Git-enhanced tools
    search      Search and navigation tools  
    text        Text processing and editors
    network     Network and monitoring tools
    sys         System utilities
    dev         Development tools
    
OPTIONS:
    --dry-run   Show what would be installed without installing
    --skip-update Skip package manager updates
    --verbose   Show detailed output
    --help      Show this help message

EXAMPLES:
    $0 all                    # Install all tools
    $0 git --verbose          # Install git tools with verbose output
    $0 search --dry-run       # Preview search tools installation
    $0 sys --skip-update      # Install system tools without updating

NOTE:
    This installer automatically elevates to sudo for system-wide installation.

EOF
}

# Parse arguments
for arg in "$@"; do
    case $arg in
        --dry-run)
            DRY_RUN=1
            ;;
        --verbose)
            VERBOSE=1
            ;;
        --skip-update)
            SKIP_UPDATE=1
            ;;
        --help)
            show_help
            exit 0
            ;;
    esac
done

# Functions
log() { printf "${BLUE}%s${NC}\n" "$*" >&2; }
log_success() { printf "${GREEN}✓ %s${NC}\n" "$*" >&2; }
log_error() { printf "${RED}✗ %s${NC}\n" "$*" >&2; }
log_warning() { printf "${YELLOW}⚠ %s${NC}\n" "$*" >&2; }
log_info() { printf "${CYAN}ℹ %s${NC}\n" "$*" >&2; }

verb() { [ "$VERBOSE" -eq 1 ] && printf "${PURPLE}[DEBUG] %s${NC}\n" "$*" >&2; }

# System detection
detect_system() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        if [ -f /etc/debian_version ]; then
            echo "debian"
        elif [ -f /etc/redhat-release ]; then
            echo "redhat"
        elif [ -f /etc/arch-release ]; then
            echo "arch"
        elif command -v apk >/dev/null 2>&1; then
            echo "alpine"
        else
            echo "linux-unknown"
        fi
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        echo "macos"
    else
        echo "unknown"
    fi
}

# Package manager detection
get_package_manager() {
    local system=$(detect_system)
    case $system in
        debian)
            echo "apt-get"
            ;;
        redhat)
            echo "yum"
            ;;
        arch)
            echo "pacman"
            ;;
        alpine)
            echo "apk"
            ;;
        macos)
            if command -v brew >/dev/null 2>&1; then
                echo "brew"
            else
                echo "brew-missing"
            fi
            ;;
        *)
            echo "unknown"
            ;;
    esac
}

# Update package manager
update_package_manager() {
    if [ "$SKIP_UPDATE" -eq 1 ]; then
        log_warning "Skipping package manager update"
        return 0
    fi

    local pm=$(get_package_manager)
    log "Updating package manager ($pm)..."
    
    case $pm in
        apt-get)
            if [ "$DRY_RUN" -eq 1 ]; then
                log_info "[DRY RUN] Would run: sudo apt-get update"
            else
                sudo apt-get update -qq
            fi
            ;;
        yum)
            if [ "$DRY_RUN" -eq 1 ]; then
                log_info "[DRY RUN] Would run: sudo yum update -y"
            else
                sudo yum update -y -q
            fi
            ;;
        pacman)
            if [ "$DRY_RUN" -eq 1 ]; then
                log_info "[DRY RUN] Would run: sudo pacman -Sy"
            else
                sudo pacman -Sy --noconfirm
            fi
            ;;
        apk)
            if [ "$DRY_RUN" -eq 1 ]; then
                log_info "[DRY RUN] Would run: sudo apk update"
            else
                sudo apk update -q
            fi
            ;;
        brew)
            if [ "$DRY_RUN" -eq 1 ]; then
                log_info "[DRY RUN] Would run: brew update"
            else
                brew update -q
            fi
            ;;
        brew-missing)
            log_error "Homebrew is not installed on macOS. Please install it first:"
            log_info '/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"'
            return 1
            ;;
        *)
            log_error "Unknown package manager"
            return 1
            ;;
    esac
    
    log_success "Package manager updated"
}

# Install package
install_pkg() {
    local cmd="$1"
    local pkg="$2"
    local description="$3"
    
    if command -v "$cmd" >/dev/null 2>&1; then
        log_success "$description is already installed ($(command -v "$cmd"))"
        return 0
    fi
    
    log "Installing $description..."
    
    local pm=$(get_package_manager)
    case $pm in
        apt-get)
            if [ "$DRY_RUN" -eq 1 ]; then
                log_info "[DRY RUN] Would run: sudo apt-get install -y $pkg"
                log_success "Would install $description"
            else
                verb "Running: sudo apt-get install -y $pkg"
                if sudo apt-get install -y "$pkg"; then
                    log_success "$description installed"
                else
                    log_error "Failed to install $description"
                    return 1
                fi
            fi
            ;;
        yum)
            if [ "$DRY_RUN" -eq 1 ]; then
                log_info "[DRY RUN] Would run: sudo yum install -y $pkg"
                log_success "Would install $description"
            else
                verb "Running: sudo yum install -y $pkg"
                if sudo yum install -y "$pkg"; then
                    log_success "$description installed"
                else
                    log_error "Failed to install $description"
                    return 1
                fi
            fi
            ;;
        pacman)
            if [ "$DRY_RUN" -eq 1 ]; then
                log_info "[DRY RUN] Would run: sudo pacman -S --noconfirm $pkg"
                log_success "Would install $description"
            else
                verb "Running: sudo pacman -S --noconfirm $pkg"
                if sudo pacman -S --noconfirm "$pkg"; then
                    log_success "$description installed"
                else
                    log_error "Failed to install $description"
                    return 1
                fi
            fi
            ;;
        apk)
            if [ "$DRY_RUN" -eq 1 ]; then
                log_info "[DRY RUN] Would run: sudo apk add $pkg"
                log_success "Would install $description"
            else
                verb "Running: sudo apk add $pkg"
                if sudo apk add "$pkg"; then
                    log_success "$description installed"
                else
                    log_error "Failed to install $description"
                    return 1
                fi
            fi
            ;;
        brew)
            if [ "$DRY_RUN" -eq 1 ]; then
                log_info "[DRY RUN] Would run: brew install $pkg"
                log_success "Would install $description"
            else
                verb "Running: brew install $pkg"
                if brew install "$pkg"; then
                    log_success "$description installed"
                else
                    log_error "Failed to install $description"
                    return 1
                fi
            fi
            ;;
        *)
            log_error "Cannot install $description: Unknown package manager"
            return 1
            ;;
    esac
}

# Install via cargo (rust)
install_cargo_pkg() {
    local cmd="$1"
    local pkg="$2"
    local description="$3"
    
    if command -v "$cmd" >/dev/null 2>&1; then
        log_success "$description is already installed ($(command -v "$cmd"))"
        return 0
    fi
    
    if ! command -v cargo >/dev/null 2>&1; then
        log_warning "Cargo is not installed. Skipping $description"
        return 1
    fi
    
    log "Installing $description via cargo..."
    
    if [ "$DRY_RUN" -eq 1 ]; then
        log_info "[DRY RUN] Would run: cargo install $pkg"
        log_success "Would install $description"
    else
        verb "Running: cargo install $pkg"
        # Run cargo install with timeout to avoid hanging
        if timeout 300 cargo install "$pkg"; then
            log_success "$description installed"
        else
            log_warning "Failed to install $description (timed out after 5 minutes or error occurred)"
            return 1
        fi
    fi
}

# Install via npm
install_npm_pkg() {
    local cmd="$1"
    local pkg="$2"
    local description="$3"
    
    if command -v "$cmd" >/dev/null 2>&1; then
        log_success "$description is already installed ($(command -v "$cmd"))"
        return 0
    fi
    
    if ! command -v npm >/dev/null 2>&1; then
        log_warning "npm is not installed. Skipping $description"
        return 1
    fi
    
    log "Installing $description via npm..."
    
    if [ "$DRY_RUN" -eq 1 ]; then
        log_info "[DRY RUN] Would run: npm install -g $pkg"
        log_success "Would install $description"
    else
        verb "Running: npm install -g $pkg"
        # Run npm install with timeout to avoid hanging
        if timeout 300 npm install -g "$pkg"; then
            log_success "$description installed"
        else
            log_warning "Failed to install $description (timed out after 5 minutes or error occurred)"
            return 1
        fi
    fi
}

# Install via Go
install_go_pkg() {
    local cmd="$1"
    local pkg="$2"
    local description="$3"
    
    if command -v "$cmd" >/dev/null 2>&1; then
        log_success "$description is already installed ($(command -v "$cmd"))"
        return 0
    fi
    
    if ! command -v go >/dev/null 2>&1; then
        log_warning "Go is not installed. Skipping $description"
        return 1
    fi
    
    log "Installing $description via Go..."
    
    if [ "$DRY_RUN" -eq 1 ]; then
        log_info "[DRY RUN] Would run: go install $pkg@latest"
        log_success "Would install $description"
    else
        verb "Running: go install $pkg@latest"
        # Run go install with timeout to avoid hanging
        if timeout 300 go install "$pkg@latest"; then
            log_success "$description installed"
        else
            log_warning "Failed to install $description (timed out after 5 minutes or error occurred)"
            return 1
        fi
    fi
}

# Tool installation functions
install_git_tools() {
    log "${BLUE}Installing Git enhancement tools...${NC}"
    
    install_pkg "git" "git" "Git version control"
    install_pkg "hub" "hub" "GitHub CLI (legacy)"
    install_pkg "gh" "gh" "GitHub CLI"
    install_pkg "git-delta" "git-delta" "Better git diffs"
    install_pkg "lazygit" "lazygit" "Terminal Git UI"
    install_pkg "tig" "tig" "Text interface for Git"
    
    # Git aliases and configuration
    if [ "$DRY_RUN" -eq 0 ]; then
        verb "Setting up useful git configurations..."
        git config --global pull.rebase false 2>/dev/null || true
        git config --global init.defaultBranch main 2>/dev/null || true
        log_info "Git configurations updated"
    fi
}

install_search_tools() {
    log "${BLUE}Installing search and navigation tools...${NC}"
    
    install_pkg "rg" "ripgrep" "Fast text search (ripgrep)"
    install_pkg "fd" "fd-find" "Modern find replacement" || install_pkg "fdfind" "fd-find" "Modern find replacement (Debian)"
    install_pkg "fzf" "fzf" "Command-line fuzzy finder"
    install_pkg "bat" "bat" "Better cat with syntax highlighting"
    install_pkg "exa" "exa" "Modern ls replacement" || install_pkg "eza" "eza" "Modern ls replacement (exa fork)"
    install_pkg "tree" "tree" "Directory tree viewer"
    install_pkg "xxd" "vim" "Hexdump tool (via vim)"
    
    # Install the silver searcher
    if [ "$(detect_system)" = "arch" ]; then
        install_pkg "ag" "the_silver_searcher" "The Silver Searcher (ag)"
    else
        install_pkg "ag" "silversearcher-ag" "The Silver Searcher (ag)"
    fi
    
    # Install more advanced tools
    install_cargo_pkg "sd" "sd" "Intuitive find and replace"
    install_cargo_pkg "jql" "jql" "JSON query language"
    install_go_pkg "gum" "github.com/charmbracelet/gum" "Interactive CLI tool"
    
    # Setup shell integrations
    if [ "$DRY_RUN" -eq 0 ] && command -v fzf >/dev/null 2>&1; then
        verb "Setting up fzf shell integrations..."
        if [ -f /usr/share/doc/fzf/examples/key-bindings.bash ]; then
            log_info "FZF key-bindings available at /usr/share/doc/fzf/examples/key-bindings.bash"
        fi
    fi
}

install_text_tools() {
    log "${BLUE}Installing text processing and editor tools...${NC}"
    
    install_pkg "vim" "vim" "Vim text editor"
    install_pkg "nano" "nano" "Nano text editor"
    install_pkg "jq" "jq" "JSON processor"
    install_pkg "yq" "yq" "YAML processor"
    install_pkg "bc" "bc" "Command-line calculator"
    install_pkg "awk" "gawk" "AWK text processing"
    install_pkg "sed" "sed" "Stream editor"
    
    # Advanced text processing
    install_cargo_pkg "miniserve" "miniserve" "File server for local directories"
    install_npm_pkg "markdownlint-cli" "markdownlint-cli" "Markdown linter"
    
    # Install neovim if not already present
    if command -v nvim >/dev/null 2>&1; then
        log_success "Neovim is already installed ($(command -v nvim))"
    else
        install_pkg "nvim" "neovim" "Neovim text editor"
    fi
}

install_network_tools() {
    log "${BLUE}Installing network and monitoring tools...${NC}"
    
    install_pkg "curl" "curl" "HTTP client"
    install_pkg "wget" "wget" "File downloader"
    install_pkg "htop" "htop" "Process viewer"
    install_pkg "iftop" "iftop" "Network bandwidth monitor"
    install_pkg "nmap" "nmap" "Network scanner"
    install_pkg "netcat" "netcat-openbsd" "Network utility"
    if [ "$(detect_system)" = "arch" ]; then
        install_pkg "dig" "bind" "DNS tools (dig)"
    else
        install_pkg "dig" "dnsutils" "DNS tools (dig)" || install_pkg "dig" "bind-utils" "DNS tools (dig) for RedHat"
    fi
    install_pkg "iproute2" "iproute2" "Network utilities (ip)" || install_pkg "net-tools" "net-tools" "Legacy network tools"
    
    # Advanced monitoring
    install_pkg "iotop" "iotop" "I/O monitor"
    install_pkg "strace" "strace" "System call tracer"
    install_pkg "lsof" "lsof" "List open files"
    install_pkg "rsync" "rsync" "File synchronization"
    
    # Modern HTTP clients
    install_cargo_pkg "xh" "xh" "Modern HTTP client (curl alternative)"
    install_cargo_pkg "bandwhich" "bandwhich" "Network bandwidth utilization"
    install_go_pkg "httpie" "github.com/httpie/httpie" "Modern HTTP client"
}

install_sys_tools() {
    log "${BLUE}Installing system utilities...${NC}"
    
    install_pkg "tmux" "tmux" "Terminal multiplexer"
    install_pkg "screen" "screen" "Terminal multiplexer"
    install_pkg "unzip" "unzip" "Unzip utility"
    install_pkg "zip" "zip" "Zip utility"
    install_pkg "tar" "tar" "Archive utility"
    if [ "$(detect_system)" = "arch" ]; then
        install_pkg "7z" "7zip" "7-Zip support"
    else
        install_pkg "7z" "p7zip-full" "7-Zip support" || install_pkg "7z" "p7zip" "7-Zip support"
    fi
    install_pkg "file" "file" "File type detector"
    install_pkg "less" "less" "Terminal pager"
    install_pkg "moreutils" "moreutils" "Unix utilities"
    
    # System analysis
    install_pkg "ncdu" "ncdu" "NCurses disk usage"
    install_pkg "dfc" "dfc" "Disk usage utility" || install_pkg "pydf" "pydf" "Disk usage utility alternative"
    
    # Tiling window managers (optional but helpful for development)
    if command -v Xorg >/dev/null 2>&1; then
        install_pkg "wmctrl" "wmctrl" "Window manager control"
        install_pkg "xdotool" "xdotool" "X11 automation tool"
    fi
}

install_dev_tools() {
    log "${BLUE}Installing development tools...${NC}"
    
    # Programming languages and package managers
    install_pkg "python3" "python3" "Python 3"
    if command -v pip >/dev/null 2>&1; then
        log_success "Python package manager is already installed ($(command -v pip))"
    else
        install_pkg "pip" "python3-pip" "Python package manager"
    fi
    if command -v node >/dev/null 2>&1; then
        log_success "Node.js JavaScript runtime is already installed ($(command -v node))"
    else
        install_pkg "node" "nodejs" "Node.js JavaScript runtime"
    fi
    if command -v npm >/dev/null 2>&1; then
        log_success "Node.js package manager is already installed ($(command -v npm))"
    else
        install_pkg "npm" "npm" "Node.js package manager"
    fi
    
    # Build tools
    if [ "$(detect_system)" = "arch" ]; then
        # Check if base-devel group is installed
        if pacman -Qg base-devel >/dev/null 2>&1; then
            log_success "Development tools (base-devel) are already installed"
        else
            install_pkg "make" "base-devel" "Development tools (Arch)"
        fi
    elif [ "$(detect_system)" = "debian" ]; then
        install_pkg "gcc" "build-essential" "Development tools (Debian)"
    else
        install_pkg "gcc" "base-devel" "Development tools" || install_pkg "gcc" "build-essential" "Development tools"
    fi
    install_pkg "cmake" "cmake" "Build system"
    install_pkg "make" "make" "Make utility" 
    install_pkg "gcc" "gcc" "C compiler"
    install_pkg "g++" "g++" "C++ compiler"
    
    # Version control enhancement
    install_pkg "tldr" "tldr" "Simplified man pages" || install_npm_pkg "tldr" "tldr" "Simplified man pages (npm)"
    
    # Container tools
    if [ "$(detect_system)" = "arch" ]; then
        install_pkg "docker" "docker" "Docker container runtime"
    else
        install_pkg "docker" "docker.io" "Docker container runtime" || install_pkg "docker" "docker" "Docker container runtime"
    fi
    install_pkg "docker-compose" "docker-compose" "Docker Compose" || install_npm_pkg "docker-compose" "docker-compose" "Docker Compose (npm)"
    
    # Cloud CLIs
    install_go_pkg "aws" "github.com/aws/aws-cli/v2/cmd/aws" "AWS CLI"
    install_go_pkg "gcloud" "github.com/GoogleCloudPlatform/cloud-sdk-core/gcloud-go" "Google Cloud CLI" || install_pkg "google-cloud-sdk" "google-cloud-sdk" "Google Cloud SDK"
    
    # Local development servers
    install_cargo_pkg "dust" "dust" "Disk usage analyzer"
    install_cargo_pkg "watchexec" "watchexec" "Command execution on file changes"
    install_go_pkg "air" "github.com/cosmtrek/air" "Live reload for Go apps"
    
    # Package manager for shell
    install_npm_pkg "serve" "serve" "Static file server"
}

install_all_tools() {
    log "${BLUE}Installing all development tools...${NC}"
    install_git_tools
    install_search_tools
    install_text_tools
    install_network_tools
    install_sys_tools
    install_dev_tools
}

# Main execution
main() {
    log "${CYAN}🚀 Advanced Tools Installer for AI Development${NC}"
    log "${CYAN}================================================${NC}"
    
    # Auto-elevate to sudo if not running as root
    if [ "$EUID" -ne 0 ]; then
        log_info "Elevating to sudo privileges for system-wide installation..."
        # Preserve environment variables for colors and terminal capabilities
        exec sudo -E TERM="$TERM" "$0" "$@"
    fi
    
    # Detect system
    local system=$(detect_system)
    local pm=$(get_package_manager)
    
    log_info "Detected system: $system"
    log_info "Package manager: $pm"
    log_info "Installation category: $TO_INSTALL_CATEGORY"
    
    if [ "$DRY_RUN" -eq 1 ]; then
        log_warning "DRY RUN MODE - No packages will be installed"
    fi
    
    # Update package manager
    update_package_manager
    
    # Install tools based on category
    case $TO_INSTALL_CATEGORY in
        all|ALL)
            install_all_tools
            ;;
        git|Git|GIT)
            install_git_tools
            ;;
        search|Search|SEARCH)
            install_search_tools
            ;;
        text|Text|TEXT)
            install_text_tools
            ;;
        network|Network|NETWORK)
            install_network_tools
            ;;
        sys|Sys|SYS)
            install_sys_tools
            ;;
        dev|Dev|DEV)
            install_dev_tools
            ;;
        *)
            log_error "Unknown category: $TO_INSTALL_CATEGORY"
            log_info "Valid categories: all, git, search, text, network, sys, dev"
            exit 1
            ;;
    esac
    
    if [ "$DRY_RUN" -eq 0 ]; then
        log_success "🎉 Installation completed!"
        log_info "You may need to restart your shell or run 'source ~/.bashrc' for some tools to be available."
    else
        log_success "🎉 Dry run completed! Use without --dry-run to actually install packages."
    fi
}

# Run main function
main "$@"