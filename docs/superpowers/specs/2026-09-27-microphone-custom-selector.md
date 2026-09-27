# Spec: Theme-Harmonized Custom Microphone Selector & Multi-Device Discovery

## Summary
Replace the raw unstyled HTML `<select>` tag in the Settings Modal with a custom Anthropic-themed microphone selector (`MicrophoneSelectDropdown.svelte`). Enhance device enumeration with native OS hardware discovery so that all physical, wireless, and virtual audio input devices (MacBook Air built-in microphone, iPhone Continuity microphone, USB audio interfaces, Bluetooth headsets, Movavi/BlackHole capture) are cleanly detected, accurately titled, and harmoniously presented matching the chosen app theme.

---

## Pain Points & Motivation
1. **Raw HTML Select Aesthetic**:
   - The native HTML `<select>` element rendered standard OS blue highlight with inconsistent browser borders that clashed with AetherGrok's dark serif and slate design system.
2. **Missing or Generic Device Labels**:
   - Web Audio's `navigator.mediaDevices.enumerateDevices()` often outputs generic placeholder names (e.g. `Microphone 1`, `Default System Microphone`) before an explicit `getUserMedia` call is initiated or when permissions are granted at the native OS layer but not yet synced with WebView.
3. **Hardware Transparency**:
   - Users cannot see transport types (Built-in, Bluetooth, USB, Virtual) or active default flags without inspecting system settings.

---

## Technical Architecture

### 1. Native Audio Device Discovery (`pkg/permissions/audio_devices*.go`)
- On macOS, execute a quick non-blocking structured query to `system_profiler SPAudioDataType -json` or CoreAudio APIs to detect:
  - Device Name (e.g., `MacBook Air Microphone`, `WIJI FIKO Microphone`)
  - Transport (`builtin`, `bluetooth`, `usb`, `virtual`, `unknown`)
  - Default input state (`coreaudio_default_audio_input_device`)
  - Manufacturer (`Apple Inc.`, `Movavi Software Inc.`, etc.)
- Expose `GetSystemAudioInputDevices() ([]AudioDeviceInfo, error)` via Wails backend bindings in `app.go`.

### 2. Dual-Layer Merging & Auto-Refresh (`frontend/src/lib/utils/voiceRecorder.ts`)
- In `getAudioInputDevices()`:
  - Attempt to retrieve native device hardware descriptions from Go backend.
  - Query `navigator.mediaDevices.enumerateDevices()`.
  - Merge device labels by fuzzy matching or fallback mapping if the browser reports blank labels.
  - Subscribe to `navigator.mediaDevices.addEventListener('devicechange', ...)` to dynamically update the list without requiring modal reopening.

### 3. Custom UI Dropdown (`frontend/src/lib/components/layout/MicrophoneSelectDropdown.svelte`)
- Themed to match `ReasoningEffortDropdown.svelte` and `AgentModeDropdown.svelte`:
  - Trigger pill button with microphone icon, active device title, badge for default status, and chevron indicator.
  - Floating popover panel with backdrop blur, deep shadow, and border styled with `border-ant-border`.
  - Item cards displaying:
    - Transport icon (`Mic` for built-in, `Headphones` for bluetooth/usb, `Radio` for virtual/unknown).
    - Device Title and manufacturer.
    - Status badges (`Default`, `Active`).
    - Checkmark icon for the selected device.
- Keyboard accessible (`Escape` to close, click-outside dismissal).
