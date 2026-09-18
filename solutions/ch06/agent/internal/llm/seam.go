package llm

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

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// SeamFor returns the renderer and parser for a vendor. This function is the
// only place in the program that switches on common.Vendor, which is why common.Vendor is an
// enum and Model is a string.
func SeamFor(v common.Vendor) (common.Renderer, common.Parser, error) {
	switch v {
	case common.VendorAnthropic:
		return anthropicSeam{}, anthropicSeam{}, nil
	case common.VendorOpenAI:
		return openAISeam{}, openAISeam{}, nil
	case common.VendorGemini:
		return geminiSeam{}, geminiSeam{}, nil
	}
	return nil, nil, fmt.Errorf("no seam for vendor %d", uint8(v))
}

// ---------------------------------------------------------------------------
// Shared rendering helpers. Everything here is vendor-independent: it is the
// part of "render" that is genuinely about the conversation rather than about
// one vendor's spelling of it.
// ---------------------------------------------------------------------------

// renderable is one entry reduced to what every vendor needs to know about it.
type renderable struct {
	Entry common.Entry
	Calls []common.ToolCallPart
	Texts []string
	Tools []common.ToolResultPart
	Blobs []common.BlobPart

	// Refs are locators that SURVIVED a redaction: the common.Ref carried forward
	// from a part that was superseded. They have no MIME type and no bytes,
	// only a place the content still is. A vendor that can act on a remote
	// reference renders them; one that cannot ignores them, because the stub
	// text has already said that something was removed.
	Refs []common.Ref

	Raw []common.OpaquePart
}

// classify splits an entry's parts by kind. Doing this once, here, is what
// stops each renderer from growing its own type switch — and a type switch
// per vendor is how three near-identical clients start.
func classify(e common.Entry, cfg common.Config) (renderable, error) {
	r := renderable{Entry: e}
	for _, p := range e.Parts {
		switch v := p.(type) {
		case common.TextPart:
			if v.Text != "" {
				r.Texts = append(r.Texts, v.Text)
			}
		case common.RedactedPart:
			r.Texts = append(r.Texts, v.Stub)
			if !v.Ref.Zero() {
				r.Refs = append(r.Refs, v.Ref)
			}
		case common.ToolCallPart:
			r.Calls = append(r.Calls, v)
		case common.ToolResultPart:
			r.Tools = append(r.Tools, v)
			// A redaction of a tool RESULT leaves its stub nested one level
			// down, inside the common.ToolResultPart that survived. The locator it
			// carried forward has to be lifted out here or it is invisible to
			// every renderer.
			for _, sub := range v.Parts {
				if rp, ok := sub.(common.RedactedPart); ok && !rp.Ref.Zero() {
					r.Refs = append(r.Refs, rp.Ref)
				}
			}
		case common.BlobPart:
			// Media asymmetry is a LOUD error. The model table says what each
			// model accepts; a missing row is an unknown model and equally loud.
			if err := checkMedia(v.MIME, cfg.Model); err != nil {
				return r, err
			}
			r.Blobs = append(r.Blobs, v)
		case common.OpaquePart:
			r.Raw = append(r.Raw, v)
		}
	}
	return r, nil
}

// checkMedia refuses to render a MIME type the model does not accept.
// A missing model row is an equally loud refusal — never guess.
func checkMedia(mime, model string) error {
	features, ok := common.LookupModel(model)
	if !ok {
		return fmt.Errorf("cannot render %s part: unknown model %q (no entry in the model table)", mime, model)
	}
	var need common.Media
	switch {
	case strings.HasPrefix(mime, "image/"):
		need = common.MediaImage
	case strings.HasPrefix(mime, "audio/"):
		need = common.MediaAudio
	case strings.HasPrefix(mime, "video/"):
		need = common.MediaVideo
	case strings.HasPrefix(mime, "application/pdf"):
		need = common.MediaDocument
	default:
		// Unknown MIME prefix — let it through, the vendor will reject if needed.
		return nil
	}
	if features.Media&need == 0 {
		return fmt.Errorf("cannot render %s part to model %q: it does not accept %s (refusing to silently drop content)", mime, model, need)
	}
	return nil
}

// resultText flattens a tool result to a string. Vendors all want a scalar
// here in the simple case, and the flattening rule is the same for all three.
func resultText(res common.ToolResultPart) string {
	var b strings.Builder
	for _, p := range res.Parts {
		switch v := p.(type) {
		case common.TextPart:
			b.WriteString(v.Text)
		case common.RedactedPart:
			b.WriteString(v.Stub)
		case common.BlobPart:
			b.WriteString("[" + v.MIME + " at " + v.Ref.Locator + "]")
		}
	}
	return b.String()
}

// synthID produces a tool-call id for a vendor that requires one when the
// issuing vendor did not supply it.
//
// Derived from common.Seq, never generated randomly. `replay` compares bytes, and a
// random id is one of the four ways non-determinism gets into a renderer — the
// others being the clock, Go's randomized map iteration, and iteration over a
// set. The seam and the determinism rule meet at exactly this field, and
// students who wire them up independently will collide here.
func synthID(prefix string, seq common.Seq, n int) string {
	return fmt.Sprintf("%s_%d_%d", prefix, uint64(seq), n)
}

// callIDFor returns the id to use for a call when rendering to target.
//
// Pass the id through when it exists: it is already unique and already
// consistent between the call and its result, and rewriting it gains nothing.
// Synthesize only when the issuing vendor gave us nothing to pass — which is
// the real motivation, and it is not "the other vendor's id is meaningless".
// A target vendor rejects a MISSING correlation id, not a foreign-looking one.
func callIDFor(c common.ToolCallPart, prefix string, seq common.Seq, n int) string {
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

// ---------------------------------------------------------------------------
// Shared parsing helpers for the streaming seam.
// ---------------------------------------------------------------------------

// isSSE reports whether a response body is a Server-Sent Events stream.
//
// The CONTENT TYPE decides, not what we asked for. A vendor is free to answer
// a streaming request with a single JSON document — an error almost always
// comes back that way even when `stream: true` was sent — and a parser that
// trusts the request instead of the response will try to read a JSON object
// as SSE frames and find nothing at all.
func isSSE(resp *http.Response) bool {
	return strings.Contains(resp.Header.Get("Content-Type"), "text/event-stream")
}

// partID converts a zero-based vendor block index into a part id.
//
// One-based, because zero is a perfectly ordinary index and a perfectly
// suspicious id. Deltas and the PartFinal that supersedes them must agree on
// this number, so it is derived from the vendor's own block index and never
// from a counter that advances per chunk.
func partID(blockIndex int) uint64 { return uint64(blockIndex) + 1 }

// emitLengthOneDeltas reports an already-complete part list as deltas.
//
// This is what makes "streaming is not a mode" true rather than aspirational.
// A response that arrived as one JSON document still produces deltas — one
// per part — so an observer downstream cannot tell, and does not need to
// know, whether the bytes trickled or landed all at once. Only the chunk
// count differs.
//
// thinking extracts the human-readable text from a vendor's opaque reasoning
// block. It returns "" when there is nothing safe to show, which is the right
// answer for redacted reasoning: the part is still carried in the event, it
// simply has no display text.
func emitLengthOneDeltas(parts common.PartList, cb common.StreamCallbacks, thinking func(common.OpaquePart) string) {
	for i, p := range parts {
		id := partID(i)
		switch v := p.(type) {
		case common.TextPart:
			if v.Text != "" {
				cb.Delta(id, common.DeltaText, v.Text)
			}
		case common.ToolCallPart:
			// The concatenation of a tool call's chunks is its NAME followed
			// by its arguments JSON, because that is the order the streaming
			// path produces them in: the name arrives with the block header,
			// the arguments in fragments after it. Both paths therefore
			// reassemble to the same display string.
			cb.Delta(id, common.DeltaToolCall, v.Name)
			if args := string(jsonObject(v.Args)); args != "{}" {
				cb.Delta(id, common.DeltaToolCall, args)
			}
		case common.OpaquePart:
			if thinking == nil {
				continue
			}
			if text := thinking(v); text != "" {
				cb.Delta(id, common.DeltaThinking, text)
			}
		}
	}
}

// emitFinals reports each part as complete, reusing the ids its deltas used.
//
// Called after the event that carries the parts, never before: an observer
// that re-renders a widget when a part finalizes should be reading a log that
// already has the part in it.
func emitFinals(parts common.PartList, cb common.StreamCallbacks) {
	for i, p := range parts {
		cb.Final(partID(i), p)
	}
}

// parseErrorResponse turns a non-200 into a single common.ErrorOccurred event.
//
// A 429 is not a response with an unfortunate mood. It is an error, and it
// gets the loud event rather than being folded into an empty assistant turn.
func parseErrorResponse(resp *http.Response, cb common.StreamCallbacks) error {
	body, _ := io.ReadAll(resp.Body)
	cb.Frame("error", body)
	cb.Emit(common.Event{Type: common.ErrorOccurred, Error: &common.ErrorData{
		Status:  resp.StatusCode,
		Message: vendorErrorMessage(body),
	}})
	return nil
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
	// common.BodyOf recovers the exact bytes later. A side table keyed by *Request
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
