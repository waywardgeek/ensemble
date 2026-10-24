package gui

import (
	"strings"
	"testing"
)

func TestSpeechTextFiltering(t *testing.T) {
	tests := []struct{ in, want string }{
		{"Hello\nsoft wrap.\n\nNew paragraph!", "Hello soft wrap.\n\nNew paragraph!"},
		{"# **Heading**\n\nUse `snake_case` and camelCase with HTTPServer.", "Heading\n\nUse snake case and camel Case with HTTP Server."},
		{"Read [the guide](https://example.test/secret) at https://example.test/path.", "Read the guide at ."},
		{"Before.\n\n```go\nsecret := 123\n\nmore()\n```\n\nAfter.", "Before.\n\ncode block\n\nAfter."},
		{"---\n\n... !!! **", ""},
		{"- One\n- Two", "One Two"},
	}
	for _, tt := range tests {
		if got := cleanSpeech(tt.in); got != tt.want {
			t.Errorf("cleanSpeech(%q)=%q; want %q", tt.in, got, tt.want)
		}
	}
}

func TestSpeechStreamMatchesFinalAndFlushesOnce(t *testing.T) {
	texts := []string{
		"First sentence. Second sentence!\n\nFinal suffix without punctuation",
		"**Bold** and `some_value`. Read [this link](https://example.test/a.b).\n\nEnd.",
		"Before.\n\n```go\nnot spoken.\n\nnope()\n```\n\nAfter.",
		"Acronyms HTTPServer and fooBar survive.\n\nUnicode café and 世界.",
	}
	for _, text := range texts {
		for step := 1; step <= len(text); step++ {
			var filter speechText
			var chunks []string
			for n := step; n < len(text); n += step {
				if out := filter.update(text[:n], false); out != "" {
					chunks = append(chunks, out)
				}
			}
			if out := filter.update(text, true); out != "" {
				chunks = append(chunks, out)
			}
			got := strings.Join(chunks, "\n\n")
			want := cleanSpeech(text)
			if got != want {
				t.Fatalf("step %d: stream %q; final %q", step, got, want)
			}
			if again := filter.update(text, true); again != "" {
				t.Fatalf("repeated final suffix: %q", again)
			}
		}
	}
}

func TestSpeechStableSentenceAndIncompleteMarkup(t *testing.T) {
	var filter speechText
	if got := filter.update("First sentence. A [link](https://example.test/a. ", false); got != "First sentence." {
		t.Fatalf("stable sentence: %q", got)
	}
	if got := filter.update("First sentence. A [link](https://example.test/a. b) remains.", true); got != "A link remains." {
		t.Fatalf("tail: %q", got)
	}
}

func TestSpeechSemanticSources(t *testing.T) {
	for _, tt := range []struct {
		card Card
		want string
	}{
		{Card{Kind: "text"}, "text"}, {Card{Kind: "thinking"}, "thinking"},
		{Card{Kind: "tool_call"}, "tool"}, {Card{Kind: "tool_result"}, "tool_result"},
		{Card{Kind: "tool_result", IsError: true}, "error"},
	} {
		if got := speechSource(tt.card); got != tt.want {
			t.Fatalf("source=%q; want %q", got, tt.want)
		}
	}
}

func TestSpeechDispatchAdmissionAndResultSilence(t *testing.T) {
	app, agent := idleAgent(t, true)
	server, err := NewServer(app, "web/gui")
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	hub := server.Hub
	hub.mu.Lock()
	defer hub.mu.Unlock()
	view := hub.view(agent)
	call := Card{Kind: "tool_call", Role: "agent", Name: "file_read", Text: "file_read\n{private arguments}", TurnID: 1}
	hub.narrate(view, 1, call, true)
	if hub.speech.active[agent.ID()] != "" {
		t.Fatal("announced before actual dispatch")
	}
	call.Dispatched = true
	hub.narrate(view, 1, call, true)
	hub.narrate(view, 1, call, true)
	active := hub.speech.tasks[hub.speech.active[agent.ID()]]
	if active == nil || active.text != "file read" || len(active.segments) != 1 || active.segments[0].Source != "tool" {
		t.Fatalf("dispatch speech: %+v", active)
	}
	hub.narrate(view, 2, Card{Kind: "tool_result", Role: "tool", Text: "full readable result", TurnID: 1}, true)
	if hub.speech.pending[agent.ID()] != nil {
		t.Fatal("tool result was automatically spoken")
	}
	hub.narrate(view, 3, Card{Kind: "thinking", Role: "agent", Text: "First thought.", TurnID: 1}, true)
	hub.narrate(view, 4, Card{Kind: "text", Role: "agent", Text: "Then answer.", TurnID: 1}, true)
	pending := hub.speech.pending[agent.ID()]
	if pending == nil || len(pending.segments) != 2 || pending.segments[0].Source != "thinking" || pending.segments[1].Source != "text" {
		t.Fatalf("semantic queue: %+v", pending)
	}
}
