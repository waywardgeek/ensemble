package gui

import (
	"embed"
	"io/fs"
	"net/http"
)

// Assets holds the GUI's static files, compiled into the binary so the
// server works from any working directory. The files sit under "web/" in
// this FS. Exported so an embedding application can mount individual
// components (artifact-scroll.js, renderers.js) without adopting the shell.
//
//go:embed web
var Assets embed.FS

// StaticHandler serves the GUI's static files. An empty dir serves the
// embedded Assets; a non-empty dir serves that directory from disk instead,
// for live development (edit a .js file, reload, no rebuild).
func StaticHandler(dir string) http.Handler {
	if dir != "" {
		return http.FileServer(http.Dir(dir))
	}
	sub, err := fs.Sub(Assets, "web")
	if err != nil {
		// The web subtree is embedded at compile time; failing to find it is
		// a programmer error (directory renamed), not a runtime condition.
		panic("gui: embedded assets missing web/ subtree: " + err.Error())
	}
	return http.FileServer(http.FS(sub))
}
