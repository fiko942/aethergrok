# Plan: AetherGrok v1.1.1 - Windows Session Resume, Updater Detached Elevation & Light Mode Polish

## Goal
Release AetherGrok v1.1.1 featuring:
1. Resilient multi-turn session continuation on Windows and macOS.
2. Auto-installer reliability on Windows with standard user Downloads destination.
3. Light mode visual enhancements for Plan Mode pills and plan objective cards.
4. Auto-reset of Active Plan Tracker upon each new prompt submission.
5. Multi-platform GitHub Actions build focused on macOS (DMG) and Windows (NSIS Setup).

## Tasks
- [x] Fix session directory resolution in `pkg/grokrunner/session_scanner.go`
- [x] Prioritize `--resume` in `pkg/grokrunner/runner.go`
- [x] Save updates directly to user `Downloads` directory in `pkg/updater/installer.go`
- [x] Implement detached batch runner for Windows NSIS setup in `pkg/updater/installer.go`
- [x] Adjust Plan mode pill and objective card styling in `AgentModeDropdown.svelte` and `ToolCallCard.svelte`
- [x] Implement `planStore.resetPlan` and invoke on new prompt submission in `App.svelte`
- [x] Update GitHub Actions release workflow to streamline macOS and Windows builds
- [ ] Bump version to 1.1.1 across `wails.json`, `app.go`, `frontend/package.json`, and `changelog.json`
- [ ] Execute release process via `npm run release` and monitor GitHub Action status
