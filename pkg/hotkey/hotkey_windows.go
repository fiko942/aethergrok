//go:build windows

package hotkey

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"
)

const (
	modAlt     = 0x0001
	modControl = 0x0002
	modShift   = 0x0004
	modWin     = 0x0008

	whKeyboardLL = 13
	wmKeyDown    = 0x0100
	wmKeyUp      = 0x0101
	wmSysKeyDown = 0x0104
	wmSysKeyUp   = 0x0105
	wmQuit       = 0x0012

	vkShift   = 0x10
	vkControl = 0x11
	vkMenu    = 0x12 // Alt
	vkLWin    = 0x5B
	vkRWin    = 0x5C
	vkLShift  = 0xA0
	vkRShift  = 0xA1
	vkLCtrl   = 0xA2
	vkRCtrl   = 0xA3
	vkLAlt    = 0xA4
	vkRAlt    = 0xA5
)

var (
	user32                  = syscall.NewLazyDLL("user32.dll")
	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	procSetWindowsHookExW   = user32.NewProc("SetWindowsHookExW")
	procUnhookWindowsHookEx = user32.NewProc("UnhookWindowsHookEx")
	procCallNextHookEx      = user32.NewProc("CallNextHookEx")
	procGetMessageW         = user32.NewProc("GetMessageW")
	procPostThreadMsgW      = user32.NewProc("PostThreadMessageW")
	procGetAsyncKeyState     = user32.NewProc("GetAsyncKeyState")
	procGetCurrentThread    = kernel32.NewProc("GetCurrentThreadId")
	procGetModuleHandleW    = kernel32.NewProc("GetModuleHandleW")
)

type kbdLLHookStruct struct {
	vkCode      uint32
	scanCode    uint32
	flags       uint32
	time        uint32
	dwExtraInfo uintptr
}

type msgStruct struct {
	hwnd    uintptr
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pt      struct{ x, y int32 }
}

var (
	hookRegistryMu sync.RWMutex
	hookRegistry   = make(map[*windowsHotkeyManager]struct{})
	globalHook     uintptr
	globalThreadID uint32
	globalRunning  bool
	globalReady    chan struct{}
)

type windowsHotkeyManager struct {
	mu          sync.Mutex
	running     bool
	handler     Handler
	keyHandler  KeyHandler
	shortcut    string
	targetMod   uint32
	targetVK    uint32
	isModifier  bool
	lastTrigger int64
	events      chan func()
	stopWorker  chan struct{}
}

func newPlatformManager() platformManager {
	return &windowsHotkeyManager{}
}

func parseWindowsShortcut(s string) (uint32, uint32, error) {
	clean := strings.TrimSpace(s)
	if clean == "" {
		return 0, 0, fmt.Errorf("empty shortcut string")
	}

	parts := strings.Split(clean, "+")
	var mod uint32
	var keyStr string

	for i, part := range parts {
		p := strings.TrimSpace(strings.ToLower(part))
		if i == len(parts)-1 && !isModifier(p) {
			keyStr = p
			break
		}

		switch p {
		case "ctrl", "control", "cmdorctrl", "cmd", "command":
			mod |= modControl
		case "alt", "option", "opt":
			mod |= modAlt
		case "shift":
			mod |= modShift
		case "win", "super", "meta":
			mod |= modWin
		default:
			if i == len(parts)-1 {
				keyStr = p
			}
		}
	}

	if keyStr == "" {
		if mod&modShift != 0 {
			return 0, vkRShift, nil
		}
		if mod&modControl != 0 {
			return 0, vkRCtrl, nil
		}
		if mod&modAlt != 0 {
			return 0, vkRAlt, nil
		}
		return 0, 0, fmt.Errorf("no key specified in shortcut %q", s)
	}

	vk, err := mapKeyToVK(keyStr)
	if err != nil {
		return 0, 0, err
	}

	return mod, vk, nil
}

func isModifier(s string) bool {
	switch s {
	case "ctrl", "control", "cmdorctrl", "cmd", "command", "alt", "option", "opt", "shift", "win", "super", "meta":
		return true
	}
	return false
}

func mapKeyToVK(key string) (uint32, error) {
	k := strings.ToLower(strings.TrimSpace(key))
	if len(k) == 1 {
		ch := k[0]
		if ch >= 'a' && ch <= 'z' {
			return uint32(ch - 'a' + 'A'), nil
		}
		if ch >= '0' && ch <= '9' {
			return uint32(ch), nil
		}
		switch ch {
		case '\\':
			return 0xDC, nil // VK_OEM_5
		case '/':
			return 0xBF, nil // VK_OEM_2
		case '-':
			return 0xBD, nil // VK_OEM_MINUS
		case '=':
			return 0xBB, nil // VK_OEM_PLUS
		case '[':
			return 0xDB, nil // VK_OEM_4
		case ']':
			return 0xDD, nil // VK_OEM_6
		case ';':
			return 0xBA, nil // VK_OEM_1
		case '\'':
			return 0xDE, nil // VK_OEM_7
		case '`':
			return 0xC0, nil // VK_OEM_3
		case ',':
			return 0xBC, nil // VK_OEM_COMMA
		case '.':
			return 0xBE, nil // VK_OEM_PERIOD
		}
	}

	switch k {
	case "space":
		return 0x20, nil
	case "return", "enter":
		return 0x0D, nil
	case "escape", "esc":
		return 0x1B, nil
	case "tab":
		return 0x09, nil
	case "backspace":
		return 0x08, nil
	case "delete", "del":
		return 0x2E, nil // VK_DELETE
	case "insert", "ins":
		return 0x2D, nil
	case "home":
		return 0x24, nil
	case "end":
		return 0x23, nil
	case "pageup", "pgup":
		return 0x21, nil
	case "pagedown", "pgdn":
		return 0x22, nil
	case "left":
		return 0x25, nil
	case "up":
		return 0x26, nil
	case "right":
		return 0x27, nil
	case "down":
		return 0x28, nil
	case "capslock", "caps":
		return 0x14, nil
	case "equal", "plus":
		return 0xBB, nil
	case "minus":
		return 0xBD, nil
	case "grave", "backquote", "tilde":
		return 0xC0, nil
	case "bracketleft":
		return 0xDB, nil
	case "bracketright":
		return 0xDD, nil
	case "semicolon":
		return 0xBA, nil
	case "quote":
		return 0xDE, nil
	case "comma":
		return 0xBC, nil
	case "period":
		return 0xBE, nil
	case "printscreen", "snapshot", "prtscn":
		return 0x2C, nil
	case "shiftright", "right shift", "rightshift", "rshift":
		return vkRShift, nil
	case "shiftleft", "left shift", "leftshift", "lshift":
		return vkLShift, nil
	case "controlright", "right ctrl", "rightctrl", "rctrl", "right control":
		return vkRCtrl, nil
	case "controlleft", "left ctrl", "leftctrl", "lctrl", "left control":
		return vkLCtrl, nil
	case "altright", "right alt", "rightalt", "ralt", "right option":
		return vkRAlt, nil
	case "altleft", "left alt", "leftalt", "lalt", "left option":
		return vkLAlt, nil
	case "backslash":
		return 0xDC, nil
	case "slash":
		return 0xBF, nil
	}

	if strings.HasPrefix(k, "f") && len(k) <= 3 {
		if num, err := strconv.Atoi(k[1:]); err == nil && num >= 1 && num <= 24 {
			return uint32(0x70 + num - 1), nil
		}
	}

	return 0, fmt.Errorf("unsupported key %q", key)
}

func isKeyDown(vk int) bool {
	ret, _, _ := procGetAsyncKeyState.Call(uintptr(vk))
	return (ret & 0x8000) != 0
}

func lowLevelKeyboardProc(nCode int32, wParam uintptr, lParam uintptr) uintptr {
	if nCode >= 0 && lParam != 0 {
		isDown := (wParam == wmKeyDown || wParam == wmSysKeyDown)
		isUp := (wParam == wmKeyUp || wParam == wmSysKeyUp)

		if isDown || isUp {
			kbd := (*kbdLLHookStruct)(unsafe.Pointer(lParam))
			vk := kbd.vkCode
			flags := kbd.flags

			hookRegistryMu.RLock()
			managers := make([]*windowsHotkeyManager, 0, len(hookRegistry))
			for m := range hookRegistry {
				managers = append(managers, m)
			}
			hookRegistryMu.RUnlock()

			for _, mgr := range managers {
				mgr.mu.Lock()
				targetMod := mgr.targetMod
				targetVK := mgr.targetVK
				isModAlone := mgr.isModifier
				h := mgr.handler
				kh := mgr.keyHandler
				mgr.mu.Unlock()

				match := false
				if isModAlone {
					if vk == targetVK {
						match = true
					} else if targetVK == vkRShift && (vk == vkShift || vk == vkRShift) && (flags&1 != 0 || kbd.scanCode == 0x36 || isKeyDown(vkRShift)) {
						match = true
					} else if targetVK == vkLShift && (vk == vkShift || vk == vkLShift) && (flags&1 == 0 || kbd.scanCode == 0x2A || isKeyDown(vkLShift)) {
						match = true
					} else if targetVK == vkRCtrl && (vk == vkControl || vk == vkRCtrl) && (flags&1 != 0 || isKeyDown(vkRCtrl)) {
						match = true
					} else if targetVK == vkLCtrl && (vk == vkControl || vk == vkLCtrl) && (flags&1 == 0 || isKeyDown(vkLCtrl)) {
						match = true
					} else if targetVK == vkRAlt && (vk == vkMenu || vk == vkRAlt) && (flags&1 != 0 || isKeyDown(vkRAlt)) {
						match = true
					} else if targetVK == vkLAlt && (vk == vkMenu || vk == vkLAlt) && (flags&1 == 0 || isKeyDown(vkLAlt)) {
						match = true
					}
				} else {
					if vk == targetVK {
						if isDown {
							ctrlDown := isKeyDown(vkControl)
							shiftDown := isKeyDown(vkShift)
							altDown := isKeyDown(vkMenu)
							winDown := isKeyDown(vkLWin) || isKeyDown(vkRWin)

							ctrlReq := (targetMod & modControl) != 0
							shiftReq := (targetMod & modShift) != 0
							altReq := (targetMod & modAlt) != 0
							winReq := (targetMod & modWin) != 0

							if ctrlDown == ctrlReq && shiftDown == shiftReq && altDown == altReq && winDown == winReq {
								match = true
							}
						} else {
							// On key up of the primary trigger key
							match = true
						}
					}
				}

				if match {
					mgr.mu.Lock()
					eventChan := mgr.events
					mgr.mu.Unlock()

					if eventChan != nil {
						if isDown {
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
						} else if isUp {
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

	ret, _, _ := procCallNextHookEx.Call(0, uintptr(nCode), wParam, lParam)
	return ret
}

func startGlobalHookLocked() error {
	if globalRunning {
		return nil
	}

	globalRunning = true
	globalReady = make(chan struct{})

	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		tid, _, _ := procGetCurrentThread.Call()
		atomic.StoreUint32(&globalThreadID, uint32(tid))

		hMod, _, _ := procGetModuleHandleW.Call(0)
		hookCallback := syscall.NewCallback(lowLevelKeyboardProc)
		hHook, _, _ := procSetWindowsHookExW.Call(
			uintptr(whKeyboardLL),
			hookCallback,
			hMod,
			0,
		)

		if hHook == 0 {
			hookRegistryMu.Lock()
			globalRunning = false
			hookRegistryMu.Unlock()
			close(globalReady)
			return
		}

		globalHook = hHook
		close(globalReady)

		var msg msgStruct
		for {
			ret, _, _ := procGetMessageW.Call(
				uintptr(unsafe.Pointer(&msg)),
				0,
				0,
				0,
			)
			if int32(ret) <= 0 || msg.message == wmQuit {
				break
			}
		}

		if globalHook != 0 {
			procUnhookWindowsHookEx.Call(globalHook)
			globalHook = 0
		}
	}()

	select {
	case <-globalReady:
		if globalHook == 0 {
			return fmt.Errorf("failed to install WH_KEYBOARD_LL hook")
		}
		return nil
	case <-time.After(1 * time.Second):
		return nil
	}
}

func stopGlobalHookLocked() {
	if !globalRunning {
		return
	}
	globalRunning = false
	tid := atomic.LoadUint32(&globalThreadID)
	if tid != 0 {
		procPostThreadMsgW.Call(uintptr(tid), wmQuit, 0, 0)
	}
}

func (w *windowsHotkeyManager) start(shortcutStr string, handler Handler) error {
	return w.startWithKeyHandler(shortcutStr, handler, nil)
}

func (w *windowsHotkeyManager) startWithKeyHandler(shortcutStr string, handler Handler, keyHandler KeyHandler) error {
	mod, vk, err := parseWindowsShortcut(shortcutStr)
	if err != nil {
		return err
	}

	w.mu.Lock()
	w.handler = handler
	w.keyHandler = keyHandler
	w.shortcut = shortcutStr
	w.targetMod = mod
	w.targetVK = vk
	w.isModifier = (vk >= vkLShift && vk <= vkRAlt) || (mod == 0 && (vk == vkShift || vk == vkControl || vk == vkMenu))
	w.running = true
	w.events = make(chan func(), 64)
	w.stopWorker = make(chan struct{})
	eventChan := w.events
	stopChan := w.stopWorker
	w.mu.Unlock()

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

	hookRegistryMu.Lock()
	hookRegistry[w] = struct{}{}
	err = startGlobalHookLocked()
	hookRegistryMu.Unlock()

	return err
}

func (w *windowsHotkeyManager) update(shortcutStr string) error {
	mod, vk, err := parseWindowsShortcut(shortcutStr)
	if err != nil {
		return err
	}

	w.mu.Lock()
	w.shortcut = shortcutStr
	w.targetMod = mod
	w.targetVK = vk
	w.isModifier = (vk >= vkLShift && vk <= vkRAlt) || (mod == 0 && (vk == vkShift || vk == vkControl || vk == vkMenu))
	w.mu.Unlock()

	return nil
}

func (w *windowsHotkeyManager) stop() {
	w.mu.Lock()
	w.running = false
	if w.stopWorker != nil {
		close(w.stopWorker)
		w.stopWorker = nil
	}
	w.mu.Unlock()

	hookRegistryMu.Lock()
	delete(hookRegistry, w)
	if len(hookRegistry) == 0 {
		stopGlobalHookLocked()
	}
	hookRegistryMu.Unlock()
}
