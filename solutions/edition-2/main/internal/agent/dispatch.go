package agent

import (
	"ensemble/internal/common"
	"fmt"
)

// Retain serial call dispatch: one-shot limits and shell calls within a batch
// have a defined order. Handler execution and its report are separate workers,
// so a callback can return a handle before even a Go-only handler finishes.
func (r *runtime) dispatch() {
	t := r.active
	if len(t.calls) == 0 {
		r.send()
		return
	}
	call := t.calls[0]
	t.calls = t.calls[1:]
	var definition *common.ToolDefinition
	for i := range r.owner.tools {
		if r.owner.tools[i].Name == call.Name {
			definition = &r.owner.tools[i]
			break
		}
	}
	limits, notice, limitErr := r.owner.jobs.Consume(call.Args)
	owner := &callContext{parent: r.owner, limits: limits}
	called := &common.ToolData{CallID: call.CallID, Name: call.Name, Args: call.Args}
	if definition == nil || !definition.Supervision {
		job, err := r.owner.jobs.Start()
		if err != nil {
			r.finish("", err)
			return
		}
		owner.job = job
		data := job.Data()
		called.Job = &data
	}
	if err := r.append(common.Event{Type: "tool_called", Tool: called}); err != nil {
		r.finish("", err)
		return
	}
	t.current = &report{turn: t, call: call, data: called}
	run := func() (string, error) {
		if limitErr != nil {
			return "", limitErr
		}
		if definition == nil {
			return "", fmt.Errorf("unknown tool %q", call.Name)
		}
		return definition.Run(owner, call.Args)
	}
	if owner.job != nil {
		r.workers++
		go func() {
			text, err := run()
			owner.job.Finish(text, err)
			r.owner.box.Post(workerDone{job: owner.job})
		}()
	}
	r.workers++
	go func() {
		var text string
		var err error
		if owner.job == nil {
			text, err = run()
			if err != nil {
				text = err.Error()
			}
		} else {
			text, err = owner.job.Wait(limits)
		}
		if notice != "" {
			text = fmt.Sprintf("[tool_limits consumed by this %s call: %s]\n", call.Name, notice) + text
		}
		data := &common.ToolData{CallID: call.CallID, Parts: []common.Part{{Type: "text", Text: text}}, IsError: err != nil}
		if owner.job != nil {
			snapshot := owner.job.Data()
			data.Job = &snapshot
		}
		r.owner.box.Post(report{turn: t, call: call, data: data})
	}()
}
