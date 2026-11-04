package app

import (
	"testing"
)

// TestStopKeepsSystemProxyChoice checks that stopping the core does not
// forget the choice of the user: the proxy in Windows goes away with the
// core, but the setting is remembered, so a restart does not lose it.
func TestStopKeepsSystemProxyChoice(t *testing.T) {
	old := applySystemProxy
	var cleared bool
	applySystemProxy = func(enable bool, _ string) error {
		if !enable {
			cleared = true
		}
		return nil
	}
	defer func() { applySystemProxy = old }()

	c := NewController(t.TempDir())
	if err := c.store.ensure(); err != nil {
		t.Fatal(err)
	}
	c.mu.Lock()
	c.settings.SystemProxy = true
	c.mu.Unlock()
	if err := c.store.SaveSettings(c.Settings()); err != nil {
		t.Fatal(err)
	}

	c.stopLocked(true)

	if !cleared {
		t.Error("the proxy in Windows stayed on after the core stopped")
	}
	if !c.Settings().SystemProxy {
		t.Error("stopping the core forgot the system proxy")
	}
	if !c.store.LoadSettings().SystemProxy {
		t.Error("stopping the core wrote the system proxy off in the settings")
	}
}
