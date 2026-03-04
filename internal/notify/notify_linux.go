//go:build linux

package notify

import (
	"bufio"
	"fmt"
	"os/exec"
)

// ShowNotification displays a desktop notification via notify-send.
// Requires libnotify-bin (or equivalent) to be installed.
// When onActivate is non-nil, the callback is invoked if the user clicks the
// notification body (requires notify-send with --wait/--action support).
func ShowNotification(n Notification) {
	dur := n.Duration
	if dur == 0 {
		dur = DefaultDuration
	}
	// dur < 0 means indefinite; notify-send --expire-time 0 lets the server decide (no auto-dismiss)
	var expireMs int64
	if dur > 0 {
		expireMs = dur.Milliseconds()
	}
	expireArg := fmt.Sprintf("%d", expireMs)

	if n.OnActivate != nil {
		go func() {
			iconPath, cleanup := writeIconTemp()
			defer cleanup()

			args := []string{"--wait", "--action", "default=Open", "--expire-time", expireArg}
			if iconPath != "" {
				args = append(args, "--icon", iconPath)
			}
			args = append(args, n.Title, n.Message)
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
					n.OnActivate()
					break
				}
			}
			_ = cmd.Wait()
		}()
		return
	}

	iconPath, cleanup := writeIconTemp()
	defer cleanup()

	args := []string{"--expire-time", expireArg}
	if iconPath != "" {
		args = append(args, "--icon", iconPath)
	}
	args = append(args, n.Title, n.Message)
	_ = exec.Command("notify-send", args...).Run()
}
