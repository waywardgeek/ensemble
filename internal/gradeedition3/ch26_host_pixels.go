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
	if len(pixels) == 0 || len(pixels) > 8<<20 || width < 16 || height < 8 || width > 2048 || height > 2048 {
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
	for band, want := range challenge.colors {
		for y := 0; y < height; y++ {
			// Ignore only resampling seams, not arbitrary private/background regions.
			for x := band*width/4 + 2; x < (band+1)*width/4-2; x++ {
				got := color.NRGBAModel.Convert(raster.At(x, y)).(color.NRGBA)
				for _, pair := range [][2]uint8{{got.R, want.R}, {got.G, want.G}, {got.B, want.B}, {got.A, want.A}} {
					delta := int(pair[0]) - int(pair[1])
					if delta < -3 || delta > 3 {
						return fmt.Errorf("host pixels differ from agreed synthetic display pattern")
					}
				}
			}
		}
	}
	return nil
}
