package common

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Config is everything about a request that is not conversation.
type Config struct {
	Vendor  Vendor
	Surface Surface
	Model   string
	BaseURL string
	APIKey  string

	// Route is what the credential's deployment permits, resolved once by
	// the engine from the credential provider's kind. It is here as DATA so
	// that a renderer can consult the rules of the route it is on without
	// learning anything about where the token came from.
	Route Route

	// Endpoints is where each vendor is reached, resolved once by the
	// composition root for every vendor rather than only the one we start
	// on. Switching models must move the endpoint with the model: the
	// renderer follows Vendor, so an endpoint left behind means we render
	// one vendor's dialect and post it to another vendor's host, carrying
	// the wrong key. BaseURL and APIKey above are simply the entry for the
	// vendor currently selected.
	//
	// The composition root owns this because resolving it needs the
	// lookup policy (environment, and for the binary the home settings
	// file) that only the root knows. An empty map is not a licence to
	// guess: a switch to a vendor with no entry is refused.
	Endpoints    map[Vendor]Endpoint
	SystemPrompt string
	MaxTokens    int
	// MaxToolRounds bounds dispatched batches per human turn. Zero uses 200.
	MaxToolRounds int
	Tools         []ToolDecl
	Cwd           string

	// DisableStreaming turns streaming off for every request.
	//
	// Named for the NON-default on purpose. Streaming on is what almost
	// everyone wants, and Go's zero value for a bool is false, so a
	// hand-built Config{} gets the default without help from a constructor.
	// A field named Stream would have made the zero value mean "off" and
	// quietly disagreed with the documented default. The cost is a negative
	// boolean, which reads worse and is worth it.
	//
	// Off is not merely slower. It is the only way to talk to a vendor whose
	// streaming is broken for the thing you need, and there is always one.
	DisableStreaming bool

	// Thinking controls the reasoning effort level.
	//
	// Zero (unset) resolves to ThinkingHigh in ThinkingFor, because that
	// is the default Bill wants and a hand-built Config{} must get it.
	// Same design as DisableStreaming: make the zero value mean the right
	// thing.
	Thinking ThinkingEffort
}

// ToolDecl is a tool as the MODEL sees it.
//
// Ephemeral tools ("round" or "turn") are auto-called by the engine — their
// results flow into context.Ephemera rather than dialogue. The LLM never sees
// them as callable tools.
type ToolDecl struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Schema      json.RawMessage `json:"schema"`
	Ephemeral   string          `json:"ephemeral,omitempty"` // "round", "turn", or ""
}

// Renderer turns a Context into one vendor's HTTP request.
type Renderer interface {
	Render(*Context, Config) (*http.Request, error)
}

// Parser turns one vendor's HTTP response back into events.
//
// Parse does not return []Event. It reports incrementally through cb, because
// a streamed response has content worth showing long before it has a
// finalized event. The events still arrive, through cb.Emit, and the event log
// is still the authority.
//
// There is deliberately no second method for streaming. Streaming is not a
// mode: a non-streamed response is a stream of length one, so it travels this
// same path and emits one delta per part. A ParseStream alongside Parse would
// have given every vendor two chances to disagree about what a tool call
// means, and would have made this interface three methods wide.
//
// The CALLER owns resp.Body and closes it. Parse only reads.
type Parser interface {
	Parse(resp *http.Response, cb StreamCallbacks) error
}

// StreamCallbacks is how a parser reports what it found, as it finds it.
//
// Every field is optional. Call them through the methods below, which
// nil-check, so a parser never has to.
type StreamCallbacks struct {
	// AllocPartID returns the next globally unique part ID. Vendor parsers
	// MUST call this instead of maintaining a local counter, so that two
	// turns never produce the same part_id. The GUI uses part_id to route
	// streaming deltas to the correct artifact; a collision routes turn N's
	// text into turn N-1's widget.
	AllocPartID func() uint64

	// OnDelta receives one incremental chunk.
	//
	// partID identifies the PART, not the chunk: every chunk belonging to
	// one part carries the SAME id, and the PartFinal for that part carries
	// it too. That shared id is the whole point. It is what lets a GUI
	// append chunks into one widget and then replace that widget's content
	// with the authoritative part when it lands. Minting a fresh id per
	// chunk would make the stream unreassemblable by anything downstream.
	OnDelta func(partID uint64, kind DeltaKind, chunk string)

	// OnEvent receives each finalized event, in order, for the event log.
	OnEvent func(Event)

	// OnPartFinal receives each completed part, with THE SAME partID its
	// deltas carried.
	//
	// The parser reports this rather than the caller reconstructing it, and
	// that is the whole reason this callback exists. A caller can only guess
	// the id by the part's position in the finished list, and the guess is
	// wrong on real wire formats: OpenAI streams a tool call's arguments
	// before it is knowable whether a text part will occupy index zero. The
	// code that CHOSE an id is the only code that can be trusted to repeat
	// it, so the id never has to mean "position", only "the same part".
	//
	// ORDERING: this is reported AFTER the event carrying the part, so an
	// observer that re-renders on a final is reading a log that already
	// contains it.
	OnPartFinal func(partID uint64, part Part)

	// OnFrame receives each raw wire frame: for SSE, one event's data bytes.
	//
	// This exists so the API log stays byte-complete when streaming. Without
	// it the response half of that log would simply go dark, because the
	// body is consumed by the parser instead of being read whole. The parser
	// is a stateless value with no Host, so logging is wired in by the
	// engine, which has one.
	OnFrame func(eventType string, data []byte)
}

// Delta reports an incremental chunk, if anyone is listening.
func (cb StreamCallbacks) Delta(partID uint64, kind DeltaKind, chunk string) {
	if cb.OnDelta != nil {
		cb.OnDelta(partID, kind, chunk)
	}
}

// Emit reports a finalized event, if anyone is listening.
func (cb StreamCallbacks) Emit(ev Event) {
	if cb.OnEvent != nil {
		cb.OnEvent(ev)
	}
}

// Final reports a completed part, if anyone is listening.
func (cb StreamCallbacks) Final(partID uint64, part Part) {
	if cb.OnPartFinal != nil {
		cb.OnPartFinal(partID, part)
	}
}

// Frame reports a raw wire frame, if anyone is listening.
func (cb StreamCallbacks) Frame(eventType string, data []byte) {
	if cb.OnFrame != nil {
		cb.OnFrame(eventType, data)
	}
}

// BodyOf returns the exact bytes the renderer produced.
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

// DefaultSurface returns the default API surface for a vendor.
//
// This is the fallback for a model nobody has recorded a surface for. Prefer
// SurfaceForModel, which consults the table first.
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

// SurfaceForModel reports which endpoint dialect a model is spoken to.
//
// The table wins where it has an opinion, and the vendor default applies
// where it does not. Doing it in that order is what lets one vendor serve two
// surfaces at once without a flag: a model migrates by gaining a row, not by
// the operator remembering to set something.
//
// The alternative — flipping the vendor default — would silently move every
// model of that vendor onto an endpoint most of them have never been tested
// against, including models that predate it. A per-model fact belongs in the
// per-model table.
func SurfaceForModel(model string, v Vendor) Surface {
	if f, ok := LookupModel(model); ok && f.Surface != 0 {
		return f.Surface
	}
	return DefaultSurface(v)
}

// Endpoint is where one vendor is reached: its base URL and the key that
// authenticates to it. The two travel together because they are useless
// apart, and sending one vendor's key to another vendor's host is the exact
// mistake this type exists to make impossible.
type Endpoint struct {
	BaseURL string
	APIKey  string
}

// ResolveEndpoints resolves an Endpoint for every vendor, reading each setting
// through look: a generic variable that overrides all vendors, a vendor
// specific variable, and a default used when neither is set.
//
// It takes a lookup function rather than reading the environment itself
// because the two composition roots are entitled to different policies. The
// binary also consults the user's home settings file; the embeddable library
// deliberately does not, since a library has no business reading its host's
// home directory. Sharing the table while injecting the policy keeps one copy
// of the vendor list, so adding a vendor cannot update one root and miss the
// other.
//
// Every vendor is resolved up front, not just the one selected at startup, so
// that switching models later is a lookup rather than a second copy of this
// table resolved under a different policy.
func ResolveEndpoints(look func(generic, specific, fallback string) string) map[Vendor]Endpoint {
	return map[Vendor]Endpoint{
		VendorAnthropic: {
			BaseURL: look("LLM_BASE_URL", "ANTHROPIC_BASE_URL", "https://api.anthropic.com"),
			APIKey:  look("LLM_API_KEY", "ANTHROPIC_API_KEY", ""),
		},
		VendorOpenAI: {
			BaseURL: look("LLM_BASE_URL", "OPENAI_BASE_URL", "https://api.openai.com"),
			APIKey:  look("LLM_API_KEY", "OPENAI_API_KEY", ""),
		},
		VendorGemini: {
			BaseURL: look("LLM_BASE_URL", "GEMINI_BASE_URL", "https://generativelanguage.googleapis.com"),
			APIKey:  look("LLM_API_KEY", "GEMINI_API_KEY", ""),
		},
	}
}
