# Global Chat Session Drag & Drop Overlay & Duplicate Fix Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Provide a responsive, animated drag-and-drop overlay across the entire active chat session viewport and fix the double attachment bug when files are dropped into the session area or composer.

**Architecture:**
1. **Root Cause Analysis of Double Attach:**
   - When dropping on Composer container (which is inside `<main>`), the drop event on Composer ran `handleContainerDrop` (`processFiles(e.dataTransfer.files)`) and then bubbled up to `<main>` in `App.svelte`, triggering `composerRef.handleExternalFiles` (`processFiles(e.dataTransfer.files)`) a second time.
   - Calling `e.stopPropagation()` in `Composer.svelte`'s `handleContainerDrop`, `handleContainerDragEnter`, `handleContainerDragOver`, and `handleContainerDragLeave` prevents double triggering.
2. **Global Drag & Drop Visual Overlay on Chat Session:**
   - In `App.svelte`, add a reactive `isMainDragOver` state and `dragEnterCounter` on the `<main>` container.
   - When files are dragged anywhere over the open session / chat area, render a prominent, beautifully styled Ant Design Dark dashed overlay ("Drop files or images anywhere in session to attach") with backdrop blur and icon.
   - Reset `isMainDragOver` properly on `dragleave` and `drop`.
3. **Execution & Verification:**
   - Run typecheck, frontend build, and Go test suite. Commit changes cleanly.

**Tech Stack:** Svelte 5 Runes, TypeScript, Tailwind CSS, Lucide Svelte.

---

### Task 1: Fix Drag Event Bubbling in `Composer.svelte`

**Files:**
- Modify: `frontend/src/lib/components/chat/Composer.svelte`

- [ ] **Step 1: Add `e.stopPropagation()` to Composer drag/drop handlers**

In `Composer.svelte`:
```typescript
function handleContainerDragEnter(e: DragEvent) {
  if (e.dataTransfer?.types?.includes('Files')) {
    e.preventDefault();
    e.stopPropagation();
    dragCounter++;
    isDragOver = true;
  }
}

function handleContainerDragOver(e: DragEvent) {
  if (e.dataTransfer?.types?.includes('Files')) {
    e.preventDefault();
    e.stopPropagation();
    if (e.dataTransfer) e.dataTransfer.dropEffect = 'copy';
    isDragOver = true;
  }
}

function handleContainerDragLeave(e: DragEvent) {
  e.preventDefault();
  e.stopPropagation();
  dragCounter--;
  if (dragCounter <= 0) {
    dragCounter = 0;
    isDragOver = false;
  }
}

function handleContainerDrop(e: DragEvent) {
  e.preventDefault();
  e.stopPropagation();
  dragCounter = 0;
  isDragOver = false;
  if (e.dataTransfer?.files && e.dataTransfer.files.length > 0) {
    processFiles(e.dataTransfer.files);
  }
}
```

- [ ] **Step 2: Verify `pnpm run check` in frontend**

---

### Task 2: Implement Chat Session Drag-and-Drop Overlay in `App.svelte`

**Files:**
- Modify: `frontend/src/App.svelte`

- [ ] **Step 1: Add `isSessionDragOver` state and drag event tracking with overlay in `App.svelte`**

Track drag enter counter and state in `App.svelte`:
```svelte
let isSessionDragOver = $state(false);
let sessionDragCounter = 0;

function handleMainDragEnter(e: DragEvent) {
  if (e.dataTransfer?.types?.includes('Files')) {
    e.preventDefault();
    sessionDragCounter++;
    isSessionDragOver = true;
  }
}

function handleMainDragOver(e: DragEvent) {
  if (e.dataTransfer?.types?.includes('Files')) {
    e.preventDefault();
    if (e.dataTransfer) e.dataTransfer.dropEffect = 'copy';
    isSessionDragOver = true;
  }
}

function handleMainDragLeave(e: DragEvent) {
  e.preventDefault();
  sessionDragCounter--;
  if (sessionDragCounter <= 0) {
    sessionDragCounter = 0;
    isSessionDragOver = false;
  }
}

function handleMainDrop(e: DragEvent) {
  e.preventDefault();
  sessionDragCounter = 0;
  isSessionDragOver = false;
  if (e.dataTransfer?.files && e.dataTransfer.files.length > 0) {
    composerRef?.handleExternalFiles?.(e.dataTransfer.files);
  }
}
```

Render visual overlay inside `<main>`:
```svelte
{#if isSessionDragOver}
  <div class="absolute inset-0 z-50 bg-ant-bg/90 border-2 border-dashed border-ant-primary/60 rounded-xl m-2 flex flex-col items-center justify-center space-y-3 backdrop-blur-md pointer-events-none animate-in fade-in zoom-in-95 duration-150 select-none">
    <div class="w-14 h-14 rounded-2xl bg-ant-primary/15 text-ant-primary flex items-center justify-center shadow-lg border border-ant-primary/20">
      <Upload size={28} />
    </div>
    <div class="text-center space-y-1">
      <p class="text-base font-serif font-semibold text-ant-text">
        Drop files or images into session
      </p>
      <p class="text-xs font-mono text-ant-text-secondary">
        Attach images, markdown, PDFs, or source code to prompt
      </p>
    </div>
  </div>
{/if}
```

- [ ] **Step 2: Verify `pnpm --prefix frontend run check` and `pnpm --prefix frontend run build`**

---

### Task 3: Full Verification & Commit

- [ ] **Step 1: Run Go tests and frontend checks**
- [ ] **Step 2: Commit all changes**
