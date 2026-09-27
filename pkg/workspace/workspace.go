package workspace

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"aethergrok/pkg/gitutil"
)

// FileItem represents a single file or directory node in the workspace
type FileItem struct {
	Name      string `json:"name"`
	Path      string `json:"path"`      // Relative path from workspace root
	FullPath  string `json:"fullPath"`  // Absolute path on disk
	IsDir     bool   `json:"isDir"`
	SizeBytes int64  `json:"sizeBytes"`
	Ext       string `json:"ext"`
}

// GitFileChange represents an uncommitted file change
type GitFileChange struct {
	Path        string `json:"path"`
	Status      string `json:"status"` // "M" (Modified), "A" (Added), "D" (Deleted), "?" (Untracked)
	AddedLines   int    `json:"addedLines"`
	RemovedLines int    `json:"removedLines"`
}

// GitStatusResult represents the current repository status
type GitStatusResult struct {
	Branch        string          `json:"branch"`
	IsClean       bool            `json:"isClean"`
	AheadCount    int             `json:"aheadCount"`
	BehindCount   int             `json:"behindCount"`
	ChangedFiles  []GitFileChange `json:"changedFiles"`
	TotalAdditions int             `json:"totalAdditions"`
	TotalDeletions int             `json:"totalDeletions"`
}

// ReadDirectory lists files and folders inside a given workspace subdirectory
func ReadDirectory(workspacePath, relativeDir string) ([]FileItem, error) {
	if workspacePath == "" {
		return nil, fmt.Errorf("workspace path cannot be empty")
	}

	targetDir := workspacePath
	if relativeDir != "" {
		targetDir = filepath.Join(workspacePath, relativeDir)
	}

	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return nil, err
	}

	var results []FileItem
	const maxEntries = 300 // Protection against freezing on massive folders like node_modules

	for _, entry := range entries {
		name := entry.Name()
		// Filter out dot files/folders by default except important configs
		if strings.HasPrefix(name, ".") && name != ".env" && name != ".gitignore" {
			continue
		}

		info, err := entry.Info()
		var size int64
		if err == nil {
			size = info.Size()
		}

		relPath := name
		if relativeDir != "" {
			relPath = filepath.Join(relativeDir, name)
		}

		isDir := entry.IsDir()
		ext := strings.ToLower(filepath.Ext(name))

		results = append(results, FileItem{
			Name:      name,
			Path:      relPath,
			FullPath:  filepath.Join(targetDir, name),
			IsDir:     isDir,
			SizeBytes: size,
			Ext:       ext,
		})

		if len(results) >= maxEntries {
			break
		}
	}

	// Sort directories first, then alphabetical
	sort.Slice(results, func(i, j int) bool {
		if results[i].IsDir != results[j].IsDir {
			return results[i].IsDir
		}
		return strings.ToLower(results[i].Name) < strings.ToLower(results[j].Name)
	})

	return results, nil
}

// ReadFileContent reads file content as text. If maxBytes > 0 it caps, if 0 it reads up to 50MB
// Supports both relative workspace paths and absolute filesystem paths.
func ReadFileContent(workspacePath, targetPath string, allowLarge bool) (string, error) {
	fullPath := targetPath
	if !filepath.IsAbs(targetPath) {
		fullPath = filepath.Join(workspacePath, targetPath)
	}

	// Expand ~ if present
	if strings.HasPrefix(fullPath, "~") {
		if homeDir, err := os.UserHomeDir(); err == nil {
			fullPath = filepath.Join(homeDir, fullPath[1:])
		}
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		return "", err
	}

	// 1MB threshold for confirmation requirement
	const standardLimit = 1024 * 1024 // 1MB
	const absoluteHardLimit = 50 * 1024 * 1024 // 50MB absolute safeguard

	if info.Size() > standardLimit && !allowLarge {
		return "", fmt.Errorf("LARGE_FILE_CONFIRM_REQUIRED:%d", info.Size())
	}

	if info.Size() > absoluteHardLimit {
		return "", fmt.Errorf("file size (%d bytes) exceeds absolute 50MB system limit", info.Size())
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return "", err
	}

	// If file contains binary null byte, return notice instead of scrambled text
	if bytes.IndexByte(data, 0) != -1 {
		return fmt.Sprintf("[Binary File Notice]\nFile '%s' is a binary file (%d bytes). Text preview is not supported.", filepath.Base(targetPath), len(data)), nil
	}

	return string(data), nil
}

// FileCheckResult holds file existence and path resolution information
type FileCheckResult struct {
	Exists    bool   `json:"exists"`
	FullPath  string `json:"fullPath"`
	RelPath   string `json:"relPath"`
	IsDir     bool   `json:"isDir"`
	SizeBytes int64  `json:"sizeBytes"`
}

// CheckFileExists checks if a file exists either in workspacePath or absolute/home filesystem path
func CheckFileExists(workspacePath, candidatePath string) *FileCheckResult {
	if strings.TrimSpace(candidatePath) == "" {
		return &FileCheckResult{Exists: false}
	}

	cleanCandidate := strings.TrimSpace(candidatePath)

	// Expand ~ if present
	resolvedPath := cleanCandidate
	if strings.HasPrefix(cleanCandidate, "~") {
		if homeDir, err := os.UserHomeDir(); err == nil {
			resolvedPath = filepath.Join(homeDir, cleanCandidate[1:])
		}
	}

	// 1. Try as absolute path
	if filepath.IsAbs(resolvedPath) {
		if info, err := os.Stat(resolvedPath); err == nil {
			return &FileCheckResult{
				Exists:    true,
				FullPath:  resolvedPath,
				RelPath:   filepath.Base(resolvedPath),
				IsDir:     info.IsDir(),
				SizeBytes: info.Size(),
			}
		}
	}

	// 2. Try relative to workspacePath
	if workspacePath != "" {
		fullWsPath := filepath.Join(workspacePath, cleanCandidate)
		if info, err := os.Stat(fullWsPath); err == nil {
			return &FileCheckResult{
				Exists:    true,
				FullPath:  fullWsPath,
				RelPath:   cleanCandidate,
				IsDir:     info.IsDir(),
				SizeBytes: info.Size(),
			}
		}
	}

	return &FileCheckResult{Exists: false}
}

// CheckMultipleFilesExists batches multiple candidate paths for performance
func CheckMultipleFilesExists(workspacePath string, candidates []string) map[string]FileCheckResult {
	results := make(map[string]FileCheckResult)
	for _, cand := range candidates {
		res := CheckFileExists(workspacePath, cand)
		if res != nil && res.Exists {
			results[cand] = *res
		}
	}
	return results
}

// GetGitStatus runs git commands to extract branch and uncommitted changes
func GetGitStatus(workspacePath string) (*GitStatusResult, error) {
	if workspacePath == "" {
		return nil, fmt.Errorf("workspace path is empty")
	}

	// Check if directory is a git repository
	if _, err := os.Stat(filepath.Join(workspacePath, ".git")); os.IsNotExist(err) {
		return &GitStatusResult{
			Branch:  "Not a git repo",
			IsClean: true,
		}, nil
	}

	// 1. Get current branch
	cmdBranch := exec.Command(gitutil.Executable(), "rev-parse", "--abbrev-ref", "HEAD")
	cmdBranch.Dir = workspacePath
	branchOut, _ := cmdBranch.Output()
	branch := strings.TrimSpace(string(branchOut))
	if branch == "" {
		branch = "HEAD"
	}

	// 2. Get status porcelain with -uall (untracked files individually listed, respecting .gitignore)
	cmdStatus := exec.Command(gitutil.Executable(), "status", "--porcelain=v1", "-uall", "--ignored=no")
	cmdStatus.Dir = workspacePath
	statusOut, err := cmdStatus.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get git status: %w", err)
	}

	lines := strings.Split(string(statusOut), "\n")
	var changes []GitFileChange
	totalAdd := 0
	totalDel := 0

	for _, line := range lines {
		trimmed := strings.TrimRight(line, "\r")
		if len(trimmed) < 4 {
			continue
		}

		code := trimmed[0:2]
		filePath := strings.TrimSpace(trimmed[3:])
		// Strip surrounding quotes if git outputs quoted paths
		if strings.HasPrefix(filePath, "\"") && strings.HasSuffix(filePath, "\"") {
			filePath = filePath[1 : len(filePath)-1]
		}

		statusCode := "M"
		if strings.Contains(code, "A") {
			statusCode = "A"
		} else if strings.Contains(code, "D") {
			statusCode = "D"
		} else if strings.Contains(code, "?") {
			statusCode = "?"
		}

		// Calculate numstat lines for tracked/staged/modified files
		addCount := 0
		delCount := 0

		if statusCode != "?" {
			cmdDiff := exec.Command(gitutil.Executable(), "diff", "--numstat", "HEAD", "--", filePath)
			cmdDiff.Dir = workspacePath
			numstatOut, _ := cmdDiff.Output()

			numParts := strings.Fields(string(numstatOut))
			if len(numParts) >= 2 {
				// git diff --numstat outputs "-" for binary files
				if numParts[0] != "-" {
					fmt.Sscanf(numParts[0], "%d", &addCount)
				}
				if numParts[1] != "-" {
					fmt.Sscanf(numParts[1], "%d", &delCount)
				}
			}
		} else {
			// For untracked new files, count total lines in file as additions (only for text files)
			fullPath := filepath.Join(workspacePath, filePath)
			if data, err := os.ReadFile(fullPath); err == nil {
				// Check if binary file (contains null byte)
				if bytes.IndexByte(data, 0) == -1 {
					if len(data) > 0 {
						addCount = bytes.Count(data, []byte("\n"))
						if !bytes.HasSuffix(data, []byte("\n")) {
							addCount++
						}
					}
				}
			}
		}

		totalAdd += addCount
		totalDel += delCount

		changes = append(changes, GitFileChange{
			Path:        filePath,
			Status:      statusCode,
			AddedLines:  addCount,
			RemovedLines: delCount,
		})
	}

	return &GitStatusResult{
		Branch:        branch,
		IsClean:       len(changes) == 0,
		ChangedFiles:  changes,
		TotalAdditions: totalAdd,
		TotalDeletions: totalDel,
	}, nil
}

// GetFileDiff returns unified diff for a single file (respecting untracked new files and directories)
func GetFileDiff(workspacePath, filePath string) (string, error) {
	if workspacePath == "" || filePath == "" {
		return "", fmt.Errorf("invalid path parameters")
	}

	// 1. Try standard git diff HEAD -- <filePath>
	cmd := exec.Command(gitutil.Executable(), "diff", "HEAD", "--", filePath)
	cmd.Dir = workspacePath
	out, err := cmd.CombinedOutput()
	if err == nil && len(bytes.TrimSpace(out)) > 0 {
		return string(out), nil
	}

	// 2. Check working tree vs staged index
	cmdStaged := exec.Command(gitutil.Executable(), "diff", "--", filePath)
	cmdStaged.Dir = workspacePath
	stagedOut, _ := cmdStaged.CombinedOutput()
	if len(bytes.TrimSpace(stagedOut)) > 0 {
		return string(stagedOut), nil
	}

	// 3. For untracked files or directories, inspect on disk
	fullPath := filepath.Join(workspacePath, filePath)
	info, statErr := os.Stat(fullPath)
	if statErr == nil {
		if info.IsDir() {
			// If it's a directory (e.g. tests/__pycache__), diff each child file inside it
			var combined strings.Builder
			_ = filepath.Walk(fullPath, func(p string, fi os.FileInfo, walkErr error) error {
				if walkErr != nil || fi.IsDir() {
					return nil
				}
				rel, rErr := filepath.Rel(workspacePath, p)
				if rErr == nil {
					cmdChild := exec.Command(gitutil.Executable(), "diff", "--no-index", "/dev/null", rel)
					cmdChild.Dir = workspacePath
					childOut, _ := cmdChild.CombinedOutput()
					if len(childOut) > 0 {
						combined.WriteString(string(childOut))
						combined.WriteString("\n")
					}
				}
				return nil
			})
			if combined.Len() > 0 {
				return combined.String(), nil
			}
		} else {
			// Single untracked file diff against /dev/null
			cmdUntracked := exec.Command(gitutil.Executable(), "diff", "--no-index", "/dev/null", filePath)
			cmdUntracked.Dir = workspacePath
			untrackedOut, _ := cmdUntracked.CombinedOutput()
			if len(untrackedOut) > 0 {
				return string(untrackedOut), nil
			}
		}
	}

	return string(out), nil
}

// ExecuteCommit performs git commit
func ExecuteCommit(workspacePath, message string) error {
	if message == "" {
		return fmt.Errorf("commit message cannot be empty")
	}

	// Add changes
	addCmd := exec.Command(gitutil.Executable(), "add", "-A")
	addCmd.Dir = workspacePath
	if out, err := addCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git add failed: %s (%w)", string(out), err)
	}

	// Commit
	commitCmd := exec.Command(gitutil.Executable(), "commit", "-m", message)
	commitCmd.Dir = workspacePath
	if out, err := commitCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git commit failed: %s (%w)", string(out), err)
	}

	return nil
}

// ExecutePush performs git push
func ExecutePush(workspacePath string) (string, error) {
	cmd := exec.Command(gitutil.Executable(), "push")
	cmd.Dir = workspacePath
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return stderr.String(), fmt.Errorf("git push failed: %s (%w)", stderr.String(), err)
	}
	return stdout.String(), nil
}

// ExecutePull performs git pull
func ExecutePull(workspacePath string) (string, error) {
	cmd := exec.Command(gitutil.Executable(), "pull")
	cmd.Dir = workspacePath
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return stderr.String(), fmt.Errorf("git pull failed: %s (%w)", stderr.String(), err)
	}
	return stdout.String(), nil
}
