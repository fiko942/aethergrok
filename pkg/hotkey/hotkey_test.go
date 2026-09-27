//go:build darwin

package hotkey

import (
	"testing"
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

	// Test Combo CmdOrCtrl+Shift+S
	kc, isMod, flags, err = parseShortcutDarwin("CmdOrCtrl+Shift+S")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isMod != 0 || kc != 1 { // 's' is keycode 1 on mac
		t.Errorf("expected keycode 1 ('s') and isMod 0, got kc=%d isMod=%d", kc, isMod)
	}
	if flags&(0x100000|0x020000) != (0x100000 | 0x020000) {
		t.Errorf("expected cmd and shift flags set, got %x", flags)
	}
}
