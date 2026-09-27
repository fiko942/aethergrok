# Superpowers Implementation Plan: Terminal Split State Persistence, Dynamic Reflow & Visual Cleanup

**Date:** 2026-09-28  
**Component:** TerminalPanel.svelte, terminal.svelte.ts, SessionTabs.svelte  
**Goal:** Deliver complete persistence across app restarts, responsive xterm buffer wrapping, and polished dark mode top session tabs.

---

## 1. Problem Statements & Identified Causes

1. **Terminal Wrapping Bug on Split Resize:**
   - Text prompts remained horizontally cut off when dragging the divider to make split pane 1 smaller.
   - Root cause: DOM flex styling had not finished applying when `fit()` was called, and missing `SIGWINCH` propagation prevented shell readline/zle from re-rendering the wrapped line.
2. **Terminal Session Memory Loss on App Restart:**
   - Tabs, multi-split configurations, divider proportions, dock positions (`bottom` vs `right`), and collapse states were stored in memory only.
   - Restarting the app erased all split state and recreated a single tab.
3. **Contrasting White Border Slop on Top Tabs:**
   - In dark mode, all session tabs in the top header had visible white rectangular borders around each inactive tab.

---

## 2. Implemented Architecture & Solutions

### A. Terminal Reflow & Dynamic Wrapping
- Added `forceRefitTerminal(termId: string)` in `TerminalPanel.svelte`.
- Connected `refitAllSplitTerminals()` to Svelte 5 `tick()` + two-pass layout settlement.
- Applied `min-w-0 min-h-0 overflow-hidden` to both horizontal and vertical split container viewports.
- Validated that terminal `cols` and `rows` are dispatched to `App.ResizeTerminal` / `pty.Setsize` on resize end.

### B. Persistent Terminal State Machine
- Created `PersistedTerminalState` in `terminal.svelte.ts` backed by `localStorage` (`aethergrok_terminal_state_v1`).
- Saved state on every mutation:
  - `createTerminal`, `splitNewTerminal`, `splitTerminal`, `unsplitTerminal`
  - `setGroupPaneSizes`, `setPanelHeight`, `setPanelWidth`
  - `switchTerminal`, `renameTerminal`, `closeTerminal`, `closeAllForSession`
  - `toggleSessionCollapse`, `setDockPosition`, `toggleDockPosition`
- Hydrated on app startup and auto-reinstantiated backend Go PTY instances via `App.CreateTerminal`.

### C. Top Tabs Design Harmonization
- Updated `SessionTabs.svelte` to remove `border-ant-border/40` on inactive tabs.
- Elevated active tabs with `border-ant-border-secondary dark:border-white/10` and primary accent line.
- Converted container divider to `border-b border-ant-border-secondary dark:border-white/5`.

---

## 3. Verification & Git Log
- Clean production build: `wails build -clean`.
- Pushed commits:
  - `26eeb24`: `fix(terminal): ensure split terminal text wraps and refits dynamically on pane resize`
  - `b9ca771`: `fix(theme): eliminate harsh white borders on top session tabs in dark mode`
  - `51abcaa`: `feat(terminal): persist split layout, pane sizes, dock position and collapse state across restarts`
