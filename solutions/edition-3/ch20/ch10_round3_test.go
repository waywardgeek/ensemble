package ensemble

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestCh10R3ToolArgumentSaveAndRenderPrefix(t *testing.T) {
	var calls atomic.Int32
	peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			fmt.Fprint(w, `{"model":"fixture","content":[{"type":"tool_use","id":"native-read","name":"read_file","input":{"path":"input.txt"}}],"stop_reason":"tool_use","usage":{"input_tokens":1,"output_tokens":1}}`)
		} else {
			fmt.Fprint(w, `{"model":"fixture","content":[{"type":"text","text":"finished"}],"stop_reason":"end_turn","usage":{"input_tokens":1,"output_tokens":1}}`)
		}
	}))
	defer peer.Close()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "input.txt"), []byte("file content"), 0600); err != nil {
		t.Fatal(err)
	}
	cfg := Config{Vendor: Anthropic, Model: "fixture", APIKey: "local-fake", BaseURL: peer.URL, DisableStreaming: true, BaseDir: dir, DataDir: "cr", Tools: []string{"read_file"}, SavePath: filepath.Join(dir, "producer.json")}
	root := New(nil)
	defer func() {
		if err := root.Shutdown(); err != nil {
			t.Error(err)
		}
	}()
	a, err := root.NewAgent(cfg)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err = a.Ask(ctx, "read input.txt"); err != nil {
		t.Fatal(err)
	}
	events := a.Log().Events()
	boundary := -1
	for i, e := range events {
		if e.Type == ToolReturned {
			boundary = i + 1
			break
		}
	}
	if boundary < 0 {
		t.Fatal("no actual tool result")
	}
	for _, tc := range []struct {
		name   string
		events []Event
	}{{"completed", events}, {"tool-result-prefix", events[:boundary]}} {
		t.Run(tc.name, func(t *testing.T) {
			run := New(nil)
			defer run.Shutdown()
			own := cfg
			own.SavePath = filepath.Join(t.TempDir(), "save.json")
			current, err := run.NewAgent(own)
			if err != nil {
				t.Fatal(err)
			}
			var input bytes.Buffer
			enc := json.NewEncoder(&input)
			for _, e := range tc.events {
				if err = enc.Encode(e); err != nil {
					t.Fatal(err)
				}
			}
			if err = current.Log().Load(&input); err != nil {
				t.Fatal(err)
			}
			before, err := current.Engine().Render(current.Log().Context())
			if err != nil {
				t.Fatal(err)
			}
			canonical, err := json.Marshal(current.Log().Events())
			if err != nil {
				t.Fatal(err)
			}
			if err = current.Save(); err != nil {
				t.Fatal(err)
			}
			if err = run.Shutdown(); err != nil {
				t.Fatal(err)
			}
			next := New(nil)
			defer next.Shutdown()
			restored, err := next.NewAgent(own)
			if err != nil {
				t.Fatal(err)
			}
			after, err := restored.Engine().Render(restored.Log().Context())
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(before, after) {
				t.Fatalf("render changed across save/reopen:\n%s\n%s", before, after)
			}
			got, err := json.Marshal(restored.Log().Events())
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, canonical) {
				t.Fatal("canonical tool IDs/arguments/anchors changed")
			}
			raw, err := os.ReadFile(own.SavePath)
			if err != nil {
				t.Fatal(err)
			}
			for _, mutation := range []string{"content", "args", "provenance", "anchor", "sequence"} {
				var s SaveFile
				if err = json.Unmarshal(raw, &s); err != nil {
					t.Fatal(err)
				}
				switch mutation {
				case "content":
					s.Context.Dialogue[0].Parts[0].Text = "changed"
				case "args":
					for i := range s.Context.Dialogue {
						for j := range s.Context.Dialogue[i].Parts {
							p := &s.Context.Dialogue[i].Parts[j]
							if p.Type == "tool_call" {
								p.Args = json.RawMessage(`{"path":"different.txt"}`)
							}
						}
					}
				case "provenance":
					for i := range s.Context.Dialogue {
						for j := range s.Context.Dialogue[i].Parts {
							p := &s.Context.Dialogue[i].Parts[j]
							if p.From != nil {
								p.From.Model = "different"
							}
						}
					}
				case "anchor":
					s.Context.Dialogue[0].Seq = s.AsOf + 1
				case "sequence":
					s.Log[0].Seq = 0
				}
				invalid, err := json.Marshal(s)
				if err != nil {
					t.Fatal(err)
				}
				bad := own
				bad.SavePath = filepath.Join(t.TempDir(), "save.json")
				if err = os.WriteFile(bad.SavePath, invalid, 0600); err != nil {
					t.Fatal(err)
				}
				rejected := New(nil)
				_, err = rejected.NewAgent(bad)
				_ = rejected.Shutdown()
				if err == nil {
					t.Fatalf("accepted %s disagreement", mutation)
				}
			}
		})
	}
}
