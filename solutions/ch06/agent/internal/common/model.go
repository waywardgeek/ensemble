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
	Media Media

	// Stream is the set of delta kinds this model actually streams.
	//
	// Capability belongs here and not in the seam code, for the same reason
	// Media does: it is a fact about a model that changes on the vendor's
	// schedule, not a branch in our logic.
	Stream Stream
}

// models is keyed by MODEL, not by vendor.
//
// This table rots. Model IDs are retired and renamed on vendor schedules that
// have nothing to do with publication dates. The table ships with a documented
// way to re-verify it, and no chapter's correctness depends on a particular
// model ID still existing. The table is an example of a shape, not a reference
// you should trust.
// models returns the feature table. A function rather than a package-level var
// so that the table is effectively immutable — no code can write to it.
func models() map[string]ModelFeatures {
	return map[string]ModelFeatures{
		// Anthropic — no audio, no video. Streams all three kinds.
		"claude-opus-5":   {Media: MediaImage | MediaDocument, Stream: StreamAll},
		"claude-sonnet-5": {Media: MediaImage | MediaDocument, Stream: StreamAll},

		// OpenAI — images yes, audio and video NO (video APIs are generation).
		// Reasoning arrives as a summary rather than as incremental deltas,
		// so thinking is not streamed.
		"gpt-6-astra": {Media: MediaImage | MediaDocument, Stream: StreamText | StreamToolArgs},
		"gpt-5.6-sol": {Media: MediaImage | MediaDocument, Stream: StreamText | StreamToolArgs},

		// Gemini — images, audio, video, documents. Text and thinking stream;
		// FUNCTION-CALL ARGUMENTS DO NOT. They arrive complete, in one frame.
		// This row is why Stream is a bitmask instead of a bool: the honest
		// description of this model needs two of three bits set, and a bool
		// would have forced us either to drop text streaming or to invent
		// argument chunks that the vendor never sent.
		"gemini-3.8-flash":       {Media: MediaImage | MediaAudio | MediaVideo | MediaDocument, Stream: StreamText | StreamThinking},
		"gemini-3.1-pro-preview": {Media: MediaImage | MediaAudio | MediaVideo | MediaDocument, Stream: StreamText | StreamThinking},

		// Fake model used by the grader — images only, to test loud refusal.
		// Streams everything, because the streaming checks need all three
		// kinds to appear.
		"fake-model": {Media: MediaImage, Stream: StreamAll},

		// Course/test models used by graders in various chapters.
		"claude-fake-course-1":    {Media: MediaImage | MediaDocument, Stream: StreamAll},
		"claude-sonnet-5-course":  {Media: MediaImage | MediaDocument, Stream: StreamAll},
		"gpt-5-course":            {Media: MediaImage | MediaDocument, Stream: StreamText | StreamToolArgs},
		"gemini-3.5-flash-course": {Media: MediaImage | MediaAudio | MediaVideo | MediaDocument, Stream: StreamText | StreamThinking},
	}
}

// LookupModel returns the feature row for a model, or false. There is
// deliberately no default row. Unknown model → loud refusal.
func LookupModel(model string) (ModelFeatures, bool) {
	m := models()
	f, ok := m[model]
	return f, ok
}
