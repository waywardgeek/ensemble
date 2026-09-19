package llm

// Server-Sent Events, the transport all three vendors chose.
//
// SSE is a line protocol with four fields (`event`, `data`, `id`, `retry`), a
// comment form (a line starting with `:`), and one framing rule: a BLANK LINE
// ends an event. Multi-line `data` fields are joined with newlines. That is
// the entire specification that matters here.
//
// One reader serves all three vendors because the framing is identical and
// only the JSON inside `data` differs. Writing this three times is how a
// vendor package starts to grow its own transport.

import (
	"bufio"
	"io"
	"strings"
)

// sseBufSize is the initial read buffer. A single `data` line can be large
// when a model returns a long text block in one frame, so the reader must not
// have a hard token ceiling. bufio.Reader grows as needed; bufio.Scanner
// would fail with ErrTooLong instead, which is why this uses the former.
const sseBufSize = 64 * 1024

// ReadSSE reads Server-Sent Events from r, calling emit for each event.
//
// An event is a (type, data) pair. Blank lines between events are consumed.
// Returns nil on clean EOF or on `data: [DONE]`.
//
// emit is called with the raw `data` bytes, NOT with parsed JSON: this
// function knows about framing and nothing about any vendor's schema. The
// bytes are only valid for the duration of the call, because the buffer is
// reused — an emit that wants to keep them copies them.
func ReadSSE(r io.Reader, emit func(eventType string, data []byte)) error {
	br := bufio.NewReaderSize(r, sseBufSize)

	var (
		eventType string
		data      []byte
		done      bool
	)

	// dispatch ends the event in progress. Called on a blank line, and once
	// more at EOF so that a stream which ends without its final blank line
	// still delivers its last event rather than silently dropping it.
	dispatch := func() {
		if eventType == "" && len(data) == 0 {
			return // nothing accumulated; a run of blank lines is not an event
		}
		// A trailing newline is an artifact of joining `data` fields, not
		// content. The spec strips exactly one.
		payload := data
		if n := len(payload); n > 0 && payload[n-1] == '\n' {
			payload = payload[:n-1]
		}
		// [DONE] is OpenAI's end sentinel. It is framing, not content, so it
		// is consumed here and never shown to a vendor parser.
		if string(payload) == "[DONE]" {
			done = true
		} else {
			emit(eventType, payload)
		}
		eventType, data = "", nil
	}

	for {
		line, err := br.ReadString('\n')

		if len(line) > 0 {
			// Trim the line terminator only. Trailing significance inside a
			// value is the vendor's business.
			l := strings.TrimRight(line, "\r\n")
			switch {
			case l == "":
				dispatch()
				if done {
					return nil
				}
			case strings.HasPrefix(l, ":"):
				// Comment. Vendors send these as keep-alives; ignoring them
				// is required, not optional.
			default:
				field, value := l, ""
				if i := strings.IndexByte(l, ':'); i >= 0 {
					field = l[:i]
					// Exactly one optional leading space is part of the
					// framing. Any further spaces are data.
					value = strings.TrimPrefix(l[i+1:], " ")
				}
				switch field {
				case "event":
					eventType = value
				case "data":
					data = append(data, value...)
					data = append(data, '\n')
				}
				// `id` and `retry` are reconnection machinery. This is a
				// single-shot request; there is nothing to reconnect to.
			}
		}

		if err != nil {
			if err == io.EOF {
				dispatch() // deliver an unterminated final event
				return nil
			}
			return err
		}
	}
}
