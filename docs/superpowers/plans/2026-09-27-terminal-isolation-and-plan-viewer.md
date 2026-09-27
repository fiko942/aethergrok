# Plan: Session-Isolated Multi-Tab Terminal & Expandable Plan Viewer

## Tasks

- [ ] Task 1: Go Backend Terminal Manager (`pkg/terminal`)
  - Create `pkg/terminal/terminal.go` and `terminal_test.go`
  - Implement PTY allocation and process group termination (`-pgid`)
  - Expose Wails methods in `app.go`:
    - `CreateTerminal(sessionId, termId, cwd, shell string) error`
    - `WriteTerminal(termId, data string) error`
    - `ResizeTerminal(termId, cols, rows int) error`
    - `CloseTerminal(termId string) error`
    - `CloseSessionTerminals(sessionId string) error`
    - `GetPlanContent(planPath string) (string, error)`

- [ ] Task 2: Expandable Plan Markdown Viewer
  - Update `ToolCallCard.svelte` to fetch plan content or load file content when `toolParsed.type === 'plan_enter'` or `'plan_exit'` is expanded
  - Render with `renderMarkdown` / formatted markdown container

- [ ] Task 3: Terminal Frontend Components & Store
  - Create `frontend/src/lib/stores/terminal.svelte.ts`
  - Create `frontend/src/lib/components/terminal/TerminalPanel.svelte`
  - Implement Xterm / ANSI canvas renderer with multi-tab support
  - Implement Drag-to-resize divider
  - Implement "Send to Agent" button which injects terminal output into active prompt composer

- [ ] Task 4: App Integration & Session Binding
  - Mount `TerminalPanel` in `App.svelte` below chat message view
  - Ensure switching tabs preserves background terminal execution per session
  - Clean up terminals on session tab close

- [ ] Task 5: Verification & Testing
  - Run `go test ./...`
  - Run frontend build `npm run build`
  - Verify process group cleanup
  - Commit & Push to Git
