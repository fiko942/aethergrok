package skills

import (
	"sort"
	"strings"
	"sync"
)

// Registry manages in-memory cached skills and search indexing
type Registry struct {
	mu          sync.RWMutex
	skills      []Skill
	directories []string
	indexed     bool
}

// DefaultDirectories contains the canonical paths per specification
var DefaultDirectories = []string{
	"~/.grok/skills/",
	"~/.grok/bundled/skills/",
	"~/.grok/installed-plugins/",
	"~/.agents/skills/",
	"~/.agents/installed-plugins/",
	"~/.claude/skills/",
	"~/.claude/installed-plugins/",
	"~/.codex/skills/",
}

// NewRegistry initializes a registry with specified search directories
func NewRegistry(directories ...string) *Registry {
	dirs := directories
	if len(dirs) == 0 {
		dirs = DefaultDirectories
	}
	return &Registry{
		directories: dirs,
		skills:      make([]Skill, 0),
	}
}

// ScanSkills rescans configured directories and refreshes the internal skill list
func (r *Registry) ScanSkills() ([]Skill, []error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	skills, errs := ScanSkillDirectories(r.directories)

	// Sort alphabetically by name
	sort.Slice(skills, func(i, j int) bool {
		return strings.ToLower(skills[i].Name) < strings.ToLower(skills[j].Name)
	})

	r.skills = skills
	r.indexed = true
	return skills, errs
}

// GetAll returns all loaded skills, triggering initial scan if needed
func (r *Registry) GetAll() []Skill {
	r.mu.RLock()
	if r.indexed {
		defer r.mu.RUnlock()
		out := make([]Skill, len(r.skills))
		copy(out, r.skills)
		return out
	}
	r.mu.RUnlock()

	r.ScanSkills()

	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Skill, len(r.skills))
	copy(out, r.skills)
	return out
}

// CalculateSkillMatchScore computes a prioritized relevance score for a skill against a query.
// Priority tiers:
// 1. Exact Name match (1000)
// 2. Name StartsWith query (800 - len)
// 3. Name Word/Kebab Segment StartsWith query (600 - len) (e.g. query "superpowers" matches segment in "using-superpowers")
// 4. Name Includes query (400 - len)
// 5. ID / Scope matches (300)
// 6. Tags or Actions match (250 / 200)
// 7. Description Word StartsWith query (100)
// 8. Description Includes query (50)
func CalculateSkillMatchScore(s *Skill, query string) int {
	cleanQuery := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(query, "/")))
	if cleanQuery == "" {
		return 1
	}

	name := strings.ToLower(s.Name)

	// 1. Exact match on name
	if name == cleanQuery {
		return 1000
	}

	// 2. Name starts with query
	if strings.HasPrefix(name, cleanQuery) {
		return 800 - len(name)
	}

	// 3. Name word / kebab-case segment starts with query
	segments := strings.FieldsFunc(name, func(r rune) bool {
		return r == '-' || r == '_' || r == ' ' || r == ':' || r == '/' || r == '.'
	})
	for _, seg := range segments {
		if strings.HasPrefix(seg, cleanQuery) {
			return 600 - len(name)
		}
	}

	// 4. Name contains query anywhere
	if strings.Contains(name, cleanQuery) {
		return 400 - len(name)
	}

	// 5. ID / Scope contains query
	if strings.Contains(strings.ToLower(s.ID), cleanQuery) {
		return 300
	}

	// 6. Tags or Actions match
	for _, tag := range s.Tags {
		tLower := strings.ToLower(tag)
		if strings.HasPrefix(tLower, cleanQuery) {
			return 250
		}
		if strings.Contains(tLower, cleanQuery) {
			return 200
		}
	}
	for _, act := range s.Actions {
		aLower := strings.ToLower(act)
		if strings.HasPrefix(aLower, cleanQuery) {
			return 250
		}
		if strings.Contains(aLower, cleanQuery) {
			return 200
		}
	}

	// 7. Description word starts with query
	desc := strings.ToLower(s.Description)
	descWords := strings.FieldsFunc(desc, func(r rune) bool {
		return r == ' ' || r == '\t' || r == '\n' || r == ',' || r == '.' || r == '-' || r == '(' || r == ')' || r == '[' || r == ']'
	})
	for _, w := range descWords {
		if strings.HasPrefix(w, cleanQuery) {
			return 100
		}
	}

	// 8. Description contains query anywhere
	if strings.Contains(desc, cleanQuery) {
		return 50
	}

	return 0
}

type scoredSkill struct {
	skill Skill
	score int
}

// Search performs filtering on skills by search query and category tab with prioritized ranking
func (r *Registry) Search(query string, category string) []Skill {
	all := r.GetAll()
	trimmedQuery := strings.TrimSpace(query)
	targetCat := strings.TrimSpace(category)

	var scored []scoredSkill

	for _, s := range all {
		// Category filtering
		if targetCat != "" && !strings.EqualFold(targetCat, string(CategoryAll)) {
			if !strings.EqualFold(string(s.Category), targetCat) {
				continue
			}
		}

		score := CalculateSkillMatchScore(&s, trimmedQuery)
		if score > 0 {
			scored = append(scored, scoredSkill{
				skill: s,
				score: score,
			})
		}
	}

	// Sort by score descending, then by name length ascending, then alphabetical
	sort.Slice(scored, func(i, j int) bool {
		if scored[i].score != scored[j].score {
			return scored[i].score > scored[j].score
		}
		if len(scored[i].skill.Name) != len(scored[j].skill.Name) {
			return len(scored[i].skill.Name) < len(scored[j].skill.Name)
		}
		return strings.ToLower(scored[i].skill.Name) < strings.ToLower(scored[j].skill.Name)
	})

	results := make([]Skill, len(scored))
	for i, item := range scored {
		results[i] = item.skill
	}

	return results
}
