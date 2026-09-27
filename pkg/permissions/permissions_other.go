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
