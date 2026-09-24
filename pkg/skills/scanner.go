package skills

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParseSkillFile reads a SKILL.md file, extracts YAML frontmatter and prompt body
func ParseSkillFile(filePath string) (*Skill, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read skill file: %w", err)
	}

	frontmatter, body, err := extractFrontmatter(data)
	if err != nil {
		return nil, fmt.Errorf("failed to extract frontmatter: %w", err)
	}

	var meta SkillFrontmatter
	if len(frontmatter) > 0 {
		if err := yaml.Unmarshal(frontmatter, &meta); err != nil {
			return nil, fmt.Errorf("failed to parse yaml frontmatter: %w", err)
		}
	}

	dir := filepath.Dir(filePath)
	dirName := filepath.Base(dir)

	name := meta.Name
	if strings.TrimSpace(name) == "" {
		name = dirName
	}

	// Infer or resolve category
	cat := inferCategory(name, meta.Category, meta.Description, meta.Tags)

	// Combine tags and actions
	tagsMap := make(map[string]struct{})
	for _, t := range meta.Tags {
		t = strings.TrimSpace(t)
		if t != "" {
			tagsMap[t] = struct{}{}
		}
	}
	for _, a := range meta.Actions {
		a = strings.TrimSpace(a)
		if a != "" {
			tagsMap[a] = struct{}{}
		}
	}

	// Include category as tag if not present
	if cat != CategoryAll {
		tagsMap[strings.ToLower(string(cat))] = struct{}{}
	}

	var tags []string
	for t := range tagsMap {
		tags = append(tags, t)
	}

	// Identify scope from path
	scope := "custom"
	normalizedPath := filepath.ToSlash(filePath)
	if strings.Contains(normalizedPath, "/.grok/skills/") {
		scope = "grok"
	} else if strings.Contains(normalizedPath, "/.agents/skills/") {
		scope = "agents"
	}

	// Create a stable ID
	id := fmt.Sprintf("%s:%s", scope, name)

	return &Skill{
		ID:          id,
		Name:        name,
		Description: meta.Description,
		Category:    cat,
		Tags:        tags,
		Actions:     meta.Actions,
		Path:        filePath,
		Directory:   dir,
		Scope:       scope,
		Prompt:      strings.TrimSpace(body),
	}, nil
}

// extractFrontmatter separates YAML frontmatter enclosed in --- delimiters from markdown body
func extractFrontmatter(data []byte) ([]byte, string, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	
	if !scanner.Scan() {
		return nil, "", nil
	}

	firstLine := strings.TrimSpace(scanner.Text())
	if firstLine != "---" {
		// No frontmatter, entire content is body
		return nil, string(data), nil
	}

	var frontmatterLines []string
	foundEnd := false

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "---" {
			foundEnd = true
			break
		}
		frontmatterLines = append(frontmatterLines, line)
	}

	if !foundEnd {
		// No closing delimiter, return entire content as body
		return nil, string(data), nil
	}

	var bodyLines []string
	for scanner.Scan() {
		bodyLines = append(bodyLines, scanner.Text())
	}

	return []byte(strings.Join(frontmatterLines, "\n")), strings.Join(bodyLines, "\n"), nil
}

// inferCategory maps names and descriptions to standard UI categories
func inferCategory(name, explicitCat, desc string, tags []string) SkillCategory {
	switch strings.ToLower(strings.TrimSpace(explicitCat)) {
	case "frontend", "ui", "web":
		return CategoryFrontend
	case "backend", "api", "database", "server":
		return CategoryBackend
	case "design", "ux", "creative":
		return CategoryDesign
	case "agents", "agent", "subagent", "orchestration":
		return CategoryAgents
	case "tools", "mcp", "tool":
		return CategoryTools
	}

	combined := strings.ToLower(fmt.Sprintf("%s %s %s", name, desc, strings.Join(tags, " ")))

	if strings.Contains(combined, "design") || strings.Contains(combined, "taste") || strings.Contains(combined, "drawio") ||
		strings.Contains(combined, "brand") || strings.Contains(combined, "banner") || strings.Contains(combined, "slides") {
		return CategoryDesign
	}

	if strings.Contains(combined, "frontend") || strings.Contains(combined, "css") || strings.Contains(combined, "html") ||
		strings.Contains(combined, "svelte") || strings.Contains(combined, "react") || strings.Contains(combined, "vue") ||
		strings.Contains(combined, "landing") || strings.Contains(combined, "web-artifacts") {
		return CategoryFrontend
	}

	if strings.Contains(combined, "backend") || strings.Contains(combined, "database") || strings.Contains(combined, "sql") ||
		strings.Contains(combined, "server") || strings.Contains(combined, "api") || strings.Contains(combined, "grpc") ||
		strings.Contains(combined, "gopls") || strings.Contains(combined, "lsp") {
		return CategoryBackend
	}

	if strings.Contains(combined, "agent") || strings.Contains(combined, "subagent") || strings.Contains(combined, "autonomous") ||
		strings.Contains(combined, "browser-use") || strings.Contains(combined, "cua") {
		return CategoryAgents
	}

	// Default to Tools (e.g. mcp, git, test, code-review, commit, debug, inspect, terminal)
	return CategoryTools
}

// ScanSkillDirectories searches provided root directories for SKILL.md files
func ScanSkillDirectories(directories []string) ([]Skill, []error) {
	var skills []Skill
	var errs []error
	seenNames := make(map[string]bool)

	for _, dir := range directories {
		// Expand ~ to user home directory if present
		expandedDir := expandHome(dir)
		info, err := os.Stat(expandedDir)
		if err != nil || !info.IsDir() {
			continue
		}

		err = filepath.WalkDir(expandedDir, func(path string, d os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return nil
			}

			// Do not descend into hidden directories other than root matching
			if d.IsDir() && strings.HasPrefix(d.Name(), ".") && path != expandedDir {
				return filepath.SkipDir
			}

			if !d.IsDir() && strings.EqualFold(d.Name(), "SKILL.md") {
				skill, err := ParseSkillFile(path)
				if err != nil {
					errs = append(errs, fmt.Errorf("error parsing %s: %w", path, err))
					return nil
				}

				// Deduplicate by name if already seen (e.g. ~/.grok prioritized over ~/.agents)
				if !seenNames[skill.Name] {
					seenNames[skill.Name] = true
					skills = append(skills, *skill)
				}
			}
			return nil
		})

		if err != nil {
			errs = append(errs, err)
		}
	}

	return skills, errs
}

// expandHome replaces leading ~ with user home directory
func expandHome(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if path == "~" {
		return home
	}
	if strings.HasPrefix(path, "~/") || strings.HasPrefix(path, `~\`) {
		return filepath.Join(home, path[2:])
	}
	return path
}
