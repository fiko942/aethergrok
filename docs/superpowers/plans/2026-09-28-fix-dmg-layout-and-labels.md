# Implementation Plan: Fix macOS DMG Installer Layout, White Space, and Icon Labels

Fix the macOS DMG installer appearance in Finder where a large white area appears at the bottom of the window and the icon text labels ("AetherGrok" and "Applications") are clipped or missing.

## Proposed Changes

### Background & Asset Geometry Calibration
1. **Calibrate Dimensions to Match Finder Viewport Perfectly**:
   - Total window bounds: `{left: 300, top: 100, right: 960, bottom: 500}` -> Width: 660px, Height: 400px.
   - Background image resolution: **660px x 400px** (@1x) and **1320px x 800px** (@2x).
   - Solid background fill `#0f1218` ensuring no transparent margins or white canvas leakage.
2. **Swift Native Multi-Scale Asset Generator**:
   - Write a helper Swift script (`scripts/generate-dmg-bg.swift`) that generates crisp @1x and @2x PNGs and combines them into a multi-resolution `resources/dmg-background.tiff` using AppKit/CoreGraphics.
   - Ensure header title, drag guide instruction, center arrow graphics (aligned at Y: 150), and footer info are positioned cleanly without overlapping icon labels.

### AppleScript Finder Window Layout in `build-macos.sh`
1. Update window bounds and icon positions:
   - Window bounds: `{300, 100, 960, 500}` (660x400)
   - Left App Icon: `{160, 150}`
   - Right Applications Folder: `{500, 150}`
   - Explicitly configure icon size `96`, label position `bottom`, text size `12`, and arrangement `not arranged`.
   - Ensure disk sync and proper `.DS_Store` write flush before detaching.

## Verification Plan
1. Run Swift asset generator to create `dmg-background.tiff`, `dmg-background.png`, `dmg-background@2x.png`.
2. Execute `bash build-macos.sh arm64` to build `AetherGrok-1.0.1-macOS-arm64.dmg`.
3. Mount the DMG using `hdiutil attach` and inspect Finder window geometry via AppleScript & screen capture.
4. Verify no white space exists at the bottom and icon labels "AetherGrok" and "Applications" are clearly visible.
