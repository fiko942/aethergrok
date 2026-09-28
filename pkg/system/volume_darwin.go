//go:build darwin

package system

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// parseDarwinVolumeSettings parses the comma-separated output from `osascript -e "get volume settings"`.
// Example output: "output volume:25, input volume:34, alert volume:100, output muted:false"
func parseDarwinVolumeSettings(output string) (int, bool, error) {
	parts := strings.Split(output, ",")
	foundVol := false
	foundMuted := false
	vol := 0
	muted := false

	for _, part := range parts {
		kv := strings.SplitN(strings.TrimSpace(part), ":", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		val := strings.TrimSpace(kv[1])

		if key == "output volume" {
			v, err := strconv.Atoi(val)
			if err != nil {
				return 0, false, fmt.Errorf("failed to parse output volume %q: %w", val, err)
			}
			vol = v
			foundVol = true
		} else if key == "output muted" {
			m, err := strconv.ParseBool(val)
			if err != nil {
				return 0, false, fmt.Errorf("failed to parse output muted %q: %w", val, err)
			}
			muted = m
			foundMuted = true
		}
	}

	if !foundVol || !foundMuted {
		return 0, false, fmt.Errorf("invalid volume settings output: %q", output)
	}

	return vol, muted, nil
}

// MuteSystemVolume queries the current macOS output volume & mute status, then mutes the output.
func MuteSystemVolume() (SystemVolumeState, error) {
	// Query current volume settings
	out, err := exec.Command("osascript", "-e", "get volume settings").Output()
	if err != nil {
		return SystemVolumeState{}, fmt.Errorf("failed to get volume settings: %w", err)
	}

	vol, muted, err := parseDarwinVolumeSettings(string(out))
	if err != nil {
		return SystemVolumeState{}, err
	}

	state := SystemVolumeState{
		OriginalVolume: vol,
		WasMuted:       muted,
	}

	// Mute output volume
	muteScript := "set volume output muted true"
	if err := exec.Command("osascript", "-e", muteScript).Run(); err != nil {
		return state, fmt.Errorf("failed to mute volume: %w", err)
	}

	return state, nil
}

// RestoreSystemVolume restores the volume level and mute state.
func RestoreSystemVolume(state SystemVolumeState) error {
	mutedStr := "false"
	if state.WasMuted {
		mutedStr = "true"
	}

	script := fmt.Sprintf("set volume output volume %d output muted %s", state.OriginalVolume, mutedStr)
	if err := exec.Command("osascript", "-e", script).Run(); err != nil {
		return fmt.Errorf("failed to restore volume: %w", err)
	}

	return nil
}
