//go:build windows

package tray

func openMenu() {}

func initPlatform() {}

func hasScreenCapturePermission() bool  { return true }
func requestScreenCapturePermission()   {}
