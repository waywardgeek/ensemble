package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"ensemble/internal/common"
)

// The slice is the lossless FIFO; wake is only a coalescing signal.
type mailbox struct {
	mu     sync.Mutex
	queue  []any
	wake   chan struct{}
	closed bool
}

func (b *mailbox) post(v any, public bool) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if public && b.closed {
		return common.ErrStopped
	}
	b.queue = append(b.queue, v)
	select {
	case b.wake <- struct{}{}:
	default:
	}
	return nil
}
func (b *mailbox) drain() []any {
	b.mu.Lock()
	defer b.mu.Unlock()
	q := b.queue
	b.queue = nil
	return q
}

type request struct {
	live        bool
	turnID      uint64
	accepted    bool
	submission  json.RawMessage
	message     *common.MessageData
	hintMessage bool
	ctx         context.Context
	text        string
	reply       chan common.Result
	done        chan struct{}
}
type completeRequest struct{ request *request }

type cancelRequest struct {
	request *request
	err     error
}
type streamNotice struct {
	turn, response uint64
	delta          common.PartDelta
}
type response struct {
	turn     uint64
	response uint64
	events   []common.Event
	err      error
}
type gathered struct {
	turn        uint64
	round, once common.PartList
	err         error
}
type report struct {
	turn uint64
	call common.ToolData
}
type stop struct{}
type invoke struct {
	ctx     context.Context
	call    common.Part
	refusal string
	reply   chan error
}
type invoked struct {
	data  common.ToolData
	reply chan error
}

// Complete enters the same terminal validation/save owner as the submit tool.
func (a *agent) Complete(ctx context.Context, text string) <-chan common.Result {
	r := &request{ctx: ctx, text: text, reply: make(chan common.Result, 1), done: make(chan struct{})}
	if err := a.box.post(completeRequest{r}, true); err != nil {
		a.reply(r, "", err)
	}
	return r.reply
}
func (a *agent) completeResult(r *request) {
	if a.stopping {
		a.reply(r, "", common.ErrStopped)
		return
	}
	if err := r.ctx.Err(); err != nil {
		a.reply(r, "", err)
		return
	}
	raw, _ := json.Marshal(map[string]string{"result": r.text})
	payload, err := common.PrepareSubmission(a.Config().SubmitSchema, raw)
	if err != nil {
		a.reply(r, "", err)
		return
	}
	fleet, ok := a.parent.(common.Fleet)
	if !ok {
		a.reply(r, "", fmt.Errorf("terminal completion requires a fleet owner"))
		return
	}
	if err = fleet.AcceptSubmission(a.ID(), raw); err != nil {
		a.reply(r, "", err)
		return
	}
	a.closeTerminalAdmission()
	t := a.active
	if t == nil {
		a.nextTurn++
		r.turnID = a.nextTurn
		ctx, cancel := context.WithCancel(r.ctx)
		t = &turn{id: r.turnID, request: r, ctx: ctx, cancel: cancel}
		a.active = t
	} else {
		r.turnID = t.id
		r.submission = payload
		t.followers = append(t.followers, r)
		if t.httpCancel != nil {
			t.httpCancel()
		}
		if t.observeCancel != nil {
			t.observeCancel()
		}
		if t.operation != nil {
			d := t.operation.Called()
			d.IsError = true
			d.Parts = common.PartList{common.Text("interrupted by terminal completion")}
			err = errors.Join(err, a.returned(t.id, d))
			t.operation = nil
			t.index++
		}
		a.skipTail("skipped after terminal completion")
	}
	t.request.accepted = true
	t.request.submission = payload
	d := common.ToolData{Name: "submit", CallID: fmt.Sprintf("complete-%d", t.id), Args: raw, Parts: common.PartList{common.Text(string(payload))}}
	_, recordErr := a.record(common.Event{Type: common.ToolCalled, TurnID: t.id, Tool: &d})
	d.FleetCompletion = true
	err = errors.Join(err, recordErr, a.returned(t.id, d))
	a.finish(string(payload), err)
}

func (a *agent) InvokeTool(ctx context.Context, call common.Part, refusal string) error {
	reply := make(chan error, 1)
	if err := a.box.post(invoke{ctx, call, refusal, reply}, true); err != nil {
		return err
	}
	select {
	case err := <-reply:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

type admittedResult struct {
	data  common.ToolData
	reply chan common.ToolData
}
type stopped struct{ err error }
type pauseChanged struct{}

type turn struct {
	gathering, turnGathered             bool
	turnEphemera                        common.PartList
	id                                  uint64
	response                            uint64
	arrivalInvalid, toolSeen, explained bool
	request                             *request
	followers                           []*request
	ctx                                 context.Context
	cancel                              context.CancelFunc
	httpCancel                          context.CancelFunc
	observeCancel                       context.CancelFunc
	calls                               common.PartList
	index                               int
	operation                           common.ToolOperation
	refusal                             string
	skip                                bool
	hinted                              bool
	wakes                               int
	rounds, limit                       int
	partIDs                             map[uint64]uint64
	speechRequested                     bool
}

// Start is called once by the composition root, after all child links exist.
func (a *agent) Start() {
	a.started.Store(true)
	go a.run()
}
func (a *agent) Post(input common.Inbound) error {
	if input == nil {
		return errors.New("nil input")
	}
	switch m := input.(type) {
	case common.Pause:
		return a.Pause()
	case common.Unpause:
		if m.TurnID == 0 {
			return a.Unpause()
		}
	}
	return a.box.post(input, true)
}
func (a *agent) Submit(ctx context.Context, text string) <-chan common.Result {
	return a.submit(ctx, text, false)
}
func (a *agent) SubmitLive(ctx context.Context, text string) <-chan common.Result {
	return a.submit(ctx, text, true)
}
func (a *agent) submit(ctx context.Context, text string, live bool) <-chan common.Result {
	r := &request{live: live, ctx: ctx, text: text, reply: make(chan common.Result, 1), done: make(chan struct{})}
	if err := a.box.post(r, true); err != nil {
		r.reply <- common.Result{Err: err}
		close(r.reply)
		close(r.done)
		return r.reply
	}
	go func() {
		select {
		case <-ctx.Done():
			_ = a.box.post(cancelRequest{r, ctx.Err()}, false)
		case <-r.done:
		case <-a.done:
		}
	}()
	return r.reply
}
func (a *agent) Ask(ctx context.Context, text string) (string, error) {
	result := <-a.Submit(ctx, text)
	return result.Text, result.Err
}

type killedReport struct {
	report   common.JobReport
	recorded chan struct{}
}

func (a *agent) ReportJob(r common.JobReport) {
	if r.Job.Status != common.JobKilledStatus {
		_ = a.box.post(r, false)
		return
	}
	recorded := make(chan struct{})
	_ = a.box.post(killedReport{r, recorded}, false)
	// A direct Kill must not return before its terminal fact is recorded.
	// This acknowledges the actor, never the uncooperative handler.
	select {
	case <-recorded:
	case <-a.done:
	}
}
func (a *agent) Shutdown() error {
	a.box.mu.Lock()
	if !a.box.closed {
		a.box.closed = true
		a.box.queue = append(a.box.queue, stop{})
		select {
		case a.box.wake <- struct{}{}:
		default:
		}
	}
	a.box.mu.Unlock()
	<-a.done
	return a.shutdownErr
}

func (a *agent) recordJob(r common.JobReport) error {
	if a.liveCalls == nil {
		a.liveCalls = map[string]bool{}
	}
	if r.Job.Status == common.JobRunning {
		a.liveCalls[r.Job.CallID] = true
	} else {
		delete(a.liveCalls, r.Job.CallID)
	}
	r.Job.CallPartID = a.jobParts[r.Job.Handle]
	typ := common.JobReported
	if r.Job.Status == common.JobKilledStatus {
		typ = common.JobKilled
	}
	_, err := a.record(common.Event{Type: typ, TurnID: r.Job.TurnID, Job: &r.Job})
	return err
}
func (a *agent) run() {
	for {
		<-a.box.wake
		for _, message := range a.box.drain() {
			switch m := message.(type) {
			case factRequest:
				_, err := a.record(m.event)
				m.reply <- err
			case modelRequest:
				m.reply <- a.changeModel(m.model)
			case contextRequest:
				a.contextRequest(m)
			case *request:
				if a.stopping {
					a.reply(m, "", common.ErrStopped)
				} else if err := m.ctx.Err(); err != nil {
					a.reply(m, "", err)
				} else if m.live && a.active != nil {
					m.turnID = a.active.id
					a.active.followers = append(a.active.followers, m)
					a.hint(common.MessageData{Actor: common.Human, Parts: common.PartList{common.Text(m.text)}})
				} else {
					a.pending = append(a.pending, m)
				}
			case cancelRequest:
				if a.active != nil && a.active.request == m.request {
					a.interrupt(m.err)
				} else {
					if a.active != nil {
						for i, r := range a.active.followers {
							if r == m.request {
								a.active.followers = append(a.active.followers[:i], a.active.followers[i+1:]...)
								a.reply(r, "", m.err)
								break
							}
						}
					}
					for i, r := range a.pending {
						if r == m.request {
							a.pending = append(a.pending[:i], a.pending[i+1:]...)
							a.reply(r, "", m.err)
							break
						}
					}
				}
			case common.UserMessage:
				if !a.stopping {
					if a.active == nil {
						a.pending = append(a.pending, &request{ctx: context.Background(), text: m.Text})
					} else {
						a.hint(common.MessageData{Actor: common.Human, Parts: common.PartList{common.Text(m.Text)}})
					}
				}
			case completeRequest:
				a.completeResult(m.request)
			case common.Ephemeral:
				_, err := a.record(common.Event{Type: common.MessageReceived, Message: &common.MessageData{Actor: common.System, Parts: common.PartList{common.Text(m.Text)}}})
				if err != nil {
					a.finish("", err)
				}
			case common.Hint:
				if !a.stopping {
					a.hint(common.MessageData{Actor: common.Human, Parts: common.PartList{common.Text(m.Text)}})
				}
			case common.Question:
				if !a.stopping {
					a.hint(common.MessageData{Actor: common.Human, Parts: common.PartList{common.Text(m.Text)}, Sender: m.Sender, Recipient: m.Recipient, RequestID: m.RequestID, TurnID: m.TurnID})
				}
			case common.Unpause:
				if m.TurnID != 0 && (a.active == nil || a.active.id != m.TurnID) {
					continue
				}
				_ = a.Unpause()
			case pauseChanged:
				paused := a.isPaused()
				a.publish(common.Paused{Paused: paused})
				if !paused {
					resume := a.gate
					a.gate = nil
					if resume != nil {
						resume()
					}
				}
			case common.SetSpeaking:
				a.speaking = m.Enabled
				a.publish(common.SpeakingChanged{Speaking: m.Enabled})
				if a.gate != nil {
					a.requestPauseSpeech()
				}
			case common.SpeakParts:
				a.publish(common.SpeechRequested{Parts: m.Parts, Reason: m.Reason})
			case common.Interrupt:
				a.interrupt(common.ErrInterrupted)
			case invoke:
				if a.stopping {
					m.reply <- common.ErrStopped
					break
				}
				if err := m.ctx.Err(); err != nil {
					m.reply <- err
					break
				}
				if a.active != nil {
					m.reply <- errors.New("direct Invoke requires an idle conversation")
					break
				}
				a.partID++
				m.call.DisplayID = a.partID
				op := a.tools.Prepare(m.ctx, m.call, m.refusal, 0)
				if err := a.admit(op, 0); err != nil {
					m.reply <- err
					break
				}
				a.workers.Add(1)
				go func() { defer a.workers.Done(); _ = a.box.post(invoked{op.Await(m.ctx), m.reply}, false) }()
			case invoked:
				m.reply <- a.returned(0, m.data)
			case common.ToolAdmission:
				var err error
				if a.stopping {
					err = common.ErrStopped
				} else {
					err = m.Context.Err()
				}
				if err != nil {
					m.Operation.Abort(err)
				} else {
					err = a.admit(m.Operation, 0)
				}
				if err != nil {
					m.Reply <- m.Operation.Await(context.Background())
					break
				}
				a.workers.Add(1)
				go func() {
					defer a.workers.Done()
					_ = a.box.post(admittedResult{m.Operation.Await(m.Context), m.Reply}, false)
				}()
			case admittedResult:
				if err := a.returned(0, m.data); err != nil {
					m.data.IsError = true
					m.data.Parts = common.PartList{common.Text(err.Error())}
				}
				m.reply <- m.data
			case streamNotice:
				t := a.active
				if t == nil || t.id != m.turn || t.response != m.response || t.httpCancel == nil || t.ctx.Err() != nil {
					a.parent.Trace(common.TraceRecord{Stage: "stale_discard", TurnID: m.turn, ResponseID: m.response, PartID: m.delta.PartID, Kind: string(m.delta.Kind)})
					continue
				}
				if m.delta.Kind == common.DeltaText && !t.toolSeen && strings.TrimSpace(m.delta.Chunk) != "" {
					t.explained = true
				}
				if m.delta.Kind == common.DeltaToolCall {
					if !t.explained {
						t.arrivalInvalid = true
					}
					t.toolSeen = true
				}
				m.delta.TurnID, m.delta.ResponseID = m.turn, m.response
				m.delta.PartID = a.displayID(t, m.delta.PartID)
				a.publish(m.delta)
				a.parent.Trace(common.TraceRecord{Stage: "delta_publish", TurnID: m.turn, ResponseID: m.response, PartID: m.delta.PartID, Kind: string(m.delta.Kind), Bytes: len(m.delta.Chunk)})
			case gathered:
				a.gathered(m)
			case response:
				a.receive(m)
			case report:
				a.report(m)
			case common.JobReport:
				if err := a.recordJob(m); err != nil {
					a.parent.Logger().Print(err)
				}
			case killedReport:
				if err := a.recordJob(m.report); err != nil {
					a.parent.Logger().Print(err)
				}
				close(m.recorded)
			case stop:
				a.stopping = true
				a.interrupt(common.ErrStopped)
				for _, r := range a.pending {
					a.reply(r, "", common.ErrStopped)
				}
				a.pending = nil
				terminal := a.terminal != nil
				go func() {
					var mcpErr error
					if terminal {
						if fleet, ok := a.parent.(common.Fleet); ok {
							mcpErr = fleet.ShutdownChildren(a.ID())
						}
					}
					if a.mcp != nil {
						mcpErr = errors.Join(mcpErr, a.mcp.Close())
					}
					err := errors.Join(mcpErr, a.jobs.Shutdown())
					a.workers.Wait()
					_ = a.box.post(stopped{err}, false)
				}()
			case stopped:
				a.shutdownErr = errors.Join(m.err, a.save(), a.Save())
				if end := a.terminal; end != nil {
					a.shutdownErr = errors.Join(a.shutdownErr, end.err)
					if a.shutdownErr == nil && end.turn.request.accepted && !end.recovered {
						a.shutdownErr = common.SaveSubmission(a.Sandbox(), a.Config(), end.turn.request.submission)
					}
				}
				a.shutdownErr = errors.Join(a.shutdownErr, a.Sandbox().Close())
				close(a.done)
				if end := a.terminal; end != nil {
					if fleet, ok := a.parent.(common.Fleet); ok {
						fleet.CompleteSubmission(a.ID(), end.turn.request.submission, a.shutdownErr)
					}
					a.reply(end.turn.request, end.text, a.shutdownErr)
					for _, follower := range end.turn.followers {
						a.reply(follower, end.text, a.shutdownErr)
					}
				}
				return
			}
			a.startNext()
		}
	}
}
func (a *agent) startNext() {
	for a.active == nil && !a.stopping && len(a.pending) > 0 {
		r := a.pending[0]
		a.pending = a.pending[1:]
		if err := r.ctx.Err(); err != nil {
			a.reply(r, "", err)
			continue
		}
		if strings.TrimSpace(r.text) == "" {
			a.reply(r, "", errors.New("user message is empty"))
			continue
		}
		a.nextTurn++
		ctx, cancel := context.WithCancel(r.ctx)
		limit := int(a.maxToolRounds.Load())
		if limit == 0 {
			limit = common.DefaultMaxToolRounds
		}
		a.active = &turn{id: a.nextTurn, request: r, ctx: ctx, cancel: cancel, limit: limit}
		r.turnID = a.active.id
		event := common.Event{Type: common.MessageReceived, TurnID: a.active.id, Message: &common.MessageData{Actor: common.Human, Parts: common.PartList{common.Text(r.text)}}}
		if r.message != nil {
			if r.hintMessage {
				event.Type = common.HintReceived
			}
			event.Message = r.message
		}
		if _, err := a.record(event); err != nil {
			a.finish("", err)
			continue
		}
		a.transition(common.InputPending)
		a.send()
	}
}
func (a *agent) hint(m common.MessageData) {
	if m.Sender != "" || m.Recipient != "" || m.RequestID != "" || m.TurnID != "" {
		metadata, _ := json.Marshal(map[string]string{"sender": m.Sender, "recipient": m.Recipient, "request_id": m.RequestID, "turn_id": m.TurnID})
		m.Parts = append(common.PartList{common.Text("Addressed question " + string(metadata) + "\n")}, m.Parts...)
	}
	if a.active == nil {
		var text strings.Builder
		for _, p := range m.Parts {
			if p.Type == "text" {
				text.WriteString(p.Text)
			}
		}
		a.pending = append(a.pending, &request{ctx: context.Background(), text: text.String(), message: &m, hintMessage: true})
		return
	}
	id := a.active.id
	if _, err := a.record(common.Event{Type: common.HintReceived, TurnID: id, Message: &m}); err != nil {
		if a.active != nil {
			a.finish("", err)
		}
		return
	}
	sender := m.Sender
	if sender == "" {
		sender = "human"
	}
	a.publish(common.HintObserved{TurnID: id, Sender: sender, Parts: m.Parts})
	if a.active != nil {
		a.active.hinted = true
		if a.active.observeCancel != nil {
			a.active.skip = true
			a.active.observeCancel()
		}
	}
}
func (a *agent) send() {
	t := a.active
	if a.waitGate("Before next model request", a.send) {
		return
	}
	if t.gathering {
		return
	}
	if a.mcp == nil {
		a.sendReady(nil)
		return
	}
	t.gathering = true
	ctx, cancel := context.WithCancel(t.ctx)
	t.httpCancel = cancel
	turnID, first := t.id, !t.turnGathered
	a.workers.Add(1)
	go func() {
		defer a.workers.Done()
		defer cancel()
		round, once := a.mcp.Gather(ctx, first)
		_ = a.box.post(gathered{turnID, round, once, ctx.Err()}, false)
	}()
}
func (a *agent) gathered(g gathered) {
	t := a.active
	if t == nil || t.id != g.turn || t.ctx.Err() != nil || g.err != nil {
		return
	}
	t.gathering = false
	t.httpCancel = nil
	t.turnGathered = true
	if len(g.once) > 0 {
		t.turnEphemera = g.once
	}
	parts := append(append(common.PartList(nil), t.turnEphemera...), g.round...)
	a.sendReady(parts)
}
func (a *agent) sendReady(automatic common.PartList) {
	if a.waitGate("Before next model request", func() { a.sendReady(automatic) }) {
		return
	}
	t := a.active
	if len(automatic) > 0 {
		parts := append(a.history.Context().Ephemera, automatic...)
		if _, err := a.record(common.Event{Type: common.MessageReceived, TurnID: t.id, Message: &common.MessageData{Actor: common.System, Parts: parts}}); err != nil {
			a.finish("", err)
			return
		}
	}
	// The final consuming request gets a private stop instruction. Keep it in
	// request-local ephemera: neither durable History nor signed assistant parts
	// may be rewritten, and later ordinary turns must not inherit this policy.
	requestContext, err := a.prepareContext()
	if err != nil {
		a.finish("", err)
		return
	}
	if t.limit > 0 && t.rounds >= t.limit {
		requestContext.Ephemera = append(requestContext.Ephemera, common.Text("The tool-round limit has been reached for this turn. Do not call any more tools. Use the results already available to give your final response."))
	}
	if _, err := a.record(common.Event{Type: common.RequestSent, TurnID: t.id, Request: &common.RequestData{To: a.config.Target()}}); err != nil {
		a.finish("", err)
		return
	}
	t.hinted = false
	a.transition(common.InFlight)
	ctx, cancel := context.WithCancel(t.ctx)
	t.httpCancel = cancel
	t.response++
	t.partIDs = make(map[uint64]uint64)
	t.arrivalInvalid, t.toolSeen, t.explained = false, false, false
	responseID := t.response
	a.parent.Trace(common.TraceRecord{Stage: "request_sent", TurnID: t.id, ResponseID: responseID})
	a.workers.Add(1)
	go func() {
		defer a.workers.Done()
		defer cancel()

		if ctx.Err() != nil {
			_ = a.box.post(response{turn: t.id, response: responseID, err: ctx.Err()}, false)
			return
		}
		body, err := a.engine.Render(requestContext)
		if err != nil {
			_ = a.box.post(response{turn: t.id, response: responseID, err: err}, false)
			return
		}
		events, err := a.engine.ExchangeRequest(ctx, body, common.ParseCallbacks{OnDelta: func(d common.PartDelta) {
			_ = a.box.post(streamNotice{t.id, responseID, d}, false)
		}})
		_ = a.box.post(response{turn: t.id, response: responseID, events: events, err: err}, false)
	}()
}
func (a *agent) receive(r response) {
	if a.active != nil && a.active.id == r.turn && a.active.ctx.Err() != nil {
		a.interrupt(a.active.ctx.Err())
	}
	stale := a.active == nil || a.active.id != r.turn || (r.response != 0 && a.active.response != r.response)
	if stale {
		a.parent.Trace(common.TraceRecord{Stage: "stale_discard", TurnID: r.turn, ResponseID: r.response})
	}
	if !stale {
		a.active.httpCancel = nil
	}
	idErr := a.identifyCalls(r.events)
	var parts common.PartList
	for _, e := range r.events {
		e.TurnID = r.turn
		e.Stale = stale
		if e.Response != nil {
			for i := range e.Response.Parts {
				p := &e.Response.Parts[i]
				if !stale && len(a.active.partIDs) > 0 {
					p.DisplayID = a.displayID(a.active, p.ID)
				} else {
					a.partID++
					p.DisplayID = a.partID
				}
			}
		}
		seq, err := a.record(e)
		if err != nil {
			if !stale {
				a.finish("", err)
			}
			return
		}
		if e.Type == common.ResponseEnded && !stale {
			parts = e.Response.Parts
			if r.err == nil && idErr == nil {
				for _, p := range e.Response.Parts {
					a.publish(common.PartFinal{Seq: uint64(seq), TurnID: r.turn, ResponseID: r.response, PartID: p.DisplayID, Part: p})
					a.parent.Trace(common.TraceRecord{Stage: "part_final", TurnID: r.turn, ResponseID: r.response, PartID: p.DisplayID, Kind: p.Type})
				}
			}
		}
	}
	if stale {
		return
	}
	if r.err != nil {
		a.finish("", r.err)
		return
	}
	if idErr != nil {
		a.finish("", idErr)
		return
	}
	var text strings.Builder
	explained := false
	t := a.active
	t.calls = nil
	t.index = 0
	t.refusal = ""
	t.skip = false
	for _, p := range parts {
		if p.Type == "text" {
			text.WriteString(p.Text)
			if len(t.calls) == 0 && strings.TrimSpace(p.Text) != "" {
				explained = true
			}
		}
		if p.Type == "tool_call" {
			t.calls = append(t.calls, p)
		}
	}
	if len(t.calls) == 0 {
		if t.hinted {
			a.send()
		} else {
			a.finish(text.String(), nil)
		}
		return
	}
	a.transition(common.ToolsPending)
	if a.config.RequireToolExplanation && (!explained || t.arrivalInvalid) {
		t.refusal = "No tools ran: provide nonblank ordinary assistant text explaining what will run and why before the first tool call, then retry."
	}
	// Input admitted during HTTP makes this entire batch undispatched.
	// Pair it without spending an allowance for a dispatched batch.
	if t.hinted {
		a.skipTail("interrupted, not executed: live input arrived before dispatch")
		a.send()
		return
	}
	if t.rounds >= t.limit {
		a.skipTail("not executed: tool round limit reached")
		a.finish("", common.ErrToolRoundLimit)
		return
	}
	t.rounds++
	a.dispatch()
}
func (a *agent) dispatch() {
	t := a.active
	if t.index == len(t.calls) {
		a.transition(common.InputPending)
		a.send()
		return
	}
	if a.waitGate("Before dispatching tool "+t.calls[t.index].Name, a.dispatch) {
		return
	}
	call := t.calls[t.index]
	op := a.tools.Prepare(t.ctx, call, t.refusal, t.id)
	if err := a.admit(op, t.id); err != nil {
		a.finish("", err)
		return
	}
	t.operation = op
	ctx, cancel := context.WithCancel(context.Background())
	t.observeCancel = cancel
	a.workers.Add(1)
	go func() {
		defer a.workers.Done()
		defer cancel()
		d := op.Await(ctx)
		_ = a.box.post(report{t.id, d}, false)
	}()
}

// History and its configured save are the release barrier for every ordinary
// operation. Failed admission settles its reservation without running a handler.
func (a *agent) admit(op common.ToolOperation, turn uint64) error {
	d := op.Called()
	d.Dispatched = false
	if d.DisplayID == 0 {
		a.partID++
		d.DisplayID = a.partID
		d.CallPartID = d.DisplayID
	}
	if d.Job != nil {
		a.jobParts[d.Job.Handle] = d.DisplayID
		d.Job.CallPartID = d.DisplayID
	}
	seq, err := a.record(common.Event{Type: common.ToolCalled, TurnID: turn, Tool: &d})
	if err != nil {
		op.Abort(err)
		// A save can fail after append. Pair that recorded call truthfully;
		// an append failure has no call to pair and must not invent one.
		if seq != 0 {
			err = errors.Join(err, a.returned(turn, op.Await(context.Background())))
		}
		return err
	}
	op.SetEventSeq(seq)
	err = op.ReleaseAfter(func() error {
		d.Dispatched = true
		if _, err := a.recordSaved(common.Event{Type: common.ToolReleased, TurnID: turn, Tool: &d}); err != nil {
			return err
		}
		a.publish(common.ToolDispatched{TurnID: turn, PartID: d.CallPartID, CallID: d.CallID, Name: d.Name, Input: append(json.RawMessage(nil), d.Args...)})
		return nil
	})
	if err != nil {
		op.Abort(err)
		if recordErr := a.returned(turn, op.Await(context.Background())); recordErr != nil {
			err = errors.Join(err, recordErr)
		}
		return err
	}
	return nil
}
func (a *agent) returned(turn uint64, d common.ToolData) error {
	if d.Job != nil {
		d.CallPartID = a.jobParts[d.Job.Handle]
		d.Job.CallPartID = d.CallPartID
	}
	a.partID++
	d.DisplayID = a.partID
	seq, err := a.record(common.Event{Type: common.ToolReturned, TurnID: turn, Tool: &d})
	if err == nil && !d.Automatic {
		a.publish(common.ToolFinished{TurnID: turn, PartID: d.DisplayID, CallID: d.CallID, Name: d.Name, Parts: d.Parts, IsError: d.IsError})
		a.final(seq, common.Part{DisplayID: d.DisplayID, Type: "tool_result", CallID: d.CallID, Name: d.Name, Parts: d.Parts, IsError: d.IsError})
	}
	if err == nil && d.Name == "micro_handoff" && !d.IsError && a.handoffText != "" {
		_, err = a.record(common.Event{Type: common.MicroHandoff, Handoff: &common.HandoffData{Text: a.handoffText}})
		a.handoffText = ""
		a.handoffSavePending = err == nil
	}
	if err == nil && a.handoffSavePending && a.history.Context().PendingHandoff == "" {
		err = a.Save()
		if err == nil {
			a.handoffSavePending = false
			if a.Config().SavePath != "" {
				if writer, ok := a.parent.Logger().Writer().(interface{ Retain() error }); ok {
					err = writer.Retain()
				}
			}
		}
	}
	return err
}
func (a *agent) report(r report) {
	t := a.active
	if t == nil || t.id != r.turn || t.operation == nil || t.calls[t.index].CallID != r.call.CallID {
		_, err := a.record(common.Event{Type: common.JobReported, TurnID: r.turn, Tool: &r.call})
		if err != nil {
			a.parent.Logger().Print(err)
		}
		return
	}
	t.observeCancel = nil
	t.operation = nil
	accepted := r.call.FleetCompletion && !r.call.IsError
	if accepted {
		// Admission closes inside this actor before a queued Submit/Invoke can
		// run. The fleet observer only publishes the already-stopped outcome.
		t.request.accepted = true
		t.request.submission, _ = common.PrepareSubmission(a.Config().SubmitSchema, r.call.Args)
		a.closeTerminalAdmission()
	}
	if err := a.returned(t.id, r.call); err != nil {
		a.finish("", err)
		return
	}
	t.index++
	if accepted {
		a.skipTail("skipped after accepted terminal submit")
		a.finish(string(t.request.submission), nil)
		return
	}
	if t.skip {
		a.skipTail("interrupted, not executed: live input interrupted this pending batch")
	}
	a.dispatch()
}
func (a *agent) skipTail(reason string) {
	t := a.active
	for t.index < len(t.calls) {
		op := a.tools.Prepare(t.ctx, t.calls[t.index], reason, t.id)
		d := op.Called()
		if _, err := a.record(common.Event{Type: common.ToolCalled, TurnID: t.id, Tool: &d}); err != nil {
			a.parent.Logger().Print(err)
		}
		// A refused operation has no worker, effects, timer, or job.
		if err := a.returned(t.id, op.Await(context.Background())); err != nil {
			a.parent.Logger().Print(err)
		}
		t.index++
	}
}
func (a *agent) interrupt(err error) {
	t := a.active
	if t == nil {
		return
	}
	if t.httpCancel != nil {
		t.httpCancel()
	}
	if t.observeCancel != nil {
		t.observeCancel()
	}
	if t.operation != nil {
		d := t.operation.Called()
		d.IsError = true
		d.Parts = common.PartList{common.Text("observation interrupted; dispatched work may still be running")}
		if recordErr := a.returned(t.id, d); recordErr != nil {
			err = errors.Join(err, recordErr)
		}
		t.operation = nil
		t.index++
	}
	a.skipTail("interrupted, not executed: turn ended before dispatch")
	if _, recordErr := a.record(common.Event{Type: common.TurnInterrupted, TurnID: t.id, Error: &common.ErrorData{Message: err.Error()}}); recordErr != nil {
		err = errors.Join(err, recordErr)
	}
	a.transition(common.Interrupted)
	a.finish("", err)
}

// closeTerminalAdmission runs on the actor before any terminal result escapes.
func (a *agent) closeTerminalAdmission() {
	a.stopping = true
	a.box.mu.Lock()
	if !a.box.closed {
		a.box.closed = true
		a.box.queue = append(a.box.queue, stop{})
		select {
		case a.box.wake <- struct{}{}:
		default:
		}
	}
	a.box.mu.Unlock()
}

type terminalCompletion struct {
	turn      *turn
	text      string
	err       error
	recovered bool
}

func (a *agent) finish(text string, err error) {
	t := a.active
	if t == nil {
		return
	}
	if err == nil && !t.request.accepted && a.Config().ParentID != "" {
		if t.wakes < 3 {
			t.wakes++
			_, wakeErr := a.record(common.Event{Type: common.HintReceived, TurnID: t.id, Message: &common.MessageData{Actor: common.Human, Parts: common.PartList{common.Text("This delegated task is not complete. Call submit with a result conforming to its advertised schema. Corrective wake " + strconv.Itoa(t.wakes) + " of 3.")}}})
			if wakeErr == nil {
				a.send()
				return
			}
			err = wakeErr
		} else {
			err = errors.New("delegated task exhausted three corrective wakes without submit")
		}
	}
	t.cancel()
	if err != nil {
		_, recordErr := a.record(common.Event{Type: common.ErrorOccurred, TurnID: t.id, Error: &common.ErrorData{Message: err.Error()}})
		err = errors.Join(err, recordErr)
	}
	_, recordErr := a.record(common.Event{Type: common.TurnEnded, TurnID: t.id, Message: &common.MessageData{Actor: common.Assistant, Parts: common.PartList{common.Text(text)}}})
	err = errors.Join(err, recordErr)
	a.active = nil
	a.gate = nil
	a.publish(common.TurnEndedObservation{TurnID: t.id, Text: text})
	if t.request.accepted || (a.Config().ParentID != "" && err != nil) {
		a.closeTerminalAdmission()
		a.terminal = &terminalCompletion{turn: t, text: text, err: err}
		return
	}
	a.transition(common.Idle)
	a.reply(t.request, text, err)
	for _, r := range t.followers {
		a.reply(r, text, err)
	}
}
func (a *agent) reply(r *request, text string, err error) {
	if r.reply != nil {
		r.reply <- common.Result{TurnID: r.turnID, Text: text, Submission: append(json.RawMessage(nil), r.submission...), Err: err}
		close(r.reply)
		close(r.done)
	}
}
func (a *agent) record(e common.Event) (common.Seq, error)      { return a.recordWithSave(e, false) }
func (a *agent) recordSaved(e common.Event) (common.Seq, error) { return a.recordWithSave(e, true) }
func (a *agent) recordWithSave(e common.Event, saveFirst bool) (common.Seq, error) {
	if e.Message != nil {
		for i := range e.Message.Parts {
			if e.Message.Parts[i].DisplayID == 0 {
				a.partID++
				e.Message.Parts[i].DisplayID = a.partID
			}
		}
	}
	var err error
	if saveFirst {
		err = a.history.AppendSaved(e, a.saveDump)
	} else {
		err = a.history.Append(e)
	}
	if err != nil {
		return 0, err
	}
	seq := a.history.LastSeq()
	// AppendSaved already persisted this exact event. A second save must not
	// introduce a post-commit failure that aborts an admitted reservation.
	if !saveFirst {
		if err := a.save(); err != nil {
			return seq, err
		}
	}
	e.Seq = seq
	a.viewState.Through = seq
	a.publish(common.HistoryCommitted{Event: e})
	return seq, nil
}
func (a *agent) save() error { return a.saveDump(a.history.Dump) }
func (a *agent) saveDump(dump func(io.Writer) error) error {
	path := a.config.LogPath
	if path == "" {
		return nil
	}
	f, err := a.Sandbox().CreateTemp(filepath.Dir(path), ".ensemble-log-")
	if err != nil {
		return fmt.Errorf("saving log: %w", err)
	}
	name := f.Name()
	defer a.Sandbox().Remove(name)
	err = dump(f)
	err = errors.Join(err, f.Close())
	if err == nil {
		err = a.Sandbox().Rename(name, path)
	}
	return err
}
func (a *agent) transition(to common.TurnState) {
	if a.state == to {
		return
	}
	from := a.state
	a.state = to
	a.publish(common.StateChanged{From: string(from), To: string(to)})
}
func (a *agent) final(seq common.Seq, p common.Part) {
	if p.DisplayID == 0 {
		a.partID++
		p.DisplayID = a.partID
	}
	turnID := uint64(0)
	if a.active != nil {
		turnID = a.active.id
	}
	a.publish(common.PartFinal{Seq: uint64(seq), TurnID: turnID, PartID: p.DisplayID, Part: p})
}

// Each observer gets a bounded queue. Neither a slow callback nor a slow Wait
// predicate can block the actor; neither queue is a completion channel.
func (a *agent) subscribe() (uint64, chan common.Observation) {
	a.observersMu.Lock()
	defer a.observersMu.Unlock()
	a.nextObserver++
	id := a.nextObserver
	ch := make(chan common.Observation, 64)
	a.observers[id] = ch
	return id, ch
}
func (a *agent) unsubscribe(id uint64) {
	a.observersMu.Lock()
	delete(a.observers, id)
	a.observersMu.Unlock()
}
func (a *agent) publish(o common.Observation) {
	a.updateView(o)
	a.parent.Publish(a, o)
	a.observersMu.Lock()
	defer a.observersMu.Unlock()
	for q := range a.queues {
		q.push(o)
	}
	for _, ch := range a.observers {
		select {
		case ch <- o:
		default:
		}
	}
}
func (a *agent) Subscribe(o common.Observer) {
	if o == nil {
		return
	}
	id, ch := a.subscribe()
	go func() {
		defer a.unsubscribe(id)
		for {
			select {
			case v := <-ch:
				o.Observe(v)
			case <-a.done:
				return
			}
		}
	}()
}
func (a *agent) Wait(ctx context.Context, predicate func(common.Observation) bool) (common.Observation, error) {
	id, ch := a.subscribe()
	defer a.unsubscribe(id)
	for {
		select {
		case o := <-ch:
			if predicate == nil || predicate(o) {
				return o, nil
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-a.done:
			return nil, common.ErrStopped
		}
	}
}

func (a *agent) isPaused() bool { a.gateMu.Lock(); defer a.gateMu.Unlock(); return a.paused }
func (a *agent) Pause() error {
	a.gateMu.Lock()
	a.paused = true
	a.gateMu.Unlock()
	return a.box.post(pauseChanged{}, true)
}
func (a *agent) Unpause() error {
	a.gateMu.Lock()
	a.paused = false
	a.gateMu.Unlock()
	return a.box.post(pauseChanged{}, true)
}

// A pause stores a continuation, never waits inside the mailbox.
func (a *agent) waitGate(reason string, resume func()) bool {
	if !a.isPaused() {
		return false
	}
	a.gate = resume
	a.gateReason = reason
	a.publish(common.Paused{Paused: true, Reason: reason})
	a.requestPauseSpeech()
	return true
}
func (a *agent) requestPauseSpeech() {
	t := a.active
	if a.speaking && t != nil && !t.speechRequested {
		t.speechRequested = true
		a.publish(common.SpeechRequested{TurnID: t.id, Parts: common.PartList{common.Text(a.gateReason)}, Reason: a.gateReason})
	}
}
func (a *agent) displayID(t *turn, parserID uint64) uint64 {
	if t.partIDs == nil {
		t.partIDs = make(map[uint64]uint64)
	}
	if id := t.partIDs[parserID]; id != 0 {
		return id
	}
	a.partID++
	t.partIDs[parserID] = a.partID
	return a.partID
}

// RestoreSubmission is construction-only; a restored terminal actor never opens
// public admission and never sends a provider request.
func (a *agent) RestoreSubmission(payload json.RawMessage) {
	a.stopping = true
	a.terminal = &terminalCompletion{turn: &turn{request: &request{accepted: true, submission: append(json.RawMessage(nil), payload...)}}, recovered: true}
	a.box.closed = true
	_ = a.box.post(stop{}, false)
}
