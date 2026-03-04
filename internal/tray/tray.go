package tray

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/atotto/clipboard"
	"github.com/getlantern/systray"
	"github.com/thismarioperez/squrl/assets"
	"github.com/thismarioperez/squrl/internal/notify"
	"github.com/thismarioperez/squrl/internal/scanner"
)

// resultItems holds the current dynamically-added result menu items so they can be
// hidden/removed before adding new ones. systray does not support removing items, so
// we pre-allocate a fixed pool and hide unused ones.
const maxResults = 20

var (
	appVersion   = "dev"
	appCtx       context.Context
	appCancel    context.CancelFunc
	scanItem     *systray.MenuItem
	clearItem    *systray.MenuItem
	statusItem   *systray.MenuItem
	resultItems  [maxResults]*systray.MenuItem
	resultTitles [maxResults]string // full (untruncated) text for each result slot
	resultMu     sync.Mutex
	scanning     bool
)

// SetVersion stores the application version to be displayed in the tray menu.
func SetVersion(v string) { appVersion = v }

// setTrayIcon sets the tray icon, accounting for the menubar background color on Linux.
// On macOS, SetTemplateIcon lets the OS automatically adapt the icon for light/dark mode.
// On Linux, template icon rendering is not supported, so we detect the desktop color
// scheme and supply a pre-colorized icon directly.
func setTrayIcon() {
	if runtime.GOOS == "linux" {
		if isLinuxDarkMode() {
			systray.SetIcon(assets.IconLight())
		} else {
			systray.SetIcon(assets.Icon())
		}
	} else {
		systray.SetTemplateIcon(assets.Icon(), assets.Icon())
	}
}

// isLinuxDarkMode returns true when the menubar background is likely dark.
func isLinuxDarkMode() bool {
	out, err := exec.Command("gsettings", "get", "org.gnome.desktop.interface", "color-scheme").Output()
	return parseDarkMode(out, err)
}

// parseDarkMode interprets gsettings output to determine if dark mode is active.
// Returns true (dark) when err is non-nil or the output does not contain "prefer-light".
func parseDarkMode(out []byte, err error) bool {
	if err != nil {
		// gsettings unavailable; assume dark (most Linux panels are dark by default).
		return true
	}
	return !strings.Contains(strings.ToLower(string(out)), "prefer-light")
}

// OnReady is called by systray once the tray icon is ready. Runs in a goroutine.
func OnReady() {
	appCtx, appCancel = context.WithCancel(context.Background())

	initPlatform()

	if !hasScreenCapturePermission() {
		slog.Warn("screen capture permission not granted")
		requestScreenCapturePermission()
		notify.ShowNotification(appCtx, notify.Notification{
			Title:   "Screen Recording permission required",
			Message: "Grant access in System Settings → Privacy & Security → Screen Recording, then relaunch Squrl.",
		})
	}

	setTrayIcon()
	systray.SetTooltip("Squrl — click to scan")

	scanItem = systray.AddMenuItem("Scan Screen", "Capture all displays and decode QR codes")
	systray.AddSeparator()
	clearItem = systray.AddMenuItem("Clear results", "")
	clearItem.Hide()
	statusItem = systray.AddMenuItem("No results yet", "")
	statusItem.Disable()

	// Pre-allocate result item pool (hidden by default).
	for i := range resultItems {
		item := systray.AddMenuItem("", "")
		item.Hide()
		resultItems[i] = item
	}

	systray.AddSeparator()
	versionItem := systray.AddMenuItem(fmt.Sprintf("Version %s", appVersion), "")
	versionItem.Disable()
	quitItem := systray.AddMenuItem("Quit", "Exit Squrl")

	// Event loop.
	go func() {
		for {
			select {
			case <-scanItem.ClickedCh:
				go runScan()
			case <-clearItem.ClickedCh:
				clearResults()
			case <-quitItem.ClickedCh:
				systray.Quit()
			}
		}
	}()

	// Listen for clicks on result items in separate goroutines.
	for i := range resultItems {
		i := i
		go func() {
			for range resultItems[i].ClickedCh {
				resultMu.Lock()
				title := resultTitles[i]
				resultMu.Unlock()
				if title != "" {
					copyToClipboard(title)
					notify.ShowNotification(appCtx, notify.Notification{Title: "Copied to clipboard", Message: title, OnActivate: openMenu})
				}
			}
		}()
	}
}

// OnExit is called by systray when the app is quitting.
func OnExit() {
	if appCancel != nil {
		appCancel()
	}
}

// runScan performs a screen capture and QR decode, then updates the menu.
func runScan() {
	resultMu.Lock()
	if scanning {
		resultMu.Unlock()
		return
	}
	scanning = true
	resultMu.Unlock()

	slog.Debug("scan started")
	scanItem.Disable()
	statusItem.SetTitle("Scanning…")

	scanCtx, cancel := context.WithTimeout(appCtx, 30*time.Second)
	defer cancel()

	results, err := scanner.ScanAllScreens(scanCtx)

	resultMu.Lock()
	scanning = false
	resultMu.Unlock()

	scanItem.Enable()

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			slog.Warn("scan timed out")
			statusItem.SetTitle("Scan timed out")
			notify.ShowNotification(appCtx, notify.Notification{Title: "Scan timed out", Message: "QR scan exceeded 30 seconds and was cancelled.", OnActivate: openMenu})
		} else if errors.Is(err, context.Canceled) {
			slog.Debug("scan cancelled (app quitting)")
		} else {
			slog.Error("scan failed", "err", err)
			statusItem.SetTitle(fmt.Sprintf("Error: %v", err))
			notify.ShowNotification(appCtx, notify.Notification{Title: "Scan failed", Message: err.Error(), OnActivate: openMenu})
		}
		return
	}

	slog.Debug("scan complete", "results", len(results))
	updateResults(appCtx, results)
}

// updateResults rebuilds the result pool with the latest scan results.
func updateResults(ctx context.Context, results []string) {
	resultMu.Lock()
	defer resultMu.Unlock()

	// Hide all slots first. Do NOT call SetTitle here: on Linux the systray
	// backend's do_add_or_update_menu_item always calls gtk_widget_show(), so
	// a SetTitle call after Hide would immediately un-hide the item.
	for i, item := range resultItems {
		item.Hide()
		resultTitles[i] = ""
	}

	if len(results) == 0 {
		clearItem.Hide()
		statusItem.SetTitle("No QR codes found")
		notify.ShowNotification(ctx, notify.Notification{Title: "Squrl", Message: "No QR codes found on screen", OnActivate: openMenu})
		return
	}

	count := len(results)
	if count > maxResults {
		count = maxResults
	}

	statusItem.SetTitle(fmt.Sprintf("Found %d QR code(s) — click to copy:", count))

	for i := 0; i < count; i++ {
		resultTitles[i] = results[i]
		resultItems[i].SetTitle(truncate(results[i], 60))
		resultItems[i].Show()
	}

	clearItem.Show()

	summary := fmt.Sprintf("Found %d QR code(s)", len(results))
	detail := strings.Join(results[:count], "\n")
	notify.ShowNotification(ctx, notify.Notification{Title: summary, Message: detail, OnActivate: openMenu})
}

// clearResults hides all result slots and resets status.
func clearResults() {
	resultMu.Lock()
	defer resultMu.Unlock()

	// Do NOT call SetTitle here: on Linux the systray backend's
	// do_add_or_update_menu_item always calls gtk_widget_show(), so a SetTitle
	// call queued after Hide would immediately un-hide the item, leaving
	// phantom blank entries in the menu.
	for i, item := range resultItems {
		item.Hide()
		resultTitles[i] = ""
	}
	clearItem.Hide()
	statusItem.SetTitle("No results yet")
}

// copyToClipboard writes text to the system clipboard.
func copyToClipboard(text string) {
	if err := clipboard.WriteAll(text); err != nil {
		slog.Error("clipboard write failed", "err", err)
	}
}

// truncate shortens s to at most n runes, appending "…" if truncated.
func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n-1]) + "…"
}
