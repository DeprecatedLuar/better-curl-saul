#!/bin/bash

# Better-Curl-Saul Install Script
# Usage: curl -sSL https://raw.githubusercontent.com/DeprecatedLuar/better-curl-saul/main/install.sh | bash

set -e

REPO="DeprecatedLuar/better-curl-saul"
BINARY_NAME="saul"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    armv7l) ARCH="arm" ;;
esac

# Try to download from releases first
echo "Checking for latest release..."
LATEST_RELEASE=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' || echo "")

if [ -n "$LATEST_RELEASE" ]; then
    DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_RELEASE/${BINARY_NAME}-${OS}-${ARCH}"

    echo "Found release $LATEST_RELEASE"
    echo "Downloading $BINARY_NAME for $OS-$ARCH..."

    if curl -L -o "$BINARY_NAME" "$DOWNLOAD_URL" 2>/dev/null; then
        chmod +x "$BINARY_NAME"
        echo "Download successful!"
    else
        echo "Release download failed, falling back to local build..."
        LATEST_RELEASE=""
    fi
fi

# Fallback to local build if no release or download failed
if [ -z "$LATEST_RELEASE" ]; then
    echo "Building from source..."

    # Check if we're in the repo directory
    if [ ! -f "go.mod" ] || [ ! -f "cmd/main.go" ]; then
        echo "Error: Not in better-curl-saul repository directory"
        echo "Either:"
        echo "1. Clone the repo: git clone https://github.com/$REPO.git"
        echo "2. Use the one-liner: curl -sSL https://raw.githubusercontent.com/$REPO/main/install.sh | bash"
        exit 1
    fi

    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        echo "Error: Go is not installed. Please install Go first."
        exit 1
    fi

    echo "Building $BINARY_NAME..."
    go build -o "$BINARY_NAME" cmd/main.go

    if [ $? -ne 0 ]; then
        echo "Build failed!"
        exit 1
    fi

    echo "Build successful!"
fi

# Install binary
echo "Installing to /usr/local/bin/..."
sudo cp "$BINARY_NAME" /usr/local/bin/
sudo chmod +x "/usr/local/bin/$BINARY_NAME"

# Clean up
rm -f "$BINARY_NAME"

echo "Installation complete! Test with: $BINARY_NAME version"