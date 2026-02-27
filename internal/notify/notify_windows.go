//go:build windows

package notify

// ShowNotification is a stub on Windows.
// TODO: implement Windows toast notifications (e.g., via go-toast or PowerShell).
func ShowNotification(_, _ string) {}
