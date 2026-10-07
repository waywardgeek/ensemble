package ensemble_test

import (
	"bytes"
	"context"
	"encoding/json"
	"example.com/ensemble"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func loadText(t *testing.T, text string) (*ensemble.Agent, error) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "log.jsonl")
	if err := os.WriteFile(path, []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
	return ensemble.New(nil).Load(path, ensemble.Config{Vendor: "anthropic", Model: "fixture-messages"})
}
func TestPublishedReplayFixtures(t *testing.T) {
	source, err := os.ReadFile("testdata/history.log")
	if err != nil {
		t.Fatal(err)
	}
	a, err := loadText(t, string(source))
	if err != nil {
		t.Fatal(err)
	}
	for _, vendor := range []string{"anthropic", "openai", "gemini"} {
		config := ensemble.Config{Vendor: vendor, Model: "offline"}
		one, err := a.Render(config)
		if err != nil {
			t.Fatal(err)
		}
		two, err := a.Render(config)
		if err != nil || !bytes.Equal(one, two) || !bytes.Contains(one, []byte("port=8080")) {
			t.Fatal("impure or incomplete render")
		}
	}
	redacted := string(source) + `{"seq":6,"type":"redacted","time":"2026-01-01T00:00:05Z","redact":{"from":4,"to":4,"level":"redact_result","reason":"test"}}` + "\n"
	b, err := loadText(t, redacted)
	if err != nil {
		t.Fatal(err)
	}
	for _, vendor := range []string{"anthropic", "openai", "gemini"} {
		data, err := b.Render(ensemble.Config{Vendor: vendor, Model: "offline"})
		if err != nil || bytes.Contains(data, []byte("port=8080")) || !bytes.Contains(data, []byte("call-config")) {
			t.Fatalf("redaction %s: %s %v", vendor, data, err)
		}
	}
	dump, _ := b.Dump()
	if !bytes.Contains(dump, []byte("port=8080")) {
		t.Fatal("redaction changed durable log")
	}
	reloaded, err := loadText(t, string(dump))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(b.Snapshot(), reloaded.Snapshot()) || !reflect.DeepEqual(b.UsageByModel(), reloaded.UsageByModel()) {
		t.Fatal("replay differs or double-counts")
	}
}
func TestMalformedLogs(t *testing.T) {
	base := `{"log_version":1}` + "\n" + `{"seq":1,"type":"message_received","time":"2026-01-01T00:00:00Z","message":{"actor":"human","purpose":"dialogue","parts":[{"type":"text","text":"hi"}]}}` + "\n"
	cases := []string{strings.Replace(base, `"log_version":1`, `"log_version":2`, 1), strings.Replace(base, `"seq":1`, `"seq":0`, 1), strings.Replace(base, `"time":"2026-01-01T00:00:00Z"`, `"time":"bad"`, 1), strings.Replace(base, `"text":"hi"`, `"text":null`, 1), strings.Replace(base, `"actor":"human"`, `"actor":"agent"`, 1), strings.Replace(base, `"type":"text","text":"hi"`, `"type":"blob","mime":"text/plain","ref":{"kind":0,"locator":"x"}`, 1), strings.Replace(base, `"message":`, `"error":null,"message":`, 1)}
	for i, input := range cases {
		if _, err := loadText(t, input); err == nil || !strings.Contains(err.Error(), "line") {
			t.Fatalf("case %d: %v", i, err)
		}
	}
}
func TestEphemeraFailureAndUsageAttribution(t *testing.T) {
	requests := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body bytes.Buffer
		_, _ = body.ReadFrom(r.Body)
		requests = append(requests, body.String())
		if len(requests) == 1 {
			w.WriteHeader(503)
			return
		}
		fmt.Fprint(w, `{"model":"reported-v1","content":[{"type":"text","text":""}],"usage":{"input_tokens":4,"cache_creation_input_tokens":2,"cache_read_input_tokens":3,"output_tokens":1}}`)
	}))
	defer server.Close()
	app := ensemble.New(nil)
	config := ensemble.Config{APIKey: "test", Model: "alias-v1", BaseURL: server.URL, LogPath: filepath.Join(t.TempDir(), "log")}
	a, err := app.NewAgent(config)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	if err = a.Ephemeral("one-use-marker"); err != nil {
		t.Fatal(err)
	}
	before := a.Events()
	_, _ = a.Render(config)
	_, _ = a.Render(config)
	if !reflect.DeepEqual(before, a.Events()) || len(a.Snapshot().Ephemera) != 1 {
		t.Fatal("render consumed ephemera")
	}
	if _, err = a.Ask(context.Background(), "failed-prompt"); err == nil {
		t.Fatal("HTTP failure accepted")
	}
	if answer, err := a.Ask(context.Background(), "accepted-prompt"); err != nil || answer != "" {
		t.Fatal("present empty text rejected", err)
	}
	if !strings.Contains(requests[0], "one-use-marker") || strings.Contains(requests[1], "one-use-marker") || strings.Contains(requests[1], "failed-prompt") {
		t.Fatal("failed attempt contaminated next context")
	}
	if a.Usage() != (ensemble.Usage{Input: 4, CacheWrite: 2, CacheRead: 3, Output: 1}) {
		t.Fatal(a.Usage())
	}
	events := a.Events()
	response := events[len(events)-1].Response
	if response.From.Model != "reported-v1" || response.Requested.Model != "alias-v1" || response.ModelReported == nil || !*response.ModelReported {
		t.Fatal("model identities conflated")
	}
	config.Model = "alias-v2"
	if err = a.SetConfig(config); err != nil {
		t.Fatal(err)
	}
	if _, ok := a.UsageByModel()[ensemble.Provenance{Vendor: "anthropic", Model: "reported-v1", Surface: "messages"}]; !ok {
		t.Fatal("usage reassigned")
	}
	dump, _ := a.Dump()
	replayed, err := loadText(t, string(dump))
	if err != nil || !reflect.DeepEqual(a.Snapshot(), replayed.Snapshot()) || !reflect.DeepEqual(a.UsageByModel(), replayed.UsageByModel()) {
		t.Fatal("failed replay", err)
	}
}
func TestToolResultsOpaqueAndOwnership(t *testing.T) {

	config := ensemble.Config{Vendor: "gemini", APIKey: "test", Model: "alias", LogPath: filepath.Join(t.TempDir(), "log")}
	app := ensemble.New(nil)
	a, err := app.NewAgent(config)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	if err = a.Append(ensemble.Event{Type: "message_received", Message: &ensemble.Entry{Actor: "human", Purpose: "dialogue", Parts: []ensemble.Part{ensemble.Text("inspect")}}}); err != nil {
		t.Fatal(err)
	}
	from := ensemble.Provenance{Vendor: "gemini", Model: "resolved-model", Surface: "generate_content"}
	id := "call-3-0"
	if err = a.Append(ensemble.Event{Type: "response_ended", Response: &ensemble.Response{From: from, Parts: []ensemble.Part{{Type: "tool_call", CallID: id, From: &from, Name: "inspect", Args: json.RawMessage(`{"path":"x"}`), Opaque: json.RawMessage(`"signed"`)}}, Usage: &ensemble.Usage{Input: 120, CacheRead: 30, Output: 40}}}); err != nil {
		t.Fatal(err)
	}
	if err = a.Append(ensemble.Event{Type: "tool_called", Tool: &ensemble.ToolEvent{CallID: id, Name: "inspect", Args: json.RawMessage(`{"path":"x"}`)}}); err != nil {
		t.Fatal(err)
	}
	if _, err = a.Ask(context.Background(), "unanswered"); err == nil {
		t.Fatal("unanswered tool ignored")
	}
	result := ensemble.Event{Type: "tool_returned", Tool: &ensemble.ToolEvent{CallID: id, Parts: []ensemble.Part{ensemble.Text("result-secret"), {Type: "blob", MIME: "text/plain", Ref: &ensemble.Ref{Kind: 2, Locator: "https://example.invalid/result"}}}}}
	if err = a.Append(result); err != nil {
		t.Fatal(err)
	}
	*result.Tool.Parts[0].Text = "caller mutation"
	if err = a.Append(result); err == nil {
		t.Fatal("duplicate result accepted")
	}
	if _, err = a.Render(config); err == nil {
		t.Fatal("signed call replayed to unproven alias")
	}
	config.ResolvedModel = "resolved-model"
	plain, err := a.Render(config)
	if err != nil || !bytes.Contains(plain, []byte("result-secret")) || !bytes.Contains(plain, []byte("signed")) {
		t.Fatal(string(plain), err)
	}
	if err = a.Redact(ensemble.Redaction{From: 4, To: 4, Level: "redact_result", Reason: "test"}); err != nil {
		t.Fatal(err)
	}
	redacted, err := a.Render(config)
	if err != nil || bytes.Contains(redacted, []byte("result-secret")) || !bytes.Contains(redacted, []byte("https://example.invalid/result")) {
		t.Fatal(string(redacted), err)
	}
	events := a.Events()
	events[0].Message.Parts[0].Text = new(string)
	if *a.Events()[0].Message.Parts[0].Text != "inspect" {
		t.Fatal("snapshot aliases authoritative events")
	}
	if a.Usage() != (ensemble.Usage{Input: 120, CacheRead: 30, Output: 40}) {
		t.Fatal(a.Usage())
	}
	dump, _ := a.Dump()
	if !bytes.Contains(dump, []byte("result-secret")) {
		t.Fatal("redaction erased log")
	}
}
func TestOpenAIParsingAndInstructionRoles(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls > 1 {
			fmt.Fprint(w, `{"choices":[{"message":{"content":""}}],"usage":{"prompt_tokens":0,"completion_tokens":0}}`)
			return
		}
		if r.Header.Get("Authorization") != "Bearer test" {
			t.Error("missing auth")
		}
		fmt.Fprint(w, `{"model":"reported","choices":[{"message":{"content":null,"tool_calls":[{"id":"issued","type":"function","function":{"name":"inspect","arguments":"{\"path\":\"x\"}"}}]}}],"usage":{"prompt_tokens":150,"prompt_tokens_details":{"cached_tokens":30,"cache_write_tokens":20},"completion_tokens":40}}`)
	}))
	defer server.Close()
	a, err := ensemble.New(nil).NewAgent(ensemble.Config{Vendor: "openai", APIKey: "test", Model: "alias", BaseURL: server.URL, LogPath: filepath.Join(t.TempDir(), "log")})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	if err = a.Append(ensemble.Event{Type: "message_received", Message: &ensemble.Entry{Actor: "system", Purpose: "instruction", Parts: []ensemble.Part{ensemble.Text("enduring")}}}); err != nil {
		t.Fatal(err)
	}
	result, err := a.Prompt(context.Background(), "call inspect")
	if err != nil || result.Text != "" {
		t.Fatal(err)
	}
	if a.Usage() != (ensemble.Usage{Input: 100, CacheRead: 30, CacheWrite: 20, Output: 40}) {
		t.Fatal(a.Usage())
	}
	var args map[string]string
	if json.Unmarshal(a.Events()[3].Response.Parts[0].Args, &args) != nil || args["path"] != "x" {
		t.Fatal("arguments not decoded into neutral object")
	}
}

func TestLoadOwnsConfiguration(t *testing.T) {
	schema := json.RawMessage(`{"type":"object"}`)
	config := ensemble.Config{Vendor: "anthropic", Model: "offline", Tools: []ensemble.ToolDefinition{{Name: "inspect", Schema: schema}}}
	a, err := ensemble.New(nil).Load("testdata/history.log", config)
	if err != nil {
		t.Fatal(err)
	}
	config.Tools[0].Name = "changed"
	schema[2] = 'X'
	got := a.Config()
	if got.Tools[0].Name != "inspect" || string(got.Tools[0].Schema) != `{"type":"object"}` {
		t.Fatal("loaded configuration aliases caller data")
	}
}

func TestDeferredHumanPreservesFactsAndProjectsAfterResults(t *testing.T) {
	source, err := os.ReadFile("testdata/history.log")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(source)), "\n")
	// Put the human event before the result while retaining strictly increasing sequences.
	deferred := strings.Replace(lines[5], `"seq":5`, `"seq":4`, 1)
	result := strings.Replace(lines[4], `"seq":4`, `"seq":5`, 1)
	prefix := strings.Join(append(lines[:4], deferred), "\n") + "\n"
	pending, err := loadText(t, prefix)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = pending.Render(ensemble.Config{Model: "offline"}); err == nil {
		t.Fatal("render accepted unanswered call")
	}
	second := strings.Replace(deferred, `"seq":4`, `"seq":5`, 1)
	if _, err = loadText(t, prefix+second+"\n"); err == nil {
		t.Fatal("second pending human accepted")
	}
	premature := strings.Replace(lines[2], `"seq":2`, `"seq":5`, 1)
	if _, err = loadText(t, prefix+premature+"\n"); err == nil {
		t.Fatal("response accepted before result")
	}
	a, err := loadText(t, prefix+result+"\n")
	if err != nil {
		t.Fatal(err)
	}
	if a.Snapshot().Pending.Seq != 4 || a.Events()[4].Seq != 5 {
		t.Fatal("source sequences changed")
	}
	for _, vendor := range []string{"anthropic", "openai", "gemini"} {
		body, err := a.Render(ensemble.Config{Vendor: vendor, Model: "offline"})
		if err != nil {
			t.Fatal(err)
		}
		text := string(body)
		if strings.Index(text, "port=8080") < 0 || strings.Index(text, "port=8080") > strings.Index(text, "now check the logs") {
			t.Fatalf("%s result is not before deferred human: %s", vendor, text)
		}
	}
}

func TestGeminiPreservesToolJSONSchema(t *testing.T) {
	a, err := ensemble.New(nil).Load("testdata/history.log", ensemble.Config{})
	if err != nil {
		t.Fatal(err)
	}
	schema := json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"}},"required":["path"],"additionalProperties":false}`)
	body, err := a.Render(ensemble.Config{Vendor: "gemini", Model: "offline", Tools: []ensemble.ToolDefinition{{Name: "inspect", Schema: schema}}})
	if err != nil {
		t.Fatal(err)
	}
	var wire struct {
		Tools []struct {
			Declarations []map[string]json.RawMessage `json:"functionDeclarations"`
		} `json:"tools"`
	}
	if err = json.Unmarshal(body, &wire); err != nil {
		t.Fatal(err)
	}
	declaration := wire.Tools[0].Declarations[0]
	if declaration["parameters"] != nil {
		t.Fatal("JSON Schema sent through OpenAPI subset")
	}
	var got, want any
	if err = json.Unmarshal(declaration["parametersJsonSchema"], &got); err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(schema, &want); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatal("tool schema lost constraints")
	}
}

func TestLoadDiagnosticsGiveSafeReasons(t *testing.T) {
	const private = "PRIVATE-LOG-CONTENT-781"
	cases := []struct{ name, record, reason string }{
		{"reference", `{"seq":1,"type":"message_received","time":"2026-01-01T00:00:00Z","message":{"actor":"human","parts":[{"type":"blob","mime":"text/plain","ref":{"kind":0,"locator":"` + private + `"}}]}}`, "invalid reference kind"},
		{"unsolicited", `{"seq":1,"type":"response_ended","time":"2026-01-01T00:00:00Z","response":{"from":{"vendor":"anthropic","model":"` + private + `","surface":"messages"},"parts":[{"type":"text","text":"` + private + `"}],"usage":{"input":0,"cache_write":0,"cache_read":0,"output":0}}}`, "unsolicited response"},
		{"sequence", `{"seq":0,"type":"message_received","time":"2026-01-01T00:00:00Z","message":{"actor":"human","parts":[{"type":"text","text":"` + private + `"}]}}`, "sequence must be positive"},
		{"payload", `{"seq":1,"type":"message_received","time":"2026-01-01T00:00:00Z","message":{"actor":"human","parts":[{"type":"text","text":"` + private + `"}]},"error":{"code":"x","message":"` + private + `"}}`, "exactly one event payload"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var diagnostics bytes.Buffer
			path := filepath.Join(t.TempDir(), "input.log")
			if err := os.WriteFile(path, []byte("{\"log_version\":1}\n"+tc.record+"\n"), 0600); err != nil {
				t.Fatal(err)
			}
			_, err := ensemble.New(&diagnostics).Load(path, ensemble.Config{APIKey: private})
			if err == nil || !strings.Contains(err.Error(), "line 2") || !strings.Contains(err.Error(), tc.reason) {
				t.Fatalf("missing safe reason: %v", err)
			}
			if !strings.Contains(diagnostics.String(), tc.reason) {
				t.Fatal("diagnostic did not reach owner logger")
			}
			if strings.Contains(err.Error(), private) || strings.Contains(diagnostics.String(), private) {
				t.Fatal("private data leaked in diagnostics")
			}
		})
	}
}
