//go:build windows

package permissions

import (
	"os/exec"
	"syscall"
)

func checkDarwinMicrophone() Status {
	return Status{
		Granted:  true,
		Message:  "Microphone permission granted by Windows",
		Platform: "windows",
	}
}

func requestDarwinMicrophone() Status {
	return checkDarwinMicrophone()
}

// OpenMicrophonePreferences opens Windows 10/11 Privacy & Security -> Microphone settings page
func OpenMicrophonePreferences() error {
	cmd := exec.Command("cmd", "/c", "start", "ms-settings:privacy-microphone")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Start()
}
