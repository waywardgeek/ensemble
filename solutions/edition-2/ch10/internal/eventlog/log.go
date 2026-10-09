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
	parent          common.Agent
	writer          io.WriteCloser
	faulted         bool
	session         bool
	bytes           int64
	count           uint64
	separator       bool
	lastRecordBytes int64
}

func New(parent common.Agent, path string) (*Log, error) {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return nil, fmt.Errorf("cannot create fresh event log")
	}
	log := &Log{parent: parent, writer: f, session: parent.Config().DataDir != ""}
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
	padding := int64(0)
	if l.separator {
		padding = 1
	}
	if l.session && (len(data)+1 > common.SkillRecordLimit || l.bytes+int64(len(data)+1)+padding > 1<<30 || l.count >= 1000000 || l.separator && l.lastRecordBytes >= common.SkillRecordLimit) {
		l.faulted = true
		return &common.SessionError{Code: "session_limit", Detail: "event record, log byte or event count bound exceeded"}
	}
	if l.separator {
		data = append([]byte{'\n'}, data...)
	}
	if err = l.write(append(data, '\n')); err != nil {
		return err
	}
	l.separator = false
	l.count++
	return nil
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
	l.bytes += int64(n)
	return nil
}
func (l *Log) Close() error { l.faulted = true; return l.writer.Close() }

// Keep the retained physical record within its budget. ReadSlice uses a fixed
// scratch buffer; overflow is detected before copying any excess into record.
func readRecord(parent common.Agent, input *bufio.Reader) ([]byte, error) {
	return readRecordLimit(parent, input, common.SkillRecordLimit)
}
func readRecordLimit(parent common.Agent, input *bufio.Reader, limit int) ([]byte, error) {
	var record []byte
	for {
		piece, err := input.ReadSlice('\n')
		need := len(record) + len(piece)
		if need > limit {
			parent.Ensemble().Logf("event log record exceeds raw byte limit")
			return nil, fmt.Errorf("record exceeds raw byte limit")
		}
		if need > cap(record) {
			capacity := max(need, 2*cap(record))
			capacity = min(capacity, limit)
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

type collector struct {
	common.Agent
	events []common.Event
	lines  []int
}

func (c *collector) AcceptReadEvent(e common.Event, line int) error {
	c.events = append(c.events, e)
	c.lines = append(c.lines, line)
	return nil
}
func Read(parent common.Agent, reader io.Reader) ([]common.Event, []int, error) {
	c := &collector{Agent: parent, events: []common.Event{}, lines: []int{}}
	err := Stream(c, reader, false)
	return c.events, c.lines, err
}

// Stream retains at most one bounded physical record before owner reduction.
func Stream(parent common.LogReaderAgent, reader io.Reader, session bool) error {
	input := bufio.NewReader(reader)
	line := 0
	header := false
	count := uint64(0)
	size := int64(0)
	fail := func(reason string) error {
		parent.Ensemble().Logf("invalid event log at line %d: %s", line, reason)
		return fmt.Errorf("invalid event log at line %d: %s", line, reason)
	}
	for {
		limit := common.SkillRecordLimit
		if !header {
			limit = 16 << 20
		}
		if session {
			limit = min(limit, int((1<<30)-size+1))
		}
		record, readErr := readRecordLimit(parent, input, limit)
		if readErr != nil && readErr != io.EOF {
			line++
			return fail("cannot read complete log record")
		}
		if len(record) == 0 && readErr == io.EOF {
			break
		}
		line++
		size += int64(len(record))
		if session && size > 1<<30 {
			return fail("session log exceeds 1 GiB")
		}
		b := bytes.TrimSpace(record)
		if len(b) == 0 {
			if !ordinaryRecordFits(record) {
				return fail("blank record exceeds bound")
			}
			continue
		}
		if err := parent.Codec().ValidateLogJSON(b, session); err != nil {
			return fail("invalid JSON or duplicate member")
		}
		if !header {
			if !ordinaryRecordFits(record) {
				return fail("header exceeds bound")
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
		var e common.Event
		if json.Unmarshal(b, &e) != nil {
			return fail("invalid event field type or value")
		}
		if !session && count == 0 && kind == "session_initialized" && e.Seq == 1 && e.Session != nil && e.Session.Identity != nil {
			if err := parent.Codec().ValidateLogJSON(b, true); err != nil {
				return fail("invalid session JSON")
			}
			session = true
		}
		if !session && kind != "skills_initialized" && kind != "skills_changed" && !ordinaryRecordFits(record) {
			return fail("ordinary record exceeds bound")
		}
		if session && (size > 1<<30 || count >= 1000000) {
			return fail("session storage bound")
		}
		payloads := 0
		for _, key := range []string{"message", "request", "response", "tool", "redact", "error", "job", "turn", "hint", "skills", "session", "limits"} {
			if raw, ok := keys[key]; ok {
				payloads++
				if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
					return fail("null event payload")
				}
			}
		}
		if payloads != 1 {
			return fail("exactly one event payload is required")
		}
		if err := parent.AcceptReadEvent(e, line); err != nil {
			return err
		}
		count++
	}
	if !header {
		return fail("missing log header")
	}
	return nil
}

// Resume opens the sole writer after the complete candidate has validated.
func Resume(parent common.Agent, path string, count uint64) (*Log, error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND, 0600)
	if err != nil {
		return nil, fmt.Errorf("cannot open validated event log")
	}
	st, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("cannot stat validated log")
	}
	l := &Log{parent: parent, writer: f, session: true, bytes: st.Size(), count: count}
	var last [1]byte
	if st.Size() > 0 {
		if _, err = f.ReadAt(last[:], st.Size()-1); err != nil {
			f.Close()
			return nil, fmt.Errorf("cannot inspect validated log ending")
		}
		l.separator = last[0] != '\n'
	}
	if l.separator {
		var block [4096]byte
		for end := st.Size(); end > 0; {
			start := max(int64(0), end-int64(len(block)))
			n, err := f.ReadAt(block[:end-start], start)
			if err != nil {
				f.Close()
				return nil, fmt.Errorf("cannot inspect validated final record")
			}
			if i := bytes.LastIndexByte(block[:n], '\n'); i >= 0 {
				l.lastRecordBytes += int64(n - i - 1)
				break
			}
			l.lastRecordBytes += int64(n)
			end = start
			if l.lastRecordBytes > common.SkillRecordLimit {
				f.Close()
				return nil, fmt.Errorf("final record exceeds session bound")
			}
		}
	}
	return l, nil
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
