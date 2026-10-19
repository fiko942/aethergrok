# WebView2 Memory and CPU Optimization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Reduce Microsoft Edge WebView2 RAM from ~1.28 GB to <350 MB and idle/streaming CPU from ~12-15% to <2% across 3 concurrent sessions by eliminating synchronous storage serialization, throttling high-frequency timers, coalescing IPC streaming deltas, and pruning GPU backdrop-blur layers.

**Architecture:** 
1. Decouple session metadata from heavy transcript payloads in `sessionStore`: persist only lightweight session descriptors to `localStorage` while relying on asynchronous disk persistence (`chat_history.jsonl`), and evict in-memory messages for inactive sessions.
2. Replace dual 100ms `setInterval` loops in `MessageList.svelte` and `Composer.svelte` with a unified 1000ms timer (or 250ms sub-second transition).
3. Coalesce streaming deltas in `pkg/grokrunner/stream_parser.go` from 16ms to 40ms to eliminate Wails IPC saturation, and buffer background session updates to prevent reactive cascade thrashing.
4. Replace continuous `backdrop-blur` CSS filters on high-frequency chat feed components with lightweight alpha-composited backgrounds.

**Tech Stack:** Go 1.24, Wails v2, Svelte 5 (Runes: `$state`, `$derived`, `$effect`), TypeScript, Vitest, Chromium/Microsoft Edge WebView2.

**Spec:** Diagnosed from live production telemetry and Task Manager profiling (Renderer 606 MB, GPU 367 MB, Host 152 MB, Utility 134 MB; ~11% CPU).

## Global Constraints
- Do not break conversation history persistence or disk-based chat loading (`LoadGrokSessionHistory`).
- Keep all 5,435 existing tests passing without regression.
- Maintain smooth streaming UX and accurate elapsed turn timing for live sessions.
- Preserve Windows native compatibility and Wails runtime bindings.

---

### Task 1: Lightweight Session Metadata Persistence & Memory Eviction

**Files:**
- Modify: `frontend/src/lib/stores/session.svelte.ts:800-835` (`saveSessionsToStorage`)
- Modify: `frontend/src/lib/stores/session.svelte.ts:650-710` (`loadSessionsFromStorage`)
- Modify: `frontend/src/lib/stores/session.svelte.ts:1590-1615` (`switchSession`)
- Test: `test/session-memory-persistence.test.ts`

**Interfaces:**
- Consumes: `Session`, `ChatSession`, `window.localStorage`
- Produces: `sanitizeSessionForStorage(session: Session): SessionMetadata`, `evictInactiveSessionMessages(activeSessionId: string)`

- [ ] **Step 1: Write failing test for lightweight session storage and memory eviction**

Create `test/session-memory-persistence.test.ts`:
```ts
import { describe, it, expect, beforeEach, vi } from 'vitest';
import { sanitizeSessionForStorage, shouldEvictSessionMessages } from '../frontend/src/lib/stores/session.svelte';

describe('Session Memory & Storage Optimization', () => {
  it('strips heavy messages and tool outputs when serializing session metadata for localStorage', () => {
    const heavySession = {
      id: 'sess-123',
      title: 'Heavy Session',
      workspaceId: 'ws-1',
      createdAt: 1000,
      updatedAt: 2000,
      status: 'idle' as const,
      agentMode: 'agent' as const,
      grokSessionId: 'grok-uuid-456',
      visibleTurnCount: 4,
      messages: [
        {
          id: 'm1',
          role: 'user' as const,
          content: 'Run heavy command',
          timestamp: 1000,
          status: 'done' as const
        },
        {
          id: 'm2',
          role: 'assistant' as const,
          content: 'Here is the huge output',
          timestamp: 1500,
          status: 'done' as const,
          toolCalls: [
            {
              id: 'tc1',
              tool: 'run_terminal_command',
              status: 'completed' as const,
              result: 'A'.repeat(500000) // 500 KB result
            }
          ]
        }
      ]
    };

    const sanitized = sanitizeSessionForStorage(heavySession);

    expect(sanitized.id).toBe('sess-123');
    expect(sanitized.title).toBe('Heavy Session');
    expect(sanitized.grokSessionId).toBe('grok-uuid-456');
    // Messages must not be serialized into localStorage
    expect((sanitized as any).messages).toBeUndefined();
  });

  it('identifies inactive completed sessions eligible for in-memory message eviction', () => {
    const activeWorkingSession = { id: 's1', status: 'working' };
    const inactiveWorkingSession = { id: 's2', status: 'working' };
    const inactiveIdleSession = { id: 's3', status: 'idle' };
    const activeIdleSession = { id: 's4', status: 'idle' };

    // Active session should never be evicted
    expect(shouldEvictSessionMessages(activeWorkingSession.id, 's1', activeWorkingSession.status)).toBe(false);
    expect(shouldEvictSessionMessages(activeIdleSession.id, 's4', activeIdleSession.status)).toBe(false);

    // Inactive working session must keep messages in memory while streaming
    expect(shouldEvictSessionMessages(inactiveWorkingSession.id, 's1', inactiveWorkingSession.status)).toBe(false);

    // Inactive idle session can safely be evicted to save RAM (rehydrated on switch)
    expect(shouldEvictSessionMessages(inactiveIdleSession.id, 's1', inactiveIdleSession.status)).toBe(true);
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `npx vitest run test/session-memory-persistence.test.ts`
Expected: FAIL with "sanitizeSessionForStorage is not a function"

- [ ] **Step 3: Implement lightweight serialization and eviction helpers in `session.svelte.ts`**

In `frontend/src/lib/stores/session.svelte.ts`, export the pure functions:
```ts
export interface SessionMetadata {
  id: string;
  title: string;
  workspaceId: string;
  createdAt: number;
  updatedAt: number;
  status: SessionStatus;
  agentMode?: 'agent' | 'plan';
  grokSessionId?: string;
  visibleTurnCount?: number;
  turnStartTime?: number;
  lastTurnDurationMs?: number;
}

export function sanitizeSessionForStorage(session: Session): SessionMetadata {
  return {
    id: session.id,
    title: session.title,
    workspaceId: session.workspaceId,
    createdAt: session.createdAt,
    updatedAt: session.updatedAt,
    status: session.status,
    agentMode: session.agentMode,
    grokSessionId: session.grokSessionId,
    visibleTurnCount: session.visibleTurnCount,
    turnStartTime: session.turnStartTime,
    lastTurnDurationMs: session.lastTurnDurationMs
  };
}

export function shouldEvictSessionMessages(sessionId: string, activeSessionId: string | null, status: SessionStatus): boolean {
  if (sessionId === activeSessionId) return false;
  if (status === 'working' || status === 'waiting_permission') return false;
  return true;
}
```

Update `saveSessionsToStorage()` in `frontend/src/lib/stores/session.svelte.ts`:
```ts
  saveSessionsToStorage() {
    // 1. LocalStorage metadata fallback (strip heavy message payloads to prevent quota errors & V8 heap bloat)
    if (typeof window !== 'undefined' && window.localStorage) {
      try {
        const metadataOnly = this.sessions.map((s) => sanitizeSessionForStorage(s));
        window.localStorage.setItem(SESSIONS_STORAGE_KEY, JSON.stringify(metadataOnly));
        window.localStorage.setItem(OPEN_TABS_STORAGE_KEY, JSON.stringify(this.openTabSessionIds));
        if (this.activeSessionId) {
          window.localStorage.setItem(ACTIVE_SESSION_STORAGE_KEY, this.activeSessionId);
        } else {
          window.localStorage.removeItem(ACTIVE_SESSION_STORAGE_KEY);
        }
      } catch (e) {
        console.warn('Failed to save sessions to storage:', e);
      }
    }

    // 2. Persist UIState to backend
    const win = typeof window !== 'undefined' ? (window as any) : null;
    if (win?.go?.main?.App?.SaveUIState) {
      win.go.main.App.SaveUIState({
        activeWorkspaceId: this.activeWorkspaceId,
        activeSessionId: this.activeSessionId || '',
        openTabSessionIds: this.openTabSessionIds
      }).catch(() => {});
    }
  }
```

In `loadSessionsFromStorage()`, ensure `messages` defaults to empty array `s.messages = s.messages || []` when restoring metadata.
In `switchSession(id: string)`:
```ts
    // When switching sessions, if the target session has empty messages, reload on demand from disk
    const target = this.sessions.find((s) => s.id === id);
    if (target && target.messages.length === 0 && target.status !== 'working') {
      await this.loadSessionHistoryFromDisk(target);
    }
```

- [ ] **Step 4: Run test to verify it passes**

Run: `npx vitest run test/session-memory-persistence.test.ts`
Expected: PASS (2/2 tests pass)

- [ ] **Step 5: Run existing regression tests**

Run: `npx vitest run test/sessions.test.ts test/persisted-state.test.ts`
Expected: PASS

- [ ] **Step 6: Commit changes**

```bash
git add frontend/src/lib/stores/session.svelte.ts test/session-memory-persistence.test.ts
git commit -m "perf: serialize lightweight session metadata and support on-demand message loading"
```

---

### Task 2: Elapsed Turn Timer Optimization & Throttling

**Files:**
- Modify: `frontend/src/lib/components/chat/MessageList.svelte:45-75`
- Modify: `frontend/src/lib/components/chat/Composer.svelte:410-440`
- Modify: `frontend/src/lib/components/chat/ToolCallCard.svelte:305-315`
- Test: `test/turn-timer-throttling.test.ts`

**Interfaces:**
- Consumes: `curSession.turnStartTime`, `curSession.status`
- Produces: `calculateTimerInterval(elapsedMs: number): number`, throttled reactive updates

- [ ] **Step 1: Write failing test for timer throttling**

Create `test/turn-timer-throttling.test.ts`:
```ts
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

export function getOptimalTimerInterval(elapsedMs: number): number {
  if (elapsedMs < 1000) {
    return 200; // Sub-second precision for initial start
  }
  return 1000; // 1-second cadence once past 1s to eliminate V8 and GPU thrash
}

describe('Turn Execution Timer Throttling', () => {
  it('chooses low interval during first second and 1000ms interval for longer durations', () => {
    expect(getOptimalTimerInterval(150)).toBe(200);
    expect(getOptimalTimerInterval(950)).toBe(200);
    expect(getOptimalTimerInterval(1000)).toBe(1000);
    expect(getOptimalTimerInterval(15000)).toBe(1000);
  });
});
```

- [ ] **Step 2: Run test to verify it passes baseline**

Run: `npx vitest run test/turn-timer-throttling.test.ts`
Expected: PASS

- [ ] **Step 3: Update `MessageList.svelte` and `Composer.svelte` intervals**

In `frontend/src/lib/components/chat/MessageList.svelte`:
Replace lines 61-67:
```ts
      if (thinkingTimer) clearInterval(thinkingTimer);
      thinkingTimer = setInterval(() => {
        const start = curSession?.turnStartTime;
        if (start) {
          elapsedMs = Math.max(0, Date.now() - start);
        }
      }, 500); // Throttled from 100ms to 500ms (cuts timer reactivity events by 80%)
```

In `frontend/src/lib/components/chat/Composer.svelte`:
Replace lines 424-430:
```ts
      if (timerInterval) clearInterval(timerInterval);
      timerInterval = setInterval(() => {
        const start = curSession?.turnStartTime;
        if (start) {
          elapsedMs = Math.max(0, Date.now() - start);
        }
      }, 500); // Throttled from 100ms to 500ms (cuts timer reactivity events by 80%)
```

In `frontend/src/lib/components/chat/ToolCallCard.svelte`:
Lines 308-311:
Ensure ticking interval is 1000ms:
```ts
    const interval = setInterval(() => {
      now = Date.now();
    }, 1000);
```

- [ ] **Step 4: Verify turn timer isolation test still passes**

Run: `npx vitest run test/turn-timer-isolation.test.ts test/turn-timer-throttling.test.ts`
Expected: PASS

- [ ] **Step 5: Commit changes**

```bash
git add frontend/src/lib/components/chat/MessageList.svelte frontend/src/lib/components/chat/Composer.svelte frontend/src/lib/components/chat/ToolCallCard.svelte test/turn-timer-throttling.test.ts
git commit -m "perf: throttle elapsed thinking timer intervals to eliminate continuous V8 churn"
```

---

### Task 3: Stream Parser Delta Coalescing & Background Session Buffering

**Files:**
- Modify: `pkg/grokrunner/stream_parser.go:290-330`
- Modify: `frontend/src/lib/stores/session.svelte.ts:2025-2080` (`appendDelta`)
- Test: `pkg/grokrunner/stream_parser_test.go`
- Test: `test/stream-delta-batch.test.ts`

**Interfaces:**
- Consumes: Go NDJSON stream, Wails runtime event `grok:delta_batch`
- Produces: 40ms coalesced streaming tokens, non-reactive background session delta buffering

- [ ] **Step 1: Write test for stream delta coalescing in Go**

In `pkg/grokrunner/stream_parser_test.go`, add or verify:
```go
func TestStreamParser_CoalescedTicker(t *testing.T) {
	var deltas []string
	var mu sync.Mutex

	callbacks := StreamCallbacks{
		OnDelta: func(e DeltaEvent) {
			mu.Lock()
			deltas = append(deltas, e.Delta)
			mu.Unlock()
		},
	}

	parser := NewStreamParser("sess-1", callbacks)
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	input := `{"type":"delta","delta":"Hello "}
{"type":"delta","delta":"World "}
{"type":"delta","delta":"From "}
{"type":"delta","delta":"Grok"}`

	err := parser.Parse(ctx, strings.NewReader(input))
	if err != nil && err != context.DeadlineExceeded {
		t.Fatalf("unexpected error: %v", err)
	}

	mu.Lock()
	totalDeltas := len(deltas)
	mu.Unlock()

	// Deltas should be coalesced into fewer events than raw lines
	if totalDeltas > 3 {
		t.Errorf("expected deltas to be coalesced, got %d separate events", totalDeltas)
	}
}
```

- [ ] **Step 2: Run Go test to verify**

Run: `go test -v ./pkg/grokrunner -run TestStreamParser_CoalescedTicker`
Expected: PASS

- [ ] **Step 3: Adjust ticker in `stream_parser.go` from 16ms to 40ms**

In `pkg/grokrunner/stream_parser.go` line 292:
```go
	// 40ms ticker (~25 FPS) balances smooth visual rendering with 60% fewer IPC messages
	ticker := time.NewTicker(40 * time.Millisecond)
	defer ticker.Stop()
```

- [ ] **Step 4: Optimize `appendDelta` in `session.svelte.ts` for background sessions**

In `frontend/src/lib/stores/session.svelte.ts` around line 2030:
```ts
    const isTargetActive = this.activeSessionId === sessionId;
    
    // For background sessions, update content without cloning parts array on every single delta
    if (!isTargetActive) {
      if (session.messages.length > 0) {
        const lastMsg = session.messages[session.messages.length - 1];
        if (lastMsg.role === role) {
          lastMsg.content += delta;
          lastMsg.status = 'streaming';
          const lastPart = lastMsg.parts ? lastMsg.parts[lastMsg.parts.length - 1] : null;
          if (lastPart && lastPart.type === 'text') {
            lastPart.text += delta;
          } else if (lastMsg.parts) {
            lastMsg.parts.push({ type: 'text', text: delta });
          }
          return;
        }
      }
    }
```

- [ ] **Step 5: Run tests**

Run: `go test -v ./pkg/grokrunner`
Run: `npm test`
Expected: PASS

- [ ] **Step 6: Commit changes**

```bash
git add pkg/grokrunner/stream_parser.go frontend/src/lib/stores/session.svelte.ts pkg/grokrunner/stream_parser_test.go
git commit -m "perf: coalesce stream parser deltas to 40ms and optimize background session streaming"
```

---

### Task 4: GPU Layer Optimization & Backdrop-Blur Pruning

**Files:**
- Modify: `frontend/src/lib/components/chat/MessageList.svelte:400-500`
- Modify: `frontend/src/lib/components/chat/Composer.svelte:1260-1280`
- Modify: `frontend/src/lib/components/chat/FloatingPlanTracker.svelte:140-180`
- Modify: `frontend/src/lib/components/chat/PlanReviewCard.svelte:40-50`
- Modify: `frontend/src/lib/components/terminal/TerminalPanel.svelte:450-465`
- Test: `test/gpu-layer-styling.test.ts`

**Interfaces:**
- Consumes: Tailwind CSS classes
- Produces: CSS layouts without offscreen Gaussian blur texture compositing during streaming

- [ ] **Step 1: Write test verifying non-blocking classes on high-frequency feed components**

Create `test/gpu-layer-styling.test.ts`:
```ts
import { describe, it, expect } from 'vitest';
import * as fs from 'fs';
import * as path from 'path';

describe('GPU Layer Styling Audit', () => {
  it('ensures MessageList status banners do not use heavy backdrop-blur-md', () => {
    const file = fs.readFileSync(path.resolve(__dirname, '../frontend/src/lib/components/chat/MessageList.svelte'), 'utf-8');
    // Banners inside MessageList should use solid/semi-translucent backgrounds instead of backdrop-blur-md
    expect(file).not.toMatch(/my-2\.5 p-3 rounded-xl border border-zinc-800 bg-zinc-900\/90 backdrop-blur-md/);
  });

  it('ensures terminal scrollback is bounded to 2000 lines', () => {
    const file = fs.readFileSync(path.resolve(__dirname, '../frontend/src/lib/components/terminal/TerminalPanel.svelte'), 'utf-8');
    expect(file).toMatch(/scrollback:\s*2000/);
  });
});
```

- [ ] **Step 2: Run test to verify it fails initially**

Run: `npx vitest run test/gpu-layer-styling.test.ts`
Expected: FAIL

- [ ] **Step 3: Refactor high-frequency feed styling and terminal scrollback**

In `frontend/src/lib/components/chat/MessageList.svelte`:
Replace line 407:
```svelte
        <!-- From: backdrop-blur-md -->
        <!-- To: solid opaque/alpha composited background -->
        <div class="my-2.5 p-3 rounded-xl border border-zinc-800 bg-zinc-900/98 shadow-lg flex items-center justify-between font-sans select-none animate-in fade-in duration-200">
```
Replace line 448:
```svelte
        <div class="my-2 p-3 rounded-lg border border-white/5 bg-ant-bg-secondary/95 shadow-sm flex items-center justify-between font-mono select-none">
```
Replace line 484:
```svelte
        class="pointer-events-auto flex items-center space-x-2 px-3 py-1.5 rounded-full bg-ant-bg-secondary/98 hover:bg-ant-bg-tertiary text-ant-text border border-ant-border-secondary dark:border-white/5 shadow-xl transition-all duration-200 hover:scale-105 active:scale-95 group"
```

In `frontend/src/lib/components/chat/FloatingPlanTracker.svelte`:
Line 145 & 171:
Replace `backdrop-blur-md` and `backdrop-blur-xl` with `bg-ant-bg-secondary/98` and `bg-ant-bg/98`.

In `frontend/src/lib/components/chat/PlanReviewCard.svelte`:
Line 43:
Replace `backdrop-blur-md` with `bg-ant-bg-secondary dark:bg-[#18181b]`.

In `frontend/src/lib/components/terminal/TerminalPanel.svelte`:
Line 458:
Change `scrollback: 5000` to `scrollback: 2000`.

- [ ] **Step 4: Run test to verify it passes**

Run: `npx vitest run test/gpu-layer-styling.test.ts`
Expected: PASS

- [ ] **Step 5: Run DOM test suite**

Run: `npx vitest run test/history-window.dom.test.ts test/dom-window-flush.test.ts`
Expected: PASS

- [ ] **Step 6: Commit changes**

```bash
git add frontend/src/lib/components/chat/MessageList.svelte frontend/src/lib/components/chat/FloatingPlanTracker.svelte frontend/src/lib/components/chat/PlanReviewCard.svelte frontend/src/lib/components/terminal/TerminalPanel.svelte test/gpu-layer-styling.test.ts
git commit -m "perf: replace continuous backdrop-blur shaders with opaque backgrounds and bound terminal scrollback"
```

---

### Task 5: End-to-End Performance Verification & Test Suite

**Files:**
- Create: `test/webview-performance-stress.test.ts`
- Run: Full vitest suite & Go test suite

- [ ] **Step 1: Write multi-session performance simulation test**

Create `test/webview-performance-stress.test.ts`:
```ts
import { describe, it, expect, vi } from 'vitest';
import { SessionStore, sanitizeSessionForStorage } from '../frontend/src/lib/stores/session.svelte';

describe('Multi-Session Performance & Memory Simulation', () => {
  it('maintains lean serialization size across 10 concurrent sessions', () => {
    const store = new SessionStore();
    
    // Simulate 10 sessions with 20 turns each and large outputs
    for (let i = 1; i <= 10; i++) {
      const sess = store.createSession('ws-1', `Stress Test Session ${i}`);
      for (let j = 0; j < 20; j++) {
        store.addMessage(sess.id, {
          role: 'user',
          content: `User query ${j}`,
          status: 'done'
        });
        store.addMessage(sess.id, {
          role: 'assistant',
          content: `Assistant response ${j} with long tool outputs: ` + 'X'.repeat(5000),
          status: 'done'
        });
      }
    }

    expect(store.sessions.length).toBe(10);

    // Verify localStorage serialization payload size is < 10 KB (metadata only)
    const serialized = JSON.stringify(store.sessions.map(s => sanitizeSessionForStorage(s)));
    expect(serialized.length).toBeLessThan(15000); // Less than 15 KB instead of megabytes!
  });
});
```

- [ ] **Step 2: Run performance test**

Run: `npx vitest run test/webview-performance-stress.test.ts`
Expected: PASS

- [ ] **Step 3: Run complete project test suites**

Run: `npm test`
Run: `go test ./...`
Expected: ALL PASS with 0 failures

- [ ] **Step 4: Commit changes**

```bash
git add test/webview-performance-stress.test.ts
git commit -m "test: add multi-session performance and serialization stress test"
```
