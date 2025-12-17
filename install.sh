#!/bin/bash

# Install script for Nexora
# Usage: ./install.sh

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
BINARY_NAME="nexora"
DEFAULT_VERSION="0.28.0"
INSTALL_DIR="$HOME/.local/bin"
TEMP_DIR="/tmp/nexora-install"

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Parse command line arguments
VERSION=${1:-$DEFAULT_VERSION}

print_status "Installing Nexora v${VERSION} to $INSTALL_DIR..."

# Ensure the installation directory exists
mkdir -p "$INSTALL_DIR"

# Add to PATH if not already there
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo "export PATH=\"\$PATH:$INSTALL_DIR\"" >> ~/.profile
    print_status "Added $INSTALL_DIR to PATH in ~/.profile"
    export PATH="$PATH:$INSTALL_DIR"
fi

# Check if Go is installed
check_go() {
    if command -v go >/dev/null 2>&1; then
        GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
        print_status "Go is already installed (version: $GO_VERSION)"
        
        # Check if Go version is adequate (>=1.19)
        if go version | grep -E 'go1\.(1[9]|[2-9][0-9])' >/dev/null 2>&1; then
            return 0
        else
            print_warning "Go version is older than recommended (>=1.19). Consider upgrading."
            return 1
        fi
    else
        return 1
    fi
}

# Build Nexora
build_nexora() {
    # Don't print status here to avoid mixing output with the path
    
    # Ensure we're in the project directory
    cd "$(dirname "$0")"
    
    # Set up build flags
    LDFLAGS="-X github.com/nexora/cli/internal/version.Version=${VERSION}"
    
    # Build the binary directly to the installation directory
    if ! go build -ldflags="${LDFLAGS}" -o "$INSTALL_DIR/nexora" .; then
        return 1
    fi
    
    # Check if binary was created
    if [ ! -f "$INSTALL_DIR/nexora" ]; then
        return 1
    fi
    
    # Return the path to the built binary
    echo "$INSTALL_DIR/nexora"
}

# Remove any existing Nexora installations
remove_existing() {
    print_status "Removing any existing Nexora installations..."
    
    # Remove from installation directory
    if [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
        rm -f "$INSTALL_DIR/$BINARY_NAME" 2>/dev/null || true
    fi
    
    # Get GOPATH safely
    GOPATH_BIN=""
    if GOPATH=$(go env GOPATH 2>/dev/null); then
        GOPATH_BIN="$GOPATH/bin"
    fi
    
    # Also remove from other directories if they exist
    for path in "$HOME/bin" "$GOPATH_BIN"; do
        if [ -d "$path" ]; then
            # Remove both possible binary names, ignore errors
            [ -f "$path/$BINARY_NAME" ] && rm -f "$path/$BINARY_NAME"
            [ -f "$path/nexora" ] && rm -f "$path/nexora"
            [ -f "$path/cli" ] && rm -f "$path/cli"
        fi
    done
    
    return 0
}

# Install the binary
install_binary() {
    local binary_path="$1"
    
    print_status "Installing $BINARY_NAME to $INSTALL_DIR..."
    
    # Ensure the install directory exists
    mkdir -p "$INSTALL_DIR"
    
    # Set executable permissions - fix issue with colored output
    # Remove any ANSI escape codes from the path
    clean_path=$(echo "$binary_path" | sed 's/\x1b\[[0-9;]*m//g')
    chmod +x "$clean_path"
    
    print_status "$BINARY_NAME installed to $clean_path"
}

# Verify installation
verify_installation() {
    print_status "Verifying installation..."
    
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        VERSION_OUTPUT=$("$BINARY_NAME" --version 2>/dev/null || echo "Version check failed")
        BINARY_LOCATION=$(which "$BINARY_NAME")
        print_status "Installation successful!"
        print_status "Binary location: $BINARY_LOCATION"
        print_status "Version: $VERSION_OUTPUT"
    else
        print_error "Installation verification failed. $BINARY_NAME not found in PATH."
        
        # Show PATH for debugging
        print_status "Current PATH: $PATH"
        print_status "Installation directory: $INSTALL_DIR"
        
        # Check if binary exists where we installed it
        if [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
            print_warning "Binary exists at $INSTALL_DIR/$BINARY_NAME but not in PATH."
            print_status "You may need to update your PATH or source your shell profile."
        fi
        
        return 1
    fi
    
    return 0
}

# Install better development tools
install_better_tools() {
    print_status "Installing faster development tools..."
    
    # Detect package manager
    if command -v apt-get >/dev/null 2>&1; then
        PM="apt"
        print_status "Detected package manager: apt (Debian/Ubuntu)"
    elif command -v dnf >/dev/null 2>&1; then
        PM="dnf"
        print_status "Detected package manager: dnf (Fedora)"
    elif command -v yum >/dev/null 2>&1; then
        PM="yum"
        print_status "Detected package manager: yum (RHEL/CentOS)"
    elif command -v pacman >/dev/null 2>&1; then
        PM="pacman"
        print_status "Detected package manager: pacman (Arch)"
    elif command -v brew >/dev/null 2>&1; then
        PM="brew"
        print_status "Detected package manager: Homebrew"
    elif command -v pkg >/dev/null 2>&1; then
        PM="pkg"
        print_status "Detected package manager: pkg (FreeBSD)"
    else
        print_warning "Unsupported package manager. You may need to install tools manually."
        return
    fi
    
    # Install tools based on package manager
    case $PM in
        "apt")
            print_status "Updating package list..."
            sudo apt-get update -qq || true
            
            print_status "Installing tools with apt..."
            sudo apt-get install -y -qq ripgrep fd-find bat fzf exa jq || {
                print_error "Failed to install some tools with apt"
            }
            
            # Create symlinks for better names
            [ -f /usr/bin/fdfind ] && [ ! -f ~/.local/bin/fd ] && ln -s /usr/bin/fdfind ~/.local/bin/fd
            [ -f /usr/bin/batcat ] && [ ! -f ~/.local/bin/bat ] && ln -s /usr/bin/batcat ~/.local/bin/bat
            ;;
        "dnf"|"yum")
            print_status "Installing tools with dnf/yum..."
            sudo $PM install -y ripgrep fd-find bat fzf exa jq || {
                print_error "Failed to install some tools with dnf/yum"
            }
            ;;
        "pacman")
            print_status "Installing tools with pacman..."
            sudo pacman -S --noconfirm ripgrep fd bat fzf exa jq || {
                print_error "Failed to install some tools with pacman"
            }
            ;;
        "brew")
            print_status "Installing tools with Homebrew..."
            brew install ripgrep fd bat fzf exa jq || {
                print_error "Failed to install some tools with brew"
            }
            ;;
        "pkg")
            print_status "Installing tools with pkg..."
            sudo pkg install -y ripgrep fd bat fzf jq || {
                print_error "Failed to install some tools with pkg"
            }
            ;;
    esac
    
    print_status "Faster tools installation completed!"
    print_status "Tools installed: ripgrep (rg), fd-find (fd), bat, fzf, exa, jq"
}

# Main installation flow
main() {
    # Check if Go is installed
    if ! check_go; then
        print_error "Go is required to build Nexora. Please install Go manually and try again."
        return 1
    fi
    
    # Remove any existing installations
    remove_existing || { print_error "Failed to remove existing installations"; return 1; }
    
    # Build Nexora directly to the install directory
    print_status "Building Nexora with version: ${VERSION}..."
    if ! BINARY_PATH=$(build_nexora); then
        print_error "Failed to build Nexora binary"
        return 1
    fi
    
    # Install the binary (set permissions)
    install_binary "$BINARY_PATH"
    
    # Verify installation
    if ! verify_installation; then
        print_error "Installation verification failed"
        return 1
    fi
    
    echo
    print_status "Would you like to install faster development tools? (ripgrep, fd-find, bat, fzf, etc.)"
    read -p "Install tools? [y/N] " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        install_better_tools
    fi
    
    print_status "Nexora v${VERSION} installation completed successfully!"
    print_status "You can now run: $BINARY_NAME --help"
    
    return 0
}

# Run main function
main "$@" || exit 1