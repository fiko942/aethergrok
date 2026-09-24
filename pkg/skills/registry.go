package skills

import (
	"sort"
	"strings"
	"sync"
)

// Registry manages in-memory cached skills and search indexing
type Registry struct {
	mu           sync.RWMutex
	skills       []Skill
	directories  []string
	indexed      bool
}

// DefaultDirectories contains the canonical paths per specification
var DefaultDirectories = []string{
	"~/.grok/skills/",
	"~/.agents/skills/",
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

// Search performs filtering on skills by search query and category tab
func (r *Registry) Search(query string, category string) []Skill {
	all := r.GetAll()
	trimmedQuery := strings.ToLower(strings.TrimSpace(query))
	targetCat := strings.TrimSpace(category)

	var results []Skill

	for _, s := range all {
		// Category filtering
		if targetCat != "" && !strings.EqualFold(targetCat, string(CategoryAll)) {
			if !strings.EqualFold(string(s.Category), targetCat) {
				continue
			}
		}

		// Text query filtering across name, description, tags, and actions
		if trimmedQuery != "" {
			nameMatch := strings.Contains(strings.ToLower(s.Name), trimmedQuery)
			descMatch := strings.Contains(strings.ToLower(s.Description), trimmedQuery)
			
			tagMatch := false
			for _, tag := range s.Tags {
				if strings.Contains(strings.ToLower(tag), trimmedQuery) {
					tagMatch = true
					break
				}
			}

			actionMatch := false
			for _, act := range s.Actions {
				if strings.Contains(strings.ToLower(act), trimmedQuery) {
					actionMatch = true
					break
				}
			}

			if !nameMatch && !descMatch && !tagMatch && !actionMatch {
				continue
			}
		}

		results = append(results, s)
	}

	return results
}
