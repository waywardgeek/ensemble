package ensemble_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"example.com/ensemble"
)

func skillSource(name, kind, fields, body string) []byte {
	return []byte("---\nname: " + name + "\ndescription: Manual " + name + "\ntype: " + kind + "\n" + fields + "---\n" + body)
}
func skillConfig(t *testing.T) ensemble.Config {
	t.Helper()
	dir := t.TempDir()
	return ensemble.Config{DisableStreaming: true, APIKey: "local-fixture", Model: "fixture", Workspace: dir, LogPath: filepath.Join(dir, "events"), Builtins: []string{"load_skill", "unload_skill", "read_file", "write_file"}, Skills: &ensemble.SkillConfig{Primary: "base", Catalog: map[string][]byte{
		"base": skillSource("base", "primary", "tools: read_file\nloadable-skills: edit\n", "Identity."),
		"edit": skillSource("edit", "loadable", "tools: write_file\n", "Use write_file."),
	}}}
}
func TestSkillPublishedHintMaterialPromptProjection(t *testing.T) {
	for _, vendor := range []string{"anthropic", "openai", "gemini"} {
		t.Run(vendor, func(t *testing.T) {
			app := ensemble.New(nil)
			defer app.Close()
			config := skillConfig(t)
			config.Vendor = vendor
			a, err := app.NewAgent(config)
			if err != nil {
				t.Fatal(err)
			}
			appendEvent := func(e ensemble.Event) {
				t.Helper()
				if err := a.Append(e); err != nil {
					t.Fatal(err)
				}
			}
			appendEvent(ensemble.Event{Type: "turn_started", Turn: &ensemble.TurnEvent{RequestID: "first"}})
			appendEvent(ensemble.Event{Type: "message_received", Message: &ensemble.Entry{Actor: "human", Purpose: "dialogue", Parts: []ensemble.Part{ensemble.Text("Read.")}}})
			surface := map[string]string{"anthropic": "messages", "openai": "chat_completions", "gemini": "generate_content"}[vendor]
			from := ensemble.Provenance{Vendor: vendor, Model: "fixture", Surface: surface}
			appendEvent(ensemble.Event{Type: "response_ended", Response: &ensemble.Response{From: from, Usage: &ensemble.Usage{}, Parts: []ensemble.Part{{Type: "tool_call", CallID: "c1", Name: "read_file", Args: json.RawMessage(`{"path":"notes"}`), From: &from}}}})
			appendEvent(ensemble.Event{Type: "tool_called", Tool: &ensemble.ToolEvent{CallID: "c1", Name: "read_file", Args: json.RawMessage(`{"path":"notes"}`)}})
			appendEvent(ensemble.Event{Type: "hint_received", Hint: &ensemble.HintEvent{RequestID: "first", Text: "Remember H."}})
			hintSeq := a.Events()[len(a.Events())-1].Seq
			if _, err = a.LoadSkill("edit"); err != nil {
				t.Fatal(err)
			}
			if _, err = a.Render(ensemble.Config{Vendor: vendor, Model: "fixture"}); err == nil {
				t.Fatal("unresolved prefix rendered")
			}
			appendEvent(ensemble.Event{Type: "tool_returned", Tool: &ensemble.ToolEvent{CallID: "c1", Parts: []ensemble.Part{ensemble.Text("done")}}})
			resultSeq := a.Events()[len(a.Events())-1].Seq
			appendEvent(ensemble.Event{Type: "turn_ended", Turn: &ensemble.TurnEvent{RequestID: "first", Outcome: "round_limit"}})
			appendEvent(ensemble.Event{Type: "turn_started", Turn: &ensemble.TurnEvent{RequestID: "second"}})
			appendEvent(ensemble.Event{Type: "message_received", Message: &ensemble.Entry{Actor: "human", Purpose: "dialogue", Parts: []ensemble.Part{ensemble.Text("Continue P.")}}})
			renderConfig := ensemble.Config{DisableStreaming: true, Vendor: vendor, Model: "fixture", MaxTokens: 512, Tools: a.Config().Tools}
			before := a.Snapshot()
			one, err := a.Render(renderConfig)
			if err != nil {
				t.Fatal(err)
			}
			two, err := a.Render(renderConfig)
			if err != nil || !bytes.Equal(one, two) || !reflect.DeepEqual(before, a.Snapshot()) {
				t.Fatal("render impure")
			}
			text := string(one)
			markers := []string{"done", "Remember H.", "[skill edit activation 2]\\nUse write_file.\\n[/skill]", "Continue P."}
			previous := -1
			for _, marker := range markers {
				pos := strings.Index(text, marker)
				if pos <= previous {
					t.Fatalf("wrong order %s: %s", marker, text)
				}
				previous = pos
			}
			appendEvent(ensemble.Event{Type: "request_sent", Request: &ensemble.RequestEvent{Delivery: "plain", To: from, Hints: []uint64{hintSeq}, Ephemera: []uint64{}, Configuration: &ensemble.RequestConfig{System: "", MaxTokens: 512, Tools: renderConfig.Tools}}})
			requestSeq := a.Events()[len(a.Events())-1].Seq
			exact, err := a.ReconstructRequest(requestSeq)
			if err != nil || !bytes.Equal(exact, one) {
				t.Fatalf("capture reconstruction: %v", err)
			}
			if _, err = a.UnloadSkill("edit"); err != nil {
				t.Fatal(err)
			}
			if err = a.Redact(ensemble.Redaction{From: resultSeq, To: resultSeq, Level: "redact_result", Reason: "result only"}); err != nil {
				t.Fatal(err)
			}
			after, err := a.Render(renderConfig)
			if err != nil {
				t.Fatal(err)
			}
			if bytes.Contains(after, []byte("Remember H.")) || !bytes.Contains(after, []byte("[skill edit activation 2]")) || bytes.Index(after, []byte("[skill edit activation 2]")) > bytes.Index(after, []byte("Continue P.")) {
				t.Fatalf("consumption/retention failed: %s", after)
			}
			dump, _ := a.Dump()
			path := filepath.Join(t.TempDir(), "offline")
			_ = os.WriteFile(path, dump, 0600)
			offline, err := app.Load(path, ensemble.Config{Vendor: vendor, Model: "fixture", Skills: &ensemble.SkillConfig{Directory: "/absent", Primary: "no"}})
			if err != nil {
				t.Fatal(err)
			}
			replay, err := offline.ReconstructRequest(requestSeq)
			if err != nil || !bytes.Equal(replay, one) {
				t.Fatalf("offline replay: %v", err)
			}
			if _, err = offline.Render(ensemble.Config{Vendor: vendor, Model: "fixture", System: "competing"}); err == nil {
				t.Fatal("competing system accepted")
			}
		})
	}
}
func TestSkillPauseHoldsManagementButTypedControlsProceed(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if requests == 1 {
			fmt.Fprint(w, `{"content":[{"type":"tool_use","id":"load","name":"load_skill","input":{"name":"edit"}}],"usage":{"input_tokens":1,"output_tokens":1}}`)
		} else {
			fmt.Fprint(w, `{"content":[{"type":"text","text":"done"}],"usage":{"input_tokens":1,"output_tokens":1}}`)
		}
	}))
	defer server.Close()
	app := ensemble.New(nil)
	defer app.Close()
	c := skillConfig(t)
	c.BaseURL = server.URL
	a, err := app.NewAgent(c)
	if err != nil {
		t.Fatal(err)
	}
	pause, _ := a.RegisterPause()
	defer pause.Close()
	_, _ = pause.Update(true, false)
	h, err := a.Submit("load")
	if err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(3 * time.Second)
	for {
		snapshot, w, err := a.Watch()
		if err != nil {
			t.Fatal(err)
		}
		w.Close()
		if snapshot.State.Lifecycle == "tools_pending" {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("no pending batch")
		}
		time.Sleep(time.Millisecond)
	}
	if _, err = a.Hint("Remember H."); err != nil {
		t.Fatal(err)
	}
	if result, err := a.LoadSkill("edit"); err != nil || !result.Changed {
		t.Fatal(result, err)
	}
	for _, e := range a.Events() {
		if e.Tool != nil {
			t.Fatal("paused model call was admitted")
		}
	}
	if _, err = a.Interrupt(); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	completion, err := h.Wait(ctx)
	if err != nil || completion.Outcome != "interrupted" {
		t.Fatal(completion, err)
	}
	state, _ := a.SkillState()
	if state.Revision != 1 || len(state.Roots) != 1 {
		t.Fatal("interrupt undid typed commit")
	}
	if len(a.Snapshot().Jobs) != 0 {
		t.Fatal("management allocated a job")
	}
}
func TestSkillRejectInvalidUTF8BeforeCopy(t *testing.T) {
	app := ensemble.New(nil)
	defer app.Close()
	c := skillConfig(t)
	c.Skills.Variables = map[string]string{"PROJECT": string([]byte{255})}
	if _, err := app.NewAgent(c); err == nil {
		t.Fatal("JSON copy normalized invalid scalar before validation")
	}
}

func TestSkillAdmissionEffectBoundary(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if requests == 1 {
			fmt.Fprint(w, `{"content":[{"type":"tool_use","id":"before","name":"write_file","input":{"path":"forbidden-before","content":"bad"}},{"type":"tool_use","id":"load","name":"load_skill","input":{"name":"edit"}},{"type":"tool_use","id":"allowed","name":"write_file","input":{"path":"allowed","content":"good"}},{"type":"tool_use","id":"unload","name":"unload_skill","input":{"name":"edit"}},{"type":"tool_use","id":"after","name":"write_file","input":{"path":"forbidden-after","content":"bad"}}],"usage":{"input_tokens":1,"output_tokens":1}}`)
		} else {
			fmt.Fprint(w, `{"content":[{"type":"text","text":"done"}],"usage":{"input_tokens":1,"output_tokens":1}}`)
		}
	}))
	defer server.Close()
	app := ensemble.New(nil)
	defer app.Close()
	config := skillConfig(t)
	config.BaseURL = server.URL
	a, err := app.NewAgent(config)
	if err != nil {
		t.Fatal(err)
	}
	// This positive declaration control stays intact in the admission deletion.
	for _, tool := range a.Config().Tools {
		if tool.Name == "write_file" {
			t.Fatal("disabled declaration visible")
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if _, err = a.Prompt(ctx, "exercise batch"); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"forbidden-before", "forbidden-after"} {
		if _, err = os.Stat(filepath.Join(config.Workspace, name)); !os.IsNotExist(err) {
			t.Fatalf("disabled tool created %s", name)
		}
	}
	data, err := os.ReadFile(filepath.Join(config.Workspace, "allowed"))
	if err != nil || string(data) != "good" {
		t.Fatal("permitted write absent", err)
	}
	if requests != 2 {
		t.Fatal("unexpected requests", requests)
	}
}

func TestSkillEscapedDurableRecordReplaysBeyondOrdinaryLineLimit(t *testing.T) {
	app := ensemble.New(nil)
	defer app.Close()
	config := skillConfig(t)
	config.Skills.Variables = map[string]string{"CHUNK": strings.Repeat("<", 4096)}
	body := strings.Repeat("${CHUNK}", 16)
	dependencies := []string{}
	for i := 0; i < 43; i++ {
		name := fmt.Sprintf("dep-%02d", i)
		dependencies = append(dependencies, name)
		config.Skills.Catalog[name] = skillSource(name, "dependency", "", body)
	}
	config.Skills.Catalog["base"] = skillSource("base", "primary", "depends: "+strings.Join(dependencies, " ")+"\n", body)
	a, err := app.NewAgent(config)
	if err != nil {
		t.Fatal(err)
	}
	stat, err := os.Stat(config.LogPath)
	if err != nil || stat.Size() <= 16*1024*1024 {
		t.Fatal("control did not cross encoded line boundary", err)
	}
	offline, err := app.Load(config.LogPath, ensemble.Config{Vendor: "anthropic", Model: "fixture"})
	if err != nil {
		t.Fatal("valid escaped transition failed replay", err)
	}
	view, err := offline.InspectSkills()
	if err != nil || len(view.Material) != 44 {
		t.Fatal("material lost", err)
	}
	live, _ := a.InspectSkills()
	if !reflect.DeepEqual(live, view) {
		t.Fatal("decoded material changed")
	}
}

func TestSkillWatchCurrentStateOutlivesMaterialWindow(t *testing.T) {
	app := ensemble.New(nil)
	defer app.Close()
	a, err := app.NewAgent(skillConfig(t))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = a.LoadSkill("edit"); err != nil {
		t.Fatal(err)
	}
	if _, err = a.UnloadSkill("edit"); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 105; i++ {
		if err = a.Ephemeral("later guidance"); err != nil {
			t.Fatal(err)
		}
	}
	snapshot, w, err := a.Watch()
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	if snapshot.State.Skills == nil || snapshot.State.Skills.Revision != 2 || len(snapshot.State.Skills.Retired) != 1 || len(snapshot.Events) != 100 {
		t.Fatal("current state lost beyond window")
	}
	for _, e := range snapshot.Events {
		if e.Skills != nil {
			t.Fatal("fixture retained a skill event")
		}
	}
	snapshot.State.Skills.Tools[0] = "corrupt"
	state, _ := a.SkillState()
	if state.Tools[0] == "corrupt" {
		t.Fatal("watch state aliases authority")
	}
}
