//go:build linux || darwin

package notify

import (
	"os"

	"github.com/thismarioperez/squrl/assets"
)

// writeIconTemp writes the embedded notification icon to a temp file and
// returns its path along with a cleanup function. If writing fails, an empty
// path and a no-op cleanup are returned.
func writeIconTemp() (string, func()) {
	f, err := os.CreateTemp("", "squrl-icon-*.png")
	if err != nil {
		return "", func() {}
	}
	if _, err := f.Write(assets.NotificationIcon()); err != nil {
		f.Close()
		os.Remove(f.Name())
		return "", func() {}
	}
	f.Close()
	return f.Name(), func() { os.Remove(f.Name()) }
}
