# AetherGrok Desktop Studio

<p align="center">
  <img src="resources/app-icon.png" width="128" height="128" alt="AetherGrok Logo" /><br>
  <b>High-Performance Native Desktop Studio for Grok Build CLI & Autonomous Engineering</b><br>
  <sub>Engineered with Go 1.24, Wails v2, Svelte 5 Runes, and Ant Design Dark Token Architecture</sub>
</p>

<p align="center">
  <a href="https://github.com/fiko942/aethergrok/releases/latest"><img src="https://img.shields.io/github/v/release/fiko942/aethergrok?color=1677ff&label=Latest%20Release" alt="Release" /></a>
  <a href="https://github.com/fiko942/aethergrok/releases"><img src="https://img.shields.io/github/downloads/fiko942/aethergrok/total?color=52c41a&label=Downloads" alt="Downloads" /></a>
  <a href="https://saweria.co/wijifikoteren"><img src="https://img.shields.io/badge/Saweria-Dukung%20Developer-E5A823?logo=coffeescript&logoColor=white" alt="Saweria" /></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License" /></a>
  <a href="https://go.dev"><img src="https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white" alt="Go" /></a>
  <a href="https://svelte.dev"><img src="https://img.shields.io/badge/Svelte-5.x-FF3E00?logo=svelte&logoColor=white" alt="Svelte" /></a>
  <a href="https://wails.io"><img src="https://img.shields.io/badge/Wails-v2-DF1A55?logo=wails&logoColor=white" alt="Wails" /></a>
</p>

<p align="center">
  <a href="#-cross-platform-compatibility">Platforms</a> •
  <a href="#-downloads--installation-guide">Downloads</a> •
  <a href="#-automatic-grok-cli-installation">Grok CLI Setup</a> •
  <a href="#-keyboard-shortcuts">Shortcuts</a> •
  <a href="#-why-aethergrok-the-problem--solution">Why AetherGrok</a> •
  <a href="#-technical-comparison-matrix">Comparison Matrix</a> •
  <a href="#-architecture--engine-flow">Architecture</a> •
  <a href="#-key-features">Features</a> •
  <a href="#-building-from-source">Build from Source</a> •
  <a href="#-sponsorship--donations">Donate</a>
</p>

---

## 🌐 Cross-Platform Compatibility

AetherGrok Desktop Studio is natively compiled and tuned across all major desktop operating systems:

| Platform | Supported Architectures | Minimum OS Version | Distribution Packages |
| :--- | :--- | :--- | :--- |
| **Windows** | **x64 (amd64)** & **ARM64** | Windows 10 / 11 (64-bit) | NSIS Setup Wizard (`AetherGrok-Setup.exe`), Portable ZIP (`aethergrok-windows-amd64.zip`) |
| **macOS** | **Apple Silicon (arm64)** & **Intel (amd64)** | macOS 11.0 Big Sur+ | Styled `.dmg` Volume Installer, Shell Installer Script |
| **Linux** | **x86_64** & **aarch64** | Ubuntu 20.04+, Fedora 36+, Arch | Standalone Binary, AppImage |

---

## 📦 Downloads & Installation Guide

Download the latest signed release directly for Windows and macOS from the [Latest GitHub Release](https://github.com/fiko942/aethergrok/releases/latest).

### 🪟 Windows Installation (Windows 10 / 11)

AetherGrok offers two distribution methods for Windows users:

| Package Type | File Name | Description |
| :--- | :--- | :--- |
| **Windows Setup Installer** | `AetherGrok-Setup.exe` | **Recommended.** NSIS installer with automatic Start Menu shortcut, Desktop shortcut, and clean uninstaller support. |
| **Portable ZIP Archive** | `aethergrok-windows-amd64.zip` | Standalone portable package. Extract anywhere and launch `aethergrok.exe` without administrative installation. |

> **🛡️ Microsoft Defender SmartScreen Notice:**  
> As an independent open-source application, Windows SmartScreen may display:  
> *"Windows protected your PC — Microsoft Defender SmartScreen prevented an unrecognized app from starting."*  
> **To proceed:** Click **More info** (*Informasi selengkapnya*) → Click **Run anyway** (*Tetap jalankan*).

---

### 🍏 macOS Installers (macOS 11+)

| Architecture | Format | Direct Download (Latest Release) | Checksum |
| :--- | :--- | :--- | :--- |
| **Apple Silicon (M1 / M2 / M3 / M4)** | Styled `.dmg` | [Download macOS ARM64 DMG (Latest)](https://github.com/fiko942/aethergrok/releases/latest) | [Verify Checksum](https://github.com/fiko942/aethergrok/releases/latest) |
| **Intel x86_64** | Styled `.dmg` | [Download macOS AMD64 DMG (Latest)](https://github.com/fiko942/aethergrok/releases/latest) | [Verify Checksum](https://github.com/fiko942/aethergrok/releases/latest) |

#### Quick Terminal Install (macOS / Linux)
Installs AetherGrok directly into `/Applications` and automatically clears Gatekeeper quarantine flags:
```bash
curl -fsSL https://raw.githubusercontent.com/fiko942/aethergrok/main/scripts/install-app.sh | bash
```

<details>
<summary><b>macOS Gatekeeper Guide (Ad-hoc Signed Notice)</b></summary>

Because AetherGrok is distributed without a paid Apple Developer ID certificate (ad-hoc signed), macOS Gatekeeper may show a verification prompt. You can open the app using either method:

**Method 1: System Settings (UI)**
1. Open **System Settings** (*Pengaturan Sistem*) on your Mac.
2. Navigate to **Privacy & Security** (*Privasi & Keamanan*) and scroll to **Security**.
3. Under the warning message, click **Open Anyway** (*Tetap Buka*) and enter your password / Touch ID.
4. Click **Open** on the confirmation dialog.

**Method 2: Terminal Command (Instant)**
```bash
xattr -d com.apple.quarantine /Applications/AetherGrok.app
```
*(Or if run from Downloads/DMG: `xattr -cr /Applications/AetherGrok.app`)*
</details>

---

## ⚡ Automatic Grok CLI Installation

AetherGrok communicates with the official Grok CLI (`grok`) to power autonomous agent workflows. Getting the CLI ready takes seconds:

### Option 1: 1-Click Integrated In-App Installation (Recommended)
When launching AetherGrok for the first time without a detected CLI, the onboarding wizard displays a **1-Click Install** button. AetherGrok's Go engine automatically executes the platform-native installation pipeline in the background and sets up the required environment variables.

### Option 2: Quick Terminal / PowerShell Command

#### Windows (PowerShell):
Run PowerShell and execute the official installation script:
```powershell
irm https://x.ai/cli/install.ps1 | iex
```

#### macOS / Linux / WSL:
Open Terminal and run:
```bash
curl -fsSL https://x.ai/cli/install.sh | bash
```
*(On macOS with Homebrew: `brew install grok`)*

#### Authenticate Grok CLI:
After installation, sign in to your account:
```bash
grok login
```
*Alternatively, you can provide an xAI API key via environment variable `XAI_API_KEY` or through AetherGrok's Settings modal (`Ctrl+,` / `⌘,`).*

---

## ⌨️ Keyboard Shortcuts

AetherGrok features comprehensive keyboard shortcut parity across macOS, Windows, and Linux:

| Action / Feature | macOS (`⌘`) | Windows / Linux (`Ctrl` / `Alt`) | Scope / Context |
| :--- | :--- | :--- | :--- |
| **Global OS Screen Snapshot** | `⌘⇧S` (`Cmd+Shift+S`) | `Ctrl+Shift+S` / `Ctrl+Alt+S` | Global (System-Wide Hotkey) |
| **Voice Dictation (Push-to-Talk / Toggle)** | `\` *(Backslash)* | `\` *(Backslash)* | Global / Composer |
| **Toggle Left Sidebar** *(Sessions & Workspaces)* | `⌘B` | `Ctrl+B` | Global Navigation |
| **Toggle Right Sidebar** *(Git Diff & Explorer)* | `⌘⌥B` (`Cmd+Option+B`) | `Ctrl+Alt+B` | Global Navigation |
| **Open Settings & Preferences** | `⌘,` | `Ctrl+,` | Global Modal |
| **Skills & MCP Discovery Catalog** | `⌘K` | `Ctrl+K` | Global Modal |
| **New Conversation Tab** | `⌘T` | `Ctrl+T` | Session Tabs |
| **Close Current Session Tab** | `⌘W` | `Ctrl+W` | Session Tabs |
| **Switch to Tab 1 through 8** | `⌘1` – `⌘8` | `Ctrl+1` – `Ctrl+8` | Session Tabs |
| **Switch to Last Session Tab** | `⌘9` | `Ctrl+9` | Session Tabs |
| **Send Message / Submit Turn** | `⌘Enter` | `Ctrl+Enter` | Composer Input |
| **Split Terminal** | `⌘⇧5` | `Ctrl+Shift+5` | Terminal Panel |
| **Filter & Insert Skill** | `/` | `/` | Composer Prompt |
| **Workspace File Context Autocomplete** | `@` | `@` | Composer Prompt |

> **Custom Keybindings**: You can rebind the Global Snapshot hotkey and Voice Dictation trigger anytime in **Settings (`Ctrl+,` / `⌘,`) → Shortcuts**.

---

## 🎯 Why AetherGrok? The Problem & Solution

### Focused Operational Considerations
Autonomous coding and agent environments require distinct architectural foundations:
1. **Long-Session UI Responsiveness**: Multi-turn agent conversations generate high volumes of rich diffs and terminal logs. Managing DOM memory pressure and scroll positioning is essential for long-running workflows.
2. **Visual & Rich Media Ergonomics**: Terminal workflows excel at raw text input, while graphical interfaces provide complementary inline side-by-side diff viewers, drag-and-drop vision snapshots, and visual skill discovery catalogs.
3. **Subprocess Lifecycle Control**: Managing background subprocess trees cleanly across cancellations, steering prompts, and session resets requires robust OS-level process management.
4. **Environment Setup**: Different developer tools utilize distinct distribution channels (package managers, binary downloads, or containerized environments).

### The AetherGrok Architecture
AetherGrok pairs a lightweight Go backend with Svelte 5 fine-grained reactivity:
- **Low Resource Usage**: Uses native OS webviews via Wails v2 (~35 MB baseline RAM) with Go concurrency.
- **10-Turn Windowing Engine**: Renders active conversational turns in a virtualized DOM window with scroll anchor preservation.
- **Subprocess Group Supervision**: Manages process groups with POSIX `setpgid` and Windows Job Objects alongside a 16ms token stream buffer.
- **In-App CLI Detection**: Checks for local CLI availability on startup with integrated setup assistance.

---

## 📊 Technical Comparison Matrix

An objective overview of architectural approaches and design characteristics across developer environments:

| Metric / Dimension | **AetherGrok Desktop** | **Anthropic Claude Code** | **OpenAI Codex CLI** | **Antigravity / CUA Agents** |
| :--- | :--- | :--- | :--- | :--- |
| **Primary Interaction Mode** | Desktop Studio (Wails v2 + Svelte 5) | Terminal REPL (Node.js CLI) | Terminal CLI / ACP Wrapper | Desktop & Web Automation Agent |
| **Interface Style** | Graphical UI with Visual Diffs & Chat | Terminal Text Interface | Terminal Text Interface | Visual Canvas & OS Interaction |
| **Host Environment** | Standalone Native Binary (Go + Webview) | Global Node.js Package (`npm`) | Python / CLI Executable | Native / Electron / Python Runner |
| **Long Session Handling** | 10-Turn DOM Virtualization Windowing | Terminal Buffer Management | Terminal Buffer Management | Session History Management |
| **Mid-Turn Steerability** | In-flight message injection & queue | Turn pause and re-prompt | ACP method protocol | Action queue adjustment |
| **Process Management** | POSIX `setpgid` & Windows Job Objects | Node.js `child_process` tree | Standard system subprocess | OS Accessibility & Input APIs |
| **Screen Context Input** | Compositor-synced OS snapshot | File reference / attachment | File reference / attachment | Native OS Screen Pixels / CUA |
| **Voice Input Support** | Integrated Push-to-Talk with visualizer | Host terminal audio dependent | Host terminal audio dependent | Host environment dependent |
| **Diff Presentation** | Visual Side-by-Side & Unified Diffs | Terminal ANSI colored diffs | Terminal ANSI colored diffs | Visual inspection / file view |
| **Ecosystem Extensibility** | Visual Skills Hub (`~/.grok/skills`) | CLI Slash Commands & Markdown | Configuration Hooks & APIs | Predefined Automation Workflows |
| **Setup & Installation** | Standalone DMG / Setup + CLI Helper | `npm install -g @anthropic-ai/claude-code` | Package manager / CLI binary | Environment setup / Python packages |

---

## 🏛️ Architecture & Engine Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         AetherGrok Desktop Studio                           │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌───────────────────────┐             ┌──────────────────────────────────┐ │
│  │   Svelte 5 Frontend   │   IPC / WS  │        Go 1.24 Core Engine       │ │
│  │ ───────────────────── │ <─────────> │ ──────────────────────────────── │ │
│  │ • Ant Design Dark UI  │  Wails v2   │ • GrokRunner (Process Group)     │ │
│  │ • 10-Turn Windowing   │   Bridge    │ • 16ms Token Batching Streamer   │ │
│  │ • Diff2Html & KaTeX   │             │ • Screen Capture Orchestrator    │ │
│  │ • Skill Hub & Catalog │             │ • Native Skills Scanner & Reg    │ │
│  │ • Live Audio Reactive │             │ • GitHub Release Auto-Updater    │ │
│  └───────────────────────┘             └──────────────────────────────────┘ │
│              │                                          │                   │
│              ▼                                          ▼                   │
│  ┌───────────────────────┐             ┌──────────────────────────────────┐ │
│  │   DOM & State Mgmt    │             │   Subprocesses & OS Integrations │ │
│  │ • Svelte $state/$effect│             │ • `grok` CLI Process Tree (pgid) │ │
│  │ • Relative Scroll Lock│             │ • Darwin screencapture / Win GDI │ │
│  │ • Persistent Settings │             │ • CoreGraphics, AVFoundation, TCC│ │
│  └───────────────────────┘             └──────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## ✨ Key Features

### ⚡ 1. 10-Turn Windowing Engine & Frame-Rate Token Batching
- **Virtual DOM Pruning**: Maintains only the most recent conversational turns in the active DOM tree to guarantee fluid 60 FPS scrolling and responsive inputs.
- **Scroll Anchor Lock**: Preserves relative scroll positioning during viewport prepend actions.
- **16ms NDJSON Token Streamer**: Buffers streaming output from subprocesses and dispatches frame-aligned updates to eliminate micro-stutters.

### 📸 2. Non-Intrusive OS-Excluding Screen Capture
- **Compositor Synchronization**: Temporarily minimizes or hides the AetherGrok window, allows compositor sync (50ms macOS / 80ms Windows), captures the designated screen display via OS APIs, and immediately refocuses the window.
- **Instant Vision Attachment**: Injects captured snapshots directly into the prompt composer as context image chips.

### 🎙️ 3. Push-to-Talk Voice Dictation & System Audio Control
- **Hands-Free Dictation**: Supports push-to-talk and double-tap toggle voice shortcuts (`\`).
- **Real-Time Visual Feedback**: Displays a dynamic audio volume equalizer.
- **System Audio Ducking**: Automatically mutes active system playback during recording and restores initial volume levels upon completion.

### 🧩 4. Skills & Agent Tools Discovery
- **Local Ecosystem Scanning**: Automatically parses skills from `~/.grok/skills/` and `~/.agents/skills/`.
- **Interactive Composer Autocomplete**: Type `/` in the prompt bar to filter, inspect descriptions, and inject skills directly.

### 🔄 5. In-App Updates & Changelog Viewer
- **GitHub Releases Integration**: Queries `fiko942/aethergrok` for updates, compares SemVer tags, and provides single-click downloads and formatted release notes.

---

## 🛠️ Building from Source

### Prerequisites

| Tool | Minimum Requirement | Installation Guide |
| :--- | :--- | :--- |
| **Go** | `1.24+` | [golang.org/dl](https://go.dev/dl/) |
| **Node.js** | `20+` (with `pnpm` or `npm`) | [nodejs.org](https://nodejs.org) |
| **Wails CLI v2** | `v2.9+` | `go install github.com/wailsapp/wails/v2/cmd/wails@latest` |
| **C/C++ Compiler** | GCC / MSVC / Clang | **Windows**: MinGW-w64 (GCC) or MSVC<br>**macOS**: Xcode Command Line Tools (`xcode-select --install`) |
| **NSIS** *(Optional)* | `3.x+` | **Windows**: For generating Setup installer (`choco install nsis` or [nsis.sourceforge.io](https://nsis.sourceforge.io/)) |

---

### 🪟 Windows Automated Build

We provide an automated PowerShell script `build-windows.ps1` that builds frontend assets with Vite, compiles the Go backend with stripped GUI flags (`-s -w -H=windowsgui`), creates the portable ZIP package, and compiles the NSIS Setup Installer:

```powershell
# Clone the repository
git clone https://github.com/fiko942/aethergrok.git
cd aethergrok

# Run the automated Windows build script
powershell -ExecutionPolicy Bypass -File .\build-windows.ps1
```

#### Build Options & Flags:
```powershell
# Build for ARM64 architecture
powershell -ExecutionPolicy Bypass -File .\build-windows.ps1 -Arch arm64

# Skip test suite verification
powershell -ExecutionPolicy Bypass -File .\build-windows.ps1 -SkipTests

# Skip NSIS installer packaging (produces binary & zip only)
powershell -ExecutionPolicy Bypass -File .\build-windows.ps1 -SkipInstaller
```

**Artifacts Generated in `build\bin\`:**
- `aethergrok.exe` — Windows standalone GUI executable
- `aethergrok-windows-amd64.zip` — Portable release archive
- `AetherGrok-Setup.exe` — Windows NSIS setup wizard installer

---

### 🍏 macOS Automated Build

```bash
# Build macOS application bundle & DMG installer
./build-macos.sh arm64   # For Apple Silicon (M1/M2/M3/M4)
./build-macos.sh amd64   # For Intel Mac
```
*Artifacts Generated: `build/bin/AetherGrok-macOS-<arch>.dmg`*

---

### 💻 Live Development Mode (Hot-Reload)

**Windows (PowerShell):**
```powershell
# Start Wails Native Desktop Studio with Hot-Reload (Default)
.\dev.ps1

# Start Frontend UI only in Web Browser (Vite dev server)
.\dev.ps1 ui

# Run Go Unit Tests and Frontend Verification
.\dev.ps1 test
```

**macOS / Linux (Bash):**
```bash
# Start Wails Native Desktop Studio with Hot-Reload
./dev.sh

# Start Frontend UI only in Web Browser
./dev.sh ui

# Run Go Unit Tests
./dev.sh test
```

---

## ☕ Sponsorship & Donations

AetherGrok is developed independently as a high-performance open-source studio. If you find the software useful, consider supporting its maintenance and development:

- **Saweria (Indonesia / QRIS / GoPay / OVO / Dana)**: [saweria.co/wijifikoteren](https://saweria.co/wijifikoteren)
- **GitHub Sponsors (International)**: [github.com/sponsors/fiko942](https://github.com/sponsors/fiko942)

---

## 📄 License & Attribution

- **License**: Released under the [MIT License](LICENSE).
- **Trademarks**: Grok and xAI are trademarks of their respective owners. AetherGrok is an independent open-source desktop client designed for developer workflows.
