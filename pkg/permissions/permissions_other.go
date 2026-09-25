//go:build !darwin

package permissions

func checkDarwinAccessibility() Status {
	return Status{
		Granted:  true,
		Message:  "Not darwin",
		Platform: "other",
	}
}
