package common

import (
	"encoding/json"
	"fmt"
)

// Zero cannot silently stand for a provider: missing provenance loses replay
// identity. Models remain strings because their set changes without a rebuild.
type Vendor int

const (
	// Anthropic selects the Messages provider translator.
	Anthropic Vendor = iota + 1
	// OpenAI selects the Chat Completions provider translator.
	OpenAI
	// Gemini selects the GenerateContent provider translator.
	Gemini
)

// Surface identifies the vendor API shape used to capture opaque content.
type Surface int

const (
	// Messages identifies Anthropic opaque replay material.
	Messages Surface = iota + 1
	// ChatCompletions identifies OpenAI chat replay material.
	ChatCompletions
	// GenerateContent identifies Gemini call signatures and thought material.
	GenerateContent
)

// String returns the canonical provider name used by configuration and JSON.
func (v Vendor) String() string {
	switch v {
	case Anthropic:
		return "anthropic"
	case OpenAI:
		return "openai"
	case Gemini:
		return "gemini"
	}
	return "unknown"
}

// String returns the canonical API-surface name used by provenance JSON.
func (s Surface) String() string {
	switch s {
	case Messages:
		return "messages"
	case ChatCompletions:
		return "chat_completions"
	case GenerateContent:
		return "generate_content"
	}
	return "unknown"
}

// JSON dispatch belongs with the enum, while vendor behavior stays in Engine.
func (v Vendor) MarshalJSON() ([]byte, error) { return json.Marshal(v.String()) }

// MarshalJSON writes the canonical surface identity for replay.
func (s Surface) MarshalJSON() ([]byte, error) { return json.Marshal(s.String()) }

// Refuse unknown names while decoding, before replay can produce plausible but
// incomplete context. A future vendor needs a renderer, not a silent fallback.
func (v *Vendor) UnmarshalJSON(b []byte) error {
	var name string
	if err := json.Unmarshal(b, &name); err != nil {
		return err
	}
	for _, candidate := range []Vendor{Anthropic, OpenAI, Gemini} {
		if name == candidate.String() {
			*v = candidate
			return nil
		}
	}
	return fmt.Errorf("unknown vendor %q", name)
}

// Canonical output uses one spelling. The one documented input alias below
// repairs inconsistent supplied fixtures without accepting arbitrary surfaces.
func (s *Surface) UnmarshalJSON(b []byte) error {
	var name string
	if err := json.Unmarshal(b, &name); err != nil {
		return err
	}
	// Two published fixture spellings denote the same surface, not two APIs.
	if name == "generatecontent" {
		name = "generate_content"
	}
	for _, candidate := range []Surface{Messages, ChatCompletions, GenerateContent} {
		if name == candidate.String() {
			*s = candidate
			return nil
		}
	}
	return fmt.Errorf("unknown surface %q", name)
}

// Reject the old blob location spelling rather than silently losing it during
// JSON decoding. It cannot be guessed into a Ref: URIs are not local paths.
func (p *Part) UnmarshalJSON(b []byte) error {
	type plain Part
	var wire struct {
		plain
		// Path detects obsolete blob JSON so decoding refuses it instead of silently
		// dropping location.
		Path json.RawMessage `json:"path"`
	}
	if err := json.Unmarshal(b, &wire); err != nil {
		return err
	}
	if wire.Path != nil {
		return fmt.Errorf("legacy part path: use Ref")
	}
	*p = Part(wire.plain)
	return nil
}
