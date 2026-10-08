package ensemble

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"example.com/ensemble/internal/common"
)

func TestMalformedResponseAppendLeavesAgentUsable(t *testing.T) {
	for _, tc := range []struct {
		name    string
		event   Event
		indexes []int
	}{
		{name: "public missing response", event: Event{Type: "response_ended"}},
		{name: "negative parser index", event: Event{Type: "response_ended", Response: &Response{Parts: []Part{Text("text")}}}, indexes: []int{-1}},
		{name: "large parser index", event: Event{Type: "response_ended", Response: &Response{}}, indexes: []int{100}},
		{name: "parser index names text", event: Event{Type: "response_ended", Response: &Response{Parts: []Part{Text("text")}}}, indexes: []int{0}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			app := New(nil)
			defer app.Close()
			path := filepath.Join(t.TempDir(), "log")
			a, err := app.NewAgent(Config{DisableStreaming: true, LogPath: path, APIKey: "fixture", Model: "fixture"})
			if err != nil {
				t.Fatal(err)
			}
			if err = a.Ephemeral("retained"); err != nil {
				t.Fatal(err)
			}
			before := a.Events()
			oldBytes, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if tc.indexes == nil {
				err = a.Append(tc.event)
			} else {
				err = a.appendPrepared(tc.event, true, true, tc.indexes)
			}
			if err == nil {
				t.Fatal("accepted malformed response")
			}
			afterBytes, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(before, a.Events()) || !bytes.Equal(oldBytes, afterBytes) {
				t.Fatal("invalid append changed accepted history")
			}
			if err = a.Ephemeral("still usable"); err != nil {
				t.Fatal(err)
			}
			if got := len(a.Events()); got != len(before)+1 {
				t.Fatal(got)
			}
		})
	}
}

func TestLogDestinationIsImmutableAndTruthful(t *testing.T) {
	app := New(nil)
	defer app.Close()
	dir := t.TempDir()
	path := filepath.Join(dir, "original.log")
	a, err := app.NewAgent(Config{DisableStreaming: true, LogPath: path, APIKey: "fixture", Model: "fixture"})
	if err != nil {
		t.Fatal(err)
	}
	initial := a.Config()
	changed := initial
	changed.LogPath = filepath.Join(dir, "other.log")
	changed.Model = "new-model"
	if err = a.SetConfig(changed); err == nil {
		t.Fatal("accepted log replacement")
	}
	if !reflect.DeepEqual(initial, a.Config()) {
		t.Fatal("failed update changed config")
	}
	changed = initial
	changed.LogPath = ""
	changed.Model = "new-model"
	if err = a.SetConfig(changed); err != nil {
		t.Fatal(err)
	}
	if a.Config().LogPath != path || a.Config().Model != "new-model" {
		t.Fatal(a.Config())
	}
	if err = a.Ephemeral("same writer"); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(path); err != nil || !bytes.Contains(data, []byte("same writer")) {
		t.Fatal(string(data), err)
	}
	if _, err = os.Stat(filepath.Join(dir, "other.log")); !os.IsNotExist(err) {
		t.Fatal(err)
	}
	replay, err := app.Load(path, Config{DisableStreaming: true, LogPath: "misleading.log"})
	if err != nil {
		t.Fatal(err)
	}
	if replay.Config().LogPath != path {
		t.Fatal(replay.Config().LogPath)
	}
}

// The parser admission interface is deliberately unavailable on public Agent.
var _ common.TurnAgent = turnAgent{}
