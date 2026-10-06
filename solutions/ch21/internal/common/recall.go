package common

// Auto-recall, as far as the hub is concerned.
//
// Recall needs a model, because a small model judges which candidates are
// genuinely relevant. The model layer needs recall, because something has to
// attach the memories before a request goes out. Left alone that is a cycle,
// and a cycle between two spokes is exactly what this architecture forbids.
//
// So both directions invert through here. `internal/recall` implements
// Recaller and consumes SnippetJudge; `internal/llm` implements SnippetJudge
// and holds a nilable Recaller. Neither package names the other, and the only
// place both concrete types appear is the wiring at the top.
//
// The alternative — handing recall a callback to break the cycle — would
// work and would be worse. A callback added solely to dodge an import is a
// cycle you have hidden rather than removed; an interface on the hub is the
// same inversion written down where a reader can find it.

// Chunk is one searchable piece of a memory file or document.
//
// It is the unit recall retrieves and the unit the judge votes on, so it
// carries enough to attribute a quote — which file it came from and which
// section of that file — and nothing about how it was scored. Ranking
// belongs to the engine that does the ranking.
type Chunk struct {
	Filename string `json:"filename"`
	Header   string `json:"header"`
	Content  string `json:"content"`
}

// Recaller produces the memories that belong with a user message.
//
// `convo` is the tail of the conversation, most recent last, as plain text.
// It is passed in rather than read from a Context because recall must not
// reach into the agent's state: it is given a query and some context and it
// answers with bytes.
//
// The returned PartList is the formatted snippets, ready to attach. An empty
// return means "nothing worth saying", which is the common case and must be
// cheap.
//
// A nil Recaller means recall is switched off. The agent must still work —
// it simply does not remember passively.
type Recaller interface {
	Recall(query string, convo []string) PartList
}

// SnippetJudge filters recall candidates by relevance rather than by keyword.
//
// It is deliberately the smallest interface that can express the job: text in,
// text out. That is not minimalism for its own sake. An implementation this
// narrow cannot hold history, cannot own a context, and has nowhere to put an
// ephemeral provider, so a judge that recalls while judging is not a failure
// mode to be guarded against — it is a sentence that cannot be written.
//
// A nil SnippetJudge means no judge: fall back to top-N by BM25 score.
type SnippetJudge interface {
	Pick(prompt string) (string, error)
}
