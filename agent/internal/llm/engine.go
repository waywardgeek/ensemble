package llm

// The engine: one turn, start to finish.
//
// Notice how little of it there is, and that none of it mentions a vendor.
// Everything vendor-shaped was pushed into the renderer and the parser, which
// is the entire point of Chapter 2.

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/waywardgeek/coding-agents-course/agent/internal/common"
)

type Engine struct {
	Log   *common.Log
	Ctx   *common.Context
	Cfg   common.Config
	HTTP  *http.Client
	Path  string // where the log is persisted, so `dump` can find it
	Jobs  common.JobManager
	Tools common.ToolRegistry
	Host  common.Host
}

func NewEngine(cfg common.Config, path string, jobs common.JobManager, tools common.ToolRegistry, host common.Host) *Engine {
	return &Engine{
		Log:   common.NewLog(),
		Ctx:   common.NewContext(),
		Cfg:   cfg,
		HTTP:  &http.Client{Timeout: 120 * time.Second},
		Path:  path,
		Jobs:  jobs,
		Tools: tools,
		Host:  host,
	}
}

// Record appends to the log and advances the context. One path in.
func (e *Engine) Record(ev common.Event) error {
	stored := e.Log.Append(ev)
	return e.Ctx.Apply(stored)
}

// Say records a human prompt.
func (e *Engine) Say(text string) error {
	return e.Record(common.Event{Type: common.MessageReceived, Message: &common.MessageData{
		Actor: common.ActorHuman, Parts: common.PartList{common.TextPart{Text: text}},
	}})
}

// Attach records volatile data — the time, the screen, live status.
//
// It is an ordinary common.MessageReceived with Actor: System. The reducer is what
// decides it is ephemeral rather than dialogue, which is the chapter's rule
// about classification made concrete: the capture site does not know, and
// cannot know, what an arriving message means.
func (e *Engine) Attach(text string) error {
	return e.Record(common.Event{Type: common.MessageReceived, Message: &common.MessageData{
		Actor: common.ActorSystem, Parts: common.PartList{common.TextPart{Text: text}},
	}})
}

// Turn renders the current context, sends it, and folds the response back in.
func (e *Engine) Turn() (string, error) {
	// Loud refusal: reject unknown models before doing anything else.
	if _, known := common.LookupModel(e.Cfg.Model); !known {
		err := fmt.Errorf("unknown model %q: not in the supported model table; refusing to proceed", e.Cfg.Model)
		_ = e.Record(common.Event{Type: common.ErrorOccurred, Error: &common.ErrorData{Message: err.Error()}})
		return "", err
	}

	renderer, parser, err := SeamFor(e.Cfg.Vendor)
	if err != nil {
		return "", err
	}

	// RENDER BEFORE RECORDING common.RequestSent.
	//
	// Reverse these two lines and the bug is subtle and expensive: the reducer
	// clears pending ephemera on common.RequestSent, so recording first means the
	// renderer never sees them and the volatile data is silently never
	// delivered. Nothing errors. The model just quietly does not know what
	// time it is. Chapter 4 hits the identical ordering trap with hints.
	req, err := renderer.Render(e.Ctx, e.Cfg)
	if err != nil {
		return "", err
	}
	if err := e.Record(common.Event{Type: common.RequestSent, Request: &common.RequestData{To: common.Provenance{
		Vendor: e.Cfg.Vendor, Model: e.Cfg.Model, Surface: e.Cfg.Surface,
	}}}); err != nil {
		return "", err
	}

	status, body, err := e.send(req)
	if err != nil {
		// A transport failure is infrastructure: it ends the turn.
		_ = e.Record(common.Event{Type: common.ErrorOccurred, Error: &common.ErrorData{Message: err.Error()}})
		return "", err
	}

	events, err := parser.Parse(status, body)
	if err != nil {
		return "", err
	}
	var vendorErr error
	for _, ev := range events {
		if err := e.Record(ev); err != nil {
			return "", err
		}
		if ev.Type == common.ErrorOccurred && ev.Error != nil {
			// The event is the record; this is the report. Without it a 404
			// on the model name prints {"assistant":""} and exits 0 — measured
			// live against Gemini, and invisible unless you read the log.
			vendorErr = fmt.Errorf("%s: %s", e.Cfg.Vendor, ev.Error.Message)
		}
	}
	return e.lastAgentText(), vendorErr
}

func (e *Engine) send(req *http.Request) (int, []byte, error) {
	// Log the outbound JSON request.
	if req.Body != nil {
		reqBody, err := io.ReadAll(req.Body)
		if err != nil {
			return 0, nil, err
		}
		req.Body.Close()
		e.Host.APILogf(">>> %s %s\n%s", req.Method, req.URL, string(reqBody))
		req.Body = io.NopCloser(strings.NewReader(string(reqBody)))
	}

	resp, err := e.HTTP.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}

	// Log the inbound JSON response.
	e.Host.APILogf("<<< %d\n%s", resp.StatusCode, string(body))

	return resp.StatusCode, body, nil
}

// MaxToolRounds bounds how many times one prompt may bounce through tools.
//
// This is a loop bound, not a timeout: nothing here is cancelled and nothing
// runs concurrently. It exists because a model that keeps asking forever, or a
// tool that keeps producing an error the model keeps retrying, should stop
// being funny at some point and hand control back to the human.
const MaxToolRounds = 16

// Ask is the loop of the chapter: say it, send it, run whatever the model
// asked for, send the results back, and keep going until it stops asking.
//
// The thing that surprises people is step 5. A tool call is not the end of a
// turn, it is the middle of one. The turn ends when the model replies without
// asking for anything.
func (e *Engine) Ask(text string) (string, error) {
	if err := e.Say(text); err != nil {
		return "", err
	}

	var reply string
	for round := 0; ; round++ {
		var err error
		reply, err = e.Turn()
		if err != nil {
			return "", err
		}

		// DISPATCH ON BLOCK TYPE. This is where Chapter 1's deferred type
		// filter finally bites: the model's reply is a LIST of typed parts,
		// and one message can carry both prose and a request to run something.
		// An implementation that walks the list looking only at text never
		// sees the call, reports the prose, and silently does nothing.
		calls := e.PendingCalls()
		if len(calls) == 0 {
			break
		}

		if round >= MaxToolRounds {
			if err := e.Record(common.Event{Type: common.ErrorOccurred, Error: &common.ErrorData{
				Message: fmt.Sprintf("stopped after %d rounds of tool calls", MaxToolRounds),
			}}); err != nil {
				return "", err
			}
			break
		}

		// SEQUENTIALLY, in the order the model asked for them. This is a
		// choice, not an oversight. Ordering is observable to the model, and a
		// shell command that changes the working tree changes what the next
		// tool sees.
		for _, call := range calls {
			if err := e.Execute(call); err != nil {
				return "", err
			}
		}
	}

	if err := e.Save(); err != nil {
		return "", err
	}
	return reply, nil
}

// PendingCalls returns the tool calls that have been asked for and not yet
// answered, in the order they were asked.
//
// It is computed from the dialogue rather than stored, for the same reason
// everything else in Chapter 2 is computed from the dialogue: the log is the
// truth, and a second copy of the truth is a second thing to get wrong.
func (e *Engine) PendingCalls() []common.ToolCallPart {
	answered := map[string]bool{}
	var calls []common.ToolCallPart
	for _, entry := range e.Ctx.Dialogue {
		for _, p := range entry.Parts {
			switch v := p.(type) {
			case common.ToolCallPart:
				calls = append(calls, v)
			case common.ToolResultPart:
				answered[v.CallID] = true
			}
		}
	}
	var out []common.ToolCallPart
	for _, c := range calls {
		if !answered[c.CallID] {
			out = append(out, c)
		}
	}
	return out
}

// Execute runs one tool call and records its result.
//
// EVERY call records a common.ToolReturned event, including the ones that fail. A
// failure is a RESULT, not an absence: the model asked a question and the
// answer is "that did not work, here is why". Dropping the result instead —
// or panicking, or ending the turn — leaves the model waiting for an answer to
// a question it can see it asked, and is the single most common way an agent
// locks up.
//
// THIS IS THE DISPATCH SITE, and the job model lives here and nowhere else.
// Before the tool runs it has a handle, an output file and a status; the tool
// runs on its own goroutine; and this goroutine waits on the JOB — until it
// is done, or the delay passes, or the pattern appears — rather than in the
// tool. Compare Chapter 3's version: the handler ran right here, and if it
// never came back, neither did the agent. Nothing about any of the six tools
// made that so. This function did.
//
// The supervision tools and tool_limits are the exception, marked NoJob: they
// act on jobs rather than being jobs, and they run inline.
func (e *Engine) Execute(call common.ToolCallPart) error {
	// common.Limits are resolved BEFORE the common.ToolCalled record, so that a pending
	// tool_limits is consumed by this call whether or not the tool exists.
	// A bad pattern is a tool error like any other: reported, not fatal.
	limits, fromPending, limErr := e.Jobs.Take(call.Args)

	tool, err := e.Tools.Lookup(call.Name)
	if err == nil {
		err = limErr
	}
	if err != nil || tool.NoJob {
		// The dispatch record: the agent decided to run this. It carries no
		// dialogue content of its own — the model already knows it asked.
		if err := e.Record(common.Event{Type: common.ToolCalled, Tool: &common.ToolData{
			CallID: call.CallID, Name: call.Name, Args: call.Args,
		}}); err != nil {
			return err
		}
		c := &common.Call{Host: e.Host, Jobs: e.Jobs, Limits: limits}
		var out string
		if err == nil {
			out, err = tool.Run(c, call.Args)
		}
		isError := false
		if err != nil {
			isError, out = true, err.Error()
		}
		if fromPending {
			out = pendingNote(call.Name, limits) + out
		}
		if err := e.Record(common.Event{Type: common.ToolReturned, Tool: &common.ToolData{
			CallID: call.CallID, Name: call.Name, Args: call.Args,
			Parts: common.PartList{common.TextPart{Text: out}}, IsError: isError,
		}}); err != nil {
			return err
		}
		for _, ev := range c.Events {
			if err := e.Record(ev); err != nil {
				return err
			}
		}
		return nil
	}

	// The handle is allocated here, for every tool, before the dispatcher
	// knows anything about what the tool will do. read_file gets one. A
	// read_file on an NFS mount that has gone away hangs exactly as well as
	// a shell command does.
	job, err := e.Jobs.Start(call.Name, call.CallID)
	if err != nil {
		return err
	}
	if err := e.Record(common.Event{Type: common.ToolCalled, Tool: &common.ToolData{
		CallID: call.CallID, Name: call.Name, Args: call.Args, Job: job.Data(),
	}}); err != nil {
		return err
	}

	// The tool runs on its own goroutine and reports into the job. There is
	// no recover here: a panic in a tool is an invariant violation, and an
	// invariant violation takes the process down loudly, as it should.
	go func() {
		out, err := tool.Run(&common.Call{Host: e.Host, Job: job, Jobs: e.Jobs, Limits: limits}, call.Args)
		job.Finish(out, err)
	}()

	// The wait is on the job, not the tool. If the tool never returns, this
	// returns anyway, with a handle, and the model decides what happens next.
	reason := job.Wait(limits)
	out := job.Report(reason, limits)
	if fromPending {
		out = pendingNote(call.Name, limits) + out
	}
	isError := job.Status() == common.StatusDone && job.Err() != nil

	return e.Record(common.Event{Type: common.ToolReturned, Tool: &common.ToolData{
		CallID:  call.CallID,
		Name:    call.Name,
		Args:    call.Args,
		Parts:   common.PartList{common.TextPart{Text: out}},
		IsError: isError,
		Job:     job.Data(),
	}})
}

// pendingNote heads the report of whichever call consumed a pending
// tool_limits. The limits are one-shot and land on the next call whatever it
// is. That is the design, and the footgun in it is that landing on the wrong
// call used to be silent: the model saw a truncated build two calls later
// with no cause in sight. This line puts the cause in the very result it
// produced. Loud, not different.
func pendingNote(tool string, l common.Limits) string {
	return fmt.Sprintf("[tool_limits consumed by this %s call: %s]\n", tool, l)
}

// Shutdown ends every job that is still running, and says so in the log.
//
// This is what happens to a job nobody killed when the agent exits: it is
// killed then, by the agent, deliberately, with a record. The alternative —
// letting it outlive the process that started it — leaves the human with a
// process they cannot find from any transcript.
func (e *Engine) Shutdown() error {
	for _, j := range e.Jobs.Running() {
		if !j.Kill("shutdown") {
			continue
		}
		data := j.Data()
		data.Reason = "shutdown"
		if err := e.Record(common.Event{Type: common.JobKilled, Job: data}); err != nil {
			return err
		}
	}
	return e.Save()
}

func (e *Engine) Save() error { return e.Log.SaveFile(e.Path) }

func (e *Engine) lastAgentText() string {
	for i := len(e.Ctx.Dialogue) - 1; i >= 0; i-- {
		if e.Ctx.Dialogue[i].Actor != common.ActorAgent {
			continue
		}
		var b strings.Builder
		for _, p := range e.Ctx.Dialogue[i].Parts {
			if t, ok := p.(common.TextPart); ok {
				b.WriteString(t.Text)
			}
		}
		return b.String()
	}
	return ""
}

// RenderOnly plays a log to a context and renders it, making no network call.
// This is what makes replay, redaction, ephemera and the seam into byte
// comparisons — and if your architecture cannot offer it cheaply, your context
// is not actually separate from your transport.
func RenderOnly(path string, cfg common.Config) ([]byte, error) {
	log, err := common.LoadLogFile(path)
	if err != nil {
		return nil, err
	}
	ctx, err := log.Replay()
	if err != nil {
		return nil, err
	}
	renderer, _, err := SeamFor(cfg.Vendor)
	if err != nil {
		return nil, err
	}
	req, err := renderer.Render(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("render: %w", err)
	}
	return common.BodyOf(req)
}
