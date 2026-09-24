# Specification Addendum: Popover Layering, Opacity & Z-Index Architecture

**Date**: 2026-09-25  
**Version**: 1.1.1  
**Status**: Implemented & Verified  
**Component**: Prompt Composer Popovers (`ModelSelectDropdown`, `ReasoningEffortDropdown`, `SlashCommandPopup`)  
**Target Repository**: `fiko942/grok-build`

---

## 1. Issue Analysis & Root Cause

From the UI screenshot audit:
1. **Translucency / See-Through Background**: The popover card was inheriting subtle alpha/backdrop transparencies on nested containers, causing behind-elements (such as diff code cards and chat bubble text) to bleed through the popover search input and model list.
2. **Layering / Stacking Context**: The composer wrapper lacked an explicit stacking context (`relative z-30`), meaning child popovers were rendered on the same composite layer as preceding viewport messages.

---

## 2. Technical Fixes Implemented

1. **Explicit Solid Backgrounds**:
   - Replaced dynamic semi-transparent classes with strict solid tokens:
     - **Light Theme**: Background `bg-[#ffffff]`, Header/Footer `bg-[#fafafa]`, border subtle neutral ring `ring-1 ring-black/10`.
     - **Dark Studio Theme**: Background `bg-[#0f1117]`, Header/Footer `bg-[#181b26]`, ring `ring-1 ring-white/10`.
2. **Elevated Stacking Context (`Composer.svelte`)**:
   - Added `relative z-30` to composer container.
   - Added `relative z-40` to bottom control bar.
   - Added `z-[100]` to the dropdown popovers.
3. **Deep Elevation Shadows**:
   - Updated box-shadow to `0 20px 40px -4px rgba(0, 0, 0, 0.45), 0 8px 16px -4px rgba(0, 0, 0, 0.25)` to guarantee distinct separation from background workspace content.

---

## 3. Verification

- `svelte-check`: 0 errors, 0 warnings.
- `vite build`: Clean production bundle.
- Code committed and pushed to `main` (`ccb32a7`).
