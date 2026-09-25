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
