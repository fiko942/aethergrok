# Implementation Plan: Session Switch DOM Flush & Progressive Infinite Scroll

> **Status:** Implemented & Verified (100% Tests Passing)

**Goal:** Ensure instant session switching with full DOM cleanup and low memory consumption by capping initial renders to 10 user turns and progressively hydrating earlier turns in 10-turn chunks on upward scrolling.

---

## Tasks Completed

- [x] **Task 1: Calibrate Window and Prepend Size to 10 Turns**
  - Configured `HISTORY_WINDOW_USER_TURNS = 10` in `media/chat.js`.
  - Configured `HISTORY_PREPEND_USER_TURNS = 10` in `media/chat.js`.
  - Configured `HISTORY_PREPEND_PX = 1000` with `IntersectionObserver` on `#history-head`.

- [x] **Task 2: Fix Upward Scroll Lock & Observer**
  - Removed restrictive `stickToBottom` checks from `maybeLoadEarlierHistory()`.
  - Ignored `transcript-loading-bar` and `history-head` in `firstLiveTranscriptChild()`.
  - Preserved scroll line via `restorePrependAnchor` calculating delta against top sentinel.

- [x] **Task 3: Session Switching DOM Flush & Memory Reclamation**
  - Verified `clearMessages` resets `state.historyPrefix` and marks transcript nodes `data-pending-clear="1"`.
  - Verified `appendTranscriptChild` removes stale transcript elements immediately when replacement nodes render.
  - Replayed sessions always re-apply `splitHistoryWindow(held, 10)` to bound DOM tree size.

- [x] **Task 4: Robust Hydration Protection**
  - Wrapped `hydrateHistoryChunkWithCounters` execution in `try ... finally` to guarantee `state.historyHydrating = false` and `state.replaying` restoration even on malformed message inputs.

- [x] **Task 5: Automated Testing & Verification**
  - Added multi-turn session switch and re-entry test in `test/history-window.dom.test.ts`.
  - Verified full test suite passes (251 files, 6,085 tests).
  - Built unpacked desktop distribution and Windows NSIS installer.
