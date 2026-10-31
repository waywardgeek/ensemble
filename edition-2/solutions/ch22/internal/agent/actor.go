package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"ensemble/internal/common"
)

// Private control messages share the same FIFO as public input. Request replies
// are buffered so an abandoned caller cannot hold up the conversation owner.
type request struct {
	ctx   context.Context
	text  string
	reply chan common.Result
	watch common.StreamCallbacks
}
type response struct {
	turn    *turn
	events  []common.Event
	partIDs []uint64
	err     error
}

// Stream progress shares the owner FIFO with input and completion. The turn
// token discards late chunks after an interrupt without ending the Agent.
type streamed struct {
	turn  *turn
	delta common.PartDelta
}
type report struct {
	turn *turn
	call common.Part
	data *common.ToolData
	wake chan struct{}
}
type record struct {
	event common.Event
	reply chan error
}
type flush struct{ reply chan error }
type stop struct{ discard bool }
type gateReady struct{ turn *turn }
type workerDone struct {
	job common.Job
	err error
}
type turn struct {
	request        request
	cancel         context.CancelFunc
	calls          []common.Part
	ids            map[uint64]uint64
	current        *report
	batches, limit int
	ephemeralTurn  bool
	presented      []common.Entry
	afterTool      bool
	// Per-response stream order is separate from parser-normalized part order.
	sawStreamCall           bool
	visibleBeforeStreamCall bool
	checkpoint              bool
}

// All fields below are owned by run. Workers own external work only, and return
// immutable messages. Interrupt detaches a turn without detaching its workers.
type runtime struct {
	owner             *agent
	active            *turn
	pending           []request
	barriers          []chan error
	workers           int
	stopping          bool
	discard           bool
	failure           error
	partID            uint64
	memoryReady       bool
	memoryDisabled    [6]bool
	forcingChecked    uint64
	memoryGeneration  [6]uint64
	compaction        *compactionAttempt
	compactionWaiting *turn
	warned            bool
	// Managed workers have one task and a lifetime send budget. Hints and a
	// queued request cannot reset it by creating another ordinary turn.
	compactionStarted bool
	compactionSends   int
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
	return a.submit(ctx, text, common.StreamCallbacks{})
}
func (a *agent) submit(ctx context.Context, text string, watch common.StreamCallbacks) <-chan common.Result {
	reply := make(chan common.Result, 1)
	if strings.TrimSpace(text) == "" {
		reply <- common.Result{Err: errors.New("user message is empty")}
		return reply
	}
	if err := a.post(request{ctx: ctx, text: text, reply: reply, watch: watch}); err != nil {
		reply <- common.Result{Err: err}
	}
	return reply
}

// Ask blocks on its own request reply or caller cancellation, never a shared idle
// event.
func (a *agent) Ask(ctx context.Context, text string) (string, error) {
	return a.AskWatching(ctx, text, common.StreamCallbacks{})
}

// AskWatching delivers this request's stream through the same actor as Ask.
// Frame callbacks run on the HTTP worker; all other callbacks run on the actor.
func (a *agent) AskWatching(ctx context.Context, text string, watch common.StreamCallbacks) (string, error) {
	reply := a.submit(ctx, text, watch)
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
func (a *agent) Shutdown() error { return a.shutdown(false) }

// DiscardConstruction drains a refused factory result without writing its rejected
// persistence paths. Factories must return before loading or starting work.
func (a *agent) DiscardConstruction() error { return a.shutdown(true) }

func (a *agent) shutdown(discard bool) error {
	a.start()
	a.lifecycle.Lock()
	if !a.stopping {
		a.stopping = true
		a.box.Post(stop{discard: discard})
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
	a.activityMu.Lock()
	a.progressTime = time.Now()
	a.activityMu.Unlock()
	a.parent.Observe(a, observation)
	if parent := a.ParentAgent(); parent != nil {
		parent.Fleet().Changed()
	}
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
	if r.active != nil && r.active.request.watch.OnEvent != nil {
		r.active.request.watch.OnEvent(event)
	}
	after := h.Context()
	if before != after.Turn {
		r.owner.notify(common.StateChanged{From: before, To: after.Turn})
	}
	if event.Type == "hint_received" {
		r.owner.notify(common.PartDelta{Kind: common.DeltaText, Hint: true, Chunk: "hint:" + event.Message.Parts[0].Text})
	}
	// Finalized dialogue has its stored Seq; transient deltas never enter history.
	if len(after.Dialogue) > 0 && event.Type == "tool_returned" {
		var entry common.Entry
		for _, candidate := range after.Dialogue {
			if candidate.Seq == h.LastSeq() {
				entry = candidate
				break
			}
		}
		for _, part := range entry.Parts {
			r.partID++
			r.owner.notify(common.PartFinal{Seq: entry.Seq, PartID: r.partID, Part: part})
		}
	}
	return nil
}

// Saving is an owner operation. In particular, do not rewrite again immediately
// after replying: the caller must be able to read the just-completed log.
// The independent crash journal is appended by History before each fact is published.
func (r *runtime) save() error {
	path := r.owner.config.LogPath
	// Constructor refusal drains workers without touching rejected persistence
	// paths, independently of the purpose actually returned by the factory.
	if r.discard || path == "" {
		return nil
	}
	file, err := r.owner.Sandbox().OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
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
	if err == nil && r.owner.spec.Compaction && r.owner.fleet.Status().State != "submitted" {
		err = errors.New("compaction worker finished without submitting memory")
	}
	if err == nil && !r.owner.spec.Compaction && r.owner.Ensemble().FleetEnabled() {
		correction, escalation := r.owner.fleet.Idle()
		if correction != "" {
			if err = r.append(common.Event{Type: "message_received", Message: &common.MessageData{Actor: common.System, Parts: []common.Part{{Type: "text", Text: correction}}}}); err == nil {
				r.owner.notify(common.PartDelta{Kind: common.DeltaText, Hint: true, Chunk: correction})
				r.send()
				return
			}
		}
		if escalation != "" {
			err = r.append(common.Event{Type: "fleet_notice", Message: &common.MessageData{Actor: common.System, Parts: []common.Part{{Type: "text", Text: escalation}}}})
			// Parent delivery may post to another actor or the human sink, never wait
			// for a reply from inside this actor.
			err = errors.Join(err, r.owner.fleet.Escalate(escalation))
		}
	}
	t.cancel()
	if r.compactionWaiting == t {
		r.compactionWaiting = nil
	}
	r.active = nil
	ended := common.TurnEnded{Text: text}
	if err != nil {
		ended.Error = err.Error()
	}
	r.owner.fleet.End(err)
	r.owner.setActivity(common.ActivityIdle, "", "")
	r.owner.notify(ended)
	err = errors.Join(err, r.save())
	if t.request.reply != nil {
		t.request.reply <- common.Result{Text: text, Err: err}
	}
}

// Snapshot the round policy once per human turn. A later queued request gets
// its own budget and cancellation context, never the previous request's state.
func (r *runtime) begin(q request) {
	if r.owner.spec.Compaction {
		if r.compactionStarted {
			if q.reply != nil {
				q.reply <- common.Result{Err: errors.New("compaction worker accepts only one task")}
			}
			return
		}
		r.compactionStarted = true
	}
	if err := r.syncMemory(); err != nil {
		if q.reply != nil {
			q.reply <- common.Result{Err: err}
		}
		return
	}
	if err := q.ctx.Err(); err != nil {
		if q.reply != nil {
			q.reply <- common.Result{Err: err}
		}
		return
	}
	r.owner.fleet.Begin()
	ctx, cancel := context.WithCancel(q.ctx)
	limit := r.owner.Config().MaxToolRounds
	if limit == 0 {
		limit = common.DefaultMaxToolRounds
	}
	if r.owner.spec.Compaction {
		limit = 3
	}
	r.active = &turn{request: q, cancel: cancel, limit: limit}
	r.active.request.ctx = ctx
	if err := r.append(common.Event{Type: "message_received", Message: &common.MessageData{Actor: common.Human, Parts: []common.Part{{Type: "text", Text: q.text}}}}); err != nil {
		r.finish("", err)
		return
	}
	r.send()
}
func (r *runtime) send() { r.ephemera() }
func (r *runtime) sendRequest() {
	limited := r.compaction != nil && r.atContextLimit()
	if err := r.settleCompaction(); err != nil {
		r.compressionWarning(err)
		if limited || r.compaction != nil {
			r.compactionWaiting = nil
			r.finish("", err)
			return
		}
	}
	if r.compaction != nil && r.atContextLimit() {
		r.compactionWaiting = r.active
		r.owner.setActivity(common.ActivityWaiting, "", "waiting for memory compaction")
		return
	}
	r.compactionWaiting = nil
	if r.owner.spec.Compaction && r.compactionSends >= 3 {
		r.finish("", errors.New("compaction worker model request limit reached"))
		return
	}
	r.owner.setActivity(common.ActivityThinking, "", "")
	t := r.active
	t.ids = make(map[uint64]uint64)
	t.sawStreamCall, t.visibleBeforeStreamCall = false, false
	if err := r.prepareContext(); err != nil {
		r.finish("", err)
		return
	}
	before := r.owner.history.Context()
	t.presented = nil
	for _, entry := range before.Dialogue {
		if entry.Seq > before.SentThrough {
			t.presented = append(t.presented, entry)
		}
	}
	body, err := r.owner.engine.Render(r.owner.history.Context())
	if err != nil {
		r.append(common.Event{Type: "error_occurred", Error: &common.ErrorData{Message: err.Error()}})
		r.finish("", err)
		return
	}
	// Capture delivery before releasing the owner. A hint arriving during HTTP
	// belongs to the NEXT request and must not be consumed by this response.
	request := &common.RequestData{To: r.owner.engine.Target()}
	if before.StartupTools == nil {
		declarations := r.owner.Declarations()
		if declarations == nil {
			declarations = []common.ToolDeclaration{}
		}
		request.StartupTools = &declarations
	}
	if err = r.append(common.Event{Type: "request_sent", Request: request}); err != nil {
		r.finish("", err)
		return
	}
	if r.owner.spec.Compaction {
		r.compactionSends++
	}
	r.workers++
	go func() {
		var ids []uint64
		cb := common.StreamCallbacks{
			OnFrame: t.request.watch.OnFrame,
			OnDelta: func(id uint64, kind common.DeltaKind, chunk string) {
				r.owner.box.Post(streamed{t, common.PartDelta{PartID: id, Kind: kind, Chunk: chunk}})
			},
			OnPartFinal: func(id uint64, _ common.Part) { ids = append(ids, id) },
		}
		events, err := r.owner.engine.Send(t.request.ctx, body, cb)
		r.owner.box.Post(response{t, events, ids, err})
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
		if event.Response != nil {
			for i := range event.Response.Parts {
				event.Response.Parts[i].ArtifactID = r.displayID(m.turn, m.partIDs[i])
			}
		}
		if err := r.append(event); err != nil {
			r.finish("", err)
			return
		}
		if event.Response != nil {
			// History may assign missing call identities. Publish the stored parts,
			// retaining the parser's mapping instead of guessing IDs from positions.
			c := r.owner.history.Context()
			entry := c.Dialogue[len(c.Dialogue)-1]
			for i, p := range entry.Parts {
				r.owner.notify(common.PartFinal{Seq: entry.Seq, PartID: p.ArtifactID, Part: p})
				if cb := m.turn.request.watch.OnPartFinal; cb != nil {
					cb(m.partIDs[i], p)
				}
			}
			for _, p := range entry.Parts {
				if p.Type == "text" {
					answer.WriteString(p.Text)
				}
				if p.Type == "tool_call" {
					m.turn.calls = append(m.turn.calls, p)
				}
			}
		}
	}
	if m.err == nil {
		if err := r.stubPrevious(m.turn); err != nil {
			r.finish("", err)
			return
		}
	}
	if m.err == nil && !r.owner.spec.Compaction && len(m.turn.calls) == 0 && len(r.owner.history.Context().Ephemera) > 0 {
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
	// Rejected explanation-less batches also consume the existing round budget:
	// corrective feedback must not create an unbounded model-only retry loop.
	m.turn.batches++
	if r.owner.spec.Compaction && (len(m.turn.calls) != 1 || m.turn.calls[0].Name != "submit") {
		// Refuse the entire batch before creating jobs; accepting a first submit
		// would conceal additional calls and prematurely mark work complete.
		for _, call := range m.turn.calls {
			if err := r.append(common.Event{Type: "tool_returned", Tool: &common.ToolData{CallID: call.CallID, IsError: true, Parts: []common.Part{{Type: "text", Text: "compaction requires exactly one submit call"}}}}); err != nil {
				r.finish("", err)
				return
			}
		}
		m.turn.calls = nil
		r.send()
		return
	}
	toolFirstStream := m.turn.sawStreamCall && !m.turn.visibleBeforeStreamCall
	if r.owner.engine.Features().RequireVisibleReasoning && (toolFirstStream || !visibleExplanationBeforeTools(m.events)) {
		r.rejectUnexplainedTools(toolFirstStream)
		return
	}
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
				if m.AfterTool && r.active != nil {
					r.active.afterTool = true
				} else {
					r.interrupt(common.ErrInterrupted)
				}
			case streamed:
				if m.turn == r.active {
					// Preserve actual stream order even when a provider parser later
					// normalizes completed parts into text-before-tools order.
					switch m.delta.Kind {
					case common.DeltaText:
						if !m.turn.sawStreamCall && strings.TrimSpace(m.delta.Chunk) != "" {
							m.turn.visibleBeforeStreamCall = true
						}
					case common.DeltaToolCall:
						m.turn.sawStreamCall = true
					}
					delta := m.delta
					delta.PartID = r.displayID(m.turn, delta.PartID)
					r.owner.notify(delta)
					if cb := m.turn.request.watch.OnDelta; cb != nil {
						cb(m.delta.PartID, m.delta.Kind, m.delta.Chunk)
					}
				}
			case gateReady:
				r.workers--
				if m.turn == r.active {
					if err := m.turn.request.ctx.Err(); err != nil {
						r.interrupt(err)
					} else {
						r.dispatch()
					}
				}
			case invocationReady:
				r.workers--
				r.invoke(m.request)
			case invocation:
				r.invoke(m)
			case ephemeralReady:
				r.workers--
				if m.turn == r.active {
					if m.text != "" {
						r.append(common.Event{Type: "message_received", Message: &common.MessageData{Actor: common.System, Parts: []common.Part{{Type: "text", Text: m.text}}}})
					}
					if err := m.turn.request.ctx.Err(); err != nil {
						r.interrupt(err)
					} else {
						r.sendRequest()
					}
				}
			case modelSelection:
				r.selectModel(m)
			case memorySettings:
				r.changeMemory(m)
			case compactionFinished:
				r.compactionFinished(m)
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
				if err == nil && kind == "tool_returned" && m.call.Name == "micro_handoff" && !m.data.IsError {
					var handoff common.MicroHandoffData
					if err = json.Unmarshal(m.call.Args, &handoff); err == nil {
						err = r.append(common.Event{Type: "micro_handoff", Handoff: &handoff})
						m.turn.checkpoint = err == nil
					}
				}
				var result strings.Builder
				for _, p := range m.data.Parts {
					result.WriteString(p.Text)
				}
				r.owner.notify(common.ToolFinished{CallID: m.data.CallID, Result: result.String(), IsError: m.data.IsError})
				if m.turn == r.active {
					m.turn.current = nil
					if err != nil {
						r.finish("", err)
					} else if m.turn.afterTool {
						r.interrupt(common.ErrInterrupted)
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
					r.discard = m.discard
					// Release reverse callers waiting at the same Agent-wide gate.
					r.owner.Unpause()
					r.interrupt(common.ErrStopped)
					for _, q := range r.pending {
						q.reply <- common.Result{Err: common.ErrStopped}
					}
					r.pending = nil
					r.workers++
					go func() {
						err := errors.Join(r.owner.jobs.Shutdown(), r.owner.fleet.Shutdown(), r.owner.mcp.Close())
						r.owner.box.Post(workerDone{err: err})
					}()
				}
			}
		}
		if r.active == nil && r.compaction != nil && r.compaction.completion != nil {
			if err := r.settleCompaction(); err != nil {
				r.compressionWarning(err)
				if r.stopping && r.compaction != nil {
					r.failure = errors.Join(r.failure, err)
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
	// Close only the current wait's signal. A message arriving with no waiter
	// cannot leave a sticky interrupt for a future job.
	if current := r.active.current; current != nil && current.wake != nil {
		close(current.wake)
		current.wake = nil
	}
}

// Parser IDs restart each response. Assign display identities once in the actor,
// preserving the same identity in the stored part and every live observation.
func (r *runtime) displayID(t *turn, local uint64) uint64 {
	if id := t.ids[local]; id != 0 {
		return id
	}
	r.partID++
	t.ids[local] = r.partID
	return r.partID
}
