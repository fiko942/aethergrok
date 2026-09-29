package grokrunner

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// GrokModelConfig holds configuration parameters for a model defined in config.toml
type GrokModelConfig struct {
	ID            string `json:"id"`
	Model         string `json:"model"`
	BaseURL       string `json:"base_url"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	APIBackend    string `json:"api_backend"`
	APIKey        string `json:"api_key"`
	ContextWindow int    `json:"context_window"`
}

// GrokConfigFile represents parsed contents of a Grok config.toml
type GrokConfigFile struct {
	DefaultModel   string                     `json:"default_model"`
	Models         map[string]GrokModelConfig `json:"models"`
	SubagentModels map[string]string          `json:"subagent_models"`
	PermissionMode string                     `json:"permission_mode"`
}

// ParseGrokConfigTOML parses a TOML string into a GrokConfigFile
func ParseGrokConfigTOML(content string) *GrokConfigFile {
	cfg := &GrokConfigFile{
		Models:         make(map[string]GrokModelConfig),
		SubagentModels: make(map[string]string),
	}

	scanner := bufio.NewScanner(strings.NewReader(content))
	currentSection := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Check for section headers [header]
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.ToLower(strings.TrimSpace(line[1 : len(line)-1]))
			continue
		}

		// Parse key = value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.ToLower(strings.TrimSpace(parts[0]))
		valStr := strings.TrimSpace(parts[1])

		// Strip inline comment if any
		if idx := strings.Index(valStr, " #"); idx != -1 {
			valStr = strings.TrimSpace(valStr[:idx])
		}

		// Strip quotes if string
		val := unquoteTomlString(valStr)

		if currentSection == "models" {
			if key == "default" {
				cfg.DefaultModel = val
			}
		} else if strings.HasPrefix(currentSection, "model.") {
			modelID := strings.TrimPrefix(currentSection, "model.")
			mCfg, exists := cfg.Models[modelID]
			if !exists {
				mCfg = GrokModelConfig{ID: modelID}
			}

			switch key {
			case "model":
				mCfg.Model = val
			case "base_url":
				mCfg.BaseURL = val
			case "name":
				mCfg.Name = val
			case "description":
				mCfg.Description = val
			case "api_backend":
				mCfg.APIBackend = val
			case "api_key":
				mCfg.APIKey = val
			case "context_window":
				if num, err := strconv.Atoi(val); err == nil {
					mCfg.ContextWindow = num
				}
			}
			cfg.Models[modelID] = mCfg
		} else if currentSection == "subagents.models" {
			cfg.SubagentModels[key] = val
		} else if currentSection == "ui" {
			if key == "permission_mode" {
				cfg.PermissionMode = val
			}
		}
	}

	return cfg
}

func unquoteTomlString(s string) string {
	s = strings.TrimSpace(s)
	if (strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`)) ||
		(strings.HasPrefix(s, `'`) && strings.HasSuffix(s, `'`)) {
		if len(s) >= 2 {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// ResolveGrokConfig searches for and parses Grok config.toml files.
// Project-level .grok/config.toml takes precedence over global ~/.grok/config.toml.
func ResolveGrokConfig(workspacePath string) *GrokConfigFile {
	result := &GrokConfigFile{
		Models:         make(map[string]GrokModelConfig),
		SubagentModels: make(map[string]string),
	}

	// 1. Read global ~/.grok/config.toml
	homeDir, err := os.UserHomeDir()
	if envGrokHome := os.Getenv("GROK_HOME"); envGrokHome != "" {
		homeDir = envGrokHome
	}
	if err == nil && homeDir != "" {
		globalConfigPath := filepath.Join(homeDir, ".grok", "config.toml")
		if data, err := os.ReadFile(globalConfigPath); err == nil {
			globalCfg := ParseGrokConfigTOML(string(data))
			if globalCfg.DefaultModel != "" {
				result.DefaultModel = globalCfg.DefaultModel
			}
			if globalCfg.PermissionMode != "" {
				result.PermissionMode = globalCfg.PermissionMode
			}
			for k, v := range globalCfg.Models {
				result.Models[k] = v
			}
			for k, v := range globalCfg.SubagentModels {
				result.SubagentModels[k] = v
			}
		}
	}

	// 2. Read workspace project .grok/config.toml if present
	if strings.TrimSpace(workspacePath) != "" {
		projConfigPath := filepath.Join(workspacePath, ".grok", "config.toml")
		if data, err := os.ReadFile(projConfigPath); err == nil {
			projCfg := ParseGrokConfigTOML(string(data))
			if projCfg.DefaultModel != "" {
				result.DefaultModel = projCfg.DefaultModel
			}
			if projCfg.PermissionMode != "" {
				result.PermissionMode = projCfg.PermissionMode
			}
			for k, v := range projCfg.Models {
				result.Models[k] = v
			}
			for k, v := range projCfg.SubagentModels {
				result.SubagentModels[k] = v
			}
		}
	}

	return result
}

// ResolveTranscriptionConfig resolves the active gateway BaseURL, APIKey, and candidate models
// for audio transcription by combining config.toml with environment variables.
func ResolveTranscriptionConfig(workspacePath string) (baseURL string, apiKey string, candidateModels []string) {
	cfg := ResolveGrokConfig(workspacePath)

	// Determine active model ID
	activeModelID := "9router"
	if cfg.DefaultModel != "" {
		activeModelID = cfg.DefaultModel
	}

	// Extract active model config
	activeModelCfg, hasActive := cfg.Models[activeModelID]
	if !hasActive && len(cfg.Models) > 0 {
		// Fallback to first available model config that has an API key or BaseURL
		for _, m := range cfg.Models {
			if m.APIKey != "" || m.BaseURL != "" {
				activeModelCfg = m
				hasActive = true
				break
			}
		}
	}

	// Resolve BaseURL
	if envURL := os.Getenv("NINEROUTER_URL"); envURL != "" {
		baseURL = envURL
	} else if envBase := os.Getenv("OPENAI_BASE_URL"); envBase != "" {
		baseURL = envBase
	} else if hasActive && activeModelCfg.BaseURL != "" {
		baseURL = activeModelCfg.BaseURL
	} else {
		baseURL = "http://127.0.0.1:20128"
	}
	// Normalize BaseURL
	baseURL = strings.TrimSuffix(strings.TrimSuffix(strings.TrimSpace(baseURL), "/"), "/v1")

	// Resolve APIKey
	if envKey := os.Getenv("NINEROUTER_KEY"); envKey != "" {
		apiKey = envKey
	} else if envKey := os.Getenv("JCODE_9ROUTER_API_KEY"); envKey != "" {
		apiKey = envKey
	} else if envKey := os.Getenv("ANTHROPIC_AUTH_TOKEN"); envKey != "" {
		apiKey = envKey
	} else if envKey := os.Getenv("OPENAI_API_KEY"); envKey != "" {
		apiKey = envKey
	} else if envKey := os.Getenv("GROK_API_KEY"); envKey != "" {
		apiKey = envKey
	} else if envKey := os.Getenv("XAI_API_KEY"); envKey != "" {
		apiKey = envKey
	} else if hasActive && activeModelCfg.APIKey != "" {
		apiKey = activeModelCfg.APIKey
	}

	// Build candidate models list
	modelSet := make(map[string]bool)
	var candidates []string

	addCandidate := func(model string) {
		model = strings.TrimSpace(model)
		if model != "" && !modelSet[model] {
			modelSet[model] = true
			candidates = append(candidates, model)
		}
	}

	// 1. Configured underlying model name from active model table (e.g. "geminigacor")
	if hasActive && activeModelCfg.Model != "" {
		addCandidate(activeModelCfg.Model)
	}

	// 2. Fallback standard multimodal audio fast models
	standardFastAudioModels := []string{
		"geminigacor",
		"ag/gemini-3.7-flash-low",
		"ag/gemini-3.8-flash-low",
		"gemini/gemini-3.7-flash",
		"ag/gemini-3.7-flash-high",
		"ag/gemini-3.8-flash",
		"gemini-2.5-flash",
		"gpt-4o-audio-preview",
		"gemini-1.5-flash",
	}

	for _, m := range standardFastAudioModels {
		addCandidate(m)
	}

	return baseURL, apiKey, candidates
}
