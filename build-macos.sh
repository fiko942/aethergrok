#!/usr/bin/env bash

# ==============================================================================
# AetherGrok Desktop GUI - Production macOS Builder (.app, .dmg, and portable)
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

# Read version dynamically from wails.json if available
VERSION="1.0.0"
if [ -f "$PROJECT_ROOT/wails.json" ]; then
  VERSION=$(grep -o '"version": *"[^"]*"' "$PROJECT_ROOT/wails.json" | cut -d '"' -f 4 || echo "1.0.0")
fi

# Target architecture: arm64 (default for Apple Silicon), amd64 (Intel), or universal
TARGET_ARCH="${1:-arm64}"

export PATH="/opt/homebrew/bin:/usr/local/bin:$(go env GOPATH 2>/dev/null || echo "$HOME/go")/bin:$PATH"

echo -e "${CYAN}================================================================${NC}"
echo -e "${CYAN}${BOLD}   🚀 AetherGrok Production Builder for macOS (${TARGET_ARCH}) v${VERSION} 🚀   ${NC}"
echo -e "${CYAN}================================================================${NC}"

# 1. Check prerequisites
echo -e "${YELLOW}Step 1/6: Verifying build dependencies...${NC}"

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
echo -e "\n${YELLOW}Step 2/6: Compiling Vite Frontend Production Bundle...${NC}"
cd "$FRONTEND_DIR"

if command -v pnpm &> /dev/null; then
  pnpm run build
else
  npm run build
fi

echo -e "${GREEN}✓ Frontend compiled successfully.${NC}"

# 3. Build macOS .app via Wails
echo -e "\n${YELLOW}Step 3/6: Compiling macOS Native Application with Wails...${NC}"
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

# Ensure macOS Privacy Usage Descriptions are present in the compiled Info.plist
APP_PLIST="$BUILD_DIR/aethergrok.app/Contents/Info.plist"
if [ -f "$APP_PLIST" ]; then
  echo "Injecting privacy usage descriptions into Info.plist..."
  /usr/libexec/PlistBuddy -c "Add :NSMicrophoneUsageDescription string 'AetherGrok requires microphone access for voice dictation and speech-to-text input.'" "$APP_PLIST" 2>/dev/null || \
  /usr/libexec/PlistBuddy -c "Set :NSMicrophoneUsageDescription 'AetherGrok requires microphone access for voice dictation and speech-to-text input.'" "$APP_PLIST" 2>/dev/null || true
  
  /usr/libexec/PlistBuddy -c "Add :NSSpeechRecognitionUsageDescription string 'AetherGrok uses speech recognition to convert dictated voice prompts into text.'" "$APP_PLIST" 2>/dev/null || \
  /usr/libexec/PlistBuddy -c "Set :NSSpeechRecognitionUsageDescription 'AetherGrok uses speech recognition to convert dictated voice prompts into text.'" "$APP_PLIST" 2>/dev/null || true
fi

# Locate compiled application bundle
APP_BUNDLE="$BUILD_DIR/aethergrok.app"
if [ ! -d "$APP_BUNDLE" ]; then
  if [ -d "$BUILD_DIR/AetherGrok.app" ]; then
    APP_BUNDLE="$BUILD_DIR/AetherGrok.app"
  else
    echo -e "${RED}Error: Application bundle not found in $BUILD_DIR${NC}"
    exit 1
  fi
fi

# 4. Prepare and package .dmg Installer
echo -e "\n${YELLOW}Step 4/6: Creating macOS .dmg Installer...${NC}"

DMG_STAGE_DIR="$BUILD_DIR/dmg_stage_${TARGET_ARCH}"
rm -rf "$DMG_STAGE_DIR"
mkdir -p "$DMG_STAGE_DIR"

DMG_OUTPUT="$BUILD_DIR/AetherGrok-${VERSION}-macOS-${TARGET_ARCH}.dmg"
rm -f "$DMG_OUTPUT"

echo "Staging application bundle..."
cp -R "$APP_BUNDLE" "$DMG_STAGE_DIR/AetherGrok.app"

echo "Creating compressed DMG installer with hdiutil..."
for attempt in 1 2 3; do
  if hdiutil create \
    -volname "$APP_NAME" \
    -srcfolder "$DMG_STAGE_DIR" \
    -ov \
    -format UDZO \
    -noanyowners \
    "$DMG_OUTPUT"; then
    break
  fi
  echo -e "${YELLOW}hdiutil create attempt $attempt failed, retrying in 2 seconds...${NC}"
  sleep 2
done

rm -rf "$DMG_STAGE_DIR"
if [ ! -f "$DMG_OUTPUT" ]; then
  echo -e "${RED}Error: Failed to create DMG installer.${NC}"
  exit 1
fi
echo -e "${GREEN}✓ DMG Installer package ready: $(basename "$DMG_OUTPUT")${NC}"

# 5. Create Portable Portable Archive (.tar.gz)
echo -e "\n${YELLOW}Step 5/6: Creating Portable Archive (.tar.gz)...${NC}"
PORTABLE_OUTPUT="$BUILD_DIR/AetherGrok-${VERSION}-macOS-${TARGET_ARCH}-portable.tar.gz"
rm -f "$PORTABLE_OUTPUT"

tar -czf "$PORTABLE_OUTPUT" -C "$BUILD_DIR" "$(basename "$APP_BUNDLE")"
echo -e "${GREEN}✓ Portable archive ready: $(basename "$PORTABLE_OUTPUT")${NC}"

# 6. Generate Checksums
echo -e "\n${YELLOW}Step 6/6: Generating SHA256 Checksums...${NC}"
cd "$BUILD_DIR"
shasum -a 256 "$(basename "$DMG_OUTPUT")" > "$(basename "$DMG_OUTPUT").sha256"
shasum -a 256 "$(basename "$PORTABLE_OUTPUT")" > "$(basename "$PORTABLE_OUTPUT").sha256"

echo -e "\n${GREEN}================================================================${NC}"
echo -e "${GREEN}${BOLD}   🎉 Production macOS Packages Created Successfully! 🎉   ${NC}"
echo -e "${GREEN}================================================================${NC}"
echo -e "📦 App Bundle:  ${CYAN}${APP_BUNDLE}${NC}"
echo -e "💿 DMG Package: ${CYAN}${DMG_OUTPUT} ($(du -h "$DMG_OUTPUT" | cut -f1))${NC}"
echo -e "🗜️  Portable:    ${CYAN}${PORTABLE_OUTPUT} ($(du -h "$PORTABLE_OUTPUT" | cut -f1))${NC}"
echo -e "🔑 Checksums:    ${CYAN}$(basename "$DMG_OUTPUT").sha256 & $(basename "$PORTABLE_OUTPUT").sha256${NC}"
