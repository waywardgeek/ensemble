package llm

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// The old actor could declare a running call lost just before it returned.
// Build the old snapshot explicitly: the current reducer must not be used to
// manufacture a fixture that is meant to contain the old reducer's damage.
func lateResultSave(t *testing.T, mode string, isError bool) *SaveFile {
	t.Helper()
	call := lateResultCall()
	lost := resultEntry(2, "late")
	lost.Parts = common.PartList{common.ToolResultPart{CallID: "late", IsError: true, Parts: common.PartList{common.TextPart{Text: "unknown outcome"}}}}
	human := common.Entry{Seq: 3, Actor: common.ActorHuman, Kind: common.KindDialogue, Parts: common.PartList{common.TextPart{Text: "continue"}}}
	real := resultEntry(4, "late")
	real.Parts = common.PartList{common.ToolResultPart{CallID: "late", IsError: isError, Parts: common.PartList{common.TextPart{Text: "actual tool output"}}}}
	events := []common.Event{
		{Seq: 1, Type: common.ResponseEnded, Response: &common.ResponseData{Parts: call.Parts, From: call.Parts[0].(common.ToolCallPart).From}},
		{Seq: 2, Type: common.ToolResultLost, Tool: &common.ToolData{CallID: "late", Parts: common.PartList{common.TextPart{Text: "unknown outcome"}}}},
		{Seq: 3, Type: common.MessageReceived, Message: &common.MessageData{Actor: common.ActorHuman, Parts: human.Parts}},
		{Seq: 4, Type: common.ToolReturned, Tool: &common.ToolData{CallID: "late", Parts: common.PartList{common.TextPart{Text: "actual tool output"}}, IsError: isError}},
	}
	sf := &SaveFile{Log: events}
	switch mode {
	case "snapshot":
		sf.AsOf = 4
		sf.Context = common.NewContext()
		sf.Context.Turn = common.InputPending
		sf.Context.Dialogue = []common.Entry{call, lost, human, real}
	case "tail":
		sf.AsOf = 3
		sf.Context = common.NewContext()
		sf.Context.Turn = common.InputPending
		sf.Context.Dialogue = []common.Entry{call, lost, human}
	case "replay":
	default:
		t.Fatalf("unknown mode %q", mode)
	}
	return sf
}

func TestLateToolResultRestoresEquallyFromSnapshotAndLog(t *testing.T) {
	for _, isError := range []bool{false, true} {
		var want *common.Context
		for _, mode := range []string{"replay", "snapshot", "tail"} {
			t.Run(fmt.Sprintf("%s/error=%v", mode, isError), func(t *testing.T) {
				sf := lateResultSave(t, mode, isError)
				before := lateResultJSON(t, sf.Log)
				var diagnostics []string
				ctx := sf.Restore(func(err error) { diagnostics = append(diagnostics, err.Error()) })
				after := lateResultJSON(t, sf.Log)
				if string(before) != string(after) {
					t.Fatal("repair changed the audit log")
				}
				if len(ctx.Dialogue) != 3 {
					t.Fatalf("got %d entries, want call, real result, human message", len(ctx.Dialogue))
				}
				result, ok := ctx.Dialogue[1].Parts[0].(common.ToolResultPart)
				if !ok || result.CallID != "late" || result.IsError != isError || !reflect.DeepEqual(result.Parts, []common.Part{common.TextPart{Text: "actual tool output"}}) {
					t.Fatalf("result is not the real tool output: %#v", result)
				}
				if ctx.Dialogue[1].Seq != 4 {
					t.Fatalf("result seq = %d, want actual result's seq 4", ctx.Dialogue[1].Seq)
				}
				if mode == "snapshot" && len(diagnostics) == 0 {
					t.Error("damaged snapshot was repaired without a diagnostic")
				}
				if want == nil {
					want = ctx
				} else if !reflect.DeepEqual(ctx, want) {
					t.Fatalf("%s differs from full replay", mode)
				}
				// A repaired snapshot must stand alone without its original event log.
				encoded, err := json.Marshal(&SaveFile{AsOf: 4, Context: ctx})
				if err != nil {
					t.Fatal(err)
				}
				var saved SaveFile
				if err := json.Unmarshal(encoded, &saved); err != nil {
					t.Fatal(err)
				}
				if got := saved.Restore(nil); string(lateResultJSON(t, got)) != string(lateResultJSON(t, ctx)) {
					t.Fatal("repair did not survive save/load")
				}
			})
		}
	}
}

func TestRestoredLateResultIsSingleAndAdjacentOnAnthropicWire(t *testing.T) {
	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
		}
		if bad := adjacencyViolations(t, string(body)); len(bad) != 0 {
			t.Errorf("nonadjacent tool results: %v", bad)
		}
		var req struct {
			Messages []struct {
				Content []struct {
					Type      string `json:"type"`
					ToolUseID string `json:"tool_use_id"`
				} `json:"content"`
			} `json:"messages"`
		}
		if err := json.Unmarshal(body, &req); err != nil {
			t.Error(err)
		}
		count := 0
		for _, msg := range req.Messages {
			for _, part := range msg.Content {
				if part.Type == "tool_result" && part.ToolUseID == "late" {
					count++
				}
			}
		}
		if count != 1 {
			t.Errorf("vendor received %d results for late, want exactly 1", count)
		}
		if !strings.Contains(string(body), "actual tool output") || strings.Contains(string(body), "unknown outcome") {
			t.Error("vendor did not receive only the real output")
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":"msg_1","type":"message","role":"assistant","model":"claude-sonnet-5","content":[{"type":"text","text":"ok"}],"stop_reason":"end_turn","usage":{"input_tokens":1,"output_tokens":1}}`)
	}))
	defer srv.Close()
	e := &Engine{Log: NewLog(), Ctx: lateResultSave(t, "snapshot", false).Restore(nil), Agent: testHost{t}, HTTP: srv.Client(), Cfg: common.Config{Model: "claude-sonnet-5", Vendor: common.VendorAnthropic, Surface: common.SurfaceForModel("claude-sonnet-5", common.VendorAnthropic), BaseURL: srv.URL, APIKey: "test"}}
	if _, err := e.Turn(common.StreamCallbacks{}); err != nil {
		t.Fatal(err)
	}
	if requests != 1 {
		t.Fatalf("got %d requests, want 1", requests)
	}
}

func TestLostResultProvenanceSurvivesSnapshotWithoutLog(t *testing.T) {
	sf := lateResultSave(t, "replay", false)
	returned := sf.Log[3]
	sf.Log = sf.Log[:3]
	ctx := sf.Restore(nil)
	data, err := json.Marshal(&SaveFile{AsOf: 3, Context: ctx})
	if err != nil {
		t.Fatal(err)
	}
	var saved SaveFile
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatal(err)
	}
	saved.Log = []common.Event{returned}
	got := saved.Restore(nil)
	want := lateResultSave(t, "replay", false).Restore(nil)
	if string(lateResultJSON(t, got)) != string(lateResultJSON(t, want)) {
		t.Fatal("late result did not replace placeholder after snapshot-only round trip")
	}
}

func TestLostResultDoesNotOverwriteAnActualToolError(t *testing.T) {
	ctx := common.NewContext()
	ctx.Dialogue = []common.Entry{lateResultCall()}
	actual := common.Event{Seq: 2, Type: common.ToolReturned, Tool: &common.ToolData{CallID: "late", IsError: true, Parts: common.PartList{common.TextPart{Text: "permission denied"}}}}
	if err := Apply(ctx, actual); err != nil {
		t.Fatal(err)
	}
	before := lateResultJSON(t, ctx)
	loss := common.Event{Seq: 3, Type: common.ToolResultLost, Tool: &common.ToolData{CallID: "late"}}
	if err := Apply(ctx, loss); err != nil {
		t.Fatal(err)
	}
	after := lateResultJSON(t, ctx)
	if string(before) != string(after) {
		t.Fatal("loss event changed an already answered call")
	}
}

func lateResultCall() common.Entry {
	entry := callEntry(1, "late", "search_files")
	call := entry.Parts[0].(common.ToolCallPart)
	call.From = common.Provenance{Vendor: common.VendorAnthropic, Model: "claude-sonnet-5", Surface: common.SurfaceForModel("claude-sonnet-5", common.VendorAnthropic)}
	entry.Parts[0] = call
	return entry
}

func lateResultJSON(t *testing.T, v any) []byte {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
