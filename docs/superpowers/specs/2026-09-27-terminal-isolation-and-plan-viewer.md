# Spec: Tab-Isolated Multi-Tab Terminal, Cross-Platform Process Management & Plan Viewer

## Summary
Implement a high-performance, session-isolated, multi-tab terminal inside AetherGrok with guaranteed total process cleanup (process group killing on macOS/Linux and process tree termination on Windows), shell login environment inheritance, terminal log attachment into agent prompt composer, and an expandable interactive Markdown viewer for `enter_plan_mode` / `exit_plan_mode` tool calls.

---

## Architecture & Implementation

### 1. Backend PTY & Cross-Platform Process Isolation (`pkg/terminal`)
- **Unix & macOS (`pkg/terminal/pty_unix.go`)**:
  - Uses `github.com/creack/pty` for native pseudo-terminal allocation.
  - Spawns shell processes inside dedicated Process Groups (`syscall.Setpgid(pid, pid)`).
  - Termination sequence on tab close / kill:
    1. Sends `syscall.Kill(-pgid, syscall.SIGTERM)` to the entire process group.
    2. Waits for a 150ms grace period.
    3. Escalates to `syscall.Kill(-pgid, syscall.SIGKILL)` to guarantee 0 orphaned processes (e.g. Vite dev server, Node daemons, Electron instances, file watchers).
- **Windows (`pkg/terminal/pty_windows.go`)**:
  - Creates dedicated process groups via `CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP`.
  - Termination uses native `taskkill /T /F /PID <pid>` to forcefully close all descendant process trees without dangling tasks.
- **Cross-Platform I/O Wrapper (`pkg/terminal/terminal.go`)**:
  - `osFileWrapper` encapsulates `io.ReadCloser`, `io.WriteCloser`, and `*os.File` for uniform PTY streaming across all supported operating systems.

### 2. Environment & PATH Resolution
- Spawns login shells (`shell, "-l"` on macOS/Unix) to ensure user configuration files (`.zprofile`, `.zshrc`, `.bashrc`) are evaluated.
- Injects standard development tool paths into `PATH` for macOS GUI desktop launches:
  - `~/.local/bin` (pnpm, node, grok CLI)
  - `~/Library/pnpm`
  - `/opt/homebrew/bin` & `/opt/homebrew/sbin`
  - `/usr/local/bin`
  - `~/go/bin` & `~/.cargo/bin`

### 3. Frontend Multi-Tab Terminal & UI/UX (`TerminalPanel.svelte`, `terminal.svelte.ts`)
- **Session Isolation**:
  - `terminalStore.svelte.ts` tracks sub-terminals per session: `sessionTerminals: Record<sessionId, TerminalTab[]>`.
  - Preserves terminal states when switching session tabs at the top navigation bar.
  - Auto-terminates session terminals via `CloseSessionTerminals(sessionId)` when a session tab is closed.
- **Visual Styling & xterm.js Integration**:
  - Imported `@xterm/xterm/css/xterm.css` globally in `app.css` to properly position and hide helper textareas/IME overlays.
  - Dark terminal theme matching AetherGrok palette (`#0e0e11` background, `#38bdf8` cyan cursor, ANSI color palette).
  - Smooth vertical drag handle for panel height resizing.
  - Maximize / Restore mode (`calc(100vh - 120px)`).
  - Sub-tab bar with active indicator, close buttons, and `+` add terminal button.
  - "Send to Agent" button: captures terminal selection (or last 80 lines of output) and formats it into the agent composer textarea without intrusive white borders.

### 4. Plan Markdown Viewer (`ToolCallCard.svelte`, `app.go`)
- Backend `GetPlanContent(planPath string)` securely reads plan Markdown files from disk.
- `ToolCallCard.svelte` detects `enter_plan_mode` and `exit_plan_mode` actions.
- Automatically fetches and displays the formatted Markdown plan inside the tool call card upon expansion, with syntax highlighting, checklist checkboxes, and clean typography.
