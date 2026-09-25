# Confirm Modal Refinement & Full-App Backdrop Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:**
1. Clean up `ModalConfirm.svelte` design:
   - Remove loud icons, loud badge backgrounds, and remove the unnecessary "Esc Cancel / Enter Confirm" keyboard instruction text.
   - Distinctively differentiate the destructive Action button (e.g. solid red danger button) from the neutral Cancel button (ghost/subtle border).
   - Minimalist, elegant design matching Anthropic Serif aesthetics without visual noise.
2. Ensure the confirmation dialog sits at the root application level (`App.svelte`) so that the entire window (including Composer prompt box, sidebar, message feed, and header) is dimmed and blurred behind the modal backdrop.

**Architecture:**
- Move `ModalConfirm` instance from `WorkspaceSidebar.svelte` to `App.svelte` or pass global confirm state via a store or callback so its backdrop overlays the whole viewport (`fixed inset-0 z-[10000]`), covering the prompt bar.
- Simplify `ModalConfirm.svelte` layout: clean typography, no colorful icon box, clear visual hierarchy between Cancel (muted secondary) and Action (solid accent or danger).

---

### Task 1: Refine `ModalConfirm.svelte` Design

**Files:**
- Modify: `frontend/src/lib/antd/ModalConfirm.svelte`

- [ ] **Step 1: Simplify structure & typography**
Remove the colorful icon square and the keyboard help text (`Esc Cancel / Confirm`).
Differentiate Cancel button (`type="default"` with muted text) and Confirm/Remove button (`type="danger"` with distinct red styling or filled primary).

---

### Task 2: Mount ModalConfirm at Root Level in `App.svelte`

**Files:**
- Modify: `frontend/src/lib/stores/session.svelte.ts` (or confirm store)
- Modify: `frontend/src/lib/components/layout/WorkspaceSidebar.svelte`
- Modify: `frontend/src/App.svelte`

- [ ] **Step 1: Create confirm modal state in store or exportable modal trigger**
Provide a global dialog trigger so any component (sidebar, tabs, messages) can request a confirmation that blurs the entire app including the prompt composer.

- [ ] **Step 2: Render ModalConfirm at the bottom of `App.svelte`**
Ensure the backdrop has `fixed inset-0 z-[99999] bg-black/70 backdrop-blur-md` covering header, sidebar, message list, and composer prompt bar.

---

### Task 3: Build & Verify

- [ ] **Step 1: Run `npm run build` and `wails build -clean`**
- [ ] **Step 2: Relaunch AetherGrok and verify that opening the delete confirmation blurs the whole app and displays clean, differentiated buttons without UI clutter.**
