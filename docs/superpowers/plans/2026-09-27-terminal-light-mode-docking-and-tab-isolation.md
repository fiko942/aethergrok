# Terminal Light Mode, Flexible Docking (Bottom/Right), Per-Session Collapse & Tab Renaming Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement full light mode support for xterm.js terminals, flexible bottom/right layout docking, per-session collapse memory bar, inline terminal tab renaming on double-click, and strict session-isolated terminal collapse state.

**Architecture:** 
1. Refactor `TerminalStore` in Svelte 5 runes to track UI layout state per session: `sessionLayout` (bottom vs right), `sessionCollapsed` (boolean per session), and terminal tab titles with custom renaming.
2. Update terminal color scheme dynamically based on `settingsStore.theme`: supply high-contrast light colors for `light-antd` and deep dark tones for `dark-studio`/`dark-high-contrast`.
3. In `TerminalPanel.svelte`, support both `bottom` and `right` docking positions with responsive drag-resizing, an inline editable tab title on double click, and a sleek mini-collapsed bar when collapsed that retains quick-expand toggles.

**Tech Stack:** Svelte 5 runes (`$state`, `$derived`, `$effect`), `@xterm/xterm`, `@xterm/addon-fit`, Tailwind CSS, Ant Design token system, Lucide icons.

---

## Global Constraints

- Never use contrastive negation in UI strings or logs.
- Preserve Svelte 5 rune reactivity conventions across stores and components.
- Terminal process execution must remain strictly isolated per session and process group.
- Collapsing or opening the terminal in Tab A must leave Tab B's terminal state intact.

---

## Detailed Task Breakdown

### Task 1: Update TerminalStore to support Session-Isolated Collapse, Docking Position & Tab Renaming

**Files:**
- Modify: `frontend/src/lib/stores/terminal.svelte.ts`
- Test: `test/terminal-manager.test.ts` or new frontend store test

**Interfaces:**
- Consumes: `TerminalTab`, `sessionId`
- Produces:
  - `isSessionCollapsed(sessionId: string): boolean`
  - `toggleSessionCollapse(sessionId: string, collapsed?: boolean): void`
  - `getSessionDockPosition(sessionId: string): 'bottom' | 'right'`
  - `setSessionDockPosition(sessionId: string, pos: 'bottom' | 'right'): void`
  - `renameTerminal(sessionId: string, termId: string, newTitle: string): void`
  - `panelWidth: number` (for right-docked terminal)

- [ ] **Step 1: Write test or verification harness for store methods**
- [ ] **Step 2: Add `sessionCollapsed`, `sessionDockPosition`, `panelWidth`, and `renameTerminal` in `TerminalStore`**
- [ ] **Step 3: Ensure session deletion cleans up session collapsed and dock preferences**
- [ ] **Step 4: Verify typecheck passes with `npm run svelte-check` or `npm run check`**
- [ ] **Step 5: Commit changes**

---

### Task 2: Implement Dynamic Theme Palette for Xterm.js (Light vs Dark)

**Files:**
- Modify: `frontend/src/lib/components/terminal/TerminalPanel.svelte`
- Consumes: `settingsStore.theme` from `$lib/stores/settings.svelte.ts`

- [ ] **Step 1: Define `LIGHT_TERMINAL_THEME` and `DARK_TERMINAL_THEME` token objects**
  - Light theme: background `#ffffff` or `#f8fafc`, foreground `#0f172a`, cursor `#0284c7`, selection `#bae6fd80`, ANSI colors optimized for light background.
  - Dark theme: current dark studio `#0e0e11` background and `#e4e4e7` foreground.
- [ ] **Step 2: Add `$effect` watching `settingsStore.theme` that applies `term.options.theme = newTheme` across all active terminal instances**
- [ ] **Step 3: Update terminal header styling with theme-aware classes (`bg-ant-bg-secondary`, `border-ant-border`, `text-ant-text`) replacing hardcoded `#0e0e11` and `#141418`**
- [ ] **Step 4: Commit changes**

---

### Task 3: Add Double-Click Inline Rename for Terminal Tabs

**Files:**
- Modify: `frontend/src/lib/components/terminal/TerminalPanel.svelte`

- [ ] **Step 1: Add state for `editingTermId: string | null` and `editingTitle: string`**
- [ ] **Step 2: Render inline input on `ondblclick` with auto-focus and auto-select**
- [ ] **Step 3: Handle `onkeydown` (Enter to save, Escape to cancel) and `onblur` to commit `terminalStore.renameTerminal(sessionId, tab.id, editingTitle.trim())`**
- [ ] **Step 4: Verify clicking other buttons (like close) does not trigger rename**
- [ ] **Step 5: Commit changes**

---

### Task 4: Flexible Docking (Bottom and Right) with Dual Resizers

**Files:**
- Modify: `frontend/src/lib/components/terminal/TerminalPanel.svelte`
- Modify: `frontend/src/App.svelte` (docking container arrangement around chat feed)

- [ ] **Step 1: Add dock toggle button in TerminalPanel header (Bottom vs Right icon / split)**
- [ ] **Step 2: In `App.svelte`, structure layout so if dock position is `right`, TerminalPanel renders side-by-side with MessageList + Composer**
- [ ] **Step 3: Implement horizontal resizer drag handle for `right` dock and vertical drag handle for `bottom` dock**
- [ ] **Step 4: Call `refitActiveTerminal()` on dock position change and during resizes**
- [ ] **Step 5: Commit changes**

---

### Task 5: Slim Collapsed Bar when Collapsed & Session Isolation

**Files:**
- Modify: `frontend/src/lib/components/terminal/TerminalPanel.svelte`

- [ ] **Step 1: Check conditions:**
  - If no terminals exist in session (`activeTerminals.length === 0`), do not render the bar or panel.
  - If at least 1 terminal exists and `isSessionCollapsed(sessionId)` is true, render a compact 28px bottom or right pill/strip showing terminal count, active tab title, and an expand icon.
  - If not collapsed, render the full terminal viewport.
- [ ] **Step 2: Confirm tab switching preserves isolated collapse state between Tab 1 and Tab 2**
- [ ] **Step 3: Commit changes**

---

### Task 6: End-to-End Visual & Functional Verification

- [ ] **Step 1: Run build verification (`npm run build` or frontend build)**
- [ ] **Step 2: Test light theme toggle: verify terminal background turns white and text remains crisp**
- [ ] **Step 3: Test bottom and right docking toggling**
- [ ] **Step 4: Test terminal tab double-click rename**
- [ ] **Step 5: Test session switching and verify Tab 1 and Tab 2 have independent collapse states**
