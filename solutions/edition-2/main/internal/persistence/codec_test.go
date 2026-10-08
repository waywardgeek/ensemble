package persistence

import (
	"bytes"
	"example.com/ensemble/internal/common"
	"strings"
	"testing"
)

type testRoot struct{ common.Ensemble }

func (*testRoot) Logf(string, ...any) {}

type testAgent struct{ common.Agent }

func (*testAgent) Ensemble() common.Ensemble { return &testRoot{} }
func TestCanonicalExactDecimals(t *testing.T) {
	c := NewCodec(&testAgent{})
	for input, want := range map[string]string{
		`{"z":-0.00e99,"a":1.000e+3}`:                                `{"a":1e3,"z":0}`,
		`[9007199254740992,9007199254740993,18446744073709551615.0]`: `[9007199254740992,9007199254740993,18446744073709551615]`,
		`[1e1000000,12.30,10e-1]`:                                    `[1e1000000,123e-1,1]`,
		`{"é":"<>&/\u0001\n"}`:                                       `{"é":"<>&/\u0001\n"}`,
	} {
		got, err := c.Canonical([]byte(input))
		if err != nil || string(got) != want {
			t.Fatalf("canonical %s: %s %v", input, got, err)
		}
	}
	for _, input := range []string{`{"a":1,"a":2}`, `{"a":{"b":0,"b":1}}`, `[NaN]`, `{} {}`, string([]byte{'"', 255, '"'}), strings.Repeat("[", 129) + "0" + strings.Repeat("]", 129)} {
		if _, err := c.Canonical([]byte(input)); err == nil {
			t.Fatalf("accepted invalid JSON")
		}
	}
}
func TestCheckpointStrictAndRaw(t *testing.T) {
	c := NewCodec(&testAgent{})
	base := "base"
	id := common.SessionIdentity{Mode: "plain", System: &base, Handlers: []common.HandlerIdentity{}}
	cp := common.Checkpoint{Version: 1, StateVersion: 1, SessionID: strings.Repeat("a", 32), Identity: id, AsOf: 1, HighWatermarks: common.Watermarks{Event: 1}}
	cp.State = &common.SemanticState{Session: common.SnapshotSession{ID: cp.SessionID, Identity: id, AsOf: 1, HighWatermarks: cp.HighWatermarks}}
	rawUsage := []byte(`{ "tokens": 9007199254740993.0 }`)
	cp.State.Context.Responses = []common.ResponseFact{{RawUsage: rawUsage}}
	raw, err := c.Encode(cp)
	if err != nil {
		t.Fatal(err)
	}
	got, err := c.Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got.State.Context.Responses[0].RawUsage, rawUsage) {
		t.Fatal("raw replay bytes changed")
	}
	for _, mutated := range [][]byte{bytes.Replace(raw, []byte(`"version":1`), []byte(`"version":1,"version":1`), 1), bytes.Replace(raw, []byte(`"state_version":1`), []byte(`"state_version":2`), 1), bytes.Replace(raw, []byte(`"as_of":1`), []byte(`"as_of":2`), 1)} {
		if _, err = c.Decode(mutated); err == nil {
			t.Fatal("accepted invalid envelope")
		}
	}
}

func TestUnicodeScalarEscapes(t *testing.T) {
	c := NewCodec(&testAgent{})
	for _, raw := range []string{`"\ud800"`, `{"\udfff":1}`, `"\ud800x"`, `"\ud800\ud800"`} {
		if _, err := c.Canonical([]byte(raw)); err == nil {
			t.Errorf("accepted lone surrogate: %s", raw)
		}
	}
	for _, raw := range []string{`"\ud83d\ude00"`, `"�"`, `"\ufffd"`, `"\\ud800"`} {
		if _, err := c.Canonical([]byte(raw)); err != nil {
			t.Errorf("rejected scalar/literal: %s: %v", raw, err)
		}
	}
	if err := c.ValidateLogJSON([]byte(`"\ud800"`), false); err != nil {
		t.Fatal("changed standalone scope", err)
	}
}
