package agent

import (
	"reflect"
	"testing"
	"time"
)

func TestTurnClockExclusiveOtherPhases(t *testing.T) {
	started := time.Now()
	at := func(seconds int) time.Time { return started.Add(time.Duration(seconds) * time.Second) }
	clock := &turnClock{started: started, last: started}
	preparation := clock.startPhase(at(1), "request_preparation")
	copying := clock.startPhase(at(2), "context_copy")
	clock.start(at(3), true)
	clock.end(at(5), true)
	clock.start(at(5), false)
	clock.end(at(6), false)
	clock.endPhase(at(7), copying)
	clock.endPhase(at(8), preparation)
	clock.endPhase(at(9), preparation)
	timing := clock.snapshot(at(10), false)
	want := map[string]int64{"request_preparation": 2000, "context_copy": 2000, "unattributed": 3000}
	if timing.ModelMS != 2000 || timing.ToolMS != 1000 || timing.OtherMS != 7000 || !reflect.DeepEqual(timing.OtherPhasesMS, want) {
		t.Fatalf("nested phases double-counted waits or each other: %+v", timing)
	}
}

func TestTurnClockConcurrentPhaseCompletion(t *testing.T) {
	started := time.Now()
	at := func(seconds int) time.Time { return started.Add(time.Duration(seconds) * time.Second) }
	clock := &turnClock{started: started, last: started}
	first := clock.startPhase(at(0), "recall")
	second := clock.startPhase(at(1), "event_recording")
	clock.endPhase(at(2), first)
	clock.endPhase(at(3), second)
	timing := clock.snapshot(at(4), false)
	want := map[string]int64{"recall": 1000, "event_recording": 2000, "unattributed": 1000}
	if !reflect.DeepEqual(timing.OtherPhasesMS, want) {
		t.Fatalf("out-of-order completion corrupted phases: %v", timing.OtherPhasesMS)
	}
}

func TestTurnClockPhaseRoundingAndLateCompletion(t *testing.T) {
	started := time.Now()
	clock := &turnClock{started: started, last: started}
	first := clock.startPhase(started, "context_copy")
	clock.endPhase(started.Add(600*time.Microsecond), first)
	clock.startPhase(started.Add(600*time.Microsecond), "event_recording")
	timing := clock.snapshot(started.Add(1200*time.Microsecond), true)
	want := map[string]int64{"context_copy": 0, "event_recording": 0, "unattributed": 1}
	if timing.OtherMS != 1 || !reflect.DeepEqual(timing.OtherPhasesMS, want) {
		t.Fatalf("rounded phases do not reconcile: %+v", timing)
	}
	clock.PhaseStarted("late")()
	if !reflect.DeepEqual(timing.OtherPhasesMS, want) || len(clock.phases) != 1 {
		t.Fatal("late worker changed sealed phases")
	}
}

func TestTurnClockAdmissionProfiling(t *testing.T) {
	started := time.Now()
	clock := &turnClock{started: started, last: started}
	runtime := &runtime{admitting: clock}
	endCopy := runtime.phase("context_copy")
	if len(clock.phases) != 1 || clock.phases[0].name != "context_copy" {
		t.Fatal("startup snapshot copying was not profiled before the turn became active")
	}
	endCopy()
	endCopy()
	if len(clock.phases) != 0 {
		t.Fatal("phase completion was not idempotent")
	}
	runtime.admitting = nil
	runtime.phase("idle")()
	if len(clock.phases) != 0 {
		t.Fatal("idle work was charged to the admitted turn")
	}
}

func TestTurnClockPartitionsWaits(t *testing.T) {
	started := time.Date(2026, 10, 25, 12, 0, 0, 0, time.UTC)
	at := func(seconds int) time.Time { return started.Add(time.Duration(seconds) * time.Second) }
	tests := []struct {
		name                     string
		steps                    []struct{ second, kind int }
		modelMS, toolMS, otherMS int64
		requests, calls          int
	}{
		{
			name:    "background lifetime does not extend its report wait",
			steps:   []struct{ second, kind int }{{0, 1}, {1, -1}, {1, 2}, {2, -2}, {2, 1}, {10, -1}},
			modelMS: 9000, toolMS: 1000, otherMS: 0, requests: 2, calls: 1,
		},
		{
			name:    "queueing and pause are other, repeated waits count separately",
			steps:   []struct{ second, kind int }{{2, 1}, {3, -1}, {5, 2}, {6, -2}, {6, 2}, {7, -2}, {8, 1}, {9, -1}},
			modelMS: 2000, toolMS: 2000, otherMS: 6000, requests: 2, calls: 2,
		},
		{
			name:    "model has priority over an overlapping foreground wait",
			steps:   []struct{ second, kind int }{{1, 2}, {2, 1}, {6, -1}, {8, -2}},
			modelMS: 4000, toolMS: 3000, otherMS: 3000, requests: 1, calls: 1,
		},
		{
			name:    "concurrent model waits count wall time once",
			steps:   []struct{ second, kind int }{{1, 1}, {2, 1}, {6, -1}, {8, -1}},
			modelMS: 7000, toolMS: 0, otherMS: 3000, requests: 2,
		},
		{
			name:    "interrupted open tool wait is bounded by the turn",
			steps:   []struct{ second, kind int }{{1, 2}},
			modelMS: 0, toolMS: 9000, otherMS: 1000, calls: 1,
		},
		{
			name:    "interrupted provider is bounded by the turn",
			steps:   []struct{ second, kind int }{{1, 1}},
			modelMS: 9000, toolMS: 0, otherMS: 1000, requests: 1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			clock := &turnClock{started: started, last: started}
			for _, step := range test.steps {
				switch step.kind {
				case 1:
					clock.start(at(step.second), true)
				case -1:
					clock.end(at(step.second), true)
				case 2:
					clock.start(at(step.second), false)
				case -2:
					clock.end(at(step.second), false)
				}
			}
			timing := clock.snapshot(at(10), true)
			if timing.TotalMS != 10000 || timing.ModelMS != test.modelMS || timing.ToolMS != test.toolMS || timing.OtherMS != test.otherMS || timing.Requests != test.requests || timing.Calls != test.calls || !timing.Interrupted {
				t.Fatalf("unexpected breakdown: %+v", timing)
			}
		})
	}
}

func TestTurnClockSealsLateWorkers(t *testing.T) {
	started := time.Now()
	clock := &turnClock{started: started, last: started}
	endModel := clock.ModelStarted()
	endTool := clock.ToolStarted()
	timing := clock.finish(true)
	endModel()
	endModel()
	endTool()
	clock.ModelStarted()()
	clock.ToolStarted()()
	if clock.models != 1 || clock.tools != 1 || clock.requests != 1 || clock.calls != 1 {
		t.Fatalf("late worker mutated sealed clock: %+v", clock)
	}
	if timing.TotalMS != timing.ModelMS+timing.ToolMS+timing.OtherMS {
		t.Fatalf("rounded categories do not reconcile: %+v", timing)
	}
}

func TestTurnClockMeasuredZero(t *testing.T) {
	started := time.Now()
	clock := &turnClock{started: started, last: started}
	clock.start(started, false)
	clock.end(started.Add(100*time.Microsecond), false)
	timing := clock.snapshot(started.Add(500*time.Microsecond), false)
	if timing.TotalMS != 0 || timing.ToolMS != 0 || timing.Calls != 1 || timing.Interrupted {
		t.Fatalf("measured zero lost dispatch count: %+v", timing)
	}
}
