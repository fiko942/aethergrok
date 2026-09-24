package grokrunner

import (
	"bufio"
	"bytes"
	"context"
	"os/exec"
	"strings"
	"time"
)

// ModelInfo describes a model discovered from Grok CLI or config
type ModelInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsDefault   bool   `json:"isDefault"`
}

// DiscoverAvailableModels queries the local Grok CLI (`grok models`) and returns discovered models
func (r *Runner) DiscoverAvailableModels() []ModelInfo {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, r.grokBinaryPath, "models")
	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &outBuf

	if err := cmd.Run(); err == nil {
		models := parseModelsOutput(outBuf.String())
		if len(models) > 0 {
			return models
		}
	}

	// Fallback standard default models
	return []ModelInfo{
		{ID: "9router", Name: "9Router (Default)", Description: "Primary router model", IsDefault: true},
		{ID: "9router-general-purpose", Name: "9Router General Purpose", Description: "General tasks & code editing", IsDefault: false},
		{ID: "9router-explore", Name: "9Router Explore", Description: "Exploration & code search", IsDefault: false},
		{ID: "9router-plan", Name: "9Router Plan", Description: "Planning & architecture", IsDefault: false},
	}
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
