package llm

import (
	"context"
	"encoding/json"
	"example.com/ensemble/internal/common"
	"io"
	"reflect"
	"strings"
	"testing"
)

type streamRoot struct{}

func (streamRoot) Logf(string, ...any)                              {}
func (streamRoot) Publish(string, common.Event)                     {}
func (streamRoot) AllocateHandle() uint64                           { return 1 }
func (streamRoot) Observe(common.Observation)                       {}
func (streamRoot) Collect([]common.RequestHandle) common.Collection { return nil }

type streamAgent struct{ fragments []common.Fragment }

func (*streamAgent) Ensemble() common.Ensemble { return streamRoot{} }
func (*streamAgent) Config() common.Config     { return common.Config{} }
func (*streamAgent) Workspace() string         { return "" }
func (a *streamAgent) ModelReady(op common.ModelOperation) {
	for {
		f, more := op.Drain()
		a.fragments = append(a.fragments, f...)
		if !more {
			return
		}
	}
}
func testOperation() (common.ModelOperation, *streamAgent) {
	a := &streamAgent{}
	e := New(a)
	return e.NewOperation("m1", "r1", common.Config{}), a
}

type splitReader struct {
	data  []byte
	first int
}

func (r *splitReader) Read(p []byte) (int, error) {
	if len(r.data) == 0 {
		return 0, io.EOF
	}
	n := 1
	if r.first > 0 {
		n = r.first
		r.first = 0
	}
	if n > len(p) {
		n = len(p)
	}
	if n > len(r.data) {
		n = len(r.data)
	}
	copy(p, r.data[:n])
	r.data = r.data[n:]
	return n, nil
}
func TestSSEPhysicalFraming(t *testing.T) {
	for _, ending := range []string{"\n", "\r\n", "\r"} {
		fixture := "\xef\xbb\xbf: keepalive" + ending + "id: ignored" + ending + "event: message" + ending + "data: {\"a\":" + ending + "data: \"é\"}" + ending + ending
		for split := 0; split < len(fixture); split++ {
			op, _ := testOperation()
			r := newSSE(op, &splitReader{[]byte(fixture), split})
			event, err := r.next()
			if err != nil || event.data != "{\"a\":\n\"é\"}" {
				t.Fatalf("ending %q split %d: %#v %v", ending, split, event, err)
			}
			if _, err = r.next(); err != io.EOF {
				t.Fatal(err)
			}
		}
	}
	for _, fixture := range []string{"data: {}\n", "data: {}", "data: \xff\n\n"} {
		op, _ := testOperation()
		if _, err := newSSE(op, strings.NewReader(fixture)).next(); err == nil {
			t.Fatalf("accepted %q", fixture)
		}
	}
}
func TestSSEExactFrameBounds(t *testing.T) {
	for _, ending := range []string{"\n", "\r\n", "\r"} {
		for _, comment := range []bool{false, true} {
			prefix, suffix := "data: \"", "\""+ending+ending
			if comment {
				prefix = ":"
				suffix = ending + ending
			}
			exact := prefix + strings.Repeat("x", pendingLimit-len(prefix)-len(suffix)) + suffix
			for _, bom := range []string{"", "\xef\xbb\xbf"} {
				op, _ := testOperation()
				r := newSSE(op, strings.NewReader(bom+exact+exact))
				if comment {
					if _, err := r.next(); err != io.EOF {
						t.Fatal(err)
					}
				} else {
					for i := 0; i < 2; i++ {
						if _, err := r.next(); err != nil {
							t.Fatal(err)
						}
					}
				}
			}
			op, _ := testOperation()
			if _, err := newSSE(op, strings.NewReader(prefix+"x"+exact[len(prefix):])).next(); err == nil {
				t.Fatal("accepted one-byte frame overflow")
			}
		}
	}
}
func messagesFixture() (string, string) {
	stream := ""
	for _, v := range []string{
		`{"type":"message_start","message":{"model":"fixture","usage":{"input_tokens":10,"output_tokens":0}}}`,
		`{"type":"content_block_start","index":0,"content_block":{"type":"text","text":"H"}}`,
		`{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"éllo."}}`,
		`{"type":"content_block_stop","index":0}`,
		`{"type":"content_block_start","index":1,"content_block":{"type":"thinking","thinking":"Check","signature":""}}`,
		`{"type":"content_block_delta","index":1,"delta":{"type":"thinking_delta","thinking":"ing."}}`,
		`{"type":"content_block_delta","index":1,"delta":{"type":"signature_delta","signature":"signed-"}}`,
		`{"type":"content_block_delta","index":1,"delta":{"type":"signature_delta","signature":"value"}}`,
		`{"type":"content_block_stop","index":1}`,
		`{"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":2}}`,
		`{"type":"message_stop"}`} {
		stream += "data: " + v + "\n\n"
	}
	plain := `{"model":"fixture","content":[{"type":"text","text":"Héllo."},{"type":"thinking","thinking":"Checking.","signature":"signed-value"}],"stop_reason":"end_turn","usage":{"input_tokens":10,"output_tokens":2}}`
	return stream, plain
}
func TestStreamPlainSemanticEquivalence(t *testing.T) {
	ms, mp := messagesFixture()
	cases := []struct{ vendor, stream, plain string }{
		{"anthropic", ms, mp},
		{"openai", `data: {"model":"fixture","choices":[{"index":1,"delta":{"content":"IGNORE"}},{"index":0,"delta":{"content":"Hé"},"finish_reason":null}]}

data: {"model":"fixture","choices":[{"index":0,"delta":{"content":"llo.","refusal":"opaque refusal"},"finish_reason":"stop"}]}

data: {"model":"fixture","choices":[],"usage":{"prompt_tokens":10,"completion_tokens":2}}

data: [DONE]

`, `{"model":"fixture","choices":[{"index":0,"message":{"content":"Héllo.","refusal":"opaque refusal"},"finish_reason":"stop"}],"usage":{"prompt_tokens":10,"completion_tokens":2}}`},
		{"gemini", `data: {"modelVersion":"fixture","candidates":[{"content":{"parts":[{"text":"Hé"},{"text":"llo."},{"text":"Check","thought":true}]}}]}

data: {"modelVersion":"fixture","candidates":[{"index":0,"content":{"parts":[{"text":"ing.","thought":true},{"text":"","thoughtSignature":"signed-value"}]},"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":10,"candidatesTokenCount":2}}

`, `{"modelVersion":"fixture","candidates":[{"content":{"parts":[{"text":"Hé"},{"text":"llo."},{"text":"Checking.","thought":true},{"text":"","thoughtSignature":"signed-value"}]},"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":10,"candidatesTokenCount":2}}`},
	}
	for _, c := range cases {
		t.Run(c.vendor, func(t *testing.T) {
			op, a := testOperation()
			config := common.Config{Vendor: c.vendor, Model: "requested"}
			streamed, err := parseStream(context.Background(), op, config, &splitReader{[]byte(c.stream), 0})
			if err != nil {
				t.Fatal(err)
			}
			plain, err := Parse(op.Engine(), config, []byte(c.plain), 0)
			if err != nil {
				t.Fatal(err)
			}
			x, _ := json.Marshal(streamed.Response)
			y, _ := json.Marshal(plain.Response)
			var xv, yv any
			json.Unmarshal(x, &xv)
			json.Unmarshal(y, &yv)
			if !reflect.DeepEqual(xv, yv) {
				t.Fatalf("stream %s\nplain %s", x, y)
			}
			if len(streamed.PartIDs) != len(plain.Response.Parts) {
				t.Fatal("part mapping mismatch")
			}
			text := ""
			thinking := ""
			for _, f := range a.fragments {
				if f.Channel == "text" {
					text += f.Text
				}
				if f.Channel == "thinking" {
					thinking += f.Text
				}
			}
			if text != "Héllo." {
				t.Fatal(text)
			}
			if c.vendor != "openai" && thinking != "Checking." {
				t.Fatal(thinking)
			}
		})
	}
}
func TestMessagesLifecycleFailures(t *testing.T) {
	stream, _ := messagesFixture()
	bad := []string{strings.Replace(stream, `"message_stop"`, `"ping"`, 1), strings.Replace(stream, `"index":0`, `"index":1`, 1), strings.Replace(stream, `"text_delta"`, `"unknown_delta"`, 1), strings.Replace(stream, `"content_block_stop","index":0`, `"content_block_delta","index":0`, 1), strings.TrimSuffix(stream, "\n"), strings.Replace(stream, `"type":"message_stop"`, `"type":"error","error":{"message":"PRIVATE"}`, 1)}
	for i, body := range bad {
		op, _ := testOperation()
		if _, err := parseStream(context.Background(), op, common.Config{Vendor: "anthropic", Model: "x"}, strings.NewReader(body)); err == nil {
			t.Fatalf("case %d accepted", i)
		} else if strings.Contains(err.Error(), "PRIVATE") {
			t.Fatal("unsafe diagnostic")
		}
	}
}
func TestPendingStoreBoundAndCancellation(t *testing.T) {
	op, _ := testOperation()
	o := op.(*operation)
	o.noticed = true
	ctx, cancel := context.WithCancel(context.Background())
	if err := o.Emit(ctx, common.Fragment{PartID: 1, Channel: "text", Text: strings.Repeat("é", pendingLimit/2)}); err != nil {
		t.Fatal(err)
	}
	result := make(chan error, 1)
	go func() { result <- o.Emit(ctx, common.Fragment{PartID: 1, Channel: "text", Text: "x"}) }()
	cancel()
	if err := <-result; err != context.Canceled {
		t.Fatal(err)
	}
	drained, more := o.Drain()
	n := 0
	for _, f := range drained {
		n += len(f.Text)
	}
	if n != drainLimit || !more || o.bytes != pendingLimit-drainLimit {
		t.Fatal(n, more, o.bytes)
	}
	o.Discard()
	if len(o.pending) != 0 || o.bytes != 0 {
		t.Fatal("discard retained fragments")
	}
}
func TestStreamMissingUsageAndTerminal(t *testing.T) {
	for _, body := range []string{
		"data: {\"choices\":[{\"index\":0,\"delta\":{\"content\":\"x\"},\"finish_reason\":\"stop\"}]}\n\ndata: [DONE]\n\n",
		"data: {\"choices\":[{\"index\":0,\"delta\":{\"content\":\"x\"},\"finish_reason\":\"stop\"}],\"usage\":{\"prompt_tokens\":1,\"completion_tokens\":1}}\n\n",
	} {
		op, _ := testOperation()
		if _, err := parseStream(context.Background(), op, common.Config{Vendor: "openai", Model: "x"}, strings.NewReader(body)); err == nil {
			t.Fatal("accepted incomplete stream")
		}
	}
}
func TestGeminiUnknownTextCombination(t *testing.T) {
	for _, middle := range []string{`{"text":"hidden","future_payload":{"tag":"retain-me"}}`, `{"text":"hidden","thought":true,"future_payload":1}`} {
		op, a := testOperation()
		config := common.Config{Vendor: "gemini", Model: "fixture"}
		body := `{"modelVersion":"fixture","candidates":[{"content":{"parts":[{"text":"A"},` + middle + `,{"text":"B","thought":false}]},"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":1,"candidatesTokenCount":1}}`
		plain, err := Parse(op.Engine(), config, []byte(body), 0)
		if err != nil {
			t.Fatal(err)
		}
		streamed, err := parseStream(context.Background(), op, config, strings.NewReader("data: "+body+"\n\n"))
		if err != nil {
			t.Fatal(err)
		}
		x, _ := json.Marshal(plain.Response)
		y, _ := json.Marshal(streamed.Response)
		var xv, yv any
		json.Unmarshal(x, &xv)
		json.Unmarshal(y, &yv)
		if !reflect.DeepEqual(xv, yv) {
			t.Fatalf("plain %s stream %s", x, y)
		}
		parts := plain.Response.Parts
		if len(parts) != 3 || parts[1].Type != "opaque" || string(parts[1].Data) != middle || TextAnswer(op.Engine(), parts) != "AB" {
			t.Fatalf("%#v", parts)
		}
		for _, f := range a.fragments {
			if f.PartID == streamed.PartIDs[1] {
				t.Fatal("opaque combination leaked delta")
			}
		}
		c := common.Context{Entries: []common.Entry{{Actor: "agent", Parts: parts}}}
		wire, err := Render(op.Engine(), c, config)
		if err != nil || !strings.Contains(string(wire), "future_payload") {
			t.Fatal(string(wire), err)
		}
		config.Model = "foreign"
		wire, err = Render(op.Engine(), c, config)
		if err != nil || strings.Contains(string(wire), "future_payload") {
			t.Fatal(string(wire), err)
		}
		for _, bad := range []string{strings.Replace(body, `{"text":"A"},`, "", 1), strings.Replace(body, `"text":"hidden"`, `"text":42`, 1)} {
			if !strings.Contains(bad, `{"text":"A"}`) {
				bad = strings.Replace(bad, `,{"text":"B","thought":false}`, "", 1)
			}
			if _, err = Parse(op.Engine(), config, []byte(bad), 0); err == nil {
				t.Fatal("accepted malformed or opaque-only")
			}
		}
	}
}
func TestInterleavedCallsKeepPartIdentity(t *testing.T) {
	for _, vendor := range []string{"anthropic", "openai", "gemini"} {
		op, a := testOperation()
		var frames []any
		if vendor == "anthropic" {
			frames = append(frames, map[string]any{"type": "message_start", "message": map[string]any{"usage": map[string]int{"input_tokens": 10, "output_tokens": 0}}})
		}
		names := []string{"read_file", "list_directory"}
		ids := []string{"read-a", "list-b"}
		prefix := []string{`{"path":`, `{"path":"`}
		suffix := []string{`"notes.txt"}`, `."}`}
		for i := 0; i < 2; i++ {
			if vendor == "anthropic" {
				frames = append(frames, map[string]any{"type": "content_block_start", "index": i, "content_block": map[string]any{"type": "tool_use", "id": ids[i], "name": names[i], "input": map[string]any{}}})
			}
		}
		for half := 0; half < 2; half++ {
			for i := 0; i < 2; i++ {
				fragment := prefix[i]
				if half == 1 {
					fragment = suffix[i]
				}
				switch vendor {
				case "anthropic":
					frames = append(frames, map[string]any{"type": "content_block_delta", "index": i, "delta": map[string]any{"type": "input_json_delta", "partial_json": fragment}})
				case "openai":
					fn := map[string]any{"arguments": fragment}
					call := map[string]any{"index": i, "function": fn}
					if half == 0 {
						call["id"] = ids[i]
						fn["name"] = names[i]
					}
					frames = append(frames, map[string]any{"choices": []any{map[string]any{"index": 0, "delta": map[string]any{"tool_calls": []any{call}}}}})
				}
			}
		}
		switch vendor {
		case "anthropic":
			for i := 0; i < 2; i++ {
				frames = append(frames, map[string]any{"type": "content_block_stop", "index": i})
			}
			frames = append(frames, map[string]any{"type": "message_delta", "delta": map[string]string{"stop_reason": "tool_use"}, "usage": map[string]int{"output_tokens": 2}}, map[string]string{"type": "message_stop"})
		case "openai":
			frames = append(frames, map[string]any{"choices": []any{map[string]any{"index": 0, "delta": map[string]any{}, "finish_reason": "tool_calls"}}}, map[string]any{"choices": []any{}, "usage": map[string]int{"prompt_tokens": 10, "completion_tokens": 2}})
		case "gemini":
			parts := []any{}
			for i := 0; i < 2; i++ {
				parts = append(parts, map[string]any{"functionCall": map[string]any{"id": ids[i], "name": names[i], "args": json.RawMessage(prefix[i] + suffix[i])}})
			}
			frames = append(frames, map[string]any{"candidates": []any{map[string]any{"content": map[string]any{"parts": parts}, "finishReason": "STOP"}}, "usageMetadata": map[string]int{"promptTokenCount": 10, "candidatesTokenCount": 2}})
		}
		var body strings.Builder
		for _, frame := range frames {
			raw, _ := json.Marshal(frame)
			body.WriteString("data: " + string(raw) + "\n\n")
		}
		if vendor == "openai" {
			body.WriteString("data: [DONE]\n\n")
		}
		parsed, err := parseStream(context.Background(), op, common.Config{Vendor: vendor, Model: "fixture"}, strings.NewReader(body.String()))
		if err != nil {
			t.Fatal(vendor, err)
		}
		args := map[int]string{}
		for _, f := range a.fragments {
			if f.Channel == "tool_args" {
				args[f.PartID] += f.Text
			}
		}
		for i, p := range parsed.Response.Parts {
			if p.Name != names[i] || p.CallID != ids[i] || string(p.Args) != prefix[i]+suffix[i] || args[parsed.PartIDs[i]] != prefix[i]+suffix[i] {
				t.Fatal(vendor, p, args)
			}
		}
	}
}
