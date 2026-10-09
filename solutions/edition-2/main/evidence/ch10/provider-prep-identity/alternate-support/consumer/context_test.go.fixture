package main

import (
	"encoding/json"
	"example.com/ensemble"
	"io"
	"testing"
)

func TestContextComparisonKeepsReplayBytes(t *testing.T) {
	app := ensemble.New(io.Discard)
	defer app.Close()
	a, err := app.OpenSession(ensemble.SessionOptions{Config: ensemble.Config{Workspace: t.TempDir(), DataDir: "session", Model: "fixture", APIKey: "local", BaseURL: "http://127.0.0.1:1", Builtins: []string{"read_file"}}})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	left, right := a.Snapshot(), a.Snapshot()
	left.Session.Identity.Handlers[0].Schema = json.RawMessage(`{"type":"object","properties":{}}`)
	right.Session.Identity.Handlers[0].Schema = json.RawMessage(`{ "properties": {}, "type": "object" }`)
	if !sameContext(left, right) {
		t.Fatal("identity schema formatting affected comparison")
	}
	left.Entries = []ensemble.Entry{{Parts: []ensemble.Part{{Type: "tool_call", Args: json.RawMessage(`{"n":1.0}`)}}}}
	right.Entries = []ensemble.Entry{{Parts: []ensemble.Part{{Type: "tool_call", Args: json.RawMessage(`{"n":1.0}`)}}}}
	if !sameContext(left, right) {
		t.Fatal("valid raw parent")
	}
	right.Entries[0].Parts[0].Args = json.RawMessage(`{"n":1}`)
	if sameContext(left, right) {
		t.Fatal("replay number lexeme normalized")
	}
	right.Entries[0].Parts[0].Args = json.RawMessage(`{ "n":1.0}`)
	if sameContext(left, right) {
		t.Fatal("replay whitespace normalized")
	}
	right.Entries[0].Parts[0].Args = json.RawMessage(`{"n":1.0}`)
	right.LastSeq++
	if sameContext(left, right) {
		t.Fatal("structural state change ignored")
	}
}
