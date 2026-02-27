package scanner

import (
	"fmt"
	"image"

	"github.com/kbinani/screenshot"
	"github.com/makiuchi-d/gozxing"
	multiqrcode "github.com/makiuchi-d/gozxing/multi/qrcode"
	"github.com/makiuchi-d/gozxing/qrcode"
)

// ScanAllScreens captures every active display and returns the decoded text from
// all QR codes found. Duplicate results across displays are deduplicated.
func ScanAllScreens() ([]string, error) {
	n := screenshot.NumActiveDisplays()
	if n == 0 {
		return nil, fmt.Errorf("no active displays found — ensure screen capture permission is granted")
	}

	seen := make(map[string]bool)
	var results []string
	captured := 0

	for i := 0; i < n; i++ {
		img, err := screenshot.CaptureDisplay(i)
		if err != nil {
			continue
		}
		captured++

		codes, err := decodeQRCodes(img)
		if err != nil {
			// gozxing returns an error when nothing is found — not fatal.
			continue
		}

		for _, text := range codes {
			if !seen[text] {
				seen[text] = true
				results = append(results, text)
			}
		}
	}

	if captured == 0 {
		return nil, fmt.Errorf("could not capture any displays — ensure screen capture permission is granted")
	}

	return results, nil
}

// decodeQRCodes attempts to find all QR codes in img using gozxing's multi-reader.
// Falls back to the single QR reader if the multi-reader returns nothing.
func decodeQRCodes(img image.Image) ([]string, error) {
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return nil, fmt.Errorf("bitmap conversion: %w", err)
	}

	hints := map[gozxing.DecodeHintType]any{
		gozxing.DecodeHintType_TRY_HARDER: true,
	}

	// Try multi-reader first (finds multiple codes in one image).
	multiReader := multiqrcode.NewQRCodeMultiReader()
	multiResults, err := multiReader.DecodeMultiple(bmp, hints)
	if err == nil && len(multiResults) > 0 {
		var texts []string
		for _, r := range multiResults {
			if t := r.GetText(); t != "" {
				texts = append(texts, t)
			}
		}
		if len(texts) > 0 {
			return texts, nil
		}
	}

	// Fall back to single QR reader — reset bitmap for fresh decode.
	bmp, err = gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return nil, err
	}
	result, err := qrcode.NewQRCodeReader().Decode(bmp, hints)
	if err != nil {
		return nil, err
	}
	return []string{result.GetText()}, nil
}
