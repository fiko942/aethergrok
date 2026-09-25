# Settings Modal "About" Section Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a dedicated **About** tab at the bottom of the navigation in `SettingsModal.svelte`. Describe the purpose, audience, and vision of AetherGrok, showcase that it is an open-source project with GitHub repository links, and feature the developer's portfolio link (`https://wijifikoteren.streampeg.com`).

**Architecture & Design:**
- In `SettingsModal.svelte`:
  - Extend `TabKey = 'general' | 'models' | 'permissions' | 'theme' | 'shortcuts' | 'about'`
  - Add `{ id: 'about', label: 'About AetherGrok', icon: Info, description: 'Project background, mission, open-source repository, and developer portfolio' }` to `tabs`.
  - Design an editorial, high-taste About page using Anthropic Serif typography, dark subtle cards (`border-white/5`), clean badges, and external links with `target="_blank"`.
- Content to include:
  1. **Vision & Purpose**: Why AetherGrok was built (A modern, fast, local-first GUI engine for Grok CLI with native macOS/Windows screen snapshot, lightweight DOM memory windowing, and Anthropic-grade typography).
  2. **Who It Is For**: AI engineers, developers, autonomous agent builders, and developers who desire a clean, high-performance desktop companion without web browser bloat.
  3. **Open-Source Repository**: Link to GitHub (`https://github.com/fiko942/grok-build`).
  4. **Creator & Developer Portfolio**: Link to `https://wijifikoteren.streampeg.com` with developer credits.
  5. **Technology Stack**: Wails v2, Go 1.24, Svelte 5 (Runes), Tailwind CSS, Ant Design Dark Tokens.

---

### Task 1: Update SettingsModal Tab Types & Navigation

**Files:**
- Modify: `frontend/src/lib/components/layout/SettingsModal.svelte`

- [ ] **Step 1: Add `'about'` to `TabKey` and `tabs` array**
Import `Info` or `Compass` / `Globe` / `Github` from `lucide-svelte`.
- [ ] **Step 2: Add About Tab View in tab content viewport**
Render cards for "Project Mission & Vision", "Target Audience", "Open-Source Repository", and "Developer Portfolio & Credits".

---

### Task 2: Build & Verification

**Files:**
- Output: `build/bin/aethergrok.app`

- [ ] **Step 1: Run `npm run build` and `wails build -clean`**
- [ ] **Step 2: Launch AetherGrok, open Settings -> About, and verify links and layout**
