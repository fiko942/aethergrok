# Implementation Plan - Skill Catalog Badge Cleanup and Animated Pill Filter Tabs

Fix the white border on the installed skills count badge and implement smooth animated sliding pill tabs for category switching in `SkillCatalog.svelte`.

## User Requirements
1. **Remove white border on `{skills.length} Installed` badge**: Replace with dark studio token styling (`bg-blue-500/10 text-blue-400 border border-blue-500/20`).
2. **Animated Pill Tab Filter**: When switching categories (`All`, `Frontend`, `Backend`, `Design`, `Agents`, `Tools`), the active pill indicator should smoothly slide and resize across the active item.

## Proposed Changes

### 1. Update `frontend/src/lib/components/skills/SkillCatalog.svelte`
- Replace badge markup:
  ```svelte
  <span class="px-2.5 py-0.5 text-[10px] font-serif font-semibold bg-blue-500/15 text-blue-400 rounded-full border border-blue-500/30">
    {skills.length} Installed
  </span>
  ```
- Implement sliding pill animation for category tabs:
  - Track active tab button element positions or use CSS sliding indicator relative to container:
  - Container with `relative flex items-center bg-ant-bg-secondary p-1 rounded-xl border border-white/5`.
  - Background sliding pill element with `transition-all duration-200 ease-out` calculated from active tab index / bounding rect or structured using layout transitions.
  - Category buttons have `relative z-10 px-3 py-1 text-xs font-medium rounded-lg transition-colors`.

## Verification Plan
1. `cd frontend && pnpm run check`
2. `cd frontend && pnpm run build`
3. `go test ./test/... -v`
