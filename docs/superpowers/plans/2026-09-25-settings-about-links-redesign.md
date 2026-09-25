# Settings Modal About Section & External Browser Links Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:**
1. Fix external links across the app so clicking GitHub repo (`https://github.com/fiko942/grok-build`), developer portfolio (`https://wijifikoteren.streampeg.com`), or header badge ("AetherGrok Studio") opens in the user's default system browser via Wails `BrowserOpenURL`.
2. Clean up and humanize the "About" tab in `SettingsModal.svelte`:
   - Replace cold, corporate AI-slop copy with warm, authentic, developer-crafted language ("Made with love by Wiji Fiko Teren").
   - Remove bold clutter (`**Bold**`) and artificial marketing triads.
   - Make the header pill and bottom-left "AetherGrok Studio" badge clickable links directly opening the repository.
   - Design clean, premium, tactile cards without robotic layout tropes.

**Tech Stack:**
- Go 1.24 backend (`wailsRuntime.BrowserOpenURL`)
- Svelte 5 with Runes (`$state`, `$derived`, `$props`)
- Tailwind CSS 3.4 & Ant Design Dark Tokens

---

### Task 1: Expose `OpenURL` method in Go Backend (`app.go`) & `frontend/src/app.d.ts`

**Files:**
- Modify: `app.go`
- Modify: `frontend/src/app.d.ts`

- [ ] **Step 1: Add `OpenURL(url string) error` to `App` struct in `app.go`**
Calls `wailsRuntime.BrowserOpenURL(a.ctx, url)` or falls back to system command `open` / `xdg-open` / `rundll32` if context is detached.
- [ ] **Step 2: Update TypeScript definitions in `frontend/src/app.d.ts`**
Add `OpenURL: (url string) => Promise<void>` and `BrowserOpenURL?: (url string) => void` to `Window.runtime`.

---

### Task 2: Implement Reusable Link Opener & Redesign About Section in `SettingsModal.svelte`

**Files:**
- Modify: `frontend/src/lib/components/layout/SettingsModal.svelte`

- [ ] **Step 1: Add `openExternalUrl(url: string)` helper in `SettingsModal.svelte`**
Handles clicking any external link using `window.go.main.App.OpenURL(url)` or `window.runtime.BrowserOpenURL(url)` with fallback to `window.open(url, '_blank')`.
- [ ] **Step 2: Make "AetherGrok Studio" badges clickable**
Both the header pill badge and the sidebar bottom branding now trigger opening `https://github.com/fiko942/grok-build` with hover states and cursor pointers.
- [ ] **Step 3: Redesign About Tab Copy & Visuals**
- Replace marketing tropes with authentic human phrasing:
  - Header: *"Handcrafted with focus by Wiji Fiko Teren"*
  - Vision: Crafted for developers who want a responsive, local-first GUI with native desktop shortcuts and transparent token control.
  - Links: Clickable cards for GitHub repo and Developer Portfolio (`https://wijifikoteren.streampeg.com`).
  - Tech footnote: Subtle monospace metadata.

---

### Task 3: Build & Verification

- [ ] **Step 1: Run Vite build (`npm run build`) in `frontend`**
- [ ] **Step 2: Run Wails production build (`wails build -skipbindings`)**
- [ ] **Step 3: Verify link opening behavior and visual aesthetics**
