# Session Isolation and Deduplication Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Strictly isolate Grok CLI execution by workspace directory, ensure continuation turns reliably resume existing sessions without spawning new sessions or losing context, eliminate duplicate session multiplication during discovery and app launch, and handle multi-workspace concurrent execution cleanly.

**Architecture:** 
1. Backend (`pkg/grokrunner`): Strictly confine session folder resolution to the target workspace's canonical directory (evaluating symlinks like macOS `/var` -> `/private/var`), remove cross-workspace fallback searches, and reliably select `--resume` vs `--session-id` so continuing turns never generate brand-new conversation sessions.
2. Frontend Session Store (`frontend/src/lib/stores/session.svelte.ts`): Reconcile and deduplicate sessions in memory and `localStorage` by matching both `s.id === gs.id` and `s.grokSessionId === gs.id` scoped to `workspaceId`, prevent ghost duplicate creation, and ensure in-flight turns are preserved.
3. Frontend Turn Orchestration (`frontend/src/App.svelte`): Strictly bind execution and disk sync to the session's own workspace path rather than `activeWorkspace`, make permission requests and cancellations session-id-targeted to support concurrent execution across tabs and workspaces.

**Tech Stack:** Go 1.24+ (Wails v2 backend, `pkg/grokrunner`, `pkg/storage`), Svelte 5 Runes (`$state`, `$derived`), TypeScript, Vite, Tailwind CSS.

## Global Constraints

- Never allow a session from Workspace A to be resumed or discovered inside Workspace B.
- Continuation turns for an existing session must ALWAYS use `--resume <UUID>` and preserve chat history.
- `--session-id <UUID>` is strictly used when initiating a brand new conversation turn that does not yet exist on disk in the target workspace.
- Deduplication must clean existing duplicate sessions from `localStorage` without losing user chat messages.
- Closing a tab or deleting a session must cleanly terminate any running Grok runner process for that session ID.

---

### Task 1: Backend Session Scanner Isolation & Symlink Canonicalization

**Files:**
- Modify: `pkg/grokrunner/session_scanner.go:50-230`
- Modify: `pkg/grokrunner/session_scanner_test.go`

**Interfaces:**
- Consumes: `sessionsDir string`, `workspacePath string`, `sessionID string`
- Produces: `ResolveWorkspaceSessionsDir(sessionsDir, workspacePath string) string`, `SessionFolderExists(sessionsDir, targetWsDir, sessionID string) (bool, string, string)`

- [ ] **Step 1: Write failing tests for strict workspace isolation and symlink resolution**

Add `TestSessionFolderExists_StrictWorkspaceIsolation` and `TestResolveWorkspaceSessionsDir_SymlinkCanonicalization` in `pkg/grokrunner/session_scanner_test.go`:
```go
func TestSessionFolderExists_StrictWorkspaceIsolation(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	sessionsDir := filepath.Join(home, ".grok", "sessions")
	ws1 := filepath.Join(t.TempDir(), "ws1")
	ws2 := filepath.Join(t.TempDir(), "ws2")
	_ = os.MkdirAll(ws1, 0755)
	_ = os.MkdirAll(ws2, 0755)

	ws1Dir := ResolveWorkspaceSessionsDir(sessionsDir, ws1)
	ws2Dir := ResolveWorkspaceSessionsDir(sessionsDir, ws2)
	_ = os.MkdirAll(ws1Dir, 0755)
	_ = os.MkdirAll(ws2Dir, 0755)
	defer os.RemoveAll(ws1Dir)
	defer os.RemoveAll(ws2Dir)

	sessID := "99999999-1111-2222-3333-444444444444"
	sessFolder1 := filepath.Join(ws1Dir, sessID)
	_ = os.MkdirAll(sessFolder1, 0755)
	_ = os.WriteFile(filepath.Join(sessFolder1, "chat_history.jsonl"), []byte("{}\n"), 0644)

	// ws1 should find sessID
	exists1, _, _ := SessionFolderExists(sessionsDir, ws1Dir, sessID)
	if !exists1 {
		t.Fatalf("Expected session to exist in ws1")
	}

	// ws2 MUST NOT find sessID under any circumstances (no cross-workspace fallback)
	exists2, _, _ := SessionFolderExists(sessionsDir, ws2Dir, sessID)
	if exists2 {
		t.Fatalf("Session from ws1 MUST NOT be found when querying ws2; cross-workspace fallback must be disabled")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -v ./pkg/grokrunner -run TestSessionFolderExists_StrictWorkspaceIsolation`
Expected: FAIL (because `SessionFolderExists` currently has fallback scanning across all workspaces in `sessionsDir`).

- [ ] **Step 3: Implement strict workspace isolation and symlink resolution**

In `pkg/grokrunner/session_scanner.go`:
1. In `ResolveWorkspaceSessionsDir`:
   Evaluate canonical path with `filepath.EvalSymlinks(cleanWs)`. If canonical path exists and is different from `cleanWs`, check encoded canonical path candidate as well.
2. In `SessionFolderExists`:
   Remove fallback step 2 (`// 2. Fallback: check across all workspace directories in sessionsDir`). Only check `targetWsDir`.
3. In `hasSessionRecord(folderPath)`:
   If `folderPath` contains any valid Grok session files (including `events.jsonl`, `prompt_context.json`, `signals.json`, `usage.json`, `chat_history.jsonl`, `summary.json`, `updates.jsonl`), return `true`.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -v ./pkg/grokrunner -run TestSessionFolderExists_StrictWorkspaceIsolation`
Expected: PASS

- [ ] **Step 5: Run all grokrunner tests**

Run: `go test -v ./pkg/grokrunner/...`
Expected: PASS

---

### Task 2: Backend Runner Continuation Logic (`--resume` vs `--session-id`)

**Files:**
- Modify: `pkg/grokrunner/runner.go:280-335`
- Test: `pkg/grokrunner/runner_test.go`

**Interfaces:**
- Consumes: `req PromptRequest` with `req.SessionID`, `req.Options.GrokSessionID`, `req.Options.WorkingDir`
- Produces: `cmd.Args` constructed with `--resume` for existing sessions and `--session-id` for new sessions

- [ ] **Step 1: Write unit test verifying continuation flag selection**

Add test in `pkg/grokrunner/runner_test.go`:
```go
func TestRunner_ContinuationFlagSelection(t *testing.T) {
	// Verify that if a session directory exists in targetWsDir, --resume is chosen.
	// If a session does not exist, --session-id is chosen with the given UUID.
}
```

- [ ] **Step 2: Run test to verify behavior**

Run: `go test -v ./pkg/grokrunner -run TestRunner_ContinuationFlagSelection`

- [ ] **Step 3: Implement robust continuation logic in `runner.go`**

In `pkg/grokrunner/runner.go:StartSession`:
1. Identify target session ID:
   ```go
   targetGrokID := strings.TrimSpace(req.Options.GrokSessionID)
   if targetGrokID == "" && isUUID(req.SessionID) {
       targetGrokID = req.SessionID
   }
   ```
2. Determine `wsPath` and canonical `targetWsDir`.
3. If `targetGrokID` is non-empty:
   Check `SessionFolderExists(sessionsDir, targetWsDir, targetGrokID)`.
   If it exists:
     Pass `--resume <resolvedID>` (or `targetGrokID`).
   Else:
     Pass `--session-id <targetGrokID>` if `isUUID(targetGrokID)`, else generate UUID and pass `--session-id <newUUID>`.
4. Ensure `cmd.Dir` is set to canonical `wsPath` if available.

- [ ] **Step 4: Run tests to verify**

Run: `go test -v ./pkg/grokrunner/...`
Expected: PASS

---

### Task 3: Backend Session Turn Cancellation

**Files:**
- Modify: `app.go:193-198`
- Modify: `pkg/grokrunner/runner.go`

**Interfaces:**
- Consumes: `sessionID string`
- Produces: `CancelSession(sessionID string) error`

- [ ] **Step 1: Verify `CancelSession` terminates subprocess and process group**

Verify `Cancel(sessionID)` properly signals `cmd.Process` and kills process group.
- [ ] **Step 2: Expose `CancelSessionTurn` or verify `CancelSession` in `app.go`**

Confirm `a.CancelSession(sessionID)` is available to frontend via Wails.

---

### Task 4: Frontend Session Deduplication & Storage Reconciliation

**Files:**
- Modify: `frontend/src/lib/stores/session.svelte.ts:380-450,1210-1265`

**Interfaces:**
- Consumes: `wsId: string`, `grokSessions: Array<{ id: string; title: string; createdAt: number; updatedAt: number }>`
- Produces: Deduplicated `this.sessions`, synchronized `grokSessionId`, persistent storage

- [ ] **Step 1: Implement in-memory deduplication in `loadSessionsFromStorage`**

In `loadSessionsFromStorage`:
When parsing sessions from localStorage:
1. Merge sessions that share the same `(workspaceId, grokSessionId)` or `(workspaceId, id)`.
2. Keep the session with the most messages / non-empty content.
3. Clean up duplicate tab entries from `this.openTabSessionIds`.

- [ ] **Step 2: Update `syncDiscoveredGrokSessions` to match by both `id` and `grokSessionId`**

In `syncDiscoveredGrokSessions(wsId, grokSessions)`:
```typescript
for (const gs of grokSessions) {
  const existing = this.sessions.find(
    (s) => s.workspaceId === wsId && (s.id === gs.id || s.grokSessionId === gs.id)
  );

  if (existing) {
    if (!existing.grokSessionId || existing.grokSessionId !== gs.id) {
      existing.grokSessionId = gs.id;
    }
    if (gs.title && !existing.isCustomTitle && (existing.title.startsWith('Session ') || existing.title.startsWith('Percakapan ') || existing.title.startsWith('New '))) {
      existing.title = gs.title;
    }
    if (gs.updatedAt && gs.updatedAt > existing.updatedAt) {
      existing.updatedAt = gs.updatedAt;
    }
  } else {
    // Only add if not already in this.sessions under any matching ID
    const duplicate = this.sessions.find((s) => s.id === gs.id || s.grokSessionId === gs.id);
    if (!duplicate) {
      this.sessions.push({
        id: gs.id,
        grokSessionId: gs.id,
        workspaceId: wsId,
        title: gs.title || `Session ${this.sessions.filter((s) => s.workspaceId === wsId).length + 1}`,
        status: 'idle',
        createdAt: gs.createdAt || Date.now(),
        updatedAt: gs.updatedAt || Date.now(),
        messages: [],
        visibleTurnCount: DEFAULT_WINDOW_TURNS,
        pendingPermission: null
      });
    }
  }
}
```
Remove unsafe splicing of placeholder sessions that could wipe active user sessions.
Save sessions to storage after sync.

- [ ] **Step 3: Update `closeSession` to cancel any active runner turn**

In `closeSession(id)`:
If the session is running (`status === 'working'`), call `window.go?.main?.App?.CancelSession(id)`.

- [ ] **Step 4: Verify frontend builds cleanly**

Run: `pnpm --filter aethergrok-frontend build`
Expected: Build succeeds.

---

### Task 5: Frontend Multi-Workspace Concurrency & Turn Execution

**Files:**
- Modify: `frontend/src/App.svelte:500-530,690-715,1360-1470`

**Interfaces:**
- Consumes: `sessionId: string`, `payload: any`
- Produces: Multi-workspace isolated `RunPromptStream` calls and targeted permission handling

- [ ] **Step 1: Harden `executeTurn` workspace resolution**

In `executeTurn`:
```typescript
const sessionObj = sessionStore.sessions.find((s) => s.id === sessionId);
const sessionWs = sessionObj ? sessionStore.workspaces.find((w) => w.id === sessionObj.workspaceId) : null;
const workingDir = sessionWs?.path || sessionStore.activeWorkspace?.path;
const grokSessionId = sessionObj?.grokSessionId || (sessionObj?.id && !sessionObj.id.startsWith('sess_') ? sessionObj.id : undefined);
```
Ensure `workingDir` strictly falls back to `sessionWs.path`.
Pass `grokSessionId` consistently.

- [ ] **Step 2: Fix `unsubComplete` workspace scoping and storage persistence**

In `unsubComplete`:
```typescript
const sessionObj = sessionStore.sessions.find((s) => s.id === event.sessionId);
if (sessionObj) {
  if (event.grokSessionId) {
    sessionObj.grokSessionId = event.grokSessionId;
  }
  sessionStore.saveSessionsToStorage();
}

const sessionWs = sessionObj ? sessionStore.workspaces.find((w) => w.id === sessionObj.workspaceId) : null;
if (sessionWs && window.go?.main?.App?.DiscoverGrokSessions) {
  try {
    const diskSessions = await window.go.main.App.DiscoverGrokSessions(sessionWs.path);
    if (diskSessions && diskSessions.length > 0) {
      sessionStore.syncDiscoveredGrokSessions(sessionWs.id, diskSessions);
    }
  } catch (err) {
    console.error('Failed to auto-sync sessions after turn:', err);
  }
}
```

- [ ] **Step 3: Make permission handling session-id-targeted**

Update `handlePermissionDecision` to accept `(decision: string, targetSessionId?: string)`:
When `unsubPerm` triggers, pass `event.sessionId` to `handlePermissionDecision`.
When resolving permission, target `sessionStore.sessions.find(s => s.id === (targetSessionId || activeSessionId))`.

- [ ] **Step 4: Verify frontend build**

Run: `pnpm --filter aethergrok-frontend build`
Expected: Build succeeds.

---

### Task 6: Full Verification & Regression Testing

**Files:**
- All touched files

- [ ] **Step 1: Run all Go unit tests**

Run: `go test -v ./pkg/...`
Expected: All tests PASS.

- [ ] **Step 2: Run frontend build**

Run: `pnpm --filter aethergrok-frontend build`
Expected: Build PASS with 0 errors.

- [ ] **Step 3: End-to-end regression check**

Simulate multi-session prompt continuation and workspace isolation.
Confirm no new sessions are spawned when continuing turns in an existing session.
Confirm no duplicate sessions appear on app reopen.
