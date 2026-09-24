# 2026-09-25: AetherGrok (Grok Desktop GUI) - Deep Research, Dependencies & Risk Mitigation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Menentukan daftar dependensi/library optimal untuk desktop GUI Grok Build (AetherGrok), mengidentifikasi seluruh potensi kendala teknis (memori, storage, konkurensi, IPC bottleneck, multi-platform), dan merancang solusi mitigasi teruji agar aplikasi tetap ringan, stabil, dan responsif pada perangkat laptop berspesifikasi rendah.

---

## 1. Inventory Library & Dependensi (Tujuan & Alasan Pemilihan)

### A. Frontend Presentation Layer (Svelte 5 + Ant Design Ecosystem)
| Package / Library | Kategori | Tujuan & Alasan Pemilihan |
| :--- | :--- | :--- |
| `svelte` (v5) | UI Framework | Compile-time reactivity via Runes (`$state`, `$derived`, `$effect`), zero virtual DOM overhead, bundle size sangat kecil (<30KB), konsumsi memori minim. |
| `@ant-design/colors` | Color Token Engine | Menyediakan palet token warna Ant Design resmi (10-step palette) untuk Dark Mode, High Contrast, dan sistem status visual. |
| `@ant-design/icons-svelte` / `lucide-svelte` | Iconography | Koleksi icon SVG teroptimasi untuk tombol snapshot, multi-session tabs, file chips, tool status, dan navigasi setting. |
| `tailwindcss` + `@tailwindcss/typography` | Utility Styling | Utilitas CSS atomic untuk layouting cepat, performa render tinggi, dan integrasi token Ant Design tanpa runtime CSS overhead. |
| `diff2html` + `diff` | Visual Diff Inspector | Render visual line-by-line / side-by-side diff perubahan kode dari tool Grok (`write`, `search_replace`) dengan highlighting warna +hijau / -merah. |
| `highlight.js` (core bundle) | Code Highlighting | Syntax highlighting untuk output blok kode Markdown, dibatasi hanya untuk bahasa umum (JS, TS, Go, Python, Rust, HTML, CSS, JSON, YAML) guna menghemat memori. |
| `katex` (KaTeX) | LaTeX Math Rendering | Render persamaan matematika inline/display secara offline, jauh lebih cepat dan ringan dibanding MathJax penuh. |
| `mermaid` (Dynamic Import) | Diagram Engine | Render diagram Mermaid secara asinkron saat blok ` ```mermaid ` terdeteksi, menghindari beban memori saat sesi tidak memuat diagram. |
| `@floating-ui/dom` | Overlay Positioning | Positioning tooltips, context menus, model selector dropdown, dan popovers yang presisi dan bebas dependency berat. |

### B. Core Desktop Engine & Native Bridge (Go / Wails v2)
| Package / Module | Kategori | Tujuan & Alasan Pemilihan |
| :--- | :--- | :--- |
| `github.com/wailsapp/wails/v2` | Desktop Bridge | Integrasi Go backend dengan WebView native OS (WebKit di macOS, WebView2 di Windows). Menghasilkan single binary mandiri (~15-25MB) dengan idle RAM <40MB. |
| `github.com/kbinani/screenshot` | Screen Capturer | Native cross-platform display capture engine (menggunakan CoreGraphics di macOS dan GDI/BitBlt di Windows). |
| `github.com/mattn/go-sqlite3` / `modernc.org/sqlite` | Local Storage | Cache indexing metadata sesi, riwayat pencarian cepat, dan metadata skill tanpa memindai ratusan file JSONL di disk setiap frame. |
| `github.com/fsnotify/fsnotify` | Filesystem Watcher | Memantau perubahan file konfigurasi (`~/.grok/config.toml`), skill baru di `~/.grok/skills/`, dan pembaruan log sesi di `~/.grok/sessions/` secara real-time berbasis OS event. |
| `golang.org/x/sys` | OS Level Interop | Akses Win32 window APIs (Windows) dan CoreGraphics/AppKit hooks (macOS) untuk auto-hide dan restore window aplikasi saat snapshot. |

---

## 2. Analisis Potensi Masalah & Bottleneck Teknis (Brainstorming & Risk Depiction)

### 1. Masalah Keterbatasan Memori & DOM Explosion (Memory Bloat & Chat Lag)
* **Penyebab:**
  * Sesi Grok berdurasi panjang dapat menghasilkan ribuan baris teks, puluhan tool calls, inline diffs besar, dan base64 images.
  * Jika seluruh riwayat pesan dirender langsung ke dalam DOM, browser WebView akan mengalami memory leak, garbage collection lag, dan UI freeze saat scroll.
* **Gejala:** Memory melonjak >500MB, scroll tersendat (<15 FPS), lag parah saat berpindah antar tab sesi.

### 2. Masalah Penyimpanan Disk & File I/O Bottleneck (Storage & Stat Flooding)
* **Penyebab:**
  * Direktori `~/.grok/sessions/` menyimpan ratusan sesi dalam format file `.jsonl` dan snapshot repository.
  * Pemindaian sinkron (`os.Stat` / `ioutil.ReadDir`) pada ratusan folder sesi di setiap render siklus sidebar menyebabkan pembacaan disk tinggi (disk IO thrashing).
  * Temporary screenshot image yang tidak dibersihkan menumpuk dan menghabiskan ruang disk lokal.

### 3. Masalah Bottleneck Streaming NDJSON & IPC Throughput
* **Penyebab:**
  * Grok CLI dengan flag `--output-format streaming-json` atau `streaming-messages-json` dapat memancarkan ratusan chunk per detik saat mode reasoning/thinking aktif.
  * Mengirim setiap token kecil satu per satu melalui jembatan IPC (Go ↔ Svelte) menyebabkan IPC congestion dan freeze pada UI main-thread.

### 4. Masalah Snapshot Visual Glitch & Screen Race Condition
* **Penyebab:**
  * Menyembunyikan jendela aplikasi (`Window.Hide()`) dan memicu capture secara instan dapat menangkap bayangan transparan jendela aplikasi (ghost window frame) jika window compositor OS (macOS Quartz / Windows DWM) belum selesai me-render ulang layar.
  * Di multi-monitor setup, penentuan layar aktif (active display vs cursor display) dapat menghasilkan koordinat snapshot yang salah.

### 5. Masalah Zombie Subprocesses & Concurrent Session Locking
* **Penyebab:**
  * Menjalankan banyak subagent atau beberapa sesi paralel dapat meninggalkan proses `grok` yang tetap berjalan di background saat jendela aplikasi ditutup paksa.
  * Dua sesi yang mengakses git worktree yang sama dapat menyebabkan file lock conflict.

---

## 3. Rencana Solusi & Strategi Mitigasi Teruji

```
+-------------------------------------------------------------------------------+
|                      MITIGATION ARCHITECTURE MATRIX                           |
+-------------------------------------------------------------------------------+
| 1. Virtualized Message Windowing : Hanya render 10-15 turn aktif dalam DOM.   |
| 2. Sliding Window Buffer        : Chunk stream buffer 16ms (60FPS batching).  |
| 3. SQLite Local Metadata Cache   : Eliminasi synchronous filesystem disk scan.|
| 4. Smart Snapshot Synchronization: Compositor flush + multi-display detection.|
| 5. Process Lifecycle Supervisor  : Auto-kill orphan subprocesses on app exit. |
+-------------------------------------------------------------------------------+
```

### Solusi 1: Virtualized Chat Windowing & Progressive History Loading
* **Mekanisme:**
  * Terapkan arsitektur *windowed turn rendering*: hanya 10 turn percakapan terakhir yang dirender di DOM aktif.
  * Turn terdahulu disimpan di memori Svelte state / SQLite cache dan hanya dimasukkan ke DOM saat pengguna melakukan scroll ke atas (*infinite scroll with virtual spacer*).
  * Card diff besar yang sudah collapsed diganti dengan lightweight placeholder.

### Solusi 2: Stream Token Batching (Sliding Window Buffer 16ms)
* **Mekanisme:**
  * Di Go backend, streaming chunk dari stdout `grok` dikumpulkan dalam buffer lokal dengan interval 16ms (disesuaikan dengan refresh rate 60 FPS).
  * Go memancarkan event `grok:delta_batch` berisi gabungan token daripada memicu ratusan event IPC individual, menjaga UI thread tetap ringan.

### Solusi 3: SQLite Session Cache & Auto Cleanup Temporary Media
* **Mekanisme:**
  * Gunakan SQLite database lokal (`~/.grok-desktop/cache.db`) untuk mengindeks ID sesi, judul, waktu modifikasi, dan token usage.
  * Pemindaian disk hanya berjalan saat ada notifikasi filesystem dari `fsnotify`.
  * Folder snapshot temporary (`/tmp/aethergrok_snaps/`) memiliki mekanisme auto-prune: gambar screenshot berumur >24 jam atau yang sudah ter-upload otomatis dihapus saat sesi ditutup.

### Solusi 4: Smart Native Snapshot Engine dengan Compositor Synchronization
* **Mekanisme:**
  * `Window.Hide()` dipanggil dengan async channel.
  * Delay minimal 50ms untuk macOS dan 80ms untuk Windows guna memastikan desktop window manager selesai me-repaint area yang sebelumnya tertutup.
  * Identifikasi posisi kursor mouse untuk memilih display index yang benar pada konfigurasi multi-monitor.
  * Setelah capture selesai, gambar dikompres ke WebP/PNG dengan resolusi optimal sebelum diserahkan ke Svelte UI.

### Solusi 5: Process Lifecycle Supervisor & Graceful Shutdown Hooks
* **Mekanisme:**
  * Setiap subprocess `grok` didaftarkan ke dalam Process Group (PGID) mandiri di backend Go.
  * Tambahkan listener `runtime.EventsOn(ctx, "app:shutdown")` dan Go signal handling (`SIGINT`, `SIGTERM`) untuk memastikan seluruh subagent dan subprocess ditutup secara bersih (`KillProcessGroup`) saat aplikasi dimatikan.
