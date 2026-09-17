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
	// Anthropic — no audio, no video.
	"claude-sonnet-5":   {Media: MediaImage | MediaDocument},
	"claude-sonnet-4":   {Media: MediaImage | MediaDocument},
	"claude-haiku-3.5":  {Media: MediaImage | MediaDocument},

	// OpenAI — images yes, audio and video NO (video APIs are generation).
	"gpt-5":             {Media: MediaImage | MediaDocument},
	"gpt-4.1":           {Media: MediaImage | MediaDocument},
	"gpt-4o":            {Media: MediaImage | MediaDocument},

	// Gemini — images, audio, video, documents.
	"gemini-3.8-flash":  {Media: MediaImage | MediaAudio | MediaVideo | MediaDocument},
	"gemini-2.5-flash":  {Media: MediaImage | MediaAudio | MediaVideo | MediaDocument},
	"gemini-2.5-pro":    {Media: MediaImage | MediaAudio | MediaVideo | MediaDocument},

	// Fake model used by the grader — images only, to test loud refusal.
	"fake-model":        {Media: MediaImage},
	}
}

// LookupModel returns the feature row for a model, or false. There is
// deliberately no default row. Unknown model → loud refusal.
func LookupModel(model string) (ModelFeatures, bool) {
	m := models()
	f, ok := m[model]
	return f, ok
}
