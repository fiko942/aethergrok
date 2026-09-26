# Plan: Zero-Tab Closing, Cmd/Ctrl+W Shortcut, Empty Workspace State & Background Session Execution

## Overview
Allow users to close all session tabs down to zero (without forcing a fallback session to reopen), close the active tab using `⌘W` / `Ctrl+W`, display an elegant Ant Design Dark empty state in the center workspace when 0 tabs are open, and guarantee continuous background prompt/tool execution and queue processing for detached/closed session tabs.

## Objectives
1. **Zero-Tab Closing Support**: Update `sessionStore.closeSessionTab` so closing the last tab sets `openTabSessionIds` to empty and `activeSessionId` to `null` instead of auto-creating or re-opening a tab.
2. **`⌘W` / `Ctrl+W` Tab Close Shortcut**: Add global keyboard listener for `⌘W` (macOS) and `Ctrl+W` (Windows/Linux) to close the active session tab.
3. **Graceful Center Empty State**: When no session tab is open (`!sessionStore.activeSession`), display a dark-studio empty canvas with quick actions (New Session `⌘T`, open recent session from sidebar, snapshot `⌘⇧S`) and hide/disable the active chat composer cleanly.
4. **Persistent Background Execution**: Ensure streaming deltas, tool executions, auto-permission bypasses, turn completions, and queued prompt dispatching continue seamlessly for any session regardless of whether its tab is open in view.

## Implementation Tasks
- **Task 1: Update Session Store (`session.svelte.ts`)**
  - Allow `closeSessionTab` to remove the tab and set `activeSessionId = null` if no tabs remain in the workspace.
  - Update `closeSession` (full deletion) to handle zero-tab state cleanly.
- **Task 2: Add Global `⌘W` / `Ctrl+W` Shortcut (`App.svelte`)**
  - Register `isMetaOrCtrl && e.key.toLowerCase() === 'w'` in `handleGlobalKeyDown`.
  - Prevent default browser window closure and invoke `sessionStore.closeSessionTab(sessionStore.activeSessionId)`.
- **Task 3: Center Workspace Zero-Tab Empty State (`App.svelte` & `SessionTabs.svelte`)**
  - Render an empty workspace view when `sessionStore.openWorkspaceTabs.length === 0` or `!sessionStore.activeSession`.
  - Provide a clean button "New Conversation" (`⌘T`), recent sessions quick-reopen list, and clear guidance.
- **Task 4: Background Execution Verification & Session Isolation**
  - Verify Wails runtime events (`grok:delta_batch`, `grok:tool_call`, `grok:complete`, `checkAndDispatchNextQueue`) update target `sessionStore.sessions` by `sessionId` even when not in `openTabSessionIds`.
- **Task 5: Verification & Commit**
  - Run `pnpm run check`, `pnpm run build`, and `go test ./test/... -v`.
  - Commit changes to git.
