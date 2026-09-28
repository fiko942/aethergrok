# UI/UX Audit & Comprehensive Enhancement Plan: Updates & Release Manager

> **Goal:** Transform the **Updates & Releases** tab in AetherGrok Studio into a world-class, intuitive, and visually stunning release center. Fix platform asset matching, humanize package downloads with OS badges, elevate typography and visual hierarchy, and streamline the 1-click update experience conforming to `/design-taste-frontend`, `/ui-ux-pro-max`, and Ant Design design system standards.

---

## 1. UX/UI Audit Findings & Root Causes

Based on visual inspection of the active application screens:

### A. Broken Platform Asset Detection & Missing 1-Click CTA
- **Root Cause**: `changelog.json` defines download keys as `macos_arm64_dmg`, `windows_x64_setup`, etc. However, `MatchPlatformAsset()` in `pkg/updater/updater.go` strictly checked for `.dmg` and `.exe` suffixes with periods (`.`). As a result, `MatchPlatformAsset` returned `nil` when parsing local changelogs or GitHub assets formatted with underscores.
- **UX Impact**: Instead of displaying the primary **"Update Now (Auto-Install)"** action with file size and speed telemetry, the UI fell back to a generic **"View Release on GitHub"** button, preventing users from using the 1-click auto-install flow.

### B. Raw Technical Package Names & Poor Asset Discoverability
- **Current State**: The package grid displays raw system IDs such as `macos_arm64_dmg`, `windows_x64_portable`, `macos_x64_dmg`, and `windows_arm64_setup`.
- **UX Impact**: End-users must decipher technical slug names. Missing visual OS branding (Apple Logo, Windows Logo) and architecture badges (Apple Silicon `arm64`, Intel `x64`, Windows `x64`) make identifying the right installer confusing.

### C. Visual Clutter & Sub-optimal Hierarchy
- **Header Section**: The version pill `v1.0.0` was placed inline next to the main title in an unformatted pill without clear semantic distinction.
- **Update Hero Banner**:
  - The version delta `v1.0.1 (Current: v1.0.0) -> v1.0.1` had overlapping blue badges and low contrast on dark backgrounds.
  - Section headers `KEY HIGHLIGHTS & CHANGES:` and `AVAILABLE PACKAGES:` used all-caps with small font sizes that lacked rhythm and visual grouping.
- **Accordion Rows**:
  - The date `28 Sep 2026` overlapped with the chevron arrow on narrow viewports.
  - Release highlights lacked structured categories (e.g. *Features*, *Improvements*, *Fixes*).

---

## 2. Proposed Design System & UX Upgrades

```
+-----------------------------------------------------------------------------------+
|  Updates & Release Changelog                            [ Check for Updates ↻ ]  |
|  Live GitHub release integration, auto-updater, and version changelog.            |
+-----------------------------------------------------------------------------------+
|  [⚡ NEW UPDATE READY: v1.0.1]                                                     |
|  -------------------------------------------------------------------------------- |
|  Installed: v1.0.0  ───►  Latest: v1.0.1  (Released 28 Sep 2026)                  |
|                                                                                   |
|  AetherGrok 1.0.1 - Backend Persistent Storage & Multi-Platform CI/CD             |
|  • Pure-Go persistent JSONL storage engine for settings and workspaces            |
|  • Automated GitHub Actions desktop builder for macOS DMG & Windows EXE          |
|                                                                                   |
|  [ ⚡ Update Now (Auto-Install) • 38.2 MB ]   [ 📥 Manual Download ]   [ Notes ↗ ]|
|  (Detected Platform: macOS Apple Silicon • arm64)                                 |
+-----------------------------------------------------------------------------------+
|  Version History (2 Releases)                                                     |
|  +-----------------------------------------------------------------------------+  |
|  | [ v1.0.1 ]  AetherGrok 1.0.1 - Backend Persistent Storage ...  [Latest] 28 Sep ▼ |
|  +-----------------------------------------------------------------------------+  |
|  | Key Highlights:                                                              |
|  |   ✦ Migrasi pengaturan ke backend storage murni Go (~/.grok/*.jsonl)         |
|  |   ✦ Penambahan toolchain rilis otomatis (pnpm run release)                   |
|  |                                                                              |
|  | Available Installers & Packages:                                             |
|  |   [  macOS Apple Silicon (DMG)      38.2 MB  ↓ ] [ ⊞ Windows x64 Installer  ↓ ] |
|  |   [  macOS Intel x64 (DMG)          39.1 MB  ↓ ] [ ⊞ Windows ARM64 Installer↓ ] |
|  |   [  macOS ARM64 Portable (tar.gz)  35.0 MB  ↓ ] [ ⊞ Windows x64 Portable   ↓ ] |
|  +-----------------------------------------------------------------------------+  |
+-----------------------------------------------------------------------------------+
```

### Visual Enhancements:
1. **Platform Asset Matcher Upgrade (`pkg/updater/updater.go`)**:
   - Normalize asset keys: Support both filenames (`AetherGrok-1.0.1-macOS-arm64.dmg`) and canonical keys (`macos_arm64_dmg`, `windows_x64_setup`).
   - Automatically detect the active user platform (`darwin arm64` -> `macOS Apple Silicon`, `darwin amd64` -> `macOS Intel`, `windows amd64` -> `Windows x64`) and match with 100% precision.
2. **Humanized Package Cards**:
   - Parse package keys into human-friendly labels:
     - `macos_arm64_dmg` -> **macOS Apple Silicon (DMG)**
     - `macos_x64_dmg` -> **macOS Intel x64 (DMG)**
     - `windows_x64_setup` -> **Windows x64 Setup (Installer)**
     - `windows_arm64_setup` -> **Windows ARM64 Setup (Installer)**
     - `*_portable` -> **Portable Archive (.zip / .tar.gz)**
   - Add Apple (`Apple` / `Laptop`) and Windows (`Monitor` / `Layers`) platform icons with architecture tags (`ARM64`, `x64`).
3. **High-End Typography & Contrast**:
   - Replace raw bullet points with styled emerald/primary accent dots (`✦` or subtle styled pills).
   - Format release dates cleanly with locale-aware short formats.
   - Clean up badge hierarchy: `Current` (emerald pill), `Latest` (amber/primary pill), `Update Available` (pulsing indicator).

---

## 3. Implementation Tasks

### Task 1: Fix Asset Key Matching in Backend Updater (`pkg/updater/updater.go`)
- Support canonical asset keys (`macos_arm64_dmg`, `windows_x64_setup`, `macos_x64_dmg`, `windows_arm64_setup`) in `MatchPlatformAsset`.
- Add test cases in `pkg/updater/updater_test.go` covering canonical keys and real GitHub asset names.

### Task 2: Build Package Formatter Utility (`frontend/src/lib/utils/packageFormatter.ts`)
- Map raw asset names to human-readable platform labels, icons, and package types:
  ```ts
  export function formatPackageInfo(name: string): {
    platform: 'macOS' | 'Windows' | 'Linux' | 'Other';
    arch: 'Apple Silicon (ARM64)' | 'Intel (x64)' | 'ARM64' | 'x64' | 'Universal';
    kind: 'Installer (.dmg)' | 'Setup (.exe)' | 'Portable (.zip)' | 'Portable (.tar.gz)' | 'Package';
    recommended: boolean;
  }
  ```

### Task 3: Redesign `UpdatesTab.svelte` with Polish & Hierarchy
- **Header & Status Bar**: Clean, unified bar with installed version badge and last-check timestamp.
- **Hero Update Card**:
  - Prominent update alert showing version jump (`v1.0.0 → v1.0.1`).
  - Detected system platform badge (`Your System: macOS Apple Silicon arm64`).
  - Prominent primary CTA: **"Update Now (Auto-Install)"** with file size and animated icon.
  - Fallback secondary actions: **"Manual Download"** and **"Full Release Notes"**.
  - Integrated live streaming progress bar (Speed MB/s, transferred MB, stage indicators).
- **Changelog Accordion**:
  - Refined headers with version badge, title, relative date, and expand/collapse animation.
  - Highlight bullets styled with clean typography and comfortable line-height.
  - Package grid displaying friendly OS titles, architecture tags, and download action.

### Task 4: Verification & Test
- Run `go test -v ./pkg/updater/...` to verify asset matching.
- Run `cd frontend && pnpm run build` to verify UI compilation.
- Commit and push to `main`.
