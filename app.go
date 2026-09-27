package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"aethergrok/pkg/gitutil"
	"aethergrok/pkg/grokrunner"
	"aethergrok/pkg/hotkey"
	"aethergrok/pkg/logger"
	"aethergrok/pkg/permissions"
	"aethergrok/pkg/screen"
	"aethergrok/pkg/skills"
	"aethergrok/pkg/terminal"
	"aethergrok/pkg/workspace"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct represents application runtime state
type App struct {
	ctx           context.Context
	runner        *grokrunner.Runner
	screenCapture *screen.Orchestrator
	skillsReg     *skills.Registry
	hotkeyMgr     *hotkey.Manager
	terminalMgr   *terminal.Manager
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		runner:        grokrunner.NewRunner(),
		screenCapture: screen.NewOrchestrator(),
		skillsReg:     skills.NewRegistry(),
		hotkeyMgr:     hotkey.NewManager(),
		terminalMgr:   terminal.NewManager(),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	if a.terminalMgr != nil {
		a.terminalMgr.SetContext(ctx)
	}
	if a.hotkeyMgr != nil {
		a.hotkeyMgr.SetHandler(func() {
			if a.ctx != nil {
				wailsRuntime.EventsEmit(a.ctx, "snapshot:trigger_global", map[string]interface{}{
					"timestamp": time.Now().UnixMilli(),
				})
			}
		})
	}
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
	if a.hotkeyMgr != nil {
		a.hotkeyMgr.Unregister()
	}
	if a.runner != nil {
		a.runner.CancelAll()
	}
	if a.terminalMgr != nil {
		// Clean up all running terminal instances and their process groups
		_ = a.terminalMgr.CloseSessionTerminals("")
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

// RegisterGlobalSnapshotShortcut registers or updates the global screen snapshot shortcut
func (a *App) RegisterGlobalSnapshotShortcut(shortcutStr string) error {
	if a.hotkeyMgr == nil {
		a.hotkeyMgr = hotkey.NewManager()
	}
	return a.hotkeyMgr.RegisterShortcut(shortcutStr)
}

// UnregisterGlobalSnapshotShortcut unregisters the global screen snapshot shortcut
func (a *App) UnregisterGlobalSnapshotShortcut() {
	if a.hotkeyMgr != nil {
		a.hotkeyMgr.Unregister()
	}
}

// GetSnapshotCacheStats calculates the total size and count of screenshot cache files
func (a *App) GetSnapshotCacheStats() (*screen.CacheStats, error) {
	return screen.GetSnapshotCacheStats()
}

// ClearSnapshotCache clears all grok screenshot cache files from disk
func (a *App) ClearSnapshotCache() (*screen.ClearCacheResult, error) {
	return screen.ClearSnapshotCache()
}

// SaveTemporaryImage caches a dropped image or screenshot to disk in the temp directory
func (a *App) SaveTemporaryImage(base64Data string, mimeType string) (*screen.SnapshotResult, error) {
	return screen.SaveTemporaryImage(base64Data, mimeType)
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

// SaveLogExport opens a native save file dialog to export system logs as .log or .json
func (a *App) SaveLogExport(defaultFilename, content, fileType string) (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("application context not initialized")
	}

	var filterName, pattern string
	if fileType == "json" {
		filterName = "JSON Log Files (*.json)"
		pattern = "*.json"
	} else {
		filterName = "Text Log Files (*.log)"
		pattern = "*.log"
	}

	filePath, err := wailsRuntime.SaveFileDialog(a.ctx, wailsRuntime.SaveDialogOptions{
		Title:           "Export System Logs",
		DefaultFilename: defaultFilename,
		Filters: []wailsRuntime.FileFilter{
			{
				DisplayName: filterName,
				Pattern:     pattern,
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
		return "", fmt.Errorf("failed to write export file: %w", err)
	}

	return filePath, nil
}

// AppendSystemLog writes a log entry asynchronously into the persistent disk log buffer
func (a *App) AppendSystemLog(entry logger.LogEntry) error {
	return logger.GetDiskLogger().Append(entry)
}

// LoadPersistedLogs retrieves up to `limit` entries from ~/.grok/logs/aethergrok.log
func (a *App) LoadPersistedLogs(limit int) ([]logger.LogEntry, error) {
	return logger.GetDiskLogger().ReadRecentEntries(limit)
}

// ClearPersistedLogs deletes all persistent log files on disk
func (a *App) ClearPersistedLogs() error {
	return logger.GetDiskLogger().Clear()
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

// CreateTerminal starts a new isolated pseudo-terminal instance
func (a *App) CreateTerminal(sessionID, termID, cwd, shell string) error {
	if a.terminalMgr == nil {
		return fmt.Errorf("terminal manager not initialized")
	}
	return a.terminalMgr.Create(sessionID, termID, cwd, shell)
}

// WriteTerminal writes user input to a pseudo-terminal
func (a *App) WriteTerminal(termID, data string) error {
	if a.terminalMgr == nil {
		return fmt.Errorf("terminal manager not initialized")
	}
	return a.terminalMgr.Write(termID, data)
}

// ResizeTerminal resizes the dimensions of a pseudo-terminal
func (a *App) ResizeTerminal(termID string, cols, rows int) error {
	if a.terminalMgr == nil {
		return fmt.Errorf("terminal manager not initialized")
	}
	return a.terminalMgr.Resize(termID, cols, rows)
}

// CloseTerminal terminates a pseudo-terminal and its full process tree
func (a *App) CloseTerminal(termID string) error {
	if a.terminalMgr == nil {
		return fmt.Errorf("terminal manager not initialized")
	}
	return a.terminalMgr.Close(termID)
}

// GetTerminalBuffer retrieves recent output history for a terminal instance
func (a *App) GetTerminalBuffer(termID string) string {
	if a.terminalMgr == nil {
		return ""
	}
	return a.terminalMgr.GetTerminalBuffer(termID)
}

// CloseSessionTerminals terminates all terminals belonging to a session
func (a *App) CloseSessionTerminals(sessionID string) error {
	if a.terminalMgr == nil {
		return fmt.Errorf("terminal manager not initialized")
	}
	return a.terminalMgr.CloseSessionTerminals(sessionID)
}

// GetPlanContent reads the full markdown content of a plan file from disk
func (a *App) GetPlanContent(planPath string) (string, error) {
	if strings.TrimSpace(planPath) == "" {
		return "", fmt.Errorf("plan path cannot be empty")
	}

	cleanPath := strings.TrimSpace(planPath)
	if strings.HasPrefix(cleanPath, "~") {
		if homeDir, err := os.UserHomeDir(); err == nil {
			cleanPath = filepath.Join(homeDir, cleanPath[1:])
		}
	}

	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
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
		cmd := exec.Command(gitutil.Executable(), args...)
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

// CheckMicrophonePermission inspects if microphone access is authorized
func (a *App) CheckMicrophonePermission() permissions.Status {
	return permissions.CheckMicrophonePermission()
}

// RequestMicrophonePermission triggers macOS system prompt if not yet determined
func (a *App) RequestMicrophonePermission() permissions.Status {
	return permissions.RequestMicrophonePermission()
}

// GetSystemAudioInputDevices returns the list of physical and virtual microphone input devices
func (a *App) GetSystemAudioInputDevices() ([]permissions.AudioDeviceInfo, error) {
	return permissions.GetAudioInputDevices()
}

// OpenMicrophoneSettings opens macOS System Settings to Microphone panel
func (a *App) OpenMicrophoneSettings() error {
	return permissions.OpenMicrophonePreferences()
}

// SaveVoiceAudioRecording writes recorded base64 audio data to a temporary file in ~/.grok/voice_cache
func (a *App) SaveVoiceAudioRecording(base64Data, ext string) (string, error) {
	if strings.TrimSpace(base64Data) == "" {
		return "", fmt.Errorf("base64 audio data is empty")
	}

	// Remove data URI prefix if present (e.g. data:audio/webm;base64,...)
	commaIdx := strings.Index(base64Data, ",")
	rawBase64 := base64Data
	if commaIdx != -1 {
		rawBase64 = base64Data[commaIdx+1:]
	}

	decoded, err := base64.StdEncoding.DecodeString(rawBase64)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 audio: %w", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home dir: %w", err)
	}

	cacheDir := filepath.Join(homeDir, ".grok", "voice_cache")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create voice cache dir: %w", err)
	}

	cleanExt := strings.TrimPrefix(ext, ".")
	if cleanExt == "" {
		cleanExt = "webm"
	}

	fileName := fmt.Sprintf("voice_dictation_%d.%s", time.Now().UnixNano(), cleanExt)
	targetPath := filepath.Join(cacheDir, fileName)

	if err := os.WriteFile(targetPath, decoded, 0600); err != nil {
		return "", fmt.Errorf("failed to write voice recording file: %w", err)
	}

	return targetPath, nil
}

// DeleteVoiceAudioRecording deletes a temporary voice recording file
func (a *App) DeleteVoiceAudioRecording(filePath string) error {
	if strings.TrimSpace(filePath) == "" {
		return nil
	}
	_ = os.Remove(filePath)
	return nil
}

// TranscribeAudioWithGrok runs high-speed audio transcription using native speech-to-text / multimodal
// audio models via the configured gateway (e.g. 9Router / OpenAI / Grok STT endpoint) with seamless fallback to Grok CLI in an isolated sandbox,
// and guarantees zero session pollution in the workspace.
func (a *App) TranscribeAudioWithGrok(workspacePath, audioFilePath string) (string, error) {
	if strings.TrimSpace(audioFilePath) == "" {
		return "", fmt.Errorf("audio file path is empty")
	}

	if _, statErr := os.Stat(audioFilePath); os.IsNotExist(statErr) {
		return "", fmt.Errorf("audio file does not exist at %s", audioFilePath)
	}

	// 1. First priority: Try high-speed direct audio transcription via 9Router / OpenAI STT endpoint (typically ~1-2 seconds)
	// Read ~/.grok/config.toml or environment to discover endpoint URL and auth token
	endpointURL := "http://127.0.0.1:20128"
	apiKey := "sk-81f5f3ce306056d3-wh99un-4249695c"

	if envURL := os.Getenv("NINEROUTER_URL"); envURL != "" {
		endpointURL = strings.TrimSuffix(envURL, "/")
	} else if envBase := os.Getenv("OPENAI_BASE_URL"); envBase != "" {
		endpointURL = strings.TrimSuffix(strings.TrimSuffix(envBase, "/v1"), "/")
	}

	if envKey := os.Getenv("NINEROUTER_KEY"); envKey != "" {
		apiKey = envKey
	} else if envKey := os.Getenv("JCODE_9ROUTER_API_KEY"); envKey != "" {
		apiKey = envKey
	} else if envKey := os.Getenv("ANTHROPIC_AUTH_TOKEN"); envKey != "" {
		apiKey = envKey
	}

	// Clean up any preamble/conversational prefixes if present
	cleanTranscript := func(text string) string {
		text = strings.TrimSpace(text)
		// Remove common conversational preamble lines generated by AI models
		preambles := []string{
			"Berikut adalah transkripsi yang akurat dari rekaman audio tersebut:",
			"Berikut adalah transkripsi dari rekaman audio tersebut:",
			"Berikut adalah transkripsi audio tersebut:",
			"Berikut transkripsi audio:",
			"Hasil transkripsi audio:",
			"Berikut adalah hasil transkripsi:",
			"Here is the transcription of the audio recording:",
			"Here is the accurate transcription of the audio:",
			"Here is the transcription:",
		}
		for _, p := range preambles {
			if strings.HasPrefix(strings.ToLower(text), strings.ToLower(p)) {
				text = strings.TrimSpace(text[len(p):])
			}
		}
		// Strip surrounding markdown quotes if returned in a quote block or triple backticks
		text = strings.Trim(text, "`\"'")
		text = strings.TrimSpace(text)
		return text
	}

	transcript, err := a.transcribeAudioViaSTTEndpoint(endpointURL, apiKey, audioFilePath)
	if err == nil && strings.TrimSpace(transcript) != "" {
		cleaned := cleanTranscript(transcript)
		if cleaned != "" {
			return cleaned, nil
		}
	}

	// 2. Second priority: Try multimodal audio chat completions via gateway (ag/gemini-3.8-flash / gemini-3.7-flash)
	transcript, errChat := a.transcribeAudioViaChatCompletions(endpointURL, apiKey, audioFilePath)
	if errChat == nil && strings.TrimSpace(transcript) != "" {
		cleaned := cleanTranscript(transcript)
		if cleaned != "" {
			return cleaned, nil
		}
	}

	// 3. Fallback: Run Grok CLI in an isolated temp directory to prevent workspace session leakage
	tmpDir, tmpErr := os.MkdirTemp("", "aethergrok_transcribe_*")
	if tmpErr == nil {
		defer os.RemoveAll(tmpDir)
	} else {
		tmpDir = os.TempDir()
	}

	grokBin := a.runner.GetBinaryPath()
	if grokBin == "" {
		grokBin = grokrunner.ResolveGrokBinary()
	}

	systemInstructions := "Kamu adalah transcriber audio programmer yang sangat akurat. Dengarkan rekaman audio teknis ini. Transkripsikan dengan jelas, gunakan istilah teknis, nama variabel, fungsi, bahasa pemrograman, dan tanda baca yang tepat dan rapi. PENTING: Keluarkan HANYA teks transkripsi murni. DILARANG KERAS menyertakan kalimat pembuka seperti 'Berikut adalah transkripsi...', kalimat penutup, tanda petik pembungkus, atau penjelasan tambahan."

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	promptText := fmt.Sprintf("%s\n\n[Attached Audio: %s]", systemInstructions, audioFilePath)
	cmd := exec.CommandContext(ctx, grokBin, "-p", promptText, "--no-subagents", "--disable-web-search")
	cmd.Dir = tmpDir
	cmd.Env = grokrunner.EnsureExecEnvironment()

	out, errExec := cmd.CombinedOutput()
	if errExec != nil {
		var detailErrs []string
		if err != nil {
			detailErrs = append(detailErrs, fmt.Sprintf("STT Error: %v", err))
		}
		if errChat != nil {
			detailErrs = append(detailErrs, fmt.Sprintf("Chat Audio Error: %v", errChat))
		}
		if ctx.Err() == context.DeadlineExceeded {
			detailErrs = append(detailErrs, "Grok CLI timed out after 20s")
		} else {
			detailErrs = append(detailErrs, fmt.Sprintf("Grok CLI Error: %v (output: %s)", errExec, strings.TrimSpace(string(out))))
		}
		return "", fmt.Errorf("transcription failed on all providers:\n%s", strings.Join(detailErrs, "\n"))
	}

	cliTranscript := strings.TrimSpace(string(out))
	cliTranscript = strings.TrimPrefix(cliTranscript, "Grok:")
	cliTranscript = cleanTranscript(cliTranscript)

	if cliTranscript == "" {
		return "", fmt.Errorf("transcription returned empty response")
	}

	return cliTranscript, nil
}

// transcribeAudioViaSTTEndpoint performs multipart speech-to-text POST to /v1/audio/transcriptions
func (a *App) transcribeAudioViaSTTEndpoint(baseURL, apiKey, audioFilePath string) (string, error) {
	audioFile, err := os.Open(audioFilePath)
	if err != nil {
		return "", err
	}
	defer audioFile.Close()

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add audio file part
	fileName := filepath.Base(audioFilePath)
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(part, audioFile); err != nil {
		return "", err
	}

	// Model candidate list: gemini-3.8-flash, whisper-1, groq/whisper-large-v3
	_ = writer.WriteField("model", "gemini/gemini-3.8-flash")
	_ = writer.WriteField("prompt", "Kamu adalah transcriber audio programmer yang sangat akurat. Dengarkan rekaman teknis ini dan transkripsikan dengan istilah teknis coding, variabel, dan tanda baca yang tepat.")
	_ = writer.WriteField("response_format", "json")

	if err := writer.Close(); err != nil {
		return "", err
	}

	targetURL := fmt.Sprintf("%s/v1/audio/transcriptions", strings.TrimSuffix(baseURL, "/"))
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", targetURL, &requestBody)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	client := &http.Client{Timeout: 12 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("STT API status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var resObj struct {
		Text  string `json:"text"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(bodyBytes, &resObj); err != nil {
		return strings.TrimSpace(string(bodyBytes)), nil
	}

	if resObj.Error != nil && resObj.Error.Message != "" {
		return "", fmt.Errorf("STT API returned error: %s", resObj.Error.Message)
	}

	return strings.TrimSpace(resObj.Text), nil
}

// transcribeAudioViaChatCompletions performs multimodal audio transcription via /v1/chat/completions
func (a *App) transcribeAudioViaChatCompletions(baseURL, apiKey, audioFilePath string) (string, error) {
	audioBytes, err := os.ReadFile(audioFilePath)
	if err != nil {
		return "", err
	}

	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(audioFilePath)), ".")
	if ext == "" {
		ext = "webm"
	}
	audioB64 := base64.StdEncoding.EncodeToString(audioBytes)

	promptText := "Kamu adalah transcriber audio programmer yang sangat akurat. Dengarkan rekaman audio teknis ini secara verbatim. Transkripsikan dengan jelas, gunakan istilah teknis, nama variabel, fungsi, bahasa pemrograman, dan tanda baca yang tepat. PENTING: Keluarkan HANYA teks transkripsi murni. Jangan menambahkan kalimat pembuka (seperti 'Berikut adalah transkripsi...', 'Here is the transcription...'), jangan menambahkan tanda petik di awal/akhir, dan jangan menambahkan penjelasan apa pun."

	// Fast transcription candidate models with 1000k context window:
	// ag/gemini-3.7-flash-low completes in ~3s, ag/gemini-3.8-flash-low in ~4s
	candidateModels := []string{"ag/gemini-3.7-flash-low", "ag/gemini-3.8-flash-low", "ag/gemini-3.8-flash"}

	var lastErr error
	for _, modelID := range candidateModels {
		payload := map[string]interface{}{
			"model":  modelID,
			"stream": false,
			"messages": []map[string]interface{}{
				{
					"role": "user",
					"content": []map[string]interface{}{
						{
							"type": "input_audio",
							"input_audio": map[string]string{
								"data":   audioB64,
								"format": ext,
							},
						},
						{
							"type": "text",
							"text": promptText,
						},
					},
				},
			},
		}

		jsonBytes, err := json.Marshal(payload)
		if err != nil {
			return "", err
		}

		targetURL := fmt.Sprintf("%s/v1/chat/completions", strings.TrimSuffix(baseURL, "/"))
		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)

		req, err := http.NewRequestWithContext(ctx, "POST", targetURL, bytes.NewBuffer(jsonBytes))
		if err != nil {
			cancel()
			lastErr = err
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		if apiKey != "" {
			req.Header.Set("Authorization", "Bearer "+apiKey)
		}

		client := &http.Client{Timeout: 25 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			cancel()
			lastErr = err
			continue
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		cancel()

		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			lastErr = fmt.Errorf("Chat audio API status %d: %s", resp.StatusCode, string(bodyBytes))
			continue
		}

		var resObj struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
			Error *struct {
				Message string `json:"message"`
			} `json:"error"`
		}

		if err := json.Unmarshal(bodyBytes, &resObj); err == nil && len(resObj.Choices) > 0 {
			textResult := strings.TrimSpace(resObj.Choices[0].Message.Content)
			if textResult != "" {
				return textResult, nil
			}
		}
	}

	if lastErr != nil {
		return "", lastErr
	}

	return "", fmt.Errorf("no response choices returned from model")
}

// PullWorkspaceChanges pulls upstream commits
func (a *App) PullWorkspaceChanges(workspacePath string) (string, error) {
	return workspace.ExecutePull(workspacePath)
}
