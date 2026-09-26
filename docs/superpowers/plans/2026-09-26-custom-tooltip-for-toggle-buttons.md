# Plan: Custom Dark Studio Tooltip for Header & Sidebar Toggle Buttons

## Overview
Currently, the sidebar toggle buttons in `App.svelte` (the left sidebar toggle `⌘B`, the collapsed floating edge button `⌘B`, and the right workspace inspector toggle `⌘⌥B`) and in `RightSidebar.svelte` use default browser HTML `title="..."` attributes. The user requested that the circular toggle button (and related toggle actions) use our custom polished Dark Studio `Tooltip.svelte` component with proper hotkey badge rendering.

## Objectives
1. **Left Sidebar Header Toggle Button (`App.svelte`)**:
   - Wrap with `<Tooltip title={settingsStore.sidebarCollapsed ? "Expand Sidebar" : "Collapse Sidebar"} shortcut={isMac ? "⌘B" : "Ctrl+B"} placement="bottom">`.
   - Remove HTML `title="..."` attribute.
2. **Left Sidebar Floating Collapsed Button (`App.svelte`)**:
   - Wrap with `<Tooltip title="Expand Sidebar" shortcut={isMac ? "⌘B" : "Ctrl+B"} placement="right">`.
   - Remove HTML `title="..."` attribute.
3. **Right Sidebar / Workspace Inspector Toggle Button (`App.svelte`)**:
   - Wrap with `<Tooltip title={currentSession?.rightSidebarOpen ? "Close Inspector" : "Open Inspector"} shortcut={isMac ? "⌘⌥B" : "Ctrl+Alt+B"} placement="bottom">`.
   - Remove HTML `title="..."` attribute.
4. **Right Sidebar Close Button (`RightSidebar.svelte`)**:
   - Wrap with `<Tooltip title="Close Inspector" shortcut={isMac ? "⌘⌥B" : "Ctrl+Alt+B"} placement="left">`.
   - Remove HTML `title="..."` attribute.
5. **Verification**:
   - Run `pnpm run check`, `pnpm run build`, and `go test ./test/... -v`.
   - Commit changes to git.

## Implementation Steps
- **Task 1**: Update `frontend/src/App.svelte` to wrap header toggle and floating collapsed toggle with `<Tooltip>`.
- **Task 2**: Update `frontend/src/lib/components/layout/RightSidebar.svelte` to wrap close toggle with `<Tooltip>`.
- **Task 3**: Verify types, build bundle, run tests, and commit.
