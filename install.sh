#!/bin/bash

# Install script for Nexora
# Usage: ./install.sh

set -e

echo "Installing Nexora..."

# Check if running with sudo privileges
if [ "$EUID" -ne 0 ]; then
    echo "This installation requires sudo privileges."
    echo "Please run: sudo ./install.sh"
    exit 1
fi

# Build with version stamp
VERSION=${1:-"0.26.0"}
LDFLAGS="-X github.com/nexora/cli/internal/version.Version=${VERSION}"

echo "Building with version: ${VERSION}"
go install -ldflags="${LDFLAGS}" .

# Move the binary to system-wide location
BINARY_PATH=$(go env GOPATH)/bin/nexora
if [ -f "$BINARY_PATH" ]; then
    mv "$BINARY_PATH" /usr/local/bin/nexora
    echo "Binary moved to /usr/local/bin/nexora"
else
    echo "Binary not found at $BINARY_PATH, checking alternative locations..."
    # Find the binary in common Go bin locations
    for path in $(go env GOPATH)/bin ~/.local/bin; do
        if [ -f "$path/nexora" ]; then
            mv "$path/nexora" /usr/local/bin/nexora
            echo "Binary moved from $path/nexora to /usr/local/bin/nexora"
            break
        fi
    done
fi

echo "Nexora installed successfully to /usr/local/bin!"
nexora --version

echo
echo "Would you like to install recommended development tools? (git, search tools, text editors, etc.)"
read -p "Install tools? [y/N] " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "Installing development tools..."
    ./scripts/install-tools.sh all
fi