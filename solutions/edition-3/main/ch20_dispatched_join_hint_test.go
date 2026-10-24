package ensemble_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"ensemble"
)

func TestCh20HumanHintDuringDispatchedJoin(t *testing.T) {
	const childPrompt = "CONTROLLED_CHILD_FOR_DISPATCHED_JOIN"
	const hintText = "ORDINARY_HUMAN_HINT_AFTER_JOIN_DISPATCH"
	const joinCall = "join-already-dispatched"
	const payload = `{"count":1,"result":"child finished after hint"}`
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	childEntered := make(chan struct{}, 1)
	releaseChild := make(chan struct{})
	var release sync.Once
	var childRequests, parentRequests atomic.Int32
	root, parent := ch20Rig(t, func(w http.ResponseWriter, raw []byte) {
		if bytes.Contains(raw, []byte(childPrompt)) {
			if n := childRequests.Add(1); n != 1 {
				t.Errorf("unexpected child provider request %d", n)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			childEntered <- struct{}{}
			<-releaseChild
			ch20Reply(w, payload, true)
			return
		}
		switch n := parentRequests.Add(1); n {
		case 1:
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"id":"parent-join","object":"chat.completion","model":"fixture","choices":[{"index":0,"message":{"role":"assistant","content":"I will join the controlled child.","tool_calls":[{"id":"long-join-observation","type":"function","function":{"name":"tool_limits","arguments":"{\"ai_callback_delay\":60,\"max_output_bytes\":65536}"}},{"id":%q,"type":"function","function":{"name":"join_agents","arguments":"{\"agent_ids\":[\"controlled-child\"]}"}}]},"finish_reason":"tool_calls"}],"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`, joinCall)
		case 2:
			if !bytes.Contains(raw, []byte(hintText)) {
				t.Error("parent's next model request omitted the ordinary human hint")
			}
			ch20Reply(w, "Human hint observed; the same join continues in the background.", false)
		default:
			t.Errorf("unexpected parent provider request %d", n)
			ch20Reply(w, "unexpected request", false)
		}
	})
	// Release before ch20Rig's shutdown/server cleanup on every failure path.
	t.Cleanup(func() { release.Do(func() { close(releaseChild) }) })
	child, err := root.Spawn(ctx, parent.ID(), ensemble.Config{
		ID: "controlled-child", Prompt: childPrompt,
		SubmitSchema: json.RawMessage(`{"type":"object","properties":{"count":{"type":"integer"},"result":{"type":"string"}},"required":["count","result"],"additionalProperties":false}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-childEntered:
	case <-ctx.Done():
		t.Fatal("child did not reach its controlled provider barrier")
	}

	observations := parent.Observe()
	defer observations.Close()
	var pending []ensemble.Observation
	waitObservation := func(match func(ensemble.Observation) bool) ensemble.Observation {
		t.Helper()
		for {
			pending = append(pending, observations.Drain()...)
			for i, o := range pending {
				if match(o) {
					pending = append(pending[:i], pending[i+1:]...)
					return o
				}
			}
			select {
			case <-observations.Wake():
			case <-ctx.Done():
				t.Fatal("timed out waiting for actor observation")
			}
		}
	}
	parentTurn := parent.Submit(ctx, "Join the controlled child and keep listening for the human.")
	dispatched := waitObservation(func(o ensemble.Observation) bool {
		d, ok := o.(ensemble.ToolDispatched)
		return ok && d.Name == "join_agents" && d.CallID == joinCall
	}).(ensemble.ToolDispatched)

	var releaseSeq ensemble.Seq
	var handle int
	for _, e := range parent.Log().Events() {
		if e.Tool == nil || e.Tool.CallID != joinCall {
			continue
		}
		if e.Type == ensemble.ToolReturned {
			t.Fatal("join observation returned before the human hint")
		}
		if e.Type == ensemble.ToolReleased && e.Tool.Dispatched && e.Tool.Job != nil {
			releaseSeq, handle = e.Seq, e.Tool.Job.Handle
		}
	}
	if releaseSeq == 0 || handle == 0 {
		t.Fatal("no canonical dispatched join with a real job handle")
	}
	job, err := parent.Jobs().Get(handle)
	if err != nil {
		t.Fatal(err)
	}
	if d := job.Data(); d.Status != ensemble.JobRunning || d.Name != "join_agents" || d.CallID != joinCall || d.TurnID != dispatched.TurnID {
		t.Fatalf("join is not actually running after dispatch: %+v", d)
	}
	t.Logf("BEFORE_HINT: turn=%d release_seq=%d handle=%d status=running child=%s provider_blocked=true", dispatched.TurnID, releaseSeq, handle, child.ID())

	if err := parent.Post(ensemble.Hint{Text: hintText}); err != nil {
		t.Fatal(err)
	}
	hint := waitObservation(func(o ensemble.Observation) bool {
		h, ok := o.(ensemble.HintObserved)
		return ok && h.TurnID == dispatched.TurnID && h.Sender == "human"
	}).(ensemble.HintObserved)
	if len(hint.Parts) != 1 || hint.Parts[0].Text != hintText {
		t.Fatalf("wrong hint acknowledged: %+v", hint)
	}
	finished := waitObservation(func(o ensemble.Observation) bool {
		d, ok := o.(ensemble.ToolFinished)
		return ok && d.Name == "join_agents" && d.CallID == joinCall
	}).(ensemble.ToolFinished)
	if finished.IsError {
		t.Fatalf("hint turned the join observation into an error: %+v", finished)
	}
	var hintSeq, reportSeq ensemble.Seq
	for _, e := range parent.Log().Events() {
		if e.Type == ensemble.HintReceived && e.Message != nil && len(e.Message.Parts) == 1 && e.Message.Parts[0].Text == hintText {
			hintSeq = e.Seq
		}
		if e.Type == ensemble.ToolReturned && e.Tool != nil && e.Tool.CallID == joinCall {
			d := e.Tool
			if !d.Dispatched || d.IsError || d.Job == nil || d.Job.Handle != handle || d.Job.Status != ensemble.JobRunning {
				t.Fatalf("join was skipped, killed, or failed instead of waking its observer: %+v", d)
			}
			reportSeq = e.Seq
		}
	}
	if !(releaseSeq < hintSeq && hintSeq < reportSeq) {
		t.Fatalf("wrong ordering: release=%d hint=%d running-report=%d", releaseSeq, hintSeq, reportSeq)
	}
	select {
	case result := <-parentTurn:
		if result.Err != nil || !strings.Contains(result.Text, "Human hint observed") {
			t.Fatalf("parent did not resume from the hint: %+v", result)
		}
	case <-ctx.Done():
		t.Fatal("parent did not resume while child remained blocked")
	}
	if d := job.Data(); d.Status != ensemble.JobRunning || d.Handle != handle || d.Reason != "" {
		t.Fatalf("underlying join did not survive the observer/parent-turn completion: %+v", d)
	}
	if status, err := root.Status(child.ID()); err != nil || status.State != "running" || status.Outcome != nil {
		t.Fatalf("child did not survive the hint: %+v, %v", status, err)
	}
	t.Logf("AFTER_HINT: hint_seq=%d report_seq=%d handle=%d status=running is_error=false child=running parent_resumed=true; observation_limit=60s test_deadline=20s", hintSeq, reportSeq, handle)

	release.Do(func() { close(releaseChild) })
	outcome, err := root.Wait(ctx, child.ID())
	if err != nil || outcome.State != "submitted" || outcome.Error != "" || string(outcome.Submission) != payload {
		t.Fatalf("genuine child submit failed: %+v, %v", outcome, err)
	}
	retained, err := parent.Jobs().Get(handle)
	if err != nil || retained != job {
		t.Fatalf("join handle was replaced or lost: %v", err)
	}
	report, err := retained.Complete(ctx)
	if err != nil || report.IsError || report.Job.Handle != handle || report.Job.Status != ensemble.JobDone {
		t.Fatalf("same join did not settle honestly: %+v, %v", report, err)
	}
	var joined struct {
		Outcomes  []ensemble.AgentOutcome `json:"outcomes"`
		Questions []json.RawMessage       `json:"questions"`
	}
	if err := json.Unmarshal([]byte(report.Text), &joined); err != nil {
		t.Fatalf("join's actual job output is not structured: %v, %q", err, report.Text)
	}
	if len(joined.Outcomes) != 1 || len(joined.Questions) != 0 || joined.Outcomes[0].AgentID != child.ID() || joined.Outcomes[0].ParentID != parent.ID() || joined.Outcomes[0].State != "submitted" || joined.Outcomes[0].Error != "" || string(joined.Outcomes[0].Submission) != payload {
		t.Fatalf("same handle lost or falsified the terminal payload: %+v", joined)
	}
	if parentRequests.Load() != 2 || childRequests.Load() != 1 {
		t.Fatalf("unexpected loopback work: parent=%d child=%d", parentRequests.Load(), childRequests.Load())
	}
	t.Logf("TERMINAL_COLLECTION: handle=%d status=%s output=%s parent_requests=%d child_requests=%d", handle, report.Job.Status, strings.TrimSpace(report.Text), parentRequests.Load(), childRequests.Load())
}
