# KaTeX LaTeX Mathematical Rendering, Subagent Activity Scoping & v1.1.16 Release Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement or audit this plan task-by-task. Steps use checkbox (`- [x]`) syntax for tracking.

**Goal:** Deliver full mathematical LaTeX rendering powered by KaTeX across primary chat turns and subagent inspector views, isolate internal subagent activities from the main session timeline, ensure instant session stop cancellation, and publish AetherGrok v1.1.16 to GitHub Releases with verified multi-platform artifacts.

**Architecture:** Integrate KaTeX with delimiter pre-parsing and currency detection heuristics in `markdownRenderer.ts`. Scope child tool calls within `SubagentDetailModal.svelte` and streamline `ToolCallCard.svelte`. Enhance process group termination in `pkg/grokrunner/runner.go`. Build and distribute via GitHub Actions release matrix.

**Tech Stack:** Svelte 5, KaTeX 0.16.21, Marked, Wails v2, Go 1.24+, TypeScript, Vitest, NSIS, macOS DMG packager.

---

### Task 1: KaTeX Integration & Mathematical Delimiter Engine

**Files:**
- Modify: `frontend/package.json`
- Modify: `frontend/src/app.css`
- Modify: `frontend/src/lib/utils/markdownRenderer.ts`
- Create: `frontend/src/lib/utils/markdownRenderer.test.ts`

- [x] **Step 1: Install KaTeX and configure stylesheet imports**
  - Added `katex: ^0.16.21` to `frontend/package.json`.
  - Added `@import 'katex/dist/katex.min.css';` to `frontend/src/app.css`.
  - Added `.katex-display-container` with horizontal scroll support.

- [x] **Step 2: Implement delimiter tokenization, currency guards, and KaTeX compilation**
  - Protected fenced code blocks (`PREFENCEDCODEBLOCK`) and inline code (`PREINLINECODESPAN`).
  - Added regex transforms for `\[...\]` and `\(...\)`.
  - Added environment regex for `\begin{equation|align|gather|alignat|multline|matrix|pmatrix|bmatrix|vmatrix|Vmatrix|cases|array|split}`.
  - Implemented `isLikelyCurrency()` heuristic to protect monetary amounts (`$100`, `$5m`, `$10-$20`).
  - Implemented `stripUnsupportedTex()` to sanitize `\label{...}` macros.
  - Wrapped display math in `<div class="katex-display-container">...</div>`.

- [x] **Step 3: Write comprehensive unit test suite**
  - Created `frontend/src/lib/utils/markdownRenderer.test.ts` with 15 test cases covering all edge cases.
  - Verified test passing with `npx vitest run frontend/src/lib/utils/markdownRenderer.test.ts`.

---

### Task 2: Subagent Card Simplification & Detail Inspector

**Files:**
- Modify: `frontend/src/lib/components/chat/ToolCallCard.svelte`
- Modify: `frontend/src/lib/components/chat/SubagentDetailModal.svelte`
- Modify: `frontend/src/lib/components/chat/MessageItem.svelte`
- Test: `test/subagent-modal.test.ts`
- Test: `test/subagent-card-parser.test.ts`

- [x] **Step 1: Streamline Subagent Card display**
  - Filtered raw subagent plan/instruction text from primary card to reduce visual clutter.
  - Highlighted clean task action items directly on the card.
  - Maintained reactive 1-second stopwatch timer updating during active subagent execution.

- [x] **Step 2: Enhance Subagent Detail Modal**
  - Wired full interactive inspector showing task instructions, execution metadata, and child tool actions.
  - Applied `renderMarkdown()` to subagent output report, rendering math and structured text.
  - Provided one-click copy buttons for clean prompt and output without XML boilerplate.

---

### Task 3: Session Stop Cancellation Reliability

**Files:**
- Modify: `frontend/src/lib/components/chat/Composer.svelte`
- Modify: `frontend/src/App.svelte`
- Modify: `pkg/grokrunner/runner.go`
- Modify: `pkg/grokrunner/proc_unix.go`
- Test: `test/session-stop-cancel.test.ts`

- [x] **Step 1: Isolate frontend stop event dispatching**
  - Prevented DOM `PointerEvent` from contaminating `sessionID` resolution.
  - Added fallback to `sessionStore.activeSession?.id`.

- [x] **Step 2: Enforce Go process group termination**
  - Sent `SIGTERM` followed by `SIGKILL` to negative process group ID (`-pgid`).
  - Cleared session from active runner registry to free process handles.

---

### Task 4: Release v1.1.16 Packaging & GitHub Actions Automation

**Files:**
- Modify: `wails.json`
- Modify: `app.go`
- Modify: `frontend/package.json`
- Modify: `build/windows/installer/wails_tools.nsh`
- Modify: `changelog.json`
- Create: Annotated Git tag `v1.1.16`

- [x] **Step 1: Bump semantic version to 1.1.16**
  - Updated all version descriptors across Go, frontend, NSIS, and changelog.

- [x] **Step 2: Build frontend distribution**
  - Executed `npm --prefix frontend run build` to generate production assets and KaTeX webfonts.

- [x] **Step 3: Commit and tag release**
  - Committed changes and created annotated tag `v1.1.16`.
  - Pushed `main` and tag `v1.1.16` to GitHub.

- [x] **Step 4: Monitor GitHub Actions build to successful completion**
  - Tracked run `39206632841` across all matrix jobs:
    - Build Windows (amd64): Succeeded (4m 52s)
    - Build Windows (arm64): Succeeded (4m 53s)
    - Build macOS (arm64): Succeeded (2m 57s)
    - Build macOS (amd64): Succeeded (2m 56s)
    - Build macOS (universal): Succeeded (4m 19s)
    - Create GitHub Release: Succeeded (24s)
  - Verified published release assets at `https://github.com/fiko942/aethergrok/releases/tag/v1.1.16`.
