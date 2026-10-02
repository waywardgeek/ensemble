package llm

import "github.com/waywardgeek/ensemble/agent/internal/common"

// Choosing what to compress.
//
// The grain of memory is the grain of checkpointing: a memory covers one or
// more WHOLE micro_handoff segments. That is a selection problem — how many
// whole units — rather than a splitting problem, and it removes a failure
// mode instead of tuning one. A memory spanning the second half of one task
// and the first half of the next is worse than a larger memory covering one
// task properly. A coherent 1.3x beats an incoherent 2x.
//
// Splitting a segment is therefore an explicitly degraded path, taken only
// when there is no whole-segment range to take, which happens when the agent
// has been talking for a long time without ever checkpointing. It is better
// than the alternative of letting the context grow without bound, and it is
// marked so that anything downstream can tell the two apart.

// Segment is a run of conversation between two checkpoints.
//
// From and To are inclusive common.Seq bounds, which is what a Redacted span wants.
// Bytes is what the segment costs in the context right now.
type Segment struct {
	From  common.Seq
	To    common.Seq
	Bytes int
}

// Segments splits the live conversation at checkpoint boundaries.
//
// Only common.KindDialogue entries are conversation. Memory, skills, tool
// declarations and the checkpoint notes themselves are survivors: they are
// removed by their own verbs and are not anybody's to summarize.
//
// Anything still in the dialogue has never been summarized, because
// compaction REMOVES the span it compressed. So there is no "already done"
// bookkeeping to keep, and none to get wrong.
func Segments(c *common.Context) []Segment {
	var out []Segment
	var cur *Segment
	for _, e := range c.Dialogue {
		if e.Kind == common.KindHandoff {
			if cur != nil {
				out = append(out, *cur)
				cur = nil
			}
			continue
		}
		if e.Kind != common.KindDialogue {
			continue
		}
		n := entryBytes(e)
		if cur == nil {
			cur = &Segment{From: e.Seq, To: e.Seq, Bytes: n}
			continue
		}
		cur.To = e.Seq
		cur.Bytes += n
	}
	if cur != nil {
		out = append(out, *cur)
	}
	return out
}

func entryBytes(e common.Entry) int {
	n := 0
	for _, p := range e.Parts {
		switch v := p.(type) {
		case common.TextPart:
			n += len(v.Text)
		case common.ToolResultPart:
			for _, q := range v.Parts {
				if t, ok := q.(common.TextPart); ok {
					n += len(t.Text)
				}
			}
		}
	}
	return n
}

// Selection is a span chosen for compression.
type Selection struct {
	From common.Seq
	To   common.Seq
	// Bytes is what the selected span currently costs.
	Bytes int
	// Split is true when no whole-segment range was large enough and a
	// segment had to be cut in the middle. The degraded path, reported
	// rather than hidden so that a log can say it happened.
	Split bool
}

// SelectForCompression picks the oldest run of whole segments worth at least
// want bytes.
//
// Oldest first, always. The newest conversation is the one still being
// reasoned about, and compressing it would take away the detail the agent is
// actively using while leaving the stale material untouched.
//
// Returns ok=false when there is nothing worth compressing, which is the
// ordinary case on most checkpoints.
func SelectForCompression(c *common.Context, want int) (Selection, bool) {
	if want <= 0 {
		return Selection{}, false
	}
	segs := Segments(c)
	if len(segs) == 0 {
		return Selection{}, false
	}

	// The last segment is the live one: it has no closing checkpoint yet,
	// so it is what the agent is working on yet. Never compress it.
	closed := segs
	if len(segs) > 0 && !endsWithCheckpoint(c) {
		closed = segs[:len(segs)-1]
	}

	// Nothing is closed: the agent has been talking without ever
	// checkpointing. The live segment is work in progress, and compacting
	// it would take away the detail being reasoned with right now. So
	// nothing happens here, and nothing needs to: the forcing rule takes
	// away every tool but micro_handoff before the window runs out, which
	// turns this case into the ordinary one before it can become a crisis.
	if len(closed) == 0 {
		return Selection{}, false
	}

	// A single closed segment can be so large that handing it to a
	// compressor whole is its own failure: the span has to fit in the
	// compressor's own context, and a segment written over days need not.
	// This is the degraded path, and the only one that cuts.
	if closed[0].Bytes > maxSpan*want {
		return splitSegment(c, closed[0], want)
	}

	total := 0
	for i, s := range closed {
		total += s.Bytes
		if total >= want {
			return Selection{From: closed[0].From, To: closed[i].To, Bytes: total}, true
		}
	}

	// No run of whole segments reaches the target. If the closed segments
	// together are worth compressing at all, take them; this is still whole
	// segments, just fewer bytes than hoped.
	if total > 0 {
		return Selection{From: closed[0].From, To: closed[len(closed)-1].To, Bytes: total}, true
	}

	return Selection{}, false
}

// maxSpan is how many times the target a single segment may be before it is
// cut rather than taken whole.
const maxSpan = 8

// endsWithCheckpoint reports whether the newest conversation entry is closed
// by a checkpoint.
func endsWithCheckpoint(c *common.Context) bool {
	for i := len(c.Dialogue) - 1; i >= 0; i-- {
		switch c.Dialogue[i].Kind {
		case common.KindHandoff:
			return true
		case common.KindDialogue:
			return false
		}
	}
	return false
}

// splitLive is the degraded path: cut the only segment there is, at the entry
// that reaches the target.
// splitSegment cuts inside one closed segment, taking entries from its start
// until the target is reached.
//
// This is the only path that breaks the whole-segment rule, and it reports
// that it did by setting Split. A memory made this way covers a piece of a
// session rather than a session, so the summary it produces begins in the
// middle of something. That is a real cost, and it is still better than
// handing a compressor a span that will not fit in its own context.
func splitSegment(c *common.Context, seg Segment, want int) (Selection, bool) {
	var from, to common.Seq
	n, started := 0, false
	for _, e := range c.Dialogue {
		if e.Kind != common.KindDialogue || e.Seq < seg.From || e.Seq > seg.To {
			continue
		}
		if !started {
			from, started = e.Seq, true
		}
		n += entryBytes(e)
		to = e.Seq
		if n >= want {
			return Selection{From: from, To: to, Bytes: n, Split: true}, true
		}
	}
	if !started {
		return Selection{}, false
	}
	return Selection{From: from, To: to, Bytes: n, Split: true}, true
}
