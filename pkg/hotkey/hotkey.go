package hotkey

import "sync"

// Handler is a callback function invoked when the hotkey triggers
type Handler func()

// Manager manages global shortcut registration and OS event listening
type Manager struct {
	mu          sync.Mutex
	shortcutStr string
	handler     Handler
	active      bool
	impl        platformManager
}

type platformManager interface {
	start(shortcutStr string, handler Handler) error
	update(shortcutStr string) error
	stop()
}

// NewManager creates a new global hotkey manager
func NewManager() *Manager {
	m := &Manager{
		impl: newPlatformManager(),
	}
	return m
}

// SetHandler sets the callback function when hotkey triggers
func (m *Manager) SetHandler(h Handler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handler = h
}

// RegisterShortcut sets and activates a global shortcut string (e.g., "ShiftRight", "Right Shift", "CmdOrCtrl+Shift+S")
func (m *Manager) RegisterShortcut(shortcutStr string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.shortcutStr = shortcutStr
	if m.impl != nil {
		if !m.active {
			err := m.impl.start(shortcutStr, func() {
				m.mu.Lock()
				h := m.handler
				m.mu.Unlock()
				if h != nil {
					h()
				}
			})
			if err == nil {
				m.active = true
			}
			return err
		} else {
			return m.impl.update(shortcutStr)
		}
	}
	return nil
}

// Unregister stops the hotkey listening
func (m *Manager) Unregister() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.active && m.impl != nil {
		m.impl.stop()
		m.active = false
	}
}
