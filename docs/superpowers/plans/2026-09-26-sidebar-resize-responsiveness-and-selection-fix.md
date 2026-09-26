# Ultra-Responsive Sidebar Resizing & Text Selection Prevention Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Eliminate lag/drag sluggishness and completely prevent unwanted text selection across chat/session contents when dragging both the left and right sidebars.

**Root Cause Analysis:**
1. **Unwanted Text Selection on Left Drag:**
   - In `App.svelte`: `handleResizeStart(e)` did NOT call `e.preventDefault()`. When a user presses down mouse on the divider handle and drags across the screen, the browser initiates native text selection across `<main>` and conversation turns.
   - `document.body.style.userSelect = 'none'` after mousedown is too late if native text selection is already triggered.
2. **Lag / Sluggishness on Left Sidebar (`App.svelte`):**
   - The `<aside>` element had CSS class: `transition-[width] duration-200 ease-out`. Every pixel movement during dragging was delayed/interpolated by CSS transition animations (200ms ease-out lag) while updating Svelte reactive state `settingsStore.sidebarWidth`!
   - Need to disable CSS transition completely when `isDraggingSidebar` is active: `isDraggingSidebar ? 'transition-none' : 'transition-[width] duration-200 ease-out'`.
   - Also use `requestAnimationFrame` throttling or direct reactive assignment to maintain a buttery 60/120fps drag response.
3. **Lag / Sluggishness on Right Sidebar (`RightSidebar.svelte`):**
   - In `RightSidebar.svelte`: `<aside>` has class `settingsStore.animationsEnabled ? 'transition-all duration-300 ease-out' : ''`.
   - During `isDragging`, it was fighting a 300ms CSS transition on width!
   - Disabling CSS transition during drag: `isDragging ? 'transition-none' : (settingsStore.animationsEnabled ? 'transition-[width] duration-200 ease-out' : '')`.
   - Set `document.body.style.userSelect = 'none'` and `cursor = 'col-resize'` in `startResize` and clean up in `handleResizeEnd`.

**Tech Stack:** Svelte 5 Runes, TypeScript, Tailwind CSS.

---

### Task 1: Fix Left Sidebar Drag Responsiveness and Text Selection in `App.svelte`

**Files:**
- Modify: `frontend/src/App.svelte`

- [ ] **Step 1: Update `handleResizeStart`, `handleResizeMove`, and `handleResizeEnd` in `App.svelte`**
- Call `e.preventDefault()` on `handleResizeStart`.
- Add `select-none` class to `document.body` or overlay during drag.
- Remove CSS `transition-[width]` while dragging (`isDraggingSidebar ? 'transition-none' : 'transition-[width] duration-200 ease-out'`).
- Ensure `window.getSelection()?.removeAllRanges()` is cleared.

- [ ] **Step 2: Verify `pnpm --prefix frontend run check`**

---

### Task 2: Fix Right Sidebar Drag Responsiveness and Text Selection in `RightSidebar.svelte`

**Files:**
- Modify: `frontend/src/lib/components/layout/RightSidebar.svelte`

- [ ] **Step 1: Update drag handlers and disable CSS transition during drag**
- Call `e.preventDefault()` and set global `document.body.style.userSelect = 'none'` and `cursor = 'col-resize'`.
- Remove CSS transition when `isDragging` is true (`isDragging ? 'transition-none' : ...`).
- Clean up styles on `mouseup`.

- [ ] **Step 2: Verify `pnpm --prefix frontend run check` and `pnpm --prefix frontend run build`**

---

### Task 3: Verification & Backend Integrity

- [ ] **Step 1: Run `go test ./test/... -v` and frontend checks**
- [ ] **Step 2: Commit all changes cleanly**
