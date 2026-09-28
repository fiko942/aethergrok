# In-App Auto-Update & Version Changelog History Architecture

## Overview
AetherGrok Desktop Studio incorporates an in-app Update & Release Changelog system backed by GitHub Releases API (`https://api.github.com/repos/fiko942/grok-build/releases`) and the central repository `changelog.json`. Users can inspect their current version, check for updates with live feedback, view chronological version histories with bullet highlights and commit diffs, and download platform-matched installers in 1 click.

## Architecture

### 1. Pure-Go Backend Updater Engine (`pkg/updater/updater.go`)
- **SemVer Comparison (`CompareSemVer`)**: Robust semantic versioning comparator supporting `v` prefixes, prerelease tags, and numeric hierarchy (`v1.0.1` vs `1.0.0`, `1.10.0` vs `1.9.0`).
- **Platform Asset Matcher (`MatchPlatformAsset`)**:
  - `darwin` / `arm64`: Matches Apple Silicon DMG installer (`*macOS-arm64.dmg`).
  - `darwin` / `amd64`: Matches Intel x64 DMG installer (`*macOS-amd64.dmg` or `*macOS-x64.dmg`).
  - `windows` / `amd64`: Matches NSIS Setup installer (`*windows-amd64-setup.exe` or `*installer*.exe`).
  - `windows` / `arm64`: Matches Windows ARM64 Setup installer (`*windows-arm64-setup.exe`).
- **Resilient Fallback**: If GitHub API fails due to network outage or rate limiting, the engine falls back to reading `changelog.json` in the current working directory, ensuring the changelog history remains browseable offline.
- **Wails Bridge (`app.go`)**:
  - `CheckForUpdates(currentVersion string) (*updater.UpdateCheckResult, error)`
  - `GetChangelogHistory() ([]updater.ReleaseInfo, error)`
  - `OpenExternalURL(targetUrl string) error`

### 2. Reactive Frontend Store (`updater.svelte.ts`)
- Manages reactive state via Svelte 5 `$state`:
  - `checking`: boolean
  - `updateAvailable`: boolean
  - `currentVersion`: string
  - `latestVersion`: string
  - `latestRelease`: ReleaseInfo | null
  - `allReleases`: ReleaseInfo[]
  - `matchedAsset`: ReleaseAsset | null
  - `lastChecked`: formatted date
  - `error`: error message if any
- Auto-initializes on startup with a gentle 3-second delay, then executes periodic background checks every 1 hour.

### 3. Design System & User Interface (`UpdatesTab.svelte`)
- Conforms to Ant Design tokens (`tokens.ts`) and Anthropic Serif typography.
- **Live Status Header**:
  - Current version pill, "Up to date" green badge or pulsing amber "Update Available" badge.
  - "Check for Updates" button with animated `RotateCw` spinner.
- **Update Hero Banner**:
  - Displayed when `updateAvailable === true`.
  - Visual version leap (`v1.0.1 → v1.1.0`), release title, highlights preview, and 1-click Download button pointing to the matched platform asset.
- **Version History & Changelog Accordion**:
  - Expandable cards for each release.
  - Formatted release notes, bulleted highlights, and a download grid for all platform packages (`.dmg`, `.exe`, `.zip`).
- **Global Badging**:
  - Small notification dots on the Settings button in the sidebar footer and header version pill when a new release is detected.
