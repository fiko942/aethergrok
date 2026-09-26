package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"aethergrok/pkg/grokrunner"
	"aethergrok/pkg/permissions"
	"aethergrok/pkg/screen"
	"aethergrok/pkg/skills"
	"aethergrok/pkg/workspace"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct represents application runtime state
type App struct {
	ctx           context.Context
	runner        *grokrunner.Runner
	screenCapture *screen.Orchestrator
	skillsReg     *skills.Registry
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		runner:        grokrunner.NewRunner(),
		screenCapture: screen.NewOrchestrator(),
		skillsReg:     skills.NewRegistry(),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// domReady is called after front-end resources have loaded
func (a *App) domReady(ctx context.Context) {
}

// beforeClose is called when the application is about to quit,
// either by clicking the window close button or calling runtime.Quit.
// Returning true will cause the application to continue, false will continue shutdown.
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	if a.runner != nil {
		a.runner.CancelAll()
	}
	return false
}

// shutdown is called at application termination
func (a *App) shutdown(ctx context.Context) {
	if a.runner != nil {
		a.runner.CancelAll()
	}
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, welcome to AetherGrok!", name)
}

// RunPromptStream starts a prompt turn and streams events to the Wails runtime
func (a *App) RunPromptStream(req grokrunner.PromptRequest) error {
	callbacks := grokrunner.StreamCallbacks{
		OnDelta: func(event grokrunner.DeltaEvent) {
			if a.ctx != nil {
				wailsRuntime.EventsEmit(a.ctx, string(grokrunner.EventDelta), event)
			}
		},
		OnToolCall: func(event grokrunner.ToolCallEvent) {
			if a.ctx != nil {
				wailsRuntime.EventsEmit(a.ctx, string(grokrunner.EventToolCall), event)
			}
		},
		OnPermissionRequest: func(event grokrunner.PermissionRequestEvent) {
			if a.ctx != nil {
				wailsRuntime.EventsEmit(a.ctx, string(grokrunner.EventPermissionRequest), event)
			}
		},
		OnComplete: func(event grokrunner.TurnCompleteEvent) {
			if a.ctx != nil {
				wailsRuntime.EventsEmit(a.ctx, string(grokrunner.EventTurnComplete), event)
			}
		},
		OnError: func(err error) {
			if a.ctx != nil {
				wailsRuntime.EventsEmit(a.ctx, string(grokrunner.EventError), map[string]string{
					"sessionId": req.SessionID,
					"error":     err.Error(),
				})
			}
		},
	}

	return a.runner.StartSession(a.ctx, req, callbacks)
}

// RespondPermission passes the user's permission choice to the running Grok process
func (a *App) RespondPermission(resp grokrunner.PermissionResponse) error {
	return a.runner.RespondPermission(resp)
}

// CancelSession cancels the active session turn
func (a *App) CancelSession(sessionID string) error {
	return a.runner.Cancel(sessionID)
}

// SetGrokBinaryPath updates the binary path used to spawn Grok CLI
func (a *App) SetGrokBinaryPath(path string) {
	a.runner.SetBinaryPath(path)
}

type wailsWindowController struct {
	ctx context.Context
}

func (w *wailsWindowController) Hide() {
	if w.ctx != nil {
		wailsRuntime.WindowHide(w.ctx)
	}
}

func (w *wailsWindowController) Show() {
	if w.ctx != nil {
		wailsRuntime.WindowShow(w.ctx)
	}
}

// CaptureScreenExcludingSelf coordinates non-intrusive snapshot capture with auto window hiding
func (a *App) CaptureScreenExcludingSelf(delayMs int) (*screen.SnapshotResult, error) {
	var winCtrl screen.WindowController
	if a.ctx != nil {
		winCtrl = &wailsWindowController{ctx: a.ctx}
	}
	return a.screenCapture.CaptureScreenExcludingWindow(context.Background(), winCtrl, delayMs)
}

// GetSnapshotCacheStats calculates the total size and count of screenshot cache files
func (a *App) GetSnapshotCacheStats() (*screen.CacheStats, error) {
	return screen.GetSnapshotCacheStats()
}

// ClearSnapshotCache clears all grok screenshot cache files from disk
func (a *App) ClearSnapshotCache() (*screen.ClearCacheResult, error) {
	return screen.ClearSnapshotCache()
}

// DeleteSessionTempFiles removes temporary files associated with a deleted session
func (a *App) DeleteSessionTempFiles(filePaths []string) error {
	return screen.DeleteSessionTempFiles(filePaths)
}

// GetInstalledSkills returns all discovered skills from ~/.grok/skills/ and ~/.agents/skills/
func (a *App) GetInstalledSkills() []skills.Skill {
	if a.skillsReg == nil {
		a.skillsReg = skills.NewRegistry()
	}
	return a.skillsReg.GetAll()
}

// SelectWorkspaceDirectory opens a native directory picker dialog
func (a *App) SelectWorkspaceDirectory() (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("application context not initialized")
	}
	dir, err := wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Select Workspace Folder",
	})
	if err != nil {
		return "", err
	}
	return dir, nil
}

// CheckDirectoryExists checks if a directory exists and is accessible on disk
func (a *App) CheckDirectoryExists(dirPath string) bool {
	if strings.TrimSpace(dirPath) == "" {
		return false
	}
	info, err := os.Stat(dirPath)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// SaveMarkdownExport opens a native save file dialog and writes markdown content to the selected path
func (a *App) SaveMarkdownExport(defaultFilename string, content string) (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("application context not initialized")
	}
	filePath, err := wailsRuntime.SaveFileDialog(a.ctx, wailsRuntime.SaveDialogOptions{
		Title:           "Export Session as Markdown",
		DefaultFilename: defaultFilename,
		Filters: []wailsRuntime.FileFilter{
			{
				DisplayName: "Markdown Files (*.md)",
				Pattern:     "*.md",
			},
			{
				DisplayName: "Text Files (*.txt)",
				Pattern:     "*.txt",
			},
			{
				DisplayName: "All Files (*.*)",
				Pattern:     "*.*",
			},
		},
	})
	if err != nil {
		return "", err
	}
	if filePath == "" {
		return "", nil // User cancelled
	}

	err = os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return filePath, nil
}

// DiscoverGrokSessions scans disk for sessions belonging to workspacePath
func (a *App) DiscoverGrokSessions(workspacePath string) ([]grokrunner.GrokSessionMetadata, error) {
	return grokrunner.DiscoverGrokSessions(workspacePath)
}

// LoadGrokSessionHistory reads and parses chat history messages for a specific session from disk
func (a *App) LoadGrokSessionHistory(workspacePath, sessionID string) ([]grokrunner.DiscoveredChatMessage, error) {
	return grokrunner.LoadGrokSessionMessages(workspacePath, sessionID)
}

// DeleteGrokSession removes a session folder from ~/.grok/sessions/
func (a *App) DeleteGrokSession(workspacePath, sessionID string) error {
	return grokrunner.DeleteGrokSessionDirectory(workspacePath, sessionID)
}

// GetSessionUsage returns the token consumption statistics for a session
func (a *App) GetSessionUsage(workspacePath, sessionID string) (*grokrunner.SessionUsageStats, error) {
	return grokrunner.GetSessionUsage(workspacePath, sessionID)
}

// CompactSession executes context compaction on a session
func (a *App) CompactSession(workspacePath, sessionID string) (*grokrunner.SessionUsageStats, error) {
	return grokrunner.CompactSession(a.ctx, a.runner.GetBinaryPath(), workspacePath, sessionID)
}

// OpenPathInSystem opens a folder or file in macOS Finder or Windows Explorer
func (a *App) OpenPathInSystem(targetPath string) error {
	if strings.TrimSpace(targetPath) == "" {
		return fmt.Errorf("target path cannot be empty")
	}

	cleanPath := strings.TrimSpace(targetPath)
	if strings.HasPrefix(cleanPath, "~") {
		if homeDir, err := os.UserHomeDir(); err == nil {
			cleanPath = filepath.Join(homeDir, cleanPath[1:])
		}
	}

	info, err := os.Stat(cleanPath)
	if err != nil {
		return err
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		if info.IsDir() {
			cmd = exec.Command("open", cleanPath)
		} else {
			cmd = exec.Command("open", "-R", cleanPath)
		}
	case "windows":
		if info.IsDir() {
			cmd = exec.Command("explorer.exe", filepath.Clean(cleanPath))
		} else {
			cmd = exec.Command("explorer.exe", fmt.Sprintf("/select,%s", filepath.Clean(cleanPath)))
		}
	default:
		cmd = exec.Command("xdg-open", cleanPath)
	}

	return cmd.Start()
}

// RevealGrokConfigFile reveals the user's ~/.grok/config.toml file in macOS Finder or Windows Explorer
func (a *App) RevealGrokConfigFile() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".grok", "config.toml")
	// If config.toml does not exist, check if ~/.grok directory exists
	grokDir := filepath.Join(homeDir, ".grok")
	targetPath := configPath
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if _, err := os.Stat(grokDir); err == nil {
			targetPath = grokDir
		} else {
			return fmt.Errorf("grok configuration file not found at %s", configPath)
		}
	}

	// Reveal in Finder on macOS or Explorer on Windows
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		if targetPath == configPath {
			cmd = exec.Command("open", "-R", targetPath)
		} else {
			cmd = exec.Command("open", targetPath)
		}
	case "windows":
		if targetPath == configPath {
			cmd = exec.Command("explorer.exe", fmt.Sprintf("/select,%s", filepath.Clean(targetPath)))
		} else {
			cmd = exec.Command("explorer.exe", filepath.Clean(targetPath))
		}
	default:
		// Linux fallback (xdg-open parent directory)
		cmd = exec.Command("xdg-open", filepath.Dir(targetPath))
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to open file manager: %w", err)
	}

	return nil
}

func (a *App) SearchSkills(query string, category string) []skills.Skill {
	if a.skillsReg == nil {
		a.skillsReg = skills.NewRegistry()
	}
	return a.skillsReg.Search(query, category)
}

// OpenExternalURL opens a web address in the user's default operating system browser
func (a *App) OpenExternalURL(targetURL string) error {
	trimmed := strings.TrimSpace(targetURL)
	if trimmed == "" {
		return fmt.Errorf("URL is empty")
	}

	if a.ctx != nil {
		wailsRuntime.BrowserOpenURL(a.ctx, trimmed)
		return nil
	}

	return fmt.Errorf("application context not initialized")
}

// GetAvailableModels returns the list of available models from Grok CLI
func (a *App) GetAvailableModels() []grokrunner.ModelInfo {
	if a.runner == nil {
		a.runner = grokrunner.NewRunner()
	}
	return a.runner.DiscoverAvailableModels()
}

// ScanGitHubSkills clones and inspects a GitHub repository for skills and dependencies
func (a *App) ScanGitHubSkills(repoURL string) (*skills.SkillAnalysisResult, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return skills.ScanGitHubRepo(ctx, repoURL)
}

// InstallDiscoveredSkills copies selected skills into user's ~/.grok/skills or ~/.agents/skills
func (a *App) InstallDiscoveredSkills(payload skills.SkillInstallPayload) (*skills.SkillInstallResult, error) {
	res, err := skills.InstallDiscoveredSkills(payload)
	if err == nil && a.skillsReg != nil {
		// Rescan registry to immediately reflect newly installed skills
		a.skillsReg.ScanSkills()
	}
	return res, err
}

// CleanupSkillImportTemp removes cloned temporary files
func (a *App) CleanupSkillImportTemp(tempPath string) error {
	return skills.CleanupTempSkills(tempPath)
}

// RevertWorkspaceFiles checks out or reverts file changes in git repository if applicable
func (a *App) RevertWorkspaceFiles(workspacePath string, filePaths []string) error {
	if workspacePath == "" || len(filePaths) == 0 {
		return nil
	}

	// Check if workspace is a git repo
	gitDir := filepath.Join(workspacePath, ".git")
	if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
		// Run git checkout -- <files>
		args := append([]string{"checkout", "--"}, filePaths...)
		cmd := exec.Command("git", args...)
		cmd.Dir = workspacePath
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git revert failed: %w (output: %s)", err, string(out))
		}
	}
	return nil
}

// ExecuteSkillSetupCommand executes a command line in temp directory and streams logs via Wails event
func (a *App) ExecuteSkillSetupCommand(workDir, commandLine string) error {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return skills.ExecuteSetupCommand(ctx, workDir, commandLine, func(line string) {
		if a.ctx != nil {
			wailsRuntime.EventsEmit(a.ctx, "skill:install_log", line)
		}
	})
}

// CheckAndRequestAccessibilityPermissions checks if the app has macOS Accessibility / Input Monitoring permission
// and prompts the user or opens System Settings if needed.
func (a *App) CheckAndRequestAccessibilityPermissions() permissions.Status {
	return permissions.CheckAndRequestAccessibility()
}

// OpenAccessibilitySettings opens macOS System Settings to Accessibility panel
func (a *App) OpenAccessibilitySettings() error {
	return permissions.OpenAccessibilityPreferences()
}

// ReadWorkspaceDirectory reads directory contents for the right sidebar file tree
func (a *App) ReadWorkspaceDirectory(workspacePath, relativeDir string) ([]workspace.FileItem, error) {
	return workspace.ReadDirectory(workspacePath, relativeDir)
}

// ReadWorkspaceFileContent reads a file's content with safety checks and confirmation flags
func (a *App) ReadWorkspaceFileContent(workspacePath, relativePath string, allowLarge bool) (string, error) {
	return workspace.ReadFileContent(workspacePath, relativePath, allowLarge)
}

// CheckFileExists checks if a file exists on disk (workspace-relative, home-relative, or absolute)
func (a *App) CheckFileExists(workspacePath, candidatePath string) *workspace.FileCheckResult {
	return workspace.CheckFileExists(workspacePath, candidatePath)
}

// CheckMultipleFilesExists batches file existence checks
func (a *App) CheckMultipleFilesExists(workspacePath string, candidates []string) map[string]workspace.FileCheckResult {
	return workspace.CheckMultipleFilesExists(workspacePath, candidates)
}

// GetWorkspaceGitStatus returns the branch, modified files, and diff stat
func (a *App) GetWorkspaceGitStatus(workspacePath string) (*workspace.GitStatusResult, error) {
	return workspace.GetGitStatus(workspacePath)
}

// GetWorkspaceFileDiff returns unified diff for a single modified file
func (a *App) GetWorkspaceFileDiff(workspacePath, filePath string) (string, error) {
	return workspace.GetFileDiff(workspacePath, filePath)
}

// CommitWorkspaceChanges commits staged/modified changes
func (a *App) CommitWorkspaceChanges(workspacePath, message string) error {
	return workspace.ExecuteCommit(workspacePath, message)
}

// PushWorkspaceChanges pushes commits to upstream
func (a *App) PushWorkspaceChanges(workspacePath string) (string, error) {
	return workspace.ExecutePush(workspacePath)
}

// PullWorkspaceChanges pulls upstream commits
func (a *App) PullWorkspaceChanges(workspacePath string) (string, error) {
	return workspace.ExecutePull(workspacePath)
}
