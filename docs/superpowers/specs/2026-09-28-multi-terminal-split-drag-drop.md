# Multi-Terminal Split & Drag-and-Drop Architecture (VS Code-Style)

## Context & Vision
In AetherGrok's desktop workspace, users require a robust, flexible terminal workspace capable of running multiple processes concurrently (e.g. backend servers, frontend dev servers, git watchers, build commands). To elevate developer productivity, AetherGrok introduces a VS Code-inspired multi-terminal split architecture featuring interactive drag-and-drop tab manipulation, multi-pane (>2) flexible layouts, and persistent PTY process isolation across both **Bottom Dock** and **Right Dock** modes.

---

## Architecture & Core Capabilities

### 1. Drag-and-Drop Terminal Splitting
- **Tab Drag Source**:
  - Each tab in the terminal header is draggable (`draggable="true"`).
  - Drag payloads convey the terminal identifier (`termId`).
- **Interactive Drop Target Overlays**:
  - Dragging a terminal tab across the active terminal viewport dynamically computes relative mouse coordinates:
    - **Right zone (`x > 60%`)**: Splits the target terminal horizontally to the right.
    - **Left zone (`x < 40%`)**: Splits the target terminal horizontally to the left.
    - **Bottom zone (`y > 60%`)**: Splits the target terminal vertically downward.
  - Visual preview overlay (`bg-ant-primary/15 border-2 border-dashed border-ant-primary rounded-lg`) provides instant visual confirmation of the target split configuration prior to dropping.

### 2. Multi-Terminal Splitting (> 2 Terminals)
- Supports arbitrary N-way splits (e.g., `Terminal 1 | Terminal 2 | Terminal 3`).
- Seamless layout management via proportional flex containers with custom interactive resize dividers (`col-resize` / `row-resize`).
- Each split pane maintains independent scrollback, input buffers, and cursor styling.

### 3. Dual Dock Compatibility (Bottom & Right)
- **Bottom Dock**:
  - Leverages wide horizontal space with default horizontal split columns.
- **Right Dock**:
  - Responsive layout adapts to narrower widths with vertical row stacking or horizontal splitting when sidebar width allows.
  - Minimum width / height thresholds safeguard readable PTY columns (minimum 180px width, 100px height).

### 4. Focused State & Un-Splitting Mechanics
- **Active Focus Ring**: Active terminal pane highlighted with `border-ant-primary` status border. Clicking anywhere in a pane transfers immediate keyboard input focus.
- **Pane Action Header**:
  - Double-click to rename terminal title inline.
  - One-click **Unsplit** button to detach a pane back into a standalone top-level tab.
  - Maximize / minimize individual pane.
  - Kill / Close (`X`) button cleanly terminates the specific PTY process group without affecting adjacent split siblings.
- **Toolbar Quick Split**:
  - Dedicated **Split Terminal** button (`Columns2` icon) on the terminal toolbar to instantly spawn and split a terminal with a single click.

---

## Technical Specifications & Lifecycle Safeguards

### A. PTY Backend Integrity (`pkg/terminal` & `app.go`)
- Terminals continue to operate via individual, isolated Go PTY process groups (`termId`).
- Splitting is handled at the presentation/viewport layout level without disrupting backend process trees or SIGKILL signals.

### B. Dynamic Resize Handling (`ResizeObserver` & `fitAddon`)
- Every split pane container attaches an independent `ResizeObserver`.
- Automatic debounced execution of `fitAddon.fit()` recalculates character dimensions and dispatches `App.ResizeTerminal(termId, cols, rows)` to the Go backend.
- Prevents zero-dimension DOM errors and ensures sharp, non-clipped typography during manual panel resizing.
