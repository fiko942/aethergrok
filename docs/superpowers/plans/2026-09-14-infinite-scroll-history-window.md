# Infinite Scroll & Virtualized History Prepending Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement a high-performance Infinite Scroll & Progressive History Windowing system in the Grok Build chat view. Long conversations will render an initial lightweight slice of the most recent turns (15–20 turns) for instant first paint, and seamlessly lazy-load older turns in small chunks (15 turns) as the user scrolls up near the top, with zero scroll jitter and an interactive loading spinner.

**Architecture:**
1. **Calibrated History Window Sizing (`media/chat.js`):**
   - Reduce default initial rendered window `HISTORY_WINDOW_USER_TURNS` from 80 down to **20** turns. This cuts initial DOM node creation by ~75% and speeds up session loading by $>4\times$.
   - Reduce prepend chunk size `HISTORY_PREPEND_USER_TURNS` from 40 down to **15** turns for ultra-smooth non-blocking hydration.
   - Increase scroll threshold `HISTORY_PREPEND_PX` to **900px** and add `IntersectionObserver` on `#history-head` for seamless automated pre-fetching before the user hits the top edge.
2. **Visual Loading Indicator at History Head (`media/chat.css` & `media/chat.js`):**
   - Enhance `#history-head` with a sleek loading spinner and progress label (`.history-head-loading` / "Loading earlier messages…") while hydration is in flight.
   - Allow clicking `#history-head` to manually expand earlier messages on demand.
3. **Rock-Solid Scroll Anchor Preservation:**
   - Double-anchor scroll preservation: record the top position of the top-most visible message node (`getBoundingClientRect().top`) before prepending, insert the new nodes, and re-anchor `scrollTop` in the same paint cycle so the user's viewport never shifts or jumps.
4. **Search Integration:**
   - Retain full-text search integration (`expandHistoryAll` on `Ctrl+F` / `findInSession`) so searching across the entire history remains 100% complete and accurate.

**Tech Stack:** JavaScript (DOM, `IntersectionObserver`, `requestAnimationFrame`), CSS, Vitest (Happy-DOM).

**Spec:** `docs/superpowers/specs/2026-09-14-infinite-scroll-history-spec.md`

## Global Constraints

- Never lose or drop earlier messages; they remain preserved in `state.historyPrefix` until scrolled into view.
- Live streaming turns must never be windowed or dropped.
- All existing tests in `test/history-window.dom.test.ts` and `test/webview-ui.dom.test.ts` must pass cleanly.

---

### Task 1: Calibrate History Window & Prepend Sizes in `media/chat.js`

**Files:**
- Modify: `media/chat.js`
- Test: `test/history-window.dom.test.ts`

- [ ] **Step 1: Update window constants and IntersectionObserver in `media/chat.js`**

```javascript
const HISTORY_WINDOW_USER_TURNS = 20;
const HISTORY_PREPEND_USER_TURNS = 15;
const HISTORY_PREPEND_PX = 900;
```

- [ ] **Step 2: Add IntersectionObserver for automatic infinite scroll trigger**

```javascript
let historyHeadObserver = null;

function setupHistoryHeadObserver(head) {
  if (!head || typeof IntersectionObserver === "undefined") return;
  if (historyHeadObserver) historyHeadObserver.disconnect();
  historyHeadObserver = new IntersectionObserver((entries) => {
    for (const entry of entries) {
      if (entry.isIntersecting) {
        maybeLoadEarlierHistory();
      }
    }
  }, { root: messagesEl, rootMargin: "600px 0px 0px 0px" });
  historyHeadObserver.observe(head);
}
```

- [ ] **Step 3: Run DOM tests to verify**

Run: `pnpm test -- test/history-window.dom.test.ts`
Expected: PASS

---

### Task 2: Implement Visual History-Head Loading Spinner in `media/chat.css` and `media/chat.js`

**Files:**
- Modify: `media/chat.css`
- Modify: `media/chat.js`
- Test: `test/webview-ui.dom.test.ts`

- [ ] **Step 1: Add styling for `#history-head` and spinner in `media/chat.css`**

```css
#history-head {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 8px 12px;
  margin: 4px 0 8px;
  font-size: 11.5px;
  color: var(--vscode-descriptionForeground);
  background: var(--vscode-editorWidget-background, rgba(255, 255, 255, 0.03));
  border: 1px dashed var(--vscode-editorWidget-border, #333);
  border-radius: 4px;
  cursor: pointer;
  user-select: none;
  transition: background 0.15s ease, border-color 0.15s ease;
}

#history-head:hover {
  background: var(--vscode-toolbar-hoverBackground, rgba(255, 255, 255, 0.08));
  border-color: var(--vscode-focusBorder, #007acc);
}

.history-head-spinner {
  width: 12px;
  height: 12px;
  border: 2px solid currentColor;
  border-right-color: transparent;
  border-radius: 50%;
  animation: history-spin 0.8s linear infinite;
  margin-right: 6px;
}

@keyframes history-spin {
  to { transform: rotate(360deg); }
}
```

- [ ] **Step 2: Update `syncHistoryHead` in `media/chat.js` to render interactive state**

Display `Loading earlier messages…` with spinner during `state.historyHydrating`, or `↑ Load earlier messages (${remaining} remaining)` when idle.

- [ ] **Step 3: Run DOM tests to verify**

Run: `pnpm test -- test/history-window.dom.test.ts test/webview-ui.dom.test.ts`
Expected: PASS

---

### Task 3: Full Test Suite Verification, Compilation, and Packaging

**Files:**
- Test: `pnpm run compile`
- Test: `pnpm test`
- Build: `pnpm run dist:dir`

- [ ] **Step 1: Compile TypeScript**
- [ ] **Step 2: Run all 251 test files (6,084+ tests)**
- [ ] **Step 3: Package to `dist-desktop/win-unpacked` and verify smooth infinite scrolling on long sessions**

---

## Self-Review

1. **Lightweight First Paint:** Initial load renders only 20 turns, reducing memory and DOM nodes by ~75%.
2. **Smooth Lazy Loading:** Older chunks load incrementally as user approaches top without screen jumps.
3. **Reliability:** Search and live turns remain unaffected.
