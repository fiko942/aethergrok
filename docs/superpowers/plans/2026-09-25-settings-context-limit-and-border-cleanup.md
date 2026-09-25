# Context Limit Setting and Dark Mode Border Polish Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Provide user-configurable context token window limit in Settings (default: 200k tokens), replace harsh high-contrast white borders with dark design tokens across SettingsModal, and refine the live reasoning and planning action card in MessageList.

**Architecture:** 
- Add dedicated UI section in `SettingsModal.svelte` for context token limit (with quick presets: 128k, 200k Default, 500k, 1M, and custom number input) wired to `settingsStore.maxContextTokens`.
- Propagate `settingsStore.maxContextTokens` into `Composer.svelte` and `session.svelte.ts` context usage calculations.
- Clean up high-contrast borders in `SettingsModal.svelte` by replacing `border-ant-border` with subtle dark tokens `border-white/5` and `border-white/10`.
- Soften the active reasoning card border in `MessageList.svelte` from stark bright box to subtle `border-white/5` with `bg-ant-bg-secondary/70`.

**Tech Stack:** Svelte 5 (Runes), Tailwind CSS 3.4, Ant Design Dark Tokens, Wails v2 / Go backend.

## Global Constraints
- All typography remains consistent with Anthropic Serif font family established across the app.
- Dark theme borders must use `border-white/5` for dividers and cards, and `border-white/10` for interactive containers.
- Default max context limit is 200,000 tokens (200k).
- Affirmative language in UI and code.

---

### Task 1: Add Context Window Limit Setting Card in SettingsModal

**Files:**
- Modify: `frontend/src/lib/components/layout/SettingsModal.svelte`
- Modify: `frontend/src/lib/components/chat/Composer.svelte`

**Interfaces:**
- Consumes: `settingsStore.maxContextTokens` (number)
- Produces: Context token limit preset selector and custom input in Settings UI, dynamic max limit in Composer usage meter

- [ ] **Step 1: Add Context Window Setting Card to `SettingsModal.svelte` under Models & Reasoning tab**
Add a dedicated card allowing the user to select preset limits (128K, 200K Default, 500K, 1M) or input a custom integer value, with helper text noting default is 200k tokens.

- [ ] **Step 2: Connect Composer context token usage calculation to `settingsStore.maxContextTokens`**
Update `frontend/src/lib/components/chat/Composer.svelte` to read `settingsStore.maxContextTokens || 200000` instead of hardcoded 200000 when computing context stats.

- [ ] **Step 3: Verify build compiles cleanly**
Run: `npm --prefix frontend run build`
Expected: 0 errors.

---

### Task 2: Dark Mode Border Cleanup across SettingsModal

**Files:**
- Modify: `frontend/src/lib/components/layout/SettingsModal.svelte`

**Interfaces:**
- Consumes: Tailwind classes `border-white/5`, `border-white/10`, `border-ant-primary/30`
- Produces: Polished dark mode settings interface with no stark white outlines

- [ ] **Step 1: Replace stark borders on inputs and controls in General tab**
In `SettingsModal.svelte`, update text inputs, dividers, and delay sliders to use `border-white/10` and `border-white/5`.

- [ ] **Step 2: Replace stark borders on Model selection cards in Models tab**
Update inactive model choice cards to `border-white/5 bg-ant-bg hover:border-ant-primary/40`.

- [ ] **Step 3: Replace stark borders on Permission cards in Permissions tab**
Update permission choice cards to `border-white/5 bg-ant-bg hover:border-ant-primary/40`.

- [ ] **Step 4: Replace stark borders on Theme choice cards and Shortcut table**
Update theme option preview borders and keyboard shortcut table borders to `border-white/5` and `border-white/10`.

- [ ] **Step 5: Replace footer divider and action toolbar borders**
Update modal footer divider to `border-white/5`.

---

### Task 3: Polish Live Reasoning and Planning Action Card in MessageList

**Files:**
- Modify: `frontend/src/lib/components/chat/MessageList.svelte`

**Interfaces:**
- Consumes: `isWorking`, `thinkingContextText`, `elapsedSeconds`
- Produces: Sleek, non-intrusive reasoning card with subtle border

- [ ] **Step 1: Update reasoning card border and container styling**
In `MessageList.svelte`, change the reasoning box outer container from `border border-ant-primary/25` to `border border-white/5 bg-ant-bg-secondary/70 backdrop-blur-md shadow-sm`.
Change the elapsed timer pill to `border border-white/5 bg-ant-bg-tertiary/60 text-ant-primary`.

- [ ] **Step 2: Verify frontend builds without errors**
Run: `npm --prefix frontend run build`
Expected: 0 errors.

---

### Task 4: Rebuild Wails App and Verify Application

**Files:**
- Build output: `build/bin/aethergrok.app`

- [ ] **Step 1: Build Wails binary**
Run: `wails build -clean`
Expected: `SUCCESS Build [darwin/arm64] built: build/bin/aethergrok.app`

- [ ] **Step 2: Relaunch application and verify dark mode UI**
Verify that:
1. Settings modal shows Context Limit setting defaulting to 200k tokens.
2. All tabs in Settings modal render seamless dark borders without stark white contrast.
3. Live reasoning card shows polished subtle borders.
