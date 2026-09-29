//go:build windows

package permissions

import (
	"testing"
)

func TestParseWindowsAudioDevicesOutput(t *testing.T) {
	sampleJSON := `[
		{"Name": "Microphone (Realtek High Definition Audio)", "Manufacturer": "Realtek", "IsDefault": true},
		{"Name": "Headset (WH-1000XM4 Hands-Free AG Audio)", "Manufacturer": "Sony", "IsDefault": false},
		{"Name": "Yeti Stereo Microphone", "Manufacturer": "Blue Microphones", "IsDefault": false},
		{"Name": "VoiceMeeter Output", "Manufacturer": "VB-Audio", "IsDefault": false}
	]`

	devices, err := parseWindowsAudioDevicesJSON([]byte(sampleJSON))
	if err != nil {
		t.Fatalf("unexpected error parsing audio devices: %v", err)
	}

	if len(devices) != 4 {
		t.Fatalf("expected 4 devices, got %d", len(devices))
	}

	if !devices[0].IsDefault || devices[0].Transport != "built-in" {
		t.Errorf("device 0 expected default built-in, got isDefault=%v transport=%s", devices[0].IsDefault, devices[0].Transport)
	}

	if devices[1].Transport != "bluetooth" {
		t.Errorf("device 1 expected bluetooth, got %s", devices[1].Transport)
	}

	if devices[2].Transport != "usb" {
		t.Errorf("device 2 expected usb, got %s", devices[2].Transport)
	}

	if devices[3].Transport != "virtual" {
		t.Errorf("device 3 expected virtual, got %s", devices[3].Transport)
	}
}

func TestGetAudioInputDevicesLive(t *testing.T) {
	devs, err := GetAudioInputDevices()
	if err != nil {
		t.Logf("GetAudioInputDevices returned error (expected in minimal CI environment): %v", err)
	} else {
		t.Logf("Discovered %d audio input devices", len(devs))
		for i, d := range devs {
			t.Logf("  [%d] %s (default: %v, transport: %s, manufacturer: %s)", i, d.Name, d.IsDefault, d.Transport, d.Manufacturer)
		}
	}
}
