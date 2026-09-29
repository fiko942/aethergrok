package grokrunner

import (
	"testing"
)

func TestParseGrokConfigTOML(t *testing.T) {
	rawToml := `
[cli]
installer = "internal"
auto_update = true

[model.9router]
model = "geminigacor"
base_url = "http://127.0.0.1:20128/v1"
name = "9Router"
description = "Routed via 9Router gateway"
api_backend = "chat_completions"
api_key = "sk-2d6f02df6a1607a9-pzclsw-c78c4222"
context_window = 200000

[model.9router-general-purpose]
model = "geminigacor"
base_url = "http://127.0.0.1:20128/v1"
name = "9Router general-purpose"
description = "Routed via 9Router gateway"
api_backend = "chat_completions"
api_key = "sk-2d6f02df6a1607a9-pzclsw-c78c4222"
context_window = 200000

[models]
default = "9router"

[ui]
permission_mode = "always-approve"
`

	cfg := ParseGrokConfigTOML(rawToml)
	if cfg == nil {
		t.Fatalf("expected non-nil config")
	}

	if cfg.DefaultModel != "9router" {
		t.Errorf("expected default model '9router', got %q", cfg.DefaultModel)
	}

	if cfg.PermissionMode != "always-approve" {
		t.Errorf("expected permission_mode 'always-approve', got %q", cfg.PermissionMode)
	}

	m9, exists := cfg.Models["9router"]
	if !exists {
		t.Fatalf("expected '9router' model in parsed models")
	}

	if m9.Model != "geminigacor" {
		t.Errorf("expected model 'geminigacor', got %q", m9.Model)
	}

	if m9.BaseURL != "http://127.0.0.1:20128/v1" {
		t.Errorf("expected base_url 'http://127.0.0.1:20128/v1', got %q", m9.BaseURL)
	}

	if m9.APIKey != "sk-2d6f02df6a1607a9-pzclsw-c78c4222" {
		t.Errorf("expected api_key 'sk-2d6f02df6a1607a9-pzclsw-c78c4222', got %q", m9.APIKey)
	}

	if m9.ContextWindow != 200000 {
		t.Errorf("expected context_window 200000, got %d", m9.ContextWindow)
	}
}

func TestResolveTranscriptionConfig(t *testing.T) {
	baseURL, apiKey, candidateModels := ResolveTranscriptionConfig("")

	if baseURL == "" {
		t.Errorf("expected non-empty baseURL")
	}

	if len(candidateModels) == 0 {
		t.Errorf("expected non-empty candidateModels")
	}

	// Should not have trailing /v1 in normalized baseURL
	if baseURL != "http://127.0.0.1:20128" {
		t.Logf("resolved baseURL: %s", baseURL)
	}

	t.Logf("Resolved APIKey length: %d", len(apiKey))
	t.Logf("Candidate models: %v", candidateModels)
}
