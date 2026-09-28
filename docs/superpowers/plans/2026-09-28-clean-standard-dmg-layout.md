# Minimal Clean DMG Layout (Standard Finder Presentation) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Configure the macOS DMG installer to use the standard default macOS Finder presentation without custom background artwork, featuring 2 clean icons (`AetherGrok.app` and `Applications` symlink) with clean positioning and standard view options.

**Architecture:** Update `build-macos.sh` to remove custom `.background` folder generation and background image assignment, simplify Finder AppleScript view configuration to clean default icon view without background pictures, set window bounds and icon positions (`AetherGrok.app` at `{150, 160}` and `Applications` at `{450, 160}` in a 600x320 window), and rebuild and visually verify the DMG layout.

**Tech Stack:** Bash, AppleScript (`osascript`), `hdiutil`, macOS Finder automation.

## Global Constraints

- Avoid custom background image files or `.background` directories in the DMG root.
- Standard default Finder window appearance with 2 icons: `AetherGrok.app` and `Applications`.
- Proper window sizing (approx 600x320) and clean centered icon spacing.
- Text labels below icons must render standard black text with macOS default icon arrangement.

---

### Task 1: Update DMG Build Script for Standard Backgroundless Finder Layout

**Files:**
- Modify: `build-macos.sh`

**Interfaces:**
- Consumes: Existing macOS DMG packaging flow in `build-macos.sh`.
- Produces: DMG installer without `.background` assets and with clean default AppleScript Finder settings.

- [ ] **Step 1: Edit `build-macos.sh` to remove background image generation and simplify AppleScript**

Update `build-macos.sh`:
- Remove `.background` directory creation and background copying (`background.png`, `background.tiff`, etc.).
- Update AppleScript in `build-macos.sh`:
  - Set window bounds to `{300, 150, 900, 470}` (600x320 window).
  - Set `arrangement` to `not arranged`.
  - Set `icon size` to `96`.
  - Set `label position` to `bottom`.
  - Set `text size` to `12`.
  - Set `position` of `AetherGrok.app` to `{150, 150}`.
  - Set `position` of `Applications` to `{450, 150}`.
  - Omit `background picture` assignment so Finder uses its default clean window background.

- [ ] **Step 2: Test building the macOS DMG**

Run: `bash build-macos.sh arm64`
Expected: Build completes successfully and outputs `build/bin/AetherGrok-1.0.1-macOS-arm64.dmg`.

---

### Task 2: Mount and Visually Verify Standard DMG Window

**Files:**
- Inspect: Mounted DMG Finder window

**Interfaces:**
- Consumes: Built DMG file `build/bin/AetherGrok-1.0.1-macOS-arm64.dmg`.
- Produces: Visual verification confirming 2 cleanly spaced icons, visible black labels on standard background, and no background image quirks.

- [ ] **Step 1: Mount the newly built DMG and capture a window screenshot**

Run:
```bash
hdiutil attach build/bin/AetherGrok-1.0.1-macOS-arm64.dmg -mountpoint /tmp/dmg_test_clean
osascript -e 'tell application "Finder" to open (POSIX file "/tmp/dmg_test_clean")'
delay 1
screencapture -x -R 150,50,750,450 /tmp/dmg_clean_test.png
```

- [ ] **Step 2: Inspect screenshot using visual verification tool**

Verify that:
1. There is no custom background image.
2. The 2 icons (`AetherGrok` and `Applications`) are cleanly positioned side by side.
3. The labels are clearly legible with default macOS system rendering.
4. Window dimensions fit the icons comfortably.

- [ ] **Step 3: Unmount DMG and cleanup temporary test mount**

Run:
```bash
hdiutil detach /tmp/dmg_test_clean
rm -f /tmp/dmg_clean_test.png
```

---

### Task 3: Commit and Push to GitHub

**Files:**
- Modify: `build-macos.sh`
- Create: `docs/superpowers/plans/2026-09-28-clean-standard-dmg-layout.md`

- [ ] **Step 1: Stage and commit changes**

```bash
git add build-macos.sh docs/superpowers/plans/2026-09-28-clean-standard-dmg-layout.md
git commit -m "style(dmg): switch DMG to clean default Finder layout with standard icons"
```

- [ ] **Step 2: Push changes to remote repository**

```bash
git push origin main
```
