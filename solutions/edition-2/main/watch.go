package ensemble

import (
	"fmt"
	"sort"

	"example.com/ensemble/internal/common"
	"example.com/ensemble/internal/llm"
)

type WatchSnapshot = common.WatchSnapshot
type Watch = common.Watch
type WatchRecord = common.WatchRecord
type WatchState = common.WatchState
type WatchPartial = common.WatchPartial
type ActiveOperation = common.ActiveOperation
type PauseRegistration = common.PauseRegistration
type PauseState = common.PauseState

const WatchItems = common.WatchItems
const WatchBytes = common.WatchBytes

func (a *Agent) Watch() (WatchSnapshot, Watch, error) {
	if a.actor == nil {
		return WatchSnapshot{}, nil, fmt.Errorf("Agent is read-only")
	}
	return a.actor.Watch()
}
func (a *Agent) RegisterPause() (PauseRegistration, error) {
	if a.actor == nil {
		return nil, fmt.Errorf("Agent is read-only")
	}
	return a.actor.RegisterPause()
}
func (e *Ensemble) Watch(id string) (WatchSnapshot, Watch, error) {
	a, err := e.Agent(id)
	if err != nil {
		return WatchSnapshot{}, nil, err
	}
	return a.Watch()
}
func (e *Ensemble) RegisterPause(id string) (PauseRegistration, error) {
	a, err := e.Agent(id)
	if err != nil {
		return nil, err
	}
	return a.RegisterPause()
}

// WatchSource copies selected immutable records while holding their owner's lock.
// Actor fills scheduling fields at the same serialized boundary.
func (a turnAgent) WatchSource() common.WatchSnapshot {
	a.mu.Lock()
	defer a.mu.Unlock()
	s := common.WatchSnapshot{LogSeq: a.context.LastSeq, Events: []Event{}, Partials: []common.WatchPartial{}, State: common.WatchState{Model: a.config.Model, Usage: []common.UsageAccount{}}}
	if a.skills != nil {
		s.State.Skills = a.skills.Inspect().State
	}
	selected := []Event{}
	for _, e := range a.events {
		if llm.RenderableEvent(a.engine, e.Type) {
			selected = append(selected, e)
		}
	}
	if len(selected) > 100 {
		s.Omitted = len(selected) - 100
		selected = selected[s.Omitted:]
	}
	s.Events, _ = llm.Clone(a.engine, selected)
	if len(s.Events) > 0 {
		first, last := s.Events[0].Seq, s.Events[len(s.Events)-1].Seq
		s.FirstSeq = &first
		s.LastSeq = &last
	}
	for from, usage := range a.engine.UsageByModel() {
		s.State.Usage = append(s.State.Usage, common.UsageAccount{From: from, Usage: usage})
	}
	sort.Slice(s.State.Usage, func(i, j int) bool {
		x, y := s.State.Usage[i].From, s.State.Usage[j].From
		return x.Vendor+"/"+x.Model+"/"+x.Surface < y.Vendor+"/"+y.Model+"/"+y.Surface
	})
	return s
}
