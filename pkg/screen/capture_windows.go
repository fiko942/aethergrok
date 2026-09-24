//go:build windows

package screen

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// WindowsCapturer implements Windows native screen snapshot capture
type WindowsCapturer struct{}

// DefaultNativeCapturer returns the WindowsCapturer implementation for Windows
func DefaultNativeCapturer() NativeCapturer {
	return &WindowsCapturer{}
}

// buildPowerShellScript returns PowerShell script utilizing System.Drawing / GDI CopyFromScreen (BitBlt equivalent)
func buildPowerShellScript(targetPath string) string {
	escapedPath := strings.ReplaceAll(targetPath, "'", "''")
	return fmt.Sprintf(`Add-Type -AssemblyName System.Windows.Forms, System.Drawing; $bounds = [System.Windows.Forms.SystemInformation]::VirtualScreen; $bmp = New-Object System.Drawing.Bitmap $bounds.Width, $bounds.Height; $graphics = [System.Drawing.Graphics]::FromImage($bmp); $graphics.CopyFromScreen($bounds.Left, $bounds.Top, 0, 0, $bounds.Size); $bmp.Save('%s', [System.Drawing.Imaging.ImageFormat]::Png); $graphics.Dispose(); $bmp.Dispose();`, escapedPath)
}

// CaptureToFile executes PowerShell script with GDI CopyFromScreen to capture virtual screen
func (w *WindowsCapturer) CaptureToFile(ctx context.Context, targetPath string) error {
	script := buildPowerShellScript(targetPath)
	cmd := exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("windows capture failed: %v, output: %s", err, string(output))
	}
	return nil
}

// CaptureBytes captures the screen and returns PNG binary bytes directly
func (w *WindowsCapturer) CaptureBytes(ctx context.Context) ([]byte, error) {
	tmpFile, err := os.CreateTemp("", "win-cap-*.png")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file for capture: %w", err)
	}
	tmpPath := tmpFile.Name()
	_ = tmpFile.Close()
	defer os.Remove(tmpPath)

	if err := w.CaptureToFile(ctx, tmpPath); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(tmpPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read captured file: %w", err)
	}

	return data, nil
}
