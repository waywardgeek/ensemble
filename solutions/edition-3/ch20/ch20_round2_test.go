package ensemble_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"ensemble"
)

func TestCh20Round2FailedDelegatesCloseAdmission(t *testing.T) {
	for _, mode := range []string{"exhausted", "provider", "budget"} {
		t.Run(mode, func(t *testing.T) {
			entered, release := make(chan struct{}, 1), make(chan struct{})
			var once sync.Once
			var calls atomic.Int32
			root, parent := ch20Rig(t, func(w http.ResponseWriter, b []byte) {
				n := calls.Add(1)
				if mode != "exhausted" || n == 4 {
					entered <- struct{}{}
					<-release
				}
				switch mode {
				case "provider":
					w.WriteHeader(400)
					fmt.Fprint(w, `{"error":{"message":"fixture provider failure"}}`)
				case "budget":
					w.Header().Set("Content-Type", "application/json")
					fmt.Fprintf(w, `{"id":"fixture","usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2},"choices":[{"message":{"role":"assistant","content":null,"tool_calls":[{"id":"read-%d","type":"function","function":{"name":"read_file","arguments":"{\"path\":\"missing.txt\"}"}}]},"finish_reason":"tool_calls"}]}`, n)
				default:
					ch20Reply(w, "ordinary finish", false)
				}
			})
			t.Cleanup(func() { once.Do(func() { close(release) }) })
			if mode == "budget" {
				if err := parent.SetMaxToolRounds(1); err != nil {
					t.Fatal(err)
				}
			}
			ctx := ch20Context(t)
			child, err := root.Spawn(ctx, parent.ID(), ensemble.Config{ID: "worker", Prompt: "initial task"})
			if err != nil {
				t.Fatal(err)
			}
			select {
			case <-entered:
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			}
			queued := child.Submit(ctx, "must not become another turn")
			once.Do(func() { close(release) })
			out, err := root.Wait(ctx, child.ID())
			if err != nil || out.State != "failed" || out.Error == "" {
				t.Fatalf("outcome %+v %v", out, err)
			}
			for name, ch := range map[string]<-chan ensemble.Result{"queued": queued, "live": child.SubmitLive(ctx, "after terminal failure"), "after": child.Submit(ctx, "after failure")} {
				select {
				case result := <-ch:
					if result.Err == nil {
						t.Fatalf("%s succeeded after terminal failure: %+v", name, result)
					}
				case <-ctx.Done():
					t.Fatal(name, ctx.Err())
				}
			}
			if _, err := root.SendMessage(ctx, parent.ID(), child.ID(), "wake again"); err == nil {
				t.Fatal("message reactivated failed child")
			}
			expected := int32(1)
			if mode == "exhausted" {
				expected = 4
			}
			if mode == "budget" {
				expected = 2
			}
			if calls.Load() != expected {
				t.Fatalf("provider calls=%d want %d", calls.Load(), expected)
			}
			if mode == "budget" && !strings.Contains(out.Error, "round") {
				t.Fatalf("not the intended budget failure: %s", out.Error)
			}
			if err := root.ShutdownAgent(child.ID()); err != nil {
				t.Fatal(err)
			}
			for _, a := range root.Agents() {
				if a.ID() == child.ID() {
					t.Fatal("shutdown child still managed")
				}
			}
			if len(root.Children(parent.ID())) != 0 {
				t.Fatal("shutdown child still listed")
			}
			if _, err := root.Status(child.ID()); err == nil {
				t.Fatal("shutdown status still managed")
			}
			retained, err := root.Outcome(child.ID())
			if err != nil || retained.Error != out.Error {
				t.Fatalf("lost outcome %+v %v", retained, err)
			}
			joined, err := root.Join(ctx, parent.ID(), []string{child.ID()})
			if err != nil || len(joined.Outcomes) != 1 {
				t.Fatalf("explicit retained join %+v %v", joined, err)
			}
			if err := root.AcceptSubmission(parent.ID(), json.RawMessage(`{"result":"resolved"}`)); err != nil {
				t.Fatal("acknowledged failure blocked parent", err)
			}
		})
	}
}

func TestCh20Round2TaskToolAndComplete(t *testing.T) {
	t.Run("complete-idle-default", func(t *testing.T) {
		var calls atomic.Int32
		root, a := ch20Rig(t, func(w http.ResponseWriter, b []byte) { calls.Add(1); ch20Reply(w, "unexpected", false) })
		result := <-a.Complete(ch20Context(t), "external result")
		if result.Err != nil || string(result.Submission) != `{"result":"external result"}` {
			t.Fatalf("idle complete %+v", result)
		}
		if calls.Load() != 0 {
			t.Fatal("terminal convenience became a chat prompt")
		}
		out, err := root.Outcome(a.ID())
		if err != nil || out.State != "submitted" {
			t.Fatalf("retained %+v %v", out, err)
		}
		raw, err := os.ReadFile(filepath.Join(a.Config().DataDir, "submit.md"))
		if err != nil {
			t.Fatal(err)
		}
		var row map[string]any
		if err = json.Unmarshal(raw, &row); err != nil || row["result"] != "external result" {
			t.Fatalf("default row %s %v", raw, err)
		}
	})

	t.Run("task-tool", func(t *testing.T) {
		var prompts atomic.Int32
		root, parent := ch20Rig(t, func(w http.ResponseWriter, b []byte) {
			var request struct {
				Messages []struct {
					Role    string          `json:"role"`
					Content json.RawMessage `json:"content"`
				} `json:"messages"`
			}
			if err := json.Unmarshal(b, &request); err != nil {
				t.Error(err)
			}
			for _, m := range request.Messages {
				if m.Role == "user" && strings.Contains(string(m.Content), "exact task once") {
					prompts.Add(1)
				}
			}
			ch20Reply(w, `{"result":"done"}`, true)
		})
		ctx := ch20Context(t)
		result := parent.InvokeTool(ctx, ensemble.Part{Type: "tool_call", Name: "spawn_sub_agent", CallID: "spawn-task", Args: json.RawMessage(`{"agent_id":"task-child","task":"exact task once"}`)}, "")
		if result != nil {
			t.Fatal(result)
		}
		for {
			if _, ok := root.Agent("task-child"); ok {
				break
			}
			select {
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			case <-time.After(time.Millisecond):
			}
		}
		out, err := root.Wait(ctx, "task-child")
		if err != nil || out.State != "submitted" {
			t.Fatalf("task spawn %+v %v; invoke %+v", out, err, result)
		}
		if prompts.Load() != 1 {
			t.Fatalf("task prompt appeared %d times", prompts.Load())
		}
	})
	t.Run("complete-active-turn", func(t *testing.T) {
		entered, release := make(chan struct{}, 1), make(chan struct{})
		var once sync.Once
		var calls atomic.Int32
		root, parent := ch20Rig(t, func(w http.ResponseWriter, b []byte) {
			calls.Add(1)
			entered <- struct{}{}
			<-release
			ch20Reply(w, "unused", false)
		})
		t.Cleanup(func() { once.Do(func() { close(release) }) })
		schema := json.RawMessage(`{"type":"object","properties":{"result":{"type":"string","minLength":8}},"required":["result"]}`)
		cfg := parent.Config().Child(filepath.Join(parent.Config().BaseDir, "completer"))
		cfg.ID = "completer"
		cfg.ParentID = ""
		cfg.SubmitSchema = schema
		var err error
		parent, err = root.NewAgent(cfg)
		if err != nil {
			t.Fatal(err)
		}
		ctx := ch20Context(t)
		active := parent.Submit(ctx, "active chat")
		select {
		case <-entered:
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
		bad := <-parent.Complete(ctx, "short")
		if bad.Err == nil {
			t.Fatal("invalid convenience submission accepted")
		}
		select {
		case r := <-active:
			t.Fatalf("refusal ended active work %+v", r)
		default:
		}
		queued := parent.Submit(ctx, "must be excluded")
		good := <-parent.Complete(ctx, "finished result")
		if good.Err != nil || string(good.Submission) != `{"result":"finished result"}` {
			t.Fatalf("complete %+v", good)
		}
		if r := <-active; r.Err != nil {
			t.Fatal(r.Err)
		}
		if r := <-queued; r.Err == nil {
			t.Fatal("queued chat survived Complete")
		}
		out, err := root.Wait(ctx, parent.ID())
		if err != nil || out.State != "submitted" {
			t.Fatalf("terminal %+v %v", out, err)
		}
		if calls.Load() != 1 {
			t.Fatalf("convenience sent another provider request: %d", calls.Load())
		}
		data, err := os.ReadFile(filepath.Join(parent.Config().DataDir, "submit.md"))
		if err != nil || !strings.Contains(string(data), `"payload":{"result":"finished result"}`) {
			t.Fatalf("submission file %s %v", data, err)
		}
		once.Do(func() { close(release) })
	})
}

func TestCh20Round2AddressedRouting(t *testing.T) {
	release := make(chan struct{})
	var once sync.Once
	root, parent := ch20Rig(t, func(w http.ResponseWriter, b []byte) { <-release; ch20Reply(w, "stopped fixture", false) })
	t.Cleanup(func() { once.Do(func() { close(release) }) })
	ctx := ch20Context(t)
	child := func(id string, owner ensemble.Agent) ensemble.Agent {
		c := owner.Config().Child(filepath.Join(parent.Config().BaseDir, id))
		c.ID = id
		c.ParentID = owner.ID()
		a, err := root.NewAgent(c)
		if err != nil {
			t.Fatal(err)
		}
		return a
	}
	left, right := child("left", parent), child("right", parent)
	grand := child("grand", left)
	route := func(sender, target ensemble.Agent, text string) {
		result := make(chan ensemble.Question, 1)
		failure := make(chan error, 1)
		go func() {
			reply, err := root.SendMessage(ctx, sender.ID(), target.ID(), text)
			if err != nil {
				failure <- err
				return
			}
			result <- reply
		}()
		deadline := time.After(3 * time.Second)
		for {
			err := root.RespondToQuestion(target.ID(), sender.ID(), "", "reply")
			if err == nil {
				break
			}
			select {
			case <-deadline:
				t.Fatal("question was not routed", err)
			case <-time.After(time.Millisecond):
			}
		}
		select {
		case reply := <-result:
			if reply.Sender != target.ID() || reply.Recipient != sender.ID() || reply.RequestID == "" || reply.Text != "reply" {
				t.Fatalf("reply envelope %+v", reply)
			}
		case err := <-failure:
			t.Fatal(err)
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
	}
	route(left, right, "sibling question")
	route(grand, parent, "ancestor question")
	route(right, grand, "cross-branch follow-up")
	if err := root.SetHumanAddress("operator"); err != nil {
		t.Fatal(err)
	}
	bad := parent.Config().Child(filepath.Join(parent.Config().BaseDir, "bad-human"))
	bad.ID = "operator"
	if _, err := root.NewAgent(bad); err == nil {
		t.Fatal("human became an Agent")
	}
	if _, err := root.Spawn(ctx, parent.ID(), ensemble.Config{ID: "operator", Prompt: "invalid"}); err == nil {
		t.Fatal("spawn human address")
	}
	if _, err := root.Status("operator"); err == nil {
		t.Fatal("human has lifecycle status")
	}
	if err := root.ShutdownAgent("operator"); err == nil {
		t.Fatal("human has shutdown lifecycle")
	}
	questions := make(chan ensemble.Question, 2)
	root.SetHumanQuestionSink(func(q ensemble.Question) { questions <- q })
	replies := make(chan ensemble.Question, 2)
	failures := make(chan error, 2)
	for _, a := range []ensemble.Agent{left, right} {
		go func(a ensemble.Agent) {
			r, err := root.SendMessage(ctx, a.ID(), "operator", "human help")
			if err != nil {
				failures <- err
				return
			}
			replies <- r
		}(a)
	}
	var qs []ensemble.Question
	for len(qs) < 2 {
		select {
		case q := <-questions:
			qs = append(qs, q)
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
	}
	if qs[0].RequestID == qs[1].RequestID {
		t.Fatal("human correlation collision")
	}
	if err := root.ReplyToHuman("", "ambiguous"); err == nil {
		t.Fatal("ambiguous reply accepted")
	}
	if err := root.ReplyToHuman(qs[0].RequestID, "first"); err != nil {
		t.Fatal(err)
	}
	if err := root.ReplyToHuman(qs[0].RequestID, "stale"); err == nil {
		t.Fatal("stale reply accepted")
	}
	if err := root.ReplyToHuman("", "second"); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		select {
		case r := <-replies:
			if r.Sender != "operator" || r.RequestID == "" || r.Recipient == "" {
				t.Fatalf("human envelope %+v", r)
			}
		case err := <-failures:
			t.Fatal(err)
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
	}
	once.Do(func() { close(release) })
}
