package engine

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

// TestSSEFraming isolates transport framing from vendor JSON: field joining,
// keepalives, line endings, EOF and large payloads must not silently lose bytes.
// Consumer failures must also stop reading rather than masquerade as success.
func TestSSEFraming(t *testing.T) {
	long := strings.Repeat("x", 128*1024)
	for _, tc := range []struct {
		name, input string
		want        []string
	}{
		{"multiline", "event: block\ndata: first\ndata: second\n\n", []string{"block:first\nsecond"}},
		{"comments", ": keepalive\nretry: 500\n\ndata: ok\n\n", []string{":ok"}},
		{"crlf", "event: x\r\ndata: value\r\n\r\n", []string{"x:value"}},
		{"eof", "data: tail", []string{":tail"}},
		{"empty data", "data:\n\n", []string{":"}},
		{"reset event", "event: old\n\ndata: fresh\n\n", []string{":fresh"}},
		{"one optional space", "data:  two\n\n", []string{": two"}},
		{"sentinel", "data: [DONE]\n\n", []string{":[DONE]"}},
		{"large", "data: " + long + "\n\n", []string{":" + long}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var got []string
			err := ReadSSE(strings.NewReader(tc.input), func(event string, data []byte) error { got = append(got, event+":"+string(data)); return nil })
			if err != nil || !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("framing error %v; got %d events, expected %d; framing payload differs", err, len(got), len(tc.want))
			}
		})
	}
	sentinel := errors.New("consumer stopped")
	if err := ReadSSE(strings.NewReader("data: first\n\ndata: second\n\n"), func(string, []byte) error { return sentinel }); err != sentinel {
		t.Fatal("consumer failure lost", err)
	}
}
