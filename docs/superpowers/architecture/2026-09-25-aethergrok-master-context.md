# AetherGrok Master Context & Architecture State (2026-09-25)

## Overview & Vision
AetherGrok is an ultra-performant, native desktop GUI companion for Grok Build CLI, developed by Wiji Fiko Teren. The application combines an anthracite/zinc dark design system, typography using the Anthropic serif font family, and high responsiveness.

---

## 1. Right Sidebar: Workspace Files Explorer & Git Inspector
### Architectural Requirements
- **Session State Isolation**: `rightSidebarOpen` and `rightSidebarTab` are stored directly on the `Session` entity in `frontend/src/lib/stores/session.svelte.ts`. Switching workspace tabs preserves each tab's individual sidebar state.
- **Collapsible & Resizable**: Clamped width between 260px and 600px with a default of 340px. Supports double-click handle reset to 340px, and snap-to-close under 180px.
- **Header & Shortcut Integration**: Toggle button `PanelRightOpen` / `PanelRightClose` in the top navigation bar, plus keyboard shortcut `⌘⌥B` (macOS) / `Ctrl+Alt+B` (Windows/Linux).
- **Tab 1: File & Folder Explorer**:
  - Recursive directory tree with level indentation.
  - Asynchronous lazy loading: children of folders are fetched on demand via Go backend (`pkg/workspace`), preventing UI freezes on directories like `node_modules` or `.git`.
  - Clear loading spinner during asynchronous expansion.
  - Tailored icons for TypeScript, Svelte, Go, JSON, Shell, Markdown, and images.
  - Debounced search filter at the top (200ms debounce).
  - Read-only file viewer (`FileViewerModal.svelte`) with syntax highlighting and 1MB size safety cap.
  - Dual-mode Markdown viewer: Raw code view vs. Rendered preview styled with Anthropic serif font.
- **Tab 2: Git Inspector & 5s Auto-Refresh**:
  - Active branch name badge with Clean/Dirty status indicator.
  - Uncommitted changes list with status badges (`M`, `A`, `D`, `U`) and addition/deletion metrics.
  - 5-second automatic silent background polling to update uncommitted status without flickering or UI freeze. Automatically pauses during active git actions.
  - Quick commit input with **Commit**, **Commit & Push**, and **Pull** actions.
  - Visual interactive diff modal (`DiffViewModal.svelte`) displaying unified diffs with green additions (`+`) and red deletions (`-`).

---

## 2. Left Sidebar: Workspace & Session Explorer
- **Collapsible & Resizable**: Left sidebar width clamped between 220px and 480px with a default of 288px. Double-click reset to 288px.
- **Fluid Child Expansion**: Element `WorkspaceSidebar.svelte` uses `w-full h-full` and `flex-1 min-w-0` to eliminate horizontal clipping and empty gaps during resize.
- **Workspace Folder Pagination**: Default display of 8 sessions per workspace folder with "+8 Show More" progressive loading and "Show Less" reset.

---

## 3. Production Build Pipeline & macOS DMG Packaging
- **Builder Script (`build-macos.sh`)**:
  - Automatically compiles Vite frontend production bundle.
  - Uses Wails v2 with stripped ldflags (`-s -w`) targeting `darwin/arm64`, `darwin/amd64`, or `darwin/universal`.
  - Packages native `.app` bundle into a compressed drag-and-drop `.dmg` installer with symlink to `/Applications`.
  - Generates SHA256 checksums automatically for release verification.

---

## 4. Key Bug Fixes & Guardrails
- **CurrentSession Reactive Rune**: Explicitly declared as `const currentSession = $derived(sessionStore.activeSession);` and `const currentSessionWorkspace = $derived.by(...)` at the top level of `App.svelte` to prevent blank screen runtime crashes and ensure accurate workspace path resolution.
- **Sidebar Width Stretch Fix**: Replaced rigid `w-64` in `WorkspaceSidebar.svelte` with dynamic `w-full` container layout.
