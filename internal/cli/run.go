package cli

import (
	"context"
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
)

// RunTUI starts the Bubble Tea program and blocks until the user quits.
// Returns 0 if the last scan found QR codes, 1 if none found, 2 on error or cancellation.
func RunTUI(ctx context.Context, opts ScanOptions, version string) int {
	m := initialModel(ctx, opts, version)
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "squrl: TUI error: %v\n", err)
		return 2
	}
	if fm, ok := finalModel.(model); ok {
		return fm.exitCode
	}
	return 0
}
