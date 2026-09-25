# Superpowers Spec: Settings Borderless Refinement, Keystroke Capture, macOS Accessibility Permissions, Tab Status Icons, Cmd+1-9 Switching, and Per-Session Prompt Isolation

**Status:** Completed & Integrated  
**Date:** 2026-09-25  
**Version:** 1.0.0  
**Stack:** Wails v2 (Go 1.24/1.27 + CGO Objective-C) + Svelte 5 (Runes) + Tailwind CSS + Anthropic Serif Font Family

---

## 1. Context & Motivation
AetherGrok Desktop provides native speed and tactile responsiveness for autonomous Grok CLI workflows.
During continuous user testing and refinement, several critical visual, interaction, and system permission issues were identified and addressed:
1. **Settings Modal Harsh Borders**: Unwanted white and high-contrast borders were cluttering the UI in dark mode (header pills, shortcut key chips `<kbd>`, record button, reset button, and about cards).
2. **Keystroke Recording in Shortcut Recorder**: The shortcut recording modal was failing to capture keystrokes due to keyboard event interception and bubbling issues.
3. **macOS Global Keyboard & Accessibility Permissions**: Global keyboard shortcut listening and screen snapshot workflows on macOS require accessibility trust (`AXIsProcessTrustedWithOptions`). The app must verify and trigger the native system prompt on macOS startup.
4. **Interactive Shortcut Reference & Tab Polish**: Shortcut reference tables needed structured platform-aware key badges rather than raw slash-separated strings.
5. **Session Tab Status Consistency**: Tabs in the top navigation bar were using plain colored dots, which lacked parity with the expressive, animated status icons in the left sidebar.
6. **Browser-Style Quick Tab Navigation (`⌘/Ctrl + 1-9`)**: Power users expect standard browser key combinations to quickly switch between open session tabs.
7. **Per-Session Prompt Isolation & Memory Efficiency**: Prompt inputs (text draft, attached snapshot images, attached context files) must be cleanly isolated per tab so switching sessions preserves work-in-progress drafts without bleeding between tabs, while maintaining lazy-mounted message rendering.

---

## 2. Technical Architecture & Delivered Features

### 2.1 Settings Modal Borderless Refinement (`frontend/src/lib/components/layout/SettingsModal.svelte`)
- Removed harsh borders:
  - Header pill `AetherGrok Studio`: `border-0 bg-ant-primary/15 text-ant-primary`.
  - Shortcut key badges `<kbd>`: Borderless `bg-ant-primary/10 text-ant-primary`.
  - `Record / Change Shortcut` button: Borderless `bg-ant-primary/10 hover:bg-ant-primary/20 text-ant-primary`.
  - `Reset Default` button: Borderless `bg-white/[0.04] hover:bg-white/[0.08]`.
  - About section cards: Removed `border border-white/5` from dedication card, GitHub card, Portfolio card, and Core Capabilities card.
- Updated author dedication to **"Made with love by Wiji Fiko Teren"**.
- Restructured `keyboardShortcuts` reference table to render individual `<kbd>` key pills with platform modifiers (`⌘` on macOS, `Ctrl` on others) and added shortcuts `⌘/Ctrl + 1-8` and `⌘/Ctrl + 9`.

### 2.2 Live Keystroke Recorder Fix (`frontend/src/lib/components/ui/KeyRecorderModal.svelte`)
- Implemented `window.addEventListener('keydown', listener, true)` in the capture phase on modal mount/open.
- Removed blocking `stopPropagation` on the modal card that previously prevented keystrokes from firing when focused.
- Added support for standalone modifiers (`ShiftLeft`, `ShiftRight`, `MetaLeft`, `MetaRight`, `ControlLeft`, `ControlRight`, `AltLeft`, `AltRight`) and single key triggers (`/`, `Delete`, `Escape`, `Space`, function keys).
- Borderless Anthropic Serif design styling with live pulsing listening indicator.

### 2.3 macOS Native Accessibility & Keyboard Permission Verification (`pkg/permissions/`)
- Created native Go package `pkg/permissions`:
  - `permissions_darwin.go`: Utilizes Objective-C cgo bridging:
    - `AXIsProcessTrusted()` to inspect existing authorization.
    - `AXIsProcessTrustedWithOptions({kAXTrustedCheckOptionPrompt: true})` to prompt the macOS System Settings permission dialog.
    - Fallback helper `OpenAccessibilityPreferences()` via AppleScript / `x-apple.systempreferences`.
  - `permissions_other.go`: No-op returning `granted: true` for Windows and Linux.
- Exposed `CheckAndRequestAccessibilityPermissions` and `OpenAccessibilitySettings` to the Wails runtime.
- Automatically invoked in `App.svelte` `onMount` on app launch.

### 2.4 Expressive Status Icons on Session Tabs (`frontend/src/lib/components/layout/SessionTabs.svelte`)
- Replaced status dots with status icons matching the sidebar:
  - `working`: `<Loader2 size={12.5} class="animate-spin text-ant-primary" />`
  - `waiting_permission`: `<AlertCircle size={12.5} class="animate-bounce text-amber-400" />`
  - `finished`: `<CheckCircle2 size={12.5} class="text-ant-success/80" />`
  - `idle`: `<MessageSquare size={12.5} />` (styled `text-ant-primary` when active).

### 2.5 Quick Tab Switching (`Cmd/Ctrl + 1-9`) (`frontend/src/App.svelte`)
- Added global listener for `⌘/Ctrl + 1-9`:
  - `1-8`: Selects tab 1 through 8 in `sessionStore.openWorkspaceTabs`.
  - `9`: Jumps directly to the last open tab.

### 2.6 Per-Session Prompt Isolation (`frontend/src/lib/stores/session.svelte.ts` & `Composer.svelte`)
- Added `draft?: SessionDraft` to `Session` model:
  ```typescript
  export interface SessionDraft {
    text: string;
    images: VisionImage[];
    attachments: AttachedFile[];
  }
  ```
- In `Composer.svelte`:
  - A reactive watcher tracks `activeSessionId`.
  - When switching away from a session, uncommitted input text, attached snapshots, and file chips are stored in `prevSession.draft`.
  - When activating a new session, its draft is restored into the composer.
  - Submitting a prompt clears the draft for that session.
  - Lazy mounting in `MessageList.svelte` ensures only the active session's conversation turns are rendered in the DOM, keeping memory usage minimal while background sessions continue streaming.

---

## 3. Verification & Compliance
- **Svelte Check**: `0 errors, 4 warnings` (clean typecheck across all components).
- **Vite Production Build**: Successfully compiled (`dist/`).
- **Go Backend**: Successfully compiled with CGO on macOS; all package unit tests passing (`ok aethergrok/pkg/grokrunner`, `ok aethergrok/test`).
