package grokrunner

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ModelInfo describes a model discovered from Grok CLI or config
type ModelInfo struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	ModelTarget   string `json:"modelTarget,omitempty"`
	ContextWindow int    `json:"contextWindow,omitempty"`
	IsDefault     bool   `json:"isDefault"`
}

// DiscoverAvailableModels queries the local Grok CLI (`grok models`) and config.toml
// and returns all available models (including custom models like geminigacor).
// Fully cross-platform supported on Windows and macOS.
func (r *Runner) DiscoverAvailableModels() []ModelInfo {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	binPath := r.grokBinaryPath
	if binPath == "" || binPath == "grok" {
		binPath = ResolveGrokBinary()
		r.grokBinaryPath = binPath
	}

	cfg := ResolveGrokConfig("")

	cmd := exec.CommandContext(ctx, binPath, "models")
	cmd.Env = EnsureExecEnvironment()
	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &outBuf

	var models []ModelInfo
	if err := cmd.Run(); err == nil {
		models = parseModelsOutput(outBuf.String())
	}

	if len(models) == 0 {
		models = []ModelInfo{
			{ID: "9router", Name: "9Router", Description: "Primary router model", IsDefault: true},
			{ID: "9router-general-purpose", Name: "9Router General Purpose", Description: "General tasks & code editing", IsDefault: false},
			{ID: "9router-explore", Name: "9Router Explore", Description: "Exploration & code search", IsDefault: false},
			{ID: "9router-plan", Name: "9Router Plan", Description: "Planning & architecture", IsDefault: false},
		}
	}

	// Merge model definitions from config.toml (e.g. underlying target models like geminigacor)
	modelMap := make(map[string]bool)
	for i := range models {
		modelMap[models[i].ID] = true
		if mCfg, exists := cfg.Models[models[i].ID]; exists {
			if mCfg.Model != "" {
				models[i].ModelTarget = mCfg.Model
				models[i].Description = fmt.Sprintf("%s (%s)", models[i].Description, mCfg.Model)
			}
			if mCfg.ContextWindow > 0 {
				models[i].ContextWindow = mCfg.ContextWindow
			}
		}
	}

	// Add any custom models declared in config.toml that were not returned by grok models CLI
	for id, mCfg := range cfg.Models {
		if !modelMap[id] {
			desc := "Configured model"
			if mCfg.Model != "" {
				desc = fmt.Sprintf("Target: %s", mCfg.Model)
			}
			models = append(models, ModelInfo{
				ID:            id,
				Name:          mCfg.Name,
				Description:   desc,
				ModelTarget:   mCfg.Model,
				ContextWindow: mCfg.ContextWindow,
				IsDefault:     cfg.DefaultModel == id,
			})
			modelMap[id] = true
		}
	}

	return models
}

// parseModelsOutput parses lines from `grok models`
func parseModelsOutput(output string) []ModelInfo {
	var models []ModelInfo
	scanner := bufio.NewScanner(strings.NewReader(output))
	inModelsSection := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(strings.ToLower(line), "available models:") {
			inModelsSection = true
			continue
		}

		if inModelsSection {
			if strings.HasPrefix(line, "*") || strings.HasPrefix(line, "-") {
				raw := strings.TrimSpace(strings.TrimLeft(line, "*- "))
				isDefault := strings.Contains(raw, "(default)") || strings.HasPrefix(line, "*")
				id := strings.TrimSpace(strings.ReplaceAll(raw, "(default)", ""))

				if id != "" {
					name := id
					desc := "Grok CLI model"
					if strings.Contains(id, "plan") {
						desc = "High-depth reasoning & planning"
					} else if strings.Contains(id, "explore") {
						desc = "Fast codebase search & discovery"
					} else if strings.Contains(id, "general-purpose") {
						desc = "Balanced agentic & code execution"
					} else if id == "9router" {
						desc = "Standard 9Router gateway model"
					}

					models = append(models, ModelInfo{
						ID:          id,
						Name:        name,
						Description: desc,
						IsDefault:   isDefault,
					})
				}
			}
		}
	}

	return models
}
