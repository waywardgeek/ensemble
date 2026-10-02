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
// Caching says HOW a vendor caches a repeated prompt prefix.
//
// EVERY model we speak to supports prompt caching. What differs is who is
// responsible for pointing at the reusable part, and that difference decides
// what our renderer must emit:
//
//   - CacheExplicit: the prefix is cached only where we mark it. Anthropic
//     works this way; a request with no cache_control marker is billed in
//     full every turn, forever, in silence.
//
//   - CacheImplicit: the vendor finds the repeated prefix itself and bills
//     it at the cached rate. OpenAI and Gemini work this way. Sending no
//     directives is not an omission, it is the entire interface. There is
//     nothing to mark and nothing to enable.
//
// Implicit is the better deal and, where a vendor offers both, the
// recommended one: it costs no marker budget, it cannot be aimed at the
// wrong byte, it needs no lifetime management, and it degrades to "no
// discount" rather than to "wrong answer" when the prefix moves. Explicit
// caching exists to buy control we mostly do not want.
//
// Implicit is not free of warm-up, and the shape is worth knowing because it
// looks like a fault. Measured on gemini-3.8-flash against the live API:
//
//   - The fixed head, meaning the system prompt and the tool declarations, is
//     served from the SECOND request. One sighting is enough for it.
//
//   - A conversation message must be stable across TWO successive requests
//     before it is cached, so a growing history is served a turn later still.
//     An agent therefore sees two cold turns at startup rather than one.
//
//   - The cached span commits in blocks of roughly four thousand tokens, so a
//     turn that appends only a little text shows no increase at all. Measured
//     counts land on multiples of the block size rather than on the true
//     prefix length, which also caps the achievable hit rate well below 100%.
//
// The practical consequence for anything we build: put large stable content in
// the system prompt rather than in the first user message, because the head
// caches a full turn earlier than the history does.
//
// The consequence for a diagnostic is the whole reason this field exists:
// "we sent no cache directives" is a DEFECT on an explicit model and CORRECT
// BEHAVIOUR on an implicit one. Without this column a lens cannot tell the
// two apart, and will report a healthy Gemini request as broken.
type Caching int

const (
	// CacheExplicit means the prefix is cached only where we mark it.
	CacheExplicit Caching = iota + 1
	// CacheImplicit means the vendor detects the repeated prefix itself.
	CacheImplicit
)

type ModelFeatures struct {
	// Vendor is the API this model is served by. It is a fact about the
	// model, not a deployment choice, so it belongs in the table rather than
	// in configuration. The zero value is deliberately invalid: a row that
	// forgets its vendor is detectable instead of silently Anthropic.
	//
	// Test models driven by a per-vendor env prefix genuinely have no single
	// vendor and correctly leave this unset.
	Vendor Vendor

	// Surface is the endpoint dialect this model is spoken to. Like Vendor
	// it is a fact about the model rather than a deployment choice: a vendor
	// can serve two surfaces at once, and a single model is not generally
	// reachable on both with the same request body.
	//
	// The zero value means "the vendor's default surface", which keeps every
	// existing row honest: a model that predates a second surface genuinely
	// has no opinion, and inventing one for it would be a guess recorded as
	// a fact. Only rows that have actually been exercised on the newer
	// endpoint name it.
	Surface Surface

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

	// NoEphemera says the model cannot be sent ephemeral content: material
	// that appears in one request and is gone from the next. Anthropic's
	// Opus 5.5 tolerates no change to a prefix it has already seen, and
	// ephemera are by definition a change. The flag is negative so that the
	// zero value, which every other row takes, means the model is fine.
	//
	// This is a fact about the model, not a branch in the caller. The
	// consequence is wider than it looks: anything delivered as ephemera is
	// silently dropped for such a model, so a feature built on ephemera must
	// be skipped at its source rather than generated and thrown away.
	NoEphemera bool

	// Caching is how this model caches a repeated prompt prefix. See the
	// Caching type. Every model in the table below supports caching; none
	// of them is CacheNone, because there is no such value.
	Caching Caching

	// MinCacheTokens is the smallest prompt, in tokens, the vendor will
	// cache at all. A request below this floor is simply not eligible, and
	// reports a zero hit that is byte-for-byte indistinguishable from a
	// broken prefix unless you know the floor — which is exactly how we
	// once concluded that a vendor had no caching at all, from three short
	// requests that were never in the running.
	//
	// Provenance: Gemini's 4096 is documented and was confirmed by
	// measurement (21,660-token prompt, 16,359 served from cache on the
	// second and every later send, with no directives of any kind).
	// Anthropic's and OpenAI's 1024 are the vendors' documented minimums;
	// the OpenAI figure is consistent with a measured first-turn hit of
	// 1,738 tokens.
	MinCacheTokens int

	// MaxCacheWrites is how many explicit cache breakpoints one request may
	// WRITE. Zero means the vendor does not let you place them and the
	// question does not arise.
	//
	// This is a separate fact from Caching, which says how the cache is
	// ADDRESSED. A vendor can accept explicit breakpoints and still cap how
	// many it will bill you for, and exceeding the cap is not an error: the
	// surplus markers are ignored, so the request succeeds, the prefix you
	// meant to bank is not banked, and nothing anywhere says so.
	//
	// Provenance: OpenAI's 4 is documented for the Responses API and was
	// confirmed today against the live endpoint, which accepted
	// prompt_cache_options {"mode":"explicit"} and echoed it back normalised
	// as {"mode":"explicit","ttl":"30m"}. Note the documented asymmetry:
	// implicit mode writes one implicit breakpoint plus up to THREE explicit
	// ones, explicit mode writes up to FOUR and creates no implicit one. The
	// four-breakpoint architecture therefore requires explicit mode.
	MaxCacheWrites int

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
		{ID: "claude-opus-5-5", Features: models()["claude-opus-5-5"]},
		{ID: "claude-sonnet-5", Features: models()["claude-sonnet-5"]},
		// OpenAI
		{ID: "gpt-6.1-sol", Features: models()["gpt-6.1-sol"]},
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
		"claude-opus-5-5":        "Claude Opus 5.5",
		"claude-sonnet-5":        "Claude Sonnet 5",
		"gpt-6-astra":            "GPT-6 Astra",
		"gpt-6.1-sol":            "GPT-6.1 Sol",
		"gpt-5.6-sol":            "GPT-5.6 Sol",
		"gemini-3.8-flash":       "Gemini 3.8 Flash",
		"gemini-3.1-pro-preview": "Gemini 3.1 Pro",
	}
	if n, ok := names[id]; ok {
		return n
	}
	return id
}

// VendorFor reports which API serves a model. The mapping lives in the model
// table, so there is exactly one of it: a model's vendor is a fact about the
// model, not a deployment choice.
//
// ok is false for an unknown model, and also for a test model that is driven
// against several vendors by env prefix and so has no single answer. A caller
// that needs to dial an API must treat !ok as a refusal rather than guessing,
// because every available guess is wrong for some model.
func VendorFor(id string) (Vendor, bool) {
	f, found := LookupModel(id)
	if !found || f.Vendor == Vendor(0) {
		return Vendor(0), false
	}
	return f.Vendor, true
}

// vendor returns the vendor group label for a model ID. It derives from the
// table so the name shown in the GUI cannot drift from the API dialed.
func vendor(id string) string {
	v, ok := VendorFor(id)
	if !ok {
		return "Other"
	}
	switch v {
	case VendorAnthropic:
		return "Anthropic"
	case VendorOpenAI:
		return "OpenAI"
	case VendorGemini:
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
		"claude-opus-4-6": {Vendor: VendorAnthropic, Caching: CacheExplicit, MinCacheTokens: 1024, Price: Pricing{Input: 5, CacheWrite: 6.25, CacheRead: 0.5, Output: 25}, ContextWindow: 1000000, Media: MediaImage | MediaDocument, Stream: StreamAll, MaxThinkingTokens: 32768, MaxOutputTokens: 128000, AdaptiveThinking: true, StubsToolResults: true, InlineTools: true},
		"claude-opus-5":   {Vendor: VendorAnthropic, Caching: CacheExplicit, MinCacheTokens: 1024, Price: Pricing{Input: 5, CacheWrite: 6.25, CacheRead: 0.5, Output: 25}, ContextWindow: 1000000, Media: MediaImage | MediaDocument, Stream: StreamAll, MaxThinkingTokens: 32768, MaxOutputTokens: 128000, AdaptiveThinking: true, StubsToolResults: true, InlineTools: true},
		"claude-opus-5-5": {Vendor: VendorAnthropic, Caching: CacheExplicit, MinCacheTokens: 1024, Price: Pricing{Input: 4, CacheWrite: 5, CacheRead: 0.2, Output: 20}, ContextWindow: 1000000, Media: MediaImage | MediaDocument, Stream: StreamAll, MaxThinkingTokens: 32768, MaxOutputTokens: 128000, AdaptiveThinking: true, StubsToolResults: true, InlineTools: true, NoEphemera: true},
		"claude-sonnet-5": {Vendor: VendorAnthropic, Caching: CacheExplicit, MinCacheTokens: 1024, Price: Pricing{Input: 3, CacheWrite: 3.75, CacheRead: 0.3, Output: 15}, ContextWindow: 1000000, Media: MediaImage | MediaDocument, Stream: StreamAll, MaxThinkingTokens: 32768, MaxOutputTokens: 128000, AdaptiveThinking: true},

		// OpenAI — images yes, audio and video NO (video APIs are generation).
		// Reasoning arrives as a summary rather than as incremental deltas,
		// so thinking is not streamed. These models support reasoning_effort
		// but do not return reasoning content on Chat Completions.
		"gpt-6-astra": {Vendor: VendorOpenAI, Caching: CacheImplicit, MinCacheTokens: 1024, Price: Pricing{Input: 10, CacheWrite: 12.50, CacheRead: 1.0, Output: 50}, ContextWindow: 128000, Media: MediaImage | MediaDocument, Stream: StreamText | StreamToolArgs, MaxThinkingTokens: 32768, MaxOutputTokens: 16384},
		"gpt-5.6-sol": {Vendor: VendorOpenAI, Caching: CacheExplicit, MinCacheTokens: 1024, Price: Pricing{Input: 4, CacheWrite: 5.0, CacheRead: 0.4, Output: 20}, ContextWindow: 1050000, Media: MediaImage | MediaDocument, Stream: StreamText | StreamToolArgs, MaxThinkingTokens: 32768, MaxOutputTokens: 128000, NoThinkingWithTools: true},

		// Chapter 19's model, and the agent's new default.
		//
		// Three things here were MEASURED against the live endpoint today
		// rather than assumed, because this row is the one the agent now
		// runs on by default and a wrong entry here is wrong everywhere.
		//
		// Measured: the model exists (it is in GET /v1/models). Measured:
		// it streams reasoning summaries, which is why this is the first
		// row in the table to claim StreamReasoningSummary. With
		// reasoning.summary "detailed" and effort "high" one response
		// produced 314 summary deltas across 3 parts; with "concise", 4.
		// Measured: it accepts prompt_cache_options {"mode":"explicit"}
		// and a prompt_cache_breakpoint on a developer content block.
		//
		// NOT measured by me, but taken from OpenAI's own model page
		// rather than guessed: pricing, context window and output cap.
		// The page states cached input is 5% of the uncached rate and
		// cache writes are 1.25x it, which is where CacheRead 0.1 and
		// CacheWrite 2.5 against Input 2 come from. Those two ratios are
		// the ones chapter 18 reasoned about, now first-party confirmed.
		//
		// Media is image only. The model page lists input modalities as
		// text and image, and names audio and video as unsupported. The
		// 5.6-sol row above claims documents; this one does not, because
		// the page does not.
		//
		// Note what is ABSENT: NoThinkingWithTools. The 5.6-sol row
		// carries it because /v1/chat/completions refuses function tools
		// and reasoning_effort together, and the comment there says the
		// proper fix is to speak /v1/responses. This chapter does that.
		// The model page agrees: "Use the Responses API for tool calling.
		// Chat Completions is supported without tool calling."
		"gpt-6.1-sol": {Vendor: VendorOpenAI, Surface: SurfaceResponses, Caching: CacheExplicit, MinCacheTokens: 1024, MaxCacheWrites: 4, Price: Pricing{Input: 2, CacheWrite: 2.5, CacheRead: 0.1, Output: 10}, ContextWindow: 1050000, Media: MediaImage, Stream: StreamText | StreamToolArgs | StreamReasoningSummary, MaxThinkingTokens: 32768, MaxOutputTokens: 128000},

		// Chapter 19's grader model. A NEW row rather than a changed one,
		// following the convention established by the course rows above:
		// every earlier chapter's grader must keep seeing exactly the wire
		// it saw before, and the surest way to guarantee that is never to
		// edit a row another chapter depends on.
		"gpt-ch19-course": {Vendor: VendorOpenAI, Surface: SurfaceResponses, Caching: CacheExplicit, MinCacheTokens: 1024, MaxCacheWrites: 4, ContextWindow: 128000, Media: MediaImage | MediaDocument, Stream: StreamText | StreamToolArgs | StreamReasoningSummary, MaxThinkingTokens: 32768, MaxOutputTokens: 16384},

		// Gemini — images, audio, video, documents. Text and thinking stream;
		// FUNCTION-CALL ARGUMENTS DO NOT. They arrive complete, in one frame.
		// This row is why Stream is a bitmask instead of a bool: the honest
		// description of this model needs two of three bits set, and a bool
		// would have forced us either to drop text streaming or to invent
		// argument chunks that the vendor never sent.
		"gemini-3.8-flash":       {Vendor: VendorGemini, Caching: CacheImplicit, MinCacheTokens: 4096, Price: Pricing{Input: 0.75, CacheWrite: 0.75, CacheRead: 0.075, Output: 3.75}, ContextWindow: 1000000, Media: MediaImage | MediaAudio | MediaVideo | MediaDocument, Stream: StreamText | StreamThinking, MaxThinkingTokens: 32768, MaxOutputTokens: 16384},
		"gemini-3.1-pro-preview": {Vendor: VendorGemini, Caching: CacheImplicit, MinCacheTokens: 4096, Price: Pricing{Input: 2, CacheWrite: 2.0, CacheRead: 0.20, Output: 12}, ContextWindow: 1000000, Media: MediaImage | MediaAudio | MediaVideo | MediaDocument, Stream: StreamText | StreamThinking, MaxThinkingTokens: 32768, MaxOutputTokens: 16384},

		// Fake model used by the grader — images only, to test loud refusal.
		// Streams everything, because the streaming checks need all three
		// kinds to appear. Thinking enabled for testing.
		"fake-model": {Caching: CacheExplicit, MinCacheTokens: 1024, ContextWindow: 200000, Media: MediaImage, Stream: StreamAll, MaxThinkingTokens: 32768, MaxOutputTokens: 16384},

		// Course/test models used by graders in various chapters.
		"claude-fake-course-1":   {Vendor: VendorAnthropic, Caching: CacheExplicit, MinCacheTokens: 1024, ContextWindow: 200000, Media: MediaImage | MediaDocument, Stream: StreamAll, MaxThinkingTokens: 32768, MaxOutputTokens: 16384, AdaptiveThinking: true},
		"claude-sonnet-5-course": {Vendor: VendorAnthropic, Caching: CacheExplicit, MinCacheTokens: 1024, ContextWindow: 200000, Media: MediaImage | MediaDocument, Stream: StreamAll, MaxThinkingTokens: 32768, MaxOutputTokens: 16384, AdaptiveThinking: true},
		// Chapter 15's grader model: the claude-opus-5 row, course-named.
		// A new row rather than a changed one, so every earlier chapter's
		// grader, which runs fake-model or claude-fake-course-1, sees
		// exactly the wire it saw before.
		"claude-opus-5-course": {Vendor: VendorAnthropic, Caching: CacheExplicit, MinCacheTokens: 1024, ContextWindow: 200000, Media: MediaImage | MediaDocument, Stream: StreamAll, MaxThinkingTokens: 32768, MaxOutputTokens: 16384, AdaptiveThinking: true, StubsToolResults: true, InlineTools: true},
		"gpt-5-course":         {Vendor: VendorOpenAI, Caching: CacheImplicit, MinCacheTokens: 1024, ContextWindow: 128000, Media: MediaImage | MediaDocument, Stream: StreamText | StreamToolArgs, MaxThinkingTokens: 32768, MaxOutputTokens: 16384},
		// Chapter 16's grader model: the opus row with a deliberately tiny
		// window, so a grader can reach ninety percent of it in a handful of
		// turns instead of a hundred thousand. A NEW row rather than a
		// changed one, so no earlier chapter's grader sees a different wire.
		"claude-ch16-course":      {Vendor: VendorAnthropic, Caching: CacheExplicit, MinCacheTokens: 1024, ContextWindow: 4096, Media: MediaImage | MediaDocument, Stream: StreamAll, MaxThinkingTokens: 32768, MaxOutputTokens: 16384, AdaptiveThinking: true, StubsToolResults: true},
		"gemini-3.5-flash-course": {Vendor: VendorGemini, Caching: CacheImplicit, MinCacheTokens: 4096, ContextWindow: 1000000, Media: MediaImage | MediaAudio | MediaVideo | MediaDocument, Stream: StreamText | StreamThinking, MaxThinkingTokens: 32768, MaxOutputTokens: 16384},
	}
}

// LookupModel returns the feature row for a model, or false. There is
// deliberately no default row. Unknown model → loud refusal.
func LookupModel(model string) (ModelFeatures, bool) {
	m := models()
	f, ok := m[model]
	return f, ok
}
