package permissions

import "runtime"

// Status represents the permission state
type Status struct {
	Granted bool   `json:"granted"`
	Message string `json:"message"`
	Platform string `json:"platform"`
}

// CheckAndRequestAccessibility checks if the process has accessibility permissions
// and prompts the OS permission dialog if not yet granted (macOS specific).
func CheckAndRequestAccessibility() Status {
	if runtime.GOOS != "darwin" {
		return Status{
			Granted:  true,
			Message:  "Accessibility permission not required on this platform",
			Platform: runtime.GOOS,
		}
	}
	return checkDarwinAccessibility()
}

// CheckMicrophonePermission inspects if microphone access is authorized
func CheckMicrophonePermission() Status {
	if runtime.GOOS != "darwin" {
		return Status{
			Granted:  true,
			Message:  "Microphone permission not restricted on this platform",
			Platform: runtime.GOOS,
		}
	}
	return checkDarwinMicrophone()
}

// RequestMicrophonePermission triggers macOS system prompt if not yet determined
func RequestMicrophonePermission() Status {
	if runtime.GOOS != "darwin" {
		return Status{
			Granted:  true,
			Message:  "Microphone permission granted",
			Platform: runtime.GOOS,
		}
	}
	return requestDarwinMicrophone()
}

