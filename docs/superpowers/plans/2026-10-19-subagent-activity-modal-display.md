# Subagent Child Activity Stream Display Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ensure all autonomous subagent child actions (`read_file`, `run_terminal_command`, `write`, `search_replace`, etc.) appear inside the Subagent Process modal and subagent card activity previews without leaking into the main chat activity.

**Architecture:** The Go backend locates and parses child subagent execution files (`updates.jsonl` and `chat_history.jsonl`) under `~/.grok/sessions/<workspace>/<subagent_id>/`. The Svelte 5 frontend binds to this backend API, maintains fine-grained reactive `$state` inside `SubagentDetailModal` and `ToolCallCard`, dynamically re-extracts the subagent task ID from both props and live stream results, and auto-populates subagent child tool calls across all chat sessions.

**Tech Stack:** Go 1.23+, Wails v2, Svelte 5 (Runes `$state`, `$derived`, `$effect`, `$props`), TypeScript, Tailwind CSS, Lucide icons.

## Global Constraints

- Never allow child subagent tool calls to leak into the top-level main session transcript.
- Subagent child actions must display in the Subagent Process Modal (`SubagentDetailModal.svelte`) under "Subagent Activity Stream" and in subagent card previews (`ToolCallCard.svelte`).
- The subagent task ID extraction must handle full UUIDs, truncated prefixes, backticked or quoted values, and live updates to `toolCall.result`.
- State updates in Svelte 5 must use Runes (`$state`, `$derived`) to guarantee reactive DOM rendering.
- State points directly in affirmative language. Avoid unnecessary contrastive negation.

---

### Task 1: Robust Subagent ID Extraction and Reactive State in SubagentDetailModal

**Files:**
- Modify: `frontend/src/lib/components/chat/SubagentDetailModal.svelte`
- Modify: `frontend/src/lib/utils/toolUtils.ts`
- Test: `test/subagent-modal.test.ts`
- Test: `test/toolUtils.test.ts`

**Interfaces:**
- Consumes: `extractSubagentMeta(params?: unknown, result?: string): SubagentParsedMeta` from `toolUtils.ts`
- Produces: Reactive `localChildCalls`, `effectiveMeta`, and `displayChildCalls` in `SubagentDetailModal.svelte`

- [ ] **Step 1: Write test for regex extraction of subagent_id with quotes, backticks, and whitespace**

In `test/toolUtils.test.ts`:
```typescript
test('extractSubagentIdFromText handles backticks, quotes, and whitespace', () => {
  expect(extractSubagentIdFromText('Subagent started in background.\nsubagent_id: `01a15226-fcb4-7973-a534-2dc31fbb1a8b`')).toBe('01a15226-fcb4-7973-a534-2dc31fbb1a8b');
  expect(extractSubagentIdFromText('subagent_id: "01a15226-fcb4-7973"')).toBe('01a15226-fcb4-7973');
  expect(extractSubagentIdFromText('subagent_id: 01a15226-fcb4-7973-a534-2dc31fbb1a8b')).toBe('01a15226-fcb4-7973-a534-2dc31fbb1a8b');
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `npm test -- test/toolUtils.test.ts`
Expected: FAIL if backtick or quote handling is missing.

- [ ] **Step 3: Update extractSubagentIdFromText and SubagentDetailModal**

In `frontend/src/lib/utils/toolUtils.ts`:
```typescript
export function extractSubagentIdFromText(text?: string): string | undefined {
  if (!text || typeof text !== 'string') return undefined;
  const match = /subagent_id:\s*[`"']?([0-9a-fA-F-]+)[`"']?/i.exec(text);
  return match ? match[1] : undefined;
}
```

In `frontend/src/lib/components/chat/SubagentDetailModal.svelte`:
Derive `effectiveMeta`:
```typescript
  const effectiveMeta = $derived.by(() => {
    return extractSubagentMeta(toolCall.params, toolCall.result);
  });
```
Add fine-grained reactive `$state` for child calls:
```typescript
  let localChildCalls = $state<ToolCall[]>(toolCall.childToolCalls || []);

  const displayChildCalls = $derived(
    localChildCalls.length > 0 ? localChildCalls : (toolCall.childToolCalls || [])
  );
```
In `loadChildToolCalls`:
```typescript
  async function loadChildToolCalls() {
    const sid = effectiveMeta.taskId || meta.taskId || toolCall.childSessionId || toolCall.id;
    if (!sid) return;

    const activeSession = sessionStore.sessions.find((s) => s.id === sessionStore.activeSessionId);
    let wsPath = '';
    if (activeSession?.workspaceId) {
      const ws = sessionStore.workspaces.find((w) => w.id === activeSession.workspaceId);
      if (ws?.path) wsPath = ws.path;
    }
    if (!wsPath && sessionStore.activeWorkspace?.path) {
      wsPath = sessionStore.activeWorkspace.path;
    }

    if (window.go?.main?.App?.LoadSubagentToolCalls) {
      try {
        const calls = await window.go.main.App.LoadSubagentToolCalls(wsPath, sid);
        if (calls && Array.isArray(calls) && calls.length > 0) {
          const mapped: ToolCall[] = calls.map((c: any) => ({
            id: c.id,
            tool: c.tool,
            params: c.params,
            result: c.result,
            status: c.status === 'failed' ? 'error' : c.status === 'completed' ? 'completed' : 'running',
            startTime: c.startTime,
            endTime: c.endTime
          }));
          toolCall.childToolCalls = mapped;
          localChildCalls = mapped;

          for (const c of calls) {
            planStore.inspectToolForPlan(
              sessionStore.activeSessionId || '',
              c.id,
              c.tool,
              c.params,
              c.result,
              { id: sid, name: effectiveMeta.description || meta.description }
            );
          }
        }
      } catch (err) {
        console.error('Failed to load subagent tool calls:', err);
      }
    }
  }
```
Update template references from `toolCall.childToolCalls` to `displayChildCalls`.

- [ ] **Step 4: Run unit tests to verify pass**

Run: `npm test -- test/toolUtils.test.ts test/subagent-modal.test.ts`
Expected: PASS

---

### Task 2: Subagent Card Preview Population in ToolCallCard and SessionStore

**Files:**
- Modify: `frontend/src/lib/components/chat/ToolCallCard.svelte`
- Modify: `frontend/src/lib/stores/session.svelte.ts`
- Test: `test/subagent-card-parser.test.ts`

**Interfaces:**
- Consumes: `window.go.main.App.LoadSubagentToolCalls`
- Produces: Auto-populated `toolCall.childToolCalls` on subagent cards in the chat transcript

- [ ] **Step 1: Write test for card childToolCalls rendering**

In `test/subagent-card-parser.test.ts`:
Verify subagent card renders child actions when `childToolCalls` has items.

- [ ] **Step 2: Add auto-loading effect in ToolCallCard.svelte**

In `frontend/src/lib/components/chat/ToolCallCard.svelte`:
Add an effect for subagent tool calls that queries `LoadSubagentToolCalls` if `toolCall.childToolCalls` is empty:
```typescript
  $effect(() => {
    if (toolParsed.type !== 'subagent') return;
    if (toolCall.childToolCalls && toolCall.childToolCalls.length > 0) return;

    const sid = toolParsed.meta.taskId || toolCall.childSessionId;
    if (!sid || !window.go?.main?.App?.LoadSubagentToolCalls) return;

    const activeSession = sessionStore.sessions.find((s) => s.id === sessionStore.activeSessionId);
    let wsPath = '';
    if (activeSession?.workspaceId) {
      const ws = sessionStore.workspaces.find((w) => w.id === activeSession.workspaceId);
      if (ws?.path) wsPath = ws.path;
    }
    if (!wsPath && sessionStore.activeWorkspace?.path) {
      wsPath = sessionStore.activeWorkspace.path;
    }

    window.go.main.App.LoadSubagentToolCalls(wsPath, sid)
      .then((calls) => {
        if (calls && Array.isArray(calls) && calls.length > 0) {
          toolCall.childToolCalls = calls.map((c: any) => ({
            id: c.id,
            tool: c.tool,
            params: c.params,
            result: c.result,
            status: c.status === 'failed' ? 'error' : c.status === 'completed' ? 'completed' : 'running',
            startTime: c.startTime,
            endTime: c.endTime
          }));
        }
      })
      .catch(() => {});
  });
```

- [ ] **Step 3: Add populateSessionSubagents to SessionStore**

In `frontend/src/lib/stores/session.svelte.ts`:
Wire `populateSessionSubagents` during `loadSessionHistoryFromDisk` to populate child tool calls for all subagents in the conversation.

- [ ] **Step 4: Run tests to verify pass**

Run: `npm test -- test/subagent-card-parser.test.ts`
Expected: PASS

---

### Task 3: Build Verification, Go Tests, and Desktop App Installation

**Files:**
- Verify: `go test ./pkg/grokrunner/...`
- Verify: `npm run check` / `npm test`
- Build: `npm run build`
- Install: `/Applications/AetherGrok.app`

**Interfaces:**
- Consumes: Go backend binaries and bundled frontend
- Produces: Installed desktop application in `/Applications/AetherGrok.app`

- [ ] **Step 1: Run Go backend tests**

Run: `go test -v ./pkg/grokrunner/...`
Expected: PASS (all tests green)

- [ ] **Step 2: Run frontend unit tests and type checks**

Run: `npm test`
Expected: PASS (all tests green)

- [ ] **Step 3: Compile production build using Wails**

Run: `wails build` (or build script)
Expected: Binary compiles cleanly with all Wails bindings included.

- [ ] **Step 4: Install to /Applications/AetherGrok.app**

Copy the newly built `.app` bundle to `/Applications/AetherGrok.app`.
Verify that the binary inside contains `LoadSubagentToolCalls`.
