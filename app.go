package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
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
	"aethergrok/pkg/storage"
	"aethergrok/pkg/system"
	"aethergrok/pkg/terminal"
	"aethergrok/pkg/updater"
	"aethergrok/pkg/workspace"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct represents application runtime state
type App struct {
	ctx                context.Context
	runner             *grokrunner.Runner
	screenCapture      *screen.Orchestrator
	skillsReg          *skills.Registry
	hotkeyMgr          *hotkey.Manager
	dictationHotkeyMgr *hotkey.Manager
	terminalMgr        *terminal.Manager
	storageMgr         *storage.StorageManager
}

// NewApp creates a new App application struct
func NewApp() *App {
	sm, _ := storage.NewStorageManager("")
	runner := grokrunner.NewRunner()
	if sm != nil {
		runner.SetStorageManager(sm)
	}
	return &App{
		runner:             runner,
		screenCapture:      screen.NewOrchestrator(),
		skillsReg:          skills.NewRegistry(),
		hotkeyMgr:          hotkey.NewManager(),
		dictationHotkeyMgr: hotkey.NewManager(),
		terminalMgr:        terminal.NewManager(),
		storageMgr:         sm,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	_ = system.InitAppProcessGroup()
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
	if a.dictationHotkeyMgr != nil {
		a.dictationHotkeyMgr.SetKeyHandler(func(action string) {
			if a.ctx != nil {
				wailsRuntime.EventsEmit(a.ctx, "dictation:trigger_global", map[string]interface{}{
					"type":      action,
					"timestamp": time.Now().UnixMilli(),
				})
			}
		})
	}

	// Automatically register saved shortcuts on startup from persistent storage
	if a.storageMgr != nil {
		st, err := a.storageMgr.GetSettings()
		if err == nil {
			if a.hotkeyMgr != nil && st.SnapshotShortcut != "" {
				_ = a.hotkeyMgr.RegisterShortcut(st.SnapshotShortcut)
			}
			if a.dictationHotkeyMgr != nil && st.DictationShortcut != "" {
				_ = a.dictationHotkeyMgr.RegisterShortcut(st.DictationShortcut)
			}
		}
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
	if a.terminalMgr != nil {
		a.terminalMgr.CloseAll()
	}
	system.CleanupOrphanProcesses()
	return false
}

// shutdown is called at application termination
func (a *App) shutdown(ctx context.Context) {
	if a.hotkeyMgr != nil {
		a.hotkeyMgr.Unregister()
	}
	if a.dictationHotkeyMgr != nil {
		a.dictationHotkeyMgr.Unregister()
	}
	if a.runner != nil {
		a.runner.CancelAll()
	}
	if a.terminalMgr != nil {
		// Clean up all running terminal instances and their process groups
		a.terminalMgr.CloseAll()
	}
	system.CleanupOrphanProcesses()
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

// GetGrokBinaryPath returns the current active Grok CLI binary path
func (a *App) GetGrokBinaryPath() string {
	if a.runner != nil {
		return a.runner.GetBinaryPath()
	}
	return grokrunner.ResolveGrokBinary()
}

// AutoDetectGrokBinaryPath searches for and returns the native Grok CLI binary location on the host
func (a *App) AutoDetectGrokBinaryPath() string {
	return grokrunner.ResolveGrokBinary()
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
		wailsRuntime.WindowUnminimise(w.ctx)
	}
}

// CaptureScreenExcludingSelf coordinates non-intrusive snapshot capture with auto window hiding
func (a *App) CaptureScreenExcludingSelf(delayMs int) (*screen.SnapshotResult, error) {
	var winCtrl screen.WindowController
	autoHide := true
	if a.storageMgr != nil {
		if st, err := a.storageMgr.GetSettings(); err == nil {
			autoHide = st.SnapshotAutoHideWindow
		}
	}
	if a.ctx != nil && autoHide {
		winCtrl = &wailsWindowController{ctx: a.ctx}
	}
	return a.screenCapture.CaptureScreenExcludingWindow(context.Background(), winCtrl, delayMs)
}

// RegisterGlobalSnapshotShortcut registers or updates the global screen snapshot shortcut
func (a *App) RegisterGlobalSnapshotShortcut(shortcutStr string) error {
	if a.hotkeyMgr == nil {
		a.hotkeyMgr = hotkey.NewManager()
	}
	a.hotkeyMgr.SetHandler(func() {
		if a.ctx != nil {
			wailsRuntime.EventsEmit(a.ctx, "snapshot:trigger_global", map[string]interface{}{
				"timestamp": time.Now().UnixMilli(),
			})
		}
	})
	return a.hotkeyMgr.RegisterShortcut(shortcutStr)
}

// UnregisterGlobalSnapshotShortcut unregisters the global screen snapshot shortcut
func (a *App) UnregisterGlobalSnapshotShortcut() {
	if a.hotkeyMgr != nil {
		a.hotkeyMgr.Unregister()
	}
}

// RegisterGlobalDictationShortcut registers or updates the global dictation shortcut
func (a *App) RegisterGlobalDictationShortcut(shortcutStr string) error {
	if a.dictationHotkeyMgr == nil {
		a.dictationHotkeyMgr = hotkey.NewManager()
	}
	a.dictationHotkeyMgr.SetKeyHandler(func(action string) {
		if a.ctx != nil {
			wailsRuntime.EventsEmit(a.ctx, "dictation:trigger_global", map[string]interface{}{
				"type":      action,
				"timestamp": time.Now().UnixMilli(),
			})
		}
	})
	return a.dictationHotkeyMgr.RegisterShortcut(shortcutStr)
}

// UnregisterGlobalDictationShortcut unregisters the global dictation shortcut
func (a *App) UnregisterGlobalDictationShortcut() {
	if a.dictationHotkeyMgr != nil {
		a.dictationHotkeyMgr.Unregister()
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

// CheckGrokInstallation returns the detection status of the Grok CLI
func (a *App) CheckGrokInstallation() grokrunner.GrokInstallStatus {
	return grokrunner.DetectGrokInstallation()
}

// InstallGrokCLI triggers the automated installer and emits progress events to the frontend
func (a *App) InstallGrokCLI() (*grokrunner.GrokInstallStatus, error) {
	return grokrunner.InstallGrokCLI(a.ctx, func(progress grokrunner.GrokInstallProgress) {
		if a.ctx != nil {
			wailsRuntime.EventsEmit(a.ctx, "grok:install_progress", progress)
		}
	})
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

// EnsureTerminal ensures a pseudo-terminal instance is alive without killing existing running sessions
func (a *App) EnsureTerminal(sessionID, termID, cwd, shell string) error {
	if a.terminalMgr == nil {
		return fmt.Errorf("terminal manager not initialized")
	}
	return a.terminalMgr.EnsureCreated(sessionID, termID, cwd, shell)
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

// InterruptTerminal sends a Ctrl+C interrupt and terminates running child processes in a terminal
func (a *App) InterruptTerminal(termID string) error {
	if a.terminalMgr == nil {
		return fmt.Errorf("terminal manager not initialized")
	}
	return a.terminalMgr.Interrupt(termID)
}

// KillTerminal terminates a pseudo-terminal instance and all child processes immediately
func (a *App) KillTerminal(termID string) error {
	if a.terminalMgr == nil {
		return fmt.Errorf("terminal manager not initialized")
	}
	return a.terminalMgr.Kill(termID)
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

// Storage Management APIs (Persistent Settings, Workspaces, UIState)

// GetAppSettings retrieves the stored application settings
func (a *App) GetAppSettings() (storage.AppSettings, error) {
	if a.storageMgr == nil {
		sm, err := storage.NewStorageManager("")
		if err != nil {
			return storage.DefaultSettings(), err
		}
		a.storageMgr = sm
	}
	settings, err := a.storageMgr.GetSettings()
	if err != nil {
		return settings, err
	}
	// Fallback to auto-detected binary path if none configured or invalid on host
	if settings.GrokBinaryPath == "" {
		settings.GrokBinaryPath = grokrunner.ResolveGrokBinary()
	} else if _, statErr := os.Stat(settings.GrokBinaryPath); statErr != nil {
		settings.GrokBinaryPath = grokrunner.ResolveGrokBinary()
	}
	return settings, nil
}

// SaveAppSettings writes updated application settings to disk
func (a *App) SaveAppSettings(settings storage.AppSettings) error {
	if a.storageMgr == nil {
		sm, err := storage.NewStorageManager("")
		if err != nil {
			return err
		}
		a.storageMgr = sm
	}
	err := a.storageMgr.SaveSettings(settings)
	if err != nil {
		return err
	}

	// Dynamically sync global OS hotkeys whenever settings are saved
	if a.hotkeyMgr != nil && settings.SnapshotShortcut != "" {
		_ = a.hotkeyMgr.RegisterShortcut(settings.SnapshotShortcut)
	}
	if a.dictationHotkeyMgr != nil && settings.DictationShortcut != "" {
		_ = a.dictationHotkeyMgr.RegisterShortcut(settings.DictationShortcut)
	}
	return nil
}

// GetWorkspaces retrieves all registered workspace folders
func (a *App) GetWorkspaces() ([]storage.Workspace, error) {
	if a.storageMgr == nil {
		sm, err := storage.NewStorageManager("")
		if err != nil {
			return []storage.Workspace{}, err
		}
		a.storageMgr = sm
	}
	return a.storageMgr.GetWorkspaces()
}

// SaveWorkspaces persists the list of workspaces
func (a *App) SaveWorkspaces(workspaces []storage.Workspace) error {
	if a.storageMgr == nil {
		sm, err := storage.NewStorageManager("")
		if err != nil {
			return err
		}
		a.storageMgr = sm
	}
	return a.storageMgr.SaveWorkspaces(workspaces)
}

// GetUIState retrieves open tabs and active workspace/session
func (a *App) GetUIState() (storage.UIState, error) {
	if a.storageMgr == nil {
		sm, err := storage.NewStorageManager("")
		if err != nil {
			return storage.UIState{}, err
		}
		a.storageMgr = sm
	}
	return a.storageMgr.GetUIState()
}

// SaveUIState persists open tabs and active workspace/session
func (a *App) SaveUIState(state storage.UIState) error {
	if a.storageMgr == nil {
		sm, err := storage.NewStorageManager("")
		if err != nil {
			return err
		}
		a.storageMgr = sm
	}
	return a.storageMgr.SaveUIState(state)
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

// CheckScreenCapturePermission inspects if screen capture / recording access is authorized
func (a *App) CheckScreenCapturePermission() permissions.Status {
	return permissions.CheckScreenCapturePermission()
}

// RequestScreenCapturePermission triggers macOS system prompt or preflight for screen capture
func (a *App) RequestScreenCapturePermission() permissions.Status {
	return permissions.RequestScreenCapturePermission()
}

// OpenScreenCaptureSettings opens macOS System Settings to Screen Recording panel
func (a *App) OpenScreenCaptureSettings() error {
	return permissions.OpenScreenCapturePreferences()
}

// CheckAllSystemPermissions inspects and aggregates the status of all required system permissions
func (a *App) CheckAllSystemPermissions() permissions.AllPermissionsStatus {
	return permissions.CheckAllSystemPermissions()
}

// GetSystemAudioInputDevices returns the list of physical and virtual microphone input devices
func (a *App) GetSystemAudioInputDevices() ([]permissions.AudioDeviceInfo, error) {
	return permissions.GetAudioInputDevices()
}

// OpenMicrophoneSettings opens macOS System Settings to Microphone panel
func (a *App) OpenMicrophoneSettings() error {
	return permissions.OpenMicrophonePreferences()
}

// VolumeMuteResult stores the system audio state prior to muting
type VolumeMuteResult struct {
	OriginalVolume int  `json:"originalVolume"`
	WasMuted       bool `json:"wasMuted"`
}

// MuteSystemVolume mutes the system audio master output and returns the previous volume state
func (a *App) MuteSystemVolume() (*VolumeMuteResult, error) {
	state, err := system.MuteSystemVolume()
	if err != nil {
		return nil, err
	}
	return &VolumeMuteResult{
		OriginalVolume: state.OriginalVolume,
		WasMuted:       state.WasMuted,
	}, nil
}

// RestoreSystemVolume restores the system audio volume to its previous state
func (a *App) RestoreSystemVolume(prevVolume int, wasMuted bool) error {
	return system.RestoreSystemVolume(system.SystemVolumeState{
		OriginalVolume: prevVolume,
		WasMuted:       wasMuted,
	})
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

func generateUUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return fmt.Sprintf("%08x-%04x-4%03x-%04x-%012x",
			time.Now().UnixNano()&0xffffffff,
			time.Now().Unix()&0xffff,
			(time.Now().UnixNano()>>16)&0x0fff,
			(time.Now().UnixNano()>>32)&0x3fff|0x8000,
			time.Now().UnixNano()&0xffffffffffff,
		)
	}
	b[6] = (b[6] & 0x0f) | 0x40 // Version 4
	b[8] = (b[8] & 0x3f) | 0x80 // Variant RFC4122
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// TranscribeAudioWithGrok runs audio transcription directly via Grok CLI in a dedicated workspace session,
// and guarantees absolute cleanup of the temporary audio file and temporary Grok session directory.
func (a *App) TranscribeAudioWithGrok(workspacePath, audioFilePath string) (string, error) {
	if strings.TrimSpace(audioFilePath) == "" {
		return "", fmt.Errorf("audio file path is empty")
	}

	if _, statErr := os.Stat(audioFilePath); os.IsNotExist(statErr) {
		return "", fmt.Errorf("audio file does not exist at %s", audioFilePath)
	}

	if strings.TrimSpace(workspacePath) == "" {
		if ws, err := os.Getwd(); err == nil {
			workspacePath = ws
		}
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

	sessionUUID := generateUUID()

	// Guaranteed cleanup of temporary Grok session directory and scratch audio file on completion
	defer func() {
		_ = grokrunner.DeleteGrokSessionDirectory(workspacePath, sessionUUID)
		_ = os.Remove(audioFilePath)
	}()

	grokBin := a.runner.GetBinaryPath()
	if grokBin == "" {
		grokBin = grokrunner.ResolveGrokBinary()
	}

	systemInstructions := fmt.Sprintf(
		"Kamu adalah transcriber audio programmer yang sangat akurat. Analisis dan dengarkan rekaman audio teknis yang terdapat pada file berikut: %s\n"+
			"Transkripsikan seluruh isi percakapan audio tersebut dengan jelas dan rapi. Gunakan istilah teknis pemrograman, nama variabel, fungsi, bahasa pemrograman, dan tanda baca yang tepat.\n"+
			"PENTING: Keluarkan HANYA teks transkripsi murni. DILARANG KERAS menyertakan kalimat pembuka (seperti 'Berikut adalah transkripsi...', 'Hasil transkripsi:'), kalimat penutup, tanda petik pembungkus, atau penjelasan tambahan apa pun.",
		audioFilePath,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, grokBin,
		"--session-id", sessionUUID,
		"--cwd", workspacePath,
		"--output-format", "plain",
		"--no-subagents",
		"--disable-web-search",
		"-p", systemInstructions,
	)
	cmd.Dir = workspacePath
	cmd.Env = grokrunner.EnsureExecEnvironment()

	out, errExec := cmd.CombinedOutput()
	if errExec != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("Grok CLI timed out after 60s")
		}
		return "", fmt.Errorf("Grok CLI Error: %w (output: %s)", errExec, strings.TrimSpace(string(out)))
	}

	cliTranscript := strings.TrimSpace(string(out))
	cliTranscript = strings.TrimPrefix(cliTranscript, "Grok:")
	cliTranscript = cleanTranscript(cliTranscript)

	if cliTranscript == "" {
		return "", fmt.Errorf("transcription returned empty response")
	}

	return cliTranscript, nil
}

// PullWorkspaceChanges pulls upstream commits
func (a *App) PullWorkspaceChanges(workspacePath string) (string, error) {
	return workspace.ExecutePull(workspacePath)
}

// CheckForUpdates inspects GitHub Releases and local changelog for available application updates
func (a *App) CheckForUpdates(currentVersion string) (*updater.UpdateCheckResult, error) {
	osName := runtime.GOOS
	arch := runtime.GOARCH
	return updater.CheckForUpdates(currentVersion, "fiko942/aethergrok", osName, arch)
}

// GetChangelogHistory retrieves parsed release history and changelog notes
func (a *App) GetChangelogHistory() ([]updater.ReleaseInfo, error) {
	return updater.FetchReleases("fiko942/aethergrok")
}

// DownloadAndInstallUpdate coordinates streaming download, checksum verification, and native installation
func (a *App) DownloadAndInstallUpdate(assetURL, sha256URL string) error {
	ctx, cancel := context.WithCancel(context.Background())
	updater.GlobalDownloadManager.SetCancel(cancel)
	defer updater.GlobalDownloadManager.SetCancel(nil)

	emitProgress := func(p updater.UpdateProgress) {
		if a.ctx != nil {
			wailsRuntime.EventsEmit(a.ctx, "updater:progress", p)
		}
	}

	emitProgress(updater.UpdateProgress{
		Stage:   "downloading",
		Percent: 0,
		Message: "Connecting to update server...",
	})

	// 1. Download asset with real-time progress callbacks
	filePath, err := updater.DownloadAssetWithProgress(ctx, assetURL, "", emitProgress)
	if err != nil {
		if ctx.Err() == context.Canceled {
			emitProgress(updater.UpdateProgress{
				Stage:   "error",
				Percent: 0,
				Message: "Download cancelled",
			})
			return fmt.Errorf("download cancelled by user")
		}
		emitProgress(updater.UpdateProgress{
			Stage:   "error",
			Percent: 0,
			Message: fmt.Sprintf("Download failed: %v", err),
		})
		return err
	}

	// 2. Checksum verification
	emitProgress(updater.UpdateProgress{
		Stage:   "verifying",
		Percent: 100,
		Message: "Verifying package integrity...",
	})

	valid, err := updater.VerifyChecksum(filePath, sha256URL)
	if err != nil || !valid {
		emitProgress(updater.UpdateProgress{
			Stage:   "error",
			Percent: 100,
			Message: fmt.Sprintf("Integrity check failed: %v", err),
		})
		return fmt.Errorf("integrity check failed: %w", err)
	}

	// 3. Platform-specific native installer
	emitProgress(updater.UpdateProgress{
		Stage:   "installing",
		Percent: 100,
		Message: "Installing update...",
	})

	switch runtime.GOOS {
	case "darwin":
		if err := updater.ApplyUpdateMacOS(filePath, emitProgress); err != nil {
			emitProgress(updater.UpdateProgress{
				Stage:   "error",
				Percent: 100,
				Message: fmt.Sprintf("Installation failed: %v", err),
			})
			return err
		}
		// Allow helper script to start before quitting Wails
		go func() {
			time.Sleep(800 * time.Millisecond)
			if a.ctx != nil {
				wailsRuntime.Quit(a.ctx)
			}
		}()
		return nil

	case "windows":
		if err := updater.ApplyUpdateWindows(filePath, emitProgress); err != nil {
			emitProgress(updater.UpdateProgress{
				Stage:   "error",
				Percent: 100,
				Message: fmt.Sprintf("Installation failed: %v", err),
			})
			return err
		}
		// Allow NSIS to spin up before quitting Wails (only quit for executable installers)
		go func() {
			time.Sleep(800 * time.Millisecond)
			if a.ctx != nil && !strings.HasSuffix(strings.ToLower(filePath), ".zip") {
				wailsRuntime.Quit(a.ctx)
			}
		}()
		return nil

	default:
		return fmt.Errorf("automatic installation is not supported on %s yet. Please download manually", runtime.GOOS)
	}
}

// CancelUpdateDownload stops any ongoing background update download
func (a *App) CancelUpdateDownload() error {
	updater.GlobalDownloadManager.CancelActive()
	if a.ctx != nil {
		wailsRuntime.EventsEmit(a.ctx, "updater:progress", updater.UpdateProgress{
			Stage:   "idle",
			Percent: 0,
			Message: "Download cancelled",
		})
	}
	return nil
}

