package terminal

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// osFileWrapper abstracts PTY file handles across platforms
type osFileWrapper struct {
	readCloser  io.ReadCloser
	writeCloser io.WriteCloser
	file        *os.File
}

func (w *osFileWrapper) Read(p []byte) (n int, err error) {
	if w.readCloser != nil {
		return w.readCloser.Read(p)
	}
	return 0, io.EOF
}

func (w *osFileWrapper) Write(p []byte) (n int, err error) {
	if w.writeCloser != nil {
		return w.writeCloser.Write(p)
	}
	return 0, io.ErrClosedPipe
}

func (w *osFileWrapper) Close() error {
	var err1, err2 error
	if w.file != nil {
		return w.file.Close()
	}
	if w.readCloser != nil {
		err1 = w.readCloser.Close()
	}
	if w.writeCloser != nil {
		err2 = w.writeCloser.Close()
	}
	if err1 != nil {
		return err1
	}
	return err2
}

// Instance represents a single pseudo-terminal process with its own process group
type Instance struct {
	ID        string
	SessionID string
	Cmd       *exec.Cmd
	PtyFile   *osFileWrapper
	mu        sync.Mutex
	closed    bool
}

// Manager manages active pseudo-terminal instances indexed by ID and SessionID
type Manager struct {
	mu        sync.RWMutex
	terminals map[string]*Instance
	ctx       context.Context
}

// NewManager creates a new terminal manager
func NewManager() *Manager {
	return &Manager{
		terminals: make(map[string]*Instance),
	}
}

// SetContext sets Wails runtime context for event emission
func (m *Manager) SetContext(ctx context.Context) {
	m.ctx = ctx
}

// Create starts a new shell inside a dedicated process group and pseudo-terminal
func (m *Manager) Create(sessionID, termID, cwd, shell string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// If a terminal with the same ID already exists, close it first
	if existing, ok := m.terminals[termID]; ok {
		_ = existing.Close()
		delete(m.terminals, termID)
	}

	if shell == "" {
		shell = os.Getenv("SHELL")
		if shell == "" {
			if os.PathSeparator == '\\' {
				shell = "powershell.exe"
			} else {
				shell = "/bin/zsh"
			}
		}
	}

	if cwd == "" {
		cwd, _ = os.Getwd()
	}

	cmd := exec.Command(shell)
	cmd.Dir = cwd
	cmd.Env = append(os.Environ(),
		"TERM=xterm-256color",
		"COLORTERM=truecolor",
		"AETHERGROK_TERMINAL=1",
	)

	ptmx, err := startPty(cmd, 24, 80)
	if err != nil {
		return fmt.Errorf("failed to start pty: %w", err)
	}

	inst := &Instance{
		ID:        termID,
		SessionID: sessionID,
		Cmd:       cmd,
		PtyFile:   ptmx,
	}

	m.terminals[termID] = inst

	// Read output in background and emit to Wails frontend
	go func() {
		buf := make([]byte, 4096)
		for {
			n, readErr := ptmx.Read(buf)
			if n > 0 {
				data := string(buf[:n])
				if m.ctx != nil {
					runtime.EventsEmit(m.ctx, fmt.Sprintf("terminal:data:%s", termID), data)
				}
			}
			if readErr != nil {
				break
			}
		}

		_ = cmd.Wait()

		m.mu.Lock()
		delete(m.terminals, termID)
		m.mu.Unlock()

		if m.ctx != nil {
			runtime.EventsEmit(m.ctx, fmt.Sprintf("terminal:exit:%s", termID), 0)
		}
	}()

	return nil
}

// Write sends user input (keystrokes, commands) to the running PTY
func (m *Manager) Write(termID, data string) error {
	m.mu.RLock()
	inst, ok := m.terminals[termID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("terminal instance %s not found", termID)
	}

	return inst.Write([]byte(data))
}

// Resize sets the terminal window size (rows and columns)
func (m *Manager) Resize(termID string, cols, rows int) error {
	m.mu.RLock()
	inst, ok := m.terminals[termID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("terminal instance %s not found", termID)
	}

	return inst.Resize(cols, rows)
}

// Close closes a specific terminal instance and terminates all child processes
func (m *Manager) Close(termID string) error {
	m.mu.Lock()
	inst, ok := m.terminals[termID]
	if ok {
		delete(m.terminals, termID)
	}
	m.mu.Unlock()

	if !ok {
		return nil
	}

	return inst.Close()
}

// CloseSessionTerminals closes all terminal instances created within a given session
func (m *Manager) CloseSessionTerminals(sessionID string) error {
	m.mu.Lock()
	var toClose []*Instance
	for id, inst := range m.terminals {
		if sessionID == "" || inst.SessionID == sessionID {
			toClose = append(toClose, inst)
			delete(m.terminals, id)
		}
	}
	m.mu.Unlock()

	for _, inst := range toClose {
		_ = inst.Close()
	}

	return nil
}

// CloseAll shuts down all running terminals (e.g. on application exit)
func (m *Manager) CloseAll() {
	m.mu.Lock()
	var all []*Instance
	for id, inst := range m.terminals {
		all = append(all, inst)
		delete(m.terminals, id)
	}
	m.mu.Unlock()

	for _, inst := range all {
		_ = inst.Close()
	}
}

func (inst *Instance) Write(p []byte) error {
	inst.mu.Lock()
	defer inst.mu.Unlock()

	if inst.closed || inst.PtyFile == nil {
		return fmt.Errorf("terminal is closed")
	}

	_, err := inst.PtyFile.Write(p)
	return err
}

func (inst *Instance) Resize(cols, rows int) error {
	inst.mu.Lock()
	defer inst.mu.Unlock()

	if inst.closed || inst.PtyFile == nil {
		return fmt.Errorf("terminal is closed")
	}

	return resizePty(inst.PtyFile, rows, cols)
}

func (inst *Instance) isClosed() bool {
	inst.mu.Lock()
	defer inst.mu.Unlock()
	return inst.closed
}

// Close kills the entire process group and closes the PTY descriptor
func (inst *Instance) Close() error {
	inst.mu.Lock()
	if inst.closed {
		inst.mu.Unlock()
		return nil
	}
	inst.closed = true
	ptmx := inst.PtyFile
	inst.PtyFile = nil
	cmd := inst.Cmd
	inst.mu.Unlock()

	if cmd != nil {
		_ = killProcessTree(cmd)
	}

	if ptmx != nil {
		_ = ptmx.Close()
	}

	return nil
}
