package gui

import (
	"embed"
	"io/fs"
)

// Keep the required embed anchor private: callers can read its contents but
// cannot replace the package's filesystem binding for other applications.
//
//go:embed web/gui
var assets embed.FS

// Assets returns read-only browser files rooted at index.html. Applications may
// serve just artifact-scroll.js and renderers.js to reuse the scroll without the
// application shell; the component does not open a socket or require that shell.
func Assets() fs.FS {
	files, err := fs.Sub(assets, "web/gui")
	if err != nil {
		panic(err) // The embedded directory is required at compile time.
	}
	return files
}
