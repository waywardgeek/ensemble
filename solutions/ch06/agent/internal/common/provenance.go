package common

// Provenance: who produced a piece of content, and with which model.
//
// This is the subtle type in the chapter, and it is where a seam that looks
// finished turns out not to be. The naive version is `Vendor string`. That is
// wrong in a way you will not discover until a user switches models mid
// conversation, because thinking signatures — the encrypted reasoning
// material — are bound to the MODEL, not the vendor.

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Vendor uint8
type Surface uint8

// iota + 1 on purpose. The zero value must be INVALID, so a Provenance that
// nobody populated is detectable instead of silently meaning "Anthropic".
// Since provenance must be captured at write time and can never be
// reconstructed, "nobody populated it" is precisely the bug you need loud.
const (
	VendorAnthropic Vendor = iota + 1
	VendorGemini
	VendorOpenAI
)

const (
	SurfaceMessages        Surface = iota + 1 // Anthropic
	SurfaceChatCompletions                    // OpenAI — what Exhibits A–C speak
	SurfaceGenerateContent                    // Gemini — what Exhibits A–C speak
	SurfaceInteractions                       // Gemini's replacement surface
	SurfaceResponses                          // OpenAI's newer surface
)

var vendorNames = map[Vendor]string{
	VendorAnthropic: "anthropic",
	VendorGemini:    "gemini",
	VendorOpenAI:    "openai",
}

var surfaceNames = map[Surface]string{
	SurfaceMessages:        "messages",
	SurfaceInteractions:    "interactions",
	SurfaceResponses:       "responses",
	SurfaceChatCompletions: "chat_completions",
	SurfaceGenerateContent: "generate_content",
}

func (v Vendor) String() string  { return vendorNames[v] }
func (s Surface) String() string { return surfaceNames[s] }

// NormalizeName lowercases and strips punctuation so that "ToolCalled",
// "tool_called" and "TOOL-CALLED" are the same name. Spelling is not a lesson.
func NormalizeName(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// Marshal enums as readable strings. The log is JSON-lines so that ordinary
// tools can grep it; `"vendor":2` destroys that for no gain.
func (v Vendor) MarshalJSON() ([]byte, error) {
	s, ok := vendorNames[v]
	if !ok {
		return nil, fmt.Errorf("refusing to marshal invalid vendor %d (zero value means it was never captured)", uint8(v))
	}
	return json.Marshal(s)
}

// UnmarshalJSON refuses an unrecognized vendor loudly. Not a default, not a
// skip — the same discipline as an unknown event type, for the same reason.
func (v *Vendor) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	for k, name := range vendorNames {
		if NormalizeName(name) == NormalizeName(s) {
			*v = k
			return nil
		}
	}
	return fmt.Errorf("unknown vendor %q: refusing to load this log", s)
}

func (s Surface) MarshalJSON() ([]byte, error) {
	name, ok := surfaceNames[s]
	if !ok {
		return nil, fmt.Errorf("refusing to marshal invalid surface %d", uint8(s))
	}
	return json.Marshal(name)
}

func (s *Surface) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	for k, name := range surfaceNames {
		if NormalizeName(name) == NormalizeName(str) {
			*s = k
			return nil
		}
	}
	return fmt.Errorf("unknown surface %q: refusing to load this log", str)
}

// Provenance is recorded at WRITE time, by the client that produced the
// content, and is never inferred afterwards. A renderer may read it; nothing
// may reconstruct it. Inference is impossible in principle: by the time you
// are rendering, the model that produced a signature three turns ago is not
// derivable from anything else in the context.
type Provenance struct {
	Vendor  Vendor  `json:"vendor"`
	Model   string  `json:"model"` // OPEN set. Never switch on it.
	Surface Surface `json:"surface"`
}

// Valid reports whether this provenance was actually captured.
func (p Provenance) Valid() bool {
	return p.Vendor != 0 && p.Surface != 0 && p.Model != ""
}

// SameModel decides whether opaque replay material may be handed back.
// Vendor is not a fine enough grain; signature validity is scoped to
// (vendor, model, surface).
func (p Provenance) SameModel(o Provenance) bool {
	return p.Vendor == o.Vendor && p.Model == o.Model && p.Surface == o.Surface
}
