package ws

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// GuiLogger writes transport-layer logs in the format:
//
//	2026-09-17T14:32:01.123Z > {"type":"subscribe","cursor":0}
//	2026-09-17T14:32:01.456Z < {"type":"part_delta",...}
//
// > is client-to-server, < is server-to-client.
type GuiLogger struct {
	mu   sync.Mutex
	file *os.File
}

// NewGuiLogger opens the gui.log file for writing. If the path is empty,
// logging is silently disabled.
func NewGuiLogger(path string) *GuiLogger {
	if path == "" {
		return &GuiLogger{}
	}
	f, err := os.Create(path)
	if err != nil {
		return &GuiLogger{}
	}
	return &GuiLogger{file: f}
}

// Log writes one line to gui.log. dir is ">" or "<".
func (g *GuiLogger) Log(dir string, data []byte) {
	if g == nil || g.file == nil {
		return
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	ts := time.Now().UTC().Format(time.RFC3339Nano)
	fmt.Fprintf(g.file, "%s %s %s\n", ts, dir, data)
}

// Close closes the underlying file.
func (g *GuiLogger) Close() {
	if g == nil || g.file == nil {
		return
	}
	g.file.Close()
}
