# Spec: Tab-Isolated Multi-Tab Terminal & Expandable Plan Viewer

## Summary
Implement a high-performance, session-isolated, multi-tab terminal inside AetherGrok with guaranteed total process cleanup (process group killing) on macOS/Unix, terminal log attachment into agent prompt composer, and an expandable interactive Markdown viewer for `enter_plan_mode` / `exit_plan_mode` tool calls.

## Architecture

### 1. Process Group & PTY Architecture (`pkg/terminal`)
- Shell processes are spawned inside dedicated Process Groups (`syscall.SysProcAttr{Setpgid: true}`).
- When a terminal is closed, restarted, or interrupted with SIGINT/SIGKILL:
  1. `syscall.Kill(-pgid, syscall.SIGTERM)` is immediately issued to all descendants.
  2. A 200ms grace period is checked, then `syscall.Kill(-pgid, syscall.SIGKILL)` is dispatched to guarantee 0 orphaned processes (such as background Vite servers, Node daemons, or Electron child processes).
- Terminal instances are indexed by `(SessionID, TerminalID)`.
- Terminal data streams to the frontend via Wails event emission: `terminal:data:{termId}`.

### 2. Tab-Isolated Multi-Tab Terminal UI/UX
- Each Session in `sessionStore` maintains its own active list of sub-terminals (`terminals: Array<{ id, title, buffer, cwd }>`).
- Responsive terminal panel with:
  - Resizable height handle (drag divider).
  - Hide/Show toggle.
  - Multi-tab header with `+` button to add new shell instances.
  - "Send to Agent" action to format recent terminal output into composer.
  - "Clear" and "Restart/Kill" buttons.

### 3. Expandable Plan Markdown Viewer
- `ToolCallCard.svelte` detects `plan_enter` and `plan_exit` tool actions.
- When expanded, it retrieves the plan content from the session / disk and renders it with `markdownRenderer` instead of displaying a bare tool string.
