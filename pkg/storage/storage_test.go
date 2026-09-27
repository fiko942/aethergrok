package storage

import (
	"os"
	"testing"
)

func TestStorageManager_Settings(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aethergrok_storage_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	sm, err := NewStorageManager(tempDir)
	if err != nil {
		t.Fatalf("failed to init storage: %v", err)
	}

	// 1. Initial should return defaults
	settings, err := sm.GetSettings()
	if err != nil {
		t.Fatalf("GetSettings failed: %v", err)
	}
	if settings.Theme != "dark-studio" || settings.DefaultModel != "9router" {
		t.Errorf("expected default settings, got %+v", settings)
	}

	// 2. Save modified settings
	settings.Theme = "light-antd"
	settings.DefaultModel = "custom-model"
	settings.PermissionMode = "bypassPermissions"
	if err := sm.SaveSettings(settings); err != nil {
		t.Fatalf("SaveSettings failed: %v", err)
	}

	// 3. Reload from disk
	reloaded, err := sm.GetSettings()
	if err != nil {
		t.Fatalf("GetSettings after save failed: %v", err)
	}
	if reloaded.Theme != "light-antd" || reloaded.DefaultModel != "custom-model" || reloaded.PermissionMode != "bypassPermissions" {
		t.Errorf("expected updated settings, got %+v", reloaded)
	}
}

func TestStorageManager_Workspaces(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aethergrok_storage_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	sm, err := NewStorageManager(tempDir)
	if err != nil {
		t.Fatalf("failed to init storage: %v", err)
	}

	workspaces := []Workspace{
		{ID: "ws-1", Name: "Project A", Path: "/path/to/a", CreatedAt: 1000},
		{ID: "ws-2", Name: "Project B", Path: "/path/to/b", CreatedAt: 2000},
	}

	if err := sm.SaveWorkspaces(workspaces); err != nil {
		t.Fatalf("SaveWorkspaces failed: %v", err)
	}

	loaded, err := sm.GetWorkspaces()
	if err != nil {
		t.Fatalf("GetWorkspaces failed: %v", err)
	}
	if len(loaded) != 2 || loaded[0].Name != "Project A" || loaded[1].ID != "ws-2" {
		t.Errorf("expected 2 workspaces, got %+v", loaded)
	}
}

func TestStorageManager_UIState(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aethergrok_storage_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	sm, err := NewStorageManager(tempDir)
	if err != nil {
		t.Fatalf("failed to init storage: %v", err)
	}

	uiState := UIState{
		ActiveWorkspaceID: "ws-1",
		ActiveSessionID:   "sess-xyz",
		OpenTabSessionIDs: []string{"sess-xyz", "sess-abc"},
	}

	if err := sm.SaveUIState(uiState); err != nil {
		t.Fatalf("SaveUIState failed: %v", err)
	}

	loaded, err := sm.GetUIState()
	if err != nil {
		t.Fatalf("GetUIState failed: %v", err)
	}
	if loaded.ActiveWorkspaceID != "ws-1" || loaded.ActiveSessionID != "sess-xyz" || len(loaded.OpenTabSessionIDs) != 2 {
		t.Errorf("expected valid ui state, got %+v", loaded)
	}
}
