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
	raster.SetNRGBA(8, 8, color.NRGBA{255, 0, 255, 255})
	if err := validateHostScreenPNG(encode(), 64, 32, challenge); err == nil {
		t.Fatal("accepted altered interior pixel")
	}
}
