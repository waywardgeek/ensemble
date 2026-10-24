package common

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCh20SchemaContract(t *testing.T) {
	schema := json.RawMessage(`{"$schema":"https://json-schema.org/draft/2020-12/schema","type":"object","properties":{"name":{"type":"string","minLength":2,"maxLength":3,"pattern":"^é"},"n":{"type":["integer","null"],"minimum":1,"exclusiveMaximum":5,"multipleOf":1},"items":{"type":"array","minItems":1,"maxItems":3,"uniqueItems":true,"items":{"enum":[1,2,3]}},"fixed":{"const":{"n":1.0}},"never":false},"required":["name","n","items","fixed"],"additionalProperties":{"type":"boolean"},"minProperties":4,"maxProperties":5}`)
	good := `{"name":"éa","n":1.0,"items":[1,2],"fixed":{"n":1},"extra":true}`
	if err := ValidateSubmission(schema, json.RawMessage(good)); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{strings.Replace(good, `"éa"`, `"é"`, 1), strings.Replace(good, `1.0`, `5`, 1), strings.Replace(good, `[1,2]`, `[1,1.0]`, 1), strings.Replace(good, `"extra":true`, `"never":true`, 1), strings.Replace(good, `"extra":true`, `"extra":1`, 1)} {
		if ValidateSubmission(schema, json.RawMessage(bad)) == nil {
			t.Fatalf("accepted %s", bad)
		}
	}
	exact := json.RawMessage(`{"type":"object","properties":{"n":{"multipleOf":0.1,"minimum":9007199254740993,"maximum":9007199254740993}},"required":["n"]}`)
	if err := ValidateSubmission(exact, json.RawMessage(`{"n":9007199254740993.0}`)); err != nil {
		t.Fatal(err)
	}
	if ValidateSubmission(exact, json.RawMessage(`{"n":9007199254740992}`)) == nil {
		t.Fatal("rounded a JSON number")
	}
	for _, bad := range []string{`true`, `{"type":"object","$ref":"x"}`, `{"type":"object","enum":[1,1.0]}`, `{"type":"object","properties":{"x":{"type":[]}}}`, `{"type":"object","items":{"pattern":"["}}`, `{"type":"object","multipleOf":0}`, `{"type":"object","minItems":-1}`, `{"type":"object","$schema":"draft-07"}`, `{"type":"object","allOf":[]}`} {
		if CheckSubmitSchema(json.RawMessage(bad)) == nil {
			t.Fatalf("admitted %s", bad)
		}
	}
	normalized, err := PrepareSubmission(nil, json.RawMessage(`{"result":"done","ignored":{"x":2}}`))
	if err != nil || string(normalized) != `{"result":"done"}` {
		t.Fatalf("default = %s, %v", normalized, err)
	}
	if err := CheckSubmitSchema(json.RawMessage(`{"type":"object","properties":{"arbitrary":true,"none":false},"additionalProperties":true,"examples":[{}],"default":{}}`)); err != nil {
		t.Fatal(err)
	}
}

func TestCh20SubmissionBudgets(t *testing.T) {
	schema := `{"type":"object"}`
	if err := CheckSubmitSchema(json.RawMessage(strings.Repeat(" ", 65536-len(schema)) + schema)); err != nil {
		t.Fatal(err)
	}
	if CheckSubmitSchema(json.RawMessage(strings.Repeat(" ", 65537-len(schema))+schema)) == nil {
		t.Fatal("raw schema budget")
	}
	if CheckSubmitSchema(json.RawMessage(`{"type":"object","description":"`+strings.Repeat("<", 11000)+`"}`)) == nil {
		t.Fatal("canonical schema budget")
	}
	for _, n := range []string{"1e1001", "1e-1001", strings.Repeat("1", 257)} {
		if ValidateSubmission(nil, json.RawMessage(`{"result":"x","n":`+n+`}`)) == nil {
			t.Fatal("numeric budget", n)
		}
	}
	if err := ValidateSubmission(nil, json.RawMessage(`{"result":"x","n":1e1000}`)); err != nil {
		t.Fatal(err)
	}
	deep := `0`
	for i := 0; i < 65; i++ {
		deep = `[` + deep + `]`
	}
	if ValidateSubmission(nil, json.RawMessage(`{"result":"x","ignored":`+deep+`}`)) == nil {
		t.Fatal("depth budget")
	}
	huge := `{"result":"` + strings.Repeat("x", 1048576) + `"}`
	if ValidateSubmission(nil, json.RawMessage(huge)) == nil {
		t.Fatal("payload byte budget")
	}
	nodes := `{"result":"x","ignored":[` + strings.Repeat("0,", 65533) + `0]}`
	if ValidateSubmission(nil, json.RawMessage(nodes)) == nil {
		t.Fatal("node budget")
	}
	expensive := json.RawMessage(`{"type":"object","examples":[` + strings.Repeat("0,", 1499) + `0]}`)
	if err := CheckSubmitSchema(expensive); err != nil {
		t.Fatal(err)
	}
	payload := json.RawMessage(`{"items":[` + strings.Repeat("0,", 999) + `0]}`)
	if err := ValidateSubmission(expensive, payload); err != nil {
		t.Fatalf("valid byte-budget input must not use the withdrawn node product: %v", err)
	}
}

func TestCh20Round2SchemaBoundaries(t *testing.T) {
	for _, depth := range []int{32, 33, 64, 65} {
		nested := strings.Repeat("[", depth-1) + "0" + strings.Repeat("]", depth-1)
		schema := json.RawMessage(`{"type":"object","default":` + nested + `}`)
		payload := json.RawMessage(`{"result":"ok","extra":` + nested + `}`)
		for name, err := range map[string]error{"schema": CheckSubmitSchema(schema), "payload": ValidateSubmission(nil, payload)} {
			if (err == nil) != (depth <= 64) {
				t.Fatalf("%s depth %d: %v", name, depth, err)
			}
		}
	}
	for _, size := range []int{128, 129, 256} {
		number := strings.Repeat("1", size)
		for name, err := range map[string]error{"schema": CheckSubmitSchema(json.RawMessage(`{"type":"object","default":` + number + `}`)), "payload": ValidateSubmission(nil, json.RawMessage(`{"result":"ok","n":`+number+`}`))} {
			if (err == nil) != (size <= 128) {
				t.Fatalf("%s literal %d: %v", name, size, err)
			}
		}
	}
	padded := func(prefix, suffix string, size int) json.RawMessage {
		return json.RawMessage(prefix + strings.Repeat("x", size-len(prefix)-len(suffix)) + suffix)
	}
	schema := padded(`{"type":"object","description":"`, `"}`, 1024)
	for _, size := range []int{65536, 65537} {
		err := ValidateSubmission(schema, padded(`{"result":"`, `"}`, size))
		if (err == nil) != (size == 65536) {
			t.Fatalf("canonical product %d * 1024: %v", size, err)
		}
	}
	escaped := json.RawMessage(`{"type":"object","description":"` + strings.Repeat("<", 1000) + `"}`)
	if ValidateSubmission(escaped, padded(`{"result":"`, `"}`, 12000)) == nil {
		t.Fatal("product must use canonical escaped bytes")
	}
	// Even ignored default fields remain subject to admission resource limits.
	if ValidateSubmission(nil, padded(`{"result":"ok","extra":"`, `"}`, 1048576)) == nil {
		t.Fatal("default normalization bypassed byte product")
	}
}
