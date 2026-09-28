# Voice Dictation Dual-Mode Shortcut & System Audio Mute Ducking Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement dual-mode voice dictation shortcuts (double-tap to lock hands-free, hold to push-to-talk) with automatic system audio mute ducking (0% volume during recording, restored to exact original level afterwards) and design-system polished UI indicators in AetherGrok Desktop Studio.

**Architecture:** 
- Go backend (`pkg/system/volume.go`) provides cross-platform system audio muting and restoring (macOS AppleScript `get/set volume` and Windows PowerShell CoreAudio fallback), exposed via Wails IPC in `app.go`.
- Frontend settings store (`settings.svelte.ts`) manages persistent preferences for dictation shortcut (`Fn` default), audio ducking toggle, and hold detection threshold.
- Global key listeners and Composer controller manage double-tap vs hold lifecycle, coordinate system audio muting, stream audio through `voiceRecorder.ts`, and insert Grok transcriptions cleanly into Composer.
- Visual tokens and components follow AetherGrok Design System (Ant Design studio tokens, subtle pulsing indicators, and custom key recorder).

**Tech Stack:** Go (Wails v2, AppleScript / CoreAudio), Svelte 5 (Runes `$state`, `$derived`, `$effect`), TypeScript, Tailwind CSS with Ant Design Studio Tokens (`--ant-*`).

## Global Constraints
- Avoid hardcoded color hex values; use design tokens (`var(--ant-*)`, `text-ant-primary`, `bg-ant-primary-bg`, `border-ant-border`).
- Ensure audio volume is restored without fail, even on errors, cancellation, or app blur/focus loss.
- Double-tap detection interval is capped at 350ms; hold push-to-talk threshold is 300ms.
- Default shortcut is `Fn` (or customizable through `KeyRecorderModal` in Settings).

---

### Task 1: Go Backend System Volume Mute & Restore Service

**Files:**
- Create: `pkg/system/volume.go`
- Create: `pkg/system/volume_darwin.go`
- Create: `pkg/system/volume_windows.go`
- Create: `pkg/system/volume_test.go`
- Modify: `app.go`
- Modify: `frontend/src/app.d.ts`

**Interfaces:**
- Consumes: OS command execution (`exec.Command`)
- Produces: 
  - `MuteSystemVolume() (SystemVolumeState, error)`
  - `RestoreSystemVolume(state SystemVolumeState) error`
  - Wails bindings: `App.MuteSystemVolume() (*VolumeMuteResult, error)` and `App.RestoreSystemVolume(prevVolume int, wasMuted bool) error`

- [ ] **Step 1: Write unit test in `pkg/system/volume_test.go`**
- [ ] **Step 2: Run test to verify it fails**
  Run: `go test -v ./pkg/system`
- [ ] **Step 3: Implement `pkg/system/volume.go`, `volume_darwin.go`, `volume_windows.go`**
  - Implement volume query and 0% mute with original state preservation.
  - Implement restore to previous volume and mute status.
- [ ] **Step 4: Expose methods in `app.go` and update `frontend/src/app.d.ts`**
  - Wire `MuteSystemVolume` and `RestoreSystemVolume` to `App` struct.
- [ ] **Step 5: Run tests to verify they pass**
  Run: `go test -v ./pkg/system`
- [ ] **Step 6: Commit changes**
  Run: `git commit -m "feat(system): add cross-platform system audio volume mute and restore service"`

---

### Task 2: Settings Store Extension for Dictation & Audio Ducking

**Files:**
- Modify: `frontend/src/lib/stores/settings.svelte.ts`
- Modify: `pkg/storage/storage.go`

**Interfaces:**
- Consumes: Backend storage JSONL
- Produces: 
  - `settingsStore.dictationShortcut` (string, default: `'Fn'`)
  - `settingsStore.dictationMuteSystemAudio` (boolean, default: `true`)
  - `settingsStore.dictationHoldThresholdMs` (number, default: `300`)

- [ ] **Step 1: Add dictation settings fields to `pkg/storage/storage.go` struct**
- [ ] **Step 2: Add reactivity and hydration in `frontend/src/lib/stores/settings.svelte.ts`**
- [ ] **Step 3: Test persistence by saving and re-reading settings**
- [ ] **Step 4: Commit changes**
  Run: `git commit -m "feat(settings): add dictation shortcut and audio ducking persistence"`

---

### Task 3: Dual-Mode Shortcut Engine (Double-Tap & Push-to-Talk)

**Files:**
- Create: `frontend/src/lib/utils/shortcutDetector.ts`
- Create: `frontend/src/lib/utils/shortcutDetector.test.ts`
- Modify: `frontend/src/App.svelte`

**Interfaces:**
- Consumes: Browser `KeyboardEvent` (`keydown` & `keyup`)
- Produces: Callback events `{ type: 'hold-start' }`, `{ type: 'hold-release' }`, `{ type: 'double-tap-lock' }`, `{ type: 'single-tap-unlock' }`

- [ ] **Step 1: Write test for `shortcutDetector.ts`**
  - Tests timing under 350ms for double-tap.
  - Tests timing over 300ms for hold push-to-talk release.
- [ ] **Step 2: Implement `shortcutDetector.ts`**
- [ ] **Step 3: Connect shortcut detector in `App.svelte`**
  - Bind global key listeners for dictation shortcut.
- [ ] **Step 4: Commit changes**
  Run: `git commit -m "feat(shortcuts): add double-tap lock and push-to-talk keyboard detector"`

---

### Task 4: Composer Integration & System Audio Ducking Lifecycle

**Files:**
- Modify: `frontend/src/lib/components/chat/Composer.svelte`
- Modify: `frontend/src/lib/utils/voiceRecorder.ts`

**Interfaces:**
- Consumes: `shortcutDetector.ts`, Wails `MuteSystemVolume` & `RestoreSystemVolume`
- Produces: Automatic volume ducking on recording start, exact restore on completion/cancel, text insertion in Composer.

- [ ] **Step 1: Add volume ducking state manager in `Composer.svelte`**
  - Save `originalVolume` and `wasMuted` when starting recording if `dictationMuteSystemAudio` is true.
  - Guarantee call to `RestoreSystemVolume` in `finally` block of `stopRecording` and `cancelRecording`.
- [ ] **Step 2: Connect double-tap lock vs hold release triggers to `Composer.svelte`**
  - Lock mode shows persistent recording pill with stop button.
  - Push-to-talk mode automatically stops recording and transcribes upon key release.
- [ ] **Step 3: Handle edge cases (window blur, modal popups, error fallback)**
  - Ensure volume is unmuted if user switches apps during dictation.
- [ ] **Step 4: Commit changes**
  Run: `git commit -m "feat(dictation): integrate system volume ducking with composer voice workflow"`

---

### Task 5: UI & Design System Polish (Settings Modal & Composer Pill)

**Files:**
- Modify: `frontend/src/lib/components/layout/SettingsModal.svelte`
- Modify: `frontend/src/lib/components/chat/Composer.svelte`
- Modify: `frontend/src/lib/antd/Tooltip.svelte`

**Interfaces:**
- Consumes: Ant Design studio tokens
- Produces: 
  - Dictation shortcut configuration & audio ducking toggle in **Settings > Voice & Dictation**.
  - Shortcut listed in **Settings > Shortcuts** table.
  - Polished pulsating recording pill with badge indicating mode (`Hands-Free` vs `Push-to-Talk`).

- [ ] **Step 1: Add Dictation Shortcut and Mute Audio Toggle in `SettingsModal.svelte` (Voice Tab)**
- [ ] **Step 2: Add Dictation entry in `SettingsModal.svelte` (Shortcuts Tab)**
- [ ] **Step 3: Refine Composer recording badge styling with Design System tokens**
- [ ] **Step 4: Commit changes**
  Run: `git commit -m "feat(ui): add dictation shortcut controls and polished recording badge"`

---

### Task 6: End-to-End Browser & System Verification

- [ ] **Step 1: Build and run test suite**
  Run: `go test -v ./...` and `npm run check` in frontend
- [ ] **Step 2: Verify in application UI**
  - Test double-tap `Fn` lock and single tap stop.
  - Test hold `Fn` push-to-talk.
  - Test system audio muting (volume becomes 0% while recording) and restore back to original level.
  - Test shortcut customization in Settings.
- [ ] **Step 3: Commit and finalize**
