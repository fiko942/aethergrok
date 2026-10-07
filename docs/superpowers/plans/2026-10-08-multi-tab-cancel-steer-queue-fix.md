# Implementation Plan - Multi-Tab Session Cancellation, Steer Reliability, Image Context Optimization, and Sequential Queue Management

Fix session stop/cancel reliability across multi-tab runs, eliminate steer deadlocks and session stuck states, prevent image payload memory/history bloat and startup timeouts, and enforce strict, safe sequential queue dispatching.

- **Status:** Proposed
- **Date:** 2026-10-08
- **Area:** Backend Process Management (`pkg/grokrunner`), Frontend Lifecycle (`App.svelte`, `session.svelte.ts`, `Composer.svelte`)

---

## 1. Problem Analysis & Root Causes

### 1.1 Stop / Cancel Failure Across Multi-Tab Sessions
- **Issue:** When running multiple tabs or stopping an active turn, the Stop button appears non-functional or fails to abort the running turn.
- **Root Causes:**
  1. Frontend `handleCancelSession()` in `App.svelte` unconditionally cancelled `activeSession.id`, instead of accepting an explicit `sessionId`. When a user triggered stop from a specific context or while switching tabs, it targeted the wrong session or failed silently.
  2. In `pkg/grokrunner/runner.go`, `Cancel(sessionID)` called `killProcessGroup(active.Cmd)` and `_ = active.Stdout.Close()`, but on Windows `taskkill /T /F /PID` can encounter race conditions where the process handle is in mid-exit or stdout scanner is blocking without context awareness.
  3. When cancel was called, `grok:complete` was emitted with status `interrupted`, which triggered `checkAndDispatchNextQueue` without checking if the user explicitly intended to stop the session queue or pause execution.

### 1.2 Steer Deadlock & "Thinking..." Stuck State
- **Issue:** Steering a queued prompt or active turn caused the UI to get stuck in thinking/spinning state, and restarting the app resumed the broken state.
- **Root Causes:**
  1. `handleSteerPrompt` ran `await handleCancelSession()`, immediately followed by `await executeTurn(..., { isSteer: true })`. However, in the Go backend, `StartSession` had to wait on `<-existing.Done` from the previous process before starting the new one. Because `Cancel` closed stdout without immediate fallback cleanup in some cases, `<-existing.Done` hung or blocked until timeout.
  2. The frontend `executeTurn` did not await backend termination settling, resulting in concurrent `RunPromptStream` calls colliding on the same session ID.
  3. On app restart, if a session was left in `status: 'working'`, it stayed in `'working'` with an animated spinner because `sessionStore` hydrated persisted `status: 'working'` without an active backend process attached.

### 1.3 Image Attachment Stalling & Timeout Failures
- **Issue:** Repeated prompts with attached images eventually caused sudden stalls, 45s startup timeouts, and unresponsive responses despite remaining context tokens.
- **Root Causes:**
  1. Temporary prompt files (`grok-prompt-*.json` and `grok-prompt-*.txt`) in `pkg/grokrunner/runner.go` created base64-encoded image payloads. When `--resume <uuid>` was used on multi-turn conversations containing multiple previous image turns, Grok CLI re-parsed prior image transcripts on disk on every prompt.
  2. The startup watchdog in `runner.go` was fixed to a hardcoded 45s `startupTimeout`. Complex multi-image payloads or heavy workspace scans exceeded this window before the first line was emitted, triggering `killProcessGroup` and terminating legitimate runs with a timeout error.
  3. Image file paths referenced in `VisionImage` were not validated for existence prior to payload construction, causing silent failures or empty data encoding.

### 1.4 Immediate Cascade / Accidental Burst of Queued Prompts on Stop
- **Issue:** When a turn was cancelled, all queued prompts flushed or triggered at once instead of remaining queued or executing sequentially.
- **Root Causes:**
  1. `grok:complete` and `grok:error` listeners both invoked `checkAndDispatchNextQueue(event.sessionId)`. When user hit "Stop", the turn completed with status `interrupted`, which immediately popped the next item from `queuedPrompts` and executed it.
  2. To the user, stopping a turn looked like it immediately started the next queue item, creating a cascade where stopping repeatedly fired every prompt in line.
  3. The queue dispatch logic lacked a paused/cancelled guard and did not verify session idle readiness before popping.

---

## 2. Proposed Changes & Technical Architecture

### 2.1 Backend (`pkg/grokrunner`)
- **File:** `pkg/grokrunner/runner.go`
  - Enhance `Cancel(sessionID string)` with graceful timeout, non-blocking cleanup channel, and immediate OS process group termination across Windows, macOS, and Linux.
  - Dynamically scale startup watchdog timeout when image attachments are present (e.g. 75s for vision prompts vs 45s standard) to accommodate image encoding and transcript loading.
  - Ensure temp files (`grok-prompt-*.json`, `grok-prompt-*.txt`) are tracked in `ActiveSession` and guaranteed to be deleted immediately upon turn completion or cancellation.
  - Add session state check so starting a turn on a session with a currently running process cleanly cancels and drains the prior process before spawning.

### 2.2 Frontend Session & Queue Lifecycle (`frontend/src/App.svelte` & `frontend/src/lib/stores/session.svelte.ts`)
- **File:** `frontend/src/lib/stores/session.svelte.ts`
  - In `loadSessionsFromStorage()` and `initialize()`, automatically sanitize session statuses: any session loaded with `status: 'working'` must be reset to `status: 'idle'` on application startup.
  - Add `clearQueuedPrompts(sessionId: string)` and `pauseQueue(sessionId: string)` methods.
- **File:** `frontend/src/App.svelte`
  - Refactor `handleCancelSession(targetSessionId?: string, clearQueue = false)`:
    - Target the specific `targetSessionId` or `activeSession.id`.
    - If `clearQueue` is true (or user clicks stop button), prevent `checkAndDispatchNextQueue` from automatically firing the next prompt.
    - Set explicit `status: 'idle'` and add user-visible notice.
  - Refactor `handleSteerPrompt(promptItem: QueuedPrompt)`:
    - Remove prompt from queue.
    - Cancel active process and await cancellation confirmation.
    - Dispatch steer prompt turn with clean payload.
  - Refactor `checkAndDispatchNextQueue(sessionId: string)`:
    - Check if session is actually `idle` and not `cancelled` or `stopped`.
    - Add a debounce and execution guard so multiple events cannot trigger duplicate concurrent turn dispatches.

---

## 3. Step-by-Step Implementation Tasks

### Task 1: Startup Session State Sanitization (Fix "Thinking..." on App Restart)
- **Files:** `frontend/src/lib/stores/session.svelte.ts`, `frontend/src/lib/stores/session.svelte.test.ts` (or unit test)
- **Changes:**
  - When loading sessions from localStorage (`loadSessionsFromStorage`), map any session with `status === 'working'` to `status: 'idle'`.
  - Reset any lingering tool calls with `status === 'running'` to `status: 'completed'` or `'error'`.

### Task 2: Robust Backend Process Cancellation & Watchdog Tuning
- **Files:** `pkg/grokrunner/runner.go`, `pkg/grokrunner/runner_test.go`
- **Changes:**
  - Ensure `Cancel(sessionID)` closes pipes, cancels context, and kills process group with error handling.
  - Increase vision prompt startup watchdog timeout and ensure prompt files are deleted in `defer`.

### Task 3: Controlled Sequential Queue Management & Stop Button Logic
- **Files:** `frontend/src/App.svelte`, `frontend/src/lib/components/chat/Composer.svelte`
- **Changes:**
  - Update `handleCancelSession(sessionId?: string, abortQueue?: boolean)`.
  - When user clicks Stop in Composer (`onCancel`), abort the turn and retain the queue without auto-dispatching.
  - Update `grok:complete` and `grok:error` event handlers to only auto-dequeue on natural `success` completions, not on user-initiated `interrupted` stops.

### Task 4: Fix Steer Prompt Synchronization
- **Files:** `frontend/src/App.svelte`
- **Changes:**
  - In `handleSteerPrompt`, ensure cancellation settles before invoking `executeTurn`.
  - Guarantee `isSteer: true` is properly flagged and does not leave orphaned queue items.

### Task 5: Testing & Verification
- **Tests to run:**
  - `go test -v ./pkg/grokrunner/...`
  - `npx vitest run`
  - End-to-end verification of multi-tab switching, prompt queueing, steer, stop button, and image prompt execution.

---

## Execution Handoff
Plan complete and saved to `docs/superpowers/plans/2026-10-08-multi-tab-cancel-steer-queue-fix.md`. Two execution options:

1. **Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration.
2. **Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints.

Which approach would you like to take?
