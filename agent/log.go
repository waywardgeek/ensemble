package main

// The log: append-only, JSON-lines, greppable with ordinary tools.
//
// That last property is not a nicety. In Chapter 3 tool output starts arriving
// by the megabyte, and a log you cannot grep is a log you cannot debug.

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// LogVersion is the semantic version of the log FORMAT. Replay happens with
// current code, not with historical code, so the version exists to let current
// code refuse a log it cannot faithfully interpret.
const LogVersion = 1

type Log struct {
	Version int
	Events  []Event

	next  Seq
	clock func() time.Time
}

func NewLog() *Log {
	return &Log{Version: LogVersion, next: 1, clock: time.Now}
}

// Append assigns the Seq and the timestamp. Seq is the ordering; Time is
// metadata that may be wrong, duplicated, or non-monotonic across machines,
// and nothing downstream is allowed to depend on it.
func (l *Log) Append(e Event) Event {
	e.Seq = l.next
	l.next++
	if e.Time.IsZero() {
		e.Time = l.clock().UTC()
	}
	l.Events = append(l.Events, e)
	return e
}

// Replay rebuilds the context from nothing but the log. This is the whole
// claim of the chapter in four lines.
func (l *Log) Replay() (*Context, error) {
	c := NewContext()
	for _, e := range l.Events {
		if err := c.Apply(e); err != nil {
			return nil, err
		}
	}
	return c, nil
}

// header is the first line of a log file. It is not an event; it carries the
// format version so that current code can refuse a log it would misread.
type header struct {
	LogVersion int `json:"log_version"`
}

func (l *Log) Write(w io.Writer) error {
	enc := json.NewEncoder(w)
	if err := enc.Encode(header{LogVersion: l.Version}); err != nil {
		return err
	}
	for _, e := range l.Events {
		if err := enc.Encode(e); err != nil {
			return err
		}
	}
	return nil
}

// ReadLog loads a log. An unknown event type is a REFUSAL TO LOAD, loudly —
// not a skip. Skipping one silently produces a context that is wrong in a way
// nothing downstream can detect, which is the worst failure mode available.
func ReadLog(r io.Reader) (*Log, error) {
	l := &Log{Version: LogVersion, next: 1, clock: time.Now}
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		// The header is optional on read: a log without one is assumed to be
		// the current format. Being lenient about a field we control, and
		// strict about event types we must interpret, is the right split.
		if !strings.Contains(line, `"seq"`) && strings.Contains(line, `"log_version"`) {
			var h header
			if err := json.Unmarshal([]byte(line), &h); err != nil {
				return nil, fmt.Errorf("line %d: %w", lineNo, err)
			}
			if h.LogVersion > LogVersion {
				return nil, fmt.Errorf("log format version %d is newer than this build understands (%d): refusing to load", h.LogVersion, LogVersion)
			}
			l.Version = h.LogVersion
			continue
		}
		var e Event
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNo, err)
		}
		l.Events = append(l.Events, e)
		if e.Seq >= l.next {
			l.next = e.Seq + 1
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return l, nil
}

func LoadLogFile(path string) (*Log, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ReadLog(f)
}

func (l *Log) SaveFile(path string) error {
	if path == "" {
		return nil
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return l.Write(f)
}
