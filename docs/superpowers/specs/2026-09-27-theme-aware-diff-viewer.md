# Specification: Theme-Aware Diff & File Inspector with Clean Diff Lines

## 1. Overview & Objective
Refine the Git Changes inspector modal (`DiffViewModal.svelte`) and the Workspace File viewer modal (`FileViewerModal.svelte`) to deliver a high-contrast, theme-adaptive experience across both Clean Light (`light-antd`) and Dark Studio (`dark-studio`, `dark-high-contrast`) themes.

## 2. Key Requirements
1. **Clean Diff Line Rendering**:
   - Strip the leading `+` symbol from addition lines (`line.slice(1)`).
   - Strip the leading `-` symbol from deletion lines (`line.slice(1)`).
   - Indication of additions and deletions is handled exclusively through row background colors and gutter line numbering.
2. **Theme Harmonization & Semantic Token Integration**:
   - Replace all hardcoded hex values (`#18181b`, `#27272a`, `#141416`, `#0f0f11`, `text-zinc-200`) with Ant Design semantic design tokens:
     - `bg-ant-bg` for modal surfaces and diff container background.
     - `bg-ant-bg-secondary` for header and footer chrome.
     - `bg-ant-bg-tertiary` for toolbar button backgrounds.
     - `border-ant-border` for dividers, table borders, and modal borders.
     - `text-ant-text`, `text-ant-text-secondary`, and `text-ant-text-muted` for hierarchical typography.
   - Diff Addition Tint: `bg-emerald-500/10 hover:bg-emerald-500/15 text-emerald-800 dark:text-emerald-200`.
   - Diff Deletion Tint: `bg-rose-500/10 hover:bg-rose-500/15 text-rose-800 dark:text-rose-200`.
   - Hunk Headers: `bg-sky-500/10 text-sky-700 dark:text-sky-300 border-y border-sky-500/20`.
3. **Markdown Preview in File Viewer**:
   - Headers, lists, blockquotes, code blocks, and tables in Markdown preview mode seamlessly inherit the active theme's font color and border tokens.

## 3. Architecture & File Mapping
- `frontend/src/lib/components/workspace/DiffViewModal.svelte`: Diff parser, line cleaner, and theme-token styled table view.
- `frontend/src/lib/components/workspace/FileViewerModal.svelte`: Theme-token styled file inspector with Markdown and Prism syntax viewer.
