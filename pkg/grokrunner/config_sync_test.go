package grokrunner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSyncGrokConfigContextWindow(t *testing.T) {
	tempHome := t.TempDir()
	grokDir := filepath.Join(tempHome, ".grok")
	_ = os.MkdirAll(grokDir, 0755)
	configPath := filepath.Join(grokDir, "config.toml")

	initialContent := `[model.9router]
model = "geminigacor"
context_window = 1048576

[model.9router-general-purpose]
model = "geminigacor"
context_window = 1048576
`
	if err := os.WriteFile(configPath, []byte(initialContent), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("USERPROFILE", tempHome)
	t.Setenv("HOME", tempHome)

	if err := SyncGrokConfigContextWindow(128000); err != nil {
		t.Fatalf("SyncGrokConfigContextWindow failed: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if strings.Contains(content, "1048576") {
		t.Errorf("Expected 1048576 to be replaced, got:\n%s", content)
	}
	if !strings.Contains(content, "context_window = 128000") {
		t.Errorf("Expected context_window = 128000 in config.toml, got:\n%s", content)
	}
}
