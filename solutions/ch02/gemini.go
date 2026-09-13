package main

// Gemini — Gemini, generateContent.
//
// Wire format verified against ai.google.dev on 2026-09-12.
//
// The genuinely alien one, and it goes last on purpose. The tempting order
// puts the most different vendor second and the most familiar one third — and
// then "the third was nearly free" is true because the third was easy, not
// because the seam was right.

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type geminiSeam struct{}

// --- request -----------------------------------------------------------

type gemRequest struct {
	// Hoisted clean out of the message list. Must be a Content OBJECT: a bare
	// string is rejected. Its `role` is accepted and ignored — while
	// role:"system" INSIDE contents is a 400.
	SystemInstruction *gemContent   `json:"systemInstruction,omitempty"`
	Contents          []gemContent  `json:"contents"`
	GenerationConfig  *gemGenConfig `json:"generationConfig,omitempty"`
}

type gemGenConfig struct {
	MaxOutputTokens int `json:"maxOutputTokens,omitempty"`
}

// gemContent: `contents`, not `messages`; `parts`, not blocks; and the
// assistant is called "model".
type gemContent struct {
	Role  string    `json:"role,omitempty"`
	Parts []gemPart `json:"parts"`
}

type gemPart struct {
	Text string `json:"text,omitempty"`
	// FileData is the REMOTE-reference form: it names bytes Gemini will fetch
	// for itself. Three of Gemini's four file input methods arrive here — a
	// File API uri, a registered gs:// object, and an external URL — and not
	// one of them is expressible as a local path, which is why BlobPart
	// carries a Ref.
	FileData         *gemFileData         `json:"fileData,omitempty"`
	FunctionCall     *gemFunctionCall     `json:"functionCall,omitempty"`
	FunctionResponse *gemFunctionResponse `json:"functionResponse,omitempty"`
	// A SIBLING key of functionCall, not a member of it. Gemini 3.x returns
	// 400 if a replayed functionCall arrives without it.
	ThoughtSignature json.RawMessage `json:"thoughtSignature,omitempty"`
}

// gemFileData is Gemini's remote file reference.
//
//	{ "mimeType": string, "fileUri": string }
//
// fileUri is required; mimeType is optional. Verified against the Gemini API
// reference: https://ai.google.dev/api/generate-content (FileData).
//
// Note what is NOT here: the base64 inline form (`inlineData`/`inline_data`,
// carrying `data`). Inlining is a decision made HERE, while building one
// request, and is never written back into the log — a log you cannot grep is a
// log you cannot debug.
type gemFileData struct {
	MIMEType string `json:"mimeType,omitempty"`
	FileURI  string `json:"fileUri"`
}

type gemFunctionCall struct {
	ID   string          `json:"id,omitempty"` // optional, model-dependent
	Name string          `json:"name"`
	Args json.RawMessage `json:"args,omitempty"` // an OBJECT, unlike OpenAI's string
}

type gemFunctionResponse struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"` // REQUIRED — and the context has no field for it
	// Must be a JSON OBJECT. A bare string, number or array is a 400, so a
	// tool result that every other vendor accepts as a scalar has to be
	// wrapped here. One more compromise that belongs in exactly one function.
	Response json.RawMessage `json:"response"`
}

type gemOutput struct {
	Output string `json:"output"`
}

// geminiFileParts renders every locator in an entry to Gemini's remote form.
//
// Only RefURI can be sent. A RefPath names a file on the machine running the
// agent, and a RefHandle names something the framework holds — possibly only in
// memory. Gemini can fetch neither, so neither is guessed at: resolving one
// into a uri is an upload, which is a job for the layer that owns the bytes.
//
// The alternative — dropping what cannot be sent — is how a multimodal request
// silently loses its attachment and comes back with a confident answer about a
// file the model never saw.
func geminiFileParts(r renderable) ([]gemPart, error) {
	var out []gemPart
	for _, b := range r.Blobs {
		if b.Ref.Kind != RefURI {
			return nil, fmt.Errorf("gemini: cannot send a blob of kind %s (%s): Gemini fetches "+
				"remote references only, so this must be resolved to a uri before rendering",
				b.Ref.Kind, b.Ref.Locator)
		}
		out = append(out, gemPart{FileData: &gemFileData{MIMEType: b.MIME, FileURI: b.Ref.Locator}})
	}
	// A locator that survived a redaction has no MIME type — only a place the
	// superseded content still is. mimeType is optional, so it is simply
	// omitted. One that cannot be fetched is dropped rather than raised: the
	// stub has already said the content is gone, and failing the whole request
	// over a recoverability hint would be worse than not offering it.
	for _, ref := range r.Refs {
		if ref.Kind == RefURI {
			out = append(out, gemPart{FileData: &gemFileData{FileURI: ref.Locator}})
		}
	}
	return out, nil
}

func (geminiSeam) Render(c *Context, cfg Config) (*http.Request, error) {
	target := Provenance{Vendor: VendorGemini, Model: cfg.Model, Surface: SurfaceGenerateContent}

	// Gemini's functionResponse requires the function NAME, and a
	// ToolResultPart carries only the call id. The information is in the
	// context, just not adjacent to where this vendor wants it — so the
	// renderer resolves it rather than the context duplicating it.
	//
	// Note what did NOT happen: the context did not grow a field. A lookup in
	// the renderer is the right home for a fact only one vendor needs.
	nameByCall := map[string]string{}
	for _, entry := range c.Dialogue {
		for _, p := range entry.Parts {
			if call, ok := p.(ToolCallPart); ok {
				nameByCall[call.CallID] = call.Name
			}
		}
	}

	var contents []gemContent
	for _, entry := range c.Dialogue {
		r, err := classify(entry, cfg)
		if err != nil {
			return nil, err
		}
		var parts []gemPart
		role := "user"

		switch entry.Actor {
		case ActorTool:
			for _, res := range r.Tools {
				payload, err := json.Marshal(gemOutput{Output: resultText(res)})
				if err != nil {
					return nil, err
				}
				parts = append(parts, gemPart{FunctionResponse: &gemFunctionResponse{
					ID:       res.CallID,
					Name:     nameByCall[res.CallID],
					Response: payload,
				}})
			}

		case ActorAgent:
			role = "model"
			for _, op := range r.Raw {
				if op.From.SameModel(target) {
					parts = append(parts, gemPart{ThoughtSignature: op.Data})
				}
			}
			for _, t := range r.Texts {
				parts = append(parts, gemPart{Text: t})
			}
			for i, call := range r.Calls {
				p := gemPart{FunctionCall: &gemFunctionCall{
					ID:   callIDFor(call, "fc", entry.Seq, i),
					Name: call.Name,
					Args: jsonObject(call.Args),
				}}
				// Replay the signature only to the exact model that issued it.
				if len(call.Opaque) > 0 && call.From.SameModel(target) {
					p.ThoughtSignature = call.Opaque
				}
				parts = append(parts, p)
			}

		default: // ActorHuman
			for _, t := range r.Texts {
				parts = append(parts, gemPart{Text: t})
			}
		}

		fileParts, err := geminiFileParts(r)
		if err != nil {
			return nil, err
		}
		parts = append(parts, fileParts...)

		if len(parts) > 0 {
			contents = append(contents, gemContent{Role: role, Parts: parts})
		}
	}

	if len(c.Ephemera) > 0 {
		if text := ephemeraText(c.Ephemera); text != "" {
			contents = append(contents, gemContent{Role: "user", Parts: []gemPart{{Text: text}}})
		}
	}

	body := gemRequest{Contents: contents}
	if cfg.SystemPrompt != "" {
		body.SystemInstruction = &gemContent{Parts: []gemPart{{Text: cfg.SystemPrompt}}}
	}
	if cfg.MaxTokens > 0 {
		body.GenerationConfig = &gemGenConfig{MaxOutputTokens: cfg.MaxTokens}
	}

	url := fmt.Sprintf("%s/v1beta/models/%s:generateContent", cfg.BaseURL, cfg.Model)
	return newJSONRequest("POST", url, body, map[string]string{
		"x-goog-api-key": cfg.APIKey,
	})
}

// --- response ----------------------------------------------------------

type gemResponse struct {
	ModelVersion string `json:"modelVersion"`
	Candidates   []struct {
		Content struct {
			Parts []gemRespPart `json:"parts"`
		} `json:"content"`
		FinishReason string `json:"finishReason"`
	} `json:"candidates"`
	UsageMetadata gemUsage `json:"usageMetadata"`
}

type gemRespPart struct {
	Text             string           `json:"text"`
	Thought          bool             `json:"thought"`
	FunctionCall     *gemFunctionCall `json:"functionCall"`
	ThoughtSignature json.RawMessage  `json:"thoughtSignature"`
}

// gemUsage disagrees with ITSELF, which is worse than the chapter claims.
//
//	input side  — SUBSET:   cachedContentTokenCount is inside promptTokenCount
//	output side — DISJOINT: thoughtsTokenCount is NOT inside candidatesTokenCount
//
// Verified by arithmetic on live samples: 76 + 792 + 1001 = 1869 = total.
// Treating thoughts as part of candidates undercounted billed output by 56% in
// one sample, and thinking tokens bill at the output rate.
type gemUsage struct {
	PromptTokenCount        int `json:"promptTokenCount"`
	CandidatesTokenCount    int `json:"candidatesTokenCount"`
	CachedContentTokenCount int `json:"cachedContentTokenCount"`
	ThoughtsTokenCount      int `json:"thoughtsTokenCount"`
}

func (geminiSeam) Parse(status int, body []byte) ([]Event, error) {
	if status != http.StatusOK {
		return []Event{{Type: ErrorOccurred, Error: &ErrorData{
			Status: status, Message: vendorErrorMessage(body),
		}}}, nil
	}
	var resp gemResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("gemini: %w", err)
	}
	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("gemini: response had no candidates")
	}
	from := Provenance{Vendor: VendorGemini, Model: resp.ModelVersion, Surface: SurfaceGenerateContent}

	var parts PartList
	for _, p := range resp.Candidates[0].Content.Parts {
		switch {
		case p.FunctionCall != nil:
			// NOTE: finishReason is "STOP" here, not a tool-call value. There
			// is no tool-call member in the enum at all. Detect tool calls by
			// inspecting the parts — a parser keyed on the stop signal, as the
			// other two vendors allow, silently never calls a tool.
			parts = append(parts, ToolCallPart{
				CallID: p.FunctionCall.ID,
				From:   from,
				Name:   p.FunctionCall.Name,
				Args:   jsonObject(p.FunctionCall.Args),
				Opaque: p.ThoughtSignature,
			})
		case p.Thought:
			raw, _ := json.Marshal(p)
			parts = append(parts, OpaquePart{From: from, Data: raw})
		case p.Text != "":
			parts = append(parts, TextPart{Text: p.Text})
		}
	}

	u := resp.UsageMetadata
	uncached := u.PromptTokenCount - u.CachedContentTokenCount
	if uncached < 0 {
		uncached = 0
	}

	return []Event{
		{Type: ResponseStarted},
		{Type: ResponseEnded, Response: &ResponseData{
			Parts: parts,
			From:  from,
			Usage: Usage{
				Input:      uncached,
				CacheWrite: 0, // Gemini reports no cache-write token count anywhere
				CacheRead:  u.CachedContentTokenCount,
				Output:     u.CandidatesTokenCount + u.ThoughtsTokenCount,
			},
		}},
	}, nil
}
