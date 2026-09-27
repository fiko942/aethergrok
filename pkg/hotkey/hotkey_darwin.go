//go:build darwin

package hotkey

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework ApplicationServices -framework Foundation -framework Cocoa
#import <ApplicationServices/ApplicationServices.h>
#import <Foundation/Foundation.h>
#import <Cocoa/Cocoa.h>

extern void triggerHotkeyCallback();

// Atomic configuration of targets
static int g_target_keycode = -1;
static int g_is_modifier_alone = 0; // 1 if standalone modifier (ShiftRight, etc.)
static uint64_t g_required_flags = 0;
static CFMachPortRef g_event_tap = NULL;
static CFRunLoopSourceRef g_run_loop_source = NULL;
static CFRunLoopRef g_run_loop = NULL;
static int g_is_running = 0;

static CGEventRef eventTapCallback(CGEventTapProxy proxy, CGEventType type, CGEventRef event, void *refcon) {
    if (type == kCGEventTapDisabledByTimeout || type == kCGEventTapDisabledByUserInput) {
        if (g_event_tap) {
            CGEventTapEnable(g_event_tap, true);
        }
        return event;
    }

    int target_kc = g_target_keycode;
    if (target_kc < 0) {
        return event;
    }

    if (g_is_modifier_alone) {
        if (type == kCGEventFlagsChanged) {
            int64_t keycode = CGEventGetIntegerValueField(event, kCGKeyboardEventKeycode);
            if (keycode == target_kc) {
                CGEventFlags flags = CGEventGetFlags(event);
                int is_pressed = 0;
                // Check if the specific modifier key is now active
                if (target_kc == 60 || target_kc == 56) { // Right shift (60) or Left shift (56)
                    is_pressed = (flags & kCGEventFlagMaskShift) != 0;
                } else if (target_kc == 54 || target_kc == 55) { // Right cmd (54) or Left cmd (55)
                    is_pressed = (flags & kCGEventFlagMaskCommand) != 0;
                } else if (target_kc == 58 || target_kc == 61) { // Left option (58) or Right option (61)
                    is_pressed = (flags & kCGEventFlagMaskAlternate) != 0;
                } else if (target_kc == 59 || target_kc == 62) { // Left ctrl (59) or Right ctrl (62)
                    is_pressed = (flags & kCGEventFlagMaskControl) != 0;
                }

                if (is_pressed) {
                    triggerHotkeyCallback();
                }
            }
        }
    } else {
        if (type == kCGEventKeyDown) {
            int64_t keycode = CGEventGetIntegerValueField(event, kCGKeyboardEventKeycode);
            if (keycode == target_kc) {
                CGEventFlags flags = CGEventGetFlags(event);
                uint64_t req = g_required_flags;
                // Mask out device independent flags
                bool cmdReq = (req & kCGEventFlagMaskCommand) != 0;
                bool shiftReq = (req & kCGEventFlagMaskShift) != 0;
                bool altReq = (req & kCGEventFlagMaskAlternate) != 0;
                bool ctrlReq = (req & kCGEventFlagMaskControl) != 0;

                bool cmdDown = (flags & kCGEventFlagMaskCommand) != 0;
                bool shiftDown = (flags & kCGEventFlagMaskShift) != 0;
                bool altDown = (flags & kCGEventFlagMaskAlternate) != 0;
                bool ctrlDown = (flags & kCGEventFlagMaskControl) != 0;

                if (cmdReq == cmdDown && shiftReq == shiftDown && altReq == altDown && ctrlReq == ctrlDown) {
                    triggerHotkeyCallback();
                }
            }
        }
    }

    return event;
}

static int startEventTap() {
    if (g_is_running) return 1;

    CGEventMask mask = (1 << kCGEventKeyDown) | (1 << kCGEventFlagsChanged);
    g_event_tap = CGEventTapCreate(
        kCGSessionEventTap,
        kCGHeadInsertEventTap,
        kCGEventTapOptionListenOnly, // Listen only, do not block or consume events
        mask,
        eventTapCallback,
        NULL
    );

    if (!g_event_tap) {
        return 0; // Failed to create event tap (usually lacking Accessibility permission)
    }

    g_run_loop_source = CFMachPortCreateRunLoopSource(kCFAllocatorDefault, g_event_tap, 0);
    g_run_loop = CFRunLoopGetCurrent();
    CFRunLoopAddSource(g_run_loop, g_run_loop_source, kCFRunLoopCommonModes);
    CGEventTapEnable(g_event_tap, true);
    g_is_running = 1;

    CFRunLoopRun();
    return 1;
}

static void stopEventTap() {
    if (!g_is_running) return;
    if (g_event_tap) {
        CGEventTapEnable(g_event_tap, false);
    }
    if (g_run_loop) {
        CFRunLoopStop(g_run_loop);
    }
    if (g_run_loop_source && g_run_loop) {
        CFRunLoopRemoveSource(g_run_loop, g_run_loop_source, kCFRunLoopCommonModes);
    }
    if (g_event_tap) {
        CFRelease(g_event_tap);
        g_event_tap = NULL;
    }
    if (g_run_loop_source) {
        CFRelease(g_run_loop_source);
        g_run_loop_source = NULL;
    }
    g_run_loop = NULL;
    g_is_running = 0;
}

static void configureTarget(int keycode, int is_modifier_alone, uint64_t req_flags) {
    g_target_keycode = keycode;
    g_is_modifier_alone = is_modifier_alone;
    g_required_flags = req_flags;
}
*/
import "C"

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

var (
	globalCallbackMu sync.Mutex
	globalCallback   Handler
	lastTriggerTime  time.Time
)

//export triggerHotkeyCallback
func triggerHotkeyCallback() {
	globalCallbackMu.Lock()
	cb := globalCallback
	now := time.Now()
	// Debounce 250ms to prevent multiple events on a single key press
	if now.Sub(lastTriggerTime) < 250*time.Millisecond {
		globalCallbackMu.Unlock()
		return
	}
	lastTriggerTime = now
	globalCallbackMu.Unlock()

	if cb != nil {
		go cb()
	}
}

type darwinManager struct {
	running bool
	stopCh  chan struct{}
}

func newPlatformManager() platformManager {
	return &darwinManager{}
}

func (d *darwinManager) start(shortcutStr string, handler Handler) error {
	globalCallbackMu.Lock()
	globalCallback = handler
	globalCallbackMu.Unlock()

	kc, isModAlone, flags, err := parseShortcutDarwin(shortcutStr)
	if err != nil {
		return err
	}

	C.configureTarget(C.int(kc), C.int(isModAlone), C.uint64_t(flags))

	d.stopCh = make(chan struct{})
	d.running = true

	startedCh := make(chan bool, 1)

	go func() {
		// Run loop inside separate goroutine / OS thread
		startedCh <- true
		ok := int(C.startEventTap())
		if ok == 0 {
			fmt.Println("[hotkey] CGEventTap failed to create. Ensure Accessibility permission is granted in macOS System Settings.")
		}
	}()

	<-startedCh
	return nil
}

func (d *darwinManager) update(shortcutStr string) error {
	kc, isModAlone, flags, err := parseShortcutDarwin(shortcutStr)
	if err != nil {
		return err
	}
	C.configureTarget(C.int(kc), C.int(isModAlone), C.uint64_t(flags))
	return nil
}

func (d *darwinManager) stop() {
	if d.running {
		C.stopEventTap()
		d.running = false
	}
}

// macOS Virtual Keycodes
// Reference: Carbon HIToolbox/Events.h
const (
	kVK_RightShift   = 60
	kVK_Shift        = 56
	kVK_RightCommand = 54
	kVK_Command      = 55
	kVK_RightOption  = 61
	kVK_Option       = 58
	kVK_RightControl = 62
	kVK_Control      = 59
	kVK_Space        = 49
	kVK_Return       = 36
	kVK_Tab          = 48
	kVK_Delete       = 51
	kVK_Escape       = 53
)

var darwinKeyMap = map[string]int{
	"a": 0, "s": 1, "d": 2, "f": 3, "h": 4, "g": 5, "z": 6, "x": 7, "c": 8, "v": 9,
	"b": 11, "q": 12, "w": 13, "e": 14, "r": 15, "y": 16, "t": 17, "1": 18, "2": 19,
	"3": 20, "4": 21, "6": 22, "5": 23, "=": 24, "9": 25, "7": 26, "-": 27, "8": 28,
	"0": 29, "]": 30, "o": 31, "u": 32, "[": 33, "i": 34, "p": 35, "l": 37, "j": 38,
	"'": 39, "k": 40, ";": 41, "\\": 42, ",": 43, "/": 44, "n": 45, "m": 46, ".": 47,
	"space": kVK_Space, "return": kVK_Return, "enter": kVK_Return, "tab": kVK_Tab,
	"delete": kVK_Delete, "backspace": kVK_Delete, "escape": kVK_Escape,
}

func parseShortcutDarwin(shortcutStr string) (keycode int, isModifierAlone int, flags uint64, err error) {
	s := strings.TrimSpace(shortcutStr)
	if s == "" {
		s = "CmdOrCtrl+Shift+S"
	}

	norm := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(s, " ", ""), "_", ""))
	// Check standalone modifier keys
	switch norm {
	case "shiftright", "rightshift":
		return kVK_RightShift, 1, 0, nil
	case "shiftleft", "leftshift", "shift":
		return kVK_Shift, 1, 0, nil
	case "metaright", "rightcmd", "cmdright", "rightcommand", "commandright":
		return kVK_RightCommand, 1, 0, nil
	case "metaleft", "leftcmd", "cmdleft", "leftcommand", "commandleft", "cmd", "meta":
		return kVK_Command, 1, 0, nil
	case "altright", "rightalt", "optionright", "rightoption", "optright", "rightopt":
		return kVK_RightOption, 1, 0, nil
	case "altleft", "leftalt", "optionleft", "leftoption", "optleft", "leftopt", "alt", "option":
		return kVK_Option, 1, 0, nil
	case "controlright", "rightctrl", "ctrlright", "rightcontrol":
		return kVK_RightControl, 1, 0, nil
	case "controlleft", "leftctrl", "ctrlleft", "leftcontrol", "ctrl", "control":
		return kVK_Control, 1, 0, nil
	}

	// Parse combination like CmdOrCtrl+Shift+S
	parts := strings.Split(s, "+")
	var reqFlags uint64
	var keyPart string

	for _, p := range parts {
		pClean := strings.ToLower(strings.TrimSpace(p))
		switch pClean {
		case "cmd", "ctrl", "cmdorctrl", "meta", "command":
			reqFlags |= 0x100000 // kCGEventFlagMaskCommand (macOS standard for CmdOrCtrl)
		case "shift":
			reqFlags |= 0x020000 // kCGEventFlagMaskShift
		case "alt", "option", "opt":
			reqFlags |= 0x080000 // kCGEventFlagMaskAlternate
		case "control":
			reqFlags |= 0x040000 // kCGEventFlagMaskControl
		default:
			keyPart = pClean
		}
	}

	if kc, ok := darwinKeyMap[keyPart]; ok {
		return kc, 0, reqFlags, nil
	}

	return -1, 0, 0, fmt.Errorf("unrecognized shortcut key: %s", keyPart)
}
