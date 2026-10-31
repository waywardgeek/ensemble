package ensemble_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"ensemble/ensemble"
)

const memorySubmitSchema = `{"type":"object","properties":{"memory":{"type":"string"}},"required":["memory"],"additionalProperties":false}`

func submissionContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func sameJSON(t *testing.T, got, want []byte) {
	t.Helper()
	// UseNumber keeps large integers honest in both actual and expected data.
	decode := func(raw []byte) any {
		t.Helper()
		d := json.NewDecoder(bytes.NewReader(raw))
		d.UseNumber()
		var value any
		if err := d.Decode(&value); err != nil {
			t.Fatalf("invalid JSON value %q: %v", raw, err)
		}
		return value
	}
	if !reflect.DeepEqual(decode(got), decode(want)) {
		t.Fatalf("JSON mismatch\ngot %s\nwant %s", got, want)
	}
}

// Invalid model and Go calls use the same Fleet contract. The second provider
// request is held after actual rejection, proving fan-in has not completed.
func TestStructuredSubmissionCorrectionAndCopies(t *testing.T) {
	rejected, release := make(chan struct{}), make(chan struct{})
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if calls.Add(1) == 1 {
			fleetCall(w, "submit", map[string]any{"memory": 7})
			return
		}
		if !strings.Contains(fleetResultText(body), "submit arguments") {
			t.Error("real tool result omitted schema rejection")
		}
		close(rejected)
		select {
		case <-release:
			fleetCall(w, "submit", map[string]string{"memory": "I preserved reasons"})
		case <-r.Context().Done():
		}
	}))
	t.Cleanup(server.Close)
	parent, _, children := fleetAgent(t, server)
	ctx := submissionContext(t)
	schema := json.RawMessage(memorySubmitSchema)
	if _, err := parent.Fleet().SpawnWithSchema("worker", "TASK", "", schema); err != nil {
		t.Fatal(err)
	}
	for i := range schema {
		schema[i] = ' '
	}
	child := retainedChild(t, children, "worker")
	select {
	case <-rejected:
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	if child.Fleet().Status().State == "submitted" {
		t.Fatal("invalid model submit completed")
	}
	if err := child.Fleet().Submit("wrong contract"); err == nil {
		t.Fatal("string API bypassed custom schema")
	}
	joinCtx, cancel := context.WithTimeout(ctx, 30*time.Millisecond)
	defer cancel()
	if _, err := parent.Fleet().Join(joinCtx, []string{"worker"}); err != context.DeadlineExceeded {
		t.Fatalf("invalid submission satisfied join: %v", err)
	}
	if _, err := os.Stat(filepath.Join(child.Config().DataDir, "submit.md")); !os.IsNotExist(err) {
		t.Fatal("invalid submission wrote durable output", err)
	}
	// Both public spec and declaration reads must be safe to reuse or edit.
	spec := child.Spec()
	for i := range spec.SubmitSchema {
		spec.SubmitSchema[i] = ' '
	}
	for _, d := range child.Declarations() {
		if d.Name == "submit" {
			sameJSON(t, d.Schema, []byte(memorySubmitSchema))
			for i := range d.Schema {
				d.Schema[i] = ' '
			}
		}
	}
	sameJSON(t, child.Fleet().SubmitSchema(), []byte(memorySubmitSchema))
	sameJSON(t, child.Spec().SubmitSchema, []byte(memorySubmitSchema))
	close(release)
	joined, err := parent.Fleet().Join(ctx, []string{"worker"})
	if err != nil {
		t.Fatal(err)
	}
	if err := child.Flush(ctx); err != nil {
		t.Fatal(err)
	}
	var statuses []ensemble.ChildStatus
	if err := json.Unmarshal([]byte(joined), &statuses); err != nil || len(statuses) != 1 {
		t.Fatal(joined, err)
	}
	want := []byte(`{"memory":"I preserved reasons"}`)
	sameJSON(t, statuses[0].Payload, want)
	if statuses[0].Result != "" {
		t.Fatal("custom payload became text", statuses[0])
	}
	for i := range statuses[0].Payload {
		statuses[0].Payload[i] = ' '
	}
	status := child.Fleet().Status()
	for i := range status.Payload {
		status.Payload[i] = ' '
	}
	data, err := parent.Fleet().SubmissionData("worker")
	if err != nil {
		t.Fatal(err)
	}
	sameJSON(t, data.Payload, want)
	for i := range data.Payload {
		data.Payload[i] = ' '
	}
	if _, err := parent.Fleet().Submission("worker"); err == nil {
		t.Fatal("custom retrieval silently became text")
	}
	child.Fleet().Begin()
	data, err = parent.Fleet().SubmissionData("worker")
	if err != nil {
		t.Fatal(err)
	}
	sameJSON(t, data.Payload, want)
	if child.Fleet().Status().State == "submitted" {
		t.Fatal("historical fallback completed new task")
	}
	// Direct raw API callers also give away no mutable storage.
	raw := json.RawMessage(`{"memory":"I own the bytes"}`)
	if err := child.Fleet().SubmitJSON(raw); err != nil {
		t.Fatal(err)
	}
	for i := range raw {
		raw[i] = ' '
	}
	sameJSON(t, child.Fleet().Status().Payload, []byte(`{"memory":"I own the bytes"}`))
}

// Construction rejects invalid contracts before invoking a custom factory, and
// factories cannot replace the promised schema. Root construction rejects before
// creating its data/sandbox paths as well.
func TestStructuredSubmissionConstruction(t *testing.T) {
	app := ensemble.New(io.Discard)
	var factories atomic.Int32
	app.SetAgentFactory(func(parent ensemble.Agent, spec ensemble.AgentSpec) (ensemble.Agent, error) {
		factories.Add(1)
		spec.SubmitSchema = nil
		return app.NewChildAgent(parent, parent.Config(), spec)
	})
	parent, err := app.NewAgent(ensemble.Config{Model: "fake", DataDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { parent.Shutdown() })
	if _, err := parent.Fleet().SpawnWithSchema("bad", "TASK", "", json.RawMessage(`{"type":"object","$ref":"file:///tmp/private"}`)); err == nil {
		t.Fatal("invalid schema reached factory")
	}
	foreign, err := ensemble.New(io.Discard).NewAgent(ensemble.Config{Model: "fake", DataDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	defer foreign.Shutdown()
	for _, invalidParent := range []ensemble.Agent{nil, foreign} {
		if _, err := app.CreateChild(invalidParent, parent.Spec()); err == nil || !strings.Contains(err.Error(), "requires parent") {
			t.Fatal("invalid parent reached child factory", err)
		}
	}
	invalidSpec := parent.Spec()
	invalidSpec.SubmitSchema = json.RawMessage(`{"type":"object","$ref":"file:///tmp/private"}`)
	if _, err := app.CreateChild(parent, invalidSpec); err == nil {
		t.Fatal("direct CreateChild bypassed schema admission")
	}
	if factories.Load() != 0 || len(parent.Fleet().List()) != 0 {
		t.Fatal("partial child after invalid schema")
	}
	if _, err := parent.Fleet().SpawnWithSchema("changed", "TASK", "", json.RawMessage(memorySubmitSchema)); err == nil || !strings.Contains(err.Error(), "changed submission contract") {
		t.Fatal("factory discarded contract", err)
	}
	path := filepath.Join(t.TempDir(), "uncreated")
	_, err = app.NewAgent(ensemble.Config{Model: "fake"}, ensemble.AgentSpec{SandboxRoot: path, SubmitSchema: json.RawMessage(`{"type":"object","format":"email"}`)})
	if err == nil {
		t.Fatal("invalid constructor schema")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("invalid schema created sandbox", err)
	}
}

// Restart uses real history and retained submit rows. The new task gets the
// default schema; old structured output remains retrievable but cannot join it.
func TestStructuredSubmissionRestartAndGrandchildDefault(t *testing.T) {
	entered := make(chan struct{}, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if strings.Contains(string(body), "RESTORED_TASK") {
			var request struct {
				Tools []struct {
					Name   string
					Schema json.RawMessage `json:"input_schema"`
				}
			}
			if err := json.Unmarshal(body, &request); err != nil {
				t.Error(err)
			}
			for _, tool := range request.Tools {
				if tool.Name == "submit" {
					sameJSON(t, tool.Schema, []byte(`{"type":"object","properties":{"result":{"type":"string"}},"required":["result"]}`))
				}
			}
			entered <- struct{}{}
			<-r.Context().Done()
			return
		}
		fleetCall(w, "submit", map[string]string{"result": "DEFAULT"})
	}))
	t.Cleanup(server.Close)
	parent, _, children := fleetAgent(t, server)
	app := parent.Ensemble().(*ensemble.Ensemble)
	path := filepath.Join(parent.Config().DataDir, "children", "custom")
	spec := ensemble.AgentSpec{ID: "custom", DataDir: path, SavePath: filepath.Join(path, "save.json"), SubmitSchema: json.RawMessage(memorySubmitSchema)}
	child, err := app.NewChildAgent(parent, parent.Config(), spec)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = child.Fleet().Spawn("leaf", "LEAF_TASK", ""); err != nil {
		t.Fatal(err)
	}
	ctx := submissionContext(t)
	if _, err = child.Fleet().Join(ctx, []string{"leaf"}); err != nil {
		t.Fatal(err)
	}
	leaf := retainedChild(t, children, "leaf")
	if len(leaf.Spec().SubmitSchema) != 0 || leaf.Fleet().Status().Result != "DEFAULT" {
		t.Fatal("grandchild inherited schema")
	}
	if len(spec.Child("another").SubmitSchema) != 0 {
		t.Fatal("Child helper inherited schema")
	}
	if err = child.Fleet().SubmitJSON(json.RawMessage(`{"memory":"I retained this"}`)); err != nil {
		t.Fatal(err)
	}
	if err = child.Save(); err != nil {
		t.Fatal(err)
	}
	if err = child.Shutdown(); err != nil {
		t.Fatal(err)
	}
	if _, err = parent.Fleet().Spawn("custom", "RESTORED_TASK", ""); err != nil {
		t.Fatal(err)
	}
	select {
	case <-entered:
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	restored := retainedChild(t, children, "custom")
	data, err := parent.Fleet().SubmissionData("custom")
	if err != nil {
		t.Fatal(err)
	}
	sameJSON(t, data.Payload, []byte(`{"memory":"I retained this"}`))
	if restored.Fleet().Status().State == "submitted" {
		t.Fatal("old custom submission completed restarted task")
	}
	// Default accepts and ignores extra keys, including a misleading payload key.
	if err = restored.Fleet().SubmitJSON(json.RawMessage(`{"result":"","payload":"ignored","extra":true}`)); err != nil {
		t.Fatal(err)
	}
	data, err = parent.Fleet().SubmissionData("custom")
	if err != nil || data.Result != "" || len(data.Payload) != 0 {
		t.Fatal("default compatibility", data, err)
	}
	if err = parent.Fleet().ShutdownChild("custom"); err != nil {
		t.Fatal(err)
	}
	rows, err := os.ReadFile(filepath.Join(path, "submit.md"))
	if err != nil || !strings.HasSuffix(string(rows), "{\"result\":\"\"}\n") {
		t.Fatal("legacy row changed", string(rows), err)
	}
}

// A malformed latest row is an error, not permission to return an older success.
func TestStructuredSubmissionMalformedDurableRecord(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fleetCall(w, "submit", map[string]string{"result": "old"})
	}))
	t.Cleanup(server.Close)
	parent, _, children := fleetAgent(t, server)
	if _, err := parent.Fleet().Spawn("worker", "TASK", ""); err != nil {
		t.Fatal(err)
	}
	ctx := submissionContext(t)
	if _, err := parent.Fleet().Join(ctx, nil); err != nil {
		t.Fatal(err)
	}
	child := retainedChild(t, children, "worker")
	if err := child.Flush(ctx); err != nil {
		t.Fatal(err)
	}
	child.Fleet().Begin()
	for _, row := range []string{`{"version":2,"submit_schema":{},"payload":{}}`, `{"version":1,"payload":{}}`, `{"result":null}`, `{"version":1,"submit_schema":` + memorySubmitSchema + `,"payload":{"memory":5}}`, `{"payload":`} {
		if err := os.WriteFile(filepath.Join(child.Config().DataDir, "submit.md"), []byte("{\"result\":\"old\"}\n"+row+"\n"), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := parent.Fleet().SubmissionData("worker"); err == nil {
			t.Fatal("malformed row fell back to old success", row)
		}
	}
	for _, id := range []string{"foreign", "absent"} {
		if _, err := parent.Fleet().SubmissionData(id); err == nil || !strings.Contains(err.Error(), "unknown agent") {
			t.Fatal("retrieval escaped ownership", err)
		}
	}
}

// The real launch tool must forward the optional contract. A custom object may
// itself have a result key; its shape must not make it legacy text on retrieval.
func TestStructuredSubmissionSpawnTool(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fleetCall(w, "submit", map[string]any{"result": "custom", "facts": []any{true, nil, 7}})
	}))
	t.Cleanup(server.Close)
	parent, _, children := fleetAgent(t, server)
	ctx := submissionContext(t)
	args := json.RawMessage(`{"agent_id":"worker","task":"TASK","submit_schema":{"type":"object","properties":{"result":{"type":"string"},"facts":{"type":"array","items":{"type":["boolean","null","integer"]}}},"required":["result","facts"],"additionalProperties":false}}`)
	if _, err := parent.Invoke(ctx, "spawn_sub_agent", args); err != nil {
		t.Fatal(err)
	}
	if _, err := parent.Fleet().Join(ctx, nil); err != nil {
		t.Fatal(err)
	}
	child := retainedChild(t, children, "worker")
	if err := child.Flush(ctx); err != nil {
		t.Fatal(err)
	}
	child.Fleet().Begin()
	data, err := parent.Fleet().SubmissionData("worker")
	if err != nil {
		t.Fatal(err)
	}
	sameJSON(t, data.Payload, []byte(`{"result":"custom","facts":[true,null,7]}`))
	if data.Result != "" {
		t.Fatal("custom result key was interpreted as legacy", data)
	}
	// Real filesystem refusal must not publish a successful submission either.
	path := filepath.Join(child.Config().DataDir, "submit.md")
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(path, 0700); err != nil {
		t.Fatal(err)
	}
	if err := child.Fleet().SubmitJSON(data.Payload); err == nil {
		t.Fatal("accepted unwritten submission")
	}
	if child.Fleet().Status().State == "submitted" {
		t.Fatal("disk failure completed child")
	}
}

// Four real adapters must declare the exact memory contract and accept its
// direct object. Visible text precedes the call on the supported Responses row.
func TestStructuredSubmissionProviderDeclarations(t *testing.T) {
	for _, route := range []struct {
		name   string
		vendor ensemble.Vendor
		model  string
	}{
		{"anthropic", ensemble.Anthropic, "fake"}, {"chat", ensemble.OpenAI, "fake"},
		{"responses", ensemble.OpenAI, "gpt-6-astra"}, {"gemini", ensemble.Gemini, "fake"},
	} {
		t.Run(route.name, func(t *testing.T) {
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				var request map[string]any
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Error(err)
					return
				}
				var declared any
				for _, raw := range request["tools"].([]any) {
					tool := raw.(map[string]any)
					switch route.name {
					case "chat":
						tool = tool["function"].(map[string]any)
					case "gemini":
						for _, raw := range tool["functionDeclarations"].([]any) {
							d := raw.(map[string]any)
							if d["name"] == "submit" {
								declared = d["parameters"]
							}
						}
					}
					if tool["name"] == "submit" {
						declared = tool["parameters"]
						if route.name == "anthropic" {
							declared = tool["input_schema"]
						}
					}
				}
				actual, _ := json.Marshal(declared)
				sameJSON(t, actual, []byte(memorySubmitSchema))
				args := map[string]string{"memory": "I retained the reason"}
				switch route.name {
				case "anthropic":
					fleetCall(w, "submit", args)
				case "chat":
					json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": map[string]any{"role": "assistant", "content": "I will submit the memory.", "tool_calls": []any{map[string]any{"id": "call", "type": "function", "function": map[string]string{"name": "submit", "arguments": `{"memory":"I retained the reason"}`}}}}}}, "usage": map[string]int{"prompt_tokens": 1, "completion_tokens": 1}})
				case "responses":
					fmt.Fprint(w, `{"status":"completed","output":[{"type":"message","role":"assistant","content":[{"type":"output_text","text":"I will submit the memory."}]},{"type":"function_call","call_id":"call","name":"submit","arguments":"{\"memory\":\"I retained the reason\"}"}],"usage":{"input_tokens":1,"output_tokens":1}}`)
				case "gemini":
					json.NewEncoder(w).Encode(map[string]any{"candidates": []any{map[string]any{"content": map[string]any{"role": "model", "parts": []any{map[string]any{"functionCall": map[string]any{"name": "submit", "args": args}}}}}}, "usageMetadata": map[string]int{"promptTokenCount": 1, "candidatesTokenCount": 1}})
				}
			}))
			defer server.Close()
			app := ensemble.New(io.Discard)
			app.SetAgentFactory(func(parent ensemble.Agent, spec ensemble.AgentSpec) (ensemble.Agent, error) {
				return app.NewChildAgent(parent, parent.Config(), spec)
			})
			app.SetUserSink("user", make(humanInbox, 1))
			a, err := app.NewAgent(ensemble.Config{Vendor: route.vendor, Model: route.model, BaseURL: server.URL, DataDir: t.TempDir(), BuiltinTools: true}, ensemble.AgentSpec{SubmitSchema: json.RawMessage(memorySubmitSchema)})
			if err != nil {
				t.Fatal(err)
			}
			defer a.Shutdown()
			text, err := a.Ask(submissionContext(t), "TASK")
			if err != nil {
				t.Fatal(err)
			}
			sameJSON(t, []byte(text), []byte(`{"memory":"I retained the reason"}`))
			if calls.Load() != 1 {
				t.Fatal("submission did not finish", calls.Load())
			}
		})
	}
}

// Refusing a bad factory must not stop the parent or another live Agent. The
// Spawn path also holds the parent fleet lock, so shutting its parent would hang.
func TestStructuredSubmissionFactoryOwnership(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"content": []any{map[string]string{"type": "text", "text": "still alive"}}, "usage": map[string]int{"input_tokens": 1, "output_tokens": 1}})
	}))
	defer server.Close()
	app := ensemble.New(io.Discard)
	cfg := ensemble.Config{Model: "fake", BaseURL: server.URL, DataDir: t.TempDir()}
	parent, err := app.NewAgent(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer parent.Shutdown()
	cfg.DataDir = t.TempDir()
	other, err := ensemble.New(io.Discard).NewAgent(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer other.Shutdown()
	for _, returned := range []ensemble.Agent{parent, other} {
		app.SetAgentFactory(func(ensemble.Agent, ensemble.AgentSpec) (ensemble.Agent, error) { return returned, nil })
		if _, err := app.CreateChild(parent, parent.Spec()); err == nil || !strings.Contains(err.Error(), "ownership") {
			t.Fatal("accepted foreign factory result", err)
		}
		if answer, err := returned.Ask(submissionContext(t), "ALIVE"); err != nil || answer != "still alive" {
			t.Fatalf("refusal stopped existing Agent: answer=%q error=%v", answer, err)
		}
		if _, err := parent.Fleet().Spawn("refused", "TASK", ""); err == nil || !strings.Contains(err.Error(), "ownership") {
			t.Fatal("spawn accepted foreign factory result", err)
		}
		if len(parent.Fleet().List()) != 0 {
			t.Fatal("refused child was registered")
		}
	}
}

// Spawn/Join completion synchronizes each retained factory reference used here.
// No management lookup API is needed to observe the real constructed child.
func retainedChild(t *testing.T, children *[]ensemble.Agent, id string) ensemble.Agent {
	t.Helper()
	for i := len(*children) - 1; i >= 0; i-- {
		if a := (*children)[i]; a.Spec().ID == id {
			return a
		}
	}
	t.Fatalf("factory did not retain child %q", id)
	return nil
}

// A shared name does not make a browser or application tool a fleet submission.
// Exercise real MCP routing beside compiled precedence, including a factory
// installed after an existing Agent's tools were composed but before it starts.
func TestSubmissionToolIdentity(t *testing.T) {
	for _, tc := range []struct {
		name                          string
		builtins, fleet, late, custom bool
	}{
		{name: "MCP without builtins"},
		{name: "MCP with builtins", builtins: true},
		{name: "compiled without builtins", fleet: true},
		{name: "compiled with builtins", builtins: true, fleet: true},
		{name: "MCP before factory", late: true},
		{name: "custom before factory", late: true, custom: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			app := ensemble.New(io.Discard)
			factory := func(parent ensemble.Agent, spec ensemble.AgentSpec) (ensemble.Agent, error) {
				return app.NewChildAgent(parent, parent.Config(), spec)
			}
			if tc.fleet {
				app.SetAgentFactory(factory)
			}
			a, err := app.NewAgent(ensemble.Config{Model: "fake", DataDir: t.TempDir(), BuiltinTools: tc.builtins}, ensemble.AgentSpec{SubmitSchema: json.RawMessage(memorySubmitSchema)})
			if err != nil {
				t.Fatal(err)
			}
			defer a.Shutdown()
			schema := json.RawMessage(`{"type":"object","properties":{"question":{"type":"string"}},"required":["question"]}`)
			if tc.custom {
				err = a.RegisterTool(ensemble.ToolDeclaration{Name: "submit", Schema: schema}, func(_ ensemble.ToolContext, raw json.RawMessage) (string, error) {
					var args struct {
						// Question belongs to the custom tool, not the fleet contract.
						Question string
					}
					if err := json.Unmarshal(raw, &args); err != nil {
						return "", err
					}
					return "custom: " + args.Question, nil
				})
				if err != nil {
					t.Fatal(err)
				}
			}
			if tc.late {
				app.SetAgentFactory(factory)
			}
			local, remote := ensemble.MCPPipe()
			defer remote.Close()
			var peerCalls atomic.Int32
			go func() {
				for {
					data, err := remote.Recv()
					if err != nil {
						return
					}
					var request struct {
						// ID correlates this protocol reply with its caller.
						ID json.RawMessage
						// Method selects discovery or an actual remote invocation.
						Method string
						// Params preserves the call that reached the external peer.
						Params struct {
							// Name identifies the selected MCP tool.
							Name string
							// Arguments proves custom input survived dispatch and transport.
							Arguments struct {
								// Question is the peer's required input field.
								Question string
							}
						}
					}
					if json.Unmarshal(data, &request) != nil {
						return
					}
					var result any
					switch request.Method {
					case "initialize":
						result = map[string]any{"protocolVersion": "2024-11-05", "capabilities": map[string]any{"tools": map[string]any{}}}
					case "tools/list":
						result = map[string]any{"tools": []any{map[string]any{"name": "submit", "inputSchema": schema}}}
					case "tools/call":
						peerCalls.Add(1)
						if request.Params.Name != "submit" {
							t.Error("wrong MCP tool", request.Params.Name)
						}
						result = map[string]any{"content": []any{map[string]any{"type": "text", "text": "peer: " + request.Params.Arguments.Question}}}
					default:
						continue
					}
					response, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": request.ID, "result": result})
					if remote.Send(response) != nil {
						return
					}
				}
			}()
			ctx := submissionContext(t)
			if err := a.MCP().Connect(ctx, "browser", local); err != nil {
				t.Fatal(err)
			}
			found := 0
			for _, declaration := range a.Declarations() {
				if declaration.Name == "submit" {
					found++
					want := schema
					if tc.fleet {
						want = json.RawMessage(memorySubmitSchema)
					}
					sameJSON(t, declaration.Schema, want)
				}
				if !tc.fleet && declaration.Name == "spawn_sub_agent" {
					t.Fatal("disabled fleet tool registered")
				}
			}
			if found != 1 {
				t.Fatal("submit declaration count", found)
			}
			args, want, wantPeer := json.RawMessage(`{"question":"IDENTITY_MARKER"}`), "peer: IDENTITY_MARKER", int32(1)
			if tc.custom {
				want, wantPeer = "custom: IDENTITY_MARKER", 0
			}
			if tc.fleet {
				args, want, wantPeer = json.RawMessage(`{"memory":"IDENTITY_MARKER"}`), `submitted: {"memory":"IDENTITY_MARKER"}`, 0
			}
			got, err := a.Invoke(ctx, "submit", args)
			if err != nil || got != want || peerCalls.Load() != wantPeer {
				t.Fatal("submit dispatch identity", got, err, peerCalls.Load())
			}
			if !tc.fleet {
				if a.Fleet().Status().State == "submitted" {
					t.Fatal("external submit completed fleet")
				}
				if _, err := a.Invoke(ctx, "spawn_sub_agent", json.RawMessage(`{"agent_id":"x","task":"x"}`)); err == nil || !strings.Contains(err.Error(), "unknown tool") {
					t.Fatal("disabled fleet dispatch", err)
				}
			}
		})
	}
}
