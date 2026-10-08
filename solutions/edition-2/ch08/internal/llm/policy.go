package llm

import (
	"encoding/json"
	"example.com/ensemble/internal/common"
)

func (a *Actor) UpdatePolicy(base uint64, patch json.RawMessage) (common.PolicyAck, error) {
	reply, err := a.ask(common.ActorMessage{Kind: "policy_update", BaseRevision: base, Patch: append(json.RawMessage(nil), patch...)})
	if _, stopped := err.(common.StoppedError); stopped {
		err = &common.SettingsError{Code: "settings_closed", Message: "Agent settings closed"}
	}
	return reply.PolicyAck, err
}
func (a *Actor) receivePolicy(m common.ActorMessage) bool {
	switch m.Kind {
	case "policy_update":
		if a.stopping {
			m.Reply <- common.ActorReply{Error: &common.SettingsError{Code: "settings_closed", Message: "Agent settings closed"}}
			return true
		}
		next, changed, err := a.parent.Policy().Prepare(m.BaseRevision, m.Patch)
		if err != nil || !changed {
			m.Reply <- common.ActorReply{PolicyAck: common.PolicyAck{Revision: next.Revision, WatchRevision: a.revision}, Error: err}
			return true
		}
		a.workers.Add(1)
		go func() {
			defer a.workers.Done()
			err := a.parent.Policy().Persist(next)
			_ = a.enqueue(common.ActorMessage{Kind: "policy_written", Policy: next, Error: err, Reply: m.Reply}, false)
		}()
		return true
	case "policy_written":
		a.parent.Policy().Apply(m.Policy, m.Error)
		if m.Error == nil {
			value := m.Policy
			a.publish(common.Observation{Kind: "policy_changed", AgentID: a.parent.ID(), ExecutionPolicy: &value})
		}
		current := a.parent.Policy().Snapshot()
		m.Reply <- common.ActorReply{PolicyAck: common.PolicyAck{Revision: current.Revision, WatchRevision: a.revision}, Error: m.Error}
		return true
	}
	return false
}
