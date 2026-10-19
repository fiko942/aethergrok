# Session Stop and Turn Cancellation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix turn stop/cancellation so that clicking the Stop button reliably and immediately terminates the running Grok CLI process group, resets session status to idle, finalizes active tool calls, and updates the UI without hanging.

**Architecture:** Isolate DOM click events in `Composer.svelte` so `PointerEvent` objects do not pollute `handleCancelSession`. Add string validation in `handleCancelSession` with fallback to `sessionStore.activeSession?.id`, and finalize in-progress tool calls. In Go backend `pkg/grokrunner/runner.go`, track `GrokSessionID` on `ActiveSession`, enhance `Cancel(sessionID)` with multi-identifier lookup (frontend session ID, Grok CLI UUID, or single-running fallback), and ensure both process group `SIGKILL` and process kill are dispatched.

**Tech Stack:** Go 1.23+ (`exec.Cmd`, `syscall.Kill`), Wails v2, Svelte 5 (`$state`, `$derived`), TypeScript, Vitest.

## Global Constraints
- Avoid contrastive negation framing in responses.
- Binary-free test suite: never spawn real Grok CLI binary during automated unit tests.
- Do not push to GitHub unless explicitly commanded by user.
- Preserve existing session isolation, workspace scoping, and active turn windowing.

---

### Task 1: Fix Frontend Stop Event Isolation and Session Cancellation

**Files:**
- Modify: `frontend/src/lib/components/chat/Composer.svelte:1526-1536`
- Modify: `frontend/src/App.svelte:668-692`
- Test: `test/session-stop-cancel.test.ts`

**Interfaces:**
- Consumes: `onCancel?: (sessionId?: string) => void` in `Composer.svelte`
- Produces: `handleCancelSession(targetSessionId?: unknown, pauseQueue?: boolean): Promise<void>` in `App.svelte`

- [ ] **Step 1: Write the failing test for cancellation event handling and session finalization**

```typescript
// test/session-stop-cancel.test.ts
import { describe, it, expect, vi } from 'vitest';

if (typeof (globalThis as any).$state === 'undefined') {
  (globalThis as any).$state = (v: any) => v;
}
if (typeof (globalThis as any).$derived === 'undefined') {
  (globalThis as any).$derived = (v: any) => v;
}

describe('Session Stop and Cancellation Logic', () => {
  it('correctly resolves session ID even when called with a DOM event object', () => {
    function resolveSessionId(targetSessionId: unknown, fallbackActiveId?: string | null): string | null {
      const valid = typeof targetSessionId === 'string' && targetSessionId.trim() ? targetSessionId.trim() : null;
      return valid || fallbackActiveId || null;
    }

    const mouseEvent = { isTrusted: true, type: 'click' };
    expect(resolveSessionId(mouseEvent, 'sess_active_123')).toBe('sess_active_123');
    expect(resolveSessionId('sess_explicit_456', 'sess_active_123')).toBe('sess_explicit_456');
    expect(resolveSessionId(undefined, 'sess_active_123')).toBe('sess_active_123');
    expect(resolveSessionId(null, null)).toBeNull();
  });

  it('finalizes in-progress running tool calls when cancelled', () => {
    const mockMessage = {
      role: 'assistant',
      content: 'Thinking...',
      status: 'streaming',
      toolCalls: [
        { id: 'tc-1', tool: 'spawn_subagent', status: 'running' as const, startTime: 1000 },
        { id: 'tc-2', tool: 'read_file', status: 'completed' as const, startTime: 500, endTime: 700 }
      ]
    };

    // Cancellation finalizer simulation
    if (mockMessage.status === 'streaming') {
      mockMessage.status = 'done';
      mockMessage.content += '\n\n*(Turn cancelled by user)*';
    }
    if (mockMessage.toolCalls) {
      for (const tc of mockMessage.toolCalls) {
        if (tc.status === 'running') {
          tc.status = 'error';
          tc.endTime = 2000;
        }
      }
    }

    expect(mockMessage.status).toBe('done');
    expect(mockMessage.content).toContain('*(Turn cancelled by user)*');
    expect(mockMessage.toolCalls[0].status).toBe('error');
    expect(mockMessage.toolCalls[0].endTime).toBe(2000);
    expect(mockMessage.toolCalls[1].status).toBe('completed');
  });
});
```

- [ ] **Step 2: Run test to verify it passes**

Run: `npx vitest run test/session-stop-cancel.test.ts`
Expected: PASS

- [ ] **Step 3: Update `Composer.svelte` to isolate the stop button click event**

In `frontend/src/lib/components/chat/Composer.svelte`:
```svelte
            <!-- Stop/Cancel Execution Button -->
            <button
              type="button"
              class="w-7 h-7 rounded-lg flex items-center justify-center bg-rose-600/90 text-white hover:bg-rose-500 transition shadow-sm cursor-pointer"
              onclick={(e) => {
                e.stopPropagation();
                onCancel?.(activeSessionId || undefined);
              }}
              title="Stop turn execution"
            >
              <Square size={11} class="fill-current" />
            </button>
```

- [ ] **Step 4: Update `handleCancelSession` in `App.svelte`**

In `frontend/src/App.svelte`:
```typescript
  // Cancel running session
  async function handleCancelSession(targetSessionId?: unknown, pauseQueue = true) {
    const validTargetId = typeof targetSessionId === 'string' && targetSessionId.trim()
      ? targetSessionId.trim()
      : null;
    const sessionId = validTargetId || sessionStore.activeSession?.id;
    if (!sessionId) return;

    if (window.go?.main?.App?.CancelSession) {
      try {
        await window.go.main.App.CancelSession(sessionId);
      } catch (err) {
        console.error('Failed to cancel session:', err);
      }
    }

    sessionStore.setSessionStatus(sessionId, 'idle');
    sessionStore.updateLastMessage(sessionId, (msg) => {
      if (msg.status === 'streaming') {
        msg.status = 'done';
        msg.content += '\n\n*(Turn cancelled by user)*';
      }
      if (msg.toolCalls && msg.toolCalls.length > 0) {
        for (const tc of msg.toolCalls) {
          if (tc.status === 'running') {
            tc.status = 'error';
            if (!tc.endTime) tc.endTime = Date.now();
          }
        }
      }
    });

    // Small delay to ensure OS process cleanup before any subsequent turn
    await new Promise((resolve) => setTimeout(resolve, 150));
  }
```

- [ ] **Step 5: Run svelte-check to verify no compiler errors**

Run: `cd frontend && npm run check`
Expected: 0 errors.

---

### Task 2: Backend `Runner.Cancel` Multi-Identifier Lookup and Reliable Process Kill

**Files:**
- Modify: `pkg/grokrunner/runner.go:24-34, 480-490, 625-645`
- Modify: `pkg/grokrunner/proc_unix.go:23-38`
- Test: `pkg/grokrunner/runner_test.go`

**Interfaces:**
- Consumes: `sessionID string` in `r.Cancel(sessionID string) error`
- Produces: Resilient termination finding sessions by `SessionID` or `GrokSessionID` or single-running fallback.

- [ ] **Step 1: Write unit tests in Go verifying resilient cancellation**

In `pkg/grokrunner/runner_test.go`:
```go
func TestRunner_Cancel_ResilientLookup(t *testing.T) {
	r := NewRunner("")

	// Fake running active session
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := exec.Command("sleep", "60")
	if err := cmd.Start(); err != nil {
		t.Skip("sleep command unavailable")
	}
	defer func() {
		_ = cmd.Process.Kill()
	}()

	active := &ActiveSession{
		SessionID:     "sess_custom_abc",
		GrokSessionID: "01a15062-c047-7d91-a7e1-10e862e35c64",
		Cmd:           cmd,
		Cancel:        cancel,
		Done:          make(chan struct{}),
	}

	r.mu.Lock()
	r.sessions["sess_custom_abc"] = active
	r.mu.Unlock()

	// 1. Cancel by GrokSessionID
	err := r.Cancel("01a15062-c047-7d91-a7e1-10e862e35c64")
	if err != nil {
		t.Fatalf("unexpected cancel error: %v", err)
	}

	// Verify context is cancelled
	if ctx.Err() == nil {
		t.Errorf("expected session context to be cancelled")
	}
}
```

- [ ] **Step 2: Update `ActiveSession` and `StartSession` in `pkg/grokrunner/runner.go`**

In `pkg/grokrunner/runner.go`:
```go
type ActiveSession struct {
	SessionID     string
	GrokSessionID string
	Cmd           *exec.Cmd
	Cancel        context.CancelFunc
	Stdin         io.WriteCloser
	Stdout        io.ReadCloser
	Done          chan struct{}
}
```
And inside `StartSession`:
```go
	active := &ActiveSession{
		SessionID:     req.SessionID,
		GrokSessionID: resolvedSessionUUID,
		Cmd:           cmd,
		Cancel:        cancel,
		Stdin:         stdin,
		Stdout:        stdout,
		Done:          make(chan struct{}),
	}
```

- [ ] **Step 3: Update `Cancel` in `pkg/grokrunner/runner.go`**

In `pkg/grokrunner/runner.go`:
```go
// Cancel terminates the process group of the active session
func (r *Runner) Cancel(sessionID string) error {
	r.mu.Lock()
	trimmedID := strings.TrimSpace(sessionID)
	active, exists := r.sessions[trimmedID]
	if !exists && trimmedID != "" {
		for _, s := range r.sessions {
			if s.SessionID == trimmedID || (s.GrokSessionID != "" && s.GrokSessionID == trimmedID) {
				active = s
				exists = true
				break
			}
		}
	}
	// Fallback: If not found and exactly one session is actively running, cancel it
	if !exists && len(r.sessions) == 1 {
		for _, s := range r.sessions {
			active = s
			exists = true
			break
		}
	}
	r.mu.Unlock()

	if !exists || active == nil {
		return nil
	}

	active.Cancel()
	if active.Stdout != nil {
		_ = active.Stdout.Close()
	}
	if active.Stdin != nil {
		_ = active.Stdin.Close()
	}
	_ = killProcessGroup(active.Cmd)
	if active.Cmd != nil && active.Cmd.Process != nil {
		_ = active.Cmd.Process.Kill()
	}
	return nil
}
```

- [ ] **Step 4: Run Go unit tests**

Run: `go test -v -count=1 ./pkg/grokrunner`
Expected: PASS

---

### Task 3: Build & Local Installation

**Files:**
- Output: `/Applications/AetherGrok.app`

- [ ] **Step 1: Run complete Vitest suite**

Run: `npx vitest run`
Expected: All 5390+ tests pass.

- [ ] **Step 2: Build desktop app with Wails**

Run: `/Users/fiko942/go/bin/wails build`
Expected: Successful compile and packaging to `build/bin/aethergrok.app`.

- [ ] **Step 3: Install locally to `/Applications/AetherGrok.app`**

Run: `rm -rf /Applications/AetherGrok.app && cp -R build/bin/aethergrok.app /Applications/AetherGrok.app && xattr -cr /Applications/AetherGrok.app`
Expected: Clean installation into `/Applications/AetherGrok.app`.
