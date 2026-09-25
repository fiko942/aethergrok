# Right Sidebar (Workspace File Explorer & Git Changes Inspector) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement a collapsible, resizable right sidebar isolated per session workspace tab, featuring an asynchronous lazy-loaded File & Folder Explorer with read-only file viewing (syntax coloring & markdown preview), debounced search, and a Git Changes Inspector with branch status, uncommitted files list, quick commit/push/pull, and visual diff preview.

**Architecture:**
- Backend Go package `pkg/workspace` exposes OS filesystem and git commands via Wails bindings (`ReadWorkspaceDirectory`, `ReadWorkspaceFileContent`, `GetWorkspaceGitStatus`, `GetWorkspaceFileDiff`, `CommitWorkspaceChanges`, `PushWorkspaceChanges`, `PullWorkspaceChanges`).
- Frontend session store maintains isolated right sidebar state (`rightSidebarOpen`, `rightSidebarTab`) per session ID so switching workspace tabs preserves individual visibility and selection.
- Svelte 5 components with Tailwind CSS: `RightSidebar.svelte` container with left drag divider (clamped 260px–600px), `FileExplorerTree.svelte` for on-demand asynchronous directory tree rendering and search filtering, `FileViewerModal.svelte` for read-only code syntax coloring and dual-mode markdown preview, and `GitChangesPanel.svelte` for git branch info, file status badges, commit actions, and interactive line-by-line diff viewing.

**Tech Stack:** Go (os, exec, path/filepath), Wails v2, Svelte 5 (Runes `$state`, `$derived`, `$effect`), TypeScript, Tailwind CSS, Lucide icons (`Folder`, `FolderOpen`, `File`, `FileCode`, `GitBranch`, `GitCommit`, `GitPullRequest`, `Upload`, `Download`, `Search`, `Eye`, `Code`, `ChevronRight`, `ChevronDown`, `Plus`, `Minus`, `PanelRightClose`, `PanelRightOpen`, `RefreshCw`, `X`).

## Global Constraints
- Sidebar default state is closed (`rightSidebarOpen: false`).
- State is strictly isolated per workspace session tab: opening in Tab A does not open in Tab B; returning to Tab A restores its state.
- Resizable width clamped between 260px and 600px with a default width of 340px, double-click reset to 340px, and snap-to-close under 180px.
- Folders load asynchronously on demand (lazy expansion) with clear spinner loading indicators to maintain a responsive UI even in massive folders such as `node_modules` or `.git`.
- File content viewing is strictly read-only with a 1MB file size safety cap.
- Markdown viewer supports dual-mode: Raw code view and rendered preview.
- Git panel displays branch name, clean/dirty status, uncommitted file list with additions/deletions, commit input with commit/push/pull actions, and interactive diff preview with green additions and red deletions.
- Dark mode theme matches existing anthracite/zinc palette (`bg-[#18181b]`, `border-[#27272a]`, `text-[#e4e4e7]`, Anthropic serif font accents).

---

### Task 1: Verify & Finalize Backend Workspace & Git Bridge in Go

**Files:**
- Existing: `pkg/workspace/workspace.go`
- Existing: `app.go`
- Test: `pkg/workspace/workspace_test.go`

**Interfaces:**
- Produces:
  - `ReadDirectory(workspacePath, relativeDir string) ([]FileItem, error)`
  - `ReadFileContent(workspacePath, relativePath string) (string, error)`
  - `GetGitStatus(workspacePath string) (*GitStatusResult, error)`
  - `GetFileDiff(workspacePath, filePath string) (string, error)`
  - `ExecuteCommit(workspacePath, message string) error`
  - `ExecutePush(workspacePath string) (string, error)`
  - `ExecutePull(workspacePath string) (string, error)`

- [ ] **Step 1: Write backend unit tests in `pkg/workspace/workspace_test.go`**
Verify directory reading, file content retrieval with size safety, git status error handling, and diff extraction.

- [ ] **Step 2: Run test to verify passes**
Run: `go test -v ./pkg/workspace`
Expected: PASS

- [ ] **Step 3: Commit backend workspace bridge**
```bash
git add pkg/workspace/ app.go
git commit -m "feat(backend): add workspace filesystem and git operations bridge"
```

---

### Task 2: Session Store Right Sidebar State Isolation

**Files:**
- Existing: `frontend/src/lib/stores/session.svelte.ts`

**Interfaces:**
- Produces:
  - `Session.rightSidebarOpen: boolean`
  - `Session.rightSidebarTab: 'files' | 'changes'`
  - `sessionStore.toggleRightSidebar(sessionId?: string): void`
  - `sessionStore.setRightSidebarTab(tab: 'files' | 'changes', sessionId?: string): void`

- [ ] **Step 1: Verify `Session` interface and store methods**
Ensure `rightSidebarOpen` defaults to `false` and `rightSidebarTab` defaults to `'files'` for all newly created or restored sessions.

- [ ] **Step 2: Commit session store state isolation**
```bash
git add frontend/src/lib/stores/session.svelte.ts
git commit -m "feat(store): add per-session right sidebar open and active tab state"
```

---

### Task 3: File Explorer Tree Component (`FileExplorerTree.svelte`)

**Files:**
- Create: `frontend/src/lib/components/workspace/FileExplorerTree.svelte`
- Create: `frontend/src/lib/components/workspace/FileIcon.svelte`

**Interfaces:**
- Consumes:
  - `ReadWorkspaceDirectory(workspacePath, relativeDir)` from Wails runtime
  - `activeSession.workspacePath` from `sessionStore`
- Produces:
  - Event `selectFile(filePath: string)` when user clicks on a file.

- [ ] **Step 1: Implement `FileIcon.svelte`**
Renders tailored icons based on extension:
- `.json`, `.js`, `.ts`, `.svelte`, `.go`, `.sh`, `.md`, `.html`, `.css`, etc.
- Folder open/closed states.

- [ ] **Step 2: Implement `FileExplorerTree.svelte`**
Features:
- Recursive tree structure with level indentation (`padding-left: depth * 14px + 8px`).
- Asynchronous lazy loading: children of a folder are requested from backend only upon clicking expand.
- Spinner indicator while directory read is in progress, preventing UI freezes on directories like `node_modules`.
- Search input at the top with 200ms debounce that filters visible tree nodes or displays matched file paths.
- Refresh button to reload the directory structure on demand.

- [ ] **Step 3: Commit File Explorer components**
```bash
git add frontend/src/lib/components/workspace/FileIcon.svelte frontend/src/lib/components/workspace/FileExplorerTree.svelte
git commit -m "feat(workspace): add lazy-loaded FileExplorerTree and FileIcon components"
```

---

### Task 4: Read-Only File Viewer Modal / Sheet (`FileViewerModal.svelte`)

**Files:**
- Create: `frontend/src/lib/components/workspace/FileViewerModal.svelte`

**Interfaces:**
- Consumes:
  - `ReadWorkspaceFileContent(workspacePath, relativePath)` from Wails runtime
  - Prop `filePath: string`
  - Prop `isOpen: boolean`
  - Prop `onClose: () => void`

- [ ] **Step 1: Implement `FileViewerModal.svelte`**
Features:
- Clean overlay or sliding side sheet displaying file name, path, and size.
- Strictly read-only view with line numbers.
- Syntax highlighting tokens for code files (JSON keys/strings/numbers, Shell comments/commands, Go keywords/strings, TypeScript/JavaScript keywords/strings).
- For Markdown (`.md`) files: Mode toggle button between **Raw Code** and **Rendered Preview** (with Anthropic serif font typography and styled headings/lists/tables).
- Close button and Escape key listener.

- [ ] **Step 2: Commit File Viewer component**
```bash
git add frontend/src/lib/components/workspace/FileViewerModal.svelte
git commit -m "feat(workspace): add read-only FileViewerModal with syntax coloring and dual-mode markdown preview"
```

---

### Task 5: Git Changes & Visual Diff Panel (`GitChangesPanel.svelte`)

**Files:**
- Create: `frontend/src/lib/components/workspace/GitChangesPanel.svelte`
- Create: `frontend/src/lib/components/workspace/DiffViewModal.svelte`

**Interfaces:**
- Consumes:
  - `GetWorkspaceGitStatus(workspacePath)`
  - `GetWorkspaceFileDiff(workspacePath, filePath)`
  - `CommitWorkspaceChanges(workspacePath, message)`
  - `PushWorkspaceChanges(workspacePath)`
  - `PullWorkspaceChanges(workspacePath)`

- [ ] **Step 1: Implement `DiffViewModal.svelte`**
Interactive modal rendering side-by-side or unified line-by-line diff:
- Green background for addition lines (`+`).
- Red background for deletion lines (`-`).
- Gray/neutral for context lines.
- File header with additions count (`+N`) and deletions count (`-N`).

- [ ] **Step 2: Implement `GitChangesPanel.svelte`**
Features:
- Top branch indicator badge: branch name with clean/dirty status badge.
- List of changed files with status letters (`M` Modified, `A` Added, `D` Deleted, `?` Untracked).
- Clicking any changed file opens `DiffViewModal` with the git diff output.
- Commit section: commit message input textarea/field with:
  - `Commit` button
  - `Commit & Push` button
  - `Pull` button
- Status notifications for git operations (success / error alerts).

- [ ] **Step 3: Commit Git Changes Panel**
```bash
git add frontend/src/lib/components/workspace/GitChangesPanel.svelte frontend/src/lib/components/workspace/DiffViewModal.svelte
git commit -m "feat(workspace): add GitChangesPanel with commit/push/pull actions and DiffViewModal"
```

---

### Task 6: Resizable Right Sidebar Container (`RightSidebar.svelte`) & App Integration

**Files:**
- Create: `frontend/src/lib/components/layout/RightSidebar.svelte`
- Modify: `frontend/src/App.svelte`
- Modify: `frontend/src/lib/components/layout/SessionTabs.svelte` or App Header

**Interfaces:**
- Consumes:
  - `sessionStore.activeSession` (for workspace path and isolated `rightSidebarOpen` / `rightSidebarTab`)
- Produces:
  - Right sidebar toggle button in the header (`PanelRightClose` / `PanelRightOpen`)
  - Shortcut listener `⌘⌥B` (Mac) / `Ctrl+Alt+B` (Windows/Linux)

- [ ] **Step 1: Implement `RightSidebar.svelte`**
Features:
- Draggable resize handle on the left edge.
- Width clamped between 260px and 600px (persisted default: 340px).
- Snap-to-close if dragged narrower than 180px.
- Double click handle resets to default 340px.
- Tab bar at top:
  - Tab 1: **Files** (with folder icon)
  - Tab 2: **Changes** (with git-branch/git-commit icon and change count badge)
- Body switches between `FileExplorerTree` and `GitChangesPanel`.

- [ ] **Step 2: Integrate into `frontend/src/App.svelte`**
- Mount `RightSidebar` in the main workspace view alongside the chat panel.
- Ensure sidebar visibility is driven strictly by `sessionStore.activeSession?.rightSidebarOpen`.
- Add right sidebar toggle icon button to top header bar.
- Add global keyboard shortcut `⌘⌥B` to toggle sidebar for the current active session.

- [ ] **Step 3: Commit layout and integration**
```bash
git add frontend/src/lib/components/layout/RightSidebar.svelte frontend/src/App.svelte
git commit -m "feat(layout): integrate resizable RightSidebar with session isolation and header toggle"
```

---

### Task 7: Full Verification, Documentation & Git Push

**Files:**
- `docs/superpowers/plans/2026-09-25-right-sidebar-explorer-git.md`
- `docs/superpowers/specs/2026-09-25-right-sidebar-spec.md`

- [ ] **Step 1: Run frontend checks & tests**
```bash
cd frontend && npm run check && npm run build && cd ..
go test ./...
```
Expected: Clean pass with 0 errors.

- [ ] **Step 2: Document architectural specifications in `docs/superpowers/`**
Write detailed spec and plan documents for future reference.

- [ ] **Step 3: Stage and commit all files to git and push to GitHub**
```bash
git add .
git commit -m "feat: complete collapsible and resizable right sidebar with lazy file explorer and git inspector"
git push origin main
```
