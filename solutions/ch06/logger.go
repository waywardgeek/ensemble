package agent

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Logger writes timestamped messages to an io.Writer. It is safe for
// concurrent use from any goroutine.
type Logger struct {
	mu sync.Mutex
	w  io.Writer
}

// NewLogger creates a logger that writes to the given writer.
func NewLogger(w io.Writer) *Logger {
	return &Logger{w: w}
}

// DefaultLogger returns a logger that writes to stderr.
func DefaultLogger() *Logger {
	return NewLogger(os.Stderr)
}

// Logf writes a timestamped, newline-terminated message.
func (l *Logger) Logf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	stamp := time.Now().Format("15:04:05.000")
	line := fmt.Sprintf("[%s] %s\n", stamp, msg)
	l.mu.Lock()
	l.w.Write([]byte(line))
	l.mu.Unlock()
}
