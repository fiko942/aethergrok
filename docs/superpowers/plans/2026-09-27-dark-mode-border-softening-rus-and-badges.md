# Dark Mode Border Softening & Border Cleanup Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Menghilangkan dan melembutkan border putih/terang yang mencolok pada mode gelap untuk: (1) garis pembatas di bawah workspace/folder "RUS" di sidebar, (2) tombol "Select" multi-select, (3) badge versi "v1.0.0" di header, dan (4) badge workspace "[RUS]" pada tab sesi, serta elemen badge serupa.

**Architecture:** Menerapkan styling token konsisten berbasis Ant Design Dark dan token Tailwind (`border-ant-border-secondary`, `border-ant-border/40`, atau `border-transparent`) serta varian `dark:border-white/5` yang harmonis. Pada mode gelap, border keras berwarna putih atau abu-abu terang digantikan dengan border beropasitas rendah atau disamakan dengan palet netral zinc/ant-border sehingga visual terasa seamless, menyatu, dan anti-slop.

**Tech Stack:** Svelte 5 (Runes), Tailwind CSS, Ant Design Dark Tokens, Wails v2.

## Global Constraints
- State points directly in affirmative language. Avoid unnecessary contrastive negation such as "X, not Y".
- Maintain complete dark and light mode compatibility without regressions.
- Keep layout geometry, keyboard shortcuts, and reactive stores intact.
- Verify browser build and TypeScript checks pass cleanly with zero errors.

---

### Task 1: Soften / Eliminate White Border on App Version Badge in Header

**Files:**
- Modify: `frontend/src/App.svelte:1135-1142`
- Test: Build and visual inspection in browser/Wails

**Interfaces:**
- Consumes: `__APP_VERSION__` global Vite variable
- Produces: Polished version tag with subtle dark-theme border matching header styling

- [ ] **Step 1: Inspect current styling in `App.svelte`**

Location: line 1139:
```svelte
<span class="px-1.5 py-0.5 text-[9px] font-mono font-medium bg-ant-bg-tertiary text-ant-text-muted rounded border border-ant-border/40 leading-none select-none">v{__APP_VERSION__}</span>
```

In dark mode, `border-ant-border/40` or explicit border can render with noticeable edge outline against dark header background. We replace it with `border border-transparent dark:border-white/[0.06] bg-ant-bg-tertiary/60` or `border border-ant-border-secondary` for a softened appearance.

- [ ] **Step 2: Update `App.svelte` version pill classes**

```svelte
<span class="px-1.5 py-0.5 text-[9px] font-mono font-medium bg-ant-bg-tertiary/70 text-ant-text-muted rounded border border-ant-border-secondary dark:border-white/5 leading-none select-none">v{__APP_VERSION__}</span>
```

- [ ] **Step 3: Verify build integrity**

Run: `npm run check` inside `frontend/`
Expected: 0 errors

---

### Task 2: Refine "Select" Multi-Select Button in Sidebar

**Files:**
- Modify: `frontend/src/lib/components/layout/WorkspaceSidebar.svelte:400-408`

**Interfaces:**
- Consumes: `sessionStore.isSelectionMode`
- Produces: Seamless "Select" toggle button without harsh border in dark mode

- [ ] **Step 1: Inspect current classes in `WorkspaceSidebar.svelte`**

Lines 400-406:
```svelte
<button
  onclick={() => sessionStore.toggleSelectionMode()}
  class="text-[11px] font-medium px-2 py-0.5 rounded transition border {sessionStore.isSelectionMode ? 'bg-ant-primary text-white border-ant-primary' : 'text-ant-text-secondary hover:text-ant-text hover:bg-ant-bg-tertiary border-ant-border/60'}"
  title="Toggle Multi-Select"
>
  {sessionStore.isSelectionMode ? 'Cancel' : 'Select'}
</button>
```

The unselected state uses `border-ant-border/60`, which on dark mode stands out as a visible light border box next to "Open Folder".

- [ ] **Step 2: Update button borders and background**

Replace with:
```svelte
<button
  onclick={() => sessionStore.toggleSelectionMode()}
  class="text-[11px] font-medium px-2 py-0.5 rounded transition border {sessionStore.isSelectionMode ? 'bg-ant-primary text-white border-ant-primary shadow-xs' : 'text-ant-text-secondary hover:text-ant-text bg-ant-bg-tertiary/40 hover:bg-ant-bg-tertiary border-ant-border-secondary dark:border-white/5'}"
  title="Toggle Multi-Select"
>
  {sessionStore.isSelectionMode ? 'Cancel' : 'Select'}
</button>
```

- [ ] **Step 3: Run svelte check**

Run: `cd frontend && npm run check`
Expected: PASS

---

### Task 3: Soften Workspace Tree Folder Border & Divider under Workspace Title

**Files:**
- Modify: `frontend/src/lib/components/layout/WorkspaceSidebar.svelte:502-518`

**Interfaces:**
- Consumes: `filteredWorkspaces`, `isExpanded`, `isMissing`
- Produces: Softened workspace container border and divider under workspace folder header (e.g. "RUS")

- [ ] **Step 1: Inspect current workspace container and header divider**

Line 502:
```svelte
<div class="rounded-lg {isMissing ? 'bg-rose-500/5 border border-rose-500/20' : 'bg-ant-bg border border-ant-border'} overflow-hidden shadow-2xs">
```
Line 515:
```svelte
class="flex items-center justify-between px-2.5 py-1.5 {isMissing ? 'bg-rose-500/10 cursor-not-allowed opacity-90' : 'bg-ant-bg-secondary hover:bg-ant-bg-tertiary cursor-pointer'} transition select-none group border-b {isExpanded ? 'border-ant-border/60' : 'border-transparent'}"
```

In dark mode, `border-ant-border` on the outer card and `border-b border-ant-border/60` directly underneath the workspace title creates the visible white/light line shown in the user's screenshot.

- [ ] **Step 2: Refine border tokens for outer card and inner divider**

Update outer container:
```svelte
<div class="rounded-lg {isMissing ? 'bg-rose-500/5 border border-rose-500/20' : 'bg-ant-bg border border-ant-border-secondary dark:border-white/5'} overflow-hidden shadow-2xs">
```

Update header bottom divider:
```svelte
class="flex items-center justify-between px-2.5 py-1.5 {isMissing ? 'bg-rose-500/10 cursor-not-allowed opacity-90' : 'bg-ant-bg-secondary hover:bg-ant-bg-tertiary cursor-pointer'} transition select-none group border-b {isExpanded ? 'border-ant-border-secondary dark:border-white/5' : 'border-transparent'}"
```

- [ ] **Step 3: Check search input and top filter container border**

Inspect `frontend/src/lib/components/layout/WorkspaceSidebar.svelte:377` and `385`:
Change `border-b border-ant-border` to `border-b border-ant-border-secondary dark:border-white/5` and input border to `border-ant-border-secondary dark:border-white/5` so the entire sidebar maintains subtle, consistent dark styling.

---

### Task 4: Soften Session Tab Workspace Badge "[RUS]"

**Files:**
- Modify: `frontend/src/lib/components/layout/SessionTabs.svelte:173-181`
- Modify: `frontend/src/lib/components/layout/NewSessionDropdown.svelte:111-114`

**Interfaces:**
- Consumes: `ws.name`, `sessionStore.sessions`
- Produces: Subtle, dark-harmonized workspace badge pill without harsh white outline

- [ ] **Step 1: Inspect `SessionTabs.svelte` workspace badge**

Line 174-179:
```svelte
<span
  class="flex-shrink-0 text-[9.5px] font-mono px-1 py-0.2 rounded bg-ant-bg-tertiary text-ant-text-muted border border-ant-border/50 max-w-[65px] truncate"
  title={`Workspace: ${ws?.name} (${ws?.path})`}
>
  {ws?.name}
</span>
```

The `border border-ant-border/50` gives it a light grey/white outline in dark mode.

- [ ] **Step 2: Update badge border in `SessionTabs.svelte`**

```svelte
<span
  class="flex-shrink-0 text-[9.5px] font-mono px-1 py-0.2 rounded bg-ant-bg-tertiary/80 text-ant-text-muted border border-ant-border-secondary dark:border-white/5 max-w-[65px] truncate"
  title={`Workspace: ${ws?.name} (${ws?.path})`}
>
  {ws?.name}
</span>
```

- [ ] **Step 3: Check `NewSessionDropdown.svelte` workspace badge**

Line 112:
```svelte
<span class="text-[10.5px] font-mono px-1.5 py-0.2 rounded bg-ant-bg-tertiary border border-ant-border/40 text-ant-text-muted">
```
Update to:
```svelte
<span class="text-[10.5px] font-mono px-1.5 py-0.2 rounded bg-ant-bg-tertiary/80 border border-ant-border-secondary dark:border-white/5 text-ant-text-muted">
```

---

### Task 5: Verify Build and Browser Rendering

**Files:**
- Validate: `frontend/src/App.svelte`
- Validate: `frontend/src/lib/components/layout/WorkspaceSidebar.svelte`
- Validate: `frontend/src/lib/components/layout/SessionTabs.svelte`
- Validate: `frontend/src/lib/components/layout/NewSessionDropdown.svelte`

- [ ] **Step 1: Run frontend typecheck**

Run: `cd frontend && npm run check`
Expected: 0 errors, 0 warnings

- [ ] **Step 2: Run frontend build**

Run: `cd frontend && npm run build`
Expected: Successful build producing `dist/`

- [ ] **Step 3: Run Go test suite**

Run: `go test ./test/...`
Expected: PASS
