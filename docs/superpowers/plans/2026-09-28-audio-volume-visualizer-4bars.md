# Implementation Plan: 4-Bar Dynamic Audio Volume Equalizer & Clean Recording Pill

> **Design Direction (Anti-Slop / UI-UX Pro Max)**:
> Remove clunky text badges like `[PUSH-TO-TALK]` or `[HANDS-FREE]`. Replace with a sleek, reactive 4-bar dynamic audio equalizer driven directly by Web Audio API's `AnalyserNode`. The bars react in real time to the user's voice intensity, smoothly scaling from 3px (whisper/silence) up to 14px (active speaking), giving immediate visual proof of microphone input without distracting text clutter.

## 1. Architecture

### 1.1 Web Audio Volume Analyzer (`voiceRecorder.ts`)
- Store `private audioCtx: AudioContext | null = null;` and `private analyser: AnalyserNode | null = null;`.
- When `audioStream` is created in `startRecording()`:
  - Create `AudioContext` and `createMediaStreamSource(this.audioStream)`.
  - Create `AnalyserNode` with `fftSize = 64` and `smoothingTimeConstant = 0.4`.
  - Connect source to analyser (do not connect to `destination` to avoid speaker loopback).
- Add `getAudioVolumeLevels(): [number, number, number, number]`:
  - Samples frequency/time domain data into 4 distinct frequency bands (representing bass, voice core 1, voice core 2, treble).
  - Normalizes values between 0.0 and 1.0.
- Clean up `AudioContext` and `AnalyserNode` in `cleanupStream()`.

### 1.2 Reactive Composer Visualizer (`Composer.svelte`)
- In `Composer.svelte`:
  - Run `requestAnimationFrame` loop while `voiceState === 'recording'`.
  - Update `voiceVolumeBars = $state<[number, number, number, number]>([0.15, 0.2, 0.25, 0.15])`.
  - Remove the badge `<span ...>{dictationMode === 'hold' ? 'PUSH-TO-TALK' : 'HANDS-FREE'}</span>`.
  - Render 4 vertical animated bars:
    - Width: 2.5px
    - Border radius: `rounded-full`
    - Height: dynamically mapped `height: ${Math.max(3, Math.min(14, bar * 14))}px`
    - Color: `bg-rose-400` with subtle glow `shadow-[0_0_6px_rgba(244,63,94,0.4)]`
  - Position: right between the pulsing red recording dot and the timer `0:05`.

---

## 2. Step-by-Step Execution

1. **Step 1: Add Real-Time Analyser in `voiceRecorder.ts`**
   - Attach `AnalyserNode` to `audioStream`.
   - Provide `getVolumeLevels(): number[]` returning 4 normalized bar heights (0.0 to 1.0).
   - Ensure complete cleanup when recording stops or cancels.

2. **Step 2: Update `Composer.svelte` Recording Pill UI**
   - Remove `[PUSH-TO-TALK]` / `[HANDS-FREE]` text badge.
   - Add animation loop reading `voiceRecorder.getVolumeLevels()` while recording.
   - Insert 4 vertical equalizer bars with smooth transitions and minimum height (3px) so they remain visible as audio pill aesthetic even when silent.
   - Keep timer `Math.floor(voiceSeconds / 60):ss` and stop button.

3. **Step 3: Verification**
   - Run Vitest tests (`npx vitest run`).
   - Run `npm run check`.
   - Run `npm run build`.
