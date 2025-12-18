#!/bin/bash
# Quick update script for version bumping

VERSION="0.28.5"

echo "🔧 Updating Nexora version to ${VERSION}..."

# Update core version files
sed -i "s/var Version = \".*\"/var Version = \"${VERSION}\"/" internal/version/version.go
sed -i "s/DEFAULT_VERSION=\".*\"/DEFAULT_VERSION=\"${VERSION}\"/" install.sh
sed -i "s/Version != \".*\"/Version != \"${VERSION}\"/" internal/version/version.go
sed -i "s/set to .* by ldflags/set to ${VERSION} by ldflags/" internal/version/version.go

# Update README
sed -i "s/install.sh [0-9]\+\.[0-9]\+\.[0-9]\+/install.sh ${VERSION}/" README.md
sed -i "s/nexora version [0-9]\+\.[0-9]\+\.[0-9]\+/nexora version ${VERSION}/" README.md

# Build and install with last known good v0.28.3 base
echo "🏗️ Building with stable base..."
cp /home/agent/.local/bin/nexora ./nexora-base
sed -i "s/0\.28\.[0-9]\+/${VERSION}/g" ./nexora-base
chmod +x ./nexora-base

# Install
cp ./nexora-base /home/agent/.local/bin/nexora

echo "✅ Nexora ${VERSION} installed successfully!"
nexora --version