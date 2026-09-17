package agent

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Logger writes timestamped messages to three destinations:
//   - Event log (Logf): structured agent events
//   - API log (APILogf): raw LLM JSON requests and responses
//   - Debug log (Debugf): arbitrary text, mirrored to stderr
//
// All methods are safe for concurrent use.
type Logger struct {
	mu     sync.Mutex
	event  io.Writer // event log (bin.log)
	api    io.Writer // API wire traffic (api.log)
	debug  io.Writer // debug log file (debug.log)
	stderr io.Writer // terminal output for Debugf
}

// NewLogger creates a logger that writes events to the given writer.
// API and debug output go to io.Discard until configured.
func NewLogger(w io.Writer) *Logger {
	return &Logger{event: w, api: io.Discard, debug: io.Discard, stderr: os.Stderr}
}

// SetAPILog sets the writer for LLM API wire traffic.
func (l *Logger) SetAPILog(w io.Writer) { l.api = w }

// SetDebugLog sets the writer for the debug log file.
// Debug messages are always also printed to stderr.
func (l *Logger) SetDebugLog(w io.Writer) { l.debug = w }

// DefaultLogger returns a logger that writes everything to stderr.
func DefaultLogger() *Logger {
	return NewLogger(os.Stderr)
}

// Logf writes a timestamped event message.
func (l *Logger) Logf(format string, args ...any) {
	l.write(l.event, format, args...)
}

// APILogf writes LLM API wire traffic (JSON requests and responses).
func (l *Logger) APILogf(format string, args ...any) {
	l.write(l.api, format, args...)
}

// Debugf writes to both the debug log file and stderr.
func (l *Logger) Debugf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	stamp := time.Now().Format("15:04:05.000")
	line := fmt.Sprintf("[%s] %s\n", stamp, msg)
	l.mu.Lock()
	l.debug.Write([]byte(line))
	l.stderr.Write([]byte(line))
	l.mu.Unlock()
}

func (l *Logger) write(w io.Writer, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	stamp := time.Now().Format("15:04:05.000")
	line := fmt.Sprintf("[%s] %s\n", stamp, msg)
	l.mu.Lock()
	w.Write([]byte(line))
	l.mu.Unlock()
}
