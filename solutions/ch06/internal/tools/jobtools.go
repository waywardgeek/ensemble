package tools

// The supervision tools. None of them is a job: they act on jobs, on the
// dispatcher's own goroutine, and return when they return.
//
// Measured across 73,777 real tool calls, send_input is the biggest of the
// three verbs by more than two to one. Job control is not mostly about
// stopping runaway work. It is mostly about talking to work that is going
// fine and is waiting for an answer.

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/waywardgeek/coding-agents-course/solutions/ch06/internal/common"
)

// lookupJob resolves a handle argument, and on a miss says what the good
// handles are — the model can only have got a handle from a result it was
// shown, so a bad one is almost always a typo it can fix on the next turn.
func lookupJob(c *common.Call, name string, handle int) (common.JobHandle, error) {
	j, ok := c.Jobs.Get(handle)
	if !ok {
		hs := c.Jobs.Handles()
		parts := make([]string, len(hs))
		for i, h := range hs {
			parts[i] = fmt.Sprint(h)
		}
		if len(parts) == 0 {
			return nil, fmt.Errorf("%s: no job with handle %d; no jobs have been started", name, handle)
		}
		return nil, fmt.Errorf("%s: no job with handle %d; handles so far: %s", name, handle, strings.Join(parts, ", "))
	}
	return j, nil
}

// --- wait_for_job ----------------------------------------------------------

// toolWaitForJob blocks until the job ends, or the delay passes, or the
// pattern shows up — the same wait the dispatcher did when it started the job,
// under limits the model chose this time.
//
// It MUST work on a job that has already finished. That case is not an edge
// case; it is the common one, because a job that finished a moment after the
// dispatcher stopped waiting is still "running" as far as the model knows.
func toolWaitForJob(c *common.Call, args json.RawMessage) (string, error) {
	var a struct {
		Handle int `json:"handle"`
		common.LimitArgs
	}
	if err := decode("wait_for_job", args, &a); err != nil {
		return "", err
	}
	j, err := lookupJob(c, "wait_for_job", a.Handle)
	if err != nil {
		return "", err
	}
	reason := j.Wait(c.Limits)
	return j.Report(reason, c.Limits), nil
}

// --- send_input ------------------------------------------------------------

// toolSendInput writes to a running process and then waits, exactly as
// wait_for_job does, for whatever it says back. The two halves are one tool
// because the model almost never wants one without the other: it typed a
// command at a prompt and it wants the prompt back.
func toolSendInput(c *common.Call, args json.RawMessage) (string, error) {
	var a struct {
		Handle        int    `json:"handle"`
		Input         string `json:"input"`
		AppendNewline *bool  `json:"append_newline"`
		common.LimitArgs
	}
	if err := decode("send_input", args, &a); err != nil {
		return "", err
	}
	j, err := lookupJob(c, "send_input", a.Handle)
	if err != nil {
		return "", err
	}
	text := a.Input
	if a.AppendNewline == nil || *a.AppendNewline {
		text += "\n"
	}
	if err := j.SendInput(text); err != nil {
		return "", fmt.Errorf("send_input: %v", err)
	}
	reason := j.Wait(c.Limits)
	return j.Report(reason, c.Limits), nil
}

// --- kill_job --------------------------------------------------------------

// toolKillJob stops a running job. For a process that means the whole
// process group. For a tool with no process it means the dispatcher stops
// listening — Go cannot kill a goroutine — and the result says so in those
// words rather than claiming a cancellation that did not happen.
//
// Killing a job that has already ended is not an error. The model asked for
// a state and the state holds; the reply says which way it got there.
func toolKillJob(c *common.Call, args json.RawMessage) (string, error) {
	var a struct {
		Handle int `json:"handle"`
	}
	if err := decode("kill_job", args, &a); err != nil {
		return "", err
	}
	j, err := lookupJob(c, "kill_job", a.Handle)
	if err != nil {
		return "", err
	}
	if !j.Kill("kill_job") {
		return fmt.Sprintf("job %d was already %s; nothing to kill. %d bytes of output at %s",
			j.Data().Handle, j.Status(), j.Bytes(), j.Data().Output.Locator), nil
	}
	data := j.Data()
	data.Reason = "kill_job"
	c.Events = append(c.Events, common.Event{Type: common.JobKilled, Job: data})

	if data.ExitCode == nil && !j.HasProcess() {
		return fmt.Sprintf("job %d marked killed. It is a %s call with no process, and Go cannot stop a goroutine: "+
			"the call may still be running and its late result will be discarded. %d bytes of output at %s",
			j.Data().Handle, j.Data().Tool, j.Bytes(), j.Data().Output.Locator), nil
	}
	return fmt.Sprintf("job %d killed (process group sent SIGKILL). %d bytes of output at %s",
		j.Data().Handle, j.Bytes(), j.Data().Output.Locator), nil
}

// --- tool_limits -----------------------------------------------------------

// toolLimits sets the wait limits for the NEXT call, whatever tool it is.
//
// It exists for tools whose argument schema we do not own and cannot add
// ai_callback_delay to — a tool loaded from someone else's server, which is
// also the tool most likely to need it. The three tools that wait take the
// limits as ordinary arguments, because the model changes how long it wants
// to wait on nearly every monitoring call and a second round trip each time
// would be a tax on the common case.
//
// One-shot, on purpose. A limit that persisted would be an escape hatch that
// is always open: a 600-second delay set once for a build silently applies
// to a read_file twenty calls later.
func toolLimits(c *common.Call, args json.RawMessage) (string, error) {
	var a common.LimitArgs
	if err := decode("tool_limits", args, &a); err != nil {
		return "", err
	}
	l, err := common.DefaultLimits().Overlay(a)
	if err != nil {
		return "", fmt.Errorf("tool_limits: %v", err)
	}
	c.Jobs.SetNext(l)
	return fmt.Sprintf("limits pending for the NEXT tool call, whichever tool that is: %s. "+
		"Call the tool you meant them for now; the next call consumes them and its result says so.", l), nil
}
