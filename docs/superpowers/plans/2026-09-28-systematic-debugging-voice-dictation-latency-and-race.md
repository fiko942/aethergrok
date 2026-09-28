# Systematic Debugging Plan: Instant-Trigger Voice Dictation & Race-Condition Resolution

## 1. Root Cause Summary (Verified)
1. **Extreme Latency ("lama banget")**:
   - Every single recording start in `voiceRecorder.startRecording()` was calling `await this.requestPermission()`, which opened and closed a throwaway `getUserMedia` stream, and THEN opened a second `getUserMedia` stream. This opened the macOS CoreAudio microphone twice (500–1500ms delay).
   - In addition, `startVoiceRecording` was awaiting `await handleMuteSystemVolume()` (AppleScript `osascript` subprocess) BEFORE opening the microphone, adding another 200–400ms delay.
2. **Hit-or-Miss Trigger ("kadang ke-trigger, kadang enggak")**:
   - Because `startVoiceRecording` took 1.5–2.5 seconds to complete, pressing and releasing a key quickly meant `hold-release` fired while `voiceState === 'checking_permission'` instead of `'recording'`.
   - In `Composer.svelte`, the handler was checking `if (voiceState === 'recording')`. Because it was still checking permission, the release event was silently dropped.
3. **Error on Stop ("Active audio recording session" / "No active audio recording session")**:
   - When a stop was invoked while `MediaRecorder` was still initializing or transitioning, `voiceRecorder.stopRecording()` immediately threw: `Error('No active audio recording session')`, showing the error modal.

---

## 2. Implementation Steps

### Step 1: Zero-Latency Microphone Startup (`voiceRecorder.ts`)
- In `voiceRecorder.startRecording()`:
  - Eliminate redundant `requestPermission()` double-open. Directly call `navigator.mediaDevices.getUserMedia(constraints)`. If it fails with NotAllowedError, then emit permission denied error.
  - Set state to `'starting'` immediately.
  - Return the initialized active stream immediately.

### Step 2: Parallel Non-Blocking Audio Ducking (`Composer.svelte`)
- In `Composer.svelte`:
  - Run `handleMuteSystemVolume()` asynchronously in parallel without blocking microphone capture so voice recording begins in under 50ms.

### Step 3: Atomic Pending-Stop Synchronization (`Composer.svelte`)
- Add `pendingStopRequested = false;` flag.
- When `hold-release` or `single-tap-unlock` is received while `voiceState === 'checking_permission'` or `voiceState === 'starting'`:
  - Set `pendingStopRequested = true`.
  - As soon as microphone finishes initializing, check `if (pendingStopRequested)`: immediately stop recording, capture audio chunk, and transcribe without dropping the turn.

### Step 4: Resilient `stopRecording()` in `voiceRecorder.ts`
- If `stopRecording()` is called while recorder is still in `starting` phase:
  - Gracefully await recorder readiness (up to 400ms) or stop active tracks cleanly without throwing an unhandled exception.
  - Always clean up streams in `finally`.

### Step 5: Snappy Shortcut Thresholds (`settings.svelte.ts` & `shortcutDetector.ts`)
- Set default `dictationHoldThresholdMs` to `200ms` so push-to-talk responds instantly.

### Step 6: Full Verification
- Run Vitest unit tests.
- Run `npm run check`.
- Run frontend build.
