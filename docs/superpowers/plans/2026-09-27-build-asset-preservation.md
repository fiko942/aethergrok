# Implementation Plan: Automated Build Cleanup and Asset Preservation

## Overview
When a developer or CI cleans or deletes the `build/` folder and runs `npm run build:prod`:
1. The build pipeline must guarantee that `build/` is automatically prepared and initialized with canonical custom assets (`resources/app-assets/`) so Wails never regenerates placeholder/default assets (e.g. default Wails icon or stripped Info.plist).
2. All previous build artifacts in `build/bin/` (and any stale outputs) must be wiped cleanly prior to starting the build.
3. The build process runs cleanly and seamlessly for both macOS and Windows targets.

## Root Cause Analysis
- Wails CLI generates default template files (`build/appicon.png`, `build/darwin/Info.plist`, etc.) if `build/` is completely empty or missing.
- When `build/` was wiped, running `npm run build:prod` caused Wails to re-populate `build/appicon.png` with its internal default icon instead of the custom 1024x1024 AetherGrok icon.
- Additionally, `build/bin/` retained older binaries unless manually cleaned.

## Proposed Changes
1. **Canonical Asset Storage**:
   - `resources/app-assets/` stores the authoritative copies of:
     - `appicon.png` (Custom 1024x1024 macOS squircle icon)
     - `darwin/Info.plist` (With `NSMicrophoneUsageDescription`)
     - `windows/icon.ico`, `windows/info.json`, `windows/wails.exe.manifest`
2. **`scripts/build-production.mjs` Enhancements**:
   - `cleanBuildBin()`: Wipes `build/bin/` before any compilation so old binaries/artifacts are cleanly removed.
   - `ensureBuildAssets()`: Ensures `build/`, `build/darwin/`, and `build/windows/` exist and synchronizes canonical assets from `resources/app-assets/` into `build/` before invoking `wails build`.
   - Post-build macOS plist enhancement: Uses `PlistBuddy` to verify/inject microphone and speech recognition descriptions into `build/bin/aethergrok.app/Contents/Info.plist` if building for macOS.
3. **Verification**:
   - Delete `build/` completely (`rm -rf build`).
   - Run `npm run build:prod -- --mac`.
   - Verify `build/appicon.png` matches canonical MD5 (`ed1d376d940d091c8920e95f139567b4`).
   - Verify `build/bin/aethergrok.app` is built and contains the correct icon and plist entries.
