package ensemble_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"ensemble/ensemble"
)

func timingSignal(t *testing.T, signal <-chan struct{}) {
	t.Helper()
	select {
	case <-signal:
	case <-time.After(3 * time.Second):
		t.Fatal("timing boundary was not reached")
	}
}

func timingHold(t *testing.T, pending <-chan ensemble.Result) int64 {
	t.Helper()
	started := time.Now()
	select {
	case result := <-pending:
		t.Fatalf("turn completed while its boundary was blocked: %+v", result)
	case <-time.After(120 * time.Millisecond):
	}
	return time.Since(started).Milliseconds()
}

func timingGate() (chan struct{}, func()) {
	gate := make(chan struct{})
	var once sync.Once
	return gate, func() { once.Do(func() { close(gate) }) }
}

func timingTurns(agent ensemble.Agent) []ensemble.Event {
	var turns []ensemble.Event
	for _, event := range agent.History().EventsAfter(0) {
		if event.Type == "turn_ended" {
			turns = append(turns, event)
		}
	}
	return turns
}

func timingSummary(t *testing.T, agent ensemble.Agent, index, count, calls, requests int, interrupted bool) ensemble.TurnTiming {
	t.Helper()
	turns := timingTurns(agent)
	if len(turns) != count {
		t.Fatalf("durable turn endings=%d, want %d", len(turns), count)
	}
	event := turns[index]
	if event.Turn == nil || event.Turn.Timing == nil {
		t.Fatalf("turn %d has no durable summary: %+v", index, event)
	}
	timing := *event.Turn.Timing
	if timing.Started.IsZero() || timing.Started.After(event.Time) {
		t.Errorf("invalid admission time: %+v, ended %v", timing, event.Time)
	}
	if timing.Calls != calls || timing.Requests != requests || timing.Interrupted != interrupted {
		t.Errorf("timing=%+v, want calls=%d requests=%d interrupted=%t", timing, calls, requests, interrupted)
	}
	if timing.TotalMS < 0 || timing.ModelMS < 0 || timing.ToolMS < 0 || timing.OtherMS < 0 ||
		absMilliseconds(timing.TotalMS-timing.ModelMS-timing.ToolMS-timing.OtherMS) > 3 {
		t.Errorf("durations do not reconcile: %+v", timing)
	}
	if timing.TotalMS > event.Time.Sub(timing.Started).Milliseconds()+3 {
		t.Errorf("timing extends beyond durable ending: %+v", timing)
	}
	var phaseMS int64
	for name, milliseconds := range timing.OtherPhasesMS {
		if name == "" || milliseconds < 0 {
			t.Errorf("invalid phase %q=%d", name, milliseconds)
		}
		phaseMS += milliseconds
	}
	if timing.OtherPhasesMS == nil || phaseMS != timing.OtherMS {
		t.Errorf("phase durations do not partition Other: %+v", timing)
	}
	if _, measured := timing.OtherPhasesMS["context_copy"]; !measured {
		t.Error("actor context snapshots were not profiled")
	}
	return timing
}

func timingBytes(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func timingWaitTool(t *testing.T, agent ensemble.Agent, name string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if _, err := agent.Wait(ctx, func(observation ensemble.Observation) bool {
		dispatched, ok := observation.(ensemble.ToolDispatched)
		return ok && dispatched.Name == name
	}); err != nil {
		t.Fatal(err)
	}
}

func TestTurnTimingQueuedAdmission(t *testing.T) {
	entered := make(chan struct{})
	gate, release := timingGate()
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		io.Copy(io.Discard, request.Body)
		if requests.Add(1) == 1 {
			close(entered)
			select {
			case <-gate:
			case <-request.Context().Done():
				return
			}
		}
		replyText(writer, "done")
	}))
	t.Cleanup(server.Close)
	agent := actorAgent(t, server, nil)
	t.Cleanup(release)
	first := agent.Submit(context.Background(), "first")
	timingSignal(t, entered)
	before := time.Now()
	second := agent.Submit(context.Background(), "queued")
	after := time.Now()
	if err := agent.Ephemeral("system arrival is not another turn"); err != nil {
		t.Fatal(err)
	}
	held := timingHold(t, second)
	release()
	for _, pending := range []<-chan ensemble.Result{first, second} {
		if result := receive(t, pending); result.Err != nil {
			t.Fatal(result.Err)
		}
	}
	timingSummary(t, agent, 0, 2, 0, 2, false)
	timing := timingSummary(t, agent, 1, 2, 0, 1, false)
	if timing.Started.Before(before) || timing.Started.After(after) {
		t.Errorf("queued start=%v, admission window=[%v, %v]", timing.Started, before, after)
	}
	if timing.OtherMS < held-3 || timing.TotalMS < held-3 || timing.ModelMS > timing.TotalMS-held+3 || timing.ToolMS != 0 {
		t.Errorf("queue wait %dms was lost or charged to work: %+v", held, timing)
	}
	if timing.OtherPhasesMS["admission_queue"] < held-3 {
		t.Errorf("queued admission not identified: %+v", timing)
	}
}

func TestTurnTimingPauseBeforeDispatchIsOther(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		io.Copy(io.Discard, request.Body)
		if requests.Add(1) == 1 {
			io.WriteString(writer, `{"content":[{"type":"tool_use","id":"paused","name":"echo","input":{}}],"usage":{"input_tokens":1,"output_tokens":1}}`)
			return
		}
		replyText(writer, "done")
	}))
	t.Cleanup(server.Close)
	agent := actorAgent(t, server, nil)
	entered := make(chan struct{})
	agent.RegisterTool(ensemble.ToolDeclaration{Name: "echo", Schema: json.RawMessage(`{"type":"object"}`)}, func(ensemble.ToolContext, json.RawMessage) (string, error) {
		close(entered)
		return "echo", nil
	})
	agent.Pause()
	t.Cleanup(agent.Unpause)
	pending := agent.Submit(context.Background(), "paused")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if _, err := agent.Wait(ctx, func(observation ensemble.Observation) bool {
		final, ok := observation.(ensemble.PartFinal)
		return ok && final.Part.Type == "tool_call"
	}); err != nil {
		t.Fatal(err)
	}
	held := timingHold(t, pending)
	select {
	case <-entered:
		t.Fatal("pause allowed tool dispatch")
	default:
	}
	agent.Unpause()
	if result := receive(t, pending); result.Err != nil {
		t.Fatal(result.Err)
	}
	timingSignal(t, entered)
	timing := timingSummary(t, agent, 0, 1, 1, 2, false)
	if timing.OtherMS < held-3 || timing.ModelMS+timing.ToolMS > timing.TotalMS-held+3 {
		t.Errorf("post-generation pause %dms was charged to model/tools: %+v", held, timing)
	}
	if timing.OtherPhasesMS["pause"] < held-3 {
		t.Errorf("pause not identified: %+v", timing)
	}
}

func TestTurnTimingBackgroundReportAndLaterWait(t *testing.T) {
	modelEntered := make(chan struct{})
	modelGate, releaseModel := timingGate()
	toolEntered := make(chan struct{})
	toolGate, releaseTool := timingGate()
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body, _ := io.ReadAll(request.Body)
		switch requests.Add(1) {
		case 1:
			io.WriteString(writer, `{"content":[{"type":"tool_use","id":"background","name":"slow","input":{"ai_callback_delay":0.01}},{"type":"tool_use","id":"after-pause","name":"echo","input":{}}],"usage":{"input_tokens":1,"output_tokens":1}}`)
		case 2:
			if !bytes.Contains(body, []byte("running")) {
				t.Error("model did not receive the continuing job report")
			}
			close(modelEntered)
			select {
			case <-modelGate:
			case <-request.Context().Done():
				return
			}
			replyText(writer, "background continues")
		case 3:
			io.WriteString(writer, `{"content":[{"type":"tool_use","id":"explicit-wait","name":"wait_for_job","input":{"handle":1,"ai_callback_delay":2}}],"usage":{"input_tokens":1,"output_tokens":1}}`)
		default:
			if !bytes.Contains(body, []byte("late real output")) {
				t.Error("explicit wait lost the job output")
			}
			replyText(writer, "wait complete")
		}
	}))
	t.Cleanup(server.Close)
	agent := actorAgent(t, server, func(config *ensemble.Config) { config.BuiltinTools = true })
	t.Cleanup(releaseModel)
	t.Cleanup(releaseTool)
	t.Cleanup(agent.Unpause)
	agent.RegisterTool(ensemble.ToolDeclaration{Name: "echo", Schema: json.RawMessage(`{"type":"object"}`)}, func(ensemble.ToolContext, json.RawMessage) (string, error) {
		return "echo", nil
	})
	agent.RegisterTool(ensemble.ToolDeclaration{Name: "slow", Schema: json.RawMessage(`{"type":"object"}`)}, func(ensemble.ToolContext, json.RawMessage) (string, error) {
		agent.Pause()
		close(toolEntered)
		<-toolGate
		return "late real output", nil
	})
	first := agent.Submit(context.Background(), "start background")
	timingSignal(t, toolEntered)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if _, err := agent.Wait(ctx, func(observation ensemble.Observation) bool {
		finished, ok := observation.(ensemble.ToolFinished)
		return ok && finished.CallID == "background"
	}); err != nil {
		t.Fatal(err)
	}
	pauseHeld := timingHold(t, first)
	agent.Unpause()
	timingSignal(t, modelEntered)
	modelHeld := timingHold(t, first)
	releaseModel()
	if result := receive(t, first); result.Err != nil {
		t.Fatal(result.Err)
	}
	job, err := agent.Jobs().Find(1)
	if err != nil {
		t.Fatal(err)
	}
	if job.Data().Status != "running" {
		t.Fatal("job must still run after its report and completed turn")
	}
	timing := timingSummary(t, agent, 0, 1, 2, 2, false)
	if timing.ModelMS < modelHeld-3 || timing.OtherMS < pauseHeld-3 || timing.ToolMS > timing.TotalMS-modelHeld-pauseHeld+4 {
		t.Errorf("background execution included post-report waits: %+v, model=%dms pause=%dms", timing, modelHeld, pauseHeld)
	}
	firstSummary := timingBytes(t, timingTurns(agent)[0])
	second := agent.Submit(context.Background(), "wait explicitly")
	timingWaitTool(t, agent, "wait_for_job")
	toolHeld := timingHold(t, second)
	releaseTool()
	if result := receive(t, second); result.Err != nil {
		t.Fatal(result.Err)
	}
	timing = timingSummary(t, agent, 1, 2, 1, 2, false)
	if timing.ToolMS < toolHeld-3 || timing.ModelMS > timing.TotalMS-toolHeld+3 {
		t.Errorf("explicit blocking wait %dms not charged to tools: %+v", toolHeld, timing)
	}
	if job.Data().Status != "done" || !bytes.Equal(firstSummary, timingBytes(t, timingTurns(agent)[0])) {
		t.Fatal("late job completion changed the already finished turn")
	}
}

func TestTurnTimingRepeatedCalls(t *testing.T) {
	for _, callID := range []string{"repeated-id", ""} {
		t.Run(fmt.Sprintf("id=%q", callID), func(t *testing.T) {
			entered := make(chan struct{}, 2)
			gate, release := timingGate()
			advance := make(chan struct{}, 2)
			var requests atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				io.Copy(io.Discard, request.Body)
				if requests.Add(1) <= 2 {
					fmt.Fprintf(writer, `{"content":[{"type":"tool_use","id":%q,"name":"step","input":{"ai_callback_delay":2}}],"usage":{"input_tokens":1,"output_tokens":1}}`, callID)
					return
				}
				replyText(writer, "two steps")
			}))
			t.Cleanup(server.Close)
			agent := actorAgent(t, server, nil)
			t.Cleanup(release)
			agent.RegisterTool(ensemble.ToolDeclaration{Name: "step", Schema: json.RawMessage(`{"type":"object"}`)}, func(ensemble.ToolContext, json.RawMessage) (string, error) {
				entered <- struct{}{}
				select {
				case <-advance:
				case <-gate:
				}
				return "step done", nil
			})
			pending := agent.Submit(context.Background(), "repeat the same tool")
			var held int64
			for range 2 {
				timingSignal(t, entered)
				held += timingHold(t, pending)
				advance <- struct{}{}
			}
			if result := receive(t, pending); result.Err != nil {
				t.Fatal(result.Err)
			}
			timing := timingSummary(t, agent, 0, 1, 2, 3, false)
			if timing.ToolMS < held-4 || timing.ModelMS > timing.TotalMS-held+4 {
				t.Errorf("repeated calls lost or overlapped %dms of waits: %+v", held, timing)
			}
			var dispatched []string
			var returned []string
			for _, event := range agent.History().EventsAfter(0) {
				if event.Type == "tool_called" {
					dispatched = append(dispatched, event.Tool.CallID)
				}
				if event.Type == "tool_returned" {
					returned = append(returned, event.Tool.CallID)
				}
			}
			if len(dispatched) != 2 || len(returned) != 2 || dispatched[0] == "" || dispatched[1] == "" || dispatched[0] != returned[0] || dispatched[1] != returned[1] {
				t.Fatalf("call/report identities differ: dispatched=%v returned=%v", dispatched, returned)
			}
			if callID == "" && dispatched[0] == dispatched[1] {
				t.Fatal("generated call IDs collided across model responses")
			}
		})
	}
}

func TestTurnTimingIncludesRecallJudge(t *testing.T) {
	entered := make(chan struct{}, 2)
	advance := make(chan struct{}, 2)
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body, _ := io.ReadAll(request.Body)
		attempt := requests.Add(1)
		entered <- struct{}{}
		select {
		case <-advance:
		case <-request.Context().Done():
			return
		}
		if attempt == 1 {
			if !bytes.Contains(body, []byte("zero-based candidate indices")) {
				t.Error("first request was not the real recall judge")
			}
			replyText(writer, "[]")
			return
		}
		replyText(writer, "done")
	}))
	t.Cleanup(server.Close)
	agent := actorAgent(t, server, func(config *ensemble.Config) {
		config.DisableStreaming = true
		config.Recall.Disabled = false
		recallCorpus(t, config.DataDir)
	})
	pending := agent.Submit(context.Background(), recallQuery)
	var held int64
	for range 2 {
		timingSignal(t, entered)
		held += timingHold(t, pending)
		advance <- struct{}{}
	}
	if result := receive(t, pending); result.Err != nil || result.Text != "done" {
		t.Fatalf("completion: %+v", result)
	}
	timing := timingSummary(t, agent, 0, 1, 0, 2, false)
	if requests.Load() != 2 || timing.ModelMS < held-4 || timing.ToolMS != 0 {
		t.Errorf("recall judge did not count both HTTP waits (%dms): %+v", held, timing)
	}
}

func TestTurnTimingInterruptedWaitIsBounded(t *testing.T) {
	for _, mode := range []string{"model", "tool"} {
		t.Run(mode, func(t *testing.T) {
			entered := make(chan struct{})
			gate, release := timingGate()
			var requests atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				io.Copy(io.Discard, request.Body)
				if requests.Add(1) == 1 {
					if mode == "tool" {
						io.WriteString(writer, `{"content":[{"type":"tool_use","id":"interrupt","name":"slow","input":{"ai_callback_delay":2}}],"usage":{"input_tokens":1,"output_tokens":1}}`)
						return
					}
					close(entered)
					<-gate
				}
				replyText(writer, "done")
			}))
			t.Cleanup(server.Close)
			agent := actorAgent(t, server, nil)
			t.Cleanup(release)
			agent.RegisterTool(ensemble.ToolDeclaration{Name: "slow", Schema: json.RawMessage(`{"type":"object"}`)}, func(ensemble.ToolContext, json.RawMessage) (string, error) {
				close(entered)
				<-gate
				return "late output", nil
			})
			pending := agent.Submit(context.Background(), "interrupt me")
			timingSignal(t, entered)
			held := timingHold(t, pending)
			if err := agent.Post(ensemble.Interrupt{}); err != nil {
				t.Fatal(err)
			}
			if result := receive(t, pending); !errors.Is(result.Err, ensemble.ErrInterrupted) {
				t.Fatalf("interrupted completion: %+v", result)
			}
			calls := 0
			if mode == "tool" {
				calls = 1
			}
			timing := timingSummary(t, agent, 0, 1, calls, 1, true)
			charged := timing.ModelMS
			if mode == "tool" {
				charged = timing.ToolMS
			}
			if charged < held-3 || agent.History().Context().Turn != "interrupted" || timingTurns(agent)[0].Turn.Error == "" {
				t.Errorf("interrupted %s wait lost its bounded duration/state: %+v", mode, timing)
			}
			before := timingBytes(t, timingTurns(agent)[0])
			if result := receive(t, agent.Submit(context.Background(), "next turn")); result.Err != nil {
				t.Fatal(result.Err)
			}
			timingSummary(t, agent, 1, 2, 0, 1, false)
			if mode == "tool" {
				job, err := agent.Jobs().Find(1)
				if err != nil || job.Data().Status != "running" {
					t.Fatal("interruption did not preserve the running job", err)
				}
			}
			release()
			if err := agent.Shutdown(); err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(before, timingBytes(t, timingTurns(agent)[0])) {
				t.Fatal("late completion extended an interrupted turn")
			}
		})
	}
}

func TestTurnTimingProviderErrorDoesNotLeakIntoNextTurn(t *testing.T) {
	entered := make(chan struct{})
	gate, release := timingGate()
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		io.Copy(io.Discard, request.Body)
		if requests.Add(1) == 1 {
			close(entered)
			select {
			case <-gate:
			case <-request.Context().Done():
				return
			}
			http.Error(writer, "provider failed", http.StatusInternalServerError)
			return
		}
		replyText(writer, "recovered")
	}))
	t.Cleanup(server.Close)
	agent := actorAgent(t, server, nil)
	t.Cleanup(release)
	pending := agent.Submit(context.Background(), "failed turn")
	timingSignal(t, entered)
	held := timingHold(t, pending)
	release()
	if result := receive(t, pending); result.Err == nil {
		t.Fatal("provider failure was acknowledged as success")
	}
	timing := timingSummary(t, agent, 0, 1, 0, 1, false)
	if timing.ModelMS < held-3 || timing.ToolMS != 0 || timingTurns(agent)[0].Turn.Error == "" {
		t.Errorf("failed HTTP request lost its timing or error: %+v", timing)
	}
	first := timingBytes(t, timingTurns(agent)[0])
	before := time.Now()
	if result := receive(t, agent.Submit(context.Background(), "recover")); result.Err != nil || result.Text != "recovered" {
		t.Fatalf("next turn failed: %+v", result)
	}
	timing = timingSummary(t, agent, 1, 2, 0, 1, false)
	if timing.Started.Before(before) || timingTurns(agent)[1].Turn.Error != "" || !bytes.Equal(first, timingBytes(t, timingTurns(agent)[0])) {
		t.Fatal("provider failure leaked into the next turn or changed its original summary")
	}
}

func TestTurnTimingDurableReplayRestartAndLegacy(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		io.Copy(io.Discard, request.Body)
		requests.Add(1)
		replyText(writer, "saved answer")
	}))
	t.Cleanup(server.Close)
	agent := actorAgent(t, server, nil)
	if err := agent.Load(); err != nil {
		t.Fatal(err)
	}
	if result := receive(t, agent.Submit(context.Background(), "save timing")); result.Err != nil {
		t.Fatal(result.Err)
	}
	timingSummary(t, agent, 0, 1, 0, 1, false)
	turn := timingTurns(agent)[0]
	if turn.Turn.Text != "saved answer" || turn.Turn.Error != "" {
		t.Fatalf("durable completion lost final answer: %+v", turn.Turn)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	observation, err := agent.Wait(ctx, func(observation ensemble.Observation) bool {
		_, ok := observation.(ensemble.TurnEnded)
		return ok
	})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(timingBytes(t, observation), timingBytes(t, turn.Turn)) {
		t.Fatal("live and durable completion summaries differ")
	}
	var wire map[string]json.RawMessage
	if err := json.Unmarshal(timingBytes(t, turn), &wire); err != nil || len(wire["turn"]) == 0 {
		t.Fatal("durable event lacks the public JSON turn payload", err)
	}
	var log bytes.Buffer
	if err := agent.History().Dump(&log); err != nil {
		t.Fatal(err)
	}
	want := timingBytes(t, timingTurns(agent))
	replay := actorAgent(t, server, nil)
	if err := replay.History().Load(bytes.NewReader(log.Bytes())); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(want, timingBytes(t, timingTurns(replay))) {
		t.Fatal("log replay changed timing bytes")
	}
	if err := agent.Shutdown(); err != nil {
		t.Fatal(err)
	}
	if err := agent.Save(); err != nil {
		t.Fatal(err)
	}
	restarted := actorAgent(t, server, func(config *ensemble.Config) { *config = agent.Config() })
	if err := restarted.Load(); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(want, timingBytes(t, timingTurns(restarted))) || requests.Load() != 1 {
		t.Fatal("restart changed timing or contacted the provider")
	}
	legacy := actorAgent(t, server, nil)
	if err := legacy.History().Load(strings.NewReader(`{"seq":1,"type":"message_received","message":{"actor":"human","parts":[{"type":"text","text":"old turn"}]}}
{"seq":2,"type":"turn_ended"}
`)); err != nil {
		t.Fatal(err)
	}
	for _, event := range legacy.History().EventsAfter(0) {
		if event.Turn != nil && event.Turn.Timing != nil {
			t.Fatal("legacy events invented timing metrics")
		}
	}
}

func TestTurnTimingAutomaticEphemeralCalls(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body, _ := io.ReadAll(request.Body)
		requests.Add(1)
		if !bytes.Contains(body, []byte("completed-startup")) || !bytes.Contains(body, []byte("completed-snapshot")) || bytes.Contains(body, []byte("failed-automatic")) {
			t.Error("automatic context did not preserve successful results and exclude failed results")
		}
		replyText(writer, "done")
	}))
	t.Cleanup(server.Close)
	agent := mcpAgent(t, server.URL)
	external := peer(t, agent, "timing", "startup", "snapshot", "broken")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pending := agent.Submit(ctx, "use automatic context")
	var held int64
	seen := make(map[string]bool)
	for range 3 {
		call := receiveRPC(t, external.requests)
		var params struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(call["params"], &params); err != nil {
			t.Fatal(err)
		}
		if params.Name != "startup" && params.Name != "snapshot" && params.Name != "broken" || seen[params.Name] {
			t.Fatalf("unexpected automatic invocation: %q", params.Name)
		}
		seen[params.Name] = true
		if len(seen) == 1 {
			text, err := agent.Invoke(ctx, "echo", json.RawMessage(`{"text":"unrelated reverse call"}`))
			if err != nil || text != "unrelated reverse call" {
				t.Fatalf("reverse call did not complete alongside automatic work: %q %v", text, err)
			}
		}
		held += timingHold(t, pending)
		if requests.Load() != 0 {
			t.Fatal("provider started before automatic context completed")
		}
		text := "completed-" + params.Name
		if params.Name == "broken" {
			text = "failed-automatic"
		}
		external.reply(call["id"], map[string]any{"content": []map[string]string{{"type": "text", "text": text}}, "isError": params.Name == "broken"})
	}
	if result := receive(t, pending); result.Err != nil || result.Text != "done" {
		t.Fatalf("automatic context completion: %+v", result)
	}
	timing := timingSummary(t, agent, 0, 1, 3, 1, false)
	if requests.Load() != 1 || timing.ToolMS < held-4 || timing.ModelMS > timing.TotalMS-held+4 {
		t.Errorf("automatic waits (%dms) were lost or attributed to model: %+v", held, timing)
	}
}

func TestTurnTimingOAuthRenewalIsOther(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	oauth := &oauthPeer{t: t, key: key, refresh: "refresh-one", initialExpiry: 3600}
	refreshEntered := make(chan struct{}, 1)
	refreshGate, releaseRefresh := timingGate()
	oauth.mu.Lock()
	oauth.server = httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/token" {
			if err := request.ParseForm(); err != nil {
				t.Error(err)
				return
			}
			if request.Form.Get("grant_type") == "refresh_token" {
				refreshEntered <- struct{}{}
				select {
				case <-refreshGate:
				case <-request.Context().Done():
					return
				}
			}
		}
		oauth.serve(writer, request)
	}))
	oauth.mu.Unlock()
	t.Cleanup(oauth.server.Close)
	root := authRoot(t, oauth, t.TempDir())
	if _, err := loginFake(root); err != nil {
		t.Fatal(err)
	}
	modelEntered := make(chan struct{}, 2)
	advance := make(chan struct{}, 2)
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		io.Copy(io.Discard, request.Body)
		attempt := requests.Add(1)
		if request.Header.Get("Authorization") != fmt.Sprintf("Bearer access-secret-%d", attempt-1) {
			t.Error("inference did not use the selected or renewed OAuth bearer")
		}
		modelEntered <- struct{}{}
		select {
		case <-advance:
		case <-request.Context().Done():
			return
		}
		if attempt == 1 {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}
		writer.Header().Set("Content-Type", "text/event-stream")
		frame(writer, wireFrame{"response.completed", decoded(`{"response":{"status":"completed","output":[{"type":"message","content":[{"type":"output_text","text":"renewed answer"}]}],"usage":{"input_tokens":1,"output_tokens":1}}}`)})
	}))
	t.Cleanup(server.Close)
	agent, err := root.NewAgent(ensemble.Config{Vendor: ensemble.OpenAI, Model: "gpt-ch19-course", BaseURL: server.URL, DataDir: t.TempDir(), Recall: ensemble.RecallSettings{Disabled: true}})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := agent.Shutdown(); err != nil {
			t.Error(err)
		}
	})
	t.Cleanup(releaseRefresh)
	pending := agent.Submit(context.Background(), "renew after 401")
	timingSignal(t, modelEntered)
	modelHeld := timingHold(t, pending)
	advance <- struct{}{}
	timingSignal(t, refreshEntered)
	refreshHeld := timingHold(t, pending)
	if requests.Load() != 1 {
		t.Fatal("inference retried before credential renewal completed")
	}
	releaseRefresh()
	timingSignal(t, modelEntered)
	modelHeld += timingHold(t, pending)
	advance <- struct{}{}
	if result := receive(t, pending); result.Err != nil || result.Text != "renewed answer" {
		t.Fatalf("renewed inference: %+v", result)
	}
	timing := timingSummary(t, agent, 0, 1, 0, 2, false)
	if timing.ModelMS < modelHeld-4 || timing.OtherMS < refreshHeld-3 || timing.ModelMS > timing.TotalMS-refreshHeld+3 || timing.ToolMS != 0 {
		t.Errorf("401/renewal attribution: model waits=%dms renewal=%dms summary=%+v", modelHeld, refreshHeld, timing)
	}
	if timing.OtherPhasesMS["credential_refresh"] < refreshHeld-3 {
		t.Errorf("renewal not identified: %+v", timing)
	}
	oauth.mu.Lock()
	rotations, tokenCalls := oauth.rotations, oauth.tokenCalls
	oauth.mu.Unlock()
	if requests.Load() != 2 || rotations != 1 || tokenCalls != 2 {
		t.Errorf("inference attempts=%d renewals=%d token exchanges=%d", requests.Load(), rotations, tokenCalls)
	}
}
