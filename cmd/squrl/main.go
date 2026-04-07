package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/getlantern/systray"
	"github.com/thismarioperez/squrl/internal/cli"
	"github.com/thismarioperez/squrl/internal/logging"
	"github.com/thismarioperez/squrl/internal/tray"
)

// version is set at build time via -ldflags "-X main.version=vX.Y.Z".
var version = "dev"

func main() {
	if len(os.Args) == 2 && (os.Args[1] == "--version" || os.Args[1] == "-version") {
		fmt.Println(version)
		os.Exit(0)
	}

	if len(os.Args) >= 2 && os.Args[1] == "scan" {
		cleanup := logging.Init()
		defer cleanup()

		opts, err := cli.ParseScanArgs(os.Args[2:])
		if err != nil {
			if !errors.Is(err, flag.ErrHelp) {
				fmt.Fprintf(os.Stderr, "squrl scan: %v\n", err)
			}
			os.Exit(2)
		}

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		os.Exit(cli.Scan(ctx, opts))
	}

	cleanup := logging.Init()
	defer cleanup()

	tray.SetVersion(version)

	// systray.Run must be called from main() on macOS — it takes ownership of the
	// Cocoa main thread and runs the AppKit event loop until Quit is triggered.
	systray.Run(tray.OnReady, tray.OnExit)
}
