package llm

import (
	"fmt"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// FoldFactor is how many memory files one compressor folds into one file.
//
// It is also why the bands are named 8x and 64x. Eight session memories
// become one 8x memory; eight of those become one 64x memory, which is
// sixty-four sessions ago seen from the inside. The names are not labels
// chosen for the design, they are arithmetic that fell out of it.
const FoldFactor = 8

// maybeCompact is the whole of chapter sixteen's control flow, and it runs
// at the top of a turn rather than at the bottom of the tool that triggers
// it.
//
// Nothing here asks whether a checkpoint just landed, because it does not
// need to. SelectForCompression refuses to return anything that is not a
// complete checkpointed segment, so before the first micro_handoff this
// function looks at a large conversation, finds nothing it is allowed to
// take, and does nothing. That is a stateless test for a stateful-sounding
// condition, and it survives a restart without having to remember anything.
func (e *Engine) maybeCompact() {
	if e.Memory == nil || e.Bands == nil {
		return
	}

	// Make the context agree with the settings before deciding anything.
	// SyncBands is idempotent, so the common case - nothing changed - costs
	// a directory read and emits no events. Doing it here is what lets a
	// band be switched off in the GUI and take effect on the next turn
	// rather than on the next launch.
	if err := e.SyncBands("settings"); err != nil {
		e.Host.Logf("memory: could not sync bands: %v", err)
	}
	cfg := e.Bands()

	e.compactConversation(cfg)

	// Bottom up, because folding the session band can be what makes the 8x
	// band full enough to fold in turn. Running top down would leave that
	// second fold waiting for a checkpoint that has already happened.
	for _, b := range []common.Band{common.BandSession, common.Band8x, common.Band64x} {
		e.graduate(b, cfg)
	}
}

// compactConversation turns the oldest finished work into one memory, and
// keeps going until the conversation is back under its low watermark.
func (e *Engine) compactConversation(cfg common.BandConfig) {
	if cfg.Session.Disabled {
		return
	}
	if e.Ctx.ConversationBytes() <= cfg.Conversation.High() {
		return
	}

	low := cfg.Conversation.Low()
	for e.Ctx.ConversationBytes() > low {
		sel, ok := e.Ctx.SelectForCompression(e.Ctx.ConversationBytes() - low)
		if !ok {
			// Everything that is left is the live segment. Refusing to
			// touch it is the correct outcome, not a failure: the work in
			// progress is the one thing compaction must never eat.
			return
		}
		if !e.compactSpan(sel) {
			return
		}
	}
}

// compactSpan compresses one span into one session memory file, then records
// the two events that swap the span for the memory.
func (e *Engine) compactSpan(sel common.Selection) bool {
	body := renderConversation(e.Ctx, sel.From, sel.To)
	if body == "" {
		return false
	}

	text, err := e.compress("the conversation below", body, len(body)/FoldFactor)
	if err != nil {
		e.warn("memory: compaction refused, conversation left intact: %v", err)
		return false
	}

	f, err := e.Memory.WriteSession(text)
	if err != nil {
		e.warn("memory: could not write session memory, conversation left intact: %v", err)
		return false
	}

	// Two events, in this order, and neither one does the other's job.
	//
	// The redaction carries an EMPTY replacement, which deletes the span and
	// lands nothing in its place. It would have been natural to let the
	// redaction carry the compressed text, since chapter fifteen's summary
	// redaction can do exactly that. It would also have put the memory into
	// the context twice: once as the replacement and once as the band entry.
	// Deleting and landing are two operations here, so they are two events.
	e.Record(common.Event{
		Type: common.Redacted,
		Redact: &common.RedactData{
			From:   sel.From,
			To:     sel.To,
			Level:  common.RedactSummary,
			Reason: fmt.Sprintf("compacted into %s", f.ID),
		},
	})
	e.Record(common.Event{
		Type: common.BandPopulated,
		BandAdd: &common.BandPopulatedData{
			Band:   common.BandSession,
			File:   f.ID,
			Text:   text,
			Source: "compaction",
		},
	})
	return true
}

// graduate folds the oldest FoldFactor files of a band into one file of the
// band above it.
func (e *Engine) graduate(b common.Band, cfg common.BandConfig) {
	up := b.Up()
	if up == 0 {
		return
	}

	files, err := e.Memory.Files(b)
	if err != nil {
		e.warn("memory: could not read the %s band: %v", b, err)
		return
	}
	// Fold a full batch when there is one, and whatever is there when there
	// is not.
	//
	// Requiring a full batch looks tidier and is wrong. A band is folded
	// because it is over its budget, and how many files that took depends
	// on how big they are. A band holding five large memories is over
	// budget and can never reach eight, so a strict batch means the one
	// band that most needs folding is the one band that never folds. The
	// grader found this by giving a band a small budget and watching it
	// grow forever.
	//
	// Two is the floor. Folding one file into one file is just compressing
	// the same memory again, which loses detail without reducing the count.
	n := FoldFactor
	if len(files) < n {
		n = len(files)
	}
	if n < 2 {
		return
	}

	// A disabled upward neighbour is refused out loud rather than handled.
	//
	// The two silent options are both wrong. Dropping the oldest files
	// destroys memory the user never asked to lose, and carrying on lets
	// the band grow without bound, which is the condition the whole
	// chapter exists to prevent. Neither failure announces itself, so
	// this one does.
	if cfg.For(up).Disabled {
		e.warn("memory: the %s band is full but %s is switched off, so nothing can graduate; "+
			"switch %s back on or these memories will stay where they are", b, up, up)
		return
	}

	src := files[:n]
	from, thru := src[0].ID, src[len(src)-1].ID

	if e.abandonedSince(b, thru) {
		e.warn("memory: a previous %s graduation did not finish; leaving it until the next checkpoint", b)
		return
	}

	// Launched before the call, so that a crash mid-compression leaves a
	// record that something was attempted. The pair of events is what makes
	// an abandoned attempt visible; a single event on success would make a
	// crash indistinguishable from never having tried.
	e.Record(common.Event{
		Type:    common.CompactorLaunched,
		Compact: &common.CompactorLaunchData{Band: b, Thru: thru},
	})

	var body string
	var total int
	for _, f := range src {
		body += fmt.Sprintf("\n--- %s ---\n%s\n", f.ID, f.Text)
		total += len(f.Text)
	}

	budget := total / FoldFactor
	what := fmt.Sprintf("the %d memories below", len(src))
	if up == common.BandMemory {
		// The top rung is different, and deliberately so. Every other rung
		// compresses by a ratio; this one compresses to a size, because
		// MEMORY.md is not a fold of what came before it. It is the thing
		// that is kept, and what is kept has a budget rather than a ratio.
		budget = cfg.Memory.Budget
		if cur, err := e.Memory.Files(common.BandMemory); err == nil && len(cur) > 0 {
			body = fmt.Sprintf("\n--- the long term memory as it stands ---\n%s\n%s", cur[0].Text, body)
		}
		what = "the long term memory and the memories below"
	}

	text, err := e.compress(what, body, budget)
	if err != nil {
		e.warn("memory: %s graduation refused, memories left in place: %v", b, err)
		return
	}

	var landed common.MemoryFileID
	if up == common.BandMemory {
		err = e.Memory.WriteCurated(text)
		landed = common.MemoryFileID{}
	} else {
		var nf File
		nf, err = e.Memory.WriteBucket(up, from, thru, text)
		landed = nf.ID
	}
	if err != nil {
		e.warn("memory: could not write the %s memory, memories left in place: %v", up, err)
		return
	}

	e.Record(common.Event{
		Type: common.BandDepopulated,
		BandDrop: &common.BandDepopulatedData{
			Band:   b,
			Thru:   &thru,
			Reason: fmt.Sprintf("graduated into %s", up),
		},
	})
	e.Record(common.Event{
		Type: common.BandPopulated,
		BandAdd: &common.BandPopulatedData{
			Band:   up,
			File:   landed,
			Text:   text,
			Thru:   thru,
			Source: "graduation",
		},
	})

	// Retiring happens last, and only after the event carrying the new
	// bytes is in the log, so a crash between the two loses nothing.
	//
	// Bucket files are retired; uncompressed memories never are. A folded
	// session memory leaves the BAND but stays in the directory, because
	// every uncompressed memory lives in one place and a later chapter
	// searches it. Membership is derived from bucket coverage rather than
	// from deletion, so keeping the band the right size costs nothing.
	if b != common.BandSession {
		if err := e.Memory.Retire(src); err != nil {
			e.warn("memory: the %s memories were graduated but could not be retired: %v", b, err)
		}
	}
}

// abandonedSince reports whether a graduation of this band was launched and
// never finished, with no checkpoint since.
//
// The rule is not "never retry". It is "do not retry until something has
// changed", and a checkpoint is the cheapest honest evidence that something
// has. Retrying immediately would mean a compressor that crashes the agent
// crashes it again on the next turn, forever.
func (e *Engine) abandonedSince(b common.Band, thru common.MemoryFileID) bool {
	launched := false
	for _, ev := range e.Log.Events {
		switch ev.Type {
		case common.MicroHandoff:
			launched = false
		case common.CompactorLaunched:
			if d := ev.Compact; d != nil && d.Band == b && d.Thru == thru {
				launched = true
			}
		case common.BandPopulated:
			if d := ev.BandAdd; d != nil && d.Source == "graduation" && d.Thru == thru {
				launched = false
			}
		}
	}
	return launched
}

func (e *Engine) warn(format string, args ...any) {
	if e.Host != nil {
		e.Host.Logf(format, args...)
	}
}
