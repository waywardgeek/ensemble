package ensemble_test

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ensemble/ensemble"
)

// A real captured vocabulary includes tool identities, opaque material, a
// redaction and pending ephemera. Persistence must preserve the renderer's
// inputs, rather than teaching a second reducer how to reconstruct them.
func TestSaveSnapshotReplayAndTail(t *testing.T) {
	prefix := conversation + firstResult + lastResult
	tail := `{"seq":6,"type":"redacted","redact":{"from":4,"to":4,"level":"redact_result"}}
{"seq":7,"type":"message_received","message":{"actor":"system","parts":[{"type":"text","text":"pending-once"}]}}
`
	for _, vendor := range []ensemble.Vendor{ensemble.Anthropic, ensemble.OpenAI, ensemble.Gemini} {
		t.Run(stringVendor(vendor), func(t *testing.T) {
			early := loaded(t, vendor, prefix)
			full := loaded(t, vendor, prefix+tail)
			var encoded bytes.Buffer
			if err := full.History().Save(&encoded); err != nil {
				t.Fatal(err)
			}
			var saved map[string]json.RawMessage
			if err := json.Unmarshal(encoded.Bytes(), &saved); err != nil {
				t.Fatal(err)
			}
			earlyContext, _ := json.Marshal(early.History().Context())
			want, err := full.Engine().Render(full.History().Context())
			if err != nil {
				t.Fatal(err)
			}
			for _, variant := range []string{"snapshot", "log-only", "tail", "snapshot-only"} {
				t.Run(variant, func(t *testing.T) {
					fields := map[string]json.RawMessage{}
					for k, v := range saved {
						fields[k] = v
					}
					switch variant {
					case "log-only":
						fields["context"] = json.RawMessage("null")
					case "tail":
						fields["context"], fields["as_of"] = earlyContext, json.RawMessage("5")
					case "snapshot-only":
						fields["log"] = json.RawMessage("[]")
					}
					data, _ := json.Marshal(fields)
					restored := loaded(t, vendor, "")
					if err := restored.History().Restore(bytes.NewReader(data)); err != nil {
						t.Fatal(err)
					}
					got, err := restored.Engine().Render(restored.History().Context())
					if err != nil || !bytes.Equal(want, got) {
						t.Fatalf("restored request differs: %s %v", got, err)
					}
					if err := restored.History().Append(ensemble.Event{Type: "message_received", Message: &ensemble.MessageData{Actor: ensemble.Human, Parts: []ensemble.Part{{Type: "text", Text: "next"}}}}); err != nil {
						t.Fatal(err)
					}
					events := restored.History().EventsAfter(7)
					if len(events) != 1 || events[0].Seq != 8 {
						t.Fatalf("sequence did not continue from anchor/tail: %+v", events)
					}
					var next bytes.Buffer
					if err := restored.History().Save(&next); err != nil {
						t.Fatal(err)
					}
					var state struct {
						// AsOf must include the first new fact after restoration.
						AsOf uint64 `json:"as_of"`
						// Log must preserve loaded facts before newly appended ones.
						Log []ensemble.Event `json:"log"`
					}
					if err := json.Unmarshal(next.Bytes(), &state); err != nil {
						t.Fatal(err)
					}
					wantCount := 8
					if variant == "snapshot-only" {
						wantCount = 1
					}
					if state.AsOf != 8 || len(state.Log) != wantCount {
						t.Fatalf("save lost anchor or original log: %s", next.Bytes())
					}
				})
			}
		})
	}
}

func stringVendor(v ensemble.Vendor) string { b, _ := json.Marshal(v); return string(b) }

// Configuration is a record only, and separate Agent directories must not
// acquire each other's conversations. These are actual public library clients.
func TestSaveOwnershipAndRuntimeConfig(t *testing.T) {
	root := ensemble.New(io.Discard)
	makeAgent := func(dir, override, model string) ensemble.Agent {
		t.Helper()
		a, err := root.NewAgent(ensemble.Config{DataDir: dir, SavePath: override, Model: model, Vendor: ensemble.OpenAI, SystemPrompt: "current constitution", APIKey: "must-not-save-key", BaseURL: "https://must-not-save-endpoint.invalid"})
		if err != nil {
			t.Fatal(err)
		}
		return a
	}
	dir := t.TempDir()
	a := makeAgent(filepath.Join(dir, "a"), "", "first")
	if err := a.Load(); err != nil {
		t.Fatal(err)
	}
	if err := a.History().Load(strings.NewReader(conversation)); err != nil {
		t.Fatal(err)
	}
	if err := a.Save(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "a", "save.json"))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(data, []byte("must-not-save")) {
		t.Fatal("transport secret or endpoint entered save")
	}
	b := makeAgent(filepath.Join(dir, "b"), "", "second")
	if err := b.Load(); err != nil {
		t.Fatal(err)
	}
	if len(b.History().Context().Dialogue) != 0 {
		t.Fatal("Agent directories shared history")
	}
	resumed := makeAgent(filepath.Join(dir, "unused"), filepath.Join(dir, "a", "save.json"), "current-model")
	if err := resumed.Load(); err != nil {
		t.Fatal(err)
	}
	body, err := resumed.Engine().Render(resumed.History().Context())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(body, []byte("current-model")) || !bytes.Contains(body, []byte("current constitution")) || !bytes.Contains(body, []byte("hello")) {
		t.Fatalf("historical config replaced runtime config or conversation lost: %s", body)
	}
}

// Refusing a malformed save must be transactional even for library users who
// already have a projection. JSON syntax and interpreted log errors both matter.
func TestBadSaveLeavesHistory(t *testing.T) {
	a := loaded(t, ensemble.OpenAI, conversation)
	var good bytes.Buffer
	if err := a.History().Save(&good); err != nil {
		t.Fatal(err)
	}
	before, _ := json.Marshal(a.History().Context())
	for _, bad := range []string{"{", "null", "{}", good.String() + "{}", strings.Replace(good.String(), "message_received", "unknown", 1)} {
		if err := a.History().Restore(strings.NewReader(bad)); err == nil {
			t.Fatalf("accepted bad save: %s", bad)
		}
		after, _ := json.Marshal(a.History().Context())
		if !bytes.Equal(before, after) {
			t.Fatal("bad restore changed existing history")
		}
	}
}

// Empty primary instructions deliberately suppress the legacy default. Save
// metadata and rendering must share that decision, not each invent a fallback.
func TestSaveIntentionallyEmptyPrimary(t *testing.T) {
	dir := skillFiles(t, map[string]string{"base": "---\nname: base\ntype: primary\n---\n"})
	a := skillAgent(t, ensemble.New(io.Discard), dir, "base", "")
	var saved bytes.Buffer
	if err := a.History().Save(&saved); err != nil {
		t.Fatal(err)
	}
	var save struct {
		// Config records request metadata without replacing runtime settings.
		Config struct {
			// SystemPrompt retains an empty primary rather than the legacy fallback.
			SystemPrompt string `json:"system_prompt"`
		} `json:"config"`
	}
	if err := json.Unmarshal(saved.Bytes(), &save); err != nil {
		t.Fatal(err)
	}
	body, err := a.Engine().Render(a.History().Context())
	if err != nil {
		t.Fatal(err)
	}
	var request struct {
		// System is Anthropic's instruction field, distinct from its conversation.
		System []json.RawMessage `json:"system"`
	}
	if err := json.Unmarshal(body, &request); err != nil {
		t.Fatal(err)
	}
	if save.Config.SystemPrompt != "" || len(request.System) != 0 {
		t.Fatalf("empty primary acquired default instructions: save=%q wire=%q", save.Config.SystemPrompt, request.System)
	}
}
