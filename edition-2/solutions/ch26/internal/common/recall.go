package common

import "context"

// RecallSettings belongs to the Agent; zero numeric values select chapter defaults.
type RecallSettings struct {
	// Disabled bypasses retrieval without deleting any corpus or recorded entries.
	Disabled bool `json:"disabled"`
	// DisableJudge selects mechanical BM25 retrieval without a model call.
	DisableJudge bool `json:"disable_judge"`
	// JudgeModel optionally selects another model on the Agent's provider.
	JudgeModel string `json:"judge_model"`
	// MinScore is the plain-recall score floor; judged candidates use 0.5.
	MinScore float64 `json:"min_score"`
	// MaxSnippets limits selected snippets per human turn.
	MaxSnippets int `json:"max_snippets"`
	// MinQueryLength skips messages shorter than this many characters.
	MinQueryLength int `json:"min_query_length"`
	// Candidates bounds the judge's numbered candidate list.
	Candidates int `json:"candidates"`
	// ContextMsgs bounds recent human/assistant messages in the judge prompt.
	ContextMsgs int `json:"context_msgs"`
	// MaxRecallBytes caps the complete formatted attachment, attribution included.
	MaxRecallBytes int `json:"max_recall_bytes"`
}

// SnippetJudge is a stateless text exchange, with no Agent loop or stored history.
// Recall alone owns the prompt and reply format.
type SnippetJudge interface {
	// Judge returns provider text without recording a conversation event.
	Judge(context.Context, string) (string, error)
}

// Recaller owns per-source indexes through its required Agent parent.
type Recaller interface {
	// Agent exposes the actual owner of corpus locations and model services.
	Agent() Agent
	// Recall searches against a snapshot; only the actor may attach the returned bytes.
	Recall(context.Context, string, *Context) (RecallData, error)
}

// RecallData records chosen bytes so replay needs neither disk nor a judge.
type RecallData struct {
	// Parts carry attributed, blockquoted text, never the user's own words.
	Parts []Part `json:"parts"`
	// Hashes identify complete source snippets actually represented, including capped ones.
	// They live only as long as the attachment, so a checkpoint permits recall again.
	Hashes []string `json:"hashes,omitempty"`
}
