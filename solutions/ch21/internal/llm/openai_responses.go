package llm

// The Responses surface.
//
// This is the same vendor as openai.go and a different endpoint, which is
// exactly why it is a separate seam rather than a flag inside the existing
// one. Chat Completions and Responses disagree about the shape of nearly
// everything: messages become typed input items, tool declarations lose a
// level of nesting, the system prompt stops being a message, and reasoning
// acquires a representation the older surface has nowhere to put.
//
// Three things the older surface could not do, which is the whole reason for
// the move:
//
//   1. Reasoning SUMMARIES stream. Chat Completions returns a reasoning token
//      COUNT and nothing else: the model's intermediate thinking is billed and
//      discarded. Here it arrives incrementally, which for an operator reading
//      along is the difference between watching the work and waiting for it.
//
//   2. Reasoning SURVIVES a turn. The encrypted reasoning item can be sent
//      back, so a model can pick up its own prior thinking on the next round
//      instead of restarting cold. openai.go drops this material with a
//      comment saying the surface has nowhere to put it. This surface does.
//
//   3. Thinking and tools travel together. Some models reject reasoning_effort
//      alongside function tools on /v1/chat/completions with a 400 that names
//      /v1/responses as the fix. NoThinkingWithTools exists to work around
//      that by discarding the effort; on this surface the workaround is not
//      needed, because the restriction is not here.
//
// The cache boundary is the subtle part, and it is a design constraint rather
// than a preference. A cache breakpoint attaches to a CONTENT BLOCK. The
// top-level "instructions" field is a plain string, so it has no block to
// attach one to: a system prompt sent that way is uncacheable, and on the
// largest, most stable, earliest span of the request that is the single most
// expensive place to lose caching. So the constitution travels as a
// developer-role message with its own content block. "instructions" is a legal
// field and we are not forbidden from using it; we decline it because of where
// the breakpoints can go.

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// responsesSeam renders and parses OpenAI's /v1/responses surface.
type responsesSeam struct{}

// --- request ---------------------------------------------------------------

type respRequest struct {
	Model string     `json:"model"`
	Input []respItem `json:"input"`
	Tools []respTool `json:"tools,omitempty"`

	// Stateless by construction. store:false means the vendor keeps no copy
	// of the conversation, which in turn means previous_response_id cannot be
	// used to continue it: there is nothing on the far side to continue from.
	// The whole conversation is resent every turn, as it has been since
	// chapter 2, and the event log stays the only authoritative record.
	//
	// This is also why Include below is not optional for us. With no server
	// copy, reasoning that is not returned to us is reasoning we can never
	// send back.
	Store  bool `json:"store"`
	Stream bool `json:"stream"`

	Include         []string       `json:"include,omitempty"`
	Reasoning       *respReasoning `json:"reasoning,omitempty"`
	MaxOutputTokens int            `json:"max_output_tokens,omitempty"`
	PromptCacheKey  string         `json:"prompt_cache_key,omitempty"`

	// Declaring explicit mode with nothing marked turns caching OFF, so this
	// is set from EVIDENCE that a breakpoint attached, never from the
	// intention to place one. Same footgun as the Chat Completions renderer,
	// same rule: see the marked variable in Render.
	PromptCacheOptions *respCacheOptions `json:"prompt_cache_options,omitempty"`
}

type respCacheOptions struct {
	Mode string `json:"mode"`
}

type respReasoning struct {
	Effort string `json:"effort,omitempty"`

	// Asking for a summary is what makes summary events stream at all. The
	// level matters more than it looks: measured against the live endpoint on
	// a single prompt, "detailed" produced 314 summary deltas, "auto" 186 and
	// "concise" 4.
	Summary string `json:"summary,omitempty"`
}

// respTool is the flat tool shape. Chat Completions nests the real
// declaration under a "function" object; here the fields sit at the top
// level. Nothing about the tool changed, only where the vendor reads it from.
type respTool struct {
	Type        string          `json:"type"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Parameters  json.RawMessage `json:"parameters,omitempty"`
}

// respItem is one element of the input array.
//
// The surface uses a tagged union: "message", "function_call",
// "function_call_output", "reasoning". Go has no sum type, so this is one
// struct with the union of the fields and omitempty doing the discriminating.
//
// raw is the exception and the important one. When replaying material the
// vendor gave us — an encrypted reasoning item — we send back the bytes we
// received, verbatim. Re-encoding through the typed fields above would risk
// changing them, and for an AEAD-encrypted blob a changed byte is a rejected
// request. Round-tripping opaque data through a struct that understands it is
// how you corrupt it.
type respItem struct {
	raw json.RawMessage

	Type string `json:"type"`

	Role    string        `json:"role,omitempty"`
	Content []respContent `json:"content,omitempty"`

	CallID    string `json:"call_id,omitempty"`
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`

	Output string `json:"output,omitempty"`
}

func (i respItem) MarshalJSON() ([]byte, error) {
	if len(i.raw) > 0 {
		return i.raw, nil
	}
	type plain respItem // sheds this method, so no infinite recursion
	return json.Marshal(plain(i))
}

type respContent struct {
	Type string `json:"type"`
	Text string `json:"text"`

	// Where a cache breakpoint can actually live. This field is the reason
	// the system prompt is a developer message instead of "instructions".
	CacheBreakpoint *respBreakpoint `json:"prompt_cache_breakpoint,omitempty"`
}

type respBreakpoint struct {
	Mode string `json:"mode"`
}

// markRespCache attaches a breakpoint to the last content block of the item
// at pos, and reports whether it managed to. A caller that marks nothing must
// not declare explicit mode.
func markRespCache(items []respItem, pos int) bool {
	if pos < 0 || pos >= len(items) {
		return false
	}
	blocks := items[pos].Content
	if len(blocks) == 0 {
		return false // a function_call item has no block to mark
	}
	blocks[len(blocks)-1].CacheBreakpoint = &respBreakpoint{Mode: "explicit"}
	return true
}

// --- render ----------------------------------------------------------------

func (responsesSeam) Render(c *common.Context, cfg common.Config) (*http.Request, error) {
	target := common.Provenance{
		Vendor:  common.VendorOpenAI,
		Model:   cfg.Model,
		Surface: common.SurfaceResponses,
	}

	items := []respItem{}

	// The constitution, as a developer-role message. See the file comment for
	// why this is not the "instructions" field.
	if cfg.SystemPrompt != "" {
		items = append(items, respItem{
			Type: "message",
			Role: "developer",
			Content: []respContent{{
				Type: "input_text",
				Text: cfg.SystemPrompt,
			}},
		})
	}

	systemItem := -1
	if len(items) > 0 {
		systemItem = 0
	}
	handoffIdx := newestHandoffIndex(c.Dialogue)
	handoffItem, stableItem := -1, -1

	for i, entry := range c.Dialogue {
		if entry.Kind == common.KindTools {
			continue // declarations travel in the tools array
		}
		if entry.Kind == common.KindRecall {
			// No system role inside the input array on this surface, and a
			// developer message mid-conversation would read as a second
			// constitution. Recalled material is context for the user's
			// request, so it renders as user input.
			if text := recallText(entry); text != "" {
				items = append(items, respItem{
					Type:    "message",
					Role:    "user",
					Content: []respContent{{Type: "input_text", Text: text}},
				})
			}
			continue
		}
		r, err := classify(entry, cfg)
		if err != nil {
			return nil, err
		}
		if len(r.Blobs) > 0 {
			return nil, fmt.Errorf("openai responses: rendering a blob part is not "+
				"implemented in this chapter (%s at %s)", r.Blobs[0].MIME, r.Blobs[0].Ref.Locator)
		}

		switch entry.Actor {
		case common.ActorTool:
			for _, res := range r.Tools {
				items = append(items, respItem{
					Type:   "function_call_output",
					CallID: res.CallID,
					Output: resultText(res),
				})
			}

		case common.ActorAgent:
			// Reasoning first. The vendor emits the reasoning item BEFORE the
			// message it supports, and sending it back out of order makes it
			// reasoning about a turn that already happened.
			//
			// SameModel is the gate. Replay material is tagged with the
			// vendor, model and surface that issued it, and a blob minted by
			// one model is not meaningful to another: at best it is rejected,
			// at worst it is silently misread. This is the binding that makes
			// opaque replay safe to carry at all.
			for _, op := range r.Raw {
				if !op.From.SameModel(target) {
					continue
				}
				items = append(items, respItem{raw: op.Data})
			}

			if text := joinTexts(r.Texts); text != "" {
				items = append(items, respItem{
					Type: "message",
					Role: "assistant",
					Content: []respContent{{
						// Assistant text the model previously produced is
						// "output_text" even when we are the ones sending it.
						// The type names the ORIGIN of the text, not the
						// direction of this particular request.
						Type: "output_text",
						Text: text,
					}},
				})
			}

			for i, call := range r.Calls {
				items = append(items, respItem{
					Type:      "function_call",
					CallID:    callIDFor(call, "call", entry.Seq, i),
					Name:      call.Name,
					Arguments: string(jsonObject(call.Args)),
				})
			}

			// Unlike Chat Completions, a turn that produced nothing needs no
			// placeholder here: an empty assistant message is simply not
			// appended, and the input array is a list of things that happened
			// rather than a strict alternation that must be filled.

		default: // common.ActorHuman
			items = append(items, respItem{
				Type:    "message",
				Role:    "user",
				Content: []respContent{{Type: "input_text", Text: joinTexts(r.Texts)}},
			})
		}

		if i == handoffIdx && len(items) > 0 {
			handoffItem = len(items) - 1
		}
	}

	// Everything below is ephemeral and must stay outside the marked prefix.
	if len(items) > 0 {
		stableItem = len(items) - 1
	}

	eph := c.EphemeraFor(cfg.Model)
	if len(eph) > 0 {
		if text := ephemeraText(eph); text != "" {
			items = append(items, respItem{
				Type:    "message",
				Role:    "user",
				Content: []respContent{{Type: "input_text", Text: text}},
			})
		}
	}

	tools := respTools(common.EffectiveTools(cfg.Tools, c))

	// Reasoning. Note what is NOT here: the NoThinkingWithTools workaround
	// from the Chat Completions renderer. That flag records a restriction on
	// the OTHER endpoint, whose own error message names this one as the fix.
	// Honouring it here would import a limitation that does not exist on this
	// surface, and would silence thinking on every tool-bearing turn — which
	// is every turn an agent takes.
	var reasoning *respReasoning
	effort, _ := common.ThinkingFor(cfg)
	if effort != common.ThinkingOff {
		reasoning = &respReasoning{Effort: effort.String()}
		if common.StreamingFor(cfg)&common.StreamReasoningSummary != 0 {
			reasoning.Summary = "auto"
		}
	}

	// With store:false the vendor keeps nothing, so reasoning we do not ask
	// for is reasoning we can never replay. Only ask when the model can
	// actually use it.
	var include []string
	if reasoning != nil {
		include = append(include, "reasoning.encrypted_content")
	}

	// Explicit breakpoints, same positions and same policy as the Anthropic
	// and Chat Completions renderers. Only the encoding is local.
	//
	// The consumer-plan route refuses explicit caching outright, markers and
	// mode alike, so on that route nothing is marked; prompt_cache_options is
	// only ever sent when something is, so this one gate covers both fields.
	marked := false
	if usesExplicitBreakpoints(cfg.Model) && !cfg.Route.Forbids("prompt_cache_breakpoint") {
		budget := 4
		if f, ok := common.LookupModel(cfg.Model); ok && f.MaxCacheWrites > 0 {
			budget = f.MaxCacheWrites
		}
		for _, pos := range []int{systemItem, handoffItem, stableItem} {
			if budget <= 0 {
				break
			}
			if markRespCache(items, pos) {
				marked = true
				budget--
			}
		}
	}

	body := respRequest{
		Model: cfg.Model,
		Input: items,
		Tools: tools,
		// Store must be false, and this is not a preference. The ChatGPT plan
		// route rejects the request outright with {"detail":"Store must be set
		// to false"} if it is true, so the literal is load-bearing. Measured
		// 2026-10-04 by flipping it and reading the 400.
		//
		// A consequence worth knowing when chasing prompt-cache misses: because
		// nothing is stored server side, every request must carry the whole
		// conversation, so the cacheable prefix is the only thing standing
		// between this agent and paying full price for the history on each turn.
		Store:     false,
		Stream:    common.StreamingFor(cfg) != 0,
		Include:   include,
		Reasoning: reasoning,
	}

	// The consumer-plan route rejects max_output_tokens outright. Not
	// ignores — refuses, so a turn that sends it does not produce a shorter
	// answer, it produces no answer. The metered route accepts it happily,
	// which is exactly why this is read from the route table rather than
	// hardcoded either way.
	if !cfg.Route.Forbids("max_output_tokens") {
		body.MaxOutputTokens = cfg.MaxTokens
	}

	// Some routes refuse a non-streaming request. We already prefer
	// streaming everywhere, so this only matters when a caller has turned it
	// off; honouring the route means a disabled stream degrades to a slower
	// answer rather than to a 400.
	if cfg.Route.RequiresStreaming {
		body.Stream = true
	}

	if marked {
		body.PromptCacheOptions = &respCacheOptions{Mode: "explicit"}
	}

	// OpenAI serves a cached prefix from one machine, and a prefix cached on
	// one machine is invisible to the others. prompt_cache_key is the
	// documented handle on that routing: requests carrying the same key are
	// steered to the same place.
	//
	// Honest accounting of what this is worth. The published guidance says
	// GPT-5.6 and later route automatically and do not need the key, and a
	// measured run on the ChatGPT plan route agreed: adding it changed
	// nothing, because that route wrote no cache at all until a request
	// crossed roughly twenty thousand tokens, and below that a prefix held
	// stable to 99% still returned zero. So this is not the fix for a cold
	// cache, and anyone arriving here hunting one should look at the route
	// and the request size first.
	//
	// It is sent anyway because the vendor's own client sends it, the field
	// is free, and it is the only documented lever on cache routing we have
	// if a future route or an API key reaches a multi-machine pool.
	//
	// The key is derived from the frozen prefix rather than from a session
	// id, which is the part worth preserving. Two runs of the same agent
	// share tools and a constitution, so they should share a machine, and
	// deriving the key from that content makes it true across restarts with
	// nothing to persist and nothing to expire. A prefix that genuinely
	// changes mints a new key, which is correct rather than unfortunate: the
	// old machine holds nothing worth reaching for.
	if !cfg.Route.Forbids("prompt_cache_key") {
		body.PromptCacheKey = respCacheKey(body.Model, tools, items)
	}

	return newJSONRequest("POST", cfg.BaseURL+"/v1/responses",
		body,
		map[string]string{"Authorization": "Bearer " + cfg.APIKey})
}

func respTools(decls []common.ToolDecl) []respTool {
	var out []respTool
	for _, d := range decls {
		out = append(out, respTool{
			Type:        "function",
			Name:        d.Name,
			Description: d.Description,
			Parameters:  d.Schema,
		})
	}
	return out
}

// --- parse -----------------------------------------------------------------

// respUsage follows the same SUBSET convention as Chat Completions:
// InputTokens is the grand total and the details are buckets carved out of
// it, so the canonical disjoint form is reached by subtraction. The field
// names differ from the older surface (input_tokens, not prompt_tokens;
// input_tokens_details, not prompt_tokens_details), which is the kind of
// gratuitous divergence that makes a per-surface parser the honest choice.
type respUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`

	InputTokensDetails struct {
		CachedTokens     int `json:"cached_tokens"`
		CacheWriteTokens int `json:"cache_write_tokens"`
	} `json:"input_tokens_details"`

	OutputTokensDetails struct {
		ReasoningTokens int `json:"reasoning_tokens"`
	} `json:"output_tokens_details"`
}

func respCanonicalUsage(u respUsage) common.Usage {
	// Reasoning tokens are already counted inside OutputTokens. They are
	// reported separately for visibility, not added on top; adding them would
	// double-count every thinking turn.
	input := u.InputTokens - u.InputTokensDetails.CachedTokens - u.InputTokensDetails.CacheWriteTokens
	if input < 0 {
		input = 0
	}
	return common.Usage{
		Input:      input,
		CacheRead:  u.InputTokensDetails.CachedTokens,
		CacheWrite: u.InputTokensDetails.CacheWriteTokens,
		Output:     u.OutputTokens,
	}
}

// respOutputItem is one element of the completed response's output array.
// Raw keeps the original bytes so a reasoning item can be replayed verbatim.
type respOutputItem struct {
	Raw json.RawMessage `json:"-"`

	Type      string `json:"type"`
	ID        string `json:"id"`
	Role      string `json:"role"`
	CallID    string `json:"call_id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`

	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`

	Summary []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"summary"`

	EncryptedContent string `json:"encrypted_content"`
}

type respResponse struct {
	Model  string            `json:"model"`
	Output []json.RawMessage `json:"output"`
	Usage  respUsage         `json:"usage"`
	Error  *struct {
		Type    string `json:"type"`
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type respEnvelope struct {
	Response respResponse `json:"response"`
}

// respAssemble turns a completed response object into canonical parts.
//
// This is the authority for what the turn produced, even while streaming.
// The deltas are for the human reading along; the final object is what gets
// recorded. Taking finals from here rather than from accumulated deltas means
// a dropped or reordered frame cannot corrupt the log, and it is the only
// place the encrypted reasoning blob appears in full.
//
// keys runs parallel to parts and names each part the way the stream names its
// deltas ("text:<item>", "call:<item>", "sum:<item>:0"), so the stream parser
// can finalize a part under the id its deltas already used. A final under a
// fresh id is a second part to every observer: the GUI draws the answer twice
// and, because the new id has no streamed text, reads it aloud twice.
func respAssemble(r respResponse, model string) (common.PartList, []string, common.Provenance, common.Usage, error) {
	from := common.Provenance{
		Vendor:  common.VendorOpenAI,
		Model:   model,
		Surface: common.SurfaceResponses,
	}
	if from.Model == "" {
		from.Model = r.Model
	}

	var parts common.PartList
	var keys []string
	for _, raw := range r.Output {
		var item respOutputItem
		if err := json.Unmarshal(raw, &item); err != nil {
			return nil, nil, from, common.Usage{}, fmt.Errorf("openai responses: output item: %w", err)
		}
		switch item.Type {
		case "reasoning":
			// Replayable only if there is something to replay. A reasoning
			// item with no encrypted content is a summary we can show but not
			// send back, and storing it as replay material would mean
			// offering the vendor an item it cannot use.
			if item.EncryptedContent == "" {
				continue
			}
			// One part per item, though its summary streamed as one part
			// per summary_index. The final takes the first summary's id; the
			// later summary parts are complete as streamed.
			parts = append(parts, common.OpaquePart{From: from, Data: raw})
			keys = append(keys, "sum:"+item.ID+":0")

		case "message":
			text := ""
			for _, c := range item.Content {
				text += c.Text
			}
			if text != "" {
				parts = append(parts, common.TextPart{Text: text})
				keys = append(keys, "text:"+item.ID)
			}

		case "function_call":
			parts = append(parts, common.ToolCallPart{
				CallID: item.CallID,
				From:   from,
				Name:   item.Name,
				Args:   jsonObject(json.RawMessage(item.Arguments)),
			})
			keys = append(keys, "call:"+item.ID)
		}
	}
	return parts, keys, from, respCanonicalUsage(r.Usage), nil
}

// respSummaryText extracts the human-readable summary from a reasoning item.
// It is what lets a reasoning part render as something other than a blob.
func respSummaryText(op common.OpaquePart) string {
	var item respOutputItem
	if err := json.Unmarshal(op.Data, &item); err != nil {
		return ""
	}
	out := ""
	for _, s := range item.Summary {
		out += s.Text
	}
	return out
}

func (s responsesSeam) Parse(resp *http.Response, cb common.StreamCallbacks) error {
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return parseErrorResponse(resp, cb)
	}
	if isSSE(resp) {
		return respParseStream(resp, cb)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("openai responses: reading body: %w", err)
	}
	cb.Frame("", body)

	var r respResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return fmt.Errorf("openai responses: %w", err)
	}
	if r.Error != nil {
		return fmt.Errorf("openai responses: %s: %s", r.Error.Code, r.Error.Message)
	}
	parts, _, from, usage, err := respAssemble(r, "")
	if err != nil {
		return err
	}

	pm := newPartIDMapper(cb)
	cb.Emit(common.Event{Type: common.ResponseStarted})
	emitLengthOneDeltas(parts, pm, cb, respSummaryText)
	emitFinals(parts, pm, cb)
	cb.Emit(common.Event{Type: common.ResponseEnded, Response: &common.ResponseData{
		Parts: parts, From: from, Usage: usage,
	}})
	return nil
}

// respParseStream consumes the event stream.
//
// Two independent jobs, and keeping them separate is the point. Deltas go out
// as they arrive so an operator hears the answer being written. The recorded
// result comes from response.completed and nothing else. If the two ever
// disagree the completed object wins, because it is the one the vendor calls
// final.
func respParseStream(resp *http.Response, cb common.StreamCallbacks) error {
	var (
		started  bool
		finished bool
		failure  error

		parts common.PartList
		keys  []string
		from  common.Provenance
		usage common.Usage
	)

	// Items as they finish, by output_index. The ChatGPT plan route sends
	// response.completed with an empty output array and delivers every item
	// only through response.output_item.done, so a parser that reads the
	// completed object alone ends every plan turn with nothing in it.
	doneItems := map[int]json.RawMessage{}

	pm := newPartIDMapper(cb)

	// One streamed part per (item, summary index). Summary parts are numbered
	// independently of output items, so the key has to carry both or two
	// summaries on one item would collapse into a single part.
	live := map[string]uint64{}
	partFor := func(key string) uint64 {
		if id, ok := live[key]; ok {
			return id
		}
		// pm.alloc, not cb.AllocPartID: callers that only want the result
		// (the recall judge, compressors) leave AllocPartID nil, and calling
		// it directly panics on the first streamed delta.
		id := pm.alloc()
		live[key] = id
		return id
	}

	err := ReadSSE(resp.Body, func(eventType string, data []byte) {
		cb.Frame(eventType, data)

		if !started {
			started = true
			cb.Emit(common.Event{Type: common.ResponseStarted})
		}

		switch eventType {
		case "response.output_text.delta":
			var ev struct {
				ItemID string `json:"item_id"`
				Delta  string `json:"delta"`
			}
			if json.Unmarshal(data, &ev) == nil && ev.Delta != "" {
				cb.Delta(partFor("text:"+ev.ItemID), common.DeltaText, ev.Delta)
			}

		case "response.reasoning_summary_text.delta":
			// The accessibility payload, and the reason for the migration.
			// It is its own kind rather than DeltaText because a summary is
			// not the answer: a reader needs to be able to tell the model's
			// thinking apart from what it decided, and a consumer that cannot
			// distinguish them will read the thinking aloud as if it were the
			// result.
			var ev struct {
				ItemID       string `json:"item_id"`
				SummaryIndex int    `json:"summary_index"`
				Delta        string `json:"delta"`
			}
			if json.Unmarshal(data, &ev) == nil && ev.Delta != "" {
				key := fmt.Sprintf("sum:%s:%d", ev.ItemID, ev.SummaryIndex)
				cb.Delta(partFor(key), common.DeltaReasoningSummary, ev.Delta)
			}

		case "response.function_call_arguments.delta":
			var ev struct {
				ItemID string `json:"item_id"`
				Delta  string `json:"delta"`
			}
			if json.Unmarshal(data, &ev) == nil && ev.Delta != "" {
				cb.Delta(partFor("call:"+ev.ItemID), common.DeltaToolCall, ev.Delta)
			}

		case "response.output_item.done":
			var ev struct {
				OutputIndex int             `json:"output_index"`
				Item        json.RawMessage `json:"item"`
			}
			if json.Unmarshal(data, &ev) == nil && len(ev.Item) > 0 {
				doneItems[ev.OutputIndex] = ev.Item
			}

		case "response.completed":
			var env respEnvelope
			if err := json.Unmarshal(data, &env); err != nil {
				failure = fmt.Errorf("openai responses: completed event: %w", err)
				return
			}
			if len(env.Response.Output) == 0 {
				env.Response.Output = orderedItems(doneItems)
			}
			p, k, f, u, err := respAssemble(env.Response, "")
			if err != nil {
				failure = err
				return
			}
			parts, keys, from, usage, finished = p, k, f, u, true

		case "response.failed", "response.incomplete":
			// A mid-stream failure arrives with HTTP 200 and a clean SSE
			// envelope: the status line was written before the vendor knew
			// anything was wrong. A parser that only checks the status code
			// reports success on a turn that produced nothing, which is how a
			// plan-usage limit turns into a silently empty answer.
			var env respEnvelope
			if json.Unmarshal(data, &env) == nil && env.Response.Error != nil {
				failure = fmt.Errorf("openai responses: %s: %s",
					env.Response.Error.Code, env.Response.Error.Message)
				return
			}
			failure = fmt.Errorf("openai responses: stream ended with %s", eventType)
		}
	})

	if failure != nil {
		return failure
	}
	if err != nil {
		return fmt.Errorf("openai responses: reading stream: %w", err)
	}
	if !finished {
		// The stream ended without response.completed. Treating that as
		// success would record a truncated turn as a whole one.
		return fmt.Errorf("openai responses: stream ended without response.completed")
	}

	// A part that streamed finalizes under its delta id; one that never
	// streamed (a call with empty arguments, a reasoning item with no summary)
	// gets a fresh id, exactly as the whole-document path would give it.
	for i, p := range parts {
		id, ok := live[keys[i]]
		if !ok {
			id = pm.id(i)
		}
		cb.Final(id, p)
	}
	cb.Emit(common.Event{Type: common.ResponseEnded, Response: &common.ResponseData{
		Parts: parts, From: from, Usage: usage,
	}})
	return nil
}

// orderedItems returns the finished items in output_index order. The
// completed object would have listed them in that order, and the order is the
// conversation: a message that explains a tool call must stay ahead of it.
func orderedItems(done map[int]json.RawMessage) []json.RawMessage {
	idx := make([]int, 0, len(done))
	for i := range done {
		idx = append(idx, i)
	}
	sort.Ints(idx)
	out := make([]json.RawMessage, 0, len(idx))
	for _, i := range idx {
		out = append(out, done[i])
	}
	return out
}

// respCacheKey fingerprints the part of a request that is meant never to
// change: the model, the tool declarations, and the leading system block.
//
// It deliberately excludes the conversation. Including it would mint a fresh
// key on every turn, steering each request to a different machine and so
// guaranteeing the miss the key exists to prevent. The failure would be
// silent, which is the reason this is a named function with a comment rather
// than an expression inlined at the call site.
func respCacheKey(model string, tools []respTool, items []respItem) string {
	h := sha256.New()
	h.Write([]byte(model))
	if b, err := json.Marshal(tools); err == nil {
		h.Write(b)
	}
	// Only the first item: on this route that is the developer block holding
	// the constitution, because top-level instructions are uncacheable here.
	if len(items) > 0 {
		if b, err := json.Marshal(items[0]); err == nil {
			h.Write(b)
		}
	}
	return "en-" + hex.EncodeToString(h.Sum(nil))[:32]
}
