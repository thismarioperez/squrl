//go:build linux

package notify

import (
	"bufio"
	"os/exec"
)

// ShowNotification displays a desktop notification via notify-send.
// Requires libnotify-bin (or equivalent) to be installed.
// When onActivate is non-nil, the callback is invoked if the user clicks the
// notification body (requires notify-send with --wait/--action support).
func ShowNotification(title, message string, onActivate func()) {
	if onActivate != nil {
		go func() {
			cmd := exec.Command("notify-send", "--wait", "--action", "default=Open", title, message)
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				return
			}
			if err := cmd.Start(); err != nil {
				return
			}
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				if scanner.Text() == "default" {
					onActivate()
					break
				}
			}
			_ = cmd.Wait()
		}()
		return
	}

	_ = exec.Command("notify-send", title, message).Run()
}
