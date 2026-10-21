# Session Switch Chat Loading, Memory Eviction & Stopped Turn Persistence Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Eliminate slow/blank chat loading during session switches, ensure background streaming renders immediately without blank delays, prevent loss of user prompts and AI responses on stopped/cancelled turns, and preserve expanded turn history across tab switches.

**Architecture:** 
1. Protect all open tabs in `openTabSessionIds` from memory eviction; only evict background sessions that are closed from tabs and whose status is idle/finished.
2. In `session.svelte.ts`, merge disk history with in-memory uncommitted turns (cancelled turns `*(Turn cancelled by user)*` and live turns) so cancelled turns are never wiped on reload.
3. Fix background session delta streaming in `appendDelta` to preserve Svelte 5 reactivity (`parts` reference refresh and `session.updatedAt`).
4. Avoid resetting `visibleTurnCount` on tab switch if the user has expanded turns.
5. In `pkg/grokrunner/session_scanner.go`, enhance `LoadGrokSessionMessages` to reliably parse user prompts regardless of tag variations, handle reasoning entries, and prevent orphan assistant message grouping.

**Tech Stack:** Go 1.24, Wails v2, Svelte 5 (Runes), TypeScript, Vitest.

**Spec:** `docs/superpowers/specs/2026-10-21-session-switch-chat-loading-and-eviction-fix.md`

## Global Constraints
- Do not break existing lightweight metadata persistence for `localStorage` (`sanitizeSessionForStorage`).
- Keep all existing tests green (5,400+ unit and DOM tests).
- Ensure instantaneous (0ms) switching between open tabs without blank loading states.
- Follow affirmative language rules in all documentation and comments.

---

### Task 1: Protect Open Tabs from Eviction & Preserve Inactive Memory

**Files:**
- Modify: `frontend/src/lib/stores/session.svelte.ts:200-220` (`shouldEvictSessionMessages`, `evictInactiveSessionMessages`)
- Modify: `frontend/src/lib/stores/session.svelte.ts:850-865` (`evictInactiveSessionMessages` method)
- Test: `test/session-memory-persistence.test.ts`

**Interfaces:**
- Consumes: `sessions: Session[]`, `activeSessionId: string | null`, `openTabSessionIds: string[]`
- Produces: `shouldEvictSessionMessages(sessionId: string, activeSessionId: string | null, status: SessionStatus, openTabIds?: string[]): boolean`

- [ ] **Step 1: Write failing test in `test/session-memory-persistence.test.ts` for open tab eviction protection**

Update `test/session-memory-persistence.test.ts` to test that sessions in `openTabIds` are never evicted:
```ts
  it('never evicts messages from sessions that are currently open in tabs', () => {
    const openTabIds = ['s1', 's2'];
    const activeSessionId = 's1';

    // s2 is an open tab, but not active, and status is idle
    expect(shouldEvictSessionMessages('s2', activeSessionId, 'idle', openTabIds)).toBe(false);

    // s3 is neither active nor an open tab -> can be evicted
    expect(shouldEvictSessionMessages('s3', activeSessionId, 'idle', openTabIds)).toBe(true);
  });
```

- [ ] **Step 2: Run test to verify it fails**

Run: `npx vitest run test/session-memory-persistence.test.ts`
Expected: FAIL (argument mismatch or returns true)

- [ ] **Step 3: Update `shouldEvictSessionMessages` and `evictInactiveSessionMessages` in `session.svelte.ts`**

In `frontend/src/lib/stores/session.svelte.ts`:
```ts
export function shouldEvictSessionMessages(
  sessionId: string,
  activeSessionId: string | null,
  status: SessionStatus,
  openTabIds?: string[]
): boolean {
  if (sessionId === activeSessionId) return false;
  if (openTabIds && openTabIds.includes(sessionId)) return false;
  if (status === 'working' || status === 'waiting_permission') return false;
  return true;
}

export function evictInactiveSessionMessages(
  sessions: Session[],
  activeSessionId: string | null,
  openTabIds?: string[]
): void {
  for (const session of sessions) {
    if (shouldEvictSessionMessages(session.id, activeSessionId, session.status, openTabIds)) {
      session.messages = [];
    }
  }
}
```
Update class method `SessionStore.evictInactiveSessionMessages`:
```ts
  evictInactiveSessionMessages(activeSessionId: string | null = this.activeSessionId): void {
    evictInactiveSessionMessages(this.sessions, activeSessionId, this.openTabSessionIds);
  }
```

- [ ] **Step 4: Run test to verify it passes**

Run: `npx vitest run test/session-memory-persistence.test.ts`
Expected: PASS

- [ ] **Step 5: Commit changes**

```bash
git add frontend/src/lib/stores/session.svelte.ts test/session-memory-persistence.test.ts
git commit -m "fix(session): protect open tabs from memory eviction on tab switch"
```

---

### Task 2: Merge In-Memory Cancelled and Uncommitted Turns with Disk History

**Files:**
- Modify: `frontend/src/lib/stores/session.svelte.ts:1600-1648` (`loadSessionHistoryFromDisk`)
- Test: `test/session-uncommitted-turns.test.ts`

**Interfaces:**
- Consumes: `session: Session`, `history: DiscoveredChatMessage[]`
- Produces: `mergeSessionHistoryWithUncommittedTurns(existingMessages: ChatMessage[], loadedHistory: ChatMessage[]): ChatMessage[]`

- [ ] **Step 1: Write unit test in `test/session-uncommitted-turns.test.ts` for preserving uncommitted turns**

Create `test/session-uncommitted-turns.test.ts`:
```ts
import { describe, it, expect } from 'vitest';
import { mergeSessionHistoryWithUncommittedTurns } from '../frontend/src/lib/stores/session.svelte';

describe('Uncommitted Turn Merging', () => {
  it('preserves user prompt and cancelled assistant response when disk history lacks the cancelled turn', () => {
    const existingMessages = [
      { id: 'm1', role: 'user', content: 'Turn 1 prompt', status: 'done' },
      { id: 'm2', role: 'assistant', content: 'Turn 1 answer', status: 'done' },
      { id: 'm3', role: 'user', content: 'Turn 2 prompt (cancelled)', status: 'done' },
      { id: 'm4', role: 'assistant', content: 'Turn 2 partial answer\n\n*(Turn cancelled by user)*', status: 'done' }
    ] as any[];

    // Disk only contains Turn 1 because Grok CLI was interrupted before writing Turn 2
    const diskMessages = [
      { id: 'd1', role: 'user', content: 'Turn 1 prompt', status: 'done' },
      { id: 'd2', role: 'assistant', content: 'Turn 1 answer', status: 'done' }
    ] as any[];

    const merged = mergeSessionHistoryWithUncommittedTurns(existingMessages, diskMessages);

    expect(merged.length).toBe(4);
    expect(merged[2].content).toBe('Turn 2 prompt (cancelled)');
    expect(merged[3].content).toContain('*(Turn cancelled by user)*');
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `npx vitest run test/session-uncommitted-turns.test.ts`
Expected: FAIL ("mergeSessionHistoryWithUncommittedTurns is not a function")

- [ ] **Step 3: Implement `mergeSessionHistoryWithUncommittedTurns` and integrate in `loadSessionHistoryFromDisk`**

In `frontend/src/lib/stores/session.svelte.ts`:
```ts
export function mergeSessionHistoryWithUncommittedTurns(
  existingMessages: ChatMessage[],
  loadedHistory: ChatMessage[]
): ChatMessage[] {
  if (!existingMessages || existingMessages.length === 0) {
    return loadedHistory;
  }
  if (!loadedHistory || loadedHistory.length === 0) {
    return existingMessages;
  }

  // Count user turns in loaded history vs existing in-memory messages
  const loadedUserCount = loadedHistory.filter((m) => m.role === 'user').length;
  const existingUserCount = existingMessages.filter((m) => m.role === 'user').length;

  if (existingUserCount <= loadedUserCount) {
    return loadedHistory;
  }

  // Find where existing messages exceed loaded history
  let userTurnsSeen = 0;
  let splitIndex = existingMessages.length;
  for (let i = 0; i < existingMessages.length; i++) {
    if (existingMessages[i].role === 'user') {
      userTurnsSeen++;
      if (userTurnsSeen > loadedUserCount) {
        splitIndex = i;
        break;
      }
    }
  }

  const uncommittedTrailing = existingMessages.slice(splitIndex);
  return [...loadedHistory, ...uncommittedTrailing];
}
```

In `loadSessionHistoryFromDisk(session: Session)`:
Update the assignment:
```ts
        if (history && history.length > 0) {
          const parsedHistory = history.map((m: any, idx: number) => ({
            id: m.id || `${session.id}_msg_${idx}`,
            role: m.role || 'assistant',
            content: m.content || '',
            timestamp: m.timestamp || (session.createdAt + idx * 1000),
            reasoningContent: m.reasoningContent,
            toolCalls: m.toolCalls,
            parts: m.parts,
            tokens: m.tokens,
            status: m.status || 'done'
          }));
          session.messages = mergeSessionHistoryWithUncommittedTurns(session.messages, parsedHistory);
          if (!session.visibleTurnCount || session.visibleTurnCount <= 0) {
            session.visibleTurnCount = getActiveWindowTurns();
          }
```

- [ ] **Step 4: Run test to verify it passes**

Run: `npx vitest run test/session-uncommitted-turns.test.ts`
Expected: PASS

- [ ] **Step 5: Commit changes**

```bash
git add frontend/src/lib/stores/session.svelte.ts test/session-uncommitted-turns.test.ts
git commit -m "fix(session): merge uncommitted and cancelled turns with disk history on session load"
```

---

### Task 3: Reactive Background Delta Streaming & History Window Preservation

**Files:**
- Modify: `frontend/src/lib/stores/session.svelte.ts:1405-1415` (`openSessionInTab`)
- Modify: `frontend/src/lib/stores/session.svelte.ts:2088-2125` (`appendDelta`)
- Test: `test/background-delta-reactivity.test.ts`

**Interfaces:**
- Consumes: `sessionId: string`, `delta: string`, `role: 'assistant' | 'user'`
- Produces: Reactive Svelte 5 array updates for background session streaming

- [ ] **Step 1: Write test for reactive background delta updates**

Create `test/background-delta-reactivity.test.ts`:
```ts
import { describe, it, expect } from 'vitest';
import { SessionStore } from '../frontend/src/lib/stores/session.svelte';

describe('Background Delta Streaming Reactivity', () => {
  it('updates parts and triggers reactivity timestamp on background session deltas', () => {
    const store = new SessionStore();
    const session = {
      id: 'bg-1',
      title: 'Background Session',
      workspaceId: 'ws-1',
      status: 'working' as const,
      messages: [
        {
          id: 'm1',
          role: 'assistant' as const,
          content: 'Hello',
          status: 'streaming' as const,
          parts: [{ type: 'text' as const, text: 'Hello' }]
        }
      ],
      updatedAt: 1000,
      visibleTurnCount: 4
    } as any;
    store.sessions = [session];
    store.activeSessionId = 'active-other'; // bg-1 is in the background

    const timeBefore = session.updatedAt;
    store.appendDelta('bg-1', ' world!', 'assistant');

    expect(session.messages[0].content).toBe('Hello world!');
    expect(session.messages[0].parts[0].text).toBe('Hello world!');
    expect(session.updatedAt).toBeGreaterThanOrEqual(timeBefore);
  });
});
```

- [ ] **Step 2: Run test to verify**

Run: `npx vitest run test/background-delta-reactivity.test.ts`
Expected: Run test and observe behavior.

- [ ] **Step 3: Update `appendDelta` and `openSessionInTab` in `session.svelte.ts`**

In `frontend/src/lib/stores/session.svelte.ts`:
Around line 2110 in `appendDelta`:
```ts
    // For background sessions, update content and clone parts to preserve Svelte 5 reactivity
    if (!isTargetActive) {
      if (session.messages.length > 0) {
        const lastMsg = session.messages[session.messages.length - 1];
        if (lastMsg.role === role) {
          lastMsg.content += delta;
          lastMsg.status = 'streaming';
          if (!lastMsg.parts) {
            lastMsg.parts = [];
          }
          const lastPart = lastMsg.parts.length > 0 ? lastMsg.parts[lastMsg.parts.length - 1] : null;
          if (lastPart && lastPart.type === 'text') {
            lastPart.text += delta;
          } else {
            lastMsg.parts.push({ type: 'text', text: delta });
          }
          lastMsg.parts = [...lastMsg.parts];
          session.updatedAt = Date.now();
          return;
        }
      }
    }
```

In `openSessionInTab` around line 1409:
Do not overwrite user-expanded turn count:
```ts
    if (!target.visibleTurnCount || target.visibleTurnCount <= 0) {
      target.visibleTurnCount = getActiveWindowTurns();
    }
```

- [ ] **Step 4: Run test to verify it passes**

Run: `npx vitest run test/background-delta-reactivity.test.ts`
Expected: PASS

- [ ] **Step 5: Commit changes**

```bash
git add frontend/src/lib/stores/session.svelte.ts test/background-delta-reactivity.test.ts
git commit -m "fix(session): ensure background delta updates maintain reactivity and preserve visible turn count"
```

---

### Task 4: Robust User Prompt and Message Parsing in Go Backend

**Files:**
- Modify: `pkg/grokrunner/session_scanner.go:480-535` (`LoadGrokSessionMessages`)
- Test: `pkg/grokrunner/session_scanner_test.go`

**Interfaces:**
- Consumes: `chat_history.jsonl` / `updates.jsonl` lines
- Produces: Correctly sequenced `DiscoveredChatMessage` with user prompts extracted even when tags vary or without `<user_query>`

- [ ] **Step 1: Write Go test in `session_scanner_test.go` for prompt tag variations**

In `pkg/grokrunner/session_scanner_test.go`:
```go
func TestLoadGrokSessionMessages_PromptTagVariations(t *testing.T) {
	testDir := t.TempDir()
	sessDir := filepath.Join(testDir, "test-sess")
	_ = os.MkdirAll(sessDir, 0755)

	lines := []string{
		`{"type":"system","content":"System prompt"}`,
		`{"type":"user","content":[{"type":"text","text":"<user_info>OS: win</user_info>\n<rules>rule</rules>\n<user_query>\nAnalyze code\n</user_query>"}]}`,
		`{"type":"assistant","content":"I analyzed the code."}`,
		`{"type":"user","content":[{"type":"text","text":"Plain text prompt without tags"}]}`,
		`{"type":"assistant","content":"Answer to plain text."}`,
	}
	_ = os.WriteFile(filepath.Join(sessDir, "chat_history.jsonl"), []byte(strings.Join(lines, "\n")+"\n"), 0644)

	msgs, err := LoadGrokSessionMessages(testDir, "test-sess")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(msgs) != 4 {
		t.Fatalf("expected 4 messages (2 user, 2 assistant), got %d", len(msgs))
	}
	if msgs[0].Content != "Analyze code" {
		t.Errorf("expected 'Analyze code', got: %s", msgs[0].Content)
	}
	if msgs[2].Content != "Plain text prompt without tags" {
		t.Errorf("expected 'Plain text prompt without tags', got: %s", msgs[2].Content)
	}
}
```

- [ ] **Step 2: Run Go test to verify it passes or fails**

Run: `go test -v ./pkg/grokrunner -run TestLoadGrokSessionMessages_PromptTagVariations`

- [ ] **Step 3: Enhance `LoadGrokSessionMessages` in `session_scanner.go`**

In `pkg/grokrunner/session_scanner.go` around line 505:
Ensure raw user queries without `<user_query>` tags are extracted cleanly rather than discarded:
```go
			userText = strings.TrimSpace(userText)
			// Extract real user prompt from <user_query>
			matches := userQueryRegex.FindStringSubmatch(userText)
			if len(matches) > 1 {
				userText = strings.TrimSpace(matches[1])
			} else {
				// If userText contains system reminders without user query, check if there's user text appended
				if strings.HasPrefix(userText, "<system-reminder>") && !strings.Contains(userText, "</system-reminder>") {
					continue
				}
				if strings.HasPrefix(userText, "<user_info>") {
					// Attempt to find text after last XML tag closing
					lastTagEnd := strings.LastIndex(userText, ">")
					if lastTagEnd != -1 && lastTagEnd < len(userText)-1 {
						trailing := strings.TrimSpace(userText[lastTagEnd+1:])
						if trailing != "" {
							userText = trailing
						} else {
							continue
						}
					} else {
						continue
					}
				}
			}
```

- [ ] **Step 4: Run Go test suite**

Run: `go test -v ./pkg/grokrunner`
Expected: PASS

- [ ] **Step 5: Commit changes**

```bash
git add pkg/grokrunner/session_scanner.go pkg/grokrunner/session_scanner_test.go
git commit -m "fix(scanner): enhance user prompt extraction across XML tag variants and plain text"
```

---

### Task 5: End-to-End Verification Across Frontend and Backend

**Files:**
- Test: `test/` full suite
- Test: `go test ./...`

- [ ] **Step 1: Run complete Go test suite**

Run: `go test -v ./...`
Expected: All packages pass.

- [ ] **Step 2: Run complete frontend unit & DOM test suite**

Run: `npm test`
Expected: 5,400+ tests pass without regression.

- [ ] **Step 3: Build frontend production bundle**

Run: `npm --prefix frontend run build`
Expected: Vite build succeeds with 0 errors.
