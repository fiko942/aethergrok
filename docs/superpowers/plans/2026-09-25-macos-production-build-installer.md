# Production macOS DMG and App Bundle Build Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to execute this plan task-by-task.

**Goal:** Create a production macOS build script `build-macos.sh` that compiles the frontend, builds the native macOS `.app` bundle via Wails v2 (optimized for Apple Silicon `darwin/arm64` and Intel `darwin/amd64` or universal), packages it into a distributable `.dmg` installer with an Applications symlink and custom volume styling using native macOS `hdiutil`, and produces a release artifact in `build/bin/`.

**Architecture:**
- `build-macos.sh`: Shell script supporting `./build-macos.sh [arm64|amd64|universal|dmg]`.
- Steps in `build-macos.sh`:
  1. Environment validation: checks `go`, `pnpm`/`npm`, `wails`, and `hdiutil`.
  2. Frontend build: builds optimized Vite bundle with Anthropic serif fonts and dark theme tokens.
  3. Wails macOS build: executes `wails build -platform darwin/<arch> -clean -nsis=false` with `-s -w` flags for stripped lightweight binary.
  4. DMG packaging: creates an `.app` installer DMG using native `hdiutil` with drag-and-drop link to `/Applications`.
  5. Checksum generation: creates `SHA256SUMS.txt` for installer verification.

---

### Task 1: Create `build-macos.sh` Production Build Script
- Produces executable `build-macos.sh` in workspace root.
- Handles Apple Silicon (`arm64`), Intel (`amd64`), and Universal binaries.
- Packages `.app` into `.dmg` installer using standard `hdiutil makehybrid` or `hdiutil create`.

### Task 2: Verify `build-macos.sh` Execution
- Run `bash build-macos.sh arm64` to verify end-to-end compilation and installer creation.
- Verify generated `.app` bundle and `.dmg` installer under `build/bin/`.

### Task 3: Commit and Push to GitHub
- Stage `build-macos.sh` and updated documentation.
- Commit and push to `origin/main`.
