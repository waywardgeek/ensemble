package ensemble_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ensemble/ensemble"
)

func TestLocalModelArtifactProductionTiming(t *testing.T) {
	for _, mode := range []struct {
		name      string
		streaming ensemble.Stream
		disabled  bool
	}{
		{"explicit", ensemble.StreamAll, false},
		{"unknown default", 0, false},
		{"disabled", ensemble.StreamAll, true},
	} {
		t.Run(mode.name, func(t *testing.T) {
			wantStream := mode.streaming != 0 && !mode.disabled
			observed := make(chan struct{}, 1)
			peer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				var body map[string]any
				if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
					t.Error(err)
					return
				}
				if (body["stream"] == true) != wantStream {
					t.Error("explicit streaming selection was not honored")
				}
				if body["stream"] != true {
					io.WriteString(writer, `{"choices":[{"message":{"content":"Hello world."},"finish_reason":"stop"}],"usage":{"prompt_tokens":8,"completion_tokens":8}}`)
					return
				}
				writer.Header().Set("Content-Type", "text/event-stream")
				frame(writer, wireFrame{"", decoded(`{"choices":[{"delta":{"content":"Hello "}}]}`)})
				select {
				case <-observed:
				case <-request.Context().Done():
					return
				}
				time.Sleep(100 * time.Millisecond)
				frame(writer, wireFrame{"", decoded(`{"choices":[{"delta":{"content":"world."}}]}`)})
				time.Sleep(250 * time.Millisecond)
				frame(writer, wireFrame{"", decoded(`{"choices":[{"delta":{},"finish_reason":"stop"}],"usage":{"prompt_tokens":8,"completion_tokens":8}}`)})
			}))
			defer peer.Close()
			root := ensemble.New(io.Discard)
			agent, err := root.NewAgent(ensemble.Config{Vendor: ensemble.OpenAI, Model: "local-timing-model", BaseURL: peer.URL, DataDir: t.TempDir(), Streaming: mode.streaming, DisableStreaming: mode.disabled, Recall: ensemble.RecallSettings{Disabled: true}})
			if err != nil {
				t.Fatal(err)
			}
			defer agent.Shutdown()
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			var parts []ensemble.Part
			var arrivals []int64
			answer, err := agent.AskWatching(ctx, "Check timing", ensemble.StreamCallbacks{
				OnDelta: func(_ uint64, kind ensemble.DeltaKind, chunk string) {
					if kind == ensemble.DeltaText && chunk != "" {
						arrivals = append(arrivals, time.Now().UnixMilli())
						if wantStream && len(arrivals) == 1 {
							observed <- struct{}{}
						}
					}
				},
				OnPartFinal: func(_ uint64, part ensemble.Part) { parts = append(parts, part) },
			})
			if err != nil || answer != "Hello world." || len(parts) != 1 {
				t.Fatalf("answer=%q error=%v parts=%d", answer, err, len(parts))
			}
			part := parts[0]
			if wantStream {
				if len(arrivals) != 2 {
					t.Fatalf("text fragments=%d, want 2", len(arrivals))
				}
				if part.Started == 0 || part.Elapsed < 70 || absMilliseconds(part.Elapsed-(arrivals[1]-arrivals[0])) > 50 {
					t.Errorf("missing text fragment duration: started=%d elapsed=%d", part.Started, part.Elapsed)
				}
			} else if part.Started != 0 || part.Elapsed != 0 {
				t.Error("buffered part invented a duration")
			}
		})
	}
}

func TestExplicitStreamingConfiguration(t *testing.T) {
	t.Setenv("EN_DISABLE_STREAMING", "0")
	for _, entry := range []struct {
		value string
		want  ensemble.Stream
		bad   bool
	}{
		{"", 0, false},
		{"all", ensemble.StreamAll, false},
		{" ALL ", ensemble.StreamAll, false},
		{"text,thinking,tools", 0, true},
		{"text", 0, true},
		{"unrecognized", 0, true},
	} {
		t.Run(fmt.Sprintf("selection=%q", entry.value), func(t *testing.T) {
			t.Setenv("EN_STREAMING", entry.value)
			cfg, err := ensemble.ConfigFromEnv()
			if (err != nil) != entry.bad {
				t.Fatalf("configuration error=%v", err)
			}
			if !entry.bad {
				cfg.Model = "local-timing-model"
				if got := ensemble.StreamingFor(cfg); got != entry.want {
					t.Errorf("effective streams=%v, want %v", got, entry.want)
				}
				cfg.DisableStreaming = true
				if ensemble.StreamingFor(cfg) != 0 {
					t.Error("no-stream did not override the explicit selection")
				}
			}
		})
	}
}
