# Session Switch Chat Loading, Memory Eviction & Stopped Turn Persistence Specification

**Author:** AetherGrok Engineering  
**Date:** 2026-10-21  
**Status:** In Progress / Investigation Complete  

---

## 1. Problem Statement & User Symptoms

When users frequently switch between sessions or open/close session tabs in the same or different workspaces:
1. **Slow or Blank Chat Rendering:** Switching to another session frequently presents an empty or blank conversation area, taking a long time ("lama banget") to display messages.
2. **Missing Latest Turn on Running Sessions:** When a session is streaming in the background and the user switches to it, the latest response text or tools appear blank or fail to update.
3. **Lost User Query & AI Response on Stopped/Cancelled Sessions:** If a turn is stopped or cancelled by the user and the user switches tabs, returning to that session shows the last chat turn completely missing (both user prompt and AI cancellation note vanish).
4. **Disappearing Earlier Turns:** Earlier conversation history randomly disappears after switching tabs due to aggressive window truncation resetting on every tab switch.

---

## 2. Root Cause Analysis

### 2.1. Aggressive Inactive Session Eviction (`evictInactiveSessionMessages`)
In v1.1.18 (commit `511a88e`), memory eviction was added:
```ts
export function shouldEvictSessionMessages(sessionId: string, activeSessionId: string | null, status: SessionStatus): boolean {
  if (sessionId === activeSessionId) return false;
  if (status === 'working' || status === 'waiting_permission') return false;
  return true;
}
```
Whenever `switchSession` or `openSessionInTab` is invoked:
- Every open tab that is currently `idle`, `finished`, `stopped`, or `error` has its messages array immediately wiped: `session.messages = []`.
- When switching back, `switchSession` sees `target.messages.length === 0` and is forced to perform an asynchronous IPC call (`window.go.main.App.LoadGrokSessionHistory`).
- While this IPC read, file system lookup, JSON parsing of hundreds of lines, and subagent inspection runs, the session messages remain `[]` (the user sees a blank screen).

### 2.2. Stopped / Cancelled Turns Are Never Persisted to `chat_history.jsonl`
When a turn is cancelled (`handleCancelSession`):
- The turn is marked `done` with content `*(Turn cancelled by user)*` in memory.
- Grok CLI was terminated via `SIGTERM` / `SIGKILL` or interrupted, so Grok CLI **never** wrote that cancelled turn to `chat_history.jsonl` on disk.
- Once the user switches tabs, `evictInactiveSessionMessages` purges `session.messages = []`.
- When switching back, `LoadGrokSessionHistory` reads `chat_history.jsonl` from disk. Because the cancelled turn was never on disk, it is completely lost from the conversation.

### 2.3. Background Delta In-Place Mutation Breaks Svelte 5 Reactivity
In v1.1.18 (commit `fe32ab1`), `appendDelta` was optimized for background sessions:
```ts
if (!isTargetActive) {
  if (session.messages.length > 0) {
    const lastMsg = session.messages[session.messages.length - 1];
    if (lastMsg.role === role) {
      lastMsg.content += delta;
      lastMsg.status = 'streaming';
      const lastPart = lastMsg.parts ? lastMsg.parts[lastMsg.parts.length - 1] : null;
      if (lastPart && lastPart.type === 'text') {
        lastPart.text += delta;
      }
      return;
    }
  }
}
```
- In-place mutation of `lastPart.text` without cloning `parts` or updating `session.updatedAt` fails to invalidate Svelte 5 runes (`$derived.by((): MessagePart[] => ...)`).
- When the user switches focus to this running session, Svelte reuses existing DOM elements without updating text, causing the latest chat to appear blank or truncated.

### 2.4. Fragile User Query Tag Stripping in `LoadGrokSessionMessages`
In `pkg/grokrunner/session_scanner.go`:
```go
matches := userQueryRegex.FindStringSubmatch(userText)
if len(matches) > 1 {
    userText = strings.TrimSpace(matches[1])
} else {
    if strings.HasPrefix(userText, "<user_info>") || strings.HasPrefix(userText, "<system-reminder>") || strings.HasPrefix(userText, "This session is being continued") {
        continue
    }
}
```
If a user prompt is prepended with `<user_info>` but lacks `<user_query>` tags, it is discarded. Furthermore, orphan assistant messages following discarded system reminder turns create turn desynchronization.

### 2.5. Forced Window Turn Reset on Every Tab Switch
In `openSessionInTab`:
`target.visibleTurnCount = getActiveWindowTurns();`
This forces `visibleTurnCount` back to 4 or 10 turns on every tab click, hiding earlier messages even if the user had already scrolled up to view them.

---

## 3. Targeted Solution Architecture

1. **Protect Open Tabs from Eviction:**
   - Only evict sessions that are **NOT** in `openTabSessionIds`. All active open tabs in the tab strip must keep their message history in memory. Switching between open tabs must be instantaneous (0ms latency, zero blank flicker).
2. **In-Memory Cache & Uncommitted Turn Merging:**
   - When loading history from disk, merge disk history with in-memory uncommitted turns (such as user-cancelled turns) so that cancelled turns or turns not yet written by Grok CLI are never dropped.
3. **Reactive Background Streaming:**
   - In `appendDelta`, ensure `lastMsg.parts = [...lastMsg.parts]` and `session.updatedAt = Date.now()` are preserved so that Svelte 5 reactivity accurately reflects background streaming upon tab focus.
4. **Preserve User Expanded History Window:**
   - Only initialize `visibleTurnCount` if undefined or 0; preserve the user's current expanded turn count across tab switches.
5. **Robust Go Parser for Chat Messages:**
   - Ensure `LoadGrokSessionMessages` gracefully extracts raw user prompts even when XML tags vary, handles `reasoning` entries, and eliminates orphan assistant messages.
