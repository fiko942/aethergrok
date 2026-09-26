# Voice-to-Text Transcription via Grok & Microphone Permissions Plan

> **Goal:** Tambahkan fitur tombol Microphone pada Prompt Box (`Composer.svelte`) untuk mendiktekan teks pemrograman dengan suara, memverifikasi izin mikrofon (macOS AVFoundation & browser permissions), merekam audio, menjalankan background transkripsi teknis menggunakan Grok di workspace aktif tanpa memunculkan sesi di UI, menempelkan hasil transkripsi ke prompt box, mengelola network resilience (auto-retry 3x & tunggu koneksi internet pulih jika offline), serta memastikan pembersihan mutlak (guaranteed cleanup) untuk temporary audio files dan temporary grok sessions baik saat sukses maupun error.

---

## 1. Arsitektur & Alur Kerja (Workflow)

```
[User clicks Mic in Composer]
       │
       ▼
[Check / Request Microphone Permission]
  ├─ macOS AVFoundation (Native Go Bridge: AVAuthorizationStatus)
  └─ Webview Navigator (navigator.mediaDevices.getUserMedia)
       │
       ▼
[Recording State in Composer]
  ├─ Pulse visual indicator & live recording timer
  ├─ Audio stream capture via MediaRecorder (audio/webm / audio/mp4 / audio/wav)
  └─ User clicks Stop / Done
       │
       ▼
[Save Temporary Audio File]
  └─ Base64 / Bytes saved to ~/.grok/voice_cache/recording_<timestamp>.mp3
       │
       ▼
[Network Resilience & Retry Loop (3x Retry + Offline Wait)]
  ├─ Check online status (navigator.onLine & backend ping/connect check)
  ├─ If offline: Wait with "Waiting for internet connection..." state until 'online' event fires
  ├─ Retry loop up to 3 attempts with exponential backoff on transient errors
       │
       ▼
[Background Transcription Run (Hidden from UI)]
  ├─ Headless Grok CLI execution in active workspace
  │   `grok --single "Here is an audio recording of a programmer talking... transcribe clearly..." --attachment <file>`
  │   OR dedicated background runner without registering into UI `sessionStore.sessions`
  └─ Extract clean transcribed text from assistant response
       │
       ▼
[Guaranteed Cleanup & Result Injection]
  ├─ If Success: Append/insert formatted transcript text into Composer prompt box
  ├─ If Fail after 3x: Show error toast to user
  └─ ALWAYS (finally block in Go & Frontend):
      ├─ Delete temporary audio file (~/.grok/voice_cache/recording_*.mp3)
      └─ Delete any created temporary Grok session directory from ~/.grok/sessions/...
```

---

## 2. Rincian Komponen & Modul yang Dikerjakan

### A. Backend Go Bridge (`pkg/permissions/` & `app.go`)
1. **macOS Microphone Permission Check & Request**:
   - `pkg/permissions/microphone_darwin.go` menggunakan Objective-C AVFoundation:
     - `[AVCaptureDevice authorizationStatusForMediaType:AVMediaTypeAudio]`
     - Nilai: `NotDetermined` (0), `Restricted` (1), `Denied` (2), `Authorized` (3).
     - Request: `[AVCaptureDevice requestAccessForMediaType:AVMediaTypeAudio completionHandler:...]`.
   - `pkg/permissions/microphone_other.go` untuk Linux/Windows fallback.
   - Info.plist update (`build/darwin/Info.plist`):
     - `NSMicrophoneUsageDescription`: `AetherGrok requires microphone access for voice-to-text dictation in the prompt composer.`
2. **Audio File Bridge & Cleanup (`app.go`)**:
   - `SaveVoiceAudioRecording(base64Data string, ext string) (string, error)`:
     - Menyimpan rekaman audio ke `~/.grok/voice_cache/recording_<timestamp>.<ext>`.
   - `DeleteVoiceAudioRecording(filePath string) error`:
     - Menghapus file rekaman audio temporary secara aman.
   - `TranscribeAudioWithGrok(workspacePath string, audioFilePath string) (string, error)`:
     - Menjalankan headless execution Grok dengan prompt teknis programmer:
       > "Kamu adalah transcriber audio programmer yang sangat akurat. Dengarkan dan baca rekaman audio teknis ini. Transkripsikan dengan jelas, gunakan istilah teknis, nama variabel, fungsi, bahasa pemrograman, dan tanda baca yang tepat dan rapi. Hanya keluarkan hasil transkrip teks murni tanpa kata pembuka atau penutup."
     - Membersihkan sesi grok yang dibuat dari disk (`DeleteGrokSessionDirectory`) dalam klausa `defer` sehingga sesi transkripsi tidak pernah tertinggal di workspace atau disk.
   - `CheckMicrophonePermission() Status`
   - `RequestMicrophonePermission() Status`
   - `OpenMicrophonePreferences() error` (Membuka `x-apple.systempreferences:com.apple.preference.security?Privacy_Microphone` di macOS).

### B. Frontend Audio Recording & Resilience (`voiceRecorder.ts` & `Composer.svelte`)
1. **Audio Recording Manager (`voiceRecorder.ts`)**:
   - Mengelola `navigator.mediaDevices.getUserMedia` dan `MediaRecorder`.
   - Format: `audio/webm;codecs=opus` atau `audio/mp4` / `audio/wav`.
   - Mengembalikan Base64 string rekaman audio.
2. **Resilience & Retry Pipeline**:
   - Mendeteksi `navigator.onLine` dan `window.addEventListener('online', ...)`.
   - Jika offline saat memulai transkripsi: Tampilkan status *"Waiting for internet connection..."* dan pause proses sampai koneksi pulih (`online` event), lalu otomatis melanjutkan transkripsi.
   - Jika terjadi error saat transkripsi: Lakukan retry otomatis hingga 3 kali (dengan delay 1s, 2s, 3s).
   - Pastikan di blok `finally`, memanggil bridge Go `DeleteVoiceAudioRecording(audioFilePath)` agar file scratch audio selalu terhapus tanpa sisa baik saat sukses maupun gagal.
3. **Tombol Mic di Prompt Box (`Composer.svelte`)**:
   - Tombol Mic dengan tooltip shortcut / fungsi.
   - State visual:
     - **Idle**: Icon Mic (`Lucide Mic`).
     - **Recording**: Icon Mic merah berdenyut (`animate-pulse`) + live timer durasi detik (`00:05`) + tombol Stop checkmark / cancel.
     - **Waiting Internet**: Status kuning "Connecting... waiting for internet".
     - **Transcribing**: Spinner loading berputar "Transcribing with Grok (Attempt 1/3)...".
   - Setelah selesai: Hasil teks disisipkan ke dalam textarea prompt box di posisi kursor dan textarea otomatis fokus.

### C. Settings Modal (`SettingsModal.svelte`)
1. **Section Baru: Voice & Dictation**:
   - Indikator Status Izin Mikrofon (macOS Permission Badge):
     - `Granted` (Badge Hijau)
     - `Denied` / `Not Determined` (Badge Kuning/Merah)
   - Tombol `Check Permission` dan `Open macOS Microphone Settings`.
   - Dropdown Pilihan Microphone Device:
     - Menampilkan daftar mikrofon yang tersedia di sistem via `navigator.mediaDevices.enumerateDevices()` (misal: *MacBook Pro Microphone*, *AirPods*, *USB Mic*).
     - Menyimpan pilihan device ID ke `settingsStore`.

---

## 3. Langkah-Langkah Eksekusi (Action Tasks)

- [ ] **Task 1**: Buat permission handler native macOS untuk Microphone (`pkg/permissions/microphone_darwin.go`, `microphone_other.go`, dan tambahkan method di `permissions.go`).
- [ ] **Task 2**: Update `build/darwin/Info.plist` dengan `NSMicrophoneUsageDescription`.
- [ ] **Task 3**: Tambahkan Go bridge methods di `app.go`:
  - `CheckMicrophonePermission() Status`
  - `RequestMicrophonePermission() Status`
  - `OpenMicrophonePreferences() error`
  - `SaveVoiceAudioRecording(base64Data, ext) (string, error)`
  - `DeleteVoiceAudioRecording(filePath) error`
  - `TranscribeAudioWithGrok(workspacePath, audioFilePath) (string, error)` (dengan auto-cleanup sesi)
- [ ] **Task 4**: Buat audio recording & resilience utility di frontend (`frontend/src/lib/utils/voiceRecorder.ts`):
  - Audio capture, retry 3x loop, wait-for-online handler, audio chunk conversion, and guaranteed cleanup invocation.
- [ ] **Task 5**: Integrasikan UI Microphone di Prompt Box (`frontend/src/lib/components/chat/Composer.svelte`):
  - Record button, active pulse indicator, timer, transcribing status, dan auto-paste ke textarea.
- [ ] **Task 6**: Tambahkan section "Voice & Dictation" di `frontend/src/lib/components/layout/SettingsModal.svelte`.
- [ ] **Task 7**: Update `app.d.ts` dan jalankan verifikasi `svelte-check` serta `npm run build`.
