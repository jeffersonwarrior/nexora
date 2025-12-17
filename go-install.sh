#!/bin/bash

# Alternative installation script that works like 'go install .' but installs to /usr/local/bin
# Usage: ./go-install.sh

set -e

echo "Building Nexora..."

# Build with version stamp
VERSION=${1:-"0.28.0"}
LDFLAGS="-X github.com/nexora/cli/internal/version.Version=${VERSION}"

# Create temporary build
go build -ldflags="${LDFLAGS}" -o /tmp/nexora-build .

# Check if we have write permissions to /usr/local/bin
if [ -w "/usr/local/bin" ]; then
    echo "Installing to /usr/local/bin/nexora..."
    mv /tmp/nexora-build /usr/local/bin/nexora
    chmod +x /usr/local/bin/nexora
    echo "Nexora installed successfully to /usr/local/bin!"
else
    echo "Need sudo privileges to install to /usr/local/bin"
    echo "Installing with sudo..."
    sudo mv /tmp/nexora-build /usr/local/bin/nexora
    sudo chmod +x /usr/local/bin/nexora
    echo "Nexora installed successfully to /usr/local/bin!"
fi

# Verify installation
if command -v nexora &> /dev/null; then
    nexora --version
    echo "Installation complete!"
else
    echo "Warning: nexora not found in PATH after installation"
    echo "You may need to add /usr/local/bin to your PATH or restart your terminal"
fi