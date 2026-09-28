package system

import (
	"runtime"
	"testing"
)

func TestParseDarwinVolumeSettings(t *testing.T) {
	tests := []struct {
		name        string
		output      string
		wantVolume  int
		wantMuted   bool
		wantErr     bool
	}{
		{
			name:        "standard output not muted",
			output:      "output volume:25, input volume:34, alert volume:100, output muted:false\n",
			wantVolume:  25,
			wantMuted:   false,
			wantErr:     false,
		},
		{
			name:        "standard output muted",
			output:      "output volume:80, input volume:50, alert volume:100, output muted:true",
			wantVolume:  80,
			wantMuted:   true,
			wantErr:     false,
		},
		{
			name:        "missing output volume key",
			output:      "input volume:34, alert volume:100, output muted:false",
			wantVolume:  0,
			wantMuted:   false,
			wantErr:     true,
		},
		{
			name:        "missing output muted key",
			output:      "output volume:50, input volume:34, alert volume:100",
			wantVolume:  0,
			wantMuted:   false,
			wantErr:     true,
		},
		{
			name:        "empty output",
			output:      "",
			wantVolume:  0,
			wantMuted:   false,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vol, muted, err := parseDarwinVolumeSettings(tt.output)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseDarwinVolumeSettings() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if vol != tt.wantVolume {
					t.Errorf("parseDarwinVolumeSettings() vol = %d, want %d", vol, tt.wantVolume)
				}
				if muted != tt.wantMuted {
					t.Errorf("parseDarwinVolumeSettings() muted = %v, want %v", muted, tt.wantMuted)
				}
			}
		})
	}
}

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
