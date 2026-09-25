# Right Sidebar (Workspace Explorer & Git Inspector) Specification

## Problem Statement
Users working with multiple workspace sessions need immediate access to workspace project structure and version control state without leaving the application. Specific requirements:
1. **Isolated Session Visibility**: Opening the right sidebar in Workspace Tab A must keep it open only in Tab A. Switching to Workspace Tab B shows Tab B's sidebar state (closed by default). Returning to Tab A restores its open state.
2. **Collapsible & Resizable**: Default state is closed. A drag handle on the left edge allows resizing clamped between 260px and 600px (default 340px). Double-clicking resets to 340px, and dragging under 180px collapses the sidebar.
3. **Files & Folders Explorer**:
   - Recursive tree view with distinct, high-quality icons for files and folders.
   - Folders expand and collapse.
   - Lazy loading of directory children: large directories like `node_modules` fetch data on demand with a spinner indicator, preventing frontend freezes or unresponsive UI.
   - Fast debounced file/folder search bar at the top of the tab.
   - Read-only file viewer (modal/sheet) with syntax highlighting for code files (JSON, Shell, TS/JS, Go).
   - Dual-mode viewer for Markdown (`.md`): Raw code vs Rendered preview with Anthropic serif font.
4. **Git Changes Inspector**:
   - Current git branch badge with clean/dirty status.
   - List of uncommitted changed files with status labels (`M`, `A`, `D`, `?`) and addition/deletion line metrics.
   - Action controls: Commit message input, Commit button, Commit & Push button, Pull button.
   - Visual diff inspection: Clicking any changed file opens an interactive line-by-line diff modal with green (+) for additions and red (-) for deletions.
5. **Theme & Style**:
   - Anthracite dark mode aesthetic matching existing Grok Desktop theme (`bg-[#18181b]`, `border-[#27272a]`, `text-[#e4e4e7]`, Anthropic serif font family).

## Architecture & Data Flow

### Backend (Go / Wails)
- Package `pkg/workspace`:
  - `ReadDirectory(workspacePath, relativeDir)`: Returns sorted list of `FileItem` (directories first, then files alphabetically). Excludes deep recursive walk to preserve memory.
  - `ReadFileContent(workspacePath, relativePath)`: Reads file content safely up to 1MB limit.
  - `GetGitStatus(workspacePath)`: Executes `git status --porcelain=v1 -b` and parses branch name and status entries.
  - `GetFileDiff(workspacePath, filePath)`: Executes `git diff HEAD -- <filePath>` or plain `git diff` for untracked files.
  - `ExecuteCommit(workspacePath, message)`: Executes `git add -A && git commit -m <message>`.
  - `ExecutePush(workspacePath)`: Executes `git push`.
  - `ExecutePull(workspacePath)`: Executes `git pull`.

### Frontend State (Svelte 5 Runes)
- `Session` object in `frontend/src/lib/stores/session.svelte.ts`:
  - `rightSidebarOpen?: boolean` (defaults to `false`)
  - `rightSidebarTab?: 'files' | 'changes'` (defaults to `'files'`)
- Methods:
  - `toggleRightSidebar(sessionId?: string)`
  - `setRightSidebarTab(tab, sessionId?: string)`
