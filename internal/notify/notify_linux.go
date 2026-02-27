//go:build linux

package notify

import "os/exec"

// ShowNotification displays a desktop notification via notify-send.
// Requires libnotify-bin (or equivalent) to be installed.
func ShowNotification(title, message string) {
	_ = exec.Command("notify-send", title, message).Run()
}
