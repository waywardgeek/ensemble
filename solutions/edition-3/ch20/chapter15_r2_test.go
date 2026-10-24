package ensemble_test

import (
	"bytes"
	"context"
	"encoding/json"
	"ensemble"
	"ensemble/internal/common"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestChapter15R2DiagnosticSectionBodiesHTTP(t *testing.T) {
	sent := make(chan []byte, 2)
	peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		sent <- b
		fmt.Fprint(w, `{"model":"claude-sonnet-4-6","content":[{"type":"text","text":"ok"}],"usage":{"input_tokens":1,"output_tokens":1,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}`)
	}))
	defer peer.Close()
	dir := t.TempDir()
	a := r2Agent(t, ensemble.Config{Vendor: ensemble.Anthropic, Model: "claude-sonnet-4-6", APIKey: "local-fixture", BaseURL: peer.URL, DataDir: dir, Tools: []string{}})
	defer a.Shutdown()
	const marker = `,"cache_control":{"type":"ephemeral"}`
	tools := "[\n {\"name\":\"find\",\"input_schema\":{ \"type\":\"object\",\"properties\":{\"cache_control\":{\"type\":\"string\"}}}" + marker + "}\n]"
	system := `[{"type":"text","text":"STATIC-SYSTEM"` + marker + `}]`
	messages := `[{ "role":"user", "content":[{"type":"text","text":"old literal \"cache_control\""` + marker + `}]}]`
	body := fmt.Sprintf("{\n \"messages\":%s,\n \"max_tokens\":32,\n \"system\":%s,\n \"model\":\"claude-sonnet-4-6\",\n \"tools\":%s,\n \"temperature\":0.25\n}\n", messages, system, tools)
	options := "{\n \n \"max_tokens\":32,\n \n \"model\":\"claude-sonnet-4-6\",\n \n \"temperature\":0.25\n}\n"
	read := func(name string) []byte {
		t.Helper()
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	send := func(body string) []byte {
		t.Helper()
		if _, err := a.Engine().ExchangeRequest(context.Background(), []byte(body)); err != nil {
			t.Fatal(err)
		}
		if got := <-sent; string(got) != body {
			t.Fatal("transport changed exact request bytes")
		}
		return read("cache-current.txt")
	}
	check := func(report []byte, body, messages string) {
		t.Helper()
		last := -1
		for _, s := range []struct{ name, raw string }{{"TOOLS", tools}, {"SYSTEM", system}, {"MESSAGES", messages}, {"NONPROMPT OPTIONS", options}} {
			want := "=== " + s.name + " ===\n" + strings.Replace(s.raw, marker, "", 1) + "\n"
			at := bytes.Index(report, []byte(want))
			if at < 0 {
				t.Fatalf("missing exact raw %s section: %s", s.name, report)
			}
			if at <= last {
				t.Fatal("sections are not in provider cache order")
			}
			last = at
		}
		at := bytes.Index(report, []byte("Exact sent request:\n"+body))
		if at <= last {
			t.Fatal("exact original wire view missing or preceding ordered section bodies")
		}
	}
	first := send(body)
	check(first, body, messages)
	changed := strings.Replace(body, "old literal", "new literal", 1)
	second := send(changed)
	check(second, changed, strings.Replace(messages, "old literal", "new literal", 1))
	if !bytes.Equal(read("cache-prior.txt"), first) {
		t.Fatal("prior file is not the first actual request report")
	}
	status := a.Engine().Status().Cache
	if status.OptionsChange != "unchanged" || len(status.Sections) != 3 {
		t.Fatalf("wrong section/options comparison: %+v", status)
	}
	for _, s := range status.Sections {
		want := "unchanged"
		if s.Name == "messages" {
			want = "changed"
		}
		if s.Change != want {
			t.Fatalf("%s=%s want %s", s.Name, s.Change, want)
		}
	}
	if !strings.HasPrefix(status.FirstDifference, "messages byte ") || !bytes.Contains(second, []byte("First difference: messages byte ")) || !strings.Contains(status.FirstDifference, "old literal") || !strings.Contains(status.FirstDifference, "new literal") {
		t.Fatalf("first differing content is not inspectable: %s", second)
	}
}

func TestChapter15R2FinalHistoryBoundaryAgentHTTP(t *testing.T) {
	for _, model := range []string{"gpt-5.6", "gpt-5.4"} {
		t.Run(model, func(t *testing.T) {
			const hint = "interleaved hint after the call"
			const result = "complete actual result; literal prompt_cache_breakpoint"
			const handoff = "surviving handoff"
			const transient = "transient status, not history"
			sent := make(chan []byte, 1)
			peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				b, _ := io.ReadAll(r.Body)
				sent <- b
				fmt.Fprintf(w, `{"model":%q,"choices":[{"message":{"content":"done"},"finish_reason":"stop"}],"usage":{"prompt_tokens":10,"completion_tokens":1,"prompt_tokens_details":{"cached_tokens":0,"cache_write_input_tokens":0}}}`, model)
			}))
			defer peer.Close()
			a := r2Agent(t, ensemble.Config{Vendor: ensemble.OpenAI, Model: model, ResolvedModel: model, APIKey: "local-fixture", BaseURL: peer.URL, DataDir: t.TempDir(), Tools: []string{}})
			defer a.Shutdown()
			from := a.Config().Target()
			facts := []common.Event{
				{Type: common.ResponseEnded, Response: &common.ResponseData{From: from, Parts: common.PartList{{Type: "opaque", From: &from, Data: json.RawMessage(`{"reasoning_content":"native reasoning preserved"}`)}, {Type: "tool_call", From: &from, CallID: "c1", Name: "held", Args: json.RawMessage(`{"n":9007199254741001}`)}}}},
				{Type: common.HintReceived, Message: &common.MessageData{Actor: common.Human, Parts: common.PartList{common.Text(hint)}}},
				{Type: common.ToolReturned, Tool: &common.ToolData{CallID: "c1", Name: "held", Parts: common.PartList{common.Text(result)}}},
				{Type: common.MessageReceived, Message: &common.MessageData{Actor: common.System, Parts: common.PartList{common.Text(transient)}}},
			}
			for _, fact := range facts {
				if err := a.Log().Append(fact); err != nil {
					t.Fatal(err)
				}
			}
			before, _ := json.Marshal(a.Log().Context())
			eventsBefore, _ := json.Marshal(a.Log().Events())
			c := a.Log().Context()
			var causal []common.Entry
			for _, entry := range c.Dialogue {
				if entry.Actor != common.System {
					causal = append(causal, entry)
				}
			}
			if len(causal) != 3 || causal[0].Actor != common.Assistant || causal[1].Parts[0].Text != hint || causal[2].Parts[0].CallID != "c1" {
				t.Fatalf("fixture is not causal call/hint/real result: %s", before)
			}
			// The typed handoff projection is already covered by R1. Add it only to
			// this disposable context, not as a fabricated retained-history message.
			c.Handoff = handoff
			contextBefore, _ := json.Marshal(c)
			rendered, err := a.Engine().Render(c)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = a.Engine().ExchangeRequest(context.Background(), rendered); err != nil {
				t.Fatal(err)
			}
			wire := <-sent
			if !bytes.Equal(wire, rendered) {
				t.Fatal("HTTP did not send the rendered bytes")
			}
			var request struct {
				Messages []struct {
					Role       string
					Content    json.RawMessage
					ToolCallID string `json:"tool_call_id"`
					ToolCalls  []struct {
						ID       string
						Function struct{ Arguments string }
					} `json:"tool_calls"`
					Reasoning string `json:"reasoning_content"`
				}
				Options json.RawMessage `json:"prompt_cache_options"`
			}
			if err = json.Unmarshal(wire, &request); err != nil {
				t.Fatal(err)
			}
			if len(request.Messages) != 5 {
				t.Fatalf("invented/removed messages: %s", wire)
			}
			ms := request.Messages
			if ms[1].Role != "assistant" || len(ms[1].ToolCalls) != 1 || ms[1].ToolCalls[0].ID != "c1" || ms[1].ToolCalls[0].Function.Arguments != `{"n":9007199254741001}` || ms[1].Reasoning != "native reasoning preserved" || ms[2].Role != "tool" || ms[2].ToolCallID != "c1" || ms[3].Role != "user" {
				t.Fatalf("call/result adjacency or native facts changed: %s", wire)
			}
			text := func(content json.RawMessage) (string, map[string]bool) {
				t.Helper()
				var s string
				if json.Unmarshal(content, &s) == nil {
					return s, nil
				}
				var blocks []struct {
					Text   string
					Marker json.RawMessage `json:"prompt_cache_breakpoint"`
				}
				if err := json.Unmarshal(content, &blocks); err != nil {
					t.Fatal(err)
				}
				marked := map[string]bool{}
				for _, block := range blocks {
					s += block.Text
					marked[block.Text] = len(block.Marker) > 0
				}
				return s, marked
			}
			resultText, resultMarks := text(ms[2].Content)
			hintText, hintMarks := text(ms[3].Content)
			tail, tailMarks := text(ms[4].Content)
			if resultText != result || hintText != hint || tail != handoff+transient {
				t.Fatalf("content invented, removed, or reordered: %s", wire)
			}
			explicit := model == "gpt-5.6"
			if resultMarks[result] || hintMarks[hint] != explicit || tailMarks[handoff] != explicit || tailMarks[transient] || (len(request.Options) > 0) != explicit {
				t.Errorf("final history boundary is not the later hint, or handoff/tail policy changed: %s", wire)
			}
			after, _ := json.Marshal(a.Log().Context())
			eventsAfter, _ := json.Marshal(a.Log().Events())
			contextAfter, _ := json.Marshal(c)
			if !bytes.Equal(before, after) || !bytes.Equal(eventsBefore, eventsAfter) || !bytes.Equal(contextBefore, contextAfter) {
				t.Fatal("render/send mutated canonical causal history, native provenance, or source context")
			}
		})
	}
}
