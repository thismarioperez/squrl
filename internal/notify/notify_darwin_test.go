//go:build darwin

package notify

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestLookupAppIconFrom_PackagedBuild simulates Contents/MacOS/squrl finding
// the icon at Contents/Resources/AppIcon.icns.
func TestLookupAppIconFrom_PackagedBuild(t *testing.T) {
	tmp := t.TempDir()
	exeDir := filepath.Join(tmp, "Contents", "MacOS")
	want := filepath.Join(tmp, "Contents", "Resources", "AppIcon.icns")
	writeFile(t, want)

	got := lookupAppIconFrom(exeDir, "")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// TestLookupAppIconFrom_BinBuild simulates bin/squrl finding the icon at
// assets/AppIcon.icns (no Resources/ sibling present).
func TestLookupAppIconFrom_BinBuild(t *testing.T) {
	tmp := t.TempDir()
	exeDir := filepath.Join(tmp, "bin")
	want := filepath.Join(tmp, "assets", "AppIcon.icns")
	writeFile(t, want)

	got := lookupAppIconFrom(exeDir, "")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// TestLookupAppIconFrom_CWDFallback simulates go run where the executable
// lives in a temp dir but the icon exists under the CWD (project root).
func TestLookupAppIconFrom_CWDFallback(t *testing.T) {
	exeDir := t.TempDir() // no icon near here
	cwd := t.TempDir()
	want := filepath.Join(cwd, "assets", "AppIcon.icns")
	writeFile(t, want)

	got := lookupAppIconFrom(exeDir, cwd)
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// TestLookupAppIconFrom_NotFound returns an empty string when no icon exists.
func TestLookupAppIconFrom_NotFound(t *testing.T) {
	got := lookupAppIconFrom(t.TempDir(), "")
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

// TestLookupAppIconFrom_PackagedBuildTakesPrecedence verifies the packaged
// Resources/ path wins when both Resources/ and assets/ exist.
func TestLookupAppIconFrom_PackagedBuildTakesPrecedence(t *testing.T) {
	tmp := t.TempDir()
	exeDir := filepath.Join(tmp, "Contents", "MacOS")
	want := filepath.Join(tmp, "Contents", "Resources", "AppIcon.icns")
	writeFile(t, want)
	writeFile(t, filepath.Join(tmp, "Contents", "assets", "AppIcon.icns"))

	got := lookupAppIconFrom(exeDir, "")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
