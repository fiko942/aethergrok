# Token Truncation Guard, Auto-Retry & Queue Pause Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Prevent and gracefully handle `max_tokens_truncation` ("response truncated by max_tokens") errors during heavy autonomous tasks and multi-tab workflows through:
1. Automatic continuation retry ("coba lagi" auto-retry up to 2 times) when `max_tokens_truncation` occurs.
2. Strict Queue Preservation & Pause: when an error occurs, never blindly dequeue or execute the next prompt; keep all queued prompts intact.
3. Pre-flight context warnings (when usage >= 70%) and structured UI recovery cards.

**Architecture:** 
1. **Go Backend (`pkg/grokrunner/stream_parser.go`)**: Detects `max_tokens_truncation` from CLI stream output, parses token usage metrics from `promptUsage`, and emits structured `error_kind: "max_tokens_truncation"` with normalized telemetry.
2. **Frontend Session Store & Queue Manager (`session.svelte.ts` & `App.svelte`)**:
   - Manages retry state per session (`autoRetryCount`, max 2 attempts).
   - Fixes the bug in `checkAndDispatchNextQueue`: when an error occurs, the queue is **strictly paused**; `checkAndDispatchNextQueue` is never triggered on failure unless auto-retry succeeds.
   - Executes an automatic continuation turn with `"Lanjutkan dan selesaikan tugas sebelumnya yang terpotong"` / `"Coba lagi"` when `max_tokens_truncation` occurs, allowing multi-turn execution to complete cleanly before processing subsequent queue items.
3. **UI Components (`Composer.svelte`, `QueueStackBar.svelte`, `MessageItem.svelte`)**:
   - Shows active auto-retry badge during recovery.
   - Shows paused queue status indicator if retries fail, with [Retry Failed Task] and [Resume Queue] actions.
   - Provides a pre-flight warning banner in Composer when active context exceeds 70% (140K/200K tokens).

**Tech Stack:** Go 1.24+, Svelte 5 (Runes), TypeScript, Tailwind CSS, Ant Design Dark Tokens, Wails v2.

## Global Constraints

- Never drop or discard remaining queued prompts when an error occurs.
- Never call `checkAndDispatchNextQueue` on an error state.
- Cap auto-retry to a maximum of 2 consecutive attempts per prompt to prevent infinite loops.
- Follow Svelte 5 runes conventions (`$state`, `$derived`, `$props`).
- Strictly adhere to Ant Design Dark token palette (`bg-ant-bg-secondary`, `border-ant-border-secondary`, `text-ant-text`, `text-rose-400`, `text-amber-400`).

---

### Task 1: Backend Structured Token Truncation Detection in Go Stream Parser

**Files:**
- Modify: `pkg/grokrunner/types.go`
- Modify: `pkg/grokrunner/stream_parser.go`
- Test: `pkg/grokrunner/stream_parser_test.go`

**Interfaces:**
- Consumes: Raw NDJSON events from Grok CLI execution stream.
- Produces: `TurnCompleteEvent` and `EventError` with `ErrorKind: "max_tokens_truncation"` and parsed `Usage` metrics.

- [ ] **Step 1: Write failing test in `pkg/grokrunner/stream_parser_test.go`**

```go
func TestStreamParser_MaxTokensTruncationError(t *testing.T) {
	rawJSON := `{"type":"error","error":"Internal error: {\n  \"message\": \"response truncated by max_tokens\",\n  \"error_kind\": \"max_tokens_truncation\",\n  \"promptUsage\": {\n    \"inputTokens\": 5777245,\n    \"outputTokens\": 23083,\n    \"totalTokens\": 5800328,\n    \"reasoningTokens\": 18377,\n    \"modelCalls\": 54,\n    \"numTurns\": 54\n  }\n}"}`

	var completedEvent *TurnCompleteEvent
	var errReceived error

	callbacks := StreamCallbacks{
		OnError: func(err error) {
			errReceived = err
		},
		OnComplete: func(evt TurnCompleteEvent) {
			completedEvent = &evt
		},
	}

	parser := NewStreamParser("sess-1", callbacks)
	err := parser.Parse(context.Background(), strings.NewReader(rawJSON+"\n"))
	if err != nil && err != io.EOF {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if completedEvent == nil {
		t.Fatal("expected TurnCompleteEvent to be emitted")
	}
	if completedEvent.ErrorKind != "max_tokens_truncation" {
		t.Fatalf("expected ErrorKind 'max_tokens_truncation', got '%s'", completedEvent.ErrorKind)
	}
	if completedEvent.TotalTokens != 5800328 {
		t.Fatalf("expected TotalTokens 5800328, got %d", completedEvent.TotalTokens)
	}
	if errReceived == nil {
		t.Fatal("expected OnError callback to be invoked")
	}
	if !strings.Contains(errReceived.Error(), "Context limit reached: Response truncated by max_tokens") {
		t.Fatalf("expected normalized error message, got: %v", errReceived)
	}
}
```

- [ ] **Step 2: Run test to verify failure**

Run: `go test -v -run TestStreamParser_MaxTokensTruncationError ./pkg/grokrunner/...`
Expected: FAIL.

- [ ] **Step 3: Update `types.go` and implement parser logic in `stream_parser.go`**

Add `ErrorKind` and `Usage` to `TurnCompleteEvent` in `pkg/grokrunner/types.go`.
Implement `parseTruncationError` in `pkg/grokrunner/stream_parser.go`.

- [ ] **Step 4: Run test to verify pass**

Run: `go test -v -run TestStreamParser_MaxTokensTruncationError ./pkg/grokrunner/...`
Expected: PASS.

- [ ] **Step 5: Run all Go unit tests**

Run: `go test ./pkg/... ./test/...`
Expected: 100% PASS.

---

### Task 2: Queue Protection & Auto-Retry Engine in `frontend/src/App.svelte`

**Files:**
- Modify: `frontend/src/App.svelte:480-575`
- Modify: `frontend/src/App.svelte:1390-1425`
- Modify: `frontend/src/lib/stores/session.svelte.ts`
- Test: `test/queue-error-isolation.test.ts`

**Interfaces:**
- Consumes: `grok:error` event, `session.queuedPrompts`, and `session.autoRetryCount`.
- Produces: 
  - On error: stops queue dispatch; does NOT call `checkAndDispatchNextQueue`.
  - On `max_tokens_truncation`: triggers automatic continuation retry with `"Lanjutkan dan selesaikan tugas sebelumnya yang terpotong"` (max 2 retries).
  - On retry success: resumes normal queue processing.

- [ ] **Step 1: Write unit test for queue error isolation & retry logic**

Create `test/queue-error-isolation.test.ts`:
```ts
import { describe, it, expect } from 'vitest';

interface QueueState {
  sessionId: string;
  queuedPrompts: string[];
  activePrompt: string | null;
  status: 'idle' | 'working' | 'error';
  retryCount: number;
}

function handleTurnError(
  state: QueueState,
  isTruncation: boolean,
  maxRetries = 2
): { action: 'retry' | 'pause'; nextPrompt?: string } {
  state.status = 'error';

  if (isTruncation && state.retryCount < maxRetries) {
    state.retryCount += 1;
    state.status = 'working';
    return {
      action: 'retry',
      nextPrompt: 'Lanjutkan dan selesaikan tugas sebelumnya yang terpotong'
    };
  }

  // Strictly pause queue: DO NOT dequeue or run next item!
  return { action: 'pause' };
}

describe('Queue Error Isolation & Auto-Retry', () => {
  it('strictly pauses queue and preserves remaining prompts on general error', () => {
    const state: QueueState = {
      sessionId: 's1',
      queuedPrompts: ['Prompt 2', 'Prompt 3'],
      activePrompt: 'Prompt 1',
      status: 'working',
      retryCount: 0
    };

    const res = handleTurnError(state, false);
    expect(res.action).toBe('pause');
    expect(state.queuedPrompts).toHaveLength(2);
    expect(state.queuedPrompts[0]).toBe('Prompt 2');
    expect(state.status).toBe('error');
  });

  it('triggers auto-retry on max_tokens_truncation up to maxRetries', () => {
    const state: QueueState = {
      sessionId: 's1',
      queuedPrompts: ['Next task'],
      activePrompt: 'Heavy reading task',
      status: 'working',
      retryCount: 0
    };

    // First truncation -> Retry 1
    const res1 = handleTurnError(state, true, 2);
    expect(res1.action).toBe('retry');
    expect(state.retryCount).toBe(1);
    expect(state.queuedPrompts).toHaveLength(1); // Queue untouched!

    // Second truncation -> Retry 2
    const res2 = handleTurnError(state, true, 2);
    expect(res2.action).toBe('retry');
    expect(state.retryCount).toBe(2);
    expect(state.queuedPrompts).toHaveLength(1); // Queue untouched!

    // Third truncation -> Exhausted -> Pause queue
    const res3 = handleTurnError(state, true, 2);
    expect(res3.action).toBe('pause');
    expect(state.status).toBe('error');
    expect(state.queuedPrompts).toHaveLength(1); // Next task safe!
  });
});
```

- [ ] **Step 2: Run test to verify it passes**

Run: `pnpm vitest run test/queue-error-isolation.test.ts`
Expected: PASS.

- [ ] **Step 3: Update `App.svelte` and `session.svelte.ts`**

In `frontend/src/lib/stores/session.svelte.ts`:
Add `autoRetryCount?: number;` and `lastExecutedPrompt?: QueuedPrompt;` to `Session` interface.
Reset `session.autoRetryCount = 0` whenever a new user prompt is manually submitted.

In `frontend/src/App.svelte`:
Remove `checkAndDispatchNextQueue` from:
1. `catch (err)` inside `executeTurn` (line 494).
2. `unsubError` inside `EventsOn('grok:error')` (line 1420).

Add auto-retry handler in `frontend/src/App.svelte`:
```ts
async function attemptAutoRetry(sessionId: string, sessionObj: Session, truncationDetails: any) {
  const currentRetries = sessionObj.autoRetryCount || 0;
  const MAX_AUTO_RETRIES = 2;

  if (currentRetries < MAX_AUTO_RETRIES) {
    sessionObj.autoRetryCount = currentRetries + 1;
    sessionStore.setSessionStatus(sessionId, 'working');

    // Notify user via message status
    sessionStore.addMessage(sessionId, {
      role: 'assistant',
      content: `⚠️ *Respon terpotong oleh batas token (max_tokens). Melanjutkan tugas secara otomatis (percobaan ${sessionObj.autoRetryCount}/${MAX_AUTO_RETRIES})...*`,
      status: 'streaming'
    });

    // Short 600ms settling delay before continuing
    setTimeout(() => {
      executeTurn(sessionId, {
        text: 'Lanjutkan dan selesaikan tugas sebelumnya yang terpotong. Fokus pada langkah yang belum terselesaikan.',
        images: [],
        attachments: [],
        model: selectedModel,
        reasoningEffort
      });
    }, 600);
    return true;
  }

  return false;
}
```

In `unsubError`:
```ts
      unsubError = window.runtime.EventsOn('grok:error', async (event: { sessionId: string; error: string }) => {
        if (event.sessionId) {
          const sessionObj = sessionStore.sessions.find((s) => s.id === event.sessionId);
          const truncationDetails = parseTokenTruncationDetails(event.error);

          // Attempt auto-retry if truncation error occurred and retry limit not reached
          if (truncationDetails && sessionObj) {
            const retried = await attemptAutoRetry(event.sessionId, sessionObj, truncationDetails);
            if (retried) {
              return; // Successfully triggered retry; do NOT pause or fail yet
            }
          }

          // Retries exhausted or general error: Mark session as error and LEAVE QUEUE PAUSED
          sessionStore.setSessionStatus(event.sessionId, 'error');
          const lastMsg = sessionObj?.messages[sessionObj.messages.length - 1];
          if (!lastMsg || lastMsg.role !== 'assistant' || (!lastMsg.content.includes(event.error) && !lastMsg.errorKind)) {
            sessionStore.addMessage(event.sessionId, {
              role: 'assistant',
              content: truncationDetails ? 'Context limit reached: Response truncated by max_tokens' : `Error: ${event.error}`,
              status: 'error',
              errorKind: truncationDetails ? 'max_tokens_truncation' : 'general_error',
              errorDetails: truncationDetails || undefined
            });
          }

          if (truncationDetails && sessionObj) {
            sessionStore.loadSessionUsage(sessionObj).catch(() => {});
          }

          if (sessionObj) {
            for (const msg of sessionObj.messages) {
              if (msg.toolCalls && msg.toolCalls.length > 0) {
                for (const tc of msg.toolCalls) {
                  if (tc.status === 'running') {
                    tc.status = 'error';
                    if (!tc.endTime) tc.endTime = Date.now();
                  }
                }
              }
            }
          }

          // NOTICE: checkAndDispatchNextQueue is DELIBERATELY NOT called here!
          // Queued prompts are preserved and paused.
        }
      });
```

---

### Task 3: Interactive Recovery Card & Queue Pause Actions in UI

**Files:**
- Modify: `frontend/src/lib/components/chat/MessageItem.svelte`
- Modify: `frontend/src/lib/components/chat/composer/QueueStackBar.svelte`
- Test: Frontend build & UI verification

**Interfaces:**
- Consumes: `message.errorKind === 'max_tokens_truncation'`, paused queue state.
- Produces: 
  - In `MessageItem.svelte`: [Coba Lagi Sekarang / Retry], [Compact Conversation], [Start Fresh Session].
  - In `QueueStackBar.svelte`: Paused queue banner when session is in `'error'` state with [Resume Queue] and [Clear Queue] actions.

- [ ] **Step 1: Implement Recovery Card in `MessageItem.svelte`**
- [ ] **Step 2: Add Paused Queue Controls in `QueueStackBar.svelte`**
- [ ] **Step 3: Run Vite build to ensure zero errors**

Run: `pnpm --filter aethergrok-frontend build`
Expected: PASS.

---

### Task 4: Composer Pre-Flight High Context Warning Banner

**Files:**
- Modify: `frontend/src/lib/components/chat/Composer.svelte`
- Test: Verification of warning banner display when tokens >= 70%

- [ ] **Step 1: Add banner in `Composer.svelte`**
- [ ] **Step 2: Build frontend and verify**

Run: `pnpm --filter aethergrok-frontend build`
Expected: PASS.

---

### Task 5: Complete Verification & Build Check

**Files:**
- Run: `go test ./pkg/... ./test/...`
- Run: `pnpm vitest run test/queue-error-isolation.test.ts`
- Run: `pnpm --filter aethergrok-frontend build`

- [ ] **Step 1: Run Go tests**
- [ ] **Step 2: Run Vitest tests**
- [ ] **Step 3: Run frontend build**
