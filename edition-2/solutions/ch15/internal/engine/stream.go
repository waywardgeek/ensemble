package engine

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"ensemble/internal/common"
)

// Parse owns IDs and finalization in both delivery modes. The SSE accumulators
// reconstruct the provider message, then reuse the existing semantic decoder:
// accounting, provenance and opaque replay never grow a second implementation.
func (e *engine) Parse(resp *http.Response, cb common.StreamCallbacks) error {
	var data []byte
	var ids []uint64
	var err error
	streamed := strings.HasPrefix(resp.Header.Get("Content-Type"), "text/event-stream")
	seen := map[uint64]bool{}
	delta := func(id uint64, kind common.DeltaKind, chunk string) {
		// Record even an empty delta: present-but-empty text is still a part.
		seen[id] = true
		if cb.OnDelta != nil {
			cb.OnDelta(id, kind, chunk)
		}
	}
	if streamed {
		switch e.parent.Config().Vendor {
		case common.Anthropic:
			data, ids, err = e.readAnthropic(resp.Body, cb.OnFrame, delta)
		case common.OpenAI:
			data, ids, err = e.readOpenAI(resp.Body, cb.OnFrame, delta)
		case common.Gemini:
			data, ids, err = e.readGemini(resp.Body, cb.OnFrame, delta)
		}
	} else {
		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return errors.New("reading provider response failed")
		}
		if cb.OnFrame != nil {
			cb.OnFrame("response", data)
		}
	}
	if err != nil {
		return err
	}
	var response common.ResponseData
	switch e.parent.Config().Vendor {
	case common.Anthropic:
		response, err = e.parseAnthropic(data)
	case common.OpenAI:
		response, err = e.parseOpenAI(data)
	case common.Gemini:
		response, err = e.parseGemini(data)
	}
	if err != nil {
		return err
	}
	if ids == nil {
		for i := range response.Parts {
			ids = append(ids, uint64(i))
		}
	}
	if len(ids) != len(response.Parts) {
		return errors.New("stream part identity mismatch")
	}
	for i, p := range response.Parts {
		if !seen[ids[i]] {
			kind, chunk := displayPart(p)
			if kind != 0 {
				delta(ids[i], kind, chunk)
			}
		}
	}
	// Facts precede finals; a consumer can replace a streamed approximation only
	// after its authoritative complete content has entered the ordered log path.
	if cb.OnEvent != nil {
		cb.OnEvent(common.Event{Type: "response_ended", Response: &response})
	}
	if cb.OnPartFinal != nil {
		for i, p := range response.Parts {
			cb.OnPartFinal(ids[i], p)
		}
	}
	return nil
}

// Opaque thinking keeps its exact vendor block for replay; only its documented
// readable field reaches the thinking display. Signatures are never prose.
func displayPart(p common.Part) (common.DeltaKind, string) {
	switch p.Type {
	case "text":
		return common.DeltaText, p.Text
	case "tool_call":
		return common.DeltaToolCall, p.Name + " " + string(p.Args)
	case "opaque":
		var fields struct {
			// Thinking is Anthropic's readable reasoning, separate from its signature.
			Thinking string `json:"thinking"`
			// Text is Gemini's readable reasoning when the part is marked thought.
			Text string `json:"text"`
			// Thought distinguishes Gemini reasoning from other opaque replay material.
			Thought bool `json:"thought"`
			// Type distinguishes Anthropic thinking from redacted thinking.
			Type string `json:"type"`
		}
		json.Unmarshal(p.Data, &fields)
		if fields.Type == "thinking" {
			return common.DeltaThinking, fields.Thinking
		}
		if fields.Thought {
			return common.DeltaThinking, fields.Text
		}
	}
	return 0, ""
}

// Frame logging happens before interpretation, including terminators and errors.
// An error payload is useful in the sanitized trace but is not echoed as an error.
func frames(r io.Reader, log func(string, []byte), consume func(string, []byte) error) error {
	return ReadSSE(r, func(kind string, data []byte) error {
		if log != nil {
			log(kind, data)
		}
		if string(data) == "[DONE]" {
			return nil
		}
		return consume(kind, data)
	})
}
