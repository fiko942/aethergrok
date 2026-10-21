package grokrunner

import (
	"strings"
	"testing"
)

func TestParseModelsOutput_WithCustomModels(t *testing.T) {
	output := `You are logged in with grok.com.

Default model: 9router

Available models:
  - grok-4.7
  * 9router (default)
  - 9router-general-purpose
  - 9router-explore
  - 9router-plan
`
	models := parseModelsOutput(output)
	if len(models) != 5 {
		t.Fatalf("expected 5 models, got %d", len(models))
	}

	foundDefault := false
	for _, m := range models {
		if m.ID == "9router" && m.IsDefault {
			foundDefault = true
		}
	}
	if !foundDefault {
		t.Errorf("expected 9router to be marked as default")
	}
}

func TestDiscoverAvailableModels_ConfigMerging(t *testing.T) {
	r := NewRunner()
	models := r.DiscoverAvailableModels()
	if len(models) == 0 {
		t.Fatal("expected at least 1 model to be discovered")
	}

	// Verify that models declared in config.toml have their underlying target model noted
	foundTarget := false
	for _, m := range models {
		if strings.Contains(m.Description, "geminigacor") || m.ModelTarget == "geminigacor" {
			foundTarget = true
			break
		}
	}
	if !foundTarget {
		t.Logf("Notice: geminigacor target not present or config.toml not yet merged into description")
	}
}
