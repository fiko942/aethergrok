package main

import (
	"context"
	"fmt"
	"os"

	"aethergrok/pkg/grokrunner"
	"aethergrok/pkg/screen"
	"aethergrok/pkg/skills"

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

// DeleteGrokSession removes a session folder from ~/.grok/sessions/
func (a *App) DeleteGrokSession(workspacePath, sessionID string) error {
	return grokrunner.DeleteGrokSessionDirectory(workspacePath, sessionID)
}

// SearchSkills queries skills by text query and category
func (a *App) SearchSkills(query string, category string) []skills.Skill {
	if a.skillsReg == nil {
		a.skillsReg = skills.NewRegistry()
	}
	return a.skillsReg.Search(query, category)
}

// GetAvailableModels returns the list of available models from Grok CLI
func (a *App) GetAvailableModels() []grokrunner.ModelInfo {
	if a.runner == nil {
		a.runner = grokrunner.NewRunner()
	}
	return a.runner.DiscoverAvailableModels()
}
