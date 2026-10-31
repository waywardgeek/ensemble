package engine

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"ensemble/internal/common"
)

// The completed output is authoritative. Deltas have response-local identities
// assigned as they arrive; keys reconnect those identities to the final parts.
type responseItem struct {
	// Type distinguishes message, reasoning and function-call items.
	Type string `json:"type"`
	// CallID correlates local tool execution with the following input item.
	CallID string `json:"call_id"`
	// Namespace preserves the provider grouping independently of the function name.
	Namespace string `json:"namespace"`
	// Name selects the caller-visible function.
	Name string `json:"name"`
	// Arguments contains the completed JSON string.
	Arguments string `json:"arguments"`
	// Content retains message content in provider order.
	Content []struct {
		// Type separates ordinary output text from a refusal.
		Type string `json:"type"`
		// Text is the assistant's visible answer.
		Text string `json:"text"`
		// Refusal is a user-visible refusal, never an inference transport failure.
		Refusal string `json:"refusal"`
	} `json:"content"`
}

type responsesResult struct {
	// Status must be completed; an HTTP success or partial stream is insufficient.
	Status string `json:"status"`
	// Model retains actual provider identity for opaque replay.
	Model string `json:"model"`
	// Output retains raw reasoning items while text and calls become neutral parts.
	Output []json.RawMessage `json:"output"`
	// Usage retains presence separately from measured zero counts.
	Usage *struct {
		// Input includes cached and cache-write input exactly once.
		Input *int `json:"input_tokens"`
		// Output includes reasoning output; never add reasoning_tokens again.
		Output *int `json:"output_tokens"`
		// Details separates the inclusive input count into disjoint categories.
		Details struct {
			// Read counts reused cache input.
			Read int `json:"cached_tokens"`
			// Write counts newly cached input.
			Write int `json:"cache_write_tokens"`
		} `json:"input_tokens_details"`
	} `json:"usage"`
}

// Only complete provider measurements may enter the meter; absent counts do
// not become an invented zero-cost result when a response fails.
func (r responsesResult) measuredUsage() (common.Usage, bool) {
	if r.Usage == nil || r.Usage.Input == nil || r.Usage.Output == nil {
		return common.Usage{}, false
	}
	u := r.Usage
	return common.Usage{Input: *u.Input - u.Details.Read - u.Details.Write, CacheRead: u.Details.Read, CacheWrite: u.Details.Write, Output: *u.Output}, true
}

func (e *engine) accountFailedResponse(result responsesResult, cfg common.Config) {
	if usage, ok := result.measuredUsage(); ok && e.parent.Spec().Compaction {
		e.account(cfg, usage)
	}
}

func (e *engine) parseResponses(resp *http.Response, cb common.StreamCallbacks, cfg common.Config) (err error) {
	ids := map[string]uint64{}
	seen := map[string]bool{}
	identify := func(key string) uint64 {
		if id, ok := ids[key]; ok {
			return id
		}
		id := uint64(len(ids))
		ids[key] = id
		return id
	}
	emit := func(key string, kind common.DeltaKind, value string) {
		seen[key] = true
		id := identify(key)
		if cb.OnDelta != nil {
			cb.OnDelta(id, kind, value)
		}
	}
	var result responsesResult
	completed := false
	// The observed plan response carried SSE bytes despite not selecting the
	// header-based branch. Peek without consuming or buffering the whole stream:
	// progress must still arrive before the provider finishes the response.
	body := bufio.NewReader(resp.Body)
	prefix, _ := body.Peek(6)
	streamed := strings.HasPrefix(resp.Header.Get("Content-Type"), "text/event-stream") || bytes.HasPrefix(prefix, []byte("event:")) || bytes.HasPrefix(prefix, []byte("data:"))
	if streamed {
		items := map[int]json.RawMessage{}
		err := frames(body, cb.OnFrame, func(event string, data []byte) error {
			var frame struct {
				// Type is also supplied in the SSE event name by supported servers.
				Type string `json:"type"`
				// Index identifies an output item independently of arrival order.
				Index int `json:"output_index"`
				// Content identifies text within a message item.
				Content int `json:"content_index"`
				// Delta is an incremental public display fragment.
				Delta string `json:"delta"`
				// Item supplies the function name before argument deltas arrive.
				Item json.RawMessage `json:"item"`
				// Response is authoritative only in response.completed.
				Response responsesResult `json:"response"`
			}
			if json.Unmarshal(data, &frame) != nil {
				return errors.New("invalid Responses stream frame")
			}
			if frame.Type != "" {
				event = frame.Type
			}
			key := fmt.Sprintf("%d", frame.Index)
			switch event {
			case "response.reasoning_summary_text.delta":
				emit(key, common.DeltaReasoningSummary, frame.Delta)
			case "response.output_text.delta", "response.refusal.delta":
				emit(fmt.Sprintf("%d/%d", frame.Index, frame.Content), common.DeltaText, frame.Delta)
			case "response.output_item.added":
				var item responseItem
				json.Unmarshal(frame.Item, &item)
				if item.Type == "function_call" {
					emit(key, common.DeltaToolCall, item.Name+" ")
				}
			case "response.output_item.done":
				items[frame.Index] = frame.Item
			case "response.function_call_arguments.delta":
				emit(key, common.DeltaToolCall, frame.Delta)
			case "response.completed":
				result = frame.Response
				completed = true
			case "response.failed", "response.incomplete", "error":
				e.accountFailedResponse(frame.Response, cfg)
				return providerError(0, data)
			}
			return nil
		})
		if err != nil {
			return err
		}
		// The plan stream completes each item separately and can leave terminal
		// output empty. These completed items are authoritative, but cannot become
		// conversation facts or tools until terminal success passes below.
		if len(result.Output) == 0 {
			indices := make([]int, 0, len(items))
			for index := range items {
				indices = append(indices, index)
			}
			sort.Ints(indices)
			for _, index := range indices {
				result.Output = append(result.Output, items[index])
			}
		}
	} else {
		data, err := io.ReadAll(body)
		if err != nil {
			return errors.New("reading Responses response failed")
		}
		if cb.OnFrame != nil {
			cb.OnFrame("response", data)
		}
		if json.Unmarshal(data, &result) != nil {
			return errors.New("invalid Responses response")
		}
		completed = result.Status == "completed"
	}
	if !completed || result.Status != "completed" {
		e.accountFailedResponse(result, cfg)
		return errors.New("Responses stream ended without response.completed")
	}
	usage, measured := result.measuredUsage()
	if !measured {
		return errors.New("Responses response missing usage")
	}
	// A complete envelope can still contain malformed function arguments.
	// Retain its measured usage without publishing a successful response fact.
	defer func() {
		if err != nil {
			e.accountFailedResponse(result, cfg)
		}
	}()
	from := common.Provenance{Vendor: common.OpenAI, Model: cfg.Model, Surface: common.Responses}
	if cfg.ResolvedModel != "" {
		from.Model = cfg.ResolvedModel
	}
	if result.Model != "" {
		from.Model = result.Model
	}
	out := common.ResponseData{From: from, Usage: usage}
	var partIDs []uint64
	add := func(key string, p common.Part, kind common.DeltaKind, value string) {
		id := identify(key)
		out.Parts = append(out.Parts, p)
		partIDs = append(partIDs, id)
		if !seen[key] && kind != 0 {
			emit(key, kind, value)
		}
	}
	for i, raw := range result.Output {
		var item responseItem
		if json.Unmarshal(raw, &item) != nil {
			return errors.New("invalid Responses output item")
		}
		key := fmt.Sprintf("%d", i)
		switch item.Type {
		case "message":
			for j, part := range item.Content {
				value := part.Text
				if part.Type == "refusal" {
					value = part.Refusal
				} else if part.Type != "output_text" {
					return errors.New("unsupported Responses message content")
				}
				add(fmt.Sprintf("%d/%d", i, j), common.Part{Type: "text", Text: value}, common.DeltaText, value)
			}
		case "function_call":
			args, err := canonical(json.RawMessage(item.Arguments))
			if err != nil {
				return err
			}
			add(key, common.Part{Type: "tool_call", CallID: item.CallID, Name: item.Name, Namespace: item.Namespace, Args: args, From: from}, common.DeltaToolCall, item.Name+" "+string(args))
		case "reasoning":
			add(key, common.Part{Type: "opaque", Data: raw, From: from}, common.DeltaReasoningSummary, reasoningSummary(raw))
		default:
			return errors.New("unsupported Responses output item")
		}
	}
	if len(out.Parts) == 0 {
		return errors.New("Responses response has no content")
	}
	if cb.OnEvent != nil {
		cb.OnEvent(common.Event{Type: "response_ended", Response: &out})
	}
	if cb.OnPartFinal != nil {
		for i, p := range out.Parts {
			cb.OnPartFinal(partIDs[i], p)
		}
	}
	return nil
}

// Summaries are intentionally visible; the encrypted field alongside them is not.
func reasoningSummary(raw json.RawMessage) string {
	var item struct {
		// Summary contains only provider-designated public explanation.
		Summary []struct {
			// Text is readable summary content, never encrypted reasoning.
			Text string `json:"text"`
		} `json:"summary"`
	}
	json.Unmarshal(raw, &item)
	var text strings.Builder
	for _, part := range item.Summary {
		text.WriteString(part.Text)
	}
	return text.String()
}
