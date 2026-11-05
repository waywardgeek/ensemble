package grade

import (
	"encoding/json"
	"testing"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
)

// Exercise the scoring boundary with otherwise complete declarations. This is
// schema-form compatibility, not a full recursive schema validator or live API test.
func TestCh3ToolDeclarationSchemaForms(t *testing.T) {
	cases := []struct {
		name, vendor, field, kind string
		conflict, want            bool
	}{
		{"gemini-schema-enum", "gemini", "parameters", "OBJECT", false, true},
		{"gemini-schema-legacy", "gemini", "parameters", "object", false, true},
		{"gemini-json-schema", "gemini", "parametersJsonSchema", "object", false, true},
		{"gemini-json-schema-snake", "gemini", "parameters_json_schema", "object", false, true},
		{"gemini-json-schema-not-enum", "gemini", "parametersJsonSchema", "OBJECT", false, false},
		{"gemini-json-schema-snake-not-enum", "gemini", "parameters_json_schema", "OBJECT", false, false},
		{"gemini-schema-not-arbitrary-case", "gemini", "parameters", "Object", false, false},
		{"gemini-schema-not-array", "gemini", "parameters", "ARRAY", false, false},
		{"gemini-conflicting-fields", "gemini", "parameters", "OBJECT", true, false},
		{"anthropic-json-schema", "anthropic", "input_schema", "object", false, true},
		{"anthropic-not-enum", "anthropic", "input_schema", "OBJECT", false, false},
		{"openai-json-schema", "openai", "parameters", "object", false, true},
		{"openai-not-enum", "openai", "parameters", "OBJECT", false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := &Ch3Result{Loop: map[string]*Ch3Session{}}
			for _, vendor := range Ch2Vendors {
				field, kind, conflict := "", "object", false
				if vendor == tc.vendor {
					field, kind, conflict = tc.field, tc.kind, tc.conflict
				}
				body := schemaFormRequest(t, vendor, field, kind, conflict)
				result.Loop[vendor] = &Ch3Session{Requests: []fakevendor.Recorded{{Body: body}}}
			}
			got := checkToolsDecl(result)
			if got.Passed != tc.want {
				t.Fatalf("passed=%v want=%v; %+v", got.Passed, tc.want, got)
			}
			wantPoints := 0
			if tc.want {
				wantPoints = 5
			}
			if got.Earned != wantPoints {
				t.Fatalf("earned=%v want=%v", got.Earned, wantPoints)
			}
		})
	}
}

func schemaFormRequest(t *testing.T, vendor, field, kind string, conflict bool) []byte {
	t.Helper()
	if field == "" {
		field = "parameters"
		if vendor == "anthropic" {
			field = "input_schema"
		}
	}
	propertyKind := "string"
	if vendor == "gemini" && field == "parameters" && kind == "OBJECT" {
		propertyKind = "STRING"
	}
	var declarations []any
	for _, name := range requiredTools {
		schema := map[string]any{
			"type":       kind,
			"properties": map[string]any{"path": map[string]any{"type": propertyKind}},
		}
		declaration := map[string]any{"name": name, "description": "test declaration", field: schema}
		if conflict {
			declaration["parametersJsonSchema"] = schema
		}
		if vendor == "openai" {
			declarations = append(declarations, map[string]any{"type": "function", "function": declaration})
		} else {
			declarations = append(declarations, declaration)
		}
	}
	tools := declarations
	if vendor == "gemini" {
		tools = []any{map[string]any{"functionDeclarations": declarations}}
	}
	body, err := json.Marshal(map[string]any{"tools": tools})
	if err != nil {
		t.Fatal(err)
	}
	return body
}
