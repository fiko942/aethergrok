# AetherGrok Desktop (Ant Design GUI for Grok Build) - Comprehensive System Plan & Progress

## Executive Summary
This document serves as the master planning, architectural state, completed milestones, pending backlog, and out-of-scope boundaries for **AetherGrok Desktop** (`grok-desktop` / `grok-build`). All specifications and progress details are preserved here to enable seamless continuity across sessions.

---

## 1. Project Architecture & Technology Stack

| Layer | Technology | Key Details |
| :--- | :--- | :--- |
| **Desktop Framework** | Go 1.24 + Wails v2 | Native OS window bindings, system APIs, cross-platform compilation (macOS/Windows) |
| **Frontend Framework** | Svelte 5 (Runes) | Fine-grained reactivity via `$state`, `$derived`, `$props`, `$effect` |
| **Styling & Theme** | Tailwind CSS + Ant Design Tokens | Authentic Ant Design Dark (`#000000`, `#141414`, `#1f1f1f`, `#262626`, `#303030`, `#177ddc`) |
| **Typography & Icons** | SF Pro / System Fonts + Lucide Svelte | Clean technical desktop aesthetic |
| **Build & Run** | `./dev.sh` / `wails dev` / `wails build` | Multiplatform build scripts, Vite frontend bundling, Go backend compilation |

---

## 2. Completed Milestones & Accomplishments

### A. Desktop Skeleton & Session Management
- [x] **Go Wails Backend Initialized**: Core Wails v2 architecture in `main.go` and `app.go`.
- [x] **Session Discovery & Scanning**: `session_scanner.go` locates and loads sessions from `~/.grok/sessions/<workspace_hash>/<session_id>/`.
- [x] **Session Multi-tab Navigation**: `SessionTabs.svelte` supports opening, closing, and switching active chat sessions.
- [x] **Batch Session Actions**: `BatchActionBar.svelte` supports multi-select, delete, and Markdown export (`markdownExport.ts`).
- [x] **Auto-naming & Rename Protection**: Auto-derives session names from the initial prompt and preserves user custom renames.

### B. Chat & Message Visualizer
- [x] **Z-Pattern Conversational Layout**: User prompts aligned to the right; Grok agent responses aligned to the left in `MessageItem.svelte`.
- [x] **Specialized Tool Output Display**:
  - `DiffCard.svelte` with clean diff rendering and syntax highlights.
  - `ToolCallCard.svelte` for shell commands and tool execution results.
  - Formatted file read/edit views with line number and offset indicators.
- [x] **Real-time Status Indicator**: Replaced mock progress indicators with actual turn state handling.

### C. Composer & Controls
- [x] **Compact Composer Layout**: Redesigned input box in `Composer.svelte` with inline control pills.
- [x] **Reasoning Effort Selector**: `ReasoningEffortDropdown.svelte` and `ModelEffortPopover.svelte` supporting low, medium, and high thinking levels.
- [x] **Model Selector**: `ModelSelectDropdown.svelte` configured for Grok models.
- [x] **Agent Mode Dropdown**: `AgentModeDropdown.svelte` with Autonomous, Plan-only, and Code-focused execution modes.
- [x] **Token Gauge Visualization**: Circular SVG donut chart in `ContextUsagePopover.svelte` showing dynamic context token usage.

### D. Snapshot & Media Capture
- [x] **macOS & Windows Screen Capture**: Native snapshot integration with visual flash effect (`ScreenFlash.svelte`) and audio feedback (`audio.ts`).

---

## 3. In-Progress & Pending Tasks

### Task 1: Complete Dark Theme & Ant Design Polish (Highest Priority)
- **Status**: `Completed`
- **Objective**: Standardize all UI components to Ant Design Dark specifications, eliminating harsh white borders (`border-white/20`, `border-white/40`) and saturated neon glow.
- **Action Items**:
  - Refactor `frontend/src/app.css` dark tokens:
    - Base Background: `#000000`
    - Container Background: `#141414`
    - Elevated Surface: `#1f1f1f`
    - Spotlight/Hover: `#262626`
    - Standard Border: `#303030` (`rgba(255, 255, 255, 0.08)`)
    - Primary Accent: `#177ddc` (Ant Design Blue)
    - Primary Text: `rgba(255, 255, 255, 0.85)`
    - Secondary Text: `rgba(255, 255, 255, 0.45)`
  - Update `tokens.ts`, `WorkspaceSidebar.svelte`, `Composer.svelte`, `MessageItem.svelte`, `SessionTabs.svelte`, and modal overlays.

### Task 2: Dynamic Per-Session Context Usage & Max Limits
- **Status**: `Completed`
- **Objective**: Ensure the context token indicator and popover dynamically reflect the active session and its specific model context limits.
- **Action Items**:
  - Calculate context token consumption per session in `session_scanner.go` (`GetSessionUsage`) and expose via Wails binding in `app.go`.
  - Bind reactive state in `session.svelte.ts` (`loadSessionUsage`) so switching tabs immediately updates token count, percentage, and limit.
  - Dynamically set max limit (e.g. 128k, 200k, 2M) based on the session's active model.

### Task 3: Native Context Compaction Routine
- **Status**: `Completed`
- **Objective**: Implement Grok CLI compaction mechanism (`CompactSession`) and wire to the "Compact conversation" button.
- **Action Items**:
  - Implement compaction runner in Go (`app.go` / `session_scanner.go`).
  - Wire compaction trigger in `ContextUsagePopover.svelte` through `sessionStore.compactActiveSession()`.
  - Reload session transcript and update token usage upon completion.

### Task 4: Visual & Operational Verification
- **Status**: `Pending`
- **Objective**: Verify UI quality and functionality using automated and manual inspection methods.
- **Action Items**:
  - Build application with `wails build` or `pnpm build`.
  - Activate window via AppleScript (`osascript -e 'tell application "aethergrok" to activate'`).
  - Capture verification screenshots using `screencapture -x`.

---

## 4. Work Boundaries (What Needs To Be Done vs Out of Scope)

### What Must Be Done (In Scope)
1. Full Ant Design Dark color calibration and border refinement across all views.
2. Dynamic per-session context token calculation and reactive UI updates.
3. Backend context compaction integration connected to the popover action button.
4. Git synchronization with `origin/main` on `fiko942/grok-build`.

### What Is Out of Scope (Not Required)
1. Replacing Svelte 5 runes with legacy Svelte 3/4 reactive syntax.
2. Introducing third-party CSS component frameworks (MUI, Chakra) outside Ant Design tokens and Tailwind CSS.
3. Rewriting the Go backend in another language.
4. Creating external web service dependencies when local file reading suffices.

---

## 5. Continuity Guide for Next Session
When continuing this project in a new session:
1. Inspect `docs/superpowers/plans/2026-09-25-aethergrok-master-plan-and-progress.md` for current system state.
2. Run `git pull origin main` to verify workspace synchronization.
3. Execute `dev.sh` or `wails dev` to start the development environment.
4. Proceed with Task 1 (Dark Theme Refinement) and Task 2 (Dynamic Context Tokens).
