package main

// The automation plugin lets a test drive the app's native UI: it lists
// the windows, reads their elements, acts on them, and captures them. It
// listens only when MYGO_AUTOMATION names a socket, so it costs nothing
// when the app runs on its own.
import _ "github.com/egoist/mygo/plugins/automation"
