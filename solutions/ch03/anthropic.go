package main

// Anthropic — Messages API.
//
// Wire format verified against platform.claude.com on 2026-09-12. Wire formats
// drift; this file is the place that has to know, and it is the only place.

import (
	"encoding/json"
	"fmt"
	"net/http"
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
}

// anthTool is Anthropic's declaration shape. The schema key is `input_schema`
// — the one vendor of the three that does not call it `parameters`.
type anthTool struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	InputSchema json.RawMessage `json:"input_schema"`
}

func anthTools(decls []ToolDecl) []anthTool {
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

func (anthropicSeam) Render(c *Context, cfg Config) (*http.Request, error) {
	target := Provenance{Vendor: VendorAnthropic, Model: cfg.Model, Surface: SurfaceMessages}

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
		switch entry.Actor {
		case ActorTool:
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

		case ActorAgent:
			var blocks []anthBlock
			// Opaque replay material goes back only to the EXACT model that
			// issued it. Vendor is not a fine enough grain.
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

		default: // ActorHuman
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
			if t, ok := p.(TextPart); ok {
				blocks = append(blocks, anthBlock{Type: "text", Text: t.Text})
			}
		}
		appendBlocks("user", blocks, false)
	}

	body := anthRequest{
		Model:     cfg.Model,
		MaxTokens: cfg.MaxTokens,
		System:    cfg.SystemPrompt, // top-level. Store it and you have picked a vendor.
		Messages:  msgs,
		Tools:     anthTools(cfg.Tools),
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

func (anthropicSeam) Parse(status int, body []byte) ([]Event, error) {
	if status != http.StatusOK {
		return []Event{{Type: ErrorOccurred, Error: &ErrorData{
			Status:  status,
			Message: vendorErrorMessage(body),
		}}}, nil
	}
	var resp anthResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("anthropic: %w", err)
	}
	// Provenance comes from the RESPONSE, not from Config: the vendor tells
	// you which model actually answered, and it is not always the one you
	// asked for.
	from := Provenance{Vendor: VendorAnthropic, Model: resp.Model, Surface: SurfaceMessages}

	var parts PartList
	for _, raw := range resp.Content {
		var h anthBlockHeader
		if err := json.Unmarshal(raw, &h); err != nil {
			return nil, fmt.Errorf("anthropic: content block: %w", err)
		}
		switch h.Type {
		case "text":
			parts = append(parts, TextPart{Text: h.Text})
		case "tool_use":
			parts = append(parts, ToolCallPart{
				CallID: h.ID, From: from, Name: h.Name, Args: jsonObject(h.Input),
			})
		default:
			// thinking, redacted_thinking, and anything shipped after this
			// book went to print. Carried verbatim, tagged with the model that
			// produced it, never interpreted.
			parts = append(parts, OpaquePart{From: from, Data: raw})
		}
	}

	return []Event{
		{Type: ResponseStarted},
		{Type: ResponseEnded, Response: &ResponseData{
			Parts: parts,
			From:  from,
			Usage: Usage{
				Input:      resp.Usage.InputTokens,
				CacheWrite: resp.Usage.CacheCreationTokens,
				CacheRead:  resp.Usage.CacheReadTokens,
				Output:     resp.Usage.OutputTokens,
			},
		}},
	}, nil
}

// vendorErrorMessage digs a human-readable message out of whatever error shape
// a vendor used, falling back to the raw body. An HTTP 429 is an
// ErrorOccurred, not a response.
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
