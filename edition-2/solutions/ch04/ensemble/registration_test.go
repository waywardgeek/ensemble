package ensemble_test

import (
	"encoding/json"
	"io"
	"testing"

	"ensemble/ensemble"
)

func TestRegistrationIsAgentOwned(t *testing.T) {
	app := ensemble.New(io.Discard)
	a, err := app.NewAgent(ensemble.Config{Model: "fake", BuiltinTools: true, DataDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	b, err := app.NewAgent(ensemble.Config{Model: "fake", BuiltinTools: true, DataDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	handler := func(ensemble.ToolContext, json.RawMessage) (string, error) { return "ok", nil }
	schema := json.RawMessage(`{"type":"object"}`)
	declaration := ensemble.ToolDeclaration{Name: "custom", Description: "external", Schema: schema}
	before := len(b.Declarations())
	if err := a.RegisterTool(declaration, handler); err != nil {
		t.Fatal(err)
	}
	if len(a.Declarations()) != before+1 || len(b.Declarations()) != before {
		t.Fatal("registration leaked or lost builtins")
	}
	// The application's buffer is not the registry's mutable storage.
	schema[0] = '!'
	if got := a.Declarations()[before]; got.Name != "custom" || string(got.Schema) != `{"type":"object"}` {
		t.Fatalf("registration did not own declaration: %+v", got)
	}
	declaration.Schema = json.RawMessage(`{}`)
	if err := a.RegisterTool(declaration, handler); err == nil {
		t.Fatal("duplicate registration accepted")
	}
	declaration.Name = "read_file"
	if err := a.RegisterTool(declaration, handler); err == nil {
		t.Fatal("builtin replaced")
	}
	for _, bad := range []ensemble.ToolDeclaration{{Schema: json.RawMessage(`{}`)}, {Name: "bad", Schema: json.RawMessage(`{`)}} {
		if err := a.RegisterTool(bad, handler); err == nil {
			t.Fatal("invalid declaration accepted")
		}
	}
	if err := a.RegisterTool(ensemble.ToolDeclaration{Name: "nil", Schema: json.RawMessage(`{}`)}, nil); err == nil {
		t.Fatal("nil handler accepted")
	}
	if len(a.Declarations()) != before+1 {
		t.Fatal("rejected registration changed visibility")
	}
}
