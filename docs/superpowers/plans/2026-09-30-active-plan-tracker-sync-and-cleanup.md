# Investigation and Plan: Active Execution Plan & Tracker Synchronization

## 1. Problem Analysis & Root Cause

### Issue 1: Floating "Active Execution Plan" widget does not auto-collapse / hide when Right Sidebar "Plan" tab is open
- **Symptom:** When the user opens the right inspector panel and switches to the "Plan" tab, the floating card `FloatingPlanTracker.svelte` on the top right of the chat area remains open/expanded, creating redundancy and visual clutter.
- **Root Cause:** In `FloatingPlanTracker.svelte` and `App.svelte`, `FloatingPlanTracker` only checks `planStore.isFloatingVisible` without checking if the active session's right sidebar is currently open and active on the `'plan'` tab (`sessionStore.activeSession?.rightSidebarOpen && sessionStore.activeSession?.rightSidebarTab === 'plan'`).
- **Solution:** In `FloatingPlanTracker.svelte`, check if `isSidebarPlanActive` (`sessionStore.activeSession?.rightSidebarOpen && sessionStore.activeSession?.rightSidebarTab === 'plan'`). When sidebar plan tab is open, automatically hide the floating widget (or collapse it). When user closes the sidebar or switches away from the 'plan' tab, restore floating widget visibility.

---

### Issue 2: Accumulated/Duplicate Todo Items when AI generates new plans in subsequent prompts
- **Symptom:** When the user sends a new prompt, `planStore.resetPlan(sessionId)` clears the state in frontend memory. However, when the AI agent subsequently writes a new plan using `todo_write` or returns `TodosUpdated`, old/previous tasks appear again along with the new tasks, or when `todo_write` with `{ merge: false }` or a fresh task set is emitted, the parser / store handling merges or retains old todos in unexpected formats.
- **Root Cause Investigation:**
  1. Grok CLI / Claude Agent `todo_write` tool has parameter `merge: boolean` (default `true`). When `merge: false`, it provides the complete fresh todo list.
  2. In `PlanStore.inspectToolForPlan`, `parseTodosUpdated` is called for both `params` and `result`.
  3. When Grok CLI finishes a tool call `todo_write`, the `result` string returned by the CLI often looks like:
     ```
     - [completed] task 1: ...
     - [in_progress] task 2: ...
     - [pending] task 3: ...
     ```
     or the CLI returns `{ "result": "- [completed] ...\n- [in_progress] ..." }`.
  4. When `todo_write` is invoked in a second prompt in the same session, if the model passes `todos: [...]`, but `parseTodosUpdated` doesn't handle full replace vs merge, or parses previous todos from raw session message reconstructions, old todos are retained or duplicated.
  5. Furthermore, in `SessionStore`, when a session is loaded from disk or created, how are messages and plan states persisted or loaded? `planStore` only stores `sessionPlans` in memory, and `resetPlan(sessionId)` clears `sessionPlans[sessionId]`.
  6. But if `App.svelte` reconstructs or if tool calls are replayed, or if `parseTodosUpdated` does not handle raw markdown list syntax `- [ ] ...` / `- [x] ...` vs `todo_write` parameter replacement correctly:
     Let's verify how `parseTodosUpdated` parses markdown checkboxes and `todo_write` inputs.

---

## 2. Implementation Steps

1. **Fix Floating Widget Visibility with Sidebar Synchronization:**
   - In `FloatingPlanTracker.svelte`:
     Derive `isSidebarPlanOpen = $derived(sessionStore.activeSession?.rightSidebarOpen && sessionStore.activeSession?.rightSidebarTab === 'plan')`.
     Condition the visibility on `!isSidebarPlanOpen`.
2. **Fix Plan Parser & Store Replace / Clean Ingestion:**
   - Update `parseTodosUpdated` to support:
     a) Markdown checkbox lists (`- [ ]`, `- [x]`, `- [/]`, `- [in_progress]`) if tool output is formatted as markdown todos.
     b) Clean handling of `todo_write` params (`todos`, `merge`). When a tool call is `todo_write` and `merge: false` (or fresh plan with all new IDs), replace the existing todo list cleanly instead of accumulating.
     c) Deduplicate todos by item `id` or normalized `content` to prevent duplicated/stale task entries.
3. **Write Unit Tests for `planParser` and `planStore`:**
   - Create `frontend/src/lib/utils/planParser.test.ts` testing:
     - Extraction from `todo_write` input (`{ todos: [...] }`).
     - Ingestion of markdown checkbox lists.
     - Reset and clean replace behavior without duplicate accumulation.
4. **Build, Test, and Deploy:**
   - Run `pnpm --dir frontend test`.
   - Run `pnpm --dir frontend build`.
   - Build macOS DMG and deploy to `/Applications/AetherGrok.app`.
