# Architectural Blueprint & Troubleshooting Guide: Virtual DOM Sliding Window & Scroll Anchoring

## 1. Executive Summary & Objective

In AetherGrok Desktop Studio, autonomous AI agents perform extensive multi-step tasks involving deep codebase research, recursive file discovery, code editing, diff generation, and terminal command execution. In extended workflows, a single session can accumulate hundreds of tool call blocks, thousands of lines of reasoning, and large token streams.

Rendering all conversation turns simultaneously in the DOM causes severe browser memory bloat, high CPU usage, DOM thrashing, and sluggish scrolling.

This blueprint establishes a **4-turn virtual sliding window** pattern with **bidirectional scroll anchoring** and **progressive hydration**. Complete session history is preserved in state and persistent storage, while the active DOM renders only the latest 4 conversational turns by default.

---

## 2. Potential Problems & Risk Analysis

### Problem 1: Scroll Jumping (Layout Drift upon Prepend)
* **Symptom**: When a user scrolls up near the top to reveal earlier turns, new DOM elements are inserted at the top of the container. The browser's default layout pushes the existing content down, causing an abrupt jump in visual position.
* **Root Cause**: The scroll offset (`scrollTop`) remains unchanged while the scroll container's `scrollHeight` increases by the height of newly prepended messages.
* **Solution**: Measure `previousScrollHeight` and `previousScrollTop` immediately before triggering state expansion. Synchronously calculate `heightDiff = container.scrollHeight - previousScrollHeight` after Svelte DOM reconciliation (`await tick()`), and restore `container.scrollTop = previousScrollTop + heightDiff`.

### Problem 2: Native Browser Heuristic Interference (`overflow-anchor`)
* **Symptom**: Modern Chromium (Windows) and WebKit (macOS) engines have built-in `overflow-anchor: auto`, which attempts to guess and anchor visible elements when content changes above them. If both the browser's heuristic and JavaScript execute scroll adjustments, they clash, producing a double-jump or jitter.
* **Solution**: Explicitly disable native heuristic anchoring on the scrollable chat container using CSS `overflow-anchor: none`. This grants the JavaScript coordinator deterministic control over scroll adjustments.

### Problem 3: History Hydration Overriding the Virtual Window
* **Symptom**: Switching between tabs or loading session history from disk displays all 50+ messages in the DOM immediately, bypassing the sliding window entirely.
* **Root Cause**: Deserialization logic setting `session.visibleTurnCount = Math.max(session.visibleTurnCount, session.messages.length)`.
* **Solution**: Always initialize and retain `session.visibleTurnCount = DEFAULT_WINDOW_TURNS` (4) when loading history from disk or creating sessions.

### Problem 4: DOM Accumulation during Continuous Interaction
* **Symptom**: If the user scrolled up to inspect older messages (expanding `visibleTurnCount` to 12) and then submitted a new message, older messages remained in the DOM indefinitely until a manual reload.
* **Root Cause**: `visibleTurnCount` never reset when new user turns began.
* **Solution**: When a new user turn is added (`addMessage` with `role === 'user'`), reset `session.visibleTurnCount = DEFAULT_WINDOW_TURNS` (4). This automatically flushes out-of-window messages from the DOM while preserving them in `session.messages`.

### Problem 5: Infinite Re-trigger Loops on Top Sentinel
* **Symptom**: If an IntersectionObserver triggers `loadEarlier()` and the scroll position adjustment lands near the sentinel threshold, the observer fires repeatedly in a loop until all history is hydrated.
* **Root Cause**: Lacking re-entrancy locks (`isHydrating`) or setting observer threshold/margins improperly.
* **Solution**: Guard `loadEarlier()` with an atomic `isHydrating` flag, check `sessionStore.remainingHiddenTurns > 0`, and adjust scroll offsets so the sentinel element is pushed outside the viewport immediately upon prepend.

### Problem 6: Auto-scroll Conflict During Long Streaming
* **Symptom**: While the AI streams long answers or executes tools, if a user scrolls up to review prior code, the auto-scroll observer forces the view back down to the bottom on every token chunk.
* **Root Cause**: Streaming updates calling `container.scrollTop = container.scrollHeight` without verifying user scroll intent.
* **Solution**: Track `autoScrollToBottom` based on proximity to bottom (`scrollHeight - (scrollTop + clientHeight) < 100`). Only auto-scroll when locked to the bottom; when scrolled up, preserve user position and illuminate the "Jump to latest activity" pill.

### Problem 7: Off-Screen DOM Bloat within Visible Turns
* **Symptom**: In a single long turn, an AI agent can execute 30+ tool calls, produce large diff summaries, and output extensive reasoning. Even when restricted to 4 turns, rendering all off-screen subtrees consumes significant memory and GPU compositing layers.
* **Solution**: Implement fine-grained viewport pruning with height preservation:
  1. Measure exact layout height via `ResizeObserver` when the message is visible.
  2. Use an `IntersectionObserver` with an **800px tolerance buffer** (`rootMargin: 800px 0px 800px 0px`) relative to the window viewport.
  3. When an element is scrolled beyond the 800px tolerance window, replace its heavy children (markdown, syntax highlighters, tool calls, and diff cards) with an off-screen virtual spacer that preserves `height: ${lastMeasuredHeight}px`.
  4. The scroll track and thumb maintain their exact proportions and position without shrinking.
  5. Apply `content-visibility: auto; contain-intrinsic-size: auto 34px;` to individual tool call cards (`ToolCallCard.svelte`) and diff summaries (`TurnDiffSummary.svelte`).

---

## 3. Architecture & Data Flow

```
+-------------------------------------------------------------------------+
|                         AetherGrok Session Store                        |
|                                                                         |
|  +-------------------------------------------------------------------+  |
|  | session.messages: Full Conversation History (State & Disk)         |  |
|  | [Turn 1] [Turn 2] [Turn 3] [Turn 4] [Turn 5] [Turn 6] [Turn 7]    |  |
|  +-------------------------------------------------------------------+  |
|                                     |                                   |
|               User turn slicing: visibleTurnCount = 4                   |
|                                     v                                   |
|  +-------------------------------------------------------------------+  |
|  | visibleMessages (Derived Svelte 5 Rune)                           |  |
|  |                  [Turn 4] [Turn 5] [Turn 6] [Turn 7]              |  |
|  +-------------------------------------------------------------------+  |
+---------------------------------|---------------------------------------+
                                  |
                                  v
+-------------------------------------------------------------------------+
|                        MessageList.svelte (DOM)                         |
|                                                                         |
|  [ Top Sentinel: "Show previous (3 earlier turns)" / Auto-Loader ]      |
|  +-------------------------------------------------------------------+  |
|  | Rendered Turn 4: User Prompt & Tool Calls                         |  |
|  | Rendered Turn 5: User Prompt & Grok Reasoning                     |  |
|  | Rendered Turn 6: User Prompt & File Edits                         |  |
|  | Rendered Turn 7: User Prompt & Live Streaming Response            |  |
|  +-------------------------------------------------------------------+  |
|  [ Bottom Anchor & Floating Activity / Jump Pill ]                      |
+-------------------------------------------------------------------------+
```

---

## 4. Implementation Rules & Constants

1. **Window Constants**:
   - `DEFAULT_WINDOW_TURNS = 4`: The default number of user conversational turns rendered in the active DOM.
   - `PREPEND_CHUNK_TURNS = 4`: The number of user conversational turns added to the window per upward hydration cycle.

2. **Turn Definition**:
   - A turn is bounded by a message with `role === 'user'`.
   - Each turn encompasses the user message and all subsequent assistant thoughts, tool calls, and final answers until the next user message.

3. **Memory Preservation**:
   - Out-of-window messages remain in `session.messages` array and on disk.
   - Compaction, search, and disk serialization read from `session.messages`.
   - Only the visual presentation layer (`visibleMessages`) filters by `visibleTurnCount`.

4. **DOM Synchronization**:
   - `overflow-anchor: none` is applied to the scroll container.
   - All scroll restorations execute after `await tick()`.

---

## 5. Maintenance Checklist for Future Features

- [ ] When adding new message types, ensure they are registered within `session.messages`.
- [ ] Ensure that no feature modifies `session.visibleTurnCount` directly without passing through `sessionStore.loadEarlierTurns()`.
- [ ] When adding complex nested components inside `MessageItem.svelte`, avoid layout shifts after initial render (use fixed aspect ratios or reserved heights for media).
- [ ] Maintain the `isHydrating` lock during any asynchronous prepend operations.
- [ ] Run `pnpm --prefix frontend check` to verify Svelte 5 rune types.
