package tray

import (
	"fmt"
	"strings"
	"sync"

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
	scanItem     *systray.MenuItem
	statusItem   *systray.MenuItem
	resultItems  [maxResults]*systray.MenuItem
	resultTitles [maxResults]string // full (untruncated) text for each result slot
	resultMu     sync.Mutex
	scanning     bool
)

// OnReady is called by systray once the tray icon is ready. Runs in a goroutine.
func OnReady() {
	systray.SetTemplateIcon(assets.Icon(), assets.Icon())
	systray.SetTooltip("Squrl — click to scan")

	scanItem = systray.AddMenuItem("Scan Screen", "Capture all displays and decode QR codes")
	systray.AddSeparator()
	statusItem = systray.AddMenuItem("No results yet", "")
	statusItem.Disable()
	systray.AddSeparator()

	// Pre-allocate result item pool (hidden by default).
	for i := range resultItems {
		item := systray.AddMenuItem("", "")
		item.Hide()
		resultItems[i] = item
	}

	systray.AddSeparator()
	quitItem := systray.AddMenuItem("Quit", "Exit Squrl")

	// Event loop.
	go func() {
		for {
			select {
			case <-scanItem.ClickedCh:
				go runScan()
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
					notify.ShowNotification("Copied to clipboard", title)
				}
			}
		}()
	}
}

// OnExit is called by systray when the app is quitting.
func OnExit() {}

// runScan performs a screen capture and QR decode, then updates the menu.
func runScan() {
	resultMu.Lock()
	if scanning {
		resultMu.Unlock()
		return
	}
	scanning = true
	resultMu.Unlock()

	scanItem.Disable()
	statusItem.SetTitle("Scanning…")

	results, err := scanner.ScanAllScreens()

	resultMu.Lock()
	scanning = false
	resultMu.Unlock()

	scanItem.Enable()

	if err != nil {
		statusItem.SetTitle(fmt.Sprintf("Error: %v", err))
		notify.ShowNotification("Scan failed", err.Error())
		return
	}

	updateResults(results)
}

// updateResults rebuilds the result pool with the latest scan results.
func updateResults(results []string) {
	resultMu.Lock()
	defer resultMu.Unlock()

	// Hide all slots first.
	for i, item := range resultItems {
		item.Hide()
		item.SetTitle("")
		resultTitles[i] = ""
	}

	if len(results) == 0 {
		statusItem.SetTitle("No QR codes found")
		notify.ShowNotification("Squrl", "No QR codes found on screen")
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

	summary := fmt.Sprintf("Found %d QR code(s)", len(results))
	detail := strings.Join(results[:count], "\n")
	notify.ShowNotification(summary, detail)
}

// copyToClipboard writes text to the system clipboard.
func copyToClipboard(text string) {
	_ = clipboard.WriteAll(text)
}

// truncate shortens s to at most n runes, appending "…" if truncated.
func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n-1]) + "…"
}
