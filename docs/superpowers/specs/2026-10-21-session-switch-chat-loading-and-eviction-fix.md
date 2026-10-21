# Session Switch Chat Loading, Terminal Defaults, macOS Stability, Token Persistence & Scroll Virtualization Specification

**Author:** AetherGrok Engineering  
**Date:** 2026-10-21  
**Version:** v1.1.20  
**Status:** Implemented, Verified & Released in v1.1.20  

---

## 1. Executive Summary

AetherGrok v1.1.19 and v1.1.20 address critical UX, stability, scroll performance, and resource management across session navigation, terminal management, context limits, and desktop runtime reliability:

1. **Instant Session Tab Switching & Memory Protection:** All sessions actively open in tabs (`openTabSessionIds`) remain in memory, eliminating blank chat screens and redundant disk IPC loads during tab navigation. Memory eviction is scoped strictly to closed background sessions.
2. **Uncommitted & Cancelled Turn Persistence:** In-memory uncommitted user queries and stopped turns (`*(Turn cancelled by user)*`) merge seamlessly with disk history, preventing loss of conversation context when Grok CLI process termination interrupts transcript flushing.
3. **Strict Context Token Limit Persistence:** User-configured token limits (e.g. 128k, 200k, 500k) from `settingsStore` strictly govern session usage and popover calculations. Disk telemetry `signals.json` (which contains raw model capacities such as 1,048,576 tokens) is prevented from overriding user settings. Changes in Settings synchronize immediately to all in-memory and open sessions.
4. **Accurate "Waiting for response" Timer:** The waiting indicator measures elapsed time since the agent's last activity/tool/message (`lastWaitStartTime`) rather than total turn inception, providing clear insight into current network/model latency. It automatically hides when a turn completes.
5. **Elimination of Scroll Thrashing & Virtualization Feedback Loops:** Removed unmounting virtual spacers and `IntersectionObserver` loops from `MessageItem.svelte`. Because `sessionStore.visibleMessages` already enforces clean 4-turn windowing, messages stay mounted without height jitter or cascading scroll feedback loops.
6. **Non-Blocking Subagent Hydration & Disk History Pre-Loading:** Chat history from disk is pre-loaded before tab activation, while child subagent details are hydrated asynchronously in the background (`.catch(() => {})`), eliminating multi-second IPC blocking freezes.
7. **Pure User Turn Counting & Synthetic Reminder Filtering:** Fixed parser in `session_scanner.go` to ignore `<system-reminder>` synthetic injections, ensuring window turn calculations (`visibleTurnCount = 4`) never misplace the active turn into off-screen space.
8. **Active Tool & Subagent Loading Indicators:** Running tool executions (search, web fetch, grep, read, terminal commands) and subagent operations display clear, animated loading indicators and live status badges instead of rendering blank cards while waiting for output.
9. **Default Closed Session Terminal with Isolated Tab Memory:** Newly created or opened sessions default to a closed/collapsed terminal state to preserve screen space. When a user explicitly opens the terminal for a session, that state is remembered across tab switches.
10. **macOS CGO EventTap Concurrency & Window Stability:** Eliminates silent macOS crashes during idle states by isolating `CGEventTap` MachPort sources to a single dedicated run loop with `kCGEventTapOptionListenOnly`. Configured `HideWindowOnClose: true` in `main.go` to prevent accidental process termination.
11. **Streaming JSONL Reader with ModTime Caching:** Replaces bulk `os.ReadFile` allocations with O(1) memory line streaming via `bufio.Reader` and early-exit detection for session titles. Added an in-memory cache validated against file `ModTime` and `FileSize` for 0.01ms history access.

---

## 2. Technical Architecture & Data Flow

```
┌────────────────────────────────────────────────────────┐
│                   Frontend (Svelte 5)                  │
│                                                        │
│  Tab Navigation & Windowed Transcript Rendering        │
│  ├─ shouldEvictSessionMessages(id, activeId, status):  │
│  │    Open tabs in tab bar are PROTECTED from eviction │
│  │    Closed idle sessions evicted from JS heap        │
│  │                                                     │
│  ├─ switchSession(id) / openSessionInTab(id):          │
│  │    Pre-loads history from disk before activeId shift│
│  │    Non-blocking subagent hydration                  │
│  │    Instant snap to bottom without slow smooth crawls│
│  │                                                     │
│  ├─ MessageItem:                                       │
│  │    Mounted directly inside 4-turn window            │
│  │    No unmounting spacers or observer feedback loops │
│  │                                                     │
│  ├─ Token Limit Management:                            │
│  │    getActiveMaxContextTokens() locks to settingsStore│
│  │    signals.json 1M raw capacity ignored             │
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
│  │    Strict exclusion of <system-reminder> turns      │
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

### 3.2. Context Token Limit Enforcement
- `getActiveMaxContextTokens()` reads `settingsStore.maxContextTokens`.
- In `loadSessionUsage`, `session.usage.maxTokens` strictly uses `getActiveMaxContextTokens()`.
- Disk `signals.json` metrics (`contextWindowTokens: 1048576`) are ignored for UI context window limits.
- When modified in `SettingsModal`, the new limit immediately synchronizes across all sessions in `sessionStore.sessions`.

### 3.3. Elimination of Self-Scrolling and Viewport Layout Loops
- In `MessageItem.svelte`, removed `IntersectionObserver` unmounting and variable-height virtual spacers.
- In `MessageList.svelte`, `ResizeObserver` only auto-scrolls when a session is actively streaming in the foreground (`sessionStore.activeSession?.status === 'working'`).
- In `scrollToBottom()`, slow smooth-scrolling animations (`behavior: 'smooth'`) were replaced with instant positioning (`containerEl.scrollTop = containerEl.scrollHeight`).
- Removed auto-load on scroll proximity (`scrollTop < 150`), reserving earlier turn expansion strictly for explicit user button clicks.

### 3.4. Accurate Waiting Timer Logic (`lastWaitStartTime`)
- `session.lastWaitStartTime` records the exact timestamp of the last incoming delta or tool execution.
- While `isWorking`, the composer and message indicator display elapsed time since this last activity.
- When execution finishes, the waiting indicator cleanly disappears.

### 3.5. Uncommitted & Cancelled Turn Merging
- When Grok CLI is cancelled via the Stop button, it exits before writing the cancelled turn to `chat_history.jsonl`.
- `mergeSessionHistoryWithUncommittedTurns` compares `loadedHistory` user turns with `existingMessages` user turns.
- Trailing uncommitted turns (including the user prompt and `*(Turn cancelled by user)*` message) are appended to the loaded history.

### 3.6. Default Closed Terminal with Isolated Memory (`terminal.svelte.ts`)
- `isSessionCollapsed(sessionId)` defaults to `true`.
- Any session that has not had its terminal opened by the user starts with the terminal collapsed.
- Opening the terminal sets `sessionCollapsed[sessionId] = false`. When switching tabs, each session remembers its own open/closed state.

### 3.7. macOS CGO MachPort Concurrency & Window Protection
- `startEventTap` in `pkg/hotkey/hotkey_darwin.go` adds the MachPort runloop source only to the dedicated locked thread `g_run_loop`, eliminating multi-threaded Mach port invalidation.
- Uses `kCGEventTapOptionListenOnly`, making the tap immune to timeout termination by the macOS window manager.
- `HideWindowOnClose: true` in `main.go` keeps the application running when the window is closed or hidden.

### 3.8. Streaming JSONL Reader & In-Memory ModTime Cache
- `forEachJSONLLine` uses `bufio.NewReaderSize` (64 KB) to process lines one at a time.
- In `DiscoverGrokSessions`, scanning stops immediately after resolving the first user prompt.
- `sessionCache` validates file modification timestamp and size, serving cached conversation history in 0.01ms.
