// Package engine owns vendor translation, HTTP exchanges and measured usage.
package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"ensemble/internal/common"
)

// This is rendered output, not conversation state. Later instruction generation
// can replace the constant without rewriting any saved conversation.
const systemPrompt = "You are a helpful assistant built from raw HTTP calls in Chapter 1 of The Art of Building AI Coding Agents. Answer briefly."

// One client and accounting accumulator belong to one Engine. Configuration
// remains reachable through Agent, so sibling Agents never share credentials
// or a mutable vendor selection by accident.
type engine struct {
	parent common.Agent
	client *http.Client
	usage  common.Usage
}

func New(parent common.Agent) common.Engine {
	if parent == nil {
		panic("Engine requires Agent")
	}
	return &engine{parent: parent, client: &http.Client{Timeout: 60 * time.Second}}
}
func (e *engine) Agent() common.Agent { return e.parent }
func (e *engine) Usage() common.Usage { return e.usage }

// An explicitly resolved identity controls replay matching while Model still
// routes the HTTP request. Without resolution, equality is conservative: an
// alias differing from the captured producer will not receive opaque bytes.
func (e *engine) target() common.Provenance {
	cfg := e.parent.Config()
	model := cfg.ResolvedModel
	if model == "" {
		model = cfg.Model
	}
	surface := common.Messages
	switch cfg.Vendor {
	case common.OpenAI:
		surface = common.ChatCompletions
	case common.Gemini:
		surface = common.GenerateContent
	}
	return common.Provenance{Vendor: cfg.Vendor, Model: model, Surface: surface}
}

// Render has no clock, I/O or mutation. Offline replay and live requests share
// these exact bytes; credentials and URL construction are transport-only.
func (e *engine) Render(c *common.Context) ([]byte, error) {
	switch e.parent.Config().Vendor {
	case common.Anthropic:
		return e.renderAnthropic(c)
	case common.OpenAI:
		return e.renderOpenAI(c)
	case common.Gemini:
		return e.renderGemini(c)
	default:
		return nil, errors.New("unknown vendor")
	}
}

func (e *engine) Send(ctx context.Context, c *common.Context) (events []common.Event, err error) {
	// Error bodies and URLs may echo credentials. Preserve useful status/error
	// identity but never return those remote bytes or log request payloads.
	defer func() {
		if err != nil {
			e.parent.Ensemble().Logger().Printf("%s request failed: %v", e.parent.Config().Vendor, err)
			events = append(events, common.Event{Type: "error_occurred", Error: &common.ErrorData{Message: err.Error()}})
		}
	}()
	body, err := e.Render(c)
	if err != nil {
		return nil, err
	}
	// URLs and authentication are deliberately outside Render. Offline replay
	// needs neither credentials nor a reachable provider, and bodies cannot
	// accidentally serialize an API key from the configuration.
	cfg := e.parent.Config()
	path := "/v1/messages"
	switch cfg.Vendor {
	case common.OpenAI:
		path = "/v1/chat/completions"
	case common.Gemini:
		path = "/v1beta/models/" + url.PathEscape(cfg.Model) + ":generateContent"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(cfg.BaseURL, "/")+path, bytes.NewReader(body))
	if err != nil {
		return nil, errors.New("invalid provider endpoint")
	}
	req.Header.Set("content-type", "application/json")
	switch cfg.Vendor {
	case common.Anthropic:
		req.Header.Set("x-api-key", cfg.APIKey)
		req.Header.Set("anthropic-version", "2023-06-01")
	case common.OpenAI:
		req.Header.Set("authorization", "Bearer "+cfg.APIKey)
	case common.Gemini:
		req.Header.Set("x-goog-api-key", cfg.APIKey)
	}
	// Record sending before consuming ephemera. Send's caller appends returned
	// facts afterward; no parser can mutate or bypass the reducer.
	events = append(events, common.Event{Type: "request_sent", Request: &common.RequestData{To: e.target()}})
	resp, err := e.client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return events, ctx.Err()
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return events, context.DeadlineExceeded
		}
		return events, errors.New("provider transport failed")
	}
	// Close even rejected responses. HTTP status is enough for a safe diagnostic;
	// provider error text is not trusted to avoid echoing supplied secrets.
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return events, fmt.Errorf("provider HTTP status %d", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return events, errors.New("reading provider response failed")
	}
	// Every parser normalizes into one response fact. There is no vendor branch
	// in the reducer, and no parser has permission to mutate conversation state.
	var response common.ResponseData
	switch cfg.Vendor {
	case common.Anthropic:
		response, err = e.parseAnthropic(data)
	case common.OpenAI:
		response, err = e.parseOpenAI(data)
	case common.Gemini:
		response, err = e.parseGemini(data)
	}
	if err != nil {
		return events, err
	}
	// Session accounting measures accepted exchanges. The replay projection
	// independently derives historical totals from stored response usage.
	// Both consume already-disjoint categories; neither reinterprets vendors.
	e.usage.Input += response.Usage.Input
	e.usage.CacheWrite += response.Usage.CacheWrite
	e.usage.CacheRead += response.Usage.CacheRead
	e.usage.Output += response.Usage.Output
	return append(events, common.Event{Type: "response_ended", Response: &response}), nil
}

// All renderers append pending ephemera at the request tail. It remains absent
// from Dialogue; RequestSent consumes it during both live application and replay.
func entries(c *common.Context) []common.Entry {
	out := append([]common.Entry(nil), c.Dialogue...)
	if len(c.Ephemera) > 0 {
		out = append(out, common.Entry{Actor: common.System, Parts: c.Ephemera})
	}
	return out
}

// A textual tool result may contain redaction stubs. Opaque bytes and tool
// arguments must not turn into visible prose merely because a vendor wants
// its result content represented as a string.
func text(parts []common.Part) string {
	var b strings.Builder
	for _, p := range parts {
		if p.Type == "text" {
			b.WriteString(p.Text)
		}
		if p.Type == "redacted" {
			b.WriteString(p.Stub)
		}
	}
	return b.String()
}

// Missing correlation IDs are deterministic, including multiple calls in one
// entry. Existing vendor-issued IDs always pass through unchanged.
func callID(entry common.Entry, index int, p common.Part) string {
	if p.CallID != "" {
		return p.CallID
	}
	return fmt.Sprintf("call_%d_%d", entry.Seq, index)
}

// Gemini needs the function name on a result, though the result fact only
// needs CallID. Resolve existing information here instead of copying a second
// name into every tool return and risking disagreement.
func callFor(c *common.Context, id string) (common.Entry, int, common.Part, error) {
	for _, entry := range c.Dialogue {
		for i, p := range entry.Parts {
			if p.Type == "tool_call" && p.CallID == id {
				return entry, i, p, nil
			}
		}
	}
	return common.Entry{}, 0, common.Part{}, errors.New("tool result has no matching call")
}

// Blob storage is deliberately a reference-only shape in this chapter. No
// media capability is implemented yet. Even a remote form that needs no local
// fetch is deferred; unsupported media must fail rather than disappear.
func (e *engine) refuseBlob(p common.Part) error {
	if p.Type == "blob" && strings.HasPrefix(p.MIME, "audio/") && !e.parent.Config().Audio {
		return errors.New("audio is unsupported by this target")
	}
	if p.Type == "blob" {
		return errors.New("blob media rendering is unsupported")
	}
	return nil
}

// OpenAI wraps arguments in a JSON string; the other providers send objects.
// Normalize object ordering at capture so that difference does not leak past
// parsing. This does not interpret or execute the function's arguments.
func canonical(raw json.RawMessage) (json.RawMessage, error) {
	var value any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber() // Arguments may contain integers beyond float64 precision.
	if err := decoder.Decode(&value); err != nil {
		return nil, errors.New("invalid tool arguments")
	}
	return json.Marshal(value)
}
