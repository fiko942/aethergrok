package system

import (
	"runtime"
	"testing"
)

func TestMuteAndRestoreSystemVolume(t *testing.T) {
	state, err := MuteSystemVolume()
	if err != nil {
		t.Fatalf("MuteSystemVolume failed: %v", err)
	}

	if runtime.GOOS == "darwin" {
		if state.OriginalVolume < 0 || state.OriginalVolume > 100 {
			t.Errorf("expected original volume between 0 and 100, got %d", state.OriginalVolume)
		}
	} else if runtime.GOOS == "windows" {
		if state.OriginalVolume < 0 || state.OriginalVolume > 100 {
			t.Errorf("expected original volume between 0 and 100 on windows, got %d", state.OriginalVolume)
		}
	} else {
		if state.OriginalVolume != 100 || state.WasMuted != false {
			t.Errorf("expected default 100/false on other OS, got %d/%v", state.OriginalVolume, state.WasMuted)
		}
	}

	// Restore original state
	if err := RestoreSystemVolume(state); err != nil {
		t.Fatalf("RestoreSystemVolume failed: %v", err)
	}
}
