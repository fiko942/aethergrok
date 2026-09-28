#!/usr/bin/env bash
# AetherGrok Desktop & CLI Installer Script for macOS & Linux
# Usage: curl -fsSL https://raw.githubusercontent.com/fiko942/aethergrok/main/scripts/install-app.sh | bash

set -euo pipefail

REPO="fiko942/aethergrok"
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

echo "==> Detecting system architecture: $OS ($ARCH)..."

case "$ARCH" in
    x86_64|amd64)
        PKG_ARCH="amd64"
        ;;
    arm64|aarch64)
        PKG_ARCH="arm64"
        ;;
    *)
        echo "Error: Unsupported architecture: $ARCH" >&2
        exit 1
        ;;
esac

if [ "$OS" = "darwin" ]; then
    echo "==> Fetching latest release metadata from GitHub ($REPO)..."
    LATEST_TAG=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    
    if [ -z "$LATEST_TAG" ]; then
        LATEST_TAG="v1.0.3"
    fi
    
    CLEAN_VER="${LATEST_TAG#v}"
    DMG_NAME="AetherGrok-${CLEAN_VER}-macOS-${PKG_ARCH}.dmg"
    DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_TAG/$DMG_NAME"
    TMP_DIR="$(mktemp -d)"
    DMG_PATH="$TMP_DIR/$DMG_NAME"
    
    echo "==> Downloading AetherGrok Desktop ($LATEST_TAG) for macOS ($PKG_ARCH)..."
    curl -fSL "$DOWNLOAD_URL" -o "$DMG_PATH"
    
    echo "==> Mounting $DMG_NAME..."
    MOUNT_DIR="$TMP_DIR/mount"
    mkdir -p "$MOUNT_DIR"
    hdiutil attach "$DMG_PATH" -nobrowse -mountpoint "$MOUNT_DIR" -quiet
    
    echo "==> Installing AetherGrok.app to /Applications..."
    if [ -d "/Applications/AetherGrok.app" ]; then
        rm -rf "/Applications/AetherGrok.app"
    fi
    cp -R "$MOUNT_DIR/AetherGrok.app" /Applications/
    
    echo "==> Unmounting installer..."
    hdiutil detach "$MOUNT_DIR" -quiet || true
    rm -rf "$TMP_DIR"
    
    echo ""
    echo "✨ AetherGrok Desktop ($LATEST_TAG) successfully installed to /Applications/AetherGrok.app!"
    echo "🚀 You can launch it from Spotlight, Launchpad, or by running: open -a AetherGrok"
else
    echo "==> Detected Linux ($PKG_ARCH)."
    echo "ℹ️  Please visit https://github.com/$REPO/releases/latest to download package assets."
fi
