//go:build windows

package notify

import "context"

// ShowNotification is a stub on Windows.
// TODO: implement Windows toast notifications (e.g., via go-toast or PowerShell).
func ShowNotification(_ context.Context, _ Notification) {}
