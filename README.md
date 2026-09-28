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

### Focused Operational Considerations
Different autonomous coding and agent environments are built with distinct architectural priorities:
1. **Long-Session UI Responsiveness**: Continuous multi-step agent conversations produce high volumes of rich diffs and terminal logs. Managing DOM memory pressure and scroll positioning is essential for long-running workflows.
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
