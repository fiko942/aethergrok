# Implementation Plan: Fix Session Isolation & Execution State in Same Workspace

## Problem Summary
When a user works across multiple sessions in the same workspace:
1. **Issue 1 (Chat Duplication/Cross-Bleed)**: Switching to a session in the same workspace causes the other session's chat to appear, or both show identical history.
   - **Root Cause A**: In `pkg/grokrunner/session_scanner.go:ResolveSessionFolder`, if a session has a frontend ID (e.g. `sess_...`) or an unrecognized ID, it falls back to picking the **most recently modified session folder** in that workspace! If session A modified the folder, session B's load history calls `GetSessionUsage` or `LoadGrokSessionHistory` with `sess_...` and automatically resolves to session A's folder, overwriting session B's memory with session A's messages.
   - **Root Cause B**: In `frontend/src/lib/stores/session.svelte.ts:openSessionInTab`, `if (target.messages.length === 0 || target.grokSessionId)` executes `loadSessionHistoryFromDisk(target)` every time! Because `target.grokSessionId` was falsely assigned to session A via fallback, opening session B loads session A's transcript from disk into session B.
2. **Issue 2 (Prompt Disappears & State Stuck in "Running" when switching sessions)**:
   - **Root Cause A**: In `pkg/grokrunner/runner.go`, when `StartSession` runs with a frontend ID (`sess_...`) and no `grokSessionId` yet, `ResolveSessionFolder` resolves to the most recent session folder (session A) instead of starting a new session or reserving a new UUID! Then `args = append(args, "--resume", resolvedID)` is executed instead of creating a brand new isolated Grok session via `--session-id <new-uuid>`.
   - **Root Cause B**: In `frontend/src/lib/stores/session.svelte.ts:createNewSessionModel`, new sessions are created with temporary IDs `sess_<random>_<timestamp>`. Because this is not a UUID, `isUUID` fails in Go, triggering the ambiguous folder search fallback.
   - **Root Cause C**: In `frontend/src/App.svelte:executeTurn`, if user switches tabs immediately after prompt submission, `sessionStore.activeSessionId` changes. Background event listeners (`grok:delta_batch`, `grok:complete`, etc.) correctly check `event.sessionId`, BUT `openSessionInTab` upon tab switch invokes `loadSessionHistoryFromDisk` while `target.status === 'working'`, wiping in-flight messages (`userMsg`) before Grok CLI finishes writing `chat_history.jsonl` to disk!
   - **Root Cause D**: If a session is currently `'working'`, switching to it must never overwrite its in-memory messages with stale disk content.

---

## Proposed Changes

### 1. Backend Hardening (`pkg/grokrunner/`)
- In `pkg/grokrunner/session_scanner.go`:
  - Modify `ResolveSessionFolder`:
    - Strict matching: Only resolve to an existing folder if the ID matches exactly, or matches case-insensitively, or is a known UUID prefix.
    - **Never** fall back to "most recently modified folder" when given an arbitrary `sess_...` ID. If not found, return empty resolved ID or return the requested ID without binding it to another session's folder.
  - In `pkg/grokrunner/runner.go`:
    - When `targetGrokID` is empty or is a temporary frontend ID (starts with `sess_`), generate a fresh UUID using `crypto/rand` (or Go standard UUID generation).
    - Pass `--session-id <new-uuid>` to Grok CLI so it creates an isolated session directory immediately.
    - Emit the new UUID back to the frontend in the first event or in `StartSession` response so the session is cleanly bound.

### 2. Frontend Store Hardening (`frontend/src/lib/stores/session.svelte.ts`)
- In `createNewSessionModel`:
  - Generate standard RFC4122 v4 UUIDs for all new sessions (`crypto.randomUUID()` available in modern webview and browsers) so every session has a guaranteed unique UUID from inception.
- In `openSessionInTab` & `switchSession`:
  - **Never** call `loadSessionHistoryFromDisk` if `target.status === 'working'` or if `target.messages.length > 0`.
  - Only load history from disk if `target.messages.length === 0` AND `target.status !== 'working'`.
  - When loading from disk, verify that `history` matches the target session ID to prevent cross-session contamination.
- In `syncDiscoveredGrokSessions`:
  - Never map one discovered disk session to multiple local session tabs.
- In `App.svelte`:
  - Ensure `executeTurn` locks the prompt and keeps `userMsg` safely in session state regardless of active tab switching.

---

## Verification Plan
1. Unit tests:
   - Add Go test in `pkg/grokrunner/session_scanner_test.go` verifying that `ResolveSessionFolder` never borrows an existing session folder for an unrelated or temporary session ID.
   - Add Go test verifying that `StartSession` assigns a unique session ID and never resumes another session's directory.
2. Build verification:
   - Run `go test ./...`
   - Run `cd frontend && npm run build`
3. Scenario simulation test:
   - Create Session 1 in workspace W, prompt "test 1".
   - Create Session 2 in workspace W, switch between Session 1 and Session 2.
   - Confirm Session 2 has separate chat history and prompt in Session 1 is never lost or overwritten.
