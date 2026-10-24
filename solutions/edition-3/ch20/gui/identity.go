package gui

import (
	"fmt"

	"ensemble"
)

// displayIDs belongs to one GUI Agent view. It assigns identities to occurrences,
// not parser IDs or native call IDs. Reserved native IDs are considered before
// legacy fallbacks, and all fallback IDs fit JavaScript's exact integer range.
// No canonical event is changed. Reconnect uses the same occurrence mapping.
type displayIDs struct {
	next     uint64
	reserved map[uint64]bool
	used     map[uint64]bool
	frames   map[ensemble.Seq][]displayIdentity
	calls    map[string]uint64
	partials map[uint64]uint64
}
type displayIdentity struct{ part, call uint64 }

func newDisplayIDs(events []ensemble.Event) *displayIDs {
	d := &displayIDs{next: 1<<53 - 1, reserved: map[uint64]bool{}, used: map[uint64]bool{}, frames: map[ensemble.Seq][]displayIdentity{}, calls: map[string]uint64{}, partials: map[uint64]uint64{}}
	for _, e := range events {
		for _, f := range projectEvent("", e) {
			d.reserved[f.PartID] = true
			if f.Part != nil {
				d.reserved[f.Part.CallPartID] = true
			}
		}
	}
	for _, e := range events {
		d.project("", e)
	}
	return d
}
func (d *displayIDs) allocate(native uint64) uint64 {
	if native != 0 && native < 1<<53 && !d.used[native] {
		d.used[native] = true
		return native
	}
	for d.reserved[d.next] || d.used[d.next] {
		d.next--
	}
	id := d.next
	d.next--
	d.used[id] = true
	return id
}
func (d *displayIDs) partial(native uint64) uint64 {
	if id := d.partials[native]; id != 0 {
		return id
	}
	id := d.allocate(native)
	d.partials[native] = id
	return id
}
func (d *displayIDs) project(agent string, e ensemble.Event) []Wire {
	frames := projectEvent(agent, e)
	ids, known := d.frames[e.Seq]
	if !known {
		for _, f := range frames {
			var id displayIdentity
			if f.Part != nil {
				key := fmt.Sprintf("%d/%s", f.TurnID, f.Part.CallID)
				id.call = d.calls[key]
				lifecycle := e.Type == ensemble.ToolCalled || e.Type == ensemble.ToolReleased || e.Type == ensemble.ToolReturned || e.Type == ensemble.JobReported || e.Type == ensemble.JobKilled
				if lifecycle && id.call == 0 && f.TurnID != 0 {
					// Snapshot entries have no TurnID. Bind the first retained
					// lifecycle event to that already displayed prefix call, then
					// retain its actual turn key for results and late job updates.
					id.call = d.calls[fmt.Sprintf("0/%s", f.Part.CallID)]
					if id.call != 0 {
						d.calls[key] = id.call
					}
				}
				if (e.Type == ensemble.ToolCalled || e.Type == ensemble.ToolReleased || e.Type == ensemble.JobReported || e.Type == ensemble.JobKilled) && id.call != 0 {
					id.part = id.call
				} else if partial := d.partials[f.PartID]; partial != 0 {
					id.part = partial
					delete(d.partials, f.PartID)
				} else {
					id.part = d.allocate(f.PartID)
				}
				if f.Part.Kind == "tool_call" {
					d.calls[key] = id.part
					id.call = id.part
				}
			}
			ids = append(ids, id)
		}
		d.frames[e.Seq] = ids
	}
	for i := range frames {
		frames[i].PartID = ids[i].part
		if frames[i].Part != nil {
			frames[i].Part.CallPartID = ids[i].call
			if frames[i].Part.Job != nil {
				job := *frames[i].Part.Job
				job.CallPartID = ids[i].call
				frames[i].Part.Job = &job
			}
		}
	}
	return frames
}

// contextEvents is display-only input when the archive omitted its historical
// log. Sequence and parts remain the saved canonical values; no synthetic event
// is appended to History or dispatched to tools.
func contextEvents(a ensemble.Agent) []ensemble.Event {
	events := a.Log().Events()
	first := a.Log().LastSeq() + 1
	if len(events) > 0 {
		first = events[0].Seq
	}
	var prefix []ensemble.Event
	for _, entry := range a.Log().Context().Dialogue {
		if entry.Seq >= first {
			break
		}
		prefix = append(prefix, ensemble.Event{Seq: entry.Seq, Type: ensemble.MessageReceived, Message: &ensemble.MessageData{Actor: entry.Actor, Parts: entry.Parts}})
	}
	return append(prefix, events...)
}
