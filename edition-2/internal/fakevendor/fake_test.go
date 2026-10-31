package fakevendor

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func post(t *testing.T, url, body string) map[string]any {
	t.Helper()
	resp, err := http.Post(url, "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("non-JSON reply from %s: %s", url, raw)
	}
	return m
}

// One server, three dialects, chosen by the path the client POSTs to. This is
// the property cmd/fakevendor relies on when it tells a chapter which vendor
// to be, and it was asserted in prose before it was asserted here.
func TestRoutesAllThreeDialectsByPath(t *testing.T) {
	s := New([]Reply{{Text: "hi", Usage: Canonical{Input: 1, Output: 1}}})
	defer s.Close()

	cases := []struct{ path, vendor, topKey string }{
		{"/v1/messages", "anthropic", "content"},
		{"/v1/chat/completions", "openai", "choices"},
		{"/v1beta/models/x:generateContent", "gemini", "candidates"},
	}
	for _, c := range cases {
		m := post(t, s.URL()+c.path, `{"messages":[]}`)
		if _, ok := m[c.topKey]; !ok {
			t.Errorf("%s: reply lacks %q, not the %s shape: %v", c.path, c.topKey, c.vendor, m)
		}
	}
	got := s.Requests()
	if len(got) != 3 || got[0].Vendor != "anthropic" || got[1].Vendor != "openai" || got[2].Vendor != "gemini" {
		t.Fatalf("recorded vendors: %+v", got)
	}
}

// New() repeats the last reply when the script runs out — the grader depends
// on that. Options.Cycle wraps instead, and gives each cycled tool call an id
// no earlier request saw, because the reference engine treats a reused id as
// already answered and silently skips the tool.
func TestCycleWrapsAndUniquifiesIDs(t *testing.T) {
	script := []Reply{
		{ToolName: "list_directory", ToolArgs: `{}`, ToolID: "call_a"},
		{Text: "done"},
	}
	var trace bytes.Buffer
	s := NewWithOptions(script, Options{Cycle: true, Trace: &trace})
	defer s.Close()

	ids := map[string]bool{}
	for i := 0; i < 4; i++ {
		m := post(t, s.URL()+"/v1/messages", `{}`)
		blocks, _ := m["content"].([]any)
		for _, b := range blocks {
			bm := b.(map[string]any)
			if bm["type"] == "tool_use" {
				id, _ := bm["id"].(string)
				if ids[id] {
					t.Fatalf("request %d reused tool id %q", i+1, id)
				}
				ids[id] = true
			}
		}
		if i%2 == 1 {
			if len(blocks) != 1 || blocks[0].(map[string]any)["type"] != "text" {
				t.Fatalf("request %d: want the text reply (script wrapped), got %v", i+1, m)
			}
		}
	}
	if len(ids) != 2 {
		t.Fatalf("want 2 distinct tool ids across two cycles, got %d", len(ids))
	}
	if n := strings.Count(trace.String(), "fake: #"); n != 4 {
		t.Fatalf("trace has %d request lines, want 4:\n%s", n, trace.String())
	}

	// Control: without Cycle the fourth request still gets the LAST reply.
	s2 := New(script)
	defer s2.Close()
	for i := 0; i < 4; i++ {
		m := post(t, s2.URL()+"/v1/messages", `{}`)
		if i == 3 {
			blocks, _ := m["content"].([]any)
			if len(blocks) != 1 || blocks[0].(map[string]any)["type"] != "text" {
				t.Fatalf("New(): request 4 should repeat the last reply, got %v", m)
			}
		}
	}
}
