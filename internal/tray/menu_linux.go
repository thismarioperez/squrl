//go:build linux

package tray

// openMenu is a no-op on Linux — there is no standard way to programmatically
// click a system tray icon across all desktop environments.
func openMenu() {}
