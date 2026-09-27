package terminal

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/creack/pty"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Instance represents a single pseudo-terminal process with its own process group
type Instance struct {
	ID        string
	SessionID string
	Cmd       *exec.Cmd
	PtyFile   *os.File
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
			shell = "/bin/zsh"
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

	// In sandbox/subagent environments or non-root macos, pty.Start handles process creation
	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{
		Rows: 24,
		Cols: 80,
	})
	if err != nil {
		// Fallback without special sysprocattr if restricted
		return fmt.Errorf("failed to start pty: %w", err)
	}

	// Set process group if not already set by pty
	if cmd.Process != nil {
		_ = syscall.Setpgid(cmd.Process.Pid, cmd.Process.Pid)
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
			n, err := ptmx.Read(buf)
			if n > 0 {
				data := string(buf[:n])
				if m.ctx != nil {
					runtime.EventsEmit(m.ctx, "terminal:data:"+termID, data)
				}
			}
			if err != nil {
				if err != io.EOF && !inst.isClosed() {
					// PTY read ended
				}
				break
			}
		}

		// Process exited, clean up
		_ = inst.Close()
		m.mu.Lock()
		delete(m.terminals, termID)
		m.mu.Unlock()

		if m.ctx != nil {
			runtime.EventsEmit(m.ctx, "terminal:exit:"+termID, map[string]interface{}{
				"termId": termID,
			})
		}
	}()

	return nil
}

// Write writes stdin data to the pseudo-terminal
func (m *Manager) Write(termID string, data string) error {
	m.mu.RLock()
	inst, ok := m.terminals[termID]
	m.mu.RUnlock()

	if !ok || inst == nil {
		return fmt.Errorf("terminal %s not found", termID)
	}

	return inst.Write([]byte(data))
}

// Resize resizes the pseudo-terminal window
func (m *Manager) Resize(termID string, cols, rows int) error {
	m.mu.RLock()
	inst, ok := m.terminals[termID]
	m.mu.RUnlock()

	if !ok || inst == nil {
		return fmt.Errorf("terminal %s not found", termID)
	}

	return inst.Resize(cols, rows)
}

// Close closes a specific terminal instance and terminates its entire process tree
func (m *Manager) Close(termID string) error {
	m.mu.Lock()
	inst, ok := m.terminals[termID]
	if ok {
		delete(m.terminals, termID)
	}
	m.mu.Unlock()

	if !ok || inst == nil {
		return nil
	}

	return inst.Close()
}

// CloseSessionTerminals closes and kills all terminals belonging to a session
func (m *Manager) CloseSessionTerminals(sessionID string) error {
	m.mu.Lock()
	var toClose []*Instance
	for id, inst := range m.terminals {
		if inst.SessionID == sessionID {
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

// Write writes data to the instance's PTY
func (inst *Instance) Write(p []byte) error {
	inst.mu.Lock()
	defer inst.mu.Unlock()

	if inst.closed || inst.PtyFile == nil {
		return fmt.Errorf("terminal is closed")
	}

	_, err := inst.PtyFile.Write(p)
	return err
}

// Resize resizes the instance PTY
func (inst *Instance) Resize(cols, rows int) error {
	inst.mu.Lock()
	defer inst.mu.Unlock()

	if inst.closed || inst.PtyFile == nil {
		return fmt.Errorf("terminal is closed")
	}

	return pty.Setsize(inst.PtyFile, &pty.Winsize{
		Rows: uint16(rows),
		Cols: uint16(cols),
	})
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

	if cmd != nil && cmd.Process != nil {
		pid := cmd.Process.Pid
		pgid, err := syscall.Getpgid(pid)

		if err == nil && pgid > 0 {
			// Sinyal SIGTERM ke seluruh process group (-pgid)
			_ = syscall.Kill(-pgid, syscall.SIGTERM)

			// Tunggu sebentar (150ms)
			done := make(chan error, 1)
			go func() {
				state, _ := cmd.Process.Wait()
				if state != nil && state.Exited() {
					done <- nil
				} else {
					done <- fmt.Errorf("still running")
				}
			}()

			select {
			case <-done:
				// Proses sudah keluar bersih
			case <-time.After(150 * time.Millisecond):
				// Eskalasi ke SIGKILL untuk membersihkan SEMUA background electron/vite/node processes
				_ = syscall.Kill(-pgid, syscall.SIGKILL)
			}
		} else {
			// Fallback jika pgid tidak bisa diambil
			_ = cmd.Process.Kill()
		}
	}

	if ptmx != nil {
		_ = ptmx.Close()
	}

	return nil
}
