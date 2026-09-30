# Voice Dictation macOS Global Event Trigger Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ensure macOS global event triggers for Voice Dictation (push-to-talk hold and double-tap hands-free lock) trigger reliably in all states, including background/unfocused windows and empty session states.

**Architecture:** 
1. Backend (`pkg/hotkey/hotkey_darwin.go`): Ensure robust CGEventTap lifecycle, modifier mask isolation (masking non-standard hardware/device flags `0x100` / `0x200` / `0x800000` so comparisons succeed cleanly), automatic recovery when event taps get disabled by macOS timeout, and reliable run loop source registration on the dedicated OS thread.
2. Frontend (`frontend/src/App.svelte` & `frontend/src/lib/components/chat/Composer.svelte`): Auto-create an active session if none exists when a global dictation trigger occurs so Composer mounts and receives the event, and synchronize the shortcut detector state between DOM keyboard events and Wails native global hook events.

**Tech Stack:** Go (cgo, ApplicationServices, CoreFoundation, Carbon keycodes), Svelte 5 / TypeScript, Wails v2.

---

### Task 1: Refine macOS CGEventTap Flag Masking & RunLoop Lifecycle in `pkg/hotkey/hotkey_darwin.go`

**Files:**
- Modify: `pkg/hotkey/hotkey_darwin.go`
- Test: `pkg/hotkey/hotkey_test.go`

- [x] **Step 1: Write the failing / unit test for flag masking with hardware/device bits**

```go
func TestDarwinKeyEventDispatch_WithHardwareModifierNoise(t *testing.T) {
	mgr := newPlatformManager()
	var actions []string
	var mu sync.Mutex

	err := mgr.startWithKeyHandler("\\", func() {}, func(action string) {
		mu.Lock()
		actions = append(actions, action)
		mu.Unlock()
	})
	if err != nil {
		t.Fatalf("failed to start: %v", err)
	}
	defer mgr.stop()

	// Simulate KeyDown with extra hardware flag bits (e.g. 0x20000000 / NonCoalesced 0x100)
	noiseFlags := uint64(0x20000100)
	triggerDarwinKeyEvent(42, cgEventKeyDown, noiseFlags)
	triggerDarwinKeyEvent(42, cgEventKeyUp, noiseFlags)

	for i := 0; i < 50; i++ {
		mu.Lock()
		count := len(actions)
		mu.Unlock()
		if count >= 2 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(actions) != 2 {
		t.Fatalf("expected 2 actions with noisy hardware flags, got %d: %v", len(actions), actions)
	}
}
```

- [x] **Step 2: Run test to verify behavior**

Run: `go test -v ./pkg/hotkey/...`

- [x] **Step 3: Update `pkg/hotkey/hotkey_darwin.go` modifier filtering and non-modifier watchdog handling**

In `pkg/hotkey/hotkey_darwin.go`:
1. Mask `flags` against standard modifier masks (`cgEventFlagMaskCommand | cgEventFlagMaskShift | cgEventFlagMaskAlternate | cgEventFlagMaskControl`) when matching shortcut modifier requirements.
2. For standalone modifier keys (`ShiftRight`), use the 60ms ticker polling `CGEventSourceKeyState`.
3. For regular character/symbol keys (e.g. `\`), avoid polling `CGEventSourceKeyState` in the background (which returns 0 when another app is focused) and rely on event-driven `CGEventTap` / Carbon release handlers with a generous 10-minute safety timeout to support continuous long push-to-talk speech without cutting off at 60s.

- [x] **Step 4: Run tests to verify they pass**

Run: `go test -v ./pkg/hotkey/...`

---

### Task 2: Ensure Session Auto-Creation on Global Dictation Trigger in `App.svelte`

**Files:**
- Modify: `frontend/src/App.svelte`

- [x] **Step 1: Update Wails `dictation:trigger_global` handler in `frontend/src/App.svelte`**

When `dictation:trigger_global` is received with `down` or `hold-start` / `double-tap-lock`:
If `!sessionStore.activeSession && sessionStore.activeWorkspace`, automatically create a new session (`sessionStore.createSession(sessionStore.activeWorkspace.id, 'Voice Dictation')`) so that `<Composer>` is immediately mounted and ready to record audio. Increased inputShield lock duration to 10 minutes to match push-to-talk hold duration.

- [x] **Step 2: Verify frontend builds and vitest tests pass**

Run: `pnpm --dir frontend test` and `pnpm --dir frontend build`

---

### Task 3: Full End-to-End Verification

- [x] **Step 1: Run all Go unit tests**

Run: `go test -v ./pkg/...`

- [x] **Step 2: Run frontend tests & build**

Run: `pnpm --dir frontend test` and `pnpm --dir frontend build`

---

### Task 4: Packaging and Release

- [x] **Step 1: Build macOS DMG package (`AetherGrok-1.1.1-macOS-arm64.dmg`)**
- [ ] **Step 2: Stage, commit, and push changes to GitHub `main`**
- [ ] **Step 3: Update GitHub Release v1.1.1 with clean macOS DMG asset and checksum**
