# Command Description & In/Out Inspector with Window Stability Plan

> **For agentic workers:** Implemented under superpowers:systematic-debugging & superpowers:writing-plans.

**Goal:** Provide human-readable command descriptions and structured `IN`/`OUT` panels for terminal tool execution matching Grok/Codex CLI standards, while eliminating spontaneous application hiding/disappearance on macOS.

**Architecture:**
1. **Window Stability & Fail-Safe Watchdog (`app.go`, `frontend/src/App.svelte`):**
   - Prevent background disappearance caused by delayed `screencapture` execution via a 1500ms fail-safe watchdog on `wailsWindowController.Hide()`, ensuring `WindowShow()` is always triggered.
   - Intercept macOS `Cmd+H` (`[NSApp hide:]`) and reinforce `Cmd+W` in `App.svelte` so accidental keypresses do not hide the window to Dock.
2. **Human-Readable Command Descriptions (`ToolCallCard.svelte`):**
   - Extract `description` (or `summary`/`reason`) from `run_terminal_command` and subagent tool payloads to replace raw truncated CLI scripts in the header with clear intent (e.g. `Run: extract task brief for Task 2`).
3. **Structured `IN` (Input) & `OUT` (Output) Panels (`ToolCallCard.svelte`):**
   - Dedicated `IN` panel displaying full command multi-line scripts with a one-click `Copy IN` button.
   - Dedicated `OUT` panel displaying exit codes, duration, and ANSI-cleaned output with a one-click `Copy OUT` button.
   - Keyboard shortcuts (`c` to copy output, `i` or `Shift+C` to copy command) when a card is expanded and focused.

**Tech Stack:** Go 1.24, Wails v2, Svelte 5 (Runes), TypeScript, Bun/Vitest, Tailwind CSS, Ant Design Dark tokens.

---

### Implementation Verification

- [x] **Backend watchdog**: 1500ms safety timer in `app.go` guaranteeing `WindowShow()` restoration.
- [x] **Keydown interception**: Intercepted `Cmd+H` and sealed `Cmd+W` propagation in `frontend/src/App.svelte`.
- [x] **Tool parsing**: Prioritized `p.description` in `ToolCallCard.svelte` for `run_terminal_command` & subagent output.
- [x] **Structured inspector**: Rendered separate `IN` and `OUT` sections with individual copy affordances.
- [x] **Card shortcuts**: Supported `c` (copy OUT) and `i` / `Shift+C` (copy IN) keyboard accelerators.
- [x] **Unit tests**: `test/toolUtils.test.ts` and `test/shortcut-window-close.test.ts` passing 100%.
- [x] **Production build**: Built and deployed to `/Applications/AetherGrok.app`.
