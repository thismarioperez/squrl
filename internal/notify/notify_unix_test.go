//go:build linux || darwin

package notify

import (
	"os"
	"testing"

	"github.com/thismarioperez/squrl/assets"
)

func TestWriteIconTemp_CreatesFile(t *testing.T) {
	path, cleanup := writeIconTemp()
	defer cleanup()

	if path == "" {
		t.Fatal("expected a non-empty path")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("icon temp file does not exist: %v", err)
	}
}

func TestWriteIconTemp_ContentMatchesAsset(t *testing.T) {
	path, cleanup := writeIconTemp()
	defer cleanup()

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading temp file: %v", err)
	}
	want := assets.NotificationIcon()
	if len(got) != len(want) {
		t.Fatalf("content length mismatch: got %d bytes, want %d bytes", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("content differs at byte %d", i)
		}
	}
}

func TestWriteIconTemp_CleanupRemovesFile(t *testing.T) {
	path, cleanup := writeIconTemp()
	cleanup()

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected file to be removed after cleanup, got: %v", err)
	}
}
