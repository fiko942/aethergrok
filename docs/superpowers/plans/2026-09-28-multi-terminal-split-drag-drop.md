# Implementation Plan: VS Code-Style Multi-Terminal Split & Drag-and-Drop

## Overview
Implement multi-terminal splitting (>2 terminals) with intuitive drag-and-drop tab splitting, interactive drop overlays, resizable dividers, and full compatibility across both Bottom Dock and Right Dock configurations in AetherGrok.

---

## Technical Work Breakdown

### Task 1: Store & Data Model Architecture (`frontend/src/lib/stores/terminal.svelte.ts`)
- [x] Create `TerminalPaneGroup` interface (`id`, `paneTermIds`, `splitDirection`, `paneSizes`).
- [x] Add reactive state:
  - `sessionSplitGroups: Record<string, TerminalPaneGroup[]>`
  - `focusedPaneTermId: Record<string, string>`
- [x] Add actions:
  - `splitTerminal(sessionId, sourceTermId, targetTermId, position)`
  - `unsplitTerminal(sessionId, termId)`
  - `setPaneSizes(sessionId, groupIdx, sizes)`
  - Update `closeTerminal(sessionId, termId)` to automatically remove from split group and re-expand sibling panes.

### Task 2: Drag-and-Drop Split Targets & Visual Overlays (`TerminalPanel.svelte`)
- [x] Make terminal header tabs draggable (`draggable="true"`, drag ghost preview).
- [x] Implement viewport drag-over coordinate detection (`left`, `right`, `bottom`, `top`).
- [x] Render animated semi-transparent drop zone overlays (`bg-ant-primary/15 border-dashed border-ant-primary`).

### Task 3: Multi-Pane Split Viewport & Resizable Dividers (`TerminalPanel.svelte`)
- [x] Render flex container for active split groups.
- [x] Implement interactive drag divider handles (`col-resize` / `row-resize`) between split panes.
- [x] Integrate mini-header on each pane with active focus ring, inline rename, unsplit button, and close (`X`) button.
- [x] Add top-level toolbar "Split Terminal" (`Columns2`) button.

### Task 4: ResizeObserver & Backend PTY Synchronization
- [x] Attach `ResizeObserver` to each split container node.
- [x] Trigger debounced `fitAddon.fit()` and `App.ResizeTerminal(termId, cols, rows)`.
- [x] Handle theme switching across all concurrently visible split terminals.

### Task 5: Testing & Verification
- [x] Verify multi-split with 2, 3, and 4 terminals in bottom dock mode.
- [x] Verify split behavior in right dock mode.
- [x] Verify tab dragging into split zones and unsplitting back to tabs.
- [x] Verify `npm run check` and `npm run build` in `frontend/`.
