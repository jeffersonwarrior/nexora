#!/bin/bash

# Install script for Nexora
# Usage: ./install.sh

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

DEFAULT_VERSION="0.29.0"
BINARY_NAME="nexora"
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

print_status "Installing Nexora v${VERSION} to $HOME/.local/bin..."

mkdir -p "$HOME/.local/bin"
# Check both current PATH and shell config files
CONFIG_FILE=""
NEEDS_UPDATE=false
CONFIG_FILE_BASENAME=""

# Determine the appropriate shell configuration file
if [[ "$SHELL" == */bash ]]; then
    if [ -f ~/.bashrc ]; then
        CONFIG_FILE="$HOME/.bashrc"
        CONFIG_FILE_BASENAME=".bashrc"
    elif [ -f ~/.bash_profile ]; then
        CONFIG_FILE="$HOME/.bash_profile"
        CONFIG_FILE_BASENAME=".bash_profile"
    else
        CONFIG_FILE="$HOME/.profile"
        CONFIG_FILE_BASENAME=".profile"
    fi
elif [[ "$SHELL" == */zsh ]]; then
    CONFIG_FILE="$HOME/.zshrc"
    CONFIG_FILE_BASENAME=".zshrc"
else
    # Fallback to .profile for other shells
    CONFIG_FILE="$HOME/.profile"
    CONFIG_FILE_BASENAME=".profile"
fi

# Create the config file if it doesn't exist
if [ ! -f "$CONFIG_FILE" ]; then
    touch "$CONFIG_FILE"
    print_status "Created $CONFIG_FILE"
fi

# Store whether we updated the PATH for later use
PATH_UPDATED=false
if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
    NEEDS_UPDATE=true
fi

if [ "$NEEDS_UPDATE" = true ]; then
    # Add PATH export to the appropriate config file
    echo "" >> "$CONFIG_FILE"
    echo "# Nexora PATH addition" >> "$CONFIG_FILE"
    echo "export PATH=\"\$PATH:$HOME/.local/bin\"" >> "$CONFIG_FILE"
    
    print_status "Added $HOME/.local/bin to PATH in $CONFIG_FILE"
    print_status "You need to restart your terminal or run 'source $CONFIG_FILE' to use nexora"
    PATH_UPDATED=true
    
export PATH="$PATH:$HOME/.local/bin"
    print_status "$HOME/.local/bin is already in PATH"
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

# Install Go if not present
install_go() {
    print_status "Installing Go..."
    
    # Detect OS and architecture
    GO_OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    GO_ARCH=$(uname -m)
    
    # Map architecture names to Go's naming convention
    case "$GO_ARCH" in
        x86_64) GO_ARCH="amd64" ;;
        aarch64|arm64) GO_ARCH="arm64" ;;
        armv7l) GO_ARCH="armv6l" ;;
        *) 
            print_error "Unsupported architecture: $GO_ARCH"
            return 1
            ;;
    esac
    
    # Set Go version to install (using a recent stable version)
    GO_VERSION="1.21.5"
    GO_TAR="go${GO_VERSION}.${GO_OS}-${GO_ARCH}.tar.gz"
    GO_URL="https://go.dev/dl/${GO_TAR}"
    GO_INSTALL_DIR="$HOME/.local/go"
    
    print_status "Downloading Go ${GO_VERSION} for ${GO_OS}-${GO_ARCH}..."
    
    # Create temp directory for download
    mkdir -p "$TEMP_DIR"
    cd "$TEMP_DIR"
    
    # Download Go
    if command -v wget >/dev/null 2>&1; then
        if ! wget -q --show-progress "$GO_URL"; then
            print_error "Failed to download Go using wget"
            return 1
        fi
    elif command -v curl >/dev/null 2>&1; then
        if ! curl -fsSL "$GO_URL" -o "$GO_TAR"; then
            print_error "Failed to download Go using curl"
            return 1
        fi
    else
        print_error "Neither wget nor curl is available. Cannot download Go."
        return 1
    fi
    
    # Remove old Go installation if it exists
    if [ -d "$GO_INSTALL_DIR" ]; then
        print_status "Removing existing Go installation..."
        rm -rf "$GO_INSTALL_DIR"
    fi
    
    # Extract Go
    print_status "Extracting Go to $GO_INSTALL_DIR..."
    if ! tar -C "$HOME/.local" -xzf "$GO_TAR"; then
        print_error "Failed to extract Go"
        return 1
    fi
    
    # Remove the downloaded tarball
    rm -f "$GO_TAR"
    
    # Add Go to PATH in shell config
    GO_BIN_DIR="$GO_INSTALL_DIR/bin"
    if ! grep -q "$GO_BIN_DIR" "$CONFIG_FILE"; then
        echo "" >> "$CONFIG_FILE"
        echo "# Go PATH addition" >> "$CONFIG_FILE"
        echo "export PATH=\"\$PATH:$GO_BIN_DIR\"" >> "$CONFIG_FILE"
        print_status "Added Go to PATH in $CONFIG_FILE"
    fi
    
    # Export to current session
    export PATH="$PATH:$GO_BIN_DIR"
    
    # Verify installation
    if [ -x "$GO_BIN_DIR/go" ]; then
        INSTALLED_VERSION=$("$GO_BIN_DIR/go" version | awk '{print $3}' | sed 's/go//')
        print_status "Go $INSTALLED_VERSION installed successfully!"
        return 0
    else
        print_error "Go installation verification failed"
        return 1
    fi
}

# Build Nexora
build_nexora() {
    # Ensure we're in the project directory (find go.mod file)
    SCRIPT_DIR="$(dirname "$0")"
    if [ -f "$SCRIPT_DIR/go.mod" ]; then
        cd "$SCRIPT_DIR"
    else
        # If go.mod not in script dir, try current directory
        if [ -f "./go.mod" ]; then
            cd "."
        else
            print_error "Could not find go.mod file. Make sure you're running this script from the project directory."
            return 1
        fi
    fi
    
    # Update dependencies before building (suppress output)
    if ! go mod tidy >/dev/null 2>&1 && go get -u ./... >/dev/null 2>&1 && go mod tidy >/dev/null 2>&1; then
        print_warning "Failed to update some dependencies, using existing ones..."
    fi
    
    # Set up build flags
    LDFLAGS="-X github.com/nexora/cli/internal/version.Version=${VERSION}"
    
    # Debug output
    print_status "Building with output path: $HOME/.local/bin/nexora"
    print_status "Current directory: $(pwd)"
    print_status "HOME is set to: $HOME"
    
    # Build the binary directly to the installation directory
    if ! go build -ldflags="${LDFLAGS}" -o "$HOME/.local/bin/nexora" .; then
        return 1
    fi
    
    # Check if binary was created
    if [ ! -f "$HOME/.local/bin/nexora" ]; then
        return 1
    fi
    
    # Return the path to the built binary
    echo "$HOME/.local/bin/nexora"
}

# Remove any existing Nexora installations
remove_existing() {
    print_status "Removing any existing Nexora installations..."
    
    # Remove from installation directory
    if [ -f "$HOME/.local/bin/nexora" ]; then
        rm -f "$HOME/.local/bin/nexora" 2>/dev/null || true
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
    
    print_status "Installing nexora to $HOME/.local/bin..."
    
    # Ensure the install directory exists
    mkdir -p "$HOME/.local/bin"
    
    # Set executable permissions - fix issue with colored output
    # Remove any ANSI escape codes from the path
    clean_path=$(echo "$binary_path" | sed 's/\x1b\[[0-9;]*m//g')
    chmod +x "$clean_path"
    
    print_status "$BINARY_NAME installed to $clean_path"
}

# Verify installation
verify_installation() {
    print_status "Verifying installation..."
    
    # First check if the binary exists in the install directory
    if [ ! -f "$HOME/.local/bin/nexora" ]; then
        print_error "Binary not found at $HOME/.local/bin/nexora"
        return 1
    fi
    
    # Then check if it's in PATH
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
        print_status "Installation directory: $HOME/.local/bin"
        
        # Check if binary exists where we installed it
        if [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
            print_warning "Binary exists at $HOME/.local/bin/nexora but not in PATH."
            print_status "Please restart your terminal or run 'source ~/.bashrc' (or ~/.zshrc for zsh users)"
            print_status "Alternatively, run: export PATH=\"\$PATH:$HOME/.local/bin\""
            
            # Try to add it to PATH for this session
            export PATH="$PATH:$HOME/.local/bin"
            
            # Verify again after adding to PATH
            if command -v "$BINARY_NAME" >/dev/null 2>&1; then
                VERSION_OUTPUT=$("$BINARY_NAME" --version 2>/dev/null || echo "Version check failed")
                print_status "Binary found after adding to PATH for current session."
                print_status "Version: $VERSION_OUTPUT"
                print_status "Remember to restart your terminal or source your shell configuration!"
            else
                return 1
            fi
        else
            return 1
        fi
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
            sudo apt-get update -qq 2>&1 | grep -v "Error connecting" || true
            # Sometimes some repos fail but the rest work, so we ignore connection errors
            
            print_status "Installing tools with apt..."
            # First try to install all tools
            sudo apt-get install -y -qq ripgrep fd-find bat fzf exa jq 2>/dev/null || {
                # If exa fails, try installing eza instead (exa replacement)
                print_warning "exa package not available, trying eza (modern replacement)..."
                sudo apt-get install -y -qq ripgrep fd-find bat fzf eza jq || {
                    # Try individual installations with more specific packages
                    print_warning "Some tools failed to install, trying individually..."
                    
                    # ripgrep
                    sudo apt-get install -y -qq ripgrep 2>/dev/null || print_warning "Could not install ripgrep"
                    
                    # fd-find (available as fd-find in Ubuntu)
                    sudo apt-get install -y -qq fd-find 2>/dev/null || print_warning "Could not install fd-find"
                    
                    # bat (available as bat in Ubuntu 24.04)
                    sudo apt-get install -y -qq bat 2>/dev/null || {
                        # Try batcat if bat not available
                        sudo apt-get install -y -qq batcat 2>/dev/null || print_warning "Could not install bat"
                    }
                    
                    # fzf
                    sudo apt-get install -y -qq fzf 2>/dev/null || print_warning "Could not install fzf"
                    
                    # Try eza (exa replacement) first
                    sudo apt-get install -y -qq eza 2>/dev/null || {
                        # Fall back to the old lsd if available
                        sudo apt-get install -y -qq lsd 2>/dev/null || print_warning "Could not install eza/lshw/lsd (file listing tool)"
                    }
                    
                    # jq
                    sudo apt-get install -y -qq jq 2>/dev/null || print_warning "Could not install jq"
                    
                    print_error "Some tools could not be installed"
                }
            }
            
            # Create symlinks for better names
            [ -f /usr/bin/fdfind ] && [ ! -f ~/.local/bin/fd ] && ln -s /usr/bin/fdfind ~/.local/bin/fd
            
            # Handle bat which might be installed as bat or batcat
            if [ -f /usr/bin/bat ] && [ ! -f ~/.local/bin/bat ]; then
                ln -s /usr/bin/bat ~/.local/bin/bat
            elif [ -f /usr/bin/batcat ] && [ ! -f ~/.local/bin/bat ]; then
                ln -s /usr/bin/batcat ~/.local/bin/bat
            fi
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
        print_status "Go not found or version too old. Installing Go..."
        if ! install_go; then
            print_error "Go installation failed. Please install Go manually and try again."
            return 1
        fi
    fi
    
    # Remove any existing installations
    remove_existing || { print_error "Failed to remove existing installations"; return 1; }
    
    # Build Nexora directly to the install directory
    print_status "Building Nexora with version: ${VERSION}..."
    print_status "Updating dependencies to latest versions..."
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
    print_status ""
    
    if [ "$PATH_UPDATED" = true ]; then
        print_status "To start using nexora immediately, run: source ~/$CONFIG_FILE_BASENAME"
        print_status "(Or restart your terminal to automatically load the configuration)"
    fi
    
    return 0
}

# Run main function
main "$@" || exit 1