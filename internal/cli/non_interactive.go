package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/thismarioperez/squrl/internal/scanner"
)

// RunNonInteractive runs a scan without the TUI. It waits for opts.Delay seconds
// (silently), scans all screens, and prints each decoded QR string to stdout one
// per line. Returns 0 (found), 1 (none found), or 2 (error/cancelled).
func RunNonInteractive(ctx context.Context, opts ScanOptions) int {
	if opts.Delay > 0 {
		select {
		case <-time.After(time.Duration(opts.Delay) * time.Second):
		case <-ctx.Done():
			fmt.Fprintf(os.Stderr, "squrl: %v\n", ctx.Err())
			return 2
		}
	}

	results, err := scanner.ScanAllScreens(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "squrl: %v\n", err)
		return 2
	}

	for _, r := range results {
		fmt.Println(r)
	}

	if len(results) == 0 {
		return 1
	}
	return 0
}
