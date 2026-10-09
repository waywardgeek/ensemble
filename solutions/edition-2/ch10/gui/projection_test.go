package gui

import (
	"bytes"
	"encoding/json"
	"example.com/ensemble"
	"testing"
)

func TestProjectionPreservesIntegerTokensAndOpaqueBoundary(t *testing.T) {
	app := ensemble.New(nil)
	defer app.Close()
	server := &Server{parent: app}
	part := ensemble.Part{Type: "tool_result", Parts: []ensemble.Part{
		{Type: "tool_call", Args: json.RawMessage(`{"revision":18446744073709551615,"text":"revision:9007199254740993"}`)},
		{Type: "opaque", Opaque: json.RawMessage(`{"private":"hidden"}`)},
	}}
	event := ensemble.Event{Type: "user", Seq: ^uint64(0), Message: &ensemble.Entry{Parts: []ensemble.Part{part}}}
	observation := ensemble.Observation{ExecutionPolicy: &ensemble.PolicySnapshot{Revision: ^uint64(0)}, Event: event}
	for name, value := range map[string]any{
		"part": ProjectPart(server, part), "event": ProjectEvent(server, event), "observation": projectObservation(server, observation),
	} {
		wire, err := json.Marshal(value)
		if err != nil || !bytes.Contains(wire, []byte(`"revision":18446744073709551615`)) ||
			!bytes.Contains(wire, []byte(`"text":"revision:9007199254740993"`)) ||
			bytes.Contains(wire, []byte("hidden")) || !bytes.Contains(wire, []byte(`"placeholder":true`)) {
			t.Fatalf("%s loses numeric or opaque boundary: %s / %v", name, wire, err)
		}
	}
}
