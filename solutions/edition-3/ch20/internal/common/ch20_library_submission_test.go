package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestCh20LibrarySubmissionAdmission(t *testing.T) {
	for _, members := range []string{
		`[]`, `[1,1.0]`,
		`[{"n":9007199254740993}, {"n":9007199254740993.0}]`,
		`[[{"n":0.1}], [{"n":0.10}]]`,
	} {
		schema := json.RawMessage(`{"type":"object","properties":{"value":{"enum":` + members + `}}}`)
		if err := CheckSubmitSchema(schema); err == nil {
			t.Fatalf("admitted noncanonical enum members %s", members)
		}
	}
	schema := json.RawMessage(`{"type":"object","properties":{"n":{"enum":[9007199254740992,9007199254740993]},"v":{"const":{"n":0.1,"a":[1,true,null]}}},"required":["n","v"]}`)
	if err := ValidateSubmission(schema, json.RawMessage(`{"n":9007199254740993.0,"v":{"a":[1.0,true,null],"n":0.10}}`)); err != nil {
		t.Fatal(err)
	}
	if ValidateSubmission(schema, json.RawMessage(`{"n":9007199254740994,"v":{"n":0.1,"a":[1,true,null]}}`)) == nil {
		t.Fatal("rounded enum comparison")
	}
	// Keywords are checked only in schemas, not arbitrary annotation/const data.
	if err := ValidateSubmission(json.RawMessage(`{"type":"object","default":{"$ref":"file:///unused"},"properties":{"value":{"const":{"unsupported":true}}}}`), json.RawMessage(`{"value":{"unsupported":true}}`)); err != nil {
		t.Fatal(err)
	}
}

func TestCh20LibrarySubmissionCountBounds(t *testing.T) {
	for _, tc := range []struct {
		minimum, maximum, instance string
	}{
		{"minLength", "maxLength", `"x"`},
		{"minItems", "maxItems", `[1]`},
		{"minProperties", "maxProperties", `{"x":1}`},
	} {
		for _, number := range []string{"1048577", "18446744073709551615", "18446744073709551616", "1e1000"} {
			for _, keyword := range []string{tc.minimum, tc.maximum} {
				t.Run(keyword+"/"+number, func(t *testing.T) {
					schema := json.RawMessage(fmt.Sprintf(`{"type":"object","properties":{"value":{"%s":%s}}}`, keyword, number))
					original := append([]byte(nil), schema...)
					canonical, err := CanonicalSubmitSchema(schema)
					if err != nil {
						t.Fatal(err)
					}
					if !bytes.Equal(schema, original) || !bytes.Contains(canonical, []byte(number)) {
						t.Fatalf("compiler normalization rewrote original schema: %s / %s", schema, canonical)
					}
					err = ValidateSubmission(schema, json.RawMessage(`{"value":`+tc.instance+`}`))
					if (err != nil) != (keyword == tc.minimum) {
						t.Fatalf("bound overflow or changed semantics: %s, %v", schema, err)
					}
				})
			}
		}
		for _, number := range []string{"-1", "1.5", `"2"`, "true"} {
			schema := json.RawMessage(fmt.Sprintf(`{"type":"object","properties":{"value":{"%s":%s}}}`, tc.minimum, number))
			if CheckSubmitSchema(schema) == nil {
				t.Fatalf("admitted malformed bound %s", schema)
			}
		}
	}
}

func TestCh20LibrarySubmissionIsolation(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		http.Error(w, "must not fetch", http.StatusInternalServerError)
	}))
	defer server.Close()
	for _, schema := range []string{
		fmt.Sprintf(`{"type":"object","$schema":%q}`, server.URL),
		fmt.Sprintf(`{"type":"object","properties":{"x":{"$ref":%q}}}`, server.URL),
		`{"type":"object","additionalProperties":{"$ref":"file:///must-not-read"}}`,
		`{"type":"object","items":{"allOf":[]}}`,
	} {
		if CheckSubmitSchema(json.RawMessage(schema)) == nil {
			t.Fatalf("admitted external/unsupported schema %s", schema)
		}
	}
	if requests.Load() != 0 {
		t.Fatal("schema admission fetched a resource")
	}
	schema := json.RawMessage(`{"type":"object","properties":{"s":{"type":"string","pattern":"(?s)^α.*β$"}},"required":["s"]}`)
	payload := json.RawMessage(`{"s":"α\nβ","extra":{"n":9007199254740993}}`)
	originalSchema, originalPayload := append([]byte(nil), schema...), append([]byte(nil), payload...)
	canonical, err := CanonicalSubmitSchema(schema)
	if err != nil {
		t.Fatal(err)
	}
	custom, err := PrepareSubmission(schema, payload)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(schema, originalSchema) || !bytes.Equal(payload, originalPayload) || !bytes.Contains(custom, []byte(`9007199254740993`)) {
		t.Fatal("input mutation or custom numeric payload loss")
	}
	canonical[0], custom[0] = '!', '!'
	if !bytes.Equal(schema, originalSchema) || !bytes.Equal(payload, originalPayload) {
		t.Fatal("returned bytes alias caller storage")
	}
	if CheckSubmitSchema(json.RawMessage(`{"type":"object","properties":{"s":{"pattern":"(?=x)"}}}`)) == nil {
		t.Fatal("accepted non-Go regexp syntax")
	}
	if err := ValidateSubmission(json.RawMessage(`{"type":"object","properties":{"answer":{"type":"integer"}}}`), json.RawMessage(`{"answer":"wrong"}`)); err == nil || !strings.Contains(err.Error(), "$.answer: violates type") {
		t.Fatalf("lost the existing path/keyword feedback: %v", err)
	}
}
