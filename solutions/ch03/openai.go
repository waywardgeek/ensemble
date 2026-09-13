package main

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
	"net/http"
)

type openAISeam struct{}

// --- request -----------------------------------------------------------

type oaiRequest struct {
	Model     string   `json:"model"`
	MaxTokens int      `json:"max_completion_tokens,omitempty"`
	Messages  []oaiMsg `json:"messages"`
	// Omitted when nothing is declared — see anthRequest.Tools.
	Tools []oaiTool `json:"tools,omitempty"`
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

func oaiTools(decls []ToolDecl) []oaiTool {
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

func (openAISeam) Render(c *Context, cfg Config) (*http.Request, error) {
	target := Provenance{Vendor: VendorOpenAI, Model: cfg.Model, Surface: SurfaceChatCompletions}

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
		case ActorTool:
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

		case ActorAgent:
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

		default: // ActorHuman
			msgs = append(msgs, oaiMsg{Role: "user", Content: strptr(joinTexts(r.Texts))})
		}
	}

	if len(c.Ephemera) > 0 {
		if text := ephemeraText(c.Ephemera); text != "" {
			msgs = append(msgs, oaiMsg{Role: "user", Content: strptr(text)})
		}
	}

	return newJSONRequest("POST", cfg.BaseURL+"/v1/chat/completions",
		oaiRequest{Model: cfg.Model, MaxTokens: cfg.MaxTokens, Messages: msgs, Tools: oaiTools(cfg.Tools)},
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

func ephemeraText(parts PartList) string {
	var texts []string
	for _, p := range parts {
		if t, ok := p.(TextPart); ok {
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

func (openAISeam) Parse(status int, body []byte) ([]Event, error) {
	if status != http.StatusOK {
		return []Event{{Type: ErrorOccurred, Error: &ErrorData{
			Status: status, Message: vendorErrorMessage(body),
		}}}, nil
	}
	var resp oaiResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("openai: %w", err)
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("openai: response had no choices")
	}
	from := Provenance{Vendor: VendorOpenAI, Model: resp.Model, Surface: SurfaceChatCompletions}

	var parts PartList
	choice := resp.Choices[0]
	if choice.Message.Content != nil && *choice.Message.Content != "" {
		parts = append(parts, TextPart{Text: *choice.Message.Content})
	}
	for _, tc := range choice.Message.ToolCalls {
		// arguments arrives as a JSON-encoded string; the context stores the
		// decoded object, because "a string that happens to contain JSON" is
		// OpenAI's encoding decision, not a fact about the conversation.
		parts = append(parts, ToolCallPart{
			CallID: tc.ID, From: from, Name: tc.Function.Name,
			Args: jsonObject(json.RawMessage(tc.Function.Arguments)),
		})
	}

	// Convert SUBSET to the canonical DISJOINT form.
	u := resp.Usage
	uncached := u.PromptTokens - u.Details.CachedTokens - u.Details.CacheWriteTokens
	if uncached < 0 {
		uncached = 0
	}

	return []Event{
		{Type: ResponseStarted},
		{Type: ResponseEnded, Response: &ResponseData{
			Parts: parts,
			From:  from,
			Usage: Usage{
				Input:      uncached,
				CacheWrite: u.Details.CacheWriteTokens,
				CacheRead:  u.Details.CachedTokens,
				Output:     u.CompletionTokens,
			},
		}},
	}, nil
}
