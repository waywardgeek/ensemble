package ensemble

import (
	"bytes"
	"context"
	"encoding/json"
	"example.com/ensemble/internal/common"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func actorAgent(t *testing.T, handler http.HandlerFunc, builtins ...string) (*Ensemble, *Agent) {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	app := New(nil)
	dir := t.TempDir()
	a, err := app.NewAgent(Config{APIKey: "fixture", Model: "fixture", BaseURL: server.URL, Workspace: dir, LogPath: filepath.Join(dir, "log"), Builtins: builtins})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { app.Close() })
	return app, a
}
func waitCompletion(t *testing.T, h RequestHandle) Completion {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	c, err := h.Wait(ctx)
	if err != nil {
		t.Fatal(err)
	}
	return c
}
func waitSignal(t *testing.T, ch <-chan struct{}) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(3 * time.Second):
		t.Fatal("barrier not reached")
	}
}
func replyText(w http.ResponseWriter, text string) {
	data, _ := json.Marshal(text)
	fmt.Fprintf(w, `{"content":[{"type":"text","text":%s}],"usage":{"input_tokens":2,"output_tokens":1}}`, data)
}
func TestActorFIFOHandlesHintsAndQueuedCancellation(t *testing.T) {
	started, release := make(chan struct{}), make(chan struct{})
	var mu sync.Mutex
	var bodies []string
	_, a := actorAgent(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		mu.Lock()
		bodies = append(bodies, string(body))
		n := len(bodies)
		mu.Unlock()
		if n == 1 {
			close(started)
			<-release
		}
		replyText(w, fmt.Sprintf("answer-%d", n))
	})
	first, _ := a.Submit("first")
	waitSignal(t, started)
	second, _ := a.Submit("second")
	canceled, _ := a.Submit("never-send")
	third, _ := a.Submit("third")
	if err := canceled.Cancel(); err != nil {
		t.Fatal(err)
	}
	if c := waitCompletion(t, canceled); c.Outcome != "canceled" {
		t.Fatal(c)
	}
	ack, err := a.Hint("HINT-ONE")
	if err != nil || ack.RequestID != first.ID() || ack.Sent {
		t.Fatal(ack, err)
	}
	config := a.Config()
	config.Model = "changed"
	if a.SetConfig(config) == nil {
		t.Fatal("active config changed")
	}
	if a.Ephemeral("busy") == nil {
		t.Fatal("active context changed")
	}
	close(release)
	c1 := waitCompletion(t, first)
	c2 := waitCompletion(t, second)
	c3 := waitCompletion(t, third)
	if c1.Text != "answer-1" || c1.PendingHints != 1 || c2.Text != "answer-2" || c2.PendingHints != 0 || c3.Text != "answer-3" {
		t.Fatal(c1, c2, c3)
	}
	if again := waitCompletion(t, first); again.Text != c1.Text {
		t.Fatal("completion consumed")
	}
	mu.Lock()
	defer mu.Unlock()
	if len(bodies) != 3 || strings.Contains(bodies[0], "HINT-ONE") || !strings.Contains(bodies[1], "HINT-ONE") || strings.Contains(bodies[2], "HINT-ONE") {
		t.Fatal(bodies)
	}
	for _, e := range a.Events() {
		if e.Turn != nil && e.Turn.RequestID == canceled.ID() {
			t.Fatal("queued cancellation entered history")
		}
	}
}
func TestActorInterruptBlockedHTTPAndLaterPrompt(t *testing.T) {
	started, release := make(chan struct{}), make(chan struct{})
	var mu sync.Mutex
	calls := 0
	_, a := actorAgent(t, func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		calls++
		n := calls
		mu.Unlock()
		if n == 1 {
			close(started)
			<-release
			replyText(w, "STALE")
			return
		}
		replyText(w, "later")
	})
	first, _ := a.Submit("abandoned")
	waitSignal(t, started)
	later, _ := a.Submit("later")
	ack, err := a.Interrupt()
	if err != nil || !ack.Interrupted || ack.RequestID != first.ID() {
		t.Fatal(ack, err)
	}
	if c := waitCompletion(t, first); c.Outcome != "interrupted" {
		t.Fatal(c)
	}
	close(release)
	if c := waitCompletion(t, later); c.Text != "later" || c.Outcome != "success" {
		t.Fatal(c)
	}
	if a.Usage() != (Usage{Input: 2, Output: 1}) {
		t.Fatal("discarded response counted", a.Usage())
	}
	for _, m := range a.History() {
		if strings.Contains(m.Content, "STALE") || strings.Contains(m.Content, "abandoned") {
			t.Fatal("stale history", a.History())
		}
	}
	if ack, err = a.Interrupt(); err != nil || ack.Interrupted {
		t.Fatal(ack, err)
	}
}
func TestActorInterruptReportPairsBatchAndRetainsLateJob(t *testing.T) {
	var mu sync.Mutex
	calls := 0
	app, a := actorAgent(t, func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		calls++
		n := calls
		mu.Unlock()
		if n == 1 {
			fmt.Fprint(w, `{"content":[{"type":"tool_use","id":"slow","name":"run_command","input":{"command":"printf ready; while [ ! -f release ]; do sleep 0.01; done; printf late-marker","ai_callback_delay":30}},{"type":"tool_use","id":"refused","name":"write_file","input":{"path":"must-not-exist","content":"bad"}}],"usage":{"input_tokens":1,"output_tokens":1}}`)
			return
		}
		replyText(w, "next-turn")
	}, "run_command", "write_file")
	dispatched, ended := make(chan struct{}), make(chan struct{})
	app.Subscribe(a.ID(), callbackObserver(func(o Observation) {
		if o.Kind == "tool_called" && o.Event.Tool.CallID == "slow" {
			close(dispatched)
		}
		if o.Kind == "job_ended" {
			close(ended)
		}
	}))
	h, _ := a.Submit("work")
	waitSignal(t, dispatched)
	ack, err := a.Hint("KEEP-HINT")
	if err != nil || ack.Sent {
		t.Fatal(ack, err)
	}
	if _, err = a.Interrupt(); err != nil {
		t.Fatal(err)
	}
	if c := waitCompletion(t, h); c.Outcome != "interrupted" || c.PendingHints != 1 {
		t.Fatal(c)
	}
	counts := map[string]int{}
	for _, e := range a.Events() {
		if e.Type == "tool_returned" {
			counts[e.Tool.CallID]++
			if e.Tool.CallID == "slow" && e.Tool.Job.Status != "running" {
				t.Fatal("interruption killed running job")
			}
			if e.Tool.CallID == "refused" && (!e.Tool.IsError || e.Tool.Job != nil) {
				t.Fatal("refusal invented effect")
			}
		}
	}
	if counts["slow"] != 1 || counts["refused"] != 1 {
		t.Fatal(counts)
	}
	if _, err = os.Stat(filepath.Join(a.Workspace(), "must-not-exist")); !os.IsNotExist(err) {
		t.Fatal("undispatched effect ran")
	}
	os.WriteFile(filepath.Join(a.Workspace(), "release"), []byte("go"), 0600)
	waitSignal(t, ended)
	next, _ := a.Submit("next")
	if c := waitCompletion(t, next); c.Outcome != "success" {
		t.Fatal(c)
	}
	terminal := 0
	for _, e := range a.Events() {
		if e.Type == "job_ended" {
			terminal++
		}
	}
	if terminal != 1 {
		t.Fatal("terminal duplicated", terminal)
	}
}
func TestActorCloseSettlesActiveQueuedAndRefusesAdmission(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	_, a := actorAgent(t, func(w http.ResponseWriter, r *http.Request) {
		close(started)
		select {
		case <-r.Context().Done():
		case <-release:
		}
	})
	t.Cleanup(func() { close(release) })
	h, _ := a.Submit("blocked")
	waitSignal(t, started)
	queued, _ := a.Submit("queued")
	if err := a.Close(); err != nil {
		t.Fatal(err)
	}
	if err := a.Close(); err != nil {
		t.Fatal(err)
	}
	for _, h := range []RequestHandle{h, queued} {
		if c := waitCompletion(t, h); c.Outcome != "stopped" {
			t.Fatal(c)
		}
	}
	if _, err := a.Submit("too-late"); err == nil || !strings.Contains(err.Error(), "stopped") {
		t.Fatal(err)
	}
}
func TestCompletionCollectionsAndCanceledWait(t *testing.T) {
	app, a := actorAgent(t, func(w http.ResponseWriter, r *http.Request) { replyText(w, "ok") })
	h1, _ := a.Submit("one")
	h2, _ := a.Submit("two")
	waitCompletion(t, h1)
	waitCompletion(t, h2)
	for n := 0; n < 2; n++ {
		collection := app.Collect([]RequestHandle{h2, h1})
		all, exhausted, err := collection.Wait(context.Background())
		if err != nil || !exhausted || len(all) != 2 || all[0].RequestID != h2.ID() || all[1].RequestID != h1.ID() {
			t.Fatal(all, exhausted, err)
		}
		all, exhausted, err = collection.Wait(context.Background())
		if err != nil || !exhausted || len(all) != 0 {
			t.Fatal(all, exhausted, err)
		}
	}
	if all, done, err := app.Collect(nil).Wait(context.Background()); err != nil || !done || len(all) != 0 {
		t.Fatal(all, done, err)
	}
	started, release := make(chan struct{}), make(chan struct{})
	app2, b := actorAgent(t, func(w http.ResponseWriter, r *http.Request) { close(started); <-release; replyText(w, "alive") })
	h, _ := b.Submit("wait")
	waitSignal(t, started)
	collection := app2.Collect([]RequestHandle{h})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := collection.Wait(ctx); err != context.Canceled {
		t.Fatal(err)
	}
	close(release)
	if all, done, err := collection.Wait(context.Background()); err != nil || !done || len(all) != 1 || all[0].Text != "alive" {
		t.Fatal(all, done, err)
	}
}
func TestObserverOverflowDoesNotConsumeCompletion(t *testing.T) {
	app, a := actorAgent(t, func(w http.ResponseWriter, r *http.Request) { replyText(w, "independent") })
	blocked, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	id, err := app.Subscribe(a.ID(), callbackObserver(func(o Observation) { once.Do(func() { close(blocked); <-release }) }))
	if err != nil {
		t.Fatal(err)
	}
	if err = a.Ephemeral("first"); err != nil {
		t.Fatal(err)
	}
	waitSignal(t, blocked)
	for i := 0; i < 300; i++ {
		if err = a.Ephemeral(fmt.Sprint(i)); err != nil {
			t.Fatal(err)
		}
	}
	if app.SubscriptionStatus(id) != "overflow" {
		t.Fatal("overflow not exposed")
	}
	h, _ := a.Submit("go")
	if c := waitCompletion(t, h); c.Text != "independent" {
		t.Fatal(c)
	}
	close(release)
}

func TestCollectionStaggeredReadinessAndIndependentWait(t *testing.T) {
	releases := []chan struct{}{make(chan struct{}), make(chan struct{})}
	started := make(chan struct{}, 2)
	app := New(nil)
	defer app.Close()
	handles := make([]RequestHandle, 2)
	servers := make([]*httptest.Server, 2)
	for i := range handles {
		i := i
		servers[i] = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			started <- struct{}{}
			<-releases[i]
			replyText(w, fmt.Sprint(i))
		}))
		defer servers[i].Close()
		a, err := app.NewAgent(Config{APIKey: "test", Model: "test", BaseURL: servers[i].URL, LogPath: filepath.Join(t.TempDir(), "log")})
		if err != nil {
			t.Fatal(err)
		}
		handles[i], _ = a.Submit("go")
	}
	<-started
	<-started
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := handles[0].Wait(ctx); err != context.Canceled {
		t.Fatal(err)
	}
	collection := app.Collect(handles)
	first := make(chan []Completion, 1)
	go func() { out, _, _ := collection.Wait(context.Background()); first <- out }()
	close(releases[0])
	out := <-first
	if len(out) != 1 || out[0].RequestID != handles[0].ID() || out[0].Text != "0" {
		t.Fatal(out)
	}
	close(releases[1])
	out, done, err := collection.Wait(context.Background())
	if err != nil || !done || len(out) != 1 || out[0].Text != "1" {
		t.Fatal(out, done, err)
	}
}
func TestRequestReplayHintAndTurnValidation(t *testing.T) {
	started, release := make(chan struct{}), make(chan struct{})
	var bodies [][]byte
	var mu sync.Mutex
	_, a := actorAgent(t, func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		mu.Lock()
		bodies = append(bodies, data)
		n := len(bodies)
		mu.Unlock()
		if n == 1 {
			close(started)
			<-release
		}
		replyText(w, "ok")
	})
	a.Ephemeral("EPHEMERAL")
	h, _ := a.Submit("first")
	waitSignal(t, started)
	a.Hint("HINT-EXACT")
	close(release)
	waitCompletion(t, h)
	h, _ = a.Submit("second")
	waitCompletion(t, h)
	requestIndex := 0
	for _, e := range a.Events() {
		if e.Request == nil {
			continue
		}
		body, err := a.ReconstructRequest(e.Seq)
		if err != nil {
			t.Fatal(err)
		}
		mu.Lock()
		want := bodies[requestIndex]
		mu.Unlock()
		if string(body) != string(want) {
			t.Fatalf("replay differs at request %d", requestIndex)
		}
		requestIndex++
	}
	if requestIndex != 2 {
		t.Fatal(requestIndex)
	}
	// Offline removal of the hint changes the replayed request bytes; originals are
	// retained and the real request never gets replaced by a regenerated receipt.
	events := a.Events()
	for i, e := range events {
		if e.Type == "hint_received" {
			events = append(events[:i], events[i+1:]...)
			break
		}
	}
	for i := range events {
		if events[i].Request != nil && len(events[i].Request.Hints) > 0 {
			events[i].Request.Hints = []uint64{}
		}
	}
	path := filepath.Join(t.TempDir(), "without-hint.log")
	f, _ := os.Create(path)
	fmt.Fprintln(f, `{"log_version":1}`)
	enc := json.NewEncoder(f)
	for _, e := range events {
		enc.Encode(e)
	}
	f.Close()
	offline, err := New(nil).Load(path, Config{Model: "fixture"})
	if err != nil {
		t.Fatal(err)
	}
	defer offline.Close()
	var seq uint64
	for _, e := range events {
		if e.Request != nil {
			seq = e.Seq
		}
	}
	body, err := offline.ReconstructRequest(seq)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(body), "HINT-EXACT") || string(body) == string(bodies[1]) {
		t.Fatal("removed-hint negative control failed")
	}
	// Explicit turn IDs cannot be reused, nor can a start/end invent success.
	for _, e := range []Event{{Type: "turn_started", Turn: &common.TurnEvent{RequestID: "r1"}}, {Type: "hint_received", Hint: &common.HintEvent{RequestID: "absent", Text: "no"}}} {
		if a.Append(e) == nil {
			t.Fatal("invalid explicit boundary accepted")
		}
	}
}

func TestProviderFailureQueueProgressAndClosePreservesAcceptedEffect(t *testing.T) {
	firstStarted, releaseFirst, continuation := make(chan struct{}), make(chan struct{}), make(chan struct{})
	releaseLast := make(chan struct{})
	var mu sync.Mutex
	calls := 0
	_, a := actorAgent(t, func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		calls++
		n := calls
		mu.Unlock()
		switch n {
		case 1:
			close(firstStarted)
			<-releaseFirst
			w.WriteHeader(503)
		case 2:
			fmt.Fprint(w, `{"content":[{"type":"tool_use","id":"accepted-effect","name":"write_file","input":{"path":"retained","content":"durable effect"}}],"usage":{"input_tokens":1,"output_tokens":1}}`)
		case 3:
			close(continuation)
			select {
			case <-r.Context().Done():
			case <-releaseLast:
			}
		default:
			t.Errorf("queued request incorrectly reached HTTP: %d", n)
		}
	}, "write_file")
	t.Cleanup(func() { close(releaseLast) })
	failed, _ := a.Submit("first failure")
	waitSignal(t, firstStarted)
	active, _ := a.Submit("second may activate before caller reads first failure")
	queued, _ := a.Submit("third remains queued")
	close(releaseFirst)
	waitSignal(t, continuation)
	if c := waitCompletion(t, failed); c.Outcome != "error" {
		t.Fatal(c)
	}
	if err := a.Close(); err != nil {
		t.Fatal(err)
	}
	for _, h := range []RequestHandle{active, queued} {
		if c := waitCompletion(t, h); c.Outcome != "stopped" {
			t.Fatal(c)
		}
	}
	data, err := os.ReadFile(filepath.Join(a.Workspace(), "retained"))
	if err != nil || string(data) != "durable effect" {
		t.Fatal("accepted effect rolled back", string(data), err)
	}
	for _, e := range a.Events() {
		if e.Turn != nil && e.Turn.RequestID == queued.ID() {
			t.Fatal("queued close invented a turn")
		}
	}
	if a.Usage() != (Usage{Input: 1, Output: 1}) {
		t.Fatal("accepted response usage lost", a.Usage())
	}
}

type announcingContext struct {
	context.Context
	once    sync.Once
	entered chan struct{}
}

func (c *announcingContext) Done() <-chan struct{} {
	c.once.Do(func() { close(c.entered) })
	return c.Context.Done()
}
func TestConcurrentCollectionWaitCanCancelWhileAnotherIsParked(t *testing.T) {
	started, release := make(chan struct{}), make(chan struct{})
	app, a := actorAgent(t, func(w http.ResponseWriter, r *http.Request) { close(started); <-release; replyText(w, "kept") })
	h, _ := a.Submit("park")
	waitSignal(t, started)
	collection := app.Collect([]RequestHandle{h})
	entered := make(chan struct{})
	firstDone := make(chan error, 1)
	go func() {
		out, done, err := collection.Wait(&announcingContext{Context: context.Background(), entered: entered})
		if err == nil && (len(out) != 1 || !done) {
			err = fmt.Errorf("first collection wait lost completion")
		}
		firstDone <- err
	}()
	waitSignal(t, entered)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	secondDone := make(chan error, 1)
	go func() { _, _, err := collection.Wait(ctx); secondDone <- err }()
	select {
	case err := <-secondDone:
		if err != context.Canceled {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		close(release)
		t.Fatal("canceled waiter blocked behind another parked waiter")
	}
	close(release)
	if err := <-firstDone; err != nil {
		t.Fatal(err)
	}
}

func TestHintFixtureRendersAfterResultsOnEverySurface(t *testing.T) {
	for _, vendor := range []string{"anthropic", "openai", "gemini"} {
		t.Run(vendor, func(t *testing.T) {
			app := New(nil)
			defer app.Close()
			a, err := app.NewAgent(Config{APIKey: "fixture", Model: "fixture", Vendor: vendor, LogPath: filepath.Join(t.TempDir(), "log")})
			if err != nil {
				t.Fatal(err)
			}
			from := Provenance{Vendor: "anthropic", Model: "fixture", Surface: "messages"}
			events := []Event{
				{Type: "turn_started", Turn: &TurnEvent{RequestID: "fixture-turn"}},
				{Type: "message_received", Message: &Entry{Actor: "human", Purpose: "dialogue", Parts: []Part{Text("Read notes")}}},
				{Type: "response_ended", Response: &Response{From: from, Parts: []Part{{Type: "tool_call", CallID: "read-1", Name: "read_file", Args: json.RawMessage(`{"path":"notes"}`), From: &from}}, Usage: &Usage{}}},
				{Type: "hint_received", Hint: &HintEvent{RequestID: "fixture-turn", Text: "HINT-AFTER-RESULT"}},
				{Type: "tool_called", Tool: &ToolEvent{CallID: "read-1", Name: "read_file", Args: json.RawMessage(`{"path":"notes"}`)}},
				{Type: "tool_returned", Tool: &ToolEvent{CallID: "read-1", Parts: []Part{Text("RESULT-CONTENT")}}},
			}
			for _, e := range events {
				if err = a.Append(e); err != nil {
					t.Fatal(err)
				}
			}
			config := Config{Model: "fixture", Vendor: vendor}
			first, err := a.Render(config)
			if err != nil {
				t.Fatal(err)
			}
			second, err := a.Render(config)
			if err != nil || string(first) != string(second) {
				t.Fatal("render changed context", err)
			}
			if strings.Index(string(first), "HINT-AFTER-RESULT") <= strings.Index(string(first), "RESULT-CONTENT") {
				t.Fatal("hint split required call/result order", string(first))
			}
			for _, hints := range [][]uint64{{4, 4}, {5}, {}} {
				if a.Append(Event{Type: "request_sent", Request: &RequestEvent{To: from, Ephemera: []uint64{}, Hints: hints}}) == nil {
					t.Fatal("invalid hint consumption accepted", hints)
				}
			}
			if err = a.Append(Event{Type: "request_sent", Request: &RequestEvent{To: from, Ephemera: []uint64{}, Hints: []uint64{4}}}); err != nil {
				t.Fatal(err)
			}
			if len(a.Snapshot().Hints) != 0 {
				t.Fatal("sent hint not consumed")
			}
		})
	}
}

func TestUnsupportedReferenceNamesModelAndMIMEWithoutLocator(t *testing.T) {
	for _, test := range []struct {
		vendor string
		kind   int
	}{{"openai", 2}, {"anthropic", 2}, {"gemini", 1}, {"gemini", 3}} {
		t.Run(fmt.Sprintf("%s-%d", test.vendor, test.kind), func(t *testing.T) {
			var diagnostics bytes.Buffer
			app := New(&diagnostics)
			defer app.Close()
			a, err := app.NewAgent(Config{APIKey: "fixture", Vendor: test.vendor, Model: "new-text-model", LogPath: filepath.Join(t.TempDir(), "log")})
			if err != nil {
				t.Fatal(err)
			}
			if _, err = a.Render(Config{Vendor: test.vendor, Model: "new-text-model"}); err != nil {
				t.Fatal("unknown text model refused", err)
			}
			locator := "private-location-credential-marker"
			err = a.Append(Event{Type: "message_received", Message: &Entry{Actor: "human", Purpose: "dialogue", Parts: []Part{{Type: "blob", MIME: "video/mp4", Ref: &Ref{Kind: test.kind, Locator: locator}}}}})
			if err != nil {
				t.Fatal(err)
			}
			_, err = a.Render(Config{Vendor: test.vendor, Model: "new-text-model"})
			if err == nil {
				t.Fatal("unsupported media silently accepted")
			}
			for _, marker := range []string{"new-text-model", "video/mp4", "mapping"} {
				if !strings.Contains(err.Error(), marker) {
					t.Fatal("missing diagnostic", marker, err)
				}
			}
			if strings.Contains(err.Error()+diagnostics.String(), locator) {
				t.Fatal("private locator leaked")
			}
		})
	}
}
