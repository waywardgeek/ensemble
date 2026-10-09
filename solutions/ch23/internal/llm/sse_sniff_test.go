package llm

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func respWith(contentType, body string) *http.Response {
	h := http.Header{}
	if contentType != "" {
		h.Set("Content-Type", contentType)
	}
	return &http.Response{Header: h, Body: io.NopCloser(strings.NewReader(body))}
}

// The ChatGPT-plan route streams with no Content-Type at all. The body's
// framing must decide, and deciding must not consume what the parser needs.
func TestIsSSEWithoutContentType(t *testing.T) {
	cases := []struct {
		name, ct, body string
		want           bool
	}{
		{"plan-route stream, no header", "", "event: response.created\ndata: {\"type\":\"response.created\"}\n\n", true},
		{"error document, no header", "", "  \n{\"error\":{\"message\":\"nope\"}}", false},
		{"header says JSON", "application/json", "event: looks like sse\n", false},
		{"header says SSE", "text/event-stream; charset=utf-8", "{}", true},
	}
	for _, c := range cases {
		resp := respWith(c.ct, c.body)
		if got := isSSE(resp); got != c.want {
			t.Errorf("%s: isSSE = %v, want %v", c.name, got, c.want)
		}
		rest, _ := io.ReadAll(resp.Body)
		if string(rest) != c.body {
			t.Errorf("%s: body after isSSE = %q, want it untouched", c.name, rest)
		}
	}
}
