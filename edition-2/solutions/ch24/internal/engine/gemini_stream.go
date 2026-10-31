package engine

import (
	"encoding/json"
	"errors"
	"io"

	"ensemble/internal/common"
)

func (e *engine) readGemini(r io.Reader, log func(string, []byte), delta func(uint64, common.DeltaKind, string)) ([]byte, []uint64, error) {
	var model string
	var finishReason string
	var usage json.RawMessage
	var parts []geminiPart
	finished := false
	err := frames(r, log, func(_ string, data []byte) error {
		var frame struct {
			// ModelVersion identifies the actual producer for opaque replay.
			ModelVersion string `json:"modelVersion"`
			// UsageMetadata is cumulative; retain the latest complete measurement.
			UsageMetadata json.RawMessage `json:"usageMetadata"`
			// Error reports a failed stream without exposing remote error prose.
			Error json.RawMessage `json:"error"`
			// Candidates contains this frame's new content, not all content so far.
			Candidates []struct {
				// Content carries new text or complete function calls in provider order.
				Content geminiContent `json:"content"`
				// FinishReason signals a completed response even for function calls.
				FinishReason string `json:"finishReason"`
			} `json:"candidates"`
		}
		if json.Unmarshal(data, &frame) != nil {
			return errors.New("invalid Gemini stream frame")
		}
		if len(frame.Error) > 0 {
			return errors.New("Gemini stream error")
		}
		if frame.ModelVersion != "" {
			model = frame.ModelVersion
		}
		if len(frame.UsageMetadata) > 0 {
			usage = frame.UsageMetadata
		}
		if len(frame.Candidates) == 0 {
			return nil
		}
		c := frame.Candidates[0]
		if c.FinishReason != "" {
			finished = true
			finishReason = c.FinishReason
		}
		for i, p := range c.Content.Parts {
			id := len(parts)
			// There is no block index. A frame's leading text continues the previous
			// text only when the thought flag agrees. A second part starts a new part.
			if i == 0 && len(parts) > 0 && p.Text != nil {
				last := &parts[len(parts)-1]
				if last.Text != nil && last.Thought == p.Thought {
					id--
					text := *last.Text + *p.Text
					last.Text = &text
					if len(p.ThoughtSignature) > 0 {
						last.ThoughtSignature = p.ThoughtSignature
					}
				}
			}
			if id == len(parts) {
				parts = append(parts, p)
			}
			switch {
			case p.FunctionCall != nil:
				// Gemini delivers a complete object. Emit one honest tool delta instead
				// of manufacturing incremental JSON from a response already in memory.
				args, err := canonical(p.FunctionCall.Args)
				if err != nil {
					return err
				}
				delta(uint64(id), common.DeltaToolCall, p.FunctionCall.Name+" "+string(args))
			case p.Text != nil:
				kind := common.DeltaText
				if p.Thought {
					kind = common.DeltaThinking
				}
				delta(uint64(id), kind, *p.Text)
			}
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	if !finished {
		return nil, nil, errors.New("incomplete Gemini stream")
	}
	ids := make([]uint64, len(parts))
	for i := range ids {
		ids[i] = uint64(i)
	}
	data, err := json.Marshal(map[string]any{"modelVersion": model, "usageMetadata": usage, "candidates": []any{map[string]any{"content": geminiContent{Role: "model", Parts: parts}, "finishReason": finishReason}}})
	return data, ids, err
}
