package engine

import (
	"encoding/json"
	"errors"
	"io"
	"sort"

	"ensemble/internal/common"
)

func (e *engine) readOpenAI(r io.Reader, log func(string, []byte), delta func(uint64, common.DeltaKind, string)) ([]byte, []uint64, error) {
	var model, content string
	var finishReason string
	var usage json.RawMessage
	calls := map[int]*openAICall{}
	hasText, finished := false, false
	err := frames(r, log, func(_ string, data []byte) error {
		var frame struct {
			// Model reports the resolved producer identity.
			Model string `json:"model"`
			// Usage arrives on the final usage-only chunk when explicitly requested.
			Usage json.RawMessage `json:"usage"`
			// Error distinguishes a provider stream failure from an empty chunk.
			Error json.RawMessage `json:"error"`
			// Choices contains incremental first-choice content; usage frames have none.
			Choices []struct {
				// FinishReason marks completion, including tool-call completion.
				FinishReason *string `json:"finish_reason"`
				// Delta holds text and indexed tool fragments, never executable arguments.
				Delta struct {
					// Content distinguishes no text from a present empty text fragment.
					Content *string `json:"content"`
					// ToolCalls groups fragments by their vendor call index.
					ToolCalls []struct {
						// Index is stable within this response, even when text arrives later.
						Index int `json:"index"`
						openAICall
					} `json:"tool_calls"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if json.Unmarshal(data, &frame) != nil {
			return errors.New("invalid OpenAI stream frame")
		}
		if len(frame.Error) > 0 {
			return errors.New("OpenAI stream error")
		}
		if frame.Model != "" {
			model = frame.Model
		}
		if len(frame.Usage) > 0 && string(frame.Usage) != "null" {
			usage = frame.Usage
		}
		if len(frame.Choices) == 0 {
			return nil
		}
		choice := frame.Choices[0]
		if choice.FinishReason != nil {
			finished = true
			finishReason = *choice.FinishReason
		}
		if choice.Delta.Content != nil {
			hasText = true
			content += *choice.Delta.Content
			// Reserve zero for text from the start. Tool indices occupy a distinct
			// namespace, so a tool-first stream cannot rename a part when text appears.
			delta(0, common.DeltaText, *choice.Delta.Content)
		}
		for _, fragment := range choice.Delta.ToolCalls {
			call := calls[fragment.Index]
			if call == nil {
				call = &openAICall{Type: "function"}
				calls[fragment.Index] = call
			}
			call.ID += fragment.ID
			call.Function.Name += fragment.Function.Name
			call.Function.Arguments += fragment.Function.Arguments
			id := uint64(fragment.Index) + 1
			if fragment.Function.Name != "" {
				delta(id, common.DeltaToolCall, fragment.Function.Name+" ")
			}
			if fragment.Function.Arguments != "" {
				delta(id, common.DeltaToolCall, fragment.Function.Arguments)
			}
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	if !finished {
		return nil, nil, errors.New("incomplete OpenAI stream")
	}
	message := map[string]any{}
	var ids []uint64
	if hasText {
		message["content"] = content
		ids = append(ids, 0)
	}
	indices := make([]int, 0, len(calls))
	for i := range calls {
		indices = append(indices, i)
	}
	sort.Ints(indices)
	var ordered []*openAICall
	for _, i := range indices {
		ordered = append(ordered, calls[i])
		ids = append(ids, uint64(i)+1)
	}
	if len(ordered) > 0 {
		message["tool_calls"] = ordered
	}
	data, err := json.Marshal(map[string]any{"model": model, "usage": usage, "choices": []any{map[string]any{"message": message, "finish_reason": finishReason}}})
	return data, ids, err
}
