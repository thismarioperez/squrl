package main

import (
	"fmt"
	"os"

	"github.com/getlantern/systray"
	"github.com/thismarioperez/squrl/internal/tray"
)

// version is set at build time via -ldflags "-X main.version=vX.Y.Z".
var version = "dev"

func main() {
	if len(os.Args) == 2 && (os.Args[1] == "--version" || os.Args[1] == "-version") {
		fmt.Println(version)
		os.Exit(0)
	}

	tray.SetVersion(version)

	// systray.Run must be called from main() on macOS — it takes ownership of the
	// Cocoa main thread and runs the AppKit event loop until Quit is triggered.
	systray.Run(tray.OnReady, tray.OnExit)
}
