# Complete Grok Session Deletion & Completed Session Read-State Sync Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:**
1. **Grok Session Deletion on Disk**:
   - When deleting a session, resolve the actual Grok session UUID directory in `~/.grok/sessions/<encoded_workspace_path>/<session_id>` (supporting `session.grokSessionId` as well as `session.id`).
   - Call Go backend `DeleteGrokSession(workspacePath, grokSessionId || sessionId)` to completely remove the session record, transcript, and directory from disk.
2. **Auto-Reset Finished Status Icon to Default Blue Chat Icon on Bottom Scroll / View**:
   - When a session has status `finished` (green checkmark), viewing and scrolling to the bottom of the conversation feed (acknowledging the latest assistant turn) marks the session read/acknowledged and transitions `session.status` to `idle`.
   - The sidebar session icon immediately transitions back to the default blue message icon (`MessageSquare`).

---

### Task 1: Robust Disk-Level Grok Session Deletion

**Files:**
- Modify: `pkg/grokrunner/session_scanner.go`
- Modify: `frontend/src/lib/components/layout/WorkspaceSidebar.svelte`

- [ ] **Step 1: Enhance `DeleteGrokSessionDirectory` in Go**
In `session_scanner.go`, ensure deleting searches and deletes both the exact directory name and any matching prefix or alternate ID.
- [ ] **Step 2: Update `WorkspaceSidebar.svelte` deletion invocation**
Pass the real Grok session ID (`session.grokSessionId || session.id`) to `window.go.main.App.DeleteGrokSession(ws.path, targetSessionId)`.

---

### Task 2: Auto-Clear "Finished" Checkmark Status on Viewing & Scrolling to Bottom

**Files:**
- Modify: `frontend/src/lib/components/chat/MessageList.svelte`
- Modify: `frontend/src/lib/stores/session.svelte.ts`

- [ ] **Step 1: Add store helper `markSessionFinishedAcknowledged(sessionId)` or set status to `idle`**
In `sessionStore`, when an active session's latest turn has been seen (scrolled to bottom or switched to), reset `status = 'idle'`.
- [ ] **Step 2: Connect `MessageList.svelte` scroll & mount to clear finished state**
In `handleScroll()` and on initial scroll to bottom when `distanceFromBottom < 100`, if `activeSession.status === 'finished'`, reset status to `idle`.

---

### Task 3: Build & Verification

- [ ] **Step 1: Run `npm run build` and `wails build -clean`**
- [ ] **Step 2: Test session deletion on disk and checkmark reset on scroll in AetherGrok**
