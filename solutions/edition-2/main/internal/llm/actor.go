package llm

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"unicode/utf8"

	"example.com/ensemble/internal/common"
)

// Actor owns turn decisions. Its mailbox mutex only guards admission and transfer;
// it never spans persistence, model I/O, job waits, or display callbacks.
type Actor struct {
	checkpointBusy                 bool
	finalCheckpoint                bool
	revision, nextWatch, nextPause uint64
	watches                        []*watch
	pauses                         map[uint64][2]bool
	partials                       map[int]map[string][]byte
	projectionInvalid              bool
	partialBytes                   int
	parent                         common.ActorAgent
	mu                             sync.Mutex
	queue                          []common.ActorMessage
	wake                           chan struct{}
	done                           chan struct{}
	accepting                      bool
	next                           uint64
	pending                        []*request
	active                         *request
	operation                      uint64
	model                          common.ModelOperation
	cancel                         context.CancelFunc
	workers                        sync.WaitGroup
	transientWorkers               atomic.Int64
	state                          string
	stopping                       bool
	closing                        bool
	fault                          error
	closeErr                       error
	faulted                        *request
}
type request struct {
	ordinal    uint64
	parent     common.Actor
	id, text   string
	done       chan struct{}
	completion common.Completion
	config     common.Config
	rounds     int
	policy     common.TurnPolicy
	parts      []common.Part
	calls      []common.Part
	index      int
	report     *common.ReportTask
	stopReason string
	ending     string
}

func NewActor(parent common.ActorAgent) *Actor {
	a := &Actor{pauses: map[uint64][2]bool{}, parent: parent, wake: make(chan struct{}, 1), done: make(chan struct{}), accepting: true, state: "idle", next: parent.TurnSnapshot().RequestCursor}
	go a.run()
	return a
}
func (a *Actor) Agent() common.ActorAgent { return a.parent }
func (r *request) ID() string             { return r.id }
func (r *request) Done() <-chan struct{}  { return r.done }
func (r *request) Cancel() error          { return r.parent.Cancel(r.id) }
func (r *request) Wait(ctx context.Context) (common.Completion, error) {
	select {
	case <-r.done:
		c, err := Clone(r.parent.Agent().Engine(), r.completion)
		return c, err
	case <-ctx.Done():
		return common.Completion{}, ctx.Err()
	}
}
func (a *Actor) enqueue(m common.ActorMessage, public bool) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if public && !a.accepting {
		return common.StoppedError{}
	}
	a.queue = append(a.queue, m)
	select {
	case a.wake <- struct{}{}:
	default:
	}
	return nil
}
func (a *Actor) Submit(text string) (common.RequestHandle, error) {
	if strings.TrimSpace(text) == "" || !utf8.ValidString(text) {
		return nil, fmt.Errorf("question must be nonempty")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.accepting {
		return nil, common.StoppedError{}
	}
	seen := a.parent.TurnSnapshot().TurnIDs
	for {
		if a.next == ^uint64(0) {
			return nil, &common.SessionError{Code: "session_limit", Detail: "request identities exhausted"}
		}
		a.next++
		if !seen[fmt.Sprintf("r%d", a.next)] {
			break
		}
	}
	r := &request{ordinal: a.next, parent: a, id: fmt.Sprintf("r%d", a.next), text: text, done: make(chan struct{})}
	a.queue = append(a.queue, common.ActorMessage{Kind: "prompt", Handle: r})
	select {
	case a.wake <- struct{}{}:
	default:
	}
	return r, nil
}
func (a *Actor) control(m common.ActorMessage) (common.ControlAck, error) {
	m.Reply = make(chan common.ActorReply, 1)
	if err := a.enqueue(m, true); err != nil {
		return common.ControlAck{}, err
	}
	reply := <-m.Reply
	return reply.Ack, reply.Error
}
func (a *Actor) Hint(text string) (common.ControlAck, error) {
	return a.control(common.ActorMessage{Kind: "hint", Text: text})
}
func (a *Actor) Interrupt() (common.ControlAck, error) {
	return a.control(common.ActorMessage{Kind: "interrupt"})
}
func (a *Actor) Cancel(id string) error {
	_, err := a.control(common.ActorMessage{Kind: "cancel", RequestID: id})
	return err
}
func (a *Actor) Append(e common.Event) error {
	_, err := a.control(common.ActorMessage{Kind: "append", Event: e})
	return err
}

// Job facts are posted while Jobs may hold its own mutex. Posting never waits for
// actor progress; the actor owns persistence failures and initiates cleanup.
func (a *Actor) Record(e common.Event) error {
	return a.enqueue(common.ActorMessage{Kind: "job", Event: e}, false)
}
func (a *Actor) Close() error {
	a.mu.Lock()
	if a.accepting {
		a.accepting = false
		a.queue = append(a.queue, common.ActorMessage{Kind: "close"})
		select {
		case a.wake <- struct{}{}:
		default:
		}
	}
	a.mu.Unlock()
	<-a.done
	return a.closeErr
}
func (a *Actor) transition(state string) {
	if state == a.state {
		return
	}
	old := a.state
	a.state = state
	id := ""
	if a.active != nil {
		id = a.active.id
	}
	a.publish(common.Observation{AgentID: a.parent.ID(), RequestID: id, Kind: "state", OldState: old, State: state})
}
func (a *Actor) run() {
	defer close(a.done)
	for {
		<-a.wake
		for {
			a.mu.Lock()
			if len(a.queue) == 0 {
				a.mu.Unlock()
				break
			}
			m := a.queue[0]
			a.queue[0] = common.ActorMessage{}
			a.queue = a.queue[1:]
			a.mu.Unlock()
			if a.receive(m) {
				return
			}
		}
		if a.active == nil && !a.stopping && len(a.pending) > 0 {
			a.activate()
		}
		if a.stopping && a.active == nil && !a.closing {
			a.shutdown()
		}
	}
}
func (a *Actor) receive(m common.ActorMessage) bool {
	if a.receiveSession(m) {
		return false
	}
	if a.receiveSkills(m) {
		return false
	}
	if a.receivePolicy(m) {
		return false
	}
	if a.receiveWatch(m) {
		return false
	}
	var ack common.ControlAck
	var err error
	switch m.Kind {
	case "resume_tools":
		if a.active != nil && a.state == "tools_pending" && a.active.report == nil {
			a.dispatch()
		}
	case "prompt":
		r := m.Handle.(*request)
		if a.stopping {
			a.complete(r, "stopped", common.StoppedError{})
		} else {
			a.pending = append(a.pending, r)
			if a.active == nil {
				a.activate()
			}
		}
	case "hint":
		if a.active == nil || strings.TrimSpace(m.Text) == "" || !utf8.ValidString(m.Text) {
			err = fmt.Errorf("hint requires nonempty text and an active turn")
			break
		}
		err = a.record(common.Event{Type: "hint_received", Hint: &common.HintEvent{RequestID: a.active.id, Text: m.Text}})
		if err == nil {
			ack = common.ControlAck{RequestID: a.active.id, Seq: a.parent.TurnSnapshot().LastSeq}
		}
	case "interrupt":
		if a.active != nil {
			ack.RequestID = a.active.id
			ack.Interrupted = true
			a.interrupt("interrupted")
		}
	case "cancel":
		if a.active != nil && a.active.id == m.RequestID {
			a.interrupt("canceled")
		} else {
			for i, r := range a.pending {
				if r.id == m.RequestID {
					a.pending = append(a.pending[:i], a.pending[i+1:]...)
					a.complete(r, "canceled", fmt.Errorf("request canceled before activation"))
					break
				}
			}
		}
	case "append":
		if a.active != nil && m.Event.Skills == nil {
			err = fmt.Errorf("Agent busy")
		} else {
			err = a.parent.RecordTurn(m.Event)
		}
	case "job":
		err = a.recordJob(m.Event)
	case "fragments":
		if a.model != m.Model {
			m.Model.Discard()
			break
		}
		fragments, more := m.Model.Drain()
		for _, f := range fragments {
			a.observeModel(common.Observation{Kind: "part_delta", PartID: f.PartID, Channel: f.Channel, Text: f.Text})
		}
		if more {
			a.ModelReady(m.Model)
		}
	case "model":
		if a.active == nil || m.Operation != a.operation {
			a.parent.Ensemble().Logf("discarded stale model operation")
			break
		}
		a.cancel = nil
		if m.Error != nil {
			a.endModel(false, 0, "error", m.Error.Error())
			a.failTurn("error", m.Error)
			break
		}
		if err = a.parent.RecordResponse(m.Response); err != nil {
			a.persistence(err)
			break
		}
		c := a.parent.TurnSnapshot()
		entry := c.Entries[len(c.Entries)-1]
		a.active.parts = entry.Parts
		a.active.stopReason = m.Response.Response.StopReason
		a.active.calls = nil
		a.active.index = 0
		for i, p := range entry.Parts {
			copy, _ := Clone(a.parent.Engine(), p)
			partID := i + 1
			if len(m.Response.PartIDs) > i {
				partID = m.Response.PartIDs[i]
			}
			a.observeModel(common.Observation{Kind: "part_final", Seq: entry.Seq, ResponseSeq: entry.Seq, Position: i, PartIndex: i, PartID: partID, Part: &copy})
			if p.Type == "tool_call" {
				a.active.calls = append(a.active.calls, p)
			}
		}
		a.endModel(true, entry.Seq, "", "")
		if len(a.active.calls) == 0 {
			a.finish("success", nil)
		} else {
			a.dispatch()
		}
	case "report":
		if a.active == nil || m.Operation != a.operation {
			break
		}
		a.cancel = nil
		if m.Error != nil {
			a.persistence(m.Error)
			break
		}
		if err = a.acceptReport(m.Report); err != nil {
			a.persistence(err)
			break
		}
		a.active.report = nil
		a.active.index++
		if a.active.ending != "" {
			a.finishInterrupted(a.active.ending)
		} else {
			a.dispatch()
		}
	case "fault":
		a.persistence(m.Error)
	case "close":
		a.stopping = true
		a.transition("stopping")
		for _, r := range a.pending {
			a.complete(r, "stopped", common.StoppedError{})
		}
		a.pending = nil
		if a.active != nil {
			a.interrupt("stopped")
		}
	case "closed":
		if a.parent.SessionState() != nil && a.fault == nil && !a.finalCheckpoint {
			a.finalCheckpoint = true
			a.startCheckpoint(common.ActorMessage{Kind: "final_checkpoint", Save: true, Error: m.Error})
			return false
		}
		for _, w := range a.watches {
			w.stop("Agent closed")
		}
		a.watches = nil
		a.pauses = nil
		if a.faulted != nil {
			a.complete(a.faulted, "error", fmt.Errorf("persistence failure; completion could not be recorded; inspect actual effects before retry"))
			a.faulted = nil
		}
		if a.fault != nil {
			a.closeErr = a.fault
		} else {
			a.closeErr = m.Error
		}
		if closeErr := a.parent.FinishClose(); a.closeErr == nil {
			a.closeErr = closeErr
		}
		return true
	}
	if m.Reply != nil {
		m.Reply <- common.ActorReply{Ack: ack, Error: err}
	}
	return false
}
func (a *Actor) record(e common.Event) error {
	err := a.parent.RecordTurn(e)
	if err != nil {
		a.persistence(err)
	}
	return err
}
func (a *Actor) recordJob(e common.Event) error {
	if e.Job != nil {
		old := a.parent.TurnSnapshot().Jobs[e.Job.Handle]
		if old.Status == e.Job.Status && old.Status != "running" {
			return nil
		}
	}
	return a.record(e)
}
func (a *Actor) activate() {
	// Activation and Close admission share a linearization point. A prompt still
	// queued when public admission closes cannot start in the gap before the
	// actor processes its close message.
	a.mu.Lock()
	if !a.accepting {
		a.mu.Unlock()
		return
	}
	r := a.pending[0]
	a.pending = a.pending[1:]
	a.active = r
	a.mu.Unlock()
	if err := a.parent.BeginTurn(); err != nil {
		a.active = nil
		a.complete(r, "error", err)
		a.wakeSelf()
		return
	}
	p := a.parent.Policy().Snapshot()
	r.policy = common.TurnPolicy{Revision: p.Revision, MaxModelRequests: p.MaxModelRequests, EffectiveMaxModelRequests: p.EffectiveMaxModelRequests}
	r.config = a.parent.Config()
	r.config.Tools = a.parent.Registry().Declarations()
	a.transition("input_pending")
	if a.record(common.Event{Type: "turn_started", Turn: &common.TurnEvent{RequestID: r.id, Policy: &r.policy, RequestIndex: a.requestIndex(r)}}) != nil {
		return
	}
	if a.record(common.Event{Type: "message_received", Message: &common.Entry{Actor: "human", Purpose: "dialogue", Parts: []common.Part{Text(r.text)}}}) != nil {
		return
	}
	a.exchange()
}
func (a *Actor) wakeSelf() {
	select {
	case a.wake <- struct{}{}:
	default:
	}
}
func (a *Actor) exchange() {
	r := a.active
	if r.rounds >= r.policy.EffectiveMaxModelRequests {
		a.failTurn("round_limit", fmt.Errorf("round_limit: %d model requests completed; final tool batch retained", r.policy.EffectiveMaxModelRequests))
		return
	}
	c := a.parent.TurnSnapshot()
	r.config.Tools = a.parent.Registry().Declarations()
	if c.SkillMode {
		r.config.System = ""
	}
	body, err := Render(a.parent.Engine(), c, r.config)
	if err != nil {
		a.failTurn("error", err)
		return
	}
	ephemera := make([]uint64, 0, len(c.Ephemera))
	hints := make([]uint64, 0, len(c.Hints))
	for _, e := range c.Ephemera {
		ephemera = append(ephemera, e.Seq)
	}
	for _, e := range c.Hints {
		hints = append(hints, e.Seq)
	}
	if a.record(common.Event{Type: "request_sent", Request: &common.RequestEvent{Delivery: delivery(r.config), To: Route(a.parent.Engine(), r.config), Ephemera: ephemera, Hints: hints, Configuration: &common.RequestConfig{System: r.config.System, MaxTokens: r.config.MaxTokens, Tools: r.config.Tools, ResolvedModel: r.config.ResolvedModel}}}) != nil {
		return
	}
	r.rounds++
	a.operation++
	op := a.operation
	a.model = a.parent.Engine().NewOperation(fmt.Sprintf("m%d", op), r.id, r.config)
	model := a.model
	a.observeModel(common.Observation{Kind: "model_begin", Delivery: model.Delivery()})
	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel
	a.transition("in_flight")
	a.workers.Add(1)
	a.transientWorkers.Add(1)
	go func() {
		defer a.workers.Done()
		defer a.transientWorkers.Add(-1)
		defer cancel()
		parsed, err := model.Exchange(ctx, body)
		_ = a.enqueue(common.ActorMessage{Kind: "model", Operation: op, Response: parsed, Error: err}, false)
	}()
}
func (a *Actor) dispatch() {
	r := a.active
	if r.index == len(r.calls) {
		a.exchange()
		return
	}
	a.transition("tools_pending")
	if a.pauseState().Paused {
		return
	}
	p := r.calls[r.index]
	registry := a.parent.Registry()
	manager := a.parent.Jobs()
	limits, note, limitErr := a.resolveLimits(p)
	if a.fault != nil {
		return
	}
	available, supervision := registry.Kind(p.Name)
	management := registry.Management(p.Name)
	var job common.Job
	var err error
	if available && !supervision && !management && limitErr == nil {
		job, err = manager.Create()
		if err != nil {
			a.persistence(err)
			return
		}
	}
	called := common.ToolEvent{CallID: p.CallID, Name: p.Name, Args: p.Args}
	if job != nil {
		snapshot := job.Snapshot()
		called.Job = &snapshot
	}
	if a.record(common.Event{Type: "tool_called", Tool: &called}) != nil {
		if job != nil {
			manager.Abort(job)
		}
		return
	}
	var result *common.ToolEvent
	var task *common.ReportTask
	if limitErr != nil || !available {
		text := "tool is unavailable to this Agent"
		if limitErr != nil {
			text = limitErr.Error()
		}
		result = &common.ToolEvent{CallID: p.CallID, IsError: true, Parts: []common.Part{Text(note + p.Name + " failed: " + text)}}
	} else if management {
		revision := uint64(0)
		if state := a.parent.SkillStateView(); state != nil {
			revision = state.Revision
		}
		op, skillErr := registry.SkillOperation(p, revision)
		var value common.SkillResult
		if skillErr == nil {
			value, skillErr = a.parent.ChangeSkill(op)
		}
		if skillErr != nil {
			if _, typed := skillErr.(*common.SkillError); !typed {
				a.persistence(skillErr)
				return
			}
		}
		ack := registry.SkillAcknowledgement(p, value, skillErr, note)
		result = &ack
	} else if supervision {
		result, task = registry.BeginSupervision(p, limits, note)
	} else {
		manager.Start(job, p)
		task = &common.ReportTask{Job: job, Request: common.JobReport{CallID: p.CallID, Limits: limits, Note: note, Original: true, MatchStart: -1}}
	}
	if result != nil {
		if p.Name == "tool_limits" && !result.IsError && r.config.DataDir != "" {
			value, err := registry.LimitCandidate(p)
			if err != nil {
				a.persistence(err)
				return
			}
			if a.record(common.Event{Type: "tool_limits_set", Limits: &common.LimitsEvent{CallID: p.CallID, Overrides: value}}) != nil {
				return
			}
		}
		if a.record(common.Event{Type: "tool_returned", Tool: result}) != nil {
			return
		}
		r.index++
		a.dispatch()
		return
	}
	r.report = task
	a.operation++
	op := a.operation
	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel
	a.workers.Add(1)
	a.transientWorkers.Add(1)
	go func() {
		defer a.workers.Done()
		defer a.transientWorkers.Add(-1)
		defer cancel()
		if task.Ready != nil {
			select {
			case <-task.Ready:
			case <-ctx.Done():
			}
		}
		request := task.Request
		var inputError error
		if task.Input != nil {
			var accepted int
			request.MatchStart, accepted, inputError = manager.SendContext(ctx, task.Job, *task.Input)
			request.Note += fmt.Sprintf("send_input accepted %d of %d bytes; accepted bytes cannot be rolled back.\n", accepted, len(*task.Input))
			if inputError != nil {
				request.Note += inputError.Error() + "\n"
				request.Limits.Delay = 0
			}
		}
		report, err := manager.PrepareReport(ctx, task.Job, request)
		if inputError != nil {
			report.Event.IsError = true
		}
		_ = a.enqueue(common.ActorMessage{Kind: "report", Operation: op, Report: report, Error: err}, false)
	}()
}
func (a *Actor) acceptReport(p common.PreparedReport) error {
	// A terminal fact can already be queued behind an interrupt. Its attributable
	// snapshot must precede a terminal report; the later queued duplicate is ignored.
	if p.Event.Job != nil && p.Event.Job.Status != "running" {
		c := a.parent.TurnSnapshot()
		if c.Jobs[p.Handle].Status == "running" {
			kind := "job_ended"
			if p.Event.Job.Status == "killed" {
				kind = "job_killed"
			}
			if err := a.recordJob(common.Event{Type: kind, Job: p.Event.Job}); err != nil {
				return err
			}
		}
	}
	// If a terminal fact overtook a prepared running report, refresh without waiting.
	if p.Event.Job != nil {
		current := a.parent.TurnSnapshot().Jobs[p.Handle]
		if current.Status != "running" && p.Event.Job.Status == "running" {
			task := a.active.report
			request := task.Request
			request.Limits.Delay = 0
			var err error
			p, err = a.parent.Jobs().PrepareReport(context.Background(), task.Job, request)
			if err != nil {
				return err
			}
		}
	}
	if err := a.parent.RecordTurn(common.Event{Type: "tool_returned", Tool: &p.Event}); err != nil {
		return err
	}
	return a.parent.Jobs().CommitReport(p)
}
func (a *Actor) interrupt(outcome string) {
	if a.active == nil {
		return
	}
	if a.active.ending != "" {
		return
	}
	a.transition("interrupted")
	a.endModel(false, 0, outcome, "turn "+outcome)
	if a.cancel != nil {
		a.cancel()
		a.cancel = nil
	}
	if a.active.report != nil && a.active.report.Input != nil {
		// Input may already have partial effects. Keep this operation current and
		// accept its settled write/report fact before ending the call or turn.
		// The mailbox remains available while the worker cancels its PTY write.
		a.active.ending = outcome
		return
	}
	a.operation++
	r := a.active
	if a.parent.TurnSnapshot().Active {
		if a.record(common.Event{Type: "error_occurred", Error: &common.EventError{Code: outcome, Message: "turn " + outcome}}) != nil {
			return
		}
	}
	if r.report != nil {
		request := r.report.Request
		request.Limits.Delay = 0
		p, err := a.parent.Jobs().PrepareReport(context.Background(), r.report.Job, request)
		if err == nil {
			err = a.acceptReport(p)
		}
		if err != nil {
			a.persistence(err)
			return
		}
		r.report = nil
		r.index++
	}
	a.finishInterrupted(outcome)
}
func (a *Actor) finishInterrupted(outcome string) {
	r := a.active
	for ; r.index < len(r.calls); r.index++ {
		p := r.calls[r.index]
		_, note, _ := a.resolveLimits(p)
		if a.fault != nil {
			return
		}
		if a.record(common.Event{Type: "tool_called", Tool: &common.ToolEvent{CallID: p.CallID, Name: p.Name, Args: p.Args}}) != nil {
			return
		}
		if a.record(common.Event{Type: "tool_returned", Tool: &common.ToolEvent{CallID: p.CallID, IsError: true, Parts: []common.Part{Text(note + "turn interrupted before execution")}}}) != nil {
			return
		}
	}
	a.finish(outcome, fmt.Errorf("turn %s", outcome))
}
func (a *Actor) failTurn(outcome string, err error) {
	if a.record(common.Event{Type: "error_occurred", Error: &common.EventError{Code: outcome, Message: err.Error()}}) != nil {
		return
	}
	a.finish(outcome, err)
}
func (a *Actor) finish(outcome string, err error) {
	r := a.active
	if a.record(common.Event{Type: "turn_ended", Turn: &common.TurnEvent{RequestID: r.id, Outcome: outcome}}) != nil {
		return
	}
	a.active = nil
	a.parent.EndTurn()
	a.complete(r, outcome, err)
	if a.stopping {
		a.transition("stopping")
	} else {
		a.transition("idle")
	}
	a.wakeSelf()
}
func (a *Actor) complete(r *request, outcome string, err error) {
	c := common.Completion{StopReason: r.stopReason, AgentID: a.parent.ID(), RequestID: r.id, Outcome: outcome, Text: TextAnswer(a.parent.Engine(), r.parts), Parts: r.parts, PendingHints: len(a.parent.TurnSnapshot().Hints), Usage: a.parent.Engine().Usage()}
	if err != nil {
		c.Error = &common.EventError{Code: outcome, Message: err.Error()}
	}
	r.completion = c
	close(r.done)
}
func (a *Actor) persistence(err error) {
	if a.fault != nil {
		return
	}
	a.endModel(false, 0, "persistence", "response acceptance failed")
	a.fault = err
	a.stopping = true
	a.mu.Lock()
	a.accepting = false
	a.mu.Unlock()
	if a.cancel != nil {
		a.cancel()
		a.cancel = nil
	}
	a.operation++
	a.parent.Ensemble().Logf("actor persistence/cleanup failure; final facts may not be durable: %v", err)
	if a.active != nil {
		r := a.active
		a.active = nil
		a.parent.EndTurn()
		// Cleanup joins owned I/O before callers can observe the failed request as
		// complete, including an input writer that may have accepted partial bytes.
		a.faulted = r
	}
	for _, r := range a.pending {
		a.complete(r, "error", fmt.Errorf("persistence failure; completion could not be recorded; inspect actual effects before retry"))
	}
	a.pending = nil
	a.transition("stopping")
	a.wakeSelf()
}
func (a *Actor) shutdown() {
	a.closing = true
	// Workers can post facts while shutdown joins them. The mailbox continues to
	// drain until Jobs' terminal facts and the final closed message are accepted.
	go func() {
		err := a.parent.Jobs().Close()
		a.workers.Wait()
		_ = a.enqueue(common.ActorMessage{Kind: "closed", Error: err}, false)
	}()
}

func (a *Actor) Fault(err error) {
	_ = a.enqueue(common.ActorMessage{Kind: "fault", Error: err}, false)
}

func delivery(config common.Config) string {
	if config.DisableStreaming {
		return "plain"
	}
	return "stream"
}
func (a *Actor) ModelReady(op common.ModelOperation) {
	_ = a.enqueue(common.ActorMessage{Kind: "fragments", Model: op}, false)
}
func (a *Actor) observeModel(o common.Observation) {
	o.AgentID = a.parent.ID()
	o.RequestID = a.model.RequestID()
	o.OperationID = a.model.ID()
	a.publish(o)
}
func (a *Actor) endModel(accepted bool, seq uint64, code, message string) {
	if a.model == nil {
		return
	}
	a.observeModel(common.Observation{Kind: "model_end", Accepted: accepted, ResponseSeq: seq, Code: code, Message: message})
	a.model.Discard()
	a.model = nil
}
