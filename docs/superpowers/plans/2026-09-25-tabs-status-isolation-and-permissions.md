# Implementation Plan: Settings Borders Refinement, Shortcut Recorder Keystroke Fix, macOS Accessibility Permissions, Tab Status Parity, Cmd+1-9 Switching, and Per-Session Prompt Isolation

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Polish Settings modal borders, resolve keystroke capture in shortcut recorder, request macOS global keyboard/accessibility permissions on app launch, align tab status icons with the sidebar, support `Cmd/Ctrl + 1-9` tab switching, and isolate prompt composer text/attachments per session tab.

**Spec:** `docs/superpowers/specs/2026-09-25-settings-recorder-permissions-and-tab-isolation.md`

---

## Completed Tasks

- [x] **Task 1: Settings Modal Borderless Refinement**
  - [x] Remove harsh border on `AetherGrok Studio` header tag in `SettingsModal.svelte`.
  - [x] Remove border on shortcut badge chips (`CmdOrCtrl`, `Shift`, `S`).
  - [x] Remove harsh border on `Record / Change Shortcut` and `Reset Default` buttons.
  - [x] Remove borders on About tab cards (dedication card, GitHub repo card, Portfolio card, Core capabilities card).
  - [x] Update developer author name to "Wiji Fiko Teren".
  - [x] Redesign keyboard shortcut reference table using individual `<kbd>` key badges with platform detection.

- [x] **Task 2: Fix Live Keystroke Recorder**
  - [x] Update `KeyRecorderModal.svelte` to listen on `window` with `capture: true` when opened.
  - [x] Remove modal container `stopPropagation` on keydown that swallowed keyboard events.
  - [x] Ensure support for modifier keys (`ShiftLeft`, `ShiftRight`, `MetaLeft`, `MetaRight`, `ControlLeft`, `ControlRight`, `AltLeft`, `AltRight`) and single keys (`/`, `Delete`, `Escape`).
  - [x] Clean up card borders and styling in accordance with the Anthropic Serif design system.

- [x] **Task 3: macOS Accessibility & Global Keyboard Permissions Integration**
  - [x] Create `pkg/permissions/permissions.go`, `permissions_darwin.go` (Objective-C CGO `AXIsProcessTrustedWithOptions`), and `permissions_other.go` (no-op).
  - [x] Expose `CheckAndRequestAccessibilityPermissions` and `OpenAccessibilitySettings` to Go `App` struct.
  - [x] Add TypeScript signatures in `frontend/src/app.d.ts`.
  - [x] Invoke permission verification on startup in `App.svelte` `onMount`.

- [x] **Task 4: Tab Status Icons & Browser-Style Quick Switching**
  - [x] Replace colored dot indicators in `SessionTabs.svelte` with `Loader2` (working), `AlertCircle` (waiting permission), `CheckCircle2` (completed), and `MessageSquare` (idle).
  - [x] Implement `⌘/Ctrl + 1-8` (switch to tab 1-8) and `⌘/Ctrl + 9` (switch to last tab) in `App.svelte`.
  - [x] Update `SettingsModal.svelte` shortcut reference.

- [x] **Task 5: Per-Session Prompt Isolation**
  - [x] Add `draft?: SessionDraft` to `Session` interface in `frontend/src/lib/stores/session.svelte.ts`.
  - [x] Implement reactive draft persistence and switching in `frontend/src/lib/components/chat/Composer.svelte`.
  - [x] Ensure clearing of session draft when prompt is sent.
  - [x] Verify lazy-mounted message feeds keep memory usage low.

- [x] **Task 6: Verification & Git Push**
  - [x] Run `svelte-check` (0 errors).
  - [x] Run `vite build` (successful compilation).
  - [x] Run `go test ./...` (all unit tests passing).
  - [x] Stage all changes cleanly and push to GitHub `main`.
