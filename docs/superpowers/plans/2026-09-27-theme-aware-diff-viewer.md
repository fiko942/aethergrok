# Plan: Theme-Aware Diff & File Inspector with Clean Diff Lines

## Problem Summary
1. In the Git Changes feature (`DiffViewModal.svelte`), additions (`+`) and deletions (`-`) have the literal `+` or `-` symbol prepended to the code line content. The user requested removing the leading `+` / `-` symbols so that additions/deletions are indicated cleanly by background color highlighting while preserving line numbers and code indentation.
2. `DiffViewModal.svelte` and `FileViewerModal.svelte` contain hardcoded dark colors (`bg-[#18181b]`, `border-[#27272a]`, `bg-[#141416]`, `bg-[#0f0f11]`, `text-zinc-200`, etc.) that do not adapt when the user switches to Clean Light (`light-antd`) theme.

## Scope of Changes
1. **`frontend/src/lib/components/workspace/DiffViewModal.svelte`**:
   - Strip the leading `+` and `-` characters when parsing or rendering lines: `line.slice(1)`.
   - Replace hardcoded dark colors with semantic theme classes (`bg-ant-bg`, `border-ant-border`, `bg-ant-bg-secondary`, `bg-ant-bg-tertiary`, `text-ant-text`, `text-ant-text-secondary`, `text-ant-text-muted`).
   - Style diff additions with clean green tints (`bg-emerald-500/10 text-emerald-700 dark:text-emerald-300 hover:bg-emerald-500/15`).
   - Style diff deletions with clean rose/red tints (`bg-rose-500/10 text-rose-700 dark:text-rose-300 hover:bg-rose-500/15`).
   - Style hunk headers (`@@`) with `bg-sky-500/10 text-sky-700 dark:text-sky-300 border-sky-500/20`.
   - Ensure line number gutter has subtle border and muted text color matching the active theme.

2. **`frontend/src/lib/components/workspace/FileViewerModal.svelte`**:
   - Update modal container, header, code container, line numbers gutter, and footer to use semantic Ant Design tokens (`bg-ant-bg`, `border-ant-border`, `text-ant-text`, `bg-ant-bg-secondary`, `bg-ant-bg-tertiary`, `text-ant-text-secondary`, `text-ant-text-muted`).
   - Update markdown view and Prism syntax container to blend seamlessly in light and dark modes.

3. **Verification**:
   - Execute `pnpm build` in `frontend/`.
   - Execute `wails build` for macOS desktop bundle.
