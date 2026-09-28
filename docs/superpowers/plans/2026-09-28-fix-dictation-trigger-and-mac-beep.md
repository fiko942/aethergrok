# Plan: Fix Dictation Shortcut Trigger & Eliminate macOS System Beep

## Problem Analysis
1. **Shortcut Not Triggering**:
   - In `App.svelte`, `<Composer>` is only mounted when `sessionStore.activeSession` is truthy. In the user's screenshot (`image-504423e5...`), all tabs were closed ("No Active Session"). If no session is active or composer is unmounted, dictation events were ignored.
   - If focus is inside a textarea or input, `matchesShortcut` or keydown handler was either letting normal typing consume `\` or not intercepting properly.
   - For `\` (Backslash), key events in different layouts can have `e.key === '\\'` or `e.code === 'Backslash'`.

2. **System Beep Sound**:
   - On macOS WKWebView / Wails, pressing any key that is not consumed by an input field and has no `e.preventDefault()` causes AppKit's `[NSResponder noResponderFor:]` to call `NSBeep()`.
   - Also, in `App.svelte` line 876: `(e.key.toLowerCase() === 'b' || e.key === '\\')` was intercepting `\` when Cmd/Ctrl was held for sidebar collapse, but when unhandled, it fell through to macOS beep.
   - Whenever the dictation shortcut key is detected on `keydown` or `keyup`, calling `e.preventDefault()` and `e.stopPropagation()` immediately stops macOS from ringing the alert sound bell.

## Implementation Steps

### 1. Global Dictation Controller & Active Session Auto-Creation
- In `App.svelte`:
  - When a dictation trigger occurs (`hold-start` or `double-tap-lock`):
    - If there is no active session (`!sessionStore.activeSession`), automatically create a new session in the active workspace and switch to it immediately so the Composer mounts and receives the voice stream.
  - In `handleGlobalKeyDown` and `handleGlobalKeyUp`:
    - Check if `shortcutDetector.matchesShortcut(e, settingsStore.dictationShortcut)`.
    - If it matches: **immediately call `e.preventDefault()` and `e.stopPropagation()`**. This directly eliminates the macOS system beep and prevents inserting `\` into the text.

### 2. ShortcutDetector Robustness
- In `shortcutDetector.ts`:
  - Ensure `matchesShortcut` handles both `e.key === '\\'` and `e.code === 'Backslash'`.
  - Fix double-tap logic:
    - Tap 1: starts a short timer (300ms). If a second tap arrives within 350ms, lock state is entered (`double-tap-lock`).
    - Hold: If key is kept pressed beyond 250ms, immediately emit `hold-start`. On release, emit `hold-release`.
    - Unlock: If currently in `locked` state, pressing the key once immediately unlocks and emits `single-tap-unlock`.

### 3. Audio Ducking & Guaranteed Restoration
- Verify volume ducking activates when recording starts and restores when recording stops or cancels.

### 4. Verification
- Test in Vitest.
- Run `npm run check`.
- Verify absence of macOS beep when pressing `\`.
- Verify hold vs double-tap behavior.
