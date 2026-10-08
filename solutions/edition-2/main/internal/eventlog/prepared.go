package eventlog

import (
	"encoding/json"
	"example.com/ensemble/internal/common"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

type preparedEvent struct {
	parent common.EventLog
	event  common.Event
	bytes  []byte
}

func (p *preparedEvent) Log() common.EventLog { return p.parent }
func (p *preparedEvent) Event() common.Event  { return p.event }

// Prepare is the sole new-session acceptance encoding. Its private value is
// scoped to this log, so AppendPrepared never accepts a foreign byte buffer.
func (l *Log) Prepare(event common.Event) (common.PreparedEvent, error) {
	data, err := encodeRecord(l.parent, event, common.SkillRecordLimit-1)
	if err != nil {
		if problem, ok := err.(*common.SessionError); ok && problem.Code == "session_limit" && (event.Type == "skills_initialized" || event.Type == "skills_changed") {
			name := ""
			revision := uint64(0)
			if event.Skills != nil {
				name = event.Skills.Name
				if event.Skills.State.Revision > 0 {
					revision = event.Skills.State.Revision - 1
				}
			}
			return nil, &common.SkillError{Code: "skill_too_large", Name: name, Revision: revision, Detail: "complete skill record exceeds 67108864 bytes"}
		}
		return nil, err
	}
	if err = l.parent.Codec().ValidateLogJSON(data, true); err != nil {
		return nil, err
	}
	var accepted common.Event
	if err = json.Unmarshal(data, &accepted); err != nil {
		return nil, &common.SessionError{Code: "session_corrupt", Detail: "prepared event cannot decode"}
	}
	return &preparedEvent{parent: l, event: accepted, bytes: append(data, '\n')}, nil
}
func (l *Log) AppendPrepared(value common.PreparedEvent) error {
	p, ok := value.(*preparedEvent)
	if !ok || p.parent != l {
		return fmt.Errorf("foreign prepared event")
	}
	padding := int64(0)
	if l.separator {
		padding = 1
	}
	if l.bytes+int64(len(p.bytes))+padding > 1<<30 || l.count >= 1000000 || l.separator && l.lastRecordBytes >= common.SkillRecordLimit {
		l.faulted = true
		return &common.SessionError{Code: "session_limit", Detail: "session log byte or event count bound exceeded"}
	}
	if l.separator {
		if err := l.write([]byte{'\n'}); err != nil {
			return err
		}
		l.separator = false
	}
	if err := l.write(p.bytes); err != nil {
		return err
	}
	l.count++
	return nil
}

// Encode directly into a capped buffer. Raw JSON is lexically compacted without
// sorting, decoding numbers, or altering number tokens. Typed fields use their
// ordinary struct order and omission rules. Every helper retains the owner chain.
func encodeRecord(owner common.Agent, event common.Event, limit int) ([]byte, error) {
	data := make([]byte, 0, 4096)
	add := func(s string) error {
		if len(s) > limit-len(data) {
			return &common.SessionError{Code: "session_limit", Detail: "session event record exceeds 64 MiB"}
		}
		need := len(data) + len(s)
		if need > cap(data) {
			next := make([]byte, len(data), min(limit, max(need, 2*cap(data))))
			copy(next, data)
			data = next
		}
		data = append(data, s...)
		return nil
	}
	quoted := func(s string) error {
		if !utf8.ValidString(s) {
			return &common.SessionError{Code: "session_corrupt", Detail: "event string is not UTF-8"}
		}
		if err := add(`"`); err != nil {
			return err
		}
		start := 0
		for i, r := range s {
			escaped := ""
			switch r {
			case '"':
				escaped = `\"`
			case '\\':
				escaped = `\\`
			case '\b':
				escaped = `\b`
			case '\f':
				escaped = `\f`
			case '\n':
				escaped = `\n`
			case '\r':
				escaped = `\r`
			case '\t':
				escaped = `\t`
			case '<':
				escaped = `\u003c`
			case '>':
				escaped = `\u003e`
			case '&':
				escaped = `\u0026`
			case '\u2028':
				escaped = `\u2028`
			case '\u2029':
				escaped = `\u2029`
			default:
				if r < 32 {
					escaped = fmt.Sprintf(`\u%04x`, r)
				}
			}
			if escaped != "" {
				if err := add(s[start:i]); err != nil {
					return err
				}
				if err := add(escaped); err != nil {
					return err
				}
				start = i + utf8.RuneLen(r)
			}
		}
		if err := add(s[start:]); err != nil {
			return err
		}
		return add(`"`)
	}
	var write func(reflect.Value) error
	write = func(v reflect.Value) error {
		if v.Kind() == reflect.Pointer {
			if v.IsNil() {
				return add("null")
			}
			return write(v.Elem())
		}
		if v.Type() == reflect.TypeOf(json.RawMessage{}) {
			if v.IsNil() {
				return add("null")
			}
			raw := v.Bytes()
			addRaw := func(raw []byte) error {
				if len(raw) > limit-len(data) {
					return &common.SessionError{Code: "session_limit", Detail: "session event record exceeds 64 MiB"}
				}
				return add(string(raw))
			}
			if !json.Valid(raw) {
				return &common.SessionError{Code: "session_corrupt", Detail: "invalid candidate raw JSON"}
			}
			inside, escaped := false, false
			start := 0
			for i, b := range raw {
				if inside {
					if escaped {
						escaped = false
					} else if b == '\\' {
						escaped = true
					} else if b == '"' {
						inside = false
					}
					continue
				}
				if b == '"' {
					inside = true
					continue
				}
				if b == ' ' || b == '\r' || b == '\n' || b == '\t' {
					if err := addRaw(raw[start:i]); err != nil {
						return err
					}
					start = i + 1
				}
			}
			return addRaw(raw[start:])
		}
		switch v.Kind() {
		case reflect.Struct:
			if err := add("{"); err != nil {
				return err
			}
			first := true
			t := v.Type()
			for i := 0; i < t.NumField(); i++ {
				f := t.Field(i)
				if !f.IsExported() {
					continue
				}
				tag := strings.Split(f.Tag.Get("json"), ",")
				name := tag[0]
				if name == "-" {
					continue
				}
				if name == "" {
					name = f.Name
				}
				field := v.Field(i)
				omit := len(tag) > 1 && tag[1] == "omitempty"
				if omit {
					empty := field.IsZero()
					switch field.Kind() {
					case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
						empty = field.Len() == 0
					}
					if empty {
						continue
					}
				}
				if !first {
					if err := add(","); err != nil {
						return err
					}
				}
				first = false
				if err := quoted(name); err != nil {
					return err
				}
				if err := add(":"); err != nil {
					return err
				}
				if err := write(field); err != nil {
					return err
				}
			}
			return add("}")
		case reflect.Slice:
			if v.IsNil() {
				return add("null")
			}
			if err := add("["); err != nil {
				return err
			}
			for i := 0; i < v.Len(); i++ {
				if i > 0 {
					if err := add(","); err != nil {
						return err
					}
				}
				if err := write(v.Index(i)); err != nil {
					return err
				}
			}
			return add("]")
		case reflect.Map:
			if v.IsNil() {
				return add("null")
			}
			keys := v.MapKeys()
			sort.Slice(keys, func(i, j int) bool { return keys[i].String() < keys[j].String() })
			if err := add("{"); err != nil {
				return err
			}
			for i, k := range keys {
				if i > 0 {
					if err := add(","); err != nil {
						return err
					}
				}
				if err := quoted(k.String()); err != nil {
					return err
				}
				if err := add(":"); err != nil {
					return err
				}
				if err := write(v.MapIndex(k)); err != nil {
					return err
				}
			}
			return add("}")
		case reflect.String:
			return quoted(v.String())
		case reflect.Bool:
			return add(strconv.FormatBool(v.Bool()))
		case reflect.Int, reflect.Int64, reflect.Int32:
			return add(strconv.FormatInt(v.Int(), 10))
		case reflect.Uint, reflect.Uint64, reflect.Uint32:
			return add(strconv.FormatUint(v.Uint(), 10))
		case reflect.Float64:
			return add(strconv.FormatFloat(v.Float(), 'g', -1, 64))
		}
		owner.Ensemble().Logf("unsupported event encoding field")
		return fmt.Errorf("unsupported event encoding field")
	}
	if err := write(reflect.ValueOf(event)); err != nil {
		return nil, err
	}
	return data, nil
}
