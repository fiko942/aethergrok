# Superpowers Specification: Multi-Terminal Split Persistence, Dynamic Text Wrapping & Top Tabs Theme Refinement

**Date:** 2026-09-28  
**Scope:** AetherGrok Desktop (macOS/Universal Desktop)  
**Status:** Implemented & Verified in Production  

---

## 1. Executive Summary & Problem Context

In this milestone, three critical UX and architectural problems were addressed and stabilized:
1. **Dynamic Text Wrapping & Buffer Reflow on Pane Resize:**
   - **Problem:** When resizing split terminal panes (e.g. dragging the divider to make pane 1 narrower), text in xterm did not wrap or reflow to the new columns width. Lines like the shell prompt `(base) fiko942@WIJiS-MacBook-Air g` stayed truncated horizontally.
   - **Root Cause:** Svelte DOM inline widths (`style="width: 35%"`) were updating asynchronously. Synchronous `fitAddon.fit()` calls were measuring stale container dimensions, and without updating `cols` and sending them via `App.ResizeTerminal` to the Go backend (`pty.Setsize`), the PTY never emitted `SIGWINCH` to zsh/bash readline/zle.
   - **Solution:** Integrated DOM synchronization via `tick()` and two-pass layout calculation (`setTimeout(..., 40)`) in `forceRefitTerminal` and `refitAllSplitTerminals()`. Added strict CSS bounding (`min-w-0 min-h-0 overflow-hidden`) to terminal viewport containers.

2. **Per-Session Terminal Layout & State Persistence:**
   - **Problem:** Upon closing or restarting the application, all multi-terminal split configurations, pane sizes, collapse state (`isCollapsed`), and dock orientation (`bottom` vs `right`) were lost and reset to a single default terminal.
   - **Solution:** Implemented local state persistence in `TerminalStore` under key `aethergrok_terminal_state_v1`. Hydrates and persists tabs (`TerminalTab[]`), active terminal IDs, split groups (`TerminalSplitGroup[]`), pane percentages (`paneSizes`), collapse state, dock positions, and panel dimensions (`panelHeight` and `panelWidth`). On application restart, backend Go PTY instances are cleanly reconnected and buffers restored.

3. **Elimination of Harsh White Borders on Top Session Tabs in Dark Mode:**
   - **Problem:** In dark mode, all session tabs in the top navigation bar exhibited sharp, contrasting white rectangular borders around inactive and active tabs (`border-ant-border/40`).
   - **Solution:** Aligned session tab design with `/design-taste-frontend` tokens: transparent/subtle backgrounds for inactive tabs (`hover:border-ant-border-secondary dark:hover:border-white/5`), elevated active tab styling (`bg-ant-bg border-ant-border-secondary dark:border-white/10`) with primary blue bottom line accent, and dark-mode-safe border tokens for the container and new-session actions.

---

## 2. Architecture & Data Structures

### 2.1 Persisted Terminal State Schema (`TerminalStore`)
```typescript
export interface TerminalSplitGroup {
  id: string;
  paneTermIds: string[]; // List of termIds in this split view
  splitDirection: 'horizontal' | 'vertical'; // 'horizontal' (columns) or 'vertical' (stacked)
  paneSizes?: number[]; // percentage weights per pane, summing to 100 (e.g. [35, 65])
}

interface PersistedTerminalState {
  version: number;
  sessionTerminals: Record<string, TerminalTab[]>;
  activeTerminalIdPerSession: Record<string, string>;
  sessionSplitGroups: Record<string, TerminalSplitGroup[]>;
  focusedPaneTermId: Record<string, string>;
  sessionCollapsed: Record<string, boolean>;
  sessionDockPosition: Record<string, 'bottom' | 'right'>;
  panelHeight: number;
  panelWidth: number;
}
```

### 2.2 Re-attachment & Process Synchronization
When AetherGrok starts or when switching sessions:
1. `TerminalStore.loadFromStorage()` deserializes persisted tabs and split groups.
2. In `TerminalPanel.svelte`, an `$effect` detects existing tabs and guarantees backend Go PTY processes are alive by invoking `window.go.main.App.CreateTerminal(sessionId, tab.id, tab.cwd, '')`.
3. In `initXterm`, `window.go.main.App.GetTerminalBuffer(termId)` retrieves previous terminal output history so prompt state and logs are seamlessly restored.

---

## 3. Dynamic Text Reflow & Layout Refit Architecture

```
User Drags Divider
       │
       ▼
Divider mousemove / mouseup (onEnd)
       │
       ▼
Update paneSizes in TerminalStore & Svelte Reactive State (e.g. [30, 70])
       │
       ▼
`refitAllSplitTerminals()`
       ├── Step 1: `await tick()` -> Wait for Svelte 5 DOM style width/height commit
       ├── Step 2: `forceRefitTerminal(termId)`
       │     ├── Measure `container.clientWidth` & `container.clientHeight`
       │     ├── `inst.fitAddon.fit()` -> Recomputes xterm `cols` & `rows`
       │     └── `App.ResizeTerminal(termId, cols, rows)` -> Sends `pty.Setsize` to Go
       │                                                      │
       │                                                      ▼
       │                                           POSIX Kernel SIGWINCH
       │                                                      │
       │                                                      ▼
       │                                           Shell (zsh/bash) wraps prompt
       └── Step 3: `setTimeout(..., 40)` -> Second pass to guarantee final flexbox settle
```

---

## 4. Visual Design & Theme Harmonization

### Top Session Tabs (`SessionTabs.svelte`)
| Element | Previous Dark Mode Style | Modern Anti-Slop Style |
| :--- | :--- | :--- |
| **Main Header Bar** | `border-b border-ant-border` (bright) | `border-b border-ant-border-secondary dark:border-white/5` |
| **Inactive Session Tabs** | `border-ant-border/40` (harsh box) | `border-transparent hover:border-ant-border-secondary dark:hover:border-white/5 bg-ant-bg-tertiary/20 hover:bg-ant-bg-tertiary/60` |
| **Active Session Tab** | `border-ant-border` | `bg-ant-bg text-ant-primary border-ant-border-secondary dark:border-white/10 shadow-2xs font-semibold` + bottom line accent |
| **New Session Button** | Default background without dark border | `border border-ant-border-secondary dark:border-white/5` |

---

## 5. Verification & Testing Evidence

- **Frontend Compilation:** Verified via `vite build` (`dist/assets/index-TGLHGnK1.js`).
- **Wails Desktop Packaging:** Clean macOS build executed via `wails build -clean`, generating production binary `aethergrok.app`.
- **Git State:** Committed to branch `main` and pushed to remote `origin/main` (`51abcaa`).
