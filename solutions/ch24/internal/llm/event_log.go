// common.Event log behavior: append, replay, and the on-disk form.
//
// The Log struct itself is shared vocabulary and stays in the hub, because
// three spokes read its Events slice. The operations over it belong here, as
// free functions, for the same reason as context_ops.go.
package llm

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

func NewLog() *common.Log {
	return &common.Log{Version: common.LogVersion, Next: 1, Clock: time.Now}
}

// Append assigns the common.Seq and the timestamp. common.Seq is the ordering; Time is
// metadata that may be wrong, duplicated, or non-monotonic across machines,
// and nothing downstream is allowed to depend on it.
func Append(l *common.Log, e common.Event) common.Event {
	e.Seq = l.Next
	l.Next++
	if e.Time.IsZero() {
		e.Time = l.Clock().UTC()
	}
	l.Events = append(l.Events, e)
	return e
}

// Len returns the number of events in the log.
func Len(l *common.Log) int {
	return len(l.Events)
}

// ResetSeq sets the next sequence number. Used when restoring
// a log from a save file, so that new events don't collide
// with the ones that were loaded.
func ResetSeq(l *common.Log, next common.Seq) {
	l.Next = next
}

// NextSeq is the common.Seq the next appended event will get. A save whose log
// was trimmed away has no last event to read the anchor off, so the
// anchor has to come from here instead.
func NextSeq(l *common.Log) common.Seq {
	return l.Next
}

// Replay rebuilds the context from nothing but the log. This is the whole
// claim of the chapter in four lines.
//
// Chapter 11 generalises this to snapshot-plus-tail, and there must be
// exactly one replay loop in the agent or the two will drift apart and
// disagree about some event nobody thought to test. So this is now the
// special case it always was: a save with no snapshot and no anchor.
func Replay(l *common.Log) (*common.Context, error) {
	sf := &SaveFile{Log: l.Events}
	// The loop is total (Chapter 15 rule 8), but a caller that asks for a
	// replay of a log FILE — render, verify — still hears about every event
	// that could not be applied, as an error, so a bad log is not quietly
	// rendered as if it were a good one.
	var skipped []error
	ctx := sf.Restore(func(err error) { skipped = append(skipped, err) })
	if len(skipped) > 0 {
		return nil, errors.Join(skipped...)
	}
	return ctx, nil
}

// header is the first line of a log file. It is not an event; it carries the
// format version so that current code can refuse a log it would misread.
type header struct {
	LogVersion int `json:"log_version"`
}

func Write(l *common.Log, w io.Writer) error {
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
func ReadLog(r io.Reader) (*common.Log, error) {
	l := &common.Log{Version: common.LogVersion, Next: 1, Clock: time.Now}
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
			if h.LogVersion > common.LogVersion {
				return nil, fmt.Errorf("log format version %d is newer than this build understands (%d): refusing to load", h.LogVersion, common.LogVersion)
			}
			l.Version = h.LogVersion
			continue
		}
		var e common.Event
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNo, err)
		}
		l.Events = append(l.Events, e)
		if e.Seq >= l.Next {
			l.Next = e.Seq + 1
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return l, nil
}

func LoadLogFile(path string) (*common.Log, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ReadLog(f)
}

func SaveLogFile(l *common.Log, path string) error {
	if path == "" {
		return nil
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return Write(l, f)
}
