# Design Cleanliness & Border De-Slop Plan (Selection, Plan Review, Timer)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to execute this plan task-by-task.

**Goal:** Eliminate unrefined, high-contrast, and glowing white/colored borders across the Selection Bar (`BatchActionBar.svelte`), Plan Gate Confirmation Card (`PlanReviewCard.svelte`), Live Execution Timer (`Composer.svelte`), and sidebar select checkboxes, aligning with our dark anthracite aesthetic.

---

### Audit of Issues Identified in Images:
1. **Sidebar Selection & Batch Action Bar (`BatchActionBar.svelte`)**:
   - `border border-ant-primary/25`: Creates a jarring rectangular outline around the batch action popup.
   - `border border-ant-primary/30` on selected badge.
   - `border border-ant-primary/25` and `border-ant-error/25` on buttons: Look rigid and unpolished.
   - Fix: Transition to seamless dark surfaces with soft zinc/white borders (`border-white/5` or `border-zinc-800`), refined subtle shadows, and tinted background fills without neon strokes.

2. **Execution Plan Review Card (`PlanReviewCard.svelte`)**:
   - `border border-ant-primary/30`: Creates an intense bounding box around the whole plan review card.
   - `border border-ant-primary/20` on "Plan Gate Active" badge.
   - `border border-white/10` on custom feedback button.
   - Fix: Use soft subtle border `border-[#27272a]` or `border-white/5`, dark anthracite background `bg-[#18181b]/95`, and refined button states.

3. **Live Execution Timer (`Composer.svelte`)**:
   - `border border-ant-primary/25`: In dark mode next to the stop button, this creates a bright, distracting white/blue bordered outline.
   - Fix: Use a clean flat dark capsule (`bg-zinc-800/80 text-zinc-300` or `bg-[#1e1e22]` with `border border-white/5`), subtle animated indicator without neon border lines.

---

### Task 1: Refactor `BatchActionBar.svelte` to Seamless Dark Theme
- Change container from `border-ant-primary/25` to `border border-[#27272a] bg-[#18181b] shadow-2xl`.
- Refactor buttons to use clean flat dark buttons with hover states instead of bright border lines.
- Refactor badge pill from `border-ant-primary/30` to `bg-zinc-800 text-zinc-300 font-mono text-[11px]`.

### Task 2: Refactor `PlanReviewCard.svelte` to Remove Bright Borders
- Change root container border to `border border-[#27272a] bg-[#18181b]/95 shadow-xl`.
- Change "Plan Gate Active" badge to `bg-indigo-500/10 text-indigo-400 border border-indigo-500/20`.
- Update button styles to clean zinc & indigo palette matching the rest of AetherGrok.

### Task 3: Refactor Live Timer in `Composer.svelte`
- Replace `border border-ant-primary/25` with `bg-[#1e1e22] border border-white/5 text-zinc-300` or clean minimal badge.

### Task 4: Verify & Build
- Run `npm --prefix frontend run build` to ensure clean compilation.
- Commit and push to GitHub.
