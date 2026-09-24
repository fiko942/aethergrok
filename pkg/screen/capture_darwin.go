//go:build darwin

package screen

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

// DarwinCapturer implements macOS native screen snapshot capture
type DarwinCapturer struct{}

// DefaultNativeCapturer returns the DarwinCapturer implementation for macOS
func DefaultNativeCapturer() NativeCapturer {
	return &DarwinCapturer{}
}

// CaptureToFile uses macOS screencapture utility (-x silent mode)
func (d *DarwinCapturer) CaptureToFile(ctx context.Context, targetPath string) error {
	cmd := exec.CommandContext(ctx, "screencapture", "-x", targetPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("screencapture failed: %v, output: %s", err, string(output))
	}
	return nil
}

// CaptureBytes captures the screen and returns PNG binary bytes directly
func (d *DarwinCapturer) CaptureBytes(ctx context.Context) ([]byte, error) {
	tmpFile, err := os.CreateTemp("", "darwin-cap-*.png")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file for capture: %w", err)
	}
	tmpPath := tmpFile.Name()
	_ = tmpFile.Close()
	defer os.Remove(tmpPath)

	if err := d.CaptureToFile(ctx, tmpPath); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(tmpPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read captured file: %w", err)
	}

	return data, nil
}
