package common

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestCh20AuthoritativeSubmissionRows(t *testing.T) {
	schema := json.RawMessage(`{"type":"object","properties":{"n":{"type":"number","const":9007199254740993.0001},"fixed":{"type":"boolean"}},"required":["n","fixed"],"additionalProperties":false}`)
	payload := json.RawMessage(`{"n":9007199254740993.0001,"fixed":false}`)
	cfg := Config{ID: "worker", ParentID: "parent", SubmitSchema: schema}
	// Public row shape is declared independently of the production row struct.
	base := map[string]json.RawMessage{"version": json.RawMessage(`1`), "agent_id": json.RawMessage(`"worker"`), "parent_id": json.RawMessage(`"parent"`), "timestamp": json.RawMessage(`"2026-01-02T03:04:05Z"`), "submit_schema": schema, "payload": payload}
	valid, _ := json.Marshal(base)
	valid = append(valid, '\n')
	out, err := parseSubmissions(append(append([]byte(nil), valid...), valid...), cfg)
	want, _ := PrepareSubmission(schema, payload)
	if err != nil || !bytes.Equal(out, want) {
		t.Fatalf("valid exact rows: %s %v", out, err)
	}
	mutations := map[string]func(map[string]json.RawMessage){
		"unknown version": func(m map[string]json.RawMessage) { m["version"] = json.RawMessage(`2`) },
		"missing version": func(m map[string]json.RawMessage) { delete(m, "version") },
		"old v custom shape": func(m map[string]json.RawMessage) {
			delete(m, "version")
			delete(m, "submit_schema")
			m["v"] = json.RawMessage(`1`)
			m["custom"] = json.RawMessage(`true`)
		},
		"missing schema": func(m map[string]json.RawMessage) { delete(m, "submit_schema") },
		"null schema":    func(m map[string]json.RawMessage) { m["submit_schema"] = json.RawMessage(`null`) },
		"unsupported schema": func(m map[string]json.RawMessage) {
			m["submit_schema"] = json.RawMessage(`{"type":"object","$ref":"https://invalid.example/schema"}`)
		},
		"changed schema": func(m map[string]json.RawMessage) { m["submit_schema"] = json.RawMessage(`{"type":"object"}`) },
		"invalid payload number": func(m map[string]json.RawMessage) {
			m["payload"] = json.RawMessage(`{"n":9007199254740993,"fixed":false}`)
		},
		"wrong payload type": func(m map[string]json.RawMessage) {
			m["payload"] = json.RawMessage(`{"n":9007199254740993.0001,"fixed":"false"}`)
		},
		"wrong owner":   func(m map[string]json.RawMessage) { m["agent_id"] = json.RawMessage(`"other"`) },
		"wrong parent":  func(m map[string]json.RawMessage) { m["parent_id"] = json.RawMessage(`"other"`) },
		"mixed default": func(m map[string]json.RawMessage) { m["result"] = json.RawMessage(`"not a custom payload"`) },
	}
	for name, change := range mutations {
		t.Run(name, func(t *testing.T) {
			row := map[string]json.RawMessage{}
			for k, v := range base {
				row[k] = append(json.RawMessage(nil), v...)
			}
			change(row)
			bad, _ := json.Marshal(row)
			file := append(append(append([]byte(nil), valid...), bad...), '\n')
			file = append(file, valid...)
			if _, err := parseSubmissions(file, cfg); err == nil {
				t.Fatalf("invalid later row accepted: %s", bad)
			}
		})
	}
	loose := cfg
	loose.SubmitSchema = json.RawMessage(`{"type":"object"}`)
	badPayload := bytes.Replace(valid, []byte(`"fixed":false`), []byte(`"fixed":"false"`), 1)
	if _, err := parseSubmissions(badPayload, loose); err == nil || !strings.Contains(err.Error(), "fixed") {
		t.Fatalf("current config blessed invalid original contract: %v", err)
	}
	for _, raw := range [][]byte{nil, []byte("\n"), append(append([]byte(nil), valid...), []byte("garbage\n")...)} {
		if _, err := parseSubmissions(raw, cfg); err == nil {
			t.Fatalf("malformed file accepted: %q", raw)
		}
	}
}
