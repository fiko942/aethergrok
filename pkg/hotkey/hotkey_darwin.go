//go:build darwin

package hotkey

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework ApplicationServices -framework Foundation -framework Cocoa
#import <ApplicationServices/ApplicationServices.h>
#import <Foundation/Foundation.h>
#import <Cocoa/Cocoa.h>

extern void triggerDarwinKeyEvent(int keycode, int eventType, uint64_t flags);

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

    int64_t keycode = CGEventGetIntegerValueField(event, kCGKeyboardEventKeycode);
    CGEventFlags flags = CGEventGetFlags(event);

    triggerDarwinKeyEvent((int)keycode, (int)type, (uint64_t)flags);

    return event;
}

static int startEventTap() {
    if (g_is_running) return 1;

    CGEventMask mask = (1 << kCGEventKeyDown) | (1 << kCGEventKeyUp) | (1 << kCGEventFlagsChanged);
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
*/
import "C"

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// CGEventType constants
const (
	cgEventKeyDown      = 10 // kCGEventKeyDown
	cgEventKeyUp        = 11 // kCGEventKeyUp
	cgEventFlagsChanged = 12 // kCGEventFlagsChanged
)

// CGEventFlagMask values
const (
	cgEventFlagMaskCommand   = 0x00100000
	cgEventFlagMaskShift     = 0x00020000
	cgEventFlagMaskAlternate = 0x00080000
	cgEventFlagMaskControl   = 0x00040000
)

var (
	darwinRegistryMu sync.RWMutex
	darwinRegistry   = make(map[*darwinHotkeyManager]struct{})
	darwinRunning    bool
)

type darwinHotkeyManager struct {
	mu          sync.Mutex
	running     bool
	handler     Handler
	keyHandler  KeyHandler
	shortcut    string
	targetKC    int
	isModAlone  int
	reqFlags    uint64
	isDown      bool
	lastTrigger int64
	events      chan func()
	stopWorker  chan struct{}
}

func newPlatformManager() platformManager {
	return &darwinHotkeyManager{}
}

//export triggerDarwinKeyEvent
func triggerDarwinKeyEvent(keycode int, eventType int, flags uint64) {
	darwinRegistryMu.RLock()
	managers := make([]*darwinHotkeyManager, 0, len(darwinRegistry))
	for m := range darwinRegistry {
		managers = append(managers, m)
	}
	darwinRegistryMu.RUnlock()

	for _, mgr := range managers {
		mgr.mu.Lock()
		targetKC := mgr.targetKC
		isModAlone := mgr.isModAlone
		reqFlags := mgr.reqFlags
		isCurrentlyDown := mgr.isDown
		h := mgr.handler
		kh := mgr.keyHandler
		eventChan := mgr.events
		mgr.mu.Unlock()

		if targetKC < 0 || eventChan == nil {
			continue
		}

		if isModAlone == 1 {
			// Standalone modifier key (ShiftRight, OptionRight, Right Command, etc.)
			if eventType == cgEventFlagsChanged && keycode == targetKC {
				isPressed := false
				if targetKC == kVK_RightShift || targetKC == kVK_Shift {
					isPressed = (flags & cgEventFlagMaskShift) != 0
				} else if targetKC == kVK_RightCommand || targetKC == kVK_Command {
					isPressed = (flags & cgEventFlagMaskCommand) != 0
				} else if targetKC == kVK_RightOption || targetKC == kVK_Option {
					isPressed = (flags & cgEventFlagMaskAlternate) != 0
				} else if targetKC == kVK_RightControl || targetKC == kVK_Control {
					isPressed = (flags & cgEventFlagMaskControl) != 0
				}

				if isPressed {
					if !isCurrentlyDown {
						mgr.mu.Lock()
						mgr.isDown = true
						mgr.mu.Unlock()

						if kh != nil {
							fn := func() { kh("down") }
							select {
							case eventChan <- fn:
							default:
							}
						}
						if h != nil {
							now := time.Now().UnixMilli()
							if now-atomic.LoadInt64(&mgr.lastTrigger) > 150 {
								atomic.StoreInt64(&mgr.lastTrigger, now)
								fn := func() { h() }
								select {
								case eventChan <- fn:
								default:
								}
							}
						}
					}
				} else {
					if isCurrentlyDown {
						mgr.mu.Lock()
						mgr.isDown = false
						mgr.mu.Unlock()

						if kh != nil {
							fn := func() { kh("up") }
							select {
							case eventChan <- fn:
							default:
							}
						}
					}
				}
			}
		} else {
			// Regular key or modifier combination (e.g. '\', '/', 'Cmd+Shift+S', 'Alt+Space')
			if keycode == targetKC {
				if eventType == cgEventKeyDown {
					cmdReq := (reqFlags & cgEventFlagMaskCommand) != 0
					shiftReq := (reqFlags & cgEventFlagMaskShift) != 0
					altReq := (reqFlags & cgEventFlagMaskAlternate) != 0
					ctrlReq := (reqFlags & cgEventFlagMaskControl) != 0

					cmdDown := (flags & cgEventFlagMaskCommand) != 0
					shiftDown := (flags & cgEventFlagMaskShift) != 0
					altDown := (flags & cgEventFlagMaskAlternate) != 0
					ctrlDown := (flags & cgEventFlagMaskControl) != 0

					if cmdReq == cmdDown && shiftReq == shiftDown && altReq == altDown && ctrlReq == ctrlDown {
						if !isCurrentlyDown {
							mgr.mu.Lock()
							mgr.isDown = true
							mgr.mu.Unlock()

							if kh != nil {
								fn := func() { kh("down") }
								select {
								case eventChan <- fn:
								default:
								}
							}
							if h != nil {
								now := time.Now().UnixMilli()
								if now-atomic.LoadInt64(&mgr.lastTrigger) > 150 {
									atomic.StoreInt64(&mgr.lastTrigger, now)
									fn := func() { h() }
									select {
									case eventChan <- fn:
									default:
									}
								}
							}
						}
					}
				} else if eventType == cgEventKeyUp {
					if isCurrentlyDown {
						mgr.mu.Lock()
						mgr.isDown = false
						mgr.mu.Unlock()

						if kh != nil {
							fn := func() { kh("up") }
							select {
							case eventChan <- fn:
							default:
							}
						}
					}
				}
			}
		}
	}
}

func (d *darwinHotkeyManager) start(shortcutStr string, handler Handler) error {
	return d.startWithKeyHandler(shortcutStr, handler, nil)
}

func (d *darwinHotkeyManager) startWithKeyHandler(shortcutStr string, handler Handler, keyHandler KeyHandler) error {
	kc, isModAlone, flags, err := parseShortcutDarwin(shortcutStr)
	if err != nil {
		return err
	}

	d.mu.Lock()
	d.shortcut = shortcutStr
	d.handler = handler
	d.keyHandler = keyHandler
	d.targetKC = kc
	d.isModAlone = isModAlone
	d.reqFlags = flags
	d.running = true
	d.isDown = false
	d.events = make(chan func(), 64)
	d.stopWorker = make(chan struct{})
	eventChan := d.events
	stopChan := d.stopWorker
	d.mu.Unlock()

	go func() {
		for {
			select {
			case fn := <-eventChan:
				if fn != nil {
					fn()
				}
			case <-stopChan:
				return
			}
		}
	}()

	darwinRegistryMu.Lock()
	darwinRegistry[d] = struct{}{}
	shouldStartGlobal := !darwinRunning
	if shouldStartGlobal {
		darwinRunning = true
	}
	darwinRegistryMu.Unlock()

	if shouldStartGlobal {
		go func() {
			ok := int(C.startEventTap())
			if ok == 0 {
				fmt.Println("[hotkey] CGEventTap failed to create. Ensure Accessibility permission is granted in macOS System Settings.")
			}
		}()
	}

	return nil
}

func (d *darwinHotkeyManager) update(shortcutStr string) error {
	kc, isModAlone, flags, err := parseShortcutDarwin(shortcutStr)
	if err != nil {
		return err
	}

	d.mu.Lock()
	d.shortcut = shortcutStr
	d.targetKC = kc
	d.isModAlone = isModAlone
	d.reqFlags = flags
	d.isDown = false
	d.mu.Unlock()

	return nil
}

func (d *darwinHotkeyManager) stop() {
	d.mu.Lock()
	d.running = false
	d.isDown = false
	if d.stopWorker != nil {
		close(d.stopWorker)
		d.stopWorker = nil
	}
	d.mu.Unlock()

	darwinRegistryMu.Lock()
	delete(darwinRegistry, d)
	shouldStopGlobal := len(darwinRegistry) == 0 && darwinRunning
	if shouldStopGlobal {
		darwinRunning = false
	}
	darwinRegistryMu.Unlock()

	if shouldStopGlobal {
		C.stopEventTap()
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
	"`": 50,
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
	case "shiftright", "rightshift", "rshift":
		return kVK_RightShift, 1, 0, nil
	case "shiftleft", "leftshift", "lshift", "shift":
		return kVK_Shift, 1, 0, nil
	case "metaright", "rightcmd", "cmdright", "rightcommand", "commandright", "rcmd":
		return kVK_RightCommand, 1, 0, nil
	case "metaleft", "leftcmd", "cmdleft", "leftcommand", "commandleft", "cmd", "meta", "lcmd":
		return kVK_Command, 1, 0, nil
	case "altright", "rightalt", "optionright", "rightoption", "optright", "rightopt", "ropt":
		return kVK_RightOption, 1, 0, nil
	case "altleft", "leftalt", "optionleft", "leftoption", "optleft", "leftopt", "alt", "option", "lopt":
		return kVK_Option, 1, 0, nil
	case "controlright", "rightctrl", "ctrlright", "rightcontrol", "rctrl":
		return kVK_RightControl, 1, 0, nil
	case "controlleft", "leftctrl", "ctrlleft", "leftcontrol", "ctrl", "control", "lctrl":
		return kVK_Control, 1, 0, nil
	}

	// Parse combination like CmdOrCtrl+Shift+S or single keys like "\"
	parts := strings.Split(s, "+")
	var reqFlags uint64
	var keyPart string

	for _, p := range parts {
		pClean := strings.ToLower(strings.TrimSpace(p))
		switch pClean {
		case "cmd", "ctrl", "cmdorctrl", "meta", "command":
			reqFlags |= cgEventFlagMaskCommand // kCGEventFlagMaskCommand
		case "control":
			reqFlags |= cgEventFlagMaskControl
		case "shift":
			reqFlags |= cgEventFlagMaskShift
		case "alt", "option", "opt":
			reqFlags |= cgEventFlagMaskAlternate
		default:
			keyPart = pClean
		}
	}

	if keyPart == "" {
		// If only modifiers specified, default to standalone modifier keycode
		if reqFlags&cgEventFlagMaskShift != 0 {
			return kVK_RightShift, 1, 0, nil
		}
		if reqFlags&cgEventFlagMaskAlternate != 0 {
			return kVK_RightOption, 1, 0, nil
		}
		if reqFlags&cgEventFlagMaskControl != 0 {
			return kVK_RightControl, 1, 0, nil
		}
		if reqFlags&cgEventFlagMaskCommand != 0 {
			return kVK_RightCommand, 1, 0, nil
		}
		return 0, 0, 0, fmt.Errorf("invalid shortcut: %s", shortcutStr)
	}

	kc, ok := darwinKeyMap[keyPart]
	if !ok {
		return 0, 0, 0, fmt.Errorf("unsupported key %q in shortcut %q", keyPart, shortcutStr)
	}

	return kc, 0, reqFlags, nil
}
