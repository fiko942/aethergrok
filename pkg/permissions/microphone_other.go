//go:build !darwin

package permissions

func checkDarwinMicrophone() Status {
	return Status{
		Granted:  true,
		Message:  "Microphone permission handled by OS/browser",
		Platform: "other",
	}
}

func requestDarwinMicrophone() Status {
	return checkDarwinMicrophone()
}

// OpenMicrophonePreferences opens microphone preferences (non-darwin stub)
func OpenMicrophonePreferences() error {
	return nil
}
