#!/usr/bin/env bash

# ==============================================================================
# AetherGrok Desktop GUI - Universal Developer Launcher
# ==============================================================================

set -e

# Export standard system and homebrew paths
export PATH="/opt/homebrew/bin:/usr/local/bin:$(go env GOPATH 2>/dev/null || echo "$HOME/go")/bin:$PATH"

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FRONTEND_DIR="$PROJECT_ROOT/frontend"

# Colors for terminal output
CYAN='\033[0;36m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${CYAN}======================================================${NC}"
echo -e "${CYAN}   ✨ AetherGrok Desktop GUI Studio - Launcher ✨   ${NC}"
echo -e "${CYAN}======================================================${NC}"

# Check prerequisites
if ! command -v pnpm &> /dev/null; then
  echo -e "${RED}Error: 'pnpm' is not installed or not found in PATH.${NC}"
  echo "Please install pnpm or ensure /opt/homebrew/bin is in your PATH."
  exit 1
fi

if ! command -v go &> /dev/null; then
  echo -e "${RED}Error: 'go' is not installed or not found in PATH.${NC}"
  echo "Please install Go or ensure /opt/homebrew/bin is in your PATH."
  exit 1
fi

# Determine mode
MODE="${1:-wails}"

case "$MODE" in
  "ui"|"web"|"browser")
    echo -e "${GREEN}Starting Svelte 5 Frontend in Web Browser Mode...${NC}"
    cd "$FRONTEND_DIR"
    if [ ! -d "node_modules" ]; then
      echo -e "${YELLOW}Installing frontend dependencies...${NC}"
      pnpm install
    fi
    pnpm run dev
    ;;

  "test")
    echo -e "${GREEN}Running Go unit tests and Frontend typechecks...${NC}"
    cd "$PROJECT_ROOT"
    echo -e "${YELLOW}1. Running Go unit tests:${NC}"
    go test ./test/... -v
    echo -e "${YELLOW}2. Running Svelte 5 typecheck:${NC}"
    cd "$FRONTEND_DIR"
    pnpm run check
    echo -e "${GREEN}✓ All tests and checks passed successfully!${NC}"
    ;;

  "build")
    echo -e "${GREEN}Building standalone executable binary...${NC}"
    cd "$FRONTEND_DIR"
    echo -e "${YELLOW}1. Building frontend production bundle...${NC}"
    pnpm run build
    cd "$PROJECT_ROOT"
    echo -e "${YELLOW}2. Compiling Go desktop binary...${NC}"
    mkdir -p build/bin
    go build -o build/bin/aethergrok main.go app.go
    echo -e "${GREEN}✓ Standalone binary created at: build/bin/aethergrok${NC}"
    ;;

  "wails"|"desktop"|*)
    if ! command -v wails &> /dev/null; then
      echo -e "${YELLOW}'wails' CLI not found. Installing Wails v2 latest...${NC}"
      go install github.com/wailsapp/wails/v2/cmd/wails@latest
    fi

    # Ensure frontend deps are installed
    if [ ! -d "$FRONTEND_DIR/node_modules" ]; then
      echo -e "${YELLOW}Installing frontend dependencies...${NC}"
      cd "$FRONTEND_DIR" && pnpm install && cd "$PROJECT_ROOT"
    fi

    echo -e "${GREEN}Starting Wails Native Desktop Live-Development...${NC}"
    cd "$PROJECT_ROOT"
    wails dev
    ;;
esac
