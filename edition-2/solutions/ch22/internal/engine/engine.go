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
	"sync"

	"ensemble/internal/common"
)

// This is rendered output, not conversation state. Later instruction generation
// can replace the constant without rewriting any saved conversation.
const systemPrompt = "You are a helpful assistant built from raw HTTP calls in Chapter 1 of The Art of Building AI Coding Agents. Answer briefly."

// One client and accounting accumulator belong to one Engine. Configuration
// remains reachable through Agent, so sibling Agents never share credentials
// or a mutable vendor selection by accident.
type engine struct {
	parent      common.Agent
	client      *http.Client
	usage       common.Usage
	last        common.Usage
	byModel     map[usageKey]common.Usage
	lens        common.CacheLens
	mu          sync.Mutex
	billingOnce sync.Once
}

// New creates an Agent-owned HTTP engine. DefaultTransport bounds connection
// setup and TLS negotiation, but model generation has no wall-clock deadline.
// The caller's context still cancels an interrupted or explicitly bounded turn.
func New(parent common.Agent) common.Engine {
	if parent == nil {
		panic("Engine requires Agent")
	}
	return &engine{parent: parent, client: &http.Client{}}
}

// Agent returns the required owning Agent rather than a copied service bundle.
func (e *engine) Agent() common.Agent { return e.parent }

// Usage returns measured session counts under the Engine's accounting lock.
func (e *engine) Usage() common.Usage {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.usage
}

// An explicitly resolved identity controls replay matching while Model still
// routes the HTTP request. Without resolution, equality is conservative: an
// alias differing from the captured producer will not receive opaque bytes.
func (e *engine) Target() common.Provenance {
	cfg := e.parent.Config()
	model := cfg.ResolvedModel
	if model == "" {
		model = cfg.Model
	}
	surface := common.Messages
	switch cfg.Vendor {
	case common.OpenAI:
		surface = common.ChatCompletions
		if e.Features().Responses {
			surface = common.Responses
		}
	case common.Gemini:
		surface = common.GenerateContent
	}
	return common.Provenance{Vendor: cfg.Vendor, Model: model, Surface: surface}
}

// Render has no clock or conversation mutation. Local media reads its Ref.
// Offline replay and live requests share these bytes; credentials and URL
// construction are transport-only. Referenced files must still contain the same bytes.
func (e *engine) Render(c *common.Context) ([]byte, error) {
	// A committed projection cannot be rendered alongside an unresolved active
	// corpus. Recovery clears this receipt only after promotion/archive succeed.
	if c.Compression != nil && c.Compression.Committed != nil {
		return nil, errors.New("memory commit requires disk recovery before rendering")
	}
	switch e.parent.Config().Vendor {
	case common.Anthropic:
		return e.renderAnthropic(c)
	case common.OpenAI:
		if e.Features().Responses {
			return e.renderResponses(c)
		}
		return e.renderOpenAI(c)
	case common.Gemini:
		return e.renderGemini(c)
	default:
		return nil, errors.New("unknown vendor")
	}
}

// Send performs one HTTP exchange from already rendered bytes and returns facts
// without changing conversation.
func (e *engine) Send(ctx context.Context, body []byte, cb common.StreamCallbacks) ([]common.Event, error) {
	return e.send(ctx, body, cb, e.parent.Config(), true)
}

// Auxiliary calls share transport and measured costs but leave main occupancy
// and cache comparisons alone. Observe main bytes only after resolving secrets.
func (e *engine) send(ctx context.Context, body []byte, cb common.StreamCallbacks, cfg common.Config, main bool) (events []common.Event, err error) {
	// Error bodies and URLs may echo credentials. Preserve useful status/error
	// identity but never return remote error prose or log authentication headers.
	defer func() {
		if err != nil {
			e.parent.Ensemble().Logf("%s request failed: %v", e.parent.Config().Vendor, err)
			events = append(events, common.Event{Type: "error_occurred", Error: &common.ErrorData{Message: err.Error()}})
		}
	}()
	secrets := []string{cfg.APIKey}
	// Resolve a request-local bearer. Rendered bytes and saved configuration never
	// receive rotating OAuth tokens.
	if cfg.Credentials != nil {
		cfg.APIKey, err = cfg.Credentials.GetBearerToken(ctx)
		if err != nil {
			return nil, err
		}
	}
	secrets = append(secrets, cfg.APIKey)
	redact := func(data []byte) []byte {
		for _, secret := range secrets {
			if secret != "" {
				data = bytes.ReplaceAll(data, []byte(secret), []byte("[REDACTED]"))
			}
		}
		return data
	}
	features, _ := LookupModel(cfg.Model)
	if cfg.Vendor == common.OpenAI && features.Responses {
		e.billingOnce.Do(func() {
			kind := common.CredentialAPIKey
			if cfg.Credentials != nil {
				kind = cfg.Credentials.Kind()
			}
			e.parent.Ensemble().Observe(e.parent, common.BillingSelected{Source: kind})
			e.parent.Ensemble().Logf("OpenAI billing source: %s", kind)
		})
	}
	// URLs and authentication are deliberately outside Render. Offline replay
	// needs neither credentials nor a reachable provider, and bodies cannot
	// accidentally serialize an API key from the configuration.
	path := "/v1/messages"
	switch cfg.Vendor {
	case common.OpenAI:
		path = "/v1/chat/completions"
		if features.Responses {
			path = "/v1/responses"
		}
	case common.Gemini:
		path = "/v1beta/models/" + url.PathEscape(cfg.Model) + ":generateContent"
		if StreamingFor(cfg) != 0 {
			path = "/v1beta/models/" + url.PathEscape(cfg.Model) + ":streamGenerateContent?alt=sse"
		}
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
		if features.InlineTools {
			req.Header.Set("anthropic-beta", "inline-tools-2026-09-15")
		}
	case common.OpenAI:
		req.Header.Set("authorization", "Bearer "+cfg.APIKey)
	case common.Gemini:
		req.Header.Set("x-goog-api-key", cfg.APIKey)
	}
	// The actor already recorded request_sent and consumed delivered ephemera.
	// This worker returns response facts without borrowing the live projection.
	safeBody := redact(body)
	e.parent.Ensemble().APILog("request", safeBody)
	// Observe one logical request after resolving and redacting its bearer.
	// A 401 renewal retries these bytes; it must not rotate the lens again.
	// Keep prediction local when interrupted workers overlap newer requests.
	predicted := 0.0
	if main && e.lens != nil {
		predicted = e.lens.ObserveRequest(fmt.Sprintf("%s/%s", cfg.Vendor, cfg.Model), safeBody)
	}
	resp, err := e.client.Do(req)
	// Retry authentication once using the same billing source. A 401 alone does
	// not delete credentials; only the token endpoint can reject their renewal.
	if err == nil && resp.StatusCode == http.StatusUnauthorized && cfg.Credentials != nil && cfg.Credentials.Kind() == common.CredentialChatGPT {
		resp.Body.Close()
		store, ok := cfg.Credentials.(common.AuthStore)
		if !ok {
			return nil, errors.New("ChatGPT credential cannot renew")
		}
		if err := store.Refresh(ctx, cfg.APIKey); err != nil {
			return nil, err
		}
		cfg.APIKey, err = store.GetBearerToken(ctx)
		if err != nil {
			return nil, err
		}
		secrets = append(secrets, cfg.APIKey)
		req.Body, err = req.GetBody()
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
		resp, err = e.client.Do(req)
	}
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
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 65536))
		return events, providerError(resp.StatusCode, data)
	}

	// Frames share the owner's trace facility. Headers and URLs never enter it;
	// remote bytes are redacted before leaving Engine, even if a server echoes a key.
	onFrame := cb.OnFrame
	cb.OnFrame = func(kind string, data []byte) {
		safe := redact(data)
		e.parent.Ensemble().APILog(kind, safe)
		if onFrame != nil {
			onFrame(kind, safe)
		}
	}
	// Remote content is untrusted too: a server echo must not turn the bearer
	// into a saved fact, tool argument or observer fragment after safe logging.
	onDelta, onFinal := cb.OnDelta, cb.OnPartFinal
	cb.OnDelta = func(id uint64, kind common.DeltaKind, chunk string) {
		if onDelta != nil {
			onDelta(id, kind, string(redact([]byte(chunk))))
		}
	}
	cb.OnPartFinal = func(id uint64, part common.Part) {
		if onFinal != nil {
			raw, _ := json.Marshal(part)
			json.Unmarshal(redact(raw), &part)
			onFinal(id, part)
		}
	}
	onEvent := cb.OnEvent
	cb.OnEvent = func(event common.Event) {
		raw, _ := json.Marshal(event)
		json.Unmarshal(redact(raw), &event)
		events = append(events, event)
		if event.Response != nil {
			// Replay independently sums these same normalized categories. No vendor
			// cache or reasoning interpretation belongs in session accounting.
			u := event.Response.Usage
			e.account(cfg, u)
			if main {
				e.mu.Lock()
				e.last = u
				e.mu.Unlock()
				if e.lens != nil {
					e.lens.ObserveUsage(u, predicted)
				}
			}
		}
		if onEvent != nil {
			onEvent(event)
		}
	}
	if cfg.Vendor == common.OpenAI && features.Responses {
		err = e.parseResponses(resp, cb, cfg)
	} else {
		err = e.parse(resp, cb, cfg)
	}
	return events, err
}

// Pending ephemera follows retained dialogue in the renderer input; Anthropic
// places inline declarations after that complete user turn. It remains absent
// from Dialogue; RequestSent consumes it during both live application and replay.
func entries(c *common.Context) []common.Entry {
	out := append([]common.Entry(nil), c.Dialogue...)
	for i, entry := range out {
		if entry.Kind == common.KindHandoff {
			// The call and its acknowledgment were cleared at the checkpoint.
			// Render the completed status alongside the note, not just its contents.
			// Keep stored notes unchanged, including those restored from old saves.
			out[i].Parts = append([]common.Part{{Type: "text", Text: "micro_handoff completed. Your checkpoint note:"}}, entry.Parts...)
		}
	}
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

// BaseSystemPrompt selects authored role instructions without model-specific
// additions. An explicitly empty primary skill suppresses the earlier clients'
// default. Child composition uses this base so changing models cannot retain a
// policy that was rendered only for the child's earlier model.
func BaseSystemPrompt(a common.Agent) string {
	if prompt := a.Config().SystemPrompt; prompt != "" || a.Skills() != nil {
		return prompt
	}
	return systemPrompt
}

// SystemPrompt resolves the Agent's role and current model policy for rendering and save metadata.
func (e *engine) SystemPrompt() string {
	prompt := BaseSystemPrompt(e.parent)
	const visibleReasoning = "Before calling tools, include ordinary user-visible text in the same response before the tool calls, explaining what you will do and why. Hidden thinking or a reasoning summary does not satisfy this requirement. Tool calls without this preceding text will be refused; retry with the explanation before the calls."
	// Saved prompts and inherited child instructions can already contain the
	// effective prompt. Do not append the same requirement again.
	if e.Features().RequireVisibleReasoning && !strings.Contains(prompt, visibleReasoning) {
		if prompt != "" {
			prompt += "\n\n"
		}
		prompt += visibleReasoning
	}
	return prompt
}

// Features follows the Agent's model selection; fixture capabilities do not imply live support.
func (e *engine) Features() common.ModelFeatures {
	features, _ := LookupModel(e.parent.Config().Model)
	return features
}

// Declarations come from recorded history after startup. Non-inline providers
// pay the prefix rewrite; inline providers retain exactly their initial list.
func (e *engine) declarations(c *common.Context) []common.ToolDeclaration {
	if c.StartupTools == nil {
		return e.parent.Declarations()
	}
	if e.Features().InlineTools {
		return *c.StartupTools
	}
	return c.Tools
}

// AttachCacheLens links both sides at construction; a foreign owner is invalid.
func (e *engine) AttachCacheLens(l common.CacheLens) {
	if l == nil || l.Engine() != e {
		panic("CacheLens requires owning Engine")
	}
	e.lens = l
}

// CacheLens exposes the optional Engine-owned instrument.
func (e *engine) CacheLens() common.CacheLens { return e.lens }
