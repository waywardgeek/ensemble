package common

// Model features table. Capability is DATA about a model, not BEHAVIOUR of
// code. A table is diffable, testable, and updatable without touching logic.
//
// A missing row is a loud failure; a default row is a silent one.

// Media is a bitmask of INPUT media types a model accepts.
type Media uint8

const (
	MediaImage    Media = 1 << iota // 1
	MediaAudio                      // 2
	MediaVideo                      // 4
	MediaDocument                   // 8
)

// mediaNames maps individual bits to human-readable names.
var mediaNames = map[Media]string{
	MediaImage:    "image",
	MediaAudio:    "audio",
	MediaVideo:    "video",
	MediaDocument: "document",
}

// String returns the name of a single media bit.
func (m Media) String() string {
	if s, ok := mediaNames[m]; ok {
		return s
	}
	return "unknown"
}

// ModelFeatures is one row: everything the seam needs to know about a model.
type ModelFeatures struct {
	// Price is the dollar cost per million tokens for each of the four
	// disjoint Usage categories. The zero sheet means unpriced, not free:
	// see Pricing.
	Price Pricing

	Media Media

	// Stream is the set of delta kinds this model actually streams.
	//
	// Capability belongs here and not in the seam code, for the same reason
	// Media does: it is a fact about a model that changes on the vendor's
	// schedule, not a branch in our logic.
	Stream Stream

	// MaxThinkingTokens is the model's ceiling for reasoning budget.
	// Zero means the model does not support thinking. ThinkingFor uses
	// this to compute the actual budget per effort level.
	MaxThinkingTokens int

	// NoThinkingWithTools says the model rejects a reasoning budget in the same
	// request as function tools. It supports both, just not together on the
	// endpoint we speak, and the API answers with a 400 rather than degrading:
	//
	//   Function tools with reasoning_effort are not supported for
	//   gpt-5.6-sol in /v1/chat/completions. To use function tools, use
	//   /v1/responses or set reasoning_effort to 'none'.
	//
	// Which makes it fatal rather than cosmetic for an agent, because an agent
	// always has tools. Every request fails, so the model is unusable rather
	// than merely unthinking.
	//
	// It is a field here and not a branch in the renderer for the reason the
	// Stream comment gives: it is a fact about a model that changes on the
	// vendor's schedule. The proper fix is to speak /v1/responses, which
	// supports both at once; that is an endpoint migration, not a flag, so this
	// records the constraint until someone does it.
	NoThinkingWithTools bool

	// MaxOutputTokens is the model's ceiling for output tokens, including
	// any thinking budget. This is DATA about the model, not a default
	// for the request — it tells the framework what ceiling is safe to
	// request. Zero means "use whatever MaxTokens the caller set".
	MaxOutputTokens int

	// AdaptiveThinking indicates the model uses Anthropic's adaptive
	// thinking API (type:"adaptive" + output_config.effort) instead of
	// manual extended thinking (type:"enabled" + budget_tokens). This
	// is a wire-format distinction: claude-opus-5 and claude-sonnet-5
	// require adaptive; older models require manual.
	AdaptiveThinking bool

	// StubsToolResults enables Chapter 15's per-round-trip stubbing: every
	// tool result above the stub threshold becomes a stub in the request
	// after the one that carried it, unless the model's next message calls
	// keep_tool_results. It is a column because it is a judgement about the
	// model: a model that curates its own context well gains from it, and
	// one that does not (Sonnet 5, in Bill's use) loses results it needed.
	// Without it, only the ladder applies.
	StubsToolResults bool

	// ContextWindow is how many tokens the model will accept in one
	// request, prompt and all. Zero means nobody filled the column in.
	//
	// Zero is not a default, it is an admission, and the forcing rule
	// treats it as one: with no window there is no such thing as ninety
	// percent of it, so nothing is forced and the agent is left alone.
	// Guessing a window here would be worse than not knowing, because the
	// consequence of guessing low is taking every tool away from an agent
	// that was doing fine.
	ContextWindow int

	// InlineTools says the vendor can carry a tool declaration inside the
	// dialog, so a skill loaded mid-session adds tools without touching the
	// frozen prefix (Chapter 15 rule 2). On the Anthropic API this is a
	// role:"system" message of tool_addition blocks (betas
	// mid-conversation-tool-changes-2026-07-01 and inline-tools-2026-09-15);
	// Sonnet 5 does not accept those messages. Without it, the renderer
	// re-declares the tools array and pays the cache miss.
	InlineTools bool
}

// models is keyed by MODEL, not by vendor.
//
// This table rots. Model IDs are retired and renamed on vendor schedules that
// have nothing to do with publication dates. The table ships with a documented
// way to re-verify it, and no chapter's correctness depends on a particular
// model ID still existing. The table is an example of a shape, not a reference
// you should trust.
// ModelEntry is a model with its ID, for ordered iteration.
type ModelEntry struct {
	ID       string
	Features ModelFeatures
}

// ModelJSON is the JSON-friendly model descriptor sent to the GUI.
type ModelJSON struct {
	ID            string `json:"id"`
	DisplayName   string `json:"display_name"`
	Vendor        string `json:"vendor"`
	ContextWindow int    `json:"context_window"`
	MaxOutput     int    `json:"max_output"`
}

// ModelListJSON returns the model catalog for the GUI dropdown.
func ModelListJSON() []ModelJSON {
	entries := ModelList()
	out := make([]ModelJSON, len(entries))
	for i, e := range entries {
		out[i] = ModelJSON{
			ID:            e.ID,
			DisplayName:   displayName(e.ID),
			Vendor:        vendor(e.ID),
			ContextWindow: e.Features.ContextWindow,
			MaxOutput:     e.Features.MaxOutputTokens,
		}
	}
	return out
}

// ModelList returns the user-facing models in display order, grouped by vendor.
// Grader/test models are excluded. The GUI builds its dropdown from this.
func ModelList() []ModelEntry {
	return []ModelEntry{
		// Anthropic
		{ID: "claude-opus-4-6", Features: models()["claude-opus-4-6"]},
		{ID: "claude-opus-5", Features: models()["claude-opus-5"]},
		{ID: "claude-sonnet-5", Features: models()["claude-sonnet-5"]},
		// OpenAI
		{ID: "gpt-6-astra", Features: models()["gpt-6-astra"]},
		{ID: "gpt-5.6-sol", Features: models()["gpt-5.6-sol"]},
		// Google
		{ID: "gemini-3.8-flash", Features: models()["gemini-3.8-flash"]},
		{ID: "gemini-3.1-pro-preview", Features: models()["gemini-3.1-pro-preview"]},
	}
}

// displayName returns a human-readable label for a model ID.
func displayName(id string) string {
	names := map[string]string{
		"claude-opus-4-6":        "Claude Opus 4.6",
		"claude-opus-5":          "Claude Opus 5",
		"claude-sonnet-5":        "Claude Sonnet 5",
		"gpt-6-astra":            "GPT-6 Astra",
		"gpt-5.6-sol":            "GPT-5.6 Sol",
		"gemini-3.8-flash":       "Gemini 3.8 Flash",
		"gemini-3.1-pro-preview": "Gemini 3.1 Pro",
	}
	if n, ok := names[id]; ok {
		return n
	}
	return id
}

// vendor returns the vendor group label for a model ID.
func vendor(id string) string {
	if len(id) >= 6 && id[:6] == "claude" {
		return "Anthropic"
	}
	if len(id) >= 3 && id[:3] == "gpt" {
		return "OpenAI"
	}
	if len(id) >= 6 && id[:6] == "gemini" {
		return "Google"
	}
	return "Other"
}

// Pricing is dollars per million tokens, with one field per disjoint Usage
// category. The categories do not overlap, so a cost is a plain dot product
// of a tally against a sheet.
//
// The multipliers are NOT uniform across vendors, which is why these are
// stored explicitly rather than derived from Input by a constant factor.
// Anthropic charges 1.25x input to write a cache entry and 0.1x to read one;
// one Anthropic model reads at 0.05x; Gemini charges no write premium at all
// (CacheWrite == Input); and several models support no caching whatsoever.
// A computed multiplier would be wrong for most rows in the table.
//
// Zero means UNPRICED, not free. Course and fake models used by graders have
// no price sheet, and a renderer must show a dash for them rather than
// "$0.00" — an unpriced model must not be able to masquerade as a free one.
type Pricing struct {
	Input      float64
	CacheWrite float64
	CacheRead  float64
	Output     float64
}

// Priced reports whether a sheet has been filled in. Tested on the two
// categories every paid model charges for, so a row cannot look priced
// merely by declaring a cache rate.
func (p Pricing) Priced() bool {
	return p.Input > 0 || p.Output > 0
}

// models returns the feature table. A function rather than a package-level var
// so that the table is effectively immutable — no code can write to it.
func models() map[string]ModelFeatures {
	return map[string]ModelFeatures{
		// Anthropic — no audio, no video. Streams all three kinds.
		// All current Anthropic models have 1M context windows as of 2026.
		"claude-opus-4-6": {Price: Pricing{Input: 5, CacheWrite: 6.25, CacheRead: 0.5, Output: 25}, ContextWindow: 1000000, Media: MediaImage | MediaDocument, Stream: StreamAll, MaxThinkingTokens: 32768, MaxOutputTokens: 128000, AdaptiveThinking: true, StubsToolResults: true, InlineTools: true},
		"claude-opus-5":   {Price: Pricing{Input: 5, CacheWrite: 6.25, CacheRead: 0.5, Output: 25}, ContextWindow: 1000000, Media: MediaImage | MediaDocument, Stream: StreamAll, MaxThinkingTokens: 32768, MaxOutputTokens: 128000, AdaptiveThinking: true, StubsToolResults: true, InlineTools: true},
		"claude-sonnet-5": {Price: Pricing{Input: 3, CacheWrite: 3.75, CacheRead: 0.3, Output: 15}, ContextWindow: 1000000, Media: MediaImage | MediaDocument, Stream: StreamAll, MaxThinkingTokens: 32768, MaxOutputTokens: 128000, AdaptiveThinking: true},

		// OpenAI — images yes, audio and video NO (video APIs are generation).
		// Reasoning arrives as a summary rather than as incremental deltas,
		// so thinking is not streamed. These models support reasoning_effort
		// but do not return reasoning content on Chat Completions.
		"gpt-6-astra": {Price: Pricing{Input: 10, CacheWrite: 12.50, CacheRead: 1.0, Output: 50}, ContextWindow: 128000, Media: MediaImage | MediaDocument, Stream: StreamText | StreamToolArgs, MaxThinkingTokens: 32768, MaxOutputTokens: 16384},
		"gpt-5.6-sol": {Price: Pricing{Input: 4, CacheWrite: 5.0, CacheRead: 0.4, Output: 20}, ContextWindow: 128000, Media: MediaImage | MediaDocument, Stream: StreamText | StreamToolArgs, MaxThinkingTokens: 32768, MaxOutputTokens: 16384, NoThinkingWithTools: true},

		// Gemini — images, audio, video, documents. Text and thinking stream;
		// FUNCTION-CALL ARGUMENTS DO NOT. They arrive complete, in one frame.
		// This row is why Stream is a bitmask instead of a bool: the honest
		// description of this model needs two of three bits set, and a bool
		// would have forced us either to drop text streaming or to invent
		// argument chunks that the vendor never sent.
		"gemini-3.8-flash":       {Price: Pricing{Input: 0.75, CacheWrite: 0.75, CacheRead: 0.075, Output: 3.75}, ContextWindow: 1000000, Media: MediaImage | MediaAudio | MediaVideo | MediaDocument, Stream: StreamText | StreamThinking, MaxThinkingTokens: 32768, MaxOutputTokens: 16384},
		"gemini-3.1-pro-preview": {Price: Pricing{Input: 2, CacheWrite: 2.0, CacheRead: 0.20, Output: 12}, ContextWindow: 1000000, Media: MediaImage | MediaAudio | MediaVideo | MediaDocument, Stream: StreamText | StreamThinking, MaxThinkingTokens: 32768, MaxOutputTokens: 16384},

		// Fake model used by the grader — images only, to test loud refusal.
		// Streams everything, because the streaming checks need all three
		// kinds to appear. Thinking enabled for testing.
		"fake-model": {ContextWindow: 200000, Media: MediaImage, Stream: StreamAll, MaxThinkingTokens: 32768, MaxOutputTokens: 16384},

		// Course/test models used by graders in various chapters.
		"claude-fake-course-1":   {ContextWindow: 200000, Media: MediaImage | MediaDocument, Stream: StreamAll, MaxThinkingTokens: 32768, MaxOutputTokens: 16384, AdaptiveThinking: true},
		"claude-sonnet-5-course": {ContextWindow: 200000, Media: MediaImage | MediaDocument, Stream: StreamAll, MaxThinkingTokens: 32768, MaxOutputTokens: 16384, AdaptiveThinking: true},
		// Chapter 15's grader model: the claude-opus-5 row, course-named.
		// A new row rather than a changed one, so every earlier chapter's
		// grader, which runs fake-model or claude-fake-course-1, sees
		// exactly the wire it saw before.
		"claude-opus-5-course": {ContextWindow: 200000, Media: MediaImage | MediaDocument, Stream: StreamAll, MaxThinkingTokens: 32768, MaxOutputTokens: 16384, AdaptiveThinking: true, StubsToolResults: true, InlineTools: true},
		"gpt-5-course":         {ContextWindow: 128000, Media: MediaImage | MediaDocument, Stream: StreamText | StreamToolArgs, MaxThinkingTokens: 32768, MaxOutputTokens: 16384},
		// Chapter 16's grader model: the opus row with a deliberately tiny
		// window, so a grader can reach ninety percent of it in a handful of
		// turns instead of a hundred thousand. A NEW row rather than a
		// changed one, so no earlier chapter's grader sees a different wire.
		"claude-ch16-course":      {ContextWindow: 4096, Media: MediaImage | MediaDocument, Stream: StreamAll, MaxThinkingTokens: 32768, MaxOutputTokens: 16384, AdaptiveThinking: true, StubsToolResults: true},
		"gemini-3.5-flash-course": {ContextWindow: 1000000, Media: MediaImage | MediaAudio | MediaVideo | MediaDocument, Stream: StreamText | StreamThinking, MaxThinkingTokens: 32768, MaxOutputTokens: 16384},
	}
}

// LookupModel returns the feature row for a model, or false. There is
// deliberately no default row. Unknown model → loud refusal.
func LookupModel(model string) (ModelFeatures, bool) {
	m := models()
	f, ok := m[model]
	return f, ok
}
