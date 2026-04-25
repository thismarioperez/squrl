package cli

import (
	"context"
	"testing"
)

func TestRunNonInteractive_CtxCancelledDuringDelay(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	opts := ScanOptions{Delay: 10}
	code := RunNonInteractive(ctx, opts)
	if code != 2 {
		t.Errorf("expected exit code 2, got %d", code)
	}
}

func TestRunNonInteractive_CtxCancelledNoDelay(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	opts := ScanOptions{Delay: 0}
	code := RunNonInteractive(ctx, opts)
	if code != 2 {
		t.Errorf("expected exit code 2, got %d", code)
	}
}
