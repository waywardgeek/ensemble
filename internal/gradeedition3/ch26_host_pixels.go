package gradeedition3

import (
	"bytes"
	"fmt"
	"image/color"
	"image/png"
)

// Native acceptance must supply dimensions and colors of the agreed full-display
// synthetic pattern. This oracle never enumerates or captures a display.
func validateHostScreenPNG(pixels []byte, width, height int, challenge imageChallenge) error {
	if len(pixels) == 0 || len(pixels) > 8<<20 || width < 32 || height < 8 || width > 2048 || height > 2048 {
		return fmt.Errorf("host PNG or expected dimensions exceed bounds")
	}
	config, err := png.DecodeConfig(bytes.NewReader(pixels))
	if err != nil || config.Width != width || config.Height != height {
		return fmt.Errorf("host PNG dimensions do not match agreed synthetic display")
	}
	raster, err := png.Decode(bytes.NewReader(pixels))
	if err != nil {
		return fmt.Errorf("host PNG cannot decode")
	}
	// Validate broad predetermined interiors, not browser/OS edge decoration.
	// Five percent per side (at least two pixels) still checks about 81% of
	// the image. Exact dimensions and byte-identical full replay are separate.
	verticalInset := max(2, height/20)
	for band, want := range challenge.colors {
		left, right := band*width/4, (band+1)*width/4
		horizontalInset := max(2, (right-left)/20)
		for y := verticalInset; y < height-verticalInset; y++ {
			for x := left + horizontalInset; x < right-horizontalInset; x++ {
				got := color.NRGBAModel.Convert(raster.At(x, y)).(color.NRGBA)
				for _, pair := range [][2]uint8{{got.R, want.R}, {got.G, want.G}, {got.B, want.B}, {got.A, want.A}} {
					delta := int(pair[0]) - int(pair[1])
					if delta < -3 || delta > 3 {
						centers := make([]color.NRGBA, 4)
						for index := range centers {
							centers[index] = color.NRGBAModel.Convert(raster.At((2*index+1)*width/8, height/2)).(color.NRGBA)
						}
						return fmt.Errorf("host pixels differ from agreed synthetic display pattern: first mismatch (%d,%d), got RGBA %v, expected %v; band-center RGBA %v", x, y, got, want, centers)
					}
				}
			}
		}
	}
	return nil
}
