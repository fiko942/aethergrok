package grokrunner

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

var contextWindowRegex = regexp.MustCompile(`(?m)^\s*context_window\s*=\s*\d+`)

// SyncGrokConfigContextWindow updates context_window values in ~/.grok/config.toml
// to match the user's configured maxContextTokens so that the Grok CLI process
// enforces the token budget at the engine level. Supported across Windows and macOS.
func SyncGrokConfigContextWindow(tokenLimit int) error {
	if tokenLimit < 16000 {
		tokenLimit = 200000
	}

	homeDir, err := os.UserHomeDir()
	if envGrokHome := os.Getenv("GROK_HOME"); envGrokHome != "" {
		homeDir = envGrokHome
	}
	if err != nil || homeDir == "" {
		return err
	}

	configPath := filepath.Join(homeDir, ".grok", "config.toml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	content := string(data)
	newLine := fmt.Sprintf("context_window = %d", tokenLimit)

	var updated string
	if contextWindowRegex.MatchString(content) {
		updated = contextWindowRegex.ReplaceAllString(content, newLine)
	} else {
		// If no context_window exists in models, append to models sections
		updated = content
	}

	if updated != content {
		return os.WriteFile(configPath, []byte(updated), 0644)
	}
	return nil
}
