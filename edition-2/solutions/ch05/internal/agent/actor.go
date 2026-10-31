package agent

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"ensemble/internal/common"
)

// Private control messages share the same FIFO as public input. Request replies
// are buffered so an abandoned caller cannot hold up the conversation owner.
type request struct {
	ctx   context.Context
	text  string
	reply chan common.Result
}
type response struct {
	turn   *turn
	events []common.Event
	err    error
}
type report struct {
	turn *turn
	call common.Part
	data *common.ToolData
}
type record struct {
	event common.Event
	reply chan error
}
type flush struct{ reply chan error }
type stop struct{}
type workerDone struct {
	job common.Job
	err error
}
type turn struct {
	request        request
	cancel         context.CancelFunc
	calls          []common.Part
	current        *report
	batches, limit int
}

// All fields below are owned by run. Workers own external work only, and return
// immutable messages. Interrupt detaches a turn without detaching its workers.
type runtime struct {
	owner    *agent
	active   *turn
	pending  []request
	barriers []chan error
	workers  int
	stopping bool
	failure  error
	partID   uint64
}

// Construction installs children first; the first runtime operation starts the
// one owner. Subsequent callers cannot accidentally create another consumer.
func (a *agent) start() { a.once.Do(func() { go (&runtime{owner: a}).run() }) }

// Admission and final shutdown share a lock. A request cannot slip behind the
// last drain and leave its private reply waiting forever. Worker facts remain
// admissible during shutdown because joined workers may still need to record.
func (a *agent) post(message any) error {
	a.start()
	a.lifecycle.Lock()
	defer a.lifecycle.Unlock()
	if a.stopping {
		if _, workerFact := message.(record); !workerFact {
			return common.ErrStopped
		}
	}
	select {
	case <-a.done:
		return common.ErrStopped
	default:
	}
	a.box.Post(message)
	return nil
}

// Post accepts live input without waiting for execution or dropping its message.
func (a *agent) Post(message common.Inbound) error { return a.post(message) }

// Submit queues a separate human turn. Post is the deliberately different API
// for live input that may become a hint when a turn is already active.
func (a *agent) Submit(ctx context.Context, text string) <-chan common.Result {
	reply := make(chan common.Result, 1)
	if strings.TrimSpace(text) == "" {
		reply <- common.Result{Err: errors.New("user message is empty")}
		return reply
	}
	if err := a.post(request{ctx: ctx, text: text, reply: reply}); err != nil {
		reply <- common.Result{Err: err}
	}
	return reply
}

// Ask blocks on its own request reply or caller cancellation, never a shared idle
// event.
func (a *agent) Ask(ctx context.Context, text string) (string, error) {
	reply := a.Submit(ctx, text)
	select {
	case result := <-reply:
		return result.Text, result.Err
	case <-ctx.Done():
		return "", ctx.Err()
	case <-a.done:
		// A completed reply wins over the shutdown signal if both became ready.
		select {
		case result := <-reply:
			return result.Text, result.Err
		default:
			return "", common.ErrStopped
		}
	}
}

// A tool may report an effect without borrowing History. Its acknowledgment
// follows reduction, so the worker can safely send its completion afterward.
func (a *agent) Record(event common.Event) error {
	reply := make(chan error, 1)
	if err := a.post(record{event, reply}); err != nil {
		return err
	}
	select {
	case err := <-reply:
		return err
	case <-a.done:
		return common.ErrStopped
	}
}

// Ephemeral records an instruction for the next request without retaining dialogue.
func (a *agent) Ephemeral(text string) error {
	return a.Record(common.Event{Type: "message_received", Message: &common.MessageData{Actor: common.System, Parts: []common.Part{{Type: "text", Text: text}}}})
}

// Flush is an input barrier for clients reaching EOF. It waits for queued turns,
// not for every background job; Shutdown owns stopping and joining those jobs.
func (a *agent) Flush(ctx context.Context) error {
	reply := make(chan error, 1)
	if err := a.post(flush{reply}); err != nil {
		return err
	}
	select {
	case err := <-reply:
		return err
	case <-ctx.Done():
		return ctx.Err()
	case <-a.done:
		return common.ErrStopped
	}
}

// Stop admission once, then join the same owner on every call. The owner closes
// done only after managed processes and both kinds of tool worker have finished.
func (a *agent) Shutdown() error {
	a.start()
	a.lifecycle.Lock()
	if !a.stopping {
		a.stopping = true
		a.box.Post(stop{})
	}
	a.lifecycle.Unlock()
	<-a.done
	return a.stopErr
}

// Observe registers a progress consumer; its callback must not block the actor.
func (a *agent) Observe(observer common.Observer) {
	a.observersMu.Lock()
	defer a.observersMu.Unlock()
	a.observers = append(a.observers, observer)
}

// This is a consuming progress queue, not a broadcast subscription or replay.
// A caller needing a particular request result must retain Submit's reply.
func (a *agent) Wait(ctx context.Context, predicate func(common.Observation) bool) (common.Observation, error) {
	for {
		select {
		case observation := <-a.progress:
			if predicate(observation) {
				return observation, nil
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-a.done:
			return nil, common.ErrStopped
		}
	}
}
func (a *agent) notify(observation common.Observation) {
	a.parent.Observe(a, observation)
	select {
	case a.progress <- observation:
	default:
	}
	a.observersMu.Lock()
	observers := append([]common.Observer(nil), a.observers...)
	a.observersMu.Unlock()
	// Observe's contract forbids blocking. Slow consumers provide their own queue;
	// the core neither spawns one goroutine per event nor makes progress reliable.
	for _, observer := range observers {
		observer.Observe(observation)
	}
}

// Publication follows successful reduction. Observers see the state and part
// identity actually stored, including IDs assigned to calls by History.
func (r *runtime) append(event common.Event) error {
	h := r.owner.history
	before := h.Context().Turn
	if err := h.Append(event); err != nil {
		return err
	}
	after := h.Context()
	if before != after.Turn {
		r.owner.notify(common.StateChanged{From: before, To: after.Turn})
	}
	if event.Type == "hint_received" {
		r.owner.notify(common.PartDelta{Chunk: "hint:" + event.Message.Parts[0].Text})
	}
	// Finalized dialogue has its stored Seq; transient deltas never enter history.
	if len(after.Dialogue) > 0 && (event.Type == "response_ended" || event.Type == "tool_returned") {
		entry := after.Dialogue[len(after.Dialogue)-1]
		for _, part := range entry.Parts {
			r.partID++
			r.owner.notify(common.PartDelta{PartID: r.partID, Chunk: part.Text})
			r.owner.notify(common.PartFinal{Seq: entry.Seq, PartID: r.partID, Part: part})
		}
	}
	return nil
}

// Saving is an owner operation. In particular, do not rewrite again immediately
// after replying: the caller must be able to read the just-completed log.
// Crash-safe journals and automatic restoration are later chapter concerns.
func (r *runtime) save() error {
	path := r.owner.config.LogPath
	if path == "" {
		return nil
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	return errors.Join(r.owner.history.Dump(file), file.Close())
}

// The request is complete only after its facts are saved. A buffered reply
// tolerates callers that stopped waiting; a failed save cannot become success.
func (r *runtime) finish(text string, err error) {
	if r.active == nil {
		return
	}
	t := r.active
	t.cancel()
	r.active = nil
	err = errors.Join(err, r.save())
	if t.request.reply != nil {
		t.request.reply <- common.Result{Text: text, Err: err}
	}
}

// Snapshot the round policy once per human turn. A later queued request gets
// its own budget and cancellation context, never the previous request's state.
func (r *runtime) begin(q request) {
	if err := q.ctx.Err(); err != nil {
		if q.reply != nil {
			q.reply <- common.Result{Err: err}
		}
		return
	}
	ctx, cancel := context.WithCancel(q.ctx)
	limit := r.owner.config.MaxToolRounds
	if limit == 0 {
		limit = common.DefaultMaxToolRounds
	}
	r.active = &turn{request: q, cancel: cancel, limit: limit}
	r.active.request.ctx = ctx
	if err := r.append(common.Event{Type: "message_received", Message: &common.MessageData{Actor: common.Human, Parts: []common.Part{{Type: "text", Text: q.text}}}}); err != nil {
		r.finish("", err)
		return
	}
	r.send()
}
func (r *runtime) send() {
	t := r.active
	body, err := r.owner.engine.Render(r.owner.history.Context())
	if err != nil {
		r.append(common.Event{Type: "error_occurred", Error: &common.ErrorData{Message: err.Error()}})
		r.finish("", err)
		return
	}
	// Capture delivery before releasing the owner. A hint arriving during HTTP
	// belongs to the NEXT request and must not be consumed by this response.
	if err = r.append(common.Event{Type: "request_sent", Request: &common.RequestData{To: r.owner.engine.Target()}}); err != nil {
		r.finish("", err)
		return
	}
	r.workers++
	go func() {
		events, err := r.owner.engine.Send(t.request.ctx, body)
		r.owner.box.Post(response{t, events, err})
	}()
}

// The turn pointer is a generation token. An interrupted HTTP exchange can
// arrive after another turn started, but cannot change that turn's state.
func (r *runtime) receive(m response) {
	r.workers--
	if m.turn != r.active {
		return
	} // cancelled HTTP cannot finish a newer human turn.
	var answer strings.Builder
	for _, event := range m.events {
		if err := r.append(event); err != nil {
			r.finish("", err)
			return
		}
		if event.Response != nil {
			for _, p := range event.Response.Parts {
				if p.Type == "text" {
					answer.WriteString(p.Text)
				}
				if p.Type == "tool_call" {
					m.turn.calls = append(m.turn.calls, p)
				}
			}
		}
	}
	if m.err == nil && len(m.turn.calls) == 0 && len(r.owner.history.Context().Ephemera) > 0 {
		r.send()
		return
	}
	if m.err != nil || len(m.turn.calls) == 0 {
		r.finish(answer.String(), m.err)
		return
	}
	// A permitted batch has already run and been consumed by the model. Refuse
	// the entire next batch and pair every call; never edit signed assistant data.
	if m.turn.batches >= m.turn.limit {
		for _, call := range m.turn.calls {
			r.append(common.Event{Type: "tool_returned", Tool: &common.ToolData{CallID: call.CallID, IsError: true, Parts: []common.Part{{Type: "text", Text: "not executed: tool round limit reached"}}}})
		}
		r.append(common.Event{Type: "turn_ended", Error: &common.ErrorData{Message: common.ErrRoundLimit.Error()}})
		r.finish("", common.ErrRoundLimit)
		return
	}
	m.turn.batches++
	r.dispatch()
}
func (r *runtime) interrupt(reason error) {
	t := r.active
	if t == nil {
		return
	}
	// Satisfy provider call/result pairing without claiming a continuing job was
	// killed. The later report is still a recorded fact, but not a duplicate result.
	if t.current != nil {
		data := *t.current.data
		text := "wait interrupted"
		if data.Job != nil {
			text = fmt.Sprintf("wait interrupted; job %d continues; output at %s", data.Job.Handle, data.Job.Output.Locator)
		}
		data.Parts = []common.Part{{Type: "text", Text: text}}
		r.append(common.Event{Type: "tool_returned", Tool: &data})
	}
	for _, call := range t.calls {
		r.append(common.Event{Type: "tool_returned", Tool: &common.ToolData{CallID: call.CallID, IsError: true, Parts: []common.Part{{Type: "text", Text: "not executed: turn interrupted"}}}})
	}
	r.append(common.Event{Type: "turn_interrupted", Error: &common.ErrorData{Message: reason.Error()}})
	r.finish("", reason)
}

// Each drain processes causal order, including hints between calls and results.
// Workers post through this same door; no worker borrows the conversation.
func (r *runtime) run() {
	for {
		<-r.owner.box.Wake
		for _, message := range r.owner.box.Drain() {
			switch m := message.(type) {
			case request:
				if r.stopping {
					m.reply <- common.Result{Err: common.ErrStopped}
				} else {
					if r.active == nil && len(r.pending) == 0 && len(r.barriers) > 0 {
						r.begin(m)
					} else {
						r.pending = append(r.pending, m)
					}
				}
			case common.UserMessage:
				r.input(m.Text)
			case common.Hint:
				r.input(m.Text)
			case common.Interrupt:
				r.interrupt(common.ErrInterrupted)
			case response:
				r.receive(m)
			case report:
				// A turn may have ended while this wait was still outstanding. Retain
				// its real report without inserting a second vendor result for one call.
				r.workers--
				kind := "tool_returned"
				if m.turn != r.active {
					kind = "job_report"
				}
				err := r.append(common.Event{Type: kind, Tool: m.data})
				if m.turn == r.active {
					m.turn.current = nil
					if err != nil {
						r.finish("", err)
					} else {
						r.dispatch()
					}
				}
			case workerDone:
				// Execution lifetime is distinct from the report deadline. Joining both
				// prevents shutdown from abandoning a handler after a callback returned.
				r.workers--
				r.failure = errors.Join(r.failure, m.err)
				if m.job != nil {
					data := m.job.Data()
					// A shell handler returns after launch; its job reader still owns
					// completion. Only a terminal snapshot is a finished-job fact.
					if data.Status != "running" {
						r.append(common.Event{Type: "job_finished", Tool: &common.ToolData{Job: &data}})
					}
				}
			case record:
				m.reply <- r.append(m.event)
			case flush:
				r.barriers = append(r.barriers, m.reply)
			case stop:
				if !r.stopping {
					r.stopping = true
					r.interrupt(common.ErrStopped)
					for _, q := range r.pending {
						q.reply <- common.Result{Err: common.ErrStopped}
					}
					r.pending = nil
					r.workers++
					go func() {
						err := r.owner.jobs.Shutdown()
						r.owner.box.Post(workerDone{err: err})
					}()
				}
			}
		}
		// Blocking callers queue as whole turns. Start the next only after all
		// messages already drained have taken effect, not from a completion worker.
		if !r.stopping && r.active == nil && len(r.pending) > 0 {
			q := r.pending[0]
			r.pending = r.pending[1:]
			r.begin(q)
			if r.active == nil {
				r.owner.box.Post(flush{})
			}
		}
		if r.active == nil && len(r.pending) == 0 && len(r.barriers) > 0 {
			err := r.save()
			for _, reply := range r.barriers {
				if reply != nil {
					reply <- err
				}
			}
			r.barriers = nil
		}
		if r.stopping && r.workers == 0 {
			r.owner.stopErr = errors.Join(r.failure, r.save())
			r.owner.lifecycle.Lock()
			close(r.owner.done)
			r.owner.lifecycle.Unlock()
			return
		}
	}
}
func (r *runtime) input(text string) {
	if r.stopping {
		return
	}
	if r.active == nil {
		r.begin(request{ctx: context.Background(), text: text})
		return
	}
	// Reducer uses active turn state to preserve a hint as pending context.
	r.append(common.Event{Type: "hint_received", Message: &common.MessageData{Actor: common.Human, Parts: []common.Part{{Type: "text", Text: text}}}})
}
