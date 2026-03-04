//go:build darwin

package notify

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	alerterPath     string
	alerterPathOnce sync.Once
)

func lookupAlerter() string {
	alerterPathOnce.Do(func() {
		// When running as a bundled .app, alerter lives next to the main binary
		// in Contents/MacOS/. Check there first.
		if exe, err := os.Executable(); err == nil {
			candidate := filepath.Join(filepath.Dir(exe), "alerter")
			if _, err := os.Stat(candidate); err == nil {
				alerterPath = candidate
				slog.Debug("alerter found next to binary", "path", alerterPath)
				return
			}
		}

		// Fall back to PATH (local development with alerter installed via Homebrew).
		if p, err := exec.LookPath("alerter"); err == nil {
			alerterPath = p
			slog.Debug("alerter found in PATH", "path", alerterPath)
		}
	})
	return alerterPath
}

// ShowNotification displays a macOS notification. Uses alerter when available so
// the callback is invoked if the user clicks the notification body.
// Falls back to osascript (no click detection) when alerter is unavailable.
func ShowNotification(ctx context.Context, n Notification) {
	if len(n.Message) > 200 {
		n.Message = n.Message[:197] + "..."
	}

	dur := n.Duration
	if dur == 0 {
		dur = DefaultDuration
	}

	osascript := func() {
		safeTitle := strings.ReplaceAll(n.Title, `"`, `'`)
		safeMsg := strings.ReplaceAll(n.Message, `"`, `'`)
		script := fmt.Sprintf(`display notification %q with title %q`, safeMsg, safeTitle)
		if err := exec.CommandContext(ctx, "osascript", "-e", script).Run(); err != nil {
			if ctx.Err() == nil {
				slog.Error("osascript notification failed", "err", err)
			}
		}
	}

	if p := lookupAlerter(); p != "" {
		go func() {
			iconPath, cleanup := writeIconTemp()
			defer cleanup()

			baseArgs := []string{
				"--title", n.Title,
				"--message", n.Message,
				"--group", "com.mario.squrl",
				"--sender", "com.apple.Terminal",
			}
			if dur > 0 {
				baseArgs = append(baseArgs, "--timeout", fmt.Sprintf("%d", int(dur/time.Second)))
			}
			// dur < 0 means indefinite: omit --timeout so alerter waits until dismissed.

			args := baseArgs
			if iconPath != "" {
				args = append(append([]string{}, baseArgs...), "--app-icon", iconPath)
			}
			out, err := exec.CommandContext(ctx, p, args...).Output()
			if err != nil && iconPath != "" {
				// --app-icon uses a private API that can fail on some macOS versions.
				// Retry without it before giving up.
				slog.Debug("alerter failed with --app-icon, retrying without it", "err", err)
				out, err = exec.CommandContext(ctx, p, baseArgs...).Output()
			}
			if err != nil {
				if ctx.Err() == nil {
					slog.Warn("alerter failed, falling back to osascript", "err", err)
					osascript()
				}
				return
			}
			if strings.TrimSpace(string(out)) == "@CONTENTCLICKED" && n.OnActivate != nil {
				n.OnActivate()
			}
		}()
		return
	}

	// Fallback: osascript (no click detection) when alerter is not available.
	slog.Debug("alerter not available, falling back to osascript")
	osascript()
}
