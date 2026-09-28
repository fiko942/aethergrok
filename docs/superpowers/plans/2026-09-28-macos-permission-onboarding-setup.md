# macOS First-Launch Permission Onboarding Setup Wizard Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Provide an Anti-Gravity style, theme-aware startup onboarding permission wizard on macOS that checks all 3 critical system permissions (Accessibility, Microphone, Screen Recording), presents an interactive gate with real-time status and 1-click system settings triggers, and smoothly transitions into the workspace once permissions are satisfied.

**Architecture:** 
1. **Backend (Go / macOS Cgo)**: Extend `pkg/permissions` with `CheckScreenCapturePermission()`, `RequestScreenCapturePermission()`, and `OpenScreenCapturePreferences()`. Provide a consolidated `CheckAllSystemPermissions()` endpoint returning detailed status for Accessibility, Microphone, and Screen Recording on macOS (and auto-granted fallback for other OSs).
2. **Frontend (Svelte 5 Runes & Ant Design Design System)**: Add `PermissionSetupModal.svelte` onboarding gate in `App.svelte` that intercepts app startup on macOS, polls or re-checks permission status on window focus / button click, and seamlessly unblocks the main UI once all required permissions are granted.

**Tech Stack:** Go (Cgo, ApplicationServices, AVFoundation, CoreGraphics), Wails v2, Svelte 5 (Runes: `$state`, `$derived`, `$effect`), Tailwind CSS, Ant Design CSS Design Tokens.

## Global Constraints
- Target platform gate: macOS only (`runtime.GOOS == "darwin"`); non-macOS platforms immediately return all permissions granted.
- Use explicit macOS deep-links (`open x-apple.systempreferences:com.apple.preference.security?Privacy_...`) to avoid triggering unnecessary osascript AppleEvents permissions.
- Theme compatibility: Strict compliance with AetherGrok CSS tokens (`data-theme`: `dark-studio`, `dark-high-contrast`, `light-antd`).
- Non-destructive & Smooth: Do not corrupt existing workspace state; auto-resume once permissions are satisfied.

---

### Task 1: Extend macOS Permissions Cgo/Backend for Screen Recording & Consolidated Status

**Files:**
- Create/Modify: `pkg/permissions/permissions_darwin.go`
- Create/Modify: `pkg/permissions/permissions_other.go`
- Create/Modify: `pkg/permissions/permissions.go`
- Modify: `app.go`

**Interfaces:**
- Consumes: macOS `CoreGraphics` framework (`CGPreflightScreenCaptureAccess`, `CGRequestScreenCaptureAccess`), `ApplicationServices`, `AVFoundation`.
- Produces: 
  - `CheckScreenCapturePermission() Status`
  - `RequestScreenCapturePermission() Status`
  - `OpenScreenCapturePreferences() error`
  - `CheckAllSystemPermissions() AllPermissionsStatus`

- [ ] **Step 1: Write backend permission types and functions in `pkg/permissions`**
Add `AllPermissionsStatus` struct:
```go
type SystemPermissionItem struct {
    ID          string `json:"id"`          // "accessibility", "microphone", "screen_capture"
    Title       string `json:"title"`
    Description string `json:"description"`
    Granted     bool   `json:"granted"`
    Message     string `json:"message"`
    Required    bool   `json:"required"`
}

type AllPermissionsStatus struct {
    Platform      string                 `json:"platform"`
    AllGranted    bool                   `json:"allGranted"`
    Items         []SystemPermissionItem `json:"items"`
}
```

- [ ] **Step 2: Implement Screen Capture Cgo and Preference opener in `permissions_darwin.go`**
```objective-c
#import <CoreGraphics/CoreGraphics.h>

static bool checkScreenCaptureAccess() {
    if (@available(macOS 10.15, *)) {
        return CGPreflightScreenCaptureAccess();
    }
    return true;
}

static bool requestScreenCaptureAccess() {
    if (@available(macOS 10.15, *)) {
        return CGRequestScreenCaptureAccess();
    }
    return true;
}
```
And add `OpenScreenCapturePreferences()`:
```go
func OpenScreenCapturePreferences() error {
    cmd := exec.Command("open", "x-apple.systempreferences:com.apple.preference.security?Privacy_ScreenCapture")
    return cmd.Run()
}
```

- [ ] **Step 3: Expose App bridge methods in `app.go`**
Expose `CheckAllSystemPermissions()`, `OpenScreenCaptureSettings()`, `RequestScreenCapturePermission()`.

- [ ] **Step 4: Verify Go build**
Run: `export PATH=$PATH:/usr/local/go/bin:~/go/bin:/opt/homebrew/bin && go build -o /dev/null .`
Expected: PASS

---

### Task 2: Create `PermissionSetupModal.svelte` (Anti-Gravity Style Onboarding Wizard)

**Files:**
- Create: `frontend/src/lib/components/setup/PermissionSetupModal.svelte`
- Modify: `frontend/src/app.d.ts`

**Interfaces:**
- Consumes: `window.go.main.App.CheckAllSystemPermissions`, `window.go.main.App.OpenAccessibilitySettings`, `window.go.main.App.OpenMicrophoneSettings`, `window.go.main.App.OpenScreenCaptureSettings`, `window.go.main.App.RequestMicrophonePermission`, `window.go.main.App.RequestScreenCapturePermission`.
- Produces: Svelte 5 component with real-time status indicators, auto-polling on window focus, and onComplete callback.

- [ ] **Step 1: Update TypeScript definitions in `frontend/src/app.d.ts`**
Add signatures for `CheckAllSystemPermissions`, `OpenScreenCaptureSettings`, and `RequestScreenCapturePermission`.

- [ ] **Step 2: Build `PermissionSetupModal.svelte`**
- Header: Shield icon with pulsing status, "macOS System Permissions Setup" title, subtitle explaining why AetherGrok requires these permissions.
- Progress bar: e.g. "2 of 3 permissions granted" with smooth progress fill.
- Item Cards:
  1. **Accessibility (Global Shortcuts)**: Key icon, status badge (Granted vs Required), "Open System Settings" button.
  2. **Microphone (Voice Dictation)**: Mic icon, status badge, "Allow / Open Settings" button.
  3. **Screen Recording (Smart Snapshot & Vision)**: Monitor icon, status badge, "Allow / Open Settings" button.
- Footer Actions:
  - "Re-check Status" button (with spin animation).
  - "Continue to AetherGrok" button (enabled automatically when `allGranted` is true, or allow optional skip if non-critical with clear warning).

- [ ] **Step 3: Add Window Focus listener to auto-refresh status**
When user switches back from macOS System Settings to AetherGrok, the modal automatically runs `checkAll()` to update badges instantly.

---

### Task 3: Integrate Permission Gate into `App.svelte` on Startup

**Files:**
- Modify: `frontend/src/App.svelte`

- [ ] **Step 1: Check permissions on app mount in `App.svelte`**
In `onMount`:
Call `CheckAllSystemPermissions()`. If `!result.allGranted` and `result.platform === 'darwin'`, set `showPermissionSetup = true`.

- [ ] **Step 2: Render `PermissionSetupModal` when `showPermissionSetup` is true**
Render on top of the UI as a backdrop modal dialog matching the active theme (`dark-studio`, `dark-high-contrast`, `light-antd`).

- [ ] **Step 3: Verify with Vitest and frontend build**
Run: `cd frontend && npx vitest run && npm run build`
Expected: PASS

---

### Task 4: End-to-End Verification & Manual Sanity Check

- [ ] **Step 1: Verify Theme switching compatibility**
Ensure colors adapt seamlessly across all 3 themes (`dark-studio`, `dark-high-contrast`, `light-antd`).
- [ ] **Step 2: Verify Go build and packaging**
Ensure binary compiles cleanly with no runtime errors.
