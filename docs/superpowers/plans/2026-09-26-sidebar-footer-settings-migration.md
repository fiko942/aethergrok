# Sidebar Footer Settings Migration & Settings Quick Access Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Clean up the left sidebar footer by removing the "Auto-Hide" and "Shutter Sound" inline toggles, moving their configuration cleanly into the Settings Modal (General & Smart Screen Snapshot tab) with persistent `settingsStore` backing, and adding a prominent "Settings" button in the bottom left sidebar footer next to the existing top-bar settings button.

**Architecture:**
1. Add `snapshotAutoHideWindow: boolean` to `settingsStore` with persistence and default `true`.
2. In `SettingsModal.svelte`, ensure both Auto-Hide Window and Shutter Sound have clear, well-described toggles in the General & Snapshot tab.
3. In `App.svelte`, remove the inline switches from the left sidebar footer and replace them with an elegant, Ant Design Dark styled "Settings" button (`SettingsModal` trigger) matching the look and feel of the desktop GUI studio.
4. Verify complete reactivity, typechecks, and build.

**Tech Stack:** Svelte 5 Runes, TypeScript, Tailwind CSS, Lucide Svelte, Go 1.24/Wails.

## Global Constraints
- Svelte 5 Runes (`$state`, `$derived`, `$props`).
- Keep existing settings button in top bar intact (now accessible in both top bar and bottom-left sidebar).
- Zero compilation errors (`pnpm --prefix frontend run check && pnpm --prefix frontend run build`).

---

### Task 1: Add `snapshotAutoHideWindow` to `settingsStore` & Update `SettingsModal.svelte`

**Files:**
- Modify: `frontend/src/lib/stores/settings.svelte.ts`
- Modify: `frontend/src/lib/components/layout/SettingsModal.svelte`

- [ ] **Step 1: Update `settings.svelte.ts` interface, defaults, and persistence methods**

Add `snapshotAutoHideWindow: boolean` (default: `true`) to `AppSettings`, `DEFAULT_SETTINGS`, class state, `loadSettings`, `saveSettings`, `updateSettings`, and `resetToDefaults`.

- [ ] **Step 2: Update `SettingsModal.svelte` with Auto-Hide Window toggle in General & Snapshot tab**

Add `editSnapshotAutoHide` state and a toggle with description in the Smart Screen Snapshot Card in `SettingsModal.svelte`.

- [ ] **Step 3: Verify frontend check passes**

Run: `export PATH=$PATH:/opt/homebrew/bin; pnpm --prefix frontend run check`

---

### Task 2: Update `App.svelte` to Use `settingsStore.snapshotAutoHideWindow` and Replace Sidebar Footer with Settings Trigger

**Files:**
- Modify: `frontend/src/App.svelte`

- [ ] **Step 1: Replace sidebar bottom switches with Settings button in `App.svelte`**

In `App.svelte`:
1. Use `settingsStore.snapshotAutoHideWindow` for window hiding logic during native screen capture.
2. Replace the bottom switches block in the `<aside>` sidebar with an Ant Design styled full-width Settings button triggering `settingsModalVisible = true`.

- [ ] **Step 2: Verify frontend check and production build**

Run: `export PATH=$PATH:/opt/homebrew/bin; pnpm --prefix frontend run check && pnpm --prefix frontend run build`

---

### Task 3: Verification & Backend Tests

**Files:**
- Run all test suites

- [ ] **Step 1: Run Go tests and Frontend checks**

Run: `export PATH=$PATH:/usr/local/go/bin:/opt/homebrew/bin; go test ./test/... -v` and `export PATH=$PATH:/opt/homebrew/bin; pnpm --prefix frontend run check`

- [ ] **Step 2: Commit all changes**
