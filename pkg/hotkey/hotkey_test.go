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

func TestDarwinWatchdogAutoRelease(t *testing.T) {
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

	// In automated tests, CGEventSourceKeyState for keycode 42 returns false (not physically held).
	// The watchdog ticker runs every 60ms and detects that the physical key is NOT held down,
	// automatically dispatching an "up" event!
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
		t.Fatalf("expected watchdog to emit 'up' action when physical key is not held, got %d: %v", len(actions), actions)
	}
	if actions[0] != "down" || actions[1] != "up" {
		t.Errorf("expected [down, up], got %v", actions)
	}
}
