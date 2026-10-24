// Package engine owns HTTP and vendor translation, never retained dialogue.
package engine

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"ensemble/internal/common"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const systemPrompt = "You are a helpful coding assistant. Use the available tools when needed. Base claims about files and commands on returned results, and answer clearly."

type engine struct {
	mu        sync.Mutex
	parent    common.Agent
	client    *http.Client
	liveUsage common.Usage // Only Exchange charges this process; Parse and replay never do.
	cache     common.Cache
	lastModel string
	last      common.Usage
	requests  int
	buckets   []common.UsageBucket
}

func New(parent common.Agent) (*engine, error) {
	if parent == nil {
		return nil, errors.New("Engine requires an Agent parent")
	}
	return &engine{parent: parent, last: common.UnknownUsage(), client: &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}, nil
}
func (e *engine) SetCache(c common.Cache) {
	if c == nil || e.cache != nil || c.Parent() != e {
		panic("invalid Engine/Cache wiring")
	}
	e.cache = c
}
func (e *engine) Cache() common.Cache  { return e.cache }
func (e *engine) Parent() common.Agent { return e.parent }
func (e *engine) Usage() common.Usage  { e.mu.Lock(); defer e.mu.Unlock(); return e.liveUsage }
func (e *engine) failure(message string) error {
	e.parent.Parent().Logger().Print(message)
	return errors.New(message)
}
func (e *engine) Render(c common.Context) ([]byte, error) {
	cfg := e.parent.Config()
	snapshot := e.parent.Skills().Snapshot()
	if len(c.ActiveSkills) == 0 {
		cfg.SystemPrompt += snapshot.Instructions
	}
	cfg.SystemPrompt += offers(e.parent.Skills().LoadableSkills())
	disclosure := anthropicDisclosure{}
	if cfg.Vendor == common.Anthropic && cfg.Skills != nil && supportsToolReferences(cfg.Model) {
		disclosure.catalog = make([]common.ToolSpec, 0, len(snapshot.Tools))
		disclosure.references = map[string][]string{}
		for _, t := range snapshot.Tools {
			if common.ToolPolicy(cfg, t.Name) != nil || (t.Skill == "web-search" && !cfg.EnableWebSearch) {
				continue
			}
			// Tools retains capability truth for admission and automatic gathering.
			// Automatic tools never enter declarations or enabling-call references.
			// Dormant compiled tools remain deferred; external declarations require
			// an enabled grant, even while their skill's instructions remain pending.
			if t.Ephemeral != "" || (!t.Enabled && (strings.HasPrefix(t.Source, "mcp:") || t.Skill != "")) {
				continue
			}
			disclosure.catalog = append(disclosure.catalog, t)
			if !t.Enabled || t.Origin != "tool" || t.EventSeq == 0 {
				continue
			}
			id := ""
			// Only live projected provenance can supply a native reference.
			// Never hydrate the lifetime archive during ordinary rendering.
			for _, entry := range c.Dialogue {
				if entry.Actor != common.Assistant {
					continue
				}
				for _, part := range entry.Parts {
					if part.Type == "tool_call" && part.Name == "load_skill" && part.AdmissionSeq == t.EventSeq && part.From != nil && part.From.Vendor == cfg.Vendor && part.From.Surface == cfg.Surface {
						id = part.CallID
					}
				}
			}
			if id == "" {
				continue
			}
			present := false
			for _, entry := range c.Dialogue {
				for _, part := range entry.Parts {
					if entry.Seq > t.EventSeq && part.Type == "tool_result" && part.CallID == id && !part.IsError {
						present = true
					}
				}
			}
			if !present {
				continue
			}

			disclosure.references[id] = append(disclosure.references[id], t.Name)
		}
	}
	return renderWithFiles(c, cfg, snapshot.Definitions, e.parent.Sandbox(), disclosure)
}

// EffectiveSystemPrompt uses the same system composition as all renderers.
func (e *engine) EffectiveSystemPrompt(c common.Context) (string, error) {
	cfg := e.parent.Config()
	snapshot := e.parent.Skills().Snapshot()
	if len(c.ActiveSkills) == 0 {
		cfg.SystemPrompt += snapshot.Instructions
	}
	cfg.SystemPrompt += offers(e.parent.Skills().LoadableSkills())
	c, cfg, _ = projectPrompt(c, cfg, snapshot.Definitions)
	return system(c, cfg)
}

// Parse owns transport framing only. The caller owns resp.Body and callbacks;
// no retained History, observation subscription, or reply channel lives here.
func (e *engine) Parse(resp *http.Response, callbacks common.ParseCallbacks) ([]common.Event, error) {
	return e.parseResponse(resp, callbacks, e.parent.Config())
}
func (e *engine) parseResponse(resp *http.Response, callbacks common.ParseCallbacks, cfg common.Config) (events []common.Event, err error) {
	defer func() {
		if callbacks.OnEvent != nil {
			for _, event := range events {
				callbacks.OnEvent(event)
			}
		}
	}()
	if resp == nil || resp.Body == nil {
		return parseFailure("missing response body")
	}
	if resp.StatusCode != http.StatusOK {
		return parseFailure(fmt.Sprintf("HTTP status %d", resp.StatusCode))
	}
	body := io.Reader(resp.Body)
	stream := strings.EqualFold(strings.TrimSpace(strings.Split(resp.Header.Get("Content-Type"), ";")[0]), "text/event-stream")
	if cfg.Surface == common.Responses && !stream {
		// Some Responses peers omit or mislabel the media type. Detect framing
		// from a small prefix, before any whole-body read can delay deltas.
		reader := bufio.NewReader(body)
		stream = sniffResponsesSSE(reader)
		body = reader
	}
	if stream {
		if cfg.Surface == common.Responses {
			events, err = parseResponsesSSE(body, cfg, callbacks)
		} else {
			events, err = parseSSE(body, cfg, callbacks)
		}
	} else {
		var raw []byte
		raw, err = io.ReadAll(body)
		if err != nil {
			return parseFailure("reading response failed")
		}
		if callbacks.OnFrame != nil && cfg.Surface != common.Responses {
			callbacks.OnFrame(common.Frame{Data: string(raw)})
		}
		events, err = parse(resp.StatusCode, raw, cfg)
		for _, event := range events {
			if event.Type == common.ResponseEnded {
				for _, p := range event.Response.Parts {
					plainDelta(p, callbacks)
				}
			}
		}
	}
	return events, err
}
func errorEvent(message string) common.Event {
	return common.Event{Type: common.ErrorOccurred, Error: &common.ErrorData{Message: message}}
}

// Exchange preserves caller-owned header/body cancellation and refuses redirects.
// Only locally chosen diagnostics escape: bodies and transport errors may hold keys.
func (e *engine) Exchange(ctx context.Context, c common.Context) ([]common.Event, error) {
	b, err := e.Render(c)
	if err != nil {
		return []common.Event{errorEvent(err.Error())}, err
	}
	return e.ExchangeRequest(ctx, b)
}

// ExchangeRequest sends the immutable bytes frozen by the actor before its
// request_sent marker. The Engine performs exactly one exchange.
func (e *engine) ExchangeRequest(ctx context.Context, b []byte, callbacks ...common.ParseCallbacks) ([]common.Event, error) {
	var err error
	config := e.parent.Config()
	credential := common.Credential{Mode: common.AuthAPI, Token: config.APIKey}
	if config.Vendor == common.OpenAI {
		credential, err = e.parent.Parent().Auth().Resolve(ctx, config)
		if err != nil {
			return []common.Event{errorEvent(err.Error())}, e.failure(err.Error())
		}
	}
	if strings.TrimSpace(credential.Token) == "" {
		err = e.failure("API key is not set")
		return []common.Event{errorEvent(err.Error())}, err
	}
	secrets := append(common.CredentialValues(config), credential.Token)
	transportTokens := []string{credential.Token}
	endpoint := config.BaseURL
	switch config.Vendor {
	case common.Anthropic:
		endpoint += "/v1/messages"
	case common.OpenAI:
		if config.Surface == common.Responses {
			endpoint = responsesEndpoint(endpoint, credential.Mode)
		} else if strings.HasSuffix(endpoint, "/v1") {
			endpoint += "/chat/completions"
		} else {
			endpoint += "/v1/chat/completions"
		}
	case common.Gemini:
		if !strings.HasSuffix(endpoint, "/v1beta") {
			endpoint += "/v1beta"
		}
		endpoint += "/models/" + url.PathEscape(config.Model)
		if config.Streaming() {
			endpoint += ":streamGenerateContent?alt=sse"
		} else {
			endpoint += ":generateContent"
		}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(b))
	if err != nil {
		return []common.Event{errorEvent("creating request failed")}, e.failure("creating request failed")
	}
	req.Header.Set("content-type", "application/json")
	if config.Streaming() {
		req.Header.Set("Accept", "text/event-stream")
	}
	switch config.Vendor {
	case common.Anthropic:
		req.Header.Set("x-api-key", config.APIKey)
		req.Header.Set("anthropic-version", "2023-06-01")
		if inlineFeatures(config) {
			req.Header.Set("anthropic-beta", "inline-tools-2026-09-15")
		}
	case common.OpenAI:
		req.Header.Set("Authorization", "Bearer "+credential.Token)

	case common.Gemini:
		req.Header.Set("x-goog-api-key", config.APIKey)
	}
	diagnosticBody := b
	if config.Surface == common.Responses {
		diagnosticBody = responsesDiagnostic(b)
	}
	e.parent.Parent().Trace(common.TraceRecord{Stage: "request_body", Payload: string(diagnosticBody)})
	var prediction common.CachePrediction
	if e.cache != nil {
		prediction = e.cache.Observe(config, diagnosticBody)
	}
	if config.Vendor == common.OpenAI {
		e.parent.Parent().Publish(e.parent, common.BillingSelected{Mode: credential.Mode})
		e.parent.Parent().Logger().Printf("OpenAI request funding: %s", credential.Mode)
	}
	resp, err := e.client.Do(req)
	if err != nil {
		err = e.failure("HTTP request failed")
		if ctx.Err() != nil {
			err = ctx.Err()
		}
		return []common.Event{errorEvent(err.Error())}, err
	}
	if credential.Mode == common.AuthChatGPT && resp.StatusCode == http.StatusUnauthorized {
		_ = resp.Body.Close()
		if err = e.parent.Parent().Auth().Reject(ctx, credential); err == nil {
			credential, err = e.parent.Parent().Auth().Resolve(ctx, config)
		}
		if err != nil {
			message := credentialReplacer(credentialSpellings(transportTokens)).Replace(err.Error())
			return []common.Event{errorEvent(message)}, e.failure(message)
		}
		secrets = append(secrets, credential.Token)
		transportTokens = append(transportTokens, credential.Token)
		req.Body = io.NopCloser(bytes.NewReader(b))
		req.Header.Set("Authorization", "Bearer "+credential.Token)
		resp, err = e.client.Do(req)
		if err != nil {
			return []common.Event{errorEvent("HTTP renewal retry failed")}, e.failure("HTTP renewal retry failed")
		}
	}
	defer resp.Body.Close()
	if credential.Mode == common.AuthChatGPT && (resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden) {
		err = e.failure("ChatGPT request rejected; API fallback refused")
		return []common.Event{errorEvent(err.Error())}, err
	}
	if resp.StatusCode != http.StatusOK {
		err = e.failure(fmt.Sprintf("HTTP status %d", resp.StatusCode))
		return []common.Event{errorEvent(err.Error())}, err
	}
	cb := common.ParseCallbacks{}
	if len(callbacks) > 0 {
		cb = callbacks[0]
	}
	originalDelta := cb.OnDelta
	cb.OnDelta = func(d common.PartDelta) {
		e.parent.Parent().Trace(common.TraceRecord{Stage: "parse_delta", PartID: d.PartID, Kind: string(d.Kind), Bytes: len(d.Chunk)})
		if originalDelta != nil {
			originalDelta(d)
		}
	}
	originalFrame := cb.OnFrame
	cb.OnFrame = func(f common.Frame) {
		f.Data = string(common.RedactDiagnostic([]byte(f.Data), secrets...))
		e.parent.Parent().Trace(common.TraceRecord{Stage: "response_frame", Event: f.Event, Payload: f.Data})
		if originalFrame != nil {
			originalFrame(f)
		}
	}
	e.parent.Parent().Trace(common.TraceRecord{Stage: "response_parse"})
	// Final parser callbacks precede its return. Publish only after removing
	// exact transport echoes from the same events returned to the Agent.
	originalEvent := cb.OnEvent
	cb.OnEvent = nil
	deltas := newCredentialDeltas(transportTokens, cb.OnDelta)
	cb.OnDelta = deltas.push
	events, err := e.parseResponse(resp, cb, config)
	deltas.flush()
	// Error diagnostics retain their existing additional known values, but
	// use the same normalized one-pass policy before callbacks, never a second
	// scrub of already-redacted event text.
	errorReplacer := credentialReplacer(credentialSpellings(secrets))
	redactCredentialEvents(events, deltas.replacer, errorReplacer)
	if originalEvent != nil {
		for _, event := range events {
			originalEvent(event)
		}
	}
	if err != nil {
		err = errors.New(errorReplacer.Replace(err.Error()))
	}
	if ctx.Err() != nil {
		err = ctx.Err()
	}

	// The same measured response facts feed two distinct owners: this process's
	// HTTP bill here, and replay totals when Agent appends them to History.
	accounted := false
	for _, event := range events {
		if event.Type == common.ResponseEnded && event.Response != nil {
			e.account(config, event.Response)
			if e.cache != nil {
				e.cache.Report(prediction, event.Response.Usage)
			}
			accounted = true
		}
	}
	if !accounted {
		// A successful HTTP response with unusable framing may still be billed.
		e.account(config, &common.ResponseData{Usage: common.UnknownUsage()})
		if e.cache != nil {
			e.cache.Report(prediction, common.UnknownUsage())
		}
	}
	if err != nil {
		e.parent.Parent().Logger().Print(err.Error())
	}
	return events, err
}

// Send preserves the accepted text API as a disposable view, not stored state.
func (e *engine) Send(ctx context.Context, messages []common.Message) (string, error) {
	c := common.Context{Turn: common.InputPending}
	for i, m := range messages {
		actor := common.Human
		if m.Role == "assistant" {
			actor = common.Assistant
		}
		c.Dialogue = append(c.Dialogue, common.Entry{Seq: common.Seq(i + 1), Actor: actor, Parts: common.PartList{common.Text(m.Content)}})
	}
	events, err := e.Exchange(ctx, c)
	return responseText(events), err
}
func responseText(events []common.Event) string {
	var b strings.Builder
	for _, e := range events {
		if e.Type == common.ResponseEnded {
			for _, p := range e.Response.Parts {
				if p.Type == "text" {
					b.WriteString(p.Text)
				}
			}
		}
	}
	return b.String()
}

// Canonical JSON preserves integers beyond float64's exact range.
func canonical(raw json.RawMessage) (json.RawMessage, error) {
	if len(raw) == 0 {
		return json.RawMessage(`{}`), nil
	}
	var v any
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	if err := d.Decode(&v); err != nil {
		return nil, err
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return nil, errors.New("invalid JSON value")
	}
	return json.Marshal(v)
}

// Plain delivery uses exactly the same notification vocabulary, in one piece.
func plainDelta(p common.Part, cb common.ParseCallbacks) {
	if cb.OnDelta == nil {
		return
	}
	d := common.PartDelta{PartID: p.ID}
	switch p.Type {
	case "text":
		d.Kind = common.DeltaText
		d.Chunk = p.Text
	case "reasoning_summary":
		d.Kind = common.DeltaReasoningSummary
		d.Chunk = p.Text
	case "tool_call":
		d.Kind = common.DeltaToolCall
		d.Name = p.Name
		d.Chunk = p.Name + string(p.Args)
	case "opaque":
		var fields map[string]json.RawMessage
		if json.Unmarshal(p.Data, &fields) != nil {
			return
		}
		switch {
		case string(fields["type"]) == `"thinking"`:
			d.Chunk = str(fields["thinking"])
		case string(fields["thought"]) == "true":
			d.Chunk = str(fields["text"])
		case fields["reasoning_content"] != nil:
			d.Chunk = str(fields["reasoning_content"])
		case fields["reasoning"] != nil:
			d.Chunk = str(fields["reasoning"])
		default:
			return
		}
		d.Kind = common.DeltaThinking
	default:
		return
	}
	cb.OnDelta(d)
}
