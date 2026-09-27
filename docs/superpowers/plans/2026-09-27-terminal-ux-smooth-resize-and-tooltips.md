# Terminal UX Improvements: Auto Select-All on Rename, Custom Tooltips, Smooth Resizing & Text Selection Guard Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement auto-select-all on double-click inline renaming, integrate Ant Design custom tooltips across all terminal action icons and tabs, optimize drag-resizing with `requestAnimationFrame` + debounce refit, and eliminate accidental text selection during resize.

**Root Causes & Architecture:**
1. **Text selection during resize:** Dragging over the window triggers browser default selection and Xterm text selection. Fix: add `select-none` to body/window during drag, set `pointer-events-none` on the xterm canvas iframe/container while `isDragging === true`, and invoke `window.getSelection()?.removeAllRanges()`.
2. **Laggy resize performance:** Firing `fitAddon.fit()` and backend IPC `ResizeTerminal` synchronously on every mousemove event causes severe IPC and DOM layout thrashing. Fix: decouple pixel dimension updates from Xterm refit using `requestAnimationFrame` for CSS dimensions and debouncing `fit()` / backend `ResizeTerminal` (e.g. 60-80ms or on `handleResizeEnd`).
3. **Double click select-all:** Svelte action / `onfocus={(e) => e.currentTarget.select()}` + `inputRef?.select()` upon mount in inline rename mode so the text is instantly selected for immediate replacement or single-click repositioning.
4. **Rich Custom Tooltips:** Wrap every action button (`Send to Agent`, `Clear Terminal`, `Dock Position Toggle`, `Maximize/Restore`, `Collapse Panel`, `Close Tab`, `New Terminal Tab`) with the app's standard `Tooltip.svelte` component.

**Tech Stack:** Svelte 5, `@xterm/xterm`, Tailwind CSS, `Tooltip.svelte` component.

---

## Global Constraints
- State points directly in affirmative language.
- Ensure terminal process management and layout state remain stable and reactive.

---

## Detailed Task Breakdown

### Task 1: Add Auto Select-All on Double Click Rename
- Modify: `frontend/src/lib/components/terminal/TerminalPanel.svelte`
- Add action `selectOnFocus(node: HTMLInputElement)`:
  ```ts
  function selectOnFocus(node: HTMLInputElement) {
    node.focus();
    node.select();
  }
  ```
- Use `use:selectOnFocus` on the inline rename `<input>` element so double-clicking immediately highlights the entire title.

### Task 2: Fix Resize Lag & Text Selection Bug
- Modify: `frontend/src/lib/components/terminal/TerminalPanel.svelte`
- On `handleResizeStart`:
  - Set `document.body.style.userSelect = 'none'` and `document.body.style.cursor = dockPosition === 'bottom' ? 'row-resize' : 'col-resize'`.
  - Clear any active text selection: `window.getSelection()?.removeAllRanges()`.
- On `handleResizeMove`:
  - Throttle state updates using `requestAnimationFrame`.
  - Only update `panelHeight` or `panelWidth`. Do not synchronously call `fit()` or `ResizeTerminal` on every pixel change.
- On `handleResizeEnd`:
  - Restore `document.body.style.userSelect = ''` and `document.body.style.cursor = ''`.
  - Trigger final `refitActiveTerminal()` to adjust PTY rows/cols accurately.
- Add `pointer-events-none` overlay or class on terminal canvas container when `isDragging === true` so Xterm doesn't intercept drag coordinates.

### Task 3: Add Custom Tooltips to All Terminal Action Buttons and Tabs
- Modify: `frontend/src/lib/components/terminal/TerminalPanel.svelte`
- Wrap all icon buttons with `<Tooltip title="..." placement="top" ...>`:
  - Add New Terminal (`+`): "New Terminal"
  - Send to Agent: "Attach recent terminal output to composer"
  - Clear (`Trash2`): "Clear buffer"
  - Dock Toggle (`PanelRight` / `PanelBottom`): "Dock to right" / "Dock to bottom"
  - Maximize / Restore (`Maximize2` / `Minimize2`): "Maximize terminal" / "Restore size"
  - Collapse (`ChevronDown`): "Collapse terminal"
  - Close Tab (`X`): "Kill process and close tab"

### Task 4: Verification & Native Build
- Run Vite compilation and build verification.
- Rebuild macOS native bundle with `./build-macos.sh`.
