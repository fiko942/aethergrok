//go:build windows

package system

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// MuteSystemVolume mutes the Windows audio master endpoint using PowerShell / Audio endpoint if available,
// returning the captured volume state.
func MuteSystemVolume() (SystemVolumeState, error) {
	// Attempt query via PowerShell audio device script
	queryScript := `
try {
	Add-Type -TypeDefinition @'
using System.Runtime.InteropServices;
[Guid("5CDF2C82-841E-4546-9722-0CF74078229A"), InterfaceType(ComInterfaceType.InterfaceIsIUnknown)]
public interface IAudioEndpointVolume {
	int f(); int g(); int h(); int j();
	int SetMasterVolumeLevelScalar(float fLevel, System.Guid pguidEventContext);
	int k();
	int GetMasterVolumeLevelScalar(out float pfLevel);
	int SetMute([MarshalAs(UnmanagedType.Bool)] bool bMute, System.Guid pguidEventContext);
	int GetMute(out bool pbMute);
}
[Guid("D666063F-1587-4E43-81F1-B948E807363F"), InterfaceType(ComInterfaceType.InterfaceIsIUnknown)]
public interface IMMDevice {
	int Activate(ref System.Guid id, int clsCtx, int activationParams, out IAudioEndpointVolume aev);
}
[Guid("A95664D2-9614-4F35-A746-DE8DB63617E6"), InterfaceType(ComInterfaceType.InterfaceIsIUnknown)]
public interface IMMDeviceEnumerator {
	int GetDefaultAudioEndpoint(int dataFlow, int role, out IMMDevice endpoint);
}
[ComImport, Guid("BCDE0395-E52F-467C-8E3D-C4579291692E")]
public class MMDeviceEnumeratorComObject {}
public class Audio {
	public static IAudioEndpointVolume GetMasterVolume() {
		var enumerator = new MMDeviceEnumeratorComObject() as IMMDeviceEnumerator;
		IMMDevice dev = null;
		enumerator.GetDefaultAudioEndpoint(0, 1, out dev);
		var IID_IAudioEndpointVolume = typeof(IAudioEndpointVolume).GUID;
		IAudioEndpointVolume epv = null;
		dev.Activate(ref IID_IAudioEndpointVolume, 23, 0, out epv);
		return epv;
	}
}
'@
	$vol = [Audio]::GetMasterVolume()
	$level = 0.0
	$muted = $false
	$vol.GetMasterVolumeLevelScalar([ref]$level) | Out-Null
	$vol.GetMute([ref]$muted) | Out-Null
	$volInt = [math]::Round($level * 100)
	Write-Output "$volInt,$muted"
	$vol.SetMute($true, [System.Guid]::Empty) | Out-Null
} catch {
	Write-Output "100,False"
}
`
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", queryScript)
	out, err := cmd.Output()
	if err != nil {
		// Fallback graceful return
		return SystemVolumeState{OriginalVolume: 100, WasMuted: false}, nil
	}

	trimmed := strings.TrimSpace(string(out))
	lines := strings.Split(trimmed, "\n")
	lastLine := strings.TrimSpace(lines[len(lines)-1])
	parts := strings.Split(lastLine, ",")
	if len(parts) >= 2 {
		volInt, errVol := strconv.Atoi(strings.TrimSpace(parts[0]))
		mutedBool, errMuted := strconv.ParseBool(strings.TrimSpace(parts[1]))
		if errVol == nil && errMuted == nil {
			return SystemVolumeState{
				OriginalVolume: volInt,
				WasMuted:       mutedBool,
			}, nil
		}
	}

	return SystemVolumeState{OriginalVolume: 100, WasMuted: false}, nil
}

// RestoreSystemVolume restores audio volume and mute state on Windows.
func RestoreSystemVolume(state SystemVolumeState) error {
	restoreScript := fmt.Sprintf(`
try {
	Add-Type -TypeDefinition @'
using System.Runtime.InteropServices;
[Guid("5CDF2C82-841E-4546-9722-0CF74078229A"), InterfaceType(ComInterfaceType.InterfaceIsIUnknown)]
public interface IAudioEndpointVolume {
	int f(); int g(); int h(); int j();
	int SetMasterVolumeLevelScalar(float fLevel, System.Guid pguidEventContext);
	int k();
	int GetMasterVolumeLevelScalar(out float pfLevel);
	int SetMute([MarshalAs(UnmanagedType.Bool)] bool bMute, System.Guid pguidEventContext);
	int GetMute(out bool pbMute);
}
[Guid("D666063F-1587-4E43-81F1-B948E807363F"), InterfaceType(ComInterfaceType.InterfaceIsIUnknown)]
public interface IMMDevice {
	int Activate(ref System.Guid id, int clsCtx, int activationParams, out IAudioEndpointVolume aev);
}
[Guid("A95664D2-9614-4F35-A746-DE8DB63617E6"), InterfaceType(ComInterfaceType.InterfaceIsIUnknown)]
public interface IMMDeviceEnumerator {
	int GetDefaultAudioEndpoint(int dataFlow, int role, out IMMDevice endpoint);
}
[ComImport, Guid("BCDE0395-E52F-467C-8E3D-C4579291692E")]
public class MMDeviceEnumeratorComObject {}
public class Audio {
	public static IAudioEndpointVolume GetMasterVolume() {
		var enumerator = new MMDeviceEnumeratorComObject() as IMMDeviceEnumerator;
		IMMDevice dev = null;
		enumerator.GetDefaultAudioEndpoint(0, 1, out dev);
		var IID_IAudioEndpointVolume = typeof(IAudioEndpointVolume).GUID;
		IAudioEndpointVolume epv = null;
		dev.Activate(ref IID_IAudioEndpointVolume, 23, 0, out epv);
		return epv;
	}
}
'@
	$vol = [Audio]::GetMasterVolume()
	$scalar = [float]%f
	$mute = $%t
	$vol.SetMasterVolumeLevelScalar($scalar, [System.Guid]::Empty) | Out-Null
	$vol.SetMute($mute, [System.Guid]::Empty) | Out-Null
} catch {}
`, float64(state.OriginalVolume)/100.0, state.WasMuted)

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", restoreScript)
	return cmd.Run()
}
