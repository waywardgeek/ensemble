package llm

// Anthropic — Messages API.
//
// Wire format verified against platform.claude.com on 2026-09-12. Wire formats
// drift; this file is the place that has to know, and it is the only place.

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/waywardgeek/coding-agents-course/agent/internal/common"
)

type anthropicSeam struct{}

// --- request -----------------------------------------------------------

type anthRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	System    string    `json:"system,omitempty"` // TOP-LEVEL, not a message
	Messages  []anthMsg `json:"messages"`
	// Tools is omitted, not empty, when nothing is declared: a chapter 2
	// request and a chapter 3 request with an empty registry are the same bytes.
	Tools []anthTool `json:"tools,omitempty"`

	// Stream asks for Server-Sent Events. Omitted when false so that a
	// non-streaming request is byte-identical to one written before
	// streaming existed.
	Stream bool `json:"stream,omitempty"`

	// Thinking requests extended reasoning from the model. Omitted when
	// thinking is off so that a non-thinking request is byte-identical to
	// one written before thinking existed.
	//
	// Two wire formats exist:
	//   Manual (older models):   { "type": "enabled", "budget_tokens": N }
	//   Adaptive (claude-5 era): { "type": "adaptive" }
	// Adaptive models control depth via the separate OutputConfig field.
	Thinking *anthThinking `json:"thinking,omitempty"`

	// OutputConfig controls thinking effort for adaptive-thinking models.
	// Omitted for manual-thinking models, where budget_tokens controls depth.
	OutputConfig *anthOutputConfig `json:"output_config,omitempty"`

	// Temperature must NOT be set when thinking is enabled — Anthropic
	// rejects temperature != 1 with thinking. We never set it anywhere,
	// so omitempty keeps it absent.
	Temperature *float64 `json:"temperature,omitempty"`
}

// anthThinking is Anthropic's thinking configuration block.
//
// Manual (pre-adaptive) models:
//
//	{ "type": "enabled", "budget_tokens": N }
//	budget_tokens must be >= 1024, and max_tokens must strictly exceed it.
//
// Adaptive models (claude-opus-5, claude-sonnet-5):
//
//	{ "type": "adaptive", "display": "summarized" }
//	Thinking depth is controlled by the separate output_config.effort field.
//
// DISPLAY IS NOT OPTIONAL, and omitting it is the expensive kind of mistake.
// On adaptive models `display` defaults to "omitted": the response still
// contains a thinking block, the block is still SIGNED, the thinking tokens
// are still GENERATED AND BILLED, and the text is empty. The symptom is
// indistinguishable from thinking being switched off, except on the invoice.
// "summarized" asks for the text to actually be sent back.
type anthThinking struct {
	Type         string `json:"type"`
	Display      string `json:"display,omitempty"`
	BudgetTokens int    `json:"budget_tokens,omitempty"`
}

// anthOutputConfig carries Anthropic's output_config block for adaptive
// thinking models.
//
//	{ "effort": "high" | "medium" | "low" }
type anthOutputConfig struct {
	Effort string `json:"effort"`
}

// anthTool is Anthropic's declaration shape. The schema key is `input_schema`
// — the one vendor of the three that does not call it `parameters`.
type anthTool struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	InputSchema json.RawMessage `json:"input_schema"`
}

func anthTools(decls []common.ToolDecl) []anthTool {
	var out []anthTool
	for _, d := range decls {
		out = append(out, anthTool{Name: d.Name, Description: d.Description, InputSchema: d.Schema})
	}
	return out
}

type anthMsg struct {
	Role    string      `json:"role"`
	Content []anthBlock `json:"content"`
}

// anthBlock marshals either a structured block or, for opaque replay material,
// the vendor's own bytes verbatim. Opaque material is carried, never
// interpreted — we could not reconstruct a thinking signature if we tried.
type anthBlock struct {
	Raw json.RawMessage

	Type      string
	Text      string
	ID        string
	Name      string
	Input     json.RawMessage
	ToolUseID string
	Content   string
	IsError   bool
}

func (b anthBlock) MarshalJSON() ([]byte, error) {
	if len(b.Raw) > 0 {
		return b.Raw, nil
	}
	// A struct with ordered fields, never a map: Go randomizes map iteration
	// order, so a map here serializes differently on some future run, on some
	// future machine, and never on the one where you tested it.
	type wire struct {
		Type      string          `json:"type"`
		Text      string          `json:"text,omitempty"`
		ID        string          `json:"id,omitempty"`
		Name      string          `json:"name,omitempty"`
		Input     json.RawMessage `json:"input,omitempty"`
		ToolUseID string          `json:"tool_use_id,omitempty"`
		Content   string          `json:"content,omitempty"`
		IsError   bool            `json:"is_error,omitempty"`
	}
	return json.Marshal(wire{b.Type, b.Text, b.ID, b.Name, b.Input, b.ToolUseID, b.Content, b.IsError})
}

func (anthropicSeam) Render(c *common.Context, cfg common.Config) (*http.Request, error) {
	target := common.Provenance{Vendor: common.VendorAnthropic, Model: cfg.Model, Surface: common.SurfaceMessages}

	var msgs []anthMsg
	// appendBlocks merges into the previous message when the role matches.
	//
	// Client-side merging is REQUIRED here, but not for the reason most people
	// give. The API does not reject consecutive same-role turns — it combines
	// them server-side. The real constraint is narrower and sharper: a
	// tool_result block must immediately follow the tool_use that produced it,
	// and within the user message the tool_result blocks must come FIRST, with
	// any text after them. Text first is a 400.
	appendBlocks := func(role string, blocks []anthBlock, resultsFirst bool) {
		if len(blocks) == 0 {
			return
		}
		if n := len(msgs); n > 0 && msgs[n-1].Role == role {
			if resultsFirst {
				// Splice tool results ahead of existing text in this message.
				at := 0
				for at < len(msgs[n-1].Content) && msgs[n-1].Content[at].Type == "tool_result" {
					at++
				}
				rest := append([]anthBlock{}, msgs[n-1].Content[at:]...)
				msgs[n-1].Content = append(append(msgs[n-1].Content[:at:at], blocks...), rest...)
				return
			}
			msgs[n-1].Content = append(msgs[n-1].Content, blocks...)
			return
		}
		msgs = append(msgs, anthMsg{Role: role, Content: blocks})
	}

	for _, entry := range c.Dialogue {
		r, err := classify(entry, cfg)
		if err != nil {
			return nil, err
		}
		// Anthropic references an uploaded file as a `source` of type "file"
		// carrying a `file_id`, on an `image` or `document` block. That form is
		// real, but mapping a common.Ref onto it needs a rule for which locators are
		// Anthropic file ids and which are something else, and Chapter 2 does
		// not have one. Raising here is the honest answer: a guessed field name
		// is worse than an unimplemented one, and dropping the blob would send
		// a request that looks fine and is missing its attachment.
		if len(r.Blobs) > 0 {
			return nil, fmt.Errorf("anthropic: rendering a blob part is not implemented in this "+
				"chapter (%s at %s)", r.Blobs[0].MIME, r.Blobs[0].Ref.Locator)
		}
		switch entry.Actor {
		case common.ActorTool:
			var blocks []anthBlock
			for _, res := range r.Tools {
				blocks = append(blocks, anthBlock{
					Type:      "tool_result",
					ToolUseID: res.CallID,
					Content:   resultText(res),
					IsError:   res.IsError,
				})
			}
			appendBlocks("user", blocks, true)

		case common.ActorAgent:
			var blocks []anthBlock
			// Opaque replay material goes back only to the EXACT model that
			// issued it. common.Vendor is not a fine enough grain.
			for _, op := range r.Raw {
				if op.From.SameModel(target) {
					blocks = append(blocks, anthBlock{Raw: op.Data})
				}
			}
			for _, t := range r.Texts {
				blocks = append(blocks, anthBlock{Type: "text", Text: t})
			}
			for i, call := range r.Calls {
				blocks = append(blocks, anthBlock{
					Type:  "tool_use",
					ID:    callIDFor(call, "toolu", entry.Seq, i),
					Name:  call.Name,
					Input: jsonObject(call.Args),
				})
			}
			appendBlocks("assistant", blocks, false)

		default: // common.ActorHuman
			var blocks []anthBlock
			for _, t := range r.Texts {
				blocks = append(blocks, anthBlock{Type: "text", Text: t})
			}
			appendBlocks("user", blocks, false)
		}
	}

	// Ephemera go LAST, in exactly one copy, and are never part of the
	// dialogue. Volatile content at the front of a prefix converts the
	// cheapest token category into the most expensive one on every request.
	if len(c.Ephemera) > 0 {
		var blocks []anthBlock
		for _, p := range c.Ephemera {
			if t, ok := p.(common.TextPart); ok {
				blocks = append(blocks, anthBlock{Type: "text", Text: t.Text})
			}
		}
		appendBlocks("user", blocks, false)
	}

	// Resolve thinking before building the request body: the budget
	// determines whether the thinking block is included, and it also
	// constrains max_tokens (which must strictly exceed budget_tokens).
	effort, thinkingBudget := common.ThinkingFor(cfg)
	maxTokens := common.EnsureMaxTokens(cfg.MaxTokens, thinkingBudget)

	// Determine whether this model uses adaptive or manual thinking.
	features, _ := common.LookupModel(cfg.Model)

	body := anthRequest{
		Model:     cfg.Model,
		MaxTokens: maxTokens,
		System:    cfg.SystemPrompt, // top-level. Store it and you have picked a vendor.
		Messages:  msgs,
		Tools:     anthTools(cfg.Tools),
		Stream:    common.StreamingFor(cfg) != 0,
	}
	if effort != common.ThinkingOff && thinkingBudget > 0 {
		if features.AdaptiveThinking {
			// Adaptive models: type:"adaptive" + output_config.effort.
			// No budget_tokens — the model manages its own thinking depth.
			// display:"summarized" is required to receive the thinking TEXT;
			// without it the block arrives empty and is still billed.
			body.Thinking = &anthThinking{Type: "adaptive", Display: "summarized"}
			body.OutputConfig = &anthOutputConfig{Effort: effortString(effort)}
		} else {
			// Manual models: type:"enabled" + budget_tokens.
			body.Thinking = &anthThinking{
				Type:         "enabled",
				BudgetTokens: thinkingBudget,
			}
		}
	}
	return newJSONRequest("POST", cfg.BaseURL+"/v1/messages", body, map[string]string{
		"x-api-key":         cfg.APIKey,
		"anthropic-version": "2023-06-01",
	})
}

// --- response ----------------------------------------------------------

type anthResponse struct {
	Model      string            `json:"model"`
	Content    []json.RawMessage `json:"content"`
	StopReason string            `json:"stop_reason"`
	Usage      anthUsage         `json:"usage"`
}

// anthUsage is DISJOINT by documented definition: input_tokens counts only
// tokens after the last cache breakpoint, so the billable input total is
// input + cache_creation + cache_read. Verified 2026-09-12; the docs give the
// formula and a worked example (200,000 read + 0 created + 50 input =
// 200,050). Reading input_tokens alone does not double-count — it undercounts
// by three orders of magnitude on a warm cache.
type anthUsage struct {
	InputTokens         int `json:"input_tokens"`
	OutputTokens        int `json:"output_tokens"`
	CacheCreationTokens int `json:"cache_creation_input_tokens"`
	CacheReadTokens     int `json:"cache_read_input_tokens"`
}

type anthBlockHeader struct {
	Type  string          `json:"type"`
	Text  string          `json:"text"`
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

// Parse handles a streamed response and a whole-document response through one
// method, because they are the same thing at different chunk counts.
func (anthropicSeam) Parse(resp *http.Response, cb common.StreamCallbacks) error {
	if resp.StatusCode != http.StatusOK {
		return parseErrorResponse(resp, cb)
	}
	if isSSE(resp) {
		return anthParseStream(resp, cb)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("anthropic: read body: %w", err)
	}
	cb.Frame("", body)
	parts, from, usage, err := anthAssemble(body)
	if err != nil {
		return err
	}
	emitLengthOneDeltas(parts, cb, anthThinkingText)
	anthEmitResponse(parts, from, usage, cb)
	emitFinals(parts, cb)
	return nil
}

// anthAssemble turns one complete JSON document into parts. Shared with the
// streaming path only in its OUTPUT shape: both must produce parts that
// render back to the same bytes, or replay stops matching.
func anthAssemble(body []byte) (common.PartList, common.Provenance, common.Usage, error) {
	var resp anthResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, common.Provenance{}, common.Usage{}, fmt.Errorf("anthropic: %w", err)
	}
	// common.Provenance comes from the RESPONSE, not from common.Config: the vendor tells
	// you which model actually answered, and it is not always the one you
	// asked for.
	from := common.Provenance{Vendor: common.VendorAnthropic, Model: resp.Model, Surface: common.SurfaceMessages}

	var parts common.PartList
	for _, raw := range resp.Content {
		var h anthBlockHeader
		if err := json.Unmarshal(raw, &h); err != nil {
			return nil, from, common.Usage{}, fmt.Errorf("anthropic: content block: %w", err)
		}
		switch h.Type {
		case "text":
			parts = append(parts, common.TextPart{Text: h.Text})
		case "tool_use":
			parts = append(parts, common.ToolCallPart{
				CallID: h.ID, From: from, Name: h.Name, Args: jsonObject(h.Input),
			})
		default:
			// thinking, redacted_thinking, and anything shipped after this
			// book went to print. Carried verbatim, tagged with the model that
			// produced it, never interpreted.
			parts = append(parts, common.OpaquePart{From: from, Data: raw})
		}
	}
	usage := common.Usage{
		Input:      resp.Usage.InputTokens,
		CacheWrite: resp.Usage.CacheCreationTokens,
		CacheRead:  resp.Usage.CacheReadTokens,
		Output:     resp.Usage.OutputTokens,
	}
	return parts, from, usage, nil
}

func anthEmitResponse(parts common.PartList, from common.Provenance, usage common.Usage, cb common.StreamCallbacks) {
	cb.Emit(common.Event{Type: common.ResponseStarted})
	cb.Emit(common.Event{Type: common.ResponseEnded, Response: &common.ResponseData{
		Parts: parts,
		From:  from,
		Usage: usage,
	}})
}

// effortString converts a ThinkingEffort enum to the wire string Anthropic
// expects in output_config.effort for adaptive-thinking models.
func effortString(e common.ThinkingEffort) string {
	switch e {
	case common.ThinkingLow:
		return "low"
	case common.ThinkingMedium:
		return "medium"
	default:
		return "high"
	}
}

// anthThinkingText pulls the readable text out of an Anthropic reasoning
// block. Returns "" for redacted_thinking, which has no text to show — the
// part is still carried in the event, it just cannot be displayed.
func anthThinkingText(p common.OpaquePart) string {
	var b struct {
		Thinking string `json:"thinking"`
	}
	if json.Unmarshal(p.Data, &b) != nil {
		return ""
	}
	return b.Thinking
}

// --- streaming ---------------------------------------------------------

// anthStreamBlock accumulates one content block as its fragments arrive.
//
// The signature field is the reason this cannot be a simple string builder.
// A thinking block must be handed BACK to the vendor byte-identical on the
// next request, signature included, or the request is refused. Streaming
// therefore has to reassemble the exact JSON the non-streaming path would
// have received, not merely the text a human wants to read.
type anthStreamBlock struct {
	Kind      string
	Text      strings.Builder
	Thinking  strings.Builder
	Signature string
	ToolID    string
	ToolName  string
	Args      strings.Builder
	Raw       json.RawMessage
}

func (b *anthStreamBlock) finalize(from common.Provenance) common.Part {
	switch b.Kind {
	case "text":
		return common.TextPart{Text: b.Text.String()}
	case "tool_use":
		return common.ToolCallPart{
			CallID: b.ToolID,
			From:   from,
			Name:   b.ToolName,
			Args:   jsonObject(json.RawMessage(b.Args.String())),
		}
	case "thinking":
		// Rebuilt, not passed through: the fragments arrived separately and
		// the vendor never sends the assembled block.
		rebuilt, err := json.Marshal(map[string]string{
			"type":      "thinking",
			"thinking":  b.Thinking.String(),
			"signature": b.Signature,
		})
		if err != nil {
			return common.OpaquePart{From: from, Data: b.Raw}
		}
		return common.OpaquePart{From: from, Data: rebuilt}
	default:
		return common.OpaquePart{From: from, Data: b.Raw}
	}
}

type anthStreamEvent struct {
	Type         string          `json:"type"`
	Index        int             `json:"index"`
	Message      *anthResponse   `json:"message"`
	ContentBlock json.RawMessage `json:"content_block"`
	Delta        struct {
		Type        string `json:"type"`
		Text        string `json:"text"`
		Thinking    string `json:"thinking"`
		Signature   string `json:"signature"`
		PartialJSON string `json:"partial_json"`
	} `json:"delta"`
	Usage *anthUsage `json:"usage"`
}

func anthParseStream(resp *http.Response, cb common.StreamCallbacks) error {
	from := common.Provenance{Vendor: common.VendorAnthropic, Surface: common.SurfaceMessages}
	var usage common.Usage
	blocks := map[int]*anthStreamBlock{}
	var order []int
	var perr error

	block := func(i int) *anthStreamBlock {
		if b, ok := blocks[i]; ok {
			return b
		}
		b := &anthStreamBlock{}
		blocks[i] = b
		order = append(order, i)
		return b
	}

	readErr := ReadSSE(resp.Body, func(eventType string, data []byte) {
		// Every frame goes to the API log before anything interprets it, so
		// the log is complete even for the frame that causes a parse failure.
		cb.Frame(eventType, data)
		if perr != nil {
			return
		}

		var ev anthStreamEvent
		if err := json.Unmarshal(data, &ev); err != nil {
			perr = fmt.Errorf("anthropic: stream frame: %w", err)
			return
		}

		switch ev.Type {
		case "message_start":
			if ev.Message != nil {
				from.Model = ev.Message.Model
				usage.Input = ev.Message.Usage.InputTokens
				usage.CacheWrite = ev.Message.Usage.CacheCreationTokens
				usage.CacheRead = ev.Message.Usage.CacheReadTokens
				usage.Output = ev.Message.Usage.OutputTokens
			}
			cb.Emit(common.Event{Type: common.ResponseStarted})

		case "content_block_start":
			b := block(ev.Index)
			b.Raw = append(json.RawMessage(nil), ev.ContentBlock...)
			var h anthBlockHeader
			if err := json.Unmarshal(ev.ContentBlock, &h); err != nil {
				perr = fmt.Errorf("anthropic: content_block_start: %w", err)
				return
			}
			b.Kind = h.Type
			switch h.Type {
			case "text":
				b.Text.WriteString(h.Text)
				if h.Text != "" {
					cb.Delta(partID(ev.Index), common.DeltaText, h.Text)
				}
			case "tool_use":
				b.ToolID, b.ToolName = h.ID, h.Name
				// The name is the first chunk of a tool call, which is why
				// the reassembled display string is name-then-arguments.
				cb.Delta(partID(ev.Index), common.DeltaToolCall, h.Name)
			}

		case "content_block_delta":
			b := block(ev.Index)
			switch ev.Delta.Type {
			case "text_delta":
				b.Text.WriteString(ev.Delta.Text)
				cb.Delta(partID(ev.Index), common.DeltaText, ev.Delta.Text)
			case "thinking_delta":
				b.Thinking.WriteString(ev.Delta.Thinking)
				cb.Delta(partID(ev.Index), common.DeltaThinking, ev.Delta.Thinking)
			case "signature_delta":
				// Accumulated, never displayed. It is a credential for the
				// next request, not content.
				b.Signature += ev.Delta.Signature
			case "input_json_delta":
				b.Args.WriteString(ev.Delta.PartialJSON)
				// Display only. This JSON is INCOMPLETE until the block
				// stops, and acting on it is how a tool runs with half its
				// arguments.
				cb.Delta(partID(ev.Index), common.DeltaToolCall, ev.Delta.PartialJSON)
			}

		case "message_delta":
			// Output tokens are only final here: message_start reports zero.
			if ev.Usage != nil && ev.Usage.OutputTokens > 0 {
				usage.Output = ev.Usage.OutputTokens
			}
		}
	})

	if perr != nil {
		return perr
	}
	if readErr != nil {
		return fmt.Errorf("anthropic: stream: %w", readErr)
	}

	// Index order, not arrival order. They agree today, and relying on that
	// is the sort of assumption that breaks quietly.
	sort.Ints(order)
	var parts common.PartList
	for _, i := range order {
		parts = append(parts, blocks[i].finalize(from))
	}

	// The finalized event still lands in the event log. The deltas were for
	// watching; THIS is the authority, and it is what a reconnecting client
	// re-renders from.
	cb.Emit(common.Event{Type: common.ResponseEnded, Response: &common.ResponseData{
		Parts: parts,
		From:  from,
		Usage: usage,
	}})

	// Finals carry the VENDOR's block id, which is the id the deltas carried.
	// Not the position in the finished list: those agree here only because
	// Anthropic numbers its blocks contiguously from zero, and depending on
	// that would be depending on a coincidence.
	for n, i := range order {
		cb.Final(partID(i), parts[n])
	}
	return nil
}

// vendorErrorMessage digs a human-readable message out of whatever error shape
// a vendor used, falling back to the raw body. An HTTP 429 is an
// common.ErrorOccurred, not a response.
func vendorErrorMessage(body []byte) string {
	var e struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &e) == nil && e.Error.Message != "" {
		return e.Error.Message
	}
	if len(body) > 300 {
		return string(body[:300])
	}
	return string(body)
}
