package engine

import (
	"encoding/json"
	"ensemble/internal/common"
	"strings"
)

// Only Engine knows the three provider declaration envelopes.
type aDefinition struct {
	CacheControl *aCacheControl  `json:"cache_control,omitempty"`
	DeferLoading bool            `json:"defer_loading,omitempty"`
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	InputSchema  json.RawMessage `json:"input_schema"`
}
type functionDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}
type oDefinition struct {
	Type     string             `json:"type"`
	Function functionDefinition `json:"function"`
}
type gDefinitions struct {
	Functions []functionDefinition `json:"functionDeclarations"`
}

func anthropicDefinitions(defs []common.ToolDefinition) []aDefinition {
	var out []aDefinition
	for _, d := range defs {
		out = append(out, aDefinition{Name: d.Name, Description: d.Description, InputSchema: d.Parameters})
	}
	return out
}
func openAIDefinitions(defs []common.ToolDefinition) []oDefinition {
	var out []oDefinition
	for _, d := range defs {
		out = append(out, oDefinition{"function", functionDefinition{d.Name, d.Description, d.Parameters}})
	}
	return out
}
func geminiDefinitions(defs []common.ToolDefinition) ([]gDefinitions, error) {
	if len(defs) == 0 {
		return nil, nil
	}
	var functions []functionDefinition
	for _, d := range defs {
		// Gemini's current parameters dialect rejects this top-level JSON Schema
		// keyword. Project only that known difference; do not mutate the common
		// declaration or imply support for arbitrary future schemas.
		var parameters map[string]json.RawMessage
		if err := json.Unmarshal(d.Parameters, &parameters); err != nil {
			return nil, err
		}
		delete(parameters, "additionalProperties")
		projected, err := json.Marshal(parameters)
		if err != nil {
			return nil, err
		}
		functions = append(functions, functionDefinition{d.Name, d.Description, projected})
	}
	return []gDefinitions{{Functions: functions}}, nil
}

// Official client-side tool search: deferred declarations remain in the request
// catalog; only a tool_reference makes them visible to Claude. This is not an
// instruction role and does not make an unloaded catalog entry dispatchable.
// https://platform.claude.com/docs/en/agents-and-tools/tool-use/tool-search-tool
// Unknown/unsupported model families deliberately use ordinary declarations.
type anthropicDisclosure struct {
	catalog    []common.ToolSpec
	references map[string][]string
}

func supportsToolReferences(model string) bool {
	for _, family := range []string{"claude-fable-5-1", "claude-mythos-5-1", "claude-fable-5", "claude-mythos-5", "claude-opus-5-5", "claude-opus-5", "claude-sonnet-5-5", "claude-sonnet-5", "claude-haiku-5-5", "claude-opus-4-8", "claude-opus-4-7", "claude-opus-4-6", "claude-sonnet-4-6", "claude-opus-4-5", "claude-sonnet-4-5", "claude-haiku-4-5"} {
		if model == family || strings.HasPrefix(model, family+"-20") {
			return true
		}
	}
	return false
}
func anthropicCatalog(tools []common.ToolSpec, disclosed map[string]bool) []aDefinition {
	out := []aDefinition{}
	for _, t := range tools {
		out = append(out, aDefinition{Name: t.Name, Description: t.Description, InputSchema: t.Parameters, DeferLoading: !t.Enabled || disclosed[t.Name]})
	}
	return out
}
