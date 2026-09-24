# Superpowers Specification: AetherGrok Settings & Smart Screen Snapshot

## 1. Overview & Architecture
AetherGrok memisahkan konfigurasi aplikasi dari file JSON acak dan memusatkannya dalam arsitektur **Superpowers** terstandarisasi (`docs/superpowers/specs/` & `docs/superpowers/plans/`) serta state storage reaktif (`SettingsStore` di Svelte 5 Runes).

---

## 2. Pengaturan Lengkap yang Dikelola

| Kategori | Parameter | Default Value | Deskripsi & UX |
| :--- | :--- | :--- | :--- |
| **Smart Snapshot** | `snapshotShortcut` | `CmdOrCtrl+Shift+S` | Pintasan keyboard global yang dapat dikustomisasi oleh pengguna. |
| **Smart Snapshot** | `snapshotSoundEnabled` | `true` | Sintesis audio mekanis rana kamera ("Cekrek") via Web Audio API. |
| **Smart Snapshot** | `snapshotFlashEnabled` | `true` | Animasi visual kilatan putih layar (*white screen flash* 260ms). |
| **Smart Snapshot** | `snapshotAutoAttach` | `true` | Otomatis melampirkan gambar snapshot ke input composer visi. |
| **Smart Snapshot** | `snapshotDelayMs` | `50` ms | Delay compositor window hiding untuk mencegah *ghost frame* jendela sendiri. |
| **Engine Controls** | `autoHideWindow` | `true` | Menyembunyikan jendela AetherGrok sesaat saat snapshot diambil. |
| **Engine Controls** | `diffPreviewer` | `Active (Green)` | Visualizer inline perbandingan diff sebelum eksekusi patch. |
| **Models & Reasoning**| `defaultModel` | `9router` | Model AI inference aktif dari backend Grok CLI (`9router`, `9router-explore`, `9router-plan`). |
| **Models & Reasoning**| `defaultReasoningEffort`| `medium` | Alokasi token berpikir model (`none`, `low`, `medium`, `high`, `max`). |
| **Permissions** | `permissionMode` | `default` | Batasan keamanan eksekusi (`default`, `acceptEdits`, `auto`, `plan`, `bypassPermissions`). |
| **Theme & Style** | `theme` | `dark-studio` | Skema palet Ant Design (`dark-studio` `#0F1117`, `dark-high-contrast` `#000000`, `light-antd` `#F5F5F5`). |
| **DOM Memory Guard** | `activeWindowTurnCount` | `10` turns | Batas DOM turn aktif untuk menjaga penggunaan RAM di bawah 60MB. |
| **Grok Binary** | `grokBinaryPath` | `/Users/fiko942/.local/bin/grok` | Lokasi binary eksekusi Grok CLI engine. |

---

## 3. Keyboard Shortcuts Matrix

| Kombinasi Tombol | Aksi | Scope |
| :--- | :--- | :--- |
| `Cmd/Ctrl + Shift + S` *(Customizable)* | **Capture Smart Screen Snapshot** | Global Desktop Studio |
| `Cmd/Ctrl + K` | Buka / Tutup **Skills & MCP Catalog** | Global |
| `Cmd/Ctrl + ,` | Buka **Settings Modal** | Global |
| `Cmd/Ctrl + T` | Buat **Sesi / Task Baru** | Workspace Aktif |
| `Shift + Enter` | Baris baru pada Composer Input | Composer Editor |
| `Enter` | Kirim Prompt Turn ke Grok CLI | Composer Editor |
| `Escape` | Tutup Modal Dialog Aktif | Global Overlays |

---

## 4. Mekanisme Snapshot Native & Non-Intrusif
1. **Window Hiding**: Jendela Wails menyembunyikan dirinya sendiri (`WindowHide`).
2. **Compositor Delay**: Menunggu `snapshotDelayMs` (50ms di macOS) agar window server menyelesaikan rendering frame desktop.
3. **Capture**: Memanggil native Go screen capture API (`pkg/screen`).
4. **Window Restoring**: Jendela dikembalikan ke posisi semula (`WindowShow`).
5. **Feedback & Context Injection**: Memutar suara "cekrek", menyalakan kilatan layar putih, dan melampirkan frame Base64 ke input composer.
