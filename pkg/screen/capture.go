package screen

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"os"
	"runtime"
	"time"
)

// Default compositor delay timings per platform
const (
	DefaultDarwinDelayMs  = 50
	DefaultWindowsDelayMs = 80
	DefaultLinuxDelayMs   = 50
)

// SnapshotResult contains data and metadata for captured screen
type SnapshotResult struct {
	FilePath  string `json:"filePath"`
	DataURL   string `json:"dataUrl"`
	Base64    string `json:"base64"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Size        int64  `json:"sizeBytes"`
	Timestamp   int64  `json:"timestamp"`
}

// WindowController interface abstracts window hide/show/focus operations
type WindowController interface {
	Hide()
	Show()
}

// NativeCapturer abstracts the OS-level capture mechanism
type NativeCapturer interface {
	CaptureToFile(ctx context.Context, targetPath string) error
	CaptureBytes(ctx context.Context) ([]byte, error)
}

// Orchestrator coordinates non-intrusive snapshot operations
type Orchestrator struct {
	capturer NativeCapturer
}

// NewOrchestrator creates a screen capture orchestrator with default OS capturer
func NewOrchestrator() *Orchestrator {
	return &Orchestrator{
		capturer: DefaultNativeCapturer(),
	}
}

// NewOrchestratorWithCapturer creates an orchestrator with custom capturer (for testing/injection)
func NewOrchestratorWithCapturer(capturer NativeCapturer) *Orchestrator {
	return &Orchestrator{
		capturer: capturer,
	}
}

// DefaultDelay returns default sleep duration in milliseconds needed for OS compositor
func DefaultDelay() time.Duration {
	switch runtime.GOOS {
	case "darwin":
		return time.Duration(DefaultDarwinDelayMs) * time.Millisecond
	case "windows":
		return time.Duration(DefaultWindowsDelayMs) * time.Millisecond
	default:
		return time.Duration(DefaultLinuxDelayMs) * time.Millisecond
	}
}

// CaptureScreenExcludingWindow performs the 4-phase non-intrusive capture:
// 1. Window Hide
// 2. OS Compositor Sleep (delayMs or default per OS)
// 3. Native Screen Capture
// 4. Window Show / Focus
func (o *Orchestrator) CaptureScreenExcludingWindow(ctx context.Context, win WindowController, delayMs int) (*SnapshotResult, error) {
	if win != nil {
		win.Hide()
		// Guarantee window restoration even on panic or error
		defer win.Show()
	}

	delay := DefaultDelay()
	if delayMs > 0 {
		delay = time.Duration(delayMs) * time.Millisecond
	}

	select {
	case <-time.After(delay):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Capture native screen
	rawBytes, err := o.capturer.CaptureBytes(ctx)
	if err != nil {
		return nil, fmt.Errorf("screen capture failed: %w", err)
	}

	if len(rawBytes) == 0 {
		return nil, fmt.Errorf("captured screen buffer is empty")
	}

	// Optimize and compress image if size exceeds 650KB limit (e.g. 2x Retina screens)
	finalBytes := rawBytes
	mimeType := "image/png"
	fileExt := "png"
	width := 0
	height := 0

	const maxSizeBytes = 650 * 1024 // 650 KB limit

	// Decode image config or full image
	img, _, decodeErr := image.Decode(bytes.NewReader(rawBytes))
	if decodeErr == nil && img != nil {
		bounds := img.Bounds()
		width = bounds.Dx()
		height = bounds.Dy()

		if len(rawBytes) > maxSizeBytes {
			// Compress to high-quality JPEG targeting <= 650KB
			quality := 85
			for quality >= 50 {
				var buf bytes.Buffer
				err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
				if err == nil {
					compressed := buf.Bytes()
					if len(compressed) <= maxSizeBytes || quality == 50 {
						finalBytes = compressed
						mimeType = "image/jpeg"
						fileExt = "jpg"
						break
					}
				}
				quality -= 10
			}
		}
	}

	// Write to temp file for immediate vision prompt attachment or file saving
	tmpFile, err := os.CreateTemp("", fmt.Sprintf("grok-snapshot-*.%s", fileExt))
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary snapshot file: %w", err)
	}
	defer tmpFile.Close()

	if _, err := tmpFile.Write(finalBytes); err != nil {
		_ = os.Remove(tmpFile.Name())
		return nil, fmt.Errorf("failed to write snapshot image data: %w", err)
	}

	base64Data := base64.StdEncoding.EncodeToString(finalBytes)
	dataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Data)

	return &SnapshotResult{
		FilePath:  tmpFile.Name(),
		DataURL:   dataURL,
		Base64:    base64Data,
		Width:     width,
		Height:    height,
		Size:      int64(len(finalBytes)),
		Timestamp: time.Now().UnixMilli(),
	}, nil
}
