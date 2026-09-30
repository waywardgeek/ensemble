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

	// PromptCacheOptions opts this request into explicit cache breakpoints.
	//
	// It MUST stay omitted unless a marker actually landed on a message.
	// Declaring explicit mode disables this vendor's implicit caching, so a
	// request that asks for explicit mode and then marks nothing caches
	// nothing at all — strictly worse than saying nothing. Anthropic has no
	// equivalent trap, because it has no implicit cache to lose.
	PromptCacheOptions *oaiCacheOptions `json:"prompt_cache_options,omitempty"`
}

type oaiCacheOptions struct {
	Mode string `json:"mode"`
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

	// CacheBreak asks for a cache breakpoint at the end of this message. It is
	// not a wire field of its own: the marker has to ride inside a content
	// block, so setting this changes how Content is encoded. See MarshalJSON.
	CacheBreak bool `json:"-"`
}

// oaiTextPart is the array-of-parts form of a message body. A breakpoint
// attaches to a content block, so a marked message cannot send its content as
// a bare string.
type oaiTextPart struct {
	Type       string          `json:"type"`
	Text       string          `json:"text"`
	Breakpoint *oaiCacheMarker `json:"prompt_cache_breakpoint,omitempty"`
}

type oaiCacheMarker struct {
	Mode string `json:"mode"`
}

// MarshalJSON emits the plain string form unless a breakpoint was requested.
//
// Flipping a message between the two forms as the rolling marker moves past it
// is safe, which is not obvious and was worth measuring. A message was sent
// marked (array form), then the same conversation was sent again with the
// marker removed and that message back in string form: 7,813 of 7,816 prompt
// tokens still read from cache. The bytes on the wire differ; the tokens the
// vendor hashes do not.
func (m oaiMsg) MarshalJSON() ([]byte, error) {
	type plain oaiMsg
	if !m.CacheBreak || m.Content == nil {
		return json.Marshal(plain(m))
	}
	return json.Marshal(struct {
		Role       string        `json:"role"`
		Content    []oaiTextPart `json:"content"`
		ToolCalls  []oaiToolCall `json:"tool_calls,omitempty"`
		ToolCallID string        `json:"tool_call_id,omitempty"`
	}{
		Role: m.Role,
		Content: []oaiTextPart{{
			Type:       "text",
			Text:       *m.Content,
			Breakpoint: &oaiCacheMarker{Mode: "explicit"},
		}},
		ToolCalls:  m.ToolCalls,
		ToolCallID: m.ToolCallID,
	})
}

// markOAICache attaches a breakpoint at pos, walking backward to the nearest
// message that can carry one.
//
// An assistant message that only issues tool calls has null content and so has
// no block to attach a marker to. Walking back lands the marker slightly
// earlier in the prefix, which costs a little coverage; dropping it silently
// would cost the whole cache entry at that boundary.
//
// Two positions resolving to the same message is fine and expected right after
// a compaction, when the bound and the anchor coincide. Setting the flag twice
// marks once.
func markOAICache(msgs []oaiMsg, pos int) bool {
	if pos < 0 || pos >= len(msgs) {
		return false
	}
	for i := pos; i >= 0; i-- {
		if msgs[i].Content != nil {
			msgs[i].CacheBreak = true
			return true
		}
	}
	return false
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

	// Breakpoint positions, mirroring the Anthropic renderer. Where to mark is
	// shared policy; only the encoding below is vendor-specific.
	systemMsg := -1
	if len(msgs) > 0 {
		systemMsg = 0
	}
	handoffIdx := newestHandoffIndex(c.Dialogue)
	handoffMsg, stableMsg, anchorMsg := -1, -1, -1

	for i, entry := range c.Dialogue {
		if entry.Kind == common.KindTools {
			continue // no in-dialog declarations here: see EffectiveTools
		}
		if entry.Kind == common.KindRecall {
			// OpenAI accepts a system message anywhere in the list, so
			// recalled material renders as one. Compare the Gemini renderer,
			// which has no such role and must do something else entirely —
			// that divergence is precisely why placement is the renderer's
			// decision and not the retriever's.
			if text := recallText(entry); text != "" {
				msgs = append(msgs, oaiMsg{Role: "system", Content: &text})
			}
			continue
		}
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
			// The anchor sits at the end of the previous exchange, so record
			// it before this turn's message lands. The last human turn wins,
			// which is what makes the pair roll forward.
			if len(msgs) > 0 {
				anchorMsg = len(msgs) - 1
			}
			msgs = append(msgs, oaiMsg{Role: "user", Content: strptr(joinTexts(r.Texts))})
		}

		if i == handoffIdx && len(msgs) > 0 {
			handoffMsg = len(msgs) - 1
		}
	}

	// Everything below is ephemeral and must stay outside the marked prefix.
	if len(msgs) > 0 {
		stableMsg = len(msgs) - 1
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

	tools := oaiTools(common.EffectiveTools(cfg.Tools, c))

	// Some models reject a reasoning budget in the same request as function
	// tools, and reject it with a 400 rather than a degraded answer. Since an
	// agent always has tools, leaving the effort in place makes such a model
	// unusable rather than merely unthinking: every turn fails.
	//
	// Drop the effort rather than the tools. An agent with no tools cannot do
	// anything at all, while one with no declared reasoning budget still works.
	// Of the two things that cannot travel together, tools are the one to keep.
	//
	// It must be sent AS "none" and not omitted, which is the part that cost a
	// live run to learn. Omitting the field does not mean "no reasoning", it
	// means "your default", and the default for these models is a reasoning
	// effort, so the request is refused with the identical error while the
	// field is nowhere in the body. The vendor's own message says to set it to
	// 'none', and it means set.
	if len(tools) > 0 {
		if f, ok := common.LookupModel(cfg.Model); ok && f.NoThinkingWithTools {
			reasoningEffort = "none"
		}
	}

	// Explicit breakpoints: the compaction bound first, then the rolling pair,
	// with the system prompt taking the fourth and last slot the vendor allows.
	//
	// marked records whether any of them actually attached. The request option
	// below is driven by that evidence rather than by the intent to place a
	// marker, because declaring explicit mode with nothing marked turns caching
	// off entirely. See usesExplicitBreakpoints.
	marked := false
	if usesExplicitBreakpoints(cfg.Model) {
		for _, pos := range []int{systemMsg, handoffMsg, stableMsg, anchorMsg} {
			if markOAICache(msgs, pos) {
				marked = true
			}
		}
	}

	body := oaiRequest{
		Model:           cfg.Model,
		MaxTokens:       cfg.MaxTokens,
		Messages:        msgs,
		Tools:           tools,
		Stream:          common.StreamingFor(cfg) != 0,
		ReasoningEffort: reasoningEffort,
	}
	if body.Stream {
		body.StreamOptions = &oaiStreamOptions{IncludeUsage: true}
	}
	if marked {
		body.PromptCacheOptions = &oaiCacheOptions{Mode: "explicit"}
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
	pm := newPartIDMapper(cb)
	emitLengthOneDeltas(parts, pm, cb, nil)
	cb.Emit(common.Event{Type: common.ResponseStarted})
	cb.Emit(common.Event{Type: common.ResponseEnded, Response: &common.ResponseData{
		Parts: parts, From: from, Usage: usage,
	}})
	emitFinals(parts, pm, cb)
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
	alloc := func() uint64 {
		if cb.AllocPartID != nil {
			return cb.AllocPartID()
		}
		return 0
	}

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
