//go:build !darwin && !windows

package permissions

// GetAudioInputDevices returns an empty slice on non-darwin platforms as fallback
func GetAudioInputDevices() ([]AudioDeviceInfo, error) {
	return []AudioDeviceInfo{}, nil
}
