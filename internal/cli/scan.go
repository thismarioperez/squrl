package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/thismarioperez/squrl/assets"
	"github.com/thismarioperez/squrl/internal/scanner"
)

// ScanOptions holds configuration for a CLI scan run.
type ScanOptions struct {
	Delay int // countdown seconds before scan; 0 = skip countdown
}

// ParseScanArgs parses the arguments after "scan" into ScanOptions.
func ParseScanArgs(args []string) (ScanOptions, error) {
	fs := flag.NewFlagSet("scan", flag.ContinueOnError)
	delay := fs.Int("delay", 3, "seconds to wait before scanning (0 to skip countdown)")
	fs.IntVar(delay, "D", 3, "shorthand for --delay")
	if err := fs.Parse(args); err != nil {
		return ScanOptions{}, err
	}
	if *delay < 0 {
		return ScanOptions{}, fmt.Errorf("--delay must be >= 0")
	}
	return ScanOptions{Delay: *delay}, nil
}

// Scan runs an optional countdown, scans all screens for QR codes, and prints
// results to stdout (one per line). Status and errors go to stderr.
// Returns 0 if results found, 1 if no QR codes found, 2 on error or cancellation.
func Scan(ctx context.Context, opts ScanOptions) int {
	slog.Debug("cli scan started", "delay", opts.Delay)

	fmt.Fprintf(os.Stderr, "%s", assets.CLIIcon())

	if opts.Delay > 0 {
		if err := countdown(ctx, opts.Delay, os.Stderr); err != nil {
			fmt.Fprintln(os.Stderr, "\nCancelled.")
			return 2
		}
	}

	fmt.Fprint(os.Stderr, "\rScanning...")

	results, err := scanner.ScanAllScreens(ctx)
	fmt.Fprintln(os.Stderr)

	if err != nil {
		if errors.Is(err, context.Canceled) {
			fmt.Fprintln(os.Stderr, "Cancelled.")
			return 2
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 2
	}

	slog.Debug("cli scan complete", "results", len(results))

	if len(results) == 0 {
		fmt.Fprintln(os.Stderr, "No QR codes found.")
		return 1
	}

	for _, r := range results {
		fmt.Println(r)
	}
	return 0
}

// countdown prints a "Scanning in N..." line to w, updating in-place each second.
// Returns ctx.Err() if the context is cancelled before the countdown finishes.
func countdown(ctx context.Context, seconds int, w io.Writer) error {
	for i := seconds; i >= 1; i-- {
		fmt.Fprintf(w, "\rScanning in %d...", i)
		select {
		case <-ctx.Done():
			fmt.Fprintln(w)
			return ctx.Err()
		case <-time.After(time.Second):
		}
	}
	return nil
}
