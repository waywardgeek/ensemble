package gradeedition3

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func validateHostPattern(pattern HostNativePattern) error {
	if pattern.Width < 32 || pattern.Height < 8 || pattern.Width > 2048 || pattern.Height > 2048 {
		return fmt.Errorf("invalid prepared synthetic-display dimensions")
	}
	for index, rgb := range pattern.Colors {
		for previous := 0; previous < index; previous++ {
			if pattern.Colors[previous] == rgb {
				return fmt.Errorf("synthetic display bands must have distinct colors")
			}
		}
	}
	return nil
}

// ReadHostPattern rejects ambiguous or oversized manifests before native work.
func ReadHostPattern(path string) (HostNativePattern, error) {
	var pattern HostNativePattern
	file, err := os.Open(path)
	if err != nil {
		return pattern, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, 8193))
	if err != nil {
		return pattern, err
	}
	if len(data) > 8192 {
		return pattern, fmt.Errorf("host pattern exceeds 8 KiB")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return pattern, fmt.Errorf("host pattern must be an object")
	}
	seen := map[string]bool{}
	for decoder.More() {
		key, err := decoder.Token()
		if err != nil {
			return pattern, err
		}
		name, ok := key.(string)
		if !ok || seen[name] {
			return pattern, fmt.Errorf("duplicate or invalid pattern field")
		}
		seen[name] = true
		switch name {
		case "prepared_width":
			err = decoder.Decode(&pattern.Width)
		case "prepared_height":
			err = decoder.Decode(&pattern.Height)
		case "colors":
			var rows [][]int
			if err = decoder.Decode(&rows); err != nil {
				break
			}
			if len(rows) != 4 {
				return pattern, fmt.Errorf("pattern requires four RGB colors")
			}
			for index, row := range rows {
				if len(row) != 3 {
					return pattern, fmt.Errorf("pattern requires RGB triples")
				}
				for channel, value := range row {
					if value < 0 || value > 255 {
						return pattern, fmt.Errorf("invalid pattern color")
					}
					pattern.Colors[index][channel] = uint8(value)
				}
			}
		default:
			return pattern, fmt.Errorf("unknown host pattern field %q", name)
		}
		if err != nil {
			return pattern, err
		}
	}
	if _, err = decoder.Token(); err != nil {
		return pattern, err
	}
	var extra any
	if err = decoder.Decode(&extra); err != io.EOF {
		return pattern, fmt.Errorf("trailing host pattern data")
	}
	if len(seen) != 3 {
		return pattern, fmt.Errorf("host pattern requires dimensions and colors")
	}
	return pattern, validateHostPattern(pattern)
}

// WriteHostPattern creates static local fixtures, never opens a browser or
// enumerates/captures a display. Existing files are refused, not overwritten.
func WriteHostPattern(directory string, width, height int) (HostNativePattern, error) {
	pattern := HostNativePattern{Width: width, Height: height}
	var random [8]byte
	if _, err := rand.Read(random[:]); err != nil {
		return pattern, err
	}
	for index := range pattern.Colors {
		pattern.Colors[index] = [3]uint8{uint8(32 + 48*index), uint8(32 + random[index*2]%192), uint8(32 + random[index*2+1]%192)}
	}
	if err := validateHostPattern(pattern); err != nil {
		return pattern, err
	}
	if err := os.MkdirAll(directory, 0700); err != nil {
		return pattern, err
	}
	manifest, err := json.MarshalIndent(pattern, "", "  ")
	if err != nil {
		return pattern, err
	}
	var bands strings.Builder
	for _, rgb := range pattern.Colors {
		fmt.Fprintf(&bands, "<div style=\"background:rgb(%d,%d,%d)\"></div>", rgb[0], rgb[1], rgb[2])
	}
	html := `<!doctype html><meta charset="utf-8"><title>Synthetic primary-display test</title>
<style>html,body{margin:0;width:100%;height:100%;overflow:hidden}#pattern{position:fixed;inset:0;display:grid;grid-template-columns:repeat(4,1fr)}#instructions{position:fixed;left:10%;right:10%;top:25%;padding:2rem;background:white;color:black;font:20px sans-serif}:fullscreen #instructions{display:none}:fullscreen,:fullscreen *{cursor:none!important}</style>
<div id="pattern">` + bands.String() + `</div><section id="instructions"><h1>Synthetic display test</h1><p>Place this window on the primary display. This button only enters full-screen; it does not take a screenshot.</p><p>Confirm that all four bands fill the entire display before separately authorizing the grader. Escape exits full-screen. Do not capture while these instructions are visible.</p><button id="fullscreen">Enter full-screen pattern</button></section>
<script>document.querySelector('#fullscreen').onclick=()=>document.documentElement.requestFullscreen().catch(error=>alert(error.message));</script>`
	writeNew := func(path string, data []byte) error {
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			return err
		}
		_, writeErr := file.Write(data)
		return errors.Join(writeErr, file.Close())
	}
	manifestPath := filepath.Join(directory, "pattern.json")
	if err := writeNew(manifestPath, manifest); err != nil {
		return pattern, err
	}
	if err := writeNew(filepath.Join(directory, "display.html"), []byte(html)); err != nil {
		return pattern, errors.Join(err, os.Remove(manifestPath))
	}
	return pattern, nil
}
