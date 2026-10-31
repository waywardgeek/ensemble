package fleet

import (
	"encoding/json"
	"strings"
	"testing"
)

// The actual validator must enforce the advertised dialect, not merely decode
// into Go structs. Data under enum/default is not itself a schema vocabulary.
func TestSubmissionSchemaSemantics(t *testing.T) {
	contract, err := CompileSubmissionSchema(json.RawMessage(`{
	 "type":"object","properties":{
	  "memory":{"type":"string","minLength":2,"maxLength":20,"pattern":"^I "},
	  "items":{"type":"array","minItems":1,"maxItems":2,"uniqueItems":true,"items":{
	   "type":"object","properties":{"n":{"type":"integer","minimum":1,"maximum":4,"multipleOf":2},
	    "kind":{"enum":["fact","reason"]},"extra":{"type":["string","null"]}},
	   "required":["n","kind","extra"],"additionalProperties":false}},
	  "literal":{"const":{"$ref":"ordinary data"},"default":{"unknown":"data"},"examples":[{"else":"data"}]}},
	 "required":["memory","items","literal"],"additionalProperties":false,"minProperties":3,"maxProperties":3}`))
	if err != nil {
		t.Fatal(err)
	}
	valid := `{"memory":"I learned","items":[{"n":2.0,"kind":"reason","extra":null}],"literal":{"$ref":"ordinary data"}}`
	if _, err := contract.validate(json.RawMessage(valid)); err != nil {
		t.Fatal(err)
	}
	for name, raw := range map[string]string{
		"required":        strings.Replace(valid, `"memory":"I learned",`, "", 1),
		"null":            strings.Replace(valid, `"I learned"`, "null", 1),
		"wrong type":      strings.Replace(valid, `"I learned"`, "3", 1),
		"nested extra":    strings.Replace(valid, `"n":2.0`, `"unknown":true,"n":2.0`, 1),
		"extra":           strings.Replace(valid, `"memory":`, `"unknown":true,"memory":`, 1),
		"integer":         strings.Replace(valid, "2.0", "2.5", 1),
		"multiple":        strings.Replace(valid, "2.0", "3", 1),
		"minimum":         strings.Replace(valid, "2.0", "0", 1),
		"maximum":         strings.Replace(valid, "2.0", "6", 1),
		"enum":            strings.Replace(valid, `"reason"`, `"fiction"`, 1),
		"const":           strings.Replace(valid, "ordinary data", "other", 1),
		"pattern":         strings.Replace(valid, "I learned", "They learned", 1),
		"min length":      strings.Replace(valid, "I learned", "I", 1),
		"max length":      strings.Replace(valid, "I learned", "I "+strings.Repeat("x", 20), 1),
		"array minimum":   `{"memory":"I learned","items":[],"literal":{"$ref":"ordinary data"}}`,
		"array duplicate": `{"memory":"I learned","items":[{"n":2,"kind":"fact","extra":null},{"n":2,"kind":"fact","extra":null}],"literal":{"$ref":"ordinary data"}}`,
		"array maximum":   `{"memory":"I learned","items":[{}, {}, {}],"literal":{"$ref":"ordinary data"}}`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := contract.validate(json.RawMessage(raw)); err == nil {
				t.Fatalf("accepted invalid submission %s", raw)
			}
		})
	}
	// Boolean subschemas and exact numbers are normal JSON Schema semantics.
	contract, err = CompileSubmissionSchema(json.RawMessage(`{"type":"object","properties":{"ok":true,"no":false,"n":{"const":9007199254740993}}}`))
	if err != nil {
		t.Fatal(err)
	}
	for _, raw := range []string{`{"ok":null,"n":9007199254740993}`, `{"ok":[true]}`} {
		if _, err := contract.validate(json.RawMessage(raw)); err != nil {
			t.Fatal(err)
		}
	}
	for _, raw := range []string{`{"no":null}`, `{"n":9007199254740992}`} {
		if _, err := contract.validate(json.RawMessage(raw)); err == nil {
			t.Fatal("accepted", raw)
		}
	}
}

// Errors fail at schema admission, including constructs the dependency itself
// would otherwise accept or ignore. No reference gets a chance to access a file.
func TestSubmissionSchemaAdmission(t *testing.T) {
	for _, raw := range []string{
		`null`, `true`, `[]`, `{}`, `{"type":"array"}`, `{"type":"object","required":"memory"}`,
		`{"type":"object","properties":{"x":{"minLength":-1}}}`,
		`{"type":"object","properties":{"x":{"pattern":"["}}}`,
		`{"type":"object","$schema":"http://json-schema.org/draft-07/schema#"}`,
		`{"type":"object","$ref":"#/properties/x"}`,
		`{"type":"object","$ref":"https://example.invalid/schema"}`,
		`{"type":"object","properties":{"x":{"$ref":"file:///tmp/private"}}}`,
		`{"type":"object","items":{"format":"email"}}`,
		`{"type":"object","additionalProperties":{"oneOf":[true,false]}}`,
		`{"type":"object","mimimum":1}`,
		`{"type":"object","default":1e1000000000}`,
		`{"type":"object"} {}`,
		`{"type":"object","description":"` + strings.Repeat("x", 65536) + `"}`,
	} {
		if _, err := CompileSubmissionSchema(json.RawMessage(raw)); err == nil {
			t.Errorf("accepted invalid schema %.150s", raw)
		}
	}
	contract, err := CompileSubmissionSchema(json.RawMessage(`{"type":"object"}`))
	if err != nil {
		t.Fatal(err)
	}
	for _, raw := range []string{
		`{"n":1e1001}`, `{"n":1e1000000000}`, `{"n":` + strings.Repeat("1", 129) + `}`,
		`{"x":"` + strings.Repeat("x", 1024*1024) + `"}`,
		`{"x":` + strings.Repeat("[", 65) + `0` + strings.Repeat("]", 65) + `}`,
		`{"x":[` + strings.Repeat("0,", 65536) + `0]}`,
	} {
		if _, err := contract.validate(json.RawMessage(raw)); err == nil {
			t.Errorf("accepted oversized submission %.150s", raw)
		}
	}
	// The schema node limit applies even when bytes/nesting are small.
	if _, err := CompileSubmissionSchema(json.RawMessage(`{"type":"object","enum":[` + strings.Repeat("0,", 2048) + `0]}`)); err == nil {
		t.Fatal("accepted excessive schema nodes")
	}
}

// Canonical escaping must not produce an accepted value that its durable reader
// rejects. The combined budget also limits repeated schema/instance work.
func TestSubmissionCanonicalAndCombinedLimits(t *testing.T) {
	for _, text := range []string{strings.Repeat("<", 11000), strings.Repeat("\u2028", 11000)} {
		raw := json.RawMessage(`{"type":"object","description":"` + text + `"}`)
		if _, err := CompileSubmissionSchema(raw); err == nil || !strings.Contains(err.Error(), "canonical") {
			t.Fatalf("canonical schema expansion accepted: %v", err)
		}
	}
	contract, err := CompileSubmissionSchema(json.RawMessage(`{"type":"object","properties":{"memory":{"type":"string"}},"required":["memory"],"additionalProperties":false}`))
	if err != nil {
		t.Fatal(err)
	}
	simple, err := CompileSubmissionSchema(json.RawMessage(`{"type":"object"}`))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := simple.validate(json.RawMessage(`{"memory":"` + strings.Repeat("<", 180000) + `"}`)); err == nil || !strings.Contains(err.Error(), "canonical") {
		t.Fatalf("canonical payload expansion accepted: %v", err)
	}
	valid := json.RawMessage(`{"memory":"` + strings.Repeat("<", 10000) + `"}`)
	owned, err := contract.validate(valid)
	if err != nil {
		t.Fatal(err)
	}
	reloaded, err := CompileSubmissionSchema(contract.Schema())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := reloaded.validate(owned); err != nil {
		t.Fatal("accepted bytes failed reload", err)
	}
	// Each input is small enough alone; their product exceeds 64 MiB. Annotation
	// bytes count conservatively too, avoiding a second schema-work evaluator.
	large, err := CompileSubmissionSchema(json.RawMessage(`{"type":"object","description":"` + strings.Repeat("x", 8192) + `"}`))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := large.validate(json.RawMessage(`{"memory":"` + strings.Repeat("x", 8192) + `"}`)); err == nil || !strings.Contains(err.Error(), "combined") {
		t.Fatalf("excessive combined validation admitted: %v", err)
	}
}
