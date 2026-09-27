# Implementation & Architecture Plan: Global Hotkey Listener, Native Git Resolution, Multi-Workspace Tabs, and Robust File Drop Ingestion

> **Goal:** Dokumentasikan arsitektur sistem, alur implementasi, dan verifikasi menyeluruh untuk fitur Global Hotkey Listener (`pkg/hotkey`), Native Git Binary Resolution (`pkg/gitutil`), Session Tabs Multi-Workspace Context Badges, serta Enhanced Drag & Drop Ingestion pada Prompt Composer.

---

## 1. Komponen yang Ditingkatkan

### 1.1 Native Global Hotkey Engine (`pkg/hotkey`)
- **Tujuan**: Memungkinkan shortcut snapshot (seperti tombol `RightShift` atau `CmdOrCtrl+Shift+S`) dipicu dari mana saja di seluruh OS, bahkan saat window AetherGrok sedang tidak aktif / diminimalkan.
- **Implementasi macOS (`hotkey_darwin.go`)**:
  - `CGEventTapCreate` pada `kCGSessionEventTap` dengan mode `kCGEventTapOptionListenOnly`.
  - Handler `kCGEventFlagsChanged` dan `kCGEventKeyDown` untuk menangani standalone modifier keys (`RightShift`, `LeftShift`, `RightCmd`, `RightOption`) dan shortcut kombinasi.
  - Debounce 250ms untuk stabilitas hardware keypress.
  - Event Wails `snapshot:trigger_global` dikirimkan ke UI secara instan.

### 1.2 Native Git Binary Resolver (`pkg/gitutil`)
- **Tujuan**: Mencegah kegagalan eksekusi Git di macOS akibat wrapper `xcrun` bawaan `/usr/bin/git` saat arsitektur berbeda (`unable to load libxcrun ... fat file`).
- **Implementasi**:
  - Mengutamakan binary langsung dari Apple Silicon Homebrew (`/opt/homebrew/bin/git`), Intel Homebrew (`/usr/local/bin/git`), CommandLineTools (`/Library/Developer/CommandLineTools/usr/bin/git`), dan Xcode.
  - Digunakan di seluruh modul Go: `pkg/workspace`, `pkg/skills`, dan `app.go`.

### 1.3 Multi-Workspace Session Tabs & Tab Reordering
- **Tujuan**: Menampilkan tab sesi dengan label workspace yang jelas saat membuka beberapa workspace sekaligus, serta mendukung drag & drop reorder tab secara presisi.
- **Implementasi**:
  - Menggunakan `openTabs` di `SessionTabs.svelte` dengan badge workspace ringkas (`ws.name`).
  - Pembaruan fungsi `reorderSessions` pada `sessionStore` untuk memindahkan index `openTabSessionIds` dan menyimpannya langsung ke storage.

### 1.4 Centralized File / Vision Drag-and-Drop Ingestion
- **Tujuan**: Memastikan drag & drop file, tangkapan layar macOS proxy icons, dan paste gambar dari web/clipboard diproses tanpa kehilangan tipe data.
- **Implementasi**:
  - Parsing mendalam untuk `dataTransfer.types` (`Files`, `public.png`, `public.tiff`, `image/*`).
  - Ekstraksi `dataTransfer.items` untuk blob file gambar dengan konversi otomatis ke `File`.

---

## 2. Status Verifikasi

| Komponen / Suite | Hasil Pengujian | Status |
|------------------|-----------------|--------|
| `go test ./test/... ./pkg/... -v` | Semua 9 test suite lulus tanpa error | PASS |
| `pnpm test test/diffUtils.test.ts` | 5/5 unit tests lulus | PASS |
| `pnpm --prefix frontend run check` | 0 errors | PASS |
| `pnpm --prefix frontend run build` | Sukses mem-build bundle frontend | PASS |
