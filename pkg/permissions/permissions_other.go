//go:build !darwin

package permissions

func checkDarwinAccessibility() Status {
	return Status{
		Granted:  true,
		Message:  "Not darwin",
		Platform: "other",
	}
}

// OpenAccessibilityPreferences is a no-op on non-darwin platforms
func OpenAccessibilityPreferences() error {
	return nil
}

// OpenScreenCapturePreferences is a no-op on non-darwin platforms
func OpenScreenCapturePreferences() error {
	return nil
}

func checkDarwinScreenCapture() Status {
	return Status{
		Granted:  true,
		Message:  "Screen capture permission not restricted on this platform",
		Platform: "other",
	}
}

func requestDarwinScreenCapture() Status {
	return Status{
		Granted:  true,
		Message:  "Screen capture permission granted",
		Platform: "other",
	}
}

func checkDarwinAllPermissions() AllPermissionsStatus {
	return AllPermissionsStatus{
		Platform:   "other",
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

