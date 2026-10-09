package engine

import (
	"encoding/json"
	"errors"

	"ensemble/internal/common"
)

type geminiCall struct {
	ID   string          `json:"id,omitempty"`
	Name string          `json:"name"`
	Args json.RawMessage `json:"args"`
}
type geminiResult struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name"`
	Response struct {
		Result  string `json:"result"`
		IsError bool   `json:"is_error,omitempty"`
	} `json:"response"`
}

// A thoughtSignature is a sibling of functionCall on the wire. Keeping it
// beside the call while parsing and rendering prevents an unrelated opaque
// part from being mistaken for that call's mandatory replay material.
type geminiPart struct {
	Text             *string         `json:"text,omitempty"`
	FunctionCall     *geminiCall     `json:"functionCall,omitempty"`
	FunctionResponse *geminiResult   `json:"functionResponse,omitempty"`
	Thought          bool            `json:"thought,omitempty"`
	ThoughtSignature json.RawMessage `json:"thoughtSignature,omitempty"`
}
type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

// Gemini calls the assistant role model, but still renders tool testimony as
// user content. systemInstruction is a Content object outside contents;
// putting a system role inside contents would be an invalid request.
func (e *engine) renderGemini(c *common.Context) ([]byte, error) {
	var contents []geminiContent
	for _, entry := range entries(c) {
		role := "user"
		if entry.Actor == common.AgentActor {
			role = "model"
		}
		content := geminiContent{Role: role}
		for _, p := range entry.Parts {
			if err := e.refuseBlob(p); err != nil {
				return nil, err
			}
			part := geminiPart{}
			switch p.Type {
			case "text":
				value := p.Text
				part.Text = &value
			case "redacted":
				value := p.Stub
				part.Text = &value
			case "tool_call":
				part.FunctionCall = &geminiCall{ID: p.CallID, Name: p.Name, Args: p.Args}
				// A call-bound signature cannot be moved to a neighboring part. Only its
				// exact producer gets it back, including the API surface and model ID.
				if p.From == e.target() {
					part.ThoughtSignature = p.Opaque
				}
			case "tool_result":
				_, _, call, err := callFor(c, p.CallID)
				if err != nil {
					return nil, err
				}
				for _, nested := range p.Parts {
					if err := e.refuseBlob(nested); err != nil {
						return nil, err
					}
				}
				// response must be an object, never a bare string. Name is resolved
				// from the producing call rather than duplicated in the log.
				part.FunctionResponse = &geminiResult{ID: p.CallID, Name: call.Name}
				part.FunctionResponse.Response.Result = text(p.Parts)
				part.FunctionResponse.Response.IsError = p.IsError
			case "opaque":
				if p.From != e.target() {
					continue
				}
				if err := json.Unmarshal(p.Data, &part); err != nil {
					return nil, errors.New("invalid Gemini replay part")
				}
			default:
				return nil, errors.New("unsupported Gemini part")
			}
			content.Parts = append(content.Parts, part)
		}
		if len(content.Parts) > 0 {
			contents = append(contents, content)
		}
	}
	prompt := systemPrompt
	return json.Marshal(struct {
		SystemInstruction geminiContent   `json:"systemInstruction"`
		Contents          []geminiContent `json:"contents"`
		GenerationConfig  struct {
			MaxOutputTokens int `json:"maxOutputTokens"`
		} `json:"generationConfig"`
	}{geminiContent{Parts: []geminiPart{{Text: &prompt}}}, contents, struct {
		MaxOutputTokens int `json:"maxOutputTokens"`
	}{1024}})
}

// STOP does not mean there were no calls: Gemini uses it for functionCall
// replies too. Inspect parts and let the common reducer decide turn state.
// modelVersion, when returned, supplies the exact producer identity.
func (e *engine) parseGemini(data []byte) (common.ResponseData, error) {
	var wire struct {
		ModelVersion string `json:"modelVersion"`
		Candidates   []struct {
			Content struct {
				Parts []json.RawMessage `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		Usage *struct {
			Input    *int `json:"promptTokenCount"`
			Output   int  `json:"candidatesTokenCount"`
			Read     int  `json:"cachedContentTokenCount"`
			Thoughts int  `json:"thoughtsTokenCount"`
		} `json:"usageMetadata"`
	}
	if err := json.Unmarshal(data, &wire); err != nil {
		return common.ResponseData{}, errors.New("invalid Gemini response")
	}
	if wire.Usage == nil || wire.Usage.Input == nil {
		return common.ResponseData{}, errors.New("Gemini response missing usage")
	}
	if len(wire.Candidates) == 0 {
		return common.ResponseData{}, errors.New("Gemini response missing candidate")
	}
	from := e.target()
	if wire.ModelVersion != "" {
		from.Model = wire.ModelVersion
	}
	// Gemini mixes conventions inside one usage object. Cached input is a
	// subset of prompt tokens, but thinking output is additional to candidate
	// output. Subtract on input and add on output to get disjoint categories.
	u := wire.Usage
	result := common.ResponseData{From: from, Usage: common.Usage{Input: *u.Input - u.Read, CacheRead: u.Read, Output: u.Output + u.Thoughts}}
	for _, raw := range wire.Candidates[0].Content.Parts {
		var part geminiPart
		if err := json.Unmarshal(raw, &part); err != nil {
			return result, errors.New("invalid Gemini part")
		}
		switch {
		case part.FunctionCall != nil:
			call := part.FunctionCall
			args, err := canonical(call.Args)
			if err != nil {
				return result, err
			}
			result.Parts = append(result.Parts, common.Part{Type: "tool_call", CallID: call.ID, Name: call.Name, Args: args, From: from, Opaque: part.ThoughtSignature})
		case part.Thought:
			result.Parts = append(result.Parts, common.Part{Type: "opaque", From: from, Data: raw})
		case part.Text != nil:
			result.Parts = append(result.Parts, common.Part{Type: "text", Text: *part.Text})
		default:
			return result, errors.New("unknown Gemini response part")
		}
	}
	if len(result.Parts) == 0 {
		return result, errors.New("Gemini response has no content")
	}
	return result, nil
}
