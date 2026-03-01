package assets

import (
	"bytes"
	_ "embed"
	"image"
	"image/color"
	"image/png"
	"sync"
)

//go:embed menubar_22.png
var menubarIcon []byte

//go:embed menubar_44.png
var menubarIcon2x []byte

//go:embed icon_64.png
var notificationIcon []byte

// Icon returns the 22×22 menu bar template icon PNG (1x).
func Icon() []byte { return menubarIcon }

// Icon2x returns the 44×44 menu bar template icon PNG (2x / Retina).
func Icon2x() []byte { return menubarIcon2x }

// NotificationIcon returns the 64×64 app icon PNG for use in desktop notifications.
func NotificationIcon() []byte { return notificationIcon }

var (
	lightOnce   sync.Once
	lightIcon   []byte
	light2xOnce sync.Once
	light2xIcon []byte
)

// IconLight returns a white-on-transparent version of the 22×22 icon,
// suitable for display on dark menu bar backgrounds (e.g. Ubuntu).
func IconLight() []byte {
	lightOnce.Do(func() { lightIcon = invertToWhite(menubarIcon) })
	return lightIcon
}

// Icon2xLight returns a white-on-transparent version of the 44×44 icon.
func Icon2xLight() []byte {
	light2xOnce.Do(func() { light2xIcon = invertToWhite(menubarIcon2x) })
	return light2xIcon
}

// invertToWhite converts a black-on-transparent PNG to white-on-transparent
// by setting all non-transparent pixels to white while preserving alpha.
func invertToWhite(src []byte) []byte {
	img, err := png.Decode(bytes.NewReader(src))
	if err != nil {
		return src
	}
	bounds := img.Bounds()
	out := image.NewNRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, a := img.At(x, y).RGBA()
			if a > 0 {
				out.Set(x, y, color.NRGBA{R: 255, G: 255, B: 255, A: uint8(a >> 8)})
			}
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, out); err != nil {
		return src
	}
	return buf.Bytes()
}
