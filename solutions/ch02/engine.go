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

// Ask is the loop of the chapter: say it, send everything, hand back the text.
func (e *Engine) Ask(text string) (string, error) {
	if err := e.Say(text); err != nil {
		return "", err
	}
	reply, err := e.Turn()
	if err != nil {
		return "", err
	}
	if err := e.Save(); err != nil {
		return "", err
	}
	return reply, nil
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
