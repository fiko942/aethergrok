# AetherGrok Master Superpowers Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement 14 interconnected core UX, performance, animation, skill scanning, drag-and-drop, and session management superpowers in AetherGrok Desktop GUI Studio.

**Architecture:** 
- Backend (Go): Expand skill discovery paths to include symlinks, `~/.grok/bundled/skills`, `~/.claude/skills`, and `~/.codex/skills`; deduplicate scanning by canonical skill name; add OS-native config file reveal in Finder/Explorer; add Plan Gate auto-bypass handler.
- Frontend (Svelte 5 Runes & Tailwind): Global drag-and-drop file attachment zone covering the entire chat window; universal MIME-type icon mapping with zero harsh white borders; real-time request/response latency timers; accurate token usage estimation; inline tab session renaming with Enter/Esc hotkeys; customizable animation engine with global settings toggle (default active); custom radio and animated checkbox components in Skill Importer.

**Tech Stack:** Go (Wails v2), Svelte 5 (Runes `$state`, `$derived`, `$effect`), Tailwind CSS, Lucide Icons.

## Global Constraints
- Clean, refined design without harsh borders or generic AI slop.
- Use affirmative language in all UI copy and documentation.
- All animations must respect the `animationsEnabled` setting (default: true).
- Preserve backwards compatibility with existing workspace and session state formats.

---

### Task 1: Comprehensive Skill Discovery & Deduplication (Backend & Autocomplete)

**Files:**
- Modify: `pkg/skills/scanner.go`
- Modify: `pkg/skills/registry.go`
- Modify: `pkg/skills/importer.go`
- Modify: `frontend/src/lib/components/chat/composer/SlashMenu.svelte`
- Test: `pkg/skills/scanner_test.go`

- [ ] **Step 1: Update Go scanner to follow symlinks and include all standard skill roots**
  - Add directories: `~/.grok/skills/`, `~/.grok/bundled/skills/`, `~/.agents/skills/`, `~/.claude/skills/`, `~/.codex/skills/`.
  - In `ScanSkillDirectories`, resolve symlinks (e.g. `~/.agents/skills/superpowers`) and recursively scan symlinked directories with cycle detection.
  - In `ScanGitHubRepo`, deduplicate discovered skills by unique name and clean path.

- [ ] **Step 2: Write unit test in Go verifying symlink discovery and deduplication**
  - Run `go test -v ./pkg/skills/...` to ensure `superpowers` symlink and all skills resolve cleanly.

- [ ] **Step 3: Update `SlashMenu.svelte` to fuzzy search across all scanned skills and actions without truncation**

---

### Task 2: Universal Chat Drag & Drop and MIME-Type File Badges

**Files:**
- Modify: `frontend/src/lib/components/chat/ChatContainer.svelte`
- Modify: `frontend/src/lib/components/chat/composer/ComposerBox.svelte`
- Modify: `frontend/src/lib/components/chat/composer/AttachmentPills.svelte`
- Modify: `frontend/src/lib/utils/fileIcons.ts` (or create if needed)

- [ ] **Step 1: Add drag & drop event listeners to ChatContainer**
  - Handle `dragover`, `dragenter`, `dragleave`, and `drop` on the entire chat container.
  - Provide visual dropzone overlay with smooth dashed ring styling.

- [ ] **Step 2: Implement rich MIME-type and file extension icon resolver**
  - Support: `.zip`, `.tar`, `.gz`, `.7z` (Archive); `.js`, `.ts`, `.svelte`, `.py`, `.go`, `.rs`, `.c`, `.cpp`, `.json`, `.yaml` (Code); `.png`, `.jpg`, `.jpeg`, `.webp`, `.gif`, `.svg` (Image); `.pdf` (PDF); `.md`, `.txt`, `.doc`, `.docx` (Document); `.mp3`, `.wav`, `.ogg` (Audio); `.mp4`, `.mov`, `.webm` (Video).
  - Eliminate white borders, use subtle translucent backgrounds (`bg-white/[0.04] border border-white/[0.08]`).

---

### Task 3: Plan Gate Bypass Mode in Settings & Execution Flow

**Files:**
- Modify: `frontend/src/lib/stores/settings.svelte.ts`
- Modify: `frontend/src/lib/components/settings/SettingsModal.svelte`
- Modify: `frontend/src/lib/components/chat/PlanReviewCard.svelte`
- Modify: `frontend/src/lib/stores/chat.svelte.ts`

- [ ] **Step 1: Add `planGateMode: 'active' | 'bypass'` to settings store**
  - Default: `'active'`.
  - Add setting option in General/Agent Settings tab with clear description.

- [ ] **Step 2: Implement auto-approve and execute logic in `PlanReviewCard.svelte` and chat loop**
  - When `planGateMode === 'bypass'` and a plan gate is triggered, automatically trigger `onApprove()` / implement without blocking on user button press.

---

### Task 4: Request/Response Latency & Waiting Live Timers

**Files:**
- Modify: `frontend/src/lib/components/chat/ToolCallItem.svelte`
- Modify: `frontend/src/lib/components/chat/MessageItem.svelte`
- Modify: `frontend/src/lib/components/chat/composer/ComposerBox.svelte`
- Modify: `frontend/src/lib/stores/chat.svelte.ts`

- [ ] **Step 1: Add live elapsed timer during active AI agent requests**
  - Track `requestStartTime` when user prompt is submitted.
  - Show ticking timer ("Waiting for response: 1.4s" / "120ms") with glowing micro-indicator while agent is thinking.
  - Display final response duration ("Responded in 1.8s") on completed assistant messages.
  - Keep individual tool execution durations ("Ran in 42ms") on tool actions.

---

### Task 5: Accurate Token Usage Estimation & Clean Compact Banner

**Files:**
- Modify: `frontend/src/lib/components/chat/TokenUsagePopover.svelte`
- Modify: `frontend/src/lib/components/chat/ContextCompactBanner.svelte`
- Modify: `frontend/src/lib/stores/session.svelte.ts`

- [ ] **Step 1: Initialize baseline system token consumption on new sessions**
  - Base prompt and system instruction tokens (typically 1.8k - 3.2k tokens) reflected accurately in token usage bar.
  - Calculate active usage against model context limit (e.g. 128k, 200k, 1M).

- [ ] **Step 2: Redesign Compact Banner with anti-AI-slop minimalist aesthetic**
  - Smooth pill with subtle gradient border and compact action triggers.

---

### Task 6: Reveal Config File in OS Finder / File Explorer

**Files:**
- Modify: `app.go`
- Modify: `pkg/config/paths.go` (or helper)
- Modify: `frontend/src/lib/components/settings/tabs/ModelsTab.svelte`

- [ ] **Step 1: Implement `RevealGrokConfigFile()` in Go backend**
  - Detect active config path (`~/.grok/user-settings.json` or `~/.grok/config.json`).
  - On macOS: run `open -R <path>`.
  - On Windows: run `explorer.exe /select,<path>`.
  - On Linux: run `xdg-open <dir>`.

- [ ] **Step 2: Wire up "Open Config File" button in Models & Reasoning settings**

---

### Task 7: Fix Session Renaming Bug & Tab Inline Edit UX

**Files:**
- Modify: `frontend/src/lib/components/layout/TabsBar.svelte`
- Modify: `frontend/src/lib/components/layout/WorkspaceSidebar.svelte`
- Modify: `frontend/src/lib/stores/session.svelte.ts`

- [ ] **Step 1: Prevent automatic title overwrite if user has custom-renamed session**
  - Add `customTitle: boolean` flag in session metadata.
  - Only auto-generate title on first prompt if `customTitle` is false.

- [ ] **Step 2: Redesign Tab Inline Rename Input**
  - Remove harsh borders; clean focus ring.
  - Keep both Confirm (Check) and Cancel (X) buttons permanently visible during edit mode with flat neutral icons.
  - Bind keyboard `Enter` to Save, `Escape` to Cancel.

---

### Task 8: Comprehensive Animation Engine & Custom UI Components

**Files:**
- Modify: `frontend/src/lib/stores/settings.svelte.ts`
- Modify: `frontend/src/lib/components/settings/tabs/GeneralTab.svelte`
- Modify: `frontend/src/lib/components/ui/CustomRadio.svelte` (create)
- Modify: `frontend/src/lib/components/skills/SkillImporterModal.svelte`
- Modify: `frontend/src/lib/components/layout/TabsBar.svelte`
- Modify: `frontend/src/lib/components/layout/WorkspaceSidebar.svelte`
- Modify: `frontend/src/lib/components/layout/RightSidebar.svelte`
- Modify: `frontend/src/lib/components/ui/AnimatedSwitch.svelte`

- [ ] **Step 1: Add `animationsEnabled` setting (default: true)**
- [ ] **Step 2: Build CustomRadio and integrate CustomCheckbox in Skill Importer**
- [ ] **Step 3: Implement unique animated transitions for Sidebar collapse, Tabs peel effect, Action groups expand, and Right Sidebar Explorer/Changes switch**
- [ ] **Step 4: Clean up all remaining harsh white borders across headers, modals, and toolbars**

---

### Task 9: Verification, Superpowers Memory Sync, Git Commit & Push

**Files:**
- Modify: Memory and repository commit records

- [ ] **Step 1: Run full frontend build (`npm run build`) and Go build (`go build`)**
- [ ] **Step 2: Verify all 14 features end-to-end**
- [ ] **Step 3: Commit and push changes to GitHub**
