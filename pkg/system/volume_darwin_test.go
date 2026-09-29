//go:build darwin

package system

import (
	"testing"
)

func TestParseDarwinVolumeSettings(t *testing.T) {
	tests := []struct {
		name       string
		output     string
		wantVolume int
		wantMuted  bool
		wantErr    bool
	}{
		{
			name:       "standard output not muted",
			output:     "output volume:25, input volume:34, alert volume:100, output muted:false\n",
			wantVolume: 25,
			wantMuted:  false,
			wantErr:    false,
		},
		{
			name:       "standard output muted",
			output:     "output volume:80, input volume:50, alert volume:100, output muted:true",
			wantVolume: 80,
			wantMuted:  true,
			wantErr:    false,
		},
		{
			name:       "missing output volume key",
			output:     "input volume:34, alert volume:100, output muted:false",
			wantVolume: 0,
			wantMuted:  false,
			wantErr:    true,
		},
		{
			name:       "missing output muted key",
			output:     "output volume:50, input volume:34, alert volume:100",
			wantVolume: 0,
			wantMuted:  false,
			wantErr:    true,
		},
		{
			name:       "empty output",
			output:     "",
			wantVolume: 0,
			wantMuted:  false,
			wantErr:    true,
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
