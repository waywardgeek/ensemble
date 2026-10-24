package tool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"ensemble/internal/common"
)

// Register extends only this Agent's registry, using the same ordinary job path.
func (r *registry) Register(d common.ToolDefinition, handler common.ToolHandler) error {
	return r.RegisterTool(common.ToolSpec{ToolDefinition: d, Handler: handler, Enabled: true, Source: "application", Origin: "manual"})
}
func (r *registry) RegisterTool(t common.ToolSpec) error { return r.parent.Skills().RegisterTool(t) }
func (r *registry) Dispatch(ctx context.Context, name string, input json.RawMessage) (common.PartList, bool) {
	result := r.Prepare(ctx, common.Part{Name: name, Args: input}, "", 0).Await(ctx)
	return result.Parts, result.IsError
}

type operation struct {
	fleetCompletion bool
	automatic       bool
	registry        *registry
	call            common.Part
	args            arguments
	limits          common.ToolLimits
	custom          bool
	job             common.Job
	control         bool
	err             error
	released        atomic.Bool
	dispatched      atomic.Bool
	eventSeq        common.Seq
}

func (o *operation) SetEventSeq(seq common.Seq) { o.eventSeq = seq }
func (o *operation) Dispatched() bool           { return o.dispatched.Load() }
func (o *operation) Called() common.ToolData {
	d := common.ToolData{Automatic: o.automatic, DisplayID: o.call.DisplayID, CallPartID: o.call.DisplayID, Dispatched: o.Dispatched(), CallID: o.call.CallID, Name: o.call.Name, Args: o.call.Args}
	if o.job != nil {
		j := o.job.Data()
		d.Job = &j
	}
	return d
}
func (o *operation) Release() { _ = o.ReleaseAfter(nil) }
func (o *operation) ReleaseAfter(before func() error) error {
	if o.err != nil {
		o.released.Store(true)
		return nil
	}
	if o.job != nil {
		started, err := o.job.StartAfter(before)
		o.dispatched.Store(started)
		o.released.Store(true)
		return err
	}
	if before != nil {
		if err := before(); err != nil {
			return err
		}
	}
	o.dispatched.Store(o.control)
	o.released.Store(true)
	return nil
}
func (o *operation) Abort(err error) {
	o.err = err
	o.released.Store(true)
	if o.job != nil {
		o.job.Abort(err)
	}
}
func (o *operation) Await(ctx context.Context) common.ToolData {
	if (o.job != nil || (o.control && o.err == nil)) && !o.released.Load() {
		reply := make(chan common.ToolData, 1)
		if err := o.registry.parent.Post(common.ToolAdmission{Context: ctx, Operation: o, Reply: reply}); err != nil {
			o.Abort(err)
		} else {
			select {
			case result := <-reply:
				return result
			case <-ctx.Done():
				d := o.Called()
				d.IsError = true
				d.Parts = common.PartList{common.Text(ctx.Err().Error())}
				return d
			}
		}
	}
	d := o.Called()
	var text string
	err := o.err
	if err == nil && o.control {
		var report *common.JobReport
		text, report, err = o.registry.control(ctx, o.call.Name, o.args, o.limits, o.fleetCompletion)
		if report != nil {
			// Controls report an existing target; they never own a job.
			text = formatReport(*report)
			d.IsError = report.IsError
		}
	} else if err == nil {
		var report common.JobReport
		var reportErr error
		if o.automatic {
			report, reportErr = o.job.Complete(ctx)
		} else {
			report, reportErr = o.job.InitialReport(ctx, o.limits)
		}
		err = reportErr
		d.Job = &report.Job
		d.IsError = report.IsError
		text = report.Text
		if !o.automatic && (report.Job.Status != common.JobDone || report.Job.Bytes > int64(o.limits.MaxOutputBytes) || report.Job.Cwd != "") {
			text = formatReport(report)
		}
	}
	if err != nil {
		text = err.Error()
		d.IsError = true
	}
	if o.fleetCompletion && err == nil && !d.IsError {
		d.FleetCompletion = true
	}
	if o.custom {
		pattern := "none"
		if o.limits.Pattern != "" {
			pattern = fmt.Sprintf("%q", o.limits.Pattern)
		}
		text = fmt.Sprintf("[tool_limits consumed by this %s call: ai_callback_delay %s, ai_callback_pattern %s, max_output_bytes %d]\n", o.call.Name, o.limits.Delay, pattern, o.limits.MaxOutputBytes) + text
	}
	text = common.Redact(text, common.CredentialValues(o.registry.parent.Config())...)
	d.Parts = common.PartList{common.Text(text)}
	return d
}
func (r *registry) Collect(ctx context.Context, name string) (common.PartList, error) {
	d := r.prepare(ctx, common.Part{Name: name, Args: json.RawMessage(`{}`)}, "", 0, true).Await(ctx)
	if d.IsError {
		var text strings.Builder
		for _, p := range d.Parts {
			text.WriteString(p.Text)
		}
		return nil, errors.New(text.String())
	}
	return d.Parts, nil
}
func (r *registry) Prepare(ctx context.Context, call common.Part, refusal string, turn uint64) common.ToolOperation {
	return r.prepare(ctx, call, refusal, turn, false)
}
func (r *registry) prepare(ctx context.Context, call common.Part, refusal string, turn uint64, automatic bool) common.ToolOperation {
	// Literal next means every position, including refusals and skipped setters.
	r.mu.Lock()
	o := &operation{registry: r, call: call, limits: common.DefaultToolLimits(), custom: !automatic && r.next != nil, automatic: automatic}
	if o.custom {
		o.limits = *r.next
	}
	if !automatic {
		r.next = nil
	}
	r.mu.Unlock()
	var selected *common.ToolSpec
	for _, t := range r.Snapshot().Tools {
		if t.Enabled && t.Name == call.Name && ((t.Ephemeral != "") == automatic) {
			copy := t
			selected = &copy
			break
		}
	}
	o.args, o.err = decode(call.Args)
	if o.err == nil && (call.Name == "run_command" || call.Name == "wait_for_job" || call.Name == "send_input" || (selected != nil && strings.HasPrefix(selected.Source, "mcp:"))) {
		o.limits, o.err = parseLimits(o.args, o.limits)
	}
	if selected == nil {
		o.err = fmt.Errorf("unknown tool %q", call.Name)
	}
	if refusal != "" {
		o.err = errors.New(refusal)
	}
	if err := common.ToolPolicy(r.parent.Config(), call.Name); err != nil {
		o.err = err
	}
	// Rejection is a tool result, never execution or job allocation.
	if o.err != nil {
		return o
	}
	if selected.Source == fleetSource && selected.Name == "submit" {
		o.fleetCompletion = true
		o.control = true
		if turn == 0 {
			o.err = errors.New("submit requires an active conversation turn")
		} else {
			o.err = common.ValidateSubmission(r.parent.Config().SubmitSchema, call.Args)
		}
		return o
	}
	if call.Name == "wait_for_job" || call.Name == "send_input" || call.Name == "kill_job" || call.Name == "tool_limits" {
		o.control = true
		if call.Name == "tool_limits" {
			o.err = o.args.only("ai_callback_delay", "ai_callback_pattern", "max_output_bytes")
			if o.err == nil {
				_, o.err = parseLimits(o.args, common.DefaultToolLimits())
			}
		}
		return o
	}
	handler := func(ctx context.Context, owner common.ToolCall, raw json.RawMessage) error {
		if call.Name == "load_skill" || call.Name == "unload_skill" {
			return r.skillCall(owner, call.Name, o.args, o.eventSeq)
		}
		if strings.HasPrefix(selected.Source, "mcp:") {
			var input map[string]json.RawMessage
			if err := json.Unmarshal(raw, &input); err != nil {
				return err
			}
			for _, name := range []string{"ai_callback_delay", "ai_callback_pattern", "max_output_bytes"} {
				delete(input, name)
			}
			raw, _ = json.Marshal(input)
		}
		return selected.Handler(ctx, owner, raw)
	}
	cwd := ""
	if call.Name == "run_command" {
		override, err := o.args.text("cwd", false)
		if err != nil {
			o.err = err
			return o
		}
		if _, present := o.args["cwd"]; present && override == "" {
			o.err = fmt.Errorf("cwd must not be empty")
			return o
		}
		// A refused cwd is not a job. Do not publish an expanded host cwd or
		// reserve an output locator/handle; preparation rechecks before launch.
		if _, err := r.parent.Sandbox().ValidatePath(override); err != nil {
			o.err = fmt.Errorf("cwd: %w", err)
			return o
		}
		if override != "" {
			cwd = r.path(override)
		}
	}
	origin := common.JobData{TurnID: turn, CallID: call.CallID, Name: call.Name, Args: append([]byte(nil), call.Args...)}
	o.job, o.err = r.parent.Jobs().Reserve(ctx, handler, call.Args, cwd, origin)
	return o
}
func formatReport(r common.JobReport) string {
	d := r.Job
	output := d.Output.Locator
	if output == "" {
		output = "unavailable"
	}
	header := fmt.Sprintf("job %d %s; %d bytes total; output: %s", d.Handle, d.Status, d.Bytes, output)
	if d.Cwd != "" {
		header += "; cwd: " + d.Cwd
	}
	if d.ExitCode != nil && d.Status != common.JobKilledStatus {
		header += fmt.Sprintf("; exit_code: %d", *d.ExitCode)
	}
	if d.Reason != "" {
		header += "; reason: " + d.Reason
	}
	if r.Text != "" {
		header += "\n" + r.Text
	}
	return header
}
func parseLimits(a arguments, l common.ToolLimits) (common.ToolLimits, error) {
	if raw, ok := a["ai_callback_delay"]; ok {
		var seconds float64
		if string(raw) == "null" || json.Unmarshal(raw, &seconds) != nil || math.IsNaN(seconds) || math.IsInf(seconds, 0) || seconds < 0 || seconds >= float64(math.MaxInt64)/float64(time.Second) {
			return l, errors.New("ai_callback_delay must be finite nonnegative seconds within duration range")
		}
		l.Delay = time.Duration(seconds * float64(time.Second))
	}
	if _, ok := a["ai_callback_pattern"]; ok {
		pattern, err := a.text("ai_callback_pattern", false)
		if err != nil {
			return l, err
		}
		if _, err = regexp.Compile(pattern); err != nil {
			return l, fmt.Errorf("ai_callback_pattern: %w", err)
		}
		l.Pattern = pattern
	}
	if _, ok := a["max_output_bytes"]; ok {
		n, err := a.integer("max_output_bytes", 0)
		if err != nil {
			return l, err
		}
		if n <= 0 {
			return l, errors.New("max_output_bytes must be a positive integer")
		}
		l.MaxOutputBytes = n
	}
	return l, nil
}
func (r *registry) control(ctx context.Context, name string, a arguments, limits common.ToolLimits, fleetCompletion bool) (string, *common.JobReport, error) {
	if fleetCompletion {
		fleet, ok := r.parent.Parent().(common.Fleet)
		if !ok {
			return "", nil, errors.New("fleet is unavailable")
		}
		raw, err := json.Marshal(a)
		if err == nil {
			err = fleet.AcceptSubmission(r.parent.ID(), raw)
		}
		return "Submission accepted; this conversation is terminal.", nil, err
	}

	if name == "tool_limits" {
		if err := a.only("ai_callback_delay", "ai_callback_pattern", "max_output_bytes"); err != nil {
			return "", nil, err
		}
		next, err := parseLimits(a, common.DefaultToolLimits())
		if err != nil {
			return "", nil, err
		}
		r.mu.Lock()
		r.next = &next
		r.mu.Unlock()
		return fmt.Sprintf("Limits for the next call, whichever tool that is: delay=%gs pattern=%q max_output_bytes=%d. That call will report the consumption; these settings are then spent.", next.Delay.Seconds(), next.Pattern, next.MaxOutputBytes), nil, nil
	}
	allowed := []string{"handle"}
	if name != "kill_job" {
		allowed = append(allowed, "ai_callback_delay", "ai_callback_pattern", "max_output_bytes")
	}
	if name == "send_input" {
		allowed = append(allowed, "input")
	}
	if err := a.only(allowed...); err != nil {
		return "", nil, err
	}
	handle, err := a.integer("handle", 0)
	if err != nil || handle <= 0 {
		return "", nil, errors.New("handle must be a positive integer")
	}
	// Validate everything before any input or kill side effect.
	limits, err = parseLimits(a, limits)
	if err != nil {
		return "", nil, err
	}
	input := ""
	if name == "send_input" {
		input, err = a.text("input", true)
		if err != nil {
			return "", nil, err
		}
	}
	job, err := r.parent.Jobs().Get(handle)
	if err != nil {
		return "", nil, err
	}
	var report common.JobReport
	switch name {
	case "wait_for_job":
		report, err = job.Wait(ctx, limits)
	case "send_input":
		report, err = job.SendInput(ctx, input, limits)
	case "kill_job":
		err = job.Kill("kill_job")
		if err == nil {
			limits.Delay = 0
			limits.Pattern = ""
			report, err = job.Wait(ctx, limits)
		}
	default:
		err = fmt.Errorf("unknown control %q", name)
	}
	if err != nil {
		return "", nil, err
	}
	// A successful kill operation is not itself a failed tool call.
	if name == "kill_job" {
		report.IsError = false
	}
	return "", &report, nil
}
func controlDefinitions() []entry {
	limits := `"ai_callback_delay":{"type":"number","minimum":0},"ai_callback_pattern":{"type":"string"},"max_output_bytes":{"type":"integer","minimum":1}`
	definition := func(name, description, properties, required string) entry {
		return entry{control: true, definition: common.ToolDefinition{Name: name, Description: description, Parameters: json.RawMessage(`{"type":"object","properties":{` + properties + `},"required":[` + required + `],"additionalProperties":false}`)}}
	}
	return []entry{
		definition("tool_limits", "Set observation limits for the literal next tool call only, including errors, setters, and supervision. Omitted values are fresh defaults: 3 seconds, no pattern, 16384 bytes. Limits never kill jobs.", limits, ""),
		definition("wait_for_job", "Observe new output from this Agent's job, including a finished job. Delay or a pattern on unseen output ends observation, never execution. Omitted delay is 3 seconds; cap is 16384 bytes.", `"handle":{"type":"integer","minimum":1},`+limits, `"handle"`),
		definition("send_input", "Submit a line to a running PTY job, appending a terminal newline if absent and preserving an existing terminator. Then observe with non-destructive callback limits; no raw-input mode.", `"handle":{"type":"integer","minimum":1},"input":{"type":"string"},`+limits, `"handle","input"`),
		definition("kill_job", "Kill and reap a managed process group. For Go-only work, mark killed and close output; Go cannot stop the handler. Repeated kills are safe.", `"handle":{"type":"integer","minimum":1}`, `"handle"`),
	}
}

// Invoke retains the direct Chapter 4 API without making the registry a History
// writer. The Agent admits, records, and joins this operation like other work.
func (r *registry) Invoke(ctx context.Context, call common.Part, refusal string) error {
	return r.parent.InvokeTool(ctx, call, refusal)
}
