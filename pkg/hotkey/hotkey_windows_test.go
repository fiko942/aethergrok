//go:build windows

package hotkey

import (
	"testing"
	"time"
	"unsafe"
)

func TestParseWindowsShortcutVariants(t *testing.T) {
	tests := []struct {
		input       string
		expectedMod uint32
		expectedVK  uint32
		wantErr     bool
	}{
		{"CmdOrCtrl+Shift+S", modControl | modShift, 'S', false},
		{"Ctrl+Shift+S", modControl | modShift, 'S', false},
		{"Cmd+D", modControl, 'D', false},
		{"ShiftRight", 0, 0xA1, false},
		{"Right Shift", 0, 0xA1, false},
		{"ShiftLeft", 0, 0xA0, false},
		{"\\", 0, 0xDC, false},
		{"/", 0, 0xBF, false},
		{"=", 0, 0xBB, false},
		{"-", 0, 0xBD, false},
		{"`", 0, 0xC0, false},
		{"[", 0, 0xDB, false},
		{"]", 0, 0xDD, false},
		{";", 0, 0xBA, false},
		{"'", 0, 0xDE, false},
		{",", 0, 0xBC, false},
		{".", 0, 0xBE, false},
		{"Alt+Space", modAlt, 0x20, false},
		{"Ctrl+Alt+S", modControl | modAlt, 'S', false},
		{"Shift+Alt+D", modShift | modAlt, 'D', false},
		{"F1", 0, 0x70, false},
		{"F12", 0, 0x7B, false},
		{"Space", 0, 0x20, false},
		{"Enter", 0, 0x0D, false},
		{"Tab", 0, 0x09, false},
	}

	for _, tt := range tests {
		mod, vk, err := parseWindowsShortcut(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("parseWindowsShortcut(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if mod != tt.expectedMod || vk != tt.expectedVK {
			t.Errorf("parseWindowsShortcut(%q) = (mod:%d, vk:0x%X), want (mod:%d, vk:0x%X)",
				tt.input, mod, vk, tt.expectedMod, tt.expectedVK)
		}
	}
}

func TestWindowsHotkeyManagerLifecycle(t *testing.T) {
	mgr := newPlatformManager()
	triggered := false

	err := mgr.start("CmdOrCtrl+Shift+S", func() {
		triggered = true
	})
	if err != nil {
		t.Fatalf("failed to start hotkey manager: %v", err)
	}

	err = mgr.update("Ctrl+Alt+A")
	if err != nil {
		t.Errorf("failed to update hotkey: %v", err)
	}

	mgr.stop()
	_ = triggered
}

func TestWindowsKeyEventDispatch(t *testing.T) {
	mgr := newPlatformManager()
	var actions []string
	triggered := false

	err := mgr.startWithKeyHandler("=", func() {
		triggered = true
	}, func(action string) {
		actions = append(actions, action)
	})
	if err != nil {
		t.Fatalf("failed to start: %v", err)
	}
	defer mgr.stop()

	// Simulate KeyDown for '=' (VK_OEM_PLUS = 0xBB)
	kbdDown := kbdLLHookStruct{vkCode: 0xBB}
	lowLevelKeyboardProc(0, wmKeyDown, uintptr(unsafe.Pointer(&kbdDown)))

	// Simulate KeyUp for '='
	kbdUp := kbdLLHookStruct{vkCode: 0xBB}
	lowLevelKeyboardProc(0, wmKeyUp, uintptr(unsafe.Pointer(&kbdUp)))

	time.Sleep(50 * time.Millisecond)

	if !triggered {
		t.Errorf("expected trigger for '=' hotkey")
	}
	if len(actions) != 2 {
		t.Fatalf("expected 2 actions (down and up), got %d: %v", len(actions), actions)
	}
	if actions[0] != "down" || actions[1] != "up" {
		t.Errorf("expected [down, up], got %v", actions)
	}
}

func TestConcurrentDualHotkeyManagers(t *testing.T) {
	snapMgr := NewManager()
	dictMgr := NewManager()

	snapTriggered := false
	var dictActions []string

	snapMgr.SetHandler(func() {
		snapTriggered = true
	})

	dictMgr.SetKeyHandler(func(action string) {
		dictActions = append(dictActions, action)
	})

	err := snapMgr.RegisterShortcut("Ctrl+Shift+S")
	if err != nil {
		t.Fatalf("failed to register snapshot hotkey: %v", err)
	}

	err = dictMgr.RegisterShortcut("ShiftRight")
	if err != nil {
		t.Fatalf("failed to register dictation hotkey: %v", err)
	}

	// Verify both managers are concurrently active in registry
	hookRegistryMu.RLock()
	activeCount := len(hookRegistry)
	isRunning := globalRunning
	hookRegistryMu.RUnlock()

	if activeCount != 2 {
		t.Errorf("expected 2 active managers, got %d", activeCount)
	}
	if !isRunning {
		t.Errorf("expected global hook to be running")
	}

	// Update dictation shortcut to backslash
	err = dictMgr.RegisterShortcut("\\")
	if err != nil {
		t.Errorf("failed to update dictation hotkey: %v", err)
	}

	// Stop one manager and verify the other is still active
	snapMgr.Unregister()

	hookRegistryMu.RLock()
	activeCount = len(hookRegistry)
	isRunning = globalRunning
	hookRegistryMu.RUnlock()

	if activeCount != 1 {
		t.Errorf("expected 1 active manager remaining, got %d", activeCount)
	}
	if !isRunning {
		t.Errorf("expected global hook to still be running with 1 manager")
	}

	// Stop remaining manager and verify hook shuts down
	dictMgr.Unregister()

	hookRegistryMu.RLock()
	activeCount = len(hookRegistry)
	hookRegistryMu.RUnlock()

	if activeCount != 0 {
		t.Errorf("expected 0 active managers, got %d", activeCount)
	}

	_ = snapTriggered
	_ = dictActions
}
