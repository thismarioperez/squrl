// Package logging configures the application-wide structured logger.
//
// Call Init() once from main(). When SQURL_DEBUG=1, output is written to a
// platform-appropriate log file and stderr at DEBUG level. Otherwise all
// output is silently discarded (zero overhead).
//
// Instrumented packages use the stdlib log/slog package-level functions
// (slog.Debug, slog.Info, slog.Error, etc.) directly — no import of this
// package is needed beyond the initial Init() call in main.
package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
)

// Init configures the default slog logger from the SQURL_DEBUG environment
// variable. Always call from main() before any other code runs.
// The returned cleanup function closes the log file; always safe to defer.
func Init() func() {
	// Default: discard all output so there is zero overhead at runtime.
	slog.SetDefault(slog.New(slog.DiscardHandler))

	if os.Getenv("SQURL_DEBUG") != "1" {
		return func() {}
	}

	logPath, err := logFilePath()
	if err != nil {
		slog.SetDefault(stderrLogger())
		return func() {}
	}

	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		slog.SetDefault(stderrLogger())
		return func() {}
	}

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		slog.SetDefault(stderrLogger())
		return func() {}
	}

	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	slog.SetDefault(slog.New(slog.NewTextHandler(io.MultiWriter(f, os.Stderr), opts)))
	slog.Info("debug logging started", "log_file", logPath)

	return func() { _ = f.Close() }
}

func stderrLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func logFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Logs", "squrl", "squrl.log"), nil
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", fmt.Errorf("APPDATA not set")
		}
		return filepath.Join(appData, "squrl", "squrl.log"), nil
	default: // linux and others
		return filepath.Join(home, ".local", "share", "squrl", "squrl.log"), nil
	}
}
