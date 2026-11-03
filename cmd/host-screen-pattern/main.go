// Create local synthetic display fixtures; never opens or captures a display.
package main

import (
	"flag"
	"fmt"
	"github.com/waywardgeek/ensemble/internal/gradeedition3"
	"os"
	"path/filepath"
)

func main() {
	width := flag.Int("width", 0, "expected prepared PNG width, at most2048")
	height := flag.Int("height", 0, "expected prepared PNG height, at most2048")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: host-screen-pattern -width N -height N NEW_DIRECTORY")
		os.Exit(2)
	}
	directory, err := filepath.Abs(flag.Arg(0))
	if err == nil {
		_, err = gradeedition3.WriteHostPattern(directory, *width, *height)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("Created", filepath.Join(directory, "display.html"), "and pattern.json. Nothing opened or captured.")
	fmt.Println("Verify prepared dimensions, open the HTML on the primary display, use its full-screen button, wait for browser overlays to disappear, and obtain separate native-test authorization.")
}
