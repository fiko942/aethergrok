//go:build !darwin

package hotkey

type dummyManager struct{}

func newPlatformManager() platformManager {
	return &dummyManager{}
}

func (d *dummyManager) start(shortcutStr string, handler Handler) error {
	return nil
}

func (d *dummyManager) update(shortcutStr string) error {
	return nil
}

func (d *dummyManager) stop() {}
