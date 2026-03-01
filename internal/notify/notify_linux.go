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
			iconPath, cleanup := writeIconTemp()
			defer cleanup()

			args := []string{"--wait", "--action", "default=Open"}
			if iconPath != "" {
				args = append(args, "--icon", iconPath)
			}
			args = append(args, title, message)
			cmd := exec.Command("notify-send", args...)
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

	iconPath, cleanup := writeIconTemp()
	defer cleanup()

	args := []string{title, message}
	if iconPath != "" {
		args = append([]string{"--icon", iconPath}, args...)
	}
	_ = exec.Command("notify-send", args...).Run()
}
