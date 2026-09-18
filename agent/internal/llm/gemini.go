package llm

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
	"io"
	"net/http"
	"strings"

	"github.com/waywardgeek/ensemble/agent/internal/common"
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
	// Omitted when nothing is declared — see anthRequest.Tools.
	Tools []gemTool `json:"tools,omitempty"`
}

// gemTool is Gemini's declaration shape: `tools` is a list of tool GROUPS,
// each holding many functionDeclarations. We send one group with everything
// in it. Gemini's schema dialect is an OpenAPI subset, which is why the
// registry's schemas avoid keys like additionalProperties that it rejects.
type gemTool struct {
	FunctionDeclarations []gemFunctionDecl `json:"functionDeclarations"`
}

type gemFunctionDecl struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Parameters  json.RawMessage `json:"parameters,omitempty"`
}

func gemTools(decls []common.ToolDecl) []gemTool {
	if len(decls) == 0 {
		return nil
	}
	var fns []gemFunctionDecl
	for _, d := range decls {
		fns = append(fns, gemFunctionDecl{Name: d.Name, Description: d.Description, Parameters: d.Schema})
	}
	return []gemTool{{FunctionDeclarations: fns}}
}

type gemGenConfig struct {
	MaxOutputTokens int              `json:"maxOutputTokens,omitempty"`
	ThinkingConfig  *gemThinkConfig  `json:"thinkingConfig,omitempty"`
}

// gemThinkConfig is Gemini's thinking configuration.
//
// includeThoughts is REQUIRED or you get billed for thinking you never
// receive. thinkingBudget is the token ceiling, same concept as Anthropic's
// budget_tokens.
type gemThinkConfig struct {
	ThinkingBudget  int  `json:"thinkingBudget"`
	IncludeThoughts bool `json:"includeThoughts"`
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
	// one of them is expressible as a local path, which is why common.BlobPart
	// carries a common.Ref.
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

// geminiFileParts renders every locator in an entry to Gemini's remote form.
//
// Only common.RefURI can be sent. A RefPath names a file on the machine running the
// agent, and a common.RefHandle names something the framework holds — possibly only in
// memory. Gemini can fetch neither, so neither is guessed at: resolving one
// into a uri is an upload, which is a job for the layer that owns the bytes.
//
// The alternative — dropping what cannot be sent — is how a multimodal request
// silently loses its attachment and comes back with a confident answer about a
// file the model never saw.
func geminiFileParts(r renderable) ([]gemPart, error) {
	var out []gemPart
	for _, b := range r.Blobs {
		if b.Ref.Kind != common.RefURI {
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
		if ref.Kind == common.RefURI {
			out = append(out, gemPart{FileData: &gemFileData{FileURI: ref.Locator}})
		}
	}
	return out, nil
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

func (geminiSeam) Render(c *common.Context, cfg common.Config) (*http.Request, error) {
	target := common.Provenance{Vendor: common.VendorGemini, Model: cfg.Model, Surface: common.SurfaceGenerateContent}

	// Gemini's functionResponse requires the function NAME, and a
	// common.ToolResultPart carries only the call id. The information is in the
	// context, just not adjacent to where this vendor wants it — so the
	// renderer resolves it rather than the context duplicating it.
	//
	// Note what did NOT happen: the context did not grow a field. A lookup in
	// the renderer is the right home for a fact only one vendor needs.
	nameByCall := map[string]string{}
	for _, entry := range c.Dialogue {
		for _, p := range entry.Parts {
			if call, ok := p.(common.ToolCallPart); ok {
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
		case common.ActorTool:
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

		case common.ActorAgent:
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

		default: // common.ActorHuman
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

	body := gemRequest{Contents: contents, Tools: gemTools(cfg.Tools)}
	if cfg.SystemPrompt != "" {
		body.SystemInstruction = &gemContent{Parts: []gemPart{{Text: cfg.SystemPrompt}}}
	}

	// Resolve thinking: same ThinkingFor resolver as all vendors.
	_, thinkingBudget := common.ThinkingFor(cfg)
	maxTokens := common.EnsureMaxTokens(cfg.MaxTokens, thinkingBudget)

	if maxTokens > 0 || thinkingBudget > 0 {
		body.GenerationConfig = &gemGenConfig{MaxOutputTokens: maxTokens}
		if thinkingBudget > 0 {
			body.GenerationConfig.ThinkingConfig = &gemThinkConfig{
				ThinkingBudget:  thinkingBudget,
				IncludeThoughts: true,
			}
		}
	}

	// Gemini signals streaming in the URL, not the body: a different method
	// and an `alt=sse` query parameter. Without alt=sse the streaming method
	// returns a JSON ARRAY of response objects instead of SSE frames, which
	// parses as neither. Two vendors put this in the request body; assuming
	// the third does too produces a request that succeeds and streams nothing.
	method := "generateContent"
	query := ""
	if common.StreamingFor(cfg) != 0 {
		method, query = "streamGenerateContent", "?alt=sse"
	}
	url := fmt.Sprintf("%s/v1beta/models/%s:%s%s", cfg.BaseURL, cfg.Model, method, query)
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

func (geminiSeam) Parse(resp *http.Response, cb common.StreamCallbacks) error {
	if resp.StatusCode != http.StatusOK {
		return parseErrorResponse(resp, cb)
	}
	if isSSE(resp) {
		return gemParseStream(resp, cb)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("gemini: read body: %w", err)
	}
	cb.Frame("", body)
	parts, from, usage, err := gemAssemble(body)
	if err != nil {
		return err
	}
	emitLengthOneDeltas(parts, cb, gemThinkingText)
	cb.Emit(common.Event{Type: common.ResponseStarted})
	cb.Emit(common.Event{Type: common.ResponseEnded, Response: &common.ResponseData{
		Parts: parts, From: from, Usage: usage,
	}})
	emitFinals(parts, cb)
	return nil
}

func gemAssemble(body []byte) (common.PartList, common.Provenance, common.Usage, error) {
	var resp gemResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, common.Provenance{}, common.Usage{}, fmt.Errorf("gemini: %w", err)
	}
	if len(resp.Candidates) == 0 {
		return nil, common.Provenance{}, common.Usage{}, fmt.Errorf("gemini: response had no candidates")
	}
	from := common.Provenance{Vendor: common.VendorGemini, Model: resp.ModelVersion, Surface: common.SurfaceGenerateContent}

	var parts common.PartList
	for _, p := range resp.Candidates[0].Content.Parts {
		switch {
		case p.FunctionCall != nil:
			// NOTE: finishReason is "STOP" here, not a tool-call value. There
			// is no tool-call member in the enum at all. Detect tool calls by
			// inspecting the parts — a parser keyed on the stop signal, as the
			// other two vendors allow, silently never calls a tool.
			parts = append(parts, common.ToolCallPart{
				CallID: p.FunctionCall.ID,
				From:   from,
				Name:   p.FunctionCall.Name,
				Args:   jsonObject(p.FunctionCall.Args),
				Opaque: p.ThoughtSignature,
			})
		case p.Thought:
			raw, _ := json.Marshal(p)
			parts = append(parts, common.OpaquePart{From: from, Data: raw})
		case p.Text != "":
			parts = append(parts, common.TextPart{Text: p.Text})
		}
	}
	return parts, from, gemCanonicalUsage(resp.UsageMetadata), nil
}

// gemCanonicalUsage converts Gemini's self-disagreeing accounting to the
// disjoint form: subset on the input side, disjoint on the output side.
func gemCanonicalUsage(u gemUsage) common.Usage {
	uncached := u.PromptTokenCount - u.CachedContentTokenCount
	if uncached < 0 {
		uncached = 0
	}
	return common.Usage{
		Input:      uncached,
		CacheWrite: 0, // Gemini reports no cache-write token count anywhere
		CacheRead:  u.CachedContentTokenCount,
		Output:     u.CandidatesTokenCount + u.ThoughtsTokenCount,
	}
}

// gemThinkingText pulls readable text out of a Gemini thought part.
func gemThinkingText(p common.OpaquePart) string {
	var b gemRespPart
	if json.Unmarshal(p.Data, &b) != nil {
		return ""
	}
	return b.Text
}

// --- streaming ---------------------------------------------------------

// Gemini's stream has NO BLOCK INDEX.
//
// Anthropic numbers its content blocks and OpenAI numbers its tool calls.
// Gemini sends a sequence of whole response objects, each carrying a `parts`
// array, and says nothing about whether the text in this frame continues the
// text in the last one or begins something new. Continuation has to be
// INFERRED, and the rule is the only one available: a part continues the open
// part when their kinds match, and starts a new part when they do not.
//
// This is also where the model table earns its keep. Gemini streams text and
// thoughts in fragments, but a functionCall arrives complete in a single
// frame — so its arguments never stream, no matter what the request asked
// for. The parser does not need to be told that: it reports what arrives, and
// a call that arrives whole simply produces one delta. The Stream bitmask
// records the fact for everyone upstream who needs to PLAN for it.
type gemOpenPart struct {
	PartID    uint64
	Kind      common.DeltaKind
	Text      strings.Builder
	Signature json.RawMessage
	Call      *gemFunctionCall
}

func gemParseStream(resp *http.Response, cb common.StreamCallbacks) error {
	from := common.Provenance{Vendor: common.VendorGemini, Surface: common.SurfaceGenerateContent}

	var (
		open    []*gemOpenPart
		usage   common.Usage
		nextID  uint64
		started bool
		perr    error
	)
	alloc := func() uint64 { nextID++; return nextID }

	// openFor returns the part that a fragment of this kind continues, or a
	// fresh one when the previous part was something else.
	openFor := func(kind common.DeltaKind) *gemOpenPart {
		if n := len(open); n > 0 && open[n-1].Kind == kind && open[n-1].Call == nil {
			return open[n-1]
		}
		p := &gemOpenPart{PartID: alloc(), Kind: kind}
		open = append(open, p)
		return p
	}

	readErr := ReadSSE(resp.Body, func(eventType string, data []byte) {
		cb.Frame(eventType, data)
		if perr != nil {
			return
		}

		var chunk gemResponse
		if err := json.Unmarshal(data, &chunk); err != nil {
			perr = fmt.Errorf("gemini: stream chunk: %w", err)
			return
		}
		if chunk.ModelVersion != "" {
			from.Model = chunk.ModelVersion
		}
		if !started {
			started = true
			cb.Emit(common.Event{Type: common.ResponseStarted})
		}
		// usageMetadata is CUMULATIVE on every frame, so the last one wins
		// rather than the counts being summed.
		if chunk.UsageMetadata != (gemUsage{}) {
			usage = gemCanonicalUsage(chunk.UsageMetadata)
		}
		if len(chunk.Candidates) == 0 {
			return
		}

		for _, p := range chunk.Candidates[0].Content.Parts {
			switch {
			case p.FunctionCall != nil:
				// Arrives complete. One part, one pair of deltas, and the
				// arguments are already whole — there is nothing to stream.
				np := &gemOpenPart{
					PartID:    alloc(),
					Kind:      common.DeltaToolCall,
					Call:      p.FunctionCall,
					Signature: p.ThoughtSignature,
				}
				open = append(open, np)
				cb.Delta(np.PartID, common.DeltaToolCall, p.FunctionCall.Name)
				if args := string(jsonObject(p.FunctionCall.Args)); args != "{}" {
					cb.Delta(np.PartID, common.DeltaToolCall, args)
				}

			case p.Thought:
				op := openFor(common.DeltaThinking)
				op.Text.WriteString(p.Text)
				if len(p.ThoughtSignature) > 0 {
					op.Signature = p.ThoughtSignature
				}
				if p.Text != "" {
					cb.Delta(op.PartID, common.DeltaThinking, p.Text)
				}

			case p.Text != "":
				op := openFor(common.DeltaText)
				op.Text.WriteString(p.Text)
				cb.Delta(op.PartID, common.DeltaText, p.Text)
			}
		}
	})

	if perr != nil {
		return perr
	}
	if readErr != nil {
		return fmt.Errorf("gemini: stream: %w", readErr)
	}

	var parts common.PartList
	var ids []uint64
	for _, op := range open {
		switch op.Kind {
		case common.DeltaToolCall:
			parts = append(parts, common.ToolCallPart{
				CallID: op.Call.ID,
				From:   from,
				Name:   op.Call.Name,
				Args:   jsonObject(op.Call.Args),
				Opaque: op.Signature,
			})
		case common.DeltaThinking:
			// Rebuilt through gemRespPart so the bytes match what the
			// whole-document path would have stored. A thought replayed
			// without its signature is a 400 on the next request, so this
			// is not cosmetic.
			raw, err := json.Marshal(gemRespPart{
				Text:             op.Text.String(),
				Thought:          true,
				ThoughtSignature: op.Signature,
			})
			if err != nil {
				return fmt.Errorf("gemini: rebuild thought: %w", err)
			}
			parts = append(parts, common.OpaquePart{From: from, Data: raw})
		default:
			if t := op.Text.String(); t != "" {
				parts = append(parts, common.TextPart{Text: t})
			} else {
				continue
			}
		}
		ids = append(ids, op.PartID)
	}

	cb.Emit(common.Event{Type: common.ResponseEnded, Response: &common.ResponseData{
		Parts: parts, From: from, Usage: usage,
	}})
	for n, p := range parts {
		cb.Final(ids[n], p)
	}
	return nil
}
