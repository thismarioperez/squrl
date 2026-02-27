//go:build darwin

package notify

import (
	"fmt"
	"os/exec"
	"strings"
)

// ShowNotification displays a macOS notification via osascript.
func ShowNotification(title, message string) {
	// Escape double-quotes to prevent osascript injection.
	safeTitle := strings.ReplaceAll(title, `"`, `'`)
	safeMsg := strings.ReplaceAll(message, `"`, `'`)
	// Truncate long messages for the notification banner.
	if len(safeMsg) > 200 {
		safeMsg = safeMsg[:197] + "..."
	}
	script := fmt.Sprintf(`display notification %q with title %q`, safeMsg, safeTitle)
	_ = exec.Command("osascript", "-e", script).Run()
}
