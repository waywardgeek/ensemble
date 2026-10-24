package gui

import (
	"context"
	"encoding/json"
	"ensemble"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/gorilla/websocket"
)

func TestChapter15R1RealUsageWireAndConsumer(t *testing.T) {
	var responses atomic.Int32
	peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if responses.Add(1) == 1 {
			fmt.Fprint(w, `{"model":"claude-sonnet-4-6","content":[{"type":"text","text":"answer"}],"usage":{"input_tokens":11,"output_tokens":13,"cache_read_input_tokens":17,"cache_creation_input_tokens":19}}`)
		} else {
			fmt.Fprint(w, `{"model":"unknown-served-model","content":[{"type":"text","text":"answer"}],"usage":{"input_tokens":3,"output_tokens":5,"cache_read_input_tokens":7,"cache_creation_input_tokens":11}}`)
		}
	}))
	defer peer.Close()
	app := ensemble.New(nil)
	defer app.Shutdown()
	a, err := app.NewAgent(ensemble.Config{Model: "claude-sonnet-4-6", APIKey: "local-fixture", BaseURL: peer.URL, BaseDir: t.TempDir(), DataDir: t.TempDir(), Tools: []string{}})
	if err != nil {
		t.Fatal(err)
	}
	server, err := NewServerWithSettings(app, "web/gui", filepath.Join(t.TempDir(), "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	server.DefaultAgent = a.ID()
	host := httptest.NewServer(server.Handler(server.Assets))
	defer host.Close()
	type observation struct {
		Type    string `json:"type"`
		AgentID string `json:"agent_id"`
		*UsageFrame
		Total *ensemble.Usage       `json:"total"`
		Meter ensemble.EngineStatus `json:"meter"`
	}
	decode := func(raw json.RawMessage) observation {
		var o observation
		if err := json.Unmarshal(raw, &o); err != nil {
			t.Fatal(err)
		}
		if o.Type == "usage" {
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(raw, &fields); err != nil {
				t.Fatal(err)
			}
			if o.Total == nil {
				t.Fatalf("usage frame lacks public nested total: %s", raw)
			}
			var total map[string]json.RawMessage
			if err := json.Unmarshal(fields["total"], &total); err != nil {
				t.Fatal(err)
			}
			if len(total) != 4 {
				t.Fatalf("total must contain exactly four disjoint counts: %s", raw)
			}
			for _, key := range []string{"input", "output", "cache_write", "cache_read"} {
				if _, ok := total[key]; !ok {
					t.Fatalf("total missing %s: %s", key, raw)
				}
				if _, ok := fields["total_"+key]; ok {
					t.Fatalf("legacy flat total alias remains: %s", raw)
				}
			}
		}
		return o
	}
	readReplay := func(ws *websocket.Conn) json.RawMessage {
		t.Helper()
		for {
			var frame struct {
				Type string
				Data struct{ Events []json.RawMessage }
			}
			if err := ws.ReadJSON(&frame); err != nil {
				t.Fatal(err)
			}
			if frame.Type != "replay" {
				continue
			}
			for _, raw := range frame.Data.Events {
				if decode(raw).Type == "usage" {
					return raw
				}
			}
			t.Fatal("replay omitted initial usage frame")
		}
	}
	first := r4Dial(t, host.URL)
	initial := readReplay(first)
	zero := decode(initial)
	if zero.UsageFrame == nil || zero.Input != 0 || zero.CacheRead != 0 || *zero.Total != (ensemble.Usage{}) || zero.CostUSD == nil || *zero.CostUSD != 0 || !zero.Priced {
		t.Fatalf("bad initial wire %s", initial)
	}
	if _, err = a.Ask(context.Background(), "question"); err != nil {
		t.Fatal(err)
	}
	var live json.RawMessage
	for {
		if err = first.ReadJSON(&live); err != nil {
			t.Fatal(err)
		}
		v := decode(live)
		if v.Type == "usage" && v.Meter.Requests == 1 {
			break
		}
	}
	got := decode(live)
	if got.Input != 11 || got.Output != 13 || got.CacheRead != 17 || got.CacheWrite != 19 || got.Total.Input != 11 || got.Total.Output != 13 || got.Total.CacheRead != 17 || got.Total.CacheWrite != 19 || got.HitRate == nil || *got.HitRate != 17.0/47.0 || got.CostUSD == nil || !got.Priced {
		t.Fatalf("incorrect disjoint usage frame: %s", live)
	}
	if _, err = a.Ask(context.Background(), "another question"); err != nil {
		t.Fatal(err)
	}
	var partial json.RawMessage
	for {
		if err = first.ReadJSON(&partial); err != nil {
			t.Fatal(err)
		}
		v := decode(partial)
		if v.Type == "usage" && v.Meter.Requests == 2 {
			break
		}
	}
	second := decode(partial)
	if second.Input != 3 || second.Output != 5 || second.CacheRead != 7 || second.CacheWrite != 11 || second.Total.Input != 14 || second.Total.Output != 18 || second.Total.CacheRead != 24 || second.Total.CacheWrite != 30 || second.HitRate == nil || *second.HitRate != 24.0/68.0 || second.CostUSD != nil || second.Priced || second.Meter.Cost.Status != "partial" || second.Meter.Cost.USD != got.Meter.Cost.USD {
		t.Fatalf("latest/session or unpriced facts wrong: %s", partial)
	}
	again := r4Dial(t, host.URL)
	reconnected := readReplay(again)
	re := decode(reconnected)
	if *re.Total != *second.Total || re.CostUSD != nil || re.Meter.Cost != second.Meter.Cost {
		t.Fatal("reconnect reset or double-counted session")
	}
	// Feed the actual server-emitted frames into the production socket/DOM code.
	fixture := filepath.Join(t.TempDir(), "usage-frames.json")
	b, _ := json.Marshal(map[string]json.RawMessage{"initial": initial, "live": live, "partial": partial, "reconnected": reconnected})
	if err = os.WriteFile(fixture, b, 0600); err != nil {
		t.Fatal(err)
	}
	command := exec.Command("node", "ch15_r1_meter_test.cjs", fixture)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("production JS consumer: %v\n%s", err, output)
	}
}
