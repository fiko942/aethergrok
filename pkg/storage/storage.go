package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// AppSettings defines all user-configurable settings and studio preferences
type AppSettings struct {
	Theme                      string  `json:"theme"`
	DefaultModel               string  `json:"defaultModel"`
	DefaultReasoningEffort     string  `json:"defaultReasoningEffort"`
	PermissionMode             string  `json:"permissionMode"`
	PlanGateMode               string  `json:"planGateMode"`
	AnimationsEnabled          bool    `json:"animationsEnabled"`
	GrokBinaryPath             string  `json:"grokBinaryPath"`
	SnapshotShortcut           string  `json:"snapshotShortcut"`
	SnapshotDelayMs            int     `json:"snapshotDelayMs"`
	SnapshotAutoHideWindow     bool    `json:"snapshotAutoHideWindow"`
	SnapshotSoundEnabled       bool    `json:"snapshotSoundEnabled"`
	SnapshotFlashEnabled       bool    `json:"snapshotFlashEnabled"`
	SnapshotAutoAttach         bool    `json:"snapshotAutoAttach"`
	ActiveWindowTurnCount      int     `json:"activeWindowTurnCount"`
	MaxContextTokens           int     `json:"maxContextTokens"`
	SidebarWidth               int     `json:"sidebarWidth"`
	SidebarCollapsed           bool    `json:"sidebarCollapsed"`
	SelectedMicrophoneDeviceID string  `json:"selectedMicrophoneDeviceId"`
	DictationShortcut          string  `json:"dictationShortcut,omitempty"`
	DictationMuteSystemAudio   *bool   `json:"dictationMuteSystemAudio,omitempty"`
	DictationHoldThresholdMs   int     `json:"dictationHoldThresholdMs,omitempty"`
	UpdatedAt                  int64   `json:"updatedAt,omitempty"`
}

// Workspace represents a registered project directory
type Workspace struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	CreatedAt int64  `json:"createdAt"`
}

// UIState records tabs and active state for instant session restoration
type UIState struct {
	ActiveWorkspaceID string   `json:"activeWorkspaceId"`
	ActiveSessionID   string   `json:"activeSessionId"`
	OpenTabSessionIDs []string `json:"openTabSessionIds"`
	UpdatedAt         int64    `json:"updatedAt,omitempty"`
}

// StorageManager manages persistent storage files in ~/.grok/
type StorageManager struct {
	mu           sync.RWMutex
	baseDir      string
	settingsFile string
	workspacesFile string
	uiStateFile  string
}

// DefaultSettings returns safe default settings matching the UI design system
func DefaultSettings() AppSettings {
	delayMs := 50
	if runtime.GOOS == "windows" {
		delayMs = 80
	}

	return AppSettings{
		Theme:                      "dark-studio",
		DefaultModel:               "9router",
		DefaultReasoningEffort:     "medium",
		PermissionMode:             "default",
		PlanGateMode:               "active",
		AnimationsEnabled:          true,
		GrokBinaryPath:             "",
		SnapshotShortcut:           "CmdOrCtrl+Shift+S",
		SnapshotDelayMs:            delayMs,
		SnapshotAutoHideWindow:     true,
		SnapshotSoundEnabled:       true,
		SnapshotFlashEnabled:       true,
		SnapshotAutoAttach:         true,
		ActiveWindowTurnCount:      10,
		MaxContextTokens:           200000,
		SidebarWidth:               288,
		SidebarCollapsed:           false,
		SelectedMicrophoneDeviceID: "",
		DictationShortcut:          "\\",
		DictationMuteSystemAudio:   boolPtr(true),
		DictationHoldThresholdMs:   300,
		UpdatedAt:                  time.Now().UnixMilli(),
	}
}

func boolPtr(b bool) *bool {
	return &b
}

// NewStorageManager initializes or creates storage files under baseDir (or ~/.grok/)
func NewStorageManager(customDir string) (*StorageManager, error) {
	dir := customDir
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		dir = filepath.Join(home, ".grok")
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory %s: %w", dir, err)
	}

	sm := &StorageManager{
		baseDir:        dir,
		settingsFile:   filepath.Join(dir, "settings.jsonl"),
		workspacesFile: filepath.Join(dir, "workspaces.jsonl"),
		uiStateFile:    filepath.Join(dir, "ui_state.jsonl"),
	}

	return sm, nil
}

// GetSettings reads the latest settings from disk, falling back to defaults if not found
func (sm *StorageManager) GetSettings() (AppSettings, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	data, err := os.ReadFile(sm.settingsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultSettings(), nil
		}
		return DefaultSettings(), err
	}

	var latest AppSettings
	lines := splitLines(data)
	if len(lines) == 0 {
		return DefaultSettings(), nil
	}

	// Read last valid entry (WAL/Append-only style)
	found := false
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		if len(line) == 0 {
			continue
		}
		if err := json.Unmarshal(line, &latest); err == nil {
			found = true
			break
		}
	}

	if !found {
		return DefaultSettings(), nil
	}

	// Apply default fallbacks and sanitize cross-platform invalid values
	if runtime.GOOS == "windows" && strings.HasPrefix(latest.GrokBinaryPath, "/Users/") {
		latest.GrokBinaryPath = ""
	} else if runtime.GOOS != "windows" && len(latest.GrokBinaryPath) >= 2 && latest.GrokBinaryPath[1] == ':' {
		latest.GrokBinaryPath = ""
	} else if latest.GrokBinaryPath != "" {
		if _, err := os.Stat(latest.GrokBinaryPath); err != nil {
			latest.GrokBinaryPath = ""
		}
	}

	if latest.SnapshotShortcut == "" {
		latest.SnapshotShortcut = "CmdOrCtrl+Shift+S"
	}

	if latest.DictationShortcut == "" {
		latest.DictationShortcut = "\\"
	}
	if latest.DictationMuteSystemAudio == nil {
		latest.DictationMuteSystemAudio = boolPtr(true)
	}
	if latest.DictationHoldThresholdMs <= 0 {
		latest.DictationHoldThresholdMs = 300
	}
	if latest.SnapshotDelayMs <= 0 {
		if runtime.GOOS == "windows" {
			latest.SnapshotDelayMs = 80
		} else {
			latest.SnapshotDelayMs = 50
		}
	}

	return latest, nil
}

// SaveSettings appends a new settings snapshot to disk atomically
func (sm *StorageManager) SaveSettings(settings AppSettings) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	settings.UpdatedAt = time.Now().UnixMilli()
	bytes, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	// Direct atomic write to keep file compact and clean
	tmpFile := sm.settingsFile + ".tmp"
	if err := os.WriteFile(tmpFile, append(bytes, '\n'), 0644); err != nil {
		return fmt.Errorf("failed to write temp settings file: %w", err)
	}

	return os.Rename(tmpFile, sm.settingsFile)
}

// GetWorkspaces retrieves all stored workspaces
func (sm *StorageManager) GetWorkspaces() ([]Workspace, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	data, err := os.ReadFile(sm.workspacesFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []Workspace{}, nil
		}
		return []Workspace{}, err
	}

	var list []Workspace
	lines := splitLines(data)
	if len(lines) == 0 {
		return []Workspace{}, nil
	}

	// Last line contains current workspace list snapshot
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		if len(line) == 0 {
			continue
		}
		if err := json.Unmarshal(line, &list); err == nil {
			return list, nil
		}
	}

	return []Workspace{}, nil
}

// SaveWorkspaces writes the current workspace collection to disk
func (sm *StorageManager) SaveWorkspaces(workspaces []Workspace) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	bytes, err := json.Marshal(workspaces)
	if err != nil {
		return fmt.Errorf("failed to marshal workspaces: %w", err)
	}

	tmpFile := sm.workspacesFile + ".tmp"
	if err := os.WriteFile(tmpFile, append(bytes, '\n'), 0644); err != nil {
		return fmt.Errorf("failed to write temp workspaces file: %w", err)
	}

	return os.Rename(tmpFile, sm.workspacesFile)
}

// GetUIState returns active workspace, active session, and open tab IDs
func (sm *StorageManager) GetUIState() (UIState, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	defaultState := UIState{
		ActiveWorkspaceID: "",
		ActiveSessionID:   "",
		OpenTabSessionIDs: []string{},
		UpdatedAt:         time.Now().UnixMilli(),
	}

	data, err := os.ReadFile(sm.uiStateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultState, nil
		}
		return defaultState, err
	}

	lines := splitLines(data)
	if len(lines) == 0 {
		return defaultState, nil
	}

	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		if len(line) == 0 {
			continue
		}
		var state UIState
		if err := json.Unmarshal(line, &state); err == nil {
			return state, nil
		}
	}

	return defaultState, nil
}

// SaveUIState saves the active UI state (open tabs and active session)
func (sm *StorageManager) SaveUIState(state UIState) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	state.UpdatedAt = time.Now().UnixMilli()
	bytes, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal ui state: %w", err)
	}

	tmpFile := sm.uiStateFile + ".tmp"
	if err := os.WriteFile(tmpFile, append(bytes, '\n'), 0644); err != nil {
		return fmt.Errorf("failed to write temp ui state file: %w", err)
	}

	return os.Rename(tmpFile, sm.uiStateFile)
}

func splitLines(data []byte) [][]byte {
	var lines [][]byte
	start := 0
	for i, b := range data {
		if b == '\n' {
			line := data[start:i]
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			lines = append(lines, line)
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, data[start:])
	}
	return lines
}
