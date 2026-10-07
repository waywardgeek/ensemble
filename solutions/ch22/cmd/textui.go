package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// textObserver renders observations as plain text for the terminal.
//
// It is an Observer and nothing more: it never calls back into the agent. That
// is what makes a text front end a drop-in alternative to the GUI rather than a
// second way to drive the agent. The CLI and the GUI watch the SAME actor loop,
// so a bug that reproduces in one is a bug in the agent, and a bug that
// reproduces in only one is a bug in that front end. Before this existed the
// terminal ran a bare engine with no actor, skills, recall or save file, so a
// GUI bug had no cheap reproduction.
type textObserver struct {
	mu      sync.Mutex
	out     io.Writer // the answer
	errOut  io.Writer // commentary about the answer
	verbose bool

	// streamed records part ids that already arrived as deltas, so the
	// authoritative final does not print the same text a second time.
	//
	// This is the same hazard the GUI hit: the Responses surface reuses delta
	// ids for finals, and a renderer that draws both shows every answer twice.
	// A renderer must choose one of the two, and deltas are the one the reader
	// has already seen.
	streamed map[uint64]bool

	section string // heading currently open, "" when none
	midOut  bool   // cursor sits after partial output on out
	midErr  bool   // cursor sits after partial output on errOut
}

// newTextObserver splits a turn across two streams.
//
// Reply text goes to OUT because that is the answer, and an answer belongs on
// stdout. Reasoning, tool calls and state go to ERROUT as commentary. Piping
// the binary therefore still yields exactly what the assistant said and
// nothing else, while a human at a terminal sees the whole turn being built.
// This rule is inherited from the bare-engine chat loop this front end
// replaced, and it is the reason the CLI is usable in a shell pipeline.
func newTextObserver(out, errOut io.Writer, verbose bool) *textObserver {
	return &textObserver{out: out, errOut: errOut, verbose: verbose, streamed: map[uint64]bool{}}
}

func deltaHeading(k common.DeltaKind) string {
	switch k {
	case common.DeltaText:
		return "assistant"
	case common.DeltaThinking:
		return "thinking"
	case common.DeltaReasoningSummary:
		return "reasoning"
	case common.DeltaToolCall:
		return "tool args"
	default:
		return k.String()
	}
}

// Observe implements common.Observer.
func (t *textObserver) Observe(o common.Observation) {
	t.mu.Lock()
	defer t.mu.Unlock()

	switch v := o.(type) {
	case common.PartDelta:
		t.streamed[v.PartID] = true
		// Tool arguments stream as JSON fragments. Printing them live is the
		// noise that showed up in the GUI as a run of empty braces, and the
		// whole, valid input arrives a moment later on ToolDispatched.
		if v.Kind == common.DeltaToolCall && !t.verbose {
			return
		}
		if v.Chunk == "" {
			return
		}
		t.openSection(deltaHeading(v.Kind))
		if v.Kind == common.DeltaText {
			t.writeOut(v.Chunk)
		} else {
			t.writeErr(v.Chunk)
		}

	case common.PartFinal:
		if t.streamed[v.PartID] {
			return // already shown as deltas
		}
		t.renderFinal(v.Part)

	case common.ToolDispatched:
		t.closeSection()
		t.line(fmt.Sprintf("  -> %s %s", v.Name, truncate(compactJSON(v.Input), 160)))

	case common.ToolFinished:
		t.closeSection()
		tag := "ok"
		if v.IsError {
			tag = "ERROR"
		}
		t.line(fmt.Sprintf("  <- %s %s", tag, truncate(oneLine(v.Result), 160)))

	case common.StateChanged:
		if t.verbose {
			t.closeSection()
			t.line(fmt.Sprintf("  [%s -> %s]", v.From, v.To))
		}

	case common.TurnEnded:
		t.closeSection()
		if v.Err != "" {
			t.line("  [turn failed: " + v.Err + "]")
		}

	case common.ConversationCleared:
		t.closeSection()
		t.line("  [conversation cleared]")
	}
}

// renderFinal prints a part that never streamed. A non-streaming vendor emits
// one delta and then a final, so this is the uncommon path.
func (t *textObserver) renderFinal(p common.Part) {
	switch v := p.(type) {
	case common.TextPart:
		if v.Text == "" {
			return
		}
		t.openSection("assistant")
		t.writeOut(v.Text)
	case common.ToolCallPart:
		// ToolDispatched already reports the call.
	default:
		// Opaque reasoning blobs and redaction stubs have no text to show. An
		// opaque part with nothing to draw must draw nothing: the GUI once
		// rendered these as stray empty braces.
		if t.verbose {
			t.closeSection()
			t.line(fmt.Sprintf("  [%T]", v))
		}
	}
}

// openSection prints a label so the reader can tell reply text from thinking.
// The label is commentary, so it goes to errOut even when the content under it
// goes to stdout. A pipe therefore collects the text and none of the labels.
func (t *textObserver) openSection(name string) {
	if t.section == name {
		return
	}
	t.closeSection()
	fmt.Fprintf(t.errOut, "%s:\n", name)
	t.section = name
}

func (t *textObserver) closeSection() {
	if t.midOut {
		fmt.Fprintln(t.out)
		t.midOut = false
	}
	if t.midErr {
		fmt.Fprintln(t.errOut)
		t.midErr = false
	}
	t.section = ""
}

func (t *textObserver) writeOut(s string) {
	fmt.Fprint(t.out, s)
	t.midOut = !strings.HasSuffix(s, "\n")
}

func (t *textObserver) writeErr(s string) {
	fmt.Fprint(t.errOut, s)
	t.midErr = !strings.HasSuffix(s, "\n")
}

// line writes one complete line of commentary.
func (t *textObserver) line(s string) {
	fmt.Fprintln(t.errOut, s)
	t.midErr = false
}

func compactJSON(raw json.RawMessage) string {
	if len(raw) == 0 {
		return "{}"
	}
	var buf bytes.Buffer
	if err := json.Compact(&buf, raw); err != nil {
		return oneLine(string(raw))
	}
	return buf.String()
}

func oneLine(s string) string {
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	return strings.TrimSpace(s)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + fmt.Sprintf("... (%d bytes)", len(s))
}

// parseTextLine turns one typed line into a protocol message.
//
// Prose is a prompt. The two things prose cannot express are steering a turn
// that is ALREADY RUNNING and stopping it, so those get slash commands. The
// JSON-lines protocol is unchanged and still available to programs; this is a
// front end on top of it, not a second protocol.
func parseTextLine(line string) (msg stdinMsg, quit bool, err error) {
	if !strings.HasPrefix(line, "/") {
		kind, text := "prompt", line
		return stdinMsg{Kind: &kind, Text: &text}, false, nil
	}

	cmd, rest, _ := strings.Cut(line, " ")
	rest = strings.TrimSpace(rest)

	switch cmd {
	case "/quit", "/exit":
		return stdinMsg{}, true, nil

	case "/interrupt":
		kind := "interrupt"
		return stdinMsg{Kind: &kind}, false, nil

	case "/hint":
		if rest == "" {
			return stdinMsg{}, false, fmt.Errorf("/hint needs text: /hint be concise")
		}
		kind := "hint"
		return stdinMsg{Kind: &kind, Text: &rest}, false, nil

	case "/help":
		return stdinMsg{}, false, fmt.Errorf("commands: /hint <text>, /interrupt, /quit. anything else is a prompt")

	default:
		return stdinMsg{}, false, fmt.Errorf("unknown command %q. /help lists them", cmd)
	}
}
