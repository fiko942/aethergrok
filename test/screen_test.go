package test

import (
	"context"
	"encoding/base64"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"aethergrok/pkg/screen"
)

// mockWindowController tracks hide/show calls and timestamps
type mockWindowController struct {
	mu        sync.Mutex
	hideCalls int
	showCalls int
	events    []string
}

func (m *mockWindowController) Hide() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.hideCalls++
	m.events = append(m.events, "hide")
}

func (m *mockWindowController) Show() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.showCalls++
	m.events = append(m.events, "show")
}

// mockCapturer implements NativeCapturer for deterministic tests
type mockCapturer struct {
	mu           sync.Mutex
	captureCalls int
	bytesToRet   []byte
	errToRet     error
	onCapture    func()
}

func (m *mockCapturer) CaptureToFile(ctx context.Context, targetPath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.captureCalls++
	if m.errToRet != nil {
		return m.errToRet
	}
	return os.WriteFile(targetPath, m.bytesToRet, 0644)
}

func (m *mockCapturer) CaptureBytes(ctx context.Context) ([]byte, error) {
	m.mu.Lock()
	if m.onCapture != nil {
		m.onCapture()
	}
	m.captureCalls++
	err := m.errToRet
	bytes := m.bytesToRet
	m.mu.Unlock()

	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func TestScreenCapture_CoordinationSequence(t *testing.T) {
	// 1x1 transparent PNG sample binary
	pngBytes := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4,
		0x89, 0x00, 0x00, 0x00, 0x0A, 0x49, 0x44, 0x41,
		0x54, 0x78, 0x9C, 0x63, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00,
		0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE,
		0x42, 0x60, 0x82,
	}

	mockWin := &mockWindowController{}
	var eventsDuringCapture []string

	capturer := &mockCapturer{
		bytesToRet: pngBytes,
		onCapture: func() {
			mockWin.mu.Lock()
			eventsDuringCapture = append([]string(nil), mockWin.events...)
			mockWin.mu.Unlock()
		},
	}

	orchestrator := screen.NewOrchestratorWithCapturer(capturer)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	start := time.Now()
	res, err := orchestrator.CaptureScreenExcludingWindow(ctx, mockWin, 20)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected error during capture: %v", err)
	}

	if elapsed < 20*time.Millisecond {
		t.Errorf("expected delay to be at least 20ms, was %v", elapsed)
	}

	if res == nil {
		t.Fatalf("expected SnapshotResult, got nil")
	}

	// Verify events during capture had only "hide"
	if len(eventsDuringCapture) != 1 || eventsDuringCapture[0] != "hide" {
		t.Errorf("expected window to be hidden during capture, events were: %v", eventsDuringCapture)
	}

	// Verify final window state had hide -> show
	mockWin.mu.Lock()
	finalEvents := mockWin.events
	hideCalls := mockWin.hideCalls
	showCalls := mockWin.showCalls
	mockWin.mu.Unlock()

	if hideCalls != 1 || showCalls != 1 {
		t.Errorf("expected 1 hide and 1 show call, got hide=%d, show=%d", hideCalls, showCalls)
	}
	if len(finalEvents) != 2 || finalEvents[0] != "hide" || finalEvents[1] != "show" {
		t.Errorf("expected sequence [hide, show], got: %v", finalEvents)
	}

	// Verify Base64 & DataURL
	expectedBase64 := base64.StdEncoding.EncodeToString(pngBytes)
	if res.Base64 != expectedBase64 {
		t.Errorf("expected base64 %q, got %q", expectedBase64, res.Base64)
	}
	if !strings.HasPrefix(res.DataURL, "data:image/png;base64,") {
		t.Errorf("expected data url prefix data:image/png;base64,, got %q", res.DataURL)
	}

	// Verify file path exists on disk
	if res.FilePath == "" {
		t.Errorf("expected non-empty FilePath")
	} else {
		fileData, err := os.ReadFile(res.FilePath)
		if err != nil {
			t.Errorf("failed to read created snapshot file: %v", err)
		} else if len(fileData) != len(pngBytes) {
			t.Errorf("file data length mismatch: expected %d, got %d", len(pngBytes), len(fileData))
		}
		// Cleanup temp file
		_ = os.Remove(res.FilePath)
	}
}

func TestScreenCapture_ErrorAndWindowRestoration(t *testing.T) {
	mockWin := &mockWindowController{}
	capturer := &mockCapturer{
		errToRet: errors.New("simulated capture failure"),
	}

	orchestrator := screen.NewOrchestratorWithCapturer(capturer)

	ctx := context.Background()
	res, err := orchestrator.CaptureScreenExcludingWindow(ctx, mockWin, 5)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if res != nil {
		t.Fatalf("expected nil result on error, got %v", res)
	}

	mockWin.mu.Lock()
	hideCalls := mockWin.hideCalls
	showCalls := mockWin.showCalls
	mockWin.mu.Unlock()

	if hideCalls != 1 || showCalls != 1 {
		t.Errorf("expected window to be restored even on error: hide=%d, show=%d", hideCalls, showCalls)
	}
}

func TestScreenCapture_ContextCancellation(t *testing.T) {
	mockWin := &mockWindowController{}
	capturer := &mockCapturer{
		bytesToRet: []byte{0x89, 0x50, 0x4E, 0x47},
	}

	orchestrator := screen.NewOrchestratorWithCapturer(capturer)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	res, err := orchestrator.CaptureScreenExcludingWindow(ctx, mockWin, 100)

	if err == nil {
		t.Fatalf("expected context cancellation error, got nil")
	}
	if res != nil {
		t.Fatalf("expected nil result on cancellation, got %v", res)
	}

	mockWin.mu.Lock()
	showCalls := mockWin.showCalls
	mockWin.mu.Unlock()

	if showCalls != 1 {
		t.Errorf("expected window to be shown even after cancellation, showCalls=%d", showCalls)
	}
}

func TestScreenCapture_RealNativeCapturer(t *testing.T) {
	// Test DefaultNativeCapturer initialization and execution on current platform
	capturer := screen.DefaultNativeCapturer()
	if capturer == nil {
		t.Fatalf("DefaultNativeCapturer returned nil")
	}

	orchestrator := screen.NewOrchestrator()
	if orchestrator == nil {
		t.Fatalf("NewOrchestrator returned nil")
	}

	// Capture without window controller (headless mode test)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := orchestrator.CaptureScreenExcludingWindow(ctx, nil, 10)
	if err != nil {
		// In CI / non-display environments, capture may fail if no screen is attached, which is acceptable
		t.Logf("native screen capture skipped or failed in current environment: %v", err)
		return
	}

	if res == nil {
		t.Fatalf("expected SnapshotResult, got nil")
	}
	if len(res.Base64) == 0 {
		t.Errorf("expected non-empty Base64 data")
	}
	if !strings.HasPrefix(res.DataURL, "data:image/png;base64,") && !strings.HasPrefix(res.DataURL, "data:image/jpeg;base64,") {
		t.Errorf("expected valid data URL prefix, got %s", res.DataURL[:min(30, len(res.DataURL))])
	}
	if res.Size > 650*1024 {
		t.Errorf("expected snapshot size <= 650KB, got %d bytes", res.Size)
	}
	if res.FilePath != "" {
		_ = os.Remove(res.FilePath)
	}
}
