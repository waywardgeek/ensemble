package engine

import (
	"encoding/json"
	"errors"
	"io"
	"sort"

	"ensemble/internal/common"
)

// A streamed block retains unknown signature fields as raw JSON. Reconstructing
// only its visible text would make a plausible display and an invalid replay.
func (e *engine) readAnthropic(r io.Reader, log func(string, []byte), delta func(uint64, common.DeltaKind, string)) ([]byte, []uint64, error) {
	message := map[string]json.RawMessage{}
	blocks := map[int]map[string]json.RawMessage{}
	arguments := map[int]string{}
	stopped := false
	err := frames(r, log, func(_ string, data []byte) error {
		var frame struct {
			// Type selects the Messages stream event.
			Type string `json:"type"`
			// Index is the provider's response-scoped content-block identity.
			Index int `json:"index"`
			// Message starts the response with its model and initial usage.
			Message map[string]json.RawMessage `json:"message"`
			// Block contains the initial content variant and private replay fields.
			Block map[string]json.RawMessage `json:"content_block"`
			// Delta contains an incremental field or a top-level message update.
			Delta map[string]json.RawMessage `json:"delta"`
			// Usage updates cumulative counts from message_start.
			Usage map[string]json.RawMessage `json:"usage"`
		}
		if json.Unmarshal(data, &frame) != nil {
			return errors.New("invalid Anthropic stream frame")
		}
		id := uint64(frame.Index)
		switch frame.Type {
		case "error":
			return errors.New("Anthropic stream error")
		case "message_start":
			message = frame.Message
		case "content_block_start":
			blocks[frame.Index] = frame.Block
			switch rawString(frame.Block["type"]) {
			case "text":
				if s := rawString(frame.Block["text"]); s != "" {
					delta(id, common.DeltaText, s)
				}
			case "thinking":
				if s := rawString(frame.Block["thinking"]); s != "" {
					delta(id, common.DeltaThinking, s)
				}
			case "tool_use":
				delta(id, common.DeltaToolCall, rawString(frame.Block["name"])+" ")
			}
		case "content_block_delta":
			block := blocks[frame.Index]
			if block == nil {
				return errors.New("Anthropic delta without block")
			}
			var field string
			var kind common.DeltaKind
			switch rawString(frame.Delta["type"]) {
			case "text_delta":
				field, kind = "text", common.DeltaText
			case "thinking_delta":
				field, kind = "thinking", common.DeltaThinking
			case "signature_delta":
				field = "signature"
			case "input_json_delta":
				chunk := rawString(frame.Delta["partial_json"])
				arguments[frame.Index] += chunk
				delta(id, common.DeltaToolCall, chunk)
			}
			if field != "" {
				chunk := rawString(frame.Delta[field])
				block[field], _ = json.Marshal(rawString(block[field]) + chunk)
				if kind != 0 {
					delta(id, kind, chunk)
				}
			}
		case "content_block_stop":
			block := blocks[frame.Index]
			if block == nil {
				return errors.New("Anthropic stop without block")
			}
			if rawString(block["type"]) == "tool_use" {
				if args, ok := arguments[frame.Index]; ok {
					block["input"] = json.RawMessage(args)
				} else {
					delta(id, common.DeltaToolCall, string(block["input"]))
				}
			}
		case "message_delta":
			var usage map[string]json.RawMessage
			json.Unmarshal(message["usage"], &usage)
			if usage == nil {
				usage = map[string]json.RawMessage{}
			}
			for key, value := range frame.Usage {
				usage[key] = value
			}
			message["usage"], _ = json.Marshal(usage)
		case "message_stop":
			stopped = true
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	if !stopped {
		return nil, nil, errors.New("incomplete Anthropic stream")
	}
	indices := make([]int, 0, len(blocks))
	for i := range blocks {
		indices = append(indices, i)
	}
	sort.Ints(indices)
	var content []map[string]json.RawMessage
	var ids []uint64
	for _, i := range indices {
		content = append(content, blocks[i])
		ids = append(ids, uint64(i))
	}
	message["content"], err = json.Marshal(content)
	if err != nil {
		return nil, nil, errors.New("invalid streamed Anthropic content")
	}
	data, err := json.Marshal(message)
	return data, ids, err
}

func rawString(raw json.RawMessage) string { var s string; json.Unmarshal(raw, &s); return s }
