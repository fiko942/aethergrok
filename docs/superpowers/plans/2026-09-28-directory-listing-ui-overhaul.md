# Directory Listing UI Overhaul Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Overhaul the directory listing (`list_dir` / `ls`) tool execution card UI so that it organizes directories and files cleanly into categorized, filterable, and collapsed groups with search/filter, item counts, and distinct badges rather than dumping all raw paths in an unorganized grid.

**Architecture:** Split items into folders vs files, provide category tab toggles (All, Folders, Files), show concise summary metrics, sort folders first, support inline instant filter, and show relative names cleanly with expandable preview.

**Tech Stack:** Svelte 5 runes (`$state`, `$derived`), Tailwind CSS, Lucide icons, Ant Design tokens.

## Global Constraints
- Must look crisp and high-contrast in both Light mode and Dark mode.
- Avoid raw dumps; group folders and files with visual hierarchy.
- No em-dashes anywhere in UI copy.
- Full type safety and build verification (`npm run build --prefix frontend`).

---

### Task 1: Refactor `dirListingItems` and create structured categories in `ToolCallCard.svelte`

**Files:**
- Modify: `frontend/src/lib/components/chat/ToolCallCard.svelte`

- [ ] **Step 1: Update directory listing item parser to extract clean relative names, path types, and extensions**
- [ ] **Step 2: Add directory filter state (search filter, tab: All / Folders / Files, collapse toggle if > 12 items)**
- [ ] **Step 3: Build organized UI with top summary bar, folder chips, file chips, and count badges**
- [ ] **Step 4: Verify build with `npm run build --prefix frontend`**
- [ ] **Step 5: Review visual contrast in Light and Dark modes**
