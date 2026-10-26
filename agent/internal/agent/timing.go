package agent

import (
	"sync"
	"time"

	"ensemble/internal/common"
)

// The turn owns this clock. HTTP and report workers only bracket their waits;
// sealing makes late completions harmless even after a newer turn has started.
type turnClock struct {
	mu                  sync.Mutex
	started, last       time.Time
	modelTime, toolTime time.Duration
	models, tools       int
	requests, calls     int
	closed              bool
	phases              []*turnPhase
	phaseTimes          map[string]time.Duration
}

type turnPhase struct{ name string }

func (clock *turnClock) advance(at time.Time) {
	elapsed := at.Sub(clock.last)
	if elapsed < 0 {
		return
	}
	// Attribute each instant once. A provider wait takes priority over any
	// overlapping foreground tool wait; background job lifetimes never enter here.
	if clock.models > 0 {
		clock.modelTime += elapsed
	} else if clock.tools > 0 {
		clock.toolTime += elapsed
	} else {
		name := "unattributed"
		if len(clock.phases) > 0 {
			name = clock.phases[len(clock.phases)-1].name
		}
		if clock.phaseTimes == nil {
			clock.phaseTimes = make(map[string]time.Duration)
		}
		clock.phaseTimes[name] += elapsed
	}
	clock.last = at
}

func (clock *turnClock) start(at time.Time, model bool) {
	// Settle the old state before changing its counters. Reversing this order
	// would charge queueing or a pause to the wait that only just started.
	clock.advance(at)
	if model {
		clock.models++
		clock.requests++
	} else {
		clock.tools++
		clock.calls++
	}
}

func (clock *turnClock) end(at time.Time, model bool) {
	// The completed wait owns the interval up to this boundary, not after it.
	clock.advance(at)
	if model {
		clock.models--
	} else {
		clock.tools--
	}
}

func (clock *turnClock) wait(model bool) func() {
	clock.mu.Lock()
	defer clock.mu.Unlock()
	if clock.closed {
		return func() {}
	}
	clock.start(time.Now(), model)
	ended := false
	return func() {
		clock.mu.Lock()
		defer clock.mu.Unlock()
		// Both repeated completion and completion after interruption are harmless.
		// The captured flag shares the clock lock rather than needing another lock.
		if !ended && !clock.closed {
			clock.end(time.Now(), model)
		}
		ended = true
	}
}

// ModelStarted observes the HTTP exchange, not rendering or credential renewal.
func (clock *turnClock) ModelStarted() func() { return clock.wait(true) }

// ToolStarted observes the report wait, not the continuing job's execution.
func (clock *turnClock) ToolStarted() func() { return clock.wait(false) }

// PhaseStarted shares the wait clock so concurrent work cannot inflate Other.
func (clock *turnClock) PhaseStarted(name string) func() {
	clock.mu.Lock()
	defer clock.mu.Unlock()
	if clock.closed {
		return func() {}
	}
	phase := clock.startPhase(time.Now(), name)
	return func() {
		clock.mu.Lock()
		defer clock.mu.Unlock()
		if !clock.closed {
			clock.endPhase(time.Now(), phase)
		}
	}
}

func (clock *turnClock) startPhase(at time.Time, name string) *turnPhase {
	clock.advance(at)
	phase := &turnPhase{name: name}
	clock.phases = append(clock.phases, phase)
	return phase
}

func (clock *turnClock) endPhase(at time.Time, phase *turnPhase) {
	for index, active := range clock.phases {
		if active == phase {
			clock.advance(at)
			clock.phases = append(clock.phases[:index], clock.phases[index+1:]...)
			return
		}
	}
}

// Actor snapshots remain independently owned; this only measures the existing
// full copy. Do not optimize the operation while measuring its baseline.
func (r *runtime) contextSnapshot() *common.Context {
	defer r.phase("context_copy")()
	return r.owner.history.Context()
}

func (r *runtime) phase(name string) func() {
	clock := r.admitting
	if r.active != nil {
		clock = r.active.timing
	}
	if clock == nil {
		return func() {}
	}
	return clock.PhaseStarted(name)
}

func (clock *turnClock) finish(interrupted bool) *common.TurnTiming {
	clock.mu.Lock()
	defer clock.mu.Unlock()
	return clock.snapshot(time.Now(), interrupted)
}

func (clock *turnClock) snapshot(at time.Time, interrupted bool) *common.TurnTiming {
	// Open waits stop at logical turn completion, even if HTTP cleanup or a
	// detached tool continues. The stored result is final, not a live stopwatch.
	clock.advance(at)
	clock.closed = true
	timing := &common.TurnTiming{
		Started: clock.started.UTC(), TotalMS: at.Sub(clock.started).Milliseconds(),
		ModelMS: clock.modelTime.Milliseconds(), ToolMS: clock.toolTime.Milliseconds(),
		Calls: clock.calls, Requests: clock.requests, Interrupted: interrupted,
	}
	// Round the residual last so displayed millisecond categories always reconcile.
	timing.OtherMS = timing.TotalMS - timing.ModelMS - timing.ToolMS
	timing.OtherPhasesMS = make(map[string]int64, len(clock.phaseTimes)+1)
	remaining := timing.OtherMS
	for name, duration := range clock.phaseTimes {
		milliseconds := duration.Milliseconds()
		timing.OtherPhasesMS[name] = milliseconds
		remaining -= milliseconds
	}
	// Each phase is rounded down independently. Retain the small residual rather
	// than claiming sub-millisecond precision or assigning it to a named operation.
	timing.OtherPhasesMS["unattributed"] += remaining
	return timing
}
