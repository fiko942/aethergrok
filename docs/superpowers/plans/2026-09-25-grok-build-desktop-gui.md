# 2026-09-25: AetherGrok (Grok Desktop GUI) Comprehensive Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Transform the Grok Desktop client into an ultra-high performance, lightweight GUI studio (AetherGrok) for the Grok Agentic AI CLI with Ant Design component system tokens, Svelte 5 lightweight client layer, native smart screen snapshots with application exclusion, multi-session management, and rich visual tool rendering.

**Architecture:** 
1. **Frontend Presentation & Design System:** Svelte 5 / Ant Design Token System (`@ant-design/colors`, Ant Design components) with optimized reactive stores.
2. **Native Orchestration Layer:** Native system bridge for smart desktop snapshot (auto-hide self window, capture active display, restore window), session persistence parser (`~/.grok/sessions`), and streaming NDJSON parser for `grok` CLI agent execution.

**Tech Stack:**
- Frontend: Svelte 5, TypeScript, Tailwind CSS with Ant Design Tokens (`#1677FF`, `#0F1117`, `#181B26`), Ant Design Icons Svelte, Lucide.
- Desktop Bridge: Go / Wails v2 (or lightweight embedded bridge runtime).
- AI Engine: Grok Build CLI (`grok --output-format streaming-json`, `grok agent stdio`).

---

## 1. Target Audience, Brand Identity, & User Persona

### A. Target Audience & Customer Demographics
* **Age Bracket:** 20 – 45 years old.
* **Gender Demographics:** All developers, engineers, and technical creators (neutral, modern, professional tech ergonomics).
* **Professions:** Full-stack Software Engineers, AI/ML Engineers, DevOps/SREs, System Architects, Technical Tech Leads.
* **Hardware Profile:** Developers working on constrained or efficiency-focused laptops (MacBook Air/Pro, Windows Ultrabooks, low-spec dual-core devices) requiring minimal RAM (<60MB idle) and zero background lag.

### B. Brand Identity & Design System Tokens (Ant Design System Theme)
* **Product Name:** **AetherGrok** (*The Grok Build Desktop GUI Studio*)
* **Typography Hierarchy:**
  * UI Primary Sans-serif: `Geist Sans`, `Inter`, `-apple-system`, `BlinkMacSystemFont`, `Segoe UI`, `Roboto`, `sans-serif`
  * Monospace Code & CLI Stream: `JetBrains Mono`, `Fira Code`, `SF Mono`, `monospace`
* **Ant Design Color Palette (Dark High-Tech Studio):**
  * `antd-primary`: `#1677FF` (Ant Design Standard Blue) / Neon Highlight: `#00F0FF`
  * `antd-bg-base`: `#0F1117` (Deep Slate Slate Canvas)
  * `antd-bg-container`: `#181B26` (Panels, Sidebars, Cards)
  * `antd-bg-elevated`: `#222634` (Dropdowns, Floating Overlays, Modals)
  * `antd-border`: `#2E3446` (Crisp 1px borders)
  * `antd-text`: `#F3F4F6` (High-contrast text)
  * `antd-text-secondary`: `#8C93A4` (Muted labels & timestamps)
  * `antd-success`: `#52C41A` (Finished subagents, successful tool execution)
  * `antd-warning`: `#FAAD14` (Permission approval required)
  * `antd-error`: `#FF4D4F` (Execution error, failed turn)

---

## 2. Core Feature Specifications

### 1. Smart Non-Intrusive Screen Snapshot (App Window Exclusion)
* **Behavior:** When the snapshot button (or hotkey `Cmd/Ctrl + Shift + S`) is triggered:
  1. Instantly hide the Grok Desktop window (`Window.Hide()`).
  2. Sleep 50ms for display compositor flush.
  3. Execute platform-native capture on the active display (macOS `CGDisplayCreateImage`, Windows `BitBlt` / GDI).
  4. Restore and focus the Grok Desktop window (`Window.Show()`, `Window.Focus()`).
  5. Attach captured image to the current composer turn as vision context with instant preview.

### 2. Multi-Session Docking Workspace
* Tabbed multi-session manager with drag-and-drop ordering.
* Status indicator dots:
  * 🔵 Blue: Processing / Reasoning
  * 🟡 Yellow: Awaiting User Approval / Tool Permission
  * 🟢 Green: Finished turn
  * 🔴 Red: Failed turn
  * ⚪ Gray: Idle
* Instant fast-switching with zero-lag DOM flush and virtualized turn rendering.

### 3. Skills & MCP Registry Hub
* Automatic discovery of skills in `~/.grok/skills/`, `~/.agents/skills/`, and `.grok/workflows/`.
* Visual card catalog with instant search, category filters, and one-click injection into composer.

### 4. Settings & Theming System
* Full Ant Design dark and high-contrast theme toggling.
* Model selection (Grok 4.6, Grok Code, custom models).
* Reasoning effort slider (`none`, `low`, `medium`, `high`, `max`).
* Permission auto-approval toggle (Auto-accept, Agent, Plan mode).

---

## 3. Implementation Tasks & Verification

### Task 1: Initialize Ant Design Svelte Component & Token Library
- [ ] Setup `@ant-design/colors` token configuration in Tailwind and CSS custom variables.
- [ ] Implement Ant Design styled Svelte components: `Button`, `Card`, `Modal`, `Switch`, `Badge`, `Tabs`.
- [ ] Verify light/dark theme transition without layout shifting.

### Task 2: Implement Native Snapshot Engine with Window Auto-Hide
- [ ] Implement window state coordinator for snapshot trigger.
- [ ] Implement macOS CoreGraphics capture adapter.
- [ ] Implement Windows Win32 GDI capture adapter.
- [ ] Connect image compression and composer vision attachment.

### Task 3: Multi-Session Docking & Tab Management
- [ ] Create reactive Svelte 5 multi-session state manager.
- [ ] Build drag-and-drop tab bar with real-time status dots.
- [ ] Integrate session forking and history loading from `~/.grok/sessions/`.

### Task 4: Visual Tool Call Cards & Diff Inspector
- [ ] Build inline tool execution cards (`read_file`, `write`, `run_terminal_cmd`).
- [ ] Implement colorized file diff viewer (+green, -red) with side-by-side / unified toggles.
- [ ] Implement Permission approval cards (Allow once, Always allow, Reject).

### Task 5: Skills & MCP Explorer Modal
- [ ] Implement filesystem scanner for `SKILL.md` files.
- [ ] Build visual catalog modal with category filters and search.
- [ ] Add one-click skill parameter form and prompt generator.

### Task 6: Packaging, Performance Profiling & Cross-Platform Verification
- [ ] Verify idle RAM consumption under 60MB.
- [ ] Verify 60FPS fluid UI interaction and smooth tab switching.
- [ ] Build macOS and Windows distribution binaries.
