package gui

import (
	"context"
	"encoding/json"
	"ensemble"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChapter15LiveMeterAndReplay(t *testing.T) {
	peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"model":"claude-sonnet-4-6","content":[{"type":"text","text":"answer"}],"usage":{"input_tokens":10,"output_tokens":4,"cache_read_input_tokens":80,"cache_creation_input_tokens":10}}`)
	}))
	defer peer.Close()
	app := ensemble.New(nil)
	defer app.Shutdown()
	a, err := app.NewAgent(ensemble.Config{Model: "claude-sonnet-4-6", APIKey: "dummy", BaseURL: peer.URL, BaseDir: t.TempDir(), DataDir: t.TempDir(), Tools: []string{}})
	if err != nil {
		t.Fatal(err)
	}
	h := NewHub(app)
	defer h.Close()
	c := h.connect()
	defer h.disconnect(c)
	var replay struct {
		Type string         `json:"type"`
		Data replayEnvelope `json:"data"`
	}
	if err = json.Unmarshal(<-c.queue, &replay); err != nil {
		t.Fatal(err)
	}
	if len(replay.Data.Agents) != 1 || replay.Data.Agents[0].Meter.Requests != 0 || replay.Data.Agents[0].Meter.Cache.Current != nil {
		t.Fatal("fresh replay carried usage")
	}
	if _, err = a.Ask(context.Background(), "question"); err != nil {
		t.Fatal(err)
	}
	// Synchronize the existing production view path before examining queued frames.
	h.mu.Lock()
	h.sync(h.view(a), a.View())
	h.mu.Unlock()
	found := false
	for len(c.queue) > 0 {
		var w Wire
		if err = json.Unmarshal(<-c.queue, &w); err != nil {
			t.Fatal(err)
		}
		if w.Type == "usage" && w.Meter.Requests == 1 {
			found = true
			if w.Meter.Session.Output != 4 || w.Meter.LastHitRate == nil || *w.Meter.LastHitRate != .8 || w.Meter.Cost.Status != "known" {
				t.Fatalf("bad live meter %+v", w.Meter)
			}
		}
	}
	if !found {
		t.Fatal("no live usage frame")
	}
	reconnect := h.connect()
	defer h.disconnect(reconnect)
	if err = json.Unmarshal(<-reconnect.queue, &replay); err != nil {
		t.Fatal(err)
	}
	if replay.Data.Agents[0].Meter.Requests != 1 || replay.Data.Agents[0].Meter.Session.Output != 4 {
		t.Fatal("reconnect lost current-process meter")
	}
}
