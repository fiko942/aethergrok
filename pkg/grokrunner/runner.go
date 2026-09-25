package grokrunner

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// ActiveSession holds running process information for a session
type ActiveSession struct {
	SessionID string
	Cmd       *exec.Cmd
	Cancel    context.CancelFunc
	Stdin     io.WriteCloser
	Done      chan struct{}
}

// Runner manages execution of Grok CLI subprocesses
type Runner struct {
	mu             sync.Mutex
	sessions       map[string]*ActiveSession
	grokBinaryPath string
}

// ResolveGrokBinary searches for the grok CLI binary across standard paths and environment
func ResolveGrokBinary() string {
	// 1. Explicit environment override
	if binPath := os.Getenv("GROK_BIN_PATH"); binPath != "" {
		if _, err := os.Stat(binPath); err == nil {
			return binPath
		}
	}

	// 2. Standard LookPath via active PATH
	if p, err := exec.LookPath("grok"); err == nil {
		return p
	}

	// 3. User home directory locations (common when launched from macOS Finder / GUI app without shell PATH)
	homeDir, err := os.UserHomeDir()
	if err == nil && homeDir != "" {
		candidates := []string{
			filepath.Join(homeDir, ".local", "bin", "grok"),
			filepath.Join(homeDir, ".grok", "bin", "grok"),
			filepath.Join(homeDir, "bin", "grok"),
			filepath.Join(homeDir, "go", "bin", "grok"),
		}
		if runtime.GOOS == "windows" {
			candidates = append(candidates,
				filepath.Join(homeDir, "AppData", "Local", "Programs", "grok", "grok.exe"),
				filepath.Join(homeDir, ".grok", "bin", "grok.exe"),
				filepath.Join(homeDir, ".grok", "bin", "grok.cmd"),
			)
		}
		for _, c := range candidates {
			if info, err := os.Stat(c); err == nil && !info.IsDir() {
				return c
			}
		}
	}

	// 4. System-wide locations
	systemCandidates := []string{
		"/opt/homebrew/bin/grok",
		"/usr/local/bin/grok",
		"/usr/bin/grok",
		"/bin/grok",
	}
	for _, sc := range systemCandidates {
		if info, err := os.Stat(sc); err == nil && !info.IsDir() {
			return sc
		}
	}

	return "grok"
}

// EnsureExecEnvironment returns a copy of the environment containing standard PATH entries
func EnsureExecEnvironment() []string {
	env := os.Environ()
	homeDir, _ := os.UserHomeDir()

	var extraPaths []string
	if homeDir != "" {
		extraPaths = append(extraPaths,
			filepath.Join(homeDir, ".local", "bin"),
			filepath.Join(homeDir, ".grok", "bin"),
			filepath.Join(homeDir, "bin"),
			filepath.Join(homeDir, "go", "bin"),
		)
	}
	extraPaths = append(extraPaths,
		"/opt/homebrew/bin",
		"/opt/homebrew/sbin",
		"/usr/local/bin",
		"/usr/bin",
		"/bin",
		"/usr/sbin",
		"/sbin",
	)

	currentPath := os.Getenv("PATH")
	existingParts := strings.Split(currentPath, string(os.PathListSeparator))
	seen := make(map[string]bool)
	for _, p := range existingParts {
		if p != "" {
			seen[p] = true
		}
	}

	merged := existingParts
	for _, ep := range extraPaths {
		if !seen[ep] {
			merged = append(merged, ep)
			seen[ep] = true
		}
	}

	newPath := strings.Join(merged, string(os.PathListSeparator))

	// Rebuild env with updated PATH
	pathKey := "PATH="
	if runtime.GOOS == "windows" {
		pathKey = "Path="
	}

	var newEnv []string
	found := false
	for _, e := range env {
		if strings.HasPrefix(strings.ToUpper(e), "PATH=") {
			newEnv = append(newEnv, pathKey+newPath)
			found = true
		} else {
			newEnv = append(newEnv, e)
		}
	}
	if !found {
		newEnv = append(newEnv, pathKey+newPath)
	}

	return newEnv
}

// NewRunner creates a new Runner instance
func NewRunner() *Runner {
	return &Runner{
		sessions:       make(map[string]*ActiveSession),
		grokBinaryPath: ResolveGrokBinary(),
	}
}

// SetBinaryPath configures the grok CLI binary path
func (r *Runner) SetBinaryPath(path string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.grokBinaryPath = path
}

// GetBinaryPath returns the configured grok binary path
func (r *Runner) GetBinaryPath() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.grokBinaryPath == "" || r.grokBinaryPath == "grok" {
		r.grokBinaryPath = ResolveGrokBinary()
	}
	return r.grokBinaryPath
}

// StartSession launches a grok subprocess for a prompt request and streams events via callbacks
func (r *Runner) StartSession(ctx context.Context, req PromptRequest, callbacks StreamCallbacks) error {
	r.mu.Lock()
	if existing, exists := r.sessions[req.SessionID]; exists {
		r.mu.Unlock()
		// Cancel existing session turn if active
		_ = r.Cancel(req.SessionID)
		<-existing.Done
		r.mu.Lock()
	}

	sessionCtx, cancel := context.WithCancel(ctx)

	args := []string{"--output-format", "streaming-json"}
	if req.Options.Model != "" {
		args = append(args, "--model", req.Options.Model)
	}
	if req.Options.ReasoningEffort != "" {
		args = append(args, "--reasoning-effort", req.Options.ReasoningEffort)
	}
	if req.Options.DisableTools {
		args = append(args, "--no-tools")
	}
	if req.Options.SystemPrompt != "" {
		args = append(args, "--system", req.Options.SystemPrompt)
	}
	for _, sDir := range req.Options.SkillDirs {
		args = append(args, "--skill-dir", sDir)
	}
	for _, flag := range req.Options.CustomFlags {
		args = append(args, flag)
	}

	for _, img := range req.Images {
		args = append(args, "--image", img)
	}

	// Non-interactive prompt turn
	args = append(args, "-p", req.Prompt)

	binPath := r.grokBinaryPath
	if binPath == "" || binPath == "grok" {
		binPath = ResolveGrokBinary()
		r.grokBinaryPath = binPath
	}

	cmd := exec.CommandContext(sessionCtx, binPath, args...)
	cmd.Env = EnsureExecEnvironment()
	if req.Options.WorkingDir != "" {
		cmd.Dir = req.Options.WorkingDir
	}

	// Set Process Group so all children / subagents can be killed cleanly
	setSysProcGroup(cmd)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		r.mu.Unlock()
		cancel()
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		r.mu.Unlock()
		cancel()
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	active := &ActiveSession{
		SessionID: req.SessionID,
		Cmd:       cmd,
		Cancel:    cancel,
		Stdin:     stdin,
		Done:      make(chan struct{}),
	}
	r.sessions[req.SessionID] = active
	r.mu.Unlock()

	if err := cmd.Start(); err != nil {
		r.mu.Lock()
		delete(r.sessions, req.SessionID)
		r.mu.Unlock()
		cancel()
		close(active.Done)
		return fmt.Errorf("failed to start grok command: %w", err)
	}

	go func() {
		defer func() {
			_ = r.cleanupSession(req.SessionID)
			close(active.Done)
			cancel()
		}()

		parser := NewStreamParser(req.SessionID, callbacks)
		_ = parser.Parse(sessionCtx, stdout)
		_ = cmd.Wait()
	}()

	return nil
}

// RespondPermission writes user permission approval or rejection to grok process stdin
func (r *Runner) RespondPermission(resp PermissionResponse) error {
	r.mu.Lock()
	active, exists := r.sessions[resp.SessionID]
	r.mu.Unlock()

	if !exists || active.Stdin == nil {
		return fmt.Errorf("active session not found: %s", resp.SessionID)
	}

	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = active.Stdin.Write(data)
	return err
}

// Cancel terminates the process group of the active session
func (r *Runner) Cancel(sessionID string) error {
	r.mu.Lock()
	active, exists := r.sessions[sessionID]
	r.mu.Unlock()

	if !exists {
		return nil
	}

	active.Cancel()
	return killProcessGroup(active.Cmd)
}

// CancelAll terminates all active sessions
func (r *Runner) CancelAll() {
	r.mu.Lock()
	ids := make([]string, 0, len(r.sessions))
	for id := range r.sessions {
		ids = append(ids, id)
	}
	r.mu.Unlock()

	for _, id := range ids {
		_ = r.Cancel(id)
	}
}

func (r *Runner) cleanupSession(sessionID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, sessionID)
	return nil
}
