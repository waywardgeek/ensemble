package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"ensemble/internal/common"
)

type invocation struct {
	ctx       context.Context
	name      string
	args      json.RawMessage
	automatic bool
	reply     chan common.Result
}
type invocationReady struct{ request invocation }

type ephemeralReady struct {
	turn *turn
	text string
}

// Invoke admits reverse calls through the actor, keeping real jobs and retained
// history. Results use job_report: a peer call is not a model-authored tool call.
func (a *agent) Invoke(ctx context.Context, name string, args json.RawMessage) (string, error) {
	return a.invoke(ctx, name, args, false)
}
func (a *agent) invoke(ctx context.Context, name string, args json.RawMessage, automatic bool) (string, error) {
	reply := make(chan common.Result, 1)
	if err := a.post(invocation{ctx, name, args, automatic, reply}); err != nil {
		return "", err
	}
	select {
	case result := <-reply:
		return result.Text, result.Err
	case <-ctx.Done():
		return "", ctx.Err()
	case <-a.done:
		return "", common.ErrStopped
	}
}

func (r *runtime) invoke(q invocation) {
	if r.stopping {
		q.reply <- common.Result{Err: common.ErrStopped}
		return
	}
	if err := q.ctx.Err(); err != nil {
		q.reply <- common.Result{Err: err}
		return
	}
	if !q.automatic {
		if resume := r.owner.pauseChannel(); resume != nil {
			r.workers++
			go func() {
				select {
				case <-resume:
				case <-q.ctx.Done():
				}
				r.owner.box.Post(invocationReady{q})
			}()
			return
		}
	}
	definition := r.owner.definition(q.name, q.automatic)
	if definition == nil {
		q.reply <- common.Result{Err: fmt.Errorf("unknown tool %q", q.name)}
		return
	}
	if len(q.args) == 0 {
		q.args = json.RawMessage(`{}`)
	}
	limits, notice, err := r.owner.jobs.Consume(q.args)
	if err != nil {
		q.reply <- common.Result{Err: err}
		return
	}
	call := &callContext{parent: r.owner, limits: limits}
	if !definition.Supervision {
		call.job, err = r.owner.jobs.Start()
		if err != nil {
			q.reply <- common.Result{Err: err}
			return
		}
	}
	data := &common.ToolData{CallID: fmt.Sprintf("external-%d", r.owner.history.LastSeq()+1), Name: q.name, Args: q.args}
	if call.job != nil {
		snapshot := call.job.Data()
		data.Job = &snapshot
	}
	if err := r.append(common.Event{Type: "tool_called", Tool: data}); err != nil {
		q.reply <- common.Result{Err: err}
		return
	}
	call.seq = r.owner.history.LastSeq()
	// These are real jobs, with the same one-shot limits and completion ownership
	// as model dispatch. Interrupting a turn still cannot kill another caller's job.
	completed := make(chan common.Result, 1)
	if call.job != nil {
		r.workers++
		go func() {
			text, err := definition.Run(call, q.args)
			call.job.Finish(text, err)
			// Automatic context needs the actual result, not a callback's running
			// notice. Ordinary callback callers may already have returned, so
			// publishing completion must not hold up the job's final history fact.
			completed <- common.Result{Text: text, Err: err}
			r.owner.box.Post(workerDone{job: call.job})
		}()
	}
	r.workers++
	go func() {
		var text string
		var err error
		if call.job == nil {
			text, err = definition.Run(call, q.args)
		} else if q.automatic {
			// invoke's caller can abandon this collection on turn interruption;
			// execution still completes and records its real result independently.
			result := <-completed
			text, err = result.Text, result.Err
			if call.job.Data().Status == "killed" {
				text, err = "", fmt.Errorf("automatic tool job was killed")
			}
		} else {
			text, err = call.job.Wait(limits)
		}
		if notice != "" {
			text = "[tool_limits consumed: " + notice + "]\n" + text
		}
		result := &common.ToolData{CallID: data.CallID, Name: q.name, Parts: []common.Part{{Type: "text", Text: text}}, IsError: err != nil}
		if call.job != nil {
			snapshot := call.job.Data()
			result.Job = &snapshot
		}
		if recordErr := r.owner.Record(common.Event{Type: "job_report", Tool: result}); recordErr != nil && err == nil {
			err = recordErr
		}
		q.reply <- common.Result{Text: text, Err: err}
		r.owner.box.Post(workerDone{})
	}()
}

// Gather outside the actor: reverse requests and job facts must remain live
// while an ephemeral peer is answering. Only this turn's completion is attached.
func (r *runtime) ephemera() {
	t := r.active
	var names []string
	for _, tool := range r.owner.definitions() {
		if r.owner.toolEnabled(tool.Name) && (tool.Ephemeral == "round" || (tool.Ephemeral == "turn" && !t.ephemeralTurn)) {
			names = append(names, tool.Name)
		}
	}
	t.ephemeralTurn = true
	if len(names) == 0 {
		r.sendRequest()
		return
	}
	r.workers++
	go func() {
		var parts []string
		for _, name := range names {
			text, err := r.owner.invoke(t.request.ctx, name, nil, true)
			if err != nil {
				r.owner.Ensemble().Logf("ephemeral tool %s: %v", name, err)
				continue
			}
			if text != "" {
				parts = append(parts, fmt.Sprintf("## %s (auto-updated)\n\n%s", name, text))
			}
		}
		r.owner.box.Post(ephemeralReady{t, strings.Join(parts, "\n\n")})
	}()
}
