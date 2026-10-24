package ensemble_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"ensemble"
)

func TestChapter3GeminiParameterProjection(t *testing.T) {
	others := []ensemble.Agent{
		c3Agent(t, ensemble.Config{Vendor: ensemble.Anthropic}),
		c3Agent(t, ensemble.Config{Vendor: ensemble.OpenAI}),
	}
	originalWire := make([][]byte, len(others))
	for i, a := range others {
		var err error
		originalWire[i], err = a.Render()
		if err != nil {
			t.Fatal(err)
		}
		var request struct {
			Tools []struct {
				Name     string          `json:"name"`
				Input    json.RawMessage `json:"input_schema"`
				Function struct {
					Name       string          `json:"name"`
					Parameters json.RawMessage `json:"parameters"`
				} `json:"function"`
			} `json:"tools"`
		}
		if err := json.Unmarshal(originalWire[i], &request); err != nil {
			t.Fatal(err)
		}
		defs := a.Tools().Definitions()
		if len(request.Tools) != len(defs) {
			t.Fatal("other provider lost declarations")
		}
		for j, d := range defs {
			name, schema := request.Tools[j].Name, request.Tools[j].Input
			if i == 1 {
				name, schema = request.Tools[j].Function.Name, request.Tools[j].Function.Parameters
			}
			if name != d.Name || !bytes.Equal(schema, d.Parameters) {
				t.Fatalf("other provider %d changed declaration %s", i, d.Name)
			}
		}
	}

	a := c3Agent(t, ensemble.Config{Vendor: ensemble.Gemini})
	before := a.Tools().Definitions()
	wire, err := a.Render()
	if err != nil {
		t.Fatal(err)
	}
	var request struct {
		Tools []struct {
			Functions []struct {
				Name        string                     `json:"name"`
				Description string                     `json:"description"`
				Parameters  map[string]json.RawMessage `json:"parameters"`
			} `json:"functionDeclarations"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(wire, &request); err != nil {
		t.Fatal(err)
	}
	if len(request.Tools) != 1 || len(request.Tools[0].Functions) != 21 || len(before) != 21 {
		t.Fatalf("wrong Gemini declaration shape: %s", wire)
	}
	for i, d := range before {
		got := request.Tools[0].Functions[i]
		if got.Name != d.Name || got.Description != d.Description {
			t.Fatalf("identity changed: %s", d.Name)
		}
		var want map[string]json.RawMessage
		if err := json.Unmarshal(d.Parameters, &want); err != nil {
			t.Fatal(err)
		}
		if d.Name == "submit" {
			// Default submit accepts a result string and ignores extra fields.
			// Assert that public schema independently, rather than imposing the
			// strict argument policy of the other twenty tools on this payload.
			var expected map[string]json.RawMessage
			if err := json.Unmarshal([]byte(`{"type":"object","properties":{"result":{"type":"string"}},"required":["result"]}`), &expected); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(want, expected) {
				t.Fatalf("default submit schema changed: %s", d.Parameters)
			}
		} else if string(want["additionalProperties"]) != "false" {
			t.Fatalf("common schema lost strictness: %s", d.Name)
		}
		delete(want, "additionalProperties")
		if _, exists := got.Parameters["additionalProperties"]; exists {
			t.Fatalf("Gemini still contains unsupported keyword: %s", d.Name)
		}
		if !reflect.DeepEqual(got.Parameters, want) {
			t.Fatalf("Gemini changed other schema fields: %s", d.Name)
		}
	}
	if !reflect.DeepEqual(before, a.Tools().Definitions()) {
		t.Fatal("Gemini projection mutated common declarations")
	}
	for i, other := range others {
		after, err := other.Render()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(after, originalWire[i]) {
			t.Fatalf("Gemini render affected other provider %d bytes", i)
		}
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "file.txt")
	marker := filepath.Join(dir, "should-not-exist")
	c3Write(t, path, "original")
	cases := []struct {
		name string
		args map[string]any
	}{
		{"read_file", map[string]any{"path": path, "max_bytes": true}},
		{"list_directory", map[string]any{"path": 42}},
		{"search_files", map[string]any{"path": dir, "pattern": "original", "context_lines": true}},
		{"write_file", map[string]any{"path": path, "text": "changed", "overwrite": "true"}},
		{"edit_file", map[string]any{"path": path, "old_text": "original", "new_text": 42}},
		{"run_command", map[string]any{"command": fmt.Sprintf("touch %q", marker), "cwd": 42}},
		{"read_file", map[string]any{"path": path, "unadvertised": true}},
	}
	for _, tc := range cases {
		if text, failed := c3Call(t, a, tc.name, tc.args); !failed {
			t.Fatalf("%s accepted invalid local arguments: %s", tc.name, text)
		}
	}
	if c3Contents(t, path) != "original" {
		t.Fatal("invalid call changed file")
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("invalid command ran: %v", err)
	}
}
