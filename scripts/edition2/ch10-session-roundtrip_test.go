package ch10_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"example.com/ensemble"
)

type exchange struct {
	body  []byte
	reply chan []byte
}
type endpoint struct {
	server   *httptest.Server
	requests chan exchange
}

func localEndpoint(t *testing.T) *endpoint {
	t.Helper()
	e := &endpoint{requests: make(chan exchange, 8)}
	e.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			return
		}
		x := exchange{data, make(chan []byte, 1)}
		select {
		case e.requests <- x:
		case <-r.Context().Done():
			return
		}
		select {
		case data = <-x.reply:
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(data)
		case <-r.Context().Done():
		}
	}))
	t.Cleanup(e.server.Close)
	return e
}
func nextRequest(t *testing.T, e *endpoint) exchange {
	t.Helper()
	select {
	case x := <-e.requests:
		return x
	case <-time.After(5 * time.Second):
		t.Fatal("actual request did not reach controlled endpoint")
		return exchange{}
	}
}
func noStartupHTTP(t *testing.T, e *endpoint) {
	t.Helper()
	select {
	case <-e.requests:
		t.Fatal("construction/inspection sent HTTP")
	default:
	}
}
func answer(t *testing.T, vendor string, x exchange) {
	t.Helper()
	var response any
	switch vendor {
	case "anthropic":
		response = map[string]any{"model": "claude-sonnet-4-6", "content": []any{map[string]any{"type": "text", "text": "accepted fixture answer"}}, "stop_reason": "end_turn", "usage": map[string]int{"input_tokens": 3, "output_tokens": 2}}
	case "openai":
		response = map[string]any{"model": "gpt-4.1-mini-2025-04-14", "choices": []any{map[string]any{"index": 0, "message": map[string]any{"role": "assistant", "content": "accepted fixture answer"}, "finish_reason": "stop"}}, "usage": map[string]int{"prompt_tokens": 3, "completion_tokens": 2}}
	case "gemini":
		response = map[string]any{"modelVersion": "gemini-3.8-flash", "candidates": []any{map[string]any{"content": map[string]any{"role": "model", "parts": []any{map[string]any{"text": "accepted fixture answer"}}}, "finishReason": "STOP"}}, "usageMetadata": map[string]int{"promptTokenCount": 3, "candidatesTokenCount": 2}}
	}
	data, err := json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}
	x.reply <- data
}
func wireOptions(t *testing.T, vendor string, e *endpoint) ensemble.SessionOptions {
	t.Helper()
	o := options(t, "source")
	o.Config.Vendor = vendor
	o.Config.Model = map[string]string{"anthropic": "claude-sonnet-4-6", "openai": "gpt-4.1-mini-2025-04-14", "gemini": "models/gemini-3.8-flash"}[vendor]
	o.Config.ResolvedModel = strings.TrimPrefix(o.Config.Model, "models/")
	o.Config.BaseURL = e.server.URL
	o.Config.DisableStreaming = true
	return o
}
func finish(t *testing.T, h ensemble.RequestHandle) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c, err := h.Wait(ctx)
	if err != nil || c.Outcome != "success" {
		t.Fatalf("request did not finish successfully: %+v %v", c, err)
	}
}
func turn(t *testing.T, a *ensemble.Agent, e *endpoint, vendor, prompt string) []byte {
	t.Helper()
	h, err := a.Submit(prompt)
	if err != nil {
		t.Fatal(err)
	}
	x := nextRequest(t, e)
	answer(t, vendor, x)
	finish(t, h)
	return x.body
}

func TestCh10PublicNonemptyTailRebuildAndImport(t *testing.T) {
	for _, vendor := range []string{"anthropic", "openai", "gemini"} {
		t.Run(vendor, func(t *testing.T) {
			e := localEndpoint(t)
			o := wireOptions(t, vendor, e)
			a := opened(t, application(t), o)
			noStartupHTTP(t, e)
			turn(t, a, e, vendor, "FIRST-CH10-MARKER")
			old := export(t, a)
			turn(t, a, e, vendor, "SECOND-CH10-MARKER")
			latest := export(t, a)
			if latest.AsOf <= old.AsOf {
				t.Fatal("fixture has no genuine newer tail")
			}
			baselineUsage := a.UsageByModel()
			var priorSend uint64
			for _, event := range a.Events() {
				if event.Type == "request_sent" {
					priorSend = event.Seq
				}
			}
			closed(t, a)
			var bodies [][]byte
			var totals []map[ensemble.Provenance]ensemble.Usage
			for _, mode := range []string{"older-snapshot", "null-rebuild", "snapshot-only-import"} {
				c := o
				c.Config.DataDir = filepath.Join(o.Config.Workspace, mode)
				app := application(t)
				var b *ensemble.Agent
				var err error
				if mode == "snapshot-only-import" {
					b, err = app.ImportSession(latest.Bytes, c)
				} else {
					copies(t, o.Config.DataDir, c.Config.DataDir)
					checkpoint := old.Bytes
					if mode == "null-rebuild" {
						var fields map[string]json.RawMessage
						if err = json.Unmarshal(checkpoint, &fields); err != nil {
							t.Fatal(err)
						}
						fields["state"], fields["state_sha256"] = json.RawMessage("null"), json.RawMessage("null")
						checkpoint, err = json.Marshal(fields)
						if err != nil {
							t.Fatal(err)
						}
					}
					if err = os.WriteFile(filepath.Join(c.Config.DataDir, "checkpoint.json"), checkpoint, 0600); err != nil {
						t.Fatal(err)
					}
					b, err = app.OpenSession(c)
				}
				if err != nil {
					t.Fatalf("valid %s refused: %v", mode, err)
				}
				noStartupHTTP(t, e)
				if !reflect.DeepEqual(baselineUsage, b.UsageByModel()) {
					t.Fatalf("%s lost/doubled accepted usage", mode)
				}
				if mode == "snapshot-only-import" {
					_, err = b.ReconstructRequest(priorSend)
					code(t, err, "history_unavailable")
					snapshot, watch, err := b.Watch()
					if err != nil {
						t.Fatal(err)
					}
					watch.Close()
					if len(b.Events()) != 1 || len(snapshot.Events) <= 1 {
						t.Fatal("import conflated actual anchor history with saved display window")
					}
				}
				body := turn(t, b, e, vendor, "THIRD-CH10-MARKER")
				for _, marker := range []string{"FIRST-CH10-MARKER", "SECOND-CH10-MARKER", "THIRD-CH10-MARKER"} {
					if bytes.Count(body, []byte(marker)) != 1 {
						t.Fatalf("%s lost/duplicated %s", mode, marker)
					}
				}
				var send uint64
				for _, event := range b.Events() {
					if event.Type == "request_sent" {
						send = event.Seq
					}
				}
				replayed, err := b.ReconstructRequest(send)
				if err != nil || !bytes.Equal(replayed, body) {
					t.Fatalf("%s available send reconstruct differs: %v", mode, err)
				}
				bodies = append(bodies, body)
				totals = append(totals, b.UsageByModel())
				closed(t, b)
			}
			for i := 1; i < len(bodies); i++ {
				if !bytes.Equal(bodies[0], bodies[i]) || !reflect.DeepEqual(totals[0], totals[i]) {
					t.Fatal("snapshot/tail, null rebuild and prefix-free import diverged")
				}
			}
		})
	}
}

func TestCh10PublicCaptureBusyDuringHTTP(t *testing.T) {
	e := localEndpoint(t)
	a := opened(t, application(t), wireOptions(t, "openai", e))
	h, err := a.Submit("held capture barrier")
	if err != nil {
		t.Fatal(err)
	}
	x := nextRequest(t, e)
	for _, operation := range []string{"export", "checkpoint"} {
		done := make(chan error, 1)
		go func() {
			var err error
			if operation == "export" {
				_, err = a.ExportCheckpoint()
			} else {
				_, err = a.Checkpoint()
			}
			done <- err
		}()
		select {
		case err := <-done:
			code(t, err, "session_busy")
		case <-time.After(time.Second):
			t.Fatal("capture waited for held HTTP instead of refusing busy")
		}
	}
	answer(t, "openai", x)
	finish(t, h)
	if _, err = a.Checkpoint(); err != nil {
		t.Fatalf("settled positive failed after busy refusal: %v", err)
	}
}
