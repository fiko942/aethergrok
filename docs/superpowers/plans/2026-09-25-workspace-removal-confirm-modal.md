# Workspace Removal, Custom Confirm Modal, and Session Limit Reset Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 
1. Auto-reset pagination limit back to 8 when a workspace accordion folder is closed / toggled off.
2. Allow any workspace (existing or missing) to be removed from the application's workspace list (without deleting disk files).
3. Create a reusable custom Ant Design Dark confirmation modal component (`ModalConfirm.svelte`) supporting keyboard shortcuts (`Escape` to cancel, `Enter` to confirm).
4. Connect the custom confirm modal to both Workspace removal and Session deletion actions.

**Architecture:**
- Create `frontend/src/lib/antd/ModalConfirm.svelte` (or `Modal.svelte`) with backdrop blur, dark tokens (`bg-ant-bg-secondary`, `border-white/10`, `text-ant-text`), danger / warning / info variants, confirm button, cancel button, and keyboard listeners (`Enter` / `Escape`).
- In `WorkspaceSidebar.svelte`:
  - When a workspace is collapsed (`toggleWorkspaceExpanded`), reset `workspaceVisibleCounts[wsId] = PAGE_SIZE (8)`.
  - Add a delete / remove action button on every workspace folder row (for both active and missing folders).
  - Replace native browser `confirm()` with the new `ModalConfirm` component for both single session deletion and workspace removal.

**Tech Stack:** Svelte 5 (Runes), Tailwind CSS 3.4, Ant Design Dark Tokens, Wails v2.

---

### Task 1: Create Reusable Custom ModalConfirm Component

**Files:**
- Create: `frontend/src/lib/antd/ModalConfirm.svelte`

**Interfaces:**
- Props:
  - `open: boolean`
  - `title: string`
  - `content: string`
  - `confirmText?: string` (default: "Confirm")
  - `cancelText?: string` (default: "Cancel")
  - `type?: 'danger' | 'warning' | 'info'` (default: "danger")
  - `onConfirm: () => void`
  - `onCancel: () => void`
- Keyboard shortcuts:
  - `Escape`: triggers `onCancel()`
  - `Enter`: triggers `onConfirm()`

- [ ] **Step 1: Write `ModalConfirm.svelte`**
Implement the modal with subtle dark borders (`border-white/10`, `bg-ant-bg-secondary`), danger styling for destructive actions, autofocus/keyboard listeners for Escape and Enter.

- [ ] **Step 2: Verify component builds**
Run: `npm --prefix frontend run build`
Expected: 0 errors.

---

### Task 2: Implement Pagination Reset on Workspace Collapse

**Files:**
- Modify: `frontend/src/lib/components/layout/WorkspaceSidebar.svelte`

**Interfaces:**
- Consumes: `workspaceVisibleCounts`, `sessionStore.toggleWorkspaceExpanded(wsId)`
- Produces: Resets `workspaceVisibleCounts[wsId] = 8` whenever workspace is collapsed.

- [ ] **Step 1: Update toggle collapse function in `WorkspaceSidebar.svelte`**
When toggling workspace open/close, if it is currently expanded and transitioning to closed (or vice versa), reset `workspaceVisibleCounts[wsId] = PAGE_SIZE`.

---

### Task 3: Integrate Workspace Removal & Session Deletion with ModalConfirm

**Files:**
- Modify: `frontend/src/lib/components/layout/WorkspaceSidebar.svelte`

**Interfaces:**
- Consumes: `ModalConfirm.svelte`, `sessionStore.removeWorkspace`, `sessionStore.closeSession`
- Produces: Workspace delete button on hover, confirms via `ModalConfirm`, deletes session cleanly.

- [ ] **Step 1: Add state for confirm dialog in `WorkspaceSidebar.svelte`**
Track `confirmState: { open: boolean, title: string, content: string, onConfirm: () => void }`.

- [ ] **Step 2: Add Remove Workspace action button to normal workspace header**
In the workspace header actions area, add a remove trash button that prompts confirmation to remove the workspace from the list.

- [ ] **Step 3: Connect Session Delete dropdown item to `ModalConfirm`**
When clicking "Delete Session", open `ModalConfirm` with session details instead of native `confirm()`.

---

### Task 4: Rebuild Wails App and Verify

**Files:**
- Build output: `build/bin/aethergrok.app`

- [ ] **Step 1: Run `npm run build` and `wails build -clean`**
- [ ] **Step 2: Verify workspace delete, session delete, and pagination reset in AetherGrok**
