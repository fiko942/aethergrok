# In-App 1-Click Auto-Download & Self-Install Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Enable seamless, 1-click in-app download and native installation of updates on both macOS and Windows. When an update is available, the user clicks "Download & Install Update" directly in the UI; the app downloads the asset in the background with real-time percentage progress, verifies the SHA256 checksum, executes the native installer/updater, and smoothly relaunches the new version.

**Architecture:**
1. **Backend Installer & Stream Engine (`pkg/updater/installer.go`)**:
   - Downloads remote binary assets to `$TMPDIR/aethergrok_update/` while streaming progress events (`updater:progress` with `{ downloadedBytes, totalBytes, percent, speedFormatted, stage: 'downloading'|'verifying'|'installing'|'ready'|'error' }`) via Wails runtime.
   - Computes and verifies SHA256 hash.
   - **macOS Pipeline**: Mounts `.dmg` using `hdiutil attach -nobrowse -noverify -noautoopen`, locates `AetherGrok.app` in the volume, runs a detached upgrade helper script that replaces the running `.app` in `/Applications` (or user directory), unmounts the disk image, and relaunches the app.
   - **Windows Pipeline**: Launches the NSIS setup executable (`-setup.exe`) with silent flags `/S` or elevated helper, then exits the current process so the installer can overwrite and launch the updated application.
2. **Frontend UX & Design System (`UpdatesTab.svelte` & `updater.svelte.ts`)**:
   - Adheres to Ant Design Dark/Light tokens & Anthropic serif typography.
   - Visual Progress Bar with smooth transitions, animated pulsing glow, transferred size (`24.5 MB / 38.2 MB`), real-time download speed (`4.8 MB/s`), and stage indicators (*Downloading...*, *Verifying checksum...*, *Installing & Restarting...*).
   - Instant "Cancel" and "Retry" safety actions.
   - Error handling with clear diagnostic feedback and fallback to manual browser download if native installation permissions require elevation.

**Tech Stack:** Go (HTTP stream, crypto/sha256, os/exec, syscall), Svelte 5 runes (`$state`), Tailwind CSS, Lucide icons, Ant Design tokens.

---

### File Structure Map
- **Backend**:
  - `pkg/updater/installer.go` - Background download streamer, SHA256 validator, macOS DMG replacer script, and Windows setup launcher.
  - `pkg/updater/installer_test.go` - Unit tests for download progress calculations and hash verifications.
  - `app.go` - Expose `DownloadAndInstallUpdate(assetUrl, expectedSha256 string)` and `CancelUpdateDownload()` via Wails IPC.
  - `frontend/src/app.d.ts` - TypeScript signatures for updater IPC and progress events.
- **Frontend**:
  - `frontend/src/lib/stores/updater.svelte.ts` - Store state for `downloading`, `progress`, `speed`, `stage`, `downloadError`, and download abort controller.
  - `frontend/src/lib/components/settings/UpdatesTab.svelte` - Premium interactive update card with animated progress bar, status badges, and 1-click update trigger.

---

### Task 1: Implement Backend Download & Native Self-Installer (`pkg/updater/installer.go`)

**Files:**
- Create: `pkg/updater/installer.go`
- Create: `pkg/updater/installer_test.go`
- Modify: `app.go`
- Modify: `frontend/src/app.d.ts`

- [ ] **Step 1: Create `pkg/updater/installer.go`**
  Implement Go structs and methods:
  ```go
  type UpdateProgress struct {
      DownloadedBytes int64   `json:"downloadedBytes"`
      TotalBytes      int64   `json:"totalBytes"`
      Percent         float64 `json:"percent"`
      SpeedFormatted  string  `json:"speedFormatted"`
      Stage           string  `json:"stage"` // "downloading" | "verifying" | "installing" | "ready" | "error"
      Message         string  `json:"message"`
  }
  ```
  Functions:
  - `DownloadAssetWithProgress(ctx context.Context, downloadUrl string, onProgress func(UpdateProgress)) (string, error)`
  - `VerifyChecksum(filePath string, expectedHash string) (bool, error)`
  - `ApplyUpdateMacOS(dmgPath string, onProgress func(UpdateProgress)) error`
  - `ApplyUpdateWindows(setupExePath string, onProgress func(UpdateProgress)) error`
  - `LaunchAndRelaunch(targetAppPath string)`

- [ ] **Step 2: Add unit tests in `pkg/updater/installer_test.go`**
  Verify checksum checking and progress tracking math.

- [ ] **Step 3: Expose IPC methods in `app.go`**
  - `DownloadAndInstallUpdate(assetUrl string, sha256Url string) error`
  - `CancelUpdateDownload() error`

- [ ] **Step 4: Run Go tests**
  ```bash
  export PATH=$PATH:/usr/local/go/bin:~/go/bin:/opt/homebrew/bin
  go test -v ./pkg/updater/...
  ```

- [ ] **Step 5: Commit backend installer**
  ```bash
  git add pkg/updater/ app.go frontend/src/app.d.ts
  git commit -m "feat(updater): implement in-app background download, checksum verification, and native self-installer"
  ```

---

### Task 2: Update Svelte Updater Store for Live Progress & Self-Install (`updater.svelte.ts`)

**Files:**
- Modify: `frontend/src/lib/stores/updater.svelte.ts`

- [ ] **Step 1: Enhance `updaterStore` state and methods**
  Add state:
  - `isInstalling = $state<boolean>(false)`
  - `installProgress = $state<UpdateProgress>({ downloadedBytes: 0, totalBytes: 0, percent: 0, speedFormatted: '0 KB/s', stage: 'idle', message: '' })`
  - `installError = $state<string | null>(null)`
  Methods:
  - `startDownloadAndInstall(assetUrl?: string, sha256Url?: string)`
  - `cancelDownload()`
  - Listen for `updater:progress` event from Wails runtime.

- [ ] **Step 2: Commit updater store updates**
  ```bash
  git add frontend/src/lib/stores/updater.svelte.ts
  git commit -m "feat(frontend): add download progress and self-install lifecycle to updater store"
  ```

---

### Task 3: Build Premium UI/UX in `UpdatesTab.svelte` with Animated Progress

**Files:**
- Modify: `frontend/src/lib/components/settings/UpdatesTab.svelte`

- [ ] **Step 1: Redesign Update Hero Card with 1-Click "Download & Install Update"**
  - High-end Ant Design styled progress bar with emerald/primary gradient, smooth width transitions, and indeterminate pulse while verifying.
  - Live progress telemetry: `Percent %`, `Transferred MB / Total MB`, `Speed MB/s`, and current stage pill (`Downloading`, `Verifying SHA-256`, `Installing & Restarting`).
  - Tactile buttons:
    - Primary CTA: **"Update Now (Download & Install)"** with icon.
    - Active State: **"Installing Update..."** with spinner, disabled click, and "Cancel" secondary button.
    - Secondary fallback: "Manual Download via Browser".

- [ ] **Step 2: Verify Frontend Build**
  ```bash
  cd frontend && pnpm run build
  ```

- [ ] **Step 3: Commit UI enhancements**
  ```bash
  git add frontend/src/lib/components/settings/UpdatesTab.svelte
  git commit -m "feat(ui): add polished 1-click self-update card with real-time download progress and status indicators"
  ```

---

### Task 4: End-to-End Verification & Superpowers Spec Documentation

**Files:**
- Update: `docs/superpowers/specs/2026-09-28-in-app-auto-update-and-changelog.md`
- Update: `/Users/fiko942/.grok/memory-v2/workspaces/grok-build-7ab8a668/topics/in-app-auto-update-and-changelog.md`

- [ ] **Step 1: Run comprehensive tests and frontend build**
  ```bash
  go test ./...
  cd frontend && pnpm run build
  ```

- [ ] **Step 2: Update Superpowers specifications and Memory**
- [ ] **Step 3: Commit and push to GitHub**
  ```bash
  git add docs/superpowers/
  git commit -m "docs(superpowers): update specs for 1-click auto-download and in-place installer"
  git push origin main
  ```
