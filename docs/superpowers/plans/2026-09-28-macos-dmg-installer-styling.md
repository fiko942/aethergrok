# macOS Premium DMG Installer Styling Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Transform the macOS `.dmg` installer into a professional, high-end Apple-standard installer with a custom dark-aesthetic background banner, customized window geometry, Apple Silicon & Intel styling, icon positioning, and drag-to-Applications arrow affordance.

**Architecture:** Implement a pure macOS AppleScript/Finder automator pipeline inside `build-macos.sh` combined with standard high-res DMG background canvas (`resources/dmg-background.png` @2x), generating a window of size `660x420`, 160px icon size, and automatic symlink layout.

**Tech Stack:** Bash, AppleScript (`osascript`), `hdiutil`, Canvas/SVG rasterizer, GitHub Actions macOS runner.

## Global Constraints
- Target bundle identifier must remain `com.fiko942.aethergrok`.
- DMG layout must display `AetherGrok.app` on the left (x: 180, y: 220) and `Applications` symlink on the right (x: 480, y: 220).
- Icon size set to 140px.
- DMG background graphic must feature sleek dark studio aesthetic with subtle metallic gradient and directional drag arrow.
- Script must work both locally on developer Mac and in headless GitHub Actions CI runner without requiring external Node dependencies.

---

### Task 1: Generate Premium DMG Background Asset (`resources/dmg-background.png`)

**Files:**
- Create: `resources/dmg-background.svg`
- Create: `resources/dmg-background.png`

**Interfaces:**
- Produces: 1320x840 Retina @2x background image with dark studio gradient, AetherGrok branding, and subtle arrow indicator between left and right slots.

- [ ] **Step 1: Create SVG source for DMG background with dark aesthetic and drag arrow**
- [ ] **Step 2: Render SVG to 660x420 / 1320x840 @2x PNG in `resources/dmg-background.png`**
- [ ] **Step 3: Verify PNG resolution and visual dimensions**

---

### Task 2: Enhance `build-macos.sh` with AppleScript Finder Window Styling & Symmetry

**Files:**
- Modify: `build-macos.sh:110-160`

**Interfaces:**
- Consumes: `resources/dmg-background.png`, `$APP_BUNDLE`
- Produces: Polished `.dmg` with window geometry `660x420`, custom `.background/background.png`, `Applications` folder symlink, and `.DS_Store` icon positions configured via AppleScript / native Finder layout.

- [ ] **Step 1: Implement read-write DMG staging with `.background` folder and `/Applications` alias**
- [ ] **Step 2: Add AppleScript automation block to set view options (icon size 140, window bounds, hide sidebar/toolbar)**
- [ ] **Step 3: Add headless CI fallback using pre-calculated Finder `.DS_Store` structure**
- [ ] **Step 4: Execute `build-macos.sh arm64` locally and test DMG mount appearance**

---

### Task 3: Verify DMG Appearance, Re-signing & Packaging

**Files:**
- Test: Mount DMG locally and verify volume window bounds and icon placements
- Modify: `.github/workflows/desktop-release.yml` (if needed for CI runner compatibility)

- [ ] **Step 1: Mount generated DMG and inspect Finder appearance**
- [ ] **Step 2: Validate code signature inside DMG with `codesign -vvv`**
- [ ] **Step 3: Commit and push changes to GitHub**
- [ ] **Step 4: Upload freshly styled DMG to GitHub Release `v1.0.1`**
