package main

import (
	"encoding/json"
	"testing"
)

func TestPublicContractJSON(t *testing.T) {
	var description struct {
		ReadFileTool string `json:"read_file_tool"`
		ReadFileArg  string `json:"read_file_arg"`
		TypeDuring   bool   `json:"supports_type_during_turn"`
		StreamingOff bool   `json:"supports_streaming_off"`
		LogFormat    string `json:"log_format"`
		Listener     struct {
			Tools []string `json:"tools"`
		} `json:"listener"`
	}
	if err := json.Unmarshal([]byte(contract), &description); err != nil {
		t.Fatal(err)
	}
	if description.ReadFileTool != "read_file" || description.ReadFileArg != "path" || !description.TypeDuring || !description.StreamingOff || description.LogFormat != "jsonl" || len(description.Listener.Tools) != 5 {
		t.Fatalf("contract: %+v", description)
	}
	want := []string{"hear", "type", "submit", "wait", "sleep"}
	for i, name := range want {
		if description.Listener.Tools[i] != name {
			t.Fatalf("unexpected listener tool %q", description.Listener.Tools[i])
		}
	}
}
