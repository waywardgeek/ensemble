package ensemble_test

import (
	"context"
	"ensemble"
	"ensemble/internal/common"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestChapter15R1DiagnosticFilesActualSend(t *testing.T) {
	var sent []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sent, _ = io.ReadAll(r.Body)
		_, _ = io.WriteString(w, `{"model":"claude-sonnet-4-6","content":[{"type":"text","text":"ok"}],"usage":{"input_tokens":20,"output_tokens":2,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}`)
	}))
	defer server.Close()
	dir := t.TempDir()
	a := r2Agent(t, ensemble.Config{Vendor: ensemble.Anthropic, Model: "claude-sonnet-4-6", APIKey: "local-fixture", BaseURL: server.URL, DataDir: dir, Tools: []string{}})
	defer a.Shutdown()
	_, err := a.Render()
	if err != nil {
		t.Fatal(err)
	}
	current := filepath.Join(dir, "cache-current.txt")
	prior := filepath.Join(dir, "cache-prior.txt")
	if b, _ := os.ReadFile(current); strings.Contains(string(b), `"messages"`) {
		t.Fatal("preview became a send")
	}
	body := []byte(`{
  "messages":[{"role":"user","content":[{"type":"text","text":"literal \"cache_control\":{\"type\":\"ephemeral\"}","cache_control":{"type":"ephemeral"}}]}],
  "model":"claude-sonnet-4-6"
}
 `)
	_, err = a.Engine().ExchangeRequest(context.Background(), body, common.ParseCallbacks{})
	if err != nil {
		t.Fatal(err)
	}
	first, err := os.ReadFile(current)
	if err != nil {
		t.Fatal(err)
	}
	if string(sent) != string(body) {
		t.Fatal("actual send reformatted frozen bytes")
	}
	clean := strings.Replace(string(body), `,"cache_control":{"type":"ephemeral"}`, "", 1)
	if !strings.Contains(string(first), "Marker-free comparison request (all other bytes unchanged):\n"+clean) {
		t.Fatalf("pretty marker-free comparison changed literal or whitespace: %s", first)
	}
	if !strings.Contains(string(first), string(sent)) {
		t.Fatalf("diagnostic omits exact sent body: %s", first)
	}
	_, err = a.Engine().ExchangeRequest(context.Background(), body, common.ParseCallbacks{})
	if err != nil {
		t.Fatal(err)
	}
	previous, err := os.ReadFile(prior)
	if err != nil || string(previous) != string(first) {
		t.Fatalf("prior is not first actual send: %v", err)
	}
}

func TestChapter15R1LateResponseKeepsOwnPrediction(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		read := int64(80)
		input := int64(20)
		if n == 1 {
			close(entered)
			<-release
			read = 0
			input = 100
		}
		fmt.Fprintf(w, `{"model":"claude-sonnet-4-6","content":[{"type":"text","text":"ok"}],"usage":{"input_tokens":%d,"output_tokens":2,"cache_read_input_tokens":%d,"cache_creation_input_tokens":0}}`, input, read)
	}))
	defer server.Close()
	a := r2Agent(t, ensemble.Config{Vendor: ensemble.Anthropic, Model: "claude-sonnet-4-6", APIKey: "local-fixture", BaseURL: server.URL, DataDir: t.TempDir(), Tools: []string{}})
	defer a.Shutdown()
	var once sync.Once
	defer once.Do(func() { close(release) })
	first := []byte(`{"model":"claude-sonnet-4-6","messages":[{"role":"user","content":"one"}]}`)
	second := []byte(`{"model":"claude-sonnet-4-6","messages":[{"role":"user","content":"two"}]}`)
	done := make(chan error, 1)
	go func() {
		_, err := a.Engine().ExchangeRequest(context.Background(), first, common.ParseCallbacks{})
		done <- err
	}()
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("first send did not arrive")
	}
	secondDone := make(chan error, 1)
	go func() {
		_, err := a.Engine().ExchangeRequest(context.Background(), second, common.ParseCallbacks{})
		secondDone <- err
	}()
	select {
	case err := <-secondDone:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("lens lock held across HTTP")
	}
	before := a.Engine().Status().Cache
	if before.Comparison == nil || before.Comparison.Prediction.RequestID != 2 || before.Comparison.Prediction.Fraction == nil || before.Comparison.ActualHitRate == nil || *before.Comparison.ActualHitRate != .8 {
		t.Fatalf("second correlation %+v", before)
	}
	once.Do(func() { close(release) })
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	after := a.Engine().Status().Cache
	if after.Current.RequestID != 2 || after.Comparison == nil || after.Comparison.Prediction.RequestID != 1 || after.Comparison.Prediction.Fraction != nil || after.Comparison.ActualHitRate == nil || *after.Comparison.ActualHitRate != 0 {
		t.Fatalf("late response borrowed current prediction %+v", after)
	}
}

func TestChapter15R1DiagnosticResumeDoesNotImportSends(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"cache-current.txt", "cache-prior.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("prior-process secret request"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	a := r2Agent(t, ensemble.Config{Vendor: ensemble.OpenAI, Model: "gpt-5.6", APIKey: "local-fixture", DataDir: dir, Tools: []string{}})
	defer a.Shutdown()
	if a.Engine().Status().Cache.Current != nil {
		t.Fatal("disk request was restored into lens")
	}
	for _, name := range []string{"cache-current.txt", "cache-prior.txt"} {
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil || !strings.Contains(string(b), "No main HTTP request") || strings.Contains(string(b), "secret") {
			t.Fatalf("dishonest resume %s: %s %v", name, b, err)
		}
	}
}
