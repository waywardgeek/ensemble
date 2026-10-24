package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestLargeJSONLine protects valid questions larger than Scanner's default physical-line limit.
func TestLargeJSONLine(t *testing.T) {
	t.Setenv("ENSEMBLE_SAVE", filepath.Join(t.TempDir(), "save.json"))
	question := strings.Repeat("x", 70*1024)
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		var body struct {
			Messages []struct{ Content string }
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("request JSON: %v", err)
		} else if len(body.Messages) != 1 || body.Messages[0].Content != question {
			t.Error("large question was truncated or changed")
		}
		fmt.Fprint(w, `{"content":[{"type":"text","text":"received"}],"usage":{"input_tokens":7,"output_tokens":1}}`)
	}))
	defer server.Close()
	t.Setenv("ANTHROPIC_BASE_URL", server.URL)
	t.Setenv("ANTHROPIC_API_KEY", "fake-key")
	t.Setenv("ANTHROPIC_MODEL", "fake-model")
	line, err := json.Marshal(map[string]string{"user": question})
	if err != nil {
		t.Fatal(err)
	}
	var out, diagnostics bytes.Buffer
	err = run(context.Background(), nil, bytes.NewReader(append(line, '\n')), &out, &diagnostics)
	if err != nil {
		t.Fatalf("well-formed large JSON line: %v; calls=%d; stdout=%q", err, calls.Load(), out.String())
	}
	if calls.Load() != 1 || out.String() != "{\"assistant\":\"received\"}\n{\"usage\":{\"input\":7,\"output\":1,\"cache_write\":0,\"cache_read\":0}}\n" || diagnostics.Len() != 0 {
		t.Fatalf("calls=%d stdout=%q stderr=%q", calls.Load(), out.String(), diagnostics.String())
	}
}

// brokenReader is an input peer that fails, not a substitute for the CLI logic.
type brokenReader struct{}

func (brokenReader) Read([]byte) (int, error) { return 0, errors.New("input-secret") }

// TestInputLines distinguishes clean EOF, final unterminated lines, and incomplete input caused by read errors.
func TestInputLines(t *testing.T) {
	const line = "{\"user\":\"hello\"}"
	const answer = "{\"assistant\":\"received\"}\n"
	for _, tc := range []struct {
		name            string
		input           io.Reader
		calls           int32
		want, wantError string
	}{
		{"ordinary", strings.NewReader(line + "\n" + line + "\r\n"), 2, answer + answer + "{\"usage\":{\"input\":14,\"output\":2,\"cache_write\":0,\"cache_read\":0}}\n", ""},
		{"final without newline", strings.NewReader(line), 1, answer + "{\"usage\":{\"input\":7,\"output\":1,\"cache_write\":0,\"cache_read\":0}}\n", ""},
		{"EOF", strings.NewReader(""), 0, "{\"usage\":{\"input\":0,\"output\":0,\"cache_write\":0,\"cache_read\":0}}\n", ""},
		{"read error", brokenReader{}, 0, "", "reading input failed"},
		{"error after complete line", io.MultiReader(strings.NewReader(line+"\n"), brokenReader{}), 1, answer, "reading input failed"},
		{"error in partial line", io.MultiReader(strings.NewReader(line), brokenReader{}), 0, "", "reading input failed"},
		{"invalid JSON", strings.NewReader("not-json\n"), 0, "", "invalid input JSON"},
		{"missing user", strings.NewReader("{}\n"), 0, "", "user message is empty"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("ENSEMBLE_SAVE", filepath.Join(t.TempDir(), "save.json"))
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				fmt.Fprint(w, `{"content":[{"type":"text","text":"received"}],"usage":{"input_tokens":7,"output_tokens":1}}`)
			}))
			defer server.Close()
			t.Setenv("ANTHROPIC_BASE_URL", server.URL)
			t.Setenv("ANTHROPIC_API_KEY", "fake-key")
			t.Setenv("ANTHROPIC_MODEL", "fake-model")
			var out, diagnostics bytes.Buffer
			err := run(context.Background(), nil, tc.input, &out, &diagnostics)
			if (tc.wantError == "" && err != nil) || (tc.wantError != "" && (err == nil || err.Error() != tc.wantError)) || calls.Load() != tc.calls || out.String() != tc.want || diagnostics.Len() != 0 {
				t.Fatalf("err=%v calls=%d stdout=%q stderr=%q", err, calls.Load(), out.String(), diagnostics.String())
			}
		})
	}
}

// completedChatOutput exposes the real CLI private-reply boundary to the feeder.
type completedChatOutput struct {
	*bytes.Buffer
	first chan struct{}
	once  sync.Once
}

func (w *completedChatOutput) Write(p []byte) (int, error) {
	n, err := w.Buffer.Write(p)
	if strings.Contains(string(p), "assistant> received") {
		w.once.Do(func() { close(w.first) })
	}
	return n, err
}

// TestChatAndConfiguration checks opt-in human output and startup rejection without paid calls.
func TestChatAndConfiguration(t *testing.T) {
	t.Setenv("ENSEMBLE_SAVE", filepath.Join(t.TempDir(), "save.json"))
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		fmt.Fprint(w, `{"content":[{"type":"text","text":"received"}],"usage":{"input_tokens":7,"output_tokens":1}}`)
	}))
	defer server.Close()
	t.Setenv("ANTHROPIC_BASE_URL", server.URL)
	t.Setenv("ANTHROPIC_API_KEY", "fake-key")
	t.Setenv("ANTHROPIC_MODEL", "fake-model")
	var out, diagnostics bytes.Buffer
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	input, feed := io.Pipe()
	defer input.Close()
	defer feed.Close()
	printed := &completedChatOutput{Buffer: &out, first: make(chan struct{})}
	fed := make(chan error, 1)
	go func() {
		_, err := io.WriteString(feed, "\nhello\n")
		if err == nil {
			select {
			case <-printed.first:
			case <-ctx.Done():
				err = ctx.Err()
			}
		}
		if err == nil {
			_, err = io.WriteString(feed, "again")
		}
		feed.CloseWithError(err)
		fed <- err
	}()
	if err := run(ctx, []string{"chat"}, input, printed, &diagnostics); err != nil {
		t.Fatal(err)
	}
	if err := <-fed; err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 2 || strings.Count(out.String(), "assistant> received") != 2 || !strings.Contains(diagnostics.String(), "[4 turns | 14 input tokens | 2 output tokens]") || !strings.Contains(diagnostics.String(), "you> ") {
		t.Fatalf("calls=%d stdout=%q stderr=%q", calls.Load(), out.String(), diagnostics.String())
	}
	out.Reset()
	if err := run(context.Background(), []string{"unknown"}, strings.NewReader(""), &out, &diagnostics); err == nil {
		t.Fatal("unknown argument accepted")
	}
	t.Setenv("ANTHROPIC_MODEL", "")
	if err := run(context.Background(), nil, strings.NewReader(""), &out, &diagnostics); err == nil || !strings.Contains(err.Error(), "GET /v1/models") || out.Len() != 0 || calls.Load() != 2 {
		t.Fatal("missing model was not rejected before HTTP")
	}
}
