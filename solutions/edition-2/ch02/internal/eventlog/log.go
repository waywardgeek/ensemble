// Package eventlog owns durable JSON-lines I/O, not conversation interpretation.
package eventlog

import (
	"bufio"
	"bytes"
	"encoding/json"
	"example.com/ensemble/internal/common"
	"fmt"
	"io"
	"os"
)

type Log struct {
	parent  common.Agent
	writer  io.WriteCloser
	faulted bool
}

func New(parent common.Agent, path string) (*Log, error) {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return nil, fmt.Errorf("cannot create fresh event log")
	}
	log := &Log{parent: parent, writer: f}
	if err = log.write([]byte("{\"log_version\":1}\n")); err != nil {
		f.Close()
		return nil, err
	}
	return log, nil
}
func (l *Log) Agent() common.Agent { return l.parent }
func (l *Log) Append(event common.Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("cannot encode event")
	}
	return l.write(append(data, '\n'))
}
func (l *Log) write(data []byte) error {
	if l.faulted {
		return fmt.Errorf("event log faulted; create a fresh Agent and log")
	}
	n, err := l.writer.Write(data)
	if err != nil || n != len(data) {
		l.faulted = true
		l.parent.Ensemble().Logf("event log write failed; Agent faulted")
		return fmt.Errorf("event log write failed; Agent faulted")
	}
	return nil
}
func (l *Log) Close() error { l.faulted = true; return l.writer.Close() }

func Read(parent common.Agent, reader io.Reader) ([]common.Event, []int, error) {
	scan := bufio.NewScanner(reader)
	scan.Buffer(make([]byte, 4096), 16*1024*1024)
	events := []common.Event{}
	lines := []int{}
	line := 0
	header := false
	fail := func(reason string) ([]common.Event, []int, error) {
		parent.Ensemble().Logf("invalid event log at line %d: %s", line, reason)
		return nil, nil, fmt.Errorf("invalid event log at line %d: %s", line, reason)
	}
	for scan.Scan() {
		line++
		b := bytes.TrimSpace(scan.Bytes())
		if len(b) == 0 {
			continue
		}
		if !header {
			var h struct {
				Version int `json:"log_version"`
			}
			if json.Unmarshal(b, &h) != nil || h.Version != 1 {
				return fail("missing or unsupported log version")
			}
			header = true
			continue
		}
		var keys map[string]json.RawMessage
		if json.Unmarshal(b, &keys) != nil {
			return fail("malformed JSON record")
		}
		count := 0
		for _, key := range []string{"message", "request", "response", "tool", "redact", "error"} {
			if raw, ok := keys[key]; ok {
				count++
				if bytes.Equal(raw, []byte("null")) {
					return fail("null event payload")
				}
			}
		}
		if count != 1 {
			return fail("exactly one event payload is required")
		}
		var e common.Event
		if json.Unmarshal(b, &e) != nil {
			return fail("invalid event field type or value")
		}
		events = append(events, e)
		lines = append(lines, line)
	}
	if scan.Err() != nil {
		return fail("cannot read complete log record")
	}
	if !header {
		return fail("missing log header")
	}
	return events, lines, nil
}
func Dump(parent common.Agent, events []common.Event) ([]byte, error) {
	var out bytes.Buffer
	out.WriteString("{\"log_version\":1}\n")
	enc := json.NewEncoder(&out)
	for _, e := range events {
		if err := enc.Encode(e); err != nil {
			parent.Ensemble().Logf("cannot encode event log")
			return nil, fmt.Errorf("cannot encode event log")
		}
	}
	return out.Bytes(), nil
}
