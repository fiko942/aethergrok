package workspace

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
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

// ReadFileContent reads file content as text, capped at 1MB to prevent memory explosion
func ReadFileContent(workspacePath, relativePath string) (string, error) {
	fullPath := filepath.Join(workspacePath, relativePath)
	info, err := os.Stat(fullPath)
	if err != nil {
		return "", err
	}

	if info.Size() > 1024*1024 {
		return "", fmt.Errorf("file size (%d bytes) exceeds 1MB preview limit", info.Size())
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return "", err
	}

	return string(data), nil
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
	cmdBranch := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmdBranch.Dir = workspacePath
	branchOut, _ := cmdBranch.Output()
	branch := strings.TrimSpace(string(branchOut))
	if branch == "" {
		branch = "HEAD"
	}

	// 2. Get status porcelain with -uall (untracked files individually listed, respecting .gitignore)
	cmdStatus := exec.Command("git", "status", "--porcelain=v1", "-uall", "--ignored=no")
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
			cmdDiff := exec.Command("git", "diff", "--numstat", "HEAD", "--", filePath)
			cmdDiff.Dir = workspacePath
			numstatOut, _ := cmdDiff.Output()

			numParts := strings.Fields(string(numstatOut))
			if len(numParts) >= 2 {
				fmt.Sscanf(numParts[0], "%d", &addCount)
				fmt.Sscanf(numParts[1], "%d", &delCount)
			}
		} else {
			// For untracked new files, count total lines in file as additions
			fullPath := filepath.Join(workspacePath, filePath)
			if data, err := os.ReadFile(fullPath); err == nil {
				if len(data) > 0 {
					addCount = bytes.Count(data, []byte("\n"))
					if !bytes.HasSuffix(data, []byte("\n")) {
						addCount++
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

// GetFileDiff returns unified diff for a single file (respecting untracked new files and .gitignore)
func GetFileDiff(workspacePath, filePath string) (string, error) {
	// First try git diff HEAD -- <filePath>
	cmd := exec.Command("git", "diff", "HEAD", "--", filePath)
	cmd.Dir = workspacePath
	out, err := cmd.CombinedOutput()
	if err == nil && len(out) > 0 {
		return string(out), nil
	}

	// Also check working tree vs index (staged)
	cmdStaged := exec.Command("git", "diff", "--", filePath)
	cmdStaged.Dir = workspacePath
	stagedOut, _ := cmdStaged.CombinedOutput()
	if len(stagedOut) > 0 {
		return string(stagedOut), nil
	}

	// For untracked new files, produce unified diff against /dev/null
	cmdUntracked := exec.Command("git", "diff", "--no-index", "/dev/null", filePath)
	cmdUntracked.Dir = workspacePath
	untrackedOut, _ := cmdUntracked.CombinedOutput()
	if len(untrackedOut) > 0 {
		return string(untrackedOut), nil
	}

	return string(out), err
}

// ExecuteCommit performs git commit
func ExecuteCommit(workspacePath, message string) error {
	if message == "" {
		return fmt.Errorf("commit message cannot be empty")
	}

	// Add changes
	addCmd := exec.Command("git", "add", "-A")
	addCmd.Dir = workspacePath
	if out, err := addCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git add failed: %s (%w)", string(out), err)
	}

	// Commit
	commitCmd := exec.Command("git", "commit", "-m", message)
	commitCmd.Dir = workspacePath
	if out, err := commitCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git commit failed: %s (%w)", string(out), err)
	}

	return nil
}

// ExecutePush performs git push
func ExecutePush(workspacePath string) (string, error) {
	cmd := exec.Command("git", "push")
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
	cmd := exec.Command("git", "pull")
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
