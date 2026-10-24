package engine

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"ensemble/internal/common"
)

// Sniff only the first SSE field prefix, not a complete frame or response.
// JSON whitespace may precede either representation. The existing parsers still
// own syntax, terminal completion and usage validation; bytes after whitespace
// remain buffered for them, including when the response is ordinary JSON.
func sniffResponsesSSE(r *bufio.Reader) bool {
	var first byte
	for {
		p, err := r.Peek(1)
		if err != nil {
			return false
		}
		first = p[0]
		if first != ' ' && first != '\t' && first != '\r' && first != '\n' {
			break
		}
		_, _ = r.ReadByte()
	}
	var prefix string
	switch first {
	case ':':
		return true
	case 'd':
		prefix = "data:"
	case 'e':
		prefix = "event:"
	case 'i':
		prefix = "id:"
	case 'r':
		prefix = "retry:"
	default:
		return false
	}
	p, err := r.Peek(len(prefix))
	return err == nil && string(p) == prefix
}

// Responses has explicit output/content indices and an authoritative completed
// response. Only public deltas are emitted; encrypted_content never leaves this
// adapter as a delta, log field, or GUI placeholder payload.
func parseResponsesSSE(body io.Reader, c common.Config, cb common.ParseCallbacks) ([]common.Event, error) {
	ids := map[string]uint64{}
	texts := map[string]string{}
	kinds := map[string]common.DeltaKind{}
	names := map[string]string{}
	arguments := map[string]string{}
	callIDs := map[string]string{}
	completed := map[int]json.RawMessage{}
	var next uint64
	model := ""
	usage := json.RawMessage(nil)
	fail := func(message string) ([]common.Event, error) {
		u, missing, _ := parseUsage(c, map[string]json.RawMessage{"usage": usage})
		actualModel := model
		if actualModel == "" {
			actualModel = c.Model
		}
		from := common.Provenance{Vendor: c.Vendor, Surface: c.Surface, Model: actualModel}
		return []common.Event{{Type: common.ResponseStarted, Response: &common.ResponseData{From: from}}, {Type: common.ResponseEnded, Response: &common.ResponseData{From: from, Usage: u, UsageMissing: missing}}, errorEvent(message)}, fmt.Errorf("%s", message)
	}
	idFor := func(key string) uint64 {
		if id, ok := ids[key]; ok {
			return id
		}
		id := next
		next++
		ids[key] = id
		return id
	}
	delta := func(key string, kind common.DeltaKind, text, name string) {
		_, exists := ids[key]
		id := idFor(key)
		kinds[key] = kind
		texts[key] += text
		if cb.OnDelta != nil && (text != "" || !exists) {
			cb.OnDelta(common.PartDelta{Kind: kind, PartID: id, Chunk: text, Name: name})
		}
	}
	process := func(data string) ([]common.Event, error, bool) {
		var ev struct {
			Type         string          `json:"type"`
			OutputIndex  int             `json:"output_index"`
			ContentIndex int             `json:"content_index"`
			SummaryIndex int             `json:"summary_index"`
			Delta        string          `json:"delta"`
			Item         json.RawMessage `json:"item"`
			Part         responseContent `json:"part"`
			Response     json.RawMessage `json:"response"`
		}
		if json.Unmarshal([]byte(data), &ev) != nil {
			return nil, fmt.Errorf("invalid Responses stream event"), true
		}
		if len(ev.Response) > 0 {
			var envelope map[string]json.RawMessage
			if json.Unmarshal(ev.Response, &envelope) != nil {
				return nil, fmt.Errorf("invalid Responses stream response"), true
			}
			if name := str(envelope["model"]); name != "" {
				if model != "" && model != name {
					return nil, fmt.Errorf("Responses model changed during stream"), true
				}
				model = name
			}
			if u, ok := envelope["usage"]; ok && string(u) != "null" {
				usage = u
			}
		}
		if ev.OutputIndex < 0 || ev.ContentIndex < 0 || ev.SummaryIndex < 0 {
			return nil, fmt.Errorf("invalid Responses stream index"), true
		}
		switch ev.Type {
		case "response.output_item.done":
			if ev.OutputIndex < 0 || len(ev.Item) == 0 || string(ev.Item) == "null" {
				return nil, fmt.Errorf("invalid completed Responses item"), true
			}
			if _, exists := completed[ev.OutputIndex]; exists {
				return nil, fmt.Errorf("duplicate completed Responses output index"), true
			}
			completed[ev.OutputIndex] = append(json.RawMessage(nil), ev.Item...)
		case "response.output_item.added":
			var item responseItem
			if json.Unmarshal(ev.Item, &item) != nil {
				return nil, fmt.Errorf("invalid Responses item"), true
			}
			if item.Type == "function_call" {
				key := responseKey(ev.OutputIndex, "tool", 0)
				if _, exists := names[key]; exists {
					return nil, fmt.Errorf("duplicate Responses tool start"), true
				}
				names[key] = item.Name
				callIDs[key] = item.CallID
				arguments[key] = item.Arguments
				delta(key, common.DeltaToolCall, item.Name+item.Arguments, item.Name)
			}
		case "response.content_part.added":
			if ev.Part.Type == "output_text" || ev.Part.Type == "refusal" {
				delta(responseKey(ev.OutputIndex, "text", ev.ContentIndex), common.DeltaText, "", "")
			}
		case "response.reasoning_summary_part.added":
			delta(responseKey(ev.OutputIndex, "summary", ev.SummaryIndex), common.DeltaReasoningSummary, "", "")
		case "response.output_text.delta", "response.refusal.delta":
			delta(responseKey(ev.OutputIndex, "text", ev.ContentIndex), common.DeltaText, ev.Delta, "")
		case "response.reasoning_summary_text.delta":
			delta(responseKey(ev.OutputIndex, "summary", ev.SummaryIndex), common.DeltaReasoningSummary, ev.Delta, "")
		case "response.function_call_arguments.delta":
			key := responseKey(ev.OutputIndex, "tool", 0)
			arguments[key] += ev.Delta
			delta(key, common.DeltaToolCall, ev.Delta, "")
		case "response.completed":
			var terminal map[string]json.RawMessage
			if json.Unmarshal(ev.Response, &terminal) != nil {
				return nil, fmt.Errorf("invalid Responses completion"), true
			}
			var output []json.RawMessage
			if rawOutput := terminal["output"]; len(rawOutput) > 0 {
				if json.Unmarshal(rawOutput, &output) != nil {
					return nil, fmt.Errorf("invalid terminal Responses output"), true
				}
			}
			var outputIndices []int
			if len(output) == 0 && len(completed) > 0 {
				// Empty terminal output is a protocol compact form, not an
				// instruction to discard completed items. Native indices keep
				// final part identity tied to the original streaming deltas.
				for index := range completed {
					outputIndices = append(outputIndices, index)
				}
				sort.Ints(outputIndices)
				for _, index := range outputIndices {
					output = append(output, completed[index])
				}
				terminal["output"], _ = json.Marshal(output)
				ev.Response, _ = json.Marshal(terminal)
			}
			events, err := parse(200, ev.Response, c)
			if err != nil {
				return nil, err, true
			}
			from := common.Provenance{Vendor: c.Vendor, Surface: c.Surface, Model: model}
			parts, keys, err := responsesParts(ev.Response, from, outputIndices)
			if err != nil {
				return nil, err, true
			}
			present := map[string]bool{}
			for i, key := range keys {
				p := &parts[i]
				p.ID = idFor(key)
				present[key] = true
				if kinds[key] == common.DeltaToolCall {
					if (names[key] != "" && names[key] != p.Name) || (callIDs[key] != "" && callIDs[key] != p.CallID) || !strings.HasPrefix(string(p.Args), arguments[key]) {
						return nil, fmt.Errorf("Responses terminal tool disagrees with stream"), true
					}
				}
				if kinds[key] == common.DeltaText || kinds[key] == common.DeltaReasoningSummary {
					if !strings.HasPrefix(p.Text, texts[key]) {
						return nil, fmt.Errorf("Responses terminal text disagrees with stream"), true
					}
				}
			}
			for key := range kinds {
				if !present[key] {
					return nil, fmt.Errorf("Responses terminal parts disagree with stream"), true
				}
			}
			for i := range events {
				if events[i].Response != nil {
					events[i].Response.From = from
				}
				if events[i].Type == common.ResponseEnded {
					events[i].Response.Parts = parts
				}
			}
			return events, nil, true
		case "response.failed", "response.incomplete", "error":
			return nil, fmt.Errorf("Responses request did not complete"), true
		}
		return nil, nil, false
	}
	reader := sseReader{r: bufio.NewReader(body)}
	for {
		frame, err := reader.next()
		if err != nil {
			if err == io.EOF {
				return fail("Responses stream ended before completion")
			}
			return fail("reading Responses stream failed")
		}
		events, err, done := process(frame.data)
		if done {
			if events != nil {
				return events, err
			}
			return fail(err.Error())
		}
	}
}
