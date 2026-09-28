# Grok CLI Auto-Detection, Native Installer & Diagnostic Logging Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Detect whether Grok CLI is installed upon startup, parse its exact version and binary location, provide an automated 1-click in-app installer for macOS (Homebrew or official `curl -fsSL https://x.ai/cli/install.sh | bash`), display a dedicated onboarding gate when missing, and record detailed diagnostic logs into `~/.grok/logs/aethergrok.log` for developer troubleshooting.

**Architecture:**
1. **Backend (`pkg/grokrunner` & `pkg/logger`)**:
   - `DetectGrokInstallation()` inspects `ResolveGrokBinary()`, executes `--version` to parse the SemVer string (e.g. `1.0.41`), and returns a detailed `GrokInstallStatus` struct.
   - `InstallGrokCLI()` executes the official installation sequence with fallback strategies (Homebrew if installed, otherwise official bash script), captures live stdout/stderr streams, writes timestamped diagnostic log entries via `logger.GetDiskLogger()`, and emits Wails events for progress tracking.
2. **Frontend (`GrokInstallModal.svelte` & `App.svelte`)**:
   - On app startup, `App.svelte` runs `CheckGrokInstallation()`.
   - If Grok CLI is missing (`!installed`), display `GrokInstallModal.svelte` as an onboarding gate matching active Ant Design design tokens.
   - Includes real-time installation progress, terminal output expander, retry handling, and 1-click copy diagnostic logs.

**Tech Stack:** Go, Wails v2, Svelte 5 (Runes), Tailwind CSS, Ant Design Design Tokens.

## Global Constraints
- Target platform: macOS first (with fallback stubs for other OSs).
- Full diagnostic logging: All installation checks, shell commands, stdout/stderr, and exit codes must be appended to `~/.grok/logs/aethergrok.log`.
- Strict theme compliance (`dark-studio`, `dark-high-contrast`, `light-antd`).
- Zero border glitches on dark mode.

---

### Task 1: Backend Grok Detection & Automated Installation Engine

**Files:**
- Create: `pkg/grokrunner/installer.go`
- Modify: `pkg/grokrunner/runner.go`
- Modify: `pkg/grokrunner/types.go`
- Modify: `app.go`

**Interfaces:**
- Produces:
  - `type GrokInstallStatus struct { Installed bool; Version string; BinaryPath string; Platform string; Error string }`
  - `type GrokInstallProgress struct { Stage string; Percent int; Message string; LogLine string }`
  - `DetectGrokInstallation() GrokInstallStatus`
  - `InstallGrokCLI() error`

- [ ] **Step 1: Implement `GrokInstallStatus` and `DetectGrokInstallation()` in `installer.go`**
- [ ] **Step 2: Implement `InstallGrokCLI()` streaming installer in `installer.go`**
  - Logs step-by-step progress via `logger.GetDiskLogger().Append(...)`.
  - Tries `curl -fsSL https://x.ai/cli/install.sh | bash` (or `brew install grok` if brew is available).
  - Emits Wails runtime events `grok:install_progress` and `grok:install_complete`.
- [ ] **Step 3: Expose `CheckGrokInstallation()` and `InstallGrokCLI()` in `app.go`**
- [ ] **Step 4: Verify Go build**
  Run: `export PATH=$PATH:/usr/local/go/bin:~/go/bin:/opt/homebrew/bin && go build -o /dev/null .`

---

### Task 2: Create `GrokInstallModal.svelte` (1-Click Grok CLI Setup Gate)

**Files:**
- Create: `frontend/src/lib/components/setup/GrokInstallModal.svelte`
- Modify: `frontend/src/app.d.ts`

- [ ] **Step 1: Update `frontend/src/app.d.ts` with Grok installer bindings & events**
- [ ] **Step 2: Build `GrokInstallModal.svelte`**
  - Theme-aware Ant Design card layout.
  - States: `not_installed`, `installing`, `success`, `error`.
  - 1-Click "Install Grok CLI" action.
  - Live animated progress bar + expandable terminal logs stream.
  - Diagnostic copy button ("Copy Error Logs") on failure.

---

### Task 3: Integrate Startup Detection & Modal in `App.svelte`

**Files:**
- Modify: `frontend/src/App.svelte`

- [ ] **Step 1: In `onMount` of `App.svelte`, check Grok installation status**
  - If missing, trigger `showGrokInstallModal = true`.
- [ ] **Step 2: Connect completion callback**
  - Once installed successfully, automatically re-initialize runner and continue to workspace.

---

### Task 4: Unit Testing & End-to-End Verification

- [ ] **Step 1: Run Vitest frontend test suite**
- [ ] **Step 2: Run Go unit tests and build**
- [ ] **Step 3: Test disk logger persistence**
