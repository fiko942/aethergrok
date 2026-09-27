# Plan: Terminal Isolation, Cross-Platform Process Cleanup, and Plan Viewer

## Execution Log & Status

- [x] **Backend PTY Engine (`pkg/terminal/terminal.go`)**:
  - Implemented `Manager` and `Instance` with PTY creation and lifecycle tracking.
  - Added Wails bindings: `CreateTerminal`, `WriteTerminal`, `ResizeTerminal`, `CloseTerminal`, `CloseSessionTerminals`.
- [x] **Cross-Platform Process Cleanup (`pkg/terminal/pty_unix.go`, `pkg/terminal/pty_windows.go`)**:
  - Unix/macOS: `syscall.Setpgid`, `syscall.Kill(-pgid, SIGTERM)` followed by `SIGKILL` to clean all descendants.
  - Windows: `CREATE_NEW_PROCESS_GROUP` and `taskkill /T /F /PID <pid>`.
  - Added cross-platform `osFileWrapper` for uniform I/O.
- [x] **Shell Login Environment & PATH Injection**:
  - Spawn login shell (`shell, "-l"`) on macOS/Unix.
  - Merge `/opt/homebrew/bin`, `~/.local/bin`, `~/Library/pnpm`, `/usr/local/bin`, and language bins into `PATH`.
  - Fix broken symlinks in `~/.local/bin/` pointing to outdated runtimes.
- [x] **Frontend Terminal Multi-Tab & Store (`TerminalPanel.svelte`, `terminal.svelte.ts`)**:
  - Per-session terminal tab isolation in Svelte store.
  - `@import '@xterm/xterm/css/xterm.css'` to properly hide xterm helper textarea and prevent visual text artifacts.
  - Minimalist styling without white borders for the "Send to Agent" button.
  - Resizable height handle, maximize toggle, and buffer clear action.
- [x] **Plan Markdown Viewer (`ToolCallCard.svelte`, `app.go`)**:
  - Added backend `GetPlanContent(planPath string)` binding.
  - Automatic markdown rendering on tool card expansion for `enter_plan_mode` and `exit_plan_mode`.
- [x] **Verification & Cross-Platform Builds**:
  - Go unit tests: `go test -count=1 -v ./...` (100% pass).
  - Wails cross-compilation: Windows (`aethergrok.exe`) and macOS (`aethergrok.app`).
  - Frontend production build (`vite build`): Succeeded.
