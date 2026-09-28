# In-App Auto-Update & Version Changelog History Architecture Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Provide an in-app Update & Version History system backed by the GitHub Releases API (and central `changelog.json`), allowing users to check for new releases, view complete multi-version changelogs with highlights and commit histories directly within the UI, download/install updates for their platform/architecture, and receive non-intrusive update notification badges.

**Architecture:**
1. **Backend Updater Engine (`pkg/updater/`)**:
   - Fetches releases from `https://api.github.com/repos/fiko942/grok-build/releases`.
   - Parses SemVer tags, release dates, notes, and platform-specific download assets (macOS DMG ARM64/Intel, Windows Setup/Portable x64/ARM64).
   - Exposes methods via Wails bridge: `CheckForUpdates(currentVersion string)`, `GetChangelogHistory()`, `DownloadAndInstallUpdate(assetUrl string)`.
2. **Frontend Design & UX (`UpdatesTab.svelte` & `SettingsModal.svelte`)**:
   - Follows strict Design System tokens (`lib/antd/tokens.ts`, `anthropic` typography, theme-aware palettes).
   - Dedicated **"Updates & Changelog"** tab in SettingsModal with:
     - Live Update Status Header (Current Version, Latest Version badge, "Check for Updates" action with animated spinner).
     - New Update Callout Banner (when update is available, with direct 1-click Download / Open DMG / Setup action).
     - Rich Chronological Version History: expandable cards for each release displaying release title, date, highlights list, commit diffs, and direct package download links.
   - Global notification indicator (subtle badge on Settings gear / header when a newer version is detected).

**Tech Stack:** Go (HTTP, SemVer, JSON), Svelte 5 (Runes `$state`, `$derived`), Tailwind CSS, Lucide icons, Ant Design design tokens.

---

### File Structure Map
- **Backend**:
  - `pkg/updater/updater.go` - Core update checker, GitHub API client, changelog parser, and asset matcher.
  - `pkg/updater/updater_test.go` - Unit tests for SemVer comparisons and release asset matching.
  - `app.go` - Expose `CheckForUpdates`, `GetChangelogHistory`, `OpenExternalURL` methods to Wails bridge.
- **Frontend**:
  - `frontend/src/lib/components/settings/UpdatesTab.svelte` - Visual updater & changelog history component.
  - `frontend/src/lib/components/layout/SettingsModal.svelte` - Register the Updates tab and tab badge.
  - `frontend/src/lib/stores/updater.svelte.ts` - Reactive store managing update state, background periodic checks, and release cache.
  - `frontend/src/App.svelte` - Mount periodic background update checker and header update badge.

---

### Task 1: Implement Backend Updater Package (`pkg/updater`)

**Files:**
- Create: `pkg/updater/updater.go`
- Create: `pkg/updater/updater_test.go`
- Modify: `app.go`
- Modify: `frontend/src/app.d.ts`

- [ ] **Step 1: Create `pkg/updater/updater.go`**
  Implement Go types:
  ```go
  type ReleaseAsset struct {
      Name        string `json:"name"`
      Size        int64  `json:"size"`
      DownloadURL string `json:"downloadUrl"`
      ContentType string `json:"contentType"`
  }

  type ReleaseInfo struct {
      Version     string         `json:"version"`
      TagName     string         `json:"tagName"`
      Title       string         `json:"title"`
      PublishedAt string         `json:"publishedAt"`
      Body        string         `json:"body"`
      Highlights  []string       `json:"highlights"`
      Assets      []ReleaseAsset `json:"assets"`
      IsLatest    bool           `json:"isLatest"`
  }

  type UpdateCheckResult struct {
      UpdateAvailable bool         `json:"updateAvailable"`
      CurrentVersion  string       `json:"currentVersion"`
      LatestVersion   string       `json:"latestVersion"`
      LatestRelease   *ReleaseInfo `json:"latestRelease"`
      AllReleases     []ReleaseInfo`json:"allReleases"`
      PlatformAsset   *ReleaseAsset`json:"platformAsset"`
      CheckedAt       string       `json:"checkedAt"`
  }
  ```
  Implement helper functions:
  - `CompareSemVer(v1, v2 string) int` (supports `v` prefix, returns 1 if v1 > v2, -1 if v1 < v2, 0 if equal).
  - `MatchPlatformAsset(assets []ReleaseAsset, osName, arch string) *ReleaseAsset` (matches `.dmg` for darwin arm64/amd64 and `.exe` / `-setup.exe` for windows).
  - `FetchGitHubReleases(repo string) ([]ReleaseInfo, error)`.

- [ ] **Step 2: Create unit tests in `pkg/updater/updater_test.go`**
  Test SemVer comparison edge cases (`1.0.1` vs `1.0.0`, `v1.2.0` vs `1.1.9`, `1.0.0-rc1`), asset matching for darwin/arm64, darwin/amd64, windows/amd64, and windows/arm64.

- [ ] **Step 3: Expose updater methods in `app.go`**
  Add `CheckForUpdates(currentVersion string) (*updater.UpdateCheckResult, error)` and `GetChangelogHistory() ([]updater.ReleaseInfo, error)` to `App`.

- [ ] **Step 4: Run Go test suite**
  ```bash
  go test -v ./pkg/updater/...
  ```

- [ ] **Step 5: Commit backend updater package**
  ```bash
  git add pkg/updater/ app.go frontend/src/app.d.ts
  git commit -m "feat(updater): implement pure-Go GitHub release update checker and changelog parser"
  ```

---

### Task 2: Create Reactive Updater Store in Svelte (`updater.svelte.ts`)

**Files:**
- Create: `frontend/src/lib/stores/updater.svelte.ts`

- [ ] **Step 1: Implement Svelte 5 Rune-based `updaterStore`**
  State properties:
  - `checking`: boolean
  - `updateAvailable`: boolean
  - `currentVersion`: string
  - `latestVersion`: string
  - `latestRelease`: ReleaseInfo | null
  - `allReleases`: ReleaseInfo[]
  - `matchedAsset`: ReleaseAsset | null
  - `lastChecked`: string | null
  - `error`: string | null
  Methods:
  - `checkForUpdates(silent?: boolean)`
  - `openDownload(url?: string)`
  - `initPeriodicCheck()`

- [ ] **Step 2: Commit updater store**
  ```bash
  git add frontend/src/lib/stores/updater.svelte.ts
  git commit -m "feat(frontend): create reactive updater store for GitHub releases"
  ```

---

### Task 3: Build Premium "Updates & Changelog" UI Component (`UpdatesTab.svelte`)

**Files:**
- Create: `frontend/src/lib/components/settings/UpdatesTab.svelte`
- Modify: `frontend/src/lib/components/layout/SettingsModal.svelte`

- [ ] **Step 1: Create `UpdatesTab.svelte`**
  Design with Ant Design Dark/Light tokens & Anthropic serif typography:
  - **Status Card**:
    - Current installed version pill (`v1.0.1`).
    - "Up to date" or "Update Available" badge with pulse animation.
    - "Check for Updates" button with loading spinner (`RotateCw`).
    - Last checked timestamp.
  - **Available Update Banner** (when `updateAvailable === true`):
    - Highlighted card with primary color accent border.
    - Version jump: `v1.0.1 → v1.1.0`.
    - Direct download action button (`Download Installer for macOS (Apple Silicon)` / `Windows Setup`).
    - Quick summary of key features.
  - **Version History & Changelog Accordion**:
    - Chronological release timeline list.
    - Release header: Tag badge, release date (formatted), release title, expand/collapse chevron.
    - Expanded content:
      - Bulleted highlights.
      - Commit log preview (hash + commit message).
      - Direct download package links table.

- [ ] **Step 2: Integrate Updates Tab into `SettingsModal.svelte`**
  - Add `'updates'` to `TabKey` and sidebar tab navigation with `Sparkles` or `DownloadCloud` icon.
  - Show a small colored indicator badge next to the "Updates" tab when an update is available.
  - Replace static version footer with clickable link to open the Updates tab.

- [ ] **Step 3: Add Update Indicator in Main Header / Settings Button**
  - In `App.svelte`, trigger initial update check on mount.
  - If an update is available, show a subtle glowing dot on the Settings icon in the sidebar footer.

- [ ] **Step 4: Commit frontend UI components**
  ```bash
  git add frontend/src/lib/components/settings/UpdatesTab.svelte frontend/src/lib/components/layout/SettingsModal.svelte frontend/src/App.svelte
  git commit -m "feat(settings): add Updates & Version Changelog history tab with live GitHub release integration"
  ```

---

### Task 4: End-to-End Verification & Documentation

**Files:**
- Create: `docs/superpowers/specs/2026-09-28-in-app-auto-update-and-changelog.md`

- [ ] **Step 1: Verify TypeScript & Frontend Build**
  ```bash
  cd frontend && pnpm run build
  ```

- [ ] **Step 2: Verify Go compilation**
  ```bash
  go test ./...
  ```

- [ ] **Step 3: Document feature in Superpowers Spec & Workspace Memory**
  Save specs and topic notes for future reference.

- [ ] **Step 4: Commit and push to GitHub**
  ```bash
  git add docs/superpowers/
  git commit -m "docs(superpowers): add in-app auto-update and changelog specifications"
  git push origin main
  ```
