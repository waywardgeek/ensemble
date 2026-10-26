package ensemble_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ensemble/ensemble"
)

// BenchmarkContextSnapshotTurn measures the real actor and renderer against a
// local HTTP peer. Resetting recorded history outside the timer keeps every
// iteration at the same size without timing restore or depending on a model.
func BenchmarkContextSnapshotTurn(b *testing.B) {
	var log bytes.Buffer
	text := strings.Repeat("retained context ", 2048)
	encoder := json.NewEncoder(&log)
	for index := 0; index < 32; index++ {
		seq := index*3 + 1
		encoder.Encode(map[string]any{"seq": seq, "type": "message_received", "message": map[string]any{"actor": "human", "parts": []map[string]string{{"type": "text", "text": text}}}})
		encoder.Encode(map[string]any{"seq": seq + 1, "type": "request_sent", "request": map[string]any{"to": map[string]string{"vendor": "anthropic", "model": "benchmark-local", "surface": "messages"}}})
		encoder.Encode(map[string]any{"seq": seq + 2, "type": "response_ended", "response": map[string]any{"from": map[string]string{"vendor": "anthropic", "model": "benchmark-local", "surface": "messages"}, "parts": []map[string]string{{"type": "text", "text": "recorded answer"}}, "usage": map[string]int{}}})
	}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		io.Copy(io.Discard, request.Body)
		replyText(writer, "done")
	}))
	defer server.Close()
	agent, err := ensemble.New(io.Discard).NewAgent(ensemble.Config{Model: "benchmark-local", BaseURL: server.URL, DataDir: b.TempDir(), Recall: ensemble.RecallSettings{Disabled: true}})
	if err != nil {
		b.Fatal(err)
	}
	defer agent.Shutdown()
	b.ReportAllocs()
	b.ResetTimer()
	for iteration := 0; iteration < b.N; iteration++ {
		b.StopTimer()
		if err := agent.History().Load(bytes.NewReader(log.Bytes())); err != nil {
			b.Fatal(err)
		}
		b.StartTimer()
		if answer, err := agent.Ask(context.Background(), "next"); err != nil || answer != "done" {
			b.Fatalf("turn: %q, %v", answer, err)
		}
	}
}
