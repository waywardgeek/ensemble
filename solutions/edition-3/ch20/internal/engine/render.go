package engine

import (
	"encoding/json"
	"ensemble/internal/common"
	"fmt"
	"strings"
)

func raw(v any) json.RawMessage { b, _ := json.Marshal(v); return b }

// Neither an absent producer model nor equality of moving routing aliases
// proves replay compatibility. ResolvedModel is explicit target evidence.
func matches(p *common.Provenance, c common.Config) bool {
	return p != nil && p.Model != "" && c.ResolvedModel != "" && p.Vendor == c.Vendor && p.Surface == c.Surface && p.Model == c.ResolvedModel
}
func textParts(ps common.PartList) (string, error) {
	var b strings.Builder
	for _, p := range ps {
		switch p.Type {
		case "text", "reasoning_summary":
			b.WriteString(p.Text)
		case "redacted":
			b.WriteString(p.Stub)
		default:
			return "", fmt.Errorf("cannot represent %s as text", p.Type)
		}
	}
	return b.String(), nil
}
func render(c common.Context, cfg common.Config, defs []common.ToolDefinition, disclosures ...anthropicDisclosure) ([]byte, error) {
	return renderWithFiles(c, cfg, defs, nil, disclosures...)
}
func renderWithFiles(c common.Context, cfg common.Config, defs []common.ToolDefinition, files common.Sandbox, disclosures ...anthropicDisclosure) ([]byte, error) {
	c, cfg, defs = projectPrompt(c, cfg, defs)
	var disclosure anthropicDisclosure
	if len(disclosures) > 0 {
		disclosure = disclosures[0]
	}
	var err error
	c, err = prepareMedia(c, cfg, files)
	if err != nil {
		return nil, err
	}
	c, err = withCallIDs(c)
	if err != nil {
		return nil, err
	}
	switch cfg.Vendor {
	case common.Anthropic:
		return renderAnthropic(c, cfg, defs, disclosure)
	case common.OpenAI:
		if cfg.Surface == common.Responses {
			return renderResponses(c, cfg, defs)
		}
		return renderOpenAI(c, cfg, defs)
	case common.Gemini:
		return renderGemini(c, cfg, defs)
	}
	return nil, fmt.Errorf("unsupported vendor")
}
func system(c common.Context, cfg common.Config) (string, error) {
	s := systemPrompt
	if cfg.Skills != nil {
		s = ""
	}
	if cfg.SystemPrompt != "" {
		if s != "" {
			s += "\n"
		}
		s += cfg.SystemPrompt
	}
	if cfg.RequireToolExplanation {
		s += " Before each tool batch, emit nonblank ordinary assistant text explaining what you will do and why, before the first tool call. Hidden reasoning is not an explanation. A batch without this explanation will be refused without effects."
	}
	for _, entry := range c.Dialogue {
		if entry.Actor == common.System && !(inlineFeatures(cfg) && entry.Kind == common.KindTools) {
			t, err := textParts(entry.Parts)
			if err != nil {
				return "", err
			}
			s += "\n" + t
		}
	}

	return s, nil
}

type aBlock struct {
	CacheControl *aCacheControl  `json:"cache_control,omitempty"`
	Type         string          `json:"type"`
	Text         *string         `json:"text,omitempty"`
	ID           string          `json:"id,omitempty"`
	Name         string          `json:"name,omitempty"`
	Input        json.RawMessage `json:"input,omitempty"`
	ToolUseID    string          `json:"tool_use_id,omitempty"`
	Content      *string         `json:"content,omitempty"`
	IsError      bool            `json:"is_error,omitempty"`
}
type aMessage struct {
	KeepBlocks bool              `json:"-"`
	Role       string            `json:"role"`
	Content    []json.RawMessage `json:"content"`
}

// Preserve the accepted text-only request shape; structured parts use blocks.
func (m aMessage) MarshalJSON() ([]byte, error) {
	var text strings.Builder
	allText := len(m.Content) > 0 && !m.KeepBlocks
	for _, b := range m.Content {
		var p aBlock
		if json.Unmarshal(b, &p) != nil || p.Type != "text" || p.Text == nil || p.CacheControl != nil {
			allText = false
			break
		}
		text.WriteString(*p.Text)
	}
	if allText {
		return json.Marshal(struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{m.Role, text.String()})
	}
	type plain aMessage
	return json.Marshal(plain(m))
}
func aText(s string) json.RawMessage { return raw(aBlock{Type: "text", Text: &s}) }
func renderAnthropic(c common.Context, cfg common.Config, defs []common.ToolDefinition, disclosures ...anthropicDisclosure) ([]byte, error) {
	var disclosure anthropicDisclosure
	if len(disclosures) > 0 {
		disclosure = disclosures[0]
	}
	sys, err := system(c, cfg)
	if err != nil {
		return nil, err
	}
	// Only an actually emitted reference can keep an enabled schema deferred.
	// Direct API loads and contexts missing the enabling pair use ordinary
	// declarations, without changing first-enable provenance or History.
	disclosed := map[string]bool{}
	messages := []aMessage{}
	// Pre-index real results so late arrivals supersede the missing-record bridge.
	calls := map[string]bool{}
	results := map[string]common.Part{}
	for _, e := range c.Dialogue {
		for _, p := range e.Parts {
			if p.Type == "tool_call" {
				if p.CallID == "" || calls[p.CallID] {
					return nil, fmt.Errorf("ambiguous tool call ID")
				}
				calls[p.CallID] = true
			}
			if p.Type == "tool_result" {
				if _, exists := results[p.CallID]; exists {
					return nil, fmt.Errorf("duplicate tool result")
				}
				results[p.CallID] = p
			}
		}
	}
	for id := range results {
		if !calls[id] {
			return nil, fmt.Errorf("orphan tool result")
		}
	}
	appendMessage := func(role string, bs []json.RawMessage) {
		if len(bs) == 0 {
			return
		}
		if len(messages) > 0 && messages[len(messages)-1].Role == role {
			messages[len(messages)-1].Content = append(messages[len(messages)-1].Content, bs...)
		} else {
			messages = append(messages, aMessage{Role: role, Content: bs, KeepBlocks: cacheEnabled(cfg)})
		}
	}
	for _, e := range c.Dialogue {
		if e.Actor == common.System && !(inlineFeatures(cfg) && e.Kind == common.KindTools) {
			continue
		}
		role := "user"
		if e.Actor == common.System {
			role = "system"
		}
		if e.Actor == common.Assistant {
			role = "assistant"
		}
		blocks := []json.RawMessage{}
		batch := []common.Part{}
		for _, p := range e.Parts {
			switch p.Type {
			case "tool_decls":
				if !inlineFeatures(cfg) {
					return nil, fmt.Errorf("inline declaration on unsupported model")
				}
				var delta common.ToolsData
				if err := json.Unmarshal(p.Data, &delta); err != nil {
					return nil, err
				}
				for _, d := range delta.Added {
					blocks = append(blocks, raw(map[string]any{"type": "tool_addition", "tool": map[string]any{"type": "tool_definition", "definition": map[string]any{"name": d.Name, "description": d.Description, "input_schema": d.Schema}}}))
				}
				for _, name := range delta.Removed {
					blocks = append(blocks, raw(map[string]any{"type": "tool_removal", "tool": map[string]any{"type": "tool_reference", "name": name}}))
				}
			case "text", "reasoning_summary":
				blocks = append(blocks, aText(p.Text))
			case "redacted":
				blocks = append(blocks, aText(p.Stub))
			case "blob", "image", "audio", "video":
				blocks = append(blocks, raw(map[string]any{"type": "image", "source": map[string]any{"type": "base64", "media_type": p.MIME, "data": mediaData(p)}}))
			case "opaque":
				if matches(p.From, cfg) {
					blocks = append(blocks, p.Data)
				}
			case "tool_call":
				if e.Actor != common.Assistant {
					return nil, fmt.Errorf("tool call must be agent-authored")
				}
				args, err := canonical(p.Args)
				if err != nil {
					return nil, err
				}
				blocks = append(blocks, raw(aBlock{Type: "tool_use", ID: p.CallID, Name: p.Name, Input: args}))
				batch = append(batch, p)
			case "tool_result": // Emitted next to its actual call, not at arrival position.
			default:
				return nil, fmt.Errorf("unsupported part %q", p.Type)
			}
		}
		appendMessage(role, blocks)
		if len(batch) > 0 {
			bs := []json.RawMessage{}
			for _, call := range batch {
				result, ok := results[call.CallID]
				text := "No result was recorded; execution outcome unknown."
				isError := true
				if ok {
					var err error
					text, err = textParts(result.Parts)
					if err != nil {
						return nil, err
					}
					isError = result.IsError
				}
				if names := disclosure.references[call.CallID]; len(names) > 0 && ok && !isError && call.Name == "load_skill" {
					// Discovery results contain only references. The handler's status
					// remains in canonical History, not this provider wire projection.
					content := make([]json.RawMessage, 0, len(names))
					for _, name := range names {
						disclosed[name] = true
						content = append(content, raw(map[string]string{"type": "tool_reference", "tool_name": name}))
					}
					bs = append(bs, raw(map[string]any{"type": "tool_result", "tool_use_id": call.CallID, "content": content, "is_error": isError}))
				} else {
					bs = append(bs, raw(aBlock{Type: "tool_result", ToolUseID: call.CallID, Content: &text, IsError: isError}))
				}
			}
			appendMessage("user", bs)
		}
	}
	conversationBreakpoints(cfg, messages)
	if len(c.Ephemera) > 0 {
		handoff, tail, err := projectedTail(c)
		if err != nil {
			return nil, err
		}
		blocks := []json.RawMessage{}
		if cacheEnabled(cfg) && handoff != "" {
			blocks = append(blocks, raw(aBlock{Type: "text", Text: &handoff, CacheControl: ephemeralCache()}))
		} else {
			tail = handoff + tail
		}
		if tail != "" {
			blocks = append(blocks, raw(aBlock{Type: "text", Text: &tail}))
		}
		appendMessage("user", blocks)
	}

	tools := anthropicDefinitions(defs)
	if disclosure.catalog != nil {
		tools = anthropicCatalog(disclosure.catalog, disclosed)
	}
	var systemValue any = sys
	if enabled, _ := common.CacheBreakpoints(cfg); enabled {
		if sys != "" {
			systemValue = []aBlock{{Type: "text", Text: &sys, CacheControl: ephemeralCache()}}
		}
		// Deferred definitions cannot carry cache_control. Preserve catalog
		// order and place the breakpoint on the last eligible declaration.
		for i := len(tools) - 1; i >= 0; i-- {
			if !tools[i].DeferLoading {
				tools[i].CacheControl = ephemeralCache()
				break
			}
		}
	}
	return json.Marshal(struct {
		Model     string        `json:"model"`
		MaxTokens int           `json:"max_tokens"`
		System    any           `json:"system"`
		Messages  []aMessage    `json:"messages"`
		Tools     []aDefinition `json:"tools,omitempty"`
		Stream    bool          `json:"stream,omitempty"`
	}{cfg.Model, 1024, systemValue, messages, tools, cfg.Streaming()})
}

// OpenAI has its own tool authorship; it does not inherit Anthropic's grouping.
type oFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}
type oCall struct {
	ID       string    `json:"id"`
	Type     string    `json:"type"`
	Function oFunction `json:"function"`
}
type oMessage struct {
	Role             string  `json:"role"`
	Content          any     `json:"content,omitempty"`
	ToolCalls        []oCall `json:"tool_calls,omitempty"`
	ToolCallID       string  `json:"tool_call_id,omitempty"`
	ReasoningContent *string `json:"reasoning_content,omitempty"`
	Reasoning        *string `json:"reasoning,omitempty"`
}

func renderOpenAI(c common.Context, cfg common.Config, defs []common.ToolDefinition) ([]byte, error) {
	sys, err := system(c, cfg)
	if err != nil {
		return nil, err
	}
	messages := []oMessage{{Role: "system", Content: &sys}}
	for _, e := range c.Dialogue {
		if e.Actor == common.System {
			continue
		}
		m := oMessage{Role: "user"}
		if e.Actor == common.Assistant {
			m.Role = "assistant"
		}
		var text strings.Builder
		present := false
		hasMedia := false
		content := []any{}
		results := []oMessage{}
		for _, p := range e.Parts {
			switch p.Type {
			case "text", "reasoning_summary":
				present = true
				text.WriteString(p.Text)
				content = append(content, map[string]any{"type": "text", "text": p.Text})
			case "redacted":
				present = true
				text.WriteString(p.Stub)
				content = append(content, map[string]any{"type": "text", "text": p.Stub})
			case "blob", "image", "audio", "video":
				hasMedia = true
				content = append(content, map[string]any{"type": "image_url", "image_url": map[string]any{"url": "data:" + p.MIME + ";base64," + mediaData(p)}})
			case "opaque":
				if matches(p.From, cfg) {
					var fields map[string]json.RawMessage
					if json.Unmarshal(p.Data, &fields) != nil || len(fields) == 0 {
						return nil, fmt.Errorf("unsupported OpenAI opaque replay material")
					}
					for name, raw := range fields {
						var text string
						if json.Unmarshal(raw, &text) != nil {
							return nil, fmt.Errorf("unsupported OpenAI opaque replay material")
						}
						switch name {
						case "reasoning_content":
							m.ReasoningContent = &text
						case "reasoning":
							m.Reasoning = &text
						default:
							return nil, fmt.Errorf("unsupported OpenAI opaque replay material")
						}
					}
				}
			case "tool_call":
				args, err := canonical(p.Args)
				if err != nil {
					return nil, err
				}
				m.ToolCalls = append(m.ToolCalls, oCall{p.CallID, "function", oFunction{p.Name, string(args)}})
			case "tool_result":
				t, err := textParts(p.Parts)
				if err != nil {
					return nil, err
				}
				results = append(results, oMessage{Role: "tool", Content: &t, ToolCallID: p.CallID})
			default:
				return nil, fmt.Errorf("unsupported part %q", p.Type)
			}
		}
		if present {
			t := text.String()
			m.Content = &t
		}
		if hasMedia {
			m.Content = content
		}
		if present || hasMedia || len(m.ToolCalls) > 0 || m.ReasoningContent != nil || m.Reasoning != nil {
			messages = append(messages, m)
		}
		messages = append(messages, results...)
	}

	// The final history boundary follows wire order, not causal History:
	// complete tool-result groups can move ahead of interleaved user hints.
	// Project before choosing markers and before adding transient ephemera.
	messages = adjacentOpenAIToolResults(messages)
	markers := 0
	if cacheEnabled(cfg) {
		for i := range messages {
			if text, ok := messages[i].Content.(*string); ok && text != nil {
				messages[i].Content = []any{map[string]any{"type": "text", "text": *text}}
			}
		}
		if markOpenAIMessage(&messages[0]) {
			markers++
		}
		for i := len(messages) - 1; i > 0; i-- {
			if markOpenAIMessage(&messages[i]) {
				markers++
				break
			}
		}
	}
	if len(c.Ephemera) > 0 {
		handoff, tail, err := projectedTail(c)
		if err != nil {
			return nil, err
		}
		m := oMessage{Role: "system"}
		if cacheEnabled(cfg) && handoff != "" {
			m.Content = &handoff
			if markOpenAIMessage(&m) {
				markers++
			}
			if tail != "" {
				m.Content = append(m.Content.([]any), map[string]any{"type": "text", "text": tail})
			}
		} else {
			tail = handoff + tail
			m.Content = &tail
		}
		messages = append(messages, m)
	}
	var cacheOptions any
	if markers > 0 {
		cacheOptions = map[string]string{"mode": "explicit"}
	}

	// Use OpenAI's documented replacement for max_tokens (rejected by GPT-5.4).
	// Keep the same 1024 native cap, including reasoning and visible output.
	return json.Marshal(struct {
		Model               string        `json:"model"`
		Messages            []oMessage    `json:"messages"`
		MaxCompletionTokens int           `json:"max_completion_tokens"`
		Tools               []oDefinition `json:"tools,omitempty"`
		Stream              bool          `json:"stream,omitempty"`
		StreamOptions       any           `json:"stream_options,omitempty"`
		CacheOptions        any           `json:"prompt_cache_options,omitempty"`
	}{cfg.Model, messages, 1024, openAIDefinitions(defs), cfg.Streaming(), streamOptions(cfg), cacheOptions})
}

// OpenAI requires a call group's results immediately after its assistant
// message. Move only complete real groups ahead of intervening user messages
// in this disposable projection. History stays causal; neither result facts nor
// within-group order change. Do not cross unrelated dialogue or invent a result.
func adjacentOpenAIToolResults(messages []oMessage) []oMessage {
	out := make([]oMessage, 0, len(messages))
	for i := 0; i < len(messages); i++ {
		m := messages[i]
		out = append(out, m)
		if len(m.ToolCalls) == 0 {
			continue
		}
		pending := make(map[string]bool, len(m.ToolCalls))
		for _, call := range m.ToolCalls {
			pending[call.ID] = true
		}
		if len(pending) != len(m.ToolCalls) {
			continue // Ambiguous IDs cannot identify a complete result group.
		}
		var results, hints []oMessage
		end := i + 1
	scan:
		for end < len(messages) && len(pending) > 0 {
			next := messages[end]
			switch next.Role {
			case "tool":
				if !pending[next.ToolCallID] {
					break scan
				}
				delete(pending, next.ToolCallID)
				results = append(results, next)
			case "user":
				hints = append(hints, next)
			default:
				break scan
			}
			end++
		}
		if len(pending) == 0 {
			out = append(out, results...)
			out = append(out, hints...)
			i = end - 1
		}
	}
	return out
}

type gCall struct {
	ID   string          `json:"id,omitempty"`
	Name string          `json:"name"`
	Args json.RawMessage `json:"args"`
}
type gResult struct {
	ID       string          `json:"id,omitempty"`
	Name     string          `json:"name"`
	Response json.RawMessage `json:"response"`
}
type gPart struct {
	Text             *string         `json:"text,omitempty"`
	FunctionCall     *gCall          `json:"functionCall,omitempty"`
	FunctionResponse *gResult        `json:"functionResponse,omitempty"`
	ThoughtSignature json.RawMessage `json:"thoughtSignature,omitempty"`
}
type gMessage struct {
	Role  string            `json:"role,omitempty"`
	Parts []json.RawMessage `json:"parts"`
}

func renderGemini(c common.Context, cfg common.Config, defs []common.ToolDefinition) ([]byte, error) {
	sys, err := system(c, cfg)
	if err != nil {
		return nil, err
	}
	messages := []gMessage{}
	names := map[string]string{}
	for _, e := range c.Dialogue {
		for _, p := range e.Parts {
			if p.Type == "tool_call" {
				names[p.CallID] = p.Name
			}
		}
	}
	for _, e := range c.Dialogue {
		if e.Actor == common.System {
			continue
		}
		m := gMessage{Role: "user", Parts: []json.RawMessage{}}
		if e.Actor == common.Assistant {
			m.Role = "model"
		}
		for _, p := range e.Parts {
			switch p.Type {
			case "text", "reasoning_summary":
				t := p.Text
				part := gPart{Text: &t}
				if matches(p.From, cfg) {
					part.ThoughtSignature = p.Opaque
				}
				m.Parts = append(m.Parts, raw(part))
			case "redacted":
				t := p.Stub
				m.Parts = append(m.Parts, raw(gPart{Text: &t}))
			case "blob", "image", "audio", "video":
				m.Parts = append(m.Parts, raw(map[string]any{"inlineData": map[string]any{"mimeType": p.MIME, "data": mediaData(p)}}))
			case "opaque":
				if matches(p.From, cfg) {
					m.Parts = append(m.Parts, p.Data)
				}
			case "tool_call":
				args, err := canonical(p.Args)
				if err != nil {
					return nil, err
				}
				part := gPart{FunctionCall: &gCall{p.CallID, p.Name, args}}
				if matches(p.From, cfg) {
					part.ThoughtSignature = p.Opaque
				}
				m.Parts = append(m.Parts, raw(part))
			case "tool_result":
				name := names[p.CallID]
				if name == "" {
					return nil, fmt.Errorf("tool result has no recorded function name")
				}
				t, err := textParts(p.Parts)
				if err != nil {
					return nil, err
				}
				response := json.RawMessage(t)
				var obj map[string]json.RawMessage
				object := json.Unmarshal(response, &obj) == nil && obj != nil
				if object && (p.IsError || obj["is_error"] != nil) {
					// Keep successful ordinary objects unchanged. On errors or
					// collisions, the outer flag is authoritative and the exact
					// original payload remains under result.
					response = raw(struct {
						Result  json.RawMessage `json:"result"`
						IsError bool            `json:"is_error"`
					}{response, p.IsError})
				} else if !object {
					response = raw(struct {
						Result  string `json:"result"`
						IsError bool   `json:"is_error,omitempty"`
					}{t, p.IsError})
				}
				m.Parts = append(m.Parts, raw(gPart{FunctionResponse: &gResult{p.CallID, name, response}}))
			default:
				return nil, fmt.Errorf("unsupported part %q", p.Type)
			}
		}
		if len(m.Parts) > 0 {
			messages = append(messages, m)
		}
	}
	if len(c.Ephemera) > 0 {
		tail, err := textParts(c.Ephemera)
		if err != nil {
			return nil, err
		}
		messages = append(messages, gMessage{Role: "user", Parts: []json.RawMessage{raw(gPart{Text: &tail})}})
	}

	tools, err := geminiDefinitions(defs)
	if err != nil {
		return nil, err
	}
	return json.Marshal(struct {
		System     gMessage       `json:"systemInstruction"`
		Contents   []gMessage     `json:"contents"`
		Tools      []gDefinitions `json:"tools,omitempty"`
		Generation struct {
			Max int `json:"maxOutputTokens"`
		} `json:"generationConfig"`
	}{gMessage{Parts: []json.RawMessage{raw(gPart{Text: &sys})}}, messages, tools, struct {
		Max int `json:"maxOutputTokens"`
	}{1024}})
}

// withCallIDs supplies stable render-only IDs when the source surface omitted
// them. Existing IDs are immutable. Result matching never guesses by tool name.
func withCallIDs(c common.Context) (common.Context, error) {
	reserved := map[string]bool{}
	for _, e := range c.Dialogue {
		for _, p := range e.Parts {
			if (p.Type == "tool_call" || p.Type == "tool_result") && p.CallID != "" {
				reserved[p.CallID] = true
			}
		}
	}
	c.Dialogue = append([]common.Entry(nil), c.Dialogue...)
	missingCalls := []string{}
	missingResults := 0
	for i := range c.Dialogue {
		e := &c.Dialogue[i]
		e.Parts = append(common.PartList(nil), e.Parts...)
		for j := range e.Parts {
			p := &e.Parts[j]
			if p.Type == "tool_call" && p.CallID == "" {
				base := fmt.Sprintf("call_%d_%d_%d", e.Seq, i, j)
				id := base
				for suffix := 1; reserved[id]; suffix++ {
					id = fmt.Sprintf("%s_%d", base, suffix)
				}
				reserved[id] = true
				p.CallID = id
				missingCalls = append(missingCalls, id)
			}
			if p.Type == "tool_result" && p.CallID == "" {
				missingResults++
			}
		}
	}
	if missingResults > 0 {
		if len(missingCalls) != 1 || missingResults != 1 {
			return c, fmt.Errorf("ambiguous missing tool result ID")
		}
		for i := range c.Dialogue {
			for j := range c.Dialogue[i].Parts {
				p := &c.Dialogue[i].Parts[j]
				if p.Type == "tool_result" && p.CallID == "" {
					p.CallID = missingCalls[0]
				}
			}
		}
	}
	return c, nil
}

func streamOptions(c common.Config) any {
	if c.Streaming() {
		return map[string]bool{"include_usage": true}
	}
	return nil
}
