package scanner

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"

	qrgen "github.com/skip2/go-qrcode"
)

// qrImage encodes text into a QR code and returns it as an image.Image.
func qrImage(t *testing.T, text string) image.Image {
	t.Helper()
	pngBytes, err := qrgen.Encode(text, qrgen.Medium, 256)
	if err != nil {
		t.Fatalf("qrgen.Encode(%q): %v", text, err)
	}
	img, err := png.Decode(bytes.NewReader(pngBytes))
	if err != nil {
		t.Fatalf("png.Decode: %v", err)
	}
	return img
}

func TestDecodeQRCodes_SingleCode(t *testing.T) {
	want := "https://example.com"
	img := qrImage(t, want)

	got, err := decodeQRCodes(img)
	if err != nil {
		t.Fatalf("decodeQRCodes returned error: %v", err)
	}
	if len(got) != 1 || got[0] != want {
		t.Errorf("got %v; want [%q]", got, want)
	}
}

func TestDecodeQRCodes_PlainText(t *testing.T) {
	want := "hello world"
	img := qrImage(t, want)

	got, err := decodeQRCodes(img)
	if err != nil {
		t.Fatalf("decodeQRCodes returned error: %v", err)
	}
	if len(got) == 0 || got[0] != want {
		t.Errorf("got %v; want [%q]", got, want)
	}
}

func TestDecodeQRCodes_NoCode(t *testing.T) {
	// Solid white image — no QR code present.
	img := image.NewRGBA(image.Rect(0, 0, 256, 256))
	for y := 0; y < 256; y++ {
		for x := 0; x < 256; x++ {
			img.Set(x, y, color.White)
		}
	}

	got, err := decodeQRCodes(img)
	if err == nil {
		t.Errorf("expected an error for blank image, got results: %v", got)
	}
}