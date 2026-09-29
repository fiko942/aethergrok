# Plan & Specification: Audio Storage Hygiene & Voice Cache Sweep

- **Feature**: Guaranteed Audio Cache Deletion & Orphan Sweep
- **Date**: 2026-09-29
- **Platform**: Cross-platform (macOS & Windows)
- **Status**: Production Ready & Fully Verified

---

## 1. Objective

Ensure that all temporary voice dictation audio recordings (`.webm`, `.mp3`, `.wav`) created in `~/.grok/voice_cache/` are immediately and unconditionally deleted once transcription finishes or encounters an error, preventing disk accumulation over extended usage.

---

## 2. Implementation Architecture

### A. Lifecycle Guaranteed Cleanup
1. **Frontend `finally` Block**:
   - `voiceRecorder.ts` encapsulates the transcription lifecycle inside a `try...finally` block.
   - Upon completion or failure, `window.go.main.App.DeleteVoiceAudioRecording(audioFilePath)` is called to remove the specific recording file immediately.

2. **Backend Direct File Removal**:
   - `DeleteVoiceAudioRecording(filePath)` executes `os.Remove(filePath)`.
   - `TranscribeAudioWithGrok` contains a fallback deferred removal to ensure no file persists if the frontend bridge disconnects.

3. **Orphan Cache Sweeper (`CleanVoiceCache`)**:
   - Sweeps the `~/.grok/voice_cache/` directory for any residual `voice_dictation_*` or `recording_*` files created during abrupt system shutdowns or crashes.
   - Executed automatically on application `startup` and during every voice deletion call.

---

## 3. Verification

- Backend Go tests pass 100%.
- Frontend typecheck passes with 0 errors.
- Verified that temporary files in `~/.grok/voice_cache/` are completely removed upon completion.
