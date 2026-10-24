package engine

import (
	"encoding/json"
	"ensemble/internal/common"
	"fmt"
	"strings"
)

func parse(status int, b []byte, c common.Config) ([]common.Event, error) {
	fail := func(message string) ([]common.Event, error) {
		return []common.Event{errorEvent(message)}, fmt.Errorf("%s", message)
	}
	if status != 200 {
		return fail(fmt.Sprintf("provider HTTP status %d", status))
	}
	// The route is not a producer identity. A missing returned model remains
	// unknown even if the caller supplied a resolved target for the next request.
	from := common.Provenance{Vendor: c.Vendor, Surface: c.Surface}
	var envelope map[string]json.RawMessage
	if json.Unmarshal(b, &envelope) != nil || envelope == nil {
		return fail("decoding response failed")
	}
	if c.Vendor == common.Gemini {
		_ = json.Unmarshal(envelope["modelVersion"], &from.Model)
	} else {
		_ = json.Unmarshal(envelope["model"], &from.Model)
	}
	usage, missing, usageErr := parseUsage(c, envelope)
	var parts common.PartList
	var contentErr error
	switch c.Vendor {
	case common.Anthropic:
		parts, contentErr = parseAnthropic(b, from)
	case common.OpenAI:
		if c.Surface == common.Responses {
			parts, contentErr = parseResponses(b, from)
		} else {
			parts, contentErr = parseOpenAI(b, from)
		}
	case common.Gemini:
		parts, contentErr = parseGemini(b, from)
	default:
		return fail("unsupported vendor")
	}
	for i := range parts {
		parts[i].ID = uint64(i)
	}
	// Even an unusable or partially parsed answer is an observed response. Its
	// measured usage survives; missing required facts remain unknown.
	events := []common.Event{{Type: common.ResponseStarted, Response: &common.ResponseData{From: from}}, {Type: common.ResponseEnded, Response: &common.ResponseData{Parts: parts, Usage: usage, UsageMissing: missing, From: from}}}
	usable := len(parts) > 0
	if c.Vendor == common.Anthropic {
		usable = false
		for _, p := range parts {
			if p.Type != "text" || strings.TrimSpace(p.Text) != "" {
				usable = true
			}
		}
	}
	err := usageErr
	if err == nil {
		err = contentErr
	}
	if err == nil && !usable {
		err = fmt.Errorf("response contained no usable text")
	}
	if err != nil {
		events = append(events, errorEvent(err.Error()))
		return events, err
	}
	return events, nil
}

// Decode counts independently. Within valid usage, omitted optional cache
// fields mean zero; supplied invalid fields remain unknown/errors. Missing whole
// usage is not a measured zero. Input excludes reported read/write cache tokens.
func parseUsage(c common.Config, envelope map[string]json.RawMessage) (common.Usage, []string, error) {
	key := "usage"
	if c.Vendor == common.Gemini {
		key = "usageMetadata"
	}
	raw, exists := envelope[key]
	unknown := []string{"input", "output", "cache_read", "cache_write"}
	if !exists || strings.TrimSpace(string(raw)) == "null" {
		return common.UnknownUsage(), unknown, fmt.Errorf("response missing usage")
	}
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil || fields == nil {
		return common.UnknownUsage(), unknown, fmt.Errorf("response contained invalid usage")
	}
	var firstErr error
	missing := []string{}
	mark := func(bucket, message string) {
		for _, old := range missing {
			if old == bucket {
				return
			}
		}
		missing = append(missing, bucket)
		if message != "" && firstErr == nil {
			firstErr = fmt.Errorf("%s", message)
		}
	}
	read := func(m map[string]json.RawMessage, key, bucket string, required bool) int64 {
		b, exists := m[key]
		if !exists {
			if !required {
				return 0
			}
			mark(bucket, "response missing usage")
			return -1
		}
		var n *int64
		if json.Unmarshal(b, &n) != nil || n == nil {
			mark(bucket, "response contained invalid usage")
			return -1
		}
		if *n < 0 {
			mark(bucket, "response contained negative usage")
			return -1
		}
		return *n
	}
	u := common.UnknownUsage()
	switch c.Vendor {
	case common.Anthropic:
		u.Input = read(fields, "input_tokens", "input", true)
		u.Output = read(fields, "output_tokens", "output", true)
		u.CacheRead = read(fields, "cache_read_input_tokens", "cache_read", false)
		u.CacheWrite = read(fields, "cache_creation_input_tokens", "cache_write", false)
	case common.OpenAI:
		inputKey, outputKey, detailsKey := "prompt_tokens", "completion_tokens", "prompt_tokens_details"
		if c.Surface == common.Responses {
			inputKey, outputKey, detailsKey = "input_tokens", "output_tokens", "input_tokens_details"
		}
		total := read(fields, inputKey, "input", true)
		u.Output = read(fields, outputKey, "output", true)
		var details map[string]json.RawMessage
		rawDetails, supplied := fields[detailsKey]
		invalidDetails := supplied && (json.Unmarshal(rawDetails, &details) != nil || details == nil)
		qualifiedWrite := cacheEnabled(c) && (c.Surface == common.ChatCompletions || c.Surface == common.Responses)
		u.CacheWrite = 0 // Older implicit routes do not have a write bucket.
		if invalidDetails {
			// An invalid container cannot be mistaken for omitted optionals.
			mark("cache_read", "response contained invalid usage")
			if qualifiedWrite {
				u.CacheWrite = -1
				mark("cache_write", "response contained invalid usage")
			}
		} else {
			u.CacheRead = read(details, "cached_tokens", "cache_read", false)
			if qualifiedWrite {
				u.CacheWrite = read(details, "cache_write_input_tokens", "cache_write", false)
			} else if _, present := details["cache_write_input_tokens"]; present {
				// A supplied field on an unqualified model is not a guessed fact.
				u.CacheWrite = -1
				mark("cache_write", "")
			}
		}
		if total >= 0 && u.CacheRead >= 0 && u.CacheWrite >= 0 {
			if total < u.CacheRead || total-u.CacheRead < u.CacheWrite {
				mark("input", "response contained inconsistent usage")
			} else {
				u.Input = total - u.CacheRead - u.CacheWrite
			}
		} else {
			mark("input", "")
		}
	case common.Gemini:
		total := read(fields, "promptTokenCount", "input", true)
		u.CacheRead = read(fields, "cachedContentTokenCount", "cache_read", false)
		visible := read(fields, "candidatesTokenCount", "output", true)
		thoughts := int64(0)
		if _, ok := fields["thoughtsTokenCount"]; ok {
			thoughts = read(fields, "thoughtsTokenCount", "output", false)
		}
		if visible >= 0 && thoughts >= 0 {
			u.Output = visible + thoughts
		}
		// This implicit generateContent path does not create explicit caches.
		u.CacheWrite = 0
		if total >= 0 && u.CacheRead >= 0 {
			if total < u.CacheRead {
				mark("input", "response contained inconsistent usage")
			} else {
				u.Input = total - u.CacheRead
			}
		} else {
			mark("input", "")
		}
	}
	return u, missing, firstErr
}

func parseAnthropic(b []byte, from common.Provenance) (common.PartList, error) {
	var response struct {
		Content []json.RawMessage `json:"content"`
	}
	if json.Unmarshal(b, &response) != nil {
		return nil, fmt.Errorf("decoding response failed")
	}
	parts := common.PartList{}
	for _, block := range response.Content {
		var p struct {
			Type  string          `json:"type"`
			Text  *string         `json:"text"`
			ID    string          `json:"id"`
			Name  string          `json:"name"`
			Input json.RawMessage `json:"input"`
		}
		if json.Unmarshal(block, &p) != nil {
			return parts, fmt.Errorf("invalid provider content")
		}
		switch p.Type {
		case "text":
			if p.Text == nil {
				return parts, fmt.Errorf("missing text")
			}
			parts = append(parts, common.Text(*p.Text))
		case "tool_use":
			args, err := canonical(p.Input)
			if err != nil {
				return parts, fmt.Errorf("invalid tool arguments")
			}
			parts = append(parts, common.Part{Type: "tool_call", CallID: p.ID, Name: p.Name, Args: args, From: &from})
		case "thinking", "redacted_thinking":
			parts = append(parts, common.Part{Type: "opaque", From: &from, Data: block})
		}
	}
	return parts, nil
}
func parseOpenAI(b []byte, from common.Provenance) (common.PartList, error) {
	var response struct {
		Choices []struct {
			Message struct {
				Content          *string `json:"content"`
				ToolCalls        []oCall `json:"tool_calls"`
				Refusal          *string `json:"refusal"`
				ReasoningContent *string `json:"reasoning_content"`
				Reasoning        *string `json:"reasoning"`
			} `json:"message"`
		} `json:"choices"`
	}
	if json.Unmarshal(b, &response) != nil {
		return nil, fmt.Errorf("decoding response failed")
	}
	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("missing choice")
	}
	message := response.Choices[0].Message
	parts := common.PartList{}
	if message.Content != nil {
		parts = append(parts, common.Text(*message.Content))
	}
	if message.Refusal != nil && *message.Refusal != "" {
		parts = append(parts, common.Text(*message.Refusal))
	}
	// Reasoning is provider replay material, not visible answer text.
	for _, field := range []struct {
		name string
		text *string
	}{{"reasoning_content", message.ReasoningContent}, {"reasoning", message.Reasoning}} {
		if field.text != nil {
			raw, _ := json.Marshal(map[string]string{field.name: *field.text})
			parts = append(parts, common.Part{Type: "opaque", From: &from, Data: raw})
		}
	}
	for _, call := range message.ToolCalls {
		args, err := canonical(json.RawMessage(call.Function.Arguments))
		if err != nil {
			return parts, fmt.Errorf("invalid tool arguments")
		}
		parts = append(parts, common.Part{Type: "tool_call", CallID: call.ID, Name: call.Function.Name, Args: args, From: &from})
	}
	return parts, nil
}
func parseGemini(b []byte, from common.Provenance) (common.PartList, error) {
	var response struct {
		Candidates []struct {
			Content struct {
				Parts []json.RawMessage `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if json.Unmarshal(b, &response) != nil {
		return nil, fmt.Errorf("decoding response failed")
	}
	if len(response.Candidates) == 0 {
		return nil, fmt.Errorf("missing candidate")
	}
	parts := common.PartList{}
	for _, block := range response.Candidates[0].Content.Parts {
		var p struct {
			Text         *string `json:"text"`
			Thought      bool    `json:"thought"`
			FunctionCall *struct {
				ID   string          `json:"id"`
				Name string          `json:"name"`
				Args json.RawMessage `json:"args"`
			} `json:"functionCall"`
			Signature json.RawMessage `json:"thoughtSignature"`
		}
		if json.Unmarshal(block, &p) != nil {
			return parts, fmt.Errorf("invalid provider content")
		}
		if p.Thought {
			parts = append(parts, common.Part{Type: "opaque", From: &from, Data: block})
		} else if p.FunctionCall != nil {
			args, err := canonical(p.FunctionCall.Args)
			if err != nil {
				return parts, fmt.Errorf("invalid tool arguments")
			}
			parts = append(parts, common.Part{Type: "tool_call", CallID: p.FunctionCall.ID, Name: p.FunctionCall.Name, Args: args, From: &from, Opaque: p.Signature})

		} else if p.Text != nil {
			part := common.Text(*p.Text)
			if len(p.Signature) > 0 {
				part.From = &from
				part.Opaque = p.Signature
			}
			parts = append(parts, part)
		} else {
			parts = append(parts, common.Part{Type: "opaque", From: &from, Data: block})
		}
	}
	return parts, nil
}
