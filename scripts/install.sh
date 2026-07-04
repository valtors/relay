#!/bin/sh
set -e

# Relay installer - downloads the latest binary for your platform
# Usage: curl -fsSL https://raw.githubusercontent.com/valtors/relay/main/scripts/install.sh | sh

REPO="valtors/relay"
INSTALL_DIR="/usr/local/bin"

# Detect OS and arch
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case "$OS" in
    linux) OS="linux" ;;
    darwin) OS="darwin" ;;
    *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

# Get latest release tag
LATEST=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"v?([^"]+)".*/\1/')

if [ -z "$LATEST" ]; then
    echo "Could not determine latest version. Check https://github.com/$REPO/releases"
    exit 1
fi

# Download
URL="https://github.com/$REPO/releases/download/v${LATEST}/relay_${LATEST}_${OS}_${ARCH}.tar.gz"
echo "Downloading relay v${LATEST} for ${OS}/${ARCH}..."

TMP=$(mktemp -d)
curl -fsSL "$URL" -o "$TMP/relay.tar.gz"
tar -xzf "$TMP/relay.tar.gz" -C "$TMP"

# Install
if [ -w "$INSTALL_DIR" ]; then
    mv "$TMP/relay" "$INSTALL_DIR/relay"
else
    echo "Installing to $INSTALL_DIR (requires sudo)..."
    sudo mv "$TMP/relay" "$INSTALL_DIR/relay"
fi

rm -rf "$TMP"
chmod +x "$INSTALL_DIR/relay"

echo ""
echo "relay v${LATEST} installed to $INSTALL_DIR/relay"
echo "Run: relay"
