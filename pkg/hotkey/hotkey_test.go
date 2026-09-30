//go:build darwin

package hotkey

import (
	"sync"
	"testing"
	"time"
)

func TestParseShortcutDarwin(t *testing.T) {
	// Test Right Shift standalone modifier
	kc, isMod, flags, err := parseShortcutDarwin("ShiftRight")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if kc != kVK_RightShift || isMod != 1 {
		t.Errorf("expected right shift keycode %d and isMod 1, got kc=%d isMod=%d", kVK_RightShift, kc, isMod)
	}

	// Test Right Shift with spaces/casing
	kc, isMod, _, err = parseShortcutDarwin("Right Shift")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if kc != kVK_RightShift || isMod != 1 {
		t.Errorf("expected right shift keycode %d, got kc=%d", kVK_RightShift, kc)
	}

	// Test Right Option standalone modifier
	kc, isMod, _, err = parseShortcutDarwin("Right Option")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if kc != kVK_RightOption || isMod != 1 {
		t.Errorf("expected right option keycode %d, got kc=%d", kVK_RightOption, kc)
	}

	// Test Backslash single key
	kc, isMod, flags, err = parseShortcutDarwin("\\")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if kc != 42 || isMod != 0 || flags != 0 {
		t.Errorf("expected backslash keycode 42, isMod 0, flags 0; got kc=%d isMod=%d flags=%x", kc, isMod, flags)
	}

	// Test Combo CmdOrCtrl+Shift+S
	kc, isMod, flags, err = parseShortcutDarwin("CmdOrCtrl+Shift+S")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isMod != 0 || kc != 1 { // 's' is keycode 1 on mac
		t.Errorf("expected keycode 1 ('s') and isMod 0, got kc=%d isMod=%d", kc, isMod)
	}
	if flags&(cgEventFlagMaskCommand|cgEventFlagMaskShift) != (cgEventFlagMaskCommand | cgEventFlagMaskShift) {
		t.Errorf("expected cmd and shift flags set, got %x", flags)
	}
}

func TestDarwinKeyEventDispatch(t *testing.T) {
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

	// Simulate KeyDown for '\' (keycode 42)
	triggerDarwinKeyEvent(42, cgEventKeyDown, 0)
	// Simulate KeyUp for '\' (keycode 42)
	triggerDarwinKeyEvent(42, cgEventKeyUp, 0)

	// Allow goroutine event channel dispatch
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
		t.Fatalf("expected 2 actions (down and up), got %d: %v", len(actions), actions)
	}
	if actions[0] != "down" || actions[1] != "up" {
		t.Errorf("expected [down, up], got %v", actions)
	}
}

func TestDarwinModifierWatchdogAutoRelease(t *testing.T) {
	mgr := newPlatformManager()
	var actions []string
	var mu sync.Mutex

	// Test with a standalone modifier key (ShiftRight) where CGEventSourceKeyState is verified
	err := mgr.startWithKeyHandler("ShiftRight", func() {}, func(action string) {
		mu.Lock()
		actions = append(actions, action)
		mu.Unlock()
	})
	if err != nil {
		t.Fatalf("failed to start: %v", err)
	}
	defer mgr.stop()

	// Simulate FlagsChanged KeyDown for Right Shift (keycode 60, flags with Shift)
	triggerDarwinKeyEvent(kVK_RightShift, cgEventFlagsChanged, cgEventFlagMaskShift)

	// In automated tests without physically holding Right Shift, the modifier watchdog
	// detects that the key is not physically held, dispatching an "up" event.
	for i := 0; i < 50; i++ {
		mu.Lock()
		count := len(actions)
		mu.Unlock()
		if count >= 2 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(actions) != 2 {
		t.Fatalf("expected modifier watchdog to emit 'up' action when physical modifier is not held, got %d: %v", len(actions), actions)
	}
	if actions[0] != "down" || actions[1] != "up" {
		t.Errorf("expected [down, up], got %v", actions)
	}
}

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
	if actions[0] != "down" || actions[1] != "up" {
		t.Errorf("expected [down, up], got %v", actions)
	}
}

func TestCarbonHotKeyDispatch(t *testing.T) {
	mgr := newPlatformManager().(*darwinHotkeyManager)
	var actions []string
	var mu sync.Mutex

	err := mgr.startWithKeyHandler("Cmd+Shift+S", func() {
		mu.Lock()
		actions = append(actions, "trigger")
		mu.Unlock()
	}, func(action string) {
		mu.Lock()
		actions = append(actions, action)
		mu.Unlock()
	})
	if err != nil {
		t.Fatalf("failed to start: %v", err)
	}
	defer mgr.stop()

	mgr.mu.Lock()
	carbonID := mgr.carbonID
	mgr.mu.Unlock()

	if carbonID == 0 {
		t.Fatalf("expected non-zero carbonID for Cmd+Shift+S")
	}

	// Dispatch Carbon hotkey down event
	dispatchCarbonHotKey(carbonID, true)
	// Dispatch Carbon hotkey up event
	dispatchCarbonHotKey(carbonID, false)

	for i := 0; i < 50; i++ {
		mu.Lock()
		count := len(actions)
		mu.Unlock()
		if count >= 3 { // down, trigger, up
			break
		}
		time.Sleep(5 * time.Millisecond)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(actions) < 2 {
		t.Fatalf("expected at least down and up actions, got: %v", actions)
	}
}
