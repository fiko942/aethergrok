# AetherGrok Desktop GUI - Final Implementation Writing Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Membangun aplikasi Desktop GUI cross-platform (macOS & Windows) berkinerja tinggi, hemat resource, dan berorientasi efisiensi untuk orkestrasi Grok Agentic AI CLI (`grok`), mengintegrasikan Svelte 5, Ant Design system styling tokens, native Go backend bridge, non-intrusive smart screen snapshot dengan window auto-hiding, virtualized multi-session management (10-turn windowing), serta Skill/MCP catalog.

**Architecture:** Arsitektur modular dua lapis:
1. **Frontend Presentation Layer:** Svelte 5 (Runes reactivity, zero-overhead compile-time reactivity) + Ant Design System styling/tokens (`@ant-design/colors`, Ant Design components) dengan Virtualized DOM turn manager (10 turn active windowing).
2. **Core OS Native & Process Engine:** Go (Golang) / Wails native bridge runtime yang mengelola IPC, streaming headless child process (`grok --output-format streaming-json`), sliding-window token batching (16ms interval), native screen capture dengan window auto-hiding, filesystem watcher `~/.grok`, serta SQLite session cache lokal.

**Tech Stack:**
- Frontend: Svelte 5, Vite, TypeScript, Tailwind CSS, Ant Design Color & Token System, Lucide Svelte / Ant Design Icons Svelte, KaTeX, diff2html.
- Native Backend: Go 1.22+, Wails v2 bridge runtime.
- Native APIs: macOS CoreGraphics (`CGDisplayCreateImage`), Windows Win32 API (`BitBlt`, `User32.dll`).
- AI Engine Driver: Grok CLI headless runner (`grok -p`, `grok agent stdio`, `--output-format streaming-json`).

---

## Global Constraints

- Platform Target: macOS (Apple Silicon & Intel) dan Windows (x64).
- Memory Budget: Idle RAM di bawah 60MB, 0% CPU saat tidak ada proses aktif.
- Visual Language: Ant Design Dark Theme Tokens (`#1677FF`, `#0F1117`, `#181B26`, `#222634`, `#2E3446`, `#F3F4F6`).
- Language: English untuk interface UI.
- No heavy frameworks: Menggunakan Svelte 5 (bukan React/Angular) untuk efisiensi komputasi maksimal.

---

## File Structure & Responsibilities

```
grok-build/
├── docs/superpowers/plans/
│   └── 2026-09-25-grok-build-desktop-gui.md   # Dokumen implementasi utama
├── app.go                                    # Entrypoint bridge Go ↔ Svelte, window control & lifecycle
├── main.go                                   # Inisialisasi Wails application & options
├── wails.json                                # Konfigurasi Wails build & metadata
├── pkg/
│   ├── grokrunner/
│   │   ├── types.go                          # Tipe data ACP NDJSON Grok stream & options
│   │   ├── stream_parser.go                  # Parser stream stdout & batching token (16ms buffer)
│   │   └── runner.go                         # Subprocess execution, cancellation, Process Group supervisor
│   ├── screen/
│   │   ├── capture.go                        # Orchestrator capture: hide window -> sleep -> capture -> show window
│   │   ├── capture_darwin.go                 # Implementasi CoreGraphics macOS capture
│   │   └── capture_windows.go                # Implementasi GDI BitBlt Windows capture
│   ├── session/
│   │   ├── cache.go                          # SQLite database lokal untuk metadata & search index
│   │   └── watcher.go                        # fsnotify watcher untuk ~/.grok/sessions
│   └── skills/
│       ├── scanner.go                        # Parser frontmatter SKILL.md dari ~/.grok/skills/
│       └── registry.go                       # In-memory registry & filtered search catalog
└── frontend/
    ├── package.json
    ├── svelte.config.js
    ├── vite.config.ts
    ├── tailwind.config.cjs
    └── src/
        ├── app.d.ts
        ├── main.ts
        ├── App.svelte
        ├── lib/
        │   ├── antd/
        │   │   ├── tokens.ts                 # Ant Design 10-level color tokens & variables
        │   │   ├── Button.svelte             # Ant Design button component
        │   │   ├── Card.svelte               # Container card dengan border subtle
        │   │   ├── Modal.svelte              # Accessible modal dialog
        │   │   ├── Switch.svelte             # Toggle switch
        │   │   └── Badge.svelte              # Status badge indicator
        │   ├── components/
        │   │   ├── layout/
        │   │   │   ├── Header.svelte         # Top navigation bar, model picker, theme switch
        │   │   │   ├── SessionTabs.svelte    # Drag & drop multi-session tabs dengan status dots
        │   │   │   └── SettingsModal.svelte  # Ant Design modal setting konfigurasi
        │   │   ├── chat/
        │   │   │   ├── MessageList.svelte    # Virtualized message list (10-turn windowing)
        │   │   │   ├── MessageItem.svelte    # Render satu turn percakapan
        │   │   │   ├── Composer.svelte       # Textarea input, vision attachment bar, snapshot trigger
        │   │   │   ├── ToolCallCard.svelte   # Collapsible tool execution card (in/out blocks)
        │   │   │   ├── DiffCard.svelte       # Code diff preview (+green / -red)
        │   │   │   └── PermissionModal.svelte# Approval dialog (Allow once, Always allow, Reject)
        │   │   ├── snapshot/
        │   │   │   └── SnapshotBar.svelte    # Thumbnail preview snapshot yang siap dikirim
        │   │   └── skills/
        │   │       ├── SkillCatalog.svelte   # Modal browser skills & MCP tools
        │   │       └── SkillCard.svelte      # Kartu individual skill dengan tombol satu klik
        │   └── stores/
        │       ├── session.svelte.ts         # Svelte 5 reactive session store (Runes)
        │       ├── grok.svelte.ts            # Handler event stream dan buffer UI
        │       └── settings.svelte.ts        # Persistent app settings store
```

---

## Bite-Sized Implementation Tasks

### Task 1: Scaffolding Wails v2 Project, Svelte 5 Setup & Ant Design Tokens

**Files:**
- Create: `wails.json`
- Create: `main.go`
- Create: `app.go`
- Create: `frontend/package.json`
- Create: `frontend/vite.config.ts`
- Create: `frontend/svelte.config.js`
- Create: `frontend/tailwind.config.cjs`
- Create: `frontend/src/lib/antd/tokens.ts`
- Create: `frontend/src/lib/antd/Button.svelte`
- Create: `frontend/src/lib/antd/Card.svelte`

**Interfaces:**
- Produces: Base executable application shell dengan Go IPC bridge aktif dan styling token Ant Design (`#1677FF`, `#0F1117`, `#181B26`).

- [ ] **Step 1: Write `wails.json` configuration**
```json
{
  "$schema": "https://wails.io/schemas/config.v2.json",
  "name": "aethergrok",
  "outputfilename": "aethergrok",
  "frontend:install": "pnpm install",
  "frontend:build": "pnpm build",
  "frontend:dev:watcher": "pnpm dev",
  "frontend:dev:serverUrl": "auto",
  "author": { "name": "fiko942", "email": "tobellord@gmail.com" },
  "version": "1.0.0"
}
```

- [ ] **Step 2: Write minimal `main.go` and `app.go` Wails lifecycle**
- [ ] **Step 3: Setup `frontend/package.json` with Svelte 5, Tailwind, and Ant Design colors**
- [ ] **Step 4: Create Ant Design token palette utility in `frontend/src/lib/antd/tokens.ts`**
- [ ] **Step 5: Verify build compilation without errors**
- [ ] **Step 6: Commit changes**
```bash
git add wails.json main.go app.go frontend/
git commit -m "feat(core): scaffold Wails v2 application with Svelte 5 and Ant Design token system"
```

---

### Task 2: Core Grok Process Runner & 16ms Sliding-Window Stream Parser (Go Backend)

**Files:**
- Create: `pkg/grokrunner/types.go`
- Create: `pkg/grokrunner/stream_parser.go`
- Create: `pkg/grokrunner/runner.go`
- Modify: `app.go`
- Test: `test/grokrunner_test.go`

**Interfaces:**
- Consumes: User prompt, model selection, reasoning effort flag.
- Produces: `RunPromptStream(sessionId, prompt, options)`, emits batched event `grok:delta_batch` (interval 16ms), `grok:tool_call`, `grok:permission_request`, `grok:complete`.

- [ ] **Step 1: Write failing unit test for 16ms token batching in `test/grokrunner_test.go`**
- [ ] **Step 2: Run test to verify it fails**
- [ ] **Step 3: Implement NDJSON struct types in `pkg/grokrunner/types.go`**
- [ ] **Step 4: Implement sliding-window stream parser with 16ms flush ticker in `pkg/grokrunner/stream_parser.go`**
- [ ] **Step 5: Implement subprocess lifecycle and Process Group kill logic in `pkg/grokrunner/runner.go`**
- [ ] **Step 6: Run test to verify it passes**
- [ ] **Step 7: Commit changes**
```bash
git add pkg/grokrunner/ test/grokrunner_test.go app.go
git commit -m "feat(runner): implement headless Grok runner with 16ms sliding-window stream parser"
```

---

### Task 3: Smart Non-Intrusive Screen Snapshot Engine with Window Auto-Hiding

**Files:**
- Create: `pkg/screen/capture.go`
- Create: `pkg/screen/capture_darwin.go`
- Create: `pkg/screen/capture_windows.go`
- Modify: `app.go`
- Test: `test/screen_test.go`

**Interfaces:**
- Produces: `CaptureScreenExcludingSelf(delayMs int) (base64Png string, err error)`

- [ ] **Step 1: Write test validating coordination sequence (Hide -> Delay -> Capture -> Show)**
- [ ] **Step 2: Implement macOS CoreGraphics screen grab in `pkg/screen/capture_darwin.go`**
- [ ] **Step 3: Implement Windows GDI BitBlt screen grab in `pkg/screen/capture_windows.go`**
- [ ] **Step 4: Implement orchestrator in `pkg/screen/capture.go` with 50ms/80ms compositor synchronization**
- [ ] **Step 5: Expose `CaptureScreenExcludingSelf` method in `app.go`**
- [ ] **Step 6: Run tests and verify zero ghost-window artifacts**
- [ ] **Step 7: Commit changes**
```bash
git add pkg/screen/ test/screen_test.go app.go
git commit -m "feat(screen): implement native non-intrusive snapshot engine with auto-window hiding"
```

---

### Task 4: Multi-Session Docking & Virtualized 10-Turn Windowing (Svelte 5)

**Files:**
- Create: `frontend/src/lib/stores/session.svelte.ts`
- Create: `frontend/src/lib/components/layout/SessionTabs.svelte`
- Create: `frontend/src/lib/components/chat/MessageList.svelte`
- Create: `frontend/src/lib/components/chat/MessageItem.svelte`

**Interfaces:**
- Consumes: Svelte 5 session state.
- Produces: Multi-tab session dock with drag & drop reordering, status indicators (🔵, 🟡, 🟢, 🔴, ⚪), and virtualized 10-turn DOM windowing.

- [ ] **Step 1: Create reactive session store in `frontend/src/lib/stores/session.svelte.ts` using Svelte 5 `$state` and `$derived`**
- [ ] **Step 2: Build `SessionTabs.svelte` with drag & drop reorder and Ant Design badge status dots**
- [ ] **Step 3: Implement `MessageList.svelte` with 10-turn windowing and scroll-to-top infinite loading**
- [ ] **Step 4: Verify fast session switching without DOM freeze**
- [ ] **Step 5: Commit changes**
```bash
git add frontend/src/lib/stores/session.svelte.ts frontend/src/lib/components/layout/SessionTabs.svelte frontend/src/lib/components/chat/
git commit -m "feat(ui): implement multi-session tabs and virtualized 10-turn chat message windowing"
```

---

### Task 5: Ant Design Chat Feed, Tool Calls, Diff Inspector & Composer

**Files:**
- Create: `frontend/src/lib/components/chat/Composer.svelte`
- Create: `frontend/src/lib/components/chat/ToolCallCard.svelte`
- Create: `frontend/src/lib/components/chat/DiffCard.svelte`
- Create: `frontend/src/lib/components/chat/PermissionModal.svelte`
- Create: `frontend/src/lib/components/snapshot/SnapshotBar.svelte`

**Interfaces:**
- Produces: Rich interactive chat feed rendering Markdown streaming, inline diff inspector (+green / -red), collapsible tool execution logs, and snapshot attachment bar.

- [ ] **Step 1: Implement `Composer.svelte` with snapshot trigger button, model selector chip, and vision image chips**
- [ ] **Step 2: Implement `ToolCallCard.svelte` rendering command execution IN/OUT blocks**
- [ ] **Step 3: Implement `DiffCard.svelte` using `diff2html` styling for visual code change review**
- [ ] **Step 4: Implement `PermissionModal.svelte` for action approvals (Allow once, Always allow, Reject)**
- [ ] **Step 5: Verify seamless interaction from prompt typing to tool inspection**
- [ ] **Step 6: Commit changes**
```bash
git add frontend/src/lib/components/chat/ frontend/src/lib/components/snapshot/
git commit -m "feat(chat): implement Ant Design composer, tool cards, diff inspector, and permission modal"
```

---

### Task 6: Skills & MCP Tools Discovery Hub

**Files:**
- Create: `pkg/skills/scanner.go`
- Create: `pkg/skills/registry.go`
- Create: `frontend/src/lib/components/skills/SkillCatalog.svelte`
- Create: `frontend/src/lib/components/skills/SkillCard.svelte`
- Modify: `app.go`

**Interfaces:**
- Produces: Visual modal exploring all installed skills in `~/.grok/skills/` and `~/.agents/skills/` with instant parameter form generator.

- [ ] **Step 1: Implement Go scanner parsing frontmatter metadata from `SKILL.md` files**
- [ ] **Step 2: Implement search and category filter in `pkg/skills/registry.go`**
- [ ] **Step 3: Build `SkillCatalog.svelte` modal with Ant Design card grid and search input**
- [ ] **Step 4: Connect one-click skill injection into active Composer input**
- [ ] **Step 5: Commit changes**
```bash
git add pkg/skills/ frontend/src/lib/components/skills/ app.go
git commit -m "feat(skills): implement local skills and MCP tools discovery catalog"
```

---

### Task 7: Settings Modal & Ant Design Theme Customization

**Files:**
- Create: `frontend/src/lib/stores/settings.svelte.ts`
- Create: `frontend/src/lib/components/layout/SettingsModal.svelte`
- Modify: `frontend/src/App.svelte`

**Interfaces:**
- Produces: Settings modal managing Grok model selection, reasoning effort, auto-accept permissions, and dark/high-contrast themes.

- [ ] **Step 1: Implement persistent localStorage settings store in Svelte 5**
- [ ] **Step 2: Build `SettingsModal.svelte` with Ant Design tabbed layout (General, Models, Themes, Shortcuts)**
- [ ] **Step 3: Wire settings parameters into headless Grok execution arguments**
- [ ] **Step 4: Commit changes**
```bash
git add frontend/src/lib/stores/settings.svelte.ts frontend/src/lib/components/layout/SettingsModal.svelte frontend/src/App.svelte
git commit -m "feat(settings): implement Ant Design settings modal and theme customization"
```

---

### Task 8: Performance Audit, Cross-Platform Packaging & Final Verification

- [ ] **Step 1: Audit RAM consumption in idle state (<60MB target)**
- [ ] **Step 2: Audit CPU consumption at rest (0% target)**
- [ ] **Step 3: Build macOS application bundle (`.app` / `.dmg`)**
- [ ] **Step 4: Build Windows binary (`.exe`)**
- [ ] **Step 5: Verify end-to-end functionality (Snapshot -> Prompt -> Tool Call -> Diff Review -> Multi-Session)**
- [ ] **Step 6: Commit final release tag**
```bash
git commit -m "release: v1.0.0 AetherGrok Desktop GUI Studio"
```
