package engine

import (
	"bufio"
	"io"
	"strings"
)

// ReadSSE yields complete data fields without imposing Scanner's token limit.
// Event names describe vendor meaning; framing itself is shared by all parsers.
// Comments and unknown fields are ignored, and a final event survives missing
// trailing blank lines. The callback's bytes belong to that invocation.
func ReadSSE(r io.Reader, emit func(string, []byte) error) error {
	reader := bufio.NewReader(r)
	var event string
	var data []string
	flush := func() error {
		if len(data) == 0 {
			event = ""
			return nil
		}
		err := emit(event, []byte(strings.Join(data, "\n")))
		event, data = "", nil
		return err
	}
	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			line = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
			if line == "" {
				if e := flush(); e != nil {
					return e
				}
			} else if !strings.HasPrefix(line, ":") {
				field, value, _ := strings.Cut(line, ":")
				value = strings.TrimPrefix(value, " ")
				switch field {
				case "event":
					event = value
				case "data":
					data = append(data, value)
				}
			}
		}
		if err == io.EOF {
			return flush()
		}
		if err != nil {
			return err
		}
	}
}
