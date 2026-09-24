# AetherGrok Master Context & Native Hot-Reload Architecture

> **Superpowers Context & Architecture Knowledge Base**
> Date: 2026-09-25
> Repository: `fiko942/grok-build`
> Target Platforms: macOS (Apple Silicon / Intel) & Windows (x64)

---

## 1. Executive Summary & Core Constraints

Project **AetherGrok** adalah aplikasi **Desktop GUI Native Studio** mandiri untuk mengorkestrasi **Grok Agentic AI CLI (`grok`)**.

### Batasan Kritis Arsitektur:
1. **Bukan Electron:** Electron memaketkan Chromium engine lengkap (~150MB installer) dan runtime Node.js yang boros memori (150MB-300MB idle RAM, konsumsi GPU tinggi). AetherGrok secara tegas menghindari Electron.
2. **Bukan Web Browser Eksternal:** Aplikasi **tidak membuka browser web** eksternal (seperti Chrome, Safari, atau Edge). Aplikasi berjalan sebagai **jendela window native desktop** mandiri.
3. **Pemanfaatan Native OS WebView (Wails v2 + Go 1.24):**
   - macOS: WebKit bawaan sistem operasi via Cocoa/AppKit.
   - Windows: Microsoft Edge WebView2 Evergreen bawaan Windows 10/11.
   - Hasil: Ukuran binary sangat kecil (~8.5MB) dan konsumsi idle RAM hanya **~35-45MB**.
4. **Svelte 5 Runes Reactivity:** Kompilasi HTML/CSS/JS tanpa Virtual DOM overhead, ukuran bundle <30KB, responsif dan ringan pada laptop spesifikasi rendah.
5. **Hot-Reloading Realtime:** Saat dalam mode development (`./dev.sh`), setiap perubahan kode pada frontend Svelte (HMR via Vite) maupun kode Go backend otomatis melakukan compile ulang dan me-reload jendela desktop tanpa perlu restart manual.

---

## 2. Fitur-Fitur Utama & Mekanisme Teknis

### A. Hot-Reload Development Workflow
* Dikonfigurasi dalam `wails.json`:
  * `"frontend:dev:watcher": "pnpm dev"`
  * `"frontend:dev:serverUrl": "auto"`
* **Cara kerja:**
  * Saat menjalankan `./dev.sh` (atau `wails dev`), Wails menjalankan Vite dev server di background dengan WebSocket HMR (Hot Module Replacement).
  * Webview native di dalam jendela desktop terhubung langsung ke stream HMR Vite. Setiap perubahan komponen Svelte langsung ter-update di layar desktop dalam hitungan milidetik.
  * Perubahan pada file `.go` otomatis memicu kompilasi ulang Go backend secara transparan.

### B. Smart Non-Intrusive Screen Snapshot (Window Auto-Exclusion)
* **Tujuan:** Mengambil tangkapan layar desktop aktif tanpa mengikutsertakan jendela AetherGrok itu sendiri.
* **Alur Eksekusi (4 Tahap):**
  1. Frontend memicu metode Go bridge `CaptureScreenExcludingSelf(delayMs)`.
  2. Jendela AetherGrok disembunyikan seketika via `Window.Hide()`.
  3. Sinkronisasi compositor OS (50ms untuk macOS Quartz, 80ms untuk Windows DWM) agar area layar di belakang jendela selesai di-repaint secara bersih tanpa ghost frame.
  4. Native capture engine mengambil bitmap display aktif (CoreGraphics pada macOS, GDI/BitBlt pada Windows).
  5. Jendela AetherGrok dipulihkan dan difokuskan kembali (`Window.Show()`, `Window.Focus()`) melalui `defer` anti-panic.
  6. Gambar dikonversi menjadi data Base64 PNG dan otomatis di-attach ke input Composer sebagai vision prompt.

### C. Multi-Session Docking Workspace & 10-Turn DOM Windowing
* **Tab Docking:** Horizontal tab bar interaktif dengan drag-and-drop reorder, rename inline, session forking, dan penutupan tab.
* **Indikator Status Warna Ant Design:**
  * 🔵 **Working:** Subagent / turn aktif memproses token.
  * 🟡 **Waiting Permission:** Menunggu persetujuan eksekusi tool.
  * 🟢 **Finished:** Turn selesai dengan sukses.
  * 🔴 **Error:** Terjadi kegagalan eksekusi.
  * ⚪ **Idle:** Sesi siap menerima instruksi.
* **10-Turn Windowing:** Membatasi jumlah node turn percakapan di DOM maksimal 10 turn aktif untuk mencegah memory leak pada percakapan panjang. Turn sebelumnya dimuat secara bertahap saat scroll ke atas (*progressive hydration*).

### D. Process Lifecycle Supervisor & 16ms Stream Batching
* Subprocess `grok` dijalankan dalam Process Group terisolasi (`Setpgid: true` di Unix, `CREATE_NEW_PROCESS_GROUP` di Windows) sehingga seluruh subagent tertutup bersih saat turn dibatalkan atau aplikasi dimatikan (zero zombie process).
* Output NDJSON stream dari `grok --output-format streaming-json` ditampung dalam buffer sliding-window 16ms (60 FPS batching) untuk mencegah kemacetan jalur IPC.

### E. Skills & MCP Catalog Hub
* Memindai file `SKILL.md` dari `~/.grok/skills/` dan `~/.agents/skills/`.
* Modal pencarian instan dengan tab kategori (Frontend, Backend, Design, Agents, Tools) dan tombol satu klik untuk menyisipkan prompt skill ke composer.

### F. Ant Design Dark Studio Theming & Settings
* Sistem tema menggunakan token resmi Ant Design (`#1677FF`, `#141414`, `#1F1F1F`, `#262626`).
* Opsi model (`grok-4.6`, `grok-code`), level reasoning effort (`none`, `low`, `medium`, `high`, `max`), dan 5 mode permissions (Default, AcceptEdits, Auto, Plan, BypassPermissions).

---

## 3. Struktur Berkas & Komponen Utama

```
grok-build/
├── docs/superpowers/plans/
│   ├── 2026-09-25-grok-build-desktop-gui.md   # Master Plan Implementasi AetherGrok
│   └── 2026-09-25-aethergrok-context-hotreload.md # Pengetahuan Konteks & Hot-Reload ini
├── dev.sh                                    # Universal Launcher (Desktop, Web, Build, Test)
├── main.go                                   # Wails native application options & lifecycle
├── app.go                                    # Go bridge API (IPC binding ke frontend)
├── wails.json                                # Konfigurasi Wails & Vite HMR watcher
├── pkg/
│   ├── grokrunner/                           # Subprocess supervisor & 16ms stream parser
│   ├── screen/                               # Native screen capture & window auto-hide
│   └── skills/                               # Parser SKILL.md & in-memory registry
├── frontend/                                 # Svelte 5 + Tailwind + Ant Design Tokens
│   ├── src/
│   │   ├── App.svelte                        # Workspace layout & global events
│   │   ├── lib/
│   │   │   ├── antd/                         # Komponen UI primitif Ant Design
│   │   │   ├── components/layout/            # Header, SessionTabs, SettingsModal
│   │   │   ├── components/chat/              # MessageList, Composer, ToolCallCard, DiffCard
│   │   │   ├── components/skills/            # SkillCatalog, SkillCard
│   │   │   └── stores/                       # Svelte 5 runes state (session, settings)
└── test/                                     # Unit test suite Go (100% PASS)
```
