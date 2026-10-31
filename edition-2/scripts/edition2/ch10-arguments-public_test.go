package ch10_test

// Narrow public controls from Chapter 10's dd1111e clarification. Existing
// external-consumer helpers supply transport/construction, never the oracle.
import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"example.com/ensemble"
)

const argumentDuplicate = `{"name": "edit", "name": "edit"}`
const argumentSpaced = `{"name": "edit"}`
const argumentReadMarker = "ARGUMENT-NEXT-READ-USES-DEFAULT-LIMITS"

func argumentOptions(t *testing.T, ep *endpoint, session bool) ensemble.SessionOptions {
	o := wireOptions(t, "openai", ep)
	o.System = nil
	o.Config.System = ""
	o.Config.Builtins = []string{"read_file", "write_file", "tool_limits", "load_skill", "unload_skill"}
	o.Config.Skills = &ensemble.SkillConfig{Primary: "base", Catalog: map[string][]byte{
		"base": []byte("---\nname: base\ndescription: Base\ntype: primary\ntools: read_file tool_limits\nloadable-skills: edit\n---\nRead the local fixture.\n$TOOLS\n$SKILLS\n"),
		"edit": []byte("---\nname: edit\ndescription: Edit\ntype: loadable\ntools: write_file\n---\nEdit the local fixture.\n"),
	}}
	if !session {
		o.Config.DataDir = ""
		o.Config.LogPath = filepath.Join(o.Config.Workspace, "standalone.log")
	}
	if err := os.WriteFile(filepath.Join(o.Config.Workspace, "probe.txt"), []byte(argumentReadMarker+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	return o
}

func argumentReply(t *testing.T, x exchange, id, name, args string) {
	t.Helper()
	encoded, err := json.Marshal(args)
	if err != nil {
		t.Fatal(err)
	}
	x.reply <- []byte(`{"model":"gpt-4.1-mini-2025-04-14","choices":[{"index":0,"message":{"role":"assistant","content":"","tool_calls":[{"id":"` + id + `","type":"function","function":{"name":"` + name + `","arguments":` + string(encoded) + `}}]},"finish_reason":"tool_calls"}],"usage":{"prompt_tokens":1,"completion_tokens":1}}`)
}

func argumentWire(t *testing.T, body []byte, id, expected string) {
	t.Helper()
	var request struct {
		Messages []struct {
			Calls []struct {
				ID       string `json:"id"`
				Function struct {
					Arguments string `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(body, &request); err != nil {
		t.Fatal(err)
	}
	found := 0
	for _, message := range request.Messages {
		for _, call := range message.Calls {
			if call.ID == id {
				found++
				if call.Function.Arguments != expected {
					t.Fatalf("%s argument string changed: got %q want %q", id, call.Function.Arguments, expected)
				}
			}
		}
	}
	if found != 1 {
		t.Fatalf("%s historical call count %d, want one", id, found)
	}
}

func argumentReplay(t *testing.T, a *ensemble.Agent, start uint64, bodies [][]byte) {
	t.Helper()
	var sends []uint64
	for _, event := range a.Events() {
		if event.Type == "request_sent" && event.Seq > start {
			sends = append(sends, event.Seq)
		}
	}
	if len(sends) != len(bodies) {
		t.Fatalf("request/replay witnesses %d/%d", len(sends), len(bodies))
	}
	for i, seq := range sends {
		raw, err := a.ReconstructRequest(seq)
		if err != nil || !bytes.Equal(raw, bodies[i]) {
			t.Fatalf("exact replay differs at sequence %d: %v", seq, err)
		}
	}
}

func argumentSeed(t *testing.T, session bool) (*ensemble.Agent, ensemble.SessionOptions, *endpoint) {
	t.Helper()
	ep := localEndpoint(t)
	o := argumentOptions(t, ep, session)
	app := application(t)
	var a *ensemble.Agent
	var err error
	if session {
		a, err = app.OpenSession(o)
	} else {
		a, err = app.NewAgent(o.Config)
	}
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = a.Close() })
	// The published acknowledgement explicitly uses revision 3. These are real
	// public transitions; the invalid call must not add a fourth transition.
	skillChange(t, a, true, "edit")
	skillChange(t, a, false, "edit")
	skillChange(t, a, true, "edit")
	before := skillView(t, a)
	h, err := a.Submit("Check controlled argument errors, then continue.")
	if err != nil {
		t.Fatal(err)
	}
	var bodies [][]byte
	for _, call := range []struct{ id, name, args string }{
		{"argument-setter", "tool_limits", `{"max_output_bytes":1}`},
		{"argument-duplicate", "load_skill", argumentDuplicate},
		{"argument-read", "read_file", `{"path":"probe.txt"}`},
		{"argument-spaced", "load_skill", argumentSpaced},
	} {
		x := nextRequest(t, ep)
		bodies = append(bodies, x.body)
		argumentReply(t, x, call.id, call.name, call.args)
	}
	x := nextRequest(t, ep)
	bodies = append(bodies, x.body)
	argumentWire(t, x.body, "argument-duplicate", argumentDuplicate)
	argumentWire(t, x.body, "argument-spaced", argumentSpaced)
	answer(t, "openai", x)
	finish(t, h)
	if !reflect.DeepEqual(before, skillView(t, a)) {
		t.Fatal("invalid/no-op management changed Skills state")
	}
	called, returned, consumed := 0, 0, 0
	ordinary, spaced := false, false
	for _, event := range a.Events() {
		if event.Type == "tool_limits_consumed" {
			consumed++
			if event.Limits == nil || event.Limits.CallID != "argument-duplicate" || event.Limits.Name != "load_skill" {
				t.Fatal("consumption assigned to wrong attempt")
			}
		}
		if event.Tool == nil {
			continue
		}
		tool := event.Tool
		var text string
		for _, part := range tool.Parts {
			if part.Text != nil {
				text += *part.Text
			}
		}
		if tool.CallID == "argument-duplicate" {
			if tool.Job != nil {
				t.Fatal("management error allocated a Job")
			}
			if event.Type == "tool_called" {
				called++
			}
			if event.Type == "tool_returned" {
				returned++
				lines := strings.Split(text, "\n")
				if !tool.IsError || len(lines) < 2 || lines[len(lines)-1] != `{"error":"invalid_skill_arguments","name":"","revision":3}` || strings.Count(text, "tool_limits") != 1 || !strings.Contains(lines[0], "load_skill") {
					t.Fatalf("controlled acknowledgement/consumption note differs: %q", text)
				}
			}
		}
		if event.Type == "tool_returned" && tool.CallID == "argument-read" {
			ordinary = !tool.IsError && strings.Contains(text, argumentReadMarker) && !strings.Contains(text, "tool_limits")
		}
		if event.Type == "tool_returned" && tool.CallID == "argument-spaced" {
			spaced = !tool.IsError && text == `{"status":"unchanged","name":"edit","revision":3,"changed":false}`
		}
	}
	if called != 1 || returned != 1 || !ordinary || !spaced || (session && consumed != 1) {
		t.Fatalf("controlled pair/continuation/once-consumption failed: called=%d returned=%d ordinary=%v spaced=%v consumed=%d", called, returned, ordinary, spaced, consumed)
	}
	argumentReplay(t, a, 0, bodies)
	return a, o, ep
}

func TestCh10ArgumentsStandalone(t *testing.T) {
	a, o, ep := argumentSeed(t, false)
	closed(t, a)
	loaded, err := application(t).Load(o.Config.LogPath, o.Config)
	if err != nil {
		t.Fatal(err)
	}
	body, err := loaded.Render(o.Config)
	if err != nil {
		t.Fatal(err)
	}
	argumentWire(t, body, "argument-duplicate", argumentDuplicate)
	argumentWire(t, body, "argument-spaced", argumentSpaced)
	noStartupHTTP(t, ep)
}

func TestCh10ArgumentsSessionRecovery(t *testing.T) {
	a, o, ep := argumentSeed(t, true)
	old := export(t, a)
	turn(t, a, ep, "openai", "GENUINE-NEWER-ARGUMENT-TAIL")
	latest := export(t, a)
	want := skillView(t, a)
	usage := a.UsageByModel()
	if latest.AsOf <= old.AsOf {
		t.Fatal("empty tail cannot prove recovery")
	}
	closed(t, a)
	var baseline []byte
	for _, mode := range []string{"checkpoint", "full-log", "snapshot-tail", "imported-tail"} {
		t.Run(mode, func(t *testing.T) {
			candidate := o
			candidate.Config.DataDir = filepath.Join(o.Config.Workspace, mode)
			app := application(t)
			var b *ensemble.Agent
			var err error
			if mode == "imported-tail" {
				b, err = app.ImportSession(latest.Bytes, candidate)
			} else {
				copies(t, o.Config.DataDir, candidate.Config.DataDir)
				if mode == "full-log" {
					err = os.Remove(filepath.Join(candidate.Config.DataDir, "checkpoint.json"))
				} else if mode == "snapshot-tail" {
					err = os.WriteFile(filepath.Join(candidate.Config.DataDir, "checkpoint.json"), old.Bytes, 0600)
				}
				if err == nil {
					b, err = app.OpenSession(candidate)
				}
			}
			if err != nil {
				t.Fatalf("settled positive %s refused: %v", mode, err)
			}
			noStartupHTTP(t, ep)
			if !reflect.DeepEqual(want, skillView(t, b)) || !reflect.DeepEqual(usage, b.UsageByModel()) {
				t.Fatal("recovery changed Skills/usage")
			}
			boundary := export(t, b).AsOf
			h, err := b.Submit("SAME-RECOVERED-REQUEST")
			if err != nil {
				t.Fatal(err)
			}
			x := nextRequest(t, ep)
			body := x.body
			argumentReply(t, x, "recovered-read", "read_file", `{"path":"probe.txt"}`)
			continuation := nextRequest(t, ep)
			answer(t, "openai", continuation)
			finish(t, h)
			readOK := false
			for _, event := range b.Events() {
				if event.Type == "tool_returned" && event.Tool != nil && event.Tool.CallID == "recovered-read" {
					text := ""
					for _, part := range event.Tool.Parts {
						if part.Text != nil {
							text += *part.Text
						}
					}
					readOK = !event.Tool.IsError && strings.Contains(text, argumentReadMarker) && !strings.Contains(text, "tool_limits")
				}
				if event.Seq > boundary && event.Type == "tool_limits_consumed" {
					t.Fatal("recovery resurrected consumed pending limits")
				}
			}
			if !readOK {
				t.Fatal("resumed ordinary call lost default limits")
			}
			argumentWire(t, body, "argument-duplicate", argumentDuplicate)
			argumentWire(t, body, "argument-spaced", argumentSpaced)
			argumentReplay(t, b, boundary, [][]byte{body, continuation.body})
			if baseline == nil {
				baseline = body
			} else if !bytes.Equal(baseline, body) {
				t.Fatal("recovery paths produced different next requests")
			}
			closed(t, b)
			// Includes a settled real post-origin turn; an unfinished-tail refusal
			// is a distinct failure and cannot stand in for this positive.
			reopened := opened(t, app, candidate)
			noStartupHTTP(t, ep)
			rendered, err := reopened.Render(reopened.Config())
			if err != nil {
				t.Fatal(err)
			}
			argumentWire(t, rendered, "argument-duplicate", argumentDuplicate)
			argumentWire(t, rendered, "argument-spaced", argumentSpaced)
		})
	}
}

func TestCh10ArgumentsSnapshotRefusals(t *testing.T) {
	a, o, ep := argumentSeed(t, true)
	x := export(t, a)
	closed(t, a)
	parentBytes := skillsStoreBytes(t, o.Config.DataDir)
	app := application(t)
	if _, err := app.InspectCheckpoint(x.Bytes); err != nil {
		t.Fatalf("genuine snapshot parent: %v", err)
	}
	command := exec.Command("python3", os.Getenv("CH10_ARGUMENT_CASES"))
	command.Stdin = bytes.NewReader(x.Bytes)
	data, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("argument mutation adapter: %v: %s", err, data)
	}
	var cases []struct {
		Name     string `json:"name"`
		Bytes    string `json:"bytes"`
		Positive bool   `json:"positive"`
		Code     string `json:"expected_code"`
	}
	if err = json.Unmarshal(data, &cases); err != nil || len(cases) != 12 {
		t.Fatalf("mutation inventory changed: %v count=%d", err, len(cases))
	}
	positives := 0
	for _, fixture := range cases {
		t.Run(fixture.Name, func(t *testing.T) {
			raw, err := base64.StdEncoding.DecodeString(fixture.Bytes)
			if err != nil {
				t.Fatal(err)
			}
			candidate := o
			candidate.Config.DataDir = filepath.Join(o.Config.Workspace, fixture.Name)
			_, inspectErr := app.InspectCheckpoint(raw)
			mounted, importErr := app.ImportSession(raw, candidate)
			if fixture.Positive {
				if inspectErr != nil || importErr != nil {
					t.Fatalf("required valid parent refused: inspect=%v import=%v", inspectErr, importErr)
				}
				positives++
				closed(t, mounted)
			} else {
				if positives != 2 {
					t.Fatal("negative control has no accepted positive parents")
				}
				code(t, inspectErr, fixture.Code)
				code(t, importErr, fixture.Code)
				for _, leaf := range []string{"events.log", "checkpoint.json", "origin.json"} {
					if _, err := os.Stat(filepath.Join(candidate.Config.DataDir, leaf)); !os.IsNotExist(err) {
						t.Fatal("refused import wrote durable session leaf")
					}
				}
			}
			noStartupHTTP(t, ep)
		})
	}
	if !reflect.DeepEqual(parentBytes, skillsStoreBytes(t, o.Config.DataDir)) {
		t.Fatal("argument refusal mutated original durable parent")
	}
}
