package ensemble_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"ensemble"
	engineowner "ensemble/internal/engine"
)

func TestChapter14KeepDecisionConsumedOnce(t *testing.T) {
	kept, newBody := strings.Repeat("KEPT-FIRST-", 1200), strings.Repeat("NEW-UNKEPT-", 600)
	var calls atomic.Int32
	captured := make(chan string, 10)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		captured <- string(b)
		switch calls.Add(1) {
		case 1:
			ch5Reply(w, ch5Call("first", "large_result", map[string]string{"which": "first"}))
		case 2:
			ch5Reply(w, ch5Call("keep", "keep_tool_results", map[string]string{}))
		case 5:
			ch5Reply(w, ch5Call("new", "large_result", map[string]string{"which": "new"}))
		default:
			ch5Reply(w, map[string]string{"type": "text", "text": "done"})
		}
	}))
	defer server.Close()
	app := ensemble.New(nil)
	defer app.Shutdown()
	a, err := app.NewAgent(ensemble.Config{Vendor: ensemble.Anthropic, Model: "claude-opus-5-course", APIKey: "local-fixture-only", BaseURL: server.URL, BaseDir: t.TempDir(), DataDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	err = a.Tools().Register(ensemble.ToolDefinition{Name: "large_result", Description: "local large result", Parameters: json.RawMessage(`{"type":"object","properties":{"which":{"type":"string"}}}`)}, func(_ context.Context, c ensemble.ToolCall, raw json.RawMessage) error {
		var args struct{ Which string }
		if err := json.Unmarshal(raw, &args); err != nil {
			return err
		}
		text := kept
		if args.Which == "new" {
			text = newBody
		}
		_, err := io.WriteString(c, text)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	for _, prompt := range []string{"get first and keep", "text-only turn", "get second", "text-only after second"} {
		if _, err = a.Ask(ctx, prompt); err != nil {
			t.Fatal(err)
		}
	}
	if calls.Load() != 7 {
		t.Fatalf("requests=%d", calls.Load())
	}
	requests := make([]string, 7)
	for i := range requests {
		requests[i] = <-captured
	}
	for _, i := range []int{1, 2, 3, 4, 5, 6} {
		if !strings.Contains(requests[i], kept) {
			t.Errorf("request %d reclassified kept result below both bands", i+1)
		}
	}
	if !strings.Contains(requests[5], newBody) || strings.Contains(requests[6], newBody) {
		t.Error("new batch was not presented once then stubbed without keep")
	}
	if err = a.SetContextTarget(20000); err != nil {
		t.Fatal(err)
	}
	if _, err = a.Ask(ctx, "apply smaller byte bands"); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(<-captured, kept) {
		t.Error("previous keep became a permanent band exemption")
	}
}

type noArchiveHistory struct {
	ensemble.History
	t *testing.T
}

func (h noArchiveHistory) EventsRange(_, _ ensemble.Seq) []ensemble.Event {
	h.t.Error("ordinary Engine render attempted archive range read")
	return nil
}

type liveRenderAgent struct {
	ensemble.Agent
	t *testing.T
}

func (a liveRenderAgent) Log() ensemble.History { return noArchiveHistory{a.Agent.Log(), a.t} }
func TestChapter14RenderUsesLiveLoadPair(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if requests.Add(1) == 1 {
			ch5Reply(w, ch5Call("live-load", "load_skill", map[string]string{"name": "writer"}))
		} else {
			ch5Reply(w, map[string]string{"type": "text", "text": "done"})
		}
	}))
	defer server.Close()
	app := ensemble.New(nil)
	defer app.Shutdown()
	a, err := app.NewAgent(ensemble.Config{Vendor: ensemble.Anthropic, Model: "claude-sonnet-4-6", APIKey: "local-fixture-only", BaseURL: server.URL, BaseDir: t.TempDir(), DataDir: t.TempDir(), SavePath: filepath.Join(t.TempDir(), "saved.archive"), Skills: &ensemble.SkillSpec{Directory: c9Fixture(t), Primary: "base"}})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := a.Ask(ctx, "load writer"); err != nil {
		t.Fatal(err)
	}
	if err := a.Save(); err != nil {
		t.Fatal(err)
	}
	e, err := engineowner.New(liveRenderAgent{a, t})
	if err != nil {
		t.Fatal(err)
	}
	c := a.Log().Context()
	for i := 0; i < 3; i++ {
		wire, err := e.Render(c)
		if err != nil {
			t.Fatal(err)
		}
		c9CheckWire(t, wire, true, "live-load")
	}
	var foreign ensemble.Context
	encoded, _ := json.Marshal(c)
	if err := json.Unmarshal(encoded, &foreign); err != nil {
		t.Fatal(err)
	}
	for i := range foreign.Dialogue {
		for j := range foreign.Dialogue[i].Parts {
			p := &foreign.Dialogue[i].Parts[j]
			if p.Type == "tool_call" {
				p.From = &ensemble.Provenance{Vendor: ensemble.Gemini, Model: "gemini-local", Surface: ensemble.GenerateContent}
			}
		}
	}
	wire, err := e.Render(foreign)
	if err != nil {
		t.Fatal(err)
	}
	var foreignWire struct {
		Tools []struct {
			Name     string `json:"name"`
			Deferred bool   `json:"defer_loading"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(wire, &foreignWire); err != nil {
		t.Fatal(err)
	}
	ordinary := false
	for _, tool := range foreignWire.Tools {
		if tool.Name == "write_file" {
			ordinary = !tool.Deferred
		}
	}
	if !ordinary {
		t.Fatal("foreign replay hid the live ordinary declaration")
	}
	if strings.Contains(string(wire), `"admission_seq"`) {
		t.Fatal("internal provenance leaked onto provider wire")
	}
	if strings.Contains(string(wire), `"tool_reference"`) {
		t.Fatal("foreign pair leaked a native reference")
	}
	// A handoff removes the actual pair, not the registry's live grant.
	if _, failed := a.Tools().Dispatch(ctx, "micro_handoff", json.RawMessage(`{"text":"keep visible progress"}`)); failed {
		t.Fatal("handoff failed")
	}
	wire, err = e.Render(a.Log().Context())
	if err != nil {
		t.Fatal(err)
	}
	c9CheckWire(t, wire, false, "")
	if strings.Contains(string(wire), `"tool_reference"`) {
		t.Fatal("handoff invented a reference without its pair")
	}
}
