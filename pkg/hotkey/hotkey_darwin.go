//go:build darwin

package hotkey

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework ApplicationServices -framework Foundation -framework Cocoa -framework Carbon
#import <ApplicationServices/ApplicationServices.h>
#import <Foundation/Foundation.h>
#import <Cocoa/Cocoa.h>
#import <Carbon/Carbon.h>
#include <stdint.h>
#include <stdbool.h>

extern void triggerDarwinKeyEvent(int64_t keycode, int64_t eventType, uint64_t flags);
extern void onDarwinTapStarted(void);
extern void onCarbonHotKeyAction(unsigned int hotKeyID, int isDown);

static CFMachPortRef g_event_tap = NULL;
static CFRunLoopSourceRef g_run_loop_source = NULL;
static CFRunLoopRef g_run_loop = NULL;
static int g_is_running = 0;
static int g_carbon_installed = 0;

static CGEventRef eventTapCallback(CGEventTapProxy proxy, CGEventType type, CGEventRef event, void *refcon) {
    if (type == kCGEventTapDisabledByTimeout || type == kCGEventTapDisabledByUserInput) {
        if (g_event_tap) {
            CGEventTapEnable(g_event_tap, true);
        }
        return event;
    }

    int64_t keycode = CGEventGetIntegerValueField(event, kCGKeyboardEventKeycode);
    CGEventFlags flags = CGEventGetFlags(event);

    triggerDarwinKeyEvent((int64_t)keycode, (int64_t)type, (uint64_t)flags);

    return event;
}

static int isPhysicalKeyDown(int64_t keycode) {
    return CGEventSourceKeyState(kCGEventSourceStateCombinedSessionState, (CGKeyCode)keycode) ? 1 : 0;
}

static pascal OSStatus globalCarbonHotKeyHandler(EventHandlerCallRef nextHandler, EventRef theEvent, void *userData) {
    UInt32 kind = GetEventKind(theEvent);
    EventHotKeyID hkID;
    GetEventParameter(theEvent, kEventParamDirectObject, typeEventHotKeyID, NULL, sizeof(hkID), NULL, &hkID);
    
    int isDown = (kind == kEventHotKeyPressed) ? 1 : 0;
    onCarbonHotKeyAction((unsigned int)hkID.id, isDown);
    return noErr;
}

static int ensureCarbonHandlerInstalled() {
    if (g_carbon_installed) return 1;
    EventTypeSpec eventTypes[2];
    eventTypes[0].eventClass = kEventClassKeyboard;
    eventTypes[0].eventKind = kEventHotKeyPressed;
    eventTypes[1].eventClass = kEventClassKeyboard;
    eventTypes[1].eventKind = kEventHotKeyReleased;
    OSStatus err = InstallApplicationEventHandler(&globalCarbonHotKeyHandler, 2, eventTypes, NULL, NULL);
    if (err == noErr) {
        g_carbon_installed = 1;
        return 1;
    }
    return 0;
}

static void *registerCarbonKeyRef(unsigned int id, unsigned int keyCode, unsigned int carbonModifiers) {
    ensureCarbonHandlerInstalled();
    EventHotKeyRef ref = NULL;
    EventHotKeyID hkID;
    hkID.signature = 'AETH';
    hkID.id = (UInt32)id;
    OSStatus err = RegisterEventHotKey((UInt32)keyCode, (UInt32)carbonModifiers, hkID, GetApplicationEventTarget(), 0, &ref);
    if (err != noErr) {
        return NULL;
    }
    return (void *)ref;
}

static void unregisterCarbonKeyRef(void *ref) {
    if (ref) {
        UnregisterEventHotKey((EventHotKeyRef)ref);
    }
}

static int startEventTap() {
    if (g_is_running) return 1;

    CGEventMask mask = (1 << kCGEventKeyDown) | (1 << kCGEventKeyUp) | (1 << kCGEventFlagsChanged);
    
    // Attempt kCGSessionEventTap with kCGEventTapOptionDefault to intercept global keyboard events reliably
    g_event_tap = CGEventTapCreate(
        kCGSessionEventTap,
        kCGHeadInsertEventTap,
        kCGEventTapOptionDefault,
        mask,
        eventTapCallback,
        NULL
    );

    // Fall back to kCGHIDEventTap if needed
    if (!g_event_tap) {
        g_event_tap = CGEventTapCreate(
            kCGHIDEventTap,
            kCGHeadInsertEventTap,
            kCGEventTapOptionDefault,
            mask,
            eventTapCallback,
            NULL
        );
    }

    // Fall back to kCGEventTapOptionListenOnly
    if (!g_event_tap) {
        g_event_tap = CGEventTapCreate(
            kCGSessionEventTap,
            kCGHeadInsertEventTap,
            kCGEventTapOptionListenOnly,
            mask,
            eventTapCallback,
            NULL
        );
    }

    if (!g_event_tap) {
        return 0; // Failed to create event tap (usually lacking Accessibility permission)
    }

    g_run_loop_source = CFMachPortCreateRunLoopSource(kCFAllocatorDefault, g_event_tap, 0);
    g_run_loop = CFRunLoopGetCurrent();
    CFRunLoopAddSource(g_run_loop, g_run_loop_source, kCFRunLoopCommonModes);

    // Also attach to main run loop for comprehensive background event dispatch
    CFRunLoopRef mainRL = CFRunLoopGetMain();
    if (mainRL && mainRL != g_run_loop) {
        CFRunLoopAddSource(mainRL, g_run_loop_source, kCFRunLoopCommonModes);
    }

    CGEventTapEnable(g_event_tap, true);
    g_is_running = 1;

    // Notify Go that event tap is active and running before entering the runloop
    onDarwinTapStarted();

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
    CFRunLoopRef mainRL = CFRunLoopGetMain();
    if (g_run_loop_source && mainRL && mainRL != g_run_loop) {
        CFRunLoopRemoveSource(mainRL, g_run_loop_source, kCFRunLoopCommonModes);
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
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
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
	darwinReadyChan  chan struct{}

	carbonMu       sync.Mutex
	carbonManagers = make(map[uint32]*darwinHotkeyManager)
	nextCarbonID   uint32 = 1
)

//export onCarbonHotKeyAction
func onCarbonHotKeyAction(hotKeyID C.uint, isDown C.int) {
	dispatchCarbonHotKey(uint32(hotKeyID), isDown == 1)
}

func dispatchCarbonHotKey(id uint32, isDown bool) {
	carbonMu.Lock()
	mgr, ok := carbonManagers[id]
	carbonMu.Unlock()
	if !ok || mgr == nil {
		return
	}

	mgr.mu.Lock()
	kh := mgr.keyHandler
	h := mgr.handler
	isCurrentlyDown := mgr.isDown
	mgr.mu.Unlock()

	if isDown {
		if !isCurrentlyDown {
			mgr.mu.Lock()
			mgr.isDown = true
			mgr.startWatchdogLocked()
			mgr.mu.Unlock()

			if kh != nil {
				mgr.enqueueEvent(func() { kh("down") })
			}
			if h != nil {
				now := time.Now().UnixMilli()
				if now-atomic.LoadInt64(&mgr.lastTrigger) > 150 {
					atomic.StoreInt64(&mgr.lastTrigger, now)
					mgr.enqueueEvent(func() { h() })
				}
			}
		}
	} else {
		if isCurrentlyDown {
			mgr.mu.Lock()
			mgr.isDown = false
			mgr.stopWatchdogLocked()
			mgr.mu.Unlock()

			if kh != nil {
				mgr.enqueueEvent(func() { kh("up") })
			}
		}
	}
}

//export onDarwinTapStarted
func onDarwinTapStarted() {
	darwinRegistryMu.Lock()
	darwinRunning = true
	if darwinReadyChan != nil {
		select {
		case <-darwinReadyChan:
		default:
			close(darwinReadyChan)
		}
	}
	darwinRegistryMu.Unlock()
}

type darwinHotkeyManager struct {
	mu           sync.Mutex
	running      bool
	handler      Handler
	keyHandler   KeyHandler
	shortcut     string
	targetKC     int
	isModAlone   int
	reqFlags     uint64
	carbonRef    unsafe.Pointer
	carbonID     uint32
	isDown       bool
	lastTrigger  int64
	events       chan func()
	stopWorker   chan struct{}
	stopWatchdog chan struct{}
}

// Helper to send events to the manager event queue with drop-resistant buffering
func (d *darwinHotkeyManager) enqueueEvent(fn func()) {
	if fn == nil {
		return
	}
	d.mu.Lock()
	ch := d.events
	d.mu.Unlock()
	if ch == nil {
		return
	}
	select {
	case ch <- fn:
	default:
		// Channel buffer is temporarily saturated; dispatch via background goroutine
		go func() {
			defer func() { _ = recover() }()
			ch <- fn
		}()
	}
}

func (d *darwinHotkeyManager) startWatchdogLocked() {
	if d.stopWatchdog != nil {
		return
	}
	stopCh := make(chan struct{})
	d.stopWatchdog = stopCh

	// If this is a standalone modifier (Shift/Option/Cmd/Ctrl), use session modifier key polling.
	// For normal characters (e.g. '\'), CGEventSourceKeyState is unreliable when another app is focused;
	// instead rely on OS KeyUp and Carbon release events, with a safety timeout.
	if d.isModAlone == 1 {
		go func(targetKC int, ch chan struct{}) {
			ticker := time.NewTicker(60 * time.Millisecond)
			defer ticker.Stop()

			for {
				select {
				case <-ch:
					return
				case <-ticker.C:
					d.mu.Lock()
					isDown := d.isDown
					kh := d.keyHandler
					running := d.running
					d.mu.Unlock()

					if !running {
						return
					}

					if isDown {
						isPhysicallyPressed := isPhysicalModifierDown(targetKC)
						if !isPhysicallyPressed {
							d.mu.Lock()
							if d.isDown {
								d.isDown = false
								d.stopWatchdogLocked()
								d.mu.Unlock()

								if kh != nil {
									d.enqueueEvent(func() { kh("up") })
								}
							} else {
								d.mu.Unlock()
							}
						}
					}
				}
			}
		}(d.targetKC, stopCh)
	} else {
		// Safety release watchdog for regular keys (generous 10-minute maximum hold timeout to prevent stuck state while allowing long dictation)
		go func(ch chan struct{}) {
			timer := time.NewTimer(10 * time.Minute)
			defer timer.Stop()

			select {
			case <-ch:
				return
			case <-timer.C:
				d.mu.Lock()
				isDown := d.isDown
				kh := d.keyHandler
				if isDown {
					d.isDown = false
					d.stopWatchdogLocked()
					d.mu.Unlock()

					if kh != nil {
						d.enqueueEvent(func() { kh("up") })
					}
				} else {
					d.mu.Unlock()
				}
			}
		}(stopCh)
	}
}

func (d *darwinHotkeyManager) stopWatchdogLocked() {
	if d.stopWatchdog != nil {
		close(d.stopWatchdog)
		d.stopWatchdog = nil
	}
}

func isPhysicalModifierDown(targetKC int) bool {
	return int(C.isPhysicalKeyDown(C.int64_t(targetKC))) != 0
}

func newPlatformManager() platformManager {
	return &darwinHotkeyManager{}
}

//export triggerDarwinKeyEvent
func triggerDarwinKeyEvent(keycode int64, eventType int64, flags uint64) {
	darwinRegistryMu.RLock()
	managers := make([]*darwinHotkeyManager, 0, len(darwinRegistry))
	for m := range darwinRegistry {
		managers = append(managers, m)
	}
	darwinRegistryMu.RUnlock()

	kc := int(keycode)
	evType := int(eventType)

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
			if evType == cgEventFlagsChanged && kc == targetKC {
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
						mgr.startWatchdogLocked()
						mgr.mu.Unlock()

						if kh != nil {
							mgr.enqueueEvent(func() { kh("down") })
						}
						if h != nil {
							now := time.Now().UnixMilli()
							if now-atomic.LoadInt64(&mgr.lastTrigger) > 150 {
								atomic.StoreInt64(&mgr.lastTrigger, now)
								mgr.enqueueEvent(func() { h() })
							}
						}
					}
				} else {
					if isCurrentlyDown {
						mgr.mu.Lock()
						mgr.isDown = false
						mgr.stopWatchdogLocked()
						mgr.mu.Unlock()

						if kh != nil {
							mgr.enqueueEvent(func() { kh("up") })
						}
					}
				}
			}
		} else {
			// Regular key or modifier combination (e.g. '\', '/', 'Cmd+Shift+S', 'Alt+Space')
			if kc == targetKC {
				if evType == cgEventKeyDown {
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
							mgr.startWatchdogLocked()
							mgr.mu.Unlock()

							if kh != nil {
								mgr.enqueueEvent(func() { kh("down") })
							}
							if h != nil {
								now := time.Now().UnixMilli()
								if now-atomic.LoadInt64(&mgr.lastTrigger) > 150 {
									atomic.StoreInt64(&mgr.lastTrigger, now)
									mgr.enqueueEvent(func() { h() })
								}
							}
						}
					}
				} else if evType == cgEventKeyUp {
					if isCurrentlyDown {
						mgr.mu.Lock()
						mgr.isDown = false
						mgr.stopWatchdogLocked()
						mgr.mu.Unlock()

						if kh != nil {
							mgr.enqueueEvent(func() { kh("up") })
						}
					}
				}
			}
		}
	}
}

const (
	carbonCmdKey     = 0x0100
	carbonShiftKey   = 0x0200
	carbonOptionKey  = 0x0800
	carbonControlKey = 0x1000
)

func parseShortcutToCarbon(shortcutStr string) (uint32, uint32, bool) {
	s := strings.TrimSpace(shortcutStr)
	if s == "" {
		return 0, 0, false
	}
	parts := strings.Split(s, "+")
	var mods uint32
	var keyPart string

	for _, p := range parts {
		clean := strings.ToLower(strings.TrimSpace(p))
		switch clean {
		case "cmd", "command", "cmdorctrl", "meta":
			mods |= carbonCmdKey
		case "ctrl", "control":
			mods |= carbonControlKey
		case "shift":
			mods |= carbonShiftKey
		case "alt", "option", "opt":
			mods |= carbonOptionKey
		default:
			keyPart = clean
		}
	}

	if keyPart == "" {
		if mods&carbonShiftKey != 0 {
			return uint32(kVK_RightShift), 0, true
		}
		if mods&carbonOptionKey != 0 {
			return uint32(kVK_RightOption), 0, true
		}
		if mods&carbonControlKey != 0 {
			return uint32(kVK_RightControl), 0, true
		}
		if mods&carbonCmdKey != 0 {
			return uint32(kVK_RightCommand), 0, true
		}
		return 0, 0, false
	}

	kc, ok := darwinKeyMap[keyPart]
	if !ok {
		return 0, 0, false
	}
	return uint32(kc), mods, true
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

	// Register with macOS Carbon EventHotKey subsystem as complementary background listener
	if carbKC, carbMods, ok := parseShortcutToCarbon(shortcutStr); ok {
		carbonMu.Lock()
		cID := nextCarbonID
		nextCarbonID++
		carbonManagers[cID] = d
		carbonMu.Unlock()

		d.carbonID = cID
		d.carbonRef = C.registerCarbonKeyRef(C.uint(cID), C.uint(carbKC), C.uint(carbMods))
	}
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
	if shouldStartGlobal && darwinReadyChan == nil {
		darwinReadyChan = make(chan struct{})
	}
	darwinRegistryMu.Unlock()

	if shouldStartGlobal {
		go func() {
			runtime.LockOSThread()
			defer runtime.UnlockOSThread()

			for attempt := 1; attempt <= 10; attempt++ {
				darwinRegistryMu.Lock()
				count := len(darwinRegistry)
				alreadyRunning := darwinRunning
				darwinRegistryMu.Unlock()
				if count == 0 || alreadyRunning {
					return
				}

				ok := int(C.startEventTap())
				if ok != 0 {
					return
				}

				// If Accessibility permission is pending user approval, retry every 2 seconds
				time.Sleep(2 * time.Second)
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

	if d.carbonRef != nil {
		C.unregisterCarbonKeyRef(d.carbonRef)
		d.carbonRef = nil
	}
	if d.carbonID != 0 {
		carbonMu.Lock()
		delete(carbonManagers, d.carbonID)
		carbonMu.Unlock()
		d.carbonID = 0
	}

	if carbKC, carbMods, ok := parseShortcutToCarbon(shortcutStr); ok {
		carbonMu.Lock()
		cID := nextCarbonID
		nextCarbonID++
		carbonManagers[cID] = d
		carbonMu.Unlock()

		d.carbonID = cID
		d.carbonRef = C.registerCarbonKeyRef(C.uint(cID), C.uint(carbKC), C.uint(carbMods))
	}
	d.mu.Unlock()

	// If global event tap is not yet running (e.g. permission was granted after startup), attempt initialization
	darwinRegistryMu.Lock()
	isRunning := darwinRunning
	if !isRunning && darwinReadyChan == nil {
		darwinReadyChan = make(chan struct{})
	}
	darwinRegistryMu.Unlock()
	if !isRunning {
		go func() {
			runtime.LockOSThread()
			defer runtime.UnlockOSThread()
			_ = int(C.startEventTap())
		}()
	}

	return nil
}

func (d *darwinHotkeyManager) stop() {
	d.mu.Lock()
	d.running = false
	d.isDown = false
	d.stopWatchdogLocked()
	if d.stopWorker != nil {
		close(d.stopWorker)
		d.stopWorker = nil
	}
	if d.carbonRef != nil {
		C.unregisterCarbonKeyRef(d.carbonRef)
		d.carbonRef = nil
	}
	if d.carbonID != 0 {
		carbonMu.Lock()
		delete(carbonManagers, d.carbonID)
		carbonMu.Unlock()
		d.carbonID = 0
	}
	d.mu.Unlock()

	darwinRegistryMu.Lock()
	delete(darwinRegistry, d)
	shouldStopGlobal := len(darwinRegistry) == 0 && darwinRunning
	if shouldStopGlobal {
		darwinRunning = false
		darwinReadyChan = nil
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
	"3": 20, "4": 21, "6": 22, "5": 23, "=": 24, "equal": 24, "plus": 24, "9": 25, "7": 26, "-": 27, "minus": 27, "8": 28,
	"0": 29, "]": 30, "bracketright": 30, "o": 31, "u": 32, "[": 33, "bracketleft": 33, "i": 34, "p": 35, "l": 37, "j": 38,
	"'": 39, "quote": 39, "k": 40, ";": 41, "semicolon": 41, "\\": 42, "backslash": 42, ",": 43, "comma": 43, "/": 44, "slash": 44, "n": 45, "m": 46, ".": 47, "period": 47,
	"`": 50, "grave": 50, "backquote": 50, "tilde": 50,
	"space": kVK_Space, "return": kVK_Return, "enter": kVK_Return, "tab": kVK_Tab,
	"delete": kVK_Delete, "del": kVK_Delete, "backspace": kVK_Delete, "escape": kVK_Escape, "esc": kVK_Escape,
	"f1": 122, "f2": 120, "f3": 99, "f4": 118, "f5": 96, "f6": 97, "f7": 98, "f8": 100, "f9": 101, "f10": 109, "f11": 103, "f12": 111,
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
