package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"example.com/ensemble"
)

func TestChatLineBoundary(t *testing.T) {
	owner := ensemble.New(nil)
	for _, ending := range []string{"\n", "\r\n", ""} {
		for _, size := range []int{chatLineLimit - 1, chatLineLimit, chatLineLimit + 1} {
			input := strings.Repeat("x", size)
			line, err := readChatLine(owner, bufio.NewReader(strings.NewReader(input+ending)))
			if size <= chatLineLimit && (err != nil || line != input) {
				t.Fatalf("size %d ending %q: %v", size, ending, err)
			}
			if size > chatLineLimit && err == nil {
				t.Fatal("accepted oversized line")
			}
		}
	}
	for _, input := range []string{"\xff\n", strings.Repeat("é", chatLineLimit/2) + "x\n"} {
		if _, err := readChatLine(owner, bufio.NewReader(strings.NewReader(input))); err == nil {
			t.Fatal("accepted invalid byte input")
		}
	}
}

func cliConfig(t *testing.T, url string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "session.log")
	for k, v := range map[string]string{"LLM_VENDOR": "anthropic", "LLM_MODEL": "fixture", "LLM_API_KEY": "secret-marker", "LLM_BASE_URL": url, "LLM_RESOLVED_MODEL": "", "CH02_LOG": path} {
		t.Setenv(k, v)
	}
	return path
}

func TestChatAndProtocolUseSamePublicPath(t *testing.T) {
	var bodies []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		bodies = append(bodies, string(body))
		fmt.Fprint(w, `{"content":[{"type":"text","text":"first\nsecond"}],"usage":{"input_tokens":7,"cache_creation_input_tokens":2,"cache_read_input_tokens":3,"output_tokens":5}}`)
	}))
	defer server.Close()
	chatPath := cliConfig(t, server.URL)
	var chat, diag bytes.Buffer
	input := " \t\n/help\n/unknown\n/ephemeral\n/redact nope 1 reason\n/usage\n/history\n/ephemeral one request\n  hello  \n//help\n"
	if err := runArgs([]string{"chat"}, strings.NewReader(input), &chat, &diag); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"You> ", "Assistant:\nfirst\nsecond\n", "No tool_returned", "Command error", "Final usage: input=14, cache write=4, cache read=6, output=10"} {
		if !strings.Contains(chat.String(), want) {
			t.Fatalf("missing %q in %s", want, chat.String())
		}
	}
	if len(bodies) != 2 || !strings.Contains(bodies[0], "  hello  ") || !strings.Contains(bodies[0], "one request") || strings.Contains(bodies[1], "one request") || !strings.Contains(bodies[1], "/help") {
		t.Fatal("chat changed prompt text or directive delivery")
	}
	data, _ := os.ReadFile(chatPath)
	if strings.Count(string(data), `"seq":`) != 11 || bytes.Contains(data, []byte("secret-marker")) {
		t.Fatal("local commands mutated the log or secret leaked")
	}
	chatBodies := append([]string(nil), bodies...)
	for _, mode := range [][]string{nil, {"protocol"}} {
		cliConfig(t, server.URL)
		bodies = nil
		var out bytes.Buffer
		if err := runArgs(mode, strings.NewReader("{\"ephemeral\":\"one request\"}\n{\"user\":\"  hello  \"}\n{\"user\":\"/help\"}\n"), &out, &diag); err != nil {
			t.Fatal(err)
		}
		want := "{\"ack\":\"ephemeral\"}\n{\"assistant\":\"first\\nsecond\"}\n{\"assistant\":\"first\\nsecond\"}\n{\"usage\":{\"input\":14,\"cache_write\":4,\"cache_read\":6,\"output\":10}}\n"
		if out.String() != want || len(bodies) != 2 || bodies[0] != chatBodies[0] || bodies[1] != chatBodies[1] {
			t.Fatal("protocol output or shared requests changed", out.String())
		}
	}
}

func TestChatEmptyAndToolContinuation(t *testing.T) {
	for _, withCall := range []bool{false, true} {
		requests := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests++
			parts := `[{"type":"text","text":""}]`
			if withCall && requests == 1 {
				parts = `[{"type":"text","text":"intermediate narration"},{"type":"tool_use","id":"call-1","name":"unavailable","input":{}}]`
			}
			if requests == 2 {
				body, _ := io.ReadAll(r.Body)
				if !bytes.Contains(body, []byte(`"tool_use_id":"call-1"`)) || !bytes.Contains(body, []byte(`"is_error":true`)) {
					t.Error("continuation did not preserve ordinary tool failure and pairing")
				}
			}
			fmt.Fprintf(w, `{"content":%s,"usage":{"input_tokens":1,"output_tokens":2}}`, parts)
		}))
		cliConfig(t, server.URL)
		var out, diag bytes.Buffer
		err := runArgs([]string{"chat"}, strings.NewReader("hello\n"), &out, &diag)
		server.Close()
		wantRequests := 1
		if withCall {
			wantRequests = 2
		}
		if err != nil || requests != wantRequests || strings.Count(out.String(), "Assistant:") != 1 || !strings.Contains(out.String(), "[No text returned]") || strings.Contains(out.String(), "intermediate narration") || !strings.Contains(out.String(), fmt.Sprintf("Final usage: input=%d", wantRequests)) {
			t.Fatal(err, out.String())
		}
	}
}

func TestChatFailuresAndEmptyExit(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(503)
		fmt.Fprint(w, "secret-marker")
	}))
	defer server.Close()
	for _, input := range []string{"\xff\n", strings.Repeat("x", chatLineLimit+1) + "\n", "/redact 1 1 reason\n", "hello\n"} {
		path := cliConfig(t, server.URL)
		var out, diag bytes.Buffer
		err := runArgs([]string{"chat"}, strings.NewReader(input), &out, &diag)
		if err == nil || strings.Contains(out.String(), "Final usage") || strings.Contains(err.Error()+diag.String(), "secret-marker") {
			t.Fatal("unsafe or successful failure", err)
		}
		data, _ := os.ReadFile(path)
		if input != "hello\n" && string(data) != "{\"log_version\":1}\n" {
			t.Fatal("rejected input mutated history")
		}
	}
	if calls != 1 {
		t.Fatal("invalid input sent HTTP or failure retried")
	}
	for _, input := range []string{"", "/quit\n"} {
		cliConfig(t, server.URL)
		var out, diag bytes.Buffer
		if err := runArgs([]string{"chat"}, strings.NewReader(input), &out, &diag); err != nil || !strings.Contains(out.String(), "Final usage: input=0, cache write=0, cache read=0, output=0") {
			t.Fatal(err, out.String())
		}
	}
	cliConfig(t, server.URL)
	t.Setenv("LLM_MODEL", "")
	t.Setenv("ANTHROPIC_MODEL", "")
	var out, diag bytes.Buffer
	if err := runArgs([]string{"chat"}, strings.NewReader(""), &out, &diag); err == nil || out.Len() != 0 {
		t.Fatal("configuration failed after banner")
	}
}

func TestChatHistoryAndRedactionOfControlledResult(t *testing.T) {
	owner := ensemble.New(nil)
	agent, err := owner.NewAgent(ensemble.Config{Model: "fixture", APIKey: "test", LogPath: filepath.Join(t.TempDir(), "log")})
	if err != nil {
		t.Fatal(err)
	}
	defer agent.Close()
	p := ensemble.Provenance{Vendor: "anthropic", Model: "fixture", Surface: "messages"}
	for _, event := range []ensemble.Event{
		{Type: "message_received", Message: &ensemble.Entry{Actor: "human", Purpose: "dialogue", Parts: []ensemble.Part{ensemble.Text("inspect")}}},
		{Type: "response_ended", Response: &ensemble.Response{From: p, Usage: &ensemble.Usage{}, Parts: []ensemble.Part{{Type: "tool_call", CallID: "call-visible", Name: "inspect", From: &p, Args: json.RawMessage(`{}`)}}}},
		{Type: "tool_returned", Tool: &ensemble.ToolEvent{CallID: "call-visible", Parts: []ensemble.Part{ensemble.Text("result-marker")}}},
	} {
		if err := agent.Append(event); err != nil {
			t.Fatal(err)
		}
	}
	var out bytes.Buffer
	if err := runChat(owner, agent, strings.NewReader("/history\n/redact 3 3 deliberate control\n/history\n/quit\n"), &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "3 tool_returned call_id=call-visible") || !strings.Contains(out.String(), "4 redacted") || strings.Contains(out.String(), "result-marker") {
		t.Fatal("history did not expose safe target metadata", out.String())
	}
	body, err := agent.Render(agent.Config())
	if err != nil || bytes.Contains(body, []byte("result-marker")) || !bytes.Contains(body, []byte("[redacted]")) {
		t.Fatal("chat redaction did not use public reducer", err)
	}
}
