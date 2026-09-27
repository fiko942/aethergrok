//go:build darwin

package permissions

import (
	"os/exec"
)

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework ApplicationServices -framework Foundation
#import <ApplicationServices/ApplicationServices.h>
#import <Foundation/Foundation.h>

static bool checkAndPromptAX() {
    NSDictionary *options = @{(__bridge id)kAXTrustedCheckOptionPrompt: @YES};
    return AXIsProcessTrustedWithOptions((__bridge CFDictionaryRef)options);
}

static bool checkAXWithoutPrompt() {
    return AXIsProcessTrusted();
}
*/
import "C"

func checkDarwinAccessibility() Status {
	// First check without prompt
	trusted := bool(C.checkAXWithoutPrompt())
	if trusted {
		return Status{
			Granted:  true,
			Message:  "Accessibility permission already granted",
			Platform: "darwin",
		}
	}

	// Trigger system prompt via AXIsProcessTrustedWithOptions
	promptedTrusted := bool(C.checkAndPromptAX())
	if promptedTrusted {
		return Status{
			Granted:  true,
			Message:  "Accessibility permission granted",
			Platform: "darwin",
		}
	}

	// If not trusted, we also provide a fallback to open the Accessibility settings pane if desired
	return Status{
		Granted:  false,
		Message:  "Accessibility permission required for global shortcuts. Please allow AetherGrok in System Settings > Privacy & Security > Accessibility.",
		Platform: "darwin",
	}
}

// OpenAccessibilityPreferences opens the macOS Accessibility system settings pane directly
func OpenAccessibilityPreferences() error {
	cmd := exec.Command("open", "x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility")
	return cmd.Run()
}
