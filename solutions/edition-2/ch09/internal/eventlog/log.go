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

// CheckSkillRecord measures the complete fact after sequence/time assignment,
// before persistence or publication. Refusal does not fault storage.
func CheckSkillRecord(parent common.Agent, event common.Event) error {
	if event.Type != "skills_initialized" && event.Type != "skills_changed" {
		return nil
	}
	data, err := json.Marshal(event)
	if err != nil {
		parent.Ensemble().Logf("cannot encode skill record")
		return fmt.Errorf("cannot encode event")
	}
	if len(data)+1 > common.SkillRecordLimit {
		failure := &common.SkillError{Code: "skill_too_large", Detail: "complete skill record exceeds 67108864 bytes"}
		if event.Skills != nil {
			failure.Name = event.Skills.Name
			if event.Type == "skills_changed" && event.Skills.State.Revision > 0 {
				failure.Revision = event.Skills.State.Revision - 1
			}
		}
		return failure
	}
	return nil
}

func (l *Log) Append(event common.Event) error {
	if err := CheckSkillRecord(l.parent, event); err != nil {
		return err
	}
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

// Keep the retained physical record within its budget. ReadSlice uses a fixed
// scratch buffer; overflow is detected before copying any excess into record.
func readRecord(parent common.Agent, input *bufio.Reader) ([]byte, error) {
	var record []byte
	for {
		piece, err := input.ReadSlice('\n')
		need := len(record) + len(piece)
		if need > common.SkillRecordLimit {
			parent.Ensemble().Logf("event log record exceeds raw byte limit")
			return nil, fmt.Errorf("record exceeds raw byte limit")
		}
		if need > cap(record) {
			capacity := max(need, 2*cap(record))
			capacity = min(capacity, common.SkillRecordLimit)
			next := make([]byte, len(record), capacity)
			copy(next, record)
			record = next
		}
		record = append(record, piece...)
		if err != bufio.ErrBufferFull {
			return record, err
		}
	}
}

func ordinaryRecordFits(record []byte) bool {
	// Scanner's inherited maximum includes LF when present, but needs spare
	// capacity to discover EOF for an unterminated token.
	return len(record) < 16*1024*1024 || len(record) == 16*1024*1024 && record[len(record)-1] == '\n'
}

func Read(parent common.Agent, reader io.Reader) ([]common.Event, []int, error) {
	input := bufio.NewReader(reader)
	events := []common.Event{}
	lines := []int{}
	line := 0
	header := false
	fail := func(reason string) ([]common.Event, []int, error) {
		parent.Ensemble().Logf("invalid event log at line %d: %s", line, reason)
		return nil, nil, fmt.Errorf("invalid event log at line %d: %s", line, reason)
	}
	for {
		record, readErr := readRecord(parent, input)
		if readErr != nil && readErr != io.EOF {
			line++
			return fail("cannot read complete log record")
		}
		if len(record) == 0 && readErr == io.EOF {
			break
		}
		line++
		b := bytes.TrimSpace(record)
		if len(b) == 0 {
			if !ordinaryRecordFits(record) {
				return fail("cannot read complete log record")
			}
			continue
		}
		if !header {
			if !ordinaryRecordFits(record) {
				return fail("cannot read complete log record")
			}
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
		var kind string
		_ = json.Unmarshal(keys["type"], &kind)
		if kind != "skills_initialized" && kind != "skills_changed" && !ordinaryRecordFits(record) {
			return fail("cannot read complete log record")
		}
		count := 0
		for _, key := range []string{"message", "request", "response", "tool", "redact", "error", "job", "turn", "hint", "skills"} {
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
