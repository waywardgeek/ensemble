package agent

import (
	"ensemble/internal/common"
)

// Settings acknowledgments follow their recorded content changes, even when idle.
type memorySettings struct {
	value common.MemorySettings
	reply chan error
}

// Running startup and re-enable reread disk so human corrections replace logged
// bytes with the same identity. Offline history replay itself never scans files.
func (r *runtime) syncMemory() error {
	state := r.owner.history.State()
	if r.compaction == nil && (!r.memoryReady || state.CompressionCommitted) {
		// Startup must settle the durable pair before scans or recall can see
		// partly promoted/archived corpus files. Do not abandon a live attempt
		// merely because ordinary settings synchronization runs again.
		outcome, err := r.owner.memory.Recover()
		if err != nil {
			r.memoryReady = false
			return err
		}
		if outcome != nil {
			if err := r.append(common.Event{Type: "compaction_outcome", CompactionOutcome: outcome}); err != nil {
				r.memoryReady = false
				return err
			}
			if _, err := r.owner.memory.Recover(); err != nil {
				r.memoryReady = false
				return err
			}
		}
	}
	for b := common.BandSoul; b <= common.BandSession; b++ {
		disabled := r.owner.memory.Settings(b).Disabled
		wasDisabled := r.memoryDisabled[b]
		if !r.memoryReady {
			wasDisabled = state.DisabledBands[b]
		}
		if !r.memoryReady || wasDisabled != disabled {
			if disabled {
				if err := r.append(common.Event{Type: "band_depopulated", Depopulated: &common.BandDepopulatedData{Band: b, Reason: "disabled"}}); err != nil {
					return err
				}
			} else {
				files, err := r.owner.memory.Scan(b)
				if err != nil {
					return err
				}
				for i := range files {
					if err := r.append(common.Event{Type: "band_populated", Populated: &files[i]}); err != nil {
						return err
					}
				}
			}
		}
		r.memoryDisabled[b] = disabled
	}
	if !r.memoryReady && r.owner.history.State().HasCompression {
		r.owner.Ensemble().Logf("memory: abandoned compressor; source retained until a later checkpoint")
	}
	r.memoryReady = true
	return nil
}

// Occupancy is the prior conversation request, never lifetime usage or the
// small compressor response. All three normalized input categories are disjoint.
func (r *runtime) forceCheckpoint() error {
	if r.owner.spec.Compaction {
		return nil
	}
	c := r.owner.history.State()
	u := c.LastUsage
	window := r.owner.engine.Features().ContextWindow
	input := u.Input + u.CacheRead + u.CacheWrite
	if window == 0 || r.forcingChecked == c.SentThrough {
		return nil
	}
	r.forcingChecked = c.SentThrough
	if input*100 < window*90 {
		r.warned = false
		return nil
	}
	if !r.warned {
		r.warned = true
		if err := r.append(common.Event{Type: "message_received", Message: &common.MessageData{Actor: common.System, Parts: []common.Part{{Type: "text", Text: "Context is at least 90% full. Finish this micro-goal and call micro_handoff now to preserve what you learned."}}}}); err != nil {
			return err
		}
	}
	if input*100 >= window*95 {
		r.owner.forceHandoff.Store(true)
	}
	return nil
}
