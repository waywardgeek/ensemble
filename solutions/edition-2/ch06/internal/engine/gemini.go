package engine

import (
	"encoding/json"
	"errors"

	"ensemble/internal/common"
)

type geminiCall struct {
	// ID is the provider-issued call identity, retained for its matching result.
	ID string `json:"id,omitempty"`
	// Name is the declared function identifier used to correlate calls and results.
	Name string `json:"name"`
	// Args retains the tool's JSON object before canonical argument capture.
	Args json.RawMessage `json:"args"`
}
type geminiResult struct {
	// ID is the provider-issued call identity, retained for its matching result.
	ID string `json:"id,omitempty"`
	// Name is the declared function identifier used to correlate calls and results.
	Name string `json:"name"`
	// Response is the object-valued Gemini tool testimony, never a bare string.
	Response struct {
		// Result carries bounded visible tool output within the result object.
		Result string `json:"result"`
		// IsError distinguishes ordinary tool failure from a successful result.
		IsError bool `json:"is_error,omitempty"`
	} `json:"response"`
}

// A thoughtSignature is a sibling of functionCall on the wire. Keeping it
// beside the call while parsing and rendering prevents an unrelated opaque
// part from being mistaken for that call's mandatory replay material.
type geminiPart struct {
	// InlineData embeds supported local media bytes with their MIME type.
	InlineData any `json:"inlineData,omitempty"`
	// Text holds visible provider or user content, including an explicitly empty string.
	Text *string `json:"text,omitempty"`
	// FunctionCall carries provider arguments and the declared tool name.
	FunctionCall *geminiCall `json:"functionCall,omitempty"`
	// FunctionResponse carries a tool result with its required function name.
	FunctionResponse *geminiResult `json:"functionResponse,omitempty"`
	// Thought marks private reasoning content that must not become visible prose.
	Thought bool `json:"thought,omitempty"`
	// ThoughtSignature binds private replay material to this exact part.
	ThoughtSignature json.RawMessage `json:"thoughtSignature,omitempty"`
}
type geminiContent struct {
	// Role is the provider spelling chosen from neutral authorship.
	Role string `json:"role,omitempty"`
	// Parts retains ordered Gemini content blocks within a turn.
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
			case "blob":
				data, err := e.blob(p)
				if err != nil {
					return nil, err
				}
				part.InlineData = map[string]string{"mimeType": p.MIME, "data": data}
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
				if p.From == e.Target() {
					part.ThoughtSignature = p.Opaque
				}
			case "tool_result":
				_, _, call, err := callFor(c, p.CallID)
				if err != nil {
					return nil, err
				}
				for _, nested := range p.Parts {
					if nested.Type == "blob" {
						return nil, errors.New("Gemini tool results do not support media on this renderer")
					}
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
				if p.From != e.Target() {
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
	var declarations []map[string]any
	for _, d := range e.parent.Declarations() {
		declarations = append(declarations, map[string]any{"name": d.Name, "description": d.Description, "parameters": d.Schema})
	}
	var tools []map[string]any
	if len(declarations) > 0 {
		tools = []map[string]any{{"functionDeclarations": declarations}}
	}
	prompt := e.systemPrompt()
	return json.Marshal(struct {
		// Tools contains only the declarations visible to this Agent.
		Tools []map[string]any `json:"tools,omitempty"`
		// SystemInstruction holds Gemini instructions outside conversation contents.
		SystemInstruction geminiContent `json:"systemInstruction"`
		// Contents carries Gemini conversation turns using its user/model vocabulary.
		Contents []geminiContent `json:"contents"`
		// GenerationConfig holds request generation settings, separate from dialogue.
		GenerationConfig struct {
			// MaxOutputTokens bounds Gemini response generation.
			MaxOutputTokens int `json:"maxOutputTokens"`
		} `json:"generationConfig"`
	}{tools, geminiContent{Parts: []geminiPart{{Text: &prompt}}}, contents, struct {
		// MaxOutputTokens bounds Gemini response generation.
		MaxOutputTokens int `json:"maxOutputTokens"`
	}{1024}})
}

// STOP does not mean there were no calls: Gemini uses it for functionCall
// replies too. Inspect parts and let the common reducer decide turn state.
// modelVersion, when returned, supplies the exact producer identity.
func (e *engine) parseGemini(data []byte) (common.ResponseData, error) {
	var wire struct {
		// ModelVersion captures the resolved producer identity returned by Gemini.
		ModelVersion string `json:"modelVersion"`
		// Candidates holds provider alternatives; this client consumes the first.
		Candidates []struct {
			// Content preserves the provider's ordered blocks or message text.
			Content struct {
				// Parts retains ordered Gemini content blocks within a turn.
				Parts []json.RawMessage `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		// Usage must be present to distinguish measured counts from missing accounting.
		Usage *struct {
			// Input retains the provider's input count before cache normalization.
			Input *int `json:"promptTokenCount"`
			// Output retains the provider's generated-token count before normalization.
			Output int `json:"candidatesTokenCount"`
			// Read records the provider's cached-input token category.
			Read int `json:"cachedContentTokenCount"`
			// Thoughts counts Gemini thinking tokens additional to candidate output.
			Thoughts int `json:"thoughtsTokenCount"`
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
	from := e.Target()
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
