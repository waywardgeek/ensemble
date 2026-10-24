package engine

import (
	"bufio"
	"encoding/json"
	"ensemble/internal/common"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type parseAgent struct {
	common.Agent
	config common.Config
}

func (a parseAgent) Config() common.Config { return a.config }
func parser(v common.Vendor, s common.Surface) *engine {
	return &engine{parent: parseAgent{config: common.Config{Vendor: v, Surface: s, Model: "route", ResolvedModel: "resolved"}}}
}
func parseBody(e *engine, body, contentType string, cb common.ParseCallbacks) ([]common.Event, error) {
	return e.Parse(&http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {contentType}}, Body: io.NopCloser(strings.NewReader(body))}, cb)
}
func ended(t *testing.T, events []common.Event) *common.ResponseData {
	t.Helper()
	for _, e := range events {
		if e.Type == common.ResponseEnded {
			return e.Response
		}
	}
	t.Fatal("missing response end")
	return nil
}
func TestSSEFraming(t *testing.T) {
	large := strings.Repeat("z", 80*1024)
	r := sseReader{bufio.NewReader(strings.NewReader(":comment\r\nevent: ignored\r\n\r\nevent: chunk\r\ndata: first\r\ndata: " + large + "\r\n\r\ndata: tail"))}
	f, err := r.next()
	if err != nil || f.event != "chunk" || f.data != "first\n"+large {
		t.Fatalf("bad frame: event=%q bytes=%d err=%v", f.event, len(f.data), err)
	}
	f, err = r.next()
	if err != nil || f.data != "tail" {
		t.Fatalf("trailing frame %v %v", f, err)
	}
	if _, err = r.next(); err != io.EOF {
		t.Fatal(err)
	}
}
func TestPlainAndStreamFinalParity(t *testing.T) {
	for _, tc := range []struct {
		name          string
		vendor        common.Vendor
		surface       common.Surface
		plain, stream string
	}{
		{"openai", common.OpenAI, common.ChatCompletions,
			`{"model":"actual","choices":[{"message":{"content":"hello"},"finish_reason":"stop"}],"usage":{"prompt_tokens":3,"completion_tokens":2}}`,
			"data: {\"model\":\"actual\",\"choices\":[{\"index\":0,\"delta\":{\"content\":\"hel\"}}]}\n\ndata: {\"choices\":[{\"index\":0,\"delta\":{\"content\":\"lo\"},\"finish_reason\":\"stop\"}],\"usage\":{\"prompt_tokens\":3,\"completion_tokens\":2}}\n\ndata: [DONE]\n\n"},
		{"anthropic", common.Anthropic, common.Messages,
			`{"model":"actual","content":[{"type":"text","text":"hello"}],"stop_reason":"end_turn","usage":{"input_tokens":3,"output_tokens":2}}`,
			"event: message_start\ndata: {\"type\":\"message_start\",\"message\":{\"model\":\"actual\",\"usage\":{\"input_tokens\":3}}}\n\ndata: {\"type\":\"content_block_start\",\"index\":0,\"content_block\":{\"type\":\"text\",\"text\":\"hel\"}}\n\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"lo\"}}\n\ndata: {\"type\":\"content_block_stop\",\"index\":0}\n\ndata: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"end_turn\"},\"usage\":{\"output_tokens\":2}}\n\ndata: {\"type\":\"message_stop\"}\n\n"},
		{"gemini", common.Gemini, common.GenerateContent,
			`{"modelVersion":"actual","candidates":[{"content":{"parts":[{"text":"hello"}]},"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":3,"candidatesTokenCount":2}}`,
			"data: {\"modelVersion\":\"actual\",\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"hel\"}]}}]}\n\ndata: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"lo\"}]},\"finishReason\":\"STOP\"}],\"usageMetadata\":{\"promptTokenCount\":3,\"candidatesTokenCount\":2}}\n\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := parser(tc.vendor, tc.surface)
			var deltas []common.PartDelta
			frames, events := 0, 0
			stream, err := parseBody(e, tc.stream, "text/event-stream; charset=utf-8", common.ParseCallbacks{OnDelta: func(d common.PartDelta) { deltas = append(deltas, d) }, OnFrame: func(common.Frame) { frames++ }, OnEvent: func(common.Event) { events++ }})
			if err != nil {
				t.Fatal(err)
			}
			var plainDeltas []common.PartDelta
			plain, err := parseBody(e, tc.plain, "application/json", common.ParseCallbacks{OnDelta: func(d common.PartDelta) { plainDeltas = append(plainDeltas, d) }})
			if err != nil {
				t.Fatal(err)
			}
			a, _ := json.Marshal(ended(t, stream))
			b, _ := json.Marshal(ended(t, plain))
			if string(a) != string(b) {
				t.Fatalf("stream %s\nplain %s", a, b)
			}
			if len(deltas) != 2 || len(plainDeltas) != 1 || deltas[0].PartID != 0 || deltas[1].PartID != 0 || frames < 2 || events != len(stream) {
				t.Fatalf("deltas=%+v plain=%+v frames=%d events=%d", deltas, plainDeltas, frames, events)
			}
			if deltas[0].Chunk+deltas[1].Chunk != "hello" {
				t.Fatal(deltas)
			}
		})
	}
}
func TestInterruptedToolNotFinal(t *testing.T) {
	e := parser(common.OpenAI, common.ChatCompletions)
	wire := "data: {\"choices\":[{\"delta\":{\"tool_calls\":[{\"index\":0,\"id\":\"call\",\"function\":{\"name\":\"run_command\",\"arguments\":\"{\\\"command\\\":\"}}]}}],\"usage\":{\"prompt_tokens\":7,\"completion_tokens\":1}}\n\n"
	var deltas []common.PartDelta
	events, err := parseBody(e, wire, "text/event-stream", common.ParseCallbacks{OnDelta: func(d common.PartDelta) { deltas = append(deltas, d) }})
	if err == nil || !strings.Contains(err.Error(), "incomplete stream") {
		t.Fatalf("expected incomplete-stream error, got %v", err)
	}
	if len(deltas) != 2 || deltas[0].Chunk != "run_command" || deltas[1].Chunk != `{"command":` {
		t.Fatalf("name/partial-args chunks=%+v", deltas)
	}
	for _, d := range deltas {
		if d.Kind != common.DeltaToolCall || d.PartID != 0 {
			t.Fatalf("changed kind or zero ID: %+v", d)
		}
	}
	response := ended(t, events)
	if len(response.Parts) != 0 || response.Usage != (common.Usage{Input: 7, Output: 1}) {
		t.Fatalf("partial executable or lost charge: %+v", response)
	}
}

func TestDeltaArrivesBeforeResponseEnds(t *testing.T) {
	e := parser(common.OpenAI, common.ChatCompletions)
	r, w := io.Pipe()
	defer r.Close()
	defer w.Close()
	first := make(chan common.PartDelta, 1)
	done := make(chan error, 1)
	go func() {
		_, err := e.Parse(&http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"text/event-stream"}}, Body: r}, common.ParseCallbacks{OnDelta: func(d common.PartDelta) { first <- d }})
		done <- err
	}()
	prefix := "data: {\"choices\":[{\"delta\":{\"content\":\"early\"}}]}\n\n"
	if _, err := io.WriteString(w, prefix); err != nil {
		t.Fatal(err)
	}
	select {
	case d := <-first:
		if d.PartID != 0 || d.Chunk != "early" {
			t.Fatal(d)
		}
	case <-time.After(time.Second):
		t.Fatal("delta waited for final response")
	}
	select {
	case err := <-done:
		t.Fatalf("response prematurely ended: %v", err)
	default:
	}
	if _, err := io.WriteString(w, "data: {\"choices\":[{\"delta\":{},\"finish_reason\":\"stop\"}],\"usage\":{\"prompt_tokens\":1,\"completion_tokens\":1}}\n\ndata: [DONE]\n\n"); err != nil {
		t.Fatal(err)
	}
	w.Close()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("parser did not finish")
	}
}
