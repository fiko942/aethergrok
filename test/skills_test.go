package test

import (
	"os"
	"path/filepath"
	"testing"

	"aethergrok/pkg/skills"
)

func TestSkills_ParseSkillFile(t *testing.T) {
	tmpDir := t.TempDir()
	skillDir := filepath.Join(tmpDir, "design-taste-frontend")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	skillFilePath := filepath.Join(skillDir, "SKILL.md")
	content := `---
name: design-taste-frontend
description: Anti-slop frontend skill for landing pages, portfolios, and redesigns.
tags:
  - design
  - frontend
  - ui
actions:
  - audit-ui
  - redesign
---

# tasteskill: Anti-Slop Frontend Skill

Follow intentional design heuristics and audit typography.
`
	if err := os.WriteFile(skillFilePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write skill file: %v", err)
	}

	skill, err := skills.ParseSkillFile(skillFilePath)
	if err != nil {
		t.Fatalf("ParseSkillFile failed: %v", err)
	}

	if skill.Name != "design-taste-frontend" {
		t.Errorf("expected name 'design-taste-frontend', got '%s'", skill.Name)
	}
	if skill.Category != skills.CategoryDesign {
		t.Errorf("expected category '%s', got '%s'", skills.CategoryDesign, skill.Category)
	}
	if len(skill.Tags) == 0 {
		t.Errorf("expected parsed tags, got empty")
	}
	if len(skill.Actions) != 2 {
		t.Errorf("expected 2 actions, got %d", len(skill.Actions))
	}
	if skill.Prompt == "" {
		t.Errorf("expected prompt body to be non-empty")
	}
}

func TestSkills_ScannerAndRegistry(t *testing.T) {
	tmpRoot := t.TempDir()
	grokDir := filepath.Join(tmpRoot, ".grok", "skills")
	agentsDir := filepath.Join(tmpRoot, ".agents", "skills")

	// Create skill 1 in ~/.grok/skills/
	skill1Dir := filepath.Join(grokDir, "code-review")
	if err := os.MkdirAll(skill1Dir, 0755); err != nil {
		t.Fatalf("failed to create skill1 dir: %v", err)
	}
	skill1Content := `---
name: code-review
description: Review diff for bugs, regressions, and tests.
category: Tools
tags:
  - git
  - review
---

# Code Review Prompt
`
	if err := os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), []byte(skill1Content), 0644); err != nil {
		t.Fatalf("failed to write skill1: %v", err)
	}

	// Create skill 2 in ~/.agents/skills/
	skill2Dir := filepath.Join(agentsDir, "backend-api")
	if err := os.MkdirAll(skill2Dir, 0755); err != nil {
		t.Fatalf("failed to create skill2 dir: %v", err)
	}
	skill2Content := `---
name: backend-api
description: API generation tool with OpenAPI and gRPC.
category: Backend
tags:
  - api
  - grpc
---

# Backend API
`
	if err := os.WriteFile(filepath.Join(skill2Dir, "SKILL.md"), []byte(skill2Content), 0644); err != nil {
		t.Fatalf("failed to write skill2: %v", err)
	}

	reg := skills.NewRegistry(grokDir, agentsDir)
	allSkills, errs := reg.ScanSkills()
	if len(errs) > 0 {
		t.Fatalf("encountered errors scanning: %v", errs)
	}
	if len(allSkills) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(allSkills))
	}

	// Test Search with query
	results := reg.Search("review", "")
	if len(results) != 1 || results[0].Name != "code-review" {
		t.Errorf("expected 1 result 'code-review', got %v", results)
	}

	// Test Search with category filter
	backendResults := reg.Search("", "Backend")
	if len(backendResults) != 1 || backendResults[0].Name != "backend-api" {
		t.Errorf("expected 1 backend result, got %v", backendResults)
	}

	// Test Search with combined filter
	toolsResults := reg.Search("review", "Tools")
	if len(toolsResults) != 1 {
		t.Errorf("expected 1 tools result, got %d", len(toolsResults))
	}

	// Test category filter mismatch
	none := reg.Search("review", "Frontend")
	if len(none) != 0 {
		t.Errorf("expected 0 results, got %d", len(none))
	}
}
