# Superpowers Architecture & Implementation Spec: OS-Wide Global Hotkey & Cross-Platform Git Utility

## Date: 2026-09-27

## Overview
This specification covers two critical desktop capabilities implemented in AetherGrok Studio:
1. **OS-Wide Global Hotkey Listener (`pkg/hotkey`)**: Allows global screenshot/snapshot triggering (such as `RightShift` or `CmdOrCtrl+Shift+S`) across macOS, Windows, and Linux even when AetherGrok is running in the background or unfocused.
2. **Robust Multi-Arch Git Binary Resolution (`pkg/gitutil`)**: Eliminates the macOS xcrun dynamic linking wrapper crash (`unable to load libxcrun ... fat file, but missing compatible architecture`) by resolving direct native git binaries (Homebrew, CommandLineTools, Xcode).
3. **Session Tabs Workspace Badge & Multi-Workspace Isolation**: Enhanced tab rendering with workspace context badges, active workspace filtering, and robust drag-and-drop tab reordering.
4. **Enhanced Diff Calculation & File Drag-and-Drop Ingestion**: Standardized unified diff stat parsing for file cards and expanded macOS proxy icon / web blob drop handling in Composer.

---

## 1. Global Hotkey Engine (`pkg/hotkey`)
- **macOS (`hotkey_darwin.go`)**:
  - Implements an asynchronous low-level `CGEventTap` (`kCGHeadInsertEventTap`, `kCGEventTapOptionListenOnly`).
  - Listens for `kCGEventKeyDown` and `kCGEventFlagsChanged`.
  - Supports standalone modifier keys (e.g. `RightShift` (kc 60), `LeftShift` (kc 56), `RightCmd` (kc 54), `RightOption` (kc 61)) and key combinations (`CmdOrCtrl+Shift+S`).
  - Includes a 250ms hardware debounce timer to avoid multi-event triggering on single key strokes.
  - Automatically dispatches Wails event `snapshot:trigger_global` to frontend runtime.
- **Windows (`hotkey_windows.go`) & Linux/Fallback (`hotkey_other.go`)**:
  - Platform interface abstraction (`platformManager`) for zero-cost compilation across all target OS architectures.

---

## 2. macOS Git Architecture Shim Resolution (`pkg/gitutil`)
- **Problem**: `/usr/bin/git` on macOS is an `xcrun` dispatch shim. Under varied Rosetta 2 / arm64 launch scenarios, `xcrun` dynamically links to `libxcrun.dylib`, failing with architecture mismatch errors.
- **Solution**:
  - `gitutil.Executable()` searches direct native binaries before falling back to PATH:
    1. `/opt/homebrew/bin/git` (Apple Silicon Homebrew)
    2. `/usr/local/bin/git` (Intel Homebrew)
    3. `/Library/Developer/CommandLineTools/usr/bin/git`
    4. `/Applications/Xcode.app/Contents/Developer/usr/bin/git`
    5. Fallback: `"git"`
  - Integrated across `pkg/workspace`, `pkg/skills/importer`, and `app.go` (`RevertWorkspaceFiles`).

---

## 3. Session Tabs & Workspace Isolation
- **Tabs Rendering**:
  - Updated `SessionTabs.svelte` to iterate over `sessionStore.openTabs`.
  - Displays concise workspace pill badges (`ws.name`) on tabs when multiple workspaces are open, ensuring clear context.
  - Double-click inline renaming and workspace tooltips.
- **Drag and Drop Reordering**:
  - `reorderSessions(fromIndex, toIndex)` now reorders `openTabSessionIds` and persists state to `localStorage` (`aethergrok_settings_v1`).

---

## 4. Verification & Testing
- `go test ./test/... ./pkg/... -v`: All unit test suites pass (100%).
- `pnpm test test/diffUtils.test.ts`: 5/5 unit tests pass.
- `pnpm --prefix frontend run check`: 0 errors.
- `pnpm --prefix frontend run build`: Clean production asset compilation.
