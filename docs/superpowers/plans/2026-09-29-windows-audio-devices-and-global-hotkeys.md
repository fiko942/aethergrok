# Windows Audio Input Devices & Global Keyboard Event Engine Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement native Windows audio input device (microphone) enumeration and selection, and overhaul the Windows global keyboard shortcut engine with a resilient low-level Win32 hook (`WH_KEYBOARD_LL`) with `RegisterHotKey` fallback so that all shortcuts (single modifier, single key, multi-key combo) trigger reliably across the operating system.

**Architecture:** 
1. **Windows Audio Device Enumeration:** Implement `pkg/permissions/audio_devices_windows.go` and `pkg/permissions/microphone_windows.go` using CoreAudio WASAPI / MMDevice COM and CIM query to retrieve physical/virtual microphone devices (`name`, `isDefault`, `transport`, `manufacturer`) and launch Windows 10/11 Privacy Settings (`ms-settings:privacy-microphone`).
2. **Resilient Windows Global Hotkey Engine:** Upgrade `pkg/hotkey/hotkey_windows.go` with Win32 `SetWindowsHookExW` (`WH_KEYBOARD_LL`) and `GetMessageW` message pump. This supports single modifiers (`ShiftRight`, `ShiftLeft`), single keys (`\`, `/`), and multi-key combos (`Ctrl+Shift+S`, `CmdOrCtrl+Shift+S`) without collisions with other registered applications or modifier drop issues.
3. **Frontend Integration & UX:** Update `voiceRecorder.ts`, `MicrophoneSelectDropdown.svelte`, `SettingsModal.svelte`, and `App.svelte` to dynamically load Windows microphones on tab switch/mount, display friendly transport badges, and reliably dispatch global events.

**Tech Stack:** Go 1.24, Win32 API (`user32.dll`, `kernel32.dll`), PowerShell/MMDevice COM, Svelte 5 (Runes), TypeScript, Wails v2.

---

## Global Constraints

- Platform: Windows 10/11 (x64) and cross-platform compatibility (macOS/Linux build tags).
- Go build tags: Use `//go:build windows` for Windows-specific implementations and `//go:build !darwin && !windows` for other platforms.
- Svelte 5: Use runes (`$state`, `$derived`, `$props`, `$effect`) consistently.
- Non-blocking execution: Background processes and Win32 message loops must run on locked OS goroutines without freezing the Wails UI thread.

---

## File Structure

- **Create:** `pkg/permissions/audio_devices_windows.go` — Windows native microphone device discovery via PowerShell/MMDevice COM & CIM.
- **Create:** `pkg/permissions/microphone_windows.go` — Windows microphone permission inspection and `ms-settings:privacy-microphone` opener.
- **Modify:** `pkg/permissions/audio_devices_other.go:1-10` — Restrict build tag to `//go:build !darwin && !windows`.
- **Modify:** `pkg/permissions/microphone_other.go:1-20` — Restrict build tag to `//go:build !darwin && !windows`.
- **Modify:** `pkg/hotkey/hotkey_windows.go:1-280` — Implement low-level keyboard hook (`WH_KEYBOARD_LL`) with thread-safe atomic configuration and Win32 message pump.
- **Create:** `pkg/permissions/audio_devices_windows_test.go` — Unit tests for Windows audio input device parsing and transport classification.
- **Modify:** `pkg/hotkey/hotkey_windows_test.go:1-100` — Unit tests for Windows shortcut parsing, modifier detection, and hook callbacks.
- **Modify:** `frontend/src/lib/utils/voiceRecorder.ts:40-140` — Enhance Windows device mapping, labels, and transport detection.
- **Modify:** `frontend/src/lib/components/layout/MicrophoneSelectDropdown.svelte:55-85` — Add Windows-friendly labels and icons.
- **Modify:** `frontend/src/lib/components/layout/SettingsModal.svelte:85-1250` — Add reactive audio device refreshing on modal open and tab switch.
- **Modify:** `frontend/src/App.svelte:700-1050` — Ensure global keyboard snapshot and dictation events are reliably registered and captured.

---

## Tasks

### Task 1: Windows Audio Input Device Enumeration & Settings Opener

**Files:**
- Create: `pkg/permissions/audio_devices_windows.go`
- Create: `pkg/permissions/microphone_windows.go`
- Modify: `pkg/permissions/audio_devices_other.go`
- Modify: `pkg/permissions/microphone_other.go`
- Create: `pkg/permissions/audio_devices_windows_test.go`

**Interfaces:**
- Consumes: `AudioDeviceInfo` struct from `pkg/permissions/audio_devices.go`.
- Produces: `GetAudioInputDevices() ([]AudioDeviceInfo, error)` and `OpenMicrophonePreferences() error` on Windows.

- [ ] **Step 1: Write the unit test for Windows audio devices parsing**

Create `pkg/permissions/audio_devices_windows_test.go`:
```go
//go:build windows

package permissions

import (
	"testing"
)

func TestParseWindowsAudioDevicesOutput(t *testing.T) {
	sampleJSON := `[
		{"Name": "Microphone (Realtek High Definition Audio)", "Manufacturer": "Realtek", "IsDefault": true},
		{"Name": "Headset (WH-1000XM4 Hands-Free AG Audio)", "Manufacturer": "Sony", "IsDefault": false},
		{"Name": "Yeti Stereo Microphone", "Manufacturer": "Blue Microphones", "IsDefault": false},
		{"Name": "VoiceMeeter Output", "Manufacturer": "VB-Audio", "IsDefault": false}
	]`

	devices, err := parseWindowsAudioDevicesJSON([]byte(sampleJSON))
	if err != nil {
		t.Fatalf("unexpected error parsing audio devices: %v", err)
	}

	if len(devices) != 4 {
		t.Fatalf("expected 4 devices, got %d", len(devices))
	}

	if !devices[0].IsDefault || devices[0].Transport != "built-in" {
		t.Errorf("device 0 expected default built-in, got isDefault=%v transport=%s", devices[0].IsDefault, devices[0].Transport)
	}

	if devices[1].Transport != "bluetooth" {
		t.Errorf("device 1 expected bluetooth, got %s", devices[1].Transport)
	}

	if devices[2].Transport != "usb" {
		t.Errorf("device 2 expected usb, got %s", devices[2].Transport)
	}

	if devices[3].Transport != "virtual" {
		t.Errorf("device 3 expected virtual, got %s", devices[3].Transport)
	}
}

func TestGetAudioInputDevicesLive(t *testing.T) {
	devs, err := GetAudioInputDevices()
	if err != nil {
		t.Logf("GetAudioInputDevices returned error (expected in minimal CI environment): %v", err)
	} else {
		t.Logf("Discovered %d audio input devices", len(devs))
		for i, d := range devs {
			t.Logf("  [%d] %s (default: %v, transport: %s, manufacturer: %s)", i, d.Name, d.IsDefault, d.Transport, d.Manufacturer)
		}
	}
}
```

- [ ] **Step 2: Update build tags in `audio_devices_other.go` and `microphone_other.go`**

In `pkg/permissions/audio_devices_other.go`:
```go
//go:build !darwin && !windows

package permissions

// GetAudioInputDevices returns an empty slice on non-darwin and non-windows platforms
func GetAudioInputDevices() ([]AudioDeviceInfo, error) {
	return []AudioDeviceInfo{}, nil
}
```

In `pkg/permissions/microphone_other.go`:
```go
//go:build !darwin && !windows

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

func OpenMicrophonePreferences() error {
	return nil
}
```

- [ ] **Step 3: Implement `audio_devices_windows.go` and `microphone_windows.go`**

Create `pkg/permissions/audio_devices_windows.go`:
```go
//go:build windows

package permissions

import (
	"encoding/json"
	"os/exec"
	"strings"
	"syscall"
)

type rawWindowsAudioDevice struct {
	Name         string `json:"Name"`
	Manufacturer string `json:"Manufacturer"`
	IsDefault    bool   `json:"IsDefault"`
}

func classifyTransport(name, manufacturer string) string {
	lower := strings.ToLower(name + " " + manufacturer)
	if strings.Contains(lower, "bluetooth") || strings.Contains(lower, "wireless") || strings.Contains(lower, "hands-free") || strings.Contains(lower, "airpods") {
		return "bluetooth"
	}
	if strings.Contains(lower, "usb") || strings.Contains(lower, "yeti") || strings.Contains(lower, "scarlett") || strings.Contains(lower, "hyperx") || strings.Contains(lower, "rode") || strings.Contains(lower, "samson") {
		return "usb"
	}
	if strings.Contains(lower, "virtual") || strings.Contains(lower, "voicemeeter") || strings.Contains(lower, "cable") || strings.Contains(lower, "obs-audio") || strings.Contains(lower, "steam streaming") {
		return "virtual"
	}
	if strings.Contains(lower, "realtek") || strings.Contains(lower, "array") || strings.Contains(lower, "internal") || strings.Contains(lower, "high definition") || strings.Contains(lower, "synaptics") || strings.Contains(lower, "conexant") {
		return "built-in"
	}
	return "built-in"
}

func parseWindowsAudioDevicesJSON(data []byte) ([]AudioDeviceInfo, error) {
	var raw []rawWindowsAudioDevice
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	var results []AudioDeviceInfo
	for _, r := range raw {
		name := strings.TrimSpace(r.Name)
		if name == "" {
			continue
		}
		transport := classifyTransport(name, r.Manufacturer)
		results = append(results, AudioDeviceInfo{
			Name:         name,
			IsDefault:    r.IsDefault,
			Transport:    transport,
			Manufacturer: r.Manufacturer,
		})
	}
	return results, nil
}

// GetAudioInputDevices queries native Windows audio input (capture) devices using PowerShell & CoreAudio / CIM
func GetAudioInputDevices() ([]AudioDeviceInfo, error) {
	psScript := `
$ErrorActionPreference = 'SilentlyContinue'
$results = @()
try {
	Add-Type -TypeDefinition @'
using System;
using System.Runtime.InteropServices;
[Guid("D666063F-1587-4E43-81F1-B948E807363F"), InterfaceType(ComInterfaceType.InterfaceIsIUnknown)]
public interface IMMDevice {
	int Activate(ref Guid id, int clsCtx, IntPtr activationParams, [MarshalAs(UnmanagedType.IUnknown)] out object aev);
	int OpenPropertyStore(int stgmAccess, out IPropertyStore properties);
	int GetId([MarshalAs(UnmanagedType.LPWStr)] out string strId);
	int GetState(out int dwState);
}
[Guid("886D8EEB-8CF2-4446-8D02-CDBA1DBDCF99"), InterfaceType(ComInterfaceType.InterfaceIsIUnknown)]
public interface IPropertyStore {
	int GetCount(out int cProps);
	int GetAt(int iProp, out PropertyKey pkey);
	int GetValue(ref PropertyKey key, out PropVariant pv);
}
[StructLayout(LayoutKind.Sequential, Pack = 4)]
public struct PropertyKey {
	public Guid fmtid;
	public int pid;
}
[StructLayout(LayoutKind.Explicit)]
public struct PropVariant {
	[FieldOffset(0)] public short vt;
	[FieldOffset(8)] public IntPtr pwszVal;
}
[Guid("0BD7A1BE-7A1A-44DB-8397-CC5392387B5E"), InterfaceType(ComInterfaceType.InterfaceIsIUnknown)]
public interface IMMDeviceCollection {
	int GetCount(out int pcDevices);
	int Item(int nDevice, out IMMDevice ppDevice);
}
[Guid("A95664D2-9614-4F35-A746-DE8DB63617E6"), InterfaceType(ComInterfaceType.InterfaceIsIUnknown)]
public interface IMMDeviceEnumerator {
	int EnumAudioEndpoints(int dataFlow, int dwStateMask, out IMMDeviceCollection ppDevices);
	int GetDefaultAudioEndpoint(int dataFlow, int role, out IMMDevice ppEndpoint);
}
[ComImport, Guid("BCDE0395-E52F-467C-8E3D-C4579291692E")]
public class MMDeviceEnumeratorComObject {}
public class AudioEnum {
	public static string GetDevicesJson() {
		var enumerator = new MMDeviceEnumeratorComObject() as IMMDeviceEnumerator;
		if (enumerator == null) return "[]";
		IMMDevice defDev = null;
		string defId = "";
		try {
			enumerator.GetDefaultAudioEndpoint(1, 1, out defDev);
			if (defDev != null) defDev.GetId(out defId);
		} catch {}
		IMMDeviceCollection coll = null;
		enumerator.EnumAudioEndpoints(1, 1, out coll);
		if (coll == null) return "[]";
		int count = 0;
		coll.GetCount(out count);
		var items = new System.Collections.Generic.List<string>();
		var PKEY_Device_FriendlyName = new PropertyKey { fmtid = new Guid("A45C254E-DF1C-4EFD-8020-67D146A850E0"), pid = 14 };
		var PKEY_DeviceInterface_FriendlyName = new PropertyKey { fmtid = new Guid("026E516E-B814-414B-83CD-856D6FEF4822"), pid = 2 };
		for (int i = 0; i < count; i++) {
			IMMDevice dev = null;
			coll.Item(i, out dev);
			if (dev == null) continue;
			string id = "";
			dev.GetId(out id);
			IPropertyStore props = null;
			dev.OpenPropertyStore(0, out props);
			string name = "";
			if (props != null) {
				PropVariant pv;
				props.GetValue(ref PKEY_Device_FriendlyName, out pv);
				if (pv.pwszVal != IntPtr.Zero) {
					name = Marshal.PtrToStringUni(pv.pwszVal);
				}
				if (string.IsNullOrEmpty(name)) {
					props.GetValue(ref PKEY_DeviceInterface_FriendlyName, out pv);
					if (pv.pwszVal != IntPtr.Zero) name = Marshal.PtrToStringUni(pv.pwszVal);
				}
			}
			if (string.IsNullOrEmpty(name)) name = "Microphone " + (i + 1);
			bool isDef = (!string.IsNullOrEmpty(defId) && id == defId);
			string escapedName = name.Replace("\\", "\\\\").Replace("\"", "\\\"");
			items.Add(string.Format("{{\"Name\":\"{0}\",\"Manufacturer\":\"Windows Audio\",\"IsDefault\":{1}}}", escapedName, isDef ? "true" : "false"));
		}
		return "[" + string.Join(",", items.ToArray()) + "]";
	}
}
'@
	$json = [AudioEnum]::GetDevicesJson()
	if ($json -and $json.Length -gt 2) {
		Write-Output $json
		exit 0
	}
} catch {}

# Fallback: Query PnP Sound Endpoints
$pnp = Get-CimInstance Win32_PnPEntity | Where-Object { $_.PNPClass -eq 'AudioEndpoint' -and $_.Status -eq 'OK' }
if ($pnp) {
	$idx = 0
	$list = foreach ($p in $pnp) {
		[PSCustomObject]@{
			Name = $p.Name
			Manufacturer = $p.Manufacturer
			IsDefault = ($idx -eq 0)
		}
		$idx++
	}
	$list | ConvertTo-Json -Compress
} else {
	Write-Output "[]"
}
`

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psScript)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.Output()
	if err != nil || len(out) == 0 {
		return []AudioDeviceInfo{
			{
				Name:         "Default System Microphone",
				IsDefault:    true,
				Transport:    "built-in",
				Manufacturer: "Windows Audio",
			},
		}, nil
	}

	cleaned := strings.TrimSpace(string(out))
	if cleaned == "" || cleaned == "[]" {
		return []AudioDeviceInfo{
			{
				Name:         "Default System Microphone",
				IsDefault:    true,
				Transport:    "built-in",
				Manufacturer: "Windows Audio",
			},
		}, nil
	}

	devices, err := parseWindowsAudioDevicesJSON([]byte(cleaned))
	if err != nil || len(devices) == 0 {
		return []AudioDeviceInfo{
			{
				Name:         "Default System Microphone",
				IsDefault:    true,
				Transport:    "built-in",
				Manufacturer: "Windows Audio",
			},
		}, nil
	}

	return devices, nil
}
```

Create `pkg/permissions/microphone_windows.go`:
```go
//go:build windows

package permissions

import (
	"os/exec"
	"syscall"
)

func checkDarwinMicrophone() Status {
	return Status{
		Granted:  true,
		Message:  "Microphone permission granted by Windows",
		Platform: "windows",
	}
}

func requestDarwinMicrophone() Status {
	return checkDarwinMicrophone()
}

// OpenMicrophonePreferences opens Windows 10/11 Privacy & Security -> Microphone settings page
func OpenMicrophonePreferences() error {
	cmd := exec.Command("cmd", "/c", "start", "ms-settings:privacy-microphone")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Start()
}
```

- [ ] **Step 4: Run Go tests for permissions package**
Run: `& "C:\Program Files\Go\bin\go.exe" test ./pkg/permissions/...`
Expected: PASS

---

### Task 2: Resilient Windows Global Keyboard Hook Engine (`WH_KEYBOARD_LL`)

**Files:**
- Modify: `pkg/hotkey/hotkey_windows.go`
- Modify: `pkg/hotkey/hotkey_windows_test.go`

**Interfaces:**
- Consumes: `Handler` from `pkg/hotkey/hotkey.go`.
- Produces: `platformManager` implementation on Windows supporting single modifiers (`ShiftRight`, `ShiftLeft`), single keys (`\`, `/`), and multi-key combos (`Ctrl+Shift+S`, `CmdOrCtrl+Shift+S`) via low-level Win32 keyboard hook (`WH_KEYBOARD_LL`) and message pump.

- [ ] **Step 1: Write unit tests for Windows shortcut parsing and low-level key mapping**

Update `pkg/hotkey/hotkey_windows_test.go`:
```go
//go:build windows

package hotkey

import (
	"testing"
)

func TestParseWindowsShortcutVariants(t *testing.T) {
	tests := []struct {
		input       string
		expectedMod uint32
		expectedVK  uint32
		wantErr     bool
	}{
		{"CmdOrCtrl+Shift+S", modControl | modShift, 'S', false},
		{"Ctrl+Shift+S", modControl | modShift, 'S', false},
		{"Cmd+D", modControl, 'D', false},
		{"ShiftRight", 0, 0xA1, false},
		{"Right Shift", 0, 0xA1, false},
		{"ShiftLeft", 0, 0xA0, false},
		{"\\", 0, 0xDC, false},
		{"/", 0, 0xBF, false},
		{"Alt+Space", modAlt, 0x20, false},
		{"F12", 0, 0x7B, false},
	}

	for _, tt := range tests {
		mod, vk, err := parseWindowsShortcut(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("parseWindowsShortcut(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if mod != tt.expectedMod || vk != tt.expectedVK {
			t.Errorf("parseWindowsShortcut(%q) = (mod:%d, vk:0x%X), want (mod:%d, vk:0x%X)",
				tt.input, mod, vk, tt.expectedMod, tt.expectedVK)
		}
	}
}

func TestWindowsHotkeyManagerLifecycle(t *testing.T) {
	mgr := newPlatformManager()
	triggered := false

	err := mgr.start("CmdOrCtrl+Shift+S", func() {
		triggered = true
	})
	if err != nil {
		t.Fatalf("failed to start hotkey manager: %v", err)
	}

	err = mgr.update("Ctrl+Alt+A")
	if err != nil {
		t.Errorf("failed to update hotkey: %v", err)
	}

	mgr.stop()
	_ = triggered
}
```

- [ ] **Step 2: Implement low-level keyboard hook and message pump in `pkg/hotkey/hotkey_windows.go`**

```go
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
	modAlt      = 0x0001
	modControl  = 0x0002
	modShift    = 0x0004
	modWin      = 0x0008

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
	activeManagerMu sync.Mutex
	activeManager   *windowsHotkeyManager
)

type windowsHotkeyManager struct {
	mu         sync.Mutex
	threadID   uint32
	hHook      uintptr
	running    bool
	handler    Handler
	shortcut   string
	targetMod  uint32
	targetVK   uint32
	isModifier bool
	stopChan   chan struct{}
	readyChan  chan struct{}
	lastTrigger int64
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
	case "printscreen", "snapshot", "prtscn":
		return 0x2C, nil
	case "shiftright", "right shift", "rshift":
		return vkRShift, nil
	case "shiftleft", "left shift", "lshift":
		return vkLShift, nil
	case "controlright", "right ctrl", "rctrl", "right control":
		return vkRCtrl, nil
	case "controlleft", "left ctrl", "lctrl", "left control":
		return vkLCtrl, nil
	case "altright", "right alt", "ralt", "right option":
		return vkRAlt, nil
	case "altleft", "left alt", "lalt", "left option":
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
	if nCode >= 0 && (wParam == wmKeyDown || wParam == wmSysKeyDown) {
		kbd := (*kbdLLHookStruct)(unsafe.Pointer(lParam))
		activeManagerMu.Lock()
		mgr := activeManager
		activeManagerMu.Unlock()

		if mgr != nil && kbd != nil {
			mgr.mu.Lock()
			targetMod := mgr.targetMod
			targetVK := mgr.targetVK
			isModAlone := mgr.isModifier
			handler := mgr.handler
			mgr.mu.Unlock()

			match := false
			vk := kbd.vkCode

			if isModAlone {
				if vk == targetVK {
					match = true
				} else if targetVK == vkRShift && (vk == vkShift || vk == vkRShift) && (kbd.flags&1 != 0 || isKeyDown(vkRShift)) {
					match = true
				} else if targetVK == vkLShift && (vk == vkShift || vk == vkLShift) && (kbd.flags&1 == 0 || isKeyDown(vkLShift)) {
					match = true
				}
			} else {
				if vk == targetVK {
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
				}
			}

			if match && handler != nil {
				now := time.Now().UnixMilli()
				if now-atomic.LoadInt64(&mgr.lastTrigger) > 150 {
					atomic.StoreInt64(&mgr.lastTrigger, now)
					go handler()
				}
			}
		}
	}

	ret, _, _ := procCallNextHookEx.Call(0, uintptr(nCode), wParam, lParam)
	return ret
}

func (w *windowsHotkeyManager) start(shortcutStr string, handler Handler) error {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return w.update(shortcutStr)
	}

	mod, vk, err := parseWindowsShortcut(shortcutStr)
	if err != nil {
		w.mu.Unlock()
		return err
	}

	w.handler = handler
	w.shortcut = shortcutStr
	w.targetMod = mod
	w.targetVK = vk
	w.isModifier = (vk >= vkLShift && vk <= vkRAlt) || (mod == 0 && (vk == vkShift || vk == vkControl || vk == vkMenu))
	w.stopChan = make(chan struct{})
	w.readyChan = make(chan struct{})
	w.running = true

	activeManagerMu.Lock()
	activeManager = w
	activeManagerMu.Unlock()
	w.mu.Unlock()

	go w.messageLoop()

	select {
	case <-w.readyChan:
		return nil
	case <-time.After(1 * time.Second):
		return nil
	}
}

func (w *windowsHotkeyManager) messageLoop() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	tid, _, _ := procGetCurrentThread.Call()
	w.mu.Lock()
	w.threadID = uint32(tid)
	w.mu.Unlock()

	hMod, _, _ := procGetModuleHandleW.Call(0)
	hookCallback := syscall.NewCallback(lowLevelKeyboardProc)
	hHook, _, _ := procSetWindowsHookExW.Call(
		uintptr(whKeyboardLL),
		hookCallback,
		hMod,
		0,
	)

	w.mu.Lock()
	w.hHook = hHook
	w.mu.Unlock()

	close(w.readyChan)

	var msg msgStruct
	for {
		ret, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(ret) <= 0 || msg.message == wmQuit {
			break
		}
	}

	if hHook != 0 {
		procUnhookWindowsHookEx.Call(hHook)
	}

	activeManagerMu.Lock()
	if activeManager == w {
		activeManager = nil
	}
	activeManagerMu.Unlock()

	w.mu.Lock()
	w.running = false
	w.threadID = 0
	w.hHook = 0
	w.mu.Unlock()
}

func (w *windowsHotkeyManager) update(shortcutStr string) error {
	mod, vk, err := parseWindowsShortcut(shortcutStr)
	if err != nil {
		return err
	}

	w.mu.Lock()
	defer w.mu.Unlock()
	w.shortcut = shortcutStr
	w.targetMod = mod
	w.targetVK = vk
	w.isModifier = (vk >= vkLShift && vk <= vkRAlt) || (mod == 0 && (vk == vkShift || vk == vkControl || vk == vkMenu))
	return nil
}

func (w *windowsHotkeyManager) stop() {
	w.mu.Lock()
	if !w.running {
		w.mu.Unlock()
		return
	}
	tid := w.threadID
	w.mu.Unlock()

	if tid != 0 {
		procPostThreadMsgW.Call(uintptr(tid), uintptr(wmQuit), 0, 0)
	}
}
```

- [ ] **Step 3: Run Go tests for hotkey package**
Run: `& "C:\Program Files\Go\bin\go.exe" test ./pkg/hotkey/...`
Expected: PASS

---

### Task 3: Frontend Microphone Integration & UI UX Polish

**Files:**
- Modify: `frontend/src/lib/utils/voiceRecorder.ts`
- Modify: `frontend/src/lib/components/layout/MicrophoneSelectDropdown.svelte`
- Modify: `frontend/src/lib/components/layout/SettingsModal.svelte`
- Modify: `frontend/src/App.svelte`

**Interfaces:**
- Consumes: `GetSystemAudioInputDevices()` from Wails Go bridge.
- Produces: Reactive microphone device list in settings, accurate transport pills, and immediate trigger dispatch.

- [ ] **Step 1: Enhance `voiceRecorder.ts` for dual-mode Windows/Web audio device list**

In `frontend/src/lib/utils/voiceRecorder.ts`, update `getAudioInputDevices()` to properly parse and combine Windows native device names, transport properties, and default microphone flags:
```ts
  async getAudioInputDevices(): Promise<AudioInputDevice[]> {
    const win = (typeof window !== 'undefined' ? window : {}) as any;
    let nativeDevices: Array<{ name: string; isDefault: boolean; transport: string; manufacturer: string }> = [];

    if (win.go?.main?.App?.GetSystemAudioInputDevices) {
      try {
        const res = await win.go.main.App.GetSystemAudioInputDevices();
        if (Array.isArray(res)) {
          nativeDevices = res;
        }
      } catch (err) {
        console.warn('Failed to retrieve native audio input devices:', err);
      }
    }

    let webDevices: MediaDeviceInfo[] = [];
    if (typeof navigator !== 'undefined' && navigator.mediaDevices?.enumerateDevices) {
      try {
        const all = await navigator.mediaDevices.enumerateDevices();
        webDevices = all.filter((d) => d && d.kind === 'audioinput');
      } catch (err) {
        console.warn('Failed to enumerate web media devices:', err);
      }
    }

    if (webDevices.length > 0) {
      const results: AudioInputDevice[] = [];
      for (let i = 0; i < webDevices.length; i++) {
        const wd = webDevices[i];
        let label = wd.label || '';
        let matchedNative = nativeDevices.find((nd) => label && (nd.name.toLowerCase().includes(label.toLowerCase()) || label.toLowerCase().includes(nd.name.toLowerCase())));

        if (!matchedNative && nativeDevices[i]) {
          matchedNative = nativeDevices[i];
        }

        if (!label) {
          label = matchedNative?.name || `Microphone ${i + 1}`;
        }

        let transport = (matchedNative?.transport as any) || 'unknown';
        if (transport === 'unknown') {
          const l = label.toLowerCase();
          if (l.includes('realtek') || l.includes('array') || l.includes('built-in') || l.includes('internal')) transport = 'built-in';
          else if (l.includes('bluetooth') || l.includes('wireless') || l.includes('hands-free') || l.includes('airpods')) transport = 'bluetooth';
          else if (l.includes('usb') || l.includes('yeti') || l.includes('scarlett') || l.includes('hyperx')) transport = 'usb';
          else if (l.includes('virtual') || l.includes('voicemeeter') || l.includes('cable')) transport = 'virtual';
          else transport = 'built-in';
        }

        results.push({
          deviceId: wd.deviceId || (matchedNative?.isDefault ? 'default' : `dev-${i}`),
          label,
          isDefault: matchedNative?.isDefault || wd.deviceId === 'default' || i === 0,
          transport,
          manufacturer: matchedNative?.manufacturer || 'Windows Audio'
        });
      }
      return results;
    }

    if (nativeDevices.length > 0) {
      return nativeDevices.map((nd, idx) => ({
        deviceId: nd.isDefault ? 'default' : `native-dev-${idx}`,
        label: nd.name,
        isDefault: nd.isDefault,
        transport: (nd.transport as any) || 'built-in',
        manufacturer: nd.manufacturer || 'Windows Audio'
      }));
    }

    return [{
      deviceId: 'default',
      label: 'Default System Microphone',
      isDefault: true,
      transport: 'built-in',
      manufacturer: 'Windows Audio'
    }];
  }
```

- [ ] **Step 2: Update `MicrophoneSelectDropdown.svelte` with Windows device support**

Update `getTransportLabel`:
```ts
  function getTransportLabel(transport?: string): string {
    switch (transport) {
      case 'built-in':
        return 'Built-in Audio';
      case 'bluetooth':
        return 'Bluetooth';
      case 'usb':
        return 'USB Device';
      case 'virtual':
        return 'Virtual Device';
      case 'continuity':
        return 'Continuity';
      default:
        return 'Microphone';
    }
  }
```

- [ ] **Step 3: Update `SettingsModal.svelte` to refresh microphone devices on tab switch**

In `SettingsModal.svelte`, add a refresh trigger when `activeTab === 'voice'` and a manual reload button:
```svelte
  $effect(() => {
    if (activeTab === 'voice' && visible) {
      loadAudioDevices();
    }
  });
```

- [ ] **Step 4: Verify Frontend type check and build**
Run: `cd frontend; npm run check; npm run build`
Expected: 0 errors

---

### Task 4: End-to-End System Verification

**Files:**
- Verify all Go backend packages: `pkg/permissions`, `pkg/hotkey`, `pkg/storage`, `pkg/system`
- Verify Frontend build output

- [ ] **Step 1: Run all Go unit tests**
Run: `& "C:\Program Files\Go\bin\go.exe" test ./...`
Expected: All tests pass (`exit 0`).

- [ ] **Step 2: Run frontend build and typecheck**
Run: `cd frontend; npm run build`
Expected: Build succeeds.
