package gradeedition3

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func TestHostScreenPixelOracle(t *testing.T) {
	challenge, _, err := newImageChallenge()
	if err != nil {
		t.Fatal(err)
	}
	raster := image.NewNRGBA(image.Rect(0, 0, 64, 32))
	for y := 0; y < 32; y++ {
		for x := 0; x < 64; x++ {
			raster.SetNRGBA(x, y, challenge.colors[x/16])
		}
	}
	encode := func() []byte {
		var b bytes.Buffer
		if err := png.Encode(&b, raster); err != nil {
			t.Fatal(err)
		}
		return b.Bytes()
	}
	pixels := encode()
	if err := validateHostScreenPNG(pixels, 64, 32, challenge); err != nil {
		t.Fatal(err)
	}
	for _, size := range [][2]int{{32, 32}, {64, 16}, {4096, 32}, {0, 0}} {
		if err := validateHostScreenPNG(pixels, size[0], size[1], challenge); err == nil {
			t.Fatal("accepted wrong dimensions", size)
		}
	}
	if err := validateHostScreenPNG([]byte("screenshot.png"), 64, 32, challenge); err == nil {
		t.Fatal("accepted filename as pixels")
	}
	for x := 0; x < 64; x++ {
		raster.SetNRGBA(x, 0, color.NRGBA{0, 0, 0, 255})
	}
	if err := validateHostScreenPNG(encode(), 64, 32, challenge); err != nil {
		t.Fatal("outer fullscreen edge was treated as an interior pixel", err)
	}
	raster.SetNRGBA(8, 2, color.NRGBA{255, 0, 255, 255})
	if err := validateHostScreenPNG(encode(), 64, 32, challenge); err == nil {
		t.Fatal("accepted alteration on first checked row")
	}
	raster.SetNRGBA(8, 2, challenge.colors[0])
	raster.SetNRGBA(8, 8, color.NRGBA{255, 0, 255, 255})
	if err := validateHostScreenPNG(encode(), 64, 32, challenge); err == nil {
		t.Fatal("accepted altered interior pixel")
	}
}

func TestHostScreenOracleRejectsEmptyInteriors(t *testing.T) {
	challenge, _, err := newImageChallenge()
	if err != nil {
		t.Fatal(err)
	}
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, image.NewNRGBA(image.Rect(0, 0, 16, 8))); err != nil {
		t.Fatal(err)
	}
	if err := validateHostScreenPNG(encoded.Bytes(), 16, 8, challenge); err == nil {
		t.Fatal("seam exclusions left no pixels to verify")
	}
}

func TestHostScreenBroadInteriorBoundaries(t *testing.T) {
	challenge, _, err := newImageChallenge()
	if err != nil {
		t.Fatal(err)
	}
	raster := image.NewNRGBA(image.Rect(0, 0, 400, 200))
	for y := 0; y < 200; y++ {
		for x := 0; x < 400; x++ {
			raster.SetNRGBA(x, y, challenge.colors[x/100])
		}
	}
	check := func() error {
		var encoded bytes.Buffer
		if err := png.Encode(&encoded, raster); err != nil {
			t.Fatal(err)
		}
		return validateHostScreenPNG(encoded.Bytes(), 400, 200, challenge)
	}
	altered := func(c color.NRGBA) color.NRGBA { c.R ^= 128; return c }
	raster.SetNRGBA(395, 4, altered(challenge.colors[3]))
	raster.SetNRGBA(50, 9, altered(challenge.colors[0]))
	if err := check(); err != nil {
		t.Fatal("edge decorations rejected", err)
	}
	for band := 0; band < 4; band++ {
		x, y := band*100+5, 10
		raster.SetNRGBA(x, y, altered(challenge.colors[band]))
		if err := check(); err == nil {
			t.Fatal("altered first interior pixel accepted", band)
		}
		raster.SetNRGBA(x, y, challenge.colors[band])
	}
}
