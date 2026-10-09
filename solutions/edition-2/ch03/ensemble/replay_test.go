package ensemble_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ensemble/ensemble"
)

func loaded(t *testing.T, v ensemble.Vendor, log string) ensemble.Agent {
	t.Helper()
	a, err := ensemble.New(io.Discard).NewAgent(ensemble.Config{Vendor: v, Model: "model", ResolvedModel: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if err = a.History().Load(strings.NewReader(log)); err != nil {
		t.Fatal(err)
	}
	return a
}

const conversation = `{"seq":1,"type":"message_received","message":{"actor":"human","parts":[{"type":"text","text":"hello"}]}}
{"seq":2,"type":"request_sent","request":{"to":{"vendor":"gemini","model":"model","surface":"generate_content"}}}
{"seq":3,"type":"response_ended","response":{"from":{"vendor":"gemini","model":"model","surface":"generate_content"},"parts":[{"type":"tool_call","call_id":"one","name":"read","args":{"file":"a"},"from":{"vendor":"gemini","model":"model","surface":"generate_content"},"opaque":"signature"},{"type":"tool_call","call_id":"two","name":"read","args":{"file":"b"},"from":{"vendor":"gemini","model":"model","surface":"generate_content"}},{"type":"text","text":"decision"}],"usage":{"input":4,"cache_write":2,"cache_read":3,"output":5}}}
`
const firstResult = `{"seq":4,"type":"tool_returned","tool":{"call_id":"one","parts":[{"type":"text","text":"secret"},{"type":"blob","mime":"text/plain","ref":{"kind":3,"locator":"result:one"}}]}}
`
const lastResult = `{"seq":5,"type":"tool_returned","tool":{"call_id":"two","parts":[{"type":"text","text":"second"}]}}
`

func TestReducerStateAndRedactionReplay(t *testing.T) {
	a := loaded(t, ensemble.Gemini, conversation)
	if a.History().Context().Turn != "tools_pending" {
		t.Fatal("tool calls did not make tools pending")
	}
	a = loaded(t, ensemble.Gemini, conversation+firstResult)
	if a.History().Context().Turn != "tools_pending" {
		t.Fatal("first result completed two-call batch")
	}
	a = loaded(t, ensemble.Gemini, conversation+firstResult+lastResult)
	if a.History().Context().Turn != "input_pending" {
		t.Fatal("last result did not complete batch")
	}
	if a.History().Context().Usage != (ensemble.Usage{Input: 4, CacheWrite: 2, CacheRead: 3, Output: 5}) {
		t.Fatal("replayed usage lost")
	}

	for _, level := range []string{"redact_result", "redact_tool", "redact_dialogue", "redact_summary"} {
		log := conversation + firstResult + lastResult + `{"seq":6,"type":"redacted","redact":{"from":1,"to":5,"level":"` + level + `","replacement":[{"type":"text","text":"summary"}]}}` + "\n"
		a = loaded(t, ensemble.Gemini, log)
		entries := a.History().Context().Dialogue
		data, _ := json.Marshal(entries)
		switch level {
		case "redact_result":
			if bytes.Contains(data, []byte("secret")) || !bytes.Contains(data, []byte("result:one")) || !bytes.Contains(data, []byte("tool_call")) {
				t.Fatal("result redaction lost reference/call or retained content")
			}
		case "redact_tool":
			if bytes.Contains(data, []byte("tool_call")) || bytes.Contains(data, []byte("tool_result")) || !bytes.Contains(data, []byte("hello")) {
				t.Fatal("tool filter removed wrong category")
			}
		case "redact_dialogue":
			if bytes.Contains(data, []byte("hello")) || !bytes.Contains(data, []byte("tool_call")) {
				t.Fatal("dialogue filter removed wrong category")
			}
		case "redact_summary":
			if len(entries) != 1 || entries[0].Seq != 1 || entries[0].Actor != "system" || string(data) == "" || strings.Count(string(data), "summary") != 1 {
				t.Fatal("summary did not fold span once")
			}
		}
		var dump bytes.Buffer
		if err := a.History().Dump(&dump); err != nil {
			t.Fatal(err)
		}
		var originalFacts bytes.Buffer
		loaded(t, ensemble.Gemini, conversation+firstResult+lastResult).History().Dump(&originalFacts)
		if !strings.HasPrefix(dump.String(), originalFacts.String()) {
			t.Fatal("redaction mutated append-only facts")
		}
		if !strings.Contains(dump.String(), "secret") {
			t.Fatal("redaction mutated append-only facts")
		}
		b := loaded(t, ensemble.Gemini, dump.String())
		replay, _ := json.Marshal(b.History().Context())
		original, _ := json.Marshal(a.History().Context())
		if !bytes.Equal(replay, original) {
			t.Fatal("dump replay changed context")
		}
	}
}

func TestLoadRefusesUnknownInterpretedData(t *testing.T) {
	good := `{"seq":1,"type":"message_received","message":{"actor":"human","parts":[{"type":"text","text":"kept"}]}}` + "\n"
	for _, bad := range []string{
		`{"log_version":2}`,
		strings.Replace(good, "message_received", "mystery", 1),
		strings.Replace(good, "human", "alien", 1),
		strings.Replace(good, `"type":"text"`, `"type":"mystery"`, 1),
		strings.Replace(conversation, "gemini", "unknown", 1),
		strings.Replace(conversation, "generate_content", "other", 1),
		`{"seq":1,"type":"message_received","message":{"actor":"human","parts":[{"type":"blob","path":"old"}]}}`,
		`{"seq":1,"type":"message_received","message":{"actor":"human","parts":[{"type":"blob","path":"old","ref":{"kind":1,"locator":"new"}}]}}`,
		good + good,
	} {
		a := loaded(t, ensemble.Anthropic, good)
		if err := a.History().Load(strings.NewReader(bad)); err == nil {
			t.Fatalf("bad log accepted: %s", bad)
		}
		if len(a.History().Context().Dialogue) != 1 || a.History().Context().Dialogue[0].Parts[0].Text != "kept" {
			t.Fatal("failed load changed retained context")
		}
	}
}

func TestOpaqueIdentityAndEmptyReply(t *testing.T) {
	// Signatures are compared on all identity axes; visible call arguments still
	// survive vendor/model changes. Repeated rendering cannot consume the parts.
	for _, variant := range []struct {
		old, new  string
		signature bool
	}{
		{"no-change", "no-change", true}, {`"model":"model"`, `"model":"other"`, false}, {"generate_content", "messages", false},
	} {
		a := loaded(t, ensemble.Gemini, strings.ReplaceAll(conversation, variant.old, variant.new))
		one, err := a.Engine().Render(a.History().Context())
		if err != nil {
			t.Fatal(err)
		}
		two, _ := a.Engine().Render(a.History().Context())
		if !bytes.Equal(one, two) {
			t.Fatal("render changed bytes")
		}
		if bytes.Contains(one, []byte(`"signature"`)) != variant.signature {
			t.Fatal("opaque crossed producer identity")
		}
	}
	// The faithful HTTP boundary preserves the live bug's actual shape: content
	// is present but empty, and the next request must contain that same string.
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 2 {
			var request struct {
				Messages []struct {
					Role    string
					Content json.RawMessage
				}
			}
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Error(err)
			}
			if len(request.Messages) != 4 || string(request.Messages[2].Content) != `""` {
				t.Error("empty reply did not round-trip as a string")
			}
		}
		io.WriteString(w, `{"model":"model","choices":[{"message":{"role":"assistant","content":""},"finish_reason":"length"}],"usage":{"prompt_tokens":3,"completion_tokens":5}}`)
	}))
	defer server.Close()
	a, err := ensemble.New(io.Discard).NewAgent(ensemble.Config{Vendor: ensemble.OpenAI, BaseURL: server.URL, Model: "model", APIKey: "fake"})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if _, err := a.Ask(context.Background(), "hello"); err != nil {
			t.Fatal(err)
		}
	}
	if calls != 2 {
		t.Fatal("round-trip path not exercised")
	}
}

func TestTurnTransitionsAndPendingEphemera(t *testing.T) {
	human := `{"seq":1,"type":"message_received","message":{"actor":"human","parts":[{"type":"text","text":"hello"}]}}` + "\n"
	sent := `{"seq":2,"type":"request_sent","request":{"to":{"vendor":"anthropic","model":"model","surface":"messages"}}}` + "\n"
	done := `{"seq":3,"type":"response_ended","response":{"parts":[{"type":"text","text":"hi"}],"from":{"vendor":"anthropic","model":"model","surface":"messages"},"usage":{}}}` + "\n"
	failed := `{"seq":3,"type":"error_occurred","error":{"message":"HTTP 429"}}` + "\n"
	for _, sample := range []struct{ log, want string }{
		{"", "idle"}, {human, "input_pending"}, {human + sent, "in_flight"}, {human + sent + done, "idle"}, {human + sent + failed, "idle"},
		{strings.Replace(done, `"seq":3`, `"seq":1`, 1), "idle"},
		{strings.Split(conversation, "\n")[2] + "\n", "idle"},
	} {
		a := loaded(t, ensemble.Anthropic, sample.log)
		if a.History().Context().Turn != sample.want {
			t.Fatalf("turn got %s want %s", a.History().Context().Turn, sample.want)
		}
	}
	a := loaded(t, ensemble.Anthropic, "")
	if err := a.Ephemeral("status-once"); err != nil {
		t.Fatal(err)
	}
	if len(a.History().Context().Dialogue) != 0 {
		t.Fatal("ephemera retained as dialogue")
	}
	var dump bytes.Buffer
	if err := a.History().Dump(&dump); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(dump.String(), "status-once") {
		t.Fatal("ephemeral arrival was not logged")
	}
	b := loaded(t, ensemble.Anthropic, dump.String())
	body, err := b.Engine().Render(b.History().Context())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(body, []byte("status-once")) {
		t.Fatal("pending ephemeral lost on replay")
	}
}

func TestRefCarryforwardAndMissingCallIDs(t *testing.T) {
	for kind := 1; kind <= 3; kind++ {
		ref := `"kind":` + string(rune('0'+kind)) + `,"locator":"result:one"`
		log := strings.Replace(firstResult, `"kind":3,"locator":"result:one"`, ref, 1)
		a := loaded(t, ensemble.Gemini, conversation+log+lastResult+`{"seq":6,"type":"redacted","redact":{"from":4,"to":4,"level":"redact_result"}}`+"\n")
		result := a.History().Context().Dialogue[2].Parts[0]
		if int(result.Parts[1].Ref.Kind) != kind || result.Parts[1].Ref.Locator != "result:one" {
			t.Fatal("redaction changed exact Ref")
		}
		body, err := a.Engine().Render(a.History().Context())
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Contains(body, []byte("secret")) || !bytes.Contains(body, []byte("redacted")) {
			t.Fatal("redacted projection was not rendered")
		}
	}
	noIDs := strings.ReplaceAll(strings.ReplaceAll(conversation, `"call_id":"one",`, ""), `"call_id":"two",`, "")
	a := loaded(t, ensemble.OpenAI, noIDs)
	first, err := a.Engine().Render(a.History().Context())
	if err != nil {
		t.Fatal(err)
	}
	second, _ := a.Engine().Render(a.History().Context())
	if !bytes.Equal(first, second) {
		t.Fatal("synthesized IDs are nondeterministic")
	}
	var request struct {
		Messages []struct {
			ToolCalls []struct{ ID string } `json:"tool_calls"`
		}
	}
	if err := json.Unmarshal(first, &request); err != nil {
		t.Fatal(err)
	}
	calls := request.Messages[2].ToolCalls
	if len(calls) != 2 || calls[0].ID == "" || calls[0].ID == calls[1].ID {
		t.Fatal("missing call IDs were not distinct")
	}
}

func TestGeminiCaptureAndResultName(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		// Chapter 3 answers even undeclared calls with an error. The provider
		// must finish that turn instead of repeating Chapter 2's call forever.
		if requests > 1 {
			io.WriteString(w, `{"candidates":[{"content":{"parts":[{"text":"done"}]}}],"usageMetadata":{"promptTokenCount":0}}`)
			return
		}
		// A routing alias deliberately differs from modelVersion, and the numeric
		// argument exceeds float64's exact integer range. Neither fact is guessed.
		io.WriteString(w, `{"modelVersion":"actual-model","candidates":[{"content":{"parts":[{"functionCall":{"id":"call","name":"read","args":{"n":9007199254740993}},"thoughtSignature":"private-signature"}]},"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":3,"candidatesTokenCount":2}}`)
	}))
	defer server.Close()
	a, err := ensemble.New(io.Discard).NewAgent(ensemble.Config{Vendor: ensemble.Gemini, BaseURL: server.URL, APIKey: "fake", Model: "routing-alias", ResolvedModel: "actual-model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.Ask(context.Background(), "read"); err != nil {
		t.Fatal(err)
	}
	if requests != 2 || !a.History().Context().Dialogue[2].Parts[0].IsError {
		t.Fatal("empty registry silently skipped the unknown tool")
	}
	p := a.History().Context().Dialogue[1].Parts[0]
	if p.From.Model != "actual-model" {
		t.Fatal("response producer replaced by routing alias")
	}
	if string(p.Opaque) != `"private-signature"` {
		t.Fatal("call signature lost during capture")
	}
	if string(p.Args) != `{"n":9007199254740993}` {
		t.Fatal("argument number changed during capture")
	}
	body, err := a.Engine().Render(a.History().Context())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(body, []byte(`"thoughtSignature":"private-signature"`)) {
		t.Fatal("captured call signature lost on replay")
	}

	// The name belongs to the call, but Gemini requires it again on the result.
	result := `{"seq":4,"type":"tool_returned","tool":{"call_id":"one","parts":[{"type":"text","text":"ok"}]}}` + "\n"
	b := loaded(t, ensemble.Gemini, conversation+result)
	body, err = b.Engine().Render(b.History().Context())
	if err != nil {
		t.Fatal(err)
	}
	var request struct {
		Contents []struct {
			Parts []struct {
				FunctionResponse *struct{ Name string } `json:"functionResponse"`
			}
		}
	}
	if err := json.Unmarshal(body, &request); err != nil {
		t.Fatal(err)
	}
	response := request.Contents[len(request.Contents)-1].Parts[0].FunctionResponse
	if response == nil || response.Name != "read" {
		t.Fatal("result lost its originating function name")
	}
}

func TestAnthropicInterleavedResultsPrecedeText(t *testing.T) {
	// The log preserves arrival order. Anthropic's merged user message must put
	// both results first even when human text arrived between their returns.
	log := conversation + `{"seq":4,"type":"tool_returned","tool":{"call_id":"one","parts":[{"type":"text","text":"first result"}]}}
{"seq":5,"type":"message_received","message":{"actor":"human","parts":[{"type":"text","text":"keep this instruction"}]}}
{"seq":6,"type":"tool_returned","tool":{"call_id":"two","parts":[{"type":"text","text":"second result"}]}}
`
	a := loaded(t, ensemble.Anthropic, log)
	before, _ := json.Marshal(a.History().Context())
	body, err := a.Engine().Render(a.History().Context())
	if err != nil {
		t.Fatal(err)
	}
	var request struct {
		Messages []struct {
			Role    string
			Content []struct {
				Type    string
				Text    string
				CallID  string `json:"tool_use_id"`
				Content []struct{ Text string }
			}
		}
	}
	if err := json.Unmarshal(body, &request); err != nil {
		t.Fatal(err)
	}
	if len(request.Messages) != 3 || request.Messages[2].Role != "user" {
		t.Fatal("results and human text were not merged into one user message")
	}
	blocks := request.Messages[2].Content
	if len(blocks) != 3 || blocks[0].Type != "tool_result" || blocks[1].Type != "tool_result" || blocks[2].Type != "text" {
		t.Fatalf("tool results must precede text: %s", body)
	}
	if blocks[0].CallID != "one" || blocks[1].CallID != "two" || blocks[2].Text != "keep this instruction" {
		t.Fatal("merge lost result identities or human text")
	}
	if len(blocks[0].Content) != 1 || blocks[0].Content[0].Text != "first result" || len(blocks[1].Content) != 1 || blocks[1].Content[0].Text != "second result" {
		t.Fatal("merge changed tool output")
	}
	after, _ := json.Marshal(a.History().Context())
	if !bytes.Equal(before, after) {
		t.Fatal("wire ordering changed the original context")
	}
}
