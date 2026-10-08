package ensemble_test

import (
	"bytes"
	"context"
	"errors"
	"example.com/ensemble"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
)

func sessionFixture(t *testing.T) (*ensemble.Ensemble, ensemble.SessionOptions) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"model":"fixture","content":[{"type":"text","text":"remembered"}],"usage":{"input_tokens":3,"output_tokens":2}}`)
	}))
	t.Cleanup(server.Close)
	root := ensemble.New(io.Discard)
	t.Cleanup(func() { root.Close() })
	workspace := t.TempDir()
	return root, ensemble.SessionOptions{Config: ensemble.Config{DataDir: "session", Workspace: workspace, Vendor: "anthropic", Model: "fixture", APIKey: "local-fixture-only", BaseURL: server.URL, DisableStreaming: true}}
}
func TestSessionResumeImportAndRawBoundary(t *testing.T) {
	root, options := sessionFixture(t)
	a, err := root.OpenSession(options)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = a.Ask(context.Background(), "remember marker"); err != nil {
		t.Fatal(err)
	}
	exported, err := a.ExportCheckpoint()
	if err != nil {
		t.Fatal(err)
	}
	before, err := a.Render(a.Config())
	if err != nil {
		t.Fatal(err)
	}
	id, _ := a.Session()
	if err = a.Close(); err != nil {
		t.Fatal(err)
	}
	b, err := root.OpenSession(options)
	if err != nil {
		t.Fatal(err)
	}
	resumed, _ := b.Session()
	if !resumed.Resumed || resumed.ID != id.ID || b.ID() == a.ID() {
		t.Fatal("identity not preserved/scoped")
	}
	after, err := b.Render(b.Config())
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("render changed: %v", err)
	}
	if b.Usage() != a.Usage() {
		t.Fatal("usage changed")
	}
	_, err = root.OpenSession(options)
	var se *ensemble.SessionError
	if !errors.As(err, &se) || se.Code != "session_in_use" {
		t.Fatalf("second owner: %v", err)
	}
	if err = b.Close(); err != nil {
		t.Fatal(err)
	}
	options.Config.DataDir = "imported"
	c, err := root.ImportSession(exported.Bytes, options)
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Events()) != 1 || c.Events()[0].Type != "session_anchor" {
		t.Fatal("invented raw history")
	}
	if _, err = c.ReconstructRequest(3); !errors.As(err, &se) || se.Code != "history_unavailable" {
		t.Fatalf("absent prefix: %v", err)
	}
	if _, err = c.Ask(context.Background(), "continue"); err != nil {
		t.Fatal(err)
	}
	if err = c.Close(); err != nil {
		t.Fatal(err)
	}
	d, err := root.OpenSession(options)
	if err != nil {
		t.Fatal(err)
	}
	if d.Usage().Input != 6 {
		t.Fatal("import replay double counted")
	}
	if err = d.Close(); err != nil {
		t.Fatal(err)
	}
	inspected, err := root.InspectSession(filepath.Join(options.Config.Workspace, "imported"))
	if err != nil {
		t.Fatal(err)
	}
	if inspected.Boundary().CompleteHistory || inspected.Boundary().OriginAsOf != exported.AsOf {
		t.Fatal("missing origin boundary")
	}
}
func TestSessionCheckpointCorruptionPreservesSource(t *testing.T) {
	root, options := sessionFixture(t)
	a, err := root.OpenSession(options)
	if err != nil {
		t.Fatal(err)
	}
	if err = a.Close(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(options.Config.Workspace, "session", "checkpoint.json")
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	corrupt := bytes.Replace(original, []byte(`"state_version":1`), []byte(`"state_version":2`), 1)
	if err = os.WriteFile(path, corrupt, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = root.OpenSession(options); err == nil {
		t.Fatal("corruption accepted")
	}
	actual, _ := os.ReadFile(path)
	if !bytes.Equal(actual, corrupt) {
		t.Fatal("refusal rewrote source")
	}
	if err = os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	b, err := root.OpenSession(options)
	if err != nil {
		t.Fatalf("failed open retained reservation: %v", err)
	}
	if err = b.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestSessionSkillsImportAndResume(t *testing.T) {
	root := ensemble.New(io.Discard)
	defer root.Close()
	config := skillConfig(t)
	config.LogPath = ""
	config.DataDir = "session"
	options := ensemble.SessionOptions{Config: config}
	a, err := root.OpenSession(options)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = a.LoadSkill("edit"); err != nil {
		t.Fatal(err)
	}
	before, err := a.Render(a.Config())
	if err != nil {
		t.Fatal(err)
	}
	exported, err := a.ExportCheckpoint()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = root.InspectCheckpoint(exported.Bytes); err != nil {
		t.Fatalf("inspect skill snapshot: %v", err)
	}
	if err = a.Close(); err != nil {
		t.Fatal(err)
	}
	b, err := root.OpenSession(options)
	if err != nil {
		t.Fatalf("resume: %v", err)
	}
	after, err := b.Render(b.Config())
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("skill render changed", err)
	}
	if err = b.Close(); err != nil {
		t.Fatal(err)
	}
	options.Config.DataDir = "imported"
	c, err := root.ImportSession(exported.Bytes, options)
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if _, err = c.UnloadSkill("edit"); err != nil {
		t.Fatal(err)
	}
	if err = c.Close(); err != nil {
		t.Fatal(err)
	}
	d, err := root.OpenSession(options)
	if err != nil {
		t.Fatalf("resume imported skills: %v", err)
	}
	if err = d.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestSessionOneShotSettingSurvivesAndConsumesUnknownCall(t *testing.T) {
	responses := []string{
		`{"type":"tool_use","id":"setter","name":"tool_limits","input":{"ai_callback_pattern":"","max_output_bytes":17}}`,
		`{"type":"text","text":"setting retained"}`,
		`{"type":"tool_use","id":"unknown","name":"missing_tool","input":{}}`,
		`{"type":"text","text":"attempt refused"}`,
	}
	var n atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i := int(n.Add(1)) - 1
		if i >= len(responses) {
			t.Error("unexpected request")
			http.Error(w, "extra request", 500)
			return
		}
		io.WriteString(w, `{"content":[`+responses[i]+`],"usage":{"input_tokens":1,"output_tokens":1}}`)
	}))
	defer server.Close()
	root := ensemble.New(io.Discard)
	defer root.Close()
	options := ensemble.SessionOptions{Config: ensemble.Config{DataDir: "session", Workspace: t.TempDir(), Model: "fixture", APIKey: "fixture", BaseURL: server.URL, DisableStreaming: true, Builtins: []string{"tool_limits"}}}
	a, err := root.OpenSession(options)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = a.Ask(context.Background(), "set a one-shot limit"); err != nil {
		t.Fatal(err)
	}
	exported, err := a.ExportCheckpoint()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = root.InspectCheckpoint(exported.Bytes); err != nil {
		t.Fatal("saved setter state invalid", err)
	}
	if err = a.Close(); err != nil {
		t.Fatal(err)
	}
	b, err := root.OpenSession(options)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = b.Ask(context.Background(), "attempt an unknown tool"); err != nil {
		t.Fatal(err)
	}
	var set, consume, dispatch uint64
	for _, e := range b.Events() {
		switch e.Type {
		case "tool_limits_set":
			set = e.Seq
		case "tool_limits_consumed":
			consume = e.Seq
		case "tool_called":
			if e.Tool.CallID == "unknown" {
				dispatch = e.Seq
			}
		}
	}
	if set == 0 || consume <= set || consume >= dispatch {
		t.Fatalf("wrong setting order: %d %d %d", set, consume, dispatch)
	}
	exported, err = b.ExportCheckpoint()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = root.InspectCheckpoint(exported.Bytes); err != nil {
		t.Fatal("consumed state invalid", err)
	}
	if err = b.Close(); err != nil {
		t.Fatal(err)
	}
	if c, err := root.OpenSession(options); err != nil {
		t.Fatal(err)
	} else {
		c.Close()
	}
}

func TestSessionCompleteUnterminatedRecordCanContinue(t *testing.T) {
	root, options := sessionFixture(t)
	a, err := root.OpenSession(options)
	if err != nil {
		t.Fatal(err)
	}
	if err = a.Close(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(options.Config.Workspace, options.Config.DataDir, "events.log")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, bytes.TrimSuffix(raw, []byte{'\n'}), 0600); err != nil {
		t.Fatal(err)
	}
	b, err := root.OpenSession(options)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = b.Ask(context.Background(), "continue after valid EOF"); err != nil {
		t.Fatal(err)
	}
	if err = b.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = root.InspectSession(filepath.Dir(path)); err != nil {
		t.Fatal(err)
	}
}
