package common

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
// From and To are inclusive Seq bounds, which is what a Redacted span wants.
// Bytes is what the segment costs in the context right now.
type Segment struct {
	From  Seq
	To    Seq
	Bytes int
}

// Segments splits the live conversation at checkpoint boundaries.
//
// Only KindDialogue entries are conversation. Memory, skills, tool
// declarations and the checkpoint notes themselves are survivors: they are
// removed by their own verbs and are not anybody's to summarize.
//
// Anything still in the dialogue has never been summarized, because
// compaction REMOVES the span it compressed. So there is no "already done"
// bookkeeping to keep, and none to get wrong.
func (c *Context) Segments() []Segment {
	var out []Segment
	var cur *Segment
	for _, e := range c.Dialogue {
		if e.Kind == KindHandoff {
			if cur != nil {
				out = append(out, *cur)
				cur = nil
			}
			continue
		}
		if e.Kind != KindDialogue {
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

func entryBytes(e Entry) int {
	n := 0
	for _, p := range e.Parts {
		switch v := p.(type) {
		case TextPart:
			n += len(v.Text)
		case ToolResultPart:
			for _, q := range v.Parts {
				if t, ok := q.(TextPart); ok {
					n += len(t.Text)
				}
			}
		}
	}
	return n
}

// Selection is a span chosen for compression.
type Selection struct {
	From Seq
	To   Seq
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
func (c *Context) SelectForCompression(want int) (Selection, bool) {
	if want <= 0 {
		return Selection{}, false
	}
	segs := c.Segments()
	if len(segs) == 0 {
		return Selection{}, false
	}

	// The last segment is the live one: it has no closing checkpoint yet,
	// so it is what the agent is working on yet. Never compress it.
	closed := segs
	if len(segs) > 0 && !c.endsWithCheckpoint() {
		closed = segs[:len(segs)-1]
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

	// Nothing is closed: the agent has been talking without ever
	// checkpointing. Cut the live segment rather than let the context grow
	// without bound, and say that is what happened.
	return c.splitLive(want)
}

// endsWithCheckpoint reports whether the newest conversation entry is closed
// by a checkpoint.
func (c *Context) endsWithCheckpoint() bool {
	for i := len(c.Dialogue) - 1; i >= 0; i-- {
		switch c.Dialogue[i].Kind {
		case KindHandoff:
			return true
		case KindDialogue:
			return false
		}
	}
	return false
}

// splitLive is the degraded path: cut the only segment there is, at the entry
// that reaches the target.
func (c *Context) splitLive(want int) (Selection, bool) {
	var from, to Seq
	n, started := 0, false
	for _, e := range c.Dialogue {
		if e.Kind != KindDialogue {
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
	return Selection{}, false
}
