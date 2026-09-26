//go:build darwin

package permissions

import (
	"os/exec"
)

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework AVFoundation -framework Foundation
#import <AVFoundation/AVFoundation.h>
#import <Foundation/Foundation.h>

static int checkMicrophoneStatus() {
    AVAuthorizationStatus status = [AVCaptureDevice authorizationStatusForMediaType:AVMediaTypeAudio];
    return (int)status;
}

static void requestMicrophoneAccess() {
    [AVCaptureDevice requestAccessForMediaType:AVMediaTypeAudio completionHandler:^(BOOL granted) {
        // Callback handled asynchronously by system
    }];
}
*/
import "C"

// checkDarwinMicrophone inspects macOS AVFoundation authorization status for microphone
func checkDarwinMicrophone() Status {
	status := int(C.checkMicrophoneStatus())
	// AVAuthorizationStatusNotDetermined = 0
	// AVAuthorizationStatusRestricted = 1
	// AVAuthorizationStatusDenied = 2
	// AVAuthorizationStatusAuthorized = 3
	switch status {
	case 3:
		return Status{
			Granted:  true,
			Message:  "Microphone permission granted",
			Platform: "darwin",
		}
	case 0:
		return Status{
			Granted:  false,
			Message:  "Microphone permission not determined (request prompt available)",
			Platform: "darwin",
		}
	case 1:
		return Status{
			Granted:  false,
			Message:  "Microphone access is restricted by device policy",
			Platform: "darwin",
		}
	default:
		return Status{
			Granted:  false,
			Message:  "Microphone permission denied. Please allow in System Settings > Privacy & Security > Microphone",
			Platform: "darwin",
		}
	}
}

func requestDarwinMicrophone() Status {
	status := int(C.checkMicrophoneStatus())
	if status == 0 {
		C.requestMicrophoneAccess()
	}
	return checkDarwinMicrophone()
}

// OpenMicrophonePreferences opens the macOS Microphone Privacy & Security settings pane
func OpenMicrophonePreferences() error {
	script := "tell application \"System Settings\" to activate\n" +
		"do shell script \"open 'x-apple.systempreferences:com.apple.preference.security?Privacy_Microphone'\""
	cmd := exec.Command("osascript", "-e", script)
	return cmd.Run()
}
