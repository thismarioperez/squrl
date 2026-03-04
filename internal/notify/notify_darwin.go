//go:build darwin

package notify

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
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
				return
			}
		}

		// Fall back to PATH (local development with alerter installed via Homebrew).
		if p, err := exec.LookPath("alerter"); err == nil {
			alerterPath = p
		}
	})
	return alerterPath
}

// ShowNotification displays a macOS notification. Uses alerter when available so
// the callback is invoked if the user clicks the notification body.
// Falls back to osascript (no click detection) when alerter is unavailable.
func ShowNotification(n Notification) {
	if len(n.Message) > 200 {
		n.Message = n.Message[:197] + "..."
	}

	if p := lookupAlerter(); p != "" {
		go func() {
			iconPath, cleanup := writeIconTemp()
			defer cleanup()

			args := []string{
				"--title", n.Title,
				"--message", n.Message,
				"--group", "com.mario.squrl",
			}
			if iconPath != "" {
				args = append(args, "--app-icon", iconPath)
			}
			out, err := exec.Command(p, args...).Output()
			if err == nil && strings.TrimSpace(string(out)) == "@CONTENTCLICKED" && n.OnActivate != nil {
				n.OnActivate()
			}
		}()
		return
	}

	// Fallback: osascript (no click detection).
	safeTitle := strings.ReplaceAll(n.Title, `"`, `'`)
	safeMsg := strings.ReplaceAll(n.Message, `"`, `'`)
	script := fmt.Sprintf(`display notification %q with title %q`, safeMsg, safeTitle)
	_ = exec.Command("osascript", "-e", script).Run()
}
