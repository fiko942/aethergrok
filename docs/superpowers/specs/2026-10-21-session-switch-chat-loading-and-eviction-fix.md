# Session Switch Chat Loading, Terminal Defaults, macOS Stability & Streaming JSONL Specification

**Author:** AetherGrok Engineering  
**Date:** 2026-10-21  
**Version:** v1.1.19  
**Status:** Implemented & Verified in v1.1.19  

---

## 1. Executive Summary

AetherGrok v1.1.19 addresses critical UX, stability, and performance bottlenecks across session navigation, terminal management, tool execution visibility, and macOS background runtime reliability:

1. **Instant Session Tab Switching & Memory Protection:** All sessions actively open in tabs (`openTabSessionIds`) remain in memory, eliminating blank chat screens and redundant disk IPC loads during tab navigation. Memory eviction is scoped strictly to closed background sessions.
2. **Uncommitted & Cancelled Turn Persistence:** In-memory uncommitted user queries and stopped turns (`*(Turn cancelled by user)*`) merge seamlessly with disk history, preventing loss of conversation context when Grok CLI process termination interrupts transcript flushing.
3. **Reactive Background Delta Streaming & History Window Preservation:** Background streaming updates trigger Svelte 5 reactivity upon tab activation, while user-expanded conversation history windows remain preserved instead of resetting to default turn caps.
4. **Active Tool & Subagent Loading Indicators:** Running tool executions (search, web fetch, grep, read, terminal commands) and subagent operations display clear, animated loading indicators and live status badges instead of rendering blank cards while waiting for output.
5. **Default Closed Session Terminal with Isolated Tab Memory:** Newly created or opened sessions default to a closed/collapsed terminal state to preserve screen space. When a user explicitly opens the terminal for a session, that state is remembered across tab switches.
6. **macOS CGO EventTap Concurrency & Window Stability:** Eliminates silent macOS crashes during idle states by isolating `CGEventTap` MachPort sources to a single dedicated run loop with `kCGEventTapOptionListenOnly`. Configured `HideWindowOnClose: true` in `main.go` to prevent accidental process termination.
7. **Streaming JSONL Reader with ModTime Caching:** Replaces bulk `os.ReadFile` allocations with O(1) memory line streaming via `bufio.Reader` and early-exit detection for session titles. Added an in-memory cache validated against file `ModTime` and `FileSize` for 0.01ms history access.

---

## 2. Technical Architecture & Data Flow

```
┌────────────────────────────────────────────────────────┐
│                   Frontend (Svelte 5)                  │
│                                                        │
│  Tab Navigation / Open Sessions                        │
│  ├─ shouldEvictSessionMessages(id, activeId, status):  │
│  │    Open tabs in tab bar are PROTECTED from eviction │
│  │    Closed idle sessions evicted from JS heap        │
│  │                                                     │
│  ├─ switchSession(id) / openSessionInTab(id):          │
│  │    Preserves target.visibleTurnCount                │
│  │    Refreshes parts array reference for live renders │
│  │    Instant 0ms switch without blank delays          │
│  │                                                     │
│  └─ mergeSessionHistoryWithUncommittedTurns():         │
│       Merges disk messages with trailing in-memory     │
│       uncommitted/cancelled turns                      │
└──────────────────────────┬─────────────────────────────┘
                           │ Wails IPC Bridge
                           ▼
┌────────────────────────────────────────────────────────┐
│                   Backend Engine (Go)                  │
│                                                        │
│  pkg/grokrunner/session_scanner.go:                    │
│  ├─ forEachJSONLLine(path, maxLines, fn):              │
│  │    Streaming bufio.Reader (64 KB buffer)            │
│  │    Early-exit on first user prompt for titles       │
│  │                                                     │
│  ├─ In-Memory ModTime & FileSize Cache:                │
│  │    Checks os.Stat(chatHistoryPath).ModTime()        │
│  │    Instant 0.01ms return on unchanged files         │
│  │                                                     │
│  pkg/hotkey/hotkey_darwin.go:                          │
│  └─ CGEventTap single runloop & ListenOnly mode:       │
│       Eliminates dual-runloop MachPort race condition  │
└────────────────────────────────────────────────────────┘
```

---

## 3. Component Deep Dive

### 3.1. Open Tab Eviction Protection (`session.svelte.ts`)
- `shouldEvictSessionMessages` checks `openTabIds && openTabIds.includes(sessionId)`.
- If a session is open in the tab strip, its transcript is preserved in memory.
- When navigating between open tabs, chat history renders immediately with zero network/disk latency.

### 3.2. Uncommitted & Cancelled Turn Merging
- When Grok CLI is cancelled via the Stop button, it exits before writing the cancelled turn to `chat_history.jsonl`.
- `mergeSessionHistoryWithUncommittedTurns` compares `loadedHistory` user turns with `existingMessages` user turns.
- Trailing uncommitted turns (including the user prompt and `*(Turn cancelled by user)*` message) are appended to the loaded history.

### 3.3. Active Loading Indicators (`ToolCallCard.svelte` & `SubagentDetailModal.svelte`)
- Added dedicated `{:else if toolCall.status === 'running'}` branch rendering animated `Loader2` indicators with dynamic verb descriptions (e.g. `Search in progress...`, `Executing command...`).
- In `SubagentDetailModal`, running child actions show animated execution status, and the final output tab informs the user that the subagent is actively executing tasks.

### 3.4. Default Closed Terminal with Isolated Memory (`terminal.svelte.ts`)
- `isSessionCollapsed(sessionId)` defaults to `true`.
- Any session that has not had its terminal opened by the user starts with the terminal collapsed.
- Opening the terminal sets `sessionCollapsed[sessionId] = false`. When switching tabs, each session remembers its own open/closed state.

### 3.5. macOS CGO MachPort Concurrency & Window Protection
- `startEventTap` in `pkg/hotkey/hotkey_darwin.go` adds the MachPort runloop source only to the dedicated locked thread `g_run_loop`, eliminating multi-threaded Mach port invalidation.
- Uses `kCGEventTapOptionListenOnly`, making the tap immune to timeout termination by the macOS window manager.
- `HideWindowOnClose: true` in `main.go` keeps the application running when the window is closed or hidden.

### 3.6. Streaming JSONL Reader & In-Memory ModTime Cache
- `forEachJSONLLine` uses `bufio.NewReaderSize` (64 KB) to process lines one at a time.
- In `DiscoverGrokSessions`, scanning stops immediately after resolving the first user prompt.
- `sessionCache` validates file modification timestamp and size, serving cached conversation history in 0.01ms.
