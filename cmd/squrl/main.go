package main

import (
	"github.com/getlantern/systray"
	"github.com/thismarioperez/squrl/internal/tray"
)

func main() {
	// systray.Run must be called from main() on macOS — it takes ownership of the
	// Cocoa main thread and runs the AppKit event loop until Quit is triggered.
	systray.Run(tray.OnReady, tray.OnExit)
}
