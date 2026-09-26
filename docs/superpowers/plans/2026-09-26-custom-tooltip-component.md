# Implementation Plan - Custom Tooltip Component for Settings & UI Elements

Implement a reusable, highly polished custom Tooltip component adhering to Ant Design Dark tokens and modern micro-interaction principles, and replace the native browser `title="..."` attribute on the sidebar Settings button.

## User Requirements
1. The tooltip on the bottom-left sidebar Settings button must use a custom-crafted UI tooltip component with dark glassmorphism, shortcut badge rendering, and smooth fade/scale micro-animations.
2. It should support customizable positions (top, bottom, left, right), hotkey shortcut display (`⌘,`), and safe hover interactions.

## Proposed Changes

### 1. Create `frontend/src/lib/antd/Tooltip.svelte`
- Built using Svelte 5 Runes (`$state`, `$props`, snippet/children).
- Props:
  - `title`: string (Tooltip text)
  - `shortcut`: string optional (e.g. `⌘,` or `Ctrl+,`)
  - `placement`: `'top' | 'bottom' | 'left' | 'right'` (default `'top'`)
  - `delay`: number in ms before showing (default `200`)
  - `children`: Snippet
- Styling:
  - `bg-[#18181c]/95 border border-white/10 text-white shadow-xl backdrop-blur-md px-2.5 py-1 rounded-md text-xs font-serif`
  - Subtle pointer arrow matching the placement.
  - Micro-animation: `animate-in fade-in zoom-in-95 duration-150`.

### 2. Update `frontend/src/App.svelte`
- Import `Tooltip` from `$lib/antd/Tooltip.svelte`.
- Wrap the bottom Settings button with `<Tooltip title="Open Settings & Preferences" shortcut={isMac ? '⌘,' : 'Ctrl+,'} placement="top">...</Tooltip>`.
- Remove native `title="..."` attribute from the `<button>` to prevent default HTML browser tooltips.

## Verification Plan
1. `cd frontend && pnpm run check`
2. `cd frontend && pnpm run build`
3. `go test ./test/... -v`
