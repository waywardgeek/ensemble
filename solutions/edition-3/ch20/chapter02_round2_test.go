package ensemble_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"ensemble"
	"ensemble/internal/common"
)

func r2Agent(t *testing.T, c ensemble.Config) ensemble.Agent {
	t.Helper()
	if c.Model == "" {
		c.Model = "routing-alias"
	}
	a, err := ensemble.New(nil).NewAgent(c)
	if err != nil {
		t.Fatal(err)
	}
	return a
}
func r2Append(t *testing.T, a ensemble.Agent, events ...common.Event) {
	t.Helper()
	for _, e := range events {
		if err := a.Log().Append(e); err != nil {
			t.Fatal(err)
		}
	}
}
func r2Dump(t *testing.T, a ensemble.Agent) []byte {
	t.Helper()
	var b bytes.Buffer
	if err := a.Log().Dump(&b); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}
func r2Render(t *testing.T, a ensemble.Agent) []byte {
	t.Helper()
	b, err := a.Render()
	if err != nil {
		t.Fatal(err)
	}
	return b
}
func r2From(v common.Vendor) common.Provenance {
	s := common.Messages
	if v == common.OpenAI {
		s = common.ChatCompletions
	}
	if v == common.Gemini {
		s = common.GenerateContent
	}
	return common.Provenance{Vendor: v, Model: "actual-model", Surface: s}
}
func r2Message(actor common.Actor, text string) common.Event {
	return common.Event{Type: common.MessageReceived, Message: &common.MessageData{Actor: actor, Parts: common.PartList{common.Text(text)}}}
}
func r2Response(v common.Vendor, parts common.PartList, u common.Usage) common.Event {
	return common.Event{Type: common.ResponseEnded, Response: &common.ResponseData{From: r2From(v), Parts: parts, Usage: u}}
}

func TestRound2LiveAccountingIsNotReplay(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		fmt.Fprint(w, `{"model":"actual-model","content":[{"type":"text","text":"paid answer"}],"usage":{"input_tokens":4,"output_tokens":2,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}`)
	}))
	defer server.Close()
	c := ensemble.Config{Model: "routing-alias", BaseURL: server.URL, APIKey: "local-fixture"}
	a := r2Agent(t, c)
	r2Append(t, a, r2Message(common.Human, "prior"), r2Response(common.Anthropic, common.PartList{common.Text("prior answer")}, common.Usage{Input: 11, Output: 7}))
	prior := r2Dump(t, a)
	if a.Usage() != (common.Usage{}) || a.Log().Context().Usage != (common.Usage{Input: 11, Output: 7}) {
		t.Fatal("replay charged live usage")
	}
	if _, err := a.Ask(context.Background(), "new"); err != nil {
		t.Fatal(err)
	}
	if a.Usage() != (common.Usage{Input: 4, Output: 2}) || a.Log().Context().Usage != (common.Usage{Input: 15, Output: 9}) {
		t.Fatalf("live=%+v historical=%+v", a.Usage(), a.Log().Context().Usage)
	}
	paid := r2Dump(t, a)
	if err := a.Log().Load(bytes.NewReader(prior)); err != nil {
		t.Fatal(err)
	}
	if a.Usage() != (common.Usage{Input: 4, Output: 2}) {
		t.Fatal("loading erased live costs")
	}
	if _, err := a.ParseResponse(200, []byte(`{"content":[{"type":"text","text":"offline"}],"usage":{"input_tokens":99,"output_tokens":99}}`)); err != nil {
		t.Fatal(err)
	}
	_ = r2Render(t, a)
	if a.Usage() != (common.Usage{Input: 4, Output: 2}) || calls != 1 {
		t.Fatal("offline operation charged usage or called HTTP")
	}
	fresh := r2Agent(t, c)
	if err := fresh.Log().Load(bytes.NewReader(paid)); err != nil {
		t.Fatal(err)
	}
	if fresh.Usage() != (common.Usage{}) || fresh.Log().Context().Usage != (common.Usage{Input: 15, Output: 9}) {
		t.Fatal("replay mixed live and historical totals")
	}
}

func r2AssertTail(t *testing.T, v common.Vendor, b []byte) {
	t.Helper()
	var req map[string]json.RawMessage
	if err := json.Unmarshal(b, &req); err != nil {
		t.Fatal(err)
	}
	field := "messages"
	if v == common.Gemini {
		field = "contents"
	}
	var messages []map[string]json.RawMessage
	if err := json.Unmarshal(req[field], &messages); err != nil || len(messages) == 0 {
		t.Fatalf("invalid request: %s", b)
	}
	if !bytes.Contains(mustJSON(t, messages[len(messages)-1]), []byte("fresh-tail")) {
		t.Fatalf("ephemera not freshest: %s", b)
	}
	if bytes.Contains(req["system"], []byte("fresh-tail")) || bytes.Contains(req["systemInstruction"], []byte("fresh-tail")) {
		t.Fatal("ephemera contaminated stable system prefix")
	}
	if v == common.OpenAI {
		if bytes.Contains(mustJSON(t, messages[0]), []byte("fresh-tail")) {
			t.Fatal("ephemera in first system message")
		}
		if string(messages[len(messages)-1]["role"]) != `"system"` {
			t.Fatal("OpenAI ephemeral instruction lost native role")
		}
	}
}
func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
func TestRound2EphemeraTailFailureAndReplay(t *testing.T) {
	for _, v := range []common.Vendor{common.Anthropic, common.OpenAI, common.Gemini} {
		t.Run(v.String(), func(t *testing.T) {
			var wire []byte
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { wire, _ = io.ReadAll(r.Body); w.WriteHeader(503) }))
			defer server.Close()
			c := ensemble.Config{Vendor: v, Model: "routing-alias", BaseURL: server.URL, APIKey: "local-fixture"}
			a := r2Agent(t, c)
			r2Append(t, a, r2Message(common.Human, "old"), r2Response(v, common.PartList{common.Text("old reply")}, common.Usage{}), r2Message(common.System, "fresh-tail"))
			before := r2Dump(t, a)
			r2AssertTail(t, v, r2Render(t, a))
			if !bytes.Equal(before, r2Dump(t, a)) {
				t.Fatal("render mutated raw log")
			}
			if _, err := a.Ask(context.Background(), "new question"); err == nil {
				t.Fatal("expected HTTP failure")
			}
			r2AssertTail(t, v, wire)
			ctx := a.Log().Context()
			if len(ctx.Ephemera) != 0 || ctx.Turn != common.Idle {
				t.Fatalf("not consumed on failure: %+v", ctx)
			}
			var causal []common.Entry
			for _, entry := range ctx.Dialogue {
				switch entry.Kind {
				case "", common.KindDialogue:
					causal = append(causal, entry)
				case common.KindSkill, common.KindTools, common.KindHandoff: // typed survivors are not causal dialogue
				default:
					t.Fatalf("unknown projected entry kind %q", entry.Kind)
				}
			}
			if len(causal) != 3 {
				t.Fatalf("causal dialogue count=%d, want 3", len(causal))
			}
			for i, want := range []string{"old", "old reply", "new question"} {
				actor := common.Human
				if i == 1 {
					actor = common.Assistant
				}
				if causal[i].Actor != actor || len(causal[i].Parts) != 1 || causal[i].Parts[0].Text != want {
					t.Fatalf("causal dialogue %d changed: %+v", i, causal[i])
				}
			}
			r2Append(t, a, common.Event{Type: common.MessageReceived, Message: &common.MessageData{Actor: common.System, Parts: common.PartList{}}})
			if bytes.Contains(r2Render(t, a), []byte("fresh-tail")) {
				t.Fatal("empty queue resurrected delivered ephemera")
			}
			fresh := r2Agent(t, c)
			if err := fresh.Log().Load(bytes.NewReader(r2Dump(t, a))); err != nil {
				t.Fatal(err)
			}
			if bytes.Contains(r2Render(t, fresh), []byte("fresh-tail")) {
				t.Fatal("replay resurrected delivered ephemera")
			}
		})
	}
}

func TestRound2ErrorPreservesProjectedFacts(t *testing.T) {
	for _, state := range []common.TurnState{common.Idle, common.InputPending, common.InFlight, common.ToolsPending} {
		t.Run(string(state), func(t *testing.T) {
			a := r2Agent(t, ensemble.Config{})
			r2Append(t, a, r2Response(common.Anthropic, common.PartList{common.Text("retained")}, common.Usage{Input: 2}))
			if state != common.Idle {
				r2Append(t, a, r2Message(common.Human, "pending"))
			}
			if state == common.InFlight || state == common.ToolsPending {
				r2Append(t, a, common.Event{Type: common.RequestSent, Request: &common.RequestData{To: r2From(common.Anthropic)}})
			}
			if state == common.ToolsPending {
				from := r2From(common.Anthropic)
				r2Append(t, a, r2Response(common.Anthropic, common.PartList{{Type: "tool_call", CallID: "c", Name: "same", Args: json.RawMessage(`{}`), From: &from}}, common.Usage{}))
			}
			before := a.Log().Context()
			r2Append(t, a, common.Event{Type: common.ErrorOccurred, Error: &common.ErrorData{Message: "observed failure"}})
			after := a.Log().Context()
			if !reflect.DeepEqual(before.Dialogue, after.Dialogue) || !reflect.DeepEqual(before.Ephemera, after.Ephemera) || before.Usage != after.Usage || !reflect.DeepEqual(before.Pending, after.Pending) {
				t.Fatal("error changed facts")
			}
			expected := state
			if state == common.InFlight {
				expected = common.Idle
			}
			if after.Turn != expected {
				t.Fatalf("turn %s want %s", after.Turn, expected)
			}
		})
	}
}

func TestRound2CallIDsAreStableRenderOnly(t *testing.T) {
	for _, v := range []common.Vendor{common.Anthropic, common.OpenAI, common.Gemini} {
		for _, missing := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/missing=%t", v, missing), func(t *testing.T) {
				a := r2Agent(t, ensemble.Config{Vendor: v})
				from := r2From(common.Gemini)
				ids := []string{"first-id", "second-id"}
				if missing {
					ids = []string{"", ""}
				}
				r2Append(t, a, r2Response(common.Gemini, common.PartList{{Type: "tool_call", CallID: ids[0], Name: "same", Args: json.RawMessage(`{"n":1}`), From: &from}, {Type: "tool_call", CallID: ids[1], Name: "same", Args: json.RawMessage(`{"n":2}`), From: &from}}, common.Usage{}))
				if !missing {
					for _, id := range []string{"second-id", "first-id"} {
						r2Append(t, a, common.Event{Type: common.ToolReturned, Tool: &common.ToolData{CallID: id, Parts: common.PartList{common.Text("result-" + id)}}})
					}
				}
				before := r2Dump(t, a)
				wire := r2Render(t, a)
				if !bytes.Equal(wire, r2Render(t, a)) || !bytes.Equal(before, r2Dump(t, a)) {
					t.Fatal("render unstable or mutated log")
				}
				var value any
				if err := json.Unmarshal(wire, &value); err != nil {
					t.Fatal(err)
				}
				calls := []string{}
				results := map[string]string{}
				var walk func(any)
				walk = func(x any) {
					switch x := x.(type) {
					case map[string]any:
						if x["type"] == "tool_use" {
							calls = append(calls, fmt.Sprint(x["id"]))
						}
						if x["type"] == "function" {
							calls = append(calls, fmt.Sprint(x["id"]))
						}
						if fc, ok := x["functionCall"].(map[string]any); ok {
							calls = append(calls, fmt.Sprint(fc["id"]))
						}
						if fr, ok := x["functionResponse"].(map[string]any); ok {
							results[fmt.Sprint(fr["id"])] = fmt.Sprint(fr["response"])
						}
						if x["type"] == "tool_result" {
							results[fmt.Sprint(x["tool_use_id"])] = fmt.Sprint(x["content"])
						}
						if x["role"] == "tool" {
							results[fmt.Sprint(x["tool_call_id"])] = fmt.Sprint(x["content"])
						}
						for _, v := range x {
							walk(v)
						}
					case []any:
						for _, v := range x {
							walk(v)
						}
					}
				}
				// Chapter 3 declarations also have type:function. Inspect actual
				// message calls/results, not the separate tool-declaration catalog.
				request := value.(map[string]any)
				if v == ensemble.Gemini {
					walk(request["contents"])
				} else {
					walk(request["messages"])
				}
				if len(calls) != 2 || calls[0] == "" || calls[1] == "" || calls[0] == "<nil>" || calls[1] == "<nil>" || calls[0] == calls[1] {
					t.Fatalf("invalid IDs %v: %s", calls, wire)
				}
				if !missing {
					if !reflect.DeepEqual(calls, ids) {
						t.Fatalf("IDs changed: %v", calls)
					}
					for _, id := range ids {
						if !strings.Contains(results[id], "result-"+id) {
							t.Fatalf("same-name result mispaired: %+v", results)
						}
					}
				}
			})
		}
	}
}

func TestRound2MissingProducerIsNotAliasProof(t *testing.T) {
	for _, returned := range []bool{false, true} {
		for _, resolved := range []bool{false, true} {
			t.Run(fmt.Sprintf("returned=%t/resolved=%t", returned, resolved), func(t *testing.T) {
				c := ensemble.Config{Model: "routing-alias"}
				if resolved {
					c.ResolvedModel = "routing-alias"
				}
				a := r2Agent(t, c)
				model := ""
				if returned {
					model = `"model":"routing-alias",`
				}
				events, err := a.ParseResponse(200, []byte(`{`+model+`"content":[{"type":"text","text":"portable"},{"type":"thinking","thinking":"private-thought","signature":"private-signature"}],"usage":{"input_tokens":2,"output_tokens":1}}`))
				if err != nil {
					t.Fatal(err)
				}
				for _, e := range events {
					if e.Response != nil && !returned && e.Response.From.Model != "" {
						t.Fatal("route fabricated producer identity")
					}
				}
				r2Append(t, a, events...)
				wire := r2Render(t, a)
				if !bytes.Contains(wire, []byte("portable")) {
					t.Fatal("lost portable text")
				}
				wantPrivate := returned && resolved
				if bytes.Contains(wire, []byte("private-thought")) != wantPrivate {
					t.Fatalf("unproven opaque replay: %s", wire)
				}
				fresh := r2Agent(t, c)
				if err := fresh.Log().Load(bytes.NewReader(r2Dump(t, a))); err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(wire, r2Render(t, fresh)) {
					t.Fatal("unknown identity changed on replay")
				}
			})
		}
	}
}

func TestRound2KnownChargesSurviveResponseErrors(t *testing.T) {
	type fixture struct {
		vendor                                common.Vendor
		content, broken, usage, partial, zero string
		paid                                  common.Usage
	}
	for _, f := range []fixture{
		{common.Anthropic, `"content":[{"type":"text","text":"answer"}]`, `"content":[{"type":"text","text":"before"},{"type":"text"}]`, `"usage":{"input_tokens":10,"output_tokens":4,"cache_read_input_tokens":3,"cache_creation_input_tokens":2}`, `"usage":{"input_tokens":9}`, `"usage":{"input_tokens":0,"output_tokens":0,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}`, common.Usage{Input: 10, Output: 4, CacheRead: 3, CacheWrite: 2}},
		{common.OpenAI, `"choices":[{"message":{"content":"answer"}}]`, `"choices":[{"message":{"content":"before","tool_calls":[{"id":"bad","type":"function","function":{"name":"same","arguments":"{"}}]}}]`, `"usage":{"prompt_tokens":15,"completion_tokens":4,"prompt_tokens_details":{"cached_tokens":3,"cache_write_tokens":2}}`, `"usage":{"prompt_tokens":9}`, `"usage":{"prompt_tokens":0,"completion_tokens":0,"prompt_tokens_details":{"cached_tokens":0}}`, common.Usage{Input: 12, Output: 4, CacheRead: 3}},
		{common.Gemini, `"candidates":[{"content":{"parts":[{"text":"answer"}]}}]`, `"candidates":[{"content":{"parts":[{"text":"before"},{"text":123}]}}]`, `"usageMetadata":{"promptTokenCount":13,"candidatesTokenCount":2,"thoughtsTokenCount":2,"cachedContentTokenCount":3}`, `"usageMetadata":{"promptTokenCount":9}`, `"usageMetadata":{"promptTokenCount":0,"candidatesTokenCount":0,"cachedContentTokenCount":0}`, common.Usage{Input: 10, Output: 4, CacheRead: 3}},
	} {
		for _, scenario := range []string{"missing", "partial", "content-error", "measured-zero"} {
			t.Run(f.vendor.String()+"/"+scenario, func(t *testing.T) {
				content, usage, want := f.content, "", common.UnknownUsage()
				wantError := true
				switch scenario {
				case "partial":
					usage = f.partial
					want = common.Usage{Input: 9, Output: -1}
				case "content-error":
					content = f.broken
					usage = f.usage
					want = f.paid
				case "measured-zero":
					usage = f.zero
					wantError = false
					want = common.Usage{}
				}
				body := "{" + content
				if usage != "" {
					body += "," + usage
				}
				body += "}"
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, body) }))
				defer server.Close()
				a := r2Agent(t, ensemble.Config{Vendor: f.vendor, BaseURL: server.URL, APIKey: "local-fixture"})
				_, err := a.Ask(context.Background(), "question")
				if (err != nil) != wantError {
					t.Fatalf("error=%v, wantError=%t", err, wantError)
				}
				if a.Usage() != want || a.Log().Context().Usage != want {
					t.Fatalf("lost/invented charges: live=%+v history=%+v want=%+v", a.Usage(), a.Log().Context().Usage, want)
				}
				if wantError && len(a.History()) != 0 {
					t.Fatal("failed answer entered compatibility History")
				}
				var response *common.ResponseData
				for _, event := range a.Log().Events() {
					if event.Type == common.ResponseEnded {
						response = event.Response
					}
				}
				if response == nil {
					t.Fatal("response accounting disappeared")
				}
				missing := strings.Join(response.UsageMissing, ",")
				if (scenario == "missing" || scenario == "partial") && !strings.Contains(missing, "output") {
					t.Fatal("missing output disguised as measured zero")
				}
				if scenario == "missing" && !strings.Contains(missing, "input") {
					t.Fatal("missing input disguised as measured zero")
				}
				if scenario == "measured-zero" && len(response.UsageMissing) != 0 {
					t.Fatal("measured zero marked unknown")
				}
				if scenario == "content-error" && (len(response.Parts) == 0 || response.Parts[0].Text != "before") {
					t.Fatal("later parser error discarded earlier observed text")
				}
				fresh := r2Agent(t, ensemble.Config{Vendor: f.vendor})
				if err := fresh.Log().Load(bytes.NewReader(r2Dump(t, a))); err != nil {
					t.Fatal(err)
				}
				if fresh.Usage() != (common.Usage{}) || fresh.Log().Context().Usage != want {
					t.Fatal("paid error replay charged live counter or lost facts")
				}
			})
		}
	}
}

func TestRound2SignedGeminiTextRemainsPortable(t *testing.T) {
	for _, producerKnown := range []bool{false, true} {
		source := r2Agent(t, ensemble.Config{Vendor: common.Gemini})
		model := ""
		if producerKnown {
			model = `"modelVersion":"actual-model",`
		}
		body := `{` + model + `"candidates":[{"content":{"parts":[{"text":"portable-visible","thoughtSignature":"text-signature"},{"text":"private-thought","thought":true,"thoughtSignature":"private-signature"}]}}],"usageMetadata":{"promptTokenCount":2,"candidatesTokenCount":1}}`
		events, err := source.ParseResponse(200, []byte(body))
		if err != nil {
			t.Fatal(err)
		}
		r2Append(t, source, events...)
		saved := r2Dump(t, source)
		for _, target := range []struct {
			name            string
			vendor          common.Vendor
			resolved        string
			surfaceMismatch bool
		}{
			{"anthropic", common.Anthropic, "actual-model", false}, {"openai", common.OpenAI, "actual-model", false}, {"gemini-foreign-model", common.Gemini, "different-model", false}, {"gemini-unresolved", common.Gemini, "", false}, {"gemini-exact", common.Gemini, "actual-model", false}, {"gemini-foreign-surface", common.Gemini, "actual-model", true},
		} {
			t.Run(fmt.Sprintf("known=%t/%s", producerKnown, target.name), func(t *testing.T) {
				a := r2Agent(t, ensemble.Config{Vendor: target.vendor, Model: "routing-alias", ResolvedModel: target.resolved})
				log := saved
				if target.surfaceMismatch {
					log = bytes.ReplaceAll(saved, []byte(`"surface":"generate_content"`), []byte(`"surface":"messages"`))
				}
				if err := a.Log().Load(bytes.NewReader(log)); err != nil {
					t.Fatal(err)
				}
				before := r2Dump(t, a)
				wire := r2Render(t, a)
				if !bytes.Contains(wire, []byte("portable-visible")) {
					t.Fatalf("lost ordinary signed text: %s", wire)
				}
				compatible := producerKnown && target.name == "gemini-exact"
				for _, private := range []string{"text-signature", "private-signature", "private-thought"} {
					if bytes.Contains(wire, []byte(private)) != compatible {
						t.Fatalf("wrong compatibility for %s: %s", private, wire)
					}
				}
				if bytes.Count(wire, []byte("portable-visible")) != 1 {
					t.Fatal("ordinary signed text duplicated")
				}
				if !bytes.Equal(before, r2Dump(t, a)) {
					t.Fatal("foreign render changed source log")
				}
			})
		}
	}
}
