package llm

import (
	"context"
	"example.com/ensemble/internal/common"
)

func (a *Actor) Session() (*common.SessionState, error) {
	r, e := a.ask(common.ActorMessage{Kind: "session"})
	return r.Session, e
}
func (a *Actor) Checkpoint(save bool) (common.ActorReply, error) {
	return a.CheckpointContext(context.Background(), save)
}
func (a *Actor) CheckpointContext(ctx context.Context, save bool) (common.ActorReply, error) {
	if err := ctx.Err(); err != nil {
		return common.ActorReply{}, err
	}
	m := common.ActorMessage{Kind: "checkpoint", Save: save, Reply: make(chan common.ActorReply, 1)}
	if err := a.enqueue(m, true); err != nil {
		return common.ActorReply{}, err
	}
	select {
	case reply := <-m.Reply:
		return reply, reply.Error
	case <-ctx.Done():
		return common.ActorReply{}, ctx.Err()
	}
}
func (a *Actor) receiveSession(m common.ActorMessage) bool {
	switch m.Kind {
	case "session":
		m.Reply <- common.ActorReply{Session: a.parent.SessionState()}
		return true
	case "checkpoint":
		a.startCheckpoint(m)
		return true
	case "checkpoint_done":
		a.parent.ApplyCheckpoint(m.Checkpoint)
		a.checkpointBusy = false
		if m.Checkpoint.Saved {
			a.publish(common.Observation{Kind: "session_changed", AgentID: a.parent.ID(), Session: a.parent.SessionState()})
		}
		if m.Reply != nil {
			m.Reply <- common.ActorReply{CheckpointAck: common.CheckpointAck{AsOf: m.Checkpoint.Export.AsOf, WatchRevision: a.revision}, CheckpointExport: m.Checkpoint.Export, Error: m.Checkpoint.Error}
		}
		if m.Text == "final_checkpoint" {
			err := m.Error
			if err == nil {
				err = m.Checkpoint.Error
			}
			_ = a.enqueue(common.ActorMessage{Kind: "closed", Error: err}, false)
		}
		return true
	}
	return false
}
func (a *Actor) startCheckpoint(m common.ActorMessage) {
	fail := func(err error) {
		if m.Reply != nil {
			m.Reply <- common.ActorReply{Error: err}
		}
		if m.Kind == "final_checkpoint" {
			_ = a.enqueue(common.ActorMessage{Kind: "closed", Error: err}, false)
		}
	}
	if a.parent.SessionState() == nil {
		fail(&common.SessionError{Code: "session_conflict", Detail: "standalone Agent has no session"})
		return
	}
	a.mu.Lock()
	queued := len(a.pending) > 0
	for _, q := range a.queue {
		if q.Kind == "prompt" {
			queued = true
		}
	}
	cursor := a.next
	a.mu.Unlock()
	if a.transientWorkers.Load() != 0 || a.checkpointBusy || a.active != nil || queued || a.model != nil || !Settled(a.parent.Engine(), a.parent.TurnSnapshot()) {
		fail(&common.SessionError{Code: "session_busy", Detail: "checkpoint requires a settled idle boundary"})
		return
	}
	cp, err := a.parent.CaptureSession(cursor)
	if err != nil {
		fail(err)
		return
	}
	done, err := a.parent.BeginCheckpoint(cp, m.Save)
	if err != nil {
		fail(err)
		return
	}
	a.checkpointBusy = true
	a.workers.Add(1)
	go func() {
		defer a.workers.Done()
		result := <-done
		_ = a.enqueue(common.ActorMessage{Kind: "checkpoint_done", Checkpoint: result, Reply: m.Reply, Text: m.Kind, Error: m.Error}, false)
	}()
}
func (a *Actor) requestIndex(r *request) uint64 {
	if r.config.DataDir != "" {
		return r.ordinal
	}
	return 0
}
func (a *Actor) resolveLimits(p common.Part) (common.Limits, string, error) {
	if a.parent.Config().DataDir == "" {
		return a.parent.Registry().ResolveLimits(p)
	}
	pending := a.parent.Jobs().PendingLimits()
	if pending != nil {
		if err := a.record(common.Event{Type: "tool_limits_consumed", Limits: &common.LimitsEvent{CallID: p.CallID, Name: p.Name, Overrides: *pending}}); err != nil {
			return common.Limits{}, "", err
		}
	}
	return a.parent.Registry().ResolveConsumed(p, pending)
}
