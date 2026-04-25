package cli

import (
	"flag"
	"fmt"
)

// ScanOptions holds configuration for a CLI scan run.
type ScanOptions struct {
	Delay          int    // countdown seconds before scan; 0 = skip countdown
	NonInteractive bool   // skip TUI; print results to stdout one per line
	Version        string // build-time version string, e.g. "v1.2.3" or "dev"
}

// ParseScanArgs parses the arguments after "scan" into ScanOptions.
func ParseScanArgs(args []string, version string) (ScanOptions, error) {
	fs := flag.NewFlagSet("squrl", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage of squrl:\n")
		fs.VisitAll(func(f *flag.Flag) {
			typeName, usage := flag.UnquoteUsage(f)
			prefix := "--"
			if len(f.Name) == 1 {
				prefix = "-"
			}
			fmt.Fprintf(fs.Output(), "  %s%s %s\n\t%s (default %v)\n", prefix, f.Name, typeName, usage, f.DefValue)
		})
	}
	delay := fs.Int("delay", 3, "seconds to wait before scanning (0 to skip countdown)")
	fs.IntVar(delay, "D", 3, "shorthand for --delay")
	nonInteractive := fs.Bool("non-interactive", false, "non-interactive mode: print results to stdout, one per line")
	fs.BoolVar(nonInteractive, "n", false, "shorthand for --non-interactive")
	if err := fs.Parse(args); err != nil {
		return ScanOptions{}, err
	}
	if fs.NArg() > 0 {
		return ScanOptions{}, fmt.Errorf("unexpected argument: %s", fs.Arg(0))
	}
	if *delay < 0 {
		return ScanOptions{}, fmt.Errorf("--delay must be >= 0")
	}
	return ScanOptions{Delay: *delay, NonInteractive: *nonInteractive, Version: version}, nil
}
