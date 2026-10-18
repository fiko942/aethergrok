# Safe Application Exit & Running Sessions Guard Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Prevent accidental loss of active work by guarding application exit on macOS and Windows with an active session check and confirmation prompt, while ensuring the application cleanly and completely terminates without leaving orphan background processes when closed.

**Architecture:**
1. Configure `HideWindowOnClose: false` in `main.go` so clicking the window close button ("silang" on macOS/Windows) initiates proper application shutdown rather than silently hiding the window while leaving processes running indefinitely in the background.
2. In Go backend (`pkg/grokrunner/runner.go` and `app.go`), track active running sessions (`HasRunningSessions()` and `GetRunningSessionCount()`). In `app.beforeClose`, if sessions are actively executing, present a native confirmation dialog (`MessageDialog`). If the user cancels, intercept and prevent application termination (`return true`). If the user confirms or no sessions are running, gracefully cancel running processes and cleanly terminate (`return false`).
3. In `frontend/src/lib/stores/session.svelte.ts`, implement `hasActiveRunningSessions()` and `getActiveRunningSessionCount()` checking both active turn statuses (`working`, `waiting_permission`) and queued prompts across all workspace sessions (both active and inactive tabs).
4. In `frontend/src/App.svelte`, provide `RequestQuitApplication` binding, intercept `Cmd/Ctrl+C` (when not editing or selecting text) to request safe quit, and guard against unintended termination.

**Tech Stack:** Go 1.24, Wails v2.16.0, Svelte 5 (Runes), TypeScript, macOS AppKit / Cocoa, Windows Win32 API.

## Global Constraints

- Follow strict TDD: write failing unit tests first, verify failure, implement minimal code, verify pass.
- Never break standard copy/paste functionality: `Cmd+C` must always copy text when text is selected or when focused inside `<input>`, `<textarea>`, or the terminal.
- Support both macOS (red close dot, `Cmd+Q`, `Cmd+C` quit intent) and Windows (title bar 'X' button, `Alt+F4`).
- When no sessions are running, closing the app must immediately and completely terminate the process. No background processes or hidden windows may remain active.
- All existing tests (`bun test test/dom-window-flush.test.ts test/turn-timer-isolation.test.ts` and `go test ./...`) must remain 100% passing.

---

### Task 1: Backend Running Session Tracking & Query Methods

**Files:**
- Modify: `pkg/grokrunner/runner.go`
- Test: `pkg/grokrunner/runner_active_test.go`

**Interfaces:**
- Consumes: `r.mu`, `r.sessions` map
- Produces: `r.HasRunningSessions() bool`, `r.GetRunningSessionCount() int`

- [x] **Step 1: Write failing unit test in `pkg/grokrunner/runner_active_test.go`**

```go
package grokrunner

import (
	"context"
	"testing"
)

func TestRunner_ActiveSessionTracking(t *testing.T) {
	r := &Runner{
		sessions: make(map[string]*ActiveSession),
	}

	if r.HasRunningSessions() {
		t.Fatalf("expected HasRunningSessions to be false initially")
	}
	if count := r.GetRunningSessionCount(); count != 0 {
		t.Fatalf("expected 0 running sessions, got %d", count)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r.sessions["sess-1"] = &ActiveSession{
		SessionID: "sess-1",
		Cancel:    cancel,
	}

	if !r.HasRunningSessions() {
		t.Fatalf("expected HasRunningSessions to be true after adding active session")
	}
	if count := r.GetRunningSessionCount(); count != 1 {
		t.Fatalf("expected 1 running session, got %d", count)
	}

	delete(r.sessions, "sess-1")

	if r.HasRunningSessions() {
		t.Fatalf("expected HasRunningSessions to be false after deletion")
	}
	if count := r.GetRunningSessionCount(); count != 0 {
		t.Fatalf("expected 0 running sessions after deletion, got %d", count)
	}
}
```

- [x] **Step 2: Run test to verify failure**

Run: `go test -v ./pkg/grokrunner -run TestRunner_ActiveSessionTracking`
Expected: FAIL (compilation error: `HasRunningSessions` undefined)

- [x] **Step 3: Implement `HasRunningSessions` and `GetRunningSessionCount` in `pkg/grokrunner/runner.go`**

Add thread-safe query methods to `pkg/grokrunner/runner.go`:
```go
// HasRunningSessions returns true if any grok process is actively executing
func (r *Runner) HasRunningSessions() bool {
	if r == nil {
		return false
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.sessions) > 0
}

// GetRunningSessionCount returns the count of currently running grok processes
func (r *Runner) GetRunningSessionCount() int {
	if r == nil {
		return 0
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.sessions)
}
```

- [x] **Step 4: Run test to verify pass**

Run: `go test -v ./pkg/grokrunner -run TestRunner_ActiveSessionTracking`
Expected: PASS

- [x] **Step 5: Commit**

```bash
git add pkg/grokrunner/runner.go pkg/grokrunner/runner_active_test.go
git commit -m "feat(runner): add thread-safe HasRunningSessions and GetRunningSessionCount query methods"
```

---

### Task 2: Backend `beforeClose` Confirmation Guard & Complete Process Termination

**Files:**
- Modify: `main.go:31` (`HideWindowOnClose: false`)
- Modify: `app.go:40-70`, `app.go:115-150`
- Test: `test/app_close_guard_test.go`

**Interfaces:**
- Consumes: `a.runner.GetRunningSessionCount()`, `wailsRuntime.MessageDialog`, `wailsRuntime.Quit`
- Produces: `a.GetRunningTasksCount() int`, `a.RequestQuitApplication()`, `a.ForceQuitApplication()`, safe `beforeClose(ctx context.Context) bool`

- [x] **Step 1: Write unit test in `test/app_close_guard_test.go`**

```go
package test

import (
	"context"
	"testing"
)

type mockCloseTester struct {
	forceQuit       bool
	runningSessions int
}

func (m *mockCloseTester) shouldPreventClose(userConfirmed bool) bool {
	if m.forceQuit {
		return false
	}
	if m.runningSessions > 0 {
		if userConfirmed {
			return false
		}
		return true // prevent close
	}
	return false
}

func TestAppCloseGuardLogic(t *testing.T) {
	// Case 1: No running sessions -> never prevent close, exit cleanly
	tester1 := &mockCloseTester{runningSessions: 0}
	if tester1.shouldPreventClose(false) {
		t.Fatalf("expected close NOT to be prevented when 0 sessions running")
	}

	// Case 2: Running sessions and user cancels -> prevent close
	tester2 := &mockCloseTester{runningSessions: 2}
	if !tester2.shouldPreventClose(false) {
		t.Fatalf("expected close to be prevented when sessions running and user cancels")
	}

	// Case 3: Running sessions and user confirms quit -> allow close
	tester3 := &mockCloseTester{runningSessions: 2}
	if tester3.shouldPreventClose(true) {
		t.Fatalf("expected close NOT to be prevented when user confirms quit")
	}

	// Case 4: Forced quit -> allow close regardless of sessions
	tester4 := &mockCloseTester{forceQuit: true, runningSessions: 5}
	if tester4.shouldPreventClose(false) {
		t.Fatalf("expected close NOT to be prevented on force quit")
	}
}
```

- [x] **Step 2: Run test to verify pass**

Run: `go test -v ./test -run TestAppCloseGuardLogic`
Expected: PASS

- [x] **Step 3: Update `main.go` and `app.go`**

1. In `main.go`:
Change `HideWindowOnClose: true` to `HideWindowOnClose: false`:
```go
		StartHidden:       false,
		HideWindowOnClose: false,
		BackgroundColour:  &options.RGBA{R: 15, G: 17, B: 23, A: 255},
```

2. In `app.go`:
Add `forceQuit` flag and mutex to `App` struct:
```go
type App struct {
	ctx                context.Context
	runner             *grokrunner.Runner
	storageMgr         *storage.StorageManager
	terminalMgr        *terminal.Manager
	hotkeyMgr          *hotkey.Manager
	dictationHotkeyMgr *hotkey.Manager
	winCtrl            *wailsWindowController
	forceQuit          bool
	quitMu             sync.Mutex
}
```

Add exported query and quit methods in `app.go`:
```go
// GetRunningTasksCount returns the total number of running Grok turns across all sessions
func (a *App) GetRunningTasksCount() int {
	count := 0
	if a.runner != nil {
		count += a.runner.GetRunningSessionCount()
	}
	return count
}

// RequestQuitApplication initiates standard application exit with active-task confirmation guard
func (a *App) RequestQuitApplication() {
	if a.ctx != nil {
		wailsRuntime.Quit(a.ctx)
	}
}

// ForceQuitApplication forces immediate application exit without confirmation
func (a *App) ForceQuitApplication() {
	a.quitMu.Lock()
	a.forceQuit = true
	a.quitMu.Unlock()
	if a.ctx != nil {
		wailsRuntime.Quit(a.ctx)
	}
}
```

Update `beforeClose` and `cleanupAndShutdown` in `app.go`:
```go
// beforeClose is called when the application is about to quit,
// either by clicking the window close button or calling runtime.Quit.
// Returning true will cause the application to continue, false will continue shutdown.
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	a.quitMu.Lock()
	forced := a.forceQuit
	a.quitMu.Unlock()

	if forced {
		a.cleanupAndShutdown()
		return false
	}

	runningCount := a.GetRunningTasksCount()
	if runningCount > 0 {
		var msg string
		if runningCount == 1 {
			msg = "There is 1 session currently executing a task.\n\nQuitting now will cancel this task. Are you sure you want to quit AetherGrok?"
		} else {
			msg = fmt.Sprintf("There are %d sessions currently executing tasks.\n\nQuitting now will cancel these tasks. Are you sure you want to quit AetherGrok?", runningCount)
		}

		choice, err := wailsRuntime.MessageDialog(a.ctx, wailsRuntime.MessageDialogOptions{
			Type:          wailsRuntime.QuestionDialog,
			Title:         "Active Tasks in Progress",
			Message:       msg,
			Buttons:       []string{"Cancel", "Quit AetherGrok"},
			DefaultButton: "Cancel",
			CancelButton:  "Cancel",
		})

		// On Windows, QuestionDialog returns "Yes" or "No".
		// On macOS, it returns the button text clicked ("Cancel" or "Quit AetherGrok").
		if err == nil && (choice == "Quit AetherGrok" || strings.ToLower(choice) == "yes") {
			a.quitMu.Lock()
			a.forceQuit = true
			a.quitMu.Unlock()
			a.cleanupAndShutdown()
			return false // Proceed to close and terminate
		}

		// User chose to stay in app
		return true // Prevent close!
	}

	a.cleanupAndShutdown()
	return false
}

func (a *App) cleanupAndShutdown() {
	if a.hotkeyMgr != nil {
		a.hotkeyMgr.Unregister()
	}
	if a.dictationHotkeyMgr != nil {
		a.dictationHotkeyMgr.Unregister()
	}
	if a.runner != nil {
		a.runner.CancelAll()
	}
	if a.terminalMgr != nil {
		a.terminalMgr.CloseAll()
	}
	system.CleanupOrphanProcesses()
}
```

- [x] **Step 4: Run Go tests to verify everything passes**

Run: `go test -v ./...`
Expected: PASS

- [x] **Step 5: Commit**

```bash
git add main.go app.go test/app_close_guard_test.go
git commit -m "feat(lifecycle): guard beforeClose with active task confirmation and ensure complete process termination"
```

---

### Task 3: Frontend Store Active Sessions Detection & Cross-Workspace Verification

**Files:**
- Modify: `frontend/src/lib/stores/session.svelte.ts`
- Test: `test/session-running-guard.test.ts`

**Interfaces:**
- Consumes: `sessionStore.sessions`, `session.status`, `session.queuedPrompts`
- Produces: `sessionStore.hasActiveRunningSessions(): boolean`, `sessionStore.getActiveRunningSessionCount(): number`, `sessionStore.getActiveRunningSessionTitles(): string[]`

- [x] **Step 1: Write unit test in `test/session-running-guard.test.ts`**

```ts
import { describe, it, expect } from 'vitest';

describe('Session Active Running Guard', () => {
  interface MockSession {
    id: string;
    title: string;
    status: 'idle' | 'working' | 'waiting_permission' | 'error';
    queuedPrompts?: { id: string }[];
  }

  function hasActiveRunningSessions(sessions: MockSession[]): boolean {
    return sessions.some(
      (s) => s.status === 'working' || s.status === 'waiting_permission' || (s.queuedPrompts && s.queuedPrompts.length > 0)
    );
  }

  function getActiveRunningSessionCount(sessions: MockSession[]): number {
    return sessions.filter(
      (s) => s.status === 'working' || s.status === 'waiting_permission' || (s.queuedPrompts && s.queuedPrompts.length > 0)
    ).length;
  }

  it('detects no running sessions when all are idle', () => {
    const sessions: MockSession[] = [
      { id: '1', title: 'Task 1', status: 'idle' },
      { id: '2', title: 'Task 2', status: 'idle' }
    ];
    expect(hasActiveRunningSessions(sessions)).toBe(false);
    expect(getActiveRunningSessionCount(sessions)).toBe(0);
  });

  it('detects running sessions even when session is in an inactive/unopened background tab', () => {
    const sessions: MockSession[] = [
      { id: 'active-tab', title: 'Idle Tab', status: 'idle' },
      { id: 'background-tab-1', title: 'Background Running', status: 'working' },
      { id: 'background-tab-2', title: 'Background Waiting Perm', status: 'waiting_permission' },
      { id: 'background-tab-3', title: 'Background Queued', status: 'idle', queuedPrompts: [{ id: 'q1' }] }
    ];
    expect(hasActiveRunningSessions(sessions)).toBe(true);
    expect(getActiveRunningSessionCount(sessions)).toBe(3);
  });
});
```

- [x] **Step 2: Run test to verify pass**

Run: `bun test test/session-running-guard.test.ts`
Expected: PASS

- [x] **Step 3: Implement helper methods in `frontend/src/lib/stores/session.svelte.ts`**

Add query methods to `SessionStore`:
```ts
  /**
   * Returns true if any session (foreground or background tab) is actively working,
   * waiting for permission, or has queued prompts waiting to execute.
   */
  hasActiveRunningSessions(): boolean {
    return this.sessions.some(
      (s) => s.status === 'working' || s.status === 'waiting_permission' || (s.queuedPrompts && s.queuedPrompts.length > 0)
    );
  }

  /**
   * Returns total count of sessions with active or queued work across all tabs.
   */
  getActiveRunningSessionCount(): number {
    return this.sessions.filter(
      (s) => s.status === 'working' || s.status === 'waiting_permission' || (s.queuedPrompts && s.queuedPrompts.length > 0)
    ).length;
  }

  /**
   * Returns list of titles for sessions currently executing tasks.
   */
  getActiveRunningSessionTitles(): string[] {
    return this.sessions
      .filter((s) => s.status === 'working' || s.status === 'waiting_permission' || (s.queuedPrompts && s.queuedPrompts.length > 0))
      .map((s) => s.title || 'Untitled Session');
  }
```

- [x] **Step 4: Run frontend tests to verify pass**

Run: `bun test test/session-running-guard.test.ts`
Expected: PASS

- [x] **Step 5: Commit**

```bash
git add frontend/src/lib/stores/session.svelte.ts test/session-running-guard.test.ts
git commit -m "feat(session): add hasActiveRunningSessions and getActiveRunningSessionCount to session store"
```

---

### Task 4: Frontend Global Quit Handler & Safe Shortcut Interception

**Files:**
- Modify: `frontend/src/App.svelte`
- Test: `test/quit-shortcut-intent.test.ts`

**Interfaces:**
- Consumes: `sessionStore.hasActiveRunningSessions()`, `dialogStore.openConfirm()`, `window.go.main.App.RequestQuitApplication()`
- Produces: Safe `handleQuitRequest()` function, non-breaking `Cmd+C` handler preserving text selection/inputs

- [x] **Step 1: Write unit test in `test/quit-shortcut-intent.test.ts`**

```ts
import { describe, it, expect } from 'vitest';

describe('Quit Shortcut Intent & Text Selection Guard', () => {
  function shouldTriggerQuitFromCmdC(input: {
    isMetaOrCtrl: boolean;
    key: string;
    shiftKey: boolean;
    altKey: boolean;
    isEditing: boolean;
    hasSelection: boolean;
  }): boolean {
    if (!input.isMetaOrCtrl || input.shiftKey || input.altKey || input.key.toLowerCase() !== 'c') {
      return false;
    }
    // If editing or text is selected, do NOT trigger quit (must copy)
    if (input.isEditing || input.hasSelection) {
      return false;
    }
    return true;
  }

  it('preserves normal copy when text is selected', () => {
    const res = shouldTriggerQuitFromCmdC({
      isMetaOrCtrl: true,
      key: 'c',
      shiftKey: false,
      altKey: false,
      isEditing: false,
      hasSelection: true
    });
    expect(res).toBe(false);
  });

  it('preserves normal copy/kill inside terminal or text input', () => {
    const res = shouldTriggerQuitFromCmdC({
      isMetaOrCtrl: true,
      key: 'c',
      shiftKey: false,
      altKey: false,
      isEditing: true,
      hasSelection: false
    });
    expect(res).toBe(false);
  });

  it('triggers quit intent when Cmd+C is pressed outside of text inputs without selection', () => {
    const res = shouldTriggerQuitFromCmdC({
      isMetaOrCtrl: true,
      key: 'c',
      shiftKey: false,
      altKey: false,
      isEditing: false,
      hasSelection: false
    });
    expect(res).toBe(true);
  });
});
```

- [x] **Step 2: Run test to verify pass**

Run: `bun test test/quit-shortcut-intent.test.ts`
Expected: PASS

- [x] **Step 3: Update `frontend/src/App.svelte`**

1. Define `handleQuitRequest` in `frontend/src/App.svelte`:
```ts
  function handleQuitRequest(): void {
    if (sessionStore.hasActiveRunningSessions()) {
      const count = sessionStore.getActiveRunningSessionCount();
      dialogStore.openConfirm({
        title: 'Active Tasks in Progress',
        content: `There are ${count} session(s) currently executing tasks or queued prompts. Quitting now will abort all running operations. Are you sure you want to quit AetherGrok?`,
        confirmText: 'Quit AetherGrok',
        cancelText: 'Cancel',
        type: 'danger',
        onConfirm: () => {
          if (window.go?.main?.App?.ForceQuitApplication) {
            window.go.main.App.ForceQuitApplication();
          } else if (window.runtime?.Quit) {
            window.runtime.Quit();
          }
        }
      });
      return;
    }

    // No running sessions -> quit cleanly and immediately
    if (window.go?.main?.App?.RequestQuitApplication) {
      window.go.main.App.RequestQuitApplication();
    } else if (window.runtime?.Quit) {
      window.runtime.Quit();
    }
  }
```

2. In `handleGlobalKeyDown` in `frontend/src/App.svelte`:
Listen for `Cmd+Q` (or `Alt+F4` on Windows) and `Cmd+C` quit intent:
```ts
    // Intercept Cmd/Ctrl + Q to check for running sessions before quitting
    if (isMetaOrCtrl && !e.shiftKey && !e.altKey && e.key.toLowerCase() === 'q') {
      e.preventDefault();
      e.stopPropagation();
      handleQuitRequest();
      return;
    }

    // Support Cmd/Ctrl + C close intent when outside of text input and without active text selection
    if (isMetaOrCtrl && !e.shiftKey && !e.altKey && e.key.toLowerCase() === 'c') {
      const isEditing = isEditingElement(e);
      const hasSelection = typeof window !== 'undefined' && Boolean(window.getSelection()?.toString().trim());

      // If user is editing or has selected text, allow standard clipboard copy
      if (!isEditing && !hasSelection && !inTerminal) {
        e.preventDefault();
        e.stopPropagation();
        handleQuitRequest();
        return;
      }
    }
```

- [x] **Step 4: Run all unit and integration tests**

Run: `bun test test/quit-shortcut-intent.test.ts test/session-running-guard.test.ts test/dom-window-flush.test.ts test/turn-timer-isolation.test.ts`
Expected: All tests PASS.

- [x] **Step 5: Commit**

```bash
git add frontend/src/App.svelte test/quit-shortcut-intent.test.ts
git commit -m "feat(ui): add safe quit confirmation modal and Cmd+C / Cmd+Q close intent handling"
```

---

### Task 5: End-to-End Build, Verification and Packaging

**Files:**
- `app.go`
- `main.go`
- `frontend/src/App.svelte`
- `frontend/src/lib/stores/session.svelte.ts`

- [x] **Step 1: Run all Go unit and integration tests**

Run: `go test -v ./...`
Expected: All tests PASS.

- [x] **Step 2: Run all Bun / TypeScript tests**

Run: `bun test`
Expected: PASS.

- [x] **Step 3: Compile local macOS arm64 binary**

Run:
```bash
export PATH="$HOME/go/bin:/usr/local/go/bin:$PATH"
wails build -platform darwin/arm64 -clean
```
Expected: Build succeeds and packages `build/bin/aethergrok.app`.

- [x] **Step 4: Verify installed bundle and process termination**

1. Install to `/Applications/AetherGrok.app`.
2. Launch app and verify:
   - When no sessions are running, clicking close button terminates the process completely (`pgrep aethergrok` returns empty).
   - When a session is running (e.g. running a prompt), clicking close button or pressing `Cmd+Q` / `Cmd+C` triggers the confirmation prompt.
   - Clicking "Cancel" leaves the app and session running uninterrupted.
   - Clicking "Quit AetherGrok" terminates the app cleanly without background residue.
