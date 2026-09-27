# Custom Theme-Matched Microphone Device Picker & Multi-Device Discovery Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the unstyled raw HTML `<select>` microphone device dropdown in SettingsModal with an Anthropic-themed custom UI picker component (`MicrophoneSelectDropdown.svelte`), and implement a dual-layer hardware discovery engine (Backend Go native CoreAudio/system profiler + Frontend Web Audio API device labels) so that all connected input microphones (Built-in, USB, Bluetooth, Virtual) are accurately named, listed, and styled according to the active app theme.

**Architecture:**
1. **Backend Go Device Discovery (`pkg/permissions/audio_devices.go`, `app.go`)**:
   - Query system audio devices natively on macOS using `system_profiler SPAudioDataType -json` / CoreAudio fallbacks, reporting exact hardware names, transport types (Built-in, USB, Bluetooth), and default input status.
   - Expose `GetSystemAudioInputDevices() ([]AudioDeviceInfo, error)` to Wails bindings.
2. **Frontend Dual-Layer Device Merger (`voiceRecorder.ts`)**:
   - Merge native system device names with browser `navigator.mediaDevices.enumerateDevices()` device IDs. If permission has not yet populated browser labels, native hardware names are displayed as rich fallbacks.
   - Automatically listen for `navigator.mediaDevices.ondevicechange` to dynamically refresh the microphone device list when headphones/USB mics are plugged in or unplugged.
3. **Custom UI Dropdown Component (`MicrophoneSelectDropdown.svelte`)**:
   - Built to match `ReasoningEffortDropdown.svelte` and `AgentModeDropdown.svelte` with full Anthropic typography, theme variables (`ant-bg`, `ant-primary`, `ant-border`), device type icons (Built-in Mic, Headset/Bluetooth, Virtual), default badges, active checkmarks, and click-outside dismissal.

**Tech Stack:** Svelte 5 (Runes), Tailwind CSS, Lucide Svelte, Go 1.22 (Wails v2 bindings), Web Audio API / CoreAudio.

---

### Task 1: Backend Native Audio Input Device Discovery (`pkg/permissions/audio_devices.go`)

**Files:**
- Create: `pkg/permissions/audio_devices.go`
- Create: `pkg/permissions/audio_devices_darwin.go`
- Create: `pkg/permissions/audio_devices_other.go`
- Modify: `app.go`
- Modify: `frontend/src/app.d.ts`

**Interfaces:**
- Consumes: System Profiler / OS Audio Interfaces
- Produces: `AudioDeviceInfo { Name string, IsDefault bool, Transport string, Manufacturer string }`
- Method: `App.GetSystemAudioInputDevices() []AudioDeviceInfo`

- [ ] **Step 1: Implement `pkg/permissions/audio_devices.go`**
- [ ] **Step 2: Implement darwin specific device extraction `pkg/permissions/audio_devices_darwin.go`**
- [ ] **Step 3: Implement non-darwin fallback `pkg/permissions/audio_devices_other.go`**
- [ ] **Step 4: Expose `GetSystemAudioInputDevices` in `app.go` and add typings in `app.d.ts`**
- [ ] **Step 5: Verify with `go test ./...` and cross-compilation**

---

### Task 2: Frontend Device Merger & Auto-Refresh (`voiceRecorder.ts`)

**Files:**
- Modify: `frontend/src/lib/utils/voiceRecorder.ts`

**Interfaces:**
- Consumes: `window.go.main.App.GetSystemAudioInputDevices()` & `navigator.mediaDevices.enumerateDevices()`
- Produces: `AudioInputDevice { deviceId: string, label: string, isDefault: boolean, transport?: string }`

- [ ] **Step 1: Update `AudioInputDevice` interface with rich metadata (`transport`, `isDefault`, `manufacturer`)**
- [ ] **Step 2: Implement smart merging of native CoreAudio names with browser MediaDeviceInfo objects**
- [ ] **Step 3: Add `onDeviceChange` listener to auto-notify UI on plug/unplug**

---

### Task 3: Build Custom Anthropic-Themed `MicrophoneSelectDropdown.svelte`

**Files:**
- Create: `frontend/src/lib/components/layout/MicrophoneSelectDropdown.svelte`
- Modify: `frontend/src/lib/components/layout/SettingsModal.svelte`

**Interfaces:**
- Props: `selectedDeviceId: string`, `devices: AudioInputDevice[]`, `disabled?: boolean`, `onselect: (deviceId: string) => void`
- Produces: Fully themed custom dropdown popover with device icons (`Mic`, `Headphones`, `Radio`), active status indicators, and keyboard navigation.

- [ ] **Step 1: Create `MicrophoneSelectDropdown.svelte` with custom styling and popover portal**
- [ ] **Step 2: Integrate `MicrophoneSelectDropdown` into `SettingsModal.svelte` replacing the unstyled HTML `<select>`**
- [ ] **Step 3: Add auto-refresh and test recording on device switch**

---

### Task 4: Verification, Theme Consistency & Git Push

**Files:**
- Test & verify: `npm run build` / `vite build`
- Modify: `docs/superpowers/specs/2026-09-27-microphone-custom-selector.md`
- Modify: `docs/superpowers/plans/2026-09-27-microphone-custom-selector.md`

- [ ] **Step 1: Run frontend build and ensure zero type or CSS compilation errors**
- [ ] **Step 2: Save spec and plan to `docs/superpowers/`**
- [ ] **Step 3: Commit and push changes to Git**
