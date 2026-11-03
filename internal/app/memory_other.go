//go:build !windows

package app

// processMemory has no portable working-set API here; the dashboard falls back
// to the Go runtime's own numbers.
func processMemory() uint64 { return 0 }
