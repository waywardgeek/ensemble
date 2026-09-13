package main

// The seam.
//
//	The context is the truth. A renderer turns truth into one vendor's
//	request. A parser turns one vendor's response back into truth.
//	Distortion lives in those two places and nowhere else.
//
// Two methods, no vendor types in either signature. That is the whole remedy,
// and its smallness is the point. Contrast the shape that cost 30,000 lines:
//
//	type AIClientInterface interface {
//	    SendMessage(msgs []ClaudeMessage) (*ClaudeResponse, error)
//	    CountTokens(msgs []ClaudeMessage) (int, error)
//	    // ...eleven more, each shaped by what ClaudeClient happened to do
//	}
//
// The tell is visible without knowing the story: vendor types in the
// signature. ClaudeMessage in the interface means the interface IS the Claude
// client, and the second implementation can only be a copy-paste.

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Config is everything about a request that is not conversation.
//
// The system prompt lives HERE, not in the context. It is an output, computed
// from context plus configuration. For now a constant string is a perfectly
// good renderer; Chapter 6 replaces it with generation from skills, and that
// is a pure addition only because it was never a stored value.
type Config struct {
	Vendor       Vendor
	Surface      Surface
	Model        string
	BaseURL      string
	APIKey       string
	SystemPrompt string
	MaxTokens    int

	// AcceptsAudio is the media-capability flag. An audio part rendered for a
	// text-only model must RAISE, never silently drop: a fallback converts an
	// invariant violation into silently-wrong output.
	AcceptsAudio bool
}

type Renderer interface {
	Render(*Context, Config) (*http.Request, error)
}

type Parser interface {
	// Parse returns EVENTS, not a message and not a context. There is exactly
	// one path into the context — append events, run the reducer — so a vendor
	// response and a human keystroke enter by the same door. Give the parser
	// the power to mutate context directly and you have quietly created a
	// second reducer that nobody will remember to keep total.
	Parse(status int, body []byte) ([]Event, error)
}

// SeamFor returns the renderer and parser for a vendor. This function is the
// only place in the program that switches on Vendor, which is why Vendor is an
// enum and Model is a string.
func SeamFor(v Vendor) (Renderer, Parser, error) {
	switch v {
	case VendorAnthropic:
		return anthropicSeam{}, anthropicSeam{}, nil
	case VendorOpenAI:
		return openAISeam{}, openAISeam{}, nil
	case VendorGemini:
		return geminiSeam{}, geminiSeam{}, nil
	}
	return nil, nil, fmt.Errorf("no seam for vendor %d", uint8(v))
}

// DefaultSurface is the surface each vendor's seam targets in this chapter.
// OpenAI's Chat Completions and Gemini's generateContent are chosen over their
// newer siblings deliberately: they are the shapes the chapter's exhibits show,
// and they preserve the difficulty ordering the chapter's prediction depends on.
func DefaultSurface(v Vendor) Surface {
	switch v {
	case VendorOpenAI:
		return SurfaceChatCompletions
	case VendorGemini:
		return SurfaceGenerateContent
	default:
		return SurfaceMessages
	}
}

// ---------------------------------------------------------------------------
// Shared rendering helpers. Everything here is vendor-independent: it is the
// part of "render" that is genuinely about the conversation rather than about
// one vendor's spelling of it.
// ---------------------------------------------------------------------------

// renderable is one entry reduced to what every vendor needs to know about it.
type renderable struct {
	Entry Entry
	Calls []ToolCallPart
	Texts []string
	Tools []ToolResultPart
	Blobs []BlobPart

	// Refs are locators that SURVIVED a redaction: the Ref carried forward
	// from a part that was superseded. They have no MIME type and no bytes,
	// only a place the content still is. A vendor that can act on a remote
	// reference renders them; one that cannot ignores them, because the stub
	// text has already said that something was removed.
	Refs []Ref

	Raw []OpaquePart
}

// classify splits an entry's parts by kind. Doing this once, here, is what
// stops each renderer from growing its own type switch — and a type switch
// per vendor is how three near-identical clients start.
func classify(e Entry, cfg Config) (renderable, error) {
	r := renderable{Entry: e}
	for _, p := range e.Parts {
		switch v := p.(type) {
		case TextPart:
			if v.Text != "" {
				r.Texts = append(r.Texts, v.Text)
			}
		case RedactedPart:
			r.Texts = append(r.Texts, v.Stub)
			if !v.Ref.zero() {
				r.Refs = append(r.Refs, v.Ref)
			}
		case ToolCallPart:
			r.Calls = append(r.Calls, v)
		case ToolResultPart:
			r.Tools = append(r.Tools, v)
			// A redaction of a tool RESULT leaves its stub nested one level
			// down, inside the ToolResultPart that survived. The locator it
			// carried forward has to be lifted out here or it is invisible to
			// every renderer.
			for _, sub := range v.Parts {
				if rp, ok := sub.(RedactedPart); ok && !rp.Ref.zero() {
					r.Refs = append(r.Refs, rp.Ref)
				}
			}
		case BlobPart:
			// Media asymmetry is a LOUD error.
			if strings.HasPrefix(v.MIME, "audio/") && !cfg.AcceptsAudio {
				return r, fmt.Errorf("cannot render %s part to model %q: it does not accept audio (refusing to silently drop content)", v.MIME, cfg.Model)
			}
			r.Blobs = append(r.Blobs, v)
		case OpaquePart:
			r.Raw = append(r.Raw, v)
		}
	}
	return r, nil
}

// resultText flattens a tool result to a string. Vendors all want a scalar
// here in the simple case, and the flattening rule is the same for all three.
func resultText(res ToolResultPart) string {
	var b strings.Builder
	for _, p := range res.Parts {
		switch v := p.(type) {
		case TextPart:
			b.WriteString(v.Text)
		case RedactedPart:
			b.WriteString(v.Stub)
		case BlobPart:
			b.WriteString("[" + v.MIME + " at " + v.Ref.Locator + "]")
		}
	}
	return b.String()
}

// synthID produces a tool-call id for a vendor that requires one when the
// issuing vendor did not supply it.
//
// Derived from Seq, never generated randomly. `replay` compares bytes, and a
// random id is one of the four ways non-determinism gets into a renderer — the
// others being the clock, Go's randomized map iteration, and iteration over a
// set. The seam and the determinism rule meet at exactly this field, and
// students who wire them up independently will collide here.
func synthID(prefix string, seq Seq, n int) string {
	return fmt.Sprintf("%s_%d_%d", prefix, uint64(seq), n)
}

// callIDFor returns the id to use for a call when rendering to target.
//
// Pass the id through when it exists: it is already unique and already
// consistent between the call and its result, and rewriting it gains nothing.
// Synthesize only when the issuing vendor gave us nothing to pass — which is
// the real motivation, and it is not "the other vendor's id is meaningless".
// A target vendor rejects a MISSING correlation id, not a foreign-looking one.
func callIDFor(c ToolCallPart, prefix string, seq Seq, n int) string {
	if c.CallID != "" {
		return c.CallID
	}
	return synthID(prefix, seq, n)
}

// jsonObject guarantees a valid JSON object for vendors that reject a bare
// null or a scalar where arguments are expected.
func jsonObject(raw json.RawMessage) json.RawMessage {
	t := strings.TrimSpace(string(raw))
	if t == "" || t == "null" {
		return json.RawMessage(`{}`)
	}
	return raw
}

// newJSONRequest builds the HTTP request. Rendering produces a request that is
// never sent by `render` — that is what makes the whole seam testable as a
// byte comparison, with no network and no key.
func newJSONRequest(method, url string, body any, headers map[string]string) (*http.Request, error) {
	buf, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		return nil, err
	}
	// http.NewRequest populates GetBody for a *strings.Reader, which is how
	// BodyOf recovers the exact bytes later. A side table keyed by *Request
	// would also work and would be a field that grows forever — the same
	// mistake the context section deletes.
	req, err := http.NewRequest(method, url, strings.NewReader(string(buf)))
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// BodyOf returns the exact bytes the renderer produced, without consuming the
// request. `render` prints these; `chat` sends them.
func BodyOf(req *http.Request) ([]byte, error) {
	if req.GetBody == nil {
		return nil, fmt.Errorf("request has no recoverable body")
	}
	rc, err := req.GetBody()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(rc)
}
