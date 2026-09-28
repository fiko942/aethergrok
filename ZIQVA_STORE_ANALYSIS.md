# Engineering Architecture & Context Analysis — AetherGrok Desktop Studio

## Overview
- **Project**: AetherGrok Desktop GUI Studio & Grok Build VS Code / Desktop Host
- **Stack**: Go 1.24, Wails v2.8+, Svelte 5 (Runes), Tailwind CSS, Ant Design Dark Design System.
- **Protocol**: Agent Client Protocol (ACP) JSON-RPC over `grok agent stdio` and Wails Go IPC bridge.

---

## Recent Engineering Patches (Session Log)

### 1. In-App Auto-Updater Checksum Match Fix (v1.0.5)
- **Component**: `frontend/src/lib/stores/updater.svelte.ts`
- **Issue**: Auto-update checksum mismatch error on Apple Silicon (macOS ARM64) due to naive `.find()` selecting the first alphabetical `.sha256` asset (AMD64) instead of matching the downloaded target file.
- **Fix**:
  - Implemented 3-tier hierarchical resolution: exact target filename match (`<filename>.sha256`), OS format & architecture substring matching (`arm64` vs `amd64`, `.dmg`/`.exe`/`.zip`), and fallback.
  - Bumped release to `v1.0.5` and published all 14 multi-platform assets.

### 2. UI Refinement & Polish
- **Component**: `frontend/src/lib/components/layout/SettingsModal.svelte`
- **Change**: 
  - Restyled and consolidated the **Grok Config** card and button into a clean single line with `whitespace-nowrap shrink-0`.
  - Replaced harsh light/white borders with subtle Ant Design Dark tokens (`border-ant-border-secondary/80 hover:border-ant-primary/40`).
  - Shortened descriptive text and title to compact format (`~/.grok/config.toml`).

### 2. Dynamic Shortcut System & Key Synchronization
- **Components**:
  - `frontend/src/lib/components/layout/SettingsModal.svelte`
  - `frontend/src/lib/components/ui/KeyRecorderModal.svelte`
  - `frontend/src/App.svelte`
  - `frontend/src/app.d.ts`
- **Change**:
  - Converted `keyboardShortcuts` table in the Shortcuts tab to a reactive Svelte 5 `$derived` state that binds to `settingsStore.snapshotShortcut`.
  - Added `formatShortcutKeys` to format standalone modifier keys (e.g., `ShiftRight` -> `Right Shift`).
  - Updated `matchesShortcut` in `App.svelte` with `normalizeKeyName` to ensure `Right Shift` triggers snapshot without intercepting regular text input.

### 3. Type Safety & Compilation Guarantees
- Resolved `PlanGateMode` type import in `SettingsModal.svelte`.
- Added `RevealGrokConfigFile` to `window.go.main.App` in `frontend/src/app.d.ts`.
- Added `handleExternalFiles` to `composerRef` signature in `App.svelte`.
- Verified 0 errors across `svelte-check`, `vite build`, `go test ./test/...`, and Go binary compilation.

---

## Architectural Invariants
1. **10-Turn DOM Windowing**: Virtualized conversation history retaining latest active nodes in DOM.
2. **Non-Intrusive OS-Excluding Screen Capture**: 4-phase coordination (hide window -> compositor delay -> native grab -> focus restoration).
3. **ACP Protocol Compliance**: JSON-RPC over stdio with client-side plan gate enforcement for terminal/fs safety.
