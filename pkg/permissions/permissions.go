package permissions

import "runtime"

// Status represents the permission state
type Status struct {
	Granted bool   `json:"granted"`
	Message string `json:"message"`
	Platform string `json:"platform"`
}

// SystemPermissionItem represents an individual system permission status
type SystemPermissionItem struct {
	ID          string `json:"id"`          // "accessibility", "microphone", "screen_capture"
	Title       string `json:"title"`
	Description string `json:"description"`
	Granted     bool   `json:"granted"`
	Message     string `json:"message"`
	Required    bool   `json:"required"`
}

// AllPermissionsStatus holds the aggregated permission states across system requirements
type AllPermissionsStatus struct {
	Platform   string                 `json:"platform"`
	AllGranted bool                   `json:"allGranted"`
	Items      []SystemPermissionItem `json:"items"`
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

// CheckScreenCapturePermission inspects if screen capture / recording access is authorized
func CheckScreenCapturePermission() Status {
	if runtime.GOOS != "darwin" {
		return Status{
			Granted:  true,
			Message:  "Screen capture permission not restricted on this platform",
			Platform: runtime.GOOS,
		}
	}
	return checkDarwinScreenCapture()
}

// RequestScreenCapturePermission triggers macOS system prompt or preflight for screen capture
func RequestScreenCapturePermission() Status {
	if runtime.GOOS != "darwin" {
		return Status{
			Granted:  true,
			Message:  "Screen capture permission granted",
			Platform: runtime.GOOS,
		}
	}
	return requestDarwinScreenCapture()
}

// CheckAllSystemPermissions inspects and aggregates the status of all required system permissions
func CheckAllSystemPermissions() AllPermissionsStatus {
	if runtime.GOOS != "darwin" {
		return AllPermissionsStatus{
			Platform:   runtime.GOOS,
			AllGranted: true,
			Items: []SystemPermissionItem{
				{
					ID:          "accessibility",
					Title:       "Accessibility & Global Shortcuts",
					Description: "Allows global hotkeys and shortcut triggering across active windows.",
					Granted:     true,
					Message:     "Not required on this platform",
					Required:    true,
				},
				{
					ID:          "screen_capture",
					Title:       "Screen Recording",
					Description: "Enables capturing desktop, application windows, and visual context for Grok Vision.",
					Granted:     true,
					Message:     "Not required on this platform",
					Required:    true,
				},
				{
					ID:          "microphone",
					Title:       "Microphone",
					Description: "Enables hands-free voice input and speech-to-text recognition.",
					Granted:     true,
					Message:     "Not required on this platform",
					Required:    false,
				},
			},
		}
	}
	return checkDarwinAllPermissions()
}


