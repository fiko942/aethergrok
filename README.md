# AetherGrok Desktop GUI Studio

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE) [![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white)](https://go.dev) [![Svelte 5](https://img.shields.io/badge/Svelte-5.x-FF3E00?logo=svelte&logoColor=white)](https://svelte.dev) [![Wails](https://img.shields.io/badge/Wails-v2-DF1A55?logo=wails&logoColor=white)](https://wails.io) [![Ant Design Dark](https://img.shields.io/badge/Design-Ant%20Design%20Dark-1677ff)](https://ant.design)

> **High-Performance Desktop GUI Studio for Grok Build CLI (including Grok 4.6)**.
> Built with Go 1.24, Wails v2, Svelte 5 Runes, and an Ant Design Dark design system. Non-intrusive screen capture, 10-turn progressive DOM windowing, real-time streaming NDJSON parser, and built-in skill discovery.

---

## 🏛️ Architectural Overview & Engine Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                      AetherGrok Desktop GUI Studio                          │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌───────────────────────┐             ┌──────────────────────────────────┐ │
│  │   Svelte 5 Frontend   │   IPC / WS  │        Go 1.24 Core Engine       │ │
│  │ ───────────────────── │ <─────────> │ ──────────────────────────────── │ │
│  │ • Ant Design Dark UI  │  Wails v2   │ • GrokRunner (Process Group)     │ │
│  │ • 10-Turn Windowing   │   Bridge    │ • 16ms Token Batching Streamer   │ │
│  │ • Diff2Html & KaTeX   │             │ • Screen Capture Orchestrator    │ │
│  │ • Skill Hub & Catalog │             │ • Native Skills Scanner & Reg    │ │
│  └───────────────────────┘             └──────────────────────────────────┘ │
│              │                                          │                   │
│              ▼                                          ▼                   │
│  ┌───────────────────────┐             ┌──────────────────────────────────┐ │
│  │   DOM & State Mgmt    │             │   Subprocesses & OS Integrations │ │
│  │ • Svelte $state/$effect│             │ • `grok` CLI Process Tree (pgid) │ │
│  │ • Relative Scroll Lock│             │ • Darwin screencapture / Win GDI │ │
│  │ • Low-overhead Render │             │ • ~/.grok/skills & ~/.agents/... │ │
│  └───────────────────────┘             └──────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 🚀 Key Features & Capabilities

### ⚡ 1. Ultra-Low Overhead & 10-Turn Windowing
- **Progressive Hydration**: Limits active DOM nodes to the latest 10 conversational turns, dynamically rendering hundreds of turns with zero UI stutter or memory bloat.
- **Scroll Anchor Preservation**: Prepend older turns seamlessly without jumping or resetting viewport position.
- **16ms Token Batching**: Streaming tokens from `grok` NDJSON are buffered and dispatched at 60fps frame rate intervals for smooth UI animations.

### 📸 2. Non-Intrusive OS-Excluding Screen Capture
- **Compositor Synchronization**: Automatically hides the AetherGrok window, sleeps for compositor sync (50ms on macOS, 80ms on Windows), captures the screen natively via OS APIs, and immediately restores and focuses the window.
- **Panic & Cancel Safe**: Guaranteed window restoration using deferred handlers and context cancellation propagation.

### 🧩 3. Built-in Skills Catalog & Hub
- **Universal Skill Discovery**: Scans `~/.grok/skills/` and `~/.agents/skills/` for `SKILL.md` definitions.
- **Frontmatter Parser**: Extracts name, version, author, description, tags, and category.
- **Direct Insertion**: Quick `/skill-name` injection into the Composer with autocomplete and search filtering.

### 🎨 4. Ant Design Dark Design System
- **Theme Tokens**: Complete integration of Ant Design dark palette (`#141414`, `#1f1f1f`, `#1677ff`, `#52c41a`, `#faad14`, `#ff4d4f`).
- **Rich Components**: Includes custom Svelte 5 Ant Design components (`Button`, `Card`, `Badge`, `Switch`, `Composer`, `DiffCard`, `ToolCallCard`, `PermissionModal`).
- **Interactive Unified Diffs**: Side-by-side and line-by-line diff viewing powered by syntax highlighting.

---

## 🖥️ UI Layout Mockup

```
┌──────────────────────────────────────────────────────────────────────────────────────┐
│ [✨ AetherGrok v1.0.0]        [✨ Skills Hub]  [🟢 Model: grok-4.6]  [⚡ Test Bridge] [⚙️]│
├───────────────┬──────────────────────────────────────────────────────────────────────┤
│ WORKSPACE     │ [ Session 1 ] [ Session 2 (Working) ] [+]                            │
│ 🤖 Active Sess├──────────────────────────────────────────────────────────────────────┤
│ 💻 Task Runner│ (turn 1) User: Check system readiness and status of AetherGrok engine│
│               │                                                                      │
│ CAPABILITIES  │ 🤖 Assistant: Diagnostic completed. Svelte 5 Runes & Tokens active.  │
│ 📸 Snap Screen│ ┌──────────────────────────────────────────────────────────────────┐ │
│ 🧩 Skills Reg │ │ 🔧 Tool Call: search_replace (src/config.ts)         [COMPLETED] │ │
│               │ │ ---------------------------------------------------------------- │ │
│ SETTINGS      │ │ - const TIMEOUT = 1000;                                          │ │
│ Model: grok4.6│ │ + const TIMEOUT = 5000;                                          │ │
│ Effort: Medium│ └──────────────────────────────────────────────────────────────────┘ │
│               ├──────────────────────────────────────────────────────────────────────┤
│               │ [ Attach Image ] [ / Skills ]                                        │
│               │ > Type your prompt or steer command...                   [ Send 🚀 ] │
└───────────────┴──────────────────────────────────────────────────────────────────────┘
```

---

## 🛠️ Project Structure

```
.
├── app.go                       # Go Wails backend application struct and bridge methods
├── main.go                      # Wails desktop window configuration and bootstrap
├── go.mod                       # Go 1.24 module definitions
├── pkg/
│   ├── grokrunner/              # Subprocess runner, process group supervisor & NDJSON parser
│   │   ├── runner.go            # Command lifecycle, cancel, stdin permission handling
│   │   ├── stream_parser.go     # 16ms token batching & event deserialization
│   │   ├── types.go             # Protocol event structs and types
│   │   ├── proc_unix.go         # SysProcAttr process grouping for Unix / macOS
│   │   └── proc_windows.go      # Process grouping and termination for Windows
│   ├── screen/                  # Non-intrusive screen capture orchestrator
│   │   ├── capture.go           # 4-phase coordination flow & WindowController
│   │   ├── capture_darwin.go    # macOS screencapture implementation
│   │   ├── capture_windows.go   # Windows GDI screen capture
│   │   └── capture_other.go     # Linux X11/Wayland screen capture
│   └── skills/                  # Extensible agent skills scanner & registry
│       ├── scanner.go           # Directory walker and SKILL.md frontmatter parser
│       ├── registry.go          # Memory cache and query engine
│       └── types.go             # Skill metadata types
├── frontend/                    # Svelte 5 desktop GUI frontend
│   ├── src/
│   │   ├── App.svelte           # Main window layout, sidebar, header, and Wails events
│   │   ├── lib/
│   │   │   ├── antd/            # Ant Design Dark tokens & primitives (Button, Card, Badge, Switch)
│   │   │   ├── components/
│   │   │   │   ├── chat/        # MessageList (10-turn window), Composer, DiffCard, ToolCallCard
│   │   │   │   ├── layout/      # SessionTabs, SettingsModal
│   │   │   │   ├── skills/      # SkillCatalog, SkillCard
│   │   │   │   └── snapshot/    # SnapshotBar
│   │   │   └── stores/          # Svelte 5 runes state (sessionStore, settingsStore)
│   ├── package.json
│   ├── vite.config.ts
│   └── tailwind.config.cjs
└── test/                        # Complete Go unit and integration test suite
    ├── runner_test.go           # Subprocess execution and cancel tests
    ├── screen_test.go           # Window hiding, delays, context cancel tests
    └── skills_test.go           # SKILL.md parsing and registry query tests
```

---

## 🏗️ Build & Development Instructions

### Prerequisites
- **Go**: `1.24+` installed ([go.dev](https://go.dev))
- **Node.js**: `18+` or `20+` & **pnpm** `9+` installed
- **Wails CLI**: `v2.8+` (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)

### 1. Run Unit Tests & Type Checks
```bash
# Run Go unit test suite
go test ./test/... -v

# Run Frontend Typecheck & Svelte Check
pnpm --prefix frontend run check

# Build Frontend Assets
pnpm --prefix frontend run build
```

### 2. Compile Desktop Application Binary
```bash
# Build standalone desktop Go binary
mkdir -p build/bin
go build -o build/bin/aethergrok main.go app.go
```

### 3. Run in Development Mode
```bash
# Start Wails live-reload development server
wails dev
```

---

## 🔒 Memory Leak & Zero-Overhead Guarantees

1. **Goroutine Cleanups**: All stream parsers and process supervisors utilize context cancellation and `sync.Mutex` locks. Active sessions are automatically purged upon turn completion or user cancellation.
2. **Process Group Isolation**: Spawns all subprocesses with `Setpgid: true` (Unix) and `CREATE_NEW_PROCESS_GROUP` (Windows) to ensure zero dangling child processes.
3. **DOM Virtualization**: Conversational turns exceeding the 10-turn active window are unmounted from the DOM, retaining only lightweight state objects in memory.
4. **Compositor Teardown**: Screen snapshot orchestrators strictly restore desktop focus and window visibility through deferred handlers under all conditions.

---

## 📜 License

This project is licensed under the [MIT License](LICENSE).
