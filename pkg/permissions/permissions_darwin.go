//go:build darwin

package permissions

import (
	"os/exec"
)

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework ApplicationServices -framework Foundation -framework CoreGraphics
#import <ApplicationServices/ApplicationServices.h>
#import <Foundation/Foundation.h>
#import <CoreGraphics/CoreGraphics.h>

static bool checkAndPromptAX() {
    NSDictionary *options = @{(__bridge id)kAXTrustedCheckOptionPrompt: @YES};
    return AXIsProcessTrustedWithOptions((__bridge CFDictionaryRef)options);
}

static bool checkAXWithoutPrompt() {
    return AXIsProcessTrusted();
}

static bool checkScreenCaptureAccess() {
    if (@available(macOS 10.15, *)) {
        return CGPreflightScreenCaptureAccess();
    }
    return true;
}

static bool requestScreenCaptureAccess() {
    if (@available(macOS 10.15, *)) {
        return CGRequestScreenCaptureAccess();
    }
    return true;
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

// checkDarwinScreenCapture checks whether macOS screen capture permission is granted
func checkDarwinScreenCapture() Status {
	granted := bool(C.checkScreenCaptureAccess())
	if granted {
		return Status{
			Granted:  true,
			Message:  "Screen capture permission granted",
			Platform: "darwin",
		}
	}
	return Status{
		Granted:  false,
		Message:  "Screen capture permission required to capture desktop and window contents. Please allow AetherGrok in System Settings > Privacy & Security > Screen & System Audio Recording.",
		Platform: "darwin",
	}
}

// requestDarwinScreenCapture prompts the macOS system dialog for screen recording if not yet authorized
func requestDarwinScreenCapture() Status {
	granted := bool(C.requestScreenCaptureAccess())
	if granted {
		return Status{
			Granted:  true,
			Message:  "Screen capture permission granted",
			Platform: "darwin",
		}
	}
	return checkDarwinScreenCapture()
}

// OpenScreenCapturePreferences opens the macOS Screen Recording settings pane directly
func OpenScreenCapturePreferences() error {
	cmd := exec.Command("open", "x-apple.systempreferences:com.apple.preference.security?Privacy_ScreenCapture")
	return cmd.Run()
}

// checkDarwinAllPermissions inspects Accessibility, Microphone, and Screen Capture
func checkDarwinAllPermissions() AllPermissionsStatus {
	axStatus := checkDarwinAccessibility()
	screenStatus := checkDarwinScreenCapture()
	micStatus := checkDarwinMicrophone()

	items := []SystemPermissionItem{
		{
			ID:          "accessibility",
			Title:       "Accessibility & Global Shortcuts",
			Description: "Allows global hotkeys and shortcut triggering across active windows.",
			Granted:     axStatus.Granted,
			Message:     axStatus.Message,
			Required:    true,
		},
		{
			ID:          "screen_capture",
			Title:       "Screen Recording",
			Description: "Enables capturing desktop, application windows, and visual context for Grok Vision.",
			Granted:     screenStatus.Granted,
			Message:     screenStatus.Message,
			Required:    true,
		},
		{
			ID:          "microphone",
			Title:       "Microphone",
			Description: "Enables hands-free voice input and speech-to-text recognition.",
			Granted:     micStatus.Granted,
			Message:     micStatus.Message,
			Required:    false,
		},
	}

	allGranted := axStatus.Granted && screenStatus.Granted

	return AllPermissionsStatus{
		Platform:   "darwin",
		AllGranted: allGranted,
		Items:      items,
	}
}

