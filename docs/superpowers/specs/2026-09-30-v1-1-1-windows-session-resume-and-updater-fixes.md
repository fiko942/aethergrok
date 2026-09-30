# Spec: AetherGrok v1.1.1 - Windows Multi-Prompt Session Resume, Installer Elevation, Downloads Directory & Light Mode UI Polish

## Overview
AetherGrok version 1.1.1 addresses critical multi-prompt session fragmentation on Windows, Windows auto-installer elevation and file-locking issues, standardizes update package downloads to the user's `Downloads` folder, and delivers light mode UI accessibility enhancements for Plan Mode and Active Plan Tracker.

---

## Architecture & Implementation Details

### 1. Cross-Platform Session Resumption (`pkg/grokrunner`)
- **Problem**: In Windows, path encoding variations (`%5C` vs `%2F` and drive letters `C:` vs `c:`) caused `os.Stat(sessionFolder)` to fail when looking up existing session directories under `~/.grok/sessions/`. The runner fell back to passing `--session-id <UUID>` (creating a brand-new turn session) rather than `--resume <UUID>`, breaking multi-turn context continuity in the same chat tab.
- **Solution**:
  - `session_scanner.go`: `ResolveWorkspaceSessionsDir` verifies both slash and backslash URL-encoded path variants and matches decoded folder names case-insensitively.
  - `runner.go`: Prioritizes `--resume <UUID>` whenever `req.Options.GrokSessionID` is provided from the frontend, ensuring every subsequent prompt in the same tab continues the existing Grok conversation turn.

### 2. Auto-Installer & User Downloads Folder (`pkg/updater`)
- **Problem**: Windows auto-install update executed with silent flag `/S` was blocked when writing to `Program Files` due to lack of UAC elevation and race conditions with file locking during application exit. Updates were also saved in hidden temp directories.
- **Solution**:
  - `installer.go`: Added `GetUserDownloadsDir()` targeting the standard user `Downloads` directory (`~/Downloads`) across Windows and macOS.
  - Windows update execution uses a detached batch runner with PID termination wait to ensure the old binary unlocks cleanly before launching the NSIS installer.

### 3. Light Mode Accessibility & Active Plan Tracker Reset
- **UI Contrast**:
  - `AgentModeDropdown.svelte`: Enhanced Plan mode pill selector with high-contrast text and border tokens (`text-violet-700 dark:text-violet-300`).
  - `ToolCallCard.svelte`: Aligned "Plan Objective / Context" heading and description with active theme tokens.
  - `FloatingPlanTracker.svelte` & `PlanPanel.svelte`: Corrected text visibility and theme-aware styling for active execution plans.
- **Plan Lifecycle Reset**:
  - `plan.svelte.ts` & `App.svelte`: Implemented `planStore.resetPlan(sessionId)` called on every new prompt submission so the sidebar and floating plan tracker reset cleanly for new turn tasks.
