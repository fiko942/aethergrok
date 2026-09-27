# Implementation Plan: Session Persistence & Multi-Turn Continuation

## 1. Problem Analysis & Root Cause
There are two core issues reported by the user:

### Issue 1: New session created on every prompt turn (Spamming ~/.grok/sessions and losing context)
- **Root Cause**:
  1. In `pkg/grokrunner/runner.go`, when `StartSession` is called for subsequent user messages, it executes `grok -p <prompt>`.
  2. Because neither `--resume <grokSessionId>` nor `--session-id <uuid>` was passed to `grok`, the CLI treated every single message as a brand new standalone conversation, writing a new UUID directory under `~/.grok/sessions/<encoded_workspace>/` every time.
  3. Consequently, context did not accumulate across turns, token context remained tiny (~28k system baseline instead of increasing with chat history), and the disk was spammed with dozens of single-turn session directories.

### Issue 2: Messages appearing cut off / disappearing when reopening sessions
- **Root Cause**:
  1. `Session.id` in the frontend was generated with custom strings like `sess_abc123_xyz`.
  2. When a session in AetherGrok already completed turns and received a real Grok UUID (or when opening an existing Grok session), `sessionStore.sessions` wasn't linking and updating `session.grokSessionId` consistently on first turn completion or when resuming.
  3. In `session_scanner.go` -> `LoadGrokSessionMessages`, if `role == "assistant"` messages contained multiple responses or interleaved reasoning / tool calls, assistant text and user queries weren't accurately preserved in sequence.
  4. In `session.svelte.ts`, when saving to `localStorage`, sessions were overwritten or stripped when `syncDiscoveredGrokSessions` re-indexed sessions. Furthermore, `openSessionInTab` only loaded history if `messages.length === 0`, but if a session had empty/partial messages saved in local storage, it never re-synced with the source on disk.

---

## 2. Architecture & Design Fixes

### A. Session ID Lifecycle & Resume in `runner.go`
1. For any session:
   - Check if `req.SessionID` (or `req.Options.GrokSessionID`) corresponds to an existing Grok session on disk or if it is an existing session turn (`--resume <id>`).
   - If a session is new (first turn) and has a valid UUID (or we assign a clean UUID), pass `--session-id <uuid>`.
   - If a session already has a Grok session ID / existing turns, pass `--resume <grokSessionId>` so Grok continues in the EXACT same conversation!
2. When Grok streams the `end` event with `sessionId`, emit `TurnCompleteEvent` containing the authoritative `GrokSessionID`.
3. Frontend updates `session.grokSessionId` (and if appropriate, adopts the UUID) so all subsequent turns continue on that session seamlessly.

### B. Accurate Session Scanner & History Loader (`session_scanner.go`)
1. Refine `LoadGrokSessionMessages`:
   - Parse both `chat_history.jsonl` and `updates.jsonl` cleanly.
   - Accurately extract all user queries (inside `<user_query>` or raw user prompt), assistant markdown responses, reasoning content, and tool calls in chronological order.
   - Preserve complete turn sequence without truncating bottom messages.
2. In `session.svelte.ts`:
   - Ensure `loadSessionHistoryFromDisk` merges and updates `session.messages` accurately.
   - Ensure `session.grokSessionId` is passed in `RunPromptStream` calls.

---

## 3. Step-by-Step Implementation Tasks
1. Update `pkg/grokrunner/types.go` and `pkg/grokrunner/runner.go` to handle `--resume` vs `--session-id`.
2. Enhance `pkg/grokrunner/session_scanner.go` to parse all turns, tool calls, and assistant content from disk.
3. Update `frontend/src/App.svelte` and `frontend/src/lib/stores/session.svelte.ts` to pass `grokSessionId` to `RunPromptStream` and properly sync disk messages.
4. Write comprehensive tests in `pkg/grokrunner/runner_test.go` and `pkg/grokrunner/session_scanner_test.go`.
5. Run full test suite and verify build.
