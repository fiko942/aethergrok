# 2026-09-25: AetherGrok Advanced Smart Snapshot Engine (macOS & Windows, Shutter Audio, Screen Flash FX, Settings)

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Mengembangkan fitur **Smart Non-Intrusive Screen Snapshot** tingkat lanjut untuk aplikasi AetherGrok Desktop GUI dengan dukungan penuh cross-platform (macOS & Windows), efek visual flash animasi layar putih, efek audio kamera shutter mekanik ("cekrek") sintetis tanpa dependensi file eksternal (via Web Audio API), serta konfigurasi lengkap di menu Settings.

**Architecture:**
1. **Visual & Auditory Feedback Architecture:**
   - **Audio Shutter Synthesizer (`frontend/src/lib/utils/audio.ts`):** Menggunakan Web Audio API native (`AudioContext`) untuk menghasilkan suara mekanik 2 tahap (tahap 1: pembukaan cermin kamera 1100Hz $\to$ 400Hz; tahap 2: pelepasan shutter curtain "cekrek" 320Hz $\to$ 90Hz). Bebas dari dependensi file `.mp3`/`.wav`.
   - **Visual Screen Flash (`frontend/src/lib/components/snapshot/ScreenFlash.svelte`):** Overlay layar penuh putih transparan dengan animasi fade out lembut (`opacity: 0.85` $\to$ `0` dalam 220ms ease-out) yang dipicu setelah proses screenshot selesai, sehingga gambar screenshot tidak terkontaminasi oleh kilatan putih.
2. **Native Capture Pipeline (Go Backend `pkg/screen`):**
   - **macOS:** Menggunakan `screencapture -x` (silent capture langsung ke memory stream PNG via CoreGraphics).
   - **Windows:** Menggunakan Win32 GDI `BitBlt` / PowerShell `CopyFromScreen` dengan flag proses tersembunyi.
   - **Window Auto-Hide Coordination:** `Window.Hide()` $\to$ Jeda sinkronisasi kompositor OS (50ms macOS Quartz / 80ms Windows DWM) $\to$ Native Grab $\to$ `Window.Show()` & `Window.Focus()` (dijamin via `defer`).
3. **Settings & Customization Integration:**
   - `snapshotDelayMs`: Rentang slider 10ms - 500ms (default 50ms).
   - `snapshotSoundEnabled`: Toggle suara shutter kamera aktif/nonaktif (default `true`).
   - `snapshotFlashEnabled`: Toggle efek visual flash layar aktif/nonaktif (default `true`).
   - `snapshotAutoAttach`: Toggle otomatis melampirkan screenshot ke prompt composer (default `true`).
   - `snapshotShortcut`: Keyboard shortcut `Cmd/Ctrl + Shift + S`.

---

## Bite-Sized Implementation Tasks

### Task 1: Web Audio API Shutter Synthesizer & Visual Flash Component
- [ ] **Step 1:** Buat modul Web Audio API synthesizer di `frontend/src/lib/utils/audio.ts` (`playCameraShutterSound()`).
- [ ] **Step 2:** Buat komponen animasi flash di `frontend/src/lib/components/snapshot/ScreenFlash.svelte`.
- [ ] **Step 3:** Verifikasi audio synthesizer tidak menimbulkan crash saat hardware audio tidak tersedia (safeguard try/catch & state suspended).

### Task 2: Snapshot Settings Store & Modal UI Enhancements
- [ ] **Step 1:** Tambahkan properti `snapshotSoundEnabled`, `snapshotFlashEnabled`, dan `snapshotAutoAttach` ke `frontend/src/lib/stores/settings.svelte.ts`.
- [ ] **Step 2:** Tambahkan kartu konfigurasi Snapshot interaktif ke Tab General di `frontend/src/lib/components/layout/SettingsModal.svelte`.
- [ ] **Step 3:** Hubungkan toggle setting langsung ke trigger capture di `Composer.svelte` dan `App.svelte`.

### Task 3: Global Hotkey Listener & Capture Coordination
- [ ] **Step 1:** Daftarkan listener global hotkey `Cmd+Shift+S` (macOS) dan `Ctrl+Shift+S` (Windows) di `frontend/src/App.svelte`.
- [ ] **Step 2:** Saat hotkey/tombol ditekan, panggil Go bridge `CaptureScreenExcludingSelf(delayMs)`.
- [ ] **Step 3:** Begitu hasil capture diterima: mainkan audio shutter (jika aktif), picu animasi flash (jika aktif), dan masukkan gambar Base64 ke list vision attachments.

### Task 4: Unit Testing & Verification
- [ ] **Step 1:** Jalankan Go unit tests `go test ./test/... -v` untuk memvalidasi alur koordinasi capture macOS & Windows.
- [ ] **Step 2:** Jalankan Svelte typecheck `pnpm run check` dan build `pnpm run build`.
- [ ] **Step 3:** Commit dan push ke repository GitHub `fiko942/grok-build`.
