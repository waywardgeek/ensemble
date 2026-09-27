package llm

import (
	"strings"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// recallConvoEntries is how many recent dialogue entries are quoted to the
// judge as context for what the user is actually asking about.
//
// It is small on purpose. The judge is deciding relevance, not participating;
// it needs enough to disambiguate a pronoun ("how did we fix that?") and no
// more. Quoting the whole conversation would make the judge call scale with
// the very thing recall exists to avoid scaling with.
const recallConvoEntries = 6

// recallConvoEntryBytes caps each quoted entry, so one enormous pasted log in
// the conversation cannot dominate the judge prompt.
const recallConvoEntryBytes = 600

// recallGeminiMarker labels retrieved material on vendors that have no
// mid-conversation system role, so it is still distinguishable from what the
// human actually typed.
const recallGeminiMarker = "[automatically retrieved from the archive; not written by the user]\n\n"

// recallText flattens a recall entry's stored parts into the text to render.
//
// It reads the parts recorded in the event log and concatenates them. It does
// NOT re-run retrieval, re-score anything, or consult the archive. The bytes
// were decided when the event was recorded; this is pure presentation, which
// is why replaying a log needs neither a model call nor a BM25 run.
func recallText(entry common.Entry) string {
	var sb strings.Builder
	for _, p := range entry.Parts {
		t, ok := p.(common.TextPart)
		if !ok {
			continue
		}
		sb.WriteString(t.Text)
	}
	return strings.TrimSpace(sb.String())
}

// attachRecall consults the recaller and, if it returns anything, lands the
// result in the Dialogue as its own permanent entry.
//
// Four properties of this function are the whole point of the chapter:
//
//   - It is best-effort. Recall failing is not a turn failing. Every exit
//     path below is a silent return, because an agent that cannot remember
//     is strictly better than an agent that refuses to talk.
//
//   - It never merges into the user's message. The material lands as a
//     separate entry with a separate kind, so the model can always tell what
//     the human said from what a retrieval system guessed might be useful.
//     Merging would forge the user's words, and a model that cannot trust the
//     user's turn to be the user's words cannot follow instructions in it.
//
//   - It records an event carrying BYTES, not a query. Replay reconstructs
//     the exact entry from the log without a model call, without a BM25 run,
//     and without the archive files still existing or still saying the same
//     thing. A query would make history a function of today's disk.
//
//   - It is permanent. The entry lands in the Dialogue and stays there for
//     the rest of the session.
func (e *Engine) attachRecall(query string) {
	if e == nil || e.Recall == nil {
		return
	}
	parts := e.Recall.Recall(query, e.recallConvo())
	if len(parts) == 0 {
		// Finding nothing is the expected outcome, not an error. Most turns
		// genuinely have no relevant history, and an agent that attaches
		// something every turn is an agent whose relevance filter is broken.
		return
	}
	_ = e.Record(common.Event{
		Type:   common.RecallAttached,
		Recall: &common.RecallData{Parts: parts},
	})
}

// recallConvo returns a short excerpt of recent dialogue for the judge.
//
// This is DATA, not messages. It is quoted inside a prompt the judge is asked
// to reason about, and is never replayed to the judge as a conversation it
// took part in.
func (e *Engine) recallConvo() []string {
	dialogue := e.Ctx.Dialogue
	start := len(dialogue) - recallConvoEntries
	if start < 0 {
		start = 0
	}
	out := make([]string, 0, recallConvoEntries)
	for _, entry := range dialogue[start:] {
		// Recall entries are skipped. Feeding previously recalled material
		// back in as conversational context would let one lucky retrieval
		// keep re-nominating itself forever, and the archive would slowly
		// converge on whatever it happened to surface first.
		if entry.Kind == common.KindRecall {
			continue
		}
		var sb strings.Builder
		for _, p := range entry.Parts {
			t, ok := p.(common.TextPart)
			if !ok {
				continue
			}
			sb.WriteString(t.Text)
		}
		text := strings.TrimSpace(sb.String())
		if text == "" {
			continue
		}
		if len(text) > recallConvoEntryBytes {
			text = text[:recallConvoEntryBytes]
		}
		out = append(out, text)
	}
	return out
}
