package ch10_test

// Contract-derived component boundaries use the actual Agent-owned codec.
// Only the request-cursor group claims end-to-end session admission coverage.
// These checks neither import implementation packages nor contact providers.
import (
	"bytes"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCh10StorageCanonical(t *testing.T) {
	a := opened(t, application(t), options(t, "canonical"))
	cases := []struct{ name, input, want string }{
		{"decimal-equivalence", `[1,1.0,1e0,10e-1,1000,1.000e+3,10e2,12.30,123e-1]`, `[1,1,1,1,1e3,1e3,1e3,123e-1,123e-1]`},
		{"uint64-and-adjacent", `[9007199254740992,9007199254740993,18446744073709551615.0]`, `[9007199254740992,9007199254740993,18446744073709551615]`},
		{"zero-and-unexpanded-exponents", `[-0,-0.00e99,0,1e1000000,10e-1000000]`, `[0,0,0,1e1000000,1e-999999]`},
		{"exponent-beyond-int64", `[10e999999999999999999999999999999,1.0e-999999999999999999999999999999]`, `[1e1000000000000000000000000000000,1e-999999999999999999999999999999]`},
		{"strings-order-and-scalars", `{"z":"<>&/","a":"\u0000\b\t\n\f\r\u001f","é":"\ud83d\ude00","�":"\ufffd","literal":"\\ud800"}`, "{\"a\":\"\\u0000\\b\\t\\n\\f\\r\\u001f\",\"literal\":\"\\\\ud800\",\"z\":\"<>&/\",\"é\":\"😀\",\"�\":\"�\"}"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := a.Codec().Canonical([]byte(c.input))
			if err != nil || string(got) != c.want {
				t.Fatalf("canonical mismatch: %q %v; want %q", got, err, c.want)
			}
		})
	}
	for _, input := range []string{`"\ud800"`, `"\udc00"`, `"\ud800x"`, `{"\ud800":0}`, `{"a":1,"a":2}`, `NaN`, `1 2`, string([]byte{'"', 0xff, '"'})} {
		if _, err := a.Codec().Canonical([]byte(input)); err == nil {
			t.Fatalf("invalid JSON/scalar accepted: %q", input)
		}
	}
	if !a.Codec().EqualJSON([]byte(`{"n":1.00}`), []byte(`{"n":1}`)) || a.Codec().EqualJSON([]byte(`9007199254740992`), []byte(`9007199254740993`)) {
		t.Fatal("semantic equality lost exact decimal distinctions")
	}
}

func TestCh10StorageNesting(t *testing.T) {
	a := opened(t, application(t), options(t, "nesting"))
	for _, kind := range []string{"array", "object"} {
		for _, depth := range []int{128, 129} {
			t.Run(fmt.Sprintf("%s-%d", kind, depth), func(t *testing.T) {
				left, right := "[", "]"
				if kind == "object" {
					left, right = `{"x":`, "}"
				}
				raw := []byte(strings.Repeat(left, depth) + "0" + strings.Repeat(right, depth))
				got, err := a.Codec().Canonical(raw)
				if depth == 128 {
					if err != nil || !bytes.Equal(got, raw) {
						t.Fatalf("exact nesting parent failed: %v", err)
					}
				} else {
					code(t, err, "session_corrupt")
				}
			})
		}
	}
}

func TestCh10StorageHandlerCount(t *testing.T) {
	a := opened(t, application(t), options(t, "handler-count"))
	cp, err := a.Codec().Decode(export(t, a).Bytes)
	if err != nil || len(cp.Identity.Handlers) == 0 {
		t.Fatalf("valid installed handler parent unavailable: %v", err)
	}
	id := cp.Identity
	prototype := id.Handlers[0]
	id.Handlers = nil
	for i := 0; i < 1024; i++ {
		h := prototype
		h.Name = fmt.Sprintf("handler_%04d", i)
		id.Handlers = append(id.Handlers, h)
	}
	if err = a.Codec().Identity(id); err != nil {
		t.Fatalf("exact 1024-handler parent refused: %v", err)
	}
	last := prototype
	last.Name = "handler_1024"
	id.Handlers = append(id.Handlers, last)
	code(t, a.Codec().Identity(id), "session_corrupt")
}

func TestCh10StorageHandlerBytes(t *testing.T) {
	a := opened(t, application(t), options(t, "handler-bytes"))
	cp, err := a.Codec().Decode(export(t, a).Bytes)
	if err != nil || len(cp.Identity.Handlers) == 0 {
		t.Fatalf("valid installed handler parent unavailable: %v", err)
	}
	id := cp.Identity
	id.Handlers = id.Handlers[:1]
	id.Handlers[0].Name = "h"
	id.Handlers[0].Schema = []byte(`{}`)
	// All description bytes are ASCII x, so this independent grammar count
	// equals canonical UTF-8 length without asking the implementation to size it.
	const grammar = `[{"description":"","name":"h","schema":{}}]`
	id.Handlers[0].Description = strings.Repeat("x", (16<<20)-len(grammar))
	if err = a.Codec().Identity(id); err != nil {
		t.Fatalf("exact 16 MiB handler parent refused: %v", err)
	}
	id.Handlers[0].Description += "x"
	code(t, a.Codec().Identity(id), "session_corrupt")
}

func TestCh10StorageRequestUint64(t *testing.T) {
	e := localEndpoint(t)
	o := wireOptions(t, "openai", e)
	a := opened(t, application(t), o)
	x := export(t, a)
	cp, err := a.Codec().Decode(x.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	// Request alone permits burned ordinals unsupported by recorded turns.
	// Do not manufacture unsupported activation/job maxima to get a parent.
	cp.HighWatermarks.Request = math.MaxUint64 - 1
	cp.State.Session.HighWatermarks.Request = math.MaxUint64 - 1
	cp.State.Context.RequestCursor = math.MaxUint64 - 1
	raw, err := a.Codec().Encode(cp)
	if err != nil {
		t.Fatal(err)
	}
	closed(t, a)
	o.Config.DataDir = filepath.Join(o.Config.Workspace, "max-request")
	b, err := application(t).ImportSession(raw, o)
	if err != nil {
		t.Fatalf("valid burned-cursor import refused: %v", err)
	}
	t.Cleanup(func() { _ = b.Close() })
	noStartupHTTP(t, e)
	turn(t, b, e, "openai", "LAST-REQUEST-IDENTITY")
	found := false
	for _, event := range b.Events() {
		if event.Type == "turn_started" {
			found = event.Turn != nil && event.Turn.RequestIndex == math.MaxUint64
		}
	}
	if !found {
		t.Fatal("last valid request did not retain exact uint64 identity")
	}
	before, err := os.ReadFile(filepath.Join(o.Config.DataDir, "events.log"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = b.Submit("MUST-NOT-WRAP")
	code(t, err, "session_limit")
	after, err := os.ReadFile(filepath.Join(o.Config.DataDir, "events.log"))
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("exhausted request mutated durable log")
	}
	noStartupHTTP(t, e)
	max := export(t, b)
	closed(t, b)
	o.Config.DataDir = filepath.Join(o.Config.Workspace, "max-reimport")
	c, err := application(t).ImportSession(max.Bytes, o)
	if err != nil {
		t.Fatalf("maximum uint64 cursor did not reimport: %v", err)
	}
	t.Cleanup(func() { _ = c.Close() })
	_, err = c.Submit("STILL-MUST-NOT-WRAP")
	code(t, err, "session_limit")
	noStartupHTTP(t, e)
}
