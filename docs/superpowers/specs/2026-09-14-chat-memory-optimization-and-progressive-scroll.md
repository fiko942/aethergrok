# Specification: Chat Memory Optimization, DOM Flushing & Progressive Infinite Scroll

## 1. Overview
This specification details the architecture and implementation of the lightweight progressive chat history system and session switching memory optimization in Grok Build.

### Key Goals
- **Minimal DOM Footprint**: Never render full multi-megabyte transcripts in the DOM all at once. Initial view renders only the most recent **10 user turns** (`HISTORY_WINDOW_USER_TURNS = 10`).
- **Progressive Upward Hydration**: Older turns are safely parked in `state.historyPrefix` and lazily hydrated in chunks of **10 user turns** (`HISTORY_PREPEND_USER_TURNS = 10`) when the user scrolls near the top or interacts with `#history-head`.
- **Complete DOM Flush on Session Switch**: When switching away from a session (e.g., Session A to Session B), Session A's transcript nodes in the DOM are completely discarded and state maps are cleared.
- **Fresh Windowing on Session Return**: When returning back to Session A, it always renders only the latest 10 turns, preserving memory and preventing accumulated DOM bloat over long work sessions.

---

## 2. Technical Architecture

### 2.1 History Windowing Partition (`media/webview-helpers.js`)
`splitHistoryWindow(messages, windowTurns)` processes the flattened replay messages and counts user turn boundaries (`userMessage`, non-steer, non-reminder):
- Identifies start indices of counted user turns.
- Splices messages into:
  - `suffix`: The most recent $N$ user turns and their assistant responses/tool outputs (rendered immediately).
  - `prefix`: Older messages (stored in `state.historyPrefix` with counter metrics preserved).

### 2.2 History Prepending & Anchor Preservation (`media/chat.js`)
When `maybeLoadEarlierHistory()` triggers:
1. `state.historyHydrating = true` prevents duplicate or overlapping prepends.
2. `splitHistoryWindow(state.historyPrefix, 10)` extracts the next chunk of 10 turns.
3. A temporary detached container (`historyPark = document.createElement("div")`) captures newly instantiated DOM elements in `try ... finally` blocks.
4. `prependHistoryNodes(nodes)` records the bounding rectangle of `sentinel` (top-most existing message node), inserts new nodes before `sentinel`, and adjusts `messagesEl.scrollTop` by `delta = sentinel.top - y` to preserve the exact visible scroll line without any jump or flicker.

### 2.3 Automatic & Manual Triggers
- **IntersectionObserver**: `#history-head` is monitored with `rootMargin: "400px 0px 0px 0px"`. Approaching the top automatically triggers `maybeLoadEarlierHistory()`.
- **Scroll Event Threshold**: `messagesEl.scrollTop <= HISTORY_PREPEND_PX (1000px)` triggers proactive loading during scroll deceleration.
- **Interactive Button**: `#history-head` displays a styled loading spinner when hydrating and a clickable button showing `↑ Load earlier messages (N remaining)` for direct click or keyboard navigation.

### 2.4 Session Switching & Memory Reclamation
When switching sessions (`clearMessages` via IPC):
- `resetForNewSession()` invokes `markTranscriptPendingClear()`, `clearHistoryWindow()`, and resets all tool/diff/card tracking Maps (`state.pendingDiffByToolCallId`, `state.subagentCards`, etc.).
- When the new session's first turn renders, all pending nodes from the previous session are unlinked and garbage collected.
- Re-focusing a previous session sends `historyReplay`, which re-executes `applyHistoryWindow` to bound the initial render to 10 turns, regardless of whether all history was previously expanded.

---

## 3. Verification & Safety Guarantees
- **No Lost Messages**: Full session history remains intact on disk in ACP `session/load` and in `session.buffer`. Full-text search (`Ctrl+F`) expands history (`expandHistoryAll`) prior to scanning.
- **Zero Lockup / Freeze**: `hydrateHistoryChunkWithCounters` is wrapped in `try ... finally` so `state.historyHydrating` and `state.replaying` can never get stuck on unexpected payload shapes.
- **Test Suite**: Verified with 10 dedicated DOM unit tests in `test/history-window.dom.test.ts` and 6,085 comprehensive tests across the suite.
