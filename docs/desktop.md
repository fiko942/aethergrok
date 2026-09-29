# AetherGrok Desktop Studio

AetherGrok Desktop Studio is a native, high-performance desktop application for the Grok Build CLI (`grok`) and autonomous engineering agents. Engineered with **Go 1.24**, **Wails v2**, **Svelte 5 Runes**, and **Ant Design Dark Tokens**.

---

## 📦 Direct Downloads & Packages

Installers and packages are published on [GitHub Releases](https://github.com/fiko942/aethergrok/releases/latest).

| Platform | Architecture | Distribution Formats |
| :--- | :--- | :--- |
| **Windows** | x64 (`amd64`) & ARM64 | `AetherGrok-Setup.exe` (NSIS Installer)<br>`aethergrok-windows-amd64.zip` (Portable ZIP) |
| **macOS** | Apple Silicon (`arm64`) & Intel (`amd64`) | `AetherGrok-macOS-arm64.dmg` (Styled DMG)<br>`AetherGrok-macOS-amd64.dmg` (Styled DMG) |
| **Linux** | x86_64 & aarch64 | Standalone Executable, AppImage |

---

## 🛠️ Build Commands

### Windows Automated Build
```powershell
# Build Windows executable, portable zip, and NSIS installer
powershell -ExecutionPolicy Bypass -File .\build-windows.ps1

# Options:
# .\build-windows.ps1 -Arch arm64         # Build for ARM64
# .\build-windows.ps1 -SkipTests          # Skip test execution
# .\build-windows.ps1 -SkipInstaller      # Skip NSIS installer creation
```

### macOS Automated Build
```bash
# Build macOS application bundle & styled DMG
./build-macos.sh arm64   # Apple Silicon
./build-macos.sh amd64   # Intel
```

### Development Mode (Live HMR)
```bash
# Install frontend packages
cd frontend && npm install && cd ..

# Launch Wails live dev server
wails dev
```

---

## 🛡️ Operating System Security & Warnings

### Windows (Microsoft Defender SmartScreen)
On first run of the installer or executable, SmartScreen may display:
> *"Windows protected your PC — Microsoft Defender SmartScreen prevented an unrecognized app from starting."*

**Resolution:**
Click **More info** (*Informasi selengkapnya*) → **Run anyway** (*Tetap jalankan*).

### macOS (Gatekeeper Ad-Hoc Notice)
If macOS displays an unverified developer warning:
- **System Settings:** Open **System Settings** → **Privacy & Security** → Scroll to **Security** → Click **Open Anyway** (*Tetap Buka*).
- **Terminal:** Run `xattr -d com.apple.quarantine /Applications/AetherGrok.app`.

---

## ⌨️ Global Shortcuts Parity

- **Global OS Screen Snapshot**: `Ctrl+Shift+S` / `Ctrl+Alt+S` (Windows) | `⌘⇧S` (macOS)
- **Voice Dictation (Push-to-Talk)**: `\` *(Backslash)*
- **Toggle Left Sidebar**: `Ctrl+B` (Windows) | `⌘B` (macOS)
- **Toggle Right Sidebar (Git & Explorer)**: `Ctrl+Alt+B` (Windows) | `⌘⌥B` (macOS)
- **Open Settings**: `Ctrl+,` (Windows) | `⌘,` (macOS)
- **Skills & MCP Catalog**: `Ctrl+K` (Windows) | `⌘K` (macOS)
- **New Session Tab**: `Ctrl+T` (Windows) | `⌘T` (macOS)
- **Close Active Tab**: `Ctrl+W` (Windows) | `⌘W` (macOS)
- **Switch Session Tab**: `Ctrl+1..8` / `Ctrl+9` (Windows) | `⌘1..8` / `⌘9` (macOS)
