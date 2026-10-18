# AetherGrok v1.1.13 Implementation Plan & Superpowers Ledger

> **For agentic workers:** Tracked under superpowers:systematic-debugging & superpowers:writing-plans.

## 1. Overview & Objectives

This release (v1.1.13) packages critical stability, performance, and UI refinements:
1. **Accurate Session-Bound Turn Timer & Grok CLI Waiting State:**
   - Isolated `turnStartTime` and `lastTurnDurationMs` per session tab.
   - Dynamic Grok CLI style `Waiting for response... [timer]` indicator next to the Stop button in `Composer.svelte`.
   - Contextual status text (`Executing shell command...`, `Waiting for response...`, `Grok is formulating response...`) in `MessageList.svelte`.
2. **DOM Active Turn Windowing & Automatic Background Flush:**
   - Dynamic binding to `settingsStore.activeWindowTurnCount` (4, 10, 30 turns).
   - Auto-flush mechanism (`flushEarlierTurnsToWindow`) when arriving at bottom or when turn finishes, conserving RAM/GPU.
   - User gesture tracking (`onwheel`, `ontouchmove`) preventing accidental auto-load cascades when streaming new chunks.
3. **Command Description & Structured IN/OUT Inspector:**
   - Display `description` parameter on tool cards (e.g. `Run: extract task brief for Task 2`).
   - Separate, copyable `IN` (full script) and `OUT` (clean terminal result) panels with keyboard shortcuts (`c` and `i`).
4. **macOS Force-Close & Window Teardown Immunity:**
   - Fail-safe 1500ms watchdog for `screencapture` ensuring window restoration.
   - Global keyboard interception for `Cmd+H` (`[NSApp hide:]`) and `Cmd+W`.

---

## 2. Release Steps

- [x] Step 1: Update version constants in `wails.json`, `app.go`, `changelog.json`, and `frontend/package.json` to `1.1.13`.
- [x] Step 2: Validate local test suites: `go test ./...` and `bun test test/dom-window-flush.test.ts test/turn-timer-isolation.test.ts`.
- [x] Step 3: Compile and verify local production build for macOS Apple Silicon.
- [x] Step 4: Commit all changes, tag `v1.1.13`, and push to `origin/main`.
- [x] Step 5: Monitor GitHub Actions multi-platform release workflow (`release.yml`) until all jobs (Windows amd64/arm64, macOS arm64/amd64/universal, and GitHub Release creation) pass.
- [x] Step 6: Verify GitHub Release availability and asset attachments.
- [x] Step 7: Re-install verified production build to `/Applications/AetherGrok.app`.
