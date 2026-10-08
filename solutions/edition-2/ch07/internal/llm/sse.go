package llm

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"unicode/utf8"

	"example.com/ensemble/internal/common"
)

type sseEvent struct{ name, data string }
type sseReader struct {
	parent common.ModelOperation
	reader *bufio.Reader
	first  bool
}

func newSSE(parent common.ModelOperation, r io.Reader) *sseReader {
	return &sseReader{parent: parent, reader: bufio.NewReader(r), first: true}
}
func (s *sseReader) fail(reason string) error { return failure(s.parent.Engine(), "%s", reason) }

// Read physical terminators before dispatch, so split CRLF belongs to one frame.
// EOF cannot supply the blank line required to dispatch a pending event.
func (s *sseReader) next() (sseEvent, error) {
	if s.first {
		s.first = false
		if prefix, _ := s.reader.Peek(3); bytes.Equal(prefix, []byte{0xef, 0xbb, 0xbf}) {
			_, _ = s.reader.Discard(3)
		}
	}
	frame := 0
	name := "message"
	var data []string
	var line []byte
	count := func() error {
		frame++
		if frame > pendingLimit {
			return s.fail("SSE frame exceeds 1 MiB")
		}
		return nil
	}
	for {
		b, err := s.reader.ReadByte()
		if err != nil {
			if err == io.EOF && frame == 0 {
				return sseEvent{}, io.EOF
			}
			if err == io.EOF {
				return sseEvent{}, s.fail("unfinished SSE frame")
			}
			return sseEvent{}, requestFailure(s.parent.Engine(), err, "stream read failed")
		}
		if err = count(); err != nil {
			return sseEvent{}, err
		}
		if b != '\r' && b != '\n' {
			line = append(line, b)
			continue
		}
		if b == '\r' {
			if next, _ := s.reader.Peek(1); len(next) > 0 && next[0] == '\n' {
				_, _ = s.reader.ReadByte()
				if err = count(); err != nil {
					return sseEvent{}, err
				}
			}
		}
		if len(line) == 0 {
			frame = 0
			if len(data) > 0 {
				return sseEvent{name, strings.Join(data, "\n")}, nil
			}
			name = "message"
			continue
		}
		if !utf8.Valid(line) {
			return sseEvent{}, s.fail("invalid UTF-8 SSE field")
		}
		text := string(line)
		line = line[:0]
		field, value, found := strings.Cut(text, ":")
		if !found {
			value = ""
		}
		value = strings.TrimPrefix(value, " ")
		switch field {
		case "event":
			name = value
			if name == "" {
				name = "message"
			}
		case "data":
			data = append(data, value)
		}
	}
}
