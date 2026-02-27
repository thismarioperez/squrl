package assets

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
)

// Icon returns a 22×22 PNG representing a simple QR-code symbol.
// It is used as the macOS menu bar template icon (SetTemplateIcon).
// Template icons should be black-on-transparent; macOS applies tinting automatically.
func Icon() []byte {
	const size = 22
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	black := color.RGBA{0, 0, 0, 255}

	// Helper to fill a square region.
	fill := func(x, y, w, h int) {
		for dy := 0; dy < h; dy++ {
			for dx := 0; dx < w; dx++ {
				img.Set(x+dx, y+dy, black)
			}
		}
	}

	// Outer border of top-left finder pattern (7×7 → scaled to fit 22px canvas).
	// We draw three finder patterns (corners) and a few data modules to look QR-like.

	// Top-left finder pattern at (1,1), 7×7
	fill(1, 1, 7, 7)
	// hollow out inner 5×5
	for dy := 0; dy < 5; dy++ {
		for dx := 0; dx < 5; dx++ {
			img.Set(2+dx, 2+dy, color.RGBA{0, 0, 0, 0})
		}
	}
	// inner 3×3 dot
	fill(3, 3, 3, 3)

	// Top-right finder pattern at (14,1), 7×7
	fill(14, 1, 7, 7)
	for dy := 0; dy < 5; dy++ {
		for dx := 0; dx < 5; dx++ {
			img.Set(15+dx, 2+dy, color.RGBA{0, 0, 0, 0})
		}
	}
	fill(16, 3, 3, 3)

	// Bottom-left finder pattern at (1,14), 7×7
	fill(1, 14, 7, 7)
	for dy := 0; dy < 5; dy++ {
		for dx := 0; dx < 5; dx++ {
			img.Set(2+dx, 15+dy, color.RGBA{0, 0, 0, 0})
		}
	}
	fill(3, 16, 3, 3)

	// A few scattered data modules in the lower-right area for realism.
	fill(10, 10, 2, 2)
	fill(14, 10, 2, 2)
	fill(10, 14, 2, 2)
	fill(18, 14, 2, 2)
	fill(14, 18, 2, 2)
	fill(18, 18, 2, 2)
	fill(10, 18, 2, 2)

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		panic("assets: failed to encode icon PNG: " + err.Error())
	}
	return buf.Bytes()
}
