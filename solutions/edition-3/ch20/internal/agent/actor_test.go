package agent

import (
	"context"
	"errors"
	"io"
	"log"
	"sync"
	"testing"
	"time"

	"ensemble/internal/common"
)

// Only contract fixtures are used here: the real actor/mailbox controls order,
// while the transport deliberately stays pending until its turn is cancelled.
type seamParent struct {
	common.Ensemble
	mu           sync.Mutex
	observations []common.Observation
}

func (p *seamParent) Trace(common.TraceRecord) {}
func (p *seamParent) Logger() *log.Logger      { return log.New(io.Discard, "", 0) }
func (p *seamParent) Publish(_ common.Agent, o common.Observation) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.observations = append(p.observations, o)
}

type seamHistory struct {
	common.History
	parent common.Agent
	mu     sync.Mutex
	events []common.Event
}

func (h *seamHistory) Parent() common.Agent { return h.parent }
func (h *seamHistory) Append(e common.Event) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	e.Seq = common.Seq(len(h.events) + 1)
	h.events = append(h.events, e)
	return nil
}
func (h *seamHistory) LastSeq() common.Seq {
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.events) == 0 {
		return 0
	}
	return h.events[len(h.events)-1].Seq
}
func (h *seamHistory) Events() []common.Event {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]common.Event(nil), h.events...)
}
func (h *seamHistory) Context() common.Context { return common.Context{Turn: common.Idle} }

type seamEngine struct {
	common.Engine
	parent  common.Agent
	started chan struct{}
}

func (e *seamEngine) Parent() common.Agent                  { return e.parent }
func (e *seamEngine) Render(common.Context) ([]byte, error) { return []byte(`{}`), nil }
func (e *seamEngine) ExchangeRequest(ctx context.Context, _ []byte, _ ...common.ParseCallbacks) ([]common.Event, error) {
	e.started <- struct{}{}
	<-ctx.Done()
	return nil, ctx.Err()
}

type seamJobs struct {
	common.Jobs
	parent common.Agent
}

func (j *seamJobs) Parent() common.Agent { return j.parent }
func (j *seamJobs) Shutdown() error      { return nil }

type seamTools struct {
	common.ToolRegistry
	parent common.Agent
}

func (r *seamTools) Parent() common.Agent { return r.parent }

func TestQueuedStaleResponseIsNotPublishedInNewTurn(t *testing.T) {
	parent := &seamParent{}
	a, err := New(parent, common.Config{Model: "offline", APIKey: "dummy", BaseURL: "http://127.0.0.1:1", BaseDir: t.TempDir(), DataDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	// This handcrafted actor-ordering seam excludes persistence; its fake
	// Engine/History and absent Skills do not model archive ownership. Public
	// root composition and real recovery are exercised by integration tests.
	a.config.SavePath = ""
	history := &seamHistory{parent: a}
	engine := &seamEngine{parent: a, started: make(chan struct{}, 4)}
	a.SetHistory(history)
	a.SetEngine(engine)
	a.SetTools(&seamTools{parent: a})
	a.SetJobs(&seamJobs{parent: a})
	a.Start()
	defer a.Shutdown()
	old := a.Submit(context.Background(), "old request")
	select {
	case <-engine.started:
	case <-time.After(3 * time.Second):
		t.Fatal("initial request not frozen")
	}
	next := &request{ctx: context.Background(), text: "new request", reply: make(chan common.Result, 1), done: make(chan struct{})}
	oldParts := common.PartList{common.Text("old response must not publish"), {Type: "tool_call", CallID: "old-call", Name: "old-tool", Args: []byte(`{}`)}}
	// This is the worker-report seam, not a timing guess at HTTP socket delivery.
	a.box.mu.Lock()
	a.box.queue = append(a.box.queue, common.Interrupt{}, next, response{turn: 1, events: []common.Event{{Type: common.ResponseEnded, Response: &common.ResponseData{Parts: oldParts, Usage: common.Usage{Input: 7, Output: 3}}}}}, common.JobReport{Job: common.JobData{TurnID: 1, CallID: "ongoing-call", Name: "held", Handle: 9, Status: common.JobDone}, Text: "valid ongoing job result"})
	select {
	case a.box.wake <- struct{}{}:
	default:
	}
	a.box.mu.Unlock()
	select {
	case r := <-old:
		if !errors.Is(r.Err, common.ErrInterrupted) {
			t.Fatalf("old private completion: %+v", r)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("old turn not interrupted")
	}
	select {
	case <-engine.started:
	case <-time.After(3 * time.Second):
		t.Fatal("new turn did not start")
	}
	if err := a.box.post(response{turn: 2, events: []common.Event{{Type: common.ResponseEnded, Response: &common.ResponseData{Parts: common.PartList{common.Text("new current answer")}}}}}, false); err != nil {
		t.Fatal("new response not queued")
	}
	select {
	case r := <-next.reply:
		if r.Err != nil || r.Text != "new current answer" || r.TurnID != 2 {
			t.Fatalf("new completion contaminated: %+v", r)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("new turn not completed")
	}
	parent.mu.Lock()
	observations := append([]common.Observation(nil), parent.observations...)
	parent.mu.Unlock()

	for _, o := range observations {
		switch o := o.(type) {
		case common.PartDelta:
			if o.Chunk == "old response must not publish" {
				t.Fatal("stale HTTP delta escaped")
			}
		case common.PartFinal:
			if o.Part.Text == "old response must not publish" || o.Part.CallID == "old-call" {
				t.Fatalf("stale HTTP part escaped: %+v", o)
			}

		}
	}
	retained, validJob := false, false
	for _, e := range history.Events() {
		if e.Type == common.JobReported && e.TurnID == 1 && e.Job != nil && e.Job.CallID == "ongoing-call" {
			validJob = true
		}
		if e.Type == common.ResponseEnded && e.TurnID == 1 {
			retained = e.Stale && e.Response.Usage.Input == 7 && e.Response.Usage.Output == 3 && len(e.Response.Parts) == 2
		}
	}
	if !validJob {
		t.Fatal("valid ongoing-job fact was fenced with stale HTTP")
	}
	if !retained {
		t.Fatal("real stale response/usage/origin was not retained")
	}
}
