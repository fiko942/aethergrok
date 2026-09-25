package skills

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// DiscoveredSkill represents an individual skill found in a repository
type DiscoveredSkill struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Category     string   `json:"category"`
	Tags         []string `json:"tags"`
	RelativePath string   `json:"relativePath"` // Relative path to directory containing SKILL.md
	SkillFile    string   `json:"skillFile"`    // Path relative to repo root
	Prereqs      []string `json:"prereqs"`
	Commands     []string `json:"commands"`
}

// SkillAnalysisResult represents full repository scan and dependency analysis
type SkillAnalysisResult struct {
	RepoURL          string            `json:"repoUrl"`
	RepoName         string            `json:"repoName"`
	TempPath         string            `json:"tempPath"`
	Skills           []DiscoveredSkill `json:"skills"`
	GlobalPrereqs    []string          `json:"globalPrereqs"`
	SuggestedScripts []string          `json:"suggestedScripts"`
}

// SkillInstallPayload represents user's selection for installation
type SkillInstallPayload struct {
	TempPath     string   `json:"tempPath"`
	SkillPaths   []string `json:"skillPaths"` // Relative paths within TempPath
	TargetScope  string   `json:"targetScope"` // "grok" (~/.grok/skills) or "agents" (~/.agents/skills)
}

// SkillInstallResult is the outcome of installing skills
type SkillInstallResult struct {
	Success        bool     `json:"success"`
	InstalledCount int      `json:"installedCount"`
	InstalledPaths []string `json:"installedPaths"`
	Errors         []string `json:"errors,omitempty"`
}

// ScanGitHubRepo clones or checks out a GitHub repo into /tmp/aethergrok-skills/ and extracts skills
func ScanGitHubRepo(ctx context.Context, repoInput string) (*SkillAnalysisResult, error) {
	cleanURL := strings.TrimSpace(repoInput)
	if cleanURL == "" {
		return nil, fmt.Errorf("repository URL or identifier is empty")
	}

	// Normalize input: e.g. "owner/repo" or "https://github.com/owner/repo" or "github.com/owner/repo"
	if !strings.HasPrefix(cleanURL, "http://") && !strings.HasPrefix(cleanURL, "https://") && !strings.HasPrefix(cleanURL, "git@") {
		cleanURL = "https://github.com/" + strings.TrimPrefix(cleanURL, "/")
	}

	// Determine repo short name
	parts := strings.Split(strings.TrimSuffix(cleanURL, ".git"), "/")
	repoName := "skills-repo"
	if len(parts) >= 2 {
		repoName = parts[len(parts)-2] + "-" + parts[len(parts)-1]
	} else if len(parts) == 1 {
		repoName = parts[0]
	}

	tempDir := filepath.Join(os.TempDir(), "aethergrok-skills", fmt.Sprintf("%s-%d", repoName, time.Now().Unix()))
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Run git clone with depth 1
	cloneCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	cmd := exec.CommandContext(cloneCtx, "git", "clone", "--depth", "1", cleanURL, tempDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		_ = os.RemoveAll(tempDir)
		return nil, fmt.Errorf("failed to clone repository (%s): %s", err.Error(), strings.TrimSpace(string(output)))
	}

	result := &SkillAnalysisResult{
		RepoURL:          cleanURL,
		RepoName:         repoName,
		TempPath:         tempDir,
		Skills:           make([]DiscoveredSkill, 0),
		GlobalPrereqs:    make([]string, 0),
		SuggestedScripts: make([]string, 0),
	}

	// Scan for SKILL.md / skill.md with deduplication
	seenSkillPaths := make(map[string]bool)
	err = filepath.Walk(tempDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if info.IsDir() {
			name := info.Name()
			if name == ".git" || name == "node_modules" || name == "vendor" || name == ".venv" {
				return filepath.SkipDir
			}
			return nil
		}

		fileName := strings.ToLower(info.Name())
		if fileName == "skill.md" || strings.HasSuffix(fileName, ".skill.md") {
			relFile, relErr := filepath.Rel(tempDir, path)
			if relErr != nil {
				return nil
			}

			skillDir := filepath.Dir(path)
			relDir, _ := filepath.Rel(tempDir, skillDir)
			if relDir == "." {
				relDir = ""
			}

			skillMeta := parseSkillFile(path, relFile, relDir)
			key := fmt.Sprintf("%s:%s", strings.ToLower(skillMeta.Name), relDir)
			if !seenSkillPaths[key] {
				seenSkillPaths[key] = true
				result.Skills = append(result.Skills, skillMeta)
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking repository: %w", err)
	}

	// Detect global repository prerequisites
	detectRepoPrereqs(tempDir, result)

	return result, nil
}

// parseSkillFile reads frontmatter and content from SKILL.md
func parseSkillFile(filePath, relFile, relDir string) DiscoveredSkill {
	skill, err := ParseSkillFile(filePath)
	if err != nil || skill == nil {
		dirName := filepath.Base(filepath.Dir(filePath))
		return DiscoveredSkill{
			Name:         dirName,
			Description:  "Skill at " + relFile,
			Category:     "Tools",
			RelativePath: relDir,
			SkillFile:    relFile,
		}
	}

	content := ""
	if data, err := os.ReadFile(filePath); err == nil {
		content = string(data)
	}

	prereqs, cmds := extractSkillRequirements(content, filepath.Dir(filePath))

	return DiscoveredSkill{
		Name:         skill.Name,
		Description:  skill.Description,
		Category:     string(skill.Category),
		Tags:         skill.Tags,
		RelativePath: relDir,
		SkillFile:    relFile,
		Prereqs:      prereqs,
		Commands:     cmds,
	}
}

// extractSkillRequirements inspects markdown content and local files for commands/tools
func extractSkillRequirements(content string, dirPath string) ([]string, []string) {
	prereqs := make([]string, 0)
	cmds := make([]string, 0)

	lower := strings.ToLower(content)

	if strings.Contains(lower, "npm install") || strings.Contains(lower, "node") || fileExists(filepath.Join(dirPath, "package.json")) {
		prereqs = append(prereqs, "Node.js (npm)")
		if fileExists(filepath.Join(dirPath, "package.json")) {
			cmds = append(cmds, "npm install")
		}
	}
	if strings.Contains(lower, "pip install") || strings.Contains(lower, "python") || fileExists(filepath.Join(dirPath, "requirements.txt")) {
		prereqs = append(prereqs, "Python 3 (pip)")
		if fileExists(filepath.Join(dirPath, "requirements.txt")) {
			cmds = append(cmds, "pip install -r requirements.txt")
		}
	}
	if strings.Contains(lower, "brew install") || strings.Contains(lower, "homebrew") {
		prereqs = append(prereqs, "Homebrew")
	}
	if strings.Contains(lower, "cargo") || fileExists(filepath.Join(dirPath, "Cargo.toml")) {
		prereqs = append(prereqs, "Rust (cargo)")
	}

	return prereqs, cmds
}

// detectRepoPrereqs inspects root files of the cloned repository
func detectRepoPrereqs(repoDir string, result *SkillAnalysisResult) {
	if fileExists(filepath.Join(repoDir, "package.json")) {
		result.GlobalPrereqs = append(result.GlobalPrereqs, "Node.js runtime")
		result.SuggestedScripts = append(result.SuggestedScripts, "npm install")
	}
	if fileExists(filepath.Join(repoDir, "requirements.txt")) {
		result.GlobalPrereqs = append(result.GlobalPrereqs, "Python 3 environment")
		result.SuggestedScripts = append(result.SuggestedScripts, "pip install -r requirements.txt")
	}
	if fileExists(filepath.Join(repoDir, "Makefile")) {
		result.GlobalPrereqs = append(result.GlobalPrereqs, "Make / Build tools")
	}

	// Deduplicate
	for _, skill := range result.Skills {
		for _, p := range skill.Prereqs {
			if !containsString(result.GlobalPrereqs, p) {
				result.GlobalPrereqs = append(result.GlobalPrereqs, p)
			}
		}
		for _, c := range skill.Commands {
			if !containsString(result.SuggestedScripts, c) {
				result.SuggestedScripts = append(result.SuggestedScripts, c)
			}
		}
	}
}

// InstallDiscoveredSkills copies selected skill directories into target scope directory
func InstallDiscoveredSkills(payload SkillInstallPayload) (*SkillInstallResult, error) {
	if payload.TempPath == "" {
		return nil, fmt.Errorf("temporary path is missing")
	}

	var targetBase string
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("unable to determine home directory: %w", err)
	}

	if payload.TargetScope == "agents" {
		targetBase = filepath.Join(home, ".agents", "skills")
	} else {
		targetBase = filepath.Join(home, ".grok", "skills")
	}

	if err := os.MkdirAll(targetBase, 0755); err != nil {
		return nil, fmt.Errorf("failed to create destination skills directory: %w", err)
	}

	res := &SkillInstallResult{
		Success:        true,
		InstalledCount: 0,
		InstalledPaths: make([]string, 0),
		Errors:         make([]string, 0),
	}

	for _, relPath := range payload.SkillPaths {
		srcDir := filepath.Join(payload.TempPath, relPath)
		skillName := filepath.Base(srcDir)
		
		// If skill is in repo root or relative path is empty, try to resolve name from SKILL.md
		if skillName == "." || skillName == "/" || skillName == "" || strings.HasPrefix(skillName, "aethergrok-skills") {
			candidateSkillFile := filepath.Join(srcDir, "SKILL.md")
			if !fileExists(candidateSkillFile) {
				candidateSkillFile = filepath.Join(srcDir, "skill.md")
			}
			if fileExists(candidateSkillFile) {
				if parsed, parseErr := ParseSkillFile(candidateSkillFile); parseErr == nil && parsed != nil && parsed.Name != "" {
					skillName = parsed.Name
				}
			}
			if skillName == "." || skillName == "/" || skillName == "" || strings.HasPrefix(skillName, "aethergrok-skills") {
				skillName = filepath.Base(payload.TempPath)
			}
		}

		destDir := filepath.Join(targetBase, skillName)

		if err := copyDirectory(srcDir, destDir); err != nil {
			res.Errors = append(res.Errors, fmt.Sprintf("failed to install %s: %s", skillName, err.Error()))
			res.Success = false
		} else {
			res.InstalledCount++
			res.InstalledPaths = append(res.InstalledPaths, destDir)
		}
	}

	return res, nil
}

// CleanupTempSkills removes the temporary clone directory
func CleanupTempSkills(tempPath string) error {
	if tempPath != "" && strings.Contains(tempPath, "aethergrok-skills") {
		return os.RemoveAll(tempPath)
	}
	return nil
}

// ExecuteSetupCommand runs an installation shell command and streams lines via callback
func ExecuteSetupCommand(ctx context.Context, workDir, commandLine string, onLog func(line string)) error {
	parts := strings.Fields(commandLine)
	if len(parts) == 0 {
		return nil
	}

	cmd := exec.CommandContext(ctx, "sh", "-c", commandLine)
	if workDir != "" && fileExists(workDir) {
		cmd.Dir = workDir
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	readLines := func(r io.Reader, prefix string) {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			if onLog != nil {
				onLog(prefix + scanner.Text())
			}
		}
	}

	go readLines(stdout, "")
	go readLines(stderr, "[stderr] ")

	return cmd.Wait()
}

func copyDirectory(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("source path inaccessible: %w", err)
	}

	// If source is a single file rather than a directory
	if !srcInfo.IsDir() {
		return copyFile(src, filepath.Join(dst, filepath.Base(src)))
	}

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(dst, rel)

		// Check if it is a directory or symlink pointing to a directory
		evalInfo, statErr := os.Stat(path)
		if statErr == nil && evalInfo.IsDir() {
			name := evalInfo.Name()
			if name == ".git" {
				return filepath.SkipDir
			}
			return os.MkdirAll(targetPath, 0755)
		}

		// Handle symlink
		if info.Mode()&os.ModeSymlink != 0 {
			resolvedTarget, resolveErr := os.Readlink(path)
			if resolveErr == nil {
				if statErr == nil && !evalInfo.IsDir() {
					return copyFile(path, targetPath)
				}
				_ = os.MkdirAll(filepath.Dir(targetPath), 0755)
				_ = os.Symlink(resolvedTarget, targetPath)
				return nil
			}
		}

		if info.IsDir() {
			name := info.Name()
			if name == ".git" {
				return filepath.SkipDir
			}
			return os.MkdirAll(targetPath, 0755)
		}

		return copyFile(path, targetPath)
	})
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	return nil
}

func fileExists(p string) bool {
	info, err := os.Stat(p)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func containsString(list []string, item string) bool {
	for _, l := range list {
		if strings.EqualFold(l, item) {
			return true
		}
	}
	return false
}
