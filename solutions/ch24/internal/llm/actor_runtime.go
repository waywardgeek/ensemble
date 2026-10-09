package llm

import (
	"context"
	"errors"
	"fmt"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// A typed constant preserves errors.Is identity without mutable package state.
type stoppedError string

func (e stoppedError) Error() string { return string(e) }

const ErrActorStopped stoppedError = "actor stopped"

// start gives an actor one owner for its entire lifetime. Ask starts lazily;
// explicit Run uses the same owner rather than creating another consumer.
func (a *Actor) start(ctx context.Context) {
	a.lifeMu.Lock()
	defer a.lifeMu.Unlock()
	if a.started {
		return
	}
	a.started = true
	a.ctx, a.cancel = context.WithCancel(ctx)
	go a.run()
}

// Run starts the owner if needed, then blocks until it stops. Cancelling any
// explicit Run stops the shared actor. An actor cannot be restarted after stop.
func (a *Actor) Run(ctx context.Context) {
	a.start(ctx)
	select {
	case <-ctx.Done():
		a.stop()
		<-a.done
	case <-a.done:
	}
}
func (a *Actor) stop() {
	a.lifeMu.Lock()
	a.cancel()
	a.lifeMu.Unlock()
}
func (a *Actor) run() {
	defer func() {
		a.cancel()
		// Kill managed processes before joining workers. An arbitrary Go tool
		// may not support cancellation; shutdown honestly waits for it to return.
		jobs := a.eng.Jobs.Running()
		for _, j := range jobs {
			// Kill intentionally suppresses late output. StopProcess also covers
			// a process attached after this point, without discarding Go results.
			j.StopProcess("shutdown")
		}
		a.workers.Wait()
		for _, j := range jobs {
			if j.Status() == common.StatusKilled {
				data := j.Data()
				data.Reason = "shutdown"
				a.shutdownErr = errors.Join(a.shutdownErr, a.eng.Record(common.Event{Type: common.JobKilled, Job: data}))
			}
		}
		for _, msg := range append(a.pending, a.mb.Drain()...) {
			switch m := msg.(type) {
			case common.ToolCompleted:
				a.shutdownErr = errors.Join(a.shutdownErr, a.completeTool(m))
			case common.ToolEvents:
				a.shutdownErr = errors.Join(a.shutdownErr, a.recordToolEvents(m.Events))
			}
		}
		a.shutdownErr = errors.Join(a.shutdownErr, a.eng.Save())
		close(a.done)
	}()
	for {
		if a.ctx.Err() != nil {
			return
		}
		if len(a.pending) > 0 {
			a.drain(a.ctx)
			continue
		}
		select {
		case <-a.ctx.Done():
			return
		case <-a.mb.Signal():
			a.drain(a.ctx)
		}
	}
}

// Ask is a synchronous adapter over the same mailbox used by GUI clients.
func (a *Actor) Ask(text string) (string, error) { return a.AskContext(context.Background(), text) }
func (a *Actor) AskContext(ctx context.Context, text string) (string, error) {
	return a.request(ctx, text, false)
}

// AttachContext serializes legacy ephemeral attachments with other actor work.
func (a *Actor) AttachContext(ctx context.Context, text string) error {
	_, err := a.request(ctx, text, true)
	return err
}
func (a *Actor) request(ctx context.Context, text string, ephemeral bool) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	a.start(context.Background())
	reply := make(chan common.TurnResult, 1)
	a.lifeMu.Lock()
	if a.ctx.Err() != nil {
		a.lifeMu.Unlock()
		return "", ErrActorStopped
	}
	a.mb.Post(common.Request{Text: text, Ephemeral: ephemeral, Context: ctx, Reply: reply})
	a.lifeMu.Unlock()
	select {
	case result := <-reply:
		return result.Text, result.Err
	case <-ctx.Done():
		return "", ctx.Err()
	case <-a.done:
		// Prefer a completed response if shutdown raced its delivery.
		select {
		case result := <-reply:
			return result.Text, result.Err
		default:
			return "", ErrActorStopped
		}
	}
}
func (a *Actor) handleRequest(r common.Request) {
	if err := r.Context.Err(); err != nil {
		r.Reply <- common.TurnResult{Err: err}
		return
	}
	if r.Ephemeral {
		r.Reply <- common.TurnResult{Err: a.eng.Attach(r.Text)}
		return
	}
	ctx, cancel := context.WithCancel(a.ctx)
	stop := context.AfterFunc(r.Context, cancel)
	defer func() { stop(); cancel() }()
	a.turnCtx = ctx
	a.result = common.TurnResult{}
	a.handleUserMessage(common.UserMessage{Text: r.Text})
	r.Reply <- a.result
}
func (a *Actor) endTurn(text string, err error) {
	err = errors.Join(err, a.eng.Save())
	a.result = common.TurnResult{Text: text, Err: err}
	ended := common.TurnEnded{Text: text}
	if err != nil {
		ended.Err = err.Error()
	}
	a.notify(ended)
}
func (a *Actor) completeTool(m common.ToolCompleted) error {
	if m.Tool != nil {
		if err := a.eng.Record(common.Event{Type: common.ToolReturned, Tool: m.Tool}); err != nil {
			return fmt.Errorf("tool result: %w", err)
		}
	}
	a.notify(common.ToolFinished{CallID: m.CallID, Result: m.Result, IsError: m.IsError})
	// A tool result is NOT model text, and it used to be announced twice:
	// once above, correctly, and once more wrapped in a TextPart. Anything
	// downstream reading that second event believed the model had said it.
	// The speech channel believed it too, and read the contents of every
	// file out loud.
	//
	// The tool card renders from ToolFinished above, so nothing visible
	// depended on the duplicate. Do not reintroduce it.
	return nil
}

func (a *Actor) recordToolEvents(events []common.Event) error {
	for _, ev := range events {
		if err := a.eng.Record(ev); err != nil {
			return err
		}
	}
	return nil
}

// Shutdown joins execution before the final save; repeated calls are safe.
// Managed processes are killed deliberately, with a record: letting a process
// outlive the agent leaves the human with a process no transcript can explain.
func (a *Actor) Shutdown() error {
	a.start(context.Background())
	a.stop()
	<-a.done
	return a.shutdownErr
}

// DefaultToolRounds is the zero-value policy, not a second orchestration cap.
const DefaultToolRounds = 200

type ToolRoundLimitError struct{ Limit int }

func (e *ToolRoundLimitError) Error() string {
	return fmt.Sprintf("stopped after %d rounds of tool calls", e.Limit)
}
func (a *Actor) toolRoundLimit() int {
	n := a.eng.Cfg.MaxToolRounds
	if a.eng.ToolRoundLimit != nil {
		n = a.eng.ToolRoundLimit()
	}
	if n <= 0 {
		return DefaultToolRounds
	}
	return n
}
