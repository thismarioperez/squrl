package assets

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

// makePNG encodes a single-pixel NRGBA image as PNG bytes for use in tests.
func makePNG(c color.NRGBA) []byte {
	img := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	img.SetNRGBA(0, 0, c)
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func TestInvertToWhite_OpaquePixelBecomesWhite(t *testing.T) {
	src := makePNG(color.NRGBA{R: 0, G: 0, B: 0, A: 255})
	got := invertToWhite(src)

	img, err := png.Decode(bytes.NewReader(got))
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	r, g, b, a := img.At(0, 0).RGBA()
	if r>>8 != 255 || g>>8 != 255 || b>>8 != 255 {
		t.Errorf("expected white pixel, got r=%d g=%d b=%d a=%d", r>>8, g>>8, b>>8, a>>8)
	}
	if a>>8 != 255 {
		t.Errorf("expected alpha=255, got %d", a>>8)
	}
}

func TestInvertToWhite_TransparentPixelStaysTransparent(t *testing.T) {
	src := makePNG(color.NRGBA{R: 0, G: 0, B: 0, A: 0})
	got := invertToWhite(src)

	img, err := png.Decode(bytes.NewReader(got))
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	_, _, _, a := img.At(0, 0).RGBA()
	if a != 0 {
		t.Errorf("expected transparent pixel, got alpha=%d", a>>8)
	}
}

func TestInvertToWhite_PreservesPartialAlpha(t *testing.T) {
	src := makePNG(color.NRGBA{R: 0, G: 0, B: 0, A: 128})
	got := invertToWhite(src)

	img, err := png.Decode(bytes.NewReader(got))
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	// Use non-pre-multiplied access to check RGB independently of alpha.
	nrgba, ok := img.(*image.NRGBA)
	if !ok {
		t.Fatal("expected *image.NRGBA from decoded PNG")
	}
	c := nrgba.NRGBAAt(0, 0)
	if c.R != 255 || c.G != 255 || c.B != 255 {
		t.Errorf("expected white pixel, got r=%d g=%d b=%d", c.R, c.G, c.B)
	}
	if c.A == 0 || c.A == 255 {
		t.Errorf("expected partial alpha, got %d", c.A)
	}
}

func TestInvertToWhite_InvalidInputReturnsOriginal(t *testing.T) {
	src := []byte("not a png")
	got := invertToWhite(src)
	if !bytes.Equal(got, src) {
		t.Error("expected original bytes returned for invalid PNG input")
	}
}

func TestIconLight_IsCached(t *testing.T) {
	first := IconLight()
	second := IconLight()
	if !bytes.Equal(first, second) {
		t.Error("IconLight should return the same bytes on repeated calls")
	}
}

func TestIcon2xLight_IsCached(t *testing.T) {
	first := Icon2xLight()
	second := Icon2xLight()
	if !bytes.Equal(first, second) {
		t.Error("Icon2xLight should return the same bytes on repeated calls")
	}
}

func TestIconLight_IsValidPNG(t *testing.T) {
	data := IconLight()
	if _, err := png.Decode(bytes.NewReader(data)); err != nil {
		t.Errorf("IconLight returned invalid PNG: %v", err)
	}
}

func TestIcon2xLight_IsValidPNG(t *testing.T) {
	data := Icon2xLight()
	if _, err := png.Decode(bytes.NewReader(data)); err != nil {
		t.Errorf("Icon2xLight returned invalid PNG: %v", err)
	}
}
