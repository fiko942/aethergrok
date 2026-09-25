# Master System Context & Session Handoff Specification

## Overview
This specification details the complete architectural state, file mappings, reactive stores, and execution plan for continuing development of AetherGrok Desktop (`grok-desktop`).

## File Map & Responsibilities

| File Path | Description | Key State & Responsibilities |
| :--- | :--- | :--- |
| `frontend/src/app.css` | Global styling & Ant Design tokens | Defines CSS variables for light and dark themes (`--ant-bg-*`, `--ant-border`, `--ant-text-*`). |
| `frontend/src/lib/antd/tokens.ts` | Theme token definitions | TypeScript token definitions for light and dark themes. |
| `frontend/src/lib/stores/session.svelte.ts` | Session reactive store | Manages active sessions, tab lists, session history, and token counts. |
| `frontend/src/lib/components/chat/Composer.svelte` | Input area | Houses model selection, effort dropdown, agent mode, and context pill. |
| `frontend/src/lib/components/chat/composer/ContextUsagePopover.svelte` | Context popover | Displays token donut gauge, usage limits, and compaction action. |
| `frontend/src/lib/components/layout/WorkspaceSidebar.svelte` | Navigation sidebar | Displays workspace folders, pinned sessions, and session switching. |
| `app.go` & `session_scanner.go` | Go backend services | Scans sessions on disk, reads transcripts, calculates tokens, and runs compaction. |

## Detailed Plan for Remaining Work

### 1. Ant Design Dark Theme Unification
- Ensure zero high-contrast harsh white borders exist in dark mode.
- Use `--ant-border: #303030` and `--ant-border-secondary: #222222`.
- Apply Ant Design Dark blue `#177ddc` for active selections.

### 2. Reactive Session Token Usage
- Calculate token count per turn from `~/.grok/sessions/<workspace_hash>/<session_id>/`.
- Pass per-session token count to the Svelte 5 store.
- Update donut gauge percentage dynamically when active tab changes.

### 3. Context Compaction Routine
- Call Grok session compaction logic from Go backend.
- Write segment summaries to `compaction/segment_*.md` and update `compaction/INDEX.md`.
- Reload active session in UI after compaction finishes.

### 4. Verification & Testing
- Automated and visual verification on macOS using AppleScript and `screencapture`.
- Validate that all changes compile cleanly and render consistently.
