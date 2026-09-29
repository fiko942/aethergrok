# Installation & Setup Guide

AetherGrok Desktop Studio is an ultra-fast, native desktop application for orchestrating the official Grok CLI (`grok`) and autonomous coding agents.

---

## 📦 Direct Downloads

Official releases are available on [GitHub Releases](https://github.com/fiko942/aethergrok/releases/latest).

### 🪟 Windows (Windows 10 / 11 64-bit & ARM64)

1. **Windows Setup Installer (`AetherGrok-Setup.exe`)**:
   - Recommended for standard desktop setups.
   - Includes full NSIS wizard, Start Menu shortcuts, Desktop shortcut, and clean Windows uninstaller registration.
2. **Portable ZIP Package (`aethergrok-windows-amd64.zip`)**:
   - Zero-installation option.
   - Extract the ZIP to any local folder and launch `aethergrok.exe`.

> **SmartScreen Notice**: If Windows Defender SmartScreen warns about an unrecognized publisher, click **More info** (*Informasi selengkapnya*) → **Run anyway** (*Tetap jalankan*).

---

### 🍏 macOS (macOS 11+ Big Sur, Monterey, Ventura, Sonoma, Sequoia)

1. Download `AetherGrok-<version>-macOS-arm64.dmg` (for Apple Silicon M1/M2/M3/M4) or `…-amd64.dmg` (for Intel x86_64).
2. Open the `.dmg` and drag `AetherGrok.app` into your `/Applications` folder.

#### Terminal 1-Line Installer:
```bash
curl -fsSL https://raw.githubusercontent.com/fiko942/aethergrok/main/scripts/install-app.sh | bash
```

#### Gatekeeper Override:
If macOS displays an unverified developer warning:
- Go to **System Settings** → **Privacy & Security** → **Security** → click **Open Anyway**.
- Or run in Terminal: `xattr -d com.apple.quarantine /Applications/AetherGrok.app`.

---

## ⚡ Grok CLI Installation & Authentication

AetherGrok requires the official Grok CLI (`grok`) to interface with xAI models.

### Option 1: 1-Click Integrated In-App Installation
Launch AetherGrok. If the CLI is not found, the onboarding screen provides a **1-Click Install** button that automatically downloads and configures the CLI for your OS.

### Option 2: Manual Terminal Installation

#### Windows (PowerShell):
```powershell
irm https://x.ai/cli/install.ps1 | iex
```

#### macOS / Linux / WSL (Bash / Zsh):
```bash
curl -fsSL https://x.ai/cli/install.sh | bash
```
*(Or via Homebrew on macOS: `brew install grok`)*

### Authenticating:
```bash
grok login
```
This opens a browser for standard OAuth login. Alternatively, set `XAI_API_KEY` in your environment or enter your API key in **Settings (`Ctrl+,` / `⌘,`)**.

---

## 🛠️ Building from Source

### Prerequisites
- **Go 1.24+**: [https://go.dev/dl/](https://go.dev/dl/)
- **Node.js 20+** with `npm` or `pnpm`: [https://nodejs.org](https://nodejs.org)
- **Wails CLI v2**: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- **C/C++ Compiler**: MinGW-w64 / MSVC (Windows) or Xcode CLI Tools (macOS)
- **NSIS 3.x** *(Optional)*: `choco install nsis` (Windows setup installer creation)

### Windows Automated Build:
```powershell
powershell -ExecutionPolicy Bypass -File .\build-windows.ps1
```

### macOS Automated Build:
```bash
./build-macos.sh arm64    # Apple Silicon
./build-macos.sh amd64    # Intel Mac
```

### Development Mode (Vite HMR + Wails Backend):
```bash
cd frontend && npm install && cd ..
wails dev
```
