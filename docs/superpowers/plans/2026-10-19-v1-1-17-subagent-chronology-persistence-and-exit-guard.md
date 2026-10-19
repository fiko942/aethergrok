# Subagent Chronology, Activity Persistence, Plan Filtering & Fail-Closed Lifecycle Guard (v1.1.17) Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement or audit this plan task-by-task. Steps use checkbox (`- [x]`) syntax for tracking.

**Goal:** Deliver chronological message/tool interleaving, subagent activity persistence across restarts, intelligent todo/diff formatting, running subagent stopwatch timers with auto-scrolling lists, "main" plan tracker defaulting, and a fail-closed lifecycle exit guard with input shield to eliminate accidental application closing while tasks run.

**Architecture:** Synchronize active frontend task counts with Go Wails backend (`SetFrontendActiveTasksCount`), enforce fail-closed confirmation in `beforeClose`, protect keyboard events in `App.svelte`, deep-scan subagent transcripts in `pkg/grokrunner/session_scanner.go`, interleave message content blocks chronologically in `MessageItem.svelte`, and default plan filters to "main" in `FloatingPlanTracker.svelte` and `PlanPanel.svelte`.

**Tech Stack:** Svelte 5 (Runes `$state`, `$derived`, `$effect`), Wails v2 (Go 1.24+), TypeScript, Vitest, GitHub Actions CI/CD.

---

### Task 1: Accidental App Exit Prevention & Fail-Closed Lifecycle Guard

**Files:**
- Modify: `app.go`
- Modify: `frontend/src/App.svelte`
- Modify: `frontend/src/lib/stores/session.svelte.ts`
- Create: `app_test.go`
- Create: `test/shortcut-window-close.test.ts`
- Modify: `test/session-running-guard.test.ts`

- [x] **Step 1: Frontend Input Shielding**
  - Suppressed single-modifier shortcuts (`ShiftRight`, `ShiftLeft`) while the user is actively typing in `<input>` or `<textarea>` elements to prevent accidental snapshot and window trigger events.
  - Intercepted `Ctrl+W` in `isEditing` state to preserve readline unix-word-rubout and prevent tab/window termination.

- [x] **Step 2: Reactive Active Tasks Synchronization**
  - Exposed `SetFrontendActiveTasksCount(count int)` on Go `App` struct.
  - Implemented reactive `$effect` in `frontend/src/App.svelte` and `session.svelte.ts` computing total running prompts, tools, and subagents across all open tabs.
  - Automatically dispatches updated count to backend on every task state change.

- [x] **Step 3: Go Lifecycle `beforeClose` Fail-Closed Safety**
  - Enhanced `beforeClose` handler in `app.go` to inspect `a.frontendActiveTasksCount > 0` and active CLI runners.
  - Handled nil dialog context or headless environments with fail-closed behavior (returns `true` to abort quit and protect running tasks).
  - Shows modal confirmation dialog ("Tasks in progress. Are you sure you want to quit?") if the user attempts to close the window while tasks are running.

---

### Task 2: Subagent Stopwatch Timer, Auto-Scroll & Activity Formatting

**Files:**
- Modify: `frontend/src/lib/components/chat/ToolCallCard.svelte`
- Modify: `frontend/src/lib/components/chat/SubagentDetailModal.svelte`
- Modify: `frontend/src/lib/utils/toolUtils.ts`
- Modify: `test/toolUtils.test.ts`
- Modify: `test/subagent-card-render.test.ts`

- [x] **Step 1: Subagent Stopwatch Timer**
  - Added live elapsed timer running alongside the lightning bolt icon and inspect button.
  - Automatically starts on subagent launch, updates every second, and latches to final elapsed duration upon completion or cancellation.

- [x] **Step 2: Activity List Auto-Scroll**
  - Implemented automatic scroll-to-bottom for subagent activity streams in both expanded card and full detail modal.
  - Ensures the most recent tool actions, file inspections, and updates remain visible without manual scrolling.

- [x] **Step 3: Clean Todo and Tool Call Formatting**
  - Extended `toolUtils.ts` to parse `todo_write` arguments (`{ merge: true, todos: [...] }`).
  - Rendered clean todo badge status lists (`completed`, `in_progress`, `pending`) rather than raw JSON strings.
  - Formatted file paths, diff snippets, and write operations cleanly.

---

### Task 3: Chronological Message Interleaving & Subagent History Persistence

**Files:**
- Modify: `frontend/src/lib/components/chat/MessageItem.svelte`
- Modify: `pkg/grokrunner/session_scanner.go`
- Modify: `pkg/grokrunner/types.go`
- Create: `test/interleaved-message-parts.test.ts`
- Modify: `pkg/grokrunner/session_scanner_test.go`

- [x] **Step 1: Chronological Text & Tool Call Interleaving**
  - Refactored `MessageItem.svelte` to render message segments in chronological sequence.
  - Prevents premature rendering of subsequent tool calls above preceding narrative text when models emit interleaved thoughts and tools.

- [x] **Step 2: Subagent Transcript Deep Scanning & Disk Persistence**
  - Enhanced `pkg/grokrunner/session_scanner.go` to traverse subagent session directories.
  - Hydrates child tool calls, prompts, instructions, and results into session state.
  - Reconstructed full subagent history upon app restart or tab reactivation without data loss.

---

### Task 4: Plan Filtering Default to "Main"

**Files:**
- Modify: `frontend/src/lib/components/chat/FloatingPlanTracker.svelte`
- Modify: `frontend/src/lib/components/workspace/PlanPanel.svelte`

- [x] **Step 1: Set Default Plan Scope to "main"**
  - Set default plan filter to "main" (Grok Plan) rather than "all" in both floating tracker and sidebar panel.
  - Eliminates visual clutter where internal subagent todos were overwhelming primary session milestone tracking.

---

### Task 5: Multi-Platform Build, CI/CD & Local Installation

**Files:**
- Modify: `wails.json`
- Modify: `app.go`
- Modify: `frontend/package.json`
- Modify: `build/windows/installer/wails_tools.nsh`
- Modify: `changelog.json`
- Create: Annotated Git tag `v1.1.17`

- [x] **Step 1: Version Bump to v1.1.17**
  - Synchronized `1.1.17` across all project manifests and changelog.
  - Generated updated frontend production bundle and MD5 hash.

- [x] **Step 2: Push and Monitor GitHub Actions**
  - Tagged `v1.1.17` and pushed to GitHub remote.
  - Monitored workflow run `39238552962` until all 6 matrix jobs finished with 100% success.

- [x] **Step 3: Verification & Local Installation**
  - Downloaded `AetherGrok-1.1.17-macOS-arm64.dmg` and verified SHA256.
  - Installed `AetherGrok.app` to `/Applications/AetherGrok.app` and verified version `1.1.17`.
