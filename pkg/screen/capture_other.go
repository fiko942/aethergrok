//go:build !darwin && !windows

package screen

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

// FallbackCapturer implements screen snapshot capture for Linux and other platforms
type FallbackCapturer struct{}

// DefaultNativeCapturer returns the FallbackCapturer implementation
func DefaultNativeCapturer() NativeCapturer {
	return &FallbackCapturer{}
}

// CaptureToFile uses import (ImageMagick) or maim/scrot
func (f *FallbackCapturer) CaptureToFile(ctx context.Context, targetPath string) error {
	cmd := exec.CommandContext(ctx, "import", "-window", "root", targetPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("fallback capture failed: %v, output: %s", err, string(output))
	}
	return nil
}

// CaptureBytes captures the screen and returns PNG binary bytes directly
func (f *FallbackCapturer) CaptureBytes(ctx context.Context) ([]byte, error) {
	tmpFile, err := os.CreateTemp("", "fallback-cap-*.png")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file for capture: %w", err)
	}
	tmpPath := tmpFile.Name()
	_ = tmpFile.Close()
	defer os.Remove(tmpPath)

	if err := f.CaptureToFile(ctx, tmpPath); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(tmpPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read captured file: %w", err)
	}

	return data, nil
}
