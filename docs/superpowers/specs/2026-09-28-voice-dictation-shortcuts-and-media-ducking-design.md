# Specification: Voice Dictation Dual-Mode Shortcut & System Audio Mute Ducking

## 1. Overview
Fitur ini menambahkan pengalaman input suara (*Voice Dictation*) profesional di AetherGrok Desktop Studio dengan:
1. **Dual-Mode Shortcut (Push-to-Talk & Lock-to-Talk)**:
   - **Double-Tap**: Menekan shortcut 2 kali secara cepat (default: tombol `Fn` atau shortcut yang dikonfigurasi) akan mengunci perekaman (*lock / hands-free mode*). Tekan 1 kali lagi untuk menyelesaikan dan mentranskripsi.
   - **Hold / Push-to-Talk**: Menahan tombol shortcut akan mulai merekam suara, dan melepaskan tombol akan langsung menghentikan perekaman serta mentranskripsi ke dalam Composer.
2. **System Audio Mute Ducking**:
   - Saat proses dictation dimulai, volume sistem operasi di-mute (0%) secara otomatis agar suara musik, video, atau speaker tidak masuk ke mikrofon.
   - Setelah dictation selesai atau dibatalkan, volume sistem dikembalikan persis ke tingkat volume semula (*unmute & restore original volume level*).
3. **UI / UX Polish (Design System Alignment)**:
   - Floating recording indicator / pulsing recording pill dengan visual waktu dan status hands-free vs push-to-talk.
   - Konfigurasi shortcut dictation, perekam tombol (*KeyRecorderModal*), dan toggle mute volume di **Settings > Voice & Dictation** serta **Settings > Shortcuts**.

---

## 2. Architecture & Components

### 2.1 Backend (Go)
- **`pkg/system/audio.go` & `pkg/system/audio_darwin.go` / `pkg/system/audio_windows.go`**:
  - `MuteSystemVolume() (originalVolume int, wasMuted bool, err error)`:
    - Di macOS: membaca volume via AppleScript `get volume settings` lalu memanggil `set volume output muted true` atau `set volume output volume 0`.
    - Di Windows: menggunakan CoreAudio IAudioEndpointVolume API atau PowerShell audio endpoint utility untuk mute dan menyimpan volume sebelum mute.
  - `RestoreSystemVolume(originalVolume int, wasMuted bool) error`:
    - Mengembalikan kondisi mute dan volume semula.
- **Wails Bindings di `app.go`**:
  - `MuteSystemVolume() (map[string]interface{}, error)`
  - `RestoreSystemVolume(prevVolume int, wasMuted bool) error`

### 2.2 Frontend State & Shortcuts
- **Settings Store (`frontend/src/lib/stores/settings.svelte.ts`)**:
  - `dictationShortcut: string` (Default: `'Fn'` atau `'DoubleFn'`).
  - `dictationMuteSystemAudio: boolean` (Default: `true`).
  - `dictationHoldThresholdMs: number` (Default: `300ms` untuk membedakan hold vs tap).
- **Global Keydown & Keyup Handler (`frontend/src/App.svelte` & Composer)**:
  - Pelacakan interval tap ganda (`< 350ms`) untuk mengaktifkan mode locked/hands-free.
  - Pelacakan `keyup` setelah durasi hold (`> 300ms`) untuk push-to-talk release stop.
- **Composer & Dictation Controller (`frontend/src/lib/components/chat/Composer.svelte`)**:
  - Menangani siklus mulai/selesai perekaman, memanggil backend mute/restore audio, dan menyisipkan hasil transkripsi ke textarea.
- **UI Settings (`frontend/src/lib/components/layout/SettingsModal.svelte`)**:
  - Menampilkan shortcut dictation di daftar shortcut dan tab Voice & Dictation.
  - Mendukung perekaman custom shortcut melalui `KeyRecorderModal`.

---

## 3. Data Flow & Interaction Lifecycle

```
[User Action: Double-Tap Fn or Hold Fn]
       │
       ├───> Frontend App.svelte detect shortcut event
       │
       ├───> Mute System Volume via Wails Go Bridge (stores original volume)
       │
       ├───> VoiceRecorderManager startRecording (Microphone stream)
       │
       ├───> Composer UI renders Pulsing Active Recording Pill
       │
[User Action: Single Tap Fn (if locked) OR Release Fn (if held)]
       │
       ├───> VoiceRecorderManager stopRecording (extracts audio blob)
       │
       ├───> Restore System Volume via Wails Go Bridge (restores original volume)
       │
       ├───> Grok Transcription API call (TranscribeAudioWithGrok)
       │
       └───> Text inserted at cursor in Composer textarea with focus
```

---

## 4. Error Handling & Edge Cases
1. **Interrupted / Cancelled Recording**: Jika user menekan tombol `Esc` atau terjadi error pada mikrofon/API, volume sistem tetap dipastikan di-restore (*guaranteed finally block*).
2. **Double-Trigger Guard**: Mencegah multiple start/stop race condition ketika tombol ditekan berulang kali secara tidak beraturan.
3. **App Minimization / Focus Loss**: Jika window kehilangan fokus saat push-to-talk sedang aktif, otomatis selesaikan rekaman dan kembalikan volume sistem.

---

## 5. Verification Plan
1. **Shortcut Verification**: Uji mode hold push-to-talk dan mode double-tap lock-to-talk di Composer.
2. **Audio Mute/Restore Verification**: Nyalakan musik/video, mulai dictation, pastikan speaker langsung hening (0%), lalu selesaikan dictation dan pastikan suara kembali ke level semula tanpa ada glitch.
3. **Settings Persistence**: Ubah konfigurasi shortcut dan toggle mute di Settings, muat ulang aplikasi, pastikan pengaturan tersimpan di backend JSONL.
