# Enterprise-Grade Global Application Logging Architecture & Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement an enterprise-grade, resource-optimized, full-spectrum logging architecture across all frontend UI components, state stores, user interactions, and backend IPC bridges, complete with a dedicated log management & export UI in Settings and strict architectural guidelines for future features.

**Architecture:** 
1. **Frontend Core Logger Store (`frontend/src/lib/stores/logger.svelte.ts`)**:
   - High-throughput circular in-memory ring buffer (default 1000 items, <500KB RAM).
   - Categorized log levels (`DEBUG`, `INFO`, `WARN`, `ERROR`).
   - Domain contexts (`UI`, `SESSION`, `TERMINAL`, `VOICE`, `BACKEND`, `SETTINGS`, `SYSTEM`).
   - Global unhandled promise rejection and window error hooks.
2. **Settings UI Log Management (`frontend/src/lib/components/layout/SettingsModal.svelte` & `frontend/src/lib/components/settings/LogsTab.svelte`)**:
   - Real-time search and level/category filter.
   - Live log streaming viewer with auto-scroll and pause toggle.
   - Multi-format exporter (JSON file, structured `.log` text file, clipboard copy).
   - Instant log purge and retention settings.
3. **Comprehensive Instrumentation**:
   - Instrument all existing key lifecycle points (SessionStore, TerminalStore, SettingsStore, Composer, VoiceRecorder, App IPC).
4. **Project Engineering Standards (`docs/superpowers/specs/logging-architecture-and-standards.md`)**:
   - Formal specification mandating how new features must implement structured logging.

**Tech Stack:** Svelte 5 Runes (`$state`, `$derived`), TypeScript, Lucide Icons, Ant Design aesthetic styling, Wails v2 Go backend.

## Global Constraints
- State points directly in affirmative language. Avoid unnecessary contrastive negation.
- Zero perceptible performance impact: logging operations must be non-blocking.
- High memory and storage efficiency: bounded circular buffers and strict size caps.
- Full light & dark mode visual compatibility conforming to Ant Design aesthetic standards.

---

### Task 1: Create Core Logging Engine & Ring Buffer Store

**Files:**
- Create: `frontend/src/lib/stores/logger.svelte.ts`

**Interfaces:**
- Produces: `logger` singleton with `.debug()`, `.info()`, `.warn()`, `.error()`, `.getLogs()`, `.clear()`, `.exportLogs(format)`.

- [ ] **Step 1: Implement `frontend/src/lib/stores/logger.svelte.ts`**

Implement the bounded circular logger with timestamps, levels, categories, optional metadata, global error traps, and export utilities.

- [ ] **Step 2: Connect global error and unhandled rejection listeners**

Ensure `window.addEventListener('error')` and `window.addEventListener('unhandledrejection')` automatically capture stack traces into `logger.error`.

---

### Task 2: Build Settings Logs Tab Component with Live Filters & Multi-Format Exporter

**Files:**
- Create: `frontend/src/lib/components/settings/LogsTab.svelte`
- Modify: `frontend/src/lib/components/layout/SettingsModal.svelte`

**Interfaces:**
- Consumes: `logger` store
- Produces: Interactive log inspection panel in Settings with search, level filtering, log detail view, JSON/TXT export, and clipboard copy.

- [ ] **Step 1: Create `LogsTab.svelte`**
- [ ] **Step 2: Add `logs` tab to `SettingsModal.svelte` with tab navigation and badge counters**

---

### Task 3: Instrument Full-Spectrum Application Events Across All Features

**Files:**
- Modify: `frontend/src/lib/stores/session.svelte.ts` (session hydration, tab switching, workspace changes, session deletion/rename)
- Modify: `frontend/src/lib/stores/terminal.svelte.ts` (terminal creation, resize, collapse, rename, close)
- Modify: `frontend/src/lib/stores/settings.svelte.ts` (settings update, shortcut customization)
- Modify: `frontend/src/App.svelte` (app boot, IPC bridge invocations, window resize, permission prompts)
- Modify: `frontend/src/lib/components/chat/Composer.svelte` (message send, queue manipulations, model selection)
- Modify: `frontend/src/lib/utils/voiceRecorder.ts` (microphone permission checks, audio recording start/stop)

- [ ] **Step 1: Instrument Session and Workspace state actions**
- [ ] **Step 2: Instrument Terminal lifecycle and PTY actions**
- [ ] **Step 3: Instrument Composer, Prompt Queue, and Voice Recorder**
- [ ] **Step 4: Instrument App initialization and IPC bridge calls**

---

### Task 4: Create Engineering Guidelines Document for Future Features

**Files:**
- Create: `docs/superpowers/specs/logging-architecture-and-standards.md`

- [ ] **Step 1: Write exhaustive specification for logging standards, rules for adding logs to new features, and resource efficiency requirements.**

---

### Task 5: Build, Verify and Package Native macOS Application

**Files:**
- Run: `cd frontend && npm run build`
- Run: `./build-macos.sh`

- [ ] **Step 1: Test TypeScript and Svelte build**
- [ ] **Step 2: Build and verify macOS native application package**
