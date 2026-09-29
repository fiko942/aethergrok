package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestAppVersionConsistency(t *testing.T) {
	app := NewApp()
	ver := app.GetAppVersion()
	if ver != AppVersion {
		t.Fatalf("expected GetAppVersion() to return %q, got %q", AppVersion, ver)
	}
	if ver != "1.0.5" {
		t.Fatalf("expected AppVersion to be 1.0.5, got %q", ver)
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
