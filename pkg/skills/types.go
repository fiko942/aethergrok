package skills

// SkillCategory represents standard categorizations for skills in the UI
type SkillCategory string

const (
	CategoryAll      SkillCategory = "All"
	CategoryFrontend SkillCategory = "Frontend"
	CategoryBackend  SkillCategory = "Backend"
	CategoryDesign   SkillCategory = "Design"
	CategoryAgents   SkillCategory = "Agents"
	CategoryTools    SkillCategory = "Tools"
)

// SkillFrontmatter defines parsed YAML header attributes from SKILL.md
type SkillFrontmatter struct {
	Name        string   `yaml:"name" json:"name"`
	Description string   `yaml:"description" json:"description"`
	License     string   `yaml:"license,omitempty" json:"license,omitempty"`
	Version     string   `yaml:"version,omitempty" json:"version,omitempty"`
	Category    string   `yaml:"category,omitempty" json:"category,omitempty"`
	Tags        []string `yaml:"tags,omitempty" json:"tags,omitempty"`
	Actions     []string `yaml:"actions,omitempty" json:"actions,omitempty"`
}

// Skill represents a discovered skill with metadata and location
type Skill struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Category    SkillCategory `json:"category"`
	Tags        []string      `json:"tags"`
	Actions     []string      `json:"actions"`
	Path        string        `json:"path"`
	Directory   string        `json:"directory"`
	Scope       string        `json:"scope"` // "grok" or "agents" or "custom"
	Prompt      string        `json:"prompt,omitempty"`
}
