# Implementation Plan: App Icon Processing, Packaging & In-App Header Logo Integration

## Overview
Process the user-provided application icon image:
1. Crop accurately to the squircle icon body and remove the outer grey background canvas (`#434449` background) with a smooth, anti-aliased alpha transparency mask.
2. Update application packaging icon assets:
   - `build/appicon.png` (1024x1024 master icon for Wails macOS/desktop build)
   - `resources/grok-icon-512.png` and `resources/grok-icon-round-512.png` (512x512 assets)
   - `frontend/public/app-icon.png` (1024x1024) and `frontend/public/app-icon-64.png` / `frontend/public/app-icon-128.png` (sharp frontend assets)
3. Integrate the new icon into the in-app UI header in `frontend/src/App.svelte` replacing the generic sparkle icon with the custom branded app icon.
4. Verify compilation, build assets, and visual presentation.

## User Review Checkpoint
Autonomous execution under superpowers conventions without blocking turn.

## Proposed Changes

### 1. Image Processing & Asset Generation
- Execute Python / Pillow script to crop the image `image-9db5ae69-9811-42be-81d1-cf3101dfc3f0.jpg` to the centered squircle bounds, generate an 8x supersampled anti-aliased macOS squircle mask, and export:
  - `build/appicon.png` (1024x1024 RGBA)
  - `resources/grok-icon-512.png` (512x512 RGBA)
  - `resources/grok-icon.png` (256x256 RGBA)
  - `frontend/public/app-icon.png` (1024x1024 RGBA)
  - `frontend/public/app-icon-128.png` (128x128 RGBA)
  - `frontend/public/app-icon-64.png` (64x64 RGBA)

### 2. Header UI Logo in `frontend/src/App.svelte`
- Update the header branding in `frontend/src/App.svelte`:
  - Replace `<div class="flex items-center justify-center w-7 h-7 rounded-lg bg-ant-primary/10 border border-ant-primary/20 text-ant-primary shadow-sm"><Sparkles size={16} /></div>`
  - With an app icon container featuring `<img src="/app-icon-64.png" alt="AetherGrok Logo" class="w-6 h-6 rounded-md object-contain drop-shadow-sm pointer-events-none select-none" />` with an ambient glow / border container that harmonizes with Ant Design Dark theme.

### 3. Verification Plan
- Run `pnpm run check` in `frontend/`
- Run `pnpm run build` in `frontend/`
- Verify that assets exist, have alpha transparency around the squircle corners, and correctly resolve in the bundle.
