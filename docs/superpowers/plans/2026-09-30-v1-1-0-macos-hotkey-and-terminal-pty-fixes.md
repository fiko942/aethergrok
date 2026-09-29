# Plan & Architecture: v1.1.0 Release - macOS Voice Dictation Key State Watchdog & Terminal PTY Fixes

**Date:** 2026-09-30  
**Status:** Completed & Shipped  
**Release Tag:** `v1.1.0`  
**Repository:** `https://github.com/fiko942/aethergrok`

---

## 1. Overview & Problem Context

In earlier versions, macOS users experienced two critical issues that degraded the desktop experience compared to Windows:

### Problem 1: Voice & Dictation Hotkey Intermittent Triggers & Stuck Key State
1. **Missed Triggers**: `CFRunLoopRun()` in `startEventTap()` was blocking synchronous execution in Go goroutines before setting `darwinRunning = true`, causing desynchronization during shortcut updates and duplicate event tap allocations.
2. **Stuck Pressed State on App Focus Switch**: When a user held the global dictation shortcut key (e.g., `\`) and switched to another application or released the key outside AetherGrok, macOS delivered the `kCGEventKeyUp` event to the target application or filtered it outside the session tap. AetherGrok remained stuck in `isDown = true`, requiring the user to press stop manually.
3. **Event Dropping under Load**: High-frequency CoreGraphics events caused non-blocking select channel writes (`select { case eventChan <- fn: default: }`) to drop `down` or `up` key state transitions.

### Problem 2: macOS Terminal Blank Screen & Non-Interactive PTY
1. When opening the built-in terminal panel on macOS, the terminal appeared as a blank black box that did not display shell prompts or accept keyboard input.
2. **Root Cause**: `pkg/terminal/pty_unix.go` explicitly set `cmd.SysProcAttr = &syscall.SysProcAttr{ Setpgid: true }`. In macOS (Darwin/BSD), `creack/pty.StartWithSize` configures the PTY slave and session control internally (`setsid`, `login_tty`, and `TIOCSCTTY`). Combining explicit `Setpgid: true` with PTY slave assignment caused kernel `fork/exec /bin/zsh: operation not permitted` (`EPERM`), failing the process spawn silently.

---

## 2. Engineering Changes

### 1. macOS Voice & Dictation Engine (`pkg/hotkey/hotkey_darwin.go`)
- **Hardware Key State Watchdog**: Added native C binding `isPhysicalKeyDown(keycode)` using `CGEventSourceKeyState(kCGEventSourceStateCombinedSessionState, keycode)`.
- **Active Safety Ticker**: Whenever `isDown == true`, an asynchronous 60ms ticker polls the physical key state. If the key is released outside the application or during focus shifts, the watchdog automatically transitions `isDown = false` and enqueues an `"up"` event.
- **Drop-Resistant Dispatcher**: Implemented `enqueueEvent(fn)` with background fallback goroutines to eliminate dropped hotkey events during channel saturation.
- **RunLoop Synchronization**: Added `onDarwinTapStarted()` C callback export to ensure Go receives the active running status before `CFRunLoopRun()` enters its event loop.

### 2. macOS Terminal Pseudo-Terminal (`pkg/terminal/pty_unix.go`)
- Removed `Setpgid: true` from `SysProcAttr` on Unix/Darwin, allowing `creack/pty` to cleanly initialize the slave PTY and spawn `/bin/zsh` without permission errors.
- Verified that process tree termination and signal forwarding (`SIGINT`, `SIGTERM`, `SIGKILL`) remain robust using dynamically retrieved PGIDs via `syscall.Getpgid(pid)`.

### 3. Application Version Bump to v1.1.0
- Updated `wails.json` version to `1.1.0`.
- Updated `app.go` `AppVersion` constant to `"1.1.0"`.

---

## 3. Verification & Evidence
- **Backend Test Suite (`go test -v ./...`)**:
  - `pkg/hotkey`: `TestDarwinWatchdogAutoRelease` and `TestDarwinKeyEventDispatch` passed 100%.
  - `pkg/terminal`: `TestTerminalManager_LifecycleAndProcessGroup`, `TestTerminalManager_CloseSessionTerminals`, `TestTerminalManager_ClsCommand`, and `TestTerminalManager_InterruptAndKill` all passed.
- **Frontend Test Suite (`pnpm test`)**: 235 test files and 5,329 Vitest tests passed with 0 regressions.
