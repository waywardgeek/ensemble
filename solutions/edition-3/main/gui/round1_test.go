package gui

import (
	"bytes"
	"context"
	"encoding/json"
	"ensemble"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func streamAgent(t *testing.T, first string) (*ensemble.Ensemble, ensemble.Agent, chan struct{}) {
	t.Helper()
	release := make(chan struct{})
	peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, first)
		w.(http.Flusher).Flush()
		select {
		case <-release:
		case <-r.Context().Done():
			return
		}
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{},\"finish_reason\":\"stop\"}]}\n\ndata: [DONE]\n\n")
	}))
	app := ensemble.New(nil)
	a, err := app.NewAgent(ensemble.Config{Vendor: ensemble.OpenAI, Model: "offline", APIKey: "fake", BaseURL: peer.URL, StreamCapability: ensemble.StreamSSE, BaseDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { app.Shutdown(); peer.Close() })
	return app, a, release
}
func replayData(t *testing.T, c *client) struct{ Events, Partials []Wire } {
	t.Helper()
	select {
	case raw := <-c.queue:
		var w struct {
			Data struct{ Events, Partials []Wire }
		}
		if err := json.Unmarshal(raw, &w); err != nil {
			t.Fatal(err)
		}
		return w.Data
	case <-time.After(time.Second):
		t.Fatal("no replay")
		return struct{ Events, Partials []Wire }{}
	}
}
func waitDelta(t *testing.T, q ensemble.ObservationQueue, n int) {
	t.Helper()
	deadline := time.After(3 * time.Second)
	for n > 0 {
		select {
		case <-q.Wake():
			for _, o := range q.Drain() {
				if _, ok := o.(ensemble.PartDelta); ok {
					n--
				}
			}
		case <-deadline:
			t.Fatal("missing partial")
		}
	}
}
func TestCh7R1CompleteActiveReplay(t *testing.T) {
	text := strings.Repeat("full-content ", 12000)
	data, _ := json.Marshal(map[string]any{"model": "actual", "choices": []any{map[string]any{"delta": map[string]any{"content": text}}}})
	app, a, release := streamAgent(t, "data: "+string(data)+"\n\n")
	hub := NewHub(app)
	defer hub.Close()
	q := a.Observe()
	defer q.Close()
	reply := a.Submit(context.Background(), "stream")
	waitDelta(t, q, 1)
	c := hub.connect()
	defer hub.disconnect(c)
	snapshot := replayData(t, c)
	if len(snapshot.Partials) != 1 {
		t.Fatalf("active partials=%d, want complete single partial", len(snapshot.Partials))
	}
	if snapshot.Partials[0].Chunk != text {
		t.Fatalf("partial bytes=%d want%d", len(snapshot.Partials[0].Chunk), len(text))
	}
	close(release)
	select {
	case <-reply:
	case <-time.After(3 * time.Second):
		t.Fatal("turn did not complete")
	}
}

func eventually(t *testing.T, fn func() bool) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for !fn() {
		if time.Now().After(deadline) {
			t.Fatal("condition not reached")
		}
		time.Sleep(time.Millisecond)
	}
}
func decodeCard(t *testing.T, w Wire) Card {
	t.Helper()
	if w.Type == "part_delta" {
		return Card{Kind: w.Kind, Text: w.Chunk, Name: w.Name, Partial: true}
	}
	if w.Part == nil {
		t.Fatalf("missing public part on %s", w.Type)
	}
	return *w.Part
}
func drain(c *client) []Wire {
	var out []Wire
	for {
		select {
		case raw := <-c.queue:
			var w Wire
			_ = json.Unmarshal(raw, &w)
			out = append(out, w)
		default:
			return out
		}
	}
}
func idleAgent(t *testing.T, speaking bool) (*ensemble.Ensemble, ensemble.Agent) {
	t.Helper()
	app := ensemble.New(nil)
	a, err := app.NewAgent(ensemble.Config{Model: "offline", BaseDir: t.TempDir(), Speaking: speaking})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = app.Shutdown() })
	return app, a
}
func TestCh7R1ManyPartOverflowRecovery(t *testing.T) {
	calls := make([]any, 300)
	for i := range calls {
		calls[i] = map[string]any{"index": i, "id": fmt.Sprint("call-", i), "function": map[string]any{"name": "read_file", "arguments": `{"path":"not-executed"}`}}
	}
	raw, _ := json.Marshal(map[string]any{"choices": []any{map[string]any{"delta": map[string]any{"tool_calls": calls}}}})
	app, a, release := streamAgent(t, "data: "+string(raw)+"\n\n")
	h := NewHub(app)
	defer h.Close()
	h.mu.Lock()
	result := a.Submit(context.Background(), "many partials")
	// Holding the real view lock must not back up the actor handoff.
	deadline := time.Now().Add(3 * time.Second)
	for len(a.View().Partials) != 300 {
		if time.Now().After(deadline) {
			h.mu.Unlock()
			t.Fatal("blocked or missing parts")
		}
		time.Sleep(time.Millisecond)
	}
	h.mu.Unlock()
	c := h.connect()
	defer h.disconnect(c)
	if got := len(replayData(t, c).Partials); got != 300 {
		t.Fatalf("parts=%d", got)
	}
	_ = a.Post(ensemble.Interrupt{})
	select {
	case <-result:
	case <-time.After(time.Second):
		t.Fatal("interrupt blocked")
	}
	close(release)
}
func TestCh7R1ClientOverflowRecoversAllFacts(t *testing.T) {
	app, a := idleAgent(t, false)
	h := NewHub(app)
	defer h.Close()
	slow := h.connect()
	_ = replayData(t, slow)
	for i := 0; i < 200; i++ {
		_ = a.Post(ensemble.Ephemeral{Text: fmt.Sprint("fact-", i)})
	}
	eventually(t, func() bool { return a.View().Through >= 200 })
	h.mu.Lock()
	h.sync(h.view(a), a.View())
	h.mu.Unlock()
	select {
	case <-slow.done:
	default:
		t.Fatal("overflowing view did not disconnect")
	}
	c := h.connect()
	defer h.disconnect(c)
	snapshot := replayData(t, c)
	count := 0
	for _, w := range snapshot.Events {
		if w.Type == "part_final" && strings.HasPrefix(decodeCard(t, w).Text, "fact-") {
			count++
		}
	}
	if count != 200 {
		t.Fatalf("recovered=%d", count)
	}
}
func TestCh7R1CoherentFinalHandoff(t *testing.T) {
	app, a, release := streamAgent(t, "data: {\"choices\":[{\"delta\":{\"content\":\"one final\"}}]}\n\n")
	h := NewHub(app)
	defer h.Close()
	old := h.connect()
	_ = replayData(t, old)
	q := a.Observe()
	defer q.Close()
	result := a.Submit(context.Background(), "handoff")
	waitDelta(t, q, 1)
	h.mu.Lock()
	checkpoint := a.View()
	h.sync(h.view(a), checkpoint)
	close(release)
	select {
	case <-result:
	case <-time.After(time.Second):
		h.mu.Unlock()
		t.Fatal("held view lock blocked final")
	}
	// The final now exists in History, but this exact captured replay frontier
	// precedes it. Both existing and new clients must get that final once live.
	c := h.connectAt(map[string]ensemble.ViewSnapshot{a.ID(): checkpoint})
	h.mu.Unlock()
	defer h.disconnect(c)
	defer h.disconnect(old)
	snapshot := replayData(t, c)
	if len(snapshot.Partials) != 1 {
		t.Fatal("captured active part missing")
	}
	for _, w := range snapshot.Events {
		if w.Type == "part_final" && decodeCard(t, w).Text == "one final" {
			t.Fatal("new final leaked into older replay frontier")
		}
	}
	h.mu.Lock()
	h.sync(h.view(a), a.View())
	h.mu.Unlock()
	for _, view := range []*client{old, c} {
		n := 0
		for _, w := range drain(view) {
			if w.Type == "part_final" && decodeCard(t, w).Text == "one final" {
				n++
				if w.Replay {
					t.Fatal("live final marked replay")
				}
			}
		}
		if n != 1 {
			t.Fatalf("live final count=%d", n)
		}
	}
}
func TestCh7R1LateJobProjection(t *testing.T) {
	for _, kill := range []bool{false, true} {
		t.Run(fmt.Sprint("killed=", kill), func(t *testing.T) {
			app, a := idleAgent(t, false)
			h := NewHub(app)
			defer h.Close()
			c := h.connect()
			defer h.disconnect(c)
			_ = replayData(t, c)
			release := make(chan struct{})
			defer close(release)
			started := make(chan struct{})
			err := a.Tools().Register(ensemble.ToolDefinition{Name: "held", Description: "Controlled real job", Parameters: json.RawMessage(`{"type":"object"}`)}, func(ctx context.Context, call ensemble.ToolCall, _ json.RawMessage) error {
				_, _ = call.Write([]byte("early output"))
				close(started)
				select {
				case <-release:
				case <-ctx.Done():
					return ctx.Err()
				}
				_, _ = call.Write([]byte(" final output"))
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			invoke := func(name, args string) {
				t.Helper()
				if err := a.Tools().Invoke(context.Background(), ensemble.Part{Type: "tool_call", CallID: "call-" + name, Name: name, Args: json.RawMessage(args)}, ""); err != nil {
					t.Fatal(err)
				}
			}
			invoke("tool_limits", `{"ai_callback_delay":0}`)
			invoke("held", `{}`)
			select {
			case <-started:
			case <-time.After(time.Second):
				t.Fatal("job not started")
			}
			var pending ensemble.JobData
			var callPart uint64
			for _, e := range a.Log().EventsRange(0, a.View().Through) {
				if e.Tool != nil && e.Tool.Name == "held" {
					if e.Type == ensemble.ToolCalled {
						callPart = e.Tool.DisplayID
					}
					if e.Type == ensemble.ToolReturned && e.Tool.Job != nil {
						pending = *e.Tool.Job
					}
				}
			}
			if pending.Handle == 0 || pending.Status != ensemble.JobRunning {
				t.Fatalf("missing pending observation: %+v", pending)
			}
			if kill {
				job, _ := app.LookupJob(pending.Handle)
				if err := job.Kill("kill_job"); err != nil {
					t.Fatal(err)
				}
			} else {
				release <- struct{}{}
			}
			var terminal ensemble.JobData
			eventually(t, func() bool {
				for _, e := range a.Log().EventsRange(0, a.View().Through) {
					if e.Job != nil && e.Job.Handle == pending.Handle {
						terminal = *e.Job
						return true
					}
				}
				return false
			})
			if terminal.CallPartID != callPart {
				t.Fatalf("job correlation=%d call=%d", terminal.CallPartID, callPart)
			}
			h.mu.Lock()
			h.sync(h.view(a), a.View())
			h.mu.Unlock()
			found := false
			for _, w := range drain(c) {
				if w.Type == "job_update" && w.PartID == callPart {
					card := decodeCard(t, w)
					found = card.Job != nil && card.Job.Status == terminal.Status
				}
			}
			if !found {
				t.Fatal("late terminal job absent live")
			}
			replay := h.connect()
			defer h.disconnect(replay)
			found = false
			for _, w := range replayData(t, replay).Events {
				if w.Type == "job_update" && w.PartID == callPart {
					card := decodeCard(t, w)
					found = card.Job != nil && card.Job.Status == terminal.Status
					if kill && (card.Job.Status != ensemble.JobKilledStatus || card.Job.Reason != "kill_job") {
						t.Fatalf("killed metadata changed: %+v", card.Job)
					}
				}
			}
			if !found {
				t.Fatal("late job absent replay")
			}
			if pending.Status != ensemble.JobRunning {
				t.Fatal("pending observation mutated")
			}
		})
	}
}

type safeBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *safeBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}
func (b *safeBuffer) String() string { b.mu.Lock(); defer b.mu.Unlock(); return b.buf.String() }
func TestCh7R1TerminalStreamAndEOF(t *testing.T) {
	app, a, release := streamAgent(t, "data: {\"choices\":[{\"delta\":{\"content\":\"EARLY-UNIQUE\"}}]}\n\n")
	s, err := NewServer(app, "web/gui")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	s.DefaultAgent = a.ID()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var out safeBuffer
	s.Terminal(ctx, strings.NewReader("stream\n"), &out)
	eventually(t, func() bool { return strings.Contains(out.String(), "EARLY-UNIQUE") })
	close(release)
	eventually(t, func() bool {
		for _, e := range a.Log().EventsRange(0, a.View().Through) {
			if e.Type == ensemble.ResponseEnded {
				return true
			}
		}
		return false
	})
	s.Hub.mu.Lock()
	s.Hub.sync(s.Hub.view(a), a.View())
	s.Hub.mu.Unlock()
	eventually(t, func() bool { return strings.Contains(out.String(), "idle") })
	if n := strings.Count(out.String(), "EARLY-UNIQUE"); n != 1 {
		t.Fatalf("stream/final printed %d times: %s", n, out.String())
	}
	web := httptest.NewServer(s.Handler(s.Assets))
	defer web.Close()
	resp, err := http.Get(web.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatal("EOF killed HTTP host")
	}
}
func TestCh7R1TerminalUnboundedInput(t *testing.T) {
	app := ensemble.New(nil)
	defer app.Shutdown()
	a, err := app.NewAgent(ensemble.Config{Model: "offline", PauseOnStart: true, BaseDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewServer(app, "web/gui")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	s.DefaultAgent = a.ID()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	input := strings.Repeat("long input ", 110000)
	want := strings.TrimSpace(input)
	if len(want) <= 1<<20 {
		t.Fatal("fixture input must exceed one MiB")
	}
	baseline := a.Log().LastSeq()
	s.Terminal(ctx, strings.NewReader(input+"\n"), io.Discard)
	// Startup facts are not an input acknowledgement. The paused/offline Agent
	// need only admit this whole human line; no provider completion is required.
	eventually(t, func() bool {
		for _, event := range a.Log().EventsRange(baseline, a.View().Through) {
			if event.Type == ensemble.MessageReceived && event.Message != nil && event.Message.Actor == "human" && len(event.Message.Parts) == 1 && event.Message.Parts[0].Type == "text" && event.Message.Parts[0].Text == want {
				return true
			}
		}
		return false
	})
}
func TestCh7R1WebSocketControls(t *testing.T) {
	app, a := idleAgent(t, false)
	s, err := NewServer(app, "web/gui")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	s.DefaultAgent = a.ID()
	web := httptest.NewServer(s.Handler(s.Assets))
	defer web.Close()
	ws, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(web.URL, "http")+"/ws", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ws.Close()
	_ = ws.SetReadDeadline(time.Now().Add(3 * time.Second))
	var replay Wire
	if err := ws.ReadJSON(&replay); err != nil {
		t.Fatal(err)
	}
	if replay.Type != "replay" {
		t.Fatal("no replay")
	}
	for _, pause := range []bool{true, false} {
		kind := "unpause"
		if pause {
			kind = "pause"
		}
		if err := ws.WriteJSON(map[string]any{"type": kind, "agent_id": a.ID(), "data": map[string]any{}}); err != nil {
			t.Fatal(err)
		}
		for {
			var w Wire
			if err := ws.ReadJSON(&w); err != nil {
				t.Fatal(err)
			}
			if w.Type == "state_changed" {
				if w.State != nil && w.State.Paused == pause {
					break
				}
			}
		}
	}
}

func TestCh7R1SpeechRetainsAllAcceptedPhrases(t *testing.T) {
	app, a := idleAgent(t, false)
	for i := 0; i < 100; i++ {
		_ = a.Post(ensemble.Ephemeral{Text: fmt.Sprintf("phrase-%03d.", i)})
	}
	eventually(t, func() bool { return a.View().Through >= 100 })
	s, err := NewServer(app, "web/gui")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	c := s.Hub.connect()
	defer s.Hub.disconnect(c)
	_ = replayData(t, c)
	for _, e := range a.Log().EventsRange(0, a.View().Through) {
		if e.Message != nil {
			for _, p := range e.Message.Parts {
				if err := s.command(c, command{Type: "speak", AgentID: a.ID(), PartID: p.DisplayID}); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
	var heard strings.Builder
	tasks := 0
	for {
		s.Hub.mu.Lock()
		id := s.Hub.speech.active[a.ID()]
		activeCount := len(s.Hub.speech.tasks)
		pendingCount := len(s.Hub.speech.pending)
		s.Hub.mu.Unlock()
		if id == "" {
			break
		}
		if activeCount > 1 || pendingCount > 1 {
			t.Fatal("speech task queue not coalesced")
		}
		if err := s.command(c, command{Type: "claim_speech", AgentID: a.ID(), SpeechID: id}); err != nil {
			t.Fatal(err)
		}
		for _, w := range drain(c) {
			if w.Type == "speech_claimed" {
				raw, _ := json.Marshal(w.Data)
				var x struct{ Text string }
				_ = json.Unmarshal(raw, &x)
				heard.WriteString(x.Text)
				heard.WriteByte('\n')
			}
		}
		if err := s.command(c, command{Type: "speech_done", AgentID: a.ID(), SpeechID: id, Status: "done"}); err != nil {
			t.Fatal(err)
		}
		tasks++
	}
	if tasks != 2 {
		t.Fatalf("coalesced tasks=%d", tasks)
	}
	for i := 0; i < 100; i++ {
		phrase := fmt.Sprintf("phrase-%03d.", i)
		if n := strings.Count(heard.String(), phrase); n != 1 {
			t.Fatalf("%s count=%d", phrase, n)
		}
	}
}
func TestCh7R1SpeechCancelClaimAndOtherTabRelease(t *testing.T) {
	app, a := idleAgent(t, true)
	s, err := NewServer(app, "web/gui")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	owner := s.Hub.connect()
	other := s.Hub.connect()
	defer s.Hub.disconnect(owner)
	defer s.Hub.disconnect(other)
	_ = replayData(t, owner)
	_ = replayData(t, other)
	request := func() string {
		s.Hub.mu.Lock()
		defer s.Hub.mu.Unlock()
		s.Hub.speech.request(a.ID(), ensemble.SpeechRequested{Parts: speechParts("read all of this"), Reason: "on demand"})
		return s.Hub.speech.active[a.ID()]
	}
	stale := request()
	previous := a.View().ContinueEpoch
	if err := s.command(other, command{Type: "unpause", AgentID: a.ID()}); err != nil {
		t.Fatal(err)
	}
	eventually(t, func() bool { return a.View().ContinueEpoch > previous })
	if err := s.command(owner, command{Type: "claim_speech", AgentID: a.ID(), SpeechID: stale}); err == nil {
		t.Fatal("cancelled claim accepted")
	}
	for _, w := range drain(owner) {
		if w.Type == "speech_claimed" {
			t.Fatal("cancelled claim published late")
		}
	}
	id := request()
	if err := s.command(owner, command{Type: "claim_speech", AgentID: a.ID(), SpeechID: id}); err != nil {
		t.Fatal(err)
	}
	if err := s.command(owner, command{Type: "pause", AgentID: a.ID()}); err != nil {
		t.Fatal(err)
	}
	eventually(t, func() bool { return a.View().Paused })
	if err := s.command(other, command{Type: "speaking", AgentID: a.ID(), Enabled: false}); err != nil {
		t.Fatal(err)
	}
	eventually(t, func() bool { return !a.View().Speaking })
	s.Hub.mu.Lock()
	s.Hub.sync(s.Hub.view(a), a.View())
	s.Hub.mu.Unlock()
	release := false
	for _, w := range drain(owner) {
		if w.Type == "speech_cancelled" {
			raw, _ := json.Marshal(w.Data)
			var x struct {
				Owner   string
				Release bool
			}
			_ = json.Unmarshal(raw, &x)
			release = x.Release && x.Owner == owner.id
		}
	}
	if !release {
		t.Fatal("other-tab Speaking off did not release rightful owner's empty queue")
	}
	if err := s.command(owner, command{Type: "speech_done", AgentID: a.ID(), SpeechID: id, Status: "done"}); err != nil {
		t.Fatal(err)
	}
	s.Hub.mu.Lock()
	remaining := len(s.Hub.speech.tasks) + len(s.Hub.speech.pending)
	s.Hub.mu.Unlock()
	if remaining != 0 {
		t.Fatal("cancelled completion resurrected speech")
	}
}
