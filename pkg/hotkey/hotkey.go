package hotkey

import "sync"

// Handler is a callback function invoked when the hotkey triggers
type Handler func()

// KeyHandler is a callback function invoked when hotkey state changes ("down" or "up")
type KeyHandler func(action string)

// Manager manages global shortcut registration and OS event listening
type Manager struct {
	mu          sync.Mutex
	shortcutStr string
	handler     Handler
	keyHandler  KeyHandler
	active      bool
	impl        platformManager
}

type platformManager interface {
	start(shortcutStr string, handler Handler) error
	startWithKeyHandler(shortcutStr string, handler Handler, keyHandler KeyHandler) error
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

// SetKeyHandler sets the callback function when hotkey key down / up occurs
func (m *Manager) SetKeyHandler(kh KeyHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.keyHandler = kh
}

// RegisterShortcut sets and activates a global shortcut string (e.g., "ShiftRight", "Right Shift", "CmdOrCtrl+Shift+S")
func (m *Manager) RegisterShortcut(shortcutStr string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.shortcutStr = shortcutStr
	if m.impl != nil {
		if !m.active {
			err := m.impl.startWithKeyHandler(
				shortcutStr,
				func() {
					m.mu.Lock()
					h := m.handler
					m.mu.Unlock()
					if h != nil {
						h()
					}
				},
				func(action string) {
					m.mu.Lock()
					kh := m.keyHandler
					m.mu.Unlock()
					if kh != nil {
						kh(action)
					}
				},
			)
			if err == nil {
				m.active = true
			}
			return err
		}
		return m.impl.update(shortcutStr)
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
