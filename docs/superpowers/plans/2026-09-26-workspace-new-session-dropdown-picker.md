# Plan: Workspace-Aware New Conversation Dropdown Picker

## Overview
Currently, clicking the `+` (New Session) button in the top tab bar or the "New Conversation" button on the zero-tab empty state blindly creates a session in whatever workspace happens to be active or fallback. The user wants both buttons to present an elegant dropdown/popover menu that lists all existing workspaces (e.g. `affilia`, `sinematika`), clearly indicating which workspace the new session will belong to, plus an option `+ Open Folder / New Workspace...` to choose a new directory via native OS folder picker.

## Objectives
1. **Interactive Workspace Picker Dropdown Component**:
   - Create a reusable, polished Dark Studio dropdown/popover (`NewSessionDropdown.svelte` or built into `SessionTabs.svelte` and zero-tab state).
   - Display a list of all configured workspace folders with their folder icons, session counts, and active workspace indicator.
   - Include a divider and `+ Open New Workspace Folder...` option that calls `window.go.main.App.SelectWorkspaceDirectory()`.
2. **Tab Bar `+` Button Integration**:
   - When clicking the `+` button in `SessionTabs.svelte`, toggle the dropdown anchored right below the plus button.
   - If there is only 1 workspace and user clicks directly without holding or clicks an item, create a session in that selected workspace.
3. **Center Zero-Tab "New Conversation" Button Integration**:
   - In `App.svelte`, clicking "New Conversation" displays the same rich workspace dropdown menu centered below the button or opens the picker so user can choose which workspace to start in.
4. **Keyboard & Outside Click Handling**:
   - Closes automatically on `Escape` or clicking outside.
   - Supports keyboard navigation (Enter/Click).
5. **Verification**:
   - `pnpm run check`, `pnpm run build`, and `go test ./test/... -v`.

## Implementation Steps
- **Task 1**: Create `frontend/src/lib/components/layout/NewSessionDropdown.svelte` (or unified dropdown controller) with Dark Studio styling, glassmorphism, workspace listing, badge counts, active checkmarks, and `+ Open Folder` action.
- **Task 2**: Integrate into `SessionTabs.svelte` on the `+` button with click-outside listener.
- **Task 3**: Integrate into `App.svelte` zero-tab empty state "New Conversation" button.
- **Task 4**: Verify and commit.
