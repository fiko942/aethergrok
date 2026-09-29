# Specification: DOM Virtual Sliding Window & Viewport Pruning with Scroll Anchoring

- **Feature**: Autonomous AI Long-Session Memory Optimization & Scroll Stability
- **Date**: 2026-09-29
- **Platform**: Cross-platform (macOS & Windows)
- **Status**: Production Ready & Fully Verified

---

## 1. Context & Motivation

During complex autonomous workflows, Grok CLI executes multi-step tasks comprising code search, recursive file inspections, iterative file editing, diff generation, and shell commands. In extended interactions, these generate dozens of tool call cards, reasoning traces, and large token streams.

Without virtualization, every message and tool call remains mounted in the DOM, consuming hundreds of megabytes of memory, degrading scroll responsiveness, and triggering layout thrashing.

---

## 2. Architecture & Design Principles

### A. 4-Turn Virtual Sliding Window
- `DEFAULT_WINDOW_TURNS = 4`: The active DOM tree renders only the latest 4 conversational turns (user message + corresponding assistant responses and tool executions).
- Out-of-window messages remain fully preserved in `session.messages` and on-disk JSON storage.
- Auto-flush: Whenever a user submits a new prompt (`role === 'user'`), `session.visibleTurnCount` resets to `DEFAULT_WINDOW_TURNS` (4), automatically pruning older turns from the rendered DOM without requiring a manual refresh.

### B. Viewport Pruning with 800px Tolerance Buffer
- Each `MessageItem` is monitored by an `IntersectionObserver` configured with `rootMargin: '800px 0px 800px 0px'` relative to the viewport window.
- When an item is scrolled outside the 800px tolerance window, its heavy internal DOM nodes (markdown body, syntax-highlighted code blocks, tool call cards, diff summary) are unmounted.
- A virtual spacer is rendered in place with `height: ${lastMeasuredHeight}px` and `min-height: ${lastMeasuredHeight}px`.
- `ResizeObserver` records the physical rendered height (`offsetHeight`) while the item is visible, guaranteeing zero scroll collapse or thumb jitter.
- Active streaming messages (`status === 'streaming'` or active agent turn) remain permanently mounted.

### C. Engine-Level Content Visibility
- Sub-components such as `ToolCallCard.svelte` and `TurnDiffSummary.svelte` use `style="content-visibility: auto; contain-intrinsic-size: auto 34px;"`.
- The browser layout engine natively skips paint, layout, and compositing passes for offscreen tool calls within long assistant responses.

### D. Deterministic Upward Scroll Restoration
- Container element enforces `style="overflow-anchor: none;"` to disable conflicting browser heuristic scroll jumps.
- During upward hydration (`loadEarlierTurns`), `previousScrollHeight` and `previousScrollTop` are recorded before expanding the window.
- Position is restored via `containerEl.scrollTop = previousScrollTop + (scrollHeight - previousScrollHeight)` after `await tick()` and settled with `requestAnimationFrame`.

---

## 3. Verification & Metrics

- Svelte 5 runes diagnostics: 0 errors (`pnpm --prefix frontend check`).
- Production bundle: Vite build succeeds with complete asset generation.
- Go backend test suite: 100% PASS across ConPTY terminal, hotkeys, screen capture, and audio transcription.
