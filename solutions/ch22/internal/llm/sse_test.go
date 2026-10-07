package llm

import (
	"errors"
	"io"
	"strings"
	"testing"
)

type sseEvent struct {
	Type string
	Data string
}

func collectSSE(t *testing.T, in string) ([]sseEvent, error) {
	t.Helper()
	var got []sseEvent
	err := ReadSSE(strings.NewReader(in), func(eventType string, data []byte) {
		// Copy: ReadSSE documents the bytes as valid only during the call.
		got = append(got, sseEvent{Type: eventType, Data: string(data)})
	})
	return got, err
}

func TestReadSSESplitsOnBlankLines(t *testing.T) {
	got, err := collectSSE(t, "event: a\ndata: one\n\nevent: b\ndata: two\n\n")
	if err != nil {
		t.Fatalf("ReadSSE: %v", err)
	}
	want := []sseEvent{{"a", "one"}, {"b", "two"}}
	if len(got) != len(want) {
		t.Fatalf("got %d events %+v, want %d", len(got), got, len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("event %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

// A vendor that sends no `event:` field at all (OpenAI, Gemini) must still
// produce events. If this regressed, both of those parsers would go silent
// while Anthropic kept working.
func TestReadSSEDataOnlyEventsHaveEmptyType(t *testing.T) {
	got, err := collectSSE(t, "data: {\"x\":1}\n\ndata: {\"x\":2}\n\n")
	if err != nil {
		t.Fatalf("ReadSSE: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d events %+v, want 2", len(got), got)
	}
	for i, ev := range got {
		if ev.Type != "" {
			t.Errorf("event %d type = %q, want empty", i, ev.Type)
		}
	}
	if got[0].Data != `{"x":1}` || got[1].Data != `{"x":2}` {
		t.Errorf("payloads = %q, %q", got[0].Data, got[1].Data)
	}
}

// Multi-line data joins with newlines and drops exactly one trailing newline.
func TestReadSSEJoinsMultiLineData(t *testing.T) {
	got, err := collectSSE(t, "data: line1\ndata: line2\ndata: line3\n\n")
	if err != nil {
		t.Fatalf("ReadSSE: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d events %+v, want 1", len(got), got)
	}
	if got[0].Data != "line1\nline2\nline3" {
		t.Errorf("data = %q, want %q", got[0].Data, "line1\nline2\nline3")
	}
}

// An empty `data:` line is content (an empty string), not an absence. It must
// still frame an event.
func TestReadSSEEmptyDataStillEmits(t *testing.T) {
	got, err := collectSSE(t, "data:\n\n")
	if err != nil {
		t.Fatalf("ReadSSE: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d events %+v, want 1", len(got), got)
	}
	if got[0].Data != "" {
		t.Errorf("data = %q, want empty", got[0].Data)
	}
}

// Keep-alive comments are mandatory to ignore. A parser that treats them as
// data hands a vendor parser garbage that is not JSON.
func TestReadSSEIgnoresComments(t *testing.T) {
	got, err := collectSSE(t, ": ping\n\ndata: real\n\n: ping\n\n")
	if err != nil {
		t.Fatalf("ReadSSE: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d events %+v, want 1", len(got), got)
	}
	if got[0].Data != "real" {
		t.Errorf("data = %q, want %q", got[0].Data, "real")
	}
}

// Exactly one leading space after the colon is framing. A second space is
// data, and a parser that trims all of them corrupts indented content.
func TestReadSSETrimsExactlyOneLeadingSpace(t *testing.T) {
	got, err := collectSSE(t, "data:   three spaces\n\n")
	if err != nil {
		t.Fatalf("ReadSSE: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d events, want 1", len(got))
	}
	if got[0].Data != "  three spaces" {
		t.Errorf("data = %q, want %q", got[0].Data, "  three spaces")
	}
}

func TestReadSSEHandlesCRLF(t *testing.T) {
	got, err := collectSSE(t, "event: a\r\ndata: one\r\n\r\n")
	if err != nil {
		t.Fatalf("ReadSSE: %v", err)
	}
	if len(got) != 1 || got[0] != (sseEvent{"a", "one"}) {
		t.Errorf("got %+v, want one event {a one}", got)
	}
}

// [DONE] is framing, not content. It must be consumed and must stop the read.
func TestReadSSEStopsAtDONEAndHidesIt(t *testing.T) {
	got, err := collectSSE(t, "data: real\n\ndata: [DONE]\n\ndata: after\n\n")
	if err != nil {
		t.Fatalf("ReadSSE: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d events %+v, want 1 (DONE consumed, nothing after)", len(got), got)
	}
	if got[0].Data != "real" {
		t.Errorf("data = %q, want %q", got[0].Data, "real")
	}
}

// A stream that ends without its final blank line still has one event in it.
// Dropping it loses the last chunk of every truncated response.
func TestReadSSEDeliversUnterminatedFinalEvent(t *testing.T) {
	got, err := collectSSE(t, "data: first\n\ndata: last")
	if err != nil {
		t.Fatalf("ReadSSE: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d events %+v, want 2", len(got), got)
	}
	if got[1].Data != "last" {
		t.Errorf("final data = %q, want %q", got[1].Data, "last")
	}
}

func TestReadSSERunsOfBlankLinesAreNotEvents(t *testing.T) {
	got, err := collectSSE(t, "\n\n\ndata: only\n\n\n\n")
	if err != nil {
		t.Fatalf("ReadSSE: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d events %+v, want 1", len(got), got)
	}
}

// `id` and `retry` are reconnection machinery and must not be mistaken for
// content, but they also must not suppress the event they accompany.
func TestReadSSEIgnoresIDAndRetryFields(t *testing.T) {
	got, err := collectSSE(t, "id: 7\nretry: 1000\ndata: payload\n\n")
	if err != nil {
		t.Fatalf("ReadSSE: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d events %+v, want 1", len(got), got)
	}
	if got[0].Data != "payload" {
		t.Errorf("data = %q, want %q", got[0].Data, "payload")
	}
}

// A line with no colon at all is a bare field name with an empty value. It is
// legal SSE and must not panic or be read as data.
func TestReadSSEBareFieldNameIsNotData(t *testing.T) {
	got, err := collectSSE(t, "data\n\n")
	if err != nil {
		t.Fatalf("ReadSSE: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d events %+v, want 1", len(got), got)
	}
	if got[0].Data != "" {
		t.Errorf("data = %q, want empty", got[0].Data)
	}
}

type errAfter struct {
	r io.Reader
}

func (e errAfter) Read(p []byte) (int, error) {
	n, err := e.r.Read(p)
	if err == io.EOF {
		return n, errors.New("connection reset")
	}
	return n, err
}

// A mid-stream transport error is a LOUD failure. Returning nil here would
// turn a truncated response into a complete-looking one.
func TestReadSSEPropagatesReadError(t *testing.T) {
	var got []sseEvent
	err := ReadSSE(errAfter{strings.NewReader("data: partial\n\n")}, func(et string, d []byte) {
		got = append(got, sseEvent{et, string(d)})
	})
	if err == nil {
		t.Fatal("ReadSSE returned nil on a transport error; a truncated stream must not look clean")
	}
	if len(got) != 1 {
		t.Errorf("got %d events, want the 1 complete event before the error", len(got))
	}
}

// A long single data line must not hit a token ceiling. bufio.Scanner would
// fail here at 64KB; bufio.Reader grows.
func TestReadSSEHandlesVeryLongDataLine(t *testing.T) {
	big := strings.Repeat("x", 512*1024)
	got, err := collectSSE(t, "data: "+big+"\n\n")
	if err != nil {
		t.Fatalf("ReadSSE: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d events, want 1", len(got))
	}
	if len(got[0].Data) != len(big) {
		t.Errorf("data length = %d, want %d", len(got[0].Data), len(big))
	}
}
