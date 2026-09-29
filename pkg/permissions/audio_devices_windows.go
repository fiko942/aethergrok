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
