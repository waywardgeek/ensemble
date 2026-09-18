package llm

// OpenAI — Chat Completions.
//
// Wire format verified against platform.openai.com on 2026-09-12.
//
// This is the second renderer, and the chapter's prediction says it should
// cost real work. It does: a tool role of its own, a flat message list,
// arguments as a JSON-encoded STRING rather than an object, and a usage
// convention that is the opposite of Anthropic's.

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

type openAISeam struct{}

// --- request -----------------------------------------------------------

type oaiRequest struct {
	Model     string   `json:"model"`
	MaxTokens int      `json:"max_completion_tokens,omitempty"`
	Messages  []oaiMsg `json:"messages"`
	// Omitted when nothing is declared — see anthRequest.Tools.
	Tools []oaiTool `json:"tools,omitempty"`

	Stream bool `json:"stream,omitempty"`

	// StreamOptions is how you get the token counts back.
	//
	// A streamed Chat Completions response omits `usage` ENTIRELY unless
	// include_usage is set. Nothing fails if you forget: every turn simply
	// reports zero tokens, the running total stays at zero, and the bug looks
	// like a display problem rather than a missing request field. This is the
	// exact failure ch2 built four disjoint usage categories to avoid, and it
	// reappears here because the streaming surface has its own opinion about
	// what is optional.
	StreamOptions *oaiStreamOptions `json:"stream_options,omitempty"`

	// ReasoningEffort controls how much reasoning the model does.
	// Accepted values: "low", "medium", "high". Omitted when thinking
	// is off or the model doesn't support it.
	//
	// Chat Completions does NOT return reasoning content, only a summary.
	// We do not fabricate thinking deltas to make it look symmetric —
	// represent the real capability honestly, the way Gemini's clear
	// StreamToolArgs bit does.
	ReasoningEffort string `json:"reasoning_effort,omitempty"`
}

type oaiStreamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

// oaiTool is OpenAI's declaration shape: one level of wrapping more than the
// others, because `tools` is a union and `function` is the arm we want.
type oaiTool struct {
	Type     string      `json:"type"` // always "function"
	Function oaiFunction `json:"function"`
}

type oaiFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Parameters  json.RawMessage `json:"parameters"`
}

func oaiTools(decls []common.ToolDecl) []oaiTool {
	var out []oaiTool
	for _, d := range decls {
		out = append(out, oaiTool{Type: "function", Function: oaiFunction{
			Name: d.Name, Description: d.Description, Parameters: d.Schema,
		}})
	}
	return out
}

type oaiMsg struct {
	Role string `json:"role"`
	// Content is a pointer because an assistant message carrying tool calls
	// must send content as null, not as "". Empty string and absent are
	// different values here, which is exactly the kind of vendor detail that
	// has no business in a context type.
	Content    *string       `json:"content"`
	ToolCalls  []oaiToolCall `json:"tool_calls,omitempty"`
	ToolCallID string        `json:"tool_call_id,omitempty"`
}

type oaiToolCall struct {
	ID       string  `json:"id"`
	Type     string  `json:"type"`
	Function oaiFunc `json:"function"`
}

type oaiFunc struct {
	Name string `json:"name"`
	// A JSON-ENCODED STRING, not an object. Anthropic sends the same
	// information as a nested object and Gemini as `args`. One fact, three
	// encodings — and a context that stored any of them would have picked a
	// vendor.
	Arguments string `json:"arguments"`
}

func strptr(s string) *string { return &s }

func (openAISeam) Render(c *common.Context, cfg common.Config) (*http.Request, error) {
	target := common.Provenance{Vendor: common.VendorOpenAI, Model: cfg.Model, Surface: common.SurfaceChatCompletions}

	msgs := []oaiMsg{}
	if cfg.SystemPrompt != "" {
		// A message INSIDE the array — the same fact Anthropic puts in a
		// top-level parameter and Gemini hoists into its own object. Newer
		// models prefer the "developer" role; "system" remains accepted
		// everywhere and is the portable choice.
		msgs = append(msgs, oaiMsg{Role: "system", Content: strptr(cfg.SystemPrompt)})
	}

	for _, entry := range c.Dialogue {
		r, err := classify(entry, cfg)
		if err != nil {
			return nil, err
		}
		// Same reasoning as the Anthropic renderer: OpenAI's file-reference
		// shape was not verified for this chapter, and a guessed field name is
		// worse than an unimplemented one.
		if len(r.Blobs) > 0 {
			return nil, fmt.Errorf("openai: rendering a blob part is not implemented in this "+
				"chapter (%s at %s)", r.Blobs[0].MIME, r.Blobs[0].Ref.Locator)
		}
		switch entry.Actor {
		case common.ActorTool:
			// Its own message, with a role that exists for exactly this.
			// No merging: OpenAI imposes no alternation requirement, so the
			// two honest, separate facts stay separate. The merge Anthropic
			// forces is a fact about a wire format, not about the
			// conversation.
			for _, res := range r.Tools {
				msgs = append(msgs, oaiMsg{
					Role:       "tool",
					ToolCallID: res.CallID,
					Content:    strptr(resultText(res)),
				})
			}

		case common.ActorAgent:
			m := oaiMsg{Role: "assistant"}
			if text := joinTexts(r.Texts); text != "" {
				m.Content = strptr(text)
			}
			for i, call := range r.Calls {
				m.ToolCalls = append(m.ToolCalls, oaiToolCall{
					ID:   callIDFor(call, "call", entry.Seq, i),
					Type: "function",
					Function: oaiFunc{
						Name:      call.Name,
						Arguments: string(jsonObject(call.Args)),
					},
				})
			}
			// Opaque replay material is deliberately NOT sent on this surface.
			// Chat Completions has nowhere to put it: encrypted reasoning is a
			// Responses-API concept. Dropping it here is a considered decision
			// recorded in one place, not an accident — and the material is
			// still in the log, still tagged with the model that issued it,
			// for a renderer that can use it.

			// An assistant turn with neither text nor tool calls still has to
			// be renderable, because we will replay it as history on the next
			// round. content:null is accepted ONLY alongside tool_calls; on a
			// bare assistant message OpenAI rejects it outright:
			//
			//   Invalid value for 'content': expected a string, got null.
			//
			// That shape is not hypothetical. A reasoning model that spends
			// its entire max_completion_tokens budget on reasoning returns
			// content:"" with finish_reason:"length" — a legal string, which
			// our parser discards as "no parts", and which this renderer would
			// then send back as null. The vendor never emitted null; we did.
			// Send back the empty string the vendor actually gave us.
			if m.Content == nil && len(m.ToolCalls) == 0 {
				m.Content = strptr("")
			}
			_ = target
			msgs = append(msgs, m)

		default: // common.ActorHuman
			msgs = append(msgs, oaiMsg{Role: "user", Content: strptr(joinTexts(r.Texts))})
		}
	}

	if len(c.Ephemera) > 0 {
		if text := ephemeraText(c.Ephemera); text != "" {
			msgs = append(msgs, oaiMsg{Role: "user", Content: strptr(text)})
		}
	}

	// Resolve thinking for OpenAI: reasoning_effort is a string, not a
	// budget. Chat Completions does not return reasoning content, so we
	// only set the effort level — no budget needed.
	effort, _ := common.ThinkingFor(cfg)
	var reasoningEffort string
	if effort != common.ThinkingOff {
		reasoningEffort = effort.String()
	}

	body := oaiRequest{
		Model:           cfg.Model,
		MaxTokens:       cfg.MaxTokens,
		Messages:        msgs,
		Tools:           oaiTools(cfg.Tools),
		Stream:          common.StreamingFor(cfg) != 0,
		ReasoningEffort: reasoningEffort,
	}
	if body.Stream {
		body.StreamOptions = &oaiStreamOptions{IncludeUsage: true}
	}

	return newJSONRequest("POST", cfg.BaseURL+"/v1/chat/completions",
		body,
		map[string]string{"Authorization": "Bearer " + cfg.APIKey})
}

func joinTexts(texts []string) string {
	out := ""
	for i, t := range texts {
		if i > 0 {
			out += "\n"
		}
		out += t
	}
	return out
}

func ephemeraText(parts common.PartList) string {
	var texts []string
	for _, p := range parts {
		if t, ok := p.(common.TextPart); ok {
			texts = append(texts, t.Text)
		}
	}
	return joinTexts(texts)
}

// --- response ----------------------------------------------------------

type oaiResponse struct {
	Model   string `json:"model"`
	Choices []struct {
		Message struct {
			Content   *string       `json:"content"`
			ToolCalls []oaiToolCall `json:"tool_calls"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage oaiUsage `json:"usage"`
}

// oaiUsage is SUBSET, the opposite of Anthropic's convention, and this is the
// purest seam bug in the chapter: nothing crashes, no test fails, the number
// is simply not the number.
//
// prompt_tokens is the GRAND TOTAL of input. cached_tokens and
// cache_write_tokens are two disjoint sub-buckets INSIDE it. Verified
// 2026-09-12 with live calls: an identical request repeated reported
// prompt_tokens 5616 on both the miss and the hit — unchanged — with
// cache_write 5613 on the first and cached 5613 on the second. Under a
// disjoint convention the second call would have reported 3.
//
// The detail fields are OPTIONAL on Chat Completions and older models omit
// them rather than zeroing them, so they must default to 0.
type oaiUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	Details          struct {
		CachedTokens     int `json:"cached_tokens"`
		CacheWriteTokens int `json:"cache_write_tokens"`
	} `json:"prompt_tokens_details"`
}

func (openAISeam) Parse(resp *http.Response, cb common.StreamCallbacks) error {
	if resp.StatusCode != http.StatusOK {
		return parseErrorResponse(resp, cb)
	}
	if isSSE(resp) {
		return oaiParseStream(resp, cb)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("openai: read body: %w", err)
	}
	cb.Frame("", body)
	parts, from, usage, err := oaiAssemble(body)
	if err != nil {
		return err
	}
	// No thinking extractor: Chat Completions has no reasoning block to show.
	emitLengthOneDeltas(parts, cb, nil)
	cb.Emit(common.Event{Type: common.ResponseStarted})
	cb.Emit(common.Event{Type: common.ResponseEnded, Response: &common.ResponseData{
		Parts: parts, From: from, Usage: usage,
	}})
	emitFinals(parts, cb)
	return nil
}

func oaiAssemble(body []byte) (common.PartList, common.Provenance, common.Usage, error) {
	var resp oaiResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, common.Provenance{}, common.Usage{}, fmt.Errorf("openai: %w", err)
	}
	if len(resp.Choices) == 0 {
		return nil, common.Provenance{}, common.Usage{}, fmt.Errorf("openai: response had no choices")
	}
	from := common.Provenance{Vendor: common.VendorOpenAI, Model: resp.Model, Surface: common.SurfaceChatCompletions}

	var parts common.PartList
	choice := resp.Choices[0]
	if choice.Message.Content != nil && *choice.Message.Content != "" {
		parts = append(parts, common.TextPart{Text: *choice.Message.Content})
	}
	for _, tc := range choice.Message.ToolCalls {
		// arguments arrives as a JSON-encoded string; the context stores the
		// decoded object, because "a string that happens to contain JSON" is
		// OpenAI's encoding decision, not a fact about the conversation.
		parts = append(parts, common.ToolCallPart{
			CallID: tc.ID, From: from, Name: tc.Function.Name,
			Args: jsonObject(json.RawMessage(tc.Function.Arguments)),
		})
	}
	return parts, from, oaiCanonicalUsage(resp.Usage), nil
}

// oaiCanonicalUsage converts OpenAI's SUBSET convention to the disjoint form.
func oaiCanonicalUsage(u oaiUsage) common.Usage {
	uncached := u.PromptTokens - u.Details.CachedTokens - u.Details.CacheWriteTokens
	if uncached < 0 {
		uncached = 0
	}
	return common.Usage{
		Input:      uncached,
		CacheWrite: u.Details.CacheWriteTokens,
		CacheRead:  u.Details.CachedTokens,
		Output:     u.CompletionTokens,
	}
}

// --- streaming ---------------------------------------------------------

// oaiStreamCall accumulates one tool call across chunks. OpenAI identifies it
// by an `index` that is scoped to the call list, not to the content blocks —
// a different numbering from Anthropic's, for the same job.
type oaiStreamCall struct {
	PartID uint64
	ID     string
	Name   string
	Args   strings.Builder
}

type oaiStreamChunk struct {
	Model   string `json:"model"`
	Choices []struct {
		Delta struct {
			Content   *string `json:"content"`
			ToolCalls []struct {
				Index    int    `json:"index"`
				ID       string `json:"id"`
				Function struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`

	// Usage is a POINTER and arrives on its own final chunk, whose `choices`
	// is empty. See oaiRequest.StreamOptions for why it arrives at all.
	Usage *oaiUsage `json:"usage"`
}

func oaiParseStream(resp *http.Response, cb common.StreamCallbacks) error {
	from := common.Provenance{Vendor: common.VendorOpenAI, Surface: common.SurfaceChatCompletions}

	var (
		text    strings.Builder
		textID  uint64
		nextID  uint64
		calls   = map[int]*oaiStreamCall{}
		order   []int
		usage   common.Usage
		started bool
		perr    error
	)

	// Part ids are allocated in the order blocks are DISCOVERED, not by their
	// position in the finished list. Position is unknowable mid-stream: a
	// tool call's arguments can arrive before it is settled whether any text
	// part exists at all. Because the same function reports the finals, the
	// id only ever has to mean "the same part".
	alloc := func() uint64 { nextID++; return nextID }

	readErr := ReadSSE(resp.Body, func(eventType string, data []byte) {
		cb.Frame(eventType, data)
		if perr != nil {
			return
		}

		var chunk oaiStreamChunk
		if err := json.Unmarshal(data, &chunk); err != nil {
			perr = fmt.Errorf("openai: stream chunk: %w", err)
			return
		}
		if chunk.Model != "" {
			from.Model = chunk.Model
		}
		if !started {
			started = true
			cb.Emit(common.Event{Type: common.ResponseStarted})
		}
		// The usage chunk carries no choices. Recording it and returning is
		// not an optimisation; indexing choices[0] here would panic.
		if chunk.Usage != nil {
			usage = oaiCanonicalUsage(*chunk.Usage)
		}
		if len(chunk.Choices) == 0 {
			return
		}
		d := chunk.Choices[0].Delta

		if d.Content != nil && *d.Content != "" {
			if textID == 0 {
				textID = alloc()
			}
			text.WriteString(*d.Content)
			cb.Delta(textID, common.DeltaText, *d.Content)
		}

		for _, tc := range d.ToolCalls {
			c, ok := calls[tc.Index]
			if !ok {
				c = &oaiStreamCall{PartID: alloc()}
				calls[tc.Index] = c
				order = append(order, tc.Index)
			}
			// id and name arrive once, on the first fragment; arguments
			// accumulate over many. Overwriting on empty would erase them.
			if tc.ID != "" {
				c.ID = tc.ID
			}
			if tc.Function.Name != "" {
				c.Name = tc.Function.Name
				cb.Delta(c.PartID, common.DeltaToolCall, tc.Function.Name)
			}
			if tc.Function.Arguments != "" {
				c.Args.WriteString(tc.Function.Arguments)
				// Display only: incomplete JSON until the stream ends.
				cb.Delta(c.PartID, common.DeltaToolCall, tc.Function.Arguments)
			}
		}
	})

	if perr != nil {
		return perr
	}
	if readErr != nil {
		return fmt.Errorf("openai: stream: %w", readErr)
	}

	// Text first, then tool calls by index: the same order the whole-document
	// path produces, so both paths render back to the same request.
	var parts common.PartList
	ids := []uint64{}
	if text.Len() > 0 {
		parts = append(parts, common.TextPart{Text: text.String()})
		ids = append(ids, textID)
	}
	sort.Ints(order)
	for _, i := range order {
		c := calls[i]
		parts = append(parts, common.ToolCallPart{
			CallID: c.ID, From: from, Name: c.Name,
			Args: jsonObject(json.RawMessage(c.Args.String())),
		})
		ids = append(ids, c.PartID)
	}

	cb.Emit(common.Event{Type: common.ResponseEnded, Response: &common.ResponseData{
		Parts: parts, From: from, Usage: usage,
	}})
	for n, p := range parts {
		cb.Final(ids[n], p)
	}
	return nil
}
