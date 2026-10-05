package grokrunner

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"aethergrok/pkg/storage"
)

// ActiveSession holds running process information for a session
type ActiveSession struct {
	SessionID string
	Cmd       *exec.Cmd
	Cancel    context.CancelFunc
	Stdin     io.WriteCloser
	Stdout    io.ReadCloser
	Done      chan struct{}
}

// Runner manages execution of Grok CLI subprocesses
type Runner struct {
	mu             sync.Mutex
	sessions       map[string]*ActiveSession
	grokBinaryPath string
	storageMgr     *storage.StorageManager
}

// SetStorageManager binds persistent storage manager for automatic permission modes
func (r *Runner) SetStorageManager(sm *storage.StorageManager) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.storageMgr = sm
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
	if runtime.GOOS == "windows" {
		if p, err := exec.LookPath("grok.exe"); err == nil {
			return p
		}
		if p, err := exec.LookPath("grok.cmd"); err == nil {
			return p
		}
	}

	// 3. User home directory locations
	homeDir, err := os.UserHomeDir()
	if err == nil && homeDir != "" {
		var candidates []string
		if runtime.GOOS == "windows" {
			candidates = []string{
				filepath.Join(homeDir, ".grok", "bin", "grok.exe"),
				filepath.Join(homeDir, ".grok", "bin", "grok.cmd"),
				filepath.Join(homeDir, "AppData", "Local", "Programs", "grok", "grok.exe"),
				filepath.Join(homeDir, "AppData", "Roaming", "npm", "grok.cmd"),
				filepath.Join(homeDir, "AppData", "Roaming", "npm", "grok.exe"),
				filepath.Join(homeDir, "AppData", "Local", "pnpm", "grok.cmd"),
				filepath.Join(homeDir, "AppData", "Local", "pnpm", "grok.exe"),
				filepath.Join(homeDir, ".cargo", "bin", "grok.exe"),
				filepath.Join(homeDir, "go", "bin", "grok.exe"),
				filepath.Join(homeDir, "bin", "grok.exe"),
			}
			if pf := os.Getenv("ProgramFiles"); pf != "" {
				candidates = append(candidates, filepath.Join(pf, "grok", "grok.exe"))
			}
			if pfx := os.Getenv("ProgramFiles(x86)"); pfx != "" {
				candidates = append(candidates, filepath.Join(pfx, "grok", "grok.exe"))
			}
		} else {
			candidates = []string{
				filepath.Join(homeDir, ".local", "bin", "grok"),
				filepath.Join(homeDir, ".grok", "bin", "grok"),
				filepath.Join(homeDir, "bin", "grok"),
				filepath.Join(homeDir, "go", "bin", "grok"),
				filepath.Join(homeDir, ".cargo", "bin", "grok"),
			}
		}
		for _, c := range candidates {
			if info, err := os.Stat(c); err == nil && !info.IsDir() {
				return c
			}
		}
	}

	// 4. System-wide locations
	if runtime.GOOS != "windows" {
		systemCandidates := []string{
			"/opt/homebrew/bin",
			"/usr/local/bin/grok",
			"/usr/bin/grok",
			"/bin/grok",
		}
		for _, sc := range systemCandidates {
			if info, err := os.Stat(sc); err == nil && !info.IsDir() {
				return sc
			}
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
		if runtime.GOOS == "windows" {
			extraPaths = append(extraPaths,
				filepath.Join(homeDir, ".grok", "bin"),
				filepath.Join(homeDir, "AppData", "Local", "Programs", "grok"),
				filepath.Join(homeDir, "AppData", "Roaming", "npm"),
				filepath.Join(homeDir, "AppData", "Local", "pnpm"),
				filepath.Join(homeDir, ".cargo", "bin"),
				filepath.Join(homeDir, "go", "bin"),
				filepath.Join(homeDir, "bin"),
			)
			if pf := os.Getenv("ProgramFiles"); pf != "" {
				extraPaths = append(extraPaths,
					filepath.Join(pf, "Go", "bin"),
					filepath.Join(pf, "Git", "cmd"),
					filepath.Join(pf, "Git", "bin"),
				)
			}
		} else {
			extraPaths = append(extraPaths,
				filepath.Join(homeDir, ".local", "bin"),
				filepath.Join(homeDir, ".grok", "bin"),
				filepath.Join(homeDir, "bin"),
				filepath.Join(homeDir, "go", "bin"),
				filepath.Join(homeDir, ".cargo", "bin"),
				"/opt/homebrew/bin",
				"/opt/homebrew/sbin",
				"/usr/local/bin",
				"/usr/bin",
				"/bin",
				"/usr/sbin",
				"/sbin",
			)
		}
	}

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
	} else if _, err := os.Stat(r.grokBinaryPath); err != nil && !filepath.IsAbs(r.grokBinaryPath) {
		if _, lookErr := exec.LookPath(r.grokBinaryPath); lookErr != nil {
			r.grokBinaryPath = ResolveGrokBinary()
		}
	} else if _, err := os.Stat(r.grokBinaryPath); err != nil && filepath.IsAbs(r.grokBinaryPath) {
		r.grokBinaryPath = ResolveGrokBinary()
	}
	return r.grokBinaryPath
}

// isUUID checks if a string is a standard UUID format (36 chars with dashes)
func isUUID(str string) bool {
	if len(str) != 36 {
		return false
	}
	for i, ch := range str {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if ch != '-' {
				return false
			}
		} else {
			if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')) {
				return false
			}
		}
	}
	return true
}

func generateUUIDv4() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40 // Version 4
	b[8] = (b[8] & 0x3f) | 0x80 // Variant RFC4122
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
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

	// Determine session continuation vs new session:
	// If GrokSessionID is specified, or req.SessionID is a valid UUID, target that UUID.
	targetGrokID := req.Options.GrokSessionID
	if targetGrokID == "" && isUUID(req.SessionID) {
		targetGrokID = req.SessionID
	}

	homeDir, _ := os.UserHomeDir()
	wsPath := req.Options.WorkingDir
	if wsPath == "" {
		wsPath, _ = os.Getwd()
	}
	sessionsDir := filepath.Join(homeDir, ".grok", "sessions")
	targetWsDir := ResolveWorkspaceSessionsDir(sessionsDir, wsPath)

	if targetGrokID != "" {
		sessionFolder, resolvedID := ResolveSessionFolder(targetWsDir, targetGrokID)
		
		// If the resolved session folder exists on disk, resume it.
		// If grokSessionId was provided, prioritize resume.
		if req.Options.GrokSessionID != "" {
			if resolvedID != "" {
				args = append(args, "--resume", resolvedID)
			} else {
				args = append(args, "--resume", req.Options.GrokSessionID)
			}
		} else if fi, err := os.Stat(sessionFolder); err == nil && fi.IsDir() {
			args = append(args, "--resume", resolvedID)
		} else if isUUID(targetGrokID) {
			// Brand new session with a valid UUID: explicit session-id
			args = append(args, "--session-id", targetGrokID)
		}
	} else {
		// If no UUID was provided (e.g. temporary legacy frontend ID), generate a fresh UUID
		// so Grok CLI creates an isolated session rather than resuming or colliding with existing ones.
		newUUID := generateUUIDv4()
		args = append(args, "--session-id", newUUID)
	}

	// Check if default model or options need fallback from storage settings
	if r.storageMgr != nil {
		if st, err := r.storageMgr.GetSettings(); err == nil {
			if req.Options.Model == "" && st.DefaultModel != "" {
				args = append(args, "--model", st.DefaultModel)
			}
			if req.Options.ReasoningEffort == "" && st.DefaultReasoningEffort != "" {
				args = append(args, "--reasoning-effort", st.DefaultReasoningEffort)
			}
		}
	}

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

	// Prepare prompt arguments:
	// If images are attached or prompt text is provided, write content blocks or text to a temporary
	// prompt file and pass `--prompt-file <path>` to prevent Win32 CreateProcess command line limit
	// (32,767 characters) errors when passing base64 encoded images.
	var promptTempFile string
	if len(req.Images) > 0 {
		var contentBlocks []map[string]interface{}
		for _, imgPath := range req.Images {
			imgPath = strings.TrimSpace(imgPath)
			if imgPath == "" {
				continue
			}
			dataBytes, err := os.ReadFile(imgPath)
			if err != nil {
				continue
			}
			mimeType := "image/png"
			ext := strings.ToLower(filepath.Ext(imgPath))
			if ext == ".jpg" || ext == ".jpeg" {
				mimeType = "image/jpeg"
			} else if ext == ".webp" {
				mimeType = "image/webp"
			} else if ext == ".gif" {
				mimeType = "image/gif"
			}
			contentBlocks = append(contentBlocks, map[string]interface{}{
				"type":     "image",
				"data":     base64.StdEncoding.EncodeToString(dataBytes),
				"mimeType": mimeType,
			})
		}

		userText := req.Prompt
		if userText == "" {
			userText = "Describe and analyze this image."
		}
		contentBlocks = append(contentBlocks, map[string]interface{}{
			"type": "text",
			"text": userText,
		})

		jsonBytes, err := json.Marshal(contentBlocks)
		if err == nil {
			tmpFile, tmpErr := os.CreateTemp("", "grok-prompt-*.json")
			if tmpErr == nil {
				_, _ = tmpFile.Write(jsonBytes)
				_ = tmpFile.Close()
				promptTempFile = tmpFile.Name()
				args = append(args, "--prompt-file", promptTempFile)
			} else {
				args = append(args, "-p", req.Prompt)
			}
		} else {
			args = append(args, "-p", req.Prompt)
		}
	} else {
		// Non-interactive text-only prompt turn
		args = append(args, "-p", req.Prompt)
	}

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

	stderr, err := cmd.StderrPipe()
	if err != nil {
		_ = stdout.Close()
		r.mu.Unlock()
		cancel()
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		_ = stdout.Close()
		_ = stderr.Close()
		r.mu.Unlock()
		cancel()
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	active := &ActiveSession{
		SessionID: req.SessionID,
		Cmd:       cmd,
		Cancel:    cancel,
		Stdin:     stdin,
		Stdout:    stdout,
		Done:      make(chan struct{}),
	}
	r.sessions[req.SessionID] = active
	r.mu.Unlock()

	if err := cmd.Start(); err != nil {
		if promptTempFile != "" {
			_ = os.Remove(promptTempFile)
		}
		_ = stdout.Close()
		_ = stderr.Close()
		r.mu.Lock()
		delete(r.sessions, req.SessionID)
		r.mu.Unlock()
		cancel()
		close(active.Done)
		return fmt.Errorf("failed to start grok command: %w", err)
	}

	var stderrBuf bytes.Buffer
	var stderrMu sync.Mutex
	go func() {
		buf := make([]byte, 4096)
		for {
			n, rErr := stderr.Read(buf)
			if n > 0 {
				stderrMu.Lock()
				if stderrBuf.Len() < 64*1024 {
					stderrBuf.Write(buf[:n])
				}
				stderrMu.Unlock()
			}
			if rErr != nil {
				break
			}
		}
	}()

	parser := NewStreamParser(req.SessionID, callbacks)

	// Startup watchdog: if Grok CLI produces zero stdout lines within 45s (e.g. startup deadlock / corrupted history),
	// terminate the process immediately and fail fast.
	const startupTimeout = 45 * time.Second
	var timedOutDuringStartup atomic.Bool
	watchdogDone := make(chan struct{})

	go func() {
		defer close(watchdogDone)
		select {
		case <-sessionCtx.Done():
		case <-parser.FirstLineChan():
		case <-time.After(startupTimeout):
			timedOutDuringStartup.Store(true)
			_ = killProcessGroup(cmd)
			_ = stdout.Close()
			cancel()
		}
	}()

	go func() {
		defer func() {
			if promptTempFile != "" {
				_ = os.Remove(promptTempFile)
			}
			_ = r.cleanupSession(req.SessionID)
			close(active.Done)
			cancel()
		}()

		parseErr := parser.Parse(sessionCtx, stdout)
		waitErr := cmd.Wait()
		<-watchdogDone

		// Guaranteed Turn Completion Fallback:
		// If parser did not receive an explicit completion signal before process exit/EOF,
		// emit completion now so the frontend session never hangs in 'working' state.
		if !parser.HasCompleted() {
			status := "success"
			var errStr string

			stderrMu.Lock()
			capturedStderr := strings.TrimSpace(stderrBuf.String())
			stderrMu.Unlock()

			if timedOutDuringStartup.Load() {
				status = "error"
				errStr = "Grok CLI startup timed out after 45s (no response from CLI process). The session history may be too large or corrupted."
				if capturedStderr != "" {
					errStr += "\nCLI Output:\n" + capturedStderr
				}
			} else if sessionCtx.Err() != nil {
				status = "interrupted"
				errStr = "Turn interrupted"
			} else if waitErr != nil {
				status = "error"
				errStr = waitErr.Error()
				if capturedStderr != "" {
					errStr += "\nCLI Output:\n" + capturedStderr
				}
			} else if parseErr != nil && parseErr != io.EOF {
				status = "error"
				errStr = parseErr.Error()
				if capturedStderr != "" {
					errStr += "\nCLI Output:\n" + capturedStderr
				}
			}

			if status == "error" && callbacks.OnError != nil {
				callbacks.OnError(fmt.Errorf("%s", errStr))
			}
			parser.EmitComplete(status, errStr)
		}
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
	if active.Stdout != nil {
		_ = active.Stdout.Close()
	}
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
