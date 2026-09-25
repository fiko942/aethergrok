#!/usr/bin/env bash

# ==============================================================================
# AetherGrok Desktop GUI - Production macOS Installer Builder (.app & .dmg)
# Author: Wiji Fiko Teren
# ==============================================================================

set -euo pipefail

# ANSI color output
CYAN='\033[0;36m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BOLD='\033[1m'
NC='\033[0m'

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BUILD_DIR="$PROJECT_ROOT/build/bin"
FRONTEND_DIR="$PROJECT_ROOT/frontend"
APP_NAME="AetherGrok"
BUNDLE_ID="com.wiji.aethergrok"
VERSION="1.0.0"

# Target architecture: arm64 (default for Apple Silicon), amd64 (Intel), or universal
TARGET_ARCH="${1:-arm64}"

export PATH="/opt/homebrew/bin:/usr/local/bin:$(go env GOPATH 2>/dev/null || echo "$HOME/go")/bin:$PATH"

echo -e "${CYAN}================================================================${NC}"
echo -e "${CYAN}${BOLD}   🚀 AetherGrok Production Builder for macOS (${TARGET_ARCH}) 🚀   ${NC}"
echo -e "${CYAN}================================================================${NC}"

# 1. Check prerequisites
echo -e "${YELLOW}Step 1/5: Verifying build dependencies...${NC}"

if ! command -v go &> /dev/null; then
  echo -e "${RED}Error: 'go' is not found in PATH.${NC}"
  exit 1
fi

if ! command -v pnpm &> /dev/null && ! command -v npm &> /dev/null; then
  echo -e "${RED}Error: Neither 'pnpm' nor 'npm' is found in PATH.${NC}"
  exit 1
fi

if ! command -v wails &> /dev/null; then
  echo -e "${YELLOW}Installing/Updating Wails v2 CLI...${NC}"
  go install github.com/wailsapp/wails/v2/cmd/wails@latest
fi

if ! command -v hdiutil &> /dev/null; then
  echo -e "${RED}Error: 'hdiutil' is required on macOS to package DMG files.${NC}"
  exit 1
fi

echo -e "${GREEN}✓ All build prerequisites verified.${NC}"

# 2. Build Frontend Production Bundle
echo -e "\n${YELLOW}Step 2/5: Compiling Vite Frontend Production Bundle...${NC}"
cd "$FRONTEND_DIR"

if command -v pnpm &> /dev/null; then
  pnpm run build
else
  npm run build
fi

echo -e "${GREEN}✓ Frontend compiled successfully.${NC}"

# 3. Build macOS .app via Wails
echo -e "\n${YELLOW}Step 3/5: Compiling macOS Native Application with Wails...${NC}"
cd "$PROJECT_ROOT"

PLATFORM_ARG="darwin/${TARGET_ARCH}"
if [ "$TARGET_ARCH" = "universal" ]; then
  PLATFORM_ARG="darwin/universal"
fi

wails build \
  -platform "$PLATFORM_ARG" \
  -clean \
  -ldflags "-s -w" \
  -webview2 embed \
  -nsis=false

echo -e "${GREEN}✓ Wails macOS bundle compiled.${NC}"

# 4. Prepare and package .dmg Installer
echo -e "\n${YELLOW}Step 4/5: Creating macOS .dmg Installer...${NC}"

APP_BUNDLE="$BUILD_DIR/aethergrok.app"
if [ ! -d "$APP_BUNDLE" ]; then
  # Check if case-sensitive named AetherGrok.app
  if [ -d "$BUILD_DIR/AetherGrok.app" ]; then
    APP_BUNDLE="$BUILD_DIR/AetherGrok.app"
  else
    echo -e "${RED}Error: Application bundle not found in $BUILD_DIR${NC}"
    exit 1
  fi
fi

# DMG staging directory
DMG_STAGE_DIR=$(mktemp -d /tmp/aethergrok_dmg_stage.XXXXXX)
DMG_OUTPUT="$BUILD_DIR/AetherGrok-${VERSION}-macOS-${TARGET_ARCH}.dmg"

# Clean prior DMG if exists
rm -f "$DMG_OUTPUT"

# Copy App into staging
echo "Staging application bundle..."
cp -R "$APP_BUNDLE" "$DMG_STAGE_DIR/AetherGrok.app"

# Create symlink to /Applications for Drag & Drop install
ln -s /Applications "$DMG_STAGE_DIR/Applications"

echo "Creating compressed DMG installer with hdiutil..."
hdiutil create \
  -volname "$APP_NAME" \
  -srcfolder "$DMG_STAGE_DIR" \
  -ov \
  -format UDZO \
  "$DMG_OUTPUT"

# Cleanup staging
rm -rf "$DMG_STAGE_DIR"

# 5. Generate Checksum
echo -e "\n${YELLOW}Step 5/5: Generating SHA256 Checksum...${NC}"
cd "$BUILD_DIR"
shasum -a 256 "$(basename "$DMG_OUTPUT")" > "$(basename "$DMG_OUTPUT").sha256"

echo -e "\n${GREEN}================================================================${NC}"
echo -e "${GREEN}${BOLD}   🎉 Production macOS Installer Created Successfully! 🎉   ${NC}"
echo -e "${GREEN}================================================================${NC}"
echo -e "📦 App Bundle:  ${CYAN}${APP_BUNDLE}${NC}"
echo -e "💿 DMG Package: ${CYAN}${DMG_OUTPUT}${NC}"
echo -e "🔑 Checksum:    ${CYAN}${DMG_OUTPUT}.sha256${NC}"
echo -e "📊 Size:        ${CYAN}$(du -h "$DMG_OUTPUT" | cut -f1)${NC}"
