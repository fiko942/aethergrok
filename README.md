# AetherGrok Desktop Studio

<p align="center">
  <img src="resources/app-icon.png" width="128" height="128" alt="AetherGrok Logo" /><br>
  <b>High-Performance Native Desktop Studio for Grok Build CLI & Autonomous Engineering</b><br>
  <sub>Engineered with Go 1.24, Wails v2, Svelte 5 Runes, and Ant Design Dark Token Architecture</sub>
</p>

<p align="center">
  <a href="https://github.com/fiko942/aethergrok/releases/latest"><img src="https://img.shields.io/github/v/release/fiko942/aethergrok?color=1677ff&label=Latest%20Release" alt="Release" /></a>
  <a href="https://github.com/fiko942/aethergrok/releases"><img src="https://img.shields.io/github/downloads/fiko942/aethergrok/total?color=52c41a&label=Downloads" alt="Downloads" /></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License" /></a>
  <a href="https://go.dev"><img src="https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white" alt="Go" /></a>
  <a href="https://svelte.dev"><img src="https://img.shields.io/badge/Svelte-5.x-FF3E00?logo=svelte&logoColor=white" alt="Svelte" /></a>
  <a href="https://wails.io"><img src="https://img.shields.io/badge/Wails-v2-DF1A55?logo=wails&logoColor=white" alt="Wails" /></a>
</p>

<p align="center">
  <a href="#-direct-binary-downloads">Downloads</a> •
  <a href="#-why-aethergrok-the-problem--solution">Why AetherGrok</a> •
  <a href="#-technical-comparison-matrix">Comparison Matrix</a> •
  <a href="#-architecture--engine-flow">Architecture</a> •
  <a href="#-key-features">Features</a> •
  <a href="#-building-from-source">Build</a>
</p>

---

## 📦 Direct Binary Downloads

Download the latest signed release directly for macOS and Windows.

### macOS Installers (macOS 11+)
| Architecture | Format | Download Link | Checksum |
| :--- | :--- | :--- | :--- |
| **Apple Silicon (M1 / M2 / M3 / M4)** | Styled `.dmg` | [Download `AetherGrok-1.0.3-macOS-arm64.dmg`](https://github.com/fiko942/aethergrok/releases/download/v1.0.3/AetherGrok-1.0.3-macOS-arm64.dmg) | [SHA256](https://github.com/fiko942/aethergrok/releases/download/v1.0.3/AetherGrok-1.0.3-macOS-arm64.dmg.sha256) |
| **Intel x86_64** | Styled `.dmg` | [Download `AetherGrok-1.0.3-macOS-amd64.dmg`](https://github.com/fiko942/aethergrok/releases/download/v1.0.3/AetherGrok-1.0.3-macOS-amd64.dmg) | [SHA256](https://github.com/fiko942/aethergrok/releases/download/v1.0.3/AetherGrok-1.0.3-macOS-amd64.dmg.sha256) |

### Windows Packages (Windows 10 / 11 64-bit)
| Architecture | Format | Download Link | Checksum |
| :--- | :--- | :--- | :--- |
| **x64 (AMD64)** | Setup Installer `.exe` | [Download `AetherGrok-1.0.3-windows-amd64-setup.exe`](https://github.com/fiko942/aethergrok/releases/download/v1.0.3/AetherGrok-1.0.3-windows-amd64-setup.exe) | [Release Assets](https://github.com/fiko942/aethergrok/releases/tag/v1.0.3) |
| **x64 (AMD64)** | Standalone Portable `.zip` | [Download `AetherGrok-1.0.3-windows-amd64-portable.zip`](https://github.com/fiko942/aethergrok/releases/download/v1.0.3/AetherGrok-1.0.3-windows-amd64-portable.zip) | [Release Assets](https://github.com/fiko942/aethergrok/releases/tag/v1.0.3) |
| **ARM64** | Setup Installer `.exe` | [Download `AetherGrok-1.0.3-windows-arm64-setup.exe`](https://github.com/fiko942/aethergrok/releases/download/v1.0.3/AetherGrok-1.0.3-windows-arm64-setup.exe) | [Release Assets](https://github.com/fiko942/aethergrok/releases/tag/v1.0.3) |
| **ARM64** | Standalone Portable `.zip` | [Download `AetherGrok-1.0.3-windows-arm64-portable.zip`](https://github.com/fiko942/aethergrok/releases/download/v1.0.3/AetherGrok-1.0.3-windows-arm64-portable.zip) | [Release Assets](https://github.com/fiko942/aethergrok/releases/tag/v1.0.3) |

> 💡 **Auto-Update**: AetherGrok includes built-in background update notifications and one-click GitHub Release synchronization.

---

## 🎯 Why AetherGrok? The Problem & Solution

### The Friction in Contemporary AI Coding
Autonomous coding agents operating purely inside terminal emulators or webviews face distinct operational limitations:
1. **DOM Degradation & Memory Leaks**: Rendering long conversational turns with dozens of diff blocks and terminal outputs degrades webviews and Electron apps after continuous sessions.
2. **Terminal Visual Occlusion**: CLI agents cannot display inline rich graphical diffs, side-by-side file comparisons, drag-and-drop screenshots, or interactive skill catalogs alongside terminal executions.
3. **Flaky Background Process Management**: Stopping or steering an agent mid-turn in standard wrappers often leaves orphaned bash processes running or terminates child processes abruptly without state synchronization.
4. **Environment Bootstrapping Hassle**: Developers without pre-installed CLI toolchains encounter cryptic setup errors instead of managed dependency resolution.

### The AetherGrok Solution
AetherGrok bridges high-throughput Go backend engineering with Svelte 5 fine-grained reactivity to produce a native desktop GUI:
- **Low Memory Footprint**: Uses OS-native webview rendering through Wails v2 without bundling multi-hundred megabyte Chromium runtimes.
- **10-Turn Windowing Engine**: Maintains constant memory consumption across 100+ turn conversations by dynamically virtualizing DOM nodes while preserving viewport scroll anchors.
- **Mid-Turn Live Steering & Process Group Supervision**: Direct Unix process group isolation (`Setpgid`) and 16ms token batching enable zero-lag steering without process death.
- **Automated 1-Click Toolchain Detection & Installer**: Detects missing `grok` CLI binaries on startup and executes automated background Homebrew or shell installations directly from the UI.

---

## 📊 Technical Comparison Matrix

A factual comparison of desktop capabilities, architecture, resource usage, and interaction models across autonomous coding tools:

| Feature / Metric | **AetherGrok Desktop** | **Anthropic Claude Code** | **OpenAI Codex CLI** | **Google Antigravity / CUA** |
| :--- | :--- | :--- | :--- | :--- |
| **Primary Interface** | Native Desktop GUI (Wails v2 + Svelte 5) | Terminal CLI (Node.js) | Terminal CLI / ACP Wrapper | Desktop CUA / Web Sandbox |
| **Runtime Footprint** | ~35 MB RAM (Go + Native OS Webview) | ~180 MB RAM (Node.js runtime) | ~140 MB RAM (Python / Node) | ~450+ MB RAM (Electron / PyAutoGUI) |
| **Long Session Performance** | **10-Turn DOM Windowing** (Zero lag at 100+ turns) | Terminal scrollback buffer limit | Terminal stdout buffer limit | High DOM/Canvas re-render load |
| **Live Mid-Turn Steering** | **Supported** (Inject prompt while working) | Requires turn interrupt / cancellation | Partial support via ACP | Limited by UI action queue |
| **Process Group Control** | **POSIX `setpgid` & Windows Job Objects** | `child_process` process tree | System subprocess | OS accessibility input injection |
| **Screen Snapshot & Vision** | **Non-Intrusive OS-Exclusion Capture** | Manual file path reference | Manual file path reference | Full screen grab + OS pixel automation |
| **Voice Dictation & Push-to-Talk**| **Supported** (Real-time live audio bars & auto-mute)| Not supported | Not supported | Audio input dependent on host |
| **Visual Diff Viewer** | **Syntax-highlighted Side-by-Side & Unified Diffs**| Unified terminal ANSI diffs | Unified terminal ANSI diffs | Screenshot comparison |
| **Skills & Extensions Catalog** | **Visual Catalog (`~/.grok/skills`)** | CLI slash commands | Config file hooks | Pre-recorded automation graphs |
| **Toolchain Auto-Install** | **1-Click Built-in Installer Gate** | Manual `npm install -g` | Manual setup | Manual Docker / environment setup |

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
- **Hands-Free Dictation**: Supports push-to-talk and double-tap toggle voice shortcuts.
- **Real-Time Visual Feedback**: Displays a 4-bar dynamic audio volume equalizer.
- **System Audio ducking**: Automatically mutes active system playback during recording and restores initial volume levels upon completion.

### 🧩 4. Skills & Agent Tools Discovery
- **Local Ecosystem Scanning**: Automatically parses skills from `~/.grok/skills/` and `~/.agents/skills/`.
- **Interactive Composer Autocomplete**: Type `/` in the prompt bar to filter, inspect descriptions, and inject skills directly.

### 🔄 5. In-App Updates & Changelog Viewer
- **GitHub Releases Integration**: Queries `fiko942/aethergrok` for updates, compares SemVer tags, and provides single-click downloads and formatted release notes.

---

## 🛠️ Building from Source

### Prerequisites
- **Go**: 1.24+ ([golang.org](https://golang.org))
- **Node.js**: 20+ & npm ([nodejs.org](https://nodejs.org))
- **Wails CLI v2**: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- **Platform Compilers**:
  - **macOS**: Xcode Command Line Tools (`xcode-select --install`)
  - **Windows**: MinGW-w64 or MSVC

### Development Mode
```bash
# Clone the repository
git clone https://github.com/fiko942/aethergrok.git
cd aethergrok

# Install frontend dependencies
cd frontend && npm install && cd ..

# Start Wails live development server
wails dev
```

### Production Build
```bash
# Build macOS application bundle & DMG installer
./build-macos.sh arm64   # For Apple Silicon
./build-macos.sh amd64   # For Intel Mac

# Build Windows installer
npm run build:windows
```

---

## 📄 License & Attribution

- **License**: Released under the [MIT License](LICENSE).
- **Trademarks**: Grok and xAI are trademarks of their respective owners. AetherGrok is an independent open-source desktop client designed for developer workflows.
