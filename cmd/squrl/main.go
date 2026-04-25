package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	xterm "github.com/charmbracelet/x/term"
	"github.com/thismarioperez/squrl/internal/cli"
	"github.com/thismarioperez/squrl/internal/logging"
)

// version is set at build time via -ldflags "-X main.version=vX.Y.Z".
var version = "dev"

func main() {
	if len(os.Args) == 2 && (os.Args[1] == "--version" || os.Args[1] == "-version") {
		fmt.Println(version)
		os.Exit(0)
	}

	cleanup := logging.Init()
	defer cleanup()

	opts, err := cli.ParseScanArgs(os.Args[1:])
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "squrl: %v\n", err)
		os.Exit(2)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if opts.NonInteractive || !isTerminal(os.Stdout) {
		os.Exit(cli.RunNonInteractive(ctx, opts))
	} else {
		os.Exit(cli.RunTUI(ctx, opts, version))
	}
}

func isTerminal(f *os.File) bool {
	return xterm.IsTerminal(f.Fd())
}
