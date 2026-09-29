package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestAppVersionConsistency(t *testing.T) {
	app := NewApp()
	ver := app.GetAppVersion()
	if ver != AppVersion {
		t.Fatalf("expected GetAppVersion() to return %q, got %q", AppVersion, ver)
	}

	// Verify wails.json matches AppVersion
	wailsData, err := os.ReadFile("wails.json")
	if err != nil {
		t.Fatalf("failed to read wails.json: %v", err)
	}
	var wailsCfg struct {
		Version string `json:"version"`
		Info    struct {
			ProductVersion string `json:"productVersion"`
		} `json:"info"`
	}
	if err := json.Unmarshal(wailsData, &wailsCfg); err != nil {
		t.Fatalf("failed to parse wails.json: %v", err)
	}

	if wailsCfg.Version != AppVersion {
		t.Errorf("wails.json version (%q) does not match AppVersion (%q)", wailsCfg.Version, AppVersion)
	}
	if wailsCfg.Info.ProductVersion != AppVersion {
		t.Errorf("wails.json info.productVersion (%q) does not match AppVersion (%q)", wailsCfg.Info.ProductVersion, AppVersion)
	}

	// Verify frontend/package.json matches AppVersion
	pkgData, err := os.ReadFile("frontend/package.json")
	if err != nil {
		t.Fatalf("failed to read frontend/package.json: %v", err)
	}
	var pkgCfg struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(pkgData, &pkgCfg); err != nil {
		t.Fatalf("failed to parse frontend/package.json: %v", err)
	}
	if pkgCfg.Version != AppVersion {
		t.Errorf("frontend/package.json version (%q) does not match AppVersion (%q)", pkgCfg.Version, AppVersion)
	}
}

func TestGetPlanContent(t *testing.T) {
	app := NewApp()

	// 1. Empty path test
	if _, err := app.GetPlanContent(""); err == nil {
		t.Errorf("expected error for empty planPath, got nil")
	}

	// 2. Create a temporary test plan.md
	tempDir, err := os.MkdirTemp("", "aethergrok-plan-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	planFile := filepath.Join(tempDir, "plan.md")
	expectedContent := "# Test Engineering Plan\n\n- Task 1\n- Task 2"
	if err := os.WriteFile(planFile, []byte(expectedContent), 0644); err != nil {
		t.Fatalf("failed to write plan.md: %v", err)
	}

	// Direct file path
	content, err := app.GetPlanContent(planFile)
	if err != nil || content != expectedContent {
		t.Errorf("failed direct file path: got %q, err: %v", content, err)
	}

	// Quoted file path
	quotedPath := `"` + planFile + `"`
	content, err = app.GetPlanContent(quotedPath)
	if err != nil || content != expectedContent {
		t.Errorf("failed quoted file path: got %q, err: %v", content, err)
	}

	// Path with trailing sentence marker
	sentencePath := planFile + ". The file exists and is empty."
	content, err = app.GetPlanContent(sentencePath)
	if err != nil || content != expectedContent {
		t.Errorf("failed sentence path: got %q, err: %v", content, err)
	}

	// Directory path (should resolve plan.md inside directory)
	content, err = app.GetPlanContent(tempDir)
	if err != nil || content != expectedContent {
		t.Errorf("failed directory path: got %q, err: %v", content, err)
	}
}

type mockWindowRuntimeOps struct {
	hideCalls       int
	showCalls       int
	maximiseCalls   int
	minimiseCalls   int
	fullscreenCalls int

	isFullscreen bool
	isMaximised  bool
	isMinimised  bool
	posX         int
	posY         int
	width        int
	height       int
}

func (m *mockWindowRuntimeOps) WindowHide() {
	m.hideCalls++
}

func (m *mockWindowRuntimeOps) WindowShow() {
	m.showCalls++
}

func (m *mockWindowRuntimeOps) WindowMaximise() {
	m.maximiseCalls++
}

func (m *mockWindowRuntimeOps) WindowMinimise() {
	m.minimiseCalls++
}

func (m *mockWindowRuntimeOps) WindowFullscreen() {
	m.fullscreenCalls++
}

func (m *mockWindowRuntimeOps) WindowIsFullscreen() bool {
	return m.isFullscreen
}

func (m *mockWindowRuntimeOps) WindowIsMaximised() bool {
	return m.isMaximised
}

func (m *mockWindowRuntimeOps) WindowIsMinimised() bool {
	return m.isMinimised
}

func (m *mockWindowRuntimeOps) WindowGetPosition() (int, int) {
	return m.posX, m.posY
}

func (m *mockWindowRuntimeOps) WindowGetSize() (int, int) {
	return m.width, m.height
}

func TestWailsWindowController_MaximizedPreservation(t *testing.T) {
	mockOps := &mockWindowRuntimeOps{
		isMaximised: true,
		posX:        -8,
		posY:        -8,
		width:       1936,
		height:      1056,
	}

	ctrl := &wailsWindowController{
		ops: mockOps,
	}

	// 1. Hide window during snapshot
	ctrl.Hide()

	if mockOps.hideCalls != 1 {
		t.Errorf("expected 1 hide call, got %d", mockOps.hideCalls)
	}
	if !ctrl.state.Captured {
		t.Errorf("expected state to be captured")
	}
	if !ctrl.state.IsMaximised {
		t.Errorf("expected IsMaximised to be true")
	}
	if ctrl.state.X != -8 || ctrl.state.Y != -8 {
		t.Errorf("expected position (-8, -8), got (%d, %d)", ctrl.state.X, ctrl.state.Y)
	}
	if ctrl.state.Width != 1936 || ctrl.state.Height != 1056 {
		t.Errorf("expected size (1936, 1056), got (%d, %d)", ctrl.state.Width, ctrl.state.Height)
	}

	// 2. Show window after snapshot
	ctrl.Show()

	if mockOps.showCalls != 1 {
		t.Errorf("expected 1 show call, got %d", mockOps.showCalls)
	}
	if mockOps.maximiseCalls != 1 {
		t.Errorf("expected WindowMaximise to be called to restore maximized state, got %d calls", mockOps.maximiseCalls)
	}
	if mockOps.fullscreenCalls != 0 || mockOps.minimiseCalls != 0 {
		t.Errorf("unexpected calls: fullscreen=%d, minimise=%d", mockOps.fullscreenCalls, mockOps.minimiseCalls)
	}
}

func TestWailsWindowController_NormalWindowPreservation(t *testing.T) {
	mockOps := &mockWindowRuntimeOps{
		isMaximised:  false,
		isFullscreen: false,
		isMinimised:  false,
		posX:         200,
		posY:         150,
		width:        1280,
		height:       850,
	}

	ctrl := &wailsWindowController{
		ops: mockOps,
	}

	ctrl.Hide()
	if !ctrl.state.Captured {
		t.Errorf("expected state to be captured")
	}
	if ctrl.state.IsMaximised || ctrl.state.IsFullscreen || ctrl.state.IsMinimised {
		t.Errorf("expected normal window state flags to be false")
	}
	if ctrl.state.X != 200 || ctrl.state.Y != 150 {
		t.Errorf("expected position (200, 150), got (%d, %d)", ctrl.state.X, ctrl.state.Y)
	}
	if ctrl.state.Width != 1280 || ctrl.state.Height != 850 {
		t.Errorf("expected size (1280, 850), got (%d, %d)", ctrl.state.Width, ctrl.state.Height)
	}

	ctrl.Show()
	if mockOps.showCalls != 1 {
		t.Errorf("expected 1 show call, got %d", mockOps.showCalls)
	}
	if mockOps.maximiseCalls != 0 || mockOps.fullscreenCalls != 0 || mockOps.minimiseCalls != 0 {
		t.Errorf("expected no special state restores for normal window, got: max=%d, full=%d, min=%d",
			mockOps.maximiseCalls, mockOps.fullscreenCalls, mockOps.minimiseCalls)
	}
}

func TestWailsWindowController_FullscreenPreservation(t *testing.T) {
	mockOps := &mockWindowRuntimeOps{
		isFullscreen: true,
		posX:         0,
		posY:         0,
		width:        1920,
		height:       1080,
	}

	ctrl := &wailsWindowController{
		ops: mockOps,
	}

	ctrl.Hide()
	if !ctrl.state.IsFullscreen {
		t.Errorf("expected IsFullscreen to be true")
	}

	ctrl.Show()
	if mockOps.showCalls != 1 {
		t.Errorf("expected 1 show call, got %d", mockOps.showCalls)
	}
	if mockOps.fullscreenCalls != 1 {
		t.Errorf("expected WindowFullscreen to be called, got %d calls", mockOps.fullscreenCalls)
	}
	if mockOps.maximiseCalls != 0 || mockOps.minimiseCalls != 0 {
		t.Errorf("unexpected calls: max=%d, min=%d", mockOps.maximiseCalls, mockOps.minimiseCalls)
	}
}

func TestWailsWindowController_MinimisedPreservation(t *testing.T) {
	mockOps := &mockWindowRuntimeOps{
		isMinimised: true,
	}

	ctrl := &wailsWindowController{
		ops: mockOps,
	}

	ctrl.Hide()
	if !ctrl.state.IsMinimised {
		t.Errorf("expected IsMinimised to be true")
	}

	ctrl.Show()
	if mockOps.showCalls != 1 {
		t.Errorf("expected 1 show call, got %d", mockOps.showCalls)
	}
	if mockOps.minimiseCalls != 1 {
		t.Errorf("expected WindowMinimise to be called, got %d calls", mockOps.minimiseCalls)
	}
}

