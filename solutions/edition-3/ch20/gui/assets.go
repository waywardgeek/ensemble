package gui

import (
	"embed"
	"io/fs"
)

//go:embed web/gui
var assets embed.FS

// Assets returns a read-only filesystem rooted at the browser assets. Consumers
// can mount individual renderer modules without constructing a Server.
func Assets() fs.FS {
	files, err := fs.Sub(assets, "web/gui")
	if err != nil {
		panic(err)
	}
	return files
}
