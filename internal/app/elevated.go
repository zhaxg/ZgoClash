package app

import (
	"os"
	"strconv"
	"strings"
)

// WaitForElevatedRestart waits for the instance that started this one as an
// administrator to go. It is asked before the single instance lock: the new
// instance would otherwise find the lock held, hand its command line over to
// the old one and quit, and the old one would quit right after, leaving
// nothing running.
func WaitForElevatedRestart() {
	for _, arg := range os.Args[1:] {
		if !strings.HasPrefix(arg, waitForPidArg+"=") {
			continue
		}
		pid, err := strconv.Atoi(strings.TrimPrefix(arg, waitForPidArg+"="))
		if err != nil {
			continue
		}
		waitForProcess(pid)
		return
	}
}
