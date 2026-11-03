package app

import "runtime/debug"

// setIdleMemory trades a little CPU for a smaller working set while no window
// shows: a tighter GC target keeps the heap small, and FreeOSMemory runs a
// collection and returns the pages the UI freed to the system. Showing the
// window puts the default target back.
func setIdleMemory(idle bool) {
	if idle {
		debug.SetGCPercent(40)
		debug.FreeOSMemory()
		return
	}
	debug.SetGCPercent(100)
}
