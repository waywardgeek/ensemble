package main

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
)

type Engine struct {
	Log  *Log
	Ctx  *Context
	Cfg  Config
	HTTP *http.Client
	Path string // where the log is persisted, so `dump` can find it
}

func NewEngine(cfg Config, path string) *Engine {
	return &Engine{
		Log:  NewLog(),
		Ctx:  NewContext(),
		Cfg:  cfg,
		HTTP: &http.Client{Timeout: 120 * time.Second},
		Path: path,
	}
}

// record appends to the log and advances the context. One path in.
func (e *Engine) record(ev Event) error {
	stored := e.Log.Append(ev)
	return e.Ctx.Apply(stored)
}

// Say records a human prompt.
func (e *Engine) Say(text string) error {
	return e.record(Event{Type: MessageReceived, Message: &MessageData{
		Actor: ActorHuman, Parts: PartList{TextPart{Text: text}},
	}})
}

// Attach records volatile data — the time, the screen, live status.
//
// It is an ordinary MessageReceived with Actor: System. The reducer is what
// decides it is ephemeral rather than dialogue, which is the chapter's rule
// about classification made concrete: the capture site does not know, and
// cannot know, what an arriving message means.
func (e *Engine) Attach(text string) error {
	return e.record(Event{Type: MessageReceived, Message: &MessageData{
		Actor: ActorSystem, Parts: PartList{TextPart{Text: text}},
	}})
}

// Turn renders the current context, sends it, and folds the response back in.
func (e *Engine) Turn() (string, error) {
	renderer, parser, err := SeamFor(e.Cfg.Vendor)
	if err != nil {
		return "", err
	}

	// RENDER BEFORE RECORDING RequestSent.
	//
	// Reverse these two lines and the bug is subtle and expensive: the reducer
	// clears pending ephemera on RequestSent, so recording first means the
	// renderer never sees them and the volatile data is silently never
	// delivered. Nothing errors. The model just quietly does not know what
	// time it is. Chapter 4 hits the identical ordering trap with hints.
	req, err := renderer.Render(e.Ctx, e.Cfg)
	if err != nil {
		return "", err
	}
	if err := e.record(Event{Type: RequestSent, Request: &RequestData{To: Provenance{
		Vendor: e.Cfg.Vendor, Model: e.Cfg.Model, Surface: e.Cfg.Surface,
	}}}); err != nil {
		return "", err
	}

	status, body, err := e.send(req)
	if err != nil {
		// A transport failure is infrastructure: it ends the turn.
		_ = e.record(Event{Type: ErrorOccurred, Error: &ErrorData{Message: err.Error()}})
		return "", err
	}

	events, err := parser.Parse(status, body)
	if err != nil {
		return "", err
	}
	var vendorErr error
	for _, ev := range events {
		if err := e.record(ev); err != nil {
			return "", err
		}
		if ev.Type == ErrorOccurred && ev.Error != nil {
			// The event is the record; this is the report. Without it a 404
			// on the model name prints {"assistant":""} and exits 0 — measured
			// live against Gemini, and invisible unless you read the log.
			vendorErr = fmt.Errorf("%s: %s", e.Cfg.Vendor, ev.Error.Message)
		}
	}
	return e.lastAgentText(), vendorErr
}

func (e *Engine) send(req *http.Request) (int, []byte, error) {
	resp, err := e.HTTP.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}
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
		calls := e.pendingCalls()
		if len(calls) == 0 {
			break
		}

		if round >= MaxToolRounds {
			if err := e.record(Event{Type: ErrorOccurred, Error: &ErrorData{
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

// pendingCalls returns the tool calls that have been asked for and not yet
// answered, in the order they were asked.
//
// It is computed from the dialogue rather than stored, for the same reason
// everything else in Chapter 2 is computed from the dialogue: the log is the
// truth, and a second copy of the truth is a second thing to get wrong.
func (e *Engine) pendingCalls() []ToolCallPart {
	answered := map[string]bool{}
	var calls []ToolCallPart
	for _, entry := range e.Ctx.Dialogue {
		for _, p := range entry.Parts {
			switch v := p.(type) {
			case ToolCallPart:
				calls = append(calls, v)
			case ToolResultPart:
				answered[v.CallID] = true
			}
		}
	}
	var out []ToolCallPart
	for _, c := range calls {
		if !answered[c.CallID] {
			out = append(out, c)
		}
	}
	return out
}

// Execute runs one tool call and records its result.
//
// EVERY call records a ToolReturned event, including the ones that fail. A
// failure is a RESULT, not an absence: the model asked a question and the
// answer is "that did not work, here is why". Dropping the result instead —
// or panicking, or ending the turn — leaves the model waiting for an answer to
// a question it can see it asked, and is the single most common way an agent
// locks up.
//
// Note what is NOT here: no goroutine, no handle, no timeout, no way to look
// at this while it runs. Execute blocks until the tool is finished. That is
// the honest simple form, and it is enough to write real code with.
func (e *Engine) Execute(call ToolCallPart) error {
	// The dispatch record: the agent decided to run this. It carries no
	// dialogue content of its own — the model already knows it asked.
	if err := e.record(Event{Type: ToolCalled, Tool: &ToolData{
		CallID: call.CallID, Name: call.Name, Args: call.Args,
	}}); err != nil {
		return err
	}

	out, err := Dispatch(call.Name, call.Args)
	isError := false
	if err != nil {
		isError, out = true, err.Error()
	}

	return e.record(Event{Type: ToolReturned, Tool: &ToolData{
		CallID:  call.CallID,
		Name:    call.Name,
		Args:    call.Args,
		Parts:   PartList{TextPart{Text: out}},
		IsError: isError,
	}})
}

func (e *Engine) Save() error { return e.Log.SaveFile(e.Path) }

func (e *Engine) lastAgentText() string {
	for i := len(e.Ctx.Dialogue) - 1; i >= 0; i-- {
		if e.Ctx.Dialogue[i].Actor != ActorAgent {
			continue
		}
		var b strings.Builder
		for _, p := range e.Ctx.Dialogue[i].Parts {
			if t, ok := p.(TextPart); ok {
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
func RenderOnly(path string, cfg Config) ([]byte, error) {
	log, err := LoadLogFile(path)
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
	return BodyOf(req)
}
