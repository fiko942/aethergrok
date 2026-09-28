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

# Resign bundle with explicit Designated Requirement (avoids Gatekeeper broken signature after Plist injection)
echo -e "\n${YELLOW}Signing macOS Application bundle with Designated Requirement...${NC}"
codesign --force --deep --sign - --requirements '= designated => identifier "com.fiko942.aethergrok"' "$APP_BUNDLE" || true
codesign -vvv "$APP_BUNDLE" || true

# 4. Prepare and package .dmg Installer with Custom Styling & Layout
echo -e "\n${YELLOW}Step 4/5: Creating Styled macOS .dmg Installer...${NC}"

DMG_TMP="$BUILD_DIR/AetherGrok_tmp_${TARGET_ARCH}.dmg"
DMG_OUTPUT="$BUILD_DIR/AetherGrok-${VERSION}-macOS-${TARGET_ARCH}.dmg"
rm -f "$DMG_TMP" "$DMG_OUTPUT"

DMG_STAGE_DIR="$BUILD_DIR/dmg_stage_${TARGET_ARCH}"
rm -rf "$DMG_STAGE_DIR"
mkdir -p "$DMG_STAGE_DIR"

echo "Staging application bundle and background image..."
cp -R "$APP_BUNDLE" "$DMG_STAGE_DIR/AetherGrok.app"
ln -s /Applications "$DMG_STAGE_DIR/Applications"

# Add custom background folder
mkdir -p "$DMG_STAGE_DIR/.background"
if [ -f "$PROJECT_ROOT/resources/dmg-background.tiff" ]; then
  cp "$PROJECT_ROOT/resources/dmg-background.tiff" "$DMG_STAGE_DIR/.background/background.tiff"
fi
if [ -f "$PROJECT_ROOT/resources/dmg-background.png" ]; then
  cp "$PROJECT_ROOT/resources/dmg-background.png" "$DMG_STAGE_DIR/.background/background.png"
fi

# Create a writable temporary disk image
echo "Creating writable disk image..."
hdiutil create \
  -volname "$APP_NAME" \
  -srcfolder "$DMG_STAGE_DIR" \
  -ov \
  -fs HFS+ \
  -format UDRW \
  "$DMG_TMP"

rm -rf "$DMG_STAGE_DIR"

# Mount the temporary image to apply Finder view layout & icon coordinates
echo "Mounting disk image to configure Finder layout..."
ATTACH_OUTPUT=$(hdiutil attach "$DMG_TMP" -noverify -noautoopen)
DEV_NODE=$(echo "$ATTACH_OUTPUT" | grep -oE '/dev/disk[0-9]+' | head -n 1)
MOUNT_POINT=$(echo "$ATTACH_OUTPUT" | grep -oE '/Volumes/[^ ]+' | tail -n 1)
VOL_NAME=$(basename "$MOUNT_POINT")

echo "Mounted on $MOUNT_POINT ($DEV_NODE), Volume: $VOL_NAME"

# AppleScript to configure Finder presentation (Window size 660x420, icon 120, left: 170, right: 490)
echo "Applying custom Finder view options, bounds, and icon positions..."
osascript -e "
tell application \"Finder\"
  set theDisk to disk \"$VOL_NAME\"
  open theDisk
  set theWindow to container window of theDisk
  set current view of theWindow to icon view
  set toolbar visible of theWindow to false
  set statusbar visible of theWindow to false
  set the bounds of theWindow to {300, 100, 960, 520}
  set opts to the icon view options of theWindow
  set arrangement of opts to not arranged
  set icon size of opts to 120
  set label position of opts to bottom
  set text size of opts to 12
  if exists file \".background:background.tiff\" of theDisk then
    set background picture of opts to file \".background:background.tiff\" of theDisk
  else if exists file \".background:background.png\" of theDisk then
    set background picture of opts to file \".background:background.png\" of theDisk
  end if
  set position of item \"AetherGrok.app\" of theDisk to {170, 215}
  set position of item \"Applications\" of theDisk to {490, 215}
  update theDisk without registering applications
  delay 1
  close theWindow
end tell
" || true

# Sync disk and detach cleanly
sync
sleep 1
hdiutil detach "$DEV_NODE" -force || true
sleep 1

# Convert temporary read-write image to compressed read-only production DMG (UDZO)
echo "Converting to compressed read-only DMG installer..."
hdiutil convert "$DMG_TMP" -format UDZO -imagekey zlib-level=9 -o "$DMG_OUTPUT" -ov
rm -f "$DMG_TMP"

if [ ! -f "$DMG_OUTPUT" ]; then
  echo -e "${RED}Error: Failed to create DMG installer.${NC}"
  exit 1
fi
echo -e "${GREEN}✓ Styled DMG Installer package ready: $(basename "$DMG_OUTPUT")${NC}"

# 5. Generate Checksums
echo -e "\n${YELLOW}Step 5/5: Generating SHA256 Checksums...${NC}"
cd "$BUILD_DIR"
shasum -a 256 "$(basename "$DMG_OUTPUT")" > "$(basename "$DMG_OUTPUT").sha256"

echo -e "\n${GREEN}================================================================${NC}"
echo -e "${GREEN}${BOLD}   🎉 Production macOS Packages Created Successfully! 🎉   ${NC}"
echo -e "${GREEN}================================================================${NC}"
echo -e "📦 App Bundle:  ${CYAN}${APP_BUNDLE}${NC}"
echo -e "💿 DMG Package: ${CYAN}${DMG_OUTPUT} ($(du -h "$DMG_OUTPUT" | cut -f1))${NC}"
echo -e "🔑 Checksums:    ${CYAN}$(basename "$DMG_OUTPUT").sha256${NC}"
