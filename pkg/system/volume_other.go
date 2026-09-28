//go:build !darwin && !windows

package system

// MuteSystemVolume returns a graceful default state on unsupported platforms.
func MuteSystemVolume() (SystemVolumeState, error) {
	return SystemVolumeState{
		OriginalVolume: 100,
		WasMuted:       false,
	}, nil
}

// RestoreSystemVolume is a no-op on unsupported platforms.
func RestoreSystemVolume(state SystemVolumeState) error {
	return nil
}
