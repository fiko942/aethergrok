//go:build darwin

package permissions

import (
	"encoding/json"
	"os/exec"
	"strings"
)

type spAudioOutput struct {
	SPAudioDataType []struct {
		Items []struct {
			Name               string `json:"_name"`
			DeviceInput        int    `json:"coreaudio_device_input"`
			DefaultAudioInput  string `json:"coreaudio_default_audio_input_device"`
			DeviceTransport    string `json:"coreaudio_device_transport"`
			DeviceManufacturer string `json:"coreaudio_device_manufacturer"`
		} `json:"_items"`
	} `json:"SPAudioDataType"`
}

// GetAudioInputDevices queries native macOS system profiler for connected input microphones
func GetAudioInputDevices() ([]AudioDeviceInfo, error) {
	cmd := exec.Command("system_profiler", "SPAudioDataType", "-json", "-timeout", "3")
	out, err := cmd.Output()
	if err != nil {
		// Fallback to fast system_profiler without timeout flag
		cmd2 := exec.Command("system_profiler", "SPAudioDataType", "-json")
		var err2 error
		out, err2 = cmd2.Output()
		if err2 != nil {
			return nil, err
		}
	}

	var parsed spAudioOutput
	if err := json.Unmarshal(out, &parsed); err != nil {
		return nil, err
	}

	var results []AudioDeviceInfo
	for _, dt := range parsed.SPAudioDataType {
		for _, item := range dt.Items {
			if item.DeviceInput > 0 {
				transport := "built-in"
				tLower := strings.ToLower(item.DeviceTransport)
				if strings.Contains(tLower, "bluetooth") {
					transport = "bluetooth"
				} else if strings.Contains(tLower, "usb") {
					transport = "usb"
				} else if strings.Contains(tLower, "virtual") {
					transport = "virtual"
				} else if strings.Contains(tLower, "unknown") {
					transport = "continuity"
				}

				results = append(results, AudioDeviceInfo{
					Name:         item.Name,
					IsDefault:    item.DefaultAudioInput == "spaudio_yes",
					Transport:    transport,
					Manufacturer: item.DeviceManufacturer,
				})
			}
		}
	}

	return results, nil
}
