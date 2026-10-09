// A public offline projection consumer. Original logs are never modified and
// no live Agent, credentials, catalog, job or provider request is constructed.
package main

import (
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"example.com/ensemble"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	path := flag.String("log", "", "original retained event log")
	result := flag.Uint64("result", 0, "actual tool_returned sequence to redact in the derived projection")
	flag.Parse()
	raw, err := os.ReadFile(*path)
	if err != nil {
		return fmt.Errorf("cannot read retained log")
	}
	app := ensemble.New(os.Stderr)
	defer app.Close()
	original, err := app.Load(*path, ensemble.Config{})
	if err != nil {
		return err
	}
	var config ensemble.Config
	var requestSeq uint64
	found := false
	for _, e := range original.Events() {
		if e.Seq == *result && e.Type == "tool_returned" {
			found = true
		}
		if e.Type == "request_sent" && e.Request.Configuration != nil {
			c := e.Request.Configuration
			config = ensemble.Config{DisableStreaming: e.Request.Delivery != "stream", Vendor: e.Request.To.Vendor, Model: e.Request.To.Model, ResolvedModel: c.ResolvedModel, System: c.System, Tools: c.Tools, MaxTokens: c.MaxTokens}
			requestSeq = e.Seq
		}
	}
	if !found || requestSeq == 0 {
		return fmt.Errorf("select a real tool result in a log with captured requests")
	}
	// Reconstruction is explicitly separate from the later end-of-log projections.
	captured, err := original.ReconstructRequest(requestSeq)
	if err != nil {
		return err
	}
	before, err := original.Render(config)
	if err != nil {
		return err
	}
	event := ensemble.Event{Seq: original.Snapshot().LastSeq + 1, Time: time.Now().UTC().Format(time.RFC3339Nano), Type: "redacted", Redact: &ensemble.Redaction{From: *result, To: *result, Level: "redact_result", Reason: "Chapter 9 result-only lifetime demonstration"}}
	appended, err := json.Marshal(event)
	if err != nil {
		return err
	}
	file, err := os.CreateTemp("", "ch09-derived-redaction-*.jsonl")
	if err != nil {
		return err
	}
	name := file.Name()
	defer os.Remove(name)
	if _, err = file.Write(append(append(append([]byte{}, raw...), '\n'), append(appended, '\n')...)); err != nil {
		file.Close()
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	// Public Load validates the derived redaction event; offline Append remains
	// read-only. No initialization or recorded skill transition is re-executed.
	derived, err := app.Load(name, ensemble.Config{})
	if err != nil {
		return err
	}
	after, err := derived.Render(config)
	if err != nil {
		return err
	}
	if original.Usage() != derived.Usage() {
		return fmt.Errorf("redaction changed accounting")
	}
	return json.NewEncoder(os.Stdout).Encode(struct {
		Scope           string          `json:"scope"`
		OriginalSHA256  string          `json:"original_sha256"`
		ResultSequence  uint64          `json:"result_sequence"`
		RequestSequence uint64          `json:"original_request_sequence"`
		Captured        json.RawMessage `json:"reconstructed_original_request"`
		Before          json.RawMessage `json:"end_projection_before_redaction"`
		After           json.RawMessage `json:"end_projection_after_redaction"`
		NewRequests     int             `json:"new_model_requests"`
	}{"offline derived projection; original receipts unchanged", fmt.Sprintf("%x", sha256.Sum256(raw)), *result, requestSeq, captured, before, after, 0})
}
